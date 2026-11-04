# chore(GEN-21): migrate ledgers and print completed migration audits in console

Explicit migration and backfill print completed steps, loss accounting, ledger paths and rollback guidance in the console, then omit completed/no-pending migration objects from persisted ledgers. Incomplete migrations retain resumable state. Metadata-only cleanup preserves every unrelated fact and timestamp; malformed raw status, duplicate members and invalid current facts fail without rewriting. Optional migrate --backup retains originals for rollback; console text alone cannot reconstruct discarded payloads.

This ticket also migrates all 119 existing ledgers to schema 6 and backfills the 585 recorded Codex sessions across 51 ledgers from exact native source IDs. History contains 1,098 sessions and 1,100 invocations, including 73 skills and 395 delegations in the refreshed ledgers. All 513 other-harness snapshots survive. Existing completed metadata is cleaned using the changed CLI with original backups; per-iteration manifests distinguish historical migration/backfill checkpoints from the final cleanup.

No shared pin or capture setup changes. The original release migration dropped one unsupported AL-55 tool-summary field, content_errors: 1, documented with the original retained in Git. GEN-21's own telemetry remains blocked by incompatible local capture-registry coordinates; no empty ledger or successful extraction is claimed.

Scope and verification: [CONTRACT.md](alspaces/GEN-21/CONTRACT.md) and [SUMMARY.md](alspaces/GEN-21/SUMMARY.md).
