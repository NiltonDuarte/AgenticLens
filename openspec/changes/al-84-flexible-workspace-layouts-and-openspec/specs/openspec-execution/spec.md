## Purpose

Keep the long-term roadmap distinct from approved execution work while preserving specialist gates, evidence and native telemetry through OpenSpec archive.

## ADDED Requirements

### Requirement: Roadmap and execution remain separate
The repository workflow SHALL retain roadmap ticket documents and `alspaces/plan/TICKETS.md` in place, creating linked ticket-prefixed OpenSpec changes only when work is selected.

#### Scenario: Ticket reprioritization
- **WHEN** a roadmap ticket is created, edited or reprioritized
- **THEN** no execution change or amended implementation approval is inferred

#### Scenario: Selected work and completion
- **WHEN** work starts for a ticket
- **THEN** the ticket and change link to each other, the change owns approved scope/design/tasks/evidence/telemetry, and archive or checkboxes alone do not complete the ticket

### Requirement: Supported pinned OpenSpec workflow
The repository SHALL use the verified pinned OpenSpec CLI/schema and supported creation/templates, preserving custom configuration and instructions without an upstream fork or assumed enforcement of advisory rules.

#### Scenario: Canonical ticket and description
- **WHEN** a selected canonical configured ID and description create a change
- **THEN** OpenSpec creates the validated lowercase folder without collisions, and first telemetry generation separately records the canonical ID supplied explicitly

### Requirement: Approval, roles and evidence survive adoption
The workflow SHALL preserve PO planning-only sessions, explicit approval before coding, specialist boundaries, configured independent review/model attribution, required verification, final telemetry and PR handoff in one execution home.

#### Scenario: Readiness versus approval
- **WHEN** planning artifacts or tasks are complete
- **THEN** neither coding approval nor verified delivery is inferred from those states

#### Scenario: Execution artifact adoption
- **WHEN** configuration, guidance and AL-84 execution artifacts move to OpenSpec
- **THEN** roadmap/history remain intact, actual writable-file moves are coordinated with owners, and reviewed Git changes plus documented rollback preserve recoverability without a separate migration framework

#### Scenario: Native lifecycle evidence
- **WHEN** the selected workflow is delivered
- **THEN** evidence covers real Codex/OpenCode source publication, CLI/API/dashboard detail, retained incomplete-review history, archive and late refresh; fixtures or text searches alone do not establish native compatibility
