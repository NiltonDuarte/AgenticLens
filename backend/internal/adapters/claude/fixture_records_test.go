package claude

// Typed builders for synthetic Claude Code 2.1.331 transcript records. They
// reproduce native keys and identities only, never transcript content. Fields
// that tests must write in malformed form (usage and content) are encoded JSON
// values, so every fixture stays typed without an empty-interface map.

import (
	"encoding/json"
	"fmt"
)

type record struct {
	Type              string           `json:"type,omitempty"`
	Subtype           string           `json:"subtype,omitempty"`
	SessionID         string           `json:"sessionId,omitempty"`
	AgentID           string           `json:"agentId,omitempty"`
	UUID              string           `json:"uuid,omitempty"`
	Timestamp         string           `json:"timestamp,omitempty"`
	CWD               string           `json:"cwd,omitempty"`
	PromptSource      string           `json:"promptSource,omitempty"`
	IsSidechain       bool             `json:"isSidechain,omitempty"`
	IsMeta            bool             `json:"isMeta,omitempty"`
	IsCompactSummary  bool             `json:"isCompactSummary,omitempty"`
	TranscriptOnly    bool             `json:"isVisibleInTranscriptOnly,omitempty"`
	IsAPIErrorMessage bool             `json:"isApiErrorMessage,omitempty"`
	Origin            *origin          `json:"origin,omitempty"`
	ForkedFrom        *forkSource      `json:"forkedFrom,omitempty"`
	RequestID         string           `json:"requestId,omitempty"`
	Message           *message         `json:"message,omitempty"`
	ToolUseResult     *resultDetails   `json:"toolUseResult,omitempty"`
	CompactMetadata   *compactMetadata `json:"compactMetadata,omitempty"`
	LogicalParentUUID string           `json:"logicalParentUuid,omitempty"`
	TotalCostUSD      float64          `json:"totalCostUSD,omitempty"`
	ModelUsage        map[string]cost  `json:"modelUsage,omitempty"`
	MessageID         string           `json:"messageId,omitempty"`
	Operation         string           `json:"operation,omitempty"`
	// OmitAgentID keeps withIdentity from adding the subagent agentId.
	OmitAgentID bool `json:"-"`
}

type origin struct {
	Kind string `json:"kind"`
}

type forkSource struct {
	SessionID   string `json:"sessionId"`
	MessageUUID string `json:"messageUuid"`
}

type compactMetadata struct {
	Trigger string `json:"trigger"`
}

type cost struct {
	CostUSD float64 `json:"costUSD"`
}

type message struct {
	ID         string          `json:"id,omitempty"`
	Model      string          `json:"model,omitempty"`
	Role       string          `json:"role,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	Usage      json.RawMessage `json:"usage,omitempty"`
	Content    json.RawMessage `json:"content,omitempty"`
}

type block struct {
	Type      string     `json:"type"`
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Text      string     `json:"text,omitempty"`
	Input     *toolInput `json:"input,omitempty"`
	ToolUseID string     `json:"tool_use_id,omitempty"`
	IsError   *bool      `json:"is_error,omitempty"`
}

type toolInput struct {
	Skill        string `json:"skill,omitempty"`
	Args         string `json:"args,omitempty"`
	SubagentType string `json:"subagent_type,omitempty"`
	Description  string `json:"description,omitempty"`
}

type resultDetails struct {
	Success     *bool  `json:"success,omitempty"`
	CommandName string `json:"commandName,omitempty"`
	AgentID     string `json:"agentId,omitempty"`
	AgentType   string `json:"agentType,omitempty"`
	Status      string `json:"status,omitempty"`
}

type usageFields struct {
	Input         *int64         `json:"input_tokens,omitempty"`
	Output        *int64         `json:"output_tokens,omitempty"`
	CacheRead     *int64         `json:"cache_read_input_tokens,omitempty"`
	CacheCreation *int64         `json:"cache_creation_input_tokens,omitempty"`
	Details       *outputDetails `json:"output_tokens_details,omitempty"`
}

type outputDetails struct {
	Thinking int64 `json:"thinking_tokens"`
}

func n64(value int64) *int64 { return &value }

func boolPtr(value bool) *bool { return &value }

// encodable lists the fixture values written as raw JSON fields.
type encodable interface {
	usageFields | []block | string | int
}

// encode marshals a typed fixture value; the fixture types cannot fail to encode.
func encode[T encodable](value T) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("synthetic fixture encoding: %v", err))
	}
	return data
}

func (u usageFields) encode() json.RawMessage { return encode(u) }

// edit returns a copy of r with an independent message, changed by change.
func (r record) edit(change func(*record)) record {
	copied := r
	if r.Message != nil {
		m := *r.Message
		copied.Message = &m
	}
	change(&copied)
	return copied
}

func textContent(text string) json.RawMessage { return encode(text) }

type metaFixture struct {
	AgentType string `json:"agentType"`
	ToolUseID string `json:"toolUseId"`
}
