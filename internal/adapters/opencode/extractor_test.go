package opencode_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/opencode"
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

	// Test Priority 4: Fallback to latest message
	t.Run("Priority 4 Latest Message Fallback", func(t *testing.T) {
		_, err := db.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
			('msg_latest', 'ses_child_1', 5000, 5000, '{"role":"assistant"}');
		`)
		if err != nil {
			t.Fatalf("failed inserting message: %v", err)
		}

		res, err := opencode.ResolveActiveRootSession(db, "", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil || *res != "ses_root_1" {
			t.Fatalf("expected ses_root_1 as parent of latest message session, got %v", res)
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

	invocations, tools, startedAt, completedAt, err := opencode.ExtractFacts(
		db,
		[]string{"ses_root"},
		"default_user@test.com",
		map[string]string{"ses_child": "special_coder@test.com"},
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

	// Invocations: ses_root (orchestrator) and ses_child (coder)
	rootInv := invocations[0]
	if rootInv.ID != "ses_root" || rootInv.Role != "orchestrator" {
		t.Errorf("expected root invocation role orchestrator, got ID=%s Role=%s", rootInv.ID, rootInv.Role)
	}
	if rootInv.User != "default_user@test.com" {
		t.Errorf("expected user default_user@test.com, got %s", rootInv.User)
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
	if childInv.User != "special_coder@test.com" {
		t.Errorf("expected user special_coder@test.com, got %s", childInv.User)
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
}

func TestExtractFacts_AdversarialEdgeCases(t *testing.T) {
	t.Run("empty db returns empty facts without panic", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		invocations, tools, startedAt, completedAt, err := opencode.ExtractFacts(
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
		if startedAt != nil || completedAt != nil {
			t.Errorf("expected nil timestamps for empty thread, got start=%v completed=%v", startedAt, completedAt)
		}
	})

	t.Run("nil database returns explicit error", func(t *testing.T) {
		_, _, _, _, err := opencode.ExtractFacts(nil, []string{"thread_1"}, "user", nil)
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

		invocations, tools, startedAt, completedAt, err := opencode.ExtractFacts(
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
			_, _, _, _, err := opencode.ExtractFacts(db, []string{payload}, "user", nil)
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

