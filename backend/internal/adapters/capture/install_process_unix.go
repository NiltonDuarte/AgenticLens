//go:build linux || darwin

package capture

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Installer inspection owns its launcher group, including inherited pipe writers.
func installCommandOutput(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 100 * time.Millisecond
	killGroup := func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.Cancel = killGroup
	output, err := command.Output()
	cleanupErr := killGroup()
	if errors.Is(cleanupErr, os.ErrProcessDone) {
		cleanupErr = nil
	}
	return output, errors.Join(err, cleanupErr, ctx.Err())
}
