import type { CostProvenance, CostSummary, EquivalentCostSummary, InvocationFact } from './types';

export function emptyCostSummary(): CostSummary {
  return { state: 'empty', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 0, total: 0 };
}

export function emptyEquivalentCostSummary(): EquivalentCostSummary {
  return {
    provenance: 'empty',
    measured_cost_usd: null,
    estimated_cost_usd: null,
    total_equivalent_cost_usd: null,
    known_subtotal_usd: null,
    measured: 0,
    estimated: 0,
    partial_measured: 0,
    unavailable: 0,
    total: 0,
  };
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

export function finalizeEquivalentCostSummary(
  measured: number,
  estimated: number,
  partialMeasured: number,
  unavailable: number,
  measuredCostUSD: number | null,
  estimatedCostUSD: number | null
): EquivalentCostSummary {
  const total = measured + estimated + partialMeasured + unavailable;
  if (total === 0) return emptyEquivalentCostSummary();

  const measuredAmount = measured === 0 && partialMeasured === 0 ? null : measuredCostUSD;
  const estimatedAmount = estimated === 0 ? null : estimatedCostUSD;
  const knownSubtotal =
    measuredAmount !== null || estimatedAmount !== null
      ? (measuredAmount ?? 0) + (estimatedAmount ?? 0)
      : null;

  let provenance: CostProvenance;
  let totalEquivalent: number | null;
  let finalMeasured = measuredAmount;
  let finalEstimated = estimatedAmount;
  let finalSubtotal = knownSubtotal;

  if (measured === total) {
    provenance = 'measured';
    finalEstimated = null;
    totalEquivalent = finalMeasured;
  } else if (estimated === total) {
    provenance = 'estimated';
    finalMeasured = null;
    totalEquivalent = finalEstimated;
  } else if (measured > 0 && estimated > 0 && measured + estimated === total) {
    provenance = 'hybrid';
    totalEquivalent = finalSubtotal;
  } else if (finalSubtotal !== null) {
    provenance = 'partial';
    totalEquivalent = null;
  } else {
    provenance = 'unavailable';
    finalMeasured = null;
    finalEstimated = null;
    totalEquivalent = null;
    finalSubtotal = null;
  }

  return {
    provenance,
    measured_cost_usd: finalMeasured,
    estimated_cost_usd: finalEstimated,
    total_equivalent_cost_usd: totalEquivalent,
    known_subtotal_usd: finalSubtotal,
    measured,
    estimated,
    partial_measured: partialMeasured,
    unavailable,
    total,
  };
}

export function invocationEquivalentCostSummary(
  invocation: Pick<InvocationFact, 'cost_usd' | 'cost_evidence' | 'cost_estimate'>
): EquivalentCostSummary {
  const evidence = invocation.cost_evidence;
  if (evidence?.state === 'partial') {
    return finalizeEquivalentCostSummary(0, 0, 1, 0, evidence.known_subtotal_usd, null);
  }
  if (evidence?.state === 'unavailable' || invocation.cost_usd == null) {
    if (invocation.cost_estimate !== undefined) {
      return finalizeEquivalentCostSummary(0, 1, 0, 0, null, invocation.cost_estimate.amount_usd);
    }
    return finalizeEquivalentCostSummary(0, 0, 0, 1, null, null);
  }
  return finalizeEquivalentCostSummary(1, 0, 0, 0, invocation.cost_usd, null);
}

export function mergeEquivalentCostSummaries(summaries: readonly EquivalentCostSummary[]): EquivalentCostSummary {
  let measured = 0;
  let estimated = 0;
  let partialMeasured = 0;
  let unavailable = 0;
  let measuredCost: number | null = null;
  let estimatedCost: number | null = null;
  for (const summary of summaries) {
    measured += summary.measured;
    estimated += summary.estimated;
    partialMeasured += summary.partial_measured;
    unavailable += summary.unavailable;
    if (summary.measured_cost_usd !== null) measuredCost = (measuredCost ?? 0) + summary.measured_cost_usd;
    if (summary.estimated_cost_usd !== null) estimatedCost = (estimatedCost ?? 0) + summary.estimated_cost_usd;
  }
  return finalizeEquivalentCostSummary(measured, estimated, partialMeasured, unavailable, measuredCost, estimatedCost);
}

export function equivalentCostSummaryForAmount(
  amount: number | null | undefined,
  measuredSummary?: CostSummary,
  equivalentSummary?: EquivalentCostSummary,
  count = 1
): EquivalentCostSummary {
  if (equivalentSummary) return equivalentSummary;
  const measured = costSummaryForAmount(amount, measuredSummary, count);
  return finalizeEquivalentCostSummary(
    measured.fully_observed,
    0,
    measured.partial,
    measured.unavailable,
    measured.fully_observed > 0 || measured.partial > 0 ? measured.known_subtotal_usd : null,
    null
  );
}

export function hasEstimatedCost(summary: EquivalentCostSummary | undefined): boolean {
  return summary !== undefined && summary.estimated > 0;
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
