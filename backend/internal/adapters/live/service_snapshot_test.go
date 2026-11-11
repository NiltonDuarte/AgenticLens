package live

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestSessionTraceSnapshotKeepsPersonaAcrossOtherSessionInvalidation(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "snapshots.db")
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := writer.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := writer.Exec(`PRAGMA journal_mode=WAL;
CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);`); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	for _, id := range []string{"target", "other"} {
		if _, err := writer.Exec(`INSERT INTO session(id,time_created,time_updated) VALUES(?,?,?)`, id, now, now); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Exec(`INSERT INTO message VALUES(?,?,?,?,?)`, id, id, now, now, `{"role":"assistant","agent":"planner","content":"original"}`); err != nil {
			t.Fatal(err)
		}
	}
	hook, err := OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := hook.Close(); err != nil {
			t.Error(err)
		}
	})
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	service := NewLiveService(hook, cfg)
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx := context.Background()
	if _, err := service.GetSessionTree(ctx); err != nil {
		t.Fatal(err)
	}
	accepted, err := service.sessionTraceSnapshot(ctx, "target")
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := service.sessionTraceSnapshot(ctx, "target")
	if err != nil {
		t.Fatal(err)
	}
	if accepted != unchanged || accepted.Persona.Resolve("") != "planner" || len(accepted.Trace.Messages) != 1 {
		t.Fatalf("unchanged trace and persona were not one reusable snapshot: %+v", accepted)
	}

	// Reproduce the interleaving deterministically: another request observes a
	// committed source revision after the tree acquires its target snapshot.
	if _, err := writer.Exec(`UPDATE message SET data='{"role":"assistant","agent":"reviewer","content":"other changed"}',time_updated=? WHERE id='other'`, now+1); err != nil {
		t.Fatal(err)
	}
	other, err := service.sessionTraceSnapshot(ctx, "other")
	if err != nil {
		t.Fatal(err)
	}
	if other.Persona.Resolve("") != "reviewer" || accepted.Persona.Resolve("") != "planner" || accepted.Trace.Messages[0].Content != "original" {
		t.Fatalf("another request invalidated accepted snapshot evidence: target=%+v other=%+v", accepted, other)
	}
	reloaded, err := service.sessionTraceSnapshot(ctx, "target")
	if err != nil {
		t.Fatal(err)
	}
	if reloaded == accepted || reloaded.Persona.Resolve("") != "planner" {
		t.Fatalf("expected a fresh named snapshot after source invalidation: %+v", reloaded)
	}
	if _, err := writer.Exec(`UPDATE message SET data='{"role":"assistant","agent":"coder","content":"target changed"}',time_updated=? WHERE id='target'`, now+2); err != nil {
		t.Fatal(err)
	}
	tree, err := service.GetSessionTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	roles := make(map[string]domain.AgentRole)
	for _, root := range tree.RootSessions {
		roles[root.ID] = root.Role
	}
	if roles["target"] != "coder" || roles["other"] != "reviewer" || accepted.Persona.Resolve("") != "planner" {
		t.Fatalf("tree projection or accepted snapshot lost source persona: roles=%v original=%+v", roles, accepted)
	}
}
