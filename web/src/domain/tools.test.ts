import { describe, expect, it } from 'vitest';
import { aggregateToolUsage } from './tools';

describe('aggregateToolUsage', () => {
  it('groups session-scoped facts by tool name and sums counts', () => {
    expect(aggregateToolUsage([
      { session_id: 'session-a', name: 'bash', calls: 4, errors: 1 },
      { session_id: 'session-b', name: 'bash', calls: 6, errors: 2 },
      { session_id: 'session-a', name: 'read', calls: 3, errors: 0 },
    ])).toEqual([
      { name: 'bash', calls: 10, errors: 3 },
      { name: 'read', calls: 3, errors: 0 },
    ]);
  });

  it('returns an empty summary for no facts', () => {
    expect(aggregateToolUsage([])).toEqual([]);
  });
});
