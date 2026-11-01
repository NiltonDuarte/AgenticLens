package live_test

import (
	"context"
	"database/sql"
	"testing"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func TestOfflineAndLiveToolExecutionParity(t *testing.T) {
	dbPath := createTestDBFile(t)
	writeDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = writeDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('ses_parity', NULL, 'build', 'model', 'AL-42 parity', 1760000000000, 1760000010000);
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('msg_parity', 'ses_parity', 1760000001000, 1760000010000, '{"role":"assistant","tokens":{"input":1}}');
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_semantic', 'ses_parity', 'msg_parity', 1760000002000, 1760000002100, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"syntax error in docs"}}'),
		('p_nomatch', 'ses_parity', 'msg_parity', 1760000003000, 1760000003100, '{"type":"tool","tool":"grep","state":{"status":"completed","output":"exit status 1"}}'),
		('p_failure_before_1', 'ses_parity', 'msg_parity', 1760000004000, 1760000004100, '{"type":"tool","tool":"edit","state":{"status":"failed","output":"write failed"}}'),
		('p_failure_before_2', 'ses_parity', 'msg_parity', 1760000005000, 1760000005100, '{"type":"tool","tool":"edit","state":{"status":"failed","output":"write failed again"}}'),
		('p_timeout_identifier', 'ses_parity', 'msg_parity', 1760000006000, 1760000006100, '{"type":"tool","tool":"read","state":{"status":"completed","output":"const timeoutMs = 5000;"}}'),
		('p_saved_panic', 'ses_parity', 'msg_parity', 1760000007000, 1760000007100, '{"type":"tool","tool":"grep","state":{"status":"completed","output":"logs/example.txt: panic: saved example"}}'),
		('p_zero_exit_fixture', 'ses_parity', 'msg_parity', 1760000008000, 1760000008100, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"exit code 0\nfixture says permission denied"}}'),
		('p_failure', 'ses_parity', 'msg_parity', 1760000009000, 1760000009100, '{"type":"tool","tool":"edit","state":{"status":"failed","output":"write failed after successful inspections"}}');`)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeDB.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	sessions, _, _, err := opencode.ExtractSessions(db, []string{"ses_parity"}, "user", nil)
	facts := domain.AgenticFactLedger{Sessions: sessions}
	tools, offlineEvents := facts.ToolFacts(), facts.EventFacts()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	trace, err := service.GetSessionTrace(context.Background(), "ses_parity")
	if err != nil {
		t.Fatal(err)
	}

	if len(tools) != 4 {
		t.Fatalf("expected 4 offline tool facts, got %d", len(tools))
	}
	var offlineErrors, offlineCalls int64
	for _, tool := range tools {
		offlineErrors += tool.Errors
		offlineCalls += tool.Calls
	}
	if offlineErrors != 3 || offlineCalls != 8 || trace.FailedToolCalls != 3 || trace.TotalToolCalls != 8 {
		t.Fatalf("execution totals changed: offline errors=%d calls=%d, live failed calls=%d calls=%d", offlineErrors, offlineCalls, trace.FailedToolCalls, trace.TotalToolCalls)
	}

	outputs := map[string]string{
		"p_timeout_identifier": "const timeoutMs = 5000;",
		"p_saved_panic":        "logs/example.txt: panic: saved example",
		"p_zero_exit_fixture":  "exit code 0\nfixture says permission denied",
	}
	statuses := make(map[string]string)
	for _, event := range offlineEvents {
		if event.ToolName != nil {
			statuses[event.EventID] = event.Status
			for id, output := range outputs {
				if event.EventID == "opencode:part:"+id && (event.ErrorMessage != nil || event.ToolOutput == nil || *event.ToolOutput != output) {
					t.Errorf("successful offline source evidence changed: %+v", event)
				}
			}
		}
	}
	liveStatuses := make(map[string]string)
	for _, message := range trace.Messages {
		for _, part := range message.Parts {
			if part.Type == "tool" {
				liveStatuses["opencode:part:"+part.ID] = part.Status
				if output, ok := outputs[part.ID]; ok && (part.IsError || part.Severity != domain.SeverityNone || part.ErrorMessage != nil || part.OutputSnippet != output) {
					t.Errorf("successful live source evidence changed: %+v", part)
				}
			}
		}
	}
	if len(statuses) != len(liveStatuses) {
		t.Fatalf("offline statuses=%v, live statuses=%v", statuses, liveStatuses)
	}
	for id, want := range map[string]string{
		"opencode:part:p_semantic":           "completed",
		"opencode:part:p_nomatch":            "completed",
		"opencode:part:p_failure_before_1":   "error",
		"opencode:part:p_failure_before_2":   "error",
		"opencode:part:p_timeout_identifier": "completed",
		"opencode:part:p_saved_panic":        "completed",
		"opencode:part:p_zero_exit_fixture":  "completed",
		"opencode:part:p_failure":            "error",
	} {
		if statuses[id] != want || liveStatuses[id] != want {
			t.Errorf("event %s: offline=%s live=%s want=%s", id, statuses[id], liveStatuses[id], want)
		}
	}
	// Two genuine failures precede the successful inspection output. That success
	// must reset the streak so the final genuine failure remains medium severity.
	if len(trace.RecentErrors) != 3 || trace.RecentErrors[0].ID != "p_failure" || trace.RecentErrors[0].Severity != domain.SeverityMedium {
		t.Fatalf("successful calls did not reset failure escalation: %+v", trace.RecentErrors)
	}
	for _, anomaly := range trace.Anomalies {
		if anomaly.RuleName == domain.RuleConsecutiveToolErrors {
			t.Fatalf("successful calls inflated the failure streak: %+v", anomaly)
		}
	}
	offlineSessions, err := domain.SessionsFromFactsWithSource(domain.SourceOpenCode, nil, tools, offlineEvents)
	if err != nil {
		t.Fatal(err)
	}
	for _, anomaly := range domain.DetectAnomalies(domain.AgenticFactLedger{Sessions: offlineSessions}) {
		if anomaly.RuleID == domain.RuleToolRetryStreak {
			t.Fatalf("successful offline calls inflated the retry streak: %+v", anomaly)
		}
	}
}
