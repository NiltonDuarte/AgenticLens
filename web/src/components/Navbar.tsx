import React, { useEffect, useState } from 'react';
import type { ActiveTab, BuildInfo, ProjectConfigResponse } from '../domain/types';
import { fetchVersion, fetchProjectConfig } from '../api/client';

export interface NavbarProps {
  readonly activeTab: ActiveTab;
  readonly onSelectTab?: ((tab: ActiveTab) => void) | undefined;
  readonly versionInfo?: BuildInfo | undefined;
  readonly projectConfig?: ProjectConfigResponse | undefined;
  readonly projectName?: string | undefined;
}

const DEFAULT_BUILD_INFO: BuildInfo = {
  version: 'v0.0.0-dev',
  commit: 'unknown',
  build_date: 'unknown',
};

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  onSelectTab,
  versionInfo,
  projectConfig,
  projectName,
}) => {
  const [internalVersion, setInternalVersion] = useState<BuildInfo | null>(versionInfo ?? null);
  const [internalConfig, setInternalConfig] = useState<ProjectConfigResponse | null>(
    projectConfig ?? null
  );

  const [configError, setConfigError] = useState<string | null>(null);
  const [configAttempt, setConfigAttempt] = useState(0);

  useEffect(() => {
    if (versionInfo) {
      setInternalVersion(versionInfo);
      return;
    }
    let mounted = true;
    void fetchVersion()
      .then((resp) => {
        if (mounted) {
          setInternalVersion(resp.data);
        }
      })
      .catch(() => {
        // Keep default build info on failure
      });
    return () => {
      mounted = false;
    };
  }, [versionInfo]);

  useEffect(() => {
    if (projectConfig) {
      setConfigError(null);
      setInternalConfig(projectConfig);
      return;
    }
    if (projectName !== undefined) {
      return;
    }
    let mounted = true;
    setConfigError(null);
    void fetchProjectConfig()
      .then((resp) => {
        if (mounted) {
          setInternalConfig(resp.data);
        }
      })
      .catch((error: unknown) => {
        if (mounted) setConfigError(error instanceof Error ? error.message : 'Failed to load workspace configuration');
      });
    return () => {
      mounted = false;
    };
  }, [projectConfig, projectName, configAttempt]);

  const displayVersion = internalVersion ?? DEFAULT_BUILD_INFO;
  const observedProject = (projectName ?? internalConfig?.project_name ?? '').trim();
  const workspaceRoots = projectConfig?.workspaces_roots ?? internalConfig?.workspaces_roots ?? [];

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
          {observedProject ? (
            <span
              className="project-badge"
              title={`Observed Project: ${observedProject}`}
            >
              {observedProject}
            </span>
          ) : null}
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
        <button
          type="button"
          className={`nav-tab ${activeTab === 'live' ? 'active' : ''}`}
          onClick={() => onSelectTab?.('live')}
        >
          Live Observability
        </button>
      </nav>

      <div className="navbar-actions">
        <span
          className="version-badge"
          title={`Commit: ${displayVersion.commit} | Built: ${displayVersion.build_date}`}
        >
          {displayVersion.version}
        </span>
        {workspaceRoots.length > 0 ? (
          <div className="workspace-roots" aria-label="Configured workspace directories">
            {workspaceRoots.map((root) => <span
              key={root.path}
              className="status-pill workspace-pill"
              title={`${root.path}; maximum depth ${root.discovery.max_depth}; directory budget ${root.discovery.max_directories}`}
              aria-label={`Workspace directory: ${root.path}`}
            >
              <span className="status-dot" />
              {root.path}
            </span>)}
          </div>
        ) : configError ? (
          <span className="status-pill workspace-pill" role="alert" title={configError}>
            Workspace configuration unavailable
            <button type="button" onClick={() => setConfigAttempt((attempt) => attempt + 1)}>Retry configuration</button>
          </span>
        ) : (
          <span
            className="status-pill workspace-pill loading"
            title="Loading workspaces configuration"
            aria-label="Loading workspaces configuration"
          >
            <span className="status-dot loading" />
            Loading...
          </span>
        )}
      </div>
    </header>
  );
};
