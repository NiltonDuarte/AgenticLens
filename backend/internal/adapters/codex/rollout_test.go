package codex_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
	if inv := transcript.Invocations[0]; inv.CostEvidence.State != domain.CostUnavailable || inv.CostEvidence.KnownSubtotalUSD != nil {
		t.Fatalf("Codex usage implied observed zero cost: %+v", inv)
	}
	if len(transcript.Tools) == 0 {
		t.Fatal("expected session-scoped tool facts")
	}
	for _, tool := range transcript.Tools {
		if tool.SessionID != transcript.ThreadID {
			t.Errorf("expected tool session_id %q, got %q", transcript.ThreadID, tool.SessionID)
		}
	}
	if len(transcript.Events) != 2 || transcript.Events[0].Status != "completed" || transcript.Events[0].DurationMs == nil || *transcript.Events[0].DurationMs != 125 {
		t.Fatalf("expected privacy-safe paired command event with duration, got %#v", transcript.Events)
	}
	sessions, err := transcript.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	version := domain.CurrentSchemaVersion
	ticketID := "TEST-101"
	factsForMetrics := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticketID, Sessions: sessions}
	ledger := domain.DeriveMetricsLedger(factsForMetrics, codexMetricConfig())
	ledger.Facts, err = domain.PersistAnomalySummaries(factsForMetrics)
	if err != nil {
		t.Fatal(err)
	}
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
	sessions, err := transcript.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	version := domain.CurrentSchemaVersion
	ticketID := "TEST-101"
	factsForMetrics := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticketID, Sessions: sessions}
	ledger := domain.DeriveMetricsLedger(factsForMetrics, codexMetricConfig())
	ledger.Facts, err = domain.PersistAnomalySummaries(factsForMetrics)
	if err != nil {
		t.Fatal(err)
	}
	if len(ledger.Facts.Sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(ledger.Facts.Sessions))
	}
	session := ledger.Facts.Sessions[0]
	if session.ID != "child-thread" || session.ParentID == nil || *session.ParentID != "root-thread" {
		t.Fatalf("Codex thread hierarchy was not preserved: %#v", session)
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
	for _, model := range []string{"gpt-real-shape", "default", "unknown"} {
		t.Run(model, func(t *testing.T) {
			input := strings.Join([]string{
				`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
				fmt.Sprintf(`{"timestamp":"2026-10-25T09:00:01Z","type":"turn_context","payload":{"turn_id":"turn-1","model":%q}}`, model),
				`{"timestamp":"2026-10-25T09:00:02Z","type":"token_usage_record","payload":{"turn_id":"turn-1","usage":{"input_tokens":10,"cached_input_tokens":4,"output_tokens":2}}}`,
				`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
			}, "\n")
			transcript, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(transcript.Invocations) != 1 {
				t.Fatalf("empty provisional row retained: %+v", transcript.Invocations)
			}
			inv := transcript.Invocations[0]
			if inv.Model != model || inv.Turns != 1 || inv.Tokens.Input != 6 || inv.Tokens.CacheRead != 4 || inv.Tokens.Output != 2 || inv.ActiveDurationSeconds != 3 {
				t.Fatalf("late model attribution changed counters: %+v", inv)
			}
		})
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

func TestCombineTranscriptsPreservesPerSessionToolFacts(t *testing.T) {
	parentID := "root-thread"
	transcripts := []codex.Transcript{
		{
			ThreadID: "root-thread", StartedAt: "2026-10-25T09:00:00Z", UpdatedAt: "2026-10-25T09:00:10Z",
			Invocations: []domain.InvocationFact{{}},
			Tools:       []domain.ToolFact{{SessionID: "root-thread", Name: "command", Calls: 2, Errors: 1}},
			Events:      []domain.EventFact{{Timestamp: "2026-10-25T09:00:02Z", SessionID: "root-thread"}},
			Usage:       codex.Usage{Input: 10, CacheRead: 4},
		},
		{
			ThreadID: "child-thread", ParentThread: parentID, StartedAt: "2026-10-25T09:00:03Z", UpdatedAt: "2026-10-25T09:00:08Z",
			Invocations: []domain.InvocationFact{{}},
			Tools:       []domain.ToolFact{{SessionID: "child-thread", Name: "command", Calls: 1}, {SessionID: "child-thread", Name: "search", Calls: 1}},
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
	if len(combined.Invocations) != 2 || len(combined.Tools) != 3 || len(combined.Events) != 2 {
		t.Fatalf("expected facts from both threads, got invocations=%d tools=%d events=%d", len(combined.Invocations), len(combined.Tools), len(combined.Events))
	}
	if combined.Tools[0].SessionID != "child-thread" || combined.Tools[0].Name != "command" || combined.Tools[0].Calls != 1 {
		t.Fatalf("expected child command facts to retain session identity, got %#v", combined.Tools)
	}
	if combined.Tools[2].SessionID != "root-thread" || combined.Tools[2].Calls != 2 || combined.Tools[2].Errors != 1 {
		t.Fatalf("expected root command facts to retain session identity, got %#v", combined.Tools)
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
		transcript.ThreadID = "root-thread"
		combined, err := codex.CombineTranscripts([]codex.Transcript{transcript, {ThreadID: "child-thread", ParentThread: "root-thread", SteeringTurns: 5}})
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
		if inv.Role != "" || inv.Model != "child-model" {
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

	for _, tc := range []struct{ name, metadata, want, parent string }{
		{"root path", `"agent_path":"/private/agents/planner"`, "planner", ""},
		{"root explicit overrides path", `"agent_path":"/private/agents/planner","agent_role":" reviewer "`, "reviewer", ""},
		{"root missing", `"agent_role":"   "`, "", ""},
		{"child missing", `"parent_thread_id":"parent","agent_role":"  "`, "", "parent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"type":"session_meta","payload":{"id":"role-session",` + tc.metadata + `}}` + "\n" +
				`{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture-model"}}` + "\n" +
				`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"usage":{"input_tokens":7}}}` + "\n" +
				`{"timestamp":"2026-01-01T00:00:03Z","type":"event_msg","payload":{"type":"exec_command_end","call_id":"call-role","exit_code":0}}` + "\n"
			transcript, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(transcript.Invocations) != 1 || string(transcript.Invocations[0].Role) != tc.want || transcript.Invocations[0].Tokens.Input != 7 {
				t.Fatalf("wrong persona or tokens: %+v", transcript.Invocations)
			}
			if len(transcript.Events) != 1 || string(transcript.Events[0].Role) != tc.want {
				t.Fatalf("event persona not propagated: %+v", transcript.Events)
			}
			if transcript.ParentThread != tc.parent {
				t.Fatalf("parent changed: %q", transcript.ParentThread)
			}
		})
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

func TestCodexSuccessfulCommandOutputPreservesExecution(t *testing.T) {
	input := strings.Join([]string{
		`{"timestamp":"2026-10-25T09:00:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
		`{"timestamp":"2026-10-25T09:00:03Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-1","item":{"id":"command-1","type":"CommandExecution","status":"completed","exit_code":0,"command":["check"],"aggregated_output":"syntax error: bad name"}}}`,
	}, "\n")
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if transcript.ToolCalls != 1 || transcript.ToolErrors != 0 || len(transcript.Events) != 1 {
		t.Fatalf("execution changed: %+v", transcript)
	}
	event := transcript.Events[0]
	if event.Status != "completed" || event.ToolOutput == nil || *event.ToolOutput != "syntax error: bad name" {
		t.Fatalf("diagnostic lost: %+v", event)
	}
	if len(transcript.Tools) != 1 || transcript.Tools[0].Calls != 1 || transcript.Tools[0].Errors != 0 {
		t.Fatalf("tool count lost: %+v", transcript.Tools)
	}
}

func TestCodexOutputEnvelopesPreserveRawEvidence(t *testing.T) {
	// Synthetic examples of supported result shapes; local runtime evidence is
	// recorded separately by the coder and is not used as a fixture here.
	const diagnostic = "syntax error: unexpected token"
	encode := func(value string) string {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	nested := `{"output":` + encode("Process exited with code 0\nFinal output:\n"+diagnostic) + `}`
	for _, tc := range []struct {
		name, raw, wantOutput string
	}{
		{"plain", encode(diagnostic), diagnostic},
		{"MCP object", `{ "content": [{"type":"text","text":"syntax error: unexpected token"}] }`, `{ "content": [{"type":"text","text":"syntax error: unexpected token"}] }`},
		{"ordered text blocks", `{"content":[{"type":"text","text":"build starting"},{"type":"text","text":"syntax error: unexpected token"}]}`, `{"content":[{"type":"text","text":"build starting"},{"type":"text","text":"syntax error: unexpected token"}]}`},
		{"input text array", `[{"type":"input_text","text":"syntax error: unexpected token"}]`, `[{"type":"input_text","text":"syntax error: unexpected token"}]`},
		{"nested command wrapper", nested, nested},
		{"serialized command wrapper in text", `{"content":[{"type":"text","text":` + encode(nested) + `}]}`, `{"content":[{"type":"text","text":` + encode(nested) + `}]}`},
		{"clean text", `{"content":[{"type":"text","text":"all checks passed"}]}`, `{"content":[{"type":"text","text":"all checks passed"}]}`},
		{"mixed clean and nontext", `{"content":[{"type":"text","text":"all checks passed"},{"type":"image","text":"syntax error: unexpected token"}]}`, `{"content":[{"type":"text","text":"all checks passed"},{"type":"image","text":"syntax error: unexpected token"}]}`},
		{"nontext", `{"content":[{"type":"image","text":"syntax error: unexpected token"}]}`, `{"content":[{"type":"image","text":"syntax error: unexpected token"}]}`},
		{"unknown envelope", `{"diagnostics":"syntax error: unexpected token"}`, `{"diagnostics":"syntax error: unexpected token"}`},
		{"unknown block", `{"content":[{"type":"unknown","text":"syntax error: unexpected token"}]}`, `{"content":[{"type":"unknown","text":"syntax error: unexpected token"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"timestamp":"2026-10-31T10:00:00Z","type":"response_item","payload":{"type":"function_call","call_id":"content-call","name":"run_check","arguments":"{}"}}` + "\n" + `{"timestamp":"2026-10-31T10:00:01Z","type":"response_item","payload":{"type":"function_call_output","call_id":"content-call","output":` + tc.raw + `}}`
			transcript, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if transcript.ToolCalls != 1 || transcript.ToolErrors != 0 || len(transcript.Events) != 1 || len(transcript.Tools) != 1 {
				t.Fatalf("invocation changed: %+v", transcript)
			}
			event := transcript.Events[0]
			tool := transcript.Tools[0]
			if event.EventID != "codex-call:content-call" || event.Status != "completed" || tool.Errors != 0 || tool.Calls != 1 {
				t.Fatalf("execution/counts changed: event=%+v tool=%+v", event, tool)
			}
			if event.ToolOutput == nil || *event.ToolOutput != tc.wantOutput {
				t.Fatalf("raw evidence changed: got=%v want=%q", event.ToolOutput, tc.wantOutput)
			}

		})
	}
}

func TestCodexMCPOutputEnvelopeRespectsExecutionStatus(t *testing.T) {
	const output = `{"content":[{"type":"text","text":"syntax error: unexpected token"}]}`
	for _, status := range []string{"completed", "failed"} {
		t.Run(status, func(t *testing.T) {
			input := `{"timestamp":"2026-10-31T10:00:00Z","type":"event_msg","payload":{"type":"mcp_tool_call_begin","call_id":"mcp-content","invocation":{"server":"compiler","tool":"check"}}}` + "\n" + `{"timestamp":"2026-10-31T10:00:01Z","type":"event_msg","payload":{"type":"mcp_tool_call_end","call_id":"mcp-content","invocation":{"server":"compiler","tool":"check"},"status":"` + status + `","output":` + output + `}}`
			transcript, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(transcript.Events) != 1 || len(transcript.Tools) != 1 || transcript.ToolCalls != 1 {
				t.Fatalf("MCP identity changed: %+v", transcript)
			}
			event, tool := transcript.Events[0], transcript.Tools[0]
			wantErrors := int64(0)
			wantStatus := "completed"
			if status == "failed" {
				wantErrors = 1
				wantStatus = "error"
			}
			if event.EventID != "codex-call:mcp-content" || event.Status != wantStatus || transcript.ToolErrors != wantErrors || tool.Errors != wantErrors {
				t.Fatalf("MCP execution status changed: event=%+v tool=%+v", event, tool)
			}
			if event.ToolOutput == nil || *event.ToolOutput != output {
				t.Fatal("MCP raw envelope changed")
			}
		})
	}
}

func TestCodexOwnedOccurrencesUseNativeCallsAndActivationEvidence(t *testing.T) {
	// Synthetic records matching supported native shapes; no live rollout reads.
	input := `{"timestamp":"2026-01-01T10:00:00Z","type":"session_meta","payload":{"id":"owner","source":"cli"}}
{"timestamp":"2026-01-01T10:00:01Z","type":"response_item","payload":{"type":"message","id":"selection","role":"user","content":[{"type":"input_text","text":"$inspect-source investigate"}]}}
{"timestamp":"2026-01-01T10:00:02Z","type":"response_item","payload":{"type":"message","id":"attachment","role":"user","content":[{"type":"input_text","text":"<skill>\n<name>inspect-source</name>\n<path>/skills/inspect-source/SKILL.md</path>\n</skill>"}]}}
{"timestamp":"2026-01-01T10:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-owner"}}
{"timestamp":"2026-01-01T10:00:02Z","type":"turn_context","payload":{"model":"model","turn_id":"turn-owner"}}
{"timestamp":"2026-01-01T10:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn-a","name":"collaboration.spawn_agent","arguments":"{\"agent_type\":\"tester\",\"task_name\":\"same-task\",\"message\":\"private prompt\"}"}}
{"timestamp":"2026-01-01T10:00:04Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn-b","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\",\"task_name\":\"same-task\",\"message\":\"private prompt\"}"}}
{"timestamp":"2026-01-01T10:00:05Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn-b","output":"{\"session_id\":\"unavailable-b\"}"}}
{"timestamp":"2026-01-01T10:00:06Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn-a","output":"{\"session_id\":\"unavailable-a\"}"}}
{"timestamp":"2026-01-01T10:00:07Z","type":"response_item","payload":{"type":"function_call","call_id":"load","name":"skill","arguments":"{\"name\":\"inspect-source\"}"}}
{"timestamp":"2026-01-01T10:00:08Z","type":"response_item","payload":{"type":"function_call_output","call_id":"load","output":"{\"status\":\"completed\"}"}}
{"timestamp":"2026-01-01T10:00:09Z","type":"response_item","payload":{"type":"function_call","call_id":"pending","name":"spawn_agent","arguments":"{\"agent_type\":\"reviewer\"}"}}
{"timestamp":"2026-01-01T10:00:10Z","type":"response_item","payload":{"type":"message","id":"catalogue","role":"user","content":[{"type":"input_text","text":"Available skills: $catalogue-only at /skills/catalogue-only/SKILL.md"}]}}
{"timestamp":"2026-01-01T10:00:11Z","type":"response_item","payload":{"type":"message","id":"quote","role":"user","content":[{"type":"input_text","text":"> $quoted is an example"}]}}
{"timestamp":"2026-01-01T10:00:12Z","type":"response_item","payload":{"type":"function_call","call_id":"plain-read","name":"exec_command","arguments":"{\"cmd\":\"cat /skills/catalogue-only/SKILL.md\"}"}}
`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := transcript.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	ledger := domain.AgenticFactLedger{Sessions: sessions}
	skills, tasks := ledger.SkillFacts(), ledger.DelegationFacts()
	if len(skills) != 2 || len(tasks) != 3 {
		t.Fatalf("missing producer or invented activity: skills=%+v tasks=%+v", skills, tasks)
	}
	if len(sessions) != 1 || sessions[0].ID != "owner" || skills[0].CallerRole == nil || *skills[0].CallerRole != "user" || skills[0].InvokedAt == nil || *skills[0].InvokedAt != "2026-01-01T10:00:01Z" {
		t.Fatalf("request became successful execution or lost owner: %+v", skills[0])
	}
	if skills[1].Name != "inspect-source" || skills[1].InvokedAt == nil || *skills[1].InvokedAt != "2026-01-01T10:00:07Z" || transcript.Skills[1].Provenance.CallID != "load" {
		t.Fatalf("native load outcome lost: %+v", skills[1])
	}
	for i, want := range []struct{ call, child string }{{"spawn-a", "unavailable-a"}, {"spawn-b", "unavailable-b"}, {"pending", ""}} {
		if transcript.Delegations[i].Provenance.CallID != want.call || tasks[i].ChildSession != want.child {
			t.Fatalf("output joined by order or unavailable child lost: %+v", tasks[i])
		}
	}
	if tasks[2].AgentType == nil || *tasks[2].AgentType != "reviewer" {
		t.Fatalf("unreturned delegation got invented outcome: %+v", tasks[2])
	}
	encoded, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "private prompt") {
		t.Fatal("delegation prompt leaked into exported facts")
	}
}

func TestCodexInheritedOccurrencesDoNotBecomeChildFacts(t *testing.T) {
	input := `{"timestamp":"2026-01-01T10:01:00Z","type":"session_meta","payload":{"id":"child","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root"}}}}}
{"timestamp":"2026-01-01T10:01:00Z","type":"session_meta","payload":{"id":"root","source":"cli"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"$inherited-skill"}]}}
{"timestamp":"2026-01-01T10:01:00Z","type":"response_item","payload":{"type":"function_call","call_id":"parent-spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"coder\"}"}}
{"timestamp":"2026-01-01T10:01:01Z","type":"event_msg","payload":{"type":"thread_settings_applied","thread_id":"child"}}
{"timestamp":"2026-01-01T10:01:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T10:01:03Z","type":"response_item","payload":{"type":"function_call","call_id":"child-spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}
`
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := transcript.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	ledger := domain.AgenticFactLedger{Sessions: sessions}
	if len(ledger.SkillFacts()) != 0 || len(ledger.DelegationFacts()) != 1 || sessions[0].ID != "child" || transcript.Delegations[0].Provenance.CallID != "child-spawn" {
		t.Fatalf("parent occurrence replay leaked into child: %+v %+v", ledger.SkillFacts(), ledger.DelegationFacts())
	}
}

func TestCodexDistinctSkillAttachmentsPreserveNativeMultiplicityAndTimeline(t *testing.T) {
	attachment := `<skill><name>inspect-source</name><path>/skills/inspect-source/SKILL.md</path></skill>`
	for _, separated := range []bool{false, true} {
		t.Run(fmt.Sprintf("intervening-user=%t", separated), func(t *testing.T) {
			input := `{"type":"session_meta","payload":{"id":"owner","source":"cli"}}` + "\n"
			message := func(id, text string) string {
				data, err := json.Marshal(text)
				if err != nil {
					t.Fatal(err)
				}
				return `{"timestamp":"2026-01-01T10:00:01Z","type":"response_item","payload":{"type":"message","id":"` + id + `","role":"user","content":[{"type":"input_text","text":` + string(data) + `}]}}` + "\n"
			}
			input += message("request-one", attachment)
			if separated {
				input += message("unrelated", "A different user instruction")
			}
			input += message("request-two", attachment)
			got, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Skills) != 2 || got.Skills[0].Provenance.MessageID != "request-one" || got.Skills[1].Provenance.MessageID != "request-two" {
				t.Fatalf("BEH-01 distinct native requests collapsed: %+v", got.Skills)
			}
			sessions, err := got.SessionFacts()
			if err != nil {
				t.Fatal(err)
			}
			if len(sessions) != 1 || len(sessions[0].Skills) != 2 || sessions[0].Skills[0].InvokedAt == nil || sessions[0].Skills[1].InvokedAt == nil || *sessions[0].Skills[0].InvokedAt != *sessions[0].Skills[1].InvokedAt {
				t.Fatalf("equal-looking lean requests collapsed: %+v", sessions)
			}
			if len(got.Events) != 2 {
				t.Fatalf("BEH-06 selected requests lost timeline occurrences: %+v", got.Events)
			}
			for i, id := range []string{"request-one", "request-two"} {
				event := got.Events[i]
				if event.EventType != domain.EventType("skill_activation") || event.SessionID != "owner" || event.MessageID == nil || *event.MessageID != id || event.Status != "unknown" {
					t.Fatalf("activation event does not preserve producer-native identity/outcome: %+v", event)
				}
			}
		})
	}
}

func TestCodexLeanOccurrencesRetainAttemptsAcrossOutputStatuses(t *testing.T) {
	for _, tc := range []struct {
		name, output string
	}{
		{"opaque", "opaque successful-looking text"},
		{"empty", ""},
		{"unknown shape", `{"future_result":"done"}`},
		{"reference only", `{"task_name":"/root/tester"}`},
		{"positive status", `{"status":"completed"}`},
		{"positive success", `{"success":true}`},
		{"recorded error", `{"error":"load failed"}`},
		{"negative success", `{"success":false}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.output)
			if err != nil {
				t.Fatal(err)
			}
			input := `{"type":"session_meta","payload":{"id":"owner"}}` + "\n" +
				`{"type":"response_item","payload":{"type":"function_call","name":"load_skill","call_id":"skill-call","arguments":"{\"name\":\"inspect-source\"}"}}` + "\n" +
				`{"type":"response_item","payload":{"type":"function_call_output","call_id":"skill-call","output":` + string(encoded) + `}}` + "\n" +
				`{"type":"response_item","payload":{"type":"function_call","name":"spawn_agent","call_id":"task-call","arguments":"{\"agent_type\":\"tester\"}"}}` + "\n" +
				`{"type":"response_item","payload":{"type":"function_call_output","call_id":"task-call","output":` + string(encoded) + `}}` + "\n"
			got, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Skills) != 1 || len(got.Delegations) != 1 {
				t.Fatalf("DATA-04 opaque output promoted or explicit evidence lost: skills=%+v tasks=%+v", got.Skills, got.Delegations)
			}
			sessions, err := got.SessionFacts()
			if err != nil {
				t.Fatal(err)
			}
			leanEncoded, err := json.Marshal(domain.AgenticFactLedger{Sessions: sessions})
			if err != nil {
				t.Fatal(err)
			}
			for _, retired := range []string{`"success"`, `"outcome"`, `"error"`} {
				if strings.Contains(string(leanEncoded), retired) {
					t.Fatalf("source output status leaked into lean occurrence: %s", leanEncoded)
				}
			}
			if len(sessions) != 1 || len(sessions[0].Skills) != 1 || len(sessions[0].Delegations) != 1 || sessions[0].Skills[0].Name != "inspect-source" || sessions[0].Delegations[0].AgentType == nil || *sessions[0].Delegations[0].AgentType != "tester" {
				t.Fatalf("output status discarded original attempts: %+v", sessions)
			}

		})
	}
}

func TestCodexDelegationKeepsOpaqueAgentSeparateFromSessionIdentity(t *testing.T) {
	for _, tc := range []struct{ name, output, wantSession, wantAgent string }{
		{"explicit session", `{"agent_id":"opaque-agent","session_id":"native-session","status":"completed"}`, "native-session", "opaque-agent"},
		{"opaque only", `{"agent_id":"opaque-agent","status":"completed"}`, "", "opaque-agent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			output, err := json.Marshal(tc.output)
			if err != nil {
				t.Fatal(err)
			}
			input := `{"type":"session_meta","payload":{"id":"owner"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":` + string(output) + `}}`
			got, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Delegations) != 1 || got.Delegations[0].ChildSession != tc.wantSession || got.Delegations[0].ChildAgentID != tc.wantAgent {
				t.Fatalf("BEH-07 opaque agent became a session or native reference lost: %+v", got.Delegations)
			}
		})
	}
}

func TestCodexMalformedRecognizedSkillOccurrenceMarksSourceIncomplete(t *testing.T) {
	for _, arguments := range []string{`"{"`, `42`, `"{}"`, `"{\"name\":42}"`, `"{\"name\":\"\"}"`} {
		t.Run(arguments, func(t *testing.T) {
			input := `{"type":"session_meta","payload":{"id":"owner"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","name":"load_skill","call_id":"native-call","arguments":` + arguments + `}}`
			got, err := codex.ParseRollout(strings.NewReader(input))
			if err != nil {
				return
			}
			if len(got.Warnings) == 0 {
				t.Fatalf("DATA-01 malformed recognized occurrence became a complete empty owner: %+v", got)
			}
		})
	}
}

func TestCodexMalformedOuterResponsePayloadMarksSourceIncomplete(t *testing.T) {
	good := `{"type":"function_call","name":"load_skill","call_id":"native-call","arguments":"{\"name\":\"inspect-source\"}"}`
	parse := func(t *testing.T, payload string) codex.Transcript {
		t.Helper()
		input := `{"type":"session_meta","payload":{"id":"owner"}}` + "\n" + `{"type":"response_item","payload":` + payload + `}`
		got, err := codex.ParseRollout(strings.NewReader(input))
		if err != nil {
			t.Fatalf("valid JSONL should report an incomplete native payload: %v", err)
		}
		return got
	}
	control := parse(t, good)
	if len(control.Warnings) != 0 || len(control.Skills) != 1 || control.Skills[0].Provenance.CallID != "native-call" {
		t.Fatalf("valid native control was not complete: %+v", control)
	}
	cases := []struct {
		name, original, replacement, diagnostic string
	}{
		{"numeric_call_id", `"call_id":"native-call"`, `"call_id":42`, "invalid response_item payload"},
		{"missing_call_id", `,"call_id":"native-call"`, ``, "requires call_id"},
		{"blank_call_id", `"call_id":"native-call"`, `"call_id":" "`, "requires call_id"},
		{"numeric_name", `"name":"load_skill"`, `"name":42`, "invalid response_item payload"},
		{"missing_name", `"name":"load_skill",`, ``, "requires name"},
		{"blank_name", `"name":"load_skill"`, `"name":" "`, "requires name"},
		{"missing_type", `"type":"function_call",`, ``, "requires type"},
		{"numeric_type", `"type":"function_call"`, `"type":42`, "invalid response_item payload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := strings.Replace(good, tc.original, tc.replacement, 1)
			if bad == good {
				t.Fatal("fixture failed to corrupt native response payload")
			}
			got := parse(t, bad)
			for _, diagnostic := range []string{"codex-rollout-line:2", tc.diagnostic} {
				if !strings.Contains(strings.Join(got.Warnings, "; "), diagnostic) {
					t.Fatalf("DATA-01 malformed outer response payload omitted diagnostic %q: %+v", diagnostic, got.Warnings)
				}
			}
		})
	}
}

func TestCodexMissingRequestedTypeRoleAndTimeRemainNullableOccurrences(t *testing.T) {
	input := `{"type":"session_meta","payload":{"id":"owner"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","call_id":"skill","name":"load_skill","arguments":"{\"name\":\"inspect\"}"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{}"}}` + "\n"
	transcript, err := codex.ParseRollout(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(transcript.Warnings) != 0 {
		t.Fatalf("missing optional source values marked unsupported: %+v", transcript.Warnings)
	}
	sessions, err := transcript.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || len(sessions[0].Skills) != 1 || len(sessions[0].Delegations) != 1 {
		t.Fatalf("unknown source values discarded occurrences: %+v", sessions)
	}
	skill, delegation := sessions[0].Skills[0], sessions[0].Delegations[0]
	if skill.CallerRole != nil || skill.InvokedAt != nil || delegation.AgentType != nil || delegation.InvokedAt != nil || delegation.Description != "" || delegation.ChildSession != "" {
		t.Fatalf("source defaults invented: %+v %+v", skill, delegation)
	}
}

func TestRolloutActiveDurationCutoffAndTerminalBoundaries(t *testing.T) {
	input := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"active-root","source":"cli"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"event_msg","payload":{"type":"task_started","turn_id":"a"}}
{"timestamp":"2026-01-01T00:00:05Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"a"}}
{"timestamp":"2026-01-01T00:00:10Z","type":"event_msg","payload":{"type":"task_started","turn_id":"b"}}
`
	for _, tc := range []struct {
		name, tail                 string
		cutoff, total, provisional int64
		invalid                    bool
	}{
		{"open", "", 20, 14, 10, false},
		{"advance", "", 30, 24, 20, false},
		{"future_completion", `{"timestamp":"2026-01-01T00:00:40Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 20, 14, 10, false},
		{"complete", `{"timestamp":"2026-01-01T00:00:40Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 50, 34, 0, false},
		{"abort", `{"timestamp":"2026-01-01T00:00:15Z","type":"event_msg","payload":{"type":"turn_aborted","turn_id":"b"}}`, 20, 9, 0, false},
		{"reversed_completion", `{"timestamp":"2026-01-01T00:00:09Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 20, 4, 0, true},
		{"malformed_completion", `{"timestamp":"not-a-date","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 20, 4, 0, true},
		{"malformed_start", `{"timestamp":"2026-01-01T00:00:15Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 20, 4, 0, true},
		{"missing_start", `{"timestamp":"2026-01-01T00:00:15Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"b"}}`, 20, 4, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cutoff := time.Date(2026, 1, 1, 0, 0, int(tc.cutoff), 0, time.UTC)
			records := input
			switch tc.name {
			case "malformed_start":
				records = strings.Replace(input, `"timestamp":"2026-01-01T00:00:10Z"`, `"timestamp":"not-a-date"`, 1)
			case "missing_start":
				records = strings.Replace(input, `"timestamp":"2026-01-01T00:00:10Z",`, "", 1)
			}
			for range 2 {
				tr, err := codex.ParseRolloutAt(strings.NewReader(records+tc.tail), cutoff)
				if err != nil {
					t.Fatal(err)
				}
				inv := tr.Invocations[0]
				if inv.ActiveDurationSeconds != tc.total || inv.ProvisionalSeconds != tc.provisional || tr.ActiveSeconds != tc.total {
					t.Fatalf("duration %+v, transcript total %d", inv, tr.ActiveSeconds)
				}
				if tc.invalid {
					if inv.ProvisionalSeconds != 0 || inv.CompletedAt == nil || *inv.CompletedAt != "2026-01-01T00:00:05Z" {
						t.Fatalf("invalid lifecycle invented provisional/completed boundary: %+v", inv)
					}
					if len(tr.Warnings) == 0 {
						t.Fatal("invalid lifecycle lost ordinary source diagnostic")
					}
				}
			}
		})
	}
}

func TestCodexCheckpointNativeIntervalsBeforeModelGrouping(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC)
	for _, tc := range []struct {
		name         string
		starts       []int64
		wantCases    []domain.CheckpointCase
		wantSubtotal int64
	}{
		{"rounded_short", []int64{179999}, nil, 180},
		{"exact_threshold", []int64{180000}, []domain.CheckpointCase{domain.CheckpointUnknownCaller}, 180},
		{"missing_native_turn_id", []int64{180000}, []domain.CheckpointCase{domain.CheckpointUnavailableInterval}, 180},
		{"future_native_completion", []int64{180000}, []domain.CheckpointCase{domain.CheckpointUnavailableInterval}, 180},
		{"superseded_short_does_not_pool", []int64{100000, 90000}, nil, 90},
		{"superseded_long_keeps_partial_qualification", []int64{200000, 180000}, []domain.CheckpointCase{domain.CheckpointUnavailableInterval, domain.CheckpointUnknownCaller}, 180},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"owner"}}` + "\n"
			for i, age := range tc.starts {
				timestamp := cutoff.Add(-time.Duration(age) * time.Millisecond).Format(time.RFC3339Nano)
				input += fmt.Sprintf(`{"timestamp":%q,"type":"turn_context","payload":{"model":"m","turn_id":"native-%d"}}`+"\n", timestamp, i)
				input += fmt.Sprintf(`{"timestamp":%q,"type":"event_msg","payload":{"type":"task_started","turn_id":"native-%d"}}`+"\n", timestamp, i)
			}
			if tc.name == "missing_native_turn_id" {
				input = strings.ReplaceAll(input, `,"turn_id":"native-0"`, "")
			}
			if tc.name == "future_native_completion" {
				input += fmt.Sprintf(`{"timestamp":%q,"type":"event_msg","payload":{"type":"task_complete","turn_id":"native-0"}}`+"\n", cutoff.Add(time.Second).Format(time.RFC3339Nano))
			}
			tr, err := codex.ParseRolloutAt(strings.NewReader(input), cutoff)
			if err != nil {
				t.Fatal(err)
			}
			if len(tr.Invocations) != 1 || tr.Invocations[0].ProvisionalSeconds != tc.wantSubtotal {
				t.Fatalf("numeric grouping changed: %+v", tr.Invocations)
			}
			session := domain.SessionFact{ID: "owner", Source: domain.SourceCodex, Invocations: tr.Invocations}
			got := domain.SelectCheckpointFeedback([]domain.SessionFact{session}, tr.CheckpointEvidence, domain.CheckpointCaller{}, cutoff, 180)
			if len(got) != len(tc.wantCases) {
				t.Fatalf("native age/subtotal classification: %+v", got)
			}
			for i, item := range got {
				if item.Case != tc.wantCases[i] || (item.Case != domain.CheckpointUnavailableInterval && item.Interval.ID == "") || item.ProvisionalSeconds != tc.wantSubtotal {
					t.Fatalf("lost native identity or subtotal: %+v", got)
				}
			}
		})
	}
}
