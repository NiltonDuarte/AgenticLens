import React from 'react';
import type { FleetAnomalySummary } from '../types/telemetry';

export interface FleetAnomalyCardProps {
  readonly summary?: FleetAnomalySummary | null | undefined;
  readonly onViewTicket?: ((ticketId: string) => void) | undefined;
}

interface RuleMeta {
  readonly id: string;
  readonly label: string;
  readonly description: string;
}

const RULES: readonly RuleMeta[] = [
  {
    id: 'TOOL_RETRY_STREAK',
    label: 'Tool Retry Streak',
    description: 'Consecutive execution failures of the same tool',
  },
  {
    id: 'TOOL_HIGH_FAILURE_RATE',
    label: 'Tool High Failure Rate',
    description: 'Tools exceeding 20% aggregate failure rate',
  },
  {
    id: 'CONTEXT_SATURATION_SPIKE',
    label: 'Context Saturation Spike',
    description: 'Sudden token jumps or critical ceiling breaches',
  },
  {
    id: 'SUBAGENT_REWORK_LOOP',
    label: 'Subagent Rework Loop',
    description: 'Cyclic role thrashing between subagents',
  },
];

export const FleetAnomalyCard: React.FC<FleetAnomalyCardProps> = ({
  summary,
}) => {
  if (summary && (summary.disabled_tickets ?? 0) > 0 && summary.total_tickets === 0) {
    return <div className="fleet-anomaly-card" data-testid="fleet-anomaly-card">
      <h3 className="card-section-title">Recorded anomaly analysis disabled</h3>
      <p>All selected ledgers omitted anomaly analysis during generation.</p>
      <ToolReliabilityTable brittleTools={summary.brittle_tools ?? []} showAnomalies={false} />
    </div>;
  }
  if (summary && (summary.unavailable_tickets ?? 0) > 0 && summary.total_anomalies === 0) {
    return <div className="fleet-anomaly-card" data-testid="fleet-anomaly-card">
      <h3 className="card-section-title">Recorded anomaly analysis unavailable</h3>
      <p>{summary.unavailable_tickets} enabled tickets include sessions whose analysis requires source-backed recovery. No clean-fleet conclusion is available.</p>
      {(summary.disabled_tickets ?? 0) > 0 && <p>{summary.disabled_tickets} tickets omitted anomaly analysis.</p>}
      <ToolReliabilityTable brittleTools={summary.brittle_tools ?? []} showAnomalies={false} />
    </div>;
  }
  if (summary && (summary.disabled_tickets ?? 0) > 0 && summary.total_anomalies === 0) {
    return <div className="fleet-anomaly-card" data-testid="fleet-anomaly-card">
      <h3 className="card-section-title">No recorded anomalies in enabled ledgers</h3>
      <p>{summary.total_tickets} enabled tickets; {summary.disabled_tickets} tickets omitted anomaly analysis.</p>
      <ToolReliabilityTable brittleTools={summary.brittle_tools ?? []} showAnomalies={false} />
    </div>;
  }
  if (!summary || summary.total_anomalies === 0) {
    return (
      <div className="fleet-anomaly-card clean-card" data-testid="fleet-anomaly-card">
        <div className="clean-card-header">
          <h3 className="card-section-title">Fleet Operating Cleanly - Zero Anomalies Detected</h3>
          <p className="clean-card-subtitle">
            All workspace executions operate within normal health thresholds with zero flagged anomalies.
          </p>
        </div>
        {summary && (summary.brittle_tools?.length ?? 0) > 0 && <ToolReliabilityTable brittleTools={summary.brittle_tools} showAnomalies={false} />}
      </div>
    );
  }

  const ruleCounts = summary.rule_counts ?? {};
  const brittleTools = summary.brittle_tools ?? [];

  return (
    <div className="fleet-anomaly-card" data-testid="fleet-anomaly-card">
      <div className="fleet-anomaly-header">
        <div>
          <h3 className="card-section-title">Fleet Anomaly and Reliability Profile</h3>
          <p className="card-section-subtitle">
            Anomaly counts cover enabled workspaces. Tool calls and errors cover all recorded workspaces.
            {(summary.disabled_tickets ?? 0) > 0 && ` ${summary.disabled_tickets} tickets omitted recorded anomaly analysis.`}
            {(summary.unavailable_tickets ?? 0) > 0 && ` ${summary.unavailable_tickets} enabled tickets have unavailable session analysis; shown findings are incomplete.`}
          </p>
        </div>
      </div>

      {/* Primary Anomaly Metrics Row */}
      <div className="fleet-anomaly-kpis">
        <div className="anomaly-kpi-item">
          <span className="anomaly-kpi-label">Total Anomalies</span>
          <span className="anomaly-kpi-val mono">{summary.total_anomalies}</span>
          <span className="anomaly-kpi-sub">
            {summary.warning_count} warning, {summary.critical_count} critical
          </span>
        </div>

        <div className="anomaly-kpi-item">
          <span className="anomaly-kpi-label">Affected Tickets</span>
          <span className="anomaly-kpi-val mono">
            {summary.tickets_with_anomalies} of {summary.total_tickets} tickets
          </span>
          <span className="anomaly-kpi-sub">
            {summary.total_tickets > 0
              ? `${Math.round((summary.tickets_with_anomalies / summary.total_tickets) * 100)}% of analyzed tickets`
              : '0% of fleet'}
          </span>
        </div>

        <div className="anomaly-kpi-item">
          <span className="anomaly-kpi-label">Anomaly Rate</span>
          <span className="anomaly-kpi-val mono">
            {summary.anomaly_rate_per_ticket.toFixed(2)}
          </span>
          <span className="anomaly-kpi-sub">average per ticket</span>
        </div>
      </div>

      {/* Rule Frequency Grid */}
      <div className="rule-frequency-section">
        <h4 className="subsection-title">Anomaly Rule Distribution</h4>
        <div className="rule-frequency-grid">
          {RULES.map((rule) => {
            const count = ruleCounts[rule.id] ?? 0;
            return (
              <div key={rule.id} className="rule-frequency-item">
                <div className="rule-header">
                  <span className="rule-name">{rule.label}</span>
                  <span className={`rule-count mono ${count > 0 ? 'has-count' : ''}`}>
                    {count}
                  </span>
                </div>
                <span className="rule-desc">{rule.description}</span>
              </div>
            );
          })}
        </div>
      </div>

      <ToolReliabilityTable brittleTools={brittleTools} showAnomalies />
    </div>
  );
};

function ToolReliabilityTable({ brittleTools, showAnomalies }: {
  readonly brittleTools: FleetAnomalySummary['brittle_tools']; readonly showAnomalies: boolean;
}): React.ReactNode {
  return (
    <div className="brittle-tools-section">
        <h4 className="subsection-title">{showAnomalies ? 'Brittle Tools' : 'Recorded Tool Reliability'}</h4>
        {brittleTools.length === 0 ? (
          <div className="brittle-tools-empty">
            {showAnomalies ? 'Zero brittle tools identified across the workspace fleet.' : 'No recorded tool errors or flagged tools in the selected fleet.'}
          </div>
        ) : (
          <div className="table-responsive">
            <table className="brittle-tools-table">
              <thead>
                <tr>
                  <th>Tool Name</th>
                  {showAnomalies && <th className="num-col">Anomalies</th>}
                  <th className="num-col">Total Calls</th>
                  <th className="num-col">Errors</th>
                  <th className="num-col">Failure Rate</th>
                </tr>
              </thead>
              <tbody>
                {brittleTools.map((t) => {
                  const failureRate =
                    t.total_calls > 0
                      ? t.failure_rate_pct ?? (t.error_count / t.total_calls) * 100
                      : 0;

                  return (
                    <tr key={t.tool_name}>
                      <td className="tool-name-col mono">{t.tool_name}</td>
                      {showAnomalies && <td className="num-col mono">{t.anomaly_count}</td>}
                      <td className="num-col mono">{t.total_calls}</td>
                      <td className="num-col mono negative">{t.error_count}</td>
                      <td className="num-col mono">{failureRate.toFixed(1)}%</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
  );
}
