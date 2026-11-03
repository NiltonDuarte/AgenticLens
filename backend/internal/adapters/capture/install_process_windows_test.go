//go:build windows

package capture

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func init() {
	if os.Getenv("AL_CAPTURE_WINDOWS_PROCESS_FIXTURE") != "1" || len(os.Args) < 4 || os.Args[1] != "installer-process-fixture" {
		return
	}
	if err := windowsInstallerFixture(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(91)
	}
	os.Exit(0)
}

func windowsInstallerFixture(args []string) error {
	if args[0] == "pipe-child" {
		connection, err := net.DialTimeout("tcp", args[1], 5*time.Second)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(connection, os.Getpid()); err != nil {
			return err
		}
		// The test opens a stable process handle before permitting launcher exit.
		if _, err := io.ReadFull(connection, make([]byte, 1)); err != nil {
			return err
		}
		owner, err := net.DialTimeout("tcp", args[2], 5*time.Second)
		if err != nil {
			return err
		}
		if _, err := owner.Write([]byte("R")); err != nil {
			return err
		}
		if err := owner.Close(); err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, connection)
		return errors.Join(err, connection.Close())
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	child := exec.Command(binary, "installer-process-fixture", "pipe-child", args[1], listener.Addr().String())
	child.Stdout, child.Stderr = os.Stdout, os.Stderr
	if err := child.Start(); err != nil {
		return err
	}
	ready, err := listener.Accept()
	if err != nil {
		return err
	}
	if _, err := io.ReadFull(ready, make([]byte, 1)); err != nil {
		return err
	}
	if err := errors.Join(ready.Close(), listener.Close()); err != nil {
		return err
	}
	if args[0] == "launcher-exit" {
		fmt.Println("selected-version")
		return nil
	}
	for {
		time.Sleep(time.Hour)
	}
}

func TestInstallWindowsInspectionOwnsLauncherTree(t *testing.T) {
	t.Setenv("AL_CAPTURE_WINDOWS_PROCESS_FIXTURE", "1")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"launcher-exit", "launcher-cancel"} {
		t.Run(mode, func(t *testing.T) {
			listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := listener.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := listener.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := installCommandOutput(ctx, binary, "installer-process-fixture", mode, listener.Addr().String())
				done <- err
			}()
			connection, err := listener.AcceptTCP()
			if err != nil {
				cancel()
				t.Fatalf("child handshake absent: %v; runner: %v", err, <-done)
			}
			defer func() {
				if err := connection.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := connection.SetDeadline(time.Now().Add(8 * time.Second)); err != nil {
				t.Fatal(err)
			}
			line, err := bufio.NewReader(connection).ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			pid, err := strconv.ParseUint(strings.TrimSpace(line), 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			process, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_TERMINATE, false, uint32(pid))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				// Also clean a failed test's deliberately blocked child.
				state, err := windows.WaitForSingleObject(process, 0)
				if err == nil && state != windows.WAIT_OBJECT_0 {
					err = windows.TerminateProcess(process, 92)
				}
				if err := errors.Join(err, windows.CloseHandle(process)); err != nil {
					t.Error(err)
				}
			}()
			if _, err := connection.Write([]byte("A")); err != nil {
				t.Fatal(err)
			}
			if mode == "launcher-cancel" {
				cancel()
			}
			err = <-done
			if mode == "launcher-exit" && !errors.Is(err, exec.ErrWaitDelay) {
				t.Fatalf("inherited pipe was not bounded independently of context: %v", err)
			}
			if mode == "launcher-cancel" && !errors.Is(err, context.Canceled) {
				t.Fatalf("caller cancellation cause lost: %v", err)
			}
			state, err := windows.WaitForSingleObject(process, 2000)
			if err != nil || state != windows.WAIT_OBJECT_0 {
				t.Fatalf("owned launcher descendant remained alive: state=%d err=%v", state, err)
			}
		})
	}
}
