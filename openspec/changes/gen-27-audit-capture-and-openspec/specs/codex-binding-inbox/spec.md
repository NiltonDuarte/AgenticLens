## Purpose

A sandboxed Codex agent can answer its own AgenticLens ticket-binding request after one-time installation. The trusted worker owns private state and publication.

## ADDED Requirements

### Requirement: Install a dedicated Codex binding inbox

The Codex user installer SHALL configure workspace-write access to a dedicated binding inbox under the effective Codex user home. The installer SHALL preserve the sandbox mode, hook trust, user settings and existing writable roots, and SHALL keep native executables and private capture/publication state outside the new grant.

#### Scenario: Preview and repeat installation

- **WHEN** an operator previews and then installs user hooks for a supported Codex project
- **THEN** the preview reports the inbox and configuration changes without applying them, installation provisions the grant, and a repeat installation leaves equivalent settings intact.

#### Scenario: Existing settings or migration

- **WHEN** a user with existing sandbox roots and hook configuration installs, migrates or removes the integration
- **THEN** the operation preserves unrelated configuration, pre-existing grants and existing capture history, and uninstall removes only the grant the installer can establish it added.

### Requirement: Bind an exact native registration using the inbox

The trusted Codex callback SHALL provide a sandbox-compatible way for the agent to submit its explicitly known ticket ID, title and optional type. A trusted consumer SHALL authenticate all submitted ticket fields against the exact registration and expected binding request before committing that binding. Inbox files SHALL omit the reusable binding challenge.

#### Scenario: Fresh root submits a binding

- **WHEN** a fresh workspace-write Codex root receives its own request and submits its explicitly supplied ticket through the installed inbox
- **THEN** it can submit without a per-run directory flag and the worker can bind its verified source identity and publish to the existing selected ledger and summary.

#### Scenario: Invalid or replayed submission

- **WHEN** a submission carries altered ticket fields, an invalid authenticator, a superseded request or a different worktree or session
- **THEN** the worker preserves the registration and previously accepted facts.
- **WHEN** an identical authenticated request is submitted more than once
- **THEN** processing is idempotent.

#### Scenario: Shared inbox and bounded consumption

- **WHEN** multiple sessions or projects use the same Codex home and the inbox contains unrelated files
- **THEN** each worker inspects only bounded response locations derived from its known registrations, preserves unrelated files, and restricts filesystem access to regular inbox entries.

#### Scenario: Verification and source failure

- **WHEN** authoritative project or source identity cannot yet be verified for an otherwise valid response
- **THEN** the worker preserves accepted history, keeps the pending work visible and retries through the capture lifecycle.

### Requirement: Select an operator-installed engine for trusted Codex capture

Trusted Codex callbacks and workers acting on their behalf SHALL execute only an operator-installed engine from a verified location outside the workspace and the sandbox’s writable grants. The trusted selection SHALL precede executable probes and survive worker restarts and handoffs.

#### Scenario: Capture from a development worktree

- **WHEN** Codex invokes a trusted lifecycle or binding wake hook in a worktree configured with a local engine
- **THEN** capture uses the protected installation and preserves the normal workspace-specific capture policy, ticket binding and destination.
- **WHEN** the worker restarts or hands work to a successor
- **THEN** the successor meets the same installation and path checks, including when the project captures multiple sources.

#### Scenario: Protected installation cannot be verified

- **WHEN** the protected executable is unavailable, incompatible, modified, or reachable through a sandbox-writable location
- **THEN** capture reports the installation problem before executing an unverified candidate, preserves pending work and previously published facts, and performs no automatic download.

#### Scenario: Ordinary development and OpenCode use

- **WHEN** an operator invokes the ordinary CLI in a development checkout or uses the supported OpenCode integration
- **THEN** ordinary development engine selection and independent OpenCode capture retain their supported behavior.
