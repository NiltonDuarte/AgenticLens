import React, { useEffect, useState, useCallback, useRef } from 'react';
import type { AgenticMetricsLedger, WorkspaceDetail } from '../domain/types';
import { fetchWorkspaceById, fetchWorkspaceLedgers } from '../api/client';
import {
  formatCurrencyUSD,
  formatDurationSeconds,
  formatIsoDateTimeShort,
  formatLeadTime,
  formatPercentage,
  formatTokensK,
  computeLeadTimeSeconds,
} from '../domain/formatters';
import { KPICard } from '../components/KPICard';
import { EmptyState } from '../components/EmptyState';
import { TicketSelector } from '../components/TicketSelector';
import { SaturationBar } from '../components/SaturationBar';
import { ThreadViewer } from '../components/ThreadViewer';
import { ArtifactViewer } from '../components/ArtifactViewer';

export interface TicketDeepDiveProps {
  readonly selectedTicketId?: string | null;
  readonly onSelectTicket?: (ticketId: string) => void;
  readonly onMockChange?: (isMock: boolean) => void;
}

export const TicketDeepDive: React.FC<TicketDeepDiveProps> = ({
  selectedTicketId,
  onSelectTicket,
  onMockChange,
}) => {
  const [ledgers, setLedgers] = useState<readonly AgenticMetricsLedger[]>([]);
  const [detail, setDetail] = useState<WorkspaceDetail | null>(null);
  const [activeTicketId, setActiveTicketId] = useState<string | null>(selectedTicketId ?? null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const ledgersRef = useRef<readonly AgenticMetricsLedger[]>([]);
  const isMockFleetRef = useRef(false);

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
      let isMockFleet = isMockFleetRef.current;

      // Only fetch fleet catalog if in-memory cache is empty or explicitly forced
      if (forceReloadFleet || currentLedgers.length === 0) {
        const ledgersRes = await fetchWorkspaceLedgers('all');
        currentLedgers = ledgersRes.data;
        isMockFleet = ledgersRes.isMock;
        ledgersRef.current = currentLedgers;
        isMockFleetRef.current = isMockFleet;
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
        onMockChange?.(isMockFleet || detailRes.isMock);
      } else {
        setDetail(null);
        onMockChange?.(isMockFleet);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load ticket details';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [activeTicketId, onMockChange]);

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
  const subTurns = eff.subagent_turn_velocity;

  const ticketOptions = ledgers.map((l) => ({
    ticket_id: l.ticket_id ?? 'unknown',
    ticket_title: l.ticket_title,
    context_health_status: l.workflow_efficiency?.context_window_health?.status,
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
            <span className="ticket-type-pill">{ledger.ticket_type ?? 'feature'}</span>
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
          tooltip="Total planner/coder/tester/reviewer iterations"
        />
        <KPICard
          title="Turnaround Lead Time"
          value={formatLeadTime(leadSecs)}
          secondary={leadSecs ? `${leadSecs.toLocaleString()}s elapsed` : 'Active'}
          tooltip="Clock duration from first event to completion"
        />
        <KPICard
          title="Active Compute Duration"
          value={eff.active_duration_formatted || formatDurationSeconds(eff.active_duration_seconds)}
          secondary={`${eff.active_duration_seconds.toLocaleString()}s total`}
          tooltip="Sum of active invocation durations"
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
            <span className="stat-val mono">{ledger.facts?.tools?.length ?? 0}</span>
          </div>
        </div>
      </div>

      {/* Row 3: Context Window Health & Saturation */}
      <SaturationBar contextHealth={eff.context_window_health} />

      {/* Row 4: Subagent Turn Velocity */}
      <div className="turn-velocity-card">
        <div className="turn-velocity-header">
          <h4 className="section-title">Subagent Turn Velocity</h4>
          <div className="turn-velocity-badges">
            <span className="velocity-badge">Avg Turns: {subTurns.avg_turns_per_invocation.toFixed(1)}</span>
            <span className="velocity-badge">Peak Turns: {subTurns.peak_turns}</span>
          </div>
        </div>
        {turnEntries.length > 0 ? (
          <div className="turns-by-role-list">
            {turnEntries.map(([role, turns]) => (
              <div key={role} className="turn-role-row">
                <span className="turn-role-label">{role}</span>
                <div className="turn-track">
                  <div
                    className="turn-fill"
                    style={{
                      width: `${Math.min(100, (turns / Math.max(1, subTurns.peak_turns)) * 100)}%`,
                    }}
                  />
                </div>
                <span className="turn-role-val">{turns.toFixed(1)} turns/inv</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="empty-turns-note">No subagent turn records available.</div>
        )}
      </div>

      {/* Row 5: Role and Model Metrics Tables */}
      <div className="tables-grid">
        <div className="table-card">
          <h4 className="table-card-title">Role Resource Consumption</h4>
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Role</th>
                  <th className="num-col">Invocations</th>
                  <th className="num-col">Tokens</th>
                  <th className="num-col">Cache Hit</th>
                  <th className="num-col">Cost</th>
                </tr>
              </thead>
              <tbody>
                {ledger.roles.map((r) => (
                  <tr key={r.role}>
                    <td className="role-col">{r.role}</td>
                    <td className="num-col">{r.invocations}</td>
                    <td className="num-col">{formatTokensK(r.input_tokens + r.output_tokens)}</td>
                    <td className="num-col">{formatPercentage(r.cache_hit_rate_pct)}</td>
                    <td className="num-col">{formatCurrencyUSD(r.cost_usd)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="table-card">
          <h4 className="table-card-title">Model Resource Consumption</h4>
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Model</th>
                  <th className="num-col">Invocations</th>
                  <th className="num-col">Tokens</th>
                  <th className="num-col">Cache Hit</th>
                  <th className="num-col">Cost</th>
                </tr>
              </thead>
              <tbody>
                {ledger.models.map((m) => (
                  <tr key={m.model}>
                    <td className="model-col">{m.model}</td>
                    <td className="num-col">{m.invocations}</td>
                    <td className="num-col">{formatTokensK(m.input_tokens + m.output_tokens)}</td>
                    <td className="num-col">{formatPercentage(m.cache_hit_rate_pct)}</td>
                    <td className="num-col">{formatCurrencyUSD(m.cost_usd)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Row 6: Session Thread Logs */}
      <ThreadViewer threads={ledger.threads} />

      {/* Row 7: Workspace Artifacts */}
      <ArtifactViewer artifacts={artifacts} />
    </div>
  );
};
