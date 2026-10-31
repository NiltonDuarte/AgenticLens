package domain_test

import (
	"encoding/json"
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

	want := domain.AgenticFactLedger{
		Skills:    []domain.SkillFact{{Name: "agentic-lens-dev", TurnIndex: 2, MessageIndex: 1, CallerRole: "coder", Timestamp: "2026-10-31T10:00:00Z", Success: false, Error: "cancelled"}},
		TaskLinks: []domain.TaskLinkFact{{SubagentType: "tester", Description: "Run checks", ParentTurn: 3, ChildSession: "child-1", Timestamp: "2026-10-31T10:01:00Z", DurationMs: 25, CostUSD: 0.25}},
	}
	encoded, err = json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Skills) != 1 || decoded.Skills[0] != want.Skills[0] || len(decoded.TaskLinks) != 1 || decoded.TaskLinks[0] != want.TaskLinks[0] {
		t.Fatalf("optional facts failed JSON round trip: got %+v want %+v", decoded, want)
	}
}
