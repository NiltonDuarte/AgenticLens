import * as fs from 'node:fs';
import * as path from 'node:path';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, within, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TicketDeepDive } from './TicketDeepDive';
import * as apiClient from '../api/client';
import { type Anomaly, type EventFact } from '../types/telemetry';
import { type AgenticMetricsLedger, type ContextHealthStatus, type WorkspaceDetail, type SessionFact, type SkillFact, type DelegationFact } from '../domain/types';
import { mockAgenticMetricsLedgers, mockWorkspaceDetails, mockAL2Events } from '../test/fixtures/mockData';
import codexOwnedDetail from '../test/fixtures/codexOwnedDetail.json';
import codexSourceEvents from '../test/fixtures/codexSourceEvents.json';

function withSessionOwners(detail: WorkspaceDetail, events: readonly EventFact[]): WorkspaceDetail {
    const owners = [...new Set(events.map((event) => event.session_id))];
    const sessions: readonly SessionFact[] = owners.map((id) => ({
        id, source: 'opencode', user: 'tester', invocations: [], tools: [],

    }));
    return { ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts, schema_version: 6 as const, sessions } } };
}

function deferred<T>() {
    let resolve!: (value: T) => void;
    let reject!: (reason: Error) => void;
    const promise = new Promise<T>((resolvePromise, rejectPromise) => {
        resolve = resolvePromise;
        reject = rejectPromise;
    });
    return { promise, resolve, reject };
}

describe('Skills and subagents autopsy', () => {
    const timestamp = '2026-10-30T12:34:56.007Z';
    const skill: SkillFact = {
        name: 'inspect-source', caller_role: 'orchestrator', invoked_at: timestamp,
    };
    const delegation: DelegationFact = {
        agent_type: 'coder', description: 'Implement the agreed change', invoked_at: timestamp,
    };
    const makeDetail = (skills?: readonly SkillFact[], delegations?: readonly DelegationFact[]): WorkspaceDetail => {
        const base = mockWorkspaceDetails['AL-1']!;
        const facts = { ...base.ledger.facts, schema_version: 6 as const };
        return { ...base, ledger: { ...base.ledger, facts: { ...facts,
            sessions: facts.sessions.map((session, index) => index === 0 ? { ...session,
                ...(skills === undefined ? {} : { skills }), ...(delegations === undefined ? {} : { delegations: delegations }),
            } : session),
        } } };
    };

    beforeEach(() => {
        vi.restoreAllMocks();
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: Object.values(mockAgenticMetricsLedgers) });
        vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockRejectedValue(new Error('Optional source unavailable'));
    });


    it('shows unknown role metrics and turn groups when optional hydration is unavailable', async () => {
        const detail = makeDetail();
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: {
            ...detail,
            ledger: {
                ...detail.ledger,
                roles: [{ ...detail.ledger.roles[0]!, role: '' }],
                workflow_efficiency: {
                    ...detail.ledger.workflow_efficiency,
                    subagent_turn_velocity: { ...detail.ledger.workflow_efficiency.subagent_turn_velocity, by_role: { '': 7 } },
                },
            },
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await screen.findByText('7.0 turns/inv');
        expect(screen.getAllByText('Unknown').length).toBeGreaterThanOrEqual(2);
    });

    it('keeps persisted provisional duration visible when source hydration is unavailable', async () => {
        const detail = makeDetail();
        const cutoff = '2026-11-01T12:00:00Z';
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: { ...detail, ledger: { ...detail.ledger,
            workflow_efficiency: { ...detail.ledger.workflow_efficiency, active_duration_seconds: 10, active_duration_formatted: '10s',
                duration_summary: { provisional_seconds: 7, provisional_invocations: 1, unknown_invocations: 1, cutoffs: [cutoff] },
            },
        } } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const card = (await screen.findByText('Active Compute Duration')).closest<HTMLElement>('.kpi-card')!;
        expect(card.querySelector('.kpi-value')).toHaveTextContent('10s (coverage incomplete, provisional)');
        expect(card.querySelector('.kpi-secondary')).toHaveTextContent('7s provisional');
        expect(card.querySelector('.kpi-secondary')).toHaveTextContent(cutoff);
        expect(card.querySelector('.kpi-secondary')).toHaveTextContent('unknown or unsupported');
    });

    it('uses compact cost presentation for the selected ticket with full evidence in the tooltip', async () => {
        const detail = makeDetail();
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: { ...detail, ledger: { ...detail.ledger, totals: {
            ...detail.ledger.totals, invocations: 3, cost_usd: null,
            cost_summary: { state: 'partial', known_subtotal_usd: 2, fully_observed: 1, partial: 1, unavailable: 1, total: 3 },
        } } } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const card = (await screen.findByText('Total Cost')).closest<HTMLElement>('.kpi-card')!;
        expect(card.querySelector('.kpi-value')).toHaveTextContent(/^\$2\.00$/);
        expect(card.querySelector('.kpi-secondary')).toHaveTextContent('Known subtotal (partial) · 3 invocations');
        expect(within(card).getByRole('tooltip')).toHaveTextContent('1 fully observed, 1 partial, 1 unavailable of 3 invocations');
    });

    it.each(['omitted', 'empty'] as const)('renders a current v6 ledger with %s optional collections and keyboard collapse', async (mode) => {
        const errors = vi.spyOn(console, 'error');
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mode === 'omitted' ? makeDetail() : (() => { const detail = makeDetail([], []); return { ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts, sessions: detail.ledger.facts.sessions.map((owner) => ({ ...owner, skills: [], delegations: [] })) } } }; })(),
        });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const panel = await screen.findByRole('region', { name: /Skills & Subagents/ });
        const toggle = within(panel).getByRole('button', { name: /Skills & Subagents/ });
        const empty = mode === 'empty' ? within(panel).getByText('No modular skills or subagent delegations recorded for this ticket.') : document.getElementById(toggle.getAttribute('aria-controls')!)!.querySelector('p')!;
        if (mode === 'omitted') expect(panel).toHaveTextContent(/not recorded/i);
        expect(empty).toBeVisible();
        expect(within(panel).queryByRole('table')).not.toBeInTheDocument();
        expect(toggle).toHaveAttribute('aria-expanded', 'true');
        expect(document.getElementById(toggle.getAttribute('aria-controls')!)).toContainElement(empty);
        toggle.focus();
        await user.keyboard('{Enter}');
        expect(toggle).toHaveAttribute('aria-expanded', 'false');
        expect(empty).not.toBeVisible();
        await user.keyboard(' ');
        expect(toggle).toHaveAttribute('aria-expanded', 'true');
        expect(empty).toBeVisible();
        expect(errors).not.toHaveBeenCalled();
    });

    it('explains unrecorded categories without guessing why source hydration is unavailable', async () => {
        const detail = makeDetail();
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts } },
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const panel = await screen.findByRole('region', { name: /Skills & Subagents/ });
        expect(panel).toHaveTextContent(/not recorded/i);
        expect(panel).not.toHaveTextContent(/disabled/i);
        expect(within(panel).queryByText('No modular skills or subagent delegations recorded for this ticket.')).not.toBeInTheDocument();
        expect(within(panel).queryByText(/0 Skills/)).not.toBeInTheDocument();
        expect(screen.queryByText(/were disabled|was disabled/i)).not.toBeInTheDocument();
    });

    it('explains unavailable historical analysis while retaining supported skills after source hydration fails', async () => {
        const detail = makeDetail([skill], [delegation]);
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts,
                sessions: detail.ledger.facts.sessions.map((owner, index) => {
                    if (index !== 0) return owner;
                    const { anomalies, ...recorded } = owner;
                    void anomalies;
                    return recorded;
                }),
            } },
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        expect(await screen.findByText(/anomaly analysis.*not recorded|anomalies.*not recorded/i)).toBeVisible();
        expect(await screen.findByRole('table', { name: 'Modular skills' })).toHaveTextContent('inspect-source');
        expect(screen.queryByText('Recorded anomaly analysis was disabled when this ledger was generated. Source trace analysis is independent.')).not.toBeInTheDocument();
    });

    it('keeps missing, recorded-empty and known skill coverage distinct when filtering owners', async () => {
        const detail = makeDetail();
        const sessions: readonly SessionFact[] = [
            { id: 'known', source: 'codex', invocations: [], tools: [], skills: [skill], delegations: [] },
            { id: 'empty', source: 'codex', invocations: [], tools: [], skills: [], delegations: [] },
            { id: 'missing', source: 'opencode', invocations: [], tools: [] },
        ];
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts, sessions },
                threads: sessions.map((owner) => ({ ...detail.ledger.threads[0]!, session_id: owner.id })),
            },
        } });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const panel = await screen.findByRole('region', { name: /Skills & Subagents/ });
        expect(panel).toHaveTextContent('1 recorded Skills');
        expect(within(panel).getByRole('table', { name: 'Modular skills' })).toHaveTextContent('inspect-source');
        const filter = within(panel).getByRole('combobox', { name: 'Filter skills and delegations by session' });
        await user.selectOptions(filter, 'empty');
        expect(panel).toHaveTextContent('0 Skills');
        expect(within(panel).queryByRole('table', { name: 'Modular skills' })).not.toBeInTheDocument();
        await user.selectOptions(filter, 'missing');
        expect(panel).toHaveTextContent(/skills not recorded/i);
        expect(panel).not.toHaveTextContent('0 Skills');
        await user.selectOptions(filter, 'known');
        expect(panel).toHaveTextContent('1 Skills');
        expect(within(panel).getByRole('table', { name: 'Modular skills' })).toHaveTextContent('inspect-source');
    });

    it('shows lean native owners and nullable activations, filters exactly by session and preserves unavailable owned facts after failed enrichment', async () => {
        const detail = makeDetail();
        const base = { ...skill, caller_role: null, invoked_at: null };
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: { ...detail, ledger: { ...detail.ledger, facts: {
            ...detail.ledger.facts, schema_version: 6,
            sessions: [
                { id: 'root', source: 'codex', invocations: [], tools: [], skills: [base, base], delegations: [{ ...delegation, child_session_id: 'unavailable-known-child' }] },
                { id: 'child', parent_id: 'root', source: 'opencode', invocations: [], tools: [], skills: [{ ...skill, name: 'child-only' }] },
                { id: 'remote-owner', source: 'codex', invocations: [], tools: [], skills: [{ ...skill, name: 'historical' }] },
            ],
        } } } });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const table = await screen.findByRole('table', { name: 'Modular skills' });
        expect(within(table).getAllByRole('row').slice(1)).toHaveLength(4);
        expect(within(table).getAllByText('Unavailable')).toHaveLength(4);
        expect(within(table).queryByText('Failed')).not.toBeInTheDocument();
        expect(within(table).getByText('remote-owner')).toBeVisible();
        expect(screen.getByText('Thread unavailable')).toBeVisible();
        expect(screen.getByText('unavailable-known-child')).toBeVisible();
        expect(screen.queryByRole('button', { name: /Jump to child session/ })).not.toBeInTheDocument();
        const select = screen.getByRole('combobox', { name: 'Filter skills and delegations by session' });
        await user.selectOptions(select, 'root');
        expect(within(table).getAllByRole('row').slice(1)).toHaveLength(2);
        expect(screen.queryByText('historical')).not.toBeInTheDocument();
        await user.selectOptions(select, 'child');
        expect(within(table).getAllByRole('row').slice(1)).toHaveLength(1);
        expect(within(table).getByText('child-only')).toBeVisible();
        expect(screen.queryByRole('table', { name: 'Subagent delegations' })).not.toBeInTheDocument();
        await user.selectOptions(select, '');
        expect(within(table).getByText('historical')).toBeVisible();
        expect(within(table).getAllByRole('row').slice(1)).toHaveLength(4);
    });

    it('preserves repeated skill invocations, nullable roles, timestamps and keyboard access', async () => {
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: makeDetail([
            skill,
            { ...skill, caller_role: 'custom-investigator' },
            { ...skill, name: 'review-change', caller_role: null, invoked_at: null },
        ]) });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const table = await screen.findByRole('table', { name: 'Modular skills' });
        const rows = within(table).getAllByRole('row').slice(1);
        expect(rows).toHaveLength(3);
        screen.getByRole('region', { name: 'Recorded skills' }).focus();
        await user.tab();
        expect(rows[0]).toHaveFocus();
        await user.tab();
        expect(rows[1]).toHaveFocus();
        expect(within(table).getAllByRole('rowheader', { name: 'inspect-source' })).toHaveLength(2);
        const success = within(rows[0]!);
        expect(within(table).queryByText(/Success|Failed|Outcome/)).not.toBeInTheDocument();
        expect(success.getByText('Orchestrator')).toBeVisible();
        const time = rows[0]!.querySelector('time');
        expect(time).toHaveAttribute('datetime', timestamp);
        expect(time).toHaveAttribute('title', expect.stringMatching(/2026.*\.007 \[UTC[+-]\d{2}:\d{2}\]/));
        expect(time).toHaveTextContent(/2026/);
        expect(time).toHaveTextContent(/\d{2}:\d{2}:\d{2}/);
        expect(time?.textContent).not.toMatch(/:\d{2}\.\d/);
        expect(within(rows[1]!).getByText('custom-investigator')).toBeVisible();
        expect(within(rows[2]!).getAllByText('Unavailable')).toHaveLength(2);
        expect(screen.queryByText('No modular skills or subagent delegations recorded for this ticket.')).not.toBeInTheDocument();
    });

    it('shows delegations without skills, preserves unavailable identities, and jumps repeatedly to the exact child by keyboard', async () => {
        const longDescription = 'Source supplied short task description';
        const detail = makeDetail(undefined, [
            { ...delegation, description: longDescription, child_session_id: 'child-a' },
            { ...delegation, agent_type: null, invoked_at: null },
            { ...delegation, description: 'Missing child thread', child_session_id: 'child-a-unavailable' },
            { ...delegation, description: 'Missing child thread', child_session_id: 'child-a-unavailable' },
        ]);
        const root = detail.ledger.threads[0]!;
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, threads: [
                { ...root, session_id: 'root' }, { ...root, session_id: 'child-ab' }, { ...root, session_id: 'child-a' },
            ] },
        } });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const table = await screen.findByRole('table', { name: 'Subagent delegations' });
        expect(screen.queryByRole('table', { name: 'Modular skills' })).not.toBeInTheDocument();
        const rows = within(table).getAllByRole('row').slice(1);
        expect(rows).toHaveLength(4);
        expect(rows[0]!.children[0]).toHaveTextContent('Coder');
        expect(rows[0]!.children[1]?.textContent).toBe(longDescription);
        expect(rows[0]!.querySelector('time')).toHaveAttribute('datetime', timestamp);
        expect(within(rows[1]!).getAllByText('Unavailable')).toHaveLength(2);
        expect(within(rows[1]!).getByText('Session identity unavailable')).toBeVisible();
        expect(within(rows[2]!).getByText('child-a-unavailable')).toBeVisible();
        expect(within(rows[2]!).getByText('Thread unavailable')).toBeVisible();
        expect(within(rows[2]!).queryByRole('button')).not.toBeInTheDocument();
        expect(within(table).queryByRole('columnheader', { name: /Duration|Cost|Parent Turn/ })).not.toBeInTheDocument();
        const child = screen.getByRole('button', { name: /^child-a User:/ });
        const sibling = screen.getByRole('button', { name: /^child-ab User:/ });
        const parent = screen.getByRole('button', { name: /^root User:/ });
        expect(parent).toHaveAttribute('aria-expanded', 'true');
        expect(child).toHaveAttribute('aria-expanded', 'false');
        expect(sibling).toHaveAttribute('aria-expanded', 'false');
        const jump = within(table).getByRole('button', { name: 'Jump to child session child-a' });
        jump.focus();
        await user.keyboard('{Enter}');
        expect(child).toHaveFocus();
        expect(child).toHaveAttribute('aria-expanded', 'true');
        expect(parent).toHaveAttribute('aria-expanded', 'true');
        expect(sibling).toHaveAttribute('aria-expanded', 'false');
        expect(child.parentElement?.querySelector('.thread-body')).toBeVisible();
        await user.keyboard(' ');
        expect(child).toHaveAttribute('aria-expanded', 'false');
        jump.focus();
        await user.keyboard(' ');
        expect(child).toHaveFocus();
        expect(child).toHaveAttribute('aria-expanded', 'true');
    });

    it('keeps persisted skill and delegation facts visible before and after optional enrichment fails', async () => {
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: makeDetail([skill], [delegation]) });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const skills = await screen.findByRole('table', { name: 'Modular skills' });
        const tasks = screen.getByRole('table', { name: 'Subagent delegations' });
        expect(within(skills).getByText('inspect-source')).toBeVisible();
        expect(within(tasks).getByText('Implement the agreed change')).toBeVisible();
        await act(async () => pending.reject(new Error('Optional enrichment failed')));
        expect(skills).toBeVisible();
        expect(tasks).toBeVisible();
        expect(screen.queryByText('Optional enrichment failed')).not.toBeInTheDocument();
    });
});

describe('Source trace and stored facts lifecycle', () => {
    const recorded: readonly EventFact[] = [
        { event_id: 'source-error', timestamp: '2026-10-29T10:00:00Z', session_id: 'owner',
            message_id: 'source-message', turn: 1, role: 'coder', event_type: 'tool_call', tool_name: 'bash', status: 'error', context_tokens: 1200 },
        { event_id: 'source-message', timestamp: '2026-10-29T10:01:00Z', session_id: 'owner', turn: 2,
            role: 'tester', event_type: 'turn_message', status: 'completed', context_tokens: 1300 },
    ];
    const coverage = { required_session_ids: ['owner'], complete_session_ids: ['owner'], diagnostics: [] };
    const makeDetail = (ticketId = 'AL-1'): WorkspaceDetail => {
        const detail = withSessionOwners(mockWorkspaceDetails['AL-1']!, recorded);
        return { ...detail, anomalies: [], ledger: { ...detail.ledger, ticket_id: ticketId,
            facts: { ...detail.ledger.facts, ticket_id: ticketId, sessions: [{ id: 'owner', invocations: [], tools: [],
                skills: [{ name: 'stored-activation', caller_role: null, invoked_at: null }],
                delegations: [{ agent_type: null, invoked_at: null, child_session_id: 'remote-child' }] }] } } };
    };
    const emptyResponse = (ticketId = 'AL-1') => ({ data: {
        ticket_id: ticketId, is_hydrated: false, availability: 'unavailable' as const, coverage: { ...coverage, complete_session_ids: [], diagnostics: ['Source unavailable'] },
        data_source: 'unavailable' as const, events: [],
    } });
    beforeEach(() => {
        vi.restoreAllMocks();
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: Object.values(mockAgenticMetricsLedgers) });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => ({ data: makeDetail(id) }));
        vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockRejectedValue(new Error('Optional source unavailable'));
    });

    it.each(['failed', 'unavailable', 'partial', 'complete'] as const)('keeps stored facts before and after %s source hydration without artificial rows', async (state) => {
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        const anomaly: Anomaly = { rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { tool_name: 'bash', count: 3 } };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: { ...makeDetail(), anomalies: [anomaly] } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const skills = await screen.findByRole('table', { name: 'Modular skills' });
        expect(within(skills).getByText('stored-activation')).toBeVisible();
        expect(screen.getByText('remote-child')).toBeVisible();
        expect(screen.getByText('Total Cost')).toBeVisible();
        expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded');
        fireEvent.click(screen.getByRole('button', { name: 'Recorded anomaly summaries (1)' }));
        fireEvent.click(screen.getByRole('button', { name: /Inspect recorded anomaly: Tool retry streak/ }));
        const recordedInspector = screen.getByRole('dialog');
        expect(recordedInspector).toHaveTextContent('3 consecutive errors');
        await act(async () => {
            if (state === 'failed') pending.reject(new Error('Cannot open source'));
            else pending.resolve({ data: { ...emptyResponse().data, availability: state, is_hydrated: state !== 'unavailable',
                data_source: state === 'unavailable' ? 'unavailable' : 'local_codex',
                coverage: state === 'complete' ? coverage : { ...coverage, complete_session_ids: [], diagnostics: ['Missing owner source'] },
                anomalies: [], events: state === 'partial' ? [recorded[0]!] : [] } });
        });
        expect(skills).toBeVisible();
        expect(screen.getByText('remote-child')).toBeVisible();
        expect(screen.getByRole('dialog')).toBe(recordedInspector);
        expect(recordedInspector).toHaveTextContent('3 consecutive errors');
        expect(screen.getByRole('button', { name: /Inspect recorded anomaly: Tool retry streak/ })).toBeVisible();
        expect(screen.queryByRole('button', { name: /Jump to child session remote-child/ })).not.toBeInTheDocument();
        expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent(`${state === 'partial' ? 1 : 0} Events Recorded`);
        if (state === 'complete') {
            expect(screen.getByText('Zero chronological events recorded for this execution trace.')).toBeVisible();
            expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
        } else {
            expect(screen.getByTestId('db-offline-banner')).toHaveTextContent(state === 'failed' ? 'Source trace request failed' : state === 'partial' ? 'Partial source trace' : 'Source trace unavailable');
            expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('1 Anomaly');
        }
        expect(within(skills).queryByText(/Success rate|Execution Status/i)).not.toBeInTheDocument();
    });

    it('ignores ticket A source response after ticket B and resets the inspected summary', async () => {
        const pendingA = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const anomaly: Anomaly = { rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { tool_name: 'bash', count: 3 } };
        vi.mocked(apiClient.fetchWorkspaceById).mockImplementation(async (id) => ({ data: { ...makeDetail(id), anomalies: id === 'AL-1' ? [anomaly] : [] } }));
        const bEvent = { ...recorded[0]!, session_id: 'b-session', turn: 8, tool_output: 'B details' };
        vi.mocked(apiClient.fetchWorkspaceEvents).mockImplementation((id) => id === 'AL-1' ? pendingA.promise : Promise.resolve({ data: {
            ...emptyResponse(id).data, availability: 'complete', is_hydrated: true, coverage, data_source: 'local_codex', events: [bEvent] } }));
        const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);
        fireEvent.click(await screen.findByRole('button', { name: 'Expand Trace Waterfall' }));
        fireEvent.click(screen.getByRole('button', { name: /Inspect anomaly: Tool retry streak/ }));
        expect(screen.getByRole('dialog')).toHaveTextContent('TOOL_RETRY_STREAK');
        rerender(<TicketDeepDive selectedTicketId="AL-2" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Local Codex'));
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        fireEvent.click(screen.getByTestId('timeline-item-8'));
        await act(async () => pendingA.resolve({ data: { ...emptyResponse().data, availability: 'complete', is_hydrated: true, coverage,
            data_source: 'local_sqlite', events: [{ ...recorded[0]!, tool_output: 'A stale details' }] } }));
        expect(screen.getByRole('dialog')).toHaveTextContent('b-session');
        expect(screen.getByTestId('timeline-item-8')).toHaveClass('selected');
        expect(screen.queryByTestId('timeline-item-1')).not.toBeInTheDocument();
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Local Codex');
    });

    it.each(['resolve', 'reject'])('ignores a stale source %s after returning to a ticket with a newer source request', async (settlement) => {
        const old = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValueOnce(old.promise).mockImplementation(async (id) => ({ data: {
            ...emptyResponse(id).data, availability: 'complete', is_hydrated: true, coverage,
            data_source: 'local_codex', events: [{ ...recorded[0]!, tool_output: 'Newest source payload' }] } }));
        const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);
        await screen.findByRole('table', { name: 'Modular skills' });
        await waitFor(() => expect(apiClient.fetchWorkspaceEvents).toHaveBeenCalledTimes(1));
        rerender(<TicketDeepDive selectedTicketId="AL-2" />);
        await waitFor(() => expect(apiClient.fetchWorkspaceEvents).toHaveBeenCalledTimes(2));
        rerender(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(apiClient.fetchWorkspaceEvents).toHaveBeenCalledTimes(3));
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Local Codex'));
        fireEvent.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        expect(screen.getByRole('dialog')).toHaveTextContent('owner');
        await act(async () => {
            if (settlement === 'reject') old.reject(new Error('Old source failed'));
            else old.resolve({ data: emptyResponse().data });
        });
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Local Codex');
        expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
        expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
        expect(screen.getByRole('dialog')).toHaveTextContent('Newest source payload');
    });

    it('renders full workspace dates visibly and preserves missing lifecycle states', async () => {
        const detail = makeDetail();
        const started = '2026-10-29T12:34:56.007Z';
        const completed = '2026-10-30T12:34:56.123Z';
        const updated = '2026-10-31T12:34:56.456Z';
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, started_at: started, completed_at: completed, updated_at: updated },
        } });
        const { container, rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);
        await screen.findByText('Started:');
        const times = container.querySelectorAll('.timestamp-group time');
        expect(times).toHaveLength(3);
        [started, completed, updated].forEach((instant, index) => {
            expect(times[index]).toHaveAttribute('datetime', instant);
            expect(times[index]).toHaveTextContent('2026');
            expect(times[index]).toHaveTextContent(/\d{2}:\d{2}:\d{2}/);
            expect(times[index]?.textContent).not.toMatch(/:\d{2}\.\d/);
            expect(times[index]).toHaveAttribute('title', expect.stringContaining('[UTC'));
        });
        expect(new Set(Array.from(times, (time) => time.textContent)).size).toBe(3);
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, started_at: '', completed_at: '', updated_at: '' },
        } });
        rerender(<TicketDeepDive selectedTicketId="AL-2" />);
        await waitFor(() => expect(container.querySelector('.timestamp-group')).toHaveTextContent('Completed: In Progress'));
        expect(container.querySelector('.timestamp-group')).toHaveTextContent('Started: N/A');
        expect(container.querySelector('.timestamp-group')).toHaveTextContent('Updated: N/A');
    });

    it('preserves malformed workspace timestamps as the visible time and tooltip fallback', async () => {
        const malformedTimestamp = 'not-a-valid-timestamp';
        const detail = makeDetail();
        const malformedDetail: WorkspaceDetail = {
            ...detail,
            ledger: { ...detail.ledger, started_at: malformedTimestamp, updated_at: malformedTimestamp },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: malformedDetail });

        render(<TicketDeepDive selectedTicketId="AL-1" />);

        await waitFor(() => expect(screen.getAllByText(malformedTimestamp)).toHaveLength(2));
        const fallbackTimes = screen.getAllByText(malformedTimestamp);
        fallbackTimes.forEach((timestamp) => expect(timestamp).toHaveAttribute('title', malformedTimestamp));
    });

    it('rebinds an inspected summary to updated anomaly context while enrichment remains event-free', async () => {
        const user = userEvent.setup();
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const anomaly: Anomaly = {
            rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
            context: { session_id: 'summary-session', tool_name: 'bash', count: 3 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1'), anomalies: [anomaly],
        } });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await user.click(await screen.findByRole('button', { name: 'Expand Trace Waterfall' }));
        const card = screen.getByTestId('trace-waterfall-card');
        expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
        await user.click(within(card).getByRole('button', { name: /Inspect anomaly: Tool retry streak/ }));
        const inspector = screen.getByRole('dialog');
        expect(inspector).toHaveTextContent('3 consecutive errors');
        await act(async () => pending.resolve({ data: {
            ...emptyResponse().data, availability: 'complete', is_hydrated: true, coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] }, data_source: 'local_codex',
            anomalies: [{ ...anomaly, severity: 'CRITICAL', context: { ...anomaly.context, count: 6 } }],
        } }));
        expect(screen.getByRole('dialog')).toBe(inspector);
        expect(inspector).toHaveTextContent('6 consecutive errors');
        expect(within(inspector).getByText('CRITICAL')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Collapse Trace Waterfall' })).toHaveAttribute('aria-expanded', 'true');
        expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
        expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('Zero Errors, 1 Anomaly');
        expect(within(card).getByRole('button', { name: /Inspect anomaly: Tool retry streak/ })).toHaveTextContent('CRITICAL');
        expect(within(inspector).queryByText('Timestamp')).not.toBeInTheDocument();
        expect(within(inspector).queryByTestId('drawer-tool-output-section')).not.toBeInTheDocument();
    });

    it.each([false, true])('preserves expanded=%s when an event-free summary receives event enrichment', async (expanded) => {
        const user = userEvent.setup();
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const anomaly: Anomaly = {
            rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
            context: { session_id: recorded[0]!.session_id, tool_name: 'bash', count: 3 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1'), anomalies: [anomaly],
        } });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded'));
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        if (expanded) {
            await user.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
            await user.click(screen.getByRole('button', { name: /Inspect anomaly: Tool retry streak/ }));
        }
        const inspector = screen.queryByRole('dialog');
        const toggle = screen.getByRole('button', { name: `${expanded ? 'Collapse' : 'Expand'} Trace Waterfall` });
        const bodyId = toggle.getAttribute('aria-controls');
        await act(async () => pending.resolve({ data: {
            ticket_id: 'AL-1', is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] }, data_source: 'local_codex',
            events: [{ ...recorded[0]!, tool_output: 'Enriched failure diagnostics' }],
            anomalies: [{ ...anomaly, severity: 'CRITICAL', context: { ...anomaly.context, count: 6 } }],
        } }));

        expect(screen.getByRole('button', { name: `${expanded ? 'Collapse' : 'Expand'} Trace Waterfall` })).toBe(toggle);
        expect(toggle).toHaveAttribute('aria-expanded', String(expanded));
        expect(toggle).toHaveAttribute('aria-controls', bodyId);
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex');
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();
        if (expanded) {
            expect(inspector).not.toBeNull();
            expect(screen.getByRole('dialog')).toBe(inspector);
            expect(inspector).toHaveTextContent('6 consecutive errors');
            expect(within(inspector!).getByText('CRITICAL')).toBeInTheDocument();
            expect(within(inspector!).queryByText('Timestamp')).not.toBeInTheDocument();
            expect(within(inspector!).queryByTestId('drawer-tool-output-section')).not.toBeInTheDocument();
        } else {
            expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('1 Events Recorded');
            expect(screen.queryByTestId('timeline-item-1')).not.toBeInTheDocument();
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            await user.click(toggle);
        }
        expect(within(screen.getByTestId('trace-waterfall-card')).getByTestId('timeline-item-1')).toBeVisible();
        expect(screen.getByText('Showing 1 of 1 events')).toBeInTheDocument();
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
    });

});
describe('TicketDeepDive View Component', () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });
    it('renders loading state initially and then displays full autopsy KPIs', async () => {
        const ledgers = Object.values(mockAgenticMetricsLedgers);
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: ledgers,
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mockWorkspaceDetails['AL-1']!,
        });
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
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
                    root_peak_tokens: 165000,
                    root_saturation_pct: 82.5,
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
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: {
                ledger: elevatedLedger,
                artifacts: { ticket_id: 'AL-1', contract_content: null, summary_content: null },
                anomalies: [],
            },
        });
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
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
                    root_peak_tokens: 240000,
                    root_saturation_pct: 120.0,
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
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: {
                ledger: criticalLedger,
                artifacts: { ticket_id: 'AL-CRIT', contract_content: null, summary_content: null },
                anomalies: [],
            },
        });
        render(<TicketDeepDive selectedTicketId="AL-CRIT"/>);
        await waitFor(() => {
            expect(screen.getByText('CRITICAL')).toBeInTheDocument();
        });
        expect(screen.getByText('100% Ceiling')).toBeInTheDocument();
        expect(screen.getByText('(120.0% of ceiling)')).toBeInTheDocument();
    });
    it('renders graceful empty state when workspace detail has missing or null ledger', async () => {
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: Object.values(mockAgenticMetricsLedgers),
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: {
                ledger: null as unknown as typeof mockAgenticMetricsLedgers['AL-1'],
                artifacts: { ticket_id: 'MISSING', contract_content: null, summary_content: null },
                anomalies: [],
            },
        });
        render(<TicketDeepDive selectedTicketId="MISSING"/>);
        await waitFor(() => {
            expect(screen.getByText(/zero telemetry workspaces available/i)).toBeInTheDocument();
        });
    });
    it('updates active ticket when selectedTicketId prop changes from parent', async () => {
        const ledgers = Object.values(mockAgenticMetricsLedgers);
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: ledgers,
        });
        const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
            const detail = mockWorkspaceDetails[id] ?? {
                ledger: mockAgenticMetricsLedgers[id]!,
                artifacts: { ticket_id: id, contract_content: null, summary_content: null },
            };
            return { data: detail };
        });
        const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1"/>);
        await waitFor(() => {
            expect(screen.getByText('AL-1')).toBeInTheDocument();
        });
        rerender(<TicketDeepDive selectedTicketId="AL-2"/>);
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
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: {
                ledger: noTurnsLedger,
                artifacts: { ticket_id: 'AL-1', contract_content: null, summary_content: null },
                anomalies: [],
            },
        });
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
        await waitFor(() => {
            expect(screen.getByText(/no subagent turn records available/i)).toBeInTheDocument();
        });
    });
    it('switches artifact tab between CONTRACT.md and SUMMARY.md', async () => {
        const ledgers = Object.values(mockAgenticMetricsLedgers);
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: ledgers,
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mockWorkspaceDetails['AL-1']!,
        });
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
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
        });
        const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
            const detail = mockWorkspaceDetails[id] ?? {
                ledger: mockAgenticMetricsLedgers[id]!,
                artifacts: { ticket_id: id, contract_content: null, summary_content: null },
            };
            return { data: detail };
        });
        const onSelectTicket = vi.fn();
        render(<TicketDeepDive selectedTicketId="AL-1" onSelectTicket={onSelectTicket}/>);
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
        });
        const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => {
            const detail = mockWorkspaceDetails[id] ?? {
                ledger: mockAgenticMetricsLedgers[id]!,
                artifacts: { ticket_id: id, contract_content: null, summary_content: null },
            };
            return { data: detail };
        });
        const onSelectTicket = vi.fn();
        render(<TicketDeepDive selectedTicketId="AL-1" onSelectTicket={onSelectTicket}/>);
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
        });
        render(<TicketDeepDive />);
        await waitFor(() => {
            expect(screen.getByText(/zero telemetry workspaces available/i)).toBeInTheDocument();
        });
    });
    it('defaults to first available ticket when selectedTicketId is not provided', async () => {
        const ledgers = Object.values(mockAgenticMetricsLedgers);
        const fetchLedgersSpy = vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: ledgers,
        });
        const fetchByIdSpy = vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mockWorkspaceDetails['AL-1']!,
        });
        render(<TicketDeepDive />);
        await waitFor(() => {
            expect(fetchByIdSpy).toHaveBeenCalledWith(ledgers[0]!.ticket_id);
        });
        expect(fetchLedgersSpy).toHaveBeenCalled();
    });
    it('displays error banner and allows retry when API call fails', async () => {
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockRejectedValueOnce(new Error('Failed to load workspaces'));
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
        await waitFor(() => {
            expect(screen.getByText('Failed to load workspaces')).toBeInTheDocument();
        });
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValueOnce({
            data: Object.values(mockAgenticMetricsLedgers),
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValueOnce({
            data: mockWorkspaceDetails['AL-1']!,
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
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mockWorkspaceDetails['AL-1']!,
        });
        render(<TicketDeepDive selectedTicketId="AL-1"/>);
        await waitFor(() => {
            expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
        });
        // Refined KPI tooltips
        expect(screen.getByText('Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel.')).toBeInTheDocument();
        expect(screen.getByText('Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution.')).toBeInTheDocument();
        // Turn velocity header tooltip
        expect(screen.getByRole('button', { name: 'Information about Subagent Turn Velocity' })).toBeInTheDocument();
        expect(screen.getByText('Average and peak conversational turns executed per agent invocation, categorized by subagent persona role.')).toBeInTheDocument();
        // Role table header tooltip
        const roleHeaderBtn = screen.getByRole('button', {
            name: 'Information about Role Resource Consumption',
        });
        expect(roleHeaderBtn).toBeInTheDocument();
        expect(screen.getByText('Telemetry breakdown of model invocations, token volume, cache hit rate, and spend by agent persona role.')).toBeInTheDocument();
        const roleTooltipId = roleHeaderBtn.getAttribute('aria-describedby');
        expect(roleTooltipId).toBeTruthy();
        const roleBubble = document.getElementById(roleTooltipId!);
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
        // Model table header tooltip
        const modelHeaderBtn = screen.getByRole('button', {
            name: 'Information about Model Resource Consumption',
        });
        expect(modelHeaderBtn).toBeInTheDocument();
        expect(screen.getByText('Telemetry breakdown of invocations, token volume, cache hit rate, and spend by model family and version.')).toBeInTheDocument();
        const modelTooltipId = modelHeaderBtn.getAttribute('aria-describedby');
        expect(modelTooltipId).toBeTruthy();
        const modelBubble = document.getElementById(modelTooltipId!);
        expect(modelBubble).toHaveClass('placement-left');
        expect(modelBubble).not.toHaveClass('is-visible');
        fireEvent.focus(modelHeaderBtn);
        expect(modelBubble).toHaveClass('is-visible');
        fireEvent.blur(modelHeaderBtn);
        expect(modelBubble).not.toHaveClass('is-visible');
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
        expect(screen.getByText('Agent persona role (orchestrator, coder, tester, reviewer, planner)')).toBeInTheDocument();
        expect(screen.getByText('Total count of model invocations initiated by this role')).toBeInTheDocument();
        expect(screen.getByText('Total input and output tokens consumed by this role')).toBeInTheDocument();
        expect(screen.getByText('Percentage of prompt tokens served from prompt cache for this role')).toBeInTheDocument();
        expect(screen.getByText('Total model inference spend in US dollars for this role')).toBeInTheDocument();
        // Exact explanatory contents for Model Resource Consumption columns
        expect(screen.getByText('Model identifier and version invoked')).toBeInTheDocument();
        expect(screen.getByText('Total count of invocations executed on this model')).toBeInTheDocument();
        expect(screen.getByText('Total input and output tokens consumed by this model')).toBeInTheDocument();
        expect(screen.getByText('Percentage of prompt tokens served from prompt cache for this model')).toBeInTheDocument();
        expect(screen.getByText('Total model inference spend in US dollars for this model')).toBeInTheDocument();
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
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: emptyWorkspace,
        });
        render(<TicketDeepDive selectedTicketId="AL-EMPTY"/>);
        await waitFor(() => {
            expect(screen.getByText('Empty Workspace Telemetry')).toBeInTheDocument();
        });
        // Check KPI tooltips intact
        expect(screen.getByText('Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel.')).toBeInTheDocument();
        expect(screen.getByText('Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution.')).toBeInTheDocument();
        // Turn Velocity tooltip intact
        expect(screen.getByRole('button', { name: 'Information about Subagent Turn Velocity' })).toBeInTheDocument();
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
        expect(screen.getByText('No session execution threads recorded for this workspace.')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Information about Session Execution Threads' })).toBeInTheDocument();
    });
    it('wraps Role and Model Resource Consumption tables in .table-responsive containers and keeps headers outside', async () => {
        const ledgers = Object.values(mockAgenticMetricsLedgers);
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: ledgers,
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mockWorkspaceDetails['AL-1']!,
        });
        const { container } = render(<TicketDeepDive selectedTicketId="AL-1"/>);
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
    it('renders custom agent roles and arbitrary models in autopsy tables and velocity', async () => {
        const customRoleLedger: AgenticMetricsLedger = {
            ...mockAgenticMetricsLedgers['AL-1']!,
            ticket_id: 'AL-CUSTOM',
            workflow_efficiency: {
                ...mockAgenticMetricsLedgers['AL-1']!.workflow_efficiency,
                subagent_turn_velocity: {
                    avg_turns_per_invocation: 5.5,
                    peak_turns: 8,
                    by_role: {
                        advisor: 4.0,
                        po: 7.0,
                    },
                },
                context_window_health: {
                    healthy_ceiling_tokens: 200000,
                    status: 'HEALTHY' as const,
                    root_peak_tokens: 60000,
                    root_saturation_pct: 30.0,
                    subagents_peak_tokens: 150000,
                    subagents_peak_role: 'advisor',
                    compactions: 0,
                    peak_by_role: {
                        orchestrator: 60000,
                        advisor: 150000,
                        po: 80000,
                    },
                },
            },
            roles: [
                {
                    role: 'orchestrator',
                    invocations: 1,
                    input_tokens: 10000,
                    output_tokens: 2000,
                    cache_read_tokens: 5000,
                    cache_hit_rate_pct: 33.3,
                    cost_usd: 1.0,
                    peak_context_tokens: 60000,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
                {
                    role: 'advisor',
                    invocations: 2,
                    input_tokens: 20000,
                    output_tokens: 4000,
                    cache_read_tokens: 10000,
                    cache_hit_rate_pct: 33.3,
                    cost_usd: 2.0,
                    peak_context_tokens: 150000,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
                {
                    role: 'po',
                    invocations: 1,
                    input_tokens: 8000,
                    output_tokens: 1000,
                    cache_read_tokens: 4000,
                    cache_hit_rate_pct: 33.3,
                    cost_usd: 0.8,
                    peak_context_tokens: 80000,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
            ],
            models: [
                {
                    model: 'claude-3-7-sonnet',
                    invocations: 2,
                    input_tokens: 20000,
                    output_tokens: 4000,
                    cache_read_tokens: 10000,
                    cache_hit_rate_pct: 33.3,
                    cost_usd: 2.0,
                    peak_context_tokens: 150000,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
                {
                    model: 'gemini-2.5-pro',
                    invocations: 2,
                    input_tokens: 18000,
                    output_tokens: 3000,
                    cache_read_tokens: 9000,
                    cache_hit_rate_pct: 33.3,
                    cost_usd: 1.8,
                    peak_context_tokens: 80000,
                    context_health: 'HEALTHY' as ContextHealthStatus,
                },
            ],
        };
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
            data: [customRoleLedger],
        });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: {
                ledger: customRoleLedger,
                artifacts: { ticket_id: 'AL-CUSTOM', contract_content: null, summary_content: null },
                anomalies: [],
            },
        });
        render(<TicketDeepDive selectedTicketId="AL-CUSTOM"/>);
        await waitFor(() => {
            expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
        });
        // Custom roles rendered in turn velocity and role breakdown
        expect(screen.getAllByText('advisor').length).toBeGreaterThanOrEqual(2);
        expect(screen.getAllByText('po').length).toBeGreaterThanOrEqual(2);
        // Arbitrary models rendered in model breakdown table
        expect(screen.getByText('claude-3-7-sonnet')).toBeInTheDocument();
        expect(screen.getByText('gemini-2.5-pro')).toBeInTheDocument();
    });
    describe('Source waterfall and AnomalyDrawer integration', () => {
        it('renders an immediately visible current empty trace without anomalies or a legacy claim', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: withSessionOwners(mockWorkspaceDetails['AL-1']!, []),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-1',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    data_source: 'local_sqlite',
                    events: []
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-1"/>);
            await waitFor(() => {
                expect(screen.getByText('Zero chronological events recorded for this execution trace.')).toBeInTheDocument();
            });
            expect(screen.queryByTestId('legacy-fallback-banner')).not.toBeInTheDocument();
            expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded');
            expect(screen.queryByTestId('waterfall-expand-btn')).not.toBeInTheDocument();
            expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
            expect(screen.queryByRole('group', { name: 'Filter events by type' })).not.toBeInTheDocument();
            expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
            expect(screen.getByText('Session Execution Threads')).toBeInTheDocument();
            expect(screen.getByText(/Contract: AL-1/i)).toBeInTheDocument();
        });
        it.each(['click', 'Enter', 'Space'] as const)('opens the view-owned aggregate inspector by %s inside an expanded empty trace', async (activation) => {
            const user = userEvent.setup();
            const summaryWithAnomalies = {
                ...withSessionOwners(mockWorkspaceDetails['AL-1']!, []),
                anomalies: [
                    {
                        rule_id: 'CONTEXT_SATURATION_SPIKE' as const,
                        severity: 'CRITICAL' as const,
                        context: { turn: 5, token_delta: 50000, context_tokens: 130000 },
                    },
                ],
            };
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: summaryWithAnomalies,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-1',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    data_source: 'local_sqlite',
                    events: []
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-1"/>);
            await user.click(await screen.findByRole('button', { name: 'Expand Trace Waterfall' }));
            const card = screen.getByTestId('trace-waterfall-card');
            const list = within(card).getByTestId('workspace-anomalies');
            expect(card).toContainElement(list);
            expect(within(list).getByText('Detected Workspace Anomalies:')).toBeVisible();
            expect(screen.getByText('Context saturation spike (+50k tokens)')).toBeInTheDocument();
            expect(screen.getByText('CRITICAL')).toBeInTheDocument();
            expect(screen.queryByTestId('legacy-fallback-banner')).not.toBeInTheDocument();
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            const anomaly = within(list).getByRole('button', { name: /Inspect anomaly: Context saturation spike/ });
            if (activation === 'click') {
                await user.click(anomaly);
            } else {
                anomaly.focus();
                await user.keyboard(activation === 'Enter' ? '{Enter}' : ' ');
            }
            const inspector = screen.getByRole('dialog');
            expect(card).not.toContainElement(inspector);
            expect(inspector).toHaveTextContent('CONTEXT_SATURATION_SPIKE');
            expect(inspector).toHaveTextContent('Turn 5');
            expect(within(inspector).getByText('CRITICAL')).toBeInTheDocument();
            expect(within(inspector).queryByText('Timestamp')).not.toBeInTheDocument();
            expect(within(inspector).queryByText('Execution Duration')).not.toBeInTheDocument();
            expect(within(inspector).queryByTestId('drawer-tool-output-section')).not.toBeInTheDocument();
            expect(screen.getByRole('button', { name: 'Collapse Trace Waterfall' })).toHaveAttribute('aria-expanded', 'true');

            await user.click(within(inspector).getByRole('button', { name: 'Close' }));
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            expect(list).toBeVisible();
            await user.click(anomaly);
            expect(screen.getByRole('dialog')).toHaveTextContent('CONTEXT_SATURATION_SPIKE');
            await user.keyboard('{Escape}');
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            expect(screen.getByRole('button', { name: 'Collapse Trace Waterfall' })).toHaveAttribute('aria-expanded', 'true');
            await user.click(screen.getByRole('button', { name: 'Collapse Trace Waterfall' }));
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        });
        it('renders TraceWaterfall component when viewing a ticket with populated events', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: mockWorkspaceDetails['AL-2']!,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: mockAL2Events,
                    data_source: 'local_sqlite',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            expect(screen.queryByTestId('legacy-fallback-banner')).not.toBeInTheDocument();
        });
        it('clicking an anomalous event in the waterfall opens AnomalyDrawer and closing restores state', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: mockWorkspaceDetails['AL-2']!,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: mockAL2Events,
                    data_source: 'local_sqlite',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            // Anomaly drawer is initially not present
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            // Click the warning anomaly badge in the waterfall
            const warningBadges = screen.getAllByText('WARNING');
            expect(warningBadges.length).toBeGreaterThan(0);
            fireEvent.click(warningBadges[0]!);
            // Anomaly drawer is now open
            await waitFor(() => {
                expect(screen.getByRole('dialog')).toBeInTheDocument();
            });
            expect(screen.getByText(/Tool retry streak: bash/i)).toBeInTheDocument();
            // Close the drawer
            const closeBtn = screen.getByRole('button', { name: 'Close' });
            fireEvent.click(closeBtn);
            await waitFor(() => {
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });
        });
        it('clicking an unflagged timeline node highlights it without opening the AnomalyDrawer', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: mockWorkspaceDetails['AL-2']!,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: mockAL2Events,
                    data_source: 'local_sqlite',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            // Switch to All Events to access Turn 1
            const allBtn = screen.getByRole('button', { name: 'All Events' });
            fireEvent.click(allBtn);
            const turn1Items = screen.getAllByTestId('timeline-item-1');
            expect(turn1Items.length).toBeGreaterThan(0);
            const unflaggedNode = turn1Items[0]!;
            // Click unflagged node
            fireEvent.click(unflaggedNode);
            // Verify node is highlighted
            expect(unflaggedNode).toHaveClass('selected');
            // Verify drawer did NOT open
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });
        it('clicking an unflagged node closes the drawer if previously opened by a flagged node', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: mockWorkspaceDetails['AL-2']!,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: mockAL2Events,
                    data_source: 'local_sqlite',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            // Switch to All Events
            const allBtn = screen.getByRole('button', { name: 'All Events' });
            fireEvent.click(allBtn);
            // Click flagged item (turn 4 error item - find the one with error)
            const turn4Items = screen.getAllByTestId('timeline-item-4');
            const flaggedNode = turn4Items.find((item) => item.classList.contains('item-error')) ?? turn4Items[1]!;
            fireEvent.click(flaggedNode);
            // Drawer opens
            await waitFor(() => {
                expect(screen.getByRole('dialog')).toBeInTheDocument();
            });
            // Now click unflagged turn 1
            const turn1Items = screen.getAllByTestId('timeline-item-1');
            const unflaggedNode = turn1Items[0]!;
            fireEvent.click(unflaggedNode);
            // Drawer should now be closed and unflagged node selected
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            expect(unflaggedNode).toHaveClass('selected');
        });
        it('inspecting a plain tool error without anomaly rule opens drawer with Tool Execution Failure title and no Rule Diagnosis', async () => {
            const plainErrorEvent: EventFact = {
                timestamp: '2026-10-18T10:10:00Z',
                session_id: 's-plain-err',
                turn: 1,
                role: 'coder',
                event_type: 'tool_call',
                tool_name: 'bash',
                status: 'error',
                error_message: 'exit status 2: command not found',
                duration_ms: 100,
                context_tokens: 20000,
            };
            const plainDetail = withSessionOwners({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, [plainErrorEvent]);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: plainDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    data_source: 'local_sqlite',
                    events: [plainErrorEvent]
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            // Click error item (visible under default Errors and Anomalies Only)
            const errItem = screen.getByTestId('timeline-item-1');
            fireEvent.click(errItem);
            // Drawer opens
            await waitFor(() => {
                expect(screen.getByRole('dialog')).toBeInTheDocument();
            });
            // Drawer title is Tool Execution Failure
            expect(screen.getByText('Tool Execution Failure')).toBeInTheDocument();
            // Rule Diagnosis section is NOT present
            expect(screen.queryByText('Rule Diagnosis')).not.toBeInTheDocument();
            // Full error message is present in the drawer
            expect(within(screen.getByRole('dialog')).getByText('exit status 2: command not found')).toBeInTheDocument();
        });
        it('clicking Inspect Failure button on tool error turn opens drawer with Tool Execution Failure title', async () => {
            const plainErrorEvent: EventFact = {
                timestamp: '2026-10-18T10:10:00Z',
                session_id: 's-plain-err',
                turn: 1,
                role: 'coder',
                event_type: 'tool_call',
                tool_name: 'bash',
                status: 'error',
                error_message: 'exit status 2: command not found',
                duration_ms: 100,
                context_tokens: 20000,
            };
            const plainDetail = withSessionOwners({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, [plainErrorEvent]);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: plainDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    data_source: 'local_sqlite',
                    events: [plainErrorEvent]
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            const inspectBtn = screen.getByRole('button', { name: 'Inspect failure for turn 1' });
            fireEvent.click(inspectBtn);
            await waitFor(() => {
                expect(screen.getByRole('dialog')).toBeInTheDocument();
            });
            expect(screen.getByText('Tool Execution Failure')).toBeInTheDocument();
            expect(within(screen.getByRole('dialog')).getByText('exit status 2: command not found')).toBeInTheDocument();
        });
        it('renders hydrated waterfall without offline banner when is_hydrated is true', async () => {
            const richEvents: EventFact[] = [
                {
                    timestamp: '2026-10-18T10:00:00Z',
                    session_id: 's-rich-1',
                    turn: 1,
                    role: 'coder',
                    event_type: 'tool_call',
                    tool_name: 'bash',
                    status: 'error',
                    error_message: 'exit status 1',
                    duration_ms: 100,
                    context_tokens: 5000,
                    prompt_content: 'Check build status',
                    reasoning_content: 'Running build command',
                    tool_input: 'make build',
                    tool_output: 'error: syntax error',
                },
            ];
            const richDetail = withSessionOwners({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, richEvents);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: richDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true, availability: 'complete', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: richEvents,
                    data_source: 'local_sqlite',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Verify Local SQLite source badge is present and offline banner is NOT rendered
            const sourceBadge = screen.getByTestId('waterfall-source-badge');
            await waitFor(() => expect(sourceBadge).toHaveTextContent('Source: Local SQLite'));
            expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
        });
        it('renders offline banner when is_hydrated is false', async () => {
            const offlineEvents: EventFact[] = [
                {
                    timestamp: '2026-10-18T10:00:00Z',
                    session_id: 's-off-1',
                    turn: 1,
                    role: 'coder',
                    event_type: 'tool_call',
                    tool_name: 'bash',
                    status: 'error',
                    error_message: 'failed',
                    context_tokens: 5000,
                },
            ];
            const offlineDetail = withSessionOwners({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, offlineEvents);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: offlineDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: false, availability: 'partial', coverage: { required_session_ids: [], complete_session_ids: [], diagnostics: [] },
                    events: offlineEvents,
                    data_source: 'unavailable',
                },
            });
            render(<TicketDeepDive selectedTicketId="AL-2"/>);
            await waitFor(() => {
                expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
            });
            // Expand the waterfall if collapsed
            const expandBtn = screen.queryByTestId('waterfall-expand-btn');
            if (expandBtn) {
                fireEvent.click(expandBtn);
            }
            expect(screen.getByTestId('db-offline-banner')).toBeInTheDocument();
            const sourceBadge = screen.getByTestId('waterfall-source-badge');
            expect(sourceBadge).toHaveTextContent('Source unavailable');
        });
        it('renders unclassified pill and does not render feature when ticket_type is null or undefined', async () => {
            const unclassifiedDetail = {
                ...mockWorkspaceDetails['AL-1']!,
                ledger: {
                    ...mockWorkspaceDetails['AL-1']!.ledger,
                    ticket_type: null,
                },
            };
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: unclassifiedDetail,
            });
            render(<TicketDeepDive selectedTicketId="AL-1"/>);
            await waitFor(() => {
                expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
            });
            const pill = screen.getByText('unclassified');
            expect(pill).toBeInTheDocument();
            expect(pill).toHaveClass('ticket-type-pill');
            expect(pill).toHaveClass('unclassified');
            expect(screen.queryByText('feature')).not.toBeInTheDocument();
        });
        it('renders unclassified badge when ticket_type is undefined and never renders feature', async () => {
            const ledgerWithoutType = { ...mockWorkspaceDetails['AL-1']!.ledger };
            delete (ledgerWithoutType as {
                ticket_type?: unknown;
            }).ticket_type;
            const unclassifiedDetail = {
                ...mockWorkspaceDetails['AL-1']!,
                ledger: ledgerWithoutType as typeof mockWorkspaceDetails['AL-1']['ledger'],
            };
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: unclassifiedDetail,
            });
            render(<TicketDeepDive selectedTicketId="AL-1"/>);
            await waitFor(() => {
                expect(screen.getByText('SDLC Iterations')).toBeInTheDocument();
            });
            const pill = screen.getByText('unclassified');
            expect(pill).toBeInTheDocument();
            expect(pill).toHaveClass('ticket-type-pill');
            expect(pill).toHaveClass('unclassified');
            expect(screen.queryByText('feature')).not.toBeInTheDocument();
        });
    });
});


describe('Native producer wire through API validation and ticket presentation', () => {
    it('renders lean activations and separate source badges without saved traces or opaque child links', async () => {
        vi.restoreAllMocks();
        const fetch = vi.spyOn(globalThis, 'fetch').mockImplementation(async (input) => {
            const route = String(input);
            const payload: unknown = route === '/api/workspaces' ? [codexOwnedDetail.ledger]
                : route === '/api/workspaces/TEST-WIRE' ? codexOwnedDetail
                : route === '/api/workspaces/TEST-WIRE/events' ? codexSourceEvents
                : undefined;
            if (payload === undefined) throw new Error(`Unexpected fixture route: ${route}`);
            return new Response(JSON.stringify(payload), { status: 200, headers: { 'Content-Type': 'application/json' } });
        });
        try {
            render(<TicketDeepDive selectedTicketId="TEST-WIRE" />);
            const skills = await screen.findByRole('table', { name: 'Modular skills' });
            expect(within(skills).getAllByRole('rowheader', { name: 'inspect-source' })).toHaveLength(2);
            expect(within(skills).queryByRole('columnheader', { name: /Status|Outcome|Error/ })).not.toBeInTheDocument();
            const delegations = screen.getByRole('table', { name: 'Subagent delegations' });
            expect(within(delegations).getAllByRole('row').slice(1)).toHaveLength(3);
            expect(within(delegations).getByText('unavailable-a')).toBeVisible();
            expect(within(delegations).getByText('unavailable-b')).toBeVisible();
            expect(within(delegations).queryByText('private prompt')).not.toBeInTheDocument();
            expect(within(delegations).queryByRole('button', { name: /Jump to child session/ })).not.toBeInTheDocument();
            await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex'));
            fireEvent.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
            fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
            expect(screen.getByText('Showing 6 of 6 events')).toBeVisible();
            expect(screen.getAllByText('[skill: inspect-source]')).toHaveLength(2);
            expect(screen.getAllByText('[task -> tester]')).toHaveLength(2);
            expect(screen.getByText('[task -> reviewer]')).toBeVisible();
            const activation = screen.getByText('Skill Activation').closest<HTMLElement>('.timeline-item')!;
            expect(within(activation).getByText('Activation')).toBeVisible();
            expect(within(activation).queryByText('Success')).not.toBeInTheDocument();
            expect(screen.getByRole('button', { name: 'Copy child session ID unavailable-a' })).toBeVisible();
            expect(screen.getByRole('button', { name: 'Copy child session ID unavailable-b' })).toBeVisible();
            expect(screen.queryByRole('button', { name: /Inspect child session/ })).not.toBeInTheDocument();
            expect(screen.queryByText('[skill: catalogue-only]')).not.toBeInTheDocument();
            expect(screen.queryByText('[skill: quoted]')).not.toBeInTheDocument();
        } finally {
            fetch.mockRestore();
        }
    });
});
