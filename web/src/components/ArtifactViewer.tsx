import React, { useState } from 'react';
import type { TicketArtifacts } from '../domain/types';

export interface ArtifactViewerProps {
  readonly artifacts: TicketArtifacts | null | undefined;
}

export const ArtifactViewer: React.FC<ArtifactViewerProps> = ({ artifacts }) => {
  const [activeArtifactTab, setActiveArtifactTab] = useState<'contract' | 'summary'>('contract');

  const contractContent = artifacts?.contract_content;
  const summaryContent = artifacts?.summary_content;

  const currentContent = activeArtifactTab === 'contract' ? contractContent : summaryContent;
  const currentTitle = activeArtifactTab === 'contract' ? 'CONTRACT.md' : 'SUMMARY.md';

  return (
    <div className="artifact-viewer-card">
      <div className="artifact-header">
        <h4 className="section-title">Workspace Artifacts</h4>
        <div className="artifact-tabs">
          <button
            type="button"
            className={`artifact-tab-btn ${activeArtifactTab === 'contract' ? 'active' : ''}`}
            onClick={() => setActiveArtifactTab('contract')}
          >
            CONTRACT.md {contractContent ? '' : '(Empty)'}
          </button>
          <button
            type="button"
            className={`artifact-tab-btn ${activeArtifactTab === 'summary' ? 'active' : ''}`}
            onClick={() => setActiveArtifactTab('summary')}
          >
            SUMMARY.md {summaryContent ? '' : '(Empty)'}
          </button>
        </div>
      </div>

      <div className="artifact-body">
        {currentContent && currentContent.trim().length > 0 ? (
          <pre className="artifact-pre">
            <code>{currentContent}</code>
          </pre>
        ) : (
          <div className="artifact-empty-note">
            No {currentTitle} document found for this workspace.
          </div>
        )}
      </div>
    </div>
  );
};
