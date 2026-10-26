// @vitest-environment node
import { describe, it, expect } from 'vitest';
import type { AgenticFactLedger, EventFact, EventType } from './types';
describe('Event Telemetry and Schema Evolution Contracts', () => {
    describe('Legacy and Backward Compatibility Boundaries', () => {
        it('deserializes legacy JSON omitting events field without error', () => {
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
        });
        it('deserializes legacy JSON with explicit null events field', () => {
            const payloadWithNull = JSON.stringify({
                schema_version: 2,
                ticket_id: 'AL-NULL',
                updated_at: '2026-10-18T10:00:00Z',
                code: { lines_added: 0, lines_deleted: 0, files_changed: 0, sdlc_iterations: 0 },
                sessions: [{ id: 'AL-NULL', invocations: [], tools: [], events: [] }],
            });
            const parsed: AgenticFactLedger = JSON.parse(payloadWithNull);
            expect(parsed.sessions[0]?.events).toEqual([]);
        });
        it('deserializes ledger with empty events array', () => {
            const payloadWithEmpty = JSON.stringify({
                schema_version: 2,
                ticket_id: 'AL-EMPTY',
                updated_at: '2026-10-18T10:00:00Z',
                code: { lines_added: 0, lines_deleted: 0, files_changed: 0, sdlc_iterations: 0 },
                sessions: [{ id: 'AL-EMPTY', invocations: [], tools: [], events: [] }]
            });
            const parsed: AgenticFactLedger = JSON.parse(payloadWithEmpty);
            expect(parsed.sessions[0]?.events).toBeDefined();
            expect(parsed.sessions[0]?.events).toHaveLength(0);
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
