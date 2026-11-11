//go:build linux || darwin

package main_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

func init() {
	if helper := os.Getenv("AL_CAPTURE_BLOCKED_HARNESS"); helper == "" || filepath.Base(os.Args[0]) != "opencode" {
		return
	}
	if err := os.WriteFile(os.Getenv("AL_CAPTURE_PROBE_STARTED"), []byte("started"), 0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(93)
	}
	if path := os.Getenv("AL_CAPTURE_PROBE_PID"); path != "" {
		if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(94)
		}
	}
	// The installer must cancel this optional process; no machine-speed delay
	// decides whether the probe completes or exhausts its own allowance.
	for {
		time.Sleep(time.Hour)
	}
}

func TestCaptureInstallSIGTERMCancelsProbeChild(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	dir := t.TempDir()
	helper := filepath.Join(dir, "opencode")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(binary, helper); err != nil {
		t.Fatal(err)
	}
	args := prepareCaptureInstall(t, f.project, []string{"install", "--repo", f.project, "--state-dir", f.state, "--source", "opencode", "--store", filepath.Join(f.store, "source.db")})
	t.Setenv("AL_CAPTURE_HARNESS_FIXTURE", "0")
	t.Setenv("AL_CAPTURE_BLOCKED_HARNESS", helper)
	t.Setenv("AL_CAPTURE_PROBE_STARTED", filepath.Join(dir, "started"))
	pidPath := filepath.Join(dir, "child-pid")
	t.Setenv("AL_CAPTURE_PROBE_PID", pidPath)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, cliBinaryPath, append([]string{"capture"}, args...)...)
	command.Dir = f.project
	output, err := os.Create(filepath.Join(dir, "cli-output"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := output.Close(); err != nil {
			t.Error(err)
		}
	}()
	command.Stdout, command.Stderr = output, output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			cancel()
			if err := command.Wait(); err == nil {
				t.Error("canceled unfinished CLI unexpectedly succeeded")
			}
		}
	}()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	var pid int
	for pid == 0 {
		if body, err := os.ReadFile(pidPath); err == nil {
			pid, err = strconv.Atoi(string(body))
			if err != nil {
				t.Fatal(err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		if pid != 0 {
			break
		}
		select {
		case <-tick.C:
		case <-ctx.Done():
			t.Fatal("optional child did not reach its blocking probe")
		}
	}
	defer func() {
		if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			t.Error(err)
		}
	}()
	if err := command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	err = command.Wait()
	waited = true
	if err == nil || ctx.Err() != nil {
		t.Fatalf("SIGTERM cancellation not observed before watchdog: %v, %v", err, ctx.Err())
	}
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Fatalf("SIGTERM exited CLI but optional child %d remains: %v", pid, err)
	}
}

func TestCaptureInstallOptionalProbeCannotConsumeRequiredValidation(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	dir := t.TempDir()
	helper := filepath.Join(dir, "opencode")
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(testBinary, helper); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "probe-started")
	t.Setenv("AL_CAPTURE_BLOCKED_HARNESS", helper)
	t.Setenv("AL_CAPTURE_PROBE_STARTED", marker)
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	args := prepareCaptureInstall(t, f.project, []string{"install", "--repo", f.project, "--state-dir", f.state, "--source", "opencode", "--store", filepath.Join(f.store, "source.db")})
	t.Setenv("AL_CAPTURE_HARNESS_FIXTURE", "0")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliBinaryPath, append([]string{"capture"}, args...)...)
	cmd.Dir = f.project
	out, err := cmd.CombinedOutput()
	if _, markerErr := os.Stat(marker); markerErr != nil {
		t.Fatalf("optional harness probe was not reached: %v; %s", markerErr, out)
	}
	if err != nil {
		t.Fatalf("optional probe exhausted required installation validation: %v; %s", err, out)
	}
	var result capture.InstallResult
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("installation result: %v; %s", err, out)
	}
	if result.Manifest.HarnessVersion != "unavailable" || !result.Manifest.Active {
		t.Fatalf("optional failure did not preserve required installation: %+v", result.Manifest)
	}
	if err := capture.CheckInstalledState(f.project, result.Manifest.StateDir, true); err != nil {
		t.Fatalf("committed required native/wrapper state unhealthy: %v", err)
	}
}
