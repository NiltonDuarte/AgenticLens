import type { DurationEvidence, DurationSummary, InvocationFact } from './types';

export function emptyDurationSummary(): DurationSummary {
  return { provisional_seconds: 0, provisional_invocations: 0, unknown_invocations: 0, cutoffs: [] };
}
export function invocationDuration(inv: InvocationFact): DurationSummary {
  const e = inv.duration_evidence;
  return e === undefined
    ? { ...emptyDurationSummary(), unknown_invocations: 1 }
    : { provisional_seconds: e.provisional_seconds, provisional_invocations: e.cutoff ? 1 : 0, unknown_invocations: e.state === 'partial' ? 1 : 0, cutoffs: e.cutoff ? [e.cutoff] : [] };
}
export function mergeDuration(a: DurationSummary, b: DurationSummary): DurationSummary {
  return { provisional_seconds: a.provisional_seconds + b.provisional_seconds, provisional_invocations: a.provisional_invocations + b.provisional_invocations, unknown_invocations: a.unknown_invocations + b.unknown_invocations, cutoffs: [...new Set([...a.cutoffs, ...b.cutoffs])].sort() };
}
export function durationDescription(s: DurationSummary | undefined): string {
  if (!s) return 'Duration coverage unknown';
  const parts: string[] = [];
  if (s.provisional_invocations > 0) parts.push(`${s.provisional_seconds}s provisional; snapshot cutoffs: ${s.cutoffs.join(', ')}`);
  if (s.unknown_invocations > 0) parts.push(`Coverage unknown or unsupported for ${s.unknown_invocations} invocation(s)`);
  return parts.length ? parts.join('; ') : 'Completed measurements';
}
export function durationLabel(value: string, s: DurationSummary | undefined): string {
  if (!s || s.unknown_invocations > 0) return `${value} (coverage incomplete${s && s.provisional_invocations > 0 ? ', provisional' : ''})`;
  return s.provisional_invocations > 0 ? `${value} (provisional)` : value;
}

export function validDurationEvidence(e: unknown, total: number): e is DurationEvidence {
  if (!e || typeof e !== 'object') return false;
  const v = e as Record<string, unknown>;
  const completed = v.completed_seconds;
  const provisional = v.provisional_seconds;
  if (typeof completed !== 'number' || !Number.isSafeInteger(completed) || completed < 0 || typeof provisional !== 'number' || !Number.isSafeInteger(provisional) || provisional < 0 || completed + provisional !== total || typeof v.basis !== 'string' || !v.basis.trim()) return false;
  const cutoff = v.cutoff;
  const limitation = v.limitation;
  if (cutoff !== undefined && (typeof cutoff !== 'string' || !Number.isFinite(Date.parse(cutoff)))) return false;
  if (provisional > 0 && !cutoff) return false;
  if (v.state === 'complete') return cutoff === undefined && provisional === 0 && limitation === undefined;
  if (v.state === 'provisional') return typeof cutoff === 'string' && limitation === undefined;
  return v.state === 'partial' && typeof limitation === 'string' && limitation.trim().length > 0;
}
