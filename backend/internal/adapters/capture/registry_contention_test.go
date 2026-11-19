package capture_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestRegistryContentionUsesRealSQLiteErrorAndRejectsOtherCauses(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, "CREATE TABLE fixture(id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	owner, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if _, err := owner.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = owner.ExecContext(ctx, "ROLLBACK") }()
	contender, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer contender.Close()
	if _, err := contender.ExecContext(ctx, "PRAGMA busy_timeout=0"); err != nil {
		t.Fatal(err)
	}
	_, busy := contender.ExecContext(ctx, "INSERT INTO fixture VALUES (1)")
	if busy == nil || !capture.IsRegistryContention(busy) {
		t.Fatalf("held SQLite writer did not produce recognized contention: %v", busy)
	}
	wrapped := fmt.Errorf("capture registry write: %w", busy)
	if !capture.IsRegistryContention(errors.Join(wrapped, busy)) {
		t.Fatal("wrapped and joined SQLite contention lost eligibility for bounded retry")
	}
	for _, nonRetryable := range []error{nil, errors.New(busy.Error()), context.DeadlineExceeded, errors.Join(wrapped, context.Canceled)} {
		if capture.IsRegistryContention(nonRetryable) {
			t.Fatalf("a permanent or untyped cause was classified as SQLite contention: %v", nonRetryable)
		}
	}
	if _, err := owner.ExecContext(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if _, err := contender.ExecContext(ctx, "INSERT INTO fixture VALUES (1)"); err != nil {
		t.Fatalf("registry write did not recover after the real lock released: %v", err)
	}
}
