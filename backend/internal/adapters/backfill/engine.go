package backfill

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

// BackfillEngine coordinates scanning, Git recomputation, OpenCode SQLite querying,
// non-destructive fallback preservation, atomic JSON writes, and summary updates.
type BackfillEngine struct {
	discoverCodexRollouts func(codex.DiscoveryOptions) ([]string, error)
	parseCodexRollout     func(string) (codex.Transcript, error)
	combineCodex          func([]codex.Transcript) (codex.Transcript, error)
}

// NewEngine creates a new BackfillEngine instance.
func NewEngine() *BackfillEngine {
	return &BackfillEngine{
		discoverCodexRollouts: codex.DiscoverRolloutPaths,
		parseCodexRollout:     codex.ParseRolloutPath,
		combineCodex:          codex.CombineTranscripts,
	}
}

// Execute runs the backfill and migration workflow across target workspaces.
func (e *BackfillEngine) Execute(opts BackfillOptions) (*BackfillReport, error) {
	cleanSource := strings.ToLower(strings.TrimSpace(opts.Source))
	if cleanSource == "" {
		cleanSource = "opencode"
	}
	if cleanSource != "opencode" && cleanSource != "codex" && cleanSource != "both" && cleanSource != "all" {
		return nil, fmt.Errorf("invalid source %q: must be 'opencode', 'codex', or 'both'", opts.Source)
	}
	opts.Source = cleanSource

	if opts.RecomputeGit {
		if strings.TrimSpace(opts.TicketID) == "" {
			return nil, fmt.Errorf("--ticket-id is required when recomputing Git code facts (--recompute-git)")
		}
		if strings.TrimSpace(opts.CodeRef) == "" {
			return nil, fmt.Errorf("a Git code ref is required when recomputing code facts")
		}
	}
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

	// Prepare SQLite DB handle if DB recompute requested and source targets OpenCode
	var opencodeDB *opencode.OpenCodeDB
	wantsOpenCode := cleanSource == "opencode" || cleanSource == "both" || cleanSource == "all"
	if opts.RecomputeDB && wantsOpenCode {
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

	canonical, err := migration.NewEngine(nil).Migrate(content, domain.CurrentSchemaVersion)
	if err != nil {
		result.Error = fmt.Errorf("normalizing ledger before backfill: %w", err)
		return result
	}
	dto := adapters.AgenticFactLedgerDTOFromDomain(canonical)
	if err := dto.Validate(); err != nil {
		result.Error = fmt.Errorf("validating tool session provenance: %w", err)
		return result
	}

	factLedger := dto.ToDomain()
	for i := range factLedger.Sessions {
		if factLedger.Sessions[i].Source == "" {
			factLedger.Sessions[i].Source = domain.SourceOpenCode
		}
	}

	cleanSource := strings.ToLower(strings.TrimSpace(opts.Source))
	if cleanSource == "" {
		cleanSource = "opencode"
	}
	wantsOpenCode := cleanSource == "opencode" || cleanSource == "both" || cleanSource == "all"
	wantsCodex := cleanSource == "codex" || cleanSource == "both" || cleanSource == "all"

	var openCodeSessions []domain.SessionFact
	var codexSessions []domain.SessionFact
	var otherSessions []domain.SessionFact
	for _, s := range factLedger.Sessions {
		switch s.Source {
		case domain.SourceCodex:
			codexSessions = append(codexSessions, s)
		case domain.SourceOpenCode:
			openCodeSessions = append(openCodeSessions, s)
		default:
			otherSessions = append(otherSessions, s)
		}
	}

	sourceVersion, _ := migration.NewEngine(nil).DetectVersion(content)
	needsNormalization := sourceVersion < domain.CurrentSchemaVersion || hasLegacyFactCollections(content)
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

	hasTargetSessions := false
	if wantsOpenCode && (len(openCodeSessions) > 0 || len(factLedger.Sessions) == 0) {
		hasTargetSessions = true
	}
	if wantsCodex && len(codexSessions) > 0 {
		hasTargetSessions = true
	}

	needsEventBackfill := false
	if opts.RecomputeDB {
		if wantsOpenCode && len(openCodeSessions) > 0 {
			eventsCount := 0
			for _, s := range openCodeSessions {
				eventsCount += len(s.Events)
			}
			if (factLedger.SchemaVersion == nil || *factLedger.SchemaVersion < domain.CurrentSchemaVersion) && eventsCount == 0 {
				needsEventBackfill = true
			}
		}
		if wantsCodex && len(codexSessions) > 0 {
			eventsCount := 0
			for _, s := range codexSessions {
				eventsCount += len(s.Events)
			}
			if (factLedger.SchemaVersion == nil || *factLedger.SchemaVersion < domain.CurrentSchemaVersion) && eventsCount == 0 {
				needsEventBackfill = true
			}
		}
	}

	isCompleted := !needsNormalization && factLedger.SchemaVersion != nil && *factLedger.SchemaVersion == domain.CurrentSchemaVersion &&
		(factLedger.Migration == nil || factLedger.Migration.Status != domain.MigrationStatusPartial) &&
		!needsEventBackfill
	cleanupCompletedMigration := factLedger.Migration != nil &&
		factLedger.Migration.Status == domain.MigrationStatusCompleted &&
		len(factLedger.Migration.StepsPending) == 0

	if !opts.Force && isCompleted && !opts.RecomputeGit && !cleanupCompletedMigration && factLedger.Migration == nil {
		result.Skipped = true
		result.NewStatus = prevStatus
		return result
	}
	if !hasTargetSessions && !opts.RecomputeGit && !opts.RecomputeDB && !needsNormalization &&
		factLedger.Migration == nil && !cleanupCompletedMigration {
		result.Skipped = true
		result.NewStatus = prevStatus
		return result
	}

	// 1. Git Recomputation
	gitSucceeded := false
	if opts.RecomputeGit {
		newCode, err := git.ComputeCodeFactsAtRef(opts.CodeRef, classificationCfg)
		if err != nil {
			result.Error = fmt.Errorf("recomputing git facts for %s: %w", ticketID, err)
			return result
		}
		if factLedger.Code.SDLCIterations > 0 {
			newCode.SDLCIterations = factLedger.Code.SDLCIterations
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

	// 2. DB Recomputation with safe fallback preservation
	dbAttempted := false
	dbSessionFound := false
	hasExistingDBFacts := len(factLedger.InvocationFacts()) > 0 || len(factLedger.EventFacts()) > 0
	recomputeDBForTicket := opts.RecomputeDB && hasTargetSessions

	if opts.RecomputeDB && wantsOpenCode && (len(openCodeSessions) > 0 || len(factLedger.Sessions) == 0) {
		dbAttempted = true
		if opencodeDB != nil && opencodeDB.DB() != nil {
			var existingOpenCodeInvocations []domain.InvocationFact
			var existingOpenCodeTools []domain.ToolFact
			var existingOpenCodeEvents []domain.EventFact
			for _, s := range openCodeSessions {
				existingOpenCodeInvocations = append(existingOpenCodeInvocations, s.Invocations...)
				for _, t := range s.Tools {
					existingOpenCodeTools = append(existingOpenCodeTools, domain.ToolFact{SessionID: s.ID, Name: t.Name, Calls: t.Calls, Errors: t.Errors})
				}
				for _, e := range s.Events {
					existingOpenCodeEvents = append(existingOpenCodeEvents, domain.EventFact{
						EventID: e.EventID, Timestamp: e.Timestamp, MessageID: e.MessageID, Turn: e.Turn,
						Role: e.Role, EventType: e.EventType, ToolName: e.ToolName, Status: e.Status,
						ErrorMessage: e.ErrorMessage, DurationMs: e.DurationMs, ContextTokens: e.ContextTokens,
						PromptContent: e.PromptContent, ReasoningContent: e.ReasoningContent,
						ToolInput: e.ToolInput, ToolOutput: e.ToolOutput, SessionID: s.ID,
					})
				}
			}
			hasExistingOpenCodeFacts := len(existingOpenCodeInvocations) > 0 || len(existingOpenCodeEvents) > 0

			var candidateRootIDs []string
			seenCandidates := make(map[string]bool)
			addCandidate := func(id string) {
				clean := strings.TrimSpace(id)
				if clean != "" && !seenCandidates[clean] {
					seenCandidates[clean] = true
					candidateRootIDs = append(candidateRootIDs, clean)
				}
			}

			for _, s := range openCodeSessions {
				if s.ParentID == nil || strings.TrimSpace(*s.ParentID) == "" {
					addCandidate(s.ID)
				} else {
					addCandidate(*s.ParentID)
				}
			}
			for _, inv := range existingOpenCodeInvocations {
				if inv.ParentID == nil || strings.TrimSpace(*inv.ParentID) == "" {
					addCandidate(inv.ID)
				} else {
					addCandidate(*inv.ParentID)
				}
			}
			if len(candidateRootIDs) == 0 {
				for _, s := range openCodeSessions {
					addCandidate(s.ID)
				}
				for _, inv := range existingOpenCodeInvocations {
					addCandidate(inv.ID)
				}
			}

			// Validate and resolve candidate IDs against the local SQLite database
			var resolvedRootIDs []string
			seenRoots := make(map[string]bool)
			for _, cid := range candidateRootIDs {
				if rootID, found := resolveSessionRootInDB(opencodeDB.DB(), cid); found && rootID != "" {
					if !seenRoots[rootID] {
						seenRoots[rootID] = true
						resolvedRootIDs = append(resolvedRootIDs, rootID)
					}
				}
			}

			// Strategy 2 & 3: Heuristic fallbacks ONLY when explicitly allowed via CLI flag
			if len(resolvedRootIDs) == 0 && opts.AllowFallbacks {
				// Strategy 2: Ticket ID matching against root session titles
				sid, err := opencode.ResolveActiveRootSession(opencodeDB.DB(), "", ticketID, "")
				if err == nil && sid != nil && strings.TrimSpace(*sid) != "" {
					cleanSID := strings.TrimSpace(*sid)
					if !seenRoots[cleanSID] {
						seenRoots[cleanSID] = true
						resolvedRootIDs = append(resolvedRootIDs, cleanSID)
					}
				}

				// Strategy 3: Workspace path match constrained by commit date and timestamp range
				if len(resolvedRootIDs) == 0 {
					minSec, maxSec, found := git.GetWorkspaceCommitTimeRange(workspaceDir)
					if found {
						minMs := (minSec - 86400) * 1000
						maxMs := (maxSec + 7200) * 1000
						sid, err := opencode.ResolveSessionByWorkspacePath(opencodeDB.DB(), ticketID, minMs, maxMs)
						if err == nil && sid != nil && strings.TrimSpace(*sid) != "" {
							cleanSID := strings.TrimSpace(*sid)
							if !seenRoots[cleanSID] {
								seenRoots[cleanSID] = true
								resolvedRootIDs = append(resolvedRootIDs, cleanSID)
							}
						}
					}
				}
			}

			if len(resolvedRootIDs) > 0 {
				turns, err := opencode.ExtractSteeringTurns(opencodeDB.DB(), resolvedRootIDs)
				if err == nil {
					factLedger.Code.SDLCIterations = turns
				}
				invs, tools, events, stISO, endISO, err := opencode.ExtractFacts(
					opencodeDB.DB(),
					resolvedRootIDs,
					gitUser,
					make(map[string]string),
				)
				if err == nil && len(invs) > 0 {
					if len(candidateRootIDs) == 0 || len(resolvedRootIDs) >= len(candidateRootIDs) {
						dbSessionFound = true
					}
					shouldUpdateInvocations := !hasExistingOpenCodeFacts || opts.Force || (factLedger.Migration != nil && factLedger.Migration.Status == domain.MigrationStatusPartial)
					updatedInvocations, updatedTools, updatedEvents := existingOpenCodeInvocations, existingOpenCodeTools, events
					if shouldUpdateInvocations {
						if len(candidateRootIDs) > 0 && len(resolvedRootIDs) < len(candidateRootIDs) {
							// Partial resolution across machines: preserve unresolved session facts
							resolvedSet := make(map[string]bool)
							for _, rid := range resolvedRootIDs {
								resolvedSet[rid] = true
							}
							for _, inv := range invs {
								resolvedSet[inv.ID] = true
								if inv.ParentID != nil && *inv.ParentID != "" {
									resolvedSet[*inv.ParentID] = true
								}
							}
							var mergedInvocations []domain.InvocationFact
							mergedInvocations = append(mergedInvocations, invs...)
							for _, exInv := range existingOpenCodeInvocations {
								if !resolvedSet[exInv.ID] && (exInv.ParentID == nil || !resolvedSet[*exInv.ParentID]) {
									mergedInvocations = append(mergedInvocations, exInv)
								}
							}
							updatedInvocations = mergedInvocations

							var mergedTools []domain.ToolFact
							mergedTools = append(mergedTools, tools...)
							for _, exTool := range existingOpenCodeTools {
								if !resolvedSet[exTool.SessionID] {
									mergedTools = append(mergedTools, exTool)
								}
							}
							updatedTools = mergedTools
						} else {
							updatedInvocations, updatedTools = invs, tools
						}
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
						updatedEvents = filteredEvents
					}
					refreshedSessions, sessionErr := domain.SessionsFromFactsWithSource(domain.SourceOpenCode, updatedInvocations, updatedTools, updatedEvents)
					if sessionErr != nil {
						result.Error = fmt.Errorf("grouping refreshed OpenCode session facts: %w", sessionErr)
						return result
					}
					openCodeSessions = refreshedSessions
					result.DBUpdated = true
				}
			}
		}
	}

	if opts.RecomputeDB && wantsCodex && len(codexSessions) > 0 {
		dbAttempted = true
		codexHome := strings.TrimSpace(opts.CodexHome)
		if codexHome == "" {
			codexHome = strings.TrimSpace(os.Getenv("CODEX_HOME"))
		}
		if codexHome == "" {
			if home, err := os.UserHomeDir(); err == nil {
				codexHome = filepath.Join(home, ".codex")
			}
		}
		if codexHome != "" {
			covered := make(map[string]struct{})
			storedIDs := make(map[string]struct{}, len(codexSessions))
			for _, session := range codexSessions {
				storedIDs[session.ID] = struct{}{}
			}
			for _, cs := range codexSessions {
				if cs.ParentID != nil {
					if _, parentStored := storedIDs[*cs.ParentID]; parentStored {
						continue
					}
				}
				if _, alreadyCovered := covered[cs.ID]; alreadyCovered {
					continue
				}
				rolloutPaths, discErr := e.discoverCodexRollouts(codex.DiscoveryOptions{
					CodexHome: codexHome,
					SessionID: cs.ID,
				})
				if discErr == nil && len(rolloutPaths) > 0 {
					transcripts := make([]codex.Transcript, 0, len(rolloutPaths))
					parseFailed := false
					for _, p := range rolloutPaths {
						t, parseErr := e.parseCodexRollout(p)
						if parseErr != nil {
							parseFailed = true
							break
						}
						transcripts = append(transcripts, t)
					}
					if !parseFailed && len(transcripts) > 0 {
						if combined, combErr := e.combineCodex(transcripts); combErr == nil {
							freshCodex, sErr := domain.SessionsFromFactsWithSource(domain.SourceCodex, combined.Invocations, combined.Tools, combined.Events)
							if sErr == nil && len(freshCodex) > 0 {
								replaceCodexTree(&codexSessions, cs.ID, freshCodex)
								for _, refreshed := range freshCodex {
									covered[refreshed.ID] = struct{}{}
								}
								dbSessionFound = true
								result.DBUpdated = true
							}
						}
					}
				}
			}
		}
	}

	result.DBFound = dbSessionFound

	// Recombine sessions preserving all non-targeted sessions untouched
	var finalSessions []domain.SessionFact
	finalSessions = append(finalSessions, openCodeSessions...)
	finalSessions = append(finalSessions, codexSessions...)
	finalSessions = append(finalSessions, otherSessions...)
	sort.SliceStable(finalSessions, func(i, j int) bool {
		return finalSessions[i].ID < finalSessions[j].ID
	})
	factLedger.Sessions = finalSessions

	// 3. Pure Migration State Evaluation
	if dbSessionFound && factLedger.Migration != nil {
		completed := append([]string(nil), factLedger.Migration.StepsCompleted...)
		pending := make([]string, 0, len(factLedger.Migration.StepsPending))
		for _, step := range factLedger.Migration.StepsPending {
			if step == domain.MigrationStepSourceBackfill {
				continue
			}
			pending = append(pending, step)
		}
		completed = append(completed, domain.MigrationStepSourceBackfill)
		factLedger.Migration.StepsCompleted = completed
		factLedger.Migration.StepsPending = pending
	}
	evalInput := domain.MigrationEvaluationInput{
		ExistingMigration:  factLedger.Migration,
		RecomputeGit:       opts.RecomputeGit,
		GitSucceeded:       gitSucceeded,
		RecomputeDB:        recomputeDBForTicket,
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
	persistedLedger, err := domain.PersistAnomalySummaries(factLedger)
	if err != nil {
		result.Error = fmt.Errorf("preparing session-owned facts: %w", err)
		return result
	}
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

func replaceCodexTree(stored *[]domain.SessionFact, requestedRoot string, refreshed []domain.SessionFact) {
	refreshedIDs := make(map[string]struct{}, len(refreshed))
	for _, session := range refreshed {
		refreshedIDs[session.ID] = struct{}{}
	}
	removeIDs := make(map[string]struct{}, len(refreshedIDs)+1)
	removeIDs[requestedRoot] = struct{}{}
	for id := range refreshedIDs {
		removeIDs[id] = struct{}{}
	}
	for changed := true; changed; {
		changed = false
		for _, session := range *stored {
			if session.ParentID == nil {
				continue
			}
			if _, parentMatched := removeIDs[*session.ParentID]; parentMatched {
				if _, exists := removeIDs[session.ID]; !exists {
					removeIDs[session.ID] = struct{}{}
					changed = true
				}
			}
		}
	}
	result := make([]domain.SessionFact, 0, len(*stored)-len(removeIDs)+len(refreshed))
	for _, session := range *stored {
		if _, remove := removeIDs[session.ID]; !remove {
			result = append(result, session)
		}
	}
	result = append(result, refreshed...)
	*stored = result
}

func resolveSessionRootInDB(db *sql.DB, id string) (string, bool) {
	curr := strings.TrimSpace(id)
	if curr == "" || db == nil {
		return "", false
	}
	visited := make(map[string]bool)
	for {
		if visited[curr] {
			return curr, true
		}
		visited[curr] = true
		var parentID sql.NullString
		err := db.QueryRow("SELECT parent_id FROM session WHERE id = ? LIMIT 1", curr).Scan(&parentID)
		if err != nil {
			return "", false
		}
		if !parentID.Valid || strings.TrimSpace(parentID.String) == "" {
			return curr, true
		}
		curr = strings.TrimSpace(parentID.String)
	}
}
