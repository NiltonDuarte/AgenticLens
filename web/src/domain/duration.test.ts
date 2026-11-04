// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { decodeLedgers } from '../api/validation';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import type { DurationEvidence } from './types';
import { durationDescription, durationLabel, emptyDurationSummary, invocationDuration, mergeDuration, validDurationEvidence } from './duration';

describe('duration coverage at the read boundary', () => {
  const cutoff = '2026-01-01T00:00:10Z';
  const evidence: DurationEvidence = { state: 'provisional', completed_seconds: 3, provisional_seconds: 7, cutoff, basis: 'codex:turn' };
  const base = Object.values(mockAgenticMetricsLedgers)[0]!;
  const invocation = base.facts.sessions.flatMap((s) => s.invocations)[0]!;
  const withEvidence = (value: unknown): unknown => ({ ...base, facts: { ...base.facts, sessions: [{ ...base.facts.sessions[0]!, invocations: [{ ...invocation, active_duration_seconds: 10, duration_evidence: value }] }] } });

  it('retains legacy unknown coverage and distinct recorded cutoffs without advancing them', () => {
    const legacy = invocationDuration(invocation);
    expect(legacy.unknown_invocations).toBe(1);
    const active = invocationDuration({ ...invocation, active_duration_seconds: 10, duration_evidence: evidence });
    const later = invocationDuration({ ...invocation, active_duration_seconds: 10, duration_evidence: { ...evidence, cutoff: '2026-01-01T00:00:20Z' } });
    const combined = mergeDuration(mergeDuration(legacy, active), later);
    expect(combined).toEqual({ provisional_seconds: 14, provisional_invocations: 2, unknown_invocations: 1, cutoffs: [cutoff, '2026-01-01T00:00:20Z'] });
    expect(active.cutoffs).toEqual([cutoff]);
    expect(durationLabel('20s', combined)).toContain('provisional');
    expect(durationDescription(combined)).toContain(cutoff);
    expect(durationDescription(combined)).toContain('unknown');
    expect(durationDescription(undefined)).not.toContain('Completed');
    expect(durationDescription(emptyDurationSummary())).toBe('Completed measurements');
  });

  it('preserves duration evidence through the actual ledger decoder', () => {
    const decoded = decodeLedgers([withEvidence(evidence)]);
    expect(decoded[0]!.facts.sessions[0]!.invocations[0]!.duration_evidence).toEqual(evidence);
    expect(validDurationEvidence(evidence, 10)).toBe(true);
    const legacy = decodeLedgers([base]);
    expect(legacy[0]!.facts.sessions.flatMap((s) => s.invocations)[0]!.active_duration_seconds).toBe(invocation.active_duration_seconds);
  });

  it.each([
    { ...evidence, completed_seconds: 4 },
    { ...evidence, provisional_seconds: -7 },
    { ...evidence, cutoff: 'not-a-date' },
    { ...evidence, cutoff: undefined },
    { ...evidence, basis: '' },
    { ...evidence, state: 'complete' },
    { ...evidence, state: 'invented' },
  ])('rejects inconsistent duration evidence through the ledger decoder: %j', (invalid) => {
    expect(validDurationEvidence(invalid, 10)).toBe(false);
    expect(() => decodeLedgers([withEvidence(invalid)])).toThrow();
  });
});
