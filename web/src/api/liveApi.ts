import type {
  LiveAnomalyAlert,
  LiveAnomalyRule,
  LiveMessageFact,
  LivePartFact,
  LiveSessionNode,
  LiveSessionTree,
  LiveTraceView,
  LiveTraceDiagnostic,
} from '../domain/live';

export type ConnectionStatus =
  | 'connecting'
  | 'connected'
  | 'fallback-polling'
  | 'paused'
  | 'error';

export interface LiveSourceDiagnostic {
  readonly message: string;
  readonly session_id?: string;
}

export interface LiveStreamCallbacks {
  readonly onSourceDiagnostic?: (diagnostic: LiveSourceDiagnostic) => void;
  readonly onSessionTree: (tree: LiveSessionTree) => void;
  readonly onAnomalyAlert?: ((alert: LiveAnomalyAlert) => void) | undefined;
  readonly onTraceUpdate?: ((trace: LiveTraceView) => void) | undefined;
  readonly onStatusChange?: ((status: ConnectionStatus) => void) | undefined;
  readonly onError?: ((error: Error) => void) | undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function isOptionalString(value: unknown): boolean {
  return value === undefined || isString(value);
}

function isOptionalNullableString(value: unknown): value is string | null | undefined {
  return value === null || value === undefined || isString(value);
}

function isCount(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0;
}

function isLiveAnomalyRule(value: unknown): value is LiveAnomalyRule {
  return value === 'ConsecutiveToolErrors' || value === 'ToolHighFailureRate' ||
    value === 'ContextCeilingApproach' || value === 'TokenVelocityRunaway' || value === 'ToolThrashing';
}

/** Validate alert facts without interpreting or normalizing source evidence. */
export function parseLiveAnomalyAlert(data: unknown): LiveAnomalyAlert {
  if (!isRecord(data) || typeof data.session_id !== 'string' || !isLiveAnomalyRule(data.rule_name) ||
    (data.severity !== 'WARNING' && data.severity !== 'HIGH' && data.severity !== 'CRITICAL') ||
    typeof data.message !== 'string' || typeof data.trigger_event !== 'string' || typeof data.timestamp !== 'string') {
    throw new Error('Invalid live anomaly alert');
  }
  const { count, tool_name: toolName, part_id: partId } = data;
  if (count !== undefined && count !== null && !isCount(count)) {
    throw new Error('Invalid live anomaly count');
  }
  if (!isOptionalNullableString(toolName)) throw new Error('Invalid live anomaly tool_name');
  if (!isOptionalNullableString(partId)) throw new Error('Invalid live anomaly part_id');
  return {
    session_id: data.session_id,
    rule_name: data.rule_name,
    severity: data.severity,
    message: data.message,
    trigger_event: data.trigger_event,
    timestamp: data.timestamp,
    ...(count !== undefined ? { count } : {}),
    ...(toolName !== undefined ? { tool_name: toolName } : {}),
    ...(partId !== undefined ? { part_id: partId } : {}),
  };
}

function parseRelationships(data: unknown): LiveTraceView['relationships'] {
  if (data === undefined) return undefined;
  if (!isRecord(data) || (data.state !== 'complete' && data.state !== 'unavailable')) {
    throw new Error('Invalid live trace relationship completeness');
  }
  return { state: data.state };
}

function parseTraceDiagnostics(data: unknown): readonly LiveTraceDiagnostic[] | undefined {
  if (data === undefined) return undefined;
  if (!Array.isArray(data)) throw new Error('Invalid live trace diagnostics');
  return data.map((value: unknown): LiveTraceDiagnostic => {
    if (!isRecord(value) || value.code !== 'child_relationship_lookup_failed' || typeof value.message !== 'string') {
      throw new Error('Invalid live trace diagnostic');
    }
    return { code: value.code, message: value.message };
  });
}

function isLiveSessionNode(data: unknown): data is LiveSessionNode {
  if (!isRecord(data) || !['id', 'agent', 'role', 'model', 'title', 'time_created', 'time_updated', 'status']
    .every((key) => isString(data[key])) ||
    typeof data.ticket_confidence !== 'number' || !Number.isFinite(data.ticket_confidence) ||
    !Array.isArray(data.sub_sessions) || !data.sub_sessions.every(isLiveSessionNode) ||
    !Array.isArray(data.anomalies)) return false;
  if (!isOptionalNullableString(data.parent_id) || !isOptionalNullableString(data.guessed_ticket_id)) return false;
  data.anomalies.forEach(parseLiveAnomalyAlert);
  return true;
}

/** A successful current API response is measured data; malformed input is unavailable. */
export function parseLiveSessionTree(data: unknown): LiveSessionTree {
  if (!isRecord(data) || !Array.isArray(data.root_sessions) || !data.root_sessions.every(isLiveSessionNode) ||
    typeof data.last_refreshed_at !== 'string' || !isCount(data.total_active_sessions) ||
    !Array.isArray(data.active_anomalies)) throw new Error('Invalid live session tree');
  return {
    root_sessions: data.root_sessions,
    last_refreshed_at: data.last_refreshed_at,
    total_active_sessions: data.total_active_sessions,
    active_anomalies: data.active_anomalies.map(parseLiveAnomalyAlert),
  };
}

function isLivePart(value: unknown): value is LivePartFact {
  if (!isRecord(value)) return false;
  const { severity } = value;
  return ['id', 'type', 'status', 'input_snippet', 'output_snippet'].every((key) => isString(value[key])) &&
    typeof value.is_error === 'boolean' &&
    ['skill_name', 'child_session_id', 'subagent_type']
      .every((key) => isOptionalString(value[key])) &&
    isOptionalNullableString(value.tool_name) && isOptionalNullableString(value.error_message) &&
    (severity === 'none' || severity === 'low' || severity === 'medium' || severity === 'high');
}

function isLiveMessage(value: unknown): value is LiveMessageFact {
  return isRecord(value) &&
    ['id', 'role', 'content', 'timestamp'].every((key) => isString(value[key])) &&
    ['context_tokens', 'input_tokens', 'output_tokens', 'cache_read_tokens', 'reasoning_tokens', 'total_tokens']
      .every((key) => isCount(value[key])) &&
    isOptionalNullableString(value.reasoning_content) &&
    Array.isArray(value.parts) && value.parts.every(isLivePart);
}

export function parseLiveTraceView(data: unknown): LiveTraceView {
  if (!isRecord(data) || !isString(data.session_id) || data.session_id.trim() === '' ||
    !Array.isArray(data.messages) || !isCount(data.total_tool_calls) || !isCount(data.failed_tool_calls) ||
    !Array.isArray(data.anomalies)) {
    throw new Error('Invalid live trace');
  }
  if (!data.messages.every(isLiveMessage)) throw new Error('Invalid live trace message or part');
  // The native producer emits null for an empty recent-errors slice.
  const recentErrors = data.recent_errors === null ? [] : data.recent_errors;
  if (!Array.isArray(recentErrors) || !recentErrors.every(isLivePart)) {
    throw new Error('Invalid live trace recent errors');
  }
  const relationships = parseRelationships(data.relationships);
  const diagnostics = parseTraceDiagnostics(data.diagnostics);
  return {
    session_id: data.session_id,
    messages: data.messages,
    total_tool_calls: data.total_tool_calls,
    failed_tool_calls: data.failed_tool_calls,
    recent_errors: recentErrors,
    anomalies: data.anomalies.map(parseLiveAnomalyAlert),
    ...(relationships !== undefined ? { relationships } : {}),
    ...(diagnostics !== undefined ? { diagnostics } : {}),
  };
}

export class LivePayloadError extends Error {
  constructor(readonly scope: 'tree' | 'trace', cause: unknown, readonly sessionId?: string) {
    super(cause instanceof Error ? cause.message : String(cause), { cause });
  }
}

function parseStreamTrace(raw: string): LiveTraceView {
  let payload: unknown;
  try {
    payload = JSON.parse(raw);
    return parseLiveTraceView(payload);
  } catch (cause: unknown) {
    const id = isRecord(payload) ? payload.session_id : undefined;
    throw new LivePayloadError('trace', cause, isString(id) && id.trim() ? id : undefined);
  }
}

// A server response can describe an unsupported or failing source even when
// the transport works. Keep that distinction when polling falls back to REST.
export class LiveResponseError extends Error {
  constructor(message: string, readonly isSourceReadFailure: boolean) {
    super(message);
  }
}

async function liveResponseError(response: Response, fallback: string): Promise<LiveResponseError> {
  let body: unknown;
  try {
    body = await response.json();
  } catch {
    // A proxy or older server may return an empty or non-JSON error body.
  }
  const message = isRecord(body) && typeof body.message === 'string' && body.message.trim()
    ? body.message : fallback;
  const code = isRecord(body) ? body.error : undefined;
  const isSourceReadFailure = (response.status === 500 && code === 'database_error') ||
    (response.status === 503 && (code === 'source_unavailable' || code === 'source_unsupported'));
  return new LiveResponseError(`HTTP ${response.status}: ${message}`, isSourceReadFailure);
}

export async function fetchLiveSessionTree(): Promise<LiveSessionTree> {
  try {
    const res = await fetch('/api/live/sessions');
    if (!res.ok) {
      throw await liveResponseError(res, 'Failed to fetch live sessions');
    }
    try {
      const data: unknown = await res.json();
      return parseLiveSessionTree(data);
    } catch (cause: unknown) {
      throw new LivePayloadError('tree', cause);
    }
  } catch (err: unknown) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error(String(err));
  }
}

export async function fetchLiveTrace(sessionId: string): Promise<LiveTraceView> {
  const cleanId = sessionId.trim();
  const encoded = encodeURIComponent(cleanId);
  try {
    const res = await fetch(`/api/live/sessions/${encoded}/trace`);
    if (!res.ok) {
      throw await liveResponseError(res, `Failed to fetch live trace for session ${cleanId}`);
    }
    const data: unknown = await res.json();
    const trace = parseLiveTraceView(data);
    if (trace.session_id !== cleanId) throw new Error('Trace response session does not match selection');
    return trace;
  } catch (err: unknown) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error(String(err));
  }
}

// Compare completed observations by request start order. A newer pending
// request does not prevent an older completed request from supplying data.
export class LiveReadOrder {
  private issued = 0;
  private acceptedTree = 0;

  begin(): number {
    return ++this.issued;
  }

  acceptTree(request: number): boolean {
    if (request <= this.acceptedTree) return false;
    this.acceptedTree = request;
    return true;
  }

  get treeRevision(): number {
    return this.acceptedTree;
  }
}

export function subscribeToLiveStream(
  callbacks: LiveStreamCallbacks, pollIntervalMs: number | undefined, reads = new LiveReadOrder(),
): () => void {
  let isClosed = false;
  let eventSource: EventSource | null = null;
  let pollingTimer: ReturnType<typeof setInterval> | null = null;
  let pollInFlight = false;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let sseConsecutiveErrors = 0;
  let pollingConsecutiveErrors = 0;
  let reconnectDelayMs = 1000;
  let sseDirectlyFailed = false;

  const reportError = (cause: unknown) => {
    if (!isClosed) callbacks.onError?.(cause instanceof Error ? cause : new Error(String(cause)));
  };

  const teardown = () => {
    isClosed = true;
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (pollingTimer !== null) {
      clearInterval(pollingTimer);
      pollingTimer = null;
    }
    if (eventSource !== null) {
      eventSource.close();
      eventSource = null;
    }
  };

  const startFallbackPolling = () => {
    if (isClosed || pollingTimer !== null) return;
    if (eventSource !== null) {
      eventSource.close();
      eventSource = null;
    }
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    if (pollIntervalMs === undefined || !Number.isSafeInteger(pollIntervalMs) || pollIntervalMs <= 0 || pollIntervalMs > 2_147_483_647) {
      callbacks.onStatusChange?.('error');
      reportError(new Error('live.poll_interval_ms must be explicitly configured as an integer from 1 through 2147483647 milliseconds for dashboard polling. Complete setup in the selected shared YAML configuration.'));
      return;
    }
    callbacks.onStatusChange?.('fallback-polling');
    if (isClosed) return;

    const poll = async () => {
      if (isClosed || pollInFlight) return;
      pollInFlight = true;
      const request = reads.begin();
      try {
        const tree = await fetchLiveSessionTree();
        if (!isClosed && reads.acceptTree(request)) {
          pollingConsecutiveErrors = 0;
          callbacks.onStatusChange?.('fallback-polling');
          if (isClosed || reads.treeRevision !== request) return;
          try {
            callbacks.onSessionTree(tree);
          } catch (callbackErr: unknown) {
            reportError(callbackErr);
          }
        }
      } catch (err: unknown) {
        if (!isClosed && reads.acceptTree(request)) {
          pollingConsecutiveErrors++;
          reportError(err);
          if (!isClosed && reads.treeRevision === request && (sseDirectlyFailed || pollingConsecutiveErrors >= 2)) {
            callbacks.onStatusChange?.('error');
          }
        }
      } finally {
        pollInFlight = false;
      }
    };

    pollingTimer = setInterval(() => {
      void poll();
    }, pollIntervalMs);
    void poll();
  };

  const connectSSE = () => {
    if (isClosed) return;
    if (typeof EventSource === 'undefined') {
      sseDirectlyFailed = true;
      reportError(new Error('EventSource is not supported in this environment'));
      startFallbackPolling();
      return;
    }

    callbacks.onStatusChange?.('connecting');
    if (isClosed) return;

    try {
      eventSource = new EventSource('/api/live/stream');
    } catch (err: unknown) {
      sseDirectlyFailed = true;
      reportError(err);
      startFallbackPolling();
      return;
    }

    // Events already queued by a previous connection must not supersede the
    // current subscription's snapshot or diagnostics after a reconnect.
    const source = eventSource;
    eventSource.onopen = () => {
      if (source !== eventSource || isClosed) return;
      sseConsecutiveErrors = 0;
      reconnectDelayMs = 1000;
      callbacks.onStatusChange?.('connected');
    };

    eventSource.addEventListener('session_tree', (event: MessageEvent<string>) => {
      if (source !== eventSource || isClosed) return;
      let tree: LiveSessionTree;
      try {
        const parsed: unknown = JSON.parse(event.data);
        tree = parseLiveSessionTree(parsed);
      } catch (err: unknown) {
        reportError(new LivePayloadError('tree', err));
        return;
      }
      reads.acceptTree(reads.begin());
      try {
        callbacks.onSessionTree(tree);
      } catch (err: unknown) {
        reportError(err);
      }
    });

    eventSource.addEventListener('source_diagnostic', (event: MessageEvent<string>) => {
      if (source !== eventSource || isClosed) return;
      try {
        const data: unknown = JSON.parse(event.data);
        if (!isRecord(data) || data.error !== 'database_error' || typeof data.message !== 'string' ||
            (data.session_id !== undefined && typeof data.session_id !== 'string')) {
          throw new Error('Invalid live source diagnostic');
        }
        reads.acceptTree(reads.begin());
        callbacks.onSourceDiagnostic?.({
          message: data.message,
          ...(typeof data.session_id === 'string' ? { session_id: data.session_id } : {}),
        });
      } catch (err: unknown) {
        reportError(err);
      }
    });

    eventSource.addEventListener('anomaly_alert', (event: MessageEvent<string>) => {
      if (source !== eventSource || isClosed) return;
      try {
        const data: unknown = JSON.parse(event.data);
        const alert = parseLiveAnomalyAlert(data);
        callbacks.onAnomalyAlert?.(alert);
      } catch (err: unknown) {
        reportError(err);
      }
    });

    eventSource.addEventListener('trace_update', (event: MessageEvent<string>) => {
      if (source !== eventSource || isClosed) return;
      try {
        const trace = parseStreamTrace(event.data);
        callbacks.onTraceUpdate?.(trace);
      } catch (err: unknown) {
        reportError(err);
      }
    });

    eventSource.onerror = () => {
      if (source !== eventSource || isClosed) return;
      sseConsecutiveErrors++;
      source.close();
      eventSource = null;
      reportError(new Error('SSE connection error'));
      if (isClosed) return;

      if (sseConsecutiveErrors >= 3) {
        startFallbackPolling();
        return;
      }

      callbacks.onStatusChange?.('connecting');
      if (isClosed) return;
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connectSSE();
      }, reconnectDelayMs);
      reconnectDelayMs = Math.min(reconnectDelayMs * 2, 16000);
    };
  };

  connectSSE();
  return teardown;
}
