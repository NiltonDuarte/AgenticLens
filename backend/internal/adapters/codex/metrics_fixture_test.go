package codex_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func codexMetricConfig() domain.ProjectConfig {
	return domain.ProjectConfig{
		RoleMappings: domain.RoleMappings{},
		RoleOrder:    []string{},
		AnomalyThresholds: domain.AnomalyThresholdsConfig{
			ErrorRateMinCalls: 5, ErrorRateThreshold: .25, ErrorRateCriticalThreshold: .7,
			ContextCeilingWarningTokens: 160000, ContextCeilingCriticalTokens: 200000,
			ConsecutiveErrorsStreakWarning: 4, ConsecutiveErrorsStreakCritical: 8,
			ToolThrashingMinFailures: 3, ToolThrashingCriticalFailures: 5,
			TurnUsageWarningTokens: 60000, TurnUsageCriticalTokens: 120000,
			ContextGrowthWarningTokens: 80000, ContextGrowthCriticalTokens: 160000,
			StallThresholdSeconds: 120,
		},
	}
}

func deriveCodexSourceMetrics(t *testing.T, facts domain.AgenticFactLedger, events []domain.EventFact) domain.AgenticMetricsLedger {
	t.Helper()
	complete := make([]string, 0, len(facts.Sessions))
	for _, owner := range facts.Sessions {
		complete = append(complete, owner.ID)
	}
	policy := codexMetricConfig()
	measured, err := domain.PersistSourceAnomalySummaries(facts, domain.SourceEvents{Events: events, CompleteSessionIDs: complete}, policy.AnomalyThresholds)
	if err != nil {
		t.Fatal(err)
	}
	return domain.DeriveMetricsLedger(measured, policy)
}
