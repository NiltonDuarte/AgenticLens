import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { TicketTable } from './TicketTable';
import type { TicketComparisonRow } from '../domain/types';

describe('TicketTable Component', () => {
  const sampleRows: readonly TicketComparisonRow[] = [
    {
      ticket_id: 'AL-10',
      ticket_title: 'Zeta feature optimization',
      ticket_type: 'feat',
      updated_at: '2026-10-15T14:32:00Z',
      sdlc_iterations: 5,
      active_duration_seconds: 3600,
      active_duration_formatted: '1h 0m',
      cost_usd: 12.5,
      cache_hit_rate_pct: 80.0,
      test_to_code_ratio: '1.5:1',
      lines_added: 500,
      lines_deleted: 100,
      files_changed: 8,
      tool_success_rate_pct: 95.0,
      context_health_status: 'HEALTHY',
      root_saturation_pct: 40.0,
      peak_context_tokens: 80000,
      started_at: '2026-10-15T10:00:00Z',
      completed_at: '2026-10-15T11:00:00Z',
      lead_duration_seconds: 3600,
      lead_duration_formatted: '1h 0m',
    },
    {
      ticket_id: 'AL-2',
      ticket_title: 'Alpha backend refactor',
      ticket_type: 'refactor',
      updated_at: '2026-10-16T16:15:00Z',
      sdlc_iterations: 1,
      active_duration_seconds: 1200,
      active_duration_formatted: '20m 0s',
      cost_usd: 2.2,
      cache_hit_rate_pct: 90.0,
      test_to_code_ratio: '2.0:1',
      lines_added: 200,
      lines_deleted: 50,
      files_changed: 3,
      tool_success_rate_pct: 100.0,
      context_health_status: 'CRITICAL',
      root_saturation_pct: 220.0,
      peak_context_tokens: 440000,
      started_at: '2026-10-16T12:00:00Z',
      completed_at: '2026-10-16T13:00:00Z',
      lead_duration_seconds: 3600,
      lead_duration_formatted: '1h 0m',
    },
    {
      ticket_id: 'GEN-9',
      ticket_title: 'Beta regex special char [.*+?^${}()|[]\\] test',
      ticket_type: 'chore',
      updated_at: '2026-10-17T10:00:00Z',
      sdlc_iterations: 2,
      active_duration_seconds: 2400,
      active_duration_formatted: '40m 0s',
      cost_usd: 8.0,
      cache_hit_rate_pct: 50.0,
      test_to_code_ratio: '0.8:1',
      lines_added: 300,
      lines_deleted: 30,
      files_changed: 4,
      tool_success_rate_pct: 88.0,
      context_health_status: 'BLOATED',
      root_saturation_pct: 120.0,
      peak_context_tokens: 240000,
      started_at: '2026-10-17T08:00:00Z',
      completed_at: '2026-10-17T09:00:00Z',
      lead_duration_seconds: 3600,
      lead_duration_formatted: '1h 0m',
    },
  ];

  it('renders all rows and columns with correct health status classes', () => {
    render(<TicketTable rows={sampleRows} />);

    expect(screen.getByText('AL-10')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('GEN-9')).toBeInTheDocument();

    const healthyBadge = screen.getByText('HEALTHY');
    expect(healthyBadge).toHaveClass('health-badge', 'HEALTHY');

    const criticalBadge = screen.getByText('CRITICAL');
    expect(criticalBadge).toHaveClass('health-badge', 'CRITICAL');

    const bloatedBadge = screen.getByText('BLOATED');
    expect(bloatedBadge).toHaveClass('health-badge', 'BLOATED');
  });

  it('handles empty rows defensively', () => {
    render(<TicketTable rows={[]} />);
    expect(screen.getByText('No Workspaces Match Filter')).toBeInTheDocument();
    expect(
      screen.getByText('No ticket workspaces are currently registered in this session.')
    ).toBeInTheDocument();
  });

  describe('Adversarial Search Filtering', () => {
    it('filters rows with regex special characters without crashing', () => {
      render(<TicketTable rows={sampleRows} />);

      const searchInput = screen.getByLabelText(/filter tickets/i);
      fireEvent.change(searchInput, { target: { value: '.*+?' } });

      expect(screen.getByText('GEN-9')).toBeInTheDocument();
      expect(screen.queryByText('AL-10')).not.toBeInTheDocument();
      expect(screen.queryByText('AL-2')).not.toBeInTheDocument();
    });

    it('shows custom empty state when search matches no tickets', () => {
      render(<TicketTable rows={sampleRows} />);

      const searchInput = screen.getByLabelText(/filter tickets/i);
      fireEvent.change(searchInput, { target: { value: 'NONEXISTENT_XYZ' } });

      expect(screen.getByText('No Workspaces Match Filter')).toBeInTheDocument();
      expect(
        screen.getByText(/No ticket records match query "NONEXISTENT_XYZ"/i)
      ).toBeInTheDocument();
    });

    it('handles whitespace-only searches as empty queries', () => {
      render(<TicketTable rows={sampleRows} />);

      const searchInput = screen.getByLabelText(/filter tickets/i);
      fireEvent.change(searchInput, { target: { value: '   ' } });

      expect(screen.getByText('AL-10')).toBeInTheDocument();
      expect(screen.getByText('AL-2')).toBeInTheDocument();
      expect(screen.getByText('GEN-9')).toBeInTheDocument();
    });
  });

  describe('Sorting Invariants', () => {
    it('sorts by ticket ID naturally ascending and descending', () => {
      render(<TicketTable rows={sampleRows} />);

      // Default sort is ticket_id asc naturally: AL-2, AL-10, GEN-9
      let cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-2');
      expect(cells[2]).toHaveTextContent('AL-10');
      expect(cells[3]).toHaveTextContent('GEN-9');

      const ticketIdHeader = screen.getByText(/ticket id/i);
      // Click ticket ID -> toggle to desc: GEN-9, AL-10, AL-2
      fireEvent.click(ticketIdHeader);

      cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('GEN-9');
      expect(cells[2]).toHaveTextContent('AL-10');
      expect(cells[3]).toHaveTextContent('AL-2');

      // Click ticket ID again -> toggle to asc: AL-2, AL-10, GEN-9
      fireEvent.click(ticketIdHeader);

      cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-2');
      expect(cells[2]).toHaveTextContent('AL-10');
      expect(cells[3]).toHaveTextContent('GEN-9');
    });

    it('sorts by cost ascending and descending upon header clicks', () => {
      render(<TicketTable rows={sampleRows} />);

      const costHeader = screen.getByText(/cost \(usd\)/i);

      // Default sort is ticket_id asc: AL-10, AL-2, GEN-9
      // Click Cost Header -> cost asc: AL-2 ($2.20), GEN-9 ($8.00), AL-10 ($12.50)
      fireEvent.click(costHeader);

      let cells = screen.getAllByRole('row');
      // header row is index 0
      expect(cells[1]).toHaveTextContent('AL-2');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-10');

      // Click Cost Header again -> cost desc: AL-10 ($12.50), GEN-9 ($8.00), AL-2 ($2.20)
      fireEvent.click(costHeader);

      cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-10');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-2');
    });

    it('sorts by active duration ascending and descending', () => {
      render(<TicketTable rows={sampleRows} />);

      const durationHeader = screen.getByText(/active duration/i);

      // Click Duration -> asc: AL-2 (1200), GEN-9 (2400), AL-10 (3600)
      fireEvent.click(durationHeader);

      let cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-2');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-10');

      // Click Duration again -> desc: AL-10 (3600), GEN-9 (2400), AL-2 (1200)
      fireEvent.click(durationHeader);

      cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-10');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-2');
    });

    it('sorts by sdlc iterations ascending and descending', () => {
      render(<TicketTable rows={sampleRows} />);

      const iterationsHeader = screen.getByText(/sdlc iterations/i);

      // Click Iterations -> asc: AL-2 (1), GEN-9 (2), AL-10 (5)
      fireEvent.click(iterationsHeader);

      let cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-2');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-10');

      // Click Iterations again -> desc: AL-10 (5), GEN-9 (2), AL-2 (1)
      fireEvent.click(iterationsHeader);

      cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-10');
      expect(cells[2]).toHaveTextContent('GEN-9');
      expect(cells[3]).toHaveTextContent('AL-2');
    });

    it('tie-breaks equal metric values with natural ticket ID comparison', () => {
      const equalRows = [
        { ...sampleRows[0]!, ticket_id: 'AL-10', cost_usd: 5.0 },
        { ...sampleRows[0]!, ticket_id: 'AL-2', cost_usd: 5.0 },
        { ...sampleRows[0]!, ticket_id: 'AL-1', cost_usd: 5.0 },
      ];

      render(<TicketTable rows={equalRows} />);
      const costHeader = screen.getByText(/cost \(usd\)/i);
      fireEvent.click(costHeader);

      const cells = screen.getAllByRole('row');
      expect(cells[1]).toHaveTextContent('AL-1');
      expect(cells[2]).toHaveTextContent('AL-2');
      expect(cells[3]).toHaveTextContent('AL-10');
    });
  });

  describe('Row Interactivity', () => {
    it('calls onSelectTicket with ticket_id when row is clicked', () => {
      const onSelect = vi.fn();
      render(<TicketTable rows={sampleRows} onSelectTicket={onSelect} />);

      fireEvent.click(screen.getByText('AL-2'));
      expect(onSelect).toHaveBeenCalledWith('AL-2');
    });
  });

  describe('Header InfoTooltips and Event Propagation', () => {
    it('clicking InfoTooltip trigger on sortable headers stops click propagation and preserves sort order', () => {
      render(<TicketTable rows={sampleRows} />);

      const expectedTooltips = [
        'Information about Ticket ID',
        'Information about Title',
        'Information about Status',
        'Information about Active Duration',
        'Information about Cost (USD)',
        'Information about Cache Hit Rate',
        'Information about Code Delivered',
        'Information about SDLC Iterations',
      ];

      for (const label of expectedTooltips) {
        const btn = screen.getByRole('button', { name: label });
        expect(btn).toBeInTheDocument();
        const tooltipId = btn.getAttribute('aria-describedby');
        expect(tooltipId).toBeTruthy();
        const tooltip = document.getElementById(tooltipId!);
        expect(tooltip).toHaveClass('placement-bottom');
      }

      // Initial default order is ticket_id asc naturally: AL-2, AL-10, GEN-9
      let rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('AL-10');
      expect(rows[3]).toHaveTextContent('GEN-9');

      // Click InfoTooltip on Cost (USD) sortable header
      const costTooltipBtn = screen.getByRole('button', { name: 'Information about Cost (USD)' });
      fireEvent.click(costTooltipBtn);

      // Order must remain unchanged!
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('AL-10');
      expect(rows[3]).toHaveTextContent('GEN-9');

      // Click InfoTooltip on Active Duration sortable header
      const durationTooltipBtn = screen.getByRole('button', { name: 'Information about Active Duration' });
      fireEvent.click(durationTooltipBtn);

      // Order must still remain unchanged!
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('AL-10');
      expect(rows[3]).toHaveTextContent('GEN-9');

      // Click InfoTooltip on SDLC Iterations sortable header
      const iterTooltipBtn = screen.getByRole('button', { name: 'Information about SDLC Iterations' });
      fireEvent.click(iterTooltipBtn);

      // Order must still remain unchanged!
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('AL-10');
      expect(rows[3]).toHaveTextContent('GEN-9');

      // Now click the Cost header text to trigger sort asc
      const costHeader = screen.getByText(/^Cost \(USD\)/i);
      fireEvent.click(costHeader);

      // Sort by cost asc: AL-2 ($2.20), GEN-9 ($8.00), AL-10 ($12.50)
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('GEN-9');
      expect(rows[3]).toHaveTextContent('AL-10');

      // Click InfoTooltip on Cost header again
      fireEvent.click(costTooltipBtn);

      // Order must NOT toggle to descending; it must remain sorted asc!
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('GEN-9');
      expect(rows[3]).toHaveTextContent('AL-10');

      // Click InfoTooltip on Ticket ID header
      const ticketIdTooltipBtn = screen.getByRole('button', { name: 'Information about Ticket ID' });
      fireEvent.click(ticketIdTooltipBtn);

      // Order must NOT change to ticket_id sort; it must remain sorted by cost asc!
      rows = screen.getAllByRole('row');
      expect(rows[1]).toHaveTextContent('AL-2');
      expect(rows[2]).toHaveTextContent('GEN-9');
      expect(rows[3]).toHaveTextContent('AL-10');
    });
  });
});
