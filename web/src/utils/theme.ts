/**
 * Shared application theme colors, status mappings, and chart tokens.
 * Aligns with CSS variables defined in index.css.
 */

export const THEME_COLORS = {
  green: '#10b981',
  blue: '#3b82f6',
  purple: '#8b5cf6',
  yellow: '#f59e0b',
  orange: '#f97316',
  red: '#ef4444',
  cyan: '#06b6d4',
  slate: '#64748b',
} as const;

export const STATUS_COLORS = {
  HEALTHY: '#10b981',
  ELEVATED: '#f59e0b',
  BLOATED: '#f97316',
  CRITICAL: '#ef4444',
  UNKNOWN: '#64748b',
} as const;

export const CHART_COLORS = {
  spendPeriod: THEME_COLORS.blue,
  spendCumulative: THEME_COLORS.green,
  cacheRead: THEME_COLORS.green,
  inputTokens: THEME_COLORS.blue,
  outputTokens: THEME_COLORS.purple,
  cacheHitRate: THEME_COLORS.cyan,
  cacheBaseline: THEME_COLORS.yellow,
  cacheTarget: THEME_COLORS.green,
  linesAdded: THEME_COLORS.green,
  linesDeleted: THEME_COLORS.red,
  linesNet: THEME_COLORS.blue,
  peakContext: THEME_COLORS.orange,
  ceilingMarker: THEME_COLORS.red,
} as const;

export function getHealthStatusColor(status: string): string {
  switch (status.toUpperCase()) {
    case 'HEALTHY':
      return STATUS_COLORS.HEALTHY;
    case 'ELEVATED':
      return STATUS_COLORS.ELEVATED;
    case 'BLOATED':
      return STATUS_COLORS.BLOATED;
    case 'CRITICAL':
      return STATUS_COLORS.CRITICAL;
    default:
      return STATUS_COLORS.UNKNOWN;
  }
}
