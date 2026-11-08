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
		if len(original.Sessions[0].Skills) != 1 || len(original.Sessions[1].Delegations) != 1 || len(original.Anomalies) != 1 {
			t.Fatal("projection mutated original facts")
		}
	}
}

func TestReadDerivationPreservesUnrecordedAndRecordedEmptyAnalysis(t *testing.T) {
	for _, recorded := range []bool{false, true} {
		ledger := AgenticFactLedger{Sessions: []SessionFact{{ID: "owner", Tools: []SessionToolFact{{Name: "bash", Calls: 6, Errors: 3}}, Invocations: []InvocationFact{{Role: "coder", Turns: 2, Tokens: InvocationTokens{Input: 31}}}}}}
		if recorded {
			ledger.Anomalies = []Anomaly{}
			ledger.Sessions[0].Anomalies = []Anomaly{}
		}
		got := DeriveMetricsLedger(ledger, DefaultProjectConfig())
		if len(got.Anomalies) != 0 {
			t.Fatalf("ordinary read invented analysis from independent tool totals: %+v", got.Anomalies)
		}
		if got.Facts.Anomalies == nil != (!recorded) || got.Facts.Sessions[0].Anomalies == nil != (!recorded) {
			t.Fatal("read changed recorded coverage")
		}
		if !recorded && got.Totals.ContextHealth != ContextHealthUnavailable {
			t.Fatalf("unrecorded analysis claimed health %s", got.Totals.ContextHealth)
		}
		if got.WorkflowEfficiency.ToolExecution.TotalCalls != 6 || got.WorkflowEfficiency.ToolExecution.ErrorsAndCorrections != 3 {
			t.Fatalf("read lost independent tool evidence: %+v", got.WorkflowEfficiency.ToolExecution)
		}
	}
}

func TestSourceAnomalyPreparationRequiresRootAndChildCoverage(t *testing.T) {
	parent := "root"
	original := AgenticFactLedger{Sessions: []SessionFact{{ID: parent}, {ID: "child", ParentID: &parent}}}
	policy := AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .35, ErrorRateCriticalThreshold: .70}
	partial, err := PersistSourceAnomalySummaries(original, SourceEvents{CompleteSessionIDs: []string{parent}}, policy)
	if err != nil {
		t.Fatal(err)
	}
	if partial.Anomalies != nil || partial.Sessions[0].Anomalies != nil || partial.Sessions[1].Anomalies != nil {
		t.Fatalf("complete parent fabricated root/child observation: %+v", partial)
	}
	complete, err := PersistSourceAnomalySummaries(original, SourceEvents{CompleteSessionIDs: []string{parent, "child"}}, policy)
	if err != nil {
		t.Fatal(err)
	}
	if complete.Anomalies == nil || complete.Sessions[0].Anomalies == nil || complete.Sessions[1].Anomalies == nil || len(complete.Anomalies) != 0 {
		t.Fatalf("complete empty source did not establish empty arrays: %+v", complete)
	}
	count := int64(3)
	tool := "recorded-source-only"
	known := Anomaly{RuleID: RuleToolRetryStreak, Severity: SeverityWarning, Context: AnomalyContext{ToolName: &tool, Count: &count}}
	original.Anomalies = []Anomaly{known}
	retained, err := PersistSourceAnomalySummaries(original, SourceEvents{CompleteSessionIDs: []string{parent}}, policy)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retained.Anomalies, original.Anomalies) {
		t.Fatal("partial source coverage erased recorded root/group history")
	}
}
