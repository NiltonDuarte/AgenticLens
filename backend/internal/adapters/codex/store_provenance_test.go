package codex_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
)

func TestRolloutStorePhysicalMembership(t *testing.T) {
	home, unrelated := t.TempDir(), t.TempDir()
	for _, family := range []string{"sessions", "archived_sessions"} {
		dir := filepath.Join(home, family)
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "owner.jsonl")
		if err := os.WriteFile(path, []byte("source"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(home, path); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(unrelated, path); err == nil {
			t.Fatal("foreign home accepted")
		}
		alias := filepath.Join(t.TempDir(), "home")
		if err := os.Symlink(home, alias); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(alias, path); err != nil {
			t.Fatalf("physical alias rejected: %v", err)
		}
		escape := filepath.Join(dir, "escape.jsonl")
		foreign := filepath.Join(unrelated, "outside.jsonl")
		if err := os.WriteFile(foreign, []byte("outside"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(foreign, escape); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(home, escape); err == nil {
			t.Fatal("physical escape accepted")
		}
		child := filepath.Join(unrelated, "child-"+family)
		if err := os.Mkdir(child, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(child, filepath.Join(dir, "hop")); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(home, dir+"/hop/../outside.jsonl"); err == nil {
			t.Fatal("raw symlink-dotdot source escape accepted")
		}
		homeChild := filepath.Join(home, "child-"+family)
		if err := os.Mkdir(homeChild, 0700); err != nil {
			t.Fatal(err)
		}
		aliasParent := t.TempDir()
		if err := os.Symlink(homeChild, filepath.Join(aliasParent, "hop")); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(aliasParent+"/hop/..", path); err != nil {
			t.Fatalf("physical whole-home traversal alias rejected: %v", err)
		}
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err := codex.ValidateRolloutStore(home, path); err != nil {
			t.Fatalf("missing accepted same-store path rejected: %v", err)
		}
	}
}

func TestNativeSessionContextKeepsChildIdentityAndStructuredParent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "child.jsonl")
	body := `{"type":"session_meta","payload":{"id":"child-session","session_id":"parent-session","cwd":"/owning-worktree","source":{"subagent":{"thread_spawn":{"parent_thread_id":"parent-session"}}}}}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	metadata, err := codex.ReadSessionContext(path)
	if err != nil || metadata.ID != "child-session" || metadata.ParentID != "parent-session" || metadata.WorkingDirectory != "/owning-worktree" {
		t.Fatalf("native child identity/ancestry rewritten: %+v %v", metadata, err)
	}
	if _, err := codex.SessionContext(path, "parent-session"); !errors.Is(err, codex.ErrSessionContextMismatch) {
		t.Fatalf("exact parent lookup accepted child metadata: %v", err)
	}
	if cwd, err := codex.SessionContext(path, "child-session"); err != nil || cwd != "/owning-worktree" {
		t.Fatalf("exact native child context: %s %v", cwd, err)
	}
	conflicting := strings.Replace(body, `"cwd":`, `"parent_thread_id":"different-root","cwd":`, 1)
	if err := os.WriteFile(path, []byte(conflicting), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := codex.ReadSessionContext(path); !errors.Is(err, codex.ErrSessionContextMismatch) {
		t.Fatalf("conflicting explicit/source parent accepted: %v", err)
	}
	for _, pending := range []string{"", `{"type":"session_meta","payload":`, `{"type":"session_meta","payload":{"id":"child-session"}}`} {
		if err := os.WriteFile(path, []byte(pending), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := codex.ReadSessionContext(path); err == nil || errors.Is(err, codex.ErrSessionContextMismatch) {
			t.Fatalf("unflushed metadata became known contradiction: %v", err)
		}
	}
}
