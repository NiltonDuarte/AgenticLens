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

func TestUpdateSummaryFile(t *testing.T) {
	tempDir := t.TempDir()
	summaryFile := filepath.Join(tempDir, "SUMMARY.md")

	tableMD := "### Workflow Efficiency\n- **Iterations:** 1\n"

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
		initial := "# Title\n\n### Workflow Efficiency\n- Old Content\n\n## Next Section\nSomething else\n"
		if err := os.WriteFile(summaryFile, []byte(initial), 0644); err != nil {
			t.Fatal(err)
		}

		newTable := "### Workflow Efficiency\n- New Content\n"
		err := adapters.UpdateSummaryFile(summaryFile, newTable)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := os.ReadFile(summaryFile)
		updatedStr := string(updated)
		if strings.Contains(updatedStr, "Old Content") {
			t.Errorf("expected Old Content to be replaced")
		}
		if !strings.Contains(updatedStr, "New Content") {
			t.Errorf("expected New Content to be present")
		}
		if !strings.Contains(updatedStr, "## Next Section") {
			t.Errorf("expected Next Section to be preserved")
		}
	})
}
