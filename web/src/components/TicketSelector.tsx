import React, { useMemo } from 'react';
import { compareTicketIds } from '../domain/formatters';

export interface TicketSelectorOption {
  readonly ticket_id: string;
  readonly ticket_title?: string | null;
  readonly context_health_status?: string;
}

export interface TicketSelectorProps {
  readonly tickets: readonly TicketSelectorOption[];
  readonly selectedTicketId: string | null;
  readonly onSelectTicket: (ticketId: string) => void;
}

export const TicketSelector: React.FC<TicketSelectorProps> = ({
  tickets,
  selectedTicketId,
  onSelectTicket,
}) => {
  const sortedTickets = useMemo(() => {
    return [...tickets].sort((a, b) => compareTicketIds(a.ticket_id, b.ticket_id));
  }, [tickets]);

  return (
    <div className="ticket-selector-bar">
      <label htmlFor="ticket-select" className="ticket-selector-label">
        Workspace Autopsy:
      </label>
      <div className="ticket-selector-control">
        <select
          id="ticket-select"
          className="ticket-selector-dropdown"
          value={selectedTicketId ?? ''}
          onChange={(e) => onSelectTicket(e.target.value)}
        >
          {sortedTickets.length === 0 && (
            <option value="" disabled>
              No workspaces available
            </option>
          )}
          {sortedTickets.map((t) => (
            <option key={t.ticket_id} value={t.ticket_id}>
              {t.ticket_id} - {t.ticket_title && t.ticket_title.trim().length > 0 ? t.ticket_title : 'Untitled'}
              {t.context_health_status ? ` [${t.context_health_status}]` : ''}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
};
