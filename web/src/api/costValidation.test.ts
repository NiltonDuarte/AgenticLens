import { describe, it, expect } from 'vitest';
import {
  validateComparisonRowCosts,
  validateEquivalentCostSummary,
  validateFleetCosts,
  validateLedgerCosts,
  validateTimeBucketCosts,
} from './costValidation';
import { decodeDetail, decodeFleet, decodeLedgers } from './validation';
import { mockAgenticMetricsLedgers, mockFleetSummary, mockWorkspaceDetails } from '../test/fixtures/mockData';

const validRate = {
  matched_pattern: 'gpt-5.4',
  input_per_million: 2.0,
  cache_read_per_million: 0.25,
  output_per_million: 8.0,
  effective_catalog_id: 'sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
};

describe('Strict equivalent-cost boundary validation', () => {
  describe('validateEquivalentCostSummary', () => {
    const validSummaries: readonly [string, unknown][] = [
      [
        'empty',
        {
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
        },
      ],
      [
        'measured',
        {
          provenance: 'measured',
          measured_cost_usd: 4.5,
          estimated_cost_usd: null,
          total_equivalent_cost_usd: 4.5,
          known_subtotal_usd: 4.5,
          measured: 2,
          estimated: 0,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
      ],
      [
        'estimated',
        {
          provenance: 'estimated',
          measured_cost_usd: null,
          estimated_cost_usd: 6.75,
          total_equivalent_cost_usd: 6.75,
          known_subtotal_usd: 6.75,
          measured: 0,
          estimated: 2,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
      ],
      [
        'hybrid',
        {
          provenance: 'hybrid',
          measured_cost_usd: 23.2,
          estimated_cost_usd: 7.12,
          total_equivalent_cost_usd: 30.32,
          known_subtotal_usd: 30.32,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
      ],
      [
        'partial',
        {
          provenance: 'partial',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: null,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 3,
        },
      ],
      [
        'unavailable',
        {
          provenance: 'unavailable',
          measured_cost_usd: null,
          estimated_cost_usd: null,
          total_equivalent_cost_usd: null,
          known_subtotal_usd: null,
          measured: 0,
          estimated: 0,
          partial_measured: 0,
          unavailable: 2,
          total: 2,
        },
      ],
    ];

    it.each(validSummaries)('accepts valid %s summary', (_name, summary) => {
      expect(() => validateEquivalentCostSummary(summary, 'test.summary')).not.toThrow();
    });

    const invalidSummaries: readonly [string, unknown][] = [
      [
        'unknown provenance',
        {
          provenance: 'approximated',
          measured_cost_usd: null,
          estimated_cost_usd: 1.0,
          total_equivalent_cost_usd: 1.0,
          known_subtotal_usd: 1.0,
          measured: 0,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 1,
        },
      ],
      [
        'count sum mismatch',
        {
          provenance: 'hybrid',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: 5.0,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 3,
        },
      ],
      [
        'non-null total_equivalent_cost_usd on partial provenance',
        {
          provenance: 'partial',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: 5.0,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 3,
        },
      ],
      [
        'mismatched provenance claiming hybrid when unavailable count > 0',
        {
          provenance: 'hybrid',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: 5.0,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 3,
        },
      ],
      [
        'negative estimated_cost_usd',
        {
          provenance: 'estimated',
          measured_cost_usd: null,
          estimated_cost_usd: -1.0,
          total_equivalent_cost_usd: -1.0,
          known_subtotal_usd: -1.0,
          measured: 0,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 1,
        },
      ],
      [
        'hybrid total does not equal measured + estimated',
        {
          provenance: 'hybrid',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: 6.0,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
      ],
    ];

    it.each(invalidSummaries)('rejects %s', (_name, summary) => {
      expect(() => validateEquivalentCostSummary(summary, 'test.summary')).toThrow(/Invalid cost payload/);
    });
  });

  describe('validateLedgerCosts invocation cost_estimate and applied_rates rules', () => {
    const baseLedger = mockAgenticMetricsLedgers['AL-1']!;

    const makeLedgerWithInvocation = (invocation: Record<string, unknown>, modelExtra: Record<string, unknown> = {}) => ({
      ...baseLedger,
      facts: {
        ...baseLedger.facts,
        schema_version: 6,
        sessions: [
          {
            id: 'ses-1',
            source: 'codex',
            invocations: [invocation],
            tools: [],
          },
        ],
      },
      totals: {
        ...baseLedger.totals,
        cost_usd: null,
        cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        measured_cost_usd: null,
        estimated_cost_usd: 1.0,
        total_equivalent_cost_usd: 1.0,
        cost_provenance: 'estimated',
        equivalent_cost_summary: {
          provenance: 'estimated',
          measured_cost_usd: null,
          estimated_cost_usd: 1.0,
          total_equivalent_cost_usd: 1.0,
          known_subtotal_usd: 1.0,
          measured: 0,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 1,
        },
      },
      roles: [],
      models: [
        {
          ...baseLedger.models[0]!,
          cost_usd: null,
          cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
          measured_cost_usd: null,
          estimated_cost_usd: 1.0,
          total_equivalent_cost_usd: 1.0,
          cost_provenance: 'estimated',
          equivalent_cost_summary: {
            provenance: 'estimated',
            measured_cost_usd: null,
            estimated_cost_usd: 1.0,
            total_equivalent_cost_usd: 1.0,
            known_subtotal_usd: 1.0,
            measured: 0,
            estimated: 1,
            partial_measured: 0,
            unavailable: 0,
            total: 1,
          },
          applied_rates: [validRate],
          ...modelExtra,
        },
      ],
      threads: [],
    });

    it('accepts valid estimated invocation with applied_rates', () => {
      const payload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 200000, output: 50000, cache_read: 800000, reasoning: 0, peak_context: 50000 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });
      expect(() => validateLedgerCosts(payload)).not.toThrow();
    });

    it('rejects cost_estimate on observed invocation', () => {
      const payload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: 1.0,
        cost_evidence: { state: 'observed', known_subtotal_usd: 1.0 },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });
      expect(() => validateLedgerCosts(payload)).toThrow(/Invalid cost payload/);
    });

    it('rejects cost_estimate on partial invocation', () => {
      const payload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'partial', known_subtotal_usd: 0.5 },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });
      expect(() => validateLedgerCosts(payload)).toThrow(/Invalid cost payload/);
    });

    it('rejects invalid applied_rates entries on model metrics', () => {
      const baseInv = {
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      };
      expect(() => validateLedgerCosts(makeLedgerWithInvocation(baseInv, { applied_rates: 'not-array' }))).toThrow(/Invalid cost payload/);
      expect(() => validateLedgerCosts(makeLedgerWithInvocation(baseInv, { applied_rates: [{ ...validRate, matched_pattern: '  ' }] }))).toThrow(/Invalid cost payload/);
      expect(() => validateLedgerCosts(makeLedgerWithInvocation(baseInv, { applied_rates: [{ ...validRate, effective_catalog_id: '' }] }))).toThrow(/Invalid cost payload/);
      expect(() => validateLedgerCosts(makeLedgerWithInvocation(baseInv, { applied_rates: [{ ...validRate, output_per_million: -5 }] }))).toThrow(/Invalid cost payload/);
    });

    it('rejects top-level cost_provenance mismatch against equivalent_cost_summary', () => {
      const payload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });
      const tampered = { ...payload, totals: { ...payload.totals, cost_provenance: 'measured' } };
      expect(() => validateLedgerCosts(tampered)).toThrow(/Invalid cost payload/);
    });

    it('rejects malformed cost_estimate on unavailable invocation (negative amount, null, missing rate fields)', () => {
      const baseInv = {
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        compactions: 0,
      };
      const malformedEstimates: readonly unknown[] = [
        null,
        { amount_usd: -1, rate: validRate },
        { amount_usd: Number.NaN, rate: validRate },
        { amount_usd: 1.0, rate: null },
        { amount_usd: 1.0 },
        { amount_usd: 1.0, rate: { ...validRate, input_per_million: undefined } },
        { amount_usd: 1.0, rate: { ...validRate, cache_read_per_million: undefined } },
        { amount_usd: 1.0, rate: { ...validRate, output_per_million: undefined } },
        { amount_usd: 1.0, rate: { ...validRate, matched_pattern: '' } },
        { amount_usd: 1.0, rate: { ...validRate, effective_catalog_id: '   ' } },
      ];
      for (const badEstimate of malformedEstimates) {
        expect(() =>
          validateLedgerCosts(makeLedgerWithInvocation({ ...baseInv, cost_estimate: badEstimate }))
        ).toThrow(/Invalid cost payload/);
      }
    });

    it('rejects missing sibling equivalent fields on totals', () => {
      const validPayload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });

      const siblingKeys = [
        'cost_provenance',
        'total_equivalent_cost_usd',
        'measured_cost_usd',
        'estimated_cost_usd',
        'equivalent_cost_summary',
      ] as const;

      for (const omitKey of siblingKeys) {
        const totalsCopy: Record<string, unknown> = { ...validPayload.totals };
        delete totalsCopy[omitKey];
        expect(() => validateLedgerCosts({ ...validPayload, totals: totalsCopy })).toThrow(/Invalid cost payload/);
      }
    });

    it('rejects invalid equivalent fields on roles and threads[].roles[].models[]', () => {
      const validPayload = makeLedgerWithInvocation({
        role: 'coder',
        model: 'gpt-5.4',
        started_at: null,
        completed_at: null,
        active_duration_seconds: 10,
        turns: 1,
        tokens: { input: 100, output: 50, cache_read: 0, reasoning: 0, peak_context: 200 },
        cost_usd: null,
        cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
        cost_estimate: { amount_usd: 1.0, rate: validRate },
        compactions: 0,
      });
      const validEquiv = {
        provenance: 'estimated',
        measured_cost_usd: null,
        estimated_cost_usd: 1.0,
        total_equivalent_cost_usd: 1.0,
        known_subtotal_usd: 1.0,
        measured: 0,
        estimated: 1,
        partial_measured: 0,
        unavailable: 0,
        total: 1,
      };
      const baseRole = {
        ...baseLedger.roles[0]!,
        cost_usd: null,
        cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        measured_cost_usd: null,
        estimated_cost_usd: 1.0,
        total_equivalent_cost_usd: 1.0,
        cost_provenance: 'estimated',
        equivalent_cost_summary: validEquiv,
      };
      const baseThreadModel = {
        ...baseLedger.models[0]!,
        model: 'gpt-5.4',
        cost_usd: null,
        cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        measured_cost_usd: null,
        estimated_cost_usd: 1.0,
        total_equivalent_cost_usd: 1.0,
        cost_provenance: 'estimated',
        equivalent_cost_summary: validEquiv,
        applied_rates: [validRate],
      };

      // Invalid role provenance mismatch
      expect(() =>
        validateLedgerCosts({
          ...validPayload,
          roles: [{ ...baseRole, cost_provenance: 'hybrid' }],
        })
      ).toThrow(/Invalid cost payload/);

      // Missing sibling on role
      const roleMissingSibling: Record<string, unknown> = { ...baseRole };
      delete roleMissingSibling.total_equivalent_cost_usd;
      expect(() =>
        validateLedgerCosts({
          ...validPayload,
          roles: [roleMissingSibling],
        })
      ).toThrow(/Invalid cost payload/);

      // Invalid threads[].roles[].models[] equivalent amount mismatch
      expect(() =>
        validateLedgerCosts({
          ...validPayload,
          threads: [
            {
              session_id: 'ses-1',
              user: 'tester',
              compactions: 0,
              roles: [{ role: 'coder', models: [{ ...baseThreadModel, estimated_cost_usd: 9.99 }] }],
            },
          ],
        })
      ).toThrow(/Invalid cost payload/);

      // Invalid threads[].roles[].models[] applied_rates
      expect(() =>
        validateLedgerCosts({
          ...validPayload,
          threads: [
            {
              session_id: 'ses-1',
              user: 'tester',
              compactions: 0,
              roles: [{ role: 'coder', models: [{ ...baseThreadModel, applied_rates: [{ ...validRate, input_per_million: -1 }] }] }],
            },
          ],
        })
      ).toThrow(/Invalid cost payload/);
    });
  });

  describe('fleet, time bucket, and comparison row equivalent-cost validation', () => {
    it('validates hybrid fleet summary and rejects mismatched measured/equivalent counts', () => {
      const hybridFleet = {
        ...mockFleetSummary,
        total_cost_usd: null,
        cost_summary: { state: 'partial', known_subtotal_usd: 10.0, fully_observed: 1, partial: 0, unavailable: 1, total: 2 },
        measured_cost_usd: 10.0,
        estimated_cost_usd: 5.0,
        total_equivalent_cost_usd: 15.0,
        cost_provenance: 'hybrid',
        equivalent_cost_summary: {
          provenance: 'hybrid',
          measured_cost_usd: 10.0,
          estimated_cost_usd: 5.0,
          total_equivalent_cost_usd: 15.0,
          known_subtotal_usd: 15.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
        role_costs: { coder: null },
        role_cost_summaries: { coder: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 } },
        role_equivalent_cost_summaries: {
          coder: {
            provenance: 'estimated',
            measured_cost_usd: null,
            estimated_cost_usd: 5.0,
            total_equivalent_cost_usd: 5.0,
            known_subtotal_usd: 5.0,
            measured: 0,
            estimated: 1,
            partial_measured: 0,
            unavailable: 0,
            total: 1,
          },
        },
        model_costs: { 'gpt-5.4': null },
        model_cost_summaries: { 'gpt-5.4': { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 } },
        model_equivalent_cost_summaries: {
          'gpt-5.4': {
            provenance: 'estimated',
            measured_cost_usd: null,
            estimated_cost_usd: 5.0,
            total_equivalent_cost_usd: 5.0,
            known_subtotal_usd: 5.0,
            measured: 0,
            estimated: 1,
            partial_measured: 0,
            unavailable: 0,
            total: 1,
          },
        },
      };
      expect(() => validateFleetCosts(hybridFleet)).not.toThrow();

      const badCounts = {
        ...hybridFleet,
        equivalent_cost_summary: { ...hybridFleet.equivalent_cost_summary, measured: 2, estimated: 1, total: 3 },
      };
      expect(() => validateFleetCosts(badCounts)).toThrow(/Invalid cost payload/);

      // Asymmetric role_equivalent_cost_summaries keys (extra key not in role_costs, or missing key from role_costs)
      expect(() =>
        validateFleetCosts({
          ...hybridFleet,
          role_equivalent_cost_summaries: {
            ...hybridFleet.role_equivalent_cost_summaries,
            reviewer: hybridFleet.role_equivalent_cost_summaries.coder,
          },
        })
      ).toThrow(/Invalid cost payload/);
      expect(() =>
        validateFleetCosts({
          ...hybridFleet,
          role_equivalent_cost_summaries: {},
        })
      ).toThrow(/Invalid cost payload/);

      // Asymmetric model_equivalent_cost_summaries keys (extra key not in model_costs, or missing key from model_costs)
      expect(() =>
        validateFleetCosts({
          ...hybridFleet,
          model_equivalent_cost_summaries: {
            ...hybridFleet.model_equivalent_cost_summaries,
            'extra-model': hybridFleet.model_equivalent_cost_summaries['gpt-5.4'],
          },
        })
      ).toThrow(/Invalid cost payload/);
      expect(() =>
        validateFleetCosts({
          ...hybridFleet,
          model_equivalent_cost_summaries: {},
        })
      ).toThrow(/Invalid cost payload/);
    });

    it('validates time bucket and cumulative equivalent costs and rejects partial non-null cumulative total', () => {
      const bucket = {
        cost_usd: null,
        cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        measured_cost_usd: null,
        estimated_cost_usd: 4.0,
        total_equivalent_cost_usd: 4.0,
        cost_provenance: 'estimated',
        equivalent_cost_summary: {
          provenance: 'estimated',
          measured_cost_usd: null,
          estimated_cost_usd: 4.0,
          total_equivalent_cost_usd: 4.0,
          known_subtotal_usd: 4.0,
          measured: 0,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 1,
        },
        cumulative_cost_usd: null,
        cumulative_cost_summary: { state: 'partial', known_subtotal_usd: 10.0, fully_observed: 1, partial: 0, unavailable: 2, total: 3 },
        cumulative_measured_cost_usd: 10.0,
        cumulative_estimated_cost_usd: 4.0,
        cumulative_total_equivalent_cost_usd: null,
        cumulative_cost_provenance: 'partial',
        cumulative_equivalent_cost_summary: {
          provenance: 'partial',
          measured_cost_usd: 10.0,
          estimated_cost_usd: 4.0,
          total_equivalent_cost_usd: null,
          known_subtotal_usd: 14.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 3,
        },
      };
      expect(() => validateTimeBucketCosts(bucket)).not.toThrow();
      expect(() => validateTimeBucketCosts({ ...bucket, cumulative_total_equivalent_cost_usd: 14.0 })).toThrow(/Invalid cost payload/);
    });

    it('validates comparison row equivalent cost fields', () => {
      const row = {
        cost_usd: null,
        cost_summary: { state: 'partial', known_subtotal_usd: 2.0, fully_observed: 1, partial: 0, unavailable: 1, total: 2 },
        measured_cost_usd: 2.0,
        estimated_cost_usd: 3.0,
        total_equivalent_cost_usd: 5.0,
        cost_provenance: 'hybrid',
        equivalent_cost_summary: {
          provenance: 'hybrid',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: 5.0,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        },
      };
      expect(() => validateComparisonRowCosts(row)).not.toThrow();
      expect(() => validateComparisonRowCosts({ ...row, estimated_cost_usd: 9.0 })).toThrow(/Invalid cost payload/);
    });
  });

  describe('Wire-level decodeDetail, decodeLedgers, and decodeFleet with AL2-6 payloads', () => {
    const baseDetail = mockWorkspaceDetails['AL-1']!;
    const estimatedEquiv = {
      provenance: 'estimated' as const,
      measured_cost_usd: null,
      estimated_cost_usd: 7.12,
      total_equivalent_cost_usd: 7.12,
      known_subtotal_usd: 7.12,
      measured: 0,
      estimated: 1,
      partial_measured: 0,
      unavailable: 0,
      total: 1,
    };
    const measuredEquiv = {
      provenance: 'measured' as const,
      measured_cost_usd: 23.2,
      estimated_cost_usd: null,
      total_equivalent_cost_usd: 23.2,
      known_subtotal_usd: 23.2,
      measured: 1,
      estimated: 0,
      partial_measured: 0,
      unavailable: 0,
      total: 1,
    };
    const hybridEquiv = {
      provenance: 'hybrid' as const,
      measured_cost_usd: 23.2,
      estimated_cost_usd: 7.12,
      total_equivalent_cost_usd: 30.32,
      known_subtotal_usd: 30.32,
      measured: 1,
      estimated: 1,
      partial_measured: 0,
      unavailable: 0,
      total: 2,
    };

    const makeAL26WireLedger = () => ({
      ...baseDetail.ledger,
      workspace_path: '/records/AL-1',
      facts: {
        ...baseDetail.ledger.facts,
        schema_version: 6,
        sessions: [
          {
            id: 'ses-oc-1',
            source: 'opencode',
            user: 'dev-a',
            invocations: [
              {
                role: 'orchestrator',
                model: 'claude-sonnet-4',
                started_at: '2026-10-15T10:05:00Z',
                completed_at: '2026-10-15T10:20:00Z',
                active_duration_seconds: 900,
                turns: 2,
                tokens: { input: 10000, output: 2000, cache_read: 5000, reasoning: 0, peak_context: 40000 },
                cost_usd: 23.2,
                cost_evidence: { state: 'observed', known_subtotal_usd: 23.2 },
                compactions: 0,
                user_iterations: 1,
              },
            ],
            tools: [],
          },
          {
            id: 'ses-codex-1',
            source: 'codex',
            user: 'dev-a',
            invocations: [
              {
                role: 'coder',
                model: 'gpt-5.4',
                started_at: '2026-10-15T10:25:00Z',
                completed_at: '2026-10-15T10:45:00Z',
                active_duration_seconds: 1200,
                turns: 4,
                tokens: { input: 200000, output: 50000, cache_read: 800000, reasoning: 10000, peak_context: 90000 },
                cost_usd: null,
                cost_evidence: { state: 'unavailable', known_subtotal_usd: null },
                cost_estimate: { amount_usd: 7.12, rate: validRate },
                compactions: 0,
                user_iterations: 1,
              },
            ],
            tools: [],
          },
        ],
      },
      totals: {
        ...baseDetail.ledger.totals,
        invocations: 2,
        cost_usd: null,
        cost_summary: { state: 'partial', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 1, total: 2 },
        measured_cost_usd: 23.2,
        estimated_cost_usd: 7.12,
        total_equivalent_cost_usd: 30.32,
        cost_provenance: 'hybrid',
        equivalent_cost_summary: hybridEquiv,
      },
      roles: [
        {
          ...baseDetail.ledger.roles[0]!,
          role: 'orchestrator',
          invocations: 1,
          cost_usd: 23.2,
          cost_summary: { state: 'complete', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 0, total: 1 },
          measured_cost_usd: 23.2,
          estimated_cost_usd: null,
          total_equivalent_cost_usd: 23.2,
          cost_provenance: 'measured',
          equivalent_cost_summary: measuredEquiv,
        },
        {
          ...baseDetail.ledger.roles[0]!,
          role: 'coder',
          invocations: 1,
          cost_usd: null,
          cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
          measured_cost_usd: null,
          estimated_cost_usd: 7.12,
          total_equivalent_cost_usd: 7.12,
          cost_provenance: 'estimated',
          equivalent_cost_summary: estimatedEquiv,
        },
      ],
      models: [
        {
          ...baseDetail.ledger.models[0]!,
          model: 'claude-sonnet-4',
          invocations: 1,
          cost_usd: 23.2,
          cost_summary: { state: 'complete', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 0, total: 1 },
          measured_cost_usd: 23.2,
          estimated_cost_usd: null,
          total_equivalent_cost_usd: 23.2,
          cost_provenance: 'measured',
          equivalent_cost_summary: measuredEquiv,
        },
        {
          ...baseDetail.ledger.models[0]!,
          model: 'gpt-5.4',
          invocations: 1,
          cost_usd: null,
          cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
          measured_cost_usd: null,
          estimated_cost_usd: 7.12,
          total_equivalent_cost_usd: 7.12,
          cost_provenance: 'estimated',
          equivalent_cost_summary: estimatedEquiv,
          applied_rates: [validRate],
        },
      ],
      threads: [
        {
          session_id: 'ses-codex-1',
          user: 'dev-a',
          compactions: 0,
          roles: [
            {
              role: 'coder',
              models: [
                {
                  ...baseDetail.ledger.models[0]!,
                  model: 'gpt-5.4',
                  invocations: 1,
                  cost_usd: null,
                  cost_summary: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
                  measured_cost_usd: null,
                  estimated_cost_usd: 7.12,
                  total_equivalent_cost_usd: 7.12,
                  cost_provenance: 'estimated',
                  equivalent_cost_summary: estimatedEquiv,
                  applied_rates: [validRate],
                },
              ],
            },
          ],
        },
      ],
    });

    it('decodes AL2-6 wire payloads through decodeDetail, decodeLedgers, and decodeFleet', () => {
      const wireLedger = makeAL26WireLedger();
      const wireDetail = {
        ...baseDetail,
        workspace_path: '/records/AL-1',
        ledger: wireLedger,
      };

      const decodedDetail = decodeDetail(JSON.parse(JSON.stringify(wireDetail)));
      expect(decodedDetail.ledger.totals.cost_provenance).toBe('hybrid');
      expect(decodedDetail.ledger.totals.total_equivalent_cost_usd).toBe(30.32);
      expect(decodedDetail.ledger.facts.sessions[1]?.invocations[0]?.cost_estimate).toEqual({
        amount_usd: 7.12,
        rate: validRate,
      });
      expect(decodedDetail.ledger.models[1]?.applied_rates).toEqual([validRate]);
      expect(decodedDetail.ledger.threads[0]?.roles[0]?.models[0]?.applied_rates).toEqual([validRate]);

      const decodedLedgers = decodeLedgers(JSON.parse(JSON.stringify([wireLedger])));
      expect(decodedLedgers).toHaveLength(1);
      expect(decodedLedgers[0]?.totals.equivalent_cost_summary?.provenance).toBe('hybrid');
      expect(decodedLedgers[0]?.models[1]?.applied_rates).toEqual([validRate]);

      const wireFleet = {
        ...mockFleetSummary,
        total_cost_usd: null,
        cost_summary: { state: 'partial', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 1, total: 2 },
        measured_cost_usd: 23.2,
        estimated_cost_usd: 7.12,
        total_equivalent_cost_usd: 30.32,
        cost_provenance: 'hybrid',
        equivalent_cost_summary: hybridEquiv,
        role_costs: { orchestrator: 23.2, coder: null },
        role_cost_summaries: {
          orchestrator: { state: 'complete', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 0, total: 1 },
          coder: { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        },
        role_equivalent_cost_summaries: {
          orchestrator: measuredEquiv,
          coder: estimatedEquiv,
        },
        model_costs: { 'claude-sonnet-4': 23.2, 'gpt-5.4': null },
        model_cost_summaries: {
          'claude-sonnet-4': { state: 'complete', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 0, total: 1 },
          'gpt-5.4': { state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 },
        },
        model_equivalent_cost_summaries: {
          'claude-sonnet-4': measuredEquiv,
          'gpt-5.4': estimatedEquiv,
        },
      };
      const decodedFleet = decodeFleet(JSON.parse(JSON.stringify(wireFleet)));
      expect(decodedFleet.cost_provenance).toBe('hybrid');
      expect(decodedFleet.total_equivalent_cost_usd).toBe(30.32);
      expect(decodedFleet.role_equivalent_cost_summaries?.coder?.provenance).toBe('estimated');
      expect(decodedFleet.model_equivalent_cost_summaries?.['gpt-5.4']?.estimated_cost_usd).toBe(7.12);

      // Malformed wire payload fails decodeDetail / decodeLedgers / decodeFleet
      const badWireDetail = JSON.parse(JSON.stringify(wireDetail));
      badWireDetail.ledger.facts.sessions[1].invocations[0].cost_estimate.amount_usd = -5;
      expect(() => decodeDetail(badWireDetail)).toThrow(/Invalid cost payload/);

      const badWireFleet = JSON.parse(JSON.stringify(wireFleet));
      delete badWireFleet.model_equivalent_cost_summaries['gpt-5.4'];
      expect(() => decodeFleet(badWireFleet)).toThrow(/Invalid cost payload/);
    });
  });
});
