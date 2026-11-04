package opencode_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

func TestOpenCodeDelegationChildProofUsesKnownNativeParentOutsideSelection(t *testing.T) {
	for _, family := range []string{"v1", "v2"} {
		for _, tc := range []struct {
			name, parent          string
			includeOther, missing bool
			want                  string
		}{
			{name: "contradiction selected", parent: "other-root", includeOther: true},
			{name: "contradiction outside selection", parent: "other-root"},
			{name: "known child is another root", parent: ""},
			{name: "proven direct child", parent: "owner", want: "returned-child"},
			{name: "explicit unavailable child", missing: true, want: "returned-child"},
		} {
			t.Run(family+"/"+tc.name, func(t *testing.T) {
				var db *sql.DB
				var err error
				owner := "owner"
				if family == "v1" {
					db = setupTestDB(t)
					_, err = db.Exec(`INSERT INTO session VALUES('owner',NULL,'coder','fixture','TEST-100 parent',1000,3000);INSERT INTO session VALUES('other-root',NULL,'tester','fixture','TEST-100 other',1000,3000);INSERT INTO message VALUES('message','owner',1500,2500,'{"role":"assistant","modelID":"fixture","time":{"created":1500,"completed":2500}}');INSERT INTO part VALUES('delegation','owner','message',1600,2000,'{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"tester","description":"Run checks"},"output":"{\"session_id\":\"returned-child\"}"}}');`)
				} else {
					db, err = sql.Open("sqlite", ":memory:")
					if err != nil {
						t.Fatal(err)
					}
					createSchemaV2TestDB(t, db)
					owner = "ses_root"
					_, err = db.Exec(`INSERT INTO session_v2 VALUES('other-root',NULL,'tester','fixture','TEST-100 other',1000,3000);UPDATE session_message SET data='{"time":{"created":1793231970238,"completed":1793231970358},"agent":"build","model":{"id":"fixture"},"content":[{"type":"tool","id":"delegation","name":"task","state":{"status":"completed","input":{"subagent_type":"tester","description":"Run checks"},"content":[{"type":"text","text":"{\"session_id\":\"returned-child\"}"}],"time":{"created":1793231970301,"completed":1793231970345}}}],"tokens":{"input":1}}' WHERE id='a1'`)
				}
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				if !tc.missing {
					var parent *string
					if tc.parent != "" {
						value := tc.parent
						if value == "owner" {
							value = owner
						}
						parent = &value
					}
					table := "session"
					if family == "v2" {
						table = "session_v2"
					}
					if _, err = db.Exec(`INSERT INTO `+table+` VALUES('returned-child',?,'tester','fixture','TEST-100 child',1000,3000)`, parent); err != nil {
						t.Fatal(err)
					}
				}
				ids := []string{owner}
				if tc.includeOther {
					ids = append(ids, "other-root")
				}
				selected, _, _, err := opencode.ExtractSessionsStrict(db, ids, "tester", nil)
				if err != nil {
					t.Fatal(err)
				}
				assertDelegationChild(t, saveReloadSourceOwners(t, selected), owner, tc.want)
				sourceEvents, sourceErr := opencode.ExtractSessionSourceEvents(db, owner, "tester")
				if sourceErr != nil {
					t.Fatal(sourceErr)
				}
				found := false
				for _, event := range sourceEvents {
					if event.ToolName != nil && *event.ToolName == "task" {
						found = true
						if event.ChildSessionID != tc.want || event.ToolOutput == nil || !strings.Contains(*event.ToolOutput, "returned-child") {
							t.Fatalf("source-only navigation escaped parent proof or raw result was erased: %+v", event)
						}
					}
				}
				if !found {
					t.Fatal("native task event lost")
				}

				snapshot, _, _, err := opencode.CaptureSessionSnapshot(context.Background(), db, owner, "tester", 1000000)
				if err != nil {
					t.Fatal(err)
				}
				assertDelegationChild(t, saveReloadSourceOwners(t, snapshot), owner, tc.want)
				// Registry reconciliation must apply the same source proof before publishing.
				sourcePath := filepath.Join(t.TempDir(), "native.db")
				if _, err = db.Exec(`VACUUM INTO ?`, sourcePath); err != nil {
					t.Fatal(err)
				}
				store, err := capture.Open(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				defer store.Close()
				identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: sourcePath, SessionID: owner}
				row, err := store.Register(context.Background(), identity, "", "session.created")
				if err != nil {
					t.Fatal(err)
				}
				if _, err = store.Bind(context.Background(), identity, "TEST-100", row.Request); err != nil {
					t.Fatal(err)
				}
				result, err := store.Reconcile(context.Background(), "TEST-100", "tester", 20, capture.DefaultByteBudget)
				if err != nil {
					t.Fatal(err)
				}
				assertDelegationChild(t, saveReloadSourceOwners(t, result.Sessions), owner, tc.want)
			})
		}
	}
}

func assertDelegationChild(t *testing.T, sessions []domain.SessionFact, owner, want string) {
	t.Helper()
	for _, session := range sessions {
		if session.ID == owner {
			if len(session.Delegations) != 1 || session.Delegations[0].ChildSession != want {
				t.Fatalf("native parent contradiction lost or explicit unavailable ID removed: owner=%s want child=%q got=%+v", owner, want, session.Delegations)
			}
			return
		}
	}
	t.Fatalf("parent occurrence owner %s disappeared: %+v", owner, sessions)
}
