package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

type installTransactionFixture struct {
	options     InstallOptions
	repo, state string
	manifest    Manifest
	changes     []fileChange
}

func newInstallTransactionFixture(t *testing.T, source string) installTransactionFixture {
	t.Helper()
	repo := t.TempDir()
	state := filepath.Join(repo, ".capture", "projects", "Fixture", "worktrees", "1")
	data, err := adaptersconfig.GenerateDefaultYAMLForVersion("v9.8.7", "Fixture", "workspaces", "", []string{"src"}, []domain.TicketPrefix{{Prefix: "T"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(repo, ".agenticlens-config.yaml"), data, 0600); err != nil {
		t.Fatal(err)
	}
	store, err := CanonicalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	options := InstallOptions{Repository: repo, StateDir: state, Source: source, SourceStore: store, Remove: true}
	config := filepath.Join(repo, ".codex", "hooks.json")
	before := []byte(`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"foreign command retained"}]},{"hooks":[{"type":"command","command":"old owned hook"}]}]}}`)
	after := []byte(`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"foreign command retained"}]}]}}`)
	if source == "opencode" {
		config = filepath.Join(repo, ".opencode", "plugins", "agenticlens-capture.ts")
		before = []byte("// previously owned plugin\n")
		after = nil
	}
	m := Manifest{Protocol: ProtocolVersion, Repository: repo, StateDir: state, Source: source, SourceStore: store, ConfigPath: config, Version: "v9.8.7", Active: false}
	encoded, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	changes := []fileChange{{path: config, before: before, content: after, existed: true, remove: source == "opencode"}, {path: filepath.Join(repo, ".gitignore"), before: []byte("# owner rule\nowner-output\n"), content: []byte("# owner rule\nowner-output\n/.capture/projects/Fixture/\n"), existed: true}, {path: manifestPath(state, source), before: []byte(`{"old":"manifest"}`), content: encoded, existed: true}, {path: filepath.Join(repo, "."+source, "agenticlens-installation.json"), before: []byte(`{"old":"anchor"}`), content: encoded, existed: true}}
	for _, c := range changes {
		if err = apply(fileChange{path: c.path, content: c.before}); err != nil {
			t.Fatal(err)
		}
	}
	return installTransactionFixture{options, repo, state, m, changes}
}
func assertInstallFiles(t *testing.T, changes []fileChange, after bool) {
	t.Helper()
	for _, c := range changes {
		expected, exists := c.before, c.existed
		if after {
			expected, exists = c.content, !c.remove
		}
		actual, err := os.ReadFile(c.path)
		if !exists {
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("expected absent target %s: %v", c.path, err)
			}
			continue
		}
		if err != nil || !bytes.Equal(expected, actual) {
			t.Fatalf("unexpected target %s: %v %s", c.path, err, actual)
		}
	}
}

func TestInstallerDurablePrefixRecoversHookPluginAndManifestCheckpoints(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		for prefix := 1; prefix <= 4; prefix++ {
			t.Run(fmt.Sprintf("%s/prefix%d", source, prefix), func(t *testing.T) {
				f := newInstallTransactionFixture(t, source)
				writes := 0
				func() {
					defer func() {
						if r := recover(); r != "simulated process death" {
							t.Fatalf("unexpected interruption: %v", r)
						}
					}()
					err := applyInstallTransactionWithWriter(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
						if err := apply(c); err != nil {
							return err
						}
						writes++
						if writes == prefix {
							panic("simulated process death")
						}
						return nil
					})
					t.Fatalf("interruption did not occur: %v", err)
				}()
				path := installJournalPath(f.state, source)
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("process interruption discarded recovery journal: %v", err)
				}
				result, err := recoverInstallTransaction(context.Background(), f.options, f.repo, f.state)
				if err != nil {
					t.Fatal(err)
				}
				if prefix == len(f.changes) {
					if result == nil || result.Manifest.Source != source {
						t.Fatal("verified completed transaction was not reused")
					}
					assertInstallFiles(t, f.changes, true)
				} else {
					if result != nil {
						t.Fatal("partial prefix reported complete")
					}
					assertInstallFiles(t, f.changes, false)
					if err = applyInstallTransaction(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes); err != nil {
						t.Fatal(err)
					}
					assertInstallFiles(t, f.changes, true)
				}
				if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("verified completion retained pending journal: %v", err)
				}
			})
		}
	}
}

func TestInstallerPendingRecoveryPreservesEditsAndExactContext(t *testing.T) {
	for _, scenario := range []string{"edited-hook", "different-state", "changed-pin", "dry-run"} {
		t.Run(scenario, func(t *testing.T) {
			f := newInstallTransactionFixture(t, "codex")
			func() {
				defer func() {
					if r := recover(); r != "simulated process death" {
						t.Fatalf("unexpected interruption: %v", r)
					}
				}()
				unexpected := applyInstallTransactionWithWriter(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
					if err := apply(c); err != nil {
						return err
					}
					panic("simulated process death")
				})
				t.Fatalf("expected interruption, got %v", unexpected)
			}()
			original, err := os.ReadFile(installJournalPath(f.state, "codex"))
			if err != nil {
				t.Fatal(err)
			}
			options, repo, state := f.options, f.repo, f.state
			switch scenario {
			case "edited-hook":
				if err = apply(fileChange{path: f.changes[0].path, content: []byte("user edits must survive")}); err != nil {
					t.Fatal(err)
				}
			case "different-state":
				state = t.TempDir()
				options.StateDir = state
				if err = apply(fileChange{path: installJournalPath(state, "codex"), content: original}); err != nil {
					t.Fatal(err)
				}
			case "dry-run":
				options.DryRun = true
			case "changed-pin":
				config := filepath.Join(repo, ".agenticlens-config.yaml")
				body, err := os.ReadFile(config)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(config, bytes.ReplaceAll(body, []byte("v9.8.7"), []byte("v9.8.8")), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := map[string][]byte{}
			for _, c := range f.changes {
				body, err := os.ReadFile(c.path)
				if err != nil {
					t.Fatal(err)
				}
				before[c.path] = body
			}
			if _, err = recoverInstallTransaction(context.Background(), options, repo, state); err == nil {
				t.Fatal("unowned pending recovery accepted")
			}
			for path, want := range before {
				body, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(want, body) {
					t.Fatalf("failed recovery overwrote file %s: %v", path, err)
				}
			}
			current, err := os.ReadFile(installJournalPath(state, "codex"))
			if err != nil || !bytes.Equal(current, original) {
				t.Fatalf("failed recovery discarded pending evidence: %v", err)
			}
		})
	}
}

func TestInstallerOrdinaryFailureRollsBackOnlyOwnedPrefix(t *testing.T) {
	f := newInstallTransactionFixture(t, "codex")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := applyInstallTransactionWithWriter(ctx, f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
		if err := apply(c); err != nil {
			return err
		}
		cancel()
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ordinary failure not retained: %v", err)
	}
	assertInstallFiles(t, f.changes, false)
	if _, err = os.Stat(installJournalPath(f.state, "codex")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successful rollback left an unfinished journal: %v", err)
	}
}

func TestInstallerFailedWriteWithVisibleOutputRestoresBefore(t *testing.T) {
	f := newInstallTransactionFixture(t, "codex")
	expected := errors.New("write reported failure after output became visible")
	err := applyInstallTransactionWithWriter(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
		if err := apply(c); err != nil {
			return err
		}
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("write failure lost: %v", err)
	}
	assertInstallFiles(t, f.changes, false)
	if _, err = os.Stat(installJournalPath(f.state, "codex")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("successful rollback left journal: %v", err)
	}
}

func TestInstallerInterruptedRollbackRetainsRecoverableOwnedOutputs(t *testing.T) {
	f := newInstallTransactionFixture(t, "codex")
	func() {
		defer func() {
			if r := recover(); r != "simulated process death" {
				t.Fatalf("unexpected interruption: %v", r)
			}
		}()
		err := applyInstallTransactionWithWriter(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
			if err := apply(c); err != nil {
				return err
			}
			if c.path == f.changes[2].path {
				panic("simulated process death")
			}
			return nil
		})
		t.Fatalf("expected interruption, got %v", err)
	}()
	path := installJournalPath(f.state, "codex")
	j, raw, err := readInstallJournal(path)
	if err != nil {
		t.Fatal(err)
	}
	j.Phase = "rollback"
	j.RollbackPrefix = 3
	if err = saveInstallJournal(path, j, raw); err != nil {
		t.Fatal(err)
	}
	// Durable reverse-prefix: the manifest has been restored, then the owner dies.
	c := f.changes[2]
	if err = apply(fileChange{path: c.path, content: c.before, remove: !c.existed}); err != nil {
		t.Fatal(err)
	}
	if _, err = recoverInstallTransaction(context.Background(), f.options, f.repo, f.state); err != nil {
		t.Fatal(err)
	}
	assertInstallFiles(t, f.changes, false)
	if _, err = os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("rollback completion did not clear journal: %v", err)
	}
}
