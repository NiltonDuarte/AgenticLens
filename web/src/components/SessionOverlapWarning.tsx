import React from 'react';
import type { SessionOverlapProjection } from '../domain/sessionOverlaps';
import type { WorkspaceDiagnostic } from '../domain/types';
import { WorkspaceLoadWarning } from './WorkspaceLoadWarning';

export type SessionOverlapInspection =
  | { readonly state: 'loading' }
  | { readonly state: 'unavailable'; readonly message: string }
  | { readonly state: 'ready'; readonly projection: SessionOverlapProjection; readonly diagnostics: readonly WorkspaceDiagnostic[] };

export interface SessionOverlapWarningProps {
  readonly inspection: SessionOverlapInspection;
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
}

export const SessionOverlapWarning: React.FC<SessionOverlapWarningProps> = ({ inspection, onSelectTicket }) => {
  if (inspection.state === 'loading') {
    return <p role="status">Inspecting loaded project ledgers for shared sessions...</p>;
  }
  if (inspection.state === 'unavailable') {
    return (
      <section className="session-overlap-warning" aria-label="Session overlap inspection" role="status">
        <strong>Session overlap inspection unavailable.</strong>
        <p>{inspection.message} Use Refresh Data to retry. Overlap outside inspected records is unknown.</p>
      </section>
    );
  }

  const { projection, diagnostics } = inspection;
  const partial = diagnostics.length > 0 || projection.omittedTicketRecords > 0 || projection.omittedSessionRecords > 0;
  if (projection.conflicts.length === 0 && !partial) return null;

  return (
    <section className="session-overlap-warning" aria-label="Session overlap inspection">
      <p role="status">
        <strong>{projection.conflicts.length > 0
          ? `${projection.conflicts.length} shared session${projection.conflicts.length === 1 ? '' : 's'} found across tickets.`
          : 'Session overlap inspection is partial.'}</strong>
      </p>
      <p>
        Inspected {projection.inspectedTickets} distinct tickets from loaded project ledgers, independently of dashboard filters.
        {partial ? ' Scope is partial; overlap in records not inspected is unknown.' : ' Scope covers the successfully loaded project ledgers.'}
      </p>
      <WorkspaceLoadWarning diagnostics={diagnostics} />
      {(projection.omittedTicketRecords > 0 || projection.omittedSessionRecords > 0) && (
        <p>{projection.omittedTicketRecords} ticket records and {projection.omittedSessionRecords} session records lacked usable IDs and were omitted.</p>
      )}
      {projection.conflicts.length > 0 && (
        <>
          <p>Metrics may overlap. Review the affected tickets and decide how to correct session attribution in your records.</p>
          <details>
            <summary>View shared sessions and affected tickets</summary>
            <ul className="session-overlap-list">
              {projection.conflicts.map(({ sessionId, tickets }) => (
                <li key={sessionId}>
                  <p>Session <code>{sessionId}</code></p>
                  <ul>
                    {tickets.map(({ ticketId, titles, sources }) => (
                      <li key={ticketId}>
                        {onSelectTicket ? (
                          <button type="button" className="btn-secondary" onClick={() => onSelectTicket(ticketId)} aria-label={`View ticket ${ticketId}`}>
                            {ticketId}{titles.length > 0 ? `: ${titles.join(' / ')}` : ''}
                          </button>
                        ) : <span>{ticketId}{titles.length > 0 ? `: ${titles.join(' / ')}` : ''}</span>}
                        <span> Sources: {sources.join(', ')}</span>
                      </li>
                    ))}
                  </ul>
                </li>
              ))}
            </ul>
          </details>
        </>
      )}
    </section>
  );
};
