package domain

import (
	"reflect"
	"regexp"
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

// LiveTraceView captures the full chronological execution trace and error summary for a target session.
type LiveTraceView struct {
	SessionID       string             `json:"session_id"`
	Messages        []LiveMessageFact  `json:"messages"`
	TotalToolCalls  int                `json:"total_tool_calls"`
	FailedToolCalls int                `json:"failed_tool_calls"`
	RecentErrors    []LivePartFact     `json:"recent_errors"`
	Anomalies       []LiveAnomalyAlert `json:"anomalies"`
}

// ToolExecutionResult is the normalized result of classifying one tool execution.
type ToolExecutionResult struct {
	IsError      bool
	Severity     FailureSeverity
	ErrorMessage *string
}

var exitStatusPattern = regexp.MustCompile(`(?i)\bexit(?:ed)?\s+(?:status|code)\s*[:=]?\s*([0-9]+)\b`)

// ClassifyToolExecution identifies process-level failures without interpreting semantic output.
func ClassifyToolExecution(toolName string, status string, output string, consecutiveErrors int) ToolExecutionResult {
	trimmedOutput := strings.TrimSpace(output)
	lowerOutput := strings.ToLower(trimmedOutput)
	cleanStatus := strings.ToLower(strings.TrimSpace(status))
	cleanTool := strings.ToLower(strings.TrimSpace(toolName))
	exitCode := -1
	if match := exitStatusPattern.FindStringSubmatch(trimmedOutput); len(match) == 2 {
		exitCode = 0
		for _, digit := range match[1] {
			exitCode = exitCode*10 + int(digit-'0')
		}
	}
	noMatch := (cleanTool == "grep" || cleanTool == "glob") && (exitCode == 1 || strings.Contains(lowerOutput, "no matches") || strings.Contains(lowerOutput, "0 matches") || strings.Contains(lowerOutput, "no matching files"))
	if noMatch && cleanStatus != "error" && cleanStatus != "failed" {
		return ToolExecutionResult{Severity: SeverityNone}
	}
	isCrash := strings.Contains(lowerOutput, "segmentation fault") ||
		strings.Contains(lowerOutput, "signal: killed") ||
		strings.Contains(lowerOutput, "killed by signal") ||
		strings.Contains(lowerOutput, "out of memory")
	hardOutputFailure := (exitCode > 0 && !((cleanTool == "grep" || cleanTool == "glob") && exitCode == 1)) || (exitCode != 0 && (isCrash || lowerOutput == "permission denied" || lowerOutput == "command timed out" || lowerOutput == "operation timed out"))
	isError := cleanStatus == "error" || cleanStatus == "failed" || hardOutputFailure
	if !isError {
		return ToolExecutionResult{Severity: SeverityNone}
	}
	severity := SeverityMedium
	if isCrash || consecutiveErrors >= 2 {
		severity = SeverityHigh
	}

	// Preserve source error evidence, bounded to 2048 bytes.
	msg := trimmedOutput
	if msg == "" {
		msg = cleanStatus
		if msg == "" {
			msg = "tool error"
		}
	}
	if len(msg) > 2048 {
		msg = msg[:2048]
	}

	return ToolExecutionResult{IsError: true, Severity: severity, ErrorMessage: &msg}
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
