//go:build !windows

package main_test

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCaptureNativeCrashReleasesOwnershipAndPreservesPendingGeneration(t *testing.T) {
	t.Run("crash takeover", func(t *testing.T) { testNativePendingGeneration(t, true) })
	t.Run("event during live pass", func(t *testing.T) { testNativePendingGeneration(t, false) })
}
func testNativePendingGeneration(t *testing.T, crash bool) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 37))
	local := []byte("telemetry:\n  capture:\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 10000\n")
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), local, 0600); err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	// Supply explicit metadata without invoking callbacks that launch detached owners.
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	if _, err = store.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle = "Crash recovery"
		r.TicketType = "feat"
		r.Activity = "idle"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = store.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	lock, err := capture.AcquirePublication(context.Background(), output)
	if err != nil {
		t.Fatal(err)
	}
	held := true
	defer func() {
		if held {
			lock.Close()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	worker := exec.CommandContext(ctx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	worker.Dir = f.project
	worker.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// Files avoid unbounded inherited pipe waits if a killed worker leaves a generator.
	log, err := os.CreateTemp(t.TempDir(), "worker-log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	worker.Stdout = log
	worker.Stderr = log
	if err = worker.Start(); err != nil {
		t.Fatal(err)
	}
	reaped := false
	defer func() {
		if !reaped {
			syscall.Kill(-worker.Process.Pid, syscall.SIGKILL)
			worker.Wait()
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	observed := false
	for time.Now().Before(deadline) {
		owner, e := capture.AcquireWorker(f.state)
		if e != nil {
			t.Fatal(e)
		}
		if owner == nil {
			observed = true
			break
		}
		owner.Close()
		time.Sleep(10 * time.Millisecond)
	}
	if !observed {
		t.Fatal("native worker never acquired ownership")
	}
	// The worker publishes its captured generation before entering the pass. The
	// output lock prevents acknowledgement while this observed generation is in flight.
	observed = false
	for time.Now().Before(deadline) {
		progress, e := capture.WorkerStatus(f.state)
		if e != nil {
			t.Fatal(e)
		}
		if progress.PID == worker.Process.Pid && strings.Contains(progress.Detail, "generation=1 acknowledged=0") {
			observed = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !observed {
		t.Fatal("worker did not expose captured generation before blocked pass")
	}
	if crash {
		if err = syscall.Kill(-worker.Process.Pid, syscall.SIGKILL); err != nil {
			t.Fatal(err)
		}
		_ = worker.Wait()
		reaped = true
		req, err := store.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if req.Generation <= req.Acknowledged {
			t.Fatalf("crashed pass swallowed request: %+v", req)
		}
		owner, err := capture.AcquireWorker(f.state)
		if err != nil || owner == nil {
			t.Fatalf("kernel ownership not released: %v", err)
		}
		owner.Close()

	}
	// Queue another generation during the interrupted pass; the replacement must
	// publish both requests rather than acknowledge the old pass optimistically.
	if err = store.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	held = false
	if crash {
		replacement := exec.CommandContext(ctx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
		replacement.Dir = f.project
		if out, e := replacement.CombinedOutput(); e != nil {
			t.Fatalf("takeover: %v %s", e, out)
		}
	} else {
		if err = worker.Wait(); err != nil {
			t.Fatalf("live pass: %v: %s", err, readCaptureFile(t, log.Name()))
		}
		reaped = true
	}
	req, err := store.PublicationRequest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if req.Generation != req.Acknowledged {
		t.Fatalf("takeover left pending work: %+v", req)
	}
	var ledger domain.AgenticFactLedger
	if err = json.Unmarshal(readCaptureFile(t, output), &ledger); err != nil {
		t.Fatal(err)
	}
	assertCaptureRootInput(t, ledger, 37)
}

func TestCaptureLateTicketMustPublishBeforeIdle(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 37))
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 100\n    max_lifetime_ms: 3000\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	if _, err = s.Change(ctx, id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle = "first"
		r.TicketType = "feat"
		r.Activity = "idle"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	helper, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	if err = os.WriteFile(filepath.Join(shim, "git"), readCaptureFile(t, helper), 0700); err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	l.(*net.TCPListener).SetDeadline(time.Now().Add(6 * time.Second))
	pctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	worker := exec.CommandContext(pctx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	worker.Dir = f.project
	worker.Env = append(os.Environ(), "PATH="+shim+string(os.PathListSeparator)+os.Getenv("PATH"), "AL_CAPTURE_TEST_GIT_HELPER=1", "AL_CAPTURE_TEST_GATE_ADDRESS="+l.Addr().String(), "AL_CAPTURE_TEST_REAL_GIT="+git, "GORACE=atexit_sleep_ms=0")
	log, err := os.CreateTemp(t.TempDir(), "log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	worker.Stdout = log
	worker.Stderr = log
	if err = worker.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if worker.ProcessState == nil {
			cancel()
			worker.Wait()
		}
	}()
	c, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(6 * time.Second))
	if _, err = io.ReadFull(c, make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	progress, err := capture.WorkerStatus(f.state)
	if err != nil {
		t.Fatal(err)
	}
	if progress.PID != worker.Process.Pid || !strings.Contains(progress.Detail, "generation=1 acknowledged=0") {
		t.Fatalf("captured generation not observable: %+v", progress)
	}
	// Worker already froze its ticket list. Bind another independently owned ticket.
	path := filepath.Join(f.store, "sessions", "late.jsonl")
	body := []byte(fmt.Sprintf(`{"type":"session_meta","payload":{"id":"late","cwd":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"late","turn_id":"one","usage":{"input_tokens":19}}}
`, f.project))
	if err = os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	late := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "late"}
	row, err := s.Register(ctx, late, path, "SessionEnd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Bind(ctx, late, "TEST-200", row.Request); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Change(ctx, late, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle = "late"
		r.TicketType = "feat"
		r.Activity = "idle"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	// Ensure the late request is older than idle_ms before the first pass ends:
	// the latest-generation ACK mutant must exit without a second pass.
	time.Sleep(150 * time.Millisecond)
	if _, err = c.Write([]byte("C")); err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			cc, e := l.Accept()
			if e != nil {
				return
			}
			cc.SetDeadline(time.Now().Add(time.Second))
			io.ReadFull(cc, make([]byte, 1))
			cc.Write([]byte("C"))
			cc.Close()
		}
	}()
	if err = worker.Wait(); err != nil {
		t.Fatalf("worker %v %s", err, readCaptureFile(t, log.Name()))
	}
	req, err := s.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(f.project, "records", "TEST-200", "agentic_metrics.json")
	if _, err = os.Stat(output); err != nil {
		t.Fatalf("acknowledged=%d generation=%d late ticket absent: %v", req.Acknowledged, req.Generation, err)
	}
	if req.Generation != req.Acknowledged {
		t.Fatalf("pending late generation: %+v", req)
	}
	var ledger domain.AgenticFactLedger
	if err = json.Unmarshal(readCaptureFile(t, output), &ledger); err != nil {
		t.Fatal(err)
	}
	if len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != "late" || len(ledger.Sessions[0].Invocations) != 1 || ledger.Sessions[0].Invocations[0].Tokens.Input != 19 {
		t.Fatalf("late source facts not published: %+v", ledger.Sessions)
	}
	if err = json.Unmarshal(readCaptureFile(t, filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")), &ledger); err != nil {
		t.Fatal(err)
	}
	assertCaptureRootInput(t, ledger, 37)
}
