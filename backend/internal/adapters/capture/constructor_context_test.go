package capture

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestRegistryConstructorSharesShortHandoffDeadlineWithSQLiteContention(t *testing.T) {
	first, second, base := sharedWorktrees(t)
	registry, err := ProjectState(base, first)
	if err != nil {
		t.Fatal(err)
	}
	store, err := OpenProject(registry, first)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	uri, err := sqliteFileURI(filepath.Join(registry, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	writer, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if _, err = writer.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	held := true
	defer func() {
		if held {
			if _, e := writer.ExecContext(context.Background(), "ROLLBACK"); e != nil {
				t.Error(e)
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	started := time.Now()
	admitted, openErr := OpenProjectContext(ctx, registry, second)
	elapsed := time.Since(started)
	if admitted != nil {
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "not-accepted-after-timeout"}
		_, registrationErr := admitted.Register(ctx, identity, "", "SessionStart")
		closeErr := admitted.Close()
		t.Fatalf("constructor admitted a scope behind held writer: open=%v registration=%v close=%v", openErr, registrationErr, closeErr)
	}
	if openErr == nil {
		t.Fatal("short constructor deadline reported success")
	}
	// Wide separation catches the old unconditional one-second SQLite wait. It is
	// a held-writer oracle, not a universal scheduling/filesystem latency promise.
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("constructor added its own busy budget: elapsed=%s error=%v", elapsed, openErr)
	}
	var scopes, registrations int
	if err = writer.QueryRowContext(context.Background(), "SELECT count(*) FROM worktrees WHERE directory=?", second).Scan(&scopes); err != nil {
		t.Fatal(err)
	}
	if err = writer.QueryRowContext(context.Background(), "SELECT count(*) FROM registrations WHERE session_id='not-accepted-after-timeout'").Scan(&registrations); err != nil {
		t.Fatal(err)
	}
	if scopes != 0 || registrations != 0 {
		t.Fatalf("expired constructor accepted work: scopes=%d registrations=%d", scopes, registrations)
	}
	if _, err = writer.ExecContext(context.Background(), "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	held = false
	active, err := OpenProjectContext(context.Background(), registry, second)
	if err != nil {
		t.Fatalf("same state failed after writer release: %v", err)
	}
	if err = active.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredRegistryConstructorDoesNotAllocateState(t *testing.T) {
	first, _, base := sharedWorktrees(t)
	registry, err := ProjectState(base, first)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store, err := OpenProjectContext(ctx, registry, first)
	if store != nil {
		store.Close()
		t.Fatal("expired constructor returned store")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expired context lost: %v", err)
	}
	if _, err = os.Stat(base); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expired constructor allocated state: %v", err)
	}
}
