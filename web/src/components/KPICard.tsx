import React from 'react';
import { InfoTooltip } from './InfoTooltip';

export interface KPICardProps {
  readonly title: string;
  readonly value: string;
  readonly secondary?: string | undefined;
  readonly tooltip?: string | undefined;
}

export const KPICard: React.FC<KPICardProps> = ({
  title,
  value,
  secondary,
  tooltip,
}) => {
  return (
    <div className="kpi-card">
      <div className="kpi-header">
        <span className="kpi-title">{title}</span>
        {tooltip ? (
          <InfoTooltip content={tooltip} ariaLabel={`Information about ${title}`} />
        ) : null}
      </div>
      <div className="kpi-value mono-number">{value || '-'}</div>
      {secondary ? <div className="kpi-secondary">{secondary}</div> : <div />}
    </div>
  );
};
