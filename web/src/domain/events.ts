import type { Anomaly, EventFact, WorkspaceEventsResponse } from '../types/telemetry';

export function sortEvents(events: readonly EventFact[]): readonly EventFact[] {
  return [...events].sort((left, right) => {
    if (left.timestamp !== right.timestamp) return left.timestamp < right.timestamp ? -1 : 1;
    return left.turn - right.turn;
  });
}

function occurrenceOf(event: EventFact): string {
  return JSON.stringify([
    event.session_id, event.timestamp, event.turn, event.event_type, event.tool_name ?? '',
  ]);
}

/** Stable IDs take precedence; missing IDs require a unique occurrence on both sides. */
function matchEventIndexes(
  baseline: readonly EventFact[],
  enrichment: readonly EventFact[],
): ReadonlyMap<number, number> {
  const matches = new Map<number, number>();
  const enrichmentUsed = new Set<number>();
  const byID = new Map<string, number[]>();
  const byOccurrence = new Map<string, number[]>();
  baseline.forEach((event, index) => {
    if (!event.session_id) return;
    if (event.event_id) {
      const key = JSON.stringify([event.session_id, event.event_id]);
      const indexes = byID.get(key) ?? [];
      indexes.push(index);
      byID.set(key, indexes);
    }
    const key = occurrenceOf(event);
    const indexes = byOccurrence.get(key) ?? [];
    indexes.push(index);
    byOccurrence.set(key, indexes);
  });
  enrichment.forEach((event, index) => {
    if (!event.session_id || !event.event_id) return;
    const key = JSON.stringify([event.session_id, event.event_id]);
    const matched = byID.get(key)?.shift();
    const recorded = matched === undefined ? undefined : baseline[matched];
    if (matched === undefined || !recorded) return;
    matches.set(matched, index);
    enrichmentUsed.add(index);
  });

  // Count before matching so consuming one row cannot resolve an ambiguous group.
  const candidates = new Map<number, number>();
  const baselineMatches = new Map<number, number>();
  enrichment.forEach((event, index) => {
    if (enrichmentUsed.has(index) || !event.session_id) return;
    const occurrenceMatches = (byOccurrence.get(occurrenceOf(event)) ?? []).filter((otherIndex) => {
      const other = baseline[otherIndex];
      return other !== undefined && !matches.has(otherIndex) &&
        !(event.event_id && other.event_id) &&
        !(event.message_id && other.message_id && event.message_id !== other.message_id);
    });
    for (const otherIndex of occurrenceMatches) {
      baselineMatches.set(otherIndex, (baselineMatches.get(otherIndex) ?? 0) + 1);
    }
    if (occurrenceMatches.length === 1 && occurrenceMatches[0] !== undefined) {
      candidates.set(index, occurrenceMatches[0]);
    }
  });
  for (const [index, matched] of candidates) {
    const recorded = baseline[matched];
    const local = enrichment[index];
    if (baselineMatches.get(matched) === 1 && recorded && local) {
      matches.set(matched, index);
      enrichmentUsed.add(index);
    }
  }
  return matches;
}

/** Retain an existing row reference or rebind using the reconciliation identity rules. */
export function rebindEvent(
  selected: EventFact | null,
  previous: readonly EventFact[],
  events: readonly EventFact[],
): EventFact | null {
  if (!selected) return null;
  if (events.includes(selected)) return selected;
  const previousIndex = previous.indexOf(selected);
  if (previousIndex < 0) return null;
  const matched = matchEventIndexes(previous, events).get(previousIndex);
  return matched === undefined ? null : events[matched] ?? null;
}

/** Trace rows come only from the transient source response, including a complete empty result. */
export function selectWorkspaceEvents(response: WorkspaceEventsResponse): WorkspaceEventsResponse {
  return { ...response, events: sortEvents(response.events) };
}

/** Only complete owner reads can replace a recorded anomaly summary. */
export function selectWorkspaceAnomalies(recorded: readonly Anomaly[], response: WorkspaceEventsResponse): readonly Anomaly[] {
  if (response.availability === 'unavailable' || response.anomalies === undefined) return recorded;
  if (response.availability === 'complete') return response.anomalies ?? recorded;
  const completeOwners = new Set(response.coverage.complete_session_ids);
  const retained = recorded.filter((anomaly) => !anomaly.context.session_id || !completeOwners.has(anomaly.context.session_id));
  const identity = (anomaly: Anomaly) => JSON.stringify([anomaly.rule_id, Object.entries(anomaly.context).sort()]);
  const retainedKeys = new Set(retained.map(identity));
  const source = response.anomalies ?? [];
  return [...source.filter((anomaly) => !retainedKeys.has(identity(anomaly))), ...retained.map((anomaly) => {
    const enrichment = source.find((candidate) => identity(candidate) === identity(anomaly));
    return enrichment?.evidence ? { ...anomaly, evidence: enrichment.evidence } : anomaly;
  })];
}
