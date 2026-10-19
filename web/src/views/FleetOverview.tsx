import React, { useEffect, useState, useCallback } from 'react';
import type { FleetSummary, TicketComparisonRow, TimeWindowPreset } from '../domain/types';
import {
  formatCurrencyUSD,
  formatPercentage,
  formatTokensK,
} from '../domain/formatters';
import { fetchFleetSummary, fetchWorkspaces } from '../api/client';
import { KPICard } from '../components/KPICard';
import { DistributionCard } from '../components/DistributionCard';
import { TicketTable } from '../components/TicketTable';
import { EmptyState } from '../components/EmptyState';
import { FleetAnomalyCard } from '../components/FleetAnomalyCard';

export interface FleetOverviewProps {
  readonly onMockChange?: (isMock: boolean) => void;
  readonly onSelectTicket?: (ticketId: string) => void;
}

export const FleetOverview: React.FC<FleetOverviewProps> = ({
  onMockChange,
  onSelectTicket,
}) => {
  const [summary, setSummary] = useState<FleetSummary | null>(null);
  const [workspaces, setWorkspaces] = useState<TicketComparisonRow[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [windowPreset, setWindowPreset] = useState<TimeWindowPreset>('all');
  const [customSince, setCustomSince] = useState<string>('');
  const [customUntil, setCustomUntil] = useState<string>('');

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    try {
      const windowParam = windowPreset === 'custom' ? 'all' : windowPreset;
      const sinceParam = windowPreset === 'custom' && customSince ? customSince : undefined;
      const untilParam = windowPreset === 'custom' && customUntil ? customUntil : undefined;

      const [fleetRes, workspacesRes] = await Promise.all([
        fetchFleetSummary(windowParam, sinceParam, untilParam),
        fetchWorkspaces(windowParam, '', '', sinceParam, untilParam),
      ]);

      setSummary(fleetRes.data);
      setWorkspaces(workspacesRes.data);
      const isAnyMock = fleetRes.isMock || workspacesRes.isMock;
      onMockChange?.(isAnyMock);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to retrieve telemetry data';
      setErrorMessage(msg);
    } finally {
      setLoading(false);
    }
  }, [windowPreset, customSince, customUntil, onMockChange]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const dateFilterControl = (
    <section className="date-filter-bar" aria-label="Date Range Filter">
      <div className="date-filter-group">
        <label htmlFor="date-range-preset" className="date-filter-label">
          Date Range Filter:
        </label>
        <select
          id="date-range-preset"
          className="date-filter-select"
          value={windowPreset}
          onChange={(e) => setWindowPreset(e.target.value as TimeWindowPreset)}
          aria-label="Date Range Preset"
        >
          <option value="all">Project Lifetime</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
          <option value="90d">Last 90 Days</option>
          <option value="custom">Custom Range</option>
        </select>
      </div>

      {windowPreset === 'custom' && (
        <div className="custom-date-inputs">
          <div className="date-input-item">
            <label htmlFor="custom-since-input" className="date-input-label">
              Since:
            </label>
            <input
              id="custom-since-input"
              type="date"
              className="date-picker-input"
              value={customSince}
              onChange={(e) => setCustomSince(e.target.value)}
              aria-label="Custom Start Date"
            />
          </div>
          <div className="date-input-item">
            <label htmlFor="custom-until-input" className="date-input-label">
              Until:
            </label>
            <input
              id="custom-until-input"
              type="date"
              className="date-picker-input"
              value={customUntil}
              onChange={(e) => setCustomUntil(e.target.value)}
              aria-label="Custom End Date"
            />
          </div>
        </div>
      )}
    </section>
  );

  return (
    <div>
      {/* Date Range Filter Bar */}
      {dateFilterControl}

      {loading ? (
        <div className="loading-container">
          <span>Loading telemetry metrics...</span>
        </div>
      ) : errorMessage ? (
        <div className="error-banner">
          <span>{errorMessage}</span>
          <button type="button" className="btn-primary" onClick={loadData}>
            Retry Connection
          </button>
        </div>
      ) : !summary || summary.total_tickets === 0 ? (
        <EmptyState
          title="Zero Telemetry Workspaces Available"
          message="No workspaces with recorded session metrics were discovered. Ensure tickets contain agentic_metrics.json ledgers."
          actionLabel="Refresh Data"
          onAction={loadData}
        />
      ) : (
        <>
          {/* KPI Cards Row */}
          <section className="kpi-row" aria-label="Fleet Key Performance Indicators">
            <KPICard
              title="Total Cost"
              value={formatCurrencyUSD(summary.total_cost_usd)}
              secondary={`${summary.total_tickets} active tickets`}
              tooltip="Cumulative LLM spend across all scanned ticket workspaces"
            />
            <KPICard
              title="Invocations"
              value={summary.total_invocations.toLocaleString()}
              secondary={`${summary.total_tickets > 0 ? (summary.total_invocations / summary.total_tickets).toFixed(1) : '0'} avg / ticket`}
              tooltip="Total atomic subagent calls dispatched"
            />
            <KPICard
              title="Total Tokens"
              value={formatTokensK(
                summary.total_input_tokens +
                summary.total_output_tokens +
                summary.total_cache_read_tokens
              )}
              secondary={`${formatTokensK(summary.total_cache_read_tokens)} cached`}
              tooltip="Total combined input, output, and cached prompt tokens"
            />
            <KPICard
              title="Cache Hit Rate"
              value={formatPercentage(summary.overall_cache_hit_rate_pct)}
              secondary="Prompt caching efficiency"
              tooltip="Percentage of prompt context read directly from KV prompt cache"
            />
            <KPICard
              title="Active Duration"
              value={summary.formatted_total_duration}
              secondary={`${summary.total_files_changed} files touched`}
              tooltip="Cumulative active agent execution time across all tickets (cumulative man-hours)."
            />
          </section>

          {/* Fleet Anomaly and Reliability Profile Card */}
          <section aria-label="Fleet Anomaly Profile">
            <FleetAnomalyCard
              summary={summary.anomaly_summary}
              onViewTicket={onSelectTicket}
            />
          </section>

          {/* Distribution Cards Row */}
          <section className="distribution-row" aria-label="Cost Distributions">
            <DistributionCard
              title="Role Cost Distribution"
              items={summary.role_costs}
              totalCostUSD={summary.total_cost_usd}
              tooltip="Breakdown of cumulative LLM inference spend categorized by agent persona role (orchestrator, coder, tester, reviewer, planner)"
            />
            <DistributionCard
              title="Model Cost Distribution"
              items={summary.model_costs}
              totalCostUSD={summary.total_cost_usd}
              tooltip="Breakdown of cumulative LLM inference spend categorized by model family and version"
            />
          </section>

          {/* Ticket Table */}
          <section aria-label="Ticket Directory">
            <TicketTable
              rows={workspaces}
              onSelectTicket={onSelectTicket}
            />
          </section>
        </>
      )}
    </div>
  );
};
