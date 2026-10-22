import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react';
import { MetricProgression } from './MetricProgression';
import * as apiClient from '../api/client';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';

describe('MetricProgression View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders loading state initially and then displays progression KPIs and trend charts', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    expect(screen.getByText(/aggregating longitudinal telemetry progression/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    expect(screen.getAllByText('Period Spend')[0]).toBeInTheDocument();

    // Summary KPIs
    expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    expect(screen.getByText('Total Token Throughput')).toBeInTheDocument();
    expect(screen.getByText('Window Cache Hit Rate')).toBeInTheDocument();
    expect(screen.getByText('Net Code Delivered')).toBeInTheDocument();

    // 5 SVG Trend Charts
    expect(
      screen.getByText('Financial Spend & Cumulative Trajectory ($ USD)')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Prompt Cache Hit Rate Trajectory (%)')
    ).toBeInTheDocument();
    expect(screen.getByText('Token Volume Distribution')).toBeInTheDocument();
    expect(
      screen.getByText('Peak Context Window Saturation vs Ceiling')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Delivered Code Line Churn & Net Growth')
    ).toBeInTheDocument();

    // Ceiling reference line
    expect(screen.getAllByText(/200k ceiling/i).length).toBeGreaterThan(0);
  });

  it('renders SVG chart elements including lines, paths, bars, and interactive hover tooltips', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Verify 5 SVG trend charts exist in the DOM
    const svgs = container.querySelectorAll('svg.trend-svg');
    expect(svgs.length).toBe(5);

    for (const svg of svgs) {
      expect(svg).toHaveAttribute('viewBox', '0 0 640 260');
      // Grid lines
      expect(svg.querySelectorAll('line.chart-grid-line').length).toBeGreaterThan(0);
    }

    // Verify bar elements exist (Token volume, delivered code, spend)
    const bars = container.querySelectorAll('rect[fill-opacity="0.8"]');
    expect(bars.length).toBeGreaterThan(0);

    // Verify line series path elements exist
    const lines = container.querySelectorAll('path[stroke-width="2.5"]');
    expect(lines.length).toBeGreaterThan(0);

    // Verify data point circles exist
    const circles = container.querySelectorAll('circle');
    expect(circles.length).toBeGreaterThan(0);

    // Verify reference line labels
    expect(screen.getAllByText(/Baseline 50%/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/Target 80%/i).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/200k Ceiling/i).length).toBeGreaterThan(0);

    // Test hover interaction on chart column overlay
    const hoverCols = container.querySelectorAll('rect[fill="transparent"]');
    expect(hoverCols.length).toBeGreaterThan(0);

    // Trigger hover on first column
    fireEvent.mouseEnter(hoverCols[0]!);
    expect(container.querySelector('.chart-tooltip')).toBeInTheDocument();

    // Mouse leave removes tooltip
    fireEvent.mouseLeave(hoverCols[0]!);
    expect(container.querySelector('.chart-tooltip')).not.toBeInTheDocument();
  });

  it('toggles granularity and verifies cumulative option is removed', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Financial Spend & Cumulative Trajectory ($ USD)')).toBeInTheDocument();
    });

    // Verify Cumulative button is strictly not present
    expect(screen.queryByRole('button', { name: /^cumulative$/i })).toBeNull();

    // Toggle to Hourly
    const hourlyBtn = screen.getByRole('button', { name: /^hourly$/i });
    fireEvent.click(hourlyBtn);
    expect(hourlyBtn).toHaveClass('active');

    // Toggle to Weekly
    const weeklyBtn = screen.getByRole('button', { name: /^weekly$/i });
    fireEvent.click(weeklyBtn);
    expect(weeklyBtn).toHaveClass('active');

    // Toggle back to Daily
    const dailyBtn = screen.getByRole('button', { name: /^daily$/i });
    fireEvent.click(dailyBtn);
    expect(dailyBtn).toHaveClass('active');
    expect(screen.getByText('Financial Spend & Cumulative Trajectory ($ USD)')).toBeInTheDocument();

    // Verify Chart 4 has zero baseline and legend labels
    expect(screen.getByText('Delivered Code Line Churn & Net Growth')).toBeInTheDocument();
    expect(screen.getByText('Lines Added (+)')).toBeInTheDocument();
    expect(screen.getByText('Lines Deleted (-)')).toBeInTheDocument();
    expect(screen.getByText('Cumulative Net Growth')).toBeInTheDocument();
    expect(container.querySelector('line.chart-zero-baseline')).toBeInTheDocument();
  });

  it('renders empty notice when ledgers have zero matches in the selected window', async () => {
    // Create a ledger where all timestamps are null or unparseable, resulting in 0 buckets
    const undatedLedger = {
      ...mockAgenticMetricsLedgers['AL-1']!,
      ticket_id: 'UNDATED-1',
      started_at: null,
      completed_at: null,
      updated_at: 'unparseable-date',
      facts: {
        ...mockAgenticMetricsLedgers['AL-1']!.facts,
        started_at: null,
        completed_at: null,
        updated_at: 'unparseable-date',
        invocations: [],
      },
      workflow_efficiency: {
        ...mockAgenticMetricsLedgers['AL-1']!.workflow_efficiency,
        code_delivered: { lines_added: 0, lines_deleted: 0, files_changed: 0 },
      },
    };

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [undatedLedger],
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(
        screen.getByText(
          /no telemetry invocations or deliverable code records found within the selected time window/i
        )
      ).toBeInTheDocument();
    });
  });

  it('toggles granularity between hourly, daily, and weekly', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Verify Cumulative button is absent
    expect(screen.queryByRole('button', { name: /^cumulative$/i })).toBeNull();

    // Toggle to Hourly
    const hourlyBtn = screen.getByRole('button', { name: /^hourly$/i });
    fireEvent.click(hourlyBtn);
    expect(hourlyBtn).toHaveClass('active');

    // Toggle to Weekly
    const weeklyBtn = screen.getByRole('button', { name: /^weekly$/i });
    fireEvent.click(weeklyBtn);
    expect(weeklyBtn).toHaveClass('active');

    // Toggle back to Daily
    const dailyBtn = screen.getByRole('button', { name: /^daily$/i });
    fireEvent.click(dailyBtn);
    expect(dailyBtn).toHaveClass('active');
  });

  it('filters data when selecting time window presets', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    const btn7d = screen.getByRole('button', { name: /7 days/i });
    fireEvent.click(btn7d);
    expect(btn7d).toHaveClass('active');

    const btn30d = screen.getByRole('button', { name: /30 days/i });
    fireEvent.click(btn30d);
    expect(btn30d).toHaveClass('active');

    const btn90d = screen.getByRole('button', { name: /90 days/i });
    fireEvent.click(btn90d);
    expect(btn90d).toHaveClass('active');

    const btnLifetime = screen.getByRole('button', { name: /lifetime/i });
    fireEvent.click(btnLifetime);
    expect(btnLifetime).toHaveClass('active');
  });

  it('renders zero telemetry state when no ledgers exist', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText(/zero progression telemetry available/i)).toBeInTheDocument();
    });
  });

  it('renders error banner and allows retry when API call fails', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockRejectedValueOnce(
      new Error('Progression network error')
    );

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Progression network error')).toBeInTheDocument();
    });

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValueOnce({
      data: Object.values(mockAgenticMetricsLedgers),
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });
  });

  it('renders InfoTooltip triggers and explanations across all five trend charts', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Chart 1
    expect(
      screen.getByRole('button', {
        name: 'Information about Financial Spend & Cumulative Trajectory ($ USD)',
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Historical model inference spend per time period alongside cumulative trajectory in US dollars.'
      )
    ).toBeInTheDocument();

    // Chart 2
    expect(
      screen.getByRole('button', {
        name: 'Information about Token Volume Distribution',
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Distribution of token throughput across input prompts, output completions, and key-value prompt cache reads.'
      )
    ).toBeInTheDocument();

    // Chart 3
    expect(
      screen.getByRole('button', {
        name: 'Information about Prompt Cache Hit Rate Trajectory (%)',
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Percentage of prompt tokens served from prompt cache per bucket against baseline (50%) and target (80%) thresholds.'
      )
    ).toBeInTheDocument();

    // Chart 4
    expect(
      screen.getByRole('button', {
        name: 'Information about Delivered Code Line Churn & Net Growth',
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Volume of production and test code lines added, deleted, and net code growth delivered per time bucket.'
      )
    ).toBeInTheDocument();

    // Chart 5
    expect(
      screen.getByRole('button', {
        name: 'Information about Peak Context Window Saturation vs Ceiling',
      })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Peak active context tokens per time bucket measured against the standard 200,000 token healthy context ceiling.'
      )
    ).toBeInTheDocument();
  });

  it('handles focus and hover transitions on trend chart tooltips in MetricProgression', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    const trigger = screen.getByRole('button', {
      name: 'Information about Token Volume Distribution',
    });
    const tooltipId = trigger.getAttribute('aria-describedby');
    expect(tooltipId).toBeTruthy();
    const bubble = document.getElementById(tooltipId!);

    expect(bubble).not.toHaveClass('is-visible');

    fireEvent.focus(trigger);
    expect(bubble).toHaveClass('is-visible');

    fireEvent.blur(trigger);
    expect(bubble).not.toHaveClass('is-visible');
  });

  it('contains zero emojis in the rendered progression document body', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    const renderedText = container.textContent ?? '';
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(renderedText)).toBe(false);
  });

  it('filters progression data and updates KPIs when ticket type and prefix filters change', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Check initial KPI value: Total Invocations = 14 across all mock ledgers
    const card = screen.getByText('Total Invocations').closest('.kpi-card');
    expect(card).not.toBeNull();
    expect(card?.textContent).toContain('14');

    // Filter by Ticket Type: Feature (AL-1, AL-2, AL-3 = 11 invocations)
    const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(typeTrigger);
    const featureCheckbox = screen.getByRole('checkbox', { name: 'Feature' });
    fireEvent.click(featureCheckbox);

    // Total Invocations should update to 11
    await waitFor(() => {
      expect(card?.textContent).toContain('11');
    });

    // Reset ticket type filter to All
    const allTypeCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allTypeCheckbox);

    await waitFor(() => {
      expect(card?.textContent).toContain('14');
    });

    // Close ticket type dropdown
    fireEvent.click(typeTrigger);

    // Filter by Prefix: PO (PO-1 has 1 invocation)
    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);
    const poCheckbox = screen.getByRole('checkbox', { name: 'PO' });
    fireEvent.click(poCheckbox);

    await waitFor(() => {
      expect(card?.textContent).toContain('1');
    });
  });

  it('toggles 24 Hours, Hourly granularity, and Custom datetime window with validation', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Toggle 24 Hours
    const btn24h = screen.getByRole('button', { name: '24 Hours' });
    expect(btn24h).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(btn24h);
    });
    expect(btn24h).toHaveClass('active');

    // Toggle Hourly Granularity
    const btnHourly = screen.getByRole('button', { name: 'Hourly' });
    expect(btnHourly).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(btnHourly);
    });
    expect(btnHourly).toHaveClass('active');

    // Toggle Custom Window
    const btnCustom = screen.getByRole('button', { name: 'Custom' });
    expect(btnCustom).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(btnCustom);
    });
    expect(btnCustom).toHaveClass('active');

    // Custom datetime inputs should now appear
    const startInput = screen.getByLabelText('Start Date and Time');
    const endInput = screen.getByLabelText('End Date and Time');
    expect(startInput).toBeInTheDocument();
    expect(endInput).toBeInTheDocument();

    // Enter valid date range
    await act(async () => {
      fireEvent.change(startInput, { target: { value: '2026-10-15T09:00' } });
      fireEvent.change(endInput, { target: { value: '2026-10-15T18:00' } });
    });
    expect(screen.queryByText(/Start datetime cannot be after End datetime/i)).not.toBeInTheDocument();

    // Enter inverted date range (start > end)
    await act(async () => {
      fireEvent.change(startInput, { target: { value: '2026-10-15T20:00' } });
    });
    expect(screen.getByText(/Start datetime cannot be after End datetime/i)).toBeInTheDocument();
  });

  it('renders gracefully with zero ledgers displaying empty state without runtime errors', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });

    expect(
      screen.getByText(/No workspace telemetry files were found/i)
    ).toBeInTheDocument();
  });

  it('keeps progression controls bar visible and interactive when zero ledgers match filters', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });

    const controlsBar = container.querySelector('.progression-controls-bar');
    expect(controlsBar).toBeInTheDocument();

    expect(screen.getByRole('combobox', { name: 'Ticket Type selector' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Prefix selector' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Lifetime' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Daily' })).toBeInTheDocument();
  });

  it('allows interactive recovery from empty filter selection without page reload', async () => {
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

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(typeTrigger);
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });
    expect(container.querySelector('.progression-controls-bar')).toBeInTheDocument();

    const allCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allCheckbox);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });
    expect(screen.queryByText('Zero Progression Telemetry Available')).not.toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(3);
  });

  it('displays empty state while keeping controls bar interactive when client-side filter matches zero ledgers', async () => {
    const singleLedger = [mockAgenticMetricsLedgers['AL-1']!];
    const fetchSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: singleLedger,
    });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    // Client-side filter: select 'Bug' which does not match AL-1
    const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(typeTrigger);
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    // Empty state should appear because filteredLedgers is empty
    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });

    // Controls bar must still be present and visible
    const controlsBar = container.querySelector('.progression-controls-bar');
    expect(controlsBar).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Ticket Type selector' })).toBeInTheDocument();
    expect(screen.getByRole('combobox', { name: 'Prefix selector' })).toBeInTheDocument();

    // Changing filter back to All should restore progression view without page reload
    const allCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allCheckbox);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });
    expect(screen.queryByText('Zero Progression Telemetry Available')).not.toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(3);
  });

  it('re-triggers data fetch when prefix filter changes from initial zero ledgers empty state', async () => {
    const fetchSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
      .mockResolvedValueOnce({
        data: [],
        })
      .mockResolvedValueOnce({
        data: Object.values(mockAgenticMetricsLedgers),
        });

    const { container } = render(<MetricProgression />);

    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });

    expect(container.querySelector('.progression-controls-bar')).toBeInTheDocument();

    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);
    const alCheckbox = screen.getByRole('checkbox', { name: 'AL' });
    fireEvent.click(alCheckbox);

    await waitFor(() => {
      expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });
    expect(screen.queryByText('Zero Progression Telemetry Available')).not.toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(2);
    expect(fetchSpy).toHaveBeenLastCalledWith('all', undefined, undefined, undefined, ['AL']);
  });

  it('renders configured prefixes in prefix dropdown even when no ledgers for those prefixes exist', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
    });

    render(<MetricProgression configuredPrefixes={['INITIATIVE', 'CORE']} />);

    await waitFor(() => {
      expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
    });

    const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
    fireEvent.click(prefixTrigger);

    expect(screen.getByRole('checkbox', { name: 'INITIATIVE' })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: 'CORE' })).toBeInTheDocument();
  });
});
