package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

// ScanWorkspaces loads all ledger-backed homes within the configured project scopes.
func ScanWorkspaces(project string) domain.ScanReport {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return domain.ScanReport{Ledgers: []domain.AgenticMetricsLedger{}, FatalError: err.Error(), Errors: []string{err.Error()}}
	}
	locations, diagnostics := DiscoverWorkspaceLocations(cfg)
	ledgers := []domain.AgenticMetricsLedger{}
	scanErrors := []string{}
	for _, location := range locations {
		ledger, err := readWorkspaceLedger(location, cfg)
		if err != nil {
			diagnostics = append(diagnostics, domain.WorkspaceDiagnostic{Path: location.LedgerPath, Message: err.Error()})
			continue
		}
		ledgers = append(ledgers, ledger)
	}
	for _, d := range diagnostics {
		scanErrors = append(scanErrors, fmt.Sprintf("%s: %s", d.Path, d.Message))
	}
	return domain.ScanReport{TicketPolicy: cfg.Ticket, Ledgers: domain.SortLedgers(ledgers, domain.SortFieldTicketID, true, cfg.Ticket), Errors: scanErrors, Diagnostics: diagnostics}
}

func ReadTicketArtifacts(ticketID, project string, configured ...[]domain.WorkspaceArtifactConfig) (result domain.TicketArtifacts, err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return domain.TicketArtifacts{}, err
	}
	location, reference, err := resolveTicketReadLocation(cfg, ticketID, "")
	if err != nil {
		return domain.TicketArtifacts{}, err
	}
	defer func() { err = errors.Join(err, reference.Close()) }()
	entries := location.Artifacts
	if len(configured) > 0 {
		if configured[0] == nil {
			return domain.TicketArtifacts{}, fmt.Errorf("artifact policy must be an explicit list")
		}
		entries = configured[0]
	}
	return readResolvedArtifacts(ticketID, location, entries)
}

func ReadTicketArtifactsAt(ticketID, project, selection string) (result domain.TicketArtifacts, err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return domain.TicketArtifacts{}, err
	}
	location, reference, err := resolveTicketReadLocation(cfg, ticketID, selection)
	if err != nil {
		return domain.TicketArtifacts{}, err
	}
	defer func() { err = errors.Join(err, reference.Close()) }()
	return readResolvedArtifacts(ticketID, location, location.Artifacts)
}
func readResolvedArtifacts(ticketID string, location WorkspaceLocation, entries []domain.WorkspaceArtifactConfig) (result domain.TicketArtifacts, err error) {
	result = domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	lease, err := openResolvedWorkspaceDirectory(location)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, lease.Close()) }()
	return readDirectoryArtifacts(ticketID, location, entries, lease)
}

func readDirectoryArtifacts(ticketID string, location WorkspaceLocation, entries []domain.WorkspaceArtifactConfig, lease *capture.ReadDirectory) (domain.TicketArtifacts, error) {
	result := domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	check := func() error {
		if _, e := lease.Lstat("agentic_metrics.json.publication-pending"); e == nil {
			return fmt.Errorf("%w at %s", ErrPublicationPending, location.Path)
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
		return nil
	}
	if err := check(); err != nil {
		return result, err
	}
	for i, entry := range entries {
		path := entry.Path
		if strings.TrimSpace(path) == "" || filepath.IsAbs(path) || strings.ContainsAny(path, "\\\x00") {
			return result, fmt.Errorf("workspaces.artifacts[%d].path must be safe and home-relative", i)
		}
		for _, part := range strings.Split(path, "/") {
			if part == "" || part == "." || part == ".." {
				return result, fmt.Errorf("unsafe artifact path %q", path)
			}
		}
		var content *string
		data, e := lease.ReadFile(path)
		if e == nil {
			value := string(data)
			content = &value
		} else if !errors.Is(e, os.ErrNotExist) {
			return result, fmt.Errorf("artifact %q is unavailable or outside selected workspace: %w", path, e)
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = filepath.Base(path)
		}
		result.Items = append(result.Items, domain.TicketArtifactItem{Path: path, Title: title, Description: entry.Description, Content: content})
	}
	if err := check(); err != nil {
		return result, err
	}
	return result, nil
}

// LoadTicketWorkspaceDetail keeps ledger and artifacts in one held physical
// directory, including when the selected directory is renamed during the read.
func LoadTicketWorkspaceDetail(ticketID, project, selection string) (result *domain.AgenticMetricsLedger, artifacts domain.TicketArtifacts, err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	location, reference, err := resolveTicketReadLocation(cfg, ticketID, selection)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	defer func() { err = errors.Join(err, reference.Close()) }()
	lease, err := openResolvedWorkspaceDirectory(location)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	defer func() { err = errors.Join(err, lease.Close()) }()
	body, err := readDirectoryLedger(lease, location.LedgerPath)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	ledger, err := decodeWorkspaceLedger(body, location, cfg)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	if !sameTicket(*ledger.TicketID, ticketID, cfg.Ticket) {
		return nil, domain.TicketArtifacts{}, &WorkspaceResolutionError{Outcome: "identity_conflict", TicketID: ticketID, Candidates: []string{location.Path}}
	}
	artifacts, err = readDirectoryArtifacts(ticketID, location, location.Artifacts, lease)
	if err != nil {
		return nil, domain.TicketArtifacts{}, err
	}
	return &ledger, artifacts, nil
}

// LoadTicketWorkspace loads a single ticket's agentic_metrics.json, converts it to domain, and derives the metrics ledger.
func LoadTicketWorkspace(ticketID, project string, selection ...string) (result *domain.AgenticMetricsLedger, err error) {
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		return nil, err
	}
	selected := ""
	if len(selection) > 0 {
		selected = selection[0]
	}
	location, reference, err := resolveTicketReadLocation(cfg, ticketID, selected)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, reference.Close()) }()
	ledger, err := readWorkspaceLedger(location, cfg)
	if err != nil {
		return nil, err
	}
	if !sameTicket(*ledger.TicketID, ticketID, cfg.Ticket) {
		return nil, &WorkspaceResolutionError{Outcome: "identity_conflict", TicketID: ticketID, Candidates: []string{location.Path}}
	}
	return &ledger, nil
}

// UpdateSummaryFile replaces only the managed metrics block, preserving other evidence.
func UpdateSummaryFile(destPath, tableMD string) error {
	destPath = capture.PublicationPath(destPath)
	parent := filepath.Dir(destPath)
	// A held root permits creating descendants without recreating a moved or
	// deleted workspace through its former path.
	if capture.PublicationOwner(destPath) != nil {
		if err := capture.MkdirPublicationDirectories(parent, 0755); err != nil {
			return fmt.Errorf("creating summary directory: %w", err)
		}
	} else if _, err := capture.StatPublicationFile(parent); err != nil {
		return fmt.Errorf("creating summary directory: %w", err)
	}
	data, err := capture.ReadPublicationFile(destPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading summary %s: %w", destPath, err)
	}
	updated, err := replaceSummaryMetrics(string(data), tableMD)
	if err != nil {
		return fmt.Errorf("updating summary %s: %w", destPath, err)
	}
	if err := capture.WritePublicationFile(destPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing summary %s: %w", destPath, err)
	}
	return nil
}

// ExistingLedgerFacts encapsulates raw facts extracted from an existing on-disk ledger.
type ExistingLedgerFacts struct {
	Tags              map[string]string
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
	path = capture.PublicationPath(path)
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}

	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	fi, err := capture.StatPublicationFile(path)
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
	return &ExistingLedgerFacts{Tags: domain.MergeTags(nil, ledger.Tags), Anomalies: ledger.Anomalies, SchemaVersion: ledger.SchemaVersion, Migration: ledger.Migration, TicketID: ledger.TicketID, TicketTitle: ledger.TicketTitle, TicketType: ledger.TicketType, StartedAt: ledger.StartedAt, CompletedAt: ledger.CompletedAt, RootThreadIDs: roots, Code: &ledger.Code, Sessions: ledger.Sessions, PersistedSessions: dto.Sessions}, nil
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

	path = capture.PublicationPath(path)
	parentDir := filepath.Dir(path)
	if _, err := capture.StatPublicationFile(parentDir); err != nil {
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

	if err := capture.WritePublicationFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write fact ledger to %s: %w", path, err)
	}
	return nil
}

// SaveMetricsLedger publishes an explicit computed snapshot in an existing,
// configured directory; it never recreates a disappeared output home.
func SaveMetricsLedger(path string, ledger domain.AgenticMetricsLedger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return saveMetricsLedger(ctx, path, ledger)
}

func saveMetricsLedger(ctx context.Context, path string, ledger domain.AgenticMetricsLedger) (err error) {
	path = capture.PublicationPath(path)
	parentDir := filepath.Dir(path)
	if _, err := capture.StatPublicationFile(parentDir); err != nil {
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
	if err = owner.ValidateLocation(); err != nil {
		return err
	}
	path = capture.PublicationPath(path)
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

	if err := durableReplace(path, data); err != nil {
		return fmt.Errorf("failed to write metrics ledger to %s: %w", path, err)
	}
	return nil
}
