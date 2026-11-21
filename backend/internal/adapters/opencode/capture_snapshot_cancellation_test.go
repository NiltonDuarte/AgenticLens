package opencode_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"modernc.org/sqlite"
)

var snapshotCollationID atomic.Uint64

func TestExactSnapshotCancellationReleasesTransaction(t *testing.T) {
	const fixtureMessages = 4096
	const v1 = `
		CREATE TABLE session(id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		CREATE TABLE message(id TEXT PRIMARY KEY, session_id TEXT COLLATE %[1]s, time_created INTEGER, time_updated INTEGER, data TEXT);
		CREATE TABLE part(id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		INSERT INTO session VALUES ('owner', NULL, 'coder', 'm', 'fixture', 1000, 2000);
		INSERT INTO message VALUES ('message', 'owner', 1000, 1000, '{"role":"assistant","modelID":"m","content":"fixture"}');
		WITH RECURSIVE fixture(n) AS (VALUES(2) UNION ALL SELECT n+1 FROM fixture WHERE n < %[2]d)
		INSERT INTO message SELECT 'unused-'||n, 'different-owner', n, n, '{}' FROM fixture;`
	const v2 = `
		CREATE TABLE session_v2(id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		CREATE TABLE session_message(id TEXT PRIMARY KEY, session_id TEXT COLLATE %[1]s, type TEXT, seq INTEGER, time_created INTEGER, time_updated INTEGER, data TEXT);
		INSERT INTO session_v2 VALUES ('owner', NULL, 'coder', 'm', 'fixture', 1000, 2000);
		INSERT INTO session_message VALUES ('message', 'owner', 'user', 1, 1000, 1000, '{"text":"fixture"}');
		WITH RECURSIVE fixture(n) AS (VALUES(2) UNION ALL SELECT n+1 FROM fixture WHERE n < %[2]d)
		INSERT INTO session_message SELECT 'unused-'||n, 'different-owner', 'user', n, n, n, '{}' FROM fixture;`
	for _, fixture := range []struct{ name, schema string }{{"v1", v1}, {"v2", v2}} {
		t.Run(fixture.name, func(t *testing.T) {
			entered, proceed := make(chan struct{}), make(chan struct{})
			var release sync.Once
			defer release.Do(func() { close(proceed) })
			var comparisons atomic.Uint64
			name := fmt.Sprintf("fixture_snapshot_cancel_%d", snapshotCollationID.Add(1))
			if err := sqlite.RegisterCollationUtf8(name, func(left, right string) int {
				// Every message participates in the unindexed byte preflight.
				// The next comparison is the start of exact body extraction.
				switch n := comparisons.Add(1); {
				case n == fixtureMessages+1:
					close(entered)
					<-proceed
				case n > fixtureMessages+1 && n < 2*fixtureMessages:
					runtime.Gosched()
				}
				return strings.Compare(left, right)
			}); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "source.sqlite")
			writer, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			_, setupErr := writer.Exec(fmt.Sprintf(fixture.schema, name, fixtureMessages))
			if err := errors.Join(setupErr, writer.Close()); err != nil {
				t.Fatal(err)
			}
			source, err := opencode.OpenDB(path, 0)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { source.Close() })
			db := source.DB()
			db.SetMaxOpenConns(1)
			guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			ctx, cancel := context.WithCancel(guard)
			defer cancel()
			evidence := domain.SourceEvents{CompleteSessionIDs: []string{"stored-owner"}}
			type outcome struct {
				count int
				err   error
			}
			done := make(chan outcome, 1)
			go func() {
				sessions, _, _, err := opencode.CaptureSessionSnapshot(ctx, db, "owner", "tester", 1<<20, &evidence)
				done <- outcome{len(sessions), err}
			}()
			select {
			case <-entered:
				// The native body read has entered SQLite with the exact transaction still open.
			case result := <-done:
				t.Fatalf("snapshot returned before starting body extraction: %+v", result)
			case <-guard.Done():
				t.Fatal("snapshot never entered body extraction")
			}
			cancel()
			release.Do(func() { close(proceed) })
			select {
			case result := <-done:
				if comparisons.Load() >= 2*fixtureMessages {
					t.Fatal("SQLite scanned all fixture rows after cancellation")
				}
				if !errors.Is(result.err, context.Canceled) || result.count != 0 ||
					!slices.Equal(evidence.CompleteSessionIDs, []string{"stored-owner"}) || len(evidence.Events) != 0 {
					t.Fatalf("cancelled snapshot returned facts or changed prior evidence: %+v, evidence=%+v", result, evidence)
				}
			case <-guard.Done():
				t.Fatal("cancelled snapshot did not release its native transaction")
			}

			sessions, _, _, err := opencode.CaptureSessionSnapshot(guard, db, "owner", "tester", 1<<20, &evidence)
			if err != nil || len(sessions) != 1 || sessions[0].ID != "owner" ||
				!slices.Contains(evidence.CompleteSessionIDs, "owner") || !slices.Contains(evidence.CompleteSessionIDs, "stored-owner") ||
				len(evidence.Events) != 1 {
				t.Fatalf("retry on the same one-connection pool did not recover: sessions=%+v, evidence=%+v, error=%v", sessions, evidence, err)
			}
		})
	}
}
