# GEN-21: Migrate existing ledgers to schema 6

User authorization: "$agentic-lens-operator GEN ticket - update all existing ledger to version 6" (2026-11-04).

Scope: explicitly migrate every existing agentic_metrics.json below configured alspaces, including nested iteration history, using published AgenticLens v3.0.0-rc.125+86bea59 at baseline 86bea59. Preserve sessions, invocation facts, costs and existing supported history. Apply only the documented schema transitions, including audited retirement of rich payloads and ownerless occurrences. Keep pre-migration originals outside Git; Git also retains committed originals. Do not backfill sources or manufacture enrichment. Record pending migration work explicitly. No application, tests, configuration or hook changes.

Acceptance: every original ledger is schema 6; identity/session sets and recorded invocation facts survive; existing schema-6 ledger remains byte-identical; migration audits retained; repeat migration makes no changes; CLI scan accepts resulting history. Publish task artifacts and migration records through a chore PR. Source-backed completion is outside this shape migration and must remain explicitly pending where reported.
