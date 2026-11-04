package codex

import (
	"agenticlens/internal/domain"
	"encoding/json"
	"regexp"
	"strings"
)

// Selected-skill syntax is interpreted only in native user messages. Catalogues,
// quoted instructions, shell reads and replayed parent messages are not activations.
var directSkillRequest = regexp.MustCompile(`^\$([A-Za-z][A-Za-z0-9_.:-]*)(?:\s|$)`)
var attachedSkill = regexp.MustCompile(`(?s)^\s*<skill>\s*<name>([^<]+)</name>\s*<path>([^<]+)</path>`)

type responseOccurrence struct {
	Type      string          `json:"type"`
	TurnID    string          `json:"turn_id"`
	ID        string          `json:"id"`
	Role      string          `json:"role"`
	CallID    string          `json:"call_id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
	Content   []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (t *Transcript) readOccurrence(payload json.RawMessage, timestamp, sourceID string) {
	var item responseOccurrence
	if json.Unmarshal(payload, &item) != nil {
		t.Warnings = append(t.Warnings, sourceID+": invalid response occurrence payload")
		return
	}
	provenance := domain.FactProvenance{SourceRef: sourceID, ItemID: item.ID, CallID: item.CallID, TurnID: item.TurnID}
	if item.Type == "message" {
		provenance.MessageID = item.ID
	}
	if item.Type == "message" {
		t.messageCount++
	}
	if item.Type == "message" && item.Role == "user" {
		previousSelection := t.pendingSelection
		t.pendingSelection = -1
		for _, content := range item.Content {
			if content.Type != "input_text" && content.Type != "text" {
				continue
			}
			text := strings.TrimSpace(content.Text)
			name, path := "", ""
			attachment := attachedSkill.FindStringSubmatch(text)
			if len(attachment) > 0 {
				name, path = strings.TrimSpace(attachment[1]), strings.TrimSpace(attachment[2])
			} else if selected := directSkillRequest.FindStringSubmatch(text); len(selected) > 0 {
				name = selected[1]
			}
			if name == "" {
				t.pendingSelection = -1
				previousSelection = -1
				continue
			}
			if len(attachment) > 0 && previousSelection >= 0 && previousSelection < len(t.Skills) {
				previous := &t.Skills[previousSelection]
				if previous.Name == name && previous.Path == "" && len(previous.EvidenceReferences) == 0 {
					previous.Path = path
					previous.EvidenceReferences = append(previous.EvidenceReferences, provenance)
					previousSelection = -1
					t.pendingSelection = -1
					continue
				}
			}
			if item.ID != "" {
				provenance.EventID = item.ID
			}
			fact := sourceSkillOccurrence{Name: name, Path: path, CallerRole: "user", Timestamp: timestamp, Provenance: provenance}
			t.Skills = append(t.Skills, fact)
			if len(attachment) == 0 {
				t.pendingSelection = len(t.Skills) - 1
			} else {
				t.pendingSelection = -1
			}
			eventID := item.ID
			var messageID *string
			if item.ID != "" {
				id := item.ID
				messageID = &id
			} else {
				eventID = sourceID
			}
			t.Events = append(t.Events, domain.EventFact{EventID: eventID, Timestamp: timestamp, SessionID: t.sessionIDOrThread(), MessageID: messageID, Turn: t.Turns, Role: "user", SkillName: name, EventType: domain.EventTypeSkillActivation, Status: "unknown", ContextTokens: t.currentContext})
			previousSelection = t.pendingSelection
		}
		return
	}
	if item.Type == "message" || item.Type == "function_call" || item.Type == "custom_tool_call" {
		t.pendingSelection = -1
	}
	if item.Type != "function_call" && item.Type != "custom_tool_call" || item.CallID == "" {
		return
	}
	for _, previous := range t.Delegations {
		if previous.Provenance.CallID == item.CallID {
			return
		}
	}
	for _, previous := range t.Skills {
		if previous.Provenance.CallID == item.CallID {
			return
		}
	}
	name := item.Name
	if index := strings.LastIndex(name, "."); index >= 0 {
		name = name[index+1:]
	}
	switch name {
	case "spawn_agent":
		var args struct {
			AgentType string `json:"agent_type"`
			TaskName  string `json:"task_name"`
			Message   string `json:"message"`
		}
		if json.Unmarshal(nativeCallArguments(item.Arguments), &args) != nil {
			t.Warnings = append(t.Warnings, "delegation arguments unavailable for call "+item.CallID)
			return
		}
		description := args.TaskName
		// The prompt body is intentionally not persisted as a description.
		t.Delegations = append(t.Delegations, sourceDelegationOccurrence{SubagentType: args.AgentType, Description: description, Timestamp: timestamp, Provenance: provenance})
	case "skill", "load_skill":
		var args struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if json.Unmarshal(nativeCallArguments(item.Arguments), &args) != nil || strings.TrimSpace(args.Name) == "" {
			t.Warnings = append(t.Warnings, "skill arguments unavailable for call "+item.CallID)
			return
		}
		t.Skills = append(t.Skills, sourceSkillOccurrence{Name: args.Name, Path: args.Path, CallerRole: string(t.role), Timestamp: timestamp, Provenance: provenance})
	}
}

func (t *Transcript) completeOccurrence(callID, output string, event domain.EventFact) {
	for i := range t.Delegations {
		link := &t.Delegations[i]
		if link.Provenance.CallID != callID {
			continue
		}
		var result struct {
			AgentID   string `json:"agent_id"`
			SessionID string `json:"session_id"`
			TaskName  string `json:"task_name"`
		}
		if json.Unmarshal([]byte(output), &result) == nil {
			link.ChildReference = result.TaskName
			link.ChildAgentID = result.AgentID
			link.ChildSession = result.SessionID
			if result.SessionID != "" {
				for i := range t.Events {
					if t.Events[i].CallID == callID {
						t.Events[i].ChildSessionID = result.SessionID
					}
				}
			}
			if link.Description == "" {
				link.Description = result.TaskName
			}
		}
	}
}

// Native Codex function-call arguments are a JSON string containing an object.
// Return invalid JSON for other shapes so recognized producers warn rather than
// authorizing an authoritative empty snapshot.
func nativeCallArguments(raw json.RawMessage) []byte {
	var encoded string
	if json.Unmarshal(raw, &encoded) != nil || strings.TrimSpace(encoded) == "" {
		return nil
	}
	return []byte(encoded)
}
