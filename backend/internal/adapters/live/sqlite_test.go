package live_test

import (
	"context"
	"database/sql"
	"encoding/json"
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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
	ruleNames := make(map[string]bool)
	for _, a := range trace.Anomalies {
		ruleNames[a.RuleName] = true
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

	cfg := domain.DefaultProjectConfig()
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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

	service := live.NewLiveService(hook, domain.DefaultProjectConfig())
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
