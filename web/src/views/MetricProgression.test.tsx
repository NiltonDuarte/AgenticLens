import userEvent from '@testing-library/user-event';
import { StrictMode } from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, act, within } from '../test/render';
import { MetricProgression } from './MetricProgression';
import * as apiClient from '../api/client';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
import workspaceWire from '../test/fixtures/progression-workspaces-wire.json';
import type { AgenticMetricsLedger } from '../domain/types';

function deferredLedgers() {
  let resolve!: (result: apiClient.ApiResponse<readonly AgenticMetricsLedger[]>) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<apiClient.ApiResponse<readonly AgenticMetricsLedger[]>>((ok, fail) => {
    resolve = ok;
    reject = fail;
  });
  return { promise, resolve, reject };
}

function kpiValue(title: string): Element | null | undefined {
  return screen.getByText(title).closest('.kpi-card')?.querySelector('.kpi-value');
}
async function selectTagKey(user: ReturnType<typeof userEvent.setup>, key: string): Promise<void> {
  if (!screen.queryByRole('dialog', { name: 'Ticket tag filters' })) {
    await user.click(screen.getByRole('button', { name: 'Ticket tags filter' }));
  }
  await user.selectOptions(screen.getByRole('combobox', { name: 'Tag key' }), key);
}

describe('MetricProgression View Component', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });
    it.each([
        ['success', 'before'],
        ['failure', 'before'],
        ['success', 'after'],
        ['failure', 'after'],
    ] as const)('ignores obsolete %s settling %s the current filter response', async (outcome, order) => {
        const older = deferredLedgers();
        const newer = deferredLedgers();
        const base = mockAgenticMetricsLedgers['AL-1']!;
        const selected = { ...base, facts: { ...base.facts, tags: { team: 'current' } } };
        const obsolete = { ...base, facts: { ...base.facts, tags: { team: 'obsolete' } } };
        const request = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
            .mockReturnValueOnce(older.promise)
            .mockReturnValueOnce(newer.promise);
        render(<MetricProgression />);
        fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
        fireEvent.click(screen.getByRole('checkbox', { name: 'Feature' }));
        expect(request).toHaveBeenCalledTimes(2);

        const settleOlder = () => act(async () => {
            if (outcome === 'success') {
                older.resolve({ data: [obsolete], diagnostics: [{ ticket_id: 'AL-OLD', message: 'Obsolete diagnostic' }] });
            } else {
                older.reject(new Error('Obsolete error'));
            }
        });
        if (order === 'before') {
            await settleOlder();
            expect(screen.getByText(/aggregating longitudinal telemetry progression/i)).toBeInTheDocument();
            expect(screen.queryByText('Total Invocations')).not.toBeInTheDocument();
        }
        await act(async () => newer.resolve({
            data: [selected], diagnostics: [{ ticket_id: 'AL-NEW', message: 'Current diagnostic' }],
        }));
        if (order === 'after') await settleOlder();
        expect(screen.getByText('Total Invocations')).toBeInTheDocument();
        expect(screen.getByRole('status')).toHaveTextContent('AL-NEW: Current diagnostic');
        expect(screen.getByRole('status')).not.toHaveTextContent('Obsolete diagnostic');
        expect(screen.queryByText('Obsolete error')).not.toBeInTheDocument();
        await selectTagKey(userEvent.setup(), 'team');
        expect(screen.getByRole('checkbox', { name: 'Value: current' })).toBeInTheDocument();
        expect(screen.queryByRole('checkbox', { name: 'Value: obsolete' })).not.toBeInTheDocument();
    });

    it('retries the same selection after its failure and ignores the previous mount request', async () => {
        const initialMount = deferredLedgers();
        const activeMount = deferredLedgers();
        const retry = deferredLedgers();
        const request = vi.spyOn(apiClient, 'fetchWorkspaceLedgers')
            .mockReturnValueOnce(initialMount.promise)
            .mockReturnValueOnce(activeMount.promise)
            .mockReturnValueOnce(retry.promise);
        render(<StrictMode><MetricProgression /></StrictMode>);
        expect(request).toHaveBeenCalledTimes(2);
        await act(async () => activeMount.reject(new Error('Current read failed')));
        expect(screen.getByText('Current read failed')).toBeInTheDocument();
        await act(async () => initialMount.reject(new Error('Disposed mount read failed')));
        expect(screen.getByText('Current read failed')).toBeInTheDocument();
        expect(screen.queryByText('Disposed mount read failed')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'Retry Connection' }));
        expect(request.mock.calls).toEqual([['all'], ['all'], ['all']]);
        expect(screen.getByText(/aggregating longitudinal telemetry progression/i)).toBeInTheDocument();
        await act(async () => retry.resolve({ data: [mockAgenticMetricsLedgers['AL-1']!] }));
        expect(screen.getByText('Total Invocations')).toBeInTheDocument();
    });

    it('keeps old preset results for an all-matching filter and moves the range only when the anchor leaves', async () => {
        // The Go handler contract test verifies these exact requests and response rows.
        const allRows = Object.values(workspaceWire);
        const responses = new Map<string, readonly unknown[]>([
            ['/api/workspaces', allRows],
            ['/api/workspaces?prefix=AL', allRows],
            ['/api/workspaces?ticket_type=feature&prefix=AL', [workspaceWire['AL-41']]],
            ['/api/workspaces?window=7d&prefix=AL', []],
        ]);
        const fetcher = vi.fn(async (url: string) => {
            const rows = responses.get(url);
            if (!rows) throw new Error(`Unexpected workspace URL: ${url}`);
            return new Response(JSON.stringify(rows), { headers: { 'Content-Type': 'application/json' } });
        });
        vi.stubGlobal('fetch', fetcher);
        try {
            const { container } = render(<MetricProgression />);
            await screen.findByText('Total Invocations');
            fireEvent.click(screen.getByRole('button', { name: '7 Days' }));
            const assertWindow = (start: string, end: string, buckets: string[], count: string, net: string) => {
                expect([...container.querySelectorAll('time')].map((el) => el.getAttribute('datetime'))).toEqual([start, end]);
                const chart = screen.getByText('Financial Spend & Cumulative Trajectory ($ USD)').closest('.trend-card')!;
                const ticks = [...chart.querySelectorAll('.chart-axis-tick')].map((el) => el.parentElement?.textContent);
                expect(ticks).toEqual(buckets);
                expect(kpiValue('Total Invocations')).toHaveTextContent(new RegExp(`^${count}$`));
                expect(kpiValue('Net Code Delivered')).toHaveTextContent(net);
            };
            assertWindow('2000-02-08', '2000-02-15', ['2000-02-15'], '1', '+100');

            fireEvent.click(screen.getByRole('combobox', { name: 'Prefix selector' }));
            fireEvent.click(screen.getByRole('checkbox', { name: 'AL' }));
            await waitFor(() => expect(fetcher).toHaveBeenLastCalledWith('/api/workspaces?prefix=AL'));
            await act(async () => { await fetcher.mock.results[1]!.value; });
            assertWindow('2000-02-08', '2000-02-15', ['2000-02-15'], '1', '+100');

            fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
            fireEvent.click(screen.getByRole('checkbox', { name: 'Feature' }));
            await waitFor(() => expect(fetcher).toHaveBeenLastCalledWith('/api/workspaces?ticket_type=feature&prefix=AL'));
            await waitFor(() => assertWindow('2000-01-01', '2000-01-08', ['2000-01-01', '2000-01-08'], '2', '+100'));

            fireEvent.click(screen.getByRole('button', { name: 'Custom' }));
            fireEvent.change(screen.getByLabelText('End Date and Time'), { target: { value: '2000-01-01T00:00' } });
            expect(kpiValue('Total Invocations')).toHaveTextContent(/^2$/);
            expect(kpiValue('Net Code Delivered')).toHaveTextContent('+0');
            expect(fetcher).toHaveBeenCalledTimes(3);
        } finally {
            vi.unstubAllGlobals();
        }
    });
    it('reconciles tag choices after filtered refetches without losing outside-cohort values', async () => {
        const user = userEvent.setup();
        const base = mockAgenticMetricsLedgers['AL-1']!;
        const original = { ...base, facts: { ...base.facts, tags: { team: 'core' } } };
        const outside = { ...base, ticket_id: 'AL-outside', facts: { ...base.facts, tags: { team: 'outside' } } };
        const updated = { ...base, facts: { ...base.facts, tags: { team: 'new' } } };
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValueOnce({ data: [original, outside] }).mockResolvedValue({ data: [updated] });
        render(<MetricProgression />);
        await screen.findByRole('button', { name: 'Ticket tags filter' });
        await user.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
        await user.click(screen.getByRole('checkbox', { name: 'Feature' }));
        await waitFor(() => expect(apiClient.fetchWorkspaceLedgers).toHaveBeenCalledTimes(2));
        await selectTagKey(user, 'team');
        expect(screen.getByRole('checkbox', { name: 'Value: new' })).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'Value: outside' })).toBeInTheDocument();
        expect(screen.queryByRole('checkbox', { name: 'Value: core' })).not.toBeInTheDocument();
    });

    it('keeps period subtotal currency compact for the selected invocation buckets', async () => {
        const base = mockAgenticMetricsLedgers['AL-1']!;
        const owner = base.facts.sessions[0]!;
        const invocation = owner.invocations[0]!;
        const completed_at = '2026-01-01T12:00:00Z';
        const ledger: typeof base = { ...base, completed_at, facts: { ...base.facts, sessions: [{ ...owner, invocations: [
            { ...invocation, completed_at, cost_usd: null, cost_evidence: { state: 'partial', known_subtotal_usd: 2 } },
            { ...invocation, completed_at, cost_usd: 0, cost_evidence: { state: 'observed', known_subtotal_usd: 0 } },
        ] }] } };
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: [ledger] });
        render(<MetricProgression />);
        const trigger = await screen.findByRole('button', { name: 'Information about Period Spend' });
        const card = trigger.closest<HTMLElement>('.kpi-card')!;
        expect(card.querySelector('.kpi-value')).toHaveTextContent(/^\$2\.00$/);
        expect(card.querySelector('.kpi-secondary')).toHaveTextContent(/^Known subtotal \(partial\)$/);
        expect(within(card).getByRole('tooltip')).toHaveTextContent('1 fully observed, 1 partial, 0 unavailable of 2 invocations');
    });

    it('shows ticket-specific skipped-workspace diagnostics alongside valid metrics', async () => {
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: Object.values(mockAgenticMetricsLedgers),
            diagnostics: [{ ticket_id: 'AL-OLD', message: 'Incomplete migration must be resumed.' }],
        });
        render(<MetricProgression />);
        expect(await screen.findByText('Some ticket workspaces were skipped.')).toBeInTheDocument();
        expect(screen.getByRole('status').textContent).toContain('AL-OLD: Incomplete migration must be resumed.');
        expect(screen.getByText('Total Invocations')).toBeInTheDocument();
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
        expect(screen.getByText('Financial Spend & Cumulative Trajectory ($ USD)')).toBeInTheDocument();
        expect(screen.getByText('Prompt Cache Hit Rate Trajectory (%)')).toBeInTheDocument();
        expect(screen.getByText('Token Volume Distribution')).toBeInTheDocument();
        expect(screen.getByText('Peak Context Window Saturation vs Ceiling')).toBeInTheDocument();
        expect(screen.getByText('Delivered Code Line Churn & Net Growth')).toBeInTheDocument();
        // Ceiling reference line
        expect(screen.getAllByText(/200k ceiling/i).length).toBeGreaterThan(0);
        // Chart 1
        expect(screen.getByRole('button', {
            name: 'Information about Financial Spend & Cumulative Trajectory ($ USD)',
        })).toBeInTheDocument();
        expect(screen.getByText('Historical model inference spend per time period alongside cumulative trajectory in US dollars.')).toBeInTheDocument();
        // Chart 2
        expect(screen.getByRole('button', {
            name: 'Information about Token Volume Distribution',
        })).toBeInTheDocument();
        expect(screen.getByText('Distribution of token throughput across input prompts, output completions, and key-value prompt cache reads.')).toBeInTheDocument();
        // Chart 3
        expect(screen.getByRole('button', {
            name: 'Information about Prompt Cache Hit Rate Trajectory (%)',
        })).toBeInTheDocument();
        expect(screen.getByText('Percentage of prompt tokens served from prompt cache per bucket against baseline (50%) and target (80%) thresholds.')).toBeInTheDocument();
        // Chart 4
        expect(screen.getByRole('button', {
            name: 'Information about Delivered Code Line Churn & Net Growth',
        })).toBeInTheDocument();
        expect(screen.getByText('Volume of production and test code lines added, deleted, and net code growth delivered per time bucket.')).toBeInTheDocument();
        // Chart 5
        expect(screen.getByRole('button', {
            name: 'Information about Peak Context Window Saturation vs Ceiling',
        })).toBeInTheDocument();
        expect(screen.getByText('Peak recorded context tokens per time bucket. A ceiling is shown only when the selected records provide one common explicit value.')).toBeInTheDocument();
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
                sessions: []
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
            expect(screen.getByText(/no telemetry invocations or deliverable code records found within the selected time window/i)).toBeInTheDocument();
        });
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
    it('renders error banner and allows retry when API call fails', async () => {
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockRejectedValueOnce(new Error('Progression network error'));
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
        expect(screen.getByText(/No workspace telemetry files were found/i)).toBeInTheDocument();
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
        render(<MetricProgression configuredPrefixes={['INITIATIVE', 'CORE']}/>);
        await waitFor(() => {
            expect(screen.getByText('Zero Progression Telemetry Available')).toBeInTheDocument();
        });
        const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
        fireEvent.click(prefixTrigger);
        expect(screen.getByRole('checkbox', { name: 'INITIATIVE' })).toBeInTheDocument();
        expect(screen.getByRole('checkbox', { name: 'CORE' })).toBeInTheDocument();
    });
});
