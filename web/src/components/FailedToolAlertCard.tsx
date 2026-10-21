import React, { useState } from 'react';
import type { LivePartFact } from '../domain/live';
import { copyTextToClipboard } from '../utils/clipboard';

export interface FailedToolAlertCardProps {
  readonly part: LivePartFact;
  readonly consecutiveErrors?: number | undefined;
  readonly onInspect?: (() => void) | undefined;
}

export const FailedToolAlertCard: React.FC<FailedToolAlertCardProps> = ({
  part,
  consecutiveErrors,
  onInspect,
}) => {
  const [expanded, setExpanded] = useState(false);
  const [isCopied, setIsCopied] = useState(false);

  const rawSeverity = (part.severity || 'HIGH').toUpperCase();
  const severityClass = rawSeverity.toLowerCase();
  const toolName = part.tool_name || 'unknown_tool';
  const errorMessage = part.output_snippet || part.error_message || 'Execution error without output message';
  const isLong = errorMessage.length > 200;
  const displayError = !expanded && isLong ? `${errorMessage.slice(0, 200)}...` : errorMessage;

  const handleCopyId = async (e: React.MouseEvent) => {
    e.stopPropagation();
    await copyTextToClipboard(part.id);
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
          <span className={`failed-tool-severity severity-${severityClass}`}>{rawSeverity}</span>
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
                aria-label={`Copy Part ID ${part.id}`}
                onClick={handleCopyId}
              >
                {isCopied ? 'Copied' : 'Copy ID'}
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
