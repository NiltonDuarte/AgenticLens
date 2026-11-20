package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

func TestCaptureOrdinarySourceWriteRejectsSupersededRestoreObservation(t *testing.T) {
	ctx := context.Background()
	store, native, observed, _ := openCodeRestoreSourceFixture(t)
	oldRevision, err := store.captureControlRevision(ctx)
	if err != nil || openCodeRestoreSourceInput(t, observed.Snapshot) != 17 {
		t.Fatalf("fixture did not establish an initial read: %v", err)
	}
	id := StableCaptureID(observed.Identity.Source, observed.Identity.SessionID)
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCaptureDiscarded(ctx, id, false, nil); err != nil {
		t.Fatal(err)
	}
	readLatest := func(input int64, updated int) domain.CaptureRegistration {
		t.Helper()
		if err := writeOpenCodeRestoreSourceMessage(native, observed.Identity.SessionID, input, updated); err != nil {
			t.Fatal(err)
		}
		reconcileOpenCodeRestoreSource(t, store, observed.TicketID, 16)
		row := openCodeRestoreSourceRegistration(t, store, observed.Identity)
		if row.SourceFingerprint == "" || openCodeRestoreSourceInput(t, row.Snapshot) != input {
			t.Fatalf("fixture did not accept %d input tokens from the new source", input)
		}
		return row
	}
	rejectStale := func(old, newer domain.CaptureRegistration, revision int64, expected error) {
		t.Helper()
		applied := false
		_, err := store.saveCaptureRead(ctx, old, revision, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			applied = true
			current.Snapshot = old.Snapshot
			current.SourceFingerprint = old.SourceFingerprint
			return current, nil
		})
		if applied || !errors.Is(err, expected) {
			t.Fatalf("stale ordinary source writer reached its update: applied=%t err=%v, want %v", applied, err, expected)
		}
		if actual := openCodeRestoreSourceRegistration(t, store, observed.Identity); !reflect.DeepEqual(actual, newer) {
			t.Fatal("rejected writer changed the accepted source record")
		}
	}

	first := readLatest(41, 4000)
	rejectStale(observed, first, oldRevision, ErrCaptureLifecycleChanged)
	currentRevision, err := store.captureControlRevision(ctx)
	if err != nil || currentRevision <= oldRevision {
		t.Fatalf("restore control did not advance its revision: %d -> %d, %v", oldRevision, currentRevision, err)
	}
	latest := readLatest(73, 5000)
	rejectStale(first, latest, currentRevision, errCaptureReadObservationChanged)
}

func TestOpenCodeRestoreRechecksCoveredOwnerAfterSilentSourceChange(t *testing.T) {
	ctx := context.Background()
	store, native, root, child := openCodeRestoreSourceFixture(t)
	id := StableCaptureID(root.Identity.Source, root.Identity.SessionID)
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	restore, err := store.SetCaptureDiscarded(ctx, id, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	controlRevision, err := store.captureControlRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Both exact native child scans finish before the deliberately one-owner reads.
	budget := &workBudget{sessions: 8, bytes: 1 << 20, controlRevision: controlRevision}
	if _, err := store.recoverChildren(ctx, root.TicketID, budget, workOpenCode); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE reconcile_cursor SET identity=? WHERE worktree_id=? AND ticket=?`, IdentityKey(child.Identity), store.worktreeID, root.TicketID); err != nil {
		t.Fatal(err)
	}
	readNext := func(wanted string, input int64) ReconcileResult {
		t.Helper()
		if _, err := store.db.ExecContext(ctx, `UPDATE recovery_phase SET phase=0 WHERE worktree_id=? AND ticket=?`, store.worktreeID, root.TicketID); err != nil {
			t.Fatal(err)
		}
		result := reconcileOpenCodeRestoreSource(t, store, root.TicketID, 1)
		if len(result.RefreshedSessions) != 1 || result.RefreshedSessions[0].ID != wanted || openCodeRestoreSourceInput(t, result.RefreshedSessions) != input {
			t.Fatalf("one-owner pass should have refreshed only %s with %d input tokens, refreshed=%+v", wanted, input, result.RefreshedSessions)
		}
		return result
	}
	first := readNext("root", 17)
	if len(first.RestoresReady) != 0 || !openCodeRestoreReadReady(t, store, restore, root.Identity) || openCodeRestoreReadReady(t, store, restore, child.Identity) {
		t.Fatal("first pass did not commit only the root's restore receipt")
	}
	covered := openCodeRestoreSourceRegistration(t, store, root.Identity)
	if err := writeOpenCodeRestoreSourceMessage(native, "late-child", 31, 3500); err != nil {
		t.Fatal(err)
	}
	if !store.restoreReadStillCurrent(ctx, restore, covered) {
		t.Fatal("an unrelated native session changed the covered root's receipt")
	}
	// No callback or registry mutation accompanies this exact-owner source edit.
	if err := writeOpenCodeRestoreSourceMessage(native, "root", 41, 4000); err != nil {
		t.Fatal(err)
	}
	second := readNext("child", 23)
	if len(second.RestoresReady) != 0 || len(second.RestoredSessions) != 0 || openCodeRestoreReadReady(t, store, restore, root.Identity) || !openCodeRestoreReadReady(t, store, restore, child.Identity) {
		t.Fatal("restore accepted the old root receipt when the remaining child became covered")
	}
	if current := openCodeRestoreSourceRegistration(t, store, root.Identity); current.CallbackObservedAt != covered.CallbackObservedAt || openCodeRestoreSourceInput(t, current.Snapshot) != 17 {
		t.Fatal("test must invalidate coverage without deleting the accepted snapshot or adding a callback")
	}
	last := readNext("root", 41)
	if len(last.RestoresReady) != 1 || last.RestoresReady[0].Revision != restore.Revision || len(last.RestoredSessions) != 2 {
		t.Fatalf("fresh root read did not complete the exact restore: %+v", last.RestoresReady)
	}
	for _, owner := range last.RestoredSessions {
		want := map[string]int64{"root": 41, "child": 23}[owner.ID]
		if want == 0 || openCodeRestoreSourceInput(t, []domain.SessionFact{owner}) != want {
			t.Fatalf("restore carried stale or unrelated facts for %s", owner.ID)
		}
	}
}

func openCodeRestoreSourceFixture(t *testing.T) (*Store, *sql.DB, domain.CaptureRegistration, domain.CaptureRegistration) {
	t.Helper()
	store, native, observe := openCodeControlFixture(t)
	for id, input := range map[string]int64{"root": 17, "child": 23} {
		if err := writeOpenCodeRestoreSourceMessage(native, id, input, 3000); err != nil {
			t.Fatal(err)
		}
	}
	root := observe("root")
	bound, err := store.BindNativeOpenCode(context.Background(), root.Identity, BindingTicket{ID: "T-100", Title: "Restore source fixture", Type: "fix"}, root.Request)
	if err != nil {
		t.Fatal(err)
	}
	child := observe("child")
	if child.TicketID != bound.TicketID || child.ParentID != bound.Identity.SessionID {
		t.Fatal("fixture child did not inherit verified root ownership")
	}
	result := reconcileOpenCodeRestoreSource(t, store, bound.TicketID, 16)
	if len(result.RefreshedSessions) != 2 {
		t.Fatalf("initial source fixture refreshed %d owners, want 2", len(result.RefreshedSessions))
	}
	return store, native, openCodeRestoreSourceRegistration(t, store, bound.Identity), openCodeRestoreSourceRegistration(t, store, child.Identity)
}

func writeOpenCodeRestoreSourceMessage(native *sql.DB, id string, input int64, updated int) error {
	body := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":1500,"completed":3000},"tokens":{"input":%d,"output":2}}`, input)
	_, err := native.Exec(`INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES (?,?,1500,?,?) ON CONFLICT(id) DO UPDATE SET time_updated=excluded.time_updated,data=excluded.data`, "m-"+id, id, updated, body)
	return err
}

func reconcileOpenCodeRestoreSource(t *testing.T, store *Store, ticket string, budget int) ReconcileResult {
	t.Helper()
	result, err := store.Reconcile(context.Background(), ticket, "tester", budget, 1<<20, store.configuration.Generation.Metrics)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func openCodeRestoreSourceRegistration(t *testing.T, store *Store, identity domain.CaptureIdentity) domain.CaptureRegistration {
	t.Helper()
	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Identity == identity {
			return row
		}
	}
	t.Fatalf("registration for %s was not retained", identity.SessionID)
	return domain.CaptureRegistration{}
}

func openCodeRestoreReadReady(t *testing.T, store *Store, restore CaptureControl, identity domain.CaptureIdentity) bool {
	t.Helper()
	progress, err := store.CaptureRestoreProgress(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return progress[restore.CaptureID].Revision == restore.Revision && progress[restore.CaptureID].Owners[IdentityKey(identity)].readReady
}

func openCodeRestoreSourceInput(t *testing.T, snapshot []domain.SessionFact) int64 {
	t.Helper()
	if len(snapshot) != 1 || len(snapshot[0].Invocations) != 1 {
		t.Fatalf("expected exactly one source invocation, got %+v", snapshot)
	}
	return snapshot[0].Invocations[0].Tokens.Input
}
