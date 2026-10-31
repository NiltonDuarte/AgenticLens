// Package codex contains experimental readers for Codex-owned telemetry.
package codex

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

const maxRolloutLineBytes = 8 * 1024 * 1024

// Transcript is a privacy-minimized summary of a Codex rollout. Content-bearing
// fields such as prompts, assistant text, command arguments, and tool results
// are intentionally not retained.
type Transcript struct {
	ThreadID       string                  `json:"thread_id,omitempty"`
	SessionID      string                  `json:"session_id,omitempty"`
	WorkingDir     string                  `json:"-"`
	Source         string                  `json:"source,omitempty"`
	CodexVersion   string                  `json:"codex_version,omitempty"`
	StartedAt      string                  `json:"started_at,omitempty"`
	Provider       string                  `json:"provider,omitempty"`
	Model          string                  `json:"model,omitempty"`
	Usage          Usage                   `json:"usage"`
	ContextWindow  int64                   `json:"context_window,omitempty"`
	RecordKinds    map[string]int64        `json:"record_kinds"`
	EventKinds     map[string]int64        `json:"event_kinds"`
	Warnings       []string                `json:"warnings,omitempty"`
	SteeringTurns  int64                   `json:"steering_turns"`
	Turns          int64                   `json:"turns"`
	ToolCalls      int64                   `json:"tool_calls"`
	ToolErrors     int64                   `json:"tool_errors"`
	ActiveSeconds  int64                   `json:"active_seconds"`
	Invocations    []domain.InvocationFact `json:"-"`
	Tools          []domain.ToolFact       `json:"-"`
	Events         []domain.EventFact      `json:"-"`
	CompletedAt    string                  `json:"-"`
	UpdatedAt      string                  `json:"updated_at"`
	modelUsage     map[string]*modelUsage  `json:"-"`
	toolUsage      map[string]*toolUsage   `json:"-"`
	pendingTools   map[string]int          `json:"-"`
	turnStarts     map[string]time.Time    `json:"-"`
	turnModels     map[string]string       `json:"-"`
	turnNumbers    map[string]int64        `json:"-"`
	toolIDs        map[string]struct{}     `json:"-"`
	compactions    int64                   `json:"-"`
	currentModel   string                  `json:"-"`
	currentContext int64                   `json:"-"`
	role           domain.AgentRole        `json:"-"`
	ParentThread   string                  `json:"parent_thread,omitempty"`
	sessions       []domain.SessionFact    `json:"-"`
	latestTime     string                  `json:"-"`
}

// Usage contains Codex-reported token counts. Cost is deliberately omitted:
// the rollout does not establish whether usage is subscription or API billed.
type Usage struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CacheRead   int64 `json:"cache_read"`
	CacheWrite  int64 `json:"cache_write"`
	Reasoning   int64 `json:"reasoning"`
	PeakContext int64 `json:"-"`
}

// ToInvocationTokens maps the Codex usage dimensions that AgenticLens already
// represents. Cache-write tokens and context-window capacity remain on the
// Codex summary because the current common contract has no equivalent fields.
func (usage Usage) ToInvocationTokens() domain.InvocationTokens {
	uncachedInput := usage.Input - usage.CacheRead
	if uncachedInput < 0 {
		uncachedInput = 0
	}
	return domain.InvocationTokens{
		Input:       uncachedInput,
		Output:      usage.Output,
		CacheRead:   usage.CacheRead,
		Reasoning:   usage.Reasoning,
		PeakContext: usage.PeakContext,
	}
}

type rolloutRecord struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp string          `json:"timestamp"`
}

type sessionMetadata struct {
	ID         string          `json:"id"`
	SessionID  string          `json:"session_id"`
	CWD        string          `json:"cwd"`
	Timestamp  string          `json:"timestamp"`
	Source     json.RawMessage `json:"source"`
	SourceName string          `json:"-"`
	CLIVersion string          `json:"cli_version"`
	Provider   string          `json:"model_provider"`
	ParentID   string          `json:"parent_thread_id"`
	AgentRole  string          `json:"agent_role"`
	AgentPath  string          `json:"-"`
}

// decodeSessionMetadata accepts both interactive source labels and structured
// subagent provenance without retaining the content-bearing metadata fields.
func decodeSessionMetadata(payload json.RawMessage) (sessionMetadata, error) {
	var metadata sessionMetadata
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return metadata, err
	}
	if err := json.Unmarshal(metadata.Source, &metadata.SourceName); err != nil {
		var source struct {
			Subagent struct {
				ThreadSpawn struct {
					ParentID  string `json:"parent_thread_id"`
					AgentRole string `json:"agent_role"`
					AgentPath string `json:"agent_path"`
				} `json:"thread_spawn"`
			} `json:"subagent"`
		}
		if err := json.Unmarshal(metadata.Source, &source); err == nil && source.Subagent.ThreadSpawn.ParentID != "" {
			metadata.SourceName = "subagent"
			if metadata.ParentID == "" {
				metadata.ParentID = source.Subagent.ThreadSpawn.ParentID
			}
			if metadata.AgentRole == "" {
				metadata.AgentRole = source.Subagent.ThreadSpawn.AgentRole
			}
			metadata.AgentPath = source.Subagent.ThreadSpawn.AgentPath
		}
	}
	return metadata, nil
}

type tokenUsage struct {
	Input       int64 `json:"input_tokens"`
	Output      int64 `json:"output_tokens"`
	CachedInput int64 `json:"cached_input_tokens"`
	CacheWrite  int64 `json:"cache_write_input_tokens"`
	Reasoning   int64 `json:"reasoning_output_tokens"`
}

type tokenUsageRecord struct {
	Usage  tokenUsage `json:"usage"`
	TurnID string     `json:"turn_id"`
	Model  string     `json:"model"`
}

type modelUsage struct {
	usage        Usage
	turns        map[string]struct{}
	activeMillis int64
	startedAt    string
	completedAt  string
	peakContext  int64
}

type toolUsage struct {
	calls  int64
	errors int64
}

type tokenUsageInfo struct {
	Total              tokenUsage `json:"total_token_usage"`
	Last               tokenUsage `json:"last_token_usage"`
	ModelContextWindow *int64     `json:"model_context_window"`
}

type tokenCountEvent struct {
	Info *tokenUsageInfo `json:"info"`
}

type eventPayload struct {
	Type string `json:"type"`
}

type turnContext struct {
	Model  string `json:"model"`
	TurnID string `json:"turn_id"`
}

// ParseRollout reads Codex rollout JSONL without retaining transcript content.
// Unknown record types are counted for forward-compatibility diagnostics.
// A malformed final unterminated line is treated as a possibly partial write;
// malformed complete lines are recorded as warnings and parsing continues.
func ParseRollout(reader io.Reader) (Transcript, error) {
	if reader == nil {
		return Transcript{}, errors.New("rollout reader is nil")
	}

	transcript := Transcript{
		RecordKinds:  make(map[string]int64),
		EventKinds:   make(map[string]int64),
		modelUsage:   make(map[string]*modelUsage),
		toolUsage:    make(map[string]*toolUsage),
		pendingTools: make(map[string]int),
		turnStarts:   make(map[string]time.Time),
		turnModels:   make(map[string]string),
		turnNumbers:  make(map[string]int64),
		toolIDs:      make(map[string]struct{}),
		currentModel: "default",
		role:         domain.RoleRootAgent,
	}
	usageRecordSeen := false
	inheritedHistory := false
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), maxRolloutLineBytes)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record rolloutRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			transcript.Warnings = append(transcript.Warnings, fmt.Sprintf("line %d: malformed JSON record", lineNumber))
			continue
		}
		if record.Type == "" {
			return Transcript{}, fmt.Errorf("line %d: record has no type", lineNumber)
		}
		// Forked rollouts start with their own identity and then replay parent
		// history, including parent metadata and rewritten timestamps. Ignore
		// that replay until the thread-settings event identifies this child thread.
		if record.Type == "session_meta" && transcript.ThreadID != "" {
			metadata, err := decodeSessionMetadata(record.Payload)
			if err == nil && metadata.ID != "" && metadata.ID != transcript.ThreadID && transcript.ParentThread != "" {
				inheritedHistory = true
			}
			continue
		}
		if inheritedHistory {
			var owner struct {
				Type     string `json:"type"`
				ThreadID string `json:"thread_id"`
			}
			if record.Type != "event_msg" || json.Unmarshal(record.Payload, &owner) != nil || owner.Type != "thread_settings_applied" || owner.ThreadID != transcript.ThreadID {
				continue
			}
			inheritedHistory = false
		}
		if record.Timestamp != "" {
			transcript.latestTime = record.Timestamp
			if transcript.StartedAt == "" || record.Timestamp < transcript.StartedAt {
				transcript.StartedAt = record.Timestamp
			}
		}
		transcript.RecordKinds[record.Type]++
		switch record.Type {
		case "session_meta":
			if err := readSessionMetadata(record.Payload, &transcript); err != nil {
				transcript.Warnings = append(transcript.Warnings, fmt.Sprintf("line %d: invalid session metadata", lineNumber))
			}
		case "token_usage_record":
			var usage tokenUsageRecord
			if err := json.Unmarshal(record.Payload, &usage); err != nil {
				transcript.Warnings = append(transcript.Warnings, fmt.Sprintf("line %d: invalid token usage record", lineNumber))
				continue
			}
			if !usageRecordSeen {
				transcript.Usage = Usage{}
				usageRecordSeen = true
			}
			addUsage(&transcript.Usage, usage.Usage)
			// Context is the last request's input, including cache reads; the
			// running usage sum is lifetime consumption, not a prompt size.
			transcript.currentContext = usage.Usage.Input
			model := usage.Model
			if model == "" {
				model = transcript.currentModel
			}
			modelAcc := transcript.modelUsageFor(model)
			addUsage(&modelAcc.usage, usage.Usage)
			modelAcc.peakContext = max(modelAcc.peakContext, usage.Usage.Input)
			if usage.TurnID != "" {
				modelAcc.turns[usage.TurnID] = struct{}{}
			}
			if record.Timestamp != "" {
				if modelAcc.startedAt == "" || record.Timestamp < modelAcc.startedAt {
					modelAcc.startedAt = record.Timestamp
				}
				if record.Timestamp > modelAcc.completedAt {
					modelAcc.completedAt = record.Timestamp
				}
			}
		case "event_msg":
			readEvent(record.Payload, record.Timestamp, &transcript, lineNumber, usageRecordSeen)
		case "response_item":
			transcript.readResponseItem(record.Payload, record.Timestamp, rolloutEventID(lineNumber))
		case "turn_context":
			var context turnContext
			if err := json.Unmarshal(record.Payload, &context); err == nil && context.Model != "" {
				transcript.assignTurnModel(context.TurnID, context.Model)
				transcript.Model = context.Model
				transcript.currentModel = context.Model
			}
		case "compacted":
			transcript.addCompactionEvent(record.Timestamp, rolloutEventID(lineNumber))
		case "world_state", "inter_agent_communication_metadata":
			// These Codex records contain harness state/provenance, not session
			// metrics. Keep their types explicit so future record types fail closed.
		default:
			return Transcript{}, fmt.Errorf("line %d: unsupported Codex rollout record type %q", lineNumber, record.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return transcript, fmt.Errorf("reading rollout JSONL: %w", err)
	}
	if transcript.RecordKinds == nil {
		transcript.RecordKinds = map[string]int64{}
	}
	if transcript.EventKinds == nil {
		transcript.EventKinds = map[string]int64{}
	}
	if inheritedHistory {
		return Transcript{}, errors.New("child rollout has inherited history without an owning-thread resume marker; refusing incomplete metrics")
	}
	// Event messages and response items can represent the same user input.
	// Prefer the semantic user-message stream when present.
	if count := transcript.EventKinds["user_message"]; count > 0 {
		transcript.SteeringTurns = count
	}
	transcript.finalizeFacts()
	return transcript, nil
}

func readSessionMetadata(payload json.RawMessage, transcript *Transcript) error {
	metadata, err := decodeSessionMetadata(payload)
	if err != nil {
		return err
	}
	transcript.ThreadID = metadata.ID
	transcript.SessionID = metadata.SessionID
	transcript.WorkingDir = metadata.CWD
	transcript.StartedAt = metadata.Timestamp
	transcript.Source = metadata.SourceName
	transcript.CodexVersion = metadata.CLIVersion
	transcript.Provider = metadata.Provider
	transcript.ParentThread = metadata.ParentID
	if metadata.ParentID != "" {
		transcript.role = domain.RoleUnassignedSubAgent
	}
	if strings.TrimSpace(metadata.AgentRole) != "" {
		transcript.role = domain.AgentRole(strings.TrimSpace(metadata.AgentRole))
	} else if metadata.AgentPath != "" && metadata.ParentID != "" {
		// Codex exposes the raw persona in agent_path even when agent_role is
		// unset. Store only the basename, matching OpenCode's raw session.agent.
		transcript.role = domain.AgentRole(filepath.Base(filepath.ToSlash(metadata.AgentPath)))
	}
	return nil
}

func readEvent(payload json.RawMessage, timestamp string, transcript *Transcript, lineNumber int, usageRecordSeen bool) {
	eventID := rolloutEventID(lineNumber)
	var event eventPayload
	if err := json.Unmarshal(payload, &event); err != nil || event.Type == "" {
		transcript.Warnings = append(transcript.Warnings, fmt.Sprintf("line %d: invalid event message", lineNumber))
		return
	}
	transcript.EventKinds[event.Type]++
	switch event.Type {
	case "turn_started", "task_started":
		var lifecycle struct {
			TurnID string `json:"turn_id"`
		}
		_ = json.Unmarshal(payload, &lifecycle)
		turnID := lifecycle.TurnID
		if turnID == "" {
			turnID = fmt.Sprintf("turn-%d", transcript.Turns+1)
		}
		if _, exists := transcript.turnStarts[turnID]; !exists {
			transcript.Turns++
			transcript.turnModels[turnID] = transcript.currentModel
			transcript.turnNumbers[turnID] = transcript.Turns
			if started, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
				transcript.turnStarts[turnID] = started
				modelAcc := transcript.modelUsageFor(transcript.currentModel)
				modelAcc.turns[turnID] = struct{}{}
				if modelAcc.startedAt == "" || timestamp < modelAcc.startedAt {
					modelAcc.startedAt = timestamp
				}
			}
		}
	case "turn_complete", "turn_completed", "task_complete", "task_completed", "turn_failed", "task_failed":
		var lifecycle struct {
			TurnID string `json:"turn_id"`
		}
		_ = json.Unmarshal(payload, &lifecycle)
		turnID := lifecycle.TurnID
		if turnID == "" && len(transcript.turnStarts) > 0 {
			for id := range transcript.turnStarts {
				turnID = id
			}
		}
		if started, exists := transcript.turnStarts[turnID]; exists {
			if ended, err := time.Parse(time.RFC3339Nano, timestamp); err == nil && !ended.Before(started) {
				millis := ended.Sub(started).Milliseconds()
				transcript.ActiveSeconds += int64(math.Round(float64(millis) / 1000))
				modelAcc := transcript.modelUsageFor(transcript.turnModels[turnID])
				modelAcc.activeMillis += millis
				modelAcc.completedAt = timestamp
			}
			delete(transcript.turnStarts, turnID)
		}
	case "shutdown_complete", "thread_ended":
		transcript.CompletedAt = timestamp
	case "exec_command_begin", "exec_command_end", "mcp_tool_call_begin", "mcp_tool_call_end", "collab_tool_call_begin", "collab_tool_call_end", "web_search_begin", "web_search_end":
		transcript.readToolEvent(event.Type, payload, timestamp, eventID)
	case "context_compacted", "compaction_started", "compaction_completed":
		transcript.addCompactionEvent(timestamp, eventID)
	case "item_completed":
		transcript.readCompletedItem(payload, timestamp, eventID)
	}
	if event.Type != "token_count" {
		return
	}
	var tokenCount tokenCountEvent
	if err := json.Unmarshal(payload, &tokenCount); err != nil || tokenCount.Info == nil {
		return
	}
	info := tokenCount.Info
	if info.Last.Input > 0 {
		transcript.currentContext = info.Last.Input
	}
	if info.ModelContextWindow != nil && *info.ModelContextWindow > transcript.ContextWindow {
		transcript.ContextWindow = *info.ModelContextWindow
	}
	// TokenCount is cumulative and can be repeated. Use it only as a fallback
	// when no per-response usage records were persisted.
	if !usageRecordSeen {
		transcript.Usage = Usage{
			Input:      info.Total.Input,
			Output:     info.Total.Output,
			CacheRead:  info.Total.CachedInput,
			CacheWrite: info.Total.CacheWrite,
			Reasoning:  info.Total.Reasoning,
		}
		modelAcc := transcript.modelUsageFor(transcript.currentModel)
		modelAcc.usage = transcript.Usage
		modelAcc.peakContext = max(modelAcc.peakContext, info.Last.Input)
		transcript.Usage.PeakContext = max(transcript.Usage.PeakContext, info.Last.Input)
	}
}

func (transcript *Transcript) modelUsageFor(model string) *modelUsage {
	cleanModel := strings.TrimSpace(model)
	if cleanModel == "" {
		cleanModel = "default"
	}
	if transcript.modelUsage[cleanModel] == nil {
		transcript.modelUsage[cleanModel] = &modelUsage{turns: make(map[string]struct{})}
	}
	return transcript.modelUsage[cleanModel]
}

func (transcript *Transcript) assignTurnModel(turnID, model string) {
	if turnID == "" || transcript.turnModels[turnID] != "default" {
		return
	}
	transcript.turnModels[turnID] = model
	previous := transcript.modelUsage["default"]
	next := transcript.modelUsageFor(model)
	if previous == nil {
		return
	}
	delete(previous.turns, turnID)
	next.turns[turnID] = struct{}{}
	if previous.startedAt != "" && (next.startedAt == "" || previous.startedAt < next.startedAt) {
		next.startedAt = previous.startedAt
	}
	if len(previous.turns) == 0 && previous.usage == (Usage{}) && previous.activeMillis == 0 {
		delete(transcript.modelUsage, "default")
	}
}

func (transcript *Transcript) readToolEvent(kind string, payload json.RawMessage, timestamp, sourceID string) {
	var raw struct {
		CallID           string          `json:"call_id"`
		TurnID           string          `json:"turn_id"`
		StartedAtMS      int64           `json:"started_at_ms"`
		DurationMS       int64           `json:"duration_ms"`
		ExitCode         *int            `json:"exit_code"`
		Status           string          `json:"status"`
		Output           json.RawMessage `json:"output"`
		AggregatedOutput string          `json:"aggregated_output"`
		Invocation       struct {
			Server string `json:"server"`
			Tool   string `json:"tool"`
		} `json:"invocation"`
		ToolName string `json:"tool_name"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return
	}
	name := "tool"
	switch {
	case strings.HasPrefix(kind, "exec_command_"):
		name = "command"
	case strings.HasPrefix(kind, "mcp_tool_call_"):
		name = strings.TrimSpace(raw.Invocation.Tool)
		if name == "" {
			name = "mcp"
		}
	case strings.HasPrefix(kind, "collab_tool_call_"):
		name = strings.TrimSpace(raw.ToolName)
		if name == "" {
			name = "collaboration"
		}
	case strings.HasPrefix(kind, "web_search_"):
		name = "web_search"
	}
	callKey := raw.CallID
	if callKey == "" {
		callKey = fmt.Sprintf("%s:%s:%d", kind, timestamp, len(transcript.Events))
	}
	eventID := sourceID
	if raw.CallID != "" {
		eventID = "codex-call:" + raw.CallID
	}
	started := strings.HasSuffix(kind, "_begin")
	if started {
		if _, exists := transcript.toolIDs[callKey]; exists {
			return
		}
		transcript.toolIDs[callKey] = struct{}{}
		tool := transcript.toolUsage[name]
		if tool == nil {
			tool = &toolUsage{}
			transcript.toolUsage[name] = tool
		}
		tool.calls++
		transcript.ToolCalls++
		transcript.pendingTools[callKey] = len(transcript.Events)
		toolName := name
		transcript.Events = append(transcript.Events, domain.EventFact{
			EventID: eventID, Timestamp: timestamp, SessionID: transcript.sessionIDOrThread(), Turn: transcript.turnNumber(raw.TurnID),
			Role: transcript.role, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: "running",
			ContextTokens: transcript.currentContext,
		})
		return
	}
	status := strings.ToLower(strings.TrimSpace(raw.Status))
	if status == "" && raw.ExitCode != nil {
		if *raw.ExitCode == 0 {
			status = "completed"
		} else {
			status = "error"
		}
	}
	if status == "" {
		status = "completed"
	}
	isError := status == "error" || status == "failed" || (raw.ExitCode != nil && *raw.ExitCode != 0)
	if isError {
		status = "error"
	} else if status != "running" {
		status = "completed"
	}
	if index, exists := transcript.pendingTools[callKey]; exists {
		transcript.Events[index].Status = status
		if raw.DurationMS > 0 {
			duration := raw.DurationMS
			transcript.Events[index].DurationMs = &duration
		} else if raw.StartedAtMS > 0 {
			if ended, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
				startedAt := time.UnixMilli(raw.StartedAtMS)
				duration := ended.Sub(startedAt).Milliseconds()
				if duration >= 0 {
					transcript.Events[index].DurationMs = &duration
				}
			}
		}
		delete(transcript.pendingTools, callKey)
	} else {
		if _, exists := transcript.toolIDs[callKey]; exists {
			return
		}
		transcript.toolIDs[callKey] = struct{}{}
		tool := transcript.toolUsage[name]
		if tool == nil {
			tool = &toolUsage{}
			transcript.toolUsage[name] = tool
		}
		tool.calls++
		transcript.ToolCalls++
		toolName := name
		transcript.Events = append(transcript.Events, domain.EventFact{
			EventID: eventID, Timestamp: timestamp, SessionID: transcript.sessionIDOrThread(), Turn: transcript.turnNumber(raw.TurnID),
			Role: transcript.role, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: status,
			ContextTokens: transcript.currentContext,
		})
	}
	output := codexToolOutput(raw.Output)
	if output == "" {
		output = raw.AggregatedOutput
	}
	for i := len(transcript.Events) - 1; i >= 0; i-- {
		if transcript.Events[i].EventID == eventID {
			transcript.recordToolOutput(i, output)
			break
		}
	}
	if isError {
		transcript.toolUsage[name].errors++
		transcript.ToolErrors++
	}
}

func (transcript *Transcript) readResponseItem(payload json.RawMessage, timestamp, sourceID string) {
	var item struct {
		Type   string          `json:"type"`
		Role   string          `json:"role"`
		CallID string          `json:"call_id"`
		Name   string          `json:"name"`
		Status string          `json:"status"`
		Output json.RawMessage `json:"output"`
	}
	if err := json.Unmarshal(payload, &item); err != nil {
		return
	}
	if item.Type == "message" && item.Role == "user" {
		transcript.SteeringTurns++
	}
	if item.CallID == "" {
		return
	}
	switch item.Type {
	case "custom_tool_call", "function_call":
		name := strings.TrimSpace(item.Name)
		if name == "" || strings.EqualFold(name, "exec") {
			// Codex exec wrappers are represented more accurately by their
			// completed CommandExecution lifecycle items.
			return
		}
		if _, exists := transcript.toolIDs[item.CallID]; exists {
			return
		}
		transcript.toolIDs[item.CallID] = struct{}{}
		tool := transcript.toolUsage[name]
		if tool == nil {
			tool = &toolUsage{}
			transcript.toolUsage[name] = tool
		}
		tool.calls++
		transcript.ToolCalls++
		transcript.pendingTools[item.CallID] = len(transcript.Events)
		toolName := name
		transcript.Events = append(transcript.Events, domain.EventFact{
			EventID: "codex-call:" + item.CallID, Timestamp: timestamp, SessionID: transcript.sessionIDOrThread(), Turn: transcript.Turns,
			Role: transcript.role, EventType: domain.EventTypeToolCall, ToolName: &toolName,
			Status: "running", ContextTokens: transcript.currentContext,
		})
	case "custom_tool_call_output", "function_call_output":
		if index, exists := transcript.pendingTools[item.CallID]; exists {
			transcript.Events[index].Status = "completed"
			output := codexToolOutput(item.Output)
			transcript.recordToolOutput(index, output)
			if started, err := time.Parse(time.RFC3339Nano, transcript.Events[index].Timestamp); err == nil {
				if ended, err := time.Parse(time.RFC3339Nano, timestamp); err == nil && !ended.Before(started) {
					duration := ended.Sub(started).Milliseconds()
					transcript.Events[index].DurationMs = &duration
				}
			}
			delete(transcript.pendingTools, item.CallID)
		}
	}
}

func (transcript *Transcript) readCompletedItem(payload json.RawMessage, timestamp, sourceID string) {
	var completed struct {
		TurnID      string `json:"turn_id"`
		StartedAtMS int64  `json:"started_at_ms"`
		CompletedMS int64  `json:"completed_at_ms"`
		Item        struct {
			ID               string          `json:"id"`
			Type             string          `json:"type"`
			Status           string          `json:"status"`
			ExitCode         *int            `json:"exit_code"`
			AggregatedOutput string          `json:"aggregated_output"`
			Output           json.RawMessage `json:"output"`
			Duration         struct {
				Seconds int64 `json:"secs"`
				Nanos   int64 `json:"nanos"`
			} `json:"duration"`
		} `json:"item"`
	}
	if err := json.Unmarshal(payload, &completed); err != nil || completed.Item.Type != "CommandExecution" {
		return
	}
	callID := completed.Item.ID
	eventID := sourceID
	if callID != "" {
		eventID = "codex-call:" + callID
		if _, exists := transcript.toolIDs[callID]; exists {
			return
		}
		transcript.toolIDs[callID] = struct{}{}
	}
	name := "command"
	tool := transcript.toolUsage[name]
	if tool == nil {
		tool = &toolUsage{}
		transcript.toolUsage[name] = tool
	}
	tool.calls++
	transcript.ToolCalls++
	status := strings.ToLower(strings.TrimSpace(completed.Item.Status))
	if status == "" {
		status = "completed"
	}
	isError := status == "failed" || status == "error" || (completed.Item.ExitCode != nil && *completed.Item.ExitCode != 0)
	if isError {
		status = "error"
		tool.errors++
		transcript.ToolErrors++
	} else {
		status = "completed"
	}
	startedAt := timestamp
	if completed.StartedAtMS > 0 {
		startedAt = time.UnixMilli(completed.StartedAtMS).UTC().Format(time.RFC3339Nano)
	}
	duration := completed.Item.Duration.Seconds*1000 + completed.Item.Duration.Nanos/1_000_000
	if duration == 0 && completed.StartedAtMS > 0 && completed.CompletedMS >= completed.StartedAtMS {
		duration = completed.CompletedMS - completed.StartedAtMS
	}
	var durationMS *int64
	if duration >= 0 {
		durationMS = &duration
	}
	toolName := name
	transcript.Events = append(transcript.Events, domain.EventFact{
		EventID: eventID, Timestamp: startedAt, SessionID: transcript.sessionIDOrThread(), Turn: transcript.turnNumber(completed.TurnID),
		Role: transcript.role, EventType: domain.EventTypeToolCall, ToolName: &toolName, Status: status,
		DurationMs: durationMS, ContextTokens: transcript.currentContext,
	})
	output := completed.Item.AggregatedOutput
	if output == "" {
		output = codexToolOutput(completed.Item.Output)
	}
	transcript.recordToolOutput(len(transcript.Events)-1, output)
}

func codexToolOutput(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	return string(raw)
}

// recordToolOutput preserves source evidence without interpreting its contents.
func (transcript *Transcript) recordToolOutput(index int, output string) {
	if output != "" {
		transcript.Events[index].ToolOutput = &output
	}
}

func (transcript *Transcript) addCompactionEvent(timestamp, eventID string) {
	transcript.EventKinds["compaction"]++
	transcript.compactions++
	transcript.Events = append(transcript.Events, domain.EventFact{
		EventID: eventID, Timestamp: timestamp, SessionID: transcript.sessionIDOrThread(), Role: transcript.role,
		EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: transcript.currentContext,
	})
}

func rolloutEventID(lineNumber int) string {
	return "codex-rollout-line:" + strconv.Itoa(lineNumber)
}

func (transcript *Transcript) sessionIDOrThread() string {
	if transcript.ThreadID != "" {
		return transcript.ThreadID
	}
	return transcript.SessionID
}

func (transcript *Transcript) turnNumber(turnID string) int64 {
	if model := transcript.turnModels[turnID]; model != "" {
		if usage := transcript.modelUsage[model]; usage != nil {
			return transcript.turnNumbers[turnID]
		}
	}
	return transcript.Turns
}

func (transcript *Transcript) finalizeFacts() {
	if transcript.StartedAt == "" {
		transcript.StartedAt = transcript.latestTime
	}
	// A persisted rollout may still be active; its last record time is an update, not a completion.
	if transcript.Model == "" {
		transcript.Model = transcript.currentModel
	}
	if len(transcript.modelUsage) == 0 {
		transcript.modelUsageFor(transcript.Model)
	}
	modelNames := make([]string, 0, len(transcript.modelUsage))
	for model := range transcript.modelUsage {
		modelNames = append(modelNames, model)
	}
	sort.Strings(modelNames)
	for _, model := range modelNames {
		usage := transcript.modelUsage[model]
		usage.usage.PeakContext = usage.peakContext
		transcript.Usage.PeakContext = max(transcript.Usage.PeakContext, usage.peakContext)
		var startedAt, completedAt *string
		if usage.startedAt != "" {
			started := usage.startedAt
			startedAt = &started
		}
		if usage.completedAt != "" {
			completed := usage.completedAt
			completedAt = &completed
		}
		if model == transcript.Model && startedAt == nil && transcript.StartedAt != "" {
			started := transcript.StartedAt
			startedAt = &started
		}
		if model == transcript.Model && transcript.CompletedAt != "" {
			completed := transcript.CompletedAt
			completedAt = &completed
		}
		turns := int64(len(usage.turns))
		activeSeconds := int64(math.Round(float64(usage.activeMillis) / 1000))
		compactions := int64(0)
		userIterations := int64(0)
		if len(transcript.Invocations) == 0 {
			compactions = transcript.compactions
			if transcript.ParentThread == "" {
				userIterations = transcript.SteeringTurns
			}
		}
		transcript.Invocations = append(transcript.Invocations, domain.InvocationFact{
			Role:                  transcript.role,
			Model:                 model,
			StartedAt:             startedAt,
			CompletedAt:           completedAt,
			ActiveDurationSeconds: activeSeconds,
			Turns:                 turns,
			Tokens:                usage.usage.ToInvocationTokens(),
			CostUSD:               0,
			CostEvidence:          domain.UnavailableCost(),
			Compactions:           compactions,
			UserIterations:        userIterations,
		})
	}
	for name, stats := range transcript.toolUsage {
		transcript.Tools = append(transcript.Tools, domain.ToolFact{
			SessionID: transcript.sessionIDOrThread(), Name: name, Calls: stats.calls, Errors: stats.errors,
		})
	}
	sort.Slice(transcript.Tools, func(i, j int) bool {
		if transcript.Tools[i].SessionID != transcript.Tools[j].SessionID {
			return transcript.Tools[i].SessionID < transcript.Tools[j].SessionID
		}
		return transcript.Tools[i].Name < transcript.Tools[j].Name
	})
	if transcript.UpdatedAt == "" {
		transcript.UpdatedAt = transcript.latestTime
	}
}

// SessionFacts converts the transcript invocations, tools, and events into normalized SessionFacts.
func (transcript Transcript) SessionFacts() ([]domain.SessionFact, error) {
	owned := transcript.sessions
	if owned == nil {
		session := domain.SessionFact{ID: transcript.sessionIDOrThread(), Source: domain.SourceCodex, Invocations: transcript.Invocations}
		if transcript.ParentThread != "" {
			parent := transcript.ParentThread
			session.ParentID = &parent
		}
		owned = []domain.SessionFact{session}
	}
	sessions, err := domain.SessionsFromFactsWithSource(domain.SourceCodex, owned, transcript.Tools, transcript.Events)
	if err != nil {
		return nil, fmt.Errorf("grouping Codex facts by session: %w", err)
	}
	return sessions, nil
}

// CombineTranscripts merges a discovered root thread and its child threads
// while preserving per-thread invocation facts and aggregating tool summaries.
func CombineTranscripts(transcripts []Transcript) (Transcript, error) {
	if len(transcripts) == 0 {
		return Transcript{}, errors.New("cannot combine an empty Codex transcript set")
	}
	combined := Transcript{RecordKinds: make(map[string]int64), EventKinds: make(map[string]int64), sessions: []domain.SessionFact{}}
	for index, transcript := range transcripts {
		sessions, err := transcript.SessionFacts()
		if err != nil {
			return Transcript{}, fmt.Errorf("grouping Codex transcript %d: %w", index, err)
		}
		combined.sessions = append(combined.sessions, sessions...)
		if index == 0 {
			// Human steering belongs to the root; child prompts are delegation.
			combined.SteeringTurns = transcript.SteeringTurns
			combined.ThreadID = transcript.ThreadID
			combined.SessionID = transcript.SessionID
			combined.WorkingDir = transcript.WorkingDir
			combined.Source = transcript.Source
			combined.CodexVersion = transcript.CodexVersion
			combined.Provider = transcript.Provider
			combined.Model = transcript.Model
		}
		combined.Usage.Input += transcript.Usage.Input
		combined.Usage.Output += transcript.Usage.Output
		combined.Usage.CacheRead += transcript.Usage.CacheRead
		combined.Usage.CacheWrite += transcript.Usage.CacheWrite
		combined.Usage.Reasoning += transcript.Usage.Reasoning
		combined.Usage.PeakContext = max(combined.Usage.PeakContext, transcript.Usage.PeakContext)
		combined.ContextWindow = max(combined.ContextWindow, transcript.ContextWindow)
		combined.Turns += transcript.Turns
		combined.ToolCalls += transcript.ToolCalls
		combined.ToolErrors += transcript.ToolErrors
		combined.ActiveSeconds += transcript.ActiveSeconds
		combined.Invocations = append(combined.Invocations, transcript.Invocations...)
		combined.Events = append(combined.Events, transcript.Events...)
		combined.Warnings = append(combined.Warnings, transcript.Warnings...)
		for kind, count := range transcript.RecordKinds {
			combined.RecordKinds[kind] += count
		}
		for kind, count := range transcript.EventKinds {
			combined.EventKinds[kind] += count
		}
		combined.Tools = append(combined.Tools, transcript.Tools...)
		if transcript.StartedAt != "" && (combined.StartedAt == "" || transcript.StartedAt < combined.StartedAt) {
			combined.StartedAt = transcript.StartedAt
		}
		if transcript.CompletedAt > combined.CompletedAt {
			combined.CompletedAt = transcript.CompletedAt
		}
		if transcript.UpdatedAt > combined.UpdatedAt {
			combined.UpdatedAt = transcript.UpdatedAt
		}
	}
	sort.Slice(combined.Tools, func(i, j int) bool {
		if combined.Tools[i].SessionID != combined.Tools[j].SessionID {
			return combined.Tools[i].SessionID < combined.Tools[j].SessionID
		}
		return combined.Tools[i].Name < combined.Tools[j].Name
	})
	sort.SliceStable(combined.Events, func(i, j int) bool { return combined.Events[i].Timestamp < combined.Events[j].Timestamp })
	return combined, nil
}

func addUsage(target *Usage, usage tokenUsage) {
	target.Input += usage.Input
	target.Output += usage.Output
	target.CacheRead += usage.CachedInput
	target.CacheWrite += usage.CacheWrite
	target.Reasoning += usage.Reasoning
}
