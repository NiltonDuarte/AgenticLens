import React from 'react';

export interface KPICardProps {
  readonly title: string;
  readonly value: string;
  readonly secondary?: string;
  readonly tooltip?: string;
}

export const KPICard: React.FC<KPICardProps> = ({
  title,
  value,
  secondary,
  tooltip,
}) => {
  return (
    <div className="kpi-card" title={tooltip}>
      <div className="kpi-header">
        <span className="kpi-title">{title}</span>
      </div>
      <div className="kpi-value mono-number">{value || '-'}</div>
      {secondary ? <div className="kpi-secondary">{secondary}</div> : <div />}
    </div>
  );
};
