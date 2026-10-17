import React, { useState, useMemo } from 'react';
import type { TicketComparisonRow } from '../domain/types';
import { formatCurrencyUSD, formatPercentage } from '../domain/formatters';
import { EmptyState } from './EmptyState';

export interface TicketTableProps {
  readonly rows: readonly TicketComparisonRow[];
  readonly onSelectTicket?: ((ticketId: string) => void) | undefined;
}

type SortField = 'ticket_id' | 'cost_usd' | 'active_duration_seconds' | 'sdlc_iterations';

export const TicketTable: React.FC<TicketTableProps> = ({
  rows,
  onSelectTicket,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState<SortField>('ticket_id');
  const [sortAscending, setSortAscending] = useState(true);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortAscending(!sortAscending);
    } else {
      setSortField(field);
      setSortAscending(true);
    }
  };

  const filteredAndSortedRows = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();
    const filtered = rows.filter((r) => {
      if (!term) return true;
      return (
        r.ticket_id.toLowerCase().includes(term) ||
        r.ticket_title.toLowerCase().includes(term)
      );
    });

    return [...filtered].sort((a, b) => {
      let diff = 0;
      switch (sortField) {
        case 'ticket_id':
          diff = a.ticket_id.localeCompare(b.ticket_id);
          break;
        case 'cost_usd':
          diff = a.cost_usd - b.cost_usd;
          break;
        case 'active_duration_seconds':
          diff = a.active_duration_seconds - b.active_duration_seconds;
          break;
        case 'sdlc_iterations':
          diff = a.sdlc_iterations - b.sdlc_iterations;
          break;
      }
      return sortAscending ? diff : -diff;
    });
  }, [rows, searchTerm, sortField, sortAscending]);

  const getSortIndicator = (field: SortField) => {
    if (sortField !== field) return '';
    return sortAscending ? ' (asc)' : ' (desc)';
  };

  return (
    <div className="ticket-section">
      <div className="ticket-header-bar">
        <h2 className="table-title">Ticket Workspaces Directory</h2>
        <input
          type="text"
          className="search-input"
          placeholder="Search ticket ID or title..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          aria-label="Filter tickets"
        />
      </div>

      {filteredAndSortedRows.length === 0 ? (
        <EmptyState
          title="No Workspaces Match Filter"
          message={
            searchTerm
              ? `No ticket records match query "${searchTerm}". Try a different search term.`
              : 'No ticket workspaces are currently registered in this session.'
          }
        />
      ) : (
        <div className="table-container">
          <table className="ticket-table">
            <thead>
              <tr>
                <th
                  className="sortable"
                  onClick={() => handleSort('ticket_id')}
                >
                  Ticket ID{getSortIndicator('ticket_id')}
                </th>
                <th>Title</th>
                <th>Status</th>
                <th
                  className="sortable"
                  onClick={() => handleSort('active_duration_seconds')}
                >
                  Active Duration{getSortIndicator('active_duration_seconds')}
                </th>
                <th
                  className="sortable"
                  onClick={() => handleSort('cost_usd')}
                >
                  Cost (USD){getSortIndicator('cost_usd')}
                </th>
                <th>Cache Hit Rate</th>
                <th>Code Delivered</th>
                <th
                  className="sortable"
                  onClick={() => handleSort('sdlc_iterations')}
                >
                  SDLC Iterations{getSortIndicator('sdlc_iterations')}
                </th>
              </tr>
            </thead>
            <tbody>
              {filteredAndSortedRows.map((r) => {
                const statusClass = ['HEALTHY', 'ELEVATED', 'BLOATED', 'CRITICAL'].includes(
                  r.context_health_status
                )
                  ? r.context_health_status
                  : 'HEALTHY';

                return (
                  <tr
                    key={r.ticket_id}
                    onClick={() => onSelectTicket?.(r.ticket_id)}
                    style={{ cursor: onSelectTicket ? 'pointer' : 'default' }}
                  >
                    <td className="mono-number" style={{ fontWeight: 600 }}>
                      {r.ticket_id}
                    </td>
                    <td>{r.ticket_title}</td>
                    <td>
                      <span className={`health-badge ${statusClass}`}>
                        {r.context_health_status}
                      </span>
                    </td>
                    <td className="mono-number">{r.active_duration_formatted}</td>
                    <td className="mono-number">{formatCurrencyUSD(r.cost_usd)}</td>
                    <td className="mono-number">{formatPercentage(r.cache_hit_rate_pct)}</td>
                    <td className="mono-number">
                      +{r.lines_added} / -{r.lines_deleted} ({r.files_changed} files)
                    </td>
                    <td className="mono-number">{r.sdlc_iterations}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
