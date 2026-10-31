import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AnomalyBadge } from './AnomalyBadge';

describe('AnomalyBadge Component', () => {
  it('renders WARNING severity with amber styling classes and accessible role status', () => {
    const { rerender } = render(<AnomalyBadge severity="WARNING" size="sm" />);
    const badge = screen.getByRole('status');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('anomaly-badge');
    expect(badge).toHaveClass('warning');
    expect(badge).toHaveTextContent('WARNING');
    expect(badge).toHaveClass('sm');
    expect(badge).not.toHaveAttribute('title');

    rerender(<AnomalyBadge severity="WARNING" size="md" title="Static Informational Title" />);
    expect(screen.getByRole('status')).toHaveClass('warning', 'md');
    expect(screen.getByRole('status')).toHaveTextContent('WARNING');
    expect(screen.getByRole('status')).toHaveAttribute('title', 'Static Informational Title');
  });

  it('renders CRITICAL severity with red styling classes and accessible role status', () => {
    render(<AnomalyBadge severity="CRITICAL" size="lg" />);
    const badge = screen.getByRole('status');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('anomaly-badge');
    expect(badge).toHaveClass('critical');
    expect(badge).toHaveTextContent('CRITICAL');
    expect(badge).toHaveClass('lg');
  });

  it('renders count accurately when count prop is provided', () => {
    render(<AnomalyBadge severity="WARNING" count={4} />);
    const countSpan = screen.getByText('4');
    expect(countSpan).toBeInTheDocument();
    expect(countSpan).toHaveClass('anomaly-badge-count');
  });

  it('invokes onClick callback when clicked and changes role to button', () => {
    const handleClick = vi.fn();
    const { rerender } = render(<AnomalyBadge severity="WARNING" label="Clickable Warning" onClick={handleClick} />);
    const button = screen.getByRole('button', { name: 'Clickable Warning' });
    expect(button).toBeInTheDocument();
    expect(screen.getByText('Clickable Warning')).toBeInTheDocument();
    expect(button).toHaveAttribute('title', 'Click to open Anomaly Inspector');
    fireEvent.click(button);
    expect(handleClick).toHaveBeenCalledTimes(1);
    rerender(
      <AnomalyBadge severity="WARNING" label="Clickable Warning" title="Custom Inspection Tooltip" onClick={handleClick} />
    );
    const updatedButton = screen.getByRole('button', { name: 'Clickable Warning' });
    expect(updatedButton).toHaveAttribute('title', 'Custom Inspection Tooltip');
    fireEvent.click(updatedButton);
    expect(handleClick).toHaveBeenCalledTimes(2);
  });
});
