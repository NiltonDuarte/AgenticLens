package http

import (
	"agenticlens/internal/domain"
)

// TicketComparisonRowDTO is the concrete nullable cost projection of TicketComparisonRow.
type TicketComparisonRowDTO struct {
	SkillsUsed              []string                       `json:"skills_used"`
	TicketID                string                         `json:"ticket_id"`
	UpdatedAt               string                         `json:"updated_at"`
	SDLCIterations          int64                          `json:"sdlc_iterations"`
	ActiveDurationSeconds   int64                          `json:"active_duration_seconds"`
	ActiveDurationFormatted string                         `json:"active_duration_formatted"`
	CostUSD                 *float64                       `json:"cost_usd"`
	CostSummary             domain.CostSummary             `json:"cost_summary"`
	CacheHitRatePct         float64                        `json:"cache_hit_rate_pct"`
	TestToCodeRatio         string                         `json:"test_to_code_ratio"`
	LinesAdded              int64                          `json:"lines_added"`
	LinesDeleted            int64                          `json:"lines_deleted"`
	FilesChanged            int64                          `json:"files_changed"`
	ToolSuccessRatePct      float64                        `json:"tool_success_rate_pct"`
	ContextHealthStatus     domain.ContextHealthStatus     `json:"context_health_status"`
	RootSaturationPct       float64                        `json:"root_saturation_pct"`
	PeakContextTokens       int64                          `json:"peak_context_tokens"`
	TicketTitle             string                         `json:"ticket_title"`
	TicketType              *domain.ConventionalTicketType `json:"ticket_type,omitempty"`
	StartedAt               *string                        `json:"started_at"`
	CompletedAt             *string                        `json:"completed_at"`
	LeadDurationSeconds     *int64                         `json:"lead_duration_seconds"`
	LeadDurationFormatted   string                         `json:"lead_duration_formatted"`
}

func TicketComparisonRowDTOFromDomain(value domain.TicketComparisonRow) TicketComparisonRowDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.CostUSD, int64(0))
	return TicketComparisonRowDTO{
		SkillsUsed:              value.SkillsUsed,
		TicketID:                value.TicketID,
		UpdatedAt:               value.UpdatedAt,
		SDLCIterations:          value.SDLCIterations,
		ActiveDurationSeconds:   value.ActiveDurationSeconds,
		ActiveDurationFormatted: value.ActiveDurationFormatted,
		CostUSD:                 summary.CompleteAmount(),
		CostSummary:             summary,
		CacheHitRatePct:         value.CacheHitRatePct,
		TestToCodeRatio:         value.TestToCodeRatio,
		LinesAdded:              value.LinesAdded,
		LinesDeleted:            value.LinesDeleted,
		FilesChanged:            value.FilesChanged,
		ToolSuccessRatePct:      value.ToolSuccessRatePct,
		ContextHealthStatus:     value.ContextHealthStatus,
		RootSaturationPct:       value.RootSaturationPct,
		PeakContextTokens:       value.PeakContextTokens,
		TicketTitle:             value.TicketTitle,
		TicketType:              value.TicketType,
		StartedAt:               value.StartedAt,
		CompletedAt:             value.CompletedAt,
		LeadDurationSeconds:     value.LeadDurationSeconds,
		LeadDurationFormatted:   value.LeadDurationFormatted,
	}
}

func TicketComparisonRowDTOsFromDomain(values []domain.TicketComparisonRow) []TicketComparisonRowDTO {
	if values == nil {
		return nil
	}
	result := make([]TicketComparisonRowDTO, len(values))
	for i, value := range values {
		result[i] = TicketComparisonRowDTOFromDomain(value)
	}
	return result
}
