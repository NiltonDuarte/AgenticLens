package adapters_test

import (
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestFormatMarkdownMetricsTable(t *testing.T) {
	tid := "AL-1"
	parentID := "ses-1"
	ledger := domain.AgenticMetricsLedger{
		TicketID: &tid,
		Facts: domain.AgenticFactLedger{Sessions: sessionFactsWithInvs([]testInvFact{
			{
				ID:                    "ses-1",
				Role:                  "orchestrator",
				Model:                 "model-a",
				User:                  "developer",
				ActiveDurationSeconds: 100,
				Tokens: domain.InvocationTokens{
					Input:       1000,
					Output:      500,
					CacheRead:   2000,
					Reasoning:   200,
					PeakContext: 50000,
				},
				CostUSD: 0.1,
			},
			{
				ID:                    "ses-2",
				ParentID:              &parentID,
				Role:                  "coder",
				Model:                 "model-b",
				User:                  "developer",
				ActiveDurationSeconds: 200,
				Turns:                 3,
				Tokens: domain.InvocationTokens{
					Input:       5000,
					Output:      1000,
					CacheRead:   10000,
					Reasoning:   500,
					PeakContext: 120000,
				},
				CostUSD: 0.4,
			},
		},
			[]domain.ToolFact{
				{SessionID: "session-a", Name: "bash", Calls: 5, Errors: 0},
			}, nil),
		},
		WorkflowEfficiency: domain.WorkflowEfficiency{
			SDLCIterations:          1,
			ActiveDurationSeconds:   300,
			ActiveDurationFormatted: "5m 0s",
			RoleEffortDistributionPct: map[string]float64{
				"coder": 100.0,
			},
			SubagentTurnVelocity: domain.SubagentTurnVelocity{
				AvgTurnsPerInvocation: 3.0,
				PeakTurns:             3,
				ByRole:                map[domain.AgentRole]float64{domain.AgentRole("coder"): 3.0},
			},
			TestToProductionRatio: domain.TestToCodeRatio{
				Ratio:          "1.0:1",
				TestLinesAdded: 50,
				ProdLinesAdded: 50,
			},
			ReasoningTokens:   700,
			ReasoningRatioPct: 31.8,
			CodeDelivered: domain.CodeDeliveredStats{
				LinesAdded:   100,
				LinesDeleted: 20,
				FilesChanged: 4,
			},
			ToolExecution: domain.ToolExecutionHealth{
				TotalCalls:           5,
				ErrorsAndCorrections: 0,
				SuccessRatePct:       100.0,
			},
			ContextWindowHealth: domain.ContextWindowHealth{
				Status:              "ELEVATED",
				RootPeakTokens:      50000,
				RootSaturationPct:   25.0,
				SubagentsPeakTokens: 120000,
				SubagentsPeakRole:   "coder",
				Compactions:         0,
			},
		},
		Totals: domain.GroupedMetric{
			Invocations:       2,
			InputTokens:       6000,
			OutputTokens:      1500,
			CacheReadTokens:   12000,
			CacheHitRatePct:   66.7,
			CostUSD:           0.5,
			PeakContextTokens: 120000,
			ContextHealth:     "ELEVATED",
		},
		Roles: []domain.RoleMetric{
			{
				Role:              "orchestrator",
				Invocations:       1,
				InputTokens:       1000,
				OutputTokens:      500,
				CacheReadTokens:   2000,
				CacheHitRatePct:   66.7,
				CostUSD:           0.1,
				PeakContextTokens: 50000,
				ContextHealth:     "HEALTHY",
			},
			{
				Role:              "coder",
				Invocations:       1,
				InputTokens:       5000,
				OutputTokens:      1000,
				CacheReadTokens:   10000,
				CacheHitRatePct:   66.7,
				CostUSD:           0.4,
				PeakContextTokens: 120000,
				ContextHealth:     "ELEVATED",
			},
		},
		Models: []domain.ModelMetric{
			{
				Model:             "model-b",
				Invocations:       1,
				InputTokens:       5000,
				OutputTokens:      1000,
				CacheReadTokens:   10000,
				CacheHitRatePct:   66.7,
				CostUSD:           0.4,
				PeakContextTokens: 120000,
				ContextHealth:     "ELEVATED",
			},
			{
				Model:             "model-a",
				Invocations:       1,
				InputTokens:       1000,
				OutputTokens:      500,
				CacheReadTokens:   2000,
				CacheHitRatePct:   66.7,
				CostUSD:           0.1,
				PeakContextTokens: 50000,
				ContextHealth:     "HEALTHY",
			},
		},
	}

	tableMD := adapters.FormatMarkdownMetricsTable(ledger)

	// Invariants check
	if !strings.Contains(tableMD, "### Workflow Efficiency") {
		t.Errorf("missing Workflow Efficiency heading")
	}
	if !strings.Contains(tableMD, "### Anomalies & Health Flags") {
		t.Errorf("missing Anomalies & Health Flags heading")
	}
	if !strings.Contains(tableMD, "No execution anomalies detected (clean session).") {
		t.Errorf("missing clean session status in clean run")
	}
	if !strings.Contains(tableMD, "### Agentic Execution Metrics") {
		t.Errorf("missing Agentic Execution Metrics heading")
	}
	if !strings.Contains(tableMD, "**Contributors:** developer") {
		t.Errorf("missing contributor developer")
	}
	if !strings.Contains(tableMD, "| **orchestrator** |") {
		t.Errorf("missing orchestrator in role table")
	}
	if !strings.Contains(tableMD, "| **coder** |") {
		t.Errorf("missing coder in role table")
	}
	if !strings.Contains(tableMD, "| **model-b** |") {
		t.Errorf("missing model-b in model table")
	}
}

func TestFormatMarkdownMetricsTable_WithAnomalies(t *testing.T) {
	tid := "AL-ANOMALIES"
	toolBash := "bash"
	toolGit := "git"
	turn3 := int64(3)
	count5 := int64(5)
	count2 := int64(2)
	total6 := int64(6)
	failPct := 33.3
	threshold := ">20%"
	contextTokens := int64(165_000)

	t.Run("multiple anomalies with mixed severities", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID: &tid,
			Anomalies: []domain.Anomaly{
				{
					RuleID:   domain.RuleToolRetryStreak,
					Severity: domain.SeverityCritical,
					Context: domain.AnomalyContext{
						ToolName: &toolBash,
						Count:    &count5,
						Turn:     &turn3,
					},
				},
				{
					RuleID:   domain.RuleToolHighFailureRate,
					Severity: domain.SeverityWarning,
					Context: domain.AnomalyContext{
						ToolName:       &toolGit,
						Count:          &count2,
						TotalCalls:     &total6,
						FailureRatePct: &failPct,
						Threshold:      &threshold,
					},
				},
			},
		}

		tableMD := adapters.FormatMarkdownMetricsTable(ledger)

		if !strings.Contains(tableMD, "### Anomalies & Health Flags") {
			t.Fatalf("expected Anomalies & Health Flags heading")
		}
		if !strings.Contains(tableMD, "- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).") {
			t.Errorf("expected 2 execution anomalies status summary, got:\n%s", tableMD)
		}
		if !strings.Contains(tableMD, "- [CRITICAL] **TOOL_RETRY_STREAK**: Tool retry streak: bash (5 consecutive errors) — Tool 'bash' encountered 5 consecutive execution failures.") {
			t.Errorf("expected formatted CRITICAL anomaly bullet, got:\n%s", tableMD)
		}
		if !strings.Contains(tableMD, "- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: git (33.3%) — Tool 'git' failed 2 out of 6 calls (33.3% failure rate, threshold: >20%).") {
			t.Errorf("expected formatted WARNING anomaly bullet, got:\n%s", tableMD)
		}
	})

	t.Run("single anomaly uses singular phrasing", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID: &tid,
			Anomalies: []domain.Anomaly{
				{
					RuleID:   domain.RuleContextSaturationSpike,
					Severity: domain.SeverityCritical,
					Context:  domain.AnomalyContext{ContextTokens: &contextTokens, Threshold: strPtr("160000")},
				},
			},
		}

		tableMD := adapters.FormatMarkdownMetricsTable(ledger)
		if !strings.Contains(tableMD, "- **Status:** 1 execution anomaly detected (1 CRITICAL).") {
			t.Errorf("expected singular phrasing for 1 anomaly, got:\n%s", tableMD)
		}
		if !strings.Contains(tableMD, "- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165k tokens) — Context tokens reached 165k, breaching the 160k critical ceiling.") {
			t.Errorf("expected formatted bullet, got:\n%s", tableMD)
		}
	})
}

func TestFormatMarkdownMetricsTable_DynamicSubagentRoles(t *testing.T) {
	t.Parallel()
	tid := "AL-24"
	parentID := "ses-orch"
	ledger := domain.AgenticMetricsLedger{
		TicketID: &tid,
		Facts: domain.AgenticFactLedger{Sessions: sessionFactsWithInvs([]testInvFact{
			{
				ID:      "ses-orch",
				Role:    "orchestrator",
				Model:   "model-a",
				User:    "developer",
				CostUSD: 1.0,
			},
			{
				ID:       "ses-adv",
				ParentID: &parentID,
				Role:     "advisor",
				Model:    "model-a",
				User:     "developer",
				CostUSD:  1.0,
				Turns:    4,
			},
			{
				ID:       "ses-po",
				ParentID: &parentID,
				Role:     "po",
				Model:    "model-b",
				User:     "developer",
				CostUSD:  1.0,
				Turns:    8,
			},
		}, nil, nil),
		},
		WorkflowEfficiency: domain.WorkflowEfficiency{
			RoleEffortDistributionPct: map[string]float64{
				"advisor": 50.0,
				"po":      50.0,
			},
			SubagentTurnVelocity: domain.SubagentTurnVelocity{
				AvgTurnsPerInvocation: 6.0,
				PeakTurns:             8,
				ByRole: map[domain.AgentRole]float64{
					"advisor": 4.0,
					"po":      8.0,
				},
			},
			ContextWindowHealth: domain.ContextWindowHealth{
				Status:              "HEALTHY",
				SubagentsPeakRole:   "advisor",
				SubagentsPeakTokens: 50000,
			},
		},
	}

	tableMD := adapters.FormatMarkdownMetricsTable(ledger)

	// Check that custom roles Advisor and Po appear in Role Effort Distribution (Subagents)
	if !strings.Contains(tableMD, "- **Role Effort Distribution (Subagents):** Advisor 50.0%, Po 50.0%") {
		t.Errorf("expected custom subagents in role effort distribution, got:\n%s", tableMD)
	}

	// Check that custom roles Advisor and Po appear in Subagent Turn Velocity
	if !strings.Contains(tableMD, "- **Subagent Turn Velocity & Thrashing:** 6.0 avg turns/run (peak: 8 turns) | Advisor 4.0, Po 8.0") {
		t.Errorf("expected custom subagents in turn velocity, got:\n%s", tableMD)
	}
}

func TestFormatMarkdownMetricsTable_MixedCanonicalAndCustomSubagentsPrecedence(t *testing.T) {
	t.Parallel()
	tid := "AL-24-PRECEDENCE"
	ledger := domain.AgenticMetricsLedger{
		TicketID: &tid,
		WorkflowEfficiency: domain.WorkflowEfficiency{
			RoleEffortDistributionPct: map[string]float64{
				"po":      10.0,
				"coder":   40.0,
				"advisor": 20.0,
				"planner": 30.0,
			},
			SubagentTurnVelocity: domain.SubagentTurnVelocity{
				AvgTurnsPerInvocation: 4.2,
				PeakTurns:             8,
				ByRole: map[domain.AgentRole]float64{
					"po":      5.0,
					"coder":   6.0,
					"advisor": 4.0,
					"planner": 2.0,
				},
			},
			ContextWindowHealth: domain.ContextWindowHealth{
				Status:              "HEALTHY",
				SubagentsPeakRole:   "planner",
				SubagentsPeakTokens: 50000,
			},
		},
	}

	customOrder := []string{"planner", "coder", "tester", "reviewer"}
	tableMD := adapters.FormatMarkdownMetricsTable(ledger, customOrder)

	// Custom subagent priority (Planner, Coder) must appear first in specified order,
	// followed by remaining subagents (Advisor, Po) in alphabetical order.
	expectedDist := "- **Role Effort Distribution (Subagents):** Planner 30.0%, Coder 40.0%, Advisor 20.0%, Po 10.0%"
	if !strings.Contains(tableMD, expectedDist) {
		t.Errorf("expected custom order precedence in role distribution:\nwant: %s\ngot:\n%s", expectedDist, tableMD)
	}

	expectedVel := "- **Subagent Turn Velocity & Thrashing:** 4.2 avg turns/run (peak: 8 turns) | Planner 2.0, Coder 6.0, Advisor 4.0, Po 5.0"
	if !strings.Contains(tableMD, expectedVel) {
		t.Errorf("expected custom order precedence in turn velocity:\nwant: %s\ngot:\n%s", expectedVel, tableMD)
	}

	// Alphabetical order default when no custom order provided
	tableMDDefault := adapters.FormatMarkdownMetricsTable(ledger)
	expectedDistDefault := "- **Role Effort Distribution (Subagents):** Advisor 20.0%, Coder 40.0%, Planner 30.0%, Po 10.0%"
	if !strings.Contains(tableMDDefault, expectedDistDefault) {
		t.Errorf("expected alphabetical order default in role distribution:\nwant: %s\ngot:\n%s", expectedDistDefault, tableMDDefault)
	}
	expectedVelDefault := "- **Subagent Turn Velocity & Thrashing:** 4.2 avg turns/run (peak: 8 turns) | Advisor 4.0, Coder 6.0, Planner 2.0, Po 5.0"
	if !strings.Contains(tableMDDefault, expectedVelDefault) {
		t.Errorf("expected alphabetical order default in turn velocity:\nwant: %s\ngot:\n%s", expectedVelDefault, tableMDDefault)
	}
}

func TestFormatMarkdownMetricsTable_EmptyAndZeroValues(t *testing.T) {
	t.Parallel()
	emptyLedger := domain.AgenticMetricsLedger{}
	tableMD := adapters.FormatMarkdownMetricsTable(emptyLedger)

	requiredHeadings := []string{
		"### Workflow Efficiency",
		"### Anomalies & Health Flags",
		"### Agentic Execution Metrics",
		"#### By Subagent Role",
		"#### By Model",
	}

	for _, heading := range requiredHeadings {
		if !strings.Contains(tableMD, heading) {
			t.Errorf("expected markdown to contain heading %q", heading)
		}
	}

	requiredKPIs := []string{
		"- **Software Development Life Cycle Iterations:**",
		"- **Active Agent Duration:**",
		"- **Context Window Health:**",
		"- **Role Effort Distribution (Subagents):**",
		"- **Subagent Turn Velocity & Thrashing:**",
		"- **Test-to-Production Code Ratio:**",
		"- **Code Delivered:**",
		"- **Reasoning Effort:**",
		"- **Tool Health & Self-Correction:**",
	}

	for _, kpi := range requiredKPIs {
		if !strings.Contains(tableMD, kpi) {
			t.Errorf("expected markdown to contain KPI row %q", kpi)
		}
	}
}

func TestFormatMarkdownMetricsTable_SingleRootAndZeroSubagents(t *testing.T) {
	t.Parallel()
	tid := "GEN-1"
	ledger := domain.AgenticMetricsLedger{
		TicketID: &tid,
		Facts: domain.AgenticFactLedger{Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:                    "ses-root-gen",
				Role:                  "general",
				ActiveDurationSeconds: 120,
				Tokens: domain.InvocationTokens{
					PeakContext: 60000,
				},
				CostUSD: 0.50,
			},
		}, nil, nil),
		},
		WorkflowEfficiency: domain.WorkflowEfficiency{
			ActiveDurationFormatted:   "2m 0s",
			RoleEffortDistributionPct: map[string]float64{},
			SubagentTurnVelocity: domain.SubagentTurnVelocity{
				AvgTurnsPerInvocation: 0.0,
				PeakTurns:             0,
				ByRole:                map[domain.AgentRole]float64{},
			},
			ContextWindowHealth: domain.ContextWindowHealth{
				Status:              "HEALTHY",
				RootPeakTokens:      60000,
				RootSaturationPct:   30.0,
				SubagentsPeakTokens: 0,
				SubagentsPeakRole:   "none",
				Compactions:         0,
			},
		},
		Totals: domain.GroupedMetric{
			Invocations: 1,
			CostUSD:     0.50,
		},
	}

	tableMD := adapters.FormatMarkdownMetricsTable(ledger)

	expectedHealth := "- **Context Window Health:** HEALTHY — Root agent peak 60k tokens (30% of 200k healthy ceiling, 0 compactions) | Subagents peak 0 (None)"
	if !strings.Contains(tableMD, expectedHealth) {
		t.Errorf("expected clean zero-subagents health format:\nwant: %s\ngot:\n%s", expectedHealth, tableMD)
	}

	expectedDist := "- **Role Effort Distribution (Subagents):** N/A"
	if !strings.Contains(tableMD, expectedDist) {
		t.Errorf("expected N/A for empty subagents effort distribution:\nwant: %s\ngot:\n%s", expectedDist, tableMD)
	}

	expectedVel := "- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)"
	if !strings.Contains(tableMD, expectedVel) {
		t.Errorf("expected zero turn velocity format:\nwant: %s\ngot:\n%s", expectedVel, tableMD)
	}
}
