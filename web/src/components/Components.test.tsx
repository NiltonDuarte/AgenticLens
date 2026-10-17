import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EmptyState } from './EmptyState';
import { KPICard } from './KPICard';
import { DistributionCard } from './DistributionCard';
import { Navbar } from './Navbar';

describe('Presentation Components Boundary Tests', () => {
  describe('EmptyState Component', () => {
    it('renders title and message without action button when omitted', () => {
      render(
        <EmptyState
          title="No Data Found"
          message="Detailed explanation of missing data."
        />
      );

      expect(screen.getByText('No Data Found')).toBeInTheDocument();
      expect(screen.getByText('Detailed explanation of missing data.')).toBeInTheDocument();
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('renders action button and triggers callback on click', () => {
      const onAction = vi.fn();
      render(
        <EmptyState
          title="Empty Test"
          message="Click below"
          actionLabel="Execute Action"
          onAction={onAction}
        />
      );

      const button = screen.getByRole('button', { name: 'Execute Action' });
      expect(button).toBeInTheDocument();
      fireEvent.click(button);
      expect(onAction).toHaveBeenCalledTimes(1);
    });
  });

  describe('KPICard Component', () => {
    it('renders value and tooltip, falling back to dash for empty value', () => {
      const { rerender } = render(
        <KPICard title="Total Cost" value="$12.50" secondary="2 tickets" tooltip="Test tooltip" />
      );

      expect(screen.getByText('Total Cost')).toBeInTheDocument();
      expect(screen.getByText('$12.50')).toBeInTheDocument();
      expect(screen.getByText('2 tickets')).toBeInTheDocument();

      rerender(<KPICard title="Zero Card" value="" />);
      expect(screen.getByText('-')).toBeInTheDocument();
    });
  });

  describe('DistributionCard Component', () => {
    it('handles zero total cost without divide-by-zero distortion', () => {
      render(
        <DistributionCard
          title="Role Cost Distribution"
          items={{ coder: 0, tester: 0 }}
          totalCostUSD={0}
        />
      );

      expect(screen.getByText('Role Cost Distribution')).toBeInTheDocument();
      expect(screen.getByText('coder')).toBeInTheDocument();
      expect(screen.getByText('tester')).toBeInTheDocument();
      const zeroPcts = screen.getAllByText('(0.0%)');
      expect(zeroPcts.length).toBeGreaterThanOrEqual(2);
    });

    it('renders empty message when items dictionary is empty', () => {
      render(
        <DistributionCard
          title="Empty Distribution"
          items={{}}
          totalCostUSD={0}
        />
      );

      expect(screen.getByText('No distribution data recorded')).toBeInTheDocument();
    });

    it('sorts distribution items descending by cost', () => {
      render(
        <DistributionCard
          title="Model Costs"
          items={{ 'cheap-model': 1.5, 'expensive-model': 15.0, 'mid-model': 5.0 }}
          totalCostUSD={21.5}
        />
      );

      const items = screen.getAllByText(/-model/);
      expect(items[0]).toHaveTextContent('expensive-model');
      expect(items[1]).toHaveTextContent('mid-model');
      expect(items[2]).toHaveTextContent('cheap-model');
    });
  });

  describe('Navbar Component', () => {
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

    it('dispatches onSelectTab when tab is clicked', () => {
      const onSelect = vi.fn();
      render(<Navbar isMock={false} activeTab="overview" onSelectTab={onSelect} />);

      const overviewBtn = screen.getByRole('button', { name: 'Fleet Overview' });
      fireEvent.click(overviewBtn);
      expect(onSelect).toHaveBeenCalledWith('overview');
    });
  });
});
