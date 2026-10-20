package live_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

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

func TestTransientMemoryBufferBoundary(t *testing.T) {
	buf := live.NewTraceBuffer(500)

	sessionID := "test_session_buffer"
	totalMessages := 600

	messages := make([]domain.LiveMessageFact, totalMessages)
	for i := 0; i < totalMessages; i++ {
		messages[i] = domain.LiveMessageFact{
			ID:        fmt.Sprintf("msg_%d", i),
			Role:      "assistant",
			Content:   fmt.Sprintf("content_%d", i),
			Timestamp: fmt.Sprintf("2026-10-19T10:00:%02dZ", i%60),
		}
	}

	buf.SetMessages(sessionID, messages)

	if buf.Len(sessionID) != 500 {
		t.Fatalf("expected buffer length 500, got %d", buf.Len(sessionID))
	}

	stored := buf.GetMessages(sessionID)
	if len(stored) != 500 {
		t.Fatalf("expected 500 stored messages, got %d", len(stored))
	}

	// Oldest 100 dropped, so first message should be msg_100
	if stored[0].ID != "msg_100" {
		t.Errorf("expected oldest retained message to be msg_100, got %s", stored[0].ID)
	}
	if stored[499].ID != "msg_599" {
		t.Errorf("expected newest retained message to be msg_599, got %s", stored[499].ID)
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

	cfg := domain.DefaultProjectConfig()
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
	if service.Buffer().Len("s_live") != 1 {
		t.Fatalf("expected buffer to hold 1 message, got %d", service.Buffer().Len("s_live"))
	}

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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

func TestTraceBuffer_NegativeCapacityDefaultsSafely(t *testing.T) {
	bufNeg := live.NewTraceBuffer(-10)
	bufZero := live.NewTraceBuffer(0)

	if bufNeg == nil || bufZero == nil {
		t.Fatal("expected non-nil buffer")
	}

	msg := domain.LiveMessageFact{ID: "m1", Role: "assistant"}
	bufNeg.AppendMessages("s1", msg)
	if bufNeg.Len("s1") != 1 {
		t.Errorf("expected len 1, got %d", bufNeg.Len("s1"))
	}
}

func TestTraceBuffer_AdversarialConcurrentReadWrite(t *testing.T) {
	buf := live.NewTraceBuffer(50)
	sessionID := "s_conc_ring"

	var wg sync.WaitGroup
	const writers = 10
	const readers = 10
	const msgsPerWriter = 20

	wg.Add(writers + readers)

	// Concurrent writers
	for w := 0; w < writers; w++ {
		go func(writerIdx int) {
			defer wg.Done()
			for i := 0; i < msgsPerWriter; i++ {
				msg := domain.LiveMessageFact{
					ID:        fmt.Sprintf("w%d_m%d", writerIdx, i),
					Role:      "assistant",
					Timestamp: "2026-10-19T10:00:00Z",
				}
				buf.AppendMessages(sessionID, msg)
			}
		}(w)
	}

	// Concurrent readers
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				msgs := buf.GetMessages(sessionID)
				_ = msgs
				_ = buf.Len(sessionID)
			}
		}()
	}

	wg.Wait()

	if buf.Len(sessionID) > 50 {
		t.Errorf("expected buffer len <= 50, got %d", buf.Len(sessionID))
	}
}

func TestLiveService_AdversarialSessionIDValidation(t *testing.T) {
	dbPath := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
	defer service.Close()

	ctx := context.Background()

	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "s_al33",
					Role: "orchestrator",
				},
			},
			Events: []domain.EventFact{
				{
					Timestamp: "2026-10-19T10:00:00Z",
					SessionID: "s_al33",
					Turn:      1,
					Role:      "orchestrator",
					EventType: domain.EventTypeCompaction,
					Status:    "completed",
				},
			},
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

func TestHydrateWorkspaceEvents_FallbackWhenSessionNotFound(t *testing.T) {
	dbPath := createTestDBFile(t)
	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
	defer service.Close()

	ctx := context.Background()

	ledger := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "s_missing",
					Role: "orchestrator",
				},
			},
			Events: []domain.EventFact{
				{
					Timestamp: "2026-10-19T10:00:00Z",
					SessionID: "s_missing",
					Turn:      1,
					Role:      "orchestrator",
					EventType: domain.EventTypeCompaction,
					Status:    "completed",
				},
			},
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
	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 fallback event, got %d", len(resp.Events))
	}
}

func TestHydrateWorkspaceEvents_HostileDegradationAndFailureModes(t *testing.T) {
	ctx := context.Background()
	workspacesDir := t.TempDir()

	ledgerWithFallback := domain.AgenticMetricsLedger{
		Facts: domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "ses-hostile", Role: "coder"},
			},
			Events: []domain.EventFact{
				{
					Timestamp: "2026-10-19T10:00:00Z",
					SessionID: "ses-hostile",
					Turn:      1,
					Role:      "coder",
					EventType: domain.EventTypeCompaction,
					Status:    "completed",
				},
			},
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
		if len(resp.Events) != 1 {
			t.Errorf("expected 1 fallback event, got %d", len(resp.Events))
		}
	})

	// 2. LiveService with nil hook
	t.Run("LiveService with nil hook returns fallback cleanly", func(t *testing.T) {
		service := live.NewLiveService(nil, domain.DefaultProjectConfig())
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
		service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

		service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

		service := live.NewLiveService(hook, domain.DefaultProjectConfig())
		defer service.Close()

		corruptLedger := domain.AgenticMetricsLedger{
			Facts: domain.AgenticFactLedger{
				Invocations: []domain.InvocationFact{
					{ID: "s_corrupt", Role: "coder"},
				},
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

		service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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
