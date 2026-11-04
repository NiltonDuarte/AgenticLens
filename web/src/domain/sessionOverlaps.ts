import type { AgenticFactLedger, AgenticMetricsLedger } from './types';

type SessionOverlapLedger = Pick<AgenticMetricsLedger, 'ticket_id' | 'ticket_title'> & {
  readonly facts: Pick<AgenticFactLedger, 'sessions'>;
};

export interface SharedSessionTicket {
  readonly ticketId: string;
  readonly titles: readonly string[];
  readonly sources: readonly string[];
}

export interface SharedSession {
  readonly sessionId: string;
  readonly tickets: readonly SharedSessionTicket[];
}

export interface SessionOverlapProjection {
  readonly conflicts: readonly SharedSession[];
  readonly inspectedTickets: number;
  readonly omittedTicketRecords: number;
  readonly omittedSessionRecords: number;
}

/** Inspect recorded identities only; source labels do not participate in matching. */
export function deriveSessionOverlaps(ledgers: readonly SessionOverlapLedger[]): SessionOverlapProjection {
  const sessions = new Map<string, Map<string, { titles: Set<string>; sources: Set<string> }>>();
  const ticketIds = new Set<string>();
  let omittedTicketRecords = 0;
  let omittedSessionRecords = 0;

  for (const ledger of ledgers) {
    const ticketId = ledger.ticket_id;
    if (ticketId === null || ticketId.trim() === '') {
      omittedTicketRecords++;
      continue;
    }
    ticketIds.add(ticketId);
    for (const session of ledger.facts.sessions) {
      if (session.id.trim() === '') {
        omittedSessionRecords++;
        continue;
      }
      let tickets = sessions.get(session.id);
      if (!tickets) {
        tickets = new Map();
        sessions.set(session.id, tickets);
      }
      let context = tickets.get(ticketId);
      if (!context) {
        context = { titles: new Set(), sources: new Set() };
        tickets.set(ticketId, context);
      }
      if (ledger.ticket_title?.trim()) context.titles.add(ledger.ticket_title);
      context.sources.add(session.source ?? 'Unknown source');
    }
  }

  const conflicts: SharedSession[] = [];
  for (const [sessionId, tickets] of sessions) {
    if (tickets.size < 2) continue;
    conflicts.push({
      sessionId,
      tickets: [...tickets].sort(([left], [right]) => left.localeCompare(right)).map(([ticketId, context]) => ({
        ticketId,
        titles: [...context.titles].sort(),
        sources: [...context.sources].sort(),
      })),
    });
  }
  conflicts.sort((left, right) => left.sessionId.localeCompare(right.sessionId));
  return { conflicts, inspectedTickets: ticketIds.size, omittedTicketRecords, omittedSessionRecords };
}
