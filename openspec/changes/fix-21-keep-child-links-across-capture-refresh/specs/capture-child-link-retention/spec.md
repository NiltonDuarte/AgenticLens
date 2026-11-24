## Purpose

Keep accepted delegation child links when automatic capture refreshes a parent without re-reading all of its children, and keep that partial coverage visible.

## ADDED Requirements

### Requirement: Retain accepted child links through partial capture refresh

When automatic capture refreshes a Codex parent and the fresh read leaves a delegation's child link empty, the system SHALL keep the link accepted for that parent's delegation with the same agent type, description and invocation time, unless the child's native parent metadata contradicts it. An unknown invocation time SHALL be treated as its own key value. The system SHALL NOT create a link that was never accepted.

#### Scenario: Parent refreshed while some children are not read
- **WHEN** an accepted ledger links every `spawn_agent` delegation of a parent whose results contain only a task name, and a capture pass reads the parent and only some of its child rollouts within its byte or session budget
- **THEN** every accepted `child_session_id` remains on its delegation, both in the published ledger and in the capture registry snapshot

#### Scenario: Accepted ledger copied to a machine without the earlier capture state
- **WHEN** automatic capture refreshes a committed ledger that has all links, on a machine whose capture registry has no earlier snapshot of the children, and the pass reads the parent but not every child
- **THEN** the published ledger keeps every accepted link, and other owners and other-harness facts are unchanged

#### Scenario: Repeated partial pass
- **WHEN** the same partial pass is repeated with unchanged sources
- **THEN** the ledger facts and capture snapshot are identical to the previous pass, and the ledger bytes and timestamps are kept

### Requirement: Retain ambiguous groups by ordinal position

When several of a parent's delegations share one key, the system SHALL compare the accepted group and the fresh group with that key, each in transcript order. If every fresh link in the group is empty and the fresh group is at least as large, the system SHALL carry accepted link i to fresh delegation i, subject to the per-child guards. Otherwise it SHALL carry none of the group and SHALL report the refused links.

#### Scenario: Ambiguous delegation provenance, ordinal retention
- **WHEN** the parent has several accepted delegations with the same agent type, description and invocation time (or unknown invocation time), the fresh read has at least as many with that key, and all of their links are empty because their children were not read
- **THEN** each accepted link is carried to the fresh delegation at the same position, except a contradicted child or one held by another delegation of the parent, and the carried children count as unresolved child evidence

#### Scenario: Ambiguous delegation provenance, refusal
- **WHEN** a fresh delegation in such a group already has a link, or the fresh group is smaller than the accepted group
- **THEN** no accepted link is moved onto those delegations, and the pass reports them as unresolved

#### Scenario: Ordinal retention repeated
- **WHEN** the same partial pass repeats after an ordinal carry
- **THEN** the group's links are unchanged, and no link is invented

### Requirement: Retained links stay out of transcript events

Request-local transcript events SHALL keep only the child links that the current source resolves. Retained links SHALL appear only on persisted delegations.

#### Scenario: Events after a partial refresh
- **WHEN** a pass keeps an accepted link for a child it did not read
- **THEN** the parent's spawn event for that delegation carries no child session ID, and the persisted delegation keeps the link

### Requirement: Fresh source evidence decides changed links

A refreshed link SHALL change only when the fresh read resolves the delegation to a different child whose native parent metadata names this parent. When fresh child metadata names a different parent or none, the link SHALL be cleared, even if it was accepted earlier. Source-level contradiction warnings SHALL behave as they do today; this change adds no new warning surface.

#### Scenario: Fresh child contradicts the accepted link
- **WHEN** a pass reads the parent and a linked child whose own metadata now names another parent
- **THEN** that delegation's link is cleared, the parent's other accepted links follow the retention rule, and source-level warning behavior is unchanged

#### Scenario: Fresh read resolves a different valid child
- **WHEN** the fresh read resolves a delegation to a different child whose metadata names this parent
- **THEN** the fresh child replaces the accepted link

#### Scenario: Different child without confirming metadata
- **WHEN** the fresh read reports a different child, this pass did not read that child, and no accepted metadata for it names this parent
- **THEN** the accepted link remains until a later read confirms the new child

### Requirement: Report retained and refused links as unresolved child evidence

A capture pass SHALL count in `pending_children` ("Unresolved child evidence") each accepted link it retained without re-proving, and each accepted link it refused to carry onto an empty, uncontradicted fresh delegation. A child already counted for the same parent SHALL be counted once. Neither kind SHALL change parent links. A non-zero figure SHALL keep the reconciliation partial.

#### Scenario: Registered children left unread
- **WHEN** a pass refreshes a parent, retains links to registered children it did not read, and one of those children is already queued as pending child work for that parent
- **THEN** unresolved child evidence increases by the number of distinct retained children not already queued, parent links are unchanged, and the reconciliation attempt is partial

#### Scenario: Accepted link refused for ambiguity
- **WHEN** a refreshed parent's accepted link cannot be carried, because its group fails the ordinal condition or another delegation of the parent already holds the child, and the fresh link is empty with no contradiction
- **THEN** the link is not moved, unresolved child evidence counts that child once, and the reconciliation attempt is partial

#### Scenario: Children unknown to the local registry
- **WHEN** generation retains links from the accepted ledger for children that this machine's capture registry has never registered
- **THEN** unresolved child evidence includes those children, and the latest recorded reconciliation for the ticket is partial rather than complete

### Requirement: Explicit generation behavior is unchanged

Manual `generate`, explicit backfill and OpenCode capture SHALL resolve delegation child links as they do today, without carrying links forward.

#### Scenario: Full manual Codex generation
- **WHEN** `generate --source codex` reads a parent with task-name-only spawn results together with all of its child rollouts
- **THEN** every delegation is linked to its child from the current source alone
