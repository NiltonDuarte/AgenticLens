//go:build !windows

package main

import (
	"os"
	"syscall"
)

// execManualNative preserves the exposed PID, inherited handles and the actual
// signal termination status for manual CLI commands. Capture workers remain a
// separate deliberate process boundary.
func execManualNative(path string, args []string) (bool, error) {
	return true, syscall.Exec(path, append([]string{path}, args...), os.Environ())
}
