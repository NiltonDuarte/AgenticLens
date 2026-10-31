import type { SessionFact } from './types';
import type { EventFact, WorkspaceEventsResponse } from '../types/telemetry';

export function sortEvents(events: readonly EventFact[]): readonly EventFact[] {
  return [...events].sort((left, right) => {
    if (left.timestamp !== right.timestamp) return left.timestamp < right.timestamp ? -1 : 1;
    return left.turn - right.turn;
  });
}

export function sessionEvents(sessions: readonly SessionFact[] | undefined): readonly EventFact[] {
  return sortEvents(sessions?.flatMap((session) =>
    (session.events ?? []).map((event) => ({ ...event, session_id: session.id }))
  ) ?? []);
}

function occurrenceOf(event: EventFact): string {
  return JSON.stringify([
    event.session_id, event.timestamp, event.turn, event.event_type, event.tool_name ?? '',
  ]);
}

function enrichEvent(recorded: EventFact, local: EventFact): EventFact {
  return {
    ...recorded,
    event_id: recorded.event_id || local.event_id,
    message_id: local.message_id || recorded.message_id,
    error_message: local.error_message || recorded.error_message,
    duration_ms: local.duration_ms ?? recorded.duration_ms,
    prompt_content: local.prompt_content || recorded.prompt_content,
    reasoning_content: local.reasoning_content || recorded.reasoning_content,
    tool_input: local.tool_input || recorded.tool_input,
    tool_output: local.tool_output || recorded.tool_output,
  };
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

/** Reconcile independent source occurrences, preserving ambiguous rows even when their values agree. */
export function reconcileEvents(
  baseline: readonly EventFact[],
  enrichment: readonly EventFact[],
): readonly EventFact[] {
  const result = [...baseline];
  const matches = matchEventIndexes(baseline, enrichment);
  const enrichmentUsed = new Set(matches.values());
  for (const [recordedIndex, localIndex] of matches) {
    const recorded = baseline[recordedIndex];
    const local = enrichment[localIndex];
    if (recorded && local) result[recordedIndex] = enrichEvent(recorded, local);
  }
  enrichment.forEach((event, index) => {
    if (!enrichmentUsed.has(index)) result.push(event);
  });
  return sortEvents(result);
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

export function selectWorkspaceEvents(
  baseline: readonly EventFact[],
  response?: WorkspaceEventsResponse,
): Pick<WorkspaceEventsResponse, 'events' | 'is_hydrated' | 'data_source'> {
  if (!response?.events.length) {
    return { events: sortEvents(baseline), is_hydrated: false, data_source: 'committed_ledger' };
  }
  return {
    events: response.includes_baseline ? sortEvents(response.events) : reconcileEvents(baseline, response.events),
    is_hydrated: response.is_hydrated,
    data_source: response.is_hydrated ? response.data_source : 'committed_ledger',
  };
}
