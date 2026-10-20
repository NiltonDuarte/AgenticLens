import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  subscribeToLiveStream,
  fetchLiveSessionTree,
  fetchLiveTrace,
  parseLiveSessionTree,
  parseLiveTraceView,
  type LiveStreamCallbacks,
} from './liveApi';
import type {
  LiveAnomalyAlert,
  LiveTraceView,
} from '../domain/live';
import { mockLiveSessionTree, mockLiveAnomalyAlerts, mockLiveTraceView } from './mockData';

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
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  describe('Boundary Parsers', () => {
    it('parseLiveSessionTree safely parses null or non-object input', () => {
      const parsed = parseLiveSessionTree(null);
      expect(parsed.root_sessions).toEqual([]);
      expect(parsed.total_active_sessions).toBe(0);
      expect(parsed.active_anomalies).toEqual([]);
    });

    it('parseLiveSessionTree handles corrupted primitive types (numbers, booleans, strings)', () => {
      for (const input of [42, true, false, 'corrupted_string', undefined]) {
        const parsed = parseLiveSessionTree(input);
        expect(parsed.root_sessions).toEqual([]);
        expect(parsed.total_active_sessions).toBe(0);
        expect(parsed.active_anomalies).toEqual([]);
        expect(typeof parsed.last_refreshed_at).toBe('string');
      }
    });

    it('parseLiveSessionTree handles empty object and corrupted properties safely', () => {
      const parsed = parseLiveSessionTree({
        root_sessions: 'not-an-array' as unknown,
        total_active_sessions: 'not-a-number' as unknown,
        active_anomalies: null as unknown,
        last_refreshed_at: '' as unknown,
      });

      expect(parsed.root_sessions).toEqual([]);
      expect(parsed.total_active_sessions).toBe(0);
      expect(parsed.active_anomalies).toEqual([]);
      expect(Date.parse(parsed.last_refreshed_at)).not.toBeNaN();
    });

    it('parseLiveSessionTree converts raw array into valid LiveSessionTree', () => {
      const parsed = parseLiveSessionTree(mockLiveSessionTree.root_sessions);
      expect(parsed.root_sessions).toHaveLength(mockLiveSessionTree.root_sessions.length);
      expect(parsed.total_active_sessions).toBe(mockLiveSessionTree.root_sessions.length);
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

  describe('subscribeToLiveStream in Mock Mode', () => {
    it('emits mock tree and mock anomalies synchronously without creating EventSource', () => {
      const onSessionTree = vi.fn();
      const onAnomalyAlert = vi.fn();
      const onStatusChange = vi.fn();

      const callbacks: LiveStreamCallbacks = {
        onSessionTree,
        onAnomalyAlert,
        onStatusChange,
      };

      const unsubscribe = subscribeToLiveStream(callbacks, true);
      expect(MockEventSource.instances).toHaveLength(0);
      expect(onStatusChange).toHaveBeenCalledWith('connected');
      expect(onSessionTree).toHaveBeenCalledWith(mockLiveSessionTree);
      expect(onAnomalyAlert).toHaveBeenCalledTimes(mockLiveAnomalyAlerts.length);

      // Fast-forward interval timer to verify periodic tick
      vi.advanceTimersByTime(2000);
      expect(onSessionTree).toHaveBeenCalledTimes(2);

      unsubscribe();
      vi.advanceTimersByTime(4000);
      expect(onSessionTree).toHaveBeenCalledTimes(2);
    });
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);
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
      const trace: LiveTraceView = mockLiveTraceView;
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);
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

      subscribeToLiveStream(callbacks, false);
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);

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

      // Poll 2: 2000ms later fetch fails, but fetchLiveSessionTree falls back to mock tree
      await vi.advanceTimersByTimeAsync(2000);
      expect(onSessionTree).toHaveBeenCalledTimes(2);

      // Poll 3: 2000ms later succeeds again
      await vi.advanceTimersByTimeAsync(2000);
      expect(onSessionTree).toHaveBeenCalledTimes(3);

      // Unsubscribe and verify polling halts
      unsubscribe();
      await vi.advanceTimersByTimeAsync(4000);
      expect(onSessionTree).toHaveBeenCalledTimes(3);
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

      const unsubscribe = subscribeToLiveStream(callbacks, false);

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

      subscribeToLiveStream(callbacks, false);
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

      subscribeToLiveStream(callbacks, false);

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

      const unsubscribe = subscribeToLiveStream(callbacks, false);
      expect(onStatusChange).toHaveBeenCalledWith('fallback-polling');

      await vi.advanceTimersByTimeAsync(10);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');

      unsubscribe();
    });
  });

  describe('fetchLiveSessionTree REST helper', () => {
    it('returns mock tree when isMock is true without network call', async () => {
      const mockFetch = vi.fn();
      vi.stubGlobal('fetch', mockFetch);

      const tree = await fetchLiveSessionTree(true);
      expect(tree).toEqual(mockLiveSessionTree);
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('fetches /api/live/sessions and parses response', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveSessionTree,
      });
      vi.stubGlobal('fetch', mockFetch);

      const tree = await fetchLiveSessionTree(false);
      expect(tree.total_active_sessions).toBe(mockLiveSessionTree.total_active_sessions);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions');
    });

    it('falls back to mock data if fetch fails', async () => {
      const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'));
      vi.stubGlobal('fetch', mockFetch);

      const tree = await fetchLiveSessionTree(false);
      expect(tree).toEqual(mockLiveSessionTree);
    });

    it('falls back to mock tree when server returns 500 internal server error', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      });
      vi.stubGlobal('fetch', mockFetch);

      const tree = await fetchLiveSessionTree(false);
      expect(tree).toEqual(mockLiveSessionTree);
    });
  });

  describe('fetchLiveTrace REST helper', () => {
    it('returns mock trace when isMock is true', async () => {
      const mockFetch = vi.fn();
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-al20-coder', true);
      expect(trace.session_id).toBe('ses-al20-coder');
      expect(mockFetch).not.toHaveBeenCalled();
    });

    it('fetches /api/live/sessions/{id}/trace and parses response', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockLiveTraceView,
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-al20-coder', false);
      expect(trace.session_id).toBe('ses-al20-coder');
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions/ses-al20-coder/trace');
    });

    it('encodes path traversal sequences, slashes, and query characters in sessionId', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      });
      vi.stubGlobal('fetch', mockFetch);

      const maliciousId = '../../etc/passwd?param=1&inject=true';
      await fetchLiveTrace(maliciousId, false);

      const encoded = encodeURIComponent(maliciousId);
      expect(mockFetch).toHaveBeenCalledWith(`/api/live/sessions/${encoded}/trace`);
    });

    it('trims leading and trailing whitespace from sessionId', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ ...mockLiveTraceView, session_id: 'ses-trimmed' }),
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('   ses-trimmed   ', false);
      expect(mockFetch).toHaveBeenCalledWith('/api/live/sessions/ses-trimmed/trace');
      expect(trace.session_id).toBe('ses-trimmed');
    });

    it('handles non-ok response gracefully without throwing', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 404,
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-non-existent', false);
      expect(trace.session_id).toBe('ses-non-existent');
    });

    it('handles server 500 status and returns fallback trace without throwing', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-error-500', false);
      expect(trace.session_id).toBe('ses-error-500');
      expect(trace.messages).toBeDefined();
    });

    it('handles network rejection (TypeError / drop) gracefully without throwing', async () => {
      const mockFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-network-drop', false);
      expect(trace.session_id).toBe('ses-network-drop');
      expect(trace.messages).toBeDefined();
    });

    it('handles malformed JSON response body gracefully without throwing', async () => {
      const mockFetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => {
          throw new SyntaxError('Unexpected token < in JSON at position 0');
        },
      });
      vi.stubGlobal('fetch', mockFetch);

      const trace = await fetchLiveTrace('ses-malformed-json', false);
      expect(trace.session_id).toBe('ses-malformed-json');
      expect(trace.messages).toBeDefined();
    });
  });
});
