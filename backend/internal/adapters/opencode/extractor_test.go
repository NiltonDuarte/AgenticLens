package opencode_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
	"reflect"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed opening in-memory db: %v", err)
	}

	schema := `
		CREATE TABLE session (
			id TEXT PRIMARY KEY,
			parent_id TEXT,
			agent TEXT,
			model TEXT,
			title TEXT,
			time_created INTEGER,
			time_updated INTEGER
		);
		CREATE TABLE message (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
		CREATE TABLE part (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			message_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed executing test schema: %v", err)
	}
	return db
}

func TestEmptyThreadSelectionStillValidatesSchema(t *testing.T) {
	unsupportedSchemas := []struct {
		name string
		ddl  string
	}{
		{name: "unknown empty database"},
		{name: "malformed partial family", ddl: "CREATE TABLE session (id TEXT);"},
		{name: "unrelated schema", ddl: "CREATE TABLE unrelated (id TEXT);"},
	}

	for _, schema := range unsupportedSchemas {
		t.Run(schema.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if schema.ddl != "" {
				if _, err := db.Exec(schema.ddl); err != nil {
					t.Fatal(err)
				}
			}

			_, _, _, factsErr := opencode.ExtractSessions(db, nil, "user", nil)
			if factsErr == nil || !strings.Contains(factsErr.Error(), "unsupported OpenCode database schema") {
				t.Errorf("ExtractSessions should reject schema, got %v", factsErr)
			}
			if count, steeringErr := opencode.ExtractSteeringTurns(db, nil); steeringErr == nil || !strings.Contains(steeringErr.Error(), "unsupported OpenCode database schema") {
				t.Errorf("ExtractSteeringTurns should reject schema (count %d), got %v", count, steeringErr)
			}
		})
	}

	t.Run("valid empty v1 selection stays empty", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()
		extractedSessions, startedAt, completedAt, err := opencode.ExtractSessions(db, nil, "user", nil)
		extractedFacts := domain.AgenticFactLedger{Sessions: extractedSessions}
		invocations := extractedFacts.InvocationFacts()
		tools := extractedFacts.ToolFacts()
		events := sourceEventsForSessions(t, db, extractedFacts.Sessions)
		if err != nil || len(invocations) != 0 || len(tools) != 0 || len(events) != 0 || startedAt != nil || completedAt != nil {
			t.Fatalf("unexpected empty facts result: invocations=%d tools=%d events=%d start=%v complete=%v err=%v", len(invocations), len(tools), len(events), startedAt, completedAt, err)
		}
		if count, err := opencode.ExtractSteeringTurns(db, nil); err != nil || count != 0 {
			t.Fatalf("unexpected empty steering result: count=%d err=%v", count, err)
		}
	})
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || filepath.Base(s) != "" && stringsContains(s, substr))
}

func stringsContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestOpenDBRemainsReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	createSchemaV2TestDB(t, db)
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

func TestSourceToolInputsPreservePathVariantsAndAllArguments(t *testing.T) {
	inputs := []string{`{"filePath":"marker.txt"}`, `{"path":"marker.txt"}`, `{ "filePath": "marker.txt", "offset": 4, "limit": 12 }`}
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
		extractedSessions, _, _, err := opencode.ExtractSessions(db, []string{"s"}, "user", nil)
		extractedFacts := domain.AgenticFactLedger{Sessions: extractedSessions}
		events := sourceEventsForSessions(t, db, extractedFacts.Sessions)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.EventType == domain.EventTypeToolCall {
				normalized = event.ToolInput
			}
		}
		if normalized == nil || *normalized != raw {
			t.Fatalf("source input %s lost original fields/bytes: %v", raw, normalized)
		}
		db.Close()
	}
}

func sourceEventsForSessions(t *testing.T, db *sql.DB, sessions []domain.SessionFact) []domain.EventFact {
	t.Helper()
	events := []domain.EventFact{}
	for _, session := range sessions {
		rows, err := opencode.ExtractSessionSourceEvents(db, session.ID, "tester")
		if err != nil {
			t.Fatal(err)
		}
		events = append(events, rows...)
	}
	return domain.SortEvents(events)
}
func assertNativeOccurrenceEvent(t *testing.T, events []domain.EventFact, owner, message, native string) {
	t.Helper()
	for _, event := range events {
		if event.SessionID == owner && event.MessageID != nil && *event.MessageID == message && strings.Contains(event.EventID, native) {
			return
		}
	}
	t.Fatalf("native source event %s/%s/%s lost: %+v", owner, message, native, events)
}

func saveReloadSourceOwners(t *testing.T, sessions []domain.SessionFact) []domain.SessionFact {
	t.Helper()
	version, ticket, title := domain.CurrentSchemaVersion, "TEST-SOURCE", "Source projection"
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := adapters.SaveFactLedger(path, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: sessions}); err != nil {
		t.Fatal(err)
	}
	loaded, err := adapters.LoadExistingFactLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	return loaded.Sessions
}
func assertOwnedInvocationTimes(t *testing.T, sessions []domain.SessionFact, id string, skillTime, delegationTime *string) {
	t.Helper()
	for _, owner := range sessions {
		if owner.ID != id {
			continue
		}
		if len(owner.Skills) != 1 || len(owner.Delegations) != 1 {
			t.Fatalf("source occurrences lost: %+v", owner)
		}
		if !reflect.DeepEqual(owner.Skills[0].InvokedAt, skillTime) || !reflect.DeepEqual(owner.Delegations[0].InvokedAt, delegationTime) {
			t.Fatalf("source-native invocation times changed: skill=%v delegation=%v expected=%v/%v owner=%+v", owner.Skills[0].InvokedAt, owner.Delegations[0].InvokedAt, skillTime, delegationTime, owner)
		}
		return
	}
	t.Fatalf("source owner %s lost", id)
}
