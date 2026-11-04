import type { EventFact, FactProvenance, SkillFact, TaskLinkFact } from './types';

export type TraceBadgeMetadata =
  | { readonly kind: 'skill'; readonly name: string }
  | { readonly kind: 'task'; readonly name: string; readonly childSessionId?: string | undefined; readonly childReference?: string | undefined; readonly childAgentId?: string | undefined; readonly childLinkDiagnostic?: string | undefined };

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

function sameOccurrence(event: EventFact, fact: { readonly session_id?: string; readonly provenance?: FactProvenance },
  timestamp: string, turn: number): boolean {
  // Session identity is required even when native call IDs happen to coincide.
  if (!fact.session_id || fact.session_id !== event.session_id) return false;
  const provenance = fact.provenance;
  if (provenance?.call_id && event.call_id) return provenance.call_id === event.call_id;
  if (provenance?.event_id) return provenance.event_id === event.event_id;
  if (provenance?.call_id) return false;
  if (provenance?.message_id) return provenance.message_id === event.message_id;
  if (provenance?.item_id) return provenance.item_id === event.message_id;
  return timestamp === event.timestamp && turn === event.turn;
}

function toolName(event: EventFact): string {
  return event.tool_name?.trim().split('.').at(-1) ?? '';
}

/** Match only within a proven owner; ticket-wide legacy indexes cannot establish ownership. */
export function traceBadgeMetadata(
  events: readonly EventFact[],
  skills: readonly SkillFact[],
  tasks: readonly TaskLinkFact[],
): ReadonlyMap<EventFact, TraceBadgeMetadata> {
  const result = new Map<EventFact, TraceBadgeMetadata>();
  for (const event of events) {
    const tool = toolName(event);
    const skillEvent = (event.event_type === 'tool_call' && ['skill', 'load_skill'].includes(tool)) || event.event_type === 'skill_activation';
    if (skillEvent) {
      const matching = skills.filter((fact) => sameOccurrence(event, fact, fact.timestamp, fact.turn_index));
      const fact = matching.length === 1 ? matching[0] : undefined;
      const unique = fact && events.filter((other) => sameOccurrence(other, fact, fact.timestamp, fact.turn_index)).length === 1;
      const name = event.event_type === 'tool_call' ? inputString(event.tool_input, 'name') : undefined;
      const resolved = name ?? (unique ? fact.name.trim() : undefined);
      if (resolved) result.set(event, { kind: 'skill', name: resolved });
    } else if (event.event_type === 'tool_call' && ['task', 'spawn_agent', 'delegate_task'].includes(tool)) {
      const matching = tasks.filter((fact) => sameOccurrence(event, fact, fact.timestamp, fact.parent_turn));
      const candidate = matching.length === 1 ? matching[0] : undefined;
      const fact = candidate && events.filter((other) => sameOccurrence(other, candidate, candidate.timestamp, candidate.parent_turn)).length === 1 ? candidate : undefined;
      const inputRole = inputString(event.tool_input, 'subagent_type') ?? inputString(event.tool_input, 'agent_type');
      const name = inputRole ?? fact?.subagent_type.trim();
      const childSessionId = !inputRole || inputRole === fact?.subagent_type.trim() ? fact?.child_session_id : undefined;
      if (name || childSessionId || fact?.child_agent_id || fact?.child_reference) result.set(event, { kind: 'task', name: name || 'unknown', childSessionId, childReference: fact?.child_reference, childAgentId: fact?.child_agent_id, childLinkDiagnostic: fact?.child_link_diagnostic });
    }
  }
  return result;
}
