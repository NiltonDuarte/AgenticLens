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
