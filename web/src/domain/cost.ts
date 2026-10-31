import type { CostSummary, InvocationFact } from './types';

export function emptyCostSummary(): CostSummary {
  return { state: 'empty', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 0, total: 0 };
}

/** Numeric legacy facts follow the approved observed-cost compatibility policy. */
export function invocationCostSummary(invocation: Pick<InvocationFact, 'cost_usd' | 'cost_evidence'>): CostSummary {
  const evidence = invocation.cost_evidence;
  if (evidence?.state === 'partial') {
    return { state: 'partial', known_subtotal_usd: evidence.known_subtotal_usd, fully_observed: 0, partial: 1, unavailable: 0, total: 1 };
  }
  if (evidence?.state === 'unavailable' || invocation.cost_usd == null) {
    return { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 };
  }
  return { state: 'complete', known_subtotal_usd: invocation.cost_usd, fully_observed: 1, partial: 0, unavailable: 0, total: 1 };
}

export function mergeCostSummaries(summaries: readonly CostSummary[]): CostSummary {
  let fully_observed = 0;
  let partial = 0;
  let unavailable = 0;
  let subtotal: number | null = null;
  for (const summary of summaries) {
    fully_observed += summary.fully_observed;
    partial += summary.partial;
    unavailable += summary.unavailable;
    if (summary.known_subtotal_usd !== null) subtotal = (subtotal ?? 0) + summary.known_subtotal_usd;
  }
  const total = fully_observed + partial + unavailable;
  return {
    state: total === 0 ? 'empty' : fully_observed === total ? 'complete' : subtotal === null ? 'unavailable' : 'partial',
    known_subtotal_usd: subtotal,
    fully_observed, partial, unavailable, total,
  };
}

export function costSummaryForAmount(amount: number | null | undefined, summary?: CostSummary, count = 1): CostSummary {
  if (summary) return summary;
  if (count === 0) return emptyCostSummary();
  return amount == null
    ? { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: count, total: count }
    : { state: 'complete', known_subtotal_usd: amount, fully_observed: count, partial: 0, unavailable: 0, total: count };
}

export function completeCost(summary: CostSummary): number | null {
  return summary.state === 'complete' ? summary.known_subtotal_usd : null;
}

/** Missing costs follow complete values regardless of the requested direction. */
export function compareCosts(a: number | null, b: number | null, ascending: boolean): number {
  if (a === null || b === null) return a === b ? 0 : a === null ? 1 : -1;
  return ascending ? a - b : b - a;
}
