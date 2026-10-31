import type { LiveTraceView } from './live';

export function isActiveLiveSession(status: string | undefined): boolean {
  return status === 'running' || status === 'active';
}

// Compare the typed snapshot contents, including edits within an existing turn.
export function liveTracesEqual(previous: LiveTraceView, next: LiveTraceView): boolean {
  return JSON.stringify(previous) === JSON.stringify(next);
}
