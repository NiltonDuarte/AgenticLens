// @vitest-environment node
import { describe, expect, it } from 'vitest';
import { skillNames } from './aggregations';
import { ledgerSkills, ledgerDelegations, type SkillFact, type SessionFact } from './types';

const skill = (name: string): SkillFact => ({ name, caller_role: 'coder', invoked_at: '2026-01-01T00:00:00Z' });

describe('recorded skill analytics', () => {
  it('deduplicates and orders skill names by Unicode code point without mutating facts', () => {
    const skills = [skill('\u{10000}'), skill('\uE000'), skill('alpha'), skill('alpha')];
    const before = JSON.stringify(skills);
    expect(skillNames({ sessions: [{ id: 'owner', invocations: [], tools: [], skills }] })).toEqual(['alpha', '\uE000', '\u{10000}']);
    expect(JSON.stringify(skills)).toBe(before);
    expect(skillNames({ sessions: [] })).toEqual([]);
    expect(skillNames({ sessions: [{ id: 'empty', invocations: [], tools: [] }] })).toEqual([]);
  });
  it('projects canonical owners once, preserves repeated occurrences and scopes exactly by session', () => {
    const repeated = skill('repeated');
    const task = { agent_type: 'tester', description: 'Verify', invoked_at: repeated.invoked_at };
    const sessions: readonly SessionFact[] = [
      { id: 'root', source: 'codex', invocations: [], tools: [], skills: [repeated, repeated], delegations: [task] },
      { id: 'child', parent_id: 'root', source: 'opencode', invocations: [], tools: [], skills: [repeated], delegations: [task] },
    ];
    const ledger = { sessions };
    expect(ledgerSkills(ledger).map((fact) => fact.session_id)).toEqual(['root', 'root', 'child']);
    expect(ledgerDelegations(ledger).map((fact) => fact.session_id)).toEqual(['root', 'child']);
    expect(ledgerSkills(ledger, 'root')).toHaveLength(2);
    expect(ledgerSkills(ledger, 'child')).toHaveLength(1);
    expect(ledgerSkills(ledger, 'missing')).toEqual([]);
    expect(ledgerDelegations(ledger, 'child')).toEqual([{ ...task, session_id: 'child' }]);
    expect(skillNames(ledger)).toEqual(['repeated']);
    expect(ledgerSkills({ sessions: [] })).toEqual([]);
    expect(ledgerSkills({ sessions: [] }, 'root')).toEqual([]);
  });

});
