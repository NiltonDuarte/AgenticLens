import React from 'react';

export interface NavbarProps {
  readonly isMock: boolean;
  readonly activeTab: 'overview' | 'detail';
  readonly onSelectTab?: (tab: 'overview' | 'detail') => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  isMock,
  activeTab,
  onSelectTab,
}) => {
  return (
    <header className="navbar">
      <div className="navbar-brand">
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
          className={`nav-tab ${activeTab === 'detail' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('detail')}
          disabled
          title="Ticket Deep Dive will be available in AL-4"
        >
          Ticket Deep Dive
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
