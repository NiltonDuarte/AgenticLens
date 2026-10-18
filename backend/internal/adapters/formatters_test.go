package adapters_test

import (
	"os"
	"path/filepath"
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
		Facts: domain.AgenticFactLedger{
			Invocations: []domain.InvocationFact{
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
			Tools: []domain.ToolFact{
				{Name: "bash", Calls: 5, Errors: 0},
			},
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
				ByRole:                map[string]float64{"coder": 3.0},
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
				Status:                    "ELEVATED",
				OrchestratorPeakTokens:    50000,
				OrchestratorSaturationPct: 25.0,
				SubagentsPeakTokens:       120000,
				SubagentsPeakRole:         "coder",
				Compactions:               0,
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

	t.Run("multiple anomalies with mixed severities", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID: &tid,
			Anomalies: []domain.Anomaly{
				{
					RuleID:      domain.RuleToolRetryStreak,
					Severity:    domain.SeverityCritical,
					Title:       "Tool retry streak: bash (5 consecutive errors)",
					Description: "Tool 'bash' encountered 5 consecutive execution failures (session: s1, turn: 3)",
					Context: domain.AnomalyContext{
						ToolName: &toolBash,
						Count:    &count5,
						Turn:     &turn3,
					},
				},
				{
					RuleID:      domain.RuleToolHighFailureRate,
					Severity:    domain.SeverityWarning,
					Title:       "High tool failure rate: git (33.3%)",
					Description: "Tool 'git' failed 2 out of 6 calls (33.3% failure rate, threshold: >20%)",
					Context: domain.AnomalyContext{
						ToolName:       &toolGit,
						Count:          &count2,
						TotalCalls:     &total6,
						FailureRatePct: &failPct,
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
		if !strings.Contains(tableMD, "- [CRITICAL] **TOOL_RETRY_STREAK**: Tool retry streak: bash (5 consecutive errors) — Tool 'bash' encountered 5 consecutive execution failures (session: s1, turn: 3)") {
			t.Errorf("expected formatted CRITICAL anomaly bullet, got:\n%s", tableMD)
		}
		if !strings.Contains(tableMD, "- [WARNING] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: git (33.3%) — Tool 'git' failed 2 out of 6 calls (33.3% failure rate, threshold: >20%)") {
			t.Errorf("expected formatted WARNING anomaly bullet, got:\n%s", tableMD)
		}
	})

	t.Run("single anomaly uses singular phrasing", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID: &tid,
			Anomalies: []domain.Anomaly{
				{
					RuleID:      domain.RuleContextSaturationSpike,
					Severity:    domain.SeverityCritical,
					Title:       "Context ceiling breached (165k tokens)",
					Description: "Context tokens reached 165k, breaching 160k critical ceiling",
				},
			},
		}

		tableMD := adapters.FormatMarkdownMetricsTable(ledger)
		if !strings.Contains(tableMD, "- **Status:** 1 execution anomaly detected (1 CRITICAL).") {
			t.Errorf("expected singular phrasing for 1 anomaly, got:\n%s", tableMD)
		}
		if !strings.Contains(tableMD, "- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (165k tokens) — Context tokens reached 165k, breaching 160k critical ceiling") {
			t.Errorf("expected formatted bullet, got:\n%s", tableMD)
		}
	})
}

func TestUpdateSummaryFile(t *testing.T) {
	tempDir := t.TempDir()
	summaryFile := filepath.Join(tempDir, "SUMMARY.md")

	tableMD := "### Workflow Efficiency\n- **Iterations:** 1\n\n### Anomalies & Health Flags\n- **Status:** No execution anomalies detected (clean session).\n"

	t.Run("append to new file", func(t *testing.T) {
		err := adapters.UpdateSummaryFile(summaryFile, tableMD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content, _ := os.ReadFile(summaryFile)
		if !strings.Contains(string(content), tableMD) {
			t.Errorf("expected table in new file, got %s", string(content))
		}
	})

	t.Run("idempotent replacement in existing file with other sections", func(t *testing.T) {
		initial := "# Title\n\n### Workflow Efficiency\n- Old Content\n\n### Anomalies & Health Flags\n- Old Anomaly\n\n## Next Section\nSomething else\n"
		if err := os.WriteFile(summaryFile, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		newTable := "### Workflow Efficiency\n- New Content\n\n### Anomalies & Health Flags\n- **Status:** 1 execution anomaly detected (1 CRITICAL).\n"
		err := adapters.UpdateSummaryFile(summaryFile, newTable)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := os.ReadFile(summaryFile)
		updatedStr := string(updated)
		if strings.Contains(updatedStr, "Old Content") || strings.Contains(updatedStr, "Old Anomaly") {
			t.Errorf("expected Old Content and Old Anomaly to be replaced")
		}
		if !strings.Contains(updatedStr, "New Content") {
			t.Errorf("expected New Content to be present")
		}
		if !strings.Contains(updatedStr, "1 execution anomaly detected") {
			t.Errorf("expected new anomaly status to be present")
		}
		if !strings.Contains(updatedStr, "## Next Section") {
			t.Errorf("expected Next Section to be preserved")
		}
		if strings.Count(updatedStr, "### Workflow Efficiency") != 1 {
			t.Errorf("expected exactly 1 Workflow Efficiency header, got %d", strings.Count(updatedStr, "### Workflow Efficiency"))
		}
	})

	t.Run("replaces legacy header cleanly", func(t *testing.T) {
		legacyFile := filepath.Join(tempDir, "LEGACY_SUMMARY.md")
		initialLegacy := "# Summary\n\n### Agentic Execution Metrics\n- Old metrics\n\n## Results\nDone\n"
		if err := os.WriteFile(legacyFile, []byte(initialLegacy), 0644); err != nil {
			t.Fatal(err)
		}

		err := adapters.UpdateSummaryFile(legacyFile, tableMD)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(legacyFile)
		contentStr := string(content)
		if strings.Contains(contentStr, "Old metrics") {
			t.Errorf("expected legacy metrics to be replaced")
		}
		if !strings.Contains(contentStr, "### Workflow Efficiency") {
			t.Errorf("expected new Workflow Efficiency header")
		}
		if !strings.Contains(contentStr, "## Results") {
			t.Errorf("expected subsequent section to be preserved")
		}
	})
}
