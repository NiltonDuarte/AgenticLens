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

**Result.** Every acceptance case passes on the current candidate, with and without `-race`, and no data race is reported.
- **D1 is resolved.** The domain rule is now idempotent after a sibling release.
- **Deferred owners.** An owner deferred only by the identification reservation is now retried before owners discovered later in the same pass. A new regression test covers this.
- **Incident scale.** The whole family is extracted in pass 1 for 50,802,494 bytes, which is 9.5 % of 512 MiB, and pass 2 links 9 of 9 children. In the baseline, every pass spent the full 512 MiB and the root settled at 3 of 9 links.

### Revision and working tree

- **Tested revision:** `030bf0ec` plus the uncommitted candidate diff, including the coder's two follow-up changes:
  - the two-phase `reconcileOwnerChildLinks` in `domain/child_links.go`;
  - the `workBudget.deferred` set, which sorts deferred owners first in the final leftover refresh, in `capture/children.go` and `reconcile.go`.
- **Product files untouched.** The tester did not edit the eleven product files. Their SHA-256 hashes, taken before the final runs, matched after the runs and after every scratch mutation.
- **Tester files:**
  - New: `domain/child_links_test.go`, `capture/reconcile_child_links_test.go`, `codex/discovery_window_test.go`, `cmd/agentic-lens/capture_child_links_test.go` and `cmd/agentic-lens/capture_corrections_isolation_test.go`.
  - Rewritten: `capture/reconcile_large_store_test.go`.
  - Extended: `capture/reconcile_archived_rollout_test.go`, `capture/reconcile_nested_codex_test.go` and `capture/reconcile_test.go`.
  - Moved: the generator is now the stdlib-only, test-support package `internal/adapters/codex/codexstoretest`, with its own self-test. `capture/large_codex_store_fixture_test.go` was removed.
- **Toolchain:** Go 1.26.0, linux/amd64 (WSL2). All fixtures are under `t.TempDir()`.

### Commands

All commands run from `backend/`.

| Command | Result | Time |
|---|---|---|
| `go test -count=1 ./internal/domain/ ./internal/adapters/codex/ ./internal/adapters/codex/codexstoretest/` | PASS | 2 s |
| `go test -count=1 -timeout 30m ./internal/adapters/capture/` | PASS | 117 s (83 s before the new tests) |
| `go test -count=1 -timeout 40m ./cmd/agentic-lens` | PASS | 387 s (316 s before the new tests) |
| `go test -race -count=1 -timeout 40m ./internal/domain/ ./internal/adapters/codex/ ./internal/adapters/capture/` | PASS, no data race. capture 410 s (baseline 379 s) | 426 s |
| `go test -count=1 -run TestLargeStore -v ./internal/adapters/capture/ -fix21.large-store-divisor=1` | PASS | 13 s (baseline 53 s) |
| `go list -deps` over `./...` | No production package depends on `codexstoretest`. Its imports are stdlib only. | |

### Accounting oracles

A header read costs `min(size, ceil(header line / 4 KiB) × 4 KiB, 64 KiB)`. That is 32,768 bytes for the fixture's 29,696-byte header. Each header is read at most once per pass. Extraction costs the file size.

- **Large store, pass 1:** family bytes, plus each family header once, plus the three unconventional names once. The non-date file swept later is already cached. Attempts: 10.
- **Large store, pass 2:** the same total, and the root links 9 of 9 children. The same holds at `max_sessions: 256`. In the baseline, a lookup failed at listing at that budget.
- **Child sweep:** charges only the family headers and the non-date file. It uses 0 session slots and 0 attempts. Baseline: 121,453,353 bytes and 1,906 slots.
- **Capture exact lookup:** claims and reads only same-UUID and unconventional names. Duplicate copies stay ambiguous. Live lookup is unchanged and still claims every rollout.
- **Reserved quarter:**
  - **Extraction ahead of discovery:** the replay runs the refresh class before the sweep. Extraction stays at three quarters or less, with the quarter untouched. Deferred owners are not attempted, and their slots are refunded. The hinted child is then discovered from the quarter. The pass total is exact.
  - **Unused share:** with one session, the root alone needs more than three quarters. It is read in one attempt, with the exact need.
  - **Ceiling:** identification starts exactly 4 child headers under the 64 KiB start rule. The root is extracted, and discovery is pending, not failed. The next pass continues.
- **Deferred owner first** (`TestLargeStoreDeferredOwnerPrecedesLaterDiscoveryInFinalRefresh`):
  - **Setup:** one session. The root needs more than three quarters, and one child is discovered from the quarter in the same pass.
  - **Assertions:** the root takes the refunded slot in the final refresh. Attempts: 1. Bytes equal the root's need plus the child's header exactly. The child stays registered but unread.
  - **Key order:** the store is relocated to directory names that make the child's registry key sort after the root's in one case, and before it in the other. The outcome is asserted for both orders.
- **CET and CEST across local midnight:** every child is found. Unconventional names inside pruned old day directories cost nothing, which proves those directories are not listed.

### Rationale for changed existing tests

- **`TestReconcileCodexRecoversNestedSpawnWithoutChildCallbacks`:** its UUIDv7 IDs (`019c0000-0000-7…`) encode 2026-01-27T15:00:18.048Z, but the files were in `sessions/2026/01/01`. The 14-hour directory rule correctly pruned that directory. The directory, file names and record times now match the encoded instant, as Codex writes them. The test's intent and assertions are unchanged.
- **`TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities`:** `SourceBytes == allowance` held only because the work equalled the cap. The oracle is now derived from the new rules: every candidate header read once, at whole-file cost because each file is under 4 KiB, plus the bound owner's extraction. It is asserted at a tight allowance, which goes through deferral and the final refresh, and at double that allowance. The original assertions are kept.
- **Reviewed and unchanged:**
  - `TestReconcileChargesDiscoveryAndExtractionOnce…` and the failed-read accounting test in `reconcile_test.go`: their exact-fit allowances still discriminate. Any extra charge would starve the extraction they assert.
  - `fix19_budget_test.go`: its oversized first record is still read in full by the exact lookup's historical tolerance.
  - `fix19_candidate_budget_internal_test.go` and `codex/discovery_budget_test.go`: they cover manual or live discovery, or have no split.
  - `codex/discovery_test.go`: manual discovery passes unchanged.

### Mutation checks

Each mutation was applied in a scratch copy of `backend/`, never in the checkout.

| Mutation | Detected by |
|---|---|
| M1: carry-forward removed. In the current two-phase rule, the accepted child is no longer assigned in the fixpoint. | 6 domain table cases. Capture partial refresh: registry links lose children 3 and 4. |
| M2: the `retainCaptureChildLinks` call in `main.go` disabled | Both CLI cases: published links all empty. |
| M3: `SkipsFile` and `SkipsDirectory` return false | Large-store sweep, pass, allowance and local-midnight tests, plus the codex window tests. |
| M4: the worker's isolated-row exclusion removed | Worker case: killed at 8 s, never idle. |
| M5: the deferred slot refund removed (`sessions++`, `attempts--`) | Allowance replay, the one-session unused-share case and the tight unassigned-roots case. |
| M6: the deferred-first sort in `refreshTicketOwners` removed | The deferred-owner test, under both key orders: the root stays pending with "Not refreshed within this pass" while the later child takes the slot. |

M1 to M5 were run on the first candidate. M1 was repeated against the current domain rule.

### Defects

- **D1, resolved.** The two-phase `reconcileOwnerChildLinks` fixed it.
  - **Trigger:** accepted delegations D1→C1 and D2→C2. The fresh read leaves D1 empty, and links D2 to C1 without confirmation.
  - **Before the fix:** the result alternated, `["" C2]` and then `[C1 C2]`.
  - **Now:** `TestRetainAcceptedChildLinksIsIdempotentAfterSiblingRelease` passes, and every domain table case passes its second-application check.
- **No open defects.**

### Limits

- **Contradiction warning.** For a task-name-only result, the codex adapter raises no warning, because the child no longer matches the parent's key. For an explicit `session_id` result, the warning is raised but capture does not surface it: `CompleteSessionIDs` comes from the transcripts. The tests assert that the link is cleared, not the warning.
- **Cache re-read on change** is proven at the `HeaderCache` level, for both a size change and an mtime-only change. A file changing in the middle of a capture pass has no deterministic capture-level trigger.
- **Rotation cursor.** A deferred owner still advances the persisted rotation cursor, by design. Holding the cursor broke `TestFIX19CodexDiscoveryCannotPermanentlyStarveFittingOpenCodeOwner`. Within the pass, deferred-first ordering compensates. Across passes, ordering follows the existing rotation.
- **Spec tension, for the planner.** The exact lookup skips a conventional name carrying another UUID, including a non-v7 UUID or a name whose metadata names a different session. The directory rule also skips unconventional names inside pruned old directories. Both are tested as implemented. They conflict with the literal "never skipped" wording of "File names never establish identity".
