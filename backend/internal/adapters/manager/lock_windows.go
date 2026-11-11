package manager

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func lockFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	var overlap windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlap); err != nil {
		return nil, errors.Join(errors.New("manager installation is busy"), err, file.Close())
	}
	return file, nil
}

// Windows directory handles do not support the POSIX directory-fsync contract.
func syncDirectory(string) error { return nil }
