package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestRegisteredGenerateRecoversRetiredCacheAndPreservesUnavailableCurrentOwners(t *testing.T) {
	for _, mode := range []string{"unchanged complete source", "missing native source", "malformed native source"} {
		t.Run(mode, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 17)
			source := append(readCaptureFile(t, path), []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"native-call","name":"load_skill","arguments":"{\"name\":\"native-skill\"}"}}`+"\n")...)
			if err := os.WriteFile(path, source, 0600); err != nil {
				t.Fatal(err)
			}
			f.bind(t, path)
			baseline := f.generate(t)
			// A valid portable current snapshot may carry facts whose native source later
			// disappears. Unsupported local cache bodies must never authorize replacing it.
			if mode != "unchanged complete source" {
				baseline.Sessions[0].Skills = append(baseline.Sessions[0].Skills, domain.SkillFact{Name: "accepted-current"})
				baseline.Sessions[0].Delegations = []domain.DelegationFact{{Description: "accepted delegation", ChildSession: "unavailable-child"}}
				baseline.Sessions[0].Anomalies = []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}}
				baseline.Anomalies = []domain.Anomaly{{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityWarning}}
				if err := adapters.SaveFactLedger(f.output, baseline); err != nil {
					t.Fatal(err)
				}
			}
			before := readCaptureFile(t, f.output)
			var acceptedWire map[string]json.RawMessage
			if err := json.Unmarshal(before, &acceptedWire); err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("sqlite", filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			readRow := func() []byte {
				t.Helper()
				var row []byte
				if err := db.QueryRow(`SELECT record FROM registrations WHERE session_id='root'`).Scan(&row); err != nil {
					t.Fatal(err)
				}
				return row
			}
			var row map[string]json.RawMessage
			if err = json.Unmarshal(readRow(), &row); err != nil {
				t.Fatal(err)
			}
			var originalRegistration domain.CaptureRegistration
			if err = json.Unmarshal(row["registration"], &originalRegistration); err != nil {
				t.Fatal(err)
			}
			var snapshot []map[string]json.RawMessage
			if err = json.Unmarshal(row["snapshot"], &snapshot); err != nil {
				t.Fatal(err)
			}
			snapshot[0]["events"] = json.RawMessage(`[]`)
			snapshot[0]["skills"] = json.RawMessage(`[{"name":"cache-only","caller_role":"coder","timestamp":"2026-01-01T00:00:01Z","success":true}]`)
			row["snapshot"], err = json.Marshal(snapshot)
			if err != nil {
				t.Fatal(err)
			}
			opaque := append([]byte(nil), row["snapshot"]...)
			legacy, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(`UPDATE registrations SET record=? WHERE session_id='root'`, legacy); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "missing native source":
				if err = os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "malformed native source":
				broken := append(append([]byte(nil), source...), []byte(`{"type":"response_item","payload":{"type":"function_call","call_id":42,"name":"load_skill","arguments":"{}"}}`+"\n")...)
				if err = os.WriteFile(path, broken, 0600); err != nil {
					t.Fatal(err)
				}
			}
			// No explicit selector/nonce: the existing binding must be discoverable even
			// though its cached body is old, and must reach actual native extraction.
			summaryPath := filepath.Join(f.project, "SUMMARY.md")
			cmd := exec.Command(cliBinaryPath, f.args(t, "--append", summaryPath)...)
			cmd.Dir = f.project
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("registered-only cache recovery blocked before native extraction: %v %s", err, out)
			}
			after := readCaptureFile(t, f.output)
			if err = adapters.ValidateRuntimeFactLedger(after); err != nil {
				t.Fatalf("registered generation published invalid current ledger: %v", err)
			}
			loaded, err := adapters.LoadExistingFactLedger(f.output)
			if err != nil {
				t.Fatal(err)
			}
			if len(loaded.Sessions) != 1 || loaded.Sessions[0].ID != "root" {
				t.Fatalf("accepted owner lost or old cache became owner data: %+v", loaded.Sessions)
			}
			registry, err := capture.OpenReadOnlyProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			rows, listErr := registry.List(context.Background())
			closeErr := registry.Close()
			if listErr != nil || closeErr != nil {
				t.Fatalf("post-generation registry not readable: %v %v", listErr, closeErr)
			}
			if len(rows) != 1 || rows[0].TicketID != "TEST-100" || rows[0].Request != originalRegistration.Request || rows[0].Identity != originalRegistration.Identity || rows[0].UpdatedAt != originalRegistration.UpdatedAt || rows[0].SourceFingerprint != originalRegistration.SourceFingerprint {
				t.Fatalf("recovery lost explicit binding: %+v", rows)
			}
			var finalRow map[string]json.RawMessage
			if err = json.Unmarshal(readRow(), &finalRow); err != nil {
				t.Fatal(err)
			}
			if mode == "unchanged complete source" {
				if rows[0].SnapshotNeedsRefresh || rows[0].State != domain.CaptureCurrent || len(loaded.Sessions[0].Skills) != 1 || loaded.Sessions[0].Skills[0].Name != "native-skill" || strings.Contains(string(finalRow["snapshot"]), `"events"`) || strings.Contains(string(finalRow["snapshot"]), "cache-only") {
					t.Fatalf("unchanged native source did not publish fresh lean cache: rows=%+v ledger=%+v", rows, loaded)
				}
				if !reflect.DeepEqual(loaded.Sessions, baseline.Sessions) {
					t.Fatalf("cache recovery changed unchanged native facts: %+v", loaded.Sessions)
				}
			} else {
				if !rows[0].SnapshotNeedsRefresh || rows[0].State == domain.CaptureCurrent || len(rows[0].Snapshot) != 0 || string(finalRow["snapshot"]) != string(opaque) {
					t.Fatalf("unavailable/invalid source cleared old opaque recovery evidence or accepted empty cache: %+v raw=%s", rows, finalRow["snapshot"])
				}
				var finalWire map[string]json.RawMessage
				if err = json.Unmarshal(after, &finalWire); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(acceptedWire["sessions"], finalWire["sessions"]) || !bytes.Equal(acceptedWire["anomalies"], finalWire["anomalies"]) || !reflect.DeepEqual(loaded.Sessions, baseline.Sessions) {
					t.Fatalf("failed source read altered accepted current owner or root summaries:\nbefore=%s\nafter=%s", before, after)
				}
				summary := readCaptureFile(t, summaryPath)
				if !strings.Contains(string(summary), "failed: 1") || !strings.Contains(string(summary), "pending processing does not establish freshness") {
					t.Fatalf("cache/source failure omitted honest incomplete-coverage report: %s", summary)
				}
				// Restoring the same exact native checkpoint must unlock the same record;
				// no new binding or private-history authorization may be required.
				if err = os.WriteFile(path, source, 0600); err != nil {
					t.Fatal(err)
				}
				retry := exec.Command(cliBinaryPath, f.args(t, "--append", summaryPath)...)
				retry.Dir = f.project
				if retryOut, retryErr := retry.CombinedOutput(); retryErr != nil {
					t.Fatalf("source restoration could not finish old-cache recovery: %v %s", retryErr, retryOut)
				}
				restored, err := adapters.LoadExistingFactLedger(f.output)
				if err != nil {
					t.Fatal(err)
				}
				if len(restored.Sessions) != 1 || len(restored.Sessions[0].Skills) != 1 || restored.Sessions[0].Skills[0].Name != "native-skill" || len(restored.Sessions[0].Delegations) != 0 {
					t.Fatalf("complete restored source did not replace its own prior snapshot: %+v", restored)
				}
				restoredRegistry, err := capture.OpenReadOnlyProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}
				restoredRows, err := restoredRegistry.List(context.Background())
				closeErr = restoredRegistry.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("recovered registry unreadable: %v %v", err, closeErr)
				}
				if len(restoredRows) != 1 || restoredRows[0].SnapshotNeedsRefresh || restoredRows[0].Request != originalRegistration.Request || restoredRows[0].State != domain.CaptureCurrent {
					t.Fatalf("restored source did not clear recovery state while preserving binding: %+v", restoredRows)
				}

			}
		})
	}
}

func TestRegisteredDraftDurationCacheNormalizesWithoutReadMutation(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := f.write(t, 17)
	f.bind(t, path)
	f.generate(t)
	db, err := sql.Open("sqlite", filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var original []byte
	if err := db.QueryRow(`SELECT record FROM registrations WHERE session_id='root'`).Scan(&original); err != nil {
		t.Fatal(err)
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(original, &record); err != nil {
		t.Fatal(err)
	}
	var owners []map[string]json.RawMessage
	if err := json.Unmarshal(record["snapshot"], &owners); err != nil {
		t.Fatal(err)
	}
	var invocations []map[string]json.RawMessage
	if err := json.Unmarshal(owners[0]["invocations"], &invocations); err != nil {
		t.Fatal(err)
	}
	invocations[0]["active_duration_seconds"] = json.RawMessage(`10`)
	invocations[0]["duration_evidence"] = json.RawMessage(`{"state":"provisional","completed_seconds":3,"provisional_seconds":7,"cutoff":"2026-01-01T00:00:10Z","basis":"codex:turn"}`)
	owners[0]["invocations"], err = json.Marshal(invocations)
	if err != nil {
		t.Fatal(err)
	}
	record["snapshot"], err = json.Marshal(owners)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE registrations SET record=? WHERE session_id='root'`, draft); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	registry, err := capture.OpenReadOnlyProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	rows, listErr := registry.List(context.Background())
	closeErr := registry.Close()
	if listErr != nil || closeErr != nil {
		t.Fatalf("read compatible cache: %v %v", listErr, closeErr)
	}
	if len(rows) != 1 || rows[0].SnapshotNeedsRefresh || len(rows[0].Snapshot) != 1 || len(rows[0].Snapshot[0].Invocations) != 1 {
		t.Fatalf("valid draft cache discarded: %+v", rows)
	}
	inv := rows[0].Snapshot[0].Invocations[0]
	if rows[0].Snapshot[0].ID != "root" || inv.ActiveDurationSeconds != 10 || inv.ProvisionalSeconds != 7 || inv.Tokens.Input != 17 {
		t.Fatalf("draft cache numeric/identity loss: %+v", rows[0])
	}
	var after []byte
	if err := db.QueryRow(`SELECT record FROM registrations WHERE session_id='root'`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(draft, after) {
		t.Fatal("read-only cache normalization mutated stored record")
	}
}
