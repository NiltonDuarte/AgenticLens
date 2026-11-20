package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func readOnlyCaptureSnapshot(t *testing.T, directory string) (string, adapters.PendingSnapshot) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one pending metrics snapshot in %s, got %v (%v)", directory, files, err)
	}
	var snapshot adapters.PendingSnapshot
	if err := json.Unmarshal(readCaptureFile(t, files[0]), &snapshot); err != nil {
		t.Fatal(err)
	}
	return files[0], snapshot
}

func TestCapturePendingSnapshotKeepsNativeFactsThroughBindingAndPublication(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	const configured = ".review/pending captures"
	body := strings.Replace(string(readCaptureFile(t, shared)), captureFixturePendingDir, configured, 1)
	if err := os.WriteFile(shared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = initialPublicationTarget(t, f, "explicit-fixture")
	decoy := initialPublicationTarget(t, f, initialPublicationTicket)
	pending := filepath.Join(f.project, configured)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	root := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	handoff := func(event string) {
		t.Helper()
		input := capture.HookInput{SessionID: root.SessionID, Event: event, WorkingDirectory: f.project, TranscriptPath: filepath.Join(f.store, "sessions", "root.jsonl")}
		if _, err := capture.ReceiveCodexUser(ctx, store, input, f.store, cliBinaryPath, f.state); err != nil {
			t.Fatal(err)
		}
	}
	run := func() {
		t.Helper()
		if output, err := runInitialPublicationWorker(t, f); err != nil {
			t.Fatalf("capture worker failed: %v\n%s\n%+v", err, output, captureDiagnostics(t, f.project, f.state))
		}
	}
	gitIndex := func() string {
		t.Helper()
		command := exec.Command("git", "diff", "--cached", "--name-only", "-z")
		command.Dir = f.project
		output, err := command.Output()
		if err != nil {
			t.Fatal(err)
		}
		return string(output)
	}
	indexBefore := gitIndex()
	path := f.write(t, 27)
	start := []byte(`{"timestamp":"2026-01-01T00:00:01Z","type":"event_msg","payload":{"type":"task_started","turn_id":"one"}}` + "\n")
	completion := []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"one"}}` + "\n")
	metadata, transcript, _ := bytes.Cut(readCaptureFile(t, path), []byte("\n"))
	source := bytes.Join([][]byte{metadata, start, transcript, completion}, []byte("\n"))
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	handoff("SessionStart")
	handoff("SessionEnd")
	run()
	name, unassigned := readOnlyCaptureSnapshot(t, pending)
	id := sha256.Sum256([]byte("codex\x00root"))
	if unassigned.State != "awaiting_ticket" || unassigned.CaptureID != hex.EncodeToString(id[:]) || filepath.Base(name) != unassigned.CaptureID+".json" ||
		unassigned.Source != domain.SourceCodex || unassigned.RootSessionID != root.SessionID || unassigned.TicketID != "" || unassigned.TicketTitle != "" || unassigned.TicketType != "" ||
		len(unassigned.Sessions) != 1 || unassigned.Sessions[0].ID != root.SessionID {
		t.Fatalf("unexpected source-backed snapshot before binding: %+v", unassigned)
	}
	assertCaptureRootInput(t, domain.AgenticFactLedger{Sessions: []domain.SessionFact{unassigned.Sessions[0].ToDomain()}}, 27)
	invocation := unassigned.Sessions[0].ToDomain().Invocations[0]
	if invocation.ActiveDurationSeconds != 2 || invocation.ProvisionalSeconds != 0 {
		t.Fatalf("unbound snapshot did not settle the two-second timing interval: %+v", invocation)
	}
	for _, field := range []string{"ticket_id", "ticket_title", "ticket_type"} {
		if bytes.Contains(readCaptureFile(t, name), []byte(fmt.Sprintf("%q", field))) {
			t.Fatalf("unassigned wire snapshot contains %s", field)
		}
	}
	assertInitialPublicationAbsent(t, f.output, decoy, filepath.Join(f.project, captureFixturePendingDir))

	rows, err := store.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].Identity != root {
		t.Fatalf("expected one exact registration: %+v %v", rows, err)
	}
	const title = "Explicit capture fixture"
	if _, err := store.Bind(ctx, root, initialPublicationTicket, rows[0].Request); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Change(ctx, root, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketTitle, row.TicketType = title, "fix"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	run()
	boundName, bound := readOnlyCaptureSnapshot(t, pending)
	if boundName != name || bound.State != "awaiting_destination" || bound.TicketID != initialPublicationTicket || bound.TicketTitle != title || bound.TicketType != "fix" ||
		!reflect.DeepEqual(unassigned.Sessions, bound.Sessions) {
		t.Fatalf("binding changed capture identity or source facts: before=%+v after=%+v", unassigned, bound)
	}
	assertInitialPublicationAbsent(t, f.output, decoy)
	if got := gitIndex(); got != indexBefore {
		t.Fatalf("capture changed the Git index: before=%q after=%q", indexBefore, got)
	}

	prepareInitialPublication(t, f, f.output)
	run()
	ledger := f.ledger(t)
	assertCaptureRootInput(t, ledger, 27)
	if ledger.TicketTitle == nil || *ledger.TicketTitle != title || len(ledger.Sessions) != 1 ||
		!reflect.DeepEqual(ledger.Sessions[0].Invocations, unassigned.Sessions[0].ToDomain().Invocations) {
		t.Fatalf("publication changed earlier source-backed facts: %+v", ledger)
	}
	assertInitialPublicationAbsent(t, name, f.output+".publication-pending", decoy)
	if got := gitIndex(); got != indexBefore {
		t.Fatalf("publication changed the Git index: before=%q after=%q", indexBefore, got)
	}
}

func TestCaptureWorkerRejectsMissingPendingDirectory(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := filepath.Join(f.project, ".agenticlens-config.yaml")
	body := strings.Replace(string(readCaptureFile(t, path)), "  pending_dir: "+captureFixturePendingDir+"\n", "", 1)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	prepareCaptureOwner(t, f, "root", initialPublicationTicket, 11)
	output, err := runInitialPublicationWorker(t, f)
	if err == nil || !strings.Contains(string(output), "agenticlens.pending_dir") {
		t.Fatalf("capture silently accepted missing project configuration: %v\n%s", err, output)
	}
	assertInitialPublicationAbsent(t, f.output, filepath.Join(f.project, captureFixturePendingDir))
	if got := string(readCaptureFile(t, path)); got != body {
		t.Fatal("capture rewrote shared configuration")
	}
}

func TestCaptureDirectBindRejectsNativeDestinationOptions(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	for _, option := range [][]string{{"--destination-directory", f.project}, {"--destination-only"}} {
		for _, source := range []string{"codex", "opencode"} {
			args := append([]string{"bind", "--source", source, "--state-dir", f.state, "--user-level"}, option...)
			output, err := captureCLI(t, f.project, "", args...)
			if err == nil || !strings.Contains(string(output), "requires the native OpenCode tool or Codex submit-binding") {
				t.Fatalf("direct bind accepted %s on %s: %v\n%s", option[0], source, err, output)
			}
		}
	}
}

func TestCaptureNativeDestinationPromotesPendingInSameDirectory(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.output = initialPublicationTarget(t, f, "explicit-native")
	pending := filepath.Dir(f.output)
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	body := strings.Replace(string(readCaptureFile(t, shared)), captureFixturePendingDir, "records/explicit-native", 1)
	if err := os.WriteFile(shared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	database := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", database)
	policy := fmt.Sprintf("telemetry:\n  default_source: opencode\n  capture:\n    max_lifetime_ms: 1400\n  sources:\n    opencode:\n      database_path: %q\n", database)
	if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
		t.Fatal(err)
	}
	enableCLIFixtureSummary(t, f.project)
	const root = "root"
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO session(id,agent,model,time_created,time_updated) VALUES ('root','build','fixture',1000,3000);
INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES ('m','root',1000,3000,'{"role":"assistant","modelID":"fixture","time":{"created":1000,"completed":3000},"tokens":{"input":17,"output":2}}');`); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, database, f.project)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: database, SessionID: root}
	observed, err := store.Register(ctx, identity, "", "session.idle")
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	ticket := capture.BindingTicket{ID: initialPublicationTicket, Title: "Exact OpenCode fixture", Type: "fix", Destination: outside}
	bound, err := store.BindNativeOpenCode(ctx, identity, ticket, observed.Request)
	if err != nil {
		t.Fatal(err)
	}
	out, err := runInitialPublicationWorker(t, f)
	if err == nil || !strings.Contains(string(out), "native destination rejected") {
		t.Fatalf("native hint escaped admission: %v\n%s", err, out)
	}
	snapshotPath, snapshot := readOnlyCaptureSnapshot(t, pending)
	if snapshot.State != "awaiting_destination" || snapshot.Source != domain.SourceOpenCode || snapshot.TicketID != ticket.ID || len(snapshot.Sessions) != 1 {
		t.Fatalf("blocked publication did not retain the native snapshot: %+v", snapshot)
	}
	assertCaptureRootInput(t, domain.AgenticFactLedger{Sessions: []domain.SessionFact{snapshot.Sessions[0].ToDomain()}}, 17)
	assertInitialPublicationAbsent(t, f.output, filepath.Join(outside, "agentic_metrics.json"), f.output+".publication-pending")

	indicate := func(previous domain.CaptureRegistration, directory string) domain.CaptureRegistration {
		t.Helper()
		submission := capture.BindingTicket{ID: ticket.ID, Destination: directory}
		row, err := store.IndicateNativeOpenCodeDestination(ctx, identity, submission, capture.NativePublicationDestinationRequest(previous))
		if err != nil {
			t.Fatal(err)
		}
		return row
	}
	corrected := indicate(bound, pending)
	run := func() {
		t.Helper()
		if output, err := runInitialPublicationWorker(t, f); err != nil {
			t.Fatalf("native destination did not publish: %v\n%s", err, output)
		}
	}
	if err := writeCaptureFixtureLocal(t, f, []byte(strings.ReplaceAll(policy, "1400", "5000"))); err != nil {
		t.Fatal(err)
	}
	run()
	ledger := f.ledger(t)
	assertCaptureRootInput(t, ledger, 17)
	if ledger.TicketTitle == nil || *ledger.TicketTitle != ticket.Title || len(ledger.Sessions) != 1 ||
		!reflect.DeepEqual(ledger.Sessions[0].Invocations, snapshot.Sessions[0].ToDomain().Invocations) {
		t.Fatalf("canonical publication did not preserve pending facts: %+v", ledger)
	}
	assertInitialPublicationAbsent(t, snapshotPath, f.output+".publication-pending")
	if len(readCaptureFile(t, filepath.Join(pending, "SUMMARY.md"))) == 0 {
		t.Fatal("canonical promotion left an empty summary")
	}

	// The canonical ledger takes precedence over a conflicting native hint.
	indicate(corrected, outside)
	if _, err := db.Exec(`UPDATE message SET data=json_set(data,'$.tokens.input',29),time_updated=3500 WHERE id='m'`); err != nil {
		t.Fatal(err)
	}
	run()
	assertCaptureRootInput(t, f.ledger(t), 29)
	assertInitialPublicationAbsent(t, snapshotPath, filepath.Join(outside, "agentic_metrics.json"))
}

func TestCapturePendingCompletionWakesWorkerDuringDestinationBackoff(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	database := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", database)
	policy := fmt.Sprintf("telemetry:\n  default_source: opencode\n  capture:\n    refresh_ms: 5000\n    debounce_ms: 100\n    idle_ms: 6000\n    max_lifetime_ms: 20000\n  sources:\n    opencode:\n      database_path: %q\n", database)
	if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const boundRoot, unboundRoot = "blocked", "unassigned"
	started := time.Now().Add(-5 * time.Second).UnixMilli()
	for _, id := range []string{boundRoot, unboundRoot} {
		if _, err := db.Exec("INSERT INTO session(id,agent,model,time_created,time_updated) VALUES (?,'build','fixture',?,?)", id, started, started); err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":%d},"tokens":{"input":17,"output":2}}`, started)
		if _, err := db.Exec("INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES (?,?,?,?,?)", "m-"+id, id, started, started, body); err != nil {
			t.Fatal(err)
		}
	}
	activeTool := fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":"running","time":{"start":%d}}}`, started)
	if _, err := db.Exec("INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES ('tool','unassigned','m-unassigned',?,?,?)", started, started, activeTool); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, database, f.project)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	bound := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: database, SessionID: boundRoot}
	unbound := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: database, SessionID: unboundRoot}
	observed, err := store.Register(ctx, bound, "", "session.created")
	if err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(f.project, "records", "missing")
	ticket := capture.BindingTicket{ID: initialPublicationTicket, Title: "Blocked fixture", Type: "fix", Destination: missing}
	if _, err := store.BindNativeOpenCode(ctx, bound, ticket, observed.Request); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Register(ctx, unbound, "", "session.created"); err != nil {
		t.Fatal(err)
	}
	if err := store.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(t.TempDir(), "pending-worker.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	worker := exec.CommandContext(workerCtx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	worker.Dir, worker.Stdout, worker.Stderr = f.project, log, log
	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cancel()
		_ = worker.Wait()
	}()

	// Observe a real blocked attempt and its retry deadline before supplying
	// the second root's independently completed native source interval.
	var blocked capture.WorkerProgress
	initialDeadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(initialDeadline) {
		status, err := capture.WorkerStatus(store.CoordinationDirectory())
		if err == nil && status.State == "failure" && status.NextAttempt > 0 {
			blocked = status
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if blocked.NextAttempt == 0 {
		t.Fatalf("worker did not back off on its rejected native hint: %s", readCaptureFile(t, log.Name()))
	}
	id := sha256.Sum256([]byte("opencode\x00" + unboundRoot))
	path := filepath.Join(f.project, captureFixturePendingDir, hex.EncodeToString(id[:])+".json")
	read := func() adapters.PendingSnapshot {
		t.Helper()
		var snapshot adapters.PendingSnapshot
		if err := json.Unmarshal(readCaptureFile(t, path), &snapshot); err != nil {
			t.Fatal(err)
		}
		if snapshot.State != "awaiting_ticket" || snapshot.RootSessionID != unboundRoot || len(snapshot.Sessions) != 1 || len(snapshot.Sessions[0].Invocations) != 1 {
			t.Fatalf("unexpected unbound snapshot: %+v", snapshot)
		}
		return snapshot
	}
	if kickoff := read().Sessions[0].Invocations[0].ToDomain(); kickoff.ProvisionalSeconds <= 0 {
		t.Fatalf("unbound interval must be provisional before native completion: %+v", kickoff)
	}
	completed := started + 4000
	if _, err := db.Exec(`UPDATE message SET data=json_set(data,'$.time.completed',?),time_updated=? WHERE id='m-unassigned'`, completed, completed); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE part SET data=json_set(data,'$.state.status','completed','$.state.time.end',?),time_updated=? WHERE id='tool'`, completed, completed); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Register(ctx, unbound, "", "session.idle"); err != nil {
		t.Fatal(err)
	}
	if err := store.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	for cutoff := time.UnixMilli(blocked.NextAttempt); time.Now().Before(cutoff); time.Sleep(25 * time.Millisecond) {
		done := read().Sessions[0].Invocations[0].ToDomain()
		if done.ProvisionalSeconds == 0 && done.ActiveDurationSeconds == 4 {
			assertInitialPublicationAbsent(t, filepath.Join(missing, "agentic_metrics.json"), f.output)
			return
		}
	}
	t.Fatalf("native completion was postponed until another ticket's backoff ended: %+v\n%s", read(), readCaptureFile(t, log.Name()))
}
