package opencode_test

import (
	"agenticlens/internal/adapters/opencode"
	"errors"
	_ "modernc.org/sqlite"
	"strings"
	"testing"
)

func TestResolveActiveRootSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed sessions
	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root_1', NULL, 'build', '{"id":"claude-3-opus"}', 'Task AL-100 feature', 1000, 2000),
		('ses_child_1', 'ses_root_1', 'coder', '{"id":"claude-3-opus"}', NULL, 1100, 1900),
		('ses_root_2', NULL, 'build', '{"id":"claude-3-opus"}', 'Task AL-200 feature', 3000, 4000);
	`)
	if err != nil {
		t.Fatalf("failed inserting test sessions: %v", err)
	}

	// Seed part with nonce in child session
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('part_1', 'ses_child_1', 'msg_1', 1500, 1500, '{"type":"tool","tool":"bash","state":{"input":"nonce_abc_123"}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting test part: %v", err)
	}

	// Test Priority 1: Exact Nonce matching
	t.Run("Priority 1 Nonce", func(t *testing.T) {
		res, err := opencode.ResolveActiveRootSession(db, "nonce_abc_123", "AL-200", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_root_1" {
			t.Fatalf("expected ses_root_1, got %v", res)
		}
	})

	// Test Priority 2: Explicit Session ID
	t.Run("Priority 2 Explicit Session ID", func(t *testing.T) {
		res, err := opencode.ResolveActiveRootSession(db, "", "AL-100", "ses_custom_explicit")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_custom_explicit" {
			t.Fatalf("expected ses_custom_explicit, got %v", res)
		}
	})

	// Test Priority 3: Ticket ID matching
	t.Run("Priority 3 Ticket ID", func(t *testing.T) {
		res, err := opencode.ResolveActiveRootSession(db, "", "AL-200", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_root_2" {
			t.Fatalf("expected ses_root_2, got %v", res)
		}
	})

	// Test that Priority 4 is deleted: fails fast without falling back to latest message
	t.Run("Fails Fast Without Tier 4 Fallback When Neither Nonce Nor Ticket ID Supplied", func(t *testing.T) {
		_, err := db.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('msg_latest', 'ses_child_1', 5000, 5000, '{"role":"assistant"}');
		`)
		if err != nil {
			t.Fatalf("failed inserting message: %v", err)
		}

		res, err := opencode.ResolveActiveRootSession(db, "", "", "")
		if err == nil {
			t.Fatal("expected error when neither nonce nor ticket ID is supplied, got nil")
		}
		if !errors.Is(err, opencode.ErrSessionNotFound) {
			t.Fatalf("expected ErrSessionNotFound, got %v", err)
		}
		if res != nil {
			t.Fatalf("expected nil session ID, got %v", res)
		}
	})

	// Test Priority 3 Fallback: Workspace Path matching when title lookup finds no rows
	t.Run("Priority 3 Workspace Path Fallback", func(t *testing.T) {
		_, err := db.Exec(`
			INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
			('ses_root_fallback', NULL, 'build', '{"id":"claude-3-opus"}', 'Unrelated Title Without Ticket', 6000, 7000),
			('ses_child_fallback', 'ses_root_fallback', 'coder', '{"id":"claude-3-opus"}', NULL, 6100, 6900);
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
			('part_fallback', 'ses_child_fallback', 'msg_fb', 6200, 6200, '{"path":"workspaces/AL-300/agentic_metrics.json"}');
		`)
		if err != nil {
			t.Fatalf("failed inserting fallback test records: %v", err)
		}

		res, err := opencode.ResolveActiveRootSession(db, "", "AL-300", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_root_fallback" {
			t.Fatalf("expected ses_root_fallback, got %v", res)
		}
	})

	t.Run("Unresolved Ticket Returns ErrSessionNotFound", func(t *testing.T) {
		res, err := opencode.ResolveActiveRootSession(db, "", "AL-NON-EXISTENT", "")
		if err == nil {
			t.Fatal("expected error for non-existent ticket, got nil")
		}
		if !errors.Is(err, opencode.ErrSessionNotFound) {
			t.Fatalf("expected ErrSessionNotFound, got %v", err)
		}
		if res != nil {
			t.Fatalf("expected nil for non-existent ticket, got %v", *res)
		}
	})

	t.Run("Tier 4 is dead: foreign session with latest message is never captured when Tiers 1-3 fail", func(t *testing.T) {
		// Insert a foreign session and the latest message in the entire DB
		_, err := db.Exec(`
			INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
			('ses_foreign_root', NULL, 'build', 'default_model', 'Foreign Root Session', 90000, 90000),
			('ses_foreign_child', 'ses_foreign_root', 'coder', 'default_model', NULL, 90001, 90001);
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('msg_foreign_latest', 'ses_foreign_child', 9999999, 9999999, '{"role":"assistant","content":"foreign message"}');
		`)
		if err != nil {
			t.Fatalf("failed inserting foreign session and message: %v", err)
		}

		scenarios := []struct {
			name      string
			nonce     string
			ticketID  string
			sessionID string
		}{
			{name: "unmatched nonce and ticket", nonce: "non-existent-nonce", ticketID: "AL-FOREIGN-GHOST", sessionID: ""},
			{name: "empty nonce and unmatched ticket", nonce: "", ticketID: "AL-FOREIGN-GHOST", sessionID: ""},
			{name: "unmatched nonce and empty ticket", nonce: "non-existent-nonce", ticketID: "", sessionID: ""},
			{name: "all empty parameters", nonce: "", ticketID: "", sessionID: ""},
		}

		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				res, err := opencode.ResolveActiveRootSession(db, sc.nonce, sc.ticketID, sc.sessionID)
				if err == nil {
					t.Fatalf("[%s] expected ErrSessionNotFound, got nil error and session %v", sc.name, res)
				}
				if !errors.Is(err, opencode.ErrSessionNotFound) {
					t.Fatalf("[%s] expected error wrapping ErrSessionNotFound, got: %v", sc.name, err)
				}
				if res != nil {
					t.Fatalf("[%s] expected nil session result, got: %v", sc.name, *res)
				}
			})
		}
	})

	t.Run("Priority 3 Fallback Database Error Handling", func(t *testing.T) {
		errDB := setupTestDB(t)
		defer errDB.Close()

		// Drop part table so ResolveSessionByWorkspacePath will fail with a database error
		if _, err := errDB.Exec("DROP TABLE part;"); err != nil {
			t.Fatalf("failed dropping part table: %v", err)
		}

		res, err := opencode.ResolveActiveRootSession(errDB, "", "AL-FALLBACK-ERR", "")
		if err == nil {
			t.Fatal("expected error when part table is missing during fallback, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil session ID on error, got %v", *res)
		}
		if !strings.Contains(err.Error(), "unsupported OpenCode database schema") {
			t.Fatalf("expected actionable unsupported-schema diagnostic, got %v", err)
		}
	})

	t.Run("Priority 3 Session Query Database Error Handling", func(t *testing.T) {
		errDB := setupTestDB(t)
		defer errDB.Close()

		// Drop session table so initial ticket query fails with a database error
		if _, err := errDB.Exec("DROP TABLE session;"); err != nil {
			t.Fatalf("failed dropping session table: %v", err)
		}

		res, err := opencode.ResolveActiveRootSession(errDB, "", "AL-SESSION-ERR", "")
		if err == nil {
			t.Fatal("expected error when session table is missing, got nil")
		}
		if res != nil {
			t.Fatalf("expected nil session ID on error, got %v", *res)
		}
		if !strings.Contains(err.Error(), "unsupported OpenCode database schema") {
			t.Fatalf("expected actionable unsupported-schema diagnostic, got %v", err)
		}
	})
}

func TestResolveSessionByWorkspacePath(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root_target', NULL, 'orchestrator', 'gpt-4o', 'Feature work', 10000, 11000),
		('ses_root_other', NULL, 'orchestrator', 'gpt-4o', 'Other work', 20000, 21000);

		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_target', 'ses_root_target', 'm1', 10500, 10500, '{"path":"workspaces/PO-3/agentic_metrics.json"}'),
		('p_other', 'ses_root_other', 'm2', 20500, 20500, '{"path":"workspaces/PO-3/agentic_metrics.json"}');
	`)
	if err != nil {
		t.Fatalf("seeding failed: %v", err)
	}

	t.Run("nil DB error", func(t *testing.T) {
		_, err := opencode.ResolveSessionByWorkspacePath(nil, "PO-3", 9000, 12000)
		if err == nil {
			t.Fatal("expected error for nil DB")
		}
	})

	t.Run("empty ticket returns nil", func(t *testing.T) {
		res, err := opencode.ResolveSessionByWorkspacePath(db, "", 9000, 12000)
		if err != nil || res != nil {
			t.Fatalf("expected nil result, got %v, err=%v", res, err)
		}
	})

	t.Run("resolves root session within timestamp window", func(t *testing.T) {
		res, err := opencode.ResolveSessionByWorkspacePath(db, "PO-3", 9000, 12000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_root_target" {
			t.Fatalf("expected ses_root_target, got %v", res)
		}
	})

	t.Run("filters out session outside timestamp window", func(t *testing.T) {
		res, err := opencode.ResolveSessionByWorkspacePath(db, "PO-3", 30000, 40000)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != nil {
			t.Fatalf("expected nil for out of window range, got %v", res)
		}
	})
}
