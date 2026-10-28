import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { WorkspaceLoadWarning } from './WorkspaceLoadWarning';

describe('WorkspaceLoadWarning', () => {
  it('names every skipped ticket and shows its remediation reason', () => {
    render(<WorkspaceLoadWarning diagnostics={[
        { ticket_id: 'AL-OLD', message: 'Schema v3 requires migration and source-backed backfill.' },
        { ticket_id: 'AL-PAUSED', message: 'Resume the incomplete migration workflow.' },
      ]} />);

    expect(screen.getByRole('status')).toHaveTextContent('Some ticket workspaces were skipped.');
    expect(screen.getByText(/AL-OLD:/)).toBeInTheDocument();
    expect(screen.getByText(/source-backed backfill/)).toBeInTheDocument();
    expect(screen.getByText(/AL-PAUSED:/)).toBeInTheDocument();
  });

  it('renders nothing when all workspaces loaded', () => {
    const { container } = render(<WorkspaceLoadWarning diagnostics={[]} />);
    expect(container).toBeEmptyDOMElement();
  });
});
