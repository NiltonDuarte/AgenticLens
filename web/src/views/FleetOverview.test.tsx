import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { FleetOverview } from './FleetOverview';
import * as apiClient from '../api/client';
import { mockFleetSummary, mockTicketComparisonRows } from '../api/mockData';

describe('FleetOverview View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders loading indicator initially then displays KPI values and tables', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
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
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
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

  it('displays empty state when total_tickets is zero', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: {
        ...mockFleetSummary,
        total_tickets: 0,
        total_cost_usd: 0,
      },
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [],
      isMock: true,
    });

    render(<FleetOverview />);

    await waitFor(() => {
      expect(screen.getByText(/Zero Telemetry Workspaces Available/i)).toBeInTheDocument();
    });
  });

  it('contains zero emojis in the rendered document body', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const { container } = render(<FleetOverview />);

    await waitFor(() => {
      expect(screen.getByText('Total Cost')).toBeInTheDocument();
    });

    const renderedText = container.textContent ?? '';
    // Regex matching emoji characters
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(renderedText)).toBe(false);
  });

  describe('Adversarial & Boundary Scenarios', () => {
    it('displays error banner and allows retry when fetch throws unexpected error', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary')
        .mockRejectedValueOnce(new Error('Fatal API crash'))
        .mockResolvedValueOnce({
          data: mockFleetSummary,
          isMock: false,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        isMock: false,
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
        },
        isMock: false,
      });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [],
        isMock: false,
      });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Zero Telemetry Workspaces Available')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: /refresh data/i })).toBeInTheDocument();
    });

    it('notifies onMockChange callback with true if either summary or workspaces is mock', async () => {
      const onMockChange = vi.fn();
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        isMock: false,
      });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        isMock: true,
      });

      render(<FleetOverview onMockChange={onMockChange} />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      expect(onMockChange).toHaveBeenCalledWith(true);
    });

    it('invokes onSelectTicket when a ticket row is clicked', async () => {
      const onSelectTicket = vi.fn();
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        isMock: false,
      });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        isMock: false,
      });

      render(<FleetOverview onSelectTicket={onSelectTicket} />);

      await waitFor(() => {
        expect(screen.getByText('AL-1')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('AL-1'));
      expect(onSelectTicket).toHaveBeenCalledWith('AL-1');
    });
  });
});
