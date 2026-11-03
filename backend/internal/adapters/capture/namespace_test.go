package capture

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNamespaceAliasesResolveOneContainedRegistry(t *testing.T) {
	for _, kind := range []string{"ordinary", "base_alias", "contained_namespace_alias"} {
		t.Run(kind, func(t *testing.T) {
			project, _, base := sharedWorktrees(t)
			selected := base
			want := filepath.Join(base, "projects", "SharedProject")
			if kind == "base_alias" {
				if err := os.MkdirAll(base, 0700); err != nil {
					t.Fatal(err)
				}
				selected = filepath.Join(filepath.Dir(base), "state-alias")
				if err := os.Symlink(base, selected); err != nil {
					t.Fatal(err)
				}
			} else if kind == "contained_namespace_alias" {
				want = filepath.Join(base, "inside", "projects", "SharedProject")
				linkNamespace(t, base, "SharedProject", want)
			}
			preview, err := PreviewStateLocation(selected, project)
			if err != nil || preview.RegistryDirectory != want {
				t.Fatalf("preview=%+v err=%v want=%s", preview, err, want)
			}
			if _, err := os.Stat(filepath.Join(want, "registry.db")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("preview allocated registry: %v", err)
			}
			registry, err := ProjectState(selected, project)
			if err != nil || registry != want {
				t.Fatalf("registry=%s err=%v want=%s", registry, err, want)
			}
			store, err := OpenProject(registry, project)
			if err != nil {
				t.Fatal(err)
			}
			coord := store.CoordinationDirectory()
			if store.RegistryDirectory() != registry {
				t.Fatalf("opened registry differs: %s != %s", store.RegistryDirectory(), registry)
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			coordAlias := filepath.Join(filepath.Dir(base), "coord-alias")
			if err := os.Symlink(coord, coordAlias); err != nil {
				t.Fatal(err)
			}
			logical := filepath.Join(selected, "projects", "SharedProject")
			for _, input := range []string{selected, logical, registry, coord, coordAlias} {
				got, err := ProjectState(input, project)
				if err != nil || got != registry {
					t.Fatalf("idempotent resolution of %s: %s %v", input, got, err)
				}
				location, err := ExistingLocation(input, project)
				if err != nil || location.RegistryDirectory != registry || location.CoordinationDirectory != coord {
					t.Fatalf("existing resolution of %s: %+v %v", input, location, err)
				}
			}
		})
	}
}

func TestNamespaceAliasesRejectEscapeAndMalformedTargetsBeforeAllocation(t *testing.T) {
	for _, kind := range []string{"escaping", "escaping_completed_namespace", "malformed_contained"} {
		t.Run(kind, func(t *testing.T) {
			project, _, base := sharedWorktrees(t)
			target := t.TempDir()
			if kind == "escaping_completed_namespace" {
				target = filepath.Join(target, "projects", "SharedProject")
			} else if kind == "malformed_contained" {
				target = filepath.Join(base, "inside")
			}
			linkNamespace(t, base, "SharedProject", target)
			logical := filepath.Join(base, "projects", "SharedProject")
			for _, input := range []string{base, logical, filepath.Join(logical, "worktrees", "1")} {
				if got, err := ProjectState(input, project); err == nil {
					t.Fatalf("unsafe namespace accepted %s => %s", input, got)
				}
				if got, err := PreviewStateLocation(input, project); err == nil {
					t.Fatalf("unsafe preview accepted %s => %+v", input, got)
				}
				if store, err := OpenProject(input, project); err == nil {
					if closeErr := store.Close(); closeErr != nil {
						t.Error(closeErr)
					}
					t.Fatalf("unsafe open accepted %s", input)
				}
			}
			for _, path := range []string{filepath.Join(target, "registry.db"), filepath.Join(target, "projects", "SharedProject", "registry.db")} {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("rejected namespace allocated %s: %v", path, err)
				}
			}
		})
	}
}

func TestRelocationRejectsNamespaceAliasBeforeSourceFence(t *testing.T) {
	for _, kind := range []string{"escaping", "escaping_completed_namespace", "malformed_contained"} {
		t.Run(kind, func(t *testing.T) {
			project, source, base, old := relocationFixture(t)
			target := t.TempDir()
			if kind == "escaping_completed_namespace" {
				target = filepath.Join(target, "projects", "Shared")
			} else if kind == "malformed_contained" {
				target = filepath.Join(base, "inside")
			}
			linkNamespace(t, base, "Shared", target)
			configPath := filepath.Join(project, ".agenticlens-config.yaml")
			before, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			for _, dryRun := range []bool{true, false} {
				result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base, DryRun: dryRun})
				if err == nil {
					t.Fatalf("unsafe relocation accepted dryRun=%t: %+v", dryRun, result)
				}
			}
			var version, count, generation, acknowledged int
			if err := old.QueryRow("SELECT version FROM capture_meta").Scan(&version); err != nil || version != 1 {
				t.Fatalf("rejected source fenced: version=%d err=%v", version, err)
			}
			if err := old.QueryRow("SELECT count(*) FROM registrations").Scan(&count); err != nil || count != 2 {
				t.Fatalf("source registrations changed: count=%d err=%v", count, err)
			}
			if err := old.QueryRow("SELECT generation,acknowledged FROM publication_request").Scan(&generation, &acknowledged); err != nil || generation != 7 || acknowledged != 3 {
				t.Fatalf("source work changed: %d/%d %v", generation, acknowledged, err)
			}
			after, err := os.ReadFile(configPath)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("rejected relocation changed config: %v", err)
			}
			if _, err := os.Stat(filepath.Join(target, "registry.db")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("rejected relocation created destination: %v", err)
			}
		})
	}
}

func TestRelocationContainedNamespaceAliasReopensPreservingEvidence(t *testing.T) {
	project, source, base, _ := relocationFixture(t)
	target := filepath.Join(base, "inside", "projects", "Shared")
	linkNamespace(t, base, "Shared", target)
	result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if err != nil || result.Phase != "completed" || result.Destination != target {
		t.Fatalf("relocation=%+v err=%v", result, err)
	}
	registry, err := ProjectState(base, project)
	if err != nil || registry != result.Destination {
		t.Fatalf("normal state differs after migration: %s %v", registry, err)
	}
	store, err := OpenProject(registry, project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertRelocationRegistrations(t, rows, project)
	request, err := store.PublicationRequest(context.Background())
	if err != nil || request.Generation != 7 || request.Acknowledged != 3 {
		t.Fatalf("publication work changed: %+v %v", request, err)
	}
	assertRelocationContinuations(t, store, project)
}

func linkNamespace(t *testing.T, base, name, target string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(base, "projects"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(base, "projects", name)); err != nil {
		t.Fatal(err)
	}
}
