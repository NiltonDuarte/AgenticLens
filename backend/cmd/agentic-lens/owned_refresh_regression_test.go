package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

// Invoke the actual CLI without fixture.args: a receiving machine has no
// private registration, and the test must not silently create that precondition.
func directOwnedGenerate(t *testing.T, project, sourceHome, state, session, output string) ([]byte, error) {
	t.Helper()
	cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", sourceHome, "--capture-dir", state, "--session-id", session, "--ticket-id", "TEST-100", "--ticket-title", "Portable owned facts", "--ticket-type", "fix", "--json-output", output)
	cmd.Dir = project
	return cmd.CombinedOutput()
}

func TestGenerateExactSourceWithoutPrivateRegistrationReplacesPortableOwner(t *testing.T) {
	a, b, _ := newCommittedHistoryFixture(t)
	baseline := b.generate(t)
	path := filepath.Join(a.store, "sessions", "prior-root.jsonl")
	original := readCaptureFile(t, path)
	updated := strings.Replace(string(original), `"input_tokens":17`, `"input_tokens":29`, 1)
	if updated == string(original) {
		t.Fatal("fixture failed to modify exact local source")
	}
	if err := os.WriteFile(path, []byte(updated), 0600); err != nil {
		t.Fatal(err)
	}
	fresh := filepath.Join(t.TempDir(), "no-private-registration")
	out, err := directOwnedGenerate(t, b.project, a.store, fresh, "prior-root", b.output)
	if err != nil {
		t.Fatalf("DATA-02 exact source required private registration/history recovery: %v %s", err, out)
	}
	got := b.ledger(t)
	owners := map[string]domain.SessionFact{}
	for _, owner := range got.Sessions {
		owners[owner.ID] = owner
	}
	if len(owners) != 2 || len(owners["prior-root"].Invocations) != 1 || owners["prior-root"].Invocations[0].Tokens.Input != 29 {
		t.Fatalf("exact native source did not replace portable owner: %+v", owners)
	}
	for _, owner := range baseline.Sessions {
		if owner.ID == "current-root" && !reflect.DeepEqual(owner, owners[owner.ID]) {
			t.Fatal("fresh exact source changed unrelated portable owner")
		}
	}
}

func TestGenerateCopiedLedgerPreservesUnavailableOwnedFactsWithoutPinnedDestination(t *testing.T) {
	_, b, _ := newCommittedHistoryFixture(t)
	ledger := b.ledger(t)
	ledger.Sessions[0].Skills = []domain.SkillFact{{Name: "unavailable-skill", Outcome: domain.OutcomeUnknown}}
	ledger.Sessions[0].TaskLinks = []domain.TaskLinkFact{{Description: "unavailable-task", Outcome: domain.OutcomeUnknown, CostEvidence: domain.UnavailableCost()}}
	copied := filepath.Join(b.project, "copied-ledger.json")
	if err := adapters.SaveFactLedger(copied, ledger); err != nil {
		t.Fatal(err)
	}
	out, err := directOwnedGenerate(t, b.project, b.store, filepath.Join(t.TempDir(), "fresh-state"), "current-root", copied)
	if err != nil {
		t.Fatalf("DATA-02 valid copied ledger required pinned destination/private receipts: %v %s", err, out)
	}
	b.output = copied
	got := b.ledger(t)
	if len(got.Sessions) != 2 || len(got.SkillFacts()) != 1 || len(got.TaskLinkFacts()) != 1 || got.SkillFacts()[0].SessionID != "prior-root" || got.TaskLinkFacts()[0].Description != "unavailable-task" {
		t.Fatalf("copied unavailable owner facts changed: %+v", got)
	}
}

func TestGenerateMalformedSkillArgumentsPreservePreviouslyOwnedSnapshot(t *testing.T) {
	good := `{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"skill-call","name":"load_skill","arguments":"{\"name\":\"inspect-source\"}"}}` + "\n"
	cases := []struct {
		name, original, replacement, diagnostic string
	}{
		{"malformed_arguments", `{\"name\":\"inspect-source\"}`, `{`, "skill arguments unavailable for call skill-call"},
		{"numeric_call_id", `"call_id":"skill-call"`, `"call_id":42`, "invalid response_item payload"},
		{"missing_call_id", `"call_id":"skill-call",`, ``, "response_item function_call requires call_id"},
		{"invalid_name_type", `"name":"load_skill"`, `"name":42`, "invalid response_item payload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 17)
			base := readCaptureFile(t, path)
			if err := os.WriteFile(path, append(append([]byte{}, base...), []byte(good)...), 0600); err != nil {
				t.Fatal(err)
			}
			state := filepath.Join(t.TempDir(), "fresh-state")
			out, err := directOwnedGenerate(t, f.project, f.store, state, "root", f.output)
			if err != nil {
				t.Fatalf("valid source control failed: %v %s", err, out)
			}
			first := f.ledger(t)
			if facts := first.SkillFacts(); len(facts) != 1 || facts[0].Name != "inspect-source" || facts[0].Provenance.CallID != "skill-call" {
				t.Fatalf("valid source failed to produce native control occurrence: %+v", facts)
			}
			before := readCaptureFile(t, f.output)
			bad := strings.Replace(good, tc.original, tc.replacement, 1)
			if bad == good {
				t.Fatal("fixture failed to corrupt recognized skill occurrence")
			}
			if err := os.WriteFile(path, append(append([]byte{}, base...), []byte(bad)...), 0600); err != nil {
				t.Fatal(err)
			}
			out, runErr := directOwnedGenerate(t, f.project, f.store, state, "root", f.output)
			if runErr == nil {
				t.Fatalf("DATA-01 malformed recognized source was published as complete: %s", out)
			}
			for _, diagnostic := range []string{"incomplete Codex owner root", "stored snapshots preserved", tc.diagnostic} {
				if !strings.Contains(string(out), diagnostic) {
					t.Fatalf("malformed source omitted diagnostic %q: %s", diagnostic, out)
				}
			}
			if tc.name != "malformed_arguments" && !strings.Contains(string(out), "codex-rollout-line:") {
				t.Fatalf("malformed outer payload omitted source location: %s", out)
			}
			if !reflect.DeepEqual(before, readCaptureFile(t, f.output)) {
				t.Fatal("malformed source erased previously owned facts or changed accepted ledger bytes")
			}
		})
	}
}
