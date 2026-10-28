# feat(AL-56)!: enforce schema v4 migration and backfill boundary

## Summary

Declares AL-56 schema v4 as a breaking change. Schema v4 changes the ledger from root-level facts to session-owned invocation, tool, event, and anomaly data. Ordinary runtime ingestion, `generate`, and server/API ledger reads accept canonical schema v4 ledgers, reject v1-v3 without parsing or merging their facts, and reject v4 ledgers that still carry incomplete migration state.

Migration metadata is temporary: it supports resuming multi-step migration and source-backed backfill across systems, then the final successful step removes it. Already-complete canonical v4 ledgers need no migration marker. The server scans ticket workspaces independently, returns valid ledgers and aggregates when others fail, and displays persistent warnings identifying skipped ticket workspaces and remediation. A single bad ledger does not make the platform unavailable.

This release correction also preserves explicit migration support for v1-v3, source-specific OpenCode/Codex backfill, and actionable HTTP/CLI remediation errors.

## Key Changes

- Validate schema version and canonical v4 structure, including session source and non-negative `user_iterations` (zero is valid), before ordinary ingestion.
- Keep resumable migration metadata only during incomplete migration/backfill; remove it on successful completion and accept metadata-free canonical v4.
- Reject schema v1, v2, and v3 and incomplete v4 migration state before parsing or merging their facts.
- Preserve partial workspace and aggregate results, with structured diagnostics exposed to the UI and persistent ticket-specific warnings.
- Preserve explicit schema migration followed by source-backed backfill as the supported upgrade workflow.
- Cover rejected-ledger no-write behavior and no-match/unavailable-source backfill behavior with regression tests.

## Review

Independent reviews and verification are recorded in `alspaces/AL-56/iterations/09/`, `10/`, and `11/`.

## Breaking change and migration

Schema v4 is structurally incompatible with schema v3 and earlier. Ordinary server/API and `generate` loaders no longer parse or merge v1-v3 facts. Operators must run the explicit schema migration and source-backed backfill with the correct harness (`opencode` or `codex`) before using those ledgers. In-progress metadata supports resumption and is removed after the final successful step. If one workspace is invalid, other valid workspaces and aggregates remain available with a warning naming the skipped ticket and reason.

BREAKING CHANGE: Runtime ingestion no longer accepts schema v1-v3 ledgers. Migrate each ledger to schema v4 and complete source-backed backfill before using it; incomplete migration state is rejected until resumed and completed. Invalid ticket workspaces are skipped with a visible warning while valid workspaces continue to load.
