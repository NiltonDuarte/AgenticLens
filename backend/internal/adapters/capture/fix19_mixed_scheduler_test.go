package capture_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestFIX19CodexDiscoveryCannotPermanentlyStarveFittingOpenCodeOwner(t *testing.T) {
	for _, variant := range []string{"no remaining bytes", "partial remaining bytes", "oversized first owner"} {
		t.Run(variant, func(t *testing.T) {
			ctx := context.Background()
			s := openStore(t, t.TempDir())
			dbPath := filepath.Join(t.TempDir(), "source.db")
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()
			if _, err := db.Exec(`CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);`); err != nil {
				t.Fatal(err)
			}
			sqlID := "opencode-fitting"
			seedCaptureFamily(t, db, "v1", sqlID)
			// Establish the cursor at the SQL owner before adding Codex. With two
			// owners, continuation then visits Codex followed by SQL regardless
			// of their identity hash ordering.
			bindOpenCodeCapture(t, s, dbPath, sqlID)
			sqlOnly, err := s.Reconcile(ctx, "TEST-100", "tester", 2, fixtureByteBudget, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			home := t.TempDir()
			dir := filepath.Join(home, "sessions")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			codexID := "codex-owner"
			path := filepath.Join(dir, "owner.jsonl")
			writeRollout(t, path, codexID, "", 10)
			registerBound(t, s, home, path, codexID, "TEST-100")
			baseline, err := s.Reconcile(ctx, "TEST-100", "tester", 2, fixtureByteBudget, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			if len(baseline.RefreshedSessions) != 2 {
				t.Fatalf("baseline failed actual source extraction: %+v", baseline)
			}
			var oldSQL domain.SessionFact
			found := false
			for _, owner := range baseline.Sessions {
				if owner.ID == sqlID {
					oldSQL = owner
					found = true
				}
			}
			if !found || len(oldSQL.Invocations) != 1 || oldSQL.Invocations[0].Tokens.Input != 11 {
				t.Fatalf("missing accepted SQL baseline: %+v", oldSQL)
			}
			codexBody := writeRollout(t, path, codexID, "", 20)
			data := `{"role":"assistant","modelID":"source-v1","time":{"created":1500,"completed":2500},"tokens":{"input":37,"output":2}}`
			if _, err := db.Exec(`UPDATE message SET data=?,time_updated=3500 WHERE session_id=?`, data, sqlID); err != nil {
				t.Fatal(err)
			}
			remainder := 0
			if variant == "partial remaining bytes" {
				remainder = len(data) / 2
			}
			allowance := 2*int64(len(codexBody)) + int64(remainder)
			if variant == "oversized first owner" {
				// Codex lacks one byte for complete discovery plus extraction
				// even when first, so it must not pin the continuation cursor.
				allowance = 2*int64(len(codexBody)) - 1
			}
			if sqlOnly.SourceBytes > allowance {
				t.Fatal("SQL owner must fit a full pass allowance")
			}
			progressed := false
			for pass := 0; pass < 4; pass++ {
				result, err := s.Reconcile(ctx, "TEST-100", "tester", 2, allowance, fixtureGenerationPolicy())
				if err != nil {
					t.Fatal(err)
				}
				if result.SourceBytes > allowance || result.SourceAttempts > 2 {
					t.Fatalf("scheduler exceeded shared budget: %+v", result)
				}
				found = false
				for _, owner := range result.Sessions {
					if owner.ID != sqlID {
						continue
					}
					found = true
					if len(owner.Invocations) != 1 {
						t.Fatalf("SQL accepted facts lost: %+v", owner)
					}
					input := owner.Invocations[0].Tokens.Input
					if input == 37 {
						progressed = true
					} else if input != 11 {
						t.Fatalf("budget-starved SQL history corrupted: %+v", owner)
					}
				}
				if !found {
					t.Fatal("budget-starved SQL owner disappeared")
				}
				if progressed {
					break
				}
				pending := false
				for _, row := range result.Registrations {
					if row.Identity.SessionID == sqlID {
						pending = row.State == domain.CapturePending && strings.Contains(strings.ToLower(row.Detail), "budget")
					}
				}
				if !pending {
					t.Fatalf("starved SQL owner lacks explicit pending budget diagnostic: %+v", result.Registrations)
				}
			}
			if !progressed {
				t.Fatal("Codex discovery repeatedly consumed the allowance and the cursor permanently starved a fitting OpenCode owner")
			}
			// Once accepted, unchanged bounded passes retain the current SQL snapshot.
			repeated, err := s.Reconcile(ctx, "TEST-100", "tester", 2, allowance, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			found = false
			for _, owner := range repeated.Sessions {
				if owner.ID == sqlID {
					found = true
					if len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != 37 {
						t.Fatalf("unchanged repetition lost or duplicated accepted SQL facts: %+v", owner)
					}
				}
			}
			if !found {
				t.Fatal("repeated bounded pass dropped SQL owner")
			}
		})
	}
}
