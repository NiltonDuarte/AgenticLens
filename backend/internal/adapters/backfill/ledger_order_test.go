package backfill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestBackfillPreservesRetainedEvidenceAndMigrationStepOrder(t *testing.T) {
	root, missingCodexHome := t.TempDir(), t.TempDir()
	version, ticket, title := domain.CurrentSchemaVersion, "TEST-ORDER", "Retained evidence order"
	workspace := createTestWorkspace(t, root, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title,
		// Existing Git facts resolve the pending Git step without a Git probe.
		Code: domain.CodeFacts{LinesAdded: 1},
		Migration: &domain.MigrationInfo{
			SchemaVersion: version, Status: domain.MigrationStatusPartial,
			StepsCompleted: []string{"z-completed", "a-completed"},
			StepsPending:   []string{"z-pending", domain.MigrationStepGitCodeFacts, "a-pending"},
		},
		Sessions: []domain.SessionFact{
			{ID: "z-unavailable", Source: domain.SourceCodex, User: "remote",
				Anomalies: []domain.Anomaly{
					{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning},
					{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical},
				}},
			{ID: "a-other-source", Source: domain.SourceOpenCode, User: "other",
				Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}}},
		},
		Anomalies: []domain.Anomaly{
			{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning},
			{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical},
		},
	})
	initial, _ := readRecoveryLedger(t, workspace)
	initial.Sessions[0].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{
		{SessionID: "z-unavailable", EventID: "z-event"},
		{SessionID: "z-unavailable", EventID: "a-event"},
	}
	initial.Sessions[1].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{
		{SessionID: "a-other-source", EventID: "other-event"},
	}
	initial.Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{
		{SessionID: "z-unavailable", EventID: "z-event"},
		{SessionID: "a-other-source", EventID: "other-event"},
	}
	body, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}

	opts := BackfillOptions{Project: root, TicketID: ticket, Source: "codex", CodexHome: missingCodexHome, RecomputeDB: true}
	wantCompleted := []string{"z-completed", "a-completed", domain.MigrationStepGitCodeFacts}
	wantPending := []string{"z-pending", "a-pending", domain.MigrationStepOpenCodeDB}
	for pass := 1; pass <= 2; pass++ {
		report, err := NewEngine().Execute(opts)
		if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil || report.Updated != 1 || !report.Results[0].RefreshIncomplete {
			t.Fatalf("pass %d: unavailable source did not exercise partial publication: %+v, %v", pass, report, err)
		}
		got, _ := readRecoveryLedger(t, workspace)
		if len(got.Sessions) != 2 || got.Sessions[0].ID != "a-other-source" || got.Sessions[1].ID != "z-unavailable" {
			t.Fatalf("pass %d: unexpected publication owner order: %+v", pass, got.Sessions)
		}
		for i, expected := range []adapters.SessionFactDTO{initial.Sessions[1], initial.Sessions[0]} {
			if !reflect.DeepEqual(got.Sessions[i].Anomalies, expected.Anomalies) {
				t.Errorf("pass %d: retained anomaly or evidence order changed for %s: got %+v, want %+v", pass, expected.ID, got.Sessions[i].Anomalies, expected.Anomalies)
			}
		}
		if !reflect.DeepEqual(got.Anomalies, initial.Anomalies) {
			t.Errorf("pass %d: retained ticket anomaly or evidence order changed: got %+v, want %+v", pass, got.Anomalies, initial.Anomalies)
		}
		if got.Migration == nil || got.Migration.Status != string(domain.MigrationStatusPartial) {
			t.Fatalf("pass %d: unresolved migration metadata missing: %+v", pass, got.Migration)
		}
		if !slices.Equal(got.Migration.StepsCompleted, wantCompleted) || !slices.Equal(got.Migration.StepsPending, wantPending) {
			t.Errorf("pass %d: migration step order changed: completed %v want %v; pending %v want %v", pass, got.Migration.StepsCompleted, wantCompleted, got.Migration.StepsPending, wantPending)
		}
	}
}

func TestBackfillPartialSourceRefreshRetainsTicketAnomalyEvidence(t *testing.T) {
	root := newBackfillProject(t)
	version, ticket, title := domain.CurrentSchemaVersion, "TEST-PARTIAL-ORDER", "Partial source order"
	parent := "local"
	issue := domain.Anomaly{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning}
	group := domain.Anomaly{RuleID: domain.RuleSubagentReworkLoop, Severity: domain.SeverityWarning}
	workspace := createTestWorkspace(t, root, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title,
		Sessions: []domain.SessionFact{
			{ID: "remote", ParentID: &parent, Source: domain.SourceCodex, User: "remote", Anomalies: []domain.Anomaly{issue}},
			{ID: parent, Source: domain.SourceCodex, User: "local", Anomalies: []domain.Anomaly{issue, group},
				Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-model"}}},
		},
		Anomalies: []domain.Anomaly{issue, {RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical}},
	})
	initial, _ := readRecoveryLedger(t, workspace)
	remote := adapters.AnomalyEvidenceRefDTO{SessionID: "remote", EventID: "remote-event"}
	local := adapters.AnomalyEvidenceRefDTO{SessionID: "local", EventID: "superseded-event"}
	initial.Sessions[0].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{remote}
	initial.Sessions[1].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{local}
	initial.Sessions[1].Anomalies[1].Evidence = []adapters.AnomalyEvidenceRefDTO{remote, local}
	initial.Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{remote, local}
	body, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	sessionsDir := filepath.Join(root, "sessions")
	writeBackfillRollout(t, sessionsDir, parent, "", "coder", "source-model", map[string]int{"read": 1}, "")

	opts := BackfillOptions{Project: root, TicketID: ticket, Source: "codex", CodexHome: root, RecomputeDB: true}
	report, err := NewEngine().Execute(opts)
	if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil || report.Updated != 1 || !report.Results[0].DBUpdated || !report.Results[0].RefreshIncomplete {
		t.Fatalf("expected one native refresh with incomplete ticket coverage: %+v, %v", report, err)
	}
	got, _ := readRecoveryLedger(t, workspace)
	if len(got.Sessions) != 2 || got.Sessions[0].ID != "local" || got.Sessions[1].ID != "remote" {
		t.Fatalf("unexpected publication owner order: %+v", got.Sessions)
	}
	fresh := got.Sessions[0]
	if len(fresh.Invocations) != 1 || fresh.Invocations[0].Model != "source-model" {
		t.Errorf("native parent refresh did not supersede its stored invocation: %+v", fresh.Invocations)
	}
	if !reflect.DeepEqual(fresh.Anomalies, initial.Sessions[1].Anomalies[1:]) {
		t.Errorf("parent refresh must retain only the incomplete group's summary and evidence: got %+v, want %+v", fresh.Anomalies, initial.Sessions[1].Anomalies[1:])
	}
	if !reflect.DeepEqual(got.Sessions[1], initial.Sessions[0]) {
		t.Error("partial refresh changed unavailable session facts or evidence")
	}
	if !reflect.DeepEqual(got.Anomalies, initial.Anomalies) {
		t.Errorf("partial source coverage changed ticket anomalies or evidence: got %+v, want %+v", got.Anomalies, initial.Anomalies)
	}

	writeBackfillRollout(t, sessionsDir, "remote", parent, "tester", "child-source", nil, "")
	report, err = NewEngine().Execute(opts)
	if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil || report.Updated != 1 || !report.Results[0].DBUpdated || report.Results[0].RefreshIncomplete {
		t.Fatalf("expected complete native group refresh: %+v, %v", report, err)
	}
	got, _ = readRecoveryLedger(t, workspace)
	if len(got.Sessions) != 2 || got.Sessions[0].ID != parent || got.Sessions[1].ID != "remote" || len(got.Sessions[1].Invocations) != 1 || got.Sessions[1].Invocations[0].Model != "child-source" {
		t.Fatalf("complete refresh did not read the child: %+v", got.Sessions)
	}
	if len(got.Sessions[0].Anomalies) != 0 || len(got.Sessions[1].Anomalies) != 0 || len(got.Anomalies) != 0 {
		t.Error("complete refresh restored stale parent, child or ticket findings")
	}
}
