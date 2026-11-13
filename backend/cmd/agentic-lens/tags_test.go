package main_test

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateTicketTagPrecedencePreservationAndAtomicFailure(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 21))
	shared := filepath.Join(f.project, ".agenticlens-config.yaml")
	local := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	originalShared, originalLocal := readCaptureFile(t, shared), readCaptureFile(t, local)
	if err := os.WriteFile(shared, append(append([]byte{}, originalShared...), []byte("\ntags:\n  team: core\n  environment: dev\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, append(append([]byte{}, originalLocal...), []byte("\ntags:\n  environment: staging\n  operator: alex\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	ledger := f.generate(t, "--tag", " environment = canary ", "--tag", "environment=prod=one", "--tag", "\ufeffcohort=\ufeffvalue\ufeff", "--tag", "\u0085experiment\u0085=\u0085Canary\u0085")
	want := map[string]string{"team": "core", "environment": "prod=one", "operator": "alex", "\ufeffcohort": "\ufeffvalue\ufeff", "experiment": "Canary"}
	if !reflect.DeepEqual(ledger.Tags, want) {
		t.Fatalf("three layer precedence: %v", ledger.Tags)
	}
	if err := os.WriteFile(shared, originalShared, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(local, originalLocal, 0600); err != nil {
		t.Fatal(err)
	}
	// Automatic publication shares the same ticket metadata even with no new tag input.
	ledger = f.generate(t, "--capture-auto")
	if !reflect.DeepEqual(ledger.Tags, want) {
		t.Fatalf("automatic absent-input erased tags: %v", ledger.Tags)
	}
	before := readCaptureFile(t, f.output)
	for _, bad := range []string{"broken", "=value", "key= "} {
		cmd := exec.Command(cliBinaryPath, f.args(t, "--tag", bad)...)
		cmd.Dir = f.project
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("invalid tag accepted %q: %s", bad, out)
		}
		if !bytes.Equal(before, readCaptureFile(t, f.output)) {
			t.Fatal("invalid tag changed saved ledger")
		}
	}
	// Reach a real publication failure after constructing the newly tagged ledger.
	blocker := filepath.Join(f.project, "summary-parent")
	if err := os.WriteFile(blocker, []byte("block directory creation"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, f.args(t, "--tag", "environment=unpublished", "--append", filepath.Join(blocker, "SUMMARY.md"))...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("publication unexpectedly succeeded: %s", out)
	}
	if !bytes.Equal(before, readCaptureFile(t, f.output)) {
		t.Fatal("failed summary publication changed saved ticket tags or ledger")
	}
}

func TestGenerateTicketTagsLateSummaryFailureRestoresReadablePriorSet(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 21))
	f.generate(t, "--tag", "team=accepted")
	beforeLedger := readCaptureFile(t, f.output)
	summary := filepath.Join(f.project, "SUMMARY.md")
	beforeSummary := []byte("# Report\n\n<!-- agentic-lens:metrics:start -->\nUnterminated existing metrics.\n")
	if err := os.WriteFile(summary, beforeSummary, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, f.args(t, "--tag", "team=unpublished", "--append", summary)...)
	cmd.Dir = f.project
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "unterminated generated metrics block") {
		t.Fatalf("did not reach late summary failure: %v %s", err, out)
	}
	if !bytes.Equal(beforeLedger, readCaptureFile(t, f.output)) {
		t.Error("handled late summary failure changed accepted ledger bytes/tags")
	}
	if !bytes.Equal(beforeSummary, readCaptureFile(t, summary)) {
		t.Error("handled late summary failure changed original summary bytes")
	}
	loaded, loadErr := adapters.LoadExistingFactLedger(f.output)
	if loadErr != nil {
		t.Errorf("failed command left supported ledger unreadable: %v", loadErr)
	} else if loaded.Tags["team"] != "accepted" {
		t.Errorf("failed command exposed unpublished tags: %v", loaded.Tags)
	}
	if _, statErr := os.Stat(f.output + ".publication-pending"); !os.IsNotExist(statErr) {
		t.Errorf("handled failure left pending publication: %v", statErr)
	}
	if t.Failed() {
		return
	}
	// Fixing the caller's malformed summary permits an ordinary successful retry.
	if err := os.WriteFile(summary, []byte("# Repaired report\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got := f.generate(t, "--tag", "team=published", "--append", summary)
	if got.Tags["team"] != "published" {
		t.Fatalf("normal retry lost new tags: %v", got.Tags)
	}
	if _, err := adapters.LoadExistingFactLedger(f.output); err != nil {
		t.Fatalf("successful retry left ledger unreadable: %v", err)
	}
}

func TestGenerateTicketTagsSurviveCopiedHistoryAndRepeatedSecondDeveloperRefresh(t *testing.T) {
	a, b := newCaptureGenerationFixture(t), newCaptureGenerationFixture(t)
	historyBind(t, a, "developer-a")
	prior := a.generate(t, "--session-id", "developer-a", "--tag", "team=core", "--tag", "environment=dev")
	if err := os.WriteFile(b.output, readCaptureFile(t, a.output), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(a.store); err != nil {
		t.Fatal(err)
	}
	historyBind(t, b, "developer-b")
	local := filepath.Join(b.project, ".agenticlens-config.local.yaml")
	if err := os.WriteFile(local, append(readCaptureFile(t, local), []byte("\ntags:\n  environment: prod\n  developer: bob\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	// Source-free reads preserve the transported map, regardless of reader defaults.
	loaded, err := adapters.LoadExistingFactLedger(b.output)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Tags, prior.Tags) {
		t.Fatalf("copied tags relabeled before generation: %v", loaded.Tags)
	}
	wantTags := map[string]string{"team": "core", "environment": "prod", "developer": "bob"}
	for pass := 0; pass < 2; pass++ {
		got := b.generate(t, "--session-id", "developer-b")
		if !reflect.DeepEqual(got.Tags, wantTags) || len(got.Sessions) != 2 {
			t.Fatalf("pass %d lost ticket keys or duplicated owners: tags=%v owners=%+v", pass, got.Tags, got.Sessions)
		}
		counts := make(map[string]int)
		for _, owner := range got.Sessions {
			counts[owner.ID]++
			if owner.ID == "developer-a" {
				if !reflect.DeepEqual(owner, prior.Sessions[0]) {
					t.Fatalf("pass %d changed unavailable developer facts: %+v", pass, owner)
				}
			}
		}
		if counts["developer-a"] != 1 || counts["developer-b"] != 1 {
			t.Fatalf("pass %d lost or duplicated exact owners: %v", pass, counts)
		}
	}
}

func TestCaptureWorkerPublishesFreshTicketTagDefaults(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	prepareCaptureOwner(t, f, "root", "TEST-100", 19)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.state)
	for path, tags := range map[string]string{
		filepath.Join(f.project, ".agenticlens-config.yaml"):       "\ntags:\n  team: core\n  environment: dev\n",
		filepath.Join(f.project, ".agenticlens-config.local.yaml"): "\ntags:\n  environment: prod\n  operator: local\n",
	} {
		if err := os.WriteFile(path, append(readCaptureFile(t, path), []byte(tags)...), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := captureCLI(t, f.project, "", "install", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath); err != nil {
		t.Fatalf("install actual worker fixture: %v %s", err, out)
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RequestPublication(context.Background()); err != nil {
		if closeErr := store.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.output); !os.IsNotExist(err) {
		t.Fatalf("worker must create fresh publication: %v", err)
	}
	if out, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state); err != nil {
		t.Fatalf("actual worker publication: %v %s", err, out)
	}
	got := f.ledger(t)
	if !reflect.DeepEqual(got.Tags, map[string]string{"team": "core", "environment": "prod", "operator": "local"}) {
		t.Fatalf("worker subprocess lost ticket defaults: %v", got.Tags)
	}
	assertCaptureRootInput(t, got, 19)
}
