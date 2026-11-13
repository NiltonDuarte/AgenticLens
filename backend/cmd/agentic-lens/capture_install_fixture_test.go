package main_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
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
		body = []byte("agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: Capture installer fixture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: \"TEST\"\n      description: Fixture ticket\nworkspaces:\n  artifacts: []\n  roots:\n    - path: workspaces\n      discovery:\n        max_depth: 8\n        max_directories: 300\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	} else if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(body, []byte("agenticlens:\n")) && !bytes.Contains(body, []byte("\nagenticlens:\n")) {
		body = append([]byte("agenticlens:\n  version: v9.8.7\n"), body...)
	}
	if !bytes.Contains(body, []byte("  install_root:")) {
		body = bytes.Replace(body, []byte("agenticlens:\n"), []byte(fmt.Sprintf("agenticlens:\n  install_root: %q\n", filepath.Join(t.TempDir(), "native cache"))), 1)
	}
	if err = os.WriteFile(shared, body, 0600); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(root, ".agenticlens-config.local.yaml")
	if _, err := os.Lstat(local); os.IsNotExist(err) {
		state := ".agenticlens-capture"
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "--state-dir" {
				state = args[i+1]
			}
		}
		policy := fmt.Sprintf("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: true\n    state_dir: %q\n", state)
		if err := os.WriteFile(local, []byte(policy), 0600); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	entry, err := adaptersconfig.ResolveEntrypoint(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(entry.InstallRoot, os.TempDir()+string(os.PathSeparator)) {
		t.Fatalf("fixture installation root is not disposable: %s", entry.InstallRoot)
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
	wrapperBody, err := os.ReadFile(managerBinaryPath)
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
