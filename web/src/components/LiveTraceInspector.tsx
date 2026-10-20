import React, { useState } from 'react';
import type { LiveMessageFact, LiveSessionNode, LiveTraceView } from '../domain/live';
import { formatTokensK } from '../domain/formatters';
import { FailedToolAlertCard } from './FailedToolAlertCard';

export interface LiveTraceInspectorProps {
  readonly sessionNode: LiveSessionNode | null;
  readonly trace: LiveTraceView | null;
  readonly loading: boolean;
  readonly onClose: () => void;
}

interface MessageTurnProps {
  readonly message: LiveMessageFact;
  readonly turnIndex: number;
}

const LiveMessageTurnItem: React.FC<MessageTurnProps> = ({ message, turnIndex }) => {
  const reasoning = message.reasoning_content;
  const isReasoningLong = (reasoning?.length ?? 0) > 300;
  const [showReasoning, setShowReasoning] = useState(!isReasoningLong);

  const displayReasoning = !showReasoning && isReasoningLong && reasoning
    ? `${reasoning.slice(0, 160)}...`
    : reasoning;

  return (
    <div className={`trace-message-turn role-${message.role.toLowerCase()}`} data-testid={`message-turn-${turnIndex}`}>
      <div className="message-turn-header">
        <div className="message-turn-meta">
          <span className={`turn-role-pill role-${message.role.toLowerCase()}`}>
            {message.role.toUpperCase()}
          </span>
          <span className="turn-timestamp">{message.timestamp}</span>
        </div>
        <div className="turn-token-chips">
          {message.total_tokens > 0 && (
            <span className="token-chip total" title="Total Tokens">
              Total: {formatTokensK(message.total_tokens)}
            </span>
          )}
          {message.context_tokens > 0 && (
            <span className="token-chip context" title="Context Tokens">
              Context: {formatTokensK(message.context_tokens)}
            </span>
          )}
          {message.input_tokens > 0 && (
            <span className="token-chip input" title="Input Tokens">
              Input: {formatTokensK(message.input_tokens)}
            </span>
          )}
          {message.output_tokens > 0 && (
            <span className="token-chip output" title="Output Tokens">
              Output: {formatTokensK(message.output_tokens)}
            </span>
          )}
          {message.reasoning_tokens > 0 && (
            <span className="token-chip reasoning" title="Reasoning Tokens">
              Reasoning: {formatTokensK(message.reasoning_tokens)}
            </span>
          )}
        </div>
      </div>

      {reasoning && reasoning.trim().length > 0 && (
        <div className="message-reasoning-bubble" data-testid="reasoning-bubble">
          <div className="reasoning-header">
            <span className="reasoning-title">Model Deliberation</span>
            <button
              type="button"
              className="reasoning-toggle-btn"
              onClick={() => setShowReasoning((prev) => !prev)}
            >
              {showReasoning ? 'Hide Reasoning' : 'Show Reasoning'}
            </button>
          </div>
          <pre className="reasoning-content">{displayReasoning}</pre>
        </div>
      )}

      {message.content && message.content.trim().length > 0 && (
        <div className="message-body-text">{message.content}</div>
      )}

      {message.parts && message.parts.length > 0 && (
        <div className="message-parts-container">
          {message.parts.map((part) => (
            <div
              key={part.id}
              className={`trace-part-item ${part.is_error ? 'part-error' : 'part-normal'}`}
            >
              <div className="part-header">
                <span className="part-type-badge">{part.type}</span>
                {part.tool_name && (
                  <span className="part-tool-name">{part.tool_name}</span>
                )}
                <span className={`part-status-badge status-${part.status.toLowerCase()}`}>
                  {part.status}
                </span>
              </div>
              {part.input_snippet && (
                <div className="part-snippet input">
                  <span className="part-snippet-label">Input:</span>
                  <pre className="part-code">{part.input_snippet}</pre>
                </div>
              )}
              {part.output_snippet && (
                <div className="part-snippet output">
                  <span className="part-snippet-label">
                    {part.is_error ? 'Error Output:' : 'Output:'}
                  </span>
                  <pre className={`part-code ${part.is_error ? 'error' : ''}`}>
                    {part.output_snippet}
                  </pre>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export const LiveTraceInspector: React.FC<LiveTraceInspectorProps> = ({
  sessionNode,
  trace,
  loading,
  onClose,
}) => {
  if (!sessionNode) {
    return (
      <div className="live-trace-inspector-placeholder" data-testid="live-trace-inspector-placeholder">
        <h3 className="placeholder-title">Select a Session to Inspect</h3>
        <p className="placeholder-desc">
          Click any session node in the hierarchy on the left to inspect real-time execution turns, token accumulations, model deliberations, and tool executions.
        </p>
      </div>
    );
  }

  const hasFailedTools =
    (trace && trace.failed_tool_calls > 0) ||
    (trace && trace.recent_errors && trace.recent_errors.length > 0);

  return (
    <div className="live-trace-inspector-container" data-testid="live-trace-inspector">
      <div className="inspector-header">
        <div className="inspector-title-group">
          <div className="inspector-session-row">
            <span className="inspector-session-id">Session: {sessionNode.id}</span>
            <span className={`node-role-badge role-${sessionNode.role.toLowerCase()}`}>
              {sessionNode.role}
            </span>
            <span className="inspector-model-badge">{sessionNode.model}</span>
          </div>
          <div className="inspector-metrics-row">
            <span className="inspector-stat">
              Agent: <strong>{sessionNode.agent}</strong>
            </span>
            {trace && (
              <>
                <span className="inspector-stat">
                  Tool Calls: <strong>{trace.total_tool_calls}</strong>
                </span>
                <span className={`inspector-stat ${trace.failed_tool_calls > 0 ? 'text-error' : ''}`}>
                  Failed Tools: <strong>{trace.failed_tool_calls}</strong>
                </span>
              </>
            )}
          </div>
        </div>
        <button
          type="button"
          className="inspector-close-btn"
          onClick={onClose}
          aria-label="Close Inspector"
        >
          Close
        </button>
      </div>

      {loading && !trace && (
        <div className="inspector-loading-indicator">
          Loading execution trace for session {sessionNode.id}...
        </div>
      )}

      {hasFailedTools && trace && trace.recent_errors.length > 0 && (
        <section
          className="inspector-failed-tools-section"
          aria-label="Failed Tool Calls"
        >
          <h4 className="failed-tools-heading">
            Failed Tool Calls ({trace.failed_tool_calls})
          </h4>
          <div className="failed-tools-cards-list">
            {trace.recent_errors.map((errorPart, idx) => (
              <FailedToolAlertCard
                key={errorPart.id || `err-${idx}`}
                part={errorPart}
                consecutiveErrors={
                  trace.recent_errors.length > 1 ? trace.recent_errors.length : undefined
                }
              />
            ))}
          </div>
        </section>
      )}

      {trace && (
        <section className="inspector-messages-section" aria-label="Chronological Message Turns">
          <h4 className="messages-heading">
            Execution Message Stream ({trace.messages.length} turns)
          </h4>
          {trace.messages.length === 0 ? (
            <div className="inspector-no-messages">
              No message turns recorded for this session yet.
            </div>
          ) : (
            <div className="inspector-messages-list">
              {trace.messages.map((msg, idx) => (
                <LiveMessageTurnItem
                  key={msg.id || `msg-${idx}`}
                  message={msg}
                  turnIndex={idx}
                />
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  );
};
