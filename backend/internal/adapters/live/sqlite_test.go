package live_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func testProjectConfig() domain.ProjectConfig {
	return domain.ProjectConfig{
		ProjectName: "AgenticLens",
		Ticket: domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "AL", Description: "Core AgenticLens development"},
				{Prefix: "GEN", Description: "General tasks", TicketType: "chore"},
				{Prefix: "PO", Description: "Product Owner specs", TicketType: "docs"},
			},
			CaseInsensitive: true,
			CustomPatterns:  []string{},
		},
		Workspaces: domain.WorkspacesConfig{
			Dir: "workspaces",
		},
		CodeMetrics: domain.CodeMetricsConfig{SourceRoots: []string{"web", "backend"}},
		Database: domain.DatabaseConfig{
			Path:           "~/.local/share/opencode/opencode.db",
			QueryTimeoutMs: 5000,
			BusyTimeoutMs:  5000,
			ReadOnly:       true,
		},
		Live: domain.LiveConfig{
			ActiveThresholdSeconds: 300,
			PollIntervalMs:         2000,
			MaxTraceBufferMessages: 500,
		},
		AnomalyThresholds: domain.AnomalyThresholdsConfig{
			StallThresholdSeconds: 120,
			ErrorRateThreshold:    0.25,
			ContextExhaustionPct:  0.85,
		},
		RoleMappings: domain.RoleMappings{
			"orchestrator": "orchestrator",
			"planner":      "planner",
			"coder":        "coder",
			"tester":       "tester",
			"reviewer":     "reviewer",
		},
		RoleOrder: []string{
			"orchestrator",
			"planner",
			"coder",
			"tester",
			"reviewer",
		},
	}
}

func createTestDBFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_opencode.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed creating test db file: %v", err)
	}
	defer db.Close()

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		t.Fatalf("failed setting WAL mode: %v", err)
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
		t.Fatalf("failed creating test schema: %v", err)
	}

	return dbPath
}

func TestReadOnlyPragmaEnforcement(t *testing.T) {
	dbPath := createTestDBFile(t)

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// Attempt write mutation
	_, err = hook.DB().Exec("INSERT INTO session (id, title) VALUES ('s1', 'Attempted write');")
	if err == nil {
		t.Fatal("expected write mutation to fail on read-only connection, but it succeeded")
	}

	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "readonly") && !strings.Contains(errStr, "read-only") {
		t.Errorf("expected read-only error message, got: %v", err)
	}
}

func TestNonLockingConcurrencyUnderWriteLoad(t *testing.T) {
	dbPath := createTestDBFile(t)

	// Seed one initial session
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_initial', NULL, 'orchestrator', 'gemini', 'Initial AL-10', ?, ?);
	`, now, now)
	if err != nil {
		t.Fatalf("failed seeding session: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// Begin an exclusive write transaction on the RW connection
	tx, err := rwDB.Begin()
	if err != nil {
		t.Fatalf("failed starting write transaction: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_writing', NULL, 'coder', 'gemini', 'In progress AL-11', ?, ?);
	`, now+100, now+100)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("failed writing inside transaction: %v", err)
	}

	// While transaction is open, perform concurrent read via read-only hook
	doneCh := make(chan error, 1)
	go func() {
		records, readErr := hook.FetchActiveSessions(300, now)
		if readErr != nil {
			doneCh <- readErr
			return
		}
		if len(records) == 0 {
			doneCh <- fmt.Errorf("expected at least 1 record, got 0")
			return
		}
		doneCh <- nil
	}()

	select {
	case readErr := <-doneCh:
		if readErr != nil {
			t.Fatalf("read during active write transaction failed: %v", readErr)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("read query blocked or timed out during active writer transaction")
	}

	// Complete writer transaction
	if err := tx.Commit(); err != nil {
		t.Fatalf("failed committing writer transaction: %v", err)
	}
}

func TestZeroDiskPersistenceVerification(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_live', NULL, 'orchestrator', 'gemini-3.8-flash', 'Task AL-17: implement live telemetry', ?, ?);
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m1', 's_live', ?, ?, '{"role":"assistant","content":"executing command"}');
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p1', 's_live', 'm1', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"completed","output":"ok"}}');
	`, now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("seeding test data: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// Snapshot workspaces directory modification times
	workspacesDir := "workspaces"
	type fileMeta struct {
		modTime time.Time
		size    int64
	}
	beforeSnapshot := make(map[string]fileMeta)

	if fi, statErr := os.Stat(workspacesDir); statErr == nil && fi.IsDir() {
		_ = filepath.Walk(workspacesDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				beforeSnapshot[path] = fileMeta{
					modTime: info.ModTime(),
					size:    info.Size(),
				}
			}
			return nil
		})
	}

	cfg := testProjectConfig()
	service := live.NewLiveService(hook, cfg)
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Execute session tree generation
	tree, err := service.GetSessionTree(ctx)
	if err != nil {
		t.Fatalf("GetSessionTree failed: %v", err)
	}
	if len(tree.RootSessions) == 0 {
		t.Fatalf("expected at least 1 root session, got 0")
	}

	// Verify inferred ticket ID from title
	root := tree.RootSessions[0]
	if root.GuessedTicketID == nil || *root.GuessedTicketID != "AL-17" {
		t.Errorf("expected guessed ticket AL-17, got %v", root.GuessedTicketID)
	}
	if root.TicketConfidence != 1.0 {
		t.Errorf("expected ticket confidence 1.0, got %f", root.TicketConfidence)
	}

	// Execute trace extraction
	trace, err := service.GetSessionTrace(ctx, "s_live")
	if err != nil {
		t.Fatalf("GetSessionTrace failed: %v", err)
	}
	if len(trace.Messages) == 0 {
		t.Fatalf("expected at least 1 message, got 0")
	}

	// Verify buffer contains message transiently

	// Verify zero persistence to workspaces/
	if fi, statErr := os.Stat(workspacesDir); statErr == nil && fi.IsDir() {
		_ = filepath.Walk(workspacesDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				before, existed := beforeSnapshot[path]
				if !existed {
					t.Errorf("Zero persistence violation: new file created in workspaces: %s", path)
				} else if info.ModTime() != before.modTime || info.Size() != before.size {
					t.Errorf("Zero persistence violation: file modified in workspaces: %s", path)
				}
			}
			return nil
		})
	}
}

func TestConcurrentReadsAndSubscriptions(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_conc', NULL, 'orchestrator', 'gemini', 'Concurrent AL-17', ?, ?);
	`, now, now)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			tree, err := service.GetSessionTree(ctx)
			if err != nil {
				t.Errorf("concurrent GetSessionTree error: %v", err)
				return
			}
			if len(tree.RootSessions) == 0 {
				t.Errorf("concurrent GetSessionTree returned 0 roots")
			}
		}()
	}

	wg.Wait()
}

func TestOpenReadOnlyHook_NonExistentFileFails(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does_not_exist.db")
	_, err := live.OpenReadOnlyHook(nonExistentPath, 1000)
	if err == nil {
		t.Fatal("expected OpenReadOnlyHook to fail for non-existent file, but got nil error")
	}
}

func TestOpenReadOnlyHook_CorruptedDatabaseFails(t *testing.T) {
	corruptedPath := filepath.Join(t.TempDir(), "corrupted.db")
	if err := os.WriteFile(corruptedPath, []byte("NOT A SQLITE DATABASE FILE - CORRUPTED HEADER"), 0644); err != nil {
		t.Fatalf("writing corrupted db file: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(corruptedPath, 1000)
	if err == nil {
		// If open succeeded, ping or subsequent queries must fail
		defer hook.Close()
		_, fetchErr := hook.FetchActiveSessions(300, time.Now().UnixMilli())
		if fetchErr == nil {
			t.Fatal("expected querying corrupted database to fail, but it succeeded")
		}
	}
}

func TestFetchActiveSessions_EmptyDatabaseReturnsEmpty(t *testing.T) {
	dbPath := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	records, err := hook.FetchActiveSessions(300, time.Now().UnixMilli())
	if err != nil {
		t.Fatalf("FetchActiveSessions on empty db failed: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records from empty db, got %d", len(records))
	}
}

func TestFetchActiveSessions_NullAndMalformedColumns(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	// Insert row with NULLs in all optional columns
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_nulls', NULL, NULL, NULL, NULL, NULL, NULL);
	`)
	if err != nil {
		t.Fatalf("inserting null-heavy session: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// Query without time cutoff to fetch all rows
	records, err := hook.FetchActiveSessions(0, 0)
	if err != nil {
		t.Fatalf("FetchActiveSessions failed scanning row with NULL columns: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	rec := records[0]
	if rec.ID != "s_nulls" {
		t.Errorf("expected ID s_nulls, got %s", rec.ID)
	}
	if rec.ParentID != nil {
		t.Errorf("expected nil ParentID, got %v", *rec.ParentID)
	}
	if rec.Agent != "" || rec.Model != "" || rec.Title != "" {
		t.Errorf("expected empty strings for NULL columns, got agent=%q model=%q title=%q", rec.Agent, rec.Model, rec.Title)
	}
	if rec.TimeCreated != 0 || rec.TimeUpdated != 0 {
		t.Errorf("expected 0 for NULL timestamps, got created=%d updated=%d", rec.TimeCreated, rec.TimeUpdated)
	}
}

func TestFetchSessionTrace_MalformedJSONPartsFallback(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_corrupt_trace', NULL, 'coder', 'gemini', 'Corrupted trace session', ?, ?);
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_bad_json', 's_corrupt_trace', ?, ?, '{invalid: json: syntax');
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_bad_json', 's_corrupt_trace', 'm_bad_json', ?, ?, 'completely random binary \x00\x01\x02 bytes');
	`, now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("seeding corrupt trace data: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	trace, err := service.GetSessionTrace(ctx, "s_corrupt_trace")
	if err != nil {
		t.Fatalf("GetSessionTrace failed on corrupted json row: %v", err)
	}
	if len(trace.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(trace.Messages))
	}
	msg := trace.Messages[0]
	if len(msg.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(msg.Parts))
	}
	part := msg.Parts[0]
	if part.ID != "p_bad_json" {
		t.Errorf("expected part ID p_bad_json, got %s", part.ID)
	}
}

func TestLiveService_AdversarialSessionIDValidation(t *testing.T) {
	dbPath := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	traversalIDs := []string{
		"../ses1",
		"..\\ses1",
		"../../etc/passwd",
		"ses1/nested",
		"ses1\\nested",
		"%2e%2e%2froot",
		"%2e%2e/root",
		"..%2froot",
		"%5c..%5cboot.ini",
		"",
		"   ",
	}

	for _, invalidID := range traversalIDs {
		t.Run("traversal_"+invalidID, func(t *testing.T) {
			_, err := service.GetSessionTrace(ctx, invalidID)
			if err == nil {
				t.Errorf("expected error for invalid session id %q, got nil", invalidID)
			}
			if !strings.Contains(err.Error(), "invalid session id") {
				t.Errorf("expected invalid session id error for %q, got: %v", invalidID, err)
			}
		})
	}
}

func TestFetchSessionTraceMulti_And_ResolveChildSessionIDs(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// 1. Seed parent session and child session
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES
			('s_parent', NULL, 'orchestrator', 'gemini-3.8-flash', 'Parent AL-33', ?, ?),
			('s_child', 's_parent', 'coder', 'gemini-3.8-flash', 'Child AL-33', ?, ?);
	`, now, now, now+1000, now+1000)
	if err != nil {
		t.Fatalf("seeding sessions: %v", err)
	}

	// 2. Seed messages and parts for both sessions
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES
			('m_p1', 's_parent', ?, ?, '{"role":"assistant","content":"parent msg"}'),
			('m_c1', 's_child', ?, ?, '{"role":"assistant","content":"child msg"}');
	`, now+100, now+200, now+1100, now+1200)
	if err != nil {
		t.Fatalf("seeding messages: %v", err)
	}

	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES
			('p_p1', 's_parent', 'm_p1', ?, ?, '{"type":"tool","tool":"read","state":{"status":"completed","input":"file.txt","output":"ok"}}'),
			('p_c1', 's_child', 'm_c1', ?, ?, '{"type":"tool","tool":"edit","state":{"status":"completed","input":"file.go","output":"done"}}');
	`, now+150, now+180, now+1150, now+1180)
	if err != nil {
		t.Fatalf("seeding parts: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// Test ResolveChildSessionIDs
	children, err := hook.ResolveChildSessionIDs([]string{"s_parent"})
	if err != nil {
		t.Fatalf("ResolveChildSessionIDs failed: %v", err)
	}
	if len(children) != 1 || children[0] != "s_child" {
		t.Fatalf("expected ['s_child'], got %v", children)
	}

	// Test FetchSessionTraceMulti
	msgs, parts, err := hook.FetchSessionTraceMulti([]string{"s_parent", "s_child"})
	if err != nil {
		t.Fatalf("FetchSessionTraceMulti failed: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
	if msgs[0].ID != "m_p1" || msgs[1].ID != "m_c1" {
		t.Errorf("messages not in chronological order: %+v", msgs)
	}
	if parts[0].ID != "p_p1" || parts[1].ID != "p_c1" {
		t.Errorf("parts not in chronological order: %+v", parts)
	}
}

func TestLiveTraceIncludesSkillAndExplicitTaskChildAttributes(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	if _, err := rwDB.Exec(`INSERT INTO session (id,parent_id,agent,model,title,time_created,time_updated) VALUES
		('parent',NULL,'orchestrator','model','parent',?,?),('child','parent','coder','model','child',?,?)`, now, now, now+1, now+1); err != nil {
		t.Fatal(err)
	}
	if _, err := rwDB.Exec(`INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('message','parent',?,?,?)`, now+2, now+3, `{"role":"assistant","content":"working"}`); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ id, data string }{
		{"skill", `{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"agentic-lens-dev"},"output":"loaded"}}`},
		{"task", `{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"coder","description":"Implement feature"},"output":{"session_id":"child"}}}`},
	} {
		if _, err := rwDB.Exec(`INSERT INTO part (id,session_id,message_id,time_created,time_updated,data) VALUES (?, 'parent','message',?,?,?)`, row.id, now+4, now+5, row.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := rwDB.Close(); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer hook.Close()
	cfg := testProjectConfig()
	cfg.Live.PollIntervalMs = 50
	service := live.NewLiveService(hook, cfg)
	defer service.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	trace, err := service.GetSessionTrace(ctx, "parent")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.Messages) != 1 || len(trace.Messages[0].Parts) != 2 {
		t.Fatalf("unexpected live trace: %+v", trace)
	}
	if trace.Messages[0].Parts[0].SkillName != "agentic-lens-dev" || trace.Messages[0].Parts[1].SubagentType != "coder" || trace.Messages[0].Parts[1].ChildSessionID != "child" {
		t.Fatalf("skill/task metadata missing from live trace: %+v", trace.Messages[0].Parts)
	}
	payload, err := json.Marshal(trace)
	if err != nil || !strings.Contains(string(payload), `"skill_name":"agentic-lens-dev"`) || !strings.Contains(string(payload), `"child_session_id":"child"`) {
		t.Fatalf("serialized trace did not preserve task metadata: %s err=%v", payload, err)
	}
	events, unsubscribe := service.Subscribe()
	defer unsubscribe()
	service.Start(ctx)
	select {
	case event := <-events:
		for event.Type != live.LiveEventTraceUpdate {
			select {
			case event = <-events:
			case <-ctx.Done():
				t.Fatal("timed out waiting for trace_update SSE payload")
			}
		}
		if !strings.Contains(string(event.Data), `"skill_name":"agentic-lens-dev"`) || !strings.Contains(string(event.Data), `"child_session_id":"child"`) {
			t.Fatalf("trace_update payload omitted skill/task metadata: %s", event.Data)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for trace_update SSE payload")
	}
}

func TestLiveTraceIgnoresMalformedOptionalSkillTaskInput(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	if _, err := rwDB.Exec(`INSERT INTO session (id,parent_id,agent,model,title,time_created,time_updated) VALUES ('optional','root','coder','model','optional',?,?)`, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := rwDB.Exec(`INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('optional-message','optional',?,?,?)`, now+1, now+2, `{"role":"assistant","content":"working"}`); err != nil {
		t.Fatal(err)
	}
	rows := []struct{ id, data string }{
		{"bad-skill", `{"type":"tool","tool":"skill","state":{"status":"completed","input":"{bad"}}`},
		{"missing-name", `{"type":"tool","tool":"skill","state":{"status":"completed","input":{}}}`},
		{"partial-task", `{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"coder"}}}`},
	}
	for i, row := range rows {
		if _, err := rwDB.Exec(`INSERT INTO part (id,session_id,message_id,time_created,time_updated,data) VALUES (?, 'optional','optional-message',?,?,?)`, row.id, now+3+int64(i), now+4+int64(i), row.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := rwDB.Close(); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer hook.Close()
	cfg := testProjectConfig()
	cfg.Live.PollIntervalMs = 50
	service := live.NewLiveService(hook, cfg)
	defer service.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	trace, err := service.GetSessionTrace(ctx, "optional")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.Messages) != 1 || len(trace.Messages[0].Parts) != 3 {
		t.Fatalf("malformed optional inputs disrupted trace retrieval: %+v", trace)
	}
	parts := trace.Messages[0].Parts
	if parts[0].SkillName != "" || parts[1].SkillName != "" || parts[2].SubagentType != "coder" {
		t.Fatalf("unexpected parsed optional attributes: %+v", parts)
	}
	payload, err := json.Marshal(trace)
	if err != nil {
		t.Fatalf("malformed optional input broke trace serialization: %v", err)
	}
	events, unsubscribe := service.Subscribe()
	defer unsubscribe()
	service.Start(ctx)
	for {
		select {
		case event := <-events:
			if event.Type != live.LiveEventTraceUpdate {
				continue
			}
			if !strings.Contains(string(event.Data), `"subagent_type":"coder"`) {
				t.Fatalf("trace_update omitted the valid partial task field: %s", event.Data)
			}
			return
		case <-ctx.Done():
			t.Fatalf("timed out waiting for trace_update; serialized trace: %s", payload)
		}
	}
}

func TestLiveTraceSurvivesChildSessionLookupFailure(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	if _, err := rwDB.Exec(`INSERT INTO session (id,parent_id,agent,model,title,time_created,time_updated) VALUES ('parent',NULL,'orchestrator','model','parent',?,?)`, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := rwDB.Exec(`INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('message','parent',?,?,?)`, now+1, now+2, `{"role":"assistant","content":"working"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := rwDB.Exec(`INSERT INTO part (id,session_id,message_id,time_created,time_updated,data) VALUES ('task','parent','message',?,?,?)`, now+3, now+4, `{"type":"tool","tool":"task","state":{"status":"completed","input":{"subagent_type":"coder","description":"Implement feature"},"output":{"session_id":"child"}}}`); err != nil {
		t.Fatal(err)
	}
	// Keep required trace tables usable while making the optional child lookup fail.
	if _, err := rwDB.Exec(`ALTER TABLE session DROP COLUMN parent_id`); err != nil {
		t.Fatalf("failed to arrange child lookup failure: %v", err)
	}
	if err := rwDB.Close(); err != nil {
		t.Fatal(err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatal(err)
	}
	defer hook.Close()
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	trace, err := service.GetSessionTrace(context.Background(), "parent")
	if err != nil {
		t.Fatalf("GetSessionTrace failed on optional child lookup error: %v", err)
	}
	if len(trace.Messages) != 1 || len(trace.Messages[0].Parts) != 1 {
		t.Fatalf("trace data was not preserved: %+v", trace)
	}
	taskPart := trace.Messages[0].Parts[0]
	if taskPart.SubagentType != "coder" || taskPart.ChildSessionID != "" {
		t.Fatalf("expected task attributes without an unresolved child id, got %+v", taskPart)
	}
}

func TestHydrateWorkspaceEvents_SuccessAndZeroDiskWrites(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	workspacesDir := t.TempDir()

	now := time.Now().UnixMilli()

	// Seed session
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('s_al33', NULL, 'orchestrator', 'gemini-3.8-flash', 'Feature AL-33 Event Hydration', ?, ?);
	`, now, now)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	// Seed user prompt followed by assistant message with reasoning and tokens
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES
			('m_user', 's_al33', ?, ?, '{"role":"user","content":"Implement dual-plane waterfall"}'),
			('m_asst', 's_al33', ?, ?, '{"role":"assistant","content":"Implementing now","reasoning":"Analyzing schema","tokens":{"input":1000,"output":200,"cache":{"read":500},"reasoning":50}}');
	`, now+100, now+100, now+200, now+300)
	if err != nil {
		t.Fatalf("seeding messages: %v", err)
	}

	// Seed tool call and compaction parts
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES
			('p_tool', 's_al33', 'm_asst', ?, ?, '{"type":"tool","tool":"write","state":{"status":"completed","input":{"path":"models.go"},"output":"file written"}}'),
			('p_comp', 's_al33', 'm_asst', ?, ?, '{"type":"compaction"}');
	`, now+220, now+250, now+260, now+260)
	if err != nil {
		t.Fatalf("seeding parts: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx := context.Background()

	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				Role: "orchestrator",
			},
		}, nil, []domain.EventFact{
			{
				Timestamp: "2026-10-19T10:00:00Z",
				SessionID: "s_al33",
				Turn:      1,
				Role:      "orchestrator",
				EventType: domain.EventTypeCompaction,
				Status:    "completed",
			},
		}),
		},
	}

	// Capture workspace directory files before hydration
	beforeEntries, err := os.ReadDir(workspacesDir)
	if err != nil {
		t.Fatalf("reading workspaces dir: %v", err)
	}

	resp, err := service.HydrateWorkspaceEvents(ctx, "AL-33", ledger)
	if err != nil {
		t.Fatalf("HydrateWorkspaceEvents returned error: %v", err)
	}

	if !resp.IsHydrated {
		t.Errorf("expected IsHydrated == true, got false")
	}
	if resp.DataSource != domain.DataSourceLocalSQLite {
		t.Errorf("expected DataSource == %q, got %q", domain.DataSourceLocalSQLite, resp.DataSource)
	}
	if resp.TicketID != "AL-33" {
		t.Errorf("expected TicketID == %q, got %q", "AL-33", resp.TicketID)
	}

	// Verify events
	if len(resp.Events) < 3 {
		t.Fatalf("expected at least 3 events, got %d: %+v", len(resp.Events), resp.Events)
	}

	// Find turn_message
	var foundTurnMsg, foundToolCall, foundCompaction bool
	for _, ev := range resp.Events {
		if ev.SessionID != "s_al33" || ev.Turn != 1 {
			t.Errorf("hydrated event lost session/turn provenance: %+v", ev)
		}
		if ev.EventType == domain.EventTypeTurnMessage && ev.EventID != "opencode:message:m_asst" {
			t.Errorf("hydrated turn lost source message identity: %+v", ev)
		}
		if ev.EventType == domain.EventTypeToolCall && (ev.MessageID == nil || *ev.MessageID != "m_asst") {
			t.Errorf("hydrated tool lost source message identity: %+v", ev)
		}
		if ev.EventType == domain.EventTypeTurnMessage {
			foundTurnMsg = true
			if ev.PromptContent == nil || *ev.PromptContent != "Implement dual-plane waterfall" {
				t.Errorf("turn message prompt content mismatch: %v", ev.PromptContent)
			}
			if ev.ReasoningContent == nil || *ev.ReasoningContent != "Analyzing schema" {
				t.Errorf("turn message reasoning content mismatch: %v", ev.ReasoningContent)
			}
			if ev.ContextTokens != 1500 { // 1000 input + 500 cache
				t.Errorf("context tokens mismatch: expected 1500, got %d", ev.ContextTokens)
			}
		}
		if ev.EventType == domain.EventTypeToolCall {
			foundToolCall = true
			if ev.ToolName == nil || *ev.ToolName != "write" {
				t.Errorf("tool name mismatch: %v", ev.ToolName)
			}
			if ev.ToolInput == nil || !strings.Contains(*ev.ToolInput, "models.go") {
				t.Errorf("tool input mismatch: %v", ev.ToolInput)
			}
			if ev.ToolOutput == nil || *ev.ToolOutput != "file written" {
				t.Errorf("tool output mismatch: %v", ev.ToolOutput)
			}
		}
		if ev.EventType == domain.EventTypeCompaction {
			foundCompaction = true
		}
	}

	if !foundTurnMsg {
		t.Errorf("expected to find turn_message event")
	}
	if !foundToolCall {
		t.Errorf("expected to find tool_call event")
	}
	if !foundCompaction {
		t.Errorf("expected to find compaction event")
	}

	// Verify zero disk persistence in workspacesDir
	afterEntries, err := os.ReadDir(workspacesDir)
	if err != nil {
		t.Fatalf("reading workspaces dir after: %v", err)
	}
	if len(beforeEntries) != len(afterEntries) {
		t.Errorf("zero disk persistence violation: file count changed in workspaces dir (%d -> %d)", len(beforeEntries), len(afterEntries))
	}
}

func TestHydrateWorkspaceEvents_CodexUsesExplicitSourceAndCODEXHome(t *testing.T) {
	home := t.TempDir()
	rolloutPath := filepath.Join(home, "sessions", "2026", "10", "26", "rollout-thread.jsonl")
	if err := os.MkdirAll(filepath.Dir(rolloutPath), 0700); err != nil {
		t.Fatal(err)
	}
	rollout := "{\"type\":\"session_meta\",\"payload\":{\"id\":\"thread-codex\",\"session_id\":\"thread-codex\",\"timestamp\":\"2026-10-26T09:00:00Z\"}}\n" +
		"{\"timestamp\":\"2026-10-26T09:00:01Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"item_completed\",\"turn_id\":\"turn-1\",\"item\":{\"id\":\"cmd-1\",\"type\":\"CommandExecution\",\"status\":\"failed\",\"exit_code\":1}}}\n"
	if err := os.WriteFile(rolloutPath, []byte(rollout), 0600); err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
	defer service.Close()
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
		ID: "thread-codex", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "orchestrator"}},
	}}}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CODEX", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsHydrated || resp.DataSource != domain.DataSourceLocalCodex {
		t.Fatalf("expected local Codex hydration, got %#v", resp)
	}
	if len(resp.Events) != 1 || resp.Events[0].EventID == "" {
		t.Fatalf("expected one transient stable-ID event, got %#v", resp.Events)
	}
}

func TestHydrateWorkspaceEventsMixedSourcesPreserveSQLiteDetailsOverCommittedEcho(t *testing.T) {
	for _, tc := range []struct {
		name         string
		includeCodex bool
	}{
		{name: "SQLite only"},
		{name: "SQLite and Codex", includeCodex: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dbPath := createTestDBFile(t)
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			created := time.Date(2026, 10, 29, 10, 0, 0, 0, time.UTC).UnixMilli()
			if _, err := db.Exec(`INSERT INTO session (id, agent, model, title, time_created, time_updated)
				VALUES ('sqlite-session', 'coder', 'test-model', 'FIX-12', ?, ?)`, created, created+25); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data)
				VALUES ('sqlite-message', 'sqlite-session', ?, ?, '{"role":"assistant","content":"Run checks","tokens":{"input":1200}}')`, created, created+25); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
				VALUES ('p1', 'sqlite-session', 'sqlite-message', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"completed","input":"make test","output":"SQLite-local output"}}')`, created, created+25); err != nil {
				t.Fatal(err)
			}
			hook, err := live.OpenReadOnlyHook(dbPath, 5000)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := hook.Close(); err != nil {
					t.Error(err)
				}
			})
			home := t.TempDir()
			tool, recordedMessage, zero := "bash", "committed-message", int64(0)
			facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
				ID: "sqlite-session", Source: domain.SourceOpenCode, User: "tester",
				Invocations: []domain.InvocationFact{{Role: "coder"}}, Tools: []domain.SessionToolFact{},
				Events: []domain.SessionEventFact{
					{EventID: "opencode:part:p1", Timestamp: "2026-10-29T10:00:00.000000+00:00", MessageID: &recordedMessage,
						Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "completed", DurationMs: &zero, ContextTokens: 1200},
					{EventID: "committed-only", Timestamp: "2026-10-29T09:00:00Z", Turn: 0, Role: "tester",
						EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: 800},
				},
			}}}
			if tc.includeCodex {
				rolloutPath := filepath.Join(home, "sessions", "2026", "10", "29", "rollout-thread.jsonl")
				if err := os.MkdirAll(filepath.Dir(rolloutPath), 0700); err != nil {
					t.Fatal(err)
				}
				rollout := `{"type":"session_meta","payload":{"id":"codex-session","session_id":"codex-session","timestamp":"2026-10-29T10:00:01Z"}}` + "\n" +
					`{"timestamp":"2026-10-29T10:00:02Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-1","item":{"id":"cmd-1","type":"CommandExecution","status":"failed","exit_code":1}}}` + "\n"
				if err := os.WriteFile(rolloutPath, []byte(rollout), 0600); err != nil {
					t.Fatal(err)
				}
				facts.Sessions = append(facts.Sessions, domain.SessionFact{
					ID: "codex-session", Source: domain.SourceCodex, User: "tester", Invocations: []domain.InvocationFact{{Role: "tester"}},
				})
			}
			before, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			service := live.NewLiveServiceWithCodexHome(hook, testProjectConfig(), home)
			t.Cleanup(func() {
				if err := service.Close(); err != nil {
					t.Error(err)
				}
			})
			ledger := domain.AgenticMetricsLedger{Facts: facts}
			response, err := service.HydrateWorkspaceEvents(context.Background(), "FIX-12", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if !response.IsHydrated || !response.IncludesBaseline || response.DataSource != domain.DataSourceLocalSQLite {
				t.Fatalf("incorrect combined-source provenance: %#v", response)
			}
			wantCount := 3
			if tc.includeCodex {
				wantCount++
			}
			if len(response.Events) != wantCount {
				t.Fatalf("expected %d baseline/local events, got %#v", wantCount, response.Events)
			}
			matched, unmatched, codex := 0, 0, 0
			for _, event := range response.Events {
				switch {
				case event.EventID == "opencode:part:p1":
					matched++
					if event.DurationMs == nil || *event.DurationMs != 25 || event.MessageID == nil || *event.MessageID != "sqlite-message" {
						t.Fatalf("committed echo overwrote SQLite duration/message ID: %#v", event)
					}
					if event.ToolInput == nil || *event.ToolInput != "make test" || event.ToolOutput == nil || *event.ToolOutput != "SQLite-local output" {
						t.Fatalf("combined source lost SQLite payloads: %#v", event)
					}
				case event.EventID == "committed-only":
					unmatched++
					if !reflect.DeepEqual(event, facts.EventFacts()[0]) {
						t.Fatalf("unmatched baseline changed: %#v", event)
					}
				case event.SessionID == "codex-session":
					codex++
					if event.EventID == "" || event.EventType != domain.EventTypeToolCall {
						t.Fatalf("Codex source occurrence lost: %#v", event)
					}
				}
			}
			if matched != 1 || unmatched != 1 || (tc.includeCodex && codex != 1) || (!tc.includeCodex && codex != 0) {
				t.Fatalf("combined trace lost or duplicated source occurrences: %#v", response.Events)
			}
			repeated, err := service.HydrateWorkspaceEvents(context.Background(), "FIX-12", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(repeated, response) {
				t.Fatalf("repeated source hydration changed the trace: %#v", repeated)
			}
			after, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatal("source enrichment mutated the committed input")
			}
		})
	}
}

func TestHydrateWorkspaceEvents_CodexSucceedsWithMultipleInvocationRows(t *testing.T) {
	home := t.TempDir()
	rolloutPath := filepath.Join(home, "sessions", "2026", "10", "26", "rollout-thread.jsonl")
	if err := os.MkdirAll(filepath.Dir(rolloutPath), 0700); err != nil {
		t.Fatal(err)
	}
	rollout := "{\"type\":\"session_meta\",\"payload\":{\"id\":\"thread-codex\",\"session_id\":\"thread-codex\",\"timestamp\":\"2026-10-26T09:00:00Z\"}}\n" +
		"{\"timestamp\":\"2026-10-26T09:00:01Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"item_completed\",\"turn_id\":\"turn-1\",\"item\":{\"id\":\"cmd-1\",\"type\":\"CommandExecution\",\"status\":\"failed\",\"exit_code\":1}}}\n"
	if err := os.WriteFile(rolloutPath, []byte(rollout), 0600); err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
	defer service.Close()

	// Multiple model rows belong to the same Codex session.
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
		ID:     "thread-codex",
		Source: domain.SourceCodex,
		User:   "alice",
		Invocations: []domain.InvocationFact{
			{Role: "orchestrator"},
			{Role: "orchestrator"},
		},
	}}}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CODEX-DISTINCT", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsHydrated || resp.DataSource != domain.DataSourceLocalCodex {
		t.Fatalf("expected local Codex hydration with distinct model rows, got %#v", resp)
	}
	if len(resp.Events) != 1 || resp.Events[0].EventID == "" {
		t.Fatalf("expected one transient event, got %#v", resp.Events)
	}
}

func TestHydrateWorkspaceEvents_AnomalyHydrationRetainsSessionHierarchy(t *testing.T) {
	home := t.TempDir()
	write := func(relativePath, id, parent string) {
		t.Helper()
		path := filepath.Join(home, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		parentField := ""
		if parent != "" {
			parentField = `,"parent_thread_id":"` + parent + `"`
		}
		content := `{"type":"session_meta","payload":{"id":"` + id + `","session_id":"` + id + `"` + parentField + `,"timestamp":"2026-10-26T09:00:00Z"}}` + "\n" +
			`{"timestamp":"2026-10-26T09:00:01Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-1","item":{"id":"cmd-1","type":"CommandExecution","status":"completed","exit_code":0}}}` + "\n"
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("sessions/2026/10/26/rollout-root.jsonl", "root-thread", "")
	write("sessions/2026/10/26/rollout-child.jsonl", "child-thread", "root-thread")

	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
	defer service.Close()

	parent := "root-thread"
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{
		{
			ID:     "root-thread",
			Source: domain.SourceCodex,
			User:   "lead",
			Invocations: []domain.InvocationFact{
				{Role: "orchestrator"},
			},
		},
		{
			ID:       "child-thread",
			ParentID: &parent,
			Source:   domain.SourceCodex,
			User:     "subagent",
			Invocations: []domain.InvocationFact{
				{Role: "coder"},
			},
		},
	}}}

	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CODEX-HIERARCHY", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsHydrated || resp.DataSource != domain.DataSourceLocalCodex {
		t.Fatalf("expected successful hydration, got %#v", resp)
	}
	if len(resp.Events) != 2 {
		t.Fatalf("expected 2 hydrated events, got %d", len(resp.Events))
	}
}

func TestHydrateWorkspaceEvents_SessionRoleMapUsesSessionID(t *testing.T) {
	dbPath := createTestDBFile(t)
	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	now := time.Now().UnixMilli()
	if _, err := rwDB.Exec("INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES ('s_custom', NULL, NULL, 'm1', 'title', ?, ?)", now, now); err != nil {
		rwDB.Close()
		t.Fatal(err)
	}
	if _, err := rwDB.Exec("INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m_custom', 's_custom', ?, ?, '{\"role\":\"assistant\",\"content\":\"msg\"}')", now, now); err != nil {
		rwDB.Close()
		t.Fatal(err)
	}
	rwDB.Close()

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{
			Sessions: []domain.SessionFact{
				{
					ID:     "s_custom",
					Source: domain.SourceOpenCode,
					Invocations: []domain.InvocationFact{
						{
							Role: "planner",
						},
					},
				},
			},
		},
	}

	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CUSTOM", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.IsHydrated {
		t.Fatalf("expected hydrated response, got %#v", resp)
	}
	if len(resp.Events) == 0 {
		t.Fatalf("expected hydrated events, got 0")
	}
	if resp.Events[0].Role != "planner" {
		t.Errorf("expected role 'planner' from sessionRoleMap, got %q", resp.Events[0].Role)
	}
}

func TestHydrateWorkspaceEvents_CodexFailsClosedWhenAnyRootIsMissing(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "sessions", "2026", "10", "26", "rollout-present.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	rollout := "{\"type\":\"session_meta\",\"payload\":{\"id\":\"root-present\",\"session_id\":\"root-present\",\"timestamp\":\"2026-10-26T09:00:00Z\"}}\n" +
		"{\"timestamp\":\"2026-10-26T09:00:01Z\",\"type\":\"event_msg\",\"payload\":{\"type\":\"item_completed\",\"turn_id\":\"turn-1\",\"item\":{\"id\":\"cmd-1\",\"type\":\"CommandExecution\",\"status\":\"failed\",\"exit_code\":1}}}\n"
	if err := os.WriteFile(path, []byte(rollout), 0600); err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
	defer service.Close()
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{
		{ID: "root-present", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "orchestrator"}}},
		{ID: "root-missing", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "orchestrator"}}},
	}}}
	ledger.Facts.Sessions[0].Events = []domain.SessionEventFact{{EventID: "committed-root", Timestamp: "2026-10-26T08:00:00Z", Turn: 1, Role: "orchestrator", EventType: domain.EventTypeTurnMessage, Status: "completed"}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsHydrated || resp.DataSource != domain.DataSourceCommittedLedger || !reflect.DeepEqual(resp.Events, ledger.Facts.EventFacts()) {
		t.Fatalf("partial Codex roots were reported as hydrated: %#v", resp)
	}
	missingPath := filepath.Join(filepath.Dir(path), "rollout-missing.jsonl")
	missingRollout := `{"type":"session_meta","payload":{"id":"root-missing","session_id":"root-missing","timestamp":"2026-10-26T09:00:00Z"}}` + "\n"
	if err := os.WriteFile(missingPath, []byte(missingRollout), 0600); err != nil {
		t.Fatal(err)
	}
	complete, err := service.HydrateWorkspaceEvents(context.Background(), "AL-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !complete.IsHydrated || complete.DataSource != domain.DataSourceLocalCodex || len(complete.Events) != 2 || complete.Events[0].EventID != "committed-root" {
		t.Fatalf("complete Codex roots failed to hydrate after restoring the missing rollout: %#v", complete)
	}
}

func TestHydrateWorkspaceEvents_CodexFailsClosedWhenChildRolloutIsMissing(t *testing.T) {
	home := t.TempDir()
	write := func(relativePath, id, parent string, withEvent bool) {
		t.Helper()
		path := filepath.Join(home, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		parentField := ""
		if parent != "" {
			parentField = `,"parent_thread_id":"` + parent + `"`
		}
		content := `{"type":"session_meta","payload":{"id":"` + id + `","session_id":"` + id + `"` + parentField + `,"timestamp":"2026-10-26T09:00:00Z"}}` + "\n"
		if withEvent {
			content += `{"timestamp":"2026-10-26T09:00:01Z","type":"event_msg","payload":{"type":"item_completed","turn_id":"turn-1","item":{"id":"cmd-1","type":"CommandExecution","status":"failed","exit_code":1}}}` + "\n"
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("sessions/2026/10/26/rollout-root.jsonl", "root-thread", "", true)
	write("sessions/2026/10/26/rollout-child-one.jsonl", "child-one", "root-thread", false)
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
	defer service.Close()
	parent := "root-thread"
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{
		{ID: "root-thread", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "orchestrator"}}},
		{ID: "child-one", Source: domain.SourceCodex, ParentID: &parent, Invocations: []domain.InvocationFact{{Role: "coder"}}},
		{ID: "child-two", Source: domain.SourceCodex, ParentID: &parent, Invocations: []domain.InvocationFact{{Role: "tester"}}},
	}}}
	ledger.Facts.Sessions[1].Events = []domain.SessionEventFact{{EventID: "committed-child", Timestamp: "2026-10-26T08:00:00Z", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, Status: "error"}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CHILD-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsHydrated || resp.DataSource != domain.DataSourceCommittedLedger || !reflect.DeepEqual(resp.Events, ledger.Facts.EventFacts()) {
		t.Fatalf("partial Codex child tree was reported as hydrated: %#v", resp)
	}
	write("sessions/2026/10/26/rollout-child-two.jsonl", "child-two", "root-thread", false)
	complete, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CHILD-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !complete.IsHydrated || complete.DataSource != domain.DataSourceLocalCodex || len(complete.Events) != 2 || complete.Events[0].EventID != "committed-child" {
		t.Fatalf("complete Codex child tree failed to hydrate after restoring the missing rollout: %#v", complete)
	}
}

func TestHydrateWorkspaceEvents_SourceLessLedgerStaysUnhydrated(t *testing.T) {
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), t.TempDir())
	defer service.Close()
	ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
		ID: "legacy", Events: []domain.SessionEventFact{{EventID: "old", Timestamp: "2026-10-26T09:00:00Z", EventType: domain.EventTypeToolCall, Status: "error"}},
		Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}},
	}}}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-LEGACY", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsHydrated || resp.DataSource != domain.DataSourceCommittedLedger || !reflect.DeepEqual(resp.Events, ledger.Facts.EventFacts()) {
		t.Fatalf("source-less ledger was incorrectly hydrated: %#v", resp)
	}
	if len(resp.Anomalies) != 1 {
		t.Fatalf("stored summary should remain available: %#v", resp.Anomalies)
	}
}

func TestHydrationFailuresRetainRecordedEventsAndLedgerBytes(t *testing.T) {
	for _, mode := range []string{"nil service", "nil hook", "missing session", "closed database", "missing tables", "empty SQLite session", "missing Codex home", "missing Codex rollout", "empty Codex rollout"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			cfg := testProjectConfig()
			cfg.Workspaces.Dir = root
			source := domain.SourceOpenCode
			if strings.Contains(mode, "Codex") {
				source = domain.SourceCodex
			}
			facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
				ID: "recorded-session", Source: source, User: "tester", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{},
				Events: []domain.SessionEventFact{
					{EventID: "later", Timestamp: "2026-10-29T11:00:00Z", Turn: 2, Role: "tester", EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: 2400},
					{Timestamp: "2026-10-29T10:00:00Z", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, Status: "error", ContextTokens: 1200},
				},
				Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}},
			}}}
			ledger := domain.AgenticMetricsLedger{Facts: facts}
			before, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			ledgerPath := filepath.Join(root, "agentic_metrics.json")
			if err := os.WriteFile(ledgerPath, before, 0600); err != nil {
				t.Fatal(err)
			}
			var service *live.LiveService
			switch mode {
			case "nil service":
			case "nil hook":
				service = live.NewLiveService(nil, cfg)
			case "missing Codex home", "missing Codex rollout", "empty Codex rollout":
				home := filepath.Join(t.TempDir(), "codex")
				if mode != "missing Codex home" {
					if err := os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "empty Codex rollout" {
					content := `{"type":"session_meta","payload":{"id":"recorded-session","session_id":"recorded-session","timestamp":"2026-10-29T10:00:00Z"}}` + "\n"
					if err := os.WriteFile(filepath.Join(home, "sessions", "rollout-empty.jsonl"), []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				service = live.NewLiveServiceWithCodexHome(nil, cfg, home)
			default:
				dbPath := createTestDBFile(t)
				if mode == "missing tables" || mode == "empty SQLite session" {
					db, err := sql.Open("sqlite", dbPath)
					if err != nil {
						t.Fatal(err)
					}
					statement := "DROP TABLE message; DROP TABLE part;"
					if mode == "empty SQLite session" {
						statement = `INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES ('recorded-session', NULL, 'coder', 'test-model', 'FIX-12', 1, 1)`
					}
					if _, err := db.Exec(statement); err != nil {
						t.Fatal(err)
					}
					if err := db.Close(); err != nil {
						t.Fatal(err)
					}
				}
				hook, err := live.OpenReadOnlyHook(dbPath, 5000)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "closed database" {
					if err := hook.Close(); err != nil {
						t.Fatal(err)
					}
				} else {
					t.Cleanup(func() {
						if err := hook.Close(); err != nil {
							t.Error(err)
						}
					})
				}
				service = live.NewLiveService(hook, cfg)
			}
			if service != nil {
				t.Cleanup(func() {
					if err := service.Close(); err != nil {
						t.Error(err)
					}
				})
			}
			response, err := service.HydrateWorkspaceEvents(context.Background(), "FIX-12", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if response.TicketID != "FIX-12" || response.IsHydrated || response.DataSource != domain.DataSourceCommittedLedger || !response.IncludesBaseline || !reflect.DeepEqual(response.Events, facts.EventFacts()) || !reflect.DeepEqual(response.Anomalies, facts.AnomalyFacts()) {
				t.Fatalf("local failure lost recorded facts or provenance: %#v", response)
			}
			after, err := os.ReadFile(ledgerPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(before) {
				t.Fatal("hydration fallback rewrote committed ledger bytes")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("hydration wrote workspace files: %#v", entries)
			}
			unchanged, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			if string(unchanged) != string(before) {
				t.Fatal("hydration mutated the ledger input")
			}
		})
	}
}

func TestHydrateWorkspaceEvents_FallbackWhenSessionNotFound(t *testing.T) {
	dbPath := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx := context.Background()

	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				Role: "orchestrator",
			},
		}, nil, []domain.EventFact{
			{
				Timestamp: "2026-10-19T10:00:00Z",
				SessionID: "s_missing",
				Turn:      1,
				Role:      "orchestrator",
				EventType: domain.EventTypeCompaction,
				Status:    "completed",
			},
		}),
		},
	}

	resp, err := service.HydrateWorkspaceEvents(ctx, "AL-UNKNOWN", ledger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsHydrated {
		t.Errorf("expected IsHydrated == false for missing session, got true")
	}
	if resp.DataSource != domain.DataSourceCommittedLedger {
		t.Errorf("expected DataSource == %q, got %q", domain.DataSourceCommittedLedger, resp.DataSource)
	}
	if !reflect.DeepEqual(resp.Events, ledger.Facts.EventFacts()) {
		t.Fatalf("missing local session discarded recorded facts: %#v", resp.Events)
	}
}

func TestHydrateWorkspaceEvents_HostileDegradationAndFailureModes(t *testing.T) {
	ctx := context.Background()
	workspacesDir := t.TempDir()

	ledgerWithFallback := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{Role: "coder"},
		}, nil, []domain.EventFact{
			{
				Timestamp: "2026-10-19T10:00:00Z",
				SessionID: "ses-hostile",
				Turn:      1,
				Role:      "coder",
				EventType: domain.EventTypeCompaction,
				Status:    "completed",
			},
		}),
		},
	}

	// 1. Nil LiveService returns fallback without panic
	t.Run("nil LiveService returns fallback cleanly", func(t *testing.T) {
		var nilService *live.LiveService
		resp, err := nilService.HydrateWorkspaceEvents(ctx, "AL-NIL", ledgerWithFallback)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false")
		}
		if resp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %q, got %q", domain.DataSourceCommittedLedger, resp.DataSource)
		}
		if !reflect.DeepEqual(resp.Events, ledgerWithFallback.Facts.EventFacts()) {
			t.Errorf("nil service discarded recorded fallback events: %#v", resp.Events)
		}
	})

	// 2. LiveService with nil hook
	t.Run("LiveService with nil hook returns fallback cleanly", func(t *testing.T) {
		service := live.NewLiveService(nil, testProjectConfig())
		resp, err := service.HydrateWorkspaceEvents(ctx, "AL-NIL-HOOK", ledgerWithFallback)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false")
		}
		if resp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %q, got %q", domain.DataSourceCommittedLedger, resp.DataSource)
		}
	})

	// 3. LiveService with closed DB handle
	t.Run("LiveService with closed DB handle returns fallback cleanly", func(t *testing.T) {
		dbPath := createTestDBFile(t)
		hook, err := live.OpenReadOnlyHook(dbPath, 5000)
		if err != nil {
			t.Fatalf("failed opening hook: %v", err)
		}
		service := live.NewLiveService(hook, testProjectConfig())
		// Close hook and DB
		_ = hook.Close()

		resp, err := service.HydrateWorkspaceEvents(ctx, "AL-CLOSED-DB", ledgerWithFallback)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false on closed db")
		}
		if resp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %q, got %q", domain.DataSourceCommittedLedger, resp.DataSource)
		}
	})

	// 4. LiveService when database tables are missing or dropped
	t.Run("LiveService when tables are dropped returns fallback cleanly", func(t *testing.T) {
		dbPath := createTestDBFile(t)
		rwDB, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatalf("failed opening rw db: %v", err)
		}

		// Drop message and part tables
		_, _ = rwDB.Exec("DROP TABLE IF EXISTS message; DROP TABLE IF EXISTS part;")
		rwDB.Close()

		hook, err := live.OpenReadOnlyHook(dbPath, 5000)
		if err != nil {
			t.Fatalf("failed opening hook: %v", err)
		}
		defer hook.Close()

		service := live.NewLiveService(hook, testProjectConfig())
		defer service.Close()

		resp, err := service.HydrateWorkspaceEvents(ctx, "AL-DROPPED-TABLES", ledgerWithFallback)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false when tables dropped")
		}
		if resp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %q, got %q", domain.DataSourceCommittedLedger, resp.DataSource)
		}
	})

	// 5. Corrupted / malformed JSON inside message and part records
	t.Run("LiveService handles corrupted JSON in SQLite records gracefully", func(t *testing.T) {
		dbPath := createTestDBFile(t)
		rwDB, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatalf("failed opening rw db: %v", err)
		}
		defer rwDB.Close()

		now := time.Now().UnixMilli()

		_, err = rwDB.Exec(`
			INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
			VALUES ('s_corrupt', NULL, 'coder', 'gemini-3.8-flash', 'Corrupted AL-33', ?, ?);
		`, now, now)
		if err != nil {
			t.Fatalf("seeding session: %v", err)
		}

		// Seed malformed message JSON and malformed part JSON
		_, err = rwDB.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data)
			VALUES ('m_corrupt', 's_corrupt', ?, ?, '{"role": "assistant", "content": [truncated');
		`, now+100, now+150)
		if err != nil {
			t.Fatalf("seeding message: %v", err)
		}

		_, err = rwDB.Exec(`
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
			VALUES ('p_corrupt', 's_corrupt', 'm_corrupt', ?, ?, '{"type": "tool", "state": corrupted json');
		`, now+110, now+120)
		if err != nil {
			t.Fatalf("seeding part: %v", err)
		}

		hook, err := live.OpenReadOnlyHook(dbPath, 5000)
		if err != nil {
			t.Fatalf("failed opening hook: %v", err)
		}
		defer hook.Close()

		service := live.NewLiveService(hook, testProjectConfig())
		defer service.Close()

		corruptLedger := domain.AgenticMetricsLedger{
			Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
				ID: "s_corrupt", Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{{Role: "coder"}},
			}},
			},
		}

		// Must not panic, must produce valid response
		resp, err := service.HydrateWorkspaceEvents(ctx, "AL-CORRUPT-RECORDS", corruptLedger)
		if err != nil {
			t.Fatalf("unexpected error on corrupted JSON records: %v", err)
		}
		if !resp.IsHydrated {
			t.Errorf("expected IsHydrated == true because session exists and message record exists")
		}
		if len(resp.Events) == 0 {
			t.Errorf("expected at least 1 turn_message event even with corrupted JSON data")
		}
	})

	// 6. Completely empty ledger facts (no invocations, no events)
	t.Run("empty ledger facts returns unhydrated empty slice", func(t *testing.T) {
		dbPath := createTestDBFile(t)
		hook, err := live.OpenReadOnlyHook(dbPath, 5000)
		if err != nil {
			t.Fatalf("failed opening hook: %v", err)
		}
		defer hook.Close()

		service := live.NewLiveService(hook, testProjectConfig())
		defer service.Close()

		emptyLedger := domain.AgenticMetricsLedger{
			Facts: domain.AgenticFactLedger{},
		}

		resp, err := service.HydrateWorkspaceEvents(ctx, "AL-EMPTY", emptyLedger)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false for empty ledger")
		}
		if resp.Events == nil {
			t.Errorf("expected non-nil events slice")
		}
		if len(resp.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(resp.Events))
		}
	})

	// 7. Verify zero disk writes to workspacesDir across all hostile tests
	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		t.Fatalf("reading workspaces dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("zero disk persistence violation: files written to %s: %v", workspacesDir, entries)
	}
}

func TestLiveService_TokenExtractionAndAnomalyDetection(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES ('s_anom_1', 'Task AL-19: Live anomaly detection', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-10000, now-1000)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	// Seed message with token metrics (150k input + 20k cache.read = 170k context tokens, 5k output = 175k total tokens)
	msgData := `{"role":"assistant","tokens":{"input":150000,"output":5000,"cache":{"read":20000},"reasoning":2000}}`
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_anom_1', 's_anom_1', ?, ?, ?)
	`, now-8000, now-8000, msgData)
	if err != nil {
		t.Fatalf("seeding message: %v", err)
	}

	// Seed 3 consecutive failing tool parts
	part1Data := `{"type":"tool","tool":"bash","state":{"status":"error","output":"command exit status 1"}}`
	part2Data := `{"type":"tool","tool":"bash","state":{"status":"error","output":"command exit status 1"}}`
	part3Data := `{"type":"tool","tool":"bash","state":{"status":"error","output":"command exit status 1"}}`

	for idx, pData := range []string{part1Data, part2Data, part3Data} {
		_, err = rwDB.Exec(`
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
			VALUES (?, 's_anom_1', 'm_anom_1', ?, ?, ?)
		`, fmt.Sprintf("p_anom_%d", idx+1), now-7000+int64(idx*1000), now-7000+int64(idx*1000), pData)
		if err != nil {
			t.Fatalf("seeding part: %v", err)
		}
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Verify GetSessionTrace token parsing and anomalies
	trace, err := service.GetSessionTrace(ctx, "s_anom_1")
	if err != nil {
		t.Fatalf("GetSessionTrace failed: %v", err)
	}

	if len(trace.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(trace.Messages))
	}

	msg := trace.Messages[0]
	if msg.InputTokens != 150000 {
		t.Errorf("expected 150000 input tokens, got %d", msg.InputTokens)
	}
	if msg.CacheReadTokens != 20000 {
		t.Errorf("expected 20000 cache read tokens, got %d", msg.CacheReadTokens)
	}
	if msg.ContextTokens != 170000 {
		t.Errorf("expected 170000 context tokens, got %d", msg.ContextTokens)
	}
	if msg.OutputTokens != 5000 {
		t.Errorf("expected 5000 output tokens, got %d", msg.OutputTokens)
	}
	if msg.ReasoningTokens != 2000 {
		t.Errorf("expected 2000 reasoning tokens, got %d", msg.ReasoningTokens)
	}
	if msg.TotalTokens != 175000 {
		t.Errorf("expected 175000 total tokens, got %d", msg.TotalTokens)
	}

	if len(trace.Anomalies) == 0 {
		t.Fatal("expected trace anomalies to be populated, got 0")
	}

	// Verify that ConsecutiveToolErrors, ContextCeilingApproach, TokenVelocityRunaway, and ToolThrashing were evaluated
	ruleNames := make(map[domain.LiveAnomalyRuleName]bool)
	for _, a := range trace.Anomalies {
		ruleNames[a.RuleName] = true
		if a.RuleName == domain.RuleConsecutiveToolErrors {
			if a.Count == nil || *a.Count != 3 {
				t.Errorf("expected ConsecutiveToolErrors Count to be 3, got %v", a.Count)
			}
		}
		if a.RuleName == domain.RuleToolThrashing {
			if a.Count == nil || *a.Count != 3 {
				t.Errorf("expected ToolThrashing Count to be 3, got %v", a.Count)
			}
		}
	}
	if !ruleNames[domain.RuleConsecutiveToolErrors] {
		t.Errorf("expected ConsecutiveToolErrors anomaly to be present")
	}
	if !ruleNames[domain.RuleContextCeilingApproach] {
		t.Errorf("expected ContextCeilingApproach anomaly to be present")
	}
	if !ruleNames[domain.RuleTokenVelocityRunaway] {
		t.Errorf("expected TokenVelocityRunaway anomaly to be present")
	}
	if !ruleNames[domain.RuleToolThrashing] {
		t.Errorf("expected ToolThrashing anomaly to be present")
	}

	// 2. Verify GetSessionTree node anomalies and active_anomalies
	tree, err := service.GetSessionTree(ctx)
	if err != nil {
		t.Fatalf("GetSessionTree failed: %v", err)
	}

	if len(tree.RootSessions) != 1 {
		t.Fatalf("expected 1 root session, got %d", len(tree.RootSessions))
	}
	node := tree.RootSessions[0]
	if len(node.Anomalies) == 0 {
		t.Error("expected session node anomalies to be populated")
	}
	if len(tree.ActiveAnomalies) == 0 {
		t.Error("expected tree active_anomalies to be populated")
	}
}

func TestLiveService_AnomalyAlertBroadcastStream(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session and high token message
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES ('s_stream_1', 'Task AL-19: Streaming alerts', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-5000, now-500)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	msgData := `{"role":"assistant","tokens":{"input":220000,"output":5000,"cache":{"read":0},"reasoning":0}}`
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_stream_1', 's_stream_1', ?, ?, ?)
	`, now-3000, now-3000, msgData)
	if err != nil {
		t.Fatalf("seeding message: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	cfg := testProjectConfig()
	cfg.Live.PollIntervalMs = 50

	service := live.NewLiveService(hook, cfg)
	defer service.Close()

	eventsCh, unsubscribe := service.Subscribe()
	defer unsubscribe()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	service.Start(ctx)

	// Wait for LiveEventAnomalyAlert event
	receivedAnomaly := false
	timeout := time.After(1 * time.Second)

	for !receivedAnomaly {
		select {
		case ev := <-eventsCh:
			if ev.Type == live.LiveEventAnomalyAlert {
				receivedAnomaly = true
				var alert domain.LiveAnomalyAlert
				if err := json.Unmarshal(ev.Data, &alert); err != nil {
					t.Fatalf("failed decoding anomaly alert JSON: %v", err)
				}
				if alert.RuleName != domain.RuleContextCeilingApproach && alert.RuleName != domain.RuleTokenVelocityRunaway {
					t.Errorf("unexpected rule name in broadcast alert: %s", alert.RuleName)
				}
				if alert.Severity != domain.LiveSeverityCritical {
					t.Errorf("expected CRITICAL severity, got %s", alert.Severity)
				}
			}
		case <-timeout:
			t.Fatal("timed out waiting for LiveEventAnomalyAlert event")
		}
	}
}

func TestLiveService_CorruptedJSONAndMalformedPayloads(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES ('s_corrupt_1', 'Corrupted payloads test', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-10000, now-1000)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	// 1. Message with invalid JSON syntax
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_bad_json', 's_corrupt_1', ?, ?, ?)
	`, now-9000, now-9000, `{"role": "assistant", "tokens": {broken syntax`)
	if err != nil {
		t.Fatalf("seeding bad json message: %v", err)
	}

	// 2. Message with null tokens
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_null_tokens', 's_corrupt_1', ?, ?, ?)
	`, now-8000, now-8000, `{"role": "assistant", "tokens": null}`)
	if err != nil {
		t.Fatalf("seeding null tokens message: %v", err)
	}

	// 3. Message with non-numeric tokens
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_str_tokens', 's_corrupt_1', ?, ?, ?)
	`, now-7000, now-7000, `{"role": "assistant", "tokens": "invalid string type"}`)
	if err != nil {
		t.Fatalf("seeding string tokens message: %v", err)
	}

	// 4. Message with partial tokens (input only)
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_partial_tokens', 's_corrupt_1', ?, ?, ?)
	`, now-6000, now-6000, `{"role": "assistant", "tokens": {"input": 45000}}`)
	if err != nil {
		t.Fatalf("seeding partial tokens message: %v", err)
	}

	// 5. Part with broken JSON syntax
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_bad_part', 's_corrupt_1', 'm_partial_tokens', ?, ?, ?)
	`, now-5000, now-5000, `{"type": "tool", state:`)
	if err != nil {
		t.Fatalf("seeding bad part json: %v", err)
	}

	// 6. Part with null state
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_null_state', 's_corrupt_1', 'm_partial_tokens', ?, ?, ?)
	`, now-4000, now-4000, `{"type": "tool", "tool": "bash", "state": null}`)
	if err != nil {
		t.Fatalf("seeding null state part: %v", err)
	}

	// 7. Part with missing fields
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_empty_part', 's_corrupt_1', 'm_partial_tokens', ?, ?, ?)
	`, now-3000, now-3000, `{}`)
	if err != nil {
		t.Fatalf("seeding empty part: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Execution must not panic and must safely parse messages
	trace, err := service.GetSessionTrace(ctx, "s_corrupt_1")
	if err != nil {
		t.Fatalf("GetSessionTrace failed on corrupted data: %v", err)
	}

	if len(trace.Messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(trace.Messages))
	}

	// Verify partial tokens message: input 45000 -> context 45000, total 45000
	partialMsg := trace.Messages[3]
	if partialMsg.InputTokens != 45000 {
		t.Errorf("expected 45000 input tokens, got %d", partialMsg.InputTokens)
	}
	if partialMsg.ContextTokens != 45000 {
		t.Errorf("expected 45000 context tokens, got %d", partialMsg.ContextTokens)
	}
	if partialMsg.TotalTokens != 45000 {
		t.Errorf("expected 45000 total tokens, got %d", partialMsg.TotalTokens)
	}

	// Verify that anomalies were evaluated safely
	if trace.Anomalies == nil {
		t.Errorf("expected non-nil anomalies slice")
	}
}

func TestLiveService_PathTraversalAndMaliciousSessionIDs(t *testing.T) {
	dbPath := createTestDBFile(t)

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	hostileIDs := []string{
		"../../etc/passwd",
		"../../../var/log",
		"s1/../../etc/shadow",
		"s1/../s2",
		"/etc/passwd",
		"\\Windows\\System32",
		"s1\x00inject",
		"s1\nnewline",
		"s1\rreturn",
		"s1\tTab",
		"' OR '1'='1",
		"'; DROP TABLE session; --",
		"\" OR \"1\"=\"1",
		"",
		"   ",
	}

	for _, hostileID := range hostileIDs {
		t.Run("hostile_id_"+hostileID, func(t *testing.T) {
			_, err := service.GetSessionTrace(ctx, hostileID)
			if err == nil {
				t.Fatalf("expected error for hostile session ID %q, but got nil", hostileID)
			}
			if !strings.Contains(err.Error(), "invalid session id") {
				t.Errorf("expected invalid session id error for %q, got: %v", hostileID, err)
			}
		})
	}
}

func TestLiveService_ZeroTokensAndZeroParts(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session without messages or parts
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES ('s_empty_1', 'Zero parts session', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-5000, now-5000)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	trace, err := service.GetSessionTrace(ctx, "s_empty_1")
	if err != nil {
		t.Fatalf("GetSessionTrace failed on empty session: %v", err)
	}

	if trace.TotalToolCalls != 0 {
		t.Errorf("expected 0 total tool calls, got %d", trace.TotalToolCalls)
	}
	if trace.FailedToolCalls != 0 {
		t.Errorf("expected 0 failed tool calls, got %d", trace.FailedToolCalls)
	}
	if len(trace.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(trace.Messages))
	}
	if len(trace.RecentErrors) != 0 {
		t.Errorf("expected 0 recent errors, got %d", len(trace.RecentErrors))
	}
	if len(trace.Anomalies) != 0 {
		t.Errorf("expected 0 anomalies, got %d", len(trace.Anomalies))
	}
}

func TestHydrateWorkspaceEvents_ConsecutiveErrorsSpansTurns(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES ('s_span_1', 'Spanning Consecutive Errors', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-10000, now-1000)
	if err != nil {
		t.Fatalf("seeding session: %v", err)
	}

	// Turn 1 message with 1 error
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_span_1', 's_span_1', ?, ?, '{"role":"assistant"}')
	`, now-8000, now-8000)
	if err != nil {
		t.Fatalf("seeding message 1: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_span_1', 's_span_1', 'm_span_1', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"error","output":"failed command 1"}}')
	`, now-7900, now-7900)
	if err != nil {
		t.Fatalf("seeding part 1: %v", err)
	}

	// Turn 2 message with 2 errors
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_span_2', 's_span_1', ?, ?, '{"role":"assistant"}')
	`, now-7000, now-7000)
	if err != nil {
		t.Fatalf("seeding message 2: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_span_2', 's_span_1', 'm_span_2', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"error","output":"failed command 2"}}')
	`, now-6900, now-6900)
	if err != nil {
		t.Fatalf("seeding part 2: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_span_3', 's_span_1', 'm_span_2', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"error","output":"failed command 3"}}')
	`, now-6800, now-6800)
	if err != nil {
		t.Fatalf("seeding part 3: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx := context.Background()
	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
			ID: "s_span_1", Source: domain.SourceOpenCode,
			Invocations: []domain.InvocationFact{{Role: "coder"}},
		}},
		},
	}

	resp, err := service.HydrateWorkspaceEvents(ctx, "SPAN-1", ledger)
	if err != nil {
		t.Fatalf("HydrateWorkspaceEvents failed: %v", err)
	}

	// Filter tool call events
	var toolEvents []domain.EventFact
	for _, ev := range resp.Events {
		if ev.EventType == domain.EventTypeToolCall {
			toolEvents = append(toolEvents, ev)
		}
	}

	if len(toolEvents) != 3 {
		t.Fatalf("expected 3 tool events, got %d", len(toolEvents))
	}
	for i, te := range toolEvents {
		if te.Status != "error" {
			t.Errorf("expected tool event %d to have status error, got %s", i, te.Status)
		}
	}
}

func TestHydrateWorkspaceEvents_MultiSessionIsolationAndStreakReset(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// Seed session A and session B
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, model, time_created, time_updated)
		VALUES
			('s_iso_A', 'Session A', 'coder', 'gemini-3.8-flash', ?, ?),
			('s_iso_B', 'Session B', 'coder', 'gemini-3.8-flash', ?, ?)
	`, now-20000, now-1000, now-20000, now-1000)
	if err != nil {
		t.Fatalf("seeding sessions: %v", err)
	}

	// Session A Turn 1: 1 error
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_A_1', 's_iso_A', ?, ?, '{"role":"assistant"}')
	`, now-18000, now-18000)
	if err != nil {
		t.Fatalf("seeding msg A1: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_A_1', 's_iso_A', 'm_A_1', ?, ?, '{"type":"tool","tool":"bash","state":{"status":"error","output":"command failure 1"}}')
	`, now-17900, now-17900)
	if err != nil {
		t.Fatalf("seeding part A1: %v", err)
	}

	// Session B Turn 1: 1 success (read tool)
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_B_1', 's_iso_B', ?, ?, '{"role":"assistant"}')
	`, now-16000, now-16000)
	if err != nil {
		t.Fatalf("seeding msg B1: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_B_1', 's_iso_B', 'm_B_1', ?, ?, '{"type":"tool","tool":"read","state":{"status":"completed","output":"ok file content"}}')
	`, now-15900, now-15900)
	if err != nil {
		t.Fatalf("seeding part B1: %v", err)
	}

	// Session A Turn 2: 1 success (resets session A streak)
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES ('m_A_2', 's_iso_A', ?, ?, '{"role":"assistant"}')
	`, now-14000, now-14000)
	if err != nil {
		t.Fatalf("seeding msg A2: %v", err)
	}
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES ('p_A_2', 's_iso_A', 'm_A_2', ?, ?, '{"type":"tool","tool":"read","state":{"status":"completed","output":"read succeeded"}}')
	`, now-13900, now-13900)
	if err != nil {
		t.Fatalf("seeding part A2: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx := context.Background()
	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{
			{ID: "s_iso_A", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder"}}},
			{ID: "s_iso_B", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder"}}},
		},
		},
	}

	resp, err := service.HydrateWorkspaceEvents(ctx, "ISO-1", ledger)
	if err != nil {
		t.Fatalf("HydrateWorkspaceEvents failed: %v", err)
	}

	var sessATools, sessBTools []domain.EventFact
	for _, ev := range resp.Events {
		if ev.EventType == domain.EventTypeToolCall {
			if ev.SessionID == "s_iso_A" {
				sessATools = append(sessATools, ev)
			} else if ev.SessionID == "s_iso_B" {
				sessBTools = append(sessBTools, ev)
			}
		}
	}

	if len(sessATools) != 2 {
		t.Fatalf("expected 2 tool events for session A, got %d", len(sessATools))
	}
	if sessATools[0].Status != "error" {
		t.Errorf("expected session A first tool call to be error, got %s", sessATools[0].Status)
	}
	if sessATools[1].Status != "completed" {
		t.Errorf("expected session A second tool call to be completed, got %s", sessATools[1].Status)
	}

	if len(sessBTools) != 1 {
		t.Fatalf("expected 1 tool event for session B, got %d", len(sessBTools))
	}
	if sessBTools[0].Status != "completed" {
		t.Errorf("expected session B tool call to be completed, got %s", sessBTools[0].Status)
	}
}

func TestHydrateWorkspaceEvents_FallbackRoleBehaviorForNamelessSessions(t *testing.T) {
	dbPath := createTestDBFile(t)

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}
	defer rwDB.Close()

	now := time.Now().UnixMilli()

	// 1. Root session with NULL agent and NULL parent_id
	// 2. Root session with whitespace agent ("   ") and NULL parent_id
	// 3. Child session with NULL agent and non-null parent_id
	// 4. Child session with whitespace agent (" \t ") and non-null parent_id
	_, err = rwDB.Exec(`
		INSERT INTO session (id, title, agent, parent_id, model, time_created, time_updated) VALUES
		('s_root_null', 'Root Null Agent', NULL, NULL, 'claude-3-7-sonnet', ?, ?),
		('s_root_ws', 'Root Whitespace Agent', '   ', NULL, 'claude-3-7-sonnet', ?, ?),
		('s_child_null', 'Child Null Agent', NULL, 's_root_null', 'gemini-3.8-flash', ?, ?),
		('s_child_ws', 'Child Whitespace Agent', '   ', 's_root_null', 'gemini-3.8-flash', ?, ?)
	`, now-10000, now-9000, now-8000, now-7000, now-6000, now-5000, now-4000, now-3000)
	if err != nil {
		t.Fatalf("seeding sessions: %v", err)
	}

	// Insert messages and tool parts for each session
	sessions := []string{"s_root_null", "s_root_ws", "s_child_null", "s_child_ws"}
	for i, sid := range sessions {
		msgID := fmt.Sprintf("m_fallback_%d", i)
		partID := fmt.Sprintf("p_fallback_%d", i)
		tMsg := now - int64(1000*(10-i))

		_, err = rwDB.Exec(`
			INSERT INTO message (id, session_id, time_created, time_updated, data)
			VALUES (?, ?, ?, ?, '{"role":"assistant"}')
		`, msgID, sid, tMsg, tMsg)
		if err != nil {
			t.Fatalf("seeding message for %s: %v", sid, err)
		}

		_, err = rwDB.Exec(`
			INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
			VALUES (?, ?, ?, ?, ?, '{"type":"tool","tool":"read","state":{"status":"completed","output":"test content"}}')
		`, partID, sid, msgID, tMsg+10, tMsg+10)
		if err != nil {
			t.Fatalf("seeding part for %s: %v", sid, err)
		}
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 2000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()

	ctx := context.Background()
	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{Sessions: []domain.SessionFact{
			{ID: "s_root_null", Source: domain.SourceOpenCode},
			{ID: "s_root_ws", Source: domain.SourceOpenCode},
			{ID: "s_child_null", Source: domain.SourceOpenCode},
			{ID: "s_child_ws", Source: domain.SourceOpenCode},
		},
		},
	}

	resp, err := service.HydrateWorkspaceEvents(ctx, "NAMELESS-1", ledger)
	if err != nil {
		t.Fatalf("HydrateWorkspaceEvents failed: %v", err)
	}

	sessionRoles := make(map[string]domain.AgentRole)
	for _, ev := range resp.Events {
		sessionRoles[ev.SessionID] = ev.Role
	}

	// Root nameless sessions must resolve to RoleRootAgent ("RootAgent")
	if r, ok := sessionRoles["s_root_null"]; !ok || r != domain.RoleRootAgent {
		t.Errorf("s_root_null: expected %q, got %q (found=%v)", domain.RoleRootAgent, r, ok)
	}
	if r, ok := sessionRoles["s_root_ws"]; !ok || r != domain.RoleRootAgent {
		t.Errorf("s_root_ws: expected %q, got %q (found=%v)", domain.RoleRootAgent, r, ok)
	}

	// Child nameless sessions must resolve to RoleUnassignedSubAgent ("UnassignedSubAgent")
	if r, ok := sessionRoles["s_child_null"]; !ok || r != domain.RoleUnassignedSubAgent {
		t.Errorf("s_child_null: expected %q, got %q (found=%v)", domain.RoleUnassignedSubAgent, r, ok)
	}
	if r, ok := sessionRoles["s_child_ws"]; !ok || r != domain.RoleUnassignedSubAgent {
		t.Errorf("s_child_ws: expected %q, got %q (found=%v)", domain.RoleUnassignedSubAgent, r, ok)
	}
}

func TestLiveTraceSnapshotReuseAndInPlaceInvalidation(t *testing.T) {
	dbPath := createTestDBFile(t)
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	now := time.Now().UnixMilli()
	if _, err := writer.Exec(`INSERT INTO session (id, title, time_created, time_updated) VALUES ('live', 'live', ?, ?);
 INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('turn', 'live', ?, ?, '{"role":"assistant","content":"before","tokens":{"input":10}}');`, now, now, now, now); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	ctx := context.Background()
	if _, err := service.GetSessionTree(ctx); err != nil {
		t.Fatal(err)
	}
	first, err := service.GetSessionTrace(ctx, "live")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.GetSessionTrace(ctx, "live")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Messages) != 1 || len(second.Messages) != 1 {
		t.Fatal("missing source message")
	}
	if &first.Messages[0] != &second.Messages[0] {
		t.Fatal("unchanged source reconstructed trace history")
	}
	if _, err := writer.Exec(`UPDATE message SET data = '{"role":"assistant","content":"after","reasoning":"thinking","tokens":{"input":20}}' WHERE id = 'turn'`); err != nil {
		t.Fatal(err)
	}
	changed, err := service.GetSessionTrace(ctx, "live")
	if err != nil {
		t.Fatal(err)
	}
	if changed.Messages[0].Content != "after" || changed.Messages[0].InputTokens != 20 || changed.Messages[0].ReasoningContent == nil {
		t.Fatalf("in-place update did not invalidate snapshot: %+v", changed.Messages[0])
	}
	if first.Messages[0].Content != "before" {
		t.Fatal("new snapshot mutated earlier accepted data")
	}
	if _, err := writer.Exec(`DROP TABLE message`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetSessionTrace(ctx, "live"); err == nil {
		t.Fatal("failed source read reported cached success")
	}
}

func TestLiveTraceLegacyRevisionRecoveryAndBoundedOnDemandCache(t *testing.T) {
	dbPath := createTestDBFile(t)
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	// Old source schemas lack optional message and part update timestamps.
	if _, err := writer.Exec(`ALTER TABLE message DROP COLUMN time_updated; ALTER TABLE part DROP COLUMN time_updated;
 INSERT INTO session (id, title, time_created, time_updated) VALUES ('one', 'one', 1, 1), ('two', 'two', 1, 1);
 INSERT INTO message (id, session_id, time_created, data) VALUES ('m1','one',1,'{"role":"assistant","content":"one"}'), ('m2','two',1,'{"role":"assistant","content":"two"}');
 INSERT INTO part (id, session_id, message_id, time_created, data) VALUES ('p1','one','m1',1,'{"type":"tool","tool":"read","state":{"status":"completed","output":"before"}}');`); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	ctx := context.Background()
	first, err := service.GetSessionTrace(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetSessionTrace(ctx, "two"); err != nil {
		t.Fatal(err)
	}
	evicted, err := service.GetSessionTrace(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	if &first.Messages[0] == &evicted.Messages[0] {
		t.Fatal("multiple undiscovered lookups were retained")
	}
	if _, err := writer.Exec(`UPDATE part SET data = '{"type":"tool","tool":"read","state":{"status":"error","output":"failed in place"}}' WHERE id = 'p1'`); err != nil {
		t.Fatal(err)
	}
	changed, err := service.GetSessionTrace(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	if changed.FailedToolCalls != 1 || len(changed.RecentErrors) != 1 || changed.Messages[0].Parts[0].ErrorMessage == nil || *changed.Messages[0].Parts[0].ErrorMessage != "failed in place" {
		t.Fatalf("legacy in-place tool update not refreshed: %+v", changed)
	}
	before, err := hook.SourceRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := service.GetSessionTrace(canceled, "one"); err == nil {
		t.Fatal("canceled revision lookup returned cached success")
	}
	after, err := hook.SourceRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.Connection == after.Connection {
		t.Fatal("observer replacement did not change connection identity")
	}
	recovered, err := service.GetSessionTrace(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	if &changed.Messages[0] == &recovered.Messages[0] || recovered.FailedToolCalls != 1 {
		t.Fatal("uncertain source generation reused old cache or lost source facts")
	}
}
