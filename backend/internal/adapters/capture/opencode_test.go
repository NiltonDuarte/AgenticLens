package capture_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestOpenCodeRegistrationNeverFallsBackToMatchingTitle(t *testing.T) {
	for _, family := range []string{"v1", "v2"} {
		t.Run(family, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "source.db")
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			schema := `CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
INSERT INTO session VALUES('unrelated',NULL,'reviewer','source-model','TEST-100 review',1000,2000);`
			if family == "v2" {
				schema = `CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);
INSERT INTO session_v2 VALUES('unrelated',NULL,'reviewer','source-model','TEST-100 review',1000,2000);`
			}
			if _, err := db.Exec(schema); err != nil {
				t.Fatal(err)
			}
			s := openStore(t, t.TempDir())
			id := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: "missing-exact-session"}
			registration, err := s.Register(context.Background(), id, "", "session.created")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Bind(context.Background(), id, "TEST-100", registration.Request); err != nil {
				t.Fatal(err)
			}
			result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Sessions) != 0 || len(result.Registrations) != 1 || result.Registrations[0].Identity.SessionID != "missing-exact-session" || result.Registrations[0].State == domain.CaptureCurrent {
				t.Fatalf("missing exact ID became a matched title/zero success: %+v", result)
			}
		})
	}
}

func mixedCaptureDB(t *testing.T) (string, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "coexisting-families.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	_, err = db.Exec(`CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);`)
	if err != nil {
		t.Fatal(err)
	}
	return path, db
}

func seedCaptureFamily(t *testing.T, db *sql.DB, family, id string) {
	t.Helper()
	if family == "v1" {
		if _, err := db.Exec(`INSERT INTO session VALUES(?,NULL,'reviewer','source-v1','TEST-100 review',1000,3000)`, id); err != nil {
			t.Fatal(err)
		}
		data := `{"role":"assistant","modelID":"source-v1","time":{"created":1500,"completed":2500},"tokens":{"input":11,"output":2}}`
		if _, err := db.Exec(`INSERT INTO message VALUES(?,?,1500,2500,?)`, "message-"+id, id, data); err != nil {
			t.Fatal(err)
		}
		return
	}
	if _, err := db.Exec(`INSERT INTO session_v2 VALUES(?,NULL,'reviewer','source-v2','TEST-100 review',1000,3000)`, id); err != nil {
		t.Fatal(err)
	}
	data := `{"agent":"reviewer","model":{"id":"source-v2"},"time":{"created":1500,"completed":2500},"content":[{"type":"text","text":"fixture reply"}],"tokens":{"input":23,"output":3}}`
	if _, err := db.Exec(`INSERT INTO session_message VALUES(?,?,'assistant',1,1500,2500,?)`, "message-"+id, id, data); err != nil {
		t.Fatal(err)
	}
}

func bindOpenCodeCapture(t *testing.T, s *capture.Store, path, id string) {
	t.Helper()
	identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: id}
	registration, err := s.Register(context.Background(), identity, "", "session.created")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Bind(context.Background(), identity, "TEST-100", registration.Request); err != nil {
		t.Fatal(err)
	}
}

func TestOpenCodeCaptureSelectsExactMembershipInCoexistingFamilies(t *testing.T) {
	for _, tc := range []struct {
		name   string
		v1, v2 bool
		model  string
		input  int64
	}{
		{"v1 exact owner", true, false, "source-v1", 11},
		{"v2 exact owner", false, true, "source-v2", 23},
		{"duplicate exact owner is ambiguous", true, true, "", 0},
		{"missing exact owner stays unresolved", false, false, "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, db := mixedCaptureDB(t)
			// Both families have the same ticket title and contain usable source facts.
			// Those ambient matches cannot select the registered owner's family.
			seedCaptureFamily(t, db, "v1", "unrelated-v1")
			seedCaptureFamily(t, db, "v2", "unrelated-v2")
			if tc.v1 {
				seedCaptureFamily(t, db, "v1", "exact-owner")
			}
			if tc.v2 {
				seedCaptureFamily(t, db, "v2", "exact-owner")
			}
			s := openStore(t, t.TempDir())
			bindOpenCodeCapture(t, s, path, "exact-owner")
			result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Registrations) != 1 {
				t.Fatalf("lost exact registration: %+v", result)
			}
			if tc.model == "" {
				if len(result.Sessions) != 0 || result.Registrations[0].State == domain.CaptureCurrent {
					t.Fatalf("ambiguous/missing exact owner became complete source facts: %+v", result)
				}
				if tc.v1 && tc.v2 && result.Registrations[0].State != domain.CaptureFailed {
					t.Fatalf("ambiguous duplicate not reported failed: %+v", result.Registrations[0])
				}
				return
			}
			if result.Registrations[0].State != domain.CaptureCurrent || len(result.Sessions) != 1 || result.Sessions[0].ID != "exact-owner" || len(result.Sessions[0].Invocations) != 1 {
				t.Fatalf("exact owner not reconciled: %+v", result)
			}
			invocation := result.Sessions[0].Invocations[0]
			if invocation.Model != tc.model || invocation.Tokens.Input != tc.input {
				t.Fatalf("selected another family's source facts: %+v", invocation)
			}
		})
	}
}

func TestOpenCodeCapturePreservesSnapshotWhenCoexistingOwnerBecomesAmbiguousOrMissing(t *testing.T) {
	for _, disposition := range []string{"ambiguous", "missing"} {
		t.Run(disposition, func(t *testing.T) {
			path, db := mixedCaptureDB(t)
			seedCaptureFamily(t, db, "v1", "exact-owner")
			seedCaptureFamily(t, db, "v2", "foreign-owner")
			s := openStore(t, t.TempDir())
			bindOpenCodeCapture(t, s, path, "exact-owner")
			before, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget)
			if err != nil {
				t.Fatal(err)
			}
			if len(before.Sessions) != 1 || before.Registrations[0].State != domain.CaptureCurrent {
				t.Fatalf("fixture did not establish an accepted source snapshot: %+v", before)
			}
			if disposition == "ambiguous" {
				seedCaptureFamily(t, db, "v2", "exact-owner")
			} else if _, err := db.Exec(`DELETE FROM session WHERE id='exact-owner'`); err != nil {
				t.Fatal(err)
			}
			after, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before.Sessions, after.Sessions) || after.Registrations[0].State == domain.CaptureCurrent {
				t.Fatalf("unresolved family ownership erased accepted source facts or claimed current: %+v", after)
			}
		})
	}
}

func TestOpenCodeChildDiscoveryContinuesPastAlreadyKnownPrefix(t *testing.T) {
	path, db := mixedCaptureDB(t)
	seedCaptureFamily(t, db, "v1", "root")
	seedCaptureFamily(t, db, "v1", "child-a")
	if _, err := db.Exec(`UPDATE session SET parent_id='root' WHERE id='child-a'`); err != nil {
		t.Fatal(err)
	}
	s := openStore(t, t.TempDir())
	bindOpenCodeCapture(t, s, path, "root")
	if _, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget); err != nil {
		t.Fatal(err)
	}
	seedCaptureFamily(t, db, "v1", "child-z")
	if _, err := db.Exec(`UPDATE session SET parent_id='root' WHERE id='child-z'`); err != nil {
		t.Fatal(err)
	}
	// Locally observed callbacks remain unassigned until native parent proof.
	for _, id := range []string{"child-z", "unrelated-root"} {
		if id == "unrelated-root" {
			seedCaptureFamily(t, db, "v1", id)
		}
		if _, err := s.Register(context.Background(), domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: id}, "", "session.created"); err != nil {
			t.Fatal(err)
		}
	}
	recovered := false
	for pass := 0; pass < 12; pass++ {
		result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 1, capture.DefaultByteBudget)
		if err != nil {
			t.Fatal(err)
		}
		if result.SourceAttempts > 1 {
			t.Fatalf("bounded discovery exceeded shared session allowance: %+v", result)
		}
		local, err := s.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range local {
			if r.Identity.SessionID == "unrelated-root" && r.TicketID != "" {
				t.Fatalf("discovery guessed unrelated root ownership: %+v", r)
			}
			if r.Identity.SessionID == "child-z" && r.TicketID == "TEST-100" && len(r.Snapshot) == 1 {
				recovered = true
			}
		}
		if recovered {
			break
		}
	}
	if !recovered {
		t.Fatal("repeated bounded passes never progressed past an already-known child prefix")
	}
}
