import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { TicketComparisons } from './TicketComparisons';
import * as apiClient from '../api/client';
import { mockAgenticMetricsLedgers } from '../api/mockData';

describe('TicketComparisons View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders loading state initially and then displays benchmarking summary KPIs and comparison matrix', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    render(<TicketComparisons />);

    expect(screen.getByText(/loading cross-ticket comparison matrix/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Benchmarked Workspaces')).toBeInTheDocument();
    });

    // Summary KPIs
    expect(screen.getByText('Average Spend / Ticket')).toBeInTheDocument();
    expect(screen.getByText('Average Active Duration')).toBeInTheDocument();
    expect(screen.getByText('Average Iterations')).toBeInTheDocument();

    // Table elements
    expect(screen.getByText('Ticket ID')).toBeInTheDocument();
    expect(screen.getByText('AL-1')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('AL-3')).toBeInTheDocument();
  });

  it('filters rows by search term input (ID, title, case-insensitive, and reset)', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText(/search by ticket id or title/i);

    // Search by Ticket ID
    fireEvent.change(searchInput, { target: { value: 'AL-2' } });
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.queryByText('AL-1')).not.toBeInTheDocument();
    expect(screen.queryByText('AL-3')).not.toBeInTheDocument();

    // Search by Title substring (case-insensitive)
    fireEvent.change(searchInput, { target: { value: 'rest api' } });
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.queryByText('AL-1')).not.toBeInTheDocument();

    // Search query with no match
    fireEvent.change(searchInput, { target: { value: 'NONEXISTENT' } });
    expect(screen.getByText(/no tickets match the query/i)).toBeInTheDocument();

    // Clear search restores all rows
    fireEvent.change(searchInput, { target: { value: '' } });
    expect(screen.getByText('AL-1')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('AL-3')).toBeInTheDocument();
  });

  it('sorts columns by completion date, cost, duration, and code lines with visual indicators', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    const { container } = render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const getFirstRowTicketId = (): string => {
      const rows = container.querySelectorAll('.clickable-comparison-row');
      return rows[0]?.querySelector('.ticket-link-btn')?.textContent?.trim() ?? '';
    };

    // 1. Sort by Cost (USD): default click sorts desc -> highest cost first (AL-1: $4.92)
    const costHeader = screen.getByRole('columnheader', { name: /cost \(usd\)/i });
    fireEvent.click(costHeader);
    expect(costHeader.querySelector('.sort-indicator.active')?.textContent).toBe('v');
    expect(getFirstRowTicketId()).toBe('AL-1');

    // Click again: sort asc -> lowest cost first (PO-1: $2.15)
    fireEvent.click(costHeader);
    expect(costHeader.querySelector('.sort-indicator.active')?.textContent).toBe('^');
    expect(getFirstRowTicketId()).toBe('PO-1');

    // 2. Sort by Active Time (duration): desc -> AL-1 (2,450s)
    const durationHeader = screen.getByRole('columnheader', { name: /active time/i });
    fireEvent.click(durationHeader);
    expect(getFirstRowTicketId()).toBe('AL-1');

    // Sort by Active Time asc -> PO-1 (1,120s)
    fireEvent.click(durationHeader);
    expect(getFirstRowTicketId()).toBe('PO-1');

    // 3. Sort by Completion Date: desc -> in-progress/null completed_at (GEN-1) floats to top
    const completedHeader = screen.getByRole('columnheader', { name: /completed/i });
    fireEvent.click(completedHeader);
    expect(getFirstRowTicketId()).toBe('GEN-1');

    // Sort by Completion Date asc -> earliest completed first (PO-1: 2026-10-14)
    fireEvent.click(completedHeader);
    expect(getFirstRowTicketId()).toBe('PO-1');

    // 4. Sort by Lines Added: desc -> AL-1 (1,250 lines added)
    const linesHeader = screen.getByRole('columnheader', { name: /code lines/i });
    fireEvent.click(linesHeader);
    expect(getFirstRowTicketId()).toBe('AL-1');
  });

  it('invokes onSelectTicket drilldown callback when user clicks a row or ticket link button', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    const onSelectTicket = vi.fn();
    const { container } = render(<TicketComparisons onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    // Test clicking button inside cell
    const ticketLink = screen.getByRole('button', { name: 'AL-1' });
    fireEvent.click(ticketLink);
    expect(onSelectTicket).toHaveBeenCalledWith('AL-1');

    // Test clicking row itself
    const rows = container.querySelectorAll('.clickable-comparison-row');
    fireEvent.click(rows[1]!); // Click second row
    expect(onSelectTicket).toHaveBeenCalledTimes(2);
  });

  it('renders zero telemetry state when no ledgers exist', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
      isMock: true,
    });

    render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText(/zero comparison workspaces available/i)).toBeInTheDocument();
    });
  });

  it('displays error banner and allows retry when API call fails', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockRejectedValueOnce(
      new Error('Failed to load comparison ledgers')
    );

    render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load comparison ledgers')).toBeInTheDocument();
    });

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValueOnce({
      data: Object.values(mockAgenticMetricsLedgers),
      isMock: true,
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('Benchmarked Workspaces')).toBeInTheDocument();
    });
  });

  it('renders all 15 expected table columns and benchmarking KPI summary values', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('Benchmarked Workspaces')).toBeInTheDocument();
    });

    // Check all KPI cards
    expect(screen.getByText('Benchmarked Workspaces')).toBeInTheDocument();
    expect(screen.getByText('Average Active Duration')).toBeInTheDocument();
    expect(screen.getByText('Average Iterations')).toBeInTheDocument();
    expect(screen.getByText('Fleet Tool Success Rate')).toBeInTheDocument();
    expect(screen.getByText('Average Spend / Ticket')).toBeInTheDocument();

    // Check all 15 column headers
    const expectedHeaders = [
      /ticket id/i,
      /title/i,
      /status/i,
      /completed/i,
      /started/i,
      /lead time/i,
      /active time/i,
      /iterations/i,
      /cost \(usd\)/i,
      /cache hit/i,
      /test:prod/i,
      /code lines/i,
      /tool success/i,
      /context health/i,
      /saturation/i,
    ];

    for (const h of expectedHeaders) {
      expect(screen.getByRole('columnheader', { name: h })).toBeInTheDocument();
    }
  });

  it('contains zero emojis in the rendered comparisons document body', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });

    const { container } = render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('Benchmarked Workspaces')).toBeInTheDocument();
    });

    const renderedText = container.textContent ?? '';
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(renderedText)).toBe(false);
  });
});
