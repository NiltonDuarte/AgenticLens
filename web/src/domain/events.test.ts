// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { rebindEvent, selectWorkspaceEvents, recordedAnomalyFacts, selectWorkspaceAnomalyEntries, rebindPresentedAnomaly } from './events';
import { mockAgenticMetricsLedgers } from '../test/fixtures/mockData';
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
    const facts = {...mockAgenticMetricsLedgers['AL-1']!.facts, sessions: [], anomalies: summaries};
    const selected = (response: Parameters<typeof selectWorkspaceAnomalyEntries>[1]) => selectWorkspaceAnomalyEntries(facts, response).map((entry) => entry.anomaly);
    const before = JSON.stringify(summaries);
    const response = { ticket_id: 'TEST-1', is_hydrated: true, availability: 'partial' as const, data_source: 'local_codex' as const,
      coverage: { required_session_ids: ['owner', 'remote-owner'], complete_session_ids: ['owner'], diagnostics: ['remote unavailable'] }, events: [recorded],
      anomalies: [updated, remoteEnriched] };
    expect(selectWorkspaceAnomalyEntries(facts, response).map((entry) => entry.roleOrigin)).toEqual(['source', 'source', 'source']);
    expect(selectWorkspaceAnomalyEntries(facts, {...response, availability: 'unavailable'}).map((entry) => entry.roleOrigin)).toEqual(['source', 'source', 'source']);
    expect(selected(response)).toEqual([updated, { ...remote, evidence: remoteEnriched.evidence }, ownerless]);
    expect(selected({ ...response, anomalies: [] })).toEqual([remote, ownerless]);
    expect(selected({ ...response, anomalies: undefined })).toEqual(summaries);
    expect(selected({ ...response, availability: 'unavailable' })).toEqual(summaries);
    expect(selected({ ...response, availability: 'complete', anomalies: [] })).toEqual([]);
    expect(JSON.stringify(summaries)).toBe(before);
  });

  it('rebinds raw occurrences without confusing distinct personas that share a presentation alias', () => {
    const raw: Anomaly = {rule_id: 'SUBAGENT_REWORK_LOOP', severity: 'WARNING', context: {session_id: 'owner', role: 'builder', secondary_role: 'tester', count: 3}};
    const updated: Anomaly = {...raw, severity: 'CRITICAL', context: {...raw.context, count: 6}};
    const selected = {anomaly: raw, roleOrigin: 'source' as const};
    const refreshed = {anomaly: updated, roleOrigin: 'source' as const};
    expect(rebindPresentedAnomaly(selected, [refreshed])).toBe(refreshed);
    expect(rebindPresentedAnomaly(selected, [{...refreshed, anomaly: {...updated, context: {...updated.context, role: 'engineer'}}}])).toBeNull();
    expect(rebindPresentedAnomaly(selected, [{...refreshed, roleOrigin: 'projected'}])).toBeNull();
  });

  it('restores canonical owners and preserves raw multiplicity and per-occurrence evidence without alias keys', () => {
    const occurrence: Anomaly = {rule_id: 'SUBAGENT_REWORK_LOOP', severity: 'WARNING', context: {role: 'builder', secondary_role: 'tester', count: 3}};
    const engineer = {...occurrence, context: {...occurrence.context, role: 'engineer'}};
    const facts = {...mockAgenticMetricsLedgers['AL-1']!.facts, anomalies: [], sessions: [
      {id: 'remote', invocations: [], tools: [], anomalies: [occurrence, {...occurrence}, engineer]},
      {id: 'current', invocations: [], tools: [], anomalies: []},
    ]};
    const before = JSON.stringify(facts);
    const raw = recordedAnomalyFacts(facts)!;
    expect(raw.map((anomaly) => [anomaly.context.session_id, anomaly.context.role])).toEqual([['remote','builder'], ['remote','builder'], ['remote','engineer']]);
    const response = {ticket_id: 'TEST-1', availability: 'partial' as const, is_hydrated: true, data_source: 'local_codex' as const,
      coverage: {required_session_ids: ['remote','current'], complete_session_ids: ['current'], diagnostics: ['remote unavailable']}, events: [],
      anomalies: raw.map((anomaly, index) => ({...anomaly, evidence: [{session_id: 'remote', event_id: `event-${index}`}]})),
    };
    const selected = selectWorkspaceAnomalyEntries(facts, response);
    expect(selected).toHaveLength(3);
    expect(selected.map((entry) => entry.anomaly.context.role)).toEqual(['builder','builder','engineer']);
    expect(selected.map((entry) => entry.anomaly.evidence?.[0]?.event_id)).toEqual(['event-0','event-1','event-2']);
    expect(selectWorkspaceAnomalyEntries(facts, {...response, anomalies: [response.anomalies[0]!]})).toHaveLength(3);
    expect(selectWorkspaceAnomalyEntries(facts, {...response, availability: 'complete', anomalies: []})).toEqual([]);
    expect(selectWorkspaceAnomalyEntries(facts, {...response, availability: 'complete', anomalies: [...response.anomalies, response.anomalies[0]!]})).toHaveLength(4);
    expect(JSON.stringify(facts)).toBe(before);
    const warning = {...raw[0]!, evidence: [{session_id: 'remote', event_id: 'warning-proof'}]};
    const critical = {...warning, severity: 'CRITICAL' as const, evidence: [{session_id: 'remote', event_id: 'critical-proof'}]};
    const severityFacts = {...facts, anomalies: [warning, critical], sessions: []};
    expect(selectWorkspaceAnomalyEntries(severityFacts, {...response, anomalies: [critical, warning]})
      .map((entry) => [entry.anomaly.severity, entry.anomaly.evidence?.[0]?.event_id])).toEqual([['WARNING','warning-proof'], ['CRITICAL','critical-proof']]);
    const group = {...facts, sessions: [facts.sessions[0]!, {id: 'child', parent_id: 'remote', invocations: [], tools: [], anomalies: []}]};
    expect(selectWorkspaceAnomalyEntries(group, {...response, coverage: {required_session_ids: ['remote','child'], complete_session_ids: ['remote'], diagnostics: ['child unavailable']}, anomalies: []})).toHaveLength(3);
    const {anomalies: category, ...withoutCategory} = facts;
    void category;
    const noCategory = {...withoutCategory, sessions: [{id: 'unknown', invocations: [], tools: []}]};
    expect(recordedAnomalyFacts(noCategory)).toBeUndefined();
    expect(recordedAnomalyFacts({...noCategory, anomalies: []})).toEqual([]);
  });

  it('keeps missing collection recovery in session metric helpers', () => {
    const sessions = [{ id: 'missing' } as unknown as SessionFact];
    expect(sessionTools(sessions)).toEqual([]);
    expect(sessionInvocations(sessions)).toEqual([]);
    expect(sessionTools(undefined)).toEqual([]);
    expect(sessionInvocations(undefined)).toEqual([]);
  });
});
