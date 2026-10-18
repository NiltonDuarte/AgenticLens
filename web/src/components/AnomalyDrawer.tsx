import React, { useEffect } from 'react';
import type { Anomaly, EventFact } from '../types/telemetry';
import { AnomalyBadge } from './AnomalyBadge';

export interface AnomalyDrawerProps {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly anomaly?: Anomaly | null | undefined;
  readonly event?: EventFact | null | undefined;
}

export const AnomalyDrawer: React.FC<AnomalyDrawerProps> = ({
  isOpen,
  onClose,
  anomaly,
  event,
}) => {
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) {
    return null;
  }

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const title =
    anomaly?.title ??
    (event?.status === 'error' ? 'Tool Execution Failure' : 'Execution Details');
  const errorMessage = event?.error_message;

  const ctx = anomaly?.context;
  const turn = ctx?.turn ?? event?.turn;
  const sessionId = ctx?.session_id ?? event?.session_id;
  const toolName = ctx?.tool_name ?? event?.tool_name;
  const role = ctx?.role ?? event?.role;
  const secondaryRole = ctx?.secondary_role;
  const count = ctx?.count;
  const totalCalls = ctx?.total_calls;
  const threshold = ctx?.threshold;
  const tokenDelta = ctx?.token_delta;
  const contextTokens = ctx?.context_tokens ?? event?.context_tokens;
  const failureRatePct = ctx?.failure_rate_pct;
  const durationMs = event?.duration_ms;
  const timestamp = event?.timestamp;

  return (
    <div
      className="anomaly-drawer-backdrop"
      onClick={handleBackdropClick}
      data-testid="anomaly-drawer-backdrop"
    >
      <div
        className="anomaly-drawer"
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <div className="drawer-header">
          <div className="drawer-header-left">
            <h3 className="drawer-title">{title}</h3>
            {anomaly && (
              <AnomalyBadge severity={anomaly.severity} size="sm" />
            )}
          </div>
          <button
            type="button"
            className="drawer-close-btn"
            onClick={onClose}
            aria-label="Close"
          >
            Close
          </button>
        </div>

        <div className="drawer-body">
          {/* Rule Diagnosis Section */}
          {anomaly && (
            <section className="drawer-section">
              <h4 className="drawer-section-title">Rule Diagnosis</h4>
              <div className="drawer-rule-id">
                <span className="drawer-label">Rule ID:</span>
                <span className="drawer-mono-val">{anomaly.rule_id}</span>
              </div>
              {anomaly.description && (
                <p className="drawer-description">{anomaly.description}</p>
              )}
            </section>
          )}

          {/* Diagnostic Context Metadata Grid */}
          <section className="drawer-section">
            <h4 className="drawer-section-title">Diagnostic Context</h4>
            <div className="drawer-meta-grid">
              {turn !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Turn Number</span>
                  <span className="drawer-val mono">Turn {turn}</span>
                </div>
              )}
              {sessionId && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Session ID</span>
                  <span className="drawer-val mono">{sessionId}</span>
                </div>
              )}
              {toolName && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Tool Name</span>
                  <span className="drawer-val mono">{toolName}</span>
                </div>
              )}
              {role && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Subagent Role</span>
                  <span className="drawer-val">{role}</span>
                </div>
              )}
              {secondaryRole && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Secondary Role</span>
                  <span className="drawer-val">{secondaryRole}</span>
                </div>
              )}
              {count !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Failure / Loop Count</span>
                  <span className="drawer-val mono">{count}</span>
                </div>
              )}
              {totalCalls !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Total Calls</span>
                  <span className="drawer-val mono">{totalCalls}</span>
                </div>
              )}
              {failureRatePct !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Failure Rate</span>
                  <span className="drawer-val mono">{failureRatePct.toFixed(1)}%</span>
                </div>
              )}
              {threshold && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Trigger Threshold</span>
                  <span className="drawer-val mono">{threshold}</span>
                </div>
              )}
              {tokenDelta !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Token Delta</span>
                  <span className="drawer-val mono">+{tokenDelta.toLocaleString()} tokens</span>
                </div>
              )}
              {contextTokens !== undefined && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Context Tokens</span>
                  <span className="drawer-val mono">{contextTokens.toLocaleString()} tokens</span>
                </div>
              )}
              {durationMs !== undefined && durationMs !== null && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Execution Duration</span>
                  <span className="drawer-val mono">{durationMs}ms</span>
                </div>
              )}
              {timestamp && (
                <div className="drawer-meta-item">
                  <span className="drawer-label">Timestamp</span>
                  <span className="drawer-val mono">{timestamp}</span>
                </div>
              )}
            </div>
          </section>

          {/* Monospace Error Payload Box */}
          {errorMessage && (
            <section className="drawer-section">
              <h4 className="drawer-section-title">Error Payload</h4>
              <div className="drawer-error-box">
                <pre className="drawer-error-content">{errorMessage}</pre>
              </div>
            </section>
          )}
        </div>
      </div>
    </div>
  );
};
