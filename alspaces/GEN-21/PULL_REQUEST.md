# chore(GEN-21): migrate all existing telemetry ledgers to schema 6

Migrate all 119 existing ledgers to the current schema: update 104 schema-4 and 14 schema-5 files and leave the existing schema-6 ledger byte-identical. Includes nested iteration history. No application or configuration changes.

Verified preservation of all 1,098 sessions and 1,100 invocations, original costs, identities and ordering. The migration adds cost evidence for legacy records; all 118 migration records report completed with no pending steps. No source-backed refresh was performed. Per-ledger hashes and audit metadata are in alspaces/GEN-21/iterations/01/MIGRATION_RESULTS.json. Originals remain in Git history and private local backups.

The release decoder also drops one unsupported legacy AL-55 tool-summary member, content_errors: 1; this exception is explicitly recorded in the verification report. No other recorded fields are removed from these inputs. Shared release pin remains unchanged and needs separate setup/upgrade for ordinary schema-6 operation.

Validation: verified published rc.125 checksum/build identity; migration dry run and actual run; recursive persisted-record comparison; successful CLI scan; repeat migration changes zero files. Full make passed, including Go race tests, 824 frontend tests, static/distribution checks and builds.

Final GEN-21 telemetry could not be generated because the local capture registry lacks the worktree mapping table; the explicit temporary --capture-dir attempt also failed. No empty ledger was fabricated or capture state reset. Migration is complete; capture repair and this operation's telemetry remain pending.
