package domain_test

import (
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

func TestPruneEventsForPersistence(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }

	t.Run("nil or empty events slice returns allocated empty non-nil slice", func(t *testing.T) {
		t.Parallel()
		resNil := domain.PruneEventsForPersistence(domain.AgenticFactLedger{Events: nil})
		if resNil == nil {
			t.Fatal("expected non-nil slice for nil events, got nil")
		}
		if len(resNil) != 0 {
			t.Fatalf("expected 0 events, got %d", len(resNil))
		}

		resEmpty := domain.PruneEventsForPersistence(domain.AgenticFactLedger{Events: []domain.EventFact{}})
		if resEmpty == nil {
			t.Fatal("expected non-nil slice for empty events, got nil")
		}
		if len(resEmpty) != 0 {
			t.Fatalf("expected 0 events, got %d", len(resEmpty))
		}
	})

	t.Run("filters out routine turn messages and successful tool calls", func(t *testing.T) {
		t.Parallel()
		dur := int64(250)
		rawErr := "file not found: config.yaml"

		events := []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:00Z",
				SessionID:     "ses-1",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "success",
				ContextTokens: 1000,
			},
			{
				Timestamp:     "2026-10-18T10:00:01Z",
				SessionID:     "ses-1",
				Turn:          2,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      strPtr("read"),
				Status:        "success",
				DurationMs:    &dur,
				ContextTokens: 1200,
			},
			{
				Timestamp:     "2026-10-18T10:00:02Z",
				SessionID:     "ses-1",
				Turn:          3,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      strPtr("bash"),
				Status:        "error",
				ErrorMessage:  &rawErr,
				DurationMs:    &dur,
				ContextTokens: 1500,
			},
			{
				Timestamp:     "2026-10-18T10:00:03Z",
				SessionID:     "ses-1",
				Turn:          4,
				Role:          "orchestrator",
				EventType:     domain.EventTypeCompaction,
				Status:        "success",
				DurationMs:    &dur,
				ContextTokens: 800,
			},
		}

		ledger := domain.AgenticFactLedger{
			Events: events,
			Tools: []domain.ToolFact{
				{Name: "read", Calls: 1, Errors: 0},
				{Name: "bash", Calls: 1, Errors: 1},
			},
		}

		pruned := domain.PruneEventsForPersistence(ledger)
		if len(pruned) != 2 {
			t.Fatalf("expected 2 pruned events (error tool call and compaction), got %d: %+v", len(pruned), pruned)
		}

		// First pruned event: error tool call
		errEvent := pruned[0]
		if errEvent.Turn != 3 || errEvent.EventType != domain.EventTypeToolCall {
			t.Errorf("expected turn 3 tool call error, got turn %d type %s", errEvent.Turn, errEvent.EventType)
		}
		if errEvent.ToolName == nil || *errEvent.ToolName != "bash" {
			t.Errorf("expected tool_name 'bash', got %v", errEvent.ToolName)
		}
		if errEvent.Status != "error" {
			t.Errorf("expected status 'error', got %s", errEvent.Status)
		}
		if errEvent.ErrorMessage != nil {
			t.Errorf("expected ErrorMessage nil, got %v", *errEvent.ErrorMessage)
		}
		if errEvent.DurationMs == nil || *errEvent.DurationMs != dur {
			t.Errorf("expected DurationMs %d, got %v", dur, errEvent.DurationMs)
		}
		if errEvent.ContextTokens != 1500 {
			t.Errorf("expected ContextTokens 1500, got %d", errEvent.ContextTokens)
		}

		// Second pruned event: compaction
		compEvent := pruned[1]
		if compEvent.Turn != 4 || compEvent.EventType != domain.EventTypeCompaction {
			t.Errorf("expected turn 4 compaction, got turn %d type %s", compEvent.Turn, compEvent.EventType)
		}
		if compEvent.ErrorMessage != nil {
			t.Errorf("expected compaction ErrorMessage nil, got %v", *compEvent.ErrorMessage)
		}
	})

	t.Run("preserves anomaly-associated events without ErrorMessage", func(t *testing.T) {
		t.Parallel()
		leakMsg := "prompt injection / confidential user data leaked in context"

		// Create a context surge: Turn 1 has 10,000 tokens, Turn 2 surges to 60,000 tokens (+50,000)
		events := []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:00Z",
				SessionID:     "ses-anomaly",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "success",
				ContextTokens: 10_000,
			},
			{
				Timestamp:     "2026-10-18T10:00:05Z",
				SessionID:     "ses-anomaly",
				Turn:          2,
				Role:          "coder",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "success",
				ErrorMessage:  &leakMsg,
				ContextTokens: 60_000,
			},
		}

		ledger := domain.AgenticFactLedger{
			Events: events,
		}

		pruned := domain.PruneEventsForPersistence(ledger)
		if len(pruned) != 1 {
			t.Fatalf("expected 1 anomaly-linked event retained, got %d", len(pruned))
		}

		anomalyEvent := pruned[0]
		if anomalyEvent.Turn != 2 {
			t.Errorf("expected turn 2 event retained, got turn %d", anomalyEvent.Turn)
		}
		if anomalyEvent.ContextTokens != 60_000 {
			t.Errorf("expected ContextTokens 60000, got %d", anomalyEvent.ContextTokens)
		}
		if anomalyEvent.ErrorMessage != nil {
			t.Errorf("expected ErrorMessage nil after pruning, got %s", *anomalyEvent.ErrorMessage)
		}
	})

	t.Run("aggregation metric derivation parity between unpruned and pruned ledgers", func(t *testing.T) {
		t.Parallel()
		dur := int64(300)
		errStr := "syntax error on line 42"

		unprunedEvents := []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:00Z",
				SessionID:     "ses-main",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "success",
				ContextTokens: 5000,
			},
			{
				Timestamp:     "2026-10-18T10:00:05Z",
				SessionID:     "ses-main",
				Turn:          2,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      strPtr("read"),
				Status:        "success",
				DurationMs:    &dur,
				ContextTokens: 5500,
			},
			{
				Timestamp:     "2026-10-18T10:00:10Z",
				SessionID:     "ses-main",
				Turn:          3,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      strPtr("bash"),
				Status:        "error",
				ErrorMessage:  &errStr,
				DurationMs:    &dur,
				ContextTokens: 6000,
			},
			{
				Timestamp:     "2026-10-18T10:00:15Z",
				SessionID:     "ses-main",
				Turn:          4,
				Role:          "tester",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "success",
				ContextTokens: 7000,
			},
			{
				Timestamp:     "2026-10-18T10:00:20Z",
				SessionID:     "ses-main",
				Turn:          5,
				Role:          "orchestrator",
				EventType:     domain.EventTypeCompaction,
				Status:        "success",
				DurationMs:    &dur,
				ContextTokens: 3000,
			},
		}

		started := "2026-10-18T10:00:00Z"
		completed := "2026-10-18T10:00:30Z"

		ledgerUnpruned := domain.AgenticFactLedger{
			TicketID:    strPtr("AL-32"),
			TicketTitle: strPtr("Event Pruning"),
			TicketType:  strPtr("feat"),
			StartedAt:   &started,
			CompletedAt: &completed,
			UpdatedAt:   completed,
			Code: domain.CodeFacts{
				LinesAdded:     120,
				LinesDeleted:   15,
				TestLinesAdded: 80,
				ProdLinesAdded: 40,
				FilesChanged:   3,
				SDLCIterations: 1,
			},
			Invocations: []domain.InvocationFact{
				{
					ID:                    "inv-1",
					Role:                  "orchestrator",
					Model:                 "claude-3-opus",
					User:                  "developer",
					StartedAt:             &started,
					CompletedAt:           &completed,
					ActiveDurationSeconds: 15,
					Turns:                 2,
					Tokens: domain.InvocationTokens{
						Input:       1000,
						Output:      500,
						CacheRead:   200,
						Reasoning:   100,
						PeakContext: 5000,
					},
					CostUSD:     0.03,
					Compactions: 1,
				},
				{
					ID:                    "inv-2",
					ParentID:              strPtr("inv-1"),
					Role:                  "coder",
					Model:                 "claude-3-sonnet",
					User:                  "developer",
					StartedAt:             &started,
					CompletedAt:           &completed,
					ActiveDurationSeconds: 15,
					Turns:                 2,
					Tokens: domain.InvocationTokens{
						Input:       2000,
						Output:      800,
						CacheRead:   400,
						Reasoning:   200,
						PeakContext: 6000,
					},
					CostUSD:     0.02,
					Compactions: 0,
				},
			},
			Tools: []domain.ToolFact{
				{Name: "read", Calls: 1, Errors: 0},
				{Name: "bash", Calls: 1, Errors: 1},
			},
			Events: unprunedEvents,
		}

		prunedEvents := domain.PruneEventsForPersistence(ledgerUnpruned)
		ledgerPruned := ledgerUnpruned
		ledgerPruned.Events = prunedEvents

		// 1. ComputeTotals parity
		totalsUnpruned := domain.ComputeTotals(ledgerUnpruned.Invocations)
		totalsPruned := domain.ComputeTotals(ledgerPruned.Invocations)
		if !reflect.DeepEqual(totalsUnpruned, totalsPruned) {
			t.Fatalf("ComputeTotals mismatch:\nunpruned: %+v\npruned: %+v", totalsUnpruned, totalsPruned)
		}

		// 2. ComputeWorkflowEfficiency parity
		effUnpruned := domain.ComputeWorkflowEfficiency(ledgerUnpruned)
		effPruned := domain.ComputeWorkflowEfficiency(ledgerPruned)
		if !reflect.DeepEqual(effUnpruned, effPruned) {
			t.Fatalf("ComputeWorkflowEfficiency mismatch:\nunpruned: %+v\npruned: %+v", effUnpruned, effPruned)
		}

		// 3. ComputeRoleMetrics parity
		rolesUnpruned := domain.ComputeRoleMetrics(ledgerUnpruned.Invocations)
		rolesPruned := domain.ComputeRoleMetrics(ledgerPruned.Invocations)
		if !reflect.DeepEqual(rolesUnpruned, rolesPruned) {
			t.Fatalf("ComputeRoleMetrics mismatch:\nunpruned: %+v\npruned: %+v", rolesUnpruned, rolesPruned)
		}

		// 4. ComputeModelMetrics parity
		modelsUnpruned := domain.ComputeModelMetrics(ledgerUnpruned.Invocations)
		modelsPruned := domain.ComputeModelMetrics(ledgerPruned.Invocations)
		if !reflect.DeepEqual(modelsUnpruned, modelsPruned) {
			t.Fatalf("ComputeModelMetrics mismatch:\nunpruned: %+v\npruned: %+v", modelsUnpruned, modelsPruned)
		}

		// 5. ComputeThreadSummaries parity
		threadsUnpruned := domain.ComputeThreadSummaries(ledgerUnpruned.Invocations)
		threadsPruned := domain.ComputeThreadSummaries(ledgerPruned.Invocations)
		if !reflect.DeepEqual(threadsUnpruned, threadsPruned) {
			t.Fatalf("ComputeThreadSummaries mismatch:\nunpruned: %+v\npruned: %+v", threadsUnpruned, threadsPruned)
		}

		// 6. DeriveMetricsLedger parity
		derivedUnpruned := domain.DeriveMetricsLedger(ledgerUnpruned)
		derivedPruned := domain.DeriveMetricsLedger(ledgerPruned)

		// Derived metrics totals and workflow efficiency must be bit-identical
		if !reflect.DeepEqual(derivedUnpruned.Totals, derivedPruned.Totals) {
			t.Errorf("DeriveMetricsLedger Totals mismatch")
		}
		if !reflect.DeepEqual(derivedUnpruned.WorkflowEfficiency, derivedPruned.WorkflowEfficiency) {
			t.Errorf("DeriveMetricsLedger WorkflowEfficiency mismatch")
		}
		if !reflect.DeepEqual(derivedUnpruned.Anomalies, derivedPruned.Anomalies) {
			t.Errorf("DeriveMetricsLedger Anomalies mismatch:\nunpruned: %+v\npruned: %+v", derivedUnpruned.Anomalies, derivedPruned.Anomalies)
		}
	})
}

func TestPruneEventsForPersistence_TableDrivenEdgeCases(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }
	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name          string
		inputEvents   []domain.EventFact
		tools         []domain.ToolFact
		expectedCount int
		verify        func(t *testing.T, pruned []domain.EventFact)
	}{
		{
			name: "routine turn messages only pruned to empty slice",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:00Z",
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "orchestrator",
					EventType:     domain.EventTypeTurnMessage,
					Status:        "success",
					ContextTokens: 1000,
				},
				{
					Timestamp:     "2026-10-18T10:00:01Z",
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeTurnMessage,
					Status:        "completed",
					ContextTokens: 1200,
				},
			},
			expectedCount: 0,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				if pruned == nil {
					t.Fatal("expected non-nil slice, got nil")
				}
			},
		},
		{
			name: "successful tool calls only pruned to empty slice",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:00Z",
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("read"),
					Status:        "completed",
					DurationMs:    int64Ptr(120),
					ContextTokens: 2000,
				},
				{
					Timestamp:     "2026-10-18T10:00:01Z",
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("glob"),
					Status:        "success",
					DurationMs:    int64Ptr(85),
					ContextTokens: 2500,
				},
			},
			expectedCount: 0,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				if pruned == nil {
					t.Fatal("expected non-nil slice, got nil")
				}
			},
		},
		{
			name: "status failed is retained alongside status error with stripped ErrorMessage",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:00Z",
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("edit"),
					Status:        "failed",
					ErrorMessage:  strPtr("file is locked"),
					DurationMs:    int64Ptr(40),
					ContextTokens: 3000,
				},
				{
					Timestamp:     "2026-10-18T10:00:01Z",
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("bash"),
					Status:        "error",
					ErrorMessage:  strPtr("exit code 127"),
					DurationMs:    int64Ptr(300),
					ContextTokens: 3500,
				},
			},
			expectedCount: 2,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				for i, ev := range pruned {
					if ev.ErrorMessage != nil {
						t.Errorf("event[%d] expected ErrorMessage nil, got %q", i, *ev.ErrorMessage)
					}
					if ev.ToolName == nil {
						t.Errorf("event[%d] expected ToolName preserved, got nil", i)
					}
					if ev.DurationMs == nil {
						t.Errorf("event[%d] expected DurationMs preserved, got nil", i)
					}
					if ev.ContextTokens <= 0 {
						t.Errorf("event[%d] expected positive ContextTokens, got %d", i, ev.ContextTokens)
					}
				}
				if pruned[0].Status != "failed" || *pruned[0].ToolName != "edit" {
					t.Errorf("pruned[0] mismatch: %+v", pruned[0])
				}
				if pruned[1].Status != "error" || *pruned[1].ToolName != "bash" {
					t.Errorf("pruned[1] mismatch: %+v", pruned[1])
				}
			},
		},
		{
			name: "compaction event retained regardless of status with ErrorMessage nil",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:00Z",
					SessionID:     "ses-1",
					Turn:          3,
					Role:          "orchestrator",
					EventType:     domain.EventTypeCompaction,
					Status:        "success",
					ErrorMessage:  strPtr("compaction completed with warning"),
					DurationMs:    int64Ptr(500),
					ContextTokens: 18000,
				},
			},
			expectedCount: 1,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				ev := pruned[0]
				if ev.EventType != domain.EventTypeCompaction {
					t.Errorf("expected compaction event type, got %s", ev.EventType)
				}
				if ev.ErrorMessage != nil {
					t.Errorf("expected compaction ErrorMessage nil, got %q", *ev.ErrorMessage)
				}
				if ev.DurationMs == nil || *ev.DurationMs != 500 {
					t.Errorf("expected DurationMs 500, got %v", ev.DurationMs)
				}
				if ev.ContextTokens != 18000 {
					t.Errorf("expected ContextTokens 18000, got %d", ev.ContextTokens)
				}
			},
		},
		{
			name: "brittle tool failure anomaly retains error tool calls with nil ErrorMessage",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:00Z",
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("bash"),
					Status:        "error",
					ErrorMessage:  strPtr("bash failed 1"),
					DurationMs:    int64Ptr(100),
					ContextTokens: 2000,
				},
				{
					Timestamp:     "2026-10-18T10:00:05Z",
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("bash"),
					Status:        "error",
					ErrorMessage:  strPtr("bash failed 2"),
					DurationMs:    int64Ptr(110),
					ContextTokens: 2200,
				},
			},
			tools: []domain.ToolFact{
				{Name: "bash", Calls: 2, Errors: 2},
			},
			expectedCount: 2,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				for i, ev := range pruned {
					if ev.ErrorMessage != nil {
						t.Errorf("event[%d] expected ErrorMessage nil, got %q", i, *ev.ErrorMessage)
					}
					if *ev.ToolName != "bash" {
						t.Errorf("event[%d] expected ToolName bash, got %s", i, *ev.ToolName)
					}
				}
			},
		},
		{
			name: "mixed chronological events sort ascending by timestamp then turn",
			inputEvents: []domain.EventFact{
				{
					Timestamp:     "2026-10-18T10:00:10Z",
					SessionID:     "ses-1",
					Turn:          5,
					Role:          "orchestrator",
					EventType:     domain.EventTypeCompaction,
					Status:        "success",
					ContextTokens: 5000,
				},
				{
					Timestamp:     "2026-10-18T10:00:02Z",
					SessionID:     "ses-1",
					Turn:          2,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("bash"),
					Status:        "error",
					ErrorMessage:  strPtr("err2"),
					ContextTokens: 4000,
				},
				{
					Timestamp:     "2026-10-18T10:00:02Z",
					SessionID:     "ses-1",
					Turn:          1,
					Role:          "coder",
					EventType:     domain.EventTypeToolCall,
					ToolName:      strPtr("read"),
					Status:        "error",
					ErrorMessage:  strPtr("err1"),
					ContextTokens: 3000,
				},
			},
			expectedCount: 3,
			verify: func(t *testing.T, pruned []domain.EventFact) {
				if pruned[0].Turn != 1 || *pruned[0].ToolName != "read" {
					t.Errorf("expected first event turn 1 read, got %+v", pruned[0])
				}
				if pruned[1].Turn != 2 || *pruned[1].ToolName != "bash" {
					t.Errorf("expected second event turn 2 bash, got %+v", pruned[1])
				}
				if pruned[2].Turn != 5 || pruned[2].EventType != domain.EventTypeCompaction {
					t.Errorf("expected third event turn 5 compaction, got %+v", pruned[2])
				}
				for _, ev := range pruned {
					if ev.ErrorMessage != nil {
						t.Errorf("expected ErrorMessage nil, got %q", *ev.ErrorMessage)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ledger := domain.AgenticFactLedger{
				Events: tt.inputEvents,
				Tools:  tt.tools,
			}

			// Verify immutability: deep copy before pruning
			var originalErrorMessages []string
			for _, ev := range tt.inputEvents {
				if ev.ErrorMessage != nil {
					originalErrorMessages = append(originalErrorMessages, *ev.ErrorMessage)
				} else {
					originalErrorMessages = append(originalErrorMessages, "")
				}
			}

			pruned := domain.PruneEventsForPersistence(ledger)

			if len(pruned) != tt.expectedCount {
				t.Fatalf("expected %d pruned events, got %d", tt.expectedCount, len(pruned))
			}

			// Verify original input events slice was not mutated in place
			for i, ev := range ledger.Events {
				var msg string
				if ev.ErrorMessage != nil {
					msg = *ev.ErrorMessage
				}
				if msg != originalErrorMessages[i] {
					t.Fatalf("ledger.Events[%d] ErrorMessage mutated in-place: was %q, now %q", i, originalErrorMessages[i], msg)
				}
			}

			if tt.verify != nil {
				tt.verify(t, pruned)
			}
		})
	}
}
