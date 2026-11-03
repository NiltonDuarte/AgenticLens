package main_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// Exclusive access is possible only after every inherited native log handle has
// closed. A successor opens its handle before its predecessor exits.
func captureWorkerCompletion(t *testing.T, directory string) func(context.Context) error {
	t.Helper()
	file, err := os.Create(filepath.Join(directory, "worker.log"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	path, err := windows.UTF16PtrFromString(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return func(ctx context.Context) error {
		for {
			handle, err := windows.CreateFile(path, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
			if err == nil {
				return windows.CloseHandle(handle)
			}
			if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
				return err
			}
			timer := time.NewTimer(10 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
}
