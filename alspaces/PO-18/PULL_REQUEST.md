# docs(PO-18): plan Phase 1 post-v2.0 follow-up tickets

Create nine Phase 1 AL tickets (AL-60 through AL-68) from the 2026-10-31 review: ticket search, historical dates, Codex skill/delegation investigation, active OpenCode verification, three-source capability audit, provisional Codex turn-duration fix, safe backfill-default research, source-owned skills/task_links with compatible migration, and configurable optional generate output.

Add a PO specification and reciprocal roadmap dependencies. The diagnosed duration bug can proceed independently; research and verification remain separate from implementation approval. Preserve existing AL-57/AL2 plans and the original AL-55 evidence. This PR delivers planning documents only and does not merge the original review or implement application fixes.

Validation: document structure/link/numbering and unchanged AL-55 SHA256 checks passed; full make passed (Go race tests, frontend typecheck/lint, 741 frontend tests, embedded assets and binary builds). Details in alspaces/PO-18/iterations/01/REVIEW.md. Runtime investigation results are acceptance work for the new tickets, not claims of this PR. Session telemetry is generated from the Codex source during finalization, with extraction limitations reported explicitly.

Iteration 02 adds AL-67: canonical facts use their narrowest proven session/invocation owner; broader views derive from those facts. Preserve legacy facts with unresolved ownership explicitly, repeated occurrences, child relationship direction, and stored baseline behavior when source hydration is unavailable. Documentation-only follow-up reuses prior make results for unchanged application code and validates the new ticket/roadmap links.

Iteration 03 adds AL-68: independent enabled-by-default config toggles for skills/task_links/anomalies. Disabled categories are omitted at all ownership scopes and from managed summary metrics, including when regenerating an existing target. Preserve unrelated facts and distinguish disabled output from unavailable evidence. Documentation-only changes reuse the verified unchanged application checks.
