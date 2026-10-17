import React from 'react';
import type { ActiveTab } from '../domain/types';

export interface NavbarProps {
  readonly isMock: boolean;
  readonly activeTab: ActiveTab;
  readonly onSelectTab?: (tab: ActiveTab) => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  isMock,
  activeTab,
  onSelectTab,
}) => {
  const handleBrandKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onSelectTab?.('landing');
    }
  };

  return (
    <header className="navbar">
      <div
        className="navbar-brand"
        role="button"
        tabIndex={0}
        aria-label="AgenticLens Home"
        onClick={() => onSelectTab?.('landing')}
        onKeyDown={handleBrandKeyDown}
      >
        <svg
          className="brand-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <circle cx="12" cy="12" r="10" />
          <path d="m10 15 5-3-5-3v6Z" />
        </svg>
        <div>
          <span className="brand-title">AgenticLens</span>
          <span className="brand-subtitle">Telemetry</span>
        </div>
      </div>

      <nav className="nav-links">
        <button
          type="button"
          className={`nav-tab ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('overview')}
        >
          Fleet Overview
        </button>
        <button
          type="button"
          className={`nav-tab ${activeTab === 'deep-dive' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('deep-dive')}
        >
          Ticket Deep Dive
        </button>
        <button
          type="button"
          className={`nav-tab ${activeTab === 'progression' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('progression')}
        >
          Metric Progression
        </button>
        <button
          type="button"
          className={`nav-tab ${activeTab === 'comparisons' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('comparisons')}
        >
          Comparisons
        </button>
      </nav>

      <div>
        {isMock ? (
          <span className="status-pill mock" title="Running with local deterministic mock fixtures">
            <span className="status-dot" />
            Mock Mode
          </span>
        ) : (
          <span className="status-pill live" title="Connected to live AgenticLens Go HTTP server">
            <span className="status-dot" />
            Live Server
          </span>
        )}
      </div>
    </header>
  );
};
