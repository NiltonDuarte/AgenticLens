package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// ErrWorkspacesDirNotFound is returned when no workspaces directory is specified via argument/flag or project configuration.
var ErrWorkspacesDirNotFound = errors.New("workspaces directory not found")

// ResolveWorkspacesDir resolves workspaces directory from argument, config.
// Prioritizes explicit overrides, followed by configured project paths.
func ResolveWorkspacesDir(overridePath string) (string, error) {
	if trimmed := strings.TrimSpace(overridePath); trimmed != "" {
		return trimmed, nil
	}

	// 1. Configured workspace dir from centralized config loader
	cfg, err := config.Load("")
	if err == nil && strings.TrimSpace(cfg.Workspaces.Dir) != "" {
		return strings.TrimSpace(cfg.Workspaces.Dir), nil
	}
	return "", ErrWorkspacesDirNotFound
}

// ScanWorkspaces scans target workspaces directory for all agentic_metrics.json ledgers.
func ScanWorkspaces(basePath string) domain.ScanReport {
	targetDir, err := ResolveWorkspacesDir(basePath)
	if err != nil {
		return domain.ScanReport{
			Ledgers: []domain.AgenticMetricsLedger{},
			Errors:  []string{fmt.Sprintf("Workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Workspaces directory is unconfigured or not found: %v", err)},
		}
	}

	fi, err := os.Stat(targetDir)
	if err != nil || !fi.IsDir() {
		return domain.ScanReport{
			Ledgers: []domain.AgenticMetricsLedger{},
			Errors:  []string{fmt.Sprintf("Workspaces directory not found: %s", targetDir)},
		}
	}

	var metricFiles []string
	err = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && d.Name() == "agentic_metrics.json" {
			metricFiles = append(metricFiles, path)
		}
		return nil
	})
	if err != nil {
		return domain.ScanReport{
			Ledgers: []domain.AgenticMetricsLedger{},
			Errors:  []string{fmt.Sprintf("Failed walking workspaces directory: %v", err)},
		}
	}

	sort.Strings(metricFiles)

	cfg, _ := config.Load(basePath)

	var ledgers []domain.AgenticMetricsLedger
	var scanErrors []string

	for _, metricFile := range metricFiles {
		content, err := os.ReadFile(metricFile)
		if err != nil {
			scanErrors = append(scanErrors, fmt.Sprintf("Failed to load %s: %v", metricFile, err))
			continue
		}

		var dto AgenticFactLedgerDTO
		if err := json.Unmarshal(content, &dto); err != nil {
			scanErrors = append(scanErrors, fmt.Sprintf("Failed to load %s: %v", metricFile, err))
			continue
		}
		if err := dto.Validate(); err != nil {
			scanErrors = append(scanErrors, fmt.Sprintf("Failed to load %s: %v", metricFile, err))
			continue
		}

		factLedger := dto.ToDomain()

		// Deduce ticket_id from parent directory if omitted
		if factLedger.TicketID == nil || *factLedger.TicketID == "" {
			parentDir := filepath.Dir(metricFile)
			parentName := filepath.Base(parentDir)
			if parentName != "workspaces" && parentName != "." {
				factLedger.TicketID = &parentName
			}
		}

		ledger := domain.DeriveMetricsLedger(factLedger, cfg)
		ledgers = append(ledgers, ledger)
	}

	sortedLedgers := domain.SortLedgers(ledgers, domain.SortFieldTicketID, true)

	return domain.ScanReport{
		Ledgers: sortedLedgers,
		Errors:  scanErrors,
	}
}

// ReadTicketArtifacts safely reads configured documentation artifacts for a ticket workspace.
func ReadTicketArtifacts(ticketID string, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) domain.TicketArtifacts {
	targetDir, err := ResolveWorkspacesDir(basePath)
	if err != nil {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}

	ticketDir := filepath.Join(absTargetDir, ticketID)
	absTicketDir, err := filepath.Abs(ticketDir)
	if err != nil {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}

	// Verify path containment to prevent directory traversal
	rel, err := filepath.Rel(absTargetDir, absTicketDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}

	fi, err := os.Stat(absTicketDir)
	if err != nil || !fi.IsDir() {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}

	var configs []domain.WorkspaceArtifactConfig
	if len(configuredArtifacts) > 0 && configuredArtifacts[0] != nil {
		configs = configuredArtifacts[0]
	} else {
		// Try to load configured artifacts from project config
		if projCfg, err := config.Load(basePath); err == nil && len(projCfg.Workspaces.Artifacts) > 0 {
			configs = projCfg.Workspaces.Artifacts
		}
	}

	if len(configs) == 0 {
		return domain.TicketArtifacts{TicketID: ticketID, Items: []domain.TicketArtifactItem{}}
	}

	items := make([]domain.TicketArtifactItem, 0, len(configs))
	var legacyContract *string
	var legacySummary *string

	for _, cfg := range configs {
		cleanRelPath := filepath.Clean(strings.TrimSpace(cfg.Path))
		if cleanRelPath == "" || filepath.IsAbs(cleanRelPath) || strings.HasPrefix(cleanRelPath, "..") || cleanRelPath == "." {
			continue
		}

		fullArtifactPath := filepath.Join(absTicketDir, cleanRelPath)
		absArtifactPath, err := filepath.Abs(fullArtifactPath)
		if err != nil {
			continue
		}

		// Ensure artifact file is inside ticket directory
		artRel, err := filepath.Rel(absTicketDir, absArtifactPath)
		if err != nil || strings.HasPrefix(artRel, "..") {
			continue
		}

		var contentPtr *string
		if afi, err := os.Stat(absArtifactPath); err == nil && !afi.IsDir() {
			if data, err := os.ReadFile(absArtifactPath); err == nil {
				s := string(data)
				contentPtr = &s
			}
		}

		title := strings.TrimSpace(cfg.Title)
		if title == "" {
			title = filepath.Base(cleanRelPath)
		}

		item := domain.TicketArtifactItem{
			Path:        cleanRelPath,
			Title:       title,
			Description: strings.TrimSpace(cfg.Description),
			Content:     contentPtr,
		}
		items = append(items, item)

		if strings.EqualFold(cleanRelPath, "CONTRACT.md") && legacyContract == nil {
			legacyContract = contentPtr
		}
		if strings.EqualFold(cleanRelPath, "SUMMARY.md") && legacySummary == nil {
			legacySummary = contentPtr
		}
	}

	return domain.TicketArtifacts{
		TicketID:        ticketID,
		ContractContent: legacyContract,
		SummaryContent:  legacySummary,
		Items:           items,
	}
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
	content, err := os.ReadFile(metricFile)
	if err != nil {
		return nil, fmt.Errorf("reading metrics file for ticket %s: %w", cleanID, err)
	}

	var dto AgenticFactLedgerDTO
	if err := json.Unmarshal(content, &dto); err != nil {
		return nil, fmt.Errorf("unmarshaling metrics for ticket %s: %w", cleanID, err)
	}
	if err := dto.Validate(); err != nil {
		return nil, fmt.Errorf("validating metrics for ticket %s: %w", cleanID, err)
	}

	factLedger := dto.ToDomain()
	if factLedger.TicketID == nil || *factLedger.TicketID == "" {
		factLedger.TicketID = &cleanID
	}

	cfg, _ := config.Load(basePath)
	ledger := domain.DeriveMetricsLedger(factLedger, cfg)
	return &ledger, nil
}

// UpdateSummaryFile idempotently updates or appends the metrics table inside the summary file.
func UpdateSummaryFile(destPath string, tableMD string) error {
	parentDir := filepath.Dir(destPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating parent directories for %s: %w", destPath, err)
	}

	header := "### Workflow Efficiency"
	legacyHeader := "### Agentic Execution Metrics"

	if data, err := os.ReadFile(destPath); err == nil {
		content := string(data)
		targetIdx := -1
		if idx := strings.Index(content, header); idx != -1 {
			targetIdx = idx
		} else if idx := strings.Index(content, legacyHeader); idx != -1 {
			targetIdx = idx
		}

		if targetIdx != -1 {
			nextHeading := strings.Index(content[targetIdx+len(header):], "\n## ")
			var newContent string
			if nextHeading != -1 {
				absNextHeading := targetIdx + len(header) + nextHeading
				newContent = strings.TrimRight(content[:targetIdx], " \t\r\n") +
					"\n\n" +
					strings.TrimSpace(tableMD) +
					"\n\n" +
					strings.TrimLeft(content[absNextHeading:], " \t\r\n")
			} else {
				newContent = strings.TrimRight(content[:targetIdx], " \t\r\n") +
					"\n\n" +
					strings.TrimSpace(tableMD) +
					"\n"
			}
			if err := os.WriteFile(destPath, []byte(newContent), 0644); err != nil {
				return fmt.Errorf("writing updated summary to %s: %w", destPath, err)
			}
			return nil
		}
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("opening %s for appending: %w", destPath, err)
	}
	defer f.Close()

	tableContent := "\n" + strings.TrimSpace(tableMD) + "\n"
	if _, err := f.WriteString(tableContent); err != nil {
		return fmt.Errorf("writing metrics table to %s: %w", destPath, err)
	}
	return nil
}

// ExistingLedgerFacts encapsulates raw facts extracted from an existing on-disk ledger.
type ExistingLedgerFacts struct {
	SchemaVersion *int
	Migration     *domain.MigrationInfo
	TicketID      *string
	TicketTitle   *string
	TicketType    *domain.ConventionalTicketType
	StartedAt     *string
	CompletedAt   *string
	RootThreadIDs []string
	Code          *domain.CodeFacts
	Invocations   []domain.InvocationFact
	Tools         []domain.ToolFact
	Events        []domain.EventFact
}

type legacyThreadItem struct {
	SessionID string `json:"session_id"`
}

type legacyCodeDelivered struct {
	LinesAdded   int64 `json:"lines_added"`
	LinesDeleted int64 `json:"lines_deleted"`
	FilesChanged int64 `json:"files_changed"`
}

type legacyTestToProdRatio struct {
	TestLinesAdded int64 `json:"test_lines_added"`
	ProdLinesAdded int64 `json:"prod_lines_added"`
}

type legacyWorkflowEfficiency struct {
	SDLCIterations        int64                 `json:"sdlc_iterations"`
	CodeDelivered         legacyCodeDelivered   `json:"code_delivered"`
	TestToProductionRatio legacyTestToProdRatio `json:"test_to_production_ratio"`
}

type rawExistingLedger struct {
	SchemaVersion      *int                      `json:"schema_version,omitempty"`
	Migration          *MigrationInfoDTO         `json:"migration,omitempty"`
	TicketID           *string                   `json:"ticket_id"`
	TicketTitle        *string                   `json:"ticket_title"`
	TicketType         *string                   `json:"ticket_type"`
	StartedAt          *string                   `json:"started_at"`
	CompletedAt        *string                   `json:"completed_at"`
	Code               *CodeFactsDTO             `json:"code"`
	Invocations        []InvocationFactDTO       `json:"invocations"`
	Tools              []ToolFactDTO             `json:"tools"`
	Events             []EventFactDTO            `json:"events,omitempty"`
	Threads            []legacyThreadItem        `json:"threads"`
	WorkflowEfficiency *legacyWorkflowEfficiency `json:"workflow_efficiency"`
}

// LoadExistingFactLedger loads raw facts from an existing agentic_metrics.json ledger on disk.
func LoadExistingFactLedger(path string) (*ExistingLedgerFacts, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}

	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading existing ledger at %s: %w", path, err)
	}

	var raw rawExistingLedger
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decoding existing ledger at %s: %w", path, err)
	}
	for i, tool := range raw.Tools {
		if strings.TrimSpace(tool.SessionID) == "" {
			return nil, fmt.Errorf("loading existing ledger at %s: %w: tools[%d] is missing; regenerate or re-extract this ledger from source logs", path, ErrToolSessionIDRequired, i)
		}
	}

	var mig *domain.MigrationInfo
	if raw.Migration != nil {
		domMig := raw.Migration.ToDomain()
		mig = &domMig
	}

	var events []domain.EventFact
	if len(raw.Events) > 0 {
		events = make([]domain.EventFact, len(raw.Events))
		for i, e := range raw.Events {
			events[i] = e.ToDomain()
		}
	}

	result := &ExistingLedgerFacts{
		SchemaVersion: raw.SchemaVersion,
		Migration:     mig,
		TicketID:      raw.TicketID,
		TicketTitle:   raw.TicketTitle,
		TicketType:    toDomainTicketType(raw.TicketType),
		StartedAt:     raw.StartedAt,
		CompletedAt:   raw.CompletedAt,
		Events:        events,
	}

	// Schema 1: Direct atomic fact schema
	if raw.Code != nil || len(raw.Invocations) > 0 || len(raw.Tools) > 0 || raw.SchemaVersion != nil {
		var invs []domain.InvocationFact
		var rootIDs []string
		for _, dto := range raw.Invocations {
			domInv := dto.ToDomain()
			invs = append(invs, domInv)
			if domInv.ParentID == nil {
				rootIDs = append(rootIDs, domInv.ID)
			}
		}
		result.Invocations = invs
		result.RootThreadIDs = rootIDs

		var tools []domain.ToolFact
		for _, dto := range raw.Tools {
			tools = append(tools, dto.ToDomain())
		}
		result.Tools = tools

		if raw.Code != nil {
			c := raw.Code.ToDomain()
			result.Code = &c
		}

		return result, nil
	}

	// Schema 2: Legacy schema with threads and workflow_efficiency
	var rootIDs []string
	for _, t := range raw.Threads {
		if strings.TrimSpace(t.SessionID) != "" {
			rootIDs = append(rootIDs, strings.TrimSpace(t.SessionID))
		}
	}
	result.RootThreadIDs = rootIDs

	if raw.WorkflowEfficiency != nil {
		we := raw.WorkflowEfficiency
		result.Code = &domain.CodeFacts{
			LinesAdded:       we.CodeDelivered.LinesAdded,
			LinesDeleted:     we.CodeDelivered.LinesDeleted,
			TestLinesAdded:   we.TestToProductionRatio.TestLinesAdded,
			TestLinesDeleted: 0,
			ProdLinesAdded:   we.TestToProductionRatio.ProdLinesAdded,
			ProdLinesDeleted: 0,
			FilesChanged:     we.CodeDelivered.FilesChanged,
			SDLCIterations:   we.SDLCIterations,
		}
	}

	return result, nil
}

// SaveFactLedger writes an AgenticFactLedger to the specified file path, creating parent directories if needed.
func SaveFactLedger(path string, ledger domain.AgenticFactLedger) error {
	if ledger.TicketID == nil || strings.TrimSpace(*ledger.TicketID) == "" {
		return fmt.Errorf("cannot save fact ledger: TicketID is required and cannot be empty")
	}
	if ledger.TicketTitle == nil || strings.TrimSpace(*ledger.TicketTitle) == "" {
		return fmt.Errorf("cannot save fact ledger: TicketTitle is required and cannot be empty")
	}
	if err := ValidateToolFacts(ledger.Tools); err != nil {
		return fmt.Errorf("cannot save fact ledger: %w", err)
	}

	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal fact ledger: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write fact ledger to %s: %w", path, err)
	}
	return nil
}

// SaveMetricsLedger writes an AgenticMetricsLedger to the specified file path after serializing to external DTO format,
// creating parent directories if needed.
func SaveMetricsLedger(path string, ledger domain.AgenticMetricsLedger) error {
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	serialized := SerializeLedger(ledger)
	data, err := json.MarshalIndent(serialized, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics ledger: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics ledger to %s: %w", path, err)
	}
	return nil
}
