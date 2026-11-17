package backfill_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/domain"
)

func TestExplicitCurrentMetadataCleanupPreservesFactsAndPendingState(t *testing.T) {
	for _, command := range []string{"migrate", "backfill"} {
		for _, pending := range []bool{false, true} {
			for _, dryRun := range []bool{false, true} {
				name := command
				if pending {
					name += "/pending"
				} else {
					name += "/complete"
				}
				if dryRun {
					name += "/dry-run"
				}
				t.Run(name, func(t *testing.T) {
					root := migrationTestRoot(t)
					workspace := filepath.Join(root, "records", "CASE-01")
					if err := os.MkdirAll(workspace, 0700); err != nil {
						t.Fatal(err)
					}
					migration := ""
					if pending {
						migration = `,"migration":{"schema_version":6,"status":"pending","steps_completed":[],"steps_pending":["source_backfill"],"notes":"Unavailable owners remain recorded"}`
					}
					prefix := `{"schema_version":6,"ticket_id":"CASE-01","updated_at":"recorded-time"`
					owner := `{"id":"portable-owner","source":"codex","user":"recorded-user","invocations":[],"tools":[],"skills":[],"delegations":[]`
					original := []byte(prefix + `,"generation_policy":{"skills":true,"delegations":true,"anomalies":true},"sessions":[` + owner + `,"anomaly_availability":"unavailable"}]` + migration + `}`)
					expected := []byte(prefix + `,"sessions":[` + owner + `}]` + migration + `}`)
					path := filepath.Join(workspace, "agentic_metrics.json")
					if err := os.WriteFile(path, original, 0600); err != nil {
						t.Fatal(err)
					}
					if err := adapters.ValidateRuntimeFactLedger(original); err == nil {
						t.Fatal("ordinary validation accepted former metadata")
					}
					if command == "migrate" {
						report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{Project: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion, Backup: true, DryRun: dryRun})
						if err != nil || report == nil || report.Migrated != 1 || len(report.Errors) != 0 {
							t.Fatalf("explicit cleanup did not run: %+v, %v", report, err)
						}
						if !dryRun {
							backup, err := os.ReadFile(path + ".bak")
							if err != nil || !bytes.Equal(backup, original) {
								t.Fatalf("cleanup backup changed original facts: %s, %v", backup, err)
							}
						}
					} else {
						report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: filepath.Dir(workspace), DryRun: dryRun})
						if err != nil || report == nil || report.Updated != 1 || len(report.Results) != 1 || report.Results[0].Error != nil {
							t.Fatalf("explicit backfill cleanup did not run: %+v, %v", report, err)
						}
					}
					after, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if dryRun {
						if !bytes.Equal(original, after) {
							t.Fatal("dry-run cleanup changed ledger bytes")
						}
					} else {
						assertRawCleanupJSONEqual(t, expected, after)
						if !pending {
							if err := adapters.ValidateRuntimeFactLedger(after); err != nil {
								t.Fatalf("explicitly cleaned complete ledger remains unreadable: %v", err)
							}
						}
					}
				})
			}
		}
	}
}

func TestMetadataSummaryRequiresContextPolicyBeforePublication(t *testing.T) {
	for _, command := range []string{"migrate", "backfill"} {
		for _, removed := range []string{"  context_ceiling_warning_tokens: 160000\n", "  context_ceiling_critical_tokens: 200000\n"} {
			for _, summary := range []bool{false, true} {
				name := command + "/" + strings.TrimSpace(strings.SplitN(removed, ":", 2)[0])
				if summary {
					name += "/summary"
				} else {
					name += "/ledger-only-control"
				}
				t.Run(name, func(t *testing.T) {
					root := migrationTestRoot(t)
					sharedPath := filepath.Join(root, ".agenticlens-config.yaml")
					shared, err := os.ReadFile(sharedPath)
					if err != nil {
						t.Fatal(err)
					}
					withoutContext := bytes.Replace(shared, []byte(removed), nil, 1)
					if bytes.Equal(withoutContext, shared) {
						t.Fatal("fixture did not remove the selected context policy field")
					}
					if err := os.WriteFile(sharedPath, withoutContext, 0600); err != nil {
						t.Fatal(err)
					}
					workspace := filepath.Join(root, "records", "CASE-01")
					if err := os.MkdirAll(workspace, 0700); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(workspace, "agentic_metrics.json")
					original := []byte(`{"schema_version":6,"ticket_id":"CASE-01","updated_at":"recorded-time","sessions":[],"generation_policy":{"skills":true,"delegations":true,"anomalies":true}}`)
					if err := os.WriteFile(path, original, 0600); err != nil {
						t.Fatal(err)
					}
					summaryPath := filepath.Join(workspace, "SUMMARY.md")
					notes := []byte("# Existing human notes\n")
					if err := os.WriteFile(summaryPath, notes, 0600); err != nil {
						t.Fatal(err)
					}
					var operationErr error
					if command == "migrate" {
						report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{Project: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion, Backup: summary, UpdateSummaries: summary})
						operationErr = err
						if err == nil {
							if report == nil || len(report.Results) != 1 {
								t.Fatalf("missing migration outcome: %+v", report)
							}
							operationErr = report.Results[0].Error
						}
					} else {
						report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: filepath.Dir(workspace), UpdateSummaries: summary})
						operationErr = err
						if err == nil {
							if report == nil || len(report.Results) != 1 {
								t.Fatalf("missing backfill outcome: %+v", report)
							}
							operationErr = report.Results[0].Error
						}
					}
					after, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if summary {
						setting := "anomaly_thresholds." + strings.TrimSpace(strings.SplitN(removed, ":", 2)[0])
						if operationErr == nil || !strings.Contains(operationErr.Error(), setting) || !bytes.Equal(after, original) {
							t.Fatalf("missing context policy did not fail before ledger publication: %v, %s", operationErr, after)
						}
						for _, suffix := range []string{".bak", ".publication-pending"} {
							if _, err := os.Stat(path + suffix); !os.IsNotExist(err) {
								t.Fatalf("rejected summary created publication output %s: %v", suffix, err)
							}
						}
					} else if operationErr != nil || bytes.Contains(after, []byte("generation_policy")) {
						t.Fatalf("ledger-only cleanup required unused context policy: %v, %s", operationErr, after)
					}
					actualNotes, err := os.ReadFile(summaryPath)
					if err != nil || !bytes.Equal(actualNotes, notes) {
						t.Fatalf("rejected/unrequested summary changed human content: %s, %v", actualNotes, err)
					}
				})
			}
		}
	}
}
