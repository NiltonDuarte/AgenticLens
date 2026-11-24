package claude

// All fixtures in this file are synthetic. They reuse the record builders of
// adapter_test.go, carry no transcript content and never read ~/.claude or a
// live ticket workspace. Expected values are computed by hand from the fixture
// timestamps and usage numbers written next to each assertion.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

var laterCutoff = time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)

func rootPath(store syntheticStore) string {
	return filepath.Join(store.home, "projects", syntheticFolder, syntheticRoot+".jsonl")
}

func appendRawLine(t *testing.T, path, line string) {
	t.Helper()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(line + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func withoutKey(r record, key string) record {
	copied := record{}
	for k, v := range r {
		if k != key {
			copied[k] = v
		}
	}
	return copied
}

func withMessageField(r record, key string, value any) record {
	copied := withoutKey(r, "message")
	message := record{}
	for k, v := range r["message"].(record) {
		message[k] = v
	}
	if value == nil {
		delete(message, key)
	} else {
		message[key] = value
	}
	copied["message"] = message
	return copied
}

// Each malformed or partial record must reach its own incomplete branch: the
// expected reason is asserted, so a fixture that failed earlier for another
// reason cannot satisfy the test.
func TestSyntheticMalformedRecordsMakeRootIncompleteThroughTheirOwnBranch(t *testing.T) {
	base := assistant("2026-11-24T10:00:02Z", "req-2", "model-a", "end_turn", usage(1, 1, 1))
	for _, tc := range []struct {
		name   string
		record record
		raw    string
		reason string
	}{
		{name: "malformed line", raw: `{"type":"assistant",`, reason: "malformed JSON record"},
		{name: "untyped record", raw: `{"sessionId":"` + syntheticRoot + `","timestamp":"2026-11-24T10:00:02Z"}`, reason: "record has no type"},
		{name: "foreign sessionId", record: func() record { r := base; r = withoutKey(r, "sessionId"); r["sessionId"] = "foreign-session"; return r }(), reason: "different sessionId"},
		{name: "foreign sessionId on ignored record type", record: record{"type": "file-history-snapshot", "sessionId": "foreign-session"}, reason: "different sessionId"},
		{name: "assistant without timestamp", record: withoutKey(base, "timestamp"), reason: "assistant record has no valid timestamp"},
		{name: "user without timestamp", record: withoutKey(prompt("2026-11-24T10:00:02Z", "sdk"), "timestamp"), reason: "user record has no valid timestamp"},
		{name: "assistant without message", record: withoutKey(base, "message"), reason: "assistant record has no readable message"},
		{name: "user without message", record: withoutKey(prompt("2026-11-24T10:00:02Z", "sdk"), "message"), reason: "user record has no readable message"},
		{name: "user with unreadable content", record: withMessageField(prompt("2026-11-24T10:00:02Z", "sdk"), "content", 7), reason: "user record has unreadable content"},
		{name: "response without any identity", record: withMessageField(withoutKey(base, "requestId"), "id", nil), reason: "neither requestId nor message.id"},
		{name: "unreadable usage", record: withMessageField(base, "usage", "not-an-object"), reason: "unreadable usage"},
		{name: "null usage", record: withMessageField(base, "usage", json.RawMessage("null")), reason: "lacks usage"},
		{name: "missing cache-read category", record: withMessageField(base, "usage", record{"input_tokens": 1, "output_tokens": 1}), reason: "lacks input, output or cache-read usage"},
		{name: "missing output category", record: withMessageField(base, "usage", record{"input_tokens": 1, "cache_read_input_tokens": 1}), reason: "lacks input, output or cache-read usage"},
		{name: "negative usage", record: withMessageField(base, "usage", usage(-1, 1, 1)), reason: "negative usage"},
		{name: "negative cache creation", record: withMessageField(base, "usage", record{"input_tokens": 1, "output_tokens": 1, "cache_read_input_tokens": 1, "cache_creation_input_tokens": -5}), reason: "negative usage"},
		{name: "served model missing", record: withMessageField(base, "model", ""), reason: "has no served model"},
		{name: "same response from another model", record: assistant("2026-11-24T10:00:02Z", "req-1", "model-b", "end_turn", usage(10, 20, 30)), reason: "conflicting usage or model"},
		{name: "tool_use without id", record: assistant("2026-11-24T10:00:02Z", "req-2", "model-a", "tool_use", usage(1, 1, 1), record{"type": "tool_use", "name": "Bash", "input": record{}}), reason: "tool_use requires id and name"},
		{name: "tool_result without tool_use_id", record: record{"type": "user", "timestamp": "2026-11-24T10:00:02Z", "message": record{"role": "user", "content": []record{{"type": "tool_result"}}}}, reason: "tool_result requires tool_use_id"},
		{name: "successful Skill without a name", record: assistant("2026-11-24T10:00:02Z", "req-2", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("s-blank", "Skill", record{"args": "x"})), reason: "has no skill name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newStore(t)
			records := []record{
				prompt("2026-11-24T10:00:00Z", "sdk"),
				assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(10, 20, 30)),
			}
			if tc.record != nil {
				records = append(records, tc.record)
			}
			if tc.name == "successful Skill without a name" {
				records = append(records, resultRecord("2026-11-24T10:00:03Z", "s-blank", false, record{"success": true}))
			}
			store.root(syntheticFolder, syntheticRoot, records)
			if tc.raw != "" {
				appendRawLine(t, rootPath(store), tc.raw)
			}
			_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
			if len(extraction.Sessions) != 0 || len(extraction.CompleteSessionIDs) != 0 {
				t.Fatalf("incomplete root must not be recorded, got %d sessions %v", len(extraction.Sessions), extraction.CompleteSessionIDs)
			}
			if !strings.Contains(strings.Join(extraction.RootIncomplete, "; "), tc.reason) {
				t.Fatalf("root incomplete for the wrong reason: want %q in %v", tc.reason, extraction.RootIncomplete)
			}
		})
	}
}

// Non-conversation record types are ignored, and a cost-state estimate never
// becomes observed spend anywhere in the extraction.
func TestSyntheticCostStateAndOtherRecordTypesAreIgnored(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		{"type": "cost-state", "totalCostUSD": 98.7654, "modelUsage": record{"model-a": record{"costUSD": 98.7654}}},
		{"type": "file-history-snapshot", "messageId": "m-1"},
		{"type": "queue-operation", "operation": "enqueue"},
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:04Z", "req-1", "model-a", "end_turn", usage(3, 4, 5)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	if len(extraction.RootIncomplete) != 0 {
		t.Fatalf("ignored record types made the root incomplete: %v", extraction.RootIncomplete)
	}
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Invocations) != 1 {
		t.Fatalf("want one row, got %+v", root.Invocations)
	}
	row := root.Invocations[0]
	if row.CostUSD != 0 || row.CostEvidence.State != domain.CostUnavailable {
		t.Fatalf("cost must stay unavailable: %v %+v", row.CostUSD, row.CostEvidence)
	}
	if row.Tokens.Input != 3 || row.Tokens.Output != 4 || row.Tokens.CacheRead != 5 || row.ActiveDurationSeconds != 4 || row.Turns != 1 {
		t.Fatalf("ignored records changed facts: %+v", row)
	}
	encoded, err := json.Marshal(struct {
		Sessions []domain.SessionFact
		Events   []domain.EventFact
	}{extraction.Sessions, extraction.Events})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "98.76") {
		t.Fatalf("cost-state estimate leaked into the extraction: %s", encoded)
	}
}

// Client-side <synthetic> notices contribute no usage and no invocation row,
// need no usage object, and still end the turn they belong to.
func TestSyntheticModelNoticeAddsNoUsageOrRowButEndsTurn(t *testing.T) {
	store := newStore(t)
	notice := record{"type": "assistant", "timestamp": "2026-11-24T10:00:20Z", "isApiErrorMessage": true,
		"message": record{"id": "synthetic-notice-1", "model": "<synthetic>", "role": "assistant", "stop_reason": "stop_sequence", "content": []record{{"type": "text", "text": "synthetic"}}}}
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:05Z", "req-1", "model-a", "tool_use", usage(5, 6, 7), toolUseBlock("t-1", "Bash", record{})),
		resultRecord("2026-11-24T10:00:08Z", "t-1", false, nil),
		notice,
		// 280 idle seconds after the notice ended the first turn.
		prompt("2026-11-24T10:05:00Z", "sdk"),
		assistant("2026-11-24T10:05:10Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	if len(extraction.RootIncomplete) != 0 {
		t.Fatalf("a usage-free synthetic notice must not make the owner incomplete: %v", extraction.RootIncomplete)
	}
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	if len(root.Invocations) != 1 || root.Invocations[0].Model != "model-a" {
		t.Fatalf("synthetic notices must not create a row: %+v", root.Invocations)
	}
	row := root.Invocations[0]
	// Usage: (5,6,7) + (1,1,1). Duration: 20s (prompt to notice) + 10s.
	if row.Tokens.Input != 6 || row.Tokens.Output != 7 || row.Tokens.CacheRead != 8 {
		t.Fatalf("synthetic notice changed usage: %+v", row.Tokens)
	}
	if row.Turns != 2 || row.ActiveDurationSeconds != 30 || row.ProvisionalSeconds != 0 || row.UserIterations != 2 {
		t.Fatalf("want 2 turns, 30s measured and 2 iterations; got turns=%d active=%d provisional=%d iterations=%d", row.Turns, row.ActiveDurationSeconds, row.ProvisionalSeconds, row.UserIterations)
	}
	if row.CompletedAt == nil || *row.CompletedAt != "2026-11-24T10:05:10Z" {
		t.Fatalf("completion must be the last terminal record: %v", row.CompletedAt)
	}
}

// A synthetic notice that is the only assistant record still ends the turn; no
// row is named <synthetic>.
func TestSyntheticModelNoticeOnlyTurnCreatesNoSyntheticRow(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:03Z", "req-1", "model-a", "end_turn", usage(2, 2, 2)),
		prompt("2026-11-24T10:01:00Z", "sdk"),
		{"type": "assistant", "timestamp": "2026-11-24T10:01:07Z", "message": record{"id": "n-1", "model": "<synthetic>", "stop_reason": "stop_sequence", "content": []record{}}},
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	for _, row := range root.Invocations {
		if row.Model == "<synthetic>" {
			t.Fatalf("synthetic model became a row: %+v", root.Invocations)
		}
	}
	// 3s + 7s, both closed; the notice-only turn is attributed to model-a.
	if len(root.Invocations) != 1 || root.Invocations[0].ActiveDurationSeconds != 10 || root.Invocations[0].ProvisionalSeconds != 0 {
		t.Fatalf("want one model-a row with 10 measured seconds: %+v", root.Invocations)
	}
	if root.Invocations[0].Tokens.Input != 2 {
		t.Fatalf("usage changed: %+v", root.Invocations[0].Tokens)
	}
}

// Root meta records, even coordinator-originated ones, start nothing.
func TestSyntheticRootMetaPromptStartsNoTurnAndNoSteering(t *testing.T) {
	store := newStore(t)
	meta := prompt("2026-11-24T10:01:00Z", "")
	meta["isMeta"] = true
	meta["origin"] = record{"kind": "coordinator"}
	summary := prompt("2026-11-24T10:02:00Z", "")
	summary["isCompactSummary"] = true
	transcriptOnly := prompt("2026-11-24T10:03:00Z", "")
	transcriptOnly["isVisibleInTranscriptOnly"] = true
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:10Z", "req-1", "model-a", "end_turn", usage(1, 1, 1)),
		meta,
		assistant("2026-11-24T10:01:05Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
		summary,
		transcriptOnly,
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	row := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0]
	if row.Turns != 1 || row.ActiveDurationSeconds != 10 || row.ProvisionalSeconds != 0 || row.UserIterations != 1 {
		t.Fatalf("meta, summary and transcript-only prompts must start nothing: turns=%d active=%d provisional=%d iterations=%d", row.Turns, row.ActiveDurationSeconds, row.ProvisionalSeconds, row.UserIterations)
	}
	if row.Tokens.Input != 2 {
		t.Fatalf("usage after a meta prompt is still counted once: %+v", row.Tokens)
	}
}

// A coordinator resume that arrives while the subagent's turn is still open
// extends that turn instead of starting another, and never counts as steering.
func TestSyntheticResumedSubagentWithOpenTurnDoesNotStartAnotherTurn(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", record{"subagent_type": "coder"})),
		resultRecord("2026-11-24T10:00:30Z", "call-child", false, record{"agentId": "achild"}),
		assistant("2026-11-24T10:00:31Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	resume := prompt("2026-11-24T10:00:10Z", "")
	resume["isMeta"] = true
	resume["origin"] = record{"kind": "coordinator"}
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "coder", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		assistant("2026-11-24T10:00:05Z", "req-c1", "model-a", "tool_use", usage(1, 1, 1)),
		resume,
		assistant("2026-11-24T10:00:22Z", "req-c2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	child := sessionByID(t, extraction.Sessions, "achild").Invocations[0]
	// One turn from 10:00:02 to 10:00:22.
	if child.Turns != 1 || child.ActiveDurationSeconds != 20 || child.UserIterations != 0 {
		t.Fatalf("want one 20s turn and no steering, got turns=%d active=%d iterations=%d", child.Turns, child.ActiveDurationSeconds, child.UserIterations)
	}
	if root := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0]; root.UserIterations != 1 {
		t.Fatalf("subagent resume changed root steering: %d", root.UserIterations)
	}
}

// A resumed subagent that never reaches end_turn is measured provisionally from
// the resume record, not from its first turn.
func TestSyntheticResumedSubagentOpenTurnIsProvisional(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", record{"subagent_type": "coder"})),
	})
	resume := prompt("2026-11-24T10:20:00Z", "")
	resume["isMeta"] = true
	resume["origin"] = record{"kind": "coordinator"}
	resume["uuid"] = "resume-uuid"
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "coder", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		assistant("2026-11-24T10:00:12Z", "req-c1", "model-a", "end_turn", usage(1, 1, 1)),
		resume,
		assistant("2026-11-24T10:20:05Z", "req-c2", "model-a", "tool_use", usage(1, 1, 1)),
	})
	cutoff := time.Date(2026, 11, 24, 10, 20, 45, 0, time.UTC)
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, cutoff)
	child := sessionByID(t, extraction.Sessions, "achild").Invocations[0]
	// 10s completed, plus 45s provisional from the resume to the cutoff.
	if child.Turns != 2 || child.ActiveDurationSeconds != 55 || child.ProvisionalSeconds != 45 || child.CompletedAt != nil {
		t.Fatalf("want 55s with 45 provisional, got turns=%d active=%d provisional=%d completed=%v", child.Turns, child.ActiveDurationSeconds, child.ProvisionalSeconds, child.CompletedAt)
	}
	var interval *domain.CheckpointInterval
	for _, evidence := range extraction.CheckpointEvidence {
		if evidence.SessionID == "achild" && len(evidence.Intervals) == 1 {
			interval = &evidence.Intervals[0]
		}
	}
	if interval == nil || interval.ID != "resume-uuid" || !interval.SourceOpen || !interval.StartedAt.Equal(time.Date(2026, 11, 24, 10, 20, 0, 0, time.UTC)) {
		t.Fatalf("open resumed turn must be checkpoint evidence keyed by the resume record: %+v", extraction.CheckpointEvidence)
	}
}

// A turn superseded by a new prompt ends at its last owned record.
func TestSyntheticSupersededTurnEndsAtLastOwnedRecord(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:10Z", "req-1", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("t-1", "Bash", record{})),
		resultRecord("2026-11-24T10:00:15Z", "t-1", false, nil),
		// No terminal record: the next prompt supersedes the turn 45s later.
		prompt("2026-11-24T10:01:00Z", "sdk"),
		assistant("2026-11-24T10:01:10Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	row := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0]
	if row.Turns != 2 || row.ActiveDurationSeconds != 25 || row.ProvisionalSeconds != 0 || row.UserIterations != 2 {
		t.Fatalf("want 15s + 10s over 2 turns, got turns=%d active=%d provisional=%d iterations=%d", row.Turns, row.ActiveDurationSeconds, row.ProvisionalSeconds, row.UserIterations)
	}
}

// A later complete read replaces an earlier provisional measurement.
func TestSyntheticCompletionReplacesProvisionalSnapshot(t *testing.T) {
	store := newStore(t)
	records := []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:10Z", "req-1", "model-a", "tool_use", usage(1, 1, 1)),
	}
	store.root(syntheticFolder, syntheticRoot, records)
	_, first := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Date(2026, 11, 24, 10, 0, 30, 0, time.UTC))
	open := sessionByID(t, first.Sessions, syntheticRoot).Invocations[0]
	if open.ActiveDurationSeconds != 30 || open.ProvisionalSeconds != 30 || open.CompletedAt != nil {
		t.Fatalf("open turn must be provisional to the cutoff: %+v", open)
	}
	store.root(syntheticFolder, syntheticRoot, append(records, assistant("2026-11-24T10:00:40Z", "req-2", "model-a", "end_turn", usage(1, 1, 1))))
	_, second := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, time.Date(2026, 11, 24, 10, 5, 0, 0, time.UTC))
	closed := sessionByID(t, second.Sessions, syntheticRoot).Invocations[0]
	if closed.ActiveDurationSeconds != 40 || closed.ProvisionalSeconds != 0 || closed.CompletedAt == nil || *closed.CompletedAt != "2026-11-24T10:00:40Z" {
		t.Fatalf("measured time must replace the provisional snapshot: %+v", closed)
	}
	for _, evidence := range second.CheckpointEvidence {
		if len(evidence.Intervals) != 0 || !evidence.Complete {
			t.Fatalf("a closed turn leaves no open interval: %+v", evidence)
		}
	}
}

// Responses without requestId deduplicate by message.id; cache creation feeds
// peak context only, and thinking tokens become reasoning.
func TestSyntheticMessageIDFallbackAndTokenMapping(t *testing.T) {
	store := newStore(t)
	rich := record{"input_tokens": 10, "output_tokens": 20, "cache_read_input_tokens": 30, "cache_creation_input_tokens": 100, "output_tokens_details": record{"thinking_tokens": 7}}
	block := func(ts string, kind string) record {
		r := withoutKey(assistant(ts, "x", "model-a", "tool_use", rich, record{"type": kind}), "requestId")
		return withMessageField(r, "id", "msg-shared")
	}
	plain := record{"input_tokens": 50, "output_tokens": 1, "cache_read_input_tokens": 0}
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		block("2026-11-24T10:00:01Z", "thinking"),
		block("2026-11-24T10:00:02Z", "text"),
		block("2026-11-24T10:00:03Z", "text"),
		assistant("2026-11-24T10:00:04Z", "req-plain", "model-a", "end_turn", plain),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	tokens := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0].Tokens
	// Input 10+50, output 20+1, cache read 30+0, reasoning 7+0,
	// peak max(10+30+100, 50+0+0).
	want := domain.InvocationTokens{Input: 60, Output: 21, CacheRead: 30, Reasoning: 7, PeakContext: 140}
	if tokens != want {
		t.Fatalf("token mapping: want %+v, got %+v", want, tokens)
	}
}

// Copied fork history adds no turns, steering, tools, skills, delegations,
// compactions or parent links, and a copied response identity is never counted
// for the copying session even when an owned record repeats it.
func TestSyntheticForkedHistoryContributesNothing(t *testing.T) {
	store := newStore(t)
	forked := func(r record) record {
		copied := withoutKey(r, "cwd")
		copied["cwd"] = "/synthetic/source-project"
		copied["forkedFrom"] = record{"sessionId": "source-session", "messageUuid": "m-1"}
		copied["sessionId"] = "source-session"
		return copied
	}
	store.root(syntheticFolder, syntheticRoot, []record{
		forked(prompt("2026-11-24T09:00:00Z", "sdk")),
		forked(assistant("2026-11-24T09:00:01Z", "req-copied", "model-a", "tool_use", usage(500, 500, 500),
			toolUseBlock("copied-agent", "Agent", record{"subagent_type": "coder"}), toolUseBlock("copied-skill", "Skill", record{"skill": "copied"}))),
		forked(resultRecord("2026-11-24T09:00:02Z", "copied-skill", false, record{"success": true})),
		forked(record{"type": "system", "subtype": "compact_boundary", "timestamp": "2026-11-24T09:00:03Z"}),
		forked(assistant("2026-11-24T09:00:04Z", "req-copied", "model-a", "end_turn", usage(500, 500, 500))),
		{"type": "attachment", "timestamp": "2026-11-24T10:00:00Z"},
		prompt("2026-11-24T10:00:00Z", "sdk"),
		// An owned record repeating a copied identity is still inherited history.
		assistant("2026-11-24T10:00:01Z", "req-copied", "model-a", "tool_use", usage(500, 500, 500)),
		assistant("2026-11-24T10:00:05Z", "req-own", "model-a", "end_turn", usage(1, 2, 3)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "acopied", "copied-agent", "coder", []record{
		prompt("2026-11-24T09:00:01Z", ""),
		assistant("2026-11-24T09:00:02Z", "req-copied-child", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	if len(extraction.RootIncomplete) != 0 {
		t.Fatalf("copied records with the source sessionId must not make the root incomplete: %v", extraction.RootIncomplete)
	}
	if len(extraction.Sessions) != 1 {
		t.Fatalf("a copied Agent call must not prove a child: %d sessions", len(extraction.Sessions))
	}
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	row := root.Invocations[0]
	if row.Tokens.Input != 1 || row.Tokens.Output != 2 || row.Tokens.CacheRead != 3 {
		t.Fatalf("copied usage counted for the fork: %+v", row.Tokens)
	}
	if row.Turns != 1 || row.UserIterations != 1 || row.Compactions != 0 || row.ActiveDurationSeconds != 5 {
		t.Fatalf("copied lifecycle counted: turns=%d iterations=%d compactions=%d active=%d", row.Turns, row.UserIterations, row.Compactions, row.ActiveDurationSeconds)
	}
	if len(root.Skills) != 0 || len(root.Delegations) != 0 || len(root.Tools) != 0 || root.ParentID != nil {
		t.Fatalf("copied tools, skills, delegations or parent leaked: %+v", root)
	}
	if extraction.WorkingDirectory != "/synthetic/project" {
		t.Fatalf("identity must come from the first owned cwd, got %q", extraction.WorkingDirectory)
	}
	if !strings.Contains(strings.Join(extraction.Diagnostics, "\n"), "acopied") {
		t.Fatalf("the unproven copied child must be named: %v", extraction.Diagnostics)
	}
}

// Project identity comes from the first recorded cwd, never the folder name.
func TestSyntheticWorkingDirectoryComesFromRecordedCWD(t *testing.T) {
	store := newStore(t)
	first := prompt("2026-11-24T10:00:00Z", "sdk")
	first["cwd"] = "/synthetic/recorded-checkout"
	later := prompt("2026-11-24T10:01:00Z", "sdk")
	later["cwd"] = "/synthetic/other-checkout"
	store.root("-home-decoy-folder-name", syntheticRoot, []record{
		{"type": "system", "subtype": "init", "timestamp": "2026-11-24T09:59:59Z"},
		first,
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1)),
		later,
		assistant("2026-11-24T10:01:01Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	files, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	if extraction.WorkingDirectory != "/synthetic/recorded-checkout" {
		t.Fatalf("want first recorded cwd, got %q", extraction.WorkingDirectory)
	}
	if filepath.Base(filepath.Dir(files.RootPath)) != "-home-decoy-folder-name" {
		t.Fatalf("fixture must select the decoy folder: %s", files.RootPath)
	}
}

// Skills and delegations are owned by the calling session, failed or
// unfinished calls record no activation, and the delegation type falls back to
// the result's agentType.
func TestSyntheticSkillAndDelegationOwnership(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", record{"description": "plan it"}),
			toolUseBlock("s-open", "Skill", record{"skill": "never-finished"})),
		resultRecord("2026-11-24T10:00:40Z", "call-child", false, record{"agentId": "achild", "agentType": "planner"}),
		assistant("2026-11-24T10:00:41Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "planner", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		assistant("2026-11-24T10:00:03Z", "req-c1", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("s-1", "Skill", record{"skill": "review"})),
		resultRecord("2026-11-24T10:00:04Z", "s-1", false, nil),
		assistant("2026-11-24T10:00:05Z", "req-c2", "model-a", "tool_use", usage(1, 1, 1), toolUseBlock("s-2", "Skill", record{"skill": "review"})),
		resultRecord("2026-11-24T10:00:06Z", "s-2", false, record{"success": true}),
		assistant("2026-11-24T10:00:07Z", "req-c3", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	root := sessionByID(t, extraction.Sessions, syntheticRoot)
	child := sessionByID(t, extraction.Sessions, "achild")
	if len(root.Skills) != 0 {
		t.Fatalf("an unfinished root Skill call is not an activation, and child skills are not the root's: %+v", root.Skills)
	}
	if len(child.Skills) != 2 {
		t.Fatalf("want two child activations of the same skill: %+v", child.Skills)
	}
	for i, want := range []string{"2026-11-24T10:00:03Z", "2026-11-24T10:00:05Z"} {
		got := child.Skills[i]
		if got.Name != "review" || got.CallerRole == nil || *got.CallerRole != "planner" || got.InvokedAt == nil || *got.InvokedAt != want {
			t.Fatalf("child activation %d: %+v", i, got)
		}
	}
	if len(root.Tools) != 2 {
		t.Fatalf("root tools: %+v", root.Tools)
	}
	for _, tool := range root.Tools {
		if tool.Calls != 1 || tool.Errors != 0 {
			t.Fatalf("an unfinished call is neither an error nor uncounted: %+v", root.Tools)
		}
	}
	if len(root.Delegations) != 1 {
		t.Fatalf("delegations: %+v", root.Delegations)
	}
	delegation := root.Delegations[0]
	if delegation.AgentType == nil || *delegation.AgentType != "planner" || delegation.Description != "plan it" || delegation.InvokedAt == nil || *delegation.InvokedAt != "2026-11-24T10:00:01Z" || delegation.ChildSession != "achild" {
		t.Fatalf("delegation must carry the result agentType, description, call time and proven child: %+v", delegation)
	}
	running := false
	for _, event := range extraction.Events {
		if event.CallID == "s-open" && event.Status == "running" {
			running = true
		}
	}
	if !running {
		t.Fatal("an unfinished tool call must stay a running event")
	}
}

func TestSyntheticDisabledOptionalCategoriesAreOmitted(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("s-1", "Skill", record{"skill": "review"}), toolUseBlock("call-x", "Agent", record{"subagent_type": "coder"})),
		resultRecord("2026-11-24T10:00:02Z", "s-1", false, record{"success": true}),
		assistant("2026-11-24T10:00:03Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	for _, policy := range []domain.GenerationPolicy{{Skills: false, Delegations: true}, {Skills: true, Delegations: false}} {
		files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot})
		if err != nil {
			t.Fatal(err)
		}
		tree, err := ReadTree(t.Context(), files, nil)
		if err != nil {
			t.Fatal(err)
		}
		extraction, err := tree.Prove().Facts(laterCutoff, &policy)
		if err != nil {
			t.Fatal(err)
		}
		root := sessionByID(t, extraction.Sessions, syntheticRoot)
		if (len(root.Skills) == 0) != !policy.Skills || (len(root.Delegations) == 0) != !policy.Delegations {
			t.Fatalf("policy %+v not applied: skills=%+v delegations=%+v", policy, root.Skills, root.Delegations)
		}
		if len(root.Tools) != 2 {
			t.Fatalf("tool counts do not depend on optional categories: %+v", root.Tools)
		}
	}
}

// Compactions are counted on the owner that recorded them.
func TestSyntheticCompactionsCountPerOwner(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", record{"subagent_type": "coder"})),
	})
	boundary := func(ts string) record {
		return record{"type": "system", "subtype": "compact_boundary", "timestamp": ts, "compactMetadata": record{"trigger": "manual"}, "logicalParentUuid": "p"}
	}
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "coder", []record{
		prompt("2026-11-24T10:00:02Z", ""),
		boundary("2026-11-24T10:00:03Z"),
		boundary("2026-11-24T10:00:04Z"),
		assistant("2026-11-24T10:00:05Z", "req-c1", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	if got := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0].Compactions; got != 0 {
		t.Fatalf("root recorded a child's compactions: %d", got)
	}
	if got := sessionByID(t, extraction.Sessions, "achild").Invocations[0].Compactions; got != 2 {
		t.Fatalf("want two child compactions, got %d", got)
	}
	compactionEvents := 0
	for _, event := range extraction.Events {
		if event.EventType == domain.EventTypeCompaction {
			compactionEvents++
			if event.SessionID != "achild" {
				t.Fatalf("compaction event owned by %s", event.SessionID)
			}
		}
	}
	if compactionEvents != 2 {
		t.Fatalf("want two compaction events, got %d", compactionEvents)
	}
}

// A terminal response written as several content-block records ends its turn at
// the last record of that response, and its usage is still counted once.
func TestSyntheticTerminalResponseBlocksExtendTheClosedTurnOnce(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:05Z", "req-1", "model-a", "end_turn", usage(4, 5, 6), record{"type": "thinking"}),
		assistant("2026-11-24T10:00:08Z", "req-1", "model-a", "end_turn", usage(4, 5, 6), record{"type": "text"}),
		// A different response after the turn closed adds no time.
		assistant("2026-11-24T10:00:30Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	_, extraction := extract(t, store.home, DiscoveryOptions{SessionID: syntheticRoot}, laterCutoff)
	row := sessionByID(t, extraction.Sessions, syntheticRoot).Invocations[0]
	if row.ActiveDurationSeconds != 8 || row.Turns != 1 || row.Tokens.Input != 5 || row.Tokens.Output != 6 || row.Tokens.CacheRead != 7 {
		t.Fatalf("want one 8s turn and usage (4,5,6)+(1,1,1): active=%d turns=%d tokens=%+v", row.ActiveDurationSeconds, row.Turns, row.Tokens)
	}
}
