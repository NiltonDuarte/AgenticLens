package capture

import (
	"context"
	"errors"
	"testing"
)

func TestOpenCodeRestoreRejectsDiscoveryCursorFromBeforeReset(t *testing.T) {
	ctx := context.Background()
	store, _, observe := openCodeControlFixture(t)
	root := observe("root")
	id := StableCaptureID(root.Identity.Source, root.Identity.SessionID)
	key := IdentityKey(root.Identity)
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	before, err := store.captureControlRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.saveOpenCodeChildCursor(ctx, key, "child", true, before); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCaptureDiscarded(ctx, id, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := store.saveOpenCodeChildCursor(ctx, key, "late-child", false, before); !errors.Is(err, ErrCaptureLifecycleChanged) {
		t.Fatalf("pre-restore discovery rewrote the reset cursor: %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT count(*) FROM child_cursor WHERE worktree_id=? AND kind='opencode' AND parent=?`, store.worktreeID, key).Scan(&count); err != nil || count != 0 {
		t.Fatalf("old discovery left cursor evidence after reset: count=%d, %v", count, err)
	}
	after, err := store.captureControlRevision(ctx)
	if err != nil || after <= before {
		t.Fatalf("restore did not advance control revision: %d -> %d, %v", before, after, err)
	}
	if err := store.saveOpenCodeChildCursor(ctx, key, "child", true, after); err != nil {
		t.Fatal(err)
	}
	if position, err := store.childCursor(ctx, "opencode", key); err != nil || position != "child" {
		t.Fatalf("current restore discovery cursor did not advance: %q, %v", position, err)
	}
}
