package backfill

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// BackfillEngine coordinates scanning, Git recomputation, OpenCode SQLite querying,
// non-destructive fallback preservation, atomic JSON writes, and summary updates.
type BackfillEngine struct{}

// NewEngine creates a new BackfillEngine instance.
func NewEngine() *BackfillEngine {
	return &BackfillEngine{}
}

// Execute runs the backfill and migration workflow across target workspaces.
func (e *BackfillEngine) Execute(opts BackfillOptions) (*BackfillReport, error) {
	targetDir, err := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	if err != nil {
		return nil, fmt.Errorf("resolving workspaces directory: %w", err)
	}

	fi, err := os.Stat(targetDir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("workspaces directory not found: %s", targetDir)
	}

	cfg, err := config.Load(targetDir)
	if err != nil {
		cfg, _ = config.Load("")
	}

	report := &BackfillReport{
		Warnings: []string{},
		Results:  []TicketMigrationResult{},
	}

	// 1. Discover all agentic_metrics.json files
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
		return nil, fmt.Errorf("walking workspaces directory %s: %w", targetDir, err)
	}

	sort.Strings(metricFiles)

	// Filter by ticket ID if specified
	var targetFiles []string
	cleanTicketID := strings.TrimSpace(opts.TicketID)
	for _, f := range metricFiles {
		parentDir := filepath.Dir(f)
		parentName := filepath.Base(parentDir)
		if cleanTicketID != "" && !strings.EqualFold(parentName, cleanTicketID) {
			continue
		}
		targetFiles = append(targetFiles, f)
	}

	report.Scanned = len(targetFiles)
	if len(targetFiles) == 0 {
		return report, nil
	}

	// 1.5. If schema migration requested, execute universal schema migration runner first
	if opts.MigrateSchema {
		migRunner := NewMigrationRunner(nil)
		migOpts := MigrationOptions{
			WorkspacesDir:   opts.WorkspacesDir,
			TicketID:        opts.TicketID,
			TargetVersion:   domain.CurrentSchemaVersion,
			DryRun:          opts.DryRun,
			Backup:          false,
			UpdateSummaries: false,
		}
		migReport, migErr := migRunner.Run(migOpts)
		report.SchemaMigration = migReport
		if migErr != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Schema migration failed: %v", migErr))
		} else if migReport != nil {
			for _, w := range migReport.Errors {
				report.Warnings = append(report.Warnings, fmt.Sprintf("Schema migration: %s", w))
			}
		}
	}

	// Prepare SQLite DB handle if DB recompute requested
	var opencodeDB *opencode.OpenCodeDB
	if opts.RecomputeDB {
		dbHandle, err := opencode.OpenDB(opts.DBPath)
		if err != nil {
			// If DB cannot be opened at all, warn and proceed without local DB (preserving telemetry)
			report.Warnings = append(report.Warnings, fmt.Sprintf("OpenCode SQLite DB unavailable (%v); will preserve existing telemetry", err))
		} else {
			opencodeDB = dbHandle
			defer opencodeDB.Close()
		}
	}

	gitUser := git.GetGitUser()
	classificationCfg, _ := git.LoadConfig("")
	nowISO := time.Now().UTC().Format(time.RFC3339Nano)

	// 2. Iterate each discovered workspace with per-ticket error isolation
	for _, metricPath := range targetFiles {
		ticketResult := e.processSingleWorkspace(
			metricPath,
			opts,
			opencodeDB,
			gitUser,
			classificationCfg,
			nowISO,
			cfg,
		)

		if ticketResult.Error != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Workspace %s: %v", metricPath, ticketResult.Error))
		}

		if ticketResult.Skipped {
			report.Skipped++
		} else if ticketResult.Error == nil {
			report.Updated++
			if ticketResult.NewStatus == string(domain.MigrationStatusCompleted) {
				report.Completed++
			} else if ticketResult.NewStatus == string(domain.MigrationStatusPartial) {
				report.Partial++
			}
		}

		report.Results = append(report.Results, ticketResult)
	}

	return report, nil
}

func (e *BackfillEngine) processSingleWorkspace(
	metricPath string,
	opts BackfillOptions,
	opencodeDB *opencode.OpenCodeDB,
	gitUser string,
	classificationCfg *git.ClassificationConfig,
	nowISO string,
	cfg domain.ProjectConfig,
) TicketMigrationResult {
	workspaceDir := filepath.Dir(metricPath)
	workspaceName := filepath.Base(workspaceDir)

	result := TicketMigrationResult{
		TicketID:      workspaceName,
		WorkspacePath: workspaceDir,
	}

	// Read raw content to check file validity
	content, err := os.ReadFile(metricPath)
	if err != nil {
		result.Error = fmt.Errorf("reading file: %w", err)
		return result
	}

	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(content, &dto); err != nil {
		result.Error = fmt.Errorf("decoding JSON: %w", err)
		return result
	}

	factLedger := dto.ToDomain()
	ticketID := workspaceName
	if factLedger.TicketID != nil && strings.TrimSpace(*factLedger.TicketID) != "" {
		ticketID = strings.TrimSpace(*factLedger.TicketID)
	}
	result.TicketID = ticketID

	prevStatus := "unversioned"
	if factLedger.Migration != nil && factLedger.Migration.Status != "" {
		prevStatus = string(factLedger.Migration.Status)
	} else if factLedger.SchemaVersion != nil && *factLedger.SchemaVersion == domain.CurrentSchemaVersion {
		prevStatus = string(domain.MigrationStatusCompleted)
	}
	result.PreviousStatus = prevStatus

	// Skip condition:
	// Fully completed tickets are skipped unless --force is specified.
	// A ticket is considered completed if it has CurrentSchemaVersion,
	// does not have a partial migration pending, and does not require event backfill when RecomputeDB is requested.
	needsEventBackfill := opts.RecomputeDB && (factLedger.SchemaVersion == nil || *factLedger.SchemaVersion < domain.CurrentSchemaVersion) && len(factLedger.Events) == 0
	isCompleted := factLedger.SchemaVersion != nil && *factLedger.SchemaVersion == domain.CurrentSchemaVersion &&
		(factLedger.Migration == nil || factLedger.Migration.Status != domain.MigrationStatusPartial) &&
		!needsEventBackfill

	if !opts.Force && isCompleted {
		result.Skipped = true
		result.NewStatus = prevStatus
		return result
	}

	// 1. Git Recomputation
	gitSucceeded := false
	if opts.RecomputeGit {
		var fallbackCode *domain.CodeFacts
		if factLedger.Code.LinesAdded > 0 || factLedger.Code.FilesChanged > 0 {
			fallbackCode = &factLedger.Code
		}
		newCode, err := git.ComputeCodeFacts(ticketID, classificationCfg, fallbackCode)
		if err != nil {
			result.Error = fmt.Errorf("recomputing git facts for %s: %w", ticketID, err)
			return result
		}
		factLedger.Code = newCode
		gitSucceeded = true
		result.GitUpdated = true
	} else if factLedger.Code.LinesAdded > 0 || factLedger.Code.FilesChanged > 0 || factLedger.Code.SDLCIterations > 0 {
		// Existing ledger already has git facts present
		gitSucceeded = true
	} else if factLedger.Migration != nil {
		// Preserve prior git status if step was completed
		for _, s := range factLedger.Migration.StepsCompleted {
			if s == domain.MigrationStepGitCodeFacts {
				gitSucceeded = true
				break
			}
		}
	}

	// 2. OpenCode DB Recomputation with safe fallback preservation
	dbAttempted := false
	dbSessionFound := false
	hasExistingDBFacts := len(factLedger.Invocations) > 0 || len(factLedger.Events) > 0

	if opts.RecomputeDB {
		dbAttempted = true
		if opencodeDB != nil && opencodeDB.DB() != nil {
			var resolvedSessionID *string

			// Strategy 1 (Main resolution method): Existing Session ID from ledger invocations
			var existingSessionID string
			for _, inv := range factLedger.Invocations {
				if inv.ParentID == nil || strings.TrimSpace(*inv.ParentID) == "" {
					existingSessionID = strings.TrimSpace(inv.ID)
					break
				}
			}
			if existingSessionID == "" && len(factLedger.Invocations) > 0 {
				if factLedger.Invocations[0].ParentID != nil && strings.TrimSpace(*factLedger.Invocations[0].ParentID) != "" {
					existingSessionID = strings.TrimSpace(*factLedger.Invocations[0].ParentID)
				} else {
					existingSessionID = strings.TrimSpace(factLedger.Invocations[0].ID)
				}
			}

			if existingSessionID != "" {
				var validID string
				err := opencodeDB.DB().QueryRow("SELECT id FROM session WHERE id = ? LIMIT 1", existingSessionID).Scan(&validID)
				if err == nil && validID != "" {
					resolvedSessionID = &validID
				}
			}

			// Strategy 2: Ticket ID matching against root session titles
			if resolvedSessionID == nil {
				sid, err := opencode.ResolveActiveRootSession(opencodeDB.DB(), "", ticketID, "")
				if err == nil && sid != nil && strings.TrimSpace(*sid) != "" {
					resolvedSessionID = sid
				}
			}

			// Strategy 3: Workspace path match constrained by commit date and timestamp range
			if resolvedSessionID == nil {
				minSec, maxSec, found := git.GetWorkspaceCommitTimeRange(workspaceDir)
				if found {
					minMs := (minSec - 86400) * 1000
					maxMs := (maxSec + 7200) * 1000
					sid, err := opencode.ResolveSessionByWorkspacePath(opencodeDB.DB(), ticketID, minMs, maxMs)
					if err == nil && sid != nil && strings.TrimSpace(*sid) != "" {
						resolvedSessionID = sid
					}
				}
			}

			if resolvedSessionID != nil && strings.TrimSpace(*resolvedSessionID) != "" {
				dbSessionFound = true
				rootID := strings.TrimSpace(*resolvedSessionID)
				turns, err := opencode.ExtractSteeringTurns(opencodeDB.DB(), []string{rootID})
				if err == nil && turns >= 1 {
					factLedger.Code.SDLCIterations = turns
				}
				invs, tools, events, stISO, endISO, err := opencode.ExtractFacts(
					opencodeDB.DB(),
					[]string{rootID},
					gitUser,
					make(map[string]string),
				)
				if err == nil && len(invs) > 0 {
					shouldUpdateInvocations := !hasExistingDBFacts || opts.Force || (factLedger.Migration != nil && factLedger.Migration.Status == domain.MigrationStatusPartial)
					if shouldUpdateInvocations {
						factLedger.Invocations = invs
						factLedger.Tools = tools
						if stISO != nil {
							factLedger.StartedAt = stISO
						}
						if endISO != nil {
							factLedger.CompletedAt = endISO
						}
					}
					if !shouldUpdateInvocations && factLedger.CompletedAt != nil && *factLedger.CompletedAt != "" {
						cutoff := *factLedger.CompletedAt
						var filteredEvents []domain.EventFact
						for _, ev := range events {
							if ev.Timestamp <= cutoff {
								filteredEvents = append(filteredEvents, ev)
							}
						}
						factLedger.Events = filteredEvents
					} else {
						factLedger.Events = events
					}
					result.DBUpdated = true
				}
			}
		}
		result.DBFound = dbSessionFound
	}

	// 3. Pure Migration State Evaluation
	evalInput := domain.MigrationEvaluationInput{
		ExistingMigration:  factLedger.Migration,
		RecomputeGit:       opts.RecomputeGit,
		GitSucceeded:       gitSucceeded,
		RecomputeDB:        opts.RecomputeDB,
		DBAttempted:        dbAttempted,
		DBSessionFound:     dbSessionFound,
		MigratedBy:         gitUser,
		MigratedAt:         nowISO,
		HasExistingDBFacts: hasExistingDBFacts,
	}

	newMigration := domain.EvaluateMigrationStatus(evalInput)
	result.NewStatus = string(newMigration.Status)
	result.PendingSteps = newMigration.StepsPending
	result.Notes = newMigration.Notes

	version := domain.CurrentSchemaVersion
	factLedger.SchemaVersion = &version
	if newMigration.Status == domain.MigrationStatusCompleted {
		factLedger.Migration = nil
	} else {
		factLedger.Migration = &newMigration
	}
	factLedger.UpdatedAt = nowISO

	if opts.DryRun {
		return result
	}

	// 4. Atomic File Write
	persistedLedger := factLedger
	persistedLedger.Events = domain.PruneEventsForPersistence(factLedger)
	dtoToSave := adapters.AgenticFactLedgerDTOFromDomain(persistedLedger)
	data, err := json.MarshalIndent(dtoToSave, "", "  ")
	if err != nil {
		result.Error = fmt.Errorf("marshaling updated JSON: %w", err)
		return result
	}
	data = append(data, '\n')

	tempPath := metricPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		result.Error = fmt.Errorf("writing temporary ledger file: %w", err)
		return result
	}

	if err := os.Rename(tempPath, metricPath); err != nil {
		_ = os.Remove(tempPath)
		result.Error = fmt.Errorf("atomic rename to %s: %w", metricPath, err)
		return result
	}

	// 5. Update Markdown Summary if requested
	if opts.UpdateSummaries {
		summaryPath := filepath.Join(workspaceDir, "SUMMARY.md")
		derivedLedger := domain.DeriveMetricsLedger(factLedger, cfg)
		tableMD := adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
		if err := adapters.UpdateSummaryFile(summaryPath, tableMD); err != nil {
			result.Error = fmt.Errorf("updating summary file %s: %w", summaryPath, err)
			return result
		}
	}

	return result
}
