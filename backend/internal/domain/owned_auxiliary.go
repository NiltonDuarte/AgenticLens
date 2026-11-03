package domain

// KnownOutcome distinguishes legacy tool outcomes from explicitly unknown execution.
func (fact SkillFact) KnownOutcome() bool {
	return fact.Outcome == "" || fact.Outcome == "succeeded" || fact.Outcome == "failed"
}
func (fact SkillFact) Succeeded() bool {
	if fact.Outcome != "" {
		return fact.Outcome == "succeeded"
	}
	return fact.Success
}

func occurrenceKeys(p FactProvenance) []string {
	if p.SessionID == "" {
		return nil
	}
	var keys []string
	if p.SourceInvocationID != "" {
		keys = append(keys, p.SessionID+"\x00call:"+p.SourceInvocationID)
	}
	if p.SourceEventID != "" {
		keys = append(keys, p.SessionID+"\x00event:"+p.SourceEventID)
	}
	return keys
}
func projectedEcho(p FactProvenance, name string, keys map[string]SourceHarness) bool {
	for _, key := range occurrenceKeys(p) {
		source, exists := keys[key+"\x00"+name]
		if exists && (p.Source == "" || source == "" || p.Source == source) {
			return true
		}
	}
	return false
}

// SkillFacts derives a projection, preserving unresolved history and multiplicity.
func (ledger AgenticFactLedger) SkillFacts() []SkillFact {
	var result []SkillFact
	keys := map[string]SourceHarness{}
	for _, owner := range ledger.Sessions {
		for _, fact := range owner.Skills {
			fact.SessionID, fact.Source = owner.ID, owner.Source
			result = append(result, fact)
			for _, key := range occurrenceKeys(fact.FactProvenance) {
				keys[key+"\x00"+fact.Name] = owner.Source
			}
		}
	}
	for _, fact := range ledger.Skills {
		if !projectedEcho(fact.FactProvenance, fact.Name, keys) {
			result = append(result, fact)
		}
	}
	return result
}

// TaskLinkFacts derives ticket-wide delegations without changing their parent owner.
func (ledger AgenticFactLedger) TaskLinkFacts() []TaskLinkFact {
	var result []TaskLinkFact
	keys := map[string]SourceHarness{}
	for _, owner := range ledger.Sessions {
		for _, fact := range owner.TaskLinks {
			fact.SessionID, fact.Source = owner.ID, owner.Source
			result = append(result, fact)
			for _, key := range occurrenceKeys(fact.FactProvenance) {
				keys[key+"\x00"] = owner.Source
			}
		}
	}
	for _, fact := range ledger.TaskLinks {
		if !projectedEcho(fact.FactProvenance, "", keys) {
			result = append(result, fact)
		}
	}
	return result
}

// NormalizeAuxiliaryOwnership moves only explicitly attributed root facts. Unknown
// historical owners remain lossless at the root with an explicit diagnostic.
func NormalizeAuxiliaryOwnership(ledger AgenticFactLedger) AgenticFactLedger {
	ledger.Sessions = append([]SessionFact(nil), ledger.Sessions...)
	byID := map[string]int{}
	for i, owner := range ledger.Sessions {
		byID[owner.ID] = i
	}
	skills, links := ledger.SkillFacts(), ledger.TaskLinkFacts()
	ledger.Skills, ledger.TaskLinks = nil, nil
	for i := range ledger.Sessions {
		ledger.Sessions[i].Skills = nil
		ledger.Sessions[i].TaskLinks = nil
	}
	for _, fact := range skills {
		i, known := byID[fact.SessionID]
		if known && (fact.Source == "" || fact.Source == ledger.Sessions[i].Source) {
			fact.SessionID, fact.Source = "", ""
			ledger.Sessions[i].Skills = append(ledger.Sessions[i].Skills, fact)
		} else {
			ledger.Skills = append(ledger.Skills, fact)
		}
	}
	for _, fact := range links {
		i, known := byID[fact.SessionID]
		if known && (fact.Source == "" || fact.Source == ledger.Sessions[i].Source) {
			fact.SessionID, fact.Source = "", ""
			ledger.Sessions[i].TaskLinks = append(ledger.Sessions[i].TaskLinks, fact)
		} else {
			ledger.TaskLinks = append(ledger.TaskLinks, fact)
		}
	}
	const diagnostic = "Legacy skill/task facts have unresolved session ownership; recorded totals remain available without session attribution."
	ledger.OwnershipDiagnostics = append([]string(nil), ledger.OwnershipDiagnostics...)
	if len(ledger.Skills)+len(ledger.TaskLinks) > 0 {
		found := false
		for _, note := range ledger.OwnershipDiagnostics {
			if note == diagnostic {
				found = true
			}
		}
		if !found {
			ledger.OwnershipDiagnostics = append(ledger.OwnershipDiagnostics, diagnostic)
		}
	}
	return ledger
}
