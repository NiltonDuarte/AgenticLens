import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FleetAnomalyCard } from './FleetAnomalyCard';
import type { FleetAnomalySummary } from '../domain/types';

describe('FleetAnomalyCard Component', () => {
  const sampleSummary: FleetAnomalySummary = {
    total_anomalies: 5,
    warning_count: 3,
    critical_count: 2,
    severity_counts: { WARNING: 3, CRITICAL: 2 },
    rule_counts: {
      TOOL_RETRY_STREAK: 2,
      TOOL_HIGH_FAILURE_RATE: 1,
      CONTEXT_SATURATION_SPIKE: 1,
      SUBAGENT_REWORK_LOOP: 1,
    },
    total_tickets: 4,
    tickets_with_anomalies: 2,
    anomaly_rate_per_ticket: 1.25,
    brittle_tools: [
      {
        tool_name: 'edit',
        anomaly_count: 3,
        total_calls: 10,
        error_count: 4,
        failure_rate_pct: 40.0,
      },
      {
        tool_name: 'bash',
        anomaly_count: 1,
        total_calls: 20,
        error_count: 2,
        failure_rate_pct: 10.0,
      },
    ],
  };

  it('displays total anomalies, warning count, critical count, and anomaly rate per ticket', () => {
    render(<FleetAnomalyCard summary={sampleSummary} />);

    expect(screen.getByText('Fleet Anomaly and Reliability Profile')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(
      screen.getByText((_, element) => element?.textContent === '3 warning, 2 critical')
    ).toBeInTheDocument();
    expect(screen.getByText('1.25')).toBeInTheDocument();
  });

  it('displays affected tickets ratio (e.g. 2 of 4 tickets)', () => {
    render(<FleetAnomalyCard summary={sampleSummary} />);
    expect(screen.getByText('2 of 4 tickets')).toBeInTheDocument();
  });

  it('renders rule count values for all four anomaly rules', () => {
    render(<FleetAnomalyCard summary={sampleSummary} />);

    expect(screen.getByText('Tool Retry Streak')).toBeInTheDocument();
    expect(screen.getByText('Tool High Failure Rate')).toBeInTheDocument();
    expect(screen.getByText('Context Saturation Spike')).toBeInTheDocument();
    expect(screen.getByText('Subagent Rework Loop')).toBeInTheDocument();
  });

  it('renders brittle tools table with tool names, anomaly counts, total calls, and failure rate percentages', () => {
    render(<FleetAnomalyCard summary={sampleSummary} />);

    expect(screen.getByText('edit')).toBeInTheDocument();
    expect(screen.getByText('40.0%')).toBeInTheDocument();
    expect(screen.getByText('bash')).toBeInTheDocument();
    expect(screen.getByText('10.0%')).toBeInTheDocument();
  });

  it('handles division by zero defensively when a brittle tool has zero calls', () => {
    const summaryWithZeroCalls: FleetAnomalySummary = {
      ...sampleSummary,
      brittle_tools: [
        {
          tool_name: 'custom_tool',
          anomaly_count: 1,
          total_calls: 0,
          error_count: 0,
          failure_rate_pct: 0,
        },
        {
          tool_name: 'uncalled_tool',
          anomaly_count: 1,
          total_calls: 0,
          error_count: 5,
          failure_rate_pct: undefined as unknown as number,
        },
      ],
    };

    render(<FleetAnomalyCard summary={summaryWithZeroCalls} />);
    expect(screen.getByText('custom_tool')).toBeInTheDocument();
    expect(screen.getByText('uncalled_tool')).toBeInTheDocument();
    expect(screen.getAllByText('0.0%')).toHaveLength(2);
  });

  it('handles division by zero defensively when total_tickets is 0', () => {
    const summaryZeroTickets: FleetAnomalySummary = {
      ...sampleSummary,
      total_tickets: 0,
      tickets_with_anomalies: 0,
    };

    render(<FleetAnomalyCard summary={summaryZeroTickets} />);
    expect(screen.getByText('0% of fleet')).toBeInTheDocument();
  });

  it('renders clean state when summary prop is null or undefined', () => {
    const { rerender } = render(<FleetAnomalyCard summary={null} />);
    expect(screen.getByText('Fleet Operating Cleanly - Zero Anomalies Detected')).toBeInTheDocument();

    rerender(<FleetAnomalyCard summary={undefined} />);
    expect(screen.getByText('Fleet Operating Cleanly - Zero Anomalies Detected')).toBeInTheDocument();
  });

  it('displays positive clean state message when total_anomalies is zero', () => {
    const cleanSummary: FleetAnomalySummary = {
      total_anomalies: 0,
      warning_count: 0,
      critical_count: 0,
      severity_counts: {},
      rule_counts: {},
      total_tickets: 3,
      tickets_with_anomalies: 0,
      anomaly_rate_per_ticket: 0,
      brittle_tools: [],
    };

    render(<FleetAnomalyCard summary={cleanSummary} />);
    expect(screen.getByText('Fleet Operating Cleanly - Zero Anomalies Detected')).toBeInTheDocument();
    expect(
      screen.getByText(
        'All workspace executions operate within normal health thresholds with zero flagged anomalies.'
      )
    ).toBeInTheDocument();
  });

  it('handles empty brittle_tools array without table errors', () => {
    const summaryNoBrittleTools: FleetAnomalySummary = {
      ...sampleSummary,
      brittle_tools: [],
    };

    render(<FleetAnomalyCard summary={summaryNoBrittleTools} />);
    expect(
      screen.getByText('Zero brittle tools identified across the workspace fleet.')
    ).toBeInTheDocument();
  });
});
