import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { TicketDeepDive } from './TicketDeepDive';
import * as apiClient from '../api/client';
import {
  mockAgenticMetricsLedgers,
  mockWorkspaceDetails,
} from '../api/mockData';

describe('TicketDeepDive View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders loading state initially and then displays full autopsy KPIs', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: mockWorkspaceDetails['AL-1']!,
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-1" />);

    expect(screen.getByText(/loading workspace autopsy telemetry/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
    });

    // Top KPI cards
    expect(screen.getByText('Turnaround Lead Time')).toBeInTheDocument();
    expect(screen.getByText('Active Compute Duration')).toBeInTheDocument();
    expect(screen.getByText('Total Cost')).toBeInTheDocument();
    expect(screen.getByText('Cache Hit Rate')).toBeInTheDocument();
    expect(screen.getByText('Test:Prod Code Ratio')).toBeInTheDocument();

    // Secondary profile cards
    expect(screen.getByText('Code Delivery')).toBeInTheDocument();
    expect(screen.getByText('Reasoning Tokens')).toBeInTheDocument();
    expect(screen.getByText('Tool Reliability')).toBeInTheDocument();

    // Context Window Saturation
    expect(screen.getByText('Context Window Saturation')).toBeInTheDocument();
    expect(screen.getByText('Ceiling: 200k tokens')).toBeInTheDocument();

    // Subagent Turn Velocity
    expect(screen.getByText('Subagent Turn Velocity')).toBeInTheDocument();

    // Tables
    expect(screen.getByText('Role Resource Consumption')).toBeInTheDocument();
    expect(screen.getByText('Model Resource Consumption')).toBeInTheDocument();

    // Specific roles in role breakdown table
    expect(screen.getAllByText('coder').length).toBeGreaterThan(0);
    expect(screen.getAllByText('tester').length).toBeGreaterThan(0);
    expect(screen.getAllByText('planner').length).toBeGreaterThan(0);
    expect(screen.getAllByText('reviewer').length).toBeGreaterThan(0);

    // Specific models in model breakdown table
    expect(screen.getAllByText('claude-3-5-sonnet').length).toBeGreaterThan(0);

    // Threads
    expect(screen.getByText('Session Execution Threads')).toBeInTheDocument();
    expect(screen.getByText('ses-al1-root')).toBeInTheDocument();

    // Artifacts
    expect(screen.getByText('Workspace Artifacts')).toBeInTheDocument();
    expect(screen.getByText(/Contract: AL-1/i)).toBeInTheDocument();
  });

  it('renders context health indicators and ceiling values correctly for ELEVATED status', async () => {
    const elevatedLedger = {
      ...mockAgenticMetricsLedgers['AL-1']!,
      workflow_efficiency: {
        ...mockAgenticMetricsLedgers['AL-1']!.workflow_efficiency,
        context_window_health: {
          healthy_ceiling_tokens: 200000,
          status: 'ELEVATED' as const,
          orchestrator_peak_tokens: 165000,
          orchestrator_saturation_pct: 82.5,
          subagents_peak_tokens: 140000,
          subagents_peak_role: 'coder',
          compactions: 2,
          peak_by_role: {
            orchestrator: 165000,
            coder: 140000,
          },
        },
      },
    };

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [elevatedLedger],
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: {
        ledger: elevatedLedger,
        artifacts: { ticket_id: 'AL-1', contract_content: null, summary_content: null },
        anomalies: [],
      },
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText('ELEVATED')).toBeInTheDocument();
    });
    expect(screen.getByText('Ceiling: 200k tokens')).toBeInTheDocument();
    expect(screen.getByText('(82.5% of ceiling)')).toBeInTheDocument();
  });

  it('renders context health indicators and ceiling marker correctly for CRITICAL ceiling breach', async () => {
    const criticalLedger = {
      ...mockAgenticMetricsLedgers['AL-1']!,
      ticket_id: 'AL-CRIT',
      workflow_efficiency: {
        ...mockAgenticMetricsLedgers['AL-1']!.workflow_efficiency,
        context_window_health: {
          healthy_ceiling_tokens: 200000,
          status: 'CRITICAL' as const,
          orchestrator_peak_tokens: 240000,
          orchestrator_saturation_pct: 120.0,
          subagents_peak_tokens: 220000,
          subagents_peak_role: 'coder',
          compactions: 5,
          peak_by_role: {
            orchestrator: 240000,
            coder: 220000,
          },
        },
      },
    };

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [criticalLedger],
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: {
        ledger: criticalLedger,
        artifacts: { ticket_id: 'AL-CRIT', contract_content: null, summary_content: null },
        anomalies: [],
      },
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-CRIT" />);

    await waitFor(() => {
      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    });
    expect(screen.getByText('100% Ceiling')).toBeInTheDocument();
    expect(screen.getByText('(120.0% of ceiling)')).toBeInTheDocument();
  });

  it('renders graceful empty state when workspace detail has missing or null ledger', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: Object.values(mockAgenticMetricsLedgers),
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: {
        ledger: null as unknown as typeof mockAgenticMetricsLedgers['AL-1'],
        artifacts: { ticket_id: 'MISSING', contract_content: null, summary_content: null },
        anomalies: [],
      },
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="MISSING" />);

    await waitFor(() => {
      expect(screen.getByText(/zero telemetry workspaces available/i)).toBeInTheDocument();
    });
  });

  it('updates active ticket when selectedTicketId prop changes from parent', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
      const detail = mockWorkspaceDetails[id] ?? {
        ledger: mockAgenticMetricsLedgers[id]!,
        artifacts: { ticket_id: id, contract_content: null, summary_content: null },
      };
      return { data: detail, isMock: true };
    });

    const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    rerender(<TicketDeepDive selectedTicketId="AL-2" />);

    await waitFor(() => {
      expect(fetchByIdSpy).toHaveBeenCalledWith('AL-2');
    });
  });

  it('handles empty subagent turn velocity records gracefully', async () => {
    const noTurnsLedger = {
      ...mockAgenticMetricsLedgers['AL-1']!,
      workflow_efficiency: {
        ...mockAgenticMetricsLedgers['AL-1']!.workflow_efficiency,
        subagent_turn_velocity: {
          avg_turns_per_invocation: 0,
          peak_turns: 0,
          by_role: {},
        },
      },
    };

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [noTurnsLedger],
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: {
        ledger: noTurnsLedger,
        artifacts: { ticket_id: 'AL-1', contract_content: null, summary_content: null },
        anomalies: [],
      },
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText(/no subagent turn records available/i)).toBeInTheDocument();
    });
  });

  it('switches artifact tab between CONTRACT.md and SUMMARY.md', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: mockWorkspaceDetails['AL-1']!,
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText(/Contract: AL-1/i)).toBeInTheDocument();
    });

    const summaryTabBtn = screen.getByRole('button', { name: /summary\.md/i });
    fireEvent.click(summaryTabBtn);

    expect(screen.getByText(/Summary: AL-1 Completed/i)).toBeInTheDocument();
  });

  it('allows user to select another workspace via dropdown', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
      const detail = mockWorkspaceDetails[id] ?? {
        ledger: mockAgenticMetricsLedgers[id]!,
        artifacts: { ticket_id: id, contract_content: null, summary_content: null },
      };
      return { data: detail, isMock: true };
    });

    const onSelectTicket = vi.fn();

    render(<TicketDeepDive selectedTicketId="AL-1" onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const select = screen.getByLabelText(/workspace autopsy/i);
    fireEvent.change(select, { target: { value: 'AL-2' } });

    expect(onSelectTicket).toHaveBeenCalledWith('AL-2');
    await waitFor(() => {
      expect(fetchByIdSpy).toHaveBeenCalledWith('AL-2');
    });
  });

  it('does not re-fetch fleet ledgers when switching tickets after initial catalog load', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    const fetchLedgersSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
      const detail = mockWorkspaceDetails[id] ?? {
        ledger: mockAgenticMetricsLedgers[id]!,
        artifacts: { ticket_id: id, contract_content: null, summary_content: null },
      };
      return { data: detail, isMock: true };
    });

    const onSelectTicket = vi.fn();

    render(<TicketDeepDive selectedTicketId="AL-1" onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    expect(fetchLedgersSpy).toHaveBeenCalledTimes(1);
    expect(fetchByIdSpy).toHaveBeenCalledWith('AL-1');

    const select = screen.getByLabelText(/workspace autopsy/i);
    fireEvent.change(select, { target: { value: 'AL-2' } });

    expect(onSelectTicket).toHaveBeenCalledWith('AL-2');
    await waitFor(() => {
      expect(fetchByIdSpy).toHaveBeenCalledWith('AL-2');
    });

    // Verify fleet ledgers was NOT called again on ticket switch
    expect(fetchLedgersSpy).toHaveBeenCalledTimes(1);
  });

  it('renders empty state when no workspaces are returned', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [],
      isMock: true,
    });

    render(<TicketDeepDive />);

    await waitFor(() => {
      expect(screen.getByText(/zero telemetry workspaces available/i)).toBeInTheDocument();
    });
  });

  it('displays error banner and allows retry when API call fails', async () => {
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockRejectedValueOnce(
      new Error('Failed to load workspaces')
    );

    render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load workspaces')).toBeInTheDocument();
    });

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValueOnce({
      data: Object.values(mockAgenticMetricsLedgers),
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValueOnce({
      data: mockWorkspaceDetails['AL-1']!,
      isMock: true,
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    fireEvent.click(retryBtn);

    await waitFor(() => {
      expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
    });
  });

  it('contains zero emojis in the rendered deep dive document body', async () => {
    const ledgers = Object.values(mockAgenticMetricsLedgers);
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: ledgers,
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: mockWorkspaceDetails['AL-1']!,
      isMock: true,
    });

    const { container } = render(<TicketDeepDive selectedTicketId="AL-1" />);

    await waitFor(() => {
      expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
    });

    const renderedText = container.textContent ?? '';
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(renderedText)).toBe(false);
  });
});
