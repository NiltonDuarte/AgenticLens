import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AnomalyBadge } from './AnomalyBadge';

describe('AnomalyBadge Component', () => {
  it('renders WARNING severity with amber styling classes and accessible role status', () => {
    render(<AnomalyBadge severity="WARNING" />);
    const badge = screen.getByRole('status');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('anomaly-badge');
    expect(badge).toHaveClass('warning');
    expect(badge).toHaveTextContent('WARNING');
  });

  it('renders CRITICAL severity with red styling classes and accessible role status', () => {
    render(<AnomalyBadge severity="CRITICAL" />);
    const badge = screen.getByRole('status');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('anomaly-badge');
    expect(badge).toHaveClass('critical');
    expect(badge).toHaveTextContent('CRITICAL');
  });

  it('renders count accurately when count prop is provided', () => {
    render(<AnomalyBadge severity="WARNING" count={4} />);
    const countSpan = screen.getByText('4');
    expect(countSpan).toBeInTheDocument();
    expect(countSpan).toHaveClass('anomaly-badge-count');
  });

  it('renders custom label string when label prop is provided', () => {
    render(<AnomalyBadge severity="CRITICAL" label="High Failure Rate" />);
    expect(screen.getByText('High Failure Rate')).toBeInTheDocument();
  });

  it('invokes onClick callback when clicked and changes role to button', () => {
    const handleClick = vi.fn();
    render(<AnomalyBadge severity="WARNING" label="Clickable Warning" onClick={handleClick} />);
    const button = screen.getByRole('button', { name: 'Clickable Warning' });
    expect(button).toBeInTheDocument();
    fireEvent.click(button);
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('renders small, medium, and large sizes according to size prop', () => {
    const { rerender } = render(<AnomalyBadge severity="WARNING" size="sm" />);
    expect(screen.getByRole('status')).toHaveClass('sm');

    rerender(<AnomalyBadge severity="WARNING" size="md" />);
    expect(screen.getByRole('status')).toHaveClass('md');

    rerender(<AnomalyBadge severity="CRITICAL" size="lg" />);
    expect(screen.getByRole('status')).toHaveClass('lg');
  });

  it('renders default title attribute on clickable badge when title prop is omitted', () => {
    render(<AnomalyBadge severity="WARNING" onClick={() => {}} />);
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('title', 'Click to open Anomaly Inspector');
  });

  it('renders custom title attribute on clickable badge when title prop is provided', () => {
    render(
      <AnomalyBadge
        severity="CRITICAL"
        title="Custom Inspection Tooltip"
        onClick={() => {}}
      />
    );
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('title', 'Custom Inspection Tooltip');
  });

  it('renders custom title attribute on non-clickable badge when title prop is provided', () => {
    render(<AnomalyBadge severity="WARNING" title="Static Informational Title" />);
    const badge = screen.getByRole('status');
    expect(badge).toHaveAttribute('title', 'Static Informational Title');
  });

  it('does not set title attribute on non-clickable badge when title prop is omitted', () => {
    render(<AnomalyBadge severity="WARNING" />);
    const badge = screen.getByRole('status');
    expect(badge).not.toHaveAttribute('title');
  });
});
