package capture

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	if os.Getenv("AL_CAPTURE_VALIDATION_HELPER") != "1" || len(os.Args) != 3 || os.Args[1] != "native-validation" {
		return
	}
	switch os.Args[2] {
	case "matching":
		fmt.Println("selected-version")
	case "mismatch":
		fmt.Println("different-version")
	case "failure":
		fmt.Fprintln(os.Stderr, "fixture execution failed")
		os.Exit(19)
	default:
		os.Exit(20)
	}
	os.Exit(0)
}

func TestInstallNativeValidationDistinguishesExecutionAndIdentity(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("AL_CAPTURE_VALIDATION_HELPER", "1")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	for _, name := range []string{"matching", "mismatch", "failure", "missing", "canceled", "deadline"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			executable, mode := binary, name
			switch name {
			case "missing":
				executable = filepath.Join(t.TempDir(), "missing-native")
			case "canceled":
				cancel()
			case "deadline":
				var expired context.CancelFunc
				ctx, expired = context.WithDeadline(context.Background(), time.Unix(0, 0))
				defer expired()
			}
			err := validateInstallNativeOutput(ctx, executable, "selected-version", "native version", "native-validation", mode)
			if name == "matching" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("failed native validation accepted")
			}
			switch name {
			case "mismatch":
				if !strings.Contains(err.Error(), `mismatch: observed "different-version", expected "selected-version"`) {
					t.Fatalf("observed identity mismatch lost: %v", err)
				}
			case "failure":
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 19 || !strings.Contains(err.Error(), "execution failed") {
					t.Fatalf("native execution failure lost: %v", err)
				}
			case "missing":
				if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "execution failed") {
					t.Fatalf("missing executable misreported as mismatch: %v", err)
				}
			case "canceled":
				if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "canceled") {
					t.Fatalf("cancellation cause lost: %v", err)
				}
			case "deadline":
				if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "timed out") {
					t.Fatalf("deadline cause lost: %v", err)
				}
			}
		})
	}
}
