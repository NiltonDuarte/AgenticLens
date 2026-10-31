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

		if ticketResult.RefreshIncomplete {
			report.RefreshIncomplete++
		}
		if ticketResult.CollectionsRetained {
			report.CollectionsRetained++
		}
		for _, diagnostic := range ticketResult.Diagnostics {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Ticket %s: %s", ticketResult.TicketID, diagnostic))
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

	if !opts.Force && isCompleted && !opts.RecomputeGit && !opts.RecomputeDB && !cleanupCompletedMigration && factLedger.Migration == nil {
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

	// 2. Recover complete owner trees before publishing any replacement snapshot.
	dbAttempted := opts.RecomputeDB && hasTargetSessions
	hasExistingDBFacts := len(factLedger.InvocationFacts()) > 0 || len(factLedger.EventFacts()) > 0
	openCodeComplete, codexComplete := false, false
	if opts.RecomputeDB && wantsOpenCode && (len(openCodeSessions) > 0 || len(factLedger.Sessions) == 0 || len(factLedger.Skills) > 0 || len(factLedger.TaskLinks) > 0) {
		var db *sql.DB
		if opencodeDB != nil {
			db = opencodeDB.DB()
		}
		recovery := recoverOpenCode(db, openCodeSessions, gitUser, ticketID, workspaceDir, opts.AllowFallbacks)
		openCodeSessions = recovery.sessions
		openCodeComplete = recovery.complete
		result.DBUpdated = recovery.updated
		result.Diagnostics = append(result.Diagnostics, recovery.diagnostics...)
		if recovery.complete {
			factLedger.Skills, factLedger.TaskLinks = opencode.ExtractSkillTaskFacts(openCodeSessions)
			if factLedger.Skills == nil {
				factLedger.Skills = []domain.SkillFact{}
			}
			if factLedger.TaskLinks == nil {
				factLedger.TaskLinks = []domain.TaskLinkFact{}
			}
			if opts.RecomputeGit && db != nil {
				var ids []string
				for _, session := range openCodeSessions {
					ids = append(ids, session.ID)
				}
				turns, err := opencode.ExtractSteeringTurns(db, ids)
				if err != nil {
					result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("steering count unavailable: %v", err))
				} else {
					factLedger.Code.SDLCIterations = turns
				}
			}
		} else {
			result.CollectionsRetained = true
			result.Diagnostics = append(result.Diagnostics, "skills/task_links retained from previous snapshot; OpenCode history incomplete or unverifiable; refreshed session trees may be newer")
		}
	}
	if opts.RecomputeDB && wantsCodex && (len(codexSessions) > 0 || len(factLedger.Sessions) == 0) {
		codexHome := strings.TrimSpace(opts.CodexHome)
		if codexHome == "" {
			codexHome = strings.TrimSpace(os.Getenv("CODEX_HOME"))
		}
		if codexHome == "" {
			if home, err := os.UserHomeDir(); err == nil {
				codexHome = filepath.Join(home, ".codex")
			}
		}
		recovery := e.recoverCodex(codexSessions, codexHome, gitUser)
		codexSessions = recovery.sessions
		codexComplete = recovery.complete
		result.DBUpdated = result.DBUpdated || recovery.updated
		result.Diagnostics = append(result.Diagnostics, recovery.diagnostics...)
	}
	selectedComplete := (!wantsOpenCode || openCodeComplete) && (!wantsCodex || codexComplete)
	// An empty selected harness does not need recovery when a different selected
	// harness owns the known sessions. Unknown ownerless collections remain unresolved.
	if len(openCodeSessions) == 0 && len(codexSessions) > 0 && len(factLedger.Skills) == 0 && len(factLedger.TaskLinks) == 0 {
		selectedComplete = (!wantsCodex || codexComplete)
	}
	if len(codexSessions) == 0 && len(openCodeSessions) > 0 {
		selectedComplete = (!wantsOpenCode || openCodeComplete)
	}
	globalComplete := len(factLedger.Sessions) > 0 && len(otherSessions) == 0 &&
		(len(openCodeSessions) == 0 || (wantsOpenCode && openCodeComplete)) &&
		(len(codexSessions) == 0 || (wantsCodex && codexComplete))
	if len(openCodeSessions) == 0 && (len(factLedger.Skills) > 0 || len(factLedger.TaskLinks) > 0) {
		globalComplete = false
	}
	dbSessionFound := selectedComplete
	if factLedger.Migration != nil {
		dbSessionFound = globalComplete
	}
	result.DBFound = dbSessionFound && opts.RecomputeDB
	result.RefreshIncomplete = opts.RecomputeDB && (!selectedComplete || len(result.Diagnostics) > 0)
	if opts.RecomputeDB && !isCompleted && factLedger.Migration != nil && !globalComplete {
		result.Diagnostics = append(result.Diagnostics, "global source recovery remains pending: not every known stored owner/child across required sources was recovered")
		result.RefreshIncomplete = true
	}

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
	if opts.RecomputeDB && globalComplete && factLedger.Migration != nil {
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
		RecomputeDB:        opts.RecomputeDB && hasTargetSessions && !isCompleted,
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
	if len(result.Diagnostics) > 0 {
		notes := append([]string(nil), result.Diagnostics...)
		if result.Notes != "" {
			notes = append([]string{result.Notes}, notes...)
		}
		result.Notes = strings.Join(notes, "; ")
	}

	version := domain.CurrentSchemaVersion
	factLedger.SchemaVersion = &version
	if newMigration.Status == domain.MigrationStatusCompleted {
		factLedger.Migration = nil
	} else {
		factLedger.Migration = &newMigration
	}
	factLedger.UpdatedAt = nowISO

	// 4. Prepare and validate the entire candidate before dry-run or publication.

	persistedLedger, err := domain.PersistAnomalySummaries(factLedger)
	if err != nil {
		result.Error = fmt.Errorf("preparing session-owned facts: %w", err)
		return result
	}
	dtoToSave := adapters.AgenticFactLedgerDTOFromDomain(persistedLedger)
	if err := dtoToSave.Validate(); err != nil {
		result.Error = fmt.Errorf("validating updated ledger: %w", err)
		return result
	}
	if opts.DryRun {
		return result
	}
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
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil {
			result.Error = fmt.Errorf("atomic rename to %s: %w; cleaning temporary file: %v", metricPath, err, cleanupErr)
			return result
		}
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

func resolveCodexSessionUsers(fresh []domain.SessionFact, existing []domain.SessionFact, gitUser string) {
	existingUsers := make(map[string]string, len(existing))
	for _, session := range existing {
		if git.IsKnownUser(session.User) {
			existingUsers[session.ID] = strings.TrimSpace(session.User)
		}
	}
	for i := range fresh {
		fresh[i].User = git.ResolveSessionUser(fresh[i].User, existingUsers[fresh[i].ID], gitUser)
	}
}

func replacementSessionIDs(stored []domain.SessionFact, requestedRoot string, refreshed []domain.SessionFact) map[string]struct{} {
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
		for _, session := range stored {
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
	return removeIDs
}

func replaceCodexTree(stored *[]domain.SessionFact, requestedRoot string, refreshed []domain.SessionFact) {
	removeIDs := replacementSessionIDs(*stored, requestedRoot, refreshed)

	result := make([]domain.SessionFact, 0, len(*stored)-len(removeIDs)+len(refreshed))
	for _, session := range *stored {
		if _, remove := removeIDs[session.ID]; !remove {
			result = append(result, session)
		}
	}
	result = append(result, refreshed...)
	*stored = result
}
