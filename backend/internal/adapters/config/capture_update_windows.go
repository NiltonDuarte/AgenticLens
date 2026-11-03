package config

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

type captureConfigLock struct {
	file  *os.File
	mutex windows.Handle
}

func (lock *captureConfigLock) Stat() (os.FileInfo, error) { return lock.file.Stat() }
func (lock *captureConfigLock) ReadAt(data []byte, offset int64) (int, error) {
	return lock.file.ReadAt(data, offset)
}

func (lock *captureConfigLock) prepareReplace() error {
	// Windows denies replacement of an open/byte-locked target on the observed
	// native filesystem. Keep kernel-mutex ownership while closing only the read
	// handle; readers continue to see the valid old path until atomic activation.
	err := lock.file.Close()
	lock.file = nil
	return err
}

func (lock *captureConfigLock) Close() error {
	var fileErr error
	if lock.file != nil {
		fileErr = lock.file.Close()
	}
	err := errors.Join(fileErr, windows.ReleaseMutex(lock.mutex), windows.CloseHandle(lock.mutex))
	runtime.UnlockOSThread()
	return err
}

func lockCaptureConfig(path string) (*captureConfigLock, error) {
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(canonical))))
	name, err := windows.UTF16PtrFromString(`Global\AgenticLensConfig-` + hex.EncodeToString(digest[:]))
	if err != nil {
		return nil, err
	}
	// Kernel mutex ownership belongs to the OS thread. Keep that thread admitted
	// until ReleaseMutex; process/thread death releases an abandoned mutex.
	runtime.LockOSThread()
	mutex, err := windows.CreateMutex(nil, false, name)
	if err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		runtime.UnlockOSThread()
		return nil, err
	}
	result, err := windows.WaitForSingleObject(mutex, 0)
	if err != nil || (result != windows.WAIT_OBJECT_0 && result != windows.WAIT_ABANDONED) {
		closeErr := windows.CloseHandle(mutex)
		runtime.UnlockOSThread()
		if err == nil {
			err = fmt.Errorf("%w (mutex wait result %d)", errCaptureConfigBusy, result)
		}
		return nil, errors.Join(err, closeErr)
	}
	fileName, err := windows.UTF16PtrFromString(path)
	if err != nil {
		releaseErr := windows.ReleaseMutex(mutex)
		closeErr := windows.CloseHandle(mutex)
		runtime.UnlockOSThread()
		return nil, errors.Join(err, releaseErr, closeErr)
	}
	handle, err := windows.CreateFile(fileName, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		releaseErr := windows.ReleaseMutex(mutex)
		closeErr := windows.CloseHandle(mutex)
		runtime.UnlockOSThread()
		return nil, errors.Join(err, releaseErr, closeErr)
	}
	return &captureConfigLock{file: os.NewFile(uintptr(handle), path), mutex: mutex}, nil
}

func replaceCaptureConfig(staged, path string) error {
	from, err := windows.UTF16PtrFromString(staged)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}

func captureConfigLockBusy(err error) bool { return errors.Is(err, errCaptureConfigBusy) }
