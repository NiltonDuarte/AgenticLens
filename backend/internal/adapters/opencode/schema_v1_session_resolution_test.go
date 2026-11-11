package opencode_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"agenticlens/internal/adapters/opencode"
)

func resolutionDB(t *testing.T, family string) (*sql.DB, string) {
	t.Helper()
	if family == "v1" {
		db := setupTestDB(t)
		db.SetMaxOpenConns(1)
		t.Cleanup(func() {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		})
		return db, "session"
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	createSchemaV2TestDB(t, db)
	if _, err := db.Exec(`DELETE FROM session_message; DELETE FROM session_v2`); err != nil {
		t.Fatal(err)
	}
	return db, "session_v2"
}

func resolutionSession(t *testing.T, db *sql.DB, table, id, parent, title string, stamp int64) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO `+table+`(id,parent_id,agent,model,title,time_created,time_updated) VALUES(?,?,'coder','model',?,?,?)`, id, sql.NullString{String: parent, Valid: parent != ""}, title, stamp, stamp); err != nil {
		t.Fatal(err)
	}
}

func resolutionRecord(t *testing.T, db *sql.DB, family, id, session, data string) {
	t.Helper()
	var err error
	if family == "v1" {
		_, err = db.Exec(`INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES(?,?,?,1500,1500,?)`, id, session, id, data)
	} else {
		_, err = db.Exec(`INSERT INTO session_message(id,session_id,type,seq,time_created,time_updated,data) VALUES(?,?,'user',1,1500,1500,?)`, id, session, data)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func exerciseExplicitRootSelectors(t *testing.T, family string) {
	t.Helper()
	db, table := resolutionDB(t, family)
	resolutionSession(t, db, table, "root_a", "", "CASE-42 matching title", 1000)
	resolutionSession(t, db, table, "child_a", "root_a", "", 1100)
	resolutionSession(t, db, table, "grand_a", "child_a", "", 1200)
	resolutionSession(t, db, table, "root_b", "", "CASE-42 newer matching title", 2000)
	resolutionRecord(t, db, family, "a1", "root_a", `{"text":"literal%_token"}`)
	resolutionRecord(t, db, family, "a2", "grand_a", `{"text":"literal%_token repeated same root"}`)
	resolutionRecord(t, db, family, "a3", "grand_a", `{"path":"custom-tickets/CASE-42/ledger.json","text":"ambiguous-token"}`)
	resolutionRecord(t, db, family, "b1", "root_b", `{"text":"literalXXtoken ambiguous-token other-root-token"}`)
	for _, tc := range []struct {
		name, nonce, explicit, want string
		ambiguous                   bool
	}{
		{name: "existing exact root", explicit: "root_a", want: "root_a"},
		{name: "literal wildcard token groups descendant matches by root", nonce: "literal%_token", want: "root_a"},
		{name: "both selectors agree", nonce: "literal%_token", explicit: "root_a", want: "root_a"},
		{name: "ticket title and configured path cannot select"},
		{name: "unmatched nonce never uses title", nonce: "absent-token"},
		{name: "unmatched explicit never uses title", explicit: "missing"},
		{name: "child cannot be supplied as root", explicit: "child_a"},
		{name: "invalid explicit selector", explicit: "bad/id"},
		{name: "unmatched nonce cannot use valid explicit", nonce: "absent-token", explicit: "root_a"},
		{name: "invalid explicit cannot use valid nonce", nonce: "literal%_token", explicit: "missing"},
		{name: "selectors disagree", nonce: "other-root-token", explicit: "root_a"},
		{name: "ambiguous nonce never chooses newest", nonce: "ambiguous-token", ambiguous: true},
		{name: "explicit root cannot disambiguate nonce", nonce: "ambiguous-token", explicit: "root_a", ambiguous: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := opencode.ResolveActiveRootSession(db, tc.nonce, "CASE-42", tc.explicit)
			if tc.want != "" {
				if err != nil || got == nil || *got != tc.want {
					t.Fatalf("root=%v error=%v; want %q", got, err, tc.want)
				}
			} else if err == nil || got != nil {
				t.Fatalf("selector failure returned output: %v, %v", got, err)
			} else if tc.ambiguous && !errors.Is(err, opencode.ErrSessionAmbiguous) {
				t.Fatalf("ambiguous root not reported: %v", err)
			}
		})
	}
	if family == "v1" {
		if _, err := db.Exec(`INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES('message_nonce','grand_a',1000,1000,'{"text":"message-only-token"}')`); err != nil {
			t.Fatal(err)
		}
		if got, err := opencode.ResolveActiveRootSession(db, "message-only-token", "", ""); err != nil || got == nil || *got != "root_a" {
			t.Fatalf("native message nonce must resolve its root: %v, %v", got, err)
		}
	}
}

func TestSchemaV1ExplicitRootSelectors(t *testing.T) {
	exerciseExplicitRootSelectors(t, "v1")
}

func TestRootSelectorRequiresSupportedSelectedSchema(t *testing.T) {
	if _, err := opencode.ResolveActiveRootSession(nil, "token", "CASE-42", ""); err == nil {
		t.Fatal("nil database accepted")
	}
	db, table := resolutionDB(t, "v1")
	resolutionSession(t, db, table, "root", "", "CASE-42", 1000)
	if _, err := db.Exec(`DROP TABLE part`); err != nil {
		t.Fatal(err)
	}
	if root, err := opencode.ResolveActiveRootSession(db, "", "CASE-42", "root"); err == nil || root != nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("incomplete selected schema used explicit/title fallback: %v, %v", root, err)
	}
}

func exerciseExplicitRecovery(t *testing.T, family string) {
	t.Helper()
	db, table := resolutionDB(t, family)
	resolutionSession(t, db, table, "root_a", "", "CASE-%_42 work", 1000)
	resolutionSession(t, db, table, "child_a", "root_a", "", 1100)
	resolutionSession(t, db, table, "root_b", "", "CASE-XX42 unrelated wildcard title", 2000)
	resolutionRecord(t, db, family, "a1", "child_a", `{"path":"custom-tickets/PATH-%_42/ledger.json"}`)
	resolutionRecord(t, db, family, "b1", "root_b", `{"path":"workspaces/PATH-%_42/ledger.json"}`)
	for _, tc := range []struct {
		name, ticket, workspace, want string
		min, max                      int64
	}{
		{name: "literal title", ticket: "CASE-%_42", workspace: "custom-tickets/CASE-%_42", want: "root_a"},
		{name: "configured child workspace", ticket: "PATH-%_42", workspace: "custom-tickets/PATH-%_42", want: "root_a"},
		{name: "no invented conventional root", ticket: "PATH-%_42", workspace: "other-tickets/PATH-%_42"},
		{name: "time window excludes match", ticket: "PATH-%_42", workspace: "custom-tickets/PATH-%_42", min: 3000, max: 4000},
		{name: "time window includes match", ticket: "PATH-%_42", workspace: "custom-tickets/PATH-%_42", min: 500, max: 1500, want: "root_a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, err := opencode.DiscoverRootSessionForRecovery(db, tc.ticket, []string{tc.workspace}, tc.min, tc.max)
			if err != nil {
				t.Fatal(err)
			}
			if tc.want == "" && root != nil || tc.want != "" && (root == nil || *root != tc.want) {
				t.Fatalf("recovery root=%v; want %q", root, tc.want)
			}
		})
	}
	if _, err := opencode.DiscoverRootSessionForRecovery(db, "CASE-%_42", nil, 0, 0); err == nil {
		t.Fatal("title recovery accepted missing configured roots")
	}
	resolutionRecord(t, db, family, "b2", "root_b", `{"path":"custom-tickets/CASE-%_42/ledger.json"}`)
	if root, err := opencode.DiscoverRootSessionForRecovery(db, "CASE-%_42", []string{"custom-tickets/CASE-%_42"}, 0, 0); root != nil || !errors.Is(err, opencode.ErrSessionAmbiguous) {
		t.Fatalf("conflicting title/path evidence chose a root: %v, %v", root, err)
	}
	resolutionRecord(t, db, family, "neighbor", "root_a", `{"path":"custom-tickets/NEIGHBOR-420/ledger.json"}`)
	if root, err := opencode.DiscoverRootSessionForRecovery(db, "NEIGHBOR-42", []string{"custom-tickets/NEIGHBOR-42"}, 0, 0); err != nil || root != nil {
		t.Fatalf("neighbor directory matched literal ticket workspace: %v, %v", root, err)
	}
}

func TestSchemaV1ExplicitRecoveryUsesConfiguredLiteralEvidence(t *testing.T) {
	exerciseExplicitRecovery(t, "v1")
}

func TestCaptureRejectsMixedFamiliesBeforeReadingExactOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO session VALUES('only_v1',NULL,'coder','model','test',1000,2000); CREATE TABLE session_v2(id TEXT,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER); CREATE TABLE session_message(id TEXT,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);`); err != nil {
		t.Fatal(err)
	}
	if sessions, _, _, err := opencode.CaptureSessionSnapshot(context.Background(), db, "only_v1", "tester", 1024); err == nil || len(sessions) != 0 || !strings.Contains(err.Error(), "mixed") {
		t.Fatalf("capture bypassed full-layout validation: %v, %v", sessions, err)
	}
	if events, err := opencode.ExtractSessionSourceEvents(db, "only_v1", "tester"); err == nil || len(events) != 0 || !strings.Contains(err.Error(), "mixed") {
		t.Fatalf("hydration bypassed full-layout validation: %v, %v", events, err)
	}
}

func TestRecoveryValidationErrors(t *testing.T) {
	db, _ := resolutionDB(t, "v1")
	for _, tc := range []struct {
		ticket   string
		paths    []string
		min, max int64
	}{
		{"", []string{"tickets/CASE-1"}, 0, 0},
		{"CASE-1", []string{""}, 0, 0},
		{"CASE-1", []string{"/"}, 0, 0},
		{"CASE-1", []string{"tickets/CASE-1"}, 10, 0},
		{"CASE-1", []string{"tickets/CASE-1"}, 20, 10},
	} {
		t.Run(fmt.Sprintf("%q/%v/%d/%d", tc.ticket, tc.paths, tc.min, tc.max), func(t *testing.T) {
			if _, err := opencode.DiscoverRootSessionForRecovery(db, tc.ticket, tc.paths, tc.min, tc.max); err == nil {
				t.Fatal("invalid recovery inputs accepted")
			}
		})
	}
}
