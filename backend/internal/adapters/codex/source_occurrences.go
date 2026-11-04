package codex

import "agenticlens/internal/domain"

// Parser-local identity exists only until envelope deduplication and native child
// binding finish. These types are never accepted by ledger or capture writers.
type sourceSkillOccurrence struct {
	EvidenceReferences []domain.FactProvenance
	Provenance         domain.FactProvenance
	Path               string
	Name               string
	CallerRole         string
	Timestamp          string
}
type sourceDelegationOccurrence struct {
	ChildAgentID   string
	ChildReference string
	Provenance     domain.FactProvenance
	SubagentType   string
	Description    string
	ChildSession   string
	Timestamp      string
}

func (t Transcript) leanSkills() []domain.SkillFact {
	result := make([]domain.SkillFact, 0, len(t.Skills))
	for _, skill := range t.Skills {
		result = append(result, domain.SkillFact{Name: skill.Name, CallerRole: domain.KnownString(skill.CallerRole), InvokedAt: domain.KnownString(skill.Timestamp)})
	}
	return result
}
func (t Transcript) leanDelegations() []domain.DelegationFact {
	result := make([]domain.DelegationFact, 0, len(t.Delegations))
	for _, link := range t.Delegations {
		result = append(result, domain.DelegationFact{AgentType: domain.KnownString(link.SubagentType), Description: link.Description, InvokedAt: domain.KnownString(link.Timestamp), ChildSession: link.ChildSession})
	}
	return result
}
func sessionIDs(sessions []domain.SessionFact) []string {
	result := make([]string, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, session.ID)
	}
	return result
}

// resolveSourceChildren consumes native metadata before projecting lean facts.
func resolveSourceChildren(transcripts []Transcript) []Transcript {
	result := append([]Transcript{}, transcripts...)
	byID := make(map[string]Transcript)
	children := make(map[string][]string)
	for _, child := range result {
		byID[child.sessionIDOrThread()] = child
		if child.ParentThread != "" && child.AgentPath != "" {
			key := child.ParentThread + "\x00" + child.AgentPath
			children[key] = append(children[key], child.sessionIDOrThread())
		}
	}
	for ti := range result {
		owner := &result[ti]
		owner.Delegations = append([]sourceDelegationOccurrence{}, owner.Delegations...)
		owner.Events = append([]domain.EventFact{}, owner.Events...)
		for i := range owner.Delegations {
			link := &owner.Delegations[i]
			if link.ChildSession == "" && link.ChildAgentID != "" {
				if child, ok := byID[link.ChildAgentID]; ok && child.ParentThread == owner.sessionIDOrThread() {
					link.ChildSession = child.sessionIDOrThread()
				}
			}
			if link.ChildSession == "" && link.ChildReference != "" {
				matches := children[owner.sessionIDOrThread()+"\x00"+link.ChildReference]
				if len(matches) == 1 {
					link.ChildSession = matches[0]
				}
			}
			if child, ok := byID[link.ChildSession]; ok && child.ParentThread != owner.sessionIDOrThread() {
				link.ChildSession = ""
				owner.Warnings = append(owner.Warnings, "delegation child metadata contradicts native parent")
			}
			{
				for ei := range owner.Events {
					event := &owner.Events[ei]
					if event.CallID == link.Provenance.CallID && event.CallID != "" {
						event.ChildSessionID = link.ChildSession
					}
				}
			}
		}
	}
	return result
}
