package migration

import (
	"agenticlens/internal/domain"
	"fmt"
)

// StepV6ToV7LeanOccurrences consumes old native relationship metadata exactly
// once and audits discarded fields without retaining their payloads.
type StepV6ToV7LeanOccurrences struct{}

func NewStepV6ToV7LeanOccurrences() *StepV6ToV7LeanOccurrences { return &StepV6ToV7LeanOccurrences{} }
func (s *StepV6ToV7LeanOccurrences) SourceVersion() int        { return 6 }
func (s *StepV6ToV7LeanOccurrences) TargetVersion() int        { return 7 }
func (s *StepV6ToV7LeanOccurrences) Description() string {
	return "Converts owned skills and delegations to lean schema v7 and discards persisted traces"
}
func (s *StepV6ToV7LeanOccurrences) Apply(ledger legacyLedger) (legacyLedger, error) {
	result := ledger
	result.Sessions = append([]legacySessionFact{}, ledger.Sessions...)
	byID := make(map[string]legacySessionFact)
	children := make(map[string][]string)
	for _, owner := range ledger.Sessions {
		byID[owner.ID] = owner
		if owner.ParentID != nil && owner.AgentPath != "" {
			key := *owner.ParentID + "\x00" + owner.AgentPath
			children[key] = append(children[key], owner.ID)
		}
	}
	skills, delegations, events, markers, removed, unresolved := 0, 0, 0, 0, 0, 0
	for i := range result.Sessions {
		owner := &result.Sessions[i]
		owner.TaskLinks = append([]legacyTaskLinkFact{}, owner.TaskLinks...)
		skills += len(owner.Skills)
		delegations += len(owner.TaskLinks)
		events += len(owner.Events)
		if owner.partialMarkerPresent || owner.EventsArePartial {
			markers++
		}
		if owner.agentPathPresent || owner.AgentPath != "" {
			removed++
		}
		for _, skill := range owner.Skills {
			removed += skill.removedFields
		}
		for j := range owner.TaskLinks {
			link := &owner.TaskLinks[j]
			removed += link.removedFields
			hadChildReference := link.ChildSession != "" || link.ChildReference != "" || link.ChildAgentID != ""
			if link.ChildSession == "" && link.ChildAgentID != "" {
				if child, ok := byID[link.ChildAgentID]; ok && child.ParentID != nil && *child.ParentID == owner.ID {
					link.ChildSession = child.ID
				}
			}
			if link.ChildSession == "" && link.ChildReference != "" {
				matches := children[owner.ID+"\x00"+link.ChildReference]
				if len(matches) == 1 {
					link.ChildSession = matches[0]
				}
			}
			if child, known := byID[link.ChildSession]; known && (child.ParentID == nil || *child.ParentID != owner.ID) {
				link.ChildSession = ""
			}
			if link.ChildSession == "" && hadChildReference {
				unresolved++
			}
			// Preserve only the values needed by the final lean projection.
			*link = legacyTaskLinkFact{SubagentType: link.SubagentType, Description: link.Description, Timestamp: link.Timestamp, ChildSession: link.ChildSession}
		}
		owner.Skills = append([]legacySkillFact{}, owner.Skills...)
		for j, skill := range owner.Skills {
			owner.Skills[j] = legacySkillFact{Name: skill.Name, CallerRole: skill.CallerRole, Timestamp: skill.Timestamp}
		}
		owner.Events = nil
		owner.EventsArePartial = false
		owner.AgentPath = ""
	}
	version := 7
	result.SchemaVersion = &version
	info := domain.MigrationInfo{Status: domain.MigrationStatusCompleted, StepsPending: []string{}}
	if ledger.Migration != nil {
		info = *ledger.Migration
	}
	info.SchemaVersion = 7
	info.StepsCompleted = append(append([]string{}, info.StepsCompleted...), "schema_v7_lean_occurrences")
	info.StepsPending = append([]string{}, info.StepsPending...)
	if info.Notes != "" {
		info.Notes += "; "
	}
	info.Notes += fmt.Sprintf("Schema v7 migration converted skills=%d delegations=%d; discarded events=%d partial_markers=%d rich_fields=%d unresolved_child_references=%d", skills, delegations, events, markers, removed, unresolved)
	result.Migration = &info
	return result, nil
}
