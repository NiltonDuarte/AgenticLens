import React from 'react';
import type { ContextWindowHealth } from '../domain/types';
import { formatTokensK, HEALTHY_CEILING_TOKENS } from '../domain/formatters';
import { getHealthStatusColor, getRoleColor, THEME_COLORS } from '../utils/theme';
import { InfoTooltip } from './InfoTooltip';

export interface SaturationBarProps {
  readonly contextHealth: ContextWindowHealth;
  readonly healthyCeilingTokens?: number;
  readonly tooltip?: string | undefined;
}

export const SaturationBar: React.FC<SaturationBarProps> = ({
  contextHealth,
  healthyCeilingTokens = HEALTHY_CEILING_TOKENS,
  tooltip,
}) => {
  const ceiling = healthyCeilingTokens > 0 ? healthyCeilingTokens : HEALTHY_CEILING_TOKENS;
  const rootTokens = contextHealth.root_peak_tokens;
  const satPct = contextHealth.root_saturation_pct;
  const status = contextHealth.status;

  const maxScaleTokens = Math.max(ceiling, rootTokens, contextHealth.subagents_peak_tokens);
  const fillPct = maxScaleTokens > 0 ? Math.min(100, (rootTokens / maxScaleTokens) * 100) : 0;
  const ceilingMarkerPct = maxScaleTokens > 0 ? (ceiling / maxScaleTokens) * 100 : 100;

  const statusColor = getHealthStatusColor(status);

  const rolePeaks = contextHealth.peak_by_role ?? {};
  const roleEntries = Object.entries(rolePeaks).sort((a, b) => b[1] - a[1]);

  return (
    <div className="saturation-card">
      <div className="saturation-header">
        <div className="saturation-title-group">
          <h4 className="saturation-title">Context Window Saturation</h4>
          <InfoTooltip
            content={
              tooltip ??
              'Context window utilization of the root agent and subagents evaluated against the 200,000 token healthy operating ceiling.'
            }
            ariaLabel="Information about Context Window Saturation"
          />
          <span className="ceiling-label">Ceiling: {formatTokensK(ceiling)} tokens</span>
        </div>
        <div className="saturation-status-badge" style={{ borderColor: statusColor, color: statusColor }}>
          <span className="status-dot" style={{ backgroundColor: statusColor }} />
          {status}
        </div>
      </div>

      <div className="saturation-metric-row">
        <div className="saturation-metric">
          <span className="metric-label">Root Agent Peak</span>
          <span className="metric-val">{formatTokensK(rootTokens)} tokens</span>
          <span className="metric-sub">({satPct.toFixed(1)}% of ceiling)</span>
        </div>
        <div className="saturation-metric">
          <span className="metric-label">Subagents Peak Role</span>
          <span className="metric-val">
            {contextHealth.subagents_peak_role !== 'none'
              ? `${contextHealth.subagents_peak_role} (${formatTokensK(contextHealth.subagents_peak_tokens)})`
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
        {ceilingMarkerPct < 100 && (
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
              const rolePct = ceiling > 0 ? (tokens / ceiling) * 100 : 0;
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
                      <span className="role-name">{role}</span>
                    </div>
                    <span className="role-tokens">{formatTokensK(tokens)}</span>
                  </div>
                  <div className="role-mini-track">
                    <div
                      className="role-mini-fill"
                      style={{
                        width: `${Math.min(100, rolePct)}%`,
                        backgroundColor: tokens > ceiling ? THEME_COLORS.red : roleColor,
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
