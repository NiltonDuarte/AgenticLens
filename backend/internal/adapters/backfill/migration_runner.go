package backfill

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/adapters"
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

	targetDir := adapters.ResolveWorkspacesDir(opts.WorkspacesDir)
	fi, err := os.Stat(targetDir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("workspaces directory not found: %s", targetDir)
	}

	cfg, err := config.Load(targetDir)
	if err != nil {
		cfg, _ = config.Load("")
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
		wsDir := filepath.Dir(metricPath)
		wsName := filepath.Base(wsDir)

		res := TicketSchemaMigrationResult{
			TicketID:      wsName,
			WorkspacePath: wsDir,
			TargetVersion: targetVer,
		}

		content, err := os.ReadFile(metricPath)
		if err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("reading file: %w", err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}

		sourceVer, err := r.engine.DetectVersion(content)
		if err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("detecting schema version: %w", err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}
		res.SourceVersion = sourceVer

		if sourceVer >= targetVer {
			res.Status = "up-to-date"
			res.DiffSummary = fmt.Sprintf("Schema version %d is already up-to-date", sourceVer)
			report.UpToDate++
			report.Results = append(report.Results, res)
			continue
		}

		migrated, err := r.engine.Migrate(content, targetVer)
		if err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("migrating schema: %w", err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}

		if migrated.TicketID != nil && strings.TrimSpace(*migrated.TicketID) != "" {
			res.TicketID = strings.TrimSpace(*migrated.TicketID)
		}
		res.DiffSummary = fmt.Sprintf("Upgraded schema v%d -> v%d", sourceVer, targetVer)

		if opts.DryRun {
			res.Status = "migrated (dry-run)"
			report.Migrated++
			report.Results = append(report.Results, res)
			continue
		}

		if opts.Backup {
			backupPath := metricPath + ".bak"
			if err := os.WriteFile(backupPath, content, 0644); err != nil {
				res.Status = "error"
				res.Error = fmt.Errorf("creating backup file %s: %w", backupPath, err)
				report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
				report.Results = append(report.Results, res)
				continue
			}
			res.BackupCreated = true
		}

		dto := adapters.AgenticFactLedgerDTOFromDomain(migrated)
		data, err := json.MarshalIndent(dto, "", "  ")
		if err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("marshaling migrated JSON: %w", err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}
		data = append(data, '\n')

		tempPath := metricPath + ".tmp"
		if err := os.WriteFile(tempPath, data, 0644); err != nil {
			res.Status = "error"
			res.Error = fmt.Errorf("writing temporary file %s: %w", tempPath, err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}

		if err := os.Rename(tempPath, metricPath); err != nil {
			_ = os.Remove(tempPath)
			res.Status = "error"
			res.Error = fmt.Errorf("atomic rename to %s: %w", metricPath, err)
			report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: %v", metricPath, res.Error))
			report.Results = append(report.Results, res)
			continue
		}

		if opts.UpdateSummaries {
			summaryPath := filepath.Join(wsDir, "SUMMARY.md")
			derivedLedger := domain.DeriveMetricsLedger(migrated, cfg)
			tableMD := adapters.FormatMarkdownMetricsTable(derivedLedger, cfg.RoleOrder)
			if err := adapters.UpdateSummaryFile(summaryPath, tableMD); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("Workspace %s: updating summary: %v", metricPath, err))
			}
		}

		res.Status = "migrated"
		report.Migrated++
		report.Results = append(report.Results, res)
	}

	return report, nil
}
