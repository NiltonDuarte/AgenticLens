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
	"strconv"
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
	cfg := "project_name: SharedCLI\ntelemetry:\n  capture:\n    state_dir: " + strconv.Quote(base) + "\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(repo, ".agenticlens-config.yaml"), []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", ".agenticlens-config.yaml")
	git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
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
			cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", source, "--capture-dir", registry, "--ticket-id", "TEST-100", "--ticket-title", "Shared ticket", "--ticket-type", "feat", "--json-output", output, "--append", summary)
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
	body := []byte("project_name: LegacyProject\nticket:\n  prefixes: [TEST]\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
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
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
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
	config := []byte("project_name: LocalCapture\ntelemetry:\n  capture:\n    state_dir: .capture-local\nticket:\n  prefixes: [TEST]\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), config, 0600); err != nil {
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
