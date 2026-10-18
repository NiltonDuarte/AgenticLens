import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AnomalyDrawer } from './AnomalyDrawer';
import type { Anomaly, EventFact } from '../domain/types';

describe('AnomalyDrawer Component', () => {
  const sampleAnomaly: Anomaly = {
    rule_id: 'TOOL_RETRY_STREAK',
    severity: 'WARNING',
    title: 'Repeated Tool Failures',
    description: 'Tool git failed 3 consecutive times.',
    context: {
      tool_name: 'git',
      session_id: 'session-123',
      turn: 4,
      role: 'coder',
      count: 3,
      threshold: '3 consecutive failures',
      token_delta: 2500,
      context_tokens: 45000,
      failure_rate_pct: 75.0,
    },
  };

  const sampleEvent: EventFact = {
    timestamp: '2026-10-18T10:15:30Z',
    session_id: 'session-123',
    turn: 4,
    role: 'coder',
    event_type: 'tool_call',
    tool_name: 'git',
    status: 'error',
    error_message: 'fatal: destination path already exists',
    duration_ms: 320,
    context_tokens: 45000,
  };

  it('returns null or remains hidden when isOpen is false', () => {
    const { container } = render(
      <AnomalyDrawer
        isOpen={false}
        onClose={vi.fn()}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders title, severity badge, rule ID, and description when isOpen is true and anomaly is provided', () => {
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    expect(screen.getByText('Repeated Tool Failures')).toBeInTheDocument();
    expect(screen.getByText('Rule Diagnosis')).toBeInTheDocument();
    expect(screen.getByText('TOOL_RETRY_STREAK')).toBeInTheDocument();
    expect(screen.getByText('Tool git failed 3 consecutive times.')).toBeInTheDocument();
    expect(screen.getByText('WARNING')).toBeInTheDocument();
  });

  it('renders diagnostic context fields: turn number, session ID, tool name, threshold, and token delta', () => {
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    expect(screen.getByText('Turn Number')).toBeInTheDocument();
    expect(screen.getByText('Turn 4')).toBeInTheDocument();
    expect(screen.getByText('session-123')).toBeInTheDocument();
    expect(screen.getByText('git')).toBeInTheDocument();
    expect(screen.getByText('3 consecutive failures')).toBeInTheDocument();
    expect(screen.getByText('+2,500 tokens')).toBeInTheDocument();
    expect(screen.getByText('75.0%')).toBeInTheDocument();
  });

  it('renders error message container in monospace font when event.error_message is present', () => {
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    const errorBlock = screen.getByText('fatal: destination path already exists');
    expect(errorBlock).toBeInTheDocument();
    expect(errorBlock).toHaveClass('drawer-error-content');
  });

  it('renders accurate Tool Execution Failure title and omits Rule Diagnosis for plain unflagged tool errors', () => {
    const plainErrorEvent: EventFact = {
      timestamp: '2026-10-18T10:20:00Z',
      session_id: 'session-456',
      turn: 2,
      role: 'coder',
      event_type: 'tool_call',
      tool_name: 'bash',
      status: 'error',
      error_message: 'exit status 1: build failed',
      duration_ms: 150,
      context_tokens: 30000,
    };

    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={undefined}
        event={plainErrorEvent}
      />
    );

    // Title must be Tool Execution Failure
    expect(screen.getByText('Tool Execution Failure')).toBeInTheDocument();

    // Rule Diagnosis section must NOT be rendered
    expect(screen.queryByText('Rule Diagnosis')).not.toBeInTheDocument();

    // Fake fallbacks must NOT appear anywhere in the document
    expect(screen.queryByText('ANOMALY_DIAGNOSTIC')).not.toBeInTheDocument();
    expect(screen.queryByText('Operational Anomaly Detected')).not.toBeInTheDocument();

    // No AnomalyBadge in header
    expect(screen.queryByRole('status')).not.toBeInTheDocument();

    // Monospace error message container must be present
    const errorBlock = screen.getByText('exit status 1: build failed');
    expect(errorBlock).toBeInTheDocument();
    expect(errorBlock).toHaveClass('drawer-error-content');

    // Context metadata populated from event
    expect(screen.getByText('Turn 2')).toBeInTheDocument();
    expect(screen.getByText('session-456')).toBeInTheDocument();
    expect(screen.getByText('bash')).toBeInTheDocument();
  });

  it('handles large 4096-character error payloads without crashing and preserves full text', () => {
    const largeError = 'Error: ' + 'x'.repeat(4096 - 7);
    const largeErrorEvent: EventFact = {
      timestamp: '2026-10-18T10:22:00Z',
      session_id: 'session-large',
      turn: 5,
      role: 'coder',
      event_type: 'tool_call',
      tool_name: 'bash',
      status: 'error',
      error_message: largeError,
      duration_ms: 500,
      context_tokens: 50000,
    };

    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={undefined}
        event={largeErrorEvent}
      />
    );

    const errorBlock = screen.getByText(largeError);
    expect(errorBlock).toBeInTheDocument();
    expect(errorBlock.textContent?.length).toBe(4096);
  });

  it('invokes onClose when close button is clicked', () => {
    const handleClose = vi.fn();
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={handleClose}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    const closeButton = screen.getByRole('button', { name: 'Close' });
    expect(closeButton).toBeInTheDocument();
    fireEvent.click(closeButton);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('invokes onClose when backdrop overlay is clicked', () => {
    const handleClose = vi.fn();
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={handleClose}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    const backdrop = screen.getByTestId('anomaly-drawer-backdrop');
    expect(backdrop).toBeInTheDocument();
    fireEvent.click(backdrop);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('invokes onClose when Escape key is pressed', () => {
    const handleClose = vi.fn();
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={handleClose}
        anomaly={sampleAnomaly}
        event={sampleEvent}
      />
    );

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('handles missing anomaly context or missing event gracefully without runtime errors', () => {
    const minimalAnomaly: Anomaly = {
      rule_id: 'CONTEXT_SATURATION_SPIKE',
      severity: 'CRITICAL',
      title: 'Context Saturation Spike',
      description: 'Context spiked sharply',
      context: {},
    };

    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={minimalAnomaly}
      />
    );

    expect(screen.getByText('Context Saturation Spike')).toBeInTheDocument();
    expect(screen.getByText('CONTEXT_SATURATION_SPIKE')).toBeInTheDocument();
    expect(screen.getByText('CRITICAL')).toBeInTheDocument();
    expect(screen.queryByText('Error Payload')).not.toBeInTheDocument();
  });

  it('handles completely undefined anomaly and event gracefully with neutral title and no fake diagnostics', () => {
    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
      />
    );

    expect(screen.getByText('Execution Details')).toBeInTheDocument();
    expect(screen.queryByText('Rule Diagnosis')).not.toBeInTheDocument();
    expect(screen.queryByText('Operational Anomaly Detected')).not.toBeInTheDocument();
    expect(screen.queryByText('ANOMALY_DIAGNOSTIC')).not.toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('falls back to event metadata when anomaly context fields are missing', () => {
    const anomalyWithoutContext: Anomaly = {
      rule_id: 'TOOL_HIGH_FAILURE_RATE',
      severity: 'WARNING',
      title: 'High Failure Rate',
      description: 'Tool failure rate exceeded threshold',
      context: {},
    };

    render(
      <AnomalyDrawer
        isOpen={true}
        onClose={vi.fn()}
        anomaly={anomalyWithoutContext}
        event={sampleEvent}
      />
    );

    // Context metadata should be populated from sampleEvent
    expect(screen.getByText('Turn 4')).toBeInTheDocument();
    expect(screen.getByText('session-123')).toBeInTheDocument();
    expect(screen.getByText('git')).toBeInTheDocument();
    expect(screen.getByText('coder')).toBeInTheDocument();
    expect(screen.getByText('320ms')).toBeInTheDocument();
    expect(screen.getByText('fatal: destination path already exists')).toBeInTheDocument();
  });
});
