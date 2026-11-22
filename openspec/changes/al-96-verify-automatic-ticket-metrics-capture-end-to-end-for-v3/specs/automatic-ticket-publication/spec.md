## Purpose

Publish and refresh ticket metrics from native session lifecycle events, including capture before ticket binding and the first ledger at an explicitly indicated destination. [Roadmap ticket](../../../../../alspaces/plan/done/AL-96%20verify%20automatic%20ticket%20metrics%20capture%20end%20to%20end%20for%20v3.md)

## ADDED Requirements

### Requirement: Configure the pending snapshot directory explicitly

Pending capture and its consumers SHALL use the shared project setting `agenticlens.pending_dir`, resolved relative to the project root for relative paths.

#### Scenario: Setup writes or updates configuration

- **WHEN** setup writes the shared configuration and no pending directory has been selected
- **THEN** it writes `.agenticlens/pending/` as the explicit value. An explicit selection is written as supplied, and configuration updates preserve an existing valid value unless an explicit replacement is requested.

#### Scenario: Runtime reads configuration

- **WHEN** pending capture or a consumer reads the shared configuration
- **THEN** it uses the configured directory. A missing, empty or invalid setting produces a configuration error without using a fallback, writing configuration or publishing to another directory.

### Requirement: Capture native sessions before ticket binding

Capture SHALL collect source-backed metrics for native sessions in capture-enabled projects without requiring a ticket ID, title, conventional type or destination at session start.

#### Scenario: Session starts without a ticket

- **WHEN** a native session starts work and source-backed metrics become available before binding
- **THEN** the worker writes a snapshot as `<capture-id>.json` in the configured pending directory with a stable capture identity and absent ticket metadata. The status reports that the session is awaiting a ticket. Sessions without a binding may remain visible as pending without acquiring an inferred ticket.

#### Scenario: Ticket is identified later

- **WHEN** the verified native session binds to an explicit ticket ID, title and conventional type
- **THEN** the same pending file gains the ticket metadata and retains previously collected facts, including descendants whose native parentage is verified. If publication has no eligible destination, the status reports that the ticket is awaiting a destination or shows the specific blocking condition.

#### Scenario: Codex archives a session before ticket binding

- **WHEN** an authorized request binds a Codex session whose registered rollout path is missing because Codex archived it
- **THEN** capture can use the archived rollout after verifying the exact native session and project. It stores the verified path with the accepted binding. If source verification fails, capture retains the pending facts and leaves the binding unresolved.

#### Scenario: Pending snapshot is visible in Git

- **WHEN** the worker creates or refreshes a pending snapshot
- **THEN** the file contains portable metrics available for review and normal Git versioning. The worker SHALL NOT stage the file in Git. Canonical ledger discovery excludes pending snapshots, and operational state, journals and raw source records remain private.

### Requirement: Discard and restore a pending capture

AgenticLens SHALL let a user or agent discard and restore a capture by its stable identity, with or without a ticket binding.

#### Scenario: Discard a pending capture

- **WHEN** the user or agent discards a pending capture
- **THEN** AgenticLens records the discard durably in local capture state and removes the pending snapshot. Discard leaves the Git index, native source records and existing canonical ledgers intact.

#### Scenario: A discarded session continues or resumes

- **WHEN** callbacks arrive, the worker restarts, the session resumes, or a descendant of the discarded root is discovered
- **THEN** the worker suppresses pending snapshot recreation and canonical publication for that capture until restore. Capture status can list the discarded root and the identity used to restore it.

#### Scenario: Restore after activity during discard

- **WHEN** the user or agent restores the discarded capture
- **THEN** the worker reconciles all available native history for the same root and its verified descendants, including activity before and during discard, without counting facts twice. The original capture identity and verified binding are retained. The restored capture follows normal pending or publication rules, and missing native history is reported as incomplete coverage.

#### Scenario: Retry an interrupted discard or restore

- **WHEN** either operation is repeated or interrupted while the worker is processing the same capture
- **THEN** recovery honors the durable capture state, retries any incomplete cleanup or reconciliation, and prevents a stale pass from undoing the latest operation.

### Requirement: Publish the first ledger after native ticket binding

Automatic capture SHALL create the first canonical ledger at an explicitly selected, eligible destination after verified ticket binding and source collection.

#### Scenario: Fresh ticket with an agent-indicated destination

- **WHEN** a bound session's agent indicates an exact existing directory within configured discovery policy and no canonical ledger exists for the ticket
- **THEN** the worker validates the destination and commits the canonical `agentic_metrics.json` and configured summary, if enabled, through the normal publication transaction without operator preparation or user confirmation. The worker incorporates facts already collected in the pending snapshot.

#### Scenario: Existing ticket has a single canonical ledger

- **WHEN** discovery resolves one canonical ledger for the verified ticket
- **THEN** the worker publishes to that ledger without requiring a destination indication and preserves its recorded identity and accepted history.

#### Scenario: Missing, unsafe or conflicting selection

- **WHEN** a bound fresh ticket has no usable selection, the destination is outside policy, a conflicting ledger exists, or discovery cannot establish an unambiguous location
- **THEN** the worker reports the blocking condition, keeps the pending snapshot available and preserves other ledgers and accepted facts. It SHALL NOT infer a location from the ticket ID, folder name or OpenSpec metadata.

#### Scenario: Retry or relocation

- **WHEN** the worker retries an interrupted first publication
- **THEN** the existing transaction recovery rules apply and accumulated facts enter the canonical ledger once. The pending snapshot is removed only after publication becomes durable.
- **WHEN** an existing ledger moves within configured roots, or its former destination or ledger is removed after first publication
- **THEN** refresh follows the ledger if found. The old initial selection cannot recreate a removed publication. Historical reads continue to require only recorded ledger facts.

### Requirement: Verify native session coverage and duration on the packaged candidate

The release gate SHALL exercise normal installed callbacks and independently compare committed ledgers with the native sources for each claimed automatic-capture combination.

#### Scenario: Real development session and a later turn

- **WHEN** the exact packaged candidate is installed through supported setup in an isolated Git project and a native root delegates a small code change and verification, waits for both, and binds through the native integration, with nested delegation where supported
- **THEN** the first committed ledger appears within a stated deadline without manual generation, backfill, seeding or direct worker invocation. The source and ledger agree by exact session ID, harness, parent link, tool activity and available usage. All same-ticket descendants appear once and an unrelated session is absent. A subsequent normal turn preserves earlier owners and counts cumulative facts once.

#### Scenario: Native completion settles previously provisional time

- **WHEN** a committed snapshot records positive provisional duration for an open interval and native source evidence later confirms its completion
- **THEN** the first automatic refresh that sees the completion preserves the full measured duration once and reports zero provisional seconds for the completed session. The ticket rollup retains only provisional contributions still open or unconfirmed. The gate records per-session and ticket values and the elapsed time to that commit, fails on overdue or unresolved publication, and requires a corrected rerun for supported-path defects.

#### Scenario: Native completion precedes binding or destination selection

- **WHEN** native completion becomes observable while a snapshot is still pending
- **THEN** the worker settles the completed interval in the pending snapshot. A later binding and publication preserve that measured duration exactly once. The packaged gate observes both pending states and subsequent publication without operator preparation or manual generation.

#### Scenario: Support and evidence are explicit

- **WHEN** a run is evaluated for release
- **THEN** its record names the exact candidate and harness versions, expected and observed session trees, and result. Manual-only combinations are identified separately. The stable gate passes only with successful runs on the corrected packaged candidate for all claimed automatic combinations.
