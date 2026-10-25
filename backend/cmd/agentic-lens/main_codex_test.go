package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
	config := "project_name: FixtureProject\nticket:\n  prefixes:\n    - prefix: " + prefix + "\nworkspaces:\n  dir: " + workspaceDir + "\n  src_folders:\n    - src\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	sessions := filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../internal/adapters/codex/testdata/rollout.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	rollout := filepath.Join(sessions, "rollout-thread-01.jsonl")
	if err := os.WriteFile(rollout, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(home, "agentic_metrics.json")
	generate := func() domain.AgenticFactLedger {
		t.Helper()
		cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", home,
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
	if len(first.Invocations) != 1 {
		t.Fatalf("expected one initial model, got %d", len(first.Invocations))
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
		if len(ledger.Invocations) != 2 {
			t.Fatalf("expected exactly two models after refresh, got %d", len(ledger.Invocations))
		}
		models := make(map[string]domain.InvocationTokens)
		for _, invocation := range ledger.Invocations {
			models[invocation.Model] = invocation.Tokens
		}
		if len(models) != 2 || models["gpt-example"].Input != 40 || models["gpt-next"].Input != 80 {
			t.Fatalf("refresh duplicated or lost token usage: %#v", models)
		}
		if len(ledger.Tools) != len(first.Tools) || len(ledger.Events) != len(first.Events) {
			t.Fatalf("refresh duplicated tool/event facts: tools=%d events=%d", len(ledger.Tools), len(ledger.Events))
		}
	}
}

func TestGenerateCodexPreservesChildIdentityAndExcludesInheritedUsage(t *testing.T) {
	project := t.TempDir()
	config := "project_name: ChildFixture\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  dir: records\n  src_folders:\n    - src\n"
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
	child := `{"timestamp":"2026-01-01T10:01:00Z","type":"session_meta","payload":{"id":"child-fixture","session_id":"root-fixture","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root-fixture","depth":1}}},"timestamp":"2026-01-01T10:01:00Z"}}
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
	for name, contents := range map[string]string{"root.jsonl": root, "child.jsonl": child} {
		if err := os.WriteFile(filepath.Join(sessions, name), []byte(contents), 0600); err != nil {
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
		if len(ledger.Invocations) != 2 {
			t.Fatalf("expected exactly root and child, got %#v", ledger.Invocations)
		}
		byID := make(map[string]domain.InvocationFact)
		for _, invocation := range ledger.Invocations {
			byID[invocation.ID] = invocation
		}
		rootInvocation, rootFound := byID["root-fixture"]
		childInvocation, childFound := byID["child-fixture"]
		if !rootFound || !childFound {
			t.Fatalf("missing distinct root/child identity: %#v", byID)
		}
		if rootInvocation.ParentID != nil || rootInvocation.Role != domain.RoleRootAgent || rootInvocation.Tokens.Input != 80 {
			t.Fatalf("incorrect root facts: %#v", rootInvocation)
		}
		if childInvocation.ParentID == nil || *childInvocation.ParentID != "root-fixture" || childInvocation.Role != domain.RoleUnassignedSubAgent {
			t.Fatalf("incorrect child attribution: %#v", childInvocation)
		}
		if childInvocation.Model != "fixture-model" || childInvocation.Turns != 1 || childInvocation.ActiveDurationSeconds != 2 || childInvocation.Tokens.Input != 20 || childInvocation.Tokens.CacheRead != 10 || childInvocation.Tokens.Output != 2 {
			t.Fatalf("inherited history leaked into child usage: %#v", childInvocation)
		}
		if ledger.Code.SDLCIterations != 1 {
			t.Fatalf("child replay inflated root steering count: %d", ledger.Code.SDLCIterations)
		}
	}
}
