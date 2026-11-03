package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommittedLedgerPinsRegularExactBlob(t *testing.T) {
	project := t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = project
		body, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, body)
		}
		return strings.TrimSpace(string(body))
	}
	run("init", "-qb", "main")
	relative := "records with spaces/ticket/ledger [literal].json"
	if err := os.MkdirAll(filepath.Dir(filepath.Join(project, relative)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, relative), []byte("committed ledger\n"), 0600); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "first")
	commit, err := ResolveHistoryCommit(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(project, relative), []byte("uncommitted ledger"), 0600); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	got, err := ResolveCommittedLedger(context.Background(), project, commit, relative)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Data) != "committed ledger\n" || got.Commit != commit || got.Blob == "" {
		t.Fatalf("read working/index state: %+v", got)
	}
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "later")
	got, err = ResolveCommittedLedger(context.Background(), project, commit, relative)
	if err != nil || string(got.Data) != "committed ledger\n" {
		t.Fatalf("pinned blob drift: %+v %v", got, err)
	}
	for _, path := range []string{"missing.json", "../outside.json", "/outside.json", "records with spaces/../ledger.json", ""} {
		if _, err = ResolveCommittedLedger(context.Background(), project, commit, path); err == nil {
			t.Fatalf("invalid/missing path accepted: %q", path)
		}
	}
	if _, err = ResolveCommittedLedger(context.Background(), project, "HEAD", relative); err == nil {
		t.Fatal("mutable ref accepted as pinned object")
	}
	// Install a symlink mode directly in the Git index, independent of host privileges.
	run("update-index", "--add", "--cacheinfo", "120000,"+got.Blob+",link.json")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "symlink")
	head, err := ResolveHistoryCommit(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ResolveCommittedLedger(context.Background(), project, head, "link.json"); err == nil {
		t.Fatal("symlink blob accepted")
	}
	large := filepath.Join(project, "large.json")
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Truncate(MaxCommittedLedgerBytes + 1); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	run("add", "large.json")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "oversized")
	head, err = ResolveHistoryCommit(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ResolveCommittedLedger(context.Background(), project, head, "large.json"); err == nil {
		t.Fatal("oversized committed ledger accepted")
	}
}

func TestCommittedHistoryIgnoresHookRepositoryAndObjectEnvironment(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	actual, foreign := t.TempDir(), t.TempDir()
	run := func(project string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = project
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	for i, project := range []string{actual, foreign} {
		run(project, "init", "-qb", "main")
		if err := os.WriteFile(filepath.Join(project, "ledger.json"), []byte([]string{"actual committed ledger\n", "foreign committed ledger\n"}[i]), 0600); err != nil {
			t.Fatal(err)
		}
		run(project, "add", "ledger.json")
		run(project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
	}
	expected := run(actual, "rev-parse", "HEAD")
	foreignCommit := run(foreign, "rev-parse", "HEAD")
	linked := filepath.Join(t.TempDir(), "legitimate linked worktree")
	run(actual, "worktree", "add", "-qb", "linked", linked)
	if err := os.WriteFile(filepath.Join(linked, "ledger.json"), []byte("linked worktree ledger\n"), 0600); err != nil {
		t.Fatal(err)
	}
	run(linked, "add", "ledger.json")
	run(linked, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "linked")
	linkedCommit := run(linked, "rev-parse", "HEAD")
	common, err := filepath.EvalSymlinks(filepath.Join(actual, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	invalidConfig := filepath.Join(t.TempDir(), "invalid config")
	if err = os.WriteFile(invalidConfig, []byte("not a valid git config\n"), 0600); err != nil {
		t.Fatal(err)
	}
	trace := filepath.Join(t.TempDir(), "trace must not be written")
	redirect := map[string]string{
		"GIT_DIR": filepath.Join(foreign, ".git"), "GIT_WORK_TREE": foreign, "GIT_COMMON_DIR": filepath.Join(foreign, ".git"),
		"GIT_OBJECT_DIRECTORY": filepath.Join(foreign, ".git", "objects"), "GIT_ALTERNATE_OBJECT_DIRECTORIES": filepath.Join(foreign, ".git", "objects"),
		"GIT_INDEX_FILE": filepath.Join(foreign, ".git", "index"), "GIT_SHALLOW_FILE": filepath.Join(foreign, ".git", "shallow"),
		"GIT_CONFIG": invalidConfig, "GIT_CONFIG_GLOBAL": invalidConfig, "GIT_CONFIG_SYSTEM": invalidConfig,
		"GIT_CONFIG_COUNT": "1", "GIT_CONFIG_KEY_0": "core.repositoryformatversion", "GIT_CONFIG_VALUE_0": "99",
		"GIT_CONFIG_PARAMETERS": "'core.repositoryformatversion'='99'", "GIT_NAMESPACE": "foreign-namespace", "GIT_PREFIX": "foreign-prefix/",
		"GIT_CEILING_DIRECTORIES": actual, "GIT_TRACE": trace,
	}
	for _, name := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_INDEX_FILE", "GIT_SHALLOW_FILE", "GIT_CONFIG", "GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM", "GIT_CONFIG_PARAMETERS", "hook bundle"} {
		t.Run(name, func(t *testing.T) {
			if name == "hook bundle" {
				for key, value := range redirect {
					t.Setenv(key, value)
				}
			} else {
				t.Setenv(name, redirect[name])
			}
			for _, project := range []string{actual, linked} {
				wantCommit, wantBody := expected, "actual committed ledger\n"
				if project == linked {
					wantCommit, wantBody = linkedCommit, "linked worktree ledger\n"
				}
				got, err := ResolveHistoryCommit(context.Background(), project)
				if err != nil || got != wantCommit {
					t.Fatalf("%s redirected pinned HEAD: got=%s want=%s err=%v", name, got, wantCommit, err)
				}
				body, err := ResolveCommittedLedger(context.Background(), project, got, "ledger.json")
				if err != nil || string(body.Data) != wantBody {
					t.Fatalf("%s redirected committed blob: %+v %v", name, body, err)
				}
				directory, err := ResolveHistoryCommonDirectory(context.Background(), project)
				if err != nil {
					t.Fatal(err)
				}
				directory, err = filepath.EvalSymlinks(directory)
				if err != nil || directory != common {
					t.Fatalf("%s redirected common-directory proof: %s %v", name, directory, err)
				}
				if _, err = ResolveCommittedLedger(context.Background(), project, foreignCommit, "ledger.json"); err == nil {
					t.Fatalf("%s supplied foreign objects as current authority", name)
				}
			}
			if _, err := os.Stat(trace); !os.IsNotExist(err) {
				t.Fatalf("history command inherited Git tracing: %v", err)
			}
		})
	}
	// The original review probe must still demonstrate Git's ambient behavior;
	// only the narrowly scoped history adapter deliberately ignores it.
	cmd := exec.Command("git", "-C", actual, "rev-parse", "HEAD")
	cmd.Env = append(os.Environ(), "GIT_DIR="+filepath.Join(foreign, ".git"))
	probe, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(probe)) != foreignCommit {
		t.Fatalf("original redirection control invalid: %s %v", probe, err)
	}
	body, err := boundedGit(context.Background(), actual, 4, "show", "HEAD:ledger.json")
	if err == nil || body != nil || strings.Contains(err.Error(), "actual committed ledger") {
		t.Fatalf("bounded output exposed partial contents: %q %v", body, err)
	}
}
