import { completeCost, costSummaryForAmount, mergeCostSummaries } from '../domain/cost';
import React, { useEffect, useState, useMemo, useCallback, useRef } from 'react';
import type { AgenticMetricsLedger, TimeGranularity, TimeWindowPreset, WorkspaceDiagnostic } from '../domain/types';
import { fetchWorkspaceLedgers } from '../api/client';
import { WorkspaceLoadWarning } from '../components/WorkspaceLoadWarning';
import { aggregateMetricsOverTime } from '../domain/progression';
import {
  combineDistinctPrefixes,
  extractTicketPrefix,
  filterLedgersByPrefixes,
  filterLedgersByTicketTypes,
  formatCurrencyUSD,
  formatLinesCompact,
  formatPercentage,
  formatTokensK,
  HEALTHY_CEILING_TOKENS,
} from '../domain/formatters';
import { KPICard } from '../components/KPICard';
import { CostKPICard } from '../components/CostKPICard';
import { TrendChart } from '../components/TrendChart';
import { EmptyState } from '../components/EmptyState';
import { CHART_COLORS } from '../utils/theme';
import { MultiSelectDropdown, type MultiSelectOption } from '../components/MultiSelectDropdown';

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

export interface MetricProgressionProps {
  readonly configuredPrefixes?: readonly string[] | undefined;
}

export const MetricProgression: React.FC<MetricProgressionProps> = ({
  configuredPrefixes,
}) => {
  const [ledgers, setLedgers] = useState<readonly AgenticMetricsLedger[]>([]);
  const [diagnostics, setDiagnostics] = useState<readonly WorkspaceDiagnostic[]>([]);
  const [knownPrefixes, setKnownPrefixes] = useState<readonly string[]>([]);
  const [selectedTicketTypes, setSelectedTicketTypes] = useState<readonly string[]>(['all']);
  const [selectedPrefixes, setSelectedPrefixes] = useState<readonly string[]>(['all']);
  const [granularity, setGranularity] = useState<TimeGranularity>('daily');
  const [preset, setPreset] = useState<TimeWindowPreset>('all');
  const [customStart, setCustomStart] = useState('');
  const [customEnd, setCustomEnd] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const hasLoadedRef = useRef(false);

  const loadLedgers = useCallback(async () => {
    if (!hasLoadedRef.current) {
      setLoading(true);
    }
    setError(null);
    try {
      const typesParam =
        selectedTicketTypes.length === 0 ||
        (selectedTicketTypes.length === 1 && selectedTicketTypes[0] === 'all')
          ? undefined
          : selectedTicketTypes;
      const prefixesParam =
        selectedPrefixes.length === 0 ||
        (selectedPrefixes.length === 1 && selectedPrefixes[0] === 'all')
          ? undefined
          : selectedPrefixes;

      const res =
        typesParam || prefixesParam
          ? await fetchWorkspaceLedgers(preset, undefined, undefined, typesParam, prefixesParam)
          : await fetchWorkspaceLedgers('all');
      setLedgers(res.data);
      setDiagnostics(res.diagnostics ?? []);
      hasLoadedRef.current = true;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to fetch telemetry ledgers';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [selectedTicketTypes, selectedPrefixes, preset]);

  useEffect(() => {
    loadLedgers();
  }, [loadLedgers]);

  // Keep track of distinct workspace prefixes discovered across ledgers
  useEffect(() => {
    if (ledgers.length > 0) {
      setKnownPrefixes((prev) => {
        const nextSet = new Set(prev);
        for (const l of ledgers) {
          if (l.ticket_id) {
            const p = extractTicketPrefix(l.ticket_id);
            if (p) nextSet.add(p);
          }
        }
        return Array.from(nextSet).sort();
      });
    }
  }, [ledgers]);

  const prefixOptions: readonly MultiSelectOption[] = useMemo(() => {
    const combined = combineDistinctPrefixes(knownPrefixes, configuredPrefixes);
    return combined.map((p) => ({ value: p, label: p }));
  }, [knownPrefixes, configuredPrefixes]);

  // Apply pure domain filtering across selected types and prefixes
  const filteredLedgers = useMemo(() => {
    let list = ledgers;
    list = filterLedgersByTicketTypes(list, selectedTicketTypes);
    list = filterLedgersByPrefixes(list, selectedPrefixes);
    return list;
  }, [ledgers, selectedTicketTypes, selectedPrefixes]);

  // Compute date boundaries based on preset and filtered ledger dates
  const { startDate, endDate } = useMemo(() => {
    if (preset === 'all' || filteredLedgers.length === 0) {
      return { startDate: null, endDate: null };
    }

    if (preset === 'custom') {
      const s = customStart.trim() || null;
      const e = customEnd.trim() || null;
      return { startDate: s, endDate: e };
    }

    // Find the latest timestamp across ledgers as the reference anchor
    let latestMs = 0;
    for (const l of filteredLedgers) {
      const ts = l.completed_at ?? l.started_at ?? l.updated_at;
      const ms = ts ? Date.parse(ts) : NaN;
      if (!Number.isNaN(ms) && ms > latestMs) {
        latestMs = ms;
      }
    }

    if (latestMs === 0) {
      return { startDate: null, endDate: null };
    }

    const refDate = new Date(latestMs);

    if (preset === '1d') {
      const startMs = refDate.getTime() - 24 * 3600_000;
      return {
        startDate: new Date(startMs).toISOString(),
        endDate: refDate.toISOString(),
      };
    }

    const endStr = refDate.toISOString().slice(0, 10);

    let days = 7;
    if (preset === '30d') days = 30;
    if (preset === '90d') days = 90;

    const startMs = refDate.getTime() - days * 86_400_000;
    const startStr = new Date(startMs).toISOString().slice(0, 10);

    return { startDate: startStr, endDate: endStr };
  }, [preset, filteredLedgers, customStart, customEnd]);

  const isCustomRangeInverted = useMemo(() => {
    if (preset !== 'custom' || !customStart || !customEnd) {
      return false;
    }
    const s = Date.parse(customStart);
    const e = Date.parse(customEnd);
    return !Number.isNaN(s) && !Number.isNaN(e) && s > e;
  }, [preset, customStart, customEnd]);

  // Compute progression buckets using pure domain algorithm
  const buckets = useMemo(() => {
    return aggregateMetricsOverTime(filteredLedgers, granularity, startDate, endDate);
  }, [filteredLedgers, granularity, startDate, endDate]);

  // Active window rollup KPIs
  const summary = useMemo(() => {
    const costSummary = mergeCostSummaries(buckets.map((b) => costSummaryForAmount(b.cost_usd, b.cost_summary, b.invocations)));
    let invocations = 0;
    let inputTokens = 0;
    let outputTokens = 0;
    let cacheTokens = 0;
    let linesAdded = 0;
    let linesDeleted = 0;

    for (const b of buckets) {

      invocations += b.invocations;
      inputTokens += b.input_tokens;
      outputTokens += b.output_tokens;
      cacheTokens += b.cache_read_tokens;
      linesAdded += b.lines_added;
      linesDeleted += b.lines_deleted;
    }

    const totalToks = inputTokens + outputTokens + cacheTokens;
    const cachedOrInp = inputTokens + cacheTokens;
    const cacheHitPct =
      cachedOrInp > 0 ? Math.round((cacheTokens / cachedOrInp) * 1000) / 10 : 0.0;

    return {
      spend: completeCost(costSummary),
      costSummary,
      invocations,
      totalToks,
      cacheHitPct,
      netLines: linesAdded - linesDeleted,
    };
  }, [buckets]);

  const chartData = useMemo(
    () =>
      buckets.map((b) => ({
        bucket_date: b.bucket_date,
        cost_usd: b.cost_usd,
        cumulative_cost_usd: b.cumulative_cost_usd,
        invocations: b.invocations,
        cumulative_invocations: b.cumulative_invocations,
        input_tokens: b.input_tokens,
        output_tokens: b.output_tokens,
        cache_read_tokens: b.cache_read_tokens,
        cache_hit_rate_pct: b.cache_hit_rate_pct,
        lines_added: b.lines_added,
        lines_deleted: -Math.abs(b.lines_deleted),
        net_lines_added: b.net_lines_added,
        cumulative_net_lines_added: b.cumulative_net_lines_added,
        peak_context_tokens: b.peak_context_tokens,
      })),
    [buckets]
  );

  return (
    <div className="progression-container">
      {/* Top Filter and Controls Bar */}
      <div className="progression-controls-bar">
        <div className="controls-left">
          <span className="control-label">Window:</span>
          <div className="preset-toggle-group">
            <button
              type="button"
              className={`preset-btn ${preset === 'all' ? 'active' : ''}`}
              onClick={() => setPreset('all')}
            >
              Lifetime
            </button>
            <button
              type="button"
              className={`preset-btn ${preset === '1d' ? 'active' : ''}`}
              onClick={() => setPreset('1d')}
            >
              24 Hours
            </button>
            <button
              type="button"
              className={`preset-btn ${preset === '7d' ? 'active' : ''}`}
              onClick={() => setPreset('7d')}
            >
              7 Days
            </button>
            <button
              type="button"
              className={`preset-btn ${preset === '30d' ? 'active' : ''}`}
              onClick={() => setPreset('30d')}
            >
              30 Days
            </button>
            <button
              type="button"
              className={`preset-btn ${preset === '90d' ? 'active' : ''}`}
              onClick={() => setPreset('90d')}
            >
              90 Days
            </button>
            <button
              type="button"
              className={`preset-btn ${preset === 'custom' ? 'active' : ''}`}
              onClick={() => setPreset('custom')}
            >
              Custom
            </button>
          </div>
        </div>

        <div className="controls-filters">
          <MultiSelectDropdown
            id="prog-ticket-type"
            label="Ticket Type"
            options={TICKET_TYPE_OPTIONS}
            selectedValues={selectedTicketTypes}
            onChange={setSelectedTicketTypes}
          />
          <MultiSelectDropdown
            id="prog-prefix"
            label="Prefix"
            options={prefixOptions}
            selectedValues={selectedPrefixes}
            onChange={setSelectedPrefixes}
          />
        </div>

        <div className="controls-right">
          <span className="control-label">Granularity:</span>
          <div className="granularity-toggle-group">
            <button
              type="button"
              className={`granularity-btn ${granularity === 'hourly' ? 'active' : ''}`}
              onClick={() => setGranularity('hourly')}
            >
              Hourly
            </button>
            <button
              type="button"
              className={`granularity-btn ${granularity === 'daily' ? 'active' : ''}`}
              onClick={() => setGranularity('daily')}
            >
              Daily
            </button>
            <button
              type="button"
              className={`granularity-btn ${granularity === 'weekly' ? 'active' : ''}`}
              onClick={() => setGranularity('weekly')}
            >
              Weekly
            </button>
          </div>
        </div>
      </div>

      {/* Custom Datetime Input Controls */}
      {preset === 'custom' && (
        <div className="custom-datetime-bar">
          <label className="datetime-input-group">
            <span className="control-label">From:</span>
            <input
              type="datetime-local"
              className="datetime-input"
              value={customStart}
              onChange={(e) => setCustomStart(e.target.value)}
              placeholder="YYYY-MM-DDTHH:mm"
              aria-label="Start Date and Time"
            />
          </label>
          <label className="datetime-input-group">
            <span className="control-label">To:</span>
            <input
              type="datetime-local"
              className="datetime-input"
              value={customEnd}
              onChange={(e) => setCustomEnd(e.target.value)}
              placeholder="YYYY-MM-DDTHH:mm"
              aria-label="End Date and Time"
            />
          </label>
          {isCustomRangeInverted && (
            <span className="datetime-warning">
              Start datetime cannot be after End datetime.
            </span>
          )}
        </div>
      )}

      <WorkspaceLoadWarning diagnostics={diagnostics} />
      {loading && !hasLoadedRef.current ? (
        <div className="loading-container">
          <div className="loading-spinner" />
          <span className="loading-text">Aggregating longitudinal telemetry progression...</span>
        </div>
      ) : error ? (
        <div className="error-banner">
          <span className="error-text">{error}</span>
          <button type="button" className="btn-secondary" onClick={loadLedgers}>
            Retry Connection
          </button>
        </div>
      ) : ledgers.length === 0 || filteredLedgers.length === 0 ? (
        <EmptyState
          title="Zero Progression Telemetry Available"
          message="No workspace telemetry files were found. Run an agentic workflow to populate longitudinal progression charts."
        />
      ) : (
        <>
          {/* Active Window KPI Summary Banner */}
          <div className="kpi-grid">
        <CostKPICard
          title="Period Spend"
          amount={summary.spend}
          summary={summary.costSummary}
          invocations={summary.invocations}
          secondary={`${buckets.length} time buckets`}
          tooltip="Total USD spend across matching date range"
        />
        <KPICard
          title="Total Invocations"
          value={summary.invocations.toLocaleString()}
          secondary="Agent executions"
          tooltip="Total invocation fact records in selected window"
        />
        <KPICard
          title="Total Token Throughput"
          value={formatTokensK(summary.totalToks)}
          secondary="Input + Output + Cache"
          tooltip="Total combined tokens processed"
        />
        <KPICard
          title="Window Cache Hit Rate"
          value={formatPercentage(summary.cacheHitPct)}
          secondary="Prompt cache efficiency"
          tooltip="Overall cache read percentage in selected window"
        />
        <KPICard
          title="Net Code Delivered"
          value={`${summary.netLines >= 0 ? '+' : ''}${summary.netLines.toLocaleString()}`}
          secondary="Lines added minus deleted"
          tooltip="Delivered code growth in window"
        />
      </div>

      {buckets.length === 0 ? (
        <div className="empty-progression-notice">
          No telemetry invocations or deliverable code records found within the selected time window.
        </div>
      ) : (
        /* Trend Charts Grid */
        <div className="progression-charts-grid">
          {/* Chart 1: Financial & Invocations */}
          <TrendChart
            title="Financial Spend & Cumulative Trajectory ($ USD)"
            height={260}
            data={chartData}
            granularity={granularity}
            tooltip="Historical model inference spend per time period alongside cumulative trajectory in US dollars."
            valueFormatter={(v) => formatCurrencyUSD(v)}
            series={[
              {
                key: 'cost_usd',
                label: 'Period Spend',
                color: CHART_COLORS.spendPeriod,
                type: 'bar',
              },
              {
                key: 'cumulative_cost_usd',
                label: 'Cumulative Spend',
                color: CHART_COLORS.spendCumulative,
                type: 'line',
              },
            ]}
          />

          {/* Chart 2: Token Throughput */}
          <TrendChart
            title="Token Volume Distribution"
            height={260}
            data={chartData}
            granularity={granularity}
            tooltip="Distribution of token throughput across input prompts, output completions, and key-value prompt cache reads."
            valueFormatter={(v) => formatTokensK(v)}
            series={[
              {
                key: 'cache_read_tokens',
                label: 'Cache Read Tokens',
                color: CHART_COLORS.cacheRead,
                type: 'bar',
              },
              {
                key: 'input_tokens',
                label: 'Input Tokens',
                color: CHART_COLORS.inputTokens,
                type: 'bar',
              },
              {
                key: 'output_tokens',
                label: 'Output Tokens',
                color: CHART_COLORS.outputTokens,
                type: 'bar',
              },
            ]}
          />

          {/* Chart 3: Prompt Cache Hit Rate Trajectory */}
          <TrendChart
            title="Prompt Cache Hit Rate Trajectory (%)"
            height={260}
            data={chartData}
            granularity={granularity}
            tooltip="Percentage of prompt tokens served from prompt cache per bucket against baseline (50%) and target (80%) thresholds."
            yAxisMin={0}
            valueFormatter={(v) => `${v.toFixed(1)}%`}
            referenceLines={[
              {
                value: 50,
                label: 'Baseline 50%',
                color: CHART_COLORS.cacheBaseline,
                strokeDasharray: '4 4',
              },
              {
                value: 80,
                label: 'Target 80%',
                color: CHART_COLORS.cacheTarget,
                strokeDasharray: '4 4',
              },
            ]}
            series={[
              {
                key: 'cache_hit_rate_pct',
                label: 'Cache Hit Rate %',
                color: CHART_COLORS.cacheHitRate,
                type: 'line',
              },
            ]}
          />

          {/* Chart 4: Delivered Code Line Churn */}
          <TrendChart
            title="Delivered Code Line Churn & Net Growth"
            height={260}
            data={chartData}
            granularity={granularity}
            tooltip="Volume of production and test code lines added, deleted, and net code growth delivered per time bucket."
            valueFormatter={(v) => `${v > 0 ? '+' : ''}${v.toLocaleString()} lines`}
            yAxisFormatter={(v) => (v > 0 ? `+${formatLinesCompact(v)}` : formatLinesCompact(v))}
            integerTicks={true}
            divergingZeroBaseline={true}
            series={[
              {
                key: 'lines_added',
                label: 'Lines Added (+)',
                color: CHART_COLORS.linesAdded,
                type: 'bar',
              },
              {
                key: 'lines_deleted',
                label: 'Lines Deleted (-)',
                color: CHART_COLORS.linesDeleted,
                type: 'bar',
              },
              {
                key: 'cumulative_net_lines_added',
                label: 'Cumulative Net Growth',
                color: CHART_COLORS.linesNet,
                type: 'line',
              },
            ]}
          />

          {/* Chart 5: Peak Context Window Saturation */}
          <TrendChart
            title="Peak Context Window Saturation vs Ceiling"
            height={260}
            data={chartData}
            granularity={granularity}
            tooltip="Peak active context tokens per time bucket measured against the standard 200,000 token healthy context ceiling."
            valueFormatter={(v) => `${formatTokensK(v)} tokens`}
            referenceLines={[
              {
                value: HEALTHY_CEILING_TOKENS,
                label: '200k Ceiling',
                color: CHART_COLORS.ceilingMarker,
                strokeDasharray: '5 5',
              },
            ]}
            series={[
              {
                key: 'peak_context_tokens',
                label: 'Peak Context Tokens',
                color: CHART_COLORS.peakContext,
                type: 'line',
              },
            ]}
          />
          </div>
        )}
      </>
    )}
  </div>
  );
};
