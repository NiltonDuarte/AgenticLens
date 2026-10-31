import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TraceWaterfall, matchesAnomaly, findMatchingAnomaly } from './TraceWaterfall';
import type { EventFact, Anomaly } from '../domain/types';

describe('FIX-12 event facts without payload hydration', () => {
  const fact: EventFact = {
    event_id: 'committed-call', timestamp: '2026-10-29T10:00:00Z', session_id: 'committed-session',
    message_id: 'committed-message', turn: 1, role: 'coder', event_type: 'tool_call',
    tool_name: 'bash', status: 'error', context_tokens: 2500,
  };

  it('keeps counts, chronological rows, identifiers, filters, and missing-details copy without hydration', () => {
    const later: EventFact = { ...fact, event_id: undefined, message_id: undefined, timestamp: '2026-10-29T11:00:00Z',
      turn: 2, status: 'completed' };
    render(<TraceWaterfall events={[later, fact]} isHydrated={false} dataSource="committed_ledger" />);
    expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('2 Events Recorded');
    expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
    fireEvent.click(screen.getByTestId('waterfall-expand-btn'));
    expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
    expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
    expect(screen.queryByTestId('timeline-item-2')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
    expect(screen.getAllByTestId(/timeline-item-/).map((item) => item.dataset.testid)).toEqual(['timeline-item-1', 'timeline-item-2']);
    fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 1' }));
    expect(screen.getByTestId('payload-details-1')).toHaveTextContent('committed-message');
    expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent('Payload details unavailable');
    expect(screen.getByRole('button', { name: 'Copy Message ID' })).toBeInTheDocument();
    expect(screen.queryByTestId('tool-output-payload-block')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Show only errors and anomalies' }));
    expect(screen.queryByTestId('timeline-item-2')).not.toBeInTheDocument();
  });

  it.each(['local_sqlite', 'local_codex'] as const)('retains %s payloads alongside a bare committed row', (source) => {
    const local = { ...fact, tool_input: 'make test', tool_output: 'error detail', prompt_content: 'Run checks', reasoning_content: 'Inspect build' };
    const bare = { ...fact, event_id: 'unmatched', turn: 2 };
    render(<TraceWaterfall events={[local, bare]} isHydrated={true} dataSource={source} defaultExpanded={true} />);
    expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent(source === 'local_codex' ? 'Source: Local Codex' : 'Source: Local SQLite');
    fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 1' }));
    expect(screen.getByTestId('tool-input-payload-block')).toHaveTextContent('make test');
    expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('error detail');
    expect(screen.getByTestId('prompt-payload-block')).toHaveTextContent('Run checks');
    expect(screen.getByTestId('reasoning-payload-block')).toHaveTextContent('Inspect build');
    fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 2' }));
    expect(within(screen.getByTestId('payload-details-2')).getByTestId('unhydrated-notice')).toHaveTextContent('Payload details unavailable');
    expect(within(screen.getByTestId('payload-details-1')).queryByTestId('unhydrated-notice')).not.toBeInTheDocument();
  });

  it('shows payloads actually supplied by an unhydrated response', () => {
    render(<TraceWaterfall events={[{ ...fact, tool_output: 'Recorded response detail' }]} isHydrated={false} defaultExpanded={true} />);
    fireEvent.click(screen.getByRole('button', { name: 'View payload for turn 1' }));
    expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('Recorded response detail');
    expect(screen.queryByTestId('unhydrated-notice')).not.toBeInTheDocument();
  });

  it.each(['stable ID', 'no IDs', 'new local ID'] as const)('rebinds an uncontrolled %s selection when its event receives payloads', (identity) => {
    const baseline: EventFact = {
      ...fact, event_id: identity === 'stable ID' ? fact.event_id : undefined, message_id: undefined,
    };
    const selected = vi.fn();
    const { rerender } = render(<TraceWaterfall events={[baseline]} defaultExpanded={true} onSelectEvent={selected} />);
    fireEvent.click(screen.getByTestId('timeline-item-1'));
    expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
    expect(screen.getByTestId('unhydrated-notice')).toBeInTheDocument();
    const enriched: EventFact = {
      ...baseline, event_id: identity === 'new local ID' ? 'local-call' : baseline.event_id,
      message_id: 'local-message', tool_output: 'new local payload',
    };
    rerender(<TraceWaterfall events={[enriched]} defaultExpanded={true} onSelectEvent={selected}
      isHydrated={true} dataSource="local_sqlite" />);
    expect(screen.getByTestId('timeline-item-1')).toHaveClass('selected');
    expect(screen.getByRole('button', { name: 'Hide payload for turn 1' })).toBeInTheDocument();
    expect(screen.getByTestId('payload-details-1')).toHaveTextContent('local-message');
    expect(screen.getByTestId('tool-output-payload-block')).toHaveTextContent('new local payload');
    expect(screen.queryByTestId('unhydrated-notice')).not.toBeInTheDocument();
    expect(selected).toHaveBeenCalledTimes(1);
  });
});

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
      context: {
        tool_name: 'edit',
        turn: 3,
        role: 'coder',
        count: 2,
      },
    },
  ];

  it('renders localized event time with full tooltip context and preserves invalid timestamp text', () => {
    const timestampEvent: EventFact = { ...sampleEvents[0]!, event_type: 'tool_call', tool_name: 'bash', status: 'error' };
    const { rerender } = render(<TraceWaterfall events={[timestampEvent]} defaultExpanded={true} />);
    const time = screen.getByTestId('timeline-item-1').querySelector('.timeline-timestamp');
    expect(time).toHaveTextContent(/\d{2}:\d{2}:\d{2}\.\d{3}/);
    expect(time).toHaveAttribute('title', expect.stringMatching(/\d{4}/));

    rerender(<TraceWaterfall events={[{ ...timestampEvent, timestamp: 'not-a-date' }]} defaultExpanded={true} />);
    expect(screen.getByTestId('timeline-item-1').querySelector('.timeline-timestamp')).toHaveTextContent('not-a-date');
  });

  it.each([undefined, []])('renders an immediately visible empty card without anomalies (%s)', (anomalies) => {
    render(<TraceWaterfall events={[]} anomalies={anomalies} />);
    expect(
      screen.getByText('Zero chronological events recorded for this execution trace.')
    ).toBeVisible();
    expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded');
    expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent('Source: Committed Ledger');
    expect(screen.queryByTestId('trace-waterfall-collapsed')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Trace Waterfall/ })).not.toBeInTheDocument();
    expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
    expect(screen.queryByRole('group', { name: 'Filter events by type' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
  });

  describe('summary-only anomaly disclosure', () => {
    it.each([
      ['committed_ledger', false, 'Source: Committed Ledger'],
      ['local_sqlite', true, 'Source: Local SQLite'],
      ['local_codex', true, 'Source: Local Codex'],
    ] as const)('keeps %s summaries and truthful counts inside the disclosure across reopening', async (dataSource, isHydrated, label) => {
      const user = userEvent.setup();
      const onSelectEvent = vi.fn();
      const onSelectAnomaly = vi.fn();
      const anomalies: readonly Anomaly[] = [...sampleAnomalies, {
        rule_id: 'CONTEXT_SATURATION_SPIKE', severity: 'CRITICAL',
        context: { turn: 8, token_delta: 50000, count: 50 },
      }];
      render(<TraceWaterfall events={[]} anomalies={anomalies} dataSource={dataSource} isHydrated={isHydrated}
        onSelectEvent={onSelectEvent} onSelectAnomaly={onSelectAnomaly} />);

      const toggle = screen.getByRole('button', { name: 'Expand Trace Waterfall' });
      const bodyId = toggle.getAttribute('aria-controls');
      expect(bodyId).toBeTruthy();
      const body = document.getElementById(bodyId!);
      expect(body).toBeInTheDocument();
      expect(body).not.toBeVisible();
      expect(toggle).toHaveAttribute('type', 'button');
      expect(toggle).toHaveAttribute('aria-expanded', 'false');
      expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('0 Events Recorded');
      expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('Zero Errors, 2 Anomalies');
      expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent(label);
      expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
      expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();
      expect(screen.getAllByRole('button')).toEqual([toggle]);

      await user.tab();
      expect(toggle).toHaveFocus();
      await user.keyboard('{Enter}');
      const card = screen.getByTestId('trace-waterfall-card');
      expect(screen.getByRole('button', { name: 'Collapse Trace Waterfall' })).toBe(toggle);
      expect(toggle).toHaveFocus();
      expect(toggle).toHaveAttribute('aria-expanded', 'true');
      expect(body).toBeVisible();
      expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
      expect(body).toContainElement(screen.getByTestId('workspace-anomalies'));
      expect(within(card).getByText('Zero chronological events recorded for this execution trace.')).toBeVisible();
      expect(within(card).getAllByRole('button', { name: /^Inspect anomaly:/ })).toHaveLength(2);
      expect(within(card).getByText('Detected Workspace Anomalies:')).toBeVisible();
      expect(screen.getByTestId('waterfall-source-badge')).toHaveTextContent(label);
      expect(screen.queryByRole('group', { name: 'Filter events by type' })).not.toBeInTheDocument();
      expect(screen.queryByTestId('trace-role-filter')).not.toBeInTheDocument();
      expect(screen.queryByTestId('trace-tool-filter')).not.toBeInTheDocument();
      expect(screen.queryByTestId('trace-anomaly-filter')).not.toBeInTheDocument();
      expect(screen.queryByTestId(/timeline-item-/)).not.toBeInTheDocument();
      expect(screen.queryByTestId('db-offline-banner')).not.toBeInTheDocument();
      expect(screen.queryByTestId('unhydrated-notice')).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /payload for turn/ })).not.toBeInTheDocument();

      await user.keyboard(' ');
      expect(toggle).toHaveFocus();
      expect(toggle).toHaveAttribute('aria-expanded', 'false');
      expect(toggle).toHaveAttribute('aria-controls', bodyId);
      expect(document.getElementById(bodyId!)).toBe(body);
      expect(body).not.toBeVisible();
      expect(screen.getAllByRole('button')).toEqual([toggle]);
      expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /^Inspect anomaly:/ })).not.toBeInTheDocument();
      expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();

      await user.click(toggle);
      expect(screen.getAllByTestId('workspace-anomalies')).toHaveLength(1);
      expect(card).toContainElement(screen.getByTestId('workspace-anomalies'));
      expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('Zero Errors, 2 Anomalies');
      expect(onSelectEvent).not.toHaveBeenCalled();
      expect(onSelectAnomaly).not.toHaveBeenCalled();
    });

    it('honors defaultExpanded and safely activates native anomaly buttons without a callback', async () => {
      const user = userEvent.setup();
      render(<TraceWaterfall events={[]} anomalies={sampleAnomalies} defaultExpanded={true} />);
      const toggle = screen.getByRole('button', { name: 'Collapse Trace Waterfall' });
      const anomaly = screen.getByRole('button', { name: /^Inspect anomaly: Tool retry streak/ });
      expect(anomaly.tagName).toBe('BUTTON');
      expect(anomaly).toHaveAttribute('type', 'button');
      expect(screen.getByText('Zero chronological events recorded for this execution trace.')).toBeVisible();
      await user.click(anomaly);
      await user.keyboard('{Enter}');
      await user.keyboard(' ');
      expect(toggle).toHaveAttribute('aria-expanded', 'true');
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
      await user.click(toggle);
      expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
    });

    it('forwards each click, Enter, and Space activation exactly once without selecting an event', async () => {
      const user = userEvent.setup();
      const onSelectAnomaly = vi.fn();
      const onSelectEvent = vi.fn();
      render(<TraceWaterfall events={[]} anomalies={sampleAnomalies} defaultExpanded={true}
        onSelectAnomaly={onSelectAnomaly} onSelectEvent={onSelectEvent} />);
      const anomaly = screen.getByRole('button', { name: /^Inspect anomaly: Tool retry streak/ });
      await user.click(anomaly);
      expect(onSelectAnomaly).toHaveBeenCalledTimes(1);
      await user.keyboard('{Enter}');
      expect(onSelectAnomaly).toHaveBeenCalledTimes(2);
      await user.keyboard(' ');
      expect(onSelectAnomaly).toHaveBeenCalledTimes(3);
      expect(onSelectAnomaly.mock.calls).toEqual([[sampleAnomalies[0]], [sampleAnomalies[0]], [sampleAnomalies[0]]]);
      expect(onSelectEvent).not.toHaveBeenCalled();
      expect(screen.getByRole('button', { name: 'Collapse Trace Waterfall' })).toHaveAttribute('aria-expanded', 'true');
    });

    it('associates each disclosure with its own body', () => {
      render(<>
        <TraceWaterfall events={[]} anomalies={sampleAnomalies} />
        <TraceWaterfall events={[]} anomalies={sampleAnomalies} />
      </>);
      const targets = screen.getAllByRole('button', { name: 'Expand Trace Waterfall' })
        .map((toggle) => toggle.getAttribute('aria-controls'));
      expect(new Set(targets).size).toBe(2);
      for (const target of targets) {
        expect(target).toBeTruthy();
        expect(document.getElementById(target!)).toBeInTheDocument();
      }
    });
  });

  it('renders collapsed summary bar by default with counts and expands when clicked', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);
    expect(screen.getByTestId('trace-waterfall-collapsed')).toBeInTheDocument();
    expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
    expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('4 Events Recorded');
    expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('1 Error, 1 Anomaly');

    const expandBtn = screen.getByTestId('waterfall-expand-btn');
    expect(expandBtn).toHaveAttribute('aria-expanded', 'false');
    expect(expandBtn).toHaveTextContent('Expand Trace Waterfall');

    // Click to expand
    fireEvent.click(expandBtn);
    expect(screen.queryByTestId('trace-waterfall-collapsed')).not.toBeInTheDocument();
    expect(screen.getByTestId('waterfall-collapse-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /^Inspect anomaly:/ })).not.toBeInTheDocument();
  });

  it('renders clean execution badge in collapsed view when no errors or anomalies exist', () => {
    const cleanEvents: readonly EventFact[] = [
      {
        timestamp: '2026-10-18T10:00:00Z',
        session_id: 's1',
        turn: 1,
        role: 'coder',
        event_type: 'turn_message',
        status: 'success',
        context_tokens: 1000,
      },
    ];
    render(<TraceWaterfall events={cleanEvents} anomalies={[]} />);
    expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('Clean Execution');
  });

  it('defaults to Errors and Anomalies Only filter mode when expanded', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} defaultExpanded={true} />);
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
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getByText('Turn 1')).toBeInTheDocument();
    expect(screen.getByText('Turn 2')).toBeInTheDocument();
    expect(screen.getByText('Turn 3')).toBeInTheDocument();
    expect(screen.getByText('Turn 4')).toBeInTheDocument();
  });

  it('displays role pills with proper role names when all events are shown', () => {
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getAllByText('orchestrator').length).toBeGreaterThan(0);
    expect(screen.getAllByText('coder').length).toBeGreaterThan(0);
    expect(screen.getAllByText('tester').length).toBeGreaterThan(0);
  });

  it('displays tool name, formatted duration, and status indicator for tool calls', () => {
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
    const allBtn = screen.getByRole('button', { name: 'All Events' });
    fireEvent.click(allBtn);

    expect(screen.getAllByText('edit').length).toBeGreaterThan(0);
    expect(screen.getByText('450ms')).toBeInTheDocument();
    expect(screen.getByText('1.25s')).toBeInTheDocument();
  });

  it('correlates anomaly by turn/tool name and renders AnomalyBadge on corresponding turn item', () => {
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} defaultExpanded={true} />);
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
        defaultExpanded={true}
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
        defaultExpanded={true}
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
        defaultExpanded={true}
      />
    );

    const badge = screen.getByRole('button', { name: 'WARNING' });
    expect(badge).toHaveAttribute('title', 'Click to open Anomaly Inspector');
    fireEvent.click(badge);

    expect(handleSelect).toHaveBeenCalledTimes(1);
    expect(handleSelect).toHaveBeenCalledWith(
      expect.objectContaining({ turn: 3, tool_name: 'edit' }),
      expect.objectContaining({ rule_id: 'TOOL_RETRY_STREAK' })
    );
  });

  it('filters events by role when role filter is changed', () => {
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
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
    render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} defaultExpanded={true} />);
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
        defaultExpanded={true}
      />
    );

    const timelineItem = screen.getByTestId('timeline-item-3');
    expect(timelineItem).toHaveClass('selected');
  });

  it('provides standalone local selection highlight when selectedEvent prop is undefined', () => {
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
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

    render(<TraceWaterfall events={[minimalEvent]} defaultExpanded={true} />);
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
        context: {},
      },
      {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'WARNING',
        context: {
          turn: 1,
        },
      },
    ];

    render(<TraceWaterfall events={events} anomalies={anomaliesWithMissingContext} defaultExpanded={true} />);
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
        defaultExpanded={true}
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
        context: {
          turn: 4,
          role: 'tester',
        },
      },
    ];

    render(<TraceWaterfall events={sampleEvents} anomalies={anomalyOnlyList} defaultExpanded={true} />);
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
        defaultExpanded={true}
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

  it('renders guidance hint in waterfall toolbar header', () => {
    render(<TraceWaterfall events={sampleEvents} defaultExpanded={true} />);
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
          defaultExpanded={true}
        />
      );

      const banner = screen.getByTestId('db-offline-banner');
      expect(banner).toBeInTheDocument();
      expect(
        screen.getByText('Committed Ledger Facts')
      ).toBeInTheDocument();
      expect(
        screen.getByText(
          /Recorded event facts remain visible from the committed ledger/i
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
          defaultExpanded={true}
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
          defaultExpanded={true}
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
          defaultExpanded={true}
        />
      );

      // Turn 3 is visible under default filter
      const viewPayloadBtn = screen.getByRole('button', { name: 'View payload for turn 3' });
      fireEvent.click(viewPayloadBtn);

      expect(screen.getByTestId('payload-details-3')).toBeInTheDocument();
      expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent(
        'Payload details unavailable for this event.'
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
          defaultExpanded={true}
        />
      );

      const allBtn = screen.getByRole('button', { name: 'All Events' });
      fireEvent.click(allBtn);

      const viewPayloadBtn = screen.getByRole('button', { name: 'View payload for turn 1' });
      fireEvent.click(viewPayloadBtn);

      expect(screen.getByTestId('payload-details-1')).toBeInTheDocument();
      expect(screen.getByTestId('unhydrated-notice')).toHaveTextContent('Payload details unavailable for this event.');
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
        context: {
          tool_name: 'edit',
          turn: 5,
          role: 'coder',
        },
      },
      {
        rule_id: 'CONTEXT_SATURATION_SPIKE',
        severity: 'CRITICAL',
        context: {
          turn: 5,
          session_id: 's-adv',
        },
      },
      {
        rule_id: 'TOOL_HIGH_FAILURE_RATE',
        severity: 'WARNING',
        context: {
          tool_name: 'bash',
          turn: 4,
          role: 'coder',
        },
      },
      {
        rule_id: 'SUBAGENT_REWORK_LOOP',
        severity: 'CRITICAL',
        context: {
          role: 'tester',
          turn: 6,
        },
      },
      {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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

        render(<TraceWaterfall events={nonToolEvents} anomalies={[]} defaultExpanded={true} />);

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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
        const { rerender } = render(<TraceWaterfall events={adversarialEvents} anomalies={[]} defaultExpanded={true} />);

        let anomalySelect = screen.getByTestId('trace-anomaly-filter');
        expect(anomalySelect).toHaveTextContent('No filter');
        fireEvent.click(anomalySelect);
        let container = anomalySelect.closest('.multiselect-container')!;
        let options = within(container as HTMLElement)
          .getAllByRole('checkbox')
          .map((cb) => cb.getAttribute('aria-label'));
        expect(options).toEqual(['No filter']);

        // Rerender with undefined anomalies
        rerender(<TraceWaterfall events={adversarialEvents} anomalies={undefined} defaultExpanded={true} />);
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
            context: {
              turn: 999,
            },
          },
        ];

        render(
          <TraceWaterfall
            events={adversarialEvents}
            defaultExpanded={true}
            anomalies={disjointAnomalies}
          />
        );

        const anomalySelect = screen.getByTestId('trace-anomaly-filter');
        const container = anomalySelect.closest('.multiselect-container')!;
        fireEvent.click(anomalySelect);
        fireEvent.click(within(container as HTMLElement).getByRole('checkbox', { name: 'CONTEXT_SATURATION_SPIKE' }));

        expect(screen.getByText('No events match the selected filters.')).toBeInTheDocument();
        expect(screen.getByText(/Showing 0 of 10 events/)).toBeInTheDocument();
        expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /^Inspect anomaly:/ })).not.toBeInTheDocument();
        expect(screen.queryByText('Zero chronological events recorded for this execution trace.')).not.toBeInTheDocument();

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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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

        render(<TraceWaterfall events={singleEvent} anomalies={[]} defaultExpanded={true} />);

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
            defaultExpanded={true}
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

    describe('Multi-Select & "No filter" Enhancements', () => {
      it('initial mount displays "No filter" on Role, Tool, and Anomaly triggers and passes through all events in All Events mode', () => {
        render(
          <TraceWaterfall
            events={adversarialEvents}
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
            defaultExpanded={true}
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
    });
  });

  describe('FIX-3 Verification: Collapsibility, Anomaly Matching, and Debug Identifiers', () => {
    it('collapsible by default: initial mount renders only summary bar and expand button', () => {
      render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies} />);

      // Summary bar is present
      expect(screen.getByTestId('trace-waterfall-collapsed')).toBeInTheDocument();
      expect(screen.getByText('Execution Trace Waterfall')).toBeInTheDocument();
      expect(screen.getByTestId('waterfall-event-count')).toHaveTextContent('4 Events Recorded');
      expect(screen.getByTestId('waterfall-issues-summary')).toHaveTextContent('1 Error, 1 Anomaly');

      // Expand button is present
      const expandBtn = screen.getByTestId('waterfall-expand-btn');
      expect(expandBtn).toBeInTheDocument();
      expect(expandBtn).toHaveAttribute('aria-expanded', 'false');

      // Filters and timeline elements must NOT be rendered while collapsed
      expect(screen.queryByRole('button', { name: 'All Events' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Tool Calls Only' })).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'Show only errors and anomalies' })).not.toBeInTheDocument();
      expect(screen.queryByRole('combobox', { name: 'Filter by subagent role' })).not.toBeInTheDocument();
      expect(screen.queryByTestId('trace-tool-filter')).not.toBeInTheDocument();
      expect(screen.queryByTestId('trace-anomaly-filter')).not.toBeInTheDocument();
      expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
    });

    it('expanding reveals filters and timeline; collapsing preserves filters, selection, and payload expansion', async () => {
      const user = userEvent.setup();
      const onSelectEvent = vi.fn();
      const onSelectAnomaly = vi.fn();
      render(<TraceWaterfall events={sampleEvents} anomalies={sampleAnomalies}
        onSelectEvent={onSelectEvent} onSelectAnomaly={onSelectAnomaly} />);

      // 1. Initially collapsed
      expect(screen.getByTestId('trace-waterfall-collapsed')).toBeInTheDocument();
      expect(screen.queryByTestId('waterfall-collapse-btn')).not.toBeInTheDocument();

      // 2. Expand
      const expandBtn = screen.getByTestId('waterfall-expand-btn');
      await user.click(expandBtn);

      // Collapsed summary gone; full waterfall revealed
      expect(screen.queryByTestId('trace-waterfall-collapsed')).not.toBeInTheDocument();
      const collapseBtn = screen.getByTestId('waterfall-collapse-btn');
      expect(collapseBtn).toBeInTheDocument();
      expect(collapseBtn).toHaveAttribute('aria-expanded', 'true');

      // Filter toolbar is visible
      expect(screen.getByRole('button', { name: 'All Events' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Show only errors and anomalies' })).toBeInTheDocument();

      // Switch to All Events to reveal all timeline turns
      await user.click(screen.getByRole('button', { name: 'All Events' }));
      expect(screen.getByText('Turn 1')).toBeInTheDocument();
      expect(screen.getByText('Turn 2')).toBeInTheDocument();
      await user.click(screen.getByRole('button', { name: 'View payload for turn 2' }));
      await user.click(screen.getByTestId('timeline-item-3'));
      expect(screen.getByTestId('payload-details-2')).toBeInTheDocument();
      expect(screen.getByTestId('timeline-item-3')).toHaveClass('selected');
      expect(onSelectEvent).toHaveBeenCalledTimes(1);

      // 3. Collapse again
      await user.click(collapseBtn);

      // Full view hidden; collapsed summary restored
      expect(screen.getByTestId('trace-waterfall-collapsed')).toBeInTheDocument();
      expect(screen.queryByTestId('waterfall-collapse-btn')).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: 'All Events' })).not.toBeInTheDocument();
      expect(screen.queryByText('Turn 1')).not.toBeInTheDocument();
      expect(screen.queryByRole('button', { name: /payload for turn/ })).not.toBeInTheDocument();
      expect(screen.queryByTestId('payload-details-2')).not.toBeInTheDocument();

      await user.click(screen.getByTestId('waterfall-expand-btn'));
      expect(screen.getByRole('button', { name: 'All Events' })).toHaveClass('active');
      expect(screen.getAllByTestId(/timeline-item-/)).toHaveLength(4);
      expect(screen.getByTestId('timeline-item-3')).toHaveClass('selected');
      expect(screen.getByTestId('payload-details-2')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Hide payload for turn 2' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Hide payload for turn 3' })).toBeInTheDocument();
      expect(screen.queryByTestId('workspace-anomalies')).not.toBeInTheDocument();
      expect(onSelectEvent).toHaveBeenCalledTimes(1);
      expect(onSelectAnomaly).not.toHaveBeenCalled();
    });

    it('anomaly matching: TOOL_RETRY_STREAK must NOT match a completed event', () => {
      const streakAnomaly: Anomaly = {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'WARNING',
        context: {
          tool_name: 'read',
          turn: 5,
          session_id: 'ses-1',
        },
      };

      const completedEvent: EventFact = {
        timestamp: '2026-10-18T10:05:00Z',
        session_id: 'ses-1',
        turn: 5,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'read',
        status: 'completed',
        context_tokens: 2000,
      };

      const successEvent: EventFact = {
        ...completedEvent,
        status: 'success',
      };

      const errorEvent: EventFact = {
        ...completedEvent,
        status: 'error',
        error_message: 'file not found',
      };

      // Pure matching function assertions
      expect(matchesAnomaly(completedEvent, streakAnomaly)).toBe(false);
      expect(matchesAnomaly(successEvent, streakAnomaly)).toBe(false);
      expect(matchesAnomaly(errorEvent, streakAnomaly)).toBe(true);

      expect(findMatchingAnomaly(completedEvent, [streakAnomaly])).toBeUndefined();
      expect(findMatchingAnomaly(successEvent, [streakAnomaly])).toBeUndefined();
      expect(findMatchingAnomaly(errorEvent, [streakAnomaly])).toEqual(streakAnomaly);

      // Component rendering assertion: completed event does NOT get AnomalyBadge
      const { container: container1, unmount: unmount1 } = render(
        <TraceWaterfall
          events={[completedEvent]}
          anomalies={[streakAnomaly]}
          defaultExpanded={true}
          isHydrated={true}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
      expect(screen.getByText('Turn 5')).toBeInTheDocument();
      expect(container1.querySelector('.anomaly-badge')).not.toBeInTheDocument();
      unmount1();

      // Rendering with error event DOES show AnomalyBadge
      const { container: container2 } = render(
        <TraceWaterfall
          events={[errorEvent]}
          anomalies={[streakAnomaly]}
          defaultExpanded={true}
          isHydrated={true}
        />
      );

      expect(screen.getByText('Turn 5')).toBeInTheDocument();
      const badge = container2.querySelector('.anomaly-badge');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveTextContent('WARNING');
    });

    it('anomaly matching: anomaly with session_id must NOT match event with different session_id even if turn is equal', () => {
      const anomalySessionA: Anomaly = {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'CRITICAL',
        context: {
          session_id: 'ses-A',
          tool_name: 'bash',
          turn: 7,
        },
      };

      const eventSessionB: EventFact = {
        timestamp: '2026-10-18T10:07:00Z',
        session_id: 'ses-B',
        turn: 7,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'error',
        error_message: 'exit status 1',
        context_tokens: 5000,
      };

      const eventSessionA: EventFact = {
        ...eventSessionB,
        session_id: 'ses-A',
      };

      // Pure function assertions
      expect(matchesAnomaly(eventSessionB, anomalySessionA)).toBe(false);
      expect(matchesAnomaly(eventSessionA, anomalySessionA)).toBe(true);

      expect(findMatchingAnomaly(eventSessionB, [anomalySessionA])).toBeUndefined();
      expect(findMatchingAnomaly(eventSessionA, [anomalySessionA])).toEqual(anomalySessionA);

      // Component rendering assertion: event in session B does NOT receive anomaly badge
      const { container: containerA, unmount: unmountA } = render(
        <TraceWaterfall
          events={[eventSessionB]}
          anomalies={[anomalySessionA]}
          defaultExpanded={true}
          isHydrated={true}
        />
      );

      // In errors_anomalies mode, eventSessionB has error, so it shows up, but must NOT have anomaly badge
      expect(screen.getByText('Turn 7')).toBeInTheDocument();
      expect(containerA.querySelector('.anomaly-badge')).not.toBeInTheDocument();
      unmountA();

      // Event in session A DOES receive anomaly badge
      const { container: containerB } = render(
        <TraceWaterfall
          events={[eventSessionA]}
          anomalies={[anomalySessionA]}
          defaultExpanded={true}
          isHydrated={true}
        />
      );

      expect(screen.getByText('Turn 7')).toBeInTheDocument();
      const badge = containerB.querySelector('.anomaly-badge');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveTextContent('CRITICAL');
    });

    it('copy ID button copies debug identifier and shows Copied feedback', async () => {
      vi.useFakeTimers();
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        configurable: true,
        writable: true,
        value: {
          writeText: writeTextMock,
        },
      });

      const eventWithMsgID: EventFact = {
        timestamp: '2026-10-18T10:00:00Z',
        session_id: 'ses-uuid-12345678',
        message_id: 'msg-uuid-87654321',
        turn: 1,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'success',
        duration_ms: 100,
        context_tokens: 1200,
      };

      render(
        <TraceWaterfall
          events={[eventWithMsgID]}
          defaultExpanded={true}
        />
      );

      // Switch to All Events to view the turn
      fireEvent.click(screen.getByRole('button', { name: 'All Events' }));

      // Find the Copy Prompt button for turn 1
      const copyBtn = screen.getByRole('button', { name: 'Copy debug prompt for turn 1' });
      expect(copyBtn).toBeInTheDocument();
      expect(copyBtn).toHaveTextContent('Copy Prompt');

      // Click Copy Prompt
      await act(async () => {
        fireEvent.click(copyBtn);
      });

      // Verifies clipboard was invoked with Layer 1 debug prompt
      expect(writeTextMock).toHaveBeenCalledTimes(1);
      const copiedText = writeTextMock.mock.calls[0][0];
      expect(copiedText).toContain('# Tool Execution: `bash`');
      expect(copiedText).toContain(
        'During turn 1 of session `ses-uuid-12345678` (message `msg-uuid-87654321`), the agent executed `bash` and it completed successfully.'
      );

      // State changes to "Copied"
      expect(copyBtn).toHaveTextContent('Copied');

      // Fast-forward 2000ms
      act(() => {
        vi.advanceTimersByTime(2000);
      });

      // Reverts back to "Copy Prompt"
      expect(copyBtn).toHaveTextContent('Copy Prompt');

      vi.useRealTimers();
    });

    it('copies Layer 3 message turn prompt when non-tool_call event is copied', async () => {
      vi.useFakeTimers();
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        configurable: true,
        writable: true,
        value: {
          writeText: writeTextMock,
        },
      });

      const messageEvent: EventFact = {
        timestamp: '2026-10-18T10:05:00Z',
        session_id: 'ses-msg-turn',
        message_id: 'msg-uuid-turn-9',
        turn: 2,
        role: 'assistant',
        event_type: 'turn_message',
        status: 'success',
        context_tokens: 2500,
        prompt_content: 'Let us start the execution plan.',
        reasoning_content: 'Analyzing requirements carefully.',
      };

      render(
        <TraceWaterfall
          events={[messageEvent]}
          defaultExpanded={true}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: 'All Events' }));

      const copyBtn = screen.getByRole('button', { name: 'Copy debug prompt for turn 2' });
      expect(copyBtn).toBeInTheDocument();

      await act(async () => {
        fireEvent.click(copyBtn);
      });

      expect(writeTextMock).toHaveBeenCalledTimes(1);
      const copiedText = writeTextMock.mock.calls[0][0];
      expect(copiedText).toContain('# Turn 2 Review (`assistant`)');
      expect(copiedText).toContain('In turn 2 of session `ses-msg-turn` (message `msg-uuid-turn-9`)');
      expect(copiedText).toContain('```\nAnalyzing requirements carefully.\n```');
      expect(copiedText).toContain('```\nLet us start the execution plan.\n```');

      vi.useRealTimers();
    });

    it('copies Layer 1 prompt with streak details for tool_call events with associated retry anomaly', async () => {
      vi.useFakeTimers();
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        configurable: true,
        writable: true,
        value: {
          writeText: writeTextMock,
        },
      });

      const failedToolEvent: EventFact = {
        timestamp: '2026-10-18T10:10:00Z',
        session_id: 'ses-wf-streak',
        message_id: 'msg-wf-88',
        turn: 4,
        role: 'coder',
        event_type: 'tool_call',
        tool_name: 'bash',
        status: 'error',
        tool_input: 'make test',
        tool_output: 'exit status 2',
        error_message: 'process exited with error',
        context_tokens: 3500,
      };

      const streakAnomaly: Anomaly = {
        rule_id: 'TOOL_RETRY_STREAK',
        severity: 'CRITICAL',
        context: {
          session_id: 'ses-wf-streak',
          tool_name: 'bash',
          turn: 4,
          count: 5,
        },
      };

      render(
        <TraceWaterfall
          events={[failedToolEvent]}
          anomalies={[streakAnomaly]}
          defaultExpanded={true}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: 'All Events' }));

      const copyBtn = screen.getByRole('button', { name: 'Copy debug prompt for turn 4' });
      expect(copyBtn).toBeInTheDocument();
      expect(copyBtn).toHaveTextContent('Copy Prompt');

      await act(async () => {
        fireEvent.click(copyBtn);
      });

      expect(writeTextMock).toHaveBeenCalledTimes(1);
      const copiedText = writeTextMock.mock.calls[0][0];
      expect(copiedText).toContain('# Tool Execution Failure: `bash`');
      expect(copiedText).toContain(
        'During turn 4 of session `ses-wf-streak` (message `msg-wf-88`), the agent executed `bash` and it failed with 5 consecutive failures on this tool.'
      );
      expect(copiedText).toContain('make test');
      expect(copiedText).toContain('exit status 2');
      expect(copyBtn).toHaveTextContent('Copied');

      act(() => {
        vi.advanceTimersByTime(2000);
      });

      expect(copyBtn).toHaveTextContent('Copy Prompt');

      vi.useRealTimers();
    });
  });
});

describe('semantic diagnostics', () => {
  it('keeps completed content warnings visible in the default errors filter with independent badges', () => {
    const base: EventFact = {timestamp:'2026-10-31T10:00:00Z',session_id:'s',turn:1,role:'coder',event_type:'tool_call',tool_name:'bash',status:'completed',context_tokens:0};
    render(<TraceWaterfall events={[{...base,has_content_error:true,content_severity:'medium',content_error_snippet:'syntax error: unexpected token'}, {...base,turn:2,status:'error'}, {...base,turn:3}]} defaultExpanded={true} />);
    expect(screen.getByText('CONTENT ERROR')).toBeInTheDocument();
    expect(screen.getByText('EXEC ERROR')).toBeInTheDocument();
    expect(screen.getByTestId('timeline-item-1')).toBeInTheDocument();
    expect(screen.queryByTestId('timeline-item-3')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', {name:'All Events'}));
    expect(screen.getByTestId('timeline-item-3')).toBeInTheDocument();
  });
});

describe('rich invocation badges', () => {
  const event: EventFact = {
    timestamp: '2026-10-30T10:00:00Z', session_id: 'parent', turn: 1, role: 'coder',
    event_type: 'tool_call', tool_name: 'task', status: 'completed', context_tokens: 100,
  };
  const task = { timestamp: event.timestamp, parent_turn: 1, subagent_type: 'tester', description: 'Verify', child_session_id: 'child' };

  it('renders committed badges and activates a child without selecting the parent event', async () => {
    const selectChild = vi.fn();
    const selectEvent = vi.fn();
    render(<TraceWaterfall events={[event, { ...event, turn: 2, tool_name: 'skill' }]}
      skills={[{ name: 'review', turn_index: 2, message_index: 0, caller_role: 'coder', timestamp: event.timestamp, success: true }]}
      taskLinks={[task]} availableSessionIds={['child']} onSelectSession={selectChild}
      onSelectEvent={selectEvent} defaultExpanded />);
    fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
    expect(screen.getByText('[skill: review]')).toHaveAttribute('title', '[skill: review]');
    const child = screen.getByRole('button', { name: 'Inspect child session child' });
    child.focus();
    await userEvent.keyboard('{Enter}');
    expect(selectChild).toHaveBeenCalledWith('child');
    expect(selectEvent).not.toHaveBeenCalled();
  });

  it('does not attribute an ambiguous task link across sessions and retains input-only badges', () => {
    render(<TraceWaterfall events={[event, { ...event, session_id: 'other', tool_input: '{"subagent_type":"planner"}' }]}
      taskLinks={[task]} availableSessionIds={['child']} onSelectSession={vi.fn()} defaultExpanded />);
    fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
    expect(screen.queryByRole('button', { name: 'Inspect child session child' })).not.toBeInTheDocument();
    expect(screen.getByText('[task -> planner]')).toBeInTheDocument();
    expect(screen.queryByText('[task -> tester]')).not.toBeInTheDocument();
  });
});


describe('retrospective invocation metadata safety', () => {
  const event: EventFact = { timestamp: '2026-10-30T10:00:00Z', session_id: 'parent', turn: 4,
    role: 'coder', event_type: 'tool_call', tool_name: 'task', status: 'completed', context_tokens: 100 };
  const task = { timestamp: event.timestamp, parent_turn: 4, subagent_type: 'tester', description: 'Verify', child_session_id: 'child' };

  it.each(['duplicate facts', 'conflicting input', 'different timestamp'])('rejects an unsafe child link for %s', (mode) => {
    const input = mode === 'conflicting input' ? '{"subagent_type":"planner"}' : '{broken-json';
    const tasks = mode === 'duplicate facts' ? [task, { ...task, child_session_id: 'other-child' }]
      : mode === 'different timestamp' ? [{ ...task, timestamp: '2026-10-30T11:00:00Z' }] : [task];
    render(<TraceWaterfall events={[{ ...event, tool_input: input }]} taskLinks={tasks}
      availableSessionIds={['child', 'other-child']} onSelectSession={vi.fn()} defaultExpanded />);
    fireEvent.click(screen.getByRole('button', { name: 'All Events' }));
    expect(screen.queryByRole('button', { name: /Inspect child session/ })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Copy child session ID/ })).not.toBeInTheDocument();
    if (mode === 'conflicting input') expect(screen.getByText('[task -> planner]')).toBeInTheDocument();
    else expect(screen.queryByText('[task -> tester]')).not.toBeInTheDocument();

  });
});
