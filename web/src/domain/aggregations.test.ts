// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { computeSkillAdoption, skillNames, skillSuccessRate } from './aggregations';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import type { AgenticFactLedger, SkillFact } from './types';

const base = Object.values(mockAgenticMetricsLedgers)[0]!.facts;
const skill = (name: string, success = true): SkillFact => ({ name, success, turn_index: 0, message_index: 0, caller_role: 'coder', timestamp: '2026-01-01T00:00:00Z' });

describe('recorded skill analytics', () => {
  it('retains repeated calls, distinct ticket counts, failure booleans, and ranking ties', () => {
    const ledgers: AgenticFactLedger[] = [
      { ...base, ticket_id: 'T-1', skills: [skill('beta'), skill('alpha', false), { ...skill('beta'), error: 'inspection text' }] },
      { ...base, ticket_id: 'T-2', skills: [skill('beta', false), skill('alpha')] },
      { ...base, ticket_id: 'T-1', skills: [skill('alpha')] },
    ];
    const before = JSON.stringify(ledgers);
    const adoption = computeSkillAdoption(ledgers);
    expect(adoption).toEqual([
      { name: 'alpha', total_invocations: 3, tickets_count: 2, failure_count: 1, failure_rate: 1 / 3 },
      { name: 'beta', total_invocations: 3, tickets_count: 2, failure_count: 1, failure_rate: 1 / 3 },
    ]);
    expect(skillSuccessRate(adoption)).toBeCloseTo(200 / 3);
    expect(computeSkillAdoption([{ ...base, skills: [skill('a'), skill('b'), skill('b'), skill('c')] }]).map((fact) => fact.name)).toEqual(['b', 'a', 'c']);
    expect(skillNames({ skills: [skill('\u{10000}'), skill('\uE000')] })).toEqual(['\uE000', '\u{10000}']);
    expect(skillNames(ledgers[0]!)).toEqual(['alpha', 'beta']);
    expect(JSON.stringify(ledgers)).toBe(before);
  });
  it('handles missing data and distinguishes unidentified ledger entries', () => {
    const { skills: omitted, ...withoutSkills } = base;
    void omitted;
    expect(computeSkillAdoption([])).toEqual([]);
    expect(computeSkillAdoption([withoutSkills, { ...base, skills: [] }])).toEqual([]);
    expect(skillSuccessRate([])).toBeNull();
    expect(computeSkillAdoption([
      { ...base, ticket_id: null, skills: [skill('x')] },
      { ...base, ticket_id: '', skills: [skill('x')] },
    ])[0]?.tickets_count).toBe(2);
  });
});
