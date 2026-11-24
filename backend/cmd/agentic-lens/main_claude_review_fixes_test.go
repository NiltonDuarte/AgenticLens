package main_test

// Synthetic regression fixtures for AL-98 review findings DATA-01 (generate
// store check), DATA-02 and DATA-03 on the generate path. Stores are temp
// directories holding native record keys only; no transcript content,
// ~/.claude or live ticket workspace is read.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

const (
	reviewCLIRoot  = "review-cli-root"
	reviewCLIChild = "areviewclichild"
)

type reviewCLIFixture struct {
	project, home, output string
}

func newReviewCLIFixture(t *testing.T) reviewCLIFixture {
	t.Helper()
	f := reviewCLIFixture{project: t.TempDir(), home: t.TempDir()}
	f.output = filepath.Join(f.project, "records", "agentic_metrics.json")
	shared := "agenticlens:\n  version: v9.8.7\ndatabase:\n  busy_timeout_ms: 0\nlive:\n  active_threshold_seconds: 300\n  poll_interval_ms: 2000\ngeneration:\n  metrics:\n    skills: true\n    delegations: true\n    anomalies: true\n  feedback:\n    provisional_threshold_seconds: 180\n  summary:\n    enabled: false\n    filename: SUMMARY.md\n    capture_observation: false\nrole_order: []\nrole_mappings: {}\nproject_name: ReviewFixture\nticket:\n  case_insensitive: true\n  prefixes:\n    - prefix: TEST\n      description: Fixture tickets\nworkspaces:\n  artifacts: []\n  roots:\n    - path: records\n      discovery:\n        max_depth: 0\n        max_directories: 300\ncode_metrics:\n  base_branch: fixture-base\n  production:\n    include: [\"src/**\"]\n    exclude: ['**/*_test.go']\n  tests:\n    include: ['**/*_test.go']\n    exclude: []\nanomaly_thresholds:\n  stall_threshold_seconds: 120\n  consecutive_errors_streak_warning: 4\n  consecutive_errors_streak_critical: 8\n  turn_usage_warning_tokens: 60000\n  turn_usage_critical_tokens: 120000\n  context_growth_warning_tokens: 80000\n  context_growth_critical_tokens: 120000\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	local := "telemetry:\n  default_source: claude\n  capture:\n    enabled: false\n    state_dir: .agenticlens-capture\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n" + fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(f.project, ".publication-state"))
	for name, body := range map[string]string{".agenticlens-config.yaml": shared, ".agenticlens-config.local.yaml": local} {
		if err := os.WriteFile(filepath.Join(f.project, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

// Typed synthetic Claude Code records mirroring the tester's fixtures in
// main_claude_test.go, written with this fixture's project as cwd.
type reviewCLIRecord struct {
	Type          string            `json:"type"`
	SessionID     string            `json:"sessionId"`
	AgentID       string            `json:"agentId,omitempty"`
	IsSidechain   bool              `json:"isSidechain,omitempty"`
	UUID          string            `json:"uuid,omitempty"`
	Timestamp     string            `json:"timestamp"`
	CWD           string            `json:"cwd"`
	PromptSource  string            `json:"promptSource,omitempty"`
	RequestID     string            `json:"requestId,omitempty"`
	ToolUseResult *reviewCLIResult  `json:"toolUseResult,omitempty"`
	Message       *reviewCLIMessage `json:"message,omitempty"`
}

type reviewCLIMessage struct {
	ID         string          `json:"id,omitempty"`
	Model      string          `json:"model,omitempty"`
	Role       string          `json:"role,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	Usage      *reviewCLIUsage `json:"usage,omitempty"`
	Content    json.RawMessage `json:"content"`
}

type reviewCLIUsage struct {
	Input     int `json:"input_tokens"`
	Output    int `json:"output_tokens"`
	CacheRead int `json:"cache_read_input_tokens"`
}

type reviewCLIBlock struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     *reviewCLIInput `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

type reviewCLIInput struct {
	SubagentType string `json:"subagent_type"`
}

type reviewCLIResult struct {
	AgentID string `json:"agentId"`
}

func reviewCLIJSON(t *testing.T, data []byte, err error) json.RawMessage {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func reviewCLIText(t *testing.T, value string) json.RawMessage {
	data, err := json.Marshal(value)
	return reviewCLIJSON(t, data, err)
}

func reviewCLIBlocks(t *testing.T, value []reviewCLIBlock) json.RawMessage {
	data, err := json.Marshal(value)
	return reviewCLIJSON(t, data, err)
}

// writeStore writes a root that delegates to one proven child.
func (f reviewCLIFixture) writeStore(t *testing.T) {
	t.Helper()
	folder := filepath.Join(f.home, "projects", "-review-cli")
	prompt := func(at, uuid, source string) reviewCLIRecord {
		return reviewCLIRecord{Type: "user", UUID: uuid, Timestamp: at, PromptSource: source, Message: &reviewCLIMessage{Role: "user", Content: reviewCLIText(t, "synthetic")}}
	}
	response := func(at, request, stop string, blocks []reviewCLIBlock) reviewCLIRecord {
		if blocks == nil {
			blocks = []reviewCLIBlock{}
		}
		return reviewCLIRecord{Type: "assistant", Timestamp: at, RequestID: request, Message: &reviewCLIMessage{ID: "m-" + request, Model: "model-review", StopReason: stop,
			Usage: &reviewCLIUsage{Input: 2, Output: 1}, Content: reviewCLIBlocks(t, blocks)}}
	}
	write := func(path string, lines []reviewCLIRecord, agent string) {
		var body strings.Builder
		for _, line := range lines {
			line.SessionID, line.CWD = reviewCLIRoot, f.project
			if agent != "" {
				line.AgentID, line.IsSidechain = agent, true
			}
			encoded, err := json.Marshal(line)
			body.Write(reviewCLIJSON(t, encoded, err))
			body.WriteByte('\n')
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body.String()), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(folder, reviewCLIRoot+".jsonl"), []reviewCLIRecord{
		prompt("2026-01-01T10:00:00Z", "u-1", "sdk"),
		response("2026-01-01T10:00:01Z", "r-1", "tool_use", []reviewCLIBlock{{Type: "tool_use", ID: "call-1", Name: "Agent", Input: &reviewCLIInput{SubagentType: "reviewer"}}}),
		{Type: "user", Timestamp: "2026-01-01T10:00:05Z", ToolUseResult: &reviewCLIResult{AgentID: reviewCLIChild}, Message: &reviewCLIMessage{Role: "user", Content: reviewCLIBlocks(t, []reviewCLIBlock{{Type: "tool_result", ToolUseID: "call-1"}})}},
		response("2026-01-01T10:00:06Z", "r-2", "end_turn", nil),
	}, "")
	subagents := filepath.Join(folder, reviewCLIRoot, "subagents")
	write(filepath.Join(subagents, "agent-"+reviewCLIChild+".jsonl"), []reviewCLIRecord{
		prompt("2026-01-01T10:00:02Z", "c-1", ""),
		response("2026-01-01T10:00:04Z", "r-c1", "end_turn", nil),
	}, reviewCLIChild)
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+reviewCLIChild+".meta.json"), []byte(`{"agentType":"reviewer","toolUseId":"call-1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
}

func (f reviewCLIFixture) generate(t *testing.T, home string) (string, error) {
	t.Helper()
	cmd := exec.Command(cliBinaryPath, "generate", "--ticket-id", "TEST-299", "--ticket-title", "Review fixture", "--ticket-type", "fix",
		"--json-output", f.output, "--source", "claude", "--claude-home", home, "--session-id", reviewCLIRoot)
	cmd.Dir = f.project
	prepareScopedGenerateFixture(t, cmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (f reviewCLIFixture) readLedger(t *testing.T) domain.AgenticFactLedger {
	t.Helper()
	data, err := os.ReadFile(f.output)
	if err != nil {
		t.Fatal(err)
	}
	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	return ledger
}

func reviewSessionIndex(t *testing.T, ledger domain.AgenticFactLedger, id string) int {
	t.Helper()
	for i, session := range ledger.Sessions {
		if session.ID == id {
			return i
		}
	}
	t.Fatalf("session %s missing from ledger", id)
	return -1
}

func TestGenerateClaudeStoreWithoutProjectsFailsBeforePublication(t *testing.T) {
	f := newReviewCLIFixture(t)
	notStore := t.TempDir()
	out, err := f.generate(t, notStore)
	if err == nil || !strings.Contains(out, "projects") || !strings.Contains(out, "--claude-home") {
		t.Fatalf("a store without projects/ must fail with an actionable error: %v %s", err, out)
	}
	if _, statErr := os.Stat(f.output); !os.IsNotExist(statErr) {
		t.Fatalf("no ledger may be published when the store is rejected: %v", statErr)
	}
}

func TestGenerateClaudeRetainsChangedParentAndReportsSmallerReplacement(t *testing.T) {
	f := newReviewCLIFixture(t)
	f.writeStore(t)
	if out, err := f.generate(t, f.home); err != nil {
		t.Fatalf("initial generate failed: %v %s", err, out)
	}
	ledger := f.readLedger(t)
	// The stored child now names a different parent and carries 99 tokens; the
	// stored root gains a second row (500 input, 2 output), so its stored total
	// is 6 + 502 = 508 tokens against a fresh total of 6.
	child := &ledger.Sessions[reviewSessionIndex(t, ledger, reviewCLIChild)]
	differentParent := "a-different-parent"
	child.ParentID = &differentParent
	child.Invocations[0].Tokens.Input = 99
	root := &ledger.Sessions[reviewSessionIndex(t, ledger, reviewCLIRoot)]
	extra := root.Invocations[0]
	extra.Model = "model-stored-extra"
	extra.Tokens.Input = 500
	root.Invocations = append(root.Invocations, extra)
	if err := adapters.SaveFactLedger(f.output, ledger); err != nil {
		t.Fatal(err)
	}

	out, err := f.generate(t, f.home)
	if err != nil {
		t.Fatalf("refresh failed: %v %s", err, out)
	}
	if !strings.Contains(out, "Claude Code owner "+reviewCLIChild+" retained: source parent linkage changed from a-different-parent to "+reviewCLIRoot) {
		t.Fatalf("the relinked child must be retained with a diagnostic: %s", out)
	}
	if !strings.Contains(out, "Claude Code owner "+reviewCLIRoot+" replaced by a complete read with smaller totals: invocations 2 -> 1, tokens 508 -> 6") {
		t.Fatalf("the smaller complete replacement must be reported: %s", out)
	}
	after := f.readLedger(t)
	storedChild := after.Sessions[reviewSessionIndex(t, after, reviewCLIChild)]
	if storedChild.ParentID == nil || *storedChild.ParentID != differentParent || storedChild.Invocations[0].Tokens.Input != 99 {
		t.Fatalf("the relinked child must keep its stored snapshot: %+v", storedChild)
	}
	if rows := after.Sessions[reviewSessionIndex(t, after, reviewCLIRoot)].Invocations; len(rows) != 1 {
		t.Fatalf("the complete root read must replace the stored snapshot: %d rows", len(rows))
	}
}
