# docs(PO-16): reconcile ticket index and remaining specifications

AL-49 and AL-50 were merged but still appeared as Ready with nonexistent to-do paths. Remove those completed entries, move dependency-free AL-57 to Ready, and synchronize AL-52 through AL-55 with the merged AL-51 prerequisite.

Annotate AL-47's existing streaming implementation and AL-55's overlap with AL-56, preserving unfinished work. Correct AL-53's empty-data acceptance wording to supported v4 ledgers; the runtime already requires explicit migration/backfill for v1-v3.

All 15 remaining to-do specifications were reviewed. No remaining ticket is marked complete on partial implementation evidence. The AL-55 original proposal remains recorded and explicitly requires replanning against the shipped v4 boundary.

Validation and per-ticket evidence: `alspaces/PO-16/iterations/01/REVIEW.md`. Documentation-only; no application or test changes.

Track this planning-maintenance audit under PO-16. AL-49 retains its original feature-delivery artifacts and telemetry; the current audit session is recorded only under PO-16.
