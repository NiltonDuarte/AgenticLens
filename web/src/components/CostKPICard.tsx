import React from 'react';
import type { CostSummary } from '../domain/types';
import { costSummaryForAmount } from '../domain/cost';
import { formatCurrencyUSD } from '../domain/formatters';
import { KPICard } from './KPICard';

export interface CostKPICardProps {
  readonly title: string;
  readonly amount: number | null;
  readonly summary?: CostSummary | undefined;
  readonly invocations: number;
  readonly secondary?: string | undefined;
  readonly tooltip?: string | undefined;
}

/** Keep currency readable while making cost scope and evidence accessible. */
export const CostKPICard: React.FC<CostKPICardProps> = ({ title, amount, summary, invocations, secondary, tooltip }) => {
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
