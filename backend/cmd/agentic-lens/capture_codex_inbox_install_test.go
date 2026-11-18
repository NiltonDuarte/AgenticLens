package main_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"github.com/pelletier/go-toml/v2"
)

// The native workspace-write profile grants access to the ordinary system
// temp directory. Keep this fixture's isolated operator home outside that root.
func newCaptureOperatorHome(t *testing.T, project string) string {
	t.Helper()
	root, err := os.MkdirTemp("/var/tmp", "agenticlens-protected-test-")
	if err != nil {
		t.Fatalf("protected fixture requires writable /var/tmp outside the sandbox's temp grant: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	home, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	manager := fixtureCaptureManager()
	digest := func(path string) string { return fmt.Sprintf("%x", sha256.Sum256(readCaptureFile(t, path))) }
	for _, args := range [][]string{
		{"self-update", "--version", "v9.8.7", "--local-bin", managerBinaryPath, "--sha256", digest(managerBinaryPath), "--dest", manager},
		{"--al-project", project, "install", "v9.8.7", "--install-root", filepath.Join(home, ".agenticlens"), "--local-bin", cliBinaryPath, "--sha256", digest(cliBinaryPath)},
	} {
		cmd := exec.Command(managerBinaryPath, args...)
		cmd.Dir = project
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("operator provisioning failed: %s %v", out, err)
		}
	}
	return manager
}

func fixtureCaptureManager() string {
	return filepath.Join(os.Getenv("HOME"), ".local", "bin", "agentic-lens")
}
func fixtureCaptureEngine() string {
	host := runtime.GOOS + "-" + runtime.GOARCH
	return filepath.Join(os.Getenv("HOME"), ".agenticlens", "bin", "v9.8.7", host, "agentic-lens-engine-"+host)
}

// User-hook installation resolves an explicit engine pin in addition to the
// generation fixture's configured ledger roots and separate private state.
func newCodexInboxFixture(t *testing.T) captureGenerationFixture {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	path := filepath.Join(f.project, ".agenticlens-config.yaml")
	if err := os.WriteFile(path, append([]byte("agenticlens:\n  version: v9.8.7\n"), readCaptureFile(t, path)...), 0600); err != nil {
		t.Fatal(err)
	}
	newCaptureOperatorHome(t, f.project)
	f.store = filepath.Join(os.Getenv("HOME"), ".codex")
	if err := os.MkdirAll(filepath.Join(f.store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", f.store)
	f.stateBase = filepath.Join(os.Getenv("HOME"), "capture-state")
	f.publication = filepath.Join(os.Getenv("HOME"), "publication-state")
	var err error
	f.state, err = capture.ProjectState(f.stateBase, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n")); err != nil {
		t.Fatal(err)
	}
	return f
}

func codexInboxInstall(t *testing.T, project, command string, extra ...string) ([]byte, error) {
	t.Helper()
	args := []string{"capture", command, "--user-source", "codex", "--repo", project, "--entrypoint", fixtureCaptureManager()}
	return exec.Command(cliBinaryPath, append(args, extra...)...).CombinedOutput()
}

func codexInboxOperation(t *testing.T, project, command string, extra ...string) capture.InstallResult {
	t.Helper()
	out, err := codexInboxInstall(t, project, command, extra...)
	if err != nil {
		t.Fatalf("%s: %s %v", command, out, err)
	}
	var result capture.InstallResult
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCodexInboxGrantPreservesUserSettingsAndSurvivesReinstall(t *testing.T) {
	f := newCodexInboxFixture(t)
	configPath := filepath.Join(f.store, "config.toml")
	inbox := filepath.Join(f.store, "agenticlens-binding-inbox")
	before := []byte("# codex settings\nsandbox_mode = 'workspace-write'\n[sandbox_workspace_write]\nnetwork_access = false\nwritable_roots = [\n  '/keep', # user grant\n]\n[hooks.state.'foreign']\ntrusted_hash = 'keep-me'\n")
	if err := os.WriteFile(configPath, before, 0600); err != nil {
		t.Fatal(err)
	}
	preview := codexInboxOperation(t, f.project, "install-user", "--dry-run")
	var changes []string
	for _, change := range preview.Changes {
		changes = append(changes, change.Path)
	}
	if !preview.DryRun || !slices.Contains(changes, inbox) || !slices.Contains(changes, configPath) {
		t.Fatalf("preview omitted inbox or grant: %+v", preview)
	}
	if _, err := os.Lstat(inbox); !os.IsNotExist(err) || !bytes.Equal(before, readCaptureFile(t, configPath)) {
		t.Fatalf("preview altered the user home: %v", err)
	}
	codexInboxOperation(t, f.project, "install-user")
	installed := readCaptureFile(t, configPath)
	var parsed struct {
		Mode    string `toml:"sandbox_mode"`
		Sandbox struct {
			Roots []string `toml:"writable_roots"`
		} `toml:"sandbox_workspace_write"`
	}
	if err := toml.Unmarshal(installed, &parsed); err != nil || parsed.Mode != "workspace-write" || (len(parsed.Sandbox.Roots) != 2 || !slices.Contains(parsed.Sandbox.Roots, "/keep") || !slices.Contains(parsed.Sandbox.Roots, inbox)) {
		t.Fatalf("unexpected sandbox configuration: %+v, %v", parsed, err)
	}
	for _, unchanged := range []string{"# codex settings", "# user grant", "[hooks.state.'foreign']\ntrusted_hash = 'keep-me'", "network_access = false"} {
		if !bytes.Contains(installed, []byte(unchanged)) {
			t.Fatalf("lost user config %q", unchanged)
		}
	}
	if repeated := codexInboxOperation(t, f.project, "install-user"); len(repeated.Changes) != 0 || !bytes.Equal(installed, readCaptureFile(t, configPath)) {
		t.Fatalf("reinstall wrote unchanged settings: %+v", repeated)
	}
	userEdits := []byte("\n[unrelated]\nsetting = 'keep this too'\n")
	if err := os.WriteFile(configPath, append(installed, userEdits...), 0600); err != nil {
		t.Fatal(err)
	}
	otherFile := filepath.Join(inbox, "other-entry")
	if err := os.WriteFile(otherFile, []byte("retain"), 0600); err != nil {
		t.Fatal(err)
	}
	codexInboxOperation(t, f.project, "uninstall-user")
	after := readCaptureFile(t, configPath)
	parsed.Sandbox.Roots = nil
	if err := toml.Unmarshal(after, &parsed); err != nil || !slices.Equal(parsed.Sandbox.Roots, []string{"/keep"}) || !bytes.Contains(after, userEdits) || !bytes.Equal(readCaptureFile(t, otherFile), []byte("retain")) {
		t.Fatalf("uninstall lost user edits or retained installer grant: %s (%v)", after, err)
	}
}

func TestCodexInboxPreexistingGrantAndConflictPreflight(t *testing.T) {
	for _, kind := range []string{"preexisting", "invalid TOML", "symlink inbox", "symlink config", "overlapping capture state", "overlapping publication state"} {
		t.Run(kind, func(t *testing.T) {
			f := newCodexInboxFixture(t)
			configPath := filepath.Join(f.store, "config.toml")
			inbox := filepath.Join(f.store, "agenticlens-binding-inbox")
			original := []byte(fmt.Sprintf("sandbox_workspace_write.writable_roots = [%q] # already granted\n", inbox))
			if kind == "invalid TOML" {
				original = []byte("sandbox_workspace_write.writable_roots = 'unexpected scalar'\n")
			}
			if kind == "overlapping capture state" || kind == "overlapping publication state" {
				privatePath := f.stateBase
				if kind == "overlapping publication state" {
					privatePath = f.publication
				}
				local := filepath.Join(f.project, ".agenticlens-config.local.yaml")
				current := readCaptureFile(t, local)
				updated := bytes.Replace(current, []byte(privatePath), []byte(filepath.Join(inbox, "worker-files")), 1)
				if bytes.Equal(current, updated) {
					t.Fatal("fixture did not overlap its private state")
				}
				if err := os.WriteFile(local, updated, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "symlink inbox" {
				if err := os.Symlink(t.TempDir(), inbox); err != nil {
					t.Fatal(err)
				}
			}
			target := configPath
			if kind == "symlink config" {
				target = filepath.Join(t.TempDir(), "personal.toml")
				if err := os.Symlink(target, configPath); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(target, original, 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "preexisting" {
				codexInboxOperation(t, f.project, "install-user")
				codexInboxOperation(t, f.project, "uninstall-user")
			} else {
				if out, err := codexInboxInstall(t, f.project, "install-user"); err == nil {
					t.Fatalf("accepted conflicting installer input: %s", out)
				}
				for _, name := range []string{"hooks.json", "agenticlens-user-installation.json"} {
					if _, err := os.Stat(filepath.Join(f.store, name)); !os.IsNotExist(err) {
						t.Fatalf("%s changed after rejected preflight: %v", name, err)
					}
				}
			}
			if !bytes.Equal(original, readCaptureFile(t, target)) {
				t.Fatal("installer modified an existing user grant or conflicting configuration")
			}
		})
	}
}

func TestCodexInboxProtectsExecutableSelectionAtInstall(t *testing.T) {
	for _, setting := range []string{"install_root", "binary_dir", "manager"} {
		t.Run(setting, func(t *testing.T) {
			f := newCodexInboxFixture(t)
			inbox := filepath.Join(f.store, "agenticlens-binding-inbox")
			manager := fixtureCaptureManager()
			if setting == "manager" {
				if err := os.Mkdir(inbox, 0700); err != nil {
					t.Fatal(err)
				}
				manager = filepath.Join(inbox, "manager")
				// Only an inert placeholder is needed to check path preflight.
				if err := os.WriteFile(manager, []byte("ordinary file"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				configureInboxExecutablePath(t, f, setting, inbox)
			}
			cmd := exec.Command(cliBinaryPath, "capture", "install-user", "--user-source", "codex", "--repo", f.project, "--entrypoint", manager, "--dry-run")
			out, err := cmd.CombinedOutput()
			if setting == "manager" {
				if err == nil || !strings.Contains(strings.ToLower(string(out)), "inbox") {
					t.Fatalf("manager must be outside the writable inbox: %s %v", out, err)
				}
			} else if err != nil {
				t.Fatalf("project %s should not select the trusted callback engine: %s %v", setting, out, err)
			}
			for _, name := range []string{"hooks.json", "config.toml", "agenticlens-user-installation.json"} {
				if _, err := os.Stat(filepath.Join(f.store, name)); !os.IsNotExist(err) {
					t.Fatalf("failed preflight created %s: %v", name, err)
				}
			}
		})
	}
}

func configureInboxExecutablePath(t *testing.T, f captureGenerationFixture, setting, inbox string) {
	t.Helper()
	path := filepath.Join(f.project, ".agenticlens-config.yaml")
	body := readCaptureFile(t, path)
	key := []byte("agenticlens:\n")
	newBody := bytes.Replace(body, key, append(key, []byte(fmt.Sprintf("  %s: %q\n", setting, filepath.Join(inbox, "selected-build")))...), 1)
	if bytes.Equal(body, newBody) {
		t.Fatal("missing fixture executable configuration")
	}
	if err := os.WriteFile(path, newBody, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCodexHookPreservesRuntimeSourceCheckAcrossProjects(t *testing.T) {
	installed := newCodexInboxFixture(t)
	manager := fixtureCaptureManager()
	home := os.Getenv("HOME")
	codexInboxOperation(t, installed.project, "install-user")
	project := newCodexInboxFixture(t)
	t.Setenv("HOME", home)
	t.Setenv("CODEX_HOME", installed.store)
	configureInboxExecutablePath(t, project, "binary_dir", filepath.Join(installed.store, "agenticlens-binding-inbox"))
	input, err := json.Marshal(capture.HookInput{SessionID: "new-root", Event: "SessionStart", WorkingDirectory: project.project})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(manager, "codex-hook")
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "runtime source") || !strings.Contains(string(out), project.store) || !strings.Contains(string(out), installed.store) {
		t.Fatalf("expected the protected handler to diagnose both source homes: %s %v", out, err)
	}
}
