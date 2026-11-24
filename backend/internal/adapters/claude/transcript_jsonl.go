package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// syntheticModel marks client-side messages, such as API error notices, that
// no model served. They carry no billable usage and create no invocation row.
const syntheticModel = "<synthetic>"

// Transcript is the request-local projection of one native transcript file.
// It keeps identities, counts and timing only; prompt, tool and result content
// is never retained.
type Transcript struct {
	// OwnerID is the root sessionId or a subagent agentId.
	OwnerID string
	// RootSessionID is the sessionId every non-copied record must carry.
	RootSessionID string
	// AgentID is set for subagent transcripts only.
	AgentID string
	// WorkingDirectory is the first recorded cwd of an owned record.
	WorkingDirectory string
	Version          string
	Entrypoint       string
	// Incomplete lists reasons that prevent replacing this owner's stored facts.
	Incomplete []string

	responses     []response
	responseIndex map[string]int
	forkedKeys    map[string]bool
	toolUses      []toolUse
	toolIndex     map[string]int
	results       map[string]toolResult
	timeline      []timelineEntry
	compactions   []string
	promptCount   int64
	sidechainSeen bool
}

type response struct {
	key           string
	model         string
	timestamp     string
	input         int64
	output        int64
	cacheRead     int64
	cacheCreation int64
	thinking      int64
}

func (r response) context() int64 { return r.input + r.cacheRead + r.cacheCreation }

type toolUse struct {
	id           string
	name         string
	timestamp    string
	responseKey  string
	turn         int64
	skill        string
	subagentType string
	description  string
}

type toolResult struct {
	timestamp string
	isError   bool
	success   *bool
	agentID   string
	agentType string
}

type entryKind int

const (
	entryOwned entryKind = iota
	entryPrompt
	entryAssistant
	// entryResume is a coordinator message that resumes an idle subagent. It
	// starts a turn only when no turn is open.
	entryResume
)

type timelineEntry struct {
	kind        entryKind
	at          time.Time
	raw         string
	nativeID    string
	system      bool
	responseKey string
	model       string
	terminal    bool
}

type rawRecord struct {
	Type              string          `json:"type"`
	Subtype           string          `json:"subtype"`
	SessionID         *string         `json:"sessionId"`
	AgentID           *string         `json:"agentId"`
	UUID              string          `json:"uuid"`
	Timestamp         string          `json:"timestamp"`
	CWD               string          `json:"cwd"`
	Version           string          `json:"version"`
	Entrypoint        string          `json:"entrypoint"`
	IsSidechain       bool            `json:"isSidechain"`
	IsMeta            bool            `json:"isMeta"`
	IsCompactSummary  bool            `json:"isCompactSummary"`
	TranscriptOnly    bool            `json:"isVisibleInTranscriptOnly"`
	PromptSource      string          `json:"promptSource"`
	Origin            json.RawMessage `json:"origin"`
	ForkedFrom        json.RawMessage `json:"forkedFrom"`
	RequestID         string          `json:"requestId"`
	IsAPIErrorMessage bool            `json:"isApiErrorMessage"`
	Message           json.RawMessage `json:"message"`
	ToolUseResult     json.RawMessage `json:"toolUseResult"`
}

type rawUsage struct {
	Input              *int64 `json:"input_tokens"`
	Output             *int64 `json:"output_tokens"`
	CacheRead          *int64 `json:"cache_read_input_tokens"`
	CacheCreation      *int64 `json:"cache_creation_input_tokens"`
	OutputTokenDetails *struct {
		Thinking *int64 `json:"thinking_tokens"`
	} `json:"output_tokens_details"`
}

type rawBlock struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	IsError   *bool           `json:"is_error"`
}

type rawAssistantMessage struct {
	ID         string          `json:"id"`
	Model      string          `json:"model"`
	StopReason *string         `json:"stop_reason"`
	Usage      json.RawMessage `json:"usage"`
	Content    json.RawMessage `json:"content"`
}

type rawUserMessage struct {
	Content json.RawMessage `json:"content"`
}

// TranscriptExpectation names the identity every owned record must carry.
type TranscriptExpectation struct {
	RootSessionID string
	AgentID       string
}

// ParseTranscript reads one native JSONL transcript. Conversation, system and
// attachment records are interpreted; other native record types are ignored.
// Malformed lines, untyped records, foreign identities and unsupported layouts
// make the owner incomplete instead of being recorded as zero. A returned error
// means the source could not be read at all.
func ParseTranscript(ctx context.Context, reader io.Reader, expect TranscriptExpectation) (Transcript, error) {
	if err := ctx.Err(); err != nil {
		return Transcript{}, err
	}
	if reader == nil {
		return Transcript{}, errors.New("Claude Code transcript reader is nil")
	}
	t := Transcript{
		OwnerID:       expect.RootSessionID,
		RootSessionID: expect.RootSessionID,
		AgentID:       expect.AgentID,
		responseIndex: make(map[string]int),
		forkedKeys:    make(map[string]bool),
		toolIndex:     make(map[string]int),
		results:       make(map[string]toolResult),
	}
	if expect.AgentID != "" {
		t.OwnerID = expect.AgentID
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), maxLineBytes)
	lineNumber := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return Transcript{}, err
		}
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record rawRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.incomplete(fmt.Sprintf("line %d: malformed JSON record", lineNumber))
			continue
		}
		if strings.TrimSpace(record.Type) == "" {
			t.incomplete(fmt.Sprintf("line %d: record has no type", lineNumber))
			continue
		}
		t.readRecord(record, lineNumber)
	}
	if err := ctx.Err(); err != nil {
		return Transcript{}, err
	}
	if err := scanner.Err(); err != nil {
		return Transcript{}, fmt.Errorf("reading Claude Code transcript JSONL: %w", err)
	}
	t.dropInheritedResponses()
	return t, nil
}

func (t *Transcript) incomplete(reason string) {
	for _, existing := range t.Incomplete {
		if existing == reason {
			return
		}
	}
	t.Incomplete = append(t.Incomplete, reason)
}

func isConversation(kind string) bool {
	return kind == "user" || kind == "assistant" || kind == "system" || kind == "attachment"
}

func (t *Transcript) readRecord(record rawRecord, line int) {
	forked := len(record.ForkedFrom) > 0 && string(record.ForkedFrom) != "null"
	if forked {
		// Copied history belongs to its original session. Remember its response
		// identities so later copies cannot be counted for this session.
		if record.Type == "assistant" {
			var message rawAssistantMessage
			if json.Unmarshal(record.Message, &message) == nil {
				if key := responseKey(record.RequestID, message.ID); key != "" {
					t.forkedKeys[key] = true
				}
			}
		}
		return
	}
	if record.SessionID != nil && *record.SessionID != t.RootSessionID {
		t.incomplete(fmt.Sprintf("line %d: record belongs to a different sessionId", line))
		return
	}
	if !isConversation(record.Type) {
		return
	}
	if t.AgentID != "" {
		if record.AgentID == nil || *record.AgentID != t.AgentID {
			t.incomplete(fmt.Sprintf("line %d: record does not carry the subagent agentId", line))
			return
		}
	} else if record.IsSidechain && (record.Type == "user" || record.Type == "assistant") {
		if !t.sidechainSeen {
			t.sidechainSeen = true
			t.incomplete("unsupported layout: root transcript contains sidechain conversation records")
		}
		return
	}
	if t.WorkingDirectory == "" && strings.TrimSpace(record.CWD) != "" {
		t.WorkingDirectory = record.CWD
	}
	if t.Version == "" {
		t.Version = record.Version
	}
	if t.Entrypoint == "" {
		t.Entrypoint = record.Entrypoint
	}
	at, timeErr := time.Parse(time.RFC3339Nano, record.Timestamp)
	switch record.Type {
	case "assistant":
		if timeErr != nil {
			t.incomplete(fmt.Sprintf("line %d: assistant record has no valid timestamp", line))
			return
		}
		t.readAssistant(record, at, line)
	case "user":
		if timeErr != nil {
			t.incomplete(fmt.Sprintf("line %d: user record has no valid timestamp", line))
			return
		}
		t.readUser(record, at, line)
	case "system":
		if record.Subtype == "compact_boundary" {
			t.compactions = append(t.compactions, record.Timestamp)
		}
		if timeErr == nil {
			t.timeline = append(t.timeline, timelineEntry{kind: entryOwned, at: at, raw: record.Timestamp})
		}
	case "attachment":
		if timeErr == nil {
			t.timeline = append(t.timeline, timelineEntry{kind: entryOwned, at: at, raw: record.Timestamp})
		}
	}
}

func responseKey(requestID, messageID string) string {
	if strings.TrimSpace(requestID) != "" {
		return "request:" + strings.TrimSpace(requestID)
	}
	if strings.TrimSpace(messageID) != "" {
		return "message:" + strings.TrimSpace(messageID)
	}
	return ""
}

func (t *Transcript) readAssistant(record rawRecord, at time.Time, line int) {
	var message rawAssistantMessage
	if err := json.Unmarshal(record.Message, &message); err != nil || len(record.Message) == 0 {
		t.incomplete(fmt.Sprintf("line %d: assistant record has no readable message", line))
		return
	}
	terminal := message.StopReason != nil && *message.StopReason != "" && *message.StopReason != "tool_use" && *message.StopReason != "pause_turn"
	if message.Model == syntheticModel {
		// A synthetic notice can end a turn but was not served by a model.
		t.timeline = append(t.timeline, timelineEntry{kind: entryAssistant, at: at, raw: record.Timestamp, terminal: terminal})
		return
	}
	key := responseKey(record.RequestID, message.ID)
	if key == "" {
		t.incomplete(fmt.Sprintf("line %d: assistant response has neither requestId nor message.id", line))
		return
	}
	usage, usageErr := decodeUsage(message.Usage)
	if usageErr != nil {
		t.incomplete(fmt.Sprintf("line %d: response %s %v", line, key, usageErr))
		return
	}
	model := strings.TrimSpace(message.Model)
	if model == "" {
		t.incomplete(fmt.Sprintf("line %d: response %s has no served model", line, key))
		return
	}
	current := response{key: key, model: model, timestamp: record.Timestamp, input: *usage.Input, output: *usage.Output, cacheRead: *usage.CacheRead}
	if usage.CacheCreation != nil {
		current.cacheCreation = *usage.CacheCreation
	}
	if usage.OutputTokenDetails != nil && usage.OutputTokenDetails.Thinking != nil {
		current.thinking = *usage.OutputTokenDetails.Thinking
	}
	if index, seen := t.responseIndex[key]; seen {
		previous := t.responses[index]
		current.timestamp = previous.timestamp
		if previous != current {
			t.incomplete(fmt.Sprintf("line %d: response %s carries conflicting usage or model", line, key))
			return
		}
	} else {
		t.responseIndex[key] = len(t.responses)
		t.responses = append(t.responses, current)
	}
	var blocks []rawBlock
	if len(message.Content) > 0 && json.Unmarshal(message.Content, &blocks) == nil {
		for _, block := range blocks {
			if block.Type == "tool_use" {
				t.addToolUse(block, record.Timestamp, key, line)
			}
		}
	}
	t.timeline = append(t.timeline, timelineEntry{kind: entryAssistant, at: at, raw: record.Timestamp, responseKey: key, model: model, terminal: terminal})
}

func decodeUsage(raw json.RawMessage) (rawUsage, error) {
	var usage rawUsage
	if len(raw) == 0 || string(raw) == "null" {
		return usage, errors.New("lacks usage")
	}
	if err := json.Unmarshal(raw, &usage); err != nil {
		return usage, errors.New("has unreadable usage")
	}
	if usage.Input == nil || usage.Output == nil || usage.CacheRead == nil {
		return usage, errors.New("lacks input, output or cache-read usage")
	}
	if *usage.Input < 0 || *usage.Output < 0 || *usage.CacheRead < 0 || (usage.CacheCreation != nil && *usage.CacheCreation < 0) {
		return usage, errors.New("has negative usage")
	}
	return usage, nil
}

func (t *Transcript) addToolUse(block rawBlock, timestamp, key string, line int) {
	if strings.TrimSpace(block.ID) == "" || strings.TrimSpace(block.Name) == "" {
		t.incomplete(fmt.Sprintf("line %d: tool_use requires id and name", line))
		return
	}
	if _, seen := t.toolIndex[block.ID]; seen {
		return
	}
	use := toolUse{id: block.ID, name: block.Name, timestamp: timestamp, responseKey: key, turn: t.promptCount}
	switch block.Name {
	case "Skill":
		var input struct {
			Skill string `json:"skill"`
		}
		if json.Unmarshal(block.Input, &input) == nil {
			use.skill = strings.TrimSpace(input.Skill)
		}
	case "Agent":
		var input struct {
			SubagentType string `json:"subagent_type"`
			Description  string `json:"description"`
		}
		if json.Unmarshal(block.Input, &input) == nil {
			use.subagentType = strings.TrimSpace(input.SubagentType)
			use.description = strings.TrimSpace(input.Description)
		}
	}
	t.toolIndex[block.ID] = len(t.toolUses)
	t.toolUses = append(t.toolUses, use)
}

func (t *Transcript) readUser(record rawRecord, at time.Time, line int) {
	var message rawUserMessage
	if err := json.Unmarshal(record.Message, &message); err != nil || len(record.Message) == 0 {
		t.incomplete(fmt.Sprintf("line %d: user record has no readable message", line))
		return
	}
	var text string
	if json.Unmarshal(message.Content, &text) == nil {
		t.readPrompt(record, at)
		return
	}
	var blocks []rawBlock
	if err := json.Unmarshal(message.Content, &blocks); err != nil {
		t.incomplete(fmt.Sprintf("line %d: user record has unreadable content", line))
		return
	}
	var results []rawBlock
	for _, block := range blocks {
		if block.Type == "tool_result" {
			results = append(results, block)
		}
	}
	if len(results) == 0 {
		t.readPrompt(record, at)
		return
	}
	for _, block := range results {
		if strings.TrimSpace(block.ToolUseID) == "" {
			t.incomplete(fmt.Sprintf("line %d: tool_result requires tool_use_id", line))
			continue
		}
		result := toolResult{timestamp: record.Timestamp, isError: block.IsError != nil && *block.IsError}
		// Record-level result details are attributable only when one result
		// shares the record.
		if len(results) == 1 {
			var details struct {
				Success   *bool  `json:"success"`
				AgentID   string `json:"agentId"`
				AgentType string `json:"agentType"`
			}
			if json.Unmarshal(record.ToolUseResult, &details) == nil {
				result.success = details.Success
				result.agentID = strings.TrimSpace(details.AgentID)
				result.agentType = strings.TrimSpace(details.AgentType)
			}
		}
		t.results[block.ToolUseID] = result
	}
	t.timeline = append(t.timeline, timelineEntry{kind: entryOwned, at: at, raw: record.Timestamp})
}

// readPrompt starts a turn only for a prompt that is not a meta record, a
// compaction summary or transcript-only text.
func (t *Transcript) readPrompt(record rawRecord, at time.Time) {
	if record.IsMeta && t.AgentID != "" && originKind(record.Origin) == "coordinator" && !record.IsCompactSummary && !record.TranscriptOnly {
		t.promptCount++
		t.timeline = append(t.timeline, timelineEntry{kind: entryResume, at: at, raw: record.Timestamp, nativeID: strings.TrimSpace(record.UUID), system: true})
		return
	}
	if record.IsMeta || record.IsCompactSummary || record.TranscriptOnly {
		t.timeline = append(t.timeline, timelineEntry{kind: entryOwned, at: at, raw: record.Timestamp})
		return
	}
	t.promptCount++
	t.timeline = append(t.timeline, timelineEntry{kind: entryPrompt, at: at, raw: record.Timestamp, nativeID: strings.TrimSpace(record.UUID), system: record.PromptSource == "system"})
}

func originKind(raw json.RawMessage) string {
	var origin struct {
		Kind string `json:"kind"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &origin) != nil {
		return ""
	}
	return strings.TrimSpace(origin.Kind)
}

// dropInheritedResponses removes responses whose identity also appears in
// copied fork history, so a response is counted only for its original session.
func (t *Transcript) dropInheritedResponses() {
	if len(t.forkedKeys) == 0 {
		return
	}
	kept := t.responses[:0]
	for _, current := range t.responses {
		if !t.forkedKeys[current.key] {
			kept = append(kept, current)
		}
	}
	t.responses = kept
	t.responseIndex = make(map[string]int, len(kept))
	for i, current := range kept {
		t.responseIndex[current.key] = i
	}
}
