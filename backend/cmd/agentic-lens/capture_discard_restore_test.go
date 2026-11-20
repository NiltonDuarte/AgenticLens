package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureDiscardRetriesCleanupAndFencesObsoletePendingWrites(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	source := f.write(t, 17)
	if _, err = capture.ReceiveCodexUser(ctx, store, capture.HookInput{
		SessionID: "root", Event: "SessionEnd", WorkingDirectory: f.project, TranscriptPath: source,
	}, f.store, cliBinaryPath, f.state); err != nil {
		t.Fatal(err)
	}
	runDiscardFixtureWorker(t, f)
	pending := filepath.Join(f.project, captureFixturePendingDir)
	path, initial := readOnlyCaptureSnapshot(t, pending)
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	if initial.CaptureID != id || initial.State != "awaiting_ticket" {
		t.Fatalf("unexpected unbound source snapshot: %+v", initial)
	}
	git := func(args ...string) []byte {
		t.Helper()
		command := exec.Command("git", args...)
		command.Dir = f.project
		out, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return out
	}
	git("add", "--", path)
	indexBefore := git("ls-files", "--stage", "-z")
	if !bytes.Contains(indexBefore, []byte(filepath.Base(path))) {
		t.Fatal("the exact pending snapshot was not staged in the fixture")
	}
	oldLease, err := capture.AcquireWorker(store.CoordinationDirectory())
	if err != nil || oldLease == nil {
		t.Fatalf("legacy worker lease: %v", err)
	}
	out, discardErr := runCaptureNative(t, f.project, "", "capture", "discard", "--repo", f.project, "--capture-id", id)
	if err := oldLease.Close(); err != nil {
		t.Fatal(err)
	}
	if discardErr == nil || !bytes.Contains(out, []byte("legacy exclusive lease")) {
		t.Fatalf("discard passed an admitted incompatible worker: %v\n%s", discardErr, out)
	}
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlActive {
		t.Fatalf("rejected discard changed lifecycle state: %+v", status)
	}
	unknown := capture.StableCaptureID(domain.SourceCodex, "unregistered")
	if status := readDiscardFixtureStatus(t, f, unknown); status.State != capture.CaptureControlUnknown {
		t.Fatalf("unregistered ID did not report unknown state: %+v", status)
	}
	if out, err := runCaptureNative(t, f.project, "", "capture", "discard", "--repo", f.project, "--capture-id", unknown); err == nil {
		t.Fatalf("discard accepted an unverified capture ID: %s", out)
	}
	stale := reconcileDiscardFixture(t, store, f, "")

	// Leave the genuine pending snapshot in place after committing the durable
	// operation, as if the actor exited before the filesystem cleanup.
	discard, err := store.SetCaptureDiscarded(ctx, id, true, nil)
	if err != nil || discard.State != capture.CaptureControlDiscarded || !discard.CleanupPending {
		t.Fatalf("durable discard: %+v, %v", discard, err)
	}
	beforeCleanup := readCaptureFile(t, path)
	if _, err := adapters.SyncPendingSnapshots(ctx, store, pending, f.project, stale, true, nil, nil); !errors.Is(err, capture.ErrCaptureLifecycleChanged) {
		t.Fatalf("obsolete pending writer reached publication: %v", err)
	}
	if !bytes.Equal(readCaptureFile(t, path), beforeCleanup) {
		t.Fatal("obsolete writer changed the actual pending snapshot")
	}
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlDiscarded || !status.CleanupPending || status.Revision != discard.Revision {
		t.Fatalf("exact status cannot discover unfinished cleanup: %+v", status)
	}
	runDiscardFixtureWorker(t, f)
	assertInitialPublicationAbsent(t, path, f.output)
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlDiscarded || status.CleanupPending || status.Revision != discard.Revision {
		t.Fatalf("worker did not acknowledge exact discard cleanup: %+v", status)
	}
	if again, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil || again.Revision != discard.Revision || again.CleanupPending {
		t.Fatalf("repeat discard restarted completed work: %+v, %v", again, err)
	}
	appendDiscardFixtureTurn(t, source, "root", "during", 19, "04")
	runDiscardFixtureWorker(t, f)
	assertInitialPublicationAbsent(t, path, f.output)
	if !bytes.Contains(readCaptureFile(t, source), []byte(`"turn_id":"during"`)) {
		t.Fatal("worker discarded native source data")
	}

	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	out, err = runCaptureNative(t, f.project, "", "capture", "restore", "--repo", f.project, "--capture-id", id)
	if err != nil {
		t.Fatalf("operator restore: %v\n%s", err, out)
	}
	var restored capture.CaptureControl
	if err := json.Unmarshal(out, &restored); err != nil || restored.State != capture.CaptureControlRestoring || restored.Revision <= discard.Revision {
		t.Fatalf("restore response did not expose pending coverage: %+v, %v\n%s", restored, err, out)
	}
	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		status, err := store.CaptureControl(deadline, id)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == capture.CaptureControlActive && !status.CoveragePending && status.Revision == restored.Revision {
			break
		}
		select {
		case <-deadline.Done():
			t.Fatalf("restoration did not settle: %+v", status)
		case <-time.After(20 * time.Millisecond):
		}
	}
	name, snapshot := readOnlyCaptureSnapshot(t, pending)
	if name != path || snapshot.CaptureID != id || snapshot.TicketID != "" || len(snapshot.Sessions) != 1 {
		t.Fatalf("restore changed pending identity or inferred binding: %+v", snapshot)
	}
	assertCaptureRootInput(t, domain.AgenticFactLedger{Sessions: []domain.SessionFact{snapshot.Sessions[0].ToDomain()}}, 36)
	sealed := readCaptureFile(t, path)
	if err := adapters.RemoveDiscardedPendingSnapshot(ctx, store, pending, f.project, discard, nil); !errors.Is(err, capture.ErrCaptureLifecycleChanged) {
		t.Fatalf("obsolete discard cleanup could delete restored facts: %v", err)
	}
	if !bytes.Equal(readCaptureFile(t, path), sealed) || !bytes.Equal(indexBefore, git("ls-files", "--stage", "-z")) {
		t.Fatal("restoration or obsolete cleanup changed the snapshot or exact staged blob")
	}
}

func TestCaptureDiscardRetainsCanonicalHistoryUntilBoundRestore(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = initialPublicationTarget(t, f, "controlled-ticket")
	prepareInitialPublication(t, f, f.output)
	source := prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	priorChild := writeDiscardFixtureChild(t, f, "prior-child", 5)
	if _, err := capture.ReceiveCodexUser(ctx, store, capture.HookInput{
		SessionID: "prior-child", Event: "SessionEnd", WorkingDirectory: f.project, TranscriptPath: priorChild,
	}, f.store, cliBinaryPath, f.state); err != nil {
		t.Fatal(err)
	}
	if err := store.QueueChild(ctx, f.store, priorChild, "root"); err != nil {
		t.Fatal(err)
	}
	const otherTicket = "TEST-200"
	other := writeDiscardFixtureChild(t, f, "independent", 11)
	child := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "independent"}
	registered, err := store.Register(ctx, child, other, "SessionEnd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Bind(ctx, child, otherTicket, registered.Request); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Change(ctx, child, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketTitle, row.TicketType = "Independent child", "fix"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	initialOther := reconcileDiscardFixture(t, store, f, otherTicket)
	if len(initialOther.Sessions) != 1 || initialOther.Sessions[0].ParentID == nil || *initialOther.Sessions[0].ParentID != "root" {
		t.Fatalf("fixture independent owner lacked a verified native parent: %+v", initialOther.Sessions)
	}
	fOther := f
	fOther.output = seedCaptureTicketHome(t, f, otherTicket)
	runDiscardFixtureWorker(t, f)
	initial := f.ledger(t)
	assertCaptureRootInput(t, initial, 17)
	canonical := readCaptureFile(t, f.output)
	if len(initial.Sessions) != 2 {
		t.Fatalf("unexpected initial canonical owners: %+v", initial.Sessions)
	}
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	discard, err := store.SetCaptureDiscarded(ctx, id, true, nil)
	if err != nil || discard.TicketID != initialPublicationTicket {
		t.Fatalf("discard did not retain bound scope: %+v, %v", discard, err)
	}
	// A later source handoff can fail verification without retracting the
	// previously proved worktree and parent membership of either owner.
	unavailable := t.TempDir()
	retiredRoot := filepath.Join(unavailable, "root.jsonl")
	retiredChild := filepath.Join(unavailable, "prior-child.jsonl")
	if err := os.Rename(source, retiredRoot); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(priorChild, retiredChild); err != nil {
		t.Fatal(err)
	}
	rows, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	failedProofs := map[string]string{"root": "", "prior-child": "root"}
	for _, row := range rows {
		parent, selected := failedProofs[row.Identity.SessionID]
		if !selected {
			continue
		}
		if row.ProjectProof != domain.CaptureProjectVerified || row.ParentID != parent {
			t.Fatalf("fixture lacks accepted source scope for %s", row.Identity.SessionID)
		}
		latest, err := store.ProveRegistration(ctx, row)
		if err != nil || latest.ProjectProof != domain.CaptureProjectUnverified || latest.ProjectDirectory != row.ProjectDirectory || latest.ParentID != parent {
			t.Fatalf("failed latest proof did not retain accepted scope for %s: %v", row.Identity.SessionID, err)
		}
		delete(failedProofs, row.Identity.SessionID)
	}
	if len(failedProofs) != 0 {
		t.Fatalf("missing source fixture owners: %v", failedProofs)
	}
	// Reappearance alone must not let an older, currently unverified root
	// escape discard. The child remains unavailable through initial restore.
	if err := os.Rename(retiredRoot, source); err != nil {
		t.Fatal(err)
	}
	appendDiscardFixtureTurn(t, source, "root", "during", 19, "04")
	appendDiscardFixtureTurn(t, retiredChild, "prior-child", "during", 3, "04")
	appendDiscardFixtureTurn(t, other, "independent", "during", 13, "04")
	late := writeDiscardFixtureChild(t, f, "late-child", 29)
	if _, err := capture.ReceiveCodexUser(ctx, store, capture.HookInput{
		SessionID: "late-child", Event: "SessionEnd", WorkingDirectory: f.project, TranscriptPath: late,
	}, f.store, cliBinaryPath, f.state); err != nil {
		t.Fatal(err)
	}
	if err := store.QueueChild(ctx, f.store, late, "root"); err != nil {
		t.Fatal(err)
	}
	runDiscardFixtureWorker(t, f)
	assertInitialPublicationAbsent(t, filepath.Join(f.project, captureFixturePendingDir, capture.StableCaptureID(domain.SourceCodex, "late-child")+".json"))
	if !bytes.Equal(readCaptureFile(t, f.output), canonical) {
		t.Fatal("discarded native activity or delayed child rewrote the accepted ledger")
	}
	if got := fOther.ledger(t).Sessions; len(got) != 1 || got[0].ID != child.SessionID || len(got[0].Invocations) != 1 || got[0].Invocations[0].Tokens.Input != 24 {
		t.Fatalf("independent child could not refresh its own ticket during discard: %+v", got)
	}

	// A separate root on the same ticket can advance without replacing the
	// discarded root's already accepted facts.
	prepareCaptureOwner(t, f, "sibling", initialPublicationTicket, 7)
	runDiscardFixtureWorker(t, f)
	retained := f.ledger(t)
	assertCaptureRootInput(t, retained, 17)
	if len(retained.Sessions) != 3 {
		t.Fatalf("same-ticket active root could not publish or admitted a discarded child: %+v", retained.Sessions)
	}
	accepted := make(map[string]domain.SessionFact, len(initial.Sessions))
	for _, owner := range initial.Sessions {
		accepted[owner.ID] = owner
	}
	for _, owner := range retained.Sessions {
		if old, ok := accepted[owner.ID]; ok && !reflect.DeepEqual(owner, old) {
			t.Fatal("same-ticket refresh replaced previously accepted discarded-owner evidence")
		}
	}
	restore, err := store.SetCaptureDiscarded(ctx, id, false, nil)
	if err != nil || restore.State != capture.CaptureControlRestoring || restore.Revision <= discard.Revision {
		t.Fatalf("bound restore: %+v, %v", restore, err)
	}
	incomplete := reconcileDiscardFixture(t, store, f, initialPublicationTicket)
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlRestoring || !status.CoveragePending || status.TicketID != initialPublicationTicket || len(incomplete.RestoresReady) != 0 {
		t.Fatalf("restore treated an unavailable native source as complete: status=%+v ready=%+v", status, incomplete.RestoresReady)
	}
	if err := os.Rename(retiredChild, priorChild); err != nil {
		t.Fatal(err)
	}
	appendDiscardFixtureTurn(t, source, "root", "after", 23, "05")
	runDiscardFixtureWorker(t, f)
	final := f.ledger(t)
	assertCaptureRootInput(t, final, 59)
	if final.TicketID == nil || *final.TicketID != initialPublicationTicket || len(final.Sessions) != 4 {
		t.Fatalf("restoration changed the canonical ticket or owner set: %+v", final)
	}
	inputs := map[string]int64{"root": 59, "prior-child": 8, "sibling": 7, "late-child": 29}
	for _, owner := range final.Sessions {
		want, expected := inputs[owner.ID]
		if !expected || len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != want {
			t.Fatalf("unexpected owner or source facts in restored ledger: %+v", owner)
		}
		if (owner.ID == "late-child" || owner.ID == "prior-child") && (owner.ParentID == nil || *owner.ParentID != "root") {
			t.Fatalf("restored delayed child lacks exact native parent: %+v", owner)
		}
		delete(inputs, owner.ID)
	}
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlActive || status.CoveragePending || status.Revision != restore.Revision || status.TicketID != initialPublicationTicket {
		t.Fatalf("bound restore did not complete the original capture: %+v", status)
	}
	if again, err := store.SetCaptureDiscarded(ctx, id, false, nil); err != nil || again.Revision != restore.Revision || again.CoveragePending {
		t.Fatalf("repeat restore restarted completed work: %+v, %v", again, err)
	}
	runDiscardFixtureWorker(t, f)
	if repeated := f.ledger(t); !reflect.DeepEqual(repeated.Sessions, final.Sessions) {
		t.Fatalf("another worker pass changed or duplicated accepted source facts: %+v", repeated.Sessions)
	}
}

func TestCaptureDiscardRecoversInterruptedFirstPublicationWithoutRepublishing(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = initialPublicationTarget(t, f, "discarded-interrupted-ticket")
	summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
	original := []byte("# Authored notes before the interrupted publication\n")
	if err := os.WriteFile(summary, original, 0600); err != nil {
		t.Fatal(err)
	}
	source := prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	runDiscardFixtureWorker(t, f)
	snapshot, _ := readOnlyCaptureSnapshot(t, filepath.Join(f.project, captureFixturePendingDir))
	prepareInitialPublication(t, f, f.output)
	stageInterruptedDiscardPublication(t, f, summary, true)
	marker := f.output + ".publication-pending"
	journal := captureFixturePrivateJournal(t, f)
	if err := adapters.CheckPublication(f.output); err == nil {
		t.Fatal("interrupted first publication was readable as complete")
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	appendDiscardFixtureTurn(t, source, "root", "during", 19, "04")
	runDiscardFixtureWorker(t, f)
	assertInitialPublicationAbsent(t, f.output, marker, journal, snapshot)
	if got := readCaptureFile(t, summary); !bytes.Equal(got, original) {
		t.Fatalf("discard did not recover the authored summary: %q", got)
	}
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlDiscarded || status.CleanupPending {
		t.Fatalf("interrupted set and pending cleanup remain unresolved: %+v", status)
	}
}

func TestCaptureDiscardRecoversOrdinaryRefreshWithoutFirstPublicationReceipt(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = seedCaptureTicketHome(t, f, initialPublicationTicket)
	source := prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	runDiscardFixtureWorker(t, f)
	assertCaptureRootInput(t, f.ledger(t), 17)
	summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
	prior, priorSummary := readCaptureFile(t, f.output), readCaptureFile(t, summary)
	stageInterruptedDiscardPublication(t, f, summary, false)
	marker, journal := f.output+".publication-pending", captureFixturePrivateJournal(t, f)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	appendDiscardFixtureTurn(t, source, "root", "during", 19, "04")
	runDiscardFixtureWorker(t, f)
	assertInitialPublicationAbsent(t, marker, journal)
	if !bytes.Equal(prior, readCaptureFile(t, f.output)) || !bytes.Equal(priorSummary, readCaptureFile(t, summary)) {
		t.Fatal("discard did not recover the exact prior canonical files")
	}
	if status := readDiscardFixtureStatus(t, f, id); status.State != capture.CaptureControlDiscarded || status.CleanupPending {
		t.Fatalf("ordinary refresh recovery left cleanup pending: %+v", status)
	}
}

func TestCaptureRestoreCoverageIsIndependentOfOtherTicketRoots(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	prepareCaptureOwner(t, f, "sibling", initialPublicationTicket, 7)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	reconcileDiscardFixture(t, store, f, initialPublicationTicket)
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	if _, err := store.SetCaptureDiscarded(ctx, id, true, nil); err != nil {
		t.Fatal(err)
	}
	if err := store.QueueChildStart(ctx, f.store, "sibling", "never-flushed-child"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCaptureDiscarded(ctx, id, false, nil); err != nil {
		t.Fatal(err)
	}
	result := reconcileDiscardFixture(t, store, f, initialPublicationTicket)
	if !result.DiscoveryPending || result.PendingChildren != 1 || len(result.RefreshedSessions) != 2 {
		t.Fatalf("unresolved child must remain globally visible: discovery=%t pending=%d refreshed=%d", result.DiscoveryPending, result.PendingChildren, len(result.RefreshedSessions))
	}
	if len(result.RestoresReady) != 1 || result.RestoresReady[0].CaptureID != id {
		t.Fatalf("unresolved child of a different root blocked our restored capture: %+v", result.RestoresReady)
	}
}

func stageInterruptedDiscardPublication(t *testing.T, f captureGenerationFixture, summary string, first bool) {
	t.Helper()
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	lease, err := capture.AcquireInitialPublication(ctx, f.project, adapters.PublicationTicketKey(initialPublicationTicket, cfg.Ticket), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lease.Close(); err != nil {
			t.Error(err)
		}
	}()
	selection, err := lease.Read()
	if err != nil || (selection != nil) != first {
		t.Fatalf("unexpected first-publication receipt for the interruption fixture: %+v, %v", selection, err)
	}
	owner, err := capture.AcquirePublication(ctx, f.output, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := owner.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := adapters.BeginInitialPublication(f.output, []string{f.output, summary}, selection); err != nil {
		t.Fatal(err)
	}
	if err := capture.WritePublicationFile(f.output, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := capture.WritePublicationFile(summary, []byte("Partial summary bytes\n"), 0600); err != nil {
		t.Fatal(err)
	}
}

func runDiscardFixtureWorker(t *testing.T, f captureGenerationFixture) {
	t.Helper()
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("automatic worker: %v\n%s", err, output)
	}
}

func readDiscardFixtureStatus(t *testing.T, f captureGenerationFixture, id string) capture.CaptureControl {
	t.Helper()
	out, err := runCaptureNative(t, f.project, "", "capture", "status", "--repo", f.project, "--capture-id", id)
	if err != nil {
		t.Fatalf("exact capture status: %v\n%s", err, out)
	}
	var status capture.CaptureControl
	if err := json.Unmarshal(out, &status); err != nil || status.CaptureID != id || bytes.Contains(out, []byte(f.store)) {
		t.Fatalf("invalid or nonportable exact status: %+v, %v\n%s", status, err, out)
	}
	return status
}

func reconcileDiscardFixture(t *testing.T, store *capture.Store, f captureGenerationFixture, ticket string) capture.ReconcileResult {
	t.Helper()
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 6; attempt++ {
		result, err := store.Reconcile(context.Background(), ticket, "fixture", 32, 64<<20, cfg.Generation.Metrics)
		if !errors.Is(err, capture.ErrCaptureLifecycleChanged) {
			if err != nil {
				t.Fatal(err)
			}
			return result
		}
	}
	t.Fatal("capture lifecycle revision did not settle after verified descendant discovery")
	return capture.ReconcileResult{}
}

func writeDiscardFixtureChild(t *testing.T, f captureGenerationFixture, id string, input int) string {
	t.Helper()
	path := filepath.Join(f.store, "sessions", id+".jsonl")
	metadata := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":%q,"parent_thread_id":"root","cwd":%q}}`+"\n", id, f.project)
	if err := os.WriteFile(path, []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}
	appendDiscardFixtureTurn(t, path, id, "one", input, "02")
	return path
}

func appendDiscardFixtureTurn(t *testing.T, path, id, turn string, input int, seconds string) {
	t.Helper()
	source := readCaptureFile(t, path)
	timestamp := "2026-01-01T00:00:" + seconds + "Z"
	turnData := fmt.Sprintf(`{"timestamp":%q,"type":"turn_context","payload":{"model":"source-model","turn_id":%q}}`+"\n"+
		`{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":%q,"usage":{"input_tokens":%d,"output_tokens":2}}}`+"\n",
		timestamp, turn, timestamp, id, turn, input)
	if err := os.WriteFile(path, append(source, turnData...), 0600); err != nil {
		t.Fatal(err)
	}
}
