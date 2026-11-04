// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { skillNames, skillSuccessRate } from './aggregations';
import { ledgerSkills, ledgerTaskLinks, type SkillFact, type SessionFact } from './types';

const skill = (name: string): SkillFact => ({ name, success: true, outcome: 'succeeded' as const, turn_index: 0, message_index: 0, caller_role: 'coder', timestamp: '2026-01-01T00:00:00Z' });

describe('recorded skill analytics', () => {
  it('deduplicates and orders skill names by Unicode code point without mutating facts', () => {
    const skills = [skill('\u{10000}'), skill('\uE000'), skill('alpha'), skill('alpha')];
    const before = JSON.stringify(skills);
    expect(skillNames({ sessions: [{ id: 'owner', invocations: [], tools: [], skills }] })).toEqual(['alpha', '\uE000', '\u{10000}']);
    expect(JSON.stringify(skills)).toBe(before);
    expect(skillNames({ sessions: [] })).toEqual([]);
    expect(skillNames({ sessions: [{ id: 'empty', invocations: [], tools: [] }] })).toEqual([]);
  });
  it('weights success by invocation count rather than averaging skill rates', () => {
    expect(skillSuccessRate([
      { name: 'alpha', total_invocations: 3, known_outcomes: 3, unknown_outcomes: 0, tickets_count: 2, failure_count: 1, failure_rate: 1 / 3 },
      { name: 'beta', total_invocations: 1, known_outcomes: 1, unknown_outcomes: 0, tickets_count: 1, failure_count: 1, failure_rate: 1 },
    ])).toBe(50);
    expect(skillSuccessRate([])).toBeNull();
    expect(skillSuccessRate([{ name: 'empty', total_invocations: 0, known_outcomes: 0, unknown_outcomes: 0, tickets_count: 0, failure_count: 0, failure_rate: 0 }])).toBeNull();
  });
  it('projects canonical owners once, preserves repeated occurrences and scopes exactly by session', () => {
    const repeated = skill('repeated');
    const task = { subagent_type: 'tester', outcome: 'unknown' as const, description: 'Verify', parent_turn: 0, timestamp: repeated.timestamp };
    const sessions: readonly SessionFact[] = [
      { id: 'root', source: 'codex', invocations: [], tools: [], skills: [repeated, repeated], task_links: [task] },
      { id: 'child', parent_id: 'root', source: 'opencode', invocations: [], tools: [], skills: [repeated], task_links: [task] },
    ];
    const ledger = { sessions };
    expect(ledgerSkills(ledger).map((fact) => fact.session_id)).toEqual(['root', 'root', 'child']);
    expect(ledgerTaskLinks(ledger).map((fact) => fact.session_id)).toEqual(['root', 'child']);
    expect(ledgerSkills(ledger, 'root')).toHaveLength(2);
    expect(ledgerSkills(ledger, 'child')).toHaveLength(1);
    expect(ledgerSkills(ledger, 'missing')).toEqual([]);
    expect(ledgerTaskLinks(ledger, 'child')).toEqual([{ ...task, session_id: 'child', source: 'opencode' }]);
    expect(skillNames(ledger)).toEqual(['repeated']);
    expect(ledgerSkills({ sessions: [] })).toEqual([]);
    expect(ledgerSkills({ sessions: [] }, 'root')).toEqual([]);
  });

  it('excludes unknown activation requests from the measured success denominator', () => {
    expect(skillSuccessRate([{ name: 'requested', total_invocations: 3, known_outcomes: 0, unknown_outcomes: 3,
      tickets_count: 1, failure_count: 0, failure_rate: 0 }])).toBeNull();
    expect(skillSuccessRate([{ name: 'mixed', total_invocations: 4, known_outcomes: 2, unknown_outcomes: 2,
      tickets_count: 1, failure_count: 1, failure_rate: 0.5 }])).toBe(50);
  });

});
