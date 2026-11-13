# Adopt OpenSpec as the AgenticLens workflow and workspace structure

## Goal

Use OpenSpec change folders as the home for planning, delivery evidence and AgenticLens telemetry. Developers should open the same change before and after archive without maintaining a separate AgenticLens workspace or losing its ticket/session associations.

## Context

The user chose OpenSpec for future repository work after the [compatibility assessment](ASSESSMENT.md). Existing `workspaces.dir` configuration can discover ledgers under `openspec/changes`, but detail reads and automatic writers reconstruct `<root>/<ticket>`. An archived `al-123-add-login` becomes `archive/2026-11-13-al-123-add-login`; scanning can still find it while detail fails or a writer recreates an active directory. Multiple roots alone do not fix that mismatch.

OpenSpec already accepts lowercase ticket-prefixed names. The remaining platform problem is stable identity and consistent location resolution. The repository also has agents, skills, setup guidance and an active backlog organized around alspaces/plan conventions. Those instructions must change together with the repository cutover, preserving their approval, role, evidence and history guarantees.

## Expected behavior and acceptance

[AL-84](../plan/to-do/AL-84%20adopt%20openspec%20across%20the%20platform%20and%20repository.md) owns the observable acceptance for the complete slice: multi-root discovery and shared location resolution, archive-safe publication, stable ticket/session binding, ticket-prefixed OpenSpec setup, reconciled agents/skills/guidance, and the live repository/backlog cutover. Platform support alone is not completion.

Future work has one change home. Historical alspaces data remains supported through an explicit legacy root rather than being copied or silently reclassified. OpenSpec owns change creation; AgenticLens owns source-backed telemetry. Existing non-OpenSpec configurations remain supported. The single slice includes the repository's operating guidance and migration evidence, preserving human scope approval, role boundaries, review attribution and portable history.

## Delivery

Deliver this initiative through the single AL-84 ticket, as the user explicitly requested after reviewing the initial decomposition. There are no separate rollout tickets or ticket dependencies for this batch; readiness is recorded only in [the roadmap index](../plan/TICKETS.md). This larger end-to-end boundary is intentional. Developers may organize internal work without treating partial platform/guidance/cutover outcomes as delivered acceptance.

## Implementation notes

This is product planning, not a developer contract or implementation approval. Preserve the [portable ledger contract](../../TECHNICAL_GUIDELINES.md#portable-ledger-contract) and [regression boundaries](../../docs/REGRESSION_PREVENTION.md). Configuration syntax, custom schema design and bounded resolver contracts belong to development. A one-change-per-ticket rule applies within a selected project; conflicting directory/ledger identities require resolution, not automatic history merging. Source package availability and a suitable pinned OpenSpec release must be verified during setup work.
