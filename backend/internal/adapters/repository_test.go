package adapters_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestResolveWorkspacesDir(t *testing.T) {
	t.Run("explicit override", func(t *testing.T) {
		got, err := adapters.ResolveWorkspacesDir("/custom/path")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/custom/path" {
			t.Errorf("expected /custom/path, got %s", got)
		}
	})

	t.Run("empty configuration returns empty string without defaulting", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgContent := `{"project_name": "Test", "ticket": {"prefixes": [{"prefix": "TEST"}]}}`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string when workspaces.dir is unconfigured, got %q", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("empty override with no config returns empty string", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string when unconfigured, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("empty override with config missing workspaces.dir returns empty string", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\ncode_metrics:\n  source_roots:\n    - src\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string when workspaces.dir is unconfigured, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("does not fall back to existing workspaces directory when unconfigured", func(t *testing.T) {
		tempDir := t.TempDir()
		// Hostile scenario: a physical ./workspaces directory exists on disk,
		// but configuration has empty workspaces.dir or does not exist.
		existingWs := filepath.Join(tempDir, "workspaces")
		if err := os.MkdirAll(existingWs, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		// 1. Without config file
		resolvedNoCfg, err := adapters.ResolveWorkspacesDir("")
		if resolvedNoCfg != "" {
			t.Errorf("expected empty string when unconfigured even if ./workspaces exists, got %s", resolvedNoCfg)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}

		// 2. With config file but missing workspaces.dir
		cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\ncode_metrics:\n  source_roots:\n    - src\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}
		resolvedEmptyWs, err := adapters.ResolveWorkspacesDir("")
		if resolvedEmptyWs != "" {
			t.Errorf("expected empty string when workspaces.dir unconfigured even if ./workspaces exists, got %s", resolvedEmptyWs)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("root yaml config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "root_yaml_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: root_yaml_workspaces\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved != "root_yaml_workspaces" {
			t.Errorf("expected root_yaml_workspaces, got %s", resolved)
		}
	})

	t.Run("root yml config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "root_yml_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: root_yml_workspaces\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved != "root_yml_workspaces" {
			t.Errorf("expected root_yml_workspaces, got %s", resolved)
		}
	})

	t.Run("root json config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "root_json_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := `{"project_name": "Test", "ticket": {"prefixes": [{"prefix": "TEST"}]}, "workspaces": {"dir": "root_json_workspaces"}, "code_metrics": {"source_roots": ["src"]}, "anomaly_thresholds": {"context_ceiling_warning_tokens": 160000, "context_ceiling_critical_tokens": 200000}}`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved != "root_json_workspaces" {
			t.Errorf("expected root_json_workspaces, got %s", resolved)
		}
	})

	t.Run("nested execution ascends to root config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "deep_root_ws")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: deep_root_ws\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		nestedDir := filepath.Join(tempDir, "sub1", "sub2", "sub3")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(nestedDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved != "deep_root_ws" {
			t.Errorf("expected deep_root_ws, got %s", resolved)
		}
	})
	t.Run("precedence between override and config", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "precedence_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Also create a default workspaces directory in the same root
		defaultWsDir := filepath.Join(tempDir, "workspaces")
		if err := os.MkdirAll(defaultWsDir, 0755); err != nil {
			t.Fatal(err)
		}

		cfgContent := "project_name: TestPrecedence\nticket:\n  prefixes:\n    - prefix: PREC\nworkspaces:\n  dir: precedence_workspaces\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		// Configured workspaces directory takes precedence over default workspaces
		resolved, err := adapters.ResolveWorkspacesDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resolved != "precedence_workspaces" {
			t.Errorf("expected precedence_workspaces, got %s", resolved)
		}

		// Explicit override takes precedence over configured workspaces directory
		overridePath := filepath.Join(tempDir, "explicit_override")
		resolvedOverride, err := adapters.ResolveWorkspacesDir(overridePath)
		if err != nil {
			t.Fatalf("unexpected error on override: %v", err)
		}
		if resolvedOverride != overridePath {
			t.Errorf("expected override %s, got %s", overridePath, resolvedOverride)
		}
	})

	t.Run("fails fast when no command flag and no config exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cwdWs := filepath.Join(tempDir, "workspaces")
		if err := os.MkdirAll(cwdWs, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("fails fast when config exists but workspaces dir is empty", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgContent := "project_name: TestEmptyWS\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: \"\"\ncode_metrics:\n  source_roots:\n    - src\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("parent traversal workspaces directory is never searched or selected", func(t *testing.T) {
		parentDir := t.TempDir()
		parentWs := filepath.Join(parentDir, "workspaces")
		if err := os.MkdirAll(parentWs, 0755); err != nil {
			t.Fatal(err)
		}
		childDir := filepath.Join(parentDir, "subproject", "nested")
		if err := os.MkdirAll(childDir, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(childDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("container path is never selected when unconfigured", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved, err := adapters.ResolveWorkspacesDir("")
		if resolved != "" {
			t.Errorf("expected empty string, got %s", resolved)
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("ScanWorkspaces returns error in scan report when unconfigured", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		report := adapters.ScanWorkspaces("")
		if len(report.Ledgers) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(report.Ledgers))
		}
		if len(report.Errors) == 0 {
			t.Fatal("expected error in report, got none")
		}
		if report.FatalError == "" {
			t.Fatal("expected unconfigured workspace directory to be a fatal scan error")
		}
		if !strings.Contains(report.Errors[0], "workspaces directory not found") {
			t.Errorf("expected 'workspaces directory not found' in error, got: %s", report.Errors[0])
		}
	})

	t.Run("LoadTicketWorkspace fails with wrapped error when unconfigured", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-1", "")
		if ledger != nil {
			t.Errorf("expected nil ledger, got %+v", ledger)
		}
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
			t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
		}
	})

	t.Run("ReadTicketArtifacts returns empty artifacts when unconfigured", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		artifacts := adapters.ReadTicketArtifacts("AL-1", "")
		if artifacts.TicketID != "AL-1" {
			t.Errorf("expected TicketID AL-1, got %s", artifacts.TicketID)
		}
		if len(artifacts.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(artifacts.Items))
		}
	})
}

func TestScanWorkspaces(t *testing.T) {
	tempDir := t.TempDir()
	ticketDir := filepath.Join(tempDir, "TEST-1")
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		t.Fatal(err)
	}

	factJSON := `{
  "schema_version": 4,
  "ticket_id": "TEST-1",
  "ticket_title": "Test Ticket",
  "updated_at": "2026-10-17T15:00:00Z",
  "code": {
    "lines_added": 10,
    "lines_deleted": 2,
    "test_lines_added": 5,
    "test_lines_deleted": 0,
    "prod_lines_added": 5,
    "prod_lines_deleted": 2,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "ses-1",
      "source": "opencode",
      "user": "test-user",
      "invocations": [
        {
          "role": "orchestrator",
          "model": "model-x",
          "started_at": null,
          "completed_at": null,
          "active_duration_seconds": 60,
          "turns": 5,
          "tokens": {
            "input": 1000,
            "output": 100,
            "cache_read": 500,
            "reasoning": 50,
            "peak_context": 20000
          },
          "cost_usd": 0.05,
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [
        {
          "name": "bash",
          "calls": 2,
          "errors": 0
        }
      ],
      "events": [],
      "anomalies": []
    }
  ]
}`

	metricPath := filepath.Join(ticketDir, "agentic_metrics.json")
	if err := os.WriteFile(metricPath, []byte(factJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a corrupt file to verify non-fatal error handling
	corruptDir := filepath.Join(tempDir, "CORRUPT-1")
	if err := os.MkdirAll(corruptDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "agentic_metrics.json"), []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	report := adapters.ScanWorkspaces(tempDir)
	if len(report.Ledgers) != 1 {
		t.Fatalf("expected 1 valid ledger, got %d", len(report.Ledgers))
	}
	if *report.Ledgers[0].TicketID != "TEST-1" {
		t.Errorf("expected ticket_id TEST-1, got %s", *report.Ledgers[0].TicketID)
	}
	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 non-fatal error for corrupt file, got %d", len(report.Errors))
	}
	if !strings.Contains(report.Errors[0], "CORRUPT-1") {
		t.Errorf("expected error to mention CORRUPT-1, got %s", report.Errors[0])
	}
	if len(report.Diagnostics) != 1 || report.Diagnostics[0].TicketID != "CORRUPT-1" || !strings.Contains(report.Diagnostics[0].Message, "CORRUPT-1") {
		t.Fatalf("expected ticket-scoped diagnostic for skipped workspace, got %+v", report.Diagnostics)
	}

	t.Run("unconfigured workspaces returns empty ledgers with descriptive error", func(t *testing.T) {
		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		report := adapters.ScanWorkspaces("")
		if len(report.Ledgers) != 0 {
			t.Fatalf("expected 0 ledgers when workspaces unconfigured, got %d", len(report.Ledgers))
		}
		if len(report.Errors) == 0 {
			t.Fatalf("expected at least 1 error in scan report when workspaces unconfigured")
		}
		foundExpected := false
		for _, e := range report.Errors {
			if strings.Contains(e, "Workspaces directory is unconfigured or not found") {
				foundExpected = true
				break
			}
		}
		if !foundExpected {
			t.Errorf("expected error containing 'Workspaces directory is unconfigured or not found', got: %v", report.Errors)
		}
	})
}

func TestScanWorkspaces_NaturalOrdering(t *testing.T) {
	tempDir := t.TempDir()

	tickets := []string{"AL-10", "AL-2", "AL-1", "GEN-1", "AL-22"}
	for _, tid := range tickets {
		dir := filepath.Join(tempDir, tid)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		factJSON := `{"schema_version":4,"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"ticket_id":"` + tid + `","updated_at":"2026-10-18T10:00:00Z","sessions":[]}`
		if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
			t.Fatal(err)
		}
	}

	report := adapters.ScanWorkspaces(tempDir)
	if len(report.Ledgers) != len(tickets) {
		t.Fatalf("expected %d ledgers, got %d", len(tickets), len(report.Ledgers))
	}

	expectedOrder := []string{"AL-1", "AL-2", "AL-10", "AL-22", "GEN-1"}
	for i, expected := range expectedOrder {
		if report.Ledgers[i].TicketID == nil || *report.Ledgers[i].TicketID != expected {
			t.Errorf("ledger index %d: expected %s, got %v", i, expected, report.Ledgers[i].TicketID)
		}
	}
}

func TestScanWorkspaces_AdversarialResilienceAndOrdering(t *testing.T) {
	tempDir := t.TempDir()

	validWorkspaces := []string{
		"AL-10",
		"AL-2",
		"123",
		"AL-01",
		"al-1",
		"unknown",
		"GEN-2",
		"AL-1",
	}

	for _, tid := range validWorkspaces {
		dir := filepath.Join(tempDir, tid)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		factJSON := `{"schema_version":4,"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"ticket_id":"` + tid + `","updated_at":"2026-10-18T10:00:00Z","sessions":[]}`
		if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Empty workspace directory (should be ignored gracefully without error)
	emptyDir := filepath.Join(tempDir, "EMPTY-WS")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Corrupted JSON workspace (should produce a scan error without crashing)
	corruptDir := filepath.Join(tempDir, "CORRUPT-JSON")
	if err := os.MkdirAll(corruptDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "agentic_metrics.json"), []byte("{invalid-json"), 0644); err != nil {
		t.Fatal(err)
	}

	report := adapters.ScanWorkspaces(tempDir)

	if len(report.Ledgers) != len(validWorkspaces) {
		t.Fatalf("expected %d valid ledgers, got %d", len(validWorkspaces), len(report.Ledgers))
	}

	// Natural ordering expectation:
	// AL-01 < AL-1 < al-1 < AL-2 < AL-10 < GEN-2 < 123 < unknown
	expectedNaturalOrder := []string{
		"AL-01",
		"AL-1",
		"al-1",
		"AL-2",
		"AL-10",
		"GEN-2",
		"123",
		"unknown",
	}

	for i, expected := range expectedNaturalOrder {
		if report.Ledgers[i].TicketID == nil || *report.Ledgers[i].TicketID != expected {
			t.Errorf("ledger index %d: expected %s, got %v", i, expected, report.Ledgers[i].TicketID)
		}
	}

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 scan error for CORRUPT-JSON, got %d: %v", len(report.Errors), report.Errors)
	}
	if !strings.Contains(report.Errors[0], "CORRUPT-JSON") {
		t.Errorf("expected error to mention CORRUPT-JSON, got %s", report.Errors[0])
	}
}

func TestReadTicketArtifacts(t *testing.T) {
	tempDir := t.TempDir()
	ticketDir := filepath.Join(tempDir, "TICKET-10")
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ticketDir, "CONTRACT.md"), []byte("# Contract"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ticketDir, "SUMMARY.md"), []byte("# Summary"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("valid read with zero artifact defaults", func(t *testing.T) {
		art := adapters.ReadTicketArtifacts("TICKET-10", tempDir)
		if art.ContractContent != nil {
			t.Errorf("expected nil ContractContent without configured artifacts, got: %v", art.ContractContent)
		}
		if art.SummaryContent != nil {
			t.Errorf("expected nil SummaryContent without configured artifacts, got: %v", art.SummaryContent)
		}
		if len(art.Items) != 0 {
			t.Fatalf("expected empty Items slice without configured artifacts, got %d", len(art.Items))
		}
	})

	t.Run("valid read with explicit configured artifacts", func(t *testing.T) {
		cfgArtifacts := []domain.WorkspaceArtifactConfig{
			{Path: "CONTRACT.md", Title: "Contract", Description: "Contract file"},
			{Path: "SUMMARY.md", Title: "Summary", Description: "Summary file"},
		}
		art := adapters.ReadTicketArtifacts("TICKET-10", tempDir, cfgArtifacts)
		if art.ContractContent == nil || *art.ContractContent != "# Contract" {
			t.Errorf("contract content mismatch: %v", art.ContractContent)
		}
		if art.SummaryContent == nil || *art.SummaryContent != "# Summary" {
			t.Errorf("summary content mismatch: %v", art.SummaryContent)
		}
		if len(art.Items) != 2 {
			t.Fatalf("expected 2 Items, got %d", len(art.Items))
		}
		var foundContract, foundSummary bool
		for _, item := range art.Items {
			if item.Path == "CONTRACT.md" && item.Content != nil && *item.Content == "# Contract" {
				foundContract = true
			}
			if item.Path == "SUMMARY.md" && item.Content != nil && *item.Content == "# Summary" {
				foundSummary = true
			}
		}
		if !foundContract || !foundSummary {
			t.Errorf("expected Items to contain CONTRACT.md and SUMMARY.md, got %+v", art.Items)
		}
	})

	t.Run("custom configured artifacts with nested paths and missing files", func(t *testing.T) {
		iterDir := filepath.Join(ticketDir, "iterations", "01")
		if err := os.MkdirAll(iterDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(iterDir, "PLAN.md"), []byte("# Iteration Plan"), 0644); err != nil {
			t.Fatal(err)
		}

		customConfigs := []domain.WorkspaceArtifactConfig{
			{
				Path:        "CONTRACT.md",
				Title:       "Frozen Contract",
				Description: "Specification boundaries",
			},
			{
				Path:        "iterations/01/PLAN.md",
				Title:       "Active Plan",
				Description: "Iteration steps",
			},
			{
				Path:        "NONEXISTENT.md",
				Title:       "Missing Doc",
				Description: "Should be nil content",
			},
			{
				Path:        "../../../etc/passwd",
				Title:       "Malicious Path",
				Description: "Should be ignored",
			},
			{
				Path:        "/etc/shadow",
				Title:       "Absolute Path",
				Description: "Should be ignored",
			},
		}

		art := adapters.ReadTicketArtifacts("TICKET-10", tempDir, customConfigs)
		if len(art.Items) != 3 {
			t.Fatalf("expected exactly 3 valid items (traversals excluded), got %d", len(art.Items))
		}

		if art.Items[0].Title != "Frozen Contract" || art.Items[0].Content == nil || *art.Items[0].Content != "# Contract" {
			t.Errorf("item 0 mismatch: %+v", art.Items[0])
		}
		if art.Items[1].Title != "Active Plan" || art.Items[1].Content == nil || *art.Items[1].Content != "# Iteration Plan" {
			t.Errorf("item 1 mismatch: %+v", art.Items[1])
		}
		if art.Items[2].Title != "Missing Doc" || art.Items[2].Content != nil {
			t.Errorf("item 2 should have nil content: %+v", art.Items[2])
		}
	})

	t.Run("path traversal attack prevention", func(t *testing.T) {
		art := adapters.ReadTicketArtifacts("../../../etc", tempDir)
		if art.ContractContent != nil || art.SummaryContent != nil {
			t.Errorf("expected traversal attempt to yield nil contents")
		}
	})

	t.Run("nested and absolute path traversal variants", func(t *testing.T) {
		variants := []string{
			"../../",
			"../",
			"TICKET-10/../../",
			"/etc/passwd",
			"./../",
		}
		for _, v := range variants {
			art := adapters.ReadTicketArtifacts(v, tempDir)
			if art.ContractContent != nil || art.SummaryContent != nil {
				t.Errorf("expected path traversal variant %q to yield nil contents", v)
			}
		}
	})

	t.Run("missing workspace directory", func(t *testing.T) {
		art := adapters.ReadTicketArtifacts("NONEXISTENT", filepath.Join(tempDir, "does_not_exist"))
		if art.ContractContent != nil || art.SummaryContent != nil {
			t.Errorf("expected missing workspace dir to yield nil contents")
		}
	})
}

func TestScanWorkspacesResilience(t *testing.T) {
	t.Run("unconfigured workspaces directory returns clear error", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgContent := `{"project_name": "Test", "ticket": {"prefixes": [{"prefix": "TEST"}]}}`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.json"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		report := adapters.ScanWorkspaces("")
		if len(report.Ledgers) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(report.Ledgers))
		}
		if len(report.Errors) != 1 || !strings.Contains(report.Errors[0], "Workspaces directory must be configured") {
			t.Errorf("expected configuration error message, got: %v", report.Errors)
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		report := adapters.ScanWorkspaces("/non/existent/path/for/sure")
		if len(report.Ledgers) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(report.Ledgers))
		}
		if len(report.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(report.Errors))
		}
		if !strings.Contains(report.Errors[0], "Workspaces directory not found") {
			t.Errorf("unexpected error message: %s", report.Errors[0])
		}
	})

	t.Run("empty directory with no files", func(t *testing.T) {
		emptyDir := t.TempDir()
		report := adapters.ScanWorkspaces(emptyDir)
		if len(report.Ledgers) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(report.Ledgers))
		}
		if len(report.Errors) != 0 {
			t.Errorf("expected 0 errors for clean empty directory, got %d", len(report.Errors))
		}
	})

	t.Run("parent directory name inference for omitted ticket ID", func(t *testing.T) {
		tempDir := t.TempDir()
		ticketDir := filepath.Join(tempDir, "INFERRED-42")
		if err := os.MkdirAll(ticketDir, 0755); err != nil {
			t.Fatal(err)
		}
		factJSON := `{"schema_version":4,"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"updated_at":"2026-10-17T12:00:00Z","sessions":[]}`
		if err := os.WriteFile(filepath.Join(ticketDir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
			t.Fatal(err)
		}

		report := adapters.ScanWorkspaces(tempDir)
		if len(report.Ledgers) != 1 {
			t.Fatalf("expected 1 ledger, got %d", len(report.Ledgers))
		}
		if report.Ledgers[0].TicketID == nil || *report.Ledgers[0].TicketID != "INFERRED-42" {
			t.Errorf("expected inferred ticket_id INFERRED-42, got %v", report.Ledgers[0].TicketID)
		}
	})
}

func TestBoundarySchemaParsingAndDefaults(t *testing.T) {
	t.Run("minimal valid json with defaults", func(t *testing.T) {
		raw := `{"updated_at": "2026-10-17T12:00:00Z"}`
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		dom := dto.ToDomain()
		if dom.UpdatedAt != "2026-10-17T12:00:00Z" {
			t.Errorf("expected updated_at preserved, got %s", dom.UpdatedAt)
		}
		if dom.Code.SDLCIterations != 0 {
			t.Errorf("expected zero SDLC iterations to remain zero, got %d", dom.Code.SDLCIterations)
		}
	})

	t.Run("canonical DTO decoding does not fabricate omitted invocation values", func(t *testing.T) {
		raw := `{
  "updated_at": "2026-10-17T12:00:00Z",
  "sessions": [
    {
      "id": "inv-1",
      "invocations": [
        {}
      ],
      "tools": [],
      "events": []
    },
    {
      "id": "fixture-session",
      "invocations": [],
      "tools": [
        {
            "name": "bash",
          "calls": 5
        }
      ],
      "events": []
    }
  ]
}`
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		dom := dto.ToDomain()
		if len(dom.InvocationFacts()) != 1 {
			t.Fatalf("expected 1 invocation, got %d", len(dom.InvocationFacts()))
		}
		if dom.InvocationFacts()[0].Role != "" {
			t.Errorf("expected omitted role to remain empty, got %s", dom.InvocationFacts()[0].Role)
		}
		if dom.InvocationFacts()[0].Model != "" {
			t.Errorf("expected omitted model to remain empty, got %s", dom.InvocationFacts()[0].Model)
		}
		if dom.Sessions[0].User != "" {
			t.Errorf("expected omitted user to remain empty, got %s", dom.Sessions[0].User)
		}
		if len(dom.ToolFacts()) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(dom.ToolFacts()))
		}
		if dom.ToolFacts()[0].Name != "bash" {
			t.Errorf("expected tool name preserved, got %s", dom.ToolFacts()[0].Name)
		}
	})

	t.Run("migration metadata round-trip DTO serialization and unmarshaling", func(t *testing.T) {
		raw := `{
			"schema_version": 2,
			"ticket_id": "AL-SCHEMA-TEST",
			"updated_at": "2026-10-17T12:00:00Z",
			"migration": {
				"schema_version": 2,
				"status": "partial",
				"steps_completed": ["git_code_facts"],
				"steps_pending": ["opencode_db"],
				"last_migrated_at": "2026-10-17T12:00:00Z",
				"last_migrated_by": "engineer-1",
				"notes": "Testing boundary serialization"
			},
			"code": {
				"lines_added": 10,
				"sdlc_iterations": 1
			}
		}`

		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if dto.SchemaVersion == nil || *dto.SchemaVersion != 2 {
			t.Fatalf("expected SchemaVersion 2, got %v", dto.SchemaVersion)
		}
		if dto.Migration == nil || dto.Migration.Status != "partial" {
			t.Fatalf("expected migration status partial, got %v", dto.Migration)
		}

		dom := dto.ToDomain()
		if dom.Migration == nil {
			t.Fatal("expected domain migration to not be nil")
		}
		if dom.Migration.Status != domain.MigrationStatusPartial {
			t.Errorf("expected domain status partial, got %s", dom.Migration.Status)
		}
		if len(dom.Migration.StepsPending) != 1 || dom.Migration.StepsPending[0] != "opencode_db" {
			t.Errorf("expected pending opencode_db, got %v", dom.Migration.StepsPending)
		}

		// Round-trip back to DTO
		roundDTO := adapters.AgenticFactLedgerDTOFromDomain(dom)
		data, err := json.Marshal(roundDTO)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("re-unmarshal error: %v", err)
		}
		if parsed.Migration == nil || parsed.Migration.Status != "partial" {
			t.Errorf("expected round-trip migration status partial, got %v", parsed.Migration)
		}
		if parsed.Migration.LastMigratedBy != "engineer-1" {
			t.Errorf("expected engineer-1, got %s", parsed.Migration.LastMigratedBy)
		}
	})
}

func TestSerializationDTO(t *testing.T) {
	tid := "AL-1"
	ledger := domain.AgenticMetricsLedger{
		TicketID: &tid,
		Totals: domain.GroupedMetric{
			Invocations: 1,
			CostUSD:     0.5,
		},
	}

	serialized := adapters.SerializeLedger(ledger)
	data, err := json.Marshal(serialized)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"ticket_id":"AL-1"`) {
		t.Errorf("expected ticket_id in json, got %s", string(data))
	}
}

func TestAdversarialBoundarySchemaAndRepository(t *testing.T) {
	t.Run("malformed and edge JSON unmarshaling", func(t *testing.T) {
		hostileJSONs := []struct {
			name        string
			payload     string
			expectError bool
		}{
			{"empty string", "", true},
			{"truncated json", `{"ticket_id": "T-1", "code": {`, true},
			{"json array root", `[{"ticket_id": "T-1"}]`, true},
			{"bad type for number", `{"ticket_id": "T-1", "code": {"lines_added": "not_a_number"}}`, true},
			{"bad type for invocations", `{"invocations": "should_be_array"}`, true},
			{"empty json object", `{}`, false},
			{"negative token counts", `{
  "updated_at": "2026-10-17T12:00:00Z",
  "sessions": [
    {
      "id": "fixture-session",
      "invocations": [
        {
          "tokens": {
            "input": -50
          }
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`, false},
		}

		for _, tc := range hostileJSONs {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				var dto adapters.AgenticFactLedgerDTO
				err := json.Unmarshal([]byte(tc.payload), &dto)
				if tc.expectError && err == nil {
					t.Errorf("expected error for payload %q, got nil", tc.payload)
				}
				if !tc.expectError && err != nil {
					t.Errorf("expected no error for payload %q, got: %v", tc.payload, err)
				}
			})
		}
	})

	t.Run("scan workspaces with mixed valid, empty, non-metric, and zero-byte files", func(t *testing.T) {
		tempDir := t.TempDir()

		// 1. Valid workspace
		validDir := filepath.Join(tempDir, "VALID-1")
		if err := os.MkdirAll(validDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(validDir, "agentic_metrics.json"), []byte(`{"schema_version":4,"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"ticket_id":"VALID-1","updated_at":"2026-10-17T10:00:00Z","sessions":[]}`), 0644); err != nil {
			t.Fatal(err)
		}

		// 2. Empty directory (no files at all)
		emptyDir := filepath.Join(tempDir, "EMPTY-DIR")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatal(err)
		}

		// 3. Directory with only non-metric files
		unrelatedDir := filepath.Join(tempDir, "UNRELATED")
		if err := os.MkdirAll(unrelatedDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(unrelatedDir, "notes.txt"), []byte("some notes"), 0644); err != nil {
			t.Fatal(err)
		}

		// 4. Zero-byte metric file
		zeroByteDir := filepath.Join(tempDir, "ZERO-BYTE")
		if err := os.MkdirAll(zeroByteDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(zeroByteDir, "agentic_metrics.json"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		// 5. Corrupted JSON file
		corruptDir := filepath.Join(tempDir, "CORRUPT-JSON")
		if err := os.MkdirAll(corruptDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(corruptDir, "agentic_metrics.json"), []byte("{not json!"), 0644); err != nil {
			t.Fatal(err)
		}

		report := adapters.ScanWorkspaces(tempDir)

		// Only VALID-1 should be parsed as a valid ledger
		if len(report.Ledgers) != 1 {
			t.Fatalf("expected 1 valid ledger, got %d", len(report.Ledgers))
		}
		if *report.Ledgers[0].TicketID != "VALID-1" {
			t.Errorf("expected VALID-1, got %s", *report.Ledgers[0].TicketID)
		}

		// Should have 2 scan errors: one for ZERO-BYTE and one for CORRUPT-JSON
		if len(report.Errors) != 2 {
			t.Errorf("expected 2 scan errors for defective files, got %d: %v", len(report.Errors), report.Errors)
		}
	})

	t.Run("hostile path traversal attempts on ReadTicketArtifacts", func(t *testing.T) {
		tempDir := t.TempDir()
		safeTicketDir := filepath.Join(tempDir, "AL-SAFE")
		if err := os.MkdirAll(safeTicketDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(safeTicketDir, "CONTRACT.md"), []byte("# Safe Contract"), 0644); err != nil {
			t.Fatal(err)
		}

		traversalPayloads := []string{
			"../../../../../../etc/passwd",
			"..",
			".",
			"",
			"/root",
			"AL-SAFE/../AL-SAFE/../../",
			"../../../",
			"./../AL-SAFE",
			"AL-SAFE/../../",
		}

		for _, payload := range traversalPayloads {
			art := adapters.ReadTicketArtifacts(payload, tempDir)
			if art.ContractContent != nil || art.SummaryContent != nil {
				t.Errorf("path traversal payload %q succeeded in reading artifacts: %+v", payload, art)
			}
		}
	})
}

func TestEventFactDTOAndRepositoryEventHandling(t *testing.T) {
	temp := t.TempDir()
	ws := filepath.Join(temp, "AL-COMPACT")
	if err := os.MkdirAll(ws, 0700); err != nil {
		t.Fatal(err)
	}
	content := `
{
  "schema_version": 4,
  "migration": {
    "schema_version": 4,
    "status": "completed",
    "steps_completed": [
      "source_backfill"
    ],
    "steps_pending": []
  },
  "ticket_id": "AL-COMPACT",
  "sessions": [
    {
      "id": "s1",
      "source": "opencode",
      "user": "u",
      "invocations": [
        {
          "role": "coder",
          "model": "m",
          "active_duration_seconds": 1,
          "turns": 1,
          "tokens": {
            "input": 10,
            "output": 2,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 10
          },
          "cost_usd": 0,
          "compactions": 0,
          "user_iterations": 0,
          "started_at": null,
          "completed_at": null
        }
      ],
      "tools": [
        {
          "name": "bash",
          "calls": 8,
          "errors": 6
        }
      ],
      "events": [
        {
          "event_id": "e1",
          "timestamp": "2026-10-26T10:00:00Z",
          "turn": 1,
          "role": "coder",
          "event_type": "tool_call",
          "tool_name": "bash",
          "status": "error"
        }
      ],
      "anomalies": []
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(ws, "agentic_metrics.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	ledger, err := adapters.LoadTicketWorkspace("AL-COMPACT", temp)
	if err != nil {
		t.Fatal(err)
	}
	events := ledger.Facts.EventFacts()
	if len(events) != 1 || events[0].EventID != "e1" || events[0].SessionID != "s1" || events[0].Status != "error" || events[0].ToolName == nil || *events[0].ToolName != "bash" {
		t.Fatalf("accepted nested v4 facts were lost on targeted read: %#v", events)
	}
	if len(ledger.Facts.ToolFacts()) != 1 || ledger.Facts.ToolFacts()[0].Calls != 8 {
		t.Fatalf("tool totals changed: %#v", ledger.Facts.ToolFacts())
	}
	if len(ledger.Facts.AnomalyFacts()) == 0 {
		t.Fatal("expected normalized summaries alongside recorded event rows")
	}
	report := adapters.ScanWorkspaces(temp)
	if len(report.Diagnostics) != 0 || len(report.Ledgers) != 1 {
		t.Fatalf("unexpected workspace scan result: %#v", report)
	}
	scanned, err := json.Marshal(report.Ledgers[0].Facts)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := json.Marshal(ledger.Facts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(scanned, loaded) {
		t.Fatalf("scan and targeted loading disagree: %s != %s", scanned, loaded)
	}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger.Facts))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"events"`) || strings.Contains(string(encoded), `"evidence"`) {
		t.Fatalf("canonical encoding retained rows or refs: %s", encoded)
	}
	after, err := os.ReadFile(filepath.Join(ws, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, []byte(content)) {
		t.Fatal("reading or converting the ledger rewrote source bytes")
	}
}

func TestLoadTicketWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	ticketDir := filepath.Join(tempDir, "AL-TARGET")
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		t.Fatal(err)
	}

	factJSON := `
{
  "schema_version": 4,
  "migration": {
    "schema_version": 4,
    "status": "completed",
    "steps_completed": [
      "source_backfill"
    ],
    "steps_pending": []
  },
  "ticket_id": "AL-TARGET",
  "ticket_title": "Target Feature",
  "updated_at": "2026-10-18T12:00:00Z",
  "code": {
    "lines_added": 40,
    "lines_deleted": 5,
    "files_changed": 3,
    "sdlc_iterations": 2
  },
  "sessions": [
    {
      "id": "ses-1",
      "source": "opencode",
      "user": "test-dev",
      "invocations": [
        {
          "role": "orchestrator",
          "model": "model-1",
          "active_duration_seconds": 90,
          "turns": 3,
          "tokens": {
            "input": 2000,
            "output": 300,
            "cache_read": 1000,
            "peak_context": 25000,
            "reasoning": 0
          },
          "cost_usd": 0.08,
          "compactions": 0,
          "started_at": null,
          "completed_at": null,
          "user_iterations": 0
        }
      ],
      "tools": [
        {
          "name": "edit",
          "calls": 4,
          "errors": 0
        }
      ],
      "events": [],
      "anomalies": []
    }
  ]
}
`

	if err := os.WriteFile(filepath.Join(ticketDir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("successful load and derivation", func(t *testing.T) {
		ledger, err := adapters.LoadTicketWorkspace("AL-TARGET", tempDir)
		if err != nil {
			t.Fatalf("unexpected error loading workspace: %v", err)
		}
		if ledger == nil {
			t.Fatal("expected non-nil ledger")
		}
		if ledger.TicketID == nil || *ledger.TicketID != "AL-TARGET" {
			t.Errorf("expected ticket ID AL-TARGET, got %v", ledger.TicketID)
		}
		if ledger.Totals.Invocations != 1 {
			t.Errorf("expected 1 invocation, got %d", ledger.Totals.Invocations)
		}
		if ledger.WorkflowEfficiency.CodeDelivered.LinesAdded != 40 {
			t.Errorf("expected 40 lines added, got %d", ledger.WorkflowEfficiency.CodeDelivered.LinesAdded)
		}
	})

	t.Run("v4 ledger with legacy unscoped tools directs user to migrate", func(t *testing.T) {
		badDir := filepath.Join(tempDir, "AL-UNSCOPED")
		if err := os.MkdirAll(badDir, 0755); err != nil {
			t.Fatal(err)
		}
		badLedger := `{"schema_version":4,"tools":[{"name":"edit","calls":1,"errors":0}]}`
		if err := os.WriteFile(filepath.Join(badDir, "agentic_metrics.json"), []byte(badLedger), 0644); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-UNSCOPED", tempDir)
		if err == nil {
			t.Fatal("expected unscoped v4 tool facts to be rejected")
		}
		if ledger != nil {
			t.Fatalf("expected nil ledger for invalid v4 tool facts, got %+v", ledger)
		}
		if !strings.Contains(err.Error(), "agentic-lens migrate") {
			t.Errorf("expected actionable migration guidance, got %v", err)
		}
	})

	t.Run("fallback to parameter ticket ID when omitted in json", func(t *testing.T) {
		noIDDir := filepath.Join(tempDir, "AL-OMITTED")
		if err := os.MkdirAll(noIDDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(noIDDir, "agentic_metrics.json"), []byte(`{"schema_version":4,"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"updated_at":"2026-10-18T12:00:00Z","sessions":[]}`), 0644); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-OMITTED", tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ledger == nil || ledger.TicketID == nil || *ledger.TicketID != "AL-OMITTED" {
			t.Errorf("expected fallback ticket ID AL-OMITTED, got %v", ledger)
		}
	})

	t.Run("non-existent workspace returns error", func(t *testing.T) {
		ledger, err := adapters.LoadTicketWorkspace("NON-EXISTENT", tempDir)
		if err == nil {
			t.Fatal("expected error for non-existent ticket workspace, got nil")
		}
		if ledger != nil {
			t.Errorf("expected nil ledger, got %+v", ledger)
		}
	})

	t.Run("unconfigured workspaces directory returns error", func(t *testing.T) {
		noCfgDir := t.TempDir()
		cfgContent := `{"project_name": "Test", "ticket": {"prefixes": [{"prefix": "TEST"}]}}`
		if err := os.WriteFile(filepath.Join(noCfgDir, ".agenticlens-config.json"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(noCfgDir); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("TEST-1", "")
		if err == nil || !strings.Contains(err.Error(), "workspaces directory is not configured") {
			t.Fatalf("expected 'workspaces directory is not configured' error, got %v", err)
		}
		if ledger != nil {
			t.Errorf("expected nil ledger, got %+v", ledger)
		}
	})

	t.Run("corrupted metrics JSON returns error", func(t *testing.T) {
		corruptDir := filepath.Join(tempDir, "AL-CORRUPT")
		if err := os.MkdirAll(corruptDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(corruptDir, "agentic_metrics.json"), []byte("{bad json"), 0644); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-CORRUPT", tempDir)
		if err == nil {
			t.Fatal("expected error for corrupted JSON, got nil")
		}
		if ledger != nil {
			t.Errorf("expected nil ledger, got %+v", ledger)
		}
	})

	t.Run("rejects path traversal attempts", func(t *testing.T) {
		hostileIDs := []string{
			"",
			"   ",
			"..",
			"../AL-TARGET",
			"../../etc/passwd",
			"/etc/passwd",
			"\\windows\\win.ini",
			"AL-TARGET/../../",
			"AL-TARGET/nested",
			"AL-TARGET\\nested",
		}

		for _, hid := range hostileIDs {
			ledger, err := adapters.LoadTicketWorkspace(hid, tempDir)
			if err == nil {
				t.Errorf("expected error for hostile ticket ID %q, got nil", hid)
			}
			if ledger != nil {
				t.Errorf("expected nil ledger for hostile ticket ID %q, got %+v", hid, ledger)
			}
		}
	})
}

func TestUpdateSummaryFile(t *testing.T) {
	tempDir := t.TempDir()
	summaryFile := filepath.Join(tempDir, "SUMMARY.md")

	tableMD := "### Workflow Efficiency\n- **Iterations:** 1\n\n### Anomalies & Health Flags\n- **Status:** No execution anomalies detected (clean session).\n"

	t.Run("append to new file", func(t *testing.T) {
		err := adapters.UpdateSummaryFile(summaryFile, tableMD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content, _ := os.ReadFile(summaryFile)
		if !strings.Contains(string(content), tableMD) {
			t.Errorf("expected table in new file, got %s", string(content))
		}
	})

	t.Run("idempotent replacement in existing file with other sections", func(t *testing.T) {
		initial := "# Title\n\n### Workflow Efficiency\n- Old Content\n\n### Anomalies & Health Flags\n- Old Anomaly\n\n## Next Section\nSomething else\n"
		if err := os.WriteFile(summaryFile, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		newTable := "### Workflow Efficiency\n- New Content\n\n### Anomalies & Health Flags\n- **Status:** 1 execution anomaly detected (1 CRITICAL).\n"
		err := adapters.UpdateSummaryFile(summaryFile, newTable)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := os.ReadFile(summaryFile)
		updatedStr := string(updated)
		if strings.Contains(updatedStr, "Old Content") || strings.Contains(updatedStr, "Old Anomaly") {
			t.Errorf("expected Old Content and Old Anomaly to be replaced")
		}
		if !strings.Contains(updatedStr, "New Content") {
			t.Errorf("expected New Content to be present")
		}
		if !strings.Contains(updatedStr, "1 execution anomaly detected") {
			t.Errorf("expected new anomaly status to be present")
		}
		if !strings.Contains(updatedStr, "## Next Section") {
			t.Errorf("expected Next Section to be preserved")
		}
		if strings.Count(updatedStr, "### Workflow Efficiency") != 1 {
			t.Errorf("expected exactly 1 Workflow Efficiency header, got %d", strings.Count(updatedStr, "### Workflow Efficiency"))
		}
	})

	t.Run("replaces legacy header cleanly", func(t *testing.T) {
		legacyFile := filepath.Join(tempDir, "LEGACY_SUMMARY.md")
		initialLegacy := "# Summary\n\n### Agentic Execution Metrics\n- Old metrics\n\n## Results\nDone\n"
		if err := os.WriteFile(legacyFile, []byte(initialLegacy), 0644); err != nil {
			t.Fatal(err)
		}

		err := adapters.UpdateSummaryFile(legacyFile, tableMD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(legacyFile)
		contentStr := string(content)
		if strings.Contains(contentStr, "Old metrics") {
			t.Errorf("expected legacy metrics to be replaced")
		}
		if !strings.Contains(contentStr, "### Workflow Efficiency") {
			t.Errorf("expected new Workflow Efficiency header")
		}
		if !strings.Contains(contentStr, "## Results") {
			t.Errorf("expected subsequent section to be preserved")
		}
	})

	t.Run("strict idempotency across repeated sequential updates", func(t *testing.T) {
		idempotentFile := filepath.Join(tempDir, "IDEMPOTENT_SUMMARY.md")
		initial := "# Work Summary\n\nExisting work log.\n\n### Workflow Efficiency\n- Initial Efficiency\n\n## Future Notes\nPreserved.\n"
		if err := os.WriteFile(idempotentFile, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		// Perform 5 successive updates with identical tableMD
		for i := 0; i < 5; i++ {
			err := adapters.UpdateSummaryFile(idempotentFile, tableMD)
			if err != nil {
				t.Fatalf("update %d failed: %v", i+1, err)
			}
		}

		data, err := os.ReadFile(idempotentFile)
		if err != nil {
			t.Fatal(err)
		}
		result := string(data)

		// Assert exactly one occurrence of each heading
		if strings.Count(result, "### Workflow Efficiency") != 1 {
			t.Errorf("expected exactly 1 occurrence of '### Workflow Efficiency', got %d", strings.Count(result, "### Workflow Efficiency"))
		}
		if strings.Count(result, "### Anomalies & Health Flags") != 1 {
			t.Errorf("expected exactly 1 occurrence of '### Anomalies & Health Flags', got %d", strings.Count(result, "### Anomalies & Health Flags"))
		}
		if !strings.Contains(result, "Existing work log.") {
			t.Errorf("expected preamble to be preserved")
		}
		if !strings.Contains(result, "## Future Notes\nPreserved.") {
			t.Errorf("expected trailing section to be preserved")
		}
	})
}

func TestSaveFactLedger(t *testing.T) {
	t.Parallel()

	t.Run("creates nested directories, writes indented JSON with trailing newline", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outPath := filepath.Join(tempDir, "sub", "dir", "facts.json")

		tid := "AL-PERSIST-1"
		title := "Test Persistence"
		schemaVer := 2
		ledger := domain.AgenticFactLedger{
			SchemaVersion: &schemaVer,
			TicketID:      &tid,
			TicketTitle:   &title, Sessions: sessionFactsWithInvs([]testInvFact{
				{
					SessionID: "inv-1",
					Role:      "coder",
					Model:     "gemini-3.8-flash",
					CostUSD:   0.05,
				},
			}, nil, nil),
		}

		err := adapters.SaveFactLedger(outPath, ledger)
		if err != nil {
			t.Fatalf("unexpected error saving fact ledger: %v", err)
		}

		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed reading saved file: %v", err)
		}

		// Trailing newline check
		if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Errorf("expected trailing newline in output")
		}

		// Two-space indentation check
		content := string(data)
		if !strings.Contains(content, "  \"ticket_id\": \"AL-PERSIST-1\"") {
			t.Errorf("expected 2-space indented JSON, got:\n%s", content)
		}

		// Unmarshal round-trip check
		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed unmarshaling saved fact ledger: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-PERSIST-1" {
			t.Errorf("expected ticket_id AL-PERSIST-1, got %v", parsed.TicketID)
		}
		if len(parsed.Sessions) != 1 || parsed.Sessions[0].ID != "inv-1" || len(parsed.Sessions[0].Invocations) != 1 || parsed.Sessions[0].Invocations[0].Model != "gemini-3.8-flash" {
			t.Errorf("expected invocation and owning session preserved, got %+v", parsed.Sessions)
		}
	})

	t.Run("returns wrapped error when directory cannot be created", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		blockingFile := filepath.Join(tempDir, "blocker")
		if err := os.WriteFile(blockingFile, []byte("file"), 0644); err != nil {
			t.Fatal(err)
		}

		// Attempting to create directory inside a regular file will fail
		invalidPath := filepath.Join(blockingFile, "sub", "facts.json")
		tid := "AL-ERR-1"
		title := "Test Error Dir"
		ledger := domain.AgenticFactLedger{
			TicketID:    &tid,
			TicketTitle: &title,
		}
		err := adapters.SaveFactLedger(invalidPath, ledger)
		if err == nil {
			t.Fatal("expected error when parent directory cannot be created, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create directory") {
			t.Errorf("expected 'failed to create directory' in error message, got: %v", err)
		}
	})

	t.Run("fails defensive validation when TicketID or TicketTitle is missing or empty", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outPath := filepath.Join(tempDir, "facts.json")

		validID := "AL-VALID"
		validTitle := "Valid Title"
		emptyStr := ""
		whitespaceStr := "   \t\n"

		cases := []struct {
			name        string
			ticketID    *string
			ticketTitle *string
			errContains string
		}{
			{"nil ticket ID", nil, &validTitle, "TicketID is required and cannot be empty"},
			{"empty ticket ID", &emptyStr, &validTitle, "TicketID is required and cannot be empty"},
			{"whitespace ticket ID", &whitespaceStr, &validTitle, "TicketID is required and cannot be empty"},
			{"nil ticket title", &validID, nil, "TicketTitle is required and cannot be empty"},
			{"empty ticket title", &validID, &emptyStr, "TicketTitle is required and cannot be empty"},
			{"whitespace ticket title", &validID, &whitespaceStr, "TicketTitle is required and cannot be empty"},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ledger := domain.AgenticFactLedger{
					TicketID:    tc.ticketID,
					TicketTitle: tc.ticketTitle,
				}
				err := adapters.SaveFactLedger(outPath, ledger)
				if err == nil {
					t.Fatalf("expected error for %s, got nil", tc.name)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
			})
		}
	})
}

func TestSaveMetricsLedger(t *testing.T) {
	t.Parallel()

	t.Run("creates nested directories, serializes DTO and writes formatted JSON", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outPath := filepath.Join(tempDir, "metrics", "computed.json")

		tid := "AL-METRICS-1"
		title := "Test Metrics Persistence"
		ledger := domain.AgenticMetricsLedger{
			TicketID:    &tid,
			TicketTitle: &title,
			Totals: domain.GroupedMetric{
				Invocations: 3,
				InputTokens: 5000,
				CostUSD:     0.1234,
			},
		}

		err := adapters.SaveMetricsLedger(outPath, ledger)
		if err != nil {
			t.Fatalf("unexpected error saving metrics ledger: %v", err)
		}

		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed reading saved file: %v", err)
		}

		// Trailing newline check
		if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Errorf("expected trailing newline in output")
		}

		// Two-space indentation check
		content := string(data)
		if !strings.Contains(content, "  \"ticket_id\": \"AL-METRICS-1\"") {
			t.Errorf("expected 2-space indented JSON, got:\n%s", content)
		}

		// DTO deserialization check
		var parsed adapters.SerializedLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed unmarshaling saved metrics ledger: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-METRICS-1" {
			t.Errorf("expected ticket_id AL-METRICS-1, got %v", parsed.TicketID)
		}
		if parsed.Totals.Invocations != 3 {
			t.Errorf("expected 3 invocations, got %d", parsed.Totals.Invocations)
		}
	})

	t.Run("returns wrapped error when directory cannot be created", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		blockingFile := filepath.Join(tempDir, "blocker")
		if err := os.WriteFile(blockingFile, []byte("file"), 0644); err != nil {
			t.Fatal(err)
		}

		invalidPath := filepath.Join(blockingFile, "nested", "metrics.json")
		err := adapters.SaveMetricsLedger(invalidPath, domain.AgenticMetricsLedger{})
		if err == nil {
			t.Fatal("expected error when parent directory cannot be created, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create directory") {
			t.Errorf("expected 'failed to create directory' in error message, got: %v", err)
		}
	})
}

func TestSaveLedgers_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("SaveFactLedger fails when destination path is an existing directory", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		dirAsFile := filepath.Join(tempDir, "existing_dir")
		if err := os.MkdirAll(dirAsFile, 0755); err != nil {
			t.Fatal(err)
		}

		tid := "AL-ERR-DIR"
		title := "Test Dir Failure"
		err := adapters.SaveFactLedger(dirAsFile, domain.AgenticFactLedger{
			TicketID:    &tid,
			TicketTitle: &title,
		})
		if err == nil {
			t.Fatal("expected error writing fact ledger to existing directory, got nil")
		}
		if !strings.Contains(err.Error(), "failed to write fact ledger") {
			t.Errorf("expected 'failed to write fact ledger' in error, got: %v", err)
		}
	})

	t.Run("SaveMetricsLedger fails when destination path is an existing directory", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		dirAsFile := filepath.Join(tempDir, "existing_metrics_dir")
		if err := os.MkdirAll(dirAsFile, 0755); err != nil {
			t.Fatal(err)
		}

		err := adapters.SaveMetricsLedger(dirAsFile, domain.AgenticMetricsLedger{})
		if err == nil {
			t.Fatal("expected error writing metrics ledger to existing directory, got nil")
		}
		if !strings.Contains(err.Error(), "failed to write metrics ledger") {
			t.Errorf("expected 'failed to write metrics ledger' in error, got: %v", err)
		}
	})

	t.Run("SaveFactLedger overwrites existing file cleanly without appending or corrupting", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "overwrite_test", "agentic_metrics.json")

		// First write: large payload
		longTitle := strings.Repeat("Initial Large Payload Content ", 10)
		tid := "AL-OVERWRITE-1"
		firstLedger := domain.AgenticFactLedger{
			TicketID:    &tid,
			TicketTitle: &longTitle,
		}
		if err := adapters.SaveFactLedger(filePath, firstLedger); err != nil {
			t.Fatalf("first write failed: %v", err)
		}

		// Second write: tiny payload
		shortTitle := "Short"
		secondLedger := domain.AgenticFactLedger{
			TicketID:    &tid,
			TicketTitle: &shortTitle,
		}
		if err := adapters.SaveFactLedger(filePath, secondLedger); err != nil {
			t.Fatalf("second write failed: %v", err)
		}

		// Read back and unmarshal to prove file is valid JSON and not corrupt trailing bytes
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("reading file failed: %v", err)
		}
		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshaling overwritten file failed: %v, content:\n%s", err, string(data))
		}
		if parsed.TicketTitle == nil || *parsed.TicketTitle != "Short" {
			t.Errorf("expected overwritten title 'Short', got: %v", parsed.TicketTitle)
		}
	})

	t.Run("SaveMetricsLedger overwrites existing file cleanly without appending or corrupting", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "overwrite_metrics_test", "metrics.json")

		tid1 := "AL-INITIAL-LONG-TICKET-ID-12345"
		firstLedger := domain.AgenticMetricsLedger{
			TicketID: &tid1,
			Totals: domain.GroupedMetric{
				Invocations: 999999,
				InputTokens: 888888,
			},
		}
		if err := adapters.SaveMetricsLedger(filePath, firstLedger); err != nil {
			t.Fatalf("first write failed: %v", err)
		}

		tid2 := "AL-2"
		secondLedger := domain.AgenticMetricsLedger{
			TicketID: &tid2,
			Totals: domain.GroupedMetric{
				Invocations: 1,
			},
		}
		if err := adapters.SaveMetricsLedger(filePath, secondLedger); err != nil {
			t.Fatalf("second write failed: %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("reading file failed: %v", err)
		}
		var parsed adapters.SerializedLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("unmarshaling overwritten file failed: %v, content:\n%s", err, string(data))
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-2" {
			t.Errorf("expected ticket_id AL-2, got %v", parsed.TicketID)
		}
		if parsed.Totals.Invocations != 1 {
			t.Errorf("expected 1 invocation, got %d", parsed.Totals.Invocations)
		}
	})

	t.Run("zero-value ledgers marshal to valid JSON with trailing newline", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		factPath := filepath.Join(tempDir, "zero_facts.json")
		tid := "AL-ZERO"
		title := "Zero Facts"
		if err := adapters.SaveFactLedger(factPath, domain.AgenticFactLedger{TicketID: &tid, TicketTitle: &title}); err != nil {
			t.Fatalf("saving zero fact ledger failed: %v", err)
		}
		factData, err := os.ReadFile(factPath)
		if err != nil {
			t.Fatal(err)
		}
		if len(factData) == 0 || factData[len(factData)-1] != '\n' {
			t.Errorf("expected trailing newline in zero fact ledger")
		}

		metricsPath := filepath.Join(tempDir, "zero_metrics.json")
		if err := adapters.SaveMetricsLedger(metricsPath, domain.AgenticMetricsLedger{}); err != nil {
			t.Fatalf("saving zero metrics ledger failed: %v", err)
		}
		metricsData, err := os.ReadFile(metricsPath)
		if err != nil {
			t.Fatal(err)
		}
		if len(metricsData) == 0 || metricsData[len(metricsData)-1] != '\n' {
			t.Errorf("expected trailing newline in zero metrics ledger")
		}
	})
}

func TestWorkspaceLoad_ConfiguredRoleMappings(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create root .agenticlens-config.yaml with custom role mappings
	cfgContent := `
project_name: TestProject
ticket:
  prefixes:
    - prefix: AL
workspaces:
  dir: workspaces
code_metrics:
  source_roots:
    - src
anomaly_thresholds:
  context_ceiling_warning_tokens: 160000
  context_ceiling_critical_tokens: 200000
role_mappings:
  build: orchestrator
  qa-bot: tester
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Create workspace with agentic_metrics.json containing raw roles 'build' and 'qa-bot'
	wsDir := filepath.Join(tempDir, "workspaces", "AL-24")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatal(err)
	}

	factJSON := `
{
  "schema_version": 4,
  "migration": {
    "schema_version": 4,
    "status": "completed",
    "steps_completed": [
      "source_backfill"
    ],
    "steps_pending": []
  },
  "ticket_id": "AL-24",
  "ticket_title": "Role Mapping Test",
  "updated_at": "2026-10-18T12:00:00Z",
  "sessions": [
    {
      "id": "ses-build",
      "source": "opencode",
      "user": "fixture-user",
      "invocations": [
        {
          "role": "build",
          "model": "claude-3-sonnet",
          "cost_usd": 1.5,
          "turns": 3,
          "started_at": null,
          "completed_at": null,
          "active_duration_seconds": 0,
          "tokens": {
            "input": 0,
            "output": 0,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [],
      "events": [],
      "anomalies": []
    },
    {
      "id": "ses-qa",
      "source": "opencode",
      "user": "fixture-user",
      "parent_id": "ses-build",
      "invocations": [
        {
          "role": "qa-bot",
          "model": "claude-3-haiku",
          "cost_usd": 0.5,
          "turns": 2,
          "started_at": null,
          "completed_at": null,
          "active_duration_seconds": 0,
          "tokens": {
            "input": 0,
            "output": 0,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [],
      "events": [],
      "anomalies": []
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(wsDir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
		t.Fatal(err)
	}

	workspacesRoot := filepath.Join(tempDir, "workspaces")

	t.Run("LoadTicketWorkspace applies role mappings while preserving raw facts", func(t *testing.T) {
		ledger, err := adapters.LoadTicketWorkspace("AL-24", workspacesRoot)
		if err != nil {
			t.Fatalf("LoadTicketWorkspace failed: %v", err)
		}

		// Roles should be mapped: orchestrator and tester
		if len(ledger.Roles) != 2 {
			t.Fatalf("expected 2 roles, got %d: %+v", len(ledger.Roles), ledger.Roles)
		}
		if ledger.Roles[0].Role != "orchestrator" {
			t.Errorf("expected role 'orchestrator', got %s", ledger.Roles[0].Role)
		}
		if ledger.Roles[1].Role != "tester" {
			t.Errorf("expected role 'tester', got %s", ledger.Roles[1].Role)
		}

		// Facts should retain raw roles 'build' and 'qa-bot'
		if len(ledger.Facts.InvocationFacts()) != 2 {
			t.Fatalf("expected 2 raw invocations, got %d", len(ledger.Facts.InvocationFacts()))
		}
		if ledger.Facts.InvocationFacts()[0].Role != "build" {
			t.Errorf("expected raw invocation 0 role 'build', got %s", ledger.Facts.InvocationFacts()[0].Role)
		}
		if ledger.Facts.InvocationFacts()[1].Role != "qa-bot" {
			t.Errorf("expected raw invocation 1 role 'qa-bot', got %s", ledger.Facts.InvocationFacts()[1].Role)
		}

		// Subagents effort distribution should reflect tester with 100%
		eff := ledger.WorkflowEfficiency.RoleEffortDistributionPct
		if eff["tester"] != 100.0 {
			t.Errorf("expected tester effort 100%%, got: %+v", eff)
		}
	})

	t.Run("ScanWorkspaces applies role mappings across all workspaces", func(t *testing.T) {
		report := adapters.ScanWorkspaces(workspacesRoot)
		if len(report.Errors) > 0 {
			t.Fatalf("unexpected scan errors: %v", report.Errors)
		}
		if len(report.Ledgers) != 1 {
			t.Fatalf("expected 1 ledger, got %d", len(report.Ledgers))
		}

		scanned := report.Ledgers[0]
		if len(scanned.Roles) != 2 {
			t.Fatalf("expected 2 roles, got %d", len(scanned.Roles))
		}
		if scanned.Roles[0].Role != "orchestrator" || scanned.Roles[1].Role != "tester" {
			t.Errorf("expected orchestrator and tester, got %+v", scanned.Roles)
		}
		if scanned.Facts.InvocationFacts()[0].Role != "build" {
			t.Errorf("expected raw role build in Facts, got %s", scanned.Facts.InvocationFacts()[0].Role)
		}
	})
}

func TestWorkspaceLoad_ConfiguredRoleMappings_RootConfig(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create root .agenticlens-config.yaml with custom role mappings
	cfgContent := `
project_name: RootTestProject
ticket:
  prefixes:
    - prefix: AL
workspaces:
  dir: root_workspaces
code_metrics:
  source_roots:
    - src
anomaly_thresholds:
  context_ceiling_warning_tokens: 160000
  context_ceiling_critical_tokens: 200000
role_mappings:
  sys-engineer: coder
  lead-tester: tester
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Create workspace with agentic_metrics.json containing raw roles 'sys-engineer' and 'lead-tester'
	wsDir := filepath.Join(tempDir, "root_workspaces", "AL-38")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatal(err)
	}

	factJSON := `
{
  "schema_version": 4,
  "migration": {
    "schema_version": 4,
    "status": "completed",
    "steps_completed": [
      "source_backfill"
    ],
    "steps_pending": []
  },
  "ticket_id": "AL-38",
  "ticket_title": "Root Config Role Mapping Test",
  "updated_at": "2026-10-20T12:00:00Z",
  "sessions": [
    {
      "id": "ses-eng",
      "source": "opencode",
      "user": "fixture-user",
      "invocations": [
        {
          "role": "sys-engineer",
          "model": "claude-3-sonnet",
          "cost_usd": 1.5,
          "active_duration_seconds": 120,
          "started_at": null,
          "completed_at": null,
          "turns": 0,
          "tokens": {
            "input": 0,
            "output": 0,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [],
      "events": [],
      "anomalies": []
    },
    {
      "id": "ses-test",
      "source": "opencode",
      "user": "fixture-user",
      "invocations": [
        {
          "role": "lead-tester",
          "model": "gpt-4o",
          "cost_usd": 0.8,
          "active_duration_seconds": 60,
          "started_at": null,
          "completed_at": null,
          "turns": 0,
          "tokens": {
            "input": 0,
            "output": 0,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [],
      "events": [],
      "anomalies": []
    }
  ]
}
`
	if err := os.WriteFile(filepath.Join(wsDir, "agentic_metrics.json"), []byte(factJSON), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	workspacesRoot, err := adapters.ResolveWorkspacesDir("")
	if err != nil {
		t.Fatalf("unexpected error resolving workspaces: %v", err)
	}
	if workspacesRoot != "root_workspaces" {
		t.Fatalf("expected workspacesRoot root_workspaces, got %s", workspacesRoot)
	}

	report := adapters.ScanWorkspaces(workspacesRoot)
	if len(report.Errors) > 0 {
		t.Fatalf("unexpected scan errors: %v", report.Errors)
	}
	if len(report.Ledgers) != 1 {
		t.Fatalf("expected 1 ledger, got %d", len(report.Ledgers))
	}

	scanned := report.Ledgers[0]
	if len(scanned.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(scanned.Roles))
	}
	if scanned.Roles[0].Role != "coder" || scanned.Roles[1].Role != "tester" {
		t.Errorf("expected coder and tester, got %+v", scanned.Roles)
	}
	if scanned.Facts.InvocationFacts()[0].Role != "sys-engineer" {
		t.Errorf("expected raw role sys-engineer in Facts, got %s", scanned.Facts.InvocationFacts()[0].Role)
	}
}

func TestLoadTicketWorkspace_AdversarialResilience(t *testing.T) {
	tempDir := t.TempDir()
	wsDir := filepath.Join(tempDir, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("blocks path traversal attempts strictly", func(t *testing.T) {
		traversalPayloads := []string{
			"../../../../../../etc/passwd",
			"/etc/passwd",
			"..",
			".",
			"",
			"   ",
			"AL-24/../../",
			"../AL-24",
			"../../",
			"AL-24/../..",
			"..\\..\\win.ini",
		}

		for _, payload := range traversalPayloads {
			ledger, err := adapters.LoadTicketWorkspace(payload, tempDir)
			if err == nil {
				t.Errorf("expected error for traversal payload %q, got nil (ledger=%v)", payload, ledger)
			}
			if ledger != nil {
				t.Errorf("expected nil ledger for traversal payload %q, got %+v", payload, ledger)
			}
		}
	})

	t.Run("missing workspace returns descriptive wrapped error", func(t *testing.T) {
		ledger, err := adapters.LoadTicketWorkspace("NON-EXISTENT-TICKET", wsDir)
		if err == nil {
			t.Fatalf("expected error for non-existent ticket, got nil")
		}
		if ledger != nil {
			t.Fatalf("expected nil ledger for non-existent ticket")
		}
		if !strings.Contains(err.Error(), "reading metrics file") {
			t.Errorf("expected 'reading metrics file' in error, got %v", err)
		}
	})

	t.Run("zero-byte metrics file returns descriptive unmarshaling error", func(t *testing.T) {
		zeroDir := filepath.Join(wsDir, "AL-ZERO")
		if err := os.MkdirAll(zeroDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(zeroDir, "agentic_metrics.json"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-ZERO", wsDir)
		if err == nil {
			t.Fatalf("expected error for zero-byte metrics file, got nil")
		}
		if ledger != nil {
			t.Fatalf("expected nil ledger, got %+v", ledger)
		}
		if !strings.Contains(err.Error(), "decoding ledger metadata") {
			t.Errorf("expected 'decoding ledger metadata' in error, got: %v", err)
		}
	})

	t.Run("corrupted JSON metrics file returns descriptive error", func(t *testing.T) {
		corruptDir := filepath.Join(wsDir, "AL-CORRUPT")
		if err := os.MkdirAll(corruptDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(corruptDir, "agentic_metrics.json"), []byte("{corrupt json"), 0644); err != nil {
			t.Fatal(err)
		}

		ledger, err := adapters.LoadTicketWorkspace("AL-CORRUPT", wsDir)
		if err == nil {
			t.Fatalf("expected error for corrupt JSON metrics file, got nil")
		}
		if ledger != nil {
			t.Fatalf("expected nil ledger, got %+v", ledger)
		}
		if !strings.Contains(err.Error(), "decoding ledger metadata") {
			t.Errorf("expected 'decoding ledger metadata' in error, got: %v", err)
		}
	})

	t.Run("serialization round-trip preserves user_iterations exactly", func(t *testing.T) {
		roundTripPath := filepath.Join(wsDir, "roundtrip.json")
		schemaVer := domain.CurrentSchemaVersion
		ticketID := "AL-ROUNDTRIP"
		ticketTitle := "Roundtrip Title"
		original := domain.AgenticFactLedger{
			SchemaVersion: &schemaVer,
			TicketID:      &ticketID,
			TicketTitle:   &ticketTitle, Sessions: sessionFacts([]domain.InvocationFact{
				{
					Role:           domain.RoleRootAgent,
					Model:          "gpt-4o",
					UserIterations: 7,
				},
				{
					Role:           domain.RoleUnassignedSubAgent,
					Model:          "gpt-4o",
					UserIterations: 0,
				},
			}, nil, nil),
		}

		if err := adapters.SaveFactLedger(roundTripPath, original); err != nil {
			t.Fatalf("failed saving fact ledger: %v", err)
		}

		loaded, err := adapters.LoadExistingFactLedger(roundTripPath)
		if err != nil {
			t.Fatalf("failed loading fact ledger: %v", err)
		}
		if loaded == nil {
			t.Fatal("expected non-nil loaded ledger")
		}
		if len((domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()) != 2 {
			t.Fatalf("expected 2 invocations, got %d", len((domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()))
		}
		if (domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()[0].UserIterations != 7 {
			t.Errorf("expected invocation 0 user_iterations 7, got %d", (domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()[0].UserIterations)
		}
		if (domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()[1].UserIterations != 0 {
			t.Errorf("expected invocation 1 user_iterations 0, got %d", (domain.AgenticFactLedger{Sessions: loaded.Sessions}).InvocationFacts()[1].UserIterations)
		}
	})
}

func TestLedgerFileReadsNormalizeOptionalSkillTaskCollections(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fields   string
		wantRows int
		invalid  bool
	}{
		{name: "omitted"},
		{name: "null", fields: `,"skills":null,"task_links":null`},
		{name: "empty", fields: `,"skills":[],"task_links":[]`},
		{name: "repeated observations", fields: `,"skills":[{"name":"review","turn_index":1},{"name":"review","turn_index":1}],"task_links":[{"subagent_type":"coder","description":"same"},{"subagent_type":"coder","description":"same"}]`, wantRows: 2},
		{name: "malformed skills", fields: `,"skills":{}`, invalid: true},
		{name: "malformed tasks", fields: `,"task_links":"unknown"`, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			ticket := "CUSTOM-READ"
			dir := filepath.Join(root, ticket)
			if err := os.Mkdir(dir, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "agentic_metrics.json")
			data := []byte(`{"schema_version":4,"ticket_id":"CUSTOM-READ","sessions":[]` + tc.fields + `}`)
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			existing, existingErr := adapters.LoadExistingFactLedger(path)
			workspace, workspaceErr := adapters.LoadTicketWorkspace(ticket, root)
			if tc.invalid {
				if existingErr == nil || workspaceErr == nil {
					t.Fatalf("malformed collections accepted: existing=%v workspace=%v", existingErr, workspaceErr)
				}
				return
			}
			if existingErr != nil || workspaceErr != nil {
				t.Fatalf("compatible collection form rejected: existing=%v workspace=%v", existingErr, workspaceErr)
			}
			if existing == nil || workspace == nil {
				t.Fatal("file read returned no ledger")
			}
			if existing.Skills == nil || existing.TaskLinks == nil || workspace.Facts.Skills == nil || workspace.Facts.TaskLinks == nil {
				t.Fatalf("internal collections must be initialized after file read: existing=%+v workspace=%+v", existing, workspace.Facts)
			}
			if len(existing.Skills) != tc.wantRows || len(existing.TaskLinks) != tc.wantRows || len(workspace.Facts.Skills) != tc.wantRows || len(workspace.Facts.TaskLinks) != tc.wantRows {
				t.Fatalf("file read changed row multiplicity: existing skills/tasks=%d/%d workspace=%d/%d", len(existing.Skills), len(existing.TaskLinks), len(workspace.Facts.Skills), len(workspace.Facts.TaskLinks))
			}
			if tc.wantRows > 0 && (existing.Skills[1].Name != "review" || workspace.Facts.TaskLinks[1].Description != "same") {
				t.Fatal("populated observation payload changed")
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, data) {
				t.Fatal("read normalization rewrote optional wire fields")
			}
		})
	}
}
