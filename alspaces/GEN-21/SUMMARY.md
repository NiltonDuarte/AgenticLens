# GEN-21: Update all existing ledgers to schema 6

Current scope includes Codex backfill (iteration 02), console-only completed migration audits (iteration 03), and verified manual capture recovery (iteration 04). GEN-21 telemetry now succeeds locally; earlier blocked-status records and manifests remain historical evidence.

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

Final iteration 02 telemetry retry after backfill verification used explicit Codex root 01a1a714-839f-76e2-b7c2-3ec19975a921 and a fresh literal nonce; exit 1 with the same "unsupported capture coordinate layout: worktree mapping table is missing" error. No GEN-21 ledger was produced and capture state was preserved. Backfill commit 9642165 and its persisted hashes passed final checks; integration with unchanged origin/main was already up to date. Updated PR #110 carries this source backfill; no new application verification is required for the final documentation-only telemetry note.


## Iteration 03: Completed migration audits move to console

User steering: "can you remove it? I prefer it to be printed out in the console, so user can revert the changes if they want to". This directly authorizes completed/no-pending metadata removal, console audit and rollback guidance; no repeated approval was required. Initial worktree clean at 8b47f0a. Planner recorded the changed boundaries in CONTRACT.md and iterations/03/PLAN.md; prior completed-audit retention is superseded for this iteration.

Implementation: explicit migrate/backfill retain completed steps and loss accounting in command reports/console while omitting completed migration objects from persisted output. Incomplete state remains resumable. Current-schema cleanup validates original duplicate-aware JSON, supported raw migration status and required fact fields before any normalization. Cleanup-only backfill preserves all unrelated raw values, including code, anomalies, timestamps and owner order. Actual source refresh still follows the established replacement rules. Console output identifies ledger paths and actual optional .bak paths; dry runs make no changes/backups. Audit text alone cannot reconstruct retired payloads, so rollback uses retained originals/Git. Canonical schema/operator guidance is updated. Supported completed metadata with null/omitted step arrays stays compatible; no new metadata shape policy was introduced.

Independent review covered behavior/integration, data/failure, architecture and test effectiveness, each on both configured profiles gpt-6.1-sol/high and gpt-6-luna/medium, in capacity-bounded waves. Execution settings were accepted by native dispatch. Final inspection of native session_meta parent relations and every observed turn_context independently confirmed the exact model/effort for the parent and all 11 child sessions; no usage/cost extraction is claimed. All eight revision-2 cells approved the frozen source af1d89e8fc585410841313a268c6694599d47ce3cbb52c3bbd8a783992c81765. Approvals carry to final source fingerprint f41f7b15cec64787f36be47b98222394ab53392ba3cb4fa3ab14649185591ea8 solely for deletion of one unreferenced private validateCurrentLedger helper. Parent verified it had no callers and compared the full patches; no reachable behavior or test changed in that final deletion.

Concrete findings/fixes: data reviewer gpt-6.1-sol/high found raw missing/unknown status could be normalized and deleted; architecture reviewer gpt-6.1-sol/high found duplicate keys could collapse before validation; architecture reviewer gpt-6-luna/medium found cleanup-only backfill changed unrelated values; test reviewer gpt-6-luna/medium found backfill tests lacked populated fact preservation; test reviewer gpt-6.1-sol/high found pre-existing retired events in identity fixtures masked their named guards. Coder gpt-6.1-sol/high corrected production; tester gpt-6.1-sol/high added populated raw equality and negative/compatibility coverage and corrected obsolete fixtures. The architecture gpt-6.1-sol/high recheck also identified the unused helper, now removed. A data gpt-6-luna/medium claim that unknown nested migration fields disappear in explicit migrate was refuted by parent CLI reproduction and the existing strict engine guard; the reviewer acknowledged the refutation. Initial/final findings and dispositions are retained in [compact review summary](iterations/03/REVIEW_SUMMARY.json). Six confirmed production/test/maintenance findings were fixed (including pre-existing fixture debt); one claimed defect was refuted. No finding or material evidence gap remains open.

Verification so far: domain migration tests and focused runner/backfill/real-CLI race tests passed; tester-revision2.md records exact commands/results. Parent broadened to go test -race ./internal/adapters/backfill: PASS, 12.265s. Independent data gpt-6.1-sol/high recheck ran 16 isolated CLI probes including original malformed-status cases, duplicates, missing required facts, incomplete preservation, compatible null arrays and mixed-harness raw equality. Behavior gpt-6.1-sol/high also verified explicit recompute is not swallowed by cleanup. Product tests use disposable fixtures, never live ticket data. Real-corpus dry-run with the changed source CLI: 119 scanned, 118 cleanup candidates, one already clean, zero errors; no writes/backups. Originals saved privately in /tmp/agenticlens-GEN-21-review03/pre-cleanup/.

Integrated make, actual cleanup, final consolidated approval and telemetry facts are recorded below after completion.


### Final integrated evidence and approval

Implementation commit 42d82e6 matches reviewed final source fingerprint f41f7b15cec64787f36be47b98222394ab53392ba3cb4fa3ab14649185591ea8. Fetched origin/main remained 86bea59; merge was already up to date. Required full make passed with exit 0 after commit/integration: static checks, distribution/installation tests, all Go race suites, all 824 frontend tests, canonical frontend/embedded assets and native build. The existing Vite size advisory is the only build warning. Working tree stayed clean before the live ledger operation. Source is built locally for this iteration; no release installation or shared pin update is claimed.

Ran the built native from 42d82e6: migrate --workspaces alspaces --target-version 6 --backup. Result: 119 scanned; 118 completed migration objects removed; one already absent and byte-identical; zero errors. Console printed each saved audit, completed steps, exact ledger path and actual backup path. JSON comparison proves every other root/session/invocation value unchanged: all 1,098 sessions and 1,100 invocations, costs, skills/delegations, owner order, anomaly/code summaries and timestamps preserved. All 118 .bak files matched original bytes. Repeated actual migration with --backup reported 119 current, zero migrated/errors, with all ledger and original-backup hashes unchanged. CLI --list accepted all 119 metadata-free schema-6 ledgers.

[Cleanup manifest](iterations/03/CLEANUP_RESULTS.json) records before/after hashes and current backup locations. Verified .bak files were moved outside Git to /tmp/agenticlens-GEN-21-review03/cli-backups/alspaces/; the command console subsequently reported that relocated backup root. Temporary paths are machine-local and not durable. Durable rollback is the committed original ledger at revision 42d82e6; restore the desired ledger paths from that revision to recover their completed metadata. No raw transcripts or backup ledgers are committed.

Parent consolidated verdict: APPROVED for this iteration after all eight revision-2 approvals, verified unreachable-helper carry-forward, full integrated make and actual metadata-only cleanup verification. Discoverers and model/effort are recorded above and in REVIEW_SUMMARY.json; verification actors were the tester, independent focused reviewers and parent, distinct from discovery credit. Resolved detailed review reports and run/finding working records were removed after compaction. No unresolved/accepted debt was hidden by cleanup. Hosting checks and release deployment have not been exercised by local approval; no merge is authorized by this review.

Final telemetry attempt and PR delivery facts follow below.


Final iteration-03 Codex telemetry generation was delayed until after reviewed integration, full make and real cleanup verification. Explicit native root 01a1a714-839f-76e2-b7c2-3ec19975a921, fresh literal nonce and conventional type chore were supplied to the built 42d82e6 native. Exit 1: "unsupported capture coordinate layout: worktree mapping table is missing; preserve this state and reset or reinstall capture into a new state directory". No GEN-21 ledger was created or empty metrics fabricated; capture state was preserved. Native session_meta confirms this root and all 11 expected child sessions, and turn_context confirms requested model/effort for each. [Telemetry status](iterations/03/TELEMETRY_STATUS.json) retains source-proven identities and the exact extraction blocker. Source availability is distinct from successful extraction; capture repair/backfill for GEN-21 remains pending on this machine. All 119 historical cleaned ledgers still match their final hashes.

Delivery uses the existing open PR #110: https://github.com/NiltonDuarte/AgenticLens/pull/110. The earlier migration/backfill commits and this reviewed implementation/cleanup are included; no merge or release deployment is performed.


## Iteration 04: Capture recovery dependency and local completion

User explicitly authorized controlling the existing fix thread, merging an eligible non-release PR and/or integrating its changes into this branch, and continuing until telemetry works here. Initial Git status clean at 132b3c7. Read the owning thread, which allocated FIX-17 and was planning explicit capture-directory recovery. Sent bounded implementation approval and runtime acceptance; no duplicate ticket/implementation owner was created.

Owner thread: <t3-thread://v2/ccbaed6c-5335-443d-9aaa-69948fbd15d1/956094cc-6f16-46af-b1d3-8b38cf841b87>. Independent owner keeps application changes and its telemetry; GEN-21 owns integration and verification in this checkout. Native source session identities are distinct from T3 thread identifiers. Coordination remains part of active GEN-21, not a switch into FIX-17.

Initial coordination state: FIX-17 implementation/review was authorized and running. Integration and successful local generation remain pending; the previously reported capture failure is not yet resolved. The source registry and all 119 historical ledgers are preserved. Host completion notification will resume coordination after the owner's run settles; a plan or owner report alone is not a success claim.


### Reviewed dependency integration and verification

Received and verified FIX-17 reviewed implementation d4a9e6edb7b36247a5c35db15c2e0e64a050d775 against its exact diff fingerprint 6cc147ec3e399a59e5589d7a19df5882cd98571b0e9db848c1cdc73772aeb439. Four capture/CLI/test paths changed. Merged without conflicts as 12048ab58a5066879762f4a321e0df4667d90ebb, preserving GEN-21 console audit and raw-ledger cleanup behavior. Final FIX-17 head 5c9d1458dac66e926b4c386318df0eca3ca9a22d adds only seven files under alspaces/FIX-17; verified zero backend/web/scripts difference from the implementation and merged it as 72e66739a9af4d5e0018c6e0b4987651b26d3f6d. The source-built native from 12048ab remains source-equivalent after this artifact-only merge.

FIX-17 [PR #111](https://github.com/NiltonDuarte/AgenticLens/pull/111) is a non-release dependency. Its eight focused persona/model cells approved; [retained findings](../FIX-17/iterations/01/REVIEW_SUMMARY.json) preserve discovery credit: behavior reviewer gpt-6.1-sol/high found lost configured-manifest validation; architecture reviewer gpt-6.1-sol/high and data reviewer gpt-6-luna/medium jointly found the explicit coordination-path manifest case. Both severity-6 issues were fixed and reapproved. Independent owner source telemetry succeeded for its own root and eleven children, with the old registry unchanged. No FIX-17 sessions are treated as native children of GEN-21.

GEN-21 combined verification: first make failed only the unchanged TestCoordinationDirectoryNamesConcurrentFirstOpenAndStableReopen at shared_state_test.go:353 with SQLITE_BUSY (1.95s). The test, project.go and registry.go are unchanged from 86bea59. Five isolated race-enabled repetitions passed (4.501s). The required full make recheck then passed with exit 0: all Go race packages, all 824 frontend tests, static/distribution/installation checks, synchronized assets and native builds. Both logs are retained privately; the initial failure is not hidden or claimed fixed. Concurrent-initialization contention remains an observed pre-existing test limitation, not a demonstrated regression in coordinate recovery. No production code or test was changed to obtain the recheck.

Parent integration review confirms only the reviewed capture changes plus the earlier GEN-21 implementation; clean automatic merges made no manual product edits. Tests cover invalid/healthy/fresh selected state and preserved configured/selected/discoverable installation checks; native/default/automatic paths remain strict. All 119 historical cleaned ledgers still match their recorded hashes. Old capture database/WAL hashes were saved immediately before final local generation; a separate fresh recovery directory will be selected explicitly without resetting or changing shared config. Source metadata still proves exactly this root and its eleven native children.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 11
- **Active Agent Duration:** 2h 9m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 276.4k tokens (61% of 450k healthy ceiling, 0 compactions) | Subagents peak 191.9k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 3 turns) | Planner 1.0, Coder 3.0, Tester 2.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 1.9:1 (+540 test lines / +290 prod lines)
- **Code Delivered:** +830 / -138 lines across 11 code files (excluding docs)
- **Reasoning Effort:** 106,059 reasoning tokens (30.9% of generated output)
- **Tool Health & Self-Correction:** 498 tool calls (35 self-corrections, 93.0% success rate)

### Anomalies & Health Flags
- **Status:** 9 execution anomalies detected (5 CRITICAL, 4 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (5 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (13 cycles) — Subagent roles 'coder' and 'tester' alternated for 13 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-architecture (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-architecture' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+109.6k tokens) — Context tokens surged by 109.6k to 109.6k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+115.7k tokens) — Context tokens surged by 115.7k to 115.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.4k tokens) — Context tokens reached 250.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (276.4k tokens) — Context tokens reached 276.4k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 121,105 | 4,205 | 1,127,552 | 90.3% | 123.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 82,110 | 22,215 | 7,900,544 | 99.0% | 184.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 86,059 | 26,839 | 10,483,968 | 99.2% | 191.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 202,339 | 121,230 | 37,960,064 | 99.5% | 276.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 251,135 | 16,065 | 2,778,880 | 91.7% | 121.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 271,835 | 14,715 | 2,171,904 | 88.9% | 99.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 182,672 | 17,283 | 2,899,328 | 94.1% | 102.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 148,667 | 14,667 | 2,424,064 | 94.2% | 102.1k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **12** | **1,345,922** | **237,219** | **67,746,304** | **98.1%** | **276.4k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 411,148 | 23,391 | 3,203,968 | 88.6% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 934,774 | 213,828 | 64,542,336 | 98.6% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **1,345,922** | **237,219** | **67,746,304** | **98.1%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Capture observation finalized: 2026-11-04T18:12:43.008730443Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.


### Verified local recovery and final telemetry

Manual generation passed with exit 0 using the integrated source-built native from 12048ab, explicit Codex source/root 01a1a714-839f-76e2-b7c2-3ec19975a921, literal nonce 8f06d3a1c79e542b and fresh capture directory /home/niltonduarte/.agenticlens/recovery/GEN-21-8f06d3a1c79e542b. An explicit title and conventional type chore were supplied. The command printed the known configured-coordinate warning, then discovered twelve rollouts and wrote the canonical GEN-21 ledger and rendered this summary's metrics table. Native wrapper-only --al-project/--al-offline flags were not passed to the direct executable.

Persisted verification: schema 6, ticket GEN-21/chore, exactly twelve unique Codex sessions equal to this root and all eleven expected source-proven children, twelve source invocations, and no completed migration object. The independent FIX-17 root is not included. Configured registry database and WAL SHA256 hashes are identical before/after; all 119 historical cleanup hashes remain identical. The new FIX-17 ledger arrived through its owner's Git commits; the new GEN-21 ledger was extracted only by this root. [Current telemetry status](iterations/04/TELEMETRY_STATUS.json) records identities, directory, hashes, source/integration revisions, successful generation and remaining limits. Earlier iteration-03 blocked status remains an honest record of the original failure, not the current result.

Source evidence preserves unavailable cost_evidence and the root's empty role value (Unknown). No successful cost reconstruction, callback binding or automatic capture health is claimed. The old configured registry remains incompatible, and the shared release pin still lacks this fix; recovery here uses the explicit healthy directory with the reviewed source-built native. The source-backed snapshot necessarily excludes subsequent verification and publication calls. No raw source transcripts or capture databases are added to Git.

Integration acceptance is satisfied: reviewed fix imported without product edits, full combined make passed after a recorded unchanged-test contention failure, actual GEN-21 generation succeeded with exact source identities, old registry/history stayed unchanged. No release or upstream PR merge is implied by local branch integration. PR #111 hosting checks were still running at the last readiness inspection; its source was integrated into this branch using the user's authorized branch-integration alternative. Final GEN-21 delivery remains PR #110, updated with the resolved capture limitation.
