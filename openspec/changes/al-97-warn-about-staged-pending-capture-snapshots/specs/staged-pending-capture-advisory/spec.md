## Purpose

Help developers notice pending AgenticLens capture snapshots before committing them. [Roadmap ticket](../../../../../alspaces/plan/to-do/AL-97%20warn%20about%20staged%20pending%20capture%20snapshots.md)

## ADDED Requirements

### Requirement: Explicit setup and hook ownership

Setup SHALL offer the existing pre-commit advisory as a separate opt-in and preserve hook ownership.

#### Scenario: Installing or declining the advisory

- **WHEN** the developer reviews the setup choices
- **THEN** setup offers a preview and requires explicit selection before installing or refreshing the advisory in Git's effective hook location. Declining leaves the hook configuration and files intact.

#### Scenario: Another integration owns the hook

- **WHEN** the effective pre-commit hook exists and is foreign or edited
- **THEN** setup preserves it and explains how its owner can add the advisory while retaining the hook's exit behavior.

### Requirement: Inspect the snapshots selected for the commit

The advisory SHALL read staged versions of pending capture snapshots added or changed relative to the current `HEAD` under the committing worktree's shared `agenticlens.pending_dir`.

#### Scenario: Staged content differs from the worktree

- **WHEN** the commit includes a valid pending snapshot and its working copy is different or absent
- **THEN** the advisory classifies the version in the index Git supplied to the hook, including a temporary index for a path-limited commit.

#### Scenario: No pending snapshot is introduced or updated

- **WHEN** the pending snapshot is unstaged, unchanged from the current `HEAD`, removed in the commit, or outside the configured directory
- **THEN** it produces no pending-snapshot warning.

#### Scenario: The pending directory setting is missing or invalid

- **WHEN** an enabled advisory cannot resolve the shared pending directory setting
- **THEN** it reports an advisory configuration diagnostic without scanning a default directory.

### Requirement: Report actionable pending states without blocking

The advisory SHALL distinguish staged snapshots awaiting a ticket from those awaiting a destination, respect a confirmed local discard, and keep commit execution and the index under Git's control.

#### Scenario: Staged snapshots have different pending states

- **WHEN** the index contains snapshots in both pending states
- **THEN** the hook reports both states in its bounded output even when capture health also has an advisory. It gives a way to review or unstage the snapshots and points to ticket binding or publication diagnostics as appropriate.

#### Scenario: Local discard is known or unavailable

- **WHEN** local capture status confirms that the exact staged capture for the committing worktree is discarded
- **THEN** the pending-snapshot warning is suppressed for that snapshot. Unknown or unavailable local status does not suppress a warning supported by the staged snapshot.

#### Scenario: Inspection is disabled or fails

- **WHEN** the local advisory preference is disabled, or enabled inspection cannot complete
- **THEN** a disabled advisory is quiet. An incomplete inspection may emit a bounded diagnostic. AgenticLens does not block the commit, change staging, or start capture or publication.

#### Scenario: An incomplete scan and an existing health finding

- **WHEN** staged inspection is incomplete, classifies no pending snapshots, and the existing health check reports a finding
- **THEN** the hook retains the actionable health finding alongside the incomplete inspection diagnostic.
