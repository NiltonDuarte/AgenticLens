package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteFormatModes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "T-1", "agentic_metrics.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	original := `{"ticket_id":"T-1","sessions":[]}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args       []string
		code       int
		diagnostic string
	}{
		{[]string{}, 2, "exactly one"}, {[]string{"--all", "--ticket-id", "T-1"}, 2, "exactly one"}, {[]string{"--ticket-id", ""}, 2, "unsafe or empty"},
		{[]string{"--ticket-ids", "T-1,,T-2"}, 2, "empty"}, {[]string{"--all", "--dry-run", "--check"}, 2, "mutually exclusive"},
		{[]string{"--ticket-id", "missing"}, 2, "no such"}, {[]string{"--ticket-id", "T-1", "--check"}, 1, ""},
		{[]string{"--ticket-ids", " T-1, T-1 ", "--dry-run"}, 0, ""},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			args := append([]string{"--workspaces", root}, tc.args...)
			if code := executeFormat(args, &out, &diagnostics); code != tc.code {
				t.Fatalf("code %d: %s %s", code, &out, &diagnostics)
			}
			if !strings.Contains(diagnostics.String(), tc.diagnostic) {
				t.Errorf("diagnostic %q missing %q", &diagnostics, tc.diagnostic)
			}
			body, err := os.ReadFile(path)
			if err != nil || string(body) != original {
				t.Fatal("non-writing mode changed ledger")
			}
		})
	}
	var out, diagnostics bytes.Buffer
	if code := executeFormat([]string{"--workspaces", root, "--all"}, &out, &diagnostics); code != 0 {
		t.Fatalf("write code %d: %s", code, &diagnostics)
	}
	out.Reset()
	if code := executeFormat([]string{"--workspaces", root, "--all", "--check"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "1 unchanged") {
		t.Fatalf("check code %d %s %s", code, &out, &diagnostics)
	}
	empty := t.TempDir()
	out.Reset()
	if code := executeFormat([]string{"--workspaces", empty, "--all"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "0 ledgers") {
		t.Fatalf("empty code %d %s", code, &out)
	}
}

func TestExecuteFormatConfiguredWorkspaceAndDiagnostics(t *testing.T) {
	project := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(project); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Error(err)
		}
	})
	configuration := "project_name: Fixture\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: custom-records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\ntelemetry:\n  capture:\n    state_dir: missing-state\n"
	if err := os.WriteFile(".agenticlens-config.yaml", []byte(configuration), 0600); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(project, "custom-records")
	if err := os.MkdirAll(filepath.Join(root, "TEST-1"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "TEST-1", "agentic_metrics.json")
	if err := os.WriteFile(path, []byte(`{"ticket_id":"TEST-1","sessions":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := executeFormat([]string{"--ticket-id", "TEST-1"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "1 changed") {
		t.Fatalf("configured workspace code %d: %s %s", code, &out, &diagnostics)
	}
	// A malformed selected input is an input failure, distinct from check differences.
	if err := os.WriteFile(path, []byte(`{"ticket_id":`), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := executeFormat([]string{"--all", "--check"}, &out, &diagnostics); code != 2 || !strings.Contains(diagnostics.String(), "invalid ledger JSON") || !strings.Contains(out.String(), "1 failed") {
		t.Fatalf("malformed check code %d: %s %s", code, &out, &diagnostics)
	}
	if _, err := os.Stat(filepath.Join(project, "missing-state")); !os.IsNotExist(err) {
		t.Fatalf("format touched capture registry: %v", err)
	}
}
