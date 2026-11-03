//go:build windows

package capture_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

const windowsWatchdogFixtureEnv = "AL_WINDOWS_WATCHDOG_TEST_FIXTURE"

type windowsWatchdogIdentity struct {
	PID        int
	Created    uint64
	Executable string
}

func TestMain(m *testing.M) {
	if directory := os.Getenv(windowsWatchdogFixtureEnv); directory != "" {
		if err := runWindowsWatchdogFixture(directory); err != nil {
			_ = os.WriteFile(filepath.Join(directory, "error"), []byte(err.Error()), 0600)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func windowsFixtureExecutable(parts ...string) string {
	return filepath.Join(append([]string{os.Getenv("SystemRoot"), "System32"}, parts...)...)
}

// The native launcher records child identities immediately after CreateProcess;
// PowerShell need not execute a single statement to establish the owned tree.
func runWindowsWatchdogFixture(directory string) error {
	if err := writeWindowsFixtureRecord(directory, "parent", os.Getpid()); err != nil {
		return err
	}
	for _, child := range []struct {
		name string
		cmd  *exec.Cmd
	}{
		{"powershell", exec.Command(windowsFixtureExecutable("WindowsPowerShell", "v1.0", "powershell.exe"), "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "[Threading.Thread]::Sleep(30000)")},
		{"native", exec.Command(windowsFixtureExecutable("ping.exe"), "-t", "127.0.0.1")},
	} {
		// Inherit the hook's pipe so leaked descendants also break its bounded
		// return, rather than being invisible behind detached fixture handles.
		child.cmd.Stdout, child.cmd.Stderr = os.Stdout, os.Stderr
		if err := child.cmd.Start(); err != nil {
			return fmt.Errorf("start %s fixture: %w", child.name, err)
		}
		defer func() { _ = child.cmd.Process.Kill(); _ = child.cmd.Wait() }()
		if err := writeWindowsFixtureRecord(directory, child.name, child.cmd.Process.Pid); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "ready"), []byte("ready"), 0600); err != nil {
		return err
	}
	time.Sleep(30 * time.Second)
	return nil
}

func writeWindowsFixtureRecord(directory, name string, pid int) error {
	identity, err := windowsFixtureIdentity(pid)
	if err != nil {
		return err
	}
	data, err := json.Marshal(identity)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, name+".json"), data, 0600)
}

func windowsFixtureIdentity(pid int) (windowsWatchdogIdentity, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return windowsWatchdogIdentity{}, err
	}
	defer windows.CloseHandle(handle)
	return windowsFixtureIdentityFromHandle(handle, pid)
}

func windowsFixtureIdentityFromHandle(handle windows.Handle, pid int) (windowsWatchdogIdentity, error) {
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &created, &exited, &kernel, &user); err != nil {
		return windowsWatchdogIdentity{}, err
	}
	var path [32768]uint16
	size := uint32(len(path))
	if err := windows.QueryFullProcessImageName(handle, 0, &path[0], &size); err != nil {
		return windowsWatchdogIdentity{}, err
	}
	return windowsWatchdogIdentity{PID: pid, Created: uint64(created.HighDateTime)<<32 | uint64(created.LowDateTime), Executable: windows.UTF16ToString(path[:size])}, nil
}

func readWindowsFixtureRecord(path string) (windowsWatchdogIdentity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return windowsWatchdogIdentity{}, err
	}
	var identity windowsWatchdogIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return identity, err
	}
	if identity.PID <= 0 || identity.Created == 0 || identity.Executable == "" {
		return identity, fmt.Errorf("invalid process identity in %s: %s", path, data)
	}
	return identity, nil
}

func windowsFixtureHandle(identity windowsWatchdogIdentity, access uint32) (windows.Handle, error) {
	handle, err := windows.OpenProcess(access|windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(identity.PID))
	if err != nil {
		return 0, err
	}
	actual, err := windowsFixtureIdentityFromHandle(handle, identity.PID)
	if err != nil || actual.Created != identity.Created || !strings.EqualFold(actual.Executable, identity.Executable) {
		// A terminated process may retain a PID handle while Windows has
		// already discarded its image metadata. It needs no termination.
		if state, waitErr := windows.WaitForSingleObject(handle, 0); waitErr == nil && state == windows.WAIT_OBJECT_0 {
			return handle, nil
		}
		windows.CloseHandle(handle)
		return 0, fmt.Errorf("process %d identity changed: actual=%+v expected=%+v error=%v", identity.PID, actual, identity, err)
	}
	return handle, nil
}

func assertWindowsFixtureRunning(identity windowsWatchdogIdentity) error {
	handle, err := windowsFixtureHandle(identity, 0)
	if err != nil {
		return fmt.Errorf("fixture %d unavailable: %w", identity.PID, err)
	}
	defer windows.CloseHandle(handle)
	state, err := windows.WaitForSingleObject(handle, 0)
	if err != nil || state != uint32(windows.WAIT_TIMEOUT) {
		return fmt.Errorf("fixture %d not running: wait=%d error=%v", identity.PID, state, err)
	}
	return nil
}

func assertWindowsFixtureStopped(identity windowsWatchdogIdentity) error {
	handle, err := windowsFixtureHandle(identity, 0)
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return nil // Windows no longer has this PID.
	}
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	state, err := windows.WaitForSingleObject(handle, 0)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("process %d (%s) still running: wait=%d error=%v", identity.PID, identity.Executable, state, err)
	}
	return nil
}

func cleanupWindowsFixtureRecord(t *testing.T, path, executable string) {
	t.Helper()
	identity, err := readWindowsFixtureRecord(path)
	if errors.Is(err, os.ErrNotExist) {
		return // Cleanup can run before fixture launch; assertions require records.
	}
	if err != nil {
		t.Errorf("read fixture cleanup identity: %v", err)
		return
	}
	if !strings.EqualFold(identity.Executable, executable) {
		t.Errorf("refusing cleanup of unexpected executable: %+v", identity)
		return
	}
	handle, err := windowsFixtureHandle(identity, windows.PROCESS_TERMINATE)
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return
	}
	if err != nil {
		t.Errorf("open fixture cleanup identity: %v", err)
		return
	}
	defer windows.CloseHandle(handle)
	state, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		t.Errorf("inspect fixture cleanup process: %v", err)
		return
	}
	if state == windows.WAIT_OBJECT_0 {
		return
	}
	if err := windows.TerminateProcess(handle, 1); err != nil {
		t.Errorf("terminate owned fixture: %v", err)
		return
	}
	if state, err := windows.WaitForSingleObject(handle, 1000); err != nil || state != windows.WAIT_OBJECT_0 {
		t.Errorf("reap owned fixture: wait=%d error=%v", state, err)
		return
	}
	t.Logf("fixture cleanup reaped process %d (%s), creation identity %d", identity.PID, identity.Executable, identity.Created)
}

func windowsFixtureDiagnostic(directory string, err error) string {
	data, readErr := os.ReadFile(filepath.Join(directory, "error"))
	return fmt.Sprintf("exit=%v helper_error=%q read_error=%v", err, data, readErr)
}
