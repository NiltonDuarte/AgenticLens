## Purpose

Publish and refresh ticket metrics from native session lifecycle events, including the first ledger for an explicitly selected destination. [Roadmap ticket](../../../../../alspaces/plan/to-do/AL-96%20verify%20automatic%20ticket%20metrics%20capture%20end%20to%20end%20for%20v3.md)

## ADDED Requirements

### Requirement: Publish the first ledger after native ticket binding

Automatic capture SHALL create the first canonical ledger at an explicitly selected, eligible destination after verified ticket binding and source collection.

#### Scenario: Fresh ticket with a valid selection

- **WHEN** the operator selects an existing directory within configured discovery policy, the directory has no ledger, and the installed lifecycle receives a verified ticket binding with its explicit ID, display title and conventional type
- **THEN** normal callbacks and the worker commit source-backed `agentic_metrics.json` and the configured summary, if enabled, through the normal publication transaction. The operator can observe whether coverage and publication are complete.

#### Scenario: Missing, unsafe or conflicting selection

- **WHEN** a bound fresh ticket has no usable selection, the destination is outside policy, a conflicting ledger exists, or discovery cannot establish an unambiguous location
- **THEN** the worker reports the blocking condition and preserves other ledgers and accepted facts. It SHALL NOT infer a location from the ticket ID, folder name or OpenSpec metadata.

#### Scenario: Retry or relocation

- **WHEN** the worker retries an interrupted first publication
- **THEN** the existing transaction recovery rules apply and at most one ledger is finalized at the authorized destination.
- **WHEN** an existing ledger moves within configured roots, or its former destination or ledger is removed after first publication
- **THEN** refresh follows the ledger if found. The old initial selection cannot recreate a removed publication. Historical reads continue to require only recorded ledger facts.

### Requirement: Verify native session coverage and duration on the packaged candidate

The release gate SHALL exercise normal installed callbacks and independently compare committed ledgers with the native sources for each claimed automatic-capture combination.

#### Scenario: Real development session and a later turn

- **WHEN** the exact packaged candidate is installed through supported setup in an isolated Git project and a native root binds, delegates a small code change and verification, and waits for both, with nested delegation where supported
- **THEN** the first committed ledger appears within a stated deadline without manual generation, backfill, seeding or direct worker invocation. The source and ledger agree by exact session ID, harness, parent link, tool activity and available usage. All same-ticket descendants appear once and an unrelated session is absent. A subsequent normal turn preserves earlier owners and counts cumulative facts once.

#### Scenario: Native completion settles previously provisional time

- **WHEN** a committed snapshot records positive provisional duration for an open interval and native source evidence later confirms its completion
- **THEN** the first automatic refresh that sees the completion preserves the full measured duration once and reports zero provisional seconds for the completed session. The ticket rollup retains only provisional contributions still open or unconfirmed. The gate records per-session and ticket values and the elapsed time to that commit, fails on overdue or unresolved publication, and requires a corrected rerun for supported-path defects.

#### Scenario: Support and evidence are explicit

- **WHEN** a run is evaluated for release
- **THEN** its record names the exact candidate and harness versions, expected and observed session trees, and result. Manual-only combinations are identified separately. The stable gate passes only with successful runs on the corrected packaged candidate for all claimed automatic combinations.
