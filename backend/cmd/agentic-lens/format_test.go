package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteFormatModes(t *testing.T) {
	root := t.TempDir()
	writeInternalCLIPublicationPolicy(t, root, ".")
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
		{[]string{}, 2, "exactly one"}, {[]string{"--all", "--ticket-id", "T-1"}, 2, "exactly one"}, {[]string{"--ticket-id", ""}, 2, "nonblank"},
		{[]string{"--ticket-ids", "T-1,,T-2"}, 2, "empty"}, {[]string{"--all", "--dry-run", "--check"}, 2, "mutually exclusive"},
		{[]string{"--ticket-id", "missing"}, 2, "not_found"}, {[]string{"--ticket-id", "T-1", "--check"}, 1, ""},
		{[]string{"--ticket-ids", " T-1, T-1 ", "--dry-run"}, 0, ""},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			args := append([]string{"--project", root}, tc.args...)
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
	if code := executeFormat([]string{"--project", root, "--all"}, &out, &diagnostics); code != 0 {
		t.Fatalf("write code %d: %s", code, &diagnostics)
	}
	out.Reset()
	if code := executeFormat([]string{"--project", root, "--all", "--check"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "1 unchanged") {
		t.Fatalf("check code %d %s %s", code, &out, &diagnostics)
	}
	empty := t.TempDir()
	writeInternalCLIPublicationPolicy(t, empty, ".")
	out.Reset()
	if code := executeFormat([]string{"--project", empty, "--all"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "0 ledgers") {
		t.Fatalf("empty code %d %s", code, &out)
	}
}

func TestExecuteFormatConfiguredWorkspaceAndDiagnostics(t *testing.T) {
	project := t.TempDir()
	writeInternalCLIPublicationPolicy(t, project, "custom-records")
	configuration := "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: Fixture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: TEST\n      description: Fixture ticket\nworkspaces:\n  artifacts: []\n  roots:\n    - path: custom-records\n      discovery:\n        max_depth: 8\n        max_directories: 300\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(configuration), 0600); err != nil {
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
	if code := executeFormat([]string{"--project", project, "--ticket-id", "TEST-1"}, &out, &diagnostics); code != 0 || !strings.Contains(out.String(), "1 changed") {
		t.Fatalf("configured workspace code %d: %s %s", code, &out, &diagnostics)
	}
	// A malformed selected input is an input failure, distinct from check differences.
	if err := os.WriteFile(path, []byte(`{"ticket_id":`), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diagnostics.Reset()
	if code := executeFormat([]string{"--project", project, "--all", "--check"}, &out, &diagnostics); code != 2 || !strings.Contains(diagnostics.String(), "invalid ledger JSON") || !strings.Contains(out.String(), "1 failed") {
		t.Fatalf("malformed check code %d: %s %s", code, &out, &diagnostics)
	}
	if _, err := os.Stat(filepath.Join(project, "missing-state")); !os.IsNotExist(err) {
		t.Fatalf("format touched capture registry: %v", err)
	}
}

// Internal command fixtures declare both shared scope and independent private
// publication state; no capture registry or installed engine is needed.
func writeInternalCLIPublicationPolicy(t *testing.T, project, searchRoot string) {
	t.Helper()
	body, err := os.ReadFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.Replace(body, []byte("    - path: records"), []byte("    - path: "+searchRoot), 1)
	body = bytes.Replace(body, []byte("max_depth: 8"), []byte("max_depth: 1"), 1)
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("agenticlens:\n  state_dir: %q\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n", filepath.Join(project, "publication-state"), filepath.Join(project, "missing-state"))
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
}
