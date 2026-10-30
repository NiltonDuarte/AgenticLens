package opencode_test

import (
	"database/sql"
	"path/filepath"
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
