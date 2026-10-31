package domain

import (
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
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	ToolName      *string         `json:"tool_name,omitempty"`
	InputSnippet  string          `json:"input_snippet"`
	OutputSnippet string          `json:"output_snippet"`
	Status        string          `json:"status"`
	IsError       bool            `json:"is_error"`
	ErrorMessage  *string         `json:"error_message,omitempty"`
	Severity      FailureSeverity `json:"severity"`
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
		strings.Contains(lowerOutput, "out of memory") ||
		strings.Contains(lowerOutput, "panic:") ||
		strings.Contains(lowerOutput, "fatal error:") ||
		strings.Contains(lowerOutput, "unhandledpromiserejection")
	hardOutputFailure := (exitCode > 0 && !((cleanTool == "grep" || cleanTool == "glob") && exitCode == 1)) || isCrash || strings.Contains(lowerOutput, "permission denied") || strings.Contains(lowerOutput, "timed out") || strings.Contains(lowerOutput, "timeout")
	isError := cleanStatus == "error" || cleanStatus == "failed" || hardOutputFailure
	if !isError {
		return ToolExecutionResult{Severity: SeverityNone}
	}
	severity := SeverityMedium
	if isCrash || consecutiveErrors >= 2 {
		severity = SeverityHigh
	}

	// Prepare sanitized error message capped at 2048 characters
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

// ClassifyToolFailure preserves the previous call signature for existing internal consumers.
func ClassifyToolFailure(toolName, status, output string, consecutiveErrors int) (bool, FailureSeverity, *string) {
	result := ClassifyToolExecution(toolName, status, output, consecutiveErrors)
	return result.IsError, result.Severity, result.ErrorMessage
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

// EventsFromLiveTrace is a pure function that converts a LiveTraceView into a slice of EventFact domain events.
// It maps messages to turn_message events, tool parts to tool_call events, and compaction parts to compaction events.
func EventsFromLiveTrace(trace *LiveTraceView) []EventFact {
	if trace == nil || len(trace.Messages) == 0 {
		return []EventFact{}
	}

	events := make([]EventFact, 0)
	consecutiveErrors := 0

	for idx, msg := range trace.Messages {
		turnNum := int64(idx + 1)
		var msgIDPtr *string
		if strings.TrimSpace(msg.ID) != "" {
			trimmedID := strings.TrimSpace(msg.ID)
			msgIDPtr = &trimmedID
		}

		var promptContentPtr *string
		if strings.TrimSpace(msg.Content) != "" {
			trimmedContent := strings.TrimSpace(msg.Content)
			promptContentPtr = &trimmedContent
		}

		// Map turn_message event
		events = append(events, EventFact{
			Timestamp:        msg.Timestamp,
			SessionID:        trace.SessionID,
			MessageID:        msgIDPtr,
			Turn:             turnNum,
			Role:             msg.Role,
			EventType:        EventTypeTurnMessage,
			Status:           "completed",
			ContextTokens:    msg.ContextTokens,
			PromptContent:    promptContentPtr,
			ReasoningContent: msg.ReasoningContent,
		})

		// Map parts
		for _, part := range msg.Parts {
			switch part.Type {
			case "tool":
				status := "completed"
				if part.IsError {
					status = "error"
					consecutiveErrors++
				} else {
					consecutiveErrors = 0
				}

				var toolInputPtr *string
				if strings.TrimSpace(part.InputSnippet) != "" {
					trimmedIn := strings.TrimSpace(part.InputSnippet)
					toolInputPtr = &trimmedIn
				}

				var toolOutputPtr *string
				if strings.TrimSpace(part.OutputSnippet) != "" {
					trimmedOut := strings.TrimSpace(part.OutputSnippet)
					toolOutputPtr = &trimmedOut
				}

				events = append(events, EventFact{
					Timestamp:     msg.Timestamp,
					SessionID:     trace.SessionID,
					MessageID:     msgIDPtr,
					Turn:          turnNum,
					Role:          msg.Role,
					EventType:     EventTypeToolCall,
					ToolName:      part.ToolName,
					Status:        status,
					ErrorMessage:  part.ErrorMessage,
					ContextTokens: msg.ContextTokens,
					ToolInput:     toolInputPtr,
					ToolOutput:    toolOutputPtr,
				})

			case "compaction":
				events = append(events, EventFact{
					Timestamp:     msg.Timestamp,
					SessionID:     trace.SessionID,
					MessageID:     msgIDPtr,
					Turn:          turnNum,
					Role:          msg.Role,
					EventType:     EventTypeCompaction,
					Status:        "completed",
					ContextTokens: msg.ContextTokens,
				})
			}
		}
	}

	return events
}
