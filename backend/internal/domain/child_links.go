package domain

// ChildParentEvidence is what the records available to one refresh prove about
// a delegation child's native parent.
type ChildParentEvidence string

const (
	// ChildParentConfirmed means the fresh child names this parent, or there is
	// no fresh child and the accepted child record names this parent.
	ChildParentConfirmed ChildParentEvidence = "confirmed"
	// ChildParentContradicted means a fresh child names another parent or none,
	// or the accepted child record names another parent.
	ChildParentContradicted ChildParentEvidence = "contradicted"
	// ChildParentUnobserved means no record exists, or the accepted record has
	// no parent.
	ChildParentUnobserved ChildParentEvidence = "unobserved"
)

// ChildLinkOutcome is the decision taken for one delegation of a refreshed owner.
type ChildLinkOutcome string

const (
	// ChildLinkRetained carries an accepted link onto an empty fresh delegation.
	ChildLinkRetained ChildLinkOutcome = "retained"
	// ChildLinkKeptPendingConfirmation keeps an accepted link over a different
	// fresh child whose parent is not confirmed.
	ChildLinkKeptPendingConfirmation ChildLinkOutcome = "kept_pending_confirmation"
	// ChildLinkReplaced accepts a different fresh child whose parent is confirmed.
	ChildLinkReplaced ChildLinkOutcome = "replaced"
	// ChildLinkCleared removes a fresh link whose child is contradicted.
	ChildLinkCleared ChildLinkOutcome = "cleared"
	// ChildLinkRefusedContradicted keeps an accepted link off the delegation
	// because the accepted child is contradicted.
	ChildLinkRefusedContradicted ChildLinkOutcome = "refused_contradicted"
	// ChildLinkRefusedAmbiguous keeps an accepted link off a delegation whose
	// agent type, description and invocation time are not unique.
	ChildLinkRefusedAmbiguous ChildLinkOutcome = "refused_ambiguous"
	// ChildLinkRefusedHeldBySibling keeps an accepted link off the delegation
	// because another delegation of the same owner already holds the child.
	ChildLinkRefusedHeldBySibling ChildLinkOutcome = "refused_held_by_sibling"
)

// ChildLink is one (parent, child) delegation link.
type ChildLink struct {
	ParentID       string
	ChildSessionID string
}

// ChildLinkDecision records one non-trivial decision. Delegation is the index
// of the delegation within the fresh owner, whose order is preserved.
type ChildLinkDecision struct {
	ParentID       string
	ChildSessionID string
	FreshChildID   string
	Outcome        ChildLinkOutcome
	Evidence       ChildParentEvidence
	Delegation     int
}

// ChildLinkReport lists the decisions of one reconciliation, in owner and
// delegation order. Delegations whose fresh link is kept as read are omitted.
type ChildLinkReport struct {
	Decisions []ChildLinkDecision
}

// RetainedLinks returns the distinct links kept without being re-proved by
// the fresh read, in decision order.
func (report ChildLinkReport) RetainedLinks() []ChildLink {
	seen := make(map[ChildLink]bool)
	links := []ChildLink{}
	for _, decision := range report.Decisions {
		if decision.Outcome != ChildLinkRetained && decision.Outcome != ChildLinkKeptPendingConfirmation {
			continue
		}
		link := ChildLink{ParentID: decision.ParentID, ChildSessionID: decision.ChildSessionID}
		if seen[link] {
			continue
		}
		seen[link] = true
		links = append(links, link)
	}
	return links
}

// ChildLinkInput selects the owners reconciled by RetainAcceptedChildLinks.
// Fresh owners are rewritten. Observed owners are read in the same refresh and
// count as fresh child evidence, but are never rewritten.
type ChildLinkInput struct {
	Harness  SourceHarness
	Accepted []SessionFact
	Fresh    []SessionFact
	Observed []SessionFact
}

// ChildLinkResult holds the reconciled fresh owners, in input order, and the
// decisions taken.
type ChildLinkResult struct {
	Sessions []SessionFact
	Report   ChildLinkReport
}

// RetainAcceptedChildLinks reconciles refreshed owners' delegation child links
// with accepted facts. A refresh that could not re-read a child keeps the
// accepted link on the delegation with the same agent type, description and
// known invocation time, when that key is unique on both sides and no evidence
// contradicts the child's parent. A fresh link to a contradicted child is
// cleared, and a different fresh child replaces an accepted link only when its
// parent is confirmed. Owners of other harnesses and owners with nil
// delegations are returned unchanged. Inputs are never modified, and applying
// the rule to its own result changes nothing.
func RetainAcceptedChildLinks(input ChildLinkInput) ChildLinkResult {
	result := ChildLinkResult{Sessions: append([]SessionFact(nil), input.Fresh...), Report: ChildLinkReport{Decisions: []ChildLinkDecision{}}}
	accepted := make(map[string]SessionFact, len(input.Accepted))
	for _, owner := range input.Accepted {
		if _, exists := accepted[owner.ID]; !exists {
			accepted[owner.ID] = owner
		}
	}
	fresh := make(map[string]SessionFact, len(input.Fresh)+len(input.Observed))
	for _, owner := range input.Observed {
		fresh[owner.ID] = owner
	}
	for _, owner := range input.Fresh {
		fresh[owner.ID] = owner
	}
	evidence := func(child, parent string) ChildParentEvidence {
		if record, ok := fresh[child]; ok {
			if record.ParentID != nil && *record.ParentID == parent {
				return ChildParentConfirmed
			}
			return ChildParentContradicted
		}
		if record, ok := accepted[child]; ok && record.ParentID != nil {
			if *record.ParentID == parent {
				return ChildParentConfirmed
			}
			return ChildParentContradicted
		}
		return ChildParentUnobserved
	}
	for i, owner := range input.Fresh {
		if owner.Source != input.Harness || owner.Delegations == nil {
			continue
		}
		prior, hasPrior := accepted[owner.ID]
		if hasPrior && prior.Source != "" && prior.Source != owner.Source {
			hasPrior = false
		}
		delegations := append([]DelegationFact(nil), owner.Delegations...)
		decisions := reconcileOwnerChildLinks(owner.ID, delegations, prior.Delegations, hasPrior, evidence)
		result.Report.Decisions = append(result.Report.Decisions, decisions...)
		if len(decisions) > 0 {
			owner.Delegations = delegations
			result.Sessions[i] = owner
		}
	}
	return result
}

type childLinkKey struct {
	agentType   string
	description string
	invokedAt   string
	typed       bool
}

func delegationLinkKey(delegation DelegationFact) (childLinkKey, bool) {
	if delegation.InvokedAt == nil || *delegation.InvokedAt == "" {
		return childLinkKey{}, false
	}
	key := childLinkKey{description: delegation.Description, invokedAt: *delegation.InvokedAt}
	if delegation.AgentType != nil {
		key.agentType = *delegation.AgentType
		key.typed = true
	}
	return key, true
}

func countLinkKeys(delegations []DelegationFact) map[childLinkKey]int {
	counts := make(map[childLinkKey]int, len(delegations))
	for _, delegation := range delegations {
		if key, known := delegationLinkKey(delegation); known {
			counts[key]++
		}
	}
	return counts
}

// reconcileOwnerChildLinks rewrites delegations in place. They are the
// caller's private copy of one fresh owner's delegations.
func reconcileOwnerChildLinks(parent string, delegations, prior []DelegationFact, hasPrior bool, evidence func(child, parent string) ChildParentEvidence) []ChildLinkDecision {
	decisions := []ChildLinkDecision{}
	for j := range delegations {
		child := delegations[j].ChildSession
		if child == "" {
			continue
		}
		if state := evidence(child, parent); state == ChildParentContradicted {
			delegations[j].ChildSession = ""
			decisions = append(decisions, ChildLinkDecision{ParentID: parent, FreshChildID: child, Outcome: ChildLinkCleared, Evidence: state, Delegation: j})
		}
	}
	if !hasPrior {
		return decisions
	}
	freshKeys := countLinkKeys(delegations)
	priorByKey := make(map[childLinkKey][]DelegationFact, len(prior))
	for _, delegation := range prior {
		if key, known := delegationLinkKey(delegation); known {
			priorByKey[key] = append(priorByKey[key], delegation)
		}
	}
	// Match each delegation with its accepted link, then decide the outcomes
	// that do not depend on which children other delegations finally hold.
	type candidate struct {
		decision ChildLinkDecision
		accepted string
		movable  bool
	}
	matches := make([]*candidate, len(delegations))
	for j := range delegations {
		key, known := delegationLinkKey(delegations[j])
		if !known {
			continue
		}
		current := delegations[j].ChildSession
		candidates := priorByKey[key]
		if len(candidates) > 1 || freshKeys[key] > 1 {
			// Report only an accepted link that this rule declined to move.
			for _, accepted := range candidates {
				if accepted.ChildSession != "" && accepted.ChildSession != current {
					matches[j] = &candidate{decision: ChildLinkDecision{ParentID: parent, FreshChildID: current, Outcome: ChildLinkRefusedAmbiguous, Delegation: j}}
					break
				}
			}
			continue
		}
		if len(candidates) == 0 || candidates[0].ChildSession == "" || candidates[0].ChildSession == current {
			continue
		}
		match := &candidate{accepted: candidates[0].ChildSession, decision: ChildLinkDecision{ParentID: parent, ChildSessionID: candidates[0].ChildSession, FreshChildID: current, Delegation: j}}
		match.decision.Evidence = evidence(match.accepted, parent)
		switch {
		case match.decision.Evidence == ChildParentContradicted:
			match.decision.Outcome = ChildLinkRefusedContradicted
		case current != "" && evidence(current, parent) == ChildParentConfirmed:
			match.decision.Outcome = ChildLinkReplaced
			match.decision.Evidence = ChildParentConfirmed
			match.decision.ChildSessionID = current
		default:
			match.movable = true
		}
		matches[j] = match
	}
	// Movable delegations take their accepted child while no other delegation
	// finally holds it. Each delegation moves at most once and a move only
	// frees its fresh child, so this fixpoint ends within len(delegations)
	// rounds. Every unmoved candidate's child is then held in the result, so
	// applying the rule to that result moves nothing more.
	held := make(map[string]int, len(delegations))
	for _, delegation := range delegations {
		if delegation.ChildSession != "" {
			held[delegation.ChildSession]++
		}
	}
	for changed := true; changed; {
		changed = false
		for j, match := range matches {
			if match == nil || !match.movable || delegations[j].ChildSession == match.accepted || held[match.accepted] > 0 {
				continue
			}
			if previous := delegations[j].ChildSession; previous != "" {
				held[previous]--
			}
			delegations[j].ChildSession = match.accepted
			held[match.accepted]++
			changed = true
		}
	}
	for j, match := range matches {
		if match == nil {
			continue
		}
		if match.movable {
			switch {
			case delegations[j].ChildSession != match.accepted:
				match.decision.Outcome = ChildLinkRefusedHeldBySibling
			case match.decision.FreshChildID == "":
				match.decision.Outcome = ChildLinkRetained
			default:
				match.decision.Outcome = ChildLinkKeptPendingConfirmation
			}
		}
		decisions = append(decisions, match.decision)
	}
	return decisions
}
