# GEN-21: Update all existing ledgers to schema 6

Current scope includes the user-authorized Codex backfill in iteration 02, documented below. Iteration 01 migration hashes and unchanged-ledger assertions describe the earlier migration checkpoint.

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

## Final verification and delivery

Reviewed and tested migration commit: 61eb84a. origin/main was already identical to the initial baseline; git merge origin/main --no-edit reported already up to date after the implementation commit. Full make passed (exit 0): Go vet/race tests, release/installation/capture distribution checks, frontend checks and 824 tests across 36 files, frontend asset build and native build. Only the existing Vite bundle-size advisory was reported. Working tree remained clean after make. No application changes or additional tests were needed for this operator operation.

Final session telemetry uses Codex native root 01a1a714-839f-76e2-b7c2-3ec19975a921, inspected directly in its source session_meta record (originator T3 Code, source vscode, exact worktree). No child agents were launched. Source session data is local to this machine; generation does not establish other developers' source availability. Native invocation uses the verified temporary release because the common wrapper is absent and the old configured release is unavailable.

PR publication follows the final telemetry commit. All task-owned paths are committed; historical originals remain available from parent 86bea59.


## Final telemetry limitation

Final Codex generation was attempted after verification with explicit source, exact source-verified root ID, literal unique nonces and conventional type chore. Both the configured capture path and a separate explicit temporary --capture-dir failed with: "unsupported capture coordinate layout: worktree mapping table is missing; preserve this state and reset or reinstall capture into a new state directory". No GEN-21 agentic_metrics.json was produced. Existing native Codex session data is present; the installed release's capture admission/installation checks block extraction before a ledger can be published. Matching-revision inspection of backend/cmd/agentic-lens/main.go and capture.go confirms configured-state inspection remains part of this path. No source completeness, empty session ledger or successful telemetry extraction is claimed. No children were launched; historical migrated ledger hashes remained unchanged. Preserve the existing registry; capture repair and subsequent GEN-21 telemetry backfill remain pending on this machine. This does not affect the 119 historical schema migrations.


## Iteration 02: Codex source backfill

Steering: user requested "backfill codex ledgers" after the migration handoff. Direct authorization extends GEN-21 to source refresh; see [iteration plan](iterations/02/PLAN.md) and amended CONTRACT.md. Initial worktree was clean at 8d57793 and existing PR #110 remained open.

Used the same checksum-verified published native rc.125 and this machine's /home/niltonduarte/.codex. Executed backfill --source codex --codex-home /home/niltonduarte/.codex --workspaces alspaces --recompute-db, first with --dry-run and then with --update-summaries. Exact recorded owners only; no --force or heuristic fallbacks. Dry run and actual operation exited 0. Both report 119 scanned/completed, zero partial, and 68 no-selected-source-owner diagnostics for ledgers without Codex sessions. Those diagnostics do not represent failed Codex refreshes.

Refreshed all 585 existing Codex sessions across 51 ledgers (including AL-68 iteration 05); 292 snapshots changed and 293 were already equivalent. All original session IDs, ordering and invocation counts remain: 1,098 sessions and 1,100 invocations across the complete ledger inventory. Within the 51 selected ledgers, source-backed occurrences now contain 73 skills and 395 delegations. These replace the earlier stored snapshots; source-derived user, invocation, tool and anomaly facts may also change. No new sessions were added. All 513 non-Codex session snapshots compare exactly to their pre-backfill values, including those sharing Codex ledgers. The other 68 ledgers and existing summaries were restored byte-for-byte after the batch CLI rewrote incidental metadata outside the selected source scope. CLI-created nested AL-68 iteration 05 SUMMARY.md is retained with its refreshed ledger.

[Per-ledger backfill results](iterations/02/BACKFILL_RESULTS.json) contain before/after hashes, owner counts and changed snapshot counts. Pre-backfill originals are also preserved privately at /tmp/agenticlens-GEN-21.MNbqCGGx/pre-backfill/; the committed pre-backfill checkpoint is 8d57793. Migration checkpoint hashes are historical; the iteration 02 hashes describe the current source-refreshed files.

Verification: exact original owner set retained; every non-Codex snapshot unchanged; all 119 files remain schema 6 with zero pending migration steps; CLI --list successfully ingests refreshed history; git diff --check passes. Full make from iteration 01 remains valid because no application, tests, runtime scripts or config changed. Fetched origin/main is unchanged at 86bea59, with no new code integration to test. Source evidence covers only this machine and recorded owners, not other unrecorded sessions or team-wide capture completeness. The prior capture-registry incompatibility remains unresolved; final GEN-21 telemetry retry is recorded below.
