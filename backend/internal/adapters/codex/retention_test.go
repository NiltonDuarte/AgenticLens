package codex_test

import (
	"fmt"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

func TestRoutineCodexEventsUseRequestContextInsteadOfCumulativeUsage(t *testing.T) {
	for _, usageSource := range []string{"token_usage_record", "token_count"} {
		t.Run(usageSource, func(t *testing.T) {
			lines := []string{`{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"thread-retention"}}`}
			for turn := 1; turn <= 8; turn++ {
				lines = append(lines, fmt.Sprintf(`{"timestamp":"2026-10-25T09:%02d:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-%d"}}`, turn, turn))
				if usageSource == "token_usage_record" {
					lines = append(lines, fmt.Sprintf(`{"timestamp":"2026-10-25T09:%02d:01Z","type":"token_usage_record","payload":{"turn_id":"turn-%d","usage":{"input_tokens":30000,"cached_input_tokens":10000,"output_tokens":100}}}`, turn, turn))
				} else {
					lines = append(lines, fmt.Sprintf(`{"timestamp":"2026-10-25T09:%02d:01Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":%d,"cached_input_tokens":%d,"output_tokens":%d},"last_token_usage":{"input_tokens":30000}}}}`, turn, turn*30000, turn*10000, turn*100))
				}
				lines = append(lines,
					fmt.Sprintf(`{"timestamp":"2026-10-25T09:%02d:02Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-%d","item":{"id":"command-%d","type":"CommandExecution","status":"completed","exit_code":0}}}`, turn, turn, turn),
					fmt.Sprintf(`{"timestamp":"2026-10-25T09:%02d:03Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-%d"}}`, turn, turn))
			}
			transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(lines, "\n")))
			if err != nil {
				t.Fatal(err)
			}
			if transcript.Usage.Input != 240000 || len(transcript.Events) != 8 {
				t.Fatalf("expected full aggregate usage and source events: input=%d events=%d", transcript.Usage.Input, len(transcript.Events))
			}
			for _, event := range transcript.Events {
				if event.ContextTokens != 30000 {
					t.Fatalf("event context is not the last request's input: %d", event.ContextTokens)
				}
			}
			ledger := transcript.ToMetricsLedger("GEN-07", "Retention regression", domain.TicketTypeChore, domain.DefaultProjectConfig())
			for _, anomaly := range ledger.Anomalies {
				if anomaly.RuleID == domain.RuleContextSaturationSpike {
					t.Fatalf("cumulative usage triggered a false context anomaly: %#v", anomaly)
				}
			}
			if len(ledger.Facts.Events) != 0 || ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 8 {
				t.Fatalf("routine events should be pruned without losing tool totals: retained=%d calls=%d", len(ledger.Facts.Events), ledger.WorkflowEfficiency.ToolExecution.TotalCalls)
			}
		})
	}
}

func TestCodexLedgerRetainsErrorsCompactionsAndAnomalyTurns(t *testing.T) {
	tool := "command"
	transcript := codex.Transcript{
		Tools: []domain.ToolFact{{Name: tool, Calls: 4, Errors: 1}},
		Events: []domain.EventFact{
			{Timestamp: "2026-10-25T09:00:00Z", SessionID: "thread", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 1000},
			{Timestamp: "2026-10-25T09:01:00Z", SessionID: "thread", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error", ContextTokens: 1000},
			{Timestamp: "2026-10-25T09:02:00Z", SessionID: "thread", Turn: 2, EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: 1000},
			{Timestamp: "2026-10-25T09:03:00Z", SessionID: "thread", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 100000},
			{Timestamp: "2026-10-25T09:04:00Z", SessionID: "thread", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 100000},
		},
	}
	ledger := transcript.ToMetricsLedger("GEN-07", "Retention regression", domain.TicketTypeChore, domain.DefaultProjectConfig())
	retained := ledger.Facts.Events
	if len(retained) != 3 || retained[0].Status != "error" || retained[1].EventType != domain.EventTypeCompaction || retained[2].Turn != 3 {
		t.Fatalf("expected error, compaction, and real context-spike turn only: %#v", retained)
	}
	if ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 4 || len(transcript.Events) != 5 {
		t.Fatal("pruning changed aggregate calls or the full source event list")
	}
}
