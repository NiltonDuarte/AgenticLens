package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureWorkerRetriesPendingCleanupAfterDurableCommit(t *testing.T) {
	for _, scenario := range []struct {
		name, pendingAnomalies string
	}{
		{"matching snapshot", ""},
		{"unavailable owner with additional anomaly", "nonempty"},
		{"unavailable owner with explicit no findings", "empty"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
			if scenario.pendingAnomalies == "empty" {
				shared := filepath.Join(f.project, ".agenticlens-config.yaml")
				policy := strings.Replace(string(readCaptureFile(t, shared)), "anomalies: true", "anomalies: false", 1)
				if err := os.WriteFile(shared, []byte(policy), 0600); err != nil {
					t.Fatal(err)
				}
			}
			f.output = initialPublicationTarget(t, f, "durable")
			rollout := prepareCaptureOwner(t, f, "root", initialPublicationTicket, 21)
			run := func() {
				t.Helper()
				if output, err := runInitialPublicationWorker(t, f); err != nil {
					t.Fatalf("automatic publication failed: %v\n%s", err, output)
				}
			}

			run()
			snapshotPath, snapshot := readOnlyCaptureSnapshot(t, filepath.Join(f.project, captureFixturePendingDir))
			captured := readCaptureFile(t, snapshotPath)
			prepareInitialPublication(t, f, f.output)
			run()
			assertInitialPublicationAbsent(t, snapshotPath, f.output+".publication-pending")
			if err := adapters.CheckPublication(f.output); err != nil {
				t.Fatal(err)
			}
			initialLedger := f.ledger(t)
			ledgerBefore := readCaptureFile(t, f.output)
			summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
			summaryBefore := readCaptureFile(t, summary)

			// Restore the actual pending record, modeling interruption after the
			// canonical commit and before removal of the corresponding record.
			if scenario.pendingAnomalies != "" {
				if len(initialLedger.Sessions) != 1 || len(snapshot.Sessions) != 1 {
					t.Fatalf("expected one owner in the durable ledger and pending snapshot: ledger=%+v snapshot=%+v", initialLedger, snapshot)
				}
				// This accepted history fixture isolates a single portable fact
				// missing from the canonical ledger. No source remains to resolve
				// the disagreement or authoritatively supersede the pending fact.
				switch scenario.pendingAnomalies {
				case "nonempty":
					tool, count := "bash", int64(6)
					snapshot.Sessions[0].Anomalies = append(snapshot.Sessions[0].Anomalies, adapters.AnomalyDTO{
						RuleID: "TOOL_RETRY_STREAK", Severity: string(domain.SeverityWarning),
						Context: adapters.AnomalyContextDTO{ToolName: &tool, Count: &count},
					})
				case "empty":
					if initialLedger.Sessions[0].Anomalies != nil || snapshot.Sessions[0].Anomalies != nil {
						t.Fatal("fixture did not omit anomaly evidence before inspection")
					}
					snapshot.Sessions[0].Anomalies = []adapters.AnomalyDTO{}
				}
				var err error
				captured, err = json.MarshalIndent(snapshot, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				captured = append(captured, '\n')
				if scenario.pendingAnomalies == "empty" && !bytes.Contains(captured, []byte(`"anomalies": []`)) {
					t.Fatal("fixture did not persist the explicit no-findings value")
				}
				if err := os.Rename(rollout, filepath.Join(t.TempDir(), "unavailable.jsonl")); err != nil {
					t.Fatal(err)
				}
				local := filepath.Join(f.project, ".agenticlens-config.local.yaml")
				policy := strings.Replace(string(readCaptureFile(t, local)), "max_lifetime_ms: 10000", "max_lifetime_ms: 1400", 1)
				if err := os.WriteFile(local, []byte(policy), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(snapshotPath, captured, 0600); err != nil {
				t.Fatal(err)
			}
			if scenario.pendingAnomalies != "" {
				output, err := runInitialPublicationWorker(t, f)
				if err == nil || !strings.Contains(string(output), "pending facts disagree for unavailable owner root") {
					t.Fatalf("retry did not report unresolved pending coverage: %v\n%s", err, output)
				}
				if !bytes.Equal(readCaptureFile(t, snapshotPath), captured) || !bytes.Equal(readCaptureFile(t, f.output), ledgerBefore) ||
					!bytes.Equal(readCaptureFile(t, summary), summaryBefore) {
					t.Fatal("failed coverage changed the accepted pending facts or durable publication")
				}
				assertInitialPublicationAbsent(t, f.output+".publication-pending")
				return
			}
			run()
			assertInitialPublicationAbsent(t, snapshotPath, f.output+".publication-pending")
			after := f.ledger(t)
			if len(after.Sessions) != 1 || !reflect.DeepEqual(after.Sessions, initialLedger.Sessions) {
				t.Fatalf("retry changed or duplicated published facts: before=%+v after=%+v", initialLedger.Sessions, after.Sessions)
			}
		})
	}
}

func TestCapturePendingCleanupPreservesReplacementAfterObservation(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	pending := filepath.Join(f.project, captureFixturePendingDir)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	observe := func() []adapters.PendingSnapshotEntry {
		t.Helper()
		result, err := store.Reconcile(ctx, initialPublicationTicket, "fixture", 16, 64<<20, cfg.Generation.Metrics)
		if err != nil {
			t.Fatal(err)
		}
		entries, err := adapters.SyncPendingSnapshots(ctx, store, pending, f.project, result, true, nil, nil)
		if err != nil || len(entries) != 1 {
			t.Fatalf("observe pending root: %v (%d entries)", err, len(entries))
		}
		return entries
	}
	older := observe()
	path, _ := readOnlyCaptureSnapshot(t, pending)
	before := readCaptureFile(t, path)
	f.writeAt(t, 29, "2026-01-01T00:00:05Z")
	newer := observe()
	after := readCaptureFile(t, path)
	assertCaptureRootInput(t, domain.AgenticFactLedger{Sessions: newer[0].Sessions()}, 29)
	if bytes.Equal(before, after) {
		t.Fatal("test did not capture a source-derived replacement")
	}
	revision, err := store.CaptureLifecycleRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = adapters.RemovePromotedPendingSnapshots(ctx, store, revision, pending, f.project, older, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "pending snapshot changed during promotion") {
		t.Fatalf("cleanup did not reject an obsolete observation: %v", err)
	}
	if !bytes.Equal(readCaptureFile(t, path), after) {
		t.Fatal("cleanup changed the newer pending snapshot")
	}
}
