package capture_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func captureProjectConfig(t *testing.T, project, name string) {
	t.Helper()
	data := fmt.Sprintf("project_name: %q\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n", name)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestProjectStateAliasesIsolationAndCopiedOwnership(t *testing.T) {
	a, b, base := t.TempDir(), t.TempDir(), t.TempDir()
	captureProjectConfig(t, a, "ProjectA")
	captureProjectConfig(t, b, "ProjectB")
	one, err := capture.ProjectState(base, a)
	if err != nil {
		t.Fatal(err)
	}
	two, err := capture.ProjectState(base, b)
	if err != nil {
		t.Fatal(err)
	}
	if one == two {
		t.Fatal("distinct project directories collapsed")
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err = os.Symlink(a, alias); err != nil {
		t.Fatal(err)
	}
	same, err := capture.ProjectState(base, alias)
	if err != nil || same != one {
		t.Fatalf("alias split: %s %v", same, err)
	}
	s, err := capture.OpenProject(one, a)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if wrong, err := capture.OpenProject(one, b); err == nil {
		wrong.Close()
		t.Fatal("copied state reassigned")
	}
	if err = capture.ValidateProject(a, b); err == nil {
		t.Fatal("caller project overrode source project")
	}
	if err = capture.ValidateDestination(a, filepath.Join(a, "..", filepath.Base(a)+"-other", "ledger")); err == nil {
		t.Fatal("string prefix authorized destination")
	}
	if err = capture.ValidateDestination(a, filepath.Join(alias, "nested", "ledger")); err != nil {
		t.Fatal(err)
	}
	escape := filepath.Join(a, "escape")
	if err = os.Symlink(b, escape); err != nil {
		t.Fatal(err)
	}
	if err = capture.ValidateDestination(a, filepath.Join(escape, "new", "ledger")); err == nil {
		t.Fatal("nested symlink escaped")
	}
}

func TestProjectDestinationResolvesDanglingSymlinks(t *testing.T) {
	for _, outside := range []bool{false, true} {
		for _, ancestor := range []bool{false, true} {
			t.Run(fmt.Sprintf("outside=%t/ancestor=%t", outside, ancestor), func(t *testing.T) {
				project := t.TempDir()
				base := project
				if outside {
					base = t.TempDir()
				}
				target := filepath.Join(base, "absent")
				link := filepath.Join(project, "link")
				// A relative target must be resolved from the link's directory.
				relative, err := filepath.Rel(project, target)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.Symlink(relative, link); err != nil {
					t.Fatal(err)
				}
				path := link
				if ancestor {
					path = filepath.Join(link, "nested", "ledger.json")
					target = filepath.Join(target, "nested", "ledger.json")
				}
				resolved, err := capture.ProjectDirectory(path)
				if err != nil || resolved != target {
					t.Fatalf("physical destination: %s %v; want %s", resolved, err, target)
				}
				if err = capture.ValidateDestination(project, path); (err != nil) != outside {
					t.Fatalf("containment outside=%t: %v", outside, err)
				}
			})
		}
	}
	project := t.TempDir()
	loop := filepath.Join(project, "loop")
	if err := os.Symlink("loop", loop); err != nil {
		t.Fatal(err)
	}
	if err := capture.ValidateDestination(project, loop); err == nil {
		t.Fatal("symlink cycle accepted")
	}
}

func TestProjectDirectoryTraversesSymlinksBeforeDotDot(t *testing.T) {
	project, outside := t.TempDir(), t.TempDir()
	child := filepath.Join(outside, "child")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(child, filepath.Join(project, "hop")); err != nil {
		t.Fatal(err)
	}
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			victim := filepath.Join(outside, "summary")
			if existing {
				if err := os.WriteFile(victim, []byte("sentinel"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			alias := filepath.Join(project, "alias-"+fmt.Sprint(existing))
			if err := os.Symlink("hop/../summary", alias); err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{alias, project + "/hop/../summary"} {
				resolved, err := capture.ProjectDirectory(path)
				if err != nil || resolved != victim {
					t.Fatalf("physical traversal: %s %v; want %s", resolved, err, victim)
				}
				if existing {
					native, err := filepath.EvalSymlinks(path)
					if err != nil || native != resolved {
						t.Fatalf("native filesystem differs: %s %v", native, err)
					}
				}
				if err := capture.ValidateDestination(project, path); err == nil {
					t.Fatal("outside target authorized")
				}
				store, err := capture.CanonicalStore(path)
				if err != nil || store != victim {
					t.Fatalf("store differs from physical traversal: %s %v", store, err)
				}
			}
		})
	}
}

func TestPublicationLockSharesPhysicalAliasBeforeAndAfterCreation(t *testing.T) {
	for _, directoryAlias := range []bool{false, true} {
		for _, state := range []string{"missing", "existing", "created-during-lock"} {
			t.Run(fmt.Sprintf("directory=%t/%s", directoryAlias, state), func(t *testing.T) {
				t.Setenv("XDG_CACHE_HOME", t.TempDir())
				project := t.TempDir()
				target := filepath.Join(project, "ledger.json")
				alias := filepath.Join(project, "alias.json")
				linkTarget := target
				if directoryAlias {
					alias = filepath.Join(t.TempDir(), "project")
					linkTarget = project
				}
				if err := os.Symlink(linkTarget, alias); err != nil {
					t.Fatal(err)
				}
				if directoryAlias {
					alias = filepath.Join(alias, "ledger.json")
				}
				if state == "existing" {
					if err := os.WriteFile(target, []byte("before"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				first, err := capture.AcquirePublication(context.Background(), alias)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if first != nil {
						first.Close()
					}
				}()
				if state == "created-during-lock" {
					if err := os.WriteFile(target, []byte("during"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				second, err := capture.AcquirePublication(ctx, target)
				if err == nil {
					second.Close()
					t.Fatal("same physical ledger has simultaneous publication owners")
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("unexpected contention error: %v", err)
				}
				if err := first.Close(); err != nil {
					t.Fatal(err)
				}
				first = nil
				second, err = capture.AcquirePublication(context.Background(), target)
				if err != nil {
					t.Fatal(err)
				}
				if err := second.Close(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestPublicationGenerationAcknowledgesOnlyObservedWork(t *testing.T) {
	s := openStore(t, t.TempDir())
	ctx := context.Background()
	if err := s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	first, err := s.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if err = s.AcknowledgePublication(ctx, first.Generation); err != nil {
		t.Fatal(err)
	}
	next, err := s.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if next.Generation == next.Acknowledged || next.Acknowledged != first.Generation {
		t.Fatalf("unseen event consumed: %+v", next)
	}
	// Same ticket and source store cannot bridge project namespaces.
	base, a, b := t.TempDir(), t.TempDir(), t.TempDir()
	captureProjectConfig(t, a, "SharedProject")
	captureProjectConfig(t, b, "SharedProject")
	one, _ := capture.ProjectState(base, a)
	two, _ := capture.ProjectState(base, b)
	sa, err := capture.OpenProject(one, a)
	if err != nil {
		t.Fatal(err)
	}
	defer sa.Close()
	sb, err := capture.OpenProject(two, b)
	if err != nil {
		t.Fatal(err)
	}
	defer sb.Close()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
	sourcePath := filepath.Join(id.Store, "sessions", "root.jsonl")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0700); err != nil {
		t.Fatal(err)
	}
	metadata := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"root","cwd":%q,"timestamp":"2026-01-01T00:00:00Z"}}`+"\n", a)
	if err := os.WriteFile(sourcePath, []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}
	row, err := sa.Register(ctx, id, sourcePath, "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = sa.Bind(ctx, id, "AL-68", row.Request); err != nil {
		t.Fatal(err)
	}
	if rows, err := sb.List(ctx); err != nil || len(rows) != 0 {
		t.Fatalf("cross-project registrations: %+v %v", rows, err)
	}
}
