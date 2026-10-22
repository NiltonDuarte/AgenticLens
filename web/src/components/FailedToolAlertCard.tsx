import React, { useState } from 'react';
import type { LivePartFact, FailureSeverity } from '../domain/live';
import type { AgentRole } from '../domain/types';
import { copyTextToClipboard } from '../utils/clipboard';
import { generateExecutionIssuePrompt } from '../utils/debugPrompt';

export interface FailedToolAlertCardProps {
  readonly part: LivePartFact;
  readonly consecutiveErrors?: number | undefined;
  readonly onInspect?: (() => void) | undefined;
  readonly sessionId?: string | undefined;
  readonly turn?: number | undefined;
  readonly role?: AgentRole | undefined;
  readonly model?: string | undefined;
  readonly messageId?: string | undefined;
}

export const FailedToolAlertCard: React.FC<FailedToolAlertCardProps> = ({
  part,
  consecutiveErrors,
  onInspect,
  sessionId,
  turn,
  role,
  model,
}) => {
  const [expanded, setExpanded] = useState(false);
  const [isCopied, setIsCopied] = useState(false);

  const effectiveSeverity: FailureSeverity = part.severity || 'high';
  const severityClass = effectiveSeverity.toLowerCase();
  const toolName = part.tool_name || 'unknown_tool';
  const errorMessage = part.output_snippet || part.error_message || 'Execution error without output message';
  const isLong = errorMessage.length > 200;
  const displayError = !expanded && isLong ? `${errorMessage.slice(0, 200)}...` : errorMessage;

  const handleCopyPrompt = async (e: React.MouseEvent) => {
    e.stopPropagation();
    const prompt = generateExecutionIssuePrompt({
      summary:
        consecutiveErrors !== undefined && consecutiveErrors > 1
          ? `Consecutive failures on \`${toolName}\``
          : `Tool execution failure on \`${toolName}\``,
      turn,
      sessionId: sessionId || 'unknown-session',
      role,
      model,
      description:
        consecutiveErrors !== undefined && consecutiveErrors > 1
          ? `The agent experienced ${consecutiveErrors} consecutive failures while executing \`${toolName}\`.`
          : `The agent encountered an execution failure while invoking \`${toolName}\`.`,
      partId: part.id || undefined,
      errorDetails: errorMessage,
    });
    await copyTextToClipboard(prompt);
    setIsCopied(true);
    window.setTimeout(() => {
      setIsCopied(false);
    }, 2000);
  };

  return (
    <div
      className={`failed-tool-alert-card severity-${severityClass}`}
      data-testid="failed-tool-alert-card"
    >
      <div className="failed-tool-header">
        <div className="failed-tool-title-group">
          <span className="failed-tool-name">{toolName}</span>
          <span className="failed-tool-badge">FAILED</span>
          <span className={`failed-tool-severity severity-${severityClass}`}>{effectiveSeverity.toUpperCase()}</span>
          {consecutiveErrors !== undefined && consecutiveErrors > 1 && (
            <span className="failed-tool-streak">
              {consecutiveErrors} Consecutive Failures
            </span>
          )}
          {part.id && (
            <div className="failed-tool-id-group">
              <span className="failed-tool-id mono">{part.id}</span>
              <button
                type="button"
                className="btn-copy-id"
                aria-label={`Copy execution issue debug prompt for ${toolName}`}
                onClick={handleCopyPrompt}
              >
                {isCopied ? 'Copied' : 'Copy Prompt'}
              </button>
            </div>
          )}
        </div>
        {onInspect && (
          <button
            type="button"
            className="failed-tool-inspect-btn"
            onClick={onInspect}
          >
            Inspect
          </button>
        )}
      </div>

      {part.input_snippet && (
        <div className="failed-tool-input-section">
          <span className="failed-tool-input-label">Command / Input:</span>
          <pre className="failed-tool-code-snippet">{part.input_snippet}</pre>
        </div>
      )}

      <div className="failed-tool-output-section">
        <span className="failed-tool-output-label">Error Output:</span>
        <pre className="failed-tool-error-snippet">{displayError}</pre>
        {isLong && (
          <button
            type="button"
            className="failed-tool-expand-btn"
            onClick={() => setExpanded((prev) => !prev)}
          >
            {expanded ? 'Collapse Output' : 'Expand Output'}
          </button>
        )}
      </div>
    </div>
  );
};
