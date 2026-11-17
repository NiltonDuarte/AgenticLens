# Support user-defined workspace layouts and adopt OpenSpec

## Goal

Let users organize tickets and telemetry in whatever workspace directories fit their workflow, with stable ticket/session binding across those locations. Apply this generic capability to adopt OpenSpec for this repository, keeping each change’s planning, delivery evidence and telemetry together before and after archive.

## Context

The user chose OpenSpec for future repository work after the [compatibility assessment](ASSESSMENT.md). Existing `workspaces.dir` configuration can discover ledgers under `openspec/changes`, but detail reads and automatic writers reconstruct `<root>/<ticket>`. An archived `al-123-add-login` becomes `archive/2026-11-13-al-123-add-login`; scanning can still find it while detail fails or a writer recreates an active directory. Multiple roots alone do not fix that mismatch.

The user explicitly clarified that multiple roots are first-class platform support, not an OpenSpec feature plus an alspaces compatibility path. A user may choose `alspaces`, `alspaces/archive`, grouped `alspaces/al`, `alspaces/al2`, `alspaces/po`, `alspaces/fix` or documentation folders, arbitrary renamed/nested directories, an OpenSpec tree, or any combination. Paths and routing belong to the user; no directory name implies a ticket type, completion state or default destination. This supersedes the earlier legacy-root/managed-OpenSpec framing.

OpenSpec already accepts lowercase ticket-prefixed names. The remaining platform problem is stable identity and consistent location resolution. The repository also has agents, skills, setup guidance and an active backlog organized around alspaces/plan conventions. Those instructions must change together with the repository cutover, preserving their approval, role, evidence and history guarantees.

## Expected behavior and acceptance

[AL-84](../plan/done/AL-84%20support%20flexible%20workspace%20layouts%20and%20adopt%20openspec.md) owns the observable acceptance for the complete slice: multi-root discovery and shared location resolution, archive-safe publication, stable ticket/session binding, ticket-prefixed OpenSpec setup, reconciled agents/skills/guidance, and the live repository/backlog cutover. Platform support alone is not completion.

Workspace discovery, location resolution, binding and publication use one generic contract for every configured root. Single-directory configuration is the same model with one root. Existing bound workspaces follow moves; new workspace creation requires an explicit selected destination or unambiguous user-configured route. Multiple possible locations require user disambiguation, not a directory-name guess.

This repository chooses OpenSpec as the home for future work; alspaces and other layouts remain equally supported ongoing platform choices. OpenSpec should therefore work out of the box when its directories are configured, using the same core read/bind/publish behavior. OpenSpec owns its change creation; AgenticLens owns source-backed telemetry. The single slice includes the repository's operating guidance and migration evidence, preserving human scope approval, role boundaries, review attribution and portable history.

## Delivery

Deliver this initiative through the single AL-84 ticket, as the user explicitly requested after reviewing the initial decomposition. There are no separate rollout tickets or ticket dependencies for this batch; readiness is recorded only in [the roadmap index](../plan/TICKETS.md). This larger end-to-end boundary is intentional. Developers may organize internal work without treating partial platform/guidance/cutover outcomes as delivered acceptance.

## Implementation notes

This is product planning, not a developer contract or implementation approval. Preserve the [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) and [regression boundaries](../../docs/REGRESSION_PREVENTION.md). Configuration syntax, custom schema design and bounded resolver contracts belong to development. Bindings select a clear workspace for the canonical ticket within its project/worktree; multiple matches or conflicting recorded identities require explicit resolution, not automatic merging. Root discovery remains bounded and containment-safe, without prescribing one global depth or a fixed folder taxonomy. Source package availability and a suitable pinned OpenSpec release must be verified during setup work.
