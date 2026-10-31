export type ContextHealthStatus = 'HEALTHY' | 'ELEVATED' | 'BLOATED' | 'CRITICAL';

export type {
  AgentRole,
  SourceHarness,
} from '../types/telemetry';

import type {
  AgentRole,
  EventFact,
  SourceHarness,
} from '../types/telemetry';

export type StandardTicketType =
  | 'feat'
  | 'fix'
  | 'feature'
  | 'bug'
  | 'chore'
  | 'docs'
  | 'refactor'
  | 'test'
  | 'ci'
  | 'perf'
  | 'build';

export type TicketType = StandardTicketType | (string & {});

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


export interface InvocationTokens {
  readonly input: number;
  readonly output: number;
  readonly cache_read: number;
  readonly reasoning: number;
  readonly peak_context: number;
}

export interface InvocationFact {
  readonly role: AgentRole;
  readonly model: string;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly active_duration_seconds: number;
  readonly turns: number;
  readonly tokens: InvocationTokens;
  readonly cost_usd: number;
  readonly compactions: number;
  readonly user_iterations?: number;
}

export interface ToolFact {
  readonly session_id: string;
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
  ConventionalTicketType,
} from '../types/telemetry';

import type {
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

export interface SkillFact {
  readonly name: string;
  readonly turn_index: number;
  readonly message_index: number;
  readonly caller_role: string;
  readonly timestamp: string;
  readonly success: boolean;
  readonly error?: string;
}

export interface TaskLinkFact {
  readonly subagent_type: string;
  readonly description: string;
  readonly parent_turn: number;
  readonly child_session_id?: string;
  readonly timestamp: string;
  readonly duration_ms?: number;
  readonly cost_usd?: number;
}

export interface AgenticFactLedger {
  readonly schema_version?: number;
  readonly migration?: MigrationInfo;
  readonly ticket_id: string | null;
  readonly ticket_title: string | null;
  readonly ticket_type: TicketType | null;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly updated_at: string;
  readonly code: CodeFacts;
  readonly sessions: readonly SessionFact[];
  readonly anomalies?: readonly Anomaly[];
  readonly skills?: readonly SkillFact[];
  readonly task_links?: readonly TaskLinkFact[];
}

export function ledgerSkills(ledger: Pick<AgenticFactLedger, 'skills'> | undefined): readonly SkillFact[] {
  return ledger?.skills ?? [];
}

export function ledgerTaskLinks(ledger: Pick<AgenticFactLedger, 'task_links'> | undefined): readonly TaskLinkFact[] {
  return ledger?.task_links ?? [];
}

export interface WorkspaceDiagnostic {
  readonly ticket_id: string;
  readonly message: string;
}

export interface SessionFact {
  readonly id: string;
  readonly parent_id?: string | null;
  readonly source?: SourceHarness;
  readonly user?: string;
  readonly invocations: readonly InvocationFact[];
  readonly tools: readonly Omit<ToolFact, 'session_id'>[];
  readonly events?: readonly SessionEventFact[];
  readonly anomalies?: readonly Anomaly[];
}

export type SessionEventFact = Pick<EventFact,
  | 'event_id'
  | 'timestamp'
  | 'message_id'
  | 'turn'
  | 'role'
  | 'event_type'
  | 'tool_name'
  | 'status'
  | 'duration_ms'
  | 'context_tokens'
>;

export function sessionInvocations(sessions: readonly SessionFact[] | undefined): readonly InvocationFact[] {
  return sessions?.flatMap((session) => session.invocations ?? []) ?? [];
}

export function sessionTools(sessions: readonly SessionFact[] | undefined): readonly ToolFact[] {
  return sessions?.flatMap((session) => (session.tools ?? []).map((tool) => ({ ...tool, session_id: session.id }))) ?? [];
}

export interface GroupedMetric {
  readonly invocations: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly cost_usd: number;
  readonly peak_context_tokens: number;
  readonly context_health: ContextHealthStatus;
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
  readonly status: ContextHealthStatus;
  readonly root_peak_tokens: number;
  readonly root_saturation_pct: number;
  readonly subagents_peak_tokens: number;
  readonly subagents_peak_role: AgentRole;
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
  readonly context_health: ContextHealthStatus;
}

export interface RoleMetric {
  readonly role: AgentRole;
  readonly invocations: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly cache_hit_rate_pct: number;
  readonly cost_usd: number;
  readonly peak_context_tokens: number;
  readonly context_health: ContextHealthStatus;
}

export interface ThreadRoleMetric {
  readonly role: AgentRole;
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
  readonly ticket_type: TicketType | null;
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

export interface SkillAdoptionFact {
  readonly name: string;
  readonly total_invocations: number;
  readonly tickets_count: number;
  readonly failure_count: number;
  readonly failure_rate: number;
}

export interface FleetSummary {
  readonly skill_adoption?: readonly SkillAdoptionFact[];
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

export interface TicketComparisonRow {
  readonly skills_used?: readonly string[];
  readonly ticket_id: string;
  readonly ticket_type: TicketType | null;
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
  readonly context_health_status: ContextHealthStatus;
  readonly root_saturation_pct: number;
  readonly peak_context_tokens: number;
  readonly ticket_title: string;
  readonly started_at: string | null;
  readonly completed_at: string | null;
  readonly lead_duration_seconds: number | null;
  readonly lead_duration_formatted: string;
}

export interface TicketArtifactItem {
  readonly path: string;
  readonly title: string;
  readonly description: string;
  readonly content: string | null;
}

export interface TicketArtifacts {
  readonly ticket_id: string;
  readonly contract_content?: string | null;
  readonly summary_content?: string | null;
  readonly items?: readonly TicketArtifactItem[];
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
  | 'root_saturation_pct';

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

export interface TicketPrefixConfig {
  readonly prefix: string;
  readonly description: string;
  readonly ticket_type?: string;
}

export interface ProjectConfigResponse {
  readonly project_name: string;
  readonly prefixes: readonly TicketPrefixConfig[];
  readonly workspaces_dir: string;
  readonly role_order: readonly string[];
  readonly version: string;
}

export type {
  FailureSeverity,
  LiveAnomalySeverity,
  LiveAnomalyRule,
  LiveAnomalyAlert,
  LivePartFact,
  LiveMessageFact,
  LiveTraceView,
  LiveSessionNode,
  LiveSessionTree,
} from './live';

export type {
  AnomalyRule,
  ToolFailureAnomalyRule,
} from './anomalies';
