import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  subscribeToLiveStream,
  fetchLiveSessionTree,
  fetchLiveTrace,
  parseLiveSessionTree,
  parseLiveTraceView,
  parseLiveAnomalyAlert,
  type LiveStreamCallbacks,
} from './liveApi';
import type {
  LiveAnomalyAlert,
  LiveTraceView,
} from '../domain/live';
import { mockLiveSessionTree, mockLiveAnomalyAlerts, mockLiveTraceView } from '../test/fixtures/mockData';

function pendingResponse() {
  let resolve!: (response: Response) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<Response>((accept, fail) => {resolve = accept; reject = fail;});
  return {promise, resolve, reject};
}

class MockEventSource {
  static instances: MockEventSource[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  onmessage: ((event: MessageEvent<string>) => void) | null = null;
  private listeners: Map<string, ((event: MessageEvent<string>) => void)[]> = new Map();
  closed = false;

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: (event: MessageEvent<string>) => void): void {
    const list = this.listeners.get(type) ?? [];
    list.push(listener);
    this.listeners.set(type, list);
  }

  removeEventListener(type: string, listener: (event: MessageEvent<string>) => void): void {
    const list = this.listeners.get(type) ?? [];
    this.listeners.set(type, list.filter((l) => l !== listener));
  }

  dispatchEvent(type: string, data: string): void {
    const event = { data } as MessageEvent<string>;
    const list = this.listeners.get(type) ?? [];
    for (const l of list) {
      l(event);
    }
  }

  close(): void {
    this.closed = true;
  }
}

describe('liveApi Streaming and Boundary Layer', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    MockEventSource.instances = [];
    vi.stubGlobal('EventSource', MockEventSource);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  describe('Boundary Parsers', () => {
    it.each([
      null, undefined, 42, true, 'corrupted', [], {},
      {...mockLiveSessionTree, root_sessions: null},
      {...mockLiveSessionTree, total_active_sessions: -1},
      {...mockLiveSessionTree, root_sessions: [{...mockLiveSessionTree.root_sessions[0]!, role: undefined}]},
    ])('rejects malformed live tree %j instead of inventing measured zero activity', (input) => {
      expect(() => parseLiveSessionTree(input)).toThrow('Invalid live session tree');
    });

    it('preserves a successful explicitly empty current tree and its recorded refresh timestamp', () => {
      const empty = {root_sessions: [], total_active_sessions: 0, active_anomalies: [], last_refreshed_at: '2026-11-11T12:00:00Z'};
      expect(parseLiveSessionTree(empty)).toEqual(empty);
    });

    it('parseLiveTraceView provides safe defaults on invalid input', () => {
      const parsed = parseLiveTraceView(null, 'fallback-id');
      expect(parsed.session_id).toBe('fallback-id');
      expect(parsed.messages).toEqual([]);
      expect(parsed.total_tool_calls).toBe(0);
      expect(parsed.failed_tool_calls).toBe(0);
    });

    it('parseLiveTraceView handles corrupted inputs and fills missing/non-numeric fields', () => {
      for (const input of [123, true, 'corrupted', undefined]) {
        const parsed = parseLiveTraceView(input, 'fallback-id');
        expect(parsed.session_id).toBe('fallback-id');
        expect(parsed.messages).toEqual([]);
        expect(parsed.total_tool_calls).toBe(0);
        expect(parsed.failed_tool_calls).toBe(0);
        expect(parsed.recent_errors).toEqual([]);
        expect(parsed.anomalies).toEqual([]);
      }

      const corruptedObj = {
        session_id: '',
        messages: 'invalid-messages' as unknown,
        total_tool_calls: 'ten' as unknown,
        failed_tool_calls: null as unknown,
        recent_errors: 'none' as unknown,
        anomalies: 1234 as unknown,
      };

      const parsedObj = parseLiveTraceView(corruptedObj, 'clean-fallback');
      expect(parsedObj.session_id).toBe('clean-fallback');
      expect(parsedObj.messages).toEqual([]);
      expect(parsedObj.total_tool_calls).toBe(0);
      expect(parsedObj.failed_tool_calls).toBe(0);
      expect(parsedObj.recent_errors).toEqual([]);
      expect(parsedObj.anomalies).toEqual([]);
    });

    it('parseLiveTraceView preserves session_id if valid string, otherwise uses fallbackSessionId', () => {
      const valid = parseLiveTraceView({ session_id: 'ses-valid-123' }, 'fallback-id');
      expect(valid.session_id).toBe('ses-valid-123');

      const emptyId = parseLiveTraceView({ session_id: '' }, 'fallback-id');
      expect(emptyId.session_id).toBe('fallback-id');
    });
  });

  it.each([undefined, 0, -1, 1.5, Infinity, 2_147_483_648])('rejects unusable polling cadence %s without starting a substitute timer', (cadence) => {
    vi.stubGlobal('EventSource', undefined);
    const fetch = vi.fn();
    vi.stubGlobal('fetch', fetch);
    const timer = vi.spyOn(globalThis, 'setInterval');
    const onError = vi.fn();
    const onStatusChange = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree: vi.fn(), onError, onStatusChange}, cadence);
    expect(timer).not.toHaveBeenCalled();
    expect(fetch).not.toHaveBeenCalled();
    expect(onError.mock.calls.at(-1)?.[0]).toHaveProperty('message', expect.stringContaining('live.poll_interval_ms'));
    expect(onStatusChange).toHaveBeenLastCalledWith('error');
    unsubscribe();
  });

  it('uses the configured nonpreset cadence, recovers REST status and cancels subsequent polls', async () => {
    vi.stubGlobal('EventSource', undefined);
    let calls = 0;
    const fetch = vi.fn(async () => ++calls === 1 ? new Response('', {status: 503})
      : new Response(JSON.stringify(mockLiveSessionTree), {status: 200}));
    vi.stubGlobal('fetch', fetch);
    const onStatusChange = vi.fn();
    const onSessionTree = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange}, 333);
    await vi.advanceTimersByTimeAsync(0);
    expect(onStatusChange).toHaveBeenLastCalledWith('error');
    await vi.advanceTimersByTimeAsync(332);
    expect(fetch).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect(fetch).toHaveBeenCalledTimes(2);
    expect(onSessionTree).toHaveBeenCalledWith(mockLiveSessionTree);
    expect(onStatusChange).toHaveBeenLastCalledWith('fallback-polling');
    unsubscribe();
    const statusCount = onStatusChange.mock.calls.length;
    await vi.advanceTimersByTimeAsync(999);
    expect(fetch).toHaveBeenCalledTimes(2);
    expect(onStatusChange).toHaveBeenCalledTimes(statusCount);
  });

  it.each(['success', 'failure'] as const)('ignores an older poll %s after a newer success', async (olderResult) => {
    vi.stubGlobal('EventSource', undefined);
    const older = pendingResponse(), newer = pendingResponse();
    const fetch = vi.fn().mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise);
    vi.stubGlobal('fetch', fetch);
    const onSessionTree = vi.fn(), onStatusChange = vi.fn(), onError = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange, onError}, 333);
    await vi.advanceTimersByTimeAsync(333);
    expect(fetch).toHaveBeenCalledTimes(2);
    const current = {...mockLiveSessionTree, last_refreshed_at: '2026-11-11T13:00:02Z'};
    newer.resolve(new Response(JSON.stringify(current), {status: 200}));
    await vi.advanceTimersByTimeAsync(0);
    const statusCount = onStatusChange.mock.calls.length, errorCount = onError.mock.calls.length;
    if (olderResult === 'success') older.resolve(new Response(JSON.stringify({...current, last_refreshed_at: '2026-11-11T13:00:01Z'}), {status: 200}));
    else older.reject(new Error('Older failed request'));
    await vi.advanceTimersByTimeAsync(0);
    expect(onSessionTree.mock.calls).toEqual([[current]]);
    expect(onStatusChange).toHaveBeenCalledTimes(statusCount);
    expect(onError).toHaveBeenCalledTimes(errorCount);
    unsubscribe();
  });

  it('keeps a newer failure authoritative over an older success, then accepts a current recovery', async () => {
    vi.stubGlobal('EventSource', undefined);
    const older = pendingResponse(), newer = pendingResponse();
    const fetch = vi.fn().mockReturnValueOnce(older.promise).mockReturnValueOnce(newer.promise)
      .mockImplementation(async () => new Response(JSON.stringify(mockLiveSessionTree), {status: 200}));
    vi.stubGlobal('fetch', fetch);
    const onSessionTree = vi.fn(), onStatusChange = vi.fn(), onError = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange, onError}, 333);
    await vi.advanceTimersByTimeAsync(333);
    newer.resolve(new Response('', {status: 503}));
    await vi.advanceTimersByTimeAsync(0);
    expect(onStatusChange).toHaveBeenLastCalledWith('error');
    const statusCount = onStatusChange.mock.calls.length;
    older.resolve(new Response(JSON.stringify(mockLiveSessionTree), {status: 200}));
    await vi.advanceTimersByTimeAsync(0);
    expect(onSessionTree).not.toHaveBeenCalled();
    expect(onStatusChange).toHaveBeenCalledTimes(statusCount);
    await vi.advanceTimersByTimeAsync(333);
    expect(onSessionTree).toHaveBeenCalledWith(mockLiveSessionTree);
    expect(onStatusChange).toHaveBeenLastCalledWith('fallback-polling');
    unsubscribe();
  });

  it('accepts useful progress while newer work is pending and suppresses completion after teardown', async () => {
    vi.stubGlobal('EventSource', undefined);
    const useful = pendingResponse(), pending = pendingResponse();
    const fetch = vi.fn().mockReturnValueOnce(useful.promise).mockReturnValueOnce(pending.promise);
    vi.stubGlobal('fetch', fetch);
    const onSessionTree = vi.fn(), onStatusChange = vi.fn(), onError = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange, onError}, 333);
    await vi.advanceTimersByTimeAsync(333);
    useful.resolve(new Response(JSON.stringify(mockLiveSessionTree), {status: 200}));
    await vi.advanceTimersByTimeAsync(0);
    expect(onSessionTree).toHaveBeenCalledWith(mockLiveSessionTree);
    unsubscribe();
    const counts = [onSessionTree.mock.calls.length, onStatusChange.mock.calls.length, onError.mock.calls.length];
    pending.resolve(new Response(JSON.stringify({...mockLiveSessionTree, last_refreshed_at: '2026-11-11T13:00:03Z'}), {status: 200}));
    await vi.advanceTimersByTimeAsync(999);
    expect([onSessionTree.mock.calls.length, onStatusChange.mock.calls.length, onError.mock.calls.length]).toEqual(counts);
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  describe('subscribeToLiveStream in Live SSE Mode', () => {
    it('initializes EventSource with /api/live/stream and routes SSE events to callbacks', () => {
      const onSessionTree = vi.fn();
      const onAnomalyAlert = vi.fn();
      const onTraceUpdate = vi.fn();
      const onStatusChange = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onAnomalyAlert,
        onTraceUpdate,
        onStatusChange,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);
      expect(MockEventSource.instances).toHaveLength(1);
      const es = MockEventSource.instances[0]!;
      expect(es.url).toBe('/api/live/stream');
      expect(onStatusChange).toHaveBeenCalledWith('connecting');

      // Trigger open
      es.onopen?.();
      expect(onStatusChange).toHaveBeenCalledWith('connected');

      // Dispatch session_tree
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
      expect(onSessionTree).toHaveBeenCalledWith(mockLiveSessionTree);

      // Dispatch anomaly_alert
      const alert: LiveAnomalyAlert = mockLiveAnomalyAlerts[0]!;
      es.dispatchEvent('anomaly_alert', JSON.stringify(alert));
      expect(onAnomalyAlert).toHaveBeenCalledWith(alert);

      // Dispatch trace_update
      const trace: LiveTraceView = { ...mockLiveTraceView, relationships:{state:"unavailable"},diagnostics:[{code:"child_relationship_lookup_failed",message:"Native query failed"}] };
      es.dispatchEvent('trace_update', JSON.stringify(trace));
      expect(onTraceUpdate).toHaveBeenCalledWith(trace);

      // Unsubscribe closes EventSource
      unsubscribe();
      expect(es.closed).toBe(true);
    });

    it('routes malformed JSON payloads in SSE events to onError callback without throwing', () => {
      const onError = vi.fn();
      const onSessionTree = vi.fn();
      const onAnomalyAlert = vi.fn();
      const onTraceUpdate = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onAnomalyAlert,
        onTraceUpdate,
        onError,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);
      const es = MockEventSource.instances[0]!;
      es.onopen?.();

      // Malformed session_tree
      es.dispatchEvent('session_tree', '{invalid-json');
      expect(onError).toHaveBeenCalledTimes(1);
      expect(onSessionTree).not.toHaveBeenCalled();

      // Malformed anomaly_alert
      es.dispatchEvent('anomaly_alert', '<not-json>');
      expect(onError).toHaveBeenCalledTimes(2);
      expect(onAnomalyAlert).not.toHaveBeenCalled();

      // Malformed trace_update
      es.dispatchEvent('trace_update', 'undefined');
      expect(onError).toHaveBeenCalledTimes(3);
      expect(onTraceUpdate).not.toHaveBeenCalled();

      unsubscribe();
    });

    it('handles EventSource constructor throwing and initiates fallback polling', async () => {
      class CrashingEventSource {
        constructor() {
          throw new Error('SecurityError: Blocked by Content Security Policy');
        }
      }
      vi.stubGlobal('EventSource', CrashingEventSource);

      const onStatusChange = vi.fn();
      const onError = vi.fn();
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const callbacks: LiveStreamCallbacks = {
        onSessionTree: vi.fn(),
        onStatusChange,
        onError,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);
      expect(onError).toHaveBeenCalledTimes(1);
      expect(onError.mock.calls[0]![0].message).toContain('Blocked by Content Security Policy');
      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      await vi.advanceTimersByTimeAsync(10);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');

      unsubscribe();
    });

    it('resets consecutive error count and reconnect delay upon successful onopen', () => {
      const onStatusChange = vi.fn();
      const callbacks: LiveStreamCallbacks = {
        onSessionTree: vi.fn(),
        onStatusChange,
      };

      subscribeToLiveStream(callbacks, 2000);
      const es1 = MockEventSource.instances[0]!;

      // 1st error
      es1.onerror?.();
      vi.advanceTimersByTime(1000);
      expect(MockEventSource.instances).toHaveLength(2);
      const es2 = MockEventSource.instances[1]!;

      // Connect succeeds on second attempt
      es2.onopen?.();
      expect(onStatusChange).toHaveBeenCalledWith('connected');

      // Subsequent error should start again with initial 1000ms delay, not 2000ms or 4000ms
      es2.onerror?.();
      vi.advanceTimersByTime(999);
      expect(MockEventSource.instances).toHaveLength(2);
      vi.advanceTimersByTime(1);
      expect(MockEventSource.instances).toHaveLength(3);
    });

    it('clears pending reconnect timer and suppresses new connections when unsubscribed', () => {
      const callbacks: LiveStreamCallbacks = {
        onSessionTree: vi.fn(),
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);
      const es1 = MockEventSource.instances[0]!;

      // Trigger error, scheduling reconnect timer
      es1.onerror?.();
      expect(MockEventSource.instances).toHaveLength(1);

      // Unsubscribe before timer fires
      unsubscribe();

      // Advance time well beyond reconnect window
      vi.advanceTimersByTime(10000);
      expect(MockEventSource.instances).toHaveLength(1);
    });

    it('continues periodic fallback polling and remains resilient during network dropouts', async () => {
      const onStatusChange = vi.fn();
      let pollCount = 0;

      const mockFetch = vi.fn().mockImplementation(() => {
        pollCount++;
        if (pollCount === 2) {
          return Promise.reject(new Error('Network offline during poll'));
        }
        return Promise.resolve({
          ok: true,
          json: async () => mockLiveSessionTree,
        });
      });
      vi.stubGlobal('fetch', mockFetch);

      const onSessionTree = vi.fn();
      const onError = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onStatusChange,
        onError,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);

      // Trigger 3 consecutive errors to enter fallback polling
      MockEventSource.instances[0]!.onerror?.();
      vi.advanceTimersByTime(1000);
      MockEventSource.instances[1]!.onerror?.();
      vi.advanceTimersByTime(2000);
      MockEventSource.instances[2]!.onerror?.();

      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      // Poll 1: initial immediate poll succeeds
      await vi.advanceTimersByTimeAsync(10);
      expect(onSessionTree).toHaveBeenCalledTimes(1);

      // Poll 2: 2000ms later fetch fails, error routed to onError
      await vi.advanceTimersByTimeAsync(2000);
      expect(onError).toHaveBeenCalledWith(
        expect.objectContaining({ message: 'Network offline during poll' })
      );
      expect(onSessionTree).toHaveBeenCalledTimes(1);

      // Poll 3: 2000ms later succeeds again
      await vi.advanceTimersByTimeAsync(2000);
      expect(onSessionTree).toHaveBeenCalledTimes(2);

      // Unsubscribe and verify polling halts
      unsubscribe();
      await vi.advanceTimersByTimeAsync(4000);
      expect(onSessionTree).toHaveBeenCalledTimes(2);
    });

    it('transitions to error status on persistent fallback polling network failure', async () => {
      const onStatusChange = vi.fn();
      const onError = vi.fn();
      const mockFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
      vi.stubGlobal('fetch', mockFetch);

      const callbacks: LiveStreamCallbacks = {
        onSessionTree: vi.fn(),
        onStatusChange,
        onError,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);

      // 3 consecutive SSE errors -> transitions to fallback-polling
      MockEventSource.instances[0]!.onerror?.();
      vi.advanceTimersByTime(1000);
      MockEventSource.instances[1]!.onerror?.();
      vi.advanceTimersByTime(2000);
      MockEventSource.instances[2]!.onerror?.();

      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      // 1st poll at t=0 fails
      await vi.advanceTimersByTimeAsync(10);
      expect(onError).toHaveBeenCalled();

      // 2nd poll at t=2000 fails -> consecutive failures >= 2 -> error status
      await vi.advanceTimersByTimeAsync(2000);
      expect(onStatusChange).toHaveBeenCalledWith('error');

      unsubscribe();
    });

    it('routes polling callback exception to onError handler', async () => {
      const onStatusChange = vi.fn();
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const onError = vi.fn();
      const onSessionTree = vi.fn().mockImplementation(() => {
        throw new Error('Consumer handler failed');
      });

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onStatusChange,
        onError,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);

      // Trigger 3 errors to start fallback polling
      MockEventSource.instances[0]!.onerror?.();
      vi.advanceTimersByTime(1000);
      MockEventSource.instances[1]!.onerror?.();
      vi.advanceTimersByTime(2000);
      MockEventSource.instances[2]!.onerror?.();

      await vi.advanceTimersByTimeAsync(10);
      expect(onError).toHaveBeenCalled();
      expect(onError.mock.calls.some((c) => c[0].message.includes('Consumer handler failed'))).toBe(true);

      unsubscribe();
    });

    it('attempts reconnection with backoff on SSE error', () => {
      const onStatusChange = vi.fn();
      const onError = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree: vi.fn(),
        onStatusChange,
        onError,
      };

      subscribeToLiveStream(callbacks, 2000);
      const es1 = MockEventSource.instances[0]!;

      // Simulate error
      es1.onerror?.();
      expect(onError).toHaveBeenCalledTimes(1);
      expect(es1.closed).toBe(true);

      // Fast-forward 1000ms reconnect timer
      vi.advanceTimersByTime(1000);
      expect(MockEventSource.instances).toHaveLength(2);
      const es2 = MockEventSource.instances[1]!;
      expect(es2.url).toBe('/api/live/stream');
    });

    it('transitions to fallback polling after consecutive SSE errors exceed threshold', async () => {
      const onStatusChange = vi.fn();
      const onSessionTree = vi.fn();
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onStatusChange,
      };

      subscribeToLiveStream(callbacks, 2000);

      // 1st error
      MockEventSource.instances[0]!.onerror?.();
      vi.advanceTimersByTime(1000);

      // 2nd error
      MockEventSource.instances[1]!.onerror?.();
      vi.advanceTimersByTime(2000);

      // 3rd error -> triggers fallback polling
      MockEventSource.instances[2]!.onerror?.();
      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      // Verify fetch was invoked
      await vi.advanceTimersByTimeAsync(10);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');
    });

    it('falls back to polling immediately if EventSource is undefined', async () => {
      vi.stubGlobal('EventSource', undefined);
      const onStatusChange = vi.fn();
      const onSessionTree = vi.fn();
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onStatusChange,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, 2000);
      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      await vi.advanceTimersByTimeAsync(10);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');

      unsubscribe();
    });
  });

  describe('fetchLiveSessionTree REST helper', () => {
    it('fetches /api/live/sessions and parses response', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const tree = await fetchLiveSessionTree();
      expect(tree.total_active_sessions).toBe(mockLiveSessionTree.total_active_sessions);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');
    });

    it('rejects with typed Error when fetch rejects with network error', async () => {
      const mockFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveSessionTree()).rejects.toBeInstanceOf(Error);
      await expect(fetchLiveSessionTree()).rejects.toThrow('Failed to fetch');
    });

    it('rejects with typed Error when server returns HTTP 500 error', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveSessionTree()).rejects.toThrow(/HTTP 500/i);
    });

    it('rejects with typed Error when server returns HTTP 404 error', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveSessionTree()).rejects.toThrow(/HTTP 404/i);
    });
  });

  describe('fetchLiveTrace REST helper', () => {
    it('fetches /api/live/sessions/{id}/trace and parses response', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveTraceView,
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-al20-coder');
      expect(trace.session_id).toBe('ses-al20-coder');
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions/ses-al20-coder/trace');
    });

    it('encodes path traversal sequences, slashes, and query characters in sessionId', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ ...mockLiveTraceView, session_id: 'clean' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      const maliciousId = '../../etc/passwd?param=1&inject=true';
      await fetchLiveTrace(maliciousId);

      const encoded = encodeURIComponent(maliciousId);
      expect(mockFetch).toHaveBeenCalledWith(`/api/live/sessions/${encoded}/trace`);
    });

    it('trims leading and trailing whitespace from sessionId', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ ...mockLiveTraceView, session_id: 'ses-trimmed' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('   ses-trimmed   ');
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions/ses-trimmed/trace');
      expect(trace.session_id).toBe('ses-trimmed');
    });

    it('rejects with typed Error mentioning sessionId and HTTP 404 status', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveTrace('ses-missing')).rejects.toThrow(/HTTP 404.*ses-missing/i);
    });

    it('rejects with typed Error on server HTTP 500 error', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveTrace('ses-error-500')).rejects.toThrow(/HTTP 500/i);
    });

    it('rejects with typed Error on network failure', async () => {
      const mockFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveTrace('ses-network-drop')).rejects.toThrow('Failed to fetch');
    });

    it('rejects with typed Error when response body contains malformed JSON', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError('Unexpected token < in JSON at position 0');
        },
      });
      vi.stubGlobal('fetch', mockFetch);

      await expect(fetchLiveTrace('ses-malformed-json')).rejects.toThrow(/Unexpected token/i);
    });
  });
});


describe('live relationship completeness boundary', () => {
 it('preserves partial relationship diagnostics alongside usable messages', () => {
  const trace = parseLiveTraceView({session_id:'parent',messages:[{id:'message',parts:[{id:'part',type:'text',text:'useful content'}]}],relationships:{state:'unavailable'},diagnostics:[{code:'child_relationship_lookup_failed',message:'child query unavailable'}]}, 'parent');
  expect(trace.relationships?.state).toBe('unavailable');
  expect(trace.diagnostics).toEqual([{code:'child_relationship_lookup_failed',message:'child query unavailable'}]);
  expect(trace.messages).toHaveLength(1);
 });
 it('distinguishes proven complete relationship lookup from legacy unknown metadata', () => {
  expect(parseLiveTraceView({session_id:'parent',messages:[],relationships:{state:'complete'},diagnostics:[]}, 'parent').relationships?.state).toBe('complete');
  expect(parseLiveTraceView({session_id:'parent',messages:[]}, 'parent').relationships).toBeUndefined();
 });
});


describe('live rate alert structured fields', () => {
 it('preserves exact structured identities and raw evidence without turning totals into streaks', () => {
  const message='  7 separated failures\n';const alert=parseLiveAnomalyAlert({session_id:'parent',rule_name:'ToolHighFailureRate',severity:'WARNING',message,trigger_event:'part-1',timestamp:'2026-11-08T10:00:00Z',tool_name:'read',part_id:'part-1',count:7});
  expect(alert?.rule_name).toBe('ToolHighFailureRate');expect(alert?.tool_name).toBe('read');expect(alert?.part_id).toBe('part-1');expect(alert?.count).toBe(7);expect(alert?.message).toBe(message);
 });
 it('rejects malformed supplied completeness instead of claiming a complete lookup', () => {
  expect(()=>parseLiveTraceView({session_id:'parent',messages:[],relationships:{state:'garbage'}},'parent')).toThrow();
  expect(()=>parseLiveTraceView({session_id:'parent',messages:[],relationships:{state:'complete'},diagnostics:[{code:'child_relationship_lookup_failed',message:17}]},'parent')).toThrow();
 });
});
