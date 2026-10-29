package main

import (
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

func TestMergeSessionsReplacesRefreshedOwnerSnapshot(t *testing.T) {
	start, firstEnd, nextEnd := "2026-10-25T09:00:00Z", "2026-10-25T09:00:07Z", "2026-10-25T09:01:02Z"
	row := domain.InvocationFact{
		Role: domain.RoleRootAgent, Model: "gpt-example", StartedAt: &start, CompletedAt: &firstEnd,
		Turns: 1, Tokens: domain.InvocationTokens{Input: 40, Output: 25},
	}
	active := row
	active.CompletedAt = nil
	grown := row
	grown.CompletedAt = &nextEnd
	grown.Turns = 2
	grown.Tokens = domain.InvocationTokens{Input: 240, Output: 75}
	changedModel := row
	changedModel.Model = "next-model"
	for _, tc := range []struct {
		name   string
		cached []domain.InvocationFact
		local  []domain.InvocationFact
	}{
		{"source growth", []domain.InvocationFact{row}, []domain.InvocationFact{grown}},
		{"active to completed", []domain.InvocationFact{active}, []domain.InvocationFact{row}},
		{"model change", []domain.InvocationFact{row}, []domain.InvocationFact{changedModel}},
		{"empty authoritative snapshot", []domain.InvocationFact{row}, []domain.InvocationFact{}},
		{"nil authoritative snapshot", []domain.InvocationFact{row}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cached := []domain.SessionFact{
				{ID: "owner-a", Source: domain.SourceCodex, Invocations: tc.cached},
				{ID: "owner-b", Source: domain.SourceOpenCode, User: "other-user", Invocations: []domain.InvocationFact{row}},
			}
			local := []domain.SessionFact{{ID: "owner-a", Invocations: tc.local}}
			want := []domain.SessionFact{
				{ID: "owner-a", Source: domain.SourceCodex, Invocations: tc.local},
				cached[1],
			}
			merged := mergeSessions(cached, local)
			if !reflect.DeepEqual(merged, want) {
				t.Fatalf("fresh owner snapshot was not authoritative: got=%+v want=%+v", merged, want)
			}
			if repeated := mergeSessions(merged, local); !reflect.DeepEqual(repeated, merged) {
				t.Fatalf("unchanged snapshot is not idempotent: before=%+v after=%+v", merged, repeated)
			}
			if unresolved := mergeSessions(cached, nil); !reflect.DeepEqual(unresolved, cached) {
				t.Fatalf("sessions without a fresh snapshot changed: %+v", unresolved)
			}
			if initial := mergeSessions(nil, want); !reflect.DeepEqual(initial, want) {
				t.Fatalf("initial snapshots changed: %+v", initial)
			}
		})
	}
}

func TestMergeSessionsPreservesRowMultiplicityAndOwnership(t *testing.T) {
	start, end := "2026-10-28T10:00:00Z", "2026-10-28T10:01:00Z"
	laterStart, laterEnd := "2026-10-28T11:00:00Z", "2026-10-28T11:01:00Z"
	row := domain.InvocationFact{Role: "coder", Model: "model-a", StartedAt: &start, CompletedAt: &end, Turns: 1}
	otherSessionRow := row
	otherSessionRow.Turns = 99
	retained := row
	retained.Model = "cached-model"
	cached := []domain.SessionFact{
		{ID: "owner-a", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{row, row, retained}},
		{ID: "owner-b", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{otherSessionRow}},
	}
	rows := []domain.InvocationFact{row, row, row, row, row, row, row}
	for i := range rows {
		rows[i].Turns = 10
	}
	rows[3].Model = "model-b"
	rows[4].Role = "reviewer"
	rows[5].StartedAt = &laterStart
	rows[6].CompletedAt = &laterEnd
	local := []domain.SessionFact{{ID: "owner-a", Invocations: rows}}

	merged := mergeSessions(cached, local)
	if len(merged) != 2 || merged[0].ID != "owner-a" || merged[0].Source != domain.SourceCodex || !reflect.DeepEqual(merged[0].Invocations, rows) {
		t.Fatalf("session-scoped rows were lost or collapsed: %+v", merged)
	}
	if !reflect.DeepEqual(merged[1], cached[1]) {
		t.Fatalf("another session's identical row was overwritten: %+v", merged[1])
	}
	if repeated := mergeSessions(merged, local); !reflect.DeepEqual(repeated, merged) {
		t.Fatalf("repeated refresh changed row multiplicity: before=%+v after=%+v", merged, repeated)
	}
	if !reflect.DeepEqual(cached[0].Invocations, []domain.InvocationFact{row, row, retained}) || !reflect.DeepEqual(local[0].Invocations, rows) {
		t.Fatal("merge mutated its input rows")
	}
}

func TestMergeToolsAndEvents_UnitAdversarial(t *testing.T) {
	t.Run("tools replace matching session facts and preserve distinct sessions", func(t *testing.T) {
		cached := []domain.ToolFact{
			{SessionID: "session-a", Name: "bash", Calls: 10, Errors: 2},
			{SessionID: "session-a", Name: "read", Calls: 5, Errors: 0},
		}
		local := []domain.ToolFact{
			{SessionID: "session-a", Name: "bash", Calls: 15, Errors: 1},
			{SessionID: "session-b", Name: "bash", Calls: 4, Errors: 1},
			{SessionID: "session-a", Name: "write", Calls: 3, Errors: 0},
		}

		merged := mergeTools(cached, local)
		if len(merged) != 4 {
			t.Fatalf("expected 4 session tool facts, got %d", len(merged))
		}
		var bashCalls, bashErrors int64
		for _, tool := range merged {
			if tool.Name == "bash" {
				bashCalls += tool.Calls
				bashErrors += tool.Errors
			}
		}
		if bashCalls != 19 || bashErrors != 2 {
			t.Errorf("expected cross-session bash calls 19, errors 2, got calls %d, errors %d", bashCalls, bashErrors)
		}
		repeated := mergeTools(merged, local)
		if len(repeated) != len(merged) {
			t.Fatalf("repeat merge duplicated facts: before=%d after=%d", len(merged), len(repeated))
		}
		var repeatedBashCalls int64
		for _, tool := range repeated {
			if tool.Name == "bash" {
				repeatedBashCalls += tool.Calls
			}
		}
		if repeatedBashCalls != bashCalls {
			t.Errorf("repeat merge changed bash calls: before=%d after=%d", bashCalls, repeatedBashCalls)
		}
	})

	t.Run("events merge deduplicates by composite key and appends novel events", func(t *testing.T) {
		cached := []domain.EventFact{
			{
				Timestamp: "2026-10-25T10:00:00Z",
				SessionID: "ses-1",
				Turn:      1,
				EventType: domain.EventTypeTurnMessage,
			},
		}
		local := []domain.EventFact{
			{
				Timestamp: "2026-10-25T10:00:00Z",
				SessionID: "ses-1",
				Turn:      1,
				EventType: domain.EventTypeTurnMessage, // duplicate
			},
			{
				Timestamp: "2026-10-25T10:00:05Z",
				SessionID: "ses-1",
				Turn:      2,
				EventType: domain.EventTypeToolCall, // novel
			},
		}

		merged := mergeEvents(cached, local)
		if len(merged) != 2 {
			t.Fatalf("expected 2 deduplicated events, got %d", len(merged))
		}
		if merged[1].Turn != 2 || merged[1].EventType != domain.EventTypeToolCall {
			t.Errorf("expected novel event at index 1, got %+v", merged[1])
		}
	})
}
