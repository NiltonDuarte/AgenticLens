package domain_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

// FIX-21 child link retention. The rule keeps an accepted Codex delegation
// link that a bounded refresh could not re-prove, and lets only fresh source
// evidence clear or replace it.

func linkString(value string) *string { return &value }

func linkDelegation(agentType, description, invokedAt, child string) domain.DelegationFact {
	delegation := domain.DelegationFact{Description: description, ChildSession: child}
	if agentType != "" {
		delegation.AgentType = linkString(agentType)
	}
	if invokedAt != "" {
		delegation.InvokedAt = linkString(invokedAt)
	}
	return delegation
}

func linkOwner(id, parent string, delegations ...domain.DelegationFact) domain.SessionFact {
	owner := domain.SessionFact{ID: id, Source: domain.SourceCodex, Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}}
	if parent != "" {
		owner.ParentID = linkString(parent)
	}
	if delegations != nil {
		owner.Delegations = delegations
	}
	return owner
}

func ownerLinks(sessions []domain.SessionFact, id string) []string {
	for _, session := range sessions {
		if session.ID != id {
			continue
		}
		links := []string{}
		for _, delegation := range session.Delegations {
			links = append(links, delegation.ChildSession)
		}
		return links
	}
	return nil
}

func outcomes(report domain.ChildLinkReport) []domain.ChildLinkOutcome {
	result := []domain.ChildLinkOutcome{}
	for _, decision := range report.Decisions {
		result = append(result, decision.Outcome)
	}
	return result
}

const (
	at1 = "2026-11-24T09:00:07.000Z"
	at2 = "2026-11-24T09:00:30.000Z"
)

func TestRetainAcceptedChildLinks(t *testing.T) {
	first := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-01", at1, child) }
	second := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-02", at2, child) }
	for _, tc := range []struct {
		name      string
		accepted  []domain.SessionFact
		fresh     []domain.SessionFact
		observed  []domain.SessionFact
		wantLinks []string
		want      []domain.ChildLinkOutcome
		evidence  domain.ChildParentEvidence
		retained  []domain.ChildLink
	}{
		{
			name:      "absent child is retained",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"), second("C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), second("C2"))},
			wantLinks: []string{"C1", "C2"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetained},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "confirmed child whose fresh link is empty is retained",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("")), linkOwner("C1", "P")},
			wantLinks: []string{"C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetained},
			evidence:  domain.ChildParentConfirmed,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "observed child read in the same refresh confirms",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""))},
			observed:  []domain.SessionFact{linkOwner("C1", "P")},
			wantLinks: []string{"C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetained},
			evidence:  domain.ChildParentConfirmed,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "fresh child naming another parent clears the link and refuses the accepted one",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"), second("C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("C1"), second("")), linkOwner("C1", "Q")},
			wantLinks: []string{"", "C2"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkCleared, domain.ChildLinkRefusedContradicted, domain.ChildLinkRetained},
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C2"}},
		},
		{
			name:      "fresh child with no parent clears the link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("C1")), linkOwner("C1", "")},
			wantLinks: []string{""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkCleared, domain.ChildLinkRefusedContradicted},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "fresh empty link to a fresh contradicting child is not retained",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("")), linkOwner("C1", "Q")},
			wantLinks: []string{""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRefusedContradicted},
			evidence:  domain.ChildParentContradicted,
			retained:  []domain.ChildLink{},
		},
		{
			name:      "accepted child record naming another parent is not retained",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1")), linkOwner("C1", "Q")},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""))},
			wantLinks: []string{""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRefusedContradicted},
			evidence:  domain.ChildParentContradicted,
			retained:  []domain.ChildLink{},
		},
		{
			name:      "accepted child record without a parent is unobserved and retained",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1")), linkOwner("C1", "")},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""))},
			wantLinks: []string{"C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetained},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "different fresh child confirmed by its fresh record replaces the link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("C9")), linkOwner("C9", "P")},
			wantLinks: []string{"C9"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkReplaced},
			evidence:  domain.ChildParentConfirmed,
			retained:  []domain.ChildLink{},
		},
		{
			name:      "different fresh child confirmed only by accepted metadata replaces the link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1")), linkOwner("C9", "P")},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("C9"))},
			wantLinks: []string{"C9"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkReplaced},
			evidence:  domain.ChildParentConfirmed,
			retained:  []domain.ChildLink{},
		},
		{
			name:      "different fresh child without confirming metadata keeps the accepted link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first("C9"))},
			wantLinks: []string{"C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkKeptPendingConfirmation},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "duplicate provenance is carried by position",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks: []string{"C1", "C2"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetainedByPosition, domain.ChildLinkRetainedByPosition},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}, {ParentID: "P", ChildSessionID: "C2"}},
		},
		{
			name:      "a larger fresh group is carried by position",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks: []string{"C1", ""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetainedByPosition},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "unknown invocation time is carried by position",
			accepted:  []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "task-01", "", "C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "task-01", "", ""))},
			wantLinks: []string{"C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRetainedByPosition},
			evidence:  domain.ChildParentUnobserved,
			retained:  []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}},
		},
		{
			name:      "a different description or agent type is another delegation",
			accepted:  []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "task-01", at1, "C1"), linkDelegation("", "task-02", at2, "C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "renamed", at1, ""), linkDelegation("worker", "task-02", at2, ""))},
			wantLinks: []string{"", ""},
			want:      []domain.ChildLinkOutcome{},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "a child already held by a sibling delegation is not moved",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"), second("C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), second("C1"))},
			observed:  []domain.SessionFact{linkOwner("C1", "P")},
			wantLinks: []string{"", "C1"},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRefusedHeldBySibling, domain.ChildLinkReplaced},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "nil fresh delegations are untouched",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "")},
			wantLinks: []string{},
			want:      []domain.ChildLinkOutcome{},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "an owner without an accepted record keeps its fresh links",
			accepted:  []domain.SessionFact{linkOwner("Q", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), second("C2"))},
			wantLinks: []string{"", "C2"},
			want:      []domain.ChildLinkOutcome{},
			retained:  []domain.ChildLink{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, err := json.Marshal([][]domain.SessionFact{tc.accepted, tc.fresh, tc.observed})
			if err != nil {
				t.Fatal(err)
			}
			input := domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: tc.accepted, Fresh: tc.fresh, Observed: tc.observed}
			result := domain.RetainAcceptedChildLinks(input)
			after, err := json.Marshal([][]domain.SessionFact{tc.accepted, tc.fresh, tc.observed})
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("the rule modified its inputs")
			}
			if len(result.Sessions) != len(tc.fresh) {
				t.Fatalf("returned %d owners, want the %d fresh owners", len(result.Sessions), len(tc.fresh))
			}
			for i := range tc.fresh {
				if result.Sessions[i].ID != tc.fresh[i].ID {
					t.Fatalf("owner order changed: %d is %s, want %s", i, result.Sessions[i].ID, tc.fresh[i].ID)
				}
			}
			if got := ownerLinks(result.Sessions, "P"); !reflect.DeepEqual(got, tc.wantLinks) {
				t.Fatalf("links = %q, want %q", got, tc.wantLinks)
			}
			if got := outcomes(result.Report); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("outcomes = %v, want %v (%+v)", got, tc.want, result.Report.Decisions)
			}
			if tc.evidence != "" && result.Report.Decisions[0].Evidence != tc.evidence {
				t.Fatalf("evidence = %s, want %s", result.Report.Decisions[0].Evidence, tc.evidence)
			}
			if got := result.Report.RetainedLinks(); !reflect.DeepEqual(got, tc.retained) {
				t.Fatalf("retained links = %+v, want %+v", got, tc.retained)
			}
			// Every other fact of every owner is the fresh read.
			for i, owner := range result.Sessions {
				want := tc.fresh[i]
				owner.Delegations, want.Delegations = nil, nil
				if !reflect.DeepEqual(owner, want) {
					t.Fatalf("owner %s changed beyond delegation links: %+v", owner.ID, owner)
				}
				if len(result.Sessions[i].Delegations) != len(tc.fresh[i].Delegations) {
					t.Fatalf("owner %s delegation count changed", owner.ID)
				}
			}
			// Applying the rule to its own result changes nothing.
			again := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: tc.accepted, Fresh: result.Sessions, Observed: tc.observed})
			if !reflect.DeepEqual(again.Sessions, result.Sessions) {
				t.Fatalf("rule is not idempotent: %+v then %+v", result.Sessions, again.Sessions)
			}
			for _, decision := range again.Report.Decisions {
				if decision.Outcome == domain.ChildLinkCleared {
					t.Fatalf("second application cleared a link: %+v", decision)
				}
			}
		})
	}
}

func TestRetainAcceptedChildLinksScopesToTheTargetHarness(t *testing.T) {
	accepted := linkOwner("P", "", linkDelegation("worker", "task-01", at1, "C1"))
	fresh := linkOwner("P", "", linkDelegation("worker", "task-01", at1, ""))
	openCode := fresh
	openCode.Source = domain.SourceOpenCode
	acceptedOpenCode := accepted
	acceptedOpenCode.Source = domain.SourceOpenCode
	for _, tc := range []struct {
		name     string
		harness  domain.SourceHarness
		accepted domain.SessionFact
		fresh    domain.SessionFact
	}{
		{"OpenCode owner under the Codex rule", domain.SourceCodex, acceptedOpenCode, openCode},
		{"Codex owner whose accepted record is another harness", domain.SourceCodex, acceptedOpenCode, fresh},
		{"Codex owner under the OpenCode rule", domain.SourceOpenCode, accepted, fresh},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: tc.harness, Accepted: []domain.SessionFact{tc.accepted}, Fresh: []domain.SessionFact{tc.fresh}})
			if !reflect.DeepEqual(result.Sessions, []domain.SessionFact{tc.fresh}) || len(result.Report.Decisions) != 0 {
				t.Fatalf("rule changed an owner outside its harness: %+v %+v", result.Sessions, result.Report)
			}
		})
	}
}

func TestRetainAcceptedChildLinksReportsEachRetainedPairOnce(t *testing.T) {
	// Two owners retain links; the same pair from a repeated input is
	// reported once, in owner and delegation order.
	accepted := []domain.SessionFact{
		linkOwner("P", "", linkDelegation("worker", "a", at1, "C1"), linkDelegation("worker", "b", at2, "C2")),
		linkOwner("R", "", linkDelegation("worker", "a", at1, "C3")),
	}
	fresh := []domain.SessionFact{
		linkOwner("P", "", linkDelegation("worker", "a", at1, ""), linkDelegation("worker", "b", at2, "")),
		linkOwner("R", "", linkDelegation("worker", "a", at1, "")),
		linkOwner("P", "", linkDelegation("worker", "a", at1, ""), linkDelegation("worker", "b", at2, "")),
	}
	result := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Fresh: fresh})
	want := []domain.ChildLink{{ParentID: "P", ChildSessionID: "C1"}, {ParentID: "P", ChildSessionID: "C2"}, {ParentID: "R", ChildSessionID: "C3"}}
	if got := result.Report.RetainedLinks(); !reflect.DeepEqual(got, want) {
		t.Fatalf("retained links = %+v, want %+v", got, want)
	}
	if len(result.Report.Decisions) != 5 {
		t.Fatalf("each delegation decision is recorded: %+v", result.Report.Decisions)
	}
	if result.Report.Decisions[1].Delegation != 1 || result.Report.Decisions[1].ParentID != "P" {
		t.Fatalf("decision does not name its delegation: %+v", result.Report.Decisions[1])
	}
}

// The design states that applying the rule again changes nothing. A fresh
// link that names an accepted sibling's child without confirmation is
// replaced by the accepted link, which frees that child only after the
// sibling delegation was already refused as "held by sibling".
func TestRetainAcceptedChildLinksIsIdempotentAfterSiblingRelease(t *testing.T) {
	accepted := []domain.SessionFact{linkOwner("P", "",
		linkDelegation("worker", "task-01", at1, "C1"),
		linkDelegation("worker", "task-02", at2, "C2"))}
	// The fresh read names C1 on the second delegation, for example through an
	// explicit agent ID, and this refresh read neither child.
	fresh := []domain.SessionFact{linkOwner("P", "",
		linkDelegation("worker", "task-01", at1, ""),
		linkDelegation("worker", "task-02", at2, "C1"))}
	first := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Fresh: fresh})
	second := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Fresh: first.Sessions})
	if !reflect.DeepEqual(first.Sessions, second.Sessions) {
		t.Fatalf("second application changed links: first %q (%v), second %q (%v)",
			ownerLinks(first.Sessions, "P"), outcomes(first.Report), ownerLinks(second.Sessions, "P"), outcomes(second.Report))
	}
	// The unconfirmed fresh C1 gives way to the accepted C2, which frees C1
	// for the first delegation in the same application.
	if got := ownerLinks(first.Sessions, "P"); !reflect.DeepEqual(got, []string{"C1", "C2"}) {
		t.Fatalf("links = %q, want [C1 C2]", got)
	}
	want := []domain.ChildLinkOutcome{domain.ChildLinkRetained, domain.ChildLinkKeptPendingConfirmation}
	if got := outcomes(first.Report); !reflect.DeepEqual(got, want) || len(first.Report.Refused) != 0 {
		t.Fatalf("outcomes = %v refused = %+v, want %v and no refusal", got, first.Report.Refused, want)
	}
}

// FIX-21 DATA-02: an accepted link the rule declines to carry, and that no
// evidence contradicts, is reported once with its typed reason. It is never
// written to a delegation, and it is unresolved child evidence.
func TestRetainAcceptedChildLinksReportsRefusedLinks(t *testing.T) {
	first := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-01", at1, child) }
	second := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-02", at2, child) }
	unknown := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-01", "", child) }
	link := func(child string) domain.ChildLink { return domain.ChildLink{ParentID: "P", ChildSessionID: child} }
	for _, tc := range []struct {
		name       string
		accepted   []domain.SessionFact
		fresh      []domain.SessionFact
		wantLinks  []string
		refused    []domain.RefusedChildLink
		unresolved []domain.ChildLink
	}{
		{
			// An empty fresh group as large as the accepted one is carried by
			// position, so nothing is refused; the carried links are unresolved.
			name:       "duplicate provenance carried by position is not refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks:  []string{"C1", "C2"},
			refused:    []domain.RefusedChildLink{},
			unresolved: []domain.ChildLink{link("C1"), link("C2")},
		},
		{
			name:       "a larger fresh group carried by position is not refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks:  []string{"C1", ""},
			refused:    []domain.RefusedChildLink{},
			unresolved: []domain.ChildLink{link("C1")},
		},
		{
			name:       "unknown invocation time carried by position is not refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", unknown("C1"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", unknown(""))},
			wantLinks:  []string{"C1"},
			refused:    []domain.RefusedChildLink{},
			unresolved: []domain.ChildLink{link("C1")},
		},
		{
			name:       "a fresh link set in the group refuses the whole group",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first("C9"), first(""))},
			wantLinks:  []string{"C9", ""},
			refused:    []domain.RefusedChildLink{{Link: link("C1"), Reason: domain.ChildLinkRefusalAmbiguous}, {Link: link("C2"), Reason: domain.ChildLinkRefusalAmbiguous}},
			unresolved: []domain.ChildLink{link("C1"), link("C2")},
		},
		{
			name:       "a smaller fresh group is refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""))},
			wantLinks:  []string{""},
			refused:    []domain.RefusedChildLink{{Link: link("C1"), Reason: domain.ChildLinkRefusalAmbiguous}, {Link: link("C2"), Reason: domain.ChildLinkRefusalAmbiguous}},
			unresolved: []domain.ChildLink{link("C1"), link("C2")},
		},
		{
			name:       "a smaller fresh group with unknown invocation time is refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", unknown("C1"), unknown("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", unknown(""))},
			wantLinks:  []string{""},
			refused:    []domain.RefusedChildLink{{Link: link("C1"), Reason: domain.ChildLinkRefusalUnknownProvenance}, {Link: link("C2"), Reason: domain.ChildLinkRefusalUnknownProvenance}},
			unresolved: []domain.ChildLink{link("C1"), link("C2")},
		},
		{
			// The ordinal carry skips a child another final delegation holds.
			name:       "a child held by another final delegation is refused within an ordinal group",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), first(""), second("C1")), linkOwner("C1", "P")},
			wantLinks:  []string{"", "C2", "C1"},
			refused:    []domain.RefusedChildLink{{Link: link("C1"), Reason: domain.ChildLinkRefusalHeldBySibling}},
			unresolved: []domain.ChildLink{link("C2"), link("C1")},
		},
		{
			name: "held by a confirmed sibling",
			// The fresh read links C1 on the second delegation and confirms
			// it; the first delegation's accepted C1 cannot move there. The
			// refusal is reported because the delegation that had C1 lost it.
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), second("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), second("C1")), linkOwner("C1", "P")},
			wantLinks:  []string{"", "C1"},
			refused:    []domain.RefusedChildLink{{Link: link("C1"), Reason: domain.ChildLinkRefusalHeldBySibling}},
			unresolved: []domain.ChildLink{link("C1")},
		},
		{
			name:       "held by an unconfirmed sibling that keeps its own accepted link",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), second("C2")), linkOwner("C1", "Q")},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), second(""))},
			wantLinks:  []string{"", "C2"},
			refused:    []domain.RefusedChildLink{},
			unresolved: []domain.ChildLink{link("C2")},
		},
		{
			// Within an ordinal group the contradicted child is skipped and is
			// not a refusal; its sibling is still carried.
			name:       "contradicted child in an ordinal group is skipped, not refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2")), linkOwner("C1", "Q")},
			fresh:      []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks:  []string{"", "C2"},
			refused:    []domain.RefusedChildLink{},
			unresolved: []domain.ChildLink{link("C2")},
		},
		{
			name:       "an accepted link the fresh read re-proves elsewhere is not refused",
			accepted:   []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:      []domain.SessionFact{linkOwner("P", "", first("C1"), first(""))},
			wantLinks:  []string{"C1", ""},
			refused:    []domain.RefusedChildLink{{Link: link("C2"), Reason: domain.ChildLinkRefusalAmbiguous}},
			unresolved: []domain.ChildLink{link("C2")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: tc.accepted, Fresh: tc.fresh})
			if got := ownerLinks(result.Sessions, "P"); !reflect.DeepEqual(got, tc.wantLinks) {
				t.Fatalf("links = %q, want %q: no refused link may be written", got, tc.wantLinks)
			}
			if !reflect.DeepEqual(result.Report.Refused, tc.refused) {
				t.Fatalf("refused = %+v, want %+v", result.Report.Refused, tc.refused)
			}
			if got := result.Report.UnresolvedLinks(); !reflect.DeepEqual(got, tc.unresolved) {
				t.Fatalf("unresolved = %+v, want %+v", got, tc.unresolved)
			}
			// Idempotence is about facts. The report of a second application
			// can differ: an ordinal group the rule filled is then partly
			// linked, so it is refused as a whole with nothing left to report.
			again := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: tc.accepted, Fresh: result.Sessions})
			if !reflect.DeepEqual(again.Sessions, result.Sessions) {
				t.Fatalf("second application changed links: %q", ownerLinks(again.Sessions, "P"))
			}
		})
	}
}

// FIX-21 DATA-04: Known records, such as the capture registry when Accepted
// is an older committed ledger, take precedence as child evidence.
func TestRetainAcceptedChildLinksPrefersKnownChildEvidence(t *testing.T) {
	first := func(child string) domain.DelegationFact { return linkDelegation("worker", "task-01", at1, child) }
	accepted := []domain.SessionFact{linkOwner("P", "", first("C1")), linkOwner("C1", "P")}
	for _, tc := range []struct {
		name      string
		known     []domain.SessionFact
		fresh     []domain.SessionFact
		wantLinks []string
		want      domain.ChildLinkOutcome
	}{
		// Divergent: the registry proves the fresh child C9, which the older
		// ledger does not know, so it replaces the ledger link.
		{"registry confirms a different fresh child", []domain.SessionFact{linkOwner("C9", "P")}, []domain.SessionFact{linkOwner("P", "", first("C9"))}, []string{"C9"}, domain.ChildLinkReplaced},
		// Divergent: the registry's newer record of C1 names another parent.
		{"registry contradicts the ledger's child", []domain.SessionFact{linkOwner("C1", "Q")}, []domain.SessionFact{linkOwner("P", "", first(""))}, []string{""}, domain.ChildLinkRefusedContradicted},
		// A registry record without a parent is unobserved, not confirmation.
		{"registry record without a parent", []domain.SessionFact{linkOwner("C1", "")}, []domain.SessionFact{linkOwner("P", "", first(""))}, []string{"C1"}, domain.ChildLinkRetained},
		// Agreement changes nothing.
		{"registry agrees with the ledger", []domain.SessionFact{linkOwner("C1", "P")}, []domain.SessionFact{linkOwner("P", "", first(""))}, []string{"C1"}, domain.ChildLinkRetained},
		// Fresh evidence still beats both.
		{"fresh record beats the registry", []domain.SessionFact{linkOwner("C1", "P")}, []domain.SessionFact{linkOwner("P", "", first("")), linkOwner("C1", "Q")}, []string{""}, domain.ChildLinkRefusedContradicted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			with := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Known: tc.known, Fresh: tc.fresh})
			if got := ownerLinks(with.Sessions, "P"); !reflect.DeepEqual(got, tc.wantLinks) || len(with.Report.Decisions) != 1 || with.Report.Decisions[0].Outcome != tc.want {
				t.Fatalf("links = %q outcomes = %v, want %q and %s", got, outcomes(with.Report), tc.wantLinks, tc.want)
			}
		})
	}
	// Without Known, the divergent cases fall back to the older ledger.
	without := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Fresh: []domain.SessionFact{linkOwner("P", "", first("C9"))}})
	if got := ownerLinks(without.Sessions, "P"); !reflect.DeepEqual(got, []string{"C1"}) {
		t.Fatalf("without registry evidence the ledger link stays pending confirmation, got %q", got)
	}
}

// FIX-21 ordinal group retention: accepted link i of a key group is carried to
// fresh delegation i when every fresh link of the group is empty and the fresh
// group is at least as large.
func TestRetainAcceptedChildLinksOrdinalGroups(t *testing.T) {
	group := func(child string) domain.DelegationFact { return linkDelegation("tester", "", at1, child) }
	accepted := []domain.SessionFact{linkOwner("P", "", group("C1"), group("C2")), linkOwner("C1", "P"), linkOwner("C2", "P")}
	apply := func(accepted, fresh []domain.SessionFact) domain.ChildLinkResult {
		return domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Fresh: fresh})
	}
	t.Run("appended delegations keep their position", func(t *testing.T) {
		fresh := []domain.SessionFact{linkOwner("P", "", group(""), group(""), group(""))}
		result := apply(accepted, fresh)
		if got := ownerLinks(result.Sessions, "P"); !reflect.DeepEqual(got, []string{"C1", "C2", ""}) {
			t.Fatalf("links = %q, want [C1 C2 \"\"]", got)
		}
		if got := outcomes(result.Report); !reflect.DeepEqual(got, []domain.ChildLinkOutcome{domain.ChildLinkRetainedByPosition, domain.ChildLinkRetainedByPosition}) {
			t.Fatalf("outcomes = %v", got)
		}
		for i, decision := range result.Report.Decisions {
			if decision.Delegation != i || decision.Evidence != domain.ChildParentConfirmed {
				t.Fatalf("decision %d = %+v, want delegation %d confirmed by the accepted child record", i, decision, i)
			}
		}
		if got := result.Report.UnresolvedLinks(); len(got) != 2 || len(result.Report.Refused) != 0 {
			t.Fatalf("unresolved = %+v refused = %+v, want the two carried links and no refusal", got, result.Report.Refused)
		}
	})
	t.Run("idempotent", func(t *testing.T) {
		fresh := []domain.SessionFact{linkOwner("P", "", group(""), group(""), group(""))}
		once := apply(accepted, fresh)
		twice := apply(accepted, once.Sessions)
		if !reflect.DeepEqual(once.Sessions, twice.Sessions) {
			t.Fatalf("f(f(x)) = %q, f(x) = %q", ownerLinks(twice.Sessions, "P"), ownerLinks(once.Sessions, "P"))
		}
	})
	t.Run("repeated pass with the published output as the accepted ledger", func(t *testing.T) {
		// Each pass publishes its result, and the next identical partial read
		// takes that publication as accepted facts.
		ledger := accepted
		fresh := []domain.SessionFact{linkOwner("P", "", group(""), group(""), group(""))}
		var first []domain.SessionFact
		for pass := 0; pass < 3; pass++ {
			result := apply(ledger, fresh)
			if pass == 0 {
				first = result.Sessions
			} else if !reflect.DeepEqual(result.Sessions, first) {
				t.Fatalf("pass %d links = %q, want %q", pass+1, ownerLinks(result.Sessions, "P"), ownerLinks(first, "P"))
			}
			if got := result.Report.UnresolvedLinks(); len(got) != 2 {
				t.Fatalf("pass %d reports %d unresolved links, want the two carried links on every pass", pass+1, len(got))
			}
			ledger = append(append([]domain.SessionFact(nil), result.Sessions...), linkOwner("C1", "P"), linkOwner("C2", "P"))
		}
	})
	t.Run("no link is invented", func(t *testing.T) {
		// An accepted delegation with no child leaves its fresh position empty.
		sparse := []domain.SessionFact{linkOwner("P", "", group(""), group("C2"))}
		result := apply(sparse, []domain.SessionFact{linkOwner("P", "", group(""), group(""))})
		if got := ownerLinks(result.Sessions, "P"); !reflect.DeepEqual(got, []string{"", "C2"}) {
			t.Fatalf("links = %q, want [\"\" C2]", got)
		}
	})
}
