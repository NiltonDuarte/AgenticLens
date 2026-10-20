import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { TraceWaterfall } from './TraceWaterfall';
import type { EventFact, Anomaly } from '../domain/types';

describe('TraceWaterfall Component', () => {
  const sampleEvents: readonly EventFact[] = [
    {
      timestamp: '2026-10-18T10:00:00Z',
      session_id: 's1',
      turn: 1,
      role: 'orchestrator',
      event_type: 'turn_message',
      status: 'success',
      context_tokens: 1200,
    },
    {
      timestamp: '2026-10-18T10:01:00Z',
      session_id: 's1',
      turn: 2,
      role: 'coder',
      event_type: 'tool_call',
      tool_name: 'bash',
      status: 'success',
      duration_ms: 450,
      context_tokens: 4500,
    },
    {
      timestamp: '2026-10-18T10:02:00Z',
      session_id: 's1',
      turn: 3,
      role: 'coder',
      event_type: 'tool_call',
      tool_name: 'edit',
      status: 'error',
      error_message: 'oldString not found in content',
      duration_ms: 120,
      context_tokens: 8200,
    },
    {
      timestamp: '2026-10-18T10:03:00Z',
      session_id: 's1',
      turn: 4,
      role: 'tester',
      event_type: 'tool_call',
      tool_name: 'bash',
      status: 'success',
      duration_ms: 1250,
      context_tokens: 11000,
    },
  ];

  const sampleAnomalies: readonly Anomaly[] = [
    {
      rule_id: 'TOOL_RETRY_STREAK',
      severity: 'WARNING',
      title: 'Repeated Tool Failures',
      description: 'Tool edit failed multiple times.',
      context: {
        tool_name: 'edit',
        turn: 3,
        role: 'coder',
        count: 2,
      },
    },
  ];

  it('renders empty message when events array is empty', () => {
    render(<TraceWaterfall events={[]} />);
    expect(
      screen.getByText('Zero chronological events recorded for this execution trace.')
    ).toBeInTheDocument();
  });

  it('defaults to Errors and Anomalies Only filter mode on initial render', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);
    expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();

    const errorsToggle = screen.getByRole('button', { name: 'Show only errors and anomalies' });
    expect(errorsToggle).toHaveClass('active');

    // Only Turn 3 (error + anomaly) is visible initially
    expect(screen.getByText('Turn 3')).toBeInTheDocument();
    expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
    expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
    expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
  });

  it('renders chronological sequence of all events when All Events tab is clicked', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getByText('Turn 1')).toBeInTheDocument();
    expect(screen.getByText('Turn 2')).toBeInTheDocument();
    expect(screen.getByText('Turn 3')).toBeInTheDocument();
    expect(screen.getByText('Turn 4')).toBeInTheDocument();
  });

  it('displays role pills with proper role names when all events are shown', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getAllByText('orchestrator').length).toBeGreaterThan(0);
    expect(screen.getAllByText('coder').length).toBeGreaterThan(0);
    expect(screen.getAllByText('tester').length).toBeGreaterThan(0);
  });

  it('displays tool name, formatted duration, and status indicator for tool calls', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getAllByText('edit').length).toBeGreaterThan(0);
    expect(screen.getByText('450ms')).toBeInTheDocument();
    expect(screen.getByText('1.25s')).toBeInTheDocument();
  });

  it('correlates anomaly by turn/tool name and renders AnomalyBadge on corresponding turn item', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);
    const warningBadge = screen.getByText('WARNING');
    expect(warningBadge).toBeInTheDocument();
  });

  it('invokes onSelectEvent callback with event and anomaly when a flagged item is clicked', () => {
    const handleSelect = vi.fn();
    render(
      <TraceWaterfall
        events={sampleEvents}
        anomalies={sampleAnomalies}
        onSelectEvent={handleSelect}
      />
    );

    const anomalousTurn = screen.getByText('Turn 3');
    fireEvent.click(anomalousTurn);

    expect(handleSelect).toHaveBeenCalledTimes(1);
    expect(handleSelect).toHaveBeenCalledWith(
      expect.objectContaining({ turn: 3, tool_name: 'edit' }),
      expect.objectContaining({ rule_id: 'TOOL_RETRY_STREAK' })
    );
  });

  it('invokes onSelectEvent callback with event and undefined anomaly when an unflagged item is clicked', () => {
    const handleSelect = vi.fn();
    render(
      <TraceWaterfall
        events={sampleEvents}
        anomalies={sampleAnomalies}
        onSelectEvent={handleSelect}
      />
    );

    // Switch to All Events to access unflagged Turn 1
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    const unflaggedTurn = screen.getByText('Turn 1');
    fireEvent.click(unflaggedTurn);

    expect(handleSelect).toHaveBeenCalledTimes(1);
    expect(handleSelect).toHaveBeenCalledWith(
      expect.objectContaining({ turn: 1, role: 'orchestrator' }),
      undefined
    );
  });

  it('invokes onSelectEvent with anomaly when explicit AnomalyBadge is clicked and stops propagation', () => {
    const handleSelect = vi.fn();
    render(
      <TraceWaterfall
        events={sampleEvents}
        anomalies={sampleAnomalies}
        onSelectEvent={handleSelect}
      />
    );

    const badge = screen.getByRole('button', { name: 'WARNING' });
    fireEvent.click(badge);

    expect(handleSelect).toHaveBeenCalledTimes(1);
    expect(handleSelect).toHaveBeenCalledWith(
      expect.objectContaining({ turn: 3, tool_name: 'edit' }),
      expect.objectContaining({ rule_id: 'TOOL_RETRY_STREAK' })
    );
  });

  it('filters events by role when role filter is changed', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
    fireEvent.change(roleSelect, { target: { value: 'tester' } });

    expect(screen.getByText('Turn 4')).toBeInTheDocument();
    expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
    expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
  });

  it('filters events to tool calls only when tool calls filter is activated', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);
    const toolsBtn = screen.getByRole('button', { name: 'Tool Calls Only' });
    fireEvent.click(toolsBtn);

    expect(screen.getByText('Turn 2')).toBeInTheDocument();
    expect(screen.getByText('Turn 3')).toBeInTheDocument();
    expect(screen.getByText('Turn 4')).toBeInTheDocument();
    expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
  });

  it('highlights the active event when selectedEvent prop matches an event in the list', () => {
    render(
      <TraceWaterfall
        events={sampleEvents}
        anomalies={sampleAnomalies}
        selectedEvent={sampleEvents[2]}
      />
    );

    const timelineItem = screen.getByTestId('timeline-item-3');
    expect(timelineItem).toHaveClass('selected');
  });

  it('provides standalone local selection highlight when selectedEvent prop is undefined', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    const item2 = screen.getByTestId('timeline-item-2');
    expect(item2).not.toHaveClass('selected');

    fireEvent.click(item2);
    expect(item2).toHaveClass('selected');
  });

  it('handles events with null and missing optional fields without throwing', () => {
    const minimalEvent: EventFact = {
      timestamp: '2026-10-18T10:00:00Z',
      session_id: 's-sparse',
      turn: 1,
      role: 'orchestrator',
      event_type: 'tool_call',
      tool_name: null,
      status: 'success',
      error_message: null,
      duration_ms: null,
      context_tokens: 500,
    };

    render(<TraceWaterfall events={[minimalEvent]} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getByText('unknown')).toBeInTheDocument();
    expect(screen.getByText('Turn 1')).toBeInTheDocument();
  });

  it('handles anomaly correlation when anomaly context has missing or partial fields', () => {
    const events: readonly EventFact[] = [
      {
        timestamp: '2026-10-18T10:00:00Z',
        session_id: 's-sparse',
        turn: 1,
        role: 'coder',
        event_type: 'turn_message',
        status: 'success',
        context_tokens: 1000,
      },
    ];

    const anomaliesWithMissingContext: readonly Anomaly[] = [
      {
        rule_id: 'CONTEXT_SATURATION_SPIKE',
        severity: 'CRITICAL',
        title: 'Empty Context Anomaly',
        description: 'Context has no fields',
        context: {},
      },
      {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'WARNING',
        title: 'Turn Only Anomaly',
        description: 'Context has only turn field',
        context: {
          turn: 1,
        },
      },
    ];

    render(<TraceWaterfall events={events} anomalies={anomaliesWithMissingContext} />);
    // Turn 1 matches the second anomaly by turn
    expect(screen.getByText('WARNING')).toBeInTheDocument();
    // The first anomaly with empty context does not crash or falsely match
    expect(screen.queryByText('CRITICAL')).not.toBeInTheDocument();
  });

  it('triggers onSelectEvent on keyboard Enter and Space press', () => {
    const handleSelect = vi.fn();
    render(
      <TraceWaterfall
        events={sampleEvents}
        onSelectEvent={handleSelect}
      />
    );

    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    const timelineItem = screen.getByTestId('timeline-item-1');
    fireEvent.keyDown(timelineItem, { key: 'Enter' });
    expect(handleSelect).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(timelineItem, { key: ' ' });
    expect(handleSelect).toHaveBeenCalledTimes(2);
  });

  it('differentiates interactive classes between unflagged, error, and anomaly rows', () => {
    const anomalyOnlyList: readonly Anomaly[] = [
      {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'CRITICAL',
        title: 'Rework Loop Detected',
        description: 'Frequent rework between planner and coder.',
        context: {
          turn: 4,
          role: 'tester',
        },
      },
    ];

    render(<TraceWaterfall events={sampleEvents} anomalies={anomalyOnlyList} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    const unflaggedItem = screen.getByTestId('timeline-item-1');
    expect(unflaggedItem).toHaveClass('timeline-item');
    expect(unflaggedItem).not.toHaveClass('item-interactive');
    expect(unflaggedItem).not.toHaveClass('item-error');
    expect(unflaggedItem).not.toHaveClass('item-anomaly');

    const errorItem = screen.getByTestId('timeline-item-3');
    expect(errorItem).toHaveClass('timeline-item');
    expect(errorItem).toHaveClass('item-error');
    expect(errorItem).toHaveClass('item-interactive');

    const anomalyOnlyItem = screen.getByTestId('timeline-item-4');
    expect(anomalyOnlyItem).toHaveClass('timeline-item');
    expect(anomalyOnlyItem).toHaveClass('item-anomaly');
    expect(anomalyOnlyItem).toHaveClass('item-interactive');
    expect(anomalyOnlyItem).not.toHaveClass('item-error');
  });

  it('renders Inspect Failure button on tool error turns with proper title and aria-label', () => {
    const handleSelect = vi.fn();
    render(
      <TraceWaterfall
        events={sampleEvents}
        anomalies={sampleAnomalies}
        onSelectEvent={handleSelect}
      />
    );

    const inspectBtn = screen.getByRole('button', { name: 'Inspect failure for turn 3' });
    expect(inspectBtn).toBeInTheDocument();
    expect(inspectBtn).toHaveClass('btn-inspect-failure');
    expect(inspectBtn).toHaveAttribute('title', 'Click to open Tool Execution Failure window');
    expect(inspectBtn).toHaveTextContent(/Inspect Failure/);

    fireEvent.click(inspectBtn);
    expect(handleSelect).toHaveBeenCalledTimes(1);
    expect(handleSelect).toHaveBeenCalledWith(
      expect.objectContaining({ turn: 3, status: 'error' }),
      expect.objectContaining({ rule_id: 'TOOL_RETRY_STREAK' })
    );
  });

  it('renders AnomalyBadge with title tooltip affordance on anomalous turns', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);
    const badge = screen.getByRole('button', { name: 'WARNING' });
    expect(badge).toHaveAttribute('title', 'Click to open Anomaly Inspector');
  });

  it('renders guidance hint in waterfall toolbar header', () => {
    render(<TraceWaterfall events={sampleEvents} />);
    expect(
      screen.getByText('Click any error or anomaly to inspect failure details')
    ).toBeInTheDocument();
  });

  describe('Dual-plane dynamic hydration and payload inspector', () => {
    const hydratedEvents: readonly EventFact[] = [
      {
        timestamp: '2026-10-18T10:00:00Z',
        session_id: 's-rich-1',
        turn: 1,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'success',
        duration_ms: 320,
        context_tokens: 12000,
        prompt_content: 'Please list files in repository root.',
        reasoning_content: 'I need to check repo structure using bash ls.',
        tool_input: 'ls -la',
        tool_output: 'total 64\ndrwxr-xr-x 12 agent agent 4096 Oct 18 10:00 .',
      },
      {
        timestamp: '2026-10-18T10:01:00Z',
        session_id: 's-rich-1',
        turn: 2,
        role: 'tester',
        event_type: 'turn_message',
        status: 'success',
        context_tokens: 15000,
        prompt_content: 'Run unit test suite',
      },
    ];

    it('renders offline banner and committed ledger badge when isHydrated is false', () => {
      render(
        <TraceWaterfall
          events={sampleEvents}
          isHydrated={false}
          dataSource="committed_ledger"
        />
      );

      const banner = screen.getByTestId('db-offline-banner');
      expect(banner).toBeInTheDocument();
      expect(
        screen.getByText('Committed Git Ledger — Local Database Offline')
      ).toBeInTheDocument();
      expect(
        screen.getByText(
          /Local SQLite database was unavailable or not found\. Showing lightweight execution events recorded in the repository ledger without raw prompt, reasoning, or tool payloads\./i
        )
      ).toBeInTheDocument();

      const sourceBadge = screen.getByTestId('waterfall-source-badge');
      expect(sourceBadge).toHaveTextContent('Source: Committed Ledger');
      expect(sourceBadge).toHaveClass('source-offline');
    });

    it('hides offline banner and renders local SQLite badge when isHydrated is true', () => {
      render(
        <TraceWaterfall
          events={hydratedEvents}
          isHydrated={true}
          dataSource="local_sqlite"
        />
      );

      expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
      const sourceBadge = screen.getByTestId('waterfall-source-badge');
      expect(sourceBadge).toHaveTextContent('Source: Local SQLite');
      expect(sourceBadge).toHaveClass('source-hydrated');
    });

    it('renders and toggles rich payload details for hydrated events', () => {
      render(
        <TraceWaterfall
          events={hydratedEvents}
          isHydrated={true}
          dataSource="local_sqlite"
        />
      );

      // Switch to All Events
      const allBtn = screen.getByRole('button', { name: 'All Events' });
      fireEvent.click(allBtn);

      const viewPayloadBtn = screen.getByRole('button', { name: 'View payload for turn 1' });
      expect(viewPayloadBtn).toBeInTheDocument();
      expect(screen.queryByTestId('payload-details-1')).not.toBeInTheDocument();

      // Click to expand payload
      fireEvent.click(viewPayloadBtn);
      expect(screen.getByTestId('payload-details-1')).toBeInTheDocument();
      expect(screen.getByTestId('prompt-payload-block')).toHaveTextContent(
        'Please list files in repository root.'
      );
      expect(screen.getByTestId('reasoning-payload-block')).toHaveTextContent(
        'I need to check repo structure using bash ls.'
      );
      expect(screen.getByTestId('tool-input-payload-block')).toHaveTextContent('ls -la');
      expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('total 64');

      // Click to hide payload
      const hidePayloadBtn = screen.getByRole('button', { name: 'Hide payload for turn 1' });
      fireEvent.click(hidePayloadBtn);
      expect(screen.queryByTestId('payload-details-1')).not.toBeInTheDocument();
    });

    it('renders unhydrated notice for events when database is offline', () => {
      render(
        <TraceWaterfall
          events={sampleEvents}
          isHydrated={false}
          dataSource="committed_ledger"
        />
      );

      // Turn 3 is visible under default filter
      const viewPayloadBtn = screen.getByRole('button', { name: 'View payload for turn 3' });
      fireEvent.click(viewPayloadBtn);

      expect(screen.getByTestId('payload-details-3')).toBeInTheDocument();
      expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent(
        'Full payload not captured in Git ledger'
      );
    });

    it('renders zero payload captured notice when hydrated event has no payload fields', () => {
      const bareHydratedEvents: readonly EventFact[] = [
        {
          timestamp: '2026-10-18T10:00:00Z',
          session_id: 's-bare-1',
          turn: 1,
          role: 'coder',
          event_type: 'turn_message',
          status: 'success',
          context_tokens: 1000,
        },
      ];

      render(
        <TraceWaterfall
          events={bareHydratedEvents}
          isHydrated={true}
          dataSource="local_sqlite"
        />
      );

      const allBtn = screen.getByRole('button', { name: 'All Events' });
      fireEvent.click(allBtn);

      const viewPayloadBtn = screen.getByRole('button', { name: 'View payload for turn 1' });
      fireEvent.click(viewPayloadBtn);

      expect(screen.getByTestId('payload-details-1')).toBeInTheDocument();
      expect(screen.getByText('Zero payload data captured for this event')).toBeInTheDocument();
    });

    it('ensures zero emojis across all rendered waterfall components in both modes', () => {
      const { container: hydratedContainer } = render(
        <TraceWaterfall
          events={hydratedEvents}
          isHydrated={true}
          dataSource="local_sqlite"
        />
      );

      const emojiRegex = /[\u{1F600}-\u{1F64F}\u{1F300}-\u{1F5FF}\u{1F680}-\u{1F6FF}\u{1F1E0}-\u{1F1FF}\u{2702}-\u{27B0}\u{24C2}-\u{1F251}\u{1F900}-\u{1F9FF}\u{1FA70}-\u{1FAFF}]/u;
      expect(emojiRegex.test(hydratedContainer.textContent ?? '')).toBe(false);

      const { container: offlineContainer } = render(
        <TraceWaterfall
          events={sampleEvents}
          isHydrated={false}
          dataSource="committed_ledger"
        />
      );
      expect(emojiRegex.test(offlineContainer.textContent ?? '')).toBe(false);
    });
  });

  describe('Adversarial Multi-dimensional Filtering', () => {
    const adversarialEvents: readonly EventFact[] = [
      {
        timestamp: '2026-10-18T10:00:00Z',
        session_id: 's-adv',
        turn: 1,
        role: 'orchestrator',
        event_type: 'turn_message',
        status: 'success',
        context_tokens: 1500,
        prompt_content: 'Coordinate decomposition and task assignment',
      },
      {
        timestamp: '2026-10-18T10:01:00Z',
        session_id: 's-adv',
        turn: 2,
        role: 'orchestrator',
        event_type: 'tool_call',
        tool_name: 'task',
        status: 'success',
        duration_ms: 1200,
        context_tokens: 3000,
        tool_input: '{"subagent": "coder"}',
      },
      {
        timestamp: '2026-10-18T10:02:00Z',
        session_id: 's-adv',
        turn: 3,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'read',
        status: 'success',
        duration_ms: 250,
        context_tokens: 4500,
        tool_input: '{"path": "CONTRACT.md"}',
      },
      {
        timestamp: '2026-10-18T10:03:00Z',
        session_id: 's-adv',
        turn: 4,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'error',
        error_message: 'command exit code 1',
        duration_ms: 800,
        context_tokens: 6000,
      },
      {
        timestamp: '2026-10-18T10:04:00Z',
        session_id: 's-adv',
        turn: 5,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'edit',
        status: 'error',
        error_message: 'oldString not found in content',
        duration_ms: 150,
        context_tokens: 18500,
      },
      {
        timestamp: '2026-10-18T10:05:00Z',
        session_id: 's-adv',
        turn: 6,
        role: 'tester',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'success',
        duration_ms: 3200,
        context_tokens: 19000,
      },
      {
        timestamp: '2026-10-18T10:06:00Z',
        session_id: 's-adv',
        turn: 7,
        role: 'reviewer',
        event_type: 'tool_call',
        tool_name: 'read',
        status: 'success',
        duration_ms: 400,
        context_tokens: 19500,
      },
      {
        timestamp: '2026-10-18T10:07:00Z',
        session_id: 's-adv',
        turn: 8,
        role: 'reviewer',
        event_type: 'compaction',
        status: 'success',
        context_tokens: 20000,
      },
      {
        timestamp: '2026-10-18T10:08:00Z',
        session_id: 's-adv',
        turn: 9,
        role: 'tester',
        event_type: 'tool_call',
        tool_name: null,
        status: 'error',
        error_message: 'unspecified execution fault',
        context_tokens: 20500,
      },
      {
        timestamp: '2026-10-18T10:09:00Z',
        session_id: 's-adv',
        turn: 10,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'edit',
        status: 'error',
        error_message: 'syntax error in patch',
        duration_ms: 95,
        context_tokens: 21000,
      },
    ];

    const adversarialAnomalies: readonly Anomaly[] = [
      {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
        title: 'Repeated Tool Failures',
        description: 'Edit failed repeatedly.',
        context: {
          tool_name: 'edit',
          turn: 5,
          role: 'coder',
        },
      },
      {
        rule_id: 'CONTEXT_SATURATION_SPIKE',
        severity: 'CRITICAL',
        title: 'Context Saturation Spike',
        description: 'Token count jumped abruptly.',
        context: {
          turn: 5,
          session_id: 's-adv',
        },
      },
      {
        rule_id: 'TOOL_HIGH_FAILURE_RATE',
        severity: 'WARNING',
        title: 'High Failure Rate',
        description: 'Bash failure rate exceeded threshold.',
        context: {
          tool_name: 'bash',
          turn: 4,
          role: 'coder',
        },
      },
      {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'CRITICAL',
        title: 'Subagent Rework Loop',
        description: 'Tester called without progress.',
        context: {
          role: 'tester',
          turn: 6,
        },
      },
      {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
        title: 'Duplicate Retry Streak',
        description: 'Another retry streak anomaly on turn 10.',
        context: {
          tool_name: 'edit',
          turn: 10,
          role: 'coder',
        },
      },
    ];

    describe('Tool Filter Invariants & Resilience', () => {
      it('renders unique tools sorted alphabetically in dropdown and excludes null/empty entries', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const toolSelect = screen.getByTestId('trace-tool-filter');
        expect(toolSelect).toBeInTheDocument();
        expect(toolSelect).toHaveAttribute('aria-label', 'Filter by tool');

        const toolOptions = Array.from(toolSelect.querySelectorAll('option')).map(
          (opt) => opt.value
        );
        expect(toolOptions).toEqual(['all', 'bash', 'edit', 'read', 'task']);
      });

      it('filters events by specific tool name across task, bash, read, edit and restores on all', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        // Switch to All Events mode so non-error tool calls are visible
        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const toolSelect = screen.getByTestId('trace-tool-filter');

        // Filter by 'task'
        fireEvent.change(toolSelect, { target: { value: 'task' } });
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Filter by 'read'
        fireEvent.change(toolSelect, { target: { value: 'read' } });
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Filter by 'bash'
        fireEvent.change(toolSelect, { target: { value: 'bash' } });
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Filter by 'edit'
        fireEvent.change(toolSelect, { target: { value: 'edit' } });
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Restore to 'all'
        fireEvent.change(toolSelect, { target: { value: 'all' } });
        for (let turn = 1; turn <= 10; turn++) {
          expect(screen.getByText(`Turn ${turn}`)).toBeInTheDocument();
        }
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();
      });

      it('excludes events with null or undefined tool_name when specific tool is active and preserves them on all', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const toolSelect = screen.getByTestId('trace-tool-filter');
        fireEvent.change(toolSelect, { target: { value: 'bash' } });

        // Turn 1 (turn_message, undefined tool), Turn 8 (compaction, undefined tool), Turn 9 (tool_call, null tool)
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 8')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();

        // Revert to 'all'
        fireEvent.change(toolSelect, { target: { value: 'all' } });
        expect(screen.getByText('Turn 1')).toBeInTheDocument();
        expect(screen.getByText('Turn 8')).toBeInTheDocument();
        expect(screen.getByText('Turn 9')).toBeInTheDocument();
      });

      it('handles events where all tool names are null, undefined, or empty whitespace without crashing', () => {
        const nonToolEvents: readonly EventFact[] = [
          {
            timestamp: '2026-10-18T10:00:00Z',
            session_id: 's-none',
            turn: 1,
            role: 'orchestrator',
            event_type: 'turn_message',
            status: 'success',
            context_tokens: 1000,
          },
          {
            timestamp: '2026-10-18T10:01:00Z',
            session_id: 's-none',
            turn: 2,
            role: 'coder',
            event_type: 'tool_call',
            tool_name: '   ',
            status: 'success',
            context_tokens: 1200,
          },
          {
            timestamp: '2026-10-18T10:02:00Z',
            session_id: 's-none',
            turn: 3,
            role: 'reviewer',
            event_type: 'compaction',
            status: 'success',
            context_tokens: 1400,
          },
        ];

        render(<TraceWaterfall events={nonToolEvents} anomalies={[]} />);

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const options = Array.from(toolSelect.querySelectorAll('option')).map((o) => o.value);
        expect(options).toEqual(['all']);
        expect(toolSelect).toHaveTextContent('All Tools');
      });
    });

    describe('Anomaly Filter Invariants & Resilience', () => {
      it('renders unique anomaly rule IDs sorted alphabetically and deduplicates whitespace variants', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toBeInTheDocument();
        expect(anomalySelect).toHaveAttribute('aria-label', 'Filter by anomaly rule');

        const anomalyOptions = Array.from(anomalySelect.querySelectorAll('option')).map(
          (opt) => opt.value
        );
        expect(anomalyOptions).toEqual([
          'all',
          'CONTEXT_SATURATION_SPIKE',
          'SUBAGENT_REWORK_LOOP',
          'TOOL_HIGH_FAILURE_RATE',
          'TOOL_RETRY_STREAK',
        ]);
      });

      it('filters events matching specific anomaly rule_id across tool, turn, and role correlation rules', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        // 1. CONTEXT_SATURATION_SPIKE (matches Turn 5 by turn and session)
        fireEvent.change(anomalySelect, { target: { value: 'CONTEXT_SATURATION_SPIKE' } });
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 10')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 2. SUBAGENT_REWORK_LOOP (matches Turn 6 by role and turn)
        fireEvent.change(anomalySelect, { target: { value: 'SUBAGENT_REWORK_LOOP' } });
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 3. TOOL_HIGH_FAILURE_RATE (matches Turn 4 by tool and turn)
        fireEvent.change(anomalySelect, { target: { value: 'TOOL_HIGH_FAILURE_RATE' } });
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 4. TOOL_RETRY_STREAK (matches Turn 5 and Turn 10)
        fireEvent.change(anomalySelect, { target: { value: 'TOOL_RETRY_STREAK' } });
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // 5. Restore to all
        fireEvent.change(anomalySelect, { target: { value: 'all' } });
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 9')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 5 of 10 events/)).toBeInTheDocument();
      });

      it('correctly matches event associated with multiple anomaly rules under each rule filter', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        // Turn 5 matches both CONTEXT_SATURATION_SPIKE and TOOL_RETRY_STREAK
        fireEvent.change(anomalySelect, { target: { value: 'CONTEXT_SATURATION_SPIKE' } });
        expect(screen.getByText('Turn 5')).toBeInTheDocument();

        fireEvent.change(anomalySelect, { target: { value: 'TOOL_RETRY_STREAK' } });
        expect(screen.getByText('Turn 5')).toBeInTheDocument();

        // But Turn 5 does not match SUBAGENT_REWORK_LOOP
        fireEvent.change(anomalySelect, { target: { value: 'SUBAGENT_REWORK_LOOP' } });
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
      });

      it('handles empty anomalies array and undefined anomalies prop defensively', () => {
        const { rerender } = render(<TraceWaterfall events={adversarialEvents} anomalies={[]} />);

        let anomalySelect = screen.getByTestId('trace-anomaly-filter');
        let options = Array.from(anomalySelect.querySelectorAll('option')).map((o) => o.value);
        expect(options).toEqual(['all']);
        expect(anomalySelect).toHaveTextContent('All Anomalies');

        // Rerender with undefined anomalies
        rerender(<TraceWaterfall events={adversarialEvents} anomalies={undefined} />);
        anomalySelect = screen.getByTestId('trace-anomaly-filter');
        options = Array.from(anomalySelect.querySelectorAll('option')).map((o) => o.value);
        expect(options).toEqual(['all']);
      });

      it('displays empty state and zero counter when selected anomaly rule matches no events in trace', () => {
        const disjointAnomalies: readonly Anomaly[] = [
          {
            rule_id: 'CONTEXT_SATURATION_SPIKE',
            severity: 'CRITICAL',
            title: 'Uncorrelated Spike',
            description: 'Uncorrelated anomaly on nonexistent turn 999.',
            context: {
              turn: 999,
            },
          },
        ];

        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={disjointAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        fireEvent.change(anomalySelect, { target: { value: 'CONTEXT_SATURATION_SPIKE' } });

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Recover on 'all'
        fireEvent.change(anomalySelect, { target: { value: 'all' } });
        expect(screen.queryByText('No events match the selected filters.')).not.toBeInTheDocument();
        // In default errors_anomalies mode, plain errors 4, 5, 9, 10 are visible (4 events)
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();
      });
    });

    describe('Multi-Filter Composition Engine', () => {
      it('simultaneously composes filterMode, roleFilter, toolFilter, and anomalyFilter across positive and negative permutations', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const toolsBtn = screen.getByRole('button', { name: 'Tool Calls Only' });
        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        // Permutation 1: Tool Calls + Coder + bash + TOOL_HIGH_FAILURE_RATE -> exactly Turn 4
        fireEvent.click(toolsBtn);
        fireEvent.change(roleSelect, { target: { value: 'coder' } });
        fireEvent.change(toolSelect, { target: { value: 'bash' } });
        fireEvent.change(anomalySelect, { target: { value: 'TOOL_HIGH_FAILURE_RATE' } });

        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Permutation 2: Orthogonal conflict between Tool and Anomaly:
        // Coder + edit + TOOL_HIGH_FAILURE_RATE (failure rate anomaly is on bash, not edit)
        fireEvent.change(toolSelect, { target: { value: 'edit' } });
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Permutation 3: Role vs Tool conflict:
        // Switch to All Events mode, Reviewer role, 'task' tool (task was invoked by orchestrator)
        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);
        fireEvent.change(roleSelect, { target: { value: 'reviewer' } });
        fireEvent.change(toolSelect, { target: { value: 'task' } });
        fireEvent.change(anomalySelect, { target: { value: 'all' } });

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Permutation 4: Non-tool event under tool filter:
        // Orchestrator role + task tool -> shows Turn 2 (task), excludes Turn 1 (turn_message)
        fireEvent.change(roleSelect, { target: { value: 'orchestrator' } });
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Permutation 5: Errors & Anomalies mode + Tester role + bash tool + SUBAGENT_REWORK_LOOP
        const errorsBtn = screen.getByRole('button', { name: 'Show only errors and anomalies' });
        fireEvent.click(errorsBtn);
        fireEvent.change(roleSelect, { target: { value: 'tester' } });
        fireEvent.change(toolSelect, { target: { value: 'bash' } });
        fireEvent.change(anomalySelect, { target: { value: 'SUBAGENT_REWORK_LOOP' } });

        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();
      });

      it('progressively narrows down events and incrementally restores them across all filter dimensions', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        const errorsBtn = screen.getByRole('button', { name: 'Show only errors and anomalies' });
        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        // Step 1: All Events -> 10 events
        fireEvent.click(allBtn);
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();

        // Step 2: Restrict role to 'coder' -> 4 events (Turns 3, 4, 5, 10)
        fireEvent.change(roleSelect, { target: { value: 'coder' } });
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Step 3: Restrict tool to 'edit' -> 2 events (Turns 5, 10)
        fireEvent.change(toolSelect, { target: { value: 'edit' } });
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Step 4: Restrict anomaly to 'CONTEXT_SATURATION_SPIKE' -> 1 event (Turn 5)
        fireEvent.change(anomalySelect, { target: { value: 'CONTEXT_SATURATION_SPIKE' } });
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Step 5: Switch to errors_anomalies mode -> Turn 5 is an error with spike -> still 1 event
        fireEvent.click(errorsBtn);
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Step 6: Step-by-step relaxation:
        // Relax anomaly to 'all' -> 2 events (Turns 5, 10)
        fireEvent.change(anomalySelect, { target: { value: 'all' } });
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Relax tool to 'all' -> 3 events in errors_anomalies mode (Turns 4, 5, 10 - Turn 3 is success without anomaly)
        fireEvent.change(toolSelect, { target: { value: 'all' } });
        expect(screen.getByText(/Showing 3 of 10 events/)).toBeInTheDocument();

        // Switch to 'all' mode -> 4 events (Turns 3, 4, 5, 10)
        fireEvent.click(allBtn);
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Relax role to 'all' -> 10 events
        fireEvent.change(roleSelect, { target: { value: 'all' } });
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();
      });

      it('dynamically toggles filterMode while tool and anomaly filters remain retained', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        const toolsBtn = screen.getByRole('button', { name: 'Tool Calls Only' });
        const errorsBtn = screen.getByRole('button', { name: 'Show only errors and anomalies' });
        const toolSelect = screen.getByTestId('trace-tool-filter');

        // Set tool to 'task'
        fireEvent.change(toolSelect, { target: { value: 'task' } });

        // Default mode is errors_anomalies: Turn 2 (task) succeeded with no anomaly -> 0 events
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Switch to All Events: Turn 2 becomes visible, tool select value remains 'task'
        fireEvent.click(allBtn);
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(toolSelect).toHaveValue('task');
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Switch to Tool Calls Only: Turn 2 is a tool call -> still visible, tool select remains 'task'
        fireEvent.click(toolsBtn);
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(toolSelect).toHaveValue('task');
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Switch back to errors_anomalies: returns to 0 matching events
        fireEvent.click(errorsBtn);
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();
      });

      it('dispatches onSelectEvent callback with filtered event and correlated anomaly when clicked in multi-filter view', () => {
        const handleSelect = vi.fn();
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
            onSelectEvent={handleSelect}
          />
        );

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        // Filter to tool 'edit' and anomaly 'TOOL_RETRY_STREAK'
        fireEvent.change(toolSelect, { target: { value: 'edit' } });
        fireEvent.change(anomalySelect, { target: { value: 'TOOL_RETRY_STREAK' } });

        // Turn 5 and Turn 10 match
        const turn5Item = screen.getByText('Turn 5');
        fireEvent.click(turn5Item);

        expect(handleSelect).toHaveBeenCalledTimes(1);
        expect(handleSelect).toHaveBeenCalledWith(
          expect.objectContaining({ turn: 5, tool_name: 'edit' }),
          expect.objectContaining({ rule_id: 'TOOL_RETRY_STREAK' })
        );
      });
    });

    describe('Counter Accuracy and Denominator Invariance', () => {
      it('maintains total event denominator invariance across arbitrary filter combinations', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const counter = screen.getByText(/Showing \d+ of \d+ events/);
        expect(counter).toHaveTextContent('Showing 5 of 10 events');

        const toolSelect = screen.getByTestId('trace-tool-filter');
        fireEvent.change(toolSelect, { target: { value: 'edit' } });
        expect(counter).toHaveTextContent('Showing 2 of 10 events');

        fireEvent.change(toolSelect, { target: { value: 'task' } });
        expect(counter).toHaveTextContent('Showing 0 of 10 events');

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);
        fireEvent.change(toolSelect, { target: { value: 'all' } });
        expect(counter).toHaveTextContent('Showing 10 of 10 events');
      });

      it('handles single-event array with accurate 1 of 1 and 0 of 1 counter', () => {
        const singleEvent: readonly EventFact[] = [
          {
            timestamp: '2026-10-18T10:00:00Z',
            session_id: 's-single',
            turn: 1,
            role: 'coder',
            event_type: 'tool_call',
            tool_name: 'bash',
            status: 'error',
            context_tokens: 500,
          },
        ];

        render(<TraceWaterfall events={singleEvent} anomalies={[]} />);

        expect(screen.getByText(/Showing 1 of 1 events/)).toBeInTheDocument();

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        fireEvent.change(roleSelect, { target: { value: 'tester' } });

        expect(screen.getByText(/Showing 0 of 1 events/)).toBeInTheDocument();
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
      });
    });

    describe('Accessibility and Screen Reader Attributes', () => {
      it('provides accessible labels, aria-labels, and semantic associations for all filter controls', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        // Tool filter accessibility
        const toolSelect = screen.getByTestId('trace-tool-filter');
        expect(toolSelect).toHaveAttribute('id', 'trace-tool-filter');
        expect(toolSelect).toHaveAttribute('aria-label', 'Filter by tool');
        expect(screen.getByLabelText('Tool:')).toBe(toolSelect);
        expect(screen.getByRole('combobox', { name: 'Filter by tool' })).toBe(toolSelect);

        // Anomaly filter accessibility
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toHaveAttribute('id', 'trace-anomaly-filter');
        expect(anomalySelect).toHaveAttribute('aria-label', 'Filter by anomaly rule');
        expect(screen.getByLabelText('Anomaly:')).toBe(anomalySelect);
        expect(screen.getByRole('combobox', { name: 'Filter by anomaly rule' })).toBe(anomalySelect);

        // Role filter accessibility
        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        expect(roleSelect).toHaveAttribute('id', 'trace-role-filter');
        expect(roleSelect).toHaveAttribute('aria-label', 'Filter by subagent role');
        expect(screen.getByLabelText('Role:')).toBe(roleSelect);

        // Mode buttons group accessibility
        const filterGroup = screen.getByRole('group', { name: 'Filter events by type' });
        expect(filterGroup).toBeInTheDocument();
      });
    });

    describe('Zero Emojis Invariant Across Multi-Filter States', () => {
      it('maintains zero emojis across all multi-filter toolbar elements, dropdowns, and empty states', () => {
        const { container } = render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const emojiRegex = /[\u{1F600}-\u{1F64F}\u{1F300}-\u{1F5FF}\u{1F680}-\u{1F6FF}\u{1F1E0}-\u{1F1FF}\u{2702}-\u{27B0}\u{24C2}-\u{1F251}\u{1F900}-\u{1F9FF}\u{1FA70}-\u{1FAFF}]/u;
        expect(emojiRegex.test(container.textContent ?? '')).toBe(false);

        // Trigger empty state
        const toolSelect = screen.getByTestId('trace-tool-filter');
        fireEvent.change(toolSelect, { target: { value: 'task' } });
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(emojiRegex.test(container.textContent ?? '')).toBe(false);
      });
    });
  });
});
