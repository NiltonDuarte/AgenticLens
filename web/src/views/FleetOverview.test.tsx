import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, within, act } from '../test/render';
import userEvent from '@testing-library/user-event';
import { FleetOverview } from './FleetOverview';
import * as apiClient from '../api/client';
import { mockFleetSummary, mockTicketComparisonRows, mockAgenticMetricsLedgers } from '../test/fixtures/mockData';

async function selectTagKey(user: ReturnType<typeof userEvent.setup>, key: string): Promise<void> {
  if (!screen.queryByRole('dialog', { name: 'Ticket tag filters' })) {
    await user.click(screen.getByRole('button', { name: 'Ticket tags filter' }));
  }
  await user.selectOptions(screen.getByRole('combobox', { name: 'Tag key' }), key);
}

describe('FleetOverview View Component', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: [] });
  });

  it('sends ticket tag cohorts to both fleet requests and restores results from empty filters', async () => {
    const user = userEvent.setup();
    const base = mockAgenticMetricsLedgers['AL-1']!;
    vi.mocked(apiClient.fetchWorkspaceLedgers).mockResolvedValue({ data: [
      { ...base, facts: { ...base.facts, tags: { team: 'core', environment: 'dev' } } },
      { ...base, facts: { ...base.facts, tags: { environment: 'prod' } } },
      base,
    ] });
    const fleet = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({ data: mockFleetSummary });
    const workspaces = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });
    render(<FleetOverview />);
    await screen.findByText('Total Cost');
    await selectTagKey(user, 'team');
    await user.click(screen.getByRole('checkbox', { name: 'Value: core' }));
    const filters = [{ key: 'team', values: ['core'], missing: false }];
    await waitFor(() => expect(fleet).toHaveBeenLastCalledWith('all', undefined, undefined, undefined, undefined, filters));
    expect(workspaces).toHaveBeenLastCalledWith('all', '', '', undefined, undefined, undefined, undefined, filters);
    fleet.mockResolvedValue({ data: { ...mockFleetSummary, total_tickets: 0 } });
    workspaces.mockResolvedValue({ data: [] });
    await selectTagKey(user, 'environment');
    await user.click(screen.getByRole('checkbox', { name: 'Value: prod' }));
    expect(await screen.findByText('No tickets match the selected tags. Clear tag filters to restore all tickets.')).toBeInTheDocument();
    fleet.mockResolvedValue({ data: mockFleetSummary });
    workspaces.mockResolvedValue({ data: [...mockTicketComparisonRows] });
    await user.click(screen.getByRole('button', { name: 'Clear all' }));
    expect(await screen.findByText('Total Cost')).toBeInTheDocument();
    expect(fleet).toHaveBeenLastCalledWith('all', undefined, undefined);
    await selectTagKey(user, 'team');
    expect(screen.getByRole('checkbox', { name: 'Value: core' })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: 'Missing (key absent)' })).toBeInTheDocument();
  });

  it('keeps the fleet cost primary compact and the partial scope in secondary and tooltip', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({ data: { ...mockFleetSummary, total_cost_usd: null, total_invocations: 3,
      cost_summary: { state: 'partial', known_subtotal_usd: 2, fully_observed: 1, partial: 1, unavailable: 1, total: 3 } } });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });
    render(<FleetOverview />);
    const card = (await screen.findByText('Total Cost')).closest<HTMLElement>('.kpi-card')!;
    expect(card.querySelector('.kpi-value')).toHaveTextContent(/^\$2\.00$/);
    expect(card.querySelector('.kpi-secondary')).toHaveTextContent(`Known subtotal (partial) · ${mockFleetSummary.total_tickets} active tickets`);
    expect(within(card).getByRole('tooltip')).toHaveTextContent('1 fully observed, 1 partial, 1 unavailable of 3 invocations');
  });

  it('shows ranked activation counts, then refreshes safely without skill facts', async () => {
    const fetchFleet = vi.spyOn(apiClient, 'fetchFleetSummary')
      .mockResolvedValueOnce({ data: { ...mockFleetSummary, skill_adoption: [
        { name: 'alpha', total_invocations: 8, tickets_count: 2 },
        { name: 'beta', total_invocations: 1, tickets_count: 1 },
        { name: 'gamma', total_invocations: 1, tickets_count: 1 },
        { name: 'omega', total_invocations: 1, tickets_count: 1 },
      ] } })
      .mockResolvedValueOnce({ data: mockFleetSummary })
      .mockResolvedValue({ data: { ...mockFleetSummary, skill_adoption: [] } });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });
    render(<FleetOverview />);
    const card = await screen.findByRole('region', { name: 'Skill Adoption' });
    expect(within(card).getByText('4 recorded skills')).toBeInTheDocument();
    expect(within(card).getByText('11 activations')).toBeInTheDocument();
    expect(within(card).queryByText(/success rate/i)).not.toBeInTheDocument();
    expect(within(card).getAllByRole('listitem').map((item) => item.textContent)).toEqual([
      'alpha: 8 invocations', 'beta: 1 invocations', 'gamma: 1 invocations',
    ]);
    fireEvent.change(screen.getByLabelText('Date Range Preset'), { target: { value: '7d' } });
    await waitFor(() => expect(fetchFleet).toHaveBeenLastCalledWith('7d', undefined, undefined));
    const emptyCard = await screen.findByRole('region', { name: 'Skill Adoption' });
    expect(await within(emptyCard).findByText('Not recorded')).toBeInTheDocument();
    expect(within(emptyCard).queryByText('0 activations')).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('Date Range Preset'), { target: { value: '30d' } });
    await waitFor(() => expect(fetchFleet).toHaveBeenLastCalledWith('30d', undefined, undefined));
    expect(await screen.findByText('0 recorded skills')).toBeInTheDocument();
    const recordedEmptyCard = screen.getByRole('region', { name: 'Skill Adoption' });
    expect(within(recordedEmptyCard).getByText('No recorded skills')).toBeInTheDocument();
    expect(within(recordedEmptyCard).getByText('0 activations')).toBeInTheDocument();
  });

  it('keeps recorded activations visible without outcome analytics', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({ data: { ...mockFleetSummary, skill_adoption: [
      { name: 'selected-skill', total_invocations: 3, tickets_count: 1 },
    ] } });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });
    render(<FleetOverview />);
    const card = await screen.findByRole('region', { name: 'Skill Adoption' });
    expect(within(card).getByText('1 recorded skills')).toBeVisible();
    expect(within(card).getByText('selected-skill: 3 invocations')).toBeVisible();
    expect(within(card).getByText('3 activations')).toBeVisible();
    expect(within(card).queryByText(/success rate/i)).not.toBeInTheDocument();
  });

  it('shows ticket-specific skipped-workspace diagnostics alongside valid workspace data', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
      diagnostics: [{ ticket_id: 'AL-OLD', message: 'Incomplete migration must be resumed.' }],
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });

    render(<FleetOverview />);

    expect(await screen.findByText('Some ticket workspaces were skipped.')).toBeInTheDocument();
    expect(screen.getByRole('status').textContent).toContain('AL-OLD: Incomplete migration must be resumed.');
    expect(screen.getByText('Total Cost')).toBeInTheDocument();
  });

  it('renders loading indicator initially then displays KPI values and tables', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
    });

    render(<FleetOverview />);

    expect(screen.getByText(/loading telemetry metrics/i)).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Total Cost')).toBeInTheDocument();
    });

    expect(screen.getByText('$14.82')).toBeInTheDocument();
    expect(screen.getByText('128')).toBeInTheDocument();
    expect(screen.getByText('78.8%')).toBeInTheDocument();
    expect(screen.getByText('2h 3m')).toBeInTheDocument();

    expect(screen.getByText('Role Cost Distribution')).toBeInTheDocument();
    expect(screen.getByText('Model Cost Distribution')).toBeInTheDocument();

    expect(screen.getByText('AL-1')).toBeInTheDocument();
    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.getByText('PO-1')).toBeInTheDocument();
    expect(screen.getByText('GEN-1')).toBeInTheDocument();
  });

  it('filters rows when user enters query in ticket table search', async () => {
    vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
      data: mockFleetSummary,
    });
    vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
      data: [...mockTicketComparisonRows],
    });

    render(<FleetOverview />);

    await waitFor(() => {
      expect(screen.getByText('AL-1')).toBeInTheDocument();
    });

    const searchInput = screen.getByLabelText(/filter tickets/i);
    fireEvent.change(searchInput, { target: { value: 'AL-2' } });

    expect(screen.getByText('AL-2')).toBeInTheDocument();
    expect(screen.queryByText('AL-1')).not.toBeInTheDocument();
    expect(screen.queryByText('PO-1')).not.toBeInTheDocument();
  });

  describe('Shared recorded sessions', () => {
    const sessionId = `stored-shared-native-session-${'long-native-id-'.repeat(24)}end`;
    const durationDisclosure = 'Recorded duration including 10s provisional';
    const base = mockAgenticMetricsLedgers['AL-1']!;
    const ledgers = ['HIDDEN-1', 'HIDDEN-2'].map((ticketId) => ({
      ...base, ticket_id: ticketId, ticket_title: `Recorded ${ticketId}`,
      facts: { ...base.facts, ticket_id: ticketId, sessions: [{ id: sessionId, source: 'codex' as const, invocations: [], tools: [] }] },
    }));

    beforeEach(() => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({ data: {
        ...mockFleetSummary,
        provisional_seconds: 10,
      } });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({ data: [...mockTicketComparisonRows] });
    });

    it('shows conflicts outside metric and table filters, navigates, and clears after corrected reload', async () => {
      const inspect = vi.mocked(apiClient.fetchWorkspaceLedgers).mockResolvedValueOnce({ data: ledgers }).mockResolvedValue({ data: [ledgers[0]!] });
      const onSelectTicket = vi.fn();
      const user = userEvent.setup();
      render(<FleetOverview onSelectTicket={onSelectTicket} />);
      const region = await screen.findByRole('region', { name: 'Session overlap inspection' });
      expect(within(region).getByText('1 shared session found across tickets.')).toBeVisible();
      const disclosure = within(region).getByText('View shared sessions and affected tickets');
      expect(disclosure.closest('details')).not.toHaveAttribute('open');
      expect(within(region).getByText(sessionId)).not.toBeVisible();
      for (let i = 0; i < 4; i++) await user.tab();
      expect(screen.getByRole('button', { name: 'Ticket tags filter' })).toHaveFocus();
      await user.tab();
      expect(screen.getByRole('button', { name: 'Refresh Data' })).toHaveFocus();
      await user.tab();
      expect(disclosure).toHaveFocus();
      // user-event/jsdom do not implement native summary keyboard activation.
      await user.click(disclosure);
      expect(disclosure.closest('details')).toHaveAttribute('open');
      expect(within(region).getByText(sessionId)).toBeVisible();
      await user.tab();
      expect(within(region).getByRole('button', { name: 'View ticket HIDDEN-1' })).toHaveFocus();
      await user.keyboard('{Enter}');
      await user.tab();
      expect(within(region).getByRole('button', { name: 'View ticket HIDDEN-2' })).toHaveFocus();
      await user.keyboard(' ');
      expect(onSelectTicket.mock.calls).toEqual([['HIDDEN-1'], ['HIDDEN-2']]);
      expect(screen.getByText('$14.82')).toBeVisible();
      expect(screen.getByText(new RegExp(durationDisclosure))).toBeVisible();
      fireEvent.change(screen.getByLabelText('Date Range Preset'), { target: { value: '7d' } });
      await screen.findByLabelText(/filter tickets/i);
      await user.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
      await user.click(screen.getByRole('checkbox', { name: 'Bug' }));
      await waitFor(() => expect(apiClient.fetchWorkspaces).toHaveBeenLastCalledWith('7d', '', '', undefined, undefined, ['bug'], undefined));
      await user.click(screen.getByRole('combobox', { name: 'Prefix selector' }));
      await user.click(screen.getByRole('checkbox', { name: 'AL' }));
      await waitFor(() => expect(apiClient.fetchWorkspaces).toHaveBeenLastCalledWith('7d', '', '', undefined, undefined, ['bug'], ['AL']));
      await screen.findByLabelText(/filter tickets/i);
      fireEvent.change(screen.getByLabelText(/filter tickets/i), { target: { value: 'AL-2' } });
      expect(within(region).getByText(sessionId)).toBeVisible();
      expect(screen.queryByText('AL-1')).not.toBeInTheDocument();
      expect(inspect).toHaveBeenCalledTimes(1);
      expect(inspect).toHaveBeenCalledWith();
      fireEvent.click(screen.getByRole('button', { name: 'Refresh Data' }));
      await waitFor(() => expect(screen.queryByRole('region', { name: 'Session overlap inspection' })).not.toBeInTheDocument());
      expect(inspect).toHaveBeenCalledTimes(2);
      expect(screen.getByText('$14.82')).toBeVisible();
      expect(screen.getByText(new RegExp(durationDisclosure))).toBeVisible();
    });

    it('keeps partial conflicts and metrics visible, then reports inspection failure without claiming absence', async () => {
      vi.mocked(apiClient.fetchWorkspaceLedgers)
        .mockResolvedValueOnce({ data: ledgers, diagnostics: [{ scope: 'response', message: 'Unreadable diagnostics' }] })
        .mockRejectedValueOnce(new Error('Inspection offline'))
        .mockResolvedValue({ data: ledgers });
      render(<FleetOverview />);
      const region = await screen.findByRole('region', { name: 'Session overlap inspection' });
      expect(within(region).getByText(/Scope is partial/)).toBeVisible();
      fireEvent.click(within(region).getByText('View shared sessions and affected tickets'));
      expect(within(region).getByText(sessionId)).toBeVisible();
      expect(within(region).queryByRole('button')).not.toBeInTheDocument();
      fireEvent.click(screen.getByRole('button', { name: 'Refresh Data' }));
      expect(await screen.findByText('Session overlap inspection unavailable.')).toBeVisible();
      expect(screen.getByText(/Inspection offline/)).toHaveTextContent('Overlap outside inspected records is unknown');
      expect(screen.queryByText(sessionId)).not.toBeInTheDocument();
      expect(screen.getByText('Total Cost')).toBeVisible();
      expect(screen.getByText(new RegExp(durationDisclosure))).toBeVisible();
      fireEvent.click(screen.getByRole('button', { name: 'Refresh Data' }));
      expect(await screen.findByText('1 shared session found across tickets.')).toBeVisible();
      expect(screen.queryByText(/Inspection offline/)).not.toBeInTheDocument();
      expect(screen.queryByText(/Scope is partial/)).not.toBeInTheDocument();
      expect(screen.getByText(new RegExp(durationDisclosure))).toBeVisible();
    });

    it.each([
      { data: ledgers.slice(0, 1), diagnostics: [{ ticket_id: 'SKIPPED-1', message: 'Rejected ledger' }], reason: 'Rejected ledger' },
      { data: ledgers.slice(0, 1), diagnostics: [{ scope: 'response' as const, message: 'Diagnostics unavailable' }], reason: 'Diagnostics unavailable' },
      { data: [{ ...ledgers[0]!, facts: { ...ledgers[0]!.facts, sessions: [{ id: ' ', invocations: [], tools: [] }] } }], reason: '1 session records lacked usable IDs' },
    ])('reports limited scope without conflicts: $reason', async ({ reason, ...response }) => {
      vi.mocked(apiClient.fetchWorkspaceLedgers).mockResolvedValue(response);
      render(<FleetOverview />);
      const region = await screen.findByRole('region', { name: 'Session overlap inspection' });
      expect(within(region).getByText('Session overlap inspection is partial.')).toBeVisible();
      expect(region).toHaveTextContent(reason);
      expect(region).toHaveTextContent('overlap in records not inspected is unknown');
      expect(within(region).queryByText('View shared sessions and affected tickets')).not.toBeInTheDocument();
      expect(await screen.findByText('Total Cost')).toBeVisible();
    });

    it.each(['success', 'error'])('ignores an obsolete initial inspection %s after a newer corrected reload', async (outcome) => {
      let resolveInitial: ((value: apiClient.ApiResponse<readonly typeof base[]>) => void) | undefined;
      let rejectInitial: ((reason: Error) => void) | undefined;
      const initial = new Promise<apiClient.ApiResponse<readonly typeof base[]>>((resolve, reject) => { resolveInitial = resolve; rejectInitial = reject; });
      vi.mocked(apiClient.fetchWorkspaceLedgers).mockReturnValueOnce(initial).mockResolvedValue({ data: [ledgers[0]!] });
      render(<FleetOverview />);
      await screen.findByText('Total Cost');
      fireEvent.click(screen.getByRole('button', { name: 'Refresh Data' }));
      await waitFor(() => expect(screen.queryByText(/Inspecting loaded project/)).not.toBeInTheDocument());
      await act(async () => {
        if (outcome === 'success') {
          resolveInitial!({ data: ledgers });
          await initial;
        } else {
          rejectInitial!(new Error('Stale inspection failure'));
          await expect(initial).rejects.toThrow('Stale inspection failure');
        }
      });
      expect(screen.queryByText(sessionId)).not.toBeInTheDocument();
      expect(screen.queryByRole('region', { name: 'Session overlap inspection' })).not.toBeInTheDocument();
      expect(screen.getByText(new RegExp(durationDisclosure))).toBeVisible();
    });

  });

  describe('Adversarial & Boundary Scenarios', () => {
    it('displays error banner and allows retry when fetch throws unexpected error', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary')
        .mockRejectedValueOnce(new Error('Fatal API crash'))
        .mockResolvedValueOnce({
          data: mockFleetSummary,
            });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Fatal API crash')).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /retry connection/i });
      fireEvent.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });
      expect(fetchFleetSpy).toHaveBeenCalledTimes(2);
    });

    it('renders empty state when workspace list is empty but total_tickets is 0', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: {
          ...mockFleetSummary,
          total_tickets: 0,
          total_cost_usd: 0,
        },
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Zero Telemetry Workspaces Available')).toBeInTheDocument();
      });
      expect(screen.getByRole('button', { name: /refresh data/i })).toBeInTheDocument();
    });

    it('invokes onSelectTicket when a ticket row is clicked', async () => {
      const onSelectTicket = vi.fn();
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview onSelectTicket={onSelectTicket} />);

      await waitFor(() => {
        expect(screen.getByText('AL-1')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('AL-1'));
      expect(onSelectTicket).toHaveBeenCalledWith('AL-1');
    });

    it('renders FleetAnomalyCard in the fleet overview layout with anomaly metrics', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Fleet Anomaly and Reliability Profile')).toBeInTheDocument();
      });

      // Fleet anomaly metrics
      expect(screen.getByText('1 warning, 1 critical')).toBeInTheDocument();
      expect(screen.getByText('Tool Retry Streak')).toBeInTheDocument();
      expect(screen.getByText('Brittle Tools')).toBeInTheDocument();
    });

    it('triggers refetch with preset window when date range preset changes', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const presetSelect = screen.getByLabelText(/date range preset/i);
      fireEvent.change(presetSelect, { target: { value: '7d' } });

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('7d', undefined, undefined);
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('7d', '', '', undefined, undefined);
      });
    });

    it('displays custom date pickers and passes since/until when custom preset is selected', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const presetSelect = screen.getByLabelText(/date range preset/i);
      fireEvent.change(presetSelect, { target: { value: 'custom' } });

      const sinceInput = await screen.findByLabelText(/custom start date/i);
      const untilInput = await screen.findByLabelText(/custom end date/i);

      fireEvent.change(sinceInput, { target: { value: '2026-03-01' } });
      fireEvent.change(untilInput, { target: { value: '2026-03-15' } });

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('all', '2026-03-01', '2026-03-15');
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('all', '', '', '2026-03-01', '2026-03-15');
      });
    });

    it('triggers refetch when ticket type and prefix filters change', async () => {
      const fetchFleetSpy = vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      const fetchWorkspacesSpy = vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [...mockTicketComparisonRows],
        });

      render(<FleetOverview />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      // Open ticket type dropdown and select Bug
      const typeTrigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
      fireEvent.click(typeTrigger);

      const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
      fireEvent.click(bugCheckbox);

      await waitFor(() => {
        expect(fetchFleetSpy).toHaveBeenCalledWith('all', undefined, undefined, ['bug'], undefined);
        expect(fetchWorkspacesSpy).toHaveBeenCalledWith('all', '', '', undefined, undefined, ['bug'], undefined);
      });
    });

    it('renders configured prefixes in prefix dropdown even when no matching workspaces exist', async () => {
      vi.spyOn(apiClient, 'fetchFleetSummary').mockResolvedValue({
        data: mockFleetSummary,
        });
      vi.spyOn(apiClient, 'fetchWorkspaces').mockResolvedValue({
        data: [],
        });

      render(<FleetOverview configuredPrefixes={['NEWINITIATIVE', 'INFRA']} />);

      await waitFor(() => {
        expect(screen.getByText('Total Cost')).toBeInTheDocument();
      });

      const prefixTrigger = screen.getByRole('combobox', { name: 'Prefix selector' });
      fireEvent.click(prefixTrigger);

      expect(screen.getByRole('checkbox', { name: 'NEWINITIATIVE' })).toBeInTheDocument();
      expect(screen.getByRole('checkbox', { name: 'INFRA' })).toBeInTheDocument();
    });
  });
});
