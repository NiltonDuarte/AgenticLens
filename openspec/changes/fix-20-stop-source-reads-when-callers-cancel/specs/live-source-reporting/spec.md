## ADDED Requirements

### Requirement: Stop cancelled live source reads

The live service SHALL stop native work when the requesting context is cancelled or its deadline expires. An interrupted read SHALL report cancellation to its caller and preserve accepted cache entries that still match the source revision.

#### Scenario: Cancellation before live work starts
- **WHEN** a caller requests the session tree or a trace with an already-cancelled context, or cancels while waiting for the trace cache
- **THEN** its request exits with a cancellation error without starting native queries

#### Scenario: Cancellation during native work
- **WHEN** a live query is interrupted during session listing, schema inspection, trace extraction or child lookup
- **THEN** the request reports cancellation, releases its read resources and cache lock, and leaves independent requests able to read normally

#### Scenario: Cache reuse and source updates
- **WHEN** an interrupted trace read is followed by an independent request
- **THEN** the service can reuse accepted snapshots for the same revision and reads fresh data when the source changes, with no interrupted result admitted to the cache

### Requirement: Report interrupted historical source coverage

Historical hydration SHALL stop native SQLite and Codex rollout work when the request is cancelled, and SHALL preserve usable evidence with explicit incomplete coverage. The interrupted owner SHALL remain outside complete coverage.

#### Scenario: Cancellation during mixed-owner hydration
- **WHEN** historical hydration completes one owner and is cancelled during another owner's SQL read, rollout discovery or JSONL or gzip parsing
- **THEN** the response retains usable evidence with a cancellation diagnostic and incomplete coverage, starts no further native owner reads and leaves the ledger bytes unchanged

#### Scenario: Cancellation before usable evidence
- **WHEN** a historical request is already cancelled or is interrupted before any usable evidence is found
- **THEN** source coverage is unavailable with a cancellation diagnostic and the ledger remains readable
