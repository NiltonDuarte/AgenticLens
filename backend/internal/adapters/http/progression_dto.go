package http

import (
	"agenticlens/internal/domain"
)

// TimeBucketMetricDTO is the concrete nullable cost projection of TimeBucketMetric.
type TimeBucketMetricDTO struct {
	BucketKey               string             `json:"bucket_key"`
	BucketDate              string             `json:"bucket_date"`
	Invocations             int64              `json:"invocations"`
	CostUSD                 *float64           `json:"cost_usd"`
	CostSummary             domain.CostSummary `json:"cost_summary"`
	InputTokens             int64              `json:"input_tokens"`
	OutputTokens            int64              `json:"output_tokens"`
	CacheReadTokens         int64              `json:"cache_read_tokens"`
	ReasoningTokens         int64              `json:"reasoning_tokens"`
	CacheHitRatePct         float64            `json:"cache_hit_rate_pct"`
	LinesAdded              int64              `json:"lines_added"`
	LinesDeleted            int64              `json:"lines_deleted"`
	FilesChanged            int64              `json:"files_changed"`
	NetLinesAdded           int64              `json:"net_lines_added"`
	PeakContextTokens       int64              `json:"peak_context_tokens"`
	ActiveDurationSeconds   int64              `json:"active_duration_seconds"`
	CumulativeCostUSD       *float64           `json:"cumulative_cost_usd"`
	CumulativeCostSummary   domain.CostSummary `json:"cumulative_cost_summary"`
	CumulativeInputTokens   int64              `json:"cumulative_input_tokens"`
	CumulativeOutputTokens  int64              `json:"cumulative_output_tokens"`
	CumulativeLinesAdded    int64              `json:"cumulative_lines_added"`
	CumulativeLinesDeleted  int64              `json:"cumulative_lines_deleted"`
	CumulativeNetLinesAdded int64              `json:"cumulative_net_lines_added"`
	CumulativeInvocations   int64              `json:"cumulative_invocations"`
}

func TimeBucketMetricDTOFromDomain(value domain.TimeBucketMetric) TimeBucketMetricDTO {
	summary := domain.EffectiveCostSummary(value.CostSummary, value.CostUSD, value.Invocations)
	cumulative := domain.EffectiveCostSummary(value.CumulativeCostSummary, value.CumulativeCostUSD, value.CumulativeInvocations)
	return TimeBucketMetricDTO{
		BucketKey:               value.BucketKey,
		BucketDate:              value.BucketDate,
		Invocations:             value.Invocations,
		CostUSD:                 summary.CompleteAmount(),
		CostSummary:             summary,
		InputTokens:             value.InputTokens,
		OutputTokens:            value.OutputTokens,
		CacheReadTokens:         value.CacheReadTokens,
		ReasoningTokens:         value.ReasoningTokens,
		CacheHitRatePct:         value.CacheHitRatePct,
		LinesAdded:              value.LinesAdded,
		LinesDeleted:            value.LinesDeleted,
		FilesChanged:            value.FilesChanged,
		NetLinesAdded:           value.NetLinesAdded,
		PeakContextTokens:       value.PeakContextTokens,
		ActiveDurationSeconds:   value.ActiveDurationSeconds,
		CumulativeCostUSD:       cumulative.CompleteAmount(),
		CumulativeCostSummary:   cumulative,
		CumulativeInputTokens:   value.CumulativeInputTokens,
		CumulativeOutputTokens:  value.CumulativeOutputTokens,
		CumulativeLinesAdded:    value.CumulativeLinesAdded,
		CumulativeLinesDeleted:  value.CumulativeLinesDeleted,
		CumulativeNetLinesAdded: value.CumulativeNetLinesAdded,
		CumulativeInvocations:   value.CumulativeInvocations,
	}
}

func TimeBucketMetricDTOsFromDomain(values []domain.TimeBucketMetric) []TimeBucketMetricDTO {
	if values == nil {
		return nil
	}
	result := make([]TimeBucketMetricDTO, len(values))
	for i, value := range values {
		result[i] = TimeBucketMetricDTOFromDomain(value)
	}
	return result
}
