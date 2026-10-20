import React, { useState } from 'react';
import type { LiveSessionNode, LiveSessionTree as LiveSessionTreeType } from '../domain/live';
import { formatDurationSeconds } from '../domain/formatters';

export interface LiveSessionTreeProps {
  readonly tree: LiveSessionTreeType;
  readonly selectedSessionId: string | null;
  readonly onSelectSession: (sessionId: string) => void;
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
}

interface NodeItemProps {
  readonly node: LiveSessionNode;
  readonly selectedSessionId: string | null;
  readonly onSelectSession: (sessionId: string) => void;
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
  readonly collapsedNodes: ReadonlySet<string>;
  readonly onToggleCollapse: (nodeId: string) => void;
  readonly depth: number;
}

const LiveSessionNodeItem: React.FC<NodeItemProps> = ({
  node,
  selectedSessionId,
  onSelectSession,
  onSelectTicket,
  collapsedNodes,
  onToggleCollapse,
  depth,
}) => {
  const isSelected = selectedSessionId === node.id;
  const isCollapsed = collapsedNodes.has(node.id);
  const hasSubSessions = node.sub_sessions && node.sub_sessions.length > 0;

  const startMs = Date.parse(node.time_created);
  const endMs = Date.parse(node.time_updated);
  const durationSec =
    !Number.isNaN(startMs) && !Number.isNaN(endMs) && endMs >= startMs
      ? Math.floor((endMs - startMs) / 1000)
      : 0;
  const formattedDuration = formatDurationSeconds(durationSec);

  const confidencePct = Math.round(node.ticket_confidence * 100);

  return (
    <div
      className={`live-session-node-wrapper depth-${depth}`}
      data-testid={`session-node-${node.id}`}
    >
      <div
        className={`live-session-node-card ${isSelected ? 'selected' : ''}`}
        tabIndex={0}
        onClick={() => onSelectSession(node.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onSelectSession(node.id);
          }
        }}
        aria-selected={isSelected}
      >
        <div className="node-header">
          <div className="node-title-group">
            <span className="node-agent-name">{node.agent}</span>
            <span className={`node-role-badge role-${node.role.toLowerCase()}`}>
              {node.role}
            </span>
            <span className="node-model-name">{node.model}</span>
          </div>
          <div className="node-status-group">
            <span className={`node-status-pill status-${node.status.toLowerCase()}`}>
              {node.status}
            </span>
            <span className="node-duration-pill">{formattedDuration}</span>
          </div>
        </div>

        <div className="node-body">
          <div className="node-task-title">{node.title}</div>
          <div className="node-badges-row">
            {node.guessed_ticket_id && (
              <span
                className="guessed-ticket-badge"
                role="button"
                tabIndex={0}
                onClick={(e) => {
                  e.stopPropagation();
                  if (node.guessed_ticket_id) {
                    onSelectTicket?.(node.guessed_ticket_id);
                  }
                }}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.stopPropagation();
                    e.preventDefault();
                    if (node.guessed_ticket_id) {
                      onSelectTicket?.(node.guessed_ticket_id);
                    }
                  }
                }}
                title={`Candidate confidence: ${confidencePct}%`}
              >
                <span className="ticket-badge-id">{node.guessed_ticket_id}</span>
                <span className="ticket-badge-confidence">{confidencePct}% confidence</span>
              </span>
            )}

            {node.anomalies && node.anomalies.length > 0 && (
              <span className="node-anomaly-pill">
                {node.anomalies.length} {node.anomalies.length === 1 ? 'anomaly' : 'anomalies'}
              </span>
            )}
          </div>
        </div>

        {hasSubSessions && (
          <div className="node-footer">
            <button
              type="button"
              className="sub-sessions-toggle-btn"
              onClick={(e) => {
                e.stopPropagation();
                onToggleCollapse(node.id);
              }}
            >
              {isCollapsed
                ? `Expand Sub-Sessions (${node.sub_sessions.length})`
                : `Collapse Sub-Sessions (${node.sub_sessions.length})`}
            </button>
          </div>
        )}
      </div>

      {hasSubSessions && !isCollapsed && (
        <div className="live-sub-sessions-container">
          {node.sub_sessions.map((subNode) => (
            <LiveSessionNodeItem
              key={subNode.id}
              node={subNode}
              selectedSessionId={selectedSessionId}
              onSelectSession={onSelectSession}
              onSelectTicket={onSelectTicket}
              collapsedNodes={collapsedNodes}
              onToggleCollapse={onToggleCollapse}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export const LiveSessionTree: React.FC<LiveSessionTreeProps> = ({
  tree,
  selectedSessionId,
  onSelectSession,
  onSelectTicket,
}) => {
  const [collapsedNodes, setCollapsedNodes] = useState<ReadonlySet<string>>(new Set());

  const handleToggleCollapse = (nodeId: string) => {
    setCollapsedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  if (!tree.root_sessions || tree.root_sessions.length === 0) {
    return (
      <div className="live-session-tree-empty" data-testid="live-session-tree-empty">
        <h3 className="empty-title">No Active Sessions</h3>
        <p className="empty-message">
          There are no active agent sessions currently running in the environment.
        </p>
      </div>
    );
  }

  return (
    <div className="live-session-tree-container" data-testid="live-session-tree">
      <div className="live-session-tree-roots">
        {tree.root_sessions.map((rootNode) => (
          <LiveSessionNodeItem
            key={rootNode.id}
            node={rootNode}
            selectedSessionId={selectedSessionId}
            onSelectSession={onSelectSession}
            onSelectTicket={onSelectTicket}
            collapsedNodes={collapsedNodes}
            onToggleCollapse={handleToggleCollapse}
            depth={0}
          />
        ))}
      </div>
    </div>
  );
};
