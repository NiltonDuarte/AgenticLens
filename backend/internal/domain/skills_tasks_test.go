package domain_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestSkillAndTaskFactLedgerFieldsAreOptionalAndRoundTrip(t *testing.T) {
	legacy := domain.AgenticFactLedger{}
	encoded, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"skills"`) || strings.Contains(string(encoded), `"task_links"`) {
		t.Fatalf("empty optional facts should be omitted: %s", encoded)
	}
	var decoded domain.AgenticFactLedger
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Skills) != 0 || len(decoded.TaskLinks) != 0 {
		t.Fatalf("legacy ledger should decode without optional facts: %+v", decoded)
	}

	want := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceCodex,
		Skills:    []domain.SkillFact{{Name: "agentic-lens-dev", TurnIndex: 2, MessageIndex: 1, CallerRole: "coder", Timestamp: "2026-10-31T10:00:00Z", Success: false, Outcome: domain.OutcomeFailed, Error: "cancelled"}},
		TaskLinks: []domain.TaskLinkFact{{Outcome: domain.OutcomeUnknown, SubagentType: "tester", Description: "Run checks", ParentTurn: 3, ChildSession: "child-1", Timestamp: "2026-10-31T10:01:00Z", DurationMs: 25, CostUSD: 0.25}},
	}}}

	encoded, err = json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.SkillFacts(), want.SkillFacts()) || !reflect.DeepEqual(decoded.TaskLinkFacts(), want.TaskLinkFacts()) {
		t.Fatalf("owned facts failed JSON round trip: got %+v want %+v", decoded, want)
	}
}

func TestOwnedSkillProjectionPreservesMultiplicityAndUnknownOutcomes(t *testing.T) {
	version := 6
	owned := domain.SkillFact{Name: "inspect", Success: false, Outcome: domain.OutcomeUnknown, EvidenceKind: "activation_request", Provenance: domain.FactProvenance{CallID: "same-local-call"}}
	ledger := domain.AgenticFactLedger{SchemaVersion: &version,
		Sessions: []domain.SessionFact{
			{ID: "root", Source: domain.SourceCodex, Skills: []domain.SkillFact{owned}},
			{ID: "child", Source: domain.SourceOpenCode, Skills: []domain.SkillFact{{Name: "inspect", Success: true, Outcome: domain.OutcomeSucceeded, Provenance: owned.Provenance}, {Name: "inspect", Outcome: domain.OutcomeFailed, Provenance: domain.FactProvenance{CallID: "second-call"}}}},
		},
	}
	before, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	canonical := ledger
	projected := canonical.SkillFacts()
	if len(projected) != 3 || len(canonical.Sessions[0].Skills) != 1 || len(canonical.Sessions[1].Skills) != 2 {
		t.Fatalf("native projection dedupe crossed owners or repeated values: %+v", projected)
	}
	byOwner := map[string]int{}
	for _, fact := range projected {
		byOwner[fact.SessionID]++
	}
	if byOwner[""] != 0 || byOwner["root"] != 1 || byOwner["child"] != 2 {
		t.Fatalf("wrong proven owners: %+v", byOwner)
	}
	adoption := domain.ComputeSkillAdoption([]domain.AgenticFactLedger{canonical})
	if len(adoption) != 1 || adoption[0].TotalInvocations != 3 || adoption[0].KnownOutcomes != 2 || adoption[0].UnknownOutcomes != 1 || adoption[0].FailureCount != 1 || adoption[0].FailureRate != 0.5 {
		t.Fatalf("unknown outcome became failure or denominator: %+v", adoption)
	}
	after, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("canonicalization mutated input")
	}
}

func TestTaskChildResolutionUsesOnlyNativeParentAndAgentPath(t *testing.T) {
	parent, other := "parent", "other-parent"
	base := []domain.SessionFact{
		{ID: parent, Source: domain.SourceCodex, TaskLinks: []domain.TaskLinkFact{
			{Description: "exact namespace", ChildReference: "/root/tester", Outcome: domain.OutcomeSucceeded},
			{Description: "contradiction", ChildReference: "/root/tester", ChildSession: "other", Outcome: domain.OutcomeSucceeded},
			{Description: "absent child", ChildReference: "/root/unavailable", Outcome: domain.OutcomeUnknown},
		}},
		{ID: "actual", Source: domain.SourceCodex, ParentID: &parent, AgentPath: "/root/tester"},
		{ID: "unrelated", Source: domain.SourceCodex, ParentID: &other, AgentPath: "/root/tester"},
	}
	resolved := domain.ResolveTaskChildren(base)
	if resolved[0].TaskLinks[0].ChildSession != "actual" {
		t.Fatalf("exact parent+native path did not resolve: %+v", resolved[0].TaskLinks[0])
	}
	contradiction := resolved[0].TaskLinks[1]
	if contradiction.ChildSession != "other" || contradiction.ChildReference != "/root/tester" || contradiction.Outcome != domain.OutcomeSucceeded || contradiction.ChildLinkDiagnostic == "" {
		t.Fatalf("contradiction silently rewrote raw relationship/execution: %+v", contradiction)
	}
	if resolved[0].TaskLinks[2].ChildSession != "" {
		t.Fatal("unavailable child guessed from unrelated namespace")
	}
	noMatch := append([]domain.SessionFact{}, base...)
	noMatch[0] = base[0]
	noMatch[0].TaskLinks = []domain.TaskLinkFact{{Outcome: domain.OutcomeSucceeded, ChildSession: "actual", ChildReference: "/root/no-match"}}
	mismatch := domain.ResolveTaskChildren(noMatch)[0].TaskLinks[0]
	if mismatch.ChildSession != "actual" || mismatch.ChildReference != "/root/no-match" || mismatch.ChildLinkDiagnostic == "" {
		t.Fatalf("BEH-04 known child path mismatch without namespace matches ignored: %+v", mismatch)
	}
	if base[0].TaskLinks[0].ChildSession != "" || base[0].TaskLinks[1].ChildLinkDiagnostic != "" {
		t.Fatal("resolution mutated input")
	}
	ambiguous := append(append([]domain.SessionFact{}, base...), domain.SessionFact{ID: "another", Source: domain.SourceCodex, ParentID: &parent, AgentPath: "/root/tester"})
	ambiguous[0] = base[0]
	ambiguous[0].TaskLinks = append([]domain.TaskLinkFact{}, base[0].TaskLinks...)
	ambiguous[0].TaskLinks = append(ambiguous[0].TaskLinks, domain.TaskLinkFact{Description: "explicit UUID disambiguates", ChildSession: "actual", ChildReference: "/root/tester", Outcome: domain.OutcomeSucceeded})
	got := domain.ResolveTaskChildren(ambiguous)
	if got[0].TaskLinks[0].ChildSession != "" || got[0].TaskLinks[0].ChildLinkDiagnostic == "" {
		t.Fatalf("ambiguous namespace was assigned arbitrarily: %+v", got[0].TaskLinks[0])
	}
	if got[0].TaskLinks[3].ChildSession != "actual" || got[0].TaskLinks[3].ChildLinkDiagnostic != "" {
		t.Fatalf("native explicit UUID lost to namespace reuse: %+v", got[0].TaskLinks[3])
	}
}
