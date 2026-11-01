import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { mergeWorkspaceDiagnostics, WorkspaceLoadWarning } from './WorkspaceLoadWarning';

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

  it('shows response warnings without inventing a ticket and recovers when diagnostics clear', () => {
    const warning = { scope: 'response' as const, message: 'Workspace diagnostics could not be read.' };
    const diagnostics = mergeWorkspaceDiagnostics([warning], [warning, { ticket_id: 'AL-OLD', message: 'Migration required.' }]);
    const { rerender } = render(<WorkspaceLoadWarning diagnostics={diagnostics} />);
    expect(screen.getByRole('status')).toHaveTextContent('Workspace response warnings.');
    expect(screen.getAllByText(warning.message)).toHaveLength(1);
    expect(screen.getByRole('status')).not.toHaveTextContent('undefined');
    expect(screen.getByRole('status')).toHaveTextContent('AL-OLD:');
    rerender(<WorkspaceLoadWarning diagnostics={[]} />);
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('renders nothing when all workspaces loaded', () => {
    const { container } = render(<WorkspaceLoadWarning diagnostics={[]} />);
    expect(container).toBeEmptyDOMElement();
  });
});
