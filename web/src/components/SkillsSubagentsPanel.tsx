import { sessionMetricCoverage } from '../domain/types';
import React, { useId, useState } from 'react';
import type { OwnedSkillFact, OwnedDelegationFact, SessionFact } from '../domain/types';
import {
  formatHistoricalDateTimeTooltip, getHistoricalTimestampDateTime,
  formatLocalDateTime, formatSubagentRole,
} from '../domain/formatters';

function InvocationTime({ value }: { readonly value: string | null }): React.ReactNode {
  return value === null ? 'Unavailable' : <time dateTime={getHistoricalTimestampDateTime(value)}
    title={formatHistoricalDateTimeTooltip(value)}>{formatLocalDateTime(value)}</time>;
}

export interface SkillsSubagentsPanelProps {
  readonly skills: readonly OwnedSkillFact[];
  readonly sessions?: readonly SessionFact[];
  readonly delegations: readonly OwnedDelegationFact[];
  readonly availableSessionIds: readonly string[];
  readonly onJumpToSession: (sessionId: string) => void;
}

export const SkillsSubagentsPanel: React.FC<SkillsSubagentsPanelProps> = ({
  skills, delegations, availableSessionIds, onJumpToSession, sessions,
}) => {
  const [expanded, setExpanded] = useState(true);
  const panelId = useId();
  const [sessionId, setSessionId] = useState('');
  const sessionIds = [...new Set([...availableSessionIds, ...skills.flatMap((fact) => fact.session_id ? [fact.session_id] : []),
    ...delegations.flatMap((fact) => fact.session_id ? [fact.session_id] : [])])];
  const visibleSkills = skills.filter((fact) => !sessionId || fact.session_id === sessionId);
  const visibleDelegations = delegations.filter((fact) => !sessionId || fact.session_id === sessionId);
  const selectedOwners = sessions?.filter((session) => !sessionId || session.id === sessionId);
  const skillsCoverage = selectedOwners ? sessionMetricCoverage(selectedOwners, 'skills') : 'recorded';
  const delegationsCoverage = selectedOwners ? sessionMetricCoverage(selectedOwners, 'delegations') : 'recorded';

  return (
    <section className="thread-viewer-card skills-subagents-card" aria-labelledby={`${panelId}-heading`}>
      <h4 id={`${panelId}-heading`} className="skills-subagents-heading">
        <button type="button" className="thread-header-btn" aria-expanded={expanded}
          aria-controls={panelId} onClick={() => setExpanded((value) => !value)}>
          <span>Skills &amp; Subagents</span>
          <span className="thread-count-badge">{skillsCoverage === 'recorded' ? `${visibleSkills.length} Skills` : visibleSkills.length > 0 ? `${visibleSkills.length} recorded Skills` : 'Skills not recorded'} · {delegationsCoverage === 'recorded' ? `${visibleDelegations.length} Delegations` : visibleDelegations.length > 0 ? `${visibleDelegations.length} recorded Delegations` : 'Delegations not recorded'}</span>
        </button>
      </h4>
      <div id={panelId} hidden={!expanded}>
        {skillsCoverage !== 'recorded' && <p>Skills not recorded for {skillsCoverage === 'partial' ? 'some selected sessions; recorded skills remain visible' : 'the selected sessions'}.</p>}
        {delegationsCoverage !== 'recorded' && <p>Delegations not recorded for {delegationsCoverage === 'partial' ? 'some selected sessions; recorded delegations remain visible' : 'the selected sessions'}.</p>}
        {sessionIds.length > 0 && (skills.length > 0 || delegations.length > 0) && <label>Session <select aria-label="Filter skills and delegations by session"
          value={sessionId} onChange={(event) => setSessionId(event.target.value)}>
          <option value="">All sessions</option>
          {sessionIds.map((id) => <option key={id} value={id}>{id}</option>)}
        </select></label>}
        {sessionId && skillsCoverage === 'recorded' && delegationsCoverage === 'recorded' && visibleSkills.length === 0 && visibleDelegations.length === 0 &&
          <p>No recorded skills or delegations for this session.</p>}
        {skillsCoverage === 'recorded' && delegationsCoverage === 'recorded' && skills.length === 0 && delegations.length === 0 && (
          <p className="empty-threads-note">No modular skills or subagent delegations recorded for this ticket.</p>
        )}
        {visibleSkills.length > 0 && (
          <div className="table-responsive" role="region" aria-label="Recorded skills" tabIndex={0}>
            <table className="data-table skills-subagents-table">
              <caption>Modular skills</caption>
              <thead><tr>
                <th scope="col">Skill Name</th><th scope="col">Invoked At</th>
                <th scope="col">Owning Session</th><th scope="col">Invoking Role</th>
              </tr></thead>
              <tbody>{visibleSkills.map((skill, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{skill.name}</th>
                  <td><InvocationTime value={skill.invoked_at} /></td>
                  <td className="source-text">{skill.session_id}</td>
                  <td>{skill.caller_role === null ? 'Unavailable' : formatSubagentRole(skill.caller_role)}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
        {visibleDelegations.length > 0 && (
          <div className="table-responsive" role="region" aria-label="Recorded subagent delegations" tabIndex={0}>
            <table className="data-table skills-subagents-table">
              <caption>Subagent delegations</caption>
              <thead><tr>
                <th scope="col">Subagent Type</th><th scope="col">Description / Task</th>
                <th scope="col">Invoked At</th>
                <th scope="col">Child Session</th><th scope="col">Owning Session</th>
              </tr></thead>
              <tbody>{visibleDelegations.map((task, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{task.agent_type === null ? 'Unavailable' : formatSubagentRole(task.agent_type)}</th>
                  <td className="source-text">{task.description ?? 'Unavailable'}</td>
                  <td><InvocationTime value={task.invoked_at} /></td>
                  <td className="source-text">{task.child_session_id ? (
                    availableSessionIds.includes(task.child_session_id) ? (
                      <button type="button" className="btn-secondary child-session-jump"
                        onClick={() => { if (task.child_session_id) onJumpToSession(task.child_session_id); }}
                        aria-label={`Jump to child session ${task.child_session_id}`}>
                        {task.child_session_id}
                      </button>
                    ) : <><span>{task.child_session_id}</span><span className="session-unavailable">Thread unavailable</span></>
                  ) : 'Session identity unavailable'}</td>
                  <td className="source-text">{task.session_id}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
};
