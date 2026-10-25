package codex_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

func TestParseRolloutMapsPrivacySafeMetadataUsageAndEventKinds(t *testing.T) {
	file, err := os.Open("testdata/rollout.jsonl")
	if err != nil {
		t.Fatalf("opening rollout fixture: %v", err)
	}
	defer file.Close()

	transcript, err := codex.ParseRollout(file)
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if transcript.ThreadID != "thread-01" || transcript.SessionID != "session-01" {
		t.Fatalf("unexpected thread/session identifiers: %#v", transcript)
	}
	if transcript.WorkingDir != "/repo/example" || transcript.Provider != "openai" || transcript.Model != "gpt-example" {
		t.Fatalf("unexpected source metadata: %#v", transcript)
	}
	if transcript.Usage != (codex.Usage{Input: 120, Output: 25, CacheRead: 80, CacheWrite: 5, Reasoning: 8, PeakContext: 120}) {
		t.Fatalf("unexpected response usage: %#v", transcript.Usage)
	}
	if got := transcript.Usage.ToInvocationTokens(); got.Input != 40 || got.Output != 25 || got.CacheRead != 80 || got.Reasoning != 8 || got.PeakContext != 120 {
		t.Fatalf("unexpected shared AgenticLens token mapping: %#v", got)
	}
	if transcript.ContextWindow != 128000 {
		t.Fatalf("expected observed context window 128000, got %d", transcript.ContextWindow)
	}
	if transcript.EventKinds["exec_command_begin"] != 1 || transcript.EventKinds["mcp_tool_call_end"] != 1 {
		t.Fatalf("expected command and MCP lifecycle observations, got %#v", transcript.EventKinds)
	}
	if transcript.Turns != 1 || transcript.ToolCalls != 2 || transcript.ToolErrors != 1 || len(transcript.Invocations) != 1 {
		t.Fatalf("rollout facts were not projected: turns=%d tools=%d tool_errors=%d invocations=%d", transcript.Turns, transcript.ToolCalls, transcript.ToolErrors, len(transcript.Invocations))
	}
	if len(transcript.Events) != 2 || transcript.Events[0].Status != "completed" || transcript.Events[0].DurationMs == nil || *transcript.Events[0].DurationMs != 125 {
		t.Fatalf("expected privacy-safe paired command event with duration, got %#v", transcript.Events)
	}
	ledger := transcript.ToMetricsLedger("GEN-07", "Codex metrics research", domain.TicketTypeChore, domain.DefaultProjectConfig())
	if ledger.Facts.SchemaVersion == nil || *ledger.Facts.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatalf("expected current AgenticLens fact schema, got %#v", ledger.Facts.SchemaVersion)
	}
	if ledger.TicketID == nil || *ledger.TicketID != "GEN-07" || ledger.Totals.Invocations != 1 || ledger.Totals.InputTokens != 40 || ledger.Totals.CacheReadTokens != 80 {
		t.Fatalf("unexpected derived AgenticLens ledger: %#v", ledger)
	}
	if ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 2 || ledger.WorkflowEfficiency.ToolExecution.ErrorsAndCorrections != 1 {
		t.Fatalf("unexpected derived tool execution metrics: %#v", ledger.WorkflowEfficiency.ToolExecution)
	}
	if transcript.RecordKinds["new_future_record"] != 1 {
		t.Fatalf("unknown record type should be counted, got %#v", transcript.RecordKinds)
	}
	if len(transcript.Warnings) != 0 {
		t.Fatalf("valid fixture produced warnings: %#v", transcript.Warnings)
	}
	encoded, err := json.Marshal(transcript)
	if err != nil {
		t.Fatalf("marshalling privacy-minimized summary: %v", err)
	}
	for _, sensitiveValue := range []string{"/repo/example", "fixture-secret-command", "fixture-secret-args", "fixture-secret-prompt"} {
		if strings.Contains(string(encoded), sensitiveValue) {
			t.Fatalf("summary retained content-bearing value %q", sensitiveValue)
		}
	}
}

func TestParseRolloutUsesCumulativeTokenCountAsFallback(t *testing.T) {
	input := `{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":90,"cached_input_tokens":30,"cache_write_input_tokens":2,"output_tokens":12,"reasoning_output_tokens":4,"total_tokens":102},"last_token_usage":{"input_tokens":20},"model_context_window":64000}}}`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if transcript.Usage != (codex.Usage{Input: 90, Output: 12, CacheRead: 30, CacheWrite: 2, Reasoning: 4, PeakContext: 20}) {
		t.Fatalf("expected cumulative usage fallback, got %#v", transcript.Usage)
	}
	if mapped := transcript.Usage.ToInvocationTokens(); mapped.Input != 60 || mapped.CacheRead != 30 {
		t.Fatalf("expected cache reads to be a subset of Codex input, got %#v", mapped)
	}
}

func TestMetricsLedgerUsesCodexThreadIDForInvocationHierarchy(t *testing.T) {
	input := `{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"child-thread","session_id":"child-session","parent_thread_id":"root-thread","agent_role":"worker"}}`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	ledger := transcript.ToMetricsLedger("GEN-07", "Codex metrics research", domain.TicketTypeChore, domain.DefaultProjectConfig())
	if len(ledger.Facts.Invocations) != 1 {
		t.Fatalf("expected one invocation, got %d", len(ledger.Facts.Invocations))
	}
	invocation := ledger.Facts.Invocations[0]
	if invocation.ID != "child-thread" || invocation.ParentID == nil || *invocation.ParentID != "root-thread" {
		t.Fatalf("Codex thread hierarchy was not preserved: %#v", invocation)
	}
}

func TestParseRolloutMapsCodexCommandItemsAndPairedCustomCalls(t *testing.T) {
	input := strings.Join([]string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
		`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-1","started_at_ms":1792918800000,"completed_at_ms":1792918803000,"item":{"id":"command-1","type":"CommandExecution","status":"failed","exit_code":1,"duration":{"secs":3,"nanos":0},"command":["private-command-marker"],"aggregated_output":"private-output-marker"}}}`,
		`{"timestamp":"2026-10-25T09:00:04Z","type":"response_item","payload":{"type":"custom_tool_call","call_id":"lookup-1","name":"lookup","input":"private-input-marker"}}`,
		`{"timestamp":"2026-10-25T09:00:06Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"lookup-1","output":"private-output-marker"}}`,
	}, "\n")
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if transcript.ToolCalls != 2 || transcript.ToolErrors != 1 || len(transcript.Events) != 2 {
		t.Fatalf("expected one failed command and one paired custom call, got calls=%d errors=%d events=%#v", transcript.ToolCalls, transcript.ToolErrors, transcript.Events)
	}
	if transcript.Events[0].Status != "error" || transcript.Events[0].DurationMs == nil || *transcript.Events[0].DurationMs != 3000 {
		t.Fatalf("expected command failure and duration from Codex completed item, got %#v", transcript.Events[0])
	}
	if transcript.Events[1].Status != "completed" || transcript.Events[1].DurationMs == nil || *transcript.Events[1].DurationMs != 2000 {
		t.Fatalf("expected paired custom tool lifecycle, got %#v", transcript.Events[1])
	}
	encoded, err := json.Marshal(transcript)
	if err != nil {
		t.Fatalf("marshalling transcript summary: %v", err)
	}
	for _, marker := range []string{"private-command-marker", "private-input-marker", "private-output-marker"} {
		if strings.Contains(string(encoded), marker) {
			t.Fatalf("transcript retained private tool payload %q", marker)
		}
	}
}

func TestParseRolloutUpdatesModelWhenTurnContextFollowsTurnStart(t *testing.T) {
	input := strings.Join([]string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
		`{"timestamp":"2026-10-25T09:00:01Z","type":"turn_context","payload":{"turn_id":"turn-1","model":"gpt-real-shape"}}`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"token_usage_record","payload":{"turn_id":"turn-1","usage":{"input_tokens":10,"cached_input_tokens":4,"output_tokens":2}}}`,
		`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
	}, "\n")
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if len(transcript.Invocations) != 1 || transcript.Invocations[0].Model != "gpt-real-shape" || transcript.Invocations[0].Turns != 1 {
		t.Fatalf("turn context should replace the pre-context default model, got %#v", transcript.Invocations)
	}
}

func TestParseRolloutSkipsMalformedRecordsAndContinues(t *testing.T) {
	input := strings.Join([]string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"turn_started"}}`,
		`{"type":"event_msg","payload":`,
		`{"timestamp":"2026-10-25T09:00:02Z","type":"event_msg","payload":{"type":"turn_complete"}}`,
	}, "\n")
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if transcript.EventKinds["turn_started"] != 1 || transcript.EventKinds["turn_complete"] != 1 {
		t.Fatalf("parser did not continue after malformed record: %#v", transcript.EventKinds)
	}
	if len(transcript.Warnings) != 1 || !strings.Contains(transcript.Warnings[0], "line 2") {
		t.Fatalf("expected one malformed-record warning, got %#v", transcript.Warnings)
	}
}

func TestParseRolloutToleratesTruncatedTrailingRecord(t *testing.T) {
	input := "{\"timestamp\":\"2026-10-25T09:00:00Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"turn_started\"}}\n" +
		`{"timestamp":"2026-10-25T09:00:01Z","type":"event_msg","payload":{"type":"turn_complete"}`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseRollout returned an error: %v", err)
	}
	if transcript.EventKinds["turn_started"] != 1 || transcript.EventKinds["turn_complete"] != 0 {
		t.Fatalf("truncated tail should be ignored, got events %#v", transcript.EventKinds)
	}
	if len(transcript.Warnings) != 1 {
		t.Fatalf("expected a warning for the truncated tail, got %#v", transcript.Warnings)
	}
}

func TestParseRolloutRejectsNilReader(t *testing.T) {
	if _, err := codex.ParseRollout(nil); err == nil {
		t.Fatal("expected an error for nil reader")
	}
}

func TestCombineTranscriptsPreservesThreadFactsAndAggregatesTools(t *testing.T) {
	parentID := "root-thread"
	transcripts := []codex.Transcript{
		{
			ThreadID: "root-thread", StartedAt: "2026-10-25T09:00:00Z", UpdatedAt: "2026-10-25T09:00:10Z",
			Invocations: []domain.InvocationFact{{ID: "root-thread"}},
			Tools:       []domain.ToolFact{{Name: "command", Calls: 2, Errors: 1}},
			Events:      []domain.EventFact{{Timestamp: "2026-10-25T09:00:02Z", SessionID: "root-thread"}},
			Usage:       codex.Usage{Input: 10, CacheRead: 4},
		},
		{
			ThreadID: "child-thread", StartedAt: "2026-10-25T09:00:03Z", UpdatedAt: "2026-10-25T09:00:08Z",
			Invocations: []domain.InvocationFact{{ID: "child-thread", ParentID: &parentID}},
			Tools:       []domain.ToolFact{{Name: "command", Calls: 1}, {Name: "search", Calls: 1}},
			Events:      []domain.EventFact{{Timestamp: "2026-10-25T09:00:01Z", SessionID: "child-thread"}},
			Usage:       codex.Usage{Input: 20, CacheRead: 7},
		},
	}

	combined, err := codex.CombineTranscripts(transcripts)
	if err != nil {
		t.Fatalf("CombineTranscripts returned an error: %v", err)
	}
	if combined.ThreadID != "root-thread" || combined.Usage.Input != 30 || combined.Usage.CacheRead != 11 {
		t.Fatalf("unexpected combined thread metadata/usage: %#v", combined)
	}
	if len(combined.Invocations) != 2 || len(combined.Tools) != 2 || len(combined.Events) != 2 {
		t.Fatalf("expected facts from both threads, got invocations=%d tools=%d events=%d", len(combined.Invocations), len(combined.Tools), len(combined.Events))
	}
	if combined.Tools[0].Name != "command" || combined.Tools[0].Calls != 3 || combined.Tools[0].Errors != 1 {
		t.Fatalf("expected aggregate command totals, got %#v", combined.Tools)
	}
	if combined.Events[0].Timestamp != "2026-10-25T09:00:01Z" {
		t.Fatalf("expected chronologically ordered events, got %#v", combined.Events)
	}
}
