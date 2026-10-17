import React, { useState } from 'react';
import type { ThreadExecutionRecord } from '../domain/types';
import { formatCurrencyUSD, formatPercentage, formatTokensK } from '../domain/formatters';

export interface ThreadViewerProps {
  readonly threads: readonly ThreadExecutionRecord[];
}

export const ThreadViewer: React.FC<ThreadViewerProps> = ({ threads }) => {
  const [expandedSessions, setExpandedSessions] = useState<ReadonlySet<string>>(() => {
    // Expand the first thread by default
    if (threads.length > 0 && threads[0]) {
      return new Set([threads[0].session_id]);
    }
    return new Set();
  });

  if (!threads || threads.length === 0) {
    return (
      <div className="thread-viewer-card empty">
        <h4 className="section-title">Session Execution Threads</h4>
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
        <h4 className="section-title">Session Execution Threads</h4>
        <span className="thread-count-badge">{threads.length} Threads</span>
      </div>

      <div className="threads-accordion-list">
        {threads.map((thread) => {
          const isExpanded = expandedSessions.has(thread.session_id);
          return (
            <div key={thread.session_id} className="thread-accordion-item">
              <button
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
                          <th>Role</th>
                          <th>Model</th>
                          <th className="num-col">Invocations</th>
                          <th className="num-col">Input</th>
                          <th className="num-col">Output</th>
                          <th className="num-col">Cache Read</th>
                          <th className="num-col">Cache Hit</th>
                          <th className="num-col">Cost</th>
                          <th className="num-col">Peak Context</th>
                          <th>Context Health</th>
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
