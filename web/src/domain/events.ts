import { rebindAnomaly } from './anomalies';
import type { AgenticFactLedger } from './types';
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

export type AnomalyRoleOrigin = 'source' | 'projected';

/** Presentation provenance is internal; canonical facts and API schemas stay unchanged. */
export interface PresentedAnomaly {
  readonly anomaly: Anomaly;
  readonly roleOrigin: AnomalyRoleOrigin;
}

/** Recover raw occurrences from canonical facts, restoring only their enclosing session owner. */
export function recordedAnomalyFacts(ledger: AgenticFactLedger): readonly Anomaly[] | undefined {
  const recorded = ledger.anomalies === undefined && ledger.sessions.every((owner) => owner.anomalies === undefined)
    ? undefined : [...(ledger.anomalies ?? [])];
  if (recorded === undefined) return undefined;
  for (const owner of ledger.sessions) {
    for (const anomaly of owner.anomalies ?? []) {
      recorded.push(anomaly.context.session_id == null ? {...anomaly, context: {...anomaly.context, session_id: owner.id}} : anomaly);
    }
  }
  return recorded;
}

export function recordedAnomalyEntries(ledger: AgenticFactLedger): readonly PresentedAnomaly[] {
  return (recordedAnomalyFacts(ledger) ?? []).map((anomaly) => ({anomaly, roleOrigin: 'source'}));
}

/** Only complete owner/group reads replace raw recorded occurrences. Display aliases are never identity. */
export function selectWorkspaceAnomalyEntries(ledger: AgenticFactLedger, response: WorkspaceEventsResponse): readonly PresentedAnomaly[] {
  const recorded = recordedAnomalyFacts(ledger) ?? [];
  const rawEntry = (anomaly: Anomaly): PresentedAnomaly => ({anomaly, roleOrigin: 'source'});
  if (response.availability === 'unavailable' || response.anomalies === undefined) return recorded.map(rawEntry);
  if (response.availability === 'complete') return response.anomalies.map(rawEntry);
  const completeOwners = new Set(response.coverage.complete_session_ids);
  const retained = recorded.filter((anomaly) => {
    const owner = anomaly.context.session_id;
    if (!owner || !completeOwners.has(owner)) return true;
    return anomaly.rule_id === 'SUBAGENT_REWORK_LOOP' && ledger.sessions.some((child) => child.parent_id === owner && !completeOwners.has(child.id));
  });
  const identity = (anomaly: Anomaly) => JSON.stringify([anomaly.rule_id, Object.entries(anomaly.context).sort()]);
  const occurrences = new Map<string, number[]>();
  retained.forEach((anomaly, index) => {
    const key = identity(anomaly);
    const indexes = occurrences.get(key) ?? [];
    indexes.push(index);
    occurrences.set(key, indexes);
  });
  const enriched = [...retained];
  const refreshed: PresentedAnomaly[] = [];
  for (const anomaly of response.anomalies) {
    const indexes = occurrences.get(identity(anomaly));
    const sameSeverity = indexes?.findIndex((index) => retained[index]!.severity === anomaly.severity) ?? -1;
    const index = sameSeverity >= 0 ? indexes!.splice(sameSeverity, 1)[0] : indexes?.shift();
    if (index === undefined) refreshed.push(rawEntry(anomaly));
    else if (anomaly.evidence !== undefined) enriched[index] = {...retained[index]!, evidence: anomaly.evidence};
  }
  return [...refreshed, ...enriched.map(rawEntry)];
}

/** Trace entries have trustworthy raw roles; projected section selections stay separately owned. */
export function rebindPresentedAnomaly(selected: PresentedAnomaly | null, entries: readonly PresentedAnomaly[]): PresentedAnomaly | null {
  if (!selected) return null;
  const candidates = entries.filter((entry) => entry.roleOrigin === selected.roleOrigin);
  const match = rebindAnomaly(selected.anomaly, candidates.map((entry) => entry.anomaly));
  return candidates.find((entry) => entry.anomaly === match) ?? null;
}
