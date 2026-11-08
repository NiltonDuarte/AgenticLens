package live_test

import (
	"agenticlens/internal/adapters"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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
	tools := facts.ToolFacts()
	offlineEvents, sourceErr := opencode.ExtractSessionSourceEvents(db, "ses_parity", "user")
	if sourceErr != nil {
		t.Fatal(sourceErr)
	}
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
	policy := testProjectConfig().AnomalyThresholds
	for _, anomaly := range domain.DetectAnomaliesFromEvents(domain.AgenticFactLedger{Sessions: offlineSessions}, offlineEvents, policy) {
		if anomaly.RuleID == domain.RuleToolRetryStreak {
			t.Fatalf("successful offline calls inflated the retry streak: %+v", anomaly)
		}
	}
	// Independent positive sequence makes the negative reset oracle non-vacuous.
	tool := "edit"
	positive := []domain.EventFact{}
	for i := 0; i < 3; i++ {
		positive = append(positive, domain.EventFact{SessionID: "positive", EventID: fmt.Sprintf("failure-%d", i), Timestamp: fmt.Sprintf("2026-11-08T10:00:0%dZ", i), Turn: int64(i + 1), EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"})
	}
	found := false
	for _, a := range domain.DetectAnomaliesFromEvents(domain.AgenticFactLedger{}, positive, policy) {
		found = found || (a.RuleID == domain.RuleToolRetryStreak && a.Context.Count != nil && *a.Context.Count == 3)
	}
	if !found {
		t.Fatal("positive source-event control did not reach retry detection")
	}
}

// Synthetic OpenCode v1 records exercise the real SQLite/live/extraction boundaries.
func TestV1StructuredExecutionAndVerbatimSourceParity(t *testing.T) {
	const rawEvidence = "  segmentation fault\nexit code 2\n\tprivate-raw-source-marker  "
	encoded, _ := json.Marshal(rawEvidence)
	for _, tc := range []struct {
		name, tool, status, output, metadata, wantOutput, wantStatus string
		wantError                                                    bool
	}{
		{"read inspection", "read", "completed", string(encoded), `{}`, rawEvidence, "completed", false},
		{"shell inspection", "bash", "completed", string(encoded), `{}`, rawEvidence, "completed", false},
		{"native nonzero exit", "bash", "completed", string(encoded), `{"exit":2}`, rawEvidence, "error", true},
		{"native zero exit", "bash", "completed", string(encoded), `{"exit":0}`, rawEvidence, "completed", false},
		{"malformed exit unknown", "bash", "completed", string(encoded), `{"exit":"2"}`, rawEvidence, "completed", false},
		{"null exit unknown", "bash", "completed", string(encoded), `{"exit":null}`, rawEvidence, "completed", false},
		{"negative native exit", "bash", "completed", string(encoded), `{"exit":-1}`, rawEvidence, "error", true},
		{"explicit failed status", "bash", "failed", string(encoded), `{}`, rawEvidence, "error", true},
		{"unsupported exit alias", "bash", "completed", string(encoded), `{"exit_code":2}`, rawEvidence, "completed", false},
		{"non-execution metadata", "read", "completed", string(encoded), `{"exit":2}`, rawEvidence, "completed", false},
		{"unknown state", "bash", "unrecognized", string(encoded), `{}`, rawEvidence, "unrecognized", false},
		{"missing status", "bash", "", string(encoded), `{}`, rawEvidence, "unknown", false},
		{"malformed numeric status", "bash", "<number>", string(encoded), `{}`, rawEvidence, "unknown", false},
		{"malformed optional error", "bash", "error", string(encoded), `{}`, rawEvidence, "error", true},
		{"structured output", "read", "completed", `{ "source": "  private-raw-source-marker  " }`, `{}`, `{ "source": "  private-raw-source-marker  " }`, "completed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dbPath := createTestDBFile(t)
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err = db.Exec(`INSERT INTO session (id,parent_id,agent,model,title,time_created,time_updated) VALUES ('source',NULL,'coder','model','source',1760000000000,1760000010000); INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('message','source',1760000001000,1760000010000,'{"role":"assistant","tokens":{"input":1}}')`); err != nil {
				t.Fatal(err)
			}
			statusJSON, _ := json.Marshal(tc.status)
			if tc.status == "<number>" {
				statusJSON = []byte("17")
			}
			payload := fmt.Sprintf(`{"type":"tool","tool":%q,"state":{"status":%s,"output":%s,"metadata":%s}}`, tc.tool, statusJSON, tc.output, tc.metadata)
			if tc.name == "malformed optional error" {
				payload = strings.Replace(payload, `"state":{`, `"state":{"error":{"unrecognized":"source detail"},`, 1)
			}
			if _, err = db.Exec(`INSERT INTO part (id,session_id,message_id,time_created,time_updated,data) VALUES ('part','source','message',1760000002000,1760000003000,?)`, payload); err != nil {
				t.Fatal(err)
			}
			owners, _, _, err := opencode.ExtractSessions(db, []string{"source"}, "user", nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(owners) != 1 || owners[0].Anomalies != nil {
				t.Fatalf("cfg-less raw source invented complete anomaly analysis: %+v", owners)
			}
			events, err := opencode.ExtractSessionSourceEvents(db, "source", "user")
			if err != nil {
				t.Fatal(err)
			}
			var offline *domain.EventFact
			for i := range events {
				if events[i].EventID == "opencode:part:part" {
					offline = &events[i]
				}
			}
			hook, err := live.OpenReadOnlyHook(dbPath, 5000)
			if err != nil {
				t.Fatal(err)
			}
			service := live.NewLiveService(hook, testProjectConfig())
			defer service.Close()
			trace, err := service.GetSessionTrace(context.Background(), "source")
			if err != nil {
				t.Fatal(err)
			}
			if offline == nil || len(trace.Messages) != 1 || len(trace.Messages[0].Parts) != 1 {
				t.Fatalf("source records unavailable offline=%+v trace=%+v", offline, trace)
			}
			part := trace.Messages[0].Parts[0]
			if offline.Status != tc.wantStatus || part.Status != tc.wantStatus || part.IsError != tc.wantError {
				t.Fatalf("classification differs: offline=%+v live=%+v", offline, part)
			}
			if offline.ToolOutput == nil || *offline.ToolOutput != tc.wantOutput || part.OutputSnippet != tc.wantOutput {
				t.Fatalf("raw source changed offline=%+v live=%q expected=%q", offline.ToolOutput, part.OutputSnippet, tc.wantOutput)
			}
			if trace.Relationships.State != domain.LiveRelationshipsComplete || len(trace.Diagnostics) != 0 {
				t.Fatalf("successful empty child lookup not complete: %+v", trace)
			}
			serialized, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(domain.AgenticFactLedger{Sessions: owners}))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(serialized), "private-raw-source-marker") {
				t.Fatalf("raw source entered persisted ledger: %s", serialized)
			}
		})
	}
}

func TestV1ErrorOnlyLongEvidenceSurvivesLiveAndOffline(t *testing.T) {
	dbPath := createTestDBFile(t)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	raw := " \t" + strings.Repeat("private-error-only-source-marker\n", 100) + "  \n"
	encoded, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO session(id,parent_id,agent,model,title,time_created,time_updated) VALUES('owner',NULL,'coder','model','fixture',1760000000000,1760000010000);INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES('message','owner',1760000001000,1760000010000,'{"role":"assistant","tokens":{"input":1}}')`); err != nil {
		t.Fatal(err)
	}
	payload := fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":"error","error":%s}}`, encoded)
	if _, err = db.Exec(`INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES('error-only','owner','message',1760000002000,1760000003000,?)`, payload); err != nil {
		t.Fatal(err)
	}
	owners, _, _, err := opencode.ExtractSessions(db, []string{"owner"}, "user", nil)
	if err != nil {
		t.Fatal(err)
	}
	events, err := opencode.ExtractSessionSourceEvents(db, "owner", "user")
	if err != nil {
		t.Fatal(err)
	}
	var source *domain.EventFact
	for i := range events {
		if events[i].EventID == "opencode:part:error-only" {
			source = &events[i]
		}
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	trace, err := service.GetSessionTrace(context.Background(), "owner")
	if err != nil {
		t.Fatal(err)
	}
	if source == nil || len(trace.Messages) != 1 || len(trace.Messages[0].Parts) != 1 {
		t.Fatalf("supported error-only record disappeared: source=%+v trace=%+v", source, trace)
	}
	part := trace.Messages[0].Parts[0]
	if source.ToolOutput == nil || *source.ToolOutput != raw || part.OutputSnippet != raw {
		t.Fatalf("full error evidence diverged: offlineOutput=%v liveBytes=%d wantBytes=%d", source.ToolOutput, len(part.OutputSnippet), len(raw))
	}
	if source.ErrorMessage == nil || len(*source.ErrorMessage) != 2048 || part.ErrorMessage == nil || len(*part.ErrorMessage) != 2048 {
		t.Fatalf("bounded diagnostics changed: source=%+v part=%+v", source, part)
	}
	wire, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(domain.AgenticFactLedger{Sessions: owners}))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(wire), "private-error-only-source-marker") {
		t.Fatalf("raw error body leaked into ledger: %s", wire)
	}
}

func TestLiveServiceAndHydrationReceiveConfiguredFailureRatePolicy(t *testing.T) {
	dbPath := createTestDBFile(t)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO session(id,parent_id,agent,model,title,time_created,time_updated) VALUES('owner',NULL,'coder','model','fixture',1760000000000,1760000010000);INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES('message','owner',1760000001000,1760000010000,'{"role":"assistant","tokens":{"input":1}}');INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES('first','owner','message',1760000002000,1760000002100,'{"type":"tool","tool":"read","state":{"status":"error","output":"failure"}}'),('second','owner','message',1760000003000,1760000003100,'{"type":"tool","tool":"read","state":{"status":"completed","output":"quoted crash"}}'),('third','owner','message',1760000004000,1760000004100,'{"type":"tool","tool":"read","state":{"status":"error","output":"failure"}}')`); err != nil {
		t.Fatal(err)
	}
	owners, _, _, err := opencode.ExtractSessions(db, []string{"owner"}, "user", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		minimum  int
		critical float64
		severity domain.AnomalySeverity
	}{{3, .8, domain.SeverityWarning}, {4, .8, ""}, {3, .6, domain.SeverityCritical}} {
		cfg := testProjectConfig()
		cfg.AnomalyThresholds.ErrorRateMinCalls = tc.minimum
		cfg.AnomalyThresholds.ErrorRateThreshold = .4
		cfg.AnomalyThresholds.ErrorRateCriticalThreshold = tc.critical
		hook, err := live.OpenReadOnlyHook(dbPath, 5000)
		if err != nil {
			t.Fatal(err)
		}
		service := live.NewLiveService(hook, cfg)
		trace, err := service.GetSessionTrace(context.Background(), "owner")
		if err != nil {
			service.Close()
			t.Fatal(err)
		}
		var liveRates []domain.LiveAnomalyAlert
		for _, a := range trace.Anomalies {
			if a.RuleName == domain.RuleLiveToolHighFailureRate {
				liveRates = append(liveRates, a)
			}
		}
		expected := 0
		if tc.severity != "" {
			expected = 1
		}
		if len(liveRates) != expected || (expected == 1 && string(liveRates[0].Severity) != string(tc.severity)) {
			service.Close()
			t.Fatalf("live service dropped configured policy: %+v", liveRates)
		}
		response, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-1", domain.DeriveMetricsLedger(domain.AgenticFactLedger{Sessions: owners}, cfg))
		service.Close()
		if err != nil {
			t.Fatal(err)
		}
		var hydratedRates []domain.Anomaly
		for _, a := range response.Anomalies {
			if a.RuleID == domain.RuleToolHighFailureRate {
				hydratedRates = append(hydratedRates, a)
			}
		}
		if len(hydratedRates) != expected || (expected == 1 && hydratedRates[0].Severity != tc.severity) {
			t.Fatalf("hydration dropped configured policy: %+v", response)
		}
	}
}
