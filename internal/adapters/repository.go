package adapters

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/domain"
)

type configDTO struct {
	WorkspacesDir string `json:"workspaces_dir"`
}

// ResolveWorkspacesDir resolves workspaces directory from argument, environment, config, or parent lookup.
func ResolveWorkspacesDir(overridePath string) string {
	if overridePath != "" {
		return overridePath
	}

	// 1. Container volume mount path
	if fi, err := os.Stat("/app/workspaces"); err == nil && fi.IsDir() {
		return "/app/workspaces"
	}

	// 2. Configured workspace dir in .agenticlens/config.json
	cwd, err := os.Getwd()
	if err == nil {
		curr := cwd
		for {
			cfgPath := filepath.Join(curr, ".agenticlens", "config.json")
			if fi, err := os.Stat(cfgPath); err == nil && !fi.IsDir() {
				if data, err := os.ReadFile(cfgPath); err == nil {
					var cfg configDTO
					if err := json.Unmarshal(data, &cfg); err == nil && cfg.WorkspacesDir != "" {
						candidate := filepath.Join(curr, cfg.WorkspacesDir)
						if cfi, err := os.Stat(candidate); err == nil && cfi.IsDir() {
							return candidate
						}
					}
				}
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	// 3. Current working directory ./workspaces
	if cwd != "" {
		candidate := filepath.Join(cwd, "workspaces")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate
		}
	}

	// 4. Parent traversal (e.g. running from a nested subfolder)
	if cwd != "" {
		curr := cwd
		for {
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			candidate := filepath.Join(parent, "workspaces")
			if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
				return candidate
			}
			curr = parent
		}
	}

	return "workspaces"
}

// ScanWorkspaces scans target workspaces directory for all agentic_metrics.json ledgers.
func ScanWorkspaces(basePath string) domain.ScanReport {
	targetDir := ResolveWorkspacesDir(basePath)

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

		factLedger := dto.ToDomain()

		// Deduce ticket_id from parent directory if omitted
		if factLedger.TicketID == nil || *factLedger.TicketID == "" {
			parentDir := filepath.Dir(metricFile)
			parentName := filepath.Base(parentDir)
			if parentName != "workspaces" && parentName != "." {
				factLedger.TicketID = &parentName
			}
		}

		ledger := domain.DeriveMetricsLedger(factLedger)
		ledgers = append(ledgers, ledger)
	}

	return domain.ScanReport{
		Ledgers: ledgers,
		Errors:  scanErrors,
	}
}

// ReadTicketArtifacts safely reads CONTRACT.md and SUMMARY.md artifacts for a ticket workspace.
func ReadTicketArtifacts(ticketID string, basePath string) domain.TicketArtifacts {
	targetDir := ResolveWorkspacesDir(basePath)
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return domain.TicketArtifacts{TicketID: ticketID}
	}

	ticketDir := filepath.Join(absTargetDir, ticketID)
	absTicketDir, err := filepath.Abs(ticketDir)
	if err != nil {
		return domain.TicketArtifacts{TicketID: ticketID}
	}

	// Verify path containment to prevent directory traversal
	rel, err := filepath.Rel(absTargetDir, absTicketDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return domain.TicketArtifacts{TicketID: ticketID}
	}

	fi, err := os.Stat(absTicketDir)
	if err != nil || !fi.IsDir() {
		return domain.TicketArtifacts{TicketID: ticketID}
	}

	var contractContent *string
	contractPath := filepath.Join(absTicketDir, "CONTRACT.md")
	if cfi, err := os.Stat(contractPath); err == nil && !cfi.IsDir() {
		if data, err := os.ReadFile(contractPath); err == nil {
			s := string(data)
			contractContent = &s
		}
	}

	var summaryContent *string
	summaryPath := filepath.Join(absTicketDir, "SUMMARY.md")
	if sfi, err := os.Stat(summaryPath); err == nil && !sfi.IsDir() {
		if data, err := os.ReadFile(summaryPath); err == nil {
			s := string(data)
			summaryContent = &s
		}
	}

	return domain.TicketArtifacts{
		TicketID:        ticketID,
		ContractContent: contractContent,
		SummaryContent:  summaryContent,
	}
}

// ExistingLedgerFacts encapsulates raw facts extracted from an existing on-disk ledger.
type ExistingLedgerFacts struct {
	SchemaVersion *int
	Migration     *domain.MigrationInfo
	TicketID      *string
	TicketTitle   *string
	TicketType    *string
	StartedAt     *string
	CompletedAt   *string
	RootThreadIDs []string
	Code          *domain.CodeFacts
	Invocations   []domain.InvocationFact
	Tools         []domain.ToolFact
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

	var mig *domain.MigrationInfo
	if raw.Migration != nil {
		domMig := raw.Migration.ToDomain()
		mig = &domMig
	}

	result := &ExistingLedgerFacts{
		SchemaVersion: raw.SchemaVersion,
		Migration:     mig,
		TicketID:      raw.TicketID,
		TicketTitle:   raw.TicketTitle,
		TicketType:    raw.TicketType,
		StartedAt:     raw.StartedAt,
		CompletedAt:   raw.CompletedAt,
	}

	// Schema 1: Direct atomic fact schema
	if len(raw.Invocations) > 0 {
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
		sdlc := we.SDLCIterations
		if sdlc <= 0 {
			sdlc = 1
		}
		result.Code = &domain.CodeFacts{
			LinesAdded:       we.CodeDelivered.LinesAdded,
			LinesDeleted:     we.CodeDelivered.LinesDeleted,
			TestLinesAdded:   we.TestToProductionRatio.TestLinesAdded,
			TestLinesDeleted: 0,
			ProdLinesAdded:   we.TestToProductionRatio.ProdLinesAdded,
			ProdLinesDeleted: 0,
			FilesChanged:     we.CodeDelivered.FilesChanged,
			SDLCIterations:   sdlc,
		}
	}

	return result, nil
}
