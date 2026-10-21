export type EventType = 'tool_call' | 'turn_message' | 'compaction';

export type DataSource = 'local_sqlite' | 'committed_ledger';

export interface EventFact {
  readonly timestamp: string;
  readonly session_id: string;
  readonly message_id?: string | null | undefined;
  readonly turn: number;
  readonly role: string;
  readonly event_type: EventType;
  readonly tool_name?: string | null | undefined;
  readonly status: string;
  readonly error_message?: string | null | undefined;
  readonly duration_ms?: number | null | undefined;
  readonly context_tokens: number;
  readonly prompt_content?: string | null | undefined;
  readonly reasoning_content?: string | null | undefined;
  readonly tool_input?: string | null | undefined;
  readonly tool_output?: string | null | undefined;
}

export interface WorkspaceEventsResponse {
  readonly ticket_id: string;
  readonly is_hydrated: boolean;
  readonly events: readonly EventFact[];
  readonly data_source: DataSource;
}

export type AnomalySeverity = 'WARNING' | 'CRITICAL';

export type AnomalyRuleID =
  | 'TOOL_RETRY_STREAK'
  | 'TOOL_HIGH_FAILURE_RATE'
  | 'CONTEXT_SATURATION_SPIKE'
  | 'SUBAGENT_REWORK_LOOP';

export interface AnomalyContext {
  readonly tool_name?: string | undefined;
  readonly session_id?: string | undefined;
  readonly turn?: number | undefined;
  readonly role?: string | undefined;
  readonly secondary_role?: string | undefined;
  readonly count?: number | undefined;
  readonly total_calls?: number | undefined;
  readonly token_delta?: number | undefined;
  readonly context_tokens?: number | undefined;
  readonly threshold?: string | undefined;
  readonly failure_rate_pct?: number | undefined;
}

export interface Anomaly {
  readonly rule_id: AnomalyRuleID;
  readonly severity: AnomalySeverity;
  readonly title: string;
  readonly description: string;
  readonly context: AnomalyContext;
}

export interface BrittleToolSummary {
  readonly tool_name: string;
  readonly anomaly_count: number;
  readonly total_calls: number;
  readonly error_count: number;
  readonly failure_rate_pct: number;
}

export interface FleetAnomalySummary {
  readonly total_anomalies: number;
  readonly warning_count: number;
  readonly critical_count: number;
  readonly severity_counts: Readonly<Record<string, number>>;
  readonly rule_counts: Readonly<Record<string, number>>;
  readonly total_tickets: number;
  readonly tickets_with_anomalies: number;
  readonly anomaly_rate_per_ticket: number;
  readonly brittle_tools: readonly BrittleToolSummary[];
}
