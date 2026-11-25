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

### Requirement: Skip local date directories only beyond the maximum UTC offset

Discovery SHALL skip a machine-local `sessions/YYYY`, `sessions/YYYY/MM` or `sessions/YYYY/MM/DD` directory only when the end of that local period, read as UTC, is more than 14 hours plus the 5-second tolerance before the parent's UTC instant. Otherwise it SHALL descend and apply the per-file rule.

#### Scenario: Local date differs from the UTC date
- **WHEN** the machine's local time is ahead of UTC, as in CET or CEST, so that a child created shortly after local midnight sits in a later local-date directory than the parent's UTC date
- **THEN** that child is found, and local year, month and day directories that end beyond that bound are not listed

### Requirement: Exact owner lookup skips other UUIDs

An exact owner lookup SHALL skip, without claiming or reading, every file whose name matches the convention but carries a different UUID. It SHALL read every file named with the requested UUID.

#### Scenario: Duplicate copies of the requested owner
- **WHEN** capture looks up one session ID and the store holds many files named with other UUIDs, plus two copies named with the requested UUID
- **THEN** only the matching-named files are read, and the duplicate copies are still reported as ambiguous

### Requirement: File names only exclude candidates

A file name SHALL NOT establish identity. Discovery SHALL accept a session, or a child of a parent, only on the candidate's own `session_meta`. A file name or a local date directory MAY only exclude candidates, through the child-window, date-directory and exact-lookup rules above.

#### Scenario: Matching name with different metadata
- **WHEN** a conventionally named candidate that is not excluded carries `session_meta` naming a different session or parent than its file name suggests
- **THEN** its metadata decides its identity and parent

#### Scenario: Permitted exclusions
- **WHEN** an exact lookup meets a conventionally named file whose UUID differs from the request, whether or not that UUID is version 7, or a sweep meets any file inside a pruned local date directory
- **THEN** that file is skipped without being read, whatever its metadata would say

### Requirement: Unconventional names are read

In a child sweep, a candidate outside a pruned date directory whose name does not match the convention, or whose UUID is not version 7, SHALL be read as it is today. In an exact lookup, a file whose name does not match the convention SHALL be read.

#### Scenario: Name outside the convention
- **WHEN** a rollout's file name is malformed, or a child-sweep candidate carries a non-v7 UUID
- **THEN** the file is read and its metadata decides whether it belongs to the family

### Requirement: Read each header once and charge actual bytes

Within one reconciliation pass, the system SHALL read each rollout header at most once, using a pass-local cache keyed by the file's physical path, size and modification time. Nothing in the cache SHALL be persisted. A header read SHALL be charged the bytes actually consumed, within the existing 64 KiB per-read bound. A header read SHALL start only when that bound, capped at the file size, fits in the remaining allowance.

#### Scenario: Several consumers inspect the same rollout
- **WHEN** child discovery, owner preflight and exact lookup need the same rollout header in one pass
- **THEN** the file's header is read once, and the pass is charged only for the bytes actually read

#### Scenario: Rollout changes during the pass
- **WHEN** a rollout's size or modification time changes after its header was cached
- **THEN** the header is read again rather than served from the cache

#### Scenario: Allowance too small for a header
- **WHEN** the remaining allowance is smaller than `min(size, 64 KiB)` for the next header
- **THEN** the read does not start, nothing is charged for it, and the work stays pending

### Requirement: Reserve a bounded identification allowance

Identification SHALL mean child discovery: Codex child sweeps, queued-child header reads, the parent header read for the window, and OpenCode child pages. During the rotating work-class rounds, identification SHALL use at most a quarter of `max_source_bytes`, reserved, and owner extraction at most the other three quarters. A known owner's own preflight and exact lookup SHALL be charged to extraction. The final leftover refresh MAY use everything left of the total.

#### Scenario: Extraction scheduled before discovery
- **WHEN** a rotation phase schedules owner extraction before child discovery, and the owners could consume the whole budget
- **THEN** extraction stops at three quarters of the budget, and the reserved quarter remains for identification in that pass

#### Scenario: Owner refused by the reservation
- **WHEN** an owner's read is refused only because it would enter the reserved quarter
- **THEN** the owner is recorded as not attempted, its session slot is returned, and the final leftover refresh retries it

#### Scenario: Unused identification bytes
- **WHEN** identification uses less than its quarter during the rotating rounds
- **THEN** the final leftover refresh can extract owners with the unused bytes, and total charged bytes never exceed `max_source_bytes`

### Requirement: Identification never borrows from extraction

Header identification SHALL stop at its ceiling and SHALL NOT draw on extraction's share. Header reads SHALL NOT consume the owner-extraction session allowance or count as source attempts. Exhausting identification SHALL leave discovery pending, not failed, with the existing budget diagnostics. No configuration key SHALL be added.

#### Scenario: Identification reaches its ceiling
- **WHEN** header identification exhausts its quarter during a pass
- **THEN** discovery is reported as pending rather than failed, known ticket owners are still extracted, and the charged bytes match the fixture's known sizes exactly

#### Scenario: Source attempts in the coverage summary
- **WHEN** a pass reads many headers and a few owners
- **THEN** "Source attempts" counts only the owner reads and never exceeds `max_sessions`, while source bytes include the header bytes

### Requirement: Manual discovery semantics are unchanged

Manual `generate` and explicit backfill family discovery SHALL keep reading the whole store. They SHALL keep their nonce, selection and ambiguity semantics, and SHALL NOT lose any family member.

#### Scenario: Manual generation on a large store
- **WHEN** `generate --source codex` selects a root by nonce or session ID on a store with members whose names fall before the root or outside the convention
- **THEN** it discovers the same family and ambiguities it does today
