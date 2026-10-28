package domain_test

import (
	"math"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestComputeTotals(t *testing.T) {
	t.Parallel()
	t.Run("empty invocations", func(t *testing.T) {
		t.Parallel()
		res := domain.ComputeTotals(nil)
		if res.Invocations != 0 {
			t.Errorf("expected 0 invocations, got %d", res.Invocations)
		}
		if res.CostUSD != 0.0 {
			t.Errorf("expected 0 cost, got %f", res.CostUSD)
		}
		if res.ContextHealth != domain.ContextHealthHealthy {
			t.Errorf("expected HEALTHY, got %s", res.ContextHealth)
		}
	})

	t.Run("multiple invocations with distinct and shared session IDs", func(t *testing.T) {
		invs := []domain.InvocationFact{
			{
				ID:   "ses-1",
				Role: "orchestrator",
				Tokens: domain.InvocationTokens{
					Input:       1000,
					Output:      200,
					CacheRead:   3000,
					PeakContext: 50000,
				},
				CostUSD: 0.12345,
			},
			{
				ID:   "ses-1", // duplicate session for same role
				Role: "orchestrator",
				Tokens: domain.InvocationTokens{
					Input:       1000,
					Output:      200,
					CacheRead:   1000,
					PeakContext: 60000,
				},
				CostUSD: 0.12345,
			},
			{
				ID:   "ses-2",
				Role: "coder",
				Tokens: domain.InvocationTokens{
					Input:       5000,
					Output:      1000,
					CacheRead:   5000,
					PeakContext: 120000,
				},
				CostUSD: 0.5,
			},
		}

		res := domain.ComputeTotals(invs)
		if res.Invocations != 2 { // 1 for orchestrator, 1 for coder
			t.Errorf("expected 2 unique role invocations, got %d", res.Invocations)
		}
		if res.InputTokens != 7000 {
			t.Errorf("expected 7000 input tokens, got %d", res.InputTokens)
		}
		if res.OutputTokens != 1400 {
			t.Errorf("expected 1400 output tokens, got %d", res.OutputTokens)
		}
		if res.CacheReadTokens != 9000 {
			t.Errorf("expected 9000 cache read tokens, got %d", res.CacheReadTokens)
		}
		// total cache (9000) / total cached or input (16000) * 100 = 56.25 -> 56.3%
		if res.CacheHitRatePct != 56.3 {
			t.Errorf("expected 56.3 cache hit rate, got %f", res.CacheHitRatePct)
		}
		if res.PeakContextTokens != 120000 {
			t.Errorf("expected 120000 peak context, got %d", res.PeakContextTokens)
		}
		if res.ContextHealth != domain.ContextHealthElevated {
			t.Errorf("expected ELEVATED health status, got %s", res.ContextHealth)
		}
	})
}

func TestComputeWorkflowEfficiency(t *testing.T) {
	t.Parallel()
	t.Run("zero division defenses", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{
			Code: domain.CodeFacts{}, Sessions: sessionFacts([]domain.InvocationFact{},
				[]domain.ToolFact{}, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ActiveDurationFormatted != "0m 0s" {
			t.Errorf("expected '0m 0s', got %q", eff.ActiveDurationFormatted)
		}
		if eff.TestToProductionRatio.Ratio != "0:0" {
			t.Errorf("expected '0:0', got %q", eff.TestToProductionRatio.Ratio)
		}
		if eff.ToolExecution.SuccessRatePct != 100.0 {
			t.Errorf("expected 100.0%% tool success rate on zero calls, got %f", eff.ToolExecution.SuccessRatePct)
		}
		if eff.ReasoningRatioPct != 0.0 {
			t.Errorf("expected 0.0%% reasoning ratio, got %f", eff.ReasoningRatioPct)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected subagents peak role 'none', got %q", eff.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("durations formatting and subagent velocities", func(t *testing.T) {
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{
			Code: domain.CodeFacts{
				LinesAdded:       150,
				LinesDeleted:     30,
				TestLinesAdded:   80,
				TestLinesDeleted: 10,
				ProdLinesAdded:   40,
				ProdLinesDeleted: 20,
				FilesChanged:     5,
				SDLCIterations:   2,
			}, Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:                    "ses-orch",
					Role:                  "orchestrator",
					ActiveDurationSeconds: 3700, // 1h 1m 40s -> 1h 1m
					Turns:                 10,
					Tokens: domain.InvocationTokens{
						Output:      500,
						Reasoning:   500,
						PeakContext: 90000,
					},
					CostUSD: 1.0,
				},
				{
					ID:                    "ses-coder-1",
					ParentID:              &parentID,
					Role:                  "coder",
					ActiveDurationSeconds: 120,
					Turns:                 4,
					Tokens: domain.InvocationTokens{
						Output:      1000,
						Reasoning:   1000,
						PeakContext: 150000,
					},
					CostUSD: 2.0,
				},
				{
					ID:                    "ses-coder-2",
					ParentID:              &parentID,
					Role:                  "coder",
					ActiveDurationSeconds: 80,
					Turns:                 6,
					Tokens: domain.InvocationTokens{
						Output:      500,
						Reasoning:   500,
						PeakContext: 160000,
					},
					CostUSD: 1.0,
				},
			},
				[]domain.ToolFact{
					{SessionID: "session-a", Name: "read", Calls: 10, Errors: 1},
				}, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ActiveDurationFormatted != "1h 5m" { // 3700 + 120 + 80 = 3900s = 1h 5m
			t.Errorf("expected '1h 5m', got %q", eff.ActiveDurationFormatted)
		}
		if eff.TestToProductionRatio.Ratio != "2.0:1" { // 80 / 40 = 2.0
			t.Errorf("expected '2.0:1', got %q", eff.TestToProductionRatio.Ratio)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 5.0 { // (4 + 6) / 2 = 5.0
			t.Errorf("expected 5.0 avg turns, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 6 {
			t.Errorf("expected 6 peak turns, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected coder as peak subagent, got %q", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 160000 {
			t.Errorf("expected 160000 peak tokens, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
	})
}

func TestComputeRoleAndModelMetrics(t *testing.T) {
	t.Parallel()
	invs := []domain.InvocationFact{
		{
			ID:    "1",
			Role:  "reviewer",
			Model: "model-b",
			Tokens: domain.InvocationTokens{
				Input:       1000,
				Output:      100,
				CacheRead:   2000,
				PeakContext: 50000,
			},
			CostUSD: 0.2,
		},
		{
			ID:    "2",
			Role:  "orchestrator",
			Model: "model-a",
			Tokens: domain.InvocationTokens{
				Input:       2000,
				Output:      200,
				CacheRead:   4000,
				PeakContext: 80000,
			},
			CostUSD: 1.5,
		},
	}

	roles := domain.ComputeRoleMetrics(invs)
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	// Check canonical order: orchestrator must come before reviewer
	if roles[0].Role != "orchestrator" {
		t.Errorf("expected first role to be orchestrator, got %s", roles[0].Role)
	}
	if roles[1].Role != "reviewer" {
		t.Errorf("expected second role to be reviewer, got %s", roles[1].Role)
	}

	models := domain.ComputeModelMetrics(invs)
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	// Sorted by cost descending
	if models[0].Model != "model-a" || models[0].CostUSD != 1.5 {
		t.Errorf("expected model-a with cost 1.5 first, got %+v", models[0])
	}
	if models[1].Model != "model-b" || models[1].CostUSD != 0.2 {
		t.Errorf("expected model-b with cost 0.2 second, got %+v", models[1])
	}
}

func TestComputeThreadSummaries(t *testing.T) {
	t.Parallel()
	invs := []domain.InvocationFact{
		{
			ID:       "child-1",
			ParentID: ptr("root-1"),
			Role:     "coder",
			Model:    "m1",
			User:     "alice",
			Tokens:   domain.InvocationTokens{Input: 100},
			CostUSD:  0.1,
		},
		{
			ID:       "root-1",
			ParentID: nil,
			Role:     "orchestrator",
			Model:    "m1",
			User:     "alice",
			Tokens:   domain.InvocationTokens{Input: 200},
			CostUSD:  0.2,
		},
	}

	records := domain.ComputeThreadSummaries(invs)
	if len(records) != 1 {
		t.Fatalf("expected 1 thread record, got %d", len(records))
	}
	if records[0].SessionID != "root-1" {
		t.Errorf("expected session_id root-1, got %s", records[0].SessionID)
	}
	if records[0].User != "alice" {
		t.Errorf("expected user alice, got %s", records[0].User)
	}
	if len(records[0].Roles) != 2 { // orchestrator and coder
		t.Errorf("expected 2 roles in thread record, got %d", len(records[0].Roles))
	}
}

func TestCalculateFleetSummary(t *testing.T) {
	t.Parallel()
	ledgers := []domain.AgenticMetricsLedger{
		{
			Totals: domain.GroupedMetric{
				Invocations:     2,
				CostUSD:         1.5,
				InputTokens:     5000,
				OutputTokens:    1000,
				CacheReadTokens: 5000,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationSeconds: 120,
				CodeDelivered: domain.CodeDeliveredStats{
					LinesAdded:   10,
					LinesDeleted: 5,
					FilesChanged: 1,
				},
				ReasoningTokens: 200,
			},
			Roles: []domain.RoleMetric{
				{Role: "orchestrator", CostUSD: 1.5},
			},
			Models: []domain.ModelMetric{
				{Model: "gpt-4", CostUSD: 1.5},
			},
		},
	}

	fleet := domain.CalculateFleetSummary(ledgers)
	if fleet.TotalTickets != 1 {
		t.Errorf("expected 1 ticket, got %d", fleet.TotalTickets)
	}
	if fleet.TotalCostUSD != 1.5 {
		t.Errorf("expected 1.5 total cost, got %f", fleet.TotalCostUSD)
	}
	if fleet.OverallCacheHitRatePct != 50.0 {
		t.Errorf("expected 50%% cache hit rate, got %f", fleet.OverallCacheHitRatePct)
	}
	if fleet.RoleCosts["orchestrator"] != 1.5 {
		t.Errorf("expected orchestrator cost 1.5, got %f", fleet.RoleCosts["orchestrator"])
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestFilterLedgersByTimeWindow(t *testing.T) {
	t.Parallel()
	ref := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	dateIn7D := "2026-10-15T10:00:00Z"
	dateIn30D := "2026-10-01T10:00:00Z"
	dateIn90D := "2026-08-15T10:00:00Z"
	dateOld := "2026-05-01T10:00:00Z"

	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID:    ptr("T-7D"),
			CompletedAt: &dateIn7D,
			UpdatedAt:   dateIn7D,
		},
		{
			TicketID:    ptr("T-30D"),
			CompletedAt: &dateIn30D,
			UpdatedAt:   dateIn30D,
		},
		{
			TicketID:    ptr("T-90D"),
			CompletedAt: &dateIn90D,
			UpdatedAt:   dateIn90D,
		},
		{
			TicketID:    ptr("T-OLD"),
			CompletedAt: &dateOld,
			UpdatedAt:   dateOld,
		},
	}

	t.Run("all window returns all ledgers", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(ledgers, domain.TimeWindowAll, ref)
		if len(res) != 4 {
			t.Errorf("expected 4 ledgers, got %d", len(res))
		}
	})

	t.Run("7d window returns within 7 days", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(ledgers, domain.TimeWindow7D, ref)
		if len(res) != 1 {
			t.Fatalf("expected 1 ledger, got %d", len(res))
		}
		if *res[0].TicketID != "T-7D" {
			t.Errorf("expected T-7D, got %s", *res[0].TicketID)
		}
	})

	t.Run("30d window returns within 30 days", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(ledgers, domain.TimeWindow30D, ref)
		if len(res) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(res))
		}
	})

	t.Run("90d window returns within 90 days", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(ledgers, domain.TimeWindow90D, ref)
		if len(res) != 3 {
			t.Fatalf("expected 3 ledgers, got %d", len(res))
		}
	})

	t.Run("1d window returns within 24 hours", func(t *testing.T) {
		dateWithin24h := "2026-10-17T02:00:00Z"
		dateOlderThan24h := "2026-10-16T10:00:00Z"
		oneDayLedgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("T-RECENT"), CompletedAt: &dateWithin24h, UpdatedAt: dateWithin24h},
			{TicketID: ptr("T-OLD-24H"), CompletedAt: &dateOlderThan24h, UpdatedAt: dateOlderThan24h},
		}
		res := domain.FilterLedgersByTimeWindow(oneDayLedgers, domain.TimeWindow1D, ref)
		if len(res) != 1 {
			t.Fatalf("expected 1 ledger, got %d", len(res))
		}
		if *res[0].TicketID != "T-RECENT" {
			t.Errorf("expected T-RECENT, got %s", *res[0].TicketID)
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(nil, domain.TimeWindow7D, ref)
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(res))
		}
	})

	t.Run("strictly excludes tickets without valid completed_at even if updated_at or started_at are within window", func(t *testing.T) {
		t.Parallel()
		dateWithin24h := "2026-10-17T02:00:00Z"
		emptyStr := ""
		wsStr := "   "
		corruptStr := "not-a-valid-date"

		testLedgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("T-NIL"), CompletedAt: nil, UpdatedAt: dateWithin24h, StartedAt: &dateWithin24h},
			{TicketID: ptr("T-EMPTY"), CompletedAt: &emptyStr, UpdatedAt: dateWithin24h, StartedAt: &dateWithin24h},
			{TicketID: ptr("T-WHITESPACE"), CompletedAt: &wsStr, UpdatedAt: dateWithin24h, StartedAt: &dateWithin24h},
			{TicketID: ptr("T-CORRUPT"), CompletedAt: &corruptStr, UpdatedAt: dateWithin24h, StartedAt: &dateWithin24h},
			{TicketID: ptr("T-VALID"), CompletedAt: &dateWithin24h, UpdatedAt: dateWithin24h, StartedAt: &dateWithin24h},
		}

		for _, window := range []domain.TimeWindow{domain.TimeWindow1D, domain.TimeWindow7D, domain.TimeWindow30D, domain.TimeWindow90D} {
			filtered := domain.FilterLedgersByTimeWindow(testLedgers, window, ref)
			if len(filtered) != 1 {
				t.Fatalf("expected exactly 1 valid ledger for window %s, got %d", window, len(filtered))
			}
			if *filtered[0].TicketID != "T-VALID" {
				t.Errorf("expected T-VALID for window %s, got %s", window, *filtered[0].TicketID)
			}
		}
	})
}

func TestSortLedgers(t *testing.T) {
	t.Parallel()
	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID:  ptr("T-1"),
			UpdatedAt: "2026-10-15T10:00:00Z",
			Totals: domain.GroupedMetric{
				CostUSD:     1.50,
				Invocations: 10,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationSeconds: 100,
			},
		},
		{
			TicketID:  ptr("T-2"),
			UpdatedAt: "2026-10-16T10:00:00Z",
			Totals: domain.GroupedMetric{
				CostUSD:     3.00,
				Invocations: 5,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationSeconds: 50,
			},
		},
		{
			TicketID:  ptr("T-3"),
			UpdatedAt: "2026-10-17T10:00:00Z",
			Totals: domain.GroupedMetric{
				CostUSD:     0.50,
				Invocations: 20,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationSeconds: 200,
			},
		},
	}

	t.Run("sort by cost ascending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldCost, true)
		if *res[0].TicketID != "T-3" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-2" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by cost descending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldCost, false)
		if *res[0].TicketID != "T-2" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-3" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by duration ascending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldDuration, true)
		if *res[0].TicketID != "T-2" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-3" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by duration descending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldDuration, false)
		if *res[0].TicketID != "T-3" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-2" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by invocations ascending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldInvocations, true)
		if *res[0].TicketID != "T-2" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-3" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by invocations descending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldInvocations, false)
		if *res[0].TicketID != "T-3" || *res[1].TicketID != "T-1" || *res[2].TicketID != "T-2" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by updated_at ascending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldUpdatedAt, true)
		if *res[0].TicketID != "T-1" || *res[1].TicketID != "T-2" || *res[2].TicketID != "T-3" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("sort by updated_at descending", func(t *testing.T) {
		res := domain.SortLedgers(ledgers, domain.SortFieldUpdatedAt, false)
		if *res[0].TicketID != "T-3" || *res[1].TicketID != "T-2" || *res[2].TicketID != "T-1" {
			t.Errorf("unexpected sort order: %v, %v, %v", *res[0].TicketID, *res[1].TicketID, *res[2].TicketID)
		}
	})

	t.Run("immutability check", func(t *testing.T) {
		originalFirst := *ledgers[0].TicketID
		sorted := domain.SortLedgers(ledgers, domain.SortFieldCost, false)
		if len(sorted) != len(ledgers) {
			t.Fatalf("expected %d sorted ledgers, got %d", len(ledgers), len(sorted))
		}
		if *ledgers[0].TicketID != originalFirst {
			t.Errorf("original slice was mutated")
		}
	})
}

func TestEnumsParsing(t *testing.T) {
	t.Parallel()
	if domain.ParseTimeWindow("7d") != domain.TimeWindow7D {
		t.Errorf("expected TimeWindow7D")
	}
	if domain.ParseTimeWindow("invalid") != domain.TimeWindowAll {
		t.Errorf("expected TimeWindowAll for invalid")
	}

	if domain.ParseLedgerSortField("cost") != domain.SortFieldCost {
		t.Errorf("expected SortFieldCost")
	}
	if domain.ParseLedgerSortField("unknown") != domain.SortFieldUpdatedAt {
		t.Errorf("expected SortFieldUpdatedAt for unknown")
	}
}

func TestBuildComparisonRows(t *testing.T) {
	t.Parallel()
	time1 := "2026-10-17T14:00:00Z"
	time2 := "2026-10-17T15:00:00Z"

	featType := domain.TicketTypeFeat
	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID:    ptr("AL-1"),
			TicketType:  &featType,
			StartedAt:   &time1,
			CompletedAt: &time2,
			UpdatedAt:   time2,
			Totals: domain.GroupedMetric{
				CostUSD: 2.0,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationFormatted: "10m 0s",
				ContextWindowHealth: domain.ContextWindowHealth{
					Status: domain.ContextHealthHealthy,
				},
				TestToProductionRatio: domain.TestToCodeRatio{
					Ratio: "1.0:1",
				},
			},
		},
	}

	rows := domain.BuildComparisonRows(ledgers)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].TicketID != "AL-1" {
		t.Errorf("expected ticket_id AL-1, got %s", rows[0].TicketID)
	}
	if rows[0].TicketType == nil || *rows[0].TicketType != domain.TicketTypeFeat {
		t.Errorf("expected ticket_type feat, got %v", rows[0].TicketType)
	}
	if rows[0].LeadDurationFormatted != "1h 0m" {
		t.Errorf("expected 1h 0m lead time, got %q", rows[0].LeadDurationFormatted)
	}
}

func TestDateFiltersAndAggregations(t *testing.T) {
	t.Parallel()
	ref := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	start, end := domain.ResolveDatePreset(string(domain.TimeWindowLast7Days), ref)
	if start == nil || end == nil {
		t.Fatalf("expected non-nil start and end for 7 days")
	}
	if end.Sub(*start) != 7*24*time.Hour {
		t.Errorf("expected 7 days diff, got %v", end.Sub(*start))
	}
}

func TestFleetSummaryAndComparisonEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("empty fleet summary", func(t *testing.T) {
		fleet := domain.CalculateFleetSummary(nil)
		if fleet.TotalTickets != 0 {
			t.Errorf("expected 0 tickets, got %d", fleet.TotalTickets)
		}
		if fleet.TotalCostUSD != 0.0 {
			t.Errorf("expected 0.0 cost, got %f", fleet.TotalCostUSD)
		}
		if fleet.FormattedTotalDuration != "0m 0s" {
			t.Errorf("expected '0m 0s', got %q", fleet.FormattedTotalDuration)
		}
		if fleet.OverallCacheHitRatePct != 0.0 {
			t.Errorf("expected 0.0 cache hit rate, got %f", fleet.OverallCacheHitRatePct)
		}
		if fleet.OverallReasoningRatioPct != 0.0 {
			t.Errorf("expected 0.0 reasoning ratio, got %f", fleet.OverallReasoningRatioPct)
		}
	})

	t.Run("empty comparison rows", func(t *testing.T) {
		rows := domain.BuildComparisonRows(nil)
		if len(rows) != 0 {
			t.Errorf("expected 0 comparison rows, got %d", len(rows))
		}
	})

	t.Run("empty thread summaries", func(t *testing.T) {
		threads := domain.ComputeThreadSummaries(nil)
		if len(threads) != 0 {
			t.Errorf("expected 0 threads, got %d", len(threads))
		}
	})

	t.Run("role effort rollup empty and zero values", func(t *testing.T) {
		rollup := domain.CalculateRoleEffortRollup(nil)
		if len(rollup) != 0 {
			t.Errorf("expected empty rollup, got %d entries", len(rollup))
		}
	})

	t.Run("time bucket metrics empty ledgers", func(t *testing.T) {
		buckets := domain.AggregateMetricsOverTime(nil, "daily", nil, nil)
		if len(buckets) != 0 {
			t.Errorf("expected empty buckets, got %d", len(buckets))
		}
	})
}

func TestLeadTimeEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("nil started_at or completed_at", func(t *testing.T) {
		if res := domain.ComputeLeadTimeSeconds(nil, nil); res != nil {
			t.Errorf("expected nil lead time, got %v", res)
		}
		timeStr := "2026-10-17T12:00:00Z"
		if res := domain.ComputeLeadTimeSeconds(&timeStr, nil); res != nil {
			t.Errorf("expected nil lead time when completed_at is nil, got %v", res)
		}
		if res := domain.ComputeLeadTimeSeconds(nil, &timeStr); res != nil {
			t.Errorf("expected nil lead time when started_at is nil, got %v", res)
		}
		if fmt := domain.FormatLeadTime(nil); fmt != "In Progress" {
			t.Errorf("expected 'In Progress', got %q", fmt)
		}
	})

	t.Run("negative seconds", func(t *testing.T) {
		neg := int64(-50)
		if fmt := domain.FormatLeadTime(&neg); fmt != "0s" {
			t.Errorf("expected '0s' for negative seconds, got %q", fmt)
		}
	})

	t.Run("days and hours", func(t *testing.T) {
		sec := int64(90000) // 1d 1h
		if fmt := domain.FormatLeadTime(&sec); fmt != "1d 1h" {
			t.Errorf("expected '1d 1h', got %q", fmt)
		}
	})

	t.Run("hours and minutes", func(t *testing.T) {
		sec := int64(7260) // 2h 1m
		if fmt := domain.FormatLeadTime(&sec); fmt != "2h 1m" {
			t.Errorf("expected '2h 1m', got %q", fmt)
		}
	})

	t.Run("minutes and seconds", func(t *testing.T) {
		sec := int64(125) // 2m 5s
		if fmt := domain.FormatLeadTime(&sec); fmt != "2m 5s" {
			t.Errorf("expected '2m 5s', got %q", fmt)
		}
	})
}

func TestAdversarialDomainAggregations(t *testing.T) {
	t.Parallel()
	t.Run("context health thresholds evaluation", func(t *testing.T) {
		cases := []struct {
			name           string
			peakTokens     int64
			expectedHealth domain.ContextHealthStatus
		}{
			{"healthy zero", 0, domain.ContextHealthHealthy},
			{"healthy 50k", 50000, domain.ContextHealthHealthy},
			{"healthy edge 99999", 99999, domain.ContextHealthHealthy},
			{"elevated lower 100000", 100000, domain.ContextHealthElevated},
			{"elevated mid 150000", 150000, domain.ContextHealthElevated},
			{"elevated upper 200000", 200000, domain.ContextHealthElevated},
			{"bloated lower 200001", 200001, domain.ContextHealthBloated},
			{"bloated mid 300000", 300000, domain.ContextHealthBloated},
			{"bloated upper 400000", 400000, domain.ContextHealthBloated},
			{"critical lower 400001", 400001, domain.ContextHealthCritical},
			{"critical high 1000000", 1000000, domain.ContextHealthCritical},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				inv := []domain.InvocationFact{
					{
						ID:   "ses-1",
						Role: "coder",
						Tokens: domain.InvocationTokens{
							PeakContext: tc.peakTokens,
						},
					},
				}
				totals := domain.ComputeTotals(inv)
				if totals.ContextHealth != tc.expectedHealth {
					t.Errorf("expected health %s for %d tokens, got %s", tc.expectedHealth, tc.peakTokens, totals.ContextHealth)
				}
			})
		}
	})

	t.Run("zero tokens, zero cost, zero duration invocation", func(t *testing.T) {
		inv := []domain.InvocationFact{
			{
				ID:                    "ses-zero",
				Role:                  "coder",
				ActiveDurationSeconds: 0,
				CostUSD:               0.0,
				Turns:                 0,
				Tokens: domain.InvocationTokens{
					Input:       0,
					Output:      0,
					CacheRead:   0,
					Reasoning:   0,
					PeakContext: 0,
				},
			},
		}

		totals := domain.ComputeTotals(inv)
		if totals.Invocations != 1 {
			t.Errorf("expected 1 invocation, got %d", totals.Invocations)
		}
		if totals.CostUSD != 0.0 {
			t.Errorf("expected 0 cost, got %f", totals.CostUSD)
		}
		if totals.CacheHitRatePct != 0.0 {
			t.Errorf("expected 0.0 cache hit rate on zero input/cache tokens, got %f", totals.CacheHitRatePct)
		}
		if totals.ContextHealth != domain.ContextHealthHealthy {
			t.Errorf("expected HEALTHY status for zero peak tokens, got %s", totals.ContextHealth)
		}
	})

	t.Run("time window exact boundary inclusion and exclusion", func(t *testing.T) {
		// Reference date is 2026-10-17 12:00:00 UTC.
		// Ref window end is 2026-10-17 23:59:59.999999999 UTC.
		// For 7D window, start is 2026-10-10 23:59:59.999999999 UTC.
		ref := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

		tsStartExact := "2026-10-10T00:00:00Z"
		tsStartBefore := "2026-10-09T23:59:59Z"
		tsEndExact := "2026-10-17T23:59:59Z"
		tsEndFuture := "2026-10-18T00:00:01Z"
		tsInvalid := "invalid-timestamp-string"

		boundaryLedgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("T-START-EXACT"), CompletedAt: &tsStartExact, UpdatedAt: tsStartExact},
			{TicketID: ptr("T-START-BEFORE"), CompletedAt: &tsStartBefore, UpdatedAt: tsStartBefore},
			{TicketID: ptr("T-END-EXACT"), CompletedAt: &tsEndExact, UpdatedAt: tsEndExact},
			{TicketID: ptr("T-END-FUTURE"), CompletedAt: &tsEndFuture, UpdatedAt: tsEndFuture},
			{TicketID: ptr("T-INVALID"), CompletedAt: &tsInvalid, UpdatedAt: tsInvalid},
		}

		filtered := domain.FilterLedgersByTimeWindow(boundaryLedgers, domain.TimeWindow7D, ref)
		matchedIDs := make(map[string]bool)
		for _, l := range filtered {
			if l.TicketID != nil {
				matchedIDs[*l.TicketID] = true
			}
		}

		if !matchedIDs["T-START-EXACT"] {
			t.Errorf("expected T-START-EXACT to be included in 7d window")
		}
		if matchedIDs["T-START-BEFORE"] {
			t.Errorf("expected T-START-BEFORE to be excluded from 7d window")
		}
		if !matchedIDs["T-END-EXACT"] {
			t.Errorf("expected T-END-EXACT to be included in 7d window")
		}
		if matchedIDs["T-END-FUTURE"] {
			t.Errorf("expected T-END-FUTURE to be excluded from 7d window (future date)")
		}
		if matchedIDs["T-INVALID"] {
			t.Errorf("expected T-INVALID to be excluded safely without error")
		}
	})

	t.Run("sorting tie-breaking and equal values", func(t *testing.T) {
		tied := []domain.AgenticMetricsLedger{
			{
				TicketID:  ptr("T-B"),
				UpdatedAt: "2026-10-17T14:00:00Z",
				Totals:    domain.GroupedMetric{CostUSD: 1.0, Invocations: 5},
				WorkflowEfficiency: domain.WorkflowEfficiency{
					ActiveDurationSeconds: 100,
				},
			},
			{
				TicketID:  ptr("T-A"),
				UpdatedAt: "2026-10-17T12:00:00Z",
				Totals:    domain.GroupedMetric{CostUSD: 1.0, Invocations: 5},
				WorkflowEfficiency: domain.WorkflowEfficiency{
					ActiveDurationSeconds: 100,
				},
			},
		}

		// On identical cost, tie-breaker should compare UpdatedAt: T-A is earlier than T-B
		asc := domain.SortLedgers(tied, domain.SortFieldCost, true)
		if *asc[0].TicketID != "T-A" || *asc[1].TicketID != "T-B" {
			t.Errorf("expected tie-breaker to order T-A then T-B, got %s, %s", *asc[0].TicketID, *asc[1].TicketID)
		}

		desc := domain.SortLedgers(tied, domain.SortFieldCost, false)
		if *desc[0].TicketID != "T-B" || *desc[1].TicketID != "T-A" {
			t.Errorf("expected tie-breaker reverse in descending, got %s, %s", *desc[0].TicketID, *desc[1].TicketID)
		}
	})

	t.Run("edge cases pure mathematical parity", func(t *testing.T) {
		t.Run("all zero values", func(t *testing.T) {
			emptyLedger := domain.AgenticFactLedger{
				Code: domain.CodeFacts{}, Sessions: sessionFacts(nil,
					nil, nil),
			}
			eff := domain.ComputeWorkflowEfficiency(emptyLedger)
			tot := domain.ComputeTotals(emptyLedger.InvocationFacts())

			if tot.Invocations != 0 {
				t.Errorf("expected 0 invocations, got %d", tot.Invocations)
			}
			if tot.CostUSD != 0.0 {
				t.Errorf("expected 0 cost, got %f", tot.CostUSD)
			}
			if tot.CacheHitRatePct != 0.0 {
				t.Errorf("expected 0 cache hit rate, got %f", tot.CacheHitRatePct)
			}
			if tot.ContextHealth != domain.ContextHealthHealthy {
				t.Errorf("expected HEALTHY, got %s", tot.ContextHealth)
			}
			if eff.ActiveDurationFormatted != "0m 0s" {
				t.Errorf("expected 0m 0s, got %s", eff.ActiveDurationFormatted)
			}
			if eff.TestToProductionRatio.Ratio != "0:0" {
				t.Errorf("expected 0:0, got %s", eff.TestToProductionRatio.Ratio)
			}
			if eff.ToolExecution.SuccessRatePct != 100.0 {
				t.Errorf("expected 100.0, got %f", eff.ToolExecution.SuccessRatePct)
			}
			if eff.ContextWindowHealth.Status != domain.ContextHealthHealthy {
				t.Errorf("expected HEALTHY, got %s", eff.ContextWindowHealth.Status)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
				t.Errorf("expected none, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("pure test code 1:0 ratio", func(t *testing.T) {
			ledger := domain.AgenticFactLedger{
				Code: domain.CodeFacts{
					TestLinesAdded: 50,
					ProdLinesAdded: 0,
				},
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.TestToProductionRatio.Ratio != "50:0" {
				t.Errorf("expected 50:0, got %s", eff.TestToProductionRatio.Ratio)
			}
		})

		t.Run("pure prod code 0:1 ratio", func(t *testing.T) {
			ledger := domain.AgenticFactLedger{
				Code: domain.CodeFacts{
					TestLinesAdded: 0,
					ProdLinesAdded: 50,
				},
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.TestToProductionRatio.Ratio != "0.0:1" {
				t.Errorf("expected 0.0:1, got %s", eff.TestToProductionRatio.Ratio)
			}
		})

		t.Run("high token saturation exceeding 200k ceiling", func(t *testing.T) {
			parentID := "ses-0"
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:   "ses-0",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						PeakContext: 250000,
					},
				},
				{
					ID:       "ses-1",
					ParentID: &parentID,
					Role:     "coder",
					Tokens: domain.InvocationTokens{
						PeakContext: 210000,
					},
					Turns: 10,
				},
			}, nil, nil),
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.Status != domain.ContextHealthBloated {
				t.Errorf("expected BLOATED, got %s", eff.ContextWindowHealth.Status)
			}
			if eff.ContextWindowHealth.RootSaturationPct != 125.0 {
				t.Errorf("expected 125.0 root saturation, got %f", eff.ContextWindowHealth.RootSaturationPct)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 210000 {
				t.Errorf("expected 210000 subagent peak, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
				t.Errorf("expected coder subagent peak role, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("critical token saturation exceeding 400k ceiling", func(t *testing.T) {
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:   "ses-0",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{
						PeakContext: 450000,
					},
				},
			}, nil, nil),
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.Status != domain.ContextHealthCritical {
				t.Errorf("expected CRITICAL, got %s", eff.ContextWindowHealth.Status)
			}
		})
	})
}

func TestDeriveMetricsLedger_EventsPreservationAndInvariants(t *testing.T) {
	t.Parallel()
	tid := "AL-6"
	title := "Chronological Event Extraction"
	ttype := domain.TicketTypeFeat
	baseLedger := domain.AgenticFactLedger{
		TicketID:    &tid,
		TicketTitle: &title,
		TicketType:  &ttype,
		Code: domain.CodeFacts{
			LinesAdded:   100,
			LinesDeleted: 20,
			FilesChanged: 4,
		}, Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:                    "ses-orch-1",
				Role:                  "orchestrator",
				Model:                 "gemini-3.8-flash",
				User:                  "engineer",
				ActiveDurationSeconds: 120,
				Turns:                 5,
				Tokens: domain.InvocationTokens{
					Input:       5000,
					Output:      1000,
					CacheRead:   15000,
					PeakContext: 20000,
				},
				CostUSD: 0.05,
			},
		},
			[]domain.ToolFact{
				{SessionID: "session-a", Name: "bash", Calls: 5, Errors: 1},
			}, nil),
	}

	t.Run("nil events preserved in Facts and does not mutate mathematical derivations", func(t *testing.T) {
		ledgerNilEvents := baseLedger
		setEventFacts(&ledgerNilEvents, nil)

		metrics := domain.DeriveMetricsLedger(ledgerNilEvents, domain.DefaultProjectConfig())
		if metrics.Facts.EventFacts() != nil {
			t.Errorf("expected nil Facts.Events, got %+v", metrics.Facts.EventFacts())
		}
		if metrics.Totals.Invocations != 1 {
			t.Errorf("expected 1 invocation, got %d", metrics.Totals.Invocations)
		}
		if metrics.Totals.CostUSD != 0.05 {
			t.Errorf("expected cost 0.05, got %f", metrics.Totals.CostUSD)
		}
	})

	t.Run("empty slice events preserved in Facts", func(t *testing.T) {
		ledgerEmptyEvents := baseLedger
		setEventFacts(&ledgerEmptyEvents, []domain.EventFact{})

		metrics := domain.DeriveMetricsLedger(ledgerEmptyEvents, domain.DefaultProjectConfig())
		if metrics.Facts.EventFacts() == nil {
			t.Fatal("expected non-nil empty Events slice")
		}
		if len(metrics.Facts.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got %d", len(metrics.Facts.EventFacts()))
		}
	})

	t.Run("populated events preserved without mutating totals or efficiency", func(t *testing.T) {
		toolBash := "bash"
		errBash := "exit status 1"
		var durMs int64 = 42

		events := []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:00.000000+00:00",
				SessionID:     "ses-orch-1",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeTurnMessage,
				Status:        "completed",
				ContextTokens: 20000,
			},
			{
				Timestamp:     "2026-10-18T10:00:01.000000+00:00",
				SessionID:     "ses-orch-1",
				Turn:          1,
				Role:          "orchestrator",
				EventType:     domain.EventTypeToolCall,
				ToolName:      &toolBash,
				Status:        "error",
				ErrorMessage:  &errBash,
				DurationMs:    &durMs,
				ContextTokens: 20000,
			},
			{
				Timestamp:     "2026-10-18T10:00:02.000000+00:00",
				SessionID:     "ses-orch-1",
				Turn:          2,
				Role:          "orchestrator",
				EventType:     domain.EventTypeCompaction,
				Status:        "completed",
				ContextTokens: 15000,
			},
		}

		ledgerWithEvents := baseLedger
		setEventFacts(&ledgerWithEvents, events)

		metricsWithEvents := domain.DeriveMetricsLedger(ledgerWithEvents, domain.DefaultProjectConfig())
		metricsWithoutEvents := domain.DeriveMetricsLedger(baseLedger, domain.DefaultProjectConfig())

		if len(metricsWithEvents.Facts.EventFacts()) != 3 {
			t.Fatalf("expected 3 events in Facts, got %d", len(metricsWithEvents.Facts.EventFacts()))
		}

		// Invariant: totals, efficiency, roles, models must be identical
		if metricsWithEvents.Totals != metricsWithoutEvents.Totals {
			t.Errorf("totals mutated by presence of events: with=%+v without=%+v",
				metricsWithEvents.Totals, metricsWithoutEvents.Totals)
		}
		if metricsWithEvents.WorkflowEfficiency.ActiveDurationFormatted != metricsWithoutEvents.WorkflowEfficiency.ActiveDurationFormatted {
			t.Errorf("active duration mutated by presence of events")
		}
		if metricsWithEvents.WorkflowEfficiency.ToolExecution != metricsWithoutEvents.WorkflowEfficiency.ToolExecution {
			t.Errorf("tool execution mutated by presence of events")
		}

		// Verify event facts fidelity
		ev0 := metricsWithEvents.Facts.EventFacts()[0]
		if ev0.EventType != domain.EventTypeTurnMessage || ev0.Turn != 1 {
			t.Errorf("unexpected event 0: %+v", ev0)
		}
		ev1 := metricsWithEvents.Facts.EventFacts()[1]
		if ev1.EventType != domain.EventTypeToolCall || ev1.Status != "error" || *ev1.ErrorMessage != errBash || *ev1.DurationMs != durMs {
			t.Errorf("unexpected event 1: %+v", ev1)
		}
		ev2 := metricsWithEvents.Facts.EventFacts()[2]
		if ev2.EventType != domain.EventTypeCompaction || ev2.Turn != 2 {
			t.Errorf("unexpected event 2: %+v", ev2)
		}
	})

	t.Run("zero division on completely empty fact ledger", func(t *testing.T) {
		emptyLedger := domain.AgenticFactLedger{}
		metrics := domain.DeriveMetricsLedger(emptyLedger, domain.DefaultProjectConfig())

		if metrics.Totals.Invocations != 0 {
			t.Errorf("expected 0 invocations, got %d", metrics.Totals.Invocations)
		}
		if metrics.Totals.CacheHitRatePct != 0.0 {
			t.Errorf("expected 0 cache hit rate, got %f", metrics.Totals.CacheHitRatePct)
		}
		if metrics.WorkflowEfficiency.ActiveDurationFormatted != "0m 0s" {
			t.Errorf("expected '0m 0s', got %q", metrics.WorkflowEfficiency.ActiveDurationFormatted)
		}
		if metrics.WorkflowEfficiency.ToolExecution.SuccessRatePct != 100.0 {
			t.Errorf("expected 100.0%% default success rate, got %f", metrics.WorkflowEfficiency.ToolExecution.SuccessRatePct)
		}
		if len(metrics.Facts.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got %d", len(metrics.Facts.EventFacts()))
		}
	})
}

func TestCompareTicketIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		// Identity and empty checks
		{"identical identifiers", "AL-1", "AL-1", 0},
		{"identical with whitespace", "  AL-1 ", "AL-1", 0},
		{"both empty", "", "", 0},
		{"empty a sorts before non-empty b", "", "AL-1", -1},
		{"non-empty a sorts after empty b", "AL-1", "", 1},

		// Standard prefix natural progression
		{"AL-1 vs AL-2", "AL-1", "AL-2", -1},
		{"AL-2 vs AL-1", "AL-2", "AL-1", 1},
		{"AL-2 vs AL-9", "AL-2", "AL-9", -1},
		{"AL-9 vs AL-10", "AL-9", "AL-10", -1},
		{"AL-10 vs AL-11", "AL-10", "AL-11", -1},
		{"AL-11 vs AL-22", "AL-11", "AL-22", -1},
		{"AL-2 vs AL-10", "AL-2", "AL-10", -1},
		{"AL-10 vs AL-2", "AL-10", "AL-2", 1},

		// Prefix without hyphen
		{"AL1 vs AL2", "AL1", "AL2", -1},
		{"AL2 vs AL10", "AL2", "AL10", -1},
		{"AL10 vs AL1", "AL10", "AL1", 1},

		// Prefix with underscore
		{"AL_1 vs AL_2", "AL_1", "AL_2", -1},
		{"AL_2 vs AL_10", "AL_2", "AL_10", -1},

		// Mixed distinct prefixes
		{"AL-10 vs GEN-1", "AL-10", "GEN-1", -1},
		{"GEN-1 vs GEN-2", "GEN-1", "GEN-2", -1},
		{"GEN-2 vs PO-1", "GEN-2", "PO-1", -1},
		{"PO-1 vs AL-10", "PO-1", "AL-10", 1},

		// Case insensitivity
		{"al-2 vs AL-10", "al-2", "AL-10", -1},
		{"AL-2 vs al-10", "AL-2", "al-10", -1},
		{"al-10 vs AL-2", "al-10", "AL-2", 1},

		// Leading zeroes tie-breaking
		{"AL-1 vs AL-01", "AL-1", "AL-01", 1},
		{"AL-01 vs AL-1", "AL-01", "AL-1", -1},

		// Conforming vs non-conforming
		{"conforming AL-1 sorts before non-conforming AL-", "AL-1", "AL-", -1},
		{"non-conforming AL- sorts after conforming AL-1", "AL-", "AL-1", 1},
		{"conforming AL-1 sorts before ticket-xyz", "AL-1", "ticket-xyz", -1},
		{"ticket-xyz sorts after conforming AL-1", "ticket-xyz", "AL-1", 1},
		{"conforming AL-2 sorts before UNKNOWN", "AL-2", "UNKNOWN", -1},
		{"UNKNOWN sorts after conforming AL-2", "UNKNOWN", "AL-2", 1},
		{"conforming AL-1 sorts before purely numeric 123", "AL-1", "123", -1},
		{"purely numeric 123 sorts after conforming AL-1", "123", "AL-1", 1},

		// Missing numbers / prefix-only (neither conforms)
		{"missing numbers lex order", "AL-", "GEN-", -1},
		{"missing numbers without hyphen", "AL", "AL-1", 1},

		// Purely numeric strings (neither conforms, fallback to lexicographical)
		{"purely numeric lex order", "123", "456", -1},
		{"purely numeric lex order reverse", "456", "123", 1},
		{"purely numeric lex order unequal length", "100", "20", -1},

		// Leading zeroes variations
		{"multiple leading zeroes AL-001 vs AL-01", "AL-001", "AL-01", -1},
		{"leading zeroes AL-02 vs AL-10", "AL-02", "AL-10", -1},
		{"leading zeroes AL-009 vs AL-1", "AL-009", "AL-1", 1},
		{"leading zeroes AL-00 vs AL-0", "AL-00", "AL-0", 1},
		{"leading zeroes AL-0 vs AL-00", "AL-0", "AL-00", -1},

		// Mixed case variations
		{"mixed case al-10 vs AL-2", "al-10", "AL-2", 1},
		{"mixed case AL-2 vs al-10", "AL-2", "al-10", -1},
		{"mixed case Al-100 vs aL-99", "Al-100", "aL-99", 1},
		{"identical numbers different case al-1 vs AL-1", "al-1", "AL-1", 1},
		{"identical numbers different case AL-1 vs al-1", "AL-1", "al-1", -1},

		// Delimiter variations
		{"delimiter hyphen vs underscore AL-1 vs AL_1", "AL-1", "AL_1", -1},
		{"delimiter underscore vs hyphen AL_1 vs AL-1", "AL_1", "AL-1", 1},
		{"mixed delimiters AL-10 vs AL_2", "AL-10", "AL_2", 1},
		{"mixed delimiters AL_2 vs AL-10", "AL_2", "AL-10", -1},
		{"non-standard delimiter AL.1 vs AL.2 lex fallback", "AL.1", "AL.2", -1},
		{"non-standard delimiter AL/1 vs AL/2 lex fallback", "AL/1", "AL/2", -1},

		// Whitespace variations
		{"whitespace tab and newline \tAL-1\n vs AL-2", "\tAL-1\n", "AL-2", -1},
		{"whitespace only vs empty", "   ", "", 0},
		{"empty vs whitespace only", "", "   ", 0},
		{"whitespace only vs AL-1", "   ", "AL-1", -1},
		{"AL-1 vs whitespace only", "AL-1", "   ", 1},
		{"whitespace surrounding AL-2 vs AL-10", "  AL-2  ", "AL-10", -1},

		// Suffix variations
		{"suffix comparison AL-1-a vs AL-1-b", "AL-1-a", "AL-1-b", -1},
		{"suffix comparison AL-1-b vs AL-1-a", "AL-1-b", "AL-1-a", 1},
		{"suffix comparison AL-1-hotfix vs AL-2-hotfix", "AL-1-hotfix", "AL-2-hotfix", -1},
		{"suffix comparison AL-2-beta vs AL-10-alpha", "AL-2-beta", "AL-10-alpha", -1},
		{"suffix comparison AL-1 vs AL-1-hotfix", "AL-1", "AL-1-hotfix", -1},
		{"suffix comparison AL-1-hotfix vs AL-1", "AL-1-hotfix", "AL-1", 1},

		// Large numbers and overflow boundaries
		{"max int64 sequence AL-9223372036854775807 vs AL-1", "AL-9223372036854775807", "AL-1", 1},
		{"AL-1 vs max int64 sequence", "AL-1", "AL-9223372036854775807", -1},
		{"overflowing int64 sequence does not panic", "AL-9999999999999999999999999999999", "AL-1", 1},

		// Neither conforms (lexicographical fallback)
		{"non-conforming lex order", "abc", "xyz", -1},
		{"non-conforming lex order reverse", "xyz", "abc", 1},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := domain.CompareTicketIDs(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("CompareTicketIDs(%q, %q) = %d; expected %d", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}

func TestSortLedgers_TicketID(t *testing.T) {
	t.Parallel()

	ticketIDs := []string{"AL-10", "AL-2", "AL-1", "GEN-5", "AL-22", "PO-1"}
	input := make([]domain.AgenticMetricsLedger, len(ticketIDs))
	for i, id := range ticketIDs {
		tid := id
		input[i] = domain.AgenticMetricsLedger{
			TicketID:  &tid,
			UpdatedAt: "2026-10-18T10:00:00Z",
		}
	}

	// Ascending sort
	ascSorted := domain.SortLedgers(input, domain.SortFieldTicketID, true)
	expectedAsc := []string{"AL-1", "AL-2", "AL-10", "AL-22", "GEN-5", "PO-1"}
	if len(ascSorted) != len(expectedAsc) {
		t.Fatalf("expected %d ledgers, got %d", len(expectedAsc), len(ascSorted))
	}
	for i, exp := range expectedAsc {
		if *ascSorted[i].TicketID != exp {
			t.Errorf("asc index %d: expected %s, got %s", i, exp, *ascSorted[i].TicketID)
		}
	}

	// Descending sort
	descSorted := domain.SortLedgers(input, domain.SortFieldTicketID, false)
	expectedDesc := []string{"PO-1", "GEN-5", "AL-22", "AL-10", "AL-2", "AL-1"}
	if len(descSorted) != len(expectedDesc) {
		t.Fatalf("expected %d ledgers, got %d", len(expectedDesc), len(descSorted))
	}
	for i, exp := range expectedDesc {
		if *descSorted[i].TicketID != exp {
			t.Errorf("desc index %d: expected %s, got %s", i, exp, *descSorted[i].TicketID)
		}
	}

	// Immutability check: input slice must not be modified
	originalIDs := []string{"AL-10", "AL-2", "AL-1", "GEN-5", "AL-22", "PO-1"}
	for i, orig := range originalIDs {
		if *input[i].TicketID != orig {
			t.Errorf("input modified at index %d: expected %s, got %s", i, orig, *input[i].TicketID)
		}
	}

	t.Run("nil and empty ticket ID handling", func(t *testing.T) {
		t.Parallel()
		tid1 := "AL-10"
		tid2 := "AL-2"
		emptyStr := ""
		ledgers := []domain.AgenticMetricsLedger{
			{TicketID: &tid1, UpdatedAt: "2026-10-18T12:00:00Z"},
			{TicketID: nil, UpdatedAt: "2026-10-18T10:00:00Z"},
			{TicketID: &emptyStr, UpdatedAt: "2026-10-18T11:00:00Z"},
			{TicketID: &tid2, UpdatedAt: "2026-10-18T13:00:00Z"},
		}

		asc := domain.SortLedgers(ledgers, domain.SortFieldTicketID, true)
		if len(asc) != 4 {
			t.Fatalf("expected 4 ledgers, got %d", len(asc))
		}
		// Empty / nil tickets sort first in ascending
		if asc[0].TicketID != nil && *asc[0].TicketID != "" {
			t.Errorf("expected first to be nil or empty, got %v", asc[0].TicketID)
		}
		if *asc[2].TicketID != "AL-2" || *asc[3].TicketID != "AL-10" {
			t.Errorf("expected AL-2 then AL-10 at end of asc, got %v then %v", *asc[2].TicketID, *asc[3].TicketID)
		}

		desc := domain.SortLedgers(ledgers, domain.SortFieldTicketID, false)
		if len(desc) != 4 {
			t.Fatalf("expected 4 ledgers, got %d", len(desc))
		}
		if *desc[0].TicketID != "AL-10" || *desc[1].TicketID != "AL-2" {
			t.Errorf("expected AL-10 then AL-2 at start of desc, got %v then %v", *desc[0].TicketID, *desc[1].TicketID)
		}
	})

	t.Run("empty and nil slices", func(t *testing.T) {
		t.Parallel()
		resNil := domain.SortLedgers(nil, domain.SortFieldTicketID, true)
		if len(resNil) != 0 {
			t.Errorf("expected 0 ledgers for nil slice, got %d", len(resNil))
		}
		resEmpty := domain.SortLedgers([]domain.AgenticMetricsLedger{}, domain.SortFieldTicketID, false)
		if len(resEmpty) != 0 {
			t.Errorf("expected 0 ledgers for empty slice, got %d", len(resEmpty))
		}
	})

	t.Run("duplicate ticket ID deterministic tie breaking by updated_at", func(t *testing.T) {
		t.Parallel()
		tid := "AL-1"
		dups := []domain.AgenticMetricsLedger{
			{TicketID: &tid, UpdatedAt: "2026-10-18T14:00:00Z"},
			{TicketID: &tid, UpdatedAt: "2026-10-18T10:00:00Z"},
			{TicketID: &tid, UpdatedAt: "2026-10-18T12:00:00Z"},
		}
		asc := domain.SortLedgers(dups, domain.SortFieldTicketID, true)
		if asc[0].UpdatedAt != "2026-10-18T10:00:00Z" ||
			asc[1].UpdatedAt != "2026-10-18T12:00:00Z" ||
			asc[2].UpdatedAt != "2026-10-18T14:00:00Z" {
			t.Errorf("expected tie-break by updated_at asc, got %s, %s, %s",
				asc[0].UpdatedAt, asc[1].UpdatedAt, asc[2].UpdatedAt)
		}
	})
}

func TestParseLedgerSortField_TicketID(t *testing.T) {
	t.Parallel()
	if domain.ParseLedgerSortField("ticket_id") != domain.SortFieldTicketID {
		t.Errorf("expected SortFieldTicketID, got %v", domain.ParseLedgerSortField("ticket_id"))
	}
}

func TestOrderRoles(t *testing.T) {
	t.Parallel()
	input := []string{"advisor", "tester", "researcher", "orchestrator", "coder", "advisor"}

	// With custom order:
	customOrder := []string{"orchestrator", "coder", "tester"}
	gotCustom := domain.OrderRoles(input, customOrder)
	expectedCustom := []string{"orchestrator", "coder", "tester", "advisor", "researcher"}
	if len(gotCustom) != len(expectedCustom) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedCustom), len(gotCustom), gotCustom)
	}
	for i := range expectedCustom {
		if gotCustom[i] != expectedCustom[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedCustom[i], gotCustom[i])
		}
	}

	// Default alphabetical sort without custom order:
	gotAlpha := domain.OrderRoles(input)
	expectedAlpha := []string{"advisor", "coder", "orchestrator", "researcher", "tester"}
	if len(gotAlpha) != len(expectedAlpha) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedAlpha), len(gotAlpha), gotAlpha)
	}
	for i := range expectedAlpha {
		if gotAlpha[i] != expectedAlpha[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedAlpha[i], gotAlpha[i])
		}
	}
}

func TestOrderSubagentRoles(t *testing.T) {
	t.Parallel()
	input := []string{"advisor", "reviewer", "orchestrator", "general", "build", "root", "po", "planner", "advisor", "General", "BUILD"}

	// With custom order:
	customOrder := []string{"planner", "reviewer"}
	gotCustom := domain.OrderSubagentRoles(input, customOrder)
	expectedCustom := []string{"planner", "reviewer", "BUILD", "General", "advisor", "build", "general", "orchestrator", "po", "root"}
	if len(gotCustom) != len(expectedCustom) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedCustom), len(gotCustom), gotCustom)
	}
	for i := range expectedCustom {
		if gotCustom[i] != expectedCustom[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedCustom[i], gotCustom[i])
		}
	}

	// Default alphabetical sort without custom order:
	gotAlpha := domain.OrderSubagentRoles(input)
	expectedAlpha := []string{"BUILD", "General", "advisor", "build", "general", "orchestrator", "planner", "po", "reviewer", "root"}
	if len(gotAlpha) != len(expectedAlpha) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedAlpha), len(gotAlpha), gotAlpha)
	}
	for i := range expectedAlpha {
		if gotAlpha[i] != expectedAlpha[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedAlpha[i], gotAlpha[i])
		}
	}
}

func TestWorkflowEfficiency_DynamicSubagents(t *testing.T) {
	t.Parallel()

	t.Run("dynamic subagent effort distribution", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "ses-orch", Role: "orchestrator", CostUSD: 1.0},
			{ID: "ses-coder", ParentID: &parentID, Role: "coder", CostUSD: 2.0},
			{ID: "ses-adv", ParentID: &parentID, Role: "advisor", CostUSD: 1.0},
			{ID: "ses-po", ParentID: &parentID, Role: "po", CostUSD: 1.0},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		dist := eff.RoleEffortDistributionPct

		if _, hasOrch := dist["orchestrator"]; hasOrch {
			t.Errorf("orchestrator should not be present in role effort distribution")
		}
		if dist["coder"] != 50.0 {
			t.Errorf("expected coder at 50.0%%, got %f", dist["coder"])
		}
		if dist["advisor"] != 25.0 {
			t.Errorf("expected advisor at 25.0%%, got %f", dist["advisor"])
		}
		if dist["po"] != 25.0 {
			t.Errorf("expected po at 25.0%%, got %f", dist["po"])
		}
	})

	t.Run("dynamic subagent turn velocity", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "ses-adv", ParentID: &parentID, Role: "advisor", Turns: 4},
			{ID: "ses-po", ParentID: &parentID, Role: "po", Turns: 8},
			{ID: "ses-coder", ParentID: &parentID, Role: "coder", Turns: 6},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		vel := eff.SubagentTurnVelocity

		if vel.AvgTurnsPerInvocation != 6.0 {
			t.Errorf("expected 6.0 avg turns, got %f", vel.AvgTurnsPerInvocation)
		}
		if vel.PeakTurns != 8 {
			t.Errorf("expected 8 peak turns, got %d", vel.PeakTurns)
		}
		if vel.ByRole["advisor"] != 4.0 {
			t.Errorf("expected advisor 4.0 turns, got %f", vel.ByRole["advisor"])
		}
		if vel.ByRole["po"] != 8.0 {
			t.Errorf("expected po 8.0 turns, got %f", vel.ByRole["po"])
		}
		if vel.ByRole["coder"] != 6.0 {
			t.Errorf("expected coder 6.0 turns, got %f", vel.ByRole["coder"])
		}
	})

	t.Run("subagent peak context tracking with custom role", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:     "ses-orch",
				Role:   "orchestrator",
				Tokens: domain.InvocationTokens{PeakContext: 50000},
			},
			{
				ID:       "ses-coder",
				ParentID: &parentID,
				Role:     "coder",
				Tokens:   domain.InvocationTokens{PeakContext: 120000},
			},
			{
				ID:       "ses-adv",
				ParentID: &parentID,
				Role:     "advisor",
				Tokens:   domain.InvocationTokens{PeakContext: 180000},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		cwh := eff.ContextWindowHealth

		if cwh.SubagentsPeakTokens != 180000 {
			t.Errorf("expected 180000 peak subagent tokens, got %d", cwh.SubagentsPeakTokens)
		}
		if cwh.SubagentsPeakRole != "advisor" {
			t.Errorf("expected advisor as peak subagent role, got %s", cwh.SubagentsPeakRole)
		}
	})

	t.Run("peak context tie-breaking precedence", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-root"
		// Default tie-break: alphabetical order breaks ties (advisor before coder)
		ledgerCanonicalVsCustom := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       "ses-coder",
				ParentID: &parentID,
				Role:     "coder",
				Tokens:   domain.InvocationTokens{PeakContext: 100000},
			},
			{
				ID:       "ses-adv",
				ParentID: &parentID,
				Role:     "advisor",
				Tokens:   domain.InvocationTokens{PeakContext: 100000},
			},
		}, nil, nil),
		}

		eff1 := domain.ComputeWorkflowEfficiency(ledgerCanonicalVsCustom)
		if eff1.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected alphabetical role 'advisor' to win default tie-breaker, got %s", eff1.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order tie-break: custom order slice takes precedence (coder before advisor)
		eff1Custom := domain.ComputeWorkflowEfficiency(ledgerCanonicalVsCustom, []string{"coder", "advisor"})
		if eff1Custom.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected custom role 'coder' to win tie-breaker, got %s", eff1Custom.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom vs Custom tie-break: alphabetical (advisor before po)
		ledgerCustomVsCustom := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       "ses-po",
				ParentID: &parentID,
				Role:     "po",
				Tokens:   domain.InvocationTokens{PeakContext: 100000},
			},
			{
				ID:       "ses-adv",
				ParentID: &parentID,
				Role:     "advisor",
				Tokens:   domain.InvocationTokens{PeakContext: 100000},
			},
		}, nil, nil),
		}

		eff2 := domain.ComputeWorkflowEfficiency(ledgerCustomVsCustom)
		if eff2.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected alphabetical role 'advisor' to win tie-breaker, got %s", eff2.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("orchestrator-only session", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:      "ses-orch",
				Role:    "orchestrator",
				Turns:   10,
				CostUSD: 2.5,
				Tokens:  domain.InvocationTokens{PeakContext: 75000},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected 0 subagent peak tokens, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected 'none' subagents peak role, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution, got %+v", eff.RoleEffortDistributionPct)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})
}

func TestComputeModelMetrics_DeterministicTieBreaking(t *testing.T) {
	t.Parallel()
	invs := []domain.InvocationFact{
		{ID: "1", Model: "model-z", CostUSD: 1.0},
		{ID: "2", Model: "model-a", CostUSD: 1.0},
		{ID: "3", Model: "model-top", CostUSD: 5.0},
	}

	models := domain.ComputeModelMetrics(invs)
	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}
	if models[0].Model != "model-top" || models[0].CostUSD != 5.0 {
		t.Errorf("expected model-top first, got %+v", models[0])
	}
	if models[1].Model != "model-a" || models[1].CostUSD != 1.0 {
		t.Errorf("expected model-a second (alphabetical tie-breaker), got %+v", models[1])
	}
	if models[2].Model != "model-z" || models[2].CostUSD != 1.0 {
		t.Errorf("expected model-z third (alphabetical tie-breaker), got %+v", models[2])
	}
}

func TestComputeRoleMetrics_DynamicOrdering(t *testing.T) {
	t.Parallel()
	invs := []domain.InvocationFact{
		{ID: "1", Role: "advisor", CostUSD: 1.0},
		{ID: "2", Role: "reviewer", CostUSD: 2.0},
		{ID: "3", Role: "planner", CostUSD: 3.0},
		{ID: "4", Role: "architect", CostUSD: 4.0},
		{ID: "5", Role: "orchestrator", CostUSD: 5.0},
	}

	// Default alphabetical ordering when no custom order provided
	metricsAlpha := domain.ComputeRoleMetrics(invs)
	if len(metricsAlpha) != 5 {
		t.Fatalf("expected 5 roles, got %d", len(metricsAlpha))
	}
	expectedAlpha := []domain.AgentRole{"advisor", "architect", "orchestrator", "planner", "reviewer"}
	for i, exp := range expectedAlpha {
		if metricsAlpha[i].Role != exp {
			t.Errorf("default at index %d: expected %s, got %s", i, exp, metricsAlpha[i].Role)
		}
	}

	// Custom order priority followed by alphabetical remainder
	customOrder := []string{"orchestrator", "planner", "reviewer"}
	metricsCustom := domain.ComputeRoleMetrics(invs, customOrder)
	if len(metricsCustom) != 5 {
		t.Fatalf("expected 5 roles, got %d", len(metricsCustom))
	}
	expectedCustom := []domain.AgentRole{"orchestrator", "planner", "reviewer", "advisor", "architect"}
	for i, exp := range expectedCustom {
		if metricsCustom[i].Role != exp {
			t.Errorf("custom at index %d: expected %s, got %s", i, exp, metricsCustom[i].Role)
		}
	}
}

func TestComputeThreadSummaries_DynamicOrderingAndTieBreaking(t *testing.T) {
	t.Parallel()
	rootID := "ses-root"
	invs := []domain.InvocationFact{
		{
			ID:      "ses-root",
			Role:    "orchestrator",
			Model:   "m-top",
			CostUSD: 5.0,
		},
		{
			ID:       "ses-child-1",
			ParentID: &rootID,
			Role:     "advisor",
			Model:    "model-z",
			CostUSD:  1.0,
		},
		{
			ID:       "ses-child-2",
			ParentID: &rootID,
			Role:     "advisor",
			Model:    "model-a",
			CostUSD:  1.0,
		},
		{
			ID:       "ses-child-3",
			ParentID: &rootID,
			Role:     "coder",
			Model:    "m-coder",
			CostUSD:  2.0,
		},
	}

	// With custom order prioritizing orchestrator, coder
	customOrder := []string{"orchestrator", "coder"}
	records := domain.ComputeThreadSummaries(invs, customOrder)
	if len(records) != 1 {
		t.Fatalf("expected 1 thread record, got %d", len(records))
	}

	rec := records[0]
	expectedRoles := []domain.AgentRole{domain.AgentRole("orchestrator"), domain.AgentRole("coder"), domain.AgentRole("advisor")}
	if len(rec.Roles) != len(expectedRoles) {
		t.Fatalf("expected %d roles, got %d", len(expectedRoles), len(rec.Roles))
	}
	for i, exp := range expectedRoles {
		if rec.Roles[i].Role != exp {
			t.Errorf("at index %d: expected role %s, got %s", i, exp, rec.Roles[i].Role)
		}
	}

	advisorRole := rec.Roles[2]
	if len(advisorRole.Models) != 2 {
		t.Fatalf("expected 2 models for advisor, got %d", len(advisorRole.Models))
	}
	if advisorRole.Models[0].Model != "model-a" {
		t.Errorf("expected model-a first, got %s", advisorRole.Models[0].Model)
	}
	if advisorRole.Models[1].Model != "model-z" {
		t.Errorf("expected model-z second, got %s", advisorRole.Models[1].Model)
	}
}

func TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty and whitespace inputs", func(t *testing.T) {
		t.Parallel()
		if len(domain.OrderRoles[string](nil)) != 0 {
			t.Errorf("expected empty slice for nil roles")
		}
		if len(domain.OrderRoles([]string{})) != 0 {
			t.Errorf("expected empty slice for empty roles")
		}
		if len(domain.OrderRoles([]string{"", "   ", "\t", "\n"})) != 0 {
			t.Errorf("expected empty slice for whitespace-only roles")
		}

		if len(domain.OrderSubagentRoles[string](nil)) != 0 {
			t.Errorf("expected empty slice for nil subagent roles")
		}
		if len(domain.OrderSubagentRoles([]string{})) != 0 {
			t.Errorf("expected empty slice for empty subagent roles")
		}
		if len(domain.OrderSubagentRoles([]string{"", "   ", "\t", "\n"})) != 0 {
			t.Errorf("expected empty slice for whitespace-only subagent roles")
		}
	})

	t.Run("only custom roles sorted alphabetically", func(t *testing.T) {
		t.Parallel()
		input := []string{"zeta", "beta", "alpha", "gamma", "beta"}
		expected := []string{"alpha", "beta", "gamma", "zeta"}

		gotRoles := domain.OrderRoles(input)
		if len(gotRoles) != len(expected) {
			t.Fatalf("OrderRoles: expected len %d, got %d", len(expected), len(gotRoles))
		}
		for i := range expected {
			if gotRoles[i] != expected[i] {
				t.Errorf("OrderRoles at %d: expected %s, got %s", i, expected[i], gotRoles[i])
			}
		}

		gotSub := domain.OrderSubagentRoles(input)
		if len(gotSub) != len(expected) {
			t.Fatalf("OrderSubagentRoles: expected len %d, got %d", len(expected), len(gotSub))
		}
		for i := range expected {
			if gotSub[i] != expected[i] {
				t.Errorf("OrderSubagentRoles at %d: expected %s, got %s", i, expected[i], gotSub[i])
			}
		}
	})

	t.Run("canonical roles order preservation", func(t *testing.T) {
		t.Parallel()
		// Reversed canonical roles
		input := []string{"reviewer", "tester", "coder", "planner", "orchestrator"}
		customOrder := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}
		expected := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}

		got := domain.OrderRoles(input, customOrder)
		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("at %d: expected %s, got %s", i, expected[i], got[i])
			}
		}

		// Subagents: custom order first, then remaining roles alphabetically
		subOrder := []string{"planner", "coder", "tester", "reviewer"}
		subExpected := []string{"planner", "coder", "tester", "reviewer", "orchestrator"}
		gotSub := domain.OrderSubagentRoles(input, subOrder)
		if len(gotSub) != len(subExpected) {
			t.Fatalf("expected len %d, got %d: %v", len(subExpected), len(gotSub), gotSub)
		}
		for i := range subExpected {
			if gotSub[i] != subExpected[i] {
				t.Errorf("subagent at %d: expected %s, got %s", i, subExpected[i], gotSub[i])
			}
		}

		// Alphabetical default when no custom order provided
		alphaExpected := []string{"coder", "orchestrator", "planner", "reviewer", "tester"}
		gotAlpha := domain.OrderRoles(input)
		for i := range alphaExpected {
			if gotAlpha[i] != alphaExpected[i] {
				t.Errorf("alpha at %d: expected %s, got %s", i, alphaExpected[i], gotAlpha[i])
			}
		}
	})
}

func TestWorkflowEfficiency_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("completely empty fact ledger zero division defenses", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{}
		eff := domain.ComputeWorkflowEfficiency(ledger)

		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected 0 subagent peak tokens, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution, got %+v", eff.RoleEffortDistributionPct)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 0 {
			t.Errorf("expected 0 peak turns, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
	})

	t.Run("zero cost subagent invocations do not divide by zero or enter distribution", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", Role: "orchestrator", CostUSD: 0.0},
			{ID: "2", ParentID: &parentID, Role: "advisor", CostUSD: 0.0, Turns: 5},
			{ID: "3", ParentID: &parentID, Role: "po", CostUSD: 0.0, Turns: 3},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution when total subagent cost is 0, got %+v", eff.RoleEffortDistributionPct)
		}
		// Turns should still be aggregated even if cost is 0
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 4.0 {
			t.Errorf("expected 4.0 avg turns, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 5 {
			t.Errorf("expected 5 peak turns, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
	})

	t.Run("subagent invocations with zero turns", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", ParentID: &parentID, Role: "advisor", Turns: 0, CostUSD: 1.0},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 0 {
			t.Errorf("expected 0 peak turns, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
		if eff.SubagentTurnVelocity.ByRole["advisor"] != 0.0 {
			t.Errorf("expected 0.0 turns for advisor, got %f", eff.SubagentTurnVelocity.ByRole["advisor"])
		}
	})

	t.Run("multi-way peak context tie-breaking across canonical and custom roles", func(t *testing.T) {
		t.Parallel()

		parentID := "ses-root"
		// Case A: 4 subagents tied at 120k tokens. Default alphabetical: coder > planner > reviewer > tester
		canonicalTiedLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", ParentID: &parentID, Role: "reviewer", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "2", ParentID: &parentID, Role: "tester", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "3", ParentID: &parentID, Role: "coder", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "4", ParentID: &parentID, Role: "planner", Tokens: domain.InvocationTokens{PeakContext: 120000}},
		}, nil, nil),
		}
		effA := domain.ComputeWorkflowEfficiency(canonicalTiedLedger)
		if effA.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected 'coder' to win 4-way alphabetical tie-breaker, got %s", effA.ContextWindowHealth.SubagentsPeakRole)
		}
		// With custom order: planner > coder > tester > reviewer
		customOrderA := []string{"planner", "coder", "tester", "reviewer"}
		effACustom := domain.ComputeWorkflowEfficiency(canonicalTiedLedger, customOrderA)
		if effACustom.ContextWindowHealth.SubagentsPeakRole != "planner" {
			t.Errorf("expected 'planner' to win with custom order, got %s", effACustom.ContextWindowHealth.SubagentsPeakRole)
		}

		// Case B: 4 custom subagents all tied at 120k tokens. Precedence: alphabetical (advisor > po > researcher > security-auditor)
		customTiedLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", ParentID: &parentID, Role: "security-auditor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "2", ParentID: &parentID, Role: "researcher", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "3", ParentID: &parentID, Role: "po", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "4", ParentID: &parentID, Role: "advisor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
		}, nil, nil),
		}
		effB := domain.ComputeWorkflowEfficiency(customTiedLedger)
		if effB.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected 'advisor' to win 4-way custom tie-breaker, got %s", effB.ContextWindowHealth.SubagentsPeakRole)
		}

		// Case C: Tied at 120k between reviewer and advisor.
		// Alphabetical default: advisor < reviewer.
		mixedTiedLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", ParentID: &parentID, Role: "advisor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			{ID: "2", ParentID: &parentID, Role: "reviewer", Tokens: domain.InvocationTokens{PeakContext: 120000}},
		}, nil, nil),
		}
		effC := domain.ComputeWorkflowEfficiency(mixedTiedLedger)
		if effC.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected alphabetical 'advisor' to beat 'reviewer' on default tie-break, got %s", effC.ContextWindowHealth.SubagentsPeakRole)
		}
		// With custom order prioritizing reviewer:
		effCCustom := domain.ComputeWorkflowEfficiency(mixedTiedLedger, []string{"reviewer", "advisor"})
		if effCCustom.ContextWindowHealth.SubagentsPeakRole != "reviewer" {
			t.Errorf("expected custom order 'reviewer' to beat 'advisor' on tie-break, got %s", effCCustom.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("subagent invocations without parent ID are root sessions: excluded from subagent metrics", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "1", Role: "advisor", ParentID: nil, CostUSD: 2.0, Turns: 10, Tokens: domain.InvocationTokens{PeakContext: 50000}},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		// Excluded from subagent effort distribution:
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution for root session, got %+v", eff.RoleEffortDistributionPct)
		}
		// Excluded from subagent turn velocity:
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns because ParentID is nil, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		// Subagents peak tokens and role should be 0 and none:
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected 0 subagents peak tokens, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected 'none' subagents peak role, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		// Root coordination peak tokens matches root session:
		if eff.ContextWindowHealth.RootPeakTokens != 50000 {
			t.Errorf("expected 50000 root peak tokens, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
	})
}

func TestComputeModelMetrics_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty invocations returns empty slice", func(t *testing.T) {
		t.Parallel()
		models := domain.ComputeModelMetrics(nil)
		if len(models) != 0 {
			t.Errorf("expected 0 models for nil invocations, got %d", len(models))
		}
		models = domain.ComputeModelMetrics([]domain.InvocationFact{})
		if len(models) != 0 {
			t.Errorf("expected 0 models for empty invocations, got %d", len(models))
		}
	})

	t.Run("empty model string preserved and sorted", func(t *testing.T) {
		t.Parallel()
		invs := []domain.InvocationFact{
			{ID: "1", Model: "", CostUSD: 1.0},
			{ID: "2", Model: "known-model", CostUSD: 2.0},
		}

		models := domain.ComputeModelMetrics(invs)
		if len(models) != 2 {
			t.Fatalf("expected 2 models, got %d", len(models))
		}
		if models[0].Model != "known-model" {
			t.Errorf("expected known-model first, got %s", models[0].Model)
		}
		if models[1].Model != "" {
			t.Errorf("expected empty string model second, got %q", models[1].Model)
		}
	})

	t.Run("multiple models with zero cost sort strictly alphabetically", func(t *testing.T) {
		t.Parallel()
		invs := []domain.InvocationFact{
			{ID: "1", Model: "model-z", CostUSD: 0.0},
			{ID: "2", Model: "model-m", CostUSD: 0.0},
			{ID: "3", Model: "model-a", CostUSD: 0.0},
			{ID: "4", Model: "model-b", CostUSD: 0.0},
		}

		models := domain.ComputeModelMetrics(invs)
		expected := []string{"model-a", "model-b", "model-m", "model-z"}
		for i, exp := range expected {
			if models[i].Model != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, models[i].Model)
			}
		}
	})
}

func TestComputeRoleMetrics_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty invocations returns empty slice", func(t *testing.T) {
		t.Parallel()
		metrics := domain.ComputeRoleMetrics(nil)
		if len(metrics) != 0 {
			t.Errorf("expected 0 metrics, got %d", len(metrics))
		}
	})

	t.Run("multiple custom and canonical roles with zero costs", func(t *testing.T) {
		t.Parallel()
		invs := []domain.InvocationFact{
			{ID: "1", Role: "security-auditor", CostUSD: 0.0},
			{ID: "2", Role: "reviewer", CostUSD: 0.0},
			{ID: "3", Role: "advisor", CostUSD: 0.0},
			{ID: "4", Role: "coder", CostUSD: 0.0},
		}

		// Default alphabetical ordering:
		metrics := domain.ComputeRoleMetrics(invs)
		expected := []domain.AgentRole{"advisor", "coder", "reviewer", "security-auditor"}
		if len(metrics) != len(expected) {
			t.Fatalf("expected %d metrics, got %d", len(expected), len(metrics))
		}
		for i, exp := range expected {
			if metrics[i].Role != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, metrics[i].Role)
			}
		}

		// With custom order prioritizing coder and reviewer:
		customOrder := []string{"coder", "reviewer"}
		metricsCustom := domain.ComputeRoleMetrics(invs, customOrder)
		expectedCustom := []domain.AgentRole{"coder", "reviewer", "advisor", "security-auditor"}
		if len(metricsCustom) != len(expectedCustom) {
			t.Fatalf("expected %d custom metrics, got %d", len(expectedCustom), len(metricsCustom))
		}
		for i, exp := range expectedCustom {
			if metricsCustom[i].Role != exp {
				t.Errorf("custom at index %d: expected %s, got %s", i, exp, metricsCustom[i].Role)
			}
		}
	})
}

func TestDeriveMetricsLedger_RoleMappings(t *testing.T) {
	t.Parallel()

	t.Run("mapping execution invariance: rolls up build under orchestrator while preserving raw facts", func(t *testing.T) {
		t.Parallel()

		factLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:      "ses-build",
				Role:    "build",
				Model:   "claude-3-sonnet",
				CostUSD: 2.00,
				Turns:   4,
				Tokens: domain.InvocationTokens{
					Input:       1000,
					Output:      500,
					PeakContext: 10000,
				},
			},
			{
				ID:      "ses-orch",
				Role:    "orchestrator",
				Model:   "claude-3-sonnet",
				CostUSD: 1.00,
				Turns:   2,
				Tokens: domain.InvocationTokens{
					Input:       500,
					Output:      200,
					PeakContext: 5000,
				},
			},
		}, nil, nil),
		}

		customMappings := domain.RoleMappings{
			"build":        "orchestrator",
			"orchestrator": "orchestrator",
		}

		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = customMappings
		metrics := domain.DeriveMetricsLedger(factLedger, cfg)

		// 1. Roles contains exactly one entry for orchestrator
		if len(metrics.Roles) != 1 {
			t.Fatalf("expected 1 role, got %d: %+v", len(metrics.Roles), metrics.Roles)
		}
		if metrics.Roles[0].Role != "orchestrator" {
			t.Errorf("expected role orchestrator, got %s", metrics.Roles[0].Role)
		}
		if metrics.Roles[0].Invocations != 2 {
			t.Errorf("expected 2 invocations, got %d", metrics.Roles[0].Invocations)
		}
		if math.Abs(metrics.Roles[0].CostUSD-3.00) > 1e-4 {
			t.Errorf("expected total cost 3.00, got %f", metrics.Roles[0].CostUSD)
		}

		// 2. Facts retains unmutated raw roles
		if len(metrics.Facts.InvocationFacts()) != 2 {
			t.Fatalf("expected 2 raw invocations in Facts, got %d", len(metrics.Facts.InvocationFacts()))
		}
		if metrics.Facts.InvocationFacts()[0].Role != "build" {
			t.Errorf("expected Facts.Invocations[0].Role 'build', got %s", metrics.Facts.InvocationFacts()[0].Role)
		}
		if metrics.Facts.InvocationFacts()[1].Role != "orchestrator" {
			t.Errorf("expected Facts.Invocations[1].Role 'orchestrator', got %s", metrics.Facts.InvocationFacts()[1].Role)
		}

		// 3. Orchestrator is excluded from subagent velocity and effort distribution
		if len(metrics.WorkflowEfficiency.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct for orchestrator-only runs, got: %+v",
				metrics.WorkflowEfficiency.RoleEffortDistributionPct)
		}
		if metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation != 0 {
			t.Errorf("expected 0 avg subagent turns, got %f",
				metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})

	t.Run("maps custom roles into canonical roles and custom subagent roles", func(t *testing.T) {
		t.Parallel()

		pid := "ses-orch"
		factLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:      "ses-orch",
				Role:    "build",
				Model:   "claude-3-sonnet",
				CostUSD: 1.00,
			},
			{
				ID:       "ses-worker-1",
				ParentID: &pid,
				Role:     "coder",
				Model:    "claude-3-opus",
				CostUSD:  2.00,
				Turns:    5,
			},
			{
				ID:       "ses-worker-2",
				ParentID: &pid,
				Role:     "custom-qa",
				Model:    "claude-3-sonnet",
				CostUSD:  2.00,
				Turns:    3,
			},
		}, nil, nil),
		}

		customMappings := domain.RoleMappings{
			"build":     "orchestrator",
			"custom-qa": "tester",
		}

		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = customMappings
		metrics := domain.DeriveMetricsLedger(factLedger, cfg)

		// Derived roles should be coder, orchestrator, tester (alphabetical order when RoleOrder is empty)
		roleNames := make([]string, len(metrics.Roles))
		for i, r := range metrics.Roles {
			roleNames[i] = string(r.Role)
		}
		expectedRoles := []string{"coder", "orchestrator", "tester"}
		if len(roleNames) != len(expectedRoles) {
			t.Fatalf("expected roles %+v, got %+v", expectedRoles, roleNames)
		}
		for i, exp := range expectedRoles {
			if roleNames[i] != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, roleNames[i])
			}
		}

		// Raw facts must still have build, coder, custom-qa
		if metrics.Facts.InvocationFacts()[0].Role != "build" {
			t.Errorf("expected raw fact build, got %s", metrics.Facts.InvocationFacts()[0].Role)
		}
		if metrics.Facts.InvocationFacts()[2].Role != "custom-qa" {
			t.Errorf("expected raw fact custom-qa, got %s", metrics.Facts.InvocationFacts()[2].Role)
		}

		// Subagents effort: coder=50%, tester=50%
		eff := metrics.WorkflowEfficiency.RoleEffortDistributionPct
		if eff["coder"] != 50.0 || eff["tester"] != 50.0 {
			t.Errorf("expected coder=50.0, tester=50.0, got: %+v", eff)
		}

		// Subagent turn velocity: coder=5.0, tester=3.0, avg=4.0
		vel := metrics.WorkflowEfficiency.SubagentTurnVelocity
		if vel.AvgTurnsPerInvocation != 4.0 {
			t.Errorf("expected avg turns 4.0, got %f", vel.AvgTurnsPerInvocation)
		}
		if vel.ByRole["coder"] != 5.0 || vel.ByRole["tester"] != 3.0 {
			t.Errorf("expected coder=5.0, tester=3.0 in velocity, got: %+v", vel.ByRole)
		}
	})
}

func TestDeriveMetricsLedger_RoleOrder(t *testing.T) {
	t.Parallel()

	ticketID := "FIX-2"
	factLedger := domain.AgenticFactLedger{
		TicketID: &ticketID, Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-1", Role: "orchestrator", Model: "gpt-4o", CostUSD: 0.10, ActiveDurationSeconds: 10},
			{ID: "inv-2", Role: "coder", Model: "gpt-4o", CostUSD: 0.20, ActiveDurationSeconds: 20},
			{ID: "inv-3", Role: "tester", Model: "gpt-4o", CostUSD: 0.15, ActiveDurationSeconds: 15},
			{ID: "inv-4", Role: "architect", Model: "gpt-4o", CostUSD: 0.05, ActiveDurationSeconds: 5},
		}, nil, nil),
	}

	t.Run("empty role order defaults strictly to alphabetical sorting", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig()
		metrics := domain.DeriveMetricsLedger(factLedger, cfg)

		roleNames := make([]string, len(metrics.Roles))
		for i, r := range metrics.Roles {
			roleNames[i] = string(r.Role)
		}
		expectedRoles := []string{"architect", "coder", "orchestrator", "tester"}
		if len(roleNames) != len(expectedRoles) {
			t.Fatalf("expected roles %+v, got %+v", expectedRoles, roleNames)
		}
		for i, exp := range expectedRoles {
			if roleNames[i] != exp {
				t.Errorf("at index %d: expected role %q, got %q", i, exp, roleNames[i])
			}
		}
	})

	t.Run("custom role order prioritizes specified sequence then alphabetical remainder", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig()
		cfg.RoleOrder = []string{"tester", "orchestrator"}
		metrics := domain.DeriveMetricsLedger(factLedger, cfg)

		roleNames := make([]string, len(metrics.Roles))
		for i, r := range metrics.Roles {
			roleNames[i] = string(r.Role)
		}
		expectedRoles := []string{"tester", "orchestrator", "architect", "coder"}
		if len(roleNames) != len(expectedRoles) {
			t.Fatalf("expected roles %+v, got %+v", expectedRoles, roleNames)
		}
		for i, exp := range expectedRoles {
			if roleNames[i] != exp {
				t.Errorf("at index %d: expected role %q, got %q", i, exp, roleNames[i])
			}
		}
	})
}

func TestDeriveMetricsLedger_SubagentClassificationUnderRoleMapping(t *testing.T) {
	t.Parallel()

	pid := "ses-orch"
	baseLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
		{
			ID:      "ses-orch",
			Role:    "orchestrator",
			Model:   "claude-3-sonnet",
			CostUSD: 1.00,
		},
		{
			ID:       "ses-custom",
			ParentID: &pid,
			Role:     "custom-bot",
			Model:    "claude-3-sonnet",
			CostUSD:  1.50,
			Turns:    4,
		},
	}, nil, nil),
	}

	t.Run("Scenario A: child invocation mapped to orchestrator is included in subagent metrics structurally", func(t *testing.T) {
		t.Parallel()
		mappings := domain.RoleMappings{
			"custom-bot": "orchestrator",
		}
		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = mappings
		metrics := domain.DeriveMetricsLedger(baseLedger, cfg)

		eff := metrics.WorkflowEfficiency.RoleEffortDistributionPct
		if eff["orchestrator"] != 100.0 {
			t.Errorf("expected orchestrator=100.0%% effort under structural classification, got: %+v", eff)
		}
		if metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation != 4.0 {
			t.Errorf("expected 4.0 avg turns, got %f",
				metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})

	t.Run("Scenario B: custom-bot mapped to coder is included in subagent metrics with 100% effort", func(t *testing.T) {
		t.Parallel()
		mappings := domain.RoleMappings{
			"custom-bot": "coder",
		}
		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = mappings
		metrics := domain.DeriveMetricsLedger(baseLedger, cfg)

		eff := metrics.WorkflowEfficiency.RoleEffortDistributionPct
		if eff["coder"] != 100.0 {
			t.Errorf("expected coder=100.0%% effort, got: %+v", eff)
		}
		if _, ok := eff["custom-bot"]; ok {
			t.Errorf("custom-bot should not appear under unmapped name in RoleEffortDistributionPct")
		}
		vel := metrics.WorkflowEfficiency.SubagentTurnVelocity
		if vel.AvgTurnsPerInvocation != 4.0 || vel.ByRole["coder"] != 4.0 {
			t.Errorf("expected coder velocity 4.0, got: %+v", vel)
		}
	})

	t.Run("Scenario C: custom-bot unmapped is treated as dynamic subagent and included in subagent metrics", func(t *testing.T) {
		t.Parallel()
		mappings := domain.RoleMappings{
			"orchestrator": "orchestrator",
			"build":        "orchestrator",
		}
		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = mappings
		metrics := domain.DeriveMetricsLedger(baseLedger, cfg)

		eff := metrics.WorkflowEfficiency.RoleEffortDistributionPct
		if eff["custom-bot"] != 100.0 {
			t.Errorf("expected custom-bot=100.0%% effort, got: %+v", eff)
		}
		vel := metrics.WorkflowEfficiency.SubagentTurnVelocity
		if vel.AvgTurnsPerInvocation != 4.0 || vel.ByRole["custom-bot"] != 4.0 {
			t.Errorf("expected custom-bot velocity 4.0, got: %+v", vel)
		}
	})
}

func TestDeriveMetricsLedger_AdversarialDeepImmutabilityAndEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("strict immutability of input FactLedger and raw persona preservation in Facts", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch-raw"
		origInvs := []domain.InvocationFact{
			{
				ID:      "ses-orch-raw",
				Role:    "build",
				Model:   "claude-3-sonnet",
				User:    "alice",
				CostUSD: 1.50,
				Turns:   2,
				Tokens: domain.InvocationTokens{
					Input: 2000, Output: 400, PeakContext: 12000,
				},
			},
			{
				ID:       "ses-sub-qa",
				ParentID: &parentID,
				Role:     "custom-qa",
				Model:    "claude-3-opus",
				User:     "bob",
				CostUSD:  2.50,
				Turns:    6,
				Tokens: domain.InvocationTokens{
					Input: 5000, Output: 1000, PeakContext: 30000,
				},
			},
			{
				ID:       "ses-sub-dyn",
				ParentID: &parentID,
				Role:     "security-auditor",
				Model:    "gemini-2.5-pro",
				User:     "charlie",
				CostUSD:  1.00,
				Turns:    4,
				Tokens: domain.InvocationTokens{
					Input: 3000, Output: 500, PeakContext: 25000,
				},
			},
		}

		origEvents := []domain.EventFact{
			{
				Timestamp: "2026-10-18T10:00:00Z",
				SessionID: "ses-orch-raw",
				Turn:      1,
				Role:      "build",
				EventType: domain.EventTypeTurnMessage,
				Status:    "completed",
			},
			{
				Timestamp: "2026-10-18T10:00:05Z",
				SessionID: "ses-sub-qa",
				Turn:      1,
				Role:      "custom-qa",
				EventType: domain.EventTypeTurnMessage,
				Status:    "completed",
			},
			{
				Timestamp: "2026-10-18T10:00:10Z",
				SessionID: "ses-sub-dyn",
				Turn:      1,
				Role:      "security-auditor",
				EventType: domain.EventTypeToolCall,
				Status:    "completed",
			},
		}

		tid := "AL-IMMUTABILITY"
		factLedger := domain.AgenticFactLedger{
			TicketID: &tid, Sessions: sessionFacts(origInvs, nil, origEvents),
		}

		// Snapshot before derivation
		invBefore0Role := factLedger.InvocationFacts()[0].Role
		invBefore1Role := factLedger.InvocationFacts()[1].Role
		invBefore2Role := factLedger.InvocationFacts()[2].Role
		evBefore0Role := factLedger.EventFacts()[0].Role
		evBefore1Role := factLedger.EventFacts()[1].Role
		evBefore2Role := factLedger.EventFacts()[2].Role

		mappings := domain.RoleMappings{
			"build":     "orchestrator",
			"custom-qa": "tester",
		}

		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = mappings
		derived := domain.DeriveMetricsLedger(factLedger, cfg)

		// 1. Assert input factLedger was not mutated in place
		if factLedger.InvocationFacts()[0].Role != invBefore0Role || invBefore0Role != "build" {
			t.Errorf("input factLedger.InvocationFacts()[0].Role mutated: got %q, want 'build'", factLedger.InvocationFacts()[0].Role)
		}
		if factLedger.InvocationFacts()[1].Role != invBefore1Role || invBefore1Role != "custom-qa" {
			t.Errorf("input factLedger.InvocationFacts()[1].Role mutated: got %q, want 'custom-qa'", factLedger.InvocationFacts()[1].Role)
		}
		if factLedger.InvocationFacts()[2].Role != invBefore2Role || invBefore2Role != "security-auditor" {
			t.Errorf("input factLedger.InvocationFacts()[2].Role mutated: got %q, want 'security-auditor'", factLedger.InvocationFacts()[2].Role)
		}
		if factLedger.EventFacts()[0].Role != evBefore0Role || evBefore0Role != "build" {
			t.Errorf("input factLedger.EventFacts()[0].Role mutated: got %q, want 'build'", factLedger.EventFacts()[0].Role)
		}
		if factLedger.EventFacts()[1].Role != evBefore1Role || evBefore1Role != "custom-qa" {
			t.Errorf("input factLedger.EventFacts()[1].Role mutated: got %q, want 'custom-qa'", factLedger.EventFacts()[1].Role)
		}
		if factLedger.EventFacts()[2].Role != evBefore2Role || evBefore2Role != "security-auditor" {
			t.Errorf("input factLedger.EventFacts()[2].Role mutated: got %q, want 'security-auditor'", factLedger.EventFacts()[2].Role)
		}

		// 2. Assert derived.Facts retains the exact raw persona strings
		if derived.Facts.InvocationFacts()[0].Role != "build" {
			t.Errorf("derived.Facts.InvocationFacts()[0].Role is not raw 'build', got %q", derived.Facts.InvocationFacts()[0].Role)
		}
		if derived.Facts.InvocationFacts()[1].Role != "custom-qa" {
			t.Errorf("derived.Facts.InvocationFacts()[1].Role is not raw 'custom-qa', got %q", derived.Facts.InvocationFacts()[1].Role)
		}
		if derived.Facts.InvocationFacts()[2].Role != "security-auditor" {
			t.Errorf("derived.Facts.InvocationFacts()[2].Role is not raw 'security-auditor', got %q", derived.Facts.InvocationFacts()[2].Role)
		}
		if derived.Facts.EventFacts()[0].Role != "build" {
			t.Errorf("derived.Facts.EventFacts()[0].Role is not raw 'build', got %q", derived.Facts.EventFacts()[0].Role)
		}
		if derived.Facts.EventFacts()[1].Role != "custom-qa" {
			t.Errorf("derived.Facts.EventFacts()[1].Role is not raw 'custom-qa', got %q", derived.Facts.EventFacts()[1].Role)
		}
		if derived.Facts.EventFacts()[2].Role != "security-auditor" {
			t.Errorf("derived.Facts.EventFacts()[2].Role is not raw 'security-auditor', got %q", derived.Facts.EventFacts()[2].Role)
		}

		// 3. Assert derived aggregates use mapped roles (orchestrator, tester) and dynamic subagent (security-auditor)
		roleMap := make(map[domain.AgentRole]domain.RoleMetric)
		for _, r := range derived.Roles {
			roleMap[r.Role] = r
		}
		if _, ok := roleMap["build"]; ok {
			t.Errorf("unmapped raw role 'build' leaked into derived.Roles: %+v", derived.Roles)
		}
		if _, ok := roleMap["custom-qa"]; ok {
			t.Errorf("unmapped raw role 'custom-qa' leaked into derived.Roles: %+v", derived.Roles)
		}
		if orch, ok := roleMap["orchestrator"]; !ok || orch.Invocations != 1 || orch.CostUSD != 1.50 {
			t.Errorf("expected mapped orchestrator with 1 inv and $1.50 cost, got: %+v", orch)
		}
		if tester, ok := roleMap["tester"]; !ok || tester.Invocations != 1 || tester.CostUSD != 2.50 {
			t.Errorf("expected mapped tester with 1 inv and $2.50 cost, got: %+v", tester)
		}
		if sec, ok := roleMap["security-auditor"]; !ok || sec.Invocations != 1 || sec.CostUSD != 1.00 {
			t.Errorf("expected dynamic subagent security-auditor with 1 inv and $1.00 cost, got: %+v", sec)
		}

		// 4. Assert subagent effort distribution excludes orchestrator, includes tester and security-auditor
		effDist := derived.WorkflowEfficiency.RoleEffortDistributionPct
		if _, ok := effDist["orchestrator"]; ok {
			t.Errorf("orchestrator must be excluded from subagent effort distribution: %+v", effDist)
		}
		if _, ok := effDist["build"]; ok {
			t.Errorf("raw build must not appear in effort distribution: %+v", effDist)
		}
		// total subagent cost = 2.50 + 1.00 = 3.50
		// tester = 2.50/3.50 = 71.4%
		// security-auditor = 1.00/3.50 = 28.6%
		if math.Abs(effDist["tester"]-71.4) > 0.1 {
			t.Errorf("expected tester effort ~71.4%%, got %f", effDist["tester"])
		}
		if math.Abs(effDist["security-auditor"]-28.6) > 0.1 {
			t.Errorf("expected security-auditor effort ~28.6%%, got %f", effDist["security-auditor"])
		}
	})

	t.Run("zero tokens, zero cost, zero duration, empty ledgers resilience", func(t *testing.T) {
		t.Parallel()

		// Case A: completely empty fact ledger
		emptyLedger := domain.AgenticFactLedger{}
		derivedEmpty := domain.DeriveMetricsLedger(emptyLedger, domain.DefaultProjectConfig())

		if derivedEmpty.Totals.Invocations != 0 {
			t.Errorf("expected 0 invocations, got %d", derivedEmpty.Totals.Invocations)
		}
		if derivedEmpty.Totals.CostUSD != 0.0 {
			t.Errorf("expected 0.0 cost, got %f", derivedEmpty.Totals.CostUSD)
		}
		if derivedEmpty.Totals.ContextHealth != domain.ContextHealthHealthy {
			t.Errorf("expected HEALTHY context health for empty ledger, got %s", derivedEmpty.Totals.ContextHealth)
		}
		if derivedEmpty.WorkflowEfficiency.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected 'none' subagents peak role, got %s", derivedEmpty.WorkflowEfficiency.ContextWindowHealth.SubagentsPeakRole)
		}
		if derivedEmpty.WorkflowEfficiency.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected 0 subagents peak tokens, got %d", derivedEmpty.WorkflowEfficiency.ContextWindowHealth.SubagentsPeakTokens)
		}
		if len(derivedEmpty.Roles) != 0 || len(derivedEmpty.Models) != 0 || len(derivedEmpty.Threads) != 0 {
			t.Errorf("expected empty slices for empty ledger, got roles=%d models=%d threads=%d",
				len(derivedEmpty.Roles), len(derivedEmpty.Models), len(derivedEmpty.Threads))
		}

		// Case B: Invocations with all zeroes across metrics
		zeroInvLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:                    "ses-zero-1",
				Role:                  "build",
				Model:                 "zero-model",
				ActiveDurationSeconds: 0,
				CostUSD:               0.0,
				Turns:                 0,
				Tokens: domain.InvocationTokens{
					Input:       0,
					Output:      0,
					CacheRead:   0,
					Reasoning:   0,
					PeakContext: 0,
				},
			},
			{
				ID:                    "ses-zero-2",
				Role:                  "coder",
				ActiveDurationSeconds: 0,
				CostUSD:               0.0,
				Turns:                 0,
				Tokens: domain.InvocationTokens{
					Input:       0,
					Output:      0,
					CacheRead:   0,
					Reasoning:   0,
					PeakContext: 0,
				},
			},
		}, nil, nil),
		}

		mappings := domain.RoleMappings{"build": "orchestrator"}
		cfg := domain.DefaultProjectConfig()
		cfg.RoleMappings = mappings
		derivedZero := domain.DeriveMetricsLedger(zeroInvLedger, cfg)

		if derivedZero.Totals.Invocations != 2 {
			t.Errorf("expected 2 invocations, got %d", derivedZero.Totals.Invocations)
		}
		if derivedZero.Totals.CostUSD != 0.0 {
			t.Errorf("expected 0.0 cost, got %f", derivedZero.Totals.CostUSD)
		}
		if derivedZero.Totals.CacheHitRatePct != 0.0 {
			t.Errorf("expected 0.0 cache hit rate, got %f", derivedZero.Totals.CacheHitRatePct)
		}
		if derivedZero.Totals.ContextHealth != domain.ContextHealthHealthy {
			t.Errorf("expected HEALTHY context health, got %s", derivedZero.Totals.ContextHealth)
		}
		if len(derivedZero.WorkflowEfficiency.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution on zero cost, got %+v", derivedZero.WorkflowEfficiency.RoleEffortDistributionPct)
		}
		if derivedZero.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns, got %f", derivedZero.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})

	t.Run("role mappings configuration handling: default, nil, custom override, and empty mappings", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-root"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "ses-root", Role: "build", CostUSD: 1.0},
			{ID: "ses-child", ParentID: &parentID, Role: "coder", CostUSD: 2.0, Turns: 3},
		}, nil, nil),
		}

		// Default configuration: with empty RoleMappings, raw roles are preserved
		cfgDefault := domain.DefaultProjectConfig()
		derivedDefault := domain.DeriveMetricsLedger(ledger, cfgDefault)
		if len(derivedDefault.Roles) != 2 || (derivedDefault.Roles[0].Role != "build" && derivedDefault.Roles[1].Role != "build") {
			t.Errorf("expected default config to preserve raw roles, got: %+v", derivedDefault.Roles)
		}

		// Nil mapping: also preserves raw roles
		cfgNil := domain.DefaultProjectConfig()
		cfgNil.RoleMappings = nil
		derivedNil := domain.DeriveMetricsLedger(ledger, cfgNil)
		if len(derivedNil.Roles) != 2 || (derivedNil.Roles[0].Role != "build" && derivedNil.Roles[1].Role != "build") {
			t.Errorf("expected nil mapping to preserve raw roles, got: %+v", derivedNil.Roles)
		}

		// Custom mapping overriding default build: maps build to custom-builder
		cfgOverride := domain.DefaultProjectConfig()
		cfgOverride.RoleMappings = domain.RoleMappings{"build": "custom-builder"}
		derivedOverride := domain.DeriveMetricsLedger(ledger, cfgOverride)
		if len(derivedOverride.Roles) != 2 || derivedOverride.Roles[1].Role != "custom-builder" {
			t.Errorf("expected custom mapping to override build to 'custom-builder', got: %+v", derivedOverride.Roles)
		}

		// Empty mapping: unmapped custom role retains raw name
		ledgerWithCustom := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "ses-custom", Role: "specialist-agent", CostUSD: 3.0},
		}, nil, nil),
		}
		cfgEmptyMap := domain.DefaultProjectConfig()
		cfgEmptyMap.RoleMappings = domain.RoleMappings{}
		derivedEmptyMap := domain.DeriveMetricsLedger(ledgerWithCustom, cfgEmptyMap)
		if len(derivedEmptyMap.Roles) != 1 || derivedEmptyMap.Roles[0].Role != "specialist-agent" {
			t.Errorf("expected empty mapping to preserve unmapped custom role 'specialist-agent', got: %+v", derivedEmptyMap.Roles)
		}
	})

	t.Run("subagent classification matrix under role mappings", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		cases := []struct {
			name               string
			rawRole            string
			parentID           *string
			targetRole         domain.AgentRole
			expectSubagentCost bool
			expectTurnCounted  bool
		}{
			{
				name:               "root session with raw build mapped to orchestrator",
				rawRole:            "build",
				parentID:           nil,
				targetRole:         domain.AgentRole("orchestrator"),
				expectSubagentCost: false,
				expectTurnCounted:  false,
			},
			{
				name:               "child session with raw build mapped to orchestrator",
				rawRole:            "build",
				parentID:           &parentID,
				targetRole:         domain.AgentRole("orchestrator"),
				expectSubagentCost: true,
				expectTurnCounted:  true,
			},
			{
				name:               "child session with custom-worker mapped to coder",
				rawRole:            "custom-worker",
				parentID:           &parentID,
				targetRole:         domain.AgentRole("coder"),
				expectSubagentCost: true,
				expectTurnCounted:  true,
			},
			{
				name:               "child session with unmapped custom role treated as dynamic subagent",
				rawRole:            "security-auditor",
				parentID:           &parentID,
				targetRole:         "security-auditor",
				expectSubagentCost: true,
				expectTurnCounted:  true,
			},
			{
				name:               "root session with unmapped custom subagent role: cost and turns excluded",
				rawRole:            "independent-agent",
				parentID:           nil,
				targetRole:         "independent-agent",
				expectSubagentCost: false,
				expectTurnCounted:  false,
			},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				factLedger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:       "ses-test",
						ParentID: tc.parentID,
						Role:     domain.AgentRole(tc.rawRole),
						CostUSD:  1.50,
						Turns:    5,
						Tokens: domain.InvocationTokens{
							Input: 1000, Output: 200, PeakContext: 10000,
						},
					},
				}, nil, nil),
				}

				mappings := domain.RoleMappings{
					tc.rawRole: string(tc.targetRole),
				}
				cfg := domain.DefaultProjectConfig()
				cfg.RoleMappings = mappings
				metrics := domain.DeriveMetricsLedger(factLedger, cfg)

				targetStr := string(tc.targetRole)
				_, hasSubCost := metrics.WorkflowEfficiency.RoleEffortDistributionPct[targetStr]
				if hasSubCost != tc.expectSubagentCost {
					t.Errorf("RoleEffortDistributionPct[%q] presence = %v, want %v", targetStr, hasSubCost, tc.expectSubagentCost)
				}

				turns, hasTurns := metrics.WorkflowEfficiency.SubagentTurnVelocity.ByRole[tc.targetRole]
				if tc.expectTurnCounted {
					if !hasTurns || turns != 5.0 {
						t.Errorf("expected 5.0 turns for %s, got: %+v", targetStr, metrics.WorkflowEfficiency.SubagentTurnVelocity)
					}
				} else {
					if hasTurns {
						t.Errorf("expected no turns recorded for %s, got %f", targetStr, turns)
					}
				}
			})
		}
	})
}

func TestParseLedgerSortField_CompletedAt(t *testing.T) {
	t.Parallel()
	if domain.ParseLedgerSortField("completed_at") != domain.SortFieldCompletedAt {
		t.Errorf("expected SortFieldCompletedAt, got %v", domain.ParseLedgerSortField("completed_at"))
	}
}

func TestGetLedgerCompletionDate(t *testing.T) {
	t.Parallel()

	t.Run("uses completed_at when present", func(t *testing.T) {
		t.Parallel()
		completed := "2026-03-10T15:30:00Z"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: &completed,
			UpdatedAt:   "2026-03-09T10:00:00Z",
			StartedAt:   &completed,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if !ok {
			t.Fatal("expected ok = true")
		}
		expected, _ := time.Parse(time.RFC3339, completed)
		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("returns false when completed_at is nil even if updated_at is present", func(t *testing.T) {
		t.Parallel()
		updated := "2026-03-09T10:00:00Z"
		started := "2026-03-08T08:00:00Z"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: nil,
			UpdatedAt:   updated,
			StartedAt:   &started,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if ok {
			t.Errorf("expected ok = false, got true with %v", got)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("returns false when completed_at is nil even if started_at is present", func(t *testing.T) {
		t.Parallel()
		started := "2026-03-08T08:00:00Z"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: nil,
			UpdatedAt:   "",
			StartedAt:   &started,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if ok {
			t.Errorf("expected ok = false, got true with %v", got)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("returns false when completed_at is whitespace only", func(t *testing.T) {
		t.Parallel()
		ws := "   "
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: &ws,
			UpdatedAt:   "2026-03-09T10:00:00Z",
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if ok {
			t.Errorf("expected ok = false for whitespace completed_at, got %v", got)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("returns false when all timestamps are absent or unparseable", func(t *testing.T) {
		t.Parallel()
		badStr := "invalid-date"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: &badStr,
			UpdatedAt:   "also-invalid",
			StartedAt:   nil,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if ok {
			t.Errorf("expected ok = false, got true with %v", got)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("adversarial table of nil, empty, whitespace, and corrupt completed_at with valid fallbacks", func(t *testing.T) {
		t.Parallel()
		validUpdated := "2026-03-09T10:00:00Z"
		validStarted := "2026-03-08T08:00:00Z"

		cases := []struct {
			name        string
			completedAt *string
		}{
			{name: "nil completed_at", completedAt: nil},
			{name: "empty string completed_at", completedAt: ptr("")},
			{name: "whitespace completed_at", completedAt: ptr("   ")},
			{name: "invalid format", completedAt: ptr("garbage-date")},
			{name: "out of range month", completedAt: ptr("2026-13-01T00:00:00Z")},
			{name: "out of range day", completedAt: ptr("2026-02-31T00:00:00Z")},
			{name: "partial date without time or proper ISO", completedAt: ptr("not-a-date")},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ledger := domain.AgenticMetricsLedger{
					CompletedAt: tc.completedAt,
					UpdatedAt:   validUpdated,
					StartedAt:   &validStarted,
				}
				got, ok := domain.GetLedgerCompletionDate(ledger)
				if ok {
					t.Errorf("[%s] expected ok = false, got true with %v", tc.name, got)
				}
				if !got.IsZero() {
					t.Errorf("[%s] expected zero time, got %v", tc.name, got)
				}
			})
		}
	})
}

func TestGetTicketEffectiveDate(t *testing.T) {
	t.Parallel()

	t.Run("parses valid completed_at", func(t *testing.T) {
		t.Parallel()
		completed := "2026-03-10T15:30:00Z"
		started := "2026-03-08T08:00:00Z"
		updated := "2026-03-09T10:00:00Z"

		got, ok := domain.GetTicketEffectiveDate(&completed, &started, updated)
		if !ok {
			t.Fatalf("expected ok = true")
		}
		expected, _ := time.Parse("2006-01-02", "2026-03-10")
		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("returns false when completed_at is nil even if started_at and updated_at are valid", func(t *testing.T) {
		t.Parallel()
		started := "2026-03-08T08:00:00Z"
		updated := "2026-03-09T10:00:00Z"

		got, ok := domain.GetTicketEffectiveDate(nil, &started, updated)
		if ok {
			t.Errorf("expected ok = false, got true with %v", got)
		}
		if !got.IsZero() {
			t.Errorf("expected zero time, got %v", got)
		}
	})

	t.Run("returns false when completed_at is empty or unparseable", func(t *testing.T) {
		t.Parallel()
		emptyStr := ""
		badStr := "invalid"
		started := "2026-03-08T08:00:00Z"
		updated := "2026-03-09T10:00:00Z"

		got1, ok1 := domain.GetTicketEffectiveDate(&emptyStr, &started, updated)
		if ok1 || !got1.IsZero() {
			t.Errorf("expected ok=false and zero time for empty completed_at, got %v, %v", ok1, got1)
		}

		got2, ok2 := domain.GetTicketEffectiveDate(&badStr, &started, updated)
		if ok2 || !got2.IsZero() {
			t.Errorf("expected ok=false and zero time for bad completed_at, got %v, %v", ok2, got2)
		}
	})

	t.Run("adversarial table of nil, empty, whitespace, and corrupt completed_at with valid fallbacks", func(t *testing.T) {
		t.Parallel()
		validStarted := "2026-03-08T08:00:00Z"
		validUpdated := "2026-03-09T10:00:00Z"

		cases := []struct {
			name        string
			completedAt *string
		}{
			{name: "nil completed_at", completedAt: nil},
			{name: "empty completed_at", completedAt: ptr("")},
			{name: "whitespace completed_at", completedAt: ptr("   ")},
			{name: "invalid timestamp format", completedAt: ptr("not-a-timestamp")},
			{name: "invalid calendar date", completedAt: ptr("2026-99-99")},
			{name: "corrupted iso", completedAt: ptr("2026-02-31T12:00:00Z")},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				got, ok := domain.GetTicketEffectiveDate(tc.completedAt, &validStarted, validUpdated)
				if ok {
					t.Errorf("[%s] expected ok = false, got true with %v", tc.name, got)
				}
				if !got.IsZero() {
					t.Errorf("[%s] expected zero time, got %v", tc.name, got)
				}
			})
		}
	})
}

func TestFilterLedgersByDateRange(t *testing.T) {
	t.Parallel()

	d1 := "2026-03-01T10:00:00Z"
	d2 := "2026-03-15T10:00:00Z"
	d3 := "2026-03-30T10:00:00Z"

	ledgers := []domain.AgenticMetricsLedger{
		{TicketID: ptr("T-1"), CompletedAt: &d1},
		{TicketID: ptr("T-2"), CompletedAt: &d2},
		{TicketID: ptr("T-3"), CompletedAt: &d3},
		{TicketID: ptr("T-NIL"), CompletedAt: nil, StartedAt: &d1, UpdatedAt: d1},
	}

	start, _ := time.Parse("2006-01-02", "2026-03-10")
	end, _ := time.Parse("2006-01-02", "2026-03-20")

	filtered := domain.FilterLedgersByDateRange(ledgers, &start, &end)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 ledger, got %d", len(filtered))
	}
	if *filtered[0].TicketID != "T-2" {
		t.Errorf("expected T-2, got %s", *filtered[0].TicketID)
	}

	t.Run("strictly excludes ledgers lacking valid completed_at even if started_at and updated_at match range", func(t *testing.T) {
		t.Parallel()
		midRange := "2026-03-15T12:00:00Z"
		emptyStr := ""
		wsStr := "   "
		corruptStr := "corrupted-date-value"

		testLedgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("T-NIL"), CompletedAt: nil, StartedAt: &midRange, UpdatedAt: midRange},
			{TicketID: ptr("T-EMPTY"), CompletedAt: &emptyStr, StartedAt: &midRange, UpdatedAt: midRange},
			{TicketID: ptr("T-WHITESPACE"), CompletedAt: &wsStr, StartedAt: &midRange, UpdatedAt: midRange},
			{TicketID: ptr("T-CORRUPT"), CompletedAt: &corruptStr, StartedAt: &midRange, UpdatedAt: midRange},
			{TicketID: ptr("T-VALID"), CompletedAt: &midRange, StartedAt: &midRange, UpdatedAt: midRange},
		}

		startDate, _ := time.Parse("2006-01-02", "2026-03-10")
		endDate, _ := time.Parse("2006-01-02", "2026-03-20")

		res := domain.FilterLedgersByDateRange(testLedgers, &startDate, &endDate)
		if len(res) != 1 {
			t.Fatalf("expected exactly 1 valid ledger, got %d", len(res))
		}
		if *res[0].TicketID != "T-VALID" {
			t.Errorf("expected T-VALID, got %s", *res[0].TicketID)
		}
	})
}

func TestFilterLedgersByCompletionDate(
	t *testing.T) {
	t.Parallel()

	date1 := "2026-03-01T10:00:00Z"
	date2 := "2026-03-10T12:00:00Z"
	date3 := "2026-03-20T18:00:00Z"

	ledgers := []domain.AgenticMetricsLedger{
		{TicketID: ptr("AL-1"), CompletedAt: &date1},
		{TicketID: ptr("AL-2"), CompletedAt: &date2},
		{TicketID: ptr("AL-3"), CompletedAt: &date3},
		{TicketID: ptr("AL-4"), CompletedAt: nil, UpdatedAt: ""},
	}

	t.Run("nil boundaries returns all ledgers", func(t *testing.T) {
		t.Parallel()
		filtered := domain.FilterLedgersByCompletionDate(ledgers, nil, nil)
		if len(filtered) != 4 {
			t.Fatalf("expected 4 ledgers, got %d", len(filtered))
		}
	})

	t.Run("since boundary inclusive filters older and dateless ledgers", func(t *testing.T) {
		t.Parallel()
		since, _ := time.Parse(time.RFC3339, "2026-03-10T12:00:00Z")
		filtered := domain.FilterLedgersByCompletionDate(ledgers, &since, nil)
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "AL-2" || *filtered[1].TicketID != "AL-3" {
			t.Errorf("unexpected ledgers: %v, %v", *filtered[0].TicketID, *filtered[1].TicketID)
		}
	})

	t.Run("until boundary inclusive filters newer and dateless ledgers", func(t *testing.T) {
		t.Parallel()
		until, _ := time.Parse(time.RFC3339, "2026-03-10T12:00:00Z")
		filtered := domain.FilterLedgersByCompletionDate(ledgers, nil, &until)
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "AL-1" || *filtered[1].TicketID != "AL-2" {
			t.Errorf("unexpected ledgers: %v, %v", *filtered[0].TicketID, *filtered[1].TicketID)
		}
	})

	t.Run("both boundaries filter to closed interval inclusive", func(t *testing.T) {
		t.Parallel()
		since, _ := time.Parse(time.RFC3339, "2026-03-05T00:00:00Z")
		until, _ := time.Parse(time.RFC3339, "2026-03-15T00:00:00Z")
		filtered := domain.FilterLedgersByCompletionDate(ledgers, &since, &until)
		if len(filtered) != 1 {
			t.Fatalf("expected 1 ledger, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "AL-2" {
			t.Errorf("expected AL-2, got %s", *filtered[0].TicketID)
		}
	})
}

func TestSortLedgersByCompletedAt(t *testing.T) {
	t.Parallel()

	dEarly := "2026-03-01T00:00:00Z"
	dMid := "2026-03-15T00:00:00Z"
	dLate := "2026-03-25T00:00:00Z"

	t.Run("descending sort places latest first and nulls last", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-1"), CompletedAt: &dEarly},
			{TicketID: ptr("AL-3"), CompletedAt: nil, UpdatedAt: ""},
			{TicketID: ptr("AL-2"), CompletedAt: &dLate},
			{TicketID: ptr("AL-4"), CompletedAt: &dMid},
			{TicketID: ptr("AL-5"), CompletedAt: nil, UpdatedAt: ""},
		}

		sorted := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		expectedIDs := []string{"AL-2", "AL-4", "AL-1", "AL-3", "AL-5"}
		for i, id := range expectedIDs {
			if *sorted[i].TicketID != id {
				t.Errorf("index %d: expected %s, got %s", i, id, *sorted[i].TicketID)
			}
		}
	})

	t.Run("ascending sort places earliest first and nulls last", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-1"), CompletedAt: &dEarly},
			{TicketID: ptr("AL-3"), CompletedAt: nil, UpdatedAt: ""},
			{TicketID: ptr("AL-2"), CompletedAt: &dLate},
			{TicketID: ptr("AL-4"), CompletedAt: &dMid},
			{TicketID: ptr("AL-5"), CompletedAt: nil, UpdatedAt: ""},
		}

		sorted := domain.SortLedgers(items, domain.SortFieldCompletedAt, true)
		expectedIDs := []string{"AL-1", "AL-4", "AL-2", "AL-3", "AL-5"}
		for i, id := range expectedIDs {
			if *sorted[i].TicketID != id {
				t.Errorf("index %d: expected %s, got %s", i, id, *sorted[i].TicketID)
			}
		}
	})

	t.Run("tie breaks identical dates using natural ticket ID comparison", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-10"), CompletedAt: &dMid},
			{TicketID: ptr("AL-2"), CompletedAt: &dMid},
			{TicketID: ptr("AL-1"), CompletedAt: &dMid},
		}

		sorted := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		expectedIDs := []string{"AL-1", "AL-2", "AL-10"}
		for i, id := range expectedIDs {
			if *sorted[i].TicketID != id {
				t.Errorf("index %d: expected %s, got %s", i, id, *sorted[i].TicketID)
			}
		}
	})
}

func TestGetLedgerCompletionDate_AdversarialTimezonesAndLeapYears(t *testing.T) {
	t.Parallel()

	t.Run("timezone offsets resolve to equivalent UTC timestamps", func(t *testing.T) {
		t.Parallel()
		jstDate := "2026-03-01T15:00:00+09:00"
		edtDate := "2026-03-01T02:00:00-04:00"
		utcDate := "2026-03-01T06:00:00Z"

		ledgerJST := domain.AgenticMetricsLedger{CompletedAt: &jstDate}
		ledgerEDT := domain.AgenticMetricsLedger{CompletedAt: &edtDate}
		ledgerUTC := domain.AgenticMetricsLedger{CompletedAt: &utcDate}

		tJST, okJST := domain.GetLedgerCompletionDate(ledgerJST)
		tEDT, okEDT := domain.GetLedgerCompletionDate(ledgerEDT)
		tUTC, okUTC := domain.GetLedgerCompletionDate(ledgerUTC)

		if !okJST || !okEDT || !okUTC {
			t.Fatalf("expected all timezone timestamps to parse successfully")
		}

		expectedUTC, _ := time.Parse(time.RFC3339, utcDate)
		if !tJST.Equal(expectedUTC) || !tEDT.Equal(expectedUTC) || !tUTC.Equal(expectedUTC) {
			t.Errorf("expected all timestamps to equal UTC %v; got JST=%v, EDT=%v, UTC=%v", expectedUTC, tJST, tEDT, tUTC)
		}
	})

	t.Run("leap year dates parse correctly on leap years and reject on non-leap years", func(t *testing.T) {
		t.Parallel()
		leap2024 := "2024-02-29T12:00:00Z"
		leap2028 := "2028-02-29T23:59:59Z"
		nonLeap := "2023-02-29T12:00:00Z"

		ledger2024 := domain.AgenticMetricsLedger{CompletedAt: &leap2024}
		t2024, ok2024 := domain.GetLedgerCompletionDate(ledger2024)
		if !ok2024 || t2024.Year() != 2024 || t2024.Month() != time.February || t2024.Day() != 29 {
			t.Errorf("failed parsing leap year 2024: %v, ok=%v", t2024, ok2024)
		}

		ledger2028 := domain.AgenticMetricsLedger{CompletedAt: &leap2028}
		t2028, ok2028 := domain.GetLedgerCompletionDate(ledger2028)
		if !ok2028 || t2028.Year() != 2028 || t2028.Month() != time.February || t2028.Day() != 29 {
			t.Errorf("failed parsing leap year 2028: %v, ok=%v", t2028, ok2028)
		}

		ledgerNonLeap := domain.AgenticMetricsLedger{CompletedAt: &nonLeap}
		_, okNonLeap := domain.GetLedgerCompletionDate(ledgerNonLeap)
		if okNonLeap {
			t.Errorf("expected non-leap year Feb 29 to fail parsing")
		}
	})

	t.Run("nano precision, date-only, and ISO without timezone parsing", func(t *testing.T) {
		t.Parallel()
		nanoStr := "2026-10-18T10:00:00.123456789Z"
		dateOnlyStr := "2026-10-18"
		isoNoTZStr := "2026-10-18T10:00:00"
		paddedStr := "   2026-10-18T10:00:00Z  \t "

		ledgerNano := domain.AgenticMetricsLedger{CompletedAt: &nanoStr}
		tNano, okNano := domain.GetLedgerCompletionDate(ledgerNano)
		if !okNano || tNano.Nanosecond() != 123456789 {
			t.Errorf("expected nano precision parsed: %v, ok=%v", tNano, okNano)
		}

		ledgerDateOnly := domain.AgenticMetricsLedger{CompletedAt: &dateOnlyStr}
		tDateOnly, okDateOnly := domain.GetLedgerCompletionDate(ledgerDateOnly)
		if !okDateOnly || tDateOnly.Day() != 18 || tDateOnly.Hour() != 0 {
			t.Errorf("expected date-only parsed at 00:00: %v, ok=%v", tDateOnly, okDateOnly)
		}

		ledgerISONoTZ := domain.AgenticMetricsLedger{CompletedAt: &isoNoTZStr}
		tISONoTZ, okISONoTZ := domain.GetLedgerCompletionDate(ledgerISONoTZ)
		if !okISONoTZ || tISONoTZ.Hour() != 10 {
			t.Errorf("expected ISO without TZ parsed: %v, ok=%v", tISONoTZ, okISONoTZ)
		}

		ledgerPadded := domain.AgenticMetricsLedger{CompletedAt: &paddedStr}
		tPadded, okPadded := domain.GetLedgerCompletionDate(ledgerPadded)
		if !okPadded || tPadded.Hour() != 10 {
			t.Errorf("expected padded string trimmed and parsed: %v, ok=%v", tPadded, okPadded)
		}
	})

	t.Run("does not fall back to started_at when completed_at is malformed", func(t *testing.T) {
		t.Parallel()
		bad1 := "garbage-completed"
		bad2 := "garbage-updated"
		validStarted := "2026-07-04T12:00:00Z"

		ledger := domain.AgenticMetricsLedger{
			CompletedAt: &bad1,
			UpdatedAt:   bad2,
			StartedAt:   &validStarted,
		}

		tGot, ok := domain.GetLedgerCompletionDate(ledger)
		if ok {
			t.Fatalf("expected ok = false when completed_at is malformed, got %v", tGot)
		}
		if !tGot.IsZero() {
			t.Errorf("expected zero time, got %v", tGot)
		}
	})
}

func TestFilterLedgersByCompletionDate_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	sinceUTC, _ := time.Parse(time.RFC3339, "2026-06-01T00:00:00Z")
	untilUTC, _ := time.Parse(time.RFC3339, "2026-06-01T23:59:59Z")

	t.Run("empty and nil ledgers return empty allocated slice without panicking", func(t *testing.T) {
		t.Parallel()
		filteredEmpty := domain.FilterLedgersByCompletionDate([]domain.AgenticMetricsLedger{}, &sinceUTC, &untilUTC)
		if filteredEmpty == nil || len(filteredEmpty) != 0 {
			t.Errorf("expected non-nil empty slice, got %v", filteredEmpty)
		}

		filteredNil := domain.FilterLedgersByCompletionDate(nil, &sinceUTC, &untilUTC)
		if filteredNil == nil || len(filteredNil) != 0 {
			t.Errorf("expected non-nil empty slice, got %v", filteredNil)
		}
	})

	t.Run("exact inclusive boundary timestamps: exact matches, 1ns before and 1ns after", func(t *testing.T) {
		t.Parallel()
		exactSince := "2026-06-01T00:00:00.000000000Z"
		beforeSince := "2026-05-31T23:59:59.999999999Z"
		exactUntil := "2026-06-01T23:59:59.000000000Z"
		afterUntil := "2026-06-01T23:59:59.000000001Z"

		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("EXACT-SINCE"), CompletedAt: &exactSince},
			{TicketID: ptr("BEFORE-SINCE"), CompletedAt: &beforeSince},
			{TicketID: ptr("EXACT-UNTIL"), CompletedAt: &exactUntil},
			{TicketID: ptr("AFTER-UNTIL"), CompletedAt: &afterUntil},
		}

		filtered := domain.FilterLedgersByCompletionDate(items, &sinceUTC, &untilUTC)
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "EXACT-SINCE" || *filtered[1].TicketID != "EXACT-UNTIL" {
			t.Errorf("unexpected filtered tickets: %v, %v", *filtered[0].TicketID, *filtered[1].TicketID)
		}
	})

	t.Run("inverted date range (since > until) returns empty slice without panic", func(t *testing.T) {
		t.Parallel()
		d1 := "2026-06-01T12:00:00Z"
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-1"), CompletedAt: &d1},
		}

		sinceInverted, _ := time.Parse(time.RFC3339, "2026-06-10T00:00:00Z")
		untilInverted, _ := time.Parse(time.RFC3339, "2026-06-01T00:00:00Z")

		filtered := domain.FilterLedgersByCompletionDate(items, &sinceInverted, &untilInverted)
		if filtered == nil || len(filtered) != 0 {
			t.Errorf("expected empty slice for inverted range, got %v", filtered)
		}

		// CalculateFleetSummary on inverted empty result must not divide by zero or panic
		fleet := domain.CalculateFleetSummary(filtered)
		if fleet.TotalTickets != 0 || fleet.TotalCostUSD != 0.0 {
			t.Errorf("expected zeroed fleet summary, got %+v", fleet)
		}
	})

	t.Run("timezone offsets resolved against UTC boundaries", func(t *testing.T) {
		t.Parallel()
		// sinceUTC: 2026-06-01T00:00:00Z, untilUTC: 2026-06-01T23:59:59Z
		// JST (+09:00): 2026-06-01 08:00:00+09:00 == 2026-05-31 23:00:00Z (before window)
		jstBefore := "2026-06-01T08:00:00+09:00"
		// JST (+09:00): 2026-06-02 01:00:00+09:00 == 2026-06-01 16:00:00Z (inside window)
		jstInside := "2026-06-02T01:00:00+09:00"
		// EDT (-04:00): 2026-05-31 21:00:00-04:00 == 2026-06-01 01:00:00Z (inside window)
		edtInside := "2026-05-31T21:00:00-04:00"
		// EDT (-04:00): 2026-06-01 21:00:00-04:00 == 2026-06-02 01:00:00Z (after window)
		edtAfter := "2026-06-01T21:00:00-04:00"

		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("JST-BEFORE"), CompletedAt: &jstBefore},
			{TicketID: ptr("JST-INSIDE"), CompletedAt: &jstInside},
			{TicketID: ptr("EDT-INSIDE"), CompletedAt: &edtInside},
			{TicketID: ptr("EDT-AFTER"), CompletedAt: &edtAfter},
		}

		filtered := domain.FilterLedgersByCompletionDate(items, &sinceUTC, &untilUTC)
		if len(filtered) != 2 {
			t.Fatalf("expected 2 ledgers matching timezone boundaries, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "JST-INSIDE" || *filtered[1].TicketID != "EDT-INSIDE" {
			t.Errorf("unexpected tickets: %v, %v", *filtered[0].TicketID, *filtered[1].TicketID)
		}
	})

	t.Run("leap year 2024-02-29 ledger inclusion", func(t *testing.T) {
		t.Parallel()
		leapSince, _ := time.Parse(time.RFC3339, "2024-02-28T00:00:00Z")
		leapUntil, _ := time.Parse(time.RFC3339, "2024-03-01T00:00:00Z")

		feb27 := "2024-02-27T12:00:00Z"
		feb29 := "2024-02-29T12:00:00Z"
		mar02 := "2024-03-02T12:00:00Z"

		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("FEB-27"), CompletedAt: &feb27},
			{TicketID: ptr("FEB-29"), CompletedAt: &feb29},
			{TicketID: ptr("MAR-02"), CompletedAt: &mar02},
		}

		filtered := domain.FilterLedgersByCompletionDate(items, &leapSince, &leapUntil)
		if len(filtered) != 1 || *filtered[0].TicketID != "FEB-29" {
			t.Fatalf("expected only FEB-29 on leap year, got %+v", filtered)
		}
	})

	t.Run("zero fallback: updated_at and started_at are not used and dateless ledgers are excluded", func(t *testing.T) {
		t.Parallel()
		validDate := "2026-06-01T12:00:00Z"
		badStr := "invalid-date"

		items := []domain.AgenticMetricsLedger{
			// Ledger 1: null completed_at, valid updated_at
			{TicketID: ptr("FALLBACK-UPDATED"), CompletedAt: nil, UpdatedAt: validDate},
			// Ledger 2: invalid completed_at, empty updated_at, valid started_at
			{TicketID: ptr("FALLBACK-STARTED"), CompletedAt: &badStr, UpdatedAt: "", StartedAt: &validDate},
			// Ledger 3: all null
			{TicketID: ptr("ALL-NULL"), CompletedAt: nil, UpdatedAt: "", StartedAt: nil},
			// Ledger 4: all invalid
			{TicketID: ptr("ALL-INVALID"), CompletedAt: &badStr, UpdatedAt: badStr, StartedAt: &badStr},
			// Ledger 5: valid completed_at
			{TicketID: ptr("VALID-COMPLETED"), CompletedAt: &validDate, UpdatedAt: ""},
		}

		filtered := domain.FilterLedgersByCompletionDate(items, &sinceUTC, &untilUTC)
		if len(filtered) != 1 {
			t.Fatalf("expected 1 completed ledger included and others excluded, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "VALID-COMPLETED" {
			t.Errorf("unexpected filtered ledger: %v", *filtered[0].TicketID)
		}
	})

	t.Run("zero tokens, zero cost, and zero duration ledgers compute cleanly via CalculateFleetSummary", func(t *testing.T) {
		t.Parallel()
		d := "2026-06-01T10:00:00Z"
		zeroLedger := domain.AgenticMetricsLedger{
			TicketID:    ptr("AL-ZERO"),
			CompletedAt: &d,
			Totals: domain.GroupedMetric{
				CostUSD:           0,
				Invocations:       0,
				InputTokens:       0,
				OutputTokens:      0,
				CacheReadTokens:   0,
				CacheHitRatePct:   0,
				PeakContextTokens: 0,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationSeconds:   0,
				ActiveDurationFormatted: "0m 0s",
				ReasoningTokens:         0,
			},
		}

		filtered := domain.FilterLedgersByCompletionDate([]domain.AgenticMetricsLedger{zeroLedger}, &sinceUTC, &untilUTC)
		if len(filtered) != 1 {
			t.Fatalf("expected 1 filtered ledger, got %d", len(filtered))
		}

		fleet := domain.CalculateFleetSummary(filtered)
		if fleet.TotalTickets != 1 {
			t.Errorf("expected 1 ticket in fleet, got %d", fleet.TotalTickets)
		}
		if fleet.TotalCostUSD != 0.0 || fleet.TotalInvocations != 0 || fleet.OverallCacheHitRatePct != 0.0 {
			t.Errorf("expected 0 for cost/invocations/cache-hit-rate, got cost=%v, inv=%v, hit=%v",
				fleet.TotalCostUSD, fleet.TotalInvocations, fleet.OverallCacheHitRatePct)
		}
	})
}

func TestSortLedgersByCompletedAt_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	dEarly := "2026-03-01T00:00:00Z"
	dMid := "2026-03-15T00:00:00Z"
	dLate := "2026-03-25T00:00:00Z"

	t.Run("multiple null candidates tie-break deterministically by natural ticket ID", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-10"), CompletedAt: nil, UpdatedAt: ""},
			{TicketID: ptr("AL-2"), CompletedAt: nil, UpdatedAt: ""},
			{TicketID: ptr("AL-1"), CompletedAt: &dMid},
			{TicketID: ptr("AL-01"), CompletedAt: nil, UpdatedAt: ""},
		}

		// Descending: non-null AL-1 first, then AL-01, AL-2, AL-10 (natural ticket ID)
		sortedDesc := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		expectedDesc := []string{"AL-1", "AL-01", "AL-2", "AL-10"}
		for i, exp := range expectedDesc {
			if *sortedDesc[i].TicketID != exp {
				t.Errorf("descending index %d: expected %s, got %s", i, exp, *sortedDesc[i].TicketID)
			}
		}

		// Ascending: non-null AL-1 still precedes nulls! Nulls are always last!
		sortedAsc := domain.SortLedgers(items, domain.SortFieldCompletedAt, true)
		expectedAsc := []string{"AL-1", "AL-01", "AL-2", "AL-10"}
		for i, exp := range expectedAsc {
			if *sortedAsc[i].TicketID != exp {
				t.Errorf("ascending index %d: expected %s, got %s", i, exp, *sortedAsc[i].TicketID)
			}
		}
	})

	t.Run("ledgers with nil completed_at sort last regardless of updated_at or started_at", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("EARLY-COMPLETED"), CompletedAt: &dEarly},
			{TicketID: ptr("MID-UPDATED"), CompletedAt: nil, UpdatedAt: dMid},
			{TicketID: ptr("LATE-STARTED"), CompletedAt: nil, UpdatedAt: "", StartedAt: &dLate},
		}

		sortedDesc := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		expectedIDs := []string{"EARLY-COMPLETED", "LATE-STARTED", "MID-UPDATED"}
		for i, exp := range expectedIDs {
			if *sortedDesc[i].TicketID != exp {
				t.Errorf("index %d: expected %s, got %s", i, exp, *sortedDesc[i].TicketID)
			}
		}
	})

	t.Run("identical timestamps tie-break by natural ticket ID comparison (AL-2 before AL-10)", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-10"), CompletedAt: &dMid},
			{TicketID: ptr("AL-2"), CompletedAt: &dMid},
		}

		sorted := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		if *sorted[0].TicketID != "AL-2" || *sorted[1].TicketID != "AL-10" {
			t.Errorf("expected natural order AL-2 before AL-10, got %s, %s", *sorted[0].TicketID, *sorted[1].TicketID)
		}
	})

	t.Run("immutability: input slice remains unmutated", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-1"), CompletedAt: &dEarly},
			{TicketID: ptr("AL-2"), CompletedAt: &dLate},
		}

		_ = domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		if *items[0].TicketID != "AL-1" || *items[1].TicketID != "AL-2" {
			t.Errorf("original slice was modified!")
		}
	})
}

func TestResolveDatePreset(t *testing.T) {
	t.Parallel()
	ref := time.Date(2026, 10, 19, 15, 30, 0, 0, time.UTC)

	t.Run("Last 24 Hours preset", func(t *testing.T) {
		start, end := domain.ResolveDatePreset(string(domain.TimeWindowLast24Hours), ref)
		if start == nil || end == nil {
			t.Fatalf("expected non-nil start and end")
		}
		expectedStart := ref.Add(-24 * time.Hour)
		if !start.Equal(expectedStart) {
			t.Errorf("expected start %v, got %v", expectedStart, *start)
		}
		if !end.Equal(ref) {
			t.Errorf("expected end %v, got %v", ref, *end)
		}
	})

	t.Run("Last 7 Days preset", func(t *testing.T) {
		start, end := domain.ResolveDatePreset(string(domain.TimeWindowLast7Days), ref)
		if start == nil || end == nil {
			t.Fatalf("expected non-nil start and end")
		}
		refMidnight := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
		expectedStart := refMidnight.AddDate(0, 0, -7)
		if !start.Equal(expectedStart) {
			t.Errorf("expected start %v, got %v", expectedStart, *start)
		}
		if !end.Equal(refMidnight) {
			t.Errorf("expected end %v, got %v", refMidnight, *end)
		}
	})

	t.Run("unknown preset returns nil", func(t *testing.T) {
		start, end := domain.ResolveDatePreset("Unknown", ref)
		if start != nil || end != nil {
			t.Errorf("expected nil for unknown preset")
		}
	})
}

func TestAggregateMetricsOverTime_Hourly(t *testing.T) {
	t.Parallel()
	ts1 := "2026-10-19T10:15:00Z"
	ts2 := "2026-10-19T10:45:00Z"
	ts3 := "2026-10-19T11:05:00Z"

	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID:    ptr("AL-1"),
			CompletedAt: &ts1,
			Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:          "inv-1",
					CompletedAt: &ts1,
					Tokens: domain.InvocationTokens{
						Input:       100,
						Output:      50,
						CacheRead:   20,
						PeakContext: 1000,
					},
					CostUSD:               0.05,
					ActiveDurationSeconds: 10,
				},
				{
					ID:          "inv-2",
					CompletedAt: &ts2,
					Tokens: domain.InvocationTokens{
						Input:       200,
						Output:      100,
						CacheRead:   80,
						PeakContext: 2000,
					},
					CostUSD:               0.10,
					ActiveDurationSeconds: 20,
				},
			}, nil, nil),
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				CodeDelivered: domain.CodeDeliveredStats{
					LinesAdded:   50,
					LinesDeleted: 10,
					FilesChanged: 2,
				},
			},
		},
		{
			TicketID:    ptr("AL-2"),
			CompletedAt: &ts3,
			Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:          "inv-3",
					CompletedAt: &ts3,
					Tokens: domain.InvocationTokens{
						Input:       300,
						Output:      150,
						CacheRead:   100,
						PeakContext: 1500,
					},
					CostUSD:               0.15,
					ActiveDurationSeconds: 15,
				},
			}, nil, nil),
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				CodeDelivered: domain.CodeDeliveredStats{
					LinesAdded:   20,
					LinesDeleted: 5,
					FilesChanged: 1,
				},
			},
		},
	}

	buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
	if len(buckets) != 2 {
		t.Fatalf("expected 2 hourly buckets, got %d", len(buckets))
	}

	b1 := buckets[0]
	if b1.BucketKey != "2026-10-19T10:00:00Z" {
		t.Errorf("expected bucket key 2026-10-19T10:00:00Z, got %s", b1.BucketKey)
	}
	if b1.Invocations != 2 {
		t.Errorf("expected 2 invocations in bucket 1, got %d", b1.Invocations)
	}
	if b1.InputTokens != 300 {
		t.Errorf("expected 300 input tokens in bucket 1, got %d", b1.InputTokens)
	}
	if b1.OutputTokens != 150 {
		t.Errorf("expected 150 output tokens in bucket 1, got %d", b1.OutputTokens)
	}
	if b1.CacheReadTokens != 100 {
		t.Errorf("expected 100 cache read tokens in bucket 1, got %d", b1.CacheReadTokens)
	}
	if b1.CostUSD != 0.15 {
		t.Errorf("expected 0.15 cost in bucket 1, got %f", b1.CostUSD)
	}
	if b1.LinesAdded != 50 {
		t.Errorf("expected 50 lines added in bucket 1, got %d", b1.LinesAdded)
	}

	if b1.CumulativeNetLinesAdded != 40 {
		t.Errorf("expected 40 cumulative net lines added in bucket 1, got %d", b1.CumulativeNetLinesAdded)
	}

	b2 := buckets[1]
	if b2.BucketKey != "2026-10-19T11:00:00Z" {
		t.Errorf("expected bucket key 2026-10-19T11:00:00Z, got %s", b2.BucketKey)
	}
	if b2.Invocations != 1 {
		t.Errorf("expected 1 invocation in bucket 2, got %d", b2.Invocations)
	}
	if b2.CumulativeInvocations != 3 {
		t.Errorf("expected cumulative 3 invocations, got %d", b2.CumulativeInvocations)
	}
	if b2.CumulativeInputTokens != 600 {
		t.Errorf("expected cumulative 600 input tokens, got %d", b2.CumulativeInputTokens)
	}
	if b2.CumulativeNetLinesAdded != 55 {
		t.Errorf("expected cumulative 55 net lines added in bucket 2, got %d", b2.CumulativeNetLinesAdded)
	}

	t.Run("boundary filtering by start and end timestamp", func(t *testing.T) {
		start := time.Date(2026, 10, 19, 10, 30, 0, 0, time.UTC)
		end := time.Date(2026, 10, 19, 10, 50, 0, 0, time.UTC)
		filteredBuckets := domain.AggregateMetricsOverTime(ledgers, "hourly", &start, &end)
		if len(filteredBuckets) != 1 {
			t.Fatalf("expected 1 hourly bucket within boundary, got %d", len(filteredBuckets))
		}
		if filteredBuckets[0].Invocations != 1 {
			t.Errorf("expected 1 invocation (inv-2), got %d", filteredBuckets[0].Invocations)
		}
	})
}

func TestParseTimeGranularity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected domain.TimeGranularity
		valid    bool
	}{
		{name: "hourly lowercase", input: "hourly", expected: domain.GranularityHourly, valid: true},
		{name: "hourly capitalized", input: "Hourly", expected: domain.GranularityHourly, valid: true},
		{name: "hourly uppercase", input: "HOURLY", expected: domain.GranularityHourly, valid: true},
		{name: "daily lowercase", input: "daily", expected: domain.GranularityDaily, valid: true},
		{name: "daily uppercase", input: "DAILY", expected: domain.GranularityDaily, valid: true},
		{name: "weekly lowercase", input: "weekly", expected: domain.GranularityWeekly, valid: true},
		{name: "cumulative rejected", input: "cumulative", expected: "", valid: false},
		{name: "cumulative uppercase rejected", input: "CUMULATIVE", expected: "", valid: false},
		{name: "whitespace padded", input: "  hourly  ", expected: domain.GranularityHourly, valid: true},
		{name: "invalid minutely", input: "minutely", expected: "", valid: false},
		{name: "invalid seconds", input: "seconds", expected: "", valid: false},
		{name: "empty string", input: "", expected: "", valid: false},
		{name: "whitespace only", input: "   ", expected: "", valid: false},
		{name: "random word", input: "yearly", expected: "", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := domain.ParseTimeGranularity(tt.input)
			if ok != tt.valid {
				t.Fatalf("ParseTimeGranularity(%q) ok = %v, want %v", tt.input, ok, tt.valid)
			}
			if got != tt.expected {
				t.Errorf("ParseTimeGranularity(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseTimeWindow_AL27(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected domain.TimeWindow
	}{
		{input: "1d", expected: domain.TimeWindow1D},
		{input: "custom", expected: domain.TimeWindowCustom},
		{input: "7d", expected: domain.TimeWindow7D},
		{input: "30d", expected: domain.TimeWindow30D},
		{input: "90d", expected: domain.TimeWindow90D},
		{input: "all", expected: domain.TimeWindowAll},
		{input: "unrecognized", expected: domain.TimeWindowAll},
		{input: "", expected: domain.TimeWindowAll},
	}

	for _, tt := range tests {
		got := domain.ParseTimeWindow(tt.input)
		if got != tt.expected {
			t.Errorf("ParseTimeWindow(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFilterLedgersByTimeWindow_1D_ExactBoundaries(t *testing.T) {
	t.Parallel()
	ref := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	start := ref.Add(-24 * time.Hour) // 2026-10-18T12:00:00Z

	tsAtStart := start.Format(time.RFC3339)
	tsBeforeStart := start.Add(-1 * time.Second).Format(time.RFC3339)
	tsAtEnd := ref.Format(time.RFC3339)
	tsAfterEnd := ref.Add(1 * time.Second).Format(time.RFC3339)
	tsInside := start.Add(12 * time.Hour).Format(time.RFC3339)

	ledgers := []domain.AgenticMetricsLedger{
		{TicketID: ptr("T-AT-START"), CompletedAt: &tsAtStart},
		{TicketID: ptr("T-BEFORE-START"), CompletedAt: &tsBeforeStart},
		{TicketID: ptr("T-AT-END"), CompletedAt: &tsAtEnd},
		{TicketID: ptr("T-AFTER-END"), CompletedAt: &tsAfterEnd},
		{TicketID: ptr("T-INSIDE"), CompletedAt: &tsInside},
		{TicketID: ptr("T-NO-DATES")},
	}

	filtered := domain.FilterLedgersByTimeWindow(ledgers, domain.TimeWindow1D, ref)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 ledgers within 24h window, got %d", len(filtered))
	}

	expectedIDs := map[string]bool{
		"T-AT-START": true,
		"T-AT-END":   true,
		"T-INSIDE":   true,
	}
	for _, l := range filtered {
		if !expectedIDs[*l.TicketID] {
			t.Errorf("unexpected ticket in 24h filtered list: %s", *l.TicketID)
		}
	}

	t.Run("nil and empty inputs", func(t *testing.T) {
		resNil := domain.FilterLedgersByTimeWindow(nil, domain.TimeWindow1D, ref)
		if resNil == nil || len(resNil) != 0 {
			t.Errorf("expected empty non-nil slice for nil input")
		}

		resEmpty := domain.FilterLedgersByTimeWindow([]domain.AgenticMetricsLedger{}, domain.TimeWindow1D, ref)
		if resEmpty == nil || len(resEmpty) != 0 {
			t.Errorf("expected empty non-nil slice for empty input")
		}
	})
}

func TestAggregateMetricsOverTime_AdversarialHourly(t *testing.T) {
	t.Parallel()

	t.Run("midnight transition and chronological sorting", func(t *testing.T) {
		tsLate := "2026-10-18T23:55:00Z"
		tsEarly := "2026-10-19T00:05:00Z"
		tsMid := "2026-10-19T01:30:00Z"

		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID:    ptr("AL-MIDNIGHT"),
				CompletedAt: &tsEarly,
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-early",
						CompletedAt: &tsEarly,
						Tokens:      domain.InvocationTokens{Input: 100, Output: 50, CacheRead: 25},
						CostUSD:     0.02,
					},
					{
						ID:          "inv-late",
						CompletedAt: &tsLate,
						Tokens:      domain.InvocationTokens{Input: 200, Output: 100, CacheRead: 50},
						CostUSD:     0.04,
					},
					{
						ID:          "inv-mid",
						CompletedAt: &tsMid,
						Tokens:      domain.InvocationTokens{Input: 300, Output: 150, CacheRead: 75},
						CostUSD:     0.06,
					},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
		if len(buckets) != 3 {
			t.Fatalf("expected 3 hourly buckets, got %d", len(buckets))
		}

		if buckets[0].BucketKey != "2026-10-18T23:00:00Z" {
			t.Errorf("bucket 0 key want 2026-10-18T23:00:00Z, got %s", buckets[0].BucketKey)
		}
		if buckets[1].BucketKey != "2026-10-19T00:00:00Z" {
			t.Errorf("bucket 1 key want 2026-10-19T00:00:00Z, got %s", buckets[1].BucketKey)
		}
		if buckets[2].BucketKey != "2026-10-19T01:00:00Z" {
			t.Errorf("bucket 2 key want 2026-10-19T01:00:00Z, got %s", buckets[2].BucketKey)
		}

		// Cumulative invocations check
		if buckets[2].CumulativeInvocations != 3 {
			t.Errorf("expected cumulative invocations 3, got %d", buckets[2].CumulativeInvocations)
		}
	})

	t.Run("timezone offsets normalized to UTC hourly bucket", func(t *testing.T) {
		// 14:30 +02:00 is 12:30 UTC
		tsPlus2 := "2026-10-19T14:30:00+02:00"
		// 08:30 -04:00 is 12:30 UTC
		tsMinus4 := "2026-10-19T08:30:00-04:00"

		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID: ptr("AL-TZ"),
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-plus2",
						CompletedAt: &tsPlus2,
						Tokens:      domain.InvocationTokens{Input: 100, Output: 20},
						CostUSD:     0.01,
					},
					{
						ID:          "inv-minus4",
						CompletedAt: &tsMinus4,
						Tokens:      domain.InvocationTokens{Input: 200, Output: 40},
						CostUSD:     0.02,
					},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
		if len(buckets) != 1 {
			t.Fatalf("expected both TZ offsets to normalize to single 12:00:00Z bucket, got %d buckets", len(buckets))
		}
		if buckets[0].BucketKey != "2026-10-19T12:00:00Z" {
			t.Errorf("expected bucket key 2026-10-19T12:00:00Z, got %s", buckets[0].BucketKey)
		}
		if buckets[0].Invocations != 2 {
			t.Errorf("expected 2 invocations, got %d", buckets[0].Invocations)
		}
	})

	t.Run("html datetime-local format without seconds", func(t *testing.T) {
		tsNoSeconds := "2026-10-19T16:45"
		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID: ptr("AL-NO-SEC"),
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-no-sec",
						CompletedAt: &tsNoSeconds,
						Tokens:      domain.InvocationTokens{Input: 50, Output: 10},
					},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
		if len(buckets) != 1 {
			t.Fatalf("expected 1 bucket, got %d", len(buckets))
		}
		if buckets[0].BucketKey != "2026-10-19T16:00:00Z" {
			t.Errorf("expected bucket key 2026-10-19T16:00:00Z, got %s", buckets[0].BucketKey)
		}
	})

	t.Run("zero tokens and zero division cache hit rate", func(t *testing.T) {
		ts := "2026-10-19T10:00:00Z"
		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID: ptr("AL-ZERO"),
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-zero",
						CompletedAt: &ts,
						Tokens:      domain.InvocationTokens{Input: 0, Output: 0, CacheRead: 0, Reasoning: 0},
						CostUSD:     0.0,
					},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
		if len(buckets) != 1 {
			t.Fatalf("expected 1 bucket, got %d", len(buckets))
		}
		b := buckets[0]
		if b.CacheHitRatePct != 0.0 {
			t.Errorf("expected CacheHitRatePct = 0.0 for zero tokens, got %f", b.CacheHitRatePct)
		}
		if b.CostUSD != 0.0 {
			t.Errorf("expected CostUSD = 0.0, got %f", b.CostUSD)
		}
	})

	t.Run("inverted date range returns empty slice", func(t *testing.T) {
		ts := "2026-10-19T10:00:00Z"
		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID: ptr("AL-RANGE"),
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-1",
						CompletedAt: &ts,
						Tokens:      domain.InvocationTokens{Input: 100, Output: 50},
					},
				}, nil, nil),
				},
			},
		}

		start := time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
		end := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC)
		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", &start, &end)
		if len(buckets) != 0 {
			t.Fatalf("expected 0 buckets for inverted date range, got %d", len(buckets))
		}
	})

	t.Run("empty and nil ledgers", func(t *testing.T) {
		resNil := domain.AggregateMetricsOverTime(nil, "hourly", nil, nil)
		if resNil == nil || len(resNil) != 0 {
			t.Errorf("expected empty non-nil slice for nil ledgers")
		}

		resEmpty := domain.AggregateMetricsOverTime([]domain.AgenticMetricsLedger{}, "hourly", nil, nil)
		if resEmpty == nil || len(resEmpty) != 0 {
			t.Errorf("expected empty non-nil slice for empty ledgers")
		}

		noInvs := []domain.AgenticMetricsLedger{
			{
				TicketID: ptr("NO-INVS"),
				Facts:    domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{}, nil, nil)},
			},
		}
		resNoInvs := domain.AggregateMetricsOverTime(noInvs, "hourly", nil, nil)
		if resNoInvs == nil || len(resNoInvs) != 0 {
			t.Errorf("expected empty non-nil slice for ledger with no invocations and no code churn")
		}
	})

	t.Run("unparseable timestamps ignored gracefully", func(t *testing.T) {
		badDate := "not-a-valid-date"
		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID:    ptr("AL-CORRUPTED"),
				CompletedAt: &badDate,
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:          "inv-bad",
						CompletedAt: &badDate,
						Tokens:      domain.InvocationTokens{Input: 100, Output: 50},
					},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "hourly", nil, nil)
		if len(buckets) != 0 {
			t.Fatalf("expected 0 buckets for unparseable dates, got %d", len(buckets))
		}
	})
}

func TestCanonicalTicketType(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
	}{
		{"feat", "feature"},
		{"FEAT", "feature"},
		{"feature", "feature"},
		{" FEATURE ", "feature"},
		{"bug", "bug"},
		{"BUG", "bug"},
		{"fix", "bug"},
		{"FIX", "bug"},
		{"bugfix", "bug"},
		{"BugFix", "bug"},
		{"doc", "docs"},
		{"DOC", "docs"},
		{"docs", "docs"},
		{" DOCS ", "docs"},
		{"chore", "chore"},
		{"CHORE", "chore"},
		{"refactor", "refactor"},
		{"ci", "ci"},
		{"test", "test"},
		{"", ""},
		{"   ", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := domain.CanonicalTicketType(tc.input)
			if got != tc.expected {
				t.Errorf("CanonicalTicketType(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestFilterLedgersByTicketTypes(t *testing.T) {
	t.Parallel()

	featLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-1"), TicketType: ptr(domain.ConventionalTicketType("feature"))}
	featAliasLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-2"), TicketType: ptr(domain.TicketTypeFeat)}
	bugLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-3"), TicketType: ptr(domain.ConventionalTicketType("bug"))}
	bugFixLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-4"), TicketType: ptr(domain.TicketTypeFix)}
	choreLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-5"), TicketType: ptr(domain.TicketTypeChore)}
	docsLedger := domain.AgenticMetricsLedger{TicketID: ptr("PO-1"), TicketType: ptr(domain.TicketTypeDocs)}
	nilTypeLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-6"), TicketType: nil}

	allLedgers := []domain.AgenticMetricsLedger{
		featLedger, featAliasLedger, bugLedger, bugFixLedger, choreLedger, docsLedger, nilTypeLedger,
	}

	t.Run("empty ledgers returns empty slice", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(nil, []string{"feature"})
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(res))
		}
	})

	t.Run("empty ticketTypes returns copy of all ledgers", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(allLedgers, nil)
		if len(res) != len(allLedgers) {
			t.Fatalf("expected %d ledgers, got %d", len(allLedgers), len(res))
		}
	})

	t.Run("ticketTypes with all or empty string returns all ledgers", func(t *testing.T) {
		t.Parallel()
		res1 := domain.FilterLedgersByTicketTypes(allLedgers, []string{"all"})
		if len(res1) != len(allLedgers) {
			t.Errorf("expected all ledgers for ['all'], got %d", len(res1))
		}
		res2 := domain.FilterLedgersByTicketTypes(allLedgers, []string{"ALL"})
		if len(res2) != len(allLedgers) {
			t.Errorf("expected all ledgers for ['ALL'], got %d", len(res2))
		}
		res3 := domain.FilterLedgersByTicketTypes(allLedgers, []string{"feature", "all"})
		if len(res3) != len(allLedgers) {
			t.Errorf("expected all ledgers when 'all' is included, got %d", len(res3))
		}
		res4 := domain.FilterLedgersByTicketTypes(allLedgers, []string{""})
		if len(res4) != len(allLedgers) {
			t.Errorf("expected all ledgers for [''], got %d", len(res4))
		}
	})

	t.Run("single selection matches canonical and aliases", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(allLedgers, []string{"feature"})
		if len(res) != 2 {
			t.Fatalf("expected 2 feature ledgers, got %d", len(res))
		}
		if *res[0].TicketID != "AL-1" || *res[1].TicketID != "AL-2" {
			t.Errorf("unexpected matches: %+v", res)
		}

		resAlias := domain.FilterLedgersByTicketTypes(allLedgers, []string{"feat"})
		if len(resAlias) != 2 {
			t.Fatalf("expected 2 feature ledgers for alias feat, got %d", len(resAlias))
		}
	})

	t.Run("multi-selection combines matching types", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(allLedgers, []string{"feature", "bug"})
		if len(res) != 4 {
			t.Fatalf("expected 4 ledgers (feature + bug), got %d", len(res))
		}
		expectedIDs := map[string]bool{"AL-1": true, "AL-2": true, "AL-3": true, "AL-4": true}
		for _, l := range res {
			if !expectedIDs[*l.TicketID] {
				t.Errorf("unexpected ticket ID in multi-filter: %s", *l.TicketID)
			}
		}
	})

	t.Run("case insensitivity and whitespace trimming", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(allLedgers, []string{" FEATURE ", "BuG"})
		if len(res) != 4 {
			t.Fatalf("expected 4 ledgers, got %d", len(res))
		}
	})

	t.Run("disjoint query returns empty slice", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketTypes(allLedgers, []string{"nonexistent"})
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers for nonexistent query, got %d", len(res))
		}
	})

	t.Run("immutability of input slice", func(t *testing.T) {
		t.Parallel()
		input := make([]domain.AgenticMetricsLedger, len(allLedgers))
		copy(input, allLedgers)
		res := domain.FilterLedgersByTicketTypes(input, []string{"feature"})
		if len(res) > 0 {
			res[0].UpdatedAt = "MUTATED"
		}
		if input[0].UpdatedAt == "MUTATED" {
			t.Errorf("input ledger was mutated!")
		}
	})
}

func TestFilterLedgersByTicketType_Wrapper(t *testing.T) {
	t.Parallel()
	featLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-1"), TicketType: ptr(domain.ConventionalTicketType("feature"))}
	bugLedger := domain.AgenticMetricsLedger{TicketID: ptr("AL-2"), TicketType: ptr(domain.ConventionalTicketType("bug"))}
	allLedgers := []domain.AgenticMetricsLedger{featLedger, bugLedger}

	t.Run("comma-separated types", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByTicketType(allLedgers, "feature,bug")
		if len(res) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res))
		}
	})

	t.Run("empty string and all pass through", func(t *testing.T) {
		t.Parallel()
		res1 := domain.FilterLedgersByTicketType(allLedgers, "")
		if len(res1) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res1))
		}
		res2 := domain.FilterLedgersByTicketType(allLedgers, "all")
		if len(res2) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res2))
		}
	})
}

func TestExtractTicketPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
	}{
		{"AL-1", "AL"},
		{"al-1", "AL"},
		{"AL-28", "AL"},
		{"GEN_4", "GEN"},
		{"gen_4", "GEN"},
		{"PO-2", "PO"},
		{"al10", "AL"},
		{"AL-", "AL"},
		{"ticket-xyz", "TICKET"},
		{"ticket_xyz", "TICKET"},
		{"123-abc", ""},
		{"123", ""},
		{"", ""},
		{"   ", ""},
		{"UNKNOWN", ""},
		{"AL.1", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := domain.ExtractTicketPrefix(tc.input)
			if got != tc.expected {
				t.Errorf("ExtractTicketPrefix(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestFilterLedgersByPrefixes(t *testing.T) {
	t.Parallel()
	al1 := domain.AgenticMetricsLedger{TicketID: ptr("AL-1")}
	al28 := domain.AgenticMetricsLedger{TicketID: ptr("al-28")}
	gen4 := domain.AgenticMetricsLedger{TicketID: ptr("GEN_4")}
	po2 := domain.AgenticMetricsLedger{TicketID: ptr("PO-2")}
	nilID := domain.AgenticMetricsLedger{TicketID: nil}
	invalidID := domain.AgenticMetricsLedger{TicketID: ptr("123")}

	allLedgers := []domain.AgenticMetricsLedger{al1, al28, gen4, po2, nilID, invalidID}

	t.Run("empty ledgers returns empty slice", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefixes(nil, []string{"AL"})
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(res))
		}
	})

	t.Run("empty prefixes returns copy of all ledgers", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefixes(allLedgers, nil)
		if len(res) != len(allLedgers) {
			t.Fatalf("expected %d ledgers, got %d", len(allLedgers), len(res))
		}
	})

	t.Run("prefixes with all or empty string returns all ledgers", func(t *testing.T) {
		t.Parallel()
		res1 := domain.FilterLedgersByPrefixes(allLedgers, []string{"all"})
		if len(res1) != len(allLedgers) {
			t.Errorf("expected all ledgers for ['all'], got %d", len(res1))
		}
		res2 := domain.FilterLedgersByPrefixes(allLedgers, []string{"AL", "all"})
		if len(res2) != len(allLedgers) {
			t.Errorf("expected all ledgers when 'all' is included, got %d", len(res2))
		}
		res3 := domain.FilterLedgersByPrefixes(allLedgers, []string{""})
		if len(res3) != len(allLedgers) {
			t.Errorf("expected all ledgers for [''], got %d", len(res3))
		}
	})

	t.Run("multi-prefix selection", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefixes(allLedgers, []string{"AL", "GEN"})
		if len(res) != 3 {
			t.Fatalf("expected 3 ledgers (AL + GEN), got %d", len(res))
		}
		expectedIDs := map[string]bool{"AL-1": true, "al-28": true, "GEN_4": true}
		for _, l := range res {
			if !expectedIDs[*l.TicketID] {
				t.Errorf("unexpected ticket ID: %s", *l.TicketID)
			}
		}
	})

	t.Run("case insensitivity", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefixes(allLedgers, []string{"al", "gen"})
		if len(res) != 3 {
			t.Fatalf("expected 3 ledgers, got %d", len(res))
		}
	})

	t.Run("disjoint prefix query returns empty slice", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefixes(allLedgers, []string{"XYZ"})
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers for XYZ, got %d", len(res))
		}
	})
}

func TestFilterLedgersByPrefix_Wrapper(t *testing.T) {
	t.Parallel()
	al1 := domain.AgenticMetricsLedger{TicketID: ptr("AL-1")}
	gen1 := domain.AgenticMetricsLedger{TicketID: ptr("GEN-1")}
	allLedgers := []domain.AgenticMetricsLedger{al1, gen1}

	t.Run("comma-separated prefixes", func(t *testing.T) {
		t.Parallel()
		res := domain.FilterLedgersByPrefix(allLedgers, "AL,GEN")
		if len(res) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res))
		}
	})

	t.Run("empty string and all pass through", func(t *testing.T) {
		t.Parallel()
		res1 := domain.FilterLedgersByPrefix(allLedgers, "")
		if len(res1) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res1))
		}
		res2 := domain.FilterLedgersByPrefix(allLedgers, "all")
		if len(res2) != 2 {
			t.Errorf("expected 2 ledgers, got %d", len(res2))
		}
	})
}

func TestExtractDistinctPrefixes(t *testing.T) {
	t.Parallel()

	t.Run("empty ledgers returns empty slice", func(t *testing.T) {
		t.Parallel()
		res := domain.ExtractDistinctPrefixes(nil)
		if len(res) != 0 {
			t.Errorf("expected 0 prefixes, got %d", len(res))
		}
	})

	t.Run("deduplicates and sorts in ascending alphabetical order", func(t *testing.T) {
		t.Parallel()
		ledgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("PO-2")},
			{TicketID: ptr("AL-10")},
			{TicketID: ptr("GEN-1")},
			{TicketID: ptr("AL-1")},
			{TicketID: ptr("PO-1")},
			{TicketID: nil},
			{TicketID: ptr("")},
			{TicketID: ptr("123")},
		}

		res := domain.ExtractDistinctPrefixes(ledgers)
		expected := []string{"AL", "GEN", "PO"}
		if len(res) != len(expected) {
			t.Fatalf("expected %d prefixes, got %d: %+v", len(expected), len(res), res)
		}
		for i, exp := range expected {
			if res[i] != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, res[i])
			}
		}
	})
}

func TestMultiFilter_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty and nil inputs across filter functions", func(t *testing.T) {
		t.Parallel()

		// FilterLedgersByTicketTypes with nil and empty slices
		resNilLedgers := domain.FilterLedgersByTicketTypes(nil, []string{"feature"})
		if resNilLedgers == nil || len(resNilLedgers) != 0 {
			t.Errorf("expected non-nil empty slice for nil ledgers, got %+v", resNilLedgers)
		}

		resEmptyLedgers := domain.FilterLedgersByTicketTypes([]domain.AgenticMetricsLedger{}, []string{"feature"})
		if resEmptyLedgers == nil || len(resEmptyLedgers) != 0 {
			t.Errorf("expected non-nil empty slice for empty ledgers, got %+v", resEmptyLedgers)
		}

		sample := []domain.AgenticMetricsLedger{
			{TicketID: ptr("AL-1"), TicketType: ptr(domain.TicketTypeFeat)},
			{TicketID: ptr("AL-2"), TicketType: ptr(domain.TicketTypeFix)},
		}

		resEmptyTypes := domain.FilterLedgersByTicketTypes(sample, []string{})
		if len(resEmptyTypes) != 2 {
			t.Errorf("expected 2 ledgers for empty types, got %d", len(resEmptyTypes))
		}

		resWhitespaceTypes := domain.FilterLedgersByTicketTypes(sample, []string{"   ", "\t", "\n"})
		if len(resWhitespaceTypes) != 2 {
			t.Errorf("expected all ledgers retained for whitespace types, got %d", len(resWhitespaceTypes))
		}

		// FilterLedgersByPrefixes with nil and empty slices
		resNilPrefixes := domain.FilterLedgersByPrefixes(nil, []string{"AL"})
		if resNilPrefixes == nil || len(resNilPrefixes) != 0 {
			t.Errorf("expected non-nil empty slice for nil ledgers, got %+v", resNilPrefixes)
		}

		resEmptyPrefixLedgers := domain.FilterLedgersByPrefixes([]domain.AgenticMetricsLedger{}, []string{"AL"})
		if resEmptyPrefixLedgers == nil || len(resEmptyPrefixLedgers) != 0 {
			t.Errorf("expected non-nil empty slice for empty ledgers, got %+v", resEmptyPrefixLedgers)
		}

		resEmptyPrefixes := domain.FilterLedgersByPrefixes(sample, []string{})
		if len(resEmptyPrefixes) != 2 {
			t.Errorf("expected 2 ledgers for empty prefixes, got %d", len(resEmptyPrefixes))
		}

		resWhitespacePrefixes := domain.FilterLedgersByPrefixes(sample, []string{"   ", "\t", "\n"})
		if len(resWhitespacePrefixes) != 2 {
			t.Errorf("expected all ledgers retained for whitespace prefixes, got %d", len(resWhitespacePrefixes))
		}

		// ExtractDistinctPrefixes with nil and empty slices
		resNilDistinct := domain.ExtractDistinctPrefixes(nil)
		if resNilDistinct == nil || len(resNilDistinct) != 0 {
			t.Errorf("expected non-nil empty slice for nil ledgers, got %+v", resNilDistinct)
		}

		resEmptyDistinct := domain.ExtractDistinctPrefixes([]domain.AgenticMetricsLedger{})
		if resEmptyDistinct == nil || len(resEmptyDistinct) != 0 {
			t.Errorf("expected non-nil empty slice for empty ledgers, got %+v", resEmptyDistinct)
		}
	})

	t.Run("nil and pathological ticket types and ticket IDs", func(t *testing.T) {
		t.Parallel()

		pathologicalLedgers := []domain.AgenticMetricsLedger{
			{TicketID: nil, TicketType: nil},
			{TicketID: ptr(""), TicketType: ptr(domain.ConventionalTicketType(""))},
			{TicketID: ptr("   "), TicketType: ptr(domain.ConventionalTicketType("   "))},
			{TicketID: ptr("123-abc"), TicketType: ptr(domain.ConventionalTicketType("\t\n"))},
			{TicketID: ptr("___"), TicketType: ptr(domain.ConventionalTicketType("unknown_custom_type"))},
			{TicketID: ptr("AL-1"), TicketType: ptr(domain.TicketTypeFeat)},
		}

		// Querying concrete ticket types excludes all nil/empty/pathological types
		resType := domain.FilterLedgersByTicketTypes(pathologicalLedgers, []string{"feature"})
		if len(resType) != 1 || *resType[0].TicketID != "AL-1" {
			t.Fatalf("expected only AL-1 for feature query, got %+v", resType)
		}

		// Querying all retains all 6 ledgers including nil/empty
		resAllType := domain.FilterLedgersByTicketTypes(pathologicalLedgers, []string{"all"})
		if len(resAllType) != len(pathologicalLedgers) {
			t.Errorf("expected all 6 ledgers retained for 'all' query, got %d", len(resAllType))
		}

		// Querying concrete prefix excludes nil/empty/unparseable IDs
		resPrefix := domain.FilterLedgersByPrefixes(pathologicalLedgers, []string{"AL"})
		if len(resPrefix) != 1 || *resPrefix[0].TicketID != "AL-1" {
			t.Fatalf("expected only AL-1 for prefix AL query, got %+v", resPrefix)
		}

		// Querying all prefixes retains all 6 ledgers
		resAllPrefix := domain.FilterLedgersByPrefixes(pathologicalLedgers, []string{"all"})
		if len(resAllPrefix) != len(pathologicalLedgers) {
			t.Errorf("expected all 6 ledgers retained for 'all' prefix query, got %d", len(resAllPrefix))
		}

		// ExtractDistinctPrefixes ignores all nil/empty/non-conforming IDs
		distinct := domain.ExtractDistinctPrefixes(pathologicalLedgers)
		if len(distinct) != 1 || distinct[0] != "AL" {
			t.Errorf("expected only ['AL'], got %+v", distinct)
		}
	})

	t.Run("symmetrical alias expansion matrix", func(t *testing.T) {
		t.Parallel()

		aliasLedgers := []domain.AgenticMetricsLedger{
			{TicketID: ptr("T-1"), TicketType: ptr(domain.TicketTypeFeat)},
			{TicketID: ptr("T-2"), TicketType: ptr(domain.ConventionalTicketType("feature"))},
			{TicketID: ptr("T-3"), TicketType: ptr(domain.ConventionalTicketType("bug"))},
			{TicketID: ptr("T-4"), TicketType: ptr(domain.TicketTypeFix)},
			{TicketID: ptr("T-5"), TicketType: ptr(domain.ConventionalTicketType("bugfix"))},
			{TicketID: ptr("T-6"), TicketType: ptr(domain.ConventionalTicketType("doc"))},
			{TicketID: ptr("T-7"), TicketType: ptr(domain.TicketTypeDocs)},
			{TicketID: ptr("T-8"), TicketType: ptr(domain.TicketTypeChore)},
		}

		// Test "feature" aliases: "feat" and "feature"
		for _, q := range []string{"feat", "FEAT", "feature", "FEATURE", "  FeAt  "} {
			matched := domain.FilterLedgersByTicketTypes(aliasLedgers, []string{q})
			if len(matched) != 2 {
				t.Errorf("query %q: expected 2 matches, got %d", q, len(matched))
			}
			for _, m := range matched {
				if *m.TicketID != "T-1" && *m.TicketID != "T-2" {
					t.Errorf("query %q: unexpected match %s", q, *m.TicketID)
				}
			}
		}

		// Test "bug" aliases: "bug", "fix", and "bugfix"
		for _, q := range []string{"bug", "BUG", "fix", "FIX", "bugfix", "BUGFIX", "  bUgFiX\t"} {
			matched := domain.FilterLedgersByTicketTypes(aliasLedgers, []string{q})
			if len(matched) != 3 {
				t.Errorf("query %q: expected 3 matches, got %d", q, len(matched))
			}
			for _, m := range matched {
				if *m.TicketID != "T-3" && *m.TicketID != "T-4" && *m.TicketID != "T-5" {
					t.Errorf("query %q: unexpected match %s", q, *m.TicketID)
				}
			}
		}

		// Test "docs" aliases: "doc" and "docs"
		for _, q := range []string{"doc", "DOC", "docs", "DOCS", "  DoC  "} {
			matched := domain.FilterLedgersByTicketTypes(aliasLedgers, []string{q})
			if len(matched) != 2 {
				t.Errorf("query %q: expected 2 matches, got %d", q, len(matched))
			}
			for _, m := range matched {
				if *m.TicketID != "T-6" && *m.TicketID != "T-7" {
					t.Errorf("query %q: unexpected match %s", q, *m.TicketID)
				}
			}
		}

		// Mixed alias combination: fix + doc -> matches T-3, T-4, T-5, T-6, T-7 (5 ledgers)
		matchedMixed := domain.FilterLedgersByTicketTypes(aliasLedgers, []string{"fix", "doc"})
		if len(matchedMixed) != 5 {
			t.Fatalf("expected 5 matches for fix+doc, got %d", len(matchedMixed))
		}

		// All conventional types combined: feat + bugfix + docs + chore -> all 8 ledgers
		matchedAllConv := domain.FilterLedgersByTicketTypes(aliasLedgers, []string{"feat", "bugfix", "docs", "chore"})
		if len(matchedAllConv) != 8 {
			t.Fatalf("expected 8 matches for feat+bugfix+docs+chore, got %d", len(matchedAllConv))
		}
	})

	t.Run("prefix extraction corner cases", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			id       string
			expected string
		}{
			{"AL-100", "AL"},
			{"al_200", "AL"},
			{"GEN-0", "GEN"},
			{"PO-9999", "PO"},
			{"PROD_FIX-1", "PROD"},
			{"FEATURE-42", "FEATURE"},
			{"al123", "AL"},
			{"AL-", "AL"},
			{"AL_", "AL"},
			{"---", ""},
			{"___", ""},
			{"-123", ""},
			{"_123", ""},
			{"123-AL", ""},
			{"123_AL", ""},
			{"123456", ""},
			{"", ""},
			{"   ", ""},
			{"\t\n", ""},
		}

		for _, tc := range cases {
			got := domain.ExtractTicketPrefix(tc.id)
			if got != tc.expected {
				t.Errorf("ExtractTicketPrefix(%q) = %q, expected %q", tc.id, got, tc.expected)
			}
		}
	})
}

func TestComputeWorkflowEfficiency_HierarchyDecoupling(t *testing.T) {
	t.Parallel()

	t.Run("Case A: Single root invocation with role general", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       "ses-root-gen",
				ParentID: nil,
				Role:     "general",
				CostUSD:  1.50,
				Turns:    12,
				Tokens: domain.InvocationTokens{
					PeakContext: 60000,
				},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.RootPeakTokens != 60000 {
			t.Errorf("expected RootPeakTokens 60000, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
		expectedSat := math.Round(60000.0/float64(domain.HealthyCeilingTokens)*1000.0) / 10.0
		if eff.ContextWindowHealth.RootSaturationPct != expectedSat {
			t.Errorf("expected RootSaturationPct %f, got %f", expectedSat, eff.ContextWindowHealth.RootSaturationPct)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct, got %+v", eff.RoleEffortDistributionPct)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected AvgTurnsPerInvocation 0.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 0 {
			t.Errorf("expected PeakTurns 0, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
	})

	t.Run("Case B: Single root invocation with role build or custom root persona", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       "ses-root-build",
				ParentID: nil,
				Role:     "build",
				CostUSD:  2.25,
				Turns:    8,
				Tokens: domain.InvocationTokens{
					PeakContext: 75000,
				},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.RootPeakTokens != 75000 {
			t.Errorf("expected RootPeakTokens 75000, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct, got %+v", eff.RoleEffortDistributionPct)
		}
	})

	t.Run("Case C: Root session with role general and child subagents", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root-1"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       rootID,
				ParentID: nil,
				Role:     "general",
				CostUSD:  1.00,
				Turns:    5,
				Tokens: domain.InvocationTokens{
					PeakContext: 40000,
				},
			},
			{
				ID:       "ses-planner",
				ParentID: &rootID,
				Role:     "planner",
				CostUSD:  2.00,
				Turns:    10,
				Tokens: domain.InvocationTokens{
					PeakContext: 80000,
				},
			},
			{
				ID:       "ses-coder",
				ParentID: &rootID,
				Role:     "coder",
				CostUSD:  4.00,
				Turns:    20,
				Tokens: domain.InvocationTokens{
					PeakContext: 150000,
				},
			},
			{
				ID:       "ses-tester",
				ParentID: &rootID,
				Role:     "tester",
				CostUSD:  2.00,
				Turns:    10,
				Tokens: domain.InvocationTokens{
					PeakContext: 90000,
				},
			},
			{
				ID:       "ses-reviewer",
				ParentID: &rootID,
				Role:     "reviewer",
				CostUSD:  2.00,
				Turns:    10,
				Tokens: domain.InvocationTokens{
					PeakContext: 70000,
				},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.RootPeakTokens != 40000 {
			t.Errorf("expected RootPeakTokens 40000, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 150000 {
			t.Errorf("expected SubagentsPeakTokens 150000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected SubagentsPeakRole 'coder', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		// Effort distribution sums only child subagents: 2 + 4 + 2 + 2 = 10.0
		// planner: 2/10 = 20%, coder: 4/10 = 40%, tester: 2/10 = 20%, reviewer: 2/10 = 20%
		if eff.RoleEffortDistributionPct["general"] != 0.0 {
			t.Errorf("general root role should not appear in RoleEffortDistributionPct, got %f", eff.RoleEffortDistributionPct["general"])
		}
		if eff.RoleEffortDistributionPct["coder"] != 40.0 {
			t.Errorf("expected coder 40.0%%, got %f", eff.RoleEffortDistributionPct["coder"])
		}
		if eff.RoleEffortDistributionPct["planner"] != 20.0 {
			t.Errorf("expected planner 20.0%%, got %f", eff.RoleEffortDistributionPct["planner"])
		}
		var totalEffort float64
		for _, pct := range eff.RoleEffortDistributionPct {
			totalEffort += pct
		}
		if totalEffort < 99.9 || totalEffort > 100.1 {
			t.Errorf("expected subagent effort distribution to sum to ~100%%, got %f", totalEffort)
		}
	})

	t.Run("Case D: Empty ledger with zero invocations", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{}, nil, nil)}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.RootPeakTokens != 0 {
			t.Errorf("expected RootPeakTokens 0, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct, got %+v", eff.RoleEffortDistributionPct)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected AvgTurnsPerInvocation 0.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 0 {
			t.Errorf("expected PeakTurns 0, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
	})

	t.Run("Case E: Multi-tier child subagents with ParentID != nil", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root-0"
		tier1ID := "ses-planner-1"
		tier2ID := "ses-coder-2"

		// Deep hierarchy: root -> planner -> coder -> tester
		// All tiers where ParentID != nil must be classified as delegated subagents
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       rootID,
				ParentID: nil,
				Role:     "general",
				CostUSD:  1.00,
				Turns:    3,
				Tokens: domain.InvocationTokens{
					PeakContext: 35000,
				},
			},
			{
				ID:       tier1ID,
				ParentID: &rootID,
				Role:     "planner",
				CostUSD:  2.00,
				Turns:    5,
				Tokens: domain.InvocationTokens{
					PeakContext: 65000,
				},
			},
			{
				ID:       tier2ID,
				ParentID: &tier1ID,
				Role:     "coder",
				CostUSD:  5.00,
				Turns:    15,
				Tokens: domain.InvocationTokens{
					PeakContext: 145000,
				},
			},
			{
				ID:       "ses-tester-3",
				ParentID: &tier2ID,
				Role:     "tester",
				CostUSD:  3.00,
				Turns:    10,
				Tokens: domain.InvocationTokens{
					PeakContext: 95000,
				},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)

		// Root coordination metrics must exclusively capture the ParentID == nil invocation
		if eff.ContextWindowHealth.RootPeakTokens != 35000 {
			t.Errorf("expected RootPeakTokens 35000, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}

		// Delegated subagents metrics must capture all tiers with ParentID != nil
		if eff.ContextWindowHealth.SubagentsPeakTokens != 145000 {
			t.Errorf("expected SubagentsPeakTokens 145000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected SubagentsPeakRole 'coder', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}

		// Total subagent cost = 2.0 + 5.0 + 3.0 = 10.0
		// Planner = 20%, Coder = 50%, Tester = 30%
		if eff.RoleEffortDistributionPct["general"] != 0.0 {
			t.Errorf("root role 'general' must not appear in RoleEffortDistributionPct")
		}
		if eff.RoleEffortDistributionPct["planner"] != 20.0 {
			t.Errorf("expected planner 20.0%%, got %f", eff.RoleEffortDistributionPct["planner"])
		}
		if eff.RoleEffortDistributionPct["coder"] != 50.0 {
			t.Errorf("expected coder 50.0%%, got %f", eff.RoleEffortDistributionPct["coder"])
		}
		if eff.RoleEffortDistributionPct["tester"] != 30.0 {
			t.Errorf("expected tester 30.0%%, got %f", eff.RoleEffortDistributionPct["tester"])
		}

		// Subagent turn velocity: total turns = 5 + 15 + 10 = 30 across 3 subagents -> 10.0 avg
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 10.0 {
			t.Errorf("expected AvgTurnsPerInvocation 10.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 15 {
			t.Errorf("expected PeakTurns 15, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
	})

	t.Run("Case F: Diverse root personas table-driven", func(t *testing.T) {
		t.Parallel()
		personas := []struct {
			role        string
			peakContext int64
			costUSD     float64
			turns       int64
		}{
			{role: "general", peakContext: 42000, costUSD: 1.10, turns: 4},
			{role: "build", peakContext: 58000, costUSD: 2.20, turns: 6},
			{role: "orchestrator", peakContext: 73000, costUSD: 3.30, turns: 8},
			{role: "po", peakContext: 89000, costUSD: 4.40, turns: 10},
			{role: "custom-root", peakContext: 61000, costUSD: 1.50, turns: 5},
			{role: "root", peakContext: 39000, costUSD: 0.90, turns: 3},
			{role: "  GENERAL  ", peakContext: 55000, costUSD: 1.20, turns: 7},
			{role: "", peakContext: 25000, costUSD: 0.50, turns: 2},
		}

		for _, tc := range personas {
			tc := tc
			t.Run("persona_"+tc.role, func(t *testing.T) {
				t.Parallel()
				ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:       "ses-single-root",
						ParentID: nil,
						Role:     domain.AgentRole(tc.role),
						CostUSD:  tc.costUSD,
						Turns:    tc.turns,
						Tokens: domain.InvocationTokens{
							PeakContext: tc.peakContext,
						},
					},
				}, nil, nil),
				}

				eff := domain.ComputeWorkflowEfficiency(ledger)
				if eff.ContextWindowHealth.RootPeakTokens != tc.peakContext {
					t.Errorf("role %q: expected RootPeakTokens %d, got %d", tc.role, tc.peakContext, eff.ContextWindowHealth.RootPeakTokens)
				}
				if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
					t.Errorf("role %q: expected SubagentsPeakTokens 0, got %d", tc.role, eff.ContextWindowHealth.SubagentsPeakTokens)
				}
				if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
					t.Errorf("role %q: expected SubagentsPeakRole 'none', got %s", tc.role, eff.ContextWindowHealth.SubagentsPeakRole)
				}
				if len(eff.RoleEffortDistributionPct) != 0 {
					t.Errorf("role %q: expected empty RoleEffortDistributionPct, got %+v", tc.role, eff.RoleEffortDistributionPct)
				}
				if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
					t.Errorf("role %q: expected AvgTurnsPerInvocation 0.0, got %f", tc.role, eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
				}
				if eff.SubagentTurnVelocity.PeakTurns != 0 {
					t.Errorf("role %q: expected PeakTurns 0, got %d", tc.role, eff.SubagentTurnVelocity.PeakTurns)
				}
			})
		}
	})

	t.Run("Case G: Child subagents with zero peak tokens and tie-breaking", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root-0"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       rootID,
				ParentID: nil,
				Role:     "general",
				CostUSD:  1.0,
				Tokens: domain.InvocationTokens{
					PeakContext: 50000,
				},
			},
			{
				ID:       "ses-coder",
				ParentID: &rootID,
				Role:     "coder",
				CostUSD:  2.0,
				Tokens: domain.InvocationTokens{
					PeakContext: 0,
				},
			},
			{
				ID:       "ses-planner",
				ParentID: &rootID,
				Role:     "planner",
				CostUSD:  1.0,
				Tokens: domain.InvocationTokens{
					PeakContext: 0,
				},
			},
			{
				ID:       "ses-reviewer",
				ParentID: &rootID,
				Role:     "reviewer",
				CostUSD:  1.0,
				Tokens: domain.InvocationTokens{
					PeakContext: 0,
				},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		// When multiple subagents share peak tokens (0), alphabetical order breaks ties by default:
		// coder comes before planner and reviewer
		if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected SubagentsPeakRole 'coder' via default tie-break, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		// When custom role order is supplied, it takes priority:
		effCustom := domain.ComputeWorkflowEfficiency(ledger, []string{"planner", "coder", "reviewer"})
		if effCustom.ContextWindowHealth.SubagentsPeakRole != "planner" {
			t.Errorf("expected SubagentsPeakRole 'planner' via custom tie-break, got %s", effCustom.ContextWindowHealth.SubagentsPeakRole)
		}
		// Subagent costs must still be computed: total = 4.0, coder = 50%, planner = 25%, reviewer = 25%
		if eff.RoleEffortDistributionPct["coder"] != 50.0 {
			t.Errorf("expected coder 50.0%%, got %f", eff.RoleEffortDistributionPct["coder"])
		}
		if eff.RoleEffortDistributionPct["planner"] != 25.0 {
			t.Errorf("expected planner 25.0%%, got %f", eff.RoleEffortDistributionPct["planner"])
		}
	})

	t.Run("Case H: Boundary conditions and division by zero defense", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root-h"

		t.Run("zero total subagent cost with active subagents", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					CostUSD:  0.0,
					Tokens:   domain.InvocationTokens{PeakContext: 10000},
				},
				{
					ID:       "ses-c1",
					ParentID: &rootID,
					Role:     "coder",
					CostUSD:  0.0,
					Tokens:   domain.InvocationTokens{PeakContext: 20000},
				},
				{
					ID:       "ses-p1",
					ParentID: &rootID,
					Role:     "planner",
					CostUSD:  0.0,
					Tokens:   domain.InvocationTokens{PeakContext: 15000},
				},
			}, nil, nil),
			}

			eff := domain.ComputeWorkflowEfficiency(ledger)
			if len(eff.RoleEffortDistributionPct) != 0 {
				t.Errorf("expected empty RoleEffortDistributionPct on 0 subagent cost, got %+v", eff.RoleEffortDistributionPct)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 20000 {
				t.Errorf("expected SubagentsPeakTokens 20000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
				t.Errorf("expected SubagentsPeakRole 'coder', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("nil invocations slice", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil, nil, nil)}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.RootPeakTokens != 0 {
				t.Errorf("expected RootPeakTokens 0, got %d", eff.ContextWindowHealth.RootPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
				t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
				t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("child invocation with non-subagent role string classified structurally", func(t *testing.T) {
			t.Parallel()
			// Under pure structural classification, child invocations (ParentID != nil) are treated
			// as subagents regardless of their persona role string (e.g. orchestrator or general).
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					Tokens:   domain.InvocationTokens{PeakContext: 30000},
				},
				{
					ID:       "ses-child-orch",
					ParentID: &rootID,
					Role:     "orchestrator",
					Tokens:   domain.InvocationTokens{PeakContext: 50000},
				},
				{
					ID:       "ses-child-gen",
					ParentID: &rootID,
					Role:     "general",
					Tokens:   domain.InvocationTokens{PeakContext: 40000},
				},
			}, nil, nil),
			}

			eff := domain.ComputeWorkflowEfficiency(ledger)
			// Root peak is 30000 from the ParentID == nil invocation
			if eff.ContextWindowHealth.RootPeakTokens != 30000 {
				t.Errorf("expected RootPeakTokens 30000, got %d", eff.ContextWindowHealth.RootPeakTokens)
			}
			// Delegated child invocations are classified structurally as subagents
			if eff.ContextWindowHealth.SubagentsPeakTokens != 50000 {
				t.Errorf("expected SubagentsPeakTokens 50000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "orchestrator" {
				t.Errorf("expected SubagentsPeakRole 'orchestrator', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("child invocation with whitespace role normalizes to unassigned", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					Tokens:   domain.InvocationTokens{PeakContext: 25000},
				},
				{
					ID:       "ses-child-blank",
					ParentID: &rootID,
					Role:     "   ",
					CostUSD:  1.20,
					Turns:    6,
					Tokens:   domain.InvocationTokens{PeakContext: 45000},
				},
			}, nil, nil),
			}

			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.SubagentsPeakTokens != 45000 {
				t.Errorf("expected SubagentsPeakTokens 45000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != domain.RoleUnassignedSubAgent {
				t.Errorf("expected SubagentsPeakRole %q, got %s", domain.RoleUnassignedSubAgent, eff.ContextWindowHealth.SubagentsPeakRole)
			}
			if eff.RoleEffortDistributionPct[string(domain.RoleUnassignedSubAgent)] != 100.0 {
				t.Errorf("expected unassigned effort 100.0%%, got %f", eff.RoleEffortDistributionPct[string(domain.RoleUnassignedSubAgent)])
			}
			if eff.SubagentTurnVelocity.ByRole[domain.RoleUnassignedSubAgent] != 6.0 {
				t.Errorf("expected unassigned turn velocity 6.0, got %f", eff.SubagentTurnVelocity.ByRole[domain.RoleUnassignedSubAgent])
			}
		})

		t.Run("single-invocation ticket with custom persona produces zero subagent metrics", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       "ses-single-custom",
					ParentID: nil,
					Role:     "custom-researcher",
					CostUSD:  3.50,
					Turns:    10,
					Tokens:   domain.InvocationTokens{PeakContext: 68000},
				},
			}, nil, nil),
			}

			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.RootPeakTokens != 68000 {
				t.Errorf("expected RootPeakTokens 68000, got %d", eff.ContextWindowHealth.RootPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
				t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
				t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
			if len(eff.RoleEffortDistributionPct) != 0 {
				t.Errorf("expected empty RoleEffortDistributionPct, got %+v", eff.RoleEffortDistributionPct)
			}
			if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
				t.Errorf("expected AvgTurnsPerInvocation 0.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
			}
		})

		t.Run("multi-tier child invocations all treated as subagents", func(t *testing.T) {
			t.Parallel()
			tier1ID := "ses-tier1"
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					CostUSD:  1.0,
					Tokens:   domain.InvocationTokens{PeakContext: 20000},
				},
				{
					ID:       tier1ID,
					ParentID: &rootID,
					Role:     "coordinator-subagent",
					CostUSD:  2.0,
					Turns:    5,
					Tokens:   domain.InvocationTokens{PeakContext: 35000},
				},
				{
					ID:       "ses-tier2",
					ParentID: &tier1ID,
					Role:     "specialist-subagent",
					CostUSD:  2.0,
					Turns:    7,
					Tokens:   domain.InvocationTokens{PeakContext: 42000},
				},
			}, nil, nil),
			}

			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.RootPeakTokens != 20000 {
				t.Errorf("expected RootPeakTokens 20000, got %d", eff.ContextWindowHealth.RootPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 42000 {
				t.Errorf("expected SubagentsPeakTokens 42000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "specialist-subagent" {
				t.Errorf("expected SubagentsPeakRole 'specialist-subagent', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
			if eff.RoleEffortDistributionPct["coordinator-subagent"] != 50.0 {
				t.Errorf("expected coordinator-subagent 50.0%%, got %f", eff.RoleEffortDistributionPct["coordinator-subagent"])
			}
			if eff.RoleEffortDistributionPct["specialist-subagent"] != 50.0 {
				t.Errorf("expected specialist-subagent 50.0%%, got %f", eff.RoleEffortDistributionPct["specialist-subagent"])
			}
			if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 6.0 {
				t.Errorf("expected AvgTurnsPerInvocation 6.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
			}
		})
	})

	t.Run("Case I: Multiple subagents tie-break canonical and custom roles", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root-i"

		t.Run("canonical tie-break tester vs reviewer", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					Tokens:   domain.InvocationTokens{PeakContext: 20000},
				},
				{
					ID:       "ses-reviewer",
					ParentID: &rootID,
					Role:     "reviewer",
					CostUSD:  1.0,
					Tokens:   domain.InvocationTokens{PeakContext: 80000},
				},
				{
					ID:       "ses-tester",
					ParentID: &rootID,
					Role:     "tester",
					CostUSD:  1.0,
					Tokens:   domain.InvocationTokens{PeakContext: 80000},
				},
			}, nil, nil),
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			// reviewer comes before tester in default alphabetical order
			if eff.ContextWindowHealth.SubagentsPeakRole != "reviewer" {
				t.Errorf("expected SubagentsPeakRole 'reviewer' via default alphabetical tie-break, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
			// When custom role order is supplied, tester comes before reviewer
			effCustom := domain.ComputeWorkflowEfficiency(ledger, []string{"tester", "reviewer"})
			if effCustom.ContextWindowHealth.SubagentsPeakRole != "tester" {
				t.Errorf("expected SubagentsPeakRole 'tester' via custom tie-break, got %s", effCustom.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("custom tie-break alphabetical order", func(t *testing.T) {
			t.Parallel()
			ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:       rootID,
					ParentID: nil,
					Role:     "general",
					Tokens:   domain.InvocationTokens{PeakContext: 20000},
				},
				{
					ID:       "ses-zebra",
					ParentID: &rootID,
					Role:     "zebra-agent",
					CostUSD:  1.0,
					Tokens:   domain.InvocationTokens{PeakContext: 75000},
				},
				{
					ID:       "ses-alpha",
					ParentID: &rootID,
					Role:     "alpha-agent",
					CostUSD:  1.0,
					Tokens:   domain.InvocationTokens{PeakContext: 75000},
				},
			}, nil, nil),
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			// alpha-agent comes before zebra-agent alphabetically
			if eff.ContextWindowHealth.SubagentsPeakRole != "alpha-agent" {
				t.Errorf("expected SubagentsPeakRole 'alpha-agent', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})
	})
}

func TestComputeWorkflowEfficiency_AdversarialStructuralHierarchy(t *testing.T) {
	t.Parallel()

	t.Run("inverted persona roles: root as coder and child as orchestrator", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-inverted-root"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       rootID,
				ParentID: nil,
				Role:     "coder",
				CostUSD:  10.0,
				Turns:    15,
				Tokens:   domain.InvocationTokens{PeakContext: 120000},
			},
			{
				ID:       "ses-inverted-child",
				ParentID: &rootID,
				Role:     "orchestrator",
				CostUSD:  5.0,
				Turns:    8,
				Tokens:   domain.InvocationTokens{PeakContext: 45000},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)

		// Root coordination is strictly ParentID == nil (even though role is coder)
		if eff.ContextWindowHealth.RootPeakTokens != 120000 {
			t.Errorf("expected RootPeakTokens 120000 from root coder, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}

		// Subagents are strictly ParentID != nil (even though role is orchestrator)
		if eff.ContextWindowHealth.SubagentsPeakTokens != 45000 {
			t.Errorf("expected SubagentsPeakTokens 45000 from child orchestrator, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "orchestrator" {
			t.Errorf("expected SubagentsPeakRole 'orchestrator', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}

		// Subagent effort distribution contains child orchestrator, never root coder
		if eff.RoleEffortDistributionPct["orchestrator"] != 100.0 {
			t.Errorf("expected orchestrator effort 100.0%%, got %f", eff.RoleEffortDistributionPct["orchestrator"])
		}
		if _, hasCoder := eff.RoleEffortDistributionPct["coder"]; hasCoder {
			t.Errorf("root coder must not appear in RoleEffortDistributionPct")
		}

		// Subagent turn velocity aggregates child orchestrator turns only
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 8.0 {
			t.Errorf("expected AvgTurnsPerInvocation 8.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 8 {
			t.Errorf("expected PeakTurns 8, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
		if eff.SubagentTurnVelocity.ByRole["orchestrator"] != 8.0 {
			t.Errorf("expected ByRole[orchestrator] 8.0, got %f", eff.SubagentTurnVelocity.ByRole["orchestrator"])
		}
		if _, hasCoder := eff.SubagentTurnVelocity.ByRole["coder"]; hasCoder {
			t.Errorf("root coder must not appear in SubagentTurnVelocity.ByRole")
		}
	})

	t.Run("multiple root sessions with zero children", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       "ses-root-1",
				ParentID: nil,
				Role:     "general",
				CostUSD:  2.0,
				Turns:    10,
				Tokens:   domain.InvocationTokens{PeakContext: 50000},
			},
			{
				ID:       "ses-root-2",
				ParentID: nil,
				Role:     "build",
				CostUSD:  3.0,
				Turns:    15,
				Tokens:   domain.InvocationTokens{PeakContext: 110000},
			},
			{
				ID:       "ses-root-3",
				ParentID: nil,
				Role:     "orchestrator",
				CostUSD:  1.0,
				Turns:    5,
				Tokens:   domain.InvocationTokens{PeakContext: 80000},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)

		// Max root peak is 110000
		if eff.ContextWindowHealth.RootPeakTokens != 110000 {
			t.Errorf("expected RootPeakTokens 110000, got %d", eff.ContextWindowHealth.RootPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected SubagentsPeakRole 'none', got %s", eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if len(eff.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct, got %+v", eff.RoleEffortDistributionPct)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected AvgTurnsPerInvocation 0.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
		if eff.SubagentTurnVelocity.PeakTurns != 0 {
			t.Errorf("expected PeakTurns 0, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
		if len(eff.SubagentTurnVelocity.ByRole) != 0 {
			t.Errorf("expected empty SubagentTurnVelocity.ByRole, got %+v", eff.SubagentTurnVelocity.ByRole)
		}
	})

	t.Run("empty string and whitespace-only roles on multiple child invocations normalize to unassigned", func(t *testing.T) {
		t.Parallel()
		rootID := "ses-root"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:       rootID,
				ParentID: nil,
				Role:     "general",
				Tokens:   domain.InvocationTokens{PeakContext: 20000},
			},
			{
				ID:       "ses-child-empty",
				ParentID: &rootID,
				Role:     "",
				CostUSD:  2.0,
				Turns:    4,
				Tokens:   domain.InvocationTokens{PeakContext: 60000},
			},
			{
				ID:       "ses-child-spaces",
				ParentID: &rootID,
				Role:     "   \t\n  ",
				CostUSD:  3.0,
				Turns:    6,
				Tokens:   domain.InvocationTokens{PeakContext: 70000},
			},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)

		// Both children are combined under role "UnassignedSubAgent"
		if eff.ContextWindowHealth.SubagentsPeakTokens != 70000 {
			t.Errorf("expected SubagentsPeakTokens 70000, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != domain.RoleUnassignedSubAgent {
			t.Errorf("expected SubagentsPeakRole %q, got %s", domain.RoleUnassignedSubAgent, eff.ContextWindowHealth.SubagentsPeakRole)
		}
		if eff.RoleEffortDistributionPct[string(domain.RoleUnassignedSubAgent)] != 100.0 {
			t.Errorf("expected unassigned effort 100.0%%, got %f", eff.RoleEffortDistributionPct[string(domain.RoleUnassignedSubAgent)])
		}
		if eff.SubagentTurnVelocity.ByRole[domain.RoleUnassignedSubAgent] != 5.0 {
			t.Errorf("expected unassigned turn velocity 5.0, got %f", eff.SubagentTurnVelocity.ByRole[domain.RoleUnassignedSubAgent])
		}
		if eff.SubagentTurnVelocity.PeakTurns != 6 {
			t.Errorf("expected PeakTurns 6, got %d", eff.SubagentTurnVelocity.PeakTurns)
		}
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 5.0 {
			t.Errorf("expected AvgTurnsPerInvocation 5.0, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})
}

func TestComputeWorkflowEfficiency_AdversarialTieBreaking(t *testing.T) {
	t.Parallel()

	rootID := "ses-root"

	t.Run("five-way tie across arbitrary unmapped roles alphabetical default and custom order", func(t *testing.T) {
		t.Parallel()

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-1", ParentID: &rootID, Role: "reviewer", Tokens: domain.InvocationTokens{PeakContext: 125000}},
			{ID: "inv-2", ParentID: &rootID, Role: "researcher", Tokens: domain.InvocationTokens{PeakContext: 125000}},
			{ID: "inv-3", ParentID: &rootID, Role: "developer", Tokens: domain.InvocationTokens{PeakContext: 125000}},
			{ID: "inv-4", ParentID: &rootID, Role: "architect", Tokens: domain.InvocationTokens{PeakContext: 125000}},
			{ID: "inv-5", ParentID: &rootID, Role: "lead-qa", Tokens: domain.InvocationTokens{PeakContext: 125000}},
		}, nil, nil),
		}

		// Alphabetical default: architect < developer < lead-qa < researcher < reviewer
		effAlpha := domain.ComputeWorkflowEfficiency(ledger)
		if effAlpha.ContextWindowHealth.SubagentsPeakTokens != 125000 {
			t.Errorf("expected SubagentsPeakTokens 125000, got %d", effAlpha.ContextWindowHealth.SubagentsPeakTokens)
		}
		if effAlpha.ContextWindowHealth.SubagentsPeakRole != "architect" {
			t.Errorf("expected 'architect' via alphabetical tie-breaking, got %q", effAlpha.ContextWindowHealth.SubagentsPeakRole)
		}

		// Empty custom order slice falls back to alphabetical
		effEmptyCustom := domain.ComputeWorkflowEfficiency(ledger, []string{})
		if effEmptyCustom.ContextWindowHealth.SubagentsPeakRole != "architect" {
			t.Errorf("expected 'architect' via empty custom order, got %q", effEmptyCustom.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing lead-qa
		effLeadQA := domain.ComputeWorkflowEfficiency(ledger, []string{"lead-qa", "developer"})
		if effLeadQA.ContextWindowHealth.SubagentsPeakRole != "lead-qa" {
			t.Errorf("expected 'lead-qa' via custom order, got %q", effLeadQA.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing researcher
		effResearcher := domain.ComputeWorkflowEfficiency(ledger, []string{"researcher", "architect"})
		if effResearcher.ContextWindowHealth.SubagentsPeakRole != "researcher" {
			t.Errorf("expected 'researcher' via custom order, got %q", effResearcher.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing reviewer
		effReviewer := domain.ComputeWorkflowEfficiency(ledger, []string{"reviewer"})
		if effReviewer.ContextWindowHealth.SubagentsPeakRole != "reviewer" {
			t.Errorf("expected 'reviewer' via custom order, got %q", effReviewer.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order with unlisted/nonexistent roles first, then matching role
		effUnlisted := domain.ComputeWorkflowEfficiency(ledger, []string{"ghost-agent", "nonexistent", "developer"})
		if effUnlisted.ContextWindowHealth.SubagentsPeakRole != "developer" {
			t.Errorf("expected 'developer' via custom order with unlisted prefix, got %q", effUnlisted.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order with ONLY nonexistent roles falls back to full alphabetical
		effGhostOnly := domain.ComputeWorkflowEfficiency(ledger, []string{"ghost-1", "ghost-2"})
		if effGhostOnly.ContextWindowHealth.SubagentsPeakRole != "architect" {
			t.Errorf("expected 'architect' fallback when custom order has no matches, got %q", effGhostOnly.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("multiple invocations per role with peak tie", func(t *testing.T) {
		t.Parallel()

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-coder-1", ParentID: &rootID, Role: "coder", Tokens: domain.InvocationTokens{PeakContext: 40000}},
			{ID: "inv-coder-2", ParentID: &rootID, Role: "coder", Tokens: domain.InvocationTokens{PeakContext: 90000}},
			{ID: "inv-tester-1", ParentID: &rootID, Role: "tester", Tokens: domain.InvocationTokens{PeakContext: 90000}},
			{ID: "inv-tester-2", ParentID: &rootID, Role: "tester", Tokens: domain.InvocationTokens{PeakContext: 60000}},
			{ID: "inv-planner-1", ParentID: &rootID, Role: "planner", Tokens: domain.InvocationTokens{PeakContext: 70000}},
		}, nil, nil),
		}

		// Alphabetical default: coder < tester
		effAlpha := domain.ComputeWorkflowEfficiency(ledger)
		if effAlpha.ContextWindowHealth.SubagentsPeakTokens != 90000 {
			t.Errorf("expected SubagentsPeakTokens 90000, got %d", effAlpha.ContextWindowHealth.SubagentsPeakTokens)
		}
		if effAlpha.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected 'coder' via default tie-break, got %q", effAlpha.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing tester over coder
		effTester := domain.ComputeWorkflowEfficiency(ledger, []string{"tester", "coder"})
		if effTester.ContextWindowHealth.SubagentsPeakRole != "tester" {
			t.Errorf("expected 'tester' via custom order, got %q", effTester.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("unassigned role participates in tie-breaking properly", func(t *testing.T) {
		t.Parallel()

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-empty", ParentID: &rootID, Role: "   ", Tokens: domain.InvocationTokens{PeakContext: 85000}},
			{ID: "inv-advisor", ParentID: &rootID, Role: "advisor", Tokens: domain.InvocationTokens{PeakContext: 85000}},
		}, nil, nil),
		}

		// Alphabetical default: UnassignedSubAgent < advisor ('U' < 'a')
		effAlpha := domain.ComputeWorkflowEfficiency(ledger)
		if effAlpha.ContextWindowHealth.SubagentsPeakTokens != 85000 {
			t.Errorf("expected SubagentsPeakTokens 85000, got %d", effAlpha.ContextWindowHealth.SubagentsPeakTokens)
		}
		if effAlpha.ContextWindowHealth.SubagentsPeakRole != domain.RoleUnassignedSubAgent {
			t.Errorf("expected %q via alphabetical tie-break, got %q", domain.RoleUnassignedSubAgent, effAlpha.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing advisor
		effAdvisor := domain.ComputeWorkflowEfficiency(ledger, []string{"advisor", string(domain.RoleUnassignedSubAgent)})
		if effAdvisor.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected 'advisor' via custom tie-break, got %q", effAdvisor.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("structural hierarchy isolation during tie-breaking", func(t *testing.T) {
		t.Parallel()

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-root-orch", ParentID: nil, Role: "orchestrator", Tokens: domain.InvocationTokens{PeakContext: 200000}},
			{ID: "inv-root-gen", ParentID: nil, Role: "general", Tokens: domain.InvocationTokens{PeakContext: 200000}},
			{ID: "inv-sub-dev", ParentID: &rootID, Role: "developer", Tokens: domain.InvocationTokens{PeakContext: 95000}},
			{ID: "inv-sub-arch", ParentID: &rootID, Role: "architect", Tokens: domain.InvocationTokens{PeakContext: 95000}},
		}, nil, nil),
		}

		effAlpha := domain.ComputeWorkflowEfficiency(ledger)
		if effAlpha.ContextWindowHealth.RootPeakTokens != 200000 {
			t.Errorf("expected RootPeakTokens 200000, got %d", effAlpha.ContextWindowHealth.RootPeakTokens)
		}
		if effAlpha.ContextWindowHealth.SubagentsPeakTokens != 95000 {
			t.Errorf("expected SubagentsPeakTokens 95000, got %d", effAlpha.ContextWindowHealth.SubagentsPeakTokens)
		}
		// Alphabetical tie-break between architect and developer
		if effAlpha.ContextWindowHealth.SubagentsPeakRole != "architect" {
			t.Errorf("expected SubagentsPeakRole 'architect', got %q", effAlpha.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing developer
		effDev := domain.ComputeWorkflowEfficiency(ledger, []string{"developer", "architect"})
		if effDev.ContextWindowHealth.SubagentsPeakRole != "developer" {
			t.Errorf("expected SubagentsPeakRole 'developer', got %q", effDev.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom order prioritizing root personas does not cause root to become SubagentsPeakRole
		effRootOrder := domain.ComputeWorkflowEfficiency(ledger, []string{"orchestrator", "general", "developer", "architect"})
		if effRootOrder.ContextWindowHealth.SubagentsPeakRole != "developer" {
			t.Errorf("expected SubagentsPeakRole 'developer' (excluding root personas), got %q", effRootOrder.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("pure root sessions always evaluate to none regardless of custom order", func(t *testing.T) {
		t.Parallel()

		ledger := domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-root-1", ParentID: nil, Role: "general", Tokens: domain.InvocationTokens{PeakContext: 110000}},
			{ID: "inv-root-2", ParentID: nil, Role: "orchestrator", Tokens: domain.InvocationTokens{PeakContext: 130000}},
		}, nil, nil),
		}

		eff := domain.ComputeWorkflowEfficiency(ledger, []string{"orchestrator", "planner", "coder", "general"})
		if eff.ContextWindowHealth.SubagentsPeakTokens != 0 {
			t.Errorf("expected SubagentsPeakTokens 0, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
		}
		if eff.ContextWindowHealth.SubagentsPeakRole != "none" {
			t.Errorf("expected SubagentsPeakRole 'none', got %q", eff.ContextWindowHealth.SubagentsPeakRole)
		}
	})
}

func TestAL34_NetLinesGrowthAndGranularityAdversarial(t *testing.T) {
	t.Parallel()

	t.Run("CumulativeNetLinesAdded accurately tracks running additions minus running deletions across buckets", func(t *testing.T) {
		t.Parallel()

		d1 := "2026-10-10T10:00:00Z"
		d2 := "2026-10-11T10:00:00Z"
		d3 := "2026-10-12T10:00:00Z"
		d4 := "2026-10-13T10:00:00Z"

		ledgers := []domain.AgenticMetricsLedger{
			{
				TicketID:    ptr("T-D1"),
				CompletedAt: &d1,
				WorkflowEfficiency: domain.WorkflowEfficiency{
					CodeDelivered: domain.CodeDeliveredStats{
						LinesAdded:   1000,
						LinesDeleted: 200,
					},
				},
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{CompletedAt: &d1, Tokens: domain.InvocationTokens{Input: 100, CacheRead: 50}},
				}, nil, nil),
				},
			},
			{
				TicketID:    ptr("T-D2"),
				CompletedAt: &d2,
				WorkflowEfficiency: domain.WorkflowEfficiency{
					CodeDelivered: domain.CodeDeliveredStats{
						LinesAdded:   100,
						LinesDeleted: 1200,
					},
				},
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{CompletedAt: &d2, Tokens: domain.InvocationTokens{Input: 100, CacheRead: 50}},
				}, nil, nil),
				},
			},
			{
				TicketID:    ptr("T-D3"),
				CompletedAt: &d3,
				WorkflowEfficiency: domain.WorkflowEfficiency{
					CodeDelivered: domain.CodeDeliveredStats{
						LinesAdded:   0,
						LinesDeleted: 0,
					},
				},
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{CompletedAt: &d3, Tokens: domain.InvocationTokens{Input: 100, CacheRead: 50}},
				}, nil, nil),
				},
			},
			{
				TicketID:    ptr("T-D4"),
				CompletedAt: &d4,
				WorkflowEfficiency: domain.WorkflowEfficiency{
					CodeDelivered: domain.CodeDeliveredStats{
						LinesAdded:   500,
						LinesDeleted: 50,
					},
				},
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{CompletedAt: &d4, Tokens: domain.InvocationTokens{Input: 100, CacheRead: 50}},
				}, nil, nil),
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(ledgers, "daily", nil, nil)
		if len(buckets) != 4 {
			t.Fatalf("expected 4 daily buckets, got %d", len(buckets))
		}

		// Expected trajectory:
		// Bucket 0 (2026-10-10): added=1000, deleted=200, net=800, cumAdded=1000, cumDeleted=200, cumNet=800
		// Bucket 1 (2026-10-11): added=100, deleted=1200, net=-1100, cumAdded=1100, cumDeleted=1400, cumNet=-300
		// Bucket 2 (2026-10-12): added=0, deleted=0, net=0, cumAdded=1100, cumDeleted=1400, cumNet=-300
		// Bucket 3 (2026-10-13): added=500, deleted=50, net=450, cumAdded=1600, cumDeleted=1450, cumNet=150
		expectedTrajectory := []struct {
			date       string
			added      int64
			deleted    int64
			net        int64
			cumAdded   int64
			cumDeleted int64
			cumNet     int64
		}{
			{"2026-10-10", 1000, 200, 800, 1000, 200, 800},
			{"2026-10-11", 100, 1200, -1100, 1100, 1400, -300},
			{"2026-10-12", 0, 0, 0, 1100, 1400, -300},
			{"2026-10-13", 500, 50, 450, 1600, 1450, 150},
		}

		for i, exp := range expectedTrajectory {
			b := buckets[i]
			if b.BucketDate != exp.date {
				t.Errorf("[%d] expected date %s, got %s", i, exp.date, b.BucketDate)
			}
			if b.LinesAdded != exp.added {
				t.Errorf("[%d] expected linesAdded %d, got %d", i, exp.added, b.LinesAdded)
			}
			if b.LinesDeleted != exp.deleted {
				t.Errorf("[%d] expected linesDeleted %d, got %d", i, exp.deleted, b.LinesDeleted)
			}
			if b.NetLinesAdded != exp.net {
				t.Errorf("[%d] expected netLinesAdded %d, got %d", i, exp.net, b.NetLinesAdded)
			}
			if b.CumulativeLinesAdded != exp.cumAdded {
				t.Errorf("[%d] expected cumLinesAdded %d, got %d", i, exp.cumAdded, b.CumulativeLinesAdded)
			}
			if b.CumulativeLinesDeleted != exp.cumDeleted {
				t.Errorf("[%d] expected cumLinesDeleted %d, got %d", i, exp.cumDeleted, b.CumulativeLinesDeleted)
			}
			if b.CumulativeNetLinesAdded != exp.cumNet {
				t.Errorf("[%d] expected cumNetLinesAdded %d, got %d", i, exp.cumNet, b.CumulativeNetLinesAdded)
			}
			// Invariant assertion: CumulativeNetLinesAdded == CumulativeLinesAdded - CumulativeLinesDeleted
			if b.CumulativeNetLinesAdded != b.CumulativeLinesAdded-b.CumulativeLinesDeleted {
				t.Errorf("[%d] CumulativeNetLinesAdded (%d) does not match CumulativeLinesAdded - CumulativeLinesDeleted (%d - %d)",
					i, b.CumulativeNetLinesAdded, b.CumulativeLinesAdded, b.CumulativeLinesDeleted)
			}
		}
	})

	t.Run("ParseTimeGranularity rejects cumulative and arbitrary non-temporal strings with case and whitespace variations", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			input    string
			valid    bool
			expected domain.TimeGranularity
		}{
			{"cumulative", false, ""},
			{"CUMULATIVE", false, ""},
			{"Cumulative", false, ""},
			{"cUmUlAtIvE", false, ""},
			{"  cumulative  ", false, ""},
			{"\tcumulative\n", false, ""},
			{"hourly", true, domain.GranularityHourly},
			{"HOURLY", true, domain.GranularityHourly},
			{"Hourly", true, domain.GranularityHourly},
			{"  hourly  ", true, domain.GranularityHourly},
			{"daily", true, domain.GranularityDaily},
			{"DAILY", true, domain.GranularityDaily},
			{"Daily", true, domain.GranularityDaily},
			{"  daily  ", true, domain.GranularityDaily},
			{"weekly", true, domain.GranularityWeekly},
			{"WEEKLY", true, domain.GranularityWeekly},
			{"Weekly", true, domain.GranularityWeekly},
			{"  weekly  ", true, domain.GranularityWeekly},
			{"minutely", false, ""},
			{"secondly", false, ""},
			{"yearly", false, ""},
			{"monthly", false, ""},
			{"decade", false, ""},
			{"all", false, ""},
			{"none", false, ""},
			{"1d", false, ""},
			{"7d", false, ""},
			{"custom", false, ""},
			{"", false, ""},
			{"   ", false, ""},
			{"../../etc/passwd", false, ""},
			{"!@#$%^&*", false, ""},
		}

		for _, tt := range tests {
			got, ok := domain.ParseTimeGranularity(tt.input)
			if ok != tt.valid {
				t.Errorf("ParseTimeGranularity(%q) valid = %v, want %v", tt.input, ok, tt.valid)
			}
			if got != tt.expected {
				t.Errorf("ParseTimeGranularity(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("AggregateMetricsOverTime edge cases: nil, empty, zero tokens, and division by zero", func(t *testing.T) {
		t.Parallel()

		// Nil ledgers
		resNil := domain.AggregateMetricsOverTime(nil, "daily", nil, nil)
		if len(resNil) != 0 {
			t.Errorf("expected empty slice for nil ledgers, got %d", len(resNil))
		}

		// Empty ledgers
		resEmpty := domain.AggregateMetricsOverTime([]domain.AgenticMetricsLedger{}, "daily", nil, nil)
		if len(resEmpty) != 0 {
			t.Errorf("expected empty slice for empty ledgers, got %d", len(resEmpty))
		}

		// Zero tokens and zero cache read (division by zero guard in CacheHitRatePct)
		d := "2026-10-15T12:00:00Z"
		zeroLedger := []domain.AgenticMetricsLedger{
			{
				TicketID:    ptr("ZERO-1"),
				CompletedAt: &d,
				Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
					{
						CompletedAt: &d,
						Tokens: domain.InvocationTokens{
							Input:     0,
							Output:    0,
							CacheRead: 0,
							Reasoning: 0,
						},
					},
				}, nil, nil),
				},
				WorkflowEfficiency: domain.WorkflowEfficiency{
					CodeDelivered: domain.CodeDeliveredStats{
						LinesAdded:   0,
						LinesDeleted: 0,
					},
				},
			},
		}

		buckets := domain.AggregateMetricsOverTime(zeroLedger, "daily", nil, nil)
		if len(buckets) != 1 {
			t.Fatalf("expected 1 bucket for zero tokens ledger, got %d", len(buckets))
		}
		b := buckets[0]
		if b.CacheHitRatePct != 0.0 {
			t.Errorf("expected CacheHitRatePct 0.0, got %f", b.CacheHitRatePct)
		}
		if b.CumulativeNetLinesAdded != 0 {
			t.Errorf("expected CumulativeNetLinesAdded 0, got %d", b.CumulativeNetLinesAdded)
		}
		if b.NetLinesAdded != 0 {
			t.Errorf("expected NetLinesAdded 0, got %d", b.NetLinesAdded)
		}

		// Inverted date range
		start := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
		bucketsInverted := domain.AggregateMetricsOverTime(zeroLedger, "daily", &start, &end)
		if len(bucketsInverted) != 0 {
			t.Errorf("expected 0 buckets for inverted date range, got %d", len(bucketsInverted))
		}
	})
}

func TestConfigurableHealthyCeilingAggregations(t *testing.T) {
	t.Parallel()

	customCeiling := int64(100_000)
	invocations := []domain.InvocationFact{
		{
			ID:    "inv-1",
			Role:  "coder",
			Model: "claude-3-5-sonnet",
			Tokens: domain.InvocationTokens{
				PeakContext: 75_000,
			},
		},
	}
	ledger := domain.AgenticFactLedger{
		Sessions: sessionFacts(invocations, nil, nil),
	}

	t.Run("ComputeTotals respects custom ceiling", func(t *testing.T) {
		totalsDefault := domain.ComputeTotals(invocations)
		// 75k under 200k default ceiling is HEALTHY (< 100k)
		if totalsDefault.ContextHealth != domain.ContextHealthHealthy {
			t.Errorf("expected HEALTHY under default ceiling, got %s", totalsDefault.ContextHealth)
		}

		totalsCustom := domain.ComputeTotals(invocations, customCeiling)
		// 75k under 100k custom ceiling is ELEVATED (>= 50k and <= 100k)
		if totalsCustom.ContextHealth != domain.ContextHealthElevated {
			t.Errorf("expected ELEVATED under custom ceiling, got %s", totalsCustom.ContextHealth)
		}
	})

	t.Run("ComputeWorkflowEfficiencyWithCeiling respects custom ceiling", func(t *testing.T) {
		effCustom := domain.ComputeWorkflowEfficiencyWithCeiling(ledger, customCeiling)
		expectedSat := math.Round(75000.0/float64(customCeiling)*1000.0) / 10.0
		if effCustom.ContextWindowHealth.HealthyCeilingTokens != customCeiling {
			t.Errorf("expected HealthyCeilingTokens %d, got %d", customCeiling, effCustom.ContextWindowHealth.HealthyCeilingTokens)
		}
		if effCustom.ContextWindowHealth.RootSaturationPct != expectedSat {
			t.Errorf("expected RootSaturationPct %f, got %f", expectedSat, effCustom.ContextWindowHealth.RootSaturationPct)
		}
	})

	t.Run("ComputeWorkflowEfficiencyWithConfig respects cfg.AnomalyThresholds.ContextCeilingCriticalTokens", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig()
		cfg.AnomalyThresholds.ContextCeilingCriticalTokens = customCeiling

		eff := domain.ComputeWorkflowEfficiencyWithConfig(ledger, cfg)
		if eff.ContextWindowHealth.HealthyCeilingTokens != customCeiling {
			t.Errorf("expected HealthyCeilingTokens %d, got %d", customCeiling, eff.ContextWindowHealth.HealthyCeilingTokens)
		}
	})

	t.Run("DeriveMetricsLedger propagates cfg.AnomalyThresholds.ContextCeilingCriticalTokens", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig()
		cfg.AnomalyThresholds.ContextCeilingCriticalTokens = customCeiling

		derived := domain.DeriveMetricsLedger(ledger, cfg)
		if derived.Totals.ContextHealth != domain.ContextHealthElevated {
			t.Errorf("expected Totals.ContextHealth ELEVATED, got %s", derived.Totals.ContextHealth)
		}
		if derived.WorkflowEfficiency.ContextWindowHealth.HealthyCeilingTokens != customCeiling {
			t.Errorf("expected WorkflowEfficiency.ContextWindowHealth.HealthyCeilingTokens %d, got %d",
				customCeiling, derived.WorkflowEfficiency.ContextWindowHealth.HealthyCeilingTokens)
		}
	})
}
