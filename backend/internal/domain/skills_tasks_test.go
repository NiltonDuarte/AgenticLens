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

func TestOwnedFactProjectionsPreserveOccurrenceAndLegacyMultiplicity(t *testing.T) {
	// Explicit source identity removes only a transitional root echo. The same
	// call ID in another session and repeated unresolved values remain distinct.
	input := `{"sessions":[
		{"id":"root","source":"codex","skills":[{"name":"same","success":false,"outcome":"unavailable","source_invocation_id":"call"},{"name":"same","success":true,"source_invocation_id":"repeat"}],"task_links":[{"description":"spawn","source_invocation_id":"spawn","child_session_id":"missing-child"}]},
		{"id":"child","parent_id":"root","source":"codex","skills":[{"name":"same","success":true,"source_invocation_id":"call"}],"task_links":[{"description":"spawn","source_invocation_id":"spawn"}]}
	],"skills":[{"name":"same","session_id":"root","source":"codex","source_invocation_id":"call","outcome":"unavailable","success":false},{"name":"legacy","success":true},{"name":"legacy","success":true}],"task_links":[{"description":"spawn","session_id":"root","source":"codex","source_invocation_id":"spawn","child_session_id":"missing-child"},{"description":"legacy"},{"description":"legacy"}]}`
	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal([]byte(input), &ledger); err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	skills, tasks := ledger.SkillFacts(), ledger.TaskLinkFacts()
	if len(skills) != 5 || len(tasks) != 4 {
		t.Fatalf("projection erased occurrences or retained known echoes: skills=%+v tasks=%+v", skills, tasks)
	}
	skillOwners := map[string]int{}
	for _, skill := range skills {
		skillOwners[skill.SessionID]++
		if skill.SessionID != "" && skill.Source != domain.SourceCodex {
			t.Fatalf("owner source not projected: %+v", skill)
		}
	}
	if skillOwners["root"] != 2 || skillOwners["child"] != 1 || skillOwners[""] != 2 {
		t.Fatalf("wrong session ownership: %+v", skillOwners)
	}
	taskOwners := map[string]int{}
	for _, task := range tasks {
		taskOwners[task.SessionID]++
		if task.SessionID == "root" && task.ChildSession != "missing-child" {
			t.Fatalf("unavailable child reference lost: %+v", task)
		}
	}
	if taskOwners["root"] != 1 || taskOwners["child"] != 1 || taskOwners[""] != 2 {
		t.Fatalf("child reference changed task owner: %+v", taskOwners)
	}
	if len(ledger.SkillFacts()) != 5 || len(ledger.TaskLinkFacts()) != 4 {
		t.Fatal("repeat projection accumulated facts")
	}
	after, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("read projection rewrote canonical facts")
	}
}
