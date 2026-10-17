import React, { useEffect, useState, useMemo, useCallback } from 'react';
import type { AgenticMetricsLedger, TimeGranularity, TimeWindowPreset } from '../domain/types';
import { fetchWorkspaceLedgers } from '../api/client';
import { aggregateMetricsOverTime } from '../domain/progression';
import {
  formatCurrencyUSD,
  formatPercentage,
  formatTokensK,
  HEALTHY_CEILING_TOKENS,
} from '../domain/formatters';
import { KPICard } from '../components/KPICard';
import { TrendChart } from '../components/TrendChart';
import { EmptyState } from '../components/EmptyState';

export interface MetricProgressionProps {
  readonly onMockChange?: (isMock: boolean) => void;
}

export const MetricProgression: React.FC<MetricProgressionProps> = ({ onMockChange }) => {
  const [ledgers, setLedgers] = useState<readonly AgenticMetricsLedger[]>([]);
  const [granularity, setGranularity] = useState<TimeGranularity>('daily');
  const [preset, setPreset] = useState<TimeWindowPreset>('all');
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
      const msg = err instanceof Error ? err.message : 'Failed to fetch telemetry ledgers';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [onMockChange]);

  useEffect(() => {
    loadLedgers();
  }, [loadLedgers]);

  // Compute date boundaries based on preset and ledger dates
  const { startDate, endDate } = useMemo(() => {
    if (preset === 'all' || ledgers.length === 0) {
      return { startDate: null, endDate: null };
    }

    // Find the latest timestamp across ledgers as the reference anchor
    let latestMs = 0;
    for (const l of ledgers) {
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
    const endStr = refDate.toISOString().slice(0, 10);

    let days = 7;
    if (preset === '30d') days = 30;
    if (preset === '90d') days = 90;

    const startMs = refDate.getTime() - days * 86_400_000;
    const startStr = new Date(startMs).toISOString().slice(0, 10);

    return { startDate: startStr, endDate: endStr };
  }, [preset, ledgers]);

  // Compute progression buckets using pure domain algorithm
  const buckets = useMemo(() => {
    return aggregateMetricsOverTime(ledgers, granularity, startDate, endDate);
  }, [ledgers, granularity, startDate, endDate]);

  // Active window rollup KPIs
  const summary = useMemo(() => {
    let spend = 0;
    let invocations = 0;
    let inputTokens = 0;
    let outputTokens = 0;
    let cacheTokens = 0;
    let linesAdded = 0;
    let linesDeleted = 0;

    for (const b of buckets) {
      spend += b.cost_usd;
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
      spend,
      invocations,
      totalToks,
      cacheHitPct,
      netLines: linesAdded - linesDeleted,
    };
  }, [buckets]);

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner" />
        <span className="loading-text">Aggregating longitudinal telemetry progression...</span>
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

  if (ledgers.length === 0) {
    return (
      <EmptyState
        title="Zero Progression Telemetry Available"
        message="No workspace telemetry files were found. Run an agentic workflow to populate longitudinal progression charts."
      />
    );
  }

  const chartData = buckets.map((b) => ({
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
    lines_deleted: b.lines_deleted,
    net_lines_added: b.net_lines_added,
    peak_context_tokens: b.peak_context_tokens,
  }));

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
          </div>
        </div>

        <div className="controls-right">
          <span className="control-label">Granularity:</span>
          <div className="granularity-toggle-group">
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
            <button
              type="button"
              className={`granularity-btn ${granularity === 'cumulative' ? 'active' : ''}`}
              onClick={() => setGranularity('cumulative')}
            >
              Cumulative
            </button>
          </div>
        </div>
      </div>

      {/* Active Window KPI Summary Banner */}
      <div className="kpi-grid">
        <KPICard
          title="Period Spend"
          value={formatCurrencyUSD(summary.spend)}
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
            title={
              granularity === 'cumulative'
                ? 'Cumulative Financial Trajectory ($ USD)'
                : 'Financial Spend & Cumulative Trajectory ($ USD)'
            }
            height={260}
            data={chartData}
            valueFormatter={(v) => formatCurrencyUSD(v)}
            series={
              granularity === 'cumulative'
                ? [
                    {
                      key: 'cumulative_cost_usd',
                      label: 'Cumulative Spend',
                      color: '#10b981',
                      type: 'area',
                    },
                  ]
                : [
                    {
                      key: 'cost_usd',
                      label: 'Period Spend',
                      color: '#3b82f6',
                      type: 'bar',
                    },
                    {
                      key: 'cumulative_cost_usd',
                      label: 'Cumulative Spend',
                      color: '#10b981',
                      type: 'line',
                    },
                  ]
            }
          />

          {/* Chart 2: Token Throughput */}
          <TrendChart
            title="Token Volume Distribution"
            height={260}
            data={chartData}
            valueFormatter={(v) => formatTokensK(v)}
            series={[
              {
                key: 'cache_read_tokens',
                label: 'Cache Read Tokens',
                color: '#10b981',
                type: 'bar',
              },
              {
                key: 'input_tokens',
                label: 'Input Tokens',
                color: '#3b82f6',
                type: 'bar',
              },
              {
                key: 'output_tokens',
                label: 'Output Tokens',
                color: '#8b5cf6',
                type: 'bar',
              },
            ]}
          />

          {/* Chart 3: Prompt Cache Hit Rate Trajectory */}
          <TrendChart
            title="Prompt Cache Hit Rate Trajectory (%)"
            height={260}
            data={chartData}
            yAxisMin={0}
            valueFormatter={(v) => `${v.toFixed(1)}%`}
            referenceLines={[
              {
                value: 50,
                label: 'Baseline 50%',
                color: '#f59e0b',
                strokeDasharray: '4 4',
              },
              {
                value: 80,
                label: 'Target 80%',
                color: '#10b981',
                strokeDasharray: '4 4',
              },
            ]}
            series={[
              {
                key: 'cache_hit_rate_pct',
                label: 'Cache Hit Rate %',
                color: '#06b6d4',
                type: 'line',
              },
            ]}
          />

          {/* Chart 4: Delivered Code Line Churn */}
          <TrendChart
            title="Delivered Code Line Churn & Net Growth"
            height={260}
            data={chartData}
            valueFormatter={(v) => `${v.toLocaleString()} lines`}
            series={[
              {
                key: 'lines_added',
                label: 'Lines Added (+)',
                color: '#10b981',
                type: 'bar',
              },
              {
                key: 'lines_deleted',
                label: 'Lines Deleted (-)',
                color: '#ef4444',
                type: 'bar',
              },
              {
                key: 'net_lines_added',
                label: 'Net Lines Growth',
                color: '#3b82f6',
                type: 'line',
              },
            ]}
          />

          {/* Chart 5: Peak Context Window Saturation */}
          <TrendChart
            title="Peak Context Window Saturation vs Ceiling"
            height={260}
            data={chartData}
            valueFormatter={(v) => `${formatTokensK(v)} tokens`}
            referenceLines={[
              {
                value: HEALTHY_CEILING_TOKENS,
                label: '200k Ceiling',
                color: '#ef4444',
                strokeDasharray: '5 5',
              },
            ]}
            series={[
              {
                key: 'peak_context_tokens',
                label: 'Peak Context Tokens',
                color: '#f97316',
                type: 'line',
              },
            ]}
          />
        </div>
      )}
    </div>
  );
};
