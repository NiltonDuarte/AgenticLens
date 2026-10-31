import type { AgenticMetricsLedger, Anomaly, BuildInfo, EventFact, FleetSummary, ProjectConfigResponse, WorkspaceDetail } from '../domain/types';

export function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}
const string = (value: unknown): value is string => typeof value === 'string';
const number = (value: unknown): value is number => typeof value === 'number' && Number.isFinite(value);
const nullableString = (value: unknown): boolean => value === null || string(value);
const optionalString = (value: unknown): boolean => value === undefined || nullableString(value);
const optionalNumber = (value: unknown): boolean => value === undefined || number(value);
const strings = (value: unknown): boolean => Array.isArray(value) && value.every(string);
const numericMap = (value: unknown): boolean => isRecord(value) && Object.values(value).every(number);
const fields = (value: Record<string, unknown>, keys: readonly string[], check: (item: unknown) => boolean): boolean => keys.every((key) => check(value[key]));
const array = (value: unknown, check: (item: unknown) => boolean): boolean => Array.isArray(value) && value.every(check);
const optionalArray = (value: unknown, check: (item: unknown) => boolean): boolean => value === undefined || array(value, check);
const health = (value: unknown): boolean => value === 'HEALTHY' || value === 'ELEVATED' || value === 'BLOATED' || value === 'CRITICAL';

function eventFields(value: Record<string, unknown>): boolean {
  return fields(value, ['timestamp', 'role', 'status'], string) && fields(value, ['turn', 'context_tokens'], number) &&
    (value.event_type === 'tool_call' || value.event_type === 'turn_message' || value.event_type === 'compaction') &&
    fields(value, ['event_id', 'message_id', 'tool_name', 'error_message', 'prompt_content', 'reasoning_content', 'tool_input', 'tool_output'], optionalString) &&
    (value.duration_ms === null || optionalNumber(value.duration_ms));
}
export function isEventFact(value: unknown): value is EventFact {
  return isRecord(value) && string(value.session_id) && eventFields(value);
}
export function isAnomaly(value: unknown): value is Anomaly {
  if (!isRecord(value) || !isRecord(value.context)) return false;
  return (value.severity === 'WARNING' || value.severity === 'CRITICAL') &&
    (value.rule_id === 'TOOL_RETRY_STREAK' || value.rule_id === 'TOOL_HIGH_FAILURE_RATE' || value.rule_id === 'CONTEXT_SATURATION_SPIKE' || value.rule_id === 'SUBAGENT_REWORK_LOOP') &&
    fields(value.context, ['tool_name', 'session_id', 'threshold'], (v) => v === undefined || string(v)) &&
    fields(value.context, ['role', 'secondary_role'], optionalString) &&
    fields(value.context, ['turn', 'count', 'total_calls', 'token_delta', 'context_tokens', 'failure_rate_pct'], optionalNumber) &&
    optionalArray(value.evidence, (ref) => isRecord(ref) && fields(ref, ['session_id', 'event_id'], string));
}
function grouped(value: unknown): boolean {
  return isRecord(value) && fields(value, ['invocations', 'input_tokens', 'output_tokens', 'cache_read_tokens', 'cache_hit_rate_pct', 'cost_usd', 'peak_context_tokens'], number) && health(value.context_health);
}
function model(value: unknown): boolean { return isRecord(value) && string(value.model) && grouped(value); }
function role(value: unknown): boolean { return isRecord(value) && string(value.role) && grouped(value); }
function invocation(value: unknown): boolean {
  return isRecord(value) && fields(value, ['role', 'model'], string) && fields(value, ['started_at', 'completed_at'], nullableString) &&
    fields(value, ['active_duration_seconds', 'turns', 'cost_usd', 'compactions'], number) && optionalNumber(value.user_iterations) &&
    isRecord(value.tokens) && fields(value.tokens, ['input', 'output', 'cache_read', 'reasoning', 'peak_context'], number);
}
function session(value: unknown): boolean {
  return isRecord(value) && string(value.id) && optionalString(value.parent_id) &&
    (value.source === undefined || value.source === 'unknown' || value.source === 'opencode' || value.source === 'codex') &&
    (value.user === undefined || string(value.user)) && array(value.invocations, invocation) &&
    array(value.tools, (v) => isRecord(v) && string(v.name) && fields(v, ['calls', 'errors'], number)) &&
    optionalArray(value.events, (v) => isRecord(v) && eventFields(v)) && optionalArray(value.anomalies, isAnomaly);
}
function ticketMetadata(value: Record<string, unknown>): boolean {
  return fields(value, ['ticket_id', 'ticket_title', 'ticket_type', 'started_at', 'completed_at'], nullableString) && string(value.updated_at);
}
function migration(value: unknown): boolean {
  return isRecord(value) && number(value.schema_version) && (value.status === 'pending' || value.status === 'partial' || value.status === 'completed') &&
    strings(value.steps_completed) && strings(value.steps_pending) && fields(value, ['last_migrated_at', 'last_migrated_by'], string) &&
    (value.notes === undefined || string(value.notes));
}
function facts(value: unknown): boolean {
  return isRecord(value) && ticketMetadata(value) && optionalNumber(value.schema_version) &&
    (value.migration === undefined || migration(value.migration)) && isRecord(value.code) &&
    fields(value.code, ['lines_added', 'lines_deleted', 'test_lines_added', 'test_lines_deleted', 'prod_lines_added', 'prod_lines_deleted', 'files_changed', 'sdlc_iterations'], number) &&
    array(value.sessions, session) && optionalArray(value.anomalies, isAnomaly) &&
    optionalArray(value.skills, (v) => isRecord(v) && fields(v, ['name', 'caller_role', 'timestamp'], string) && fields(v, ['turn_index', 'message_index'], number) && typeof v.success === 'boolean' && (v.error === undefined || string(v.error))) &&
    optionalArray(value.task_links, (v) => isRecord(v) && fields(v, ['subagent_type', 'description', 'timestamp'], string) && number(v.parent_turn) && (v.child_session_id === undefined || string(v.child_session_id)) && fields(v, ['duration_ms', 'cost_usd'], optionalNumber));
}
function workflow(value: unknown): boolean {
  if (!isRecord(value)) return false;
  const velocity = value.subagent_turn_velocity, ratio = value.test_to_production_ratio, code = value.code_delivered, tool = value.tool_execution, context = value.context_window_health;
  return fields(value, ['sdlc_iterations', 'active_duration_seconds', 'reasoning_tokens', 'reasoning_ratio_pct'], number) && string(value.active_duration_formatted) && numericMap(value.role_effort_distribution_pct) &&
    isRecord(velocity) && fields(velocity, ['avg_turns_per_invocation', 'peak_turns'], number) && numericMap(velocity.by_role) &&
    isRecord(ratio) && string(ratio.ratio) && fields(ratio, ['test_lines_added', 'prod_lines_added'], number) &&
    isRecord(code) && fields(code, ['lines_added', 'lines_deleted', 'files_changed'], number) &&
    isRecord(tool) && fields(tool, ['total_calls', 'errors_and_corrections', 'success_rate_pct'], number) &&
    isRecord(context) && fields(context, ['healthy_ceiling_tokens', 'root_peak_tokens', 'root_saturation_pct', 'subagents_peak_tokens', 'compactions'], number) && health(context.status) && string(context.subagents_peak_role) && numericMap(context.peak_by_role);
}
function isLedger(value: unknown): value is AgenticMetricsLedger {
  return isRecord(value) && ticketMetadata(value) && facts(value.facts) && workflow(value.workflow_efficiency) && grouped(value.totals) &&
    array(value.roles, role) && array(value.models, model) && optionalArray(value.anomalies, isAnomaly) &&
    array(value.threads, (v) => isRecord(v) && fields(v, ['session_id', 'user'], string) && number(v.compactions) &&
      array(v.roles, (r) => isRecord(r) && string(r.role) && array(r.models, model)));
}

// Go encodes nil collections as null. Normalize only known collection positions;
// missing required fields and malformed non-null values still fail validation.
function collections(value: unknown, arrays: readonly string[] = [], maps: readonly string[] = []): unknown {
  if (!isRecord(value)) return value;
  const result = { ...value };
  for (const key of arrays) if (result[key] === null) result[key] = [];
  for (const key of maps) if (result[key] === null) result[key] = {};
  return result;
}
function prepareLedger(value: unknown): unknown {
  const result = collections(value, ['threads', 'roles', 'models', 'anomalies']);
  if (!isRecord(result)) return result;
  const rawFacts = collections(result.facts, ['sessions', 'anomalies', 'skills', 'task_links']);
  if (isRecord(rawFacts)) {
    if (Array.isArray(rawFacts.sessions)) rawFacts.sessions = rawFacts.sessions.map((s: unknown) => collections(s, ['invocations', 'tools', 'events', 'anomalies']));
    if (rawFacts.migration !== undefined) rawFacts.migration = collections(rawFacts.migration, ['steps_completed', 'steps_pending']);
    result.facts = rawFacts;
  }
  const efficiency = collections(result.workflow_efficiency, [], ['role_effort_distribution_pct']);
  if (isRecord(efficiency)) {
    efficiency.subagent_turn_velocity = collections(efficiency.subagent_turn_velocity, [], ['by_role']);
    efficiency.context_window_health = collections(efficiency.context_window_health, [], ['peak_by_role']);
    result.workflow_efficiency = efficiency;
  }
  if (Array.isArray(result.threads)) result.threads = result.threads.map((thread: unknown) => {
    const item = collections(thread, ['roles']);
    if (isRecord(item) && Array.isArray(item.roles)) item.roles = item.roles.map((r: unknown) => collections(r, ['models']));
    return item;
  });
  return result;
}
export function decodeLedgers(value: unknown): AgenticMetricsLedger[] {
  if (!Array.isArray(value)) throw new Error('Invalid workspace ledgers response');
  return value.map((entry: unknown) => {
    const ledger = prepareLedger(entry);
    if (!isLedger(ledger)) throw new Error('Invalid workspace ledgers response');
    return ledger;
  });
}
function isFleet(value: unknown): value is FleetSummary {
  if (!isRecord(value) || !isRecord(value.anomaly_summary)) return false;
  const anomaly = value.anomaly_summary;
  return fields(value, ['total_tickets', 'total_cost_usd', 'total_invocations', 'total_input_tokens', 'total_output_tokens', 'total_cache_read_tokens', 'overall_cache_hit_rate_pct', 'total_duration_seconds', 'total_lines_added', 'total_lines_deleted', 'total_files_changed', 'total_reasoning_tokens', 'overall_reasoning_ratio_pct'], number) &&
    string(value.formatted_total_duration) && numericMap(value.role_costs) && numericMap(value.model_costs) &&
    fields(anomaly, ['total_anomalies', 'warning_count', 'critical_count', 'total_tickets', 'tickets_with_anomalies', 'anomaly_rate_per_ticket'], number) && numericMap(anomaly.severity_counts) && numericMap(anomaly.rule_counts) &&
    array(anomaly.brittle_tools, (v) => isRecord(v) && string(v.tool_name) && fields(v, ['anomaly_count', 'total_calls', 'error_count', 'failure_rate_pct'], number)) &&
    optionalArray(value.skill_adoption, (v) => isRecord(v) && string(v.name) && fields(v, ['total_invocations', 'tickets_count', 'failure_count', 'failure_rate'], number));
}
export function decodeFleet(value: unknown): FleetSummary {
  const result = collections(value, ['skill_adoption'], ['role_costs', 'model_costs']);
  if (isRecord(result)) result.anomaly_summary = collections(result.anomaly_summary, ['brittle_tools'], ['severity_counts', 'rule_counts']);
  if (!isFleet(result)) throw new Error('Invalid fleet summary response');
  return result;
}
function isDetail(value: unknown): value is WorkspaceDetail {
  if (!isRecord(value) || !isLedger(value.ledger) || !isRecord(value.artifacts)) return false;
  const artifacts = value.artifacts;
  return string(artifacts.ticket_id) && fields(artifacts, ['contract_content', 'summary_content'], optionalString) &&
    optionalArray(artifacts.items, (v) => isRecord(v) && fields(v, ['path', 'title', 'description'], string) && nullableString(v.content)) && array(value.anomalies, isAnomaly);
}
export function decodeDetail(value: unknown): WorkspaceDetail {
  const result = collections(value, ['anomalies']);
  if (isRecord(result)) {
    result.ledger = prepareLedger(result.ledger);
    const artifacts = collections(result.artifacts, ['items']);
    if (isRecord(artifacts) && Array.isArray(artifacts.items)) {
      // Go omits an empty description; the frontend model keeps it as a string.
      artifacts.items = artifacts.items.map((item: unknown) =>
        isRecord(item) && item.description === undefined ? { ...item, description: '' } : item);
    }
    result.artifacts = artifacts;
  }
  if (!isDetail(result)) throw new Error('Invalid workspace detail response');
  return result;
}
export function isBuildInfo(value: unknown): value is BuildInfo {
  return isRecord(value) && fields(value, ['version', 'commit', 'build_date'], string);
}
export function isProjectConfig(value: unknown): value is ProjectConfigResponse {
  return isRecord(value) && fields(value, ['project_name', 'workspaces_dir', 'version'], string) && strings(value.role_order) &&
    array(value.prefixes, (v) => isRecord(v) && fields(v, ['prefix', 'description'], string) && (v.ticket_type === undefined || string(v.ticket_type)));
}
