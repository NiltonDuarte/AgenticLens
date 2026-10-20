import type {
  LiveAnomalyAlert,
  LiveSessionNode,
  LiveSessionTree,
  LiveTraceView,
} from '../domain/live';

export type ConnectionStatus =
  | 'connecting'
  | 'connected'
  | 'fallback-polling'
  | 'paused'
  | 'error';

export interface LiveStreamCallbacks {
  readonly onSessionTree: (tree: LiveSessionTree) => void;
  readonly onAnomalyAlert?: ((alert: LiveAnomalyAlert) => void) | undefined;
  readonly onTraceUpdate?: ((trace: LiveTraceView) => void) | undefined;
  readonly onStatusChange?: ((status: ConnectionStatus) => void) | undefined;
  readonly onError?: ((error: Error) => void) | undefined;
}

export function parseLiveSessionTree(data: unknown): LiveSessionTree {
  if (data && typeof data === 'object') {
    if (Array.isArray(data)) {
      return {
        root_sessions: data as readonly LiveSessionNode[],
        last_refreshed_at: new Date().toISOString(),
        total_active_sessions: data.length,
        active_anomalies: [],
      };
    }
    const obj = data as Partial<LiveSessionTree>;
    const rootSessions = Array.isArray(obj.root_sessions) ? obj.root_sessions : [];
    const activeAnomalies = Array.isArray(obj.active_anomalies) ? obj.active_anomalies : [];
    const totalActive = typeof obj.total_active_sessions === 'number'
      ? obj.total_active_sessions
      : rootSessions.length;
    const lastRefreshed = typeof obj.last_refreshed_at === 'string' && obj.last_refreshed_at.length > 0
      ? obj.last_refreshed_at
      : new Date().toISOString();

    return {
      root_sessions: rootSessions,
      last_refreshed_at: lastRefreshed,
      total_active_sessions: totalActive,
      active_anomalies: activeAnomalies,
    };
  }

  return {
    root_sessions: [],
    last_refreshed_at: new Date().toISOString(),
    total_active_sessions: 0,
    active_anomalies: [],
  };
}

export function parseLiveTraceView(data: unknown, fallbackSessionId: string): LiveTraceView {
  if (data && typeof data === 'object') {
    const obj = data as Partial<LiveTraceView>;
    return {
      session_id: typeof obj.session_id === 'string' && obj.session_id.length > 0
        ? obj.session_id
        : fallbackSessionId,
      messages: Array.isArray(obj.messages) ? obj.messages : [],
      total_tool_calls: typeof obj.total_tool_calls === 'number' ? obj.total_tool_calls : 0,
      failed_tool_calls: typeof obj.failed_tool_calls === 'number' ? obj.failed_tool_calls : 0,
      recent_errors: Array.isArray(obj.recent_errors) ? obj.recent_errors : [],
      anomalies: Array.isArray(obj.anomalies) ? obj.anomalies : [],
    };
  }

  return {
    session_id: fallbackSessionId,
    messages: [],
    total_tool_calls: 0,
    failed_tool_calls: 0,
    recent_errors: [],
    anomalies: [],
  };
}

export async function fetchLiveSessionTree(): Promise<LiveSessionTree> {
  try {
    const res = await fetch('/api/live/sessions');
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}: Failed to fetch live sessions`);
    }
    const data: unknown = await res.json();
    return parseLiveSessionTree(data);
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
      throw new Error(`HTTP ${res.status}: Failed to fetch live trace for session ${cleanId}`);
    }
    const data: unknown = await res.json();
    return parseLiveTraceView(data, cleanId);
  } catch (err: unknown) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error(String(err));
  }
}

export function subscribeToLiveStream(callbacks: LiveStreamCallbacks): () => void {
  let isClosed = false;
  let eventSource: EventSource | null = null;
  let pollingTimer: ReturnType<typeof setInterval> | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let sseConsecutiveErrors = 0;
  let pollingConsecutiveErrors = 0;
  let reconnectDelayMs = 1000;
  let sseDirectlyFailed = false;

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
    if (isClosed) return;
    if (eventSource !== null) {
      eventSource.close();
      eventSource = null;
    }
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }

    callbacks.onStatusChange?.('fallback-polling');

    const poll = async () => {
      if (isClosed) return;
      try {
        const tree = await fetchLiveSessionTree();
        if (!isClosed) {
          pollingConsecutiveErrors = 0;
          try {
            callbacks.onSessionTree(tree);
          } catch (callbackErr: unknown) {
            const error =
              callbackErr instanceof Error
                ? callbackErr
                : new Error(String(callbackErr));
            callbacks.onError?.(error);
          }
        }
      } catch (err: unknown) {
        if (!isClosed) {
          pollingConsecutiveErrors++;
          const error = err instanceof Error ? err : new Error(String(err));
          callbacks.onError?.(error);
          if (sseDirectlyFailed || pollingConsecutiveErrors >= 2) {
            callbacks.onStatusChange?.('error');
          }
        }
      }
    };

    void poll();
    if (pollingTimer === null) {
      pollingTimer = setInterval(() => {
        void poll();
      }, 2000);
    }
  };

  const connectSSE = () => {
    if (isClosed) return;
    if (typeof EventSource === 'undefined') {
      sseDirectlyFailed = true;
      const error = new Error('EventSource is not supported in this environment');
      callbacks.onError?.(error);
      startFallbackPolling();
      return;
    }

    callbacks.onStatusChange?.('connecting');

    try {
      eventSource = new EventSource('/api/live/stream');
    } catch (err: unknown) {
      sseDirectlyFailed = true;
      const error = err instanceof Error ? err : new Error(String(err));
      callbacks.onError?.(error);
      startFallbackPolling();
      return;
    }

    eventSource.onopen = () => {
      if (isClosed) return;
      sseConsecutiveErrors = 0;
      reconnectDelayMs = 1000;
      callbacks.onStatusChange?.('connected');
    };

    eventSource.addEventListener('session_tree', (event: MessageEvent<string>) => {
      if (isClosed) return;
      try {
        const parsed: unknown = JSON.parse(event.data);
        callbacks.onSessionTree(parseLiveSessionTree(parsed));
      } catch (err: unknown) {
        const error = err instanceof Error ? err : new Error(String(err));
        callbacks.onError?.(error);
      }
    });

    eventSource.addEventListener('anomaly_alert', (event: MessageEvent<string>) => {
      if (isClosed) return;
      try {
        const alert = JSON.parse(event.data) as LiveAnomalyAlert;
        callbacks.onAnomalyAlert?.(alert);
      } catch (err: unknown) {
        const error = err instanceof Error ? err : new Error(String(err));
        callbacks.onError?.(error);
      }
    });

    eventSource.addEventListener('trace_update', (event: MessageEvent<string>) => {
      if (isClosed) return;
      try {
        const trace = JSON.parse(event.data) as LiveTraceView;
        callbacks.onTraceUpdate?.(trace);
      } catch (err: unknown) {
        const error = err instanceof Error ? err : new Error(String(err));
        callbacks.onError?.(error);
      }
    });

    eventSource.onerror = () => {
      if (isClosed) return;
      sseConsecutiveErrors++;
      const error = new Error('SSE connection error');
      callbacks.onError?.(error);

      if (eventSource !== null) {
        eventSource.close();
        eventSource = null;
      }

      if (sseConsecutiveErrors >= 3) {
        startFallbackPolling();
        return;
      }

      callbacks.onStatusChange?.('connecting');
      reconnectTimer = setTimeout(() => {
        if (!isClosed) {
          connectSSE();
        }
      }, reconnectDelayMs);
      reconnectDelayMs = Math.min(reconnectDelayMs * 2, 16000);
    };
  };

  connectSSE();
  return teardown;
}
