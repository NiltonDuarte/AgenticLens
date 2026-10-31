import React, { useEffect, useLayoutEffect, useRef, useState, useCallback } from 'react';
import type {
  LiveAnomalyAlert,
  LiveSessionNode,
  LiveSessionTree as LiveSessionTreeType,
} from '../domain/live';
import {
  type ConnectionStatus,
  fetchLiveSessionTree,
  subscribeToLiveStream,
} from '../api/liveApi';
import { LiveSessionTree } from '../components/LiveSessionTree';
import { LiveTraceInspector } from '../components/LiveTraceInspector';
import { useLiveTrace } from '../hooks/useLiveTrace';
import { LiveAnomalyBanner } from '../components/LiveAnomalyBanner';

export interface LiveObservabilityProps {
  readonly onSelectTicket: (ticketId: string) => void;
}

const EMPTY_TREE: LiveSessionTreeType = {
  root_sessions: [],
  last_refreshed_at: '',
  total_active_sessions: 0,
  active_anomalies: [],
};

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

export const LiveObservability: React.FC<LiveObservabilityProps> = ({
  onSelectTicket,
}) => {
  const [isPaused, setIsPaused] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('connecting');
  const [sessionTree, setSessionTree] = useState<LiveSessionTreeType>(EMPTY_TREE);
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  const selectedNode = selectedSessionId
    ? findSessionNode(sessionTree.root_sessions, selectedSessionId)
    : null;
  const { trace: traceView, loading: loadingTrace, error: traceError, onTraceUpdate, captureRefresh } =
    useLiveTrace(selectedSessionId, isPaused, selectedNode);

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

  const manualOwner = useRef<object | null>(null);
  useLayoutEffect(() => {
    manualOwner.current = {};
    return () => { manualOwner.current = null; };
  }, [selectedSessionId, isPaused, retryCount]);

  const handleManualRefresh = useCallback(async () => {
    const owner = manualOwner.current;
    const refreshTrace = captureRefresh();
    refreshTrace();
    try {
      const tree = await fetchLiveSessionTree();
      if (manualOwner.current !== owner) return;
      setSessionTree(tree);
      setErrorMessage(null);
    } catch (err: unknown) {
      if (manualOwner.current !== owner) return;
      setErrorMessage(err instanceof Error ? err.message : String(err));
    }
  }, [captureRefresh]);

  // Reactive streaming lifecycle
  useEffect(() => {
    if (isPaused) {
      setConnectionStatus('paused');
      return;
    }

    let subscribed = true;
    const unsubscribe = subscribeToLiveStream({
      onSessionTree: (tree) => {
        if (!subscribed) return;
        setSessionTree(tree);
        setErrorMessage(null);
      },
      onAnomalyAlert: (alert: LiveAnomalyAlert) => {
        if (!subscribed) return;
        setSessionTree((prev) => {
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
      onTraceUpdate: (trace) => { if (subscribed) onTraceUpdate(trace); },
      onStatusChange: (status) => {
        if (!subscribed) return;
        setConnectionStatus(status);
      },
      onError: (err) => {
        if (!subscribed) return;
        setErrorMessage(err.message);
      },
    });

    return () => {
      subscribed = false;
      unsubscribe();
    };
  }, [isPaused, retryCount, onTraceUpdate]);

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
            {connectionStatus === 'error' && 'Connection Error'}
          </span>

          <span className="status-bar-stat">
            Active Sessions: <strong>{sessionTree.total_active_sessions}</strong>
          </span>

          <span className="status-bar-stat">
            Active Anomalies: <strong>{sessionTree.active_anomalies.length}</strong>
          </span>

          {sessionTree.last_refreshed_at && (
            <span className="status-bar-stat refreshed-time">
              Refreshed: {sessionTree.last_refreshed_at}
            </span>
          )}

          {(traceError || errorMessage) && connectionStatus !== 'error' && (
            <span className="status-bar-error" role="alert">
              Error: {traceError || errorMessage}
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
              Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running.
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

      {sessionTree.active_anomalies && sessionTree.active_anomalies.length > 0 && (
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
              {sessionTree.root_sessions.length} Root {sessionTree.root_sessions.length === 1 ? 'Session' : 'Sessions'}
            </span>
          </div>
          <LiveSessionTree
            tree={sessionTree}
            selectedSessionId={selectedSessionId}
            onSelectSession={handleSelectSession}
            onSelectTicket={onSelectTicket}
          />
        </div>

        <div className="live-split-pane right-inspector-pane">
          <LiveTraceInspector
            key={selectedSessionId ?? 'no-selection'}
            sessionNode={selectedNode}
            trace={traceView}
            loading={loadingTrace}
            onClose={() => setSelectedSessionId(null)}
          />
        </div>
      </div>
    </div>
  );
};
