import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within, act } from '@testing-library/react';
import { LiveObservability } from './LiveObservability';
import { mockLiveSessionTree, mockLiveTraceView } from '../test/fixtures/mockData';

class MockEventSource {
  static instances: MockEventSource[] = [];
  url: string;
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
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

describe('LiveObservability View Integration', () => {
  beforeEach(() => {
    MockEventSource.instances = [];
    vi.stubGlobal('EventSource', MockEventSource);
    vi.stubGlobal('fetch', vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/trace')) {
        return {
          ok: true,
          json: async () => mockLiveTraceView,
        };
      }
      return {
        ok: true,
        json: async () => mockLiveSessionTree,
      };
    }));
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('mounts, initializes stream, and displays status bar and session tree', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    expect(screen.getByTestId('live-observability-view')).toBeInTheDocument();
    expect(screen.getByTestId('live-status-bar')).toBeInTheDocument();
    expect(screen.getByText('Connecting...')).toBeInTheDocument();

    // Trigger open and dispatch session_tree
    expect(MockEventSource.instances.length).toBeGreaterThan(0);
    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('Connected (SSE)')).toBeInTheDocument();
    });

    expect(screen.getByText('Active Sessions:')).toBeInTheDocument();
    expect(
      screen.getByText('AL-20 TypeScript Live Observability Implementation')
    ).toBeInTheDocument();
  });

  it('selects session from tree and updates trace inspector', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(
        screen.getByText('AL-20 TypeScript Live Observability Implementation')
      ).toBeInTheDocument();
    });

    // Initially inspector is placeholder
    expect(screen.getByText('Select a Session to Inspect')).toBeInTheDocument();

    // Click on coder session node
    const coderTitle = screen.getByText('TypeScript Frontend Implementation');
    const coderCard = coderTitle.closest('.live-session-node-card');
    expect(coderCard).not.toBeNull();
    fireEvent.click(coderCard!);

    // Trace inspector should update
    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });

    const inspector = screen.getByTestId('live-trace-inspector');
    expect(within(inspector).getByText('Session: ses-al20-coder')).toBeInTheDocument();
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeInTheDocument();
  });

  it('handles pause and resume controls properly', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('Connected (SSE)')).toBeInTheDocument();
    });

    const pauseBtn = screen.getByRole('button', { name: /pause streaming/i });
    fireEvent.click(pauseBtn);

    // Status changes to Paused
    await waitFor(() => {
      expect(screen.getByText('Paused')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /resume streaming/i })).toBeInTheDocument();

    // Click resume
    fireEvent.click(screen.getByRole('button', { name: /resume streaming/i }));
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /pause streaming/i })).toBeInTheDocument();
    });
  });

  it('displays connection error banner with retry button when stream and fallback polling fail to connect', async () => {
    class FailingEventSource {
      constructor() {
        throw new Error('Connection refused');
      }
    }
    vi.stubGlobal('EventSource', FailingEventSource);
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));

    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByTestId('live-connection-error-banner')).toBeInTheDocument();
    });

    const banner = screen.getByTestId('live-connection-error-banner');
    expect(banner).toHaveAttribute('role', 'alert');
    expect(
      within(banner).getByText(
        'Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running.'
      )
    ).toBeInTheDocument();
    expect(
      within(banner).getByRole('button', { name: /retry connection/i })
    ).toBeInTheDocument();
  });

  it('attempts reconnection and dismisses error banner when retry connection button is clicked', async () => {
    class FailingEventSource {
      constructor() {
        throw new Error('Connection refused');
      }
    }
    vi.stubGlobal('EventSource', FailingEventSource);
    const mockFetch = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
    vi.stubGlobal('fetch', mockFetch);

    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByTestId('live-connection-error-banner')).toBeInTheDocument();
    });

    // Now restore EventSource and fetch to succeed
    vi.stubGlobal('EventSource', MockEventSource);
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => mockLiveSessionTree,
    });

    const retryBtn = screen.getByRole('button', { name: /retry connection/i });
    fireEvent.click(retryBtn);

    // EventSource connects
    await waitFor(() => {
      expect(MockEventSource.instances.length).toBeGreaterThan(0);
    });
    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('Connected (SSE)')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('live-connection-error-banner')).not.toBeInTheDocument();
  });

  it('does not render any mock mode controls in the UI', () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    expect(screen.queryByRole('button', { name: /mock mode/i })).not.toBeInTheDocument();
  });

  it('navigates to ticket autopsy when guessed ticket badge is clicked', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('AL-20')).toBeInTheDocument();
    });

    const al20Badge = screen.getByText('AL-20').closest('.guessed-ticket-badge');
    expect(al20Badge).not.toBeNull();
    fireEvent.click(al20Badge!);

    expect(onSelectTicket).toHaveBeenCalledWith('AL-20');
  });

  it('focuses session in inspector when anomaly alert item in banner is clicked', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByRole('alert', { name: /active live anomalies/i })).toBeInTheDocument();
    });

    // Click the first anomaly in the banner (targeting ses-al20-coder)
    const anomalyItem = screen.getByText(mockLiveSessionTree.active_anomalies[0]!.message);
    fireEvent.click(anomalyItem);

    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });
    const inspector = screen.getByTestId('live-trace-inspector');
    expect(within(inspector).getByText('Session: ses-al20-coder')).toBeInTheDocument();
  });

  it('transitions to Fallback Polling (REST) status when EventSource is unavailable or throws', async () => {
    vi.stubGlobal('EventSource', undefined);
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByText('Fallback Polling (REST)')).toBeInTheDocument();
    });
  });

  it('triggers manual refresh to reload tree and selected trace', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('AL-20 TypeScript Live Observability Implementation')).toBeInTheDocument();
    });

    // Select session
    const rootCard = screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card')!;
    fireEvent.click(rootCard);

    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });

    // Click Manual Refresh
    const refreshBtn = screen.getByRole('button', { name: /manual refresh/i });
    fireEvent.click(refreshBtn);

    // View should remain stable and retain selected session
    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });
  });

  it('clears selected session and returns to placeholder state when inspector close button is clicked', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('AL-20 TypeScript Live Observability Implementation')).toBeInTheDocument();
    });

    // Select session
    fireEvent.click(screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card')!);

    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });

    // Click Close button in inspector
    const closeBtn = screen.getByRole('button', { name: /close inspector/i });
    fireEvent.click(closeBtn);

    // Inspector should revert to placeholder
    await waitFor(() => {
      expect(screen.getByText('Select a Session to Inspect')).toBeInTheDocument();
    });
    expect(screen.queryByTestId('live-trace-inspector')).not.toBeInTheDocument();
  });

  it('dynamically surfaces anomaly alert received via SSE into banner without page reload', async () => {
    const onSelectTicket = vi.fn();
    const cleanTree = {
      ...mockLiveSessionTree,
      active_anomalies: [],
    };

    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(cleanTree));
    });

    await waitFor(() => {
      expect(screen.getByText('Active Anomalies:')).toBeInTheDocument();
    });

    // Banner is not rendered when anomalies is 0
    expect(screen.queryByRole('alert', { name: /active live anomalies/i })).not.toBeInTheDocument();

    // Now dispatch an SSE anomaly_alert
    act(() => {
      es.dispatchEvent(
        'anomaly_alert',
        JSON.stringify({
          session_id: 'ses-dynamically-alerted',
          rule_name: 'ContextCeilingApproach',
          severity: 'CRITICAL',
          message: 'Dynamic runaway token context alert',
          trigger_event: 'evt-dyn-01',
          timestamp: '2026-10-20T12:45:00Z',
        })
      );
    });

    // Banner should immediately appear
    await waitFor(() => {
      expect(screen.getByRole('alert', { name: /active live anomalies/i })).toBeInTheDocument();
    });
    expect(screen.getByText('Dynamic runaway token context alert')).toBeInTheDocument();
  });

  it('updates trace inspector dynamically when trace_update event arrives for inspected session', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });

    await waitFor(() => {
      expect(screen.getByText('TypeScript Frontend Implementation')).toBeInTheDocument();
    });

    // Select coder session
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);

    await waitFor(() => {
      expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    });

    // Dispatch updated trace with new message turn
    const updatedTrace = {
      ...mockLiveTraceView,
      session_id: 'ses-al20-coder',
      messages: [
        ...mockLiveTraceView.messages,
        {
          id: 'turn-dyn-new',
          role: 'assistant' as const,
          content: 'Brand new dynamically updated turn arrived via SSE stream.',
          reasoning_content: null,
          total_tokens: 88,
          input_tokens: 44,
          output_tokens: 44,
          reasoning_tokens: 0,
          context_tokens: 88,
          timestamp: '2026-10-20T12:50:00Z',
          parts: [],
        },
      ],
    };

    act(() => {
      es.dispatchEvent('trace_update', JSON.stringify(updatedTrace));
    });

    await waitFor(() => {
      expect(
        screen.getByText('Brand new dynamically updated turn arrived via SSE stream.')
      ).toBeInTheDocument();
    });
  });

  it('recovers active trace edits through REST while SSE stays connected and stops after completion', async () => {
    vi.useFakeTimers();
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    await act(async () => { await Promise.resolve(); });
    expect(screen.getByText(/Execution Message Stream/)).toBeInTheDocument();
    const next = { ...mockLiveTraceView, messages: mockLiveTraceView.messages.map((message, index) =>
      index === 0 ? { ...message, content: 'Recovered same-turn update', total_tokens: 9999 } : message) };
    const fetcher = vi.mocked(fetch);
    fetcher.mockClear();
    fetcher.mockResolvedValue({ ok: true, json: async () => next } as Response);
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.getByText('Recovered same-turn update')).toBeInTheDocument();
    expect(fetcher).toHaveBeenCalledTimes(1);
    const coder = mockLiveSessionTree.root_sessions[0]!.sub_sessions.find((node) => node.id === 'ses-al20-coder')!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify({
      ...mockLiveSessionTree, root_sessions: [{ ...coder, status: 'completed' }],
    })));
    await act(async () => { await Promise.resolve(); });
    fetcher.mockClear();
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).not.toHaveBeenCalled();
  });

  it('keeps an accepted SSE snapshot when an older REST response finishes later', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    let finish: ((response: Response) => void) | undefined;
    vi.mocked(fetch).mockImplementation(() => new Promise<Response>((resolve) => { finish = resolve; }));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    const pushed = { ...mockLiveTraceView, messages: mockLiveTraceView.messages.map((message, index) =>
      index === 0 ? { ...message, content: 'Newest pushed snapshot' } : message) };
    act(() => es.dispatchEvent('trace_update', JSON.stringify(pushed)));
    expect(screen.getByText('Newest pushed snapshot')).toBeInTheDocument();
    await act(async () => { finish?.({ ok: true, json: async () => mockLiveTraceView } as Response); });
    expect(screen.getByText('Newest pushed snapshot')).toBeInTheDocument();
  });

  it('coalesces refreshes, preserves expanded reasoning while loading, and retains trace errors across tree updates', async () => {
    vi.useFakeTimers();
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    await act(async () => { await Promise.resolve(); });
    const bubble = screen.getAllByTestId('reasoning-bubble').find((element) => within(element).queryByRole('button', { name: /show reasoning/i }))!;
    fireEvent.click(within(bubble).getByRole('button', { name: /show reasoning/i }));
    const turn = bubble.closest('.trace-message-turn');
    let rejectPending: ((error: Error) => void) | undefined;
    const fetcher = vi.mocked(fetch);
    fetcher.mockClear();
    fetcher.mockImplementation(() => new Promise<Response>((_resolve, reject) => { rejectPending = reject; }));
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).toHaveBeenCalledTimes(1);
    expect(bubble.closest('.trace-message-turn')).toBe(turn);
    expect(within(bubble).getByRole('button', { name: /hide reasoning/i })).toBeInTheDocument();
    // The pending triggers become exactly one follow-up after the failed request.
    fetcher.mockRejectedValue(new Error('trace recovery unavailable'));
    await act(async () => { rejectPending?.(new Error('trace recovery unavailable')); });
    expect(fetcher).toHaveBeenCalledTimes(2);
    expect(within(screen.getByTestId('live-status-bar')).getByRole('alert')).toHaveTextContent('trace recovery unavailable');
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    expect(within(screen.getByTestId('live-status-bar')).getByRole('alert')).toHaveTextContent('trace recovery unavailable');
    expect(bubble.closest('.trace-message-turn')).toBe(turn);
    fetcher.mockResolvedValue({ ok: true, json: async () => mockLiveTraceView } as Response);
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(within(screen.getByTestId('live-status-bar')).queryByRole('alert')).not.toBeInTheDocument();
    expect(within(bubble).getByRole('button', { name: /hide reasoning/i })).toBeInTheDocument();
  });

  it('discards obsolete selection and manual responses without hiding the new session', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    const pending = new Map<string, (response: Response) => void>();
    vi.mocked(fetch).mockImplementation((input) => new Promise<Response>((resolve) => { pending.set(String(input), resolve); }));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    fireEvent.click(screen.getByRole('button', { name: /manual refresh/i }));
    fireEvent.click(screen.getByText('Contract and Architecture Plan').closest('.live-session-node-card')!);
    const plannerTrace = { ...mockLiveTraceView, session_id: 'ses-al20-planner', messages: [] };
    await act(async () => { pending.get('/api/live/sessions/ses-al20-planner/trace')?.({ ok: true, json: async () => plannerTrace } as Response); });
    expect(screen.getByText('Session: ses-al20-planner')).toBeInTheDocument();
    await act(async () => {
      pending.get('/api/live/sessions/ses-al20-coder/trace')?.({ ok: true, json: async () => mockLiveTraceView } as Response);
      pending.get('/api/live/sessions')?.({ ok: true, json: async () => ({ ...mockLiveSessionTree, root_sessions: [] }) } as Response);
    });
    expect(screen.getByText('Session: ses-al20-planner')).toBeInTheDocument();
    expect(screen.getByText('No message turns recorded for this session yet.')).toBeInTheDocument();
    expect(within(screen.getByTestId('live-trace-inspector')).queryByText('Session: ses-al20-coder')).not.toBeInTheDocument();
  });

  it('invalidates in-flight work on pause and close, permits paused manual refresh, and reconciles on resume', async () => {
    vi.useFakeTimers();
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    await act(async () => { await Promise.resolve(); });
    let finishOld: ((response: Response) => void) | undefined;
    const fetcher = vi.mocked(fetch);
    fetcher.mockImplementation(() => new Promise<Response>((resolve) => { finishOld = resolve; }));
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    fireEvent.click(screen.getByRole('button', { name: /pause streaming/i }));
    fetcher.mockClear();
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).not.toHaveBeenCalled();
    const stale = { ...mockLiveTraceView, messages: [] };
    await act(async () => { finishOld?.({ ok: true, json: async () => stale } as Response); });
    act(() => es.dispatchEvent('trace_update', JSON.stringify(stale)));
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeInTheDocument();
    fetcher.mockImplementation(async (input) => ({ ok: true, json: async () => String(input).includes('/trace') ? mockLiveTraceView : mockLiveSessionTree } as Response));
    fireEvent.click(screen.getByRole('button', { name: /manual refresh/i }));
    await act(async () => { await Promise.resolve(); });
    expect(fetcher).toHaveBeenCalledTimes(2);
    fetcher.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /resume streaming/i }));
    await act(async () => { await Promise.resolve(); });
    expect(fetcher).toHaveBeenCalledTimes(1);
    fireEvent.click(screen.getByRole('button', { name: /close inspector/i }));
    fetcher.mockClear();
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).not.toHaveBeenCalled();
  });

  it.each(['stalled', 'missing'])('stops automatic trace requests when the selected node becomes %s', async (status) => {
    vi.useFakeTimers();
    const { unmount } = render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    await act(async () => { await Promise.resolve(); });
    const coder = mockLiveSessionTree.root_sessions[0]!.sub_sessions.find((node) => node.id === 'ses-al20-coder')!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify({ ...mockLiveSessionTree, root_sessions: status === 'missing' ? [] : [{ ...coder, status }] })));
    await act(async () => { await Promise.resolve(); });
    const fetcher = vi.mocked(fetch);
    fetcher.mockClear();
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).not.toHaveBeenCalled();
    unmount();
    await act(async () => { await vi.advanceTimersByTimeAsync(6000); });
    expect(fetcher).not.toHaveBeenCalled();
  });

  it('continues active trace recovery when EventSource is unavailable', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await act(async () => { await Promise.resolve(); });
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation').closest('.live-session-node-card')!);
    await act(async () => { await Promise.resolve(); });
    expect(screen.getByText('Fallback Polling (REST)')).toBeInTheDocument();
    const recovered = { ...mockLiveTraceView, messages: [] };
    vi.mocked(fetch).mockImplementation(async (input) => ({ ok: true, json: async () => String(input).includes('/trace') ? recovered : mockLiveSessionTree } as Response));
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.getByText('No message turns recorded for this session yet.')).toBeInTheDocument();
  });

  it('renders empty tree state gracefully when SSE sends zero active sessions', async () => {
    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    const es = MockEventSource.instances[0]!;
    act(() => {
      es.onopen?.();
      es.dispatchEvent(
        'session_tree',
        JSON.stringify({
          root_sessions: [],
          total_active_sessions: 0,
          active_anomalies: [],
          last_refreshed_at: '2026-10-20T12:00:00Z',
        })
      );
    });

    await waitFor(() => {
      expect(screen.getByText('Connected (SSE)')).toBeInTheDocument();
      expect(screen.getByTestId('live-session-tree-empty')).toBeInTheDocument();
    });
    expect(screen.getByText('No Active Sessions')).toBeInTheDocument();
    expect(screen.queryByTestId('live-connection-error-banner')).not.toBeInTheDocument();
  });

  it('renders empty tree state when fallback polling connects and backend returns zero active sessions', async () => {
    vi.stubGlobal('EventSource', undefined);
    const mockFetch = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/api/live/sessions')) {
        return {
          ok: true,
          json: async () => ({
            root_sessions: [],
            total_active_sessions: 0,
            active_anomalies: [],
            last_refreshed_at: '2026-10-20T12:00:00Z',
          }),
        };
      }
      return {
        ok: true,
        json: async () => mockLiveTraceView,
      };
    });
    vi.stubGlobal('fetch', mockFetch);

    const onSelectTicket = vi.fn();
    render(<LiveObservability onSelectTicket={onSelectTicket} />);

    await waitFor(() => {
      expect(screen.getByText('Fallback Polling (REST)')).toBeInTheDocument();
      expect(screen.getByTestId('live-session-tree-empty')).toBeInTheDocument();
    });
    expect(screen.getByText('No Active Sessions')).toBeInTheDocument();
    expect(screen.queryByTestId('live-connection-error-banner')).not.toBeInTheDocument();
  });
});
