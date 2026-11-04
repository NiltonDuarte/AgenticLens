import React, { useState, useMemo, useId } from 'react';
import type { EventFact, Anomaly, DataSource } from '../types/telemetry';
import { formatRoleLabel, formatTokensK, formatLocalDateTime, formatHistoricalDateTimeTooltip, getHistoricalTimestampDateTime } from '../domain/formatters';
import { rebindEvent, sortEvents } from '../domain/events';
import { traceBadgeMetadata } from '../domain/traceBadges';
import { TraceInvocationBadge, type ChildSessionNavigation } from './TraceInvocationBadge';
import { AnomalyBadge } from './AnomalyBadge';
import { MultiSelectDropdown, type MultiSelectOption } from './MultiSelectDropdown';
import { copyTextToClipboard } from '../utils/clipboard';
import {
  generateToolExecutionPrompt,
  generateMessageTurnPrompt,
} from '../utils/debugPrompt';
import {
  isToolFailureAnomalyRule,
  matchesAnomaly,
  findMatchingAnomaly,
  formatAnomaly,
  type AnomalyRule,
  type ToolFailureAnomalyRule,
} from '../domain/anomalies';

// Re-export domain anomaly matching utilities for backwards compatibility
export {
  isToolFailureAnomalyRule,
  matchesAnomaly,
  findMatchingAnomaly,
  type AnomalyRule,
  type ToolFailureAnomalyRule,
};

export interface TraceWaterfallProps extends ChildSessionNavigation {
  readonly traceAvailability?: 'pending' | 'failed' | 'complete' | 'partial' | 'unavailable' | undefined;
  readonly traceDiagnostics?: readonly string[] | undefined;
  readonly events: readonly EventFact[];
  readonly anomalies?: readonly Anomaly[] | undefined;
  readonly selectedEvent?: EventFact | null | undefined;
  readonly onSelectEvent?: ((event: EventFact, anomaly?: Anomaly | undefined) => void) | undefined;
  readonly onSelectAnomaly?: ((anomaly: Anomaly) => void) | undefined;
  readonly isHydrated?: boolean | undefined;
  readonly dataSource?: DataSource | undefined;
  readonly defaultExpanded?: boolean | undefined;
}

type FilterMode = 'all' | 'tools' | 'errors_anomalies';

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
    case 'skill_activation':
      return 'Skill Activation';
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
  if (a === b) return true;
  return Boolean(a.session_id && a.event_id && b.event_id &&
    a.session_id === b.session_id && a.event_id === b.event_id);
};

export const TraceWaterfall: React.FC<TraceWaterfallProps> = ({
  events,
  traceAvailability = 'complete',
  traceDiagnostics = [],
  availableSessionIds,
  onSelectSession,
  anomalies,
  selectedEvent,
  onSelectEvent,
  onSelectAnomaly,
  isHydrated = false,
  dataSource = 'unavailable',
  defaultExpanded = false,
}) => {
  const [isWaterfallExpanded, setIsWaterfallExpanded] = useState<boolean>(defaultExpanded);
  const [filterMode, setFilterMode] = useState<FilterMode>('errors_anomalies');
  const [selectedRoles, setSelectedRoles] = useState<readonly string[]>(['all']);
  const [selectedTools, setSelectedTools] = useState<readonly string[]>(['all']);
  const [selectedAnomalies, setSelectedAnomalies] = useState<readonly string[]>(['all']);
  const [localSelection, setLocalSelection] = useState<{
    readonly event: EventFact;
    readonly events: readonly EventFact[];
  } | null>(null);
  const [expandedKeys, setExpandedKeys] = useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const waterfallBodyId = useId();
  const badges = useMemo(() => traceBadgeMetadata(events), [events]);

  const activeSelectedEvent = selectedEvent !== undefined
    ? selectedEvent
    : localSelection ? rebindEvent(localSelection.event, localSelection.events, events) : null;
  const hasLocalSource = isHydrated && ['local_sqlite', 'local_codex', 'mixed_sources'].includes(dataSource);
  const sourceLabel = (hasLocalSource
    ? (dataSource === 'mixed_sources' ? 'Source: Mixed native sources' : dataSource === 'local_codex' ? 'Source: Local Codex' : 'Source: Local SQLite')
    : 'Source unavailable') + (traceAvailability === 'partial' ? ' · Partial trace' : '');

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

  const handleCopyTimelinePrompt = async (
    ev: EventFact,
    key: string,
    e: React.MouseEvent
  ) => {
    e.stopPropagation();
    let prompt: string;
    if (ev.event_type === 'tool_call') {
      const matched = findMatchingAnomaly(ev, anomalies);
      const streakCount = matched?.context?.count;
      prompt = generateToolExecutionPrompt({
        toolName: ev.tool_name || 'unknown_tool',
        turn: ev.turn,
        sessionId: ev.session_id,
        messageId: ev.message_id ?? undefined,
        status: ev.status,
        consecutiveErrors: typeof streakCount === 'number' && streakCount > 1 ? streakCount : undefined,
        input: ev.tool_input ?? undefined,
        output: ev.tool_output ?? undefined,
        errorMessage: ev.error_message ?? undefined,
      });
    } else {
      prompt = generateMessageTurnPrompt({
        turn: ev.turn,
        role: ev.role,
        sessionId: ev.session_id,
        messageId: ev.message_id ?? undefined,
        reasoningContent: ev.reasoning_content ?? undefined,
        messageContent: ev.prompt_content ?? undefined,
        toolsExecuted: [],
      });
    }
    await copyTextToClipboard(prompt);
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
    setLocalSelection({ event: ev, events });
    const isFlagged = ev.status === 'error' || ev.status === 'failed' || Boolean(matchingAnomaly);
    onSelectEvent?.(ev, isFlagged ? matchingAnomaly : undefined);
  };

  // Sorted events chronologically
  const sortedEvents = useMemo(() => sortEvents(events ?? []), [events]);

  // Unique roles for filter dropdown
  const availableRoles = useMemo(() => {
    if (!events || events.length === 0) return [];
    const roleSet = new Set<string>();
    for (const e of events) {
      roleSet.add(e.role.trim());
    }
    return Array.from(roleSet).sort((a, b) => a.localeCompare(b));
  }, [events]);

  const roleOptions: readonly MultiSelectOption[] = useMemo(() => {
    return availableRoles.map((role) => ({
      value: role,
      label: formatRoleLabel(role),
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
        const hasErr = ev.status === 'error' || ev.status === 'failed';
        const hasAnomaly = Boolean(findMatchingAnomaly(ev, anomalies));
        if (!hasErr && !hasAnomaly) {
          return false;
        }
      }

      // 2. Multi-role filter
      if (isRoleFilterActive) {
        if (!selectedRoles.includes(ev.role.trim())) {
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

  const availabilityNotice = traceAvailability === 'complete' ? null : <div role="status" className="db-offline-banner" data-testid="db-offline-banner">
    <p>{traceAvailability === 'pending' ? 'Loading source trace' : traceAvailability === 'failed' ? 'Source trace request failed' : traceAvailability === 'partial' ? 'Partial source trace' : 'Source trace unavailable'}</p>
    <p>Stored skills, delegations, metrics and anomaly summaries remain available. Trace details require the original session sources.</p>
    {traceDiagnostics.map((diagnostic, index) => <p key={index}>{diagnostic}</p>)}
  </div>;

  if ((!events || events.length === 0) && (!anomalies || anomalies.length === 0)) {
    return (
      <div className="trace-waterfall-card" data-testid="trace-waterfall-card">
        <div className="waterfall-header">
          <div className="waterfall-title-group">
            <h3 className="waterfall-title">Execution Trace Waterfall</h3>
            <span
              className={`waterfall-source-badge ${hasLocalSource ? 'source-hydrated' : 'source-offline'}`}
              data-testid="waterfall-source-badge"
            >
              {sourceLabel}
            </span>
            <span className="waterfall-summary-stat mono" data-testid="waterfall-event-count">
              0 Events Recorded
            </span>
          </div>
        </div>
        {availabilityNotice}
        <div className="trace-empty-state">
          {traceAvailability === 'complete' ? 'Zero chronological events recorded for this execution trace.' : 'No source trace events available.'}
        </div>
      </div>
    );
  }

  // Summary statistics for collapsed bar and header
  const totalEventsCount = events.length;
  const totalErrorsCount = events.filter((ev) => ev.status === 'error' || ev.status === 'failed').length;
  const totalAnomaliesCount = anomalies ? anomalies.length : 0;

  const issuesSummaryText =
    totalErrorsCount === 0 && totalAnomaliesCount === 0
      ? (traceAvailability === 'complete' ? 'Clean Execution' : 'Trace coverage incomplete')
      : `${totalErrorsCount === 0 && traceAvailability !== 'complete' ? 'Errors unavailable' : `${totalErrorsCount === 0 ? 'Zero' : totalErrorsCount} ${totalErrorsCount === 1 ? 'Error' : 'Errors'}`}, ${totalAnomaliesCount === 0 ? 'Zero' : totalAnomaliesCount} ${
          totalAnomaliesCount === 1 ? 'Anomaly' : 'Anomalies'
        }`;

  const renderCard = (body?: React.ReactNode) => (
    <div
      className={`trace-waterfall-card${isWaterfallExpanded ? '' : ' trace-waterfall-collapsed'}`}
      data-testid={isWaterfallExpanded ? 'trace-waterfall-card' : 'trace-waterfall-collapsed'}
    >
      <div className="waterfall-summary-bar">
        <div className="waterfall-summary-left">
          <h3 className="waterfall-title">Execution Trace Waterfall</h3>
          <span
            className={`waterfall-source-badge ${hasLocalSource ? 'source-hydrated' : 'source-offline'}`}
            data-testid="waterfall-source-badge"
          >
            {sourceLabel}
          </span>
        </div>

        {(!isWaterfallExpanded || totalEventsCount === 0) && (
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
        )}

        <div className="waterfall-summary-right">
          <button
            type="button"
            className="waterfall-toggle-btn"
            data-testid={isWaterfallExpanded ? 'waterfall-collapse-btn' : 'waterfall-expand-btn'}
            aria-expanded={isWaterfallExpanded}
            aria-controls={waterfallBodyId}
            onClick={() => setIsWaterfallExpanded((expanded) => !expanded)}
          >
            {isWaterfallExpanded ? 'Collapse' : 'Expand'} Trace Waterfall
          </button>
        </div>
      </div>

      {availabilityNotice}
      <div id={waterfallBodyId} className="waterfall-body" hidden={!isWaterfallExpanded}>
        {body}
      </div>
    </div>
  );

  if (!isWaterfallExpanded) return renderCard();

  if (totalEventsCount === 0) {
    return renderCard(
      <>
        <div className="trace-empty-state">
          {traceAvailability === 'complete' ? 'Zero chronological events recorded for this execution trace.' : 'No source trace events available.'}
        </div>
        <div className="legacy-anomalies-section" data-testid="workspace-anomalies">
          <span className="legacy-anomalies-heading">Detected Workspace Anomalies:</span>
          <div className="legacy-anomalies-list">
            {anomalies?.map((anomaly, index) => {
              const display = formatAnomaly(anomaly);
              return (
                <button
                  key={`${anomaly.rule_id}-${index}`}
                  type="button"
                  className="legacy-anomaly-item"
                  onClick={() => onSelectAnomaly?.(anomaly)}
                  aria-label={`Inspect anomaly: ${display.title}`}
                >
                  <span className="legacy-anomaly-info">
                    <span className="legacy-anomaly-title">{display.title}</span>
                    <span className="legacy-anomaly-desc">{display.description}</span>
                  </span>
                  <AnomalyBadge severity={anomaly.severity} label={anomaly.severity} size="sm" />
                </button>
              );
            })}
          </div>
        </div>
      </>
    );
  }

  return renderCard(
    <>
      <div className="waterfall-header">
        <div className="waterfall-title-group">
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
            const hasError = ev.status === 'error' || ev.status === 'failed';
            const isIssue = hasError || Boolean(matchingAnomaly);
            const roleClass = `role-pill ${ev.role.toLowerCase()}`;
            const itemKey = `${ev.timestamp}-${ev.turn}-${ev.session_id}-${idx}`;
            const isExpanded = Boolean(expandedKeys[itemKey]) || isSelected;
            const itemClasses = `timeline-item ${isSelected ? 'selected' : ''} ${hasError ? 'item-error' : ''} ${matchingAnomaly ? 'item-anomaly' : ''} ${isIssue ? 'item-interactive' : ''}`
              .replace(/\s+/g, ' ')
              .trim();

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
                aria-label={`Event Turn ${ev.turn} ${formatRoleLabel(ev.role)} ${ev.event_type}`}
              >
                <div className="timeline-rail">
                  <div className={`timeline-node ${hasError ? 'node-error' : ''}`} />
                  {idx < filteredEvents.length - 1 && <div className="timeline-connector" />}
                </div>

                <div className="timeline-content">
                  <div className="timeline-top-row">
                    <span className="turn-pill">Turn {ev.turn}</span>
                    <span className={roleClass}>{formatRoleLabel(ev.role)}</span>
                    <span className="event-type-badge">
                      {formatEventTypeLabel(ev.event_type)}
                    </span>
                    <time className="timeline-timestamp mono" dateTime={getHistoricalTimestampDateTime(ev.timestamp)} title={formatHistoricalDateTimeTooltip(ev.timestamp)}>{formatLocalDateTime(ev.timestamp)}</time>

                    {/* Concrete Debug Identifier Badge with Copy Button */}
                    <div className="timeline-id-badge" onClick={(e) => e.stopPropagation()}>
                      <span className="timeline-id-label mono" title={`Session: ${ev.session_id}${ev.message_id ? ` | Message: ${ev.message_id}` : ''}`}>
                        {ev.message_id ? `${ev.session_id.slice(0, 8)}...:${ev.message_id.slice(0, 8)}` : ev.session_id}
                      </span>
                      <button
                        type="button"
                        className="btn-copy-id"
                        aria-label={`Copy debug prompt for turn ${ev.turn}`}
                        onClick={(e) => handleCopyTimelinePrompt(ev, copyKey, e)}
                      >
                        {isCopied ? 'Copied' : 'Copy Prompt'}
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
                          setLocalSelection({ event: ev, events });
                          onSelectEvent?.(ev, matchingAnomaly);
                        }}
                      />
                    )}
                  </div>

                  <div className="timeline-details-row">
                    {ev.event_type === 'skill_activation' && <span className="status-pill">Activation</span>}
                    {ev.event_type !== 'tool_call' && <TraceInvocationBadge metadata={badges.get(ev)} availableSessionIds={availableSessionIds} onSelectSession={onSelectSession} />}
                    {ev.event_type === 'tool_call' && (
                      <div className="tool-details-group">
                        <span className="tool-label">Tool:</span>
                        <span className="tool-name mono">{ev.tool_name ?? 'unknown'}</span>
                        <TraceInvocationBadge metadata={badges.get(ev)} availableSessionIds={availableSessionIds} onSelectSession={onSelectSession} />
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
                        <div className="payload-unhydrated-notice" data-testid="unhydrated-notice">
                          Payload details unavailable for this event.
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
    </>
  );
};
