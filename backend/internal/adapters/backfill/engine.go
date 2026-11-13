package backfill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
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
	snapshotCutoff        time.Time
	discoverCodexRollouts func(codex.DiscoveryOptions) ([]string, error)
	freshCode             *domain.CodeFacts
	projectRoot           string
	budget                *sourceReadBudget
	generationPolicy      *domain.GenerationPolicy
}

// NewEngine creates a new BackfillEngine instance.
func NewEngine() *BackfillEngine {
	return &BackfillEngine{
		discoverCodexRollouts: codex.DiscoverRolloutPaths,
	}
}

// Execute runs the backfill and migration workflow across target workspaces.
func (e *BackfillEngine) Execute(opts BackfillOptions) (*BackfillReport, error) {
	run := *e
	if opts.SnapshotCutoff.IsZero() {
		opts.SnapshotCutoff = time.Now().UTC()
	}
	run.snapshotCutoff = opts.SnapshotCutoff
	e = &run
	cleanSource := strings.ToLower(strings.TrimSpace(opts.Source))
	if cleanSource == "" && opts.RecomputeDB {
		return nil, fmt.Errorf("--source is required for source backfill: select opencode, codex or both")
	}
	if cleanSource != "" && cleanSource != "opencode" && cleanSource != "codex" && cleanSource != "both" && cleanSource != "all" {
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
	targetDir, cfg, err := adapters.LoadWorkspaceProject(opts.Project)
	if err != nil {
		return nil, fmt.Errorf("loading required backfill configuration and workspaces: %w", errors.Join(err))
	}
	e.projectRoot = filepath.Dir(cfg.ConfigPath)

	opts.Project = targetDir

	report := &BackfillReport{
		Warnings: []string{},
		Results:  []TicketMigrationResult{},
	}

	// 1. Resolve targeted maintenance or discover an unfiltered batch.
	targetFiles, discoveryDiagnostics, err := discoverWorkspaceLedgers(cfg, opts.TicketID, opts.Workspace)
	if err != nil {
		return nil, err
	}

	report.Warnings = append([]string{}, discoveryDiagnostics...)
	report.Scanned = len(targetFiles)
	if len(targetFiles) == 0 {
		return report, nil
	}

	if opts.RecomputeDB {
		cfg, err = config.Load(e.projectRoot)
		if err != nil {
			return nil, fmt.Errorf("loading selected backfill source policy: %w", err)
		}
		if err := config.ValidateGenerationPolicy(cfg); err != nil {
			return nil, err
		}
		if err := cfg.Telemetry.Capture.ValidateBudgets(); err != nil {
			return nil, err
		}
		e.budget = &sourceReadBudget{sessions: cfg.Telemetry.Capture.MaxSessions, bytes: cfg.Telemetry.Capture.MaxSourceBytes}
		policy := cfg.Generation.Metrics
		e.generationPolicy = &policy
	}

	classificationCfg := &git.ClassificationConfig{BaseBranch: cfg.CodeMetrics.BaseBranch, Production: cfg.CodeMetrics.Production, Tests: cfg.CodeMetrics.Tests}
	if opts.RecomputeGit {
		fresh, err := git.ComputeCodeFactsAtRefInDir(e.projectRoot, opts.CodeRef, classificationCfg)
		if err != nil {
			for _, metricPath := range targetFiles {
				report.Results = append(report.Results, TicketMigrationResult{TicketID: func() string { id, _ := recordedTicket(metricPath); return id }(), WorkspacePath: filepath.Dir(metricPath), LedgerPath: metricPath, Error: fmt.Errorf("recomputing Git facts: %w", err)})
			}
			report.Warnings = append(report.Warnings, fmt.Sprintf("requested Git measurement failed before publication: %v", err))
			return report, nil
		}
		e.freshCode = &fresh
	}

	// Open only the explicitly selected stores. Unavailable selected paths fail
	// the requested extraction before any schema or ledger publication.
	var opencodeDB *opencode.OpenCodeDB
	wantsOpenCode := cleanSource == "opencode" || cleanSource == "both" || cleanSource == "all"
	wantsCodex := cleanSource == "codex" || cleanSource == "both" || cleanSource == "all"
	if opts.RecomputeDB && wantsOpenCode {
		if err := config.ValidateDatabasePolicy(cfg); err != nil {
			return nil, err
		}
		opts.DBPath, err = config.ResolveOpenCodePath(opts.DBPath, cfg.Telemetry.Sources.OpenCodeDatabasePath)
		if err != nil {
			return nil, err
		}
		opencodeDB, err = opencode.OpenDB(opts.DBPath, cfg.Database.BusyTimeoutMs)
		if err != nil {
			return nil, fmt.Errorf("opening selected OpenCode source for backfill: %w", err)
		}
		defer func() {
			if err := opencodeDB.Close(); err != nil {
				report.Warnings = append(report.Warnings, fmt.Sprintf("closing OpenCode source: %v", err))
			}
		}()
	}
	if opts.RecomputeDB && wantsCodex {
		opts.CodexHome, err = config.ResolveCodexHome(opts.CodexHome, cfg.Telemetry.Sources.CodexHome)
		if err != nil {
			return nil, err
		}
		opts.CodexHome, err = capture.ProjectDirectory(opts.CodexHome)
		if err != nil {
			return nil, fmt.Errorf("resolving selected Codex store: %w", err)
		}
		info, err := os.Stat(opts.CodexHome)
		if err != nil {
			return nil, fmt.Errorf("accessing selected Codex store %q: %w", opts.CodexHome, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("selected Codex store %q must be a directory", opts.CodexHome)
		}
	}

	gitUser := git.GetGitUserInDir(e.projectRoot)

	// 1.5. If schema migration requested, execute universal schema migration runner first
	if opts.MigrateSchema {
		migRunner := NewMigrationRunner(nil)
		migOpts := MigrationOptions{
			Project:       opts.Project,
			TicketID:      opts.TicketID,
			Workspace:     opts.Workspace,
			TargetVersion: domain.CurrentSchemaVersion,
			// Validate/report the existing migration chain without publishing
			// ahead of the integrated source/Git/summary candidate.
			DryRun:          true,
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

	nowISO := opts.SnapshotCutoff.UTC().Format(time.RFC3339Nano)

	// 2. Iterate each discovered workspace with per-ticket error isolation
	for _, metricPath := range targetFiles {
		lockCtx, lockCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		ticketResult := e.processSingleWorkspace(
			lockCtx, metricPath,
			opts,
			opencodeDB,
			gitUser,
			classificationCfg,
			nowISO,
			cfg,
		)

		lockCancel()

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

		if report.SchemaMigration != nil && !opts.DryRun {
			for i := range report.SchemaMigration.Results {
				schema := &report.SchemaMigration.Results[i]
				if schema.LedgerPath != ticketResult.LedgerPath {
					continue
				}
				if ticketResult.Error != nil {
					if schema.Status == "migrated (dry-run)" {
						report.SchemaMigration.Migrated--
					}
					schema.Status = "error"
					schema.Error = ticketResult.Error
					report.SchemaMigration.Errors = append(report.SchemaMigration.Errors, ticketResult.Error.Error())
				} else if schema.Status == "migrated (dry-run)" {
					schema.Status = "migrated"
				}
			}
		}
		report.Results = append(report.Results, ticketResult)
	}

	return report, nil
}

func (e *BackfillEngine) processSingleWorkspace(
	ctx context.Context,
	metricPath string,
	opts BackfillOptions,
	opencodeDB *opencode.OpenCodeDB,
	gitUser string,
	classificationCfg *git.ClassificationConfig,
	nowISO string,
	cfg domain.ProjectConfig,
) (result TicketMigrationResult) {
	workspaceDir := filepath.Dir(metricPath)
	workspaceName, _ := recordedTicket(metricPath)

	result = TicketMigrationResult{
		TicketID:      workspaceName,
		WorkspacePath: workspaceDir,
		LedgerPath:    metricPath,
	}

	// One target is owned at a time, before any baseline read and through the
	// ledger/summary writes. Batch callers never retain a lock while taking the
	// next, so they cannot invert the formatter's canonical multi-lock order.
	canonicalPath, summaryPath, err := backfillPublicationPaths(metricPath, opts.UpdateSummaries, cfg)
	if err != nil {
		result.Error = fmt.Errorf("resolving ledger publication: %w", err)
		return result
	}
	var owner io.Closer
	if opts.DryRun {
		directory, e := capture.ReadPublicationDirectory(canonicalPath)
		err = e
		if e == nil {
			owner = directory
		}
	} else {
		publisher, e := capture.AcquirePublication(ctx, canonicalPath, publicationProject(cfg, canonicalPath))
		err = e
		if e == nil {
			owner = publisher
		}
	}

	if err != nil {
		result.Error = fmt.Errorf("acquiring ledger publication ownership: %w", err)
		return result
	}
	defer func() { result.Error = errors.Join(result.Error, owner.Close()) }()
	metricPath = capture.PublicationPath(canonicalPath)
	result.LedgerPath = canonicalPath
	if err := adapters.CheckPublication(metricPath); err != nil {
		result.Error = err
		return result
	}

	// Read raw content to check file validity
	content, identity, err := readMaintenanceLedger(owner, metricPath, opts.TicketID, cfg.Ticket.CaseInsensitive)
	if err != nil {
		result.Error = fmt.Errorf("reading file: %w", err)
		return result
	}
	result.TicketID = identity

	sourceVersion, err := migration.NewEngine(nil).DetectVersion(content)
	if err != nil {
		result.Error = fmt.Errorf("detecting ledger schema: %w", err)
		return result
	}
	var current *currentLedgerPreparation
	var canonical domain.AgenticFactLedger
	if sourceVersion == domain.CurrentSchemaVersion {
		prepared, prepareErr := prepareCurrentLedger(content)
		if prepareErr != nil {
			result.Error = fmt.Errorf("validating current ledger before backfill: %w", prepareErr)
			return result
		}
		current = &prepared
		canonical = prepared.Ledger
	} else {
		canonical, err = migration.NewEngine(nil).Migrate(content, domain.CurrentSchemaVersion)
		if err != nil {
			result.Error = fmt.Errorf("normalizing ledger before backfill: %w", err)
			return result
		}
	}
	if current != nil && (current.Completed || current.FormerMetadataRemoved) && !opts.RecomputeGit && !opts.RecomputeDB {
		// An explicit metadata-only cleanup is not a source refresh or a new
		// observability snapshot. Preserve every other raw member unchanged.
		if canonical.TicketID != nil && strings.TrimSpace(*canonical.TicketID) != "" {
			result.TicketID = strings.TrimSpace(*canonical.TicketID)
		}
		status := domain.MigrationStatusCompleted
		if canonical.Migration != nil && !current.Completed {
			status = canonical.Migration.Status
			result.PendingSteps = append([]string{}, canonical.Migration.StepsPending...)
		}
		result.PreviousStatus = string(status)
		result.NewStatus = string(status)
		result.Notes = migrationAuditSummary(canonical.Migration)
		if opts.DryRun {
			return result
		}
		var tableMD string
		if summaryPath != "" {
			derivedLedger := domain.DeriveMetricsLedger(canonical, cfg)
			tableMD = adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
		}
		if err := publishBackfillArtifacts(metricPath, current.CleanupCandidate, summaryPath, tableMD); err != nil {
			result.Error = err
		}

		return result
	}
	dto := adapters.AgenticFactLedgerDTOFromDomain(canonical)
	if err := dto.Validate(); err != nil {
		result.Error = fmt.Errorf("validating tool session provenance: %w", err)
		return result
	}

	factLedger := dto.ToDomain()
	if info := factLedger.Migration; info != nil {
		completed := info.Status == domain.MigrationStatusCompleted
		incomplete := info.Status == domain.MigrationStatusPartial || info.Status == domain.MigrationStatusPending
		if completed && len(info.StepsPending) > 0 || incomplete && len(info.StepsPending) == 0 {
			result.Error = fmt.Errorf("inconsistent migration status %q with %d pending steps; preserved ledger for explicit recovery", info.Status, len(info.StepsPending))
			return result
		}
	}
	for i := range factLedger.Sessions {
		if factLedger.Sessions[i].Source == "" {
			factLedger.Sessions[i].Source = domain.SourceOpenCode
		}
	}

	cleanSource := strings.ToLower(strings.TrimSpace(opts.Source))
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

	needsEventBackfill := opts.RecomputeDB && (factLedger.SchemaVersion == nil || *factLedger.SchemaVersion < domain.CurrentSchemaVersion)

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
		if e.freshCode == nil {
			result.Error = fmt.Errorf("fresh Git measurement must be prepared for the selected target before publication")
			return result
		}
		newCode := *e.freshCode

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
	hasExistingDBFacts := len(factLedger.InvocationFacts()) > 0 || len(factLedger.Sessions) > 0
	openCodeComplete, codexComplete := false, false
	var sourceEvents domain.SourceEvents
	if opts.RecomputeDB && wantsOpenCode && (len(openCodeSessions) > 0 || len(factLedger.Sessions) == 0) {
		var db *sql.DB
		if opencodeDB != nil {
			db = opencodeDB.DB()
		}
		protected := make(map[string]bool, len(codexSessions)+len(otherSessions))
		for _, owner := range append(append([]domain.SessionFact{}, codexSessions...), otherSessions...) {
			protected[owner.ID] = true
		}
		workspacePaths := []string{workspaceDir}
		relative, err := filepath.Rel(filepath.Dir(cfg.ConfigPath), workspaceDir)
		if err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			workspacePaths = append(workspacePaths, filepath.ToSlash(relative))
		}
		recovery := recoverOpenCodeAt(db, openCodeSessions, gitUser, ticketID, workspacePaths, opts.AllowFallbacks, opts.SnapshotCutoff, e.budget, e.generationPolicy, protected)
		openCodeSessions = recovery.sessions
		sourceEvents = domain.MergeSourceEvents(sourceEvents, recovery.source)
		openCodeComplete = recovery.complete
		result.DBUpdated = recovery.updated
		result.Diagnostics = append(result.Diagnostics, recovery.diagnostics...)
		if recovery.complete {
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
			result.Diagnostics = append(result.Diagnostics, "skills/delegations retained from previous snapshot; OpenCode history incomplete or unverifiable; refreshed session trees may be newer")
		}
	}
	if opts.RecomputeDB && wantsCodex && (len(codexSessions) > 0 || len(factLedger.Sessions) == 0) {
		protected := make(map[string]bool, len(openCodeSessions)+len(otherSessions))
		for _, owner := range append(append([]domain.SessionFact{}, openCodeSessions...), otherSessions...) {
			protected[owner.ID] = true
		}
		recovery := e.recoverCodex(codexSessions, opts.CodexHome, gitUser, protected)
		codexSessions = recovery.sessions
		sourceEvents = domain.MergeSourceEvents(sourceEvents, recovery.source)
		codexComplete = recovery.complete
		result.DBUpdated = result.DBUpdated || recovery.updated
		result.Diagnostics = append(result.Diagnostics, recovery.diagnostics...)
	}
	selectedComplete := (!wantsOpenCode || openCodeComplete) && (!wantsCodex || codexComplete)
	// An empty selected harness does not need recovery when a different selected
	// harness owns the known sessions. Unknown ownerless collections remain unresolved.
	if wantsCodex && len(openCodeSessions) == 0 && len(codexSessions) > 0 {
		selectedComplete = (!wantsCodex || codexComplete)
	}
	if wantsOpenCode && len(codexSessions) == 0 && len(openCodeSessions) > 0 {
		selectedComplete = (!wantsOpenCode || openCodeComplete)
	}
	if opts.RecomputeDB && !hasTargetSessions {
		selectedComplete = false
		result.CollectionsRetained = true
		result.Diagnostics = append(result.Diagnostics, "No exact selected-source owner is recorded; previous snapshot preserved without heuristic source reassignment")
	}
	globalComplete := len(factLedger.Sessions) > 0 && len(otherSessions) == 0 && (len(openCodeSessions) == 0 || wantsOpenCode && openCodeComplete) && (len(codexSessions) == 0 || wantsCodex && codexComplete)

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
	hasOwnershipAudit := factLedger.Migration != nil && strings.Contains(factLedger.Migration.Notes, "Schema v6 migration dropped ownerless")
	if hasOwnershipAudit {
		if newMigration.Notes != "" {
			newMigration.Notes += "; "
		}
		if !strings.Contains(newMigration.Notes, factLedger.Migration.Notes) {
			newMigration.Notes += factLedger.Migration.Notes
		}
	}
	result.NewStatus = string(newMigration.Status)
	result.PendingSteps = newMigration.StepsPending
	result.Notes = migrationAuditSummary(&newMigration)
	if factLedger.Migration != nil {
		previousNotes := strings.TrimSpace(factLedger.Migration.Notes)
		if previousNotes != "" && !strings.Contains(result.Notes, previousNotes) {
			if result.Notes != "" {
				result.Notes += "; "
			}
			result.Notes += previousNotes
		}
	}
	if len(result.Diagnostics) > 0 {
		notes := append([]string(nil), result.Diagnostics...)
		if result.Notes != "" {
			notes = append([]string{result.Notes}, notes...)
		}
		result.Notes = strings.Join(notes, "; ")
	}

	version := domain.CurrentSchemaVersion
	factLedger.SchemaVersion = &version
	if newMigration.Status == domain.MigrationStatusCompleted && len(newMigration.StepsPending) == 0 {
		factLedger.Migration = nil
	} else {
		factLedger.Migration = &newMigration
	}
	factLedger.UpdatedAt = nowISO

	// 4. Prepare and validate the entire candidate before dry-run or publication.

	persistedLedger := factLedger
	if len(sourceEvents.CompleteSessionIDs) > 0 {
		if e.generationPolicy == nil {
			result.Error = fmt.Errorf("explicit generation policy is required for source backfill")
			return result
		}
		if e.generationPolicy.Anomalies {
			if err := config.ValidateAnomalyPolicy(cfg); err != nil {
				result.Error = err
				return result
			}
		}
		var err error
		persistedLedger, err = domain.PersistSourceAnomalySummariesWithPolicy(factLedger, sourceEvents, *e.generationPolicy, cfg.AnomalyThresholds)
		if err != nil {
			result.Error = fmt.Errorf("preparing session-owned facts: %w", err)
			return result
		}
	}
	dtoToSave := adapters.AgenticFactLedgerDTOFromDomain(persistedLedger)
	if err := dtoToSave.Validate(); err != nil {
		result.Error = fmt.Errorf("validating updated ledger: %w", err)
		return result
	}
	if opts.DryRun {
		return result
	}
	data, err := json.Marshal(dtoToSave)
	if err != nil {
		result.Error = fmt.Errorf("marshaling updated JSON: %w", err)
		return result
	}
	data, err = adapters.FormatLedgerJSON(data)
	if err != nil {
		result.Error = fmt.Errorf("formatting updated ledger: %w", err)
		return result
	}

	var tableMD string
	if summaryPath != "" {
		derivedLedger := domain.DeriveMetricsLedger(persistedLedger, cfg)
		tableMD = adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
	}
	if err := publishBackfillArtifacts(metricPath, data, summaryPath, tableMD); err != nil {
		result.Error = err
		return result
	}

	return result
}

// publishBackfillLedger replaces one already-owned ledger atomically.
func publishBackfillLedger(metricPath string, data []byte) error {
	metricPath = capture.PublicationPath(metricPath)
	tempPath := metricPath + ".tmp"
	if err := capture.WritePublicationFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("writing temporary ledger file: %w", err)
	}
	if err := capture.RenamePublicationFile(tempPath, metricPath); err != nil {
		if cleanupErr := capture.RemovePublicationFile(tempPath); cleanupErr != nil {
			return fmt.Errorf("atomic rename to %s: %w; cleaning temporary file: %v", metricPath, err, cleanupErr)
		}
		return fmt.Errorf("atomic rename to %s: %w", metricPath, err)
	}
	return nil
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

// replaceRecoveredOwners applies complete exact-owner snapshots while retaining
// all owners outside the successful refresh set, including unavailable children.
func replaceRecoveredOwners(stored *[]domain.SessionFact, refreshed []domain.SessionFact) {
	refreshed = domain.RetainUnrecordedCategories(*stored, domain.RetainIncompleteGroupSummaries(*stored, refreshed))
	replaced := make(map[string]bool, len(refreshed))
	for _, owner := range refreshed {
		replaced[owner.ID] = true
	}
	result := make([]domain.SessionFact, 0, len(*stored)+len(refreshed))
	for _, owner := range *stored {
		if !replaced[owner.ID] {
			result = append(result, owner)
		}
	}
	*stored = append(result, refreshed...)
}
