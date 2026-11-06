// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { decodeLedgers, decodeFleet } from '../api/validation';
import { mockAgenticMetricsLedgers, mockFleetSummary } from '../test/fixtures/mockData';
import { durationDescription, durationLabel, validDuration } from './duration';

describe('numeric provisional duration at the read boundary', () => {
  const base = Object.values(mockAgenticMetricsLedgers)[0]!;
  const invocation = base.facts.sessions.flatMap((s) => s.invocations)[0]!;
  const withProvisional = (value: unknown): unknown => ({ ...base, facts: { ...base.facts, sessions: [{ ...base.facts.sessions[0]!, invocations: [{ ...invocation, active_duration_seconds: 10, provisional_seconds: value }] }] } });

  it('uses neutral recorded labels for absent and zero contributions', () => {
    for (const value of [undefined, 0]) {
      expect(durationLabel('10s', value)).toBe('10s');
      expect(durationDescription(value)).toBe('Recorded duration');
      expect(durationDescription(value)).not.toMatch(/complete|verified|coverage|cutoff/i);
    }
    expect(durationLabel('10s', 7)).toContain('7s provisional');
    expect(durationDescription(7)).toBe('Recorded duration including 7s provisional');
  });

  it.each([undefined, 0, 7, 10])('retains optional numeric contribution %s through the actual ledger decoder', (value) => {
    const decoded = decodeLedgers([withProvisional(value)]);
    const inv = decoded[0]!.facts.sessions[0]!.invocations[0]!;
    expect(inv.active_duration_seconds).toBe(10);
    expect(inv.provisional_seconds).toBe(value);
    expect(validDuration(10, value)).toBe(true);
    expect(decodeLedgers([base])[0]!.facts.sessions.flatMap((s) => s.invocations)[0]!.active_duration_seconds).toBe(invocation.active_duration_seconds);
  });

  it.each([-1, 1.5, 11, Number.MAX_SAFE_INTEGER + 1, null, '7', Infinity, NaN])('rejects inconsistent numeric duration through the ledger decoder: %s', (value) => {
    expect(validDuration(10, value)).toBe(false);
    expect(() => decodeLedgers([withProvisional(value)])).toThrow();
  });

  it('validates aggregate provisional duration against its recorded total', () => {
    const fleet = { ...mockFleetSummary, total_duration_seconds: 10, provisional_seconds: 7 };
    expect(decodeFleet(fleet).provisional_seconds).toBe(7);
    expect(() => decodeFleet({ ...fleet, provisional_seconds: 11 })).toThrow();
    expect(() => decodeFleet({ ...fleet, provisional_seconds: 0.5 })).toThrow();
  });
});
