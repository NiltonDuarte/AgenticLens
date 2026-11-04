import { describe, expect, it } from 'vitest';
import { deriveSessionOverlaps } from './sessionOverlaps';
import type { SessionFact } from './types';

function record(ticketId: string | null, sessions: readonly SessionFact[], title: string | null = null) {
  return { ticket_id: ticketId, ticket_title: title, facts: { sessions } };
}

const shared: SessionFact = { id: 'native-shared-id', source: 'codex', invocations: [], tools: [] };

describe('recorded session overlap projection', () => {
  it('indexes exact IDs across distinct tickets and merges context without mutating repeated input', () => {
    const ledgers = [
      record('TASK-2', [shared, shared, { ...shared, id: 'another-native-id' }], 'Second'),
      record('TASK-1', [{ ...shared, source: 'opencode' }, { ...shared, id: 'another-native-id', source: 'opencode' }], 'First'),
      record('TASK-2', [{ ...shared, source: 'opencode' }], 'Other title'),
      record('TASK-3', [{ id: shared.id, invocations: [], tools: [] }]),
    ];
    const before = JSON.stringify(ledgers);
    expect(deriveSessionOverlaps(ledgers)).toEqual({
      conflicts: [{ sessionId: 'another-native-id', tickets: [
        { ticketId: 'TASK-1', titles: ['First'], sources: ['opencode'] },
        { ticketId: 'TASK-2', titles: ['Second'], sources: ['codex'] },
      ] }, { sessionId: shared.id, tickets: [
        { ticketId: 'TASK-1', titles: ['First'], sources: ['opencode'] },
        { ticketId: 'TASK-2', titles: ['Other title', 'Second'], sources: ['codex', 'opencode'] },
        { ticketId: 'TASK-3', titles: [], sources: ['Unknown source'] },
      ] }],
      inspectedTickets: 3, omittedTicketRecords: 0, omittedSessionRecords: 0,
    });
    expect(JSON.stringify(ledgers)).toBe(before);
  });

  it('does not invent overlap from same-ticket repetition, similar IDs, or absent identities', () => {
    expect(deriveSessionOverlaps([
      record('TASK-1', [shared, shared]), record('TASK-1', [shared]),
      record('TASK-2', [{ ...shared, id: `${shared.id} ` }]),
      record('TASK-3', [{ ...shared, id: ' ' }]),
      record(null, [shared]), record(' ', [shared]),
    ])).toEqual({ conflicts: [], inspectedTickets: 3, omittedTicketRecords: 2, omittedSessionRecords: 1 });
    expect(deriveSessionOverlaps([]).conflicts).toEqual([]);
    expect(deriveSessionOverlaps([record('TASK-1', [shared])]).conflicts).toEqual([]);
  });
});
