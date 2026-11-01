package domain_test

import (
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

func TestSessionsFromFactsAggregatesDuplicateToolNamesWithinSession(t *testing.T) {
	sessions, err := domain.SessionsFromFacts(nil, []domain.ToolFact{
		{SessionID: "session-a", Name: "bash", Calls: 3, Errors: 1},
		{SessionID: "session-a", Name: "bash", Calls: 2, Errors: 1},
		{SessionID: "session-b", Name: "bash", Calls: 4, Errors: 0},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 || len(sessions[0].Tools) != 1 || sessions[0].Tools[0].Calls != 5 || sessions[0].Tools[0].Errors != 2 {
		t.Fatalf("same-session rows must aggregate once by tool name without crossing sessions: %+v", sessions)
	}
	if sessions[1].Tools[0].Calls != 4 || sessions[1].Tools[0].Errors != 0 {
		t.Fatalf("other session totals changed: %+v", sessions[1])
	}
}

func TestSessionsFromFactsPreservesOwnersAndEveryMetricRow(t *testing.T) {
	rootID := "opaque-root"
	started, completed := "2026-10-28T10:00:00Z", "2026-10-28T10:01:00Z"
	row := domain.InvocationFact{Role: "coder", Model: "model-a", StartedAt: &started, CompletedAt: &completed, CostUSD: 0.1}
	otherModel := row
	otherModel.Model = "model-b"
	childRow := row
	childRow.Role = "reviewer"
	owned := []domain.SessionFact{
		{ID: rootID, Source: domain.SourceCodex, User: "alice", Invocations: []domain.InvocationFact{row, row, otherModel}},
		{ID: "opaque-child", ParentID: &rootID, Source: domain.SourceCodex, User: "alice", Invocations: []domain.InvocationFact{childRow}},
	}
	tools := []domain.ToolFact{{SessionID: "opaque-child", Name: "bash", Calls: 2}}
	events := []domain.EventFact{{SessionID: rootID, EventID: "turn", Timestamp: started, Role: "coder", EventType: domain.EventTypeTurnMessage}}
	sessions, err := domain.SessionsFromFacts(owned, tools, events)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 || sessions[0].ID != rootID || sessions[1].ID != "opaque-child" || sessions[1].ParentID == nil || *sessions[1].ParentID != rootID {
		t.Fatalf("explicit ownership changed: %+v", sessions)
	}
	for i := range owned {
		if !reflect.DeepEqual(sessions[i].Invocations, owned[i].Invocations) || sessions[i].Source != domain.SourceCodex || sessions[i].User != "alice" {
			t.Fatalf("session metadata or metric rows changed: %+v", sessions[i])
		}
	}
	if len(sessions[0].Events) != 1 || len(sessions[0].Tools) != 0 || len(sessions[1].Events) != 0 || len(sessions[1].Tools) != 1 {
		t.Fatalf("tools or events crossed session boundaries: %+v", sessions)
	}
	metrics := domain.DeriveMetricsLedger(domain.AgenticFactLedger{Sessions: sessions}, domain.DefaultProjectConfig())
	if metrics.Totals.Invocations != 4 || len(metrics.Models) != 2 || len(metrics.Roles) != 2 {
		t.Fatalf("rows collapsed during aggregation: %+v", metrics)
	}
	for _, model := range metrics.Models {
		want := int64(1)
		if model.Model == "model-a" {
			want = 3
		}
		if model.Invocations != want {
			t.Errorf("model %s: expected %d rows, got %d", model.Model, want, model.Invocations)
		}
	}
	for _, role := range metrics.Roles {
		want := int64(1)
		if role.Role == "coder" {
			want = 3
		}
		if role.Invocations != want {
			t.Errorf("role %s: expected %d rows, got %d", role.Role, want, role.Invocations)
		}
	}
	threads := domain.ComputeThreadSummariesWithCeiling(sessions, domain.ContextCeilingCriticalTokens)
	if len(threads) != 1 || threads[0].SessionID != rootID || threads[0].User != "alice" {
		t.Fatalf("thread summary lost enclosing session hierarchy: %+v", threads)
	}
	sessions[0].Invocations[0].Model = "changed"
	*sessions[1].ParentID = "changed"
	if owned[0].Invocations[0].Model != "model-a" || *owned[1].ParentID != "opaque-root" {
		t.Fatal("grouping aliased the input rows or parent identity")
	}
}

func TestSessionsFromFactsRequiresExplicitInvocationOwner(t *testing.T) {
	if _, err := domain.SessionsFromFacts([]domain.SessionFact{{Invocations: []domain.InvocationFact{{Model: "model-a", Role: "coder"}}}}, nil, nil); err == nil {
		t.Fatal("unowned invocation rows must fail instead of receiving a fabricated session")
	}
}
