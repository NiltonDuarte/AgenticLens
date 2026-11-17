package main_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
)

// Re-executed disposable harnesses handle only the optional inspection protocol.
// The installed AgenticLens executable and wrapper remain real boundaries.
func init() {
	source := filepath.Base(os.Args[0])
	if source != "codex" && source != "opencode" {
		return
	}
	if external := os.Getenv("AL_CAPTURE_EXTERNAL_HARNESS_DIR"); external != "" {
		executable, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(91)
		}
		currentDir, currentErr := os.Stat(filepath.Dir(executable))
		externalDir, externalErr := os.Stat(external)
		if currentErr != nil || externalErr != nil {
			fmt.Fprintln(os.Stderr, currentErr, externalErr)
			os.Exit(91)
		}
		if os.SameFile(currentDir, externalDir) {
			if err := os.WriteFile(filepath.Join(external, "invoked"), []byte(source), 0600); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(91)
			}
			os.Exit(42)
		}
	}
	if os.Getenv("AL_CAPTURE_HARNESS_FIXTURE") != "1" {
		return
	}
	args := strings.Join(os.Args[1:], " ")
	switch {
	case source == "opencode" && args == "--version":
		fmt.Println("1.19.5")
	case source == "codex" && args == "--version":
		fmt.Println("codex-cli 0.166.0")
	case source == "codex" && args == "features list":
		fmt.Println("hooks enabled")
	default:
		fmt.Fprintln(os.Stderr, "unsupported disposable harness inspection", source, args)
		os.Exit(2)
	}
	os.Exit(0)
}

func writeCaptureHarnessExecutable(t *testing.T, dir, source string) string {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := source
	target := filepath.Join(dir, name)
	if err := os.Link(binary, target); err != nil {
		body, readErr := os.ReadFile(binary)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if writeErr := os.WriteFile(target, body, 0700); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	return target
}

func prepareCaptureHarnessFixture(t *testing.T) {
	t.Helper()
	// A child dispatch flag inherited from the runner does not establish fixture
	// ownership. Each installer setup provisions its own executables and PATH.
	dir := t.TempDir()
	for _, source := range []string{"codex", "opencode"} {
		writeCaptureHarnessExecutable(t, dir, source)
	}
	t.Setenv("AL_CAPTURE_HARNESS_FIXTURE", "1")
	t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestCaptureHarnessFixtureOwnsInheritedDispatchGuard(t *testing.T) {
	for _, inherited := range []string{"0", "1"} {
		for _, source := range []string{"codex", "opencode"} {
			t.Run("guard="+inherited+"/"+source, func(t *testing.T) {
				external := t.TempDir()
				harness := writeCaptureHarnessExecutable(t, external, source)
				marker := filepath.Join(external, "invoked")
				t.Setenv("AL_CAPTURE_EXTERNAL_HARNESS_DIR", external)
				t.Setenv("AL_CAPTURE_HARNESS_FIXTURE", inherited)
				t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
				t.Setenv("PATH", external+string(os.PathListSeparator)+os.Getenv("PATH"))
				// Positive control proves that the inherited PATH executable records
				// invocation, even though optional inspection failure permits install.
				if out, err := exec.Command(harness, "--version").CombinedOutput(); err == nil {
					t.Fatalf("hostile external harness control unexpectedly succeeded: %s", out)
				}
				if _, err := os.Stat(marker); err != nil {
					t.Fatalf("external harness control did not record invocation: %v", err)
				}
				if err := os.Remove(marker); err != nil {
					t.Fatal(err)
				}
				f := newCaptureGenerationFixture(t)
				store := f.store
				if source == "opencode" {
					store = filepath.Join(store, "source.db")
				}
				out, err := captureCLI(t, f.project, "", "install", "--repo", f.project, "--state-dir", f.state, "--source", source, "--store", store)
				if err != nil {
					t.Fatalf("installation with inherited dispatch flag failed: %v %s", err, out)
				}
				if _, err := os.Stat(marker); !os.IsNotExist(err) {
					t.Fatalf("inherited dispatch guard bypassed disposable harness provisioning: %v", err)
				}
				var installed capture.InstallResult
				if err := json.Unmarshal(out, &installed); err != nil {
					t.Fatalf("installation result: %v %s", err, out)
				}
				want := "codex-cli 0.166.0"
				if source == "opencode" {
					want = "1.19.5"
				}
				if installed.Manifest.HarnessVersion != want {
					t.Fatalf("test-owned harness probe unavailable: %+v", installed.Manifest)
				}
			})
		}
	}
}
