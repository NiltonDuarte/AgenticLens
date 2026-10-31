import * as fs from 'node:fs';
import * as path from 'node:path';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent, within, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TicketDeepDive } from './TicketDeepDive';
import * as apiClient from '../api/client';
import { type Anomaly, type EventFact } from '../types/telemetry';
import * as clipboard from '../utils/clipboard';
import { type AgenticMetricsLedger, type ContextHealthStatus, type WorkspaceDetail, type SessionFact, type SkillFact, type TaskLinkFact } from '../domain/types';
import { mockAgenticMetricsLedgers, mockWorkspaceDetails, mockAL2Events } from '../test/fixtures/mockData';

function withRecordedEvents(detail: WorkspaceDetail, events: readonly EventFact[]): WorkspaceDetail {
    const owners = [...new Set(events.map((event) => event.session_id))];
    const sessions: readonly SessionFact[] = owners.map((id) => ({
        id, source: 'opencode', user: 'tester', invocations: [], tools: [],
        events: events.filter((event) => event.session_id === id).map((event) => ({
            event_id: event.event_id, timestamp: event.timestamp, message_id: event.message_id,
            turn: event.turn, role: event.role, event_type: event.event_type, tool_name: event.tool_name,
            status: event.status, duration_ms: event.duration_ms, context_tokens: event.context_tokens,
        })),
    }));
    return { ...detail, ledger: { ...detail.ledger, facts: { ...detail.ledger.facts, schema_version: 4, sessions } } };
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
        name: 'inspect-source', turn_index: 7, message_index: 2,
        caller_role: 'orchestrator', timestamp, success: true,
    };
    const delegation: TaskLinkFact = {
        subagent_type: 'coder', description: 'Implement the agreed change', parent_turn: 8, timestamp,
    };
    const makeDetail = (skills?: readonly SkillFact[], taskLinks?: readonly TaskLinkFact[]): WorkspaceDetail => {
        const base = mockWorkspaceDetails['AL-1']!;
        const facts = { ...base.ledger.facts, schema_version: 4 };
        delete facts.skills;
        delete facts.task_links;
        return { ...base, ledger: { ...base.ledger, facts: {
            ...facts,
            ...(skills === undefined ? {} : { skills }),
            ...(taskLinks === undefined ? {} : { task_links: taskLinks }),
        } } };
    };

    beforeEach(() => {
        vi.restoreAllMocks();
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: Object.values(mockAgenticMetricsLedgers) });
        vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockRejectedValue(new Error('Optional source unavailable'));
    });

    it.each(['omitted', 'empty'] as const)('renders a supported v4 ledger with %s optional collections and keyboard collapse', async (mode) => {
        const errors = vi.spyOn(console, 'error');
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
            data: mode === 'omitted' ? makeDetail() : makeDetail([], []),
        });
        const user = userEvent.setup();
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        const panel = await screen.findByRole('region', { name: /Skills & Subagents/ });
        const toggle = within(panel).getByRole('button', { name: /Skills & Subagents/ });
        const empty = within(panel).getByText('No modular skills or subagent delegations recorded for this ticket.');
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

    it('preserves repeated skill invocations, recorded status, roles, timestamps and complete diagnostics', async () => {
        const error = 'Source error: validation failed\n' + 'Detailed diagnostic context '.repeat(12);
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({ data: makeDetail([
            { ...skill, error: 'Inspection of error text completed successfully' },
            { ...skill, success: false, caller_role: 'custom-investigator', turn_index: 9, error },
            { ...skill, name: 'review-change', caller_role: '' },
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
        expect(success.getByText('Success')).toBeVisible();
        expect(success.getByText('Orchestrator')).toBeVisible();
        expect(success.getByText('Inspection of error text completed successfully')).toBeVisible();
        expect(rows[0]).toHaveTextContent('Turn 7 /');
        const time = rows[0]!.querySelector('time');
        expect(time).toHaveAttribute('datetime', timestamp);
        expect(time).toHaveAttribute('title', expect.stringMatching(/2026.*\.007 \[UTC[+-]\d{2}:\d{2}\]/));
        expect(time).toHaveTextContent(/\d{2}:\d{2}:\d{2}\.007/);
        expect(within(rows[1]!).getByText('Failed')).toBeVisible();
        expect(within(rows[1]!).getByText('custom-investigator')).toBeVisible();
        expect(rows[1]).toHaveTextContent('Turn 9 /');
        expect(rows[1]!.lastElementChild?.textContent).toBe(error);
        expect(within(rows[2]!).getByText('Unknown')).toBeVisible();
        expect(within(rows[2]!).getByText('None recorded')).toBeVisible();
        expect(screen.queryByText('No modular skills or subagent delegations recorded for this ticket.')).not.toBeInTheDocument();
    });

    it('shows delegations without skills, distinguishes missing metrics, and jumps repeatedly to the exact child by keyboard', async () => {
        const longDescription = 'Preserve the whole task prompt: ' + 'task details '.repeat(30);
        const detail = makeDetail(undefined, [
            { ...delegation, description: longDescription, child_session_id: 'child-a', duration_ms: 62500, cost_usd: 1.25 },
            { ...delegation, description: 'Measured zero', duration_ms: 0, cost_usd: 0 },
            { ...delegation, description: 'Missing optional measurements', child_session_id: 'child-a-unavailable' },
            { ...delegation, description: 'Missing optional measurements', child_session_id: 'child-a-unavailable' },
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
        expect(rows[0]!.children[2]).toHaveTextContent('8');
        expect(rows[0]!.children[3]).toHaveTextContent('1m 2s');
        expect(rows[0]!.children[4]).toHaveTextContent('$1.25');
        expect(rows[1]!.children[3]).toHaveTextContent('0m 0s');
        expect(rows[1]!.children[4]).toHaveTextContent('$0.00');
        expect(within(rows[1]!).getByText('Not recorded')).toBeVisible();
        expect(rows[2]!.children[3]).toHaveTextContent('Unavailable');
        expect(rows[2]!.children[4]).toHaveTextContent('Unavailable');
        expect(within(rows[2]!).getByText('child-a-unavailable')).toBeVisible();
        expect(within(rows[2]!).getByText('Thread unavailable')).toBeVisible();
        expect(within(rows[2]!).queryByRole('button')).not.toBeInTheDocument();
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

describe('FIX-12 committed trace and enrichment lifecycle', () => {
    const recorded: readonly EventFact[] = [
        { event_id: 'recorded-error', timestamp: '2026-10-29T10:00:00Z', session_id: 'recorded-session',
            message_id: 'recorded-message', turn: 1, role: 'coder', event_type: 'tool_call',
            tool_name: 'bash', status: 'error', context_tokens: 1200 },
        { event_id: 'recorded-success', timestamp: '2026-10-29T10:01:00Z', session_id: 'recorded-session',
            turn: 2, role: 'tester', event_type: 'turn_message', status: 'completed', context_tokens: 1300 },
    ];
    const makeDetail = (ticketId = 'AL-1', events = recorded): WorkspaceDetail => {
        const detail = withRecordedEvents({ ...mockWorkspaceDetails['AL-1']!, anomalies: [] }, events);
        return { ...detail, ledger: { ...detail.ledger, ticket_id: ticketId,
            facts: { ...detail.ledger.facts, ticket_id: ticketId } } };
    };
    const emptyResponse = (ticketId = 'AL-1') => ({ data: {
        ticket_id: ticketId, is_hydrated: false, data_source: 'committed_ledger' as const, events: [],
    } });

    beforeEach(() => {
        vi.restoreAllMocks();
        vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({ data: Object.values(mockAgenticMetricsLedgers) });
        vi.spyOn(apiClient, 'fetchWorkspaceById').mockImplementation(async (id) => ({ data: makeDetail(id) }));
        vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockRejectedValue(new Error('Optional source unavailable'));
    });

    it.each(['rejected', 'empty offline', 'empty hydrated'])('keeps baseline facts after %s enrichment', async (mode) => {
        if (mode !== 'rejected') {
            vi.mocked(apiClient.fetchWorkspaceEvents).mockResolvedValue({ data: {
                ...emptyResponse().data, is_hydrated: mode === 'empty hydrated', data_source: 'local_sqlite',
            } });
        }
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        const headerTimes = document.querySelectorAll('time[title]');
        expect(headerTimes.length).toBeGreaterThan(0);
        expect(headerTimes[0]).toHaveTextContent(/\d{2}:\d{2}:\d{2}\.\d{3}/);
        expect(headerTimes[0]).toHaveAttribute('title', expect.stringMatching(/\d{4}/));
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
        expect(screen.queryByTestId('legacy-fallback-banner')).not.toBeInTheDocument();
        expect(screen.queryByText(/Failed to load ticket details|Optional source unavailable/)).not.toBeInTheDocument();
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
        expect(screen.queryByTestId('timeline-item-2')).not.toBeInTheDocument();
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        expect(screen.getByRole('dialog')).toHaveTextContent('Tool Execution Failure');
        fireEvent.click(screen.getByRole('button', { name: 'Close' }));
        fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
        fireEvent.click(screen.getByTestId('timeline-item-2'));
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
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

    it.each(['resolve', 'reject'])('renders the baseline before a deferred enrichment request can %s', async (settlement) => {
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        expect(screen.queryByText(/loading workspace autopsy telemetry/i)).not.toBeInTheDocument();
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
        await act(async () => {
            if (settlement === 'resolve') pending.resolve(emptyResponse());
            else pending.reject(new Error('Late optional failure'));
        });
        expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
    });

    it.each(['stable ID', 'no IDs', 'new local ID'] as const)('rebinds the selected %s row and open inspector when pending enrichment completes', async (identity) => {
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const copy = vi.spyOn(clipboard, 'copyTextToClipboard').mockResolvedValue(true);
        const baseline: EventFact = {
            ...recorded[0]!, duration_ms: 0,
            event_id: identity === 'stable ID' ? 'recorded-error' : undefined,
            message_id: identity === 'stable ID' ? 'recorded-message' : undefined,
        };
        const anomaly: Anomaly = {
            rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
            context: { session_id: baseline.session_id, tool_name: 'bash', turn: 1, count: 3 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1', [baseline, recorded[1]!]), anomalies: [anomaly],
        } });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        const inspector = screen.getByRole('dialog');
        expect(inspector).toHaveTextContent('3 consecutive errors');
        expect(inspector).toHaveTextContent('0ms');
        expect(within(inspector).queryByTestId('drawer-tool-output-section')).not.toBeInTheDocument();
        expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
        expect(screen.getByRole('button', { name: 'Hide payload for turn 1' })).toBeInTheDocument();

        const enriched: EventFact = {
            ...baseline,
            event_id: identity === 'new local ID' ? 'local-error' : baseline.event_id,
            message_id: 'local-message', duration_ms: 25, error_message: 'local command exited with status 2',
            prompt_content: 'Run the focused regression checks', reasoning_content: 'Inspect the failing command',
            tool_input: 'make test', tool_output: 'local failure diagnostics',
        };
        await act(async () => pending.resolve({ data: {
            ticket_id: 'AL-1', is_hydrated: true, data_source: 'local_sqlite', events: [enriched],
            anomalies: [{ ...anomaly, severity: 'CRITICAL', context: { ...anomaly.context, count: 6 } }],
        } }));

        expect(screen.getByRole('dialog')).toBe(inspector);
        expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
        expect(screen.getByText('Showing 1 of 2 events')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Hide payload for turn 1' })).toBeInTheDocument();
        expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('local failure diagnostics');
        expect(inspector).toHaveTextContent('local-message');
        expect(inspector).not.toHaveTextContent('recorded-message');
        expect(inspector).toHaveTextContent('25ms');
        expect(inspector).toHaveTextContent('local command exited with status 2');
        expect(inspector).toHaveTextContent('6 consecutive errors');
        expect(within(inspector).getByText('CRITICAL')).toBeInTheDocument();
        expect(within(inspector).getByTestId('drawer-prompt-section')).toHaveTextContent(enriched.prompt_content!);
        expect(within(inspector).getByTestId('drawer-reasoning-section')).toHaveTextContent(enriched.reasoning_content!);
        expect(within(inspector).getByTestId('drawer-tool-input-section')).toHaveTextContent('make test');
        expect(within(inspector).getByTestId('drawer-tool-output-section')).toHaveTextContent('local failure diagnostics');
        await act(async () => fireEvent.click(within(inspector).getByRole('button', { name: 'Copy execution issue debug prompt' })));
        expect(copy).toHaveBeenCalledWith(expect.stringContaining('local command exited with status 2'));
    });

    it('keeps an ID-less selection and its open inspector across a known unhydrated baseline echo', async () => {
        const baseline = { ...recorded[0]!, event_id: undefined, message_id: undefined };
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: makeDetail('AL-1', [baseline]) });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockReturnValue(pending.promise);
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('1 Events Recorded'));
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        const inspector = screen.getByRole('dialog');
        await act(async () => pending.resolve({ data: {
            ...emptyResponse().data, includes_baseline: true, events: [{ ...baseline }],
        } }));
        expect(screen.getByRole('dialog')).toBe(inspector);
        expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
        expect(screen.getByRole('button', { name: 'Hide payload for turn 1' })).toBeInTheDocument();
        expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent('Payload details unavailable');
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
        expect(screen.getByText('Showing 1 of 1 events')).toBeInTheDocument();
    });

    it('rebinds an inspected summary to updated anomaly context while enrichment remains event-free', async () => {
        const user = userEvent.setup();
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const anomaly: Anomaly = {
            rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
            context: { session_id: 'summary-session', tool_name: 'bash', count: 3 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1', []), anomalies: [anomaly],
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
            ...emptyResponse().data,
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
            ...makeDetail('AL-1', []), anomalies: [anomaly],
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
            ticket_id: 'AL-1', is_hydrated: true, data_source: 'local_codex',
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

    it.each([false, true])('never creates a summary list for an event-bearing trace with zero filter matches (hydrated=%s)', async (isHydrated) => {
        const user = userEvent.setup();
        const anomaly: Anomaly = {
            rule_id: 'CONTEXT_SATURATION_SPIKE', severity: 'WARNING', context: { turn: 999, token_delta: 50000 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1', [recorded[1]!]), anomalies: [anomaly],
        } });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockResolvedValue({ data: {
            ...emptyResponse().data, is_hydrated: isHydrated,
            data_source: isHydrated ? 'local_codex' : 'committed_ledger', events: [recorded[1]!],
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge'))
            .toHaveTextContent(isHydrated ? 'Source: Local Codex' : 'Source: Committed Ledger'));
        expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('1 Events Recorded');
        await user.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        expect(screen.getByText('No events match the selected filters.')).toBeVisible();
        expect(screen.getByText('Showing 0 of 1 events')).toBeInTheDocument();
        expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /^Inspect anomaly:/ })).not.toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'All Events' }));
        expect(screen.getByTestId('timeline-item-2')).toBeVisible();
        await user.click(screen.getByRole('button', { name: 'Collapse Trace Waterfall' }));
        expect(screen.queryByTestId('timeline-item-2')).not.toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        expect(screen.getByRole('button', { name: 'All Events' })).toHaveClass('active');
        expect(screen.getByTestId('timeline-item-2')).toBeVisible();
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it.each(['local_sqlite', 'local_codex'] as const)('adds partial %s payloads without losing unmatched facts', async (source) => {
        vi.mocked(apiClient.fetchWorkspaceEvents).mockResolvedValue({ data: {
            ticket_id: 'AL-1', is_hydrated: true, data_source: source,
            events: [{ ...recorded[0]!, tool_input: 'make test', tool_output: 'failure details' }],
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent(source === 'local_codex' ? 'Source: Local Codex' : 'Source: Local SQLite'));
        expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded');
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
        expect(screen.getByTestId('timeline-item-2')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 1' }));
        expect(screen.getByTestId('tool-input-payload-block')).toHaveTextContent('make test');
        fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 2' }));
        expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent('Payload details unavailable');
    });

    it('renders a nonempty unhydrated response even when the detail contains no events', async () => {
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: makeDetail('AL-1', []) });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockResolvedValue({ data: {
            ...emptyResponse().data, events: recorded,
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
    });

    it.each(['omitted', 'empty'])('keeps current summary-only anomalies, threads, and artifacts with %s events', async (mode) => {
        const user = userEvent.setup();
        const detail = makeDetail('AL-1', []);
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...detail, anomalies: [{ rule_id: 'CONTEXT_SATURATION_SPIKE', severity: 'WARNING', context: { turn: 5, token_delta: 50000 } }],
            ledger: { ...detail.ledger, facts: { ...detail.ledger.facts, sessions: [{
                id: 'summary-only', source: 'opencode', user: 'tester', invocations: [], tools: [],
                ...(mode === 'empty' ? { events: [] } : {}),
            }] } },
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded'));
        expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('Zero Errors, 1 Anomaly');
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();
        expect(screen.queryByTestId('legacy-fallback-banner')).not.toBeInTheDocument();
        expect(screen.getByText('Session Execution Threads')).toBeInTheDocument();
        expect(screen.getByText('ses-al1-root')).toBeInTheDocument();
        expect(screen.getByText(/Contract: AL-1/i)).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        const card = screen.getByTestId('trace-waterfall-card');
        expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
        expect(within(card).getByText('Zero chronological events recorded for this execution trace.')).toBeVisible();
        expect(screen.queryByRole('group', { name: 'Filter events by type' })).not.toBeInTheDocument();
        expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();

        await user.click(screen.getByRole('button', { name: 'Collapse Trace Waterfall' }));
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        expect(screen.getByText('Session Execution Threads')).toBeInTheDocument();
        expect(screen.getByText(/Contract: AL-1/i)).toBeInTheDocument();
        await user.click(screen.getByRole('button', { name: 'Expand Trace Waterfall' }));
        expect(screen.getAllByTestId('workspace-anomalies')).toHaveLength(1);
        expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
        await user.click(within(card).getByRole('button', { name: /Inspect anomaly: Context saturation spike/ }));
        expect(screen.getByRole('dialog')).toHaveTextContent('CONTEXT_SATURATION_SPIKE');
        await user.click(screen.getByRole('button', { name: 'Close' }));
        await user.click(screen.getByRole('button', { name: /summary\.md/i }));
        expect(screen.getByText(/Summary: AL-1 Completed/i)).toBeInTheDocument();
    });

    it.each([1, undefined])('schema metadata %s cannot hide supplied events or infer legacy', async (version) => {
        const detail = makeDetail();
        const { schema_version, ...facts } = detail.ledger.facts;
        expect(schema_version).toBe(4);
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...detail, ledger: { ...detail.ledger, facts: { ...facts, ...(version === undefined ? {} : { schema_version: version }) } },
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        expect(Boolean(screen.queryByTestId('legacy-fallback-banner'))).toBe(version === 1);
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
    });

    it.each(['recorded', 'summary-only'])('ignores ticket A %s enrichment after selecting ticket B and resets the drawer and disclosure', async (mode) => {
        const pendingA = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        const bEvent: EventFact = { ...recorded[0]!, event_id: 'b-event', session_id: 'b-session', turn: 8 };
        const anomaly: Anomaly = {
            rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
            context: { session_id: recorded[0]!.session_id, tool_name: 'bash', turn: 1, count: 3 },
        };
        vi.mocked(apiClient.fetchWorkspaceById).mockImplementation(async (id) => ({ data: id === 'AL-2'
            ? makeDetail(id, [bEvent])
            : { ...makeDetail(id, mode === 'summary-only' ? [] : recorded), anomalies: [anomaly] },
        }));
        vi.mocked(apiClient.fetchWorkspaceEvents).mockImplementation((id) => id === 'AL-1' ? pendingA.promise : Promise.resolve({ data: {
            ticket_id: 'AL-2', is_hydrated: true, data_source: 'local_codex', events: [{ ...bEvent, tool_output: 'B details' }],
        } }));
        const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count'))
            .toHaveTextContent(`${mode === 'summary-only' ? 0 : 2} Events Recorded`));
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(mode === 'summary-only'
            ? within(screen.getByTestId('trace-waterfall-card')).getByRole('button', { name: /Inspect anomaly: Tool retry streak/ })
            : screen.getByTestId('timeline-item-1'));
        expect(screen.getByRole('dialog')).toHaveTextContent('TOOL_RETRY_STREAK');
        rerender(<TicketDeepDive selectedTicketId="AL-2" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex'));
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        expect(screen.getByTestId('waterfall-expand-btn')).toHaveAttribute('aria-expanded', 'false');
        expect(screen.queryByTestId('timeline-item-8')).not.toBeInTheDocument();
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByText('Detected Workspace Anomalies:')).not.toBeInTheDocument();
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByTestId('timeline-item-8'));
        await act(async () => pendingA.resolve({ data: {
            ticket_id: 'AL-1', is_hydrated: true, data_source: 'local_sqlite', events: [{ ...recorded[0]!, tool_output: 'A stale details' }],
        } }));
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex');
        expect(screen.getByText('Showing 1 of 1 events')).toBeInTheDocument();
        expect(screen.getByTestId('timeline-item-8')).toHaveClass('selected');
        expect(screen.queryByTestId('timeline-item-1')).not.toBeInTheDocument();
        expect(screen.getByTestId('waterfall-collapse-btn')).toHaveAttribute('aria-expanded', 'true');
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.getByRole('dialog')).toHaveTextContent('b-session');
        expect(screen.getByRole('dialog')).not.toHaveTextContent('TOOL_RETRY_STREAK');
        expect(screen.queryByText('A stale details')).not.toBeInTheDocument();
    });

    it('ignores an older refresh response after a newer request completes', async () => {
        const pending = deferred<Awaited<ReturnType<typeof apiClient.fetchWorkspaceEvents>>>();
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValueOnce({ data: {
            ...makeDetail(), ledger: null as unknown as AgenticMetricsLedger,
        } });
        vi.mocked(apiClient.fetchWorkspaceEvents).mockResolvedValueOnce(emptyResponse()).mockReturnValueOnce(pending.promise)
            .mockImplementation(async (id) => ({ data: {
                ticket_id: id, is_hydrated: true, data_source: 'local_codex', events: [{ ...recorded[0]!, tool_output: 'New details' }],
            } }));
        const { rerender } = render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByRole('button', { name: 'Refresh Data' })).toBeInTheDocument());
        fireEvent.click(screen.getByRole('button', { name: 'Refresh Data' }));
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded'));
        expect(screen.getByTestId('waterfall-expand-btn')).toHaveAttribute('aria-expanded', 'false');
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        expect(screen.getByRole('dialog')).toBeInTheDocument();
        rerender(<TicketDeepDive selectedTicketId="AL-2" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex'));
        expect(screen.getByTestId('waterfall-expand-btn')).toHaveAttribute('aria-expanded', 'false');
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        rerender(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(apiClient.fetchWorkspaceById).toHaveBeenCalledTimes(4));
        await waitFor(() => expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex'));
        expect(screen.getByTestId('waterfall-expand-btn')).toHaveAttribute('aria-expanded', 'false');
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(screen.getByTestId('timeline-item-1'));
        await act(async () => pending.resolve({ data: {
            ticket_id: 'AL-1', is_hydrated: true, data_source: 'local_sqlite', events: [{ ...recorded[0]!, tool_output: 'Old details' }],
        } }));
        expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Local Codex');
        expect(screen.getByText('Showing 1 of 2 events')).toBeInTheDocument();
        expect(screen.getByRole('dialog')).toHaveTextContent('recorded-session');
        expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
        expect(screen.getByRole('button', { name: 'Hide payload for turn 1' })).toBeInTheDocument();
        expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('New details');
        expect(screen.getByTestId('tool-output-payload-block')).not.toHaveTextContent('Old details');
    });

    it.each(['error', 'failed'])('opens a correlated anomaly from a bare recorded %s fact', async (status) => {
        const event = { ...recorded[0]!, status };
        vi.mocked(apiClient.fetchWorkspaceById).mockResolvedValue({ data: {
            ...makeDetail('AL-1', [event]), anomalies: [{
                rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING',
                context: { session_id: event.session_id, tool_name: 'bash', turn: 1, count: 3 },
            }],
        } });
        render(<TicketDeepDive selectedTicketId="AL-1" />);
        await waitFor(() => expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('1 Events Recorded'));
        fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
        fireEvent.click(within(screen.getByTestId('timeline-item-1')).getByText('WARNING'));
        expect(screen.getByRole('dialog')).toHaveTextContent('TOOL_RETRY_STREAK');
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
    describe('Waterfall, Legacy Fallback, and AnomalyDrawer Integration', () => {
        it('renders an immediately visible current empty trace without anomalies or a legacy claim', async () => {
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: withRecordedEvents(mockWorkspaceDetails['AL-1']!, []),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-1',
                    is_hydrated: false,
                    data_source: 'committed_ledger',
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
                ...withRecordedEvents(mockWorkspaceDetails['AL-1']!, []),
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
                    is_hydrated: false,
                    data_source: 'committed_ledger',
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
                    is_hydrated: false,
                    events: mockAL2Events,
                    data_source: 'committed_ledger',
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
                    is_hydrated: false,
                    events: mockAL2Events,
                    data_source: 'committed_ledger',
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
                    is_hydrated: false,
                    events: mockAL2Events,
                    data_source: 'committed_ledger',
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
                    is_hydrated: false,
                    events: mockAL2Events,
                    data_source: 'committed_ledger',
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
            const plainDetail = withRecordedEvents({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, [plainErrorEvent]);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: plainDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: false,
                    data_source: 'committed_ledger',
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
            const plainDetail = withRecordedEvents({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, [plainErrorEvent]);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: plainDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: false,
                    data_source: 'committed_ledger',
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
            const richDetail = withRecordedEvents({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, richEvents);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: richDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: true,
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
            const offlineDetail = withRecordedEvents({ ...mockWorkspaceDetails['AL-2']!, anomalies: [] }, offlineEvents);
            vi.spyOn(apiClient, 'fetchWorkspaceLedgers').mockResolvedValue({
                data: Object.values(mockAgenticMetricsLedgers),
            });
            vi.spyOn(apiClient, 'fetchWorkspaceById').mockResolvedValue({
                data: offlineDetail,
            });
            vi.spyOn(apiClient, 'fetchWorkspaceEvents').mockResolvedValue({
                data: {
                    ticket_id: 'AL-2',
                    is_hydrated: false,
                    events: offlineEvents,
                    data_source: 'committed_ledger',
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
            expect(sourceBadge).toHaveTextContent('Source: Committed Ledger');
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
