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
	// ChildParentNotEvaluated means the rule evaluated no child, because the
	// delegation's provenance is ambiguous.
	ChildParentNotEvaluated ChildParentEvidence = "not_evaluated"
)

// ChildLinkOutcome is the decision taken for one delegation of a refreshed owner.
type ChildLinkOutcome string

const (
	// ChildLinkRetained carries an accepted link onto an empty fresh delegation
	// whose agent type, description and known invocation time are unique.
	ChildLinkRetained ChildLinkOutcome = "retained"
	// ChildLinkRetainedByPosition carries an accepted link onto the fresh
	// delegation at the same position among delegations that share agent type,
	// description and invocation time, known or not, when every fresh link of
	// that group is empty. Codex transcripts only append, so the accepted group
	// is a prefix of the fresh one.
	ChildLinkRetainedByPosition ChildLinkOutcome = "retained_by_position"
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
	// ChildLinkRefusedAmbiguous keeps accepted links off a delegation whose
	// agent type, description and invocation time are shared, when a fresh link
	// of that group is already set or the fresh group is smaller.
	ChildLinkRefusedAmbiguous ChildLinkOutcome = "refused_ambiguous"
	// ChildLinkRefusedHeldBySibling keeps an accepted link off the delegation
	// because another delegation of the same owner already holds the child.
	ChildLinkRefusedHeldBySibling ChildLinkOutcome = "refused_held_by_sibling"
)

// ChildLinkRefusal is why an accepted link that no evidence contradicts was
// not carried onto an empty fresh delegation.
type ChildLinkRefusal string

const (
	// ChildLinkRefusalAmbiguous means delegations share the link's agent type,
	// description and known invocation time, and the group could not be
	// aligned by position: a fresh link is set or the fresh group is smaller.
	ChildLinkRefusalAmbiguous ChildLinkRefusal = "ambiguous"
	// ChildLinkRefusalUnknownProvenance means the accepted delegation has no
	// known invocation time and its group could not be aligned by position.
	ChildLinkRefusalUnknownProvenance ChildLinkRefusal = "unknown_provenance"
	// ChildLinkRefusalHeldBySibling means another delegation of the owner holds
	// the accepted child.
	ChildLinkRefusalHeldBySibling ChildLinkRefusal = "held_by_sibling"
)

// RefusedChildLink is an accepted link the rule declined to carry. It is never
// written to a delegation; it only keeps the refresh visibly incomplete.
type RefusedChildLink struct {
	Link   ChildLink
	Reason ChildLinkRefusal
}

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
// Refused lists the accepted links, not contradicted, that the rule declined
// to carry while their child is not linked from the delegation they had.
type ChildLinkReport struct {
	Decisions []ChildLinkDecision
	Refused   []RefusedChildLink
}

// RetainedLinks returns the distinct links kept without being re-proved by
// the fresh read, in decision order.
func (report ChildLinkReport) RetainedLinks() []ChildLink {
	seen := make(map[ChildLink]bool)
	links := []ChildLink{}
	for _, decision := range report.Decisions {
		if decision.Outcome != ChildLinkRetained && decision.Outcome != ChildLinkRetainedByPosition && decision.Outcome != ChildLinkKeptPendingConfirmation {
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

// UnresolvedLinks returns the distinct accepted links this refresh did not
// re-prove: the links it retained and the links it refused to carry. Each is
// unresolved child evidence. Decision order comes first, then refusals.
func (report ChildLinkReport) UnresolvedLinks() []ChildLink {
	links := report.RetainedLinks()
	seen := make(map[ChildLink]bool, len(links)+len(report.Refused))
	for _, link := range links {
		seen[link] = true
	}
	for _, refused := range report.Refused {
		if !seen[refused.Link] {
			seen[refused.Link] = true
			links = append(links, refused.Link)
		}
	}
	return links
}

// ChildLinkInput selects the owners reconciled by RetainAcceptedChildLinks.
// Fresh owners are rewritten. Observed owners are read in the same refresh and
// count as fresh child evidence, but are never rewritten. Known owners are
// accepted records of another store, such as the capture registry when
// Accepted is the committed ledger; for child evidence they take precedence
// over Accepted, and a Known record without a parent is unobserved.
type ChildLinkInput struct {
	Harness  SourceHarness
	Accepted []SessionFact
	Known    []SessionFact
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
// contradicts the child's parent. Delegations sharing a key, or with an
// unknown invocation time, keep accepted links by position when every fresh
// link of the group is empty and the fresh group is at least as large. A fresh link to a contradicted child is
// cleared, and a different fresh child replaces an accepted link only when its
// parent is confirmed. Owners of other harnesses and owners with nil
// delegations are returned unchanged. Inputs are never modified, and applying
// the rule to its own result changes nothing.
func RetainAcceptedChildLinks(input ChildLinkInput) ChildLinkResult {
	result := ChildLinkResult{Sessions: append([]SessionFact(nil), input.Fresh...), Report: ChildLinkReport{Decisions: []ChildLinkDecision{}, Refused: []RefusedChildLink{}}}
	accepted := make(map[string]SessionFact, len(input.Accepted))
	for _, owner := range input.Accepted {
		if _, exists := accepted[owner.ID]; !exists {
			accepted[owner.ID] = owner
		}
	}
	known := make(map[string]SessionFact, len(input.Known))
	for _, owner := range input.Known {
		if _, exists := known[owner.ID]; !exists {
			known[owner.ID] = owner
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
		record, ok := known[child]
		if !ok {
			record, ok = accepted[child]
		}
		if ok && record.ParentID != nil {
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
		decisions, refused := reconcileOwnerChildLinks(owner.ID, delegations, prior.Delegations, hasPrior, evidence)
		result.Report.Decisions = append(result.Report.Decisions, decisions...)
		result.Report.Refused = append(result.Report.Refused, refused...)
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
	timed       bool
}

// delegationLinkKey groups delegations by agent type, description and
// invocation time. An unknown invocation time is its own key value.
func delegationLinkKey(delegation DelegationFact) childLinkKey {
	key := childLinkKey{description: delegation.Description}
	if delegation.InvokedAt != nil && *delegation.InvokedAt != "" {
		key.invokedAt = *delegation.InvokedAt
		key.timed = true
	}
	if delegation.AgentType != nil {
		key.agentType = *delegation.AgentType
		key.typed = true
	}
	return key
}

// childLinkGroups returns delegation positions by key, in transcript order,
// and the keys in order of their first delegation.
func childLinkGroups(delegations []DelegationFact) (map[childLinkKey][]int, []childLinkKey) {
	groups := make(map[childLinkKey][]int, len(delegations))
	order := []childLinkKey{}
	for i, delegation := range delegations {
		key := delegationLinkKey(delegation)
		if _, seen := groups[key]; !seen {
			order = append(order, key)
		}
		groups[key] = append(groups[key], i)
	}
	return groups, order
}

// childLinkAlignment is how one key group's accepted links were matched.
type childLinkAlignment int

const (
	alignmentNone childLinkAlignment = iota
	alignmentUnique
	alignmentPosition
	alignmentRefused
)

// reconcileOwnerChildLinks rewrites delegations in place. They are the
// caller's private copy of one fresh owner's delegations.
func reconcileOwnerChildLinks(parent string, delegations, prior []DelegationFact, hasPrior bool, evidence func(child, parent string) ChildParentEvidence) ([]ChildLinkDecision, []RefusedChildLink) {
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
		return decisions, nil
	}
	freshGroups, order := childLinkGroups(delegations)
	priorGroups, _ := childLinkGroups(prior)
	// Match accepted links group by group, then decide the outcomes that do
	// not depend on which children other delegations finally hold.
	type candidate struct {
		decision ChildLinkDecision
		accepted string
		movable  bool
	}
	matches := make([]*candidate, len(delegations))
	alignments := make(map[childLinkKey]childLinkAlignment, len(order))
	for _, key := range order {
		fresh, accepted := freshGroups[key], priorGroups[key]
		if len(accepted) == 0 {
			continue
		}
		if key.timed && len(fresh) == 1 && len(accepted) == 1 {
			alignments[key] = alignmentUnique
			j, match := fresh[0], prior[accepted[0]].ChildSession
			current := delegations[j].ChildSession
			if match == "" || match == current {
				continue
			}
			entry := &candidate{accepted: match, decision: ChildLinkDecision{ParentID: parent, ChildSessionID: match, FreshChildID: current, Delegation: j}}
			entry.decision.Evidence = evidence(match, parent)
			switch {
			case entry.decision.Evidence == ChildParentContradicted:
				entry.decision.Outcome = ChildLinkRefusedContradicted
			case current != "" && evidence(current, parent) == ChildParentConfirmed:
				entry.decision.Outcome = ChildLinkReplaced
				entry.decision.Evidence = ChildParentConfirmed
				entry.decision.ChildSessionID = current
			default:
				entry.movable = true
			}
			matches[j] = entry
			continue
		}
		empty := true
		for _, j := range fresh {
			empty = empty && delegations[j].ChildSession == ""
		}
		if !empty || len(fresh) < len(accepted) {
			alignments[key] = alignmentRefused
			if !key.timed {
				continue
			}
			// Report only accepted links this rule declined to move.
			for _, j := range fresh {
				current := delegations[j].ChildSession
				for _, i := range accepted {
					if child := prior[i].ChildSession; child != "" && child != current {
						matches[j] = &candidate{decision: ChildLinkDecision{ParentID: parent, FreshChildID: current, Outcome: ChildLinkRefusedAmbiguous, Evidence: ChildParentNotEvaluated, Delegation: j}}
						break
					}
				}
			}
			continue
		}
		// Transcripts only append: accepted position i is fresh position i.
		alignments[key] = alignmentPosition
		for position, i := range accepted {
			match := prior[i].ChildSession
			if match == "" {
				continue
			}
			j := fresh[position]
			entry := &candidate{accepted: match, decision: ChildLinkDecision{ParentID: parent, ChildSessionID: match, Delegation: j}}
			entry.decision.Evidence = evidence(match, parent)
			if entry.decision.Evidence == ChildParentContradicted {
				entry.decision.Outcome = ChildLinkRefusedContradicted
			} else {
				entry.movable = true
			}
			matches[j] = entry
		}
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
			case alignments[delegationLinkKey(delegations[j])] == alignmentPosition:
				match.decision.Outcome = ChildLinkRetainedByPosition
			case match.decision.FreshChildID == "":
				match.decision.Outcome = ChildLinkRetained
			default:
				match.decision.Outcome = ChildLinkKeptPendingConfirmation
			}
		}
		decisions = append(decisions, match.decision)
	}
	return decisions, refusedChildLinks(parent, delegations, prior, decisions, alignments, evidence)
}

// refusedChildLinks lists the accepted links, not contradicted, that the rule
// declined to carry onto an empty fresh delegation: a child held by a sibling,
// or a key group that could not be aligned by position. Unaligned links count
// only while no final delegation links their child.
func refusedChildLinks(parent string, delegations, prior []DelegationFact, decisions []ChildLinkDecision, alignments map[childLinkKey]childLinkAlignment, evidence func(child, parent string) ChildParentEvidence) []RefusedChildLink {
	refused := []RefusedChildLink{}
	seen := map[ChildLink]bool{}
	add := func(child string, reason ChildLinkRefusal) {
		link := ChildLink{ParentID: parent, ChildSessionID: child}
		if child == "" || seen[link] || evidence(child, parent) == ChildParentContradicted {
			return
		}
		seen[link] = true
		refused = append(refused, RefusedChildLink{Link: link, Reason: reason})
	}
	for _, decision := range decisions {
		if decision.Outcome == ChildLinkRefusedHeldBySibling && decision.FreshChildID == "" {
			add(decision.ChildSessionID, ChildLinkRefusalHeldBySibling)
		}
	}
	linked := make(map[string]bool, len(delegations))
	for _, delegation := range delegations {
		linked[delegation.ChildSession] = true
	}
	for _, delegation := range prior {
		key := delegationLinkKey(delegation)
		if delegation.ChildSession == "" || linked[delegation.ChildSession] || alignments[key] != alignmentRefused {
			continue
		}
		if key.timed {
			add(delegation.ChildSession, ChildLinkRefusalAmbiguous)
		} else {
			add(delegation.ChildSession, ChildLinkRefusalUnknownProvenance)
		}
	}
	return refused
}
