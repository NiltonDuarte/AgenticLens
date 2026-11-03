package capture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoordinationTokenAliasesRejectUnknownAndRedirectedState(t *testing.T) {
	project, _, base := sharedWorktrees(t)
	store, err := OpenProject(base, project)
	if err != nil {
		t.Fatal(err)
	}
	coordinate, registry := store.CoordinationDirectory(), store.RegistryDirectory()
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	preview, err := PreviewStateLocation(coordinate, project)
	if err != nil || preview.Provisional || preview.CoordinationDirectory != coordinate {
		t.Fatalf("persisted token alias: %+v %v", preview, err)
	}
	unknown := filepath.Join(registry, "worktrees", "unknown-token")
	if _, err = PreviewStateLocation(unknown, project); err == nil {
		t.Fatal("unknown token preview accepted")
	}
	if opened, e := OpenProject(unknown, project); e == nil {
		opened.Close()
		t.Fatal("unknown token allocated another scope")
	}
	if _, err = os.Stat(unknown); !os.IsNotExist(err) {
		t.Fatalf("unknown token was created: %v", err)
	}
	retained := coordinate + "-retained"
	if err = os.Rename(coordinate, retained); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err = os.Symlink(outside, coordinate); err != nil {
		t.Fatal(err)
	}
	if opened, e := OpenReadOnlyProject(base, project); e == nil {
		opened.Close()
		t.Fatal("persisted token followed a redirected coordination path")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("redirected state was modified: %v %v", entries, err)
	}
}
