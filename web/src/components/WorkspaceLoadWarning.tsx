import React from 'react';
import type { WorkspaceDiagnostic } from '../domain/types';

export interface WorkspaceLoadWarningProps {
  readonly diagnostics: readonly WorkspaceDiagnostic[];
}

export function mergeWorkspaceDiagnostics(
  ...groups: (readonly WorkspaceDiagnostic[])[]
): readonly WorkspaceDiagnostic[] {
  const uniqueDiagnostics = new Map<string, WorkspaceDiagnostic>();
  for (const item of groups.flat()) {
    uniqueDiagnostics.set(item.scope === 'response' ? `response:${item.message}` : JSON.stringify(['workspace', item.path, item.ticket_id, item.message]), item);
  }
  return [...uniqueDiagnostics.values()];
}

export const WorkspaceLoadWarning: React.FC<WorkspaceLoadWarningProps> = ({ diagnostics }) => {
  if (diagnostics.length === 0) return null;

  return (
    <section className="workspace-load-warning" role="status" aria-live="polite">
      <strong>{diagnostics.some((item) => item.scope === 'response') ? 'Workspace response warnings.' : 'Some ticket workspaces were skipped.'}</strong>
      <ul>
        {diagnostics.map((item) => (
          <li key={item.scope === 'response' ? `response:${item.message}` : JSON.stringify(['workspace', item.path, item.ticket_id, item.message])}>
            {item.scope !== 'response' && <>
              {item.ticket_id && <strong>{item.ticket_id}: </strong>}
              {item.path && <code className="workspace-location">{item.path}: </code>}
            </>}{item.message}
          </li>
        ))}
      </ul>
    </section>
  );
};
