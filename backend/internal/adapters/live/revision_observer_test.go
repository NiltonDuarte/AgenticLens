package live

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestSourceRevisionReopensAfterActualObserverFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "native.sqlite")
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { writer.Close() })
	if _, err := writer.Exec("CREATE TABLE native (value TEXT)"); err != nil {
		t.Fatal(err)
	}
	hook, err := OpenReadOnlyHook(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { hook.Close() })
	ctx := context.Background()
	before, err := hook.SourceRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := hook.revisionConn.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := hook.SourceRevision(ctx); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("closed source observer returned %v", err)
	}
	after, err := hook.SourceRevision(ctx)
	if err != nil || before.Connection == after.Connection {
		t.Fatalf("failed observer did not recover with a new revision identity: before=%v after=%v error=%v", before, after, err)
	}
}
