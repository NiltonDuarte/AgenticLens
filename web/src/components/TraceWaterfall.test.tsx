import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
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
    const roleContainer = roleSelect.closest('.multiselect-container')!;
    fireEvent.click(roleSelect);
    fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'tester' }));

    expect(screen.getByText('Turn 4')).toBeInTheDocument();
    expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
    expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
    expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
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
        expect(toolSelect.textContent).toContain('No filter');

        fireEvent.click(toolSelect);
        const container = toolSelect.closest('.multiselect-container')!;
        const checkboxes = within(container as HTMLElement).getAllByRole('checkbox');
        const toolOptions = checkboxes.map((cb) => cb.getAttribute('aria-label'));
        expect(toolOptions).toEqual(['No filter', 'bash', 'edit', 'read', 'task']);
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
        const container = toolSelect.closest('.multiselect-container')!;

        // Filter by 'task'
        fireEvent.click(toolSelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'task' }));
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Filter by 'read' (deselect task, select read)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'read' }));
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Filter by 'bash' (deselect read, select bash)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'read' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Filter by 'edit' (deselect bash, select edit)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Restore to 'all' via 'No filter'
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
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
        const container = toolSelect.closest('.multiselect-container')!;
        fireEvent.click(toolSelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'bash' }));

        // Turn 1 (turn_message, undefined tool), Turn 8 (compaction, undefined tool), Turn 9 (tool_call, null tool)
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 8')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();

        // Revert to 'all' via 'No filter'
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
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
        expect(toolSelect).toHaveTextContent('No filter');

        fireEvent.click(toolSelect);
        const container = toolSelect.closest('.multiselect-container')!;
        const options = within(container as HTMLElement)
          .getAllByRole('checkbox')
          .map((cb) => cb.getAttribute('aria-label'));
        expect(options).toEqual(['No filter']);
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
        expect(anomalySelect).toHaveTextContent('No filter');

        fireEvent.click(anomalySelect);
        const container = anomalySelect.closest('.multiselect-container')!;
        const checkboxes = within(container as HTMLElement).getAllByRole('checkbox');
        const anomalyOptions = checkboxes.map((cb) => cb.getAttribute('aria-label'));
        expect(anomalyOptions).toEqual([
          'No filter',
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
        const container = anomalySelect.closest('.multiselect-container')!;

        // 1. CONTEXT_SATURATION_SPIKE (matches Turn 5 by turn and session)
        fireEvent.click(anomalySelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 10')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 2. SUBAGENT_REWORK_LOOP (matches Turn 6 by role and turn)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 3. TOOL_HIGH_FAILURE_RATE (matches Turn 4 by tool and turn)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'TOOL_HIGH_FAILURE_RATE' }));
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // 4. TOOL_RETRY_STREAK (matches Turn 5 and Turn 10)
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'TOOL_HIGH_FAILURE_RATE' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // 5. Restore to all via No filter
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
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
        const container = anomalySelect.closest('.multiselect-container')!;

        // Turn 5 matches both CONTEXT_SATURATION_SPIKE and TOOL_RETRY_STREAK
        fireEvent.click(anomalySelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        expect(screen.getByText('Turn 5')).toBeInTheDocument();

        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        expect(screen.getByText('Turn 5')).toBeInTheDocument();

        // But Turn 5 does not match SUBAGENT_REWORK_LOOP
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
      });

      it('handles empty anomalies array and undefined anomalies prop defensively', () => {
        const { rerender } = render(<TraceWaterfall events={adversarialEvents} anomalies={[]} />);

        let anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toHaveTextContent('No filter');
        fireEvent.click(anomalySelect);
        let container = anomalySelect.closest('.multiselect-container')!;
        let options = within(container as HTMLElement)
          .getAllByRole('checkbox')
          .map((cb) => cb.getAttribute('aria-label'));
        expect(options).toEqual(['No filter']);

        // Rerender with undefined anomalies
        rerender(<TraceWaterfall events={adversarialEvents} anomalies={undefined} />);
        anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toHaveTextContent('No filter');
        if (anomalySelect.getAttribute('aria-expanded') !== 'true') {
          fireEvent.click(anomalySelect);
        }
        container = anomalySelect.closest('.multiselect-container')!;
        options = within(container as HTMLElement)
          .getAllByRole('checkbox')
          .map((cb) => cb.getAttribute('aria-label'));
        expect(options).toEqual(['No filter']);
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
        const container = anomalySelect.closest('.multiselect-container')!;
        fireEvent.click(anomalySelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Recover on 'all' via 'No filter'
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
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

        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const toolContainer = toolSelect.closest('.multiselect-container')!;
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // Permutation 1: Tool Calls + Coder + bash + TOOL_HIGH_FAILURE_RATE -> exactly Turn 4
        fireEvent.click(toolsBtn);
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(roleSelect);

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(toolSelect);

        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_HIGH_FAILURE_RATE' }));
        fireEvent.click(anomalySelect);

        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Permutation 2: Orthogonal conflict between Tool and Anomaly:
        // Coder + edit + TOOL_HIGH_FAILURE_RATE (failure rate anomaly is on bash, not edit)
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Permutation 3: Role vs Tool conflict:
        // Switch to All Events mode, Reviewer role, 'task' tool (task was invoked by orchestrator)
        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'reviewer' }));
        fireEvent.click(roleSelect);

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);

        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(anomalySelect);

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Permutation 4: Non-tool event under tool filter:
        // Orchestrator role + task tool -> shows Turn 2 (task), excludes Turn 1 (turn_message)
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'reviewer' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'orchestrator' }));
        fireEvent.click(roleSelect);

        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Permutation 5: Errors & Anomalies mode + Tester role + bash tool + SUBAGENT_REWORK_LOOP
        const errorsBtn = screen.getByRole('button', { name: 'Show only errors and anomalies' });
        fireEvent.click(errorsBtn);

        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'orchestrator' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'tester' }));
        fireEvent.click(roleSelect);

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(toolSelect);

        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        fireEvent.click(anomalySelect);

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

        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const toolContainer = toolSelect.closest('.multiselect-container')!;
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // Step 1: All Events -> 10 events
        fireEvent.click(allBtn);
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();

        // Step 2: Restrict role to 'coder' -> 4 events (Turns 3, 4, 5, 10)
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(roleSelect);
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Step 3: Restrict tool to 'edit' -> 2 events (Turns 5, 10)
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Step 4: Restrict anomaly to 'CONTEXT_SATURATION_SPIKE' -> 1 event (Turn 5)
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        fireEvent.click(anomalySelect);
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Step 5: Switch to errors_anomalies mode -> Turn 5 is an error with spike -> still 1 event
        fireEvent.click(errorsBtn);
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Step 6: Step-by-step relaxation:
        // Relax anomaly to 'No filter' -> 2 events (Turns 5, 10)
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(anomalySelect);
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Relax tool to 'No filter' -> 3 events in errors_anomalies mode (Turns 4, 5, 10 - Turn 3 is success without anomaly)
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(toolSelect);
        expect(screen.getByText(/Showing 3 of 10 events/)).toBeInTheDocument();

        // Switch to 'all' mode -> 4 events (Turns 3, 4, 5, 10)
        fireEvent.click(allBtn);
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Relax role to 'No filter' -> 10 events
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(roleSelect);
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
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Set tool to 'task'
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);

        // Default mode is errors_anomalies: Turn 2 (task) succeeded with no anomaly -> 0 events
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Switch to All Events: Turn 2 becomes visible, tool select summary remains 'task'
        fireEvent.click(allBtn);
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(toolSelect.textContent).toContain('task');
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Switch to Tool Calls Only: Turn 2 is a tool call -> still visible, tool select remains 'task'
        fireEvent.click(toolsBtn);
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(toolSelect.textContent).toContain('task');
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
        const toolContainer = toolSelect.closest('.multiselect-container')!;
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // Filter to tool 'edit' and anomaly 'TOOL_RETRY_STREAK'
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(anomalySelect);

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
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);
        expect(counter).toHaveTextContent('Showing 2 of 10 events');

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);
        expect(counter).toHaveTextContent('Showing 0 of 10 events');

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(toolSelect);
        expect(counter).toHaveTextContent('Showing 10 of 10 events');
      });

      it('handles single-event array with accurate 1 of 1 and 0 of 1 counter', () => {
        const singleEvent: readonly EventFact[] = [
          {
            timestamp: '2026-10-18T10:00:00Z',
            session_id: 's-single',
            turn: 1,
            role: 'coder',
            event_type: 'turn_message',
            status: 'error',
            context_tokens: 500,
          },
        ];

        render(<TraceWaterfall events={singleEvent} anomalies={[]} />);

        expect(screen.getByText(/Showing 1 of 1 events/)).toBeInTheDocument();

        // Switch to Tool Calls Only -> Turn 1 is turn_message -> 0 of 1 events
        const toolsBtn = screen.getByRole('button', { name: 'Tool Calls Only' });
        fireEvent.click(toolsBtn);

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
        expect(toolSelect).toHaveAttribute('id', 'trace-tool-filter-trigger');
        expect(toolSelect).toHaveAttribute('aria-label', 'Filter by tool');
        expect(screen.getByLabelText('Tool:')).toBe(toolSelect);
        expect(screen.getByRole('combobox', { name: 'Filter by tool' })).toBe(toolSelect);

        // Anomaly filter accessibility
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toHaveAttribute('id', 'trace-anomaly-filter-trigger');
        expect(anomalySelect).toHaveAttribute('aria-label', 'Filter by anomaly rule');
        expect(screen.getByLabelText('Anomaly:')).toBe(anomalySelect);
        expect(screen.getByRole('combobox', { name: 'Filter by anomaly rule' })).toBe(anomalySelect);

        // Role filter accessibility
        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        expect(roleSelect).toHaveAttribute('id', 'trace-role-filter-trigger');
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

        // Trigger empty state by filtering tool to task (which has no errors in default mode)
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const toolContainer = toolSelect.closest('.multiselect-container')!;
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(emojiRegex.test(container.textContent ?? '')).toBe(false);
      });
    });

    describe('Multi-Select & "No filter" Enhancements', () => {
      it('initial mount displays "No filter" on Role, Tool, and Anomaly triggers and passes through all events in All Events mode', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        expect(roleSelect.textContent).toContain('No filter');
        expect(toolSelect.textContent).toContain('No filter');
        expect(anomalySelect.textContent).toContain('No filter');

        // Switch to All Events mode
        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        // With "No filter" active on all 3, all events pass including non-tool events (Turns 1, 8, 9)
        expect(screen.getByText('Turn 1')).toBeInTheDocument();
        expect(screen.getByText('Turn 8')).toBeInTheDocument();
        expect(screen.getByText('Turn 9')).toBeInTheDocument();
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();
      });

      it('supports multi-tool selection (e.g. bash + edit) simultaneously', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Select bash
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        // Select edit without closing
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('bash, edit');

        // Turns 4, 6 (bash) and Turns 5, 10 (edit) are visible
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        // Other tools and non-tool turns excluded
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 7')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 8')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();

        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();
      });

      it('supports multi-anomaly selection (e.g. TOOL_RETRY_STREAK + CONTEXT_SATURATION_SPIKE)', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));
        fireEvent.click(anomalySelect);

        // Turn 5 (spike & retry) and Turn 10 (retry streak) match
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        // Turn 4 (failure rate only) and Turn 6 (rework loop only) excluded
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();

        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();
      });

      it('supports multi-role selection (e.g. coder + tester)', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const roleContainer = roleSelect.closest('.multiselect-container')!;

        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'tester' }));
        fireEvent.click(roleSelect);

        expect(roleSelect.textContent).toContain('coder, tester');

        // Coder (Turns 3, 4, 5, 10) and Tester (Turns 6, 9)
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 9')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        // Orchestrator (Turns 1, 2) and Reviewer (Turns 7, 8) excluded
        expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 2')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 7')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 8')).not.toBeInTheDocument();

        expect(screen.getByText(/Showing 6 of 10 events/)).toBeInTheDocument();
      });

      it('displays count summary format when three or more options are selected', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'read' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('3 selected');
      });

      it('toggling an option on and then off restores "No filter" pass-through', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Select task
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Deselect task
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'task' }));
        fireEvent.click(toolSelect);

        // Reverted to No filter
        expect(toolSelect.textContent).toContain('No filter');
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();
      });

      it('composes multi-tool and multi-role conjunction in errors_anomalies mode', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Select coder
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(roleSelect);

        // Select bash + edit
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        // In default errors_anomalies mode:
        // Turn 4 (coder, bash, error)
        // Turn 5 (coder, edit, error)
        // Turn 10 (coder, edit, error)
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument(); // coder read success
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument(); // tester
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument(); // tester null tool

        expect(screen.getByText(/Showing 3 of 10 events/)).toBeInTheDocument();
      });

      it('verifies non-tool event pass-through when Tool is "No filter" and exclusion when specific tool is active', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Filter role to 'reviewer'
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'reviewer' }));
        fireEvent.click(roleSelect);

        // Reviewer has Turn 7 (tool_call: read) and Turn 8 (compaction: no tool)
        // Since Tool filter is "No filter", Turn 8 passes through
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.getByText('Turn 8')).toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Now restrict Tool to 'read'
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'read' }));
        fireEvent.click(toolSelect);

        // Turn 8 (compaction, no tool) is now excluded
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.queryByText('Turn 8')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 1 of 10 events/)).toBeInTheDocument();

        // Restore Tool to "No filter" via checkbox
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(toolSelect);

        // Turn 8 immediately passes through again
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.getByText('Turn 8')).toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();
      });

      it('verifies non-anomaly event pass-through when Anomaly is "No filter" and exclusion when specific anomaly is active', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // Filter role to 'coder'
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(roleSelect);

        // Coder has Turn 3 (read, success, no anomaly), Turn 4 (bash, error, anomaly), Turn 5 (edit, error, anomaly), Turn 10 (edit, error, anomaly)
        // With Anomaly filter at "No filter", all 4 turns pass through (including non-anomaly Turn 3)
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Now filter Anomaly to 'TOOL_RETRY_STREAK'
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(anomalySelect);

        // Non-anomaly Turn 3 and failure-rate Turn 4 are excluded
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 4')).not.toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Restore Anomaly to "No filter"
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(anomalySelect);

        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();
      });

      it('handles three-tool selection, step-down deselection, and return to "No filter"', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const toolSelect = screen.getByTestId('trace-tool-filter');
        const toolContainer = toolSelect.closest('.multiselect-container')!;

        // Select bash, edit, and read
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'read' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('3 selected');
        // Turns 3, 7 (read), Turns 4, 6 (bash), Turns 5, 10 (edit) = 6 events
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 6 of 10 events/)).toBeInTheDocument();

        // Deselect read -> 2 tools (bash, edit)
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'read' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('bash, edit');
        expect(screen.queryByText('Turn 3')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 7')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Deselect edit -> 1 tool (bash)
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('bash');
        expect(screen.queryByText('Turn 5')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 10')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 2 of 10 events/)).toBeInTheDocument();

        // Deselect bash (last option) -> reverts to ['all'] ("No filter")
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(toolSelect);

        expect(toolSelect.textContent).toContain('No filter');
        expect(screen.getByText(/Showing 10 of 10 events/)).toBeInTheDocument();
      });

      it('handles three-anomaly selection and excludes error turns lacking matching anomaly rule', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // Select 3 anomalies: TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, SUBAGENT_REWORK_LOOP
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_HIGH_FAILURE_RATE' }));
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        fireEvent.click(anomalySelect);

        expect(anomalySelect.textContent).toContain('3 selected');

        // Turns 4 (TOOL_HIGH_FAILURE_RATE), 5 (TOOL_RETRY_STREAK), 6 (SUBAGENT_REWORK_LOOP), 10 (TOOL_RETRY_STREAK)
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        // Turn 9 is an error turn but has NO matching anomaly rule -> strictly excluded
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();
      });

      it('handles three-role selection and excludes unselected roles', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const allBtn = screen.getByRole('button', { name: 'All Events' });
        fireEvent.click(allBtn);

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const roleContainer = roleSelect.closest('.multiselect-container')!;

        // Select orchestrator, coder, reviewer
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'orchestrator' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'reviewer' }));
        fireEvent.click(roleSelect);

        expect(roleSelect.textContent).toContain('3 selected');

        // Orchestrator (1, 2), Coder (3, 4, 5, 10), Reviewer (7, 8) -> 8 events
        expect(screen.getByText('Turn 1')).toBeInTheDocument();
        expect(screen.getByText('Turn 2')).toBeInTheDocument();
        expect(screen.getByText('Turn 3')).toBeInTheDocument();
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 7')).toBeInTheDocument();
        expect(screen.getByText('Turn 8')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();

        // Tester turns (6, 9) excluded
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.queryByText('Turn 9')).not.toBeInTheDocument();

        expect(screen.getByText(/Showing 8 of 10 events/)).toBeInTheDocument();
      });

      it('evaluates full 4-dimensional conjunction across filterMode, multi-role, multi-tool, and multi-anomaly', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        const toolsBtn = screen.getByRole('button', { name: 'Tool Calls Only' });
        fireEvent.click(toolsBtn);

        const roleSelect = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolSelect = screen.getByTestId('trace-tool-filter');
        const anomalySelect = screen.getByTestId('trace-anomaly-filter');

        const roleContainer = roleSelect.closest('.multiselect-container')!;
        const toolContainer = toolSelect.closest('.multiselect-container')!;
        const anomalyContainer = anomalySelect.closest('.multiselect-container')!;

        // 1. Roles: coder, tester
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'tester' }));
        fireEvent.click(roleSelect);

        // 2. Tools: bash, edit
        fireEvent.click(toolSelect);
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' }));
        fireEvent.click(within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'edit' }));
        fireEvent.click(toolSelect);

        // 3. Anomalies: TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_RETRY_STREAK' }));
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'TOOL_HIGH_FAILURE_RATE' }));
        fireEvent.click(anomalySelect);

        // Turn 4: coder, bash, tool_call, error, TOOL_HIGH_FAILURE_RATE -> matches
        // Turn 5: coder, edit, tool_call, error, TOOL_RETRY_STREAK -> matches
        // Turn 10: coder, edit, tool_call, error, TOOL_RETRY_STREAK -> matches
        // Turn 6: tester, bash, tool_call, SUBAGENT_REWORK_LOOP -> excluded by anomaly filter
        // Turn 3: coder, read -> excluded by tool filter
        // Turn 9: tester, null tool -> excluded by tool filter
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.queryByText('Turn 6')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 3 of 10 events/)).toBeInTheDocument();

        // Now add SUBAGENT_REWORK_LOOP to anomaly filter
        fireEvent.click(anomalySelect);
        fireEvent.click(within(anomalyContainer as HTMLElement).getByRole('checkbox', { name: 'SUBAGENT_REWORK_LOOP' }));
        fireEvent.click(anomalySelect);

        // Turn 6 now matches as well!
        expect(screen.getByText('Turn 4')).toBeInTheDocument();
        expect(screen.getByText('Turn 5')).toBeInTheDocument();
        expect(screen.getByText('Turn 6')).toBeInTheDocument();
        expect(screen.getByText('Turn 10')).toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();

        // Conflict producing 0 events: set role to 'orchestrator' only
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'coder' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'tester' }));
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'orchestrator' }));
        fireEvent.click(roleSelect);

        // Orchestrator has no bash or edit tools -> 0 events
        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();

        // Recovery: toggle role back to "No filter"
        fireEvent.click(roleSelect);
        fireEvent.click(within(roleContainer as HTMLElement).getByRole('checkbox', { name: 'No filter' }));
        fireEvent.click(roleSelect);

        expect(screen.queryByText('No events match the selected filters.')).not.toBeInTheDocument();
        expect(screen.getByText(/Showing 4 of 10 events/)).toBeInTheDocument();
      });

      it('closes dropdowns via keyboard Escape key and clicking outside', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        // 1. Role dropdown: open and close via Escape on trigger
        const roleTrigger = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        fireEvent.click(roleTrigger);
        expect(roleTrigger).toHaveAttribute('aria-expanded', 'true');

        fireEvent.keyDown(roleTrigger, { key: 'Escape' });
        expect(roleTrigger).toHaveAttribute('aria-expanded', 'false');

        // 2. Tool dropdown: open and close via Escape on a checkbox within the dropdown
        const toolTrigger = screen.getByTestId('trace-tool-filter');
        fireEvent.click(toolTrigger);
        expect(toolTrigger).toHaveAttribute('aria-expanded', 'true');

        const toolContainer = toolTrigger.closest('.multiselect-container')!;
        const bashCheckbox = within(toolContainer as HTMLElement).getByRole('checkbox', { name: 'bash' });
        fireEvent.keyDown(bashCheckbox, { key: 'Escape' });
        expect(toolTrigger).toHaveAttribute('aria-expanded', 'false');

        // 3. Anomaly dropdown: open and close by clicking outside
        const anomalyTrigger = screen.getByTestId('trace-anomaly-filter');
        fireEvent.click(anomalyTrigger);
        expect(anomalyTrigger).toHaveAttribute('aria-expanded', 'true');

        // Click outside on the card header
        const cardHeader = screen.getByText('Execution Trace Waterfall');
        fireEvent.mouseDown(cardHeader);
        expect(anomalyTrigger).toHaveAttribute('aria-expanded', 'false');
      });

      it('asserts zero emojis across all rendered elements and multi-select states', () => {
        const { container } = render(
          <TraceWaterfall
            events={adversarialEvents}
            anomalies={adversarialAnomalies}
          />
        );

        // Open each dropdown to ensure options are also rendered in DOM
        const roleTrigger = screen.getByRole('combobox', { name: 'Filter by subagent role' });
        const toolTrigger = screen.getByTestId('trace-tool-filter');
        const anomalyTrigger = screen.getByTestId('trace-anomaly-filter');

        fireEvent.click(roleTrigger);
        fireEvent.click(toolTrigger);
        fireEvent.click(anomalyTrigger);

        const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
        expect(emojiRegex.test(container.textContent ?? '')).toBe(false);
      });
    });
  });
});
