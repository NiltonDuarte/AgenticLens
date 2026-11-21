package adapters_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestFactLedgerStableSessionOrderAcrossPartialRefreshes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agentic_metrics.json")
	version, ticket, title := domain.CurrentSchemaVersion, "FIX-ORDER", "Stable ledger order"
	ledger := domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title,
		Sessions: []domain.SessionFact{
			{ID: "z-retained", Source: domain.SourceCodex, User: "fixture", Anomalies: []domain.Anomaly{
				{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning},
				{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical},
			}},
			{ID: "b-codex", Source: domain.SourceCodex, User: "fixture", Tools: []domain.SessionToolFact{{Name: "read", Calls: 2}}},
			{ID: "a-opencode", Source: domain.SourceOpenCode, User: "fixture", Tools: []domain.SessionToolFact{{Name: "read", Calls: 3}}},
		},
	}
	legacy := adapters.AgenticFactLedgerDTOFromDomain(ledger)
	legacy.Sessions[0].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{
		{SessionID: "z-retained", EventID: "z-event"}, {SessionID: "z-retained", EventID: "a-event"},
	}
	body, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	initial, err := adapters.LoadExistingFactLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	retained := initial.PersistedSessions[0]
	wantCalls := map[string]int64{"a-opencode": 3, "b-codex": 2}
	for _, refreshed := range []string{"b-codex", "a-opencode"} {
		previous, err := adapters.LoadExistingFactLedger(path)
		if err != nil {
			t.Fatal(err)
		}
		ledger.Sessions = previous.Sessions
		var history []adapters.SessionFactDTO
		for i, owner := range previous.PersistedSessions {
			if owner.ID == refreshed {
				wantCalls[refreshed]++
				ledger.Sessions[i].Tools[0].Calls = wantCalls[refreshed]
				continue
			}
			history = append(history, owner)
		}
		if err := adapters.SaveFactLedgerPreservingHistory(path, ledger, history); err != nil {
			t.Fatalf("refresh %s: %v", refreshed, err)
		}
		published, err := adapters.LoadExistingFactLedger(path)
		if err != nil {
			t.Fatal(err)
		}
		if ids := ledgerArrayKeys(published.PersistedSessions, func(s adapters.SessionFactDTO) string { return s.ID }); !slices.Equal(ids, []string{"a-opencode", "b-codex", "z-retained"}) {
			t.Fatalf("refresh %s produced session order %v", refreshed, ids)
		}
		for _, owner := range published.PersistedSessions[:2] {
			if owner.Tools[0].Calls != wantCalls[owner.ID] {
				t.Errorf("refresh %s reassigned tool count for %s: got %d, want %d", refreshed, owner.ID, owner.Tools[0].Calls, wantCalls[owner.ID])
			}
		}
		if !reflect.DeepEqual(published.PersistedSessions[2], retained) {
			t.Fatalf("refresh %s changed retained nested arrays or evidence", refreshed)
		}
	}
}

func TestMarshalFactLedgerStableNestedArrayOrder(t *testing.T) {
	stamp := func(value string) *string { return &value }
	late := stamp("2026-01-01T00:02:00Z")
	early := stamp("2026-01-01T01:00:00+01:00")
	tied := stamp("2026-01-01T00:00:00Z")
	version, ticket, title := domain.CurrentSchemaVersion, "FIX-ORDER", "Stable ledger order"
	ledger := domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title,
		Migration: &domain.MigrationInfo{SchemaVersion: version, Status: domain.MigrationStatusPartial, StepsCompleted: []string{"z-first", "a-next"}, StepsPending: []string{"y-first", "b-next"}},
		Sessions: []domain.SessionFact{
			{ID: "owner", Source: domain.SourceCodex, User: "fixture",
				Invocations: []domain.InvocationFact{{Model: "late", StartedAt: late}, {Model: "unknown"}, {Model: "early", StartedAt: early}, {Model: "tied", StartedAt: tied}},
				Skills:      []domain.SkillFact{{Name: "late", InvokedAt: late}, {Name: "unknown"}, {Name: "early", InvokedAt: early}, {Name: "tied", InvokedAt: tied}, {Name: "tied", InvokedAt: tied}},
				Delegations: []domain.DelegationFact{{Description: "late", InvokedAt: late}, {Description: "unknown"}, {Description: "early", InvokedAt: early}, {Description: "tied", InvokedAt: tied}},
				Tools:       []domain.SessionToolFact{{Name: "zeta", Calls: 7}, {Name: "alpha", Calls: 3}},
				Anomalies: []domain.Anomaly{
					{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning},
					{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical},
				}},
			{ID: "empty", Source: domain.SourceOpenCode, User: "fixture", Anomalies: []domain.Anomaly{}},
		},
		Anomalies: []domain.Anomaly{
			{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning},
			{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityCritical},
		},
	}
	candidate := adapters.AgenticFactLedgerDTOFromDomain(ledger)
	candidate.Sessions[0].Anomalies[0].Evidence = []adapters.AnomalyEvidenceRefDTO{
		{SessionID: "owner", EventID: "z"}, {SessionID: "empty", EventID: "z"},
		{SessionID: "owner", EventID: "a"}, {SessionID: "owner", EventID: "a"},
	}
	original, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := adapters.MarshalFactLedger(candidate, adapters.FactLedgerHistory{})
	if err != nil {
		t.Fatal(err)
	}
	var got adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Sessions) != 2 || got.Sessions[0].ID != "empty" || got.Sessions[1].ID != "owner" {
		t.Fatalf("unexpected sessions: %+v", got.Sessions)
	}
	empty, owner := got.Sessions[0], got.Sessions[1]
	if empty.Invocations == nil || empty.Tools == nil || empty.Anomalies == nil || empty.Skills != nil || empty.Delegations != nil {
		t.Fatalf("optional absence and explicit empty arrays changed: %+v", empty)
	}
	assertKeys := func(label string, actual, want []string) {
		t.Helper()
		if !slices.Equal(actual, want) {
			t.Errorf("%s order: got %v, want %v", label, actual, want)
		}
	}
	timeline := []string{"early", "unknown", "tied", "late"}
	assertKeys("invocations", ledgerArrayKeys(owner.Invocations, func(i adapters.InvocationFactDTO) string { return i.Model }), timeline)
	assertKeys("skills", ledgerArrayKeys(owner.Skills, func(s adapters.SkillFactDTO) string { return s.Name }), []string{"early", "unknown", "tied", "tied", "late"})
	assertKeys("delegations", ledgerArrayKeys(owner.Delegations, func(d adapters.DelegationFactDTO) string { return d.Description }), timeline)
	if owner.Invocations[0].StartedAt == nil || *owner.Invocations[0].StartedAt != *early {
		t.Error("timestamp spelling changed when ordering by absolute time")
	}
	assertKeys("tools", ledgerArrayKeys(owner.Tools, func(tool adapters.SessionToolFactDTO) string { return tool.Name }), []string{"alpha", "zeta"})
	severity := []string{string(domain.SeverityCritical), string(domain.SeverityWarning)}
	assertKeys("session anomalies", ledgerArrayKeys(owner.Anomalies, func(a adapters.AnomalyDTO) string { return a.Severity }), severity)
	assertKeys("ticket anomalies", ledgerArrayKeys(got.Anomalies, func(a adapters.AnomalyDTO) string { return a.Severity }), severity)
	if len(owner.Anomalies) == 2 {
		assertKeys("evidence", ledgerArrayKeys(owner.Anomalies[1].Evidence, func(e adapters.AnomalyEvidenceRefDTO) string { return e.SessionID + "/" + e.EventID }), []string{"empty/z", "owner/a", "owner/a", "owner/z"})
	}
	assertKeys("completed migration steps", got.Migration.StepsCompleted, candidate.Migration.StepsCompleted)
	assertKeys("pending migration steps", got.Migration.StepsPending, candidate.Migration.StepsPending)
	unchanged, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, unchanged) {
		t.Fatal("serialization mutated caller-owned arrays")
	}
	repeated, err := adapters.MarshalFactLedger(got, adapters.FactLedgerHistory{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, repeated) {
		t.Fatal("serializing the same facts twice changed ledger bytes")
	}
}

func ledgerArrayKeys[T any](items []T, key func(T) string) []string {
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = key(item)
	}
	return keys
}
