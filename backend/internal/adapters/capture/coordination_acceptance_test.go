package capture

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoordinationStateBaseUnderWorktreesRemainsAStateBase(t *testing.T) {
	project, _, _ := sharedWorktrees(t)
	for _, alias := range []bool{false, true} {
		name := "direct"
		if alias {
			name = "physical-alias"
		}
		t.Run(name, func(t *testing.T) {
			physical := filepath.Join(t.TempDir(), "worktrees", "capture-state")
			base := physical
			if alias {
				if err := os.MkdirAll(physical, 0700); err != nil {
					t.Fatal(err)
				}
				base = filepath.Join(t.TempDir(), "state-alias")
				if err := os.Symlink(physical, base); err != nil {
					t.Fatal(err)
				}
			}
			registry := filepath.Join(physical, "projects", "SharedProject")
			preview, err := PreviewStateLocation(base, project)
			if err != nil || !preview.Provisional || preview.RegistryDirectory != registry {
				t.Fatalf("configured base was treated as a coordinate: %+v %v", preview, err)
			}
			if _, err := os.Stat(registry); !os.IsNotExist(err) {
				t.Fatalf("preview initialized state: %v", err)
			}
			store, err := OpenProject(base, project)
			if err != nil {
				t.Fatal(err)
			}
			coordinate := store.CoordinationDirectory()
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			preview, err = PreviewStateLocation(base, project)
			if err != nil || preview.Provisional || preview.CoordinationDirectory != coordinate {
				t.Fatalf("configured base did not reopen its committed scope: %+v %v", preview, err)
			}
		})
	}
}

func TestCoordinationDirectoryNameAliasesRejectUnknownAndRedirectedState(t *testing.T) {
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
		t.Fatalf("persisted directory name alias: %+v %v", preview, err)
	}
	unknown := filepath.Join(registry, "worktrees", "unknown-directory-name")
	if _, err = PreviewStateLocation(unknown, project); err == nil {
		t.Fatal("unknown directory name preview accepted")
	}
	if opened, e := OpenProject(unknown, project); e == nil {
		opened.Close()
		t.Fatal("unknown directory name allocated another scope")
	}
	if _, err = os.Stat(unknown); !os.IsNotExist(err) {
		t.Fatalf("unknown directory name was created: %v", err)
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
		t.Fatal("persisted directory name followed a redirected coordination path")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("redirected state was modified: %v %v", entries, err)
	}
}
