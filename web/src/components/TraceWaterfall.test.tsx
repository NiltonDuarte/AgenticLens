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

    expect(screen.getByText('edit')).toBeInTheDocument();
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
});
