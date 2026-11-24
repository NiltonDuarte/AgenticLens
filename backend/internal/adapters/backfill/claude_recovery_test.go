package backfill

// All Claude Code stores in this file are synthetic fixtures written to temp
// directories. They reproduce native record keys without transcript content and
// never read ~/.claude, CLAUDE_CONFIG_DIR or a live ticket workspace. Expected
// values are hand-computed from the usage and timestamps written below.

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

const (
	claudeFixtureRoot  = "claude-root-synthetic"
	claudeFixtureChild = "aplanner"
)

// Typed synthetic Claude Code records. Only native keys are written.
type claudeRecord struct {
	Type          string        `json:"type"`
	SessionID     string        `json:"sessionId"`
	AgentID       string        `json:"agentId,omitempty"`
	IsSidechain   bool          `json:"isSidechain,omitempty"`
	UUID          string        `json:"uuid,omitempty"`
	Timestamp     string        `json:"timestamp"`
	CWD           string        `json:"cwd"`
	PromptSource  string        `json:"promptSource,omitempty"`
	RequestID     string        `json:"requestId,omitempty"`
	ToolUseResult *claudeResult `json:"toolUseResult,omitempty"`
	Message       claudeMessage `json:"message"`
}

type claudeMessage struct {
	ID         string          `json:"id,omitempty"`
	Model      string          `json:"model,omitempty"`
	Role       string          `json:"role,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	Usage      *claudeUsage    `json:"usage,omitempty"`
	Content    json.RawMessage `json:"content"`
}

type claudeUsage struct {
	Input     int `json:"input_tokens"`
	Output    int `json:"output_tokens"`
	CacheRead int `json:"cache_read_input_tokens"`
}

type claudeBlock struct {
	Type      string       `json:"type"`
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Input     *claudeInput `json:"input,omitempty"`
	ToolUseID string       `json:"tool_use_id,omitempty"`
}

type claudeInput struct {
	SubagentType string `json:"subagent_type"`
	Description  string `json:"description"`
}

type claudeResult struct {
	AgentID   string `json:"agentId"`
	AgentType string `json:"agentType"`
}

func claudeText(t *testing.T, text string) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(text)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func claudeBlocks(t *testing.T, blocks []claudeBlock) json.RawMessage {
	t.Helper()
	if blocks == nil {
		blocks = []claudeBlock{}
	}
	data, err := json.Marshal(blocks)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeClaudeLines(t *testing.T, path, sessionID, agentID string, lines []claudeRecord) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	for _, line := range lines {
		line.SessionID, line.CWD = sessionID, "/synthetic/project"
		if agentID != "" {
			line.AgentID, line.IsSidechain = agentID, true
		}
		encoded, err := json.Marshal(line)
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

// claudeRootRecords is one 30s root turn with usage (100,10,5)+(200,20,0) and
// an Agent call whose result names claudeFixtureChild. Without usage, the
// final response carries no usage object, which makes the root incomplete.
func claudeRootRecords(t *testing.T, finalUsage bool) []claudeRecord {
	t.Helper()
	final := &claudeUsage{Input: 200, Output: 20}
	if !finalUsage {
		final = nil
	}
	return []claudeRecord{
		{Type: "user", UUID: "u-1", Timestamp: "2026-01-01T00:00:00Z", PromptSource: "sdk", Message: claudeMessage{Role: "user", Content: claudeText(t, "synthetic")}},
		{Type: "assistant", Timestamp: "2026-01-01T00:00:01Z", RequestID: "r-1", Message: claudeMessage{ID: "m-1", Model: "claude-synthetic-model", StopReason: "tool_use", Usage: &claudeUsage{Input: 100, Output: 10, CacheRead: 5},
			Content: claudeBlocks(t, []claudeBlock{{Type: "tool_use", ID: "call-1", Name: "Agent", Input: &claudeInput{SubagentType: "planner", Description: "plan"}}})}},
		{Type: "user", Timestamp: "2026-01-01T00:00:20Z", ToolUseResult: &claudeResult{AgentID: claudeFixtureChild, AgentType: "planner"},
			Message: claudeMessage{Role: "user", Content: claudeBlocks(t, []claudeBlock{{Type: "tool_result", ToolUseID: "call-1"}})}},
		{Type: "assistant", Timestamp: "2026-01-01T00:00:30Z", RequestID: "r-2", Message: claudeMessage{ID: "m-2", Model: "claude-synthetic-model", StopReason: "end_turn", Usage: final, Content: claudeBlocks(t, nil)}},
	}
}

// writeClaudeTree writes root rootID with claudeRootRecords and, when asked,
// its proven planner child with usage (40,4,1) and one 10s turn.
func writeClaudeTree(t *testing.T, home, rootID string, withChild, rootUsage bool) {
	t.Helper()
	folder := filepath.Join(home, "projects", "-synthetic-project")
	writeClaudeLines(t, filepath.Join(folder, rootID+".jsonl"), rootID, "", claudeRootRecords(t, rootUsage))
	if !withChild {
		return
	}
	subagents := filepath.Join(folder, rootID, "subagents")
	writeClaudeLines(t, filepath.Join(subagents, "agent-"+claudeFixtureChild+".jsonl"), rootID, claudeFixtureChild, []claudeRecord{
		{Type: "user", UUID: "c-1", Timestamp: "2026-01-01T00:00:02Z", Message: claudeMessage{Role: "user", Content: claudeText(t, "synthetic")}},
		{Type: "assistant", Timestamp: "2026-01-01T00:00:12Z", RequestID: "r-c1", Message: claudeMessage{ID: "m-c1", Model: "claude-synthetic-model", StopReason: "end_turn", Usage: &claudeUsage{Input: 40, Output: 4, CacheRead: 1}, Content: claudeBlocks(t, nil)}},
	})
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+claudeFixtureChild+".meta.json"), []byte(`{"agentType":"planner","toolUseId":"call-1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
}

// writeClaudeFixtureStore writes root claudeFixtureRoot with one proven planner
// child. Root: usage (100,10,5)+(200,20,0), one 30s turn. Child: usage
// (40,4,1), one 10s turn.
func writeClaudeFixtureStore(t *testing.T, home string, withChild bool) {
	t.Helper()
	writeClaudeTree(t, home, claudeFixtureRoot, withChild, true)
}

func claudeStoredLedger(ticket string, extra ...domain.SessionFact) domain.AgenticFactLedger {
	version := domain.CurrentSchemaVersion
	root := claudeFixtureRoot
	sessions := []domain.SessionFact{
		{ID: claudeFixtureRoot, Source: domain.SourceClaude, User: "local", Invocations: []domain.InvocationFact{{Model: "stale-model", Tokens: domain.InvocationTokens{Input: 1}}}},
		{ID: claudeFixtureChild, ParentID: &root, Source: domain.SourceClaude, User: "local", Invocations: []domain.InvocationFact{{Role: "planner", Model: "stale-model", Tokens: domain.InvocationTokens{Input: 2}}}},
		{ID: "codex-owner", Source: domain.SourceCodex, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-codex"}}, Tools: []domain.SessionToolFact{{Name: "exec", Calls: 3}}},
		{ID: "opencode-owner", Source: domain.SourceOpenCode, User: "remote", Invocations: []domain.InvocationFact{{Role: "tester", Model: "stored-opencode"}}, Tools: []domain.SessionToolFact{{Name: "bash", Calls: 10}}},
		{ID: "claude-other-machine", Source: domain.SourceClaude, User: "remote", Skills: []domain.SkillFact{{Name: "remote-skill"}}, Invocations: []domain.InvocationFact{{Role: "", Model: "remote-model", ActiveDurationSeconds: 99}}},
	}
	return domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: append(sessions, extra...)}
}

func sessionsByID(t *testing.T, workspace string) map[string]domain.SessionFact {
	t.Helper()
	dto, _ := readRecoveryLedger(t, workspace)
	byID := map[string]domain.SessionFact{}
	for _, session := range dto.ToDomain().Sessions {
		if _, duplicate := byID[session.ID]; duplicate {
			t.Fatalf("duplicate session %s", session.ID)
		}
		byID[session.ID] = session
	}
	return byID
}

func assertClaudeFixtureRefreshed(t *testing.T, byID map[string]domain.SessionFact) {
	t.Helper()
	root, child := byID[claudeFixtureRoot], byID[claudeFixtureChild]
	if len(root.Invocations) != 1 || root.Source != domain.SourceClaude {
		t.Fatalf("root not refreshed: %+v", root)
	}
	wantRoot := domain.InvocationTokens{Input: 300, Output: 30, CacheRead: 5, PeakContext: 200}
	if got := root.Invocations[0]; got.Model != "claude-synthetic-model" || got.Tokens != wantRoot || got.ActiveDurationSeconds != 30 || got.Turns != 1 || got.UserIterations != 1 || got.CostEvidence.State != domain.CostUnavailable {
		t.Fatalf("root refreshed with wrong facts: %+v", got)
	}
	wantChild := domain.InvocationTokens{Input: 40, Output: 4, CacheRead: 1, PeakContext: 41}
	if len(child.Invocations) != 1 || child.ParentID == nil || *child.ParentID != claudeFixtureRoot || child.Invocations[0].Role != "planner" || child.Invocations[0].Tokens != wantChild || child.Invocations[0].ActiveDurationSeconds != 10 {
		t.Fatalf("child refreshed with wrong facts: %+v", child)
	}
	if len(root.Delegations) != 1 || root.Delegations[0].ChildSession != claudeFixtureChild {
		t.Fatalf("root delegation lost its proven child: %+v", root.Delegations)
	}
}

func overwriteBackfillLocal(t *testing.T, root, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestBackfillClaudeExactRefreshPreservesOtherOwnersAndRepeatsEqually(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-REFRESH"
	workspace := createTestWorkspace(t, root, ticket, claudeStoredLedger(ticket))
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	opts := BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true}
	var previous map[string]domain.SessionFact
	for pass := 0; pass < 2; pass++ {
		report, err := NewEngine().Execute(opts)
		if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
			t.Fatalf("pass %d failed: %+v %v", pass, report, err)
		}
		result := report.Results[0]
		if !result.RefreshIncomplete || result.DBFound || !strings.Contains(strings.Join(result.Diagnostics, "\n"), "claude-other-machine") {
			t.Fatalf("an owner without a local transcript must keep the refresh incomplete and be named: %+v", result)
		}
		after := sessionsByID(t, workspace)
		if len(after) != len(before) {
			t.Fatalf("session count changed: %d -> %d", len(before), len(after))
		}
		assertClaudeFixtureRefreshed(t, after)
		for _, id := range []string{"codex-owner", "opencode-owner", "claude-other-machine"} {
			if !reflect.DeepEqual(after[id], before[id]) {
				t.Fatalf("%s changed during a Claude Code refresh:\nbefore %+v\nafter  %+v", id, before[id], after[id])
			}
		}
		if previous != nil && !reflect.DeepEqual(previous, after) {
			t.Fatal("repeat on an unchanged store changed facts")
		}
		previous = after
	}
}

func TestBackfillClaudeRetainsChangedParentLinkAndUnavailableChild(t *testing.T) {
	t.Run("parent link changed", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-PARENT"
		ledger := claudeStoredLedger(ticket)
		other := "recorded-elsewhere"
		ledger.Sessions[1].ParentID = &other
		workspace := createTestWorkspace(t, root, ticket, ledger)
		before := sessionsByID(t, workspace)
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, true)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		after := sessionsByID(t, workspace)
		if !reflect.DeepEqual(after[claudeFixtureChild], before[claudeFixtureChild]) {
			t.Fatalf("an owner whose parent link changed must be retained: %+v", after[claudeFixtureChild])
		}
		if after[claudeFixtureRoot].Invocations[0].Tokens.Input != 300 {
			t.Fatalf("the root must still refresh: %+v", after[claudeFixtureRoot])
		}
		if res := report.Results[0]; !res.RefreshIncomplete || !strings.Contains(strings.Join(res.Diagnostics, "\n"), "parent linkage changed") {
			t.Fatalf("retention must be reported: %+v", res)
		}
	})
	t.Run("child transcript cleaned up", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-CLEANUP"
		workspace := createTestWorkspace(t, root, ticket, claudeStoredLedger(ticket))
		before := sessionsByID(t, workspace)
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, false)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		after := sessionsByID(t, workspace)
		if !reflect.DeepEqual(after[claudeFixtureChild], before[claudeFixtureChild]) {
			t.Fatalf("a cleaned-up child must keep its stored facts: %+v", after[claudeFixtureChild])
		}
		if after[claudeFixtureRoot].Invocations[0].Tokens.Input != 300 {
			t.Fatalf("the root must still refresh: %+v", after[claudeFixtureRoot])
		}
		diagnostics := strings.Join(report.Results[0].Diagnostics, "\n")
		if !report.Results[0].RefreshIncomplete || !strings.Contains(diagnostics, claudeFixtureChild) {
			t.Fatalf("the unavailable child must be reported: %+v", report.Results[0])
		}
	})
}

func TestBackfillClaudeBudgetExhaustionRetainsStoredFacts(t *testing.T) {
	root := newBackfillProject(t)
	overwriteBackfillLocal(t, root, "agenticlens:\n  state_dir: .publication-state\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: fixture-state\n    max_sessions: 1\n    max_source_bytes: 67108864\n")
	ticket := "TEST-CLAUDE-BUDGET"
	workspace := createTestWorkspace(t, root, ticket, claudeStoredLedger(ticket))
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("budget exhaustion must report incomplete recovery, not fail: %+v %v", report, err)
	}
	after := sessionsByID(t, workspace)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("a tree read past the session budget must not replace stored facts:\nbefore %+v\nafter  %+v", before, after)
	}
	t.Logf("budget diagnostics: %q", report.Results[0].Diagnostics)
	if res := report.Results[0]; !res.RefreshIncomplete || !strings.Contains(strings.Join(res.Diagnostics, "\n"), "budget") {
		t.Fatalf("the budget must be named in an incomplete refresh: %+v", res)
	}
}

func TestBackfillClaudeOwnerConflictingWithAnotherHarnessIsRetained(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-CONFLICT"
	ledger := claudeStoredLedger(ticket)
	// The recorded Codex owner shares the native agent ID of the Claude child.
	ledger.Sessions = append(ledger.Sessions[:1], ledger.Sessions[2:]...)
	ledger.Sessions = append(ledger.Sessions, domain.SessionFact{ID: claudeFixtureChild, Source: domain.SourceCodex, User: "remote", Invocations: []domain.InvocationFact{{Role: "coder", Model: "codex-history"}}})
	workspace := createTestWorkspace(t, root, ticket, ledger)
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	after := sessionsByID(t, workspace)
	if !reflect.DeepEqual(after[claudeFixtureChild], before[claudeFixtureChild]) || after[claudeFixtureChild].Source != domain.SourceCodex {
		t.Fatalf("the Codex owner must keep its history: %+v", after[claudeFixtureChild])
	}
	if after[claudeFixtureRoot].Invocations[0].Tokens.Input != 300 {
		t.Fatalf("the non-conflicting root must refresh: %+v", after[claudeFixtureRoot])
	}
	if res := report.Results[0]; !res.RefreshIncomplete || !strings.Contains(strings.Join(res.Diagnostics, "\n"), "conflicts with another recorded harness") {
		t.Fatalf("the conflict must be reported: %+v", res)
	}
}

func TestBackfillSourceAllRequiresReadableClaudeStoreBeforePublication(t *testing.T) {
	for _, tc := range []struct {
		name  string
		store func(t *testing.T, root string) string
	}{
		{name: "not configured", store: func(*testing.T, string) string { return "" }},
		{name: "missing", store: func(t *testing.T, _ string) string { return filepath.Join(t.TempDir(), "absent-store") }},
		{name: "file", store: func(t *testing.T, _ string) string {
			path := filepath.Join(t.TempDir(), "store-file")
			if err := os.WriteFile(path, []byte("synthetic"), 0o600); err != nil {
				t.Fatal(err)
			}
			return path
		}},
		{name: "unreadable", store: func(t *testing.T, _ string) string {
			if os.Geteuid() == 0 {
				t.Skip("directory permissions do not restrict root; the missing and file cases cover the failure before publication")
			}
			path := t.TempDir()
			writeClaudeFixtureStore(t, path, true)
			if err := os.Chmod(path, 0o000); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
			return path
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newBackfillProject(t)
			ticket := "TEST-CLAUDE-ALL"
			workspace := createTestWorkspace(t, root, ticket, claudeStoredLedger(ticket))
			_, ledgerBefore := readRecoveryLedger(t, workspace)
			summaryBefore, err := os.ReadFile(filepath.Join(workspace, "SUMMARY.md"))
			if err != nil {
				t.Fatal(err)
			}
			codexHome := t.TempDir()
			if err := os.MkdirAll(filepath.Join(codexHome, "sessions"), 0o700); err != nil {
				t.Fatal(err)
			}
			writeBackfillRollout(t, filepath.Join(codexHome, "sessions"), "codex-owner", "", "coder", "fresh-codex", nil, "")
			dbPath := filepath.Join(root, "opencode.db")
			setupMockOpenCodeDB(t, dbPath, "opencode-owner", ticket)
			_, err = NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "all", ClaudeHome: tc.store(t, root), CodexHome: codexHome, DBPath: dbPath, RecomputeDB: true, UpdateSummaries: true})
			t.Logf("source all error: %v", err)
			if err == nil || !(strings.Contains(err.Error(), "--claude-home") || strings.Contains(err.Error(), "telemetry.sources.claude.home")) {
				t.Fatalf("source all must fail naming the Claude Code store choice: %v", err)
			}
			_, ledgerAfter := readRecoveryLedger(t, workspace)
			summaryAfter, err := os.ReadFile(filepath.Join(workspace, "SUMMARY.md"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(ledgerBefore, ledgerAfter) || !bytes.Equal(summaryBefore, summaryAfter) {
				t.Fatal("a failed source all selection changed the ledger or summary")
			}
		})
	}
}

func TestBackfillSourceAllRefreshesEveryHarnessAndBothLeavesClaudeOwners(t *testing.T) {
	setup := func(t *testing.T, ticket string) (string, string, BackfillOptions) {
		root := newBackfillProject(t)
		// No Claude Code store is configured locally; "all" receives it explicitly.
		workspace := createTestWorkspace(t, root, ticket, claudeStoredLedger(ticket))
		codexHome := t.TempDir()
		if err := os.MkdirAll(filepath.Join(codexHome, "sessions"), 0o700); err != nil {
			t.Fatal(err)
		}
		writeBackfillRollout(t, filepath.Join(codexHome, "sessions"), "codex-owner", "", "coder", "fresh-codex", nil, "")
		dbPath := filepath.Join(root, "opencode.db")
		setupMockOpenCodeDB(t, dbPath, "opencode-owner", ticket)
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE session SET model='fresh-opencode'`); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		return root, workspace, BackfillOptions{Project: root, TicketID: ticket, CodexHome: codexHome, DBPath: dbPath, RecomputeDB: true}
	}
	t.Run("all", func(t *testing.T) {
		_, workspace, opts := setup(t, "TEST-CLAUDE-ALL3")
		before := sessionsByID(t, workspace)
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, true)
		opts.Source, opts.ClaudeHome = "all", home
		report, err := NewEngine().Execute(opts)
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("source all failed: %+v %v", report, err)
		}
		after := sessionsByID(t, workspace)
		assertClaudeFixtureRefreshed(t, after)
		if got := after["codex-owner"].Invocations; len(got) != 1 || got[0].Model != "fresh-codex" {
			t.Fatalf("Codex owner not refreshed by all: %+v", after["codex-owner"])
		}
		if tools := after["opencode-owner"].Tools; reflect.DeepEqual(after["opencode-owner"], before["opencode-owner"]) || len(tools) != 1 || tools[0].Name != "bash" || tools[0].Calls != 1 {
			t.Fatalf("OpenCode owner not refreshed by all: %+v", after["opencode-owner"])
		}
		if !reflect.DeepEqual(after["claude-other-machine"], before["claude-other-machine"]) {
			t.Fatalf("other-machine Claude Code owner changed: %+v", after["claude-other-machine"])
		}
	})
	t.Run("both", func(t *testing.T) {
		_, workspace, opts := setup(t, "TEST-CLAUDE-BOTH")
		before := sessionsByID(t, workspace)
		opts.Source = "both"
		report, err := NewEngine().Execute(opts)
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("source both must not need a Claude Code store: %+v %v", report, err)
		}
		after := sessionsByID(t, workspace)
		if got := after["codex-owner"].Invocations; len(got) != 1 || got[0].Model != "fresh-codex" {
			t.Fatalf("Codex owner not refreshed by both: %+v", after["codex-owner"])
		}
		if reflect.DeepEqual(after["opencode-owner"], before["opencode-owner"]) {
			t.Fatalf("OpenCode owner not refreshed by both: %+v", after["opencode-owner"])
		}
		for _, id := range []string{claudeFixtureRoot, claudeFixtureChild, "claude-other-machine"} {
			if !reflect.DeepEqual(after[id], before[id]) {
				t.Fatalf("both changed Claude Code owner %s:\nbefore %+v\nafter  %+v", id, before[id], after[id])
			}
		}
	})
}

func TestBackfillClaudeAddsNewlyProvenChildOfRecordedRoot(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-NEWCHILD"
	ledger := claudeStoredLedger(ticket)
	ledger.Sessions = append(ledger.Sessions[:1], ledger.Sessions[2:]...)
	workspace := createTestWorkspace(t, root, ticket, ledger)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	assertClaudeFixtureRefreshed(t, sessionsByID(t, workspace))
}

// claudeOnlyLedger records only owners the fixture store can refresh.
func claudeOnlyLedger(ticket string) domain.AgenticFactLedger {
	ledger := claudeStoredLedger(ticket)
	ledger.Sessions = ledger.Sessions[:2]
	return ledger
}

func TestBackfillClaudeCompleteRefreshIsReportedComplete(t *testing.T) {
	t.Run("claude", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-COMPLETE"
		workspace := createTestWorkspace(t, root, ticket, claudeOnlyLedger(ticket))
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, true)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		res := report.Results[0]
		if res.RefreshIncomplete || !res.DBFound || len(res.Diagnostics) != 0 || res.CollectionsRetained {
			t.Fatalf("a fully recovered Claude Code ledger must be reported complete: %+v", res)
		}
		assertClaudeFixtureRefreshed(t, sessionsByID(t, workspace))
	})
	t.Run("all with three complete harnesses", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-COMPLETE-ALL"
		ledger := claudeStoredLedger(ticket)
		ledger.Sessions = ledger.Sessions[:4]
		workspace := createTestWorkspace(t, root, ticket, ledger)
		codexHome := t.TempDir()
		if err := os.MkdirAll(filepath.Join(codexHome, "sessions"), 0o700); err != nil {
			t.Fatal(err)
		}
		writeBackfillRollout(t, filepath.Join(codexHome, "sessions"), "codex-owner", "", "coder", "fresh-codex", nil, "")
		dbPath := filepath.Join(root, "opencode.db")
		setupMockOpenCodeDB(t, dbPath, "opencode-owner", ticket)
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, true)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "all", ClaudeHome: home, CodexHome: codexHome, DBPath: dbPath, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		if res := report.Results[0]; res.RefreshIncomplete || !res.DBFound || len(res.Diagnostics) != 0 {
			t.Fatalf("all three harnesses recovered must be reported complete: %+v", res)
		}
		assertClaudeFixtureRefreshed(t, sessionsByID(t, workspace))
	})
}

// A ledger whose source backfill step is pending discharges it only when every
// recorded owner, including Claude Code owners, was recovered.
func TestBackfillPendingSourceStepRequiresClaudeOwners(t *testing.T) {
	pendingLedger := func(ticket string, sessions []domain.SessionFact) domain.AgenticFactLedger {
		version := domain.CurrentSchemaVersion
		return domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: sessions, Migration: &domain.MigrationInfo{SchemaVersion: version, Status: domain.MigrationStatusPartial,
			StepsCompleted: []string{domain.MigrationStepGitCodeFacts}, StepsPending: []string{domain.MigrationStepSourceBackfill}}}
	}
	t.Run("codex refresh leaves Claude owners pending", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-PENDING"
		stored := claudeStoredLedger(ticket).Sessions
		workspace := createTestWorkspace(t, root, ticket, pendingLedger(ticket, []domain.SessionFact{stored[0], stored[1], stored[2]}))
		codexHome := t.TempDir()
		if err := os.MkdirAll(filepath.Join(codexHome, "sessions"), 0o700); err != nil {
			t.Fatal(err)
		}
		writeBackfillRollout(t, filepath.Join(codexHome, "sessions"), "codex-owner", "", "coder", "fresh-codex", nil, "")
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "codex", CodexHome: codexHome, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		res := report.Results[0]
		if res.DBFound || !containsMigrationStep(res.PendingSteps, domain.MigrationStepSourceBackfill) {
			t.Fatalf("unrefreshed Claude Code owners must keep source backfill pending: %+v", res)
		}
		if got := sessionsByID(t, workspace)["codex-owner"].Invocations[0].Model; got != "fresh-codex" {
			t.Fatalf("Codex owner not refreshed: %s", got)
		}
	})
	t.Run("claude refresh of every owner completes the step", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-DISCHARGED"
		createTestWorkspace(t, root, ticket, pendingLedger(ticket, claudeOnlyLedger(ticket).Sessions))
		home := t.TempDir()
		writeClaudeFixtureStore(t, home, true)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
		if err != nil || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		if res := report.Results[0]; !res.DBFound || res.RefreshIncomplete || containsMigrationStep(res.PendingSteps, domain.MigrationStepSourceBackfill) {
			t.Fatalf("complete Claude Code recovery must discharge source backfill: %+v", res)
		}
	})
}

func TestBackfillClaudeSourceWithoutClaudeOwnersPreservesLedger(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-NONE"
	ledger := claudeStoredLedger(ticket)
	ledger.Sessions = ledger.Sessions[2:4]
	workspace := createTestWorkspace(t, root, ticket, ledger)
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	res := report.Results[0]
	if !res.RefreshIncomplete || res.DBFound || !res.CollectionsRetained || !strings.Contains(strings.Join(res.Diagnostics, "\n"), "No exact selected-source owner is recorded") {
		t.Fatalf("a ledger without Claude Code owners must be preserved and reported: %+v", res)
	}
	if after := sessionsByID(t, workspace); !reflect.DeepEqual(after, before) {
		t.Fatalf("store sessions were adopted into a ledger without Claude Code owners:\nbefore %+v\nafter  %+v", before, after)
	}
}

// With a session budget of one, the first recorded root is read completely and
// replaced; the second is retained and named.
func TestBackfillClaudePartialBudgetReplacesOnlyCompletelyReadRoots(t *testing.T) {
	root := newBackfillProject(t)
	overwriteBackfillLocal(t, root, "agenticlens:\n  state_dir: .publication-state\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: fixture-state\n    max_sessions: 1\n    max_source_bytes: 67108864\n")
	ticket := "TEST-CLAUDE-PARTIAL-BUDGET"
	version := domain.CurrentSchemaVersion
	stale := func(id string) domain.SessionFact {
		return domain.SessionFact{ID: id, Source: domain.SourceClaude, User: "local", Invocations: []domain.InvocationFact{{Model: "stale-model", Tokens: domain.InvocationTokens{Input: 1}}}}
	}
	workspace := createTestWorkspace(t, root, ticket, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: []domain.SessionFact{stale("claude-root-a"), stale("claude-root-b")}})
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeTree(t, home, "claude-root-a", false, true)
	writeClaudeTree(t, home, "claude-root-b", false, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	after := sessionsByID(t, workspace)
	if got := after["claude-root-a"].Invocations; len(got) != 1 || got[0].Tokens.Input != 300 {
		t.Fatalf("the root read within the budget must be replaced: %+v", after["claude-root-a"])
	}
	if !reflect.DeepEqual(after["claude-root-b"], before["claude-root-b"]) {
		t.Fatalf("the root beyond the budget must keep its stored facts: %+v", after["claude-root-b"])
	}
	diagnostics := strings.Join(report.Results[0].Diagnostics, "\n")
	if !report.Results[0].RefreshIncomplete || !strings.Contains(diagnostics, "claude-root-b retained: configured source recovery budget reached") {
		t.Fatalf("the unread root must be named with the budget: %s", diagnostics)
	}
}

// A recorded root whose transcript lacks usage keeps its stored facts instead
// of recording zero; its complete child is still replaced.
func TestBackfillClaudeIncompleteRootKeepsStoredFacts(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-INCOMPLETE-ROOT"
	workspace := createTestWorkspace(t, root, ticket, claudeOnlyLedger(ticket))
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeTree(t, home, claudeFixtureRoot, true, false)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	after := sessionsByID(t, workspace)
	if !reflect.DeepEqual(after[claudeFixtureRoot], before[claudeFixtureRoot]) || after[claudeFixtureRoot].Invocations[0].Tokens.Input != 1 {
		t.Fatalf("an incomplete root must keep its stored facts, never zero: %+v", after[claudeFixtureRoot])
	}
	if got := after[claudeFixtureChild].Invocations; len(got) != 1 || got[0].Tokens.Input != 40 {
		t.Fatalf("the complete child must still be replaced: %+v", after[claudeFixtureChild])
	}
	res := report.Results[0]
	diagnostics := strings.Join(res.Diagnostics, "\n")
	if !res.RefreshIncomplete || res.DBFound || !strings.Contains(diagnostics, "is incomplete; previous snapshot retained") || !strings.Contains(diagnostics, "lacks usage") {
		t.Fatalf("the incomplete root must be reported: %+v", res)
	}
}

// A recorded root whose transcript is gone, listed before a recoverable root,
// is retained without stopping recovery of the roots after it.
func TestBackfillClaudeUnavailableRootDoesNotStopLaterRoots(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-ORDER"
	ledger := claudeOnlyLedger(ticket)
	missing := domain.SessionFact{ID: "claude-aaa-missing-root", Source: domain.SourceClaude, User: "remote", Invocations: []domain.InvocationFact{{Model: "remote-model", Tokens: domain.InvocationTokens{Input: 9}}}}
	ledger.Sessions = append([]domain.SessionFact{missing}, ledger.Sessions...)
	workspace := createTestWorkspace(t, root, ticket, ledger)
	before := sessionsByID(t, workspace)
	home := t.TempDir()
	writeClaudeFixtureStore(t, home, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
	if err != nil || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	after := sessionsByID(t, workspace)
	if !reflect.DeepEqual(after[missing.ID], before[missing.ID]) {
		t.Fatalf("the unavailable root must keep its facts: %+v", after[missing.ID])
	}
	assertClaudeFixtureRefreshed(t, after)
	if !report.Results[0].RefreshIncomplete {
		t.Fatalf("an unavailable root keeps the refresh incomplete: %+v", report.Results[0])
	}
}

// A complete read that replaces a recorded owner with fewer invocation rows,
// or with a lower input+output+cache-read total, is reported; each condition
// triggers the notice on its own, and a larger replacement is not reported.
func TestBackfillClaudeSmallerReplacementNotice(t *testing.T) {
	// The fixture root is one row with 300 input, 30 output and 5 cache-read
	// tokens (335); the child is one row with 40+4+1 = 45.
	for _, tc := range []struct {
		name   string
		stored []domain.InvocationFact
		want   string
	}{
		{name: "fewer rows only", stored: []domain.InvocationFact{{Model: "model-a", Tokens: domain.InvocationTokens{Input: 1}}, {Model: "model-b", Tokens: domain.InvocationTokens{Input: 1}}},
			want: "Claude Code owner " + claudeFixtureRoot + " replaced by a complete read with smaller totals: invocations 2 -> 1, tokens 2 -> 335"},
		{name: "lower tokens only", stored: []domain.InvocationFact{{Model: "claude-synthetic-model", Tokens: domain.InvocationTokens{Input: 9000, Output: 900, CacheRead: 100}}},
			want: "Claude Code owner " + claudeFixtureRoot + " replaced by a complete read with smaller totals: invocations 1 -> 1, tokens 10000 -> 335"},
		{name: "larger replacement", stored: []domain.InvocationFact{{Model: "claude-synthetic-model", Tokens: domain.InvocationTokens{Input: 1}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newBackfillProject(t)
			ticket := "TEST-CLAUDE-NOTICE"
			ledger := claudeOnlyLedger(ticket)
			ledger.Sessions[0].Invocations = tc.stored
			workspace := createTestWorkspace(t, root, ticket, ledger)
			home := t.TempDir()
			writeClaudeFixtureStore(t, home, true)
			report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: home, RecomputeDB: true})
			if err != nil || report.Results[0].Error != nil {
				t.Fatalf("backfill failed: %+v %v", report, err)
			}
			assertClaudeFixtureRefreshed(t, sessionsByID(t, workspace))
			notices := report.Results[0].Notices
			if tc.want == "" {
				if len(notices) != 0 {
					t.Fatalf("a larger replacement must not be reported: %q", notices)
				}
				return
			}
			if len(notices) != 1 || notices[0] != tc.want {
				t.Fatalf("want exactly %q, got %q", tc.want, notices)
			}
		})
	}
}
