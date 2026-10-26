package main

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestMergeInvocations_UnitAdversarial(t *testing.T) {
	parentID := "ses_parent"
	cached := []domain.InvocationFact{
		{
			ID:             "ses_parent",
			Role:           domain.RoleRootAgent,
			Model:          "gpt-4o",
			Turns:          4,
			UserIterations: 2,
			Tokens:         domain.InvocationTokens{Input: 100, Output: 50},
			Compactions:    0,
		},
		{
			ID:             "ses_subagent",
			ParentID:       &parentID,
			Role:           domain.RoleUnassignedSubAgent,
			Model:          "claude-3-5-sonnet",
			Turns:          2,
			UserIterations: 0,
			Tokens:         domain.InvocationTokens{Input: 50, Output: 25},
			Compactions:    0,
		},
		{
			ID:             "ses_device_prev",
			Role:           domain.RoleRootAgent,
			Model:          "claude-3-opus",
			Turns:          10,
			UserIterations: 5,
			Tokens:         domain.InvocationTokens{Input: 500, Output: 200},
			Compactions:    1,
		},
	}

	t.Run("updates existing invocation in place preserving slice index", func(t *testing.T) {
		local := []domain.InvocationFact{
			{
				ID:             "ses_parent",
				Role:           domain.RoleRootAgent,
				Model:          "gpt-4o",
				Turns:          8,
				UserIterations: 4,
				Tokens:         domain.InvocationTokens{Input: 300, Output: 150},
				Compactions:    2,
			},
		}

		merged := mergeInvocations(cached, local)
		if len(merged) != 3 {
			t.Fatalf("expected 3 invocations, got %d", len(merged))
		}
		// Index 0 must be updated in place
		if merged[0].ID != "ses_parent" || merged[0].Turns != 8 || merged[0].UserIterations != 4 || merged[0].Compactions != 2 {
			t.Errorf("expected updated ses_parent at index 0, got %+v", merged[0])
		}
		// Other records must remain intact
		if merged[1].ID != "ses_subagent" || merged[2].ID != "ses_device_prev" {
			t.Errorf("expected preserved subsequent records, got %s, %s", merged[1].ID, merged[2].ID)
		}
	})

	t.Run("multi-model session on same ID preserves distinct model invocations", func(t *testing.T) {
		local := []domain.InvocationFact{
			{
				ID:             "ses_parent",
				Role:           domain.RoleRootAgent,
				Model:          "gemini-2.0-flash", // different model for same session ID
				Turns:          2,
				UserIterations: 0,
				Tokens:         domain.InvocationTokens{Input: 80, Output: 40},
			},
		}

		merged := mergeInvocations(cached, local)
		if len(merged) != 4 {
			t.Fatalf("expected 4 invocations (3 cached + 1 distinct model), got %d", len(merged))
		}
		if merged[3].ID != "ses_parent" || merged[3].Model != "gemini-2.0-flash" {
			t.Errorf("expected distinct model invocation appended, got %+v", merged[3])
		}
	})

	t.Run("handles empty cached or empty local cleanly", func(t *testing.T) {
		merged1 := mergeInvocations(nil, cached)
		if len(merged1) != len(cached) {
			t.Errorf("expected %d invocations from empty cached, got %d", len(cached), len(merged1))
		}

		merged2 := mergeInvocations(cached, nil)
		if len(merged2) != len(cached) {
			t.Errorf("expected %d invocations from empty local, got %d", len(cached), len(merged2))
		}
	})
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
