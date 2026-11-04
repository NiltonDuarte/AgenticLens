import React, { useId, useState } from 'react';
import { skillOutcome, type FactProvenance, type SkillFact, type TaskLinkFact } from '../domain/types';
import {
  formatFactCostUSD,
  formatDurationSeconds,
  formatHistoricalDateTimeTooltip, getHistoricalTimestampDateTime,
  formatLocalDateTime,
  formatSubagentRole,
} from '../domain/formatters';

function OccurrenceOwner({ fact }: { readonly fact: { readonly session_id?: string; readonly source?: string; readonly provenance?: FactProvenance; readonly evidence_references?: readonly FactProvenance[] } }): React.ReactNode {
  return <td className="source-text">
    <span>{fact.session_id ?? 'Owner unavailable'}</span>
    {fact.source && <div>Source: {fact.source}</div>}
    {[...(fact.provenance ? [fact.provenance] : []), ...(fact.evidence_references ?? [])].map((reference, index) =>
      <div key={index}>{index > 0 && <span>Additional evidence reference</span>}
        {Object.entries(reference).map(([kind, identity]) =>
          <div key={kind}>{kind === 'source_ref' ? 'Diagnostic source reference' : kind.replace('_id', '')}: {identity}</div>)}
      </div>)}
  </td>;
}

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
  const [sessionId, setSessionId] = useState('');
  const sessionIds = [...new Set([...availableSessionIds, ...skills.flatMap((fact) => fact.session_id ? [fact.session_id] : []),
    ...delegations.flatMap((fact) => fact.session_id ? [fact.session_id] : [])])];
  const visibleSkills = skills.filter((fact) => !sessionId || fact.session_id === sessionId);
  const visibleDelegations = delegations.filter((fact) => !sessionId || fact.session_id === sessionId);

  return (
    <section className="thread-viewer-card skills-subagents-card" aria-labelledby={`${panelId}-heading`}>
      <h4 id={`${panelId}-heading`} className="skills-subagents-heading">
        <button type="button" className="thread-header-btn" aria-expanded={expanded}
          aria-controls={panelId} onClick={() => setExpanded((value) => !value)}>
          <span>Skills &amp; Subagents</span>
          <span className="thread-count-badge">{visibleSkills.length} Skills · {visibleDelegations.length} Delegations</span>
        </button>
      </h4>
      <div id={panelId} hidden={!expanded}>
        {sessionIds.length > 0 && (skills.length > 0 || delegations.length > 0) && <label>Session <select aria-label="Filter skills and delegations by session"
          value={sessionId} onChange={(event) => setSessionId(event.target.value)}>
          <option value="">All sessions</option>
          {sessionIds.map((id) => <option key={id} value={id}>{id}</option>)}
        </select></label>}
        {sessionId && visibleSkills.length === 0 && visibleDelegations.length === 0 &&
          <p>No recorded skills or delegations for this session.</p>}
        {skills.length === 0 && delegations.length === 0 && (
          <p className="empty-threads-note">No modular skills or subagent delegations recorded for this ticket.</p>
        )}
        {visibleSkills.length > 0 && (
          <div className="table-responsive" role="region" aria-label="Recorded skills" tabIndex={0}>
            <table className="data-table skills-subagents-table">
              <caption>Modular skills</caption>
              <thead><tr>
                <th scope="col">Skill Name</th><th scope="col">Recorded At (Turn / Time)</th>
                <th scope="col">Owning Session / Source Identity</th><th scope="col">Invoking Role</th><th scope="col">Execution Status</th><th scope="col">Error Details</th>
              </tr></thead>
              <tbody>{visibleSkills.map((skill, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{skill.name}</th>
                  <td>Turn {skill.turn_index} / <time dateTime={getHistoricalTimestampDateTime(skill.timestamp)}
                    title={formatHistoricalDateTimeTooltip(skill.timestamp)}>{formatLocalDateTime(skill.timestamp)}</time></td>
                  <OccurrenceOwner fact={skill} />
                  <td>{formatSubagentRole(skill.caller_role)}</td>
                  <td><span className={`skill-status ${skillOutcome(skill) === 'succeeded' ? 'success' : skillOutcome(skill) === 'failed' ? 'failed' : ''}`}>
                    {skillOutcome(skill) === 'unknown' ? 'Outcome unavailable' : skillOutcome(skill) === 'succeeded' ? 'Success' : 'Failed'}
                  </span>{skill.evidence_kind === 'activation_request' && <div>Activation requested</div>}</td>
                  <td className="source-text">{skill.error ?? 'None recorded'}</td>
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
                <th scope="col">Parent Turn</th><th scope="col">Duration</th><th scope="col">Cost (USD)</th>
                <th scope="col">Child Session</th><th scope="col">Owning Session / Source Identity</th>
              </tr></thead>
              <tbody>{visibleDelegations.map((task, index) => (
                <tr key={index} tabIndex={0}>
                  <th scope="row">{formatSubagentRole(task.subagent_type)}</th>
                  <td className="source-text">{task.description}</td>
                  <td>{task.parent_turn}</td>
                  <td>{task.duration_ms === undefined ? 'Unavailable' : formatDurationSeconds(task.duration_ms / 1000)}</td>
                  <td>{formatFactCostUSD(task.cost_usd, task.cost_evidence)}</td>
                  <td className="source-text">{task.child_session_id ? (
                    availableSessionIds.includes(task.child_session_id) && !task.child_link_diagnostic ? (
                      <button type="button" className="btn-secondary child-session-jump"
                        onClick={() => { if (task.child_session_id) onJumpToSession(task.child_session_id); }}
                        aria-label={`Jump to child session ${task.child_session_id}`}>
                        {task.child_session_id}
                      </button>
                    ) : <><span>{task.child_session_id}</span><span className="session-unavailable">{task.child_link_diagnostic ? 'Child relationship unresolved' : 'Thread unavailable'}</span></>
                  ) : task.child_reference ? <><span>{task.child_reference}</span><span className="session-unavailable">{task.child_link_diagnostic ? 'Child relationship unresolved' : 'Thread unavailable'}</span></> : task.child_agent_id ? <span className="session-unavailable">Session identity unavailable</span> : 'Not recorded'}
                    {task.child_agent_id && <div>Agent reference: {task.child_agent_id}</div>}
                    {task.child_session_id && task.child_reference && <div>Source reference: {task.child_reference}</div>}
                    {task.child_link_diagnostic && <div role="note">Child relationship: {task.child_link_diagnostic}</div>}
                  </td>
                  <OccurrenceOwner fact={task} />
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  );
};
