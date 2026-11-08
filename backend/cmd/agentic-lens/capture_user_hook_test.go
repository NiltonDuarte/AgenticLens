package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	config "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestNativeUserHookDisabledSessionUsesContextAndAvoidsCaptureIO(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	if err := os.WriteFile(shared, append([]byte("agenticlens:\n  version: v9.8.7\n  install_root: "+filepath.Join(f.project, "absent-cache")+"\n"), readCaptureFile(t, shared)...), 0600); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(f.project, "unavailable-source")
	state := filepath.Join(f.project, "uncreated-state")
	local := "telemetry:\n  capture:\n    enabled: false\n    state_dir: " + state + "\n  sources:\n    codex:\n      home: " + source + "\n"
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(capture.HookInput{SessionID: "exact-root", Event: "SessionStart", WorkingDirectory: f.project, TranscriptPath: filepath.Join(source, "sessions", "not-flushed.jsonl")})
	if err != nil {
		t.Fatal(err)
	}
	for _, invocation := range [][]string{{cliBinaryPath, "capture", "user-hook"}, {managerBinaryPath, "codex-hook"}} {
		command := exec.Command(invocation[0], invocation[1:]...)
		command.Dir = t.TempDir()
		command.Stdin = strings.NewReader(string(payload))
		command.Env = append(os.Environ(), "CODEX_HOME="+filepath.Join(f.project, "different-runtime-home"))
		out, err := command.CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) != "{}" {
			t.Fatalf("disabled callback %v touched source/readiness: %s %v", invocation, out, err)
		}
	}
	for _, path := range []string{source, state, filepath.Join(f.project, "absent-cache")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("disabled callback created %s: %v", path, err)
		}
	}
}

func TestNativeUserBindingUsesSelectedHandlerDespiteStaleWrapper(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	t.Setenv("PATH", t.TempDir())
	path := f.write(t, 17)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	scope := store.CoordinationDirectory()
	stale := capture.Manifest{Protocol: capture.ManifestProtocolVersion, Repository: f.project, StateDir: scope, Source: "codex", SourceStore: f.store, Wrapper: filepath.Join(f.project, "missing-historical-wrapper"), Active: true}
	encoded, err := json.Marshal(stale)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scope, "install-codex.json"), encoded, 0600); err != nil {
		t.Fatal(err)
	}
	completion := captureWorkerCompletion(t, scope)
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"receive", "--user-level", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state}
	out, err := captureCLI(t, f.project, string(payload), args...)
	if err != nil {
		t.Fatalf("user callback: %s %v", out, err)
	}
	var response capture.HookOutput
	if err := json.Unmarshal(out, &response); err != nil {
		t.Fatal(err)
	}
	if response.Context == nil || !strings.Contains(response.Context.Context, cliBinaryPath) || strings.Contains(response.Context.Context, "missing-historical-wrapper") {
		t.Fatalf("binding prompt did not use admitted native: %s", out)
	}
	rows, err := store.List(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("exact native registration missing: %+v %v", rows, err)
	}
	out, err = captureCLI(t, f.project, "", "bind", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--session-id", "root", "--request", rows[0].Request, "--ticket-id", "TEST-100", "--ticket-title", "User native binding", "--ticket-type", "feat")
	if err != nil {
		t.Fatalf("selected native binding command failed: %s %v", out, err)
	}
	payload, err = json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionEnd", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, string(payload), args...); err != nil {
		t.Fatalf("end native fixture: %s %v", out, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := completion(ctx); err != nil {
		t.Fatal(err)
	}
	rows, err = store.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].TicketID != "TEST-100" {
		t.Fatalf("exact response not bound: %+v %v", rows, err)
	}
}

func TestBaseLocalSetupRequiresChoiceAndPreservesExistingValues(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(cliBinaryPath, "init", "--local-only", "--source", "codex", "--codex-home", f.store, "--state-dir", f.stateBase)
	command.Dir = f.project
	if out, err := command.CombinedOutput(); err == nil || !strings.Contains(string(out), "capture.enabled") {
		t.Fatalf("noninteractive setup guessed enablement: %s %v", out, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("missing choice mutated local: %v", err)
	}
	original := "telemetry:\n  capture:\n    enabled: false\n    max_sessions: 7\n  default_source: codex\n  sources:\n    codex:\n      home: " + f.store + "\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	shared := readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))
	command = exec.Command(cliBinaryPath, "init", "--agenticlens-version", "v9.8.7", "--project-name", "Fixture", "--src-folders", "src", "--ticket-prefixes", "TEST", "--capture-enabled", "true", "--source", "codex", "--codex-home", f.store, "--state-dir", f.stateBase)
	command.Dir = f.project
	if out, err := command.CombinedOutput(); err == nil || !strings.Contains(string(out), "--local-only") {
		t.Fatalf("ordinary init silently left missing local choices: %s %v", out, err)
	}
	if string(readCaptureFile(t, path)) != original || string(readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))) != string(shared) {
		t.Fatal("rejected init mutated existing configuration")
	}
	command = exec.Command(cliBinaryPath, "init", "--local-only", "--capture-enabled", "true", "--source", "codex", "--codex-home", f.store, "--state-dir", f.stateBase)
	command.Dir = f.project
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fill missing explicit base choice: %s %v", out, err)
	}
	policy, err := config.ResolveCapture(f.project, "")
	if err != nil || policy.Enabled || policy.MaxSessions != 7 || policy.StateDir != f.stateBase {
		t.Fatalf("setup reset retained local values: %+v %v", policy, err)
	}
}

func TestNativeStatusShowsObservedStagesWithoutProjectReceipt(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	args := []string{"status", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store}
	if out, err := captureCLI(t, f.project, "", args...); err != nil {
		t.Fatalf("preevent native status: %s %v", out, err)
	}
	if _, err := os.Stat(filepath.Join(f.stateBase, "projects")); !os.IsNotExist(err) {
		t.Fatalf("preevent status allocated capture scope: %v", err)
	}
	f.bind(t, f.write(t, 17))
	f.generate(t)
	out, err := captureCLI(t, f.project, "", args...)
	if err != nil {
		t.Fatalf("native observed status: %s %v", out, err)
	}
	var status struct {
		Verification domain.CaptureHealth `json:"verification"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		t.Fatal(err)
	}
	h := status.Verification
	if h.State != "verified" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
		t.Fatalf("manifestless observed exchange hidden: %+v %s", h, out)
	}
	if _, err := os.Stat(filepath.Join(f.coordination(t), "install-codex.json")); !os.IsNotExist(err) {
		t.Fatalf("status required/created project receipt: %v", err)
	}
}

func TestNativeSubagentStartUsesProvenChildTranscriptWithoutRelabelingParent(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	rootPath := f.write(t, 17)
	f.bind(t, rootPath)
	local := "telemetry:\n  capture:\n  sources:\n    codex:\n      home: '" + filepath.ToSlash(f.store) + "'\n"
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	// Provision only the selected executable; user adoption has no project manifest.
	prepareCaptureInstall(t, f.project, []string{"install", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state})
	childPath := filepath.Join(f.store, "sessions", "child.jsonl")
	_, tail, found := strings.Cut(string(readCaptureFile(t, rootPath)), "\n")
	if !found {
		t.Fatal("root fixture metadata missing")
	}
	childMetadata := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"child","session_id":"root","cwd":%q,"timestamp":"2026-01-01T00:00:00Z","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root"}}}}}`, f.project)
	childBody := childMetadata + "\n" + strings.ReplaceAll(strings.ReplaceAll(tail, `"thread_id":"root"`, `"thread_id":"child"`), `"input_tokens":17`, `"input_tokens":29`)
	if err := os.WriteFile(childPath, []byte(childBody), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	lease, err := capture.AcquireWorker(store.CoordinationDirectory())
	if err != nil || lease == nil {
		t.Fatalf("fixture publisher serialization: %v", err)
	}
	defer lease.Close()
	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SubagentStart", AgentID: "child", TranscriptPath: childPath, WorkingDirectory: f.project})
	if err != nil {
		t.Fatal(err)
	}
	invoke := func() ([]byte, error) {
		t.Helper()
		command := exec.Command(managerBinaryPath, "codex-hook")
		command.Dir = t.TempDir()
		command.Stdin = strings.NewReader(string(payload))
		return command.CombinedOutput()
	}
	if out, err := invoke(); err != nil {
		t.Fatalf("source-proven native child callback rejected: %s %v", out, err)
	}
	assertParent := func() {
		t.Helper()
		rows, err := store.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if row.Identity.SessionID == "root" {
				if row.TranscriptPath != rootPath || row.ParentID != "" || row.TicketID != "TEST-100" {
					t.Fatalf("child callback poisoned/reassigned parent: %+v", row)
				}
				return
			}
		}
		t.Fatal("callback parent identity lost")
	}
	assertParent()
	result, err := store.Reconcile(context.Background(), "TEST-100", "tester", capture.DefaultSessionBudget, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	seenRoot, seenChild := false, false
	for _, session := range result.Sessions {
		switch session.ID {
		case "root":
			seenRoot = true
			if session.ParentID != nil || len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != 17 {
				t.Fatalf("parent metrics relabeled: %+v", session)
			}
		case "child":
			seenChild = true
			if session.ParentID == nil || *session.ParentID != "root" || len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != 29 {
				t.Fatalf("child native identity/metrics lost: %+v", session)
			}
		}
	}
	if !seenRoot || !seenChild {
		t.Fatalf("source-proven ancestry did not retain both owners: %+v", result.Sessions)
	}
	unrelated := strings.Replace(childBody, `"parent_thread_id":"root"`, `"parent_thread_id":"unrelated-root"`, 1)
	if err := os.WriteFile(childPath, []byte(unrelated), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := invoke(); err == nil || !strings.Contains(string(out), "mismatch") {
		t.Fatalf("unrelated source ancestry accepted: %s %v", out, err)
	}
	assertParent()
	foreignCwd := t.TempDir()
	foreignBody := strings.Replace(childBody, fmt.Sprintf("%q", f.project), fmt.Sprintf("%q", foreignCwd), 1)
	if err := os.WriteFile(childPath, []byte(foreignBody), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := invoke(); err == nil {
		t.Fatalf("unrelated child source worktree accepted: %s", out)
	}
	assertParent()
	if err := os.WriteFile(childPath, []byte(`{"type":"session_meta","payload":`), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := invoke(); err != nil {
		t.Fatalf("unflushed child metadata became false contradiction: %s %v", out, err)
	}
	assertParent()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := completion(ctx); err != nil {
		t.Fatal(err)
	}
}
