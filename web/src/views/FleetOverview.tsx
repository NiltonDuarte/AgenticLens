import { durationDescription } from '../domain/duration';
import React, { useEffect, useState, useCallback, useMemo, useRef } from 'react';
import type { FleetSummary, TicketComparisonRow, TimeWindowPreset, WorkspaceDiagnostic } from '../domain/types';
import {
  formatPercentage,
  formatTokensK,
  extractTicketPrefix,
  combineDistinctPrefixes,
} from '../domain/formatters';
import { fetchFleetSummary, fetchWorkspaces, fetchWorkspaceLedgers } from '../api/client';
import { deriveSessionOverlaps } from '../domain/sessionOverlaps';
import { SessionOverlapWarning, type SessionOverlapInspection } from '../components/SessionOverlapWarning';
import { KPICard } from '../components/KPICard';
import { CostKPICard } from '../components/CostKPICard';
import { DistributionCard } from '../components/DistributionCard';
import { TicketTable } from '../components/TicketTable';
import { EmptyState } from '../components/EmptyState';
import { FleetAnomalyCard } from '../components/FleetAnomalyCard';
import { MultiSelectDropdown, type MultiSelectOption } from '../components/MultiSelectDropdown';
import { mergeWorkspaceDiagnostics, WorkspaceLoadWarning } from '../components/WorkspaceLoadWarning';

const TICKET_TYPE_OPTIONS: readonly MultiSelectOption[] = [
  { value: 'feature', label: 'Feature' },
  { value: 'bug', label: 'Bug' },
  { value: 'chore', label: 'Chore' },
  { value: 'docs', label: 'Docs' },
  { value: 'refactor', label: 'Refactor' },
  { value: 'test', label: 'Test' },
  { value: 'ci', label: 'CI' },
  { value: 'perf', label: 'Perf' },
  { value: 'build', label: 'Build' },
];

export interface FleetOverviewProps {
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
  readonly configuredPrefixes?: readonly string[] | undefined;
}

export const FleetOverview: React.FC<FleetOverviewProps> = ({
  onSelectTicket,
  configuredPrefixes,
}) => {
  const [summary, setSummary] = useState<FleetSummary | null>(null);
  const [workspaces, setWorkspaces] = useState<TicketComparisonRow[]>([]);
  const [diagnostics, setDiagnostics] = useState<readonly WorkspaceDiagnostic[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const [inspection, setInspection] = useState<SessionOverlapInspection>({ state: 'loading' });
  const inspectionRequest = useRef(0);
  const metricsRequest = useRef(0);

  const loadInspection = useCallback(async () => {
    const request = ++inspectionRequest.current;
    setInspection({ state: 'loading' });
    try {
      const response = await fetchWorkspaceLedgers();
      if (request !== inspectionRequest.current) return;
      setInspection({ state: 'ready', projection: deriveSessionOverlaps(response.data), diagnostics: response.diagnostics ?? [] });
    } catch (err: unknown) {
      if (request !== inspectionRequest.current) return;
      setInspection({ state: 'unavailable', message: err instanceof Error ? err.message : 'Failed to inspect project ledgers.' });
    }
  }, []);

  useEffect(() => {
    const requests = inspectionRequest;
    void loadInspection();
    return () => { requests.current++; };
  }, [loadInspection]);

  const [windowPreset, setWindowPreset] = useState<TimeWindowPreset>('all');
  const [customSince, setCustomSince] = useState<string>('');
  const [customUntil, setCustomUntil] = useState<string>('');
  const [selectedTicketTypes, setSelectedTicketTypes] = useState<readonly string[]>(['all']);
  const [selectedPrefixes, setSelectedPrefixes] = useState<readonly string[]>(['all']);
  const [knownPrefixes, setKnownPrefixes] = useState<readonly string[]>([]);

  useEffect(() => {
    if (workspaces.length > 0) {
      setKnownPrefixes((prev) => {
        const nextSet = new Set(prev);
        for (const w of workspaces) {
          const p = extractTicketPrefix(w.ticket_id);
          if (p) nextSet.add(p);
        }
        return Array.from(nextSet).sort();
      });
    }
  }, [workspaces]);

  const prefixOptions = useMemo(() => {
    const combined = combineDistinctPrefixes(knownPrefixes, configuredPrefixes);
    return combined.map((p) => ({ value: p, label: p }));
  }, [knownPrefixes, configuredPrefixes]);

  const loadData = useCallback(async () => {
    const request = ++metricsRequest.current;
    setLoading(true);
    setErrorMessage(null);
    try {
      const windowParam = windowPreset === 'custom' ? 'all' : windowPreset;
      const sinceParam = windowPreset === 'custom' && customSince ? customSince : undefined;
      const untilParam = windowPreset === 'custom' && customUntil ? customUntil : undefined;

      const typesParam =
        selectedTicketTypes.length === 0 || (selectedTicketTypes.length === 1 && selectedTicketTypes[0] === 'all')
          ? undefined
          : selectedTicketTypes;
      const prefixesParam =
        selectedPrefixes.length === 0 || (selectedPrefixes.length === 1 && selectedPrefixes[0] === 'all')
          ? undefined
          : selectedPrefixes;

      const [fleetRes, workspacesRes] = await Promise.all([
        typesParam || prefixesParam
          ? fetchFleetSummary(windowParam, sinceParam, untilParam, typesParam, prefixesParam)
          : fetchFleetSummary(windowParam, sinceParam, untilParam),
        typesParam || prefixesParam
          ? fetchWorkspaces(windowParam, '', '', sinceParam, untilParam, typesParam, prefixesParam)
          : fetchWorkspaces(windowParam, '', '', sinceParam, untilParam),
      ]);

      if (request !== metricsRequest.current) return;
      setSummary(fleetRes.data);
      setWorkspaces(workspacesRes.data);
      setDiagnostics(mergeWorkspaceDiagnostics(fleetRes.diagnostics ?? [], workspacesRes.diagnostics ?? []));
    } catch (err: unknown) {
      if (request !== metricsRequest.current) return;
      const msg = err instanceof Error ? err.message : 'Failed to retrieve telemetry data';
      setErrorMessage(msg);
    } finally {
      if (request === metricsRequest.current) setLoading(false);
    }
  }, [windowPreset, customSince, customUntil, selectedTicketTypes, selectedPrefixes]);

  useEffect(() => {
    const requests = metricsRequest;
    void loadData();
    return () => { requests.current++; };
  }, [loadData]);

  const refreshData = () => {
    void loadData();
    void loadInspection();
  };

  const skillAdoption = summary?.skill_adoption ?? [];

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

      <MultiSelectDropdown
        id="fleet-ticket-type"
        label="Ticket Type"
        options={TICKET_TYPE_OPTIONS}
        selectedValues={selectedTicketTypes}
        onChange={setSelectedTicketTypes}
      />

      <MultiSelectDropdown
        id="fleet-prefix"
        label="Prefix"
        options={prefixOptions}
        selectedValues={selectedPrefixes}
        onChange={setSelectedPrefixes}
      />
    </section>
  );

  return (
    <div>
      <WorkspaceLoadWarning diagnostics={diagnostics} />
      {/* Date Range Filter Bar */}
      {dateFilterControl}
      <div className="fleet-refresh-controls">
        <button type="button" className="btn-secondary" onClick={refreshData}>Refresh Data</button>
      </div>
      <SessionOverlapWarning inspection={inspection} onSelectTicket={onSelectTicket} />

      {loading ? (
        <div className="loading-container">
          <span>Loading telemetry metrics...</span>
        </div>
      ) : errorMessage ? (
        <div className="error-banner">
          <span>{errorMessage}</span>
          <button type="button" className="btn-primary" onClick={refreshData}>
            Retry Connection
          </button>
        </div>
      ) : !summary || summary.total_tickets === 0 ? (
        <EmptyState
          title="Zero Telemetry Workspaces Available"
          message="No workspaces with recorded session metrics were discovered. Ensure tickets contain agentic_metrics.json ledgers."
        />
      ) : (
        <>
          {/* KPI Cards Row */}
          <section className="kpi-row" aria-label="Fleet Key Performance Indicators">
            <CostKPICard
              title="Total Cost"
              amount={summary.total_cost_usd}
              summary={summary.cost_summary}
              invocations={summary.total_invocations}
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
              secondary={`${summary.total_files_changed} files touched; ${durationDescription(summary.provisional_seconds)}`}
              tooltip="Cumulative active agent execution time across all tickets (cumulative man-hours)."
            />
          </section>

          <section className="distribution-row" aria-label="Skill Adoption">
            <div className="kpi-card">
              <h2 className="kpi-title">Skill Adoption</h2>
              <div className="kpi-value mono-number">{summary.skill_adoption === undefined ? 'Not recorded' : `${skillAdoption.length} recorded skills`}</div>
              <p>{summary.skill_adoption === undefined ? 'Skill metrics not recorded for a complete population.' : `${skillAdoption.reduce((total, skill) => total + skill.total_invocations, 0)} activations`}</p>
              {skillAdoption.length === 0 ? summary.skill_adoption === undefined ? null : <p>No recorded skills</p> : (
                <ol aria-label="Top skills">
                  {skillAdoption.slice(0, 3).map((skill) => (
                    <li key={skill.name}>{skill.name}: {skill.total_invocations} invocations</li>
                  ))}
                </ol>
              )}
            </div>
          </section>

          {/* Fleet Anomaly and Reliability Profile Card */}
          <section aria-label="Fleet Anomaly Profile">
            <FleetAnomalyCard
              summary={summary.anomaly_summary}
              selectedTickets={summary.total_tickets}
              onViewTicket={onSelectTicket}
            />
          </section>

          {/* Distribution Cards Row */}
          <section className="distribution-row" aria-label="Cost Distributions">
            <DistributionCard
              title="Role Cost Distribution"
              items={summary.role_costs}
              costSummaries={summary.role_cost_summaries}
              totalCostUSD={summary.total_cost_usd}
              tooltip="Breakdown of cumulative LLM inference spend categorized by agent persona role (orchestrator, coder, tester, reviewer, planner)"
            />
            <DistributionCard
              title="Model Cost Distribution"
              items={summary.model_costs}
              costSummaries={summary.model_cost_summaries}
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
