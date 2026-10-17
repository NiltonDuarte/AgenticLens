import React from 'react';

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
          {tickets.length === 0 && (
            <option value="" disabled>
              No workspaces available
            </option>
          )}
          {tickets.map((t) => (
            <option key={t.ticket_id} value={t.ticket_id}>
              {t.ticket_id} - {t.ticket_title ?? 'Untitled'}
              {t.context_health_status ? ` [${t.context_health_status}]` : ''}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
};
