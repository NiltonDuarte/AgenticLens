import React, { useState, useMemo } from 'react';
import type { EventFact, Anomaly } from '../types/telemetry';
import { formatTokensK } from '../domain/formatters';
import { AnomalyBadge } from './AnomalyBadge';

export interface TraceWaterfallProps {
  readonly events: readonly EventFact[];
  readonly anomalies?: readonly Anomaly[] | undefined;
  readonly selectedEvent?: EventFact | null | undefined;
  readonly onSelectEvent?: ((event: EventFact, anomaly?: Anomaly | undefined) => void) | undefined;
}

type FilterMode = 'all' | 'tools' | 'errors_anomalies';

const findMatchingAnomaly = (
  event: EventFact,
  anomalies?: readonly Anomaly[]
): Anomaly | undefined => {
  if (!anomalies || anomalies.length === 0) {
    return undefined;
  }

  return anomalies.find((a) => {
    const ctx = a.context;
    // 1. Tool call match: tool_name matches and turn or session_id matches
    if (
      ctx.tool_name &&
      event.tool_name &&
      ctx.tool_name.toLowerCase() === event.tool_name.toLowerCase()
    ) {
      if (ctx.turn !== undefined && event.turn !== undefined) {
        return ctx.turn === event.turn;
      }
      if (ctx.session_id && event.session_id) {
        return ctx.session_id === event.session_id;
      }
      return true;
    }

    // 2. Turn and session match (e.g. context saturation spike)
    if (ctx.turn !== undefined && ctx.turn === event.turn) {
      if (ctx.session_id && event.session_id) {
        return ctx.session_id === event.session_id;
      }
      return true;
    }

    // 3. Role match
    if (
      ctx.role &&
      event.role &&
      (ctx.role.toLowerCase() === event.role.toLowerCase() ||
        ctx.secondary_role?.toLowerCase() === event.role.toLowerCase())
    ) {
      if (ctx.turn !== undefined) {
        return ctx.turn === event.turn;
      }
    }

    return false;
  });
};

const formatDurationMs = (ms: number): string => {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const sec = ms / 1000;
  return `${Number(sec.toFixed(2))}s`;
};

const formatEventTypeLabel = (type: string): string => {
  switch (type) {
    case 'tool_call':
      return 'Tool Call';
    case 'turn_message':
      return 'Turn Message';
    case 'compaction':
      return 'Compaction';
    default:
      return type;
  }
};

const isSameEvent = (a?: EventFact | null, b?: EventFact | null): boolean => {
  if (!a || !b) return false;
  return (
    a.timestamp === b.timestamp &&
    a.turn === b.turn &&
    a.session_id === b.session_id &&
    a.event_type === b.event_type &&
    a.tool_name === b.tool_name
  );
};

export const TraceWaterfall: React.FC<TraceWaterfallProps> = ({
  events,
  anomalies,
  selectedEvent,
  onSelectEvent,
}) => {
  const [filterMode, setFilterMode] = useState<FilterMode>('errors_anomalies');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [localSelectedEvent, setLocalSelectedEvent] = useState<EventFact | null>(null);

  const activeSelectedEvent = selectedEvent !== undefined ? selectedEvent : localSelectedEvent;

  const handleItemSelect = (ev: EventFact, matchingAnomaly?: Anomaly) => {
    setLocalSelectedEvent(ev);
    const isFlagged = ev.status === 'error' || Boolean(matchingAnomaly);
    onSelectEvent?.(ev, isFlagged ? matchingAnomaly : undefined);
  };

  // Sorted events chronologically
  const sortedEvents = useMemo(() => {
    if (!events || events.length === 0) return [];
    return [...events].sort((a, b) => {
      const cmp = a.timestamp.localeCompare(b.timestamp);
      if (cmp !== 0) return cmp;
      return a.turn - b.turn;
    });
  }, [events]);

  // Unique roles for filter dropdown
  const availableRoles = useMemo(() => {
    if (!events) return [];
    const roleSet = new Set<string>();
    for (const e of events) {
      if (e.role) {
        roleSet.add(e.role);
      }
    }
    return Array.from(roleSet).sort();
  }, [events]);

  // Filtered event list
  const filteredEvents = useMemo(() => {
    return sortedEvents.filter((e) => {
      // Role filter
      if (roleFilter !== 'all' && e.role !== roleFilter) {
        return false;
      }

      // Filter mode
      if (filterMode === 'tools') {
        return e.event_type === 'tool_call';
      }

      if (filterMode === 'errors_anomalies') {
        const isError = e.status === 'error';
        const hasAnomaly = !!findMatchingAnomaly(e, anomalies);
        return isError || hasAnomaly;
      }

      return true;
    });
  }, [sortedEvents, roleFilter, filterMode, anomalies]);

  if (!events || events.length === 0) {
    return (
      <div className="trace-waterfall-card">
        <div className="trace-empty-state">
          Zero chronological events recorded for this execution trace.
        </div>
      </div>
    );
  }

  return (
    <div className="trace-waterfall-card">
      <div className="waterfall-header">
        <div className="waterfall-title-group">
          <h3 className="waterfall-title">Execution Trace Waterfall</h3>
          <span className="waterfall-counter">
            Showing {filteredEvents.length} of {events.length} events
          </span>
          <span className="waterfall-hint">
            Click any error or anomaly to inspect failure details
          </span>
        </div>

        {/* Filter Toolbar */}
        <div className="waterfall-toolbar">
          <div className="filter-button-group" role="group" aria-label="Filter events by type">
            <button
              type="button"
              className={`filter-btn ${filterMode === 'all' ? 'active' : ''}`}
              onClick={() => setFilterMode('all')}
            >
              All Events
            </button>
            <button
              type="button"
              className={`filter-btn ${filterMode === 'tools' ? 'active' : ''}`}
              onClick={() => setFilterMode('tools')}
            >
              Tool Calls Only
            </button>
            <button
              type="button"
              className={`filter-btn ${filterMode === 'errors_anomalies' ? 'active' : ''}`}
              onClick={() => setFilterMode('errors_anomalies')}
              aria-label="Show only errors and anomalies"
            >
              Errors and Anomalies Only
            </button>
          </div>

          <div className="role-filter-group">
            <label htmlFor="trace-role-filter" className="role-filter-label">
              Role:
            </label>
            <select
              id="trace-role-filter"
              aria-label="Filter by subagent role"
              className="role-filter-select"
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
            >
              <option value="all">All Roles</option>
              {availableRoles.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Waterfall Timeline Track */}
      {filteredEvents.length === 0 ? (
        <div className="trace-empty-state">
          No events match the selected filters.
        </div>
      ) : (
        <div className="waterfall-timeline">
          {filteredEvents.map((ev, idx) => {
            const matchingAnomaly = findMatchingAnomaly(ev, anomalies);
            const isSelected = isSameEvent(activeSelectedEvent, ev);
            const hasError = ev.status === 'error';
            const isIssue = ev.status === 'error' || Boolean(matchingAnomaly);
            const roleClass = `role-pill ${ev.role.toLowerCase()}`;
            const itemKey = `${ev.timestamp}-${ev.turn}-${ev.session_id}-${idx}`;
            const itemClasses = `timeline-item ${isSelected ? 'selected' : ''} ${hasError ? 'item-error' : ''} ${matchingAnomaly ? 'item-anomaly' : ''} ${isIssue ? 'item-interactive' : ''}`
              .replace(/\s+/g, ' ')
              .trim();

            return (
              <div
                key={itemKey}
                data-testid={`timeline-item-${ev.turn}`}
                className={itemClasses}
                onClick={() => handleItemSelect(ev, matchingAnomaly)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleItemSelect(ev, matchingAnomaly);
                  }
                }}
                aria-label={`Event Turn ${ev.turn} ${ev.role} ${ev.event_type}`}
              >
                <div className="timeline-rail">
                  <div className={`timeline-node ${hasError ? 'node-error' : ''}`} />
                  {idx < filteredEvents.length - 1 && <div className="timeline-connector" />}
                </div>

                <div className="timeline-content">
                  <div className="timeline-top-row">
                    <span className="turn-pill">Turn {ev.turn}</span>
                    <span className={roleClass}>{ev.role}</span>
                    <span className="event-type-badge">
                      {formatEventTypeLabel(ev.event_type)}
                    </span>
                    <span className="timeline-timestamp mono">{ev.timestamp}</span>

                    {matchingAnomaly && (
                      <AnomalyBadge
                        severity={matchingAnomaly.severity}
                        label={matchingAnomaly.severity}
                        size="sm"
                        title="Click to open Anomaly Inspector"
                        onClick={(e) => {
                          e.stopPropagation();
                          setLocalSelectedEvent(ev);
                          onSelectEvent?.(ev, matchingAnomaly);
                        }}
                      />
                    )}
                  </div>

                  <div className="timeline-details-row">
                    {ev.event_type === 'tool_call' && (
                      <div className="tool-details-group">
                        <span className="tool-label">Tool:</span>
                        <span className="tool-name mono">{ev.tool_name ?? 'unknown'}</span>
                        <span
                          className={`status-pill ${hasError ? 'status-error' : 'status-success'}`}
                        >
                          {hasError ? 'Error' : 'Success'}
                        </span>
                        {ev.duration_ms !== undefined && ev.duration_ms !== null && (
                          <span className="tool-duration mono">{formatDurationMs(ev.duration_ms)}</span>
                        )}
                        {hasError && (
                          <button
                            type="button"
                            className="btn-inspect-failure"
                            title="Click to open Tool Execution Failure window"
                            aria-label={`Inspect failure for turn ${ev.turn}`}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleItemSelect(ev, matchingAnomaly);
                            }}
                          >
                            Inspect Failure →
                          </button>
                        )}
                      </div>
                    )}

                    {ev.event_type !== 'tool_call' && hasError && (
                      <button
                        type="button"
                        className="btn-inspect-failure"
                        title="Click to open Tool Execution Failure window"
                        aria-label={`Inspect failure for turn ${ev.turn}`}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleItemSelect(ev, matchingAnomaly);
                        }}
                      >
                        Inspect Failure →
                      </button>
                    )}

                    {ev.event_type === 'compaction' && (
                      <div className="compaction-details-group">
                        <span className="compaction-note">Context window compacted</span>
                      </div>
                    )}

                    <div className="timeline-context-tokens">
                      <span className="context-tokens-tag mono">
                        {formatTokensK(ev.context_tokens)} tokens
                      </span>
                    </div>
                  </div>

                  {ev.error_message && (
                    <div
                      className="timeline-error-preview mono"
                      title="Click to open Tool Execution Failure window"
                    >
                      {ev.error_message}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
