import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react';
import { TicketComparisons } from './TicketComparisons';
import type { SkillFact } from '../domain/types';
import * as apiClient from '../api/client';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';

describe('TicketComparisons View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('matches literal skill names with OR semantics, composes filters, and recovers from empty results', async () => {
    const fact = (name: string): SkillFact => ({ name, success: true, caller_role: 'coder', turn_index: 0, message_index: 0, timestamp: '2026-01-01T00:00:00Z' });
    const ids = ['AL-1', 'AL-2', 'AL-3', 'PO-1'];
    const ledgers = ids.map((id, index) => {
      const ledger = mockAgenticMetricsLedgers[id]!;
      const amount = index === 0 ? 6 : index === 1 ? 0 : null;
      const cost_summary = index === 2
        ? { state: 'partial' as const, known_subtotal_usd: 4, fully_observed: 0, partial: 1, unavailable: 0, total: 1 }
        : index === 3
          ? { state: 'unavailable' as const, known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 }
          : { state: 'complete' as const, known_subtotal_usd: amount, fully_observed: 1, partial: 0, unavailable: 0, total: 1 };
      return { ...ledger, totals: { ...ledger.totals, cost_usd: amount, cost_summary },
        workflow_efficiency: { ...ledger.workflow_efficiency, active_duration_seconds: [60, 120, 900, 300][index]!, sdlc_iterations: [2, 4, 12, 20][index]! },
        facts: { ...ledger.facts, skills: index === 0 ? [fact('all'), fact('all')] : index === 1 ? [fact('ALL')] : index === 3 ? [fact('all')] : [] } };
    });
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: ledgers });
    const { container } = render(<TicketComparisons />);
    await screen.findByText('AL-1');
    const table = screen.getByRole('table');
    const visibleIDs = () => [...container.querySelectorAll('.clickable-comparison-row .ticket-link-btn')].map((node) => node.textContent?.trim()).sort();
    const points = () => container.querySelectorAll('circle.ticket-scatter-point');
    const expectSpend = (value: string, population: string) => {
      const card = screen.getByText('Average Spend / Ticket').closest<HTMLElement>('.kpi-card')!;
      expect(card.querySelector('.kpi-value')?.textContent).toBe(value);
      expect(card.querySelector('.kpi-secondary')?.textContent).toBe(population);
    };
    expectSpend('$3.00', '2 of 4 tickets with cost');
    expect(within(table).getByText('No recorded skills')).toBeInTheDocument();
    expect(within(table).getAllByText('all')).toHaveLength(2);
    fireEvent.click(screen.getByRole('combobox', { name: 'Filter by Skill selector' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'all' }));
    expect(visibleIDs()).toEqual(['AL-1', 'PO-1']);
    expectSpend('$6.00', '1 of 2 tickets with cost');
    fireEvent.click(screen.getByRole('checkbox', { name: 'ALL' }));
    expect(visibleIDs()).toEqual(['AL-1', 'AL-2', 'PO-1']);
    expectSpend('$3.00', '2 of 3 tickets with cost');
    expect(screen.queryByText('AL-3')).not.toBeInTheDocument();
    expect(points()).toHaveLength(2); // Missing cost omits its cost-axis point, not its row.
    expect(screen.getByText('Benchmarked Workspaces').closest('.kpi-card')).toHaveTextContent('3');
    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'Feature' }));
    await waitFor(() => expect(visibleIDs()).toEqual(['AL-1', 'AL-2']));
    expectSpend('$3.00', '2 of 2 tickets with cost');
    fireEvent.click(screen.getByRole('combobox', { name: 'Prefix selector' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'AL' }));
    await waitFor(() => expect(visibleIDs()).toEqual(['AL-1', 'AL-2']));
    expectSpend('$3.00', '2 of 2 tickets with cost');
    const search = screen.getByPlaceholderText(/search by ticket id or title/i);
    fireEvent.change(search, { target: { value: 'AL-3' } });
    expect(screen.getByText(/No tickets match the query/)).toBeInTheDocument();
    expectSpend('Unavailable', '0 of 0 tickets with cost');
    // Options are still derived from source ledgers while no displayed row matches.
    expect(screen.getByRole('checkbox', { name: 'all' })).toBeChecked();
    fireEvent.click(screen.getByRole('checkbox', { name: 'Select all Filter by Skill' }));
    expect(visibleIDs()).toEqual(['AL-3']);
    expectSpend('Unavailable', '0 of 1 tickets with cost');
    expect(points()).toHaveLength(0); // The partial row has no complete cost-axis value.
    fireEvent.change(search, { target: { value: '' } });
    expect(visibleIDs()).toEqual(['AL-1', 'AL-2', 'AL-3']);
    expectSpend('$3.00', '2 of 3 tickets with cost');
    expect(screen.getByText('Average Active Duration').closest('.kpi-card')?.querySelector('.kpi-value')?.textContent).toBe('6m 0s');
    expect(screen.getByText('Average Iterations').closest('.kpi-card')?.querySelector('.kpi-value')?.textContent).toBe('6.0');
    expect(screen.getByText('Benchmarked Workspaces').closest('.kpi-card')?.querySelector('.kpi-value')?.textContent).toBe('3');
    expect(screen.queryByText('PO-1')).not.toBeInTheDocument();
  });

  it('averages complete filtered ticket costs including zero and recovers after unknown and empty searches', async () => {
    const ledgers = ['AL-1', 'AL-2', 'AL-3'].map((id, index) => {
      const base = mockAgenticMetricsLedgers[id]!;
      const amount = index === 0 ? 6 : index === 1 ? 0 : null;
      return { ...base, totals: { ...base.totals, cost_usd: amount,
        cost_summary: index === 2
          ? { state: 'partial' as const, known_subtotal_usd: 4, fully_observed: 0, partial: 1, unavailable: 0, total: 1 }
          : { state: 'complete' as const, known_subtotal_usd: amount, fully_observed: 1, partial: 0, unavailable: 0, total: 1 },
      } };
    });
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: ledgers });
    render(<TicketComparisons />);
    const card = (await screen.findByText('Average Spend / Ticket')).closest<HTMLElement>('.kpi-card')!;
    const value = () => card.querySelector('.kpi-value')?.textContent;
    const population = () => card.querySelector('.kpi-secondary')?.textContent;
    expect(value()).toBe('$3.00');
    expect(population()).toBe('2 of 3 tickets with cost');
    expect(screen.getByText('Benchmarked Workspaces').closest('.kpi-card')?.querySelector('.kpi-value')).toHaveTextContent('3');
    expect(within(card).getByRole('tooltip')).toHaveTextContent('excludes unavailable and partial costs');
    const search = screen.getByPlaceholderText(/search by ticket id or title/i);
    fireEvent.change(search, { target: { value: 'AL-1' } });
    expect(value()).toBe('$6.00');
    expect(population()).toBe('1 of 1 tickets with cost');
    fireEvent.change(search, { target: { value: 'AL-2' } });
    expect(value()).toBe('$0.00');
    expect(population()).toBe('1 of 1 tickets with cost');
    fireEvent.change(search, { target: { value: 'AL-3' } });
    expect(value()).toBe('Unavailable');
    expect(population()).toBe('0 of 1 tickets with cost');
    fireEvent.change(search, { target: { value: 'missing-ticket' } });
    expect(value()).toBe('Unavailable');
    expect(population()).toBe('0 of 0 tickets with cost');
    fireEvent.change(search, { target: { value: '' } });
    expect(value()).toBe('$3.00');
    expect(population()).toBe('2 of 3 tickets with cost');
  });

  it('replaces eligible cost populations when filter refetches return changed ledgers', async () => {
    const row = (id: string, amount: number | null) => {
      const base = mockAgenticMetricsLedgers[id]!;
      return { ...base, totals: { ...base.totals, cost_usd: amount } };
    };
    const fetchLedgers = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
      .mockResolvedValueOnce({ data: [row('AL-1', 6), row('PO-1', null)] })
      .mockResolvedValueOnce({ data: [row('AL-1', 0), row('AL-2', null), row('AL-3', null)] })
      .mockResolvedValueOnce({ data: [row('AL-1', null), row('AL-2', null)] })
      .mockResolvedValue({ data: [row('AL-1', 8), row('AL-2', 0)] });
    const { container } = render(<TicketComparisons />);
    const card = (await screen.findByText('Average Spend / Ticket')).closest<HTMLElement>('.kpi-card')!;
    const value = () => card.querySelector('.kpi-value')?.textContent;
    const population = () => card.querySelector('.kpi-secondary')?.textContent;
    expect(value()).toBe('$6.00');
    expect(population()).toBe('1 of 2 tickets with cost');
    fireEvent.click(screen.getByRole('combobox', { name: 'Prefix selector' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'AL' }));
    await waitFor(() => expect(population()).toBe('1 of 3 tickets with cost'));
    expect(fetchLedgers).toHaveBeenLastCalledWith('all', undefined, undefined, undefined, ['AL']);
    expect(value()).toBe('$0.00');
    expect(container.querySelectorAll('.clickable-comparison-row')).toHaveLength(3);
    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'Feature' }));
    await waitFor(() => expect(population()).toBe('0 of 2 tickets with cost'));
    expect(value()).toBe('Unavailable');
    expect(container.querySelectorAll('.clickable-comparison-row')).toHaveLength(2);
    fireEvent.click(screen.getByRole('checkbox', { name: 'Select all Prefix' }));
    await waitFor(() => expect(value()).toBe('$4.00'));
    expect(population()).toBe('2 of 2 tickets with cost');
    expect(fetchLedgers).toHaveBeenCalledTimes(4);
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

  it('integrates the scatter plot with default metrics and ticket deep dive navigation', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: Object.values(mockAgenticMetricsLedgers) });
    const onSelectTicket = vi.fn();
    render(<TicketComparisons onSelectTicket={onSelectTicket} />);
    expect(await screen.findByRole('heading', { name: 'Cross-Ticket Metric Correlation' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'X-Axis' })).toHaveValue('cost_usd');
    expect(screen.getByRole('combobox', { name: 'Y-Axis' })).toHaveValue('net_lines_added');
    fireEvent.click(screen.getByRole('button', { name: /AL-1:/ }));
    expect(onSelectTicket).toHaveBeenCalledWith('AL-1');
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

  it('renders all 16 expected table columns and benchmarking KPI summary values', async () => {
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

    // Check all 16 column headers
    const expectedHeaders = [
      /ticket id/i,
      /title/i,
      /skills used/i,
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
