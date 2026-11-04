# chore(GEN-21): migrate ledgers and recover Codex telemetry

Migrate all 119 existing ledgers to schema 6, backfill the 585 recorded Codex sessions across 51 ledgers, and remove 118 completed migration objects. Explicit migration/backfill now print completed steps, loss accounting, ledger paths and rollback guidance in the console. Incomplete migration state stays resumable; metadata-only cleanup preserves every unrelated fact and timestamp. Original history retains all 1,098 sessions and 1,100 invocations, including 73 skills and 395 delegations in the refreshed ledgers; all 513 other-harness snapshots survive.

Integrates reviewed FIX-17 from PR #111: an explicit manual healthy/fresh capture directory can recover past known unsupported configured coordinate layouts. Configured/selected/discoverable installation checks still apply, and native/default/automatic capture stays strict. GEN-21's final source-backed extraction and original-registry preservation evidence are recorded in iteration 04. This is source-built recovery; shared release pin and existing registry/config are unchanged, and automatic capture health is not claimed.

Optional migrate --backup retains originals; console audit alone cannot reconstruct discarded payloads. The earlier release migration removed one unsupported AL-55 tool-summary field, content_errors: 1, documented with its original preserved in Git. Per-iteration manifests distinguish historical migration, backfill, metadata cleanup and capture recovery.

Scope, findings and verification: [CONTRACT.md](alspaces/GEN-21/CONTRACT.md), [SUMMARY.md](alspaces/GEN-21/SUMMARY.md).
