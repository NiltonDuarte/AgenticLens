## ADDED Requirements

### Requirement: Accept only usable live payloads

The dashboard SHALL validate live trees, trace identity, messages, parts and anomalies, including nested collections and required scalars, consistently at REST and SSE ingress. It SHALL preserve supported optional omissions and source diagnostics. See the [AL-71 proposal](../../proposal.md).

#### Scenario: A malformed observation follows a valid observation

- **WHEN** REST or SSE supplies a malformed body or a malformed required element inside a live tree, trace or alert
- **THEN** the malformed payload cannot replace the last valid observation or appear as a successful empty result, and the dashboard surfaces a diagnostic

#### Scenario: A measured empty or partial trace arrives

- **WHEN** a valid trace reports no recent errors through the current Go producer's null representation, omits supported optional metadata or reports unavailable child relationships with usable messages
- **THEN** the dashboard renders the available facts and retains the reported completeness and diagnostics

#### Scenario: A trace belongs to another session

- **WHEN** a REST trace reports an identity different from the requested session, or an SSE trace belongs to another session
- **THEN** it cannot replace the selected trace

### Requirement: Bound fallback polling and report recovery

The dashboard SHALL keep at most one in-flight fallback poll per active subscription at its configured cadence. A current valid poll SHALL restore fallback-polling status after an error and publish its snapshot. The dashboard SHALL preserve ordering across polling, SSE and manual reads.

#### Scenario: Poll ticks arrive during a pending read

- **WHEN** a poll remains unresolved through multiple timer ticks
- **THEN** no concurrent poll starts and polling resumes when the pending read has settled

#### Scenario: Polling recovers after the error threshold

- **WHEN** polling reaches its existing failure threshold and a later current poll returns a valid snapshot
- **THEN** the status returns to fallback polling, the recovered snapshot is visible and its superseded error clears

#### Scenario: A read becomes obsolete

- **WHEN** a poll settles after the subscription is replaced or closed, or after a newer tree observation was accepted
- **THEN** its result or failure cannot update the snapshot or status
