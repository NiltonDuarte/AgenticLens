package adapters

import "agenticlens/internal/domain"

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

func (value GroupedMetricDTO) ToDomain() domain.GroupedMetric {
	return domain.GroupedMetric{
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           costAccumulator(value.CostUSD, value.CostSummary),
		CostSummary:       value.CostSummary,
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

func (value RoleMetricDTO) ToDomain() domain.RoleMetric {
	return domain.RoleMetric{
		Role:              value.Role,
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           costAccumulator(value.CostUSD, value.CostSummary),
		CostSummary:       value.CostSummary,
		PeakContextTokens: value.PeakContextTokens,
		ContextHealth:     value.ContextHealth,
	}
}

func RoleMetricDTOsToDomain(values []RoleMetricDTO) []domain.RoleMetric {
	if values == nil {
		return nil
	}
	result := make([]domain.RoleMetric, len(values))
	for i, value := range values {
		result[i] = value.ToDomain()
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

func (value ModelMetricDTO) ToDomain() domain.ModelMetric {
	return domain.ModelMetric{
		Model:             value.Model,
		Invocations:       value.Invocations,
		InputTokens:       value.InputTokens,
		OutputTokens:      value.OutputTokens,
		CacheReadTokens:   value.CacheReadTokens,
		CacheHitRatePct:   value.CacheHitRatePct,
		CostUSD:           costAccumulator(value.CostUSD, value.CostSummary),
		CostSummary:       value.CostSummary,
		PeakContextTokens: value.PeakContextTokens,
		ContextHealth:     value.ContextHealth,
	}
}

func ModelMetricDTOsToDomain(values []ModelMetricDTO) []domain.ModelMetric {
	if values == nil {
		return nil
	}
	result := make([]domain.ModelMetric, len(values))
	for i, value := range values {
		result[i] = value.ToDomain()
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

func (value WorkflowEfficiencyDTO) ToDomain() domain.WorkflowEfficiency {
	return domain.WorkflowEfficiency{
		SDLCIterations:              value.SDLCIterations,
		ActiveDurationSeconds:       value.ActiveDurationSeconds,
		ActiveDurationFormatted:     value.ActiveDurationFormatted,
		RoleEffortDistributionPct:   roleEffortAccumulators(value.RoleEffortDistributionPct),
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

// ThreadRoleMetricDTO is the concrete nullable cost projection of ThreadRoleMetric.
type ThreadRoleMetricDTO struct {
	Role   domain.AgentRole `json:"role"`
	Models []ModelMetricDTO `json:"models"`
}

func ThreadRoleMetricDTOFromDomain(value domain.ThreadRoleMetric) ThreadRoleMetricDTO {
	return ThreadRoleMetricDTO{
		Role:   value.Role,
		Models: ModelMetricDTOsFromDomain(value.Models),
	}
}

func ThreadRoleMetricDTOsFromDomain(values []domain.ThreadRoleMetric) []ThreadRoleMetricDTO {
	if values == nil {
		return nil
	}
	result := make([]ThreadRoleMetricDTO, len(values))
	for i, value := range values {
		result[i] = ThreadRoleMetricDTOFromDomain(value)
	}
	return result
}

// ThreadExecutionRecordDTO is the concrete nullable cost projection of ThreadExecutionRecord.
type ThreadExecutionRecordDTO struct {
	SessionID   string                `json:"session_id"`
	User        string                `json:"user"`
	Roles       []ThreadRoleMetricDTO `json:"roles"`
	Compactions int64                 `json:"compactions"`
}

func ThreadExecutionRecordDTOFromDomain(value domain.ThreadExecutionRecord) ThreadExecutionRecordDTO {
	return ThreadExecutionRecordDTO{
		SessionID:   value.SessionID,
		User:        value.User,
		Roles:       ThreadRoleMetricDTOsFromDomain(value.Roles),
		Compactions: value.Compactions,
	}
}

func ThreadExecutionRecordDTOsFromDomain(values []domain.ThreadExecutionRecord) []ThreadExecutionRecordDTO {
	if values == nil {
		return nil
	}
	result := make([]ThreadExecutionRecordDTO, len(values))
	for i, value := range values {
		result[i] = ThreadExecutionRecordDTOFromDomain(value)
	}
	return result
}

// SessionFactResponseDTO is the concrete nullable cost projection of SessionFact.
type SessionFactResponseDTO struct {
	ID          string                    `json:"id"`
	ParentID    *string                   `json:"parent_id,omitempty"`
	Source      domain.SourceHarness      `json:"source,omitempty"`
	User        string                    `json:"user,omitempty"`
	Invocations []InvocationFactDTO       `json:"invocations"`
	Tools       []domain.SessionToolFact  `json:"tools"`
	Events      []domain.SessionEventFact `json:"events,omitempty"`
	Anomalies   []domain.Anomaly          `json:"anomalies"`
}

func SessionFactResponseDTOFromDomain(value domain.SessionFact) SessionFactResponseDTO {
	return SessionFactResponseDTO{
		ID:          value.ID,
		ParentID:    value.ParentID,
		Source:      value.Source,
		User:        value.User,
		Invocations: InvocationFactDTOsFromDomain(value.Invocations),
		Tools:       value.Tools,
		Events:      value.Events,
		Anomalies:   value.Anomalies,
	}
}

func SessionFactResponseDTOsFromDomain(values []domain.SessionFact) []SessionFactResponseDTO {
	if values == nil {
		return nil
	}
	result := make([]SessionFactResponseDTO, len(values))
	for i, value := range values {
		result[i] = SessionFactResponseDTOFromDomain(value)
	}
	return result
}

// FactLedgerResponseDTO is the concrete nullable cost projection of AgenticFactLedger.
type FactLedgerResponseDTO struct {
	SchemaVersion *int                           `json:"schema_version,omitempty"`
	Migration     *domain.MigrationInfo          `json:"migration,omitempty"`
	TicketID      *string                        `json:"ticket_id"`
	TicketTitle   *string                        `json:"ticket_title"`
	TicketType    *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt     *string                        `json:"started_at"`
	CompletedAt   *string                        `json:"completed_at"`
	UpdatedAt     string                         `json:"updated_at"`
	Code          domain.CodeFacts               `json:"code"`
	Sessions      []SessionFactResponseDTO       `json:"sessions"`
	Anomalies     []domain.Anomaly               `json:"anomalies"`
	Skills        []domain.SkillFact             `json:"skills,omitempty"`
	TaskLinks     []TaskLinkFactDTO              `json:"task_links,omitempty"`
}

func FactLedgerResponseDTOFromDomain(value domain.AgenticFactLedger) FactLedgerResponseDTO {
	return FactLedgerResponseDTO{
		SchemaVersion: value.SchemaVersion,
		Migration:     value.Migration,
		TicketID:      value.TicketID,
		TicketTitle:   value.TicketTitle,
		TicketType:    value.TicketType,
		StartedAt:     value.StartedAt,
		CompletedAt:   value.CompletedAt,
		UpdatedAt:     value.UpdatedAt,
		Code:          value.Code,
		Sessions:      SessionFactResponseDTOsFromDomain(value.Sessions),
		Anomalies:     value.Anomalies,
		Skills:        value.Skills,
		TaskLinks:     TaskLinkFactDTOsFromDomain(value.TaskLinks),
	}
}

// MetricsLedgerResponseDTO is the concrete nullable cost projection of AgenticMetricsLedger.
type MetricsLedgerResponseDTO struct {
	TicketID           *string                        `json:"ticket_id"`
	TicketTitle        *string                        `json:"ticket_title"`
	TicketType         *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt          *string                        `json:"started_at"`
	CompletedAt        *string                        `json:"completed_at"`
	UpdatedAt          string                         `json:"updated_at"`
	Facts              FactLedgerResponseDTO          `json:"facts"`
	WorkflowEfficiency WorkflowEfficiencyDTO          `json:"workflow_efficiency"`
	Threads            []ThreadExecutionRecordDTO     `json:"threads"`
	Totals             GroupedMetricDTO               `json:"totals"`
	Roles              []RoleMetricDTO                `json:"roles"`
	Models             []ModelMetricDTO               `json:"models"`
	Anomalies          []domain.Anomaly               `json:"anomalies,omitempty"`
}

func MetricsLedgerResponseDTOFromDomain(value domain.AgenticMetricsLedger) MetricsLedgerResponseDTO {
	return MetricsLedgerResponseDTO{
		TicketID:           value.TicketID,
		TicketTitle:        value.TicketTitle,
		TicketType:         value.TicketType,
		StartedAt:          value.StartedAt,
		CompletedAt:        value.CompletedAt,
		UpdatedAt:          value.UpdatedAt,
		Facts:              FactLedgerResponseDTOFromDomain(value.Facts),
		WorkflowEfficiency: WorkflowEfficiencyDTOFromDomain(value.WorkflowEfficiency),
		Threads:            ThreadExecutionRecordDTOsFromDomain(value.Threads),
		Totals:             GroupedMetricDTOFromDomain(value.Totals),
		Roles:              RoleMetricDTOsFromDomain(value.Roles),
		Models:             ModelMetricDTOsFromDomain(value.Models),
		Anomalies:          value.Anomalies,
	}
}

func MetricsLedgerResponseDTOsFromDomain(values []domain.AgenticMetricsLedger) []MetricsLedgerResponseDTO {
	if values == nil {
		return nil
	}
	result := make([]MetricsLedgerResponseDTO, len(values))
	for i, value := range values {
		result[i] = MetricsLedgerResponseDTOFromDomain(value)
	}
	return result
}

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
	return FleetSummaryDTO{
		SkillAdoption:            value.SkillAdoption,
		TotalTickets:             value.TotalTickets,
		TotalCostUSD:             summary.CompleteAmount(),
		CostSummary:              summary,
		RoleCostSummaries:        effectiveRoleCostSummaries(value.RoleCostSummaries, value.RoleCosts),
		ModelCostSummaries:       effectiveModelCostSummaries(value.ModelCostSummaries, value.ModelCosts),
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
		RoleCosts:                roleCostAmounts(effectiveRoleCostSummaries(value.RoleCostSummaries, value.RoleCosts)),
		ModelCosts:               modelCostAmounts(effectiveModelCostSummaries(value.ModelCostSummaries, value.ModelCosts)),
		AnomalySummary:           value.AnomalySummary,
	}
}

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
		CumulativeCostUSD:       domain.EffectiveCostSummary(value.CumulativeCostSummary, value.CumulativeCostUSD, value.CumulativeInvocations).CompleteAmount(),
		CumulativeCostSummary:   domain.EffectiveCostSummary(value.CumulativeCostSummary, value.CumulativeCostUSD, value.CumulativeInvocations),
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

func costAccumulator(amount *float64, summary domain.CostSummary) float64 {
	if summary.KnownSubtotalUSD != nil {
		return *summary.KnownSubtotalUSD
	}
	if amount != nil {
		return *amount
	}
	return 0
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
func roleEffortAccumulators(values map[string]*float64) map[string]float64 {
	if values == nil {
		return nil
	}
	result := make(map[string]float64, len(values))
	for key, value := range values {
		if value != nil {
			result[key] = *value
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
