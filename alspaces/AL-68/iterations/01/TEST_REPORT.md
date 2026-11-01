# AL-68 iteration 01 test report

Status: affected Go regression checks pass. Real harness capture/agent-interaction evidence and integrated delivery checks remain incomplete in this tester run. This is test evidence, not the consolidated review verdict.

## Revision and execution

Baseline HEAD: 7bc670a7f5c9e93a220f3596f3d0e05cb6007332 with uncommitted AL-68 product/tests. The coder paused the affected Go paths during the full focused run. Identical before/after SHA256: `0071b2308f100a882dcc7787f1f6151a2a40363854e8b26c4d30f12adcd0a78c`.

Fingerprint definition: sort every recursive regular file under `backend/internal/domain`, `backend/internal/adapters/capture`, `backend/internal/adapters/codex`, `backend/internal/adapters/opencode`, `backend/internal/adapters/backfill` and `backend/cmd/agentic-lens`; hash repository-relative path bytes, NUL, file bytes, NUL in order. This includes tests and the embedded plugin template. Changes after that fingerprint need impact-selected rechecks.

The embedded asset prerequisite had been prepared by the coder; this tester verified `backend/internal/adapters/web/dist/index.html` exists before direct Go checks, and the CLI integration tests built and executed their native test binary successfully.

From `backend/`, full affected-package execution:

`go test -race ./internal/domain ./internal/adapters/capture ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill ./cmd/agentic-lens -count=1`

| Package | Result | Reported package duration |
| --- | --- | --- |
| internal/domain | PASS | 1.183s |
| internal/adapters/capture | PASS | 2.200s |
| internal/adapters/codex | PASS | 1.101s |
| internal/adapters/opencode | PASS | 2.085s |
| internal/adapters/backfill | PASS | 6.187s |
| cmd/agentic-lens | PASS | 5.650s |

Afterward, one intentional request-wire regression was added to protect the OpenCode prompt defect below. Its test file SHA256 is `fed83872934545bb299c7f8e51e50796427906bfb8727276c7dd667200e7b9a2`. Command `go test -race ./cmd/agentic-lens -run '^TestCaptureCLIRequestExposesUnassignedTicketToPlugin$' -count=1` passed in 1.756s. Completion recorded at 2026-11-01 18:11:42 UTC. No product changes were made by the tester. `git diff --check` also passed.

## Added regression ownership and oracles

- `backend/internal/domain/capture_test.go`: explicit/inherited binding, rejected conflicting ownership and unsafe ticket values; cumulative snapshot truncation/time precedence including fractional and offset-equivalent RFC3339 instants. Domain tests use pure values with no mocks or I/O.
- `backend/internal/adapters/capture/hooks_test.go`: production hook input parsing and model-visible binding request, exact durable request challenge, explicit response acceptance, no ticket/parent assignment from unknown payload fields, subagent parent ID not treated as child, malformed/oversized payload and unsupported event rejection. These are synthetic callback fixtures, not actual harness invocation evidence.
- `store_test.go`: exact source/store/session challenge matching, missing/wrong/other-session tokens, conflicting tickets, unchanged binding on resume, simultaneous SQLite writers, aborted transaction rollback and reopen. A real subprocess exits while holding a write transaction; reopening and binding with the original committed request proves the crash did not strand the lock or publish the uncommitted mutation.
- `reconcile_test.go`: source-derived usage/cost evidence, unchanged and resumed snapshot idempotence, truncation and missing-source preservation, source-proven child registration/inheritance, independent ticket isolation and unassigned visibility, shared byte/session budget continuation, canonical ID collisions across distinct stores, partial-source preservation and publication redaction without mutating local state.
- `opencode_test.go`: temporary SQLite fixtures for both schema families contain an unrelated matching ticket title while the registered exact ID is missing. Reconciliation must not select that title as a fallback or certify an empty/current capture.
- `backend/cmd/agentic-lens/capture_test.go`: native CLI installer dry run, existing user hooks, repeated installation/disable/removal, edited owned-hook/plugin conflict preservation, missing handler rejection, explicit empty `ticket_id` request wire for the plugin, receive/request/bind/generate integration across two independent roots, another ticket and an unassigned root. Published canonical v5 usage remains idempotent and survives a missing registered source. The actual final report digest matches the ledger; final `capture_report.json` and appended `SUMMARY.md` exclude binding request tokens, local store/state paths and transcript/snapshot fields.

No existing tests were removed. Tests use temporary repositories, SQLite files and synthetic source records; none depend on live ticket workspaces, developer source sessions or credentials. Existing source extraction/preservation tests were inspected before adding coverage, as mapped in TEST_PLAN.md.

## Defects discovered and verification

| Tester finding | Before-fix evidence | Correction and regression evidence |
| --- | --- | --- |
| Unbound OpenCode registration did not trigger agent feedback | Source trace: plugin required present empty `ticket_id`, but registration omitted it through `omitempty` | Coder serializes explicit empty ticket field. Native request-wire test passes. Real plugin-to-agent delivery still requires smoke evidence |
| RFC3339 lexical comparison rejected later fractional/equivalent-offset source instants | Domain precedence table failed fractional and offset cases | Coder uses parsed instants. Full domain race suite passes |
| Two-pass reconciliation charged source bytes twice and could starve later registered owners | Exact one-source byte-budget fixture ended in failed state after successfully reading the source | Coder separates source reads from parent binding propagation and persists continuation. Exact budget and successive bounded-pass regression passes |
| Same harness session ID from different stores could enter an indistinguishable v5 result | Distinct stores with same exact session ID produced two result rows without an error | Coder rejects canonical identity collisions. Regression passes |
| Warning-bearing source refresh replaced previously accepted facts with partial data | Larger fixture with valid reduced usage and a malformed tail replaced 10 accepted input tokens with 1 and reported current | Coder rejects warning-bearing parse before replacement. Previous facts and incomplete status regression pass |

These findings were reported directly to the product owner/coder; the tester did not implement their corrections. They are separate from independent focused-review attribution.

## Evidence limits and remaining gates

- This tester did not run real Codex/OpenCode sessions. The fixture receive/bind exchange proves the implemented command/feedback contract, not that a harness delivers context or that a real agent responds. The user's required actual hook-to-agent interaction and independent reviewer/native child capture remain real smoke gates. Codex trust review was reported unavailable by the parent at test assignment; no trust was changed here.
- Source-store availability, version observations and a successfully installed configuration are not capture completeness evidence. No T3 path, additional version/schema family, macOS or Windows runtime compatibility is claimed by these Linux fixture checks.
- Installer tests cover conflict and preflight-failure preservation plus successful removal. A deterministic failure after one installer write and before the remaining writes was not injected in this tester run; installer transaction rollback still needs independent inspection or an appropriate fault-path probe.
- The full frontend suite, downloader/distribution script suite, integrated `make`, upstream merge and final standalone distribution checks were not run by this tester. The parent/coder own those coordinated stages and must record their actual results separately. Passing these focused tests cannot replace them.
- Parent-provenance callbacks with unavailable/mismatched child source data and other unregistered sessions must remain explicitly incomplete; the focused tests cannot establish full coverage of sessions that never delivered a hook.

No required evidence is silently skipped or fabricated. Current passing fixture evidence does not satisfy the remaining real-runtime or consolidated delivery gates.

## Real-smoke correction: coexisting OpenCode schema families

At the parent's request, tested the capture-only correction after the coder found both complete schema families in the installed OpenCode source store. This is a correction within the existing exact-source contract, not permission for heuristic family selection. The general resolver's mixed-layout rejection remains unchanged.

New fixture coverage in `backend/internal/adapters/capture/opencode_test.go` uses one temporary real SQLite database with all five required v1/v2 tables and unrelated sessions carrying the same ticket title. Through `Store.Reconcile`, an exact registered ID present only in v1 must produce v1's model and 11 input tokens; one present only in v2 must produce v2's model and 23 input tokens. The same exact ID in both families must fail rather than pick a family, and a missing ID must remain unresolved without matching unrelated titles. After accepting a valid snapshot, both a newly duplicated ID and removal of the exact owner must preserve the prior facts while ceasing to claim a current snapshot.

`go test -race ./internal/adapters/capture -run '^TestOpenCodeCapture' -count=1` passed (1.831s).

Then `go test -race ./internal/adapters/opencode ./internal/adapters/capture -count=1` passed: opencode 1.912s, capture 2.711s. Identical before/after SHA256 over the two complete adapter directories, using the path/NUL/content/NUL definition above: `9c180210452f962cf12b7c27c60d0fc758d5708fd019048e77d26e42e9715552`. Completion recorded at 2026-11-01 18:16:46 UTC. `git diff --check` passed. The tester changed only its assigned test file and this report. No duplicate full suite was run.

These fixture checks establish the exact-membership dispatch and preservation behavior. They do not establish that the real agent smoke has now completed source reconciliation; that evidence remains with the coder/parent.

## Review corrections: corrected candidate verification

Completed 2026-11-01 19:00:00 UTC. Product and test edits were frozen by agreement with the primary coder. The same six-directory fingerprint definition used earlier produced identical before/after SHA256 `5d2e00d47ce4a17e5ee4c1da1183c05851ebcd528ce6d13cee553ffb92f00d4e`. This supersedes earlier fingerprints only for the corrected source/test revision; earlier failed/passed results remain evidence of their own revisions.

### Correction oracles and before/after evidence

| Finding | Durable production-path oracle | Before-fix evidence | Corrected result |
| --- | --- | --- | --- |
| CAP-01 | `TestCaptureGenerationRetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot`: registry accepted 10, a different supported generation publishes 30, registry source disappears; 30 and stored other-source facts/repeated skill/task rows survive. Separate case archives the registered path and successfully extracts 30 explicitly while the retained registry still has 10 | Native before-fix reproductions in behavior-sol, data-sol, architecture-sol and tests-sol reports. The tester's first new CLI run hit a transient partial product API save; it does not count as a before-fix behavior reproduction. A subsequent tester fixture initially failed canonical validation for its missing stored user; corrected to reach the intended merge branch | PASS, both cases; valid stored-user and AL-62 multiplicity fixtures |
| CAP-02 | `TestCaptureGenerationOverlappingPublishersCannotRegressFinalSnapshot`: real Git forwarding shim pauses older generation after its source phase, advances the source from 10 to 50, starts a newer native publisher, then releases the old process. Final ledger must be 50. Covers a newer registered publisher and a newer explicit publisher with an absent registry | Deterministic native before-fix schedule in data-sol report. The tester reused that proven trigger and did not rerun its CLI oracle against an older binary | PASS. A TCP barrier establishes the source/publication ordering; a bounded release deadline permits the corrected implementation to serialize the newer process behind the older owner. Registered publication also verifies the final report digest. The explicit generation case does not claim a new capture report was published |
| CAP-03 | `TestCaptureGenerationRejectsExplicitSelectionOfAnotherTicketOwner`: exact root bound to TEST-100 cannot publish to TEST-200 through explicit selection; no conflicting ledger is written | Native before-fix reproduction in data-sol report; no separate tester pre-fix native run | PASS |
| CAP-04 | `TestReconcileRecoversSourceProvenChildAfterStartWithoutStop`: SubagentStart retains opaque interface agent evidence, then source metadata supplies a different exact child ID and proven root parent. No Stop event is delivered | Tester regression failed before coherent correction: the source child was absent and pending start evidence was lost | PASS; the opaque agent ID never appears as an invented source session |
| CAP-05 | `TestReconcileFailedReadsConsumeSharedByteBudget`: two malformed bound sources each fit the entire allowance. Only one is parsed/failed; the other stays pending, and reserved-byte/attempt counters remain within the shared bound | Tester regression failed before correction with two failed sources under one-source byte allowance | PASS |
| CAP-06 | `TestCaptureInstallerRendersMarkerBearingPathsOnceAndRollsBackMidWrite/literal_marker_paths`: actual native install receives handler/state/store paths containing mutually referring template markers and a quote; the three emitted constant values must parse and recover the exact original paths | Security-sol's repeated native syntax probe demonstrated the before-fix defect. The durable fixture uses cyclic markers so a repeated-substitution implementation cannot pass merely because of one map iteration order; no repeated trial quota is needed | PASS |
| CAP-07 | `TestCaptureReceiveNativeTimeoutBoundsOpenStdin`: valid JSON is written to a native receive process while stdin remains open; the native 50ms timeout must return nonzero before an independent three-second kill deadline, without registering an event | Security-sol's native before-fix open-pipe probe remained running beyond its requested timeout. No separate tester pre-fix native run | PASS |
| CAP-10 | `TestReconcileReportsCompletionAfterOwnerCheckpoint`: a successful source read exposes `refreshed_at` between actual call start and completion cutoff; declared pass start and cutoff semantics are present and ordered | Tester regression failed before correction with no owner read checkpoint; data-luna additionally traced the earlier misleading start-time cutoff | PASS |
| CAP-11 | `TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities`: four unrelated unassigned registrations sort before a bound owner; a one-owner/one-source budget still refreshes that owner. Ticket report excludes unrelated IDs while local List retains all five registrations | Tester regression failed before correction: unassigned roots consumed work, bound owner stayed pending and unrelated IDs appeared in the result | PASS |

The existing child/isolation test was intentionally changed: it now rejects unassigned identities in ticket results and separately asserts their continued presence without a guessed ticket in local diagnostics. The direct redaction fixture now sets its explicit TEST-100 binding, so it still exercises an owned published row after ticket scoping. These changes implement the reviewed privacy/ownership contract rather than deleting assertions.

The actual installer mid-write rollback probe is now a durable native CLI subcase of `TestCaptureInstallerRendersMarkerBearingPathsOnceAndRollsBackMidWrite`. A state path nested under `.codex/hooks.json/state` permits planning but fails manifest apply after hooks/ignore writes. The new hooks file must be absent and original `.gitignore` byte-identical after failure. This passes, closing the specific unexecuted fault-injection limitation recorded earlier. It does not claim crash-atomic multi-file installation or concurrent-edit proof.

### Coordinated affected checks

The tester ran, from `backend/`:

`go test -race ./internal/domain ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill -count=1`

All passed: domain 1.118s, codex 1.082s, opencode 1.899s, backfill 5.710s. Before/after corrected fingerprint matched as recorded above. `git diff --check` passed.

To avoid duplicating the concurrent expensive run, reused the primary coder's exact command/result on the acknowledged frozen product/test files:

`go test -race ./internal/adapters/capture ./cmd/agentic-lens`

The coder reported exit zero: capture 3.379s, CLI 13.178s. This includes the new corrected regression tests and existing package suites. The coder did not collect its own before/after hash; the tester recorded the matching six-directory hash during this coordinated freeze. The coder separately reported successful Windows/Darwin cross-builds and Linux native build; those are reported build evidence, not this tester's own native platform execution.

An initial targeted correction run while product files were being saved failed CLI compilation for the not-yet-added RefreshedSessions/Skills/TaskLinks fields and reproduced CAP-04/05/10/11 in the adapter package. After the coherent save, affected correction cases passed except for the tester's missing-user fixture validation error, which was fixed without changing product code and rerun successfully. These setup/transient failures are distinguished from product regressions; no successful extraction was fabricated.

CAP-08/09 distribution/update/manual corrections and their tests belong to the setup coder and are not certified by these Go checks. Full integrated `make`, independent corrected-candidate re-review and final delivery remain parent-owned stages. Real Codex trust/callback/agent interaction still requires the normal human trust action; all synthetic callback and concurrent-publication fixtures above remain fixture evidence, not a substitute for that runtime gate. No new real harness support claim is made here.

Tester product/test writes are now frozen. No additional exploratory tests, product edits, staging, commits, telemetry generation or publication were performed by the tester.

## Candidate-02 remaining cases and final test freeze

The parent assigned the remaining CAP-01/CAP-03 and bounded-child continuation triggers after reviewing candidate 02. The tester added only those concrete oracles, preserved the prior tests and reproduced all four before the coherent fixes:

- `TestCaptureGenerationRetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot/successful_unchanged_older_source_checkpoint`: accepted old source bytes produce 10 at source timestamp 02; a later appended record at source timestamp 03 produces cumulative 30 through explicit generation with an absent registry; restoring the exact old bytes and successfully rereading them through the original registry must retain canonical 30. The native before-fix test failed with 10. This establishes actual source chronology and does not infer ordering from totals.
- `TestCaptureGenerationResolvesStoreAliasesWithoutConflatingDistinctStores`: a real directory symlink cannot bypass TEST-100 ownership when generating TEST-200, in either registration/selection direction. Both directions failed before the fix by publishing the conflicting ledger. After normalization was implemented, the stored-alias case was strengthened with an actual pre-normalization alias row in temporary SQLite, so it still reaches legacy-alias comparison. A control then proves a genuinely separate real store reusing the same harness-local ID is allowed and stays distinct.
- `TestOpenCodeChildDiscoveryContinuesPastAlreadyKnownPrefix`: after accepting root/child-a, a later source-proven child-z is locally registered but unassigned. Repeated one-attempt passes must eventually inherit and capture child-z, while an unrelated source root remains unassigned and every pass stays within its work allowance. Before correction, 12 passes never progressed beyond the known prefix.
- `TestQueuedCodexChildRecoveryContinuesPastMalformedAndMismatchedMetadata`: two malformed stop paths and a source-parent mismatch sort before a valid child. Repeated two-attempt passes must eventually publish the valid child's exact source facts, retain all three unresolved queue entries, and never bind the different-parent session. Before correction, 20 passes never reached the valid child.

Before-fix command from `backend/`: `go test -race ./internal/adapters/capture ./cmd/agentic-lens -run 'TestOpenCodeChildDiscoveryContinues|TestQueuedCodexChildRecovery|TestCaptureGenerationRetained|TestCaptureGenerationResolvesStoreAliases' -count=1`. Both packages failed with the specific behaviors above (capture 0.870s, CLI 1.726s). This was a working-checkout reproduction during correction, rather than an asserted match to the immutable candidate-02 fingerprint. The matching frozen-candidate reviewer probes remain recorded in the assigned re-review reports.

After the first coherent save, the same cases plus the existing overlapping-publisher regression passed under the race detector (capture 1.853s, CLI 9.340s). The parent then explicitly confirmed a uniform conservative ordering rule for both capture and explicit source selection: differing facts with equal or unavailable authoritative source checkpoints preserve the accepted whole owner and expose unresolved precedence.

`TestCaptureGenerationPreservesDifferingSnapshotsWithoutProvenSourceOrder` now protects equal-checkpoint registered refresh, equal-checkpoint explicit refresh and an explicit refresh whose accepted canonical checkpoint is unavailable. Each preserves the complete baseline; registered coverage remains pending and explicit selection prints an unresolved checkpoint diagnostic. The larger differing input total is deliberately rejected in these cases, guarding against a fieldwise/token-maximum shortcut. The shared source-independent comparison is exercised through the actual native publication path.

### Chronology fixture clarification

The helper's later refresh scenarios now pass an explicit later source timestamp through `writeAt`, and the overlap schedule appends a genuine later source record. They no longer assume a different total at the same timestamp proves source order. The unchanged-old-source regression restores the original timestamp and exact bytes deliberately.

The full affected race run then passed domain/capture/Codex/OpenCode/backfill but exposed the existing `TestCLIGenerate_InvocationsUpdateInPlaceOnSameDevice` fixture's unavailable invocation completion checkpoint. Its relational row timestamps already expressed valid storage chronology; they were not fabricated. This parser obtains source invocation completion from assistant JSON `time.completed`, which those three assistant records omitted. Relational storage timestamps are not a replacement for that source-owned field under the approved ordering contract.

The parent explicitly assigned the tester the exact m2/m4/m6 fixture adjustment in `backend/cmd/agentic-lens/main_test.go`: supply JSON `time.created`/`time.completed` using the already-declared now+50/+150/+250 values. Every original update-in-place, invocation multiplicity, token, steering, compaction and SDLC assertion remains unchanged. The separate unknown-checkpoint native guard above preserves the required unavailable-evidence behavior; no source timestamp is invented by production code.

### Final coordinated results and fingerprints

`go test -race ./internal/domain ./internal/adapters/capture ./internal/adapters/codex ./internal/adapters/opencode ./internal/adapters/backfill ./cmd/agentic-lens -count=1` produced:

| Package | Result before the exact legacy fixture clarification |
| --- | --- |
| domain | PASS 1.147s |
| capture | PASS 5.255s |
| codex | PASS 1.107s |
| opencode | PASS 2.180s |
| backfill | PASS 6.459s |
| CLI | Only the existing unavailable-checkpoint fixture failed; new remaining-case and equal/unknown oracles passed |

After the three source JSON fixture changes, `go test -race ./cmd/agentic-lens -count=1` passed the entire CLI package in 13.469s. The five passing package results remain applicable: the sole subsequent executable/test change was the authorized CLI fixture, and the primary coder acknowledged that product paths stayed frozen without competing writes.

Final six-directory source/test fingerprint, matching before and after the last CLI run: `d3a5ccacfcbf838e7ba7011224b1cbbebb284084f43e6d71bab69ed5b8351a8e`. Definition is the six-directory path/NUL/content/NUL definition recorded above.

Final full candidate fingerprint, also matching before/after: `26898644f86547fb9013c161120b291c1c6ed7f130307f214351b25c1741534d`. Definition matches the parent's candidate files: sorted unique tracked and nonignored untracked regular repository files outside `alspaces`, hashing path/NUL/content/NUL. Recorded completion: 2026-11-01 19:30:15 UTC. `git diff --check` passed.

All tester files are now frozen. No unresolved reproduced regression remains in this tested correction scope. Required real Codex trust/callback/agent-response evidence remains unavailable; these fixture/native CLI checks cannot substitute for it. No full `make` was duplicated, no product files were edited by the tester, and final review/integration/publication remain parent-owned gates.

## CAP-12 remaining byte and cross-queue fairness

The parent assigned only the two remaining scheduler reproductions after candidate-03 review. CAP-01/CAP-03 and their existing source-order/alias oracles remain unchanged. Added two coherent adapter regressions in `reconcile_test.go`:

- `TestReconcileSchedulerEventuallyRefreshesOwnerWhenFailedMetadataConsumesByteAllowance`: a bound root first accepts 10. A persistent malformed queued stop is exactly 1,024 bytes. The root then appends an authoritative later record; its valid complete source is exactly 509 bytes and yields cumulative 30. Eight repeated passes retain the same two-attempt/1,024-byte allowance. The root must eventually accept 30 with a new successful-read checkpoint, every pass must respect both counters, and the failed child evidence must remain pending. The test does not increase budgets or discard failed work to obtain progress.
- `TestReconcileSchedulerEventuallyDiscoversStartOnlyChildWithPersistentFailedStop`: under an unchanged two-attempt allowance, one malformed stop and a separate source-proven start-only child must both receive fair service. Twenty repeated passes must eventually capture the exact native child/parent/usage while retaining failed-stop evidence and never substituting the opaque interface agent ID. The same bounded scenario without a failed stop is the control.

The first Go attempt encountered the coder's partially saved scheduler declarations and did not compile; this is not recorded as a behavior failure. For actual before-fix evidence, the tester independently reran the existing candidate-03 native scripts while the working code was being corrected:

- `python3 /tmp/al68-arch03-budget.py`: fixture `/tmp/al68-arch03-budget-m6g6_fps`; all eight two-attempt passes reserved 1,024 bytes, retained input 10 and the identical `refreshed_at`, despite the 509-byte source fitting the total allowance. The script's one-attempt control read the same source and accepted 30.
- `python3 /tmp/al68-data-sol-03-interactions.py`: fixture `/tmp/al68-data03-interaction-z2bl50pb`; after twenty two-attempt passes, only root was registered and two child-work entries remained pending. The tester then removed only that synthetic failed-stop queue row and repeated the unchanged bounded command: exact `child-only-start` appeared on control pass three. These disposable mutations were confined to synthetic `/tmp` fixtures.

Both replay executables, `/tmp/al68-arch03-cli` and `/tmp/al68-data-sol-03-cli`, had SHA256 `d4e305a3d62f09448f76169c7d0e27042fc44bc02499d378c40d8143dfd1a9d1`. Their candidate-03 build provenance is documented in the assigned reviewer reports; they were not silently rebuilt from the partially saved correction.

After the primary coder froze the general work-class priority correction, `go test -race ./internal/adapters/capture -run '^TestReconcileScheduler' -count=1` passed in 1.567s. Then `go test -race ./internal/domain ./internal/adapters/capture ./cmd/agentic-lens -count=1` passed all three complete packages: domain 1.172s, capture 5.140s, CLI 13.835s.

Matching before/after six-directory source/test fingerprint: `1a4a6c3ab0138374161d9ff51f6ffbc108cede7034978f034399f83432a174b1`. Matching full candidate fingerprint: `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec`. Definitions match the preceding sections. Completion recorded at 2026-11-01 19:47:18 UTC; `git diff --check` passed.

The coder confirmed that only capture `reconcile.go` and `children.go` changed in this correction; the tester changed only its assigned regression test and this report. Existing passing direct Codex/OpenCode/backfill source-package tests therefore carry forward for their unchanged source boundaries. No competing or full `make` suite was duplicated. The new full capture/CLI checks cover the scheduler's consumers and preserve previous precedence, alias, queue, publication and binding guards.

Tests and report are frozen again, with no unresolved reproduced regression in the assigned scheduler scope. Independent re-review and the parent-owned integrated gate remain necessary. Actual Codex trust/callback/agent interaction is still unverified; neither these scheduler fixtures nor the native synthetic replays close that runtime evidence gap.
