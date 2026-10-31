package domain_test

import (
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func testTicketConfig() domain.TicketConfig {
	return domain.TicketConfig{
		Prefixes: []domain.TicketPrefix{
			{Prefix: "AL", Description: "Core development"},
			{Prefix: "GEN", Description: "General tasks", TicketType: "chore"},
			{Prefix: "PO", Description: "Product owner", TicketType: "docs"},
		},
		CaseInsensitive: true,
	}
}

func TestInferTicketID(t *testing.T) {
	cfg := testTicketConfig()

	tests := []struct {
		name           string
		title          string
		expectedID     *string
		expectedConf   float64
		expectedTarget string
	}{
		{
			name:           "exact prefix uppercase",
			title:          "Task AL-17: implement live database hook",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "case-insensitive lowercase prefix",
			title:          "working on al-17",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "alternate configured prefix GEN",
			title:          "GEN-42 bugfix",
			expectedConf:   1.0,
			expectedTarget: "GEN-42",
		},
		{
			name:           "alternate configured prefix lowercase po",
			title:          "po-3 spec decomposition",
			expectedConf:   1.0,
			expectedTarget: "PO-3",
		},
		{
			name:         "general debugging title without ticket",
			title:        "General debugging session",
			expectedConf: 0.0,
		},
		{
			name:         "quick question title without ticket",
			title:        "Quick question",
			expectedConf: 0.0,
		},
		{
			name:         "empty string",
			title:        "",
			expectedConf: 0.0,
		},
		{
			name:         "whitespace string",
			title:        "    \t\n  ",
			expectedConf: 0.0,
		},
		{
			name:         "unconfigured prefix PROJ",
			title:        "PROJ-1234 other project task",
			expectedConf: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, conf := domain.InferTicketID(tt.title, cfg)

			if conf != tt.expectedConf {
				t.Errorf("InferTicketID(%q) confidence = %f, expected %f", tt.title, conf, tt.expectedConf)
			}

			if tt.expectedConf == 0.0 {
				if gotID != nil {
					t.Errorf("InferTicketID(%q) expected nil, got %v", tt.title, *gotID)
				}
			} else {
				if gotID == nil {
					t.Fatalf("InferTicketID(%q) expected %s, got nil", tt.title, tt.expectedTarget)
				}
				if *gotID != tt.expectedTarget {
					t.Errorf("InferTicketID(%q) = %s, expected %s", tt.title, *gotID, tt.expectedTarget)
				}
			}
		})
	}
}

func TestInferTicketID_AdversarialSuite(t *testing.T) {
	cfg := testTicketConfig()

	tests := []struct {
		name           string
		title          string
		expectedConf   float64
		expectedTarget string
	}{
		// Case insensitivity variations
		{
			name:           "mixed case Al-17",
			title:          "Working on Al-17 refactoring",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "mixed case aL-17",
			title:          "Fix for aL-17 here",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "mixed case Gen-4",
			title:          "Session Gen-4 cleanup",
			expectedConf:   1.0,
			expectedTarget: "GEN-4",
		},
		{
			name:           "mixed case Po-2",
			title:          "Requirements Po-2 spec",
			expectedConf:   1.0,
			expectedTarget: "PO-2",
		},
		// Delimiters and punctuation surrounding ticket
		{
			name:           "bracket enclosed [AL-17]",
			title:          "[AL-17] Non-locking SQLite database hook",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "parentheses enclosed (PO-2)",
			title:          "Spec review (PO-2)",
			expectedConf:   1.0,
			expectedTarget: "PO-2",
		},
		{
			name:           "conventional commit style feat(al-17):",
			title:          "feat(al-17): implement live telemetry endpoint",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "colon prefix AL-17:",
			title:          "AL-17: live query and streaming",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "dash separator AL-17 - summary",
			title:          "AL-17 - summary of live tree",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "trailing period AL-17.",
			title:          "Completed work on AL-17.",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		// Whitespace torture tests
		{
			name:           "heavy leading and trailing whitespace",
			title:          "   \t\t\n  AL-17  \r\n\t  ",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:           "extra spaces between words and ticket",
			title:          "Investigating    database    lock    in    al-17    now",
			expectedConf:   1.0,
			expectedTarget: "AL-17",
		},
		{
			name:         "pure spaces",
			title:        "     ",
			expectedConf: 0.0,
		},
		{
			name:         "mixed whitespace escape sequences",
			title:        "\n\r\t \v\f",
			expectedConf: 0.0,
		},
		// Unconfigured and non-matching prefixes
		{
			name:         "unconfigured prefix JIRA-123",
			title:        "Working on JIRA-123 task",
			expectedConf: 0.0,
		},
		{
			name:         "unconfigured prefix BUG-999",
			title:        "Resolving BUG-999 fatal crash",
			expectedConf: 0.0,
		},
		{
			name:         "unconfigured prefix TASK-1",
			title:        "TASK-1 initialize project",
			expectedConf: 0.0,
		},
		// Malformed numbers and missing parts
		{
			name:         "prefix without dash AL17",
			title:        "Working on AL17",
			expectedConf: 0.0,
		},
		{
			name:         "prefix with dash but no digits AL-",
			title:        "Task AL- unfinished",
			expectedConf: 0.0,
		},
		{
			name:         "prefix with non-digits after dash AL-abc",
			title:        "Branch AL-abc",
			expectedConf: 0.0,
		},
		{
			name:         "prefix separated by space AL - 17",
			title:        "Task AL - 17",
			expectedConf: 0.0,
		},
		// Plain natural language titles
		{
			name:         "generic conversation",
			title:        "Can you explain how SQLite WAL mode works?",
			expectedConf: 0.0,
		},
		{
			name:         "code snippet in title",
			title:        "func main() { fmt.Println() }",
			expectedConf: 0.0,
		},
		{
			name:         "pure numeric string",
			title:        "123456789",
			expectedConf: 0.0,
		},
		{
			name:         "punctuation only",
			title:        "!@#$%^&*()_+=-[]{}|;:',.<>/?",
			expectedConf: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, conf := domain.InferTicketID(tt.title, cfg)

			if conf != tt.expectedConf {
				t.Errorf("InferTicketID(%q) confidence = %f, expected %f", tt.title, conf, tt.expectedConf)
			}

			if tt.expectedConf == 0.0 {
				if gotID != nil {
					t.Errorf("InferTicketID(%q) expected nil, got %q", tt.title, *gotID)
				}
			} else {
				if gotID == nil {
					t.Fatalf("InferTicketID(%q) expected %s, got nil", tt.title, tt.expectedTarget)
				}
				if *gotID != tt.expectedTarget {
					t.Errorf("InferTicketID(%q) = %s, expected %s", tt.title, *gotID, tt.expectedTarget)
				}
			}
		})
	}
}

func TestClassifyToolExecution(t *testing.T) {
	tests := []struct {
		name, tool, status, output string
		consecutiveErrors          int
		wantError                  bool
		wantSeverity               domain.FailureSeverity
	}{
		{"completed semantic error text", "bash", "completed", "syntax error in documentation", 0, false, domain.SeverityNone},
		{"completed traceback text", "read", "completed", "Traceback (most recent call last): example", 0, false, domain.SeverityNone},
		{"explicit status error", "edit", "error", "", 0, true, domain.SeverityMedium},
		{"explicit failed status", "bash", "failed", "", 0, true, domain.SeverityMedium},
		{"nonzero exit", "bash", "completed", "exit status 2", 0, true, domain.SeverityMedium},
		{"grep no match exit one", "grep", "completed", "exit status 1", 0, false, domain.SeverityNone},
		{"grep no match cannot override explicit error status", "grep", "error", "exit status 1", 0, true, domain.SeverityMedium},
		{"glob no match cannot override explicit failed status", "glob", "failed", "no matches", 0, true, domain.SeverityMedium},
		{"glob zero matches", "glob", "completed", "0 matches", 0, false, domain.SeverityNone},
		{"timeout", "bash", "completed", "operation timed out", 0, true, domain.SeverityMedium},
		{"permission denied", "bash", "completed", "permission denied", 0, true, domain.SeverityMedium},
		{"completed panic diagnostic", "bash", "completed", "panic: runtime failure", 0, false, domain.SeverityNone},
		{"consecutive errors escalate", "bash", "failed", "exit status 1", 2, true, domain.SeverityHigh},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ClassifyToolExecution(tt.tool, tt.status, tt.output, tt.consecutiveErrors)
			if got.IsError != tt.wantError || got.Severity != tt.wantSeverity {
				t.Fatalf("got error=%v severity=%s; want error=%v severity=%s", got.IsError, got.Severity, tt.wantError, tt.wantSeverity)
			}
			if got.IsError != (got.ErrorMessage != nil) {
				t.Fatalf("error message presence does not match failure state: %#v", got)
			}
		})
	}
}

func TestClassifyToolFailure_MessageTruncation(t *testing.T) {
	// Create an oversized error message of 5000 characters
	largeOutput := "Error: " + strings.Repeat("A", 4993)
	isErr, sev, errMsg := domain.ClassifyToolFailure("bash", "error", largeOutput, 0)

	if !isErr {
		t.Fatal("expected isError to be true")
	}
	if sev != "medium" {
		t.Errorf("expected medium severity, got %s", sev)
	}
	if errMsg == nil {
		t.Fatal("expected non-nil error message")
	}
	if len(*errMsg) != 2048 {
		t.Errorf("expected error message to be truncated to exactly 2048 characters, got %d", len(*errMsg))
	}
}

func TestBuildSessionTree(t *testing.T) {
	parentID := "ses_root_1"
	parent2ID := "ses_root_2"
	subParentID := "ses_child_1"

	nodes := []*domain.LiveSessionNode{
		{
			ID:          "ses_child_2",
			ParentID:    &parentID,
			Agent:       "tester",
			Role:        domain.AgentRole("tester"),
			Title:       "Testing phase",
			TimeCreated: "2026-10-19T10:15:00Z",
			TimeUpdated: "2026-10-19T10:20:00Z",
		},
		{
			ID:          "ses_child_1",
			ParentID:    &parentID,
			Agent:       "coder",
			Role:        domain.AgentRole("coder"),
			Title:       "Coding phase",
			TimeCreated: "2026-10-19T10:05:00Z",
			TimeUpdated: "2026-10-19T10:14:00Z",
		},
		{
			ID:          "ses_root_1",
			ParentID:    nil,
			Agent:       "build",
			Role:        domain.AgentRole("orchestrator"),
			Title:       "Root session 1",
			TimeCreated: "2026-10-19T10:00:00Z",
			TimeUpdated: "2026-10-19T10:25:00Z",
		},
		{
			ID:          "ses_root_2",
			ParentID:    nil,
			Agent:       "build",
			Role:        domain.AgentRole("orchestrator"),
			Title:       "Root session 2",
			TimeCreated: "2026-10-19T09:00:00Z",
			TimeUpdated: "2026-10-19T10:30:00Z", // Latest update
		},
		{
			ID:          "ses_grandchild",
			ParentID:    &subParentID,
			Agent:       "reviewer",
			Role:        domain.AgentRole("reviewer"),
			Title:       "Review child",
			TimeCreated: "2026-10-19T10:10:00Z",
			TimeUpdated: "2026-10-19T10:12:00Z",
		},
		{
			ID:          "ses_orphan",
			ParentID:    &parent2ID, // Will be child of root 2
			Agent:       "planner",
			Role:        domain.AgentRole("planner"),
			Title:       "Planning",
			TimeCreated: "2026-10-19T09:05:00Z",
			TimeUpdated: "2026-10-19T09:10:00Z",
		},
	}

	roots := domain.BuildSessionTree(nodes)

	if len(roots) != 2 {
		t.Fatalf("expected 2 root sessions, got %d", len(roots))
	}

	// Root 2 has TimeUpdated 10:30:00Z, so it must be first
	if roots[0].ID != "ses_root_2" {
		t.Errorf("expected first root to be ses_root_2, got %s", roots[0].ID)
	}
	if roots[1].ID != "ses_root_1" {
		t.Errorf("expected second root to be ses_root_1, got %s", roots[1].ID)
	}

	// Root 2 has 1 sub-session
	if len(roots[0].SubSessions) != 1 {
		t.Fatalf("expected root 2 to have 1 sub-session, got %d", len(roots[0].SubSessions))
	}
	if roots[0].SubSessions[0].ID != "ses_orphan" {
		t.Errorf("expected root 2 child to be ses_orphan, got %s", roots[0].SubSessions[0].ID)
	}

	// Root 1 has 2 sub-sessions (child 1 and child 2), ordered by TimeCreated ascending
	if len(roots[1].SubSessions) != 2 {
		t.Fatalf("expected root 1 to have 2 sub-sessions, got %d", len(roots[1].SubSessions))
	}
	if roots[1].SubSessions[0].ID != "ses_child_1" {
		t.Errorf("expected first child of root 1 to be ses_child_1 (created earlier), got %s", roots[1].SubSessions[0].ID)
	}
	if roots[1].SubSessions[1].ID != "ses_child_2" {
		t.Errorf("expected second child of root 1 to be ses_child_2, got %s", roots[1].SubSessions[1].ID)
	}

	// ses_child_1 has 1 grandchild (ses_grandchild)
	child1 := roots[1].SubSessions[0]
	if len(child1.SubSessions) != 1 {
		t.Fatalf("expected ses_child_1 to have 1 grandchild, got %d", len(child1.SubSessions))
	}
	if child1.SubSessions[0].ID != "ses_grandchild" {
		t.Errorf("expected grandchild to be ses_grandchild, got %s", child1.SubSessions[0].ID)
	}
}

func TestBuildSessionTree_OrphanWithoutParentBecomesRoot(t *testing.T) {
	nonExistentParent := "ses_missing_parent"
	nodes := []*domain.LiveSessionNode{
		{
			ID:          "ses_isolated",
			ParentID:    &nonExistentParent,
			Agent:       "coder",
			Role:        domain.AgentRole("coder"),
			Title:       "Isolated task",
			TimeCreated: "2026-10-19T10:00:00Z",
			TimeUpdated: "2026-10-19T10:00:00Z",
		},
	}

	roots := domain.BuildSessionTree(nodes)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root session for orphaned node without parent, got %d", len(roots))
	}
	if roots[0].ID != "ses_isolated" {
		t.Errorf("expected ses_isolated to be promoted to root, got %s", roots[0].ID)
	}
}

func TestBuildSessionTree_AdversarialEdgeCases(t *testing.T) {
	t.Run("empty input returns non-nil empty slice", func(t *testing.T) {
		roots := domain.BuildSessionTree([]*domain.LiveSessionNode{})
		if roots == nil {
			t.Fatal("expected non-nil empty slice, got nil")
		}
		if len(roots) != 0 {
			t.Errorf("expected 0 roots, got %d", len(roots))
		}
	})

	t.Run("nil entries in nodes slice are safely ignored", func(t *testing.T) {
		nodes := []*domain.LiveSessionNode{
			nil,
			{
				ID:          "ses_valid",
				ParentID:    nil,
				Title:       "Valid root",
				TimeCreated: "2026-10-19T10:00:00Z",
				TimeUpdated: "2026-10-19T10:05:00Z",
			},
			nil,
		}
		roots := domain.BuildSessionTree(nodes)
		if len(roots) != 1 {
			t.Fatalf("expected 1 root, got %d", len(roots))
		}
		if roots[0].ID != "ses_valid" {
			t.Errorf("expected ses_valid, got %s", roots[0].ID)
		}
	})

	t.Run("self-referencing parent is promoted to root safely", func(t *testing.T) {
		selfID := "ses_loop"
		nodes := []*domain.LiveSessionNode{
			{
				ID:          selfID,
				ParentID:    &selfID, // Self-referential
				Title:       "Loop session",
				TimeCreated: "2026-10-19T10:00:00Z",
				TimeUpdated: "2026-10-19T10:00:00Z",
			},
		}
		roots := domain.BuildSessionTree(nodes)
		if len(roots) != 1 {
			t.Fatalf("expected self-referencing session to become root, got %d roots", len(roots))
		}
		if roots[0].ID != selfID {
			t.Errorf("expected root ID %s, got %s", selfID, roots[0].ID)
		}
	})

	t.Run("multi-root sorting with identical updated times uses created time tie-breaker", func(t *testing.T) {
		nodes := []*domain.LiveSessionNode{
			{
				ID:          "ses_older_created",
				ParentID:    nil,
				TimeCreated: "2026-10-19T09:00:00Z",
				TimeUpdated: "2026-10-19T10:00:00Z",
			},
			{
				ID:          "ses_newer_created",
				ParentID:    nil,
				TimeCreated: "2026-10-19T09:30:00Z",
				TimeUpdated: "2026-10-19T10:00:00Z",
			},
		}
		roots := domain.BuildSessionTree(nodes)
		if len(roots) != 2 {
			t.Fatalf("expected 2 roots, got %d", len(roots))
		}
		if roots[0].ID != "ses_newer_created" {
			t.Errorf("expected ses_newer_created first due to tie-breaker, got %s", roots[0].ID)
		}
		if roots[1].ID != "ses_older_created" {
			t.Errorf("expected ses_older_created second, got %s", roots[1].ID)
		}
	})

	t.Run("sibling sorting with identical created times uses ID tie-breaker", func(t *testing.T) {
		parentID := "ses_parent"
		nodes := []*domain.LiveSessionNode{
			{
				ID:          parentID,
				ParentID:    nil,
				TimeCreated: "2026-10-19T09:00:00Z",
				TimeUpdated: "2026-10-19T10:00:00Z",
			},
			{
				ID:          "child_b",
				ParentID:    &parentID,
				TimeCreated: "2026-10-19T09:10:00Z",
				TimeUpdated: "2026-10-19T09:20:00Z",
			},
			{
				ID:          "child_a",
				ParentID:    &parentID,
				TimeCreated: "2026-10-19T09:10:00Z",
				TimeUpdated: "2026-10-19T09:20:00Z",
			},
		}
		roots := domain.BuildSessionTree(nodes)
		if len(roots) != 1 {
			t.Fatalf("expected 1 root, got %d", len(roots))
		}
		children := roots[0].SubSessions
		if len(children) != 2 {
			t.Fatalf("expected 2 children, got %d", len(children))
		}
		// Alphabetical ID tie-breaker: child_a < child_b
		if children[0].ID != "child_a" {
			t.Errorf("expected first child to be child_a, got %s", children[0].ID)
		}
		if children[1].ID != "child_b" {
			t.Errorf("expected second child to be child_b, got %s", children[1].ID)
		}
	})

	t.Run("deep 4-level hierarchy preserves recursive nesting", func(t *testing.T) {
		rID := "l0"
		c1ID := "l1"
		c2ID := "l2"
		c3ID := "l3"

		nodes := []*domain.LiveSessionNode{
			{ID: rID, ParentID: nil, TimeCreated: "2026-10-19T10:00:00Z", TimeUpdated: "2026-10-19T10:40:00Z"},
			{ID: c1ID, ParentID: &rID, TimeCreated: "2026-10-19T10:10:00Z", TimeUpdated: "2026-10-19T10:30:00Z"},
			{ID: c2ID, ParentID: &c1ID, TimeCreated: "2026-10-19T10:15:00Z", TimeUpdated: "2026-10-19T10:25:00Z"},
			{ID: c3ID, ParentID: &c2ID, TimeCreated: "2026-10-19T10:20:00Z", TimeUpdated: "2026-10-19T10:22:00Z"},
		}

		roots := domain.BuildSessionTree(nodes)
		if len(roots) != 1 {
			t.Fatalf("expected 1 root, got %d", len(roots))
		}
		if len(roots[0].SubSessions) != 1 {
			t.Fatalf("expected root to have 1 child, got %d", len(roots[0].SubSessions))
		}
		level1 := roots[0].SubSessions[0]
		if level1.ID != c1ID {
			t.Errorf("expected level 1 ID %s, got %s", c1ID, level1.ID)
		}
		if len(level1.SubSessions) != 1 {
			t.Fatalf("expected level 1 to have 1 child, got %d", len(level1.SubSessions))
		}
		level2 := level1.SubSessions[0]
		if level2.ID != c2ID {
			t.Errorf("expected level 2 ID %s, got %s", c2ID, level2.ID)
		}
		if len(level2.SubSessions) != 1 {
			t.Fatalf("expected level 2 to have 1 child, got %d", len(level2.SubSessions))
		}
		level3 := level2.SubSessions[0]
		if level3.ID != c3ID {
			t.Errorf("expected level 3 ID %s, got %s", c3ID, level3.ID)
		}
		if len(level3.SubSessions) != 0 {
			t.Errorf("expected leaf node to have 0 sub-sessions, got %d", len(level3.SubSessions))
		}
	})
}

func TestEventsFromLiveTrace(t *testing.T) {
	t.Run("nil and empty trace return empty slice", func(t *testing.T) {
		if evs := domain.EventsFromLiveTrace(nil); len(evs) != 0 {
			t.Fatalf("expected empty events for nil trace, got %d", len(evs))
		}
		if evs := domain.EventsFromLiveTrace(&domain.LiveTraceView{}); len(evs) != 0 {
			t.Fatalf("expected empty events for empty trace, got %d", len(evs))
		}
	})

	t.Run("maps message turns, tool calls, and compactions with provenance", func(t *testing.T) {
		toolBash := "bash"
		toolRead := "read"
		errMsg := "exit status 1"
		trace := &domain.LiveTraceView{
			SessionID: "ses-123",
			Messages: []domain.LiveMessageFact{
				{
					ID:            "msg-1",
					Role:          "assistant",
					Content:       "Starting build",
					Timestamp:     "2026-10-21T10:00:00Z",
					ContextTokens: 5000,
					Parts: []domain.LivePartFact{
						{
							ID:            "part-1",
							Type:          "tool",
							ToolName:      &toolBash,
							InputSnippet:  "make build",
							OutputSnippet: "make: error",
							Status:        "error",
							IsError:       true,
							ErrorMessage:  &errMsg,
						},
						{
							ID:            "part-2",
							Type:          "tool",
							ToolName:      &toolRead,
							InputSnippet:  "Makefile",
							OutputSnippet: "all: ...",
							Status:        "completed",
							IsError:       false,
						},
					},
				},
				{
					ID:            "msg-2",
					Role:          "assistant",
					Content:       "Compaction turn",
					Timestamp:     "2026-10-21T10:01:00Z",
					ContextTokens: 2500,
					Parts: []domain.LivePartFact{
						{
							ID:   "part-3",
							Type: "compaction",
						},
					},
				},
			},
		}

		events := domain.EventsFromLiveTrace(trace)
		if len(events) != 5 {
			t.Fatalf("expected 5 events (2 turn_messages, 2 tool_calls, 1 compaction), got %d", len(events))
		}

		// Event 0: Turn 1 message
		ev0 := events[0]
		if ev0.EventType != domain.EventTypeTurnMessage || ev0.Turn != 1 || ev0.SessionID != "ses-123" {
			t.Errorf("ev0 mismatch: %+v", ev0)
		}
		if ev0.MessageID == nil || *ev0.MessageID != "msg-1" {
			t.Errorf("expected message_id msg-1, got %v", ev0.MessageID)
		}

		// Event 1: Tool call 1 (bash, error)
		ev1 := events[1]
		if ev1.EventType != domain.EventTypeToolCall || ev1.Turn != 1 || ev1.Status != "error" {
			t.Errorf("ev1 mismatch: %+v", ev1)
		}
		if ev1.ToolName == nil || *ev1.ToolName != "bash" {
			t.Errorf("expected tool name bash, got %v", ev1.ToolName)
		}
		if ev1.MessageID == nil || *ev1.MessageID != "msg-1" {
			t.Errorf("expected message_id msg-1, got %v", ev1.MessageID)
		}

		// Event 2: Tool call 2 (read, completed)
		ev2 := events[2]
		if ev2.EventType != domain.EventTypeToolCall || ev2.Turn != 1 || ev2.Status != "completed" {
			t.Errorf("ev2 mismatch: %+v", ev2)
		}

		// Event 3: Turn 2 message
		ev3 := events[3]
		if ev3.EventType != domain.EventTypeTurnMessage || ev3.Turn != 2 {
			t.Errorf("ev3 mismatch: %+v", ev3)
		}

		// Event 4: Compaction
		ev4 := events[4]
		if ev4.EventType != domain.EventTypeCompaction || ev4.Turn != 2 {
			t.Errorf("ev4 mismatch: %+v", ev4)
		}
		if ev4.MessageID == nil || *ev4.MessageID != "msg-2" {
			t.Errorf("expected message_id msg-2, got %v", ev4.MessageID)
		}
	})

	t.Run("adversarial: messages with empty or whitespace IDs, reasoning, and context token mappings", func(t *testing.T) {
		reasoningText := "Planning next step after analyzing code"
		toolEdit := "edit"
		trace := &domain.LiveTraceView{
			SessionID: "ses-adversarial",
			Messages: []domain.LiveMessageFact{
				{
					ID:               "", // Empty ID
					Role:             "assistant",
					Content:          "   ", // Whitespace content
					Timestamp:        "2026-10-21T11:00:00Z",
					ReasoningContent: &reasoningText,
					ContextTokens:    12345,
					Parts: []domain.LivePartFact{
						{
							ID:            "p-edit",
							Type:          "tool",
							ToolName:      &toolEdit,
							InputSnippet:  "   ", // Whitespace snippet
							OutputSnippet: "",    // Empty snippet
							Status:        "completed",
							IsError:       false,
						},
						{
							ID:   "p-unknown",
							Type: "non_existent_type", // Unrecognized type should be safely skipped
						},
					},
				},
				{
					ID:            "   ", // Whitespace ID
					Role:          "user",
					Content:       "Please continue",
					Timestamp:     "2026-10-21T11:01:00Z",
					ContextTokens: 14000,
					Parts: []domain.LivePartFact{
						{
							ID:   "p-compact",
							Type: "compaction",
						},
					},
				},
			},
		}

		events := domain.EventsFromLiveTrace(trace)
		// Should have 3 events: turn 1 message, tool call (unknown skipped), turn 2 message, compaction
		if len(events) != 4 {
			t.Fatalf("expected 4 events, got %d", len(events))
		}

		// Turn 1 message
		msg0 := events[0]
		if msg0.MessageID != nil {
			t.Errorf("expected nil MessageID for empty string ID, got %v", *msg0.MessageID)
		}
		if msg0.PromptContent != nil {
			t.Errorf("expected nil PromptContent for whitespace content, got %v", *msg0.PromptContent)
		}
		if msg0.ReasoningContent == nil || *msg0.ReasoningContent != reasoningText {
			t.Errorf("expected reasoning content %q, got %v", reasoningText, msg0.ReasoningContent)
		}
		if msg0.ContextTokens != 12345 {
			t.Errorf("expected 12345 context tokens, got %d", msg0.ContextTokens)
		}

		// Turn 1 tool call
		tool0 := events[1]
		if tool0.MessageID != nil {
			t.Errorf("expected nil MessageID on tool call from empty ID message, got %v", *tool0.MessageID)
		}
		if tool0.ToolInput != nil {
			t.Errorf("expected nil ToolInput for whitespace snippet, got %v", *tool0.ToolInput)
		}
		if tool0.ToolOutput != nil {
			t.Errorf("expected nil ToolOutput for empty snippet, got %v", *tool0.ToolOutput)
		}
		if tool0.ContextTokens != 12345 {
			t.Errorf("expected 12345 context tokens propagated to tool call, got %d", tool0.ContextTokens)
		}

		// Turn 2 message
		msg1 := events[2]
		if msg1.MessageID != nil {
			t.Errorf("expected nil MessageID for whitespace ID, got %v", *msg1.MessageID)
		}
		if msg1.PromptContent == nil || *msg1.PromptContent != "Please continue" {
			t.Errorf("expected PromptContent 'Please continue', got %v", msg1.PromptContent)
		}
		if msg1.ContextTokens != 14000 {
			t.Errorf("expected 14000 context tokens, got %d", msg1.ContextTokens)
		}

		// Turn 2 compaction
		comp0 := events[3]
		if comp0.MessageID != nil {
			t.Errorf("expected nil MessageID on compaction from whitespace ID message, got %v", *comp0.MessageID)
		}
		if comp0.ContextTokens != 14000 {
			t.Errorf("expected 14000 context tokens propagated to compaction, got %d", comp0.ContextTokens)
		}
	})
}

func TestLiveTracesEqualDetectsInPlaceChanges(t *testing.T) {
	base := func() domain.LiveTraceView {
		return domain.LiveTraceView{SessionID: "session", Messages: []domain.LiveMessageFact{{ID: "turn", Content: "before", Parts: []domain.LivePartFact{{ID: "part", Status: "running"}}}}}
	}
	cases := []struct {
		name   string
		change func(*domain.LiveTraceView)
	}{
		{"content", func(trace *domain.LiveTraceView) { trace.Messages[0].Content = "after" }},
		{"tokens", func(trace *domain.LiveTraceView) { trace.Messages[0].ReasoningTokens = 20 }},
		{"reasoning", func(trace *domain.LiveTraceView) { text := "thought"; trace.Messages[0].ReasoningContent = &text }},
		{"tool output", func(trace *domain.LiveTraceView) { trace.Messages[0].Parts[0].OutputSnippet = "result" }},
		{"tool status", func(trace *domain.LiveTraceView) { trace.Messages[0].Parts[0].Status = "completed" }},
		{"child link", func(trace *domain.LiveTraceView) { trace.Messages[0].Parts[0].ChildSessionID = "child" }},
		{"failure count", func(trace *domain.LiveTraceView) { trace.FailedToolCalls = 1 }},
		{"anomaly", func(trace *domain.LiveTraceView) {
			trace.Anomalies = []domain.LiveAnomalyAlert{{SessionID: "session", Message: "alert"}}
		}},
	}
	if !domain.LiveTracesEqual(base(), base()) {
		t.Fatal("equal values should compare equal")
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before, after := base(), base()
			tc.change(&after)
			if domain.LiveTracesEqual(before, after) {
				t.Fatal("same-count source edit was suppressed")
			}
		})
	}
}
