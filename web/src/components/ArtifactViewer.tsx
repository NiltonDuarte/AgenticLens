import React, { useState, useRef, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { Components } from 'react-markdown';
import type { TicketArtifacts } from '../domain/types';

export type ArtifactTab = 'contract' | 'summary';
export type ViewMode = 'rich' | 'raw';
export type CopyStatus = 'Copy' | 'Copied';

export interface ArtifactViewerProps {
  readonly artifacts: TicketArtifacts | null | undefined;
}

const markdownComponents: Components = {
  table: ({ node: _node, ...props }) => (
    <div className="markdown-table-wrapper">
      <table className="markdown-table" {...props} />
    </div>
  ),
  th: ({ node: _node, ...props }) => <th className="markdown-th" {...props} />,
  td: ({ node: _node, ...props }) => <td className="markdown-td" {...props} />,
  blockquote: ({ node: _node, ...props }) => <blockquote className="markdown-blockquote" {...props} />,
  input: ({ node: _node, ...props }) => {
    if (props.type === 'checkbox') {
      return <input className="markdown-checkbox" disabled readOnly {...props} />;
    }
    return <input {...props} />;
  },
  li: ({ node: _node, className, children, ...props }) => {
    const isTask = typeof className === 'string' && className.includes('task-list-item');
    return (
      <li
        className={isTask ? `markdown-task-list-item ${className}` : className}
        {...props}
      >
        {children}
      </li>
    );
  },
  pre: ({ node: _node, ...props }) => <pre className="markdown-pre" {...props} />,
  code: ({ node: _node, className, children, ...props }) => {
    const isInline = !className && typeof children === 'string' && !children.includes('\n');
    return (
      <code
        className={isInline ? 'markdown-code-inline' : `markdown-code-block ${className ?? ''}`.trim()}
        {...props}
      >
        {children}
      </code>
    );
  },
  a: ({ node: _node, href, children, ...props }) => (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="markdown-link"
      {...props}
    >
      {children}
    </a>
  ),
  h1: ({ node: _node, ...props }) => <h1 className="markdown-h1" {...props} />,
  h2: ({ node: _node, ...props }) => <h2 className="markdown-h2" {...props} />,
  h3: ({ node: _node, ...props }) => <h3 className="markdown-h3" {...props} />,
  h4: ({ node: _node, ...props }) => <h4 className="markdown-h4" {...props} />,
  h5: ({ node: _node, ...props }) => <h5 className="markdown-h5" {...props} />,
  h6: ({ node: _node, ...props }) => <h6 className="markdown-h6" {...props} />,
};

export const ArtifactViewer: React.FC<ArtifactViewerProps> = ({ artifacts }) => {
  const [activeArtifactTab, setActiveArtifactTab] = useState<ArtifactTab>('contract');
  const [viewMode, setViewMode] = useState<ViewMode>('rich');
  const [copyStatus, setCopyStatus] = useState<CopyStatus>('Copy');

  const copyTimerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
    };
  }, []);

  const contractContent = artifacts?.contract_content ?? null;
  const summaryContent = artifacts?.summary_content ?? null;

  const currentContent = activeArtifactTab === 'contract' ? contractContent : summaryContent;
  const currentTitle = activeArtifactTab === 'contract' ? 'CONTRACT.md' : 'SUMMARY.md';

  const hasContractContent = Boolean(contractContent && contractContent.trim().length > 0);
  const hasSummaryContent = Boolean(summaryContent && summaryContent.trim().length > 0);
  const hasCurrentContent = Boolean(currentContent && currentContent.trim().length > 0);

  const handleTabChange = (tab: ArtifactTab) => {
    setActiveArtifactTab(tab);
    setCopyStatus('Copy');
    if (copyTimerRef.current !== null) {
      window.clearTimeout(copyTimerRef.current);
      copyTimerRef.current = null;
    }
  };

  const handleCopy = async () => {
    if (!currentContent || currentContent.trim().length === 0) {
      return;
    }
    try {
      if (typeof navigator !== 'undefined' && navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
        await navigator.clipboard.writeText(currentContent);
      } else if (typeof document !== 'undefined') {
        const textarea = document.createElement('textarea');
        textarea.value = currentContent;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
      }
      setCopyStatus('Copied');
      if (copyTimerRef.current !== null) {
        window.clearTimeout(copyTimerRef.current);
      }
      copyTimerRef.current = window.setTimeout(() => {
        setCopyStatus('Copy');
      }, 2000);
    } catch {
      // Defensive fallback for clipboard write failures
    }
  };

  return (
    <div className="artifact-viewer-card">
      <div className="artifact-header">
        <h4 className="section-title">Workspace Artifacts</h4>
        <div className="artifact-toolbar">
          <div className="artifact-tabs">
            <button
              type="button"
              className={`artifact-tab-btn ${activeArtifactTab === 'contract' ? 'active' : ''}`}
              onClick={() => handleTabChange('contract')}
            >
              CONTRACT.md {hasContractContent ? '' : '(Empty)'}
            </button>
            <button
              type="button"
              className={`artifact-tab-btn ${activeArtifactTab === 'summary' ? 'active' : ''}`}
              onClick={() => handleTabChange('summary')}
            >
              SUMMARY.md {hasSummaryContent ? '' : '(Empty)'}
            </button>
          </div>
          <div className="view-mode-toggle" role="group" aria-label="View Mode">
            <button
              type="button"
              className={`view-mode-btn ${viewMode === 'rich' ? 'active' : ''}`}
              onClick={() => setViewMode('rich')}
            >
              Rich Document
            </button>
            <button
              type="button"
              className={`view-mode-btn ${viewMode === 'raw' ? 'active' : ''}`}
              onClick={() => setViewMode('raw')}
            >
              Raw Source
            </button>
          </div>
          <button
            type="button"
            className={`artifact-copy-btn ${copyStatus === 'Copied' ? 'copied' : ''}`}
            onClick={handleCopy}
            disabled={!hasCurrentContent}
            title={copyStatus === 'Copied' ? 'Copied to clipboard' : 'Copy Content'}
          >
            {copyStatus}
          </button>
        </div>
      </div>

      <div className="artifact-body">
        {hasCurrentContent && currentContent ? (
          viewMode === 'raw' ? (
            <pre className="artifact-pre">
              <code>{currentContent}</code>
            </pre>
          ) : (
            <div className="artifact-markdown-container">
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={markdownComponents}
              >
                {currentContent}
              </ReactMarkdown>
            </div>
          )
        ) : (
          <div className="artifact-empty-note">
            No {currentTitle} document found for this workspace.
          </div>
        )}
      </div>
    </div>
  );
};
