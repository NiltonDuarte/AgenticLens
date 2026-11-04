package domain

import (
	"reflect"
	"testing"
)

func TestGenerationProjectionPreservesExecutionAndInput(t *testing.T) {
	parent := "root"
	anomaly := Anomaly{RuleID: RuleSubagentReworkLoop, Severity: SeverityWarning}
	original := AgenticFactLedger{Anomalies: []Anomaly{anomaly}, Sessions: []SessionFact{
		{ID: "root", Source: SourceCodex, Skills: []SkillFact{{Name: "inspect"}}, Delegations: []DelegationFact{{Description: "review"}}, Anomalies: []Anomaly{anomaly}, Invocations: []InvocationFact{{Role: "coder", Turns: 2, Tokens: InvocationTokens{Input: 23}, CostUSD: 0.5}}, Tools: []SessionToolFact{{Name: "load_skill", Calls: 1, Errors: 1}}},
		{ID: "child", ParentID: &parent, Source: SourceCodex, Skills: []SkillFact{{Name: "child-inspect"}}, Delegations: []DelegationFact{{Description: "child-review"}}, Anomalies: []Anomaly{anomaly}},
	}}
	for mask := 0; mask < 8; mask++ {
		policy := GenerationPolicy{Skills: mask&1 != 0, Delegations: mask&2 != 0, Anomalies: mask&4 != 0}
		got := ProjectGenerationFacts(original, policy)
		if !reflect.DeepEqual(got.Sessions[0].Invocations, original.Sessions[0].Invocations) || !reflect.DeepEqual(got.Sessions[0].Tools, original.Sessions[0].Tools) || got.Sessions[1].ParentID == nil || *got.Sessions[1].ParentID != "root" {
			t.Fatalf("policy %+v changed execution/tree", policy)
		}
		for _, owner := range got.Sessions {
			if (len(owner.Skills) > 0) != policy.Skills || (len(owner.Delegations) > 0) != policy.Delegations || (len(owner.Anomalies) > 0) != policy.Anomalies {
				t.Fatalf("policy %+v not applied to owner %+v", policy, owner)
			}
		}
		if (len(got.Anomalies) > 0) != policy.Anomalies {
			t.Fatal("root anomaly policy ignored")
		}
		if len(original.Sessions[0].Skills) != 1 || len(original.Sessions[1].Delegations) != 1 || len(original.Anomalies) != 1 || original.GenerationPolicy != nil {
			t.Fatal("projection mutated original facts")
		}
	}
	if got := (AgenticFactLedger{}).EffectiveGenerationPolicy(); !got.Skills || !got.Delegations || !got.Anomalies {
		t.Fatalf("historical missing policy disabled categories: %+v", got)
	}
}
