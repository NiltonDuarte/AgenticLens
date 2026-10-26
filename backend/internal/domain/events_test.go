package domain_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
		resNil := domain.PruneEventsForPersistence(domain.AgenticFactLedger{Sessions: sessionFacts(nil, nil, nil)})
		if resNil == nil {
			t.Fatal("expected non-nil slice for nil events, got nil")
		}
		if len(resNil) != 0 {
			t.Fatalf("expected 0 events, got %d", len(resNil))
		}

		resEmpty := domain.PruneEventsForPersistence(domain.AgenticFactLedger{Sessions: sessionFacts(nil, nil, []domain.EventFact{})})
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

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil, []domain.ToolFact{
			{SessionID: "session-a", Name: "read", Calls: 1, Errors: 0},
			{SessionID: "session-a", Name: "bash", Calls: 1, Errors: 1},
		}, events),
		}

		pruned := domain.PruneEventsForPersistence(ledger)
		if len(pruned) != 0 {
			t.Fatalf("expected no direct anomaly evidence for isolated error/compaction, got %d: %+v", len(pruned), pruned)
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

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil, nil, events)}

		pruned := domain.PruneEventsForPersistence(ledger)
		if len(pruned) != 2 {
			t.Fatalf("expected before/after context evidence, got %d", len(pruned))
		}

		anomalyEvent := pruned[1]
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
		tt := domain.TicketTypeFeat

		ledgerUnpruned := domain.AgenticFactLedger{
			TicketID:    strPtr("AL-32"),
			TicketTitle: strPtr("Event Pruning"),
			TicketType:  &tt,
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
			}, Sessions: sessionFacts([]domain.InvocationFact{
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
				[]domain.ToolFact{
					{SessionID: "session-a", Name: "read", Calls: 1, Errors: 0},
					{SessionID: "session-a", Name: "bash", Calls: 1, Errors: 1},
				},
				unprunedEvents),
		}

		ledgerPruned, err := domain.PersistSparseAnomalyEvidence(ledgerUnpruned)
		if err != nil {
			t.Fatal(err)
		}

		// 1. ComputeTotals parity
		totalsUnpruned := domain.ComputeTotals(ledgerUnpruned.InvocationFacts())
		totalsPruned := domain.ComputeTotals(ledgerPruned.InvocationFacts())
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
		rolesUnpruned := domain.ComputeRoleMetrics(ledgerUnpruned.InvocationFacts())
		rolesPruned := domain.ComputeRoleMetrics(ledgerPruned.InvocationFacts())
		if !reflect.DeepEqual(rolesUnpruned, rolesPruned) {
			t.Fatalf("ComputeRoleMetrics mismatch:\nunpruned: %+v\npruned: %+v", rolesUnpruned, rolesPruned)
		}

		// 4. ComputeModelMetrics parity
		modelsUnpruned := domain.ComputeModelMetrics(ledgerUnpruned.InvocationFacts())
		modelsPruned := domain.ComputeModelMetrics(ledgerPruned.InvocationFacts())
		if !reflect.DeepEqual(modelsUnpruned, modelsPruned) {
			t.Fatalf("ComputeModelMetrics mismatch:\nunpruned: %+v\npruned: %+v", modelsUnpruned, modelsPruned)
		}

		// 5. ComputeThreadSummaries parity
		threadsUnpruned := domain.ComputeThreadSummaries(ledgerUnpruned.InvocationFacts())
		threadsPruned := domain.ComputeThreadSummaries(ledgerPruned.InvocationFacts())
		if !reflect.DeepEqual(threadsUnpruned, threadsPruned) {
			t.Fatalf("ComputeThreadSummaries mismatch:\nunpruned: %+v\npruned: %+v", threadsUnpruned, threadsPruned)
		}

		// 6. DeriveMetricsLedger parity
		derivedUnpruned := domain.DeriveMetricsLedger(ledgerUnpruned, domain.DefaultProjectConfig())
		derivedPruned := domain.DeriveMetricsLedger(ledgerPruned, domain.DefaultProjectConfig())

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
			name: "isolated failed and error events are omitted",
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
			expectedCount: 0,
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
			name: "unlinked compaction event is omitted",
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
			expectedCount: 0,
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
			name: "unaggregated tool errors have no direct event evidence",
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
				{SessionID: "session-a", Name: "bash", Calls: 2, Errors: 2},
			},
			expectedCount: 0,
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
			name: "isolated event rows are omitted",
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
			expectedCount: 0,
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

			ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil, tt.tools, tt.inputEvents)}
			beforeEvents := ledger.EventFacts()

			pruned := domain.PruneEventsForPersistence(ledger)

			if len(pruned) != tt.expectedCount {
				t.Fatalf("expected %d pruned events, got %d", tt.expectedCount, len(pruned))
			}

			// Verify pruning did not mutate source facts.
			for i, event := range ledger.EventFacts() {
				if event.ErrorMessage != beforeEvents[i].ErrorMessage {
					t.Fatalf("ledger.EventFacts()[%d] changed after pruning", i)
				}
			}

			if tt.verify != nil && tt.expectedCount > 0 {
				tt.verify(t, pruned)
			}
		})
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

func TestPruneEventsForPersistence_DiagnosticSanitization(t *testing.T) {
	t.Parallel()

	errMsg := "permission denied"
	prompt := "Sensitive user prompt with credentials"
	reasoning := "Internal secret reasoning thought"
	toolIn := `{"secret_key": "abc123xyz"}`
	toolOut := "DATABASE_PASSWORD=supersecret"

	ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil, nil, []domain.EventFact{
		{
			Timestamp:        "2026-10-19T10:00:00Z",
			SessionID:        "ses-1",
			Turn:             1,
			Role:             "coder",
			EventType:        domain.EventTypeToolCall,
			Status:           "error",
			ErrorMessage:     &errMsg,
			PromptContent:    &prompt,
			ReasoningContent: &reasoning,
			ToolInput:        &toolIn,
			ToolOutput:       &toolOut,
		},
		{
			Timestamp:        "2026-10-19T10:01:00Z",
			SessionID:        "ses-1",
			Turn:             2,
			Role:             "orchestrator",
			EventType:        domain.EventTypeCompaction,
			Status:           "completed",
			PromptContent:    &prompt,
			ReasoningContent: &reasoning,
		},
	}),
	}

	pruned := domain.PruneEventsForPersistence(ledger)
	if len(pruned) != 0 {
		t.Fatalf("expected no direct evidence for isolated error/compaction, got %d", len(pruned))
	}

	for i, ev := range pruned {
		if ev.ErrorMessage != nil {
			t.Errorf("event %d: ErrorMessage not nil: %q", i, *ev.ErrorMessage)
		}
		if ev.PromptContent != nil {
			t.Errorf("event %d: PromptContent not nil: %q", i, *ev.PromptContent)
		}
		if ev.ReasoningContent != nil {
			t.Errorf("event %d: ReasoningContent not nil: %q", i, *ev.ReasoningContent)
		}
		if ev.ToolInput != nil {
			t.Errorf("event %d: ToolInput not nil: %q", i, *ev.ToolInput)
		}
		if ev.ToolOutput != nil {
			t.Errorf("event %d: ToolOutput not nil: %q", i, *ev.ToolOutput)
		}
	}

	// Verify original ledger is not mutated
	if ledger.EventFacts()[0].PromptContent == nil || *ledger.EventFacts()[0].PromptContent != prompt {
		t.Errorf("original ledger PromptContent was mutated")
	}
	if ledger.EventFacts()[0].ToolInput == nil || *ledger.EventFacts()[0].ToolInput != toolIn {
		t.Errorf("original ledger ToolInput was mutated")
	}
}

func TestPersistSparseAnomalyEvidenceKeepsOnlyDirectEvidenceAndCompactSummaries(t *testing.T) {
	t.Run("retry evidence and session summary are preserved", func(t *testing.T) {
		tool := "bash"
		events := []domain.EventFact{}
		for i := int64(1); i <= 3; i++ {
			events = append(events, domain.EventFact{EventID: fmt.Sprintf("call-%d", i), Timestamp: fmt.Sprintf("2026-10-01T00:00:0%dZ", i), SessionID: "s1", Turn: i, Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error", ContextTokens: 1000})
		}
		sessions, err := domain.SessionsFromFacts(nil, []domain.ToolFact{{SessionID: "s1", Name: tool, Calls: 3, Errors: 3}}, events)
		if err != nil {
			t.Fatal(err)
		}
		full := domain.AgenticFactLedger{Sessions: sessions}
		persisted, err := domain.PersistSparseAnomalyEvidence(full)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(persisted.EventFacts()); got != 3 {
			t.Fatalf("retry evidence rows = %d, want 3", got)
		}
		if len(persisted.Sessions[0].Anomalies) != 1 || persisted.Sessions[0].Anomalies[0].Context.SessionID != nil {
			t.Fatalf("session summary should be nested without session_id: %#v", persisted.Sessions[0].Anomalies)
		}
		metrics := domain.DeriveMetricsLedger(persisted, domain.DefaultProjectConfig())
		for i := range metrics.Anomalies {
			metrics.Anomalies[i].Evidence = nil
		}
		if !reflect.DeepEqual(metrics.Anomalies, domain.DetectAnomalies(full)) {
			t.Fatalf("stored summary differs from full stream: %#v != %#v", metrics.Anomalies, domain.DetectAnomalies(full))
		}
	})

	t.Run("cross-session failure summary is ticket-scoped and references failed events", func(t *testing.T) {
		tools := []domain.ToolFact{{SessionID: "s1", Name: "bash", Calls: 5, Errors: 2}, {SessionID: "s2", Name: "bash", Calls: 5, Errors: 1}}
		errorEvents := []domain.EventFact{
			{EventID: "call-a", Timestamp: "2026-10-01T00:00:01Z", SessionID: "s1", EventType: domain.EventTypeToolCall, ToolName: ptrString("bash"), Status: "error"},
			{EventID: "call-b", Timestamp: "2026-10-01T00:00:02Z", SessionID: "s1", EventType: domain.EventTypeToolCall, ToolName: ptrString("bash"), Status: "failed"},
			{EventID: "call-c", Timestamp: "2026-10-01T00:00:03Z", SessionID: "s2", EventType: domain.EventTypeToolCall, ToolName: ptrString("bash"), Status: "error"},
		}
		invocations := []domain.InvocationFact{
			{ID: "s1", Role: "coder", Model: "model-a", Turns: 2, Tokens: domain.InvocationTokens{Input: 100, Output: 20}},
			{ID: "s2", Role: "tester", Model: "model-b", Turns: 3, Tokens: domain.InvocationTokens{Input: 200, Output: 30}},
		}
		sessions, err := domain.SessionsFromFacts(invocations, tools, errorEvents)
		if err != nil {
			t.Fatal(err)
		}
		full := domain.AgenticFactLedger{Sessions: sessions}
		persisted, err := domain.PersistSparseAnomalyEvidence(full)
		if err != nil {
			t.Fatal(err)
		}
		if len(persisted.Anomalies) != 1 || len(persisted.EventFacts()) != 3 {
			t.Fatalf("expected ticket aggregate with all failed-call evidence: anomalies=%d events=%d", len(persisted.Anomalies), len(persisted.EventFacts()))
		}
		if !reflect.DeepEqual(domain.ComputeTotals(full.InvocationFacts()), domain.ComputeTotals(persisted.InvocationFacts())) {
			t.Fatalf("invocation totals changed during event pruning: full=%+v sparse=%+v", domain.ComputeTotals(full.InvocationFacts()), domain.ComputeTotals(persisted.InvocationFacts()))
		}
		if !reflect.DeepEqual(full.ToolFacts(), persisted.ToolFacts()) {
			t.Fatalf("per-session tool calls/errors changed during event pruning: full=%+v sparse=%+v", full.ToolFacts(), persisted.ToolFacts())
		}
		a := persisted.Anomalies[0]
		if a.Context.SessionID != nil || a.Context.ToolName == nil || *a.Context.ToolName != "bash" || a.Context.Count == nil || *a.Context.Count != 3 || a.Context.TotalCalls == nil || *a.Context.TotalCalls != 10 || a.Context.FailureRatePct == nil || *a.Context.FailureRatePct != 30 || a.Context.Threshold == nil {
			t.Fatalf("ticket anomaly missing aggregate context: %#v", a)
		}
		if len(a.Evidence) != 3 || a.Evidence[0] != (domain.AnomalyEvidenceRef{SessionID: "s1", EventID: "call-a"}) || a.Evidence[2] != (domain.AnomalyEvidenceRef{SessionID: "s2", EventID: "call-c"}) {
			t.Fatalf("ticket anomaly evidence references are incomplete: %#v", a.Evidence)
		}
		metrics := domain.DeriveMetricsLedger(persisted, domain.DefaultProjectConfig())
		if !reflect.DeepEqual(metrics.Anomalies, []domain.Anomaly{a}) {
			t.Fatalf("aggregate summary not used by metrics: %#v", metrics.Anomalies)
		}
	})
}

func TestPersistSparseAnomalyEvidenceSelectsMinimalReworkRoleSequence(t *testing.T) {
	root, coder, tester := "root", "child-coder", "child-tester"
	toolEvents := make([]domain.EventFact, 0, 8)
	for i, role := range []string{"coder", "tester", "coder", "tester", "coder", "tester", "coder"} {
		session := coder
		if role == "tester" {
			session = tester
		}
		toolEvents = append(toolEvents, domain.EventFact{EventID: fmt.Sprintf("role-event-%d", i), Timestamp: fmt.Sprintf("2026-10-01T00:00:0%dZ", i+1), SessionID: session, MessageID: ptrString(fmt.Sprintf("role-%d", i)), Turn: int64(i + 1), Role: domain.AgentRole(role), EventType: domain.EventTypeTurnMessage, Status: "completed", ContextTokens: 1000})
	}
	toolEvents = append(toolEvents, domain.EventFact{Timestamp: "2026-10-01T00:00:08Z", SessionID: root, Turn: 8, Role: "reviewer", EventType: domain.EventTypeTurnMessage, Status: "completed", ContextTokens: 1000})
	parent := root
	sessions, err := domain.SessionsFromFacts([]domain.InvocationFact{{ID: root, Role: "general"}, {ID: coder, ParentID: &parent, Role: "coder"}, {ID: tester, ParentID: &parent, Role: "tester"}}, nil, toolEvents)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := domain.PersistSparseAnomalyEvidence(domain.AgenticFactLedger{Sessions: sessions})
	if err != nil {
		t.Fatal(err)
	}
	retained := persisted.EventFacts()
	if len(retained) != 6 {
		t.Fatalf("retained %d events; want exactly the six role transitions in three cycles: %#v", len(retained), retained)
	}
	for i, event := range retained {
		if event.Turn != int64(i+1) {
			t.Errorf("retained turn %d at index %d, want %d", event.Turn, i, i+1)
		}
	}
	if len(persisted.Anomalies) != 1 || persisted.Anomalies[0].RuleID != domain.RuleSubagentReworkLoop || persisted.Anomalies[0].Context.Count == nil || *persisted.Anomalies[0].Context.Count != 3 {
		t.Fatalf("missing ticket-scoped rework summary: %#v", persisted.Anomalies)
	}
}

func TestPersistSparseAnomalyEvidenceRejectsDirectEvidenceWithoutStableID(t *testing.T) {
	tool := "bash"
	events := []domain.EventFact{
		{Timestamp: "2026-10-01T00:00:01Z", SessionID: "s", Turn: 1, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{Timestamp: "2026-10-01T00:00:02Z", SessionID: "s", Turn: 2, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
		{Timestamp: "2026-10-01T00:00:03Z", SessionID: "s", Turn: 3, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"},
	}
	sessions, err := domain.SessionsFromFacts(nil, []domain.ToolFact{{SessionID: "s", Name: tool, Calls: 3, Errors: 3}}, events)
	if err != nil {
		t.Fatal(err)
	}
	_, err = domain.PersistSparseAnomalyEvidence(domain.AgenticFactLedger{Sessions: sessions})
	if err == nil || !strings.Contains(err.Error(), "stable event_id") {
		t.Fatalf("expected record-level error for selected evidence without stable ID, got %v", err)
	}
}

func TestPersistSparseAnomalyEvidenceDoesNotJoinInterleavedParentGroups(t *testing.T) {
	parentCoder, parentTester := "parent-coder", "parent-tester"
	invocations := []domain.InvocationFact{
		{ID: parentCoder, Role: "general"}, {ID: parentTester, Role: "general"},
		{ID: "coder-1", ParentID: &parentCoder, Role: "coder"}, {ID: "coder-2", ParentID: &parentCoder, Role: "coder"}, {ID: "coder-3", ParentID: &parentCoder, Role: "coder"},
		{ID: "tester-1", ParentID: &parentTester, Role: "tester"}, {ID: "tester-2", ParentID: &parentTester, Role: "tester"}, {ID: "tester-3", ParentID: &parentTester, Role: "tester"},
	}
	events := []domain.EventFact{
		{EventID: "c1", Timestamp: "2026-10-01T00:00:01Z", SessionID: "coder-1", Role: "coder", EventType: domain.EventTypeTurnMessage},
		{EventID: "t1", Timestamp: "2026-10-01T00:00:02Z", SessionID: "tester-1", Role: "tester", EventType: domain.EventTypeTurnMessage},
		{EventID: "c2", Timestamp: "2026-10-01T00:00:03Z", SessionID: "coder-2", Role: "coder", EventType: domain.EventTypeTurnMessage},
		{EventID: "t2", Timestamp: "2026-10-01T00:00:04Z", SessionID: "tester-2", Role: "tester", EventType: domain.EventTypeTurnMessage},
		{EventID: "c3", Timestamp: "2026-10-01T00:00:05Z", SessionID: "coder-3", Role: "coder", EventType: domain.EventTypeTurnMessage},
		{EventID: "t3", Timestamp: "2026-10-01T00:00:06Z", SessionID: "tester-3", Role: "tester", EventType: domain.EventTypeTurnMessage},
	}
	sessions, err := domain.SessionsFromFacts(invocations, nil, events)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := domain.PersistSparseAnomalyEvidence(domain.AgenticFactLedger{Sessions: sessions})
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted.Anomalies) != 0 || len(persisted.EventFacts()) != 0 {
		t.Fatalf("independent parent groups formed false evidence: anomalies=%#v events=%#v", persisted.Anomalies, persisted.EventFacts())
	}
}

func TestPersistSparseAnomalyEvidenceIsIdempotentForStoredReworkSummary(t *testing.T) {
	root, coder, tester := "root", "child-coder", "child-tester"
	parent := root
	firstRole, secondRole, cycles := domain.AgentRole("coder"), domain.AgentRole("tester"), int64(3)
	unrelated := domain.EventFact{Timestamp: "2026-10-01T00:00:00Z", SessionID: coder, Turn: 40, Role: "coder", EventType: domain.EventTypeTurnMessage, Status: "completed", ContextTokens: 1000}
	sessions, err := domain.SessionsFromFacts([]domain.InvocationFact{{ID: root, Role: "general"}, {ID: coder, ParentID: &parent, Role: firstRole}, {ID: tester, ParentID: &parent, Role: secondRole}}, nil, []domain.EventFact{unrelated})
	if err != nil {
		t.Fatal(err)
	}
	stored := domain.Anomaly{RuleID: domain.RuleSubagentReworkLoop, Severity: domain.SeverityWarning, Title: "Subagent rework loop", Description: "three cycles", Context: domain.AnomalyContext{Role: &firstRole, SecondaryRole: &secondRole, Count: &cycles}}
	first, err := domain.PersistSparseAnomalyEvidence(domain.AgenticFactLedger{Sessions: sessions, Anomalies: []domain.Anomaly{stored}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.PersistSparseAnomalyEvidence(first)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("reapplying sparse persistence changed the canonical ledger:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if len(second.Anomalies) != 1 || second.Anomalies[0].RuleID != domain.RuleSubagentReworkLoop {
		t.Fatalf("stored rework summary was dropped: %#v", second.Anomalies)
	}
}

func TestSparseEvidenceMeetsLedgerSizeTargets(t *testing.T) {
	const toolCalls, compactions, errors = 608, 11, 132
	tool := "bash"
	events := make([]domain.EventFact, 0, toolCalls+compactions)
	for i := 0; i < toolCalls; i++ {
		status := "completed"
		if i%4 == 0 && i < errors*4 {
			status = "error"
		}
		events = append(events, domain.EventFact{EventID: fmt.Sprintf("synthetic-call-%d", i), Timestamp: fmt.Sprintf("2026-10-01T00:%02d:%02dZ", i/60, i%60), SessionID: "synthetic-session", MessageID: ptrString(fmt.Sprintf("m-%d", i)), Turn: int64(i + 1), Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: status, DurationMs: ptrInt64(300), ContextTokens: 5000})
	}
	for i := 0; i < compactions; i++ {
		events = append(events, domain.EventFact{Timestamp: fmt.Sprintf("2026-10-02T00:00:%02dZ", i), SessionID: "synthetic-session", Turn: int64(toolCalls + i), Role: "coder", EventType: domain.EventTypeCompaction, Status: "completed", ContextTokens: 5000})
	}
	sessions, err := domain.SessionsFromFacts(nil, []domain.ToolFact{{SessionID: "synthetic-session", Name: tool, Calls: toolCalls, Errors: errors}}, events)
	if err != nil {
		t.Fatal(err)
	}
	full := domain.AgenticFactLedger{Sessions: sessions, Anomalies: nil}
	persisted, err := domain.PersistSparseAnomalyEvidence(full)
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(full)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(persisted)
	if err != nil {
		t.Fatal(err)
	}
	rows := len(persisted.EventFacts())
	t.Logf("synthetic measurement: baseline_event_rows=%d retained_event_rows=%d baseline_bytes=%d sparse_bytes=%d reduction_pct=%.1f", len(events), rows, len(before), len(after), 100*float64(len(before)-len(after))/float64(len(before)))
	if rows*4 > len(events) {
		t.Fatalf("retained %d of %d event rows; max is 25%%", rows, len(events))
	}
	if len(after)*100 > len(before)*60 {
		t.Fatalf("ledger size reduction is below 40%%: before=%d after=%d", len(before), len(after))
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
