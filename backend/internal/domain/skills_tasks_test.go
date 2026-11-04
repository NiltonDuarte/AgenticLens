package domain_test

import (
	"agenticlens/internal/domain"
	"encoding/json"
	"reflect"
	"testing"
)

func TestLeanOwnedOccurrencesPreserveMultiplicityAndNulls(t *testing.T) {
	same := domain.SkillFact{Name: "inspect"}
	ledger := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: "root", Source: domain.SourceCodex, Skills: []domain.SkillFact{same, same}, Delegations: []domain.DelegationFact{{}, {ChildSession: "unavailable"}}}, {ID: "child", Source: domain.SourceOpenCode, Skills: []domain.SkillFact{same}}}}
	before, _ := json.Marshal(ledger)
	if len(ledger.SkillFacts()) != 3 || len(ledger.DelegationFacts()) != 2 {
		t.Fatal("projection collapsed equal occurrences")
	}
	adoption := domain.ComputeSkillAdoption([]domain.AgenticFactLedger{ledger})
	if len(adoption) != 1 || adoption[0].TotalInvocations != 3 || adoption[0].TicketsCount != 1 {
		t.Fatalf("adoption count mismatch: %+v", adoption)
	}
	var decoded domain.AgenticFactLedger
	if err := json.Unmarshal(before, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Sessions, ledger.Sessions) {
		t.Fatalf("nullable occurrences changed: %+v", decoded)
	}
	after, _ := json.Marshal(ledger)
	if string(before) != string(after) {
		t.Fatal("projection mutated canonical input")
	}
}
