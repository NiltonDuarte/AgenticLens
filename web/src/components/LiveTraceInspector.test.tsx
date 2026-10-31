import { afterEach, describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LiveTraceInspector } from './LiveTraceInspector';
import type { LiveSessionNode, LiveTraceView, LivePartFact } from '../domain/live';
import { mockLiveTraceView } from '../test/fixtures/mockData';

const originalClipboardDescriptor = Object.getOwnPropertyDescriptor(navigator, 'clipboard');

function mockClipboard(writeText: (text: string) => Promise<void>): void {
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText },
  });
}

afterEach(() => {
  if (originalClipboardDescriptor) {
    Object.defineProperty(navigator, 'clipboard', originalClipboardDescriptor);
  } else {
    Reflect.deleteProperty(navigator, 'clipboard');
  }
});

describe('LiveTraceInspector Component', () => {
  const mockNode: LiveSessionNode = {
    id: 'ses-al20-coder',
    parent_id: 'ses-al20-root',
    agent: 'coder',
    role: 'coder',
    model: 'gemini-3.8-flash',
    title: 'TypeScript Frontend Implementation',
    time_created: '2026-10-20T12:15:00Z',
    time_updated: '2026-10-20T12:35:00Z',
    status: 'running',
    ticket_confidence: 0.0,
    anomalies: [],
    sub_sessions: [],
  };

  it('shows rich live invocations, navigates known children, and copies unavailable identifiers', async () => {
    const select = vi.fn();
    const writeText = vi.fn().mockResolvedValue(undefined);
    mockClipboard(writeText);
    const baseMessage = mockLiveTraceView.messages[0];
    if (!baseMessage) throw new Error('Missing trace fixture message');
    const trace: LiveTraceView = { ...mockLiveTraceView, messages: [{ ...baseMessage, parts: [
      { id: 'skill-part', type: 'tool', tool_name: 'skill', skill_name: 'agentic-lens-dev', status: 'completed', is_error: false },
      { id: 'known-task', type: 'tool', tool_name: 'task', subagent_type: 'coder', child_session_id: 'known-child', status: 'completed', is_error: false },
      { id: 'pending-task', type: 'tool', tool_name: 'task', subagent_type: 'tester', child_session_id: 'pending-child', status: 'running', is_error: false },
    ] }] };
    render(<LiveTraceInspector sessionNode={mockNode} trace={trace} loading={false} onClose={vi.fn()}
      availableSessionIds={['known-child']} onSelectSession={select} />);
    expect(screen.getByText('[skill: agentic-lens-dev]')).toHaveAttribute('title', '[skill: agentic-lens-dev]');
    fireEvent.click(screen.getByRole('button', { name: 'Inspect child session known-child' }));
    expect(select).toHaveBeenCalledWith('known-child');
    expect(screen.getByTitle('Child session pending-child is pending or unavailable in this view')).toBeInTheDocument();
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: 'Copy child session ID pending-child' })); });
    expect(writeText).toHaveBeenCalledWith('pending-child');
    const copyChild = screen.getByRole('button', { name: 'Copy child session ID pending-child' });
    expect(copyChild).toHaveTextContent('Copied');
    writeText.mockRejectedValueOnce(new Error('Clipboard denied'));
    await act(async () => { fireEvent.click(copyChild); });
    expect(copyChild).toHaveTextContent('Copy failed');
    expect(screen.getByTitle('Child session pending-child is pending or unavailable in this view')).toHaveTextContent('pending-child');
    expect(select).toHaveBeenCalledTimes(1);
  });

  it('renders placeholder prompt when sessionNode is null', () => {
    render(
      <LiveTraceInspector
        sessionNode={null}
        trace={null}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByTestId('live-trace-inspector-placeholder')).toBeInTheDocument();
    expect(screen.getByText('Select a Session to Inspect')).toBeInTheDocument();
  });

  it('renders loading indicator when loading is true and trace is null', () => {
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={null}
        loading={true}
        onClose={vi.fn()}
      />
    );

    expect(
      screen.getByText('Loading execution trace for session ses-al20-coder...')
    ).toBeInTheDocument();
  });

  it('renders latest-first message stream with role tags and formatted token summaries', () => {
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByTestId('live-trace-inspector')).toBeInTheDocument();
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeInTheDocument();

    expect(screen.getByTestId('message-turn-0')).toBeInTheDocument();
    expect(screen.getByTestId('message-turn-1')).toBeInTheDocument();
    expect(screen.getByTestId('message-turn-2')).toBeInTheDocument();
    expect(screen.getByTestId('message-turn-3')).toBeInTheDocument();
    const messageList = document.querySelector('.inspector-messages-list');
    expect(messageList?.firstElementChild).toHaveAttribute('data-testid', 'message-turn-3');

    expect(
      screen.getByText('Implement Live Observability Dashboard components and streaming client for ticket AL-20.')
    ).toBeInTheDocument();

    const timestamp = screen.getAllByText(/\d{2}:\d{2}:\d{2}\.\d{3}/)[0]!;
    expect(timestamp).toHaveClass('turn-timestamp');
    expect(timestamp).toHaveAttribute('title', expect.stringMatching(/\d{4}/));
    expect(screen.getByText(`TZ: ${Intl.DateTimeFormat().resolvedOptions().timeZone}`)).toHaveClass('inspector-timezone-badge');
  });

  it('preserves malformed message timestamps as the visible time and tooltip fallback', () => {
    const malformedTimestamp = 'not-a-valid-timestamp';
    const trace: LiveTraceView = {
      ...mockLiveTraceView,
      messages: mockLiveTraceView.messages.map((message, index) =>
        index === 0 ? { ...message, timestamp: malformedTimestamp } : message
      ),
    };

    render(<LiveTraceInspector sessionNode={mockNode} trace={trace} loading={false} onClose={vi.fn()} />);

    const timestamp = screen.getByText(malformedTimestamp);
    expect(timestamp).toHaveClass('turn-timestamp');
    expect(timestamp).toHaveAttribute('title', malformedTimestamp);
  });

  it('renders collapsible model reasoning block with expand/collapse toggle', () => {
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const bubbles = screen.getAllByTestId('reasoning-bubble');
    expect(bubbles.length).toBeGreaterThan(0);

    // Find a collapsed long reasoning bubble without depending on display order.
    const collapsedBubble = bubbles.find((bubble) =>
      within(bubble).getByRole('button').textContent?.includes('Show Reasoning'),
    );
    expect(collapsedBubble).toBeDefined();
    const toggleBtn = within(collapsedBubble!).getByRole('button');

    // Click to show reasoning
    fireEvent.click(toggleBtn);
    expect(toggleBtn).toHaveTextContent('Hide Reasoning');

    // Click to hide reasoning again
    fireEvent.click(toggleBtn);
    expect(toggleBtn).toHaveTextContent('Show Reasoning');
  });

  it('shows a prominent collapsed failure summary while keeping the message stream visible', () => {
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const failedSection = screen.getByRole('region', { name: /failed tool calls/i });
    expect(failedSection).toBeInTheDocument();
    expect(screen.getByText('Failed Tool Calls: 3')).toBeInTheDocument();
    const toggle = screen.getByTestId('failed-tools-toggle-btn');
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByTestId('failed-tool-alert-card')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Message Turns')).toBeInTheDocument();
    expect(screen.getByText('Execution Message Stream (4 turns)')).toBeInTheDocument();
  });

  it('expands and collapses failed tool details with the toggle button', () => {
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const toggle = screen.getByTestId('failed-tools-toggle-btn');
    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    const failedCards = screen.getAllByTestId('failed-tool-alert-card');
    expect(failedCards.length).toBe(3);

    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByTestId('failed-tool-alert-card')).not.toBeInTheDocument();
  });

  it('supports Enter and Space keyboard activation on the failure toggle', async () => {
    const user = userEvent.setup();
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const toggle = screen.getByTestId('failed-tools-toggle-btn');
    toggle.focus();
    await user.keyboard('{Enter}');
    expect(toggle).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getAllByTestId('failed-tool-alert-card')).toHaveLength(3);

    await user.keyboard(' ');
    expect(toggle).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByTestId('failed-tool-alert-card')).not.toBeInTheDocument();
  });

  it('invokes onClose when Close button is clicked', () => {
    const onClose = vi.fn();
    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={mockLiveTraceView}
        loading={false}
        onClose={onClose}
      />
    );

    const closeBtn = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('middle-truncates visible message and part IDs while titles and copied prompts retain full IDs', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    const messageId = 'ses_0192a83e-unique-x-304a';
    const partId = 'part_1234567_middle_987654';
    const trace: LiveTraceView = {
      session_id: 'ses-truncation-test',
      messages: [
        {
          id: messageId,
          role: 'assistant',
          content: 'Identifier truncation check',
          timestamp: '2026-10-31T12:00:00Z',
          parts: [
            {
              id: partId,
              type: 'tool',
              tool_name: 'bash',
              status: 'completed',
              is_error: false,
              severity: 'none',
              input_snippet: 'true',
              output_snippet: '',
            },
          ],
          total_tokens: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 0,
        },
      ],
      total_tool_calls: 1,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={trace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const messageToken = screen.getByText('ses_0192...x-304a');
    expect(messageToken).toHaveAttribute('title', messageId);
    const partToken = screen.getByText('part_123...987654');
    expect(partToken).toHaveAttribute('title', partId);

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /copy turn 1 debug prompt/i }));
    });
    expect(writeTextMock.mock.calls[0]?.[0]).toContain(`message \`${messageId}\``);

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: new RegExp(partId) }));
    });
    expect(writeTextMock.mock.calls[1]?.[0]).toContain(`part \`${partId}\``);

    act(() => {
      vi.runOnlyPendingTimers();
    });
    vi.useRealTimers();
  });

  it('renders safely when trace contains zero messages or tool calls', () => {
    const emptyTrace: LiveTraceView = {
      session_id: 'ses-al20-coder',
      messages: [],
      total_tool_calls: 0,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={emptyTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(
      screen.getByText('No message turns recorded for this session yet.')
    ).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: /failed tool calls/i })).not.toBeInTheDocument();
  });

  it('preserves source turn identity across system, user, and assistant turns', () => {
    const chronologicalTrace: LiveTraceView = {
      session_id: 'ses-order-test',
      messages: [
        {
          id: 'turn-0-sys',
          role: 'system',
          content: 'Turn 0: System prompt initialized.',
          reasoning_content: null,
          total_tokens: 100,
          input_tokens: 100,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
          timestamp: '2026-10-20T12:00:00Z',
          parts: [],
        },
        {
          id: 'turn-1-usr',
          role: 'user',
          content: 'Turn 1: User prompt submitted.',
          reasoning_content: null,
          total_tokens: 50,
          input_tokens: 50,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 150,
          timestamp: '2026-10-20T12:00:05Z',
          parts: [],
        },
        {
          id: 'turn-2-ast',
          role: 'assistant',
          content: 'Turn 2: Assistant response generated.',
          reasoning_content: 'Quick thought.',
          total_tokens: 300,
          input_tokens: 0,
          output_tokens: 300,
          cache_read_tokens: 0,
          reasoning_tokens: 50,
          context_tokens: 200,
          timestamp: '2026-10-20T12:00:10Z',
          parts: [],
        },
      ],
      total_tool_calls: 0,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={chronologicalTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const turns = [
      screen.getByTestId('message-turn-0'),
      screen.getByTestId('message-turn-1'),
      screen.getByTestId('message-turn-2'),
    ];

    expect(within(turns[0]!).getByText('SYSTEM')).toBeInTheDocument();
    expect(within(turns[0]!).getByText('Turn 0: System prompt initialized.')).toBeInTheDocument();

    expect(within(turns[1]!).getByText('USER')).toBeInTheDocument();
    expect(within(turns[1]!).getByText('Turn 1: User prompt submitted.')).toBeInTheDocument();

    expect(within(turns[2]!).getByText('ASSISTANT')).toBeInTheDocument();
    expect(within(turns[2]!).getByText('Turn 2: Assistant response generated.')).toBeInTheDocument();
    expect(screen.getByTestId('message-turn-2').querySelector('.turn-number-badge')).toHaveTextContent('Turn 3');
    expect(screen.getByTestId('message-turn-0').querySelector('.turn-number-badge')).toHaveTextContent('Turn 1');
  });

  it('toggles between latest-first and oldest-first without changing logical turn numbers', () => {
    render(
      <LiveTraceInspector sessionNode={mockNode} trace={mockLiveTraceView} loading={false} onClose={vi.fn()} />
    );

    const messageList = document.querySelector('.inspector-messages-list');
    expect(messageList?.firstElementChild).toHaveAttribute('data-testid', 'message-turn-3');
    fireEvent.click(screen.getByTestId('turn-order-toggle-btn'));
    expect(messageList?.firstElementChild).toHaveAttribute('data-testid', 'message-turn-0');
    expect(screen.getByTestId('message-turn-0').querySelector('.turn-number-badge')).toHaveTextContent('Turn 1');
    expect(screen.getByTestId('message-turn-3').querySelector('.turn-number-badge')).toHaveTextContent('Turn 4');
  });

  it('preserves the reading position and offers a jump when new turns arrive while paused', () => {
    const initialTrace: LiveTraceView = { ...mockLiveTraceView, messages: mockLiveTraceView.messages.slice(0, 2) };
    const { rerender } = render(
      <LiveTraceInspector sessionNode={mockNode} trace={initialTrace} loading={false} onClose={vi.fn()} />
    );
    const messageList = document.querySelector('.inspector-messages-list') as HTMLDivElement;
    Object.defineProperties(messageList, {
      scrollHeight: { configurable: true, value: 200 },
      clientHeight: { configurable: true, value: 100 },
    });
    messageList.scrollTop = 80;
    fireEvent.scroll(messageList);
    Object.defineProperty(messageList, 'scrollHeight', { configurable: true, value: 300 });

    rerender(
      <LiveTraceInspector sessionNode={mockNode} trace={mockLiveTraceView} loading={false} onClose={vi.fn()} />
    );

    expect(messageList.scrollTop).toBe(180);
    expect(screen.getByRole('button', { name: /new turns available/i })).toBeInTheDocument();
  });

  it('quick navigation buttons scroll to each end of the message list', () => {
    render(
      <LiveTraceInspector sessionNode={mockNode} trace={mockLiveTraceView} loading={false} onClose={vi.fn()} />
    );
    const messageList = document.querySelector('.inspector-messages-list') as HTMLDivElement;
    Object.defineProperty(messageList, 'scrollHeight', { configurable: true, value: 500 });
    const scrollTo = vi.fn();
    messageList.scrollTo = scrollTo;

    fireEvent.click(screen.getByTestId('scroll-to-top-btn'));
    expect(scrollTo).toHaveBeenLastCalledWith({ top: 0, behavior: 'smooth' });
    fireEvent.click(screen.getByTestId('scroll-to-bottom-btn'));
    expect(scrollTo).toHaveBeenLastCalledWith({ top: 500, behavior: 'smooth' });
  });

  it('handles short model reasoning blocks by showing full deliberation initially with Hide Reasoning button', () => {
    const shortReasoning = 'Short direct plan: inspect and edit files.';
    const shortTrace: LiveTraceView = {
      session_id: 'ses-short-reasoning',
      messages: [
        {
          id: 'msg-short',
          role: 'assistant',
          content: 'I have formulated the plan.',
          reasoning_content: shortReasoning,
          total_tokens: 250,
          input_tokens: 100,
          output_tokens: 150,
          cache_read_tokens: 0,
          reasoning_tokens: 20,
          context_tokens: 200,
          timestamp: '2026-10-20T12:00:00Z',
          parts: [],
        },
      ],
      total_tool_calls: 0,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={shortTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    const bubble = screen.getByTestId('reasoning-bubble');
    expect(bubble).toBeInTheDocument();

    // Short deliberation is not clamped initially
    const toggleBtn = within(bubble).getByRole('button', { name: /hide reasoning/i });
    expect(toggleBtn).toBeInTheDocument();
    expect(screen.getByText(shortReasoning)).toBeInTheDocument();

    // Clicking Hide Reasoning toggles state
    fireEvent.click(toggleBtn);
    expect(within(bubble).getByRole('button', { name: /show reasoning/i })).toBeInTheDocument();
  });

  it('omits reasoning bubble when reasoning_content is empty, null, or whitespace-only', () => {
    const blankReasoningTrace: LiveTraceView = {
      session_id: 'ses-blank-reasoning',
      messages: [
        {
          id: 'msg-blank',
          role: 'assistant',
          content: 'No reasoning included here.',
          reasoning_content: '   ',
          total_tokens: 100,
          input_tokens: 50,
          output_tokens: 50,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
          timestamp: '2026-10-20T12:00:00Z',
          parts: [],
        },
      ],
      total_tool_calls: 0,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={blankReasoningTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(screen.queryByTestId('reasoning-bubble')).not.toBeInTheDocument();
  });

  it('renders selective token chips only for positive token counts and formats large numbers', () => {
    const tokenTrace: LiveTraceView = {
      session_id: 'ses-token-test',
      messages: [
        {
          id: 'msg-zero-tokens',
          role: 'user',
          content: 'Zero tokens turn.',
          reasoning_content: null,
          total_tokens: 0,
          input_tokens: 0,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 0,
          timestamp: '2026-10-20T12:00:00Z',
          parts: [],
        },
        {
          id: 'msg-large-tokens',
          role: 'assistant',
          content: 'Large token turn.',
          reasoning_content: null,
          total_tokens: 1250000,
          input_tokens: 1200000,
          output_tokens: 50000,
          cache_read_tokens: 0,
          reasoning_tokens: 10000,
          context_tokens: 1500000,
          timestamp: '2026-10-20T12:01:00Z',
          parts: [],
        },
      ],
      total_tool_calls: 0,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={tokenTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    // Turn 0 should have zero token chips
    const turn0 = screen.getByTestId('message-turn-0');
    expect(turn0.querySelectorAll('.token-chip').length).toBe(0);

    // Turn 1 should have formatted token chips
    const turn1 = screen.getByTestId('message-turn-1');
    expect(within(turn1).getByText('Total: 1.3M')).toBeInTheDocument();
    expect(within(turn1).getByText('Context: 1.5M')).toBeInTheDocument();
    expect(within(turn1).getByText('Input: 1.2M')).toBeInTheDocument();
    expect(within(turn1).getByText('Output: 50k')).toBeInTheDocument();
    expect(within(turn1).getByText('Reasoning: 10k')).toBeInTheDocument();
  });

  it('renders message parts with correct success/error distinction and snippets', () => {
    const partsTrace: LiveTraceView = {
      session_id: 'ses-parts-test',
      messages: [
        {
          id: 'msg-parts',
          role: 'assistant',
          content: 'Ran tools below:',
          reasoning_content: null,
          total_tokens: 100,
          input_tokens: 50,
          output_tokens: 50,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
          timestamp: '2026-10-20T12:00:00Z',
          parts: [
            {
              id: 'part-ok',
              type: 'tool',
              tool_name: 'read_file',
              status: 'completed',
              is_error: false,
              severity: 'low',
              input_snippet: 'path: config.yaml',
              output_snippet: 'version: 1.0.0',
            },
            {
              id: 'part-fail',
              type: 'tool',
              tool_name: 'bash',
              status: 'failed',
              is_error: true,
              severity: 'high',
              input_snippet: 'cat missing.txt',
              output_snippet: 'cat: missing.txt: No such file or directory',
            },
          ],
        },
      ],
      total_tool_calls: 2,
      failed_tool_calls: 1,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={mockNode}
        trace={partsTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByText('read_file')).toBeInTheDocument();
    expect(screen.getByText('path: config.yaml')).toBeInTheDocument();
    expect(screen.getByText('version: 1.0.0')).toBeInTheDocument();

    expect(screen.getByText('bash')).toBeInTheDocument();
    expect(screen.getByText('cat missing.txt')).toBeInTheDocument();
    expect(screen.getByText('cat: missing.txt: No such file or directory')).toBeInTheDocument();

    const okPart = screen.getByText('path: config.yaml').closest('.trace-part-item');
    expect(okPart).toHaveClass('part-normal');

    const errorPart = screen.getByText('cat missing.txt').closest('.trace-part-item');
    expect(errorPart).toHaveClass('part-error');
  });

  it('renders copy buttons for session ID, message turn ID, and part ID with temporary copied feedback', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    mockClipboard(writeTextMock);

    const debugTrace: LiveTraceView = {
      session_id: 'ses-debug-1',
      messages: [
        {
          id: 'msg-debug-101',
          role: 'assistant',
          content: 'Debugging turn',
          timestamp: '2026-10-20T12:00:00Z',
          parts: [
            {
              id: 'part-debug-202',
              type: 'tool',
              tool_name: 'bash',
              status: 'completed',
              is_error: false,
              severity: 'none',
              input_snippet: 'echo hello',
              output_snippet: 'hello',
            },
          ],
          total_tokens: 100,
          input_tokens: 100,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
        },
      ],
      total_tool_calls: 1,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={{ ...mockNode, id: 'ses-debug-1' }}
        trace={debugTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    // 1. Session Overview Prompt Copy Button
    const sessionCopyBtn = screen.getByRole('button', { name: /copy session overview debug prompt/i });
    expect(sessionCopyBtn).toHaveTextContent('Copy Prompt');
    await act(async () => {
      fireEvent.click(sessionCopyBtn);
    });
    expect(writeTextMock).toHaveBeenCalledTimes(1);
    expect(writeTextMock.mock.calls[0][0]).toContain('# Session Overview: `ses-debug-1`');
    expect(sessionCopyBtn).toHaveTextContent('Copied');
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(sessionCopyBtn).toHaveTextContent('Copy Prompt');

    // 2. Message Turn Prompt Copy Button
    expect(screen.getByText('Turn 1')).toBeInTheDocument();
    const msgCopyBtn = screen.getByRole('button', { name: /copy turn 1 debug prompt/i });
    expect(msgCopyBtn).toHaveTextContent('Copy Prompt');
    await act(async () => {
      fireEvent.click(msgCopyBtn);
    });
    expect(writeTextMock).toHaveBeenCalledTimes(2);
    expect(writeTextMock.mock.calls[1][0]).toContain('# Turn 1 Review (`assistant` / `gemini-3.8-flash`)');
    expect(writeTextMock.mock.calls[1][0]).toContain(
      'In turn 1 of session `ses-debug-1` (message `msg-debug-101`), the agent deliberated and executed tools.'
    );
    expect(msgCopyBtn).toHaveTextContent('Copied');
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(msgCopyBtn).toHaveTextContent('Copy Prompt');

    // 3. Tool Execution Part Prompt Copy Button
    const partCopyBtn = screen.getByRole('button', {
      name: /copy tool execution debug prompt for part part-debug-202/i,
    });
    expect(partCopyBtn).toHaveTextContent('Copy Prompt');
    await act(async () => {
      fireEvent.click(partCopyBtn);
    });
    expect(writeTextMock).toHaveBeenCalledTimes(3);
    expect(writeTextMock.mock.calls[2][0]).toContain('# Tool Execution: `bash`');
    expect(writeTextMock.mock.calls[2][0]).toContain(
      'During turn 1 of session `ses-debug-1` (message `msg-debug-101`, part `part-debug-202`), the agent executed `bash` and it completed successfully.'
    );
    expect(partCopyBtn).toHaveTextContent('Copied');
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(partCopyBtn).toHaveTextContent('Copy Prompt');

    vi.useRealTimers();
  });

  it('omits copy buttons when message ID or part ID is absent', () => {
    const traceWithoutIds: LiveTraceView = {
      session_id: 'ses-no-ids',
      messages: [
        {
          id: '',
          role: 'assistant',
          content: 'No IDs turn',
          timestamp: '2026-10-20T12:00:00Z',
          parts: [
            {
              id: '',
              type: 'tool',
              tool_name: 'bash',
              status: 'completed',
              is_error: false,
              severity: 'none',
            },
          ],
          total_tokens: 50,
          input_tokens: 50,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 50,
        },
      ],
      total_tool_calls: 1,
      failed_tool_calls: 0,
      recent_errors: [],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={{ ...mockNode, id: 'ses-no-ids' }}
        trace={traceWithoutIds}
        loading={false}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByText('Turn 1')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /copy turn 1 debug prompt/i })).toBeNull();
    expect(screen.queryByRole('button', { name: /copy tool execution debug prompt/i })).toBeNull();
  });

  it('does not display streak badge when anomaly count is 1 or missing', () => {
    const errorPart: LivePartFact = {
      id: 'part-fail-single',
      type: 'tool',
      tool_name: 'bash',
      status: 'error',
      is_error: true,
      severity: 'medium',
      input_snippet: 'cat foo.txt',
      output_snippet: 'not found',
    };

    const traceSingleError: LiveTraceView = {
      session_id: 'ses-single',
      messages: [
        {
          id: 'msg-s-1',
          role: 'assistant',
          content: 'Single failure',
          timestamp: '2026-10-20T12:00:00Z',
          parts: [errorPart],
          total_tokens: 100,
          input_tokens: 100,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
        },
      ],
      total_tool_calls: 1,
      failed_tool_calls: 1,
      recent_errors: [errorPart],
      anomalies: [],
    };

    render(
      <LiveTraceInspector
        sessionNode={{ ...mockNode, id: 'ses-single' }}
        trace={traceSingleError}
        loading={false}
        onClose={vi.fn()}
      />
    );

    fireEvent.click(screen.getByTestId('failed-tools-toggle-btn'));
    expect(screen.queryByText(/consecutive failures/i)).toBeNull();
  });

  it('wires exact anomaly streak count to FailedToolAlertCard', () => {
    const errorPart: LivePartFact = {
      id: 'part-fail-streak',
      type: 'tool',
      tool_name: 'bash',
      status: 'error',
      is_error: true,
      severity: 'high',
      input_snippet: 'make test',
      output_snippet: 'exit status 2',
    };

    const traceWithStreak: LiveTraceView = {
      session_id: 'ses-streak',
      messages: [
        {
          id: 'msg-streak-1',
          role: 'assistant',
          content: 'Error',
          timestamp: '2026-10-20T12:00:00Z',
          parts: [errorPart],
          total_tokens: 100,
          input_tokens: 100,
          output_tokens: 0,
          cache_read_tokens: 0,
          reasoning_tokens: 0,
          context_tokens: 100,
        },
      ],
      total_tool_calls: 1,
      failed_tool_calls: 1,
      recent_errors: [errorPart],
      anomalies: [
        {
          session_id: 'ses-streak',
          rule_name: 'ConsecutiveToolErrors',
          severity: 'CRITICAL',
          message: 'Tool execution error streak reached 4 consecutive failures on bash',
          trigger_event: 'part-fail-streak: exit status 2',
          timestamp: '2026-10-20T12:00:00Z',
          count: 4,
        },
      ],
    };

    render(
      <LiveTraceInspector
        sessionNode={{ ...mockNode, id: 'ses-streak' }}
        trace={traceWithStreak}
        loading={false}
        onClose={vi.fn()}
      />
    );

    fireEvent.click(screen.getByTestId('failed-tools-toggle-btn'));
    expect(screen.getByText('4 Consecutive Failures')).toBeInTheDocument();
  });

  it('adversarially tests all 4 layer copy buttons and 2000ms feedback states with streak and error details', async () => {
    vi.useFakeTimers();
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    mockClipboard(writeTextMock);

    const errorPart: LivePartFact = {
      id: 'part-adv-404',
      type: 'tool',
      tool_name: 'bash',
      status: 'error',
      is_error: true,
      severity: 'high',
      input_snippet: 'cargo test --all',
      output_snippet: 'error[E0432]: unresolved import',
      error_message: 'compilation failed',
    };

    const advTrace: LiveTraceView = {
      session_id: 'ses-adv-live',
      messages: [
        {
          id: 'msg-adv-1',
          role: 'assistant',
          content: 'Executing build verification step.',
          timestamp: '2026-10-20T12:00:00Z',
          reasoning_content: 'Let me compile the packages and examine failures.',
          parts: [errorPart],
          total_tokens: 150,
          input_tokens: 100,
          output_tokens: 50,
          cache_read_tokens: 0,
          reasoning_tokens: 20,
          context_tokens: 150,
        },
      ],
      total_tool_calls: 3,
      failed_tool_calls: 3,
      recent_errors: [errorPart],
      anomalies: [
        {
          session_id: 'ses-adv-live',
          rule_name: 'ConsecutiveToolErrors',
          severity: 'CRITICAL',
          message: 'Tool execution error streak reached 5 consecutive failures on bash',
          trigger_event: 'part-adv-404: unresolved import',
          timestamp: '2026-10-20T12:00:00Z',
          count: 5,
        },
      ],
    };

    render(
      <LiveTraceInspector
        sessionNode={{ ...mockNode, id: 'ses-adv-live', role: 'coder', model: 'gemini-3.8-flash', status: 'running' }}
        trace={advTrace}
        loading={false}
        onClose={vi.fn()}
      />
    );

    // 1. Session Overview Copy Button (Layer 4)
    const sessionCopyBtn = screen.getByRole('button', { name: 'Copy session overview debug prompt' });
    expect(sessionCopyBtn).toBeInTheDocument();
    expect(sessionCopyBtn).toHaveTextContent('Copy Prompt');

    await act(async () => {
      fireEvent.click(sessionCopyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(1);
    const sessionPrompt = writeTextMock.mock.calls[0][0];
    expect(sessionPrompt).toContain('# Session Overview: `ses-adv-live`');
    expect(sessionPrompt).toContain('Session `ses-adv-live` is a `coder` subagent running `gemini-3.8-flash`');
    expect(sessionPrompt).toContain('**running** after 1 turns and 3 tool executions (3 failed)');
    expect(sessionPrompt).toContain('### Flagged Issues');
    expect(sessionPrompt).toContain('Consecutive tool failure streak');
    expect(sessionCopyBtn).toHaveTextContent('Copied');

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(sessionCopyBtn).toHaveTextContent('Copy Prompt');

    // 2. Turn Copy Button (Layer 3)
    const turnCopyBtn = screen.getByRole('button', { name: 'Copy turn 1 debug prompt' });
    expect(turnCopyBtn).toBeInTheDocument();
    expect(turnCopyBtn).toHaveTextContent('Copy Prompt');

    await act(async () => {
      fireEvent.click(turnCopyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(2);
    const turnPrompt = writeTextMock.mock.calls[1][0];
    expect(turnPrompt).toContain('# Turn 1 Review (`assistant` / `gemini-3.8-flash`)');
    expect(turnPrompt).toContain('In turn 1 of session `ses-adv-live` (message `msg-adv-1`)');
    expect(turnPrompt).toContain('```\nLet me compile the packages and examine failures.\n```');
    expect(turnPrompt).toContain('```\nExecuting build verification step.\n```');
    expect(turnPrompt).toContain('- bash (part-adv-404): error');
    expect(turnCopyBtn).toHaveTextContent('Copied');

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(turnCopyBtn).toHaveTextContent('Copy Prompt');

    // 3. Part Copy Button in Part Header (Layer 1)
    const partCopyBtn = screen.getByRole('button', {
      name: 'Copy tool execution debug prompt for part part-adv-404',
    });
    expect(partCopyBtn).toBeInTheDocument();
    expect(partCopyBtn).toHaveTextContent('Copy Prompt');

    await act(async () => {
      fireEvent.click(partCopyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(3);
    const partPrompt = writeTextMock.mock.calls[2][0];
    expect(partPrompt).toContain('# Tool Execution Failure: `bash`');
    expect(partPrompt).toContain(
      'During turn 1 of session `ses-adv-live` (message `msg-adv-1`, part `part-adv-404`), the agent executed `bash` and it failed with 5 consecutive failures on this tool.'
    );
    expect(partPrompt).toContain('cargo test --all');
    expect(partPrompt).toContain('error[E0432]: unresolved import');
    expect(partCopyBtn).toHaveTextContent('Copied');

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(partCopyBtn).toHaveTextContent('Copy Prompt');

    // 4. Failed Tool Alert Card Copy Button (Layer 2)
    fireEvent.click(screen.getByTestId('failed-tools-toggle-btn'));
    const failedCardCopyBtn = screen.getByRole('button', {
      name: 'Copy execution issue debug prompt for bash',
    });
    expect(failedCardCopyBtn).toBeInTheDocument();
    expect(failedCardCopyBtn).toHaveTextContent('Copy Prompt');

    await act(async () => {
      fireEvent.click(failedCardCopyBtn);
    });

    expect(writeTextMock).toHaveBeenCalledTimes(4);
    const issuePrompt = writeTextMock.mock.calls[3][0];
    expect(issuePrompt).toContain('# Agent Execution Issue: Consecutive failures on `bash`');
    expect(issuePrompt).toContain(
      'During turn 1 of session `ses-adv-live` (agent `assistant` running `gemini-3.8-flash`); failure part `part-adv-404`.'
    );
    expect(issuePrompt).toContain('### Error Details');
    expect(issuePrompt).toContain('```\nerror[E0432]: unresolved import\n```');
    expect(failedCardCopyBtn).toHaveTextContent('Copied');

    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(failedCardCopyBtn).toHaveTextContent('Copy Prompt');

    vi.useRealTimers();
  });
});
