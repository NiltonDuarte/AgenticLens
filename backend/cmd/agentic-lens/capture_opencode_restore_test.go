package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureOpenCodeIdleRestoreContinuesBoundedDescendantDiscovery(t *testing.T) {
	const siblings = 8
	f, native, store, root := prepareOpenCodeRestoreFixture(t, "", 1, siblings)
	id := capture.StableCaptureID(root.Identity.Source, root.Identity.SessionID)
	assertCaptureRootInput(t, f.ledger(t), 17)

	// These completed descendants are written after discard without invoking a
	// callback. Each budgeted pass may discover or refresh at most one owner.
	addOpenCodeRestoreSession(t, native, f.project, "child", "root", 23)
	addOpenCodeRestoreSession(t, native, f.project, "grandchild", "child", 29)
	addOpenCodeRestoreTask(t, native, "root", "child")
	addOpenCodeRestoreTask(t, native, "child", "grandchild")
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != 1+siblings {
		t.Fatalf("descendants were already registered before restore: %d rows, %v", len(rows), err)
	}
	for _, row := range rows {
		if row.ParentID != "" || domain.CaptureNeedsPeriodicRefresh(row) {
			t.Fatalf("fixture must have only independent idle roots, got %s", row.Identity.SessionID)
		}
	}
	positionBeforeUnrelatedOpenCodeOwners(t, store, root.Identity)

	restoreOpenCodeFixtureAndWait(t, f, store, id)
	wanted := map[string]openCodeRestoreOwner{
		"root":       {input: 17},
		"child":      {parent: "root", input: 23},
		"grandchild": {parent: "child", input: 29},
	}
	for number := 0; number < siblings; number++ {
		wanted[fmt.Sprintf("other-%d", number)] = openCodeRestoreOwner{input: int64(50 + number)}
	}
	assertOpenCodeRestoredOwners(t, f.ledger(t), wanted)
}

func TestCaptureOpenCodeIndependentRestoreKeepsUnavailableUpstreamParent(t *testing.T) {
	const upstream = "unavailable-parent"
	f, native, store, root := prepareOpenCodeRestoreFixture(t, upstream, 8, 0)
	if root.ParentID != upstream || root.Binding != "explicit_agent_response" || root.TicketID != initialPublicationTicket {
		t.Fatalf("fixture lacks an admitted independent native root: %+v", root)
	}
	assertCaptureRootInput(t, f.ledger(t), 17)
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].Identity != root.Identity {
		t.Fatalf("upstream owner must remain unavailable in the local registry: %d rows, %v", len(rows), err)
	}
	if _, err := native.Exec(`UPDATE message SET data=json_set(data,'$.tokens.input',41),time_updated=3500 WHERE id='m-root'`); err != nil {
		t.Fatal(err)
	}
	restoreOpenCodeFixtureAndWait(t, f, store, capture.StableCaptureID(root.Identity.Source, root.Identity.SessionID))
	assertOpenCodeRestoredOwners(t, f.ledger(t), map[string]openCodeRestoreOwner{
		"root": {parent: upstream, input: 41},
	})
}

func positionBeforeUnrelatedOpenCodeOwners(t *testing.T, store *capture.Store, root domain.CaptureIdentity) {
	t.Helper()
	registry, err := sql.Open("sqlite", filepath.Join(store.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	// Ordinary rotation visits IDs after each cursor, then wraps. Put the
	// restored root behind every unrelated root without generating a wake.
	key := capture.IdentityKey(root)
	for _, statement := range []string{
		`UPDATE reconcile_cursor SET identity=? WHERE ticket=?`,
		`UPDATE child_cursor SET key=? WHERE kind='parents-opencode' AND parent=?`,
	} {
		result, err := registry.Exec(statement, key, initialPublicationTicket)
		if err != nil {
			t.Fatal(err)
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			t.Fatalf("fixture requires one existing ticket cursor, got %d, %v", affected, err)
		}
	}
}

func prepareOpenCodeRestoreFixture(t *testing.T, parent string, restoreBudget, siblings int) (captureGenerationFixture, *sql.DB, *capture.Store, domain.CaptureRegistration) {
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
	if out, err := runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", "opencode", "--store", path, "--state-dir", f.state, "--binary", cliBinaryPath); err != nil {
		t.Fatalf("install isolated OpenCode worker selection: %v\n%s", err, out)
	}
	native, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := native.Close(); err != nil {
			t.Error(err)
		}
	})
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
	bind := func(id, parent string, input int) domain.CaptureIdentity {
		t.Helper()
		addOpenCodeRestoreSession(t, native, f.project, id, parent, input)
		identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: id}
		row, err := store.Register(ctx, identity, "", "session.idle")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.BindNativeOpenCode(ctx, identity, capture.BindingTicket{ID: initialPublicationTicket, Title: "Idle restore fixture", Type: "fix"}, row.Request); err != nil {
			t.Fatal(err)
		}
		return identity
	}
	identity := bind("root", parent, 17)
	for number := 0; number < siblings; number++ {
		bind(fmt.Sprintf("other-%d", number), "", 50+number)
	}
	runDiscardFixtureWorker(t, f)
	rows, err := store.List(ctx)
	if err != nil || len(rows) != 1+siblings {
		t.Fatalf("fixture did not register native roots: %d rows, %v", len(rows), err)
	}
	var root domain.CaptureRegistration
	for _, row := range rows {
		if row.Identity == identity {
			root = row
		}
	}
	if root.Identity != identity || root.ParentID != parent || domain.CaptureNeedsPeriodicRefresh(root) {
		t.Fatal("fixture needs an exact verified idle native root")
	}
	if _, err := store.SetCaptureDiscarded(ctx, capture.StableCaptureID(identity.Source, identity.SessionID), true, nil); err != nil {
		t.Fatal(err)
	}
	runDiscardFixtureWorker(t, f)
	settings(restoreBudget)
	return f, native, store, root
}

func restoreOpenCodeFixtureAndWait(t *testing.T, f captureGenerationFixture, store *capture.Store, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var log bytes.Buffer
	completion := captureWorkerCompletionWithLog(t, store.CoordinationDirectory(), &log)
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	previous, err := capture.WorkerStatus(store.CoordinationDirectory())
	if err != nil {
		t.Fatal(err)
	}
	out, err := runCaptureNative(t, f.project, "", "capture", "restore", "--repo", f.project, "--capture-id", id)
	if err != nil {
		t.Fatalf("operator restore: %v\n%s", err, out)
	}
	var queued capture.CaptureControl
	if err := json.Unmarshal(out, &queued); err != nil || queued.CaptureID != id || queued.State != capture.CaptureControlRestoring || !queued.CoveragePending {
		t.Fatalf("restore did not queue an exact coverage check: %+v, %v\n%s", queued, err, out)
	}
	done := make(chan error, 1)
	go func() { done <- completion(ctx) }()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	seen := make(map[int]string)
	var settled error
	for done != nil {
		select {
		case settled = <-done:
			done = nil
		case <-tick.C:
			if progress, err := capture.WorkerStatus(store.CoordinationDirectory()); err == nil && progress.PID != previous.PID && progress.PID > 0 {
				seen[progress.PID] = progress.State
			}
		}
	}
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
		entry, entryErr := adaptersconfig.ResolveEntrypoint(f.project)
		logText := "worker log is still open"
		if settled == nil {
			logText = log.String()
		}
		t.Fatalf("restore did not finish on the original wake: status=%+v request=%+v read=%v completion=%v; owners=%v (%v); proofs=%v (%v); retained passes=%v (%v); worker=%+v (%v); observed worker PIDs=%v; manager entry=%+v (%v)\n%s", status, request, err, settled, owners, listErr, proofs, proofErr, attempts, diagnosticErr, progress, progressErr, seen, entry, entryErr, logText)
	}
	t.Logf("exact restore completed across %d observed worker processes", len(seen))
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
