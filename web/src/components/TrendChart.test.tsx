import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { TrendChart } from './TrendChart';
import { formatLinesCompact } from '../domain/formatters';

describe('TrendChart Component', () => {
  it('omits unknown cost points and starts a new segment after a gap', () => {
    const { container } = render(<TrendChart
      title="Reported cost"
      series={[{ key: 'cost', label: 'Cost', color: '#123456' }]}
      data={[{ bucket_date: 'first', cost: 0 }, { bucket_date: 'missing', cost: null }, { bucket_date: 'last', cost: 2 }]}
    />);
    expect(container.querySelectorAll('circle')).toHaveLength(2);
    const path = container.querySelector('path[stroke-width="2.5"]')?.getAttribute('d');
    expect(path?.match(/M /g)).toHaveLength(2);
    expect(path).not.toContain('L ');
    expect(screen.getByText(/1 unavailable values omitted/)).toBeInTheDocument();
    const columns = container.querySelectorAll('rect[fill="transparent"]');
    fireEvent.mouseEnter(columns[1]!);
    expect(screen.getByText('Unavailable')).toBeInTheDocument();
  });

  it('renders empty message when data is empty', () => {
    render(
      <TrendChart
        title="Empty Test Chart"
        series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
        data={[]}
        emptyMessage="Custom empty progression message"
      />
    );

    expect(screen.getByText('Custom empty progression message')).toBeInTheDocument();
    expect(screen.getByText('Empty Test Chart')).toBeInTheDocument();
  });

  it('renders standard positive-only chart with grid lines, bars, and line series', () => {
    const data = [
      { bucket_date: '2026-03-01', period_spend: 10, cum_spend: 10 },
      { bucket_date: '2026-03-02', period_spend: 25, cum_spend: 35 },
    ];

    const { container } = render(
      <TrendChart
        title="Standard Chart"
        data={data}
        series={[
          { key: 'period_spend', label: 'Spend', color: '#10b981', type: 'bar' },
          { key: 'cum_spend', label: 'Cumulative', color: '#3b82f6', type: 'line' },
        ]}
        valueFormatter={(v) => `$${v.toFixed(2)}`}
      />
    );

    expect(screen.getByText('Standard Chart')).toBeInTheDocument();
    expect(screen.getByText('Spend')).toBeInTheDocument();
    expect(screen.getByText('Cumulative')).toBeInTheDocument();

    const svg = container.querySelector('svg.trend-svg');
    expect(svg).toBeInTheDocument();

    const bars = container.querySelectorAll('rect[fill-opacity="0.8"]');
    expect(bars.length).toBe(2);

    const lines = container.querySelectorAll('path[stroke-width="2.5"]');
    expect(lines.length).toBe(1);

    // No zero baseline needed for non-negative chart
    const zeroBaseline = container.querySelector('line.chart-zero-baseline');
    expect(zeroBaseline).toBeNull();
  });

  it('renders diverging bi-directional chart with negative and positive values', () => {
    const data = [
      { bucket_date: '2026-03-01', lines_added: 1200, lines_deleted: -450, cum_net: 750 },
      { bucket_date: '2026-03-02', lines_added: 800, lines_deleted: -1100, cum_net: 450 },
    ];

    const { container } = render(
      <TrendChart
        title="Delivered Code Line Churn & Net Growth"
        data={data}
        divergingZeroBaseline={true}
        integerTicks={true}
        yAxisFormatter={(v) => formatLinesCompact(v)}
        valueFormatter={(v) => `${v > 0 ? '+' : ''}${v.toLocaleString()} lines`}
        series={[
          { key: 'lines_added', label: 'Lines Added (+)', color: '#10b981', type: 'bar' },
          { key: 'lines_deleted', label: 'Lines Deleted (-)', color: '#ef4444', type: 'bar' },
          { key: 'cum_net', label: 'Cumulative Net Growth', color: '#3b82f6', type: 'line' },
        ]}
      />
    );

    expect(screen.getByText('Delivered Code Line Churn & Net Growth')).toBeInTheDocument();

    // Verify prominent zero baseline exists
    const zeroBaseline = container.querySelector('line.chart-zero-baseline');
    expect(zeroBaseline).toBeInTheDocument();
    const baselineY = parseFloat(zeroBaseline?.getAttribute('y1') ?? '0');
    expect(baselineY).toBeGreaterThan(0);

    // Verify bars (2 per bucket = 4 total)
    const bars = container.querySelectorAll('rect[fill-opacity="0.8"]');
    expect(bars.length).toBe(4);

    // Find added bars (green, upward from baseline) and deleted bars (red, downward from baseline)
    const addedBars = Array.from(bars).filter((b) => b.getAttribute('fill') === '#10b981');
    const deletedBars = Array.from(bars).filter((b) => b.getAttribute('fill') === '#ef4444');

    expect(addedBars.length).toBe(2);
    expect(deletedBars.length).toBe(2);

    for (const bar of addedBars) {
      const y = parseFloat(bar.getAttribute('y') ?? '0');
      const height = parseFloat(bar.getAttribute('height') ?? '0');
      // Upward bar: top is y, bottom is baselineY = y + height
      expect(y).toBeLessThan(baselineY);
      expect(Math.abs(y + height - baselineY)).toBeLessThan(1.0);
    }

    for (const bar of deletedBars) {
      const y = parseFloat(bar.getAttribute('y') ?? '0');
      const height = parseFloat(bar.getAttribute('height') ?? '0');
      // Downward bar: top is baselineY = y, bottom is y + height
      expect(Math.abs(y - baselineY)).toBeLessThan(1.0);
      expect(height).toBeGreaterThan(0);
    }

    // Verify Y-axis tick labels format via formatLinesCompact and include 0
    const axisLabels = Array.from(
      container.querySelectorAll('.chart-grid-group text.chart-axis-label')
    ).map((el) => el.textContent);
    expect(axisLabels).toContain('0');
    // Labels should have minus prefixes for negative ticks and compact notation for large values
    const hasMinus = axisLabels.some((l) => l?.startsWith('-'));
    const hasCompact = axisLabels.some((l) => l?.endsWith('k'));
    expect(hasMinus).toBe(true);
    expect(hasCompact).toBe(true);
  });

  it('renders tooltip on hover', () => {
    const data = [
      { bucket_date: '2026-03-01', val: 100 },
      { bucket_date: '2026-03-02', val: 200 },
    ];

    const { container } = render(
      <TrendChart
        title="Interactive Chart"
        data={data}
        series={[{ key: 'val', label: 'Value', color: '#3b82f6', type: 'line' }]}
        valueFormatter={(v) => `${v} units`}
      />
    );

    const hoverColumns = container.querySelectorAll('rect[fill="transparent"]');
    expect(hoverColumns.length).toBe(2);

    fireEvent.mouseEnter(hoverColumns[1]!);
    expect(container.querySelector('.chart-tooltip')).toBeInTheDocument();
    expect(screen.getByText('200 units')).toBeInTheDocument();

    fireEvent.mouseLeave(hoverColumns[1]!);
    expect(container.querySelector('.chart-tooltip')).toBeNull();
  });
});
