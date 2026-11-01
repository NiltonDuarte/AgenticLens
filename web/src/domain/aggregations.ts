import { ledgerSkills, type AgenticFactLedger, type SkillAdoptionFact } from './types';

// Compare Unicode code points to match Go string ordering, including non-ASCII names.
export function compareSkillNames(a: string, b: string): number {
  const left = Array.from(a, (character) => character.codePointAt(0)!);
  const right = Array.from(b, (character) => character.codePointAt(0)!);
  for (let i = 0; i < Math.min(left.length, right.length); i++) {
    if (left[i] !== right[i]) return left[i]! - right[i]!;
  }
  return left.length - right.length;
}

export function skillNames(ledger: Pick<AgenticFactLedger, 'skills'>): string[] {
  return [...new Set(ledgerSkills(ledger).map((skill) => skill.name))].sort(compareSkillNames);
}

export function skillSuccessRate(adoption: readonly SkillAdoptionFact[]): number | null {
  const invocations = adoption.reduce((total, skill) => total + skill.total_invocations, 0);
  const failures = adoption.reduce((total, skill) => total + skill.failure_count, 0);
  return invocations === 0 ? null : (invocations - failures) / invocations * 100;
}
