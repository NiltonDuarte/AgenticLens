import React from 'react';
import type { AppliedRateSnapshot, CostSummary, EquivalentCostSummary } from '../domain/types';
import { hasEstimatedCost } from '../domain/cost';
import { formatAppliedRatesTooltip, formatCostUSD, formatCurrencyUSD } from '../domain/formatters';
import { InfoTooltip } from './InfoTooltip';

export interface EstimatedCostCellProps {
  readonly amount: number | null;
  readonly summary?: CostSummary | undefined;
  readonly equivalentSummary?: EquivalentCostSummary | undefined;
  readonly appliedRates?: readonly AppliedRateSnapshot[] | undefined;
  readonly subjectLabel: string;
}

/** Render baseline reported spend when no estimates exist, or explicit measured vs EST values with rate details. */
export const EstimatedCostCell: React.FC<EstimatedCostCellProps> = ({
  amount,
  summary,
  equivalentSummary,
  appliedRates,
  subjectLabel,
}) => {
  if (!hasEstimatedCost(equivalentSummary) || !equivalentSummary) {
    return <>{formatCostUSD(amount, summary)}</>;
  }

  let primaryText = '';
  let suffixText = '';

  if (equivalentSummary.provenance === 'estimated') {
    primaryText = formatCurrencyUSD(equivalentSummary.estimated_cost_usd);
  } else if (equivalentSummary.provenance === 'hybrid') {
    primaryText = `${formatCurrencyUSD(equivalentSummary.measured_cost_usd)} measured + ${formatCurrencyUSD(equivalentSummary.estimated_cost_usd)}`;
    suffixText = `(${formatCurrencyUSD(equivalentSummary.total_equivalent_cost_usd)} equiv)`;
  } else if (equivalentSummary.measured_cost_usd !== null) {
    primaryText = `${formatCurrencyUSD(equivalentSummary.measured_cost_usd)} measured + ${formatCurrencyUSD(equivalentSummary.estimated_cost_usd)}`;
    suffixText = '(partial)';
  } else {
    primaryText = formatCurrencyUSD(equivalentSummary.estimated_cost_usd);
    suffixText = '(partial)';
  }

  const tooltipAriaLabel =
    appliedRates && appliedRates.length > 0
      ? `Applied API-equivalent rate details for ${subjectLabel}`
      : `API-equivalent estimate details for ${subjectLabel}`;

  return (
    <span className="cost-cell-with-est">
      <span>{primaryText}</span>
      <span className="cost-est-badge" aria-label="API-equivalent estimate">
        EST
      </span>
      {suffixText ? <span className="cost-cell-suffix">{suffixText}</span> : null}
      <InfoTooltip
        content={formatAppliedRatesTooltip(appliedRates, equivalentSummary)}
        placement="left"
        ariaLabel={tooltipAriaLabel}
        portal={true}
      />
    </span>
  );
};
