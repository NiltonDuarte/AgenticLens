import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { LiveSessionTree } from './LiveSessionTree';
import type { LiveSessionTree as LiveSessionTreeType } from '../domain/live';
import { mockLiveSessionTree } from '../api/mockData';

describe('LiveSessionTree Component', () => {
  it('renders informative empty state when root_sessions is empty', () => {
    const emptyTree: LiveSessionTreeType = {
      root_sessions: [],
      last_refreshed_at: '2026-10-20T12:00:00Z',
      total_active_sessions: 0,
      active_anomalies: [],
    };

    render(
      <LiveSessionTree
        tree={emptyTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    expect(screen.getByTestId('live-session-tree-empty')).toBeInTheDocument();
    expect(screen.getByText('No Active Sessions')).toBeInTheDocument();
    expect(
      screen.getByText('There are no active agent sessions currently running in the environment.')
    ).toBeInTheDocument();
  });

  it('renders multiple concurrent root sessions with agent names, roles, and status pills', () => {
    render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    expect(screen.getByTestId('live-session-tree')).toBeInTheDocument();
    expect(screen.getAllByText('orchestrator')[0]).toBeInTheDocument();
    expect(screen.getByText('agentic-lens-po')).toBeInTheDocument();
    expect(
      screen.getByText('AL-20 TypeScript Live Observability Implementation')
    ).toBeInTheDocument();
    expect(screen.getByText('PO-2 Roadmap Ticket Decomposition')).toBeInTheDocument();
  });

  it('renders deeply nested child sub-sessions with correct depth styling and independent toggle states', () => {
    const deepTree: LiveSessionTreeType = {
      root_sessions: [
        {
          id: 'root-1',
          parent_id: null,
          agent: 'orchestrator',
          role: 'orchestrator',
          model: 'claude-3-opus',
          title: 'Orchestrator Root Session',
          time_created: '2026-10-20T10:00:00Z',
          time_updated: '2026-10-20T10:30:00Z',
          status: 'running',
          ticket_confidence: 0.9,
          anomalies: [],
          sub_sessions: [
            {
              id: 'child-level-1',
              parent_id: 'root-1',
              agent: 'coder',
              role: 'coder',
              model: 'gemini-3.8-flash',
              title: 'Child Coder Session',
              time_created: '2026-10-20T10:05:00Z',
              time_updated: '2026-10-20T10:25:00Z',
              status: 'running',
              ticket_confidence: 0.8,
              anomalies: [],
              sub_sessions: [
                {
                  id: 'grandchild-level-2',
                  parent_id: 'child-level-1',
                  agent: 'tester',
                  role: 'tester',
                  model: 'gemini-3.8-flash',
                  title: 'Deep Grandchild Tester Session',
                  time_created: '2026-10-20T10:10:00Z',
                  time_updated: '2026-10-20T10:20:00Z',
                  status: 'completed',
                  ticket_confidence: 0.7,
                  anomalies: [],
                  sub_sessions: [],
                },
              ],
            },
          ],
        },
      ],
      last_refreshed_at: '2026-10-20T10:30:00Z',
      total_active_sessions: 3,
      active_anomalies: [],
    };

    render(
      <LiveSessionTree
        tree={deepTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    const rootWrapper = screen.getByTestId('session-node-root-1');
    const childWrapper = screen.getByTestId('session-node-child-level-1');
    const grandchildWrapper = screen.getByTestId('session-node-grandchild-level-2');

    expect(rootWrapper).toHaveClass('depth-0');
    expect(childWrapper).toHaveClass('depth-1');
    expect(grandchildWrapper).toHaveClass('depth-2');

    // Collapse child-level-1 (which has grandchild-level-2)
    const childCollapseBtn = within(childWrapper).getByRole('button', { name: /collapse sub-sessions \(1\)/i });
    fireEvent.click(childCollapseBtn);

    // Grandchild is hidden, but root and child remain visible
    expect(screen.queryByText('Deep Grandchild Tester Session')).not.toBeInTheDocument();
    expect(screen.getByText('Child Coder Session')).toBeInTheDocument();
    expect(screen.getByText('Orchestrator Root Session')).toBeInTheDocument();
  });

  it('toggles hierarchical child sub-sessions visibility on toggle button click', () => {
    render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    // By default, sub-sessions are visible
    expect(screen.getByText('TypeScript Frontend Implementation')).toBeInTheDocument();
    expect(screen.getByText('Contract and Architecture Plan')).toBeInTheDocument();

    // Click collapse button
    const collapseBtn = screen.getByRole('button', { name: /collapse sub-sessions \(3\)/i });
    fireEvent.click(collapseBtn);

    // Now sub-sessions should be hidden
    expect(screen.queryByText('TypeScript Frontend Implementation')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /expand sub-sessions \(3\)/i })).toBeInTheDocument();

    // Click expand button to restore
    fireEvent.click(screen.getByRole('button', { name: /expand sub-sessions \(3\)/i }));
    expect(screen.getByText('TypeScript Frontend Implementation')).toBeInTheDocument();
  });

  it('renders guessed ticket candidate badges with confidence score and invokes onSelectTicket without triggering onSelectSession', () => {
    const onSelectTicket = vi.fn();
    const onSelectSession = vi.fn();

    render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={onSelectSession}
        onSelectTicket={onSelectTicket}
      />
    );

    expect(screen.getByText('AL-20')).toBeInTheDocument();
    expect(screen.getByText('100% confidence')).toBeInTheDocument();
    expect(screen.getByText('PO-2')).toBeInTheDocument();
    expect(screen.getByText('85% confidence')).toBeInTheDocument();

    const al20Badge = screen.getByText('AL-20').closest('.guessed-ticket-badge');
    expect(al20Badge).not.toBeNull();
    fireEvent.click(al20Badge!);

    // onSelectTicket called, but onSelectSession NOT called due to stopPropagation
    expect(onSelectTicket).toHaveBeenCalledWith('AL-20');
    expect(onSelectSession).not.toHaveBeenCalled();
  });

  it('activates guessed ticket badge via keyboard Enter and Space keys', () => {
    const onSelectTicket = vi.fn();
    render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
        onSelectTicket={onSelectTicket}
      />
    );

    const al20Badge = screen.getByText('AL-20').closest('.guessed-ticket-badge')!;

    fireEvent.keyDown(al20Badge, { key: 'Enter' });
    expect(onSelectTicket).toHaveBeenCalledWith('AL-20');

    fireEvent.keyDown(al20Badge, { key: ' ' });
    expect(onSelectTicket).toHaveBeenCalledTimes(2);

    // Other keys do nothing
    fireEvent.keyDown(al20Badge, { key: 'Escape' });
    expect(onSelectTicket).toHaveBeenCalledTimes(2);
  });

  it('handles edge case confidence values: 0%, 50%, 100%, and omits badge when ticket id is missing', () => {
    const edgeTree: LiveSessionTreeType = {
      root_sessions: [
        {
          id: 'ses-zero',
          parent_id: null,
          agent: 'coder',
          role: 'coder',
          model: 'gemini-3.8-flash',
          title: 'Zero Confidence Node',
          time_created: '2026-10-20T12:00:00Z',
          time_updated: '2026-10-20T12:05:00Z',
          status: 'running',
          guessed_ticket_id: 'AL-99',
          ticket_confidence: 0.0,
          anomalies: [],
          sub_sessions: [],
        },
        {
          id: 'ses-half',
          parent_id: null,
          agent: 'tester',
          role: 'tester',
          model: 'gemini-3.8-flash',
          title: 'Half Confidence Node',
          time_created: '2026-10-20T12:00:00Z',
          time_updated: '2026-10-20T12:05:00Z',
          status: 'running',
          guessed_ticket_id: 'AL-50',
          ticket_confidence: 0.504,
          anomalies: [],
          sub_sessions: [],
        },
        {
          id: 'ses-no-badge',
          parent_id: null,
          agent: 'reviewer',
          role: 'reviewer',
          model: 'gemini-3.8-flash',
          title: 'No Guessed Ticket Node',
          time_created: '2026-10-20T12:00:00Z',
          time_updated: '2026-10-20T12:05:00Z',
          status: 'running',
          ticket_confidence: 0.0,
          anomalies: [],
          sub_sessions: [],
        },
      ],
      last_refreshed_at: '2026-10-20T12:05:00Z',
      total_active_sessions: 3,
      active_anomalies: [],
    };

    render(
      <LiveSessionTree
        tree={edgeTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    expect(screen.getByText('0% confidence')).toBeInTheDocument();
    expect(screen.getByText('50% confidence')).toBeInTheDocument();
    expect(screen.queryByText('AL-99')).toBeInTheDocument();
    expect(screen.queryByText('AL-50')).toBeInTheDocument();

    const noBadgeCard = screen.getByText('No Guessed Ticket Node').closest('.live-session-node-card')!;
    expect(noBadgeCard.querySelector('.guessed-ticket-badge')).toBeNull();
  });

  it('renders anomaly count pills accurately for singular and plural counts', () => {
    const anomalyTree: LiveSessionTreeType = {
      root_sessions: [
        {
          id: 'ses-one-anomaly',
          parent_id: null,
          agent: 'coder',
          role: 'coder',
          model: 'gemini-3.8-flash',
          title: 'Single Anomaly Session',
          time_created: '2026-10-20T12:00:00Z',
          time_updated: '2026-10-20T12:05:00Z',
          status: 'running',
          ticket_confidence: 0.0,
          anomalies: [
            {
              session_id: 'ses-one-anomaly',
              rule_name: 'ToolThrashing',
              severity: 'WARNING',
              message: 'Repeated tool execution',
              trigger_event: 'evt-1',
              timestamp: '2026-10-20T12:01:00Z',
            },
          ],
          sub_sessions: [],
        },
        {
          id: 'ses-multi-anomalies',
          parent_id: null,
          agent: 'coder',
          role: 'coder',
          model: 'gemini-3.8-flash',
          title: 'Multiple Anomalies Session',
          time_created: '2026-10-20T12:00:00Z',
          time_updated: '2026-10-20T12:05:00Z',
          status: 'running',
          ticket_confidence: 0.0,
          anomalies: [
            {
              session_id: 'ses-multi-anomalies',
              rule_name: 'ToolThrashing',
              severity: 'WARNING',
              message: 'Repeated tool execution',
              trigger_event: 'evt-1',
              timestamp: '2026-10-20T12:01:00Z',
            },
            {
              session_id: 'ses-multi-anomalies',
              rule_name: 'ContextCeilingApproach',
              severity: 'CRITICAL',
              message: 'Out of context',
              trigger_event: 'evt-2',
              timestamp: '2026-10-20T12:02:00Z',
            },
          ],
          sub_sessions: [],
        },
      ],
      last_refreshed_at: '2026-10-20T12:05:00Z',
      total_active_sessions: 2,
      active_anomalies: [],
    };

    render(
      <LiveSessionTree
        tree={anomalyTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    expect(screen.getByText('1 anomaly')).toBeInTheDocument();
    expect(screen.getByText('2 anomalies')).toBeInTheDocument();
  });

  it('safely handles corrupted timestamps and clock skew in duration calculations', () => {
    const corruptTimestampTree: LiveSessionTreeType = {
      root_sessions: [
        {
          id: 'ses-corrupt-date',
          parent_id: null,
          agent: 'coder',
          role: 'coder',
          model: 'gemini-3.8-flash',
          title: 'Corrupted Dates Node',
          time_created: 'not-a-valid-date',
          time_updated: 'not-a-valid-date-either',
          status: 'running',
          ticket_confidence: 0.0,
          anomalies: [],
          sub_sessions: [],
        },
        {
          id: 'ses-clock-skew',
          parent_id: null,
          agent: 'tester',
          role: 'tester',
          model: 'gemini-3.8-flash',
          title: 'Clock Skew Node',
          time_created: '2026-10-20T12:30:00Z',
          time_updated: '2026-10-20T12:00:00Z', // Updated before created!
          status: 'running',
          ticket_confidence: 0.0,
          anomalies: [],
          sub_sessions: [],
        },
      ],
      last_refreshed_at: '2026-10-20T12:05:00Z',
      total_active_sessions: 2,
      active_anomalies: [],
    };

    render(
      <LiveSessionTree
        tree={corruptTimestampTree}
        selectedSessionId={null}
        onSelectSession={vi.fn()}
      />
    );

    const durationPills = screen.getAllByText('0m 0s');
    expect(durationPills.length).toBe(2);
  });

  it('invokes onSelectSession and marks clicked node with selected styling', () => {
    const onSelectSession = vi.fn();

    const { rerender } = render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={onSelectSession}
      />
    );

    const rootCard = screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card');
    expect(rootCard).not.toBeNull();
    expect(rootCard).not.toHaveClass('selected');

    fireEvent.click(rootCard!);
    expect(onSelectSession).toHaveBeenCalledWith('ses-al20-root');

    // Rerender with selectedSessionId set
    rerender(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId="ses-al20-root"
        onSelectSession={onSelectSession}
      />
    );

    const updatedCard = screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card');
    expect(updatedCard).toHaveClass('selected');
    expect(updatedCard).toHaveAttribute('aria-selected', 'true');
  });

  it('activates session selection via keyboard Enter and Space keys and ignores other keys', () => {
    const onSelectSession = vi.fn();

    render(
      <LiveSessionTree
        tree={mockLiveSessionTree}
        selectedSessionId={null}
        onSelectSession={onSelectSession}
      />
    );

    const rootCard = screen.getByText('AL-20 TypeScript Live Observability Implementation').closest('.live-session-node-card')!;

    fireEvent.keyDown(rootCard, { key: 'Enter' });
    expect(onSelectSession).toHaveBeenCalledWith('ses-al20-root');

    fireEvent.keyDown(rootCard, { key: ' ' });
    expect(onSelectSession).toHaveBeenCalledTimes(2);

    fireEvent.keyDown(rootCard, { key: 'Tab' });
    fireEvent.keyDown(rootCard, { key: 'Escape' });
    expect(onSelectSession).toHaveBeenCalledTimes(2);
  });
});
