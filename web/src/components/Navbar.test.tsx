import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { Navbar } from './Navbar';
import type { BuildInfo } from '../domain/types';
import { mockBuildInfo, mockProjectConfig } from '../test/fixtures/mockData';

describe('Navbar Component', () => {
  it('renders all five operational navigation tabs', () => {
    render(<Navbar activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    expect(screen.getByRole('button', { name: /fleet overview/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /metric progression/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /comparisons/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /live observability/i })).toBeInTheDocument();
  });

  it('marks the active tab button with the active class', () => {
    const { rerender } = render(<Navbar activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /fleet overview/i })).toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');

    rerender(<Navbar activeTab="deep-dive" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toHaveClass('active');

    rerender(<Navbar activeTab="progression" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /metric progression/i })).toHaveClass('active');

    rerender(<Navbar activeTab="comparisons" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /comparisons/i })).toHaveClass('active');

    rerender(<Navbar activeTab="live" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /live observability/i })).toHaveClass('active');
  });

  it('exclusively activates the live observability tab when activeTab is live while deactivating all other tabs', () => {
    const { container } = render(<Navbar activeTab="live" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    expect(screen.getByRole('button', { name: /live observability/i })).toHaveClass('active');
    expect(screen.getByRole('button', { name: /fleet overview/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /metric progression/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /comparisons/i })).not.toHaveClass('active');

    const activeButtons = container.querySelectorAll('.nav-tab.active');
    expect(activeButtons.length).toBe(1);
    expect(activeButtons[0]).toHaveTextContent('Live Observability');
  });

  it('ensures zero navigation tabs have active class when activeTab is landing', () => {
    const { container } = render(<Navbar activeTab="landing" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    expect(screen.getByRole('button', { name: /fleet overview/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /metric progression/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /comparisons/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /live observability/i })).not.toHaveClass('active');

    const activeTabs = container.querySelectorAll('.nav-tab.active');
    expect(activeTabs.length).toBe(0);
  });

  it('invokes onSelectTab with landing when brand container, title, subtitle, or icon is clicked', () => {
    const onSelectTab = vi.fn();
    const { container } = render(
      <Navbar activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />
    );

    const brandButton = screen.getByRole('button', { name: /agenticlens home/i });
    expect(brandButton).toBeInTheDocument();

    // Click brand container
    fireEvent.click(brandButton);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

    // Click brand title text
    const brandTitle = container.querySelector('.brand-title');
    expect(brandTitle).not.toBeNull();
    if (brandTitle) {
      fireEvent.click(brandTitle);
      expect(onSelectTab).toHaveBeenLastCalledWith('landing');
    }

    // Click brand subtitle text
    const brandSubtitle = screen.getByText('Telemetry');
    fireEvent.click(brandSubtitle);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

    // Click brand SVG icon
    const svgIcon = container.querySelector('.brand-icon');
    expect(svgIcon).not.toBeNull();
    if (svgIcon) {
      fireEvent.click(svgIcon);
      expect(onSelectTab).toHaveBeenLastCalledWith('landing');
    }

    expect(onSelectTab).toHaveBeenCalledTimes(4);
  });

  it('invokes onSelectTab with landing on Enter or Space key and ignores other keys', () => {
    const onSelectTab = vi.fn();
    render(<Navbar activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    const brandButton = screen.getByRole('button', { name: /agenticlens home/i });

    // Enter key triggers
    fireEvent.keyDown(brandButton, { key: 'Enter' });
    expect(onSelectTab).toHaveBeenCalledTimes(1);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

    // Space key triggers
    fireEvent.keyDown(brandButton, { key: ' ' });
    expect(onSelectTab).toHaveBeenCalledTimes(2);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

    // Non-triggering keys
    fireEvent.keyDown(brandButton, { key: 'Tab' });
    fireEvent.keyDown(brandButton, { key: 'Escape' });
    fireEvent.keyDown(brandButton, { key: 'ArrowDown' });
    fireEvent.keyDown(brandButton, { key: 'a' });
    expect(onSelectTab).toHaveBeenCalledTimes(2);
  });

  it('handles brand and tab activations safely when onSelectTab callback is omitted', () => {
    render(<Navbar activeTab="landing" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    const brandButton = screen.getByRole('button', { name: /agenticlens home/i });

    expect(() => {
      fireEvent.click(brandButton);
      fireEvent.keyDown(brandButton, { key: 'Enter' });
      fireEvent.keyDown(brandButton, { key: ' ' });
      fireEvent.click(screen.getByRole('button', { name: /fleet overview/i }));
    }).not.toThrow();
  });

  it('invokes onSelectTab when user clicks navigation tabs', () => {
    const onSelectTab = vi.fn();
    render(<Navbar activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    fireEvent.click(screen.getByRole('button', { name: /ticket deep dive/i }));
    expect(onSelectTab).toHaveBeenCalledWith('deep-dive');

    fireEvent.click(screen.getByRole('button', { name: /metric progression/i }));
    expect(onSelectTab).toHaveBeenCalledWith('progression');

    fireEvent.click(screen.getByRole('button', { name: /comparisons/i }));
    expect(onSelectTab).toHaveBeenCalledWith('comparisons');

    fireEvent.click(screen.getByRole('button', { name: /live observability/i }));
    expect(onSelectTab).toHaveBeenCalledWith('live');

    fireEvent.click(screen.getByRole('button', { name: /fleet overview/i }));
    expect(onSelectTab).toHaveBeenCalledWith('overview');
  });

  it('renders active workspaces directory when projectConfig is provided', () => {
    render(
      <Navbar
        activeTab="overview"
        versionInfo={mockBuildInfo}
        projectConfig={{
          ...mockProjectConfig,
          workspaces_dir: 'custom/workspaces',
        }}
      />
    );
    expect(screen.getByText('custom/workspaces')).toBeInTheDocument();
    const pill = screen.getByText('custom/workspaces').closest('.status-pill.workspace-pill');
    expect(pill).toBeInTheDocument();
    expect(pill).toHaveAttribute('title', 'Active Workspaces Directory: custom/workspaces');
    expect(pill).toHaveAttribute('aria-label', 'Active Workspaces Directory: custom/workspaces');
    expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
    expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
  });

  it('renders loading indicator and does not guess workspaces when projectConfig is undefined or empty', () => {
    render(
      <Navbar
        activeTab="overview"
        versionInfo={mockBuildInfo}
        projectConfig={undefined}
      />
    );
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    const pill = screen.getByText('Loading...').closest('.status-pill.workspace-pill.loading');
    expect(pill).toBeInTheDocument();
    expect(pill).toHaveAttribute('title', 'Loading workspaces configuration');
    expect(pill).toHaveAttribute('aria-label', 'Loading workspaces configuration');
    expect(screen.queryByText('workspaces')).not.toBeInTheDocument();
    expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
    expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
  });

  it('renders loading placeholder and does not guess workspaces when workspaces_dir is empty string', () => {
    render(
      <Navbar
        activeTab="overview"
        versionInfo={mockBuildInfo}
        projectConfig={{
          ...mockProjectConfig,
          workspaces_dir: '',
        }}
      />
    );
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.queryByText('workspaces')).not.toBeInTheDocument();
  });

  it('renders fallback version badge and resolves via fetch when versionInfo is omitted', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        version: 'v2.0.0',
        commit: 'abcdef1',
        build_date: '2026-10-18T10:00:00Z',
      }),
    }));

    render(<Navbar activeTab="overview" projectConfig={mockProjectConfig} />);

    const badge = await screen.findByText('v2.0.0');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');
    expect(badge).toHaveAttribute('title', 'Commit: abcdef1 | Built: 2026-10-18T10:00:00Z');

    vi.unstubAllGlobals();
  });

  it('renders fallback version badge when fetch fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

    render(<Navbar activeTab="overview" projectConfig={mockProjectConfig} />);

    const badge = await screen.findByText('v0.0.0-dev');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');

    vi.unstubAllGlobals();
  });

  it('renders complex semver tags with prerelease and build metadata in badge', () => {
    const complexInfo: BuildInfo = {
      version: 'v2.1.0-rc.3+20261018.commit.abcdef0-dirty',
      commit: 'abcdef0',
      build_date: '2026-10-18T19:45:00.123Z',
    };
    const { rerender } = render(<Navbar activeTab="overview" versionInfo={complexInfo} projectConfig={mockProjectConfig} />);

    const badge = screen.getByText('v2.1.0-rc.3+20261018.commit.abcdef0-dirty');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');
    expect(badge).toHaveAttribute('title', 'Commit: abcdef0 | Built: 2026-10-18T19:45:00.123Z');
    const updatedInfo: BuildInfo = {
      version: 'v2.2.0',
      commit: '2222222',
      build_date: '2026-10-19T02:00:00Z',
    };
    rerender(<Navbar activeTab="overview" versionInfo={updatedInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByText('v2.2.0')).toBeInTheDocument();
    expect(screen.getByText('v2.2.0')).toHaveAttribute('title', 'Commit: 2222222 | Built: 2026-10-19T02:00:00Z');
    expect(screen.queryByText(complexInfo.version)).not.toBeInTheDocument();
  });

  it('displays fallback v0.0.0-dev immediately prior to async resolution and updates upon resolve', async () => {
    let resolvePromise: (value: { ok: boolean; json: () => Promise<BuildInfo> }) => void = () => {};
    const pendingPromise = new Promise<{ ok: boolean; json: () => Promise<BuildInfo> }>((resolve) => {
      resolvePromise = resolve;
    });

    vi.stubGlobal('fetch', vi.fn().mockReturnValue(pendingPromise));

    render(<Navbar activeTab="overview" projectConfig={mockProjectConfig} />);

    // Prior to resolution, fallback is immediately visible
    expect(screen.getByText('v0.0.0-dev')).toBeInTheDocument();

    // Resolve promise wrapped in act
    await act(async () => {
      resolvePromise({
        ok: true,
        json: async () => ({
          version: 'v3.5.0',
          commit: '9999999',
          build_date: '2026-10-18T20:00:00Z',
        }),
      });
    });

    const badge = await screen.findByText('v3.5.0');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveAttribute('title', 'Commit: 9999999 | Built: 2026-10-18T20:00:00Z');

    vi.unstubAllGlobals();
  });

  it('invokes onSelectTab with landing when project badge is clicked', () => {
    const onSelectTab = vi.fn();
    const { container, rerender } = render(
      <Navbar
        activeTab="overview"
        onSelectTab={onSelectTab}
        versionInfo={mockBuildInfo} projectConfig={mockProjectConfig}
        projectName="ClickableProject"
      />
    );

    const badge = screen.getByText('ClickableProject');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('project-badge');
    expect(badge).toHaveAttribute('title', 'Observed Project: ClickableProject');
    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();
    fireEvent.click(badge);
    expect(onSelectTab).toHaveBeenCalledWith('landing');
    const projectConfig = { ...mockProjectConfig, project_name: 'ConfiguredRepo' };
    rerender(<Navbar activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={projectConfig} />);
    const configuredBadge = screen.getByText('ConfiguredRepo');
    expect(configuredBadge).toBeInTheDocument();
    expect(configuredBadge).toHaveClass('project-badge');
    expect(configuredBadge).toHaveAttribute('title', 'Observed Project: ConfiguredRepo');
    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();

    rerender(<Navbar activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={projectConfig} projectName="" />);
    expect(container.querySelector('.project-badge')).toBeNull();
    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();
  });

  it('renders fallback project badge and resolves via fetch when projectConfig and projectName are omitted', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        project_name: 'FetchedProject',
        prefixes: [],
        workspaces_dir: 'workspaces',
        role_order: [],
        version: 'v1.0.0',
      }),
    }));

    render(
      <Navbar
        activeTab="overview"
        versionInfo={mockBuildInfo}
      />
    );

    const badge = await screen.findByText('FetchedProject');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveAttribute('title', 'Observed Project: FetchedProject');

    vi.unstubAllGlobals();
  });
});
