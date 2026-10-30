// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { rebindEvent, reconcileEvents, selectWorkspaceEvents, sessionEvents } from './events';
import {
    sessionTools,
    sessionInvocations,
    type AgenticFactLedger,
    type EventFact,
    type EventType,
    type SessionFact,
    ledgerSkills,
    ledgerTaskLinks,
} from './types';
describe('Event Telemetry and Schema Evolution Contracts', () => {
    describe('Legacy and Backward Compatibility Boundaries', () => {
        it('deserializes legacy JSON omitting sessions without error', () => {
            const legacyPayload = JSON.stringify({
                schema_version: 2,
                ticket_id: 'AL-LEGACY',
                ticket_title: 'Legacy ticket',
                updated_at: '2026-10-18T10:00:00Z',
                code: {
                    lines_added: 50,
                    lines_deleted: 10,
                    files_changed: 2,
                    sdlc_iterations: 1,
                },
                sessions: []
            });
            const parsed: AgenticFactLedger = JSON.parse(legacyPayload);
            expect(parsed.ticket_id).toBe('AL-LEGACY');
            expect(parsed.sessions).toEqual([]);
            expect(ledgerSkills(parsed)).toEqual([]);
            expect(ledgerTaskLinks(parsed)).toEqual([]);
        });
        it('reads additive skill and task facts while keeping legacy omissions empty', () => {
            const payload: AgenticFactLedger = {
                schema_version: 4,
                ticket_id: 'AL-51', ticket_title: null, ticket_type: 'feat',
                started_at: null, completed_at: null, updated_at: '',
                code: { lines_added: 0, lines_deleted: 0, test_lines_added: 0, test_lines_deleted: 0,
                    prod_lines_added: 0, prod_lines_deleted: 0, files_changed: 0, sdlc_iterations: 0 },
                sessions: [],
                skills: [{ name: 'agentic-lens-dev', turn_index: 1, message_index: 0, caller_role: 'coder',
                    timestamp: '2026-10-31T10:00:00Z', success: true }],
                task_links: [{ subagent_type: 'tester', description: 'Run checks', parent_turn: 2,
                    child_session_id: 'child-1', timestamp: '2026-10-31T10:01:00Z' }],
            };
            expect(ledgerSkills(payload)).toHaveLength(1);
            expect(ledgerTaskLinks(payload)[0]?.child_session_id).toBe('child-1');
        });
        it('handles undefined tools and invocations in session helpers without throwing', () => {
            const sessionsWithoutArrays = [
                { id: 'ses-1' } as unknown as SessionFact,
            ];
            expect(sessionTools(sessionsWithoutArrays)).toEqual([]);
            expect(sessionInvocations(sessionsWithoutArrays)).toEqual([]);
            expect(sessionTools(undefined)).toEqual([]);
            expect(sessionInvocations(undefined)).toEqual([]);
        });
    });
    describe('EventFact Type Discrimination and Property Fidelity', () => {
        const mockEvents: readonly EventFact[] = [
            {
                timestamp: '2026-10-18T10:00:01.000000+00:00',
                session_id: 'ses-orch-1',
                turn: 1,
                role: 'orchestrator',
                event_type: 'turn_message',
                status: 'completed',
                duration_ms: 1200,
                context_tokens: 5000,
            },
            {
                timestamp: '2026-10-18T10:00:02.000000+00:00',
                session_id: 'ses-orch-1',
                turn: 1,
                role: 'orchestrator',
                event_type: 'tool_call',
                tool_name: 'bash',
                status: 'error',
                error_message: 'Traceback (most recent call last):\n  File script.py, line 2',
                duration_ms: 450,
                context_tokens: 5000,
            },
            {
                timestamp: '2026-10-18T10:00:03.000000+00:00',
                session_id: 'ses-orch-1',
                turn: 2,
                role: 'orchestrator',
                event_type: 'compaction',
                status: 'completed',
                context_tokens: 3800,
            },
        ];
        it('correctly discriminates event_type enums', () => {
            const types: EventType[] = mockEvents.map((e) => e.event_type);
            expect(types).toEqual(['turn_message', 'tool_call', 'compaction']);
        });
        it('preserves tool call error excerpt and execution duration', () => {
            const toolEvent = mockEvents[1];
            expect(toolEvent.event_type).toBe('tool_call');
            expect(toolEvent.tool_name).toBe('bash');
            expect(toolEvent.status).toBe('error');
            expect(toolEvent.error_message).toContain('Traceback');
            expect(toolEvent.duration_ms).toBe(450);
            expect(toolEvent.context_tokens).toBe(5000);
        });
        it('preserves compaction events with turn and context volume', () => {
            const compEvent = mockEvents[2];
            expect(compEvent.event_type).toBe('compaction');
            expect(compEvent.status).toBe('completed');
            expect(compEvent.turn).toBe(2);
            expect(compEvent.context_tokens).toBe(3800);
            expect(compEvent.tool_name).toBeUndefined();
            expect(compEvent.error_message).toBeUndefined();
        });
        it('verifies chronological ascending ordering and turn tie-breaking logic', () => {
            const unsortedEvents: EventFact[] = [
                {
                    timestamp: '2026-10-18T10:00:02.000000+00:00',
                    session_id: 'ses-1',
                    turn: 2,
                    role: 'coder',
                    event_type: 'turn_message',
                    status: 'completed',
                    context_tokens: 8000,
                },
                {
                    timestamp: '2026-10-18T10:00:01.000000+00:00',
                    session_id: 'ses-1',
                    turn: 2,
                    role: 'coder',
                    event_type: 'compaction',
                    status: 'completed',
                    context_tokens: 6000,
                },
                {
                    timestamp: '2026-10-18T10:00:01.000000+00:00',
                    session_id: 'ses-1',
                    turn: 1,
                    role: 'coder',
                    event_type: 'turn_message',
                    status: 'completed',
                    context_tokens: 4000,
                },
            ];
            const sorted = [...unsortedEvents].sort((a, b) => {
                if (a.timestamp !== b.timestamp) {
                    return a.timestamp.localeCompare(b.timestamp);
                }
                if (a.turn !== b.turn) {
                    return a.turn - b.turn;
                }
                return a.session_id.localeCompare(b.session_id);
            });
            expect(sorted[0].turn).toBe(1);
            expect(sorted[0].event_type).toBe('turn_message');
            expect(sorted[1].turn).toBe(2);
            expect(sorted[1].event_type).toBe('compaction');
            expect(sorted[2].turn).toBe(2);
            expect(sorted[2].timestamp).toBe('2026-10-18T10:00:02.000000+00:00');
        });
    });
});

describe('Recorded event projection and optional enrichment', () => {
    const recorded: EventFact = Object.freeze({
        event_id: 'call-1', timestamp: '2026-10-29T10:00:00Z', session_id: 'session-a',
        message_id: 'message-1', turn: 2, role: 'coder', event_type: 'tool_call',
        tool_name: 'bash', status: 'error', duration_ms: 12, context_tokens: 3400,
    });

    it('projects every nested field with its owner and stable chronological ordering without mutation', () => {
        const { session_id: owner, ...nested } = recorded;
        const sessions: readonly SessionFact[] = Object.freeze([
            Object.freeze({ id: owner, invocations: [], tools: [], events: Object.freeze([
                Object.freeze({ ...nested, event_id: 'later', timestamp: '2026-10-29T11:00:00Z' }),
                Object.freeze(nested),
                Object.freeze({ ...nested, event_id: 'tie' }),
            ]) }),
            Object.freeze({ id: 'session-b', invocations: [], tools: [], events: Object.freeze([
                Object.freeze({ ...nested, event_id: 'other-owner' }),
                Object.freeze({ ...nested, event_id: 'earlier-turn', turn: 1 }),
            ]) }),
        ]);
        const before = JSON.stringify(sessions);
        const projected = sessionEvents(sessions);
        expect(projected.map((event) => event.event_id)).toEqual(['earlier-turn', 'call-1', 'tie', 'other-owner', 'later']);
        expect(projected[1]).toEqual(recorded);
        expect(projected[3]).toEqual({ ...recorded, event_id: 'other-owner', session_id: 'session-b' });
        expect(projected[1]).not.toBe(nested);
        expect(JSON.stringify(sessions)).toBe(before);
    });

    it.each([
        undefined, [], [{ id: 'omitted', invocations: [], tools: [] }],
        [{ id: 'empty', invocations: [], tools: [], events: [] }],
    ])('accepts absent or empty optional collections without inferring a schema: %j', (sessions) => {
        expect(sessionEvents(sessions)).toEqual([]);
    });

    it('keeps a v4 row without optional identifiers or payloads', () => {
        const bare = { timestamp: recorded.timestamp, turn: 0, role: 'tester', event_type: 'turn_message' as const, status: 'completed', context_tokens: 0 };
        expect(sessionEvents([{ id: 'bare', invocations: [], tools: [], events: [bare] }])).toEqual([{ ...bare, session_id: 'bare' }]);
    });

    it('enriches matching IDs, retains unmatched facts, adds local rows, and is idempotent', () => {
        const unmatched = Object.freeze({ ...recorded, event_id: 'unmatched', turn: 3 });
        const local = Object.freeze({ ...recorded, message_id: 'local-message', duration_ms: 0,
            prompt_content: 'Run checks', reasoning_content: 'Inspect failures', tool_input: 'make', tool_output: 'failed', error_message: 'exit 1' });
        const added = Object.freeze({ ...local, event_id: 'local-only', session_id: 'session-b', turn: 1 });
        const baseline = Object.freeze([unmatched, recorded]);
        const enrichment = Object.freeze([added, local]);
        const before = JSON.stringify({ baseline, enrichment });
        const merged = reconcileEvents(baseline, enrichment);
        expect(merged).toEqual([added, { ...recorded, ...local }, unmatched]);
        expect(reconcileEvents(merged, enrichment)).toEqual(merged);
        expect(reconcileEvents(baseline, merged)).toEqual(merged);
        expect(JSON.stringify({ baseline, enrichment })).toBe(before);
    });

    it('keeps different IDs and different sessions distinct even on the same turn', () => {
        expect(reconcileEvents([recorded], [{ ...recorded, event_id: 'different' }])).toHaveLength(2);
        expect(reconcileEvents([recorded], [{ ...recorded, session_id: 'session-b' }])).toHaveLength(2);
    });

    it('matches missing IDs only for an unambiguous occurrence and compatible message IDs', () => {
        const bare = { ...recorded, event_id: undefined };
        const local = { ...recorded, tool_output: 'details' };
        expect(reconcileEvents([bare], [local])).toEqual([local]);
        for (const different of [
            { ...local, timestamp: '2026-10-29T12:00:00Z' }, { ...local, turn: 9 },
            { ...local, event_type: 'turn_message' as const }, { ...local, tool_name: 'read' },
            { ...local, message_id: 'different' }, { ...local, session_id: '' },
        ]) {
            expect(reconcileEvents([bare], [different])).toHaveLength(2);
        }
        const ambiguous = reconcileEvents([bare, { ...bare }], [local]);
        expect(ambiguous).toHaveLength(3);
        expect(ambiguous.filter((event) => event.tool_output)).toEqual([local]);
        expect(reconcileEvents(ambiguous, [local])).toEqual(ambiguous);
        expect(reconcileEvents([bare], [local, { ...local, event_id: 'second' }])).toHaveLength(3);
    });

    it.each([[2, 1], [1, 2], [2, 2]])('preserves %i baseline and %i fresh identical ID-less occurrences', (baselineCount, freshCount) => {
        const bare = { ...recorded, event_id: undefined, message_id: undefined };
        const baseline = Object.freeze(Array.from({ length: baselineCount }, () => Object.freeze({ ...bare })));
        const fresh = Object.freeze(Array.from({ length: freshCount }, () => Object.freeze({ ...bare })));
        const before = JSON.stringify({ baseline, fresh });
        expect(reconcileEvents(baseline, fresh)).toEqual([...baseline, ...fresh]);
        expect(JSON.stringify({ baseline, fresh })).toBe(before);
    });

    it.each([undefined, false])('treats identical ambiguous response rows as fresh without baseline provenance=%s', (provenance) => {
        const bare = { ...recorded, event_id: undefined, message_id: undefined };
        const baseline = [bare, { ...bare }];
        expect(selectWorkspaceEvents(baseline, {
            ticket_id: 'FIX-12', is_hydrated: true, data_source: 'local_sqlite',
            includes_baseline: provenance, events: [{ ...bare }],
        }).events).toEqual([...baseline, bare]);
    });

    it.each([false, true])('repeated known baseline responses preserve ambiguous multiplicity with hydration=%s', (hydrated) => {
        const bare = { ...recorded, event_id: undefined, message_id: undefined };
        const baseline = Object.freeze([Object.freeze({ ...bare }), Object.freeze({ ...bare })]);
        const response = Object.freeze({
            ticket_id: 'FIX-12', is_hydrated: hydrated, data_source: 'local_sqlite' as const,
            includes_baseline: true, events: Object.freeze([...baseline.map((event) => Object.freeze({ ...event })), Object.freeze({ ...bare })]),
        });
        let events: readonly EventFact[] = baseline;
        for (let round = 0; round < 3; round++) {
            const selected = selectWorkspaceEvents(events, response);
            expect(selected.events).toEqual(response.events);
            expect(selected.events).toHaveLength(3);
            expect(selected.is_hydrated).toBe(hydrated);
            expect(selected.data_source).toBe(hydrated ? 'local_sqlite' : 'committed_ledger');
            events = selected.events;
        }
        expect(baseline).toHaveLength(2);
    });

    it('does not rebind a missing-ID selection to an ambiguous or different occurrence', () => {
        const bare = { ...recorded, event_id: undefined, message_id: undefined };
        expect(rebindEvent(bare, [bare, { ...bare }], [{ ...bare, tool_output: 'ambiguous' }])).toBeNull();
        expect(rebindEvent(bare, [bare], [{ ...bare }, { ...bare }])).toBeNull();
        expect(rebindEvent(bare, [bare], [{ ...bare, session_id: 'other-session' }])).toBeNull();
        expect(rebindEvent(recorded, [recorded], [{ ...recorded, event_id: 'different' }])).toBeNull();
    });

    it.each([false, true])('empty enrichment retains committed provenance even with hydration=%s', (hydrated) => {
        expect(selectWorkspaceEvents([recorded], {
            ticket_id: 'FIX-12', is_hydrated: hydrated, data_source: 'local_sqlite', events: [],
        })).toEqual({ events: [recorded], is_hydrated: false, data_source: 'committed_ledger' });
        expect(selectWorkspaceEvents([recorded])).toEqual({ events: [recorded], is_hydrated: false, data_source: 'committed_ledger' });
    });

    it.each(['local_sqlite', 'local_codex'] as const)('uses nonempty %s enrichment regardless of the hydration flag', (source) => {
        const local = { ...recorded, tool_output: 'result' };
        for (const hydrated of [false, true]) {
            expect(selectWorkspaceEvents([], { ticket_id: 'FIX-12', is_hydrated: hydrated, data_source: source, events: [local] }))
                .toEqual({ events: [local], is_hydrated: hydrated, data_source: hydrated ? source : 'committed_ledger' });
        }
    });
});
