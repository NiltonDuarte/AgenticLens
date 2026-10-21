import React, { useState, useMemo } from 'react';
import type { EventFact, Anomaly, DataSource } from '../types/telemetry';
import { formatTokensK } from '../domain/formatters';
import { AnomalyBadge } from './AnomalyBadge';
import { MultiSelectDropdown, type MultiSelectOption } from './MultiSelectDropdown';
import { copyTextToClipboard } from '../utils/clipboard';

export interface TraceWaterfallProps {
  readonly events: readonly EventFact[];
  readonly anomalies?: readonly Anomaly[] | undefined;
  readonly selectedEvent?: EventFact | null | undefined;
  readonly onSelectEvent?: ((event: EventFact, anomaly?: Anomaly | undefined) => void) | undefined;
  readonly isHydrated?: boolean | undefined;
  readonly dataSource?: DataSource | undefined;
  readonly defaultExpanded?: boolean | undefined;
}

type FilterMode = 'all' | 'tools' | 'errors_anomalies';

const isToolFailureAnomalyRule = (ruleId: string): boolean => {
  return (
    ruleId === 'TOOL_RETRY_STREAK' ||
    ruleId === 'TOOL_HIGH_FAILURE_RATE' ||
    ruleId.startsWith('TOOL_') ||
    ruleId.includes('TOOL_FAILURE')
  );
};

export const matchesAnomaly = (event: EventFact, anomaly: Anomaly): boolean => {
  const ctx = anomaly.context;

  // Strict Session ID Guard: if both context and event specify session_id, they must match
  if (ctx.session_id && event.session_id && ctx.session_id !== event.session_id) {
    return false;
  }

  // Tool Failure Anomaly Status Criterion:
  // If anomaly is a tool failure rule, only match if the event is an error
  if (isToolFailureAnomalyRule(anomaly.rule_id)) {
    if (event.status !== 'error') {
      return false;
    }
  }

  // 1. Tool call match: tool_name matches
  if (
    ctx.tool_name &&
    event.tool_name &&
    ctx.tool_name.toLowerCase() === event.tool_name.toLowerCase()
  ) {
    if (ctx.turn !== undefined && event.turn !== undefined) {
      return ctx.turn === event.turn;
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
};

export const findMatchingAnomaly = (
  event: EventFact,
  anomalies?: readonly Anomaly[]
): Anomaly | undefined => {
  if (!anomalies || anomalies.length === 0) {
    return undefined;
  }

  return anomalies.find((a) => matchesAnomaly(event, a));
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
  isHydrated = false,
  dataSource = 'committed_ledger',
  defaultExpanded = false,
}) => {
  const [isWaterfallExpanded, setIsWaterfallExpanded] = useState<boolean>(defaultExpanded);
  const [filterMode, setFilterMode] = useState<FilterMode>('errors_anomalies');
  const [selectedRoles, setSelectedRoles] = useState<readonly string[]>(['all']);
  const [selectedTools, setSelectedTools] = useState<readonly string[]>(['all']);
  const [selectedAnomalies, setSelectedAnomalies] = useState<readonly string[]>(['all']);
  const [localSelectedEvent, setLocalSelectedEvent] = useState<EventFact | null>(null);
  const [expandedKeys, setExpandedKeys] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const activeSelectedEvent = selectedEvent !== undefined ? selectedEvent : localSelectedEvent;

  const handleCopyIdentifier = async (
    idText: string,
    key: string,
    e: React.MouseEvent
  ) => {
    e.stopPropagation();
    await copyTextToClipboard(idText);
    setCopiedId(key);
    window.setTimeout(() => {
      setCopiedId((current) => (current === key ? null : current));
    }, 2000);
  };

  const toggleExpand = (key: string, e?: React.SyntheticEvent) => {
    e?.stopPropagation();
    setExpandedKeys((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

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
    if (!events || events.length === 0) return [];
    const roleSet = new Set<string>();
    for (const e of events) {
      if (e.role && e.role.trim().length > 0) {
        roleSet.add(e.role.trim());
      }
    }
    return Array.from(roleSet).sort((a, b) => a.localeCompare(b));
  }, [events]);

  const roleOptions: readonly MultiSelectOption[] = useMemo(() => {
    return availableRoles.map((role) => ({
      value: role,
      label: role,
    }));
  }, [availableRoles]);

  // Unique tools for filter dropdown
  const availableTools = useMemo(() => {
    if (!events || events.length === 0) return [];
    const toolSet = new Set<string>();
    for (const e of events) {
      if (e.tool_name && e.tool_name.trim().length > 0) {
        toolSet.add(e.tool_name.trim());
      }
    }
    return Array.from(toolSet).sort((a, b) => a.localeCompare(b));
  }, [events]);

  const toolOptions: readonly MultiSelectOption[] = useMemo(() => {
    return availableTools.map((tool) => ({
      value: tool,
      label: tool,
    }));
  }, [availableTools]);

  // Unique anomaly rules for filter dropdown
  const activeAnomalies = useMemo(() => {
    if (!anomalies || anomalies.length === 0) return [];
    const ruleSet = new Set<string>();
    for (const a of anomalies) {
      if (a.rule_id && a.rule_id.trim().length > 0) {
        ruleSet.add(a.rule_id.trim());
      }
    }
    return Array.from(ruleSet).sort((a, b) => a.localeCompare(b));
  }, [anomalies]);

  const anomalyOptions: readonly MultiSelectOption[] = useMemo(() => {
    return activeAnomalies.map((rule) => ({
      value: rule,
      label: rule,
    }));
  }, [activeAnomalies]);

  const isRoleFilterActive = !selectedRoles.includes('all') && selectedRoles.length > 0;
  const isToolFilterActive = !selectedTools.includes('all') && selectedTools.length > 0;
  const isAnomalyFilterActive = !selectedAnomalies.includes('all') && selectedAnomalies.length > 0;

  // Filtered events
  const filteredEvents = useMemo(() => {
    return sortedEvents.filter((ev) => {
      // 1. FilterMode condition
      if (filterMode === 'tools' && ev.event_type !== 'tool_call') {
        return false;
      }
      if (filterMode === 'errors_anomalies') {
        const hasErr = ev.status === 'error';
        const hasAnomaly = Boolean(findMatchingAnomaly(ev, anomalies));
        if (!hasErr && !hasAnomaly) {
          return false;
        }
      }

      // 2. Multi-role filter
      if (isRoleFilterActive) {
        if (!ev.role || !selectedRoles.includes(ev.role.trim())) {
          return false;
        }
      }

      // 3. Multi-tool filter
      if (isToolFilterActive) {
        if (!ev.tool_name || !selectedTools.includes(ev.tool_name.trim())) {
          return false;
        }
      }

      // 4. Multi-anomaly filter
      if (isAnomalyFilterActive) {
        const hasSelectedAnomaly = anomalies?.some(
          (a) => selectedAnomalies.includes(a.rule_id) && matchesAnomaly(ev, a)
        );
        if (!hasSelectedAnomaly) {
          return false;
        }
      }

      return true;
    });
  }, [
    sortedEvents,
    filterMode,
    isRoleFilterActive,
    selectedRoles,
    anomalies,
    isToolFilterActive,
    selectedTools,
    isAnomalyFilterActive,
    selectedAnomalies,
  ]);

  if (!events || events.length === 0) {
    return (
      <div className="trace-waterfall-card">
        <div className="trace-empty-state">
          Zero chronological events recorded for this execution trace.
        </div>
      </div>
    );
  }

  // Summary statistics for collapsed bar and header
  const totalEventsCount = events.length;
  const totalErrorsCount = events.filter((ev) => ev.status === 'error').length;
  const totalAnomaliesCount = anomalies ? anomalies.length : 0;

  const issuesSummaryText =
    totalErrorsCount === 0 && totalAnomaliesCount === 0
      ? 'Clean Execution'
      : `${totalErrorsCount === 0 ? 'Zero' : totalErrorsCount} ${
          totalErrorsCount === 1 ? 'Error' : 'Errors'
        }, ${totalAnomaliesCount === 0 ? 'Zero' : totalAnomaliesCount} ${
          totalAnomaliesCount === 1 ? 'Anomaly' : 'Anomalies'
        }`;

  // Collapsed View
  if (!isWaterfallExpanded) {
    return (
      <div className="trace-waterfall-card trace-waterfall-collapsed" data-testid="trace-waterfall-collapsed">
        <div className="waterfall-summary-bar">
          <div className="waterfall-summary-left">
            <h3 className="waterfall-title">Execution Trace Waterfall</h3>
            <span
              className={`waterfall-source-badge ${isHydrated ? 'source-hydrated' : 'source-offline'}`}
              data-testid="waterfall-source-badge"
            >
              {isHydrated ? 'Source: Local SQLite' : 'Source: Committed Ledger'}
            </span>
          </div>

          <div className="waterfall-summary-metrics">
            <span className="waterfall-summary-stat mono" data-testid="waterfall-event-count">
              {totalEventsCount} Events Recorded
            </span>
            <span
              className={`waterfall-summary-issues ${
                totalErrorsCount > 0 || totalAnomaliesCount > 0 ? 'issues-flagged' : 'issues-clean'
              }`}
              data-testid="waterfall-issues-summary"
            >
              {issuesSummaryText}
            </span>
          </div>

          <div className="waterfall-summary-right">
            <button
              type="button"
              className="waterfall-toggle-btn"
              data-testid="waterfall-expand-btn"
              aria-expanded="false"
              onClick={() => setIsWaterfallExpanded(true)}
            >
              Expand Trace Waterfall
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Expanded Full View
  return (
    <div className="trace-waterfall-card" data-testid="trace-waterfall-card">
      <div className="waterfall-header">
        <div className="waterfall-title-group">
          <h3 className="waterfall-title">Execution Trace Waterfall</h3>
          <span
            className={`waterfall-source-badge ${isHydrated ? 'source-hydrated' : 'source-offline'}`}
            data-testid="waterfall-source-badge"
          >
            {isHydrated ? 'Source: Local SQLite' : 'Source: Committed Ledger'}
          </span>
          <span className="waterfall-counter">
            Showing {filteredEvents.length} of {events.length} events
          </span>
          <span className="waterfall-hint">
            Click any error or anomaly to inspect failure details
          </span>
          <button
            type="button"
            className="waterfall-toggle-btn"
            data-testid="waterfall-collapse-btn"
            aria-expanded="true"
            onClick={() => setIsWaterfallExpanded(false)}
          >
            Collapse Trace Waterfall
          </button>
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

          <MultiSelectDropdown
            id="trace-role-filter"
            testId="trace-role-filter"
            label="Role"
            allLabel="No filter"
            ariaLabel="Filter by subagent role"
            className="role-filter-group"
            options={roleOptions}
            selectedValues={selectedRoles}
            onChange={setSelectedRoles}
          />

          <MultiSelectDropdown
            id="trace-tool-filter"
            testId="trace-tool-filter"
            label="Tool"
            allLabel="No filter"
            ariaLabel="Filter by tool"
            className="tool-filter-group"
            options={toolOptions}
            selectedValues={selectedTools}
            onChange={setSelectedTools}
          />

          <MultiSelectDropdown
            id="trace-anomaly-filter"
            testId="trace-anomaly-filter"
            label="Anomaly"
            allLabel="No filter"
            ariaLabel="Filter by anomaly rule"
            className="anomaly-filter-group"
            options={anomalyOptions}
            selectedValues={selectedAnomalies}
            onChange={setSelectedAnomalies}
          />
        </div>
      </div>

      {/* Dynamic Hydration Fallback Banner */}
      {(!isHydrated || dataSource === 'committed_ledger') && (
        <div className="db-offline-banner" data-testid="db-offline-banner" role="status">
          <div className="offline-banner-header">
            <span className="offline-indicator-badge">Committed Git Ledger — Local Database Offline</span>
          </div>
          <p className="offline-banner-text">
            Local SQLite database was unavailable or not found. Showing lightweight execution events recorded in the repository ledger without raw prompt, reasoning, or tool payloads.
          </p>
        </div>
      )}

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
            const isExpanded = Boolean(expandedKeys[itemKey]) || isSelected;
            const itemClasses = `timeline-item ${isSelected ? 'selected' : ''} ${hasError ? 'item-error' : ''} ${matchingAnomaly ? 'item-anomaly' : ''} ${isIssue ? 'item-interactive' : ''}`
              .replace(/\s+/g, ' ')
              .trim();

            const turnDebugId = `turn-${ev.turn}`;
            const copyKey = `item-id-${itemKey}`;
            const isCopied = copiedId === copyKey;

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

                    {/* Concrete Debug Identifier Badge with Copy Button */}
                    <div className="timeline-id-badge" onClick={(e) => e.stopPropagation()}>
                      <span className="timeline-id-label mono" title={`Session: ${ev.session_id}${ev.message_id ? ` | Message: ${ev.message_id}` : ''}`}>
                        {ev.message_id ? `${ev.session_id.slice(0, 8)}...:${ev.message_id.slice(0, 8)}` : ev.session_id}
                      </span>
                      <button
                        type="button"
                        className="btn-copy-id"
                        aria-label={`Copy debug identifier for turn ${ev.turn}`}
                        onClick={(e) =>
                          handleCopyIdentifier(
                            ev.message_id ?? ev.session_id ?? turnDebugId,
                            copyKey,
                            e
                          )
                        }
                      >
                        {isCopied ? 'Copied' : 'Copy ID'}
                      </button>
                    </div>

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

                    <button
                      type="button"
                      className="btn-toggle-payload"
                      aria-label={`${isExpanded ? 'Hide' : 'View'} payload for turn ${ev.turn}`}
                      onClick={(e) => toggleExpand(itemKey, e)}
                    >
                      {isExpanded ? 'Hide Payload' : 'View Payload'}
                    </button>
                  </div>

                  {ev.error_message && (
                    <div
                      className="timeline-error-preview mono"
                      title="Click to open Tool Execution Failure window"
                    >
                      {ev.error_message}
                    </div>
                  )}

                  {isExpanded && (
                    <div className="payload-details-container" data-testid={`payload-details-${ev.turn}`}>
                      {isHydrated ? (
                        <>
                          {ev.message_id && (
                            <div className="payload-id-row">
                              <span className="drawer-label">Message ID:</span>
                              <span className="drawer-val mono">{ev.message_id}</span>
                              <button
                                type="button"
                                className="btn-copy-id"
                                aria-label="Copy Message ID"
                                onClick={(e) =>
                                  handleCopyIdentifier(ev.message_id!, `msg-${itemKey}`, e)
                                }
                              >
                                {copiedId === `msg-${itemKey}` ? 'Copied' : 'Copy'}
                              </button>
                            </div>
                          )}
                          {ev.prompt_content && (
                            <div className="payload-block prompt-block" data-testid="prompt-payload-block">
                              <div className="payload-block-title">Prompt Payload</div>
                              <pre className="mono payload-content">{ev.prompt_content}</pre>
                            </div>
                          )}
                          {ev.reasoning_content && (
                            <div className="payload-block reasoning-block" data-testid="reasoning-payload-block">
                              <div className="payload-block-title">Model Reasoning</div>
                              <pre className="mono payload-content">{ev.reasoning_content}</pre>
                            </div>
                          )}
                          {ev.tool_input && (
                            <div className="payload-block tool-input-block" data-testid="tool-input-payload-block">
                              <div className="payload-block-title">Tool Input</div>
                              <pre className="mono payload-content">{ev.tool_input}</pre>
                            </div>
                          )}
                          {ev.tool_output && (
                            <div className="payload-block tool-output-block" data-testid="tool-output-payload-block">
                              <div className="payload-block-title">Tool Output</div>
                              <pre className="mono payload-content">{ev.tool_output}</pre>
                            </div>
                          )}
                          {!ev.prompt_content && !ev.reasoning_content && !ev.tool_input && !ev.tool_output && (
                            <div className="payload-unhydrated-notice">Zero payload data captured for this event</div>
                          )}
                        </>
                      ) : (
                        <div className="payload-unhydrated-notice" data-testid="unhydrated-notice">
                          Full payload not captured in Git ledger
                        </div>
                      )}
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
