package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCaptureSharedDatabasePublishesSameTicketToExactWorktrees(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.Mkdir(repo, 0700); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "state")
	source := filepath.Join(root, "codex")
	if err := os.MkdirAll(filepath.Join(source, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	git("init", "-qb", "main")
	cfg := "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: SharedCLI\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: \"TEST\"\n      description: Fixture ticket\nworkspaces:\n  artifacts: []\n  roots:\n    - path: records\n      discovery:\n        max_depth: 8\n        max_directories: 300\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(repo, ".agenticlens-config.yaml"), []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeCaptureFixtureLocal(t, captureGenerationFixture{project: repo, state: base, store: source}, []byte("telemetry:\n  capture:\n    enabled: true\n")); err != nil {
		t.Fatal(err)
	}
	git("add", ".agenticlens-config.yaml")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
	git("branch", "fixture-base")
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	git("worktree", "add", "-qb", "a", a)
	git("worktree", "add", "-qb", "b", b)
	registry, err := capture.ProjectState(base, a)
	if err != nil {
		t.Fatal(err)
	}
	other, err := capture.ProjectState(base, b)
	if err != nil {
		t.Fatal(err)
	}
	if registry != other {
		t.Fatal("worktrees did not share configured-name registry")
	}
	for i, project := range []string{a, b} {
		if err := writeCaptureFixtureLocal(t, captureGenerationFixture{project: project, state: base, store: source}, []byte("telemetry:\n  capture:\n")); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(project, "records", "TEST-100"), 0700); err != nil {
			t.Fatal(err)
		}
		id := fmt.Sprintf("root-%d", i)
		path := filepath.Join(source, "sessions", id+".jsonl")
		data := fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q,\"timestamp\":\"2026-01-01T00:00:00Z\"}}\n{\"timestamp\":\"2026-01-01T00:00:01Z\",\"type\":\"turn_context\",\"payload\":{\"model\":\"source-model\",\"turn_id\":\"one\"}}\n{\"timestamp\":\"2026-01-01T00:00:02Z\",\"type\":\"token_usage_record\",\"payload\":{\"thread_id\":%q,\"turn_id\":\"one\",\"usage\":{\"input_tokens\":%d,\"output_tokens\":2}}}\n", id, project, id, 11+i*20)
		if err = os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		store, e := capture.OpenProject(registry, project)
		if e != nil {
			t.Fatal(e)
		}
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: source, SessionID: id}
		row, e := store.Register(context.Background(), identity, path, "SessionStart")
		if e == nil {
			_, e = store.Bind(context.Background(), identity, "TEST-100", row.Request)
		}
		e2 := store.Close()
		if e != nil || e2 != nil {
			t.Fatalf("bind %s: %v %v", id, e, e2)
		}
	}
	var group sync.WaitGroup
	errors := make(chan string, 2)
	for _, project := range []string{a, b} {
		group.Add(1)
		go func(project string) {
			defer group.Done()
			output := filepath.Join(project, "records", "TEST-100", "agentic_metrics.json")
			summary := filepath.Join(project, "records", "TEST-100", "SUMMARY.md")
			cmd := exec.Command(cliBinaryPath, "generate", "--capture-auto", "--source", "codex", "--codex-home", source, "--capture-dir", registry, "--ticket-id", "TEST-100", "--ticket-title", "Shared ticket", "--ticket-type", "feat", "--json-output", output, "--append", summary)
			cmd.Dir = project
			if out, e := cmd.CombinedOutput(); e != nil {
				errors <- fmt.Sprintf("%s: %v %s", project, e, out)
			}
		}(project)
	}
	group.Wait()
	close(errors)
	for e := range errors {
		t.Fatal(e)
	}
	for i, project := range []string{a, b} {
		path := filepath.Join(project, "records", "TEST-100", "agentic_metrics.json")
		data, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var ledger domain.AgenticFactLedger
		if e = json.Unmarshal(data, &ledger); e != nil {
			t.Fatal(e)
		}
		if len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != fmt.Sprintf("root-%d", i) {
			t.Fatalf("cross-worktree publication: %+v", ledger.Sessions)
		}
		total := int64(0)
		for _, inv := range ledger.Sessions[0].Invocations {
			total += inv.Tokens.Input
		}
		if total != int64(11+i*20) {
			t.Fatalf("worktree tokens mixed: %d", total)
		}
		store, e := capture.OpenReadOnlyProject(registry, project)
		if e != nil {
			t.Fatal(e)
		}
		ds, e := store.Diagnostics(context.Background())
		if e2 := store.Close(); e == nil {
			e = e2
		}
		if e != nil {
			t.Fatal(e)
		}
		for _, d := range ds {
			if d.Phase == "publication" && d.Destination != "" && d.Destination != path {
				t.Fatalf("foreign publication diagnostic: %+v", d)
			}
		}
	}
}

func TestCaptureConfigUpdateRefusesLegacyStateWithoutConfiguredPath(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, ".agenticlens-config.yaml")
	body := []byte("workspaces:\n  artifacts: []\n  roots:\n    - path: records\n      discovery:\n        max_depth: 8\n        max_directories: 300\nagenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: LegacyProject\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: \"TEST\"\n      description: Fixture ticket\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	if err := os.WriteFile(config, body, 0600); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(root, ".agenticlens-capture")
	if err := os.Mkdir(legacy, 0700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(legacy, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("CREATE TABLE capture_meta(version INTEGER NOT NULL);INSERT INTO capture_meta VALUES(1);CREATE TABLE retained_pending(value TEXT);INSERT INTO retained_pending VALUES('pending capture')"); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, "init", "--update-capture-config")
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "CODEX_HOME="+filepath.Join(home, ".codex"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("configuration update orphaned retained legacy state: %s", output)
	}
	actual, e := os.ReadFile(config)
	if e != nil || !bytes.Equal(actual, body) {
		t.Fatalf("refused update changed original configuration: %v", e)
	}
}

func TestCaptureInstallIgnoresLogicalRegistryAndPreservesUserRules(t *testing.T) {
	project := t.TempDir()
	source := t.TempDir()
	config := []byte("workspaces:\n  artifacts: []\n  roots:\n    - path: records\n      discovery:\n        max_depth: 8\n        max_directories: 300\nagenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: LocalCapture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: \"TEST\"\n      description: Fixture ticket\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), config, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: true\n    state_dir: .capture-local\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ownerRules := []byte("# Owner rules retained\n/retained-legacy/\nunrelated-output\n")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), ownerRules, 0600); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) []byte {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = project
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return out
	}
	git("init", "-q")
	args := prepareCaptureInstall(t, project, []string{"install", "--repo", project, "--source", "codex", "--store", source})
	wrapper := fixtureWrapperFromHook(t, project)
	cmd := exec.Command(wrapper, append([]string{"--al-offline", "--al-project", project, "capture"}, args...)...)
	cmd.Dir = project
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("real installed wrapper/native install: %v %s", err, output)
	}
	registry, err := capture.ProjectState(filepath.Join(project, ".capture-local"), project)
	if err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(registry, project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = store.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "completed"}); err != nil {
		t.Fatal(err)
	}
	coord := store.CoordinationDirectory()
	lock, err := capture.AcquireWorker(coord)
	if err != nil || lock == nil {
		t.Fatalf("worker coordination: %v", err)
	}
	defer lock.Close()
	if err = capture.WriteWorkerProgress(coord, capture.WorkerProgress{State: "completed"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(registry, "registry.db"), filepath.Join(registry, "registry.db-wal"), filepath.Join(registry, "registry.db-shm"), filepath.Join(coord, "install-codex.json"), filepath.Join(coord, "worker.lock"), filepath.Join(coord, "worker-status.json")} {
		if _, err = os.Stat(path); err != nil {
			t.Fatalf("expected real state file absent: %s: %v", path, err)
		}
		git("check-ignore", "--", path)
	}
	for _, relative := range []string{"retained-legacy/pending-state", "unrelated-output"} {
		git("check-ignore", "--", relative)
	}
	ignore, err := os.ReadFile(filepath.Join(project, ".gitignore"))
	if err != nil || !bytes.HasPrefix(ignore, ownerRules) {
		t.Fatalf("user/legacy ignore content changed: %v %s", err, ignore)
	}
	expected := "/.capture-local/projects/LocalCapture/"
	if !bytes.Contains(ignore, []byte(expected+"\n")) {
		t.Fatalf("logical registry ignore absent: %s", ignore)
	}
	status := git("status", "--short", "--untracked-files=all")
	if bytes.Contains(status, []byte("registry.db")) {
		t.Fatalf("private capture database leaked to Git status: %s", status)
	}
}

func TestCaptureInspectRetainedStateThroughManagerAfterWorktreeDeletion(t *testing.T) {
	deleted := newCaptureGenerationFixture(t)
	survivor := newCaptureGenerationFixture(t)
	prepareCaptureInstall(t, survivor.project, []string{"install", "--repo", survivor.project, "--source", "codex", "--store", survivor.store})
	store, err := capture.OpenProject(deleted.state, deleted.project)
	if err != nil {
		t.Fatal(err)
	}
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: deleted.store, SessionID: "retained-deleted-root"}
	if _, err := store.Register(context.Background(), identity, "", "SessionStart"); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if err := store.RequestPublication(context.Background()); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	database := filepath.Join(deleted.state, "registry.db")
	before := readCaptureFile(t, database)
	if err := os.RemoveAll(deleted.project); err != nil {
		t.Fatal(err)
	}
	args := []string{"capture", "inspect-state", "--repo", deleted.project, "--state-dir", deleted.state, "--project-name", "CaptureFixture"}
	for _, invocation := range [][]string{{cliBinaryPath}, {managerBinaryPath, "--al-project", survivor.project}} {
		command := exec.Command(invocation[0], append(append([]string{}, invocation[1:]...), args...)...)
		command.Dir = survivor.project
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("retained inspection through %v failed: %v %s", invocation, err, output)
		}
		var inspection capture.RetainedStateInspection
		if err := json.Unmarshal(output, &inspection); err != nil {
			t.Fatal(err)
		}
		if inspection.Project != deleted.project || inspection.ProjectName != "CaptureFixture" || len(inspection.Registrations) != 1 || inspection.Registrations[0].Identity != identity || inspection.Publication.Generation != 1 || inspection.Publication.Acknowledged != 0 {
			t.Fatalf("retained observations or pending work changed: %+v", inspection)
		}
		if !bytes.Equal(before, readCaptureFile(t, database)) {
			t.Fatal("read-only inspection mutated retained registry")
		}
		for _, path := range []string{deleted.project, survivor.state} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("inspection recreated checkout or initialized survivor state %s: %v", path, err)
			}
		}
	}
	// The read-only exception must not admit an ordinary cross-project command.
	command := exec.Command(managerBinaryPath, "--al-project", survivor.project, "capture", "status", "--repo", deleted.project)
	command.Dir = survivor.project
	if output, err := command.CombinedOutput(); err == nil || !strings.Contains(string(output), "AgenticLens manager:") || !strings.Contains(string(output), deleted.project) {
		t.Fatalf("ordinary capture unexpectedly bypassed manager's deleted-project guard: %v %s", err, output)
	}
}
