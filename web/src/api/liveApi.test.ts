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
import nativeLiveTrace from '../test/fixtures/live-trace-wire.json';

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
    const root = mockLiveSessionTree.root_sessions[0]!;
    const alert = mockLiveAnomalyAlerts[0]!;

    it.each<{ label: string; payload: unknown }>([
      { label: 'null body', payload: null },
      { label: 'primitive body', payload: 42 },
      { label: 'array body', payload: [] },
      { label: 'missing root collection', payload: {} },
      { label: 'null root collection', payload: {...mockLiveSessionTree, root_sessions: null} },
      { label: 'invalid count', payload: {...mockLiveSessionTree, total_active_sessions: -1} },
      { label: 'invalid root', payload: {...mockLiveSessionTree, root_sessions: [{...root, role: undefined}]} },
      { label: 'invalid descendant', payload: {...mockLiveSessionTree, root_sessions: [{...root, sub_sessions: [null]}]} },
      { label: 'invalid descendant anomaly', payload: {...mockLiveSessionTree, root_sessions: [{
        ...root,
        sub_sessions: [{...root.sub_sessions[0]!, anomalies: [{...alert, trigger_event: 42}]}],
      }]} },
    ])('rejects a tree with $label through REST and SSE', async ({ payload }) => {
      vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(payload))));
      await expect(fetchLiveSessionTree()).rejects.toThrow(/Invalid live/);
      const onSessionTree = vi.fn();
      const onError = vi.fn();
      const unsubscribe = subscribeToLiveStream({onSessionTree, onError}, 333);
      const stream = MockEventSource.instances[0]!;
      stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
      stream.dispatchEvent('session_tree', JSON.stringify(payload));
      expect(onSessionTree.mock.calls).toEqual([[mockLiveSessionTree]]);
      expect(onError.mock.calls).toEqual([[expect.objectContaining({message: expect.stringMatching(/Invalid live/)})]]);
      unsubscribe();
    });

    it('preserves a successful explicitly empty current tree and its recorded refresh timestamp', () => {
      const empty = {root_sessions: [], total_active_sessions: 0, active_anomalies: [], last_refreshed_at: '2026-11-11T12:00:00Z'};
      expect(parseLiveSessionTree(empty)).toEqual(empty);
    });

    const firstMessage = mockLiveTraceView.messages[0]!;
    const firstError = mockLiveTraceView.recent_errors[0]!;

    it.each<{ label: string; payload: unknown }>([
      { label: 'null body', payload: null },
      { label: 'primitive body', payload: 'trace' },
      { label: 'array body', payload: [] },
      { label: 'missing identity', payload: { ...mockLiveTraceView, session_id: undefined } },
      { label: 'blank identity', payload: { ...mockLiveTraceView, session_id: ' ' } },
      { label: 'non-numeric counter', payload: { ...mockLiveTraceView, total_tool_calls: '12' } },
      { label: 'negative counter', payload: { ...mockLiveTraceView, failed_tool_calls: -1 } },
      { label: 'missing messages', payload: { ...mockLiveTraceView, messages: undefined } },
      { label: 'null message', payload: { ...mockLiveTraceView, messages: [null] } },
      { label: 'missing message role', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, role: undefined }] } },
      { label: 'wrong token scalar', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, input_tokens: '1' }] } },
      { label: 'negative token count', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, reasoning_tokens: -1 }] } },
      { label: 'unsafe token count', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, total_tokens: Number.MAX_SAFE_INTEGER + 1 }] } },
      { label: 'null parts', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, parts: null }] } },
      { label: 'invalid nested part', payload: { ...mockLiveTraceView, messages: [{ ...firstMessage, parts: [{ ...firstError, is_error: 'false' }] }] } },
      { label: 'invalid recent error', payload: { ...mockLiveTraceView, recent_errors: [{ ...firstError, tool_name: 42 }] } },
      { label: 'invalid nested anomaly', payload: { ...mockLiveTraceView, anomalies: [{ ...alert, message: 42 }] } },
    ])('rejects a $label through REST and SSE', async ({ payload }) => {
      vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(payload))));
      await expect(fetchLiveTrace(mockLiveTraceView.session_id)).rejects.toThrow(/Invalid live/);

      const onTraceUpdate = vi.fn();
      const onError = vi.fn();
      const unsubscribe = subscribeToLiveStream({onSessionTree: vi.fn(), onTraceUpdate, onError}, 333);
      const stream = MockEventSource.instances[0]!;
      stream.dispatchEvent('trace_update', JSON.stringify(mockLiveTraceView));
      stream.dispatchEvent('trace_update', JSON.stringify(payload));
      expect(onTraceUpdate.mock.calls).toEqual([[mockLiveTraceView]]);
      expect(onError.mock.calls).toEqual([[expect.objectContaining({message: expect.stringMatching(/Invalid live/)})]]);
      unsubscribe();
    });

    it('accepts the native REST and SSE trace representations of no recent errors', async () => {
      const restTrace = {
        ...nativeLiveTrace,
        recent_errors: [],
      };
      vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(restTrace))));
      const trace = await fetchLiveTrace('wire-root');
      expect(trace).toEqual(restTrace);
      expect(trace.messages[0]?.parts[0]).toMatchObject({id: 'wire-part', type: 'text', output_snippet: 'Native trace payload'});
      expect(trace.messages[0]?.parts[0]).not.toHaveProperty('tool_name');
      const onTraceUpdate = vi.fn();
      const onError = vi.fn();
      const unsubscribe = subscribeToLiveStream({onSessionTree: vi.fn(), onTraceUpdate, onError}, 333);
      MockEventSource.instances[0]!.dispatchEvent('trace_update', JSON.stringify(nativeLiveTrace));
      expect(onTraceUpdate.mock.calls).toEqual([[trace]]);
      expect(onError).not.toHaveBeenCalled();
      unsubscribe();
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

  it.each(['success', 'failure'] as const)('serializes polls and ignores pending %s after teardown', async (outcome) => {
    vi.stubGlobal('EventSource', undefined);
    const first = pendingResponse();
    const closing = pendingResponse();
    const fetch = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(closing.promise);
    vi.stubGlobal('fetch', fetch);
    const onSessionTree = vi.fn();
    const onStatusChange = vi.fn();
    const onError = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange, onError}, 333);
    await vi.advanceTimersByTimeAsync(999);
    expect(fetch).toHaveBeenCalledTimes(1);

    first.resolve(new Response(JSON.stringify(mockLiveSessionTree)));
    await vi.advanceTimersByTimeAsync(332);
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(onSessionTree.mock.calls).toEqual([[mockLiveSessionTree]]);
    await vi.advanceTimersByTimeAsync(1);
    expect(fetch).toHaveBeenCalledTimes(2);

    unsubscribe();
    const counts = [onSessionTree.mock.calls.length, onStatusChange.mock.calls.length, onError.mock.calls.length];
    if (outcome === 'success') closing.resolve(new Response(JSON.stringify(mockLiveSessionTree)));
    else closing.reject(new Error('Pending poll failed'));
    await vi.advanceTimersByTimeAsync(999);
    expect([onSessionTree.mock.calls.length, onStatusChange.mock.calls.length, onError.mock.calls.length]).toEqual(counts);
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it('stops a poll before reporting an error status when its error callback closes the subscription', async () => {
    vi.stubGlobal('EventSource', undefined);
    const pending = pendingResponse();
    const fetch = vi.fn().mockReturnValue(pending.promise);
    vi.stubGlobal('fetch', fetch);
    const onStatusChange = vi.fn();
    const onSessionTree = vi.fn();
    const onError = vi.fn();
    const unsubscribe = subscribeToLiveStream({onSessionTree, onStatusChange, onError}, 333);
    onError.mockImplementation(unsubscribe);
    pending.reject(new Error('Polling failed'));
    await vi.advanceTimersByTimeAsync(999);
    expect(onError).toHaveBeenLastCalledWith(expect.objectContaining({message: 'Polling failed'}));
    expect(onStatusChange.mock.calls).toEqual([['fallback-polling']]);
    expect(onSessionTree).not.toHaveBeenCalled();
    expect(fetch).toHaveBeenCalledTimes(1);
  });

  describe('subscribeToLiveStream in Live SSE Mode', () => {
    it('initializes EventSource with /api/live/stream and routes SSE events to callbacks', () => {
      const onSessionTree = vi.fn();
      const onAnomalyAlert = vi.fn();
      const onTraceUpdate = vi.fn();
      const onStatusChange = vi.fn();
      const onError = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onAnomalyAlert,
        onTraceUpdate,
        onStatusChange,
        onError,
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
      es.dispatchEvent('anomaly_alert', JSON.stringify({...alert, count: '3'}));
      expect(onAnomalyAlert.mock.calls).toEqual([[alert]]);
      expect(onError.mock.calls).toEqual([[expect.objectContaining({message: 'Invalid live anomaly count'})]]);

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

    it('transitions to error at the fallback failure threshold and recovers on a valid poll', async () => {
      const onStatusChange = vi.fn();
      const onError = vi.fn();
      const mockFetch = vi.fn()
        .mockRejectedValueOnce(new TypeError('First poll failed'))
        .mockRejectedValueOnce(new TypeError('Second poll failed'))
        .mockResolvedValueOnce(new Response(JSON.stringify(mockLiveSessionTree)));
      vi.stubGlobal('fetch', mockFetch);
      const onSessionTree = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
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
      expect(onError).toHaveBeenLastCalledWith(expect.objectContaining({message: 'First poll failed'}));
      expect(onStatusChange).toHaveBeenLastCalledWith('fallback-polling');

      await vi.advanceTimersByTimeAsync(2000);
      expect(onError).toHaveBeenLastCalledWith(expect.objectContaining({message: 'Second poll failed'}));
      expect(onStatusChange).toHaveBeenLastCalledWith('error');
      expect(onSessionTree).not.toHaveBeenCalled();

      await vi.advanceTimersByTimeAsync(2000);
      expect(onSessionTree.mock.calls).toEqual([[mockLiveSessionTree]]);
      expect(onStatusChange).toHaveBeenLastCalledWith('fallback-polling');

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
      const mockFetch = vi.fn().mockResolvedValue(new Response('', {status: 400}));
      vi.stubGlobal('fetch', mockFetch);

      const maliciousId = '../../etc/passwd?param=1&inject=true';
      await expect(fetchLiveTrace(maliciousId)).rejects.toThrow(/HTTP 400/);

      const encoded = encodeURIComponent(maliciousId);
      expect(mockFetch).toHaveBeenCalledWith(`/api/live/sessions/${encoded}/trace`);
    });

    it('rejects a well-formed trace returned for the wrong session', async () => {
      vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(mockLiveTraceView))));
      await expect(fetchLiveTrace('another-session')).rejects.toThrow(/session does not match selection/i);
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

    it('preserves the selected trace read failure from the server envelope', async () => {
      const response = new Response(JSON.stringify({
        error: 'database_error', message: 'Reading session trace failed', status_code: 500,
      }), {status: 500});
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue(response));

      await expect(fetchLiveTrace('ses-error-500')).rejects.toThrow(/HTTP 500.*Reading session trace failed/i);
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
  it('keeps a partial relationship diagnosis alongside usable source messages', () => {
    const diagnostics = [{code: 'child_relationship_lookup_failed', message: 'child query unavailable'}];
    const trace = parseLiveTraceView({...nativeLiveTrace, relationships: {state: 'unavailable'}, diagnostics});
    expect(trace.relationships?.state).toBe('unavailable');
    expect(trace.diagnostics).toEqual(diagnostics);
    expect(trace.messages).toEqual(nativeLiveTrace.messages);
  });

  it('distinguishes proven complete relationship lookup from omitted support metadata', () => {
    expect(parseLiveTraceView(nativeLiveTrace).relationships?.state).toBe('complete');
    const trace = parseLiveTraceView({...nativeLiveTrace, relationships: undefined, diagnostics: undefined});
    expect(trace).not.toHaveProperty('relationships');
    expect(trace).not.toHaveProperty('diagnostics');
  });

  it('rejects malformed supplied completeness instead of claiming a complete lookup', () => {
    expect(() => parseLiveTraceView({...nativeLiveTrace, relationships: {state: 'garbage'}})).toThrow(/relationship completeness/);
    expect(() => parseLiveTraceView({...nativeLiveTrace, diagnostics: [{code: 'child_relationship_lookup_failed', message: 17}]})).toThrow(/trace diagnostic/);
  });
});

describe('live rate alert structured fields', () => {
 it('preserves exact structured identities and raw evidence without turning totals into streaks', () => {
  const message='  7 separated failures\n';const alert=parseLiveAnomalyAlert({session_id:'parent',rule_name:'ToolHighFailureRate',severity:'WARNING',message,trigger_event:'part-1',timestamp:'2026-11-08T10:00:00Z',tool_name:'read',part_id:'part-1',count:7});
  expect(alert?.rule_name).toBe('ToolHighFailureRate');expect(alert?.tool_name).toBe('read');expect(alert?.part_id).toBe('part-1');expect(alert?.count).toBe(7);expect(alert?.message).toBe(message);
 });
});
