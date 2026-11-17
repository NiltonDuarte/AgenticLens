/** API location is a selection coordinate, never a replacement ticket identity. */
export interface LocatedTicket {
  readonly ticket_id: string | null;
  readonly workspace_path?: string;
}

export function workspaceKey(workspace: LocatedTicket): string {
  return workspace.workspace_path ? JSON.stringify([workspace.ticket_id, workspace.workspace_path]) : workspace.ticket_id ?? '';
}

export function selectWorkspace(
  onSelect: ((ticketId: string, workspacePath?: string) => void) | undefined,
  workspace: LocatedTicket,
): void {
  if (!onSelect || !workspace.ticket_id) return;
  if (workspace.workspace_path) onSelect(workspace.ticket_id, workspace.workspace_path);
  else onSelect(workspace.ticket_id);
}
