import type { ToolFact } from './types';

export interface ToolUsageSummary {
  readonly name: string;
  readonly calls: number;
  readonly errors: number;
}

export function aggregateToolUsage(facts: readonly ToolFact[]): ToolUsageSummary[] {
  const byName = new Map<string, ToolUsageSummary>();
  for (const fact of facts) {
    const current = byName.get(fact.name);
    byName.set(fact.name, {
      name: fact.name,
      calls: (current?.calls ?? 0) + fact.calls,
      errors: (current?.errors ?? 0) + fact.errors,
    });
  }
  return [...byName.values()].sort((left, right) => left.name.localeCompare(right.name));
}
