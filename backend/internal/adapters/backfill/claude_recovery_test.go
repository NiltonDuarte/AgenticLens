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

type claudeLine map[string]any

func writeClaudeLines(t *testing.T, path string, lines []claudeLine) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	for _, line := range lines {
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

func claudeUsage(input, output, cacheRead int) claudeLine {
	return claudeLine{"input_tokens": input, "output_tokens": output, "cache_read_input_tokens": cacheRead}
}

// writeClaudeFixtureStore writes root claudeFixtureRoot with one proven planner
// child. Root: usage (100,10,5)+(200,20,0), one 30s turn. Child: usage
// (40,4,1), one 10s turn.
func writeClaudeFixtureStore(t *testing.T, home string, withChild bool) {
	t.Helper()
	folder := filepath.Join(home, "projects", "-synthetic-project")
	identity := func(line claudeLine, agent string) claudeLine {
		line["sessionId"] = claudeFixtureRoot
		line["cwd"] = "/synthetic/project"
		if agent != "" {
			line["agentId"] = agent
			line["isSidechain"] = true
		}
		return line
	}
	writeClaudeLines(t, filepath.Join(folder, claudeFixtureRoot+".jsonl"), []claudeLine{
		identity(claudeLine{"type": "user", "uuid": "u-1", "timestamp": "2026-01-01T00:00:00Z", "promptSource": "sdk", "message": claudeLine{"role": "user", "content": "synthetic"}}, ""),
		identity(claudeLine{"type": "assistant", "timestamp": "2026-01-01T00:00:01Z", "requestId": "r-1", "message": claudeLine{"id": "m-1", "model": "claude-synthetic-model", "stop_reason": "tool_use", "usage": claudeUsage(100, 10, 5),
			"content": []claudeLine{{"type": "tool_use", "id": "call-1", "name": "Agent", "input": claudeLine{"subagent_type": "planner", "description": "plan"}}}}}, ""),
		identity(claudeLine{"type": "user", "timestamp": "2026-01-01T00:00:20Z", "toolUseResult": claudeLine{"agentId": claudeFixtureChild, "agentType": "planner"}, "message": claudeLine{"role": "user", "content": []claudeLine{{"type": "tool_result", "tool_use_id": "call-1"}}}}, ""),
		identity(claudeLine{"type": "assistant", "timestamp": "2026-01-01T00:00:30Z", "requestId": "r-2", "message": claudeLine{"id": "m-2", "model": "claude-synthetic-model", "stop_reason": "end_turn", "usage": claudeUsage(200, 20, 0), "content": []claudeLine{}}}, ""),
	})
	if !withChild {
		return
	}
	subagents := filepath.Join(folder, claudeFixtureRoot, "subagents")
	writeClaudeLines(t, filepath.Join(subagents, "agent-"+claudeFixtureChild+".jsonl"), []claudeLine{
		identity(claudeLine{"type": "user", "uuid": "c-1", "timestamp": "2026-01-01T00:00:02Z", "message": claudeLine{"role": "user", "content": "synthetic"}}, claudeFixtureChild),
		identity(claudeLine{"type": "assistant", "timestamp": "2026-01-01T00:00:12Z", "requestId": "r-c1", "message": claudeLine{"id": "m-c1", "model": "claude-synthetic-model", "stop_reason": "end_turn", "usage": claudeUsage(40, 4, 1), "content": []claudeLine{}}}, claudeFixtureChild),
	})
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+claudeFixtureChild+".meta.json"), []byte(`{"agentType":"planner","toolUseId":"call-1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
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
		if !result.RefreshIncomplete || !strings.Contains(strings.Join(result.Diagnostics, "\n"), "claude-other-machine") {
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
