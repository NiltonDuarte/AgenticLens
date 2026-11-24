## Purpose

Let automatic capture identify and extract a ticket's Codex family within its configured budget, even on a large source store, without spending the budget on unrelated rollouts.

## ADDED Requirements

### Requirement: Prune child candidates by UUIDv7 creation instant

In a capture child sweep, discovery SHALL skip a candidate without claiming or reading it when the candidate's file name matches `rollout-YYYY-MM-DDThh-mm-ss-<uuid>.jsonl` (compressed or not), its UUID is version 7, and the UTC creation instant in that UUID's first 48 bits is more than 5 seconds earlier than the parent's instant. The parent's instant SHALL come from its own UUIDv7, or else from its `session_meta` timestamp. When neither is available, that sweep SHALL NOT prune.

#### Scenario: Large store with one active family
- **WHEN** a capture pass runs on a synthetic store with thousands of earlier live and archived rollouts, plus one root and its children created after it
- **THEN** the root and all its children are extracted within the configured budget, and candidates outside the window consume no bytes and no session slots

#### Scenario: Child created just after its parent
- **WHEN** a child's UUIDv7 instant is later than its parent's, or earlier by less than the 5-second tolerance
- **THEN** the child is read and found

#### Scenario: Parent instant unavailable
- **WHEN** the parent has neither a UUIDv7 identity nor a readable `session_meta` timestamp
- **THEN** the sweep reads candidates as it does today

### Requirement: Skip local day directories only beyond the maximum UTC offset

Discovery SHALL skip a `sessions/YYYY/MM/DD` directory, whose date is machine-local, only when the end of that local day, read as UTC, is more than 14 hours before the parent's UTC instant. Otherwise it SHALL descend and apply the per-file rule.

#### Scenario: Local date differs from the UTC date
- **WHEN** the machine's local time is ahead of UTC, as in CET or CEST, so that a child created shortly after local midnight sits in a later local-date directory than the parent's UTC date
- **THEN** that child is found, and local day directories that end more than 14 hours before the parent's instant are not listed

### Requirement: Exact owner lookup skips other UUIDs

An exact owner lookup SHALL skip, without claiming or reading, every file whose name matches the convention but carries a different UUID. It SHALL read every file named with the requested UUID.

#### Scenario: Duplicate copies of the requested owner
- **WHEN** capture looks up one session ID and the store holds many files named with other UUIDs, plus two copies named with the requested UUID
- **THEN** only the matching-named files are read, and the duplicate copies are still reported as ambiguous

### Requirement: File names never establish identity

The file name SHALL serve only as a pre-filter. Session identity and parent SHALL come from the candidate's `session_meta`. A file whose name does not match the convention, or whose UUID is not version 7, SHALL be read as it is today and never skipped.

#### Scenario: Name outside the convention
- **WHEN** a rollout's file name is malformed or carries a non-v7 UUID, or matches but its metadata names a different session
- **THEN** the file is read and its metadata decides its identity, and no member of the family is skipped

### Requirement: Read each header once and charge actual bytes

Within one reconciliation pass, the system SHALL read each rollout header at most once, using a pass-local cache keyed by the file's physical path, size and modification time. Nothing in the cache SHALL be persisted. A header read SHALL be charged the bytes actually consumed, within the existing per-read safety bound.

#### Scenario: Several consumers inspect the same rollout
- **WHEN** child discovery, owner preflight and exact lookup need the same rollout header in one pass
- **THEN** the file's header is read once, and the pass is charged only for the bytes actually read

#### Scenario: Rollout changes during the pass
- **WHEN** a rollout's size or modification time changes after its header was cached
- **THEN** the header is read again rather than served from the cache

### Requirement: Reserve a bounded identification allowance

During the rotating work-class rounds of a pass, header identification SHALL use at most a quarter of `max_source_bytes`. That quarter is reserved, so owner extraction in those rounds SHALL use at most the other three quarters. In the final leftover refresh, extraction MAY use everything left of the total, including identification's unused share.

#### Scenario: Extraction scheduled before discovery
- **WHEN** a rotation phase schedules owner extraction before child discovery, and the owners could consume the whole budget
- **THEN** extraction stops at three quarters of the budget, and the reserved quarter remains for identification in that pass

#### Scenario: Unused identification bytes
- **WHEN** identification uses less than its quarter during the rotating rounds
- **THEN** the final leftover refresh can extract owners with the unused bytes, and total charged bytes never exceed `max_source_bytes`

### Requirement: Identification never borrows from extraction

Header identification SHALL stop at its ceiling and SHALL NOT draw on extraction's share. Header reads SHALL NOT consume the owner-extraction session allowance. Exhausting identification SHALL leave discovery pending, not failed, with the existing budget diagnostics. No configuration key SHALL be added.

#### Scenario: Identification reaches its ceiling
- **WHEN** header identification exhausts its quarter during a pass
- **THEN** discovery is reported as pending rather than failed, known ticket owners are still extracted, and the charged bytes match the fixture's known sizes exactly

### Requirement: Manual discovery semantics are unchanged

Manual `generate` and explicit backfill family discovery SHALL keep reading the whole store. They SHALL keep their nonce, selection and ambiguity semantics, and SHALL NOT lose any family member.

#### Scenario: Manual generation on a large store
- **WHEN** `generate --source codex` selects a root by nonce or session ID on a store with members whose names fall before the root or outside the convention
- **THEN** it discovers the same family and ambiguities it does today
