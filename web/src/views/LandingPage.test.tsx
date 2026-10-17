import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { LandingPage } from './LandingPage';
import * as apiClient from '../api/client';
import { mockFleetSummary, mockTicketComparisonRows } from '../api/mockData';

describe('LandingPage View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders loading indicator initially then displays hero banner and descriptions', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const onNavigate = vi.fn();
    const onSelectTicket = vi.fn();

    render(
      <LandingPage
        onNavigate={onNavigate}
        onSelectTicket={onSelectTicket}
      />
    );

    expect(screen.getByText(/loading observability portal/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });

    expect(screen.getByText('Autonomous Multi-Agent Telemetry')).toBeInTheDocument();
    expect(
      screen.getByText(/real-time telemetry, context health tracking, and unit economics/i)
    ).toBeInTheDocument();
    expect(
      screen.getByText(/functional core, imperative shell \(fcis\) architecture/i)
    ).toBeInTheDocument();
  });

  it('renders all four feature navigation cards with titles and action buttons', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const onNavigate = vi.fn();
    const onSelectTicket = vi.fn();

    render(
      <LandingPage
        onNavigate={onNavigate}
        onSelectTicket={onSelectTicket}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });

    expect(screen.getByRole('heading', { level: 2, name: 'Fleet Overview' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Ticket Deep Dive' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Metric Progression' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: 'Ticket Comparisons' })).toBeInTheDocument();

    expect(screen.getByRole('button', { name: /open fleet overview/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /explore deep dive/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /view progression/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /compare tickets/i })).toBeInTheDocument();
  });

  it('triggers onNavigate with the expected tab identifier when clicking feature cards', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const onNavigate = vi.fn();
    const onSelectTicket = vi.fn();

    render(
      <LandingPage
        onNavigate={onNavigate}
        onSelectTicket={onSelectTicket}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: /open fleet overview/i }));
    expect(onNavigate).toHaveBeenCalledWith('overview');

    fireEvent.click(screen.getByRole('button', { name: /explore deep dive/i }));
    expect(onNavigate).toHaveBeenCalledWith('deep-dive');

    fireEvent.click(screen.getByRole('button', { name: /view progression/i }));
    expect(onNavigate).toHaveBeenCalledWith('progression');

    fireEvent.click(screen.getByRole('button', { name: /compare tickets/i }));
    expect(onNavigate).toHaveBeenCalledWith('comparisons');
  });

  it('correctly calculates and renders pulse metrics without NaN or null', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const { container } = render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Total Ticket Workspaces')).toBeInTheDocument();
    });

    expect(screen.getByText('Cumulative Fleet Spend')).toBeInTheDocument();
    expect(screen.getByText('Total Iterations')).toBeInTheDocument();
    expect(screen.getByText('Cache Hit Efficiency')).toBeInTheDocument();

    expect(screen.getByText('$14.82')).toBeInTheDocument();
    expect(screen.getByText('78.8%')).toBeInTheDocument();

    const renderedText = container.textContent ?? '';
    expect(renderedText).not.toContain('NaN');
    expect(renderedText).not.toContain('null');
    expect(renderedText).not.toContain('undefined');
  });

  it('renders pulse metrics correctly with zero values and zero-division defenses', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: {
        total_tickets: 0,
        total_cost_usd: 0,
        total_invocations: 0,
        total_input_tokens: 0,
        total_output_tokens: 0,
        total_cache_read_tokens: 0,
        overall_cache_hit_rate_pct: 0,
        total_duration_seconds: 0,
        formatted_total_duration: '0s',
        total_lines_added: 0,
        total_lines_deleted: 0,
        total_files_changed: 0,
        total_reasoning_tokens: 0,
        overall_reasoning_ratio_pct: 0,
        role_costs: {},
        model_costs: {},
      },
      isMock: false,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [],
      isMock: false,
    });

    const { container } = render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Total Ticket Workspaces')).toBeInTheDocument();
    });

    expect(screen.getByText('$0.00')).toBeInTheDocument();
    expect(screen.getByText('0.0%')).toBeInTheDocument();

    const renderedText = container.textContent ?? '';
    expect(renderedText).not.toContain('NaN');
    expect(renderedText).not.toContain('null');
    expect(renderedText).not.toContain('undefined');
    expect(renderedText).not.toContain('Infinity');
  });

  it('renders pulse metrics with fallback calculations when summary is null and handles non-finite iteration counts', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: null as unknown as typeof mockFleetSummary,
      isMock: false,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [
        {
          ticket_id: 'ZERO-1',
          ticket_title: 'Zero Iteration Ticket',
          updated_at: '2026-10-17T01:00:00Z',
          started_at: '2026-10-17T00:00:00Z',
          completed_at: '2026-10-17T01:00:00Z',
          cost_usd: 0,
          active_duration_seconds: 0,
          active_duration_formatted: '0s',
          lead_duration_seconds: 3600,
          lead_duration_formatted: '1h',
          sdlc_iterations: 0,
          cache_hit_rate_pct: 0,
          tool_success_rate_pct: 0,
          test_to_code_ratio: '0.00',
          lines_added: 0,
          lines_deleted: 0,
          files_changed: 0,
          orchestrator_saturation_pct: 0,
          peak_context_tokens: 0,
          context_health_status: 'HEALTHY',
        },
        {
          ticket_id: 'NAN-1',
          ticket_title: 'Corrupted Iteration Ticket',
          updated_at: '2026-10-17T01:00:00Z',
          started_at: '2026-10-17T00:00:00Z',
          completed_at: '2026-10-17T01:00:00Z',
          cost_usd: 0,
          active_duration_seconds: 0,
          active_duration_formatted: '0s',
          lead_duration_seconds: 3600,
          lead_duration_formatted: '1h',
          sdlc_iterations: Number.NaN,
          cache_hit_rate_pct: 0,
          tool_success_rate_pct: 0,
          test_to_code_ratio: '0.00',
          lines_added: 0,
          lines_deleted: 0,
          files_changed: 0,
          orchestrator_saturation_pct: 0,
          peak_context_tokens: 0,
          context_health_status: 'HEALTHY',
        },
      ],
      isMock: false,
    });

    const { container } = render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('ZERO-1')).toBeInTheDocument();
    });

    // Falls back to workspaces.length (2)
    const workspacesCard = screen.getByText('Total Ticket Workspaces').closest('.pulse-card');
    expect(workspacesCard).not.toBeNull();
    expect(workspacesCard?.textContent).toContain('2');

    // Falls back to '$0.00' and '0.0%'
    expect(screen.getByText('$0.00')).toBeInTheDocument();
    expect(screen.getByText('0.0%')).toBeInTheDocument();

    const renderedText = container.textContent ?? '';
    expect(renderedText).not.toContain('NaN');
    expect(renderedText).not.toContain('null');
    expect(renderedText).not.toContain('undefined');
  });

  it('renders quick-launch workspace list and triggers onSelectTicket on click', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    const onSelectTicket = vi.fn();

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={onSelectTicket}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('PO-1')).toBeInTheDocument();
    expect(screen.getByText('GEN-1')).toBeInTheDocument();

    fireEvent.click(screen.getByText('AL-1'));
    expect(onSelectTicket).toHaveBeenCalledWith('AL-1');

    const inspectBtn = screen.getByRole('button', { name: /open deep dive for AL-2/i });
    fireEvent.click(inspectBtn);
    expect(onSelectTicket).toHaveBeenCalledWith('AL-2');
    // Ensure stopPropagation prevented double execution
    expect(onSelectTicket).toHaveBeenCalledTimes(2);
  });

  it('notifies onMockChange callback with true when either endpoint falls back to mock', async () => {
    const onMockChange = vi.fn();
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: false,
    });

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
        onMockChange={onMockChange}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });

    expect(onMockChange).toHaveBeenCalledWith(true);
  });

  it('renders error banner and recovers on retry', async () => {
    const fetchSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockRejectedValueOnce(new Error('Network failure'));
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Network failure')).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    expect(retryBtn).toBeInTheDocument();

    fetchSpy.mockResolvedValueOnce({
      data: mockFleetSummary,
      isMock: true,
    });

    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });
  });

  it('handles non-Error rejection and displays fallback error message', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockRejectedValueOnce('raw-timeout-string');
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Failed to retrieve telemetry data')).toBeInTheDocument();
    });

    expect(screen.getByRole('button', { name: /retry connection/i })).toBeInTheDocument();
  });

  it('renders empty state notice when zero workspaces are returned', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: { ...mockFleetSummary, total_tickets: 0 },
      isMock: false,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [],
      isMock: false,
    });

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Zero Workspaces Discovered')).toBeInTheDocument();
    });
  });

  it('handles unexpected context health status with fallback class without throwing', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [
        {
          ...mockTicketComparisonRows[0],
          ticket_id: 'CUSTOM-STATUS-1',
          context_health_status: 'UNRECOGNIZED_STATUS' as unknown as 'HEALTHY',
        },
      ],
      isMock: true,
    });

    const { container } = render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('CUSTOM-STATUS-1')).toBeInTheDocument();
    });

    const badge = container.querySelector('.health-badge');
    expect(badge).not.toBeNull();
    expect(badge).toHaveClass('HEALTHY');
  });

  it('preserves error banner and allows repeated retry attempts when retry also fails', async () => {
    const fetchSpy = vi.spyOn(apiClient, 'fetchFleetSummary')
      .mockRejectedValueOnce(new Error('First failure'))
      .mockRejectedValueOnce(new Error('Second failure'));

    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
      isMock: true,
    });

    render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('First failure')).toBeInTheDocument();
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('Second failure')).toBeInTheDocument();
    });

    expect(screen.getByRole('button', { name: /retry connection/i })).toBeInTheDocument();
    expect(fetchSpy).toHaveBeenCalledTimes(2);
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

    const { container } = render(
      <LandingPage
        onNavigate={vi.fn()}
        onSelectTicket={vi.fn()}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('AgenticLens Observability')).toBeInTheDocument();
    });

    const renderedText = container.textContent ?? '';
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(renderedText)).toBe(false);
  });
});
