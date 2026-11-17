package backfill_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/domain"
)

func TestMaintenanceSelectionDoesNotInferUniquenessFromIncompleteDiscovery(t *testing.T) {
	for _, command := range []string{"backfill", "migration", "backfill-with-migration"} {
		t.Run(command, func(t *testing.T) {
			project := migrationTestRoot(t)
			if err := os.Remove(filepath.Join(project, ".agenticlens-config.local.yaml")); err != nil {
				t.Fatal(err)
			}
			privateHome := t.TempDir()
			t.Setenv("HOME", privateHome)
			t.Setenv("XDG_CACHE_HOME", filepath.Join(privateHome, "cache"))
			policyPath := filepath.Join(project, ".agenticlens-config.yaml")
			priorPolicy, err := os.ReadFile(policyPath)
			if err != nil {
				t.Fatal(err)
			}
			policy := bytes.Replace(priorPolicy, []byte("max_depth: 8\n        max_directories: 300"), []byte("max_depth: 1\n        max_directories: 1"), 1)
			if bytes.Equal(policy, priorPolicy) {
				t.Fatal("fixture did not configure the bounded discovery policy")
			}
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), policy, 0600); err != nil {
				t.Fatal(err)
			}
			hidden := filepath.Join(project, "unvisited-copy")
			if err := os.Mkdir(hidden, 0700); err != nil {
				t.Fatal(err)
			}
			body := []byte(`{"schema_version":6,"ticket_id":"TASK-9","sessions":[],"generation_policy":{"skills":false,"delegations":false,"anomalies":false}}`)
			if command != "backfill" {
				body = []byte(`{"schema_version":4,"ticket_id":"TASK-9","sessions":[]}`)
			}
			for _, home := range []string{project, hidden} {
				if err := os.WriteFile(filepath.Join(home, "agentic_metrics.json"), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			for _, selection := range []struct {
				name, ticket, home string
				rejected           bool
			}{
				{"implicit-ticket", "TASK-9", "", true},
				{"explicit-observed-home", "TASK-9", project, false},
				{"explicit-unvisited-home", "TASK-9", hidden, false},
				{"explicit-unvisited-home-without-ticket", "", hidden, false},
				{"explicit-identity-conflict", "OTHER-10", hidden, true},
				{"unfiltered-partial-batch", "", "", false},
			} {
				t.Run(selection.name, func(t *testing.T) {
					before := maintenanceFixtureSnapshot(t, project)
					var runErr error
					var scanned int
					var paths, tickets, diagnostics []string
					if command != "migration" {
						report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: project, TicketID: selection.ticket, Workspace: selection.home, DryRun: true, MigrateSchema: command == "backfill-with-migration"})
						runErr = err
						if report != nil {
							scanned = report.Scanned
							diagnostics = report.Warnings
							for _, result := range report.Results {
								if result.Error != nil {
									t.Fatalf("selected dry run failed: %v", result.Error)
								}
								paths = append(paths, result.WorkspacePath)
								tickets = append(tickets, result.TicketID)
							}
							if command == "backfill-with-migration" {
								migration := report.SchemaMigration
								if migration == nil || migration.Scanned != 1 || migration.Migrated != 1 || len(migration.Results) != 1 || migration.Results[0].Error != nil || len(paths) != 1 || migration.Results[0].WorkspacePath != paths[0] || migration.Results[0].TicketID != "TASK-9" {
									t.Fatalf("nested migration lost explicit selection: %+v", migration)
								}
								if strings.Contains(strings.Join(report.Warnings, " "), "Schema migration failed") {
									t.Fatalf("nested migration silently retried implicit selection: %v", report.Warnings)
								}
							}
						}
					} else {
						report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{Project: project, TicketID: selection.ticket, Workspace: selection.home, DryRun: true, Backup: true, TargetVersion: domain.CurrentSchemaVersion})
						runErr = err
						if report != nil {
							scanned = report.Scanned
							diagnostics = report.Errors
							for _, result := range report.Results {
								if result.Error != nil {
									t.Fatalf("selected historical dry run failed: %v", result.Error)
								}
								paths = append(paths, result.WorkspacePath)
								tickets = append(tickets, result.TicketID)
							}
						}
					}
					if selection.rejected {
						var resolution *adapters.WorkspaceResolutionError
						want := "unavailable"
						if selection.home != "" {
							want = "identity_conflict"
						}
						if !errors.As(runErr, &resolution) || resolution.Outcome != want {
							t.Fatalf("targeted maintenance silently selected or reached wrong failure: want %s, got %v", want, runErr)
						}
					} else {
						wantHome := selection.home
						if wantHome == "" {
							wantHome = project
						}
						if runErr != nil || scanned != 1 || len(paths) != 1 || paths[0] != wantHome || len(tickets) != 1 || tickets[0] != "TASK-9" {
							t.Fatalf("selection lost recorded home/identity: scanned=%d paths=%v tickets=%v err=%v", scanned, paths, tickets, runErr)
						}
						if selection.name == "unfiltered-partial-batch" && !strings.Contains(strings.Join(diagnostics, " "), "budget") {
							t.Fatalf("partial batch lost discovery warning: %v", diagnostics)
						}
					}
					if after := maintenanceFixtureSnapshot(t, project); !reflect.DeepEqual(after, before) {
						t.Fatalf("dry run changed ledger, backup, policy or operational state: before=%v after=%v", before, after)
					}
					if entries, err := os.ReadDir(privateHome); err != nil || len(entries) != 0 {
						t.Fatalf("dry run initialized private state: %v %v", entries, err)
					}
				})
			}
		})
	}
}

// A conclusive missing-ticket result remains an empty batch rather than a
// resolution failure. It must also remain a no-op for the existing other home.
func TestBackfillCompleteDiscoveryMissingTicketIsNoOp(t *testing.T) {
	project := migrationTestRoot(t)
	if err := os.Remove(filepath.Join(project, ".agenticlens-config.local.yaml")); err != nil {
		t.Fatal(err)
	}
	privateHome := t.TempDir()
	t.Setenv("HOME", privateHome)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(privateHome, "cache"))
	body := []byte(`{"schema_version":6,"ticket_id":"TASK-9","sessions":[],"generation_policy":{"skills":false,"delegations":false,"anomalies":false}}`)
	if err := os.WriteFile(filepath.Join(project, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	before := maintenanceFixtureSnapshot(t, project)
	report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: project, TicketID: "DOES-NOT-EXIST"})
	if err != nil || report == nil || report.Scanned != 0 || report.Updated != 0 || len(report.Results) != 0 || len(report.Warnings) != 0 {
		t.Fatalf("complete absence is not an empty batch: %+v %v", report, err)
	}
	if after := maintenanceFixtureSnapshot(t, project); !reflect.DeepEqual(after, before) {
		t.Fatalf("missing ticket changed the existing home: before=%v after=%v", before, after)
	}
	if entries, err := os.ReadDir(privateHome); err != nil || len(entries) != 0 {
		t.Fatalf("empty batch initialized private state: %v %v", entries, err)
	}
}
