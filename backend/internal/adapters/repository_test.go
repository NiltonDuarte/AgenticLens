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

	t.Run("yaml config file resolution", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatal(err)
		}
		wsDir := filepath.Join(tempDir, "yaml_workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		cfgContent := "workspaces:\n  dir: yaml_workspaces\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
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

func TestEventFactDTOAndRepositoryEventHandling(t *testing.T) {
	t.Run("legacy payload omitting events decodes without error", func(t *testing.T) {
		legacyJSON := `{
			"schema_version": 2,
			"ticket_id": "AL-LEGACY",
			"updated_at": "2026-10-18T10:00:00Z",
			"invocations": [
				{
					"id": "ses-legacy",
					"role": "orchestrator",
					"model": "gemini-3.8-flash",
					"user": "eng",
					"active_duration_seconds": 30,
					"turns": 2,
					"tokens": {"input": 100, "output": 20, "cache_read": 50, "reasoning": 0, "peak_context": 150},
					"cost_usd": 0.01,
					"compactions": 0
				}
			]
		}`

		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(legacyJSON), &dto); err != nil {
			t.Fatalf("failed to unmarshal legacy ledger without events: %v", err)
		}
		if dto.Events != nil {
			t.Errorf("expected dto.Events to be nil, got: %+v", dto.Events)
		}

		dom := dto.ToDomain()
		if len(dom.Events) != 0 {
			t.Errorf("expected empty domain events slice, got len %d", len(dom.Events))
		}
	})

	t.Run("ledger with events parses and round-trips cleanly", func(t *testing.T) {
		toolName := "bash"
		errMsg := "command not found: foobar"
		var durMs int64 = 150

		eventJSON := `{
			"schema_version": 2,
			"ticket_id": "AL-EVENTS",
			"updated_at": "2026-10-18T10:00:00Z",
			"events": [
				{
					"timestamp": "2026-10-18T10:00:01.100000+00:00",
					"session_id": "ses-1",
					"turn": 1,
					"role": "orchestrator",
					"event_type": "turn_message",
					"status": "completed",
					"duration_ms": 1200,
					"context_tokens": 5000
				},
				{
					"timestamp": "2026-10-18T10:00:02.300000+00:00",
					"session_id": "ses-1",
					"turn": 1,
					"role": "orchestrator",
					"event_type": "tool_call",
					"tool_name": "bash",
					"status": "error",
					"error_message": "command not found: foobar",
					"duration_ms": 150,
					"context_tokens": 5000
				},
				{
					"timestamp": "2026-10-18T10:00:03.000000+00:00",
					"session_id": "ses-1",
					"turn": 2,
					"role": "orchestrator",
					"event_type": "compaction",
					"status": "completed",
					"context_tokens": 4200
				}
			]
		}`

		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(eventJSON), &dto); err != nil {
			t.Fatalf("failed to unmarshal ledger with events: %v", err)
		}
		if len(dto.Events) != 3 {
			t.Fatalf("expected 3 events in DTO, got %d", len(dto.Events))
		}

		dom := dto.ToDomain()
		if len(dom.Events) != 3 {
			t.Fatalf("expected 3 events in domain model, got %d", len(dom.Events))
		}

		e0 := dom.Events[0]
		if e0.EventType != domain.EventTypeTurnMessage || e0.Turn != 1 || e0.ContextTokens != 5000 {
			t.Errorf("unexpected event 0: %+v", e0)
		}

		e1 := dom.Events[1]
		if e1.EventType != domain.EventTypeToolCall || e1.ToolName == nil || *e1.ToolName != toolName {
			t.Errorf("unexpected tool name in event 1: %+v", e1)
		}
		if e1.ErrorMessage == nil || *e1.ErrorMessage != errMsg {
			t.Errorf("unexpected error message in event 1: %+v", e1)
		}
		if e1.DurationMs == nil || *e1.DurationMs != durMs {
			t.Errorf("unexpected duration in event 1: %+v", e1)
		}

		e2 := dom.Events[2]
		if e2.EventType != domain.EventTypeCompaction || e2.Turn != 2 {
			t.Errorf("unexpected event 2: %+v", e2)
		}

		// Round-trip
		roundDTO := adapters.AgenticFactLedgerDTOFromDomain(dom)
		roundBytes, err := json.Marshal(roundDTO)
		if err != nil {
			t.Fatalf("failed to marshal round-trip DTO: %v", err)
		}

		var reloadedDTO adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(roundBytes, &reloadedDTO); err != nil {
			t.Fatalf("failed to unmarshal reloaded DTO: %v", err)
		}
		if len(reloadedDTO.Events) != 3 {
			t.Fatalf("expected 3 reloaded events, got %d", len(reloadedDTO.Events))
		}
		if *reloadedDTO.Events[1].ToolName != toolName {
			t.Errorf("expected tool_name %s, got %s", toolName, *reloadedDTO.Events[1].ToolName)
		}
	})

	t.Run("LoadExistingFactLedger loads events and handles absence", func(t *testing.T) {
		tempDir := t.TempDir()

		// 1. Workspace without events
		noEventsDir := filepath.Join(tempDir, "NO-EVENTS")
		if err := os.MkdirAll(noEventsDir, 0755); err != nil {
			t.Fatal(err)
		}
		noEventsFile := filepath.Join(noEventsDir, "agentic_metrics.json")
		if err := os.WriteFile(noEventsFile, []byte(`{"ticket_id":"NO-EVENTS","updated_at":"2026-10-18T10:00:00Z"}`), 0644); err != nil {
			t.Fatal(err)
		}

		factsNoEvents, err := adapters.LoadExistingFactLedger(noEventsFile)
		if err != nil {
			t.Fatalf("LoadExistingFactLedger failed on ledger without events: %v", err)
		}
		if len(factsNoEvents.Events) != 0 {
			t.Errorf("expected empty events slice, got %d", len(factsNoEvents.Events))
		}

		// 2. Workspace with events
		withEventsDir := filepath.Join(tempDir, "WITH-EVENTS")
		if err := os.MkdirAll(withEventsDir, 0755); err != nil {
			t.Fatal(err)
		}
		withEventsFile := filepath.Join(withEventsDir, "agentic_metrics.json")
		payload := `{
			"ticket_id":"WITH-EVENTS",
			"updated_at":"2026-10-18T10:00:00Z",
			"events":[
				{"timestamp":"2026-10-18T10:00:00Z","session_id":"s1","turn":1,"role":"coder","event_type":"tool_call","tool_name":"edit","status":"completed","context_tokens":1000}
			]
		}`
		if err := os.WriteFile(withEventsFile, []byte(payload), 0644); err != nil {
			t.Fatal(err)
		}

		factsWithEvents, err := adapters.LoadExistingFactLedger(withEventsFile)
		if err != nil {
			t.Fatalf("LoadExistingFactLedger failed on ledger with events: %v", err)
		}
		if len(factsWithEvents.Events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(factsWithEvents.Events))
		}
		if factsWithEvents.Events[0].EventType != domain.EventTypeToolCall {
			t.Errorf("expected tool_call event, got %s", factsWithEvents.Events[0].EventType)
		}
	})

	t.Run("explicit events null decodes gracefully with empty slice", func(t *testing.T) {
		nullEventsJSON := `{"schema_version":2,"ticket_id":"AL-NULL","updated_at":"2026-10-18T10:00:00Z","events":null}`
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(nullEventsJSON), &dto); err != nil {
			t.Fatalf("failed to unmarshal explicit null events: %v", err)
		}
		if dto.Events != nil {
			t.Errorf("expected dto.Events to be nil, got %+v", dto.Events)
		}
		dom := dto.ToDomain()
		if len(dom.Events) != 0 {
			t.Errorf("expected empty domain events slice, got len %d", len(dom.Events))
		}
	})

	t.Run("explicit events empty array decodes gracefully with empty slice", func(t *testing.T) {
		emptyArrayJSON := `{"schema_version":2,"ticket_id":"AL-EMPTY","updated_at":"2026-10-18T10:00:00Z","events":[]}`
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(emptyArrayJSON), &dto); err != nil {
			t.Fatalf("failed to unmarshal explicit empty events array: %v", err)
		}
		if dto.Events == nil {
			t.Fatal("expected dto.Events to be non-nil empty slice")
		}
		if len(dto.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(dto.Events))
		}
		dom := dto.ToDomain()
		if len(dom.Events) != 0 {
			t.Errorf("expected empty domain events slice, got len %d", len(dom.Events))
		}
	})

	t.Run("event with empty role and status defaults safely", func(t *testing.T) {
		eventDTO := adapters.EventFactDTO{
			Timestamp:     "2026-10-18T10:00:00Z",
			SessionID:     "ses-default",
			Turn:          1,
			Role:          "", // should default to orchestrator
			EventType:     "turn_message",
			Status:        "", // should default to completed
			ContextTokens: 100,
		}
		dom := eventDTO.ToDomain()
		if dom.Role != "orchestrator" {
			t.Errorf("expected default role orchestrator, got %q", dom.Role)
		}
		if dom.Status != "completed" {
			t.Errorf("expected default status completed, got %q", dom.Status)
		}
	})

	t.Run("corrupted event json payload returns unmarshal error", func(t *testing.T) {
		corruptedEventsJSON := `{"schema_version":2,"ticket_id":"AL-CORRUPT","events":[{"timestamp": 12345}]}` // timestamp should be string
		var dto adapters.AgenticFactLedgerDTO
		err := json.Unmarshal([]byte(corruptedEventsJSON), &dto)
		if err == nil {
			t.Fatal("expected unmarshal error for non-string timestamp in events, got nil")
		}
	})
}

func TestLoadTicketWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	ticketDir := filepath.Join(tempDir, "AL-TARGET")
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		t.Fatal(err)
	}

	factJSON := `{
		"ticket_id": "AL-TARGET",
		"ticket_title": "Target Feature",
		"updated_at": "2026-10-18T12:00:00Z",
		"code": {
			"lines_added": 40,
			"lines_deleted": 5,
			"files_changed": 3,
			"sdlc_iterations": 2
		},
		"invocations": [
			{
				"id": "ses-1",
				"role": "orchestrator",
				"model": "model-1",
				"user": "test-dev",
				"active_duration_seconds": 90,
				"turns": 3,
				"tokens": {
					"input": 2000,
					"output": 300,
					"cache_read": 1000,
					"peak_context": 25000
				},
				"cost_usd": 0.08,
				"compactions": 0
			}
		],
		"tools": [
			{
				"name": "edit",
				"calls": 4,
				"errors": 0
			}
		]
	}`

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

	t.Run("fallback to parameter ticket ID when omitted in json", func(t *testing.T) {
		noIDDir := filepath.Join(tempDir, "AL-OMITTED")
		if err := os.MkdirAll(noIDDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(noIDDir, "agentic_metrics.json"), []byte(`{"updated_at": "2026-10-18T12:00:00Z"}`), 0644); err != nil {
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
			TicketTitle:   &title,
			Invocations: []domain.InvocationFact{
				{
					ID:      "inv-1",
					Role:    "coder",
					Model:   "gemini-3.8-flash",
					CostUSD: 0.05,
				},
			},
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
		if len(parsed.Invocations) != 1 || parsed.Invocations[0].ID != "inv-1" {
			t.Errorf("expected 1 invocation with ID inv-1, got %+v", parsed.Invocations)
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
		err := adapters.SaveFactLedger(invalidPath, domain.AgenticFactLedger{})
		if err == nil {
			t.Fatal("expected error when parent directory cannot be created, got nil")
		}
		if !strings.Contains(err.Error(), "failed to create directory") {
			t.Errorf("expected 'failed to create directory' in error message, got: %v", err)
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

		err := adapters.SaveFactLedger(dirAsFile, domain.AgenticFactLedger{})
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
		firstLedger := domain.AgenticFactLedger{
			TicketTitle: &longTitle,
		}
		if err := adapters.SaveFactLedger(filePath, firstLedger); err != nil {
			t.Fatalf("first write failed: %v", err)
		}

		// Second write: tiny payload
		shortTitle := "Short"
		secondLedger := domain.AgenticFactLedger{
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
		if err := adapters.SaveFactLedger(factPath, domain.AgenticFactLedger{}); err != nil {
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
