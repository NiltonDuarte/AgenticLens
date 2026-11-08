# REV-7: Review source read guarantees

## Goal

Produce an evidence-backed, bounded read contract and remediation plan that lets consumers share meaningful operations without erasing source-specific capabilities.

## Context

The sources normalize into common facts but consumers still know storage details and follow different read rules. Backfill chooses the v2 table when it exists; exact-owner hydration can read a uniquely owned v1 session in a mixed database. Exact SQLite capture is transactional, ordinary reads span queries, and Codex discovery/parsing are separate file reads. V1 timestamp ties also have weaker ordering than v2 native sequence.

Example: a mixed database contains a v1-only owner A and unrelated v2 rows. Exact historical hydration can find A, while backfill's v2 tree query cannot. That is a documented policy divergence, not evidence that mixed-schema recovery was promised. Concurrent writers and tied timestamps are investigation cases, not established corruption incidents.

## Acceptance criteria

* State what complete, partial, unavailable, unsupported and complete-empty mean for each operation, and which result permits owner replacement. Preserve portable recorded history without private receipts.
* Investigate the mixed-family backfill/exact-hydration discrepancy. Explain the supported policy, show its consumer impact, and propose a bounded correction or an explicitly approved compatibility change; do not silently broaden generic generate's mixed-layout acceptance.
* Establish observation guarantees separately from parsing completion. Inspect concurrent writer behavior, source-revision changes and transaction/file lifetime; label unexecuted probes and hypotheses honestly.
* Define native sequence versus timestamp/tie behavior for the consumers that need ordering, preserving distinct equal-valued occurrences. State whether a user-visible defect is proven before recommending a fix.
* Propose the smallest useful consumer-owned read capabilities. Keep ledger persistence, root discovery, optional children, historical events and live observation distinct; retain legitimate v1/v2 adapter dispatch. Reject a single all-purpose repository if it only forwards existing coupling.
* Include a proposed dependency map, compatibility guarantees, concrete verification cases and prioritized follow-up slices. Reuse AL-67 evidence instead of repeating the whole audit. AL-63 keeps capability/version-matrix ownership.

## Implementation notes

Finding references: F07, F08, F09, F12

Review only: no application, schema, runtime-script or new test changes and no broad refactor authorization. FIX-18/FIX-19/FIX-20 and AL-69/AL-70/AL-74 own their concrete corrections; their completion is not a prerequisite for this investigation.

* [AL-67 assessment](../../AL-67/ARCHITECTURE_ASSESSMENT.md), F07, F08, F09, F12; application baseline a84adbe5be46c2ef52ede0987a417562715d38e1. Findings are unimplemented.
* [Follow-up ownership](../../AL-67/iterations/02/FOLLOW_UP_TICKETS.md).

*DEVELOPMENT STATE REVIEW (2026-11-05)*

Ready, review-only. AL-67 delivered the broad architecture assessment; this narrower follow-up owns remaining mixed-family, snapshot, ordering and resource guarantees. It does not block concrete fixes.

Evidence and review limits: [PO-19 reconciliation](../../PO-19/SPEC.md), merged baseline `03f5b05d`. This planning review does not certify runtime behavior or close unverified acceptance criteria.

[Engineering rules](../../../TECHNICAL_GUIDELINES.md) and [relevant regression boundaries](../../../docs/REGRESSION_PREVENTION.md) apply. Historical findings require revalidation; this rewrite is not runtime verification or implementation approval.
