package adapters

import (
	"agenticlens/internal/domain"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SkillFactDTO domain.SkillFact
type DelegationFactDTO domain.DelegationFact

func validateOccurrenceShape(data []byte, allowed, required []string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		found := false
		for _, k := range allowed {
			if key == k {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("retired or unsupported occurrence field %q", key)
		}
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("occurrence requires %s (null is supported where unknown)", key)
		}
	}
	for _, key := range []string{"description", "child_session_id"} {
		if raw, ok := fields[key]; ok {
			var value *string
			if err := json.Unmarshal(raw, &value); err != nil || value == nil || strings.TrimSpace(*value) == "" {
				return fmt.Errorf("optional %s must be a nonempty string when present", key)
			}
		}
	}
	return nil
}
func (dto *SkillFactDTO) UnmarshalJSON(data []byte) error {
	if err := validateOccurrenceShape(data, []string{"name", "caller_role", "invoked_at"}, []string{"name", "caller_role", "invoked_at"}); err != nil {
		return err
	}
	type alias SkillFactDTO
	var value alias
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*dto = SkillFactDTO(value)
	return validateSkill(domain.SkillFact(*dto))
}
func (dto *DelegationFactDTO) UnmarshalJSON(data []byte) error {
	if err := validateOccurrenceShape(data, []string{"agent_type", "description", "invoked_at", "child_session_id"}, []string{"agent_type", "invoked_at"}); err != nil {
		return err
	}
	type alias DelegationFactDTO
	var value alias
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*dto = DelegationFactDTO(value)
	return validateDelegation(domain.DelegationFact(*dto))
}
func validateOccurrenceTime(value *string) error {
	if value == nil {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, *value); err != nil {
		return fmt.Errorf("invalid invoked_at: %w", err)
	}
	return nil
}
func validateSkill(fact domain.SkillFact) error {
	if strings.TrimSpace(fact.Name) == "" {
		return fmt.Errorf("skill name is required")
	}
	if fact.CallerRole != nil && strings.TrimSpace(*fact.CallerRole) == "" {
		return fmt.Errorf("caller_role must be nonempty or null")
	}
	return validateOccurrenceTime(fact.InvokedAt)
}
func validateDelegation(fact domain.DelegationFact) error {
	if fact.Description != "" && strings.TrimSpace(fact.Description) == "" {
		return fmt.Errorf("description must be nonempty when present")
	}
	if fact.ChildSession != "" && strings.TrimSpace(fact.ChildSession) == "" {
		return fmt.Errorf("child_session_id must be nonempty when present")
	}
	if fact.AgentType != nil && strings.TrimSpace(*fact.AgentType) == "" {
		return fmt.Errorf("agent_type must be nonempty or null")
	}
	return validateOccurrenceTime(fact.InvokedAt)
}
func validateOwnedOccurrences(owner SessionFactDTO) error {
	for i, fact := range owner.Skills {
		if err := validateSkill(domain.SkillFact(fact)); err != nil {
			return fmt.Errorf("skills[%d]: %w", i, err)
		}
	}
	for i, fact := range owner.Delegations {
		if err := validateDelegation(domain.DelegationFact(fact)); err != nil {
			return fmt.Errorf("delegations[%d]: %w", i, err)
		}
	}
	return nil
}
func skillFactDTOsToDomain(values []SkillFactDTO) []domain.SkillFact {
	if values == nil {
		return nil
	}
	result := make([]domain.SkillFact, len(values))
	for i, v := range values {
		result[i] = domain.SkillFact(v)
	}
	return result
}
func skillFactDTOsFromDomain(values []domain.SkillFact) []SkillFactDTO {
	if values == nil {
		return nil
	}
	result := make([]SkillFactDTO, len(values))
	for i, v := range values {
		result[i] = SkillFactDTO(v)
	}
	return result
}
func delegationFactDTOsToDomain(values []DelegationFactDTO) []domain.DelegationFact {
	if values == nil {
		return nil
	}
	result := make([]domain.DelegationFact, len(values))
	for i, v := range values {
		result[i] = domain.DelegationFact(v)
	}
	return result
}
func delegationFactDTOsFromDomain(values []domain.DelegationFact) []DelegationFactDTO {
	if values == nil {
		return nil
	}
	result := make([]DelegationFactDTO, len(values))
	for i, v := range values {
		result[i] = DelegationFactDTO(v)
	}
	return result
}
