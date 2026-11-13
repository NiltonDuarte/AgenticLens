package backfill_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/domain"
)

func TestMaintenanceDryRunsDoNotRequireOrInitializeOperationalState(t *testing.T) {
	for _, command := range []string{"format-dry-run", "format-check", "migration", "backfill"} {
		for _, state := range []string{"absent", "malformed-unused-local", "unwritable-target"} {
			t.Run(command+"/"+state, func(t *testing.T) {
				project := migrationTestRoot(t)
				home := t.TempDir()
				t.Setenv("HOME", home)
				t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
				local := filepath.Join(project, ".agenticlens-config.local.yaml")
				if err := os.Remove(local); err != nil {
					t.Fatal(err)
				}
				switch state {
				case "malformed-unused-local":
					if err := os.WriteFile(local, []byte("agenticlens: [invalid unused operational state]\n"), 0600); err != nil {
						t.Fatal(err)
					}
				case "unwritable-target":
					if err := os.WriteFile(local, []byte("agenticlens:\n  state_dir: blocked-state\n"), 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(project, "blocked-state"), []byte("unrelated file"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				directory := filepath.Join(project, "records", "arbitrary-folder")
				if err := os.MkdirAll(directory, 0700); err != nil {
					t.Fatal(err)
				}
				body := []byte(`{"schema_version":6,"ticket_id":"TASK-9","sessions":[]}`)
				if command == "migration" {
					body = []byte(`{"schema_version":4,"ticket_id":"TASK-9","sessions":[]}`)
				}
				if command == "backfill" {
					body = []byte(`{"schema_version":6,"ticket_id":"TASK-9","sessions":[],"generation_policy":{"skills":false,"delegations":false,"anomalies":false}}`)
				}
				ledger := filepath.Join(directory, "agentic_metrics.json")
				if err := os.WriteFile(ledger, body, 0600); err != nil {
					t.Fatal(err)
				}
				before := maintenanceFixtureSnapshot(t, project)
				switch command {
				case "format-dry-run", "format-check":
					mode := adapters.LedgerFormatDryRun
					if command == "format-check" {
						mode = adapters.LedgerFormatCheck
					}
					results, err := adapters.FormatLedgerFiles(context.Background(), adapters.LedgerFormatOptions{Project: project, All: true, Mode: mode})
					if err != nil || len(results) != 1 || !results[0].Changed || results[0].Written {
						t.Fatalf("read-only format depended on private state: %+v %v", results, err)
					}
				case "migration":
					report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{Project: project, TicketID: "TASK-9", TargetVersion: domain.CurrentSchemaVersion, DryRun: true, Backup: true})
					if err != nil || report == nil || report.Scanned != 1 || report.Migrated != 1 || len(report.Errors) != 0 {
						t.Fatalf("read-only migration depended on private state: %+v %v", report, err)
					}
				case "backfill":
					report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: project, TicketID: "TASK-9", DryRun: true})
					if err != nil || report == nil || report.Scanned != 1 || report.Updated != 1 || len(report.Results) != 1 || report.Results[0].Error != nil {
						t.Fatalf("read-only backfill depended on private state: %+v %v", report, err)
					}
				}
				if after := maintenanceFixtureSnapshot(t, project); !reflect.DeepEqual(after, before) {
					t.Fatalf("dry run/check changed project state: before=%v after=%v", before, after)
				}
				if after, err := os.ReadFile(ledger); err != nil || !bytes.Equal(after, body) {
					t.Fatalf("dry run/check changed ledger bytes: %v", err)
				}
				entries, err := os.ReadDir(home)
				if err != nil || len(entries) != 0 {
					t.Fatalf("dry run/check initialized private home/cache: %v %v", entries, err)
				}
			})
		}
	}
}

func maintenanceFixtureSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			files[rel] = "directory"
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = string(body)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}
