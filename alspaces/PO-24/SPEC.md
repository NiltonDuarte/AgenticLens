# Adopt OpenSpec as the AgenticLens workflow and workspace structure

## Goal

Use OpenSpec change folders as the home for planning, delivery evidence and AgenticLens telemetry. Developers should open the same change before and after archive without maintaining a separate AgenticLens workspace or losing its ticket/session associations.

## Context

The user chose OpenSpec for future repository work after the [compatibility assessment](ASSESSMENT.md). Existing `workspaces.dir` configuration can discover ledgers under `openspec/changes`, but detail reads and automatic writers reconstruct `<root>/<ticket>`. An archived `al-123-add-login` becomes `archive/2026-11-13-al-123-add-login`; scanning can still find it while detail fails or a writer recreates an active directory. Multiple roots alone do not fix that mismatch.

OpenSpec already accepts lowercase ticket-prefixed names. The remaining platform problem is stable identity and consistent location resolution. The repository also has agents, skills, setup guidance and an active backlog organized around alspaces/plan conventions. Those instructions must change together with the repository cutover, preserving their approval, role, evidence and history guarantees.

## Expected behavior and acceptance

- **One change home.** Future work uses ticket-prefixed OpenSpec changes with telemetry and delivery evidence inside that change. Canonical ticket identity remains stable when the description or archive path changes. OpenSpec owns change creation; AgenticLens owns source-backed telemetry. No separate per-ticket AgenticLens directory is required.
- **Consistent multi-root operation.** The platform supports explicitly configured active/archive roots and bounded immediate-child discovery for this mode. All readers and publishers resolve the same unique change. Existing single-root recursive configurations remain supported; missing or ambiguous identity is visible rather than silently selecting a destination.
- **Safe lifecycle.** Archive, late telemetry updates and interrupted/concurrent publication preserve supported history and never recreate a missing active change. Binding remains scoped to the existing project/worktree and exact native session; changing a directory does not reassign ownership.
- **OpenSpec-native guidance.** Agent roles, all affected skills/templates, setup materials and repository instructions use one authoritative set of OpenSpec artifacts. Planning readiness, approval to implement, implementation progress, review evidence and archive are distinct. OpenSpec's defaults cannot bypass human scope approval or verification gates.
- **Repository adoption.** Cut over active roadmap work and future ticket creation while preserving ticket IDs, dependency/readiness information and existing links. Historical alspaces ledgers remain readable through an explicit legacy root without copying or rewriting their facts. No parallel editable backlog or new telemetry workspace is introduced.

Ticket-specific acceptance is owned by the linked tickets. Implementing phase analytics, additional harness adapters, cross-repository OpenSpec stores and compressed archive packages is outside this initiative.

## Delivery

[AL-84](../plan/to-do/AL-84%20resolve%20workspaces%20across%20configured%20roots.md) makes multi-root changes readable. [AL-85](../plan/to-do/AL-85%20publish%20telemetry%20safely%20across%20change%20moves.md) preserves publication/binding through moves. [AL-86](../plan/to-do/AL-86%20set%20up%20ticket%20prefixed%20openspec%20changes.md) supplies a usable OpenSpec setup and creation workflow. [AL-87](../plan/to-do/AL-87%20align%20agents%20and%20skills%20with%20openspec.md) aligns operating guidance. [AL-88](../plan/to-do/AL-88%20cut%20over%20the%20repository%20to%20openspec.md) migrates the live repository and backlog. Readiness and dependency edges are recorded only in [the roadmap index](../plan/TICKETS.md).

## Implementation notes

This is product planning, not a developer contract or implementation approval. Preserve the [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) and [regression boundaries](../../docs/REGRESSION_PREVENTION.md). Configuration syntax, custom schema design and bounded resolver contracts belong to development. A one-change-per-ticket rule applies within a selected project; conflicting directory/ledger identities require resolution, not automatic history merging. Source package availability and a suitable pinned OpenSpec release must be verified during setup work.
