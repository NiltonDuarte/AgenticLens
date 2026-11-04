# GEN-21: Update all existing ledgers to schema 6

User directly authorized the operator migration on 2026-11-04. Initial Git status was clean at 86bea59 on generated placeholder branch GEN-<NUMBER>/update-ledgers-to-v6. Checked existing workspaces and ticket index; allocated unused GEN-21 and renamed the branch. No matching planning specification or ticket index entry existed.

Migrated 118 of 119 existing ledgers (104 schema 4; 14 schema 5). The existing schema-6 AL-62 ledger remained byte-identical. Recursive scope includes AL-68/iterations/05. All 119 now use schema 6. All 1,098 recorded sessions and 1,100 invocations retain their identities, order, counts and recorded fields, including original costs; schema 4 gained cost_evidence on 692 invocations. No source backfill was run and no new source completeness is claimed. All 118 migration statuses report completed with no pending steps. All schema-6 occurrence retirement audit counts are zero for these inputs.

Compatibility detail: the release decoder also removed the noncanonical content_errors: 1 field from AL-55's command tool summary. Standard calls/errors and all invocation facts remain preserved. This removal is absent from the CLI occurrence audit; it is recorded explicitly in MIGRATION_RESULTS.json. The original remains recoverable through Git and the private pre-migration backup. No other session/tool/root changes occurred beyond schema/migration metadata and cost_evidence additions.

## Operation and verification

Installed common wrapper and native cache were missing. Shared pin v2.0.0-rc.111+99fbcd7 has no available GitHub release. Used temporary published native v3.0.0-rc.125+86bea59 matching the checkout, verified against release SHA256SUMS (native SHA256 5c6602796e4e20e0603334dc699ac0558e4ce6fbdc4f9d7ff11c61726f91f804). Shared pin, configuration and hooks are unchanged. Future operation through the old pin still requires separate setup/upgrade; migrated files require schema-6-capable readers.

- migrate --workspaces alspaces --target-version 6 --dry-run: 119 scanned, 118 planned, one current, zero skipped/errors.
- migrate --workspaces alspaces --target-version 6 --backup: 118 migrated, zero errors.
- Recursive JSON comparison against originals: all session IDs, invocation counts and existing financial evidence preserved. Existing schema-6 bytes unchanged. Recorded field-level exception above.
- CLI --list --workspaces alspaces: successful ingestion of migrated history.
- Repeated actual migrate: 119 current, zero migrated/skipped/errors; subsequent digest comparison confirms unchanged output.
- [Per-ledger evidence](iterations/01/MIGRATION_RESULTS.json) retains original/result hashes and migration metadata.

Private backup and command logs: /tmp/agenticlens-GEN-21.MNbqCGGx/ (machine-local, not portable or durable). CLI .bak files were checked byte-for-byte against originals and moved there. No raw transcripts or backup ledgers added to Git; committed baseline is the durable rollback source.

Required repository make check and delivery: pending below.
