import type { DelegationFact, SessionFact, SkillFact } from '../domain/types';

type Assert<Condition extends true> = Condition;
type Equal<Left, Right> = (<Value>() => Value extends Left ? 1 : 2) extends
  (<Value>() => Value extends Right ? 1 : 2) ? true : false;

// Current canonical types cannot silently reacquire retired persistence fields.
export type SkillKeys = Assert<Equal<keyof SkillFact, 'name' | 'caller_role' | 'invoked_at'>>;
export type DelegationKeys = Assert<Equal<keyof DelegationFact, 'agent_type' | 'description' | 'invoked_at' | 'child_session_id'>>;
export type SkillUnknowns = Assert<Equal<SkillFact['caller_role' | 'invoked_at'], string | null>>;
export type DelegationUnknowns = Assert<Equal<DelegationFact['agent_type' | 'invoked_at'], string | null>>;
export type RetiredSessionKeys = Assert<Equal<Extract<keyof SessionFact, 'task_links' | 'events' | 'events_are_partial' | 'agent_path'>, never>>;
