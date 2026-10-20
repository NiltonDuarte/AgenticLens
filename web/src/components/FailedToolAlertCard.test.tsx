import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FailedToolAlertCard } from './FailedToolAlertCard';
import type { LivePartFact } from '../domain/live';

describe('FailedToolAlertCard Component', () => {
  const basePart: LivePartFact = {
    id: 'part-err-01',
    type: 'tool',
    tool_name: 'bash',
    input_snippet: 'cargo test --workspace',
    output_snippet: 'error[E0432]: unresolved import `std::collections::NonExistentMap`',
    status: 'error',
    is_error: true,
    error_message: 'Compilation failure in workspace crate',
    severity: 'HIGH',
  };

  it('renders tool name, failed status badge, and high severity tag with red accent styling', () => {
    render(<FailedToolAlertCard part={basePart} />);

    expect(screen.getByText('bash')).toBeInTheDocument();
    expect(screen.getByText('FAILED')).toBeInTheDocument();
    expect(screen.getByText('HIGH')).toBeInTheDocument();
    expect(screen.getByText('Command / Input:')).toBeInTheDocument();
    expect(screen.getByText('cargo test --workspace')).toBeInTheDocument();
    expect(screen.getByText(/unresolved import/)).toBeInTheDocument();

    const card = screen.getByTestId('failed-tool-alert-card');
    expect(card).toHaveClass('severity-high');
  });

  it('truncates error output exceeding 200 characters and toggles full text on button click', () => {
    const longOutput =
      'Stack trace line 1: failed to execute database migration.\n' +
      'Stack trace line 2: duplicate column definition detected at column_id_001.\n' +
      'Stack trace line 3: terminating active worker pool due to unhandled fatal constraint.\n' +
      'Stack trace line 4: connection terminated prematurely by remote host.';

    expect(longOutput.length).toBeGreaterThan(200);

    const longPart: LivePartFact = {
      ...basePart,
      output_snippet: longOutput,
    };

    render(<FailedToolAlertCard part={longPart} />);

    // Initially truncated with ellipsis
    const toggleBtn = screen.getByRole('button', { name: /expand output/i });
    expect(toggleBtn).toBeInTheDocument();
    expect(screen.getByText(/Stack trace line 1/)).toBeInTheDocument();

    // Click Expand Output
    fireEvent.click(toggleBtn);
    expect(screen.getByRole('button', { name: /collapse output/i })).toBeInTheDocument();
    expect(screen.getByText(longOutput, { collapseWhitespace: false })).toBeInTheDocument();

    // Click Collapse Output
    fireEvent.click(screen.getByRole('button', { name: /collapse output/i }));
    expect(screen.getByRole('button', { name: /expand output/i })).toBeInTheDocument();
  });

  it('renders consecutive failure streak counter when consecutiveErrors is greater than 1', () => {
    render(<FailedToolAlertCard part={basePart} consecutiveErrors={4} />);

    expect(screen.getByText('4 Consecutive Failures')).toBeInTheDocument();
  });

  it('does not render streak counter when consecutiveErrors is 1 or undefined', () => {
    const { rerender } = render(<FailedToolAlertCard part={basePart} consecutiveErrors={1} />);
    expect(screen.queryByText(/Consecutive Failures/)).not.toBeInTheDocument();

    rerender(<FailedToolAlertCard part={basePart} />);
    expect(screen.queryByText(/Consecutive Failures/)).not.toBeInTheDocument();
  });

  it('invokes onInspect callback when Inspect button is clicked', () => {
    const onInspect = vi.fn();
    render(<FailedToolAlertCard part={basePart} onInspect={onInspect} />);

    const inspectBtn = screen.getByRole('button', { name: /inspect/i });
    fireEvent.click(inspectBtn);
    expect(onInspect).toHaveBeenCalledTimes(1);
  });

  it('renders all severity levels (CRITICAL, HIGH, MEDIUM, LOW) and defaults to HIGH when omitted', () => {
    const severities = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'] as const;

    for (const sev of severities) {
      const partWithSev: LivePartFact = {
        ...basePart,
        severity: sev,
      };

      const { unmount } = render(<FailedToolAlertCard part={partWithSev} />);
      const card = screen.getByTestId('failed-tool-alert-card');
      expect(card).toHaveClass(`severity-${sev.toLowerCase()}`);
      expect(screen.getByText(sev)).toBeInTheDocument();
      unmount();
    }

    // Omitted severity defaults to HIGH
    const omittedSevPart: LivePartFact = {
      id: 'part-omitted-sev',
      type: 'tool',
      status: 'error',
      is_error: true,
      severity: '',
    };
    render(<FailedToolAlertCard part={omittedSevPart} />);
    const card = screen.getByTestId('failed-tool-alert-card');
    expect(card).toHaveClass('severity-high');
    expect(screen.getByText('HIGH')).toBeInTheDocument();
  });

  it('handles output snippet length boundary: 200 chars is not truncated, 201 chars triggers expand button', () => {
    // Exactly 200 characters
    const exact200 = 'x'.repeat(200);
    const part200: LivePartFact = {
      ...basePart,
      output_snippet: exact200,
    };

    const { unmount } = render(<FailedToolAlertCard part={part200} />);
    expect(screen.queryByRole('button', { name: /expand output/i })).not.toBeInTheDocument();
    expect(screen.getByText(exact200)).toBeInTheDocument();
    unmount();

    // 201 characters
    const text201 = 'y'.repeat(201);
    const part201: LivePartFact = {
      ...basePart,
      output_snippet: text201,
    };

    render(<FailedToolAlertCard part={part201} />);
    const toggleBtn = screen.getByRole('button', { name: /expand output/i });
    expect(toggleBtn).toBeInTheDocument();
    expect(screen.getByText(`${'y'.repeat(200)}...`)).toBeInTheDocument();
  });

  it('falls back to error_message when output_snippet is missing, and to default text when both are missing', () => {
    const onlyErrorMsgPart: LivePartFact = {
      id: 'part-only-msg',
      type: 'tool',
      status: 'error',
      is_error: true,
      severity: 'HIGH',
      error_message: 'Explicit error reason',
    };

    const { unmount } = render(<FailedToolAlertCard part={onlyErrorMsgPart} />);
    expect(screen.getByText('Explicit error reason')).toBeInTheDocument();
    unmount();

    const noSnippetPart: LivePartFact = {
      id: 'part-no-snippet',
      type: 'tool',
      status: 'error',
      is_error: true,
      severity: 'HIGH',
    };

    render(<FailedToolAlertCard part={noSnippetPart} />);
    expect(screen.getByText('Execution error without output message')).toBeInTheDocument();
  });

  it('handles missing tool name and omitted input snippet gracefully', () => {
    const missingFieldsPart: LivePartFact = {
      id: 'part-no-tool-input',
      type: 'tool',
      status: 'error',
      is_error: true,
      severity: 'HIGH',
    };

    render(<FailedToolAlertCard part={missingFieldsPart} />);
    expect(screen.getByText('unknown_tool')).toBeInTheDocument();
    expect(screen.queryByText('Command / Input:')).not.toBeInTheDocument();
  });

  it('safely renders ANSI escape sequences, HTML characters, and injection vectors as text', () => {
    const dangerousPart: LivePartFact = {
      ...basePart,
      input_snippet: '<script>alert("xss")</script> && cargo check',
      output_snippet: '\u001b[31mFATAL ERROR\u001b[0m: <div class="trap">Vec<Option<&str>> failed</div>',
    };

    render(<FailedToolAlertCard part={dangerousPart} />);

    // Verifies text is safely rendered literally
    expect(screen.getByText('<script>alert("xss")</script> && cargo check')).toBeInTheDocument();
    expect(screen.getByText(/FATAL ERROR/)).toBeInTheDocument();
    expect(screen.getByText(/Vec<Option<&str>>/)).toBeInTheDocument();
  });
});
