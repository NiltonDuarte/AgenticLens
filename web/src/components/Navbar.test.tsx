import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { Navbar } from './Navbar';
import type { ActiveTab, BuildInfo } from '../domain/types';
import { mockBuildInfo, mockProjectConfig } from '../test/fixtures/mockData';

describe('Navbar Component', () => {
  it('renders all five operational navigation tabs', () => {
    render(<Navbar isMock={false} activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

    expect(screen.getByRole('button', { name: /fleet overview/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /metric progression/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /comparisons/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /live observability/i })).toBeInTheDocument();
  });

  it('marks the active tab button with the active class', () => {
    const { rerender } = render(<Navbar isMock={false} activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /fleet overview/i })).toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="deep-dive" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="progression" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /metric progression/i })).toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="comparisons" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /comparisons/i })).toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="live" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByRole('button', { name: /live observability/i })).toHaveClass('active');
  });

  it('exclusively activates the live observability tab when activeTab is live while deactivating all other tabs', () => {
    const { container } = render(<Navbar isMock={false} activeTab="live" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

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
    const { container } = render(<Navbar isMock={false} activeTab="landing" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

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
      <Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />
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
    render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

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
    render(<Navbar isMock={false} activeTab="landing" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
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
    render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);

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

  it('renders Live Server status pill when isMock is false', () => {
    render(<Navbar isMock={false} activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByText('Live Server')).toBeInTheDocument();
    expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
  });

  it('renders Mock Mode status pill when isMock is true', () => {
    render(<Navbar isMock={true} activeTab="overview" versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByText('Mock Mode')).toBeInTheDocument();
    expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
  });

  it('renders version badge with provided versionInfo prop and displays tooltip', () => {
    const versionInfo: BuildInfo = {
      version: 'v1.4.2',
      commit: 'fedcba9',
      build_date: '2026-10-18T14:00:00Z',
    };
    render(<Navbar isMock={false} activeTab="overview" versionInfo={versionInfo} projectConfig={mockProjectConfig} />);

    const badge = screen.getByText('v1.4.2');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');
    expect(badge).toHaveAttribute('title', 'Commit: fedcba9 | Built: 2026-10-18T14:00:00Z');
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

    render(<Navbar isMock={false} activeTab="overview" projectConfig={mockProjectConfig} />);

    const badge = await screen.findByText('v2.0.0');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');
    expect(badge).toHaveAttribute('title', 'Commit: abcdef1 | Built: 2026-10-18T10:00:00Z');

    vi.unstubAllGlobals();
  });

  it('renders fallback version badge when fetch fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('Network error')));

    render(<Navbar isMock={false} activeTab="overview" projectConfig={mockProjectConfig} />);

    const badge = await screen.findByText('v0.1.0-dev');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('version-badge');

    vi.unstubAllGlobals();
  });

  it('dynamically updates version badge when versionInfo prop changes on rerender', () => {
    const initialInfo: BuildInfo = {
      version: 'v1.0.0',
      commit: '1111111',
      build_date: '2026-10-18T01:00:00Z',
    };
    const updatedInfo: BuildInfo = {
      version: 'v2.0.0',
      commit: '2222222',
      build_date: '2026-10-18T02:00:00Z',
    };

    const { rerender } = render(<Navbar isMock={false} activeTab="overview" versionInfo={initialInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByText('v1.0.0')).toBeInTheDocument();
    expect(screen.getByText('v1.0.0')).toHaveAttribute('title', 'Commit: 1111111 | Built: 2026-10-18T01:00:00Z');

    rerender(<Navbar isMock={false} activeTab="overview" versionInfo={updatedInfo} projectConfig={mockProjectConfig} />);
    expect(screen.getByText('v2.0.0')).toBeInTheDocument();
    expect(screen.getByText('v2.0.0')).toHaveAttribute('title', 'Commit: 2222222 | Built: 2026-10-18T02:00:00Z');
    expect(screen.queryByText('v1.0.0')).not.toBeInTheDocument();
  });

  it('renders complex semver tags with prerelease and build metadata in badge', () => {
    const complexInfo: BuildInfo = {
      version: 'v2.1.0-rc.3+20261018.commit.abcdef0-dirty',
      commit: 'abcdef0',
      build_date: '2026-10-18T19:45:00.123Z',
    };
    render(<Navbar isMock={false} activeTab="overview" versionInfo={complexInfo} projectConfig={mockProjectConfig} />);

    const badge = screen.getByText('v2.1.0-rc.3+20261018.commit.abcdef0-dirty');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveAttribute('title', 'Commit: abcdef0 | Built: 2026-10-18T19:45:00.123Z');
  });

  it('displays fallback v0.1.0-dev immediately prior to async resolution and updates upon resolve', async () => {
    let resolvePromise: (value: { ok: boolean; json: () => Promise<BuildInfo> }) => void = () => {};
    const pendingPromise = new Promise<{ ok: boolean; json: () => Promise<BuildInfo> }>((resolve) => {
      resolvePromise = resolve;
    });

    vi.stubGlobal('fetch', vi.fn().mockReturnValue(pendingPromise));

    render(<Navbar isMock={false} activeTab="overview" projectConfig={mockProjectConfig} />);

    // Prior to resolution, fallback is immediately visible
    expect(screen.getByText('v0.1.0-dev')).toBeInTheDocument();

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

  it('renders project badge when projectName prop is provided', () => {
    render(
      <Navbar
        isMock={false}
        activeTab="overview"
        versionInfo={mockBuildInfo} projectConfig={mockProjectConfig}
        projectName="CustomObservedProject"
      />
    );

    const badge = screen.getByText('CustomObservedProject');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('project-badge');
    expect(badge).toHaveAttribute('title', 'Observed Project: CustomObservedProject');
    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();
  });

  it('renders project badge when projectConfig prop is provided', () => {
    render(
      <Navbar
        isMock={false}
        activeTab="overview"
        versionInfo={mockBuildInfo}
        projectConfig={{
          ...mockProjectConfig,
          project_name: 'ConfiguredRepo',
        }}
      />
    );

    const badge = screen.getByText('ConfiguredRepo');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveClass('project-badge');
    expect(badge).toHaveAttribute('title', 'Observed Project: ConfiguredRepo');
    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();
  });

  it('invokes onSelectTab with landing when project badge is clicked', () => {
    const onSelectTab = vi.fn();
    render(
      <Navbar
        isMock={false}
        activeTab="overview"
        onSelectTab={onSelectTab}
        versionInfo={mockBuildInfo} projectConfig={mockProjectConfig}
        projectName="ClickableProject"
      />
    );

    const badge = screen.getByText('ClickableProject');
    fireEvent.click(badge);
    expect(onSelectTab).toHaveBeenCalledWith('landing');
  });

  it('preserves permanent AgenticLens and Telemetry branding when project is empty', () => {
    const { container } = render(
      <Navbar
        isMock={false}
        activeTab="overview"
        versionInfo={mockBuildInfo}
        projectConfig={mockProjectConfig}
        projectName=""
      />
    );

    expect(screen.getByText('AgenticLens')).toBeInTheDocument();
    expect(screen.getByText('Telemetry')).toBeInTheDocument();
    expect(container.querySelector('.project-badge')).toBeNull();
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
        isMock={false}
        activeTab="overview"
        versionInfo={mockBuildInfo}
      />
    );

    const badge = await screen.findByText('FetchedProject');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveAttribute('title', 'Observed Project: FetchedProject');

    vi.unstubAllGlobals();
  });

  it('contains zero emojis across all navigation states and modes', () => {
    const tabs: readonly ActiveTab[] = ['landing', 'overview', 'deep-dive', 'progression', 'comparisons', 'live'];
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}\u{1F000}-\u{1F02F}\u{1FA00}-\u{1FAFF}]/u;

    for (const tab of tabs) {
      const { container: liveContainer } = render(
        <Navbar isMock={false} activeTab={tab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />
      );
      expect(emojiRegex.test(liveContainer.textContent ?? '')).toBe(false);
      expect(emojiRegex.test(liveContainer.innerHTML)).toBe(false);

      const { container: mockContainer } = render(
        <Navbar isMock={true} activeTab={tab} versionInfo={mockBuildInfo} projectConfig={mockProjectConfig} />
      );
      expect(emojiRegex.test(mockContainer.textContent ?? '')).toBe(false);
      expect(emojiRegex.test(mockContainer.innerHTML)).toBe(false);
    }
  });
});
