import React from 'react';
import { formatCurrencyUSD, formatPercentage } from '../domain/formatters';
import { getRoleColor, getModelColor, THEME_COLORS } from '../utils/theme';
import { InfoTooltip } from './InfoTooltip';

export interface DistributionCardProps {
  readonly title: string;
  readonly items: Readonly<Record<string, number>>;
  readonly totalCostUSD: number;
  readonly colorType?: 'role' | 'model' | 'default';
  readonly tooltip?: string | undefined;
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
  colorType,
  tooltip,
}) => {
  const resolvedColorType = colorType ?? (
    title.toLowerCase().includes('role')
      ? 'role'
      : title.toLowerCase().includes('model')
      ? 'model'
      : 'default'
  );

  const getItemColor = (name: string): string => {
    if (resolvedColorType === 'role') {
      return getRoleColor(name);
    }
    if (resolvedColorType === 'model') {
      return getModelColor(name);
    }
    return THEME_COLORS.blue;
  };

  const sorted: SortedItem[] = Object.entries(items)
    .map(([name, cost]) => {
      const pct = totalCostUSD > 0 && !isNaN(cost) ? (cost / totalCostUSD) * 100 : 0;
      return {
        name,
        cost,
        percentage: isNaN(pct) ? 0 : pct,
      };
    })
    .sort((a, b) => b.cost - a.cost);

  return (
    <div className="distribution-card">
      <div className="distribution-header">
        <h3 className="card-title">{title}</h3>
        {tooltip ? (
          <InfoTooltip content={tooltip} ariaLabel={`Information about ${title}`} />
        ) : null}
      </div>
      {sorted.length === 0 ? (
        <div className="empty-state" style={{ padding: '24px' }}>
          <p className="empty-state-message">No distribution data recorded</p>
        </div>
      ) : (
        <div className="distribution-list">
          {sorted.map((item) => {
            const itemColor = getItemColor(item.name);
            const clampedPct = Math.min(Math.max(item.percentage, 0), 100);
            return (
              <div key={item.name} className="distribution-item">
                <div className="distribution-label-row">
                  <div className="distribution-name-group" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span
                      className="distribution-color-dot"
                      style={{
                        width: '8px',
                        height: '8px',
                        borderRadius: '50%',
                        backgroundColor: itemColor,
                        display: 'inline-block',
                        flexShrink: 0,
                      }}
                      aria-hidden="true"
                    />
                    <span className="distribution-name">{item.name}</span>
                  </div>
                  <div className="distribution-values mono-number">
                    <span>{formatCurrencyUSD(item.cost)}</span>
                    <span>({formatPercentage(item.percentage)})</span>
                  </div>
                </div>
                <div className="progress-track" aria-hidden="true">
                  <div
                    className="progress-fill"
                    style={{
                      width: `${clampedPct}%`,
                      backgroundColor: itemColor,
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
