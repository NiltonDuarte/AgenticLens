package main_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	config := "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nrole_mappings: {}\nproject_name: FixtureProject\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: " + prefix + "\n      description: Fixture tickets\nworkspaces:\n  artifacts: []\n  dir: " + workspaceDir + "\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	localConfig := "telemetry:\n  capture:\n    enabled: false\n    state_dir: .agenticlens-capture\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n  default_source: codex\n  sources:\n    codex:\n      home: " + home + "\n"
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
		prepareGenerateFixture(t, cmd)
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
		if len(ledger.ToolFacts()) != len(first.ToolFacts()) {
			t.Fatalf("refresh duplicated tool facts: tools=%d", len(ledger.ToolFacts()))
		}
	}
}

func TestGenerateCodexPreservesChildIdentityAndExcludesInheritedUsage(t *testing.T) {
	project := t.TempDir()
	config := "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\nrole_order: []\nproject_name: ChildFixture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: TEST\n      description: Fixture ticket\nworkspaces:\n  artifacts: []\n  dir: records\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\nrole_mappings:\n  coder: tester\n"
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
		prepareGenerateFixture(t, cmd)
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
		encodedLedger, err := json.Marshal(ledger)
		if err != nil {
			t.Fatal(err)
		}

		if strings.Contains(string(encodedLedger), `"agent_path"`) {
			t.Fatal("canonical ledger retained native agent path")
		}
		if ledger.Code.SDLCIterations != 1 {
			t.Fatalf("child replay inflated root steering count: %d", ledger.Code.SDLCIterations)
		}
	}
}

func TestGenerateSelectedSourceDoesNotFallback(t *testing.T) {
	for _, scenario := range []string{"missing-source", "invalid-source", "blank-source", "missing-codex-home", "invalid-codex-home", "blank-codex-home", "missing-db-path", "invalid-db-path", "blank-db-path", "missing-selector"} {
		t.Run(scenario, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.write(t, 17)
			dbPath := setupIsolatedEmptyDB(t)
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(`INSERT INTO session(id)VALUES('root')`); err != nil {
				t.Fatal(err)
			}
			if err = db.Close(); err != nil {
				t.Fatal(err)
			}
			setOpenCodeFixtureProject(t, dbPath, f.project)
			local := fmt.Sprintf("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n    opencode:\n      database_path: %q\n  capture:\n    enabled: false\n    state_dir: %q\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n", f.store, dbPath, f.stateBase)
			args := []string{"generate", "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Explicit selection", "--json-output", f.output, "--append", filepath.Join(f.project, "SUMMARY.md")}
			want := ""
			switch scenario {
			case "missing-source":
				local = strings.Replace(local, "  default_source: codex\n", "", 1)
				want = "source is required"
			case "invalid-source":
				args = append(args, "--source", "unknown")
				want = "source"
			case "blank-source":
				args = append(args, "--source", "")
				want = "--source"
			case "missing-codex-home":
				local = strings.Replace(local, fmt.Sprintf("    codex:\n      home: %q\n", f.store), "", 1)
				want = "codex.home"
			case "invalid-codex-home":
				args = append(args, "--codex-home", filepath.Join(t.TempDir(), "missing-selected-store"))
				want = "missing-selected-store"
			case "blank-codex-home":
				args = append(args, "--codex-home", "")
				want = "--codex-home"
			case "missing-db-path":
				local = strings.Replace(local, fmt.Sprintf("    opencode:\n      database_path: %q\n", dbPath), "", 1)
				args = append(args, "--source", "opencode")
				want = "database_path"
			case "invalid-db-path":
				args = append(args, "--source", "opencode", "--db-path", filepath.Join(t.TempDir(), "missing-selected.db"))
				want = "missing-selected.db"
			case "blank-db-path":
				args = append(args, "--source", "opencode", "--db-path", "")
				want = "--db-path"
			case "missing-selector":
				args = append(args[:1], args[3:]...)
				want = "session-id"
			}
			if err = os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = f.project
			// A usable competing environment store must never satisfy omitted policy.
			cmd.Env = append(os.Environ(), "CODEX_HOME="+f.store)
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), want) {
				t.Fatalf("selected failure %q: %v %s", want, err, out)
			}
			for _, path := range []string{f.output, filepath.Join(f.project, "SUMMARY.md")} {
				if _, err = os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("selection failure wrote %s: %v", path, err)
				}
			}
		})
	}
}
