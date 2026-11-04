import type { EventFact } from './types';

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

/** Badges describe native event payloads; lean stored occurrences cannot prove event identity. */
export function traceBadgeMetadata(events: readonly EventFact[]): ReadonlyMap<EventFact, TraceBadgeMetadata> {
  const result = new Map<EventFact, TraceBadgeMetadata>();
  for (const event of events) {
    const tool = event.tool_name?.trim().split('.').at(-1) ?? '';
    if ((event.event_type === 'tool_call' && ['skill', 'load_skill'].includes(tool)) || event.event_type === 'skill_activation') {
      const name = event.event_type === 'skill_activation' ? event.skill_name : inputString(event.tool_input, 'name');
      if (name) result.set(event, { kind: 'skill', name });
    } else if (event.event_type === 'tool_call' && ['task', 'spawn_agent', 'delegate_task'].includes(tool)) {
      const name = inputString(event.tool_input, 'subagent_type') ?? inputString(event.tool_input, 'agent_type');
      result.set(event, { kind: 'task', name: name ?? 'unknown', childSessionId: event.child_session_id });
    }
  }
  return result;
}
