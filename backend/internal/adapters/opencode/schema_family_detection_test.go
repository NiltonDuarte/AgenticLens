package opencode_test

import (
	"database/sql"
	"strings"
	"testing"

	"agenticlens/internal/adapters/opencode"
	_ "modernc.org/sqlite"
)

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
			t.Parallel()
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err = db.Exec(tc.sql); err != nil {
				t.Fatal(err)
			}
			_, _, _, err = opencode.ExtractSessions(db, []string{"ses"}, "", nil)
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
				t.Parallel()
				db, err := sql.Open("sqlite", ":memory:")
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
				_, _, _, err = opencode.ExtractSessions(db, []string{"s"}, "", nil)
				if err == nil || !containsSubstring(err.Error(), "unsupported mixed OpenCode schema") {
					t.Fatalf("omitting %s.%s should reject competing/incomplete families, got %v", table, omitted, err)
				}
			})
		}
	}
}

func TestSchemaFamilyV1AuxiliaryBootstrapTable(t *testing.T) {
	const v1 = `CREATE TABLE session(id TEXT,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
 CREATE TABLE message(id TEXT,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
 CREATE TABLE part(id TEXT,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
 INSERT INTO session VALUES('s',NULL,'coder','m','fixture',1767225600000,1767225604000);
 INSERT INTO message VALUES('a','s',1767225601000,1767225604000,'{"role":"assistant","modelID":"m","time":{"created":1767225601000,"completed":1767225604000}}');`
	const auxiliary = `CREATE TABLE session_message(id TEXT,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);`
	for _, tc := range []struct {
		name, sql string
		reject    bool
	}{
		{"native empty bootstrap", auxiliary, false},
		{"populated auxiliary", auxiliary + `INSERT INTO session_message VALUES('unexpected','s','assistant',1,1767225601000,1767225604000,'{}');`, true},
		{"malformed auxiliary", `CREATE TABLE session_message(id TEXT);`, true},
		{"empty partial owner", auxiliary + `CREATE TABLE session_v2(id TEXT);`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err := db.Exec(v1 + tc.sql); err != nil {
				t.Fatal(err)
			}
			sessions, _, _, err := opencode.ExtractSessions(db, []string{"s"}, "tester", nil)
			if tc.reject {
				if err == nil {
					t.Fatalf("ambiguous/incomplete family accepted: %+v", sessions)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(sessions) != 1 || len(sessions[0].Invocations) != 1 || sessions[0].Invocations[0].ActiveDurationSeconds != 3 {
				t.Fatalf("native v1 bootstrap hid completed duration: %+v", sessions)
			}
		})
	}
}
