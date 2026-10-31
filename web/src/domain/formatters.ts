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

export const HEALTHY_CEILING_TOKENS = 200_000;

export function formatSkillStatus(success: boolean): 'Success' | 'Failed' {
  return success ? 'Success' : 'Failed';
}

export function formatSubagentRole(role: string): string {
  switch (role) {
    case 'orchestrator': return 'Orchestrator';
    case 'planner': return 'Planner';
    case 'coder': return 'Coder';
    case 'tester': return 'Tester';
    case 'reviewer': return 'Reviewer';
    default: return role || 'Unknown';
  }
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

export function formatIsoDateOnly(isoString: string | null | undefined): string {
  if (!isoString || isoString.length < 10) {
    return 'N/A';
  }
  return isoString.slice(0, 10);
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

const TICKET_ID_REGEX = /^([a-zA-Z]+)[-_]?([0-9]+)(.*)$/;

export function compareTicketIds(a: string, b: string): number {
  const cleanA = a.trim();
  const cleanB = b.trim();

  if (cleanA === cleanB) {
    return 0;
  }
  if (cleanA === '') {
    return -1;
  }
  if (cleanB === '') {
    return 1;
  }

  const matchA = TICKET_ID_REGEX.exec(cleanA);
  const matchB = TICKET_ID_REGEX.exec(cleanB);

  const hasMatchA = matchA !== null;
  const hasMatchB = matchB !== null;

  if (hasMatchA && hasMatchB) {
    const prefixA = matchA[1]!.toUpperCase();
    const prefixB = matchB[1]!.toUpperCase();
    if (prefixA !== prefixB) {
      return prefixA < prefixB ? -1 : 1;
    }

    const numA = parseInt(matchA[2]!, 10);
    const numB = parseInt(matchB[2]!, 10);

    if (numA !== numB) {
      return numA - numB;
    }

    return cleanA < cleanB ? -1 : cleanA > cleanB ? 1 : 0;
  }

  if (hasMatchA && !hasMatchB) {
    return -1;
  }
  if (!hasMatchA && hasMatchB) {
    return 1;
  }

  return cleanA < cleanB ? -1 : cleanA > cleanB ? 1 : 0;
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
  b: TicketDateComparable
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
  return compareTicketIds(a.ticket_id ?? '', b.ticket_id ?? '');
}

export function filterLedgersByCompletionDate<T extends TicketDateComparable>(
  ledgers: readonly T[],
  since?: string | number | null,
  until?: string | number | null
): T[] {
  if (!since && !until) {
    return [...ledgers];
  }

  let startMs: number | null = null;
  if (typeof since === 'number' && !Number.isNaN(since)) {
    startMs = since;
  } else if (typeof since === 'string' && since.trim().length > 0) {
    const cleanSince = since.trim();
    const parsed = cleanSince.length === 10
      ? Date.parse(`${cleanSince}T00:00:00.000Z`)
      : Date.parse(cleanSince);
    if (!Number.isNaN(parsed)) {
      startMs = parsed;
    }
  }

  let endMs: number | null = null;
  if (typeof until === 'number' && !Number.isNaN(until)) {
    endMs = until;
  } else if (typeof until === 'string' && until.trim().length > 0) {
    const cleanUntil = until.trim();
    const parsed = cleanUntil.length === 10
      ? Date.parse(`${cleanUntil}T23:59:59.999Z`)
      : Date.parse(cleanUntil);
    if (!Number.isNaN(parsed)) {
      endMs = parsed;
    }
  }

  if (startMs === null && endMs === null) {
    return [...ledgers];
  }

  return ledgers.filter((ledger) => {
    const effectiveMs = getEffectiveCompletionDateMs(ledger);
    if (effectiveMs === null) {
      return false;
    }
    if (startMs !== null && effectiveMs < startMs) {
      return false;
    }
    if (endMs !== null && effectiveMs > endMs) {
      return false;
    }
    return true;
  });
}

export function deriveTicketComparisonRows(
  ledgers: readonly AgenticMetricsLedger[],
  sortField?: ComparisonSortField,
  sortOrder: ComparisonSortOrder = 'desc'
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
      skills_used: skillNames(ledger.facts),
      ticket_type: ledger.ticket_type,
      updated_at: ledger.updated_at,
      sdlc_iterations: eff.sdlc_iterations,
      active_duration_seconds: eff.active_duration_seconds,
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

      return compareTicketIds(b.ticket_id, a.ticket_id);
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
        cmp = compareTicketIds(a.ticket_id, b.ticket_id);
        break;
      case 'cost_usd':
        return compareCosts(a.cost_usd, b.cost_usd, sortOrder === 'asc') || compareTicketIds(a.ticket_id, b.ticket_id);
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
    return compareTicketIds(a.ticket_id, b.ticket_id);
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

/**
 * Extracts ticket prefix in uppercase (e.g. "AL" from "AL-12" or "GEN" from "GEN_4").
 * Returns an empty string if no valid alphabetic prefix is identified.
 */
export function extractTicketPrefix(ticketId: string): string {
  const clean = ticketId.trim();
  if (!clean) return '';

  const delimMatch = /^([A-Za-z]+)[-_]/.exec(clean);
  if (delimMatch) {
    return delimMatch[1].toUpperCase();
  }

  const digitMatch = /^([A-Za-z]+)[0-9]/.exec(clean);
  if (digitMatch) {
    return digitMatch[1].toUpperCase();
  }

  return '';
}

/**
 * Extracts distinct ticket prefixes from ledgers, sorted ascending alphabetically.
 */
export function extractDistinctPrefixes(ledgers: readonly AgenticMetricsLedger[]): readonly string[] {
  const prefixSet = new Set<string>();
  for (const ledger of ledgers) {
    if (ledger.ticket_id) {
      const p = extractTicketPrefix(ledger.ticket_id);
      if (p) {
        prefixSet.add(p);
      }
    }
  }
  return Array.from(prefixSet).sort((a, b) => a.localeCompare(b));
}

/**
 * Combines discovered ticket prefixes with configured prefixes, deduplicates,
 * trims tokens, normalizes to uppercase, sorts ascending alphabetically,
 * and falls back to canonical defaults ['AL', 'GEN', 'PO'] if empty.
 */
export function combineDistinctPrefixes(
  discovered: readonly string[],
  configured?: readonly string[]
): readonly string[] {
  const prefixSet = new Set<string>();

  const addToken = (token: string) => {
    const clean = token.replace(/^[\s-_]+|[\s-_]+$/g, '');
    if (clean) {
      prefixSet.add(clean.toUpperCase());
    }
  };

  for (const item of discovered) {
    addToken(item);
  }

  if (configured) {
    for (const item of configured) {
      addToken(item);
    }
  }

  if (prefixSet.size === 0) {
    return ['AL', 'GEN', 'PO'];
  }

  return Array.from(prefixSet).sort((a, b) => a.localeCompare(b));
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

/**
 * Filters ledgers by workspace prefixes case-insensitively.
 * Returns all ledgers if prefixes is empty or contains 'all'.
 */
export function filterLedgersByPrefixes(
  ledgers: readonly AgenticMetricsLedger[],
  prefixes: readonly string[]
): readonly AgenticMetricsLedger[] {
  if (prefixes.length === 0) return ledgers;
  const normalizedPrefixes = prefixes.map((p) => p.trim().toUpperCase()).filter(Boolean);
  if (normalizedPrefixes.length === 0 || normalizedPrefixes.includes('ALL')) {
    return ledgers;
  }
  const prefixSet = new Set(normalizedPrefixes);

  return ledgers.filter((ledger) => {
    if (!ledger.ticket_id) return false;
    const p = extractTicketPrefix(ledger.ticket_id);
    return prefixSet.has(p);
  });
}

/**
 * Pure multi-filter for ticket comparison rows by ticket types and prefixes.
 */
export function filterComparisonRows(
  rows: readonly TicketComparisonRow[],
  types: readonly string[],
  prefixes: readonly string[]
): readonly TicketComparisonRow[] {
  let filtered = rows;

  const normalizedTypes = types.map((t) => t.trim().toLowerCase()).filter(Boolean);
  if (normalizedTypes.length > 0 && !normalizedTypes.includes('all')) {
    const canonicalSet = new Set(normalizedTypes.map(canonicalTicketType));
    filtered = filtered.filter((row) => {
      if (!row.ticket_type) return false;
      const rowCanonical = canonicalTicketType(row.ticket_type);
      return canonicalSet.has(rowCanonical);
    });
  }

  const normalizedPrefixes = prefixes.map((p) => p.trim().toUpperCase()).filter(Boolean);
  if (normalizedPrefixes.length > 0 && !normalizedPrefixes.includes('ALL')) {
    const prefixSet = new Set(normalizedPrefixes);
    filtered = filtered.filter((row) => {
      const p = extractTicketPrefix(row.ticket_id);
      return prefixSet.has(p);
    });
  }

  return filtered;
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
