package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func coordinatorRecordPath(directory string) string {
	return filepath.Join(directory, userInstallPendingFile)
}

func stopCoordinatedUserMigration(t *testing.T, f userTransactionFixture, prefix string) {
	t.Helper()
	stop := errors.New("interrupted actual coordinated migration prefix")
	coordinator := filepath.Dir(f.legacyHooks)
	var changePaths []string
	_, err := installUserCodexWithCheckpoint(context.Background(), f.options, func(stage userInstallStage, index int) error {
		if stage == userInstallReferenced {
			record, exists, err := readPendingUserRecord(coordinator)
			if err != nil || !exists || record.Kind != userInstallReferenceKind || record.Reference.OwnerHome != f.options.Home || record.Reference.Coordinator != coordinator {
				t.Fatalf("reference checkpoint lacks typed owner record: %+v %t %v", record, exists, err)
			}
			var wire map[string]json.RawMessage
			if err := json.Unmarshal(mustReadUserInstall(t, coordinatorRecordPath(coordinator)), &wire); err != nil {
				t.Fatal(err)
			}
			if len(wire["changes"]) != 0 || len(wire["homes"]) != 0 {
				t.Fatal("shared migration reference contains private hook or undo payload")
			}
			if _, err := os.Stat(coordinatorRecordPath(f.options.Home)); !os.IsNotExist(err) {
				t.Fatalf("private canonical journal existed before reference checkpoint: %v", err)
			}
			if prefix == "reference-only" {
				return stop
			}
		}
		if stage == userInstallJournaled {
			journal, actualCoordinator, err := readUserInstallJournal(f.options.Home)
			if err != nil || actualCoordinator != f.options.Home || journal.OwnerHome != f.options.Home || journal.Coordinator != coordinator || journal.Kind != userInstallJournalKind {
				t.Fatalf("canonical checkpoint lacks owned production journal: %+v %q %v", journal, actualCoordinator, err)
			}
			for _, change := range journal.Changes {
				changePaths = append(changePaths, change.Path)
			}
			if prefix == "canonical" {
				return stop
			}
		}
		if stage == userInstallApplied {
			if index < 0 || index >= len(changePaths) {
				t.Fatalf("business checkpoint lacks production journal index: %d", index)
			}
			if prefix == "legacy-hooks" && changePaths[index] == f.legacyHooks || prefix == "home-hooks" && changePaths[index] == filepath.Join(f.options.Home, "hooks.json") {
				return stop
			}
		}
		if prefix == "before-commit" && stage == userInstallBeforeCommit {
			return stop
		}
		if prefix == "committed" && stage == userInstallCommitted {
			if _, err := os.Stat(coordinatorRecordPath(f.options.Home)); !os.IsNotExist(err) {
				t.Fatalf("commit checkpoint retained rollback authority: %v", err)
			}
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("did not reach actual %s prefix: %v", prefix, err)
	}
}

// Call the recovery boundary with the fixture's complete known resource leases.
// Resource discovery/race behavior remains covered through the real installer.
func recoverUserCoordinatorWithLeases(t *testing.T, home string, resources ...string) error {
	t.Helper()
	sort.Strings(resources)
	seen := map[string]bool{}
	for _, resource := range resources {
		if seen[resource] {
			continue
		}
		seen[resource] = true
		lease, err := acquireInstallLock(resource)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := lease.Close(); err != nil {
				t.Error(err)
			}
		}()
	}
	return recoverUserInstallation(home)
}

func coordinatorHomeFixture(t *testing.T, f userTransactionFixture) userTransactionFixture {
	t.Helper()
	other := f
	other.options.Home = filepath.Join(t.TempDir(), "second-home")
	if err := os.Mkdir(other.options.Home, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other.options.Home, "hooks.json"), []byte(`{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"foreign home callback from second home"}]}]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	return other
}

func TestUserMigrationCoordinatorCrashPrefixesPreserveCommitBoundary(t *testing.T) {
	for _, prefix := range []string{"reference-only", "canonical", "legacy-hooks", "home-hooks", "before-commit", "committed"} {
		t.Run(prefix, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			before := snapshotUserArtifacts(t, f.paths)
			database := mustReadUserInstall(t, f.pendingDB)
			stopCoordinatedUserMigration(t, f, prefix)
			afterPrefix := snapshotUserArtifacts(t, f.paths)
			expected := before
			if prefix == "committed" {
				expected = afterPrefix
			}
			if prefix != "reference-only" && prefix != "canonical" && reflect.DeepEqual(before, afterPrefix) {
				t.Fatal("selected business prefix applied no actual change")
			}
			recoveryContext := f.options.Home
			if prefix == "reference-only" || prefix == "committed" {
				recoveryContext = filepath.Dir(f.legacyHooks)
			}
			if err := recoverUserCoordinatorWithLeases(t, recoveryContext, f.options.Home, filepath.Dir(f.legacyHooks), f.state); err != nil {
				t.Fatal(err)
			}
			if got := snapshotUserArtifacts(t, f.paths); !reflect.DeepEqual(got, expected) {
				t.Fatalf("recovery crossed commit boundary: prefix=%s got=%+v want=%+v", prefix, got, expected)
			}
			for _, directory := range []string{f.options.Home, filepath.Dir(f.legacyHooks)} {
				if _, err := os.Stat(coordinatorRecordPath(directory)); !os.IsNotExist(err) {
					t.Fatalf("recovery retained private transaction record: %s %v", directory, err)
				}
			}
			if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
				t.Fatal(err)
			}
			assertCompletedUserMigration(t, f, database)
		})
	}
}

func TestUserMigrationDifferentHomeRecoversCanonicalOwnerBeforeCommit(t *testing.T) {
	f := newUserTransactionFixture(t)
	database := mustReadUserInstall(t, f.pendingDB)
	beforeA := snapshotUserArtifacts(t, []string{filepath.Join(f.options.Home, "hooks.json"), filepath.Join(f.options.Home, "agenticlens-user-installation.json")})
	stopCoordinatedUserMigration(t, f, "home-hooks")
	other := coordinatorHomeFixture(t, f)
	if _, err := InstallUserCodex(context.Background(), other.options); err != nil {
		t.Fatalf("different home did not recover canonical owner: %v", err)
	}
	if got := snapshotUserArtifacts(t, []string{filepath.Join(f.options.Home, "hooks.json"), filepath.Join(f.options.Home, "agenticlens-user-installation.json")}); !reflect.DeepEqual(got, beforeA) {
		t.Fatal("second home committed without restoring the interrupted first home")
	}
	protected := append(append([]string{}, f.paths...), filepath.Join(other.options.Home, "hooks.json"), filepath.Join(other.options.Home, "agenticlens-user-installation.json"))
	committed := snapshotUserArtifacts(t, protected)
	if err := recoverUserCoordinatorWithLeases(t, f.options.Home, f.options.Home); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(committed, snapshotUserArtifacts(t, protected)) || !bytes.Equal(database, mustReadUserInstall(t, f.pendingDB)) {
		t.Fatal("later first-home recovery undid another home's committed migration")
	}
	assertCompletedUserMigration(t, other, database)
}

func TestUserMigrationStaleReferenceCannotRecoverAnotherOwnersJournal(t *testing.T) {
	f := newUserTransactionFixture(t)
	stopCoordinatedUserMigration(t, f, "committed")
	committedA := snapshotUserArtifacts(t, f.paths)
	// The same private owner later starts a different legacy migration. The old
	// shared reference must not acquire authority over this new private journal.
	other := newUserTransactionFixture(t)
	other.options.Home = f.options.Home
	otherPaths := []string{other.legacyHooks, other.legacyAnchor, other.legacyState, other.pendingDB, filepath.Join(f.options.Home, "hooks.json"), filepath.Join(f.options.Home, "agenticlens-user-installation.json")}
	beforeB := snapshotUserArtifacts(t, otherPaths)
	stopCoordinatedUserMigration(t, other, "legacy-hooks")
	protected := append(append(append([]string{}, f.paths...), otherPaths...), coordinatorRecordPath(f.options.Home), coordinatorRecordPath(filepath.Dir(other.legacyHooks)))
	beforeCleanup := snapshotUserArtifacts(t, protected)
	resources, err := pendingUserHomes(filepath.Dir(f.legacyHooks))
	if err != nil || len(resources) != 0 {
		t.Fatalf("stale reference enrolled another owner's recovery: %v %v", resources, err)
	}
	busy, err := acquireInstallLock(f.options.Home)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if busy != nil {
			if err := busy.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if err := recoverUserCoordinatorWithLeases(t, filepath.Dir(f.legacyHooks), filepath.Dir(f.legacyHooks)); err != nil {
		t.Fatalf("stale-reference cleanup required another owner: %v", err)
	}
	if !reflect.DeepEqual(beforeCleanup, snapshotUserArtifacts(t, protected)) {
		t.Fatal("stale reference changed another owner's canonical journal or business bytes")
	}
	caller := coordinatorHomeFixture(t, other)
	if _, err := InstallUserCodex(context.Background(), caller.options); err == nil || !strings.Contains(err.Error(), "progress") {
		t.Fatalf("explicit same-legacy migration bypassed canonical owner lease: %v", err)
	}
	if !reflect.DeepEqual(beforeCleanup, snapshotUserArtifacts(t, protected)) {
		t.Fatal("busy canonical recovery changed protected files")
	}
	if err := busy.Close(); err != nil {
		t.Fatal(err)
	}
	busy = nil
	if err := recoverUserCoordinatorWithLeases(t, f.options.Home, f.options.Home, filepath.Dir(other.legacyHooks), other.state); err != nil {
		t.Fatalf("leased canonical owner could not recover: %v", err)
	}
	if !reflect.DeepEqual(committedA, snapshotUserArtifacts(t, f.paths)) || !reflect.DeepEqual(beforeB, snapshotUserArtifacts(t, otherPaths)) {
		t.Fatal("canonical recovery failed to preserve A's commit and restore B's interrupted before-image")
	}
}

func TestUserMigrationCoordinatorCorruptionPreservesUnknownRecords(t *testing.T) {
	for _, corruption := range []string{"malformed reference", "edited reference owner", "edited reference coordinator", "malformed canonical", "edited canonical before-image"} {
		t.Run(corruption, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			stopCoordinatedUserMigration(t, f, "home-hooks")
			referencePath := coordinatorRecordPath(filepath.Dir(f.legacyHooks))
			canonicalPath := coordinatorRecordPath(f.options.Home)
			path := referencePath
			body := []byte("{invalid interrupted record")
			if corruption == "malformed canonical" {
				path = canonicalPath
			} else if corruption == "edited canonical before-image" {
				journal, _, err := readUserInstallJournal(f.options.Home)
				if err != nil {
					t.Fatal(err)
				}
				changed := false
				for index := range journal.Changes {
					if journal.Changes[index].Path == filepath.Join(f.options.Home, "hooks.json") {
						journal.Changes[index].Before = append(journal.Changes[index].Before, '\n')
						changed = true
					}
				}
				if !changed {
					t.Fatal("actual canonical journal lacks home before-image")
				}
				body, err = json.Marshal(journal)
				if err != nil {
					t.Fatal(err)
				}
				path = canonicalPath
			} else if corruption != "malformed reference" {
				record, exists, err := readPendingUserRecord(filepath.Dir(f.legacyHooks))
				if err != nil || !exists {
					t.Fatalf("actual reference missing: %t %v", exists, err)
				}
				if corruption == "edited reference owner" {
					record.Reference.OwnerHome = filepath.Join(t.TempDir(), "foreign-home")
				} else {
					record.Reference.Coordinator = filepath.Dir(f.options.Home)
				}
				body, err = json.Marshal(record.Reference)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			paths := append(append([]string{}, f.paths...), referencePath, canonicalPath)
			before := snapshotUserArtifacts(t, paths)
			if _, err := InstallUserCodex(context.Background(), f.options); err == nil {
				t.Fatal("unknown edited recovery record was accepted")
			}
			if !reflect.DeepEqual(before, snapshotUserArtifacts(t, paths)) {
				t.Fatal("corrupt-record recovery changed user bytes or pending capture facts")
			}
		})
	}
}

func TestUserMigrationDeletedCoordinatorRetainsPrivateUndo(t *testing.T) {
	for _, prefix := range []string{"reference-only", "home-hooks", "committed"} {
		t.Run(prefix, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			stopCoordinatedUserMigration(t, f, prefix)
			if err := os.RemoveAll(f.project); err != nil {
				t.Fatal(err)
			}
			paths := append(append([]string{}, f.paths...), coordinatorRecordPath(f.options.Home))
			before := snapshotUserArtifacts(t, paths)
			err := recoverUserCoordinatorWithLeases(t, f.options.Home, f.options.Home)
			if prefix == "home-hooks" {
				if err == nil {
					t.Fatal("deleted destination became successful pending recovery")
				}
				if _, err := os.Stat(coordinatorRecordPath(f.options.Home)); err != nil {
					t.Fatalf("deleted destination lost private undo images: %v", err)
				}
			} else if err != nil {
				t.Fatalf("no private rollback authority should preserve idle/committed bytes: %v", err)
			}
			if !reflect.DeepEqual(before, snapshotUserArtifacts(t, paths)) {
				t.Fatal("unknown deleted coordinator changed retained facts or reference")
			}
			if _, err := os.Stat(f.project); !os.IsNotExist(err) {
				t.Fatalf("recovery recreated deleted checkout: %v", err)
			}
		})
	}
}

func TestUserMigrationDeletedOwnerPreservesUnknownSharedReference(t *testing.T) {
	for _, prefix := range []string{"home-hooks", "committed"} {
		t.Run(prefix, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			stopCoordinatedUserMigration(t, f, prefix)
			if err := os.RemoveAll(f.options.Home); err != nil {
				t.Fatal(err)
			}
			coordinator := filepath.Dir(f.legacyHooks)
			paths := append(append([]string{}, f.paths...), coordinatorRecordPath(coordinator))
			before := snapshotUserArtifacts(t, paths)
			if err := recoverUserCoordinatorWithLeases(t, coordinator, coordinator); err == nil || !strings.Contains(err.Error(), "owner") {
				t.Fatalf("deleted private owner became a proven commit: %v", err)
			}
			if !reflect.DeepEqual(before, snapshotUserArtifacts(t, paths)) {
				t.Fatal("unknown owner deletion changed retained facts or shared reference")
			}
			if _, err := os.Stat(f.options.Home); !os.IsNotExist(err) {
				t.Fatalf("recovery recreated deleted private owner: %v", err)
			}
		})
	}
}

func TestUserMigrationPublicInstallPreservesDeletedOwner(t *testing.T) {
	for _, prefix := range []string{"home-hooks", "committed"} {
		t.Run(prefix, func(t *testing.T) {
			f := newUserTransactionFixture(t)
			f.options.Repository = f.project
			stopCoordinatedUserMigration(t, f, prefix)
			if err := os.RemoveAll(f.options.Home); err != nil {
				t.Fatal(err)
			}
			paths := append(append([]string{}, f.paths...), coordinatorRecordPath(filepath.Dir(f.legacyHooks)))
			before := snapshotUserArtifacts(t, paths)
			if _, err := InstallUserCodex(context.Background(), f.options); err == nil || !strings.Contains(err.Error(), "owner") {
				t.Fatalf("public installer manufactured a successful private owner: %v", err)
			}
			if _, err := os.Stat(f.options.Home); !os.IsNotExist(err) {
				t.Fatalf("public installer recreated unknown private authority: %v", err)
			}
			if !reflect.DeepEqual(before, snapshotUserArtifacts(t, paths)) {
				t.Fatal("public installer changed unknown recovery reference or pending facts")
			}
		})
	}
	t.Run("fresh home without pending authority", func(t *testing.T) {
		f := newUserTransactionFixture(t)
		f.options.Repository = f.project
		f.options.Home = filepath.Join(t.TempDir(), "new-private-home")
		database := mustReadUserInstall(t, f.pendingDB)
		if _, err := InstallUserCodex(context.Background(), f.options); err != nil {
			t.Fatalf("legitimate fresh home was rejected: %v", err)
		}
		var receipt userInstallation
		if err := json.Unmarshal(mustReadUserInstall(t, filepath.Join(f.options.Home, "agenticlens-user-installation.json")), &receipt); err != nil {
			t.Fatal(err)
		}
		if receipt.Protocol != userCaptureOwnershipProtocol || len(receipt.Groups) != 6 || bytes.Count(mustReadUserInstall(t, filepath.Join(f.options.Home, "hooks.json")), []byte(userHookMarker)) != 6 {
			t.Fatalf("fresh owner lacks complete native hook ownership: %+v", receipt)
		}
		if !bytes.Equal(database, mustReadUserInstall(t, f.pendingDB)) {
			t.Fatal("fresh home installation changed retained capture facts")
		}
		for _, directory := range []string{f.options.Home, filepath.Dir(f.legacyHooks)} {
			if _, err := os.Stat(coordinatorRecordPath(directory)); !os.IsNotExist(err) {
				t.Fatalf("successful fresh owner retains pending transaction: %s %v", directory, err)
			}
		}
	})
}
