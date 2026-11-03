package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCaptureVerifyNativeExchangeBelongsToCurrentInstallationAndSelector(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := f.write(t, 23)
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n", f.state)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, "", "install", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state); err != nil {
		t.Fatalf("install: %v %s", err, out)
	}
	verify := func(session, store string) domain.CaptureHealth {
		t.Helper()
		out, err := captureCLI(t, f.project, "", "verify", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", store, "--session-id", session)
		if err != nil {
			t.Fatalf("verify: %v %s", err, out)
		}
		var health domain.CaptureHealth
		if err := json.Unmarshal(out, &health); err != nil {
			t.Fatalf("verify JSON: %v %s", err, out)
		}
		return health
	}
	if h := verify("root", f.store); h.State != "verification_pending" || h.CallbackObserved {
		t.Fatalf("installation alone established callback health: %+v", h)
	}
	payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, string(payload), "receive", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state); err != nil {
		t.Fatalf("callback: %v %s", err, out)
	}
	request, err := captureCLI(t, f.project, "", "request", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--session-id", "root")
	if err != nil {
		t.Fatalf("binding request: %v %s", err, request)
	}
	var registration domain.CaptureRegistration
	if err := json.Unmarshal(request, &registration); err != nil || registration.Request == "" {
		t.Fatalf("exact binding request unavailable: %v %s", err, request)
	}
	if out, err := captureCLI(t, f.project, "", "bind", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--session-id", "root", "--ticket-id", "TEST-100", "--request", registration.Request); err != nil {
		t.Fatalf("explicit binding: %v %s", err, out)
	}
	if h := verify("root", f.store); h.State == "verified" || h.PublicationAccepted {
		t.Fatalf("binding without source publication established health: %+v", h)
	}
	f.generate(t)
	if h := verify("root", f.store); h.State != "verified" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
		t.Fatalf("actual current-installation exchange not verified: %+v", h)
	}
	for _, selection := range []struct{ session, store string }{{"other-root", f.store}, {"root", t.TempDir()}} {
		if h := verify(selection.session, selection.store); h.State == "verified" || h.CallbackObserved {
			t.Fatalf("unrelated selector inherited another session's health: %+v", h)
		}
	}
	manifest := filepath.Join(f.state, "install-codex.json")
	var current map[string]json.RawMessage
	if err := json.Unmarshal(readCaptureFile(t, manifest), &current); err != nil {
		t.Fatal(err)
	}
	current["source_store"], err = json.Marshal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, body, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, "", "request", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--session-id", "root"); err != nil {
		t.Fatalf("request after installation change: %v %s", err, out)
	}
	if h := verify("root", f.store); h.State == "verified" || h.CallbackObserved {
		t.Fatalf("new installation reused an old callback/source/publication chain: %+v", h)
	}
	// An observed installation problem remains actionable by default, while the
	// explicit machine-local advisory policy applies before drift inspection.
	anchor := filepath.Join(f.project, ".codex", "agenticlens-installation.json")
	var installed map[string]json.RawMessage
	if err := json.Unmarshal(readCaptureFile(t, anchor), &installed); err != nil {
		t.Fatal(err)
	}
	installed["template_version"] = json.RawMessage(`"stale-template"`)
	anchorBody, err := json.Marshal(installed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(anchor, anchorBody, 0600); err != nil {
		t.Fatal(err)
	}
	check := []string{"check", "--repo", f.project, "--state-dir", f.state}
	if out, err := captureCLI(t, f.project, "", check...); err != nil || !bytes.Contains(out, []byte("Owned capture installation is missing, edited or stale")) {
		t.Fatalf("installation drift control was not actionable: %v %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local+"    commit_warning: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, "", check...); err != nil || len(out) != 0 {
		t.Fatalf("explicit local advisory disable was bypassed by installation drift: %v %s", err, out)
	}
}

func TestCaptureSourceAcceptanceSurvivesFailedPublicationWithoutClaimingSuccess(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 10))
	summary := filepath.Join(f.project, "SUMMARY.md")
	f.generate(t, "--append", summary)
	beforeLedger, beforeSummary := readCaptureFile(t, f.output), readCaptureFile(t, summary)
	f.writeAt(t, 31, "2026-01-01T00:00:03Z")
	if err := os.Remove(summary); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(summary, 0700); err != nil {
		t.Fatal(err)
	}
	out, err := runCaptureNative(t, f.project, "", f.args(t, "--append", summary)...)
	if err == nil || !bytes.Contains(out, []byte("Cannot stage recoverable publication")) {
		t.Fatalf("publication obstruction was accepted: %v %s", err, out)
	}
	if !bytes.Equal(beforeLedger, readCaptureFile(t, f.output)) {
		t.Fatal("failed publication replaced accepted ledger")
	}
	diagnostics := captureDiagnostics(t, f.project, f.state)
	var lastSuccess, lastAttempt *domain.CaptureDiagnostic
	for i := range diagnostics {
		d := &diagnostics[i]
		if d.Phase != "publication" {
			continue
		}
		if d.Outcome == "committed" {
			lastSuccess = d
		}
		if lastAttempt == nil || d.Sequence > lastAttempt.Sequence {
			lastAttempt = d
		}
	}
	if lastSuccess == nil || lastAttempt == nil || lastAttempt.Outcome != "staging" || lastAttempt.Sequence <= lastSuccess.Sequence {
		t.Fatalf("unfinished publication overwrote/hid last success: %+v", diagnostics)
	}
	if len(lastSuccess.Owners) != 1 || lastSuccess.Owners[0].Checkpoint != "2026-01-01T00:00:02Z" || len(lastAttempt.Owners) != 1 || lastAttempt.Owners[0].Checkpoint != "2026-01-01T00:00:03Z" {
		t.Fatalf("source acceptance conflated with artifact commit: previous=%+v attempt=%+v", lastSuccess, lastAttempt)
	}
	store, err := capture.OpenReadOnlyProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := store.List(context.Background())
	if closeErr := store.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if err != nil || len(rows) != 1 {
		t.Fatalf("accepted source unavailable after restart: %+v %v", rows, err)
	}
	assertCaptureRootInput(t, domain.AgenticFactLedger{Sessions: rows[0].Snapshot}, 31)
	if out, err := captureCLI(t, f.project, "", "check", "--repo", f.project, "--state-dir", f.state); err != nil || !bytes.Contains(out, []byte("unresolved operational")) {
		t.Fatalf("unfinished publication not advisory: %v %s", err, out)
	}
	if err := os.Remove(summary); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(summary, beforeSummary, 0600); err != nil {
		t.Fatal(err)
	}
	ledger := f.generate(t, "--append", summary)
	assertCaptureRootInput(t, ledger, 31)
	assertCapturePublishedOwners(t, captureDiagnostics(t, f.project, f.state), f.output, summary, ledger.Sessions)
	if out, err := captureCLI(t, f.project, "", "check", "--repo", f.project, "--state-dir", f.state); err != nil || len(out) != 0 {
		t.Fatalf("successful recovery retained obsolete failure advisory: %v %s", err, out)
	}
}

func TestCaptureStatusReadsCommittedWALWithExplicitBoundedContinuation(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCapturePolicy(t, f, false, 0)
	writer, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			t.Error(err)
		}
	}()
	requests := map[string]string{}
	for _, session := range []string{"first", "second"} {
		row, err := writer.Register(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: session}, "", "SessionStart")
		if err != nil {
			t.Fatal(err)
		}
		requests[session] = row.Request
	}
	for _, limit := range []int{1, 2} {
		out, err := runCaptureNative(t, f.project, "", "capture", "status", "--repo", f.project, "--state-dir", f.state, "--max-sessions", fmt.Sprint(limit))
		if err != nil {
			t.Fatalf("live-WAL status: %v %s", err, out)
		}
		var status struct {
			Registrations []domain.CaptureRegistration `json:"registrations"`
			Continuation  bool                         `json:"registrations_continuation"`
			Limits        []string                     `json:"limits"`
		}
		if err := json.Unmarshal(out, &status); err != nil {
			t.Fatal(err)
		}
		if len(status.Registrations) != limit || status.Continuation != (limit == 1) {
			t.Fatalf("bounded inspection lost committed rows/continuation: %s", out)
		}
		for _, row := range status.Registrations {
			if row.Request == "" || row.Request != requests[row.Identity.SessionID] {
				t.Fatalf("live committed identity/request omitted or fabricated: %+v", row)
			}
		}
		if limit == 1 && !strings.Contains(strings.Join(status.Limits, " "), "partial") {
			t.Fatalf("truncated status omitted inspection limit: %s", out)
		}
	}
	if out, err := captureCLI(t, f.project, "", "check", "--repo", f.project, "--state-dir", f.state, "--max-sessions", "1"); err != nil || !bytes.Contains(out, []byte("session limit")) || bytes.Contains(out, []byte("unbound")) || bytes.Count(out, []byte("AgenticLens capture advisory:")) != 1 {
		t.Fatalf("bounded check lost continuation or guessed unverified bindings: %v %s", err, out)
	}
}

func TestCaptureNativeCheckExemptsOnlyAcceptedExactCallerTail(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := f.write(t, 10)
	f.bind(t, path)
	f.generate(t)
	var publication time.Time
	for _, diagnostic := range captureDiagnostics(t, f.project, f.state) {
		if diagnostic.Phase == "publication" && diagnostic.Outcome == "committed" {
			var err error
			publication, err = time.Parse(time.RFC3339Nano, diagnostic.ObservedAt)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if publication.IsZero() {
		t.Fatal("missing accepted publication checkpoint")
	}
	// Establish source chronology from the actual publication acknowledgement;
	// no callback timestamp or active flag substitutes for this ordering proof.
	f.writeAt(t, 31, publication.Add(time.Second).Format(time.RFC3339Nano))
	reconcile := func() {
		t.Helper()
		if out, err := captureCLI(t, f.project, "", "reconcile", "--repo", f.project, "--state-dir", f.state, "--ticket-id", "TEST-100"); err != nil {
			t.Fatalf("source refresh: %v %s", err, out)
		}
	}
	reconcile()
	args := []string{"check", "--repo", f.project, "--state-dir", f.state}
	if out, err := captureCLI(t, f.project, "", args...); err != nil || !bytes.Contains(out, []byte("newer than")) {
		t.Fatalf("known unpublished source was hidden without caller proof: %v %s", err, out)
	}
	args = append(args, "--source", "codex", "--store", f.store, "--session-id", "root")
	if out, err := captureCLI(t, f.project, "", args...); err != nil || len(out) != 0 {
		t.Fatalf("verified published caller tail was noisy: %v %s", err, out)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	reconcile()
	if out, err := captureCLI(t, f.project, "", args...); err != nil || !bytes.Contains(out, []byte("source refresh failed")) {
		t.Fatalf("exact caller exemption hid a failed refresh: %v %s", err, out)
	}
	assertCaptureRootInput(t, f.ledger(t), 10)
}

func TestCaptureNativeGitAdvisoryKeepsActualWorktreeHealthIsolated(t *testing.T) {
	first := newCaptureGenerationFixture(t)
	base := t.TempDir()
	var err error
	first.state, err = capture.ProjectState(base, first.project)
	if err != nil {
		t.Fatal(err)
	}
	prepareCaptureInstall(t, first.project, []string{"install"})
	git := func(project string, args ...string) []byte {
		t.Helper()
		command := exec.Command("git", append([]string{"-C", project, "-c", "commit.gpgsign=false"}, args...)...)
		command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_AUTHOR_NAME=Capture Test", "GIT_AUTHOR_EMAIL=capture@example.invalid", "GIT_COMMITTER_NAME=Capture Test", "GIT_COMMITTER_EMAIL=capture@example.invalid")
		out, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return out
	}
	git(first.project, "init", "--quiet")
	git(first.project, "add", ".agenticlens-config.yaml")
	git(first.project, "commit", "--quiet", "-m", "initial fixture")
	second := captureGenerationFixture{project: filepath.Join(t.TempDir(), "linked worktree"), store: first.store}
	git(first.project, "worktree", "add", "--quiet", "-b", "linked", second.project)
	second.state, err = capture.ProjectState(base, second.project)
	if err != nil {
		t.Fatal(err)
	}
	second.output = filepath.Join(second.project, "ledger.json")
	writeLocal := func(f captureGenerationFixture, warning bool) {
		t.Helper()
		body := fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n    commit_warning: %t\n", base, warning)
		if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for i, f := range []captureGenerationFixture{first, second} {
		writeLocal(f, true)
		prepareCaptureOwner(t, f, fmt.Sprintf("worktree-%d", i), "TEST-100", 23+i)
		f.generate(t)
	}
	sharedHooks := filepath.Join(t.TempDir(), "shared hooks")
	git(first.project, "config", "core.hooksPath", sharedHooks)
	if out, err := captureCLI(t, first.project, "", "install-check", "--repo", first.project, "--wrapper", fixtureWrapperFromHook(t, first.project)); err != nil {
		t.Fatalf("native hook install: %v %s", err, out)
	}
	if out := git(first.project, "hook", "run", "pre-commit"); len(out) != 0 {
		t.Fatalf("first accepted worktree warned: %s", out)
	}
	if err := os.Remove(second.output); err != nil {
		t.Fatal(err)
	}
	native := exec.Command(cliBinaryPath, "capture", "check", "--repo", second.project, "--state-dir", second.state)
	native.Dir = second.project
	var nativeStderr bytes.Buffer
	native.Stderr = &nativeStderr
	if out, err := native.Output(); err != nil || len(out) != 0 || !bytes.Contains(nativeStderr.Bytes(), []byte("no readable metrics")) {
		t.Fatalf("native missing-metrics control did not reach advisory stderr: %v stdout=%s stderr=%s", err, out, nativeStderr.String())
	}
	config := filepath.Join(first.project, ".git", "config")
	index := strings.TrimSpace(string(git(second.project, "rev-parse", "--path-format=absolute", "--git-path", "index")))
	beforeConfig, beforeIndex := readCaptureFile(t, config), readCaptureFile(t, index)
	if out := git(first.project, "hook", "run", "pre-commit"); len(out) != 0 {
		t.Fatalf("other same-ticket/source worktree contaminated first health: %s", out)
	}
	if out := git(second.project, "hook", "run", "pre-commit"); !bytes.Contains(out, []byte("no readable metrics")) || bytes.Count(out, []byte("AgenticLens capture advisory:")) != 1 {
		t.Fatalf("native shared hook inspected installing worktree instead of caller: %s", out)
	}
	if !bytes.Equal(beforeConfig, readCaptureFile(t, config)) || !bytes.Equal(beforeIndex, readCaptureFile(t, index)) {
		t.Fatal("native advisory changed Git configuration/index")
	}
	if out := git(second.project, "commit", "--quiet", "--allow-empty", "-m", "advisory permits commit"); !bytes.Contains(out, []byte("no readable metrics")) {
		t.Fatalf("real advisory commit lost evidence-backed warning: %s", out)
	}
	writeLocal(second, false)
	if out := git(second.project, "hook", "run", "pre-commit"); len(out) != 0 {
		t.Fatalf("native hook ignored local advisory disable: %s", out)
	}
}
