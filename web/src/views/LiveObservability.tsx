import { useProjectConfig } from '../contexts/ProjectConfigContext';
import React, { useEffect, useLayoutEffect, useRef, useState, useCallback } from 'react';
import type {
  LiveAnomalyAlert,
  LiveSessionNode,
  LiveSessionTree as LiveSessionTreeType,
} from '../domain/live';
import {
  type ConnectionStatus,
  fetchLiveSessionTree,
  LiveResponseError,
  LiveReadOrder,
  subscribeToLiveStream,
} from '../api/liveApi';
import { LiveSessionTree } from '../components/LiveSessionTree';
import { LiveTraceInspector } from '../components/LiveTraceInspector';
import { useLiveTrace } from '../hooks/useLiveTrace';
import { LiveAnomalyBanner } from '../components/LiveAnomalyBanner';

export interface LiveObservabilityProps {
  readonly onSelectTicket: (ticketId: string) => void;
}

function findSessionNode(
  nodes: readonly LiveSessionNode[],
  targetId: string
): LiveSessionNode | null {
  for (const node of nodes) {
    if (node.id === targetId) {
      return node;
    }
    if (node.sub_sessions && node.sub_sessions.length > 0) {
      const found = findSessionNode(node.sub_sessions, targetId);
      if (found) {
        return found;
      }
    }
  }
  return null;
}

function sessionIds(nodes: readonly LiveSessionNode[]): readonly string[] {
  return nodes.flatMap((node) => [node.id, ...sessionIds(node.sub_sessions)]);
}

export const LiveObservability: React.FC<LiveObservabilityProps> = ({
  onSelectTicket,
}) => {
  const projectConfig = useProjectConfig();
  const pollIntervalMs = projectConfig?.live_poll_interval_ms;
  const [isPaused, setIsPaused] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('connecting');
  const [sessionTree, setSessionTree] = useState<LiveSessionTreeType | null>(null);
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [treeDiagnostic, setTreeDiagnostic] = useState<string | null>(null);
  const [traceDiagnostics, setTraceDiagnostics] = useState<ReadonlyMap<string, string>>(new Map());
  const [reads] = useState(() => new LiveReadOrder());
  const [retryCount, setRetryCount] = useState(0);

  const selectedNode = selectedSessionId
    ? findSessionNode(sessionTree?.root_sessions ?? [], selectedSessionId)
    : null;
  const clearTraceDiagnostic = useCallback((id: string) => {
    setTraceDiagnostics((previous) => {
      if (!previous.has(id)) return previous;
      const next = new Map(previous);
      next.delete(id);
      return next;
    });
  }, []);

  const { trace: traceView, loading: loadingTrace, error: traceError, onTraceUpdate, onTraceReadFailure, onCompleteTree, captureRefresh } =
    useLiveTrace(selectedSessionId, isPaused, selectedNode, clearTraceDiagnostic, reads);

  const acceptTree = useCallback((tree: LiveSessionTreeType) => {
    onCompleteTree(sessionIds(tree.root_sessions), reads.treeRevision);
    setSessionTree(tree);
    setTreeDiagnostic(null);
    setErrorMessage(null);
    // A complete tree requires a successful trace read for every included
    // session. It also supersedes diagnostics for sessions that have left.
    setTraceDiagnostics((previous) => previous.size ? new Map() : previous);
  }, [onCompleteTree, reads]);

  const handleSelectSession = useCallback((sessionId: string) => {
    setSelectedSessionId(sessionId);
  }, []);

  const handleTogglePause = () => {
    setIsPaused((prev) => !prev);
  };

  const handleRetryConnection = useCallback(() => {
    setErrorMessage(null);
    setConnectionStatus('connecting');
    setRetryCount((prev) => prev + 1);
  }, []);

  const fallbackPolling = useRef(false);
  const manualOwner = useRef<object | null>(null);
  useLayoutEffect(() => {
    manualOwner.current = {};
    return () => { manualOwner.current = null; };
  }, [selectedSessionId, isPaused, retryCount]);

  const handleManualRefresh = useCallback(async () => {
    const owner = manualOwner.current;
    const refreshTrace = captureRefresh();
    refreshTrace();
    const revision = reads.begin();
    try {
      const tree = await fetchLiveSessionTree();
      if (manualOwner.current !== owner || !reads.acceptTree(revision)) return;
      acceptTree(tree);
      if (fallbackPolling.current) setConnectionStatus('fallback-polling');
    } catch (err: unknown) {
      if (manualOwner.current !== owner || !reads.acceptTree(revision)) return;
      if (err instanceof LiveResponseError) setTreeDiagnostic(err.message);
      else setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }, [acceptTree, captureRefresh, reads]);

  // Reactive streaming lifecycle
  useEffect(() => {
    fallbackPolling.current = false;
    if (isPaused) {
      setConnectionStatus('paused');
      return;
    }

    let subscribed = true;
    const unsubscribe = subscribeToLiveStream({
      onSessionTree: (tree) => {
        if (!subscribed) return;
        acceptTree(tree);
      },
      onAnomalyAlert: (alert: LiveAnomalyAlert) => {
        if (!subscribed) return;
        setSessionTree((prev) => {
          if (prev === null) return prev;
          const alreadyExists = prev.active_anomalies.some(
            (a) => a.session_id === alert.session_id && a.rule_name === alert.rule_name
          );
          if (alreadyExists) {
            return prev;
          }
          return {
            ...prev,
            active_anomalies: [alert, ...prev.active_anomalies],
          };
        });
      },
      onTraceUpdate: (trace) => {
        if (!subscribed) return;
        clearTraceDiagnostic(trace.session_id);
        onTraceUpdate(trace);
      },
      onSourceDiagnostic: (diagnostic) => {
        if (!subscribed) return;
        const id = diagnostic.session_id;
        if (id === undefined) {
          setTreeDiagnostic(diagnostic.message);
        } else {
          onTraceReadFailure(id);
          setTraceDiagnostics((previous) => new Map(previous).set(id, diagnostic.message));
        }
      },
      onStatusChange: (status) => {
        if (!subscribed) return;
        if (status === 'fallback-polling') fallbackPolling.current = true;
        setConnectionStatus(status);
      },
      onError: (err) => {
        if (!subscribed) return;
        if (err instanceof LiveResponseError) {
          setTreeDiagnostic(err.message);
        } else {
          setErrorMessage(err.message);
        }
      },
    }, pollIntervalMs, reads);

    return () => {
      subscribed = false;
      unsubscribe();
    };
  }, [isPaused, retryCount, acceptTree, clearTraceDiagnostic, onTraceUpdate, onTraceReadFailure, pollIntervalMs, reads]);

  const sourceMessage = [
    ...(treeDiagnostic ? [treeDiagnostic] : []),
    ...Array.from(traceDiagnostics, ([id, message]) => `Session ${id}: ${message}`),
    ...(traceError ? [traceError] : []),
  ].join(' · ');

  return (
    <div className="live-observability-view" data-testid="live-observability-view">
      <header className="live-status-bar" data-testid="live-status-bar">
        <div className="status-bar-left">
          <span className={`live-connection-pill status-${connectionStatus}`}>
            <span className="live-dot" />
            {connectionStatus === 'connected' && 'Connected (SSE)'}
            {connectionStatus === 'fallback-polling' && 'Fallback Polling (REST)'}
            {connectionStatus === 'connecting' && 'Connecting...'}
            {connectionStatus === 'paused' && 'Paused'}
            {connectionStatus === 'error' && (treeDiagnostic ? 'Live source error' : 'Connection Error')}
          </span>

          <span className="status-bar-stat">
            Active Sessions: <strong>{sessionTree?.total_active_sessions ?? 'Not available'}</strong>
          </span>

          <span className="status-bar-stat">
            Active Anomalies: <strong>{sessionTree?.active_anomalies.length ?? 'Not available'}</strong>
          </span>

          {sessionTree?.last_refreshed_at && (
            <span className="status-bar-stat refreshed-time">
              Refreshed: {sessionTree.last_refreshed_at}
            </span>
          )}

          {(sourceMessage || errorMessage) && connectionStatus !== 'error' && (
            <span className="status-bar-error" role="alert">
              Error: {sourceMessage || errorMessage}
            </span>
          )}
        </div>

        <div className="status-bar-actions">
          <button
            type="button"
            className="action-btn pause-resume-btn"
            onClick={handleTogglePause}
          >
            {isPaused ? 'Resume Streaming' : 'Pause Streaming'}
          </button>

          <button
            type="button"
            className="action-btn refresh-btn"
            onClick={handleManualRefresh}
          >
            Manual Refresh
          </button>
        </div>
      </header>

      {connectionStatus === 'error' && (
        <div
          className="live-connection-error-banner"
          role="alert"
          data-testid="live-connection-error-banner"
        >
          <div className="live-connection-error-content">
            <span className="live-connection-error-text">
              {sourceMessage ||
                'Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running.'}
            </span>
            <button
              type="button"
              className="live-retry-btn"
              onClick={handleRetryConnection}
            >
              Retry Connection
            </button>
          </div>
        </div>
      )}

      {sessionTree && sessionTree.active_anomalies.length > 0 && (
        <LiveAnomalyBanner
          anomalies={sessionTree.active_anomalies}
          onSelectSession={handleSelectSession}
        />
      )}

      <div className="live-split-layout">
        <div className="live-split-pane left-tree-pane">
          <div className="pane-section-header">
            <h2 className="pane-title">Active Root Sessions & Lineage</h2>
            <span className="pane-count-badge">
              {sessionTree ? `${sessionTree.root_sessions.length} Root ${sessionTree.root_sessions.length === 1 ? 'Session' : 'Sessions'}` : 'Sessions unavailable'}
            </span>
          </div>
          {sessionTree ? <LiveSessionTree
            tree={sessionTree}
            selectedSessionId={selectedSessionId}
            onSelectSession={handleSelectSession}
            onSelectTicket={onSelectTicket}
          /> : <div className="empty-state" role="status">
            {treeDiagnostic || errorMessage || connectionStatus === 'error' ? 'Live session data unavailable.' : 'Waiting for live session data.'}
          </div>}
        </div>

        <div className="live-split-pane right-inspector-pane">
          <LiveTraceInspector
            key={selectedSessionId ?? 'no-selection'}
            sessionNode={selectedNode}
            trace={traceView?.session_id === selectedSessionId ? traceView : null}
            availableSessionIds={sessionIds(sessionTree?.root_sessions ?? [])}
            onSelectSession={handleSelectSession}
            loading={loadingTrace}
            onClose={() => setSelectedSessionId(null)}
          />
        </div>
      </div>
    </div>
  );
};
