import React from 'react';
import { formatCurrencyUSD, formatPercentage } from '../domain/formatters';

export interface DistributionCardProps {
  readonly title: string;
  readonly items: Readonly<Record<string, number>>;
  readonly totalCostUSD: number;
}

interface SortedItem {
  readonly name: string;
  readonly cost: number;
  readonly percentage: number;
}

export const DistributionCard: React.FC<DistributionCardProps> = ({
  title,
  items,
  totalCostUSD,
}) => {
  const sorted: SortedItem[] = Object.entries(items)
    .map(([name, cost]) => {
      const pct = totalCostUSD > 0 ? (cost / totalCostUSD) * 100 : 0;
      return {
        name,
        cost,
        percentage: pct,
      };
    })
    .sort((a, b) => b.cost - a.cost);

  return (
    <div className="distribution-card">
      <h3 className="card-title">{title}</h3>
      {sorted.length === 0 ? (
        <div className="empty-state" style={{ padding: '24px' }}>
          <p className="empty-state-message">No distribution data recorded</p>
        </div>
      ) : (
        <div className="distribution-list">
          {sorted.map((item) => (
            <div key={item.name} className="distribution-item">
              <div className="distribution-label-row">
                <span className="distribution-name">{item.name}</span>
                <div className="distribution-values mono-number">
                  <span>{formatCurrencyUSD(item.cost)}</span>
                  <span>({formatPercentage(item.percentage)})</span>
                </div>
              </div>
              <div className="progress-track" aria-hidden="true">
                <div
                  className="progress-fill"
                  style={{ width: `${Math.min(Math.max(item.percentage, 0), 100)}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
