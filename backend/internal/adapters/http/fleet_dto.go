package http

import (
	"agenticlens/internal/domain"
)

// FleetSummaryDTO is the concrete nullable cost projection of FleetSummary.
type FleetSummaryDTO struct {
	ProvisionalSeconds           int64                                             `json:"provisional_seconds,omitempty"`
	SkillAdoption                []domain.SkillAdoptionFact                        `json:"skill_adoption,omitzero"`
	TotalTickets                 int64                                             `json:"total_tickets"`
	TotalCostUSD                 *float64                                          `json:"total_cost_usd"`
	CostSummary                  domain.CostSummary                                `json:"cost_summary"`
	MeasuredCostUSD              *float64                                          `json:"measured_cost_usd"`
	EstimatedCostUSD             *float64                                          `json:"estimated_cost_usd"`
	TotalEquivalentCostUSD       *float64                                          `json:"total_equivalent_cost_usd"`
	CostProvenance               domain.CostProvenance                             `json:"cost_provenance"`
	EquivalentCostSummary        domain.EquivalentCostSummary                      `json:"equivalent_cost_summary"`
	RoleCostSummaries            map[domain.AgentRole]domain.CostSummary           `json:"role_cost_summaries"`
	ModelCostSummaries           map[string]domain.CostSummary                     `json:"model_cost_summaries"`
	RoleEquivalentCostSummaries  map[domain.AgentRole]domain.EquivalentCostSummary `json:"role_equivalent_cost_summaries"`
	ModelEquivalentCostSummaries map[string]domain.EquivalentCostSummary           `json:"model_equivalent_cost_summaries"`
	TotalInvocations             int64                                             `json:"total_invocations"`
	TotalInputTokens             int64                                             `json:"total_input_tokens"`
	TotalOutputTokens            int64                                             `json:"total_output_tokens"`
	TotalCacheReadTokens         int64                                             `json:"total_cache_read_tokens"`
	OverallCacheHitRatePct       float64                                           `json:"overall_cache_hit_rate_pct"`
	TotalDurationSeconds         int64                                             `json:"total_duration_seconds"`
	FormattedTotalDuration       string                                            `json:"formatted_total_duration"`
	TotalLinesAdded              int64                                             `json:"total_lines_added"`
	TotalLinesDeleted            int64                                             `json:"total_lines_deleted"`
	TotalFilesChanged            int64                                             `json:"total_files_changed"`
	TotalReasoningTokens         int64                                             `json:"total_reasoning_tokens"`
	OverallReasoningRatioPct     float64                                           `json:"overall_reasoning_ratio_pct"`
	RoleCosts                    map[domain.AgentRole]*float64                     `json:"role_costs"`
	ModelCosts                   map[string]*float64                               `json:"model_costs"`
	AnomalySummary               domain.FleetAnomalySummary                        `json:"anomaly_summary"`
}

func FleetSummaryDTOFromDomain(value domain.FleetSummary) FleetSummaryDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.TotalCostUSD, value.TotalInvocations)
	equiv := domain.EffectiveEquivalentCostSummary(value.EquivalentCostSummary, summary, value.TotalCostUSD, value.TotalInvocations)
	roles := effectiveRoleCostSummaries(value.RoleCostSummaries, value.RoleCosts)
	models := effectiveModelCostSummaries(value.ModelCostSummaries, value.ModelCosts)
	roleEquivs := effectiveRoleEquivalentCostSummaries(value.RoleEquivalentCostSummaries, roles)
	modelEquivs := effectiveModelEquivalentCostSummaries(value.ModelEquivalentCostSummaries, models)
	return FleetSummaryDTO{
		ProvisionalSeconds:           value.ProvisionalSeconds,
		SkillAdoption:                value.SkillAdoption,
		TotalTickets:                 value.TotalTickets,
		TotalCostUSD:                 summary.CompleteAmount(),
		CostSummary:                  summary,
		MeasuredCostUSD:              equiv.MeasuredCostUSD,
		EstimatedCostUSD:             equiv.EstimatedCostUSD,
		TotalEquivalentCostUSD:       equiv.TotalEquivalentCostUSD,
		CostProvenance:               equiv.Provenance,
		EquivalentCostSummary:        equiv,
		RoleCostSummaries:            roles,
		ModelCostSummaries:           models,
		RoleEquivalentCostSummaries:  roleEquivs,
		ModelEquivalentCostSummaries: modelEquivs,
		TotalInvocations:             value.TotalInvocations,
		TotalInputTokens:             value.TotalInputTokens,
		TotalOutputTokens:            value.TotalOutputTokens,
		TotalCacheReadTokens:         value.TotalCacheReadTokens,
		OverallCacheHitRatePct:       value.OverallCacheHitRatePct,
		TotalDurationSeconds:         value.TotalDurationSeconds,
		FormattedTotalDuration:       value.FormattedTotalDuration,
		TotalLinesAdded:              value.TotalLinesAdded,
		TotalLinesDeleted:            value.TotalLinesDeleted,
		TotalFilesChanged:            value.TotalFilesChanged,
		TotalReasoningTokens:         value.TotalReasoningTokens,
		OverallReasoningRatioPct:     value.OverallReasoningRatioPct,
		RoleCosts:                    roleCostAmounts(roles),
		ModelCosts:                   modelCostAmounts(models),
		AnomalySummary:               value.AnomalySummary,
	}
}

func effectiveRoleEquivalentCostSummaries(values map[domain.AgentRole]domain.EquivalentCostSummary, summaries map[domain.AgentRole]domain.CostSummary) map[domain.AgentRole]domain.EquivalentCostSummary {
	result := make(map[domain.AgentRole]domain.EquivalentCostSummary, len(summaries))
	for key, value := range values {
		result[key] = domain.FinalizeEquivalentCostSummary(value)
	}
	for key, summary := range summaries {
		if _, exists := result[key]; !exists {
			result[key] = domain.EffectiveEquivalentCostSummary(domain.EquivalentCostSummary{}, summary, summary.Subtotal(), summary.Total)
		}
	}
	return result
}

func effectiveModelEquivalentCostSummaries(values map[string]domain.EquivalentCostSummary, summaries map[string]domain.CostSummary) map[string]domain.EquivalentCostSummary {
	result := make(map[string]domain.EquivalentCostSummary, len(summaries))
	for key, value := range values {
		result[key] = domain.FinalizeEquivalentCostSummary(value)
	}
	for key, summary := range summaries {
		if _, exists := result[key]; !exists {
			result[key] = domain.EffectiveEquivalentCostSummary(domain.EquivalentCostSummary{}, summary, summary.Subtotal(), summary.Total)
		}
	}
	return result
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
