import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within, act } from '../test/render';
import { LiveObservability } from './LiveObservability';
import nativeLiveTree from '../test/fixtures/live-session-wire.json';
import nativeLiveTrace from '../test/fixtures/live-trace-wire.json';
import { ProjectConfigContext } from '../contexts/ProjectConfigContext';
import { mockProjectConfig, mockLiveSessionTree, mockLiveTraceView } from '../test/fixtures/mockData';

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

  it('shows unavailable initial source separately from a successful measured-empty response', async () => {
    class UnavailableStream { constructor() { throw new Error('Live source unavailable'); } }
    vi.stubGlobal('EventSource', UnavailableStream);
    const fetch = vi.fn().mockResolvedValue(new Response(JSON.stringify({error: 'source_unsupported', message: 'OpenCode v2 live observation is unsupported', status_code: 503}), {status: 503}));
    vi.stubGlobal('fetch', fetch);
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await screen.findByText('Live session data unavailable.');
    expect(screen.queryByText('No Active Sessions')).not.toBeInTheDocument();
    expect(screen.getAllByText('Not available')).toHaveLength(2);
    expect(screen.getByText('Sessions unavailable')).toBeVisible();
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('OpenCode v2 live observation is unsupported'));

    const empty = {root_sessions: [], active_anomalies: [], total_active_sessions: 0, last_refreshed_at: '2026-11-11T12:00:00Z'};
    fetch.mockImplementation(async () => new Response(JSON.stringify(empty), {status: 200}));
    fireEvent.click(screen.getByRole('button', {name: 'Manual Refresh'}));
    await screen.findByText('No Active Sessions');
    expect(screen.queryByText('Live session data unavailable.')).not.toBeInTheDocument();
    expect(screen.queryByText('Not available')).not.toBeInTheDocument();
    expect(screen.getByText('0 Root Sessions')).toBeVisible();
    expect(screen.queryByText(/OpenCode v2 live observation is unsupported/)).not.toBeInTheDocument();
  });

  it('retains the last successful live tree when a later connection fails', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    act(() => {
      stream.onopen?.();
      stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });
    await screen.findByText('AL-20 TypeScript Live Observability Implementation');
    act(() => stream.onerror?.());
    expect(screen.getByText('AL-20 TypeScript Live Observability Implementation')).toBeVisible();
    expect(within(screen.getByTestId('live-status-bar')).getByRole('alert')).toHaveTextContent('SSE connection error');
    expect(screen.queryByText('No Active Sessions')).not.toBeInTheDocument();
  });

  it('keeps the accepted trace visible when a later REST trace is malformed', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    act(() => {
      stream.onopen?.();
      stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree));
    });
    fireEvent.click(await screen.findByText('TypeScript Frontend Implementation'));
    const accepted = await screen.findByText('Execution Message Stream (4 turns)');
    const malformed = {...mockLiveTraceView, messages: [null]};
    vi.mocked(fetch).mockImplementation(async (input) => new Response(JSON.stringify(
      String(input).endsWith('/trace') ? malformed : mockLiveSessionTree,
    )));
    fireEvent.click(screen.getByRole('button', {name: 'Manual Refresh'}));
    const status = within(screen.getByTestId('live-status-bar'));
    expect(await status.findByRole('alert')).toHaveTextContent(/invalid live trace/i);
    expect(accepted).toBeVisible();
    expect(screen.queryByText('No message turns recorded for this session yet.')).not.toBeInTheDocument();
  });

  it('keeps the accepted trace and its SSE warning until the same session recovers', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    const send = (name: string, data: unknown) => act(() => stream.dispatchEvent(name, JSON.stringify(data)));
    act(() => stream.onopen?.());
    send('session_tree', mockLiveSessionTree);
    fireEvent.click(await screen.findByText('TypeScript Frontend Implementation'));
    const accepted = await screen.findByText('Execution Message Stream (4 turns)');
    const status = within(screen.getByTestId('live-status-bar'));
    const sessionId = mockLiveTraceView.session_id;
    const sourceDiagnostic = 'Native trace read temporarily failed';
    send('source_diagnostic', {error: 'database_error', message: sourceDiagnostic, session_id: sessionId});
    send('trace_update', {...mockLiveTraceView, messages: [null]});
    expect(accepted).toBeVisible();
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live trace message or part');
    expect(status.getByRole('alert')).toHaveTextContent(sourceDiagnostic);

    send('trace_update', {...mockLiveTraceView, session_id: mockLiveSessionTree.root_sessions[0]!.id});
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live trace message or part');
    expect(status.getByRole('alert')).toHaveTextContent(sourceDiagnostic);
    expect(accepted).toBeVisible();

    const recovered = {...mockLiveTraceView.messages[0]!, content: 'Recovered selected trace'};
    send('trace_update', {...mockLiveTraceView, messages: [recovered]});
    expect(screen.getByText(recovered.content)).toBeVisible();
    expect(status.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('keeps a malformed SSE trace warning when an older selected REST refresh resolves', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    const send = (name: string, data: unknown) => act(() => stream.dispatchEvent(name, JSON.stringify(data)));
    act(() => stream.onopen?.());
    send('session_tree', mockLiveSessionTree);
    fireEvent.click(await screen.findByText('TypeScript Frontend Implementation'));
    const accepted = await screen.findByText('Execution Message Stream (4 turns)');
    const status = within(screen.getByTestId('live-status-bar'));

    let finishOlder!: (response: Response) => void;
    const fetcher = vi.mocked(fetch);
    fetcher.mockImplementation((input) => String(input).endsWith('/trace')
      ? new Promise<Response>((resolve) => { finishOlder = resolve; })
      : Promise.resolve(new Response(JSON.stringify(mockLiveSessionTree))));
    fireEvent.click(screen.getByRole('button', {name: 'Manual Refresh'}));
    expect(fetcher).toHaveBeenCalledWith(`/api/live/sessions/${mockLiveTraceView.session_id}/trace`);
    expect(finishOlder).toBeTypeOf('function');
    send('trace_update', {...mockLiveTraceView, messages: [null]});
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live trace message or part');

    const firstMessage = mockLiveTraceView.messages[0]!;
    const oldRead = {...firstMessage, content: 'Older selected REST refresh'};
    await act(async () => finishOlder(new Response(JSON.stringify({...mockLiveTraceView, messages: [oldRead]}))));
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live trace message or part');
    expect(accepted).toBeVisible();
    expect(screen.queryByText(oldRead.content)).not.toBeInTheDocument();

    const currentRead = {...firstMessage, content: 'Current selected SSE trace'};
    send('trace_update', {...mockLiveTraceView, messages: [currentRead]});
    expect(screen.getByText(currentRead.content)).toBeVisible();
    expect(status.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('keeps separate malformed trace warnings for active sessions and one unattributed warning', () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    const send = (name: string, data: unknown) => act(() => stream.dispatchEvent(name, JSON.stringify(data)));
    act(() => stream.onopen?.());
    send('session_tree', mockLiveSessionTree);
    const status = within(screen.getByTestId('live-status-bar'));
    const sessionA = mockLiveTraceView.session_id;
    const sessionB = mockLiveSessionTree.root_sessions[0]!.id;
    const warningA = `Session ${sessionA}: Invalid live trace message or part`;
    const warningB = `Session ${sessionB}: Invalid live trace recent errors`;
    send('trace_update', {...mockLiveTraceView, session_id: sessionA, messages: [null]});
    send('trace_update', {...mockLiveTraceView, session_id: sessionB, recent_errors: [null]});
    expect(status.getByRole('alert')).toHaveTextContent(warningA);
    expect(status.getByRole('alert')).toHaveTextContent(warningB);

    send('trace_update', {...mockLiveTraceView, session_id: 'untracked-session-one', anomalies: [{...mockLiveTraceView.anomalies[0]!, count: 'invalid'}]});
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live anomaly count');
    send('trace_update', {...mockLiveTraceView, session_id: 'untracked-session-two', relationships: {state: 'invalid'}});
    expect(status.getByRole('alert')).toHaveTextContent('Invalid live trace relationship completeness');
    expect(status.getByRole('alert')).not.toHaveTextContent('Invalid live anomaly count');
    expect(status.getByRole('alert')).not.toHaveTextContent('untracked-session');

    send('trace_update', {...mockLiveTraceView, session_id: sessionB});
    expect(status.getByRole('alert')).toHaveTextContent(warningA);
    expect(status.getByRole('alert')).not.toHaveTextContent(warningB);
    expect(status.getByRole('alert')).not.toHaveTextContent('Invalid live trace relationship completeness');
    send('trace_update', {...mockLiveTraceView, session_id: sessionA});
    expect(status.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('keeps tree and per-session SSE diagnostics until the corresponding read recovers', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    const send = (name: string, data: unknown) => act(() => stream.dispatchEvent(name, JSON.stringify(data)));
    const status = within(screen.getByTestId('live-status-bar'));
    const root = mockLiveSessionTree.root_sessions[0]!;
    const child = root.sub_sessions.find((node) => node.id === mockLiveTraceView.session_id)!;
    send('session_tree', mockLiveSessionTree);
    fireEvent.click(await screen.findByText(child.title));
    await screen.findByText('Execution Message Stream (4 turns)');
    send('source_diagnostic', {error: 'database_error', message: 'Tree query interrupted'});
    send('source_diagnostic', {error: 'database_error', message: 'Trace query interrupted', session_id: child.id});
    expect(status.getByRole('alert')).toHaveTextContent('Tree query interrupted');
    expect(status.getByRole('alert')).toHaveTextContent('Trace query interrupted');
    expect(within(screen.getByTestId(`session-node-${child.id}`)).getByText(child.title)).toBeVisible();
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeVisible();
    send('trace_update', {...mockLiveTraceView, session_id: root.id});
    expect(status.getByRole('alert')).toHaveTextContent('Tree query interrupted');
    expect(status.getByRole('alert')).toHaveTextContent('Trace query interrupted');
    send('trace_update', {...mockLiveTraceView, session_id: child.id});
    expect(status.getByRole('alert')).not.toHaveTextContent('Trace query interrupted');
    expect(status.getByRole('alert')).toHaveTextContent('Tree query interrupted');
    send('session_tree', mockLiveSessionTree);
    expect(status.queryByRole('alert')).not.toBeInTheDocument();
  });

  it.each(['SSE', 'REST fallback'])('recovers an unselected trace diagnosis from the next complete %s tree after resume', async (transport) => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    act(() => stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    let finish!: (response: Response) => void;
    vi.mocked(fetch).mockImplementationOnce(() => new Promise<Response>((resolve) => { finish = resolve; }));
    fireEvent.click(screen.getByRole('button', {name: /manual refresh/i}));
    const message = 'Reading an active session trace failed';
    act(() => stream.dispatchEvent('source_diagnostic', JSON.stringify({
      error: 'database_error', message, session_id: mockLiveTraceView.session_id,
    })));
    await act(async () => finish(new Response(JSON.stringify(mockLiveSessionTree))));
    expect(screen.getByText(message, {exact: false})).toBeVisible();

    fireEvent.click(screen.getByRole('button', {name: /pause streaming/i}));
    expect(screen.getByText(message, {exact: false})).toBeVisible();
    if (transport === 'REST fallback') vi.stubGlobal('EventSource', undefined);
    fireEvent.click(screen.getByRole('button', {name: /resume streaming/i}));
    if (transport === 'SSE') {
      const resumedStream = MockEventSource.instances[1]!;
      act(() => resumedStream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    }
    await waitFor(() => expect(screen.queryByText(message, {exact: false})).not.toBeInTheDocument());
    expect(screen.getByText('Select a Session to Inspect')).toBeVisible();
    expect(vi.mocked(fetch).mock.calls.every(([url]) => url === '/api/live/sessions')).toBe(true);
  });

  it.each([
    { older: 'poll', newerFails: false },
    { older: 'poll', newerFails: true },
    { older: 'manual', newerFails: false },
    { older: 'manual', newerFails: true },
  ])('orders an older $older and newer tree read across manual and polling paths (newer failure: $newerFails)', async ({older, newerFails}) => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await act(async () => { await Promise.resolve(); });
    const complete = new Response(JSON.stringify(mockLiveSessionTree));
    const message = 'Tree read failed while accessing the selected database';
    const failed = new Response(JSON.stringify({error: 'database_error', message, status_code: 500}), {status: 500});
    let finishOlder!: (response: Response) => void;
    vi.mocked(fetch)
      .mockImplementationOnce(() => new Promise<Response>((resolve) => { finishOlder = resolve; }))
      .mockResolvedValueOnce(newerFails ? failed : complete);
    const manual = () => fireEvent.click(screen.getByRole('button', {name: /manual refresh/i}));
    const scheduled = () => act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    if (older === 'poll') {
      await scheduled();
      manual();
    } else {
      manual();
      await scheduled();
    }
    await act(async () => { await Promise.resolve(); });
    expect(vi.mocked(fetch)).toHaveBeenCalledTimes(3);
    if (newerFails) expect(screen.getByText(message, {exact: false})).toBeVisible();
    await act(async () => finishOlder(newerFails ? complete : failed));
    if (newerFails) {
      expect(screen.getByText(message, {exact: false})).toBeVisible();
    } else {
      expect(screen.queryByText(message, {exact: false})).not.toBeInTheDocument();
      expect(screen.getByText('Fallback Polling (REST)')).toBeVisible();
    }
  });

  it('uses an available fallback tree while a newer manual request remains pending', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    const pending: ((response: Response) => void)[] = [];
    vi.mocked(fetch).mockImplementation(() => new Promise<Response>((resolve) => pending.push(resolve)));
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', {name: /manual refresh/i}));
    await act(async () => pending[0]!(new Response(JSON.stringify(mockLiveSessionTree))));
    expect(screen.getByText('TypeScript Frontend Implementation')).toBeVisible();
    const empty = {...mockLiveSessionTree, root_sessions: [], active_anomalies: [], total_active_sessions: 0};
    await act(async () => pending[1]!(new Response(JSON.stringify(empty))));
    expect(screen.getByText('No Active Sessions')).toBeVisible();
  });

  it('preserves observed anomalies when REST fallback fails to read an unselected session', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    const client = vi.mocked(fetch);
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await act(async () => { await Promise.resolve(); });
    const alerts = () => screen.getByRole('alert', {name: /active live anomalies/i});
    const observed = mockLiveSessionTree.active_anomalies[0]!.message;
    expect(alerts()).toHaveTextContent(observed);
    const failure = 'Failed to read session ses-al20-coder: no such table: message';
    client.mockResolvedValueOnce(new Response(JSON.stringify({
      error: 'database_error', message: failure, status_code: 500,
    }), {status: 500}));
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.getByText(failure, {exact: false})).toBeVisible();
    expect(screen.getByText('Live source error')).toBeVisible();
    expect(alerts()).toHaveTextContent(observed);
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.queryByText(failure, {exact: false})).not.toBeInTheDocument();
    expect(screen.getByText('Fallback Polling (REST)')).toBeVisible();
    expect(alerts()).toHaveTextContent(observed);
    expect(client.mock.calls.map(([route]) => route)).toEqual(Array(3).fill('/api/live/sessions'));
  });

  it('reports a malformed HTTP 200 fallback tree and clears its warning on the next valid poll', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    const fetcher = vi.mocked(fetch);
    fetcher
      .mockResolvedValueOnce(new Response(JSON.stringify(mockLiveSessionTree)))
      .mockResolvedValueOnce(new Response('null', {status: 200}));
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    const session = screen.getByText('TypeScript Frontend Implementation');

    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.getByTestId('live-connection-error-banner')).toHaveTextContent('Invalid live session tree');
    expect(screen.getByText('Live data error')).toBeVisible();
    expect(session).toBeVisible();
    expect(screen.queryByText('No Active Sessions')).not.toBeInTheDocument();

    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    expect(screen.getByText('Fallback Polling (REST)')).toBeVisible();
    expect(screen.queryByText(/Invalid live session tree/i)).not.toBeInTheDocument();
    expect(session).toBeVisible();
    expect(fetcher.mock.calls.map(([route]) => route)).toEqual(Array(3).fill('/api/live/sessions'));
  });

  it('renders the shared native SQLite-to-HTTP wire through API validation without projecting roles twice', async () => {
    // The Go producer test verifies these same SQLite-to-wire fixtures.
    vi.stubGlobal('EventSource', undefined);
    const fetch = vi.fn(async (input: string | URL | Request) => {
      const route = String(input);
      const payload: unknown = route === '/api/live/sessions' ? nativeLiveTree
        : route === '/api/live/sessions/wire-root/trace' ? {...nativeLiveTrace, recent_errors: []} : undefined;
      if (payload === undefined) throw new Error(`Unexpected native fixture route: ${route}`);
      return new Response(JSON.stringify(payload), {status: 200, headers: {'Content-Type': 'application/json'}});
    });
    vi.stubGlobal('fetch', fetch);
    render(<ProjectConfigContext.Provider value={{...mockProjectConfig,
      prefixes: [{prefix: 'T-', description: 'Wire fixture'}], case_insensitive: false,
      role_mappings: {builder: 'planner', planner: 'other'}, role_order: ['planner'],
    }}><LiveObservability onSelectTicket={vi.fn()} /></ProjectConfigContext.Provider>);
    const node = await screen.findByTestId('session-node-wire-root');
    expect(within(node).getByText('builder')).toBeVisible();
    expect(within(node).getByText('planner')).toHaveClass('node-role-badge');
    expect(within(node).getByText('T-02 wire projection')).toBeVisible();
    fireEvent.click(within(node).getByText('T-02 wire projection'));
    const inspector = await screen.findByTestId('live-trace-inspector');
    expect(within(inspector).getByText('planner')).toHaveClass('node-role-badge');
    expect(await within(inspector).findByText('Native execution turn')).toBeVisible();
    expect(screen.queryByText('other')).not.toBeInTheDocument();
    expect(fetch).toHaveBeenCalledWith('/api/live/sessions');
    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/live/sessions/wire-root/trace'));
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

  it('opens a delegated child and hides the parent stream until the child trace arrives', async () => {
    const parentId = 'ses-al20-root';
    const childId = 'ses-al20-coder';
    const baseMessage = mockLiveTraceView.messages[0]!;
    const parentTrace = { ...mockLiveTraceView, session_id: parentId, recent_errors: [], failed_tool_calls: 0,
      messages: [{ ...baseMessage, content: 'Parent delegation evidence', parts: [{ id: 'delegation', type: 'tool',
        tool_name: 'task', input_snippet: '', output_snippet: '', status: 'completed', is_error: false,
        severity: 'none', subagent_type: 'coder', child_session_id: childId }] }] };
    let finishChild!: (value: Response) => void;
    const childResponse = new Promise<Response>((resolve) => { finishChild = resolve; });
    vi.stubGlobal('fetch', vi.fn().mockImplementation(async (url: string) => {
      if (url.includes(childId) && url.includes('/trace')) return childResponse;
      return { ok: true, json: async () => url.includes('/trace') ? parentTrace : mockLiveSessionTree };
    }));
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const es = MockEventSource.instances[0]!;
    act(() => es.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    fireEvent.click(screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card')!);
    const jump = await screen.findByRole('button', { name: `Inspect child session ${childId}` });
    expect(screen.getByText('Parent delegation evidence')).toBeInTheDocument();
    fireEvent.click(jump);
    expect(screen.getByText(`Loading execution trace for session ${childId}...`)).toBeInTheDocument();
    expect(screen.queryByText('Parent delegation evidence')).not.toBeInTheDocument();
    expect(screen.queryByText('Execution Message Stream (1 turns)')).not.toBeInTheDocument();
    await act(async () => finishChild({ ok: true, json: async () => mockLiveTraceView } as Response));
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeInTheDocument();
    expect(within(screen.getByTestId('live-trace-inspector')).getByText(`Session: ${childId}`)).toBeInTheDocument();
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
    expect(banner).toHaveTextContent('Failed to fetch');
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
          cache_read_tokens: 0,
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

  it('keeps a newer trace diagnostic when a pending REST read completes', async () => {
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    const stream = MockEventSource.instances[0]!;
    act(() => stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
    let finish!: (response: Response) => void;
    vi.mocked(fetch).mockImplementationOnce(() => new Promise<Response>((resolve) => { finish = resolve; }));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation'));
    const message = 'The source could not read the selected trace';
    act(() => stream.dispatchEvent('source_diagnostic', JSON.stringify({
      error: 'database_error', message, session_id: mockLiveTraceView.session_id,
    })));
    expect(screen.getByText(message, {exact: false})).toBeVisible();
    await act(async () => finish(new Response(JSON.stringify(mockLiveTraceView))));
    expect(screen.getByText(message, {exact: false})).toBeVisible();
    expect(screen.queryByText('Execution Message Stream (4 turns)')).not.toBeInTheDocument();
    act(() => stream.dispatchEvent('trace_update', JSON.stringify(mockLiveTraceView)));
    expect(screen.queryByText(message, {exact: false})).not.toBeInTheDocument();
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeVisible();
  });

  it('retains a newer selected trace source error across an older complete fallback response', async () => {
    vi.useFakeTimers();
    vi.stubGlobal('EventSource', undefined);
    render(<LiveObservability onSelectTicket={vi.fn()} />);
    await act(async () => { await Promise.resolve(); });
    let finishTree!: (response: Response) => void;
    vi.mocked(fetch).mockImplementationOnce(() => new Promise<Response>((resolve) => { finishTree = resolve; }));
    await act(async () => { await vi.advanceTimersByTimeAsync(2000); });
    const message = 'Newer selected trace read failed';
    vi.mocked(fetch).mockResolvedValueOnce(new Response(JSON.stringify({
      error: 'database_error', message, status_code: 500,
    }), {status: 500}));
    fireEvent.click(screen.getByText('TypeScript Frontend Implementation'));
    await act(async () => { await Promise.resolve(); });
    expect(screen.getByText(message, {exact: false})).toBeVisible();
    await act(async () => finishTree(new Response(JSON.stringify(mockLiveSessionTree))));
    expect(screen.getByText(message, {exact: false})).toBeVisible();
  });

  it.each(['failure before tree', 'failure after tree', 'success after tree', 'lookup error after tree'])(
    'reconciles a selected trace read with a newer complete tree: %s', async (ordering) => {
      vi.useFakeTimers();
      render(<LiveObservability onSelectTicket={vi.fn()} />);
      const stream = MockEventSource.instances[0]!;
      const sendTree = () => act(() => stream.dispatchEvent('session_tree', JSON.stringify(mockLiveSessionTree)));
      sendTree();
      let finish!: (response: Response) => void;
      vi.mocked(fetch).mockImplementationOnce(() => new Promise<Response>((resolve) => { finish = resolve; }));
      fireEvent.click(screen.getByText('TypeScript Frontend Implementation'));
      const sourceMessage = 'Temporary database read failure';
      const lookupMessage = 'Requested session was not found';
      let outcome: Response;
      if (ordering === 'success after tree') {
        outcome = new Response(JSON.stringify(mockLiveTraceView));
      } else if (ordering === 'lookup error after tree') {
        outcome = new Response(JSON.stringify({error: 'session_not_found', message: lookupMessage, status_code: 404}), {status: 404});
      } else {
        outcome = new Response(JSON.stringify({error: 'database_error', message: sourceMessage, status_code: 500}), {status: 500});
      }
      if (ordering === 'failure before tree') {
        await act(async () => finish(outcome));
        expect(screen.getByText(sourceMessage, {exact: false})).toBeVisible();
        sendTree();
      } else {
        sendTree();
        await act(async () => finish(outcome));
      }
      expect(screen.queryByText(sourceMessage, {exact: false})).not.toBeInTheDocument();
      if (ordering === 'success after tree') expect(screen.getByText('Execution Message Stream (4 turns)')).toBeVisible();
      if (ordering === 'lookup error after tree') expect(screen.getByText(lookupMessage, {exact: false})).toBeVisible();
    },
  );

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
