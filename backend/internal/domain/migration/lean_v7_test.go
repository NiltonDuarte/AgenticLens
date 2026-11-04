package migration_test

import (
	"encoding/json"
	"strings"
	"testing"

	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

func TestV7MigrationPreservesOwnersMultiplicityProvenChildrenAndSummaries(t *testing.T) {
	input := []byte(`{"schema_version":6,"ticket_id":"TEST-LEAN","ticket_title":"Lean migration","sessions":[
 {"id":"parent","source":"codex","user":"alice","invocations":[{"role":"coder","model":"fixture","cost_usd":1.25,"cost_evidence":{"state":"observed","known_subtotal_usd":1.25},"tokens":{"input":10}}],"tools":[{"name":"bash","calls":5,"errors":3}],"skills":[{"name":"same","caller_role":"coder","timestamp":"2026-01-01T00:00:01Z","success":true,"path":"private-path","provenance":{"call_id":"one"}},{"name":"same","caller_role":"coder","timestamp":"2026-01-01T00:00:01Z","success":false,"provenance":{"call_id":"two"}},{"name":"unknown"}],"task_links":[{"subagent_type":"tester","description":"Unique child","timestamp":"2026-01-01T00:00:02Z","child_reference":"/root/unique","cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}},{"child_reference":"/root/ambiguous","cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}},{"child_session_id":"unavailable-child","cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}},{"child_session_id":"wrong-parent-child","cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}}],"events":[{"event_id":"sparse","event_type":"tool_call","tool_name":"load_skill","status":"error","timestamp":"2026-01-01T00:00:01Z","turn":1,"role":"coder","tool_input":"private-source"}],"events_are_partial":true,"anomalies":[{"rule_id":"TOOL_RETRY_STREAK","severity":"WARNING","context":{"count":3,"tool_name":"bash"}}]},
 {"id":"unique-child","source":"codex","parent_id":"parent","agent_path":"/root/unique","invocations":[],"tools":[],"anomalies":[]},
 {"id":"ambiguous-a","source":"codex","parent_id":"parent","agent_path":"/root/ambiguous","invocations":[],"tools":[],"anomalies":[]},
 {"id":"ambiguous-b","source":"codex","parent_id":"parent","agent_path":"/root/ambiguous","invocations":[],"tools":[],"anomalies":[]},
 {"id":"wrong-parent-child","source":"codex","parent_id":"other-parent","invocations":[],"tools":[],"anomalies":[]},{"id":"other-parent","source":"codex","invocations":[],"tools":[],"anomalies":[]}]}`)
	got, err := migration.NewEngine(nil).Migrate(input, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion == nil || *got.SchemaVersion != 7 || len(got.Sessions) != 6 {
		t.Fatalf("schema/tree changed: %+v", got)
	}
	parent := got.Sessions[0]
	if len(parent.Skills) != 3 || parent.Skills[0].Name != "same" || parent.Skills[1].Name != "same" || parent.Skills[2].CallerRole != nil || parent.Skills[2].InvokedAt != nil {
		t.Fatalf("multiplicity or missing values changed: %+v", parent.Skills)
	}
	if parent.Skills[0].CallerRole == nil || *parent.Skills[0].CallerRole != "coder" || parent.Skills[0].InvokedAt == nil || *parent.Skills[0].InvokedAt != "2026-01-01T00:00:01Z" {
		t.Fatalf("recorded role/time lost: %+v", parent.Skills[0])
	}
	if len(parent.Delegations) != 4 || parent.Delegations[0].ChildSession != "unique-child" || parent.Delegations[1].ChildSession != "" || parent.Delegations[2].ChildSession != "unavailable-child" || parent.Delegations[3].ChildSession != "" || parent.Delegations[1].AgentType != nil || parent.Delegations[1].InvokedAt != nil {
		t.Fatalf("exact proof or unknowns changed: %+v", parent.Delegations)
	}
	if len(parent.Anomalies) != 1 || parent.Anomalies[0].RuleID != domain.RuleToolRetryStreak || parent.Anomalies[0].Context.Count == nil || *parent.Anomalies[0].Context.Count != 3 {
		t.Fatalf("partial old events replaced full summary: %+v", parent.Anomalies)
	}
	if len(parent.Invocations) != 1 || parent.Invocations[0].CostUSD != 1.25 || parent.Invocations[0].CostEvidence.State != domain.CostObserved || parent.Invocations[0].Tokens.Input != 10 || parent.Source != domain.SourceCodex || parent.User != "alice" {
		t.Fatalf("unrelated recorded facts changed: %+v", parent)
	}
	if got.Migration == nil {
		t.Fatal("missing audit")
	}
	for _, fragment := range []string{"converted skills=3 delegations=4", "discarded events=1 partial_markers=1", "rich_fields=18", "unresolved_child_references=2"} {
		if !strings.Contains(got.Migration.Notes, fragment) {
			t.Fatalf("missing count %q: %s", fragment, got.Migration.Notes)
		}
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{`"events"`, `"events_are_partial"`, `"agent_path"`, `"task_links"`, `"provenance"`, `"success"`, `"outcome"`, `"private-path"`, `"private-source"`} {
		if strings.Contains(string(data), retired) {
			t.Fatalf("retired payload %s persisted: %s", retired, data)
		}
	}
	again, err := migration.NewEngine(nil).Migrate(data, 7)
	if err != nil {
		t.Fatal(err)
	}
	stable, _ := json.Marshal(again)
	if string(stable) != string(data) {
		t.Fatalf("unchanged current migration not stable:\n%s\n%s", data, stable)
	}
}
