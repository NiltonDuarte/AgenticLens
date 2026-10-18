import React, { useEffect, useState, useCallback } from 'react';
import type { FleetSummary, TicketComparisonRow } from '../domain/types';
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

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    try {
      const [fleetRes, workspacesRes] = await Promise.all([
        fetchFleetSummary(),
        fetchWorkspaces(),
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
  }, [onMockChange]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  if (loading) {
    return (
      <div className="loading-container">
        <span>Loading telemetry metrics...</span>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div>
        <div className="error-banner">
          <span>{errorMessage}</span>
          <button type="button" className="btn-primary" onClick={loadData}>
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  if (!summary || summary.total_tickets === 0) {
    return (
      <EmptyState
        title="Zero Telemetry Workspaces Available"
        message="No workspaces with recorded session metrics were discovered. Ensure tickets contain agentic_metrics.json ledgers."
        actionLabel="Refresh Data"
        onAction={loadData}
      />
    );
  }

  const totalTokens =
    summary.total_input_tokens +
    summary.total_output_tokens +
    summary.total_cache_read_tokens;

  return (
    <div>
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
          value={formatTokensK(totalTokens)}
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
    </div>
  );
};
