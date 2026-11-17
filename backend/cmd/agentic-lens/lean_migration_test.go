package main_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
)

func TestCLIMigrationUsesDirectV5ToFinalV6AndRejectsUnreleasedSchemas(t *testing.T) {
	legacy := `{"schema_version":5,"ticket_id":"TEST-100","ticket_title":"Direct lean migration","sessions":[{"id":"owner","source":"codex","user":"alice","invocations":[{"role":"coder","model":"fixture","cost_usd":1.25,"cost_evidence":{"state":"observed","known_subtotal_usd":1.25},"tokens":{"input":10}}],"tools":[],"anomalies":[]}],"skills":[{"session_id":"owner","name":"kept","caller_role":"coder","timestamp":"2026-01-01T00:00:00Z","success":true},{"name":"ownerless","success":true}],"task_links":[{"session_id":"owner","child_session_id":"unavailable-child","cost_usd":0,"cost_evidence":{"state":"observed","known_subtotal_usd":0}}]}`
	for _, tc := range []struct {
		name, input string
		success     bool
	}{
		{"released v5", legacy, true},
		{"unreleased rich6", strings.Replace(legacy, `"schema_version":5`, `"schema_version":6`, 1), false},
		{"unsupported7", strings.Replace(legacy, `"schema_version":5`, `"schema_version":7`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newCaptureGenerationFixture(t)
			root := filepath.Join(fixture.project, "records")
			dir := filepath.Join(root, "TEST-100")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "agentic_metrics.json")
			before := []byte(tc.input)
			if err := os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			if tc.success {
				dry := exec.Command(cliBinaryPath, "migrate", "--target-version", "6", "--backup", "--dry-run")
				dry.Dir = fixture.project
				preview, err := dry.CombinedOutput()
				if err != nil || !strings.Contains(string(preview), "dropped ownerless skills=1 task_links=0") || !strings.Contains(string(preview), "Dry run: no ledgers or backups were written") || strings.Contains(string(preview), "  Backup:") {
					t.Fatalf("dry-run audit or backup guidance was misleading: %v\n%s", err, preview)
				}
				if !bytes.Equal(readCaptureFile(t, path), before) {
					t.Fatal("dry-run changed original legacy ledger")
				}
				if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
					t.Fatalf("dry-run created rollback backup: %v", err)
				}
			}
			cmd := exec.Command(cliBinaryPath, "migrate", "--ticket-id", "TEST-100", "--target-version", "6", "--backup")
			cmd.Dir = fixture.project
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("migration runner failed before reporting its per-ticket result: %v %s", err, out)
			}
			after := readCaptureFile(t, path)
			if !tc.success {
				if !bytes.Equal(before, after) || !strings.Contains(string(out), "Migrated: 0 | Up to date: 0") || !strings.Contains(string(out), "Errors: 1") {
					t.Fatalf("unsupported/interim input was rewritten or reported up-to-date: %s\n%s", out, after)
				}
				if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
					t.Fatalf("rejected input created a publication backup: %v", err)
				}
				return
			}
			if err := adapters.ValidateRuntimeFactLedger(after); err != nil {
				t.Fatal(err)
			}
			got, err := adapters.LoadExistingFactLedger(path)
			if err != nil {
				t.Fatal(err)
			}
			if got.SchemaVersion == nil || *got.SchemaVersion != 6 || len(got.Sessions) != 1 || len(got.Sessions[0].Skills) != 1 || got.Sessions[0].Skills[0].Name != "kept" || len(got.Sessions[0].Delegations) != 1 || got.Sessions[0].Delegations[0].ChildSession != "unavailable-child" || got.Sessions[0].Invocations[0].CostUSD != 1.25 || got.Sessions[0].Invocations[0].Tokens.Input != 10 {
				t.Fatalf("direct migration lost owned facts or unrelated costs/tokens: %+v", got)
			}
			if got.Migration != nil || !strings.Contains(string(out), "dropped ownerless skills=1 task_links=0") || !strings.Contains(string(out), "completed steps: [schema_v6_lean_occurrences]") {
				t.Fatalf("completed audit must be console-only with direct step/drop counts: metadata=%+v\n%s", got.Migration, out)
			}
			if backup := readCaptureFile(t, path+".bak"); !bytes.Equal(before, backup) {
				t.Fatal("migration backup changed original released-schema bytes")
			}
			if !strings.Contains(string(out), "Ledger: "+path) || !strings.Contains(string(out), "Backup: "+path+".bak; restore this file over the ledger to revert it.") {
				t.Fatalf("migration output omitted actual ledger or rollback backup path:\n%s", out)
			}
		})
	}
}

func TestCLIBackfillReportsCompletedAuditBeforeRemovingMetadata(t *testing.T) {
	fixture := newCaptureGenerationFixture(t)
	root := filepath.Join(fixture.project, "records")
	dir := filepath.Join(root, "TEST-100", "iterations", "07")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "agentic_metrics.json")
	before := []byte(`{"schema_version":6,"ticket_id":"TEST-100","ticket_title":"Portable history","ticket_type":"chore","updated_at":"2026-01-01T00:00:00Z","code":{"lines_added":19,"lines_deleted":3,"test_lines_added":7,"test_lines_deleted":1,"prod_lines_added":12,"prod_lines_deleted":2,"files_changed":4,"sdlc_iterations":3},"migration":{"schema_version":6,"status":"completed","steps_completed":["schema_v6_lean_occurrences"],"steps_pending":[],"notes":"Schema v6 migration dropped ownerless skills=2 task_links=3"},"anomalies":[{"rule_id":"TOOL_RETRY_STREAK","severity":"WARNING","context":{"count":5}}],"sessions":[{"id":"remote-owner","source":"codex","user":"remote-user","invocations":[{"role":"coder","model":"fixture","cost_usd":1.25,"cost_evidence":{"state":"observed","known_subtotal_usd":1.25},"started_at":null,"completed_at":null,"active_duration_seconds":11,"turns":3,"compactions":1,"user_iterations":2,"tokens":{"input":10,"output":2,"cache_read":3,"reasoning":4,"peak_context":19}}],"tools":[{"name":"bash","calls":3,"errors":1}],"anomalies":[{"rule_id":"TOOL_HIGH_FAILURE_RATE","severity":"WARNING","context":{"count":1,"total_calls":3}}],"skills":[{"name":"kept","caller_role":null,"invoked_at":null}],"delegations":[{"agent_type":null,"invoked_at":null,"child_session_id":"unavailable-child"}]}]}`)
	if err := os.WriteFile(path, before, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, "backfill", "--source", "codex")
	cmd.Dir = fixture.project
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("backfill failed: %v\n%s", err, out)
	}
	after := readCaptureFile(t, path)
	if bytes.Contains(after, []byte(`"migration"`)) {
		t.Fatalf("completed migration remained persisted: %s", after)
	}
	if err := adapters.ValidateRuntimeFactLedger(after); err != nil {
		t.Fatal(err)
	}
	var expectedFields, actualFields map[string]json.RawMessage
	if err := json.Unmarshal(before, &expectedFields); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(after, &actualFields); err != nil {
		t.Fatal(err)
	}
	delete(expectedFields, "migration")
	if len(expectedFields) != len(actualFields) {
		t.Fatalf("backfill cleanup changed root member count: %s", after)
	}
	for key, raw := range expectedFields {
		expected, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := json.Marshal(actualFields[key])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expected, actual) {
			t.Fatalf("backfill cleanup changed non-migration field %q: before=%s after=%s", key, expected, actual)
		}
	}
	for _, expected := range []string{"Ledger: " + path, "completed steps: [schema_v6_lean_occurrences]", "dropped ownerless skills=2 task_links=3", "Rollback:"} {
		if !strings.Contains(string(out), expected) {
			t.Fatalf("backfill console omitted %q:\n%s", expected, out)
		}
	}
	if strings.Contains(string(out), "  Backup:") {
		t.Fatalf("backfill claimed an uncreated rollback backup:\n%s", out)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatalf("backfill unexpectedly created a backup: %v", err)
	}
}
