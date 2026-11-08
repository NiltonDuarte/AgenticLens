//go:build linux

package main_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
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

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// The disposable Git helper reports the actual generator executable before
// pausing it. This observes native A/B identity without product test branches.
func init() {
	if os.Getenv("AL_CAPTURE_SWITCH_GIT_HELPER") != "1" {
		return
	}
	if strings.Join(os.Args[1:], " ") == "rev-parse --show-toplevel" {
		native, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", os.Getppid()))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(91)
		}
		conn, err := net.DialTimeout("tcp", os.Getenv("AL_CAPTURE_SWITCH_GATE"), 5*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(92)
		}
		if err = conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(93)
		}
		if _, err = fmt.Fprintln(conn, native); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(94)
		}
		if _, err = io.ReadFull(conn, make([]byte, 1)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(95)
		}
		if err = conn.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(96)
		}
	}
	command := exec.Command(os.Getenv("AL_CAPTURE_SWITCH_REAL_GIT"), os.Args[1:]...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(97)
	}
	os.Exit(0)
}

func TestCaptureWorkerRetainsAdmittedBuildAcrossPreferredSwitch(t *testing.T) {
	t.Run("pending generation", func(t *testing.T) { testCaptureWorkerReleaseSwitch(t, true) })
	t.Run("active acknowledged root", func(t *testing.T) { testCaptureWorkerReleaseSwitch(t, false) })
}
func testCaptureWorkerReleaseSwitch(t *testing.T, queueNext bool) {
	f := newCaptureGenerationFixture(t)
	prepareCaptureOwner(t, f, "root", "TEST-100", 37)
	originalConfig := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
	cache := t.TempDir()
	writePin := func(version string) {
		t.Helper()
		header := fmt.Sprintf("agenticlens:\n  version: %s\n  install_root: %q\n", version, cache)
		if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), append([]byte(header), originalConfig...), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writePin("v9.7.1")
	native := func(version string) string {
		t.Helper()
		writePin(version)
		entry, err := adaptersconfig.ResolveEntrypoint(f.project)
		if err != nil {
			t.Fatal(err)
		}
		path, err := entry.NativePath()
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		build := exec.Command("go", "build", "-ldflags", "-X agenticlens/internal/domain.Version="+version, "-o", path, ".")
		if out, err := build.CombinedOutput(); err != nil {
			t.Fatalf("native %s build: %v %s", version, err, out)
		}
		return path
	}
	first, second := native("v9.7.1"), native("v9.7.2")
	writePin("v9.7.1")
	local := fmt.Sprintf("telemetry:\n  capture:\n    state_dir: %q\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 10000\n", f.state)
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(cache, "common wrapper")
	if err := os.WriteFile(wrapper, readCaptureFile(t, filepath.Join("..", "..", "..", "scripts", "agentic-lens-wrapper.sh")), 0700); err != nil {
		t.Fatal(err)
	}
	install := func(binary string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := capture.Install(ctx, capture.InstallOptions{Repository: f.project, StateDir: f.coordination(t), Binary: binary, Wrapper: wrapper, Source: "codex", SourceStore: f.store}); err != nil {
			t.Fatal(err)
		}
	}
	install(first)
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	}()
	if !queueNext {
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
		if _, err = registry.Register(context.Background(), id, filepath.Join(f.store, "sessions", "root.jsonl"), "UserPromptSubmit"); err != nil {
			t.Fatal(err)
		}
		if _, err = registry.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			row.Activity = "busy"
			return row, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err = registry.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	realGit, err := exec.LookPath("git")
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
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			t.Error(err)
		}
	}()
	if err = listener.(*net.TCPListener).SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "PATH="+shim+string(os.PathListSeparator)+os.Getenv("PATH"), "AL_CAPTURE_SWITCH_GIT_HELPER=1", "AL_CAPTURE_SWITCH_GATE="+listener.Addr().String(), "AL_CAPTURE_SWITCH_REAL_GIT="+realGit, "GORACE=atexit_sleep_ms=0")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	old := exec.CommandContext(ctx, first, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	old.Dir = f.project
	old.Env = env
	old.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	log, err := os.CreateTemp(t.TempDir(), "old-worker")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := log.Close(); err != nil {
			t.Error(err)
		}
	}()
	old.Stdout, old.Stderr = log, log
	if err = old.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if old.ProcessState == nil {
			if err := syscall.Kill(-old.Process.Pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
				t.Error(err)
			}
			if err := old.Wait(); err == nil {
				t.Error("interrupted old worker unexpectedly succeeded")
			}
		}
	}()
	accept := func() (net.Conn, string) {
		t.Helper()
		conn, err := listener.Accept()
		if err != nil {
			t.Fatalf("native pass gate unavailable: %v; old log=%s", err, readCaptureFile(t, log.Name()))
		}
		if err = conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			t.Fatal(err)
		}
		name, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		return conn, strings.TrimSpace(name)
	}
	release := func(conn net.Conn) {
		t.Helper()
		if _, err := conn.Write([]byte("C")); err != nil {
			t.Fatal(err)
		}
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}
	paused, identity := accept()
	if identity == first || !bytes.Equal(readCaptureFile(t, identity), readCaptureFile(t, first)) {
		t.Fatalf("first pass did not use an immutable copy of admitted build: %s", identity)
	}
	retained := identity
	progress, err := capture.WorkerStatus(f.coordination(t))
	if err != nil {
		t.Fatal(err)
	}
	if progress.PID != old.Process.Pid || !strings.Contains(progress.Detail, "generation=1 acknowledged=0") {
		t.Fatalf("old captured generation not established: %+v", progress)
	}
	writePin("v9.7.2")
	install(second)
	wanted := int64(1)
	if queueNext {
		wanted = 2
		if err = registry.RequestPublication(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	contender := exec.CommandContext(ctx, second, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	contender.Dir = f.project
	contender.Env = env
	if out, err := contender.CombinedOutput(); err != nil {
		t.Fatalf("selected worker lock contention: %v %s", err, out)
	}
	// Replace the mutable source at the same path while A is still running.
	replacement := first + ".new"
	if err := os.WriteFile(replacement, readCaptureFile(t, second), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, first); err != nil {
		t.Fatal(err)
	}
	release(paused)
	later, laterIdentity := accept()
	if laterIdentity != retained {
		t.Fatalf("active worker changed later child build: %s want %s", laterIdentity, retained)
	}
	release(later)
	// Continue to release real A children until the admitted owner exits normally.
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
				_ = conn.Close()
				return
			}
			_, err = bufio.NewReader(conn).ReadString('\n')
			if err == nil {
				_, err = conn.Write([]byte("C"))
			}
			closeErr := conn.Close()
			if err != nil || closeErr != nil {
				return
			}
		}
	}()
	if !queueNext {
		f.writeAt(t, 73, "2026-01-01T00:00:03Z")
		waitCorrectionLedger(t, filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json"), "root", 73)
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
		if _, err = registry.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			row.Activity = "idle"
			return row, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := old.Wait(); err != nil {
		t.Fatalf("admitted worker failed normal completion: %v %s", err, readCaptureFile(t, log.Name()))
	}
	request, err := registry.PublicationRequest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if request.Generation < wanted || request.Acknowledged != request.Generation {
		t.Fatalf("admitted build stranded pending work: %+v", request)
	}
	entry, err := adaptersconfig.ResolveEntrypoint(f.project)
	if err != nil {
		t.Fatal(err)
	}
	preferred, err := entry.NativePath()
	if err != nil || preferred != second {
		t.Fatalf("new invocation did not prefer B: %s %v", preferred, err)
	}
	out, err := exec.Command(preferred, "version", "--short").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "v9.7.2" {
		t.Fatalf("preferred build identity: %s %v", out, err)
	}
}
