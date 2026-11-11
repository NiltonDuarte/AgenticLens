package main_test

import (
	"bytes"
	"fmt"
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
	ledger.Sessions[0].Skills = []domain.SkillFact{{Name: "unavailable-skill"}}
	ledger.Sessions[0].Delegations = []domain.DelegationFact{{Description: "unavailable-task"}}
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
	owned := false
	for _, owner := range got.Sessions {
		if owner.ID == "prior-root" && len(owner.Skills) == 1 && owner.Skills[0].Name == "unavailable-skill" && len(owner.Delegations) == 1 {
			owned = true
		}
	}
	if !owned {
		t.Fatal("copied unavailable facts lost their native owner")
	}
	if len(got.Sessions) != 2 || len(got.SkillFacts()) != 1 || len(got.DelegationFacts()) != 1 || got.DelegationFacts()[0].Description != "unavailable-task" {
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
			if facts := first.SkillFacts(); len(facts) != 1 || facts[0].Name != "inspect-source" || facts[0].InvokedAt == nil || *facts[0].InvokedAt != "2026-01-01T00:00:03Z" {
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

func TestGenerateCompleteParentRetainsGroupSummaryForUnavailableChild(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.write(t, 17)
	state := filepath.Join(t.TempDir(), "fresh-state")
	if out, err := directOwnedGenerate(t, f.project, f.store, state, "root", f.output); err != nil {
		t.Fatalf("valid source control: %v %s", err, out)
	}
	facts := f.ledger(t)
	first, second := domain.AgentRole("coder"), domain.AgentRole("tester")
	count := int64(3)
	summary := domain.Anomaly{RuleID: domain.RuleSubagentReworkLoop, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{Role: &first, SecondaryRole: &second, Count: &count}}
	facts.Sessions[0].Anomalies = []domain.Anomaly{summary}
	parent := "root"
	facts.Sessions = append(facts.Sessions, domain.SessionFact{ID: "unavailable-child", ParentID: &parent, Source: domain.SourceCodex, User: "recorded-user", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}, Anomalies: []domain.Anomaly{}})
	if err := adapters.SaveFactLedger(f.output, facts); err != nil {
		t.Fatal(err)
	}
	for pass := 0; pass < 2; pass++ {
		if out, err := directOwnedGenerate(t, f.project, f.store, state, "root", f.output); err != nil {
			t.Fatalf("complete parent refresh %d: %v %s", pass, err, out)
		}
		got := f.ledger(t)
		foundParent, foundChild := false, false
		for _, owner := range got.Sessions {
			if owner.ID == "root" {
				for _, anomaly := range owner.Anomalies {
					if anomaly.RuleID == domain.RuleSubagentReworkLoop && anomaly.Context.Count != nil && *anomaly.Context.Count == 3 {
						foundParent = true
					}
				}
			}
			if owner.ID == "unavailable-child" && owner.ParentID != nil && *owner.ParentID == "root" {
				foundChild = true
			}
		}
		if !foundParent || !foundChild {
			t.Fatalf("pass %d lost incomplete group's recorded summary or child: %+v", pass, got.Sessions)
		}
	}
}

func TestGeneratePostGroupingWarningsPreserveAcceptedLedgerAndSummary(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := f.write(t, 17)
	base := readCaptureFile(t, path)
	spawn := `{"type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\",\"task_name\":\"check\"}"}}
{"type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":"{\"session_id\":\"grandchild\"}"}}
`
	if err := os.WriteFile(path, append(base, []byte(spawn)...), 0600); err != nil {
		t.Fatal(err)
	}
	metadata := func(id, parent string) []byte {
		return []byte(fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"parent_thread_id\":%q,\"cwd\":%q}}\n", id, parent, f.project))
	}
	for _, id := range []string{"child", "grandchild"} {
		if err := os.WriteFile(filepath.Join(f.store, "sessions", id+".jsonl"), metadata(id, "root"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	summary := filepath.Join(f.project, "SUMMARY.md")
	run := func() ([]byte, error) {
		command := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", f.store, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Grouping completeness", "--ticket-type", "fix", "--json-output", f.output, "--append", summary)
		command.Dir = f.project
		return command.CombinedOutput()
	}
	if out, err := run(); err != nil {
		t.Fatalf("complete grouping control: %v %s", err, out)
	}
	accepted := f.ledger(t)
	if len(accepted.Sessions) != 3 || len(accepted.DelegationFacts()) != 1 || accepted.DelegationFacts()[0].ChildSession != "grandchild" {
		t.Fatalf("control lacks source-proven owner/link: %+v", accepted)
	}
	ledgerBefore, summaryBefore := readCaptureFile(t, f.output), readCaptureFile(t, summary)
	// Parsing each owner still succeeds; only cross-owner grouping can discover
	// that the source-proven spawn output disagrees with the child's native parent.
	if err := os.WriteFile(filepath.Join(f.store, "sessions", "grandchild.jsonl"), metadata("grandchild", "child"), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := run()
	if err == nil {
		t.Fatalf("post-parse grouping warning published incomplete owner: %s", out)
	}
	if !strings.Contains(string(out), "incomplete Codex grouping") || !strings.Contains(string(out), "stored snapshots preserved") {
		t.Fatalf("missing grouping completeness diagnostic: %s", out)
	}
	if !bytes.Equal(ledgerBefore, readCaptureFile(t, f.output)) || !bytes.Equal(summaryBefore, readCaptureFile(t, summary)) {
		t.Fatal("post-grouping warning changed accepted ledger or summary")
	}
}

func TestGenerateDiscoveryErrorsRespectSelectedNativeFamily(t *testing.T) {
	for _, tc := range []struct {
		name, selector, corrupt string
		wantFailure             bool
	}{
		{"nonce_selected_child", "--nonce", "child", true},
		{"session_selected_child", "--session-id", "child", true},
		{"nonce_unrelated_native_owner", "--nonce", "unrelated", false},
		{"nonce_unidentifiable_file", "--nonce", "unidentifiable", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			rootPath := f.write(t, 17)
			rootBody := append(readCaptureFile(t, rootPath), []byte(`{"type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"discovery-regression-nonce"}]}}`+"\n")...)
			if err := os.WriteFile(rootPath, rootBody, 0600); err != nil {
				t.Fatal(err)
			}
			metadata := func(id, parent string) string {
				return fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"parent_thread_id\":%q,\"cwd\":%q}}\n", id, parent, f.project)
			}
			childPath := filepath.Join(f.store, "sessions", "child.jsonl")
			if err := os.WriteFile(childPath, []byte(metadata("child", "root")), 0600); err != nil {
				t.Fatal(err)
			}
			summary := filepath.Join(f.project, "SUMMARY.md")
			selectorValue := "root"
			if tc.selector == "--nonce" {
				selectorValue = "discovery-regression-nonce"
			}
			run := func() ([]byte, error) {
				cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", f.store, tc.selector, selectorValue, "--ticket-id", "TEST-100", "--ticket-title", "Native discovery boundary", "--ticket-type", "fix", "--json-output", f.output, "--append", summary)
				cmd.Dir = f.project
				return cmd.CombinedOutput()
			}
			if out, err := run(); err != nil {
				t.Fatalf("valid root/child control: %v %s", err, out)
			}
			control := f.ledger(t)
			if len(control.Sessions) != 2 {
				t.Fatalf("control lacks native child: %+v", control.Sessions)
			}
			beforeLedger, beforeSummary := readCaptureFile(t, f.output), readCaptureFile(t, summary)
			corruptPath := childPath
			// The scanner-boundary failure occurs after valid native identity. A nonce
			// scan must retain that identity when classifying selected-family failures.
			corruptBody := metadata("child", "root") + strings.Repeat("x", 9*1024*1024) + "\n"
			if tc.corrupt == "unrelated" {
				corruptPath = filepath.Join(f.store, "sessions", "unrelated.jsonl")
				corruptBody = metadata("unrelated", "other-root") + strings.Repeat("x", 9*1024*1024) + "\n"
			}
			if tc.corrupt == "unidentifiable" {
				corruptPath = filepath.Join(f.store, "sessions", "unidentifiable.jsonl")
				corruptBody = "invalid native metadata\n"
			}
			if err := os.WriteFile(corruptPath, []byte(corruptBody), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(rootPath, []byte(strings.Replace(string(rootBody), `"input_tokens":17`, `"input_tokens":29`, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			out, err := run()
			if tc.wantFailure {
				if err == nil {
					t.Fatalf("incomplete selected child admitted as complete discovery: %s", out)
				}
				if !strings.Contains(string(out), "child") || (tc.selector == "--nonce" && !strings.Contains(string(out), "incomplete")) || (tc.selector == "--session-id" && !strings.Contains(string(out), "Error reading selected Codex source")) {
					t.Fatalf("missing attributable child source diagnostic: %s", out)
				}
				if !bytes.Equal(beforeLedger, readCaptureFile(t, f.output)) || !bytes.Equal(beforeSummary, readCaptureFile(t, summary)) {
					t.Fatal("selected-child discovery failure changed accepted ledger or summary")
				}
				return
			}
			if err != nil {
				t.Fatalf("unrelated/unidentifiable source poisoned selected family: %v %s", err, out)
			}
			after := f.ledger(t)
			if len(after.Sessions) != 2 {
				t.Fatalf("unrelated source imported or known child lost: %+v", after.Sessions)
			}
			for _, owner := range after.Sessions {
				if owner.ID == "root" && (len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != 29) {
					t.Fatalf("complete selected root did not refresh: %+v", owner)
				}
				if owner.ID == "child" {
					for _, original := range control.Sessions {
						if original.ID == "child" && !reflect.DeepEqual(owner, original) {
							t.Fatalf("unchanged selected child changed: %+v", owner)
						}
					}
				}
			}
		})
	}
}
