package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

const (
	stagedPendingDirectory = ".review/pending [captures]"
	stagedAwaitingTicket   = "awaiting a ticket"
	stagedAwaitingTarget   = "awaiting a publication destination"
)

func stagedPendingGit(t *testing.T, project string, args ...string) []byte {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", project, "--literal-pathspecs", "-c", "commit.gpgsign=false"}, args...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return output
}

func writeStagedPendingFile(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
}

func newStagedPendingAdvisoryFixture(t *testing.T) captureGenerationFixture {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	policy := bytes.Replace(readCaptureFile(t, shared), []byte(captureFixturePendingDir), []byte(fmt.Sprintf("%q", stagedPendingDirectory)), 1)
	writeStagedPendingFile(t, shared, policy)
	prepareCaptureInstall(t, f.project, []string{"install"})
	stagedPendingGit(t, f.project, "config", "core.hooksPath", filepath.Join(t.TempDir(), "shared hooks"))
	output, err := captureCLI(t, f.project, "", "install-check", "--repo", f.project, "--wrapper", fixtureWrapperFromHook(t, f.project))
	if err != nil {
		t.Fatalf("install real advisory hook: %v\n%s", err, output)
	}
	return f
}

func registerUnboundStagedCapture(t *testing.T, f captureGenerationFixture) string {
	t.Helper()
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	source := f.write(t, 17)
	if _, err := capture.ReceiveCodexUser(ctx, store, capture.HookInput{
		SessionID: "root", Event: "SessionEnd", WorkingDirectory: f.project, TranscriptPath: source,
	}, f.store, cliBinaryPath, f.state); err != nil {
		t.Fatal(err)
	}
	return capture.StableCaptureID(domain.SourceCodex, "root")
}

func assertStagedPendingStates(t *testing.T, output []byte, ticket, destination bool) {
	t.Helper()
	body := strings.ToLower(string(output))
	if strings.Contains(body, stagedAwaitingTicket) != ticket || strings.Contains(body, stagedAwaitingTarget) != destination {
		t.Fatalf("expected pending categories ticket=%t destination=%t, got %q", ticket, destination, body)
	}
	if (ticket || destination) && (!strings.Contains(body, "review or unstage") || len(bytes.TrimSpace(output)) > 512 || bytes.Count(output, []byte("\n")) != 1) {
		t.Fatalf("expected a bounded, actionable advisory, got %q", body)
	}
	if ticket && !strings.Contains(body, "bind") || destination && !strings.Contains(body, "capture diagnostics") {
		t.Fatalf("expected a next step for the observed pending state, got %q", body)
	}
}

// writePortableStagedSnapshot supplies the persisted contract directly. The
// producer and discard-cleanup integration are exercised in their own suites.
func writePortableStagedSnapshot(t *testing.T, project, session, ticket string) string {
	t.Helper()
	snapshot := adapters.PendingSnapshot{
		Version:       1,
		CaptureID:     capture.StableCaptureID(domain.SourceCodex, session),
		Source:        domain.SourceCodex,
		RootSessionID: session,
		SourceCutoff:  "2026-01-01T00:00:02Z",
		State:         "awaiting_ticket",
		Sessions: []adapters.SessionFactDTO{{
			ID: session, Source: string(domain.SourceCodex), User: "git advisory fixture",
			Invocations: []adapters.InvocationFactDTO{}, Tools: []adapters.SessionToolFactDTO{},
		}},
	}
	if ticket != "" {
		snapshot.State = "awaiting_destination"
		snapshot.TicketID = ticket
		snapshot.TicketTitle = "Fixture ticket"
		snapshot.TicketType = "feat"
	}
	body, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(stagedPendingDirectory, snapshot.CaptureID+".json")
	writeStagedPendingFile(t, filepath.Join(project, path), body)
	return path
}

func addMalformedStagedSnapshot(t *testing.T, project string) string {
	t.Helper()
	path := filepath.Join(stagedPendingDirectory, capture.StableCaptureID(domain.SourceCodex, "malformed")+".json")
	writeStagedPendingFile(t, filepath.Join(project, path), []byte("{"))
	stagedPendingGit(t, project, "add", "--", path)
	return path
}

func TestCaptureStagedPendingAdvisoryUsesCommitIndex(t *testing.T) {
	f := newStagedPendingAdvisoryFixture(t)
	git := func(args ...string) []byte {
		t.Helper()
		return stagedPendingGit(t, f.project, args...)
	}
	registerUnboundStagedCapture(t, f)
	prepareCaptureOwner(t, f, "bound", "TEST-200", 23)
	unbound := writePortableStagedSnapshot(t, f.project, "root", "")
	bound := writePortableStagedSnapshot(t, f.project, "bound", "TEST-200")
	unboundPath := filepath.Join(f.project, unbound)
	boundPath := filepath.Join(f.project, bound)
	unboundBody := readCaptureFile(t, unboundPath)
	boundBody := readCaptureFile(t, boundPath)

	// The initial commit has no parent to diff against. A removed working copy
	// has no bearing on the staged blob, and the other snapshot is unstaged.
	git("symbolic-ref", "HEAD", "refs/heads/initial-pending")
	git("add", "--", ".agenticlens-config.yaml", unbound)
	if err := os.Remove(unboundPath); err != nil {
		t.Fatal(err)
	}
	indexPath := strings.TrimSpace(string(git("rev-parse", "--path-format=absolute", "--git-path", "index")))
	indexBefore := readCaptureFile(t, indexPath)
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), true, false)
	// The current filesystem can stop representing this directory altogether.
	// Its index path and bytes remain valid in both cases.
	pending := filepath.Join(f.project, stagedPendingDirectory)
	displaced := pending + "-displaced"
	if err := os.Rename(pending, displaced); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), pending); err != nil {
		t.Fatal(err)
	}
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), true, false)
	if err := os.Remove(pending); err != nil {
		t.Fatal(err)
	}
	writeStagedPendingFile(t, pending, []byte("replaced by a regular file"))
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), true, false)
	if err := os.Remove(pending); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(displaced, pending); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(indexBefore, readCaptureFile(t, indexPath)) {
		t.Fatal("advisory changed the raw index")
	}
	assertStagedPendingStates(t, git("commit", "-qm", "initial snapshot"), true, false)
	if committed := git("show", "HEAD:"+unbound); !bytes.Equal(committed, unboundBody) {
		t.Fatal("initial commit did not retain the exact staged snapshot")
	}

	// Reformat the snapshots for distinct Git objects with the same facts and
	// pending states.
	git("add", "--", bound)
	git("commit", "--no-verify", "-qm", "previous bound snapshot")
	health := git("hook", "run", "pre-commit")
	assertStagedPendingStates(t, health, false, false)
	if !bytes.Contains(health, []byte("no readable metrics")) {
		t.Fatalf("expected the fixture's standing capture-health warning, got %q", health)
	}

	malformed := addMalformedStagedSnapshot(t, f.project)
	incomplete := git("hook", "run", "pre-commit")
	assertStagedPendingStates(t, incomplete, false, false)
	for _, warning := range []string{"inspection incomplete", "no readable metrics"} {
		if !bytes.Contains(bytes.ToLower(incomplete), []byte(warning)) || bytes.Count(incomplete, []byte("\n")) != 1 {
			t.Fatalf("incomplete staged inspection hid hook warning %q: %q", warning, incomplete)
		}
	}
	git("rm", "--cached", "--", malformed)

	updatedBound := append([]byte("\n"), boundBody...)
	writeStagedPendingFile(t, boundPath, updatedBound)
	writeStagedPendingFile(t, unboundPath, append([]byte("\n"), unboundBody...))
	// Operate on a real alternate commit object database. The bound update has
	// never been written to the repository's default object directory.
	objectID := strings.TrimSpace(string(git("hash-object", "--", bound)))
	mainObjects := strings.TrimSpace(string(git("rev-parse", "--path-format=absolute", "--git-path", "objects")))
	alternateObjects := t.TempDir()
	t.Setenv("GIT_OBJECT_DIRECTORY", alternateObjects)
	t.Setenv("GIT_ALTERNATE_OBJECT_DIRECTORIES", mainObjects)
	git("add", "--", unbound, bound)
	assertInitialPublicationAbsent(t, filepath.Join(mainObjects, objectID[:2], objectID[2:]))
	if _, err := os.Stat(filepath.Join(alternateObjects, objectID[:2], objectID[2:])); err != nil {
		t.Fatalf("bound update did not enter the alternate object database: %v", err)
	}
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), true, true)

	// A path-limited Git commit supplies a temporary index containing the bound
	// working copy and excluding the already staged unbound update.
	git("restore", "--staged", "--", bound)
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), true, false)
	assertStagedPendingStates(t, git("commit", "-qm", "bound snapshot only", "--", bound), false, true)
	if committed := git("show", "HEAD:"+bound); !bytes.Equal(committed, updatedBound) {
		t.Fatal("path-limited commit did not contain the expected bound snapshot")
	}
	if remaining := git("diff", "--cached", "--name-only", "-z"); string(remaining) != unbound+"\x00" {
		t.Fatalf("path-limited commit changed unrelated staging: %q", remaining)
	}

	// A staged deletion and an unstaged edit cannot introduce a snapshot.
	git("restore", "--staged", "--", unbound)
	git("rm", "--", bound)
	assertStagedPendingStates(t, git("hook", "run", "pre-commit"), false, false)
}

func TestCaptureStagedPendingAdvisoryScopesDiscardAndSurvivesMissingState(t *testing.T) {
	f := newStagedPendingAdvisoryFixture(t)
	id := registerUnboundStagedCapture(t, f)
	prepareCaptureOwner(t, f, "unrelated", "TEST-200", 23)
	relative := writePortableStagedSnapshot(t, f.project, "root", "")
	unrelated := writePortableStagedSnapshot(t, f.project, "unrelated", "TEST-200")
	path := filepath.Join(f.project, relative)
	snapshot := readCaptureFile(t, path)
	stagedPendingGit(t, f.project, "add", "--", ".agenticlens-config.yaml")
	stagedPendingGit(t, f.project, "commit", "-qm", "shared advisory configuration")
	stagedPendingGit(t, f.project, "add", "--", relative, unrelated)
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, true)

	// Record a genuine discard in this worktree. Remove the worktree file to
	// model the post-cleanup state while retaining its exact staged bytes.
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx := context.Background()
	discard, err := store.SetCaptureDiscarded(ctx, id, true, nil)
	if err != nil || discard.State != capture.CaptureControlDiscarded {
		t.Fatalf("record durable discard: %+v, %v", discard, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	indexPath := strings.TrimSpace(string(stagedPendingGit(t, f.project, "rev-parse", "--path-format=absolute", "--git-path", "index")))
	indexBefore := readCaptureFile(t, indexPath)
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), false, true)
	if !bytes.Equal(indexBefore, readCaptureFile(t, indexPath)) || !bytes.Equal(stagedPendingGit(t, f.project, "show", ":"+relative), snapshot) {
		t.Fatal("discard suppression changed the index")
	}
	stagedPendingGit(t, f.project, "rm", "--cached", "--", unrelated)

	linked := filepath.Join(t.TempDir(), "linked worktree")
	stagedPendingGit(t, f.project, "worktree", "add", "-qb", "linked-pending", linked)
	linkedFixture := captureGenerationFixture{project: linked, stateBase: f.stateBase}
	if err := writeCaptureFixtureLocal(t, linkedFixture, []byte("telemetry:\n  capture:\n    enabled: false\n")); err != nil {
		t.Fatal(err)
	}
	linkedStore, err := capture.OpenProject(f.stateBase, linked)
	if err != nil {
		t.Fatal(err)
	}
	sameRegistry := linkedStore.RegistryDirectory() == store.RegistryDirectory()
	if err := linkedStore.Close(); err != nil {
		t.Fatal(err)
	}
	if !sameRegistry {
		t.Fatal("linked worktree did not enroll in the same capture registry")
	}
	writeStagedPendingFile(t, filepath.Join(linked, relative), snapshot)
	stagedPendingGit(t, linked, "add", "-N", "--", relative)
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), false, false)
	stagedPendingGit(t, linked, "add", "--", relative)
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), true, false)

	// The same staged capture remains meaningful when its selected private
	// state and native source are unavailable.
	missingState := filepath.Join(t.TempDir(), "missing capture state")
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "missing native source"))
	linkedFixture.stateBase = missingState
	if err := writeCaptureFixtureLocal(t, linkedFixture, []byte("telemetry:\n  capture:\n    enabled: false\n")); err != nil {
		t.Fatal(err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), true, false)
	assertInitialPublicationAbsent(t, missingState)

	invalidState := filepath.Join(t.TempDir(), "loop")
	if err := os.Symlink("loop", invalidState); err != nil {
		t.Fatal(err)
	}
	linkedFixture.stateBase = invalidState
	if err := writeCaptureFixtureLocal(t, linkedFixture, []byte("telemetry:\n  capture:\n    enabled: false\n")); err != nil {
		t.Fatal(err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), true, false)
	linkedFixture.stateBase = missingState
	if err := writeCaptureFixtureLocal(t, linkedFixture, []byte("telemetry:\n  capture:\n    enabled: false\n")); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(linked, ".agenticlens-config.local.yaml")
	localBody := readCaptureFile(t, local)
	stateLine := []byte(fmt.Sprintf("    state_dir: '%s'\n", strings.ReplaceAll(missingState, "'", "''")))
	withoutState := bytes.Replace(localBody, stateLine, nil, 1)
	if bytes.Equal(localBody, withoutState) {
		t.Fatal("fixture could not remove the explicit local state directory")
	}
	writeStagedPendingFile(t, local, withoutState)
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), true, false)
	writeStagedPendingFile(t, local, localBody)

	// An unreadable candidate must leave valid staged evidence available.
	malformed := addMalformedStagedSnapshot(t, linked)
	incomplete := stagedPendingGit(t, linked, "hook", "run", "pre-commit")
	assertStagedPendingStates(t, incomplete, true, false)
	if !bytes.Contains(bytes.ToLower(incomplete), []byte("inspection incomplete")) {
		t.Fatalf("malformed staged snapshot was silently skipped: %q", incomplete)
	}
	stagedPendingGit(t, linked, "rm", "--cached", "--", malformed)

	// A missing required setting is visible and must not fall back to the
	// conventional directory, even when that directory has a staged snapshot.
	shared := filepath.Join(linked, ".agenticlens-config.yaml")
	config := readCaptureFile(t, shared)
	withoutPending := bytes.Replace(config, []byte(fmt.Sprintf("  pending_dir: %q\n", stagedPendingDirectory)), nil, 1)
	writeStagedPendingFile(t, shared, withoutPending)
	fallback := filepath.Join(captureFixturePendingDir, id+".json")
	writeStagedPendingFile(t, filepath.Join(linked, fallback), snapshot)
	stagedPendingGit(t, linked, "add", "--", fallback)
	output := stagedPendingGit(t, linked, "hook", "run", "pre-commit")
	assertStagedPendingStates(t, output, false, false)
	if !bytes.Contains(output, []byte("agenticlens.pending_dir")) {
		t.Fatalf("incomplete pending inspection was not explained: %q", output)
	}
	writeStagedPendingFile(t, shared, config)
	stagedPendingGit(t, linked, "rm", "--cached", "--", relative)
	assertStagedPendingStates(t, stagedPendingGit(t, linked, "hook", "run", "pre-commit"), false, false)
	stagedPendingGit(t, linked, "add", "--", relative)
	if err := writeCaptureFixtureLocal(t, linkedFixture, []byte("telemetry:\n  capture:\n    enabled: false\n    commit_warning: false\n")); err != nil {
		t.Fatal(err)
	}
	if output := stagedPendingGit(t, linked, "hook", "run", "pre-commit"); len(output) != 0 {
		t.Fatalf("disabled warning ran against staged snapshots: %q", output)
	}

	// A second store can report the same native ID before establishing that its
	// source belongs to this worktree. Only an accepted association introduces
	// ambiguity, and losing the original source doesn't erase that association.
	competing := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
	transcript := filepath.Join(competing.Store, "sessions", "root.jsonl")
	writeStagedPendingFile(t, transcript, readCaptureFile(t, filepath.Join(f.store, "sessions", "root.jsonl")))
	row, err := store.Register(ctx, competing, transcript, "SessionEnd")
	if err != nil || row.ProjectDirectory != "" {
		t.Fatalf("fixture did not retain an unassociated competing handoff: %+v, %v", row, err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), false, false)
	t.Run("corrupt competing registration", func(t *testing.T) {
		db, err := sql.Open("sqlite", filepath.Join(store.RegistryDirectory(), "registry.db"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		})
		var original []byte
		if err := db.QueryRowContext(ctx, "SELECT record FROM registrations WHERE source=? AND store=? AND session_id=?", competing.Source, competing.Store, competing.SessionID).Scan(&original); err != nil {
			t.Fatal(err)
		}
		// SQLite length(TEXT) counts characters, so this valid competing row
		// exceeds the byte budget without exceeding the character count.
		if _, err := db.ExecContext(ctx, "UPDATE registrations SET record=json_set(CAST(? AS TEXT),'$.registration.detail',?) WHERE source=? AND store=? AND session_id=?", original, strings.Repeat("é", 600000), competing.Source, competing.Store, competing.SessionID); err != nil {
			t.Fatal(err)
		}
		assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
		if _, err := db.ExecContext(ctx, "UPDATE registrations SET record=CAST(json_set(CAST(? AS TEXT),'$.registration.identity',42) AS BLOB) WHERE source=? AND store=? AND session_id=?", original, competing.Source, competing.Store, competing.SessionID); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := db.ExecContext(ctx, "UPDATE registrations SET record=? WHERE source=? AND store=? AND session_id=?", original, competing.Source, competing.Store, competing.SessionID); err != nil {
				t.Error(err)
			}
		})
		if control, err := store.CaptureControl(ctx, id); err == nil {
			t.Fatalf("corrupted registration should make exact discard status unavailable: %+v", control)
		}
		assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
		if _, err := db.ExecContext(ctx, "UPDATE registrations SET record=CAST(json_remove(CAST(? AS TEXT),'$.registration.identity') AS BLOB) WHERE source=? AND store=? AND session_id=?", original, competing.Source, competing.Store, competing.SessionID); err != nil {
			t.Fatal(err)
		}
		assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
	})
	row, err = store.ProveRegistration(ctx, row)
	if err != nil || row.ProjectDirectory != f.project || row.ProjectProof != domain.CaptureProjectVerified {
		t.Fatalf("fixture did not prove the competing association: %+v, %v", row, err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
	if err := os.Remove(transcript); err != nil {
		t.Fatal(err)
	}
	row, err = store.ProveRegistration(ctx, row)
	if err != nil || row.ProjectDirectory != f.project || row.ProjectProof != domain.CaptureProjectUnverified {
		t.Fatalf("fixture did not retain the previously proved association after source loss: %+v, %v", row, err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
}

func TestCaptureStagedPendingAdvisoryPreservesPhysicalParentTraversal(t *testing.T) {
	f := newStagedPendingAdvisoryFixture(t)
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	policy := bytes.Replace(readCaptureFile(t, shared), []byte(fmt.Sprintf("%q", stagedPendingDirectory)), []byte(`"logical/link/../pending"`), 1)
	writeStagedPendingFile(t, shared, policy)

	fixture := writePortableStagedSnapshot(t, f.project, "portable", "")
	snapshot := readCaptureFile(t, filepath.Join(f.project, fixture))
	name := filepath.Base(fixture)
	lexical := filepath.Join("logical", "pending", name)
	physical := filepath.Join("physical", "pending", name)
	writeStagedPendingFile(t, filepath.Join(f.project, lexical), snapshot)
	writeStagedPendingFile(t, filepath.Join(f.project, physical), snapshot)
	child := filepath.Join(f.project, "physical", "child")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(child, filepath.Join(f.project, "logical", "link")); err != nil {
		t.Fatal(err)
	}

	stagedPendingGit(t, f.project, "add", "--", lexical)
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), false, false)
	stagedPendingGit(t, f.project, "add", "--", physical)
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)

	// Private state can also traverse a link before its parent. The lexical
	// path overlaps this staged directory, while the physical path does not.
	privateChild := filepath.Join(t.TempDir(), "child")
	if err := os.Mkdir(privateChild, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(privateChild, filepath.Join(f.project, "physical", "state-link")); err != nil {
		t.Fatal(err)
	}
	if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    enabled: false\n    state_dir: 'physical/state-link/../pending'\n")); err != nil {
		t.Fatal(err)
	}
	assertStagedPendingStates(t, stagedPendingGit(t, f.project, "hook", "run", "pre-commit"), true, false)
}
