package backfill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

// MigrationRunner discovers workspace ledgers, detects schema versions,
// invokes the pure domain migration engine, manages backups, and persists upgrades atomically.
type MigrationRunner struct {
	engine *migration.Engine
}

// NewMigrationRunner creates a MigrationRunner instance.
// If engine is nil, the default domain migration engine is used.
func NewMigrationRunner(engine *migration.Engine) *MigrationRunner {
	if engine == nil {
		engine = migration.NewEngine(nil)
	}
	return &MigrationRunner{engine: engine}
}

// Run executes batch schema migration across target workspace ledgers.
func (r *MigrationRunner) Run(opts MigrationOptions) (*SchemaMigrationReport, error) {
	targetVer := opts.TargetVersion
	if targetVer <= 0 {
		targetVer = domain.CurrentSchemaVersion
	}
	if targetVer != domain.CurrentSchemaVersion {
		return nil, fmt.Errorf("migration publication requires current target version %d; intermediate transformations are migration-engine internals", domain.CurrentSchemaVersion)
	}

	_, cfg, err := adapters.LoadWorkspaceProject(opts.Project)
	if err != nil {
		return nil, fmt.Errorf("loading required migration target configuration and workspaces: %w", errors.Join(err))
	}

	targetFiles, discoveryDiagnostics, err := discoverWorkspaceLedgers(cfg, opts.TicketID, opts.Workspace)
	if err != nil {
		return nil, err
	}

	report := &SchemaMigrationReport{
		Scanned: len(targetFiles),
		Errors:  append([]string{}, discoveryDiagnostics...),
		Results: []TicketSchemaMigrationResult{},
	}

	if len(targetFiles) == 0 {
		return report, nil
	}

	for _, metricPath := range targetFiles {
		// Release each target before the next acquisition; no batch lock order
		// can conflict with format's sorted canonical ownership.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		res, issues := r.migrateWorkspace(ctx, metricPath, opts, targetVer, cfg)
		cancel()
		report.Errors = append(report.Errors, issues...)
		report.Results = append(report.Results, res)
		switch res.Status {
		case "up-to-date":
			report.UpToDate++
		case "migrated", "migrated (dry-run)":
			report.Migrated++
		}
	}

	return report, nil
}

// migrateWorkspace owns the canonical output before observing its baseline.
// Generation, format and backfill take this same lock through publication.
func (r *MigrationRunner) migrateWorkspace(ctx context.Context, metricPath string, opts MigrationOptions, targetVer int, cfg domain.ProjectConfig) (res TicketSchemaMigrationResult, issues []string) {

	wsDir := filepath.Dir(metricPath)
	wsName, _ := recordedTicket(metricPath)

	res = TicketSchemaMigrationResult{
		TicketID:      wsName,
		WorkspacePath: wsDir,
		LedgerPath:    metricPath,
		TargetVersion: targetVer,
	}

	canonicalPath, summaryPath, err := backfillPublicationPaths(metricPath, opts.UpdateSummaries, cfg)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("resolving ledger publication: %w", err)
		return res, []string{res.Error.Error()}
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
		res.Status = "error"
		res.Error = fmt.Errorf("acquiring ledger publication ownership: %w", err)
		return res, []string{res.Error.Error()}
	}
	defer func() {
		if err := owner.Close(); err != nil {
			res.Status = "error"
			res.Error = errors.Join(res.Error, err)
			issues = append(issues, fmt.Sprintf("Workspace %s: publication unlock: %v", metricPath, err))
		}
	}()
	metricPath = capture.PublicationPath(canonicalPath)
	res.LedgerPath = canonicalPath
	if err := adapters.CheckPublication(metricPath); err != nil {
		res.Status = "error"
		res.Error = err
		return res, []string{err.Error()}
	}

	content, identity, err := readMaintenanceLedger(owner, metricPath, opts.TicketID, cfg.Ticket.CaseInsensitive)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("reading file: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	res.TicketID = identity

	sourceVer, err := r.engine.DetectVersion(content)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("detecting schema version: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	res.SourceVersion = sourceVer
	if sourceVer > domain.CurrentSchemaVersion {
		res.Status = "error"
		res.Error = fmt.Errorf("%w: schema %d is newer than supported schema %d", migration.ErrUnsupportedVersion, sourceVer, domain.CurrentSchemaVersion)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	hasLegacyFacts := hasLegacyFactCollections(content)
	missingSource := hasMissingSessionSource(content)
	needsSanitization := sourceVer >= 4 && sourceVer <= domain.CurrentSchemaVersion && (needsEventSanitization(content) || missingSource)
	var current *currentLedgerPreparation
	if sourceVer == domain.CurrentSchemaVersion {
		prepared, err := prepareCurrentLedger(content)
		if err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("validating current schema ledger: %w", err)
			return res, []string{res.Error.Error()}
		}
		current = &prepared
	}
	cleanupCompleted := current != nil && current.Completed
	cleanupFormer := current != nil && current.FormerMetadataRemoved
	cleanupCurrent := cleanupCompleted || cleanupFormer
	if sourceVer >= targetVer && !cleanupCurrent && !(sourceVer >= 4 && sourceVer <= domain.CurrentSchemaVersion && (hasLegacyFacts || needsSanitization)) {
		res.Status = "up-to-date"
		res.DiffSummary = fmt.Sprintf("Schema version %d is already up-to-date", sourceVer)
		return res, issues
	}

	var migrated domain.AgenticFactLedger
	if current != nil {
		migrated = current.Ledger
	} else {
		migrated, err = r.engine.Migrate(content, targetVer)
	}
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("migrating schema: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}

	if migrated.TicketID != nil && strings.TrimSpace(*migrated.TicketID) != "" {
		res.TicketID = strings.TrimSpace(*migrated.TicketID)
	}
	res.DiffSummary = fmt.Sprintf("Upgraded schema v%d -> v%d", sourceVer, targetVer)
	if cleanupCompleted {
		res.DiffSummary = "Removed completed migration metadata"
	}
	if cleanupFormer {
		if cleanupCompleted {
			res.DiffSummary += " and former generation metadata"
		} else {
			res.DiffSummary = "Removed former generation metadata"
		}
	}
	if audit := migrationAuditSummary(migrated.Migration); audit != "" {
		res.DiffSummary += "; " + audit
	}
	if migrated.Migration != nil && migrated.Migration.Status == domain.MigrationStatusCompleted && len(migrated.Migration.StepsPending) == 0 {
		migrated.Migration = nil
	}

	// Validate and prepare the complete candidate before dry run or backup. Current
	// schema cleanup deletes only the metadata member, preserving every other raw value.
	dto := adapters.AgenticFactLedgerDTOFromDomain(migrated)
	if err := dto.Validate(); err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("validating migrated schema: %w", err)
		return res, []string{res.Error.Error()}
	}
	var data []byte
	if cleanupCurrent {
		data = current.CleanupCandidate
	} else {
		data, err = json.Marshal(dto)
	}
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("marshaling migrated JSON: %w", err)
		return res, []string{res.Error.Error()}
	}
	data, err = adapters.FormatLedgerJSON(data)
	if err != nil {
		res.Status = "error"
		res.Error = err
		return res, []string{err.Error()}
	}

	if opts.DryRun {
		res.Status = "migrated (dry-run)"
		return res, issues
	}

	if opts.Backup {
		backupPath := metricPath + ".bak"
		if err := capture.WritePublicationFile(backupPath, content, 0644); err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("creating backup file %s: %w", backupPath, err)
			issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			return res, issues
		}
		res.BackupCreated = true
		res.BackupPath = canonicalPath + ".bak"
	}

	var tableMD string
	if summaryPath != "" {
		derivedLedger := domain.DeriveMetricsLedger(migrated, cfg)
		tableMD = adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
	}
	if err := publishBackfillArtifacts(metricPath, data, summaryPath, tableMD); err != nil {
		res.Status = "error"
		res.Error = err
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, err))
		return res, issues
	}

	res.Status = "migrated"
	return res, issues
}

func needsEventSanitization(content []byte) bool {
	var input struct {
		Events   json.RawMessage `json:"events"`
		Partial  json.RawMessage `json:"events_are_partial"`
		Sessions []struct {
			Events  json.RawMessage `json:"events"`
			Partial json.RawMessage `json:"events_are_partial"`
		} `json:"sessions"`
	}
	if json.Unmarshal(content, &input) != nil {
		return false
	}
	if len(input.Events) > 0 || len(input.Partial) > 0 {
		return true
	}
	for _, owner := range input.Sessions {
		if len(owner.Events) > 0 || len(owner.Partial) > 0 {
			return true
		}
	}
	return false
}

func hasLegacyFactCollections(content []byte) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(content, &fields) != nil {
		return false
	}
	for _, key := range []string{"invocations", "tools", "events"} {
		if _, ok := fields[key]; ok {
			return true
		}
	}
	return false
}

func hasMissingSessionSource(content []byte) bool {
	var raw struct {
		Sessions []struct {
			Source string `json:"source"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return false
	}
	if len(raw.Sessions) == 0 {
		return false
	}
	for _, s := range raw.Sessions {
		if strings.TrimSpace(s.Source) == "" {
			return true
		}
	}
	return false
}
