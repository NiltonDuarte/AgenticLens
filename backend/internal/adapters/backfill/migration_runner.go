package backfill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
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

	targetDir, err := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	if err != nil {
		return nil, fmt.Errorf("resolving workspaces directory: %w", err)
	}
	fi, err := os.Stat(targetDir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("workspaces directory not found: %s", targetDir)
	}

	cfg, err := config.Load(targetDir)
	if errors.Is(err, config.ErrConfigNotFound) {
		cfg, err = config.Load("")
	}
	if err != nil && !errors.Is(err, config.ErrConfigNotFound) {
		return nil, fmt.Errorf("loading backfill configuration: %w", err)
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
		return nil, fmt.Errorf("walking workspaces directory %s: %w", targetDir, err)
	}

	sort.Strings(metricFiles)

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

	report := &SchemaMigrationReport{
		Scanned: len(targetFiles),
		Errors:  []string{},
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
	wsName := filepath.Base(wsDir)

	res = TicketSchemaMigrationResult{
		TicketID:      wsName,
		WorkspacePath: wsDir,
		TargetVersion: targetVer,
	}

	canonicalPath, err := capture.ProjectDirectory(metricPath)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("resolving ledger publication: %w", err)
		return res, []string{res.Error.Error()}
	}
	owner, err := capture.AcquirePublication(ctx, canonicalPath)
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
	metricPath = canonicalPath
	if err := adapters.CheckPublication(metricPath); err != nil {
		res.Status = "error"
		res.Error = err
		return res, []string{err.Error()}
	}

	content, err := os.ReadFile(metricPath)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("reading file: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}

	sourceVer, err := r.engine.DetectVersion(content)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("detecting schema version: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	res.SourceVersion = sourceVer
	hasLegacyFacts := hasLegacyFactCollections(content)
	missingSource := hasMissingSessionSource(content)
	needsSanitization := sourceVer >= 4 && sourceVer <= domain.CurrentSchemaVersion && (needsEventSanitization(content) || missingSource)
	if sourceVer == domain.CurrentSchemaVersion && !hasLegacyFacts {
		if err := validateCurrentLedger(content); err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("validating current schema ledger: %w", err)
			issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			return res, issues
		}
	}
	if sourceVer >= targetVer && !(sourceVer >= 4 && sourceVer <= domain.CurrentSchemaVersion && (hasLegacyFacts || needsSanitization)) {
		res.Status = "up-to-date"
		res.DiffSummary = fmt.Sprintf("Schema version %d is already up-to-date", sourceVer)
		return res, issues
	}

	migrated, err := r.engine.Migrate(content, targetVer)
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
	if migrated.Migration != nil && strings.TrimSpace(migrated.Migration.Notes) != "" {
		res.DiffSummary += "; " + strings.TrimSpace(migrated.Migration.Notes)
	}

	if opts.DryRun {
		res.Status = "migrated (dry-run)"
		return res, issues
	}

	if opts.Backup {
		backupPath := metricPath + ".bak"
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("creating backup file %s: %w", backupPath, err)
			issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			return res, issues
		}
		res.BackupCreated = true
	}

	dto := adapters.AgenticFactLedgerDTOFromDomain(migrated)
	if err := dto.Validate(); err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("validating migrated schema: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	data, err := json.Marshal(dto)
	if err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("marshaling migrated JSON: %w", err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}
	data, err = adapters.FormatLedgerJSON(data)
	if err != nil {
		res.Status = "error"
		res.Error = err
		issues = append(issues, err.Error())
		return res, issues
	}

	tempPath := metricPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		res.Status = "error"
		res.Error = fmt.Errorf("writing temporary file %s: %w", tempPath, err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}

	if err := os.Rename(tempPath, metricPath); err != nil {
		_ = os.Remove(tempPath)
		res.Status = "error"
		res.Error = fmt.Errorf("atomic rename to %s: %w", metricPath, err)
		issues = append(issues, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
		return res, issues
	}

	if opts.UpdateSummaries {
		summaryPath := filepath.Join(wsDir, "SUMMARY.md")
		derivedLedger := domain.DeriveMetricsLedger(migrated, cfg)
		tableMD := adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
		if err := adapters.UpdateSummaryFile(summaryPath, tableMD); err != nil {
			issues = append(issues, fmt.Sprintf("Workspace %s: updating summary: %v", metricPath, err))
		}
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

func validateCurrentLedger(content []byte) error {
	// Pre-iteration v4 ledgers may contain payload keys that this normalization removes.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(content, &raw); err != nil {
		return err
	}
	if sessions, ok := raw["sessions"]; ok {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(sessions, &items); err == nil {
			for i := range items {
				if events, exists := items[i]["events"]; exists {
					var rows []map[string]json.RawMessage
					if json.Unmarshal(events, &rows) == nil {
						for j := range rows {
							for _, key := range []string{"error_message", "prompt_content", "reasoning_content", "tool_input", "tool_output"} {
								delete(rows[j], key)
							}
						}
						items[i]["events"], _ = json.Marshal(rows)
					}
				}
				if _, hasSrc := items[i]["source"]; !hasSrc {
					items[i]["source"], _ = json.Marshal(domain.SourceOpenCode)
				}
				if _, hasUser := items[i]["user"]; !hasUser {
					items[i]["user"], _ = json.Marshal("unknown")
				}
			}
			raw["sessions"], _ = json.Marshal(items)
		}
	}
	cleaned, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(cleaned, &dto); err != nil {
		return err
	}
	return dto.Validate()
}
