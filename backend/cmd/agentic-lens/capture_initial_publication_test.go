package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

const initialPublicationTicket = "TEST-100"

func initialPublicationTarget(t *testing.T, f captureGenerationFixture, name string) string {
	t.Helper()
	dir := filepath.Join(f.project, "records", name)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "agentic_metrics.json")
}

func prepareInitialPublication(t *testing.T, f captureGenerationFixture, path string) {
	t.Helper()
	output, err := runCaptureNative(t, f.project, "", "capture", "prepare-publication", "--repo", f.project, "--ticket-id", initialPublicationTicket, "--json-output", path)
	if err != nil {
		t.Fatalf("prepare first publication: %v\n%s", err, output)
	}
}

func runInitialPublicationWorker(t *testing.T, f captureGenerationFixture) ([]byte, error) {
	t.Helper()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = store.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	output, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	request, requestErr := store.PublicationRequest(context.Background())
	if requestErr != nil {
		t.Fatal(requestErr)
	}
	if err == nil && request.Generation != request.Acknowledged {
		t.Fatalf("successful publication left its generation pending: %+v\n%s", request, output)
	}
	if err != nil && request.Generation == request.Acknowledged {
		t.Fatalf("failed publication acknowledged its generation: %+v\n%s", request, output)
	}
	return output, err
}

func assertInitialPublicationAbsent(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s must be absent: %v", path, err)
		}
	}
}

func TestCapturePreparedFirstPublicationMovesAndRemainsPortable(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = initialPublicationTarget(t, f, "chosen-home")
	decoy := initialPublicationTarget(t, f, initialPublicationTicket)
	summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
	introduction := []byte("# Existing operator notes\n")
	if err := os.WriteFile(summary, introduction, 0600); err != nil {
		t.Fatal(err)
	}

	prepareInitialPublication(t, f, f.output)
	prepareInitialPublication(t, f, f.output)
	assertInitialPublicationAbsent(t, f.output, decoy, f.output+".publication-pending", filepath.Join(f.state, "registry.db"))
	if !bytes.Equal(readCaptureFile(t, summary), introduction) {
		t.Fatal("preparing the destination changed the operator's summary")
	}

	prepareCaptureOwner(t, f, "root", initialPublicationTicket, 19)
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("first automatic publication: %v\n%s", err, output)
	}
	first := f.ledger(t)
	assertCaptureRootInput(t, first, 19)
	if first.TicketID == nil || *first.TicketID != initialPublicationTicket ||
		first.TicketTitle == nil || *first.TicketTitle != "Explicit "+initialPublicationTicket ||
		first.TicketType == nil || *first.TicketType != domain.TicketTypeFeat ||
		first.CompletedAt == nil || *first.CompletedAt != first.UpdatedAt ||
		len(first.Sessions) != 1 || first.Sessions[0].Source != domain.SourceCodex {
		t.Fatalf("first committed ledger lacks explicit bound metadata or exact source: %+v", first)
	}
	if body := readCaptureFile(t, summary); !bytes.HasPrefix(body, introduction) || bytes.Equal(body, introduction) {
		t.Fatalf("summary did not retain its authored introduction and append metrics: %s", body)
	}
	assertCapturePublishedOwners(t, captureDiagnostics(t, f.project, f.state), f.output, summary, first.Sessions)
	assertInitialPublicationAbsent(t, decoy, f.output+".publication-pending")

	old := f.output
	f.output = filepath.Join(f.project, "records", "relocated-home", "agentic_metrics.json")
	if err := os.Rename(filepath.Dir(old), filepath.Dir(f.output)); err != nil {
		t.Fatal(err)
	}
	f.writeAt(t, 29, "2026-01-01T00:00:03Z")
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("automatic refresh after move: %v\n%s", err, output)
	}
	later := f.ledger(t)
	assertCaptureRootInput(t, later, 29)
	if len(later.Sessions) != 1 || !reflect.DeepEqual(first.TicketTitle, later.TicketTitle) {
		t.Fatalf("move or refresh changed owner multiplicity or ticket title: %+v", later)
	}
	assertInitialPublicationAbsent(t, filepath.Dir(old), f.output+".publication-pending")

	copy := newCaptureGenerationFixture(t)
	copy.output = initialPublicationTarget(t, copy, "different-machine-home")
	writeCaptureSuccessfulPublicationPolicy(t, copy.project, copy.stateBase)
	if err := os.WriteFile(copy.output, readCaptureFile(t, f.output), 0600); err != nil {
		t.Fatal(err)
	}
	prepareCaptureOwner(t, copy, "local-owner", initialPublicationTicket, 13)
	if output, err := runInitialPublicationWorker(t, copy); err != nil {
		t.Fatalf("automatic refresh with portable ledger and independent state: %v\n%s", err, output)
	}
	fromCopy := copy.ledger(t)
	owners := make(map[string]domain.SessionFact)
	for _, session := range fromCopy.Sessions {
		owners[session.ID] = session
	}
	if len(fromCopy.Sessions) != 2 || !reflect.DeepEqual(later.Sessions[0], owners["root"]) {
		t.Fatalf("portable capture lost or changed unavailable owner: %+v", fromCopy.Sessions)
	}
	added := owners["local-owner"]
	if len(added.Invocations) != 1 || added.Invocations[0].Tokens.Input != 13 {
		t.Fatalf("portable refresh did not add the new machine's source facts: %+v", added)
	}
}

func TestCaptureConcurrentFirstSelectionsAdmitOneDestination(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	first := initialPublicationTarget(t, f, "choice-one")
	second := initialPublicationTarget(t, f, "choice-two")

	type result struct {
		path   string
		output []byte
		err    error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for i, path := range []string{first, second} {
		ticket := initialPublicationTicket
		if i == 1 {
			ticket = strings.ToLower(ticket)
		}
		go func() {
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, cliBinaryPath, "capture", "prepare-publication", "--repo", f.project, "--ticket-id", ticket, "--json-output", path)
			cmd.Dir = f.project
			output, err := cmd.CombinedOutput()
			results <- result{path: path, output: output, err: err}
		}()
	}
	close(start)
	a, b := <-results, <-results
	if (a.err == nil) == (b.err == nil) {
		t.Fatalf("want exactly one accepted destination: %v\n%s\nand %v\n%s", a.err, a.output, b.err, b.output)
	}
	if a.err != nil {
		a, b = b, a
	}
	assertInitialPublicationAbsent(t, a.path, b.path)

	f.output = a.path
	prepareCaptureOwner(t, f, "root", initialPublicationTicket, 17)
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("automatic publication at the sole selected destination: %v\n%s", err, output)
	}
	assertCaptureRootInput(t, f.ledger(t), 17)
	assertInitialPublicationAbsent(t, b.path, f.output+".publication-pending")
}

// Exercise the actual manager, protected engine, binding inbox and native
// worker with a fixture transcript. The separate release gate uses real Codex.
func TestCaptureProtectedCodexPublishesFirstPreparedDestination(t *testing.T) {
	f := newCodexInboxFixture(t)
	codexInboxOperation(t, f.project, "install-user")
	enableCLIFixtureSummary(t, f.project)
	f.output = initialPublicationTarget(t, f, "protected-first-home")
	preparer := exec.Command(fixtureCaptureManager(), "--al-offline", "--al-project", f.project,
		"capture", "prepare-publication", "--repo", f.project, "--ticket-id", initialPublicationTicket, "--json-output", f.output)
	preparer.Dir = f.project
	if output, err := preparer.CombinedOutput(); err != nil {
		t.Fatalf("installed manager did not route first publication preparation: %v\n%s", err, output)
	}
	assertInitialPublicationAbsent(t, f.output, f.output+".publication-pending")
	rollout := f.write(t, 31)
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	completion := captureWorkerCompletion(t, registry.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })

	receive := func(event string) capture.HookOutput {
		t.Helper()
		payload, err := json.Marshal(capture.HookInput{Event: event, WorkingDirectory: f.project, SessionID: "root", TranscriptPath: rollout})
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(fixtureCaptureManager(), "codex-hook")
		cmd.Dir = f.project
		cmd.Stdin = bytes.NewReader(payload)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("protected %s callback: %v\n%s", event, err, output)
		}
		var reply capture.HookOutput
		if err := json.Unmarshal(output, &reply); err != nil {
			t.Fatalf("invalid %s callback response: %v\n%s", event, err, output)
		}
		return reply
	}
	credential := inboxCredentialFromHook(t, receive("SessionStart"))
	output, err := captureCLI(t, f.project, "", "submit-binding", "--repo", f.project, "--store", f.store, "--session-id", "root",
		"--ticket-id", initialPublicationTicket, "--ticket-title", "Protected first publication", "--ticket-type", "fix", "--request", credential)
	if err != nil {
		t.Fatalf("submit authenticated binding: %v\n%s", err, output)
	}
	receive("SessionEnd")
	waitCorrectionLedger(t, f.output, "root", 31)

	deadline, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	if err := completion(deadline); err != nil {
		t.Fatalf("protected worker did not finalize and exit: %v", err)
	}
	ledger := f.ledger(t)
	if ledger.TicketTitle == nil || *ledger.TicketTitle != "Protected first publication" ||
		ledger.TicketType == nil || *ledger.TicketType != domain.TicketTypeFix {
		t.Fatalf("protected first publication lost native binding metadata: %+v", ledger)
	}
	if err := adapters.CheckPublication(f.output); err != nil {
		t.Fatal(err)
	}
	assertCapturePublishedOwners(t, captureDiagnostics(t, f.project, f.state), f.output, filepath.Join(filepath.Dir(f.output), "SUMMARY.md"), ledger.Sessions)
}

func TestCaptureOpenCodeCompletionWakesWorkerBeforePeriodicRefresh(t *testing.T) {
	const refresh, debounce, wakeBudget = 15 * time.Second, 500 * time.Millisecond, 6 * time.Second
	f := newCodexInboxFixture(t)
	f.output = initialPublicationTarget(t, f, "completion-wake")
	database := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", database)
	policy := fmt.Sprintf("telemetry:\n  default_source: opencode\n  capture:\n    refresh_ms: %d\n    debounce_ms: %d\n    idle_ms: 30000\n    max_lifetime_ms: 60000\n  sources:\n    opencode:\n      database_path: %q\n", refresh.Milliseconds(), debounce.Milliseconds(), database)
	if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
		t.Fatal(err)
	}
	prepareInitialPublication(t, f, f.output)
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	created := time.Now().Add(-8 * time.Second).UnixMilli()
	if _, err := db.Exec("INSERT INTO session(id,agent,model,time_created,time_updated) VALUES ('root','build','fixture',?,?)", created, created); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, database, f.project)
	activeMessage := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":%d},"tokens":{"input":17,"output":2}}`, created)
	activePart := fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":"running","time":{"start":%d}}}`, created)
	if _, err := db.Exec("INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES ('m','root',?,?,?)", created, created, activeMessage); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO part(id,session_id,message_id,time_created,time_updated,data) VALUES ('p','root','m',?,?,?)", created, created, activePart); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	})
	completion := captureWorkerCompletion(t, registry.CoordinationDirectory())
	t.Cleanup(func() {
		if err := writeCaptureFixtureLocal(t, f, []byte(strings.Replace(policy, "  capture:\n", "  capture:\n    enabled: false\n", 1))); err != nil {
			t.Error(err)
		}
		if err := registry.RequestPublication(ctx); err != nil {
			t.Error(err)
		}
		cleanupCtx, cancel := context.WithTimeout(ctx, refresh+2*time.Second)
		defer cancel()
		if err := completion(cleanupCtx); err != nil {
			captureCleanupWorkers(t, completion)
		}
	})
	identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: database, SessionID: "root"}
	row, err := registry.Register(ctx, identity, "", "session.created")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Bind(ctx, identity, initialPublicationTicket, row.Request); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Change(ctx, identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle, r.TicketType, r.Activity = "Completion wake fixture", "fix", "busy"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if err := capture.LaunchWorker(cliBinaryPath, f.project, registry.CoordinationDirectory()); err != nil {
		t.Fatal(err)
	}
	wait := func(deadline time.Time, after int64, provisional bool) (capture.PublicationRequest, domain.AgenticFactLedger) {
		t.Helper()
		for time.Now().Before(deadline) {
			request, err := registry.PublicationRequest(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if request.Generation > after && request.Generation == request.Acknowledged && adapters.CheckPublication(f.output) == nil {
				body, err := os.ReadFile(f.output)
				var ledger domain.AgenticFactLedger
				if err == nil && json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) == 1 && len(ledger.Sessions[0].Invocations) == 1 {
					invocation := ledger.Sessions[0].Invocations[0]
					if (invocation.ProvisionalSeconds > 0) == provisional {
						return request, ledger
					}
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("worker did not commit the %s snapshot before deadline, ledger=%s", map[bool]string{true: "provisional", false: "complete"}[provisional], f.output)
		return capture.PublicationRequest{}, domain.AgenticFactLedger{}
	}
	before, active := wait(time.Now().Add(8*time.Second), 0, true)
	assertCaptureRootInput(t, active, 17)
	oldWorker, err := capture.WorkerStatus(registry.CoordinationDirectory())
	if err != nil || oldWorker.PID <= 0 {
		t.Fatalf("first publisher was not observed: %+v %v", oldWorker, err)
	}

	completed := created + 8000
	finishedMessage := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":%d,"completed":%d},"tokens":{"input":17,"output":2}}`, created, completed)
	finishedPart := fmt.Sprintf(`{"type":"tool","tool":"bash","state":{"status":"completed","time":{"start":%d,"end":%d}}}`, created, completed)
	if _, err := db.Exec("UPDATE message SET time_updated=?,data=? WHERE id='m'", completed, finishedMessage); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE part SET time_updated=?,data=? WHERE id='p'", completed, finishedPart); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(capture.HookInput{WorkingDirectory: f.project, SessionID: "root", Event: "session.idle"})
	if err != nil {
		t.Fatal(err)
	}
	notified := time.Now()
	if out, err := runCaptureNative(t, f.project, string(payload), "capture", "opencode-hook", "receive"); err != nil {
		t.Fatalf("OpenCode idle callback: %v\n%s", err, out)
	}
	request, done := wait(notified.Add(wakeBudget), before.Generation, false)
	assertCaptureRootInput(t, done, 17)
	if duration := done.Sessions[0].Invocations[0].ActiveDurationSeconds; duration != 8 {
		t.Fatalf("completed native interval = %ds, want eight source seconds exactly once", duration)
	}
	publishedAt, err := time.Parse(time.RFC3339Nano, done.UpdatedAt)
	if err != nil || publishedAt.Before(time.UnixMilli(request.RequestedAt).Add(debounce)) {
		t.Fatalf("completion bypassed the configured debounce: updated=%q request=%d err=%v", done.UpdatedAt, request.RequestedAt, err)
	}
	newWorker, err := capture.WorkerStatus(registry.CoordinationDirectory())
	if err != nil || newWorker.PID != oldWorker.PID {
		t.Fatalf("terminal request changed worker instead of waking it: before=%d after=%d err=%v", oldWorker.PID, newWorker.PID, err)
	}
	t.Logf("native idle callback to committed acknowledgement: %s with refresh %s and debounce %s", time.Since(notified), refresh, debounce)
}

func TestCaptureInterruptedInitialPublicationRecoversExactTransaction(t *testing.T) {
	for _, committed := range []bool{false, true} {
		phase := "partial artifacts"
		if committed {
			phase = "committed artifacts with uncleared marker"
		}
		t.Run(phase, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
			f.output = initialPublicationTarget(t, f, "interrupted-home")
			summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
			original := []byte("# Authored notes before publication\n")
			if err := os.WriteFile(summary, original, 0600); err != nil {
				t.Fatal(err)
			}
			prepareInitialPublication(t, f, f.output)
			paths := []string{f.output, summary}
			durableLedger := []byte(sampleFactJSON(initialPublicationTicket, "Initial durable publication", 8, 0))
			durableSummary := []byte("# Authored notes with committed metrics\n")
			marker := f.output + ".publication-pending"

			func() {
				cfg, err := adaptersconfig.Load(f.project)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				initial, err := capture.AcquireInitialPublication(ctx, f.project, adapters.PublicationTicketKey(initialPublicationTicket, cfg.Ticket), nil)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := initial.Close(); err != nil {
						t.Error(err)
					}
				}()
				selection, err := initial.Read()
				if err != nil || selection == nil {
					t.Fatalf("private selection unavailable after prepare: %+v %v", selection, err)
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
				if err := adapters.BeginInitialPublication(f.output, paths, selection); err != nil {
					t.Fatal(err)
				}
				target := []byte("{")
				if committed {
					target = durableLedger
				}
				if err := capture.WritePublicationFile(f.output, target, 0600); err != nil {
					t.Fatal(err)
				}
				if !committed {
					return
				}
				if err := capture.WritePublicationFile(summary, durableSummary, 0600); err != nil {
					t.Fatal(err)
				}
				markerBody := readCaptureFile(t, marker)
				journal := captureFixturePrivateJournal(t, f)
				journalBody := readCaptureFile(t, journal)
				if err := adapters.CommitPublication(f.output, paths); err != nil {
					t.Fatal(err)
				}
				// Restore the original marker and recovery journal to reproduce
				// interruption after durable consumption and before public cleanup.
				if err := os.WriteFile(journal, journalBody, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(marker, markerBody, 0600); err != nil {
					t.Fatal(err)
				}
				if err := capture.WritePublicationFile(summary, []byte("changed before recovery"), 0600); err != nil {
					t.Fatal(err)
				}
				recovered, err := adapters.RecoverPublication(f.output, paths)
				if err == nil || !recovered {
					t.Fatalf("recovery accepted a modified durable set: recovered=%t %v", recovered, err)
				}
				if err := adapters.CheckPublication(f.output); err == nil {
					t.Fatal("the changed set was reported as complete")
				}
				if err := capture.WritePublicationFile(summary, durableSummary, 0600); err != nil {
					t.Fatal(err)
				}
				recovered, err = adapters.RecoverPublication(f.output, paths)
				if err != nil || !recovered {
					t.Fatalf("committed transaction did not finish cleanup: recovered=%t %v", recovered, err)
				}
				if !bytes.Equal(readCaptureFile(t, f.output), durableLedger) || !bytes.Equal(readCaptureFile(t, summary), durableSummary) {
					t.Fatal("recovery reverted a durable first publication")
				}
				assertInitialPublicationAbsent(t, marker, journal)
			}()

			if !committed {
				if err := adapters.CheckPublication(f.output); err == nil {
					t.Fatal("an interrupted set was readable as complete")
				}
				prepareCaptureOwner(t, f, "root", initialPublicationTicket, 23)
				if output, err := runInitialPublicationWorker(t, f); err != nil {
					t.Fatalf("worker failed to recover and retry the exact initial selection: %v\n%s", err, output)
				}
				assertCaptureRootInput(t, f.ledger(t), 23)
				if body := readCaptureFile(t, summary); !bytes.HasPrefix(body, original) || bytes.Equal(body, original) {
					t.Fatalf("retry did not publish metrics over the recovered original summary: %s", body)
				}
				assertInitialPublicationAbsent(t, marker)
			}
			if err := adapters.CheckPublication(f.output); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCaptureRecoversTwoInterruptedPreparedTicketsInOneProject(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	fixtures := []struct {
		ticket, owner, destination string
		input                      int
	}{
		{"TEST-100", "root", initialPublicationTarget(t, f, "first-staged-home"), 17},
		{"TEST-200", "second-root", initialPublicationTarget(t, f, "second-staged-home"), 31},
	}
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range fixtures {
		output, err := runCaptureNative(t, f.project, "", "capture", "prepare-publication",
			"--repo", f.project, "--ticket-id", entry.ticket, "--json-output", entry.destination)
		if err != nil {
			t.Fatalf("prepare %s before either publication starts: %v\n%s", entry.ticket, err, output)
		}
	}
	const introduction = "# Preserve the authored introduction\n"
	for _, entry := range fixtures {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			lease, err := capture.AcquireInitialPublication(ctx, f.project, adapters.PublicationTicketKey(entry.ticket, cfg.Ticket), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := lease.Close(); err != nil {
					t.Error(err)
				}
			}()
			selection, err := lease.Read()
			if err != nil || selection == nil {
				t.Fatalf("read prepared selection for %s: %+v %v", entry.ticket, selection, err)
			}
			summary := filepath.Join(filepath.Dir(entry.destination), "SUMMARY.md")
			if err := os.WriteFile(summary, []byte(introduction), 0600); err != nil {
				t.Fatal(err)
			}
			owner, err := capture.AcquirePublication(ctx, entry.destination, f.project)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := owner.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := adapters.BeginInitialPublication(entry.destination, []string{entry.destination, summary}, selection); err != nil {
				t.Fatal(err)
			}
			// The old process died after staging these artifacts. The valid
			// replacement ledger must still be read as incomplete.
			replacement := []byte(sampleFactJSON(entry.ticket, "Interrupted replacement", 8, 0))
			if err := capture.WritePublicationFile(entry.destination, replacement, 0600); err != nil {
				t.Fatal(err)
			}
			if err := capture.WritePublicationFile(summary, []byte("Uncommitted summary"), 0600); err != nil {
				t.Fatal(err)
			}
		}()
		if err := adapters.CheckPublication(entry.destination); !errors.Is(err, adapters.ErrPublicationPending) {
			t.Fatalf("%s was not marked incomplete before restart: %v", entry.ticket, err)
		}
		prepareCaptureOwner(t, f, entry.owner, entry.ticket, entry.input)
	}

	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("worker failed to recover both prepared destinations: %v\n%s", err, output)
	}
	for _, entry := range fixtures {
		f.output = entry.destination
		ledger := f.ledger(t)
		if ledger.TicketID == nil || *ledger.TicketID != entry.ticket ||
			ledger.TicketTitle == nil || *ledger.TicketTitle != "Explicit "+entry.ticket ||
			len(ledger.Sessions) != 1 {
			t.Fatalf("recovered ledger retained an uncommitted artifact or wrong ticket: %+v", ledger)
		}
		session := ledger.Sessions[0]
		if session.ID != entry.owner || session.Source != domain.SourceCodex ||
			len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != int64(entry.input) {
			t.Fatalf("recovered %s lost its own exact source: %+v", entry.ticket, session)
		}
		summary := readCaptureFile(t, filepath.Join(filepath.Dir(entry.destination), "SUMMARY.md"))
		if !bytes.HasPrefix(summary, []byte(introduction)) || bytes.Contains(summary, []byte("Uncommitted summary")) || len(summary) <= len(introduction) {
			t.Fatalf("recovered %s did not publish over its own accepted introduction: %s", entry.ticket, summary)
		}
		assertInitialPublicationAbsent(t, entry.destination+".publication-pending")
	}
}

func TestCaptureConsumedSelectionPreservesCanonicalLedgerOwnership(t *testing.T) {
	for _, scenario := range []string{"recover ordinary publication", "recover ordinary publication after same-home import", "reject prior ledger for another ticket", "follow moved ledger past former marker"} {
		t.Run(scenario, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
			f.output = initialPublicationTarget(t, f, "original-home")
			prepareInitialPublication(t, f, f.output)
			const introduction = "# Authored introduction survives recovery\n"
			summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
			if err := os.WriteFile(summary, []byte(introduction), 0600); err != nil {
				t.Fatal(err)
			}
			wantTitle := "Explicit " + initialPublicationTicket
			imported := scenario == "recover ordinary publication after same-home import"
			if imported {
				wantTitle = "Portable ticket title"
				portable := []byte(sampleFactJSON(initialPublicationTicket, wantTitle, 8, 0))
				if err := os.WriteFile(f.output, portable, 0600); err != nil {
					t.Fatal(err)
				}
			}
			prepareCaptureOwner(t, f, "root", initialPublicationTicket, 11)
			if output, err := runInitialPublicationWorker(t, f); err != nil {
				t.Fatalf("initial automatic commit: %v\n%s", err, output)
			}
			previous := f.ledger(t)
			assertCaptureRootInput(t, previous, 11)
			if imported && (len(previous.Sessions) != 3 || previous.TicketTitle == nil || *previous.TicketTitle != wantTitle) {
				t.Fatalf("worker did not recognize the same-home imported ledger: %+v", previous)
			}
			cfg, err := adaptersconfig.Load(f.project)
			if err != nil {
				t.Fatal(err)
			}
			readSelection := func(ticket string) *capture.InitialPublication {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				lease, err := capture.AcquireInitialPublication(ctx, f.project, adapters.PublicationTicketKey(ticket, cfg.Ticket), nil)
				if err != nil {
					t.Fatal(err)
				}
				selection, readErr := lease.Read()
				if err := errors.Join(readErr, lease.Close()); err != nil || selection == nil {
					t.Fatalf("read publication selection for %s: %+v %v", ticket, selection, err)
				}
				return selection
			}
			before := readSelection(initialPublicationTicket)
			if imported {
				if !before.ExistingLedgerObserved || before.CommittedTransaction != "" {
					t.Fatalf("same-home import was not retired as an existing ledger: %+v", before)
				}
			} else if before.CommittedTransaction == "" {
				t.Fatal("first automatic commit did not consume the selection")
			}
			stage := func(destination string, selection *capture.InitialPublication, writeReplacement bool) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				owner, err := capture.AcquirePublication(ctx, destination, f.project)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := owner.Close(); err != nil {
						t.Error(err)
					}
				}()
				outputSummary := filepath.Join(filepath.Dir(destination), "SUMMARY.md")
				if err := adapters.BeginInitialPublication(destination, []string{destination, outputSummary}, selection); err != nil {
					t.Fatal(err)
				}
				if !writeReplacement {
					return
				}
				// The interrupted replacement must not determine recovery
				// authority. Only the backed-up canonical ledger can do so.
				interrupted := []byte(sampleFactJSON(initialPublicationTicket, "Uncommitted replacement title", 9, 0))
				if err := capture.WritePublicationFile(destination, interrupted, 0600); err != nil {
					t.Fatal(err)
				}
				if err := capture.WritePublicationFile(outputSummary, []byte("Uncommitted replacement summary"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			original := f.output
			marker := original + ".publication-pending"
			switch scenario {
			case "recover ordinary publication", "recover ordinary publication after same-home import":
				stage(original, nil, true)
			case "reject prior ledger for another ticket":
				foreign := []byte(sampleFactJSON("TEST-200", "Another ticket's baseline", 13, 0))
				if err := os.WriteFile(original, foreign, 0600); err != nil {
					t.Fatal(err)
				}
				stage(original, nil, true)
			case "follow moved ledger past former marker":
				f.output = filepath.Join(f.project, "records", "moved-home", "agentic_metrics.json")
				if err := os.Rename(filepath.Dir(original), filepath.Dir(f.output)); err != nil {
					t.Fatal(err)
				}
				summary = filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
				if err := os.Mkdir(filepath.Dir(original), 0700); err != nil {
					t.Fatal(err)
				}
				other := "TEST-200"
				output, err := runCaptureNative(t, f.project, "", "capture", "prepare-publication", "--repo", f.project, "--ticket-id", other, "--json-output", original)
				if err != nil {
					t.Fatalf("prepare other ticket at former home: %v\n%s", err, output)
				}
				stage(original, readSelection(other), false)
			}
			markerBefore := readCaptureFile(t, marker)
			f.writeAt(t, 43, "2026-01-01T00:00:04Z")
			if scenario == "reject prior ledger for another ticket" {
				local := filepath.Join(f.project, ".agenticlens-config.local.yaml")
				short := bytes.Replace(readCaptureFile(t, local), []byte("max_lifetime_ms: 10000"), []byte("max_lifetime_ms: 1800"), 1)
				if err := os.WriteFile(local, short, 0600); err != nil {
					t.Fatal(err)
				}
				interruptedLedger, interruptedSummary := readCaptureFile(t, original), readCaptureFile(t, summary)
				if output, err := runInitialPublicationWorker(t, f); err == nil {
					t.Fatalf("the consumed selection authorized a different ticket's backup: %s", output)
				}
				if !bytes.Equal(interruptedLedger, readCaptureFile(t, original)) ||
					!bytes.Equal(interruptedSummary, readCaptureFile(t, summary)) ||
					!bytes.Equal(markerBefore, readCaptureFile(t, marker)) {
					t.Fatal("rejected recovery touched the different ticket's interrupted set")
				}
				return
			}

			if output, err := runInitialPublicationWorker(t, f); err != nil {
				t.Fatalf("automatic refresh of canonical ledger: %v\n%s", err, output)
			}
			ledger := f.ledger(t)
			assertCaptureRootInput(t, ledger, 43)
			if ledger.TicketTitle == nil || *ledger.TicketTitle != wantTitle || len(ledger.Sessions) != len(previous.Sessions) {
				t.Fatalf("refresh accepted the wrong baseline: %+v", ledger)
			}
			for _, session := range previous.Sessions {
				if session.ID == "root" {
					continue
				}
				preserved := false
				for _, current := range ledger.Sessions {
					if reflect.DeepEqual(current, session) {
						preserved = true
						break
					}
				}
				if !preserved {
					t.Fatalf("recovery lost a prior portable owner: %s", session.ID)
				}
			}
			if body := readCaptureFile(t, summary); !bytes.HasPrefix(body, []byte(introduction)) || bytes.Contains(body, []byte("Uncommitted replacement")) {
				t.Fatalf("refresh did not preserve the authored summary: %s", body)
			}
			if after := readSelection(initialPublicationTicket); !reflect.DeepEqual(before, after) {
				t.Fatalf("ordinary refresh changed the consumed first receipt: before=%+v after=%+v", before, after)
			}
			assertInitialPublicationAbsent(t, f.output+".publication-pending")
			if scenario == "follow moved ledger past former marker" {
				assertInitialPublicationAbsent(t, original)
				if after := readCaptureFile(t, marker); !bytes.Equal(markerBefore, after) {
					t.Fatal("refresh changed the other ticket's pending marker at the former home")
				}
			}
		})
	}
}

func TestCaptureTerminalCodexDiscoveryContinuesOnlyWhileProgressing(t *testing.T) {
	const childID = "019c0000-0000-7000-8000-000000000074"
	for _, childAvailable := range []bool{true, false} {
		t.Run(map[bool]string{true: "child after multiple passes", false: "unresolved hint eventually stops"}[childAvailable], func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.output = initialPublicationTarget(t, f, "terminal-discovery")
			policy := []byte("telemetry:\n  capture:\n    refresh_ms: 100\n    idle_ms: 250\n    max_lifetime_ms: 5000\n    max_sessions: 5\n")
			if err := writeCaptureFixtureLocal(t, f, policy); err != nil {
				t.Fatal(err)
			}
			prepareInitialPublication(t, f, f.output)
			path := prepareCaptureOwner(t, f, "root", initialPublicationTicket, 11)
			result := fmt.Sprintf(`{"agent_id":%q}`, childID)
			events := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{}"}}
{"timestamp":"2026-01-01T00:00:04Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":%q}}
`, result)
			if err := os.WriteFile(path, append(readCaptureFile(t, path), events...), 0600); err != nil {
				t.Fatal(err)
			}
			sessionDir := filepath.Join(f.store, "sessions")
			// A directory walk has its own entry budget. Keep the real
			// rollouts within the exact-source extraction budget.
			for number := range 320 {
				name := filepath.Join(sessionDir, fmt.Sprintf("auxiliary-%03d.json", number))
				if err := os.WriteFile(name, []byte("{}\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			for number := range 3 {
				decoy := fmt.Sprintf("019c0000-0000-7000-8000-%012d", number+100)
				body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q,"source":"cli"}}`, decoy, f.project)
				if err := os.WriteFile(filepath.Join(sessionDir, "rollout-2026-01-01T00-00-00-"+decoy+".jsonl"), []byte(body+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if childAvailable {
				body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q,"source":{"subagent":{"thread_spawn":{"parent_thread_id":"root","depth":1}}}}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"child-turn","usage":{"input_tokens":29,"output_tokens":2}}}
`, childID, f.project, childID)
				if err := os.WriteFile(filepath.Join(sessionDir, "rollout-2026-01-01T00-00-01-"+childID+".jsonl"), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx := context.Background()
			registry, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := registry.Close(); err != nil {
					t.Error(err)
				}
			})
			completion := captureWorkerCompletion(t, registry.CoordinationDirectory())
			workerLog, err := os.Create(filepath.Join(t.TempDir(), "terminal-worker.log"))
			if err != nil {
				t.Fatal(err)
			}
			defer workerLog.Close()
			workerCtx, stop := context.WithCancel(ctx)
			t.Cleanup(stop)
			cmd := exec.CommandContext(workerCtx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
			cmd.Dir, cmd.Stdout, cmd.Stderr = f.project, workerLog, workerLog
			if err := registry.RequestPublication(ctx); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			workerDone := make(chan struct{})
			go func() {
				_ = cmd.Wait()
				close(workerDone)
			}()
			t.Cleanup(func() {
				disabled := bytes.Replace(policy, []byte("  capture:\n"), []byte("  capture:\n    enabled: false\n"), 1)
				if err := writeCaptureFixtureLocal(t, f, disabled); err != nil {
					t.Error(err)
				}
				if err := registry.RequestPublication(ctx); err != nil {
					t.Error(err)
				}
				select {
				case <-workerDone:
				case <-time.After(2 * time.Second):
					stop()
					<-workerDone
					t.Error("terminal fixture worker did not stop after disabling capture")
				}
				captureCleanupWorkers(t, completion)
			})
			// No further callbacks or requests. Only the worker may advance discovery.
			if childAvailable {
				deadline := time.Now().Add(6 * time.Second)
				for time.Now().Before(deadline) {
					body, err := os.ReadFile(f.output)
					var ledger domain.AgenticFactLedger
					if err == nil && adapters.CheckPublication(f.output) == nil && json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) == 2 {
						assertCaptureRootInput(t, ledger, 11)
						for _, session := range ledger.Sessions {
							if session.ID == childID && session.ParentID != nil && *session.ParentID == "root" &&
								len(session.Invocations) == 1 && session.Invocations[0].Tokens.Input == 29 {
								return
							}
						}
						t.Fatalf("second owner was not the exact source child: %+v", ledger.Sessions)
					}
					time.Sleep(20 * time.Millisecond)
				}
				body, readErr := os.ReadFile(f.output)
				t.Fatalf("terminal worker abandoned a child after bounded discovery: read=%v ledger=%s worker=%s", readErr, body, readCaptureFile(t, workerLog.Name()))
			}

			deadline, cancel := context.WithTimeout(ctx, 6*time.Second)
			defer cancel()
			select {
			case <-workerDone:
			case <-deadline.Done():
				t.Fatalf("an unavailable child hint kept the worker running: %s", readCaptureFile(t, workerLog.Name()))
			}
			if err := completion(deadline); err != nil {
				t.Fatalf("an unavailable child hint retained the worker indefinitely: %v", err)
			}
			if _, err := os.Stat(f.output); err != nil {
				t.Fatalf("terminal worker did not commit the proven root: %v worker=%s", err, readCaptureFile(t, workerLog.Name()))
			}
			ledger := f.ledger(t)
			if len(ledger.Sessions) != 1 {
				t.Fatalf("an unavailable child hint fabricated an owner: %+v", ledger.Sessions)
			}
			assertCaptureRootInput(t, ledger, 11)
			sawPending := false
			for _, diagnostic := range captureDiagnostics(t, f.project, f.state) {
				if diagnostic.TicketID == initialPublicationTicket && diagnostic.Reconciliation != nil && diagnostic.Reconciliation.PendingChildren > 0 {
					sawPending = true
					break
				}
			}
			if !sawPending {
				t.Fatal("unavailable child evidence was silently dropped")
			}
		})
	}
}

func TestCaptureTerminalTicketSweepsStayClosedAcrossWorkerLifetimes(t *testing.T) {
	f := newCodexInboxFixture(t)
	first := initialPublicationTarget(t, f, "first-terminal")
	second := initialPublicationTarget(t, f, "second-terminal")
	// Two native rollouts fit the source budget. The auxiliary directory
	// entries force more bounded sweep passes than one worker can perform.
	const passBudget = 2
	policy := []byte("telemetry:\n  capture:\n    refresh_ms: 100\n    idle_ms: 100\n    max_lifetime_ms: 1500\n    max_sessions: 2\n")
	if err := writeCaptureFixtureLocal(t, f, policy); err != nil {
		t.Fatal(err)
	}
	for ticket, destination := range map[string]string{"TEST-100": first, "TEST-200": second} {
		if output, err := runCaptureNative(t, f.project, "", "capture", "prepare-publication",
			"--repo", f.project, "--ticket-id", ticket, "--json-output", destination); err != nil {
			t.Fatalf("prepare %s: %v\n%s", ticket, err, output)
		}
	}
	for number := range 550 {
		path := filepath.Join(f.store, "sessions", fmt.Sprintf("auxiliary-%04d.json", number))
		if err := os.WriteFile(path, []byte("{}\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	prepareCaptureOwner(t, f, "root", "TEST-100", 19)
	prepareCaptureOwner(t, f, "second-root", "TEST-200", 31)
	ctx := context.Background()
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	})
	rows, err := registry.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	firstKey := ""
	for _, row := range rows {
		if row.TicketID == "TEST-100" {
			firstKey = capture.IdentityKey(row.Identity)
		}
		missing := "missing-child-for-" + row.Identity.SessionID
		if err := registry.QueueChildStart(ctx, row.Identity.Store, row.Identity.SessionID, missing); err != nil {
			t.Fatal(err)
		}
	}
	cursors, err := registry.CodexChildDiscoveryCursors(ctx)
	_, firstQueued := cursors[firstKey]
	if err != nil || !firstQueued || len(cursors) != 2 {
		t.Fatalf("fixture did not start with two pending sweeps: %+v %v", cursors, err)
	}
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	reconcile := func(ticket string) {
		t.Helper()
		result, err := registry.ReconcileAt(ctx, ticket, "Fixture owner", passBudget, 64<<20, time.Now().UTC(), cfg.Generation.Metrics)
		if err != nil || result.PendingChildren != 1 || len(result.Sessions) != 1 {
			t.Fatalf("%s lost the unresolved hint or its exact owner: %+v %v", ticket, result, err)
		}
	}
	closed := false
	for range 64 {
		reconcile("TEST-100")
		cursors, err = registry.CodexChildDiscoveryCursors(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, pending := cursors[firstKey]; !pending {
			closed = true
			break
		}
	}
	if !closed || len(cursors) != 1 {
		t.Fatalf("fixture did not close only the first sweep: %+v", cursors)
	}
	completion := captureWorkerCompletion(t, registry.CoordinationDirectory())
	t.Cleanup(func() {
		disabled := bytes.Replace(policy, []byte("  capture:\n"), []byte("  capture:\n    enabled: false\n"), 1)
		if err := writeCaptureFixtureLocal(t, f, disabled); err != nil {
			t.Error(err)
		}
		if err := registry.RequestPublication(ctx); err != nil {
			t.Error(err)
		}
		captureCleanupWorkers(t, completion)
	})
	if err := registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if err := capture.LaunchWorker(fixtureCaptureEngine(), f.project, registry.CoordinationDirectory()); err != nil {
		t.Fatal(err)
	}
	// The race suite shares CPU with other native process tests. Give the
	// observer time to see the fixed 1.5s workers advance without changing
	// the worker lifetime, cursor assertions, or handoff requirement.
	const observationBudget = 45 * time.Second
	wait, cancel := context.WithTimeout(ctx, observationBudget)
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- completion(wait) }()
	ticker := time.NewTicker(35 * time.Millisecond)
	defer ticker.Stop()
	workers := make(map[int]bool)
	var workerPIDs []int
	lastCursors := cursors
	var lastProgress capture.WorkerProgress
	for {
		select {
		case err := <-finished:
			if err != nil {
				t.Fatalf("terminal worker chain did not drain within %s: %v; observed worker PIDs=%v; last active cursors=%+v; last worker status=%+v",
					observationBudget, err, workerPIDs, lastCursors, lastProgress)
			}
			if len(workers) < 2 {
				t.Fatalf("fixture did not exercise a finite worker handoff: observed worker PIDs=%v; last worker status=%+v", workerPIDs, lastProgress)
			}
			for _, ticket := range []string{"TEST-100", "TEST-200", "TEST-100", "TEST-200"} {
				reconcile(ticket)
			}
			cursors, err := registry.CodexChildDiscoveryCursors(ctx)
			if err != nil || len(cursors) != 0 {
				t.Fatalf("static reconciliation reopened a completed sweep: %+v %v", cursors, err)
			}
			for _, fixture := range []struct {
				path, id string
				input    int64
			}{{first, "root", 19}, {second, "second-root", 31}} {
				f.output = fixture.path
				ledger := f.ledger(t)
				if len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != fixture.id ||
					len(ledger.Sessions[0].Invocations) != 1 || ledger.Sessions[0].Invocations[0].Tokens.Input != fixture.input {
					t.Fatalf("terminal sweep did not publish only its exact bound owner: %+v", ledger)
				}
			}
			t.Logf("%d bounded workers completed both terminal sweeps", len(workers))
			return
		case <-ticker.C:
			cursors, err := registry.CodexChildDiscoveryCursors(ctx)
			if _, reopened := cursors[firstKey]; err != nil || reopened {
				t.Fatalf("later worker pass reopened the first static sweep: cursors=%+v err=%v; observed worker PIDs=%v; last worker status=%+v",
					cursors, err, workerPIDs, lastProgress)
			}
			lastCursors = cursors
			progress, err := capture.WorkerStatus(registry.CoordinationDirectory())
			if err != nil {
				t.Fatal(err)
			}
			lastProgress = progress
			if progress.PID > 0 && !workers[progress.PID] {
				workers[progress.PID] = true
				workerPIDs = append(workerPIDs, progress.PID)
			}
		}
	}
}

func TestCaptureFirstPublicationRefusesMissingUnsafeAndAmbiguousSelection(t *testing.T) {
	for _, scenario := range []struct {
		name, failure string
	}{
		{"missing", "prepare-publication"},
		{"outside roots", "prepare-publication"},
		{"ambiguous ledgers", "ambiguous"},
		{"replaced directory", "no longer matches"},
		{"consumed then deleted", "selection was consumed"},
		{"portable ledger then deleted", "retired after observing an existing ledger"},
		{"unrelated pending transaction", "does not match its prepared selection"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
			f.output = initialPublicationTarget(t, f, "selected")
			decoy := initialPublicationTarget(t, f, initialPublicationTicket)
			unchanged := make(map[string][]byte)
			marker := f.output + ".publication-pending"
			switch scenario.name {
			case "outside roots":
				outside := filepath.Join(t.TempDir(), "agentic_metrics.json")
				output, err := runCaptureNative(t, f.project, "", "capture", "prepare-publication", "--repo", f.project, "--ticket-id", initialPublicationTicket, "--json-output", outside)
				if err == nil {
					t.Fatalf("outside path selected: %s", output)
				}
				assertInitialPublicationAbsent(t, outside)
			case "ambiguous ledgers":
				prepareInitialPublication(t, f, f.output)
				for _, dir := range []string{"existing-one", "existing-two"} {
					path := initialPublicationTarget(t, f, dir)
					if err := os.WriteFile(path, []byte(sampleFactJSON(initialPublicationTicket, dir, 3, 0)), 0600); err != nil {
						t.Fatal(err)
					}
					unchanged[path] = readCaptureFile(t, path)
				}
			case "replaced directory":
				prepareInitialPublication(t, f, f.output)
				if err := os.Rename(filepath.Dir(f.output), filepath.Join(f.project, "records", "held-original")); err != nil {
					t.Fatal(err)
				}
				initialPublicationTarget(t, f, "selected")
			case "consumed then deleted":
				prepareInitialPublication(t, f, f.output)
				prepareCaptureOwner(t, f, "root", initialPublicationTicket, 11)
				if output, err := runInitialPublicationWorker(t, f); err != nil {
					t.Fatalf("initial publication before deletion: %v\n%s", err, output)
				}
				assertCaptureRootInput(t, f.ledger(t), 11)
				if err := os.Remove(f.output); err != nil {
					t.Fatal(err)
				}
			case "portable ledger then deleted":
				prepareInitialPublication(t, f, f.output)
				portable := f
				portable.output = initialPublicationTarget(t, f, "portable-home")
				body := []byte(sampleFactJSON(initialPublicationTicket, "Portable ticket title", 8, 0))
				if err := os.WriteFile(portable.output, body, 0600); err != nil {
					t.Fatal(err)
				}
				prepareCaptureOwner(t, f, "root", initialPublicationTicket, 11)
				if output, err := runInitialPublicationWorker(t, f); err != nil {
					t.Fatalf("automatic refresh of portable ledger: %v\n%s", err, output)
				}
				ledger := portable.ledger(t)
				assertCaptureRootInput(t, ledger, 11)
				if ledger.TicketTitle == nil || *ledger.TicketTitle != "Portable ticket title" || len(ledger.Sessions) != 3 {
					t.Fatalf("refresh did not use the imported ledger and preserve its two owners: %+v", ledger)
				}
				assertInitialPublicationAbsent(t, f.output)
				if err := os.Remove(portable.output); err != nil {
					t.Fatal(err)
				}
			case "unrelated pending transaction":
				prepareInitialPublication(t, f, f.output)
				owner, err := capture.AcquirePublication(context.Background(), f.output, f.project)
				if err != nil {
					t.Fatal(err)
				}
				summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
				beginErr := adapters.BeginPublication(f.output, []string{f.output, summary})
				if err := errors.Join(beginErr, owner.Close()); err != nil {
					t.Fatal(err)
				}
				unchanged[marker] = readCaptureFile(t, marker)
			}

			prepareCaptureOwner(t, f, "root", initialPublicationTicket, 13)
			localPath := filepath.Join(f.project, ".agenticlens-config.local.yaml")
			local := string(readCaptureFile(t, localPath))
			local = strings.Replace(local, "max_lifetime_ms: 10000", "max_lifetime_ms: 1400", 1)
			if err := os.WriteFile(localPath, []byte(local), 0600); err != nil {
				t.Fatal(err)
			}
			output, err := runInitialPublicationWorker(t, f)
			if err == nil || !bytes.Contains(output, []byte(scenario.failure)) {
				t.Fatalf("worker did not diagnose %q: %v\n%s", scenario.failure, err, output)
			}
			failed := false
			for _, diagnostic := range captureDiagnostics(t, f.project, f.state) {
				if diagnostic.TicketID == initialPublicationTicket && diagnostic.Phase == "publication" && diagnostic.Outcome == "failed" {
					failed = true
				}
			}
			if !failed {
				t.Fatalf("worker did not record the failed ticket publication: %v\n%s", err, output)
			}
			assertInitialPublicationAbsent(t, f.output, decoy)
			if scenario.name != "unrelated pending transaction" {
				assertInitialPublicationAbsent(t, marker)
			}
			for path, before := range unchanged {
				if !bytes.Equal(readCaptureFile(t, path), before) {
					t.Fatalf("rejected publication changed an existing artifact: %s", path)
				}
			}
		})
	}
}
