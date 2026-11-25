## Purpose

Keep reading an unassigned Codex rollout after Codex moves it to `archived_sessions/`, without weakening attribution, and keep a source that cannot be recovered from stalling the capture worker.

## ADDED Requirements

### Requirement: Recover a moved unassigned rollout only on exact proof

When an unassigned registration's rollout path no longer exists, reconciliation SHALL select a replacement under the same registered store only when all of these hold:
- exactly one rollout there carries the registration's session ID in its own `session_meta`
- the rollout's working directory satisfies the existing project proof
- its parent is consistent with the accepted parent

Recovery SHALL NOT assign a ticket, change the binding or adopt a rollout outside the registered store.

#### Scenario: Unassigned child rollout archived
- **WHEN** an unassigned Codex child registration's rollout moves from `sessions/YYYY/MM/DD/` to `archived_sessions/` in the same store
- **THEN** the next unassigned pass reads it from the archive, updates its registered path, keeps it unassigned with the same parent, and its facts stay available for binding

#### Scenario: Proof is missing or ambiguous
- **WHEN** no rollout proves the session, or several copies do, or the candidate's project or parent conflicts with the registration
- **THEN** the row keeps its accepted facts, records the reason, and no other file's facts are adopted

### Requirement: Isolate an unrecoverable unassigned source

An unassigned row whose source cannot be recovered SHALL be recorded as unavailable or pending, with a diagnostic naming the reason. It SHALL NOT make the worker's pass fail, and it SHALL NOT cause retries until the worker's lifetime runs out. It SHALL be retried on the next integration event or an explicit `capture reconcile`.

#### Scenario: Unrecoverable moved rollout
- **WHEN** an unassigned row's rollout is missing and cannot be proven anywhere in its store
- **THEN** the worker completes its pass for the other rows and tickets, exits normally when it is otherwise idle, and capture diagnostics show the row as unavailable
