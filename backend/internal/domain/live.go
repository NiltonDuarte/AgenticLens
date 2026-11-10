package domain

import (
	"reflect"
	"sort"
	"strings"
)

// FailureSeverity represents the severity level of a tool failure.
type FailureSeverity string

const (
	SeverityNone   FailureSeverity = "none"
	SeverityLow    FailureSeverity = "low"
	SeverityMedium FailureSeverity = "medium"
	SeverityHigh   FailureSeverity = "high"
)

// LiveSessionNode represents a single session within the hierarchical live session tree.
type LiveSessionNode struct {
	ID               string             `json:"id"`
	ParentID         *string            `json:"parent_id,omitempty"`
	Agent            string             `json:"agent"`
	Role             AgentRole          `json:"role"`
	Model            string             `json:"model"`
	Title            string             `json:"title"`
	TimeCreated      string             `json:"time_created"`
	TimeUpdated      string             `json:"time_updated"`
	Status           string             `json:"status"`
	GuessedTicketID  *string            `json:"guessed_ticket_id,omitempty"`
	TicketConfidence float64            `json:"ticket_confidence"`
	Anomalies        []LiveAnomalyAlert `json:"anomalies"`
	SubSessions      []*LiveSessionNode `json:"sub_sessions"`
}

// LiveSessionTree represents the assembled hierarchical tree of active running sessions.
type LiveSessionTree struct {
	RootSessions        []*LiveSessionNode `json:"root_sessions"`
	LastRefreshedAt     string             `json:"last_refreshed_at"`
	TotalActiveSessions int                `json:"total_active_sessions"`
	ActiveAnomalies     []LiveAnomalyAlert `json:"active_anomalies"`
}

// LivePartFact captures a granular execution part (tool invocation, reasoning block, or compaction).
type LivePartFact struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	ToolName       *string         `json:"tool_name,omitempty"`
	InputSnippet   string          `json:"input_snippet"`
	OutputSnippet  string          `json:"output_snippet"`
	Status         string          `json:"status"`
	IsError        bool            `json:"is_error"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	Severity       FailureSeverity `json:"severity"`
	SkillName      string          `json:"skill_name,omitempty"`
	ChildSessionID string          `json:"child_session_id,omitempty"`
	SubagentType   string          `json:"subagent_type,omitempty"`
}

// LiveMessageFact captures a chat message turn with associated execution parts.
type LiveMessageFact struct {
	ID               string         `json:"id"`
	Role             AgentRole      `json:"role"`
	Content          string         `json:"content"`
	Timestamp        string         `json:"timestamp"`
	ReasoningContent *string        `json:"reasoning_content,omitempty"`
	ContextTokens    int64          `json:"context_tokens"`
	InputTokens      int64          `json:"input_tokens"`
	OutputTokens     int64          `json:"output_tokens"`
	CacheReadTokens  int64          `json:"cache_read_tokens"`
	ReasoningTokens  int64          `json:"reasoning_tokens"`
	TotalTokens      int64          `json:"total_tokens"`
	Parts            []LivePartFact `json:"parts"`
}

// LiveRelationshipState describes native child lookup independently of trace content.
type LiveRelationshipState string

const (
	LiveRelationshipsComplete    LiveRelationshipState = "complete"
	LiveRelationshipsUnavailable LiveRelationshipState = "unavailable"
)

type LiveRelationships struct {
	State LiveRelationshipState `json:"state"`
}
type LiveTraceDiagnosticCode string

const DiagnosticChildRelationshipLookupFailed LiveTraceDiagnosticCode = "child_relationship_lookup_failed"

type LiveTraceDiagnostic struct {
	Code    LiveTraceDiagnosticCode `json:"code"`
	Message string                  `json:"message"`
}

// LiveTraceView captures the full chronological execution trace and error summary for a target session.
type LiveTraceView struct {
	Relationships   LiveRelationships     `json:"relationships"`
	Diagnostics     []LiveTraceDiagnostic `json:"diagnostics"`
	SessionID       string                `json:"session_id"`
	Messages        []LiveMessageFact     `json:"messages"`
	TotalToolCalls  int                   `json:"total_tool_calls"`
	FailedToolCalls int                   `json:"failed_tool_calls"`
	RecentErrors    []LivePartFact        `json:"recent_errors"`
	Anomalies       []LiveAnomalyAlert    `json:"anomalies"`
}

// ToolProcessEvidence contains only source-supplied structured process facts.
// A nil field means the source did not report that fact.
type ToolProcessEvidence struct {
	ExitCode *int
	Signal   *string
	TimedOut *bool
}

// ToolExecutionResult retains the source tool status independently of process coverage.
type ToolExecutionResult struct {
	IsError             bool
	Severity            FailureSeverity
	ErrorMessage        *string
	Status              string
	ProcessOutcomeKnown bool
}

// ClassifyToolExecution never interprets free-form text as execution evidence.
// Output is retained only as diagnostic content after structured failure is established.
func ClassifyToolExecution(toolName, status, output string, consecutiveErrors int, process ...ToolProcessEvidence) ToolExecutionResult {
	cleanStatus := strings.ToLower(strings.TrimSpace(status))
	if cleanStatus == "" {
		cleanStatus = "unknown"
	}
	if cleanStatus == "success" || cleanStatus == "succeeded" {
		cleanStatus = "completed"
	}
	result := ToolExecutionResult{Status: cleanStatus, Severity: SeverityNone}
	isError := cleanStatus == "error" || cleanStatus == "failed"
	highSeverity := false
	if len(process) > 0 {
		facts := process[0]
		if facts.ExitCode != nil {
			result.ProcessOutcomeKnown = true
			noMatch := (strings.EqualFold(toolName, "grep") || strings.EqualFold(toolName, "glob")) && *facts.ExitCode == 1
			if *facts.ExitCode != 0 && !noMatch {
				isError = true
			}
		}
		if facts.Signal != nil && strings.TrimSpace(*facts.Signal) != "" {
			isError = true
			highSeverity = true
			result.ProcessOutcomeKnown = true
		}
		if facts.TimedOut != nil && *facts.TimedOut {
			isError = true
			highSeverity = true
			result.ProcessOutcomeKnown = true
		}
	}
	if !isError {
		return result
	}
	result.IsError = true
	result.Status = "error"
	result.Severity = SeverityMedium
	if highSeverity || consecutiveErrors >= 2 {
		result.Severity = SeverityHigh
	}
	msg := output
	if msg == "" {
		msg = cleanStatus
	}
	if len(msg) > 2048 {
		msg = msg[:2048]
	}
	result.ErrorMessage = &msg
	return result
}

// InferTicketID inspects the session title exclusively, matching candidate ticket identifiers
// case-insensitively against configured prefixes, returning canonical uppercase format on match
// or nil with 0.0 confidence when absent.
func InferTicketID(sessionTitle string, cfg TicketConfig) (*string, float64) {
	trimmed := strings.TrimSpace(sessionTitle)
	if trimmed == "" {
		return nil, 0.0
	}

	canonical, ok := NormalizeTicketID(trimmed, cfg)
	if !ok || canonical == "" {
		return nil, 0.0
	}

	return &canonical, 1.0
}

// BuildSessionTree constructs a hierarchical session tree from flat session records,
// assigning root sessions and nesting sub-sessions recursively.
// Roots are sorted by TimeUpdated descending; child sessions are sorted by TimeCreated ascending.
func BuildSessionTree(nodes []*LiveSessionNode) []*LiveSessionNode {
	if len(nodes) == 0 {
		return []*LiveSessionNode{}
	}

	// Create shallow copies of nodes with initialized empty sub-sessions slice
	nodeMap := make(map[string]*LiveSessionNode, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		clone := *n
		clone.SubSessions = make([]*LiveSessionNode, 0)
		if clone.Anomalies == nil {
			clone.Anomalies = make([]LiveAnomalyAlert, 0)
		}
		nodeMap[n.ID] = &clone
	}

	var roots []*LiveSessionNode

	for _, n := range nodes {
		if n == nil {
			continue
		}
		curr := nodeMap[n.ID]
		if curr == nil {
			continue
		}

		if curr.ParentID == nil || strings.TrimSpace(*curr.ParentID) == "" {
			roots = append(roots, curr)
			continue
		}

		if *curr.ParentID == curr.ID {
			// Self-referential parent: promote to root
			roots = append(roots, curr)
			continue
		}

		parent, exists := nodeMap[*curr.ParentID]
		if !exists {
			// Parent is not in the set: treat orphaned node as a root
			roots = append(roots, curr)
			continue
		}

		parent.SubSessions = append(parent.SubSessions, curr)
	}

	// Sort roots by TimeUpdated descending, breaking ties by TimeCreated descending and ID
	sort.SliceStable(roots, func(i, j int) bool {
		if roots[i].TimeUpdated != roots[j].TimeUpdated {
			return roots[i].TimeUpdated > roots[j].TimeUpdated
		}
		if roots[i].TimeCreated != roots[j].TimeCreated {
			return roots[i].TimeCreated > roots[j].TimeCreated
		}
		return roots[i].ID > roots[j].ID
	})

	// Recursively sort sub-sessions by TimeCreated ascending, breaking ties by ID
	visited := make(map[string]bool)
	var sortChildren func(node *LiveSessionNode)
	sortChildren = func(node *LiveSessionNode) {
		if node == nil || visited[node.ID] {
			return
		}
		visited[node.ID] = true
		if len(node.SubSessions) > 1 {
			sort.SliceStable(node.SubSessions, func(i, j int) bool {
				if node.SubSessions[i].TimeCreated != node.SubSessions[j].TimeCreated {
					return node.SubSessions[i].TimeCreated < node.SubSessions[j].TimeCreated
				}
				return node.SubSessions[i].ID < node.SubSessions[j].ID
			})
		}
		for _, child := range node.SubSessions {
			sortChildren(child)
		}
	}

	for _, root := range roots {
		sortChildren(root)
	}

	return roots
}

// LiveTracesEqual compares the complete source-derived snapshot, including in-place edits.
func LiveTracesEqual(previous, next LiveTraceView) bool {
	return reflect.DeepEqual(previous, next)
}
