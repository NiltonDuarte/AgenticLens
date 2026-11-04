package opencode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"agenticlens/internal/domain"
)

// ExtractSkillTaskFacts projects each source occurrence once, preserving native sequence.
func ExtractSkillTaskFacts(events []domain.EventFact) ([]domain.SkillFact, []domain.DelegationFact) {
	skills := []domain.SkillFact{}
	delegations := []domain.DelegationFact{}
	seen := make(map[string]bool)
	for _, event := range events {
		if event.EventType != domain.EventTypeToolCall || event.ToolName == nil {
			continue
		}
		key := event.SessionID + "\x00" + event.EventID
		if event.EventID != "" {
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		var input map[string]json.RawMessage
		if event.ToolInput == nil || json.Unmarshal([]byte(*event.ToolInput), &input) != nil {
			continue
		}
		switch strings.TrimSpace(*event.ToolName) {
		case "skill":
			name := rawString(input["name"])
			if name == "" {
				continue
			}
			skills = append(skills, domain.SkillFact{Name: name, CallerRole: domain.KnownString(string(event.Role)), InvokedAt: event.InvokedAt})
		case "task":
			child := event.ChildSessionID
			delegations = append(delegations, domain.DelegationFact{AgentType: domain.KnownString(rawString(input["subagent_type"])), Description: rawString(input["description"]), InvokedAt: event.InvokedAt, ChildSession: child})
		}
	}
	return skills, delegations
}

func rawString(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func outputSessionID(output string) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal([]byte(output), &fields) != nil {
		return ""
	}
	for _, key := range []string{"child_session_id", "session_id", "sessionID"} {
		if value := rawString(fields[key]); value != "" {
			return value
		}
	}
	return ""
}

// WithOwnedSkillTaskFacts consumes transient source rows before persistence.
func WithOwnedSkillTaskFacts(sessions []domain.SessionFact, events []domain.EventFact) []domain.SessionFact {
	result := append([]domain.SessionFact{}, sessions...)
	for i, owner := range result {
		if owner.Source != "" && owner.Source != domain.SourceOpenCode {
			continue
		}
		rows := []domain.EventFact{}
		for _, event := range events {
			if event.SessionID == owner.ID {
				rows = append(rows, event)
			}
		}
		result[i].Skills, result[i].Delegations = ExtractSkillTaskFacts(rows)
	}
	return result
}
func projectSourceFacts(sessions []domain.SessionFact, events []domain.EventFact, strict bool) ([]domain.SessionFact, error) {
	for _, event := range events {
		if !strict || event.EventType != domain.EventTypeToolCall || event.ToolName == nil || (*event.ToolName != "skill" && *event.ToolName != "task") {
			continue
		}
		var fields map[string]json.RawMessage
		if event.ToolInput == nil || json.Unmarshal([]byte(*event.ToolInput), &fields) != nil || fields == nil {
			return nil, fmt.Errorf("incomplete %s source call in session %s", *event.ToolName, event.SessionID)
		}
		if *event.ToolName == "skill" && rawString(fields["name"]) == "" {
			return nil, fmt.Errorf("skill call has no native name in session %s", event.SessionID)
		}
	}
	sessions = WithOwnedSkillTaskFacts(sessions, events)
	complete := make([]string, 0, len(sessions))
	for _, owner := range sessions {
		complete = append(complete, owner.ID)
	}
	facts, err := domain.PersistSourceAnomalySummaries(domain.AgenticFactLedger{Sessions: sessions}, domain.SourceEvents{Events: events, CompleteSessionIDs: complete})
	if err != nil {
		return nil, err
	}
	return facts.Sessions, nil
}

// proveSourceChildLinks checks explicit output references against native metadata
// even when the child was outside this exact owner's extraction. Unknown native
// IDs remain recorded references; a known different parent contradicts the link.
func proveSourceChildLinks(db *sql.DB, table string, events []domain.EventFact) error {
	for i := range events {
		event := &events[i]
		if event.EventType != domain.EventTypeToolCall || event.ToolName == nil || *event.ToolName != "task" || event.ToolOutput == nil {
			continue
		}
		child := outputSessionID(*event.ToolOutput)
		if child == "" {
			continue
		}
		var parent sql.NullString
		err := db.QueryRow("SELECT parent_id FROM "+table+" WHERE id=?", child).Scan(&parent)
		if errors.Is(err, sql.ErrNoRows) {
			event.ChildSessionID = child
			continue
		}
		if err != nil {
			return fmt.Errorf("checking native child %s for parent %s: %w", child, event.SessionID, err)
		}
		if parent.Valid && strings.TrimSpace(parent.String) == event.SessionID {
			event.ChildSessionID = child
		} else {
			event.ChildSessionID = ""
		}
	}
	return nil
}
