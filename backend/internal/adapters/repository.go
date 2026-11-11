package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
)

// ErrWorkspacesDirNotFound is returned when no workspaces directory is specified via argument/flag or project configuration.
var ErrWorkspacesDirNotFound = errors.New("workspaces directory not found")

// ResolveWorkspaceContext selects the target project's own policy and physical root.
func ResolveWorkspaceContext(overridePath string) (string, domain.ProjectConfig, error) {
	root, err := ResolveWorkspacesDir(overridePath)
	if err != nil {
		return "", domain.ProjectConfig{}, fmt.Errorf("select workspaces with --workspaces or shared workspaces.dir; complete project setup if required: %w", err)
	}
	cfg, err := config.LoadShared(root)
	if err != nil {
		return "", cfg, fmt.Errorf("loading target workspace policy: %w", err)
	}
	return root, cfg, nil
}

// ResolveWorkspacesDir resolves only the selected path. Explicit representation-only
// operations do not need unrelated project/classification/source settings.
func ResolveWorkspacesDir(overridePath string) (string, error) {
	selected := strings.TrimSpace(overridePath)
	if selected == "" {
		cfg, err := config.LoadShared("")
		if err != nil {
			return "", errors.Join(ErrWorkspacesDirNotFound, err)
		}
		if strings.TrimSpace(cfg.Workspaces.Dir) == "" {
			return "", fmt.Errorf("workspaces root is required: supply --workspaces or configure workspaces.dir in %s", cfg.ConfigPath)
		}
		selected, err = config.ResolveConfiguredPath(filepath.Dir(cfg.ConfigPath), cfg.Workspaces.Dir)
		if err != nil {
			return "", err
		}
	} else {
		var err error
		selected, err = config.ExpandHomePath(selected)
		if err != nil {
			return "", err
		}
	}
	root, err := filesystem.PhysicalPath(selected)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("selected workspaces root %s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("selected workspaces root %s must be a directory", root)
	}
	return root, nil
}

// ScanWorkspaces scans target workspaces directory for all agentic_metrics.json ledgers.
func ScanWorkspaces(basePath string) domain.ScanReport {
	targetDir, err := ResolveWorkspacesDir(basePath)
	if err != nil {
		return domain.ScanReport{
			Ledgers:    []domain.AgenticMetricsLedger{},
			Errors:     []string{fmt.Sprintf("Workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Workspaces directory is unconfigured or not found: %v", err)},
			FatalError: fmt.Sprintf("Workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag: %v", err),
		}
	}

	fi, err := os.Stat(targetDir)
	if err != nil || !fi.IsDir() {
		return domain.ScanReport{
			Ledgers:    []domain.AgenticMetricsLedger{},
			Errors:     []string{fmt.Sprintf("Workspaces directory not found: %s", targetDir)},
			FatalError: fmt.Sprintf("Workspaces directory not found: %s", targetDir),
		}
	}

	var metricFiles []string
	type scanWalkFailure struct {
		path string
		err  error
	}
	var walkFailures []scanWalkFailure
	err = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == targetDir {
				return err
			}
			walkFailures = append(walkFailures, scanWalkFailure{path: path, err: err})
			return nil
		}
		if !d.IsDir() && d.Name() == "agentic_metrics.json" {
			metricFiles = append(metricFiles, path)
		}
		return nil
	})
	if err != nil {
		return domain.ScanReport{
			Ledgers:    []domain.AgenticMetricsLedger{},
			Errors:     []string{fmt.Sprintf("Failed walking workspaces directory: %v", err)},
			FatalError: fmt.Sprintf("Failed walking workspaces directory: %v", err),
		}
	}

	sort.Strings(metricFiles)

	cfg, configErr := config.LoadShared(targetDir)
	if configErr != nil {
		return domain.ScanReport{Ledgers: []domain.AgenticMetricsLedger{}, Errors: []string{configErr.Error()}, FatalError: configErr.Error()}
	}

	var ledgers []domain.AgenticMetricsLedger
	var scanErrors []string
	var diagnostics []domain.WorkspaceDiagnostic
	addDiagnostic := func(metricFile string, reason error) {
		ticketID := filepath.Base(filepath.Dir(metricFile))
		scanErrors = append(scanErrors, fmt.Sprintf("Failed to load %s: %v", metricFile, reason))
		guidance := reason.Error()
		if !strings.Contains(guidance, "run `agentic-lens") && !strings.Contains(guidance, "resume the explicit") {
			guidance += "; repair or regenerate this workspace's agentic_metrics.json"
		}
		diagnostics = append(diagnostics, domain.WorkspaceDiagnostic{
			TicketID: ticketID,
			Message:  fmt.Sprintf("Ticket workspace %s was skipped: %s", ticketID, guidance),
		})
	}

	for _, failure := range walkFailures {
		metricPath := failure.path
		if filepath.Base(metricPath) != "agentic_metrics.json" {
			metricPath = filepath.Join(metricPath, "agentic_metrics.json")
		}
		addDiagnostic(metricPath, fmt.Errorf("workspace traversal unavailable: %w", failure.err))
	}

	for _, metricFile := range metricFiles {
		content, err := readPublishedLedger(metricFile)
		if err != nil {
			addDiagnostic(metricFile, err)
			continue
		}
		if err := ValidateRuntimeFactLedger(content); err != nil {
			addDiagnostic(metricFile, err)
			continue
		}

		var dto AgenticFactLedgerDTO
		if err := json.Unmarshal(content, &dto); err != nil {
			addDiagnostic(metricFile, err)
			continue
		}
		if err := dto.Validate(); err != nil {
			addDiagnostic(metricFile, err)
			continue
		}

		factLedger := dto.ToDomain()
		factLedger, err = domain.NormalizeLedgerForRead(factLedger)
		if err != nil {
			addDiagnostic(metricFile, fmt.Errorf("normalization failed: %w", err))
			continue
		}

		ledger := domain.DeriveMetricsLedger(factLedger, cfg)
		ledgers = append(ledgers, ledger)
	}

	sortedLedgers := domain.SortLedgers(ledgers, domain.SortFieldTicketID, true, cfg.Ticket)

	return domain.ScanReport{
		TicketPolicy: cfg.Ticket,
		Ledgers:      sortedLedgers,
		Errors:       scanErrors,
		Diagnostics:  diagnostics,
	}
}

// ReadTicketArtifacts safely reads configured documentation artifacts for a ticket workspace.
func ReadTicketArtifacts(ticketID, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) (domain.TicketArtifacts, error) {
	result := domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	if strings.TrimSpace(ticketID) == "" || ticketID == "." || ticketID == ".." || strings.ContainsAny(ticketID, "/\\\x00") {
		return result, fmt.Errorf("invalid ticket workspace ID %q", ticketID)
	}
	root, err := ResolveWorkspacesDir(basePath)
	if err != nil {
		return result, err
	}
	ticketDir, err := filesystem.PhysicalPath(filepath.Join(root, ticketID))
	if err != nil {
		return result, err
	}
	rel, err := filepath.Rel(root, ticketDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return result, fmt.Errorf("ticket workspace is outside the selected workspace root")
	}
	if info, err := os.Stat(ticketDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return result, nil
		}
		return result, err
	} else if !info.IsDir() {
		return result, fmt.Errorf("ticket workspace must be a directory")
	}
	var entries []domain.WorkspaceArtifactConfig
	if len(configuredArtifacts) > 0 {
		entries = configuredArtifacts[0]
		if entries == nil {
			return result, fmt.Errorf("workspaces.artifacts must be an explicit list")
		}
	} else {
		cfg, err := config.LoadShared(root)
		if err != nil {
			return result, err
		}
		entries = cfg.Workspaces.Artifacts
	}
	for i, entry := range entries {
		path := entry.Path
		if strings.TrimSpace(path) == "" || filepath.IsAbs(path) || strings.ContainsAny(path, "\\\x00") {
			return result, fmt.Errorf("workspaces.artifacts[%d].path must be a safe ticket-workspace-relative path", i)
		}
		for _, part := range strings.Split(path, "/") {
			if part == "" || part == "." || part == ".." {
				return result, fmt.Errorf("workspaces.artifacts[%d].path must be a safe ticket-workspace-relative path", i)
			}
		}
		artifactPath, err := filesystem.PhysicalPath(filepath.Join(ticketDir, path))
		if err != nil {
			return result, err
		}
		relative, err := filepath.Rel(ticketDir, artifactPath)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return result, fmt.Errorf("artifact %q resolves outside its ticket workspace", path)
		}
		var content *string
		if info, err := os.Stat(artifactPath); err == nil && !info.IsDir() {
			if data, err := os.ReadFile(artifactPath); err == nil {
				text := string(data)
				content = &text
			}
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = filepath.Base(path)
		}
		result.Items = append(result.Items, domain.TicketArtifactItem{Path: path, Title: title, Description: entry.Description, Content: content})
	}
	return result, nil
}

// LoadTicketWorkspace loads a single ticket's agentic_metrics.json, converts it to domain, and derives the metrics ledger.
func LoadTicketWorkspace(ticketID, basePath string) (*domain.AgenticMetricsLedger, error) {
	cleanID := strings.TrimSpace(ticketID)
	if cleanID == "" || strings.Contains(cleanID, "..") || strings.Contains(cleanID, "/") || strings.Contains(cleanID, "\\") {
		return nil, fmt.Errorf("invalid ticket ID: %q", ticketID)
	}

	targetDir, err := ResolveWorkspacesDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("workspaces directory is not configured: %w", err)
	}
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolving workspaces directory: %w", err)
	}

	ticketDir := filepath.Join(absTargetDir, cleanID)
	absTicketDir, err := filepath.Abs(ticketDir)
	if err != nil {
		return nil, fmt.Errorf("resolving ticket directory: %w", err)
	}

	// Verify path containment to prevent directory traversal
	rel, err := filepath.Rel(absTargetDir, absTicketDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return nil, fmt.Errorf("invalid ticket path containment: %q", ticketID)
	}

	metricFile := filepath.Join(absTicketDir, "agentic_metrics.json")
	content, err := readPublishedLedger(metricFile)
	if err != nil {
		return nil, fmt.Errorf("reading metrics file for ticket %s: %w", cleanID, err)
	}
	if err := ValidateRuntimeFactLedger(content); err != nil {
		return nil, fmt.Errorf("loading metrics for ticket %s: %w", cleanID, err)
	}

	var dto AgenticFactLedgerDTO
	if err := json.Unmarshal(content, &dto); err != nil {
		return nil, fmt.Errorf("unmarshaling metrics for ticket %s: %w", cleanID, err)
	}
	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("validating metrics for ticket %s: %w", cleanID, err)
	}

	factLedger := dto.ToDomain()
	factLedger, err = domain.NormalizeLedgerForRead(factLedger)
	if err != nil {
		return nil, fmt.Errorf("normalizing metrics for ticket %s: %w", cleanID, err)
	}
	cfg, configErr := config.LoadShared(targetDir)
	if configErr != nil {
		return nil, fmt.Errorf("loading ticket configuration: %w", configErr)
	}
	ledger := domain.DeriveMetricsLedger(factLedger, cfg)
	return &ledger, nil
}

// UpdateSummaryFile replaces only the managed metrics block, preserving other evidence.
func UpdateSummaryFile(destPath, tableMD string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("creating summary directory: %w", err)
	}
	data, err := os.ReadFile(destPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading summary %s: %w", destPath, err)
	}
	updated, err := replaceSummaryMetrics(string(data), tableMD)
	if err != nil {
		return fmt.Errorf("updating summary %s: %w", destPath, err)
	}
	if err := os.WriteFile(destPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing summary %s: %w", destPath, err)
	}
	return nil
}

// ExistingLedgerFacts encapsulates raw facts extracted from an existing on-disk ledger.
type ExistingLedgerFacts struct {
	Anomalies         []domain.Anomaly
	SchemaVersion     *int
	Migration         *domain.MigrationInfo
	TicketID          *string
	TicketTitle       *string
	TicketType        *domain.ConventionalTicketType
	StartedAt         *string
	CompletedAt       *string
	RootThreadIDs     []string
	Code              *domain.CodeFacts
	Sessions          []domain.SessionFact
	PersistedSessions []SessionFactDTO
}

// LoadExistingFactLedger loads raw facts from an existing agentic_metrics.json ledger on disk.
func LoadExistingFactLedger(path string) (*ExistingLedgerFacts, error) {
	return loadExistingFactLedger(path, false)
}

// LoadOwnedFactLedger requires the caller to hold the canonical publication lock.
func LoadOwnedFactLedger(path string) (*ExistingLedgerFacts, error) {
	return loadExistingFactLedger(path, true)
}
func loadExistingFactLedger(path string, owned bool) (*ExistingLedgerFacts, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}

	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return nil, nil
	}

	var data []byte
	if owned {
		data, err = readOwnedLedger(path)
	} else {
		data, err = readPublishedLedger(path)
	}
	if err != nil {
		return nil, fmt.Errorf("reading existing ledger at %s: %w", path, err)
	}
	result, err := DecodeExistingFactLedger(data)
	if err != nil {
		return nil, fmt.Errorf("loading existing ledger at %s: %w", path, err)
	}
	return result, nil
}

// DecodeExistingFactLedger applies the canonical persisted-ledger boundary to an immutable blob.
func DecodeExistingFactLedger(data []byte) (*ExistingLedgerFacts, error) {
	if err := ValidateRuntimeFactLedger(data); err != nil {
		return nil, fmt.Errorf("validating ledger blob: %w", err)
	}
	var dto AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("decoding ledger blob: %w", err)
	}
	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("validating ledger blob: %w", err)
	}
	ledger := dto.ToDomain()
	roots := []string{}
	for _, session := range ledger.Sessions {
		if session.ParentID == nil {
			roots = append(roots, session.ID)
		}
	}
	return &ExistingLedgerFacts{Anomalies: ledger.Anomalies, SchemaVersion: ledger.SchemaVersion, Migration: ledger.Migration, TicketID: ledger.TicketID, TicketTitle: ledger.TicketTitle, TicketType: ledger.TicketType, StartedAt: ledger.StartedAt, CompletedAt: ledger.CompletedAt, RootThreadIDs: roots, Code: &ledger.Code, Sessions: ledger.Sessions, PersistedSessions: dto.Sessions}, nil
}

// SaveFactLedger writes an AgenticFactLedger to the specified file path, creating parent directories if needed.
func SaveFactLedger(path string, ledger domain.AgenticFactLedger) error {
	return SaveFactLedgerPreservingHistory(path, ledger, nil)
}

// NormalizeHistorySessionForPublication clones and validates the current schema
// representation without source hydration or legacy shape/cost conversion.
func NormalizeHistorySessionForPublication(owner SessionFactDTO, schemaVersion int) (SessionFactDTO, error) {
	if schemaVersion != domain.CurrentSchemaVersion {
		return SessionFactDTO{}, fmt.Errorf("unsupported historical schema version %d", schemaVersion)
	}
	body, err := json.Marshal(owner)
	if err != nil {
		return SessionFactDTO{}, fmt.Errorf("cloning historical fact: %w", err)
	}
	var normalized SessionFactDTO
	if err = json.Unmarshal(body, &normalized); err != nil {
		return SessionFactDTO{}, fmt.Errorf("validating historical fact: %w", err)
	}

	version := domain.CurrentSchemaVersion
	if err = (AgenticFactLedgerDTO{SchemaVersion: &version, Sessions: []SessionFactDTO{normalized}}).validateCostFields(true); err != nil {
		return SessionFactDTO{}, fmt.Errorf("historical fact cannot publish under the current cost contract: %w", err)
	}
	return normalized, nil
}

// SaveFactLedgerPreservingHistory compacts current owners and serializes
// previously authorized immutable historical DTOs through the same canonical
// validation and writer. Supported lean facts, summaries, order and nullable
// values in retained rows never pass through source persistence transforms.
func SaveFactLedgerPreservingHistory(path string, ledger domain.AgenticFactLedger, history []SessionFactDTO) error {

	if ledger.TicketID == nil || strings.TrimSpace(*ledger.TicketID) == "" {
		return fmt.Errorf("cannot save fact ledger: TicketID is required and cannot be empty")
	}
	if ledger.TicketTitle == nil || strings.TrimSpace(*ledger.TicketTitle) == "" {
		return fmt.Errorf("cannot save fact ledger: TicketTitle is required and cannot be empty")
	}
	historical := make(map[string]SessionFactDTO, len(history))
	for _, owner := range history {
		if _, exists := historical[owner.ID]; exists {
			return fmt.Errorf("duplicate retained history owner %s", owner.ID)
		}
		historical[owner.ID] = owner
	}
	current := ledger
	current.Sessions = nil
	matched := make(map[string]bool, len(history))
	for _, owner := range ledger.Sessions {
		prior, retained := historical[owner.ID]
		if !retained {
			current.Sessions = append(current.Sessions, owner)
			continue
		}
		if matched[owner.ID] {
			return fmt.Errorf("duplicate publication history owner %s", owner.ID)
		}
		if !reflect.DeepEqual(prior.ToDomain(), owner) {
			return fmt.Errorf("retained history owner %s changed before serialization", owner.ID)
		}
		matched[owner.ID] = true
	}
	if len(matched) != len(history) {
		return errors.New("retained history owner missing from publication")
	}
	allSummaries, err := domain.PersistAnomalySummaries(ledger)
	if err != nil {
		return fmt.Errorf("cannot derive resulting ledger summaries: %w", err)
	}
	ledger, err = domain.PersistAnomalySummaries(current)
	if err != nil {
		return fmt.Errorf("cannot prepare anomaly summaries: %w", err)
	}
	// Aggregate summaries cover refreshed and retained native owners together.
	// The retained owner DTOs below remain independent immutable source records.
	ledger.Anomalies = allSummaries.Anomalies
	dto := AgenticFactLedgerDTOFromDomain(ledger)
	// Classification already authorized and normalized historical facts. Validate
	// their current representation again without compacting their evidence.
	for _, owner := range history {
		cloned, err := NormalizeHistorySessionForPublication(owner, domain.CurrentSchemaVersion)
		if err != nil {
			return err
		}
		dto.Sessions = append(dto.Sessions, cloned)
	}
	if err := dto.Validate(); err != nil {
		return fmt.Errorf("cannot save fact ledger: %w", err)
	}
	if err := dto.ValidatePersistedEvidence(); err != nil {
		return fmt.Errorf("cannot save fact ledger: %w", err)
	}

	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal fact ledger: %w", err)
	}
	data, err = FormatLedgerJSON(data)
	if err != nil {
		return fmt.Errorf("formatting ledger: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write fact ledger to %s: %w", path, err)
	}
	return nil
}

// SaveMetricsLedger writes an AgenticMetricsLedger to the specified file path after serializing to external DTO format,
// creating parent directories if needed.
func SaveMetricsLedger(path string, ledger domain.AgenticMetricsLedger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return saveMetricsLedger(ctx, path, ledger)
}

func saveMetricsLedger(ctx context.Context, path string, ledger domain.AgenticMetricsLedger) (err error) {
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// Computed export is an explicit snapshot, not a baseline merge. Own the
	// output here without taking source scanner locks recursively.
	path, err = capture.ProjectDirectory(path)
	if err != nil {
		return fmt.Errorf("resolving metrics output: %w", err)
	}
	owner, err := capture.AcquirePublication(ctx, path)
	if err != nil {
		return fmt.Errorf("acquiring metrics publication ownership: %w", err)
	}
	defer func() { err = errors.Join(err, owner.Close()) }()
	if err := CheckPublication(path); err != nil {
		return err
	}

	serialized := SerializeLedger(ledger)
	data, err := json.Marshal(serialized)

	if err != nil {
		return fmt.Errorf("failed to marshal metrics ledger: %w", err)
	}
	data, err = FormatLedgerJSON(data)
	if err != nil {
		return fmt.Errorf("formatting ledger: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics ledger to %s: %w", path, err)
	}
	return nil
}
