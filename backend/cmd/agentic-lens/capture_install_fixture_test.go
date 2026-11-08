package main_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// prepareCaptureInstall provisions the explicit shared release authority and
// real offline wrapper/native boundary required by installer behavior tests.
// It preserves each fixture's source/state policy and negative binary argument.
func prepareCaptureInstall(t *testing.T, cwd string, args []string) []string {
	t.Helper()
	if len(args) == 0 || args[0] != "install" {
		return args
	}
	prepareCaptureHarnessFixture(t)
	project, binary := cwd, cliBinaryPath
	for i := 0; i+1 < len(args); i++ {
		switch args[i] {
		case "--repo":
			project = args[i+1]
		case "--binary":
			binary = args[i+1]
		}
	}
	root, err := adaptersconfig.ProjectRoot(project)
	if err != nil {
		t.Fatal(err)
	}
	shared := filepath.Join(root, ".agenticlens-config.yaml")
	body, err := os.ReadFile(shared)
	if os.IsNotExist(err) {
		body = []byte("project_name: Capture installer fixture\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: workspaces\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	} else if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("agenticlens:\n")) && !bytes.Contains(body, []byte("\nagenticlens:\n")) {
		installRoot := filepath.Join(t.TempDir(), "native cache")
		header := fmt.Sprintf("agenticlens:\n  version: %s\n  install_root: %q\n", domain.DefaultVersion, installRoot)
		body = append([]byte(header), body...)
		if err = os.WriteFile(shared, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	entry, err := adaptersconfig.ResolveEntrypoint(root)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := entry.NativePath()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(selected), 0700); err != nil {
		t.Fatal(err)
	}
	// A symlink keeps the native executable identity exact without copying a
	// full test binary per installer invocation. Runtime code still validates it.
	if old, err := os.Readlink(selected); err == nil && old != binary {
		if err = os.Remove(selected); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = os.Lstat(selected); os.IsNotExist(err) {
		if err = os.Symlink(binary, selected); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(entry.InstallRoot, "wrapper __BINARY__ __STATE__ __STORE__")
	source := filepath.Join("..", "..", "..", "scripts", "agentic-lens-wrapper.sh")
	wrapperBody, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(wrapper, wrapperBody, 0700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--wrapper" {
			return args
		}
	}
	return append(append([]string{}, args...), "--wrapper", wrapper)
}

func fixtureWrapperFromHook(t *testing.T, project string) string {
	t.Helper()
	entry, err := adaptersconfig.ResolveEntrypoint(project)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(entry.InstallRoot, "wrapper __BINARY__ __STATE__ __STORE__")
}
