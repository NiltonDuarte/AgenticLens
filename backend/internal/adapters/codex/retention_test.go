package codex_test

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func TestCodexAndOpenCodeAdaptersPersistEquivalentCompactSummaries(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		INSERT INTO session VALUES ('shared-session', NULL, 'coder', 'model', 'fixture', 1792904400000, 1792904410000);
	`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 5; i++ {
		status := "completed"
		if i <= 3 {
			status = "error"
		}
		_, err = db.Exec(`INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES (?, 'shared-session', ?, ?, ?, ?)`,
			fmt.Sprintf("part-%d", i), fmt.Sprintf("message-%d", i), 1792904400000+i*1000, 1792904400000+i*1000,
			fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":%q}}`, status))
		if err != nil {
			t.Fatal(err)
		}
	}
	opencodeSessions, _, _, err := opencode.ExtractSessions(db, []string{"shared-session"}, "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	opencodeTransient := domain.AgenticFactLedger{Sessions: opencodeSessions}
	opencodeLiveAnomalies, err := domain.DetectAnomaliesWithEvidence(opencodeTransient)
	if err != nil {
		t.Fatal(err)
	}
	opencodeFacts, err := domain.PersistAnomalySummaries(opencodeTransient)
	if err != nil {
		t.Fatal(err)
	}

	lines := []string{`{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"shared-session"}}`}
	for i := 1; i <= 5; i++ {
		status := "completed"
		if i <= 3 {
			status = "error"
		}
		lines = append(lines, fmt.Sprintf(`{"timestamp":"2026-10-25T09:00:%02dZ","type":"event_msg","payload":{"type":"mcp_tool_call_end","call_id":"part-%d","turn_id":"turn-%d","status":%q,"invocation":{"tool":"bash"}}}`, i, i, i, status))
	}
	transcript, err := codex.ParseRollout(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	codexTransientSessions, err := domain.SessionsFromFactsWithSource(domain.SourceCodex, nil, nil, transcript.Events)
	if err != nil {
		t.Fatal(err)
	}
	codexLiveAnomalies, err := domain.DetectAnomaliesWithEvidence(domain.AgenticFactLedger{Sessions: codexTransientSessions})
	if err != nil {
		t.Fatal(err)
	}
	if len(opencodeLiveAnomalies) != len(codexLiveAnomalies) {
		t.Fatalf("transient summary parity mismatch: opencode=%#v codex=%#v", opencodeLiveAnomalies, codexLiveAnomalies)
	}
	for i := range opencodeLiveAnomalies {
		if opencodeLiveAnomalies[i].RuleID != codexLiveAnomalies[i].RuleID || len(opencodeLiveAnomalies[i].Evidence) != len(codexLiveAnomalies[i].Evidence) {
			t.Fatalf("transient event association mismatch: opencode=%#v codex=%#v", opencodeLiveAnomalies[i], codexLiveAnomalies[i])
		}
	}
	codexLedger, err := transcript.ToMetricsLedger("", "", "", domain.DefaultProjectConfig())
	if err != nil {
		t.Fatal(err)
	}

	opencodeAnomalies, codexAnomalies := opencodeFacts.AnomalyFacts(), codexLedger.Facts.AnomalyFacts()
	if len(opencodeAnomalies) != len(codexAnomalies) {
		t.Fatalf("summary parity mismatch: opencode=%#v codex=%#v", opencodeAnomalies, codexAnomalies)
	}
	for i, anomaly := range opencodeAnomalies {
		other := codexAnomalies[i]
		if anomaly.RuleID != other.RuleID || anomaly.Context.Count == nil || other.Context.Count == nil || *anomaly.Context.Count != *other.Context.Count {
			t.Fatalf("summary mismatch at %d: opencode=%#v codex=%#v", i, anomaly, other)
		}
	}
	for _, facts := range []domain.AgenticFactLedger{opencodeFacts, codexLedger.Facts} {
		if len(facts.EventFacts()) != 0 {
			t.Fatalf("persisted ledger retained events: %#v", facts.EventFacts())
		}
		for _, anomaly := range facts.AnomalyFacts() {
			if len(anomaly.Evidence) != 0 {
				t.Fatalf("persisted summary retained evidence refs: %#v", anomaly.Evidence)
			}
		}
	}
}

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
			ledger, err := transcript.ToMetricsLedger("TEST-101", "Retention regression", domain.TicketTypeChore, domain.DefaultProjectConfig())
			if err != nil {
				t.Fatal(err)
			}
			for _, anomaly := range ledger.Anomalies {
				if anomaly.RuleID == domain.RuleContextSaturationSpike {
					t.Fatalf("cumulative usage triggered a false context anomaly: %#v", anomaly)
				}
			}
			if len(ledger.Facts.EventFacts()) != 0 || ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 8 {
				t.Fatalf("events should be omitted without losing tool totals: retained=%d calls=%d", len(ledger.Facts.EventFacts()), ledger.WorkflowEfficiency.ToolExecution.TotalCalls)
			}
		})
	}
}

func TestCodexLedgerOmitsEventsButRetainsAnomalySummaries(t *testing.T) {
	tool := "command"
	transcript := codex.Transcript{
		ThreadID: "thread",
		Tools:    []domain.ToolFact{{SessionID: "thread", Name: tool, Calls: 4, Errors: 1}},
		Events: []domain.EventFact{
			{EventID: "codex-event-1", Timestamp: "2026-10-25T09:00:00Z", SessionID: "thread", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 1000},
			{EventID: "codex-event-2", Timestamp: "2026-10-25T09:01:00Z", SessionID: "thread", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error", ContextTokens: 1000},
			{EventID: "codex-event-3", Timestamp: "2026-10-25T09:02:00Z", SessionID: "thread", Turn: 2, EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: 1000},
			{EventID: "codex-event-4", Timestamp: "2026-10-25T09:03:00Z", SessionID: "thread", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 100000},
			{EventID: "codex-event-5", Timestamp: "2026-10-25T09:04:00Z", SessionID: "thread", Turn: 4, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", ContextTokens: 100000},
		},
	}
	ledger, err := transcript.ToMetricsLedger("TEST-101", "Retention regression", domain.TicketTypeChore, domain.DefaultProjectConfig())
	if err != nil {
		t.Fatal(err)
	}
	if retained := ledger.Facts.EventFacts(); len(retained) != 0 {
		t.Fatalf("expected no persisted events: %#v", retained)
	}
	if len(ledger.Anomalies) == 0 {
		t.Fatal("expected anomaly summaries to be retained")
	}
	if ledger.WorkflowEfficiency.ToolExecution.TotalCalls != 4 || len(transcript.Events) != 5 {
		t.Fatal("pruning changed aggregate calls or the full source event list")
	}
}
