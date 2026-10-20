import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within, act } from '@testing-library/react';
import { LiveObservability } from './LiveObservability';
import { mockLiveSessionTree, mockLiveTraceView } from '../api/mockData';

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
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockLiveTraceView,
    }));
  });

  afterEach(() => {
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

  it('toggles mock mode on and off invoking onMockChange callback', async () => {
    const onSelectTicket = vi.fn();
    const onMockChange = vi.fn();

    render(<LiveObservability onSelectTicket={onSelectTicket} onMockChange={onMockChange} />);

    const mockBtn = screen.getByRole('button', { name: /mock mode: off/i });
    fireEvent.click(mockBtn);

    expect(onMockChange).toHaveBeenCalledWith(true);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /mock mode: on/i })).toBeInTheDocument();
    });

    // Clicking again disables mock mode
    fireEvent.click(screen.getByRole('button', { name: /mock mode: on/i }));
    expect(onMockChange).toHaveBeenCalledWith(false);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /mock mode: off/i })).toBeInTheDocument();
    });
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
      expect(screen.getByTestId('live-session-tree-empty')).toBeInTheDocument();
    });
    expect(screen.getByText('No Active Sessions')).toBeInTheDocument();
  });
});
