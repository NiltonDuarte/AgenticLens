import { selectWorkspace } from '../domain/workspaces';
import { useProjectConfig, ticketPolicy } from '../contexts/ProjectConfigContext';
import React, { useState, useMemo } from 'react';
import type { AutopsySortMode, ContextHealthStatus } from '../domain/types';
import { compareTicketIds, compareTicketCompletionDate } from '../domain/formatters';

export interface TicketSelectorOption {
  readonly ticket_id: string;
  readonly workspace_path?: string;
  readonly ticket_title?: string | null;
  readonly context_health_status?: ContextHealthStatus;
  readonly completed_at?: string | null;
  readonly updated_at?: string;
  readonly started_at?: string | null;
}

export interface TicketSelectorProps {
  readonly tickets: readonly TicketSelectorOption[];
  readonly selectedTicketId: string | null;
  readonly selectedWorkspacePath?: string | undefined;
  readonly onSelectTicket: (ticketId: string, workspacePath?: string) => void;
}

export const TicketSelector: React.FC<TicketSelectorProps> = ({
  tickets,
  selectedTicketId,
  selectedWorkspacePath,
  onSelectTicket,
}) => {
  const projectConfig = useProjectConfig();
  const recognitionPolicy = useMemo(() => ticketPolicy(projectConfig), [projectConfig]);
  const [sortMode, setSortMode] = useState<AutopsySortMode>('ticket_id');

  const sortedTickets = useMemo(() => {
    if (sortMode === 'completed_at') {
      return [...tickets].sort((a, b) => compareTicketCompletionDate(a, b, recognitionPolicy));
    }
    return [...tickets].sort((a, b) => compareTicketIds(a.ticket_id, b.ticket_id, recognitionPolicy));
  }, [tickets, sortMode, recognitionPolicy]);

  const matchingTickets = tickets.filter((ticket) => ticket.ticket_id === selectedTicketId);
  const selectedValue = selectedWorkspacePath ?? (matchingTickets.length === 1
    ? matchingTickets[0]?.workspace_path ?? selectedTicketId ?? '' : '');

  return (
    <div className="ticket-selector-bar">
      <label htmlFor="ticket-select" className="ticket-selector-label">
        Workspace Autopsy:
      </label>
      <div className="ticket-selector-control">
        <select
          id="ticket-select"
          className="ticket-selector-dropdown"
          value={selectedValue}
          onChange={(e) => {
            const ticket = tickets.find((item) => (item.workspace_path ?? item.ticket_id) === e.target.value);
            if (ticket) selectWorkspace(onSelectTicket, ticket);
          }}
        >
          {sortedTickets.length > 0 && selectedValue === '' && <option value="" disabled>Select a workspace location</option>}
          {selectedWorkspacePath && !tickets.some((item) => item.workspace_path === selectedWorkspacePath) &&
            <option value={selectedWorkspacePath} disabled>{selectedTicketId} — unavailable location: {selectedWorkspacePath}</option>}
          {sortedTickets.length === 0 && (
            <option value="" disabled>
              No workspaces available
            </option>
          )}
          {sortedTickets.map((t) => (
            <option key={t.workspace_path ?? t.ticket_id} value={t.workspace_path ?? t.ticket_id}>
              {t.ticket_id} - {t.ticket_title && t.ticket_title.trim().length > 0 ? t.ticket_title : 'Untitled'}
              {t.context_health_status ? ` [${t.context_health_status}]` : ''}
              {t.workspace_path ? ` — ${t.workspace_path}` : ''}
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
