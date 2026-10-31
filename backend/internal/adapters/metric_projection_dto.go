package adapters

import (
	"agenticlens/internal/domain"
)

// GroupedMetricDTO is the concrete nullable cost projection of GroupedMetric.
type GroupedMetricDTO struct {
	Invocations       int64                      `json:"invocations"`
	InputTokens       int64                      `json:"input_tokens"`
	OutputTokens      int64                      `json:"output_tokens"`
	CacheReadTokens   int64                      `json:"cache_read_tokens"`
	CacheHitRatePct   float64                    `json:"cache_hit_rate_pct"`
	CostUSD           *float64                   `json:"cost_usd"`
	CostSummary       domain.CostSummary         `json:"cost_summary"`
	PeakContextTokens int64                      `json:"peak_context_tokens"`
	ContextHealth     domain.ContextHealthStatus `json:"context_health"`
}

func GroupedMetricDTOFromDomain(value domain.GroupedMetric) GroupedMetricDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.CostUSD, value.Invocations)
	return GroupedMetricDTO{
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           summary.CompleteAmount(),
		CostSummary:       summary,
		PeakContextTokens: value.PeakContextTokens,
		ContextHealth:     value.ContextHealth,
	}
}

// RoleMetricDTO is the concrete nullable cost projection of RoleMetric.
type RoleMetricDTO struct {
	Role              domain.AgentRole           `json:"role"`
	Invocations       int64                      `json:"invocations"`
	InputTokens       int64                      `json:"input_tokens"`
	OutputTokens      int64                      `json:"output_tokens"`
	CacheReadTokens   int64                      `json:"cache_read_tokens"`
	CacheHitRatePct   float64                    `json:"cache_hit_rate_pct"`
	CostUSD           *float64                   `json:"cost_usd"`
	CostSummary       domain.CostSummary         `json:"cost_summary"`
	PeakContextTokens int64                      `json:"peak_context_tokens"`
	ContextHealth     domain.ContextHealthStatus `json:"context_health"`
}

func RoleMetricDTOFromDomain(value domain.RoleMetric) RoleMetricDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.CostUSD, value.Invocations)
	return RoleMetricDTO{
		Role:              value.Role,
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           summary.CompleteAmount(),
		CostSummary:       summary,
		PeakContextTokens: value.PeakContextTokens,
		ContextHealth:     value.ContextHealth,
	}
}

func RoleMetricDTOsFromDomain(values []domain.RoleMetric) []RoleMetricDTO {
	if values == nil {
		return nil
	}
	result := make([]RoleMetricDTO, len(values))
	for i, value := range values {
		result[i] = RoleMetricDTOFromDomain(value)
	}
	return result
}

// ModelMetricDTO is the concrete nullable cost projection of ModelMetric.
type ModelMetricDTO struct {
	Model             string                     `json:"model"`
	Invocations       int64                      `json:"invocations"`
	InputTokens       int64                      `json:"input_tokens"`
	OutputTokens      int64                      `json:"output_tokens"`
	CacheReadTokens   int64                      `json:"cache_read_tokens"`
	CacheHitRatePct   float64                    `json:"cache_hit_rate_pct"`
	CostUSD           *float64                   `json:"cost_usd"`
	CostSummary       domain.CostSummary         `json:"cost_summary"`
	PeakContextTokens int64                      `json:"peak_context_tokens"`
	ContextHealth     domain.ContextHealthStatus `json:"context_health"`
}

func ModelMetricDTOFromDomain(value domain.ModelMetric) ModelMetricDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.CostUSD, value.Invocations)
	return ModelMetricDTO{
		Model:             value.Model,
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           summary.CompleteAmount(),
		CostSummary:       summary,
		PeakContextTokens: value.PeakContextTokens,
		ContextHealth:     value.ContextHealth,
	}
}

func ModelMetricDTOsFromDomain(values []domain.ModelMetric) []ModelMetricDTO {
	if values == nil {
		return nil
	}
	result := make([]ModelMetricDTO, len(values))
	for i, value := range values {
		result[i] = ModelMetricDTOFromDomain(value)
	}
	return result
}

// WorkflowEfficiencyDTO is the concrete nullable cost projection of WorkflowEfficiency.
type WorkflowEfficiencyDTO struct {
	SDLCIterations              int64                       `json:"sdlc_iterations"`
	ActiveDurationSeconds       int64                       `json:"active_duration_seconds"`
	ActiveDurationFormatted     string                      `json:"active_duration_formatted"`
	RoleEffortDistributionPct   map[string]*float64         `json:"role_effort_distribution_pct"`
	RoleEffortUnavailableReason string                      `json:"role_effort_unavailable_reason,omitempty"`
	SubagentTurnVelocity        domain.SubagentTurnVelocity `json:"subagent_turn_velocity"`
	TestToProductionRatio       domain.TestToCodeRatio      `json:"test_to_production_ratio"`
	ReasoningTokens             int64                       `json:"reasoning_tokens"`
	ReasoningRatioPct           float64                     `json:"reasoning_ratio_pct"`
	CodeDelivered               domain.CodeDeliveredStats   `json:"code_delivered"`
	ToolExecution               domain.ToolExecutionHealth  `json:"tool_execution"`
	ContextWindowHealth         domain.ContextWindowHealth  `json:"context_window_health"`
}

func WorkflowEfficiencyDTOFromDomain(value domain.WorkflowEfficiency) WorkflowEfficiencyDTO {
	return WorkflowEfficiencyDTO{
		SDLCIterations:              value.SDLCIterations,
		ActiveDurationSeconds:       value.ActiveDurationSeconds,
		ActiveDurationFormatted:     value.ActiveDurationFormatted,
		RoleEffortDistributionPct:   roleEffortAmounts(value.RoleEffortDistributionPct, value.RoleEffortUnavailableReason),
		RoleEffortUnavailableReason: value.RoleEffortUnavailableReason,
		SubagentTurnVelocity:        value.SubagentTurnVelocity,
		TestToProductionRatio:       value.TestToProductionRatio,
		ReasoningTokens:             value.ReasoningTokens,
		ReasoningRatioPct:           value.ReasoningRatioPct,
		CodeDelivered:               value.CodeDelivered,
		ToolExecution:               value.ToolExecution,
		ContextWindowHealth:         value.ContextWindowHealth,
	}
}

func roleEffortAmounts(values map[string]float64, reason string) map[string]*float64 {
	if values == nil {
		return nil
	}
	result := make(map[string]*float64, len(values))
	for key, value := range values {
		if reason != "" {
			result[key] = nil
		} else {
			result[key] = domain.CostNumber(value)
		}
	}
	return result
}
