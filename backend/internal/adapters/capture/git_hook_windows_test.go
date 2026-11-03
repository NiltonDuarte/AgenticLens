//go:build windows

package capture_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

// Run the rendered production hook through Git for Windows. Neither case needs
// PowerShell to finish its own startup before the three-second watchdog fires.
func TestGitCheckWindowsWatchdogKillsOwnedTree(t *testing.T) {
	t.Run("PowerShellLateBody", func(t *testing.T) {
		repo, fixture := checkRepository(t), t.TempDir()
		marker := filepath.Join(fixture, "late-body.pid")
		wrapper := filepath.Join(fixture, "wrapper with spaces.ps1")
		// A PID written by the script body is not evidence that a native process
		// was launched: the outer watchdog can kill it before this statement.
		body := "[Threading.Thread]::Sleep(10000)\n[IO.File]::WriteAllText('" + strings.ReplaceAll(marker, "'", "''") + "', [string]$PID)\n"
		checkWrite(t, wrapper, body, 0600)
		installCheck(t, capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper})
		runWindowsWatchdogCommit(t, repo, nil, nil)
		if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("watchdog allowed the deliberately late PowerShell body to run: %v", err)
		}
	})

	t.Run("CreatedNativeTree", func(t *testing.T) {
		repo, fixture := checkRepository(t), t.TempDir()
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		// Reuse the running test image. Native launch records do not depend on
		// PowerShell loading modules or entering its script body within 3s.
		installCheck(t, capture.CheckInstallOptions{Repository: repo, Wrapper: executable})
		control := exec.Command(windowsFixtureExecutable("ping.exe"), "-t", "127.0.0.1")
		if err := control.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := control.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				t.Errorf("stop independent fixture process: %v", err)
			}
			var terminated *exec.ExitError
			if err := control.Wait(); err != nil && !errors.As(err, &terminated) {
				t.Errorf("reap independent fixture process: %v", err)
			}
		})
		controlIdentity, err := windowsFixtureIdentity(control.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		records := []struct{ name, executable string }{
			{"parent", executable},
			{"powershell", windowsFixtureExecutable("WindowsPowerShell", "v1.0", "powershell.exe")},
			{"native", windowsFixtureExecutable("ping.exe")},
		}
		for _, record := range records {
			t.Cleanup(func() { cleanupWindowsFixtureRecord(t, filepath.Join(fixture, record.name+".json"), record.executable) })
		}
		var identities []windowsWatchdogIdentity
		runWindowsWatchdogCommit(t, repo, []string{windowsWatchdogFixtureEnv + "=" + fixture}, func(done <-chan error) error {
			// Readiness is mandatory. A missing record is a fixture failure, never
			// a successful cleanup assertion for a process that did not exist.
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				if _, err := os.Stat(filepath.Join(fixture, "ready")); err == nil {
					break
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}
				select {
				case err := <-done:
					return errors.New("Git returned before native fixture readiness: " + windowsFixtureDiagnostic(fixture, err))
				case <-ticker.C:
				}
			}
			for _, record := range records {
				identity, err := readWindowsFixtureRecord(filepath.Join(fixture, record.name+".json"))
				if err != nil {
					return err
				}
				if !strings.EqualFold(identity.Executable, record.executable) {
					return errors.New("unexpected fixture executable: " + identity.Executable)
				}
				if err := assertWindowsFixtureRunning(identity); err != nil {
					return err
				}
				identities = append(identities, identity)
			}
			return nil
		})
		for _, identity := range identities {
			if err := assertWindowsFixtureStopped(identity); err != nil {
				t.Errorf("owned Windows process survived watchdog: %v", err)
			}
		}
		if err := assertWindowsFixtureRunning(controlIdentity); err != nil {
			t.Errorf("watchdog affected independent native sibling: %v", err)
		}
		t.Log("checked native parent, PowerShell/native descendants and independent sibling using recorded creation identities")
	})
}

func runWindowsWatchdogCommit(t *testing.T, repo string, environment []string, ready func(<-chan error) error) {
	t.Helper()
	command := checkGitCommand(repo, "commit", "--quiet", "--allow-empty", "-m", "owned process tree deadline")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	bounded := exec.CommandContext(ctx, command.Path, command.Args[1:]...)
	bounded.Env = append(command.Env, environment...)
	bounded.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	bounded.Stdout, bounded.Stderr = &stdout, &stderr
	started := time.Now()
	if err := bounded.Start(); err != nil {
		t.Fatal(err)
	}
	// Readiness and Wait observe completion independently; neither can consume
	// the other's result and accidentally turn the 8s parent bound into a hang.
	done := make(chan error, 1)
	finished := make(chan struct{})
	var waitErr error
	go func() { waitErr = bounded.Wait(); done <- waitErr; close(finished) }()
	var readyErr error
	if ready != nil {
		readyErr = ready(done)
	}
	<-finished
	elapsed := time.Since(started)
	if readyErr != nil {
		t.Fatalf("Windows fixture was not observed ready: %v; Git=%v elapsed=%v stdout=%q stderr=%q", readyErr, waitErr, elapsed, stdout.String(), stderr.String())
	}
	if waitErr != nil || elapsed > 7*time.Second || stdout.Len() != 0 || !strings.Contains(stderr.String(), "check unavailable") || strings.Count(stderr.String(), "\n") != 1 {
		t.Fatalf("Windows advisory did not bound its process tree: %v elapsed=%v stdout=%q stderr=%q", waitErr, elapsed, stdout.String(), stderr.String())
	}
	t.Logf("Windows Git commit returned zero after %v with one bounded advisory", elapsed)
}
