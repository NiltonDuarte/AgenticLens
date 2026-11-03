//go:build !windows

package capture_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

func init() {
	if os.Getenv("AL_CAPTURE_ADAPTER_BLOCKING_PROBE") == "" || filepath.Base(os.Args[0]) != "opencode" {
		return
	}
	connection, err := net.DialTimeout("tcp", os.Getenv("AL_CAPTURE_ADAPTER_BLOCKING_PROBE"), 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(91)
	}
	if _, err := connection.Write([]byte("P")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(92)
	}
	if err := connection.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(93)
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestInstallCallerCancellationDuringOptionalProbe(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary, wrapper := installFixture(t, project)
	dir := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(executable, filepath.Join(dir, "opencode")); err != nil {
		t.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := listener.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AL_CAPTURE_ADAPTER_BLOCKING_PROBE", listener.Addr().String())
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	parent, parentCancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer parentCancel()
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	store := filepath.Join(t.TempDir(), "source.db")
	done := make(chan error, 1)
	go func() {
		_, err := capture.Install(ctx, capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "opencode", SourceStore: store})
		done <- err
	}()
	connection, err := listener.AcceptTCP()
	if err != nil {
		cancel()
		select {
		case installErr := <-done:
			t.Fatalf("optional probe not reached: %v; install: %v", err, installErr)
		case <-parent.Done():
			t.Fatalf("optional probe not reached: %v", err)
		}
	}
	if err := connection.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(connection, make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("caller cancellation ignored or cause lost: %v", err)
		}
	case <-parent.Done():
		t.Fatal("installation did not honor caller cancellation")
	}
	if _, err := os.Stat(filepath.Join(state, "install-opencode.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled installation committed a manifest: %v", err)
	}
}
