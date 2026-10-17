import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Navbar } from './Navbar';
import type { ActiveTab } from '../domain/types';

describe('Navbar Component', () => {
  it('renders all four operational navigation tabs', () => {
    render(<Navbar isMock={false} activeTab="overview" />);

    expect(screen.getByRole('button', { name: /fleet overview/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /metric progression/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /comparisons/i })).toBeInTheDocument();
  });

  it('marks the active tab button with the active class', () => {
    const { rerender } = render(<Navbar isMock={false} activeTab="overview" />);
    expect(screen.getByRole('button', { name: /fleet overview/i })).toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="deep-dive" />);
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="progression" />);
    expect(screen.getByRole('button', { name: /metric progression/i })).toHaveClass('active');

    rerender(<Navbar isMock={false} activeTab="comparisons" />);
    expect(screen.getByRole('button', { name: /comparisons/i })).toHaveClass('active');
  });

  it('ensures zero navigation tabs have active class when activeTab is landing', () => {
    const { container } = render(<Navbar isMock={false} activeTab="landing" />);

    expect(screen.getByRole('button', { name: /fleet overview/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /ticket deep dive/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /metric progression/i })).not.toHaveClass('active');
    expect(screen.getByRole('button', { name: /comparisons/i })).not.toHaveClass('active');

    const activeTabs = container.querySelectorAll('.nav-tab.active');
    expect(activeTabs.length).toBe(0);
  });

  it('invokes onSelectTab with landing when brand container, title, subtitle, or icon is clicked', () => {
    const onSelectTab = vi.fn();
    const { container } = render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} />);

    const brandButton = screen.getByRole('button', { name: /agenticlens home/i });
    expect(brandButton).toBeInTheDocument();

    // Click brand container
    fireEvent.click(brandButton);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

    // Click brand title text
    const brandTitle = screen.getByText('AgenticLens');
    fireEvent.click(brandTitle);
    expect(onSelectTab).toHaveBeenLastCalledWith('landing');

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
    render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} />);

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
    render(<Navbar isMock={false} activeTab="landing" />);
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
    render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelectTab} />);

    fireEvent.click(screen.getByRole('button', { name: /ticket deep dive/i }));
    expect(onSelectTab).toHaveBeenCalledWith('deep-dive');

    fireEvent.click(screen.getByRole('button', { name: /metric progression/i }));
    expect(onSelectTab).toHaveBeenCalledWith('progression');

    fireEvent.click(screen.getByRole('button', { name: /comparisons/i }));
    expect(onSelectTab).toHaveBeenCalledWith('comparisons');

    fireEvent.click(screen.getByRole('button', { name: /fleet overview/i }));
    expect(onSelectTab).toHaveBeenCalledWith('overview');
  });

  it('renders Live Server status pill when isMock is false', () => {
    render(<Navbar isMock={false} activeTab="overview" />);
    expect(screen.getByText('Live Server')).toBeInTheDocument();
    expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
  });

  it('renders Mock Mode status pill when isMock is true', () => {
    render(<Navbar isMock={true} activeTab="overview" />);
    expect(screen.getByText('Mock Mode')).toBeInTheDocument();
    expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
  });

  it('contains zero emojis across all navigation states and modes', () => {
    const tabs: readonly ActiveTab[] = ['landing', 'overview', 'deep-dive', 'progression', 'comparisons'];
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}\u{1F000}-\u{1F02F}\u{1FA00}-\u{1FAFF}]/u;

    for (const tab of tabs) {
      const { container: liveContainer } = render(<Navbar isMock={false} activeTab={tab} />);
      expect(emojiRegex.test(liveContainer.textContent ?? '')).toBe(false);

      const { container: mockContainer } = render(<Navbar isMock={true} activeTab={tab} />);
      expect(emojiRegex.test(mockContainer.textContent ?? '')).toBe(false);
    }
  });
});
