import React from 'react';
import type { AnomalySeverity } from '../types/telemetry';

export interface AnomalyBadgeProps {
  readonly severity: AnomalySeverity;
  readonly count?: number | undefined;
  readonly label?: string | undefined;
  readonly size?: ('sm' | 'md' | 'lg') | undefined;
  readonly title?: string | undefined;
  readonly onClick?: ((e: React.MouseEvent<HTMLButtonElement>) => void) | undefined;
}

export const AnomalyBadge: React.FC<AnomalyBadgeProps> = ({
  severity,
  count,
  label,
  size = 'md',
  title,
  onClick,
}) => {
  const sevClass = severity === 'CRITICAL' ? 'critical' : 'warning';
  const sizeClass = size;
  const classes = `anomaly-badge ${sevClass} ${sizeClass} ${onClick ? 'clickable' : ''}`.trim();

  const content = (
    <>
      {label ? (
        <span className="anomaly-badge-label">{label}</span>
      ) : (
        <>
          <span className="anomaly-badge-severity">{severity}</span>
          {count !== undefined && (
            <span className="anomaly-badge-count">{count}</span>
          )}
        </>
      )}
    </>
  );

  if (onClick) {
    return (
      <button
        type="button"
        role="button"
        className={classes}
        onClick={onClick}
        title={title ?? 'Click to open Anomaly Inspector'}
        aria-label={label ?? (count !== undefined ? `${severity} ${count}` : severity)}
      >
        {content}
      </button>
    );
  }

  return (
    <span
      role="status"
      className={classes}
      title={title}
      aria-label={label ?? (count !== undefined ? `${severity} ${count}` : severity)}
    >
      {content}
    </span>
  );
};
