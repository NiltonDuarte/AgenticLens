package opencode_test

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
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
			name        string
			nonce       string
			ticketID    string
			sessionID   string
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
		if !strings.Contains(err.Error(), "resolving session by workspace path fallback") {
			t.Fatalf("expected wrapped fallback error, got %v", err)
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
		if !strings.Contains(err.Error(), "resolving session by ticket id") {
			t.Fatalf("expected wrapped ticket id error, got %v", err)
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

func TestExtractFacts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed sessions
	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root', NULL, 'build', 'default_model', 'AL-10 Root', 1760000000000, 1760000100000),
		('ses_child', 'ses_root', 'coder', 'default_model', 'AL-10 Child', 1760000010000, 1760000090000);
	`)
	if err != nil {
		t.Fatalf("failed inserting sessions: %v", err)
	}

	// Seed compactions
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_comp_1', 'ses_root', 'msg_comp', 1760000020000, 1760000020000, '{"type":"compaction"}'),
		('p_comp_2', 'ses_root', 'msg_comp', 1760000030000, 1760000030000, '{"type":"compaction"}');
	`)
	if err != nil {
		t.Fatalf("failed inserting compactions: %v", err)
	}

	// Seed assistant messages
	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m1', 'ses_root', 1760000001000, 1760000005000, '{"role":"assistant","modelID":"claude-3-sonnet","time":{"created":1760000001000,"completed":1760000005000},"tokens":{"input":1000,"output":200,"cache":{"read":5000},"reasoning":50},"cost":0.05}'),
		('m2', 'ses_root', 1760000010000, 1760000020000, '{"role":"assistant","modelID":"claude-3-sonnet","time":{"created":1760000010000,"completed":1760000020000},"tokens":{"input":2000,"output":300,"cache":{"read":6000},"reasoning":100},"cost":0.10}'),
		('m3', 'ses_child', 1760000030000, 1760000045000, '{"role":"assistant","modelID":"claude-3-opus","time":{"created":1760000030000,"completed":1760000045000},"tokens":{"input":4000,"output":800,"cache":{"read":8000},"reasoning":200},"cost":0.50}');
	`)
	if err != nil {
		t.Fatalf("failed inserting messages: %v", err)
	}

	// Seed tool parts
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('t1', 'ses_root', 'm1', 1760000002000, 1760000002000, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"all good"}}'),
		('t2', 'ses_root', 'm2', 1760000011000, 1760000011000, '{"type":"tool","tool":"bash","state":{"status":"error","output":"permission denied"}}'),
		('t3', 'ses_child', 'm3', 1760000031000, 1760000031000, '{"type":"tool","tool":"read","state":{"status":"completed","output":"content"}}'),
		('t4', 'ses_child', 'm3', 1760000032000, 1760000032000, '{"type":"tool","tool":"read","state":{"status":"failed","output":"Error: file not found"}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting tools: %v", err)
	}

	invocations, tools, events, startedAt, completedAt, err := opencode.ExtractFacts(
		db,
		[]string{"ses_root"},
		"test-user",
		map[string]string{"ses_child": "test-user"},
	)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	if startedAt == nil || completedAt == nil {
		t.Fatalf("expected valid start and completed timestamps")
	}

	if len(invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(invocations))
	}

	// Invocations: ses_root (build) and ses_child (coder)
	rootInv := invocations[0]
	if rootInv.ID != "ses_root" || rootInv.Role != "build" {
		t.Errorf("expected root invocation role build, got ID=%s Role=%s", rootInv.ID, rootInv.Role)
	}
	if rootInv.User != "test-user" {
		t.Errorf("expected user test-user, got %s", rootInv.User)
	}
	if rootInv.Compactions != 2 {
		t.Errorf("expected 2 compactions on root invocation, got %d", rootInv.Compactions)
	}
	if rootInv.Tokens.Input != 3000 || rootInv.Tokens.Output != 500 || rootInv.Tokens.CacheRead != 11000 || rootInv.Tokens.Reasoning != 150 {
		t.Errorf("unexpected root tokens: %+v", rootInv.Tokens)
	}
	if rootInv.Tokens.PeakContext != 8000 { // max(1000+5000, 2000+6000) = 8000
		t.Errorf("expected peak context 8000, got %d", rootInv.Tokens.PeakContext)
	}
	if rootInv.ActiveDurationSeconds != 14 { // (5-1) + (20-10) = 14s
		t.Errorf("expected active duration 14s, got %d", rootInv.ActiveDurationSeconds)
	}
	if rootInv.CostUSD != 0.15 {
		t.Errorf("expected cost 0.15, got %f", rootInv.CostUSD)
	}

	childInv := invocations[1]
	if childInv.ID != "ses_child" || childInv.Role != "coder" {
		t.Errorf("expected child invocation role coder, got ID=%s Role=%s", childInv.ID, childInv.Role)
	}
	if childInv.User != "test-user" {
		t.Errorf("expected user test-user, got %s", childInv.User)
	}
	if childInv.Compactions != 0 {
		t.Errorf("expected 0 compactions on child, got %d", childInv.Compactions)
	}
	if childInv.Tokens.Input != 4000 || childInv.Tokens.Output != 800 || childInv.Tokens.CacheRead != 8000 || childInv.Tokens.Reasoning != 200 {
		t.Errorf("unexpected child tokens: %+v", childInv.Tokens)
	}
	if childInv.Tokens.PeakContext != 12000 {
		t.Errorf("expected peak context 12000, got %d", childInv.Tokens.PeakContext)
	}
	if childInv.ActiveDurationSeconds != 15 {
		t.Errorf("expected active duration 15s, got %d", childInv.ActiveDurationSeconds)
	}

	// Tools
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	for _, tool := range tools {
		if tool.Name == "bash" {
			if tool.Calls != 2 || tool.Errors != 1 {
				t.Errorf("expected bash calls=2 errors=1, got calls=%d errors=%d", tool.Calls, tool.Errors)
			}
		} else if tool.Name == "read" {
			if tool.Calls != 2 || tool.Errors != 1 {
				t.Errorf("expected read calls=2 errors=1, got calls=%d errors=%d", tool.Calls, tool.Errors)
			}
		} else {
			t.Errorf("unexpected tool %s", tool.Name)
		}
	}

	// Chronological Events
	if len(events) != 9 {
		t.Fatalf("expected 9 events, got %d", len(events))
	}

	// Check chronological ordering
	for i := 1; i < len(events); i++ {
		if events[i].Timestamp < events[i-1].Timestamp {
			t.Errorf("events not in chronological order: index %d (%s) < index %d (%s)",
				i, events[i].Timestamp, i-1, events[i-1].Timestamp)
		}
	}

	// Verify specific event properties
	ev0 := events[0] // m1 turn message
	if ev0.EventType != domain.EventTypeTurnMessage || ev0.Turn != 1 || ev0.ContextTokens != 6000 || ev0.Role != "build" {
		t.Errorf("unexpected event 0: %+v", ev0)
	}

	ev1 := events[1] // t1 tool call (bash)
	if ev1.EventType != domain.EventTypeToolCall || ev1.ToolName == nil || *ev1.ToolName != "bash" || ev1.Status != "completed" || ev1.Turn != 1 || ev1.ContextTokens != 6000 {
		t.Errorf("unexpected event 1: %+v", ev1)
	}

	ev2 := events[2] // m2 turn message
	if ev2.EventType != domain.EventTypeTurnMessage || ev2.Turn != 2 || ev2.ContextTokens != 8000 || ev2.Role != "build" {
		t.Errorf("unexpected event 2: %+v", ev2)
	}

	ev3 := events[3] // t2 tool call (bash error)
	if ev3.EventType != domain.EventTypeToolCall || ev3.Status != "error" || ev3.ErrorMessage != nil {
		t.Errorf("unexpected event 3: %+v", ev3)
	}

	ev4 := events[4] // p_comp_1 compaction
	if ev4.EventType != domain.EventTypeCompaction || ev4.Status != "completed" {
		t.Errorf("unexpected event 4: %+v", ev4)
	}

	ev8 := events[8] // t4 tool call (read error)
	if ev8.EventType != domain.EventTypeToolCall || ev8.Status != "error" || ev8.ErrorMessage != nil {
		t.Errorf("unexpected event 8: %+v", ev8)
	}
}

func TestExtractFacts_AdversarialEdgeCases(t *testing.T) {
	t.Run("empty db returns empty facts without panic", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		invocations, tools, events, startedAt, completedAt, err := opencode.ExtractFacts(
			db,
			[]string{"non_existent_thread"},
			"default_user",
			nil,
		)
		if err != nil {
			t.Fatalf("unexpected error on non-existent thread: %v", err)
		}
		if len(invocations) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(invocations))
		}
		if len(tools) != 0 {
			t.Errorf("expected 0 tools, got %d", len(tools))
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
		if startedAt != nil || completedAt != nil {
			t.Errorf("expected nil timestamps for empty thread, got start=%v completed=%v", startedAt, completedAt)
		}
	})

	t.Run("nil database returns explicit error", func(t *testing.T) {
		_, _, _, _, _, err := opencode.ExtractFacts(nil, []string{"thread_1"}, "user", nil)
		if err == nil {
			t.Fatal("expected error for nil db, got nil")
		}

		res, err := opencode.ResolveActiveRootSession(nil, "nonce", "ticket", "")
		if err == nil {
			t.Fatal("expected error for nil db in ResolveActiveRootSession, got nil")
		}
		if res != nil {
			t.Errorf("expected nil result, got %v", res)
		}
	})

	t.Run("unparseable JSON part data and corrupted message payloads are skipped gracefully", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		_, err := db.Exec(`
			INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
			('ses_corrupt', NULL, 'build', 'default_model', 'AL-Corrupt Session', 1760000000000, 1760000050000);
		`)
		if err != nil {
			t.Fatalf("failed inserting session: %v", err)
		}

		// Message 1: Completely invalid JSON syntax
		// Message 2: Null/empty data
		// Message 3: Missing role (not assistant)
		// Message 4: Valid assistant message with zero tokens
		_, err = db.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('m_bad_1', 'ses_corrupt', 1760000001000, 1760000002000, '{unparseable json!!'),
			('m_bad_2', 'ses_corrupt', 1760000003000, 1760000004000, ''),
			('m_bad_3', 'ses_corrupt', 1760000005000, 1760000006000, '{"role":"user","content":"hello"}'),
			('m_good', 'ses_corrupt', 1760000007000, 1760000009000, '{"role":"assistant","tokens":{"input":0,"output":0}}');
		`)
		if err != nil {
			t.Fatalf("failed inserting messages: %v", err)
		}

		// Part 1: Corrupted JSON
		// Part 2: Non-tool type
		// Part 3: Tool without tool name or state
		// Part 4: Tool with empty output
		_, err = db.Exec(`
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
			('p_bad_1', 'ses_corrupt', 'm_good', 1760000007500, 1760000007500, '<<<corrupted json>>>'),
			('p_bad_2', 'ses_corrupt', 'm_good', 1760000008000, 1760000008000, '{"type":"thinking","text":"just thinking"}'),
			('p_tool_unnamed', 'ses_corrupt', 'm_good', 1760000008500, 1760000008500, '{"type":"tool"}'),
			('p_tool_normal', 'ses_corrupt', 'm_good', 1760000008700, 1760000008700, '{"type":"tool","tool":"grep","state":{"status":"completed","output":""}}');
		`)
		if err != nil {
			t.Fatalf("failed inserting parts: %v", err)
		}

		invocations, tools, events, startedAt, completedAt, err := opencode.ExtractFacts(
			db,
			[]string{"ses_corrupt"},
			"test-user",
			nil,
		)
		if err != nil {
			t.Fatalf("ExtractFacts failed on corrupted rows: %v", err)
		}

		if startedAt == nil || completedAt == nil {
			t.Fatalf("expected non-nil startedAt and completedAt")
		}

		if len(invocations) != 1 {
			t.Fatalf("expected 1 invocation, got %d", len(invocations))
		}
		inv := invocations[0]
		if inv.Tokens.Input != 0 || inv.Tokens.Output != 0 || inv.CostUSD != 0.0 {
			t.Errorf("expected zero tokens and zero cost, got %+v, cost=%f", inv.Tokens, inv.CostUSD)
		}

		if len(events) != 3 {
			t.Errorf("expected 3 valid events (1 turn message, 2 tool calls), got %d: %+v", len(events), events)
		}

		if len(tools) != 2 {
			t.Fatalf("expected 2 tools (unknown and grep), got %d: %+v", len(tools), tools)
		}
		for _, tool := range tools {
			if tool.Name == "unknown" {
				if tool.Calls != 1 || tool.Errors != 0 {
					t.Errorf("unexpected unnamed tool stats: %+v", tool)
				}
			} else if tool.Name == "grep" {
				if tool.Calls != 1 || tool.Errors != 0 {
					t.Errorf("unexpected grep tool stats: %+v", tool)
				}
			} else {
				t.Errorf("unexpected tool name %s", tool.Name)
			}
		}
	})

	t.Run("thread ID sanitization blocks injection", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		injectionPayloads := []string{
			"ses_1' OR '1'='1",
			"ses_1; DROP TABLE session;--",
			"ses_1\x00malicious",
			" ",
		}

		for _, payload := range injectionPayloads {
			_, _, _, _, _, err := opencode.ExtractFacts(db, []string{payload}, "user", nil)
			if err == nil {
				t.Errorf("expected error for malicious thread ID %q, got nil", payload)
			}
		}
	})

	t.Run("corrupted SQLite file handling", func(t *testing.T) {
		tempDir := t.TempDir()
		corruptDBPath := filepath.Join(tempDir, "corrupted.db")
		// Write random garbage into file
		if err := os.WriteFile(corruptDBPath, []byte("NOT A SQLITE DATABASE HEADER AT ALL"), 0644); err != nil {
			t.Fatal(err)
		}

		client, err := opencode.OpenDB(corruptDBPath)
		if err == nil && client != nil {
			// modernc.org/sqlite might open and ping, but query fails
			defer client.Close()
			_, err = opencode.ResolveActiveRootSession(client.DB(), "nonce", "ticket", "")
			if err == nil {
				t.Error("expected error querying corrupted sqlite database")
			}
		}
	})
}

func TestExtractFacts_ChronologicalOrderingAndTieBreaking(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed session
	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_ord', NULL, 'build', 'default_model', 'AL-Order', 1000, 5000);
	`)
	if err != nil {
		t.Fatalf("failed inserting session: %v", err)
	}

	// Insert messages with identical timestamp 1760000000000
	// m1 (turn 1), m2 (turn 2), m3 (turn 3)
	// And one earlier message m0 (turn 0) at 1760000000000 - 1000
	// And one later message m4 (turn 4) at 1760000000000 + 1000
	const tieTime = 1760000010000
	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m0', 'ses_ord', 1760000005000, 1760000005000, '{"role":"assistant","tokens":{"input":100,"output":10}}'),
		('m1', 'ses_ord', ?, ?, '{"role":"assistant","tokens":{"input":200,"output":20}}'),
		('m2', 'ses_ord', ?, ?, '{"role":"assistant","tokens":{"input":300,"output":30}}'),
		('m3', 'ses_ord', ?, ?, '{"role":"assistant","tokens":{"input":400,"output":40}}'),
		('m4', 'ses_ord', 1760000020000, 1760000020000, '{"role":"assistant","tokens":{"input":500,"output":50}}');
	`, tieTime, tieTime, tieTime, tieTime, tieTime, tieTime)
	if err != nil {
		t.Fatalf("failed inserting messages: %v", err)
	}

	// Insert tool calls and compactions at the exact same tieTime (1760000010000)
	// correlated to different turns
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_tool_t1', 'ses_ord', 'm1', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"turn 2 tool"}}'),
		('p_comp_t2', 'ses_ord', 'm2', ?, ?, '{"type":"compaction"}'),
		('p_tool_t3', 'ses_ord', 'm3', ?, ?, '{"type":"tool","tool":"read","state":{"status":"completed","output":"turn 4 tool"}}');
	`, tieTime, tieTime, tieTime, tieTime, tieTime, tieTime)
	if err != nil {
		t.Fatalf("failed inserting parts: %v", err)
	}

	_, _, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_ord"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	if len(events) != 8 {
		t.Fatalf("expected 8 events (5 messages, 2 tools, 1 compaction), got %d", len(events))
	}

	// 1. Strict primary ordering by timestamp ascending
	for i := 1; i < len(events); i++ {
		if events[i].Timestamp < events[i-1].Timestamp {
			t.Fatalf("timestamp ordering violation at index %d: %s < %s",
				i, events[i].Timestamp, events[i-1].Timestamp)
		}
	}

	// 2. Strict secondary ordering by turn when timestamps are identical
	for i := 1; i < len(events); i++ {
		if events[i].Timestamp == events[i-1].Timestamp {
			if events[i].Turn < events[i-1].Turn {
				t.Fatalf("turn tie-break ordering violation at index %d: turn %d < previous turn %d (both at %s)",
					i, events[i].Turn, events[i-1].Turn, events[i].Timestamp)
			}
		}
	}
}

func TestExtractFacts_ToolCallErrorParsingAndDuration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_tools', NULL, 'build', 'default_model', 'AL-Tools', 1000, 10000);
	`)
	if err != nil {
		t.Fatalf("failed inserting session: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m1', 'ses_tools', 1760000001000, 1760000001000, '{"role":"assistant","tokens":{"input":1000,"cache":{"read":500}}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting message: %v", err)
	}

	// Long error string > 4096 chars (5000 chars)
	longErrOutput := "Error: " + strings.Repeat("x", 5000-len("Error: "))

	// Tool 1: Clean completed tool with 350ms duration (time_created=1000, time_updated=1350)
	// Tool 2: Tool with status "error" and 0ms duration
	// Tool 3: Tool with Traceback in output (status "completed")
	// Tool 4: Tool with "Error:" prefix in output (status "success")
	// Tool 5: Tool with long error output exceeding 4096 characters (truncated to 4096)
	// Tool 6: Tool with invalid inverted timestamps (time_updated < time_created) -> duration should be nil
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_clean', 'ses_tools', 'm1', 1760000002000, 1760000002350, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"all tests passed"}}'),
		('p_stat_err', 'ses_tools', 'm1', 1760000003000, 1760000003000, '{"type":"tool","tool":"bash","state":{"status":"error","output":"command failed with status 1"}}'),
		('p_traceback', 'ses_tools', 'm1', 1760000004000, 1760000004200, '{"type":"tool","tool":"python","state":{"status":"completed","output":"Traceback (most recent call last):\n  File script.py, line 5\nZeroDivisionError"}}'),
		('p_err_prefix', 'ses_tools', 'm1', 1760000005000, 1760000005100, '{"type":"tool","tool":"read","state":{"status":"success","output":"Error: path not found: /etc/missing"}}'),
		('p_long_err', 'ses_tools', 'm1', 1760000006000, 1760000006050, ?),
		('p_inv_time', 'ses_tools', 'm1', 1760000007000, 1760000006900, '{"type":"tool","tool":"grep","state":{"status":"completed","output":"matched"}}');
	`, `{"type":"tool","tool":"edit","state":{"status":"failed","output":"`+longErrOutput+`"}}`)
	if err != nil {
		t.Fatalf("failed inserting parts: %v", err)
	}

	_, tools, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_tools"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	// Verify ToolFact aggregations
	toolMap := make(map[string]domain.ToolFact)
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	// bash: 2 calls, 1 error
	if bash, ok := toolMap["bash"]; !ok || bash.Calls != 2 || bash.Errors != 1 {
		t.Errorf("expected bash calls=2 errors=1, got: %+v", bash)
	}
	// python: 1 call, 1 error (due to Traceback)
	if py, ok := toolMap["python"]; !ok || py.Calls != 1 || py.Errors != 1 {
		t.Errorf("expected python calls=1 errors=1 (Traceback), got: %+v", py)
	}
	// read: 1 call, 1 error (due to Error: prefix)
	if rd, ok := toolMap["read"]; !ok || rd.Calls != 1 || rd.Errors != 1 {
		t.Errorf("expected read calls=1 errors=1 (Error: prefix), got: %+v", rd)
	}
	// edit: 1 call, 1 error (due to Error: prefix)
	if ed, ok := toolMap["edit"]; !ok || ed.Calls != 1 || ed.Errors != 1 {
		t.Errorf("expected edit calls=1 errors=1 (Error: prefix), got: %+v", ed)
	}
	// grep: 1 call, 0 errors
	if gr, ok := toolMap["grep"]; !ok || gr.Calls != 1 || gr.Errors != 0 {
		t.Errorf("expected grep calls=1 errors=0, got: %+v", gr)
	}

	// Find tool call events
	eventMap := make(map[string]domain.EventFact)
	for _, ev := range events {
		if ev.ToolName != nil {
			eventMap[*ev.ToolName+"_"+ev.Status] = ev
		}
	}

	// Verify Clean Tool Call
	cleanEv := events[1] // p_clean
	if cleanEv.Status != "completed" || cleanEv.ErrorMessage != nil {
		t.Errorf("expected clean tool completed with nil error message, got: %+v", cleanEv)
	}
	if cleanEv.DurationMs == nil || *cleanEv.DurationMs != 350 {
		t.Errorf("expected clean tool duration 350ms, got: %v", cleanEv.DurationMs)
	}
	if cleanEv.ContextTokens != 1500 {
		t.Errorf("expected 1500 context tokens, got %d", cleanEv.ContextTokens)
	}

	// Verify Status Error Tool Call
	statErrEv := events[2] // p_stat_err
	if statErrEv.Status != "error" || statErrEv.ErrorMessage != nil {
		t.Errorf("expected status error tool call with nil error message, got: %+v", statErrEv)
	}
	if statErrEv.DurationMs == nil || *statErrEv.DurationMs != 0 {
		t.Errorf("expected 0ms duration, got: %v", statErrEv.DurationMs)
	}

	// Verify Traceback Detection Tool Call
	traceEv := events[3] // p_traceback
	if traceEv.Status != "error" || traceEv.ErrorMessage != nil {
		t.Errorf("expected Traceback detection error with nil error message, got: %+v", traceEv)
	}
	if traceEv.DurationMs == nil || *traceEv.DurationMs != 200 {
		t.Errorf("expected 200ms duration, got: %v", traceEv.DurationMs)
	}

	// Verify Error: Prefix Detection Tool Call
	errPrefixEv := events[4] // p_err_prefix
	if errPrefixEv.Status != "error" || errPrefixEv.ErrorMessage != nil {
		t.Errorf("expected Error: prefix detection with nil error message, got: %+v", errPrefixEv)
	}

	// Verify Long Error Message sets ErrorMessage to nil
	longErrEv := events[5] // p_long_err
	if longErrEv.Status != "error" || longErrEv.ErrorMessage != nil {
		t.Fatalf("expected long error to be flagged as error with nil ErrorMessage, got: %+v", longErrEv)
	}

	// Verify Inverted Timestamps yields nil DurationMs
	invTimeEv := events[6] // p_inv_time
	if invTimeEv.DurationMs != nil {
		t.Errorf("expected nil duration for inverted timestamps, got: %v", *invTimeEv.DurationMs)
	}
}

func TestExtractFacts_Sub4096ErrorPreservation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_sub', NULL, 'build', 'default_model', 'AL-Sub4096', 1000, 10000);
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_sub', 'ses_sub', 1760000001000, 1760000001000, '{"role":"assistant","tokens":{"input":100}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting session and message: %v", err)
	}

	sub4096Err := "Error: " + strings.Repeat("z", 1500-len("Error: "))
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_sub_err', 'ses_sub', 'm_sub', 1760000002000, 1760000002100, ?);
	`, `{"type":"tool","tool":"edit","state":{"status":"failed","output":"`+sub4096Err+`"}}`)
	if err != nil {
		t.Fatalf("failed inserting part: %v", err)
	}

	_, _, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_sub"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	ev := events[1]
	if ev.Status != "error" {
		t.Fatalf("expected error event, got: %+v", ev)
	}
	if ev.ErrorMessage != nil {
		t.Errorf("expected ErrorMessage nil under zero message persistence, got: %s", *ev.ErrorMessage)
	}
	if ev.ToolName == nil || *ev.ToolName != "edit" {
		t.Errorf("expected tool_name 'edit', got %v", ev.ToolName)
	}
}

func TestExtractFacts_ErrorPayloadAdversarialBoundaries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_adv_err', NULL, 'build', 'default_model', 'AL-AdvErr', 1000, 10000);
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_adv_err', 'ses_adv_err', 1760000001000, 1760000001000, '{"role":"assistant","tokens":{"input":100}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting session and message: %v", err)
	}

	exact4096 := "Error: " + strings.Repeat("A", 4096-len("Error: "))
	offByOne4097 := "Error: " + strings.Repeat("B", 4097-len("Error: "))
	huge5000 := "Error: " + strings.Repeat("C", 5000-len("Error: "))
	small100 := "Error: " + strings.Repeat("D", 100-len("Error: "))

	cases := []struct {
		name           string
		partID         string
		partJSON       string
		expectedStatus string
	}{
		{
			name:           "exact_4096_boundary_error_detected",
			partID:         "p_exact_4096",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":"` + exact4096 + `"}}`,
			expectedStatus: "error",
		},
		{
			name:           "off_by_one_4097_error_detected",
			partID:         "p_off_by_one_4097",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":"` + offByOne4097 + `"}}`,
			expectedStatus: "error",
		},
		{
			name:           "over_limit_5000_error_detected",
			partID:         "p_huge_5000",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":"` + huge5000 + `"}}`,
			expectedStatus: "error",
		},
		{
			name:           "sub_4096_payload_error_detected",
			partID:         "p_small_100",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":"` + small100 + `"}}`,
			expectedStatus: "error",
		},
		{
			name:           "empty_error_output_falls_back_to_status",
			partID:         "p_empty_err_output",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":""}}`,
			expectedStatus: "error",
		},
		{
			name:           "whitespace_only_error_output_trimmed_and_falls_back_to_status",
			partID:         "p_whitespace_err_output",
			partJSON:       `{"type":"tool","tool":"bash","state":{"status":"error","output":"   \n\t  "}}`,
			expectedStatus: "error",
		},
		{
			name:           "empty_status_with_error_prefix_handled",
			partID:         "p_err_prefix_clean",
			partJSON:       `{"type":"tool","tool":"bash","state":{"output":"Error: simple failure"}}`,
			expectedStatus: "error",
		},
	}

	for i, tc := range cases {
		_, err := db.Exec(`
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
			(?, 'ses_adv_err', 'm_adv_err', ?, ?, ?)
		`, tc.partID, 1760000002000+int64(i*1000), 1760000002000+int64(i*1000)+100, tc.partJSON)
		if err != nil {
			t.Fatalf("failed inserting part %s: %v", tc.partID, err)
		}
	}

	_, _, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_adv_err"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	var toolEvents []domain.EventFact
	for _, ev := range events {
		if ev.EventType == domain.EventTypeToolCall {
			toolEvents = append(toolEvents, ev)
		}
	}

	if len(toolEvents) != len(cases) {
		t.Fatalf("expected %d tool events, got %d", len(cases), len(toolEvents))
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := toolEvents[i]
			if ev.Status != tc.expectedStatus {
				t.Errorf("status mismatch: expected %q, got %q", tc.expectedStatus, ev.Status)
			}
			if ev.ErrorMessage != nil {
				t.Errorf("expected ErrorMessage nil under zero message persistence, got %q", *ev.ErrorMessage)
			}
			if ev.ToolName == nil || *ev.ToolName != "bash" {
				t.Errorf("expected tool name bash, got %v", ev.ToolName)
			}
		})
	}
}

func TestExtractFacts_CompactionEvents(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_comp', NULL, 'build', 'default_model', 'AL-Compaction', 1000, 10000);
	`)
	if err != nil {
		t.Fatalf("failed inserting session: %v", err)
	}

	// Message with tokens
	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_comp_1', 'ses_comp', 1760000001000, 1760000001500, '{"role":"assistant","tokens":{"input":8000,"cache":{"read":2000}}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting message: %v", err)
	}

	// 1. Compaction linked to message m_comp_1
	// 2. Compaction unlinked (message_id is NULL)
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_comp_linked', 'ses_comp', 'm_comp_1', 1760000002000, 1760000002000, '{"type":"compaction"}'),
		('p_comp_unlinked', 'ses_comp', NULL, 1760000003000, 1760000003000, '{"type":"compaction"}');
	`)
	if err != nil {
		t.Fatalf("failed inserting compaction parts: %v", err)
	}

	invocations, _, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_comp"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	// Compactions counted in invocation
	if len(invocations) != 1 {
		t.Fatalf("expected 1 invocation, got %d", len(invocations))
	}
	if invocations[0].Compactions != 2 {
		t.Errorf("expected 2 compactions on invocation, got %d", invocations[0].Compactions)
	}

	// Events: m_comp_1, p_comp_linked, p_comp_unlinked
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Linked compaction: turn=1, contextTokens=10000 (8000+2000)
	linkedComp := events[1]
	if linkedComp.EventType != domain.EventTypeCompaction || linkedComp.Status != "completed" {
		t.Errorf("unexpected linked compaction: %+v", linkedComp)
	}
	if linkedComp.Turn != 1 {
		t.Errorf("expected linked compaction turn 1, got %d", linkedComp.Turn)
	}
	if linkedComp.ContextTokens != 10000 {
		t.Errorf("expected linked compaction context tokens 10000, got %d", linkedComp.ContextTokens)
	}
	if linkedComp.DurationMs != nil {
		t.Errorf("expected nil duration on compaction, got %v", linkedComp.DurationMs)
	}

	// Unlinked compaction: defaults to turn=1, contextTokens=0
	unlinkedComp := events[2]
	if unlinkedComp.EventType != domain.EventTypeCompaction || unlinkedComp.Status != "completed" {
		t.Errorf("unexpected unlinked compaction: %+v", unlinkedComp)
	}
	if unlinkedComp.Turn != 1 {
		t.Errorf("expected unlinked compaction default turn 1, got %d", unlinkedComp.Turn)
	}
	if unlinkedComp.ContextTokens != 0 {
		t.Errorf("expected unlinked compaction default context tokens 0, got %d", unlinkedComp.ContextTokens)
	}
}

func TestExtractFacts_RawPersonaPreservation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root_build', NULL, 'build', 'default_model', 'AL-24 Root Build', 1000, 2000),
		('ses_child_sec', 'ses_root_build', 'security-scanner', 'default_model', 'AL-24 Child Security', 1100, 1900),
		('ses_root_empty', NULL, '', 'default_model', 'AL-24 Root Empty', 3000, 4000),
		('ses_child_empty', 'ses_root_empty', '   ', 'default_model', 'AL-24 Child Empty', 3100, 3900);
	`)
	if err != nil {
		t.Fatalf("failed inserting test sessions: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_build', 'ses_root_build', 1050, 1060, '{"role":"assistant","tokens":{"input":100,"output":20}}'),
		('m_sec', 'ses_child_sec', 1150, 1160, '{"role":"assistant","tokens":{"input":200,"output":40}}'),
		('m_rt_empty', 'ses_root_empty', 3050, 3060, '{"role":"assistant","tokens":{"input":50,"output":10}}'),
		('m_ch_empty', 'ses_child_empty', 3150, 3160, '{"role":"assistant","tokens":{"input":60,"output":12}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting test messages: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_tool_sec', 'ses_child_sec', 'm_sec', 1155, 1155, '{"type":"tool","tool":"trivy","state":{"status":"completed","output":"no vulnerabilities"}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting test parts: %v", err)
	}

	t.Run("preserves raw personas for build and security-scanner", func(t *testing.T) {
		invocations, tools, events, _, _, err := opencode.ExtractFacts(
			db,
			[]string{"ses_root_build"},
			"test-user",
			nil,
		)
		if err != nil {
			t.Fatalf("ExtractFacts failed: %v", err)
		}

		if len(invocations) != 2 {
			t.Fatalf("expected 2 invocations, got %d", len(invocations))
		}

		var rootInv, childInv *domain.InvocationFact
		for i := range invocations {
			if invocations[i].ID == "ses_root_build" {
				rootInv = &invocations[i]
			} else if invocations[i].ID == "ses_child_sec" {
				childInv = &invocations[i]
			}
		}

		if rootInv == nil || rootInv.Role != "build" {
			t.Fatalf("expected root invocation role 'build', got: %+v", rootInv)
		}
		if childInv == nil || childInv.Role != "security-scanner" {
			t.Fatalf("expected child invocation role 'security-scanner', got: %+v", childInv)
		}

		if len(tools) != 1 || tools[0].Name != "trivy" {
			t.Fatalf("expected 1 trivy tool fact, got: %+v", tools)
		}

		var foundBuildEvent, foundSecEvent, foundToolEvent bool
		for _, ev := range events {
			if ev.SessionID == "ses_root_build" && ev.Role == "build" && ev.EventType == domain.EventTypeTurnMessage {
				foundBuildEvent = true
			}
			if ev.SessionID == "ses_child_sec" && ev.Role == "security-scanner" && ev.EventType == domain.EventTypeTurnMessage {
				foundSecEvent = true
			}
			if ev.SessionID == "ses_child_sec" && ev.Role == "security-scanner" && ev.EventType == domain.EventTypeToolCall {
				foundToolEvent = true
			}
		}

		if !foundBuildEvent {
			t.Errorf("missing turn event with role 'build'")
		}
		if !foundSecEvent {
			t.Errorf("missing turn event with role 'security-scanner'")
		}
		if !foundToolEvent {
			t.Errorf("missing tool call event with role 'security-scanner'")
		}
	})

	t.Run("empty agent fallback for root and child", func(t *testing.T) {
		invocations, _, events, _, _, err := opencode.ExtractFacts(
			db,
			[]string{"ses_root_empty"},
			"test-user",
			nil,
		)
		if err != nil {
			t.Fatalf("ExtractFacts failed: %v", err)
		}

		if len(invocations) != 2 {
			t.Fatalf("expected 2 invocations, got %d", len(invocations))
		}

		var rootInv, childInv *domain.InvocationFact
		for i := range invocations {
			if invocations[i].ID == "ses_root_empty" {
				rootInv = &invocations[i]
			} else if invocations[i].ID == "ses_child_empty" {
				childInv = &invocations[i]
			}
		}

		if rootInv == nil || rootInv.Role != domain.RoleRootAgent {
			t.Fatalf("expected root invocation fallback role %q, got: %+v", domain.RoleRootAgent, rootInv)
		}
		if childInv == nil || childInv.Role != domain.RoleUnassignedSubAgent {
			t.Fatalf("expected child invocation fallback role %q, got: %+v", domain.RoleUnassignedSubAgent, childInv)
		}

		for _, ev := range events {
			if ev.SessionID == "ses_root_empty" && ev.Role != domain.RoleRootAgent {
				t.Errorf("expected root event role %q, got: %s", domain.RoleRootAgent, ev.Role)
			}
			if ev.SessionID == "ses_child_empty" && ev.Role != domain.RoleUnassignedSubAgent {
				t.Errorf("expected child event role %q, got: %s", domain.RoleUnassignedSubAgent, ev.Role)
			}
		}
	})
}

func TestExtractFacts_AdversarialPersonaPreservationAndWhitespace(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed sessions with whitespace-padded agent personas and complex identifiers
	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root_ws', NULL, '   build   ', 'default_model', 'AL-24 Whitespace Build', 1000, 2000),
		('ses_child_ws', 'ses_root_ws', '   sec-ops/v1-alpha   ', 'default_model', 'AL-24 Security Ops', 1100, 1900),
		('ses_child_deep', 'ses_root_ws', '   deep.researcher+tools   ', 'default_model', 'AL-24 Deep Researcher', 1200, 1800);
	`)
	if err != nil {
		t.Fatalf("failed inserting sessions: %v", err)
	}

	// Seed assistant messages
	_, err = db.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_ws_1', 'ses_root_ws', 1010, 1050, '{"role":"assistant","tokens":{"input":100,"output":20}}'),
		('m_ws_2', 'ses_child_ws', 1110, 1150, '{"role":"assistant","tokens":{"input":200,"output":40}}'),
		('m_ws_3', 'ses_child_deep', 1210, 1250, '{"role":"assistant","tokens":{"input":300,"output":60}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting messages: %v", err)
	}

	// Seed tool parts
	_, err = db.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_ws_1', 'ses_child_ws', 'm_ws_2', 1120, 1120, '{"type":"tool","tool":"scanner","state":{"status":"completed","output":"clean"}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting tool part: %v", err)
	}

	invocations, _, events, _, _, err := opencode.ExtractFacts(
		db,
		[]string{"ses_root_ws"},
		"test-user",
		nil,
	)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	if len(invocations) != 3 {
		t.Fatalf("expected 3 invocations, got %d", len(invocations))
	}

	invMap := make(map[string]domain.InvocationFact)
	for _, inv := range invocations {
		invMap[inv.ID] = inv
	}

	// Assert whitespace was trimmed and raw personas preserved without coercion
	if inv, ok := invMap["ses_root_ws"]; !ok || inv.Role != "build" {
		t.Errorf("expected trimmed role 'build' for root session, got: %+v", inv)
	}
	if inv, ok := invMap["ses_child_ws"]; !ok || inv.Role != "sec-ops/v1-alpha" {
		t.Errorf("expected trimmed role 'sec-ops/v1-alpha', got: %+v", inv)
	}
	if inv, ok := invMap["ses_child_deep"]; !ok || inv.Role != "deep.researcher+tools" {
		t.Errorf("expected trimmed role 'deep.researcher+tools', got: %+v", inv)
	}

	// Verify events inherited the trimmed raw persona
	for _, ev := range events {
		switch ev.SessionID {
		case "ses_root_ws":
			if ev.Role != "build" {
				t.Errorf("expected event role 'build', got %q", ev.Role)
			}
		case "ses_child_ws":
			if ev.Role != "sec-ops/v1-alpha" {
				t.Errorf("expected event role 'sec-ops/v1-alpha', got %q", ev.Role)
			}
		case "ses_child_deep":
			if ev.Role != "deep.researcher+tools" {
				t.Errorf("expected event role 'deep.researcher+tools', got %q", ev.Role)
			}
		}
	}
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

func TestExtractSteeringTurns(t *testing.T) {
	t.Run("nil database returns error", func(t *testing.T) {
		count, err := opencode.ExtractSteeringTurns(nil, []string{"ses_1"})
		if err == nil {
			t.Errorf("expected error for nil db, got nil")
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("empty thread IDs returns zero with nil error", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		count, err := opencode.ExtractSteeringTurns(db, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("invalid session id returns validation error", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		count, err := opencode.ExtractSteeringTurns(db, []string{"bad session; drop table message;"})
		if err == nil {
			t.Errorf("expected error for invalid session id, got nil")
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("single and multiple user turns counted accurately", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		_, err := db.Exec(`
			INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES
			('ses_root', NULL, 'build', '{"id":"claude-3-opus"}', 1000, 5000),
			('ses_child', 'ses_root', 'coder', '{"id":"claude-3-opus"}', 1100, 4000);
		`)
		if err != nil {
			t.Fatalf("failed inserting test sessions: %v", err)
		}

		_, err = db.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('m1', 'ses_root', 1010, 1010, '{"role":"user","tokens":{"input":10}}'),
			('m2', 'ses_root', 1020, 1020, '{"role":"assistant","tokens":{"output":20}}'),
			('m3', 'ses_root', 1030, 1030, '{"role":"User","tokens":{"input":15}}'),
			('m4', 'ses_root', 1040, 1040, '{"role":"USER ","tokens":{"input":12}}'),
			('m5', 'ses_child', 1050, 1050, '{"role":"user","tokens":{"input":8}}'),
			('m6', 'ses_root', 1060, 1060, '{"role":"system","tokens":{"input":5}}'),
			('m7', 'ses_root', 1070, 1070, 'invalid json content'),
			('m8', 'ses_root', 1080, 1080, NULL);
		`)
		if err != nil {
			t.Fatalf("failed inserting test messages: %v", err)
		}

		// When querying root session: 3 user turns ('user', 'User', 'USER '), child message ignored, invalid JSON ignored
		count, err := opencode.ExtractSteeringTurns(db, []string{"ses_root"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 steering turns for ses_root, got %d", count)
		}

		// When querying child session
		childCount, err := opencode.ExtractSteeringTurns(db, []string{"ses_child"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if childCount != 1 {
			t.Errorf("expected 1 steering turn for ses_child, got %d", childCount)
		}
	})

	t.Run("only assistant messages returns zero turns", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		_, err := db.Exec(`
			INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES
			('ses_asst', NULL, 'build', '{"id":"claude-3-opus"}', 1000, 2000);
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('m1', 'ses_asst', 1010, 1010, '{"role":"assistant"}'),
			('m2', 'ses_asst', 1020, 1020, '{"role":"assistant"}');
		`)
		if err != nil {
			t.Fatalf("failed setting up assistant-only messages: %v", err)
		}

		count, err := opencode.ExtractSteeringTurns(db, []string{"ses_asst"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 steering turns, got %d", count)
		}
	})
}

func TestExtractFacts_ContentInspectionTools_CompletedStatusAndMessageID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_insp', NULL, 'build', 'default_model', 'AL-Inspection', 1000, 10000);
		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m_insp_1', 'ses_insp', 1760000001000, 1760000001500, '{"role":"assistant","tokens":{"input":100}}'),
		('m_insp_2', 'ses_insp', 1760000005000, 1760000005500, '{"role":"assistant","tokens":{"input":200}}');
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_read_syntax', 'ses_insp', 'm_insp_1', 1760000002000, 1760000002100, '{"type":"tool","tool":"read","state":{"status":"completed","output":"def main():\n  syntax error: unexpected EOF while parsing"}}'),
		('p_read_tb', 'ses_insp', 'm_insp_1', 1760000003000, 1760000003100, '{"type":"tool","tool":"read","state":{"status":"completed","output":"# Troubleshooting\nTraceback (most recent call last):\n  File run.py, line 5\nZeroDivisionError"}}'),
		('p_grep_err', 'ses_insp', 'm_insp_1', 1760000004000, 1760000004100, '{"type":"tool","tool":"grep","state":{"status":"completed","output":"Error: log pattern matched in syslog"}}'),
		('p_glob_panic', 'ses_insp', 'm_insp_1', 1760000004500, 1760000004600, '{"type":"tool","tool":"glob","state":{"status":"completed","output":"panic: runtime error in search target"}}'),
		('p_webfetch_err', 'ses_insp', 'm_insp_1', 1760000004700, 1760000004800, '{"type":"tool","tool":"webfetch","state":{"status":"completed","output":"Error: remote service temporarily unavailable"}}'),
		('p_read_real_err', 'ses_insp', 'm_insp_2', 1760000006000, 1760000006100, '{"type":"tool","tool":"read","state":{"status":"error","output":"file not found: /etc/missing.conf"}}'),
		('p_bash_err', 'ses_insp', 'm_insp_2', 1760000007000, 1760000007100, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"Error: command execution failed with non-zero code"}}');
	`)
	if err != nil {
		t.Fatalf("failed inserting inspection test data: %v", err)
	}

	_, tools, events, _, _, err := opencode.ExtractFacts(db, []string{"ses_insp"}, "user", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	toolMap := make(map[string]domain.ToolFact)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	// read: 3 calls: 2 completed with syntax error / Traceback (0 errors), 1 error with file not found (1 error)
	readTool, ok := toolMap["read"]
	if !ok {
		t.Fatalf("expected read tool in extracted facts")
	}
	if readTool.Calls != 3 {
		t.Errorf("expected read calls=3, got %d", readTool.Calls)
	}
	if readTool.Errors != 1 {
		t.Errorf("expected read errors=1 (only from status error part), got %d", readTool.Errors)
	}

	// grep: 1 call completed -> 0 errors despite "Error:" in output
	grepTool, ok := toolMap["grep"]
	if !ok || grepTool.Errors != 0 {
		t.Errorf("expected grep errors=0, got %+v", grepTool)
	}

	// glob: 1 call completed -> 0 errors despite "panic:" in output
	globTool, ok := toolMap["glob"]
	if !ok || globTool.Errors != 0 {
		t.Errorf("expected glob errors=0, got %+v", globTool)
	}

	// webfetch: 1 call completed -> 0 errors despite "Error:" in output
	webfetchTool, ok := toolMap["webfetch"]
	if !ok || webfetchTool.Errors != 0 {
		t.Errorf("expected webfetch errors=0, got %+v", webfetchTool)
	}

	// bash: 1 call completed with "syntax error" -> non-inspection tool, so it MUST flag as error
	bashTool, ok := toolMap["bash"]
	if !ok || bashTool.Errors != 1 {
		t.Errorf("expected bash errors=1 (syntax error on non-inspection tool), got %+v", bashTool)
	}

	// Verify EventFact message_id propagation and statuses
	for _, ev := range events {
		switch ev.EventType {
		case domain.EventTypeTurnMessage:
			if ev.MessageID == nil {
				t.Errorf("expected non-nil MessageID on turn message event: %+v", ev)
			} else if *ev.MessageID != "m_insp_1" && *ev.MessageID != "m_insp_2" {
				t.Errorf("unexpected MessageID on turn message: %s", *ev.MessageID)
			}
		case domain.EventTypeToolCall:
			if ev.ToolName == nil {
				continue
			}
			switch *ev.ToolName {
			case "read":
				if ev.MessageID == nil {
					t.Errorf("expected non-nil MessageID on read tool call event")
				}
				if *ev.MessageID == "m_insp_1" {
					if ev.Status != "completed" {
						t.Errorf("expected completed status for read tool on m_insp_1, got %s", ev.Status)
					}
				} else if *ev.MessageID == "m_insp_2" {
					if ev.Status != "error" {
						t.Errorf("expected error status for read tool on m_insp_2 with file not found, got %s", ev.Status)
					}
				}
			case "grep", "glob", "webfetch":
				if ev.Status != "completed" {
					t.Errorf("expected completed status for %s, got %s", *ev.ToolName, ev.Status)
				}
				if ev.MessageID == nil || *ev.MessageID != "m_insp_1" {
					t.Errorf("expected MessageID m_insp_1 for %s, got %v", *ev.ToolName, ev.MessageID)
				}
			}
		}
	}
}

func TestExtractFacts_UserIterationsRootAndChild(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed opening in-memory db: %v", err)
	}
	defer db.Close()

	createSchema := `
	CREATE TABLE session (
		id TEXT PRIMARY KEY,
		parent_id TEXT,
		agent TEXT,
		model TEXT,
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
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatalf("failed creating test tables: %v", err)
	}

	// Root session with 2 models and 3 user messages
	_, _ = db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES ('root_sess', NULL, 'root', 'gpt-4o', 1000, 2000)`)
	// Child session with 1 user message
	_, _ = db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES ('child_sess', 'root_sess', 'coder', 'gpt-4o', 1500, 2000)`)

	// Root messages: 3 user messages, assistant msg on model1, assistant msg on model2
	uData := `{"role": "user"}`
	a1Data := `{"role": "assistant", "modelID": "model-1", "tokens": {"input": 10}}`
	a2Data := `{"role": "assistant", "modelID": "model-2", "tokens": {"input": 20}}`
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('rm1', 'root_sess', 1000, 1000, ?),
		('rm2', 'root_sess', 1100, 1100, ?),
		('rm3', 'root_sess', 1200, 1200, ?),
		('rm4', 'root_sess', 1300, 1300, ?),
		('rm5', 'root_sess', 1400, 1400, ?)`,
		uData, a1Data, uData, a2Data, uData,
	)

	// Child messages: 1 user message, 1 assistant message
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('cm1', 'child_sess', 1500, 1500, ?),
		('cm2', 'child_sess', 1600, 1600, ?)`,
		uData, a1Data,
	)

	invs, _, _, _, _, err := opencode.ExtractFacts(db, []string{"root_sess"}, "testuser", nil)
	if err != nil {
		t.Fatalf("ExtractFacts failed: %v", err)
	}

	// Expect 2 invocations for root (model-1, model-2) and 1 invocation for child
	if len(invs) != 3 {
		t.Fatalf("expected 3 invocations, got %d", len(invs))
	}

	for _, inv := range invs {
		if inv.ID == "root_sess" {
			if inv.Model == "model-1" {
				// Primary model on root session gets all user iterations (3)
				if inv.UserIterations != 3 {
					t.Errorf("expected primary root invocation to have UserIterations=3, got %d", inv.UserIterations)
				}
			} else if inv.Model == "model-2" {
				// Secondary model on root session gets 0
				if inv.UserIterations != 0 {
					t.Errorf("expected secondary root invocation to have UserIterations=0, got %d", inv.UserIterations)
				}
			}
		} else if inv.ID == "child_sess" {
			// Child session gets 0
			if inv.UserIterations != 0 {
				t.Errorf("expected child invocation to have UserIterations=0, got %d", inv.UserIterations)
			}
		}
	}
}
