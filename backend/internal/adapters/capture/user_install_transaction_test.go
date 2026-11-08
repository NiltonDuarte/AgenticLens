package capture

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

type userTransactionFixture struct {
	options                                                           UserInstallOptions
	project, state, legacyHooks, legacyAnchor, legacyState, pendingDB string
	paths                                                             []string
}

func newUserTransactionFixture(t *testing.T) userTransactionFixture {
	t.Helper()
	root := t.TempDir()
	binary, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	f := userTransactionFixture{options: UserInstallOptions{Home: filepath.Join(root, "home"), Entrypoint: binary}, project: filepath.Join(root, "project"), state: filepath.Join(root, "legacy-state")}
	for _, dir := range []string{f.options.Home, filepath.Join(f.project, ".codex"), f.state} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	f.options.LegacyRepository = f.project
	f.legacyHooks = filepath.Join(f.project, ".codex", "hooks.json")
	f.legacyAnchor = filepath.Join(f.project, ".codex", "agenticlens-installation.json")
	f.legacyState = filepath.Join(f.state, "install-codex.json")
	owned := json.RawMessage(`{"hooks":[{"type":"command","command":"owned prior native callback"}]}`)
	config := `{"foreign_setting":true,"hooks":{"SessionStart":[` + string(owned) + `],"Stop":[{"hooks":[{"type":"command","command":"foreign project callback"}]}]}}`
	if err := os.WriteFile(f.legacyHooks, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	legacy := Manifest{Protocol: ManifestProtocolVersion, Repository: f.project, StateDir: f.state, Source: "codex", SourceStore: filepath.Join(root, "source"), ConfigPath: f.legacyHooks, Groups: []OwnedGroup{{Event: "SessionStart", Value: owned}}, Active: true}
	body, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{f.legacyAnchor, f.legacyState} {
		if err := os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	store, err := Open(f.state)
	if err != nil {
		t.Fatal(err)
	}
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: legacy.SourceStore, SessionID: "durable-unassigned-owner"}
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
	f.pendingDB = filepath.Join(f.state, "registry.db")
	foreign := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"foreign home callback"}]}]}}`
	if err := os.WriteFile(filepath.Join(f.options.Home, "hooks.json"), []byte(foreign), 0600); err != nil {
		t.Fatal(err)
	}
	f.paths = []string{f.legacyHooks, f.legacyAnchor, f.legacyState, f.pendingDB, filepath.Join(f.options.Home, "hooks.json"), filepath.Join(f.options.Home, "agenticlens-user-installation.json")}
	return f
}

type userArtifactSnapshot struct {
	Exists bool
	Digest string
	Mode   os.FileMode
}

func snapshotUserArtifacts(t *testing.T, paths []string) map[string]userArtifactSnapshot {
	t.Helper()
	snapshot := map[string]userArtifactSnapshot{}
	for _, path := range paths {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			snapshot[path] = userArtifactSnapshot{}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		snapshot[path] = userArtifactSnapshot{true, fmt.Sprintf("%x", digest), info.Mode().Perm()}
	}
	return snapshot
}
func readUserTransactionJournal(t *testing.T, root string) userInstallJournal {
	t.Helper()
	var journal userInstallJournal
	if err := json.Unmarshal(mustReadUserInstall(t, filepath.Join(root, "agenticlens-user-install.pending.json")), &journal); err != nil {
		t.Fatal(err)
	}
	if journal.Version != 1 || len(journal.Changes) == 0 {
		t.Fatalf("invalid interrupted production journal: %+v", journal)
	}
	return journal
}
func stopUserTransactionAt(t *testing.T, f userTransactionFixture, target string) error {
	t.Helper()
	interrupted := errors.New("test interruption after durable production checkpoint")
	var changePaths []string
	_, err := installUserCodexWithCheckpoint(context.Background(), f.options, func(stage userInstallStage, index int) error {
		if stage == userInstallJournaled {
			// Production writes the complete journal once before applying changes.
			// Keep its actual path order without decoding executable bytes per prefix.
			journal := readUserTransactionJournal(t, f.options.Home)
			for _, change := range journal.Changes {
				changePaths = append(changePaths, change.Path)
			}
			if target == "journaled" {
				return interrupted
			}
		}
		if target == "before-commit" && stage == userInstallBeforeCommit {
			return interrupted
		}
		if stage == userInstallApplied {
			if index < 0 || index >= len(changePaths) {
				t.Fatalf("applied checkpoint lacks actual journal change: %d", index)
			}
			desired := map[string]string{"legacy-hooks": f.legacyHooks, "home-hooks": filepath.Join(f.options.Home, "hooks.json")}[target]
			if changePaths[index] == desired {
				return interrupted
			}
		}
		return nil
	})
	if !errors.Is(err, interrupted) {
		t.Fatalf("did not interrupt actual %s prefix: %v", target, err)
	}
	readUserTransactionJournal(t, f.options.Home)
	return interrupted
}
func assertCompletedUserMigration(t *testing.T, f userTransactionFixture, database []byte) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(f.options.Home, "agenticlens-user-install.pending.json")); !os.IsNotExist(err) {
		t.Fatalf("completed migration retains unresolved journal: %v", err)
	}
	if !bytes.Equal(database, mustReadUserInstall(t, f.pendingDB)) {
		t.Fatal("native migration changed retained capture database")
	}
	hooks := mustReadUserInstall(t, f.legacyHooks)
	if bytes.Contains(hooks, []byte("owned prior native callback")) || !bytes.Contains(hooks, []byte("foreign project callback")) {
		t.Fatalf("explicit migration failed ownership boundary: %s", hooks)
	}
	for _, path := range []string{f.legacyAnchor, f.legacyState} {
		var legacy Manifest
		if err := json.Unmarshal(mustReadUserInstall(t, path), &legacy); err != nil {
			t.Fatal(err)
		}
		if legacy.Active {
			t.Fatalf("migration left legacy ownership active: %s", path)
		}
	}
	home := mustReadUserInstall(t, filepath.Join(f.options.Home, "hooks.json"))
	if !bytes.Contains(home, []byte("foreign home callback")) || bytes.Count(home, []byte(userHookMarker)) != 6 {
		t.Fatalf("global migration lost foreign content or duplicated owned hooks: %s", home)
	}

}

func TestUserInstallationDurablePrefixesRecoverExplicitMigration(t *testing.T) {
	for _, target := range []string{"journaled", "legacy-hooks", "home-hooks", "before-commit"} {
		t.Run(target, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			before := snapshotUserArtifacts(t, f.paths)
			database := mustReadUserInstall(t, f.pendingDB)
			stopUserTransactionAt(t, f, target)
			if target != "journaled" && reflect.DeepEqual(before, snapshotUserArtifacts(t, f.paths)) {
				t.Fatal("interrupted prefix applied no actual owned mutation")
			}
			recovered := false
			stop := errors.New("inspect recovered before new mutation")
			_, err := installUserCodexWithCheckpoint(context.Background(), f.options, func(stage userInstallStage, _ int) error {
				if stage == userInstallJournaled {
					recovered = true
					if got := snapshotUserArtifacts(t, f.paths); !reflect.DeepEqual(before, got) {
						t.Fatalf("recovery failed to restore exact before-images: before=%+v after=%+v", before, got)
					}
					return stop
				}
				return nil
			})
			if !recovered || !errors.Is(err, stop) {
				t.Fatalf("restart did not recover before next transaction: %v", err)
			}
			if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
				t.Fatalf("resumed migration: %v", err)
			}
			assertCompletedUserMigration(t, f, database)
		})
	}
}

func TestInterruptedUserMigrationPreservesEditAndResumesAfterResolution(t *testing.T) {
	f := newUserTransactionFixture(t)
	database := mustReadUserInstall(t, f.pendingDB)
	stopUserTransactionAt(t, f, "home-hooks")
	path := filepath.Join(f.options.Home, "hooks.json")
	transactionBytes := mustReadUserInstall(t, path)
	edited := append(append([]byte{}, transactionBytes...), []byte("\nuser edit after interruption\n")...)
	if err := os.WriteFile(path, edited, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallUserCodex(context.Background(), f.options); err == nil || !strings.Contains(err.Error(), "preserv") {
		t.Fatalf("recovery silently overwrote later user edit: %v", err)
	}
	if !bytes.Equal(edited, mustReadUserInstall(t, path)) || !bytes.Equal(database, mustReadUserInstall(t, f.pendingDB)) {
		t.Fatal("conflicting recovery changed user edit or pending capture state")
	}
	readUserTransactionJournal(t, f.options.Home)
	// Resolve only this known test-owned conflict; the same durable journal resumes.
	if err := os.WriteFile(path, transactionBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
		t.Fatalf("resolved conflict remained stuck: %v", err)
	}
	assertCompletedUserMigration(t, f, database)
}

func TestRootRecoveryLocksInterruptedHomeEvenWhenCurrentHomeDiffers(t *testing.T) {
	for _, legacyShape := range []bool{false, true} {
		for _, resource := range []string{"previous-home", "legacy-project-home", "legacy-state"} {
			t.Run(fmt.Sprintf("old-shape-%t/%s", legacyShape, resource), func(t *testing.T) {
				f := newUserTransactionFixture(t)
				database := mustReadUserInstall(t, f.pendingDB)
				stopUserTransactionAt(t, f, "home-hooks")
				journalPath := filepath.Join(f.options.Home, "agenticlens-user-install.pending.json")
				if legacyShape {
					journal := readUserTransactionJournal(t, f.options.Home)
					journal.Homes = nil
					body, err := json.Marshal(journal)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(journalPath, body, 0600); err != nil {
						t.Fatal(err)
					}
				}
				journalBefore := mustReadUserInstall(t, journalPath)
				before := snapshotUserArtifacts(t, f.paths)
				lockHome := f.options.Home
				if resource == "legacy-project-home" {
					lockHome = filepath.Join(f.project, ".codex")
				}
				if resource == "legacy-state" {
					lockHome = f.state
				}
				lock, err := acquireInstallLock(lockHome)
				if err != nil {
					t.Fatal(err)
				}
				other := f.options
				other.Home = filepath.Join(t.TempDir(), "different-current-home")
				_, installErr := InstallUserCodex(context.Background(), other)
				if installErr == nil || !strings.Contains(installErr.Error(), "progress") {
					lock.Close()
					t.Fatalf("recovery ignored affected %s lock: %v", resource, installErr)
				}
				if !reflect.DeepEqual(before, snapshotUserArtifacts(t, f.paths)) || !bytes.Equal(journalBefore, mustReadUserInstall(t, journalPath)) {
					lock.Close()
					t.Fatal("busy recovery restored bytes before acquiring every affected-home lock")
				}
				if err := lock.Close(); err != nil {
					t.Fatal(err)
				}
				if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
					t.Fatalf("unlocked original recovery failed: %v", err)
				}
				assertCompletedUserMigration(t, f, database)
			})
		}
	}
}

func TestUserRecoveryRejectsResourcesAddedAfterDiscovery(t *testing.T) {
	f := newUserTransactionFixture(t)
	second := newUserTransactionFixture(t)
	second.options.Home = filepath.Dir(f.legacyHooks)
	otherHome := filepath.Dir(second.legacyHooks)
	journalPath := filepath.Join(second.options.Home, "agenticlens-user-install.pending.json")
	paths := []string{second.legacyHooks, second.legacyAnchor, second.legacyState, filepath.Join(second.options.Home, "hooks.json"), filepath.Join(second.options.Home, "agenticlens-user-installation.json")}
	var lease *PublicationLock
	defer func() {
		if lease != nil {
			lease.Close()
		}
	}()
	var journalBefore []byte
	var before map[string]userArtifactSnapshot
	scheduled := false
	_, installErr := installUserCodexWithCheckpoint(context.Background(), f.options, func(stage userInstallStage, _ int) error {
		if stage != userInstallResourcesDiscovered || scheduled {
			return nil
		}
		scheduled = true
		interrupted := errors.New("second real installer interrupted after newly discovered legacy hooks")
		_, err := installUserCodexWithCheckpoint(context.Background(), second.options, func(stage userInstallStage, index int) error {
			if stage == userInstallApplied {
				journal := readUserTransactionJournal(t, second.options.Home)
				if journal.Changes[index].Path == second.legacyHooks {
					return interrupted
				}
			}
			return nil
		})
		if !errors.Is(err, interrupted) {
			t.Fatalf("second actual installation did not reach durable prefix: %v", err)
		}
		journalBefore = mustReadUserInstall(t, journalPath)
		before = snapshotUserArtifacts(t, paths)
		lease, err = acquireInstallLock(otherHome)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if !scheduled || installErr == nil || !strings.Contains(installErr.Error(), "retry") {
		t.Fatalf("stale resource discovery did not require safe retry: scheduled=%t err=%v", scheduled, installErr)
	}
	if !reflect.DeepEqual(before, snapshotUserArtifacts(t, paths)) || !bytes.Equal(journalBefore, mustReadUserInstall(t, journalPath)) {
		t.Fatal("recovery changed newly discovered resources without their lease")
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	lease = nil
	if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
		t.Fatalf("fresh discovery failed to recover and resume: %v", err)
	}
	if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
		t.Fatalf("resumed recovery retained old journal: %v", err)
	}
	if hooks := mustReadUserInstall(t, filepath.Join(f.options.Home, "hooks.json")); bytes.Count(hooks, []byte(userHookMarker)) != 6 {
		t.Fatalf("resumed home installation missing owned hooks: %s", hooks)
	}
}
