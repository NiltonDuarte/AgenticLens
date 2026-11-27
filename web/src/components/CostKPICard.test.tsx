import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '../test/render';
import { CostKPICard } from './CostKPICard';
import { EstimatedCostCell } from './EstimatedCostCell';
import type { AppliedRateSnapshot, EquivalentCostSummary } from '../domain/types';

const sampleRate: AppliedRateSnapshot = {
  matched_pattern: 'gpt-5.4',
  input_per_million: 2.0,
  cache_read_per_million: 0.25,
  output_per_million: 8.0,
  effective_catalog_id: 'sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
};

const secondRate: AppliedRateSnapshot = {
  matched_pattern: '*gemini-3.8*',
  input_per_million: 0.5,
  cache_read_per_million: 0.05,
  output_per_million: 1.5,
  effective_catalog_id: 'sha256:fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321',
};

describe('CostKPICard equivalent-cost presentation', () => {
  it('renders pure measured spend without EST label', () => {
    const measured: EquivalentCostSummary = {
      provenance: 'measured',
      measured_cost_usd: 4.5,
      estimated_cost_usd: null,
      total_equivalent_cost_usd: 4.5,
      known_subtotal_usd: 4.5,
      measured: 2,
      estimated: 0,
      partial_measured: 0,
      unavailable: 0,
      total: 2,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={4.5}
        summary={{ state: 'complete', known_subtotal_usd: 4.5, fully_observed: 2, partial: 0, unavailable: 0, total: 2 }}
        equivalentSummary={measured}
        invocations={2}
        secondary="2 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('$4.50');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent(/^2 invocations$/);
    expect(container.querySelector('.cost-est-badge')).not.toBeInTheDocument();
  });

  it('renders pure estimated spend explicitly with EST badge, appliedRates tooltip, and not billed spend', () => {
    const estimated: EquivalentCostSummary = {
      provenance: 'estimated',
      measured_cost_usd: null,
      estimated_cost_usd: 6.74,
      total_equivalent_cost_usd: 6.74,
      known_subtotal_usd: 6.74,
      measured: 0,
      estimated: 3,
      partial_measured: 0,
      unavailable: 0,
      total: 3,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={null}
        summary={{ state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 3, total: 3 }}
        equivalentSummary={estimated}
        appliedRates={[sampleRate]}
        invocations={3}
        secondary="3 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('$6.74');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent('API-equivalent estimate (EST) · 3 invocations');
    const tooltip = screen.getByRole('tooltip');
    expect(tooltip).toHaveTextContent('not billed spend');
    expect(tooltip).toHaveTextContent('0 measured, 3 estimated, 0 partial, 0 unavailable of 3 invocations');
    expect(tooltip).toHaveTextContent('gpt-5.4: $2.00 in / $0.25 cache read / $8.00 out per 1M tokens');
    expect(tooltip).toHaveTextContent(sampleRate.effective_catalog_id);
  });

  it('renders hybrid spend with EST badge, measured + EST breakdown, appliedRates tooltip, and equivalent total', () => {
    const hybrid: EquivalentCostSummary = {
      provenance: 'hybrid',
      measured_cost_usd: 23.2,
      estimated_cost_usd: 7.12,
      total_equivalent_cost_usd: 30.32,
      known_subtotal_usd: 30.32,
      measured: 2,
      estimated: 2,
      partial_measured: 0,
      unavailable: 0,
      total: 4,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={null}
        summary={{ state: 'partial', known_subtotal_usd: 23.2, fully_observed: 2, partial: 0, unavailable: 2, total: 4 }}
        equivalentSummary={hybrid}
        appliedRates={[sampleRate, secondRate]}
        invocations={4}
        secondary="4 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('$30.32');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent('$23.20 measured + $7.12 EST (equiv) · 4 invocations');
    const tooltip = screen.getByRole('tooltip');
    expect(tooltip).toHaveTextContent('2 measured, 2 estimated, 0 partial, 0 unavailable of 4 invocations');
    expect(tooltip).toHaveTextContent('gpt-5.4: $2.00 in / $0.25 cache read / $8.00 out per 1M tokens');
    expect(tooltip).toHaveTextContent('*gemini-3.8*: $0.50 in / $0.05 cache read / $1.50 out per 1M tokens');
  });

  it('renders partial mixed spend honestly as partial subtotal without claiming a complete equivalent total', () => {
    const partial: EquivalentCostSummary = {
      provenance: 'partial',
      measured_cost_usd: 2.0,
      estimated_cost_usd: 3.0,
      total_equivalent_cost_usd: null,
      known_subtotal_usd: 5.0,
      measured: 1,
      estimated: 1,
      partial_measured: 0,
      unavailable: 1,
      total: 3,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={null}
        summary={{ state: 'partial', known_subtotal_usd: 2.0, fully_observed: 1, partial: 0, unavailable: 2, total: 3 }}
        equivalentSummary={partial}
        invocations={3}
        secondary="3 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('$5.00');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent('Partial ($2.00 measured + $3.00 EST) · 3 invocations');
    expect(container.querySelector('.kpi-secondary')).not.toHaveTextContent('(equiv)');
    expect(screen.getByRole('tooltip')).toHaveTextContent('complete equivalent cost is unavailable');
  });

  it('renders estimated-only partial spend honestly with EST badge and no measured subtotal', () => {
    const partialEstOnly: EquivalentCostSummary = {
      provenance: 'partial',
      measured_cost_usd: null,
      estimated_cost_usd: 3.0,
      total_equivalent_cost_usd: null,
      known_subtotal_usd: 3.0,
      measured: 0,
      estimated: 1,
      partial_measured: 0,
      unavailable: 1,
      total: 2,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={null}
        summary={{ state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 2, total: 2 }}
        equivalentSummary={partialEstOnly}
        appliedRates={[sampleRate]}
        invocations={2}
        secondary="2 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('$3.00');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent('Partial ($3.00 EST) · 2 invocations');
    expect(container.querySelector('.kpi-secondary')).not.toHaveTextContent('measured');
    expect(screen.getByRole('tooltip')).toHaveTextContent('complete equivalent cost is unavailable');
  });

  it('renders unavailable spend without EST badge', () => {
    const unavailable: EquivalentCostSummary = {
      provenance: 'unavailable',
      measured_cost_usd: null,
      estimated_cost_usd: null,
      total_equivalent_cost_usd: null,
      known_subtotal_usd: null,
      measured: 0,
      estimated: 0,
      partial_measured: 0,
      unavailable: 2,
      total: 2,
    };
    const { container } = render(
      <CostKPICard
        title="Total Cost"
        amount={null}
        summary={{ state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 2, total: 2 }}
        equivalentSummary={unavailable}
        invocations={2}
        secondary="2 invocations"
      />
    );
    expect(container.querySelector('.kpi-value')).toHaveTextContent('-');
    expect(container.querySelector('.kpi-secondary')).toHaveTextContent('Unavailable · 2 invocations');
    expect(container.querySelector('.cost-est-badge')).not.toBeInTheDocument();
  });
});

describe('EstimatedCostCell presentation', () => {
  it('omits EST badge for pure measured rows and renders badge + portaled accessible rate tooltip for estimated rows', () => {
    const { container, rerender } = render(
      <table>
        <tbody>
          <tr>
            <td data-testid="cost-cell">
              <EstimatedCostCell
                amount={4.5}
                summary={{ state: 'complete', known_subtotal_usd: 4.5, fully_observed: 1, partial: 0, unavailable: 0, total: 1 }}
                equivalentSummary={{
                  provenance: 'measured',
                  measured_cost_usd: 4.5,
                  estimated_cost_usd: null,
                  total_equivalent_cost_usd: 4.5,
                  known_subtotal_usd: 4.5,
                  measured: 1,
                  estimated: 0,
                  partial_measured: 0,
                  unavailable: 0,
                  total: 1,
                }}
                subjectLabel="claude-sonnet-4"
              />
            </td>
          </tr>
        </tbody>
      </table>
    );
    expect(container).toHaveTextContent('$4.50');
    expect(container.querySelector('.cost-est-badge')).not.toBeInTheDocument();

    rerender(
      <table>
        <tbody>
          <tr>
            <td data-testid="cost-cell">
              <EstimatedCostCell
                amount={null}
                summary={{ state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 1, total: 1 }}
                equivalentSummary={{
                  provenance: 'estimated',
                  measured_cost_usd: null,
                  estimated_cost_usd: 1.0,
                  total_equivalent_cost_usd: 1.0,
                  known_subtotal_usd: 1.0,
                  measured: 0,
                  estimated: 1,
                  partial_measured: 0,
                  unavailable: 0,
                  total: 1,
                }}
                appliedRates={[sampleRate]}
                subjectLabel="gpt-5.4"
              />
            </td>
          </tr>
        </tbody>
      </table>
    );
    const cell = screen.getByTestId('cost-cell');
    expect(cell.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    const trigger = screen.getByRole('button', { name: 'Applied API-equivalent rate details for gpt-5.4' });
    fireEvent.focus(trigger);
    const tooltip = document.getElementById(trigger.getAttribute('aria-describedby')!);
    expect(tooltip).not.toBeNull();
    // Tooltip MUST be portaled into document.body (outside the table cell) with .is-portaled
    expect(tooltip!.parentElement).toBe(document.body);
    expect(cell.contains(tooltip)).toBe(false);
    expect(tooltip).toHaveClass('is-portaled');
    expect(tooltip).toHaveClass('is-visible');
    expect(tooltip).toHaveTextContent('gpt-5.4: $2.00 in / $0.25 cache read / $8.00 out per 1M tokens');
    expect(tooltip).toHaveTextContent(sampleRate.effective_catalog_id);
    expect(tooltip).toHaveTextContent('CacheWrite');
  });

  it('renders hybrid, partial-with-measured, partial-estimated-only, multi-rate tooltip, and honest role ariaLabel without appliedRates', () => {
    const { container, rerender } = render(
      <EstimatedCostCell
        amount={null}
        summary={{ state: 'partial', known_subtotal_usd: 23.2, fully_observed: 1, partial: 0, unavailable: 1, total: 2 }}
        equivalentSummary={{
          provenance: 'hybrid',
          measured_cost_usd: 23.2,
          estimated_cost_usd: 7.12,
          total_equivalent_cost_usd: 30.32,
          known_subtotal_usd: 30.32,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 0,
          total: 2,
        }}
        appliedRates={[sampleRate, secondRate]}
        subjectLabel="gpt-5.4"
      />
    );
    expect(container).toHaveTextContent('$23.20 measured + $7.12');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.cost-cell-suffix')).toHaveTextContent('($30.32 equiv)');
    const multiTrigger = screen.getByRole('button', { name: 'Applied API-equivalent rate details for gpt-5.4' });
    fireEvent.focus(multiTrigger);
    const multiTooltip = document.getElementById(multiTrigger.getAttribute('aria-describedby')!);
    expect(multiTooltip).toHaveTextContent('gpt-5.4: $2.00 in / $0.25 cache read / $8.00 out per 1M tokens');
    expect(multiTooltip).toHaveTextContent('*gemini-3.8*: $0.50 in / $0.05 cache read / $1.50 out per 1M tokens');
    expect(multiTooltip).toHaveTextContent(sampleRate.effective_catalog_id);
    expect(multiTooltip).toHaveTextContent(secondRate.effective_catalog_id);

    // Partial with measured + estimated
    rerender(
      <EstimatedCostCell
        amount={null}
        summary={{ state: 'partial', known_subtotal_usd: 2.0, fully_observed: 1, partial: 0, unavailable: 2, total: 3 }}
        equivalentSummary={{
          provenance: 'partial',
          measured_cost_usd: 2.0,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: null,
          known_subtotal_usd: 5.0,
          measured: 1,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 3,
        }}
        subjectLabel="coder"
      />
    );
    expect(container).toHaveTextContent('$2.00 measured + $3.00');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.cost-cell-suffix')).toHaveTextContent('(partial)');
    // Honest ariaLabel when appliedRates is absent (e.g. Role row)
    const roleTrigger = screen.getByRole('button', { name: 'API-equivalent estimate details for coder' });
    fireEvent.focus(roleTrigger);
    const roleTooltip = document.getElementById(roleTrigger.getAttribute('aria-describedby')!);
    expect(roleTooltip).toHaveTextContent('API-equivalent rate-card estimate (not billed spend).');
    expect(roleTooltip).toHaveClass('is-portaled');

    // Partial estimated-only (no measured spend)
    rerender(
      <EstimatedCostCell
        amount={null}
        summary={{ state: 'unavailable', known_subtotal_usd: null, fully_observed: 0, partial: 0, unavailable: 2, total: 2 }}
        equivalentSummary={{
          provenance: 'partial',
          measured_cost_usd: null,
          estimated_cost_usd: 3.0,
          total_equivalent_cost_usd: null,
          known_subtotal_usd: 3.0,
          measured: 0,
          estimated: 1,
          partial_measured: 0,
          unavailable: 1,
          total: 2,
        }}
        subjectLabel="coder"
      />
    );
    expect(container).toHaveTextContent('$3.00');
    expect(container).not.toHaveTextContent('measured');
    expect(container.querySelector('.cost-est-badge')).toHaveTextContent('EST');
    expect(container.querySelector('.cost-cell-suffix')).toHaveTextContent('(partial)');
  });
});
