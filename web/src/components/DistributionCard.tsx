import type { CostSummary } from '../domain/types';
import { compareCosts } from '../domain/cost';
import React from 'react';
import { formatCostUSD, formatPercentage } from '../domain/formatters';
import { getRoleColor, getModelColor, THEME_COLORS } from '../utils/theme';
import { InfoTooltip } from './InfoTooltip';

export interface DistributionCardProps {
  readonly title: string;
  readonly items: Readonly<Record<string, number | null>>;
  readonly totalCostUSD: number | null;
  readonly costSummaries?: Readonly<Record<string, CostSummary>> | undefined;
  readonly colorType?: 'role' | 'model' | 'default';
  readonly tooltip?: string | undefined;
}

interface SortedItem {
  readonly name: string;
  readonly cost: number | null;
  readonly percentage: number | null;
}

export const DistributionCard: React.FC<DistributionCardProps> = ({
  title,
  items,
  totalCostUSD,
  costSummaries,
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
      const pct = totalCostUSD !== null && totalCostUSD > 0 && cost !== null && Number.isFinite(cost) ? (cost / totalCostUSD) * 100 : null;
      return {
        name,
        cost,
        percentage: pct,
      };
    })
    .sort((a, b) => compareCosts(a.cost, b.cost, false) || a.name.localeCompare(b.name));

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
          {(totalCostUSD === null || totalCostUSD === 0) && <p>Cost shares unavailable: a complete, nonzero total is required.</p>}
          {sorted.map((item) => {
            const itemColor = getItemColor(item.name);
            const clampedPct = item.percentage === null ? 0 : Math.min(Math.max(item.percentage, 0), 100);
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
                    <span>{formatCostUSD(item.cost, costSummaries?.[item.name])}</span>
                    <span>({formatPercentage(item.percentage)})</span>
                  </div>
                </div>
                <div className="progress-track" aria-hidden="true" hidden={item.percentage === null}>
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
