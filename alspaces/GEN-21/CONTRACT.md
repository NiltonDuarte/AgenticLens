# GEN-21: Migrate existing ledgers to schema 6

User authorization: "$agentic-lens-operator GEN ticket - update all existing ledger to version 6" (2026-11-04).

Scope: explicitly migrate every existing agentic_metrics.json below configured alspaces, including nested iteration history, using published AgenticLens v3.0.0-rc.125+86bea59 at baseline 86bea59. Preserve sessions, invocation facts, costs and existing supported history. Apply only the documented schema transitions, including audited retirement of rich payloads and ownerless occurrences. Keep pre-migration originals outside Git; Git also retains committed originals. Do not backfill sources or manufacture enrichment. Record pending migration work explicitly. No application, tests, configuration or hook changes.

Acceptance: every original ledger is schema 6; identity/session sets and recorded invocation facts survive; existing schema-6 ledger remains byte-identical; migration audits retained; repeat migration makes no changes; CLI scan accepts resulting history. Publish task artifacts and migration records through a chore PR. Source-backed completion is outside this shape migration and must remain explicitly pending where reported.

## Iteration 02 authorized extension

User subsequently requested "backfill codex ledgers". The current contract includes exact-ID Codex source refresh from this machine across existing ledgers, preserving unavailable/partial and other-harness owners. Successful complete local snapshots replace recorded snapshots even with lower totals. This supersedes the original no-backfill boundary and byte-identical-AL-62 requirement for the new iteration only. Original migration verification remains historical evidence. See iterations/02/PLAN.md.

## Iteration 03 authorized extension

The user subsequently requested removal of completed migration objects and console output so users can revert changes. This explicitly authorizes application changes limited to completed/no-pending migration metadata cleanup, console/report audit and rollback guidance, plus cleaning the existing migrated ledgers through the changed CLI. Pending/partial state remains persisted and resumable. Retain the pure migration audit as a transient result; omit completed metadata only at explicit persistence boundaries, never during ordinary reads. Preserve all other recorded ledger facts and existing explicit backup behavior. This supersedes the earlier requirement to retain completed migration audit inside canonical ledgers. See iterations/03/PLAN.md for acceptance, ownership, verification and review assignments.
