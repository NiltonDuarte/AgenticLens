import { describe, expect, it } from 'vitest';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import { aggregateMetricsOverTime } from './progression';
import type { AgenticMetricsLedger, InvocationFact } from './types';

describe('Session-owned invocation rows', () => {
  it.each(['hourly', 'daily', 'weekly'] as const)('counts repeated rows across owners in %s progression', (granularity) => {
    const fixture = mockAgenticMetricsLedgers['AL-1'];
    if (!fixture) throw new Error('Required AL-1 fixture is missing');
    const row: InvocationFact = {
      role: 'coder',
      model: 'shared-model',
      started_at: '2026-10-28T10:00:00Z',
      completed_at: '2026-10-28T10:01:00Z',
      active_duration_seconds: 20,
      turns: 1,
      tokens: { input: 10, output: 4, cache_read: 2, reasoning: 1, peak_context: 500 },
      cost_usd: 0.25,
      compactions: 0,
      user_iterations: 0,
    };
    const ledger: AgenticMetricsLedger = {
      ...fixture,
      completed_at: row.completed_at,
      facts: {
        ...fixture.facts,
        sessions: [
          { id: 'opaque-root', source: 'codex', user: 'alice', invocations: [row, row, { ...row, model: 'other-model' }], tools: [] },
          { id: 'opaque-child', parent_id: 'opaque-root', source: 'codex', user: 'bob', invocations: [row, row], tools: [] },
          { id: 'other-root', source: 'opencode', user: 'carol', invocations: [row], tools: [] },
        ],
      },
    };
    const before = JSON.stringify(ledger);

    const buckets = aggregateMetricsOverTime([ledger], granularity);

    expect(buckets).toHaveLength(1);
    expect(buckets[0]).toMatchObject({
      invocations: 6,
      cumulative_invocations: 6,
      input_tokens: 60,
      output_tokens: 24,
      cache_read_tokens: 12,
      reasoning_tokens: 6,
      cost_usd: 1.5,
      cumulative_cost_usd: 1.5,
      active_duration_seconds: 120,
      peak_context_tokens: 500,
      lines_added: fixture.workflow_efficiency.code_delivered.lines_added,
    });
    expect(JSON.stringify(ledger)).toBe(before);
  });
});
