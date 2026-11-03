import type { EventFact, SkillFact, TaskLinkFact } from './types';

export type TraceBadgeMetadata =
  | { readonly kind: 'skill'; readonly name: string }
  | { readonly kind: 'task'; readonly name: string; readonly childSessionId?: string | undefined; readonly childReference?: string | undefined };

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

/** Source IDs bind invocation facts; positions are fallback only within a proven owner. */
export function traceBadgeMetadata(
  events: readonly EventFact[],
  skills: readonly SkillFact[],
  tasks: readonly TaskLinkFact[],
): ReadonlyMap<EventFact, readonly TraceBadgeMetadata[]> {
  const result = new Map<EventFact, readonly TraceBadgeMetadata[]>();
  const matches = (event: EventFact, fact: SkillFact | TaskLinkFact, turn: number, tool: 'skill' | 'task'): boolean => {
    if (!fact.session_id || fact.session_id !== event.session_id) return false;
    if (fact.source_event_id && fact.source_event_id === event.event_id) return true;
    if (fact.source_invocation_id && (fact.source_invocation_id === event.source_invocation_id || fact.source_invocation_id === event.event_id)) return true;
    if (fact.source_event_id || fact.source_invocation_id) return false;
    const positionalMatch = (candidate: EventFact): boolean => candidate.event_type === 'tool_call' &&
      candidate.session_id === fact.session_id && candidate.tool_name?.trim() === tool &&
      fact.timestamp === candidate.timestamp && turn === candidate.turn &&
      (!fact.source_message_id || fact.source_message_id === candidate.message_id);
    return positionalMatch(event) && events.filter(positionalMatch).length === 1;
  };
  for (const event of events) {
    if (event.event_type !== 'tool_call') continue;
    const skillMatches = skills.filter((fact) => matches(event, fact, fact.turn_index, 'skill'));
    const taskMatches = tasks.filter((fact) => matches(event, fact, fact.parent_turn, 'task'));
    const metadata: TraceBadgeMetadata[] = [];
    const provenSkills = skillMatches.length === 1 || skillMatches.every((fact) => fact.source_event_id || fact.source_invocation_id);
    if (provenSkills && skillMatches.length > 0) {
      for (const name of new Set(skillMatches.map((fact) => fact.name))) metadata.push({ kind: 'skill', name });
    } else if (event.tool_name?.trim() === 'skill') {
      const name = inputString(event.tool_input, 'name');
      if (name) metadata.push({ kind: 'skill', name });
    }
    // Conflicting delegation facts cannot choose one particular child link.
    if (taskMatches.length === 1) {
      const fact = taskMatches[0]!;
      const inputRole = inputString(event.tool_input, 'subagent_type');
      metadata.push(!inputRole || inputRole === fact.subagent_type.trim()
        ? { kind: 'task', name: fact.subagent_type || 'unknown', childSessionId: fact.child_session_id, childReference: fact.child_reference }
        : { kind: 'task', name: inputRole });
    } else if (event.tool_name?.trim() === 'task') {
      const name = inputString(event.tool_input, 'subagent_type');
      if (name) metadata.push({ kind: 'task', name });
    }
    if (metadata.length) result.set(event, metadata);
  }
  return result;
}
