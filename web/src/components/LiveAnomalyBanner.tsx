import React, { useState } from 'react';
import type { LiveAnomalyAlert, LiveAnomalySeverity } from '../domain/live';

export interface LiveAnomalyBannerProps {
  readonly anomalies: readonly LiveAnomalyAlert[];
  readonly onSelectSession: (sessionId: string) => void;
}

function resolveHighestSeverity(
  anomalies: readonly LiveAnomalyAlert[]
): LiveAnomalySeverity {
  if (anomalies.some((a) => a.severity === 'CRITICAL')) {
    return 'CRITICAL';
  }
  if (anomalies.some((a) => a.severity === 'HIGH')) {
    return 'HIGH';
  }
  return 'WARNING';
}

export const LiveAnomalyBanner: React.FC<LiveAnomalyBannerProps> = ({
  anomalies,
  onSelectSession,
}) => {
  const [expanded, setExpanded] = useState(false);

  if (!anomalies || anomalies.length === 0) {
    return null;
  }

  const highestSeverity = resolveHighestSeverity(anomalies);
  const severityClass = highestSeverity.toLowerCase();
  const visibleAnomalies = expanded ? anomalies : anomalies.slice(0, 2);
  const hasMore = anomalies.length > 2;

  return (
    <aside
      className={`live-anomaly-banner severity-${severityClass}`}
      role="alert"
      aria-label="Active Live Anomalies"
    >
      <div className="live-anomaly-banner-header">
        <div className="live-anomaly-banner-title">
          <span className={`live-anomaly-badge severity-${severityClass}`}>
            {highestSeverity}
          </span>
          <span className="live-anomaly-count-text">
            {anomalies.length} Active {anomalies.length === 1 ? 'Anomaly' : 'Anomalies'} Detected
          </span>
        </div>
        {hasMore && (
          <button
            type="button"
            className="live-anomaly-toggle-btn"
            onClick={() => setExpanded((prev) => !prev)}
          >
            {expanded ? 'Show Less' : `+${anomalies.length - 2} More`}
          </button>
        )}
      </div>

      <div className="live-anomaly-list">
        {visibleAnomalies.map((alert, idx) => (
          <div
            key={`${alert.session_id}-${alert.rule_name}-${idx}`}
            className={`live-anomaly-item severity-${alert.severity.toLowerCase()}`}
            onClick={() => onSelectSession(alert.session_id)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                onSelectSession(alert.session_id);
              }
            }}
          >
            <div className="live-anomaly-meta">
              <span className={`live-anomaly-rule-tag severity-${alert.severity.toLowerCase()}`}>
                {alert.rule_name}
              </span>
              <span className="live-anomaly-session-tag">
                Session: {alert.session_id}
              </span>
              <span className="live-anomaly-timestamp">
                {alert.timestamp}
              </span>
            </div>
            <div className="live-anomaly-message">{alert.message}</div>
          </div>
        ))}
      </div>
    </aside>
  );
};
