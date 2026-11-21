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
