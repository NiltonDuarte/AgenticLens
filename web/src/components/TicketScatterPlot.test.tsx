import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockTicketComparisonRows } from '../test/fixtures/mockData';
import { TicketScatterPlot } from './TicketScatterPlot';

describe('TicketScatterPlot', () => {
  it('defaults to cost and net lines and renders one SVG point per ticket', () => {
    const { container } = render(<TicketScatterPlot rows={mockTicketComparisonRows} />);
    expect(screen.getByRole('combobox', { name: 'X-Axis' })).toHaveValue('cost_usd');
    expect(screen.getByRole('combobox', { name: 'Y-Axis' })).toHaveValue('net_lines_added');
    expect(screen.getByRole('heading', { name: 'Cross-Ticket Metric Correlation' })).toBeInTheDocument();
    expect(container.querySelectorAll('circle.ticket-scatter-point')).toHaveLength(mockTicketComparisonRows.length);
    const points = container.querySelectorAll('circle.ticket-scatter-point');
    expect(points[0]).toHaveAttribute('fill', '#4ea1ff');
    expect(points[4]).toHaveAttribute('fill', '#f2a93b');
    expect(points[0]?.querySelector('title')?.textContent).toContain(mockTicketComparisonRows[0]!.ticket_title);
  });

  it('filters excluded tickets and presents accessible details and routes clicks', () => {
    const onSelectTicket = vi.fn();
    render(<TicketScatterPlot rows={mockTicketComparisonRows} onSelectTicket={onSelectTicket} />);
    fireEvent.click(screen.getByText(/Exclude tickets/));
    const checkbox = screen.getByRole('checkbox', { name: /AL-1/ });
    fireEvent.click(checkbox);
    expect(screen.getByText('Exclude tickets (1 excluded)')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /AL-1:/ })).not.toBeInTheDocument();
    const point = screen.getByRole('button', { name: /AL-2:/ });
    expect(point).toHaveAttribute('aria-label', expect.stringContaining('Total Cost (USD)'));
    fireEvent.click(point);
    expect(onSelectTicket).toHaveBeenCalledWith('AL-2');
  });

  it('searches IDs and titles case-insensitively and recovers from no matches without changing the chart', async () => {
    const user = userEvent.setup();
    render(<TicketScatterPlot rows={mockTicketComparisonRows} />);
    await user.click(screen.getByText(/Exclude tickets/));
    const search = screen.getByRole('searchbox', { name: 'Search tickets' });
    const chartBeforeSearch = screen.getByRole('img').innerHTML;
    const clear = screen.getByRole('button', { name: 'Clear search' });
    expect(clear).toBeDisabled();

    for (const [query, expectedId] of [['aL-2', 'AL-2'], ['dAsHbOaRd', 'AL-3']] as const) {
      await user.clear(search);
      await user.type(search, query);
      expect(screen.getAllByRole('checkbox')).toHaveLength(1);
      expect(screen.getByRole('checkbox', { name: new RegExp(expectedId) })).toBeChecked();
      expect(screen.getByText('Exclude tickets (0 excluded)')).toBeInTheDocument();
      expect(screen.getByRole('img').innerHTML).toBe(chartBeforeSearch);
    }

    await user.clear(search);
    await user.type(search, 'no-such-ticket');
    expect(screen.queryAllByRole('checkbox')).toHaveLength(0);
    expect(screen.getByRole('status')).toHaveTextContent('No tickets match');
    expect(search).toHaveValue('no-such-ticket');
    expect(screen.getByRole('img').innerHTML).toBe(chartBeforeSearch);
    await user.tab();
    expect(clear).toHaveFocus();
    await user.keyboard('{Enter}');
    expect(search).toHaveValue('');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
    expect(screen.getAllByRole('checkbox')).toHaveLength(mockTicketComparisonRows.length);
    expect(screen.getByRole('img').innerHTML).toBe(chartBeforeSearch);
    expect(screen.getByText('Exclude tickets (0 excluded)')).toBeInTheDocument();
  });

  it('preserves hidden exclusions and supports keyboard toggling, restoration and point navigation while searching', async () => {
    const user = userEvent.setup();
    const onSelectTicket = vi.fn();
    render(<TicketScatterPlot rows={mockTicketComparisonRows} onSelectTicket={onSelectTicket} />);
    await user.click(screen.getByText(/Exclude tickets/));
    await user.click(screen.getByRole('checkbox', { name: /AL-1/ }));
    const search = screen.getByRole('searchbox', { name: 'Search tickets' });
    const chartAfterExclusion = screen.getByRole('img').innerHTML;
    await user.type(search, 'al-2');
    expect(screen.queryByRole('checkbox', { name: /AL-1/ })).not.toBeInTheDocument();
    expect(screen.getByText('Exclude tickets (1 excluded)')).toBeInTheDocument();
    expect(screen.getByRole('img').innerHTML).toBe(chartAfterExclusion);
    await user.tab();
    await user.tab();
    expect(screen.getByRole('checkbox', { name: /AL-2/ })).toHaveFocus();
    await user.keyboard(' ');
    expect(screen.getByRole('checkbox', { name: /AL-2/ })).not.toBeChecked();
    expect(screen.getByText('Exclude tickets (2 excluded)')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /AL-[12]:/ })).not.toBeInTheDocument();

    const point = screen.getByRole('button', { name: /AL-3:/ });
    point.focus();
    await user.keyboard('{Enter} ');
    expect(onSelectTicket.mock.calls).toEqual([['AL-3'], ['AL-3']]);

    await user.clear(search);
    await user.type(search, 'al-1');
    expect(screen.getByRole('checkbox', { name: /AL-1/ })).not.toBeChecked();
    await user.tab();
    await user.tab();
    await user.keyboard(' ');
    expect(screen.getByRole('checkbox', { name: /AL-1/ })).toBeChecked();
    expect(screen.getByRole('button', { name: /AL-1:/ })).toBeInTheDocument();
    expect(screen.getByText('Exclude tickets (1 excluded)')).toBeInTheDocument();
    const chartAfterRestoration = screen.getByRole('img').innerHTML;
    await user.click(screen.getByRole('button', { name: 'Clear search' }));
    expect(screen.getByRole('checkbox', { name: /AL-1/ })).toBeChecked();
    expect(screen.getByRole('checkbox', { name: /AL-2/ })).not.toBeChecked();
    expect(screen.getByRole('img').innerHTML).toBe(chartAfterRestoration);
    expect(screen.queryByRole('button', { name: /AL-2:/ })).not.toBeInTheDocument();
  });

  it('includes recorded skills and explicit missing-data copy in accessible point tooltips', () => {
    const rows = mockTicketComparisonRows.slice(0, 2).map((row, index) => ({ ...row, skills_used: index === 0 ? ['alpha', 'beta'] : [] }));
    render(<TicketScatterPlot rows={rows} />);
    const recorded = screen.getByRole('button', { name: /AL-1:/ });
    expect(recorded).toHaveAttribute('aria-label', expect.stringContaining('Skills Used: alpha, beta'));
    expect(recorded.querySelector('title')).toHaveTextContent('Skills Used: alpha, beta');
    expect(screen.getByRole('button', { name: /AL-2:/ })).toHaveAttribute('aria-label', expect.stringContaining('No recorded skills'));
  });

  it('allows independent metric selection and switching to log scale', () => {
    render(<TicketScatterPlot rows={mockTicketComparisonRows} />);
    fireEvent.change(screen.getByRole('combobox', { name: 'X-Axis' }), { target: { value: 'active_duration_seconds' } });
    fireEvent.change(screen.getByRole('combobox', { name: 'Y-Axis' }), { target: { value: 'cache_hit_rate_pct' } });
    fireEvent.change(screen.getByRole('combobox', { name: 'Scale' }), { target: { value: 'logarithmic' } });
    expect(screen.getByRole('img')).toHaveAttribute('aria-label', 'Cache Hit Rate by Active Duration');
    expect(screen.getByRole('img')).toHaveTextContent('(log10)');
  });
  it('colors ticket tags with stable accessible identities without moving or duplicating points', () => {
    const rows = mockTicketComparisonRows.slice(0, 3).map((row, index) => ({ ...row, ...(index < 2 ? { tags: { team: index === 0 ? 'Missing' : 'Mixed' } } : {}) }));
    const { container, rerender } = render(<TicketScatterPlot rows={rows} categoryRows={rows} />);
    const before = [...container.querySelectorAll('circle')].map((point) => [point.getAttribute('cx'), point.getAttribute('cy')]);
    fireEvent.change(screen.getByRole('combobox', { name: 'Color by tag' }), { target: { value: 'team' } });
    const points = [...container.querySelectorAll('circle')];
    expect(points.map((point) => [point.getAttribute('cx'), point.getAttribute('cy')])).toEqual(before);
    expect(points).toHaveLength(3);
    const first = screen.getByRole('button', { name: /AL-1:/ });
    expect(first).toHaveAccessibleName(expect.stringContaining('Tag team: Value: Missing'));
    expect(first).toHaveAccessibleName(expect.stringContaining('Status: Completed'));
    expect(screen.getByRole('button', { name: /AL-3:/ })).toHaveAccessibleName(expect.stringContaining('Tag team: Missing'));
    expect(screen.getByText('Value: Mixed')).toBeInTheDocument();
    const fill = first.getAttribute('fill');
    rerender(<TicketScatterPlot rows={[rows[0]!]} categoryRows={rows} />);
    expect(screen.getByRole('button', { name: /AL-1:/ })).toHaveAttribute('fill', fill);
    expect(screen.getByText('Value: Mixed')).toBeInTheDocument();
    fireEvent.change(screen.getByRole('combobox', { name: 'Color by tag' }), { target: { value: '' } });
    expect(screen.getByRole('button', { name: /AL-1:/ })).toHaveAttribute('fill', '#4ea1ff');
  });

  it('assigns current values colors and legend entries when a supplied catalog is stale', () => {
    const base = mockTicketComparisonRows[0]!;
    const previous = { ...base, tags: { team: 'old' } };
    const current = { ...base, tags: { team: 'new' } };
    const outside = { ...base, ticket_id: 'outside', tags: { team: 'outside' } };
    render(<TicketScatterPlot rows={[current]} categoryRows={[previous, outside]} />);
    fireEvent.change(screen.getByRole('combobox', { name: 'Color by tag' }), { target: { value: 'team' } });
    expect(screen.getByRole('button', { name: /AL-1:.*Tag team: Value: new/ }).getAttribute('fill')).toBeTruthy();
    expect(screen.getByText('Value: new')).toBeInTheDocument();
    expect(screen.getByText('Value: outside')).toBeInTheDocument();
    expect(screen.queryByText('Value: old')).not.toBeInTheDocument();
  });

});
