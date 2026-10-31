// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { skillNames, skillSuccessRate } from './aggregations';
import type { SkillFact } from './types';

const skill = (name: string): SkillFact => ({ name, success: true, turn_index: 0, message_index: 0, caller_role: 'coder', timestamp: '2026-01-01T00:00:00Z' });

describe('recorded skill analytics', () => {
  it('deduplicates and orders skill names by Unicode code point without mutating facts', () => {
    const skills = [skill('\u{10000}'), skill('\uE000'), skill('alpha'), skill('alpha')];
    const before = JSON.stringify(skills);
    expect(skillNames({ skills })).toEqual(['alpha', '\uE000', '\u{10000}']);
    expect(JSON.stringify(skills)).toBe(before);
    expect(skillNames({})).toEqual([]);
    expect(skillNames({ skills: [] })).toEqual([]);
  });
  it('weights success by invocation count rather than averaging skill rates', () => {
    expect(skillSuccessRate([
      { name: 'alpha', total_invocations: 3, tickets_count: 2, failure_count: 1, failure_rate: 1 / 3 },
      { name: 'beta', total_invocations: 1, tickets_count: 1, failure_count: 1, failure_rate: 1 },
    ])).toBe(50);
    expect(skillSuccessRate([])).toBeNull();
    expect(skillSuccessRate([{ name: 'empty', total_invocations: 0, tickets_count: 0, failure_count: 0, failure_rate: 0 }])).toBeNull();
  });
});
