package adapters

import (
	"agenticlens/internal/domain"
	"fmt"
	"strings"
)

func validateOutcome(outcome domain.FactOutcome) error {
	switch outcome {
	case domain.OutcomeUnknown, domain.OutcomeSucceeded, domain.OutcomeFailed:
		return nil
	default:
		return fmt.Errorf("unsupported occurrence outcome %q", outcome)
	}
}

func validateOwnedOccurrences(owner SessionFactDTO) error {
	skills := make(map[string]bool)
	tasks := make(map[string]bool)
	for i, fact := range owner.Skills {
		if fact.SessionID != "" || fact.Source != "" {
			return fmt.Errorf("skills[%d] must inherit session/source from its owner", i)
		}
		if strings.TrimSpace(fact.Name) == "" {
			return fmt.Errorf("skills[%d].name is required", i)
		}
		if fact.Outcome == "" {
			return fmt.Errorf("skills[%d].outcome is required", i)
		}
		if err := validateOutcome(fact.Outcome); err != nil {
			return fmt.Errorf("skills[%d]: %w", i, err)
		}
		if fact.Outcome != "" && fact.Success != (fact.Outcome == domain.OutcomeSucceeded) {
			return fmt.Errorf("skills[%d] success contradicts explicit outcome", i)
		}
		if err := uniqueOccurrence(skills, fact.Provenance); err != nil {
			return fmt.Errorf("skills[%d]: %w", i, err)
		}
	}
	for i, fact := range owner.TaskLinks {
		if fact.SessionID != "" || fact.Source != "" {
			return fmt.Errorf("task_links[%d] must inherit session/source from its owner", i)
		}
		if err := validateOutcome(fact.Outcome); err != nil {
			return fmt.Errorf("task_links[%d]: %w", i, err)
		}
		if err := uniqueOccurrence(tasks, fact.Provenance); err != nil {
			return fmt.Errorf("task_links[%d]: %w", i, err)
		}
	}
	return nil
}
func uniqueOccurrence(seen map[string]bool, provenance domain.FactProvenance) error {
	key := ""
	if provenance.CallID != "" {
		key = "call:" + provenance.CallID
	} else if provenance.EventID != "" {
		key = "event:" + provenance.EventID
	}
	if key == "" {
		return nil
	}
	if seen[key] {
		return fmt.Errorf("duplicate native occurrence %q", key)
	}
	seen[key] = true
	return nil
}
