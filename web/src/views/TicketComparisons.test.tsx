import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { TicketComparisons } from './TicketComparisons';
import * as apiClient from '../api/client';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';

describe('TicketComparisons View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('shows ticket-specific skipped-workspace diagnostics alongside valid rows', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: Object.values(mockAgenticMetricsLedgers),
      diagnostics: [{ ticket_id: 'AL-OLD', message: 'Schema version 2 needs migration and source-backed backfill.' }],
    });

    render(<TicketComparisons />);

    expect(await screen.findByText('Some ticket workspaces were skipped.')).toBeInTheDocument();
    expect(screen.getByRole('status').textContent).toContain('AL-OLD: Schema version 2 needs migration and source-backed backfill.');
    expect(screen.getByText('AL-1')).toBeInTheDocument();
  });

  it('renders loading state initially and then displays benchmarking summary KPIs and comparison matrix', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
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

  it('filters comparison rows by ticket type and workspace prefix dropdowns', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    // Verify all rows initially visible
    expect(screen.getByText('AL-1')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('AL-3')).toBeInTheDocument();
    expect(screen.getByText('PO-1')).toBeInTheDocument();
    expect(screen.getByText('GEN-1')).toBeInTheDocument();

    // Filter by Ticket Type: Feature
    const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(typeTrigger);
    const featureCheckbox = screen.getByRole('checkbox', { name: 'Feature' });
    fireEvent.click(featureCheckbox);

    // Only AL-1, AL-2, AL-3 should be visible
    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('AL-3')).toBeInTheDocument();
    expect(screen.queryByText('PO-1')).not.toBeInTheDocument();
    expect(screen.queryByText('GEN-1')).not.toBeInTheDocument();

    // Reset ticket type filter to All
    const allTypeCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allTypeCheckbox);

    // All should be back
    await waitFor(() => {
      expect(screen.getByText('PO-1')).toBeInTheDocument();
    });
    expect(screen.getByText('AL-1')).toBeInTheDocument();

    // Close ticket type dropdown
    fireEvent.click(typeTrigger);

    // Filter by Prefix: PO
    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);
    const poCheckbox = screen.getByRole('checkbox', { name: 'PO' });
    fireEvent.click(poCheckbox);

    // Only PO-1 should be visible
    await waitFor(() => {
      expect(screen.getByText('PO-1')).toBeInTheDocument();
    });
    expect(screen.queryByText('AL-1')).not.toBeInTheDocument();
    expect(screen.queryByText('AL-2')).not.toBeInTheDocument();
    expect(screen.queryByText('AL-3')).not.toBeInTheDocument();
    expect(screen.queryByText('GEN-1')).not.toBeInTheDocument();

    // Reset Prefix filter to All
    const allPrefixCheckbox = screen.getByRole('checkbox', { name: 'Select all Prefix' });
    fireEvent.click(allPrefixCheckbox);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    // Close prefix dropdown
    fireEvent.click(prefixTrigger);

    // Unmatched filter displays fallback table message
    fireEvent.click(typeTrigger);
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    await waitFor(() => {
      expect(screen.getByText('No tickets match the selected filters.')).toBeInTheDocument();
    });
  });

  it('keeps table controls bar visible and interactive when zero ledgers match filters', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    const { container } = render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('Zero Comparison Workspaces Available')).toBeInTheDocument();
    });

    expect(container.querySelector('.kpi-grid')).toBeInTheDocument();
    expect(container.querySelector('.table-controls-bar')).toBeInTheDocument();

    expect(screen.getByRole('combobox', { name: 'Ticket Type selector' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Prefix selector' })).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/search by ticket id or title/i)).toBeInTheDocument();
  });

  it('allows interactive recovery from empty filter selection in TicketComparisons', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    const fetchSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
      .mockResolvedValueOnce({
        data: ledgers,
        })
      .mockResolvedValueOnce({
        data: [],
        })
      .mockResolvedValueOnce({
        data: ledgers,
        });

    const { container } = render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(typeTrigger);
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    await waitFor(() => {
      expect(screen.getByText('Zero Comparison Workspaces Available')).toBeInTheDocument();
    });
    expect(container.querySelector('.table-controls-bar')).toBeInTheDocument();

    const allCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allCheckbox);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });
    expect(screen.queryByText('Zero Comparison Workspaces Available')).not.toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(3);
  });

  it('re-triggers data fetch when filters change from initial zero ledgers empty state', async () => {
    const fetchSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
      .mockResolvedValueOnce({
        data: [],
        })
      .mockResolvedValueOnce({
        data: Object.values(mockAgenticMetricsLedgers),
        });

    const { container } = render(<TicketComparisons />);

    await waitFor(() => {
      expect(screen.getByText('Zero Comparison Workspaces Available')).toBeInTheDocument();
    });

    expect(container.querySelector('.table-controls-bar')).toBeInTheDocument();

    // Change prefix filter from empty state
    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);
    const alCheckbox = screen.getByRole('checkbox', { name: 'AL' });
    fireEvent.click(alCheckbox);

    // Should re-fetch and render the returned rows
    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });
    expect(screen.queryByText('Zero Comparison Workspaces Available')).not.toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(2);
    expect(fetchSpy).toHaveBeenLastCalledWith('all', undefined, undefined, undefined, ['AL']);
  });

  it('renders configured prefixes in prefix dropdown even when no ledgers for those prefixes exist', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    render(<TicketComparisons configuredPrefixes={['NEWFEATURE', 'PLATFORM']} />);

    await waitFor(() => {
      expect(screen.getByText('Zero Comparison Workspaces Available')).toBeInTheDocument();
    });

    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);

    expect(screen.getByRole('checkbox', { name: 'NEWFEATURE' })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: 'PLATFORM' })).toBeInTheDocument();
  });
});
