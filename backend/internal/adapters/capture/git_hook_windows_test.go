//go:build windows

package capture_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

// This native Windows regression requires Git for Windows and normal PowerShell
// script execution. It fails when either boundary is unavailable; Linux tests
// exercise the POSIX watchdog; native Windows test runs exercise this case.
func TestGitCheckWindowsWatchdogKillsOwnedTree(t *testing.T) {
	repo := checkRepository(t)
	fixture := t.TempDir()
	childPID := filepath.Join(fixture, "child.pid")
	parentPID := filepath.Join(fixture, "parent.pid")
	wrapper := filepath.Join(fixture, "wrapper with spaces.ps1")
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	powershell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	// An independent sibling must survive: the watchdog owns only its mapped
	// child tree, even when another native process has the same executable name.
	control := exec.Command(filepath.Join(os.Getenv("SystemRoot"), "System32", "ping.exe"), "-t", "127.0.0.1")
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
	body := "[IO.File]::WriteAllText(" + quote(parentPID) + ", ([string]$PID + '|' + [string]([Diagnostics.Process]::GetCurrentProcess().StartTime.ToUniversalTime().Ticks)))\n" +
		"$start = [Diagnostics.ProcessStartInfo]::new()\n" +
		"$start.FileName = [IO.Path]::Combine($env:SystemRoot, 'System32', 'ping.exe')\n" +
		"$start.Arguments = '-t 127.0.0.1'\n$start.UseShellExecute = $false\n" +
		"$child = [Diagnostics.Process]::Start($start)\n" +
		"[IO.File]::WriteAllText(" + quote(childPID) + ", ([string]$child.Id + '|' + [string]$child.StartTime.ToUniversalTime().Ticks))\n" +
		// The outer watchdog must act while the inner deadline is delayed.
		"Start-Sleep -Seconds 30\n$child.WaitForExit()\n"
	checkWrite(t, wrapper, body, 0600)
	// Cleanup is scoped to recorded fixture PIDs with a verified executable path.
	// It also runs if the regression strands a process or the assertion fails.
	for _, record := range []struct{ path, executable string }{
		{childPID, filepath.Join(os.Getenv("SystemRoot"), "System32", "ping.exe")},
		{parentPID, powershell},
	} {
		t.Cleanup(func() {
			data, err := os.ReadFile(record.path)
			if os.IsNotExist(err) {
				return
			}
			if err != nil {
				t.Errorf("read owned process cleanup identity: %v", err)
				return
			}
			identity := strings.SplitN(strings.TrimSpace(string(data)), "|", 2)
			if len(identity) != 2 {
				t.Errorf("invalid owned process identity %q", data)
				return
			}
			pid, err := strconv.Atoi(identity[0])
			if err != nil || pid <= 0 {
				t.Errorf("invalid owned process identity %q: %v", data, err)
				return
			}
			started, err := strconv.ParseInt(identity[1], 10, 64)
			if err != nil || started <= 0 {
				t.Errorf("invalid owned process start identity %q: %v", data, err)
				return
			}
			code := "$p = Get-Process -Id " + strconv.Itoa(pid) + " -ErrorAction SilentlyContinue; if ($p) { if ($p.Path -ne " + quote(record.executable) + " -or $p.StartTime.ToUniversalTime().Ticks -ne [long]" + quote(strconv.FormatInt(started, 10)) + ") { throw 'Owned process identity changed' }; Stop-Process -Id $p.Id -Force }; exit 0"
			if output, err := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", code).CombinedOutput(); err != nil {
				t.Errorf("cleanup owned process: %v %s", err, output)
			}
		})
	}
	installCheck(t, capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper})
	command := checkGitCommand(repo, "commit", "--quiet", "--allow-empty", "-m", "owned process tree deadline")
	// Bound the test's parent wait independently if production cleanup regresses.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	bounded := exec.CommandContext(ctx, command.Path, command.Args[1:]...)
	bounded.Env = command.Env
	bounded.WaitDelay = time.Second
	var stdout, stderr bytes.Buffer
	bounded.Stdout, bounded.Stderr = &stdout, &stderr
	started := time.Now()
	err := bounded.Run()
	elapsed := time.Since(started)
	if err != nil || elapsed > 7*time.Second || stdout.Len() != 0 || !strings.Contains(stderr.String(), "check unavailable") || strings.Count(stderr.String(), "\n") != 1 {
		t.Fatalf("Windows advisory did not bound its process tree: %v elapsed=%v stdout=%q stderr=%q", err, elapsed, stdout.String(), stderr.String())
	}
	for _, path := range []string{parentPID, childPID} {
		data := checkRead(t, path)
		pid, err := strconv.Atoi(strings.SplitN(strings.TrimSpace(string(data)), "|", 2)[0])
		if err != nil || pid <= 0 {
			t.Fatalf("invalid actual Windows process identity: %q %v", data, err)
		}
		code := "if (Get-Process -Id " + strconv.Itoa(pid) + " -ErrorAction SilentlyContinue) { exit 1 }; exit 0"
		if output, err := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", code).CombinedOutput(); err != nil {
			t.Fatalf("owned process %d survived outer watchdog: %v %s", pid, err, output)
		}
	}
	code := "if (Get-Process -Id " + strconv.Itoa(control.Process.Pid) + " -ErrorAction SilentlyContinue) { exit 0 }; exit 1"
	if output, err := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", code).CombinedOutput(); err != nil {
		t.Fatalf("watchdog affected independent native sibling: %v %s", err, output)
	}
	t.Logf("Windows Git commit returned zero after %v; actual PowerShell/native child PIDs are absent and independent sibling survived", elapsed)
}
