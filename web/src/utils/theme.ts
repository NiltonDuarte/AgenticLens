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

export const DYNAMIC_PALETTE: readonly string[] = [
  '#ec4899',
  '#14b8a6',
  '#f43f5e',
  '#84cc16',
  '#a855f7',
  '#0ea5e9',
  '#eab308',
  '#6366f1',
  '#d946ef',
  '#10b981',
  '#f97316',
  '#06b6d4',
] as const;

export function hashStringToColor(str: string): string {
  const normalized = str.trim().toLowerCase();
  if (normalized === '') {
    return THEME_COLORS.slate;
  }
  let hash = 0;
  for (let i = 0; i < normalized.length; i++) {
    hash = ((hash << 5) - hash) + normalized.charCodeAt(i);
    hash |= 0;
  }
  const index = Math.abs(hash) % DYNAMIC_PALETTE.length;
  return DYNAMIC_PALETTE[index] ?? THEME_COLORS.slate;
}

export function getRoleColor(role: string): string {
  return hashStringToColor(role);
}

export function getModelColor(model: string): string {
  return hashStringToColor(model);
}
