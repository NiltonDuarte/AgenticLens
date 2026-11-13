package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func historyAuthorityFixture(t *testing.T) (captureGenerationFixture, captureGenerationFixture) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	actual, foreign := newCaptureGenerationFixture(t), newCaptureGenerationFixture(t)
	foreign.state = actual.state
	if err := os.WriteFile(filepath.Join(foreign.project, ".agenticlens-config.yaml"), readCaptureFile(t, filepath.Join(actual.project, ".agenticlens-config.yaml")), 0600); err != nil {
		t.Fatal(err)
	}
	for i, f := range []captureGenerationFixture{actual, foreign} {
		historyGit(t, f.project, "init", "-qb", "main")
		historyGit(t, f.project, "config", "user.name", "Fixture")
		historyGit(t, f.project, "config", "user.email", "fixture@example.invalid")
		historyGit(t, f.project, "add", ".agenticlens-config.yaml")
		historyGit(t, f.project, "commit", "-qm", "isolated authority fixture")
		historyBind(t, f, []string{"actual-root", "foreign-root"}[i])
		f.generate(t, "--append", filepath.Join(f.project, "SUMMARY.md"))
	}
	return actual, foreign
}
func historyAuthorityAuxiliary(t *testing.T, f captureGenerationFixture, label string) []byte {
	t.Helper()
	ledger := f.ledger(t)
	ledger.Sessions[0].Skills = []domain.SkillFact{{Name: label + "-skill", CallerRole: domain.KnownString("coder")}}
	ledger.Sessions[0].Delegations = []domain.DelegationFact{{AgentType: domain.KnownString("tester"), Description: label + "-task", ChildSession: "descriptive-only"}}
	body, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(f.output, body, 0600); err != nil {
		t.Fatal(err)
	}
	return body
}
func runHistoryAuthorityGeneration(t *testing.T, f captureGenerationFixture) ([]byte, error) {
	t.Helper()
	owner := f.ledger(t).Sessions[0].ID
	cmd := exec.Command(cliBinaryPath, f.args(t, "--append", filepath.Join(f.project, "SUMMARY.md"), "--session-id", owner, "--codex-home", f.store, "--capture-dir", filepath.Join(f.project, "explicit-state"))...)
	cmd.Dir = f.project
	return cmd.CombinedOutput()
}

// Security finding severity6, security specialist / gpt-6-luna / medium: matching
// ownerless rows in another repository must not become committed authority via
// the Git environment inherited by a native hook.
func TestCaptureExactSourceRemainsAuthoritativeDespiteForeignGitEnvironment(t *testing.T) {
	for _, mode := range []string{"foreign GIT_DIR", "hook override bundle"} {
		t.Run(mode, func(t *testing.T) {
			actual, foreign := historyAuthorityFixture(t)
			trusted := historyAuthorityAuxiliary(t, actual, "actual-committed")
			historyGit(t, actual.project, "add", "agentic_metrics.json")
			historyGit(t, actual.project, "commit", "-qm", "actual auxiliary authority")
			historyAuthorityAuxiliary(t, foreign, "foreign-forged")
			historyGit(t, foreign.project, "add", "agentic_metrics.json")
			historyGit(t, foreign.project, "commit", "-qm", "foreign same-ticket auxiliary rows")
			forged := historyAuthorityAuxiliary(t, actual, "foreign-forged")
			summary := readCaptureFile(t, filepath.Join(actual.project, "SUMMARY.md"))
			foreignBefore := foreignHistoryState(t, foreign)
			t.Setenv("GIT_DIR", filepath.Join(foreign.project, ".git"))
			if mode == "hook override bundle" {
				for key, value := range map[string]string{"GIT_WORK_TREE": foreign.project, "GIT_COMMON_DIR": filepath.Join(foreign.project, ".git"), "GIT_OBJECT_DIRECTORY": filepath.Join(foreign.project, ".git", "objects"), "GIT_ALTERNATE_OBJECT_DIRECTORIES": filepath.Join(foreign.project, ".git", "objects"), "GIT_INDEX_FILE": filepath.Join(foreign.project, ".git", "index"), "GIT_CONFIG_COUNT": "1", "GIT_CONFIG_KEY_0": "core.repositoryformatversion", "GIT_CONFIG_VALUE_0": "99"} {
					t.Setenv(key, value)
				}
			}
			if out, err := runHistoryAuthorityGeneration(t, actual); err != nil {
				t.Fatalf("exact local source was hidden by foreign Git environment: %v %s", err, out)
			}
			result := actual.ledger(t)
			if len(result.SkillFacts()) != 0 || len(result.DelegationFacts()) != 0 {
				t.Fatal("complete empty local snapshot retained manual auxiliary facts")
			}
			if reflect.DeepEqual(forged, readCaptureFile(t, actual.output)) {
				t.Fatal("source authority failed to replace manual differences")
			}
			if len(summary) == 0 {
				t.Fatal("fixture summary absent")
			}

			if err := adapters.CheckPublication(actual.output); err != nil {
				t.Fatalf("foreign authority was rejected after staging: %v", err)
			}
			if foreignHistoryState(t, foreign) != foreignBefore {
				t.Fatal("foreign auxiliary authority check mutated foreign scope")
			}
			// The same hostile environment must still permit actual committed rows.
			// Rejection of the first case cannot merely result from unusable Git commands.
			if err := os.WriteFile(actual.output, trusted, 0600); err != nil {
				t.Fatal(err)
			}
			for pass := 0; pass < 2; pass++ {
				if out, err := runHistoryAuthorityGeneration(t, actual); err != nil {
					t.Fatalf("actual authority failed under %s: %v %s", mode, err, out)
				}
				ledger := actual.ledger(t)
				if len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != "actual-root" || len(ledger.SkillFacts()) != 0 || len(ledger.DelegationFacts()) != 0 {
					t.Fatalf("wrong repository supplied evidence: %+v", ledger)
				}
				if foreignHistoryState(t, foreign) != foreignBefore {
					t.Fatal("actual committed publication mutated foreign scope")
				}
			}
		})
	}
}

func TestCaptureExactSourceDoesNotRequireAncestorGitAuthority(t *testing.T) {
	actual, foreign := historyAuthorityFixture(t)
	historyAuthorityAuxiliary(t, actual, "ancestor-committed")
	historyGit(t, actual.project, "add", "agentic_metrics.json")
	historyGit(t, actual.project, "commit", "-qm", "ancestor auxiliary ledger")
	nested := actual
	nested.project = filepath.Join(actual.project, "nested project")
	nested.output = filepath.Join(nested.project, "agentic_metrics.json")
	nested.store = t.TempDir()
	if err := os.Mkdir(nested.project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested.project, ".agenticlens-config.yaml"), readCaptureFile(t, filepath.Join(actual.project, ".agenticlens-config.yaml")), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeCaptureFixtureLocal(t, nested, []byte("telemetry:\n  capture:\n")); err != nil {
		t.Fatal(err)
	}
	historyBind(t, nested, "nested-root")
	// Ordinary current-scope publication remains supported without an observed
	// worktree Git marker or committed-history dependency.
	nested.generate(t, "--append", filepath.Join(nested.project, "SUMMARY.md"))
	candidate := historyAuthorityAuxiliary(t, nested, "ancestor-committed")
	historyGit(t, actual.project, "add", "nested project/agentic_metrics.json")
	historyGit(t, actual.project, "commit", "-qm", "nested scope committed auxiliary ledger")
	foreignBefore := foreignHistoryState(t, foreign)
	if out, err := runHistoryAuthorityGeneration(t, nested); err != nil {
		t.Fatalf("exact source required ancestor Git authority: %v %s", err, out)
	}
	if reflect.DeepEqual(candidate, readCaptureFile(t, nested.output)) || len(nested.ledger(t).SkillFacts()) != 0 {
		t.Fatal("complete exact source did not replace manually supplied auxiliary facts")
	}

	if err := adapters.CheckPublication(nested.output); err != nil {
		t.Fatalf("missing physical authority staged publication: %v", err)
	}
	if foreignHistoryState(t, foreign) != foreignBefore {
		t.Fatal("missing physical authority mutated unrelated foreign scope")
	}
}
