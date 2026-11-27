import { describe, it, expect } from 'vitest';
import {
  emptyEquivalentCostSummary,
  equivalentCostSummaryForAmount,
  finalizeEquivalentCostSummary,
  hasEstimatedCost,
  invocationEquivalentCostSummary,
  mergeEquivalentCostSummaries,
} from './cost';
import type { AppliedRateSnapshot, InvocationFact } from './types';

const sampleRate: AppliedRateSnapshot = {
  matched_pattern: 'gpt-5.4',
  input_per_million: 2.0,
  cache_read_per_million: 0.25,
  output_per_million: 8.0,
  effective_catalog_id: 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
};

describe('Equivalent cost domain helpers', () => {
  it('produces all 6 provenance states with exact nullable amounts and coverage counts', () => {
    // 1. empty
    const empty = emptyEquivalentCostSummary();
    expect(empty).toEqual({
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
    });
    expect(hasEstimatedCost(empty)).toBe(false);

    // 2. measured (including genuine $0.00)
    const measuredZero = invocationEquivalentCostSummary({
      cost_usd: 0,
      cost_evidence: { state: 'observed', known_subtotal_usd: 0 },
    });
    expect(measuredZero).toEqual({
      provenance: 'measured',
      measured_cost_usd: 0,
      estimated_cost_usd: null,
      total_equivalent_cost_usd: 0,
      known_subtotal_usd: 0,
      measured: 1,
      estimated: 0,
      partial_measured: 0,
      unavailable: 0,
      total: 1,
    });
    expect(hasEstimatedCost(measuredZero)).toBe(false);

    const measured = invocationEquivalentCostSummary({
      cost_usd: 23.2,
      cost_evidence: { state: 'observed', known_subtotal_usd: 23.2 },
    });
    expect(measured.provenance).toBe('measured');
    expect(measured.total_equivalent_cost_usd).toBeCloseTo(23.2, 9);

    // 3. estimated
    const estimated = invocationEquivalentCostSummary({
      cost_usd: null,
      cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
      cost_estimate: { amount_usd: 7.12, rate: sampleRate },
    });
    expect(estimated).toEqual({
      provenance: 'estimated',
      measured_cost_usd: null,
      estimated_cost_usd: 7.12,
      total_equivalent_cost_usd: 7.12,
      known_subtotal_usd: 7.12,
      measured: 0,
      estimated: 1,
      partial_measured: 0,
      unavailable: 0,
      total: 1,
    });
    expect(hasEstimatedCost(estimated)).toBe(true);

    // 4. hybrid (measured + estimated, zero unpriced)
    const hybrid = mergeEquivalentCostSummaries([measured, estimated]);
    expect(hybrid.provenance).toBe('hybrid');
    expect(hybrid.measured_cost_usd).toBeCloseTo(23.2, 9);
    expect(hybrid.estimated_cost_usd).toBeCloseTo(7.12, 9);
    expect(hybrid.total_equivalent_cost_usd).toBeCloseTo(30.32, 9);
    expect(hybrid.known_subtotal_usd).toBeCloseTo(30.32, 9);
    expect(hybrid.measured).toBe(1);
    expect(hybrid.estimated).toBe(1);
    expect(hybrid.partial_measured).toBe(0);
    expect(hybrid.unavailable).toBe(0);
    expect(hybrid.total).toBe(2);
    expect(hasEstimatedCost(hybrid)).toBe(true);

    // 5. partial (measured + estimated + partial_measured + unavailable => total_equivalent_cost_usd is null)
    const partialMeasured = invocationEquivalentCostSummary({
      cost_usd: null,
      cost_evidence: { state: 'partial', known_subtotal_usd: 2.0 },
    });
    const unavailable = invocationEquivalentCostSummary({
      cost_usd: null,
      cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
    });
    const partialMixed = mergeEquivalentCostSummaries([measured, estimated, partialMeasured, unavailable]);
    expect(partialMixed.provenance).toBe('partial');
    expect(partialMixed.measured_cost_usd).toBeCloseTo(25.2, 9);
    expect(partialMixed.estimated_cost_usd).toBeCloseTo(7.12, 9);
    expect(partialMixed.total_equivalent_cost_usd).toBeNull();
    expect(partialMixed.known_subtotal_usd).toBeCloseTo(32.32, 9);
    expect(partialMixed.measured).toBe(1);
    expect(partialMixed.estimated).toBe(1);
    expect(partialMixed.partial_measured).toBe(1);
    expect(partialMixed.unavailable).toBe(1);
    expect(partialMixed.total).toBe(4);
    expect(hasEstimatedCost(partialMixed)).toBe(true);

    // 6. unavailable
    expect(unavailable).toEqual({
      provenance: 'unavailable',
      measured_cost_usd: null,
      estimated_cost_usd: null,
      total_equivalent_cost_usd: null,
      known_subtotal_usd: null,
      measured: 0,
      estimated: 0,
      partial_measured: 0,
      unavailable: 1,
      total: 1,
    });
    expect(hasEstimatedCost(unavailable)).toBe(false);
  });

  it('keeps genuine $0.00 configured estimate distinct from unavailable', () => {
    const zeroEstimate: Pick<InvocationFact, 'cost_usd' | 'cost_evidence' | 'cost_estimate'> = {
      cost_usd: null,
      cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
      cost_estimate: { amount_usd: 0, rate: { ...sampleRate, input_per_million: 0, cache_read_per_million: 0, output_per_million: 0 } },
    };
    const summary = invocationEquivalentCostSummary(zeroEstimate);
    expect(summary.provenance).toBe('estimated');
    expect(summary.estimated_cost_usd).toBe(0);
    expect(summary.total_equivalent_cost_usd).toBe(0);
    expect(hasEstimatedCost(summary)).toBe(true);
  });

  it('falls back to measured projection in equivalentCostSummaryForAmount when equivalentSummary is omitted', () => {
    const fallbackMeasured = equivalentCostSummaryForAmount(4.5, undefined, undefined, 2);
    expect(fallbackMeasured.provenance).toBe('measured');
    expect(fallbackMeasured.measured_cost_usd).toBe(4.5);
    expect(fallbackMeasured.estimated_cost_usd).toBeNull();
    expect(fallbackMeasured.total_equivalent_cost_usd).toBe(4.5);

    const fallbackUnavailable = equivalentCostSummaryForAmount(null, undefined, undefined, 3);
    expect(fallbackUnavailable.provenance).toBe('unavailable');
    expect(fallbackUnavailable.total_equivalent_cost_usd).toBeNull();

    const explicit = finalizeEquivalentCostSummary(0, 2, 0, 0, null, 6.75);
    expect(equivalentCostSummaryForAmount(null, undefined, explicit, 2)).toBe(explicit);
  });
});
