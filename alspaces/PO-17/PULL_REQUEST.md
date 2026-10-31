# docs(PO-17): reconcile approved AL-55 backfill scope

AL-55 still asked developers to recreate schema v4 and add a nonexistent client migration. Replace those stale requirements with the user-approved outcome: safely refresh existing ledgers, preserve unavailable history and whole skill/task snapshots during partial recovery, and report incomplete recovery truthfully.

Keep v4 fields optional at the file boundary and normalize omitted collections to required empty arrays internally. Explicit recompute works without Force; routine skips, all-required-owner completion, valid baseline eligibility after optional refresh failure, and existing migration-backup/atomic-write interfaces remain the agreed boundaries. No new format or mandatory migration campaign is required.

PO-17 also records the independent AL-57 review and AL2 conflicts. Cost semantics remain proposals, cost changes are deferred, and AL2 development is queued. Only the released AL-55 specification and PO-17 artifacts change; index transitions and all application implementation remain with separate owners. The AL-55 developer receives lifecycle ownership of the spec after this PR merges and manager handoff.

Validation: documentation/path-scope review and integrated full make; detailed results and independent Codex session telemetry are recorded in PO-17 SUMMARY.md and iterations/02/REVIEW.md. This PR specifies future AL-55 behavior; it does not implement or prove those fixes. No release action.
