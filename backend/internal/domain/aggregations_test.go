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
		if res.ContextHealth != string(domain.ContextHealthHealthy) {
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
		if res.ContextHealth != string(domain.ContextHealthElevated) {
			t.Errorf("expected ELEVATED health status, got %s", res.ContextHealth)
		}
	})
}

func TestComputeWorkflowEfficiency(t *testing.T) {
	t.Parallel()
	t.Run("zero division defenses", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{
			Code:        domain.CodeFacts{},
			Invocations: []domain.InvocationFact{},
			Tools:       []domain.ToolFact{},
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
				LinesAdded:      150,
				LinesDeleted:    30,
				TestLinesAdded:  80,
				TestLinesDeleted: 10,
				ProdLinesAdded:  40,
				ProdLinesDeleted: 20,
				FilesChanged:    5,
				SDLCIterations:  2,
			},
			Invocations: []domain.InvocationFact{
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
			Tools: []domain.ToolFact{
				{Name: "read", Calls: 10, Errors: 1},
			},
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
				Invocations:      2,
				CostUSD:          1.5,
				InputTokens:      5000,
				OutputTokens:     1000,
				CacheReadTokens:  5000,
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

func ptr(s string) *string {
	return &s
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
			TicketID:  ptr("T-7D"),
			UpdatedAt: dateIn7D,
		},
		{
			TicketID:  ptr("T-30D"),
			UpdatedAt: dateIn30D,
		},
		{
			TicketID:  ptr("T-90D"),
			UpdatedAt: dateIn90D,
		},
		{
			TicketID:  ptr("T-OLD"),
			UpdatedAt: dateOld,
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

	t.Run("empty input returns empty slice", func(t *testing.T) {
		res := domain.FilterLedgersByTimeWindow(nil, domain.TimeWindow7D, ref)
		if len(res) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(res))
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

	ledgers := []domain.AgenticMetricsLedger{
		{
			TicketID:    ptr("AL-1"),
			StartedAt:   &time1,
			CompletedAt: &time2,
			UpdatedAt:   time2,
			Totals: domain.GroupedMetric{
				CostUSD: 2.0,
			},
			WorkflowEfficiency: domain.WorkflowEfficiency{
				ActiveDurationFormatted: "10m 0s",
				ContextWindowHealth: domain.ContextWindowHealth{
					Status: "HEALTHY",
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
				if totals.ContextHealth != string(tc.expectedHealth) {
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
		if totals.ContextHealth != string(domain.ContextHealthHealthy) {
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
			{TicketID: ptr("T-START-EXACT"), UpdatedAt: tsStartExact},
			{TicketID: ptr("T-START-BEFORE"), UpdatedAt: tsStartBefore},
			{TicketID: ptr("T-END-EXACT"), UpdatedAt: tsEndExact},
			{TicketID: ptr("T-END-FUTURE"), UpdatedAt: tsEndFuture},
			{TicketID: ptr("T-INVALID"), UpdatedAt: tsInvalid},
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
				Code:        domain.CodeFacts{},
				Invocations: nil,
				Tools:       nil,
			}
			eff := domain.ComputeWorkflowEfficiency(emptyLedger)
			tot := domain.ComputeTotals(emptyLedger.Invocations)

			if tot.Invocations != 0 {
				t.Errorf("expected 0 invocations, got %d", tot.Invocations)
			}
			if tot.CostUSD != 0.0 {
				t.Errorf("expected 0 cost, got %f", tot.CostUSD)
			}
			if tot.CacheHitRatePct != 0.0 {
				t.Errorf("expected 0 cache hit rate, got %f", tot.CacheHitRatePct)
			}
			if tot.ContextHealth != string(domain.ContextHealthHealthy) {
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
			if eff.ContextWindowHealth.Status != string(domain.ContextHealthHealthy) {
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
			ledger := domain.AgenticFactLedger{
				Invocations: []domain.InvocationFact{
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
				},
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.Status != string(domain.ContextHealthBloated) {
				t.Errorf("expected BLOATED, got %s", eff.ContextWindowHealth.Status)
			}
			if eff.ContextWindowHealth.OrchestratorSaturationPct != 125.0 {
				t.Errorf("expected 125.0 orchestrator saturation, got %f", eff.ContextWindowHealth.OrchestratorSaturationPct)
			}
			if eff.ContextWindowHealth.SubagentsPeakTokens != 210000 {
				t.Errorf("expected 210000 subagent peak, got %d", eff.ContextWindowHealth.SubagentsPeakTokens)
			}
			if eff.ContextWindowHealth.SubagentsPeakRole != "coder" {
				t.Errorf("expected coder subagent peak role, got %s", eff.ContextWindowHealth.SubagentsPeakRole)
			}
		})

		t.Run("critical token saturation exceeding 400k ceiling", func(t *testing.T) {
			ledger := domain.AgenticFactLedger{
				Invocations: []domain.InvocationFact{
					{
						ID:   "ses-0",
						Role: "orchestrator",
						Tokens: domain.InvocationTokens{
							PeakContext: 450000,
						},
					},
				},
			}
			eff := domain.ComputeWorkflowEfficiency(ledger)
			if eff.ContextWindowHealth.Status != string(domain.ContextHealthCritical) {
				t.Errorf("expected CRITICAL, got %s", eff.ContextWindowHealth.Status)
			}
		})
	})
}

func TestDeriveMetricsLedger_EventsPreservationAndInvariants(t *testing.T) {
	t.Parallel()
	tid := "AL-6"
	title := "Chronological Event Extraction"
	ttype := "feature"
	baseLedger := domain.AgenticFactLedger{
		TicketID:    &tid,
		TicketTitle: &title,
		TicketType:  &ttype,
		Code: domain.CodeFacts{
			LinesAdded:   100,
			LinesDeleted: 20,
			FilesChanged: 4,
		},
		Invocations: []domain.InvocationFact{
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
		Tools: []domain.ToolFact{
			{Name: "bash", Calls: 5, Errors: 1},
		},
	}

	t.Run("nil events preserved in Facts and does not mutate mathematical derivations", func(t *testing.T) {
		ledgerNilEvents := baseLedger
		ledgerNilEvents.Events = nil

		metrics := domain.DeriveMetricsLedger(ledgerNilEvents)
		if metrics.Facts.Events != nil {
			t.Errorf("expected nil Facts.Events, got %+v", metrics.Facts.Events)
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
		ledgerEmptyEvents.Events = []domain.EventFact{}

		metrics := domain.DeriveMetricsLedger(ledgerEmptyEvents)
		if metrics.Facts.Events == nil {
			t.Fatal("expected non-nil empty Events slice")
		}
		if len(metrics.Facts.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(metrics.Facts.Events))
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
		ledgerWithEvents.Events = events

		metricsWithEvents := domain.DeriveMetricsLedger(ledgerWithEvents)
		metricsWithoutEvents := domain.DeriveMetricsLedger(baseLedger)

		if len(metricsWithEvents.Facts.Events) != 3 {
			t.Fatalf("expected 3 events in Facts, got %d", len(metricsWithEvents.Facts.Events))
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
		ev0 := metricsWithEvents.Facts.Events[0]
		if ev0.EventType != domain.EventTypeTurnMessage || ev0.Turn != 1 {
			t.Errorf("unexpected event 0: %+v", ev0)
		}
		ev1 := metricsWithEvents.Facts.Events[1]
		if ev1.EventType != domain.EventTypeToolCall || ev1.Status != "error" || *ev1.ErrorMessage != errBash || *ev1.DurationMs != durMs {
			t.Errorf("unexpected event 1: %+v", ev1)
		}
		ev2 := metricsWithEvents.Facts.Events[2]
		if ev2.EventType != domain.EventTypeCompaction || ev2.Turn != 2 {
			t.Errorf("unexpected event 2: %+v", ev2)
		}
	})

	t.Run("zero division on completely empty fact ledger", func(t *testing.T) {
		emptyLedger := domain.AgenticFactLedger{}
		metrics := domain.DeriveMetricsLedger(emptyLedger)

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
		if len(metrics.Facts.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(metrics.Facts.Events))
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

func TestIsSubagentRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{name: "empty role", role: "", expected: false},
		{name: "whitespace role", role: "   ", expected: false},
		{name: "orchestrator lowercase", role: "orchestrator", expected: false},
		{name: "orchestrator mixed case", role: "Orchestrator", expected: false},
		{name: "orchestrator with whitespace", role: "  orchestrator  ", expected: false},
		{name: "canonical subagent planner", role: "planner", expected: true},
		{name: "canonical subagent coder", role: "coder", expected: true},
		{name: "canonical subagent tester", role: "tester", expected: true},
		{name: "canonical subagent reviewer", role: "reviewer", expected: true},
		{name: "custom role advisor", role: "advisor", expected: true},
		{name: "custom role po", role: "po", expected: true},
		{name: "custom role researcher", role: "researcher", expected: true},
		{name: "custom role security-auditor", role: "security-auditor", expected: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := domain.IsSubagentRole(tc.role)
			if got != tc.expected {
				t.Errorf("IsSubagentRole(%q) = %v, expected %v", tc.role, got, tc.expected)
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
	expected := []string{"orchestrator", "coder", "tester", "advisor", "researcher"}

	got := domain.OrderRoles(input)
	if len(got) != len(expected) {
		t.Fatalf("expected len %d, got %d: %v", len(expected), len(got), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expected[i], got[i])
		}
	}
}

func TestOrderSubagentRoles(t *testing.T) {
	t.Parallel()
	input := []string{"advisor", "reviewer", "orchestrator", "po", "planner", "advisor"}
	expected := []string{"planner", "reviewer", "advisor", "po"}

	got := domain.OrderSubagentRoles(input)
	if len(got) != len(expected) {
		t.Fatalf("expected len %d, got %d: %v", len(expected), len(got), got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expected[i], got[i])
		}
	}
}

func TestWorkflowEfficiency_DynamicSubagents(t *testing.T) {
	t.Parallel()

	t.Run("dynamic subagent effort distribution", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-orch"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "ses-orch", Role: "orchestrator", CostUSD: 1.0},
				{ID: "ses-coder", ParentID: &parentID, Role: "coder", CostUSD: 2.0},
				{ID: "ses-adv", ParentID: &parentID, Role: "advisor", CostUSD: 1.0},
				{ID: "ses-po", ParentID: &parentID, Role: "po", CostUSD: 1.0},
			},
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
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "ses-adv", ParentID: &parentID, Role: "advisor", Turns: 4},
				{ID: "ses-po", ParentID: &parentID, Role: "po", Turns: 8},
				{ID: "ses-coder", ParentID: &parentID, Role: "coder", Turns: 6},
			},
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
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-orch",
					Role: "orchestrator",
					Tokens: domain.InvocationTokens{PeakContext: 50000},
				},
				{
					ID:   "ses-coder",
					Role: "coder",
					Tokens: domain.InvocationTokens{PeakContext: 120000},
				},
				{
					ID:   "ses-adv",
					Role: "advisor",
					Tokens: domain.InvocationTokens{PeakContext: 180000},
				},
			},
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
		// Canonical vs Custom tie-break: canonical (coder) takes precedence
		ledgerCanonicalVsCustom := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-coder",
					Role: "coder",
					Tokens: domain.InvocationTokens{PeakContext: 100000},
				},
				{
					ID:   "ses-adv",
					Role: "advisor",
					Tokens: domain.InvocationTokens{PeakContext: 100000},
				},
			},
		}

		eff1 := domain.ComputeWorkflowEfficiency(ledgerCanonicalVsCustom)
		if eff1.ContextWindowHealth.SubagentsPeakRole != "coder" {
			t.Errorf("expected canonical role 'coder' to win tie-breaker, got %s", eff1.ContextWindowHealth.SubagentsPeakRole)
		}

		// Custom vs Custom tie-break: alphabetical (advisor before po)
		ledgerCustomVsCustom := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:   "ses-po",
					Role: "po",
					Tokens: domain.InvocationTokens{PeakContext: 100000},
				},
				{
					ID:   "ses-adv",
					Role: "advisor",
					Tokens: domain.InvocationTokens{PeakContext: 100000},
				},
			},
		}

		eff2 := domain.ComputeWorkflowEfficiency(ledgerCustomVsCustom)
		if eff2.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected alphabetical role 'advisor' to win tie-breaker, got %s", eff2.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("orchestrator-only session", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{
					ID:      "ses-orch",
					Role:    "orchestrator",
					Turns:   10,
					CostUSD: 2.5,
					Tokens:  domain.InvocationTokens{PeakContext: 75000},
				},
			},
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

	metrics := domain.ComputeRoleMetrics(invs)
	if len(metrics) != 5 {
		t.Fatalf("expected 5 roles, got %d", len(metrics))
	}

	expectedRoles := []string{"orchestrator", "planner", "reviewer", "advisor", "architect"}
	for i, exp := range expectedRoles {
		if metrics[i].Role != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, metrics[i].Role)
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

	records := domain.ComputeThreadSummaries(invs)
	if len(records) != 1 {
		t.Fatalf("expected 1 thread record, got %d", len(records))
	}

	rec := records[0]
	expectedRoles := []string{"orchestrator", "coder", "advisor"}
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
		if len(domain.OrderRoles(nil)) != 0 {
			t.Errorf("expected empty slice for nil roles")
		}
		if len(domain.OrderRoles([]string{})) != 0 {
			t.Errorf("expected empty slice for empty roles")
		}
		if len(domain.OrderRoles([]string{"", "   ", "\t", "\n"})) != 0 {
			t.Errorf("expected empty slice for whitespace-only roles")
		}

		if len(domain.OrderSubagentRoles(nil)) != 0 {
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
		expected := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}

		got := domain.OrderRoles(input)
		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("at %d: expected %s, got %s", i, expected[i], got[i])
			}
		}

		// Subagents: orchestrator must be excluded
		subExpected := []string{"planner", "coder", "tester", "reviewer"}
		gotSub := domain.OrderSubagentRoles(input)
		if len(gotSub) != len(subExpected) {
			t.Fatalf("expected len %d, got %d: %v", len(subExpected), len(gotSub), gotSub)
		}
		for i := range subExpected {
			if gotSub[i] != subExpected[i] {
				t.Errorf("subagent at %d: expected %s, got %s", i, subExpected[i], gotSub[i])
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
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", Role: "orchestrator", CostUSD: 0.0},
				{ID: "2", ParentID: &parentID, Role: "advisor", CostUSD: 0.0, Turns: 5},
				{ID: "3", ParentID: &parentID, Role: "po", CostUSD: 0.0, Turns: 3},
			},
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
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", ParentID: &parentID, Role: "advisor", Turns: 0, CostUSD: 1.0},
			},
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

		// Case A: 4 canonical subagents all tied at 120k tokens. Precedence: planner > coder > tester > reviewer
		canonicalTiedLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", Role: "reviewer", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "2", Role: "tester", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "3", Role: "coder", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "4", Role: "planner", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			},
		}
		effA := domain.ComputeWorkflowEfficiency(canonicalTiedLedger)
		if effA.ContextWindowHealth.SubagentsPeakRole != "planner" {
			t.Errorf("expected 'planner' to win 4-way canonical tie-breaker, got %s", effA.ContextWindowHealth.SubagentsPeakRole)
		}

		// Case B: 4 custom subagents all tied at 120k tokens. Precedence: alphabetical (advisor > po > researcher > security-auditor)
		customTiedLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", Role: "security-auditor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "2", Role: "researcher", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "3", Role: "po", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "4", Role: "advisor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			},
		}
		effB := domain.ComputeWorkflowEfficiency(customTiedLedger)
		if effB.ContextWindowHealth.SubagentsPeakRole != "advisor" {
			t.Errorf("expected 'advisor' to win 4-way custom tie-breaker, got %s", effB.ContextWindowHealth.SubagentsPeakRole)
		}

		// Case C: Lowest-precedence canonical (reviewer) vs highest-precedence custom (advisor) tied at 120k.
		// Canonical must win over custom.
		mixedTiedLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", Role: "advisor", Tokens: domain.InvocationTokens{PeakContext: 120000}},
				{ID: "2", Role: "reviewer", Tokens: domain.InvocationTokens{PeakContext: 120000}},
			},
		}
		effC := domain.ComputeWorkflowEfficiency(mixedTiedLedger)
		if effC.ContextWindowHealth.SubagentsPeakRole != "reviewer" {
			t.Errorf("expected canonical 'reviewer' to beat custom 'advisor' on tie-break, got %s", effC.ContextWindowHealth.SubagentsPeakRole)
		}
	})

	t.Run("subagent invocations without parent ID excluded from turn velocity but included in cost", func(t *testing.T) {
		t.Parallel()
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "1", Role: "advisor", ParentID: nil, CostUSD: 2.0, Turns: 10},
			},
		}

		eff := domain.ComputeWorkflowEfficiency(ledger)
		// Cost included:
		if eff.RoleEffortDistributionPct["advisor"] != 100.0 {
			t.Errorf("expected 100%% effort for advisor, got %f", eff.RoleEffortDistributionPct["advisor"])
		}
		// Turns excluded because ParentID is nil:
		if eff.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns because ParentID is nil, got %f", eff.SubagentTurnVelocity.AvgTurnsPerInvocation)
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

		metrics := domain.ComputeRoleMetrics(invs)
		expected := []string{"coder", "reviewer", "advisor", "security-auditor"}
		if len(metrics) != len(expected) {
			t.Fatalf("expected %d metrics, got %d", len(expected), len(metrics))
		}
		for i, exp := range expected {
			if metrics[i].Role != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, metrics[i].Role)
			}
		}
	})
}

func TestDeriveMetricsLedger_RoleMappings(t *testing.T) {
	t.Parallel()

	t.Run("mapping execution invariance: rolls up build under orchestrator while preserving raw facts", func(t *testing.T) {
		t.Parallel()

		factLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
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
			},
		}

		customMappings := domain.RoleMappings{
			"build":        "orchestrator",
			"orchestrator": "orchestrator",
		}

		metrics := domain.DeriveMetricsLedger(factLedger, customMappings)

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
		if len(metrics.Facts.Invocations) != 2 {
			t.Fatalf("expected 2 raw invocations in Facts, got %d", len(metrics.Facts.Invocations))
		}
		if metrics.Facts.Invocations[0].Role != "build" {
			t.Errorf("expected Facts.Invocations[0].Role 'build', got %s", metrics.Facts.Invocations[0].Role)
		}
		if metrics.Facts.Invocations[1].Role != "orchestrator" {
			t.Errorf("expected Facts.Invocations[1].Role 'orchestrator', got %s", metrics.Facts.Invocations[1].Role)
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
		factLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
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
			},
		}

		customMappings := domain.RoleMappings{
			"build":     "orchestrator",
			"custom-qa": "tester",
		}

		metrics := domain.DeriveMetricsLedger(factLedger, customMappings)

		// Derived roles should be orchestrator, coder, tester
		roleNames := make([]string, len(metrics.Roles))
		for i, r := range metrics.Roles {
			roleNames[i] = r.Role
		}
		expectedRoles := []string{"orchestrator", "coder", "tester"}
		if len(roleNames) != len(expectedRoles) {
			t.Fatalf("expected roles %+v, got %+v", expectedRoles, roleNames)
		}
		for i, exp := range expectedRoles {
			if roleNames[i] != exp {
				t.Errorf("at index %d: expected %s, got %s", i, exp, roleNames[i])
			}
		}

		// Raw facts must still have build, coder, custom-qa
		if metrics.Facts.Invocations[0].Role != "build" {
			t.Errorf("expected raw fact build, got %s", metrics.Facts.Invocations[0].Role)
		}
		if metrics.Facts.Invocations[2].Role != "custom-qa" {
			t.Errorf("expected raw fact custom-qa, got %s", metrics.Facts.Invocations[2].Role)
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

func TestDeriveMetricsLedger_SubagentClassificationUnderRoleMapping(t *testing.T) {
	t.Parallel()

	pid := "ses-orch"
	baseLedger := domain.AgenticFactLedger{
		Invocations: []domain.InvocationFact{
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
		},
	}

	t.Run("Scenario A: custom-bot mapped to orchestrator is excluded from subagent metrics", func(t *testing.T) {
		t.Parallel()
		mappings := domain.RoleMappings{
			"custom-bot": "orchestrator",
		}
		metrics := domain.DeriveMetricsLedger(baseLedger, mappings)

		if len(metrics.WorkflowEfficiency.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty RoleEffortDistributionPct, got: %+v",
				metrics.WorkflowEfficiency.RoleEffortDistributionPct)
		}
		if metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation != 0 {
			t.Errorf("expected 0 avg turns, got %f",
				metrics.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})

	t.Run("Scenario B: custom-bot mapped to coder is included in subagent metrics with 100% effort", func(t *testing.T) {
		t.Parallel()
		mappings := domain.RoleMappings{
			"custom-bot": "coder",
		}
		metrics := domain.DeriveMetricsLedger(baseLedger, mappings)

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
		metrics := domain.DeriveMetricsLedger(baseLedger, mappings)

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
			TicketID:    &tid,
			Invocations: origInvs,
			Events:      origEvents,
		}

		// Snapshot before derivation
		invBefore0Role := factLedger.Invocations[0].Role
		invBefore1Role := factLedger.Invocations[1].Role
		invBefore2Role := factLedger.Invocations[2].Role
		evBefore0Role := factLedger.Events[0].Role
		evBefore1Role := factLedger.Events[1].Role
		evBefore2Role := factLedger.Events[2].Role

		mappings := domain.RoleMappings{
			"build":     "orchestrator",
			"custom-qa": "tester",
		}

		derived := domain.DeriveMetricsLedger(factLedger, mappings)

		// 1. Assert input factLedger was not mutated in place
		if factLedger.Invocations[0].Role != invBefore0Role || invBefore0Role != "build" {
			t.Errorf("input factLedger.Invocations[0].Role mutated: got %q, want 'build'", factLedger.Invocations[0].Role)
		}
		if factLedger.Invocations[1].Role != invBefore1Role || invBefore1Role != "custom-qa" {
			t.Errorf("input factLedger.Invocations[1].Role mutated: got %q, want 'custom-qa'", factLedger.Invocations[1].Role)
		}
		if factLedger.Invocations[2].Role != invBefore2Role || invBefore2Role != "security-auditor" {
			t.Errorf("input factLedger.Invocations[2].Role mutated: got %q, want 'security-auditor'", factLedger.Invocations[2].Role)
		}
		if factLedger.Events[0].Role != evBefore0Role || evBefore0Role != "build" {
			t.Errorf("input factLedger.Events[0].Role mutated: got %q, want 'build'", factLedger.Events[0].Role)
		}
		if factLedger.Events[1].Role != evBefore1Role || evBefore1Role != "custom-qa" {
			t.Errorf("input factLedger.Events[1].Role mutated: got %q, want 'custom-qa'", factLedger.Events[1].Role)
		}
		if factLedger.Events[2].Role != evBefore2Role || evBefore2Role != "security-auditor" {
			t.Errorf("input factLedger.Events[2].Role mutated: got %q, want 'security-auditor'", factLedger.Events[2].Role)
		}

		// 2. Assert derived.Facts retains the exact raw persona strings
		if derived.Facts.Invocations[0].Role != "build" {
			t.Errorf("derived.Facts.Invocations[0].Role is not raw 'build', got %q", derived.Facts.Invocations[0].Role)
		}
		if derived.Facts.Invocations[1].Role != "custom-qa" {
			t.Errorf("derived.Facts.Invocations[1].Role is not raw 'custom-qa', got %q", derived.Facts.Invocations[1].Role)
		}
		if derived.Facts.Invocations[2].Role != "security-auditor" {
			t.Errorf("derived.Facts.Invocations[2].Role is not raw 'security-auditor', got %q", derived.Facts.Invocations[2].Role)
		}
		if derived.Facts.Events[0].Role != "build" {
			t.Errorf("derived.Facts.Events[0].Role is not raw 'build', got %q", derived.Facts.Events[0].Role)
		}
		if derived.Facts.Events[1].Role != "custom-qa" {
			t.Errorf("derived.Facts.Events[1].Role is not raw 'custom-qa', got %q", derived.Facts.Events[1].Role)
		}
		if derived.Facts.Events[2].Role != "security-auditor" {
			t.Errorf("derived.Facts.Events[2].Role is not raw 'security-auditor', got %q", derived.Facts.Events[2].Role)
		}

		// 3. Assert derived aggregates use mapped roles (orchestrator, tester) and dynamic subagent (security-auditor)
		roleMap := make(map[string]domain.RoleMetric)
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
		derivedEmpty := domain.DeriveMetricsLedger(emptyLedger)

		if derivedEmpty.Totals.Invocations != 0 {
			t.Errorf("expected 0 invocations, got %d", derivedEmpty.Totals.Invocations)
		}
		if derivedEmpty.Totals.CostUSD != 0.0 {
			t.Errorf("expected 0.0 cost, got %f", derivedEmpty.Totals.CostUSD)
		}
		if derivedEmpty.Totals.ContextHealth != string(domain.ContextHealthHealthy) {
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
		zeroInvLedger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
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
			},
		}

		mappings := domain.RoleMappings{"build": "orchestrator"}
		derivedZero := domain.DeriveMetricsLedger(zeroInvLedger, mappings)

		if derivedZero.Totals.Invocations != 2 {
			t.Errorf("expected 2 invocations, got %d", derivedZero.Totals.Invocations)
		}
		if derivedZero.Totals.CostUSD != 0.0 {
			t.Errorf("expected 0.0 cost, got %f", derivedZero.Totals.CostUSD)
		}
		if derivedZero.Totals.CacheHitRatePct != 0.0 {
			t.Errorf("expected 0.0 cache hit rate, got %f", derivedZero.Totals.CacheHitRatePct)
		}
		if derivedZero.Totals.ContextHealth != string(domain.ContextHealthHealthy) {
			t.Errorf("expected HEALTHY context health, got %s", derivedZero.Totals.ContextHealth)
		}
		if len(derivedZero.WorkflowEfficiency.RoleEffortDistributionPct) != 0 {
			t.Errorf("expected empty effort distribution on zero cost, got %+v", derivedZero.WorkflowEfficiency.RoleEffortDistributionPct)
		}
		if derivedZero.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation != 0.0 {
			t.Errorf("expected 0.0 avg turns, got %f", derivedZero.WorkflowEfficiency.SubagentTurnVelocity.AvgTurnsPerInvocation)
		}
	})

	t.Run("variadic role mappings handling: nil, empty, and omitted mappings", func(t *testing.T) {
		t.Parallel()
		parentID := "ses-root"
		ledger := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "ses-root", Role: "build", CostUSD: 1.0},
				{ID: "ses-child", ParentID: &parentID, Role: "coder", CostUSD: 2.0, Turns: 3},
			},
		}

		// Omitted variadic: uses DefaultProjectConfig().RoleMappings where build -> orchestrator
		derivedDefault := domain.DeriveMetricsLedger(ledger)
		if len(derivedDefault.Roles) != 2 || derivedDefault.Roles[0].Role != "orchestrator" {
			t.Errorf("expected default mapping to map build to orchestrator, got: %+v", derivedDefault.Roles)
		}

		// Nil mapping: also falls back to DefaultProjectConfig().RoleMappings
		derivedNil := domain.DeriveMetricsLedger(ledger, nil)
		if len(derivedNil.Roles) != 2 || derivedNil.Roles[0].Role != "orchestrator" {
			t.Errorf("expected nil mapping to fallback to default, got: %+v", derivedNil.Roles)
		}

		// Custom mapping overriding default build: maps build to custom-builder
		derivedOverride := domain.DeriveMetricsLedger(ledger, domain.RoleMappings{"build": "custom-builder"})
		if len(derivedOverride.Roles) != 2 || derivedOverride.Roles[1].Role != "custom-builder" {
			t.Errorf("expected custom mapping to override build to 'custom-builder', got: %+v", derivedOverride.Roles)
		}

		// Empty mapping: build resolves via domain default to orchestrator, while unmapped custom role retains raw name
		ledgerWithCustom := domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
				{ID: "ses-custom", Role: "specialist-agent", CostUSD: 3.0},
			},
		}
		derivedEmptyMap := domain.DeriveMetricsLedger(ledgerWithCustom, domain.RoleMappings{})
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
				targetRole:         domain.RoleOrchestrator,
				expectSubagentCost: false,
				expectTurnCounted:  false,
			},
			{
				name:               "child session with raw build mapped to orchestrator",
				rawRole:            "build",
				parentID:           &parentID,
				targetRole:         domain.RoleOrchestrator,
				expectSubagentCost: false,
				expectTurnCounted:  false,
			},
			{
				name:               "child session with custom-worker mapped to coder",
				rawRole:            "custom-worker",
				parentID:           &parentID,
				targetRole:         domain.RoleCoder,
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
				name:               "root session with unmapped custom subagent role: cost counted, turns excluded",
				rawRole:            "independent-agent",
				parentID:           nil,
				targetRole:         "independent-agent",
				expectSubagentCost: true,
				expectTurnCounted:  false,
			},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				factLedger := domain.AgenticFactLedger{
					Invocations: []domain.InvocationFact{
						{
							ID:       "ses-test",
							ParentID: tc.parentID,
							Role:     tc.rawRole,
							CostUSD:  1.50,
							Turns:    5,
							Tokens: domain.InvocationTokens{
								Input: 1000, Output: 200, PeakContext: 10000,
							},
						},
					},
				}

				mappings := domain.RoleMappings{
					tc.rawRole: string(tc.targetRole),
				}
				metrics := domain.DeriveMetricsLedger(factLedger, mappings)

				targetStr := string(tc.targetRole)
				_, hasSubCost := metrics.WorkflowEfficiency.RoleEffortDistributionPct[targetStr]
				if hasSubCost != tc.expectSubagentCost {
					t.Errorf("RoleEffortDistributionPct[%q] presence = %v, want %v", targetStr, hasSubCost, tc.expectSubagentCost)
				}

				turns, hasTurns := metrics.WorkflowEfficiency.SubagentTurnVelocity.ByRole[targetStr]
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

	t.Run("falls back to updated_at when completed_at is nil", func(t *testing.T) {
		t.Parallel()
		updated := "2026-03-09T10:00:00Z"
		started := "2026-03-08T08:00:00Z"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: nil,
			UpdatedAt:   updated,
			StartedAt:   &started,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if !ok {
			t.Fatal("expected ok = true")
		}
		expected, _ := time.Parse(time.RFC3339, updated)
		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("falls back to started_at when completed_at and updated_at are absent", func(t *testing.T) {
		t.Parallel()
		started := "2026-03-08T08:00:00Z"
		ledger := domain.AgenticMetricsLedger{
			CompletedAt: nil,
			UpdatedAt:   "",
			StartedAt:   &started,
		}
		got, ok := domain.GetLedgerCompletionDate(ledger)
		if !ok {
			t.Fatal("expected ok = true")
		}
		expected, _ := time.Parse(time.RFC3339, started)
		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
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

	t.Run("secondary fallback to started_at when completed_at and updated_at are malformed", func(t *testing.T) {
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
		if !ok {
			t.Fatalf("expected fallback to valid started_at")
		}
		expected, _ := time.Parse(time.RFC3339, validStarted)
		if !tGot.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, tGot)
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

	t.Run("fallback hierarchy: updated_at, started_at, and all-null exclusion", func(t *testing.T) {
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
		}

		filtered := domain.FilterLedgersByCompletionDate(items, &sinceUTC, &untilUTC)
		if len(filtered) != 2 {
			t.Fatalf("expected 2 fallbacks included and 2 dateless excluded, got %d", len(filtered))
		}
		if *filtered[0].TicketID != "FALLBACK-UPDATED" || *filtered[1].TicketID != "FALLBACK-STARTED" {
			t.Errorf("unexpected filtered ledgers: %v, %v", *filtered[0].TicketID, *filtered[1].TicketID)
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

	t.Run("fallback timestamps across completed_at, updated_at, and started_at sort accurately", func(t *testing.T) {
		t.Parallel()
		items := []domain.AgenticMetricsLedger{
			{TicketID: ptr("EARLY-COMPLETED"), CompletedAt: &dEarly},
			{TicketID: ptr("MID-UPDATED"), CompletedAt: nil, UpdatedAt: dMid},
			{TicketID: ptr("LATE-STARTED"), CompletedAt: nil, UpdatedAt: "", StartedAt: &dLate},
		}

		sortedDesc := domain.SortLedgers(items, domain.SortFieldCompletedAt, false)
		expectedIDs := []string{"LATE-STARTED", "MID-UPDATED", "EARLY-COMPLETED"}
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
