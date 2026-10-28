# feat(AL-56)!: enforce schema v4 migration and backfill boundary

## Summary

Declares AL-56 schema v4 as a breaking change. Schema v4 changes the ledger from root-level facts to session-owned invocation, tool, event, and anomaly data. Ordinary runtime ingestion, `generate`, and server/API ledger reads now accept only complete, runtime-ready schema v4 ledgers. They reject v1-v3 without parsing or merging their facts, and reject migration-only or incomplete v4 ledgers.

Legacy ledgers must be explicitly migrated to schema v4 and then source-backed backfilled before runtime use. Shape migration alone leaves `source_backfill` pending. Backfill marks a ledger ready only when the selected source can be read and facts are successfully recovered. Missing or unavailable source data leaves the ledger incomplete and runtime-ineligible.

This release correction also preserves explicit migration support for v1-v3, source-specific OpenCode/Codex backfill, and actionable HTTP/CLI remediation errors.

## Key Changes

- Validate schema version and canonical v4 structure, including session source and non-negative `user_iterations` (zero is valid), before ordinary ingestion.
- Persist explicit source-backfill readiness; reject missing readiness metadata, migration-only ledgers, and pending/failed backfills.
- Reject schema v1, v2, and v3 at workspace listing, ticket detail, and events API boundaries without partial successful responses.
- Preserve explicit schema migration followed by source-backed backfill as the supported upgrade workflow.
- Cover rejected-ledger no-write behavior and no-match/unavailable-source backfill behavior with regression tests.

## Review

Independent iteration 10 review: APPROVED. Implementation and acceptance details are recorded in `alspaces/AL-56/iterations/10/`.

## Breaking change and migration

Schema v4 is structurally incompatible with schema v3 and earlier. Ordinary server/API and `generate` loaders no longer parse or merge v1-v3 facts. Operators must run schema migration, then source-backed backfill with the correct harness (`opencode` or `codex`) before using those ledgers. A shape-only migration is not sufficient.

BREAKING CHANGE: Runtime ingestion no longer accepts schema v1-v3 ledgers. Migrate each ledger to schema v4 and complete source-backed backfill before using it; migration without backfill remains incomplete and is rejected.
