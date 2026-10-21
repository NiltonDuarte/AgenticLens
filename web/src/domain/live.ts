export type LiveAnomalySeverity = 'WARNING' | 'HIGH' | 'CRITICAL';

export type LiveAnomalyRule =
  | 'ConsecutiveToolErrors'
  | 'ContextCeilingApproach'
  | 'TokenVelocityRunaway'
  | 'ToolThrashing';

export interface LiveAnomalyAlert {
  readonly session_id: string;
  readonly rule_name: LiveAnomalyRule | string;
  readonly severity: LiveAnomalySeverity;
  readonly message: string;
  readonly trigger_event: string;
  readonly timestamp: string;
  readonly count?: number | null;
}

export interface LivePartFact {
  readonly id: string;
  readonly type: string;
  readonly tool_name?: string | null;
  readonly input_snippet?: string;
  readonly output_snippet?: string;
  readonly status: string;
  readonly is_error: boolean;
  readonly error_message?: string | null;
  readonly severity: string;
}

export interface LiveMessageFact {
  readonly id: string;
  readonly role: string;
  readonly content: string;
  readonly timestamp: string;
  readonly reasoning_content?: string | null;
  readonly parts: readonly LivePartFact[];
  readonly context_tokens: number;
  readonly input_tokens: number;
  readonly output_tokens: number;
  readonly cache_read_tokens: number;
  readonly reasoning_tokens: number;
  readonly total_tokens: number;
}

export interface LiveTraceView {
  readonly session_id: string;
  readonly messages: readonly LiveMessageFact[];
  readonly total_tool_calls: number;
  readonly failed_tool_calls: number;
  readonly recent_errors: readonly LivePartFact[];
  readonly anomalies: readonly LiveAnomalyAlert[];
}

export interface LiveSessionNode {
  readonly id: string;
  readonly parent_id?: string | null;
  readonly agent: string;
  readonly role: string;
  readonly model: string;
  readonly title: string;
  readonly time_created: string;
  readonly time_updated: string;
  readonly status: string;
  readonly guessed_ticket_id?: string | null;
  readonly ticket_confidence: number;
  readonly sub_sessions: readonly LiveSessionNode[];
  readonly anomalies: readonly LiveAnomalyAlert[];
}

export interface LiveSessionTree {
  readonly root_sessions: readonly LiveSessionNode[];
  readonly last_refreshed_at: string;
  readonly total_active_sessions: number;
  readonly active_anomalies: readonly LiveAnomalyAlert[];
}
