import React, { useState } from 'react';
import type { TraceBadgeMetadata } from '../domain/traceBadges';
import { copyTextToClipboard } from '../utils/clipboard';

export interface ChildSessionNavigation {
  readonly availableSessionIds?: readonly string[] | undefined;
  readonly onSelectSession?: ((sessionId: string) => void) | undefined;
}

interface TraceInvocationBadgeProps extends ChildSessionNavigation {
  readonly metadata: TraceBadgeMetadata | undefined;
}

export const TraceInvocationBadge: React.FC<TraceInvocationBadgeProps> = ({
  metadata, availableSessionIds, onSelectSession,
}) => {
  const [copyStatus, setCopyStatus] = useState('');
  if (!metadata) return null;
  const label = metadata.kind === 'skill' ? `[skill: ${metadata.name}]` : `[task -> ${metadata.name}]`;
  const childId = metadata.kind === 'task' ? metadata.childSessionId : undefined;
  const canNavigate = childId && availableSessionIds?.includes(childId) && onSelectSession;
  return (
    <span className="trace-invocation-controls" onClick={(event) => event.stopPropagation()}
      onKeyDown={(event) => event.stopPropagation()}>
      {canNavigate ? (
        <button type="button" className={`trace-invocation-badge badge-${metadata.kind}`}
          title={`${label} — Inspect child session ${childId}`} aria-label={`Inspect child session ${childId}`}
          onClick={() => onSelectSession(childId)}>{label}</button>
      ) : (
        <span className={`trace-invocation-badge badge-${metadata.kind}`} title={label}>{label}</span>
      )}
      {childId && !canNavigate && (
        <span className="trace-child-unavailable" title={`Child session ${childId} is pending or unavailable in this view`}>
          <span className="trace-child-id">Child: {childId}</span>
          <button type="button" className="btn-copy-id" aria-label={`Copy child session ID ${childId}`}
            onClick={async () => setCopyStatus(await copyTextToClipboard(childId) ? 'Copied' : 'Copy failed')}>
            {copyStatus || 'Copy ID'}
          </button>
          <span className="sr-only" role="status">{copyStatus}</span>
        </span>
      )}
    </span>
  );
};
