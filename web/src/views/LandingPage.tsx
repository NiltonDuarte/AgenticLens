import React, { useEffect, useState, useCallback, useMemo } from 'react';
import type { ActiveTab, FleetSummary, ProjectConfigResponse, TicketComparisonRow, WorkspaceDiagnostic } from '../domain/types';
import { compareTicketIds, formatCurrencyUSD, formatPercentage } from '../domain/formatters';
import { fetchFleetSummary, fetchWorkspaces } from '../api/client';
import { EmptyState } from '../components/EmptyState';
import { InfoTooltip } from '../components/InfoTooltip';
import { mergeWorkspaceDiagnostics, WorkspaceLoadWarning } from '../components/WorkspaceLoadWarning';

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
  readonly projectConfig?: ProjectConfigResponse | null | undefined;
  readonly projectName?: string | undefined;
}

export const LandingPage: React.FC<LandingPageProps> = ({
  onNavigate,
  onSelectTicket,
  projectConfig,
  projectName,
}) => {
  const [summary, setSummary] = useState<FleetSummary | null>(null);
  const [workspaces, setWorkspaces] = useState<readonly TicketComparisonRow[]>([]);
  const [diagnostics, setDiagnostics] = useState<readonly WorkspaceDiagnostic[]>([]);
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
      setDiagnostics(mergeWorkspaceDiagnostics(fleetRes.diagnostics ?? [], workspacesRes.diagnostics ?? []));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to retrieve telemetry data';
      setErrorMessage(msg);
    } finally {
      setLoading(false);
    }
  }, []);

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
  const observedProject = (projectName ?? projectConfig?.project_name ?? '').trim();

  return (
    <div className="landing-container">
      <WorkspaceLoadWarning diagnostics={diagnostics} />
      {/* Hero Section */}
      <section className="landing-hero" aria-label="Observability Platform Overview">
        <div className="hero-badge">Autonomous Multi-Agent Telemetry</div>
        <h1 className="hero-title">AgenticLens Observability</h1>
        {observedProject ? (
          <div className="hero-project-context" data-testid="landing-project-context">
            <span className="hero-project-label">Observed Project:</span>{' '}
            <span className="hero-project-name">{observedProject}</span>
          </div>
        ) : null}
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
          <div className="pulse-card-header">
            <span className="pulse-label">Total Ticket Workspaces</span>
            <InfoTooltip
              content="Total number of registered agentic development workspaces with recorded telemetry"
              ariaLabel="Information about Total Ticket Workspaces"
            />
          </div>
          <span className="pulse-value mono-number">{totalWorkspaces}</span>
          <span className="pulse-subtext">Registered SDLC workspaces</span>
        </div>
        <div className="pulse-card">
          <div className="pulse-card-header">
            <span className="pulse-label">Cumulative Fleet Spend</span>
            <InfoTooltip
              content="Total monetary expenditure for LLM model inference across all workspaces"
              ariaLabel="Information about Cumulative Fleet Spend"
            />
          </div>
          <span className="pulse-value mono-number">{totalSpend}</span>
          <span className="pulse-subtext">Total model inference costs</span>
        </div>
        <div className="pulse-card">
          <div className="pulse-card-header">
            <span className="pulse-label">Total Iterations</span>
            <InfoTooltip
              content="Total number of completed SDLC development iterations across all workspaces"
              ariaLabel="Information about Total Iterations"
            />
          </div>
          <span className="pulse-value mono-number">{totalIterations}</span>
          <span className="pulse-subtext">Completed SDLC cycles</span>
        </div>
        <div className="pulse-card">
          <div className="pulse-card-header">
            <span className="pulse-label">Cache Hit Efficiency</span>
            <InfoTooltip
              content="Overall percentage of prompt context tokens served from prompt cache"
              ariaLabel="Information about Cache Hit Efficiency"
            />
          </div>
          <span className="pulse-value mono-number">{cacheHitEfficiency}</span>
          <span className="pulse-subtext">Prompt cache read ratio</span>
        </div>
      </section>

      {/* Feature Navigation Cards */}
      <section className="feature-cards-grid" aria-label="Analytical Perspectives">
        <div className="feature-nav-card">
          <div className="feature-card-header">
            <div className="feature-card-top-row">
              <span className="feature-card-badge">Macro Fleet</span>
              <InfoTooltip
                content="Macro fleet-level analytics, cumulative spend, token throughput, and role/model distributions"
                ariaLabel="Information about Fleet Overview"
              />
            </div>
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
            <div className="feature-card-top-row">
              <span className="feature-card-badge">Single Ticket</span>
              <InfoTooltip
                content="Detailed autopsy of single ticket workspaces, context saturation, subagent turns, and artifacts"
                ariaLabel="Information about Ticket Deep Dive"
              />
            </div>
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
            <div className="feature-card-top-row">
              <span className="feature-card-badge">Temporal Trends</span>
              <InfoTooltip
                content="Longitudinal temporal progression tracking period spend, invocations, tokens, and code churn"
                ariaLabel="Information about Metric Progression"
              />
            </div>
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
            <div className="feature-card-top-row">
              <span className="feature-card-badge">Benchmarking</span>
              <InfoTooltip
                content="Cross-workspace benchmarking matrix comparing lead time, compute, test ratios, and health"
                ariaLabel="Information about Ticket Comparisons"
              />
            </div>
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
                    <span className="th-content">
                      <span>Ticket ID{getSortIndicator('ticket_id')}</span>
                      <InfoTooltip
                        content="Unique identifier for the development ticket and workspace directory"
                        placement="bottom"
                        ariaLabel="Information about Ticket ID"
                      />
                    </span>
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('ticket_title')}
                    aria-sort={sortField === 'ticket_title' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    <span className="th-content">
                      <span>Title{getSortIndicator('ticket_title')}</span>
                      <InfoTooltip
                        content="Summary title of the feature, bugfix, or task tracked by this ticket"
                        placement="bottom"
                        ariaLabel="Information about Title"
                      />
                    </span>
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('context_health_status')}
                    aria-sort={sortField === 'context_health_status' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    <span className="th-content">
                      <span>Status{getSortIndicator('context_health_status')}</span>
                      <InfoTooltip
                        content="Context window saturation health status (HEALTHY, ELEVATED, BLOATED, or CRITICAL)"
                        placement="bottom"
                        ariaLabel="Information about Status"
                      />
                    </span>
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('sdlc_iterations')}
                    aria-sort={sortField === 'sdlc_iterations' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    <span className="th-content">
                      <span>Iterations{getSortIndicator('sdlc_iterations')}</span>
                      <InfoTooltip
                        content="Number of SDLC feedback cycles executed for this ticket"
                        placement="bottom"
                        ariaLabel="Information about Iterations"
                      />
                    </span>
                  </th>
                  <th
                    className="sortable"
                    onClick={() => handleSort('active_duration_seconds')}
                    aria-sort={sortField === 'active_duration_seconds' ? (sortOrder === 'asc' ? 'ascending' : 'descending') : 'none'}
                  >
                    <span className="th-content">
                      <span>Active Duration{getSortIndicator('active_duration_seconds')}</span>
                      <InfoTooltip
                        content="Cumulative active invocation duration spent on this ticket (cumulative man-hours)."
                        placement="bottom"
                        ariaLabel="Information about Active Duration"
                      />
                    </span>
                  </th>
                  <th>
                    <span className="th-content">
                      <span>Action</span>
                      <InfoTooltip
                        content="Direct navigation action to inspect single ticket autopsy"
                        placement="bottom"
                        ariaLabel="Information about Action"
                      />
                    </span>
                  </th>
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
