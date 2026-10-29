package opencode_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func createV2DB(t *testing.T, db *sql.DB) {
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

func TestV2RuntimeEvidenceExtractsNormalizedFacts(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createV2DB(t, db)
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

func TestOpenDBRemainsReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	createV2DB(t, db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	client, err := opencode.OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.DB().Exec(`UPDATE session_v2 SET title='changed' WHERE id='ses_root'`); err == nil {
		t.Fatal("read-only OpenCode connection accepted a write")
	}
}

func TestV1FilePathAndV2PathNormalizeToSameToolInput(t *testing.T) {
	inputs := []string{`{"filePath":"marker.txt"}`, `{"path":"marker.txt"}`}
	for _, raw := range inputs {
		var normalized *string
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);INSERT INTO session VALUES('s',NULL,'build','{}','AL-58',1000,2000);INSERT INTO message VALUES('m','s',1000,1100,'{"role":"assistant","modelID":"mock-model","time":{"created":1000,"completed":1100}}');`)
		if err != nil {
			t.Fatal(err)
		}
		toolData := `{"type":"tool","tool":"read","state":{"status":"completed","input":` + raw + `,"output":"found marker"}}`
		if _, err = db.Exec(`INSERT INTO part VALUES('p','s','m',1050,1080,?)`, toolData); err != nil {
			t.Fatal(err)
		}
		_, _, events, _, _, err := opencode.ExtractFacts(db, []string{"s"}, "user", nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.EventType == domain.EventTypeToolCall {
				normalized = event.ToolInput
			}
		}
		if normalized == nil || *normalized != `{"path":"marker.txt"}` {
			t.Fatalf("input %s normalized to %v", raw, normalized)
		}
		db.Close()
	}
}

func TestSchemaFamilyRejectsPartialAndMixedLayouts(t *testing.T) {
	cases := []struct {
		name, sql string
		want      string
	}{
		{"unknown", "CREATE TABLE nope(id TEXT);", "unsupported OpenCode database schema"},
		{"partial", "CREATE TABLE session(id TEXT);", "unsupported OpenCode database schema"},
		{"mixed", "CREATE TABLE session(id TEXT,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER); CREATE TABLE message(id TEXT,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT); CREATE TABLE part(id TEXT,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT); CREATE TABLE session_v2(id TEXT,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);", "unsupported mixed OpenCode schema"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err = db.Exec(tc.sql); err != nil {
				t.Fatal(err)
			}
			_, _, _, _, _, err = opencode.ExtractFacts(db, []string{"ses"}, "", nil)
			if err == nil || !containsSubstring(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestSchemaFamilyRejectsEveryRequiredTableColumnOmission(t *testing.T) {
	// Keep the table inventories complete, then remove each required column in
	// turn so schema validation cannot accidentally depend on only table names.
	tables := map[string][]string{
		"session":         {"id", "parent_id", "agent", "model", "title", "time_created", "time_updated"},
		"message":         {"id", "session_id", "time_created", "time_updated", "data"},
		"part":            {"id", "session_id", "message_id", "time_created", "time_updated", "data"},
		"session_v2":      {"id", "parent_id", "agent", "model", "title", "time_created", "time_updated"},
		"session_message": {"id", "session_id", "type", "seq", "time_created", "time_updated", "data"},
	}
	for table, columns := range tables {
		for _, omitted := range columns {
			t.Run(table+"_missing_"+omitted, func(t *testing.T) {
				db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				for name, cols := range tables {
					if name == table {
						filtered := make([]string, 0, len(cols)-1)
						for _, col := range cols {
							if col != omitted {
								filtered = append(filtered, col)
							}
						}
						cols = filtered
					}
					defs := make([]string, 0, len(cols))
					for _, col := range cols {
						defs = append(defs, col+" TEXT")
					}
					if _, err := db.Exec("CREATE TABLE \"" + name + "\" (" + strings.Join(defs, ",") + ")"); err != nil {
						t.Fatal(err)
					}
				}
				_, _, _, _, _, err = opencode.ExtractFacts(db, []string{"s"}, "", nil)
				if err == nil || !containsSubstring(err.Error(), "unsupported mixed OpenCode schema") {
					t.Fatalf("omitting %s.%s should reject competing/incomplete families, got %v", table, omitted, err)
				}
			})
		}
	}
}

func TestV2MalformedJSONAndUnknownMessageVariantFailWithContext(t *testing.T) {
	for _, tc := range []struct {
		name, kind, data, want string
	}{
		{"malformed_json", "user", "{", "decoding v2 user message bad in session ses_root"},
		{"unknown_variant", "system", `{"text":"unexpected"}`, `unsupported v2 message type "system"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			createV2DB(t, db)
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

func TestV2ToolPendingRunningAndUnknownStatusesFailClosed(t *testing.T) {
	for _, status := range []string{"pending", "running", "mystery"} {
		t.Run(status, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			createV2DB(t, db)
			payload := `{"time":{"created":1793231970238},"model":{"id":"mock-model"},"content":[{"type":"tool","id":"call1","name":"read","state":{"status":"` + status + `","input":{"path":"marker.txt"}}}]}`
			if _, err := db.Exec(`UPDATE session_message SET data=? WHERE id='a1'`, payload); err != nil {
				t.Fatal(err)
			}
			_, _, _, _, _, err = opencode.ExtractFacts(db, []string{"ses_root"}, "", nil)
			if err == nil || !containsSubstring(err.Error(), `unsupported v2 tool status "`+status+`"`) {
				t.Fatalf("status %q should not claim success, got %v", status, err)
			}
		})
	}
}

func TestV2TiedTimestampsPreserveSourceAndSameMessageOrder(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createV2DB(t, db)
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

func TestV2EmptySessionProducesEmptyFactsAndRootResolutionIsExplicit(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createV2DB(t, db)
	if _, err := db.Exec(`INSERT INTO session_v2 VALUES('empty',NULL,'build','mock-model','AL-58 empty',1793231971000,1793231971000)`); err != nil {
		t.Fatal(err)
	}
	root, err := opencode.ResolveActiveRootSession(db, "", "", "empty")
	if err != nil || root == nil || *root != "empty" {
		t.Fatalf("explicit v2 root=%v err=%v", root, err)
	}
	invs, tools, events, start, end, err := opencode.ExtractFacts(db, []string{"empty"}, "tester", nil)
	if err != nil || len(invs) != 1 || len(tools) != 0 || len(events) != 0 || start == nil || end == nil {
		t.Fatalf("empty session facts inv=%+v tools=%+v events=%+v bounds=%v/%v err=%v", invs, tools, events, start, end, err)
	}
}
