package live_test

import (
	"context"
	"database/sql"
	"testing"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/adapters/opencode"
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
		VALUES ('ses_parity', NULL, 'build', 'model', 'AL-42 parity', 1760000000000, 1760000005000);
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('msg_parity', 'ses_parity', 1760000001000, 1760000005000, '{"role":"assistant","tokens":{"input":1}}');
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_semantic', 'ses_parity', 'msg_parity', 1760000002000, 1760000002100, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"syntax error in docs"}}'),
		('p_nomatch', 'ses_parity', 'msg_parity', 1760000003000, 1760000003100, '{"type":"tool","tool":"grep","state":{"status":"completed","output":"exit status 1"}}'),
		('p_failure', 'ses_parity', 'msg_parity', 1760000004000, 1760000004100, '{"type":"tool","tool":"edit","state":{"status":"failed","output":"write failed"}}');`)
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
	_, tools, offlineEvents, _, _, err := opencode.ExtractFacts(db, []string{"ses_parity"}, "user", nil)
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

	if len(tools) != 3 {
		t.Fatalf("expected 3 offline tool facts, got %d", len(tools))
	}
	var offlineErrors int64
	for _, tool := range tools {
		offlineErrors += tool.Errors
	}
	if offlineErrors != 1 || trace.FailedToolCalls != 1 || trace.TotalToolCalls != 3 {
		t.Fatalf("execution totals changed: offline errors=%d, live failed calls=%d, live calls=%d", offlineErrors, trace.FailedToolCalls, trace.TotalToolCalls)
	}

	statuses := make(map[string]string)
	for _, event := range offlineEvents {
		if event.ToolName != nil {
			statuses[event.EventID] = event.Status
		}
	}
	liveStatuses := make(map[string]string)
	for _, message := range trace.Messages {
		for _, part := range message.Parts {
			if part.Type == "tool" {
				liveStatuses["opencode:part:"+part.ID] = part.Status
			}
		}
	}
	if len(statuses) != len(liveStatuses) {
		t.Fatalf("offline statuses=%v, live statuses=%v", statuses, liveStatuses)
	}
	for id, want := range map[string]string{
		"opencode:part:p_semantic": "completed",
		"opencode:part:p_nomatch":  "completed",
		"opencode:part:p_failure":  "error",
	} {
		if statuses[id] != want || liveStatuses[id] != want {
			t.Errorf("event %s: offline=%s live=%s want=%s", id, statuses[id], liveStatuses[id], want)
		}
	}
}
