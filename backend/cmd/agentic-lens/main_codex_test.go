package main_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestGenerateCodexRefreshReplacesSnapshotAndCapturesModelChange(t *testing.T) {
	for _, tc := range []struct {
		ticketID     string
		prefix       string
		workspaceDir string
	}{
		{"TEST-101", "TEST", "records"},
		{"CUSTOM-902", "CUSTOM", "nested/custom records"},
	} {
		t.Run(tc.ticketID, func(t *testing.T) {
			testCodexRefresh(t, tc.ticketID, tc.prefix, tc.workspaceDir)
		})
	}
}

func testCodexRefresh(t *testing.T, ticketID, prefix, workspaceDir string) {
	t.Helper()
	// Run outside the checkout so configuration, Git history, and ticket
	// directories cannot be inherited from the development repository.
	project := t.TempDir()
	config := "project_name: FixtureProject\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes:\n    - prefix: " + prefix + "\nworkspaces:\n  dir: " + workspaceDir + "\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	localConfig := "telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: " + home + "\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(localConfig), 0600); err != nil {
		t.Fatal(err)
	}
	sessions := filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../internal/adapters/codex/testdata/rollout.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	fixture = []byte(strings.ReplaceAll(string(fixture), `"cwd":"/repo/example"`, `"cwd":`+strconv.Quote(project)))
	rollout := filepath.Join(sessions, "rollout-thread-01.jsonl")
	if err := os.WriteFile(rollout, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(project, "agentic_metrics.json")
	generate := func() domain.AgenticFactLedger {
		t.Helper()
		cmd := exec.Command(cliBinaryPath, "generate",
			"--session-id", "thread-01", "--ticket-id", ticketID, "--ticket-title", "Codex refresh regression",
			"--ticket-type", "chore", "--json-output", output)
		cmd.Dir = project
		if result, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generate failed: %v\n%s", err, result)
		}
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var ledger domain.AgenticFactLedger
		if err := json.Unmarshal(data, &ledger); err != nil {
			t.Fatal(err)
		}
		if ledger.TicketID == nil || *ledger.TicketID != ticketID {
			t.Fatalf("expected supplied ticket %q, got %v", ticketID, ledger.TicketID)
		}
		return ledger
	}
	first := generate()
	if len(first.InvocationFacts()) != 1 {
		t.Fatalf("expected one initial model, got %d", len(first.InvocationFacts()))
	}
	modelChange := `
{"type":"response_item","payload":{"type":"message","role":"user"}}
{"timestamp":"2026-10-25T09:01:00Z","type":"turn_context","payload":{"model":"gpt-next","turn_id":"turn-next"}}
{"timestamp":"2026-10-25T09:01:01Z","type":"token_usage_record","payload":{"model":"gpt-next","turn_id":"turn-next","usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":10}}}
`
	if err := os.WriteFile(rollout, append(fixture, []byte(modelChange)...), 0600); err != nil {
		t.Fatal(err)
	}
	second := generate()
	third := generate()
	for _, ledger := range []domain.AgenticFactLedger{second, third} {
		if ledger.Code.SDLCIterations != 2 {
			t.Fatalf("expected two user steering iterations, got %+v", ledger.Code)
		}
		if len(ledger.InvocationFacts()) != 2 {
			t.Fatalf("expected exactly two models after refresh, got %d", len(ledger.InvocationFacts()))
		}
		models := make(map[string]domain.InvocationTokens)
		for _, invocation := range ledger.InvocationFacts() {
			models[invocation.Model] = invocation.Tokens
		}
		if len(models) != 2 || models["gpt-example"].Input != 40 || models["gpt-next"].Input != 80 {
			t.Fatalf("refresh duplicated or lost token usage: %#v", models)
		}
		if len(ledger.ToolFacts()) != len(first.ToolFacts()) || len(ledger.EventFacts()) != len(first.EventFacts()) {
			t.Fatalf("refresh duplicated tool/event facts: tools=%d events=%d", len(ledger.ToolFacts()), len(ledger.EventFacts()))
		}
	}
}

func TestGenerateCodexPreservesChildIdentityAndExcludesInheritedUsage(t *testing.T) {
	project := t.TempDir()
	config := "project_name: ChildFixture\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots:\n    - src\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\nrole_mappings:\n  coder: tester\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	sessions := filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	root := `{"timestamp":"2026-01-01T10:00:00Z","type":"session_meta","payload":{"id":"root-fixture","session_id":"root-fixture","source":"cli","timestamp":"2026-01-01T10:00:00Z"}}
{"timestamp":"2026-01-01T10:00:01Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"fixture-nonce-child-discovery"}]}}
{"timestamp":"2026-01-01T10:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"root-turn"}}
{"timestamp":"2026-01-01T10:00:02Z","type":"turn_context","payload":{"model":"fixture-model","turn_id":"root-turn"}}
{"timestamp":"2026-01-01T10:00:03Z","type":"token_usage_record","payload":{"thread_id":"root-fixture","model":"fixture-model","turn_id":"root-turn","usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":10}}}
{"timestamp":"2026-01-01T10:00:05Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"root-turn"}}
`
	// The inherited records have replay timestamps, so elapsed-time filtering
	// cannot establish ownership. The explicit child thread marker ends replay.
	child := `{"timestamp":"2026-01-01T10:01:00Z","type":"session_meta","payload":{"id":"child-fixture","session_id":"root-fixture","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root-fixture","depth":1,"agent_path":"/private/agents/coder"}}},"timestamp":"2026-01-01T10:01:00Z"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"session_meta","payload":{"id":"root-fixture","session_id":"root-fixture","source":"cli","timestamp":"2026-01-01T10:00:00Z"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"response_item","payload":{"type":"message","role":"user"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"root-turn"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"turn_context","payload":{"model":"inherited-model","turn_id":"root-turn"}}
{"timestamp":"2026-01-01T10:01:00Z","type":"token_usage_record","payload":{"thread_id":"root-fixture","model":"inherited-model","turn_id":"root-turn","usage":{"input_tokens":100,"cached_input_tokens":20,"output_tokens":10}}}
{"timestamp":"2026-01-01T10:01:00Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"root-turn"}}
{"timestamp":"2026-01-01T10:01:01Z","type":"event_msg","payload":{"type":"thread_settings_applied","thread_id":"child-fixture"}}
{"timestamp":"2026-01-01T10:01:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T10:01:02Z","type":"turn_context","payload":{"model":"fixture-model","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T10:01:03Z","type":"token_usage_record","payload":{"thread_id":"child-fixture","model":"fixture-model","turn_id":"child-turn","usage":{"input_tokens":30,"cached_input_tokens":10,"output_tokens":2}}}
{"timestamp":"2026-01-01T10:01:04Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"child-turn"}}
`
	// Synthetic source-format calls deliberately share IDs across inherited history.
	calls := codexOwnedFactCalls("shared", "root-skill")
	root += calls
	child = strings.Replace(child, `{"timestamp":"2026-01-01T10:01:01Z","type":"event_msg"`, calls+`{"timestamp":"2026-01-01T10:01:01Z","type":"event_msg"`, 1)
	child += codexOwnedFactCalls("child-call", "child-skill")
	for name, contents := range map[string]string{"root.jsonl": root, "child.jsonl": child} {
		if err := os.WriteFile(filepath.Join(sessions, name), []byte(strings.ReplaceAll(contents, `"type":"session_meta","payload":{`, `"type":"session_meta","payload":{"cwd":`+strconv.Quote(project)+`,`)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	output := filepath.Join(project, "metrics.json")
	for attempt := 0; attempt < 2; attempt++ {
		cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", home,
			"--nonce", "fixture-nonce-child-discovery", "--ticket-id", "TEST-303", "--ticket-title", "Child attribution regression",
			"--ticket-type", "test", "--json-output", output)
		cmd.Dir = project
		if result, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generate attempt %d failed: %v\n%s", attempt, err, result)
		}
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var ledger domain.AgenticFactLedger
		if err := json.Unmarshal(data, &ledger); err != nil {
			t.Fatal(err)
		}
		if len(ledger.InvocationFacts()) != 2 {
			t.Fatalf("expected exactly root and child, got %#v", ledger.InvocationFacts())
		}
		byID := make(map[string]domain.SessionFact)
		for _, session := range ledger.Sessions {
			byID[session.ID] = session
		}
		rootSession, rootFound := byID["root-fixture"]
		childSession, childFound := byID["child-fixture"]
		if !rootFound || !childFound {
			t.Fatalf("missing distinct root/child identity: %#v", byID)
		}
		if rootSession.ParentID != nil || len(rootSession.Invocations) == 0 || rootSession.Invocations[0].Role != "" || rootSession.Invocations[0].Tokens.Input != 80 {
			t.Fatalf("incorrect root facts: %#v", rootSession)
		}
		if childSession.ParentID == nil || *childSession.ParentID != "root-fixture" || len(childSession.Invocations) == 0 || childSession.Invocations[0].Role != domain.AgentRole("coder") {
			t.Fatalf("incorrect child attribution: %#v", childSession)
		}
		childInvocation := childSession.Invocations[0]
		if childInvocation.Model != "fixture-model" || childInvocation.Turns != 1 || childInvocation.ActiveDurationSeconds != 2 || childInvocation.Tokens.Input != 20 || childInvocation.Tokens.CacheRead != 10 || childInvocation.Tokens.Output != 2 {
			t.Fatalf("inherited history leaked into child usage: %#v", childInvocation)
		}
		if len(rootSession.Skills) != 1 || rootSession.Skills[0].Name != "root-skill" || len(rootSession.TaskLinks) != 1 {
			t.Fatalf("root source facts lost: %+v", rootSession)
		}
		if len(childSession.Skills) != 1 || childSession.Skills[0].Name != "child-skill" || len(childSession.TaskLinks) != 1 {
			t.Fatalf("inherited source facts leaked into child or own facts lost: %+v", childSession)
		}
		if len(ledger.SkillFacts()) != 2 || len(ledger.TaskLinkFacts()) != 2 || len(ledger.Skills) != 0 || len(ledger.TaskLinks) != 0 {
			t.Fatalf("canonical facts duplicated or dropped: %+v", ledger)
		}
		encodedLedger, err := json.Marshal(ledger)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encodedLedger), "/private/agents/coder") {
			t.Fatal("agent path leaked into exported ledger")
		}
		if ledger.Code.SDLCIterations != 1 {
			t.Fatalf("child replay inflated root steering count: %d", ledger.Code.SDLCIterations)
		}
	}
}

func TestGenerateSelectedSourceDoesNotFallback(t *testing.T) {
	sharedConfig := "project_name: SourceFixture\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes: [TEST]\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\ndatabase:\n  path: ./legacy-opencode.db\n"
	for _, tc := range []struct {
		name        string
		local       string
		args        []string
		wantMessage string
		wantPath    string
	}{
		{
			name:        "omitted source honors Codex-only local default and does not fall back",
			local:       "telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: ./missing-codex-home\n    opencode:\n      database_path: ./available-opencode.db\n",
			wantMessage: "selected Codex source", wantPath: "missing-codex-home",
		},
		{
			name:        "explicit source overrides local default and does not fall back",
			local:       "telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: ./available-codex-home\n    opencode:\n      database_path: ./local-opencode.db\n",
			args:        []string{"--source", "opencode", "--db-path", "./missing-cli-opencode.db"},
			wantMessage: "selected OpenCode source", wantPath: "missing-cli-opencode.db",
		},
		{
			name:        "legacy shared OpenCode path works without local config",
			wantMessage: "selected OpenCode source", wantPath: "legacy-opencode.db",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := t.TempDir()
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(sharedConfig), 0600); err != nil {
				t.Fatal(err)
			}
			if tc.local != "" {
				if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(tc.local), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := append([]string{"generate"}, tc.args...)
			args = append(args, "--ticket-id", "TEST-59", "--ticket-title", "Source selection", "--ticket-type", "test", "--json-output", filepath.Join(project, "metrics.json"))
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = project
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected selected source failure, output: %s", output)
			}
			if !strings.Contains(string(output), tc.wantMessage) || !strings.Contains(string(output), tc.wantPath) {
				t.Fatalf("expected selected source error containing %q and %q, got:\n%s", tc.wantMessage, tc.wantPath, output)
			}
			if _, statErr := os.Stat(filepath.Join(project, "metrics.json")); !os.IsNotExist(statErr) {
				t.Fatalf("source failure unexpectedly wrote metrics: stat err=%v", statErr)
			}
		})
	}
}

// These synthetic records mirror the function-call envelope observed in Codex,
// but contain no live ticket data. A file load proves loading, not execution.
func codexOwnedFactCalls(id, skill string) string {
	return `{"timestamp":"2026-01-01T10:00:06Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"` + id + `-read","arguments":"{\"cmd\":\"cat .agents/skills/` + skill + `/SKILL.md\"}"}}
{"timestamp":"2026-01-01T10:00:07Z","type":"response_item","payload":{"type":"function_call_output","call_id":"` + id + `-read","output":"Exit code: 0\nOutput:\nfixture instructions"}}
{"timestamp":"2026-01-01T10:00:08Z","type":"response_item","payload":{"type":"function_call","name":"spawn_agent","call_id":"` + id + `-spawn","arguments":"{\"task_name\":\"coder\",\"message\":\"Run fixture\",\"agent_type\":\"coder\"}"}}
{"timestamp":"2026-01-01T10:00:09Z","type":"response_item","payload":{"type":"function_call_output","call_id":"` + id + `-spawn","output":"{\"task_name\":\"/root/coder\"}"}}
`
}

func TestGenerateCodexPreservesUnavailableOwnersAndReplacesLocalFactSnapshot(t *testing.T) {
	// Reuse actual cross-worktree capture provenance and committed-history
	// authorization rather than bypassing ownership checks with an arbitrary file.
	remoteFixture, localFixture, _ := newCommittedHistoryFixture(t)
	stored := remoteFixture.ledger(t)
	if len(stored.Sessions) != 1 || stored.Sessions[0].ID != "prior-root" {
		t.Fatalf("unexpected history fixture: %+v", stored)
	}
	var facts struct {
		Skills []domain.SkillFact    `json:"skills"`
		Tasks  []domain.TaskLinkFact `json:"task_links"`
	}
	if err := json.Unmarshal([]byte(`{"skills":[{"name":"remote-skill","success":true,"source_invocation_id":"remote-read"}],"task_links":[{"description":"remote-task","source_invocation_id":"remote-spawn","child_session_id":"remote-child"}]}`), &facts); err != nil {
		t.Fatal(err)
	}
	stored.Sessions[0].Skills, stored.Sessions[0].TaskLinks = facts.Skills, facts.Tasks
	stored.Skills = []domain.SkillFact{{Name: "ambiguous-history", Success: true}}
	stored.TaskLinks = []domain.TaskLinkFact{{Description: "ambiguous-task"}}
	if err := adapters.SaveFactLedger(remoteFixture.output, stored); err != nil {
		t.Fatal(err)
	}
	historyGit(t, remoteFixture.project, "add", "ledger.json")
	historyGit(t, remoteFixture.project, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "record historical auxiliary facts")
	historyGit(t, localFixture.project, "merge", "--ff-only", "history-a")
	rollout := filepath.Join(localFixture.store, "sessions", "current-root.jsonl")
	base, err := os.ReadFile(rollout)
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range []struct {
		name, calls, checkpoint string
		localCount              int
	}{
		{"first", codexOwnedFactCalls("first", "local-skill"), "2026-01-01T10:00:10Z", 1},
		{"growing", codexOwnedFactCalls("first", "local-skill") + codexOwnedFactCalls("second", "local-skill"), "2026-01-01T10:01:10Z", 2},
		{"unchanged", codexOwnedFactCalls("first", "local-skill") + codexOwnedFactCalls("second", "local-skill"), "2026-01-01T10:01:10Z", 2},
		{"authoritative local empty", "", "2026-01-01T10:02:10Z", 0},
		{"restore before malformed refresh", codexOwnedFactCalls("restore", "local-skill"), "2026-01-01T10:03:10Z", 1},
	} {
		// A newer source-reported checkpoint proves replacement authority;
		// filesystem rewrite time alone cannot authorize differing facts.
		checkpoint := `{"timestamp":"` + stage.checkpoint + `","type":"token_usage_record","payload":{"thread_id":"current-root","model":"history-model","turn_id":"one","usage":{"input_tokens":17,"output_tokens":2}}}` + "\n"
		if err := os.WriteFile(rollout, append(append([]byte{}, base...), []byte(stage.calls+checkpoint)...), 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--capture-dir", localFixture.state, "--codex-home", localFixture.store, "--session-id", "current-root", "--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "feat", "--json-output", localFixture.output)
		cmd.Dir = localFixture.project
		if result, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s generate: %v\n%s", stage.name, err, result)
		}
		got := localFixture.ledger(t)
		owners := map[string]domain.SessionFact{}
		for _, session := range got.Sessions {
			owners[session.ID] = session
		}
		remote, found := owners["prior-root"]
		if !found || remote.Source != domain.SourceCodex || len(remote.Skills) != 1 || remote.Skills[0].Name != "remote-skill" || len(remote.TaskLinks) != 1 || remote.TaskLinks[0].ChildSession != "remote-child" {
			t.Fatalf("%s unavailable owner facts lost: %+v", stage.name, got)
		}
		local := owners["current-root"]
		if len(local.Skills) != stage.localCount || len(local.TaskLinks) != stage.localCount {
			t.Fatalf("%s local replacement/multiplicity: %+v", stage.name, local)
		}
		if len(got.Skills) != 1 || got.Skills[0].Name != "ambiguous-history" || len(got.TaskLinks) != 1 || got.TaskLinks[0].Description != "ambiguous-task" {
			t.Fatalf("%s unresolved history lost: %+v", stage.name, got)
		}
		if len(got.SkillFacts()) != stage.localCount+2 || len(got.TaskLinkFacts()) != stage.localCount+2 {
			t.Fatalf("%s projection double counted: %+v", stage.name, got)
		}
		for _, skill := range local.Skills {
			if skill.Outcome != "unavailable" || skill.SourceInvocationID == "" {
				t.Fatalf("file read fabricated success or lost occurrence: %+v", skill)
			}
		}
		for _, task := range local.TaskLinks {
			if task.ChildSession != "" || task.ChildReference != "/root/coder" || task.SourceInvocationID == "" {
				t.Fatalf("namespace fabricated child UUID or lost source reference: %+v", task)
			}
		}
	}
	accepted, err := os.ReadFile(localFixture.output)
	if err != nil {
		t.Fatal(err)
	}
	partial := string(base) + codexOwnedFactCalls("partial", "must-not-publish") + `{"timestamp":"2026-01-01T10:04:10Z","type":"token_usage_record","payload":{"thread_id":"current-root","model":"history-model","turn_id":"one","usage":{"input_tokens":17,"output_tokens":2}}}` + "\n" + "{\"type\":\n"
	if err := os.WriteFile(rollout, []byte(partial), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--capture-dir", localFixture.state, "--codex-home", localFixture.store, "--session-id", "current-root", "--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "feat", "--json-output", localFixture.output)
	cmd.Dir = localFixture.project
	if result, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(result), "Partial Codex source parse; previous ledger preserved") {
		t.Fatalf("malformed source did not reach preservation branch: %v %s", err, result)
	}
	after, err := os.ReadFile(localFixture.output)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(accepted) {
		t.Fatal("malformed selected source changed accepted granular ledger")
	}
	retained := localFixture.ledger(t)
	if len(retained.SkillFacts()) != 3 || len(retained.TaskLinkFacts()) != 3 {
		t.Fatalf("aborted malformed refresh lost populated facts: %+v", retained)
	}

}

func TestGenerateUpgradesAuxiliaryCoverageWithoutChangingEqualCheckpointMetrics(t *testing.T) {
	for _, tc := range []struct{ explicit, conflictingMetrics bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		t.Run("explicit-"+strconv.FormatBool(tc.explicit)+"-conflicting-"+strconv.FormatBool(tc.conflictingMetrics), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 10)
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(body, []byte(codexOwnedFactCalls("upgrade", "upgraded"))...), 0600); err != nil {
				t.Fatal(err)
			}
			f.bind(t, path)
			initial := f.generate(t)
			if len(initial.Sessions) != 1 || len(initial.Sessions[0].Skills) != 1 || len(initial.Sessions[0].TaskLinks) != 1 {
				t.Fatalf("source facts missing before cache simulation: %+v", initial)
			}
			baseline := initial
			baseline.Sessions = append([]domain.SessionFact{}, initial.Sessions...)
			baseline.Sessions[0].Skills, baseline.Sessions[0].TaskLinks = nil, nil
			baseline.Sessions[0].AuxiliaryExtractionVersion = 0
			if tc.conflictingMetrics {
				baseline.Sessions[0].Invocations = append([]domain.InvocationFact{}, baseline.Sessions[0].Invocations...)
				baseline.Sessions[0].Invocations[0].Tokens.Input++
			}
			if err := adapters.SaveFactLedger(f.output, baseline); err != nil {
				t.Fatal(err)
			}
			registry, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := registry.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 {
				t.Fatalf("unexpected registry owners: %+v", rows)
			}
			_, err = registry.Change(context.Background(), rows[0].Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				current.ExtractionVersion = 0
				current.Snapshot[0].Skills, current.Snapshot[0].TaskLinks = nil, nil
				current.Snapshot[0].AuxiliaryExtractionVersion = 0
				return current, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := registry.Close(); err != nil {
				t.Fatal(err)
			}
			var extra []string
			if tc.explicit {
				extra = []string{"--session-id", "root", "--codex-home", f.store}
			}
			upgraded := f.generate(t, extra...)
			if tc.conflictingMetrics {
				if len(upgraded.Sessions) != 1 || !reflect.DeepEqual(upgraded.Sessions[0], baseline.Sessions[0]) {
					t.Fatalf("coverage upgrade bypassed equal-checkpoint metric preservation: baseline=%+v after=%+v", baseline.Sessions, upgraded.Sessions)
				}
				return
			}
			if len(upgraded.Sessions) != 1 || !reflect.DeepEqual(upgraded.Sessions[0].Skills, initial.Sessions[0].Skills) || !reflect.DeepEqual(upgraded.Sessions[0].TaskLinks, initial.Sessions[0].TaskLinks) || upgraded.Sessions[0].AuxiliaryExtractionVersion != domain.CurrentAuxiliaryExtractionVersion {
				t.Fatalf("unchanged source cache did not upgrade canonical granular coverage: %+v", upgraded)
			}
			if !reflect.DeepEqual(upgraded.Sessions[0].Invocations, initial.Sessions[0].Invocations) || !reflect.DeepEqual(upgraded.Sessions[0].Tools, initial.Sessions[0].Tools) {
				t.Fatal("coverage upgrade changed equal-checkpoint metric facts")
			}
		})
	}
}
