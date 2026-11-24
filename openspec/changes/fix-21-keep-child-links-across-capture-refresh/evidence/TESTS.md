# FIX-21: Test evidence

## Baseline measurement

This is what each capture consumer spends on a large synthetic Codex store before any discovery-budget code change. It serves as the reference for the [discovery budget spec](../specs/capture-discovery-budget/spec.md).

**Result.** The incident pattern reproduces. At the incident budget, identification uses 95 to 97 % of every pass's 512 MiB. Each pass extracts at most 4 of the 10 family owners. The root is never read in the same pass as more than 3 of its 9 children, so its accepted links settle at 3 of 9. The real incident lost 7 of 9. The family is never extracted together in one pass. With the example `max_sessions: 256`, the root is never extracted at all.

### Revision and commands

- **Tested revision:** `030bf0ec` (the approved plan), with no product change. The diff consists of two new, untracked test files:
  - `backend/internal/adapters/capture/large_codex_store_fixture_test.go`: the generator.
  - `backend/internal/adapters/capture/reconcile_large_store_test.go`: the measurement, `TestLargeStoreFamilyWithinCaptureBudget`.
- **Toolchain:** Go 1.26.0, linux/amd64 (WSL2), 12 CPUs.
- **Commands**, all run from `backend/`:

| Command | Result | Time |
|---|---|---|
| `go test -count=1 -run TestLargeStoreFamilyWithinCaptureBudget -v ./internal/adapters/capture/ -fix21.large-store-divisor=1 -fix21.large-store-passes=8` | PASS | 53.4 s |
| `go test -count=1 -run TestLargeStoreFamilyWithinCaptureBudget -v ./internal/adapters/capture/` (defaults: divisor 4, 3 passes) | PASS | 17.7 s |
| `go test -race -count=1 -run TestLargeStoreFamilyWithinCaptureBudget -v ./internal/adapters/capture/` | PASS | 74.2 s |
| `go test -race -count=1 ./internal/adapters/capture/` (whole package, new test included) | PASS | 379 s |

The figures below come from the incident-scale run (divisor 1). A second run produced the same numbers.

### Fixture shape

Everything is generated under `t.TempDir()`. Nothing reads `~/.codex` or a ledger.

| Property | Divisor 1 (incident scale) | Divisor 4 (default) |
|---|---|---|
| Rollouts | 1,907: 1,506 live, 401 archived | 489 |
| Store bytes (apparent) | 2,951,555,623 (2.75 GiB) | 751,920,694 (0.70 GiB) |
| Allocated on disk | 113 MiB | 30.5 MiB |
| Family | root 12,601,143 B + 9 children of 4,197,263 B = 50,376,510 B (48.0 MiB) | 12,576,576 B (12.0 MiB) |
| `max_source_bytes` | 536,870,912 (512 MiB) | 134,217,728 (128 MiB) |
| Budget : store | 1 : 5.50 | 1 : 5.60 |
| Family : budget | 0.094 | 0.094 |

The incident had 1,909 rollouts, about 3 GB, a family of about 50 MB and 512 MiB, so about 1 : 6. Divisor 4 shrinks the rollout count, the family sizes and the budget together. Header size, the per-file size distribution and the family member count stay the same. Divisor 4 reproduced the same per-pass pattern.

- **Names.** Live rollouts follow `sessions/YYYY/MM/DD/rollout-YYYY-MM-DDThh-mm-ss-<uuid>.jsonl` in Europe/Berlin local time, which covers both CET and CEST. The UUIDv7 encodes the UTC creation millisecond. Archived rollouts are flat in `archived_sessions/`.
- **Headers.** Every first line is a `session_meta` of exactly 29,696 bytes, padded through `base_instructions`.
- **Unrelated rollouts.** Every 10th is a dense file under 64 KiB. The others have a header and two real records, then a sparse hole up to an apparent size between 96 KiB and 3.2 MiB.
- **Irregular names.** The store contains each of these once:
  - a name without a UUID
  - a custom name
  - a non-date directory
  - a UUIDv4 name
  - a name whose UUID differs from its metadata
  - a gzip-compressed rollout
  - a same-UUID archived copy of one unrelated live rollout
- **Family.** The root is created after every unrelated rollout. It makes 9 task-name-only `spawn_agent` calls, and its output carries only `task_name`.
  - Each child names the root only in `source.subagent.thread_spawn.parent_thread_id`, with a matching `agent_path`. The children are created 7 s to 50 min after the root.
  - The family is dense, real JSONL. A full combined parse links all 9 children with no warnings.
- **Registry.** The root is registered and bound to `TEST-100`. One `QueueChildStart` hint per child mirrors Codex `SubagentStart` hooks.
- **Entry point.** Each pass uses `OpenProject`, `SetSourceBudgets` and `ReconcileAt`, as the worker does.

### Per-consumer cost

Each consumer was measured in isolation, with unbounded budgets so it could finish.

| Consumer | Bytes charged | Claims and slots | Notes |
|---|---|---|---|
| Exact lookup per owner (`DiscoverExactRolloutPathWithBudget`) | 121,486,121 (115.9 MiB) | 1,907 discovery claims, from listing every rollout | Equals Σ min(size, 64 KiB) over the 1,906 uncompressed rollouts (121,453,353), plus 32,768 for the gzip header. Identical for the root and for a child. |
| Exact lookup of the duplicated UUID | 26,072,408 | 1,907 claims | Stops at the second copy and reports it as ambiguous. |
| Exact lookup with `max_sessions: 256` | 0 | 256 claims, then `ErrDiscoveryBudget` | The listing claims every file before reading any. |
| Owner preflight (`validateReadOwner`) | 65,536 per owner | none | A flat 64 KiB reservation, against a 29,696-byte header. |
| Child sweep (`recoverChildren`, starts class) | 121,453,353 (115.8 MiB) | 1,906 header reads, each taking 1 pass session slot and 1 `SourceAttempt` | Reads every `.jsonl`, including the family. It never reads `.gz`. All 9 children are inherited. |
| Extraction (`extractExactAt`) | root 12,601,143; child 4,197,263 | 1 session slot and 1 attempt per owner | Equals the file size. |

**How the first pass splits.** The test calls the phase-0 work classes, `refreshTicketOwners` and `recoverChildren`, with one `workBudget`. It reads the lookup share from `workBudget.codexDiscovery.SourceBytesRead` and the extraction share from the accepted rows' `Revision`. The preflight share is the remainder. To validate this replay, the real `ReconcileAt` runs on a fresh registry over the same store and reports the same totals: `SourceBytes` 536,870,912 and `SourceAttempts` 1,916.

| Pass 1 step | Preflight | Exact lookup | Child sweep | Extraction |
|---|---|---|---|---|
| Refresh | 65,536 | 121,486,121 | | 12,601,143 (root) |
| Starts class | | | 121,453,353 (1,906 reads) | |
| Leftover refresh, 9 owners attempted | 196,608 | 272,673,625 (2 full lookups and 1 cut short) | | 8,394,526 (2 children) |
| **Total: 536,870,912** | 262,144 | 394,159,746 | 121,453,353 | 20,995,669 (3.9 %) |

In the leftover refresh, the third lookup ran out mid-scan with "Codex discovery session/byte budget reached". The other six children were left pending with "Source byte budget reached".

### Repeated worker passes

These are 8 passes, a full scheduler rotation, at `max_sessions: 4096` and 512 MiB.

| Pass | `SourceBytes` | Root refreshed | Children refreshed | Root links | Identification share |
|---|---|---|---|---|---|
| 1 | 536,870,912 | yes | 2/9 | 0/9 | 96.1 % |
| 2 | 536,870,912 | no | 4/9 | 0/9 | 96.9 % |
| 3 | 536,870,912 | yes | 3/9 | 3/9 | 95.3 % |
| 4, 5 | 536,870,912 | no | 4/9 | 3/9 | 96.9 % |
| 6 | 536,870,912 | yes | 3/9 | 3/9 | 95.3 % |
| 7 | 536,870,912 | no | 4/9 | 3/9 | 96.9 % |
| 8 | 536,870,912 | yes | 3/9 | 3/9 | 95.3 % |

- **Pass state.** Every pass reported `DiscoveryPending=false`, `PendingChildren=0` and `UnresolvedParentLinks=0`. The diagnostic was still partial, because six owners stayed pending.
- **Session budget.** With `max_sessions: 256`, the root's exact lookup fails at listing in every pass.
  - In passes 1 and 2, the sweep spends the rest of the session allowance on 255 headers (about 15.5 MiB).
  - In pass 2, the sweep runs before the refresh class, so the root is not attempted ("Not refreshed within this pass").
  - Identification consumes 100 % of each pass.

### Limits

- **Uncharged reads.** Two header reads are not charged to the pass, so they are not measured. One is the validation of each inherited child in `inheritChild`, through `validateOwner` and `ReadSessionContextWithinBudget`. The other is binding verification, which happens outside the pass.
- **Run-to-run variation.** Which children share a pass varies between runs, because the registry orders owners by a hash of the identity, which includes the temporary store path. The counts and bytes do not vary.
- **Replay scope.** The replay mirrors only phase 0. The later passes are measured only through `ReconcileAt`.
- **Incident session budget.** The incident's `max_sessions` is unknown. The byte-bound failure implies at least the store size, so both 4096 and 256 were measured.
- **Sparse files.** The sparse unrelated bodies are not JSONL. A manual nonce scan, which reads whole files, would need dense bodies (`SparseUnrelated: false`) on a smaller store.
- **Runtime.** The default divisor 4 adds about 74 s to the package's race run. The whole package took 379 s, against the 10-minute default timeout. Divisor 1 is selected explicitly with `-fix21.large-store-divisor=1`.
- **Shared generator.** `codex/discovery_window_test.go` also needs the generator, and Go test files cannot be shared across packages. The proposal is a stdlib-only, non-test helper package, for example `internal/adapters/codex/codexstoretest`. Until then, the generator stays as a `capture` package test helper.
- **Assertions.**
  - **Asserted now:** fixture validity; the byte ceiling; the live exact-lookup oracle, which stays unchanged by design; the 256-session listing exhaustion; sweep completeness when unbounded; and extraction equal to the file size.
  - **Logged only:** per-pass outcomes and the preflight and sweep reservation charges. A later task converts these into acceptance assertions.

The baseline section above is the historical "before". Its generator and measurement files have since moved, as recorded below.

## Candidate verification

**Result.** Every acceptance case, every review round 1 regression and the approved ordinal group retention pass on the current candidate. No data race is reported. The review findings are in [REVIEW.md](REVIEW.md), and the ordinal rule is the design's Steering entry "Ordinal group retention".
- **Mutations.** Each confirmed product fix and each test gap has a test that fails when its fix is removed in a scratch copy.
- **No new product defect.**
- **Incident scale.** The fixture now also holds unrelated sessions after the root. The whole family is extracted in pass 1 for 51,064,638 bytes, which is 9.5 % of 512 MiB, and pass 2 links 9 of 9 children. In the baseline, every pass spent the full 512 MiB and the root settled at 3 of 9 links.

### Revision and working tree

- **Tested revision:** commit `0f73f759`, plus the coder's uncommitted round 1 fixes, the ordinal group retention in `domain/child_links.go`, and the tester's uncommitted tests. Since round 1, only `domain/child_links.go` changed among the product files.
  - **Coder's fixes:** `domain/child_links.go`, `capture/reconcile.go`, `children.go` and `archive_projection.go`, `codex/context.go`, `discovery.go` and `discovery_budget.go`, `cmd/agentic-lens/capture.go` and `main.go`.
  - **Product files untouched.** The tester did not edit the product files. Their SHA-256 hashes, taken before the mutations and the final runs, matched afterwards.
- **Tester files:**
  - New: `capture/retain_child_links_test.go`, `capture/children_sweep_internal_test.go` and `codex/codexstoretest/import_guard_test.go`.
  - Extended: `domain/child_links_test.go`, `capture/reconcile_child_links_test.go`, `capture/reconcile_archived_rollout_test.go`, `capture/reconcile_large_store_test.go`, `codex/discovery_window_test.go`, `codex/codexstoretest/store.go` and `store_test.go`, `cmd/agentic-lens/capture_child_links_test.go` and `capture_corrections_isolation_test.go`.
  - `tasks.md` gains a DATA-02 tester task line and an ordinal-retention tester task line.
- **Toolchain:** Go 1.26.0, linux/amd64 (WSL2). All fixtures are under `t.TempDir()`.

### Commands

All commands run from `backend/`.

| Command | Result | Time |
|---|---|---|
| `go test -count=1 ./internal/domain/` | PASS | under 1 s |
| `go test -count=1 -timeout 30m ./internal/adapters/capture/` | PASS | 104 s |
| `go test -count=1 -timeout 40m ./cmd/agentic-lens` | PASS | 365 s |
| `go test -race -count=1 ./internal/domain/` | PASS, no data race | 8 s |

The following round 1 runs are reused, because their packages' product code is unchanged since then: codex and codexstoretest (PASS), `-race` on codex and capture (PASS, capture 417 s), and the large-store test at divisor 1 (PASS: pass 1 is 51,064,638 bytes, 9.5 % of 512 MiB, and pass 2 links 9 of 9). Capture under `-race` was not rerun after the domain change. The domain package itself passes under `-race`.

### Review round 1: finding to test

| Finding | Test | What it asserts |
|---|---|---|
| DATA-01 | `TestReconcileInterruptedAnnotationKeepsRegistryLinks` | A registry trigger aborts the parent's first write after its refreshed snapshot commits, which is the family annotation, the same seam the archived checkpoint tests use. The registry keeps all 4 links. The follow-up partial pass keeps them and reports 2 unresolved. |
| BEH-01 | `TestResumedSweepAtTheCeilingKeepsItsFileCursor`, `TestBrandNewSweepAdmittingNothingWritesNoCursor` | Two parents under one ticket. The second parent, with a real file cursor, finds identification below its next candidate's 64 KiB start bound, so its cursor, hints and discoveries are unchanged and discovery is pending. Later `ReconcileAt` passes find all 8 children. A brand-new sweep that admits nothing writes no cursor row and is found later. |
| DATA-02 | `TestRetainAcceptedChildLinksReportsRefusedLinks` (domain), `TestRetainAcceptedChildLinksCountsRefusedLinksOnceBesideFreshLinks` (capture API), CLI case "ambiguous accepted provenance" | Typed reasons, and no link is moved. The count is exact, the child also queued as a start is counted once, and the attempt is partial. stderr reports "could not place 2 accepted link(s)". |
| BEH-02/ARCH-03 | `TestRetainAcceptedChildLinksReportsRegistryContention` | A held `BEGIN IMMEDIATE` surfaces as registry contention, and nothing is recorded. After release, a new request records the partial diagnostic with 4 unresolved children. |
| ARCH-04 | `TestHeaderCacheAdmitsCompressedHeaderOnlyWithTheWholeBound` | A `.gz` header that fits its compressed size but not its decompressed header reads nothing below 64 KiB. With the whole bound, it is charged exactly the decompressed bytes. |
| DATA-03 | CLI case "queued path of an unregistered child" | The queued path and the kept link of the same child count once: 4 unresolved, not 5. |
| DATA-04 | `TestRetainAcceptedChildLinksPrefersKnownChildEvidence` (domain), `TestRetainAcceptedChildLinksUsesRegistrySessionsAsChildEvidence` (capture API) | Registry proof replaces, or contradicts, the older ledger link. Agreement changes nothing, and a fresh read still wins. |
| TEST-01 | `TestReconcileArchivedRegroupKeepsAcceptedChildLink` | An archived child is superseded after the annotation. The regrouped parent keeps its accepted link in the refreshed and published owners, counted once, and its spawn event carries no child. |
| TEST-02 | CLI case "some children read with the parent on the second run" | Exact counts: run 1 keeps 4 (4 unresolved). Run 2 resolves children 1 and 2 freshly and keeps 3 and 4. Child 3's hint and kept link count once: 2 unresolved, not 3. |
| TEST-03 | `TestChildSweepWindowFromNonV7ParentTimestamp` | A UUIDv4 parent's `session_meta` timestamp sets the window. The children and a later rollout are read, earlier rollouts cost nothing, and the byte total is exact. |
| TEST-04 | `TestRetainAcceptedChildLinksRewritesRestoredSessions` | A restored parent regains C1. C2, which the same request refreshed with another parent, stays cleared. |
| TEST-05 | `TestLargeStoreFamilyWithinCaptureBudget` with 6 later unrelated rollouts and two irregular names in the root's day directory | The sweep and pass byte oracles include them exactly. The exact lookup still reads only unconventional and same-UUID names. |
| TEST-06 | `TestRetainAcceptedChildLinksIsIdempotentAfterSiblingRelease`; `TestNoProductionPackageImportsCodexStoreTest`; `TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout`; `TestManualGenerateLeavesLinkToMissingChildEmpty`; CLI case "contradicting child read on machine B" | Exact `[C1 C2]` with its outcomes. A `go list`-based import guard. The worker is judged by its outcome, under a harness allowance of 3 times its lifetime, with no wall-clock bound. Manual generation leaves a missing child's link empty. `--capture-auto` clears a contradicted link. |

| Ordinal group retention | `TestRetainAcceptedChildLinks` (3 renamed cases), `TestRetainAcceptedChildLinksReportsRefusedLinks` (4 renamed and 4 new cases), `TestRetainAcceptedChildLinksOrdinalGroups`, `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses` | Accepted link i is carried to fresh delegation i, and appended delegations stay empty. A group with one fresh link set, or a smaller fresh group, is refused and reported, including with unknown invoked_at. A contradicted child is skipped, not refused. A child held by another final delegation is refused as held by a sibling. The rule is idempotent and invents no link. Repeated passes, each taking the previous published output as the accepted ledger, keep the links and report them every time. In capture, two consecutive partial passes keep both links of an ambiguous group in the registry and the published owners, with 2 unresolved each pass. This removes the earlier "reported once, then lost" limit. |

All CLI cases now assert exact per-run counts (kept, refused, unresolved) and ledger byte identity on the repeat. The CLI ambiguous case is unchanged by the ordinal rule, because its fresh group is smaller than the accepted group, which is still a refusal.

### Accounting oracles

A header read costs `min(size, ceil(header line / 4 KiB) × 4 KiB, 64 KiB)`, and decompressed bytes for `.gz`. Each header is read at most once per pass. Extraction costs the file size.

- **Large store, pass 1:** family bytes, plus each family header once, plus the unconventional names once, plus the later and in-window non-v7 rollouts that only the sweep reads. Attempts: 10.
- **Large store, pass 2:** family bytes, family headers and unconventional names, and the root links 9 of 9 children. The same holds at `max_sessions: 256`.
- **Child sweep:** charges the family headers and the in-window rollouts only. It uses 0 session slots and 0 attempts. Baseline: 121,453,353 bytes and 1,906 slots.
- **Reserved quarter, deferral and deferred-first ordering:** as in the first round. These are unchanged, and the tests still pass.

### Rationale for changed existing tests

- **`TestReconcileCodexRecoversNestedSpawnWithoutChildCallbacks`:** the files' dates now match their UUIDv7 instants. The test's intent and assertions are unchanged.
- **`TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities`:** the oracle is now derived from the new rules, at a tight and a double allowance. The original assertions are kept.
- **CLI cross-machine test:** `PendingChildren >= 2` became exact per-run counts (TEST-02), and five cases now cover the review findings.
- **Worker isolation test:** the 8 s wall-clock bound was replaced by the outcome under a 30 s harness allowance (TEST-06).
- **Large-store fixture:** it gains opt-in later and in-window rollouts. Existing identities are unchanged, because the new files are drawn after the family.
- **Domain oracles changed by the ordinal rule (design Steering "Ordinal group retention").** They are renamed to their new meaning:
  - "duplicate provenance is carried by position", "a larger fresh group is carried by position" and "unknown invocation time is carried by position" were refusals. They are now carries, with the `retained_by_position` outcome.
  - In the refusal test, the three corresponding cases now assert no refusal and the carried links as unresolved.
  - "contradicted child in an ordinal group is skipped, not refused" now carries the uncontradicted sibling.
  - The refusal branch keeps its own cases: a fresh link set in the group, a smaller fresh group (timed and untimed), and a child held by another final delegation.
  - The second-application check in the refusal test now compares facts only. After an ordinal carry, the rule's own output is a partly linked group, which is refused as a whole with nothing left to report. The design's idempotence is about facts, so the links are unchanged.

### Mutation checks

Each mutation was applied in a fresh scratch copy of `backend/`, never in the checkout.

| Mutation | Detected by |
|---|---|
| M1 to M6 (first rounds: carry-forward, CLI call, window pre-filter, worker isolation exclusion, slot refund, deferred-first sort) | As recorded earlier. M1 was repeated against the current rule. |
| M7: per-owner retention removed (DATA-01) | Interrupted-annotation test: registry links `["" "" "" ""]`. |
| M8c: both sweep fixes reverted, back to candidate `0f73f759` behaviour (BEH-01) | Both sweep tests: cursor moved to `""` (closed), and an empty sweep wrote `""`. |
| M8a: only `last := cursor` reverted | Not detected. It is an equivalent mutant: the new "admitted nothing, write no row" branch alone keeps the saved row. |
| M8b: only the no-row branch removed | Brand-new sweep test. |
| M9: refusals not reported | The domain, capture API and CLI ambiguous cases. |
| M10: contention type lost (`errors.New(err.Error())`) | The contention test. |
| M11: `.gz` admitted by compressed size | The gzip test: 49,151 bytes read, then abandoned. |
| M12: name-based dedup removed (DATA-03) | CLI queued-path case: 5 unresolved instead of 4. |
| M13: `Known` dropped from `capture.RetainAcceptedChildLinks` | The registry-evidence API test: the older ledger link wins. |
| M14: regroup retention removed (TEST-01) | The regroup test: an empty link after the regroup. |
| M15: parent timestamp shifted by +48 h (TEST-03) | The non-v7 window test: 312 bytes charged against 2,476, children pruned. |
| M16: the restored rewrite removed (TEST-04) | The restored-sessions test. |
| M17: non-v7 names pruned by their UUID bits (TEST-05) | Large-store sweep and pass-1 oracles: 32,768 bytes short. |
| M18: a production import of `codexstoretest` (TEST-06) | The import guard lists every dependent production package. |
| M19: the ordinal branch disabled, so every non-unique group is refused | 11 domain subcases. The capture two-pass test: published group links `["" ""]` on pass 1. |

### Defects

- **D1 is resolved** (first round).
- **Round 1 product findings DATA-01 to DATA-04, BEH-01, BEH-02, ARCH-03 and ARCH-04:** fixed by the coder and protected by the tests above.
- **No open product defect.**

### Limits

- **BEH-02, end to end.** The busy exit code is asserted at the capture API boundary only. A lock that contends with the retention write also contends with the reconciliation that precedes it, and the product has no seam between them. The CLI maps contention to `capturePublicationBusyExit` in one branch, verified by reading.
- **DATA-01 trigger.** The test interrupts the pass at the annotation write with a registry trigger, not with a cancelled context. A counting context, as the review probe used, depends on how many context checks the code makes, so it is not stable.
- **Held by sibling counts even when the child stays linked.** A child confirmed on another delegation of the same parent is still counted once as unresolved, which is the documented "not linked from the delegation they had" rule. It over-reports for one pass.
- **Refused links are reported once.** This now applies only to the refusal branch: a fresh link set in the group, or a smaller fresh group. A refused link is never moved, so the publication of that pass holds the fresh delegations, and later passes no longer see it. The ambiguous CLI case asserts this. Groups that the ordinal rule can align keep their links across passes.
- **Small `max_source_bytes` (accepted limit).** With an identification quarter under 64 KiB, an in-window rollout larger than the quarter, and not cached by another consumer, can stop a sweep at its cursor indefinitely. It shows as pending. A 12 KiB probe stayed at 10 of 18 owners for 80 passes. The BEH-01 tests therefore keep the quarter above the largest in-window rollout.
- **Shared dedup state.** A copied `ReconcileResult` shares its counted-links map, so a retention call on a copy affects the original's deduplication. Tests use a separate request for each call.
- **Contradiction warning.** It is still not surfaced by capture (accepted, pre-existing).
- **Spec tension.** The exact-UUID and directory pre-filters versus the literal "never skipped" wording are unchanged from the first round.
