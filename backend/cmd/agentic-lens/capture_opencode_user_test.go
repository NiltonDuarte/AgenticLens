package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestOpenCodeUserInstallPreservesCustomHomeAndDrift(t *testing.T) {
	project := t.TempDir()
	initCLIGitFixture(t, project)
	writeCLISharedPolicy(t, project)
	home := t.TempDir()
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("OPENCODE_CONFIG_DIR", home)
	foreign := filepath.Join(home, "plugins", "unrelated.ts")
	if err := os.Mkdir(filepath.Dir(foreign), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foreign, []byte("export const myPlugin = async () => ({})"), 0600); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(home, "opencode.jsonc")
	configBody := []byte("{ /* personal settings */ }\n")
	if err := os.WriteFile(config, configBody, 0600); err != nil {
		t.Fatal(err)
	}
	operate := func(command string, extra ...string) ([]byte, error) {
		args := []string{"capture", command, "--source", "opencode", "--repo", project, "--entrypoint", managerBinaryPath}
		return exec.Command(cliBinaryPath, append(args, extra...)...).CombinedOutput()
	}
	path := filepath.Join(home, "plugins", "agenticlens-capture.ts")
	if out, err := operate("install-user", "--dry-run"); err != nil {
		t.Fatalf("dry run: %s %v", out, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dry run created plugin: %v", err)
	}

	duplicate := filepath.Join(xdg, "opencode", "plugin", "agenticlens-capture.js")
	if err := os.MkdirAll(filepath.Dir(duplicate), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicate, []byte("user managed"), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := operate("install-user"); err == nil {
		t.Fatalf("duplicate in second loaded global directory allowed: %s", out)
	}
	if err := os.Remove(duplicate); err != nil {
		t.Fatal(err)
	}
	if out, err := operate("install-user"); err != nil {
		t.Fatalf("install: %s %v", out, err)
	}
	first := readCaptureFile(t, path)
	if out, err := operate("install-user"); err != nil || !bytes.Equal(first, readCaptureFile(t, path)) {
		t.Fatalf("idempotent install: %s %v", out, err)
	}
	edited := append(append([]byte{}, first...), []byte("\n// local edit\n")...)
	if err := os.WriteFile(path, edited, 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"install-user", "uninstall-user"} {
		if out, err := operate(command); err == nil || !bytes.Equal(edited, readCaptureFile(t, path)) {
			t.Fatalf("%s discarded user edit: %s %v", command, out, err)
		}
	}
	if err := os.WriteFile(path, first, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := operate("uninstall-user"); err != nil {
		t.Fatalf("uninstall: %s %v", out, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("owned plugin retained: %v", err)
	}
	if !bytes.Equal(configBody, readCaptureFile(t, config)) || !bytes.Contains(readCaptureFile(t, foreign), []byte("myPlugin")) {
		t.Fatal("unrelated config or plugin changed")
	}
}

func TestOpenCodeHookAdmissionUsesNativeWorktreeBeforeCaptureIO(t *testing.T) {
	base := t.TempDir()
	initCLIGitFixture(t, base)
	writeCLISharedPolicy(t, base)
	historyGit(t, base, "add", ".agenticlens-config.yaml")
	historyGit(t, base, "commit", "-m", "fixture policy")
	worktree := filepath.Join(t.TempDir(), "linked")
	historyGit(t, base, "worktree", "add", "-b", "linked", worktree)
	nested := filepath.Join(worktree, "nested")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}
	localFile := ".agenticlens-config.local.yaml"
	disabled := "agenticlens:\n  state_dir: .private\ntelemetry:\n  default_source: opencode\n  sources:\n    opencode:\n      database_path: relative-source.db\n  capture:\n    enabled: false\n    state_dir: capture-state\n    pass_timeout_ms: 1000\n    max_sessions: 32\n    max_source_bytes: 4096\n    refresh_ms: 500\n    debounce_ms: 100\n    idle_ms: 500\n    max_lifetime_ms: 2000\n    commit_warning: true\n"
	if err := os.WriteFile(filepath.Join(base, localFile), []byte(disabled), 0600); err != nil {
		t.Fatal(err)
	}
	cwd := t.TempDir()
	t.Setenv("OPENCODE_DB", filepath.Join(cwd, "missing.db"))
	call := func() ([]byte, error) {
		event, err := json.Marshal(capture.HookInput{WorkingDirectory: nested, SessionID: "observed-exact-id", Event: "session.created"})
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(managerBinaryPath, "opencode-hook", "receive")
		cmd.Dir = cwd
		cmd.Stdin = bytes.NewReader(event)
		return cmd.CombinedOutput()
	}
	if out, err := call(); err != nil {
		t.Fatalf("disabled base fallback should skip: %s %v", out, err)
	}
	if err := os.WriteFile(filepath.Join(worktree, localFile), []byte(strings.ReplaceAll(disabled, "    enabled: false\n", "")), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := call(); err == nil || !strings.Contains(string(out), "enabled") {
		t.Fatalf("selected incomplete local file inherited base switch or routed to callback cwd: %s %v", out, err)
	}
	// The base local path remains relative to the base even for a nested session in the linked checkout.
	if err := os.Remove(filepath.Join(worktree, localFile)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, localFile), []byte(strings.Replace(disabled, "enabled: false", "enabled: true", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	prepareCaptureInstall(t, base, []string{"install", "--repo", base})
	shared := readCaptureFile(t, filepath.Join(base, ".agenticlens-config.yaml"))
	if err := os.WriteFile(filepath.Join(worktree, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := call(); err == nil || !strings.Contains(string(out), filepath.Join(base, "relative-source.db")) || !strings.Contains(string(out), filepath.Join(cwd, "missing.db")) {
		t.Fatalf("source disagreement must report both configured and runtime origins before registration: %s %v", out, err)
	}
	if err := os.Remove(filepath.Join(worktree, ".agenticlens-config.yaml")); err != nil {
		t.Fatal(err)
	}
	if out, err := call(); err != nil {
		t.Fatalf("missing active shared file should skip despite invalid local: %s %v", out, err)
	}
	for _, dir := range []string{base, worktree, cwd} {
		for _, name := range []string{"capture-state", "relative-source.db", ".private", "missing.db"} {
			if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
				t.Fatalf("callback allocated %s/%s: %v", dir, name, err)
			}
		}
	}
}

func TestOpenCodeUserMigrationVerifiesLegacyPluginAcrossLinkedWorktree(t *testing.T) {
	base := t.TempDir()
	initCLIGitFixture(t, base)
	writeCLISharedPolicy(t, base)
	xdg := t.TempDir()
	t.Setenv("OPENCODE_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	state := filepath.Join(base, "local-state")
	store := filepath.Join(base, "source.db")
	out, err := captureCLI(t, base, "", "install", "--repo", base, "--source", "opencode", "--store", store, "--state-dir", state, "--binary", cliBinaryPath)
	if err != nil {
		t.Fatalf("legacy install: %s %v", out, err)
	}
	var install capture.InstallResult
	if err := json.Unmarshal(out, &install); err != nil {
		t.Fatal(err)
	}
	historyGit(t, base, "add", ".agenticlens-config.yaml")
	historyGit(t, base, "commit", "-m", "share fixture policy")
	linked := filepath.Join(t.TempDir(), "linked")
	historyGit(t, base, "worktree", "add", "-b", "linked", linked)
	run := func(extra ...string) ([]byte, error) {
		args := append([]string{"capture", "install-user", "--source", "opencode", "--repo", linked, "--entrypoint", managerBinaryPath}, extra...)
		return exec.Command(cliBinaryPath, args...).CombinedOutput()
	}
	if out, err := run(); err == nil {
		t.Fatalf("user installation duplicated active base plugin: %s", out)
	}
	plugin := filepath.Join(base, ".opencode", "plugins", "agenticlens-capture.ts")
	before := readCaptureFile(t, plugin)
	edited := append(append([]byte{}, before...), []byte("\n// local edit\n")...)
	if err := os.WriteFile(plugin, edited, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := run("--migrate-legacy-repo", base); err == nil || !bytes.Equal(edited, readCaptureFile(t, plugin)) {
		t.Fatalf("explicit migration discarded user edit: %s %v", out, err)
	}
	userPlugin := filepath.Join(xdg, "opencode", "plugins", "agenticlens-capture.ts")
	if _, err := os.Stat(userPlugin); !os.IsNotExist(err) {
		t.Fatalf("refused migration partially activated global plugin: %v", err)
	}
	if err := os.WriteFile(plugin, before, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := run("--migrate-legacy-repo", base); err != nil {
		t.Fatalf("explicit unchanged migration: %s %v", out, err)
	}
	if _, err := os.Stat(plugin); !os.IsNotExist(err) {
		t.Fatalf("migrated project plugin remains: %v", err)
	}
	for _, path := range []string{filepath.Join(base, ".opencode", "agenticlens-installation.json"), filepath.Join(install.Manifest.StateDir, "install-opencode.json")} {
		var manifest capture.Manifest
		if err := json.Unmarshal(readCaptureFile(t, path), &manifest); err != nil || manifest.Active {
			t.Fatalf("retired manifest %s: %+v %v", path, manifest, err)
		}
	}
	if len(readCaptureFile(t, userPlugin)) == 0 {
		t.Fatal("empty global plugin after migration")
	}

	custom := filepath.Join(t.TempDir(), "custom-config")
	t.Setenv("OPENCODE_CONFIG_DIR", custom)
	t.Run("worker recognizes the still-loaded user integration", func(t *testing.T) {
		installed, err := capture.UserOpenCodeInstalled()
		if err != nil || !installed {
			t.Fatalf("worker missed its migrated plugin in the loaded XDG directory: %t %v", installed, err)
		}
		if _, err := os.Stat(custom); !os.IsNotExist(err) {
			t.Fatalf("read-only worker detection created a custom config home: %v", err)
		}
	})
	t.Run("legacy install cannot reactivate a duplicate", func(t *testing.T) {
		out, err := captureCLI(t, base, "", "install", "--repo", base, "--source", "opencode", "--store", store, "--state-dir", state, "--binary", cliBinaryPath)
		if err == nil {
			t.Fatalf("reactivated a duplicate project plugin after changing config homes: %s", out)
		}
		if _, err := os.Stat(plugin); !os.IsNotExist(err) {
			t.Fatalf("duplicate legacy plugin created: %v", err)
		}
	})
}

// A global Codex installation can serve two linked worktrees while only one
// has admitted a native Codex callback. Each process reads its own worktree.
func TestOpenCodeWorktreeKeepsCustomEngineAcrossCodexAdmission(t *testing.T) {
	mixed := newCodexInboxFixture(t)
	configureInboxExecutablePath(t, mixed, "binary_dir", mixed.project)
	dev := filepath.Join(mixed.project, "selected-build", filepath.Base(fixtureCaptureEngine()))
	if err := os.MkdirAll(filepath.Dir(dev), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dev, readCaptureFile(t, cliBinaryPath), 0700); err != nil {
		t.Fatal(err)
	}
	shared := ".agenticlens-config.yaml"
	historyGit(t, mixed.project, "add", shared)
	historyGit(t, mixed.project, "commit", "-m", "fixture shared policy")
	other := mixed
	other.project = filepath.Join(t.TempDir(), "opencode-worktree")
	historyGit(t, mixed.project, "worktree", "add", "-b", "opencode", other.project)
	nested := filepath.Join(mixed.project, ".worktrees", "nested")
	historyGit(t, mixed.project, "worktree", "add", "-b", "nested", nested)
	var err error
	other.state, err = capture.ProjectState(other.stateBase, other.project)
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(os.Getenv("HOME"), "other-project-cache")
	alternate := "v9.8.8"
	otherShared := filepath.Join(other.project, shared)
	body := strings.Replace(string(readCaptureFile(t, otherShared)), "  binary_dir: "+strconv.Quote(filepath.Dir(dev)), "  install_root: "+strconv.Quote(cache), 1)
	body = strings.Replace(body, "version: v9.8.7", "version: "+alternate, 1)
	if err := os.WriteFile(otherShared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	altEngine := filepath.Join(t.TempDir(), "engine")
	if out, err := exec.Command("go", "build", "-ldflags", "-X agenticlens/internal/domain.Version="+alternate, "-o", altEngine, ".").CombinedOutput(); err != nil {
		t.Fatalf("build alternate engine: %s %v", out, err)
	}
	install := exec.Command(fixtureCaptureManager(), "--al-project", other.project, "install", alternate, "--install-root", cache, "--local-bin", altEngine, "--sha256", fmt.Sprintf("%x", sha256.Sum256(readCaptureFile(t, altEngine))))
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("install independent OpenCode engine: %s %v", out, err)
	}
	t.Setenv("OPENCODE_CONFIG_DIR", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	database := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", database)
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	setOpenCodeFixtureProject(t, database, mixed.project)
	for id, project := range map[string]string{"mixed-session": mixed.project, "independent-session": other.project} {
		if _, err := db.Exec("INSERT INTO session (id,agent,model,time_created,time_updated,directory) VALUES (?, 'root','fixture',1,2,?)", id, project); err != nil {
			t.Fatal(err)
		}
	}
	configure := func(f captureGenerationFixture, withCodex bool) {
		t.Helper()
		sources := fmt.Sprintf("    opencode:\n      database_path: %q\n", database)
		if withCodex {
			sources += fmt.Sprintf("    codex:\n      home: %q\n", f.store)
		}
		policy := "telemetry:\n  default_source: opencode\n  capture:\n  sources:\n" + sources
		if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
			t.Fatal(err)
		}
	}
	configure(mixed, true)
	configure(other, false)
	codexInboxOperation(t, mixed.project, "install-user")
	if out, err := captureCLI(t, other.project, "", "install-user", "--source", "opencode", "--repo", other.project, "--entrypoint", fixtureCaptureManager()); err != nil {
		t.Fatalf("install native OpenCode hook: %s %v", out, err)
	}
	for _, f := range []captureGenerationFixture{mixed, other} {
		s, err := capture.OpenProject(f.state, f.project)
		if err != nil {
			t.Fatal(err)
		}
		completion := captureWorkerCompletion(t, s.CoordinationDirectory())
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    enabled: false\n")); err != nil {
				t.Error(err)
			}
			captureCleanupWorkers(t, completion)
		})
	}
	hook := func(project string, event capture.HookInput, args ...string) ([]byte, error) {
		t.Helper()
		event.WorkingDirectory = project
		payload, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(fixtureCaptureManager(), args...)
		cmd.Dir = project
		cmd.Stdin = bytes.NewReader(payload)
		return cmd.CombinedOutput()
	}
	openCode := func(project, id string) ([]byte, error) {
		return hook(project, capture.HookInput{Event: "session.idle", SessionID: id}, "opencode-hook", "receive")
	}
	verifyOpenCode := func(project, id string) {
		t.Helper()
		out, err := openCode(project, id)
		var row domain.CaptureRegistration
		if err != nil || json.Unmarshal(out, &row) != nil || row.Request == "" || row.Identity.SessionID != id {
			t.Fatalf("native OpenCode callback for %s: %s %v", id, out, err)
		}
	}
	verifyOpenCode(other.project, "independent-session")
	rollout := mixed.write(t, 27)
	codexStart := capture.HookInput{Event: "SessionStart", SessionID: "root", TranscriptPath: rollout}
	registry, err := capture.OpenProject(mixed.state, mixed.project)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	lease, err := capture.AcquireWorker(registry.CoordinationDirectory())
	if err != nil || lease == nil {
		t.Fatalf("acquire active ordinary worker lease: %v", err)
	}
	ctx := context.Background()
	if err := registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	before, err := registry.PublicationRequest(ctx)
	if err != nil || before.Generation <= before.Acknowledged {
		t.Fatalf("ordinary worker has no pending generation: %+v %v", before, err)
	}
	blocked, busyErr := hook(mixed.project, codexStart, "codex-hook")
	after, readErr := registry.PublicationRequest(ctx)
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if busyErr == nil || !bytes.Contains(blocked, []byte("ordinary capture worker")) || readErr != nil || after != before {
		t.Fatalf("first admission changed pending generation or failed for the wrong reason: %s %v %+v %v", blocked, busyErr, after, readErr)
	}
	out, err := hook(mixed.project, codexStart, "codex-hook")
	var reply capture.HookOutput
	if err != nil || json.Unmarshal(out, &reply) != nil || reply.Context == nil || !strings.Contains(reply.Context.Context, fixtureCaptureEngine()) {
		t.Fatalf("Codex callback failed to admit mixed worktree: %s %v", out, err)
	}
	verifyOpenCode(other.project, "independent-session")
	verifyOpenCode(mixed.project, "mixed-session")
	configure(mixed, false)
	verifyOpenCode(mixed.project, "mixed-session")
	path := fixtureCaptureEngine()
	if err := os.Rename(path, path+".maintenance"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Rename(path+".maintenance", path); err != nil {
			t.Error(err)
		}
	}()
	for _, project := range []string{mixed.project, nested} {
		out, err = openCode(project, "mixed-session")
		if err == nil || !bytes.Contains(out, []byte("protected engine")) {
			t.Fatalf("removing the Codex source downgraded admitted worktree at %s: %s %v", project, out, err)
		}
	}
	verifyOpenCode(other.project, "independent-session")
}
