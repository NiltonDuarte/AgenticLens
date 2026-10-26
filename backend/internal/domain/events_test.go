package domain_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"agenticlens/internal/domain"
)

func TestSortEvents(t *testing.T) {
	t.Parallel()

	t.Run("nil slice returns allocated empty non-nil slice", func(t *testing.T) {
		t.Parallel()
		res := domain.SortEvents(nil)
		if res == nil {
			t.Fatal("expected non-nil slice, got nil")
		}
		if len(res) != 0 {
			t.Fatalf("expected 0 events, got %d", len(res))
		}
	})

	t.Run("empty slice returns allocated empty non-nil slice", func(t *testing.T) {
		t.Parallel()
		res := domain.SortEvents([]domain.EventFact{})
		if res == nil {
			t.Fatal("expected non-nil slice, got nil")
		}
		if len(res) != 0 {
			t.Fatalf("expected 0 events, got %d", len(res))
		}
	})

	t.Run("single event returns equivalent single element slice", func(t *testing.T) {
		t.Parallel()
		single := []domain.EventFact{
			{
				Timestamp: "2026-10-18T10:00:00Z",
				SessionID: "ses-1",
				Turn:      1,
				Role:      "coder",
				EventType: domain.EventTypeToolCall,
			},
		}
		res := domain.SortEvents(single)
		if len(res) != 1 {
			t.Fatalf("expected 1 event, got %d", len(res))
		}
		if !reflect.DeepEqual(res[0], single[0]) {
			t.Fatalf("expected %+v, got %+v", single[0], res[0])
		}
	})

	t.Run("pre-sorted slice maintains ascending order", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "ses-1"},
			{Timestamp: "2026-10-18T10:00:05Z", Turn: 2, SessionID: "ses-2"},
			{Timestamp: "2026-10-18T10:00:10Z", Turn: 3, SessionID: "ses-3"},
		}
		res := domain.SortEvents(input)
		if len(res) != 3 {
			t.Fatalf("expected 3 events, got %d", len(res))
		}
		for i := range res {
			if res[i].Timestamp != input[i].Timestamp || res[i].Turn != input[i].Turn {
				t.Errorf("mismatch at index %d: expected %+v, got %+v", i, input[i], res[i])
			}
		}
	})

	t.Run("reversed slice sorts to ascending timestamps", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:10Z", Turn: 3, SessionID: "ses-3"},
			{Timestamp: "2026-10-18T10:00:05Z", Turn: 2, SessionID: "ses-2"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "ses-1"},
		}
		res := domain.SortEvents(input)
		if len(res) != 3 {
			t.Fatalf("expected 3 events, got %d", len(res))
		}
		if res[0].Timestamp != "2026-10-18T10:00:00Z" || res[1].Timestamp != "2026-10-18T10:00:05Z" || res[2].Timestamp != "2026-10-18T10:00:10Z" {
			t.Errorf("unexpected ordering: %+v", res)
		}
	})

	t.Run("identical timestamps sort by ascending turn number", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 4, SessionID: "ses-4"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "ses-1"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 3, SessionID: "ses-3"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 2, SessionID: "ses-2"},
		}
		res := domain.SortEvents(input)
		if len(res) != 4 {
			t.Fatalf("expected 4 events, got %d", len(res))
		}
		for i := 0; i < 4; i++ {
			expectedTurn := int64(i + 1)
			if res[i].Turn != expectedTurn {
				t.Errorf("expected turn %d at index %d, got %d", expectedTurn, i, res[i].Turn)
			}
		}
	})

	t.Run("identical timestamps and turn numbers preserve stable order", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "first"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "second"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "third"},
		}
		res := domain.SortEvents(input)
		if len(res) != 3 {
			t.Fatalf("expected 3 events, got %d", len(res))
		}
		if res[0].SessionID != "first" || res[1].SessionID != "second" || res[2].SessionID != "third" {
			t.Errorf("expected stable order preserved: got %s, %s, %s", res[0].SessionID, res[1].SessionID, res[2].SessionID)
		}
	})

	t.Run("immutability - input slice remains unmodified", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:10Z", Turn: 3, SessionID: "ses-3"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "ses-1"},
		}
		inputCopy := make([]domain.EventFact, len(input))
		copy(inputCopy, input)

		res := domain.SortEvents(input)

		if !reflect.DeepEqual(input, inputCopy) {
			t.Fatalf("input slice was mutated: original %+v, modified %+v", inputCopy, input)
		}
		if &res[0] == &input[0] {
			t.Fatal("result slice shares underlying array with input slice")
		}
	})
}

func TestSortEvents_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("negative and zero turns sort in ascending numerical order", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 5, SessionID: "turn-5"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: -10, SessionID: "turn-neg10"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 0, SessionID: "turn-0"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: -1, SessionID: "turn-neg1"},
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 2, SessionID: "turn-2"},
		}

		res := domain.SortEvents(input)
		if len(res) != 5 {
			t.Fatalf("expected 5 events, got %d", len(res))
		}

		expectedTurns := []int64{-10, -1, 0, 2, 5}
		for i, expected := range expectedTurns {
			if res[i].Turn != expected {
				t.Errorf("at index %d: expected turn %d, got %d", i, expected, res[i].Turn)
			}
		}
	})

	t.Run("empty timestamp strings sort lexicographically before non-empty timestamps", func(t *testing.T) {
		t.Parallel()
		input := []domain.EventFact{
			{Timestamp: "2026-10-18T10:00:00Z", Turn: 1, SessionID: "normal-ts"},
			{Timestamp: "", Turn: 2, SessionID: "empty-ts-2"},
			{Timestamp: "", Turn: 1, SessionID: "empty-ts-1"},
		}

		res := domain.SortEvents(input)
		if len(res) != 3 {
			t.Fatalf("expected 3 events, got %d", len(res))
		}

		if res[0].SessionID != "empty-ts-1" || res[1].SessionID != "empty-ts-2" || res[2].SessionID != "normal-ts" {
			t.Errorf("unexpected ordering for empty timestamps: %+v", res)
		}
	})

	t.Run("complete metadata and optional pointer fields preserved without truncation", func(t *testing.T) {
		t.Parallel()
		tool := "bash"
		errMsg := "exit status 1"
		dur := int64(1500)

		input := []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:01Z",
				SessionID:     "ses-detail",
				Turn:          3,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      &tool,
				Status:        "error",
				ErrorMessage:  &errMsg,
				DurationMs:    &dur,
				ContextTokens: 42000,
			},
			{
				Timestamp:     "2026-10-18T10:00:00Z",
				SessionID:     "ses-detail-0",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeTurnMessage,
				ToolName:      nil,
				Status:        "success",
				ErrorMessage:  nil,
				DurationMs:    nil,
				ContextTokens: 12000,
			},
		}

		res := domain.SortEvents(input)
		if len(res) != 2 {
			t.Fatalf("expected 2 events, got %d", len(res))
		}

		first := res[0]
		if first.SessionID != "ses-detail-0" || first.ToolName != nil || first.ErrorMessage != nil {
			t.Errorf("unexpected fields in first sorted event: %+v", first)
		}

		second := res[1]
		if second.SessionID != "ses-detail" || second.ToolName == nil || *second.ToolName != "bash" {
			t.Errorf("tool name not preserved: %+v", second)
		}
		if second.ErrorMessage == nil || *second.ErrorMessage != "exit status 1" {
			t.Errorf("error message not preserved: %+v", second)
		}
		if second.DurationMs == nil || *second.DurationMs != 1500 {
			t.Errorf("duration not preserved: %+v", second)
		}
		if second.ContextTokens != 42000 {
			t.Errorf("context tokens not preserved: %d", second.ContextTokens)
		}
	})

	t.Run("large collection with repeated timestamps and turns verifies stable sort order", func(t *testing.T) {
		t.Parallel()
		count := 200
		input := make([]domain.EventFact, count)
		for i := 0; i < count; i++ {
			// Deliberately generate collisions across 5 timestamps and 4 turns
			ts := "2026-10-18T10:00:0" + string(rune('0'+(i%5))) + "Z"
			turn := int64(i % 4)
			input[i] = domain.EventFact{
				Timestamp: ts,
				Turn:      turn,
				SessionID: "event-" + string(rune('A'+(i%26))),
			}
		}

		res := domain.SortEvents(input)
		if len(res) != count {
			t.Fatalf("expected %d events, got %d", count, len(res))
		}

		// Verify non-decreasing order: Timestamp[i] <= Timestamp[i+1], and if equal Turn[i] <= Turn[i+1]
		for i := 0; i < count-1; i++ {
			if res[i].Timestamp > res[i+1].Timestamp {
				t.Fatalf("sort violation at index %d: %s > %s", i, res[i].Timestamp, res[i+1].Timestamp)
			}
			if res[i].Timestamp == res[i+1].Timestamp && res[i].Turn > res[i+1].Turn {
				t.Fatalf("turn sort violation at index %d: turn %d > %d", i, res[i].Turn, res[i+1].Turn)
			}
		}
	})
}

func TestPersistAnomalySummariesDropsEventsAndEvidenceButKeepsFacts(t *testing.T) {
	tool := "bash"
	sessionID := "s1"
	facts := domain.AgenticFactLedger{Source: domain.SourceCodex, Sessions: []domain.SessionFact{{
		ID:          sessionID,
		Invocations: []domain.InvocationFact{{ID: sessionID, Role: "orchestrator"}},
		Tools:       []domain.SessionToolFact{{Name: tool, Calls: 6, Errors: 6}},
		Events:      []domain.SessionEventFact{{EventID: "e1", Timestamp: "2026-01-01T00:00:00Z", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"}},
	}}}
	persisted, err := domain.PersistAnomalySummaries(facts)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Source != domain.SourceCodex || len(persisted.EventFacts()) != 0 {
		t.Fatalf("source/events mismatch: %#v", persisted)
	}
	if len(persisted.InvocationFacts()) != 1 || len(persisted.ToolFacts()) != 1 || persisted.ToolFacts()[0].Calls != 6 {
		t.Fatalf("invocation or tool totals changed: %#v", persisted)
	}
	if len(persisted.AnomalyFacts()) == 0 {
		t.Fatal("anomaly summary was not retained")
	}
	for _, anomaly := range persisted.AnomalyFacts() {
		if len(anomaly.Evidence) != 0 {
			t.Fatalf("evidence refs persisted: %#v", anomaly.Evidence)
		}
	}
}

func TestPersistAnomalySummariesRefreshesNewSessionsAndAggregates(t *testing.T) {
	tool := "bash"
	oldSummary := domain.Anomaly{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityWarning}
	newEvents := []domain.SessionEventFact{
		{EventID: "new-1", Timestamp: "2026-10-26T10:00:01Z", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{EventID: "new-2", Timestamp: "2026-10-26T10:00:02Z", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{EventID: "new-3", Timestamp: "2026-10-26T10:00:03Z", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
	}
	merged := domain.AgenticFactLedger{Sessions: []domain.SessionFact{
		{ID: "old-root", Invocations: []domain.InvocationFact{{ID: "old-root", Role: "orchestrator"}}, Tools: []domain.SessionToolFact{{Name: tool, Calls: 6, Errors: 5}}, Anomalies: []domain.Anomaly{oldSummary}},
		{ID: "new-root", Invocations: []domain.InvocationFact{{ID: "new-root", Role: "orchestrator"}}, Tools: []domain.SessionToolFact{{Name: tool, Calls: 2, Errors: 1}}, Events: newEvents, Anomalies: []domain.Anomaly{}},
	}}
	persisted, err := domain.PersistAnomalySummaries(merged)
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted.Sessions[0].Anomalies) != 1 {
		t.Fatalf("cached session summary was lost: %#v", persisted.Sessions)
	}
	var retryFound bool
	for _, anomaly := range persisted.Sessions[1].Anomalies {
		if anomaly.RuleID == domain.RuleToolRetryStreak && anomaly.Context.Count != nil && *anomaly.Context.Count == 3 {
			retryFound = true
		}
	}
	if !retryFound {
		t.Fatalf("new source events did not generate the new session summary: %#v", persisted.Sessions[1].Anomalies)
	}
	var aggregate *domain.Anomaly
	for i := range persisted.Anomalies {
		if persisted.Anomalies[i].RuleID == domain.RuleToolHighFailureRate {
			aggregate = &persisted.Anomalies[i]
		}
	}
	if aggregate == nil || aggregate.Context.TotalCalls == nil || *aggregate.Context.TotalCalls != 8 || aggregate.Context.Count == nil || *aggregate.Context.Count != 6 {
		t.Fatalf("aggregate summary was not recomputed from merged tool facts: %#v", persisted.Anomalies)
	}
	if len(persisted.EventFacts()) != 0 {
		t.Fatalf("event rows persisted: %#v", persisted.EventFacts())
	}
	projected := domain.DeriveMetricsLedger(merged, domain.DefaultProjectConfig())
	projectedOld, projectedRetry, projectedAggregate := false, false, false
	for _, anomaly := range projected.Anomalies {
		projectedOld = projectedOld || anomaly.RuleID == domain.RuleContextSaturationSpike
		projectedRetry = projectedRetry || anomaly.RuleID == domain.RuleToolRetryStreak
		projectedAggregate = projectedAggregate || (anomaly.RuleID == domain.RuleToolHighFailureRate && anomaly.Context.TotalCalls != nil && *anomaly.Context.TotalCalls == 8)
	}
	if !projectedOld || !projectedRetry || !projectedAggregate {
		t.Fatalf("projection lost cached/new/aggregate anomaly coverage: %#v", projected.Anomalies)
	}
}

func TestWorkspaceEventsResponseSerialization(t *testing.T) {
	t.Parallel()

	prompt := "Write tests for events"
	reasoning := "Need to ensure sanitization"
	toolInput := `{"filePath": "events.go"}`
	toolOutput := "file contents"
	duration := int64(150)

	tests := []struct {
		name         string
		input        domain.WorkspaceEventsResponse
		expectedJSON string
	}{
		{
			name: "hydrated local sqlite response with full diagnostic context",
			input: domain.WorkspaceEventsResponse{
				TicketID:   "AL-33",
				IsHydrated: true,
				Events: []domain.EventFact{
					{
						Timestamp:        "2026-10-19T10:00:00Z",
						SessionID:        "ses-123",
						Turn:             1,
						Role:             "coder",
						EventType:        domain.EventTypeToolCall,
						ToolName:         func() *string { s := "edit"; return &s }(),
						Status:           "completed",
						DurationMs:       &duration,
						ContextTokens:    4500,
						PromptContent:    &prompt,
						ReasoningContent: &reasoning,
						ToolInput:        &toolInput,
						ToolOutput:       &toolOutput,
					},
				},
				DataSource: domain.DataSourceLocalSQLite,
			},
		},
		{
			name: "unhydrated committed ledger response with empty events",
			input: domain.WorkspaceEventsResponse{
				TicketID:   "AL-33",
				IsHydrated: false,
				Events:     []domain.EventFact{},
				DataSource: domain.DataSourceCommittedLedger,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("failed to marshal WorkspaceEventsResponse: %v", err)
			}

			var decoded domain.WorkspaceEventsResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal WorkspaceEventsResponse: %v", err)
			}

			if decoded.TicketID != tt.input.TicketID {
				t.Errorf("TicketID mismatch: expected %q, got %q", tt.input.TicketID, decoded.TicketID)
			}
			if decoded.IsHydrated != tt.input.IsHydrated {
				t.Errorf("IsHydrated mismatch: expected %v, got %v", tt.input.IsHydrated, decoded.IsHydrated)
			}
			if decoded.DataSource != tt.input.DataSource {
				t.Errorf("DataSource mismatch: expected %q, got %q", tt.input.DataSource, decoded.DataSource)
			}
			if len(decoded.Events) != len(tt.input.Events) {
				t.Fatalf("Events count mismatch: expected %d, got %d", len(tt.input.Events), len(decoded.Events))
			}
			if len(decoded.Events) > 0 {
				ev := decoded.Events[0]
				if ev.PromptContent == nil || *ev.PromptContent != prompt {
					t.Errorf("PromptContent mismatch: expected %q, got %v", prompt, ev.PromptContent)
				}
				if ev.ReasoningContent == nil || *ev.ReasoningContent != reasoning {
					t.Errorf("ReasoningContent mismatch: expected %q, got %v", reasoning, ev.ReasoningContent)
				}
				if ev.ToolInput == nil || *ev.ToolInput != toolInput {
					t.Errorf("ToolInput mismatch: expected %q, got %v", toolInput, ev.ToolInput)
				}
				if ev.ToolOutput == nil || *ev.ToolOutput != toolOutput {
					t.Errorf("ToolOutput mismatch: expected %q, got %v", toolOutput, ev.ToolOutput)
				}
			}
		})
	}
}

func ptrString(v string) *string { return &v }
func ptrInt64(v int64) *int64    { return &v }

func TestWorkspaceEventsResponse_AdversarialBoundaries(t *testing.T) {
	t.Parallel()

	// Assert canonical provenance constants
	if domain.DataSourceLocalSQLite != "local_sqlite" {
		t.Errorf("expected DataSourceLocalSQLite to be %q, got %q", "local_sqlite", domain.DataSourceLocalSQLite)
	}
	if domain.DataSourceCommittedLedger != "committed_ledger" {
		t.Errorf("expected DataSourceCommittedLedger to be %q, got %q", "committed_ledger", domain.DataSourceCommittedLedger)
	}

	zeroDuration := int64(0)
	negativeDuration := int64(-50)
	largeDuration := int64(999999999)
	emptyString := ""
	unicodePayload := "Testing unicode: \u00e9\u00e0\u00fc \u4e16\u754c"

	tests := []struct {
		name     string
		resp     domain.WorkspaceEventsResponse
		validate func(t *testing.T, marshaled string, unmarshaled domain.WorkspaceEventsResponse)
	}{
		{
			name: "zero tokens and zero duration boundary",
			resp: domain.WorkspaceEventsResponse{
				TicketID:   "AL-ZERO",
				IsHydrated: true,
				Events: []domain.EventFact{
					{
						Timestamp:     "2026-10-19T00:00:00Z",
						SessionID:     "s-zero",
						Turn:          0,
						Role:          "coder",
						EventType:     domain.EventTypeToolCall,
						Status:        "completed",
						DurationMs:    &zeroDuration,
						ContextTokens: 0,
					},
				},
				DataSource: domain.DataSourceLocalSQLite,
			},
			validate: func(t *testing.T, raw string, decoded domain.WorkspaceEventsResponse) {
				if len(decoded.Events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(decoded.Events))
				}
				ev := decoded.Events[0]
				if ev.ContextTokens != 0 {
					t.Errorf("expected context tokens 0, got %d", ev.ContextTokens)
				}
				if ev.DurationMs == nil || *ev.DurationMs != 0 {
					t.Errorf("expected duration 0, got %v", ev.DurationMs)
				}
				if ev.PromptContent != nil || ev.ReasoningContent != nil || ev.ToolInput != nil || ev.ToolOutput != nil {
					t.Errorf("expected omitted pointer fields to decode as nil")
				}
			},
		},
		{
			name: "negative duration and large context tokens",
			resp: domain.WorkspaceEventsResponse{
				TicketID:   "AL-EXTREME",
				IsHydrated: false,
				Events: []domain.EventFact{
					{
						Timestamp:        "2026-10-19T23:59:59Z",
						SessionID:        "s-extreme",
						Turn:             9999,
						Role:             "tester",
						EventType:        domain.EventTypeTurnMessage,
						Status:           "completed",
						DurationMs:       &negativeDuration,
						ContextTokens:    2000000,
						PromptContent:    &unicodePayload,
						ReasoningContent: &emptyString,
					},
				},
				DataSource: domain.DataSourceCommittedLedger,
			},
			validate: func(t *testing.T, raw string, decoded domain.WorkspaceEventsResponse) {
				if len(decoded.Events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(decoded.Events))
				}
				ev := decoded.Events[0]
				if ev.ContextTokens != 2000000 {
					t.Errorf("expected 2000000 context tokens, got %d", ev.ContextTokens)
				}
				if ev.DurationMs == nil || *ev.DurationMs != -50 {
					t.Errorf("expected duration -50, got %v", ev.DurationMs)
				}
				if ev.PromptContent == nil || *ev.PromptContent != unicodePayload {
					t.Errorf("expected unicode payload, got %v", ev.PromptContent)
				}
				if ev.ReasoningContent == nil || *ev.ReasoningContent != "" {
					t.Errorf("expected empty string reasoning content, got %v", ev.ReasoningContent)
				}
			},
		},
		{
			name: "nil events slice marshals cleanly",
			resp: domain.WorkspaceEventsResponse{
				TicketID:   "AL-NIL-SLICE",
				IsHydrated: false,
				Events:     nil,
				DataSource: domain.DataSourceCommittedLedger,
			},
			validate: func(t *testing.T, raw string, decoded domain.WorkspaceEventsResponse) {
				if decoded.TicketID != "AL-NIL-SLICE" {
					t.Errorf("expected ticket ID AL-NIL-SLICE, got %s", decoded.TicketID)
				}
				if decoded.IsHydrated {
					t.Errorf("expected IsHydrated == false")
				}
			},
		},
		{
			name: "large duration and compaction event fact",
			resp: domain.WorkspaceEventsResponse{
				TicketID:   "AL-COMPACTION",
				IsHydrated: true,
				Events: []domain.EventFact{
					{
						Timestamp:     "2026-10-19T12:00:00Z",
						SessionID:     "s-comp",
						Turn:          42,
						Role:          "orchestrator",
						EventType:     domain.EventTypeCompaction,
						Status:        "completed",
						DurationMs:    &largeDuration,
						ContextTokens: 128000,
					},
				},
				DataSource: domain.DataSourceLocalSQLite,
			},
			validate: func(t *testing.T, raw string, decoded domain.WorkspaceEventsResponse) {
				if len(decoded.Events) != 1 {
					t.Fatalf("expected 1 event, got %d", len(decoded.Events))
				}
				ev := decoded.Events[0]
				if ev.EventType != domain.EventTypeCompaction {
					t.Errorf("expected event type compaction, got %s", ev.EventType)
				}
				if ev.DurationMs == nil || *ev.DurationMs != largeDuration {
					t.Errorf("expected large duration, got %v", ev.DurationMs)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.resp)
			if err != nil {
				t.Fatalf("marshaling failed: %v", err)
			}

			var decoded domain.WorkspaceEventsResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshaling failed: %v", err)
			}

			tt.validate(t, string(data), decoded)
		})
	}
}
