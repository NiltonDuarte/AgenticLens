package adapters_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestResolveWorkspacesDir(t *testing.T) {
	t.Run("explicit override", func(t *testing.T) {
		got := adapters.ResolveWorkspacesDir("/custom/path")
		if got != "/custom/path" {
			t.Errorf("expected /custom/path, got %s", got)
		}
	})

	t.Run("config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		wsDir := filepath.Join(tempDir, "custom_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := `{"workspaces_dir": "custom_workspaces"}`
		if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}

		resolved := adapters.ResolveWorkspacesDir("")
		if resolved != wsDir {
			t.Errorf("expected %s, got %s", wsDir, resolved)
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
		"invocations": [
			{
				"id": "ses-1",
				"role": "orchestrator",
				"model": "model-x",
				"user": "test-user",
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
				"compactions": 0
			}
		],
		"tools": [
			{
				"name": "bash",
				"calls": 2,
				"errors": 0
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

	t.Run("valid read", func(t *testing.T) {
		art := adapters.ReadTicketArtifacts("TICKET-10", tempDir)
		if art.ContractContent == nil || *art.ContractContent != "# Contract" {
			t.Errorf("contract content mismatch: %v", art.ContractContent)
		}
		if art.SummaryContent == nil || *art.SummaryContent != "# Summary" {
			t.Errorf("summary content mismatch: %v", art.SummaryContent)
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
		factJSON := `{"updated_at": "2026-10-17T12:00:00Z"}`
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
		if dom.Code.SDLCIterations != 1 {
			t.Errorf("expected default SDLC iterations 1, got %d", dom.Code.SDLCIterations)
		}
	})

	t.Run("invocation and tool defaults when fields omitted", func(t *testing.T) {
		raw := `{
			"updated_at": "2026-10-17T12:00:00Z",
			"invocations": [
				{"id": "inv-1"}
			],
			"tools": [
				{"calls": 5}
			]
		}`
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		dom := dto.ToDomain()
		if len(dom.Invocations) != 1 {
			t.Fatalf("expected 1 invocation, got %d", len(dom.Invocations))
		}
		if dom.Invocations[0].Role != "orchestrator" {
			t.Errorf("expected default role orchestrator, got %s", dom.Invocations[0].Role)
		}
		if dom.Invocations[0].Model != "default" {
			t.Errorf("expected default model default, got %s", dom.Invocations[0].Model)
		}
		if dom.Invocations[0].User != "unknown" {
			t.Errorf("expected default user unknown, got %s", dom.Invocations[0].User)
		}
		if len(dom.Tools) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(dom.Tools))
		}
		if dom.Tools[0].Name != "unknown" {
			t.Errorf("expected default tool name unknown, got %s", dom.Tools[0].Name)
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
			{"negative token counts", `{"updated_at": "2026-10-17T12:00:00Z", "invocations": [{"tokens": {"input": -50}}]}`, false},
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
		if err := os.WriteFile(filepath.Join(validDir, "agentic_metrics.json"), []byte(`{"ticket_id": "VALID-1", "updated_at": "2026-10-17T10:00:00Z"}`), 0644); err != nil {
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
