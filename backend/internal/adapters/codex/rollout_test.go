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
	ledger := transcript.ToMetricsLedger("TEST-101", "Fixture metrics", domain.TicketTypeChore, domain.DefaultProjectConfig())
	if ledger.Facts.SchemaVersion == nil || *ledger.Facts.SchemaVersion != domain.CurrentSchemaVersion {
		t.Fatalf("expected current AgenticLens fact schema, got %#v", ledger.Facts.SchemaVersion)
	}
	if ledger.TicketID == nil || *ledger.TicketID != "TEST-101" || ledger.Totals.Invocations != 1 || ledger.Totals.InputTokens != 40 || ledger.Totals.CacheReadTokens != 80 {
		t.Fatalf("unexpected derived AgenticLens ledger: %#v", ledger)
	}
	if ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 2 || ledger.WorkflowEfficiency.ToolExecution.ErrorsAndCorrections != 1 {
		t.Fatalf("unexpected derived tool execution metrics: %#v", ledger.WorkflowEfficiency.ToolExecution)
	}
	if transcript.RecordKinds["world_state"] != 1 || transcript.RecordKinds["inter_agent_communication_metadata"] != 1 {
		t.Fatalf("known non-metric records should be explicitly accounted, got %#v", transcript.RecordKinds)
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
	encodedLedger, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"/repo/example", "fixture-secret-command", "fixture-secret-args", "fixture-secret-prompt"} {
		if strings.Contains(string(encodedLedger), marker) {
			t.Fatalf("exported ledger retained content-bearing value %q", marker)
		}
	}
}

func TestParseRolloutRejectsUnknownRecordType(t *testing.T) {
	input := `{"timestamp":"2026-10-25T09:00:00Z","type":"future_record_v99","payload":{}}`
	_, err := codex.ParseRollout(strings.NewReader(input))
	if err == nil || !strings.Contains(err.Error(), `unsupported Codex rollout record type "future_record_v99"`) {
		t.Fatalf("expected unknown record type to fail extraction, got %v", err)
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
	ledger := transcript.ToMetricsLedger("TEST-101", "Fixture metrics", domain.TicketTypeChore, domain.DefaultProjectConfig())
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

func TestSteeringTurnsCountUserMessagesWithoutDuplicateRepresentations(t *testing.T) {
	responses := `{"type":"response_item","payload":{"type":"message","role":"user"}}
{"type":"response_item","payload":{"type":"message","role":"assistant"}}
{"type":"response_item","payload":{"type":"message","role":"user"}}
`
	for _, input := range []string{responses, responses + `{"type":"event_msg","payload":{"type":"user_message"}}
{"type":"event_msg","payload":{"type":"user_message"}}
`} {
		transcript, err := codex.ParseRollout(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if transcript.SteeringTurns != 2 {
			t.Fatalf("expected two user inputs, got %d", transcript.SteeringTurns)
		}
		combined, err := codex.CombineTranscripts([]codex.Transcript{transcript, {SteeringTurns: 5}})
		if err != nil {
			t.Fatal(err)
		}
		if combined.SteeringTurns != 2 {
			t.Fatalf("child delegation inflated steering: %d", combined.SteeringTurns)
		}
	}
}

func TestChildRolloutPreservesIdentityAndSkipsInheritedReplay(t *testing.T) {
	inherited := `{"timestamp":"2026-01-01T10:00:00Z","type":"session_meta","payload":{"id":"child","session_id":"root","timestamp":"2026-01-01T10:00:00Z","cli_version":"child-version","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root"}}}}}
{"timestamp":"2026-01-01T10:00:00Z","type":"session_meta","payload":{"id":"root","session_id":"root","timestamp":"2025-01-01T10:00:00Z","cli_version":"parent-version","source":"cli"}}
{"timestamp":"2026-01-01T10:00:00Z","type":"turn_context","payload":{"model":"parent-model","turn_id":"parent-turn"}}
{"timestamp":"2026-01-01T10:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"parent-turn"}}
{"timestamp":"2026-01-01T10:00:00Z","type":"response_item","payload":{"type":"message","role":"user"}}
{"timestamp":"2026-01-01T10:00:00Z","type":"event_msg","payload":{"type":"exec_command_end","call_id":"parent-call","exit_code":1}}
{"timestamp":"2026-01-01T10:00:00Z","type":"compacted","payload":{}}
{"timestamp":"2026-01-01T10:00:00Z","type":"token_usage_record","payload":{"thread_id":"root","model":"parent-model","usage":{"input_tokens":9000,"output_tokens":900}}}
{"timestamp":"2026-01-01T10:00:00Z","type":"token_usage_record","payload":{"thread_id":"child","model":"inherited-model","usage":{"input_tokens":8000}}}
`
	own := `{"timestamp":"2026-01-01T10:00:01Z","type":"event_msg","payload":{"type":"thread_settings_applied","thread_id":"child"}}
{"timestamp":"2026-01-01T10:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T10:00:02Z","type":"turn_context","payload":{"model":"child-model","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T10:00:03Z","type":"event_msg","payload":{"type":"exec_command_end","call_id":"child-call","turn_id":"child-turn","exit_code":1}}
{"timestamp":"2026-01-01T10:00:04Z","type":"token_usage_record","payload":{"thread_id":"child","model":"child-model","turn_id":"child-turn","usage":{"input_tokens":30,"cached_input_tokens":10,"output_tokens":2}}}
{"timestamp":"2026-01-01T10:00:04Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"child-turn"}}
`
	t.Run("own activity only", func(t *testing.T) {
		transcript, err := codex.ParseRollout(strings.NewReader(inherited + own))
		if err != nil {
			t.Fatal(err)
		}
		if transcript.ThreadID != "child" || transcript.Source != "subagent" || transcript.CodexVersion != "child-version" || transcript.StartedAt != "2026-01-01T10:00:00Z" {
			t.Fatalf("child identity was overwritten: %+v", transcript)
		}
		if len(transcript.Invocations) != 1 {
			t.Fatalf("inherited model counted: %+v", transcript.Invocations)
		}
		inv := transcript.Invocations[0]
		if inv.ID != "child" || inv.ParentID == nil || *inv.ParentID != "root" || inv.Role != domain.RoleUnassignedSubAgent || inv.Model != "child-model" {
			t.Fatalf("incorrect child invocation attribution: %+v", inv)
		}
		if inv.Turns != 1 || inv.ActiveDurationSeconds != 2 || inv.Compactions != 0 || inv.Tokens.Input != 20 || inv.Tokens.CacheRead != 10 || inv.Tokens.Output != 2 {
			t.Fatalf("inherited facts counted: %+v", inv)
		}
		if transcript.SteeringTurns != 0 || transcript.ToolCalls != 1 || len(transcript.Events) != 1 || transcript.Events[0].SessionID != "child" {
			t.Fatalf("inherited tool/event/steering facts counted: %+v", transcript)
		}
	})
	t.Run("missing resume marker fails extraction", func(t *testing.T) {
		_, err := codex.ParseRollout(strings.NewReader(inherited))
		if err == nil || !strings.Contains(err.Error(), "without an owning-thread resume marker") {
			t.Fatalf("expected incomplete child rollout to fail extraction, got %v", err)
		}
	})
}

func TestCodexExtractionPreservesRawPersonaAndExplicitRole(t *testing.T) {
	input := `{"type":"session_meta","payload":{"id":"child","parent_thread_id":"root","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root","agent_path":"/private/agents/coder"}}}}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture-model"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"usage":{"input_tokens":1}}}
`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if transcript.Invocations[0].Role != domain.AgentRole("coder") {
		t.Fatalf("expected raw persona in extracted facts, got %s", transcript.Invocations[0].Role)
	}
	encoded, err := json.Marshal(transcript)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "/private/agents/coder") {
		t.Fatal("private agent path was exposed")
	}

	explicitInput := strings.Replace(input, `"agent_path":"/private/agents/coder"`, `"agent_path":"/private/agents/coder","agent_role":"reviewer"`, 1)
	explicit, err := codex.ParseRollout(strings.NewReader(explicitInput))
	if err != nil {
		t.Fatal(err)
	}
	if explicit.Invocations[0].Role != domain.AgentRole("reviewer") {
		t.Fatalf("explicit role lost precedence: %s", explicit.Invocations[0].Role)
	}
}

func TestCodexUserIterationsAttribution(t *testing.T) {
	rootJSON := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"root_thread"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"event_msg","payload":{"type":"user_message"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"event_msg","payload":{"type":"user_message"}}
{"timestamp":"2026-01-01T00:00:03Z","type":"token_usage_record","payload":{"model":"model-a","usage":{"input_tokens":10}}}
{"timestamp":"2026-01-01T00:00:04Z","type":"token_usage_record","payload":{"model":"model-b","usage":{"input_tokens":20}}}
`
	childJSON := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"child_thread","parent_thread_id":"root_thread"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"event_msg","payload":{"type":"user_message"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"model":"model-a","usage":{"input_tokens":15}}}
`
	rootTranscript, err := codex.ParseRollout(strings.NewReader(rootJSON))
	if err != nil {
		t.Fatalf("failed parsing root rollout: %v", err)
	}
	childTranscript, err := codex.ParseRollout(strings.NewReader(childJSON))
	if err != nil {
		t.Fatalf("failed parsing child rollout: %v", err)
	}

	if len(rootTranscript.Invocations) != 2 {
		t.Fatalf("expected 2 invocations for root, got %d", len(rootTranscript.Invocations))
	}
	// Primary model on root thread receives SteeringTurns (2)
	if rootTranscript.Invocations[0].UserIterations != 2 {
		t.Errorf("expected root primary invocation UserIterations=2, got %d", rootTranscript.Invocations[0].UserIterations)
	}
	// Secondary model on root thread receives 0
	if rootTranscript.Invocations[1].UserIterations != 0 {
		t.Errorf("expected root secondary invocation UserIterations=0, got %d", rootTranscript.Invocations[1].UserIterations)
	}

	// Child rollout invocation receives 0
	if len(childTranscript.Invocations) != 1 {
		t.Fatalf("expected 1 invocation for child, got %d", len(childTranscript.Invocations))
	}
	if childTranscript.Invocations[0].UserIterations != 0 {
		t.Errorf("expected child invocation UserIterations=0, got %d", childTranscript.Invocations[0].UserIterations)
	}
}
