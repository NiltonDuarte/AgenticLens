package opencode_test

import (
	"database/sql"
	"testing"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func createSchemaV2TestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);
INSERT INTO session_v2 VALUES('ses_root',NULL,'build','{"id":"mock-model","providerID":"al58local"}','AL-58 v2',1793231969869,1793231970472);
INSERT INTO session_message VALUES
('u','ses_root','user',5,1793231970186,1793231970189,'{"time":{"created":1793231970186},"text":"\"Read marker.txt and report its contents.\""}'),
('a1','ses_root','assistant',6,1793231970238,1793231970363,'{"time":{"created":1793231970238,"completed":1793231970358},"agent":"build","model":{"id":"mock-model","providerID":"al58local"},"content":[{"type":"tool","id":"call1","name":"read","state":{"status":"completed","input":{"path":"marker.txt"},"content":[{"type":"text","text":"found marker"}],"time":{"created":1793231970288,"completed":1793231970345}}}],"cost":0,"tokens":{"input":23,"output":11,"reasoning":0,"cache":{"read":0,"write":0}}}'),
('a2','ses_root','assistant',13,1793231970413,1793231970476,'{"time":{"created":1793231970413,"completed":1793231970472},"model":{"id":"mock-model","providerID":"al58local"},"content":[{"type":"text","text":"AL-58 deterministic runtime reply."}],"cost":0,"tokens":{"input":23,"output":11,"reasoning":0,"cache":{"read":0,"write":0}}}'),
('i','ses_root','idle',18,1793231970482,1793231970485,'{"time":{"created":1793231970482},"outcome":"succeeded"}');`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchemaV2SessionRuntimeEvidenceExtractsNormalizedFacts(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	id, err := opencode.ResolveActiveRootSession(db, "", "AL-58", "")
	if err != nil || id == nil || *id != "ses_root" {
		t.Fatalf("resolve=%v err=%v", id, err)
	}
	invs, tools, events, start, end, err := opencode.ExtractFacts(db, []string{"ses_root"}, "tester", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(invs) != 1 || invs[0].Model != "mock-model" || invs[0].Tokens.Input != 46 || invs[0].Tokens.Output != 22 || invs[0].CostUSD != 0 {
		t.Fatalf("unexpected v2 invocation: %+v", invs)
	}
	if start == nil || end == nil || len(tools) != 1 || tools[0].Name != "read" || tools[0].Calls != 1 || tools[0].Errors != 0 {
		t.Fatalf("unexpected bounds/tools: %v %v %+v", start, end, tools)
	}
	turns, err := opencode.ExtractSteeringTurns(db, []string{"ses_root"})
	if err != nil || turns != 1 {
		t.Fatalf("user turns=%d err=%v", turns, err)
	}
	var assistant, tool int
	for _, e := range events {
		if e.EventType == domain.EventTypeTurnMessage && e.MessageID != nil && *e.MessageID != "u" {
			assistant++
		}
		if e.EventType == domain.EventTypeToolCall && e.ToolName != nil && *e.ToolName == "read" && e.Status == "completed" {
			tool++
			if e.ToolInput == nil || *e.ToolInput != `{"path":"marker.txt"}` || e.ToolOutput == nil || !containsSubstring(*e.ToolOutput, "found marker") {
				t.Fatalf("tool input/result were not normalized: %+v", e)
			}
		}
	}
	if assistant != 2 || tool != 1 {
		t.Fatalf("assistant=%d successful read=%d events=%+v", assistant, tool, events)
	}
	var userText, assistantText string
	for _, event := range events {
		if event.MessageID == nil {
			continue
		}
		if *event.MessageID == "u" && event.PromptContent != nil {
			userText = *event.PromptContent
		}
		if *event.MessageID == "a2" && event.PromptContent != nil {
			assistantText = *event.PromptContent
		}
	}
	if userText != `"Read marker.txt and report its contents."` || assistantText != "AL-58 deterministic runtime reply." {
		t.Fatalf("runtime message text was not preserved exactly: user=%q assistant=%q", userText, assistantText)
	}
}

func TestSchemaV2ExtractsSkillAndTaskFactsWithExplicitChildSession(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	if _, err := db.Exec(`INSERT INTO session_v2 VALUES('ses_child','ses_root','coder','mock-model','child',1793231971000,1793231971000)`); err != nil {
		t.Fatal(err)
	}
	payload := `{"time":{"created":1793231970238,"completed":1793231970358},"agent":"build","model":{"id":"mock-model"},"content":[{"type":"tool","id":"skill-call","name":"skill","state":{"status":"completed","input":{"name":"agentic-lens-dev"},"content":[{"type":"text","text":"loaded"}],"time":{"created":1793231970288,"completed":1793231970300}}},{"type":"tool","id":"task-call","name":"task","state":{"status":"completed","input":{"subagent_type":"coder","description":"Implement AL-51"},"content":[{"type":"text","text":"{\"session_id\":\"ses_child\"}"}],"time":{"created":1793231970301,"completed":1793231970345}}}],"cost":0,"tokens":{"input":23,"output":11}}`
	if _, err := db.Exec(`UPDATE session_message SET data=? WHERE id='a1'`, payload); err != nil {
		t.Fatal(err)
	}
	sessions, _, _, err := opencode.ExtractSessions(db, []string{"ses_root"}, "tester", nil)
	if err != nil {
		t.Fatal(err)
	}
	skills, tasks := opencode.ExtractSkillTaskFacts(sessions)
	if len(skills) != 1 || skills[0].Name != "agentic-lens-dev" || skills[0].TurnIndex != 1 || skills[0].MessageIndex != 1 || skills[0].CallerRole != "build" || !skills[0].Success {
		t.Fatalf("unexpected skill facts: %+v sessions=%+v", skills, sessions)
	}
	if len(tasks) != 1 || tasks[0].SubagentType != "coder" || tasks[0].Description != "Implement AL-51" || tasks[0].ParentTurn != 1 || tasks[0].ChildSession != "ses_child" {
		t.Fatalf("unexpected task facts: %+v", tasks)
	}
}

func TestSchemaV2MalformedOrIncompleteSkillTaskInputsPreserveOtherFacts(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	payload := `{"time":{"created":1793231970238,"completed":1793231970358},"agent":"build","model":{"id":"mock-model"},"content":[{"type":"tool","id":"bad-skill","name":"skill","state":{"status":"completed","input":"{bad"}},{"type":"tool","id":"missing-name","name":"skill","state":{"status":"completed","input":{}}},{"type":"tool","id":"partial-task","name":"task","state":{"status":"completed","input":{"subagent_type":"coder"}}}],"cost":0,"tokens":{"input":23,"output":11}}`
	if _, err := db.Exec(`UPDATE session_message SET data=? WHERE id='a1'`, payload); err != nil {
		t.Fatal(err)
	}
	sessions, _, _, err := opencode.ExtractSessions(db, []string{"ses_root"}, "tester", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || len(sessions[0].Invocations) == 0 || len(sessions[0].Events) < 4 {
		t.Fatalf("optional input handling lost session, invocation, or event facts: %+v", sessions)
	}
	skills, tasks := opencode.ExtractSkillTaskFacts(sessions)
	if len(skills) != 0 {
		t.Errorf("malformed or unnamed skill input should be skipped, got %+v", skills)
	}
	if len(tasks) != 1 || tasks[0].SubagentType != "coder" || tasks[0].Description != "" {
		t.Errorf("expected task fact to preserve only available fields, got %+v", tasks)
	}
}

func TestSchemaV2SessionEmptySessionProducesEmptyFacts(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	if _, err := db.Exec(`INSERT INTO session_v2 VALUES('empty',NULL,'build','mock-model','AL-58 empty',1793231971000,1793231971000)`); err != nil {
		t.Fatal(err)
	}
	invs, tools, events, start, end, err := opencode.ExtractFacts(db, []string{"empty"}, "tester", nil)
	if err != nil || len(invs) != 1 || len(tools) != 0 || len(events) != 0 || start == nil || end == nil {
		t.Fatalf("empty session facts inv=%+v tools=%+v events=%+v bounds=%v/%v err=%v", invs, tools, events, start, end, err)
	}
}

func TestSchemaV2SessionMalformedJSONAndUnknownMessageVariantFailWithContext(t *testing.T) {
	for _, tc := range []struct {
		name, kind, data, want string
	}{
		{"malformed_json", "user", "{", "OpenCode schema family v2: decoding user message bad in session ses_root"},
		{"unknown_variant", "system", `{"text":"unexpected"}`, `OpenCode schema family v2: unsupported message type "system"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			createSchemaV2TestDB(t, db)
			if _, err := db.Exec(`INSERT INTO session_message VALUES('bad','ses_root',?,20,1793231970500,1793231970500,?)`, tc.kind, tc.data); err != nil {
				t.Fatal(err)
			}
			_, _, _, _, _, err = opencode.ExtractFacts(db, []string{"ses_root"}, "", nil)
			if err == nil || !containsSubstring(err.Error(), tc.want) {
				t.Fatalf("expected contextual %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestSchemaV2SessionToolPendingRunningAndUnknownStatusesFailClosed(t *testing.T) {
	for _, status := range []string{"pending", "running", "mystery"} {
		t.Run(status, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			createSchemaV2TestDB(t, db)
			payload := `{"time":{"created":1793231970238},"model":{"id":"mock-model"},"content":[{"type":"tool","id":"call1","name":"read","state":{"status":"` + status + `","input":{"path":"marker.txt"}}}]}`
			if _, err := db.Exec(`UPDATE session_message SET data=? WHERE id='a1'`, payload); err != nil {
				t.Fatal(err)
			}
			_, _, _, _, _, err = opencode.ExtractFacts(db, []string{"ses_root"}, "", nil)
			if err == nil || !containsSubstring(err.Error(), `OpenCode schema family v2: unsupported tool status "`+status+`"`) {
				t.Fatalf("status %q should not claim success, got %v", status, err)
			}
		})
	}
}

func TestSchemaV2SessionTiedTimestampsPreserveSourceAndSameMessageOrder(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	const tie = int64(1793231970238)
	if _, err := db.Exec(`UPDATE session_message SET time_created=?,data=CASE id WHEN 'u' THEN '{"time":{"created":1793231970238},"text":"prompt"}' WHEN 'a1' THEN '{"time":{"created":1793231970238,"completed":1793231970238},"model":{"id":"mock-model"},"content":[{"type":"tool","id":"call1","name":"read","state":{"status":"completed","input":{},"content":[]}}]}' ELSE data END WHERE id IN ('u','a1')`, tie); err != nil {
		t.Fatal(err)
	}
	_, _, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_root"}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var tied []string
	for _, event := range events {
		if event.MessageID != nil && (*event.MessageID == "u" || *event.MessageID == "a1") {
			messageID := ""
			if event.MessageID != nil {
				messageID = *event.MessageID
			}
			tied = append(tied, string(event.EventType)+":"+messageID)
		}
	}
	want := []string{"turn_message:u", "turn_message:a1", "tool_call:a1"}
	if len(tied) != len(want) {
		t.Fatalf("same-time source order not preserved: %v", tied)
	}
	for i := range want {
		if tied[i] != want[i] {
			t.Fatalf("same-time source order=%v, want %v", tied, want)
		}
	}
}

func TestSchemaV2CostEvidenceCoalescesReportedAndMissingRecords(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	if _, err := db.Exec(`UPDATE session_message SET data=json_set(data,'$.cost',2) WHERE id='a1'; UPDATE session_message SET data=json_remove(data,'$.cost') WHERE id='a2';`); err != nil {
		t.Fatal(err)
	}
	invs, _, _, _, _, err := opencode.ExtractFacts(db, []string{"ses_root"}, "tester", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(invs) != 1 || invs[0].CostEvidence.State != domain.CostPartial || invs[0].CostEvidence.KnownSubtotalUSD == nil || *invs[0].CostEvidence.KnownSubtotalUSD != 2 {
		t.Fatalf("coalesced cost=%+v", invs)
	}
	summary := domain.ComputeTotals(invs).CostSummary
	if summary.CompleteAmount() != nil || summary.Partial != 1 || summary.Total != 1 {
		t.Fatalf("partial invocation counted as complete: %+v", summary)
	}
	for _, tc := range []struct {
		name, first, second string
		state               domain.CostState
		invalid             bool
	}{
		{"explicit zero remains observed", "0", "0", domain.CostObserved, false},
		{"known zero plus null is partial", "0", "null", domain.CostPartial, false},
		{"only null is unavailable", "null", "null", domain.CostUnavailable, false},
		{"negative cost is rejected", "-1", "null", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.Exec(`UPDATE session_message SET data=json_set(data,'$.cost',json(?)) WHERE id=?`, tc.first, "a1"); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE session_message SET data=json_set(data,'$.cost',json(?)) WHERE id=?`, tc.second, "a2"); err != nil {
				t.Fatal(err)
			}
			rows, _, _, _, _, err := opencode.ExtractFacts(db, []string{"ses_root"}, "tester", nil)
			if tc.invalid {
				if err == nil {
					t.Fatalf("invalid source amount was accepted: %+v", rows)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].CostEvidence.State != tc.state {
				t.Fatalf("source availability=%+v", rows)
			}
			if tc.state == domain.CostUnavailable {
				if rows[0].CostEvidence.KnownSubtotalUSD != nil {
					t.Fatalf("unknown source became known zero: %+v", rows[0])
				}
			} else if rows[0].CostEvidence.KnownSubtotalUSD == nil || *rows[0].CostEvidence.KnownSubtotalUSD != 0 {
				t.Fatalf("known zero contribution lost: %+v", rows[0])
			}
		})
	}

}
