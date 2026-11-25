package backfill

// Synthetic regression fixtures for AL-98 review findings DATA-01, DATA-03,
// DATA-04 and DATA-05. Every Claude Code store is written to a temp directory
// with native record keys only; no transcript content, ~/.claude or live ticket
// workspace is read. Expected values are hand-computed from the records below.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

const (
	reviewRoot  = "review-root-synthetic"
	reviewChild = "areviewchild"
)

// Typed synthetic Claude Code records mirroring the tester's fixtures in
// claude_recovery_test.go, with an explicit cwd per store. An empty Type
// encodes a record with no type, which makes its owner incomplete.
type reviewRecord struct {
	Type          string         `json:"type"`
	SessionID     string         `json:"sessionId"`
	AgentID       string         `json:"agentId,omitempty"`
	IsSidechain   bool           `json:"isSidechain,omitempty"`
	UUID          string         `json:"uuid,omitempty"`
	Timestamp     string         `json:"timestamp"`
	CWD           string         `json:"cwd"`
	PromptSource  string         `json:"promptSource,omitempty"`
	RequestID     string         `json:"requestId,omitempty"`
	ToolUseResult *reviewResult  `json:"toolUseResult,omitempty"`
	Message       *reviewMessage `json:"message,omitempty"`
}

type reviewMessage struct {
	ID         string          `json:"id,omitempty"`
	Model      string          `json:"model,omitempty"`
	Role       string          `json:"role,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	Usage      *reviewUsage    `json:"usage,omitempty"`
	Content    json.RawMessage `json:"content"`
}

type reviewUsage struct {
	Input     int `json:"input_tokens"`
	Output    int `json:"output_tokens"`
	CacheRead int `json:"cache_read_input_tokens"`
}

type reviewBlock struct {
	Type      string       `json:"type"`
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Input     *reviewInput `json:"input,omitempty"`
	ToolUseID string       `json:"tool_use_id,omitempty"`
}

type reviewInput struct {
	SubagentType string `json:"subagent_type"`
}

type reviewResult struct {
	AgentID string `json:"agentId"`
}

type reviewStore struct {
	t    *testing.T
	home string
}

func newReviewStore(t *testing.T) reviewStore {
	return reviewStore{t: t, home: t.TempDir()}
}

func (s reviewStore) check(data []byte, err error) json.RawMessage {
	s.t.Helper()
	if err != nil {
		s.t.Fatal(err)
	}
	return data
}

func (s reviewStore) text(value string) json.RawMessage {
	data, err := json.Marshal(value)
	return s.check(data, err)
}

func (s reviewStore) blocks(value []reviewBlock) json.RawMessage {
	data, err := json.Marshal(value)
	return s.check(data, err)
}

func (s reviewStore) line(value reviewRecord) json.RawMessage {
	data, err := json.Marshal(value)
	return s.check(data, err)
}

func (s reviewStore) write(path, cwd, agentID string, lines []reviewRecord) {
	s.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		s.t.Fatal(err)
	}
	var body strings.Builder
	for _, line := range lines {
		line.SessionID, line.CWD = reviewRoot, cwd
		if agentID != "" {
			line.AgentID, line.IsSidechain = agentID, true
		}
		body.Write(s.line(line))
		body.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(body.String()), 0o600); err != nil {
		s.t.Fatal(err)
	}
}

func (s reviewStore) prompt(at, uuid, source string) reviewRecord {
	return reviewRecord{Type: "user", UUID: uuid, Timestamp: at, PromptSource: source, Message: &reviewMessage{Role: "user", Content: s.text("synthetic")}}
}

func (s reviewStore) response(at, request string, input int, stop string, blocks []reviewBlock) reviewRecord {
	if blocks == nil {
		blocks = []reviewBlock{}
	}
	return reviewRecord{Type: "assistant", Timestamp: at, RequestID: request, Message: &reviewMessage{ID: "m-" + request, Model: "model-review", StopReason: stop,
		Usage: &reviewUsage{Input: input, Output: 1}, Content: s.blocks(blocks)}}
}

// root writes a root that delegates to reviewChild. A malformed root makes the
// tree unrecoverable while its child file stays readable.
func (s reviewStore) root(cwd string, input int, malformed bool) {
	s.t.Helper()
	folder := filepath.Join(s.home, "projects", "-review-project")
	lines := []reviewRecord{
		s.prompt("2026-01-01T00:00:00Z", "u-1", "sdk"),
		s.response("2026-01-01T00:00:01Z", "r-1", input, "tool_use", []reviewBlock{{Type: "tool_use", ID: "call-1", Name: "Agent", Input: &reviewInput{SubagentType: "reviewer"}}}),
		{Type: "user", Timestamp: "2026-01-01T00:00:05Z", ToolUseResult: &reviewResult{AgentID: reviewChild}, Message: &reviewMessage{Role: "user", Content: s.blocks([]reviewBlock{{Type: "tool_result", ToolUseID: "call-1"}})}},
		s.response("2026-01-01T00:00:06Z", "r-2", input, "end_turn", nil),
	}
	if malformed {
		lines = append(lines, reviewRecord{Timestamp: "2026-01-01T00:00:07Z"})
	}
	s.write(filepath.Join(folder, reviewRoot+".jsonl"), cwd, "", lines)
	subagents := filepath.Join(folder, reviewRoot, "subagents")
	s.write(filepath.Join(subagents, "agent-"+reviewChild+".jsonl"), cwd, reviewChild, []reviewRecord{
		s.prompt("2026-01-01T00:00:02Z", "c-1", ""),
		s.response("2026-01-01T00:00:04Z", "r-c1", 7, "end_turn", nil),
	})
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+reviewChild+".meta.json"), []byte(`{"agentType":"reviewer","toolUseId":"call-1"}`), 0o600); err != nil {
		s.t.Fatal(err)
	}
}

func reviewLedger(ticket string, sessions ...domain.SessionFact) domain.AgenticFactLedger {
	version := domain.CurrentSchemaVersion
	return domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: sessions}
}

func reviewOwner(id string, parent *string, input int64, rows int) domain.SessionFact {
	owner := domain.SessionFact{ID: id, ParentID: parent, Source: domain.SourceClaude, User: "local"}
	for i := 0; i < rows; i++ {
		owner.Invocations = append(owner.Invocations, domain.InvocationFact{Model: "stored-model-" + string(rune('a'+i)), Tokens: domain.InvocationTokens{Input: input}})
	}
	return owner
}

func reviewSessions(t *testing.T, workspace string) map[string]domain.SessionFact {
	t.Helper()
	dto, _ := readRecoveryLedger(t, workspace)
	byID := map[string]domain.SessionFact{}
	for _, session := range dto.ToDomain().Sessions {
		byID[session.ID] = session
	}
	return byID
}

func TestBackfillClaudeStoreWithoutProjectsFailsBeforePublication(t *testing.T) {
	for _, source := range []string{"claude", "all"} {
		t.Run(source, func(t *testing.T) {
			root := newBackfillProject(t)
			ticket := "TEST-CLAUDE-NO-PROJECTS-" + strings.ToUpper(source)
			parent := reviewRoot
			workspace := createTestWorkspace(t, root, ticket, reviewLedger(ticket, reviewOwner(reviewRoot, nil, 5, 1), reviewOwner(reviewChild, &parent, 5, 1)))
			_, before := readRecoveryLedger(t, workspace)
			// A readable directory that is not a Claude Code store, such as a home
			// directory, must not be accepted.
			notStore := t.TempDir()
			if err := os.WriteFile(filepath.Join(notStore, "unrelated.txt"), []byte("synthetic"), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: source, ClaudeHome: notStore, RecomputeDB: true})
			if err == nil || !strings.Contains(err.Error(), "projects") || !strings.Contains(err.Error(), "--claude-home") {
				t.Fatalf("a store without projects/ must fail with an actionable error, got %v", err)
			}
			if _, after := readRecoveryLedger(t, workspace); string(after) != string(before) {
				t.Fatal("ledger changed although the selected Claude Code store was rejected")
			}
		})
	}
}

func TestBackfillClaudeReportsSmallerCompleteReplacement(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-SMALLER"
	parent := reviewRoot
	// Stored root has two rows and 900 tokens; the source has one row and 6 tokens.
	workspace := createTestWorkspace(t, root, ticket, reviewLedger(ticket, reviewOwner(reviewRoot, nil, 450, 2), reviewOwner(reviewChild, &parent, 1, 1)))
	store := newReviewStore(t)
	store.root(root, 2, false)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: store.home, RecomputeDB: true})
	if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	notice := "Claude Code owner " + reviewRoot + " replaced by a complete read with smaller totals: invocations 2 -> 1, tokens 900 -> 6"
	if !strings.Contains(strings.Join(report.Warnings, "\n"), notice) {
		t.Fatalf("smaller complete replacement must be reported as %q, got %v", notice, report.Warnings)
	}
	if report.Results[0].RefreshIncomplete {
		t.Fatalf("a reported replacement is not an incomplete refresh: %+v", report.Results[0])
	}
	if got := reviewSessions(t, workspace)[reviewRoot]; len(got.Invocations) != 1 || got.Invocations[0].Tokens.Input != 4 {
		t.Fatalf("complete read must replace the stored snapshot: %+v", got)
	}
}

func TestBackfillClaudeReadsEachUnrecoverableTreeOnce(t *testing.T) {
	root := newBackfillProject(t)
	ticket := "TEST-CLAUDE-READ-ONCE"
	parent := reviewRoot
	workspace := createTestWorkspace(t, root, ticket, reviewLedger(ticket, reviewOwner(reviewRoot, nil, 5, 1), reviewOwner(reviewChild, &parent, 5, 1)))
	store := newReviewStore(t)
	store.root(root, 2, true)
	report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: store.home, RecomputeDB: true})
	if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
		t.Fatalf("backfill failed: %+v %v", report, err)
	}
	joined := strings.Join(report.Results[0].Diagnostics, "\n")
	if count := strings.Count(joined, "is incomplete; previous snapshot retained"); count != 1 {
		t.Fatalf("an unrecoverable tree must be read and reported once, got %d reports:\n%s", count, joined)
	}
	if got := reviewSessions(t, workspace)[reviewRoot]; got.Invocations[0].Tokens.Input != 5 {
		t.Fatalf("an incomplete root keeps its stored snapshot: %+v", got)
	}
	if owner := claudeCoverageOwner(t, report.Results[0], reviewRoot); owner.Status != OwnerRetained || !strings.Contains(owner.Reason, "source session incomplete") {
		t.Fatalf("an incomplete root must be retained with its reason: %+v", owner)
	}
}

func TestBackfillClaudeAdoptsUnrecordedRootOnlyInsideProject(t *testing.T) {
	t.Run("inside the project", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-ADOPT-INSIDE"
		parent := reviewRoot
		workspace := createTestWorkspace(t, root, ticket, reviewLedger(ticket, reviewOwner(reviewChild, &parent, 5, 1)))
		store := newReviewStore(t)
		store.root(root, 2, false)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: store.home, RecomputeDB: true})
		if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		if _, added := reviewSessions(t, workspace)[reviewRoot]; !added {
			t.Fatal("a proven root inside the project must be adopted")
		}
		if !strings.Contains(strings.Join(report.Warnings, "\n"), "Claude Code owner "+reviewRoot+" added from recorded tree "+reviewRoot) {
			t.Fatalf("a newly added owner must be named: %v", report.Warnings)
		}
	})
	t.Run("outside the project", func(t *testing.T) {
		root := newBackfillProject(t)
		ticket := "TEST-CLAUDE-ADOPT-OUTSIDE"
		parent := reviewRoot
		workspace := createTestWorkspace(t, root, ticket, reviewLedger(ticket, reviewOwner(reviewChild, &parent, 5, 1)))
		store := newReviewStore(t)
		store.root(t.TempDir(), 2, false)
		report, err := NewEngine().Execute(BackfillOptions{Project: root, TicketID: ticket, Source: "claude", ClaudeHome: store.home, RecomputeDB: true})
		if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v %v", report, err)
		}
		after := reviewSessions(t, workspace)
		if _, added := after[reviewRoot]; added {
			t.Fatal("a root recorded outside the project must not be adopted")
		}
		if !strings.Contains(strings.Join(report.Results[0].Diagnostics, "\n"), "Claude Code root "+reviewRoot+" not added") {
			t.Fatalf("the skipped root must be diagnosed: %v", report.Results[0].Diagnostics)
		}
		if got := after[reviewChild]; len(got.Invocations) != 1 || got.Invocations[0].Tokens.Input != 7 {
			t.Fatalf("the recorded child in the tree still refreshes: %+v", got)
		}
	})
}
