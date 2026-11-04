package domain

// RecordedOutcome preserves unknown evidence; historical boolean conversion
// belongs solely to the explicit migration boundary.
func (fact SkillFact) RecordedOutcome() FactOutcome {
	if fact.Outcome != "" {
		return fact.Outcome
	}
	return OutcomeUnknown
}

// SkillFacts flattens only canonical session-owned occurrences.
func (ledger AgenticFactLedger) SkillFacts() []SkillFact {
	result := []SkillFact{}
	for _, session := range ledger.Sessions {
		for _, fact := range session.Skills {
			fact.SessionID, fact.Source = session.ID, session.Source
			result = append(result, fact)
		}
	}
	return result
}
func (ledger AgenticFactLedger) TaskLinkFacts() []TaskLinkFact {
	result := []TaskLinkFact{}
	for _, session := range ResolveTaskChildren(ledger.Sessions) {
		for _, fact := range session.TaskLinks {
			fact.SessionID, fact.Source = session.ID, session.Source
			result = append(result, fact)
		}
	}
	return result
}

// ResolveTaskChildren uses explicit source parent identity and agent namespace.
// A namespace with zero or multiple matching children remains unresolved; role,
// ordering and timestamp never establish a child relationship. It is pure and
// usable from persisted metadata without re-reading a private source store.
func ResolveTaskChildren(sessions []SessionFact) []SessionFact {
	result := append([]SessionFact(nil), sessions...)
	children := make(map[string][]string)
	byID := make(map[string]SessionFact)
	for _, child := range sessions {
		byID[child.ID] = child
		if child.Source == SourceCodex && child.ParentID != nil && child.AgentPath != "" {
			key := *child.ParentID + "\x00" + child.AgentPath
			children[key] = append(children[key], child.ID)
		}
	}
	for si := range result {
		owner := &result[si]
		owner.TaskLinks = append([]TaskLinkFact(nil), owner.TaskLinks...)
		for i := range owner.TaskLinks {
			link := &owner.TaskLinks[i]
			if link.ChildSession == "" && link.ChildAgentID != "" {
				if child, known := byID[link.ChildAgentID]; known && child.ParentID != nil && *child.ParentID == owner.ID {
					link.ChildSession = child.ID
				}
			}
			if child, exists := byID[link.ChildSession]; exists && (child.ParentID == nil || *child.ParentID != owner.ID) {
				link.ChildLinkDiagnostic = "Recorded child metadata contradicts the delegating parent; native child reference preserved"
				continue
			}
			if owner.Source == SourceCodex && link.ChildReference != "" && link.ChildSession != "" {
				if child, known := byID[link.ChildSession]; known && child.AgentPath != "" && child.AgentPath != link.ChildReference {
					link.ChildLinkDiagnostic = "Recorded explicit child session has a different native namespace; native references preserved"
					continue
				}
			}
			if owner.Source != SourceCodex || link.ChildReference == "" {
				continue
			}
			matches := children[owner.ID+"\x00"+link.ChildReference]
			if link.ChildSession != "" {
				if len(matches) > 0 {
					found := false
					for _, id := range matches {
						if id == link.ChildSession {
							found = true
						}
					}
					if !found {
						link.ChildLinkDiagnostic = "Recorded child namespace and explicit child session disagree; native references preserved"
					} else {
						link.ChildLinkDiagnostic = ""
					}
				}
				continue
			}
			if len(matches) == 1 {
				link.ChildSession = matches[0]
				link.ChildLinkDiagnostic = ""
			}
			if len(matches) > 1 {
				link.ChildLinkDiagnostic = "Multiple recorded children share this native namespace; no child session was inferred"
			}
		}
	}
	return result
}

// PersistedOccurrenceEvents retains safe metadata only from actual source events
// proven to describe an owned occurrence. Facts never synthesize event rows.
func PersistedOccurrenceEvents(session SessionFact) []SessionEventFact {
	eventIDs, callIDs, messageIDs, sourceRefs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	add := func(p FactProvenance) {
		if p.EventID != "" {
			eventIDs[p.EventID] = true
		}
		if p.CallID != "" {
			callIDs[p.CallID] = true
		}
		if p.MessageID != "" {
			messageIDs[p.MessageID] = true
		}
		if p.SourceRef != "" {
			sourceRefs[p.SourceRef] = true
		}
	}
	for _, fact := range session.Skills {
		add(fact.Provenance)
	}
	for _, fact := range session.TaskLinks {
		add(fact.Provenance)
	}
	var result []SessionEventFact
	for _, event := range session.Events {
		if event.EventType != EventTypeToolCall && event.EventType != EventTypeSkillActivation {
			continue
		}
		matched := eventIDs[event.EventID] || event.CallID != "" && callIDs[event.CallID]
		if event.EventType == EventTypeSkillActivation {
			matched = matched || event.MessageID != nil && messageIDs[*event.MessageID] || sourceRefs[event.EventID]
		}
		if !matched {
			continue
		}
		result = append(result, SessionEventFact{CallID: event.CallID, EventID: event.EventID, Timestamp: event.Timestamp, MessageID: event.MessageID, Turn: event.Turn, Role: event.Role, EventType: event.EventType, ToolName: event.ToolName, Status: event.Status, DurationMs: event.DurationMs, ContextTokens: event.ContextTokens})
	}
	return result
}
