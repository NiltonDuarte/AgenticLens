import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
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
});
