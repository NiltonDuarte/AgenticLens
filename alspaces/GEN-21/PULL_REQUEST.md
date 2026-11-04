# chore(GEN-21): migrate ledgers to schema 6 and backfill Codex history

Migrate all 119 existing ledgers to schema 6, then refresh the 585 recorded Codex sessions across 51 ledgers from this machine's native Codex source store. Refresh their summary tables, including nested AL-68 iteration history. Source-backed history now includes 73 skill occurrences and 395 delegations in the selected ledgers. Exact IDs only; no heuristic discovery.

All 1,098 session identities and 1,100 invocation counts remain intact. All 513 non-Codex snapshots and the 68 ledgers without Codex owners remain unchanged by backfill. Complete native Codex snapshots replace their recorded facts; 292 changed and 293 remained equivalent. All ledgers pass schema-6 ingestion, with no pending migration steps or incomplete selected-source refreshes. The batch CLI's 68 diagnostics refer to ledgers without Codex owners; incidental metadata/summary rewrites for those ledgers were excluded.

Migration used the published checksum-verified rc.125 native matching source baseline 86bea59. Its decoder removed one unsupported AL-55 tool-summary field, content_errors: 1, during the original migration. This is explicitly audited. Original ledgers remain in Git and private backups. No application/configuration changes. Shared release pin remains unchanged and requires separate setup/upgrade for ordinary schema-6 operation.

Validation: full make passed for unchanged application code, including Go race tests, 824 frontend tests and builds. Migration dry run, actual migration and idempotence check passed. Codex backfill dry run and persisted refresh passed, with exact owner-preservation checks and successful CLI ingestion. Per-ledger evidence is in alspaces/GEN-21/iterations/01/MIGRATION_RESULTS.json and iterations/02/BACKFILL_RESULTS.json; the latter records current post-backfill hashes.

GEN-21's own session telemetry remains blocked by incompatible local capture-registry coordinates. No empty ledger was fabricated or existing capture state reset. This limits the operation's telemetry, not the completed historical migration/backfill.
