package http

import (
	"agenticlens/internal/domain"
)

// FleetSummaryDTO is the concrete nullable cost projection of FleetSummary.
type FleetSummaryDTO struct {
	SkillAdoption            []domain.SkillAdoptionFact              `json:"skill_adoption"`
	TotalTickets             int64                                   `json:"total_tickets"`
	TotalCostUSD             *float64                                `json:"total_cost_usd"`
	CostSummary              domain.CostSummary                      `json:"cost_summary"`
	RoleCostSummaries        map[domain.AgentRole]domain.CostSummary `json:"role_cost_summaries"`
	ModelCostSummaries       map[string]domain.CostSummary           `json:"model_cost_summaries"`
	TotalInvocations         int64                                   `json:"total_invocations"`
	TotalInputTokens         int64                                   `json:"total_input_tokens"`
	TotalOutputTokens        int64                                   `json:"total_output_tokens"`
	TotalCacheReadTokens     int64                                   `json:"total_cache_read_tokens"`
	OverallCacheHitRatePct   float64                                 `json:"overall_cache_hit_rate_pct"`
	TotalDurationSeconds     int64                                   `json:"total_duration_seconds"`
	FormattedTotalDuration   string                                  `json:"formatted_total_duration"`
	TotalLinesAdded          int64                                   `json:"total_lines_added"`
	TotalLinesDeleted        int64                                   `json:"total_lines_deleted"`
	TotalFilesChanged        int64                                   `json:"total_files_changed"`
	TotalReasoningTokens     int64                                   `json:"total_reasoning_tokens"`
	OverallReasoningRatioPct float64                                 `json:"overall_reasoning_ratio_pct"`
	RoleCosts                map[domain.AgentRole]*float64           `json:"role_costs"`
	ModelCosts               map[string]*float64                     `json:"model_costs"`
	AnomalySummary           domain.FleetAnomalySummary              `json:"anomaly_summary"`
}

func FleetSummaryDTOFromDomain(value domain.FleetSummary) FleetSummaryDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.TotalCostUSD, value.TotalInvocations)
	roles := effectiveRoleCostSummaries(value.RoleCostSummaries, value.RoleCosts)
	models := effectiveModelCostSummaries(value.ModelCostSummaries, value.ModelCosts)
	return FleetSummaryDTO{
		SkillAdoption:            value.SkillAdoption,
		TotalTickets:             value.TotalTickets,
		TotalCostUSD:             summary.CompleteAmount(),
		CostSummary:              summary,
		RoleCostSummaries:        roles,
		ModelCostSummaries:       models,
		TotalInvocations:         value.TotalInvocations,
		TotalInputTokens:         value.TotalInputTokens,
		TotalOutputTokens:        value.TotalOutputTokens,
		TotalCacheReadTokens:     value.TotalCacheReadTokens,
		OverallCacheHitRatePct:   value.OverallCacheHitRatePct,
		TotalDurationSeconds:     value.TotalDurationSeconds,
		FormattedTotalDuration:   value.FormattedTotalDuration,
		TotalLinesAdded:          value.TotalLinesAdded,
		TotalLinesDeleted:        value.TotalLinesDeleted,
		TotalFilesChanged:        value.TotalFilesChanged,
		TotalReasoningTokens:     value.TotalReasoningTokens,
		OverallReasoningRatioPct: value.OverallReasoningRatioPct,
		RoleCosts:                roleCostAmounts(roles),
		ModelCosts:               modelCostAmounts(models),
		AnomalySummary:           value.AnomalySummary,
	}
}

func effectiveRoleCostSummaries(values map[domain.AgentRole]domain.CostSummary, amounts map[domain.AgentRole]float64) map[domain.AgentRole]domain.CostSummary {
	result := make(map[domain.AgentRole]domain.CostSummary, len(values))
	for key, value := range values {
		result[key] = value
	}
	for key, amount := range amounts {
		if _, exists := result[key]; !exists {
			result[key] = domain.EffectiveCostSummary(domain.CostSummary{}, amount, 1)
		}
	}
	return result
}

func effectiveModelCostSummaries(values map[string]domain.CostSummary, amounts map[string]float64) map[string]domain.CostSummary {
	result := make(map[string]domain.CostSummary, len(values))
	for key, value := range values {
		result[key] = value
	}
	for key, amount := range amounts {
		if _, exists := result[key]; !exists {
			result[key] = domain.EffectiveCostSummary(domain.CostSummary{}, amount, 1)
		}
	}
	return result
}

func roleCostAmounts(values map[domain.AgentRole]domain.CostSummary) map[domain.AgentRole]*float64 {
	result := make(map[domain.AgentRole]*float64, len(values))
	for key, value := range values {
		result[key] = value.CompleteAmount()
	}
	return result
}

func modelCostAmounts(values map[string]domain.CostSummary) map[string]*float64 {
	result := make(map[string]*float64, len(values))
	for key, value := range values {
		result[key] = value.CompleteAmount()
	}
	return result
}
