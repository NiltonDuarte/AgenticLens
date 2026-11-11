package domain_test

import (
	"encoding/json"
	"fmt"
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
	facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{
		ID:          sessionID,
		Source:      domain.SourceCodex,
		Invocations: []domain.InvocationFact{{Role: "orchestrator"}},
		Tools:       []domain.SessionToolFact{{Name: tool, Calls: 6, Errors: 6}},
	}}}
	persisted, err := domain.PersistSourceAnomalySummaries(facts, domain.SourceEvents{Events: []domain.EventFact{{SessionID: sessionID, EventID: "e1", Timestamp: "2026-01-01T00:00:00Z", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"}}, CompleteSessionIDs: []string{sessionID}}, fixtureAnomalyPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted.Sessions) == 0 || persisted.Sessions[0].Source != domain.SourceCodex {
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
	newEvents := []domain.EventFact{
		{SessionID: "new-root", EventID: "new-1", Timestamp: "2026-10-26T10:00:01Z", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{SessionID: "new-root", EventID: "new-2", Timestamp: "2026-10-26T10:00:02Z", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{SessionID: "new-root", EventID: "new-3", Timestamp: "2026-10-26T10:00:03Z", Turn: 3, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
	}
	oldCalls, oldErrors, oldRate := int64(6), int64(5), 83.3
	oldAggregate := domain.Anomaly{RuleID: domain.RuleToolHighFailureRate, Severity: domain.SeverityCritical, Context: domain.AnomalyContext{ToolName: &tool, TotalCalls: &oldCalls, Count: &oldErrors, FailureRatePct: &oldRate}}
	merged := domain.AgenticFactLedger{Anomalies: []domain.Anomaly{oldAggregate}, Sessions: []domain.SessionFact{
		{ID: "old-root", Invocations: []domain.InvocationFact{{Role: "orchestrator"}}, Tools: []domain.SessionToolFact{{Name: tool, Calls: 6, Errors: 5}}, Anomalies: []domain.Anomaly{oldSummary}},
		{ID: "new-root", Invocations: []domain.InvocationFact{{Role: "orchestrator"}}, Tools: []domain.SessionToolFact{{Name: tool, Calls: 3, Errors: 3}}, Anomalies: []domain.Anomaly{}},
	}}
	persisted, err := domain.PersistSourceAnomalySummaries(merged, domain.SourceEvents{Events: newEvents, CompleteSessionIDs: []string{"new-root"}}, fixtureAnomalyPolicy())
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
	if aggregate == nil || aggregate.Context.TotalCalls == nil || *aggregate.Context.TotalCalls != 6 || aggregate.Context.Count == nil || *aggregate.Context.Count != 5 {
		t.Fatalf("partial source erased historical root/group summary: %#v", persisted.Anomalies)
	}
	projected := domain.DeriveMetricsLedger(persisted, testMetricsProjectConfig())
	projectedOld, projectedRetry, projectedAggregate := false, false, false
	for _, anomaly := range projected.Anomalies {
		projectedOld = projectedOld || anomaly.RuleID == domain.RuleContextSaturationSpike
		projectedRetry = projectedRetry || anomaly.RuleID == domain.RuleToolRetryStreak
		projectedAggregate = projectedAggregate || (anomaly.RuleID == domain.RuleToolHighFailureRate && anomaly.Context.TotalCalls != nil && *anomaly.Context.TotalCalls == 6)
	}
	if !projectedOld || !projectedRetry || !projectedAggregate {
		t.Fatalf("projection lost cached/new/aggregate anomaly coverage: %#v", projected.Anomalies)
	}
	complete, err := domain.PersistSourceAnomalySummaries(persisted, domain.SourceEvents{Events: newEvents, CompleteSessionIDs: []string{"old-root", "new-root"}}, fixtureAnomalyPolicy())
	if err != nil {
		t.Fatal(err)
	}
	var completeAggregate *domain.Anomaly
	for i := range complete.Anomalies {
		if complete.Anomalies[i].RuleID == domain.RuleToolHighFailureRate {
			completeAggregate = &complete.Anomalies[i]
		}
	}
	if completeAggregate == nil || completeAggregate.Context.TotalCalls == nil || *completeAggregate.Context.TotalCalls != 9 || completeAggregate.Context.Count == nil || *completeAggregate.Context.Count != 8 {
		t.Fatalf("complete source did not replace aggregate from whole owner population: %#v", complete.Anomalies)
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
				DataSource: domain.DataSourceLocalSQLite, Availability: domain.TraceComplete, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{"ses-123"}, CompleteSessionIDs: []string{"ses-123"}, Diagnostics: []string{}},
			},
		},
		{
			name: "unavailable source response with empty events",
			input: domain.WorkspaceEventsResponse{
				TicketID:   "AL-33",
				IsHydrated: false,
				Events:     []domain.EventFact{},
				DataSource: domain.DataSourceUnavailable, Availability: domain.TraceUnavailable, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{}, CompleteSessionIDs: []string{}, Diagnostics: []string{}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.input.Validate(); err != nil {
				t.Fatal(err)
			}
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

func TestReconcileEventsEnrichesPartialTraceWithoutLoss(t *testing.T) {
	first := domain.EventFact{EventID: "call", SessionID: "owner", Timestamp: "2026-10-29T10:00:00Z", Turn: 2, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: ptrString("bash"), Status: "error", ContextTokens: 1200, DurationMs: ptrInt64(10)}
	unmatched := first
	unmatched.EventID, unmatched.Turn = "unmatched", 3
	local := first
	local.MessageID, local.DurationMs = ptrString("message"), ptrInt64(0)
	local.PromptContent, local.ReasoningContent = ptrString("Run checks"), ptrString("Inspect failures")
	local.ToolInput, local.ToolOutput, local.ErrorMessage = ptrString("make"), ptrString("failed"), ptrString("exit 1")
	additional := local
	additional.EventID, additional.SessionID, additional.Turn = "local-only", "other", 1
	baseline, enrichment := []domain.EventFact{unmatched, first}, []domain.EventFact{additional, local}
	beforeBaseline, beforeEnrichment := append([]domain.EventFact{}, baseline...), append([]domain.EventFact{}, enrichment...)
	merged := domain.ReconcileEvents(baseline, enrichment)
	if !reflect.DeepEqual(merged, []domain.EventFact{additional, local, unmatched}) {
		t.Fatalf("partial enrichment lost facts/details: %#v", merged)
	}
	if !reflect.DeepEqual(domain.ReconcileEvents(merged, enrichment), merged) || !reflect.DeepEqual(domain.ReconcileEvents(baseline, merged), merged) {
		t.Fatal("reconciliation duplicated rows when repeated at the HTTP boundary")
	}
	if !reflect.DeepEqual(baseline, beforeBaseline) || !reflect.DeepEqual(enrichment, beforeEnrichment) {
		t.Fatal("reconciliation mutated an input")
	}

	t.Run("authoritative source projection differs from ordinary enrichment", func(t *testing.T) {
		for _, role := range []domain.AgentRole{"", "Build"} {
			source := local
			source.Role = role
			ordinary := domain.ReconcileEvents([]domain.EventFact{first}, []domain.EventFact{source})
			authoritative := domain.ReconcileSourceEvents([]domain.EventFact{first}, []domain.EventFact{source})
			if len(ordinary) != 1 || ordinary[0].Role != "coder" {
				t.Fatalf("ordinary enrichment changed baseline persona: %+v", ordinary)
			}
			if len(authoritative) != 1 || authoritative[0].Role != role || authoritative[0].ToolOutput == nil {
				t.Fatalf("source projection lost persona or payload: %+v", authoritative)
			}
			if first.Role != "coder" || source.Role != role {
				t.Fatal("projection mutated source or baseline")
			}
		}
	})

}

func TestReconcileEventsIdentityAndAmbiguity(t *testing.T) {
	base := domain.EventFact{EventID: "id", SessionID: "session", Timestamp: "2026-10-29T10:00:00Z", Turn: 1, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: ptrString("bash"), Status: "error", MessageID: ptrString("message")}
	bare := base
	bare.EventID = ""
	local := base
	local.ToolOutput = ptrString("details")
	changed := func(change func(*domain.EventFact)) domain.EventFact { event := local; change(&event); return event }
	tests := []struct {
		name                 string
		baseline, enrichment []domain.EventFact
		count                int
	}{
		{"empty", nil, nil, 0},
		{"baseline only", []domain.EventFact{base}, nil, 1},
		{"local only", nil, []domain.EventFact{local}, 1},
		{"same ID", []domain.EventFact{base}, []domain.EventFact{local}, 1},
		{"missing ID unique", []domain.EventFact{bare}, []domain.EventFact{local}, 1},
		{"different IDs", []domain.EventFact{base}, []domain.EventFact{changed(func(e *domain.EventFact) { e.EventID = "different" })}, 2},
		{"different sessions", []domain.EventFact{base}, []domain.EventFact{changed(func(e *domain.EventFact) { e.SessionID = "other" })}, 2},
		{"different timestamp", []domain.EventFact{bare}, []domain.EventFact{changed(func(e *domain.EventFact) { e.Timestamp = "later" })}, 2},
		{"different turn", []domain.EventFact{bare}, []domain.EventFact{changed(func(e *domain.EventFact) { e.Turn = 2 })}, 2},
		{"different event type", []domain.EventFact{bare}, []domain.EventFact{changed(func(e *domain.EventFact) { e.EventType = domain.EventTypeTurnMessage })}, 2},
		{"different tool", []domain.EventFact{bare}, []domain.EventFact{changed(func(e *domain.EventFact) { e.ToolName = ptrString("read") })}, 2},
		{"different message", []domain.EventFact{bare}, []domain.EventFact{changed(func(e *domain.EventFact) { e.MessageID = ptrString("other") })}, 2},
		{"ambiguous baseline", []domain.EventFact{bare, bare}, []domain.EventFact{local}, 3},
		{"ambiguous local", []domain.EventFact{bare}, []domain.EventFact{local, changed(func(e *domain.EventFact) { e.EventID = "other" })}, 3},
		{"identical fresh row ambiguous against baseline", []domain.EventFact{bare, bare}, []domain.EventFact{bare}, 3},
		{"identical fresh rows ambiguous against one baseline", []domain.EventFact{bare}, []domain.EventFact{bare, bare}, 3},
		{"identical ambiguous independent collections", []domain.EventFact{bare, bare}, []domain.EventFact{bare, bare}, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merged := domain.ReconcileEvents(tc.baseline, tc.enrichment)
			if merged == nil || len(merged) != tc.count {
				t.Fatalf("expected %d distinct events, got %#v", tc.count, merged)
			}
			sourceMerged := domain.ReconcileSourceEvents(tc.baseline, tc.enrichment)
			if sourceMerged == nil || len(sourceMerged) != tc.count {
				t.Fatalf("source projection changed occurrence matching: %#v", sourceMerged)
			}
		})
	}
}

func TestWorkspaceEventsResponse_AdversarialBoundaries(t *testing.T) {
	t.Parallel()

	// Assert canonical provenance constants
	if domain.DataSourceLocalSQLite != "local_sqlite" {
		t.Errorf("expected DataSourceLocalSQLite to be %q, got %q", "local_sqlite", domain.DataSourceLocalSQLite)
	}
	if domain.DataSourceUnavailable != "unavailable" {
		t.Errorf("invalid unavailable source discriminator: %q", domain.DataSourceUnavailable)
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
				DataSource: domain.DataSourceLocalSQLite, Availability: domain.TraceComplete, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{"s-zero"}, CompleteSessionIDs: []string{"s-zero"}, Diagnostics: []string{}},
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
				IsHydrated: true,
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
				DataSource: domain.DataSourceLocalSQLite, Availability: domain.TraceComplete, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{"s-extreme"}, CompleteSessionIDs: []string{"s-extreme"}, Diagnostics: []string{}},
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
				DataSource: domain.DataSourceUnavailable, Availability: domain.TraceUnavailable, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{}, CompleteSessionIDs: []string{}, Diagnostics: []string{}},
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
				DataSource: domain.DataSourceLocalSQLite, Availability: domain.TraceComplete, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{"s-comp"}, CompleteSessionIDs: []string{"s-comp"}, Diagnostics: []string{}},
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

			if err := tt.resp.Validate(); err != nil {
				t.Fatal(err)
			}
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

func TestTransientTraceAvailabilityRequiresHonestOwnerCoverage(t *testing.T) {
	base := domain.WorkspaceEventsResponse{Availability: domain.TraceComplete, IsHydrated: true, DataSource: domain.DataSourceLocalCodex, Events: []domain.EventFact{}, Coverage: domain.TraceCoverage{RequiredSessionIDs: []string{"owner"}, CompleteSessionIDs: []string{"owner"}, Diagnostics: []string{}}}
	cases := []struct {
		name   string
		change func(*domain.WorkspaceEventsResponse)
		valid  bool
	}{
		{"complete empty", func(*domain.WorkspaceEventsResponse) {}, true},
		{"missing owner read", func(r *domain.WorkspaceEventsResponse) { r.Coverage.CompleteSessionIDs = nil }, false},
		{"unknown complete owner", func(r *domain.WorkspaceEventsResponse) { r.Coverage.CompleteSessionIDs = []string{"other"} }, false},
		{"duplicate owner", func(r *domain.WorkspaceEventsResponse) { r.Coverage.RequiredSessionIDs = []string{"owner", "owner"} }, false},
		{"unrecorded event", func(r *domain.WorkspaceEventsResponse) { r.Events = []domain.EventFact{{SessionID: "other"}} }, false},
		{"committed fallback", func(r *domain.WorkspaceEventsResponse) { r.DataSource = domain.DataSourceCommittedLedger }, false},
		{"partial empty successful owner", func(r *domain.WorkspaceEventsResponse) {
			r.Availability = domain.TracePartial
			r.Coverage.RequiredSessionIDs = []string{"owner", "missing"}
			r.Coverage.Diagnostics = []string{"missing: unavailable"}
		}, true},
		{"partial rows incomplete owner", func(r *domain.WorkspaceEventsResponse) {
			r.Availability = domain.TracePartial
			r.Events = []domain.EventFact{{SessionID: "owner"}}
			r.Coverage.CompleteSessionIDs = nil
			r.Coverage.Diagnostics = []string{"owner: incomplete"}
		}, true},
		{"partial without evidence", func(r *domain.WorkspaceEventsResponse) {
			r.Availability = domain.TracePartial
			r.Coverage.CompleteSessionIDs = nil
		}, false},
		{"unavailable", func(r *domain.WorkspaceEventsResponse) {
			r.Availability = domain.TraceUnavailable
			r.IsHydrated = false
			r.DataSource = domain.DataSourceUnavailable
			r.Coverage.CompleteSessionIDs = nil
			r.Coverage.Diagnostics = []string{"owner: unavailable"}
		}, true},
		{"unavailable claims native source", func(r *domain.WorkspaceEventsResponse) {
			r.Availability = domain.TraceUnavailable
			r.IsHydrated = false
			r.Coverage.CompleteSessionIDs = nil
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := base
			tc.change(&r)
			if err := r.Validate(); (err == nil) != tc.valid {
				t.Fatalf("valid=%t err=%v response=%+v", tc.valid, err, r)
			}
		})
	}
}

func TestSourceSummaryRefreshRequiresCompleteChildGroup(t *testing.T) {
	parent := "parent"
	count := int64(3)
	coder, tester := domain.AgentRole("coder"), domain.AgentRole("tester")
	rework := domain.Anomaly{RuleID: domain.RuleSubagentReworkLoop, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{Role: &coder, SecondaryRole: &tester, Count: &count}}
	facts := domain.AgenticFactLedger{Sessions: []domain.SessionFact{{ID: parent, Anomalies: []domain.Anomaly{rework}}, {ID: "child", ParentID: &parent}}, Anomalies: []domain.Anomaly{{RuleID: domain.RuleContextSaturationSpike, Severity: domain.SeverityWarning}}}
	for _, complete := range [][]string{nil, {parent}} {
		t.Run(fmt.Sprintf("coverage=%v", complete), func(t *testing.T) {
			got := domain.RefreshAnomalySummariesFromSource(facts, domain.SourceEvents{CompleteSessionIDs: complete})
			if len(got) != 2 {
				t.Fatalf("incomplete source erased parent or ticket cached summary: %+v", got)
			}
			found := false
			for _, a := range got {
				found = found || a.RuleID == domain.RuleSubagentReworkLoop && a.Context.SessionID != nil && *a.Context.SessionID == parent && a.Context.Count != nil && *a.Context.Count == 3
			}
			if !found {
				t.Fatalf("parent rework summary lost: %+v", got)
			}
		})
	}
	cleared := domain.RefreshAnomalySummariesFromSource(facts, domain.SourceEvents{CompleteSessionIDs: []string{parent, "child"}})
	if len(cleared) != 0 {
		t.Fatalf("complete empty group did not replace old summaries: %+v", cleared)
	}
	fresh := []domain.SessionFact{{ID: parent}}
	retained := domain.RetainIncompleteGroupSummaries(facts.Sessions, fresh)
	if len(retained) != 1 || len(retained[0].Anomalies) != 1 || retained[0].Anomalies[0].RuleID != domain.RuleSubagentReworkLoop || len(fresh[0].Anomalies) != 0 {
		t.Fatalf("per-owner refresh erased group summary or mutated fresh input: %+v", retained)
	}
	unobserved := domain.RetainIncompleteGroupSummaries(facts.Sessions, []domain.SessionFact{{ID: parent}, {ID: "child", ParentID: &parent}})
	if len(unobserved[0].Anomalies) != 1 {
		t.Fatal("returned identities alone erased unobserved group history")
	}
	completeFresh := domain.RetainIncompleteGroupSummaries(facts.Sessions, []domain.SessionFact{{ID: parent, Anomalies: []domain.Anomaly{}}, {ID: "child", ParentID: &parent, Anomalies: []domain.Anomaly{}}})
	if len(completeFresh[0].Anomalies) != 0 {
		t.Fatalf("complete group kept obsolete summary: %+v", completeFresh)
	}
}

func TestUnconfiguredRateRefreshPreservesRecordedPolicyEvidence(t *testing.T) {
	tool := "read"
	calls, count := int64(10), int64(6)
	rate := 60.0
	threshold := "saved configured threshold"
	recorded := domain.Anomaly{RuleID: domain.RuleToolHighFailureRate, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{ToolName: &tool, Count: &count, TotalCalls: &calls, FailureRatePct: &rate, Threshold: &threshold}}
	facts := domain.AgenticFactLedger{Anomalies: []domain.Anomaly{recorded}, Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceOpenCode, Tools: []domain.SessionToolFact{{Name: tool, Calls: 10, Errors: 0}}, Anomalies: []domain.Anomaly{}}}}
	source := domain.SourceEvents{Events: []domain.EventFact{}, CompleteSessionIDs: []string{"owner"}}
	preserved := domain.RefreshAnomalySummariesFromSource(facts, source)
	if len(preserved) != 1 || !reflect.DeepEqual(preserved[0], recorded) {
		t.Fatalf("cfg-less analysis erased recorded policy: %+v", preserved)
	}
	for _, invalid := range []domain.AnomalyThresholdsConfig{{ErrorRateMinCalls: 5, ErrorRateCriticalThreshold: .70}, {ErrorRateMinCalls: 5, ErrorRateThreshold: .90, ErrorRateCriticalThreshold: .70}} {
		retained := domain.RefreshAnomalySummariesFromSource(facts, source, invalid)
		if len(retained) != 1 || !reflect.DeepEqual(retained[0], recorded) {
			t.Fatalf("invalid rate policy erased recorded summary: %+v", retained)
		}
	}
	configured := domain.AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .35, ErrorRateCriticalThreshold: .70}
	refreshed := domain.RefreshAnomalySummariesFromSource(facts, source, configured)
	if len(refreshed) != 0 {
		t.Fatalf("complete supported refresh retained stale rate summary: %+v", refreshed)
	}
}
