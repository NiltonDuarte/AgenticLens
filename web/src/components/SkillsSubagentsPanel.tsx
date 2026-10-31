import React, { useId, useState } from 'react';
import type { SkillFact, TaskLinkFact } from '../domain/types';
import {
  formatCurrencyUSD,
  formatDurationSeconds,
  formatLocalDateTimeTooltip,
  formatLocalTimeWithMs,
  formatSkillStatus,
  formatSubagentRole,
} from '../domain/formatters';

export interface SkillsSubagentsPanelProps {
  readonly skills: readonly SkillFact[];
  readonly delegations: readonly TaskLinkFact[];
  readonly availableSessionIds: readonly string[];
  readonly onJumpToSession: (sessionId: string) => void;
}

export const SkillsSubagentsPanel: React.FC<SkillsSubagentsPanelProps> = ({
  skills, delegations, availableSessionIds, onJumpToSession,
}) => {
  const [expanded, setExpanded] = useState(true);
  const panelId = useId();

  return (
    <section className="thread-viewer-card skills-subagents-card" aria-labelledby={`${panelId}-heading`}>
      <h4 id={`${panelId}-heading`} className="skills-subagents-heading">
        <button type="button" className="thread-header-btn" aria-expanded={expanded}
          aria-controls={panelId} onClick={() => setExpanded((value) => !value)}>
          <span>Skills &amp; Subagents</span>
          <span className="thread-count-badge">{skills.length} Skills · {delegations.length} Delegations</span>
        </button>
      </h4>
      <div id={panelId} hidden={!expanded}>
        {skills.length === 0 && delegations.length === 0 && (
          <p className="empty-threads-note">No modular skills or subagent delegations recorded for this ticket.</p>
        )}
        {skills.length > 0 && (
          <div className="table-responsive" role="region" aria-label="Recorded skills" tabIndex={0}>
            <table className="data-table skills-subagents-table">
              <caption>Modular skills</caption>
              <thead><tr>
                <th scope="col">Skill Name</th><th scope="col">Loaded At (Turn / Time)</th>
                <th scope="col">Invoking Role</th><th scope="col">Execution Status</th><th scope="col">Error Details</th>
              </tr></thead>
              <tbody>{skills.map((skill, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{skill.name}</th>
                  <td>Turn {skill.turn_index} / <time dateTime={skill.timestamp}
                    title={formatLocalDateTimeTooltip(skill.timestamp)}>{formatLocalTimeWithMs(skill.timestamp)}</time></td>
                  <td>{formatSubagentRole(skill.caller_role)}</td>
                  <td><span className={`skill-status ${skill.success ? 'success' : 'failed'}`}>{formatSkillStatus(skill.success)}</span></td>
                  <td className="source-text">{skill.error ?? 'None recorded'}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
        {delegations.length > 0 && (
          <div className="table-responsive" role="region" aria-label="Recorded subagent delegations" tabIndex={0}>
            <table className="data-table skills-subagents-table">
              <caption>Subagent delegations</caption>
              <thead><tr>
                <th scope="col">Subagent Type</th><th scope="col">Description / Task</th>
                <th scope="col">Parent Turn</th><th scope="col">Duration</th><th scope="col">Cost (USD)</th>
                <th scope="col">Child Session</th>
              </tr></thead>
              <tbody>{delegations.map((task, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{formatSubagentRole(task.subagent_type)}</th>
                  <td className="source-text">{task.description}</td>
                  <td>{task.parent_turn}</td>
                  <td>{task.duration_ms === undefined ? 'Unavailable' : formatDurationSeconds(task.duration_ms / 1000)}</td>
                  <td>{task.cost_usd === undefined ? 'Unavailable' : formatCurrencyUSD(task.cost_usd)}</td>
                  <td className="source-text">{task.child_session_id ? (
                    availableSessionIds.includes(task.child_session_id) ? (
                      <button type="button" className="btn-secondary child-session-jump"
                        onClick={() => { if (task.child_session_id) onJumpToSession(task.child_session_id); }}
                        aria-label={`Jump to child session ${task.child_session_id}`}>
                        {task.child_session_id}
                      </button>
                    ) : <><span>{task.child_session_id}</span><span className="session-unavailable">Thread unavailable</span></>
                  ) : 'Not recorded'}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
};
