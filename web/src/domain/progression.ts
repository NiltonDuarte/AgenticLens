import type { AgenticMetricsLedger, TimeBucketMetric, TimeGranularity } from './types';

interface MutableBucketAccumulator {
  invocations: number;
  cost_usd: number;
  input_tokens: number;
  output_tokens: number;
  cache_read_tokens: number;
  reasoning_tokens: number;
  active_duration_seconds: number;
  peak_context_tokens: number;
  lines_added: number;
  lines_deleted: number;
  files_changed: number;
}

export function parseIsoDateToUtc(ts: string | null | undefined): Date | null {
  if (!ts || ts.length < 10) {
    return null;
  }
  const clean = ts.slice(0, 10);
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(clean);
  if (!match) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);

  if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day)) {
    return null;
  }

  const d = new Date(Date.UTC(year, month - 1, day, 0, 0, 0, 0));
  if (Number.isNaN(d.getTime())) {
    return null;
  }
  if (d.getUTCFullYear() !== year || d.getUTCMonth() !== month - 1 || d.getUTCDate() !== day) {
    return null;
  }

  return d;
}

function formatUtcDate(d: Date): string {
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, '0');
  const day = String(d.getUTCDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function getBucketKey(d: Date, isWeekly: boolean): string {
  if (isWeekly) {
    let weekday = d.getUTCDay();
    if (weekday === 0) {
      weekday = 7;
    }
    const mondayMs = d.getTime() - (weekday - 1) * 86_400_000;
    const monday = new Date(mondayMs);
    return formatUtcDate(monday);
  }
  return formatUtcDate(d);
}

export function aggregateMetricsOverTime(
  ledgers: readonly AgenticMetricsLedger[],
  granularity: TimeGranularity,
  startDate?: string | null,
  endDate?: string | null
): readonly TimeBucketMetric[] {
  if (!ledgers || ledgers.length === 0) {
    return [];
  }

  const isWeekly = granularity === 'weekly';
  const startUtc = parseIsoDateToUtc(startDate);
  const endUtc = parseIsoDateToUtc(endDate);

  const accumulators = new Map<string, MutableBucketAccumulator>();

  const getOrCreate = (key: string): MutableBucketAccumulator => {
    const existing = accumulators.get(key);
    if (existing) {
      return existing;
    }
    const initial: MutableBucketAccumulator = {
      invocations: 0,
      cost_usd: 0,
      input_tokens: 0,
      output_tokens: 0,
      cache_read_tokens: 0,
      reasoning_tokens: 0,
      active_duration_seconds: 0,
      peak_context_tokens: 0,
      lines_added: 0,
      lines_deleted: 0,
      files_changed: 0,
    };
    accumulators.set(key, initial);
    return initial;
  };

  for (const ledger of ledgers) {
    const invocations = ledger.facts?.invocations ?? [];
    for (const inv of invocations) {
      const invDate =
        parseIsoDateToUtc(inv.completed_at) ??
        parseIsoDateToUtc(inv.started_at) ??
        parseIsoDateToUtc(ledger.completed_at) ??
        parseIsoDateToUtc(ledger.started_at) ??
        parseIsoDateToUtc(ledger.updated_at);

      if (!invDate) {
        continue;
      }
      if (startUtc && invDate.getTime() < startUtc.getTime()) {
        continue;
      }
      if (endUtc && invDate.getTime() > endUtc.getTime()) {
        continue;
      }

      const bkey = getBucketKey(invDate, isWeekly);
      const acc = getOrCreate(bkey);

      acc.invocations += 1;
      acc.cost_usd += inv.cost_usd;
      acc.input_tokens += inv.tokens.input;
      acc.output_tokens += inv.tokens.output;
      acc.cache_read_tokens += inv.tokens.cache_read;
      acc.reasoning_tokens += inv.tokens.reasoning;
      acc.active_duration_seconds += inv.active_duration_seconds;
      if (inv.tokens.peak_context > acc.peak_context_tokens) {
        acc.peak_context_tokens = inv.tokens.peak_context;
      }
    }

    const codeDate =
      parseIsoDateToUtc(ledger.completed_at) ??
      parseIsoDateToUtc(ledger.started_at) ??
      parseIsoDateToUtc(ledger.updated_at);

    if (codeDate) {
      let inRange = true;
      if (startUtc && codeDate.getTime() < startUtc.getTime()) {
        inRange = false;
      }
      if (endUtc && codeDate.getTime() > endUtc.getTime()) {
        inRange = false;
      }

      if (inRange) {
        const bkey = getBucketKey(codeDate, isWeekly);
        const acc = getOrCreate(bkey);
        const code = ledger.workflow_efficiency.code_delivered;
        acc.lines_added += code.lines_added;
        acc.lines_deleted += code.lines_deleted;
        acc.files_changed += code.files_changed;
      }
    }
  }

  if (accumulators.size === 0) {
    return [];
  }

  const sortedKeys = Array.from(accumulators.keys()).sort();

  let runningCumCost = 0;
  let runningCumInput = 0;
  let runningCumOutput = 0;
  let runningCumLinesAdded = 0;
  let runningCumLinesDeleted = 0;
  let runningCumInvocations = 0;

  const results: TimeBucketMetric[] = [];

  for (const bkey of sortedKeys) {
    const acc = accumulators.get(bkey)!;
    runningCumCost += acc.cost_usd;
    runningCumInput += acc.input_tokens;
    runningCumOutput += acc.output_tokens;
    runningCumLinesAdded += acc.lines_added;
    runningCumLinesDeleted += acc.lines_deleted;
    runningCumInvocations += acc.invocations;

    const cachedOrInp = acc.input_tokens + acc.cache_read_tokens;
    const hitPct =
      cachedOrInp > 0
        ? Math.round((acc.cache_read_tokens / cachedOrInp) * 1000) / 10
        : 0.0;

    results.push({
      bucket_key: bkey,
      bucket_date: bkey,
      invocations: acc.invocations,
      cost_usd: Math.round(acc.cost_usd * 10_000) / 10_000,
      input_tokens: acc.input_tokens,
      output_tokens: acc.output_tokens,
      cache_read_tokens: acc.cache_read_tokens,
      reasoning_tokens: acc.reasoning_tokens,
      cache_hit_rate_pct: hitPct,
      lines_added: acc.lines_added,
      lines_deleted: acc.lines_deleted,
      files_changed: acc.files_changed,
      net_lines_added: acc.lines_added - acc.lines_deleted,
      peak_context_tokens: acc.peak_context_tokens,
      active_duration_seconds: acc.active_duration_seconds,
      cumulative_cost_usd: Math.round(runningCumCost * 10_000) / 10_000,
      cumulative_input_tokens: runningCumInput,
      cumulative_output_tokens: runningCumOutput,
      cumulative_lines_added: runningCumLinesAdded,
      cumulative_lines_deleted: runningCumLinesDeleted,
      cumulative_invocations: runningCumInvocations,
    });
  }

  return results;
}
