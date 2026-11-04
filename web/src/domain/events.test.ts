// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { rebindEvent, selectWorkspaceEvents, selectWorkspaceAnomalies } from './events';
import { sessionTools, sessionInvocations, type Anomaly, type EventFact, type SessionFact } from './types';

describe('transient source traces', () => {
  const recorded: EventFact = Object.freeze({ event_id: 'call-1', timestamp: '2026-10-29T10:00:00Z', session_id: 'owner',
    message_id: 'message-1', turn: 2, role: 'coder', event_type: 'tool_call', tool_name: 'bash', status: 'error', duration_ms: 12, context_tokens: 3400 });
  const coverage = { required_session_ids: ['owner'], complete_session_ids: ['owner'], diagnostics: [] };

  it('sorts actual source events without mutating or deduplicating equal occurrences', () => {
    const events = Object.freeze([{ ...recorded, timestamp: '2026-10-29T11:00:00Z' }, recorded, { ...recorded }]);
    const response = Object.freeze({ ticket_id: 'TEST-1', is_hydrated: true, availability: 'complete' as const, data_source: 'local_codex' as const, coverage, events });
    const before = JSON.stringify(response);
    const selected = selectWorkspaceEvents(response);
    expect(selected.events).toEqual([recorded, recorded, events[0]]);
    expect(JSON.stringify(response)).toBe(before);
  });

  it.each(['complete', 'partial', 'unavailable'] as const)('preserves explicit %s coverage for zero events', (availability) => {
    const response = { ticket_id: 'TEST-1', is_hydrated: availability !== 'unavailable', availability, coverage,
      data_source: 'local_codex' as const, events: [] };
    expect(selectWorkspaceEvents(response)).toEqual(response);
  });

  it('rebinds only an unambiguous native source occurrence within its owner', () => {
    const enriched = { ...recorded, tool_output: 'failure detail' };
    expect(rebindEvent(recorded, [recorded], [enriched])).toBe(enriched);
    expect(rebindEvent(recorded, [recorded], [{ ...enriched, session_id: 'other' }])).toBeNull();
    expect(rebindEvent(recorded, [recorded], [{ ...enriched, event_id: 'different' }])).toBeNull();
    const bare = { ...recorded, event_id: undefined, message_id: undefined };
    expect(rebindEvent(bare, [bare, { ...bare }], [{ ...bare }])).toBeNull();
    expect(rebindEvent(bare, [bare], [{ ...bare }, { ...bare }])).toBeNull();
    expect(rebindEvent(null, [], [])).toBeNull();
  });

  it('refreshes complete owner anomaly associations during partial hydration and preserves unavailable owners', () => {
    const summary: Anomaly = { rule_id: 'TOOL_RETRY_STREAK', severity: 'WARNING', context: { session_id: 'owner', count: 3 } };
    const remote = { ...summary, context: { session_id: 'remote-owner', count: 4 } };
    const ownerless = { ...summary, context: { count: 5 } };
    const updated: Anomaly = { ...summary, severity: 'CRITICAL', context: { session_id: 'owner', count: 6 }, evidence: [{ session_id: 'owner', event_id: 'native-call' }] };
    const remoteEnriched = { ...remote, severity: 'CRITICAL' as const, evidence: [{ session_id: 'remote-owner', event_id: 'remote-call' }] };
    const summaries = Object.freeze([Object.freeze(summary), Object.freeze(remote), Object.freeze(ownerless)]);
    const before = JSON.stringify(summaries);
    const response = { ticket_id: 'TEST-1', is_hydrated: true, availability: 'partial' as const, data_source: 'local_codex' as const,
      coverage: { required_session_ids: ['owner', 'remote-owner'], complete_session_ids: ['owner'], diagnostics: ['remote unavailable'] }, events: [recorded],
      anomalies: [updated, remoteEnriched] };
    expect(selectWorkspaceAnomalies(summaries, response)).toEqual([updated, { ...remote, evidence: remoteEnriched.evidence }, ownerless]);
    expect(selectWorkspaceAnomalies(summaries, { ...response, anomalies: [] })).toEqual([remote, ownerless]);
    expect(selectWorkspaceAnomalies(summaries, { ...response, anomalies: undefined })).toBe(summaries);
    expect(selectWorkspaceAnomalies(summaries, { ...response, availability: 'unavailable' })).toBe(summaries);
    expect(selectWorkspaceAnomalies(summaries, { ...response, availability: 'complete', anomalies: [] })).toEqual([]);
    expect(JSON.stringify(summaries)).toBe(before);
  });

  it('keeps missing collection recovery in session metric helpers', () => {
    const sessions = [{ id: 'missing' } as unknown as SessionFact];
    expect(sessionTools(sessions)).toEqual([]);
    expect(sessionInvocations(sessions)).toEqual([]);
    expect(sessionTools(undefined)).toEqual([]);
    expect(sessionInvocations(undefined)).toEqual([]);
  });
});
