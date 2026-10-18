package domain_test

import (
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestComputeTotals(t *testing.T) {
	t.Run("empty invocations", func(t *testing.T) {
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
	t.Run("zero division defenses", func(t *testing.T) {
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
