package live_test

import (
	adaptershttp "agenticlens/internal/adapters/http"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
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
		},
		Workspaces: domain.WorkspacesConfig{
			Roots: []domain.WorkspaceRootConfig{{Path: "workspaces", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 8, MaxDirectories: 300}}}, Artifacts: []domain.WorkspaceArtifactConfig{},
		},
		CodeMetrics: domain.CodeMetricsConfig{BaseBranch: "fixture-base", Production: domain.CodeCategoryConfig{Include: []string{}, Exclude: []string{}}, Tests: domain.CodeCategoryConfig{Include: []string{}, Exclude: []string{}}},
		Database: domain.DatabaseConfig{
			BusyTimeoutMs: 5000,
		},
		Live: domain.LiveConfig{
			ActiveThresholdSeconds: 300,
			PollIntervalMs:         2000,
		},
		AnomalyThresholds: domain.AnomalyThresholdsConfig{
			StallThresholdSeconds:       120,
			ErrorRateThreshold:          0.25,
			ErrorRateMinCalls:           5,
			ErrorRateCriticalThreshold:  0.70,
			ContextCeilingWarningTokens: 160000, ContextCeilingCriticalTokens: 200000, ConsecutiveErrorsStreakWarning: 3, ConsecutiveErrorsStreakCritical: 5, ToolThrashingMinFailures: 3, ToolThrashingCriticalFailures: 5, TurnUsageWarningTokens: 30000, TurnUsageCriticalTokens: 60000, ContextGrowthWarningTokens: 40000, ContextGrowthCriticalTokens: 80000,
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

	workspacesDir := t.TempDir()
	for _, name := range []string{"T-1/agentic_metrics.json", "T-1/METRICS.md"} {
		path := filepath.Join(workspacesDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("retained portable history: "+name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := func() map[string][32]byte {
		t.Helper()
		files := map[string][32]byte{}
		if err := filepath.Walk(workspacesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = sha256.Sum256(body)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return files
	}
	beforeSnapshot := snapshot()
	if len(beforeSnapshot) != 2 {
		t.Fatal("owned persistence sentinels missing")
	}
	cfg := testProjectConfig()
	cfg.Workspaces.Roots[0].Path = workspacesDir
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

	if after := snapshot(); !reflect.DeepEqual(beforeSnapshot, after) {
		t.Fatalf("live read changed configured workspace files: before=%v after=%v", beforeSnapshot, after)
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
		INSERT INTO session(id,parent_id,time_updated) VALUES('active-child','s_nulls',1);
	`)
	if err != nil {
		t.Fatalf("inserting null-heavy session: %v", err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	// An active child makes its NULL-heavy direct parent part of the result.
	records, err := hook.FetchActiveSessions(300, 1)
	if err != nil {
		t.Fatalf("FetchActiveSessions failed scanning row with NULL columns: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected parent and active child, got %d records", len(records))
	}
	rec := records[0]
	if rec.ID != "s_nulls" {
		rec = records[1]
	}
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
	if trace.Relationships.State != domain.LiveRelationshipsComplete || len(trace.Diagnostics) != 0 {
		t.Fatalf("linked native relationships not reported complete: %+v", trace)
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
	if trace.Relationships.State != domain.LiveRelationshipsUnavailable || len(trace.Diagnostics) != 1 || trace.Diagnostics[0].Code != domain.DiagnosticChildRelationshipLookupFailed || trace.Diagnostics[0].Message == "" {
		t.Fatalf("optional lookup failure was indistinguishable from an empty child set: %+v", trace)
	}
	payload, err := json.Marshal(trace)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"relationships":{"state":"unavailable"}`) || !strings.Contains(string(payload), `"code":"child_relationship_lookup_failed"`) {
		t.Fatalf("serialized trace lost typed completeness: %s", payload)
	}
	request := httptest.NewRequest("GET", "/api/live/trace/parent", nil)
	request.SetPathValue("id", "parent")
	response := httptest.NewRecorder()
	adaptershttp.LiveTraceHandler(service)(response, request)
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"state":"unavailable"`) || !strings.Contains(response.Body.String(), `"code":"child_relationship_lookup_failed"`) || !strings.Contains(response.Body.String(), `"subagent_type":"coder"`) {
		t.Fatalf("HTTP discarded partial trace/diagnostic status=%d body=%s", response.Code, response.Body.String())
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
	if resp.Events[0].Role != "" {
		t.Errorf("unknown native persona was invented from stored metrics: %q", resp.Events[0].Role)
	}

	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()

	for _, tc := range []struct{ name, stored, sessionAgent, first, second, want string }{
		{"unknown recovered from assistant", "", "", "tester", "tester", "tester"},
		{"legacy root resolved from assistant", "RootAgent", "", "reviewer", "reviewer", "reviewer"},
		{"legacy child resolved from assistant", "UnassignedSubAgent", "", "coder", "", "coder"},
		{"stale named persona recovered", "planner", "", "reviewer", "reviewer", "reviewer"},
		{"explicit raw source wins mapping", "RootAgent", "Build", "reviewer", "coder", "Build"},
		{"conflicting assistants remain unknown", "planner", "", "planner", "coder", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ledger.Facts.Sessions[0].Invocations[0].Role = domain.AgentRole(tc.stored)
			if _, err := writer.Exec(`UPDATE session SET agent=? WHERE id='s_custom'`, tc.sessionAgent); err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Exec(`UPDATE message SET data=? WHERE id='m_custom'`, fmt.Sprintf(`{"role":"assistant","agent":%q,"content":"msg"}`, tc.first)); err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Exec(`INSERT OR REPLACE INTO message(id,session_id,time_created,time_updated,data) VALUES('m_custom_second','s_custom',?,?,?)`, now+1, now+1, fmt.Sprintf(`{"role":"assistant","agent":%q,"content":"msg2"}`, tc.second)); err != nil {
				t.Fatal(err)
			}

			before, err := json.Marshal(ledger)
			if err != nil {
				t.Fatal(err)
			}
			result, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CUSTOM", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsHydrated || result.Availability != domain.TraceComplete || len(result.Events) != 2 {
				t.Fatalf("hydration lost or duplicated occurrences: %+v", result)
			}
			matched, appended := 0, 0
			for _, event := range result.Events {
				if event.EventID == "opencode:message:m_custom" {
					matched++
				} else if event.EventID == "opencode:message:m_custom_second" {
					appended++
				} else {
					t.Errorf("unexpected source identity: %+v", event)
				}
				if string(event.Role) != tc.want || event.PromptContent == nil {
					t.Errorf("hydrated source role/payload=%+v want role=%q", event, tc.want)
				}
			}
			if matched != 1 || appended != 1 {
				t.Fatalf("occurrence ownership changed: %+v", result.Events)
			}
			after, err := json.Marshal(ledger)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("read projection mutated stored facts")
			}
			var unavailable *live.LiveService
			fallback, err := unavailable.HydrateWorkspaceEvents(context.Background(), "AL-CUSTOM", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if fallback.IsHydrated || len(fallback.Events) != 0 || fallback.Availability != domain.TraceUnavailable {
				t.Fatalf("unavailable source rewrote baseline: %+v", fallback)
			}
		})
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

	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Availability != domain.TracePartial || !resp.IsHydrated || resp.DataSource != domain.DataSourceLocalCodex || len(resp.Events) != 1 || len(resp.Coverage.Diagnostics) == 0 {
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
	if !complete.IsHydrated || complete.DataSource != domain.DataSourceLocalCodex || len(complete.Events) != 1 || complete.Availability != domain.TraceComplete {
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

	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CHILD-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Availability != domain.TracePartial || !resp.IsHydrated || resp.DataSource != domain.DataSourceLocalCodex || len(resp.Events) != 1 || len(resp.Coverage.Diagnostics) == 0 {
		t.Fatalf("partial Codex child tree was reported as hydrated: %#v", resp)
	}
	write("sessions/2026/10/26/rollout-child-two.jsonl", "child-two", "root-thread", false)
	complete, err := service.HydrateWorkspaceEvents(context.Background(), "AL-CHILD-PARTIAL", ledger)
	if err != nil {
		t.Fatal(err)
	}
	if !complete.IsHydrated || complete.DataSource != domain.DataSourceLocalCodex || len(complete.Events) != 1 || complete.Availability != domain.TraceComplete {
		t.Fatalf("complete Codex child tree failed to hydrate after restoring the missing rollout: %#v", complete)
	}
}

func TestHydrateWorkspaceEvents_SourceLessLedgerStaysUnavailable(t *testing.T) {
	service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), t.TempDir())
	defer service.Close()
	facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "legacy", Skills: []domain.SkillFact{{Name: "stored"}}, Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}}}}}
	resp, err := service.HydrateWorkspaceEvents(context.Background(), "AL-LEGACY", domain.AgenticMetricsLedger{Facts: facts})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Availability != domain.TraceUnavailable || resp.IsHydrated || resp.DataSource != domain.DataSourceUnavailable || len(resp.Events) != 0 || len(resp.Anomalies) != 1 || len(resp.Coverage.Diagnostics) == 0 {
		t.Fatalf("unknown source lost summaries or invented trace: %+v", resp)
	}
}

func TestHydrationAvailabilityPreservesLeanFactsAndLedgerBytes(t *testing.T) {
	modes := []string{"nil service", "nil hook", "missing session", "closed database", "missing tables", "missing persona column", "empty SQLite session", "partial SQLite session", "missing Codex home", "missing Codex rollout", "empty Codex rollout", "partial Codex rollout"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			cfg := testProjectConfig()
			cfg.Workspaces.Roots[0].Path = root
			source := domain.SourceOpenCode
			if strings.Contains(mode, "Codex") {
				source = domain.SourceCodex
			}
			facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "recorded-session", Source: source, User: "tester", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}, Skills: []domain.SkillFact{{Name: "stored"}}, Delegations: []domain.DelegationFact{{ChildSession: "unavailable-child"}}, Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}}}}}
			ledger := domain.AgenticMetricsLedger{Facts: facts}
			before, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			ledgerPath := filepath.Join(root, "agentic_metrics.json")
			if err = os.WriteFile(ledgerPath, before, 0600); err != nil {
				t.Fatal(err)
			}
			var service *live.LiveService
			switch mode {
			case "nil service":
			case "nil hook":
				service = live.NewLiveService(nil, cfg)
			case "missing Codex home", "missing Codex rollout", "empty Codex rollout", "partial Codex rollout":
				home := filepath.Join(t.TempDir(), "codex")
				if mode != "missing Codex home" {
					if err = os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "empty Codex rollout" || mode == "partial Codex rollout" {
					content := `{"type":"session_meta","payload":{"id":"recorded-session","session_id":"recorded-session"}}` + "\n"
					if mode == "partial Codex rollout" {
						content += `{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"good","name":"load_skill","arguments":"{\"name\":\"fresh\"}"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","name":"load_skill","call_id":42}}` + "\n"
					}
					if err = os.WriteFile(filepath.Join(home, "sessions", "recorded.jsonl"), []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				service = live.NewLiveServiceWithCodexHome(nil, cfg, home)
			default:
				dbPath := createTestDBFile(t)
				writer, err := sql.Open("sqlite", dbPath)
				if err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "empty SQLite session", "partial SQLite session":
					if _, err = writer.Exec(`INSERT INTO session(id,parent_id,agent,model,title,time_created,time_updated) VALUES('recorded-session',NULL,'coder','fixture','TEST-100',1,1)`); err != nil {
						t.Fatal(err)
					}
					if mode == "partial SQLite session" {
						if _, err = writer.Exec(`INSERT INTO message VALUES('good','recorded-session',1000,1000,'{"role":"assistant","content":"source-only-body"}');INSERT INTO message VALUES('bad','recorded-session',2000,2000,'{broken')`); err != nil {
							t.Fatal(err)
						}
					}
				case "missing tables":
					if _, err = writer.Exec(`DROP TABLE message;DROP TABLE part;`); err != nil {
						t.Fatal(err)
					}
				case "missing persona column":
					if _, err = writer.Exec(`INSERT INTO session(id,agent,time_created,time_updated) VALUES('recorded-session','planner',1,1);ALTER TABLE session DROP COLUMN agent;`); err != nil {
						t.Fatal(err)
					}
				}
				if err = writer.Close(); err != nil {
					t.Fatal(err)
				}
				hook, err := live.OpenReadOnlyHook(dbPath, 5000)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "closed database" {
					if err = hook.Close(); err != nil {
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
			resp, err := service.HydrateWorkspaceEvents(context.Background(), "TEST-100", ledger)
			if err != nil {
				t.Fatal(err)
			}
			want := domain.TraceUnavailable
			if strings.HasPrefix(mode, "empty ") {
				want = domain.TraceComplete
			}
			if strings.HasPrefix(mode, "partial ") {
				want = domain.TracePartial
			}
			if resp.Availability != want || resp.IsHydrated != (want != domain.TraceUnavailable) || !reflect.DeepEqual(resp.Coverage.RequiredSessionIDs, []string{"recorded-session"}) {
				t.Fatalf("availability/owner coverage mismatch: %+v want %s", resp, want)
			}
			if err = resp.Validate(); err != nil {
				t.Fatal(err)
			}
			if want == domain.TraceComplete {
				if len(resp.Events) != 0 || !reflect.DeepEqual(resp.Coverage.CompleteSessionIDs, []string{"recorded-session"}) || len(resp.Coverage.Diagnostics) != 0 {
					t.Fatalf("successful empty read not complete: %+v", resp)
				}
			} else {
				if len(resp.Coverage.CompleteSessionIDs) != 0 || len(resp.Coverage.Diagnostics) == 0 {
					t.Fatalf("failed/partial source certified owner: %+v", resp)
				}
				wantAnomalies := facts.AnomalyFacts()
				for i := range wantAnomalies {
					wantAnomalies[i].Evidence = []domain.AnomalyEvidenceRef{}
				}
				if !reflect.DeepEqual(resp.Anomalies, wantAnomalies) {
					t.Fatalf("incomplete source erased stored summary: %+v", resp)
				}
			}
			if mode == "partial SQLite session" && (len(resp.Events) != 1 || resp.Events[0].PromptContent == nil || *resp.Events[0].PromptContent != "source-only-body") {
				t.Fatalf("supported partial source payload lost: %+v", resp.Events)
			}
			if want == domain.TracePartial && len(resp.Events) == 0 {
				t.Fatalf("supported partial source rows discarded: %+v", resp)
			}
			if want == domain.TraceUnavailable && (len(resp.Events) != 0 || resp.DataSource != domain.DataSourceUnavailable) {
				t.Fatalf("invented unavailable events: %+v", resp)
			}
			after, err := os.ReadFile(ledgerPath)
			if err != nil {
				t.Fatal(err)
			}
			unchanged, err := json.Marshal(facts)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) || string(before) != string(unchanged) {
				t.Fatal("hydration mutated input or ledger bytes")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("hydration wrote workspace sidecars: %+v", entries)
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
	if resp.DataSource != domain.DataSourceUnavailable {
		t.Errorf("expected DataSource == %q, got %q", domain.DataSourceUnavailable, resp.DataSource)
	}
	if len(resp.Events) != 0 {
		t.Fatalf("missing local session discarded recorded facts: %#v", resp.Events)
	}

	for _, legacy := range []domain.AgentRole{domain.AgentRole("RootAgent"), domain.AgentRole("UnassignedSubAgent")} {
		t.Run(string(legacy), func(t *testing.T) {
			for i := range ledger.Facts.Sessions {
				for j := range ledger.Facts.Sessions[i].Invocations {
					ledger.Facts.Sessions[i].Invocations[j].Role = legacy
				}
			}
			fallback, err := service.HydrateWorkspaceEvents(ctx, "LEGACY-UNKNOWN", ledger)
			if err != nil {
				t.Fatal(err)
			}
			if fallback.IsHydrated || len(fallback.Events) != 0 {
				t.Fatalf("unavailable source rewrote legacy persona: %+v", fallback)
			}
		})
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

	// Unknown root and child personas remain blank regardless of ancestry.
	if r, ok := sessionRoles["s_root_null"]; !ok || r != domain.AgentRole("") {
		t.Errorf("s_root_null: expected %q, got %q (found=%v)", domain.AgentRole(""), r, ok)
	}
	if r, ok := sessionRoles["s_root_ws"]; !ok || r != domain.AgentRole("") {
		t.Errorf("s_root_ws: expected %q, got %q (found=%v)", domain.AgentRole(""), r, ok)
	}

	// Child ancestry does not manufacture a persona.
	if r, ok := sessionRoles["s_child_null"]; !ok || r != domain.AgentRole("") {
		t.Errorf("s_child_null: expected %q, got %q (found=%v)", domain.AgentRole(""), r, ok)
	}
	if r, ok := sessionRoles["s_child_ws"]; !ok || r != domain.AgentRole("") {
		t.Errorf("s_child_ws: expected %q, got %q (found=%v)", domain.AgentRole(""), r, ok)
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
 INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('turn', 'live', ?, ?, '{"role":"assistant","content":"before","agent":"planner","tokens":{"input":10}}');`, now, now, now, now); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	ctx := context.Background()
	tree, err := service.GetSessionTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.RootSessions) != 1 || tree.RootSessions[0].Role != "planner" {
		t.Fatalf("assistant persona missing from tree: %+v", tree)
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
	if _, err := writer.Exec(`UPDATE message SET data = '{"role":"assistant","content":"after","agent":"coder","reasoning":"thinking","tokens":{"input":20}}' WHERE id = 'turn'`); err != nil {
		t.Fatal(err)
	}
	changed, err := service.GetSessionTrace(ctx, "live")
	if err != nil {
		t.Fatal(err)
	}
	if changed.Messages[0].Content != "after" || changed.Messages[0].InputTokens != 20 || changed.Messages[0].ReasoningContent == nil {
		t.Fatalf("in-place update did not invalidate snapshot: %+v", changed.Messages[0])
	}
	updatedTree, err := service.GetSessionTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(updatedTree.RootSessions) != 1 || updatedTree.RootSessions[0].Role != "coder" {
		t.Fatalf("persona cache failed invalidation: %+v", updatedTree)
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

func TestLiveModelResolutionPreservesUnknownAndObservedNames(t *testing.T) {
	cases := []struct {
		id, raw, want string
		null          bool
	}{
		{id: "absent", want: "", null: true},
		{id: "empty", raw: "", want: ""},
		{id: "blank", raw: " \t\n ", want: ""},
		{id: "observed-default", raw: "default", want: "default"},
		{id: "observed-unknown", raw: "unknown", want: "unknown"},
		{id: "observed-known", raw: "source-model", want: "source-model"},
		{id: "observed-trimmed", raw: " source-model ", want: "source-model"},
		{id: "object-default", raw: `{"id":"default"}`, want: "default"},
		{id: "object-known", raw: `{"id":"source-model"}`, want: "source-model"},
		// A12 preserves existing nonempty wrapper interpretation rather than
		// broadening accepted model metadata or normalizing stored names.
		{id: "object-empty-id", raw: `{"id":" "}`, want: `{"id":" "}`},
	}
	path := createTestDBFile(t)
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UnixMilli()
	for _, tc := range cases {
		if _, err := writer.Exec(`INSERT INTO session(id,agent,model,title,time_created,time_updated) VALUES(?, 'builder', ?, 'Model identity', ?, ?)`, tc.id, sql.NullString{String: tc.raw, Valid: !tc.null}, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	defer service.Close()
	response := httptest.NewRecorder()
	adaptershttp.LiveSessionsHandler(service)(response, httptest.NewRequest("GET", "/api/live/sessions", nil))
	if response.Code != 200 {
		t.Fatalf("live models HTTP %d: %s", response.Code, response.Body.String())
	}
	var tree domain.LiveSessionTree
	if err := json.Unmarshal(response.Body.Bytes(), &tree); err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Roots []struct {
			ID    string  `json:"id"`
			Model *string `json:"model"`
		} `json:"root_sessions"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	for _, node := range wire.Roots {
		if node.Model == nil {
			t.Fatalf("model string became absent/null for %s: %s", node.ID, response.Body.String())
		}
	}
	if tree.TotalActiveSessions != len(cases) || len(tree.RootSessions) != len(cases) {
		t.Fatalf("missing model changed session count: %+v", tree)
	}
	nodes := make(map[string]*domain.LiveSessionNode, len(tree.RootSessions))
	for _, node := range tree.RootSessions {
		nodes[node.ID] = node
	}
	for _, tc := range cases {
		if node := nodes[tc.id]; node == nil || node.Model != tc.want {
			t.Errorf("source %s model changed: node=%+v want=%q", tc.id, node, tc.want)
		}
	}
	if after, err := os.ReadFile(path); err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("live read changed source database bytes: %v", err)
	}
}

func TestRecordedModelNamesAndCountersSurviveRepeatedReads(t *testing.T) {
	root := t.TempDir()
	shared, err := os.ReadFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	workspace := filepath.Join(root, "records")
	ticket := filepath.Join(workspace, "T-02")
	if err := os.MkdirAll(ticket, 0700); err != nil {
		t.Fatal(err)
	}
	ledger := []byte(`{"schema_version":6,"ticket_id":"T-02","ticket_title":"Recorded models","ticket_type":"feat","started_at":"2026-10-17T10:00:00Z","completed_at":"2026-10-17T11:00:00Z","updated_at":"2026-10-17T11:00:00Z","code":{"lines_added":0,"lines_deleted":0,"test_lines_added":0,"test_lines_deleted":0,"prod_lines_added":0,"prod_lines_deleted":0,"files_changed":0,"sdlc_iterations":0},"sessions":[{"id":"recorded-owner","source":"opencode","user":"fixture","invocations":[{"role":"builder","model":"default","started_at":null,"completed_at":null,"active_duration_seconds":7,"turns":2,"tokens":{"input":17,"output":3,"cache_read":0,"reasoning":0,"peak_context":17},"cost_usd":0.1,"cost_evidence":{"state":"observed","known_subtotal_usd":0.1},"compactions":1,"user_iterations":2},{"role":"builder","model":"unknown","started_at":null,"completed_at":null,"active_duration_seconds":9,"turns":3,"tokens":{"input":19,"output":5,"cache_read":0,"reasoning":0,"peak_context":19},"cost_usd":0.2,"cost_evidence":{"state":"observed","known_subtotal_usd":0.2},"compactions":2,"user_iterations":3}],"tools":[],"anomalies":[]}]}`)
	path := filepath.Join(ticket, "agentic_metrics.json")
	if err := os.WriteFile(path, ledger, 0600); err != nil {
		t.Fatal(err)
	}
	get := func() adaptershttp.WorkspaceDetailResponse {
		t.Helper()
		request := httptest.NewRequest("GET", "/api/workspaces/T-02", nil)
		request.SetPathValue("ticketId", "T-02")
		response := httptest.NewRecorder()
		adaptershttp.WorkspaceDetailHandler(workspace)(response, request)
		if response.Code != 200 {
			t.Fatalf("recorded models HTTP %d: %s", response.Code, response.Body.String())
		}
		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil {
			t.Fatal(err)
		}
		return detail
	}
	for i := 0; i < 2; i++ {
		detail := get()
		owners := detail.Ledger.Facts.Sessions
		if len(owners) != 1 || len(owners[0].Invocations) != 2 {
			t.Fatalf("recorded rows changed: %+v", owners)
		}
		rows := owners[0].Invocations
		if rows[0].Model != "default" || rows[1].Model != "unknown" || rows[0].ActiveDurationSeconds != 7 || rows[1].ActiveDurationSeconds != 9 || rows[0].Turns != 2 || rows[1].Turns != 3 || rows[0].Compactions != 1 || rows[1].Compactions != 2 || rows[0].UserIterations != 2 || rows[1].UserIterations != 3 {
			t.Fatalf("recorded model identity/counters changed: %+v", rows)
		}
		if rows[0].Tokens.Input != 17 || rows[0].Tokens.Output != 3 || rows[1].Tokens.Input != 19 || rows[1].Tokens.Output != 5 || rows[0].CostUSD == nil || *rows[0].CostUSD != 0.1 || rows[1].CostUSD == nil || *rows[1].CostUSD != 0.2 {
			t.Fatalf("recorded tokens/cost changed: %+v", rows)
		}
		modelCounts := map[string]int64{}
		for _, model := range detail.Ledger.Models {
			modelCounts[model.Model] = model.Invocations
		}
		if !reflect.DeepEqual(modelCounts, map[string]int64{"default": 1, "unknown": 1}) {
			t.Fatalf("recorded model groups changed: %+v", detail.Ledger.Models)
		}
		if after, err := os.ReadFile(path); err != nil || !reflect.DeepEqual(after, ledger) {
			t.Fatalf("recorded read rewrote original ledger bytes: %v", err)
		}
	}
}
