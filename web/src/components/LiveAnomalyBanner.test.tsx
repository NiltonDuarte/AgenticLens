import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { LiveAnomalyBanner } from './LiveAnomalyBanner';
import type { LiveAnomalyAlert } from '../domain/live';

describe('LiveAnomalyBanner Component', () => {
  const mockAnomalies: readonly LiveAnomalyAlert[] = [
    {
      session_id: 'ses-crit-01',
      rule_name: 'ConsecutiveToolErrors',
      severity: 'CRITICAL',
      message: 'Persistent tool failures observed in subagent execution',
      trigger_event: 'part_fail_03',
      timestamp: '2026-10-20T12:00:00Z',
    },
    {
      session_id: 'ses-warn-02',
      rule_name: 'ToolThrashing',
      severity: 'WARNING',
      message: 'High frequency repeated file reads without mutations',
      trigger_event: 'part_read_09',
      timestamp: '2026-10-20T12:01:00Z',
    },
    {
      session_id: 'ses-high-03',
      rule_name: 'ContextCeilingApproach',
      severity: 'HIGH',
      message: 'Approaching context window boundary at 170k tokens',
      trigger_event: 'msg_05',
      timestamp: '2026-10-20T12:02:00Z',
    },
  ];

  it('renders null when anomalies array is empty', () => {
    const onSelectSession = vi.fn();
    const { container } = render(
      <LiveAnomalyBanner anomalies={[]} onSelectSession={onSelectSession} />
    );

    expect(container.firstChild).toBeNull();
  });

  it('applies CRITICAL severity badge and styling when critical anomaly is present', () => {
    const onSelectSession = vi.fn();
    render(
      <LiveAnomalyBanner anomalies={mockAnomalies} onSelectSession={onSelectSession} />
    );

    const banner = screen.getByRole('alert', { name: /active live anomalies/i });
    expect(banner).toHaveClass('severity-critical');
    expect(screen.getByText('3 Active Anomalies Detected')).toBeInTheDocument();
  });

  it('applies WARNING severity badge and styling when only warning anomalies exist', () => {
    const onSelectSession = vi.fn();
    const warningOnly: readonly LiveAnomalyAlert[] = [mockAnomalies[1]!];

    render(
      <LiveAnomalyBanner anomalies={warningOnly} onSelectSession={onSelectSession} />
    );

    const banner = screen.getByRole('alert', { name: /active live anomalies/i });
    expect(banner).toHaveClass('severity-warning');
    expect(screen.getByText('1 Active Anomaly Detected')).toBeInTheDocument();
  });

  it('invokes onSelectSession callback with session ID when anomaly item is clicked or activated via keyboard', () => {
    const onSelectSession = vi.fn();
    render(
      <LiveAnomalyBanner anomalies={mockAnomalies} onSelectSession={onSelectSession} />
    );

    const firstItem = screen.getByText('Persistent tool failures observed in subagent execution');
    fireEvent.click(firstItem);
    expect(onSelectSession).toHaveBeenCalledWith('ses-crit-01');

    const secondItem = screen.getByText('High frequency repeated file reads without mutations');
    fireEvent.keyDown(secondItem.closest('.live-anomaly-item')!, { key: 'Enter' });
    expect(onSelectSession).toHaveBeenCalledWith('ses-warn-02');
  });

  it('limits displayed anomalies to 2 by default and toggles full list on button click', () => {
    const onSelectSession = vi.fn();
    render(
      <LiveAnomalyBanner anomalies={mockAnomalies} onSelectSession={onSelectSession} />
    );

    expect(screen.getByText('Persistent tool failures observed in subagent execution')).toBeInTheDocument();
    expect(screen.getByText('High frequency repeated file reads without mutations')).toBeInTheDocument();
    expect(
      screen.queryByText('Approaching context window boundary at 170k tokens')
    ).not.toBeInTheDocument();

    const toggleBtn = screen.getByRole('button', { name: /\+1 more/i });
    fireEvent.click(toggleBtn);

    expect(screen.getByText('Approaching context window boundary at 170k tokens')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /show less/i })).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /show less/i }));
    expect(
      screen.queryByText('Approaching context window boundary at 170k tokens')
    ).not.toBeInTheDocument();
  });

  it('applies HIGH severity badge and theme when highest severity is HIGH without CRITICAL', () => {
    const onSelectSession = vi.fn();
    const highAndWarn: readonly LiveAnomalyAlert[] = [
      mockAnomalies[1]!, // WARNING
      mockAnomalies[2]!, // HIGH
    ];

    render(
      <LiveAnomalyBanner anomalies={highAndWarn} onSelectSession={onSelectSession} />
    );

    const banner = screen.getByRole('alert', { name: /active live anomalies/i });
    expect(banner).toHaveClass('severity-high');
    expect(screen.getByText('HIGH')).toBeInTheDocument();
    expect(screen.getByText('2 Active Anomalies Detected')).toBeInTheDocument();
  });

  it('does not render expand/collapse toggle when anomalies count is 1 or 2', () => {
    const onSelectSession = vi.fn();
    const oneAlert: readonly LiveAnomalyAlert[] = [mockAnomalies[0]!];

    const { rerender } = render(
      <LiveAnomalyBanner anomalies={oneAlert} onSelectSession={onSelectSession} />
    );

    expect(screen.queryByRole('button', { name: /more/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /show less/i })).not.toBeInTheDocument();

    const twoAlerts: readonly LiveAnomalyAlert[] = [mockAnomalies[0]!, mockAnomalies[1]!];
    rerender(
      <LiveAnomalyBanner anomalies={twoAlerts} onSelectSession={onSelectSession} />
    );

    expect(screen.queryByRole('button', { name: /more/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /show less/i })).not.toBeInTheDocument();
    expect(screen.getByText('Persistent tool failures observed in subagent execution')).toBeInTheDocument();
    expect(screen.getByText('High frequency repeated file reads without mutations')).toBeInTheDocument();
  });

  it('activates onSelectSession on Space key and ignores irrelevant keys', () => {
    const onSelectSession = vi.fn();
    render(
      <LiveAnomalyBanner anomalies={mockAnomalies} onSelectSession={onSelectSession} />
    );

    const firstItem = screen.getByText('Persistent tool failures observed in subagent execution').closest('.live-anomaly-item')!;

    // Space key triggers callback
    fireEvent.keyDown(firstItem, { key: ' ' });
    expect(onSelectSession).toHaveBeenCalledWith('ses-crit-01');

    // Irrelevant keys do not trigger callback
    fireEvent.keyDown(firstItem, { key: 'ArrowDown' });
    fireEvent.keyDown(firstItem, { key: 'Escape' });
    fireEvent.keyDown(firstItem, { key: 'Tab' });
    expect(onSelectSession).toHaveBeenCalledTimes(1);
  });
});
