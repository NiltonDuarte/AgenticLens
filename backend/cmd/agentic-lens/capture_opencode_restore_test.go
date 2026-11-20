package main_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCaptureOpenCodeIdleRestoreContinuesBoundedDescendantDiscovery(t *testing.T) {
	f, native, store, root := prepareOpenCodeRestoreFixture(t, "", 1)
	id := capture.StableCaptureID(root.Identity.Source, root.Identity.SessionID)
	assertCaptureRootInput(t, f.ledger(t), 17)

	// These completed descendants are written after discard without invoking a
	// callback. Each budgeted pass may discover or refresh at most one owner.
	addOpenCodeRestoreSession(t, native, f.project, "child", "root", 23)
	addOpenCodeRestoreSession(t, native, f.project, "grandchild", "child", 29)
	addOpenCodeRestoreTask(t, native, "root", "child")
	addOpenCodeRestoreTask(t, native, "child", "grandchild")
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].Identity != root.Identity {
		t.Fatalf("descendants were already registered before restore: %d rows, %v", len(rows), err)
	}

	restoreOpenCodeFixtureAndWait(t, f, store, id)
	assertOpenCodeRestoredOwners(t, f.ledger(t), map[string]openCodeRestoreOwner{
		"root":       {input: 17},
		"child":      {parent: "root", input: 23},
		"grandchild": {parent: "child", input: 29},
	})
}

func TestCaptureOpenCodeIndependentRestoreKeepsUnavailableUpstreamParent(t *testing.T) {
	const upstream = "unavailable-parent"
	f, _, store, root := prepareOpenCodeRestoreFixture(t, upstream, 8)
	if root.ParentID != upstream || root.Binding != "explicit_agent_response" || root.TicketID != initialPublicationTicket {
		t.Fatalf("fixture lacks an admitted independent native root: %+v", root)
	}
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].Identity != root.Identity {
		t.Fatalf("upstream owner must remain unavailable in the local registry: %d rows, %v", len(rows), err)
	}
	restoreOpenCodeFixtureAndWait(t, f, store, capture.StableCaptureID(root.Identity.Source, root.Identity.SessionID))
	assertOpenCodeRestoredOwners(t, f.ledger(t), map[string]openCodeRestoreOwner{
		"root": {parent: upstream, input: 17},
	})
}

func prepareOpenCodeRestoreFixture(t *testing.T, parent string, restoreBudget int) (captureGenerationFixture, *sql.DB, *capture.Store, domain.CaptureRegistration) {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	enableCLIFixtureSummary(t, f.project)
	f.output = seedCaptureTicketHome(t, f, initialPublicationTicket)
	path := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", path)
	setOpenCodeFixtureProject(t, path, f.project)
	settings := func(maxSessions int) {
		t.Helper()
		policy := fmt.Sprintf("telemetry:\n  default_source: opencode\n  sources:\n    opencode:\n      database_path: %q\n  capture:\n    refresh_ms: 100\n    idle_ms: 100\n    max_sessions: %d\n    max_lifetime_ms: 6000\n", path, maxSessions)
		if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
			t.Fatal(err)
		}
	}
	settings(32)
	native, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := native.Close(); err != nil {
			t.Error(err)
		}
	})
	addOpenCodeRestoreSession(t, native, f.project, "root", parent, 17)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: "root"}
	row, err := store.Register(ctx, identity, "", "session.idle")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.BindNativeOpenCode(ctx, identity, capture.BindingTicket{ID: initialPublicationTicket, Title: "Idle restore fixture", Type: "fix"}, row.Request); err != nil {
		t.Fatal(err)
	}
	runDiscardFixtureWorker(t, f)
	rows, err := store.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].Identity != identity || rows[0].ParentID != parent || domain.CaptureNeedsPeriodicRefresh(rows[0]) {
		t.Fatalf("fixture must start with one verified idle native root: %d rows, %v", len(rows), err)
	}
	root := rows[0]
	if _, err := store.SetCaptureDiscarded(ctx, capture.StableCaptureID(identity.Source, identity.SessionID), true, nil); err != nil {
		t.Fatal(err)
	}
	runDiscardFixtureWorker(t, f)
	settings(restoreBudget)
	return f, native, store, root
}

func restoreOpenCodeFixtureAndWait(t *testing.T, f captureGenerationFixture, store *capture.Store, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	out, err := runCaptureNative(t, f.project, "", "capture", "restore", "--repo", f.project, "--capture-id", id)
	if err != nil {
		t.Fatalf("operator restore: %v\n%s", err, out)
	}
	var queued capture.CaptureControl
	if err := json.Unmarshal(out, &queued); err != nil || queued.CaptureID != id || queued.State != capture.CaptureControlRestoring || !queued.CoveragePending {
		t.Fatalf("restore did not queue an exact coverage check: %+v, %v\n%s", queued, err, out)
	}
	settled := completion(ctx)
	status := readDiscardFixtureStatus(t, f, id)
	request, err := store.PublicationRequest(context.Background())
	if settled != nil || err != nil || status.State != capture.CaptureControlActive || status.CoveragePending || status.Revision != queued.Revision || request.Generation != request.Acknowledged {
		rows, listErr := store.List(context.Background())
		owners := make(map[string]string, len(rows))
		for _, row := range rows {
			owners[row.Identity.SessionID] = string(row.State) + ": " + row.Detail
		}
		restores, proofErr := store.CaptureRestoreProgress(context.Background())
		proofs := make(map[string]capture.CaptureRestoreOwnerProgress, len(restores[id].Owners))
		for _, row := range rows {
			if proof, ok := restores[id].Owners[capture.IdentityKey(row.Identity)]; ok {
				proofs[row.Identity.SessionID] = proof
			}
		}
		diagnostics, diagnosticErr := store.Diagnostics(context.Background())
		var attempts []string
		for _, diagnostic := range diagnostics {
			if r := diagnostic.Reconciliation; diagnostic.TicketID == queued.TicketID && r != nil {
				attempts = append(attempts, fmt.Sprintf("%d:%s attempts=%d discovery_pending=%t pending_children=%d", diagnostic.Sequence, r.Attempt, r.SourceAttempts, r.DiscoveryPending, r.PendingChildren))
			}
		}
		progress, progressErr := capture.WorkerStatus(store.CoordinationDirectory())
		t.Fatalf("restore did not finish on the original wake: status=%+v request=%+v read=%v completion=%v; owners=%v (%v); proofs=%v (%v); retained passes=%v (%v); worker=%+v (%v)", status, request, err, settled, owners, listErr, proofs, proofErr, attempts, diagnosticErr, progress, progressErr)
	}
}

func addOpenCodeRestoreSession(t *testing.T, native *sql.DB, project, id, parent string, input int) {
	t.Helper()
	if _, err := native.Exec(`INSERT INTO session(id,parent_id,directory,agent,model,time_created,time_updated) VALUES (?,NULLIF(?,''),?,'build','fixture',1000,3000)`, id, parent, project); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":1000,"completed":3000},"tokens":{"input":%d,"output":2}}`, input)
	if _, err := native.Exec(`INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES (?,?,1000,3000,?)`, "m-"+id, id, body); err != nil {
		t.Fatal(err)
	}
}

func addOpenCodeRestoreTask(t *testing.T, native *sql.DB, parent, child string) {
	t.Helper()
	output := fmt.Sprintf(`{"session_id":%q}`, child)
	body := fmt.Sprintf(`{"type":"tool","tool":"task","state":{"status":"completed","time":{"start":1200,"end":2400},"input":{"subagent_type":"tester","description":"Read the bounded fixture"},"output":%q,"metadata":{"sessionId":%q}}}`, output, child)
	if _, err := native.Exec(`INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES (?,?,?,1200,2400,?)`, "task-"+child, parent, "m-"+parent, body); err != nil {
		t.Fatal(err)
	}
}

type openCodeRestoreOwner struct {
	parent string
	input  int64
}

func assertOpenCodeRestoredOwners(t *testing.T, ledger domain.AgenticFactLedger, wanted map[string]openCodeRestoreOwner) {
	t.Helper()
	if len(ledger.Sessions) != len(wanted) {
		t.Fatalf("restored ledger owner count = %d, want %d", len(ledger.Sessions), len(wanted))
	}
	for _, owner := range ledger.Sessions {
		want, ok := wanted[owner.ID]
		parent := ""
		if owner.ParentID != nil {
			parent = *owner.ParentID
		}
		if !ok || owner.Source != domain.SourceOpenCode || parent != want.parent || len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != want.input || owner.Invocations[0].ProvisionalSeconds != 0 {
			t.Fatalf("unexpected, duplicate or incomplete owner: %+v", owner)
		}
		delete(wanted, owner.ID)
	}
}
