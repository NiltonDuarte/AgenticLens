import type { EventFact, SkillFact, TaskLinkFact } from './types';

export type TraceBadgeMetadata =
  | { readonly kind: 'skill'; readonly name: string }
  | { readonly kind: 'task'; readonly name: string; readonly childSessionId?: string | undefined };

function inputString(input: string | null | undefined, key: string): string | undefined {
  if (!input) return undefined;
  try {
    const parsed: unknown = JSON.parse(input);
    if (typeof parsed !== 'object' || parsed === null || !(key in parsed)) return undefined;
    const value: unknown = Reflect.get(parsed, key);
    return typeof value === 'string' && value.trim() ? value.trim() : undefined;
  } catch {
    return undefined;
  }
}

function groupBy<T>(values: readonly T[], key: (value: T) => string): ReadonlyMap<string, readonly T[]> {
  const groups = new Map<string, T[]>();
  for (const value of values) {
    const identity = key(value);
    const group = groups.get(identity);
    if (group) group.push(value);
    else groups.set(identity, [value]);
  }
  return groups;
}

function occurrence(tool: string, timestamp: string, turn: number, role = ''): string {
  return JSON.stringify([tool, timestamp, turn, role]);
}

/** Collections lack session identity: only unique invocation matches are safe. */
export function traceBadgeMetadata(
  events: readonly EventFact[],
  skills: readonly SkillFact[],
  tasks: readonly TaskLinkFact[],
): ReadonlyMap<EventFact, TraceBadgeMetadata> {
  const eventKey = (event: EventFact): string => occurrence(event.tool_name?.trim() ?? '', event.timestamp,
    event.turn, event.tool_name?.trim() === 'skill' ? event.role : '');
  const occurrences = groupBy(events.filter((event) => event.event_type === 'tool_call'), eventKey);
  const skillOccurrences = groupBy(skills, (fact) => occurrence('skill', fact.timestamp, fact.turn_index, fact.caller_role));
  const taskOccurrences = groupBy(tasks, (fact) => occurrence('task', fact.timestamp, fact.parent_turn));
  const result = new Map<EventFact, TraceBadgeMetadata>();
  for (const event of events) {
    if (event.event_type !== 'tool_call') continue;
    const tool = event.tool_name?.trim();
    const key = eventKey(event);
    const uniqueEvent = occurrences.get(key)?.length === 1;
    if (tool === 'skill') {
      const name = inputString(event.tool_input, 'name');
      const matching = skillOccurrences.get(key) ?? [];
      const resolved = name ?? (uniqueEvent && matching.length === 1 ? matching[0]?.name.trim() : undefined);
      if (resolved) result.set(event, { kind: 'skill', name: resolved });
    } else if (tool === 'task') {
      const matching = taskOccurrences.get(key) ?? [];
      const fact = uniqueEvent && matching.length === 1 ? matching[0] : undefined;
      const inputRole = inputString(event.tool_input, 'subagent_type');
      const name = inputRole ?? fact?.subagent_type.trim();
      const childSessionId = !inputRole || inputRole === fact?.subagent_type.trim() ? fact?.child_session_id : undefined;
      if (name || childSessionId) result.set(event, { kind: 'task', name: name || 'unknown', childSessionId });
    }
  }
  return result;
}
