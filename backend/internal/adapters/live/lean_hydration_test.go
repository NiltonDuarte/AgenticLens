package live_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

func TestHydrationMixedSourcesRetainsUsableOwnersAndNativePayloads(t *testing.T) {
	dbPath := createTestDBFile(t)
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = writer.Exec(`INSERT INTO session(id,parent_id,agent,model,title,time_created,time_updated) VALUES('opencode-owner',NULL,'custom-reviewer','fixture','TEST-MIXED',1,1);INSERT INTO message VALUES('message','opencode-owner',1000,1000,'{"role":"assistant","content":"native message","reasoning":"native reasoning"}');INSERT INTO part VALUES('tool','opencode-owner','message',1100,1200,'{"type":"tool","tool":"bash","state":{"status":"completed","input":{"cmd":"native command"},"output":"native output"}}');`)
	if err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer hook.Close()
	home := t.TempDir()
	if err = os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	source := `{"type":"session_meta","payload":{"id":"codex-owner","source":"cli"}}` + "\n" + `{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"native-call","name":"load_skill","arguments":"{\"name\":\"selected\"}"}}` + "\n"
	sourcePath := filepath.Join(home, "sessions", "owner.jsonl")
	if err = os.WriteFile(sourcePath, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(hook, testProjectConfig(), home)
	defer service.Close()
	facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "opencode-owner", Source: domain.SourceOpenCode, User: "tester", Skills: []domain.SkillFact{{Name: "stored-only"}}}, {ID: "codex-owner", Source: domain.SourceCodex, User: "tester", Delegations: []domain.DelegationFact{{ChildSession: "unavailable-child"}}}}}
	before, _ := json.Marshal(facts)
	first, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-MIXED", domain.AgenticMetricsLedger{Facts: facts})
	if err != nil {
		t.Fatal(err)
	}
	if first.Availability != domain.TraceComplete || first.DataSource != domain.DataSourceMixed || len(first.Events) != 3 || !reflect.DeepEqual(first.Coverage.CompleteSessionIDs, []string{"opencode-owner", "codex-owner"}) {
		t.Fatalf("mixed source coverage lost: %+v", first)
	}
	foundSQL, foundCodex := false, false
	for _, event := range first.Events {
		if event.EventID == "opencode:part:tool" {
			foundSQL = event.Role == "custom-reviewer" && event.ToolInput != nil && *event.ToolInput == `{"cmd":"native command"}` && event.ToolOutput != nil && *event.ToolOutput == "native output"
		}
		if event.CallID == "native-call" {
			foundCodex = event.SessionID == "codex-owner" && event.ToolInput != nil && *event.ToolInput == `{"name":"selected"}`
		}
	}
	if !foundSQL || !foundCodex {
		t.Fatalf("source-native payload/identity lost: %+v", first.Events)
	}
	second, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-MIXED", domain.AgenticMetricsLedger{Facts: facts})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(second, first) {
		t.Fatal("unchanged hydration accumulated or changed rows")
	}
	if err = os.Remove(sourcePath); err != nil {
		t.Fatal(err)
	}
	partial, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-MIXED", domain.AgenticMetricsLedger{Facts: facts})
	if err != nil {
		t.Fatal(err)
	}
	if partial.Availability != domain.TracePartial || partial.DataSource != domain.DataSourceLocalSQLite || len(partial.Events) != 2 || !reflect.DeepEqual(partial.Coverage.CompleteSessionIDs, []string{"opencode-owner"}) || len(partial.Coverage.Diagnostics) == 0 {
		t.Fatalf("unavailable Codex owner hid supported SQLite rows: %+v", partial)
	}
	for _, event := range partial.Events {
		if event.SessionID != "opencode-owner" {
			t.Fatalf("lean stored occurrence became fabricated trace: %+v", event)
		}
	}
	after, _ := json.Marshal(facts)
	if string(before) != string(after) {
		t.Fatal("source-only hydration mutated lean input")
	}
}
