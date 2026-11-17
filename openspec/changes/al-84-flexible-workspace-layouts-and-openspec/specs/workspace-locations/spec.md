## Purpose

Discover telemetry ledgers in explicitly selected directories and preserve their identity and history through movement and publication.

## ADDED Requirements

### Requirement: Explicit roots and bounded ledger-only discovery
The system SHALL accept only `workspaces.roots` entries containing a path and explicit `discovery.max_depth` and `discovery.max_directories`, without root IDs or minimum depth. Writers SHALL initialize depth zero and budget 300; selected repository containers SHALL explicitly use depth one and budget 300.

#### Scenario: Root-only and container selection
- **WHEN** a root uses depth zero
- **THEN** only that directory is checked for a ledger, while depth one also checks immediate children

#### Scenario: No ledger
- **WHEN** a directory contains only planning files or a ticket-like name
- **THEN** AgenticLens creates no listing, placeholder, registration or inferred identity for it

#### Scenario: Partial discovery
- **WHEN** one root is unavailable or a directory budget interrupts the selected scope
- **THEN** valid sibling results remain usable with explicit incomplete diagnostics; reaching the selected depth alone is normal

### Requirement: Ledger identity and shared location resolution
The system SHALL derive canonical ticket identity solely from the supported ledger and use one resolver for listing, detail, artifacts, events/traces, comparison, live selection, manual destination selection and automatic publication.

#### Scenario: Real move and aliases
- **WHEN** a ledger moves between configured roots or is reached through overlapping physical aliases
- **THEN** its recorded identity and supported facts remain accessible at its actual location and it is enumerated once

#### Scenario: Duplicate ticket locations
- **WHEN** distinct ledgers record the same ticket ID
- **THEN** they remain distinct and require explicit location selection rather than merging or first-root precedence

### Requirement: First publication and move-safe refresh
The system SHALL accept an explicit ticket ID and destination for first generation without prior folder registration, and SHALL refresh existing ledgers at their resolved locations without recreating old parents.

#### Scenario: Archive and late refresh
- **WHEN** OpenSpec archives a ledger-bearing change and a later refresh runs
- **THEN** the archived ledger is updated without reactivating or recreating its former folder or inferring lifecycle facts

#### Scenario: Concurrent movement or interruption
- **WHEN** a workspace moves, disappears or is replaced during publication
- **THEN** publication yields one valid update or explicit recoverable retry, preserving history and avoiding writes to an unrelated replacement

### Requirement: Private operational state and portable pending detection
The system SHALL store journals/locks in an explicitly configured writable local `agenticlens.state_dir`, initialized by setup to `~/.agenticlens/state`, separately from versioned binaries and capture registry storage.

#### Scenario: Manual publication without capture
- **WHEN** a manual writer initializes its configured operational state
- **THEN** it requires no capture enablement, installation or registry access, and unavailable operational state stops before public artifact changes

#### Scenario: Reads on another machine
- **WHEN** supported completed history is copied to a machine without sources or private state
- **THEN** ordinary reads remain available without initializing private directories

#### Scenario: Interrupted set without originating journal
- **WHEN** a pending ledger set is moved or copied without its private recovery data
- **THEN** its adjacent transient transaction indicator preserves visible incomplete status; readers do not accept mixed artifacts or clear the indicator to claim success

### Requirement: Preserve source ownership and honest facts
The system SHALL retain exact ticket/source/session ownership within project/worktree boundaries, preserving unavailable or partial owners, other harnesses and legitimate multiplicity while authoritative complete local refresh remains idempotent.

#### Scenario: Mixed local and unavailable history
- **WHEN** one native source is refreshed and another owner is unavailable
- **THEN** only successfully refreshed owners change, other supported history survives, and missing source/cost facts remain explicitly unavailable

### Requirement: Explicit breaking update
The system SHALL reject retired scalar directory configuration, old runtime root overrides and unsupported publication state rather than retaining compatibility parsers, API aliases or automatic conversion.

#### Scenario: Retired configuration or journal
- **WHEN** old configuration or an unsupported pending publication journal is supplied
- **THEN** actionable update/recovery guidance is returned without silently translating it or discarding pending data; established explicit ledger-schema migration remains separate
