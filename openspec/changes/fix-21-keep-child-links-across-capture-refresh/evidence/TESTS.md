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

**Result.** Every acceptance case, every review round 1, 2 and 3 regression, and the approved position alignment pass on the current candidate. No data race is reported. The review findings are in [REVIEW_SUMMARY.json](REVIEW_SUMMARY.json), and the alignment rule is the design's Steering entry "Position alignment".
- **Mutations.** Each confirmed product fix and each test gap has a test that fails when its fix is removed in a scratch copy.
- **No new product defect.**
- **Incident scale.** The fixture now also holds unrelated sessions after the root. The whole family is extracted in pass 1 for 51,064,638 bytes, which is 9.5 % of 512 MiB, and pass 2 links 9 of 9 children. In the baseline, every pass spent the full 512 MiB and the root settled at 3 of 9 links.

### Revision and working tree

- **Tested revision:** `34068194`. The round-3 changes were tested before commit, and the committed tree is identical to the tested working tree.
  - **Coder's round 3 changes:**
    - R3-01 (duplicated fresh child; displaced child) and R3-03 (`ReplacedChildID`, docs) in `domain/child_links.go`.
    - R3-02 (`errRegistryWaitExpired`) in `capture/store.go` and `reconcile.go`, and the pure helper `childLinkRetentionOutcome` in `cmd/agentic-lens/capture.go`.
    - R3-03 renames in `capture/reconcile.go` and `archive_projection.go`.
    - R3-04: no early return on an empty ledger.
  - **Product files untouched.** The tester did not edit the product files. Their SHA-256 hashes, taken before the mutations and the final runs, matched afterwards.
- **Tester files:**
  - New in round 1: `capture/retain_child_links_test.go`, `capture/children_sweep_internal_test.go` and `codex/codexstoretest/import_guard_test.go`.
  - Changed in round 2: `domain/child_links_test.go`, `capture/retain_child_links_test.go`, `capture/reconcile_child_links_test.go`, `capture/children_sweep_internal_test.go`, `codex/discovery_window_test.go`, `codex/codexstoretest/import_guard_test.go` and `cmd/agentic-lens/capture_child_links_test.go`.
  - Changed in round 3: `domain/child_links_test.go`, `capture/retain_child_links_test.go` and `capture/reconcile_child_links_test.go`. New: `cmd/agentic-lens/child_link_retention_internal_test.go`.
  - Extended: `domain/child_links_test.go`, `capture/reconcile_child_links_test.go`, `capture/reconcile_archived_rollout_test.go`, `capture/reconcile_large_store_test.go`, `codex/discovery_window_test.go`, `codex/codexstoretest/store.go` and `store_test.go`, `cmd/agentic-lens/capture_child_links_test.go` and `capture_corrections_isolation_test.go`.
  - `tasks.md` gains tester task lines for DATA-02, ordinal retention, and position alignment with the round 2 gaps.
- **Toolchain:** Go 1.26.0, linux/amd64 (WSL2). All fixtures are under `t.TempDir()`.

### Commands

All commands run from `backend/`.

| Command | Result | Time |
|---|---|---|
| `go vet ./internal/domain/ ./internal/adapters/capture/ ./internal/adapters/codex/... ./cmd/agentic-lens/` | clean | 1 s |
| `go test -count=1 ./internal/domain/ ./internal/adapters/codex/ ./internal/adapters/codex/codexstoretest/` | PASS | 1 s |
| `go test -count=1 -timeout 30m ./internal/adapters/capture/` | PASS | 98 s |
| `go test -count=1 -timeout 40m ./cmd/agentic-lens` | PASS | 342 s |
| `go test -race -count=1 ./internal/domain/` | PASS, no data race | 6 s |
| `go test -race -count=1 -timeout 40m ./internal/adapters/capture/` | PASS, no data race | 366 s |

The round 2 runs are reused for the code round 3 did not change:
- `-race` on codex, three times on capture, and 20 focused contention runs;
- the large-store test at divisor 1 (pass 1 is 51,064,638 bytes, 9.5 % of 512 MiB, and pass 2 links 9 of 9);
- the import guard without `web/dist`.

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

### Review round 2: finding to test

| Finding | Test | What it asserts |
|---|---|---|
| R2-01, domain | "a partly re-read group keeps the unread position" (refusal table); `TestRetainAcceptedChildLinksAlignsGroupsByPosition/a partly re-read group keeps the unread link on every pass` | Fresh `[C1 ""]` against accepted `[C1 C2]` gives `[C1 C2]`, with C2 the only unresolved link and nothing refused. This holds on three passes, each taking the previous output as its accepted ledger. |
| R2-01, capture | `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses/second child re-read` | One child of a same-key `agent_id` group is re-read each pass. In two consecutive partial passes, the registry, refreshed and published owners keep both links, with exactly 1 unresolved. |
| R2-01, CLI second trigger; R2-T8 | `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger` | Machine B's registry stage freshly fills positions 1 and 2. The ledger stage keeps child 3 by position, with "Capture kept 1" and "could not place 0". On two runs: partial with 1 unresolved, and identical ledger bytes. |
| R2-02 | `AlignsGroupsByPosition/an untimed unique delegation keeps the accepted link against an unconfirmed child`, plus the confirmed-child counterpart | An unconfirmed child is kept pending confirmation, and a confirmed child replaces the accepted link. |
| R2-03 | "an accepted key without a fresh counterpart is reported as unmatched"; "an unmatched key whose child is linked elsewhere is not reported" | Reason `unmatched`, counted as unresolved, and the link is not placed. |
| R2-04 | `AlignsGroupsByPosition/a complete group emits no refusal decisions`; "a smaller fresh group is refused without decisions" | There are no decisions. A smaller group is reported only through `Refused`. |
| R2-05 | `TestCaptureAutoReportsOnlyClearedLinksWithoutPartial` | A run that only clears says "cleared 1 contradicted link(s); unresolved child evidence: 1." with no "remains partial", and its retention adds no unresolved evidence. |
| R2-06 | None | No seam. See Limits. |
| R2-T1 | `TestRetainAcceptedChildLinksReportsRegistryContention` | The retry now has its own context. The test passed 3 times within the capture package under `-race` and 20 times on its own under `-race`. |
| R2-T2 | `AlignsGroupsByPosition`: "a group interleaved with another key and offset between sides" and "two interleaved groups each keep their own order" | `["" C1 C2]` and `[C3 C1 "" C2]`. |
| R2-T3 | "a contradicted member of a refused group is not reported" | A smaller group reports only its uncontradicted C2. |
| R2-T4 | `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure` | A registry trigger rejects the partial diagnostic insert. The error is returned, and the latest recorded attempt stays completed. |
| R2-T5 | `TestHeaderSessionContextRejectsConflictingAncestry` | On the header-cache path, disagreeing direct and nested parents set `ConflictingParent`, and `SessionContext` returns `ErrSessionContextMismatch`. The agreeing, nested-only and direct-only shapes pass with the right parent. |
| R2-T6 | `TestQueuedChildPathsResumeAtTheRefusedPath` | At the ceiling the queued cursor is the last admitted path. The next one-read pass reads the refused path, not the one after it. |
| R2-T7 | `TestNoProductionPackageImportsCodexStoreTest` | `go list -e` lists the dependencies without the frontend build. |
| R2-T8 | `TestRetainAcceptedChildLinksReportsALinkKeptAndRefusedAsKept`, plus the CLI alignment case above | A link kept for refreshed owners and unmatched for restored owners counts as kept, not as refused, and once. |

### Review round 3: finding to test

| Finding | Test | What it asserts |
|---|---|---|
| R3-01(a), domain | Refusal table: "a fresh child linked from two delegations does not replace the accepted link" and "… duplicated within one key group …"; `TestRetainAcceptedChildLinksReplacementsAndDisplacedChildren/duplicated fresh child` | Accepted `[A B]` against fresh `[B B]` gives `[A B]`, with A kept pending confirmation and no replacement or `ReplacedChildID`. It is idempotent and stable when fed back. |
| R3-01(a), capture | `TestRetainAcceptedChildLinksIgnoresADuplicatedFreshChild` | Through `capture.RetainAcceptedChildLinks`: `[A B]`, one kept, counted once. |
| R3-01(b), domain | "a confirmed child moving position displaces the accepted child", "a swap of confirmed children displaces nothing", "a contradicted displaced child is not reported"; `ReplacementsAndDisplacedChildren/displaced` | Fresh `["" A]` gives B `displaced`, B unresolved exactly once and A not reported. The replacement decision records `ReplacedChildID` B. It is idempotent. Fed back, B is no longer reported. |
| R3-01(b), capture | `TestReconcileConfirmedChildMovingPositionCountsTheDisplacedChild` | A real re-read moves child a to the second delegation. The registry, refreshed and published owners give `["" a]`, with exactly 1 unresolved. |
| R3-02 | `TestRetainAcceptedChildLinksMarksAnExpiredRegistryWaitAsContention`; `TestChildLinkRetentionOutcome` | An expired wait on a held registry is a deadline marked as contention. An unmarked `DeadlineExceeded` or `Canceled` is not. The helper gives the busy exit for SQLite contention, joined contention and the capture-marked expiry, and exit 1 for other errors, including an unmarked deadline. |
| R3-T2 | `TestChildLinkRetentionOutcome` | The exact "remains partial" line when `Added > 0`. "." when links were only cleared, or kept but already counted. Nothing when there is nothing to report. |
| R3-03 | `ReplacementsAndDisplacedChildren` | `ReplacedChildID` on the replacement, and `ChildSessionID` on the held-by-sibling decision. |
| R3-04 | `TestRetainAcceptedChildLinksClearsContradictedLinksWithoutALedger` | With no ledger facts, a fresh link that the registry contradicts is cleared, with nothing counted. |
| R3-T1 | Refusal table: "a different agent type", "a different known invocation time" and "a different description is an unmatched key" | Each key field separates groups. |
| R3-T3 | Comments | The stale "ordinal" and "refused as a whole" comments were rewritten. `TestRetainAcceptedChildLinksOrdinalGroups` was renamed `TestRetainAcceptedChildLinksKeepsUnreadGroupsByPosition`. |

A CLI end-to-end case for R3-01 was not added. The CLI stage calls the same `capture.RetainAcceptedChildLinks`, which the API tests cover. The registry-stage path is covered by the real `Reconcile` case.

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
- **Domain oracles changed by round 3 (R3-01(b)).** A held-by-sibling child stays linked, so it is no longer reported. A child displaced by a confirmed replacement is reported as `displaced`.
  - "a child held by another final delegation is refused within an ordinal group" became "a child held by another final delegation is not duplicated or reported": nothing refused, only C2 unresolved, and the links are unchanged.
  - "held by a confirmed sibling" became "a confirmed child moving position displaces the accepted child": `["" C1]`, with C2 `displaced` and unresolved.
  - In `AlignsGroupsByPosition`, the untimed confirmed case now also reports the replaced C1 as `displaced`.
  - The refusal table's second-application comment no longer describes the removed whole-group refusal.
- **Domain oracles changed by position alignment (design Steering "Position alignment").** Two cases in `TestRetainAcceptedChildLinksReportsRefusedLinks` change by design, and are renamed:
  - "a fresh link set in the group refuses the whole group" became "an unconfirmed fresh link in a group keeps the accepted link at its position". It now gives `[C1 C2]`, with nothing refused and C1 and C2 unresolved.
  - "an accepted link the fresh read re-proves elsewhere is not refused" became "a partly re-read group keeps the unread position". It now gives `[C1 C2]`, with nothing refused and C2 unresolved.
  - A group is now refused only when the fresh group is smaller. Those cases are unchanged and gain a contradicted-member variant.
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
| M19 (superseded): the ordinal branch disabled, so every non-unique group is refused | Applied to the ordinal candidate, where 11 domain subcases and the capture two-pass test caught it. Position alignment removed that branch. M20 is its replacement. |
| M20: the whole-group refusal of round 2 reinstated when any fresh link is set (R2-01) | Domain: both renamed oracles and the alignment test. Capture "second child re-read": `["" group-child-2]` on pass 1. CLI alignment case: `[group-child-1 group-child-2 ""]`. |
| M21b: the accepted absolute index used in place of the group position (R2-T2) | Both interleaved cases: `[C1 C2 ""]` and `[C1 C3 C2 ""]`. |
| M22: an untimed unique delegation accepts an unconfirmed child (R2-02) | The untimed alignment case: `[C9]`. |
| M23: an unmatched key is not reported (R2-03) | The unmatched case: refused `[]`. |
| M24: the contradicted filter removed from refusals (R2-T3) | The contradicted-member case: C1 reported. |
| M25: complete positions are decided (R2-04) | The complete-group case: three `kept_pending_confirmation` decisions. |
| M26: the stderr line always claims partial (R2-05) | The cleared-only CLI case. |
| M27: not used | R2-T1 changed only a test (the retry's own context), so there is no product mutation. Its evidence is the repeated `-race` runs in round 2. |
| M28: the `RecordDiagnostic` error is discarded (R2-T4) | The diagnostic-failure test: `nil` returned. |
| M29: the header conflict flag is dropped (R2-T5) | The header ancestry test. |
| M30: the queued cursor is not restored at the ceiling (R2-T6) | The queued-path test: cursor on the refused path. |
| M31: the `delete(refused)` guard removed (R2-T8) | The kept-and-refused test: `Refused` 1. |
| M18 without `web/dist` (R2-T7) | The guard still lists every dependent production package. |
| M32: a duplicated fresh child confirms a replacement (R3-01(a)) | The two domain duplicated cases: `[C2 C2]`. The capture API case: `[B B]`. |
| M33: a displaced child is not reported (R3-01(b)) | The domain displaced and untimed-confirmed cases. The capture `Reconcile` case: 0 unresolved instead of 1. |
| M34: a replaced child still linked elsewhere is reported (R3-01(b) guard) | The swap case: C1 and C2 reported. |
| M35: `ReplacedChildID` not recorded (R3-03) | The replacement-decision assertion. |
| M36: the expired wait not marked (R3-02) | The capture expired-wait test, and the helper case for a capture-marked error: exit 1 instead of 75. |
| M37: the sentinel ignored by `IsRegistryContention` (R3-02) | The same two tests. |
| M38a and M38b: the wording always or never partial (R3-T2) | The helper's cleared-only and already-counted cases, and its "remains partial" case. |
| M39: the busy mapping removed from the helper | The helper's two contention cases. |
| M40: the early return on an empty ledger restored (R3-04) | The no-ledger clearing test. |
| M41a and M41b: agent type, or invoked_at, dropped from the key (R3-T1) | The matching unmatched-key row. |

### Defects

- **D1 is resolved** (first round).
- **Round 1 product findings DATA-01 to DATA-04, BEH-01, BEH-02, ARCH-03 and ARCH-04:** fixed by the coder and protected by the tests above.
- **Round 2 product findings R2-01 to R2-05:** fixed by the coder and protected by the tests above.
- **R2-06:** fixed by the coder, then reworked by R3-02.
- **Round 3 findings R3-01 to R3-04:** fixed by the coder and protected by the tests above.
- **No open product defect.**

### Limits

- **BEH-02, R2-06 and R3-02, end to end.** The busy exit is now tested at two boundaries: capture's marking of an expired wait, and the pure `childLinkRetentionOutcome` mapping. Only the `os.Exit` wiring in `retainCaptureChildLinks` is verified by reading. A lock that contends with the retention write also contends with the reconciliation that precedes it.
- **A cancelled wait on a held registry** returns the SQLite busy error itself, so it is contention. Only a deadline is marked.
- **An already expired retention context** is marked as contention whatever the cause, because capture checks only `ctx.Err()`. The CLI always passes a fresh 2 s context, so this only affects a caller that breaks the "bound only this call" contract.
- **DATA-01 trigger.** The test interrupts the pass at the annotation write with a registry trigger, not with a cancelled context. A counting context, as the review probe used, depends on how many context checks the code makes, so it is not stable.
- **Held by sibling (resolved in round 3).** A child that another final delegation holds is no longer reported or counted, because it stays linked.
- **Refused links are reported once.** This applies to a smaller fresh group, a key without a fresh counterpart (`unmatched`), and a child displaced by a confirmed replacement (`displaced`). A refused link is never moved, so the publication of that pass holds the fresh delegations, and later passes no longer see it. The ambiguous CLI case asserts this. A partly re-read group keeps its unread links across passes.
- **R2-P1 (pre-existing).** When two delegations share a task name and only one child is read, `resolveSourceChildren` points both at that child. FIX-21 tests do not cover it; it is recorded in REVIEW_SUMMARY.json as a follow-up.
- **Small `max_source_bytes` (accepted limit).** With an identification quarter under 64 KiB, an in-window rollout larger than the quarter, and not cached by another consumer, can stop a sweep at its cursor indefinitely. It shows as pending. A 12 KiB probe stayed at 10 of 18 owners for 80 passes. The BEH-01 tests therefore keep the quarter above the largest in-window rollout.
- **Shared dedup state.** A copied `ReconcileResult` shares its counted-links map, so a retention call on a copy affects the original's deduplication. Tests use a separate request for each call.
- **Contradiction warning.** It is still not surfaced by capture (accepted, pre-existing).
- **Spec tension.** The exact-UUID and directory pre-filters versus the literal "never skipped" wording are unchanged from the first round.
