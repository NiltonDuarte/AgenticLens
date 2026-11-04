package domain

import "strings"

// SkillFacts preserves owner order and legitimate equal-looking occurrences.
func (ledger AgenticFactLedger) SkillFacts() []SkillFact {
	result := []SkillFact{}
	for _, session := range ledger.Sessions {
		result = append(result, session.Skills...)
	}
	return result
}
func (ledger AgenticFactLedger) DelegationFacts() []DelegationFact {
	result := []DelegationFact{}
	for _, session := range ledger.Sessions {
		result = append(result, session.Delegations...)
	}
	return result
}

// KnownString preserves unknown source values as explicit nulls.
func KnownString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
