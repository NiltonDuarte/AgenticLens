import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EmptyState } from './EmptyState';
import { KPICard } from './KPICard';
import { DistributionCard } from './DistributionCard';
import { Navbar } from './Navbar';
import { TrendChart } from './TrendChart';
import { SaturationBar } from './SaturationBar';
import { ThreadViewer } from './ThreadViewer';
import { ArtifactViewer } from './ArtifactViewer';
import { TicketSelector } from './TicketSelector';

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

  describe('TrendChart Component', () => {
    it('renders empty placeholder when data array is empty', () => {
      render(
        <TrendChart
          title="Empty Trend"
          data={[]}
          series={[{ key: 'val', label: 'Value', color: '#10b981', type: 'line' }]}
          emptyMessage="No data points recorded"
        />
      );

      expect(screen.getByText('Empty Trend')).toBeInTheDocument();
      expect(screen.getByText('No data points recorded')).toBeInTheDocument();
    });

    it('renders svg and series elements when data is present', () => {
      const { container } = render(
        <TrendChart
          title="Cost Trend"
          data={[
            { date: '2026-10-14', cost: 1.2 },
            { date: '2026-10-15', cost: 3.5 },
          ]}
          series={[{ key: 'cost', label: 'Cost', color: '#3b82f6', type: 'line' }]}
          referenceLines={[{ value: 2.0, label: 'Target', color: '#ef4444' }]}
        />
      );

      expect(screen.getByText('Cost Trend')).toBeInTheDocument();
      expect(screen.getByText('Target (2)')).toBeInTheDocument();
      const svg = container.querySelector('svg');
      expect(svg).not.toBeNull();
      const path = container.querySelector('path');
      expect(path).not.toBeNull();
    });

    it('renders bar series rect elements', () => {
      const { container } = render(
        <TrendChart
          title="Invocations Bar"
          data={[
            { date: '2026-10-14', count: 5 },
            { date: '2026-10-15', count: 12 },
          ]}
          series={[{ key: 'count', label: 'Count', color: '#8b5cf6', type: 'bar' }]}
        />
      );

      const rects = container.querySelectorAll('rect');
      expect(rects.length).toBeGreaterThan(0);
    });
  });

  describe('SaturationBar Component', () => {
    it('renders healthy saturation status and 200k ceiling', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'HEALTHY',
            orchestrator_peak_tokens: 50000,
            orchestrator_saturation_pct: 25.0,
            subagents_peak_tokens: 30000,
            subagents_peak_role: 'coder',
            compactions: 0,
            peak_by_role: { orchestrator: 50000, coder: 30000 },
          }}
          healthyCeilingTokens={200000}
        />
      );

      expect(screen.getByText('HEALTHY')).toBeInTheDocument();
      expect(screen.getByText('50k tokens')).toBeInTheDocument();
      expect(screen.getByText('(25.0% of ceiling)')).toBeInTheDocument();
      expect(screen.getByText('Ceiling: 200k tokens')).toBeInTheDocument();
      expect(screen.getByText('orchestrator')).toBeInTheDocument();
      expect(screen.getByText('coder')).toBeInTheDocument();
    });

    it('renders critical saturation badge when status is CRITICAL', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'CRITICAL',
            orchestrator_peak_tokens: 210000,
            orchestrator_saturation_pct: 105.0,
            subagents_peak_tokens: 40000,
            subagents_peak_role: 'coder',
            compactions: 2,
            peak_by_role: {},
          }}
          healthyCeilingTokens={200000}
        />
      );

      expect(screen.getByText('CRITICAL')).toBeInTheDocument();
      expect(screen.getByText(/105\.0%/)).toBeInTheDocument();
    });
  });

  describe('ThreadViewer Component', () => {
    it('renders empty notice when threads array is empty', () => {
      render(<ThreadViewer threads={[]} />);
      expect(
        screen.getByText('No session execution threads recorded for this workspace.')
      ).toBeInTheDocument();
    });

    it('renders thread headers and expands child models on toggle click', () => {
      const dummyThread = {
        session_id: 'ses-root-01',
        user: 'dev',
        compactions: 0,
        roles: [
          {
            role: 'coder',
            models: [
              {
                model: 'gemini-3.8-flash',
                invocations: 4,
                input_tokens: 1000,
                output_tokens: 200,
                cache_read_tokens: 500,
                cache_hit_rate_pct: 33.3,
                cost_usd: 0.85,
                peak_context_tokens: 1200,
                context_health: 'HEALTHY',
              },
            ],
          },
        ],
      };

      render(<ThreadViewer threads={[dummyThread]} />);

      expect(screen.getByText('ses-root-01')).toBeInTheDocument();
      expect(screen.getByText('User: dev')).toBeInTheDocument();
      expect(screen.getByText('gemini-3.8-flash')).toBeInTheDocument();
      expect(screen.getByText('coder')).toBeInTheDocument();

      // Click to toggle collapse
      const toggleBtn = screen.getByRole('button', { name: /ses-root-01/i });
      fireEvent.click(toggleBtn);
      expect(screen.queryByText('gemini-3.8-flash')).not.toBeInTheDocument();

      // Click to expand again
      fireEvent.click(toggleBtn);
      expect(screen.getByText('gemini-3.8-flash')).toBeInTheDocument();
    });
  });

  describe('ArtifactViewer Component', () => {
    it('renders markdown artifact content and toggles between contract and summary', () => {
      render(
        <ArtifactViewer
          artifacts={{
            ticket_id: 'AL-1',
            contract_content: '# Contract Spec\nScope defined here.',
            summary_content: '# Summary Report\nAll tests pass.',
          }}
        />
      );

      expect(screen.getByText(/Scope defined here/)).toBeInTheDocument();

      const summaryTabBtn = screen.getByRole('button', { name: /summary\.md/i });
      fireEvent.click(summaryTabBtn);

      expect(screen.getByText(/All tests pass/)).toBeInTheDocument();
    });

    it('renders empty notice when content is null', () => {
      render(<ArtifactViewer artifacts={null} />);
      expect(
        screen.getByText('No CONTRACT.md document found for this workspace.')
      ).toBeInTheDocument();

      const summaryTabBtn = screen.getByRole('button', { name: /summary\.md/i });
      fireEvent.click(summaryTabBtn);
      expect(
        screen.getByText('No SUMMARY.md document found for this workspace.')
      ).toBeInTheDocument();
    });
  });

  describe('TicketSelector Component', () => {
    it('renders options for all tickets and calls onSelect on change', () => {
      const onSelect = vi.fn();
      render(
        <TicketSelector
          tickets={[
            { ticket_id: 'AL-1', ticket_title: 'Ticket One' },
            { ticket_id: 'AL-2', ticket_title: 'Ticket Two' },
          ]}
          selectedTicketId="AL-1"
          onSelectTicket={onSelect}
        />
      );

      const select = screen.getByLabelText(/workspace autopsy/i);
      expect(select).toBeInTheDocument();
      expect(screen.getByText('AL-1 - Ticket One')).toBeInTheDocument();
      expect(screen.getByText('AL-2 - Ticket Two')).toBeInTheDocument();

      fireEvent.change(select, { target: { value: 'AL-2' } });
      expect(onSelect).toHaveBeenCalledWith('AL-2');
    });
  });
});
