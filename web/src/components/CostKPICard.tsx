import React from 'react';
import type { AppliedRateSnapshot, CostSummary, EquivalentCostSummary } from '../domain/types';
import { costSummaryForAmount, hasEstimatedCost } from '../domain/cost';
import { formatAppliedRatesTooltip, formatCurrencyUSD } from '../domain/formatters';
import { KPICard } from './KPICard';

export interface CostKPICardProps {
  readonly title: string;
  readonly amount: number | null;
  readonly summary?: CostSummary | undefined;
  readonly equivalentSummary?: EquivalentCostSummary | undefined;
  readonly appliedRates?: readonly AppliedRateSnapshot[] | undefined;
  readonly invocations: number;
  readonly secondary?: string | undefined;
  readonly tooltip?: string | undefined;
}

/** Keep currency readable while making cost scope and evidence accessible. */
export const CostKPICard: React.FC<CostKPICardProps> = ({
  title,
  amount,
  summary,
  equivalentSummary,
  appliedRates,
  invocations,
  secondary,
  tooltip,
}) => {
  if (hasEstimatedCost(equivalentSummary) && equivalentSummary) {
    let value = '-';
    let status = '';
    let explanation = '';

    if (equivalentSummary.provenance === 'estimated') {
      value = formatCurrencyUSD(equivalentSummary.estimated_cost_usd);
      status = 'API-equivalent estimate (EST)';
      explanation = 'API-equivalent rate-card estimate (not billed spend).';
    } else if (equivalentSummary.provenance === 'hybrid') {
      value = formatCurrencyUSD(equivalentSummary.total_equivalent_cost_usd);
      status = `${formatCurrencyUSD(equivalentSummary.measured_cost_usd)} measured + ${formatCurrencyUSD(equivalentSummary.estimated_cost_usd)} EST (equiv)`;
      explanation = 'Hybrid equivalent total combining source-reported measured spend and configured API-equivalent rate-card estimates (not billed invoice spend).';
    } else {
      value = equivalentSummary.known_subtotal_usd !== null
        ? formatCurrencyUSD(equivalentSummary.known_subtotal_usd)
        : '-';
      status = equivalentSummary.measured_cost_usd !== null && equivalentSummary.estimated_cost_usd !== null
        ? `Partial (${formatCurrencyUSD(equivalentSummary.measured_cost_usd)} measured + ${formatCurrencyUSD(equivalentSummary.estimated_cost_usd)} EST)`
        : `Partial (${formatCurrencyUSD(equivalentSummary.estimated_cost_usd)} EST)`;
      explanation = 'Known partial subtotal only; unpriced invocations remain so complete equivalent cost is unavailable.';
    }

    const rateDetails = formatAppliedRatesTooltip(appliedRates, equivalentSummary);
    const tooltipText = [
      tooltip,
      rateDetails.startsWith(explanation) ? '' : explanation,
      rateDetails,
    ]
      .filter(Boolean)
      .join(' ');

    return (
      <KPICard
        title={title}
        value={value}
        valueBadge={
          <span className="cost-est-badge" aria-label="API-equivalent estimate">
            EST
          </span>
        }
        secondary={[status, secondary].filter(Boolean).join(' · ')}
        tooltip={tooltipText}
      />
    );
  }

  const cost = costSummaryForAmount(amount, summary, invocations);
  const value = cost.state === 'complete'
    ? formatCurrencyUSD(amount)
    : cost.state === 'partial' && cost.known_subtotal_usd !== null
      ? formatCurrencyUSD(cost.known_subtotal_usd)
      : '-';
  const status = cost.state === 'partial'
    ? 'Known subtotal (partial)'
    : cost.state === 'unavailable'
      ? 'Unavailable'
      : cost.state === 'empty' ? 'No invocation data' : '';
  const explanation = cost.state === 'complete'
    ? 'Complete reported cost.'
    : cost.state === 'partial'
      ? 'Known subtotal only; the complete cost is unavailable.'
      : cost.state === 'empty' ? 'No invocation data establishes a cost.' : 'Reported cost is unavailable.';
  const coverage = `Coverage: ${cost.fully_observed.toLocaleString()} fully observed, ${cost.partial.toLocaleString()} partial, ${cost.unavailable.toLocaleString()} unavailable of ${cost.total.toLocaleString()} invocations. Invocation coverage is not financial coverage.`;

  return <KPICard title={title} value={value}
    secondary={[status, secondary].filter(Boolean).join(' · ')}
    tooltip={[tooltip, explanation, coverage].filter(Boolean).join(' ')} />;
};
