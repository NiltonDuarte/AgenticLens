import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { fetchLiveTrace, LiveResponseError, type LiveReadOrder } from '../api/liveApi';
import type { LiveSessionNode, LiveTraceView } from '../domain/live';
import { isActiveLiveSession, liveTracesEqual } from '../domain/liveTraceRefresh';

interface TraceOwner {
  readonly sessionId: string;
  inFlight: boolean;
  pending: boolean;
  pushVersion: number;
  healthySourceRead: number;
}

interface TraceFailure {
  readonly message: string;
  readonly sourceRead: number | null;
}

export function useLiveTrace(
  sessionId: string | null, paused: boolean, node: LiveSessionNode | null,
  onTraceRead: (id: string) => void, reads: LiveReadOrder,
) {
  const [trace, setTrace] = useState<LiveTraceView | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<TraceFailure | null>(null);
  const ownerRef = useRef<TraceOwner | null>(null);
  const lastSelection = useRef<string | null>(null);

  const accept = useCallback((next: LiveTraceView) => {
    setTrace((previous) => previous && liveTracesEqual(previous, next) ? previous : next);
    setError(null);
    setLoading(false);
    onTraceRead(next.session_id);
  }, [onTraceRead]);

  const request = useCallback(function refresh(owner: TraceOwner): void {
    if (ownerRef.current !== owner) return;
    if (owner.inFlight) { owner.pending = true; return; }
    owner.inFlight = true;
    const pushVersion = owner.pushVersion;
    const readRevision = reads.begin();
    setLoading(true);
    void fetchLiveTrace(owner.sessionId).then((next) => {
      if (ownerRef.current !== owner || owner.pushVersion !== pushVersion) return;
      if (next.session_id !== owner.sessionId) throw new Error('Trace response session does not match selection');
      accept(next);
    }).catch((cause: unknown) => {
      if (ownerRef.current !== owner || owner.pushVersion !== pushVersion) return;
      const sourceRead = cause instanceof LiveResponseError && cause.isSourceReadFailure ? readRevision : null;
      if (sourceRead !== null && owner.healthySourceRead > sourceRead) return;
      setError({ message: cause instanceof Error ? cause.message : String(cause), sourceRead });
    }).finally(() => {
      owner.inFlight = false;
      if (ownerRef.current !== owner) return;
      setLoading(false);
      if (owner.pending) { owner.pending = false; refresh(owner); }
    });
  }, [accept, reads]);

  useLayoutEffect(() => {
    setTrace((previous) => previous?.session_id === sessionId ? previous : null);
    setError(null);
    setLoading(false);
    const selectionChanged = lastSelection.current !== sessionId;
    lastSelection.current = sessionId;
    if (!sessionId) { ownerRef.current = null; return; }
    const owner: TraceOwner = { sessionId, inFlight: false, pending: false, pushVersion: 0, healthySourceRead: 0 };
    ownerRef.current = owner;
    if (!paused || selectionChanged) request(owner);
    return () => { if (ownerRef.current === owner) ownerRef.current = null; };
  }, [sessionId, paused, request]);

  const active = isActiveLiveSession(node?.status);
  useEffect(() => {
    const owner = ownerRef.current;
    if (!owner || !active || paused) return;
    const timer = window.setInterval(() => request(owner), 2000);
    return () => { window.clearInterval(timer); owner.pending = false; };
  }, [sessionId, paused, active, request]);

  // Revision changes and terminal transitions use the same coalesced recovery path.
  const revision = node ? `${node.id}:${node.time_updated}:${node.status}` : null;
  const previousRevision = useRef<{ id: string | null; revision: string | null; active: boolean }>({ id: null, revision: null, active: false });
  useEffect(() => {
    const previous = previousRevision.current;
    previousRevision.current = { id: sessionId, revision, active };
    const owner = ownerRef.current;
    if (!paused && owner && revision && previous.id === sessionId && previous.revision && previous.revision !== revision && (active || previous.active)) request(owner);
  }, [sessionId, paused, revision, active, request]);

  const onTraceUpdate = useCallback((next: LiveTraceView) => {
    const owner = ownerRef.current;
    if (!owner || next.session_id !== owner.sessionId) return;
    owner.pushVersion++;
    accept(next);
  }, [accept]);

  const onTraceReadFailure = useCallback((id: string) => {
    const owner = ownerRef.current;
    if (owner?.sessionId === id) owner.pushVersion++;
  }, []);

  // A complete tree verifies source reads for its sessions but does not
  // contain the selected trace payload or diagnose a failed HTTP transport.
  const onCompleteTree = useCallback((ids: readonly string[], revision: number) => {
    const owner = ownerRef.current;
    if (!owner || !ids.includes(owner.sessionId)) return;
    owner.healthySourceRead = revision;
    setError((previous) => previous && previous.sourceRead !== null && previous.sourceRead < revision ? null : previous);
  }, []);

  // Capturing ownership before an asynchronous manual tree fetch prevents selection races.
  const captureRefresh = useCallback(() => {
    const owner = ownerRef.current;
    return () => { if (owner) request(owner); };
  }, [request]);

  return { trace: trace?.session_id === sessionId ? trace : null, loading, error: error?.message, onTraceUpdate, onTraceReadFailure, onCompleteTree, captureRefresh };
}
