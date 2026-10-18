import React, { useEffect, useState, useMemo, useCallback } from 'react';
import type {
  AgenticMetricsLedger,
  ComparisonSortField,
  ComparisonSortOrder,
} from '../domain/types';
import { fetchWorkspaceLedgers } from '../api/client';
import {
  deriveTicketComparisonRows,
  formatCurrencyUSD,
  formatDurationSeconds,
  formatIsoDateTimeShort,
  formatPercentage,
} from '../domain/formatters';
import { KPICard } from '../components/KPICard';
import { EmptyState } from '../components/EmptyState';
import { InfoTooltip } from '../components/InfoTooltip';

export interface TicketComparisonsProps {
  readonly onSelectTicket?: (ticketId: string) => void;
  readonly onMockChange?: (isMock: boolean) => void;
}

export const TicketComparisons: React.FC<TicketComparisonsProps> = ({
  onSelectTicket,
  onMockChange,
}) => {
  const [ledgers, setLedgers] = useState<readonly AgenticMetricsLedger[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [sortField, setSortField] = useState<ComparisonSortField | undefined>(undefined);
  const [sortOrder, setSortOrder] = useState<ComparisonSortOrder>('desc');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadLedgers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetchWorkspaceLedgers('all');
      setLedgers(res.data);
      onMockChange?.(res.isMock);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch comparison workspaces';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [onMockChange]);

  useEffect(() => {
    loadLedgers();
  }, [loadLedgers]);

  const allRows = useMemo(() => {
    return deriveTicketComparisonRows(ledgers, sortField, sortOrder);
  }, [ledgers, sortField, sortOrder]);

  const filteredRows = useMemo(() => {
    if (!searchQuery.trim()) {
      return allRows;
    }
    const q = searchQuery.toLowerCase().trim();
    return allRows.filter(
      (r) =>
        r.ticket_id.toLowerCase().includes(q) ||
        r.ticket_title.toLowerCase().includes(q)
    );
  }, [allRows, searchQuery]);

  // Fleet benchmarking stats
  const benchmarks = useMemo(() => {
    const total = allRows.length;
    if (total === 0) {
      return {
        totalTickets: 0,
        avgDurationFormatted: '0m 0s',
        avgIterations: '0.0',
        avgCostFormatted: '$0.00',
        fleetToolSuccessPct: '0.0%',
      };
    }

    let totDuration = 0;
    let totIterations = 0;
    let totCost = 0;
    let totToolSuccess = 0;

    for (const r of allRows) {
      totDuration += r.active_duration_seconds;
      totIterations += r.sdlc_iterations;
      totCost += r.cost_usd;
      totToolSuccess += r.tool_success_rate_pct;
    }

    const avgDuration = Math.round(totDuration / total);
    const avgIter = (totIterations / total).toFixed(1);
    const avgCost = totCost / total;
    const avgToolPct = totToolSuccess / total;

    return {
      totalTickets: total,
      avgDurationFormatted: formatDurationSeconds(avgDuration),
      avgIterations: avgIter,
      avgCostFormatted: formatCurrencyUSD(avgCost),
      fleetToolSuccessPct: formatPercentage(avgToolPct),
    };
  }, [allRows]);

  const handleHeaderSort = (field: ComparisonSortField) => {
    if (sortField === field) {
      setSortOrder((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortOrder('desc');
    }
  };

  const renderSortIndicator = (field: ComparisonSortField) => {
    if (sortField !== field) {
      return <span className="sort-indicator neutral">-</span>;
    }
    return (
      <span className="sort-indicator active">
        {sortOrder === 'asc' ? '^' : 'v'}
      </span>
    );
  };

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner" />
        <span className="loading-text">Loading cross-ticket comparison matrix...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-banner">
        <span className="error-text">{error}</span>
        <button type="button" className="btn-secondary" onClick={loadLedgers}>
          Retry Connection
        </button>
      </div>
    );
  }

  if (allRows.length === 0) {
    return (
      <EmptyState
        title="Zero Comparison Workspaces Available"
        message="No workspace telemetry files were found. Run an agentic workflow to generate comparison data."
      />
    );
  }

  return (
    <div className="comparisons-container">
      {/* Benchmarking Summary Banner */}
      <div className="kpi-grid">
        <KPICard
          title="Benchmarked Workspaces"
          value={benchmarks.totalTickets.toString()}
          secondary="Catalog volume"
          tooltip="Total active and completed workspaces analyzed"
        />
        <KPICard
          title="Average Active Duration"
          value={benchmarks.avgDurationFormatted}
          secondary="Compute time per ticket"
          tooltip="Average cumulative active compute time across all agent invocations per ticket (cumulative man-hours)."
        />
        <KPICard
          title="Average Iterations"
          value={benchmarks.avgIterations}
          secondary="SDLC turns per ticket"
          tooltip="Average feedback iterations to delivery"
        />
        <KPICard
          title="Fleet Tool Success Rate"
          value={benchmarks.fleetToolSuccessPct}
          secondary="Execution reliability"
          tooltip="Overall tool call success rate across fleet"
        />
        <KPICard
          title="Average Spend / Ticket"
          value={benchmarks.avgCostFormatted}
          secondary="Cost efficiency"
          tooltip="Average USD spend across analyzed tickets"
        />
      </div>

      {/* Search and Filters Bar */}
      <div className="table-controls-bar">
        <div className="search-box">
          <label htmlFor="comparison-search" className="search-label">
            Filter Comparison Rows:
          </label>
          <input
            id="comparison-search"
            type="search"
            className="search-input"
            placeholder="Search by ticket ID or title..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <div className="table-row-count">
          Showing {filteredRows.length} of {allRows.length} tickets
        </div>
      </div>

      {/* Comparison Matrix Table */}
      <div className="table-container">
        <div className="table-responsive">
          <table className="comparison-matrix-table">
            <thead>
              <tr>
                <th
                  onClick={() => handleHeaderSort('ticket_id')}
                  className="sortable-col"
                  aria-label="Ticket ID"
                >
                  <span className="th-content">
                    <span>Ticket ID {renderSortIndicator('ticket_id')}</span>
                    <InfoTooltip
                      content="Unique identifier for the development ticket workspace"
                      placement="bottom"
                      ariaLabel="Information about Ticket ID"
                    />
                  </span>
                </th>
                <th aria-label="Title">
                  <span className="th-content">
                    <span>Title</span>
                    <InfoTooltip
                      content="Descriptive title of the ticket or work item"
                      placement="bottom"
                      ariaLabel="Information about Title"
                    />
                  </span>
                </th>
                <th aria-label="Status">
                  <span className="th-content">
                    <span>Status</span>
                    <InfoTooltip
                      content="Workspace lifecycle status (Completed or In Progress)"
                      placement="bottom"
                      ariaLabel="Information about Status"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('completed_at')}
                  className="sortable-col"
                  aria-label="Completed"
                >
                  <span className="th-content">
                    <span>Completed {renderSortIndicator('completed_at')}</span>
                    <InfoTooltip
                      content="Timestamp when the ticket workspace execution completed"
                      placement="bottom"
                      ariaLabel="Information about Completed"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('started_at')}
                  className="sortable-col"
                  aria-label="Started"
                >
                  <span className="th-content">
                    <span>Started {renderSortIndicator('started_at')}</span>
                    <InfoTooltip
                      content="Timestamp when the ticket workspace execution began"
                      placement="bottom"
                      ariaLabel="Information about Started"
                    />
                  </span>
                </th>
                <th aria-label="Lead Time">
                  <span className="th-content">
                    <span>Lead Time</span>
                    <InfoTooltip
                      content="Total wall-clock elapsed time from first event to completion."
                      placement="bottom"
                      ariaLabel="Information about Lead Time"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('active_duration_seconds')}
                  className="sortable-col num-col"
                  aria-label="Active Time"
                >
                  <span className="th-content">
                    <span>Active Time {renderSortIndicator('active_duration_seconds')}</span>
                    <InfoTooltip
                      content="Cumulative active compute duration across all agent invocations (cumulative man-hours); can exceed elapsed lead time when subagents run in parallel."
                      placement="bottom"
                      ariaLabel="Information about Active Time"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('sdlc_iterations')}
                  className="sortable-col num-col"
                  aria-label="Iterations"
                >
                  <span className="th-content">
                    <span>Iterations {renderSortIndicator('sdlc_iterations')}</span>
                    <InfoTooltip
                      content="Total number of SDLC feedback cycles executed"
                      placement="bottom"
                      ariaLabel="Information about Iterations"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('cost_usd')}
                  className="sortable-col num-col"
                  aria-label="Cost (USD)"
                >
                  <span className="th-content">
                    <span>Cost (USD) {renderSortIndicator('cost_usd')}</span>
                    <InfoTooltip
                      content="Total LLM inference spend in US dollars"
                      placement="bottom"
                      ariaLabel="Information about Cost (USD)"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('cache_hit_rate_pct')}
                  className="sortable-col num-col"
                  aria-label="Cache Hit"
                >
                  <span className="th-content">
                    <span>Cache Hit {renderSortIndicator('cache_hit_rate_pct')}</span>
                    <InfoTooltip
                      content="Percentage of prompt tokens read from KV prompt cache"
                      placement="bottom"
                      ariaLabel="Information about Cache Hit"
                    />
                  </span>
                </th>
                <th className="num-col" aria-label="Test:Prod">
                  <span className="th-content">
                    <span>Test:Prod</span>
                    <InfoTooltip
                      content="Ratio of test lines added to production lines added"
                      placement="bottom"
                      ariaLabel="Information about Test:Prod"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('lines_added')}
                  className="sortable-col num-col"
                  aria-label="Code Lines"
                >
                  <span className="th-content">
                    <span>Code Lines {renderSortIndicator('lines_added')}</span>
                    <InfoTooltip
                      content="Total code churn: lines added, lines deleted, and files changed"
                      placement="bottom"
                      ariaLabel="Information about Code Lines"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('tool_success_rate_pct')}
                  className="sortable-col num-col"
                  aria-label="Tool Success"
                >
                  <span className="th-content">
                    <span>Tool Success {renderSortIndicator('tool_success_rate_pct')}</span>
                    <InfoTooltip
                      content="Percentage of subagent tool executions that succeeded without error"
                      placement="bottom"
                      ariaLabel="Information about Tool Success"
                    />
                  </span>
                </th>
                <th aria-label="Context Health">
                  <span className="th-content">
                    <span>Context Health</span>
                    <InfoTooltip
                      content="Context window saturation evaluation status (HEALTHY, ELEVATED, BLOATED, CRITICAL)"
                      placement="bottom"
                      ariaLabel="Information about Context Health"
                    />
                  </span>
                </th>
                <th
                  onClick={() => handleHeaderSort('orchestrator_saturation_pct')}
                  className="sortable-col num-col"
                  aria-label="Saturation"
                >
                  <span className="th-content">
                    <span>Saturation {renderSortIndicator('orchestrator_saturation_pct')}</span>
                    <InfoTooltip
                      content="Peak context window utilization percentage relative to the 200k token ceiling"
                      placement="bottom"
                      ariaLabel="Information about Saturation"
                    />
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredRows.length === 0 ? (
                <tr>
                  <td colSpan={15} className="empty-table-cell">
                    No tickets match the query "{searchQuery}".
                  </td>
                </tr>
              ) : (
                filteredRows.map((row) => {
                  const isCompleted = row.completed_at !== null;
                  return (
                    <tr
                      key={row.ticket_id}
                      className="clickable-comparison-row"
                      onClick={() => onSelectTicket?.(row.ticket_id)}
                      title="Click to view Ticket Deep Dive autopsy"
                    >
                      <td className="ticket-id-cell">
                        <button
                          type="button"
                          className="ticket-link-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            onSelectTicket?.(row.ticket_id);
                          }}
                        >
                          {row.ticket_id}
                        </button>
                      </td>
                      <td className="ticket-title-cell" title={row.ticket_title}>
                        {row.ticket_title}
                      </td>
                      <td>
                        <span className={`status-pill ${isCompleted ? 'live' : 'mock'}`}>
                          {isCompleted ? 'Completed' : 'In Progress'}
                        </span>
                      </td>
                      <td className="date-cell">{formatIsoDateTimeShort(row.completed_at)}</td>
                      <td className="date-cell">{formatIsoDateTimeShort(row.started_at)}</td>
                      <td className="mono">{row.lead_duration_formatted}</td>
                      <td className="num-col mono">{row.active_duration_formatted}</td>
                      <td className="num-col mono">{row.sdlc_iterations}</td>
                      <td className="num-col mono">{formatCurrencyUSD(row.cost_usd)}</td>
                      <td className="num-col mono">{formatPercentage(row.cache_hit_rate_pct)}</td>
                      <td className="num-col mono">{row.test_to_code_ratio}</td>
                      <td className="num-col mono">
                        +{row.lines_added} / -{row.lines_deleted} ({row.files_changed}f)
                      </td>
                      <td className="num-col mono">
                        {formatPercentage(row.tool_success_rate_pct)}
                      </td>
                      <td>
                        <span className={`health-pill mini ${row.context_health_status.toLowerCase()}`}>
                          {row.context_health_status}
                        </span>
                      </td>
                      <td className="num-col mono">
                        {formatPercentage(row.orchestrator_saturation_pct)}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
