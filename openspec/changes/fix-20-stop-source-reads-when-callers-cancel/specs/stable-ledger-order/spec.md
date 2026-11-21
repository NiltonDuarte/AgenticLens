# Stable ledger order

## ADDED Requirements

### Requirement: Stable order of persisted ledger arrays

Raw metrics ledger writers SHALL order sessions by ID regardless of refresh status. When serializing refreshed or newly recorded facts, writers SHALL use stable keys for tools, anomalies and anomaly evidence. Invocations, skills and delegations SHALL use chronological order with the source sequence preserved for equal or unknown timestamps. Untouched historical session records and migration step lists SHALL preserve their recorded sequence.

#### Scenario: A different subset of existing sessions is refreshed
- **WHEN** two publications share the same session IDs and refresh different subsets of them
- **THEN** the session array has the same order in both publications while updated values remain attached to their original sessions

#### Scenario: Backfill or schema migration serializes facts
- **WHEN** source backfill or schema migration writes a raw metrics ledger
- **THEN** the same ordering policy applies to the root and nested arrays, with repeated occurrences and retained facts preserved

#### Scenario: Metadata-only cleanup or formatting
- **WHEN** a supported ledger is formatted or rewritten solely to remove retired metadata
- **THEN** the existing array sequence is preserved

### Requirement: Stable timestamps for equivalent ledger facts

Ledger publication SHALL preserve the existing file and its root `updated_at` and `completed_at` values when the candidate contains equivalent persisted facts.

#### Scenario: An automatic capture pass finds the same facts
- **WHEN** automatic capture produces the same recorded facts, including session-level dates and evidence
- **THEN** the ledger remains byte-for-byte unchanged and capture can acknowledge the pass

#### Scenario: A persisted fact changes
- **WHEN** generation produces a different recorded fact
- **THEN** it publishes the changed ledger with the new publication timestamps
