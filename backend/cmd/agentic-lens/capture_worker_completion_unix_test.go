//go:build !windows

package main_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
)

// Workers, their generators and successors inherit this log descriptor. Once
// callbacks have returned, EOF proves all their writers have closed, including
// registry cleanup after the worker lease is released.
func captureWorkerCompletion(t *testing.T, directory string) func(context.Context) error {
	t.Helper()
	path := filepath.Join(directory, "worker.log")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	// Keep EOF from occurring between fixture setup and the final callback.
	keeper, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		if closeErr := keeper.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		t.Fatal(err)
	}
	var once sync.Once
	var closeErr error
	release := func() { once.Do(func() { closeErr = keeper.Close() }) }
	finished := make(chan struct{})
	var drainErr error
	go func() {
		_, drainErr = io.Copy(io.Discard, reader)
		close(finished)
	}()
	t.Cleanup(func() {
		release()
		if err := errors.Join(closeErr, reader.Close()); err != nil {
			t.Error(err)
		}
	})
	return func(ctx context.Context) error {
		release()
		select {
		case <-finished:
			return errors.Join(closeErr, drainErr)
		case <-ctx.Done():
			return errors.Join(closeErr, ctx.Err())
		}
	}
}
