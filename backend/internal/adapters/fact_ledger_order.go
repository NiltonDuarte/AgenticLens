package adapters

import (
	"cmp"
	"encoding/json"
	"slices"
	"time"

	"agenticlens/internal/domain"
)

// FactLedgerHistory selects already validated stored evidence whose sequence is
// retained during publication.
type FactLedgerHistory struct {
	SessionIDs       map[string]bool
	IncompleteGroups map[string]bool
	TicketAnomalies  bool
}

// MarshalFactLedger orders the final persisted fact arrays without mutating the
// DTO. Retained sessions participate in the ledger-wide order and keep their
// existing nested evidence. Migration steps already carry their execution order.
func MarshalFactLedger(dto AgenticFactLedgerDTO, history FactLedgerHistory) ([]byte, error) {
	dto.Sessions = slices.Clone(dto.Sessions)
	for i := range dto.Sessions {
		session := &dto.Sessions[i]
		if history.SessionIDs[session.ID] {
			continue
		}
		session.Invocations = chronologicalFacts(session.Invocations, func(invocation InvocationFactDTO) *string { return invocation.StartedAt })
		session.Skills = chronologicalFacts(session.Skills, func(skill SkillFactDTO) *string { return skill.InvokedAt })
		session.Delegations = chronologicalFacts(session.Delegations, func(delegation DelegationFactDTO) *string { return delegation.InvokedAt })
		session.Tools = slices.Clone(session.Tools)
		slices.SortFunc(session.Tools, func(a, b SessionToolFactDTO) int {
			return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Calls, b.Calls), cmp.Compare(a.Errors, b.Errors))
		})
		session.Anomalies = orderedFactAnomalies(session.Anomalies, history.IncompleteGroups[session.ID])
	}
	slices.SortFunc(dto.Sessions, func(a, b SessionFactDTO) int { return cmp.Compare(a.ID, b.ID) })
	if !history.TicketAnomalies {
		dto.Anomalies = orderedFactAnomalies(dto.Anomalies, false)
	}
	return json.Marshal(dto)
}

// chronologicalFacts sorts only timestamped rows. Unknown timestamps retain
// their slots and equivalent instants retain the source's tie order.
func chronologicalFacts[T any](facts []T, timestamp func(T) *string) []T {
	if len(facts) < 2 {
		return facts
	}
	type datedFact struct {
		fact T
		time time.Time
	}
	var dated []datedFact
	var slots []int
	for i, fact := range facts {
		stamp := timestamp(fact)
		if stamp == nil {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, *stamp)
		if err == nil {
			dated = append(dated, datedFact{fact: fact, time: at})
			slots = append(slots, i)
		}
	}
	if len(dated) < 2 {
		return facts
	}
	slices.SortStableFunc(dated, func(a, b datedFact) int { return a.time.Compare(b.time) })
	ordered := slices.Clone(facts)
	for i, slot := range slots {
		ordered[slot] = dated[i].fact
	}
	return ordered
}

func orderedFactAnomalies(anomalies []AnomalyDTO, retainGroupEvidence bool) []AnomalyDTO {
	ordered := slices.Clone(anomalies)
	for i := range ordered {
		if retainGroupEvidence && ordered[i].RuleID == string(domain.RuleSubagentReworkLoop) {
			continue
		}
		ordered[i].Evidence = slices.Clone(ordered[i].Evidence)
		slices.SortFunc(ordered[i].Evidence, compareFactEvidence)
	}
	slices.SortFunc(ordered, compareFactAnomaly)
	return ordered
}

func compareFactAnomaly(a, b AnomalyDTO) int {
	ac, bc := a.Context, b.Context
	if order := cmp.Or(
		cmp.Compare(anomalySeverityOrder(a.Severity), anomalySeverityOrder(b.Severity)),
		cmp.Compare(a.Severity, b.Severity),
		cmp.Compare(a.RuleID, b.RuleID),
		compareOptionalFact(ac.Turn, bc.Turn),
		compareOptionalFact(ac.ToolName, bc.ToolName),
		compareOptionalFact(ac.SessionID, bc.SessionID),
		compareOptionalFact(ac.Role, bc.Role),
		compareOptionalFact(ac.SecondaryRole, bc.SecondaryRole),
	); order != 0 {
		return order
	}
	return cmp.Or(
		slices.CompareFunc(a.Evidence, b.Evidence, compareFactEvidence),
		compareOptionalFact(ac.Count, bc.Count),
		compareOptionalFact(ac.TotalCalls, bc.TotalCalls),
		compareOptionalFact(ac.TokenDelta, bc.TokenDelta),
		compareOptionalFact(ac.ContextTokens, bc.ContextTokens),
		compareOptionalFact(ac.Threshold, bc.Threshold),
		compareOptionalFact(ac.FailureRatePct, bc.FailureRatePct),
	)
}

func anomalySeverityOrder(severity string) int {
	switch domain.AnomalySeverity(severity) {
	case domain.SeverityCritical:
		return 0
	case domain.SeverityWarning:
		return 1
	default:
		return 2
	}
}

func compareOptionalFact[T cmp.Ordered](a, b *T) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return cmp.Compare(*a, *b)
	}
}

func compareFactEvidence(a, b AnomalyEvidenceRefDTO) int {
	return cmp.Or(cmp.Compare(a.SessionID, b.SessionID), cmp.Compare(a.EventID, b.EventID))
}
