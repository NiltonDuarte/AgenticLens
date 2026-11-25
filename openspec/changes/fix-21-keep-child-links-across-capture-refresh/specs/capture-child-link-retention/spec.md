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

### Requirement: Align delegation key groups by position

For one parent, the system SHALL group accepted and fresh delegations by key, each in transcript order, and pair accepted position i with fresh position i. A unique key is a group of one. At each position, the unique-delegation rules SHALL apply. Only when the fresh group is smaller SHALL the system refuse the group: it places none of its accepted links, not even at overlapping positions, and reports them.

#### Scenario: Ambiguous delegation provenance, unread group
- **WHEN** the parent has several accepted delegations with the same agent type, description and invocation time (or unknown invocation time), and every fresh link of that group is empty because the children were not read
- **THEN** each accepted link stays on the fresh delegation at the same position, except a contradicted child or one held by another delegation of the parent, and the kept children count as unresolved child evidence

#### Scenario: Partly re-read group
- **WHEN** a fresh read of such a group resolves some positions and leaves others empty
- **THEN** resolved positions follow fresh evidence, and the empty positions keep their accepted links and count as unresolved child evidence

#### Scenario: Ambiguous delegation provenance, smaller fresh group
- **WHEN** the fresh group with a key is smaller than the accepted group, as in a truncated transcript
- **THEN** no accepted link of that group is placed, including at positions both groups share, and the pass reports them as unresolved, refused as ambiguous when the invocation time is known and as unknown provenance when it is not

#### Scenario: Untimed unique delegation against an unconfirmed child
- **WHEN** a parent's only delegation with an unknown invocation time has an accepted link, and the fresh read reports a different child whose parent is not confirmed
- **THEN** the accepted link is kept until a later read confirms the new child

#### Scenario: Group alignment repeated
- **WHEN** the same partial pass repeats after links were kept by position
- **THEN** the group's links are unchanged, and no link is invented

### Requirement: Report accepted keys without a fresh counterpart

When an accepted delegation's key has no fresh delegation, the system SHALL NOT place its link anywhere. It SHALL report the link as refused with an unmatched reason, and SHALL count the child as unresolved child evidence.

#### Scenario: Key changed between reads
- **WHEN** an accepted delegation had an unknown invocation time and the fresh read records a known one, so its key has no fresh counterpart
- **THEN** its accepted link is not placed, it is reported as refused and unmatched, and unresolved child evidence counts that child once

### Requirement: Retained links stay out of transcript events

Request-local transcript events SHALL keep only the child links that the current source resolves. Retained links SHALL appear only on persisted delegations.

#### Scenario: Events after a partial refresh
- **WHEN** a pass keeps an accepted link for a child it did not read
- **THEN** the parent's spawn event for that delegation carries no child session ID, and the persisted delegation keeps the link

### Requirement: Fresh source evidence decides changed links

A refreshed link SHALL change only when the fresh read resolves the delegation to a different child whose native parent metadata names this parent, and which no other fresh delegation of the parent also links. When fresh child metadata names a different parent or none, the link SHALL be cleared, even if it was accepted earlier. Source-level contradiction warnings SHALL behave as they do today; this change adds no new warning surface.

#### Scenario: Fresh child contradicts the accepted link
- **WHEN** a pass reads the parent and a linked child whose own metadata now names another parent
- **THEN** that delegation's link is cleared, the parent's other accepted links follow the retention rule, and source-level warning behavior is unchanged

#### Scenario: Fresh read resolves a different valid child
- **WHEN** the fresh read resolves a delegation to a different child whose metadata names this parent
- **THEN** the fresh child replaces the accepted link

#### Scenario: Fresh child duplicated across delegations
- **WHEN** the fresh read links the same child from more than one delegation of the parent, for example accepted `[A, B, C]` and fresh `[B, B, A]`
- **THEN** the duplicated fresh child does not confirm a replacement, and at each such position the accepted link is kept unless another final delegation holds that child
- **AND THEN** any accepted child left unlinked is reported as displaced: the result is `[B, B, A]`, `A` stays on its confirmed replacement position, and `C` is reported as displaced

#### Scenario: Different child without confirming metadata
- **WHEN** the fresh read reports a different child, this pass did not read that child, and no accepted metadata for it names this parent
- **THEN** the accepted link remains until a later read confirms the new child

### Requirement: Report retained and refused links as unresolved child evidence

A capture pass SHALL count in `pending_children` ("Unresolved child evidence") each accepted link it kept without re-proving, and each uncontradicted accepted link it refused: smaller fresh group, unmatched key, or child displaced by a replacement. A child still linked elsewhere in the owner's final result SHALL NOT be reported as refused. Each child counts once per parent. Parent links SHALL NOT change. A non-zero figure SHALL keep the reconciliation partial.

#### Scenario: Registered children left unread
- **WHEN** a pass refreshes a parent, retains links to registered children it did not read, and one of those children is already queued as pending child work for that parent
- **THEN** unresolved child evidence increases by the number of distinct retained children not already queued, parent links are unchanged, and the reconciliation attempt is partial

#### Scenario: Accepted link refused
- **WHEN** a refreshed parent's accepted link that nothing contradicts cannot be kept, because its fresh group is smaller, its key has no fresh counterpart, or a replacement displaced the child
- **THEN** the link is not moved, unresolved child evidence counts that child once, and the reconciliation attempt is partial

#### Scenario: Child held by another delegation
- **WHEN** an accepted link cannot be kept at its position because another final delegation of the parent already holds that child
- **THEN** the link is not duplicated, and the child is not reported as refused, because it remains linked

#### Scenario: Child displaced by a replacement
- **WHEN** a confirmed replacement takes a delegation whose accepted child is then linked nowhere in the owner's final result, for example accepted `[A, B]` and fresh `["", A]`
- **THEN** the displaced child is reported as refused with reason `displaced` and counted once as unresolved, while a child still linked elsewhere, such as `A`, is not reported

#### Scenario: Children unknown to the local registry
- **WHEN** generation retains links from the accepted ledger for children that this machine's capture registry has never registered
- **THEN** unresolved child evidence includes those children, and the latest recorded reconciliation for the ticket is partial rather than complete

### Requirement: Explicit generation behavior is unchanged

Manual `generate`, explicit backfill and OpenCode capture SHALL resolve delegation child links as they do today, without carrying links forward.

#### Scenario: Full manual Codex generation
- **WHEN** `generate --source codex` reads a parent with task-name-only spawn results together with all of its child rollouts
- **THEN** every delegation is linked to its child from the current source alone
