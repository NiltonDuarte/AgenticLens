import { useRolePresentation } from '../contexts/ProjectConfigContext';
import React from 'react';
import type { ContextWindowHealth } from '../domain/types';
import { formatRoleLabel, formatTokensK } from '../domain/formatters';
import { getHealthStatusColor, getRoleColor, THEME_COLORS } from '../utils/theme';
import { InfoTooltip } from './InfoTooltip';

export interface SaturationBarProps {
  readonly contextHealth: ContextWindowHealth;
  readonly healthyCeilingTokens?: number;
  readonly tooltip?: string | undefined;
}

export const SaturationBar: React.FC<SaturationBarProps> = ({
  contextHealth,
  healthyCeilingTokens,
  tooltip,
}) => {
  const rolePresentation = useRolePresentation();
  const recordedCeiling = healthyCeilingTokens ?? contextHealth.healthy_ceiling_tokens;
  const ceiling = Number.isFinite(recordedCeiling) && recordedCeiling > 0 ? recordedCeiling : null;
  const rootTokens = contextHealth.root_peak_tokens;
  const satPct = contextHealth.root_saturation_pct;
  const status = contextHealth.status;

  const maxScaleTokens = Math.max(ceiling ?? 0, rootTokens, contextHealth.subagents_peak_tokens);
  const fillPct = maxScaleTokens > 0 ? Math.min(100, (rootTokens / maxScaleTokens) * 100) : 0;
  const ceilingMarkerPct = ceiling !== null && maxScaleTokens > 0 ? (ceiling / maxScaleTokens) * 100 : null;

  const statusColor = getHealthStatusColor(status);

  const rolePeaks = contextHealth.peak_by_role ?? {};
  const roleEntries = Object.entries(rolePeaks).sort((a, b) => rolePresentation.compareProjected(a[0], b[0]));

  return (
    <div className="saturation-card">
      <div className="saturation-header">
        <div className="saturation-title-group">
          <h4 className="saturation-title">Context Window Saturation</h4>
          <InfoTooltip
            content={
              tooltip ??
              (status === 'UNAVAILABLE' ? 'Recorded context usage. Complete anomaly analysis is not recorded for this ledger.' : (ceiling === null ? 'Recorded context usage; a context ceiling is unavailable.' : `Context usage evaluated against the recorded ${ceiling.toLocaleString()} token ceiling.`))
            }
            ariaLabel="Information about Context Window Saturation"
          />
          <span className="ceiling-label">Ceiling: {ceiling === null ? 'Not recorded' : `${formatTokensK(ceiling)} tokens`}</span>
        </div>
        <div className="saturation-status-badge" style={{ borderColor: statusColor, color: statusColor }}>
          <span className="status-dot" style={{ backgroundColor: statusColor }} />
          {status === 'UNAVAILABLE' ? 'Not recorded' : status}
        </div>
      </div>

      <div className="saturation-metric-row">
        <div className="saturation-metric">
          <span className="metric-label">Root Agent Peak</span>
          <span className="metric-val">{formatTokensK(rootTokens)} tokens</span>
          <span className="metric-sub">{ceiling === null ? 'Saturation not recorded' : `(${satPct.toFixed(1)}% of ceiling)`}</span>
        </div>
        <div className="saturation-metric">
          <span className="metric-label">Subagents Peak Role</span>
          <span className="metric-val">
            {contextHealth.subagents_peak_role !== 'none'
              ? `${formatRoleLabel(contextHealth.subagents_peak_role)} (${formatTokensK(contextHealth.subagents_peak_tokens)})`
              : 'None'}
          </span>
          <span className="metric-sub">Compactions: {contextHealth.compactions}</span>
        </div>
      </div>

      {/* Progress Bar with Ceiling Tick */}
      <div className="saturation-bar-track">
        <div
          className="saturation-bar-fill"
          style={{
            width: `${Math.max(2, fillPct)}%`,
            backgroundColor: statusColor,
          }}
        />
        {ceilingMarkerPct !== null && ceilingMarkerPct < 100 && ceiling !== null && (
          <div
            className="saturation-ceiling-marker"
            style={{ left: `${ceilingMarkerPct}%` }}
            title={`Healthy Ceiling: ${ceiling.toLocaleString()} tokens`}
          >
            <div className="marker-line" />
            <span className="marker-tag">100% Ceiling</span>
          </div>
        )}
      </div>

      {/* Role Peak Context Breakdown */}
      {roleEntries.length > 0 && (
        <div className="saturation-roles-list">
          <h5 className="roles-breakdown-title">Peak Context by Agent Role</h5>
          <div className="roles-grid">
            {roleEntries.map(([role, tokens]) => {
              const rolePct = ceiling !== null ? (tokens / ceiling) * 100 : 0;
              const roleColor = getRoleColor(role);
              return (
                <div key={role} className="role-token-item">
                  <div className="role-token-header">
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <span
                        className="role-color-dot"
                        style={{
                          width: '8px',
                          height: '8px',
                          borderRadius: '50%',
                          backgroundColor: roleColor,
                          display: 'inline-block',
                          flexShrink: 0,
                        }}
                        aria-hidden="true"
                      />
                      <span className="role-name">{formatRoleLabel(role)}</span>
                    </div>
                    <span className="role-tokens">{formatTokensK(tokens)}</span>
                  </div>
                  <div className="role-mini-track">
                    <div
                      className="role-mini-fill"
                      style={{
                        width: `${Math.min(100, rolePct)}%`,
                        backgroundColor: ceiling !== null && tokens > ceiling ? THEME_COLORS.red : roleColor,
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
