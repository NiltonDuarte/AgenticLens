package http_test

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

// pausedDiagnosticWriter stalls the HTTP transport only after the client has
// received one diagnostic. The service and its SQLite reads remain real.
type pausedDiagnosticWriter struct {
	http.ResponseWriter
	ctx         context.Context
	paused      chan struct{}
	resume      chan struct{}
	diagnostics int
}

func (w *pausedDiagnosticWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *pausedDiagnosticWriter) Flush()                      { w.ResponseWriter.(http.Flusher).Flush() }

func (w *pausedDiagnosticWriter) Write(data []byte) (int, error) {
	if strings.HasPrefix(string(data), "event: source_diagnostic\n") {
		w.diagnostics++
		if w.diagnostics == 2 {
			close(w.paused)
			select {
			case <-w.resume:
			case <-w.ctx.Done():
				return 0, w.ctx.Err()
			}
		}
	}
	return w.ResponseWriter.Write(data)
}

func liveStreamSQLiteFixture(t *testing.T) (*sql.DB, *live.LiveService) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "live.db")
	writer, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := writer.Close(); err != nil {
			t.Errorf("close live fixture writer: %v", err)
		}
	})
	writer.SetMaxOpenConns(1)
	if _, err := writer.Exec(`PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=1000;
CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,data TEXT);`); err != nil {
		t.Fatal(err)
	}
	hook, err := live.OpenReadOnlyHook(dbPath, 1000)
	if err != nil {
		t.Fatal(err)
	}
	cfg := httpFixturePolicy()
	cfg.Live.PollIntervalMs = 10
	service := live.NewLiveService(hook, cfg)
	t.Cleanup(func() {
		if err := service.Close(); err != nil {
			t.Errorf("close live fixture service: %v", err)
		}
		// Closing the pool prevents further reads. A poll already using a
		// connection can still touch WAL files until the driver releases it.
		ticks := time.NewTicker(time.Millisecond)
		defer ticks.Stop()
		deadline := time.NewTimer(3 * time.Second)
		defer deadline.Stop()
		for open := hook.DB().Stats().OpenConnections; open > 0; open = hook.DB().Stats().OpenConnections {
			select {
			case <-ticks.C:
			case <-deadline.C:
				t.Errorf("live fixture still has %d reader connections after close", open)
				return
			}
		}
	})
	return writer, service
}

func TestLiveStreamReconnectsAfterSubscriberOverflow(t *testing.T) {
	writer, service := liveStreamSQLiteFixture(t)
	fast, cleanup := service.Subscribe()
	defer cleanup()
	paused, resume := make(chan struct{}), make(chan struct{})
	stream := adaptershttp.LiveStreamHandler(service)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			w = &pausedDiagnosticWriter{ResponseWriter: w, ctx: r.Context(), paused: paused, resume: resume}
		}
		stream.ServeHTTP(w, r)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	openStream := func(path string) (*http.Response, *bufio.Scanner) {
		t.Helper()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { response.Body.Close() })
		if response.StatusCode != http.StatusOK {
			t.Fatalf("stream %s returned HTTP %d", path, response.StatusCode)
		}
		return response, bufio.NewScanner(response.Body)
	}
	readEvent := func(scanner *bufio.Scanner, want live.LiveEventType) string {
		t.Helper()
		for scanner.Scan() {
			if scanner.Text() == "event: "+string(want) && scanner.Scan() && strings.HasPrefix(scanner.Text(), "data: ") {
				return strings.TrimPrefix(scanner.Text(), "data: ")
			}
		}
		t.Fatalf("stream ended before %s: %v", want, scanner.Err())
		return ""
	}
	readTree := func(scanner *bufio.Scanner) domain.LiveSessionTree {
		t.Helper()
		var tree domain.LiveSessionTree
		if err := json.Unmarshal([]byte(readEvent(scanner, live.LiveEventSessionTree)), &tree); err != nil {
			t.Fatal(err)
		}
		tree.LastRefreshedAt = ""
		return tree
	}
	readFast := func() live.LiveEventType {
		t.Helper()
		select {
		case event, ok := <-fast:
			if !ok {
				t.Fatal("responsive subscriber disconnected during slow subscriber overflow")
			}
			return event.Type
		case <-ctx.Done():
			t.Fatalf("responsive subscriber made no progress: %v", ctx.Err())
			return ""
		}
	}

	slow, scanner := openStream("/slow")
	initial := readTree(scanner)
	if err := service.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if event := readFast(); event != live.LiveEventSessionTree {
		t.Fatalf("responsive subscriber missed initial tree: %s", event)
	}
	if _, err := writer.Exec("ALTER TABLE session RENAME TO unavailable"); err != nil {
		t.Fatal(err)
	}
	if diagnostic := readEvent(scanner, live.LiveEventSourceDiagnostic); !strings.Contains(diagnostic, "database_error") {
		t.Fatalf("slow subscriber missed source failure: %s", diagnostic)
	}
	select {
	case <-paused:
	case <-ctx.Done():
		t.Fatalf("stream transport did not pause: %v", ctx.Err())
	}
	// One diagnostic reached the slow client and a second is blocked in its
	// transport. The remaining capacity fills and the next event overflows.
	for i := 0; i < cap(fast)+3; i++ {
		if event := readFast(); event != live.LiveEventSourceDiagnostic {
			t.Fatalf("responsive subscriber missed failure diagnostic: %s", event)
		}
	}
	if _, err := writer.Exec("ALTER TABLE unavailable RENAME TO session"); err != nil {
		t.Fatal(err)
	}
	for {
		event := readFast()
		if event == live.LiveEventSessionTree {
			break
		}
		if event != live.LiveEventSourceDiagnostic {
			t.Fatalf("responsive subscriber missed recovery tree: %s", event)
		}
	}
	close(resume)
	for scanner.Scan() {
	}
	if scanner.Err() != nil || ctx.Err() != nil {
		t.Fatalf("slow stream stayed open after overflow instead of allowing reconnect: %v", scanner.Err())
	}
	slow.Body.Close()
	_, reconnected := openStream("/reconnect")
	if recovered := readTree(reconnected); !reflect.DeepEqual(recovered, initial) {
		t.Fatalf("new stream did not recover the unchanged initial tree: %+v vs %+v", recovered, initial)
	}
}
