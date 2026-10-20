export type ContextHealthStatus = 'HEALTHY' | 'ELEVATED' | 'BLOATED' | 'CRITICAL';

export type AgentRole = string;

export type MigrationStatus = 'pending' | 'partial' | 'completed';

export interface MigrationInfo {
  readonly schema_version: number;
  readonly status: MigrationStatus;
  readonly steps_completed: readonly string[];
  readonly steps_pending: readonly string[];
  readonly last_migrated_at: string;
  readonly last_migrated_by: string;
  readonly notes?: string;
}

export const CANONICAL_ROLE_ORDER: readonly string[] = [
  'orchestrator',
  'planner',
  'coder',
  'tester',
  'reviewer',
];

export const SUBAGENT_ROLE_ORDER: readonly string[] = [
  'planner',
  'coder',
  'tester',
  'reviewer',
];

export function isCanonicalRole(role: string): boolean {
  return CANONICAL_ROLE_ORDER.includes(role.toLowerCase().trim());
}

export interface InvocationTokens {
  readonly input: number;
  readonly output: number;
  readonly cache_read: number;
  readonly reasoning: number;
  readonly peak_context: number;
}

export interface InvocationFact {
  readonly id: string;
  readonly parent_id: string | null;
  readonly role: string;
  readonly model: string;
  readonly user: string;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly active_duration_seconds: number;
  readonly turns: number;
  readonly tokens: InvocationTokens;
  readonly cost_usd: number;
  readonly compactions: number;
}

export interface ToolFact {
  readonly name: string;
  readonly calls: number;
  readonly errors: number;
}

export type {
  EventType,
  EventFact,
  DataSource,
  WorkspaceEventsResponse,
  AnomalySeverity,
  AnomalyRuleID,
  AnomalyContext,
  Anomaly,
  BrittleToolSummary,
  FleetAnomalySummary,
} from '../types/telemetry';

import type {
  EventFact,
  Anomaly,
  FleetAnomalySummary,
} from '../types/telemetry';

export interface CodeFacts {
  readonly lines_added: number;
  readonly lines_deleted: number;
  readonly test_lines_added: number;
  readonly test_lines_deleted: number;
  readonly prod_lines_added: number;
  readonly prod_lines_deleted: number;
  readonly files_changed: number;
  readonly sdlc_iterations: number;
}

export interface AgenticFactLedger {
  readonly schema_version?: number;
  readonly migration?: MigrationInfo;
  readonly ticket_id: string | null;
  readonly ticket_title: string | null;
  readonly ticket_type: string | null;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly updated_at: string;
  readonly code: CodeFacts;
  readonly invocations: readonly InvocationFact[];
  readonly tools: readonly ToolFact[];
  readonly events?: readonly EventFact[];
}

export interface GroupedMetric {
  readonly invocations: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly cost_usd: number;
  readonly peak_context_tokens: number;
  readonly context_health: string;
}

export interface TestToCodeRatio {
  readonly ratio: string;
  readonly test_lines_added: number;
  readonly prod_lines_added: number;
}

export interface SubagentTurnVelocity {
  readonly avg_turns_per_invocation: number;
  readonly peak_turns: number;
  readonly by_role: Readonly<Record<string, number>>;
}

export interface CodeDeliveredStats {
  readonly lines_added: number;
  readonly lines_deleted: number;
  readonly files_changed: number;
}

export interface ToolExecutionHealth {
  readonly total_calls: number;
  readonly errors_and_corrections: number;
  readonly success_rate_pct: number;
}

export interface ContextWindowHealth {
  readonly healthy_ceiling_tokens: number;
  readonly status: string;
  readonly orchestrator_peak_tokens: number;
  readonly orchestrator_saturation_pct: number;
  readonly subagents_peak_tokens: number;
  readonly subagents_peak_role: string;
  readonly compactions: number;
  readonly peak_by_role: Readonly<Record<string, number>>;
}

export interface WorkflowEfficiency {
  readonly sdlc_iterations: number;
  readonly active_duration_seconds: number;
  readonly active_duration_formatted: string;
  readonly role_effort_distribution_pct: Readonly<Record<string, number>>;
  readonly subagent_turn_velocity: SubagentTurnVelocity;
  readonly test_to_production_ratio: TestToCodeRatio;
  readonly reasoning_tokens: number;
  readonly reasoning_ratio_pct: number;
  readonly code_delivered: CodeDeliveredStats;
  readonly tool_execution: ToolExecutionHealth;
  readonly context_window_health: ContextWindowHealth;
}

export interface ModelMetric {
  readonly model: string;
  readonly invocations: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly cost_usd: number;
  readonly peak_context_tokens: number;
  readonly context_health: string;
}

export interface RoleMetric {
  readonly role: string;
  readonly invocations: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly cost_usd: number;
  readonly peak_context_tokens: number;
  readonly context_health: string;
}

export interface ThreadRoleMetric {
  readonly role: string;
  readonly models: readonly ModelMetric[];
}

export interface ThreadExecutionRecord {
  readonly session_id: string;
  readonly user: string;
  readonly roles: readonly ThreadRoleMetric[];
  readonly compactions: number;
}

export interface AgenticMetricsLedger {
  readonly ticket_id: string | null;
  readonly ticket_title: string | null;
  readonly ticket_type: string | null;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly updated_at: string;
  readonly facts: AgenticFactLedger;
  readonly workflow_efficiency: WorkflowEfficiency;
  readonly threads: readonly ThreadExecutionRecord[];
  readonly totals: GroupedMetric;
  readonly roles: readonly RoleMetric[];
  readonly models: readonly ModelMetric[];
  readonly anomalies?: readonly Anomaly[];
}

export interface FleetSummary {
  readonly total_tickets: number;
  readonly total_cost_usd: number;
  readonly total_invocations: number;
  readonly total_input_tokens: number;
  readonly total_output_tokens: number;
  readonly total_cache_read_tokens: number;
  readonly overall_cache_hit_rate_pct: number;
  readonly total_duration_seconds: number;
  readonly formatted_total_duration: string;
  readonly total_lines_added: number;
  readonly total_lines_deleted: number;
  readonly total_files_changed: number;
  readonly total_reasoning_tokens: number;
  readonly overall_reasoning_ratio_pct: number;
  readonly role_costs: Readonly<Record<string, number>>;
  readonly model_costs: Readonly<Record<string, number>>;
  readonly anomaly_summary: FleetAnomalySummary;
}

export type StandardTicketType =
  | 'feature'
  | 'bug'
  | 'chore'
  | 'docs'
  | 'refactor'
  | 'test'
  | 'ci'
  | 'perf'
  | 'build';

export interface TicketComparisonRow {
  readonly ticket_id: string;
  readonly ticket_type?: StandardTicketType | string | null;
  readonly updated_at: string;
  readonly sdlc_iterations: number;
  readonly active_duration_seconds: number;
  readonly active_duration_formatted: string;
  readonly cost_usd: number;
  readonly cache_hit_rate_pct: number;
  readonly test_to_code_ratio: string;
  readonly lines_added: number;
  readonly lines_deleted: number;
  readonly files_changed: number;
  readonly tool_success_rate_pct: number;
  readonly context_health_status: string;
  readonly orchestrator_saturation_pct: number;
  readonly peak_context_tokens: number;
  readonly ticket_title: string;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly lead_duration_seconds: number | null;
  readonly lead_duration_formatted: string;
}

export interface TicketArtifacts {
  readonly ticket_id: string;
  readonly contract_content: string | null;
  readonly summary_content: string | null;
}

export interface WorkspaceDetail {
  readonly ledger: AgenticMetricsLedger;
  readonly artifacts: TicketArtifacts;
  readonly anomalies: readonly Anomaly[];
}

export type TimeGranularity = 'hourly' | 'daily' | 'weekly';

export type TimeWindowPreset = '1d' | '7d' | '30d' | '90d' | 'all' | 'custom';

export type AutopsySortMode = 'ticket_id' | 'completed_at';

export type ActiveTab = 'landing' | 'overview' | 'deep-dive' | 'progression' | 'comparisons' | 'live';

export type ComparisonSortField =
  | 'completed_at'
  | 'started_at'
  | 'ticket_id'
  | 'cost_usd'
  | 'active_duration_seconds'
  | 'sdlc_iterations'
  | 'cache_hit_rate_pct'
  | 'tool_success_rate_pct'
  | 'lines_added'
  | 'orchestrator_saturation_pct';

export type ComparisonSortOrder = 'asc' | 'desc';

export interface TimeBucketMetric {
  readonly bucket_key: string;
  readonly bucket_date: string;
  readonly invocations: number;
  readonly cost_usd: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly reasoning_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly lines_added: number;
  readonly lines_deleted: number;
  readonly files_changed: number;
  readonly net_lines_added: number;
  readonly peak_context_tokens: number;
  readonly active_duration_seconds: number;
  readonly cumulative_cost_usd: number;
  readonly cumulative_input_tokens: number;
  readonly cumulative_output_tokens: number;
  readonly cumulative_lines_added: number;
  readonly cumulative_lines_deleted: number;
  readonly cumulative_net_lines_added: number;
  readonly cumulative_invocations: number;
}

export interface BuildInfo {
  readonly version: string;
  readonly commit: string;
  readonly build_date: string;
}

export type {
  LiveAnomalySeverity,
  LiveAnomalyRule,
  LiveAnomalyAlert,
  LivePartFact,
  LiveMessageFact,
  LiveTraceView,
  LiveSessionNode,
  LiveSessionTree,
} from './live';
