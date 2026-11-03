package opencode

import (
	"encoding/json"
	"strings"

	"agenticlens/internal/domain"
)

// ExtractSkillTaskFacts derives optional skill and task facts from normalized
// OpenCode session events. Child links require an output identifier that matches
// a direct child session's parent relationship.
func ExtractSkillTaskFacts(sessions []domain.SessionFact) ([]domain.SkillFact, []domain.TaskLinkFact) {
	childrenByParent := make(map[string]map[string]struct{})
	for _, session := range sessions {
		if session.ParentID == nil || strings.TrimSpace(*session.ParentID) == "" {
			continue
		}
		parent := strings.TrimSpace(*session.ParentID)
		if childrenByParent[parent] == nil {
			childrenByParent[parent] = make(map[string]struct{})
		}
		childrenByParent[parent][session.ID] = struct{}{}
	}

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
					FactProvenance: eventProvenance(session, event), Evidence: "skill_tool", Outcome: recordedSkillOutcome(event),
					Name: name, TurnIndex: int(event.Turn), MessageIndex: messageIndex,
					CallerRole: string(event.Role), Timestamp: event.Timestamp,
					Success: recordedSkillOutcome(event) == "succeeded",
				}
				if skill.Outcome == "failed" {
					if event.ErrorMessage != nil {
						skill.Error = strings.TrimSpace(*event.ErrorMessage)
					}
					if skill.Error == "" && event.ToolOutput != nil {
						skill.Error = strings.TrimSpace(*event.ToolOutput)
					}
				}
				skills = append(skills, skill)
			case "task":
				link := domain.TaskLinkFact{
					FactProvenance: eventProvenance(session, event),
					CostEvidence:   domain.UnavailableCost(),
					SubagentType:   rawString(input["subagent_type"]),
					Description:    rawString(input["description"]),
					ParentTurn:     int(event.Turn), Timestamp: event.Timestamp,
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
		// The native task tool also returns an explicit metadata envelope.
		if metadata, ok := strings.CutPrefix(strings.TrimSpace(output), "<task_metadata>"); ok {
			metadata, _, closed := strings.Cut(metadata, "</task_metadata>")
			if closed {
				for _, line := range strings.Split(metadata, "\n") {
					if key, value, found := strings.Cut(strings.TrimSpace(line), ":"); found && key == "session_id" {
						return strings.TrimSpace(value)
					}
				}
			}
		}
		return ""
	}
	for _, key := range []string{"child_session_id", "session_id", "sessionID"} {
		if value := rawString(fields[key]); value != "" {
			return value
		}
	}
	return ""
}

func eventProvenance(session domain.SessionFact, event domain.SessionEventFact) domain.FactProvenance {
	p := domain.FactProvenance{SessionID: session.ID, Source: session.Source, SourceEventID: event.EventID, SourceInvocationID: event.SourceInvocationID}
	if event.MessageID != nil {
		p.SourceMessageID = *event.MessageID
	}
	return p
}
func recordedSkillOutcome(event domain.SessionEventFact) string {
	if event.RecordedStatus != nil {
		return skillOutcome(*event.RecordedStatus)
	}
	return skillOutcome(event.Status)
}
func skillOutcome(status string) string {
	switch strings.ToLower(status) {
	case "completed", "success":
		return "succeeded"
	case "error", "failed":
		return "failed"
	default:
		return "unavailable"
	}
}

// WithSkillTaskFacts attaches extracted occurrences to their proven owners.
func WithSkillTaskFacts(sessions []domain.SessionFact) []domain.SessionFact {
	skills, links := ExtractSkillTaskFacts(sessions)
	ledger := domain.NormalizeAuxiliaryOwnership(domain.AgenticFactLedger{Sessions: sessions, Skills: skills, TaskLinks: links})
	for i := range ledger.Sessions {
		ledger.Sessions[i].AuxiliaryExtractionVersion = domain.CurrentAuxiliaryExtractionVersion
	}
	return ledger.Sessions
}
