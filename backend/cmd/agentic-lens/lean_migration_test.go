package main_test

import (
	"bytes"
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
			cmd := exec.Command(cliBinaryPath, "migrate", "--workspaces", root, "--ticket-id", "TEST-100", "--target-version", "6", "--backup")
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
			if got.Migration == nil || !strings.Contains(got.Migration.Notes, "dropped ownerless skills=1 task_links=0") || len(got.Migration.StepsCompleted) != 1 || got.Migration.StepsCompleted[0] != "schema_v6_lean_occurrences" {
				t.Fatalf("direct migration audit exposed an interim step or lost drop counts: %+v", got.Migration)
			}
			if backup := readCaptureFile(t, path+".bak"); !bytes.Equal(before, backup) {
				t.Fatal("migration backup changed original released-schema bytes")
			}
		})
	}
}
