import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { EmptyState } from './EmptyState';
import { KPICard } from './KPICard';
import { DistributionCard } from './DistributionCard';
import { InfoTooltip, InfoIcon } from './InfoTooltip';
import { Navbar } from './Navbar';
import { TrendChart } from './TrendChart';
import { SaturationBar } from './SaturationBar';
import { ThreadViewer } from './ThreadViewer';
import { ArtifactViewer } from './ArtifactViewer';
import { TicketSelector } from './TicketSelector';
import type { ThreadExecutionRecord } from '../domain/types';

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
      expect(screen.getByRole('tooltip')).toHaveTextContent('Test tooltip');
      expect(
        screen.getByRole('button', { name: 'Information about Total Cost' })
      ).toBeInTheDocument();

      rerender(<KPICard title="Zero Card" value="" />);
      expect(screen.getByText('-')).toBeInTheDocument();
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    });

    it('renders zero numeric, currency, and token quantities without falling back to dash', () => {
      const { rerender } = render(<KPICard title="Zero Numeric" value="0" />);
      expect(screen.getByText('0')).toBeInTheDocument();
      expect(screen.queryByText('-')).not.toBeInTheDocument();

      rerender(<KPICard title="Zero Tokens" value="0 tokens" />);
      expect(screen.getByText('0 tokens')).toBeInTheDocument();

      rerender(<KPICard title="Zero Dollars" value="$0.00" />);
      expect(screen.getByText('$0.00')).toBeInTheDocument();

      rerender(<KPICard title="Empty String" value="" />);
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

    it('renders custom roles with color indicators and correct percentage breakdown', () => {
      render(
        <DistributionCard
          title="Role Cost Distribution"
          items={{ advisor: 2.5, po: 5.0, researcher: 2.5 }}
          totalCostUSD={10.0}
        />
      );

      expect(screen.getByText('po')).toBeInTheDocument();
      expect(screen.getByText('advisor')).toBeInTheDocument();
      expect(screen.getByText('researcher')).toBeInTheDocument();
      expect(screen.getByText('(50.0%)')).toBeInTheDocument();
      const twentyFives = screen.getAllByText('(25.0%)');
      expect(twentyFives.length).toBe(2);
    });

    it('renders arbitrary models with correct labels and percentages', () => {
      render(
        <DistributionCard
          title="Model Cost Distribution"
          items={{ 'claude-3-7-sonnet': 7.5, 'gemini-2.5-pro': 2.5 }}
          totalCostUSD={10.0}
        />
      );

      expect(screen.getByText('claude-3-7-sonnet')).toBeInTheDocument();
      expect(screen.getByText('gemini-2.5-pro')).toBeInTheDocument();
      expect(screen.getByText('(75.0%)')).toBeInTheDocument();
      expect(screen.getByText('(25.0%)')).toBeInTheDocument();
    });

    it('handles zero total cost without NaN or division by zero', () => {
      render(
        <DistributionCard
          title="Role Cost Distribution"
          items={{ advisor: 0, po: 0 }}
          totalCostUSD={0}
        />
      );

      expect(screen.getByText('advisor')).toBeInTheDocument();
      expect(screen.getByText('po')).toBeInTheDocument();
      const zeroPcts = screen.getAllByText('(0.0%)');
      expect(zeroPcts.length).toBe(2);
      expect(screen.queryByText(/NaN/)).not.toBeInTheDocument();
    });

    it('renders empty notice when items dictionary is empty', () => {
      render(
        <DistributionCard
          title="Role Cost Distribution"
          items={{}}
          totalCostUSD={0}
        />
      );

      expect(screen.getByText('No distribution data recorded')).toBeInTheDocument();
    });
    it('renders InfoTooltip when tooltip prop is provided', () => {
      render(
        <DistributionCard
          title="Role Breakdown"
          items={{ coder: 10 }}
          totalCostUSD={10}
          tooltip="Detailed explanation of role inference spend"
        />
      );

      expect(screen.getByRole('tooltip')).toHaveTextContent(
        'Detailed explanation of role inference spend'
      );
      expect(
        screen.getByRole('button', { name: 'Information about Role Breakdown' })
      ).toBeInTheDocument();
    });

    it('handles single item with 100% cost allocation cleanly', () => {
      render(
        <DistributionCard
          title="Single Model"
          items={{ 'gemini-3.8-flash': 50.0 }}
          totalCostUSD={50.0}
        />
      );

      expect(screen.getByText('gemini-3.8-flash')).toBeInTheDocument();
      expect(screen.getByText('(100.0%)')).toBeInTheDocument();
    });
  });

  describe('InfoTooltip Component', () => {
    it('renders trigger button with accessible aria attributes and role="tooltip"', () => {
      render(<InfoTooltip content="Helpful information" ariaLabel="Help details" />);
      const button = screen.getByRole('button', { name: 'Help details' });
      expect(button).toBeInTheDocument();
      const tooltip = screen.getByRole('tooltip');
      expect(tooltip).toBeInTheDocument();
      expect(tooltip).toHaveTextContent('Helpful information');
      expect(button).toHaveAttribute('aria-describedby', tooltip.id);
    });

    it('uses default aria-label when not provided', () => {
      render(<InfoTooltip content="Default label info" />);
      expect(screen.getByRole('button', { name: 'Information' })).toBeInTheDocument();
    });

    it('returns null when content is empty string', () => {
      const { container } = render(<InfoTooltip content="" />);
      expect(container.firstChild).toBeNull();
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('toggles visibility class on mouseenter and mouseleave', () => {
      render(<InfoTooltip content="Interactive tooltip" />);
      const button = screen.getByRole('button', { name: 'Information' });
      const tooltip = screen.getByRole('tooltip');

      expect(tooltip).not.toHaveClass('is-visible');

      fireEvent.mouseEnter(button);
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.mouseLeave(button);
      expect(tooltip).not.toHaveClass('is-visible');
    });

    it('toggles visibility class on focus, blur, and dismisses on Escape key', () => {
      render(<InfoTooltip content="Keyboard tooltip" />);
      const button = screen.getByRole('button', { name: 'Information' });
      const tooltip = screen.getByRole('tooltip');

      fireEvent.focus(button);
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.keyDown(button, { key: 'Escape' });
      expect(tooltip).not.toHaveClass('is-visible');

      fireEvent.focus(button);
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.blur(button);
      expect(tooltip).not.toHaveClass('is-visible');
    });

    it('retains visibility when non-Escape keys are pressed while focused', () => {
      render(<InfoTooltip content="Key check" />);
      const button = screen.getByRole('button', { name: 'Information' });
      const tooltip = screen.getByRole('tooltip');

      fireEvent.focus(button);
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.keyDown(button, { key: 'Enter' });
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.keyDown(button, { key: 'Tab' });
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.keyDown(button, { key: 'ArrowDown' });
      expect(tooltip).toHaveClass('is-visible');

      fireEvent.keyDown(button, { key: ' ' });
      expect(tooltip).toHaveClass('is-visible');
    });

    it('applies custom className to container element', () => {
      const { container } = render(
        <InfoTooltip content="Class test" className="header-tip-custom" />
      );
      const wrapper = container.querySelector('.info-tooltip-container');
      expect(wrapper).toHaveClass('info-tooltip-container', 'header-tip-custom');
    });

    it('applies custom placement classes correctly', () => {
      const { rerender } = render(<InfoTooltip content="Top tip" placement="top" />);
      expect(screen.getByRole('tooltip')).toHaveClass('placement-top');

      rerender(<InfoTooltip content="Bottom tip" placement="bottom" />);
      expect(screen.getByRole('tooltip')).toHaveClass('placement-bottom');

      rerender(<InfoTooltip content="Left tip" placement="left" />);
      expect(screen.getByRole('tooltip')).toHaveClass('placement-left');

      rerender(<InfoTooltip content="Right tip" placement="right" />);
      expect(screen.getByRole('tooltip')).toHaveClass('placement-right');
    });

    it('prevents click event propagation to parent elements', () => {
      const parentClick = vi.fn();
      render(
        <div onClick={parentClick}>
          <InfoTooltip content="Nested info" />
        </div>
      );
      const button = screen.getByRole('button', { name: 'Information' });
      fireEvent.click(button);
      expect(parentClick).not.toHaveBeenCalled();
    });

    it('exports InfoIcon alias identically', () => {
      expect(InfoIcon).toBe(InfoTooltip);
    });
  });

  describe('Navbar Component', () => {
    const dummyVersion = {
      version: 'v0.1.0-dev',
      commit: 'unknown',
      build_date: 'unknown',
    };

    it('renders workspace pill when workspaces directory is configured', () => {
      render(
        <Navbar
          activeTab="overview"
          versionInfo={dummyVersion}
          projectName=""
          projectConfig={{
            project_name: 'TestProject',
            prefixes: [],
            workspaces_dir: 'custom/workspaces',
            role_order: [],
            version: 'v1.0.0',
          }}
        />
      );
      expect(screen.getByText('custom/workspaces')).toBeInTheDocument();
      expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
      expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
    });

    it('renders loading state and does not default to workspaces when configuration is missing or empty', () => {
      render(<Navbar activeTab="overview" versionInfo={dummyVersion} projectName="" />);
      expect(screen.getByText('Loading...')).toBeInTheDocument();
      expect(screen.queryByText('workspaces')).not.toBeInTheDocument();
      expect(screen.queryByText('Live Server')).not.toBeInTheDocument();
      expect(screen.queryByText('Mock Mode')).not.toBeInTheDocument();
    });

    it('dispatches onSelectTab when tab is clicked', () => {
      const onSelect = vi.fn();
      render(<Navbar activeTab="overview" onSelectTab={onSelect} versionInfo={dummyVersion} projectName="" />);

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

    it('renders InfoTooltip when tooltip prop is provided in both populated and empty states', () => {
      const { rerender } = render(
        <TrendChart
          title="Spend Velocity"
          data={[{ date: '2026-10-18', val: 10 }]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          tooltip="Spend velocity trajectory explanation"
        />
      );

      expect(
        screen.getByRole('button', { name: 'Information about Spend Velocity' })
      ).toBeInTheDocument();
      expect(screen.getByRole('tooltip')).toHaveTextContent('Spend velocity trajectory explanation');

      // Empty state
      rerender(
        <TrendChart
          title="Spend Velocity"
          data={[]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          tooltip="Spend velocity trajectory explanation"
        />
      );

      expect(
        screen.getByRole('button', { name: 'Information about Spend Velocity' })
      ).toBeInTheDocument();
      expect(screen.getByRole('tooltip')).toHaveTextContent('Spend velocity trajectory explanation');
    });

    it('does not render InfoTooltip when tooltip prop is omitted or empty in populated and empty states', () => {
      const { rerender } = render(
        <TrendChart
          title="Unannotated Velocity"
          data={[{ date: '2026-10-18', val: 10 }]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
        />
      );

      expect(
        screen.queryByRole('button', { name: 'Information about Unannotated Velocity' })
      ).not.toBeInTheDocument();

      // Empty state without tooltip
      rerender(
        <TrendChart
          title="Unannotated Velocity"
          data={[]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
        />
      );
      expect(
        screen.queryByRole('button', { name: 'Information about Unannotated Velocity' })
      ).not.toBeInTheDocument();

      // Populated state with empty string tooltip
      rerender(
        <TrendChart
          title="Unannotated Velocity"
          data={[{ date: '2026-10-18', val: 10 }]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          tooltip=""
        />
      );
      expect(
        screen.queryByRole('button', { name: 'Information about Unannotated Velocity' })
      ).not.toBeInTheDocument();

      // Empty state with empty string tooltip
      rerender(
        <TrendChart
          title="Unannotated Velocity"
          data={[]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          tooltip=""
        />
      );
      expect(
        screen.queryByRole('button', { name: 'Information about Unannotated Velocity' })
      ).not.toBeInTheDocument();
    });

    it('handles hover and keyboard focus states on TrendChart InfoTooltip trigger', () => {
      render(
        <TrendChart
          title="Interactive Velocity"
          data={[{ date: '2026-10-18', val: 10 }]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          tooltip="Detailed explanation of interactive velocity"
        />
      );

      const trigger = screen.getByRole('button', {
        name: 'Information about Interactive Velocity',
      });
      const bubble = screen.getByRole('tooltip');

      expect(bubble).not.toHaveClass('is-visible');

      // Mouse hover
      fireEvent.mouseEnter(trigger);
      expect(bubble).toHaveClass('is-visible');

      fireEvent.mouseLeave(trigger);
      expect(bubble).not.toHaveClass('is-visible');

      // Keyboard focus
      fireEvent.focus(trigger);
      expect(bubble).toHaveClass('is-visible');

      // Escape key dismiss
      fireEvent.keyDown(trigger, { key: 'Escape' });
      expect(bubble).not.toHaveClass('is-visible');
    });

    it('formats x-axis labels using hourly granularity when provided', () => {
      render(
        <TrendChart
          title="Hourly Throughput"
          granularity="hourly"
          data={[
            { bucket_date: '2026-10-19T10:00:00Z', val: 50 },
            { bucket_date: '2026-10-19T11:00:00Z', val: 75 },
          ]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
        />
      );

      expect(screen.getByText('10:00')).toBeInTheDocument();
      expect(screen.getByText('11:00')).toBeInTheDocument();
    });

    it('respects custom xAxisFormatter and renders empty state message on empty data', () => {
      const customFormatter = (v: string) => `TICK:${v.slice(11, 16)}`;
      const { rerender } = render(
        <TrendChart
          title="Custom Axis Formatter"
          xAxisFormatter={customFormatter}
          data={[
            { bucket_date: '2026-10-19T14:00:00Z', val: 100 },
          ]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
        />
      );

      expect(screen.getByText('TICK:14:00')).toBeInTheDocument();

      rerender(
        <TrendChart
          title="Empty Progression"
          data={[]}
          series={[{ key: 'val', label: 'Value', color: '#3b82f6' }]}
          emptyMessage="No progression data available in this time window"
        />
      );

      expect(screen.getByText('No progression data available in this time window')).toBeInTheDocument();
    });
  });

  describe('SaturationBar Component', () => {
    it('renders healthy saturation status and 200k ceiling', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'HEALTHY',
            root_peak_tokens: 50000,
            root_saturation_pct: 25.0,
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
      expect(
        screen.getByRole('button', { name: 'Information about Context Window Saturation' })
      ).toBeInTheDocument();
      expect(screen.getByRole('tooltip')).toHaveTextContent(
        'Context window utilization of the root agent and subagents evaluated against the 200,000 token healthy operating ceiling.'
      );
    });

    it('renders custom tooltip prop on SaturationBar when provided', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'HEALTHY',
            root_peak_tokens: 10000,
            root_saturation_pct: 5.0,
            subagents_peak_tokens: 5000,
            subagents_peak_role: 'none',
            compactions: 0,
            peak_by_role: {},
          }}
          tooltip="Custom saturation analysis description"
        />
      );

      expect(screen.getByRole('tooltip')).toHaveTextContent(
        'Custom saturation analysis description'
      );
    });

    it('renders dynamic subagent peak role and custom roles in peak breakdown', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'HEALTHY',
            root_peak_tokens: 50000,
            root_saturation_pct: 25.0,
            subagents_peak_tokens: 180000,
            subagents_peak_role: 'advisor',
            compactions: 0,
            peak_by_role: {
              orchestrator: 50000,
              advisor: 180000,
              po: 90000,
              researcher: 60000,
            },
          }}
          healthyCeilingTokens={200000}
        />
      );

      expect(screen.getByText('advisor (180k)')).toBeInTheDocument();
      expect(screen.getByText('advisor')).toBeInTheDocument();
      expect(screen.getByText('po')).toBeInTheDocument();
      expect(screen.getByText('researcher')).toBeInTheDocument();
      expect(screen.getByText('180k')).toBeInTheDocument();
      expect(screen.getByText('90k')).toBeInTheDocument();
    });

    it('renders critical saturation badge when status is CRITICAL', () => {
      render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'CRITICAL',
            root_peak_tokens: 210000,
            root_saturation_pct: 105.0,
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

    it('handles zero or negative healthy ceiling tokens with defensive fallback and tooltip rendering', () => {
      const { rerender } = render(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 0,
            status: 'HEALTHY',
            root_peak_tokens: 0,
            root_saturation_pct: 0.0,
            subagents_peak_tokens: 0,
            subagents_peak_role: 'none',
            compactions: 0,
            peak_by_role: {},
          }}
          healthyCeilingTokens={0}
        />
      );

      expect(screen.getByText('None')).toBeInTheDocument();
      expect(screen.getByText('HEALTHY')).toBeInTheDocument();
      expect(screen.getByText('Ceiling: 200k tokens')).toBeInTheDocument();
      expect(screen.getByText('(0.0% of ceiling)')).toBeInTheDocument();
      expect(screen.queryByText(/NaN/)).not.toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: 'Information about Context Window Saturation' })
      ).toBeInTheDocument();

      // Negative ceiling tokens falls back to 200k
      rerender(
        <SaturationBar
          contextHealth={{
            healthy_ceiling_tokens: 200000,
            status: 'ELEVATED',
            root_peak_tokens: 150000,
            root_saturation_pct: 75.0,
            subagents_peak_tokens: 20000,
            subagents_peak_role: 'coder',
            compactions: 1,
            peak_by_role: { orchestrator: 150000 },
          }}
          healthyCeilingTokens={-50000}
          tooltip="Negative ceiling fallback explanation"
        />
      );

      expect(screen.getByText('Ceiling: 200k tokens')).toBeInTheDocument();
      expect(screen.getByText('ELEVATED')).toBeInTheDocument();
      expect(screen.getByRole('tooltip')).toHaveTextContent(
        'Negative ceiling fallback explanation'
      );
    });
  });

  describe('ThreadViewer Component', () => {
    it('renders empty notice and section header tooltip when threads array is empty', () => {
      render(<ThreadViewer threads={[]} />);
      expect(
        screen.getByText('No session execution threads recorded for this workspace.')
      ).toBeInTheDocument();
      expect(screen.getByText('Session Execution Threads')).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: 'Information about Session Execution Threads' })
      ).toBeInTheDocument();
      expect(
        screen.getByText(
          'Hierarchical session execution threads tracking multi-agent turns, context compactions, and model usage per role.'
        )
      ).toBeInTheDocument();
    });

    it('renders thread headers and expands child models on toggle click', () => {
      const dummyThread: ThreadExecutionRecord = {
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

    it('renders InfoTooltip on section header and all 10 column headers of thread models table', () => {
      const dummyThread: ThreadExecutionRecord = {
        session_id: 'ses-root-02',
        user: 'dev',
        compactions: 0,
        roles: [
          {
            role: 'orchestrator',
            models: [
              {
                model: 'claude-3-5-sonnet',
                invocations: 2,
                input_tokens: 2000,
                output_tokens: 400,
                cache_read_tokens: 1000,
                cache_hit_rate_pct: 33.3,
                cost_usd: 1.25,
                peak_context_tokens: 3500,
                context_health: 'HEALTHY',
              },
            ],
          },
        ],
      };

      render(<ThreadViewer threads={[dummyThread]} />);

      // Section header tooltip
      expect(
        screen.getByRole('button', { name: 'Information about Session Execution Threads' })
      ).toBeInTheDocument();
      expect(
        screen.getByText(
          'Hierarchical session execution threads tracking multi-agent turns, context compactions, and model usage per role.'
        )
      ).toBeInTheDocument();

      // All 10 column header tooltips with placement-bottom and exact explanatory content
      const expectedColumns: readonly { readonly label: string; readonly content: string }[] = [
        { label: 'Role', content: 'Agent persona role executing in this thread' },
        { label: 'Model', content: 'Model family and version used in this session' },
        { label: 'Invocations', content: 'Total count of model invocations in this thread' },
        { label: 'Input', content: 'Input prompt tokens sent to the model' },
        { label: 'Output', content: 'Completion output tokens generated by the model' },
        { label: 'Cache Read', content: 'Prompt tokens read directly from prompt cache' },
        { label: 'Cache Hit', content: 'Percentage of prompt context served from cache' },
        { label: 'Cost', content: 'Total inference cost in US dollars for this model' },
        {
          label: 'Peak Context',
          content: 'Maximum single-turn context window size encountered in this thread',
        },
        {
          label: 'Context Health',
          content:
            'Context window saturation status against healthy ceiling (HEALTHY, ELEVATED, BLOATED, CRITICAL)',
        },
      ];

      for (const col of expectedColumns) {
        const btn = screen.getByRole('button', { name: `Information about ${col.label}` });
        expect(btn).toBeInTheDocument();
        const tooltipId = btn.getAttribute('aria-describedby');
        expect(tooltipId).toBeTruthy();
        const bubble = document.getElementById(tooltipId!);
        expect(bubble).toHaveClass('placement-bottom');
        expect(bubble).toHaveTextContent(col.content);
      }
    });

    it('clicking column header InfoTooltip triggers stops propagation and does not toggle thread details', () => {
      const dummyThread: ThreadExecutionRecord = {
        session_id: 'ses-root-03',
        user: 'dev',
        compactions: 0,
        roles: [
          {
            role: 'orchestrator',
            models: [
              {
                model: 'claude-3-5-sonnet',
                invocations: 1,
                input_tokens: 1000,
                output_tokens: 200,
                cache_read_tokens: 500,
                cache_hit_rate_pct: 50.0,
                cost_usd: 0.5,
                peak_context_tokens: 1500,
                context_health: 'HEALTHY',
              },
            ],
          },
        ],
      };

      render(<ThreadViewer threads={[dummyThread]} />);

      expect(screen.getByText('claude-3-5-sonnet')).toBeInTheDocument();

      // Click column header tooltip button
      const roleTooltipBtn = screen.getByRole('button', { name: 'Information about Role' });
      fireEvent.click(roleTooltipBtn);

      // Model table must remain open and expanded
      expect(screen.getByText('claude-3-5-sonnet')).toBeInTheDocument();

      const costTooltipBtn = screen.getByRole('button', { name: 'Information about Cost' });
      fireEvent.click(costTooltipBtn);

      expect(screen.getByText('claude-3-5-sonnet')).toBeInTheDocument();
    });

    it('wraps thread models table in .table-responsive container and keeps section header outside', () => {
      const dummyThread: ThreadExecutionRecord = {
        session_id: 'ses-root-04',
        user: 'dev',
        compactions: 0,
        roles: [
          {
            role: 'orchestrator',
            models: [
              {
                model: 'claude-3-5-sonnet',
                invocations: 1,
                input_tokens: 1000,
                output_tokens: 200,
                cache_read_tokens: 500,
                cache_hit_rate_pct: 50.0,
                cost_usd: 0.5,
                peak_context_tokens: 1500,
                context_health: 'HEALTHY',
              },
            ],
          },
        ],
      };

      const { container } = render(<ThreadViewer threads={[dummyThread]} />);

      const table = container.querySelector('table.thread-models-table');
      expect(table).not.toBeNull();

      const responsiveWrapper = table!.parentElement;
      expect(responsiveWrapper).not.toBeNull();
      expect(responsiveWrapper).toHaveClass('table-responsive');

      // Section header must be outside .table-responsive
      const header = container.querySelector('.thread-viewer-header');
      expect(header).not.toBeNull();
      expect(responsiveWrapper!.contains(header)).toBe(false);

      // Verify all 10 column header tooltips in table have placement-bottom
      const columnBubbles = table!.querySelectorAll('.info-tooltip-bubble');
      expect(columnBubbles.length).toBe(10);
      columnBubbles.forEach((bubble) => {
        expect(bubble).toHaveClass('placement-bottom');
      });
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

    it('sorts options naturally by ticket id', () => {
      render(
        <TicketSelector
          tickets={[
            { ticket_id: 'AL-10', ticket_title: 'Ticket Ten' },
            { ticket_id: 'AL-2', ticket_title: 'Ticket Two' },
            { ticket_id: 'AL-1', ticket_title: 'Ticket One' },
          ]}
          selectedTicketId="AL-1"
          onSelectTicket={vi.fn()}
        />
      );

      const options = screen.getAllByRole('option');
      expect(options[0]).toHaveTextContent('AL-1 - Ticket One');
      expect(options[1]).toHaveTextContent('AL-2 - Ticket Two');
      expect(options[2]).toHaveTextContent('AL-10 - Ticket Ten');
    });

    it('handles adversarial tickets with leading zeroes, mixed case, non-conforming IDs, and health badges', () => {
      render(
        <TicketSelector
          tickets={[
            { ticket_id: 'AL-10', ticket_title: 'Ten' },
            { ticket_id: '123', ticket_title: 'Numeric' },
            { ticket_id: 'AL-01', ticket_title: 'Zero One', context_health_status: 'HEALTHY' },
            { ticket_id: 'al-2', ticket_title: 'Lower Two' },
            { ticket_id: 'GEN-2', ticket_title: 'Generic' },
            { ticket_id: 'AL-1', ticket_title: '' },
          ]}
          selectedTicketId="AL-01"
          onSelectTicket={vi.fn()}
        />
      );

      const ticketSelect = screen.getByLabelText(/workspace autopsy/i);
      const options = within(ticketSelect).getAllByRole('option');
      expect(options).toHaveLength(6);
      expect(options[0]).toHaveTextContent('AL-01 - Zero One [HEALTHY]');
      expect(options[1]).toHaveTextContent('AL-1 - Untitled');
      expect(options[2]).toHaveTextContent('al-2 - Lower Two');
      expect(options[3]).toHaveTextContent('AL-10 - Ten');
      expect(options[4]).toHaveTextContent('GEN-2 - Generic');
      expect(options[5]).toHaveTextContent('123 - Numeric');
    });

    it('renders disabled placeholder when tickets list is empty', () => {
      render(
        <TicketSelector
          tickets={[]}
          selectedTicketId={null}
          onSelectTicket={vi.fn()}
        />
      );

      const ticketSelect = screen.getByLabelText(/workspace autopsy/i);
      const option = within(ticketSelect).getByRole('option');
      expect(option).toBeDisabled();
      expect(option).toHaveTextContent(/no workspaces available/i);
    });

    it('toggles sort mode between Ticket ID and Completion Date', () => {
      render(
        <TicketSelector
          tickets={[
            { ticket_id: 'AL-1', ticket_title: 'One', completed_at: '2026-03-01T00:00:00Z' },
            { ticket_id: 'AL-2', ticket_title: 'Two', completed_at: '2026-03-15T00:00:00Z' },
            { ticket_id: 'AL-3', ticket_title: 'Three', completed_at: null },
          ]}
          selectedTicketId="AL-1"
          onSelectTicket={vi.fn()}
        />
      );

      const ticketSelect = screen.getByLabelText(/workspace autopsy/i);
      let options = within(ticketSelect).getAllByRole('option');
      // Default: natural ticket ID order
      expect(options[0]).toHaveTextContent('AL-1 - One');
      expect(options[1]).toHaveTextContent('AL-2 - Two');
      expect(options[2]).toHaveTextContent('AL-3 - Three');

      // Change sort mode to completion date
      const sortSelect = screen.getByLabelText(/sort/i);
      fireEvent.change(sortSelect, { target: { value: 'completed_at' } });

      options = within(ticketSelect).getAllByRole('option');
      // Descending by completion date, nulls last
      expect(options[0]).toHaveTextContent('AL-2 - Two');
      expect(options[1]).toHaveTextContent('AL-1 - One');
      expect(options[2]).toHaveTextContent('AL-3 - Three');
    });
  });
});
