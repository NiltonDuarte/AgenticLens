import React, { useEffect, useRef, useState } from 'react';
import type { ThreadExecutionRecord } from '../domain/types';
import { formatCurrencyUSD, formatPercentage, formatTokensK } from '../domain/formatters';
import { InfoTooltip } from './InfoTooltip';

export interface ThreadViewerProps {
  readonly focusSession?: { readonly id: string } | undefined;
  readonly threads: readonly ThreadExecutionRecord[];
}

export const ThreadViewer: React.FC<ThreadViewerProps> = ({ threads, focusSession }) => {
  const [expandedSessions, setExpandedSessions] = useState<ReadonlySet<string>>(() => {
    // Expand the first thread by default
    if (threads.length > 0 && threads[0]) {
      return new Set([threads[0].session_id]);
    }
    return new Set();
  });

  const focusedHeader = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (!focusSession || !threads.some((thread) => thread.session_id === focusSession.id)) return;
    setExpandedSessions((previous) => new Set([...previous, focusSession.id]));
    focusedHeader.current?.focus();
    focusedHeader.current?.scrollIntoView?.({ block: 'nearest' });
  }, [focusSession, threads]);

  if (!threads || threads.length === 0) {
    return (
      <div className="thread-viewer-card empty">
        <div className="thread-viewer-header">
          <div className="thread-title-group">
            <h4 className="section-title">Session Execution Threads</h4>
            <InfoTooltip
              content="Hierarchical session execution threads tracking multi-agent turns, context compactions, and model usage per role."
              ariaLabel="Information about Session Execution Threads"
            />
          </div>
        </div>
        <div className="empty-threads-note">No session execution threads recorded for this workspace.</div>
      </div>
    );
  }

  const toggleSession = (sessionId: string) => {
    setExpandedSessions((prev) => {
      const next = new Set(prev);
      if (next.has(sessionId)) {
        next.delete(sessionId);
      } else {
        next.add(sessionId);
      }
      return next;
    });
  };

  return (
    <div className="thread-viewer-card">
      <div className="thread-viewer-header">
        <div className="thread-title-group">
          <h4 className="section-title">Session Execution Threads</h4>
          <InfoTooltip
            content="Hierarchical session execution threads tracking multi-agent turns, context compactions, and model usage per role."
            ariaLabel="Information about Session Execution Threads"
          />
        </div>
        <span className="thread-count-badge">{threads.length} Threads</span>
      </div>

      <div className="threads-accordion-list">
        {threads.map((thread) => {
          const isExpanded = expandedSessions.has(thread.session_id);
          return (
            <div key={thread.session_id} className="thread-accordion-item">
              <button
                ref={thread.session_id === focusSession?.id ? focusedHeader : undefined}
                type="button"
                className={`thread-header-btn ${isExpanded ? 'open' : ''}`}
                onClick={() => toggleSession(thread.session_id)}
                aria-expanded={isExpanded}
              >
                <div className="thread-header-main">
                  <span className="thread-id-mono">{thread.session_id}</span>
                  <span className="thread-user-tag">User: {thread.user}</span>
                  <span className="thread-compactions-tag">
                    Compactions: {thread.compactions}
                  </span>
                </div>
                <span className="thread-chevron">{isExpanded ? '[-]' : '[+]'}</span>
              </button>

              {isExpanded && (
                <div className="thread-body">
                  <div className="table-responsive">
                    <table className="thread-models-table">
                      <thead>
                        <tr>
                          <th>
                            <span className="th-content">
                              <span>Role</span>
                              <InfoTooltip
                                content="Agent persona role executing in this thread"
                                placement="bottom"
                                ariaLabel="Information about Role"
                              />
                            </span>
                          </th>
                          <th>
                            <span className="th-content">
                              <span>Model</span>
                              <InfoTooltip
                                content="Model family and version used in this session"
                                placement="bottom"
                                ariaLabel="Information about Model"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Invocations</span>
                              <InfoTooltip
                                content="Total count of model invocations in this thread"
                                placement="bottom"
                                ariaLabel="Information about Invocations"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Input</span>
                              <InfoTooltip
                                content="Input prompt tokens sent to the model"
                                placement="bottom"
                                ariaLabel="Information about Input"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Output</span>
                              <InfoTooltip
                                content="Completion output tokens generated by the model"
                                placement="bottom"
                                ariaLabel="Information about Output"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Cache Read</span>
                              <InfoTooltip
                                content="Prompt tokens read directly from prompt cache"
                                placement="bottom"
                                ariaLabel="Information about Cache Read"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Cache Hit</span>
                              <InfoTooltip
                                content="Percentage of prompt context served from cache"
                                placement="bottom"
                                ariaLabel="Information about Cache Hit"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Cost</span>
                              <InfoTooltip
                                content="Total inference cost in US dollars for this model"
                                placement="bottom"
                                ariaLabel="Information about Cost"
                              />
                            </span>
                          </th>
                          <th className="num-col">
                            <span className="th-content">
                              <span>Peak Context</span>
                              <InfoTooltip
                                content="Maximum single-turn context window size encountered in this thread"
                                placement="bottom"
                                ariaLabel="Information about Peak Context"
                              />
                            </span>
                          </th>
                          <th>
                            <span className="th-content">
                              <span>Context Health</span>
                              <InfoTooltip
                                content="Context window saturation status against healthy ceiling (HEALTHY, ELEVATED, BLOATED, CRITICAL)"
                                placement="bottom"
                                ariaLabel="Information about Context Health"
                              />
                            </span>
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {thread.roles.flatMap((r) =>
                          r.models.map((m, mIdx) => (
                            <tr key={`${r.role}-${m.model}-${mIdx}`}>
                              <td className="role-col">{r.role}</td>
                              <td className="model-col">{m.model}</td>
                              <td className="num-col">{m.invocations}</td>
                              <td className="num-col">{formatTokensK(m.input_tokens)}</td>
                              <td className="num-col">{formatTokensK(m.output_tokens)}</td>
                              <td className="num-col">{formatTokensK(m.cache_read_tokens)}</td>
                              <td className="num-col">{formatPercentage(m.cache_hit_rate_pct)}</td>
                              <td className="num-col">{formatCurrencyUSD(m.cost_usd)}</td>
                              <td className="num-col">{formatTokensK(m.peak_context_tokens)}</td>
                              <td>
                                <span className={`health-pill mini ${m.context_health.toLowerCase()}`}>
                                  {m.context_health}
                                </span>
                              </td>
                            </tr>
                          ))
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};
