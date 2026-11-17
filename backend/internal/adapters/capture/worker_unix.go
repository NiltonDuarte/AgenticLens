//go:build linux || darwin

package capture

import (
	"os/exec"
	"syscall"
)

func configureWorkerProcess(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} }
