## Purpose

Stop interrupted native refresh work and preserve each workspace's recorded telemetry and publication consistency.

## ADDED Requirements

### Requirement: Stop cancelled backfill source work

Source-backed backfill SHALL propagate caller cancellation and the workspace deadline through native discovery and extraction. An interrupted read SHALL release its resources and SHALL NOT certify an owner as completely read or empty.

#### Scenario: Interruption inside an owner
- **WHEN** a backfill caller cancels or its workspace deadline expires during OpenCode v1 or v2 discovery or snapshot reads, or during Codex rollout discovery or parsing
- **THEN** the current native read stops at its query or read boundary, rows, files and gzip readers close, and an exact SQLite snapshot ends on its dedicated handle
- **AND THEN** the result reports cancellation or deadline expiry without starting further reads within the cancelled workspace

### Requirement: Preserve the workspace when cancellation precedes publication

Backfill SHALL leave the current workspace's ledger and configured summary byte-for-byte intact when it observes cancellation before publication starts. Workspaces already committed SHALL retain their updates. Publication already underway SHALL resolve through the existing commit or recovery protocol.

#### Scenario: Complete owner followed by an interrupted owner
- **WHEN** backfill fully reads one owner and is cancelled during a second owner's read before publication
- **THEN** the current workspace retains its original ledger and summary, the report identifies cancellation, and previously committed workspaces remain published

#### Scenario: Ordinary partial refresh
- **WHEN** an owner is unavailable or unreadable and the caller remains active
- **THEN** the existing partial refresh can publish successfully read owners, preserve unavailable owners and other harness facts, and report incomplete coverage
