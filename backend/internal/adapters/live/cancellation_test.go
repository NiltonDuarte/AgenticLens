package live_test

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
	"modernc.org/sqlite"
)

var traceGateID atomic.Uint64

type observedTraceWait struct {
	context.Context
	reached chan struct{}
	once    sync.Once
}

func (ctx *observedTraceWait) Done() <-chan struct{} {
	ctx.once.Do(func() { close(ctx.reached) })
	return ctx.Context.Done()
}

type codexPrefixCancelContext struct {
	context.Context
	cancel       context.CancelFunc
	cancelAt     int
	parserChecks int
}

func (ctx *codexPrefixCancelContext) Err() error {
	if pc, _, _, ok := runtime.Caller(1); ok && strings.HasSuffix(runtime.FuncForPC(pc).Name(), "/codex.ParseRolloutAtContext") {
		ctx.parserChecks++
		if ctx.parserChecks == ctx.cancelAt {
			ctx.cancel()
		}
	}
	return ctx.Context.Err()
}

func TestLiveTraceCancellationInterruptsSQLiteAndPreservesCache(t *testing.T) {
	t.Run("first step", func(t *testing.T) { assertLiveTraceCancellation(t, false) })
	t.Run("later indexed step", func(t *testing.T) { assertLiveTraceCancellation(t, true) })
}

func assertLiveTraceCancellation(t *testing.T, laterStep bool) {
	t.Helper()
	path := createTestDBFile(t)
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	if _, err = writer.Exec(`
		INSERT INTO session(id, title, time_created, time_updated)
		  VALUES ('cached','cached',1,1), ('slow','slow',1,1);
		INSERT INTO message(id, session_id, time_created, time_updated, data)
		  VALUES ('cached-message','cached',1,1,'{"role":"assistant","content":"initial cache"}'),
		         ('slow-first','slow',1,1,'{"role":"assistant","content":"first row"}'),
	         ('slow-second','slow',2,2,'{"role":"assistant","content":"second row"}');
	`); err != nil {
		t.Fatal(err)
	}
	blockedRow := "slow-first"
	if laterStep {
		blockedRow = "slow-second"
		if _, err := writer.Exec("CREATE INDEX fixture_message_order ON message(session_id, time_created)"); err != nil {
			t.Fatal(err)
		}
	}

	started, release := make(chan struct{}), make(chan struct{})
	var entered, released sync.Once
	var halt, slowReadCompleted atomic.Bool
	function := fmt.Sprintf("fixture_trace_gate_%d", traceGateID.Add(1))
	if err := sqlite.RegisterScalarFunction(function, 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		n := args[0].(int64)
		switch n {
		case 0:
			entered.Do(func() { close(started) })
			<-release
		case -1:
			slowReadCompleted.Store(true)
		}
		return !halt.Load(), nil
	}); err != nil {
		t.Fatal(err)
	}

	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	t.Cleanup(func() { service.Close() })
	t.Cleanup(func() {
		halt.Store(true)
		released.Do(func() { close(release) })
	})
	if err := service.CheckAvailability(); err != nil {
		t.Fatal(err)
	}
	alreadyCancelled, cancelBeforeStart := context.WithCancel(context.Background())
	cancelBeforeStart()
	if _, err := service.GetSessionTree(alreadyCancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("already cancelled session listing returned %v", err)
	}
	// The default case gates the initial execution. In the indexed case,
	// SQLite can return the first row without evaluating the later row.
	// The view makes the selected native step enter a bounded VM loop.
	if _, err := writer.Exec(fmt.Sprintf(`
		ALTER TABLE message RENAME TO fixture_message;
		CREATE VIEW message AS
		  SELECT id, session_id, time_created, time_updated,
		    CASE WHEN id = '%s' THEN
		      (WITH RECURSIVE scan(n) AS (
		        VALUES (0)
		        UNION ALL SELECT n + 1 FROM scan WHERE n < 10000 AND %s(n)
		      ) SELECT CASE WHEN %s(-1) THEN data ELSE NULL END FROM scan ORDER BY n DESC LIMIT 1)
		    ELSE data END AS data
		  FROM fixture_message;
	`, blockedRow, function, function)); err != nil {
		t.Fatal(err)
	}
	if laterStep {
		assertIndexedTracePlan(t, hook.DB())
	}
	first, err := service.GetSessionTrace(context.Background(), "cached")
	if err != nil || len(first.Messages) != 1 {
		t.Fatalf("preparing accepted trace: %+v, %v", first, err)
	}
	guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	readCtx, cancelRead := context.WithCancel(context.Background())
	defer cancelRead()
	readDone := make(chan error, 1)
	go func() {
		_, err := service.GetSessionTrace(readCtx, "slow")
		readDone <- err
	}()
	select {
	case <-started:
	case err := <-readDone:
		t.Fatalf("trace returned before reaching the SQLite read: %v", err)
	case <-guard.Done():
		t.Fatal("trace did not enter the SQLite read")
	}

	waitCtx, stopWaiting := context.WithCancel(guard)
	defer stopWaiting()
	wait := &observedTraceWait{Context: waitCtx, reached: make(chan struct{})}
	waitDone := make(chan error, 1)
	go func() {
		_, err := service.GetSessionTrace(wait, "cached")
		waitDone <- err
	}()
	select {
	case <-wait.reached:
		stopWaiting()
	case err := <-waitDone:
		t.Fatalf("trace waiter returned while the slow native query was still held: %v", err)
	case <-guard.Done():
		t.Fatal("trace waiter did not register for cancellation")
	}
	select {
	case err := <-waitDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled cache waiter returned %v", err)
		}
	case <-guard.Done():
		t.Fatal("trace cache waiter could not cancel during another read")
	}
	cancelRead()
	released.Do(func() { close(release) })
	select {
	case err := <-readDone:
		if !errors.Is(err, context.Canceled) || slowReadCompleted.Load() {
			t.Fatalf("SQLite work completed despite cancellation: %v", err)
		}
	case <-guard.Done():
		t.Fatal("cancelled request kept running native SQLite work")
	}
	after, err := service.GetSessionTrace(guard, "cached")
	if err != nil || len(after.Messages) != 1 || &after.Messages[0] != &first.Messages[0] {
		t.Fatalf("independent lookup lost its accepted cached snapshot: %+v, %v", after, err)
	}
	if _, err := writer.Exec(`UPDATE fixture_message SET data = '{"role":"assistant","content":"revised cache"}' WHERE id='cached-message'`); err != nil {
		t.Fatal(err)
	}
	revised, err := service.GetSessionTrace(guard, "cached")
	if err != nil || len(revised.Messages) != 1 || revised.Messages[0].Content != "revised cache" {
		t.Fatalf("independent read hid a genuine SQLite revision: %+v, %v", revised, err)
	}
}

func assertIndexedTracePlan(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`EXPLAIN QUERY PLAN
		SELECT id, session_id, time_created, time_updated, data
		FROM message WHERE session_id=? ORDER BY time_created`, "slow")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var usesIndex bool
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		usesIndex = usesIndex || strings.Contains(detail, "USING INDEX fixture_message_order")
		if parent == 0 && strings.Contains(detail, "TEMP B-TREE FOR ORDER BY") {
			t.Fatalf("fixture plan buffers the outer trace rows: %s", detail)
		}
	}
	if err := rows.Err(); err != nil || !usesIndex {
		t.Fatalf("fixture did not select the indexed trace plan: %v", err)
	}
}

func TestHistoricalEventsCancellationKeepsEarlierCodexOwner(t *testing.T) {
	path := createTestDBFile(t)
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	if _, err = writer.Exec(`
		INSERT INTO session(id,agent,title,time_created,time_updated)
		  VALUES ('sqlite-owner','tester','Fixture',1000,2000);
		INSERT INTO message(id,session_id,time_created,time_updated,data)
		  VALUES ('sqlite-message','sqlite-owner',1000,1000,'{"role":"assistant","modelID":"m","content":"native SQLite row"}');
	`); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	rolloutPath := filepath.Join(home, "sessions", "rollout.jsonl")
	if err := os.MkdirAll(filepath.Dir(rolloutPath), 0700); err != nil {
		t.Fatal(err)
	}
	rollout := `{"type":"session_meta","payload":{"id":"codex-owner","source":"cli"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"kept-call","name":"read","arguments":"{}"}}
`
	if err := os.WriteFile(rolloutPath, []byte(rollout), 0600); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Workspaces.Roots[0].Path = "."
	shared, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	workspace := filepath.Join(root, "fixture")
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(workspace, "agentic_metrics.json")
	version, ticket, title, ticketType := domain.CurrentSchemaVersion, "FIX-63", "Mixed source fixture", domain.TicketTypeFix
	before, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, TicketType: &ticketType,
		Sessions: []domain.SessionFact{
			{ID: "codex-owner", Source: domain.SourceCodex, User: "tester"},
			{ID: "sqlite-owner", Source: domain.SourceOpenCode, User: "tester"},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledger, before, 0600); err != nil {
		t.Fatal(err)
	}

	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(hook, testProjectConfig(), home)
	t.Cleanup(func() { service.Close() })
	db := hook.DB()
	db.SetMaxOpenConns(1)
	guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	held, err := db.Conn(guard)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	waitCount := db.Stats().WaitCount
	ctx, cancel := context.WithCancel(guard)
	defer cancel()
	handler := adaptershttp.WorkspaceEventsHandler(root, service)
	read := func(ctx context.Context) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/FIX-63/events", nil).WithContext(ctx)
		req.SetPathValue("ticketId", "FIX-63")
		rec := httptest.NewRecorder()
		handler(rec, req)
		return rec
	}
	requestDone := make(chan *httptest.ResponseRecorder, 1)
	go func() { requestDone <- read(ctx) }()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for db.Stats().WaitCount == waitCount {
		select {
		case rec := <-requestDone:
			t.Fatalf("GET returned before attempting the SQLite owner: %d %s", rec.Code, rec.Body)
		case <-guard.Done():
			t.Fatal("GET never requested the occupied SQLite connection")
		case <-ticker.C:
		}
	}
	cancel()

	var partial domain.WorkspaceEventsResponse
	select {
	case rec := <-requestDone:
		if err := json.Unmarshal(rec.Body.Bytes(), &partial); rec.Code != http.StatusOK || err != nil {
			t.Fatalf("GET did not return a usable partial response: status %d, body %s, error %v", rec.Code, rec.Body, err)
		}
	case <-guard.Done():
		t.Fatal("cancelled historical GET waited for the occupied source connection")
	}
	if partial.Availability != domain.TracePartial || partial.DataSource != domain.DataSourceLocalCodex ||
		!slices.Equal(partial.Coverage.CompleteSessionIDs, []string{"codex-owner"}) ||
		!slices.Equal(partial.Coverage.RequiredSessionIDs, []string{"codex-owner", "sqlite-owner"}) ||
		!strings.Contains(strings.Join(partial.Coverage.Diagnostics, " "), context.Canceled.Error()) ||
		len(partial.Events) != 1 || partial.Events[0].CallID != "kept-call" {
		t.Fatalf("interruption changed earlier evidence or certified the SQLite owner: %+v", partial)
	}
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	rec := read(ctx)
	var unavailable domain.WorkspaceEventsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &unavailable); rec.Code != http.StatusOK || err != nil ||
		unavailable.Availability != domain.TraceUnavailable || len(unavailable.Events) != 0 ||
		!slices.Equal(unavailable.Coverage.RequiredSessionIDs, []string{"codex-owner", "sqlite-owner"}) ||
		len(unavailable.Coverage.CompleteSessionIDs) != 0 ||
		!strings.Contains(strings.Join(unavailable.Coverage.Diagnostics, " "), context.Canceled.Error()) {
		t.Fatalf("already cancelled historical read hid coverage or accepted evidence: status %d, body %s, error %v", rec.Code, rec.Body, err)
	}
	rec = read(guard)
	var complete domain.WorkspaceEventsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &complete); rec.Code != http.StatusOK || err != nil || complete.Availability != domain.TraceComplete || len(complete.Events) != 2 {
		t.Fatalf("independent historical read did not recover: status %d, body %s, error %v", rec.Code, rec.Body, err)
	}
	after, err := os.ReadFile(ledger)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("historical GET changed persisted ledger bytes: %v", err)
	}

	t.Run("running native SQLite read", func(t *testing.T) {
		const fixtureMessages = 4096
		started, proceed := make(chan struct{}), make(chan struct{})
		var release sync.Once
		defer release.Do(func() { close(proceed) })
		var armed, nativeFinished atomic.Bool
		var comparisons atomic.Uint64
		collation := fmt.Sprintf("fixture_historical_gate_%d", traceGateID.Add(1))
		if err := sqlite.RegisterCollationUtf8(collation, func(left, right string) int {
			if armed.Load() {
				switch comparisons.Add(1) {
				case 1:
					close(started)
					<-proceed
				case fixtureMessages:
					nativeFinished.Store(true)
				}
				runtime.Gosched()
			}
			return strings.Compare(left, right)
		}); err != nil {
			t.Fatal(err)
		}
		fixtureWriter, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { fixtureWriter.Close() })
		if _, err := fixtureWriter.Exec(fmt.Sprintf(`
			ALTER TABLE message RENAME TO original_message;
			CREATE TABLE message(id TEXT PRIMARY KEY, session_id TEXT COLLATE %[1]s, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO message SELECT * FROM original_message;
			DROP TABLE original_message;
			WITH RECURSIVE fixture(n) AS (VALUES(2) UNION ALL SELECT n+1 FROM fixture WHERE n < %[2]d)
			INSERT INTO message SELECT 'unused-'||n, 'different-owner', n, n, '{"role":"assistant"}' FROM fixture;
		`, collation, fixtureMessages)); err != nil {
			t.Fatal(err)
		}
		// Reopen the read-only pool so its SQLite connections register
		// the fixture collation before validating or reading the schema.
		if err := service.Close(); err != nil {
			t.Fatal(err)
		}
		nativeHook, err := live.OpenReadOnlyHook(path, 0)
		if err != nil {
			t.Fatal(err)
		}
		service = live.NewLiveServiceWithCodexHome(nativeHook, testProjectConfig(), home)
		t.Cleanup(func() { service.Close() })
		handler = adaptershttp.WorkspaceEventsHandler(root, service)
		ctx, cancel := context.WithCancel(guard)
		defer cancel()
		armed.Store(true)
		done := make(chan *httptest.ResponseRecorder, 1)
		go func() { done <- read(ctx) }()
		select {
		case <-started:
		case rec := <-done:
			t.Fatalf("GET returned before native SQLite extraction began: %d %s", rec.Code, rec.Body)
		case <-guard.Done():
			t.Fatal("historical extraction did not reach the native row comparison")
		}
		cancel()
		release.Do(func() { close(proceed) })
		select {
		case rec := <-done:
			if nativeFinished.Load() {
				t.Fatal("SQLite scanned all fixture rows after cancellation")
			}
			var result domain.WorkspaceEventsResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &result); rec.Code != http.StatusOK || err != nil ||
				result.Availability != domain.TracePartial || !slices.Equal(result.Coverage.CompleteSessionIDs, []string{"codex-owner"}) ||
				len(result.Events) != 1 || result.Events[0].CallID != "kept-call" ||
				!strings.Contains(strings.Join(result.Coverage.Diagnostics, " "), context.Canceled.Error()) {
				t.Fatalf("interrupted SQLite query did not preserve usable evidence: status %d, body %s, error %v", rec.Code, rec.Body, err)
			}
		case <-guard.Done():
			t.Fatal("cancelled historical GET held the running SQLite query")
		}
		armed.Store(false)
		rec := read(guard)
		var fresh domain.WorkspaceEventsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &fresh); rec.Code != http.StatusOK || err != nil ||
			fresh.Availability != domain.TraceComplete || len(fresh.Events) != 2 {
			t.Fatalf("historical SQLite extraction failed after cancellation: status %d, body %s, error %v", rec.Code, rec.Body, err)
		}
		after, err := os.ReadFile(ledger)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("interrupted SQLite GET changed persisted ledger bytes: %v", err)
		}
	})

	for _, variant := range []struct {
		name       string
		parent     string
		wantPrefix bool
	}{{"exact interrupted owner", "", true}, {"conflicting native parent", "unrelated", false}} {
		t.Run(variant.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"codex-owner","source":"cli","parent_thread_id":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"kept-call","name":"read","arguments":"{}"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"response_item","payload":{"type":"function_call_output","call_id":"kept-call","output":"accepted result"}}
{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"later-call","name":"read","arguments":"{}"}}
`, variant.parent)
			if err := os.WriteFile(rolloutPath, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(guard)
			defer cancel()
			// The entry check, metadata, call and output have been reached.
			beforeNextRecord := &codexPrefixCancelContext{Context: ctx, cancel: cancel, cancelAt: 5}
			rec := read(beforeNextRecord)
			var response domain.WorkspaceEventsResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &response); rec.Code != http.StatusOK || err != nil ||
				beforeNextRecord.parserChecks < 5 ||
				!slices.Equal(response.Coverage.RequiredSessionIDs, []string{"codex-owner", "sqlite-owner"}) ||
				len(response.Coverage.CompleteSessionIDs) != 0 ||
				!strings.Contains(strings.Join(response.Coverage.Diagnostics, " "), context.Canceled.Error()) {
				t.Fatalf("interruption did not produce incomplete coverage: checks=%d status=%d body=%s error=%v", beforeNextRecord.parserChecks, rec.Code, rec.Body, err)
			}
			if variant.wantPrefix {
				if response.Availability != domain.TracePartial || response.DataSource != domain.DataSourceLocalCodex ||
					len(response.Events) != 1 || response.Events[0].CallID != "kept-call" || response.Events[0].SessionID != "codex-owner" {
					t.Fatalf("interrupted owner lost its completed event or admitted an interrupted record: %+v", response)
				}
			} else if response.Availability != domain.TraceUnavailable || len(response.Events) != 0 ||
				!strings.Contains(strings.Join(response.Coverage.Diagnostics, " "), "source parent contradicts recorded owner") {
				t.Fatalf("contradictory native parent admitted unverified evidence: %+v", response)
			}
			after, err := os.ReadFile(ledger)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("interrupted historical GET changed persisted ledger bytes: %v", err)
			}
		})
	}
}

func TestCancelledHistoricalCodexSpawnRespectsNativeChildParent(t *testing.T) {
	const interruptedOwner = "interrupted-owner"
	for _, tc := range []struct {
		name, parent, wantChild string
	}{
		{"contradictory parent", "other-parent", ""},
		{"matching parent", interruptedOwner, "child"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			sessions := filepath.Join(home, "sessions")
			if err := os.Mkdir(sessions, 0700); err != nil {
				t.Fatal(err)
			}
			child := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"child","source":"cli","parent_thread_id":%q}}`+"\n", tc.parent)
			parent := `{"type":"session_meta","payload":{"id":"interrupted-owner","source":"cli"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":"{\"session_id\":\"child\"}"}}
{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"later","name":"read","arguments":"{}"}}
`
			if err := os.WriteFile(filepath.Join(sessions, "child.jsonl"), []byte(child), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(sessions, "parent.jsonl"), []byte(parent), 0600); err != nil {
				t.Fatal(err)
			}
			ledger := domain.AgenticMetricsLedger{Facts: domain.AgenticFactLedger{
				Sessions: []domain.SessionFact{
					{ID: "child", Source: domain.SourceCodex, ParentID: &tc.parent},
					{ID: interruptedOwner, Source: domain.SourceCodex},
				},
			}}
			service := live.NewLiveServiceWithCodexHome(nil, testProjectConfig(), home)
			t.Cleanup(func() { service.Close() })
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// The child makes three parser checks. In the second rollout,
			// cancel at its fifth check, after the spawn output was read.
			beforeNextRecord := &codexPrefixCancelContext{Context: ctx, cancel: cancel, cancelAt: 8}
			result, err := service.HydrateWorkspaceEvents(beforeNextRecord, "FIX-63", ledger)
			if err != nil || beforeNextRecord.parserChecks < 8 ||
				result.Availability != domain.TracePartial || result.DataSource != domain.DataSourceLocalCodex ||
				!slices.Equal(result.Coverage.CompleteSessionIDs, []string{"child"}) ||
				!strings.Contains(strings.Join(result.Coverage.Diagnostics, " "), context.Canceled.Error()) ||
				len(result.Events) != 1 || result.Events[0].CallID != "spawn" ||
				result.Events[0].Status != "completed" || result.Events[0].SessionID != interruptedOwner {
				t.Fatalf("interrupted rollout lost its completed spawn or certified incomplete evidence: checks=%d response=%+v error=%v", beforeNextRecord.parserChecks, result, err)
			}
			if event := result.Events[0]; event.ChildSessionID != tc.wantChild {
				t.Fatalf("native child parent %q: spawn child = %q, want %q", tc.parent, event.ChildSessionID, tc.wantChild)
			}
		})
	}
}

type treeCacheAdmissionGate struct {
	context.Context
	reached, proceed chan struct{}
	once             sync.Once
}

func (ctx *treeCacheAdmissionGate) Err() error {
	if pc, _, _, ok := runtime.Caller(2); ok && strings.HasSuffix(runtime.FuncForPC(pc).Name(), ".(*LiveService).acceptTreeCache") {
		ctx.once.Do(func() {
			close(ctx.reached)
			<-ctx.proceed
		})
	}
	return ctx.Context.Err()
}

func TestOlderLiveTreeCannotEvictNewerCache(t *testing.T) {
	path := createTestDBFile(t)
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	t.Cleanup(func() { service.Close() })
	guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	gate := &treeCacheAdmissionGate{Context: guard, reached: make(chan struct{}), proceed: make(chan struct{})}
	var released sync.Once
	release := func() { released.Do(func() { close(gate.proceed) }) }
	defer release()
	olderDone := make(chan error, 1)
	go func() {
		tree, err := service.GetSessionTree(gate)
		if err == nil && tree.TotalActiveSessions != 0 {
			err = fmt.Errorf("older tree found %d active sessions", tree.TotalActiveSessions)
		}
		olderDone <- err
	}()
	select {
	case <-gate.reached:
	case err := <-olderDone:
		t.Fatalf("older tree returned before staged cache admission: %v", err)
	case <-guard.Done():
		t.Fatal("older tree did not reach cache admission")
	}
	now := time.Now().UnixMilli()
	if _, err := writer.Exec("INSERT INTO session(id,title,time_created,time_updated) VALUES ('new','new',?,?)", now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Exec(`INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES ('newmsg','new',?,?,'{"role":"assistant","content":"new"}')`, now, now); err != nil {
		t.Fatal(err)
	}
	tree, err := service.GetSessionTree(guard)
	if err != nil || tree.TotalActiveSessions != 1 {
		t.Fatalf("newer tree was not accepted: %+v, %v", tree, err)
	}
	before, err := service.GetSessionTrace(guard, "new")
	if err != nil || len(before.Messages) != 1 {
		t.Fatalf("newer snapshot was unavailable: %+v, %v", before, err)
	}
	release()
	select {
	case err := <-olderDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-guard.Done():
		t.Fatal("older tree did not return")
	}
	after, err := service.GetSessionTrace(guard, "new")
	if err != nil || len(after.Messages) != 1 || &before.Messages[0] != &after.Messages[0] {
		t.Fatalf("older membership evicted the newer trace at its accepted source revision: %+v, %v", after, err)
	}
}

func TestLiveSessionListingCancellationInterruptsSQLite(t *testing.T) {
	path := createTestDBFile(t)
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	now := time.Now().UnixMilli()
	if _, err := writer.Exec("INSERT INTO session(id,title,time_created,time_updated) VALUES('active','active',?,?)", now, now); err != nil {
		t.Fatal(err)
	}
	started, proceed := make(chan struct{}), make(chan struct{})
	var entered, released sync.Once
	release := func() { released.Do(func() { close(proceed) }) }
	defer release()
	var armed, nativeFinished atomic.Bool
	function := fmt.Sprintf("fixture_active_sessions_gate_%d", traceGateID.Add(1))
	if err := sqlite.RegisterScalarFunction(function, 1, func(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
		if armed.Load() {
			switch args[0].(int64) {
			case 0:
				entered.Do(func() { close(started) })
				<-proceed
			case -1:
				nativeFinished.Store(true)
			default:
				runtime.Gosched()
			}
		}
		return armed.Load(), nil
	}); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveService(hook, testProjectConfig())
	t.Cleanup(func() { service.Close() })
	if _, err := writer.Exec(fmt.Sprintf(`
		ALTER TABLE session RENAME TO fixture_session;
		CREATE VIEW session AS SELECT id,parent_id,agent,model,
		(WITH RECURSIVE scan(n) AS (VALUES(0) UNION ALL SELECT n+1 FROM scan WHERE n<10000 AND %[1]s(n))
		 SELECT CASE WHEN %[1]s(-1) THEN title ELSE title END FROM scan ORDER BY n DESC LIMIT 1) AS title,
		time_created,time_updated FROM fixture_session;
	`, function)); err != nil {
		t.Fatal(err)
	}

	guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	ctx, cancel := context.WithCancel(guard)
	defer cancel()
	armed.Store(true)
	done := make(chan error, 1)
	go func() {
		_, err := service.GetSessionTree(ctx)
		done <- err
	}()
	select {
	case <-started:
	case err := <-done:
		t.Fatalf("listing returned before starting its native query: %v", err)
	case <-guard.Done():
		t.Fatal("listing did not enter the native SQLite step")
	}
	cancel()
	release()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) || nativeFinished.Load() {
			t.Fatalf("native active listing completed despite cancellation: %v", err)
		}
	case <-guard.Done():
		t.Fatal("cancelled listing still holds its native read")
	}
	armed.Store(false)
	tree, err := service.GetSessionTree(guard)
	if err != nil || tree.TotalActiveSessions != 1 {
		t.Fatalf("new active listing did not recover: %+v, %v", tree, err)
	}
}
