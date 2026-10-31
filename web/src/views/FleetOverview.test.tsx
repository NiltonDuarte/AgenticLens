import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { FleetOverview } from './FleetOverview';
import * as apiClient from '../api/client';
import { mockFleetSummary, mockTicketComparisonRows } from '../test/fixtures/mockData';

describe('FleetOverview View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('shows ticket-specific skipped-workspace diagnostics alongside valid workspace data', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      diagnostics: [{ ticket_id: 'AL-OLD', message: 'Incomplete migration must be resumed.' }],
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });

    render(<FleetOverview />);

    expect(await screen.findByText('Some ticket workspaces were skipped.')).toBeInTheDocument();
    expect(screen.getByRole('status').textContent).toContain('AL-OLD: Incomplete migration must be resumed.');
    expect(screen.getByText('Total Cost')).toBeInTheDocument();
  });

  it('renders loading indicator initially then displays KPI values and tables', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
    });

    render(<FleetOverview />);

    expect(screen.getByText(/loading telemetry metrics/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Total Cost')).toBeInTheDocument();
    });

    expect(screen.getByText('$14.82')).toBeInTheDocument();
    expect(screen.getByText('128')).toBeInTheDocument();
    expect(screen.getByText('78.8%')).toBeInTheDocument();
    expect(screen.getByText('2h 3m')).toBeInTheDocument();

    expect(screen.getByText('Role Cost Distribution')).toBeInTheDocument();
    expect(screen.getByText('Model Cost Distribution')).toBeInTheDocument();

    expect(screen.getByText('AL-1')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('PO-1')).toBeInTheDocument();
    expect(screen.getByText('GEN-1')).toBeInTheDocument();
  });

  it('filters rows when user enters query in ticket table search', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
    });

    render(<FleetOverview />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByLabelText(/filter tickets/i);
    fireEvent.change(searchInput, { target: { value: 'AL-2' } });

    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.queryByText('AL-1')).not.toBeInTheDocument();
    expect(screen.queryByText('PO-1')).not.toBeInTheDocument();
  });

  describe('Adversarial & Boundary Scenarios', () => {
    it('displays error banner and allows retry when fetch throws unexpected error', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary')
        .mockRejectedValueOnce(new Error('Fatal API crash'))
        .mockResolvedValueOnce({
          data: mockFleetSummary,
            });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Fatal API crash')).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /retry connection/i });
      fireEvent.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });
      expect(fetchFleetSpy).toHaveBeenCalledTimes(2);
    });

    it('renders empty state when workspace list is empty but total_tickets is 0', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: {
          ...mockFleetSummary,
          total_tickets: 0,
          total_cost_usd: 0,
        },
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Zero Telemetry Workspaces Available')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: /refresh data/i })).toBeInTheDocument();
    });

    it('invokes onSelectTicket when a ticket row is clicked', async () => {
      const onSelectTicket = vi.fn();
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview onSelectTicket={onSelectTicket} />);

      await waitFor(() => {
        expect(screen.getByText('AL-1')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('AL-1'));
      expect(onSelectTicket).toHaveBeenCalledWith('AL-1');
    });

    it('renders FleetAnomalyCard in the fleet overview layout with anomaly metrics', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Fleet Anomaly and Reliability Profile')).toBeInTheDocument();
      });

      // Fleet anomaly metrics
      expect(screen.getByText('1 warning, 1 critical')).toBeInTheDocument();
      expect(screen.getByText('Tool Retry Streak')).toBeInTheDocument();
      expect(screen.getByText('Brittle Tools')).toBeInTheDocument();
    });

    it('triggers refetch with preset window when date range preset changes', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const presetSelect = screen.getByLabelText(/date range preset/i);
      fireEvent.change(presetSelect, { target: { value: '7d' } });

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('7d', undefined, undefined);
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('7d', '', '', undefined, undefined);
      });
    });

    it('displays custom date pickers and passes since/until when custom preset is selected', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const presetSelect = screen.getByLabelText(/date range preset/i);
      fireEvent.change(presetSelect, { target: { value: 'custom' } });

      const sinceInput = await screen.findByLabelText(/custom start date/i);
      const untilInput = await screen.findByLabelText(/custom end date/i);

      fireEvent.change(sinceInput, { target: { value: '2026-03-01' } });
      fireEvent.change(untilInput, { target: { value: '2026-03-15' } });

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('all', '2026-03-01', '2026-03-15');
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('all', '', '', '2026-03-01', '2026-03-15');
      });
    });

    it('triggers refetch when ticket type and prefix filters change', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      // Open ticket type dropdown and select Bug
      const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
      fireEvent.click(typeTrigger);

      const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
      fireEvent.click(bugCheckbox);

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('all', undefined, undefined, ['bug'], undefined);
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('all', '', '', undefined, undefined, ['bug'], undefined);
      });
    });

    it('renders configured prefixes in prefix dropdown even when no matching workspaces exist', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [],
        });

      render(<FleetOverview configuredPrefixes={['NEWINITIATIVE', 'INFRA']} />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
      fireEvent.click(prefixTrigger);

      expect(screen.getByRole('checkbox', { name: 'NEWINITIATIVE' })).toBeInTheDocument();
      expect(screen.getByRole('checkbox', { name: 'INFRA' })).toBeInTheDocument();
    });
  });
});
