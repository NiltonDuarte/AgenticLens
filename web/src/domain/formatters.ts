import type {
  AgenticMetricsLedger,
  ComparisonSortField,
  ComparisonSortOrder,
  TicketComparisonRow,
} from './types';

export const HEALTHY_CEILING_TOKENS = 200_000;

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

export function formatCurrencyUSD(cost: number): string {
  if (!Number.isFinite(cost) || cost <= 0) {
    return '$0.00';
  }
  if (cost < 1) {
    return `$${cost.toFixed(4)}`;
  }
  return `$${cost.toFixed(2)}`;
}

export function formatPercentage(pct: number): string {
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
      updated_at: ledger.updated_at,
      sdlc_iterations: eff.sdlc_iterations,
      active_duration_seconds: eff.active_duration_seconds,
      active_duration_formatted: eff.active_duration_formatted,
      cost_usd: tot.cost_usd,
      cache_hit_rate_pct: tot.cache_hit_rate_pct,
      test_to_code_ratio: eff.test_to_production_ratio.ratio,
      lines_added: eff.code_delivered.lines_added,
      lines_deleted: eff.code_delivered.lines_deleted,
      files_changed: eff.code_delivered.files_changed,
      tool_success_rate_pct: eff.tool_execution.success_rate_pct,
      context_health_status: cwh.status,
      orchestrator_saturation_pct: cwh.orchestrator_saturation_pct,
      peak_context_tokens: cwh.orchestrator_peak_tokens,
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

      return b.ticket_id.localeCompare(a.ticket_id);
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
        cmp = a.ticket_id.localeCompare(b.ticket_id);
        break;
      case 'cost_usd':
        cmp = a.cost_usd - b.cost_usd;
        break;
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
      case 'orchestrator_saturation_pct':
        cmp = a.orchestrator_saturation_pct - b.orchestrator_saturation_pct;
        break;
    }

    if (cmp !== 0) {
      return sortOrder === 'asc' ? cmp : -cmp;
    }
    return a.ticket_id.localeCompare(b.ticket_id);
  });
}
