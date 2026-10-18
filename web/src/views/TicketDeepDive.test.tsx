import * as fs from 'node:fs';
import * as path from 'node:path';
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

  it('renders information tooltips on velocity, table headers, and refined KPI tooltips', async () => {
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
      expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
    });

    // Refined KPI tooltips
    expect(
      screen.getByText(
        'Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel.'
      )
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution.'
      )
    ).toBeInTheDocument();

    // Turn velocity header tooltip
    expect(
      screen.getByRole('button', { name: 'Information about Subagent Turn Velocity' })
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Average and peak conversational turns executed per agent invocation, categorized by subagent persona role.'
      )
    ).toBeInTheDocument();

    // Role table header tooltip
    const roleHeaderBtn = screen.getByRole('button', {
      name: 'Information about Role Resource Consumption',
    });
    expect(roleHeaderBtn).toBeInTheDocument();
    expect(
      screen.getByText(
        'Telemetry breakdown of model invocations, token volume, cache hit rate, and spend by agent persona role.'
      )
    ).toBeInTheDocument();
    const roleTooltipId = roleHeaderBtn.getAttribute('aria-describedby');
    expect(roleTooltipId).toBeTruthy();
    expect(document.getElementById(roleTooltipId!)).toHaveClass('placement-left');

    // Model table header tooltip
    const modelHeaderBtn = screen.getByRole('button', {
      name: 'Information about Model Resource Consumption',
    });
    expect(modelHeaderBtn).toBeInTheDocument();
    expect(
      screen.getByText(
        'Telemetry breakdown of invocations, token volume, cache hit rate, and spend by model family and version.'
      )
    ).toBeInTheDocument();
    const modelTooltipId = modelHeaderBtn.getAttribute('aria-describedby');
    expect(modelTooltipId).toBeTruthy();
    expect(document.getElementById(modelTooltipId!)).toHaveClass('placement-left');

    // Column headers across tables
    const roleColButtons = screen.getAllByRole('button', { name: 'Information about Role' });
    const modelColButtons = screen.getAllByRole('button', { name: 'Information about Model' });
    const invocationsColButtons = screen.getAllByRole('button', { name: 'Information about Invocations' });
    const tokensColButtons = screen.getAllByRole('button', { name: 'Information about Tokens' });
    const cacheHitColButtons = screen.getAllByRole('button', { name: 'Information about Cache Hit' });
    const costColButtons = screen.getAllByRole('button', { name: 'Information about Cost' });

    expect(roleColButtons).toHaveLength(2);
    expect(modelColButtons).toHaveLength(2);
    expect(invocationsColButtons).toHaveLength(3);
    expect(tokensColButtons).toHaveLength(2);
    expect(cacheHitColButtons).toHaveLength(3);
    expect(costColButtons).toHaveLength(3);

    // Assert that table column header tooltips retain placement="bottom" and .placement-bottom class
    const allTableColumnButtons = [
      ...roleColButtons,
      ...modelColButtons,
      ...invocationsColButtons,
      ...tokensColButtons,
      ...cacheHitColButtons,
      ...costColButtons,
    ];
    for (const btn of allTableColumnButtons) {
      const tipId = btn.getAttribute('aria-describedby');
      expect(tipId).toBeTruthy();
      const bubble = document.getElementById(tipId!);
      expect(bubble).toHaveClass('placement-bottom');
      expect(bubble).not.toHaveClass('placement-left');
      expect(bubble).not.toHaveClass('placement-top');
    }

    // Exact explanatory contents for Role Resource Consumption columns
    expect(
      screen.getByText('Agent persona role (orchestrator, coder, tester, reviewer, planner)')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Total count of model invocations initiated by this role')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Total input and output tokens consumed by this role')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Percentage of prompt tokens served from prompt cache for this role')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Total model inference spend in US dollars for this role')
    ).toBeInTheDocument();

    // Exact explanatory contents for Model Resource Consumption columns
    expect(screen.getByText('Model identifier and version invoked')).toBeInTheDocument();
    expect(screen.getByText('Total count of invocations executed on this model')).toBeInTheDocument();
    expect(
      screen.getByText('Total input and output tokens consumed by this model')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Percentage of prompt tokens served from prompt cache for this model')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Total model inference spend in US dollars for this model')
    ).toBeInTheDocument();
  });

  it('renders cleanly when ledger has zero roles, zero models, and empty threads with all tooltips intact', async () => {
    const emptyLedger = {
      ...mockAgenticMetricsLedgers['AL-1']!,
      ticket_id: 'AL-EMPTY',
      ticket_title: 'Empty Workspace Telemetry',
      active_duration_seconds: 0,
      active_duration_formatted: '0s',
      lead_duration_seconds: 0,
      lead_duration_formatted: '0s',
      cost_usd: 0,
      roles: [],
      models: [],
      threads: [],
    };

    const emptyWorkspace = {
      ledger: emptyLedger,
      artifacts: {
        ticket_id: 'AL-EMPTY',
        contract_content: null,
        summary_content: null,
      },
      anomalies: [],
    };

    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
      data: [emptyLedger],
      isMock: true,
    });
    vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
      data: emptyWorkspace,
      isMock: true,
    });

    render(<TicketDeepDive selectedTicketId="AL-EMPTY" />);

    await waitFor(() => {
      expect(screen.getByText('Empty Workspace Telemetry')).toBeInTheDocument();
    });

    // Check KPI tooltips intact
    expect(
      screen.getByText(
        'Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel.'
      )
    ).toBeInTheDocument();
    expect(
      screen.getByText(
        'Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution.'
      )
    ).toBeInTheDocument();

    // Turn Velocity tooltip intact
    expect(
      screen.getByRole('button', { name: 'Information about Subagent Turn Velocity' })
    ).toBeInTheDocument();

    // Table title tooltips intact with placement-left
    const emptyRoleHeaderBtn = screen.getByRole('button', {
      name: 'Information about Role Resource Consumption',
    });
    expect(emptyRoleHeaderBtn).toBeInTheDocument();
    const emptyRoleTooltipId = emptyRoleHeaderBtn.getAttribute('aria-describedby');
    expect(emptyRoleTooltipId).toBeTruthy();
    expect(document.getElementById(emptyRoleTooltipId!)).toHaveClass('placement-left');

    const emptyModelHeaderBtn = screen.getByRole('button', {
      name: 'Information about Model Resource Consumption',
    });
    expect(emptyModelHeaderBtn).toBeInTheDocument();
    const emptyModelTooltipId = emptyModelHeaderBtn.getAttribute('aria-describedby');
    expect(emptyModelTooltipId).toBeTruthy();
    expect(document.getElementById(emptyModelTooltipId!)).toHaveClass('placement-left');

    // Column headers intact even with 0 rows
    const emptyRoleButtons = screen.getAllByRole('button', { name: 'Information about Role' });
    const emptyModelButtons = screen.getAllByRole('button', { name: 'Information about Model' });
    const emptyTokensButtons = screen.getAllByRole('button', { name: 'Information about Tokens' });
    const emptyInvocationsButtons = screen.getAllByRole('button', { name: 'Information about Invocations' });
    const emptyCacheHitButtons = screen.getAllByRole('button', { name: 'Information about Cache Hit' });
    const emptyCostButtons = screen.getAllByRole('button', { name: 'Information about Cost' });

    expect(emptyRoleButtons).toHaveLength(1);
    expect(emptyModelButtons).toHaveLength(1);
    expect(emptyTokensButtons).toHaveLength(2);
    expect(emptyInvocationsButtons).toHaveLength(2);
    expect(emptyCacheHitButtons).toHaveLength(2);
    expect(emptyCostButtons).toHaveLength(2);

    const emptyAllColButtons = [
      ...emptyRoleButtons,
      ...emptyModelButtons,
      ...emptyTokensButtons,
      ...emptyInvocationsButtons,
      ...emptyCacheHitButtons,
      ...emptyCostButtons,
    ];
    for (const btn of emptyAllColButtons) {
      const tipId = btn.getAttribute('aria-describedby');
      expect(tipId).toBeTruthy();
      const bubble = document.getElementById(tipId!);
      expect(bubble).toHaveClass('placement-bottom');
      expect(bubble).not.toHaveClass('placement-left');
    }

    // ThreadViewer empty notice and header tooltip intact
    expect(
      screen.getByText('No session execution threads recorded for this workspace.')
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'Information about Session Execution Threads' })
    ).toBeInTheDocument();
  });

  it('handles hover and focus on Role and Model Resource Consumption header tooltips with placement-left', async () => {
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
      expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
    });

    const roleHeaderBtn = screen.getByRole('button', {
      name: 'Information about Role Resource Consumption',
    });
    const roleBubble = document.getElementById(roleHeaderBtn.getAttribute('aria-describedby')!);
    expect(roleBubble).toHaveClass('placement-left');
    expect(roleBubble).not.toHaveClass('is-visible');

    // Mouse hover reveals
    fireEvent.mouseEnter(roleHeaderBtn);
    expect(roleBubble).toHaveClass('is-visible');

    fireEvent.mouseLeave(roleHeaderBtn);
    expect(roleBubble).not.toHaveClass('is-visible');

    // Keyboard focus reveals
    fireEvent.focus(roleHeaderBtn);
    expect(roleBubble).toHaveClass('is-visible');

    // Escape dismisses
    fireEvent.keyDown(roleHeaderBtn, { key: 'Escape' });
    expect(roleBubble).not.toHaveClass('is-visible');

    // Model Resource Consumption tooltip focus
    const modelHeaderBtn = screen.getByRole('button', {
      name: 'Information about Model Resource Consumption',
    });
    const modelBubble = document.getElementById(modelHeaderBtn.getAttribute('aria-describedby')!);
    expect(modelBubble).toHaveClass('placement-left');
    expect(modelBubble).not.toHaveClass('is-visible');

    fireEvent.focus(modelHeaderBtn);
    expect(modelBubble).toHaveClass('is-visible');

    fireEvent.blur(modelHeaderBtn);
    expect(modelBubble).not.toHaveClass('is-visible');
  });

  it('ensures .table-card and .thread-accordion-item in web/src/index.css have overflow: visible to prevent tooltip clipping', () => {
    const cssPath = path.resolve(__dirname, '../index.css');
    const cssContent = fs.readFileSync(cssPath, 'utf-8');

    // Verify .table-card rule explicitly defines overflow: visible and not overflow: hidden
    const tableCardBlock = cssContent.match(/\.table-card\s*\{[^}]+\}/)?.[0] ?? '';
    expect(tableCardBlock).toMatch(/overflow:\s*visible;/);
    expect(tableCardBlock).not.toMatch(/overflow:\s*hidden;/);

    // Verify .thread-accordion-item rule explicitly defines overflow: visible and not overflow: hidden
    const threadAccordionBlock = cssContent.match(/\.thread-accordion-item\s*\{[^}]+\}/)?.[0] ?? '';
    expect(threadAccordionBlock).toMatch(/overflow:\s*visible;/);
    expect(threadAccordionBlock).not.toMatch(/overflow:\s*hidden;/);
  });

  it('ensures .table-responsive, .data-table, and .thread-models-table rules in web/src/index.css satisfy responsive constraints', () => {
    const cssPath = path.resolve(__dirname, '../index.css');
    const cssContent = fs.readFileSync(cssPath, 'utf-8');

    // Verify .table-responsive utility definition
    const tableResponsiveBlock = cssContent.match(/\.table-responsive\s*\{[^}]+\}/)?.[0] ?? '';
    expect(tableResponsiveBlock).toMatch(/width:\s*100%;/);
    expect(tableResponsiveBlock).toMatch(/overflow-x:\s*auto;/);
    expect(tableResponsiveBlock).toMatch(/-webkit-overflow-scrolling:\s*touch;/);

    // Verify .data-table min-width and nowrap td
    const dataTableBlock = cssContent.match(/\.data-table\s*\{[^}]+\}/)?.[0] ?? '';
    expect(dataTableBlock).toMatch(/min-width:\s*480px;/);

    const dataTableTdBlock = cssContent.match(/\.data-table\s+td\s*\{[^}]+\}/)?.[0] ?? '';
    expect(dataTableTdBlock).toMatch(/white-space:\s*nowrap;/);

    // Verify .thread-models-table min-width and nowrap td
    const threadModelsTableBlock = cssContent.match(/\.thread-models-table\s*\{[^}]+\}/)?.[0] ?? '';
    expect(threadModelsTableBlock).toMatch(/min-width:\s*850px;/);

    const threadModelsTableTdBlock = cssContent.match(/\.thread-models-table\s+td\s*\{[^}]+\}/)?.[0] ?? '';
    expect(threadModelsTableTdBlock).toMatch(/white-space:\s*nowrap;/);
  });

  it('wraps Role and Model Resource Consumption tables in .table-responsive containers and keeps headers outside', async () => {
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

    const dataTables = container.querySelectorAll('table.data-table');
    expect(dataTables.length).toBe(2);

    dataTables.forEach((table) => {
      const parent = table.parentElement;
      expect(parent).not.toBeNull();
      expect(parent).toHaveClass('table-responsive');

      const card = table.closest('.table-card');
      expect(card).not.toBeNull();
      const cardHeader = card!.querySelector('.table-card-header');
      expect(cardHeader).not.toBeNull();
      expect(parent!.contains(cardHeader)).toBe(false);

      const columnBubbles = table.querySelectorAll('.info-tooltip-bubble');
      expect(columnBubbles.length).toBeGreaterThan(0);
      columnBubbles.forEach((bubble) => {
        expect(bubble).toHaveClass('placement-bottom');
      });
    });

    const roleHeaderBtn = screen.getByRole('button', {
      name: 'Information about Role Resource Consumption',
    });
    const roleBubble = document.getElementById(roleHeaderBtn.getAttribute('aria-describedby')!);
    expect(roleBubble).toHaveClass('placement-left');

    const modelHeaderBtn = screen.getByRole('button', {
      name: 'Information about Model Resource Consumption',
    });
    const modelBubble = document.getElementById(modelHeaderBtn.getAttribute('aria-describedby')!);
    expect(modelBubble).toHaveClass('placement-left');
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
