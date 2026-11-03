//go:build !windows

package config

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

type captureConfigLock struct{ *os.File }

func (lock *captureConfigLock) prepareReplace() error { return nil }

func lockCaptureConfig(path string) (*captureConfigLock, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if err = unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return &captureConfigLock{file}, nil
}

func replaceCaptureConfig(staged, path string) error { return os.Rename(staged, path) }

func captureConfigLockBusy(err error) bool {
	return errors.Is(err, errCaptureConfigBusy) || errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)
}
