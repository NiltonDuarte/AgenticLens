package codex_test

import (
	"os"
	"path/filepath"
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
