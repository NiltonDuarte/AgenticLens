import type { TicketComparisonRow } from './types';

export type ScatterMetricKey =
  | 'cost_usd'
  | 'net_lines_added'
  | 'active_duration_seconds'
  | 'lead_duration_seconds'
  | 'sdlc_iterations'
  | 'cache_hit_rate_pct'
  | 'lines_added'
  | 'lines_deleted'
  | 'tool_success_rate_pct'
  | 'orchestrator_saturation_pct';

export type ScatterScale = 'linear' | 'logarithmic';

export interface ScatterMetric {
  readonly key: ScatterMetricKey;
  readonly label: string;
  readonly format: (value: number) => string;
}

export interface ScatterPoint {
  readonly row: TicketComparisonRow;
  readonly x: number;
  readonly y: number;
}

export interface ScatterBounds {
  readonly min: number;
  readonly max: number;
}

export interface ScatterPlotData {
  readonly points: readonly ScatterPoint[];
  readonly xBounds: ScatterBounds;
  readonly yBounds: ScatterBounds;
  readonly recommendsLogarithmic: boolean;
  readonly unavailableCount: number;
}

const formatNumber = (value: number): string => new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value);
const formatCurrency = (value: number): string => `$${value.toFixed(value < 1 ? 4 : 2)}`;
const formatPercent = (value: number): string => `${value.toFixed(1)}%`;
const formatDuration = (value: number): string => {
  const seconds = Math.max(0, Math.floor(value));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remaining = seconds % 60;
  return hours > 0 ? `${hours}h ${minutes}m` : minutes > 0 ? `${minutes}m ${remaining}s` : `${remaining}s`;
};

export const SCATTER_METRICS: readonly ScatterMetric[] = [
  { key: 'cost_usd', label: 'Total Cost (USD)', format: formatCurrency },
  { key: 'net_lines_added', label: 'Net Lines of Code', format: formatNumber },
  { key: 'active_duration_seconds', label: 'Active Duration', format: formatDuration },
  { key: 'lead_duration_seconds', label: 'Lead Duration', format: formatDuration },
  { key: 'sdlc_iterations', label: 'SDLC Iterations', format: formatNumber },
  { key: 'cache_hit_rate_pct', label: 'Cache Hit Rate', format: formatPercent },
  { key: 'lines_added', label: 'Lines Added', format: formatNumber },
  { key: 'lines_deleted', label: 'Lines Deleted', format: formatNumber },
  { key: 'tool_success_rate_pct', label: 'Tool Success Rate', format: formatPercent },
  { key: 'orchestrator_saturation_pct', label: 'Orchestrator Saturation', format: formatPercent },
];

export function getScatterMetricValue(row: TicketComparisonRow, key: ScatterMetricKey): number | null {
  switch (key) {
    case 'cost_usd': return row.cost_usd;
    case 'net_lines_added': return row.lines_added - row.lines_deleted;
    case 'active_duration_seconds': return row.active_duration_seconds;
    case 'lead_duration_seconds': return row.lead_duration_seconds;
    case 'sdlc_iterations': return row.sdlc_iterations;
    case 'cache_hit_rate_pct': return row.cache_hit_rate_pct;
    case 'lines_added': return row.lines_added;
    case 'lines_deleted': return row.lines_deleted;
    case 'tool_success_rate_pct': return row.tool_success_rate_pct;
    case 'orchestrator_saturation_pct': return row.root_saturation_pct;
  }
}

export function transformScatterValue(value: number, scale: ScatterScale): number | null {
  if (!Number.isFinite(value)) return null;
  if (scale === 'linear') return value;
  if (value < 0) return null;
  return Math.log10(value + 1);
}

function getBounds(values: readonly number[]): ScatterBounds {
  if (values.length === 0) return { min: 0, max: 1 };
  const min = Math.min(...values);
  const max = Math.max(...values);
  if (min === max) {
    const padding = Math.max(Math.abs(min) * 0.1, 0.5);
    return { min: min - padding, max: max + padding };
  }
  const padding = (max - min) * 0.1;
  return { min: min - padding, max: max + padding };
}

function shouldRecommendLog(values: readonly number[]): boolean {
  const nonNegative = values.filter((value) => value >= 0 && Number.isFinite(value));
  if (nonNegative.length < 2) return false;
  const min = Math.min(...nonNegative);
  const max = Math.max(...nonNegative);
  return max > 0 && (min === 0 || max / min > 50);
}

export function deriveScatterPlotData(
  rows: readonly TicketComparisonRow[],
  xMetric: ScatterMetricKey,
  yMetric: ScatterMetricKey,
  scale: ScatterScale,
  excludedTicketIds: ReadonlySet<string>,
): ScatterPlotData {
  let unavailableCount = 0;
  const activeRows: Array<{ row: TicketComparisonRow; x: number; y: number }> = [];
  for (const row of rows) {
    if (excludedTicketIds.has(row.ticket_id)) continue;
    const rawX = getScatterMetricValue(row, xMetric);
    const rawY = getScatterMetricValue(row, yMetric);
    if (rawX === null || rawY === null || !Number.isFinite(rawX) || !Number.isFinite(rawY)) { unavailableCount += 1; continue; }
    activeRows.push({ row, x: rawX, y: rawY });
  }

  const recommendsLogarithmic = shouldRecommendLog(activeRows.map((point) => point.x)) ||
    shouldRecommendLog(activeRows.map((point) => point.y));
  const points: ScatterPoint[] = [];
  for (const point of activeRows) {
    const x = transformScatterValue(point.x, scale);
    const y = transformScatterValue(point.y, scale);
    if (x !== null && y !== null) points.push({ row: point.row, x, y });
  }
  return {
    points,
    xBounds: getBounds(points.map((point) => point.x)),
    yBounds: getBounds(points.map((point) => point.y)),
    recommendsLogarithmic,
    unavailableCount,
  };
}
