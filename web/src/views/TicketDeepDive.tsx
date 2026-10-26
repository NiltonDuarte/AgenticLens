import React, { useEffect, useState, useCallback, useRef } from 'react';
import type {
  AgenticMetricsLedger,
  Anomaly,
  DataSource,
  EventFact,
  WorkspaceDetail,
} from '../domain/types';
import { sessionEvents, sessionTools } from '../domain/types';
import {
  fetchWorkspaceById,
  fetchWorkspaceEvents,
  fetchWorkspaceLedgers,
} from '../api/client';
import {
  formatCurrencyUSD,
  formatDurationSeconds,
  formatIsoDateTimeShort,
  formatLeadTime,
  formatPercentage,
  formatTokensK,
  computeLeadTimeSeconds,
} from '../domain/formatters';
import { aggregateToolUsage } from '../domain/tools';
import { KPICard } from '../components/KPICard';
import { EmptyState } from '../components/EmptyState';
import { TicketSelector } from '../components/TicketSelector';
import { SaturationBar } from '../components/SaturationBar';
import { ThreadViewer } from '../components/ThreadViewer';
import { ArtifactViewer } from '../components/ArtifactViewer';
import { getRoleColor, getModelColor } from '../utils/theme';
import { TraceWaterfall } from '../components/TraceWaterfall';
import { AnomalyDrawer } from '../components/AnomalyDrawer';
import { formatAnomaly } from '../domain/anomalies';
import { AnomalyBadge } from '../components/AnomalyBadge';
import { InfoTooltip } from '../components/InfoTooltip';

export interface TicketDeepDiveProps {
  readonly selectedTicketId?: string | null;
  readonly onSelectTicket?: (ticketId: string) => void;
}

export const TicketDeepDive: React.FC<TicketDeepDiveProps> = ({
  selectedTicketId,
  onSelectTicket,
}) => {
  const [ledgers, setLedgers] = useState<readonly AgenticMetricsLedger[]>([]);
  const [detail, setDetail] = useState<WorkspaceDetail | null>(null);
  const [events, setEvents] = useState<readonly EventFact[]>([]);
  const [isHydrated, setIsHydrated] = useState(false);
  const [dataSource, setDataSource] = useState<DataSource>('committed_ledger');
  const [traceAnomalies, setTraceAnomalies] = useState<readonly Anomaly[]>([]);
  const [selectedEvent, setSelectedEvent] = useState<EventFact | null>(null);
  const [selectedAnomaly, setSelectedAnomaly] = useState<Anomaly | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [activeTicketId, setActiveTicketId] = useState<string | null>(selectedTicketId ?? null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const ledgersRef = useRef<readonly AgenticMetricsLedger[]>([]);

  // Sync activeTicketId when parent prop changes
  useEffect(() => {
    if (selectedTicketId && selectedTicketId !== activeTicketId) {
      setActiveTicketId(selectedTicketId);
    }
  }, [selectedTicketId, activeTicketId]);

  const loadData = useCallback(async (forceReloadFleet = false) => {
    setLoading(true);
    setError(null);
    try {
      let currentLedgers = ledgersRef.current;

      // Only fetch fleet catalog if in-memory cache is empty or explicitly forced
      if (forceReloadFleet || currentLedgers.length === 0) {
        const ledgersRes = await fetchWorkspaceLedgers('all');
        currentLedgers = ledgersRes.data;
        ledgersRef.current = currentLedgers;
        setLedgers(currentLedgers);
      }

      let targetId = activeTicketId;
      if (!targetId && currentLedgers.length > 0 && currentLedgers[0]?.ticket_id) {
        targetId = currentLedgers[0].ticket_id;
        setActiveTicketId(targetId);
      }

      if (targetId) {
        const detailRes = await fetchWorkspaceById(targetId);
        setDetail(detailRes.data);

        let loadedEvents: readonly EventFact[] = sessionEvents(detailRes.data.ledger?.facts?.sessions);
        let loadedAnomalies: readonly Anomaly[] = detailRes.data.anomalies ?? [];
        let hydrated = false;
        let source: DataSource = 'committed_ledger';
        try {
          const eventsRes = await fetchWorkspaceEvents(targetId);
          if (eventsRes.data && eventsRes.data.events && eventsRes.data.events.length > 0) {
            loadedEvents = eventsRes.data.events;
          }
          if (eventsRes.data?.anomalies) loadedAnomalies = eventsRes.data.anomalies;
          hydrated = eventsRes.data?.is_hydrated ?? false;
          source = eventsRes.data?.data_source ?? 'committed_ledger';
        } catch {
          // Gracefully keep events from detail.ledger.facts.events if fetch fails
        }

        setEvents(loadedEvents);
        setTraceAnomalies(loadedAnomalies);
        setIsHydrated(hydrated);
        setDataSource(source);
      } else {
        setDetail(null);
        setEvents([]);
        setTraceAnomalies([]);
        setIsHydrated(false);
        setDataSource('committed_ledger');
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load ticket details';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [activeTicketId]);

  useEffect(() => {
    loadData(false);
  }, [loadData]);

  const handleRefresh = useCallback(() => {
    loadData(true);
  }, [loadData]);

  const handleSelectTicket = (ticketId: string) => {
    setActiveTicketId(ticketId);
    onSelectTicket?.(ticketId);
  };

  const handleSelectEvent = (event: EventFact, anomaly?: Anomaly) => {
    setSelectedEvent(event);
    const isFlagged = Boolean(anomaly) || event.status === 'error';
    if (isFlagged) {
      setSelectedAnomaly(anomaly ?? null);
      setIsDrawerOpen(true);
    } else {
      setSelectedAnomaly(null);
      setIsDrawerOpen(false);
    }
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner" />
        <span className="loading-text">Loading workspace autopsy telemetry...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-banner">
        <span className="error-text">{error}</span>
        <button type="button" className="btn-secondary" onClick={handleRefresh}>
          Retry Connection
        </button>
      </div>
    );
  }

  if (!detail || !detail.ledger) {
    return (
      <EmptyState
        title="Zero Telemetry Workspaces Available"
        message="No workspace autopsy records found. Ensure telemetry ledgers are present in the workspaces directory."
        actionLabel="Refresh Data"
        onAction={handleRefresh}
      />
    );
  }

  const { ledger, artifacts } = detail;
  const eff = ledger.workflow_efficiency;
  const totals = ledger.totals;
  const leadSecs = computeLeadTimeSeconds(ledger.started_at, ledger.completed_at);
  const code = eff.code_delivered;
  const tools = eff.tool_execution;
  const toolUsageByName = aggregateToolUsage(sessionTools(ledger.facts?.sessions));
  const subTurns = eff.subagent_turn_velocity;

  const ticketOptions = ledgers.map((l) => ({
    ticket_id: l.ticket_id ?? 'unknown',
    ticket_title: l.ticket_title,
    context_health_status: l.workflow_efficiency?.context_window_health?.status,
    completed_at: l.completed_at,
    updated_at: l.updated_at,
    started_at: l.started_at,
  }));

  const turnEntries = Object.entries(subTurns.by_role ?? {}).sort((a, b) => b[1] - a[1]);

  return (
    <div className="deep-dive-container">
      {/* Ticket Selection Bar */}
      <div className="deep-dive-top-bar">
        <TicketSelector
          tickets={ticketOptions}
          selectedTicketId={activeTicketId}
          onSelectTicket={handleSelectTicket}
        />
      </div>

      {/* Workspace Header Information */}
      <div className="workspace-header-card">
        <div className="header-meta-row">
          <div className="header-left">
            <span className="ticket-id-tag">{ledger.ticket_id ?? 'UNKNOWN'}</span>
            <span className={`ticket-type-pill ${ledger.ticket_type ? '' : 'unclassified'}`.trim()}>{ledger.ticket_type ?? 'unclassified'}</span>
            <h2 className="workspace-title">{ledger.ticket_title ?? 'Untitled Workspace'}</h2>
          </div>
          <div className="header-right">
            <div className="timestamp-group">
              <span className="timestamp-item">
                <span className="timestamp-label">Started:</span>{' '}
                {formatIsoDateTimeShort(ledger.started_at)}
              </span>
              <span className="timestamp-item">
                <span className="timestamp-label">Completed:</span>{' '}
                {ledger.completed_at ? formatIsoDateTimeShort(ledger.completed_at) : 'In Progress'}
              </span>
              <span className="timestamp-item">
                <span className="timestamp-label">Updated:</span>{' '}
                {formatIsoDateTimeShort(ledger.updated_at)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Row 1: Primary Workflow KPIs */}
      <div className="kpi-grid">
        <KPICard
          title="SDLC Iterations"
          value={eff.sdlc_iterations.toString()}
          secondary="Feedback iterations"
          tooltip="Total SDLC feedback iterations across all participating agent roles"
        />
        <KPICard
          title="Turnaround Lead Time"
          value={formatLeadTime(leadSecs)}
          secondary={leadSecs ? `${leadSecs.toLocaleString()}s elapsed` : 'Active'}
          tooltip="Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel."
        />
        <KPICard
          title="Active Compute Duration"
          value={eff.active_duration_formatted || formatDurationSeconds(eff.active_duration_seconds)}
          secondary={`${eff.active_duration_seconds.toLocaleString()}s total`}
          tooltip="Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution."
        />
        <KPICard
          title="Total Cost"
          value={formatCurrencyUSD(totals.cost_usd)}
          secondary={`${totals.invocations} invocations`}
          tooltip="Total USD spend across all model calls"
        />
        <KPICard
          title="Cache Hit Rate"
          value={formatPercentage(totals.cache_hit_rate_pct)}
          secondary={`${formatTokensK(totals.cache_read_tokens)} read`}
          tooltip="Prompt cache reuse efficiency"
        />
        <KPICard
          title="Test:Prod Code Ratio"
          value={eff.test_to_production_ratio.ratio}
          secondary={`${eff.test_to_production_ratio.test_lines_added} test lines`}
          tooltip="Ratio of test lines added to production lines added"
        />
      </div>

      {/* Row 2: Secondary Operational Profile Cards */}
      <div className="profile-grid">
        <div className="profile-card">
          <h4 className="card-section-title">Code Delivery</h4>
          <div className="stat-row">
            <span className="stat-label">Lines Added:</span>
            <span className="stat-val mono positive">+{code.lines_added.toLocaleString()}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Lines Deleted:</span>
            <span className="stat-val mono negative">-{code.lines_deleted.toLocaleString()}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Files Modified:</span>
            <span className="stat-val mono">{code.files_changed}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Net Code Growth:</span>
            <span className="stat-val mono">{(code.lines_added - code.lines_deleted).toLocaleString()} lines</span>
          </div>
        </div>

        <div className="profile-card">
          <h4 className="card-section-title">Reasoning Tokens</h4>
          <div className="stat-row">
            <span className="stat-label">Reasoning Volume:</span>
            <span className="stat-val mono">{formatTokensK(eff.reasoning_tokens)} tokens</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Reasoning Ratio:</span>
            <span className="stat-val mono">{formatPercentage(eff.reasoning_ratio_pct)}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Output Tokens:</span>
            <span className="stat-val mono">{formatTokensK(totals.output_tokens)} tokens</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Input Tokens:</span>
            <span className="stat-val mono">{formatTokensK(totals.input_tokens)} tokens</span>
          </div>
        </div>

        <div className="profile-card">
          <h4 className="card-section-title">Tool Reliability</h4>
          <div className="stat-row">
            <span className="stat-label">Total Calls:</span>
            <span className="stat-val mono">{tools.total_calls.toLocaleString()}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Errors / Corrections:</span>
            <span className="stat-val mono negative">{tools.errors_and_corrections.toLocaleString()}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Success Rate:</span>
            <span className="stat-val mono positive">{formatPercentage(tools.success_rate_pct)}</span>
          </div>
          <div className="stat-row">
            <span className="stat-label">Tool Tools Count:</span>
            <span className="stat-val mono">{toolUsageByName.length}</span>
          </div>
        </div>
      </div>

      {/* Row 3: Context Window Health & Saturation */}
      <SaturationBar contextHealth={eff.context_window_health} />

      {/* Row 4: Subagent Turn Velocity */}
      <div className="turn-velocity-card">
        <div className="turn-velocity-header">
          <div className="turn-velocity-title-group">
            <h4 className="section-title">Subagent Turn Velocity</h4>
            <InfoTooltip
              content="Average and peak conversational turns executed per agent invocation, categorized by subagent persona role."
              ariaLabel="Information about Subagent Turn Velocity"
            />
          </div>
          <div className="turn-velocity-badges">
            <span className="velocity-badge">Avg Turns: {subTurns.avg_turns_per_invocation.toFixed(1)}</span>
            <span className="velocity-badge">Peak Turns: {subTurns.peak_turns}</span>
          </div>
        </div>
        {turnEntries.length > 0 ? (
          <div className="turns-by-role-list">
            {turnEntries.map(([role, turns]) => {
              const roleColor = getRoleColor(role);
              return (
                <div key={role} className="turn-role-row">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '6px', minWidth: '100px' }}>
                    <span
                      className="turn-role-dot"
                      style={{
                        width: '8px',
                        height: '8px',
                        borderRadius: '50%',
                        backgroundColor: roleColor,
                        display: 'inline-block',
                        flexShrink: 0,
                      }}
                      aria-hidden="true"
                    />
                    <span className="turn-role-label">{role}</span>
                  </div>
                  <div className="turn-track">
                    <div
                      className="turn-fill"
                      style={{
                        width: `${Math.min(100, (turns / Math.max(1, subTurns.peak_turns)) * 100)}%`,
                        backgroundColor: roleColor,
                      }}
                    />
                  </div>
                  <span className="turn-role-val">{turns.toFixed(1)} turns/inv</span>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="empty-turns-note">No subagent turn records available.</div>
        )}
      </div>

      {/* Row 5: Role and Model Metrics Tables */}
      <div className="tables-grid">
        <div className="table-card">
          <div className="table-card-header">
            <h4 className="table-card-title">Role Resource Consumption</h4>
            <InfoTooltip
              content="Telemetry breakdown of model invocations, token volume, cache hit rate, and spend by agent persona role."
              placement="left"
              ariaLabel="Information about Role Resource Consumption"
            />
          </div>
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th>
                    <span className="th-content">
                      <span>Role</span>
                      <InfoTooltip
                        content="Agent persona role (orchestrator, coder, tester, reviewer, planner)"
                        placement="bottom"
                        ariaLabel="Information about Role"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Invocations</span>
                      <InfoTooltip
                        content="Total count of model invocations initiated by this role"
                        placement="bottom"
                        ariaLabel="Information about Invocations"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Tokens</span>
                      <InfoTooltip
                        content="Total input and output tokens consumed by this role"
                        placement="bottom"
                        ariaLabel="Information about Tokens"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Cache Hit</span>
                      <InfoTooltip
                        content="Percentage of prompt tokens served from prompt cache for this role"
                        placement="bottom"
                        ariaLabel="Information about Cache Hit"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Cost</span>
                      <InfoTooltip
                        content="Total model inference spend in US dollars for this role"
                        placement="bottom"
                        ariaLabel="Information about Cost"
                      />
                    </span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {ledger.roles.map((r) => {
                  const roleColor = getRoleColor(r.role);
                  return (
                    <tr key={r.role}>
                      <td className="role-col">
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <span
                            className="role-color-dot"
                            style={{
                              width: '8px',
                              height: '8px',
                              borderRadius: '50%',
                              backgroundColor: roleColor,
                              display: 'inline-block',
                              flexShrink: 0,
                            }}
                            aria-hidden="true"
                          />
                          <span>{r.role}</span>
                        </div>
                      </td>
                      <td className="num-col">{r.invocations}</td>
                      <td className="num-col">{formatTokensK(r.input_tokens + r.output_tokens)}</td>
                      <td className="num-col">{formatPercentage(r.cache_hit_rate_pct)}</td>
                      <td className="num-col">{formatCurrencyUSD(r.cost_usd)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>

        <div className="table-card">
          <div className="table-card-header">
            <h4 className="table-card-title">Model Resource Consumption</h4>
            <InfoTooltip
              content="Telemetry breakdown of invocations, token volume, cache hit rate, and spend by model family and version."
              placement="left"
              ariaLabel="Information about Model Resource Consumption"
            />
          </div>
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th>
                    <span className="th-content">
                      <span>Model</span>
                      <InfoTooltip
                        content="Model identifier and version invoked"
                        placement="bottom"
                        ariaLabel="Information about Model"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Invocations</span>
                      <InfoTooltip
                        content="Total count of invocations executed on this model"
                        placement="bottom"
                        ariaLabel="Information about Invocations"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Tokens</span>
                      <InfoTooltip
                        content="Total input and output tokens consumed by this model"
                        placement="bottom"
                        ariaLabel="Information about Tokens"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Cache Hit</span>
                      <InfoTooltip
                        content="Percentage of prompt tokens served from prompt cache for this model"
                        placement="bottom"
                        ariaLabel="Information about Cache Hit"
                      />
                    </span>
                  </th>
                  <th className="num-col">
                    <span className="th-content">
                      <span>Cost</span>
                      <InfoTooltip
                        content="Total model inference spend in US dollars for this model"
                        placement="bottom"
                        ariaLabel="Information about Cost"
                      />
                    </span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {ledger.models.map((m) => {
                  const modelColor = getModelColor(m.model);
                  return (
                    <tr key={m.model}>
                      <td className="model-col">
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <span
                            className="model-color-dot"
                            style={{
                              width: '8px',
                              height: '8px',
                              borderRadius: '50%',
                              backgroundColor: modelColor,
                              display: 'inline-block',
                              flexShrink: 0,
                            }}
                            aria-hidden="true"
                          />
                          <span>{m.model}</span>
                        </div>
                      </td>
                      <td className="num-col">{m.invocations}</td>
                      <td className="num-col">{formatTokensK(m.input_tokens + m.output_tokens)}</td>
                      <td className="num-col">{formatPercentage(m.cache_hit_rate_pct)}</td>
                      <td className="num-col">{formatCurrencyUSD(m.cost_usd)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Row 6: Visual Trace Waterfall or Legacy v1 Fallback */}
      {events && events.length > 0 ? (
        <TraceWaterfall
          events={events}
          isHydrated={isHydrated}
          dataSource={dataSource}
          anomalies={traceAnomalies}
          selectedEvent={selectedEvent}
          onSelectEvent={handleSelectEvent}
        />
      ) : (
        <div className="legacy-fallback-banner" data-testid="legacy-fallback-banner">
          <h4 className="legacy-banner-title">Legacy Telemetry Schema Notice</h4>
          <p className="legacy-banner-text">
            Chronological trace waterfall is unavailable for legacy v1 telemetry ledgers. Aggregate workspace anomalies and thread records are displayed below.
          </p>
          {detail.anomalies && detail.anomalies.length > 0 && (
            <div className="legacy-anomalies-section">
              <span className="legacy-anomalies-heading">Detected Workspace Anomalies:</span>
              <div className="legacy-anomalies-list">
                {detail.anomalies.map((a, i) => (
                  (() => {
                    const display = formatAnomaly(a);
                    return (
                  <div
                    key={`${a.rule_id}-${i}`}
                    className="legacy-anomaly-item"
                    onClick={() => {
                      setSelectedAnomaly(a);
                      setSelectedEvent(null);
                      setIsDrawerOpen(true);
                    }}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setSelectedAnomaly(a);
                        setSelectedEvent(null);
                        setIsDrawerOpen(true);
                      }
                    }}
                    aria-label={`Inspect anomaly: ${display.title}`}
                  >
                    <div className="legacy-anomaly-info">
                      <span className="legacy-anomaly-title">{display.title}</span>
                      <span className="legacy-anomaly-desc">{display.description}</span>
                    </div>
                    <AnomalyBadge severity={a.severity} label={a.severity} size="sm" />
                  </div>
                    );
                  })()
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Row 7: Session Thread Logs */}
      <ThreadViewer threads={ledger.threads} />

      {/* Row 8: Workspace Artifacts */}
      <ArtifactViewer artifacts={artifacts} />

      {/* Anomaly Inspection Drawer */}
      <AnomalyDrawer
        isOpen={isDrawerOpen}
        onClose={() => setIsDrawerOpen(false)}
        anomaly={selectedAnomaly}
        event={selectedEvent}
      />
    </div>
  );
};
