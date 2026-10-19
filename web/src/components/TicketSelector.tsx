import React, { useState, useMemo } from 'react';
import type { AutopsySortMode } from '../domain/types';
import { compareTicketIds, compareTicketCompletionDate } from '../domain/formatters';

export interface TicketSelectorOption {
  readonly ticket_id: string;
  readonly ticket_title?: string | null;
  readonly context_health_status?: string;
  readonly completed_at?: string | null;
  readonly updated_at?: string;
  readonly started_at?: string | null;
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
  const [sortMode, setSortMode] = useState<AutopsySortMode>('ticket_id');

  const sortedTickets = useMemo(() => {
    if (sortMode === 'completed_at') {
      return [...tickets].sort(compareTicketCompletionDate);
    }
    return [...tickets].sort((a, b) => compareTicketIds(a.ticket_id, b.ticket_id));
  }, [tickets, sortMode]);

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
      <div className="autopsy-sort-control">
        <label htmlFor="autopsy-sort-select" className="autopsy-sort-label">
          Sort:
        </label>
        <select
          id="autopsy-sort-select"
          className="autopsy-sort-dropdown"
          value={sortMode}
          onChange={(e) => setSortMode(e.target.value as AutopsySortMode)}
          aria-label="Sort Workspaces"
        >
          <option value="ticket_id">Ticket ID</option>
          <option value="completed_at">Completion Date</option>
        </select>
      </div>
    </div>
  );
};
