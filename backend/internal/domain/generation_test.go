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

func TestExternalParentAnomalyCoverageRemainsPartial(t *testing.T) {
	parent := "external-parent"
	known := Anomaly{RuleID: RuleToolRetryStreak, Severity: SeverityWarning}
	policy := AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .35, ErrorRateCriticalThreshold: .70}
	for _, history := range []string{"fresh", "recorded-empty", "recorded-warning"} {
		t.Run(history, func(t *testing.T) {
			original := AgenticFactLedger{Sessions: []SessionFact{{ID: "child", ParentID: &parent, Source: SourceCodex, Invocations: []InvocationFact{{Role: "coder", Tokens: InvocationTokens{Input: 31}}}}}}
			switch history {
			case "recorded-empty":
				original.Anomalies = []Anomaly{}
				original.Sessions[0].Anomalies = []Anomaly{}
			case "recorded-warning":
				original.Anomalies = []Anomaly{known}
				original.Sessions[0].Anomalies = []Anomaly{known}
			}
			unavailable, err := PersistSourceAnomalySummaries(original, SourceEvents{})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(unavailable, original) {
				t.Fatal("unavailable source erased recorded external-parent history")
			}
			fresh, err := PersistSourceAnomalySummaries(original, SourceEvents{CompleteSessionIDs: []string{"child"}}, policy)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(fresh.Anomalies, original.Anomalies) || fresh.Sessions[0].Anomalies == nil {
				t.Fatalf("complete child replaced unavailable group history or lost independent owner analysis: %+v", fresh)
			}
			unrecordedScopes := 0
			if history == "fresh" {
				unrecordedScopes = 1
			}
			if fresh.CompleteAnomalyCoverage() || fresh.UnrecordedAnomalyScopes() != unrecordedScopes || fresh.UnrecordedAnomalySessions() != 0 {
				t.Fatalf("external-parent context fabricated completeness or discarded recorded coverage: %+v", fresh)
			}
			if len(fresh.Sessions) != 1 || fresh.Sessions[0].ParentID == nil || *fresh.Sessions[0].ParentID != parent || !reflect.DeepEqual(fresh.Sessions[0].Invocations, original.Sessions[0].Invocations) {
				t.Fatal("partial group analysis changed ancestry or owner metrics")
			}
			if history == "fresh" && (original.Anomalies != nil || original.Sessions[0].Anomalies != nil) || history == "recorded-warning" && (len(original.Anomalies) != 1 || len(original.Sessions[0].Anomalies) != 1) {
				t.Fatal("analysis mutated caller-owned source facts")
			}
		})
	}
}

func TestAnomalyCompletenessRequiresPolicyAndNativeGroupCoverage(t *testing.T) {
	valid := AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .35, ErrorRateCriticalThreshold: .70}
	for _, history := range []string{"fresh", "recorded-group"} {
		for _, tc := range []struct {
			name          string
			includeParent bool
			policy        AnomalyThresholdsConfig
			canAnalyze    bool
		}{
			{"closed group configured", true, valid, true},
			{"external parent configured", false, valid, true},
			{"closed group missing policy", true, AnomalyThresholdsConfig{}, false},
			{"external parent missing policy", false, AnomalyThresholdsConfig{}, false},
			{"closed group invalid policy", true, AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .90, ErrorRateCriticalThreshold: .70}, false},
		} {
			t.Run(history+"/"+tc.name, func(t *testing.T) {
				parent := "parent"
				original := AgenticFactLedger{Sessions: []SessionFact{{ID: "child", ParentID: &parent, Source: SourceCodex, Invocations: []InvocationFact{{Role: "coder", Tokens: InvocationTokens{Input: 31}}}}}}
				complete := []string{"child"}
				if tc.includeParent {
					original.Sessions = append(original.Sessions, SessionFact{ID: parent, Source: SourceCodex})
					complete = append(complete, parent)
				}
				if history == "recorded-group" {
					original.Anomalies = []Anomaly{{RuleID: RuleToolHighFailureRate, Severity: SeverityWarning}}
					for i := range original.Sessions {
						original.Sessions[i].Anomalies = []Anomaly{}
					}
				}
				got, err := PersistSourceAnomalySummaries(original, SourceEvents{CompleteSessionIDs: complete}, tc.policy)
				if err != nil {
					t.Fatal(err)
				}
				if tc.includeParent && tc.canAnalyze {
					if got.Anomalies == nil || len(got.Anomalies) != 0 || !got.CompleteAnomalyCoverage() {
						t.Fatalf("complete configured native group did not establish empty analysis: %+v", got)
					}
				} else {
					if !reflect.DeepEqual(got.Anomalies, original.Anomalies) {
						t.Fatalf("partial group or missing policy replaced recorded group facts: %+v", got)
					}
					if history == "fresh" && got.CompleteAnomalyCoverage() {
						t.Fatalf("partial source/policy fabricated complete clean analysis: %+v", got)
					}
				}
				if !tc.includeParent && got.CompleteAnomalyCoverage() {
					t.Fatalf("external native parent was treated as complete: %+v", got)
				}
				ownerRecorded := tc.canAnalyze || history == "recorded-group"
				if (got.Sessions[0].Anomalies != nil) != ownerRecorded {
					t.Fatalf("independent child analysis presence changed incorrectly: %+v", got.Sessions[0])
				}
				if got.Sessions[0].ParentID == nil || *got.Sessions[0].ParentID != parent || got.Sessions[0].Invocations[0].Tokens.Input != 31 {
					t.Fatal("coverage handling changed ancestry or numeric facts")
				}
			})
		}
	}
}
