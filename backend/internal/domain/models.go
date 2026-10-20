package domain

// InvocationTokens stores immutable token counts for a single agent invocation.
type InvocationTokens struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CacheRead   int64 `json:"cache_read"`
	Reasoning   int64 `json:"reasoning"`
	PeakContext int64 `json:"peak_context"`
}

// InvocationFact represents an atomic execution record for an agent invocation.
type InvocationFact struct {
	ID                    string           `json:"id"`
	ParentID              *string          `json:"parent_id"`
	Role                  string           `json:"role"`
	Model                 string           `json:"model"`
	User                  string           `json:"user"`
	StartedAt             *string          `json:"started_at"`
	CompletedAt           *string          `json:"completed_at"`
	ActiveDurationSeconds int64            `json:"active_duration_seconds"`
	Turns                 int64            `json:"turns"`
	Tokens                InvocationTokens `json:"tokens"`
	CostUSD               float64          `json:"cost_usd"`
	Compactions           int64            `json:"compactions"`
}

// ToolFact stores tool usage execution counts and errors.
type ToolFact struct {
	Name   string `json:"name"`
	Calls  int64  `json:"calls"`
	Errors int64  `json:"errors"`
}

// EventFact captures a single discrete, ordered lifecycle occurrence in an agent session.
type EventFact struct {
	Timestamp        string    `json:"timestamp"`
	SessionID        string    `json:"session_id"`
	Turn             int64     `json:"turn"`
	Role             string    `json:"role"`
	EventType        EventType `json:"event_type"`
	ToolName         *string   `json:"tool_name,omitempty"`
	Status           string    `json:"status"`
	ErrorMessage     *string   `json:"error_message,omitempty"`
	DurationMs       *int64    `json:"duration_ms,omitempty"`
	ContextTokens    int64     `json:"context_tokens"`
	PromptContent    *string   `json:"prompt_content,omitempty"`
	ReasoningContent *string   `json:"reasoning_content,omitempty"`
	ToolInput        *string   `json:"tool_input,omitempty"`
	ToolOutput       *string   `json:"tool_output,omitempty"`
}

// WorkspaceEventsResponse wraps execution events for a ticket workspace with hydration status and data source provenance.
type WorkspaceEventsResponse struct {
	TicketID   string      `json:"ticket_id"`
	IsHydrated bool        `json:"is_hydrated"`
	Events     []EventFact `json:"events"`
	DataSource string      `json:"data_source"`
}

// CodeFacts stores lines of code modified and SDLC iteration counts.
type CodeFacts struct {
	LinesAdded      int64 `json:"lines_added"`
	LinesDeleted    int64 `json:"lines_deleted"`
	TestLinesAdded  int64 `json:"test_lines_added"`
	TestLinesDeleted int64 `json:"test_lines_deleted"`
	ProdLinesAdded  int64 `json:"prod_lines_added"`
	ProdLinesDeleted int64 `json:"prod_lines_deleted"`
	FilesChanged    int64 `json:"files_changed"`
	SDLCIterations  int64 `json:"sdlc_iterations"`
}

// MigrationInfo captures migration state, completed steps, pending requirements, and audit metadata.
type MigrationInfo struct {
	SchemaVersion  int             `json:"schema_version"`
	Status         MigrationStatus `json:"status"`
	StepsCompleted []string        `json:"steps_completed"`
	StepsPending   []string        `json:"steps_pending"`
	LastMigratedAt string          `json:"last_migrated_at"`
	LastMigratedBy string          `json:"last_migrated_by"`
	Notes          string          `json:"notes,omitempty"`
}

// AgenticFactLedger is the root raw session facts ledger extracted for a ticket workspace.
type AgenticFactLedger struct {
	SchemaVersion *int             `json:"schema_version,omitempty"`
	Migration     *MigrationInfo   `json:"migration,omitempty"`
	TicketID      *string          `json:"ticket_id"`
	TicketTitle   *string          `json:"ticket_title"`
	TicketType    *string          `json:"ticket_type"`
	StartedAt     *string          `json:"started_at"`
	CompletedAt   *string          `json:"completed_at"`
	UpdatedAt     string           `json:"updated_at"`
	Code          CodeFacts        `json:"code"`
	Invocations   []InvocationFact `json:"invocations"`
	Tools         []ToolFact       `json:"tools"`
	Events        []EventFact      `json:"events,omitempty"`
}

// GroupedMetric aggregates metrics across invocations.
type GroupedMetric struct {
	Invocations       int64   `json:"invocations"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	CacheHitRatePct   float64 `json:"cache_hit_rate_pct"`
	CostUSD           float64 `json:"cost_usd"`
	PeakContextTokens int64   `json:"peak_context_tokens"`
	ContextHealth     string  `json:"context_health"`
}

// TestToCodeRatio captures test to production code lines added and ratio string.
type TestToCodeRatio struct {
	Ratio          string `json:"ratio"`
	TestLinesAdded int64  `json:"test_lines_added"`
	ProdLinesAdded int64  `json:"prod_lines_added"`
}

// SubagentTurnVelocity records turn efficiency and thrashing for subagents.
type SubagentTurnVelocity struct {
	AvgTurnsPerInvocation float64            `json:"avg_turns_per_invocation"`
	PeakTurns             int64              `json:"peak_turns"`
	ByRole                map[string]float64 `json:"by_role"`
}

// CodeDeliveredStats summarizes delivered code lines and file count.
type CodeDeliveredStats struct {
	LinesAdded   int64 `json:"lines_added"`
	LinesDeleted int64 `json:"lines_deleted"`
	FilesChanged int64 `json:"files_changed"`
}

// ToolExecutionHealth summarizes tool call volume, corrections, and success rate.
type ToolExecutionHealth struct {
	TotalCalls           int64   `json:"total_calls"`
	ErrorsAndCorrections int64   `json:"errors_and_corrections"`
	SuccessRatePct       float64 `json:"success_rate_pct"`
}

// ContextWindowHealth captures context saturation, ceiling thresholds, and peak token tracking.
type ContextWindowHealth struct {
	HealthyCeilingTokens      int64            `json:"healthy_ceiling_tokens"`
	Status                    string           `json:"status"`
	OrchestratorPeakTokens    int64            `json:"orchestrator_peak_tokens"`
	OrchestratorSaturationPct float64          `json:"orchestrator_saturation_pct"`
	SubagentsPeakTokens       int64            `json:"subagents_peak_tokens"`
	SubagentsPeakRole         string           `json:"subagents_peak_role"`
	Compactions               int64            `json:"compactions"`
	PeakByRole                map[string]int64 `json:"peak_by_role"`
}

// WorkflowEfficiency aggregates SDLC duration, code delivery, and operational health metrics.
type WorkflowEfficiency struct {
	SDLCIterations            int64                `json:"sdlc_iterations"`
	ActiveDurationSeconds     int64                `json:"active_duration_seconds"`
	ActiveDurationFormatted   string               `json:"active_duration_formatted"`
	RoleEffortDistributionPct map[string]float64   `json:"role_effort_distribution_pct"`
	SubagentTurnVelocity      SubagentTurnVelocity `json:"subagent_turn_velocity"`
	TestToProductionRatio     TestToCodeRatio      `json:"test_to_production_ratio"`
	ReasoningTokens           int64                `json:"reasoning_tokens"`
	ReasoningRatioPct         float64              `json:"reasoning_ratio_pct"`
	CodeDelivered             CodeDeliveredStats   `json:"code_delivered"`
	ToolExecution             ToolExecutionHealth  `json:"tool_execution"`
	ContextWindowHealth       ContextWindowHealth  `json:"context_window_health"`
}

// ModelMetric summarizes aggregated usage and performance for a single model.
type ModelMetric struct {
	Model             string  `json:"model"`
	Invocations       int64   `json:"invocations"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	CacheHitRatePct   float64 `json:"cache_hit_rate_pct"`
	CostUSD           float64 `json:"cost_usd"`
	PeakContextTokens int64   `json:"peak_context_tokens"`
	ContextHealth     string  `json:"context_health"`
}

// RoleMetric summarizes aggregated usage and performance for an agent role.
type RoleMetric struct {
	Role              string  `json:"role"`
	Invocations       int64   `json:"invocations"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	CacheHitRatePct   float64 `json:"cache_hit_rate_pct"`
	CostUSD           float64 `json:"cost_usd"`
	PeakContextTokens int64   `json:"peak_context_tokens"`
	ContextHealth     string  `json:"context_health"`
}

// ThreadRoleMetric groups model metrics within an agent role under a thread session.
type ThreadRoleMetric struct {
	Role   string        `json:"role"`
	Models []ModelMetric `json:"models"`
}

// ThreadExecutionRecord groups invocations by root session into a hierarchical execution tree.
type ThreadExecutionRecord struct {
	SessionID   string             `json:"session_id"`
	User        string             `json:"user"`
	Roles       []ThreadRoleMetric `json:"roles"`
	Compactions int64              `json:"compactions"`
}

// AgenticMetricsLedger is the full projected metrics ledger combining raw facts with derived rollups.
type AgenticMetricsLedger struct {
	TicketID           *string                 `json:"ticket_id"`
	TicketTitle        *string                 `json:"ticket_title"`
	TicketType         *string                 `json:"ticket_type"`
	StartedAt          *string                 `json:"started_at"`
	CompletedAt        *string                 `json:"completed_at"`
	UpdatedAt          string                  `json:"updated_at"`
	Facts              AgenticFactLedger       `json:"facts"`
	WorkflowEfficiency WorkflowEfficiency      `json:"workflow_efficiency"`
	Threads            []ThreadExecutionRecord `json:"threads"`
	Totals             GroupedMetric           `json:"totals"`
	Roles              []RoleMetric            `json:"roles"`
	Models             []ModelMetric           `json:"models"`
	Anomalies          []Anomaly               `json:"anomalies,omitempty"`
}

// TimeBucketMetric captures telemetry aggregates for a chronological time bucket.
type TimeBucketMetric struct {
	BucketKey               string  `json:"bucket_key"`
	BucketDate              string  `json:"bucket_date"`
	Invocations             int64   `json:"invocations"`
	CostUSD                 float64 `json:"cost_usd"`
	InputTokens             int64   `json:"input_tokens"`
	OutputTokens            int64   `json:"output_tokens"`
	CacheReadTokens         int64   `json:"cache_read_tokens"`
	ReasoningTokens         int64   `json:"reasoning_tokens"`
	CacheHitRatePct         float64 `json:"cache_hit_rate_pct"`
	LinesAdded              int64   `json:"lines_added"`
	LinesDeleted            int64   `json:"lines_deleted"`
	FilesChanged            int64   `json:"files_changed"`
	NetLinesAdded           int64   `json:"net_lines_added"`
	PeakContextTokens       int64   `json:"peak_context_tokens"`
	ActiveDurationSeconds   int64   `json:"active_duration_seconds"`
	CumulativeCostUSD       float64 `json:"cumulative_cost_usd"`
	CumulativeInputTokens   int64   `json:"cumulative_input_tokens"`
	CumulativeOutputTokens  int64   `json:"cumulative_output_tokens"`
	CumulativeLinesAdded    int64   `json:"cumulative_lines_added"`
	CumulativeLinesDeleted  int64   `json:"cumulative_lines_deleted"`
	CumulativeNetLinesAdded int64   `json:"cumulative_net_lines_added"`
	CumulativeInvocations   int64   `json:"cumulative_invocations"`
}

// BrittleToolSummary represents aggregated failure and anomaly metrics for a tool across the fleet.
type BrittleToolSummary struct {
	ToolName       string  `json:"tool_name"`
	AnomalyCount   int64   `json:"anomaly_count"`
	TotalCalls     int64   `json:"total_calls"`
	ErrorCount     int64   `json:"error_count"`
	FailureRatePct float64 `json:"failure_rate_pct"`
}

// FleetAnomalySummary aggregates anomaly patterns, severity counts, and brittle tools across the fleet.
type FleetAnomalySummary struct {
	TotalAnomalies       int64                `json:"total_anomalies"`
	WarningCount         int64                `json:"warning_count"`
	CriticalCount        int64                `json:"critical_count"`
	SeverityCounts       map[string]int64     `json:"severity_counts"`
	RuleCounts           map[string]int64     `json:"rule_counts"`
	TotalTickets         int64                `json:"total_tickets"`
	TicketsWithAnomalies int64                `json:"tickets_with_anomalies"`
	AnomalyRatePerTicket float64              `json:"anomaly_rate_per_ticket"`
	BrittleTools         []BrittleToolSummary `json:"brittle_tools"`
}

// FleetSummary aggregates high-level metrics across multiple tickets.
type FleetSummary struct {
	TotalTickets             int64               `json:"total_tickets"`
	TotalCostUSD             float64             `json:"total_cost_usd"`
	TotalInvocations         int64               `json:"total_invocations"`
	TotalInputTokens         int64               `json:"total_input_tokens"`
	TotalOutputTokens        int64               `json:"total_output_tokens"`
	TotalCacheReadTokens     int64               `json:"total_cache_read_tokens"`
	OverallCacheHitRatePct   float64             `json:"overall_cache_hit_rate_pct"`
	TotalDurationSeconds     int64               `json:"total_duration_seconds"`
	FormattedTotalDuration   string              `json:"formatted_total_duration"`
	TotalLinesAdded          int64               `json:"total_lines_added"`
	TotalLinesDeleted        int64               `json:"total_lines_deleted"`
	TotalFilesChanged        int64               `json:"total_files_changed"`
	TotalReasoningTokens     int64               `json:"total_reasoning_tokens"`
	OverallReasoningRatioPct float64             `json:"overall_reasoning_ratio_pct"`
	RoleCosts                map[string]float64  `json:"role_costs"`
	ModelCosts               map[string]float64  `json:"model_costs"`
	AnomalySummary           FleetAnomalySummary `json:"anomaly_summary"`
}

// TicketComparisonRow summarizes a single ticket for cross-ticket matrix presentation.
type TicketComparisonRow struct {
	TicketID                  string  `json:"ticket_id"`
	UpdatedAt                 string  `json:"updated_at"`
	SDLCIterations            int64   `json:"sdlc_iterations"`
	ActiveDurationSeconds     int64   `json:"active_duration_seconds"`
	ActiveDurationFormatted   string  `json:"active_duration_formatted"`
	CostUSD                   float64 `json:"cost_usd"`
	CacheHitRatePct           float64 `json:"cache_hit_rate_pct"`
	TestToCodeRatio           string  `json:"test_to_code_ratio"`
	LinesAdded                int64   `json:"lines_added"`
	LinesDeleted              int64   `json:"lines_deleted"`
	FilesChanged              int64   `json:"files_changed"`
	ToolSuccessRatePct        float64 `json:"tool_success_rate_pct"`
	ContextHealthStatus       string  `json:"context_health_status"`
	OrchestratorSaturationPct float64 `json:"orchestrator_saturation_pct"`
	PeakContextTokens         int64   `json:"peak_context_tokens"`
	TicketTitle               string  `json:"ticket_title"`
	TicketType                *string `json:"ticket_type,omitempty"`
	StartedAt                 *string `json:"started_at"`
	CompletedAt               *string `json:"completed_at"`
	LeadDurationSeconds       *int64  `json:"lead_duration_seconds"`
	LeadDurationFormatted     string  `json:"lead_duration_formatted"`
}

// ScanReport holds all ledgers loaded from workspace traversal alongside non-fatal scan errors.
type ScanReport struct {
	Ledgers []AgenticMetricsLedger `json:"ledgers"`
	Errors  []string               `json:"errors"`
}

// TicketArtifacts represents markdown documentation artifacts for a ticket workspace.
type TicketArtifacts struct {
	TicketID        string  `json:"ticket_id"`
	ContractContent *string `json:"contract_content"`
	SummaryContent  *string `json:"summary_content"`
}

// WorkspaceDetail combines an AgenticMetricsLedger with its associated TicketArtifacts.
type WorkspaceDetail struct {
	Ledger    AgenticMetricsLedger `json:"ledger"`
	Artifacts TicketArtifacts      `json:"artifacts"`
}
