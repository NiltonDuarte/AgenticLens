package main_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"golang.org/x/sys/unix"
)

func TestCaptureDiscardDefersWhileLegacyPassOutlivesWorker(t *testing.T) {
	ctx := context.Background()
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := capture.ReceiveCodexUser(ctx, store, capture.HookInput{
		SessionID: "root", Event: "SessionEnd", WorkingDirectory: f.project, TranscriptPath: f.write(t, 17),
	}, f.store, cliBinaryPath, f.state); err != nil {
		t.Fatal(err)
	}
	id := capture.StableCaptureID(domain.SourceCodex, "root")
	statusBefore, err := store.CaptureControl(ctx, id)
	if err != nil || statusBefore.State != capture.CaptureControlActive {
		t.Fatalf("fixture root did not become active: %+v, %v", statusBefore, err)
	}

	oldLease, err := capture.AcquireWorker(store.CoordinationDirectory())
	if err != nil || oldLease == nil {
		t.Fatalf("legacy fixture could not take ownership: %v", err)
	}
	t.Cleanup(func() {
		if oldLease != nil {
			oldLease.Close()
		}
	})
	group, retire := startLegacyOrphanPass(t)
	if err := capture.WriteWorkerProgress(store.CoordinationDirectory(), capture.WorkerProgress{PID: group, State: "running"}); err != nil {
		t.Fatal(err)
	}
	if err := oldLease.Close(); err != nil {
		t.Fatal(err)
	}
	oldLease = nil
	out, err := runCaptureNative(t, f.project, "", "capture", "discard", "--repo", f.project, "--capture-id", id)
	if err == nil || !bytes.Contains(out, []byte("may still be active; retry after it retires")) {
		t.Fatalf("discard did not defer for the orphaned legacy pass: %v\n%s", err, out)
	}
	if status, err := store.CaptureControl(ctx, id); err != nil || status.State != statusBefore.State || status.Revision != statusBefore.Revision || status.CleanupPending {
		t.Fatalf("refused discard changed durable state: %+v, %v", status, err)
	}
	worker, err := capture.AcquireLifecycleWorker(store.CoordinationDirectory())
	if worker != nil {
		worker.Close()
	}
	if worker != nil || err == nil || !strings.Contains(err.Error(), "may still be active") {
		t.Fatalf("compatible worker did not defer for the orphaned legacy pass: %v", err)
	}

	retire()
	if err := syscall.Kill(-group, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("legacy fixture group still exists after retirement: %v", err)
	}
	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	if out, err := runCaptureNative(t, f.project, "", "capture", "discard", "--repo", f.project, "--capture-id", id); err != nil {
		t.Fatalf("discard failed after legacy group retired: %v\n%s", err, out)
	}
	if status, err := store.CaptureControl(ctx, id); err != nil || status.State != capture.CaptureControlDiscarded || status.Revision <= statusBefore.Revision {
		t.Fatalf("successful retry did not persist discarding this root: %+v, %v", status, err)
	}
}

func startLegacyOrphanPass(t *testing.T) (int, func()) {
	t.Helper()
	// Adoption lets the fixture reap its own orphan even when the runner's PID 1
	// does not reap. The test runs serially and restores the previous setting.
	var previous int32
	if err := unix.Prctl(unix.PR_GET_CHILD_SUBREAPER, uintptr(unsafe.Pointer(&previous)), 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(previous), 0, 0, 0); err != nil {
			t.Error(err)
		}
	})
	passInput, release, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer passInput.Close()
	deferClose := true
	defer func() {
		if deferClose {
			release.Close()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	leader := exec.CommandContext(ctx, "sh", "-c", `(IFS= read -r done) <&3 >/dev/null 2>&1 & printf '%s\n' "$!"`)
	leader.ExtraFiles = []*os.File{passInput}
	leader.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	out, err := leader.Output()
	if err != nil {
		t.Fatal(err)
	}
	group := leader.Process.Pid
	child, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || child <= 1 {
		t.Fatalf("legacy fixture did not report its pass PID: %q, %v", out, err)
	}
	retired := false
	retire := func() {
		t.Helper()
		if retired {
			return
		}
		retired = true
		if err := release.Close(); err != nil {
			t.Error(err)
		}
		deadline := time.Now().Add(2 * time.Second)
		for {
			var state syscall.WaitStatus
			pid, err := syscall.Wait4(child, &state, syscall.WNOHANG, nil)
			if err != nil || pid == child {
				if err != nil {
					t.Error(err)
				}
				return
			}
			if !time.Now().Before(deadline) {
				_ = syscall.Kill(child, syscall.SIGKILL)
				_, _ = syscall.Wait4(child, &state, 0, nil)
				t.Error("legacy fixture pass did not stop when its input closed")
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	deferClose = false
	t.Cleanup(retire)
	if err := syscall.Kill(group, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("wanted an exited, reaped group leader: %v", err)
	}
	if pgid, err := syscall.Getpgid(child); err != nil || pgid != group {
		t.Fatalf("legacy pass %d did not retain group %d: got %d, %v", child, group, pgid, err)
	}
	return group, retire
}
