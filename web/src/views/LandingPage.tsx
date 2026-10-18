import React, { useEffect, useState, useCallback, useMemo } from 'react';
import type { ActiveTab, FleetSummary, TicketComparisonRow } from '../domain/types';
import { compareTicketIds, formatCurrencyUSD, formatPercentage } from '../domain/formatters';
import { fetchFleetSummary, fetchWorkspaces } from '../api/client';
import { EmptyState } from '../components/EmptyState';

export type QuickLaunchSortField =
  | 'ticket_id'
  | 'ticket_title'
  | 'context_health_status'
  | 'sdlc_iterations'
  | 'active_duration_seconds';

export type QuickLaunchSortOrder = 'asc' | 'desc';

export interface LandingPageProps {
  readonly onNavigate: (tab: ActiveTab) => void;
  readonly onSelectTicket: (ticketId: string) => void;
  readonly onMockChange?: (isMock: boolean) => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({
  onNavigate,
  onSelectTicket,
  onMockChange,
}) => {
  const [summary, setSummary] = useState<FleetSummary | null>(null);
  const [workspaces, setWorkspaces] = useState<readonly TicketComparisonRow[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [sortField, setSortField] = useState<QuickLaunchSortField>('ticket_id');
  const [sortOrder, setSortOrder] = useState<QuickLaunchSortOrder>('asc');

  const handleSort = (field: QuickLaunchSortField) => {
    if (sortField === field) {
      setSortOrder((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  const getSortIndicator = (field: QuickLaunchSortField): string => {
    if (sortField !== field) return '';
    return sortOrder === 'asc' ? ' (asc)' : ' (desc)';
  };

  const sortedWorkspaces = useMemo(() => {
    return [...workspaces].sort((a, b) => {
      let diff = 0;
      switch (sortField) {
        case 'ticket_id':
          diff = compareTicketIds(a.ticket_id, b.ticket_id);
          break;
        case 'ticket_title':
          diff = a.ticket_title.localeCompare(b.ticket_title);
          break;
        case 'context_health_status':
          diff = a.context_health_status.localeCompare(b.context_health_status);
          break;
        case 'sdlc_iterations':
          diff = a.sdlc_iterations - b.sdlc_iterations;
          break;
        case 'active_duration_seconds':
          diff = a.active_duration_seconds - b.active_duration_seconds;
          break;
      }
      if (diff !== 0) {
        return sortOrder === 'asc' ? diff : -diff;
      }
      return compareTicketIds(a.ticket_id, b.ticket_id);
    });
  }, [workspaces, sortField, sortOrder]);

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
        <span>Loading observability portal...</span>
      </div>
    );
  }

  if (errorMessage) {
    return (
      <div className="landing-container">
        <div className="error-banner">
          <span>{errorMessage}</span>
          <button type="button" className="btn-primary" onClick={loadData}>
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  const totalWorkspaces = summary?.total_tickets ?? workspaces.length;
  const totalSpend = summary ? formatCurrencyUSD(summary.total_cost_usd) : '$0.00';
  const totalIterations = workspaces.reduce(
    (acc, w) => acc + (Number.isFinite(w.sdlc_iterations) ? w.sdlc_iterations : 0),
    0
  );
  const cacheHitEfficiency = summary
    ? formatPercentage(summary.overall_cache_hit_rate_pct)
    : '0.0%';

  return (
    <div className="landing-container">
      {/* Hero Section */}
      <section className="landing-hero" aria-label="Observability Platform Overview">
        <div className="hero-badge">Autonomous Multi-Agent Telemetry</div>
        <h1 className="hero-title">AgenticLens Observability</h1>
        <p className="hero-subtitle">
          Real-time telemetry, context health tracking, and unit economics for autonomous multi-agent software engineering fleets.
        </p>
        <p className="hero-lead">
          Built upon the Functional Core, Imperative Shell (FCIS) architecture, AgenticLens delivers granular operational intelligence for AI-driven development. Monitor context window saturation against model ceilings, evaluate subagent turn velocity, audit token unit economics, and inspect SDLC execution threads across every ticket workspace.
        </p>
        <div className="hero-tags">
          <span className="hero-tag">Functional Core, Imperative Shell</span>
          <span className="hero-tag">Context Window Saturation</span>
          <span className="hero-tag">Token Unit Economics</span>
          <span className="hero-tag">Multi-Agent Verification</span>
        </div>
      </section>

      {/* Pulse Metrics Bar */}
      <section className="pulse-metrics-grid" aria-label="Pulse Metrics">
        <div className="pulse-card">
          <span className="pulse-label">Total Ticket Workspaces</span>
          <span className="pulse-value mono-number">{totalWorkspaces}</span>
          <span className="pulse-subtext">Registered SDLC workspaces</span>
        </div>
        <div className="pulse-card">
          <span className="pulse-label">Cumulative Fleet Spend</span>
          <span className="pulse-value mono-number">{totalSpend}</span>
          <span className="pulse-subtext">Total model inference costs</span>
        </div>
        <div className="pulse-card">
          <span className="pulse-label">Total Iterations</span>
          <span className="pulse-value mono-number">{totalIterations}</span>
          <span className="pulse-subtext">Completed SDLC cycles</span>
        </div>
        <div className="pulse-card">
          <span className="pulse-label">Cache Hit Efficiency</span>
          <span className="pulse-value mono-number">{cacheHitEfficiency}</span>
          <span className="pulse-subtext">Prompt cache read ratio</span>
        </div>
      </section>

      {/* Feature Navigation Cards */}
      <section className="feature-cards-grid" aria-label="Analytical Perspectives">
        <div className="feature-nav-card">
          <div className="feature-card-header">
            <span className="feature-card-badge">Macro Fleet</span>
            <h2 className="feature-card-title">Fleet Overview</h2>
          </div>
          <p className="feature-card-desc">
            Macro-level spend, total invocations, token throughput, and role/model distributions across all active tickets.
          </p>
          <button
            type="button"
            className="feature-card-btn"
            onClick={() => onNavigate('overview')}
          >
            Open Fleet Overview
          </button>
        </div>

        <div className="feature-nav-card">
          <div className="feature-card-header">
            <span className="feature-card-badge">Single Ticket</span>
            <h2 className="feature-card-title">Ticket Deep Dive</h2>
          </div>
          <p className="feature-card-desc">
            Single-ticket autopsy, context ceiling saturation, subagent turn velocity, and workspace artifact reader.
          </p>
          <button
            type="button"
            className="feature-card-btn"
            onClick={() => onNavigate('deep-dive')}
          >
            Explore Deep Dive
          </button>
        </div>

        <div className="feature-nav-card">
          <div className="feature-card-header">
            <span className="feature-card-badge">Temporal Trends</span>
            <h2 className="feature-card-title">Metric Progression</h2>
          </div>
          <p className="feature-card-desc">
            Longitudinal temporal velocity analysis tracking period spend, invocations, token throughput, and code churn.
          </p>
          <button
            type="button"
            className="feature-card-btn"
            onClick={() => onNavigate('progression')}
          >
            View Progression
          </button>
        </div>

        <div className="feature-nav-card">
          <div className="feature-card-header">
            <span className="feature-card-badge">Benchmarking</span>
            <h2 className="feature-card-title">Ticket Comparisons</h2>
          </div>
          <p className="feature-card-desc">
            Cross-ticket benchmarking matrix comparing turnaround lead time, active compute, test-to-prod ratios, and context saturation.
          </p>
          <button
            type="button"
            className="feature-card-btn"
            onClick={() => onNavigate('comparisons')}
          >
            Compare Tickets
          </button>
        </div>
      </section>

      {/* Quick-Launch Workspace Directory */}
      <section className="quick-launch-section" aria-label="Registered Workspaces">
        <div className="section-header-bar">
          <h2 className="section-title">Quick-Launch Workspaces</h2>
          <span className="section-subtitle">Direct access to ticket deep-dive autopsies</span>
        </div>

        {workspaces.length === 0 ? (
          <EmptyState
            title="Zero Workspaces Discovered"
            message="No ticket workspaces with recorded telemetry were discovered. Ensure tickets contain valid agentic_metrics.json ledgers."
          />
        ) : (
          <div className="quick-launch-table-wrapper">
            <table className="quick-launch-table">
              <thead>
                <tr>
                  <th
                    className="sortable"
                    onClick={() => handleSort('ticket_id')}
                    aria-sort={sortField === 'ticket_id' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    Ticket ID{getSortIndicator('ticket_id')}
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('ticket_title')}
                    aria-sort={sortField === 'ticket_title' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    Title{getSortIndicator('ticket_title')}
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('context_health_status')}
                    aria-sort={sortField === 'context_health_status' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    Status{getSortIndicator('context_health_status')}
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('sdlc_iterations')}
                    aria-sort={sortField === 'sdlc_iterations' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    Iterations{getSortIndicator('sdlc_iterations')}
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('active_duration_seconds')}
                    aria-sort={sortField === 'active_duration_seconds' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    Active Duration{getSortIndicator('active_duration_seconds')}
                  </th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {sortedWorkspaces.map((w) => {
                  const statusClass = ['HEALTHY', 'ELEVATED', 'BLOATED', 'CRITICAL'].includes(
                    w.context_health_status
                  )
                    ? w.context_health_status
                    : 'HEALTHY';

                  return (
                    <tr
                      key={w.ticket_id}
                      className="quick-launch-row"
                      onClick={() => onSelectTicket(w.ticket_id)}
                    >
                      <td className="mono-number" style={{ fontWeight: 600 }}>
                        {w.ticket_id}
                      </td>
                      <td className="ticket-title-cell">{w.ticket_title}</td>
                      <td>
                        <span className={`health-badge ${statusClass}`}>
                          {w.context_health_status}
                        </span>
                      </td>
                      <td className="mono-number">
                        {Number.isFinite(w.sdlc_iterations) ? w.sdlc_iterations : 0}
                      </td>
                      <td className="mono-number">{w.active_duration_formatted}</td>
                      <td>
                        <button
                          type="button"
                          className="quick-launch-btn"
                          onClick={(e) => {
                            e.stopPropagation();
                            onSelectTicket(w.ticket_id);
                          }}
                          aria-label={`Open deep dive for ${w.ticket_id}`}
                        >
                          Inspect Ticket
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
};
