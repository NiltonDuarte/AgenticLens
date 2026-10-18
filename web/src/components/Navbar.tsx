import React, { useEffect, useState } from 'react';
import type { ActiveTab, BuildInfo } from '../domain/types';
import { fetchVersion } from '../api/client';

export interface NavbarProps {
  readonly isMock: boolean;
  readonly activeTab: ActiveTab;
  readonly onSelectTab?: (tab: ActiveTab) => void;
  readonly versionInfo?: BuildInfo;
}

const DEFAULT_BUILD_INFO: BuildInfo = {
  version: 'v0.1.0-dev',
  commit: 'unknown',
  build_date: 'unknown',
};

export const Navbar: React.FC<NavbarProps> = ({
  isMock,
  activeTab,
  onSelectTab,
  versionInfo,
}) => {
  const [internalVersion, setInternalVersion] = useState<BuildInfo | null>(versionInfo ?? null);

  useEffect(() => {
    if (versionInfo) {
      setInternalVersion(versionInfo);
      return;
    }
    let mounted = true;
    void fetchVersion().then((resp) => {
      if (mounted) {
        setInternalVersion(resp.data);
      }
    });
    return () => {
      mounted = false;
    };
  }, [versionInfo]);

  const displayVersion = internalVersion ?? DEFAULT_BUILD_INFO;

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

      <div className="navbar-actions">
        <span
          className="version-badge"
          title={`Commit: ${displayVersion.commit} | Built: ${displayVersion.build_date}`}
        >
          {displayVersion.version}
        </span>
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
