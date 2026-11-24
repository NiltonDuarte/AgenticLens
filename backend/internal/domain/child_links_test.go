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
			name:      "duplicate accepted provenance moves no link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"), first("C2"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks: []string{"", ""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRefusedAmbiguous, domain.ChildLinkRefusedAmbiguous},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "duplicate fresh provenance moves no link",
			accepted:  []domain.SessionFact{linkOwner("P", "", first("C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", first(""), first(""))},
			wantLinks: []string{"", ""},
			want:      []domain.ChildLinkOutcome{domain.ChildLinkRefusedAmbiguous, domain.ChildLinkRefusedAmbiguous},
			retained:  []domain.ChildLink{},
		},
		{
			name:      "unknown invocation time moves no link",
			accepted:  []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "task-01", "", "C1"))},
			fresh:     []domain.SessionFact{linkOwner("P", "", linkDelegation("worker", "task-01", "", ""))},
			wantLinks: []string{""},
			want:      []domain.ChildLinkOutcome{},
			retained:  []domain.ChildLink{},
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
}
