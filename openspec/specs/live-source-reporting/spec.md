# live-source-reporting Specification

## Purpose
Operators can distinguish a supported live source with no active sessions from a source that cannot supply current observations, while retaining historical evidence.

## Requirements

### Requirement: Select the live observer by query capability

The server SHALL start the live poller only for a supported v1 layout and convey the selected source's live status consistently at startup and through live REST and SSE.

#### Scenario: Working v1 reader
- **WHEN** the selected store provides the tables and columns used by the v1 live queries
- **THEN** sessions, traces, polling and subscription cleanup work, including a measured empty tree when no sessions are active, optional message/part update timestamps and the recognized native empty auxiliary table

#### Scenario: Unsupported family
- **WHEN** the selected database opens and contains only a complete v2 family, or contains competing OpenCode families
- **THEN** startup identifies the unsupported layout, live listing, trace and stream requests report that reason and polling stays inactive

#### Scenario: Source diagnosis
- **WHEN** no OpenCode path is configured, a selected file is missing or inaccessible, a SQLite layout is unknown or incomplete, or source inspection fails
- **THEN** operators receive a reason distinguishing these cases through startup and live requests while the dashboard server remains available

### Requirement: Deliver read failures to the live consumer

The live service SHALL communicate a failed read to the requesting client or active stream subscriber and surface that diagnostic in the live view.

#### Scenario: Failure before or during streaming
- **WHEN** a supported source fails the first snapshot or a subsequent tree or trace read
- **THEN** the consumer sees the failure, retains usable last observations with a visible diagnostic, and can observe recovery on a successful refresh, including when the tree data is unchanged

#### Scenario: Supported trace lookup
- **WHEN** a client supplies an invalid ID or requests an absent session on a healthy v1 source
- **THEN** existing invalid-ID and missing-session responses remain distinct from source failure

### Requirement: Preserve independent historical access

Historical events and recorded facts SHALL remain readable according to the existing source and ledger contracts, independently of the live status.

#### Scenario: Exact owner on a clean or mixed source
- **WHEN** historical events are requested for an exact recorded OpenCode owner
- **THEN** clean v1 and v2 layouts can hydrate, while a competing mixed layout reports the current family-selection diagnostic for the affected owner

#### Scenario: Optional enrichment unavailable
- **WHEN** a copied ledger is read without its original machine's sources or with only a subset of its sources present
- **THEN** recorded facts and independently available owner evidence remain accessible, including historical Codex data where configured, and GET leaves the ledger bytes unchanged
