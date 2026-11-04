package migration

import (
	"agenticlens/internal/domain"
	"fmt"
)

// StepV5ToV6OwnedOccurrences explicitly drops ownerless legacy rows. It never
// guesses ownership from a root, role, model, time or local turn number.
type StepV5ToV6OwnedOccurrences struct{}

func NewStepV5ToV6OwnedOccurrences() *StepV5ToV6OwnedOccurrences {
	return &StepV5ToV6OwnedOccurrences{}
}
func (s *StepV5ToV6OwnedOccurrences) SourceVersion() int { return 5 }
func (s *StepV5ToV6OwnedOccurrences) TargetVersion() int { return 6 }
func (s *StepV5ToV6OwnedOccurrences) Description() string {
	return "Moves reliably attributed occurrences and drops ownerless legacy skill/task rows"
}
func (s *StepV5ToV6OwnedOccurrences) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	result := ledger
	result.Sessions = append([]domain.SessionFact(nil), ledger.Sessions...)
	byID := make(map[string]int)
	for i := range result.Sessions {
		byID[result.Sessions[i].ID] = i
		result.Sessions[i].Skills = append([]domain.SkillFact(nil), result.Sessions[i].Skills...)
		for j := range result.Sessions[i].Skills {
			normalizeLegacySkill(&result.Sessions[i].Skills[j])
		}
		result.Sessions[i].TaskLinks = append([]domain.TaskLinkFact(nil), result.Sessions[i].TaskLinks...)
		for j := range result.Sessions[i].TaskLinks {
			if result.Sessions[i].TaskLinks[j].Outcome == "" {
				result.Sessions[i].TaskLinks[j].Outcome = domain.OutcomeUnknown
			}
		}
	}
	droppedSkills, droppedTasks := 0, 0
	for _, fact := range ledger.Skills {
		i, ok := byID[fact.SessionID]
		if !ok || fact.SessionID == "" || fact.Source != "" && fact.Source != result.Sessions[i].Source {
			droppedSkills++
			continue
		}
		fact.SessionID, fact.Source = "", ""
		normalizeLegacySkill(&fact)
		duplicate := false
		for _, previous := range result.Sessions[i].Skills {
			if sameNativeOccurrence(previous.Provenance, fact.Provenance) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result.Sessions[i].Skills = append(result.Sessions[i].Skills, fact)
		}
	}
	for _, fact := range ledger.TaskLinks {
		if fact.Outcome == "" {
			fact.Outcome = domain.OutcomeUnknown
		}
		i, ok := byID[fact.SessionID]
		if !ok || fact.SessionID == "" || fact.Source != "" && fact.Source != result.Sessions[i].Source {
			droppedTasks++
			continue
		}
		fact.SessionID, fact.Source = "", ""
		duplicate := false
		for _, previous := range result.Sessions[i].TaskLinks {
			if sameNativeOccurrence(previous.Provenance, fact.Provenance) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result.Sessions[i].TaskLinks = append(result.Sessions[i].TaskLinks, fact)
		}
	}
	result.Skills, result.TaskLinks = nil, nil
	version := 6
	result.SchemaVersion = &version
	info := domain.MigrationInfo{Status: domain.MigrationStatusCompleted, StepsPending: []string{}}
	if ledger.Migration != nil {
		info = *ledger.Migration
	}
	info.SchemaVersion = 6
	info.StepsCompleted = append(append([]string{}, info.StepsCompleted...), "schema_v6_owned_occurrences")
	info.StepsPending = append([]string{}, info.StepsPending...)
	if info.Notes != "" {
		info.Notes += "; "
	}
	info.Notes += fmt.Sprintf("Schema v6 migration dropped ownerless skills=%d task_links=%d", droppedSkills, droppedTasks)
	result.Migration = &info
	return result, nil
}
func sameNativeOccurrence(a, b domain.FactProvenance) bool {
	if a.CallID != "" && b.CallID != "" {
		return a.CallID == b.CallID
	}
	if a.ItemID != "" && b.ItemID != "" {
		return a.ItemID == b.ItemID
	}
	return a.EventID != "" && b.EventID != "" && a.EventID == b.EventID
}

func normalizeLegacySkill(fact *domain.SkillFact) {
	if fact.Outcome == "" {
		fact.Outcome = domain.OutcomeFailed
		if fact.Success {
			fact.Outcome = domain.OutcomeSucceeded
		}
	}
}
