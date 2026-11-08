package opencode

import (
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
)

// SchemaV1PartData is the OpenCode v1 part representation shared by live and
// retrospective projections. It is transient source data, never a ledger DTO.
type SchemaV1PartData struct {
	Text  *string            `json:"text"`
	Type  *string            `json:"type"`
	Tool  *string            `json:"tool"`
	State *SchemaV1PartState `json:"state"`
}
type SchemaV1PartState struct {
	Status   *string         `json:"status"`
	Output   json.RawMessage `json:"output"`
	Input    json.RawMessage `json:"input"`
	Metadata json.RawMessage `json:"metadata"`
	Error    json.RawMessage `json:"error"`
}

// UnmarshalJSON isolates unsupported optional/status values from supported
// input/output fields. An unreadable status remains unknown, never successful.
func (s *SchemaV1PartState) UnmarshalJSON(data []byte) error {
	var fields struct {
		Status   json.RawMessage `json:"status"`
		Output   json.RawMessage `json:"output"`
		Input    json.RawMessage `json:"input"`
		Metadata json.RawMessage `json:"metadata"`
		Error    json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*s = SchemaV1PartState{Output: fields.Output, Input: fields.Input, Metadata: fields.Metadata, Error: fields.Error}
	var status string
	if len(fields.Status) > 0 && !bytes.Equal(bytes.TrimSpace(fields.Status), []byte("null")) && json.Unmarshal(fields.Status, &status) == nil {
		s.Status = &status
	}
	return nil
}

// ProcessEvidence reads the documented bash-tool metadata.exit field only.
// OpenCode v1.2.15 tool/bash.ts returns metadata.exit = proc.exitCode;
// missing, null or malformed exit metadata does not establish a process result.
func (p SchemaV1PartData) ProcessEvidence() domain.ToolProcessEvidence {
	if p.State == nil || p.Tool == nil || *p.Tool != "bash" {
		return domain.ToolProcessEvidence{}
	}
	var metadata struct {
		Exit *int `json:"exit"`
	}
	if json.Unmarshal(p.State.Metadata, &metadata) != nil {
		return domain.ToolProcessEvidence{}
	}
	return domain.ToolProcessEvidence{ExitCode: metadata.Exit}
}

// SchemaV1RawValue preserves string contents and structured values verbatim.
func SchemaV1RawValue(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return ""
	}
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	return string(raw)
}
