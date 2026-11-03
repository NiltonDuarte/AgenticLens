//go:build windows

package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// A suspended start assigns ownership before a launcher can create children.
// Closing the job cleans descendants even when the launcher exits first.
func installCommandOutput(ctx context.Context, binary string, args ...string) (output []byte, returnErr error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create installer inspection job: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, windows.CloseHandle(job)) }()
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, fmt.Errorf("bound installer inspection job: %w", err)
	}
	command := exec.CommandContext(ctx, binary, args...)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED, HideWindow: true}
	command.WaitDelay = 100 * time.Millisecond
	command.Cancel = func() error {
		jobErr := windows.TerminateJobObject(job, 1)
		// Cancellation can precede assignment while the child is still suspended.
		killErr := command.Process.Kill()
		if errors.Is(killErr, os.ErrProcessDone) {
			killErr = nil
		}
		return errors.Join(jobErr, killErr)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Start(); err != nil {
		return nil, err
	}
	abort := func(cause error) ([]byte, error) {
		killErr := command.Process.Kill()
		if errors.Is(killErr, os.ErrProcessDone) {
			killErr = nil
		}
		return nil, errors.Join(cause, killErr, command.Wait(), ctx.Err())
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(command.Process.Pid))
	if err != nil {
		return abort(fmt.Errorf("open suspended installer inspection: %w", err))
	}
	assignErr := windows.AssignProcessToJobObject(job, process)
	if err := errors.Join(assignErr, windows.CloseHandle(process)); err != nil {
		return abort(fmt.Errorf("own installer inspection job: %w", err))
	}
	if err := resumeInstallerInspection(uint32(command.Process.Pid)); err != nil {
		return abort(err)
	}
	err = command.Wait()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		exit.Stderr = append([]byte(nil), stderr.Bytes()...)
	}
	return stdout.Bytes(), errors.Join(err, ctx.Err())
}

func resumeInstallerInspection(pid uint32) (returnErr error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return fmt.Errorf("inspect suspended installer thread: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, windows.CloseHandle(snapshot)) }()
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	err = windows.Thread32First(snapshot, &entry)
	for err == nil {
		if entry.OwnerProcessID == pid {
			thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
			if err != nil {
				return fmt.Errorf("open suspended installer thread: %w", err)
			}
			_, resumeErr := windows.ResumeThread(thread)
			return errors.Join(resumeErr, windows.CloseHandle(thread))
		}
		err = windows.Thread32Next(snapshot, &entry)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return fmt.Errorf("enumerate suspended installer thread: %w", err)
	}
	return errors.New("suspended installer inspection thread unavailable")
}
