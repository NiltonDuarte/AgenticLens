package opencode

import (
	"encoding/json"
	"strings"

	"agenticlens/internal/domain"
)

// ExtractSkillTaskFacts derives optional skill and task facts from normalized
// OpenCode session events. Explicit output child references survive even when
// the child session is unavailable locally.
func ExtractSkillTaskFacts(sessions []domain.SessionFact) ([]domain.SkillFact, []domain.TaskLinkFact) {

	var skills []domain.SkillFact
	var taskLinks []domain.TaskLinkFact
	for _, session := range sessions {
		messageIndexes := make(map[string]int)
		nextIndex := 0
		for _, event := range session.Events {
			if event.MessageID != nil {
				if _, exists := messageIndexes[*event.MessageID]; !exists {
					messageIndexes[*event.MessageID] = nextIndex
					nextIndex++
				}
			}
			if event.EventType != domain.EventTypeToolCall || event.ToolName == nil {
				continue
			}
			tool := strings.TrimSpace(*event.ToolName)
			var input map[string]json.RawMessage
			if event.ToolInput == nil || json.Unmarshal([]byte(*event.ToolInput), &input) != nil {
				continue
			}
			switch tool {
			case "skill":
				name := rawString(input["name"])
				if name == "" {
					continue
				}
				messageIndex := 0
				if event.MessageID != nil {
					messageIndex = messageIndexes[*event.MessageID]
				}
				skill := domain.SkillFact{
					SessionID: session.ID, Source: session.Source, Provenance: eventProvenance(event), EvidenceKind: "tool_call", Outcome: eventOutcome(event.Status),
					Name: name, TurnIndex: int(event.Turn), MessageIndex: messageIndex,
					CallerRole: string(event.Role), Timestamp: event.Timestamp,
					Success: strings.EqualFold(event.Status, "completed"),
				}
				if skill.Outcome == domain.OutcomeFailed {
					if event.ErrorMessage != nil {
						skill.Error = *event.ErrorMessage
					}
					if skill.Error == "" && event.ToolOutput != nil {
						skill.Error = *event.ToolOutput
					}
				}
				skills = append(skills, skill)
			case "task":
				link := domain.TaskLinkFact{
					SessionID: session.ID, Source: session.Source, Provenance: eventProvenance(event), Outcome: eventOutcome(event.Status),
					CostEvidence: domain.UnavailableCost(),
					SubagentType: rawString(input["subagent_type"]),
					Description:  rawString(input["description"]),
					ParentTurn:   int(event.Turn), Timestamp: event.Timestamp,
				}
				if link.Outcome == domain.OutcomeFailed && event.ErrorMessage != nil {
					link.Error = *event.ErrorMessage
				}
				if event.DurationMs != nil && *event.DurationMs > 0 {
					link.DurationMs = *event.DurationMs
				}
				if event.ToolOutput != nil {
					childID := outputSessionID(*event.ToolOutput)
					if childID != "" {
						link.ChildSession = childID
					}
				}
				taskLinks = append(taskLinks, link)
			}
		}
	}
	return skills, taskLinks
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

func eventProvenance(event domain.SessionEventFact) domain.FactProvenance {
	p := domain.FactProvenance{EventID: event.EventID}
	if event.MessageID != nil {
		p.MessageID = *event.MessageID
	}
	return p
}
func eventOutcome(status string) domain.FactOutcome {
	switch strings.ToLower(status) {
	case "completed":
		return domain.OutcomeSucceeded
	case "error", "failed":
		return domain.OutcomeFailed
	default:
		return domain.OutcomeUnknown
	}
}

// WithOwnedSkillTaskFacts derives a complete source snapshot before event payloads
// are compacted. Other-source owners and their recorded snapshots are untouched.
func WithOwnedSkillTaskFacts(sessions []domain.SessionFact) []domain.SessionFact {
	result := append([]domain.SessionFact(nil), sessions...)
	for i, session := range result {
		if session.Source != "" && session.Source != domain.SourceOpenCode {
			continue
		}
		skills, links := ExtractSkillTaskFacts([]domain.SessionFact{session})
		for j := range skills {
			skills[j].SessionID, skills[j].Source = "", ""
		}
		for j := range links {
			links[j].SessionID, links[j].Source = "", ""
		}
		result[i].Skills, result[i].TaskLinks = skills, links
	}
	return result
}
