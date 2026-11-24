package claude

// All fixtures in this file are synthetic. They reproduce the Claude Code
// 2.1.331 record shapes (keys and identities) without any transcript content,
// and they never read ~/.claude or a live ticket workspace.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

const (
	syntheticRoot   = "00000000-0000-4000-8000-000000000001"
	syntheticFolder = "-synthetic-project"
)

func prompt(ts, source string) record {
	return record{Type: "user", Timestamp: ts, UUID: "u-" + ts, CWD: "/synthetic/project", PromptSource: source,
		Message: &message{Role: "user", Content: textContent("synthetic prompt")}}
}

// assistant builds one response record; nil usage omits the usage object.
func assistant(ts, requestID, model, stop string, usage json.RawMessage, blocks ...block) record {
	if len(blocks) == 0 {
		blocks = []block{{Type: "text", Text: "synthetic"}}
	}
	return record{Type: "assistant", Timestamp: ts, RequestID: requestID,
		Message: &message{ID: "msg-" + requestID, Model: model, Role: "assistant", StopReason: stop, Usage: usage, Content: encode(blocks)}}
}

func usage(input, output, cacheRead int64) json.RawMessage {
	return usageFields{Input: n64(input), Output: n64(output), CacheRead: n64(cacheRead), CacheCreation: n64(0)}.encode()
}

func toolUseBlock(id, name string, input toolInput) block {
	return block{Type: "tool_use", ID: id, Name: name, Input: &input}
}

func resultRecord(ts, id string, isError bool, details *resultDetails) record {
	return record{Type: "user", Timestamp: ts, ToolUseResult: details,
		Message: &message{Role: "user", Content: encode([]block{{Type: "tool_result", ToolUseID: id, IsError: boolPtr(isError)}})}}
}

func withIdentity(records []record, sessionID, agentID string) []record {
	out := make([]record, 0, len(records))
	for _, r := range records {
		copied := r
		if copied.SessionID == "" {
			copied.SessionID = sessionID
		}
		if agentID != "" {
			if copied.AgentID == "" && !copied.OmitAgentID {
				copied.AgentID = agentID
			}
			copied.IsSidechain = true
		}
		out = append(out, copied)
	}
	return out
}

func writeJSONL(t *testing.T, path string, records []record) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var builder strings.Builder
	for _, r := range records {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		builder.Write(data)
		builder.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(builder.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

type syntheticStore struct {
	t    *testing.T
	home string
}

func newStore(t *testing.T) syntheticStore {
	return syntheticStore{t: t, home: t.TempDir()}
}

func (s syntheticStore) root(folder, id string, records []record) {
	writeJSONL(s.t, filepath.Join(s.home, "projects", folder, id+".jsonl"), withIdentity(records, id, ""))
}

func (s syntheticStore) subagent(folder, root, agent, toolUseID, agentType string, records []record) {
	dir := filepath.Join(s.home, "projects", folder, root, "subagents")
	writeJSONL(s.t, filepath.Join(dir, "agent-"+agent+".jsonl"), withIdentity(records, root, agent))
	if toolUseID != "" {
		meta, _ := json.Marshal(metaFixture{AgentType: agentType, ToolUseID: toolUseID})
		if err := os.WriteFile(filepath.Join(dir, "agent-"+agent+".meta.json"), meta, 0o644); err != nil {
			s.t.Fatal(err)
		}
	}
}

func extract(t *testing.T, home string, options DiscoveryOptions, cutoff time.Time) (TreeFiles, Extraction) {
	t.Helper()
	options.ClaudeHome = home
	files, err := DiscoverTree(options)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	tree, err := ReadTree(context.Background(), files, nil)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	policy := domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
	extraction, err := tree.Prove().Facts(cutoff, &policy)
	if err != nil {
		t.Fatalf("facts: %v", err)
	}
	return files, extraction
}

func sessionByID(t *testing.T, sessions []domain.SessionFact, id string) domain.SessionFact {
	t.Helper()
	for _, s := range sessions {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("session %s not extracted; have %d sessions", id, len(sessions))
	return domain.SessionFact{}
}

func TestSyntheticRepeatedContentBlocksCountOnceAndForkCopiesAreExcluded(t *testing.T) {
	store := newStore(t)
	fork := assistant("2026-11-24T10:00:01Z", "req-inherited", "model-a", "end_turn", usage(1000, 1000, 1000))
	fork.ForkedFrom = &forkSource{SessionID: "source-session", MessageUUID: "m1"}
	store.root(syntheticFolder, syntheticRoot, []record{
		fork,
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:02Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), block{Type: "thinking"}),
		assistant("2026-11-24T10:00:03Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), block{Type: "text"}),
		assistant("2026-11-24T10:00:04Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), toolUseBlock("t1", "Bash", toolInput{})),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Date(2026, 11, 24, 11, 0, 0, 0, time.UTC))
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Invocations) != 1 {
		t.Fatalf("want one served-model row, got %d", len(root.Invocations))
	}
	tokens := root.Invocations[0].Tokens
	if tokens.Input != 10 || tokens.Output != 20 || tokens.CacheRead != 30 || tokens.PeakContext != 40 {
		t.Fatalf("usage must count each response once and exclude fork copies: %+v", tokens)
	}
	if root.Invocations[0].CostEvidence.State != domain.CostUnavailable {
		t.Fatalf("cost must be unavailable: %+v", root.Invocations[0].CostEvidence)
	}
	if root.ParentID != nil {
		t.Fatalf("a fork must not create a parent link")
	}
}

func TestSyntheticMissingOrConflictingUsageMakesOwnerIncomplete(t *testing.T) {
	for name, second := range map[string]record{
		"missing":     assistant("2026-11-24T10:00:02Z", "req-2", "model-a", "end_turn", nil),
		"conflicting": assistant("2026-11-24T10:00:02Z", "req-1", "model-a", "end_turn", usage(11, 20, 30)),
		"partial":     assistant("2026-11-24T10:00:02Z", "req-3", "model-a", "end_turn", usageFields{Input: n64(1), Output: n64(1)}.encode()),
	} {
		t.Run(name, func(t *testing.T) {
			store := newStore(t)
			store.root(syntheticFolder, syntheticRoot, []record{
				prompt("2026-11-24T10:00:00Z", "sdk"),
				assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(10, 20, 30)),
				second,
			})
			_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Now())
			if len(extraction.RootIncomplete) == 0 || len(extraction.Sessions) != 0 {
				t.Fatalf("owner must be incomplete and omitted, got reasons=%v sessions=%d", extraction.RootIncomplete, len(extraction.Sessions))
			}
		})
	}
}

func TestSyntheticOneInvocationRowPerServedModel(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 2, 3)),
		assistant("2026-11-24T10:00:02Z", "req-2", "model-b-20261124", "end_turn", usage(4, 5, 6)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Now())
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Invocations) != 2 || root.Invocations[0].Model != "model-a" || root.Invocations[1].Model != "model-b-20261124" {
		t.Fatalf("rows must be keyed by verbatim served model: %+v", root.Invocations)
	}
}

func TestSyntheticDurationCompletedIdleAndOpenTurns(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:02:00Z", "req-1", "model-a", "end_turn", usage(1, 1, 1)),
		// 600 idle seconds, then a task notification starts an open turn.
		prompt("2026-11-24T10:12:00Z", "system"),
		assistant("2026-11-24T10:12:10Z", "req-2", "model-a", "tool_use", usage(1, 1, 1)),
	})
	cutoff := time.Date(2026, 11, 24, 10, 12, 30, 0, time.UTC)
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, cutoff)
	row := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0]
	if row.ActiveDurationSeconds != 150 || row.ProvisionalSeconds != 30 {
		t.Fatalf("want 150s with 30s provisional, got %d/%d", row.ActiveDurationSeconds, row.ProvisionalSeconds)
	}
	if row.UserIterations != 1 || row.Turns != 2 {
		t.Fatalf("task notification is a turn but not steering: iterations=%d turns=%d", row.UserIterations, row.Turns)
	}
	if row.CompletedAt != nil {
		t.Fatalf("a provisional row has no completion time")
	}
}

func TestSyntheticProvenNestedUnprovenAndExplicitChildren(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", toolInput{SubagentType: "planner", Description: "plan"})),
		resultRecord("2026-11-24T10:00:30Z", "call-child", false, &resultDetails{AgentID: "achild", Status: "completed"}),
		assistant("2026-11-24T10:00:31Z", "req-2", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-missing", "Agent", toolInput{SubagentType: "coder", Description: "code"})),
		resultRecord("2026-11-24T10:00:32Z", "call-missing", false, &resultDetails{AgentID: "agone", Status: "async_launched"}),
		assistant("2026-11-24T10:00:33Z", "req-3", "model-a", "end_turn", usage(1, 1, 1)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "planner", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		assistant("2026-11-24T10:00:03Z", "req-c1", "model-a", "tool_use", usage(2, 2, 2),
			toolUseBlock("call-nested", "Agent", toolInput{SubagentType: "reviewer", Description: "review"})),
		resultRecord("2026-11-24T10:00:20Z", "call-nested", false, &resultDetails{AgentID: "anested"}),
		assistant("2026-11-24T10:00:21Z", "req-c2", "model-a", "end_turn", usage(2, 2, 2)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "anested", "call-nested", "reviewer", []record{
		prompt("2026-11-24T10:00:04Z", ""),
		assistant("2026-11-24T10:00:05Z", "req-n1", "model-a", "end_turn", usage(3, 3, 3)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "aunproven", "call-unknown", "coder", []record{
		prompt("2026-11-24T10:00:06Z", ""),
		assistant("2026-11-24T10:00:07Z", "req-u1", "model-a", "end_turn", usage(4, 4, 4)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "anometa", "", "", []record{prompt("2026-11-24T10:00:08Z", "")})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: "anested"}, time.Now())
	if len(extraction.Sessions) != 3 {
		t.Fatalf("want root, child and nested child, got %d", len(extraction.Sessions))
	}
	child := sessionByID(t, extraction.Sessions, "achild")
	nested := sessionByID(t, extraction.Sessions, "anested")
	if child.ParentID == nil || *child.ParentID != syntheticRoot || nested.ParentID == nil || *nested.ParentID != "achild" {
		t.Fatalf("parentage must follow the transcript holding the Agent tool_use")
	}
	if child.Invocations[0].Role != "planner" || nested.Invocations[0].Role != "reviewer" {
		t.Fatalf("role must be the native agentType")
	}
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Delegations) != 2 || root.Delegations[0].ChildSession != "achild" || root.Delegations[1].ChildSession != "agone" {
		t.Fatalf("delegations must reference proven or explicit agent IDs: %+v", root.Delegations)
	}
	joined := strings.Join(extraction.Diagnostics, "\n")
	for _, id := range []string{"aunproven", "anometa", "agone"} {
		if !strings.Contains(joined, id) {
			t.Fatalf("missing diagnostic for %s in %q", id, joined)
		}
	}
}

func TestSyntheticSidechainInRootIsUnsupportedLayout(t *testing.T) {
	store := newStore(t)
	sidechain := assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1))
	sidechain.IsSidechain = true
	store.root(syntheticFolder, syntheticRoot, []record{prompt("2026-11-24T10:00:00Z", "sdk"), sidechain})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Now())
	if len(extraction.RootIncomplete) == 0 || !strings.Contains(extraction.RootIncomplete[0], "unsupported layout") {
		t.Fatalf("sidechain records inside a root must be reported unsupported: %v", extraction.RootIncomplete)
	}
}

func TestSyntheticSkillActivationsAndFailures(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("s-ok", "Skill", toolInput{Skill: "review"}), toolUseBlock("s-false", "Skill", toolInput{Skill: "review"})),
		resultRecord("2026-11-24T10:00:02Z", "s-ok", false, &resultDetails{Success: boolPtr(true), CommandName: "review"}),
		resultRecord("2026-11-24T10:00:03Z", "s-false", false, &resultDetails{Success: boolPtr(false)}),
		assistant("2026-11-24T10:00:04Z", "req-2", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("s-err", "Skill", toolInput{Skill: "deploy"})),
		resultRecord("2026-11-24T10:00:05Z", "s-err", true, nil),
		assistant("2026-11-24T10:00:06Z", "req-3", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Now())
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Skills) != 1 || root.Skills[0].Name != "review" {
		t.Fatalf("only the successful Skill call is an activation: %+v", root.Skills)
	}
	if len(root.Tools) != 1 || root.Tools[0].Name != "Skill" || root.Tools[0].Calls != 3 || root.Tools[0].Errors != 2 {
		t.Fatalf("failed Skill calls count as tool errors: %+v", root.Tools)
	}
}

func TestSyntheticCompactionCountsAndResumedSubagentTurn(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", toolInput{SubagentType: "coder"})),
		{Type: "system", Subtype: "compact_boundary", Timestamp: "2026-11-24T10:00:02Z"},
		assistant("2026-11-24T10:00:03Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	resume := prompt("2026-11-24T10:10:00Z", "")
	resume.IsMeta = true
	resume.Origin = &origin{Kind: "coordinator"}
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "coder", []record{
		prompt("2026-11-24T10:00:01Z", ""),
		assistant("2026-11-24T10:00:11Z", "req-c1", "model-a", "end_turn", usage(1, 1, 1)),
		resume,
		assistant("2026-11-24T10:10:20Z", "req-c2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC))
	if got := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0].Compactions; got != 1 {
		t.Fatalf("want one compaction, got %d", got)
	}
	child := sessionByID(t, extraction.Sessions, "achild").Invocations[0]
	// A coordinator resume of an idle subagent starts a measured turn; the idle
	// gap between 10:00:11 and 10:10:00 is excluded.
	if child.Turns != 2 || child.ActiveDurationSeconds != 30 || child.ProvisionalSeconds != 0 {
		t.Fatalf("want 2 turns and 30s measured, got turns=%d active=%d provisional=%d", child.Turns, child.ActiveDurationSeconds, child.ProvisionalSeconds)
	}
}

func TestSyntheticDiscoveryNonceAmbiguityAndBudget(t *testing.T) {
	store := newStore(t)
	other := "00000000-0000-4000-8000-000000000002"
	nonce := "synthetic-nonce-7f3a"
	withNonce := prompt("2026-11-24T10:00:00Z", "sdk")
	withNonce.Message = &message{Role: "user", Content: textContent("run " + nonce)}
	store.root("-worktree-a", syntheticRoot, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	store.subagent("-worktree-a", syntheticRoot, "achild", "call-x", "coder", []record{withNonce})
	store.root("-worktree-b", other, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	writeJSONL(t, filepath.Join(store.home, "projects", "-worktree-a", syntheticRoot, "tool-results", "ignored.jsonl"), []record{{Type: "user"}})

	files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce})
	if err != nil || files.RootID != syntheticRoot {
		t.Fatalf("nonce in a subagent transcript must select its root: %v %+v", err, files)
	}
	store.root("-worktree-b", other, []record{withNonce})
	if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce}); err == nil || !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("nonce in two roots must be ambiguous: %v", err)
	}
	store.root("-worktree-c", other, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: other}); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("one session ID in two folders must be ambiguous: %v", err)
	}
	if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: "absent"}); !errors.Is(err, ErrOwnerUnavailable) {
		t.Fatalf("absent ID must report unavailable: %v", err)
	}
	budget := &DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 20}
	files, err = DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot, Budget: budget})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTree(context.Background(), files, budget); !errors.Is(err, ErrDiscoveryBudget) {
		t.Fatalf("a tree beyond the session budget must stop with the typed budget error: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReadTree(cancelled, files, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation must stop reads: %v", err)
	}
}
