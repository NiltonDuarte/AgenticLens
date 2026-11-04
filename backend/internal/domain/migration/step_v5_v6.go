package migration

import (
	"agenticlens/internal/domain"
	"fmt"
)

// StepV5ToV6LeanOccurrences relocates proven owners and produces the sole new
// lean schema in one transition. Ownerless drops and retired payload removal
// have independent audit counts; no interim rich schema is accepted at runtime.
type StepV5ToV6LeanOccurrences struct{}

func NewStepV5ToV6LeanOccurrences() *StepV5ToV6LeanOccurrences { return &StepV5ToV6LeanOccurrences{} }
func (s *StepV5ToV6LeanOccurrences) SourceVersion() int        { return 5 }
func (s *StepV5ToV6LeanOccurrences) TargetVersion() int        { return 6 }
func (s *StepV5ToV6LeanOccurrences) Description() string {
	return "Creates owned lean skills and delegations, drops ownerless rows and retires persisted traces"
}
func (s *StepV5ToV6LeanOccurrences) Apply(ledger legacyLedger) (legacyLedger, error) {
	result, droppedSkills, droppedTasks := relocateLegacyOwners(ledger)
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
		for j, skill := range owner.Skills {
			owner.Skills[j] = legacySkillFact{Name: skill.Name, CallerRole: skill.CallerRole, Timestamp: skill.Timestamp}
		}
		owner.Events = nil
		owner.EventsArePartial = false
		owner.AgentPath = ""
	}

	result.Skills, result.TaskLinks = nil, nil
	version := 6
	result.SchemaVersion = &version
	info := domain.MigrationInfo{Status: domain.MigrationStatusCompleted, StepsPending: []string{}}
	if ledger.Migration != nil {
		info = *ledger.Migration
	}
	info.SchemaVersion = 6
	info.StepsCompleted = append(append([]string{}, info.StepsCompleted...), "schema_v6_lean_occurrences")
	info.StepsPending = append([]string{}, info.StepsPending...)
	if info.Notes != "" {
		info.Notes += "; "
	}
	info.Notes += fmt.Sprintf("Schema v6 migration dropped ownerless skills=%d task_links=%d; converted skills=%d delegations=%d; discarded events=%d partial_markers=%d rich_fields=%d unresolved_child_references=%d", droppedSkills, droppedTasks, skills, delegations, events, markers, removed, unresolved)
	result.Migration = &info
	return result, nil
}

// relocateLegacyOwners uses recorded native session identity only. Native
// occurrence IDs reconcile duplicate representations; equal rows without an ID
// retain their multiplicity.
func relocateLegacyOwners(ledger legacyLedger) (legacyLedger, int, int) {
	result := ledger
	result.Sessions = append([]legacySessionFact{}, ledger.Sessions...)
	byID := make(map[string]int)
	for i := range result.Sessions {
		byID[result.Sessions[i].ID] = i
		result.Sessions[i].Skills = append([]legacySkillFact(nil), result.Sessions[i].Skills...)
		result.Sessions[i].TaskLinks = append([]legacyTaskLinkFact(nil), result.Sessions[i].TaskLinks...)
	}
	droppedSkills, droppedTasks := 0, 0
	for _, fact := range ledger.Skills {
		i, ok := byID[fact.SessionID]
		if !ok || fact.SessionID == "" || fact.Source != "" && fact.Source != result.Sessions[i].Source {
			droppedSkills++
			continue
		}
		fact.SessionID, fact.Source = "", ""
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

	return result, droppedSkills, droppedTasks
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
