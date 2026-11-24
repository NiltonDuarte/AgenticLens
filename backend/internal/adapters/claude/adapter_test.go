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

type record map[string]any

func prompt(ts, source string) record {
	r := record{"type": "user", "timestamp": ts, "uuid": "u-" + ts, "cwd": "/synthetic/project", "message": record{"role": "user", "content": "synthetic prompt"}}
	if source != "" {
		r["promptSource"] = source
	}
	return r
}

func assistant(ts, requestID, model, stop string, usage record, blocks ...record) record {
	if len(blocks) == 0 {
		blocks = []record{{"type": "text", "text": "synthetic"}}
	}
	message := record{"id": "msg-" + requestID, "model": model, "role": "assistant", "stop_reason": stop, "content": blocks}
	if usage != nil {
		message["usage"] = usage
	}
	return record{"type": "assistant", "timestamp": ts, "requestId": requestID, "message": message}
}

func usage(input, output, cacheRead int64) record {
	return record{"input_tokens": input, "output_tokens": output, "cache_read_input_tokens": cacheRead, "cache_creation_input_tokens": int64(0)}
}

func toolUseBlock(id, name string, input record) record {
	return record{"type": "tool_use", "id": id, "name": name, "input": input}
}

func resultRecord(ts, id string, isError bool, details record) record {
	r := record{"type": "user", "timestamp": ts, "message": record{"role": "user", "content": []record{{"type": "tool_result", "tool_use_id": id, "is_error": isError}}}}
	if details != nil {
		r["toolUseResult"] = details
	}
	return r
}

func withIdentity(records []record, sessionID, agentID string) []record {
	out := make([]record, 0, len(records))
	for _, r := range records {
		copied := record{}
		for k, v := range r {
			copied[k] = v
		}
		if _, set := copied["sessionId"]; !set {
			copied["sessionId"] = sessionID
		}
		if agentID != "" {
			if _, set := copied["agentId"]; !set {
				copied["agentId"] = agentID
			}
			copied["isSidechain"] = true
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
		meta, _ := json.Marshal(record{"agentType": agentType, "toolUseId": toolUseID})
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
	fork["forkedFrom"] = record{"sessionId": "source-session", "messageUuid": "m1"}
	store.root(syntheticFolder, syntheticRoot, []record{
		fork,
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:02Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), record{"type": "thinking"}),
		assistant("2026-11-24T10:00:03Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), record{"type": "text"}),
		assistant("2026-11-24T10:00:04Z", "req-1", "model-a", "end_turn", usage(10, 20, 30), toolUseBlock("t1", "Bash", record{})),
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
		"partial":     assistant("2026-11-24T10:00:02Z", "req-3", "model-a", "end_turn", record{"input_tokens": 1, "output_tokens": 1}),
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
			toolUseBlock("call-child", "Agent", record{"subagent_type": "planner", "description": "plan"})),
		resultRecord("2026-11-24T10:00:30Z", "call-child", false, record{"agentId": "achild", "status": "completed"}),
		assistant("2026-11-24T10:00:31Z", "req-2", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-missing", "Agent", record{"subagent_type": "coder", "description": "code"})),
		resultRecord("2026-11-24T10:00:32Z", "call-missing", false, record{"agentId": "agone", "status": "async_launched"}),
		assistant("2026-11-24T10:00:33Z", "req-3", "model-a", "end_turn", usage(1, 1, 1)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "planner", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		assistant("2026-11-24T10:00:03Z", "req-c1", "model-a", "tool_use", usage(2, 2, 2),
			toolUseBlock("call-nested", "Agent", record{"subagent_type": "reviewer", "description": "review"})),
		resultRecord("2026-11-24T10:00:20Z", "call-nested", false, record{"agentId": "anested"}),
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
	sidechain["isSidechain"] = true
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
			toolUseBlock("s-ok", "Skill", record{"skill": "review"}), toolUseBlock("s-false", "Skill", record{"skill": "review"})),
		resultRecord("2026-11-24T10:00:02Z", "s-ok", false, record{"success": true, "commandName": "review"}),
		resultRecord("2026-11-24T10:00:03Z", "s-false", false, record{"success": false}),
		assistant("2026-11-24T10:00:04Z", "req-2", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("s-err", "Skill", record{"skill": "deploy"})),
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
			toolUseBlock("call-child", "Agent", record{"subagent_type": "coder"})),
		{"type": "system", "subtype": "compact_boundary", "timestamp": "2026-11-24T10:00:02Z"},
		assistant("2026-11-24T10:00:03Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	resume := prompt("2026-11-24T10:10:00Z", "")
	resume["isMeta"] = true
	resume["origin"] = record{"kind": "coordinator"}
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
	withNonce["message"] = record{"role": "user", "content": "run " + nonce}
	store.root("-worktree-a", syntheticRoot, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	store.subagent("-worktree-a", syntheticRoot, "achild", "call-x", "coder", []record{withNonce})
	store.root("-worktree-b", other, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	writeJSONL(t, filepath.Join(store.home, "projects", "-worktree-a", syntheticRoot, "tool-results", "ignored.jsonl"), []record{{"type": "user"}})

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
