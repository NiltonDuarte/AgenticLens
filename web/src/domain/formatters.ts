import { ledgerMetricCoverage } from './types';
import { compareCosts, costSummaryForAmount } from './cost';
import type { CostEvidence, CostSummary } from './types';
import { skillNames } from './aggregations';
import type {
  AgenticMetricsLedger,
  ComparisonSortField,
  ComparisonSortOrder,
  TicketComparisonRow,
  TimeGranularity,
} from './types';

import { compareConfiguredTicketIds, compareText, ticketNumber, lowercaseTicketText, type TicketRecognitionPolicy } from './tickets';

export function formatSkillStatus(success: boolean): 'Success' | 'Failed' {
  return success ? 'Success' : 'Failed';
}

/** Empty is the canonical missing-model identity; observed names remain verbatim. */
export function formatModelLabel(model: string): string {
  return model === '' ? 'Unknown' : model;
}

// Preserve persona spelling while making unavailable source roles visible.
export function formatRoleLabel(role: string): string {
  return role.trim() || 'Unknown';
}

export function truncateMiddle(
  str: string,
  maxLeading = 8,
  maxTrailing = 6
): string {
  const leading = Number.isFinite(maxLeading) && maxLeading >= 0
    ? Math.floor(maxLeading)
    : 8;
  const trailing = Number.isFinite(maxTrailing) && maxTrailing >= 0
    ? Math.floor(maxTrailing)
    : 6;

  if (str.length <= leading + trailing) {
    return str;
  }
  return `${str.slice(0, leading)}...${trailing === 0 ? '' : str.slice(-trailing)}`;
}

export function formatLocalTimeWithMs(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso || '';
  return `${new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23' }).format(date)}.${String(date.getMilliseconds()).padStart(3, '0')}`;
}

/** Returns a source instant only when its RFC 3339 calendar and clock fields are valid. */
export function getHistoricalTimestampDateTime(iso: string | null | undefined): string | undefined {
  if (!iso) return undefined;
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/i.exec(iso);
  if (!match) return undefined;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const leapYear = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const daysInMonth = [31, leapYear ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  const maxDay = daysInMonth[month - 1];
  if (maxDay === undefined || day < 1 || day > maxDay ||
    Number(match[4]) > 23 || Number(match[5]) > 59 || Number(match[6]) > 59 ||
    Number.isNaN(Date.parse(iso))) return undefined;
  return iso;
}

/** Retains malformed historical source text instead of normalizing it in a tooltip. */
export function formatHistoricalDateTimeTooltip(iso: string): string {
  return getHistoricalTimestampDateTime(iso) ? formatLocalDateTimeTooltip(iso) : iso;
}

/** Formats historical instants in the viewer timezone with visible calendar context. */
export function formatLocalDateTime(iso: string | null | undefined): string {
  if (!iso) return 'N/A';
  if (!getHistoricalTimestampDateTime(iso)) return iso;
  const date = new Date(iso);
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
  }).format(date);
}

export function getClientTimezoneBadge(): string {
  try {
    return `TZ: ${Intl.DateTimeFormat().resolvedOptions().timeZone || 'Local'}`;
  } catch {
    return 'TZ: Local';
  }
}

export function formatLocalDateTimeTooltip(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso || '';
  const dateTime = new Intl.DateTimeFormat(undefined, { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23' }).format(date);
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? '+' : '-';
  const offset = `${sign}${String(Math.floor(Math.abs(offsetMinutes) / 60)).padStart(2, '0')}:${String(Math.abs(offsetMinutes) % 60).padStart(2, '0')}`;
  return `${dateTime}.${String(date.getMilliseconds()).padStart(3, '0')} [UTC${offset}]`;
}

export function formatTokensK(toks: number): string {
  if (!Number.isFinite(toks) || toks < 0) {
    return '0';
  }
  const rounded = Math.floor(toks);
  if (rounded >= 1_000_000) {
    const val = rounded / 1_000_000;
    if (Number.isInteger(val)) {
      return `${val.toFixed(0)}M`;
    }
    return `${val.toFixed(1)}M`;
  }
  if (rounded >= 1_000) {
    const val = rounded / 1_000;
    if (Number.isInteger(val)) {
      return `${val.toFixed(0)}k`;
    }
    return `${val.toFixed(1)}k`;
  }
  return rounded.toString();
}

export function formatCurrencyUSD(cost: number | null | undefined): string {
  if (cost == null || !Number.isFinite(cost)) return 'Unavailable';
  if (cost <= 0) return '$0.00';
  if (cost < 1) {
    return `$${cost.toFixed(4)}`;
  }
  return `$${cost.toFixed(2)}`;
}

export function formatCostUSD(amount: number | null | undefined, summary?: CostSummary): string {
  const cost = costSummaryForAmount(amount, summary);
  if (cost.state === 'empty') return 'No invocation data';
  if (cost.state === 'complete') return formatCurrencyUSD(amount);
  const coverage = `${cost.fully_observed}/${cost.total} observed, ${cost.partial} partial, ${cost.unavailable} unavailable`;
  return cost.known_subtotal_usd === null
    ? `Unavailable (${coverage})`
    : `${formatCurrencyUSD(cost.known_subtotal_usd)} known subtotal (${coverage})`;
}

/** References carry fact evidence but do not establish invocation coverage. */
export function formatFactCostUSD(amount: number | null | undefined, evidence?: CostEvidence): string {
  if (evidence?.state === 'partial') return `${formatCurrencyUSD(evidence.known_subtotal_usd)} known subtotal (partial)`;
  if (evidence?.state === 'unavailable') return 'Unavailable';
  return formatCurrencyUSD(amount);
}

export function formatPercentage(pct: number | null): string {
  if (pct === null) return 'Unavailable';
  if (!Number.isFinite(pct) || pct <= 0) {
    return '0.0%';
  }
  return `${pct.toFixed(1)}%`;
}

export function formatDurationSeconds(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return '0m 0s';
  }
  const total = Math.floor(seconds);
  const hours = Math.floor(total / 3600);
  const rem = total % 3600;
  const mins = Math.floor(rem / 60);
  const remSecs = rem % 60;

  if (hours > 0) {
    return `${hours}h ${mins}m`;
  }
  return `${mins}m ${remSecs}s`;
}

export function computeLeadTimeSeconds(
  startedAt: string | null,
  completedAt: string | null
): number | null {
  if (!startedAt || !completedAt) {
    return null;
  }
  const startMs = Date.parse(startedAt);
  const compMs = Date.parse(completedAt);
  if (Number.isNaN(startMs) || Number.isNaN(compMs)) {
    return null;
  }
  const diffSec = Math.floor((compMs - startMs) / 1000);
  return diffSec >= 0 ? diffSec : 0;
}

export function formatLeadTime(seconds: number | null): string {
  if (seconds === null || !Number.isFinite(seconds)) {
    return 'In Progress';
  }
  const sec = Math.floor(seconds);
  if (sec <= 0) {
    return '0s';
  }
  if (sec >= 86400) {
    const days = Math.floor(sec / 86400);
    const hours = Math.floor((sec % 86400) / 3600);
    return `${days}d ${hours}h`;
  }
  if (sec >= 3600) {
    const hours = Math.floor(sec / 3600);
    const mins = Math.floor((sec % 3600) / 60);
    return `${hours}h ${mins}m`;
  }
  const mins = Math.floor(sec / 60);
  const remSecs = sec % 60;
  return `${mins}m ${remSecs}s`;
}

export function formatIsoDateTimeShort(isoString: string | null | undefined): string {
  if (!isoString || isoString.length < 10) {
    return 'N/A';
  }
  const datePart = isoString.slice(0, 10);
  if (isoString.length >= 16) {
    const timePart = isoString.slice(11, 16);
    return `${datePart} ${timePart}`;
  }
  return datePart;
}

export function compareTicketIds(a: string, b: string, policy?: TicketRecognitionPolicy): number {
  return compareConfiguredTicketIds(a, b, policy);
}

export interface TicketDateComparable {
  readonly ticket_id?: string | null | undefined;
  readonly completed_at?: string | null | undefined;
  readonly updated_at?: string | undefined;
  readonly started_at?: string | null | undefined;
}

export function getEffectiveCompletionDateMs(item: TicketDateComparable): number | null {
  const candidate = item.completed_at;
  if (candidate && typeof candidate === 'string' && candidate.trim().length > 0) {
    const ms = Date.parse(candidate.trim());
    if (!Number.isNaN(ms)) {
      return ms;
    }
  }
  return null;
}

export function compareTicketCompletionDate(
  a: TicketDateComparable,
  b: TicketDateComparable,
  policy?: TicketRecognitionPolicy
): number {
  const dateA = getEffectiveCompletionDateMs(a);
  const dateB = getEffectiveCompletionDateMs(b);

  if (dateA !== null && dateB === null) {
    return -1;
  }
  if (dateA === null && dateB !== null) {
    return 1;
  }
  if (dateA !== null && dateB !== null && dateA !== dateB) {
    return dateB - dateA;
  }
  return compareTicketIds(a.ticket_id ?? '', b.ticket_id ?? '', policy);
}

export function deriveTicketComparisonRows(
  ledgers: readonly AgenticMetricsLedger[],
  sortField?: ComparisonSortField,
  sortOrder: ComparisonSortOrder = 'desc',
  policy?: TicketRecognitionPolicy
): TicketComparisonRow[] {
  const rows: TicketComparisonRow[] = ledgers.map((ledger) => {
    const tid = ledger.ticket_id && ledger.ticket_id.length > 0 ? ledger.ticket_id : 'unknown';
    const eff = ledger.workflow_efficiency;
    const tot = ledger.totals;
    const cwh = eff.context_window_health;
    const leadSecs = computeLeadTimeSeconds(ledger.started_at, ledger.completed_at);
    const title = ledger.ticket_title && ledger.ticket_title.length > 0 ? ledger.ticket_title : 'N/A';

    return {
      ticket_id: tid,
      ...(ledger.workspace_path ? { workspace_path: ledger.workspace_path } : {}),
      ...(ledger.facts.tags ? { tags: ledger.facts.tags } : {}),
      ...(ledgerMetricCoverage(ledger.facts, 'skills') === 'recorded' || skillNames(ledger.facts).length > 0 ? { skills_used: skillNames(ledger.facts) } : {}),
      ticket_type: ledger.ticket_type,
      updated_at: ledger.updated_at,
      sdlc_iterations: eff.sdlc_iterations,
      active_duration_seconds: eff.active_duration_seconds,
 ...(eff.provisional_seconds ? { provisional_seconds: eff.provisional_seconds } : {}),
      active_duration_formatted: eff.active_duration_formatted,
      cost_usd: tot.cost_usd,
      ...(tot.cost_summary ? { cost_summary: tot.cost_summary } : {}),
      cache_hit_rate_pct: tot.cache_hit_rate_pct,
      test_to_code_ratio: eff.test_to_production_ratio.ratio,
      lines_added: eff.code_delivered.lines_added,
      lines_deleted: eff.code_delivered.lines_deleted,
      files_changed: eff.code_delivered.files_changed,
      tool_success_rate_pct: eff.tool_execution.success_rate_pct,
      context_health_status: cwh.status,
      root_saturation_pct: cwh.root_saturation_pct,
      peak_context_tokens: cwh.root_peak_tokens,
      ticket_title: title,
      started_at: ledger.started_at,
      completed_at: ledger.completed_at,
      lead_duration_seconds: leadSecs,
      lead_duration_formatted: formatLeadTime(leadSecs),
    };
  });

  if (!sortField) {
    return rows.sort((a, b) => {
      const compAHasVal = a.completed_at !== null;
      const compBHasVal = b.completed_at !== null;
      if (compAHasVal !== compBHasVal) {
        return compAHasVal ? -1 : 1;
      }
      const compAStr = a.completed_at ?? '';
      const compBStr = b.completed_at ?? '';
      if (compAStr !== compBStr) {
        return compBStr.localeCompare(compAStr);
      }

      const startAStr = a.started_at ?? '';
      const startBStr = b.started_at ?? '';
      if (startAStr !== startBStr) {
        return startBStr.localeCompare(startAStr);
      }

      return compareTicketIds(b.ticket_id, a.ticket_id, policy);
    });
  }

  return rows.sort((a, b) => {
    let cmp = 0;
    switch (sortField) {
      case 'completed_at': {
        const valA = a.completed_at ?? '';
        const valB = b.completed_at ?? '';
        if (a.completed_at === null && b.completed_at === null) {
          cmp = 0;
        } else if (a.completed_at === null) {
          cmp = 1;
        } else if (b.completed_at === null) {
          cmp = -1;
        } else {
          cmp = valA.localeCompare(valB);
        }
        break;
      }
      case 'started_at': {
        const valA = a.started_at ?? '';
        const valB = b.started_at ?? '';
        if (a.started_at === null && b.started_at === null) {
          cmp = 0;
        } else if (a.started_at === null) {
          cmp = 1;
        } else if (b.started_at === null) {
          cmp = -1;
        } else {
          cmp = valA.localeCompare(valB);
        }
        break;
      }
      case 'ticket_id':
        cmp = compareTicketIds(a.ticket_id, b.ticket_id, policy);
        break;
      case 'cost_usd':
        return compareCosts(a.cost_usd, b.cost_usd, sortOrder === 'asc') || compareTicketIds(a.ticket_id, b.ticket_id, policy);
      case 'active_duration_seconds':
        cmp = a.active_duration_seconds - b.active_duration_seconds;
        break;
      case 'sdlc_iterations':
        cmp = a.sdlc_iterations - b.sdlc_iterations;
        break;
      case 'cache_hit_rate_pct':
        cmp = a.cache_hit_rate_pct - b.cache_hit_rate_pct;
        break;
      case 'tool_success_rate_pct':
        cmp = a.tool_success_rate_pct - b.tool_success_rate_pct;
        break;
      case 'lines_added':
        cmp = a.lines_added - b.lines_added;
        break;
      case 'root_saturation_pct':
        cmp = a.root_saturation_pct - b.root_saturation_pct;
        break;
    }

    if (cmp !== 0) {
      return sortOrder === 'asc' ? cmp : -cmp;
    }
    return compareTicketIds(a.ticket_id, b.ticket_id, policy);
  });
}

/**
 * Normalizes raw ticket types by trimming, converting to lower case,
 * and mapping conventional aliases.
 */
export function canonicalTicketType(raw: string): string {
  const clean = raw.trim().toLowerCase();
  switch (clean) {
    case 'feat':
    case 'feature':
      return 'feature';
    case 'bug':
    case 'fix':
    case 'bugfix':
      return 'bug';
    case 'doc':
    case 'docs':
      return 'docs';
    case 'chore':
      return 'chore';
    case 'refactor':
      return 'refactor';
    case 'ci':
      return 'ci';
    case 'test':
      return 'test';
    case 'perf':
      return 'perf';
    case 'build':
      return 'build';
    default:
      return clean;
  }
}

/** Discover only prefix spelling present in a numeric recorded identifier. */
export function extractTicketPrefix(ticketId: string, policy?: TicketRecognitionPolicy): string {
  if (policy) {
    const matching = policy.prefixes.filter((prefix) => ticketNumber(ticketId, prefix, policy.caseInsensitive) !== null);
    matching.sort((a, b) => b.length - a.length || compareText(a, b));
    if (matching.length > 0) return matching[0]!;
  }
  const match = /^(.*?[^0-9])[0-9]+$/.exec(ticketId);
  return match?.[1] ?? '';
}

/** Preserve literal configured/observed spelling; empty evidence produces no options. */
export function combineDistinctPrefixes(
  discovered: readonly string[],
  configured: readonly string[] = [],
  caseInsensitive?: boolean
): readonly string[] {
  const values = new Map<string, string>();
  for (const token of [...configured, ...discovered]) {
    const clean = token.trim();
    if (!clean) continue;
    const key = caseInsensitive ? lowercaseTicketText(clean) : clean;
    if (!values.has(key)) values.set(key, clean);
  }
  return [...values.values()].sort(compareText);
}

/**
 * Filters ledgers by ticket types supporting canonical alias matching.
 * Returns all ledgers if types is empty or contains 'all'.
 */
export function filterLedgersByTicketTypes(
  ledgers: readonly AgenticMetricsLedger[],
  types: readonly string[]
): readonly AgenticMetricsLedger[] {
  if (types.length === 0) return ledgers;
  const normalizedTypes = types.map((t) => t.trim().toLowerCase()).filter(Boolean);
  if (normalizedTypes.length === 0 || normalizedTypes.includes('all')) {
    return ledgers;
  }
  const canonicalSet = new Set(normalizedTypes.map(canonicalTicketType));

  return ledgers.filter((ledger) => {
    if (!ledger.ticket_type) return false;
    const ledgerCanonical = canonicalTicketType(ledger.ticket_type);
    return canonicalSet.has(ledgerCanonical);
  });
}

/** A selected literal prefix may overlap another selected prefix. */
export function filterLedgersByPrefixes(
  ledgers: readonly AgenticMetricsLedger[],
  prefixes: readonly string[],
  caseInsensitive: boolean | undefined
): readonly AgenticMetricsLedger[] {
  if (caseInsensitive === undefined || prefixes.length === 0) return ledgers;
  const selected = prefixes.map((prefix) => prefix.trim()).filter(Boolean);
  if (selected.length === 0) return ledgers;
  return ledgers.filter((ledger) => selected.some((prefix) => ticketNumber(ledger.ticket_id ?? '', prefix, caseInsensitive) !== null));
}

export function formatChartAxisTimestamp(
  timestamp: string,
  granularity: TimeGranularity
): string {
  if (!timestamp) {
    return '';
  }
  if (granularity === 'hourly') {
    const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):/.exec(timestamp);
    if (match) {
      return `${match[4]}:00`;
    }
    const fallbackHour = /T(\d{2}):/.exec(timestamp);
    if (fallbackHour) {
      return `${fallbackHour[1]}:00`;
    }
    return timestamp;
  }
  if (timestamp.length >= 10) {
    return timestamp.slice(0, 10);
  }
  return timestamp;
}

export function formatCompactCount(
  count: number,
  options?: { showSign?: boolean; explicitPlus?: boolean }
): string {
  if (!Number.isFinite(count)) {
    return '0';
  }
  const explicitPlus = Boolean(options?.explicitPlus || options?.showSign);
  if (count === 0) {
    return explicitPlus ? '+0' : '0';
  }

  const isNegative = count < 0;
  const sign = isNegative ? '-' : explicitPlus ? '+' : '';
  const abs = Math.abs(count);

  if (abs >= 1_000_000) {
    const val = Math.round((abs / 1_000_000) * 10) / 10;
    const formatted = Number.isInteger(val) ? val.toFixed(0) : val.toFixed(1);
    return `${sign}${formatted}M`;
  }

  if (abs >= 1_000) {
    const val = Math.round((abs / 1_000) * 10) / 10;
    if (val >= 1000) {
      return `${sign}1M`;
    }
    const formatted = Number.isInteger(val) ? val.toFixed(0) : val.toFixed(1);
    return `${sign}${formatted}k`;
  }

  return `${sign}${Math.round(abs)}`;
}

export function formatLinesCompact(
  lines: number,
  options?: { explicitPlus?: boolean }
): string {
  if (!Number.isFinite(lines)) {
    return '0';
  }
  return formatCompactCount(Math.round(lines), options);
}
