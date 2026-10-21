import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { LiveTraceInspector } from './LiveTraceInspector';
import type { LiveSessionNode, LiveTraceView } from '../domain/live';
import { mockLiveTraceView } from '../test/fixtures/mockData';

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

  it('renders chronological message stream with role tags and formatted token summaries', () => {
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

    expect(
      screen.getByText('Implement Live Observability Dashboard components and streaming client for ticket AL-20.')
    ).toBeInTheDocument();
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

    // First bubble reasoning is long (> 300 chars), so initially shows 'Show Reasoning'
    const firstBubble = bubbles[0]!;
    const toggleBtn = within(firstBubble).getByRole('button');
    expect(toggleBtn).toHaveTextContent('Show Reasoning');

    // Click to show reasoning
    fireEvent.click(toggleBtn);
    expect(toggleBtn).toHaveTextContent('Hide Reasoning');

    // Click to hide reasoning again
    fireEvent.click(toggleBtn);
    expect(toggleBtn).toHaveTextContent('Show Reasoning');
  });

  it('prioritizes failed tool calls in prominent top-level alert cards above message turns', () => {
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
    expect(screen.getByText('Failed Tool Calls (3)')).toBeInTheDocument();

    const failedCards = screen.getAllByTestId('failed-tool-alert-card');
    expect(failedCards.length).toBe(3);
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

  it('strictly preserves chronological message turn order across system, user, and assistant turns', () => {
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
              severity: 'LOW',
              input_snippet: 'path: config.yaml',
              output_snippet: 'version: 1.0.0',
            },
            {
              id: 'part-fail',
              type: 'tool',
              tool_name: 'bash',
              status: 'failed',
              is_error: true,
              severity: 'HIGH',
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
});
