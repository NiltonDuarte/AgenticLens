import React from 'react';
import type { WorkspaceDiagnostic } from '../domain/types';

export interface WorkspaceLoadWarningProps {
  readonly diagnostics: readonly WorkspaceDiagnostic[];
}

export function mergeWorkspaceDiagnostics(
  ...groups: (readonly WorkspaceDiagnostic[])[]
): readonly WorkspaceDiagnostic[] {
  const byTicket = new Map<string, WorkspaceDiagnostic>();
  for (const item of groups.flat()) {
    byTicket.set(item.scope === 'response' ? `response:${item.message}` : `workspace:${item.ticket_id}`, item);
  }
  return [...byTicket.values()];
}

export const WorkspaceLoadWarning: React.FC<WorkspaceLoadWarningProps> = ({ diagnostics }) => {
  if (diagnostics.length === 0) return null;

  return (
    <section className="workspace-load-warning" role="status" aria-live="polite">
      <strong>{diagnostics.some((item) => item.scope === 'response') ? 'Workspace response warnings.' : 'Some ticket workspaces were skipped.'}</strong>
      <ul>
        {diagnostics.map((item) => (
          <li key={item.scope === 'response' ? `response:${item.message}` : `workspace:${item.ticket_id}:${item.message}`}>
            {item.scope !== 'response' && <strong>{item.ticket_id}: </strong>}{item.message}
          </li>
        ))}
      </ul>
    </section>
  );
};
