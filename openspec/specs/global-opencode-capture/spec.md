# global-opencode-capture Specification

## Purpose
A single user-installed OpenCode plugin captures each configured session in its own checkout and publishes source-backed ticket metrics through the common native worker.

## Requirements

### Requirement: Additive user installation
The installer SHALL respect the effective OpenCode user configuration directory, preserve unrelated integrations and prevent duplicate owned plugin loading.

#### Scenario: Existing plugin and legacy integration
- **WHEN** a user installs, inspects, explicitly migrates or removes the integration in a custom configuration home
- **THEN** other plugins, configuration and pending capture survive. Migration removes only unchanged owned legacy files. Conflicting or modified files are identified and preserved.

### Requirement: Runtime project and source selection
The integration SHALL use the runtime session's actual directory and native project configuration for lifecycle, binding request and binding response.

#### Scenario: One server serves several checkouts
- **WHEN** a session runs in a nested directory of an ordinary checkout or either linked worktree while the server runs elsewhere
- **THEN** the active checkout's shared file and its whole local override or whole base-local fallback determine policy. Relative local paths start at the selected file. Missing shared policy skips, invalid required configuration errors, disabled capture avoids capture-only I/O, inspection is read-only and the first enabled event initializes that worktree.

#### Scenario: Source or executable selection fails
- **WHEN** runtime evidence contradicts the configured source, the selected family cannot be read, or the selected development output is invalid
- **THEN** the failure is visible and known facts survive. Omitting development mode selects the exact installed pin. Missing output in a configured development directory reports an error. A running worker retains its selected build.

### Requirement: Exact session binding and publication
The integration SHALL bind exact runtime sessions, collect from their authoritative source and publish only bound facts.

#### Scenario: Root and child perform the same task
- **WHEN** the root receives an explicit current-task ticket and a native binding request
- **THEN** the binding tool answers directly using runtime identity. Ambiguous or missing context requests clarification and leaves local facts unassigned until binding. Only a source-proven same-task child inherits, and an existing explicit child binding prevails.

#### Scenario: Activity, checkpoint and retry
- **WHEN** notifications overlap, a source read is incomplete, or publication is interrupted
- **THEN** the shared worker coalesces notifications, publishes during activity and at checkpoints, retries within its budget, and retains unresolved work and previously accepted owners. Deleting a destination preserves state and reports the missing location.

#### Scenario: Worktrees and portable ledgers
- **WHEN** multiple sessions publish different tickets or share one ticket, or a copied mixed-harness ledger is refreshed without the producer's private state
- **THEN** each checkout writes to its selected destination, concurrent contributors survive, and only exact owners with complete local snapshots are refreshed. Other harnesses and unavailable owners retain recorded facts.
