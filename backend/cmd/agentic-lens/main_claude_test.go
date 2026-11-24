package main_test

// All Claude Code stores here are synthetic fixtures in temp directories. They
// reproduce native record keys without transcript content and never read
// ~/.claude, CLAUDE_CONFIG_DIR or a live ticket workspace. Expected values are
// hand-computed from the usage and timestamps written in each fixture.

import (
	"bytes"
	"encoding/json"
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

const claudeCLIRoot = "claude-cli-root"

type claudeCLIFixture struct {
	project, home, output string
}

func claudeCLISharedConfig() string {
	return "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\n    capture_observation: false\nrole_order: []\nrole_mappings: {}\nproject_name: ClaudeFixture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: TEST\n      description: Fixture tickets\nworkspaces:\n  artifacts: []\n  roots:\n    - path: records\n      discovery:\n        max_depth: 0\n        max_directories: 300\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
}

// newClaudeCLIFixture writes a project whose local default source is codex
// with an absent Codex store: reading it would fail the command, so success
// proves that only the selected Claude Code store was read.
func newClaudeCLIFixture(t *testing.T) claudeCLIFixture {
	t.Helper()
	f := claudeCLIFixture{project: t.TempDir(), home: t.TempDir()}
	f.output = filepath.Join(f.project, "records", "agentic_metrics.json")
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), []byte(claudeCLISharedConfig()), 0o600); err != nil {
		t.Fatal(err)
	}
	f.writeLocal(t, fmt.Sprintf("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n", filepath.Join(f.project, "absent-codex-store")))
	return f
}

func (f claudeCLIFixture) writeLocal(t *testing.T, telemetry string) {
	t.Helper()
	body := telemetry + "  capture:\n    enabled: false\n    state_dir: .agenticlens-capture\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n" + fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(f.project, ".publication-state"))
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Typed synthetic Claude Code records; only native keys are written.
type claudeCLIRecord struct {
	Type          string           `json:"type"`
	SessionID     string           `json:"sessionId"`
	AgentID       string           `json:"agentId,omitempty"`
	IsSidechain   bool             `json:"isSidechain,omitempty"`
	UUID          string           `json:"uuid,omitempty"`
	Timestamp     string           `json:"timestamp"`
	CWD           string           `json:"cwd,omitempty"`
	PromptSource  string           `json:"promptSource,omitempty"`
	RequestID     string           `json:"requestId,omitempty"`
	ToolUseResult *claudeCLIResult `json:"toolUseResult,omitempty"`
	Message       claudeCLIMessage `json:"message"`
}

type claudeCLIMessage struct {
	ID         string           `json:"id,omitempty"`
	Model      string           `json:"model,omitempty"`
	Role       string           `json:"role,omitempty"`
	StopReason string           `json:"stop_reason,omitempty"`
	Usage      *claudeCLIUsage  `json:"usage,omitempty"`
	Content    claudeCLIContent `json:"content"`
}

type claudeCLIUsage struct {
	Input     int `json:"input_tokens"`
	Output    int `json:"output_tokens"`
	CacheRead int `json:"cache_read_input_tokens"`
}

// claudeCLIContent encodes as a prompt string, or as content blocks when
// Blocks is non-nil.
type claudeCLIContent struct {
	Text   string
	Blocks []claudeCLIBlock
}

func (c claudeCLIContent) MarshalJSON() ([]byte, error) {
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(c.Text)
}

type claudeCLIBlock struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     *claudeCLIInput `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type claudeCLIInput struct {
	SubagentType string `json:"subagent_type,omitempty"`
	Description  string `json:"description,omitempty"`
}

type claudeCLIResult struct {
	AgentID   string `json:"agentId,omitempty"`
	AgentType string `json:"agentType,omitempty"`
	Status    string `json:"status,omitempty"`
}

func claudeCLIPrompt(cwd, at string) claudeCLIRecord {
	return claudeCLIRecord{Type: "user", UUID: "u-" + at, Timestamp: at, CWD: cwd, PromptSource: "sdk", Message: claudeCLIMessage{Role: "user", Content: claudeCLIContent{Text: "synthetic"}}}
}

func claudeCLIResponse(at, request, model, stop string, input, output, cacheRead int, blocks ...claudeCLIBlock) claudeCLIRecord {
	if blocks == nil {
		blocks = []claudeCLIBlock{}
	}
	return claudeCLIRecord{Type: "assistant", Timestamp: at, RequestID: request, Message: claudeCLIMessage{ID: "m-" + request, Model: model, StopReason: stop,
		Usage: &claudeCLIUsage{Input: input, Output: output, CacheRead: cacheRead}, Content: claudeCLIContent{Blocks: blocks}}}
}

func claudeCLIAgentCall(id, agentType string) claudeCLIBlock {
	return claudeCLIBlock{Type: "tool_use", ID: id, Name: "Agent", Input: &claudeCLIInput{SubagentType: agentType, Description: "synthetic"}}
}

func claudeCLIAgentResult(at, callID, agentID string) claudeCLIRecord {
	return claudeCLIRecord{Type: "user", Timestamp: at, ToolUseResult: &claudeCLIResult{AgentID: agentID, Status: "completed"},
		Message: claudeCLIMessage{Role: "user", Content: claudeCLIContent{Blocks: []claudeCLIBlock{{Type: "tool_result", ToolUseID: callID}}}}}
}

func writeClaudeCLITranscript(t *testing.T, path, sessionID, agentID string, records []claudeCLIRecord) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	for _, record := range records {
		record.SessionID = sessionID
		if agentID != "" {
			record.AgentID, record.IsSidechain = agentID, true
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		body.Write(encoded)
		body.WriteByte('\n')
	}
	if err := os.WriteFile(path, body.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}

func claudeCLIRootPath(home, folder, root string) string {
	return filepath.Join(home, "projects", folder, root+".jsonl")
}

func (f claudeCLIFixture) generate(t *testing.T, extra ...string) (domain.AgenticFactLedger, string, error) {
	t.Helper()
	args := append([]string{"generate", "--ticket-id", "TEST-198", "--ticket-title", "Claude Code fixture", "--ticket-type", "feat", "--json-output", f.output}, extra...)
	cmd := exec.Command(cliBinaryPath, args...)
	cmd.Dir = f.project
	prepareScopedGenerateFixture(t, cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return domain.AgenticFactLedger{}, string(out), err
	}
	data, readErr := os.ReadFile(f.output)
	if readErr != nil {
		t.Fatal(readErr)
	}
	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	return ledger, string(out), nil
}

func claudeCLISessions(ledger domain.AgenticFactLedger) map[string]domain.SessionFact {
	byID := map[string]domain.SessionFact{}
	for _, session := range ledger.Sessions {
		byID[session.ID] = session
	}
	return byID
}

func TestGenerateClaudeRefreshReplacesSnapshotRepeatsEquallyAndRestoresSmallerSnapshot(t *testing.T) {
	f := newClaudeCLIFixture(t)
	path := claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot)
	first := []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:05Z", "r-1", "model-a", "end_turn", 100, 10, 20),
	}
	writeClaudeCLITranscript(t, path, claudeCLIRoot, "", first)
	selection := []string{"--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot}
	initial, out, err := f.generate(t, selection...)
	if err != nil {
		t.Fatalf("generate failed: %v %s", err, out)
	}
	if got := claudeCLISessions(initial)[claudeCLIRoot]; got.Source != domain.SourceClaude || len(got.Invocations) != 1 || got.Invocations[0].Tokens.Input != 100 || got.Invocations[0].ActiveDurationSeconds != 5 {
		t.Fatalf("initial snapshot: %+v", got)
	}
	// The ticket starts at the root's first recorded prompt, and a new owner is
	// attributed to the selected project's Git user.
	if initial.StartedAt == nil || *initial.StartedAt != "2026-01-01T10:00:00Z" {
		t.Fatalf("ledger started_at must be the first Claude Code record: %v", initial.StartedAt)
	}
	if got := claudeCLISessions(initial)[claudeCLIRoot].User; got != projectGitUser {
		t.Fatalf("new owner user: %q, want %q", got, projectGitUser)
	}
	grown := append(append([]claudeCLIRecord{}, first...),
		claudeCLIPrompt(f.project, "2026-01-01T10:01:00Z"),
		claudeCLIResponse("2026-01-01T10:01:10Z", "r-2", "model-b", "end_turn", 50, 5, 0))
	writeClaudeCLITranscript(t, path, claudeCLIRoot, "", grown)
	var previous domain.AgenticFactLedger
	for pass := 0; pass < 2; pass++ {
		ledger, out, err := f.generate(t, selection...)
		if err != nil {
			t.Fatalf("refresh %d failed: %v %s", pass, err, out)
		}
		root := claudeCLISessions(ledger)[claudeCLIRoot]
		if len(ledger.Sessions) != 1 || len(root.Invocations) != 2 {
			t.Fatalf("refresh %d must replace the snapshot with one row per model: %+v", pass, ledger.Sessions)
		}
		models := map[string]int64{}
		var iterations int64
		for _, row := range root.Invocations {
			models[row.Model] = row.Tokens.Input
			iterations += row.UserIterations
		}
		if models["model-a"] != 100 || models["model-b"] != 50 || iterations != 2 {
			t.Fatalf("refresh %d added or lost usage: %v iterations=%d", pass, models, iterations)
		}
		if pass == 1 && !reflect.DeepEqual(previous.Sessions, ledger.Sessions) {
			t.Fatalf("unchanged repeat changed facts:\n%+v\n%+v", previous.Sessions, ledger.Sessions)
		}
		previous = ledger
	}
	// An older but complete transcript replaces the larger stored snapshot.
	writeClaudeCLITranscript(t, path, claudeCLIRoot, "", first)
	restored, out, err := f.generate(t, selection...)
	if err != nil {
		t.Fatalf("restored generate failed: %v %s", err, out)
	}
	if root := claudeCLISessions(restored)[claudeCLIRoot]; len(restored.Sessions) != 1 || len(root.Invocations) != 1 || root.Invocations[0].Model != "model-a" || root.Invocations[0].Tokens.Input != 100 {
		t.Fatalf("a complete restored snapshot must replace the larger one: %+v", root)
	}
}

func TestGenerateClaudeMixedHarnessLedgerKeepsOtherOwnersAndUnavailableChild(t *testing.T) {
	f := newClaudeCLIFixture(t)
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-linked-worktree-folder", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "tool_use", 7, 1, 0,
			claudeCLIBlock{Type: "tool_use", ID: "call-gone", Name: "Agent", Input: &claudeCLIInput{SubagentType: "coder", Description: "code"}}),
		{Type: "user", Timestamp: "2026-01-01T10:00:02Z", ToolUseResult: &claudeCLIResult{AgentID: "agone", Status: "async_launched"}, Message: claudeCLIMessage{Role: "user", Content: claudeCLIContent{Blocks: []claudeCLIBlock{{Type: "tool_result", ToolUseID: "call-gone"}}}}},
		claudeCLIResponse("2026-01-01T10:00:03Z", "r-2", "model-a", "end_turn", 3, 1, 0),
	})
	version := domain.CurrentSchemaVersion
	ticket, title := "TEST-198", "Claude Code fixture"
	root := claudeCLIRoot
	stored := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{
		{ID: "codex-owner", Source: domain.SourceCodex, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "codex-model", ActiveDurationSeconds: 11, Tokens: domain.InvocationTokens{Input: 11}}}, Tools: []domain.SessionToolFact{{Name: "exec", Calls: 2}}},
		{ID: "opencode-owner", Source: domain.SourceOpenCode, User: "remote", Invocations: []domain.InvocationFact{{Role: "tester", Model: "opencode-model", ActiveDurationSeconds: 13, Tokens: domain.InvocationTokens{Input: 13}}}},
		{ID: "claude-other-machine", Source: domain.SourceClaude, User: "remote", Skills: []domain.SkillFact{{Name: "remote-skill"}}, Invocations: []domain.InvocationFact{{Model: "remote-model", ActiveDurationSeconds: 17, Tokens: domain.InvocationTokens{Input: 17}}}},
		{ID: "agone", ParentID: &root, Source: domain.SourceClaude, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "child-model", ActiveDurationSeconds: 19, Tokens: domain.InvocationTokens{Input: 19}}}},
	}}
	if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.SaveFactLedger(f.output, stored); err != nil {
		t.Fatal(err)
	}
	var before domain.AgenticFactLedger
	if err := json.Unmarshal(readCaptureFile(t, f.output), &before); err != nil {
		t.Fatal(err)
	}
	ledger, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
	if err != nil {
		t.Fatalf("generate failed: %v %s", err, out)
	}
	after, previous := claudeCLISessions(ledger), claudeCLISessions(before)
	if len(after) != 5 {
		t.Fatalf("want 4 preserved owners and the new root, got %+v", ledger.Sessions)
	}
	for _, id := range []string{"codex-owner", "opencode-owner", "claude-other-machine", "agone"} {
		if !reflect.DeepEqual(after[id], previous[id]) {
			t.Fatalf("%s changed:\nbefore %+v\nafter  %+v", id, previous[id], after[id])
		}
	}
	if got := after[claudeCLIRoot]; len(got.Invocations) != 1 || got.Invocations[0].Tokens.Input != 10 || len(got.Delegations) != 1 || got.Delegations[0].ChildSession != "agone" {
		t.Fatalf("local root not refreshed with its explicit child reference: %+v", got)
	}
	if !strings.Contains(out, "agone unavailable") {
		t.Fatalf("the unavailable child must be reported: %s", out)
	}
	total := domain.ComputeWorkflowEfficiency(ledger)
	// 11 + 13 + 17 + 19 stored seconds plus the root's 3 measured seconds.
	if total.ActiveDurationSeconds != 63 {
		t.Fatalf("aggregates must cover refreshed and preserved owners together: %d", total.ActiveDurationSeconds)
	}
}

func TestGenerateClaudeOwnershipAndHarnessCollisionFailBeforePublication(t *testing.T) {
	t.Run("recorded cwd outside the project", func(t *testing.T) {
		f := newClaudeCLIFixture(t)
		writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
			claudeCLIPrompt(t.TempDir(), "2026-01-01T10:00:00Z"),
			claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "end_turn", 1, 1, 1),
		})
		_, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
		if err == nil || !strings.Contains(out, "Source ownership conflict") {
			t.Fatalf("an outside root must be rejected: %v %s", err, out)
		}
		if _, statErr := os.Stat(f.output); !os.IsNotExist(statErr) {
			t.Fatalf("rejected root published a ledger: %v", statErr)
		}
	})
	t.Run("session ID recorded for another harness", func(t *testing.T) {
		f := newClaudeCLIFixture(t)
		writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
			claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
			claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "end_turn", 1, 1, 1),
		})
		version := domain.CurrentSchemaVersion
		ticket, title := "TEST-198", "Claude Code fixture"
		if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := adapters.SaveFactLedger(f.output, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{
			{ID: claudeCLIRoot, Source: domain.SourceCodex, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "codex-model"}}},
		}}); err != nil {
			t.Fatal(err)
		}
		before := readCaptureFile(t, f.output)
		_, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
		if err == nil || !strings.Contains(out, "collides across source harnesses") {
			t.Fatalf("a cross-harness ID collision must fail: %v %s", err, out)
		}
		if !bytes.Equal(before, readCaptureFile(t, f.output)) {
			t.Fatal("a rejected collision changed the ledger")
		}
	})
	t.Run("incomplete root", func(t *testing.T) {
		f := newClaudeCLIFixture(t)
		path := claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot)
		writeClaudeCLITranscript(t, path, claudeCLIRoot, "", []claudeCLIRecord{
			claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
			{Type: "assistant", Timestamp: "2026-01-01T10:00:01Z", RequestID: "r-1", Message: claudeCLIMessage{ID: "m-1", Model: "model-a", StopReason: "end_turn", Content: claudeCLIContent{Blocks: []claudeCLIBlock{}}}},
		})
		_, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
		if err == nil || !strings.Contains(out, "incomplete Claude Code root session") || !strings.Contains(out, "lacks usage") {
			t.Fatalf("missing usage must fail instead of recording zero: %v %s", err, out)
		}
		if _, statErr := os.Stat(f.output); !os.IsNotExist(statErr) {
			t.Fatalf("incomplete root published a ledger: %v", statErr)
		}
	})
}

// The flag wins over local configuration, the configured default selects the
// configured store, and CLAUDE_CONFIG_DIR never substitutes for either.
func TestGenerateClaudeStorePrecedenceAndConfiguredDefault(t *testing.T) {
	f := newClaudeCLIFixture(t)
	configured := t.TempDir()
	for home, input := range map[string]int{configured: 1, f.home: 2} {
		writeClaudeCLITranscript(t, claudeCLIRootPath(home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
			claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
			claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "end_turn", input, 1, 1),
		})
	}
	f.writeLocal(t, fmt.Sprintf("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n    claude:\n      home: %q\n", filepath.Join(f.project, "absent-codex-store"), configured))
	run := func(t *testing.T, args ...string) int64 {
		t.Helper()
		ledger, out, err := f.generate(t, args...)
		if err != nil {
			t.Fatalf("generate failed: %v %s", err, out)
		}
		return claudeCLISessions(ledger)[claudeCLIRoot].Invocations[0].Tokens.Input
	}
	t.Setenv("CLAUDE_CONFIG_DIR", configured)
	if got := run(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot); got != 2 {
		t.Fatalf("--claude-home must win over local configuration: input %d", got)
	}
	f.writeLocal(t, fmt.Sprintf("telemetry:\n  default_source: claude\n  sources:\n    codex:\n      home: %q\n    claude:\n      home: %q\n", filepath.Join(f.project, "absent-codex-store"), configured))
	t.Setenv("CLAUDE_CONFIG_DIR", f.home)
	if got := run(t, "--session-id", claudeCLIRoot); got != 1 {
		t.Fatalf("configured default must read only the configured store: input %d", got)
	}
}

// A local default source of claude does not change automatic publication of a
// registered Codex owner.
func TestAutomaticGenerationIgnoresClaudeDefaultSource(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 10))
	claudeStore := t.TempDir()
	local := fmt.Sprintf("telemetry:\n  default_source: claude\n  sources:\n    codex:\n      home: %q\n    claude:\n      home: %q\n  capture:\n", f.store, claudeStore)
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	for _, tokens := range []int{10, 31} {
		f.writeAt(t, tokens, "2026-01-01T00:00:03Z")
		command := exec.Command(cliBinaryPath, "generate", "--capture-auto", "--capture-dir", f.state, "--ticket-id", "TEST-100", "--ticket-title", "Automatic Codex owner", "--ticket-type", "feat", "--json-output", f.output)
		command.Dir = f.project
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("automatic publication with a claude default failed: %v %s", err, out)
		}
		ledger := f.ledger(t)
		assertCaptureRootInput(t, ledger, int64(tokens))
		if len(ledger.Sessions) != 1 || ledger.Sessions[0].Source != domain.SourceCodex {
			t.Fatalf("automatic publication changed source ownership: %+v", ledger.Sessions)
		}
	}
}

func TestCaptureInstallRejectsClaudeSource(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	before, err := os.ReadDir(f.project)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(cliBinaryPath, "capture", "install", "--repo", f.project, "--source", "claude", "--store", t.TempDir(), "--dry-run")
	command.Dir = f.project
	out, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "codex or opencode") {
		t.Fatalf("capture install must reject claude: %v %s", err, out)
	}
	after, err := os.ReadDir(f.project)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("rejected installation changed the project: %v -> %v", before, after)
	}
}

func TestInitLocalOnlyWritesClaudeSettings(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	path := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	store := filepath.Join(t.TempDir(), "claude-synthetic-store")
	command := exec.Command(cliBinaryPath, "init", "--local-only", "--capture-enabled", "false", "--source", "claude", "--claude-home", store, "--state-dir", f.stateBase, "--publication-state-dir", filepath.Join(f.project, ".publication-state"))
	command.Dir = f.project
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("local Claude Code setup failed: %v %s", err, out)
	}
	written := string(readCaptureFile(t, path))
	if !strings.Contains(written, "default_source: claude") || !strings.Contains(written, store) || strings.Contains(written, "codex") {
		t.Fatalf("local setup must record only the Claude Code selection:\n%s", written)
	}
	blank := exec.Command(cliBinaryPath, "init", "--local-only", "--source", "claude", "--claude-home", " ")
	blank.Dir = f.project
	if out, err := blank.CombinedOutput(); err == nil || !strings.Contains(string(out), "--claude-home") {
		t.Fatalf("a blank --claude-home must be rejected: %v %s", err, out)
	}
	if string(readCaptureFile(t, path)) != written {
		t.Fatal("a rejected setup changed local configuration")
	}
}

func TestBackfillCLIClaudeSourceSelection(t *testing.T) {
	f := newClaudeCLIFixture(t)
	codexStore := t.TempDir()
	f.writeLocal(t, fmt.Sprintf("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n", codexStore))
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:04Z", "r-1", "model-a", "end_turn", 42, 2, 0),
	})
	version := domain.CurrentSchemaVersion
	ticket, title := "TEST-198", "Claude Code fixture"
	if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.SaveFactLedger(f.output, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{
		{ID: claudeCLIRoot, Source: domain.SourceClaude, User: "local", Invocations: []domain.InvocationFact{{Model: "stale", Tokens: domain.InvocationTokens{Input: 1}}}},
		{ID: "codex-owner", Source: domain.SourceCodex, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "codex-model", Tokens: domain.InvocationTokens{Input: 5}}}},
	}}); err != nil {
		t.Fatal(err)
	}
	backfill := func(env []string, args ...string) (string, error) {
		cmd := exec.Command(cliBinaryPath, append([]string{"backfill", "--ticket-id", ticket}, args...)...)
		cmd.Dir = f.project
		cmd.Env = append(os.Environ(), env...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	before := readCaptureFile(t, f.output)
	for _, tc := range []struct {
		name string
		env  []string
		args []string
		want string
	}{
		{name: "all without a Claude Code store", env: []string{"CLAUDE_CONFIG_DIR=" + f.home}, args: []string{"--source", "all", "--recompute-db"}, want: "--claude-home"},
		{name: "all with a missing store", args: []string{"--source", "all", "--recompute-db", "--claude-home", filepath.Join(f.project, "missing-claude-store")}, want: "missing-claude-store"},
		{name: "blank store", args: []string{"--source", "claude", "--recompute-db", "--claude-home", ""}, want: "--claude-home must be nonblank"},
		{name: "unsupported source", args: []string{"--source", "claude-code", "--recompute-db"}, want: "'claude'"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := backfill(tc.env, tc.args...)
			if err == nil || !strings.Contains(out, tc.want) {
				t.Fatalf("want failure naming %q: %v %s", tc.want, err, out)
			}
			if !bytes.Equal(before, readCaptureFile(t, f.output)) {
				t.Fatal("a rejected backfill changed the ledger")
			}
		})
	}
	help, err := exec.Command(cliBinaryPath, "backfill", "--help").CombinedOutput()
	if err != nil || !strings.Contains(string(help), "--claude-home") || !strings.Contains(string(help), "all (OpenCode, Codex and Claude Code)") || !strings.Contains(string(help), "both (OpenCode and Codex)") {
		t.Fatalf("backfill help must describe the Claude Code store and source values: %v %s", err, help)
	}
	out, err := backfill(nil, "--source", "claude", "--recompute-db", "--claude-home", f.home)
	if err != nil {
		t.Fatalf("Claude Code backfill failed: %v %s", err, out)
	}
	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal(readCaptureFile(t, f.output), &ledger); err != nil {
		t.Fatal(err)
	}
	sessions := claudeCLISessions(ledger)
	if got := sessions[claudeCLIRoot].Invocations; len(got) != 1 || got[0].Model != "model-a" || got[0].Tokens.Input != 42 || got[0].ActiveDurationSeconds != 4 {
		t.Fatalf("Claude Code owner not refreshed: %+v\n%s", sessions[claudeCLIRoot], out)
	}
	if got := sessions["codex-owner"].Invocations; len(got) != 1 || got[0].Model != "codex-model" || got[0].Tokens.Input != 5 {
		t.Fatalf("Codex owner changed during a Claude Code backfill: %+v", sessions["codex-owner"])
	}
}

// projectGitUser is the repository-local user.name that initCLIGitFixture
// writes into every generate fixture project.
const projectGitUser = "CLI Fixture"

// writeClaudeCLITree stores a root under folder with one proven planner child,
// one subagent without proof, and a literal nonce only in the child's prompt.
func writeClaudeCLITree(t *testing.T, home, folder, cwd, nonce string) {
	t.Helper()
	writeClaudeCLITranscript(t, claudeCLIRootPath(home, folder, claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(cwd, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "tool_use", 3, 1, 0, claudeCLIAgentCall("call-1", "planner")),
		claudeCLIAgentResult("2026-01-01T10:00:20Z", "call-1", "achild"),
		claudeCLIResponse("2026-01-01T10:00:30Z", "r-2", "model-a", "end_turn", 4, 1, 0),
	})
	subagents := filepath.Join(home, "projects", folder, claudeCLIRoot, "subagents")
	childPrompt := claudeCLIPrompt(cwd, "2026-01-01T10:00:02Z")
	childPrompt.PromptSource = ""
	childPrompt.Message.Content = claudeCLIContent{Text: "synthetic " + nonce}
	writeClaudeCLITranscript(t, filepath.Join(subagents, "agent-achild.jsonl"), claudeCLIRoot, "achild", []claudeCLIRecord{
		childPrompt,
		claudeCLIResponse("2026-01-01T10:00:12Z", "r-c1", "model-a", "end_turn", 7, 1, 0),
	})
	writeClaudeCLITranscript(t, filepath.Join(subagents, "agent-astray.jsonl"), claudeCLIRoot, "astray", []claudeCLIRecord{
		claudeCLIPrompt(cwd, "2026-01-01T10:00:03Z"),
		claudeCLIResponse("2026-01-01T10:00:04Z", "r-s1", "model-a", "end_turn", 9, 1, 0),
	})
	for agent, toolUse := range map[string]string{"achild": "call-1", "astray": "call-not-in-tree"} {
		meta := []byte(`{"agentType":"planner","toolUseId":"` + toolUse + `"}`)
		if err := os.WriteFile(filepath.Join(subagents, "agent-"+agent+".meta.json"), meta, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerateClaudeSelectsByNonceAndAgentID(t *testing.T) {
	const nonce = "synthetic-nonce-5d1e"
	for _, tc := range []struct {
		name      string
		selection []string
	}{
		{name: "literal nonce in a linked-worktree folder", selection: []string{"--nonce", nonce}},
		{name: "proven agent ID", selection: []string{"--session-id", "achild"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClaudeCLIFixture(t)
			// The folder name is unrelated to the project; identity comes from cwd.
			writeClaudeCLITree(t, f.home, "-elsewhere-linked-worktree-folder", f.project, nonce)
			ledger, out, err := f.generate(t, append([]string{"--source", "claude", "--claude-home", f.home}, tc.selection...)...)
			if err != nil {
				t.Fatalf("generate failed: %v %s", err, out)
			}
			sessions := claudeCLISessions(ledger)
			root, child := sessions[claudeCLIRoot], sessions["achild"]
			if len(sessions) != 2 || len(root.Invocations) != 1 || root.Invocations[0].Tokens.Input != 7 {
				t.Fatalf("the selection must publish the root tree only: %+v", ledger.Sessions)
			}
			if child.ParentID == nil || *child.ParentID != claudeCLIRoot || child.Invocations[0].Role != "planner" || child.Invocations[0].Tokens.Input != 7 || child.User != projectGitUser {
				t.Fatalf("proven child: %+v", child)
			}
			if !strings.Contains(out, "astray unresolved") {
				t.Fatalf("the unproven subagent must be reported: %s", out)
			}
		})
	}
	t.Run("unproven agent ID", func(t *testing.T) {
		f := newClaudeCLIFixture(t)
		writeClaudeCLITree(t, f.home, "-elsewhere-linked-worktree-folder", f.project, nonce)
		_, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", "astray")
		if err == nil || !strings.Contains(out, "astray is not a proven subagent") {
			t.Fatalf("an unproven agent ID must be rejected: %v %s", err, out)
		}
		if _, statErr := os.Stat(f.output); !os.IsNotExist(statErr) {
			t.Fatalf("a rejected agent selection published its root tree: %v", statErr)
		}
	})
}

// A root already recorded for the ticket refreshes even when its recorded cwd
// is no longer inside this project (moved checkout or clone at another path);
// a new root with that cwd is still rejected.
func TestGenerateClaudeRecordedRootRefreshesAfterCheckoutMove(t *testing.T) {
	f := newClaudeCLIFixture(t)
	const original = "/synthetic/original-checkout"
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-original-checkout", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(original, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:06Z", "r-1", "model-a", "end_turn", 21, 2, 0),
	})
	version := domain.CurrentSchemaVersion
	ticket, title := "TEST-198", "Claude Code fixture"
	if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.SaveFactLedger(f.output, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{
		{ID: claudeCLIRoot, Source: domain.SourceClaude, User: "Recorded Developer", Invocations: []domain.InvocationFact{{Model: "stale-model", Tokens: domain.InvocationTokens{Input: 1}}}},
	}}); err != nil {
		t.Fatal(err)
	}
	ledger, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
	if err != nil {
		t.Fatalf("a recorded root must refresh after its checkout moved: %v %s", err, out)
	}
	root := claudeCLISessions(ledger)[claudeCLIRoot]
	if len(root.Invocations) != 1 || root.Invocations[0].Tokens.Input != 21 || root.Invocations[0].ActiveDurationSeconds != 6 {
		t.Fatalf("recorded root not refreshed: %+v", root)
	}
	if root.User != "Recorded Developer" {
		t.Fatalf("a refreshed owner keeps its recorded user: %q", root.User)
	}
	const newRoot = "claude-cli-new-root"
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-original-checkout", newRoot), newRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(original, "2026-01-01T11:00:00Z"),
		claudeCLIResponse("2026-01-01T11:00:01Z", "r-1", "model-a", "end_turn", 1, 1, 0),
	})
	before := readCaptureFile(t, f.output)
	if _, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", newRoot); err == nil || !strings.Contains(out, "Source ownership conflict") {
		t.Fatalf("a new root outside the project must still be rejected: %v %s", err, out)
	}
	if !bytes.Equal(before, readCaptureFile(t, f.output)) {
		t.Fatal("a rejected new root changed the ledger")
	}
}

// seedClaudeCLILedger publishes a stored ledger for the fixture ticket.
func seedClaudeCLILedger(t *testing.T, f claudeCLIFixture, sessions []domain.SessionFact) domain.AgenticFactLedger {
	t.Helper()
	version := domain.CurrentSchemaVersion
	ticket, title := "TEST-198", "Claude Code fixture"
	if err := os.MkdirAll(filepath.Dir(f.output), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.SaveFactLedger(f.output, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: sessions}); err != nil {
		t.Fatal(err)
	}
	var stored domain.AgenticFactLedger
	if err := json.Unmarshal(readCaptureFile(t, f.output), &stored); err != nil {
		t.Fatal(err)
	}
	return stored
}

// A complete read that replaces a recorded root with fewer rows, or with a
// lower input+output+cache-read total, is reported; each condition alone
// triggers the notice and a larger replacement is silent.
func TestGenerateClaudeSmallerReplacementNotice(t *testing.T) {
	// The fresh root is one model-a row with 100+10+20 = 130 tokens.
	for _, tc := range []struct {
		name   string
		stored []domain.InvocationFact
		want   string
	}{
		{name: "fewer rows only", stored: []domain.InvocationFact{{Model: "model-a", Tokens: domain.InvocationTokens{Input: 1}}, {Model: "model-b", Tokens: domain.InvocationTokens{Input: 1}}},
			want: "Claude Code owner " + claudeCLIRoot + " replaced by a complete read with smaller totals: invocations 2 -> 1, tokens 2 -> 130"},
		{name: "lower tokens only", stored: []domain.InvocationFact{{Model: "model-a", Tokens: domain.InvocationTokens{Input: 9000, Output: 900, CacheRead: 100}}},
			want: "Claude Code owner " + claudeCLIRoot + " replaced by a complete read with smaller totals: invocations 1 -> 1, tokens 10000 -> 130"},
		{name: "larger replacement", stored: []domain.InvocationFact{{Model: "model-a", Tokens: domain.InvocationTokens{Input: 1}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClaudeCLIFixture(t)
			writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
				claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
				claudeCLIResponse("2026-01-01T10:00:05Z", "r-1", "model-a", "end_turn", 100, 10, 20),
			})
			seedClaudeCLILedger(t, f, []domain.SessionFact{{ID: claudeCLIRoot, Source: domain.SourceClaude, User: "local", Invocations: tc.stored}})
			ledger, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
			if err != nil {
				t.Fatalf("generate failed: %v %s", err, out)
			}
			if got := claudeCLISessions(ledger)[claudeCLIRoot].Invocations; len(got) != 1 || got[0].Tokens.Input != 100 {
				t.Fatalf("the complete read must replace the stored snapshot: %+v", got)
			}
			if tc.want == "" {
				if strings.Contains(out, "replaced by a complete read") {
					t.Fatalf("a larger replacement must not be reported: %s", out)
				}
				return
			}
			if strings.Count(out, "replaced by a complete read") != 1 || !strings.Contains(out, tc.want) {
				t.Fatalf("want one notice %q in:\n%s", tc.want, out)
			}
		})
	}
}

// A child retained because its recorded parent changed keeps its whole stored
// snapshot: fresh events from its transcript must not add anomaly summaries or
// checkpoint feedback to it.
func TestGenerateClaudeRetainedChildKeepsStoredSnapshot(t *testing.T) {
	f := newClaudeCLIFixture(t)
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:01Z", "r-1", "model-a", "tool_use", 3, 1, 0, claudeCLIAgentCall("call-1", "planner")),
		claudeCLIAgentResult("2026-01-01T10:00:20Z", "call-1", "achild"),
		claudeCLIResponse("2026-01-01T10:00:30Z", "r-2", "model-a", "end_turn", 4, 1, 0),
	})
	// Six consecutive failing Bash calls, then the child stays open: enough for
	// streak, error-rate and thrashing anomalies and an open-turn interval.
	child := []claudeCLIRecord{claudeCLIPrompt(f.project, "2026-01-01T10:00:02Z")}
	child[0].PromptSource = ""
	for i := 0; i < 6; i++ {
		id := fmt.Sprintf("bash-%d", i)
		at := fmt.Sprintf("2026-01-01T10:00:%02d", 3+2*i)
		child = append(child,
			claudeCLIResponse(at+"Z", "r-c"+id, "model-a", "tool_use", 1, 1, 0, claudeCLIBlock{Type: "tool_use", ID: id, Name: "Bash", Input: &claudeCLIInput{}}),
			claudeCLIRecord{Type: "user", Timestamp: fmt.Sprintf("2026-01-01T10:00:%02dZ", 4+2*i), Message: claudeCLIMessage{Role: "user", Content: claudeCLIContent{Blocks: []claudeCLIBlock{{Type: "tool_result", ToolUseID: id, IsError: true}}}}})
	}
	subagents := filepath.Join(f.home, "projects", "-synthetic-project", claudeCLIRoot, "subagents")
	writeClaudeCLITranscript(t, filepath.Join(subagents, "agent-achild.jsonl"), claudeCLIRoot, "achild", child)
	if err := os.WriteFile(filepath.Join(subagents, "agent-achild.meta.json"), []byte(`{"agentType":"planner","toolUseId":"call-1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	elsewhere := "recorded-elsewhere"
	stored := seedClaudeCLILedger(t, f, []domain.SessionFact{
		{ID: claudeCLIRoot, Source: domain.SourceClaude, User: "local", Invocations: []domain.InvocationFact{{Model: "model-a", Tokens: domain.InvocationTokens{Input: 1}}}},
		{ID: "achild", ParentID: &elsewhere, Source: domain.SourceClaude, User: "remote", Anomalies: []domain.Anomaly{},
			Skills: []domain.SkillFact{{Name: "stored-skill"}}, Tools: []domain.SessionToolFact{{Name: "Read", Calls: 2}},
			Invocations: []domain.InvocationFact{{Role: "planner", Model: "model-a", ActiveDurationSeconds: 500, ProvisionalSeconds: 500, Tokens: domain.InvocationTokens{Input: 77}}}},
	})
	ledger, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
	if err != nil {
		t.Fatalf("generate failed: %v %s", err, out)
	}
	after := claudeCLISessions(ledger)
	if !reflect.DeepEqual(after["achild"], claudeCLISessions(stored)["achild"]) {
		t.Fatalf("the retained child changed:\nbefore %+v\nafter  %+v", claudeCLISessions(stored)["achild"], after["achild"])
	}
	if !strings.Contains(out, "Claude Code owner achild retained: source parent linkage changed from recorded-elsewhere to "+claudeCLIRoot) {
		t.Fatalf("the retention must be reported: %s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "session=achild") && !strings.Contains(line, "native interval evidence is unavailable") {
			t.Fatalf("checkpoint feedback for the retained child used fresh source evidence: %s", line)
		}
	}
	if root := after[claudeCLIRoot]; len(root.Invocations) != 1 || root.Invocations[0].Tokens.Input != 7 {
		t.Fatalf("the root must still refresh: %+v", root)
	}
}

// Fresh source events reach the publication boundary: a complete Claude Code
// owner gets a recorded (empty) anomaly summary, and an open turn produces
// checkpoint feedback that cites the native interval.
func TestGenerateClaudeSourceEventsReachSummariesAndCheckpointFeedback(t *testing.T) {
	f := newClaudeCLIFixture(t)
	writeClaudeCLITranscript(t, claudeCLIRootPath(f.home, "-synthetic-project", claudeCLIRoot), claudeCLIRoot, "", []claudeCLIRecord{
		claudeCLIPrompt(f.project, "2026-01-01T10:00:00Z"),
		claudeCLIResponse("2026-01-01T10:00:05Z", "r-1", "model-a", "tool_use", 2, 1, 0),
	})
	ledger, out, err := f.generate(t, "--source", "claude", "--claude-home", f.home, "--session-id", claudeCLIRoot)
	if err != nil {
		t.Fatalf("generate failed: %v %s", err, out)
	}
	root := claudeCLISessions(ledger)[claudeCLIRoot]
	if root.Anomalies == nil || len(root.Anomalies) != 0 {
		t.Fatalf("a complete clean owner must record an empty anomaly summary, not leave it absent: %#v", root.Anomalies)
	}
	want := "AgenticLens checkpoint: source=claude; session=" + claudeCLIRoot + "; invocation=1"
	if !strings.Contains(out, want) || !strings.Contains(out, "Source reports open turn u-2026-01-01T10:00:00Z, started at 2026-01-01T10:00:00Z") {
		t.Fatalf("the open turn must be reported from native interval evidence:\n%s", out)
	}
}
