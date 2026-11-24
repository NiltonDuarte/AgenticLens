# FIX-21 test-effectiveness review: tests-sonnet

- Run: FIX-21 discovery run tests-sonnet, focus reviewer-tests, requested profile Claude Sonnet 5.5 / high (runtime self-report: claude-sonnet-5-5; effort not independently verifiable).
- Reviewed revision: 0f73f759 (frozen candidate), base 133100bb. Working tree had only an unrelated tasks.md edit.
- Independence: other reviewers' reports were not read.

## Verdict: CHANGES_REQUESTED (low severity; no defect in product behavior found)

The suite reaches every claimed behavior through real boundaries with strong oracles. Three product paths introduced by this change are not protected by any assertion (verified by surviving mutations). None is a blocker on its own; they are test-gap findings for the tester/parent to close or consciously accept.

## Findings

### T1. Retention in `regroupArchiveFamily` is unprotected (severity 4, introduced)
- Location: `backend/internal/adapters/capture/archive_projection.go` (the added `RetainAcceptedChildLinks` call, `owners = retention.Sessions`).
- Expectation: design D3-A says the rule runs in `finishCaptureAnnotations` and `regroupArchiveFamily`. Task list claims both.
- Trigger: a partial pass whose archive family is regrouped after a stale owner is discarded, with an accepted child link the regrouped read cannot re-prove.
- Consequence: a regression (or deletion) of the regroup retention re-opens the original loss of accepted `child_session_id` links on that path, silently.
- Evidence: scratch mutation removing `owners = retention.Sessions` in a copy of the candidate. The whole `./internal/adapters/capture/` package passed (180 s; the only failure was `TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject`, which also fails unmutated in the scratch copy because the copy lacks `scripts/`). No new test builds a stale archive family with an accepted link.
- Suggested evidence: one `reconcile_archived_rollout_test.go` case where a family is regrouped after a stale member and the accepted parent link survives.

### T2. Identification stop/resume bookkeeping in `recoverChildren` is unasserted (severity 4, introduced)
- Location: `backend/internal/adapters/capture/children.go` (queued-path loop cursor restore `saveChildCursor(... previous ...)`; starts-sweep `last = previous` on `ErrCaptureBudget`).
- Expectation (design, spec): "identification stops at its ceiling with discovery pending, not failed" and the next pass "continues"; the unread candidate must be first next pass.
- Trigger: identification reaches its quarter in the middle of a sweep or of the queued-path list.
- Consequence: the unread candidate can be skipped until the cursor wraps, which delays a child's discovery (bounded, not data loss). `TestLargeStoreIdentificationAllowance/identification stops at its ceiling...` only asserts `discovered > reads` for the next pass, not which candidate is read.
- Evidence: two scratch mutations, each survived the full capture package (same single environment-only failure):
  - G: `last = previous` replaced by `_ = previous` in the starts sweep.
  - D: the queued-path cursor restore removed.
- Suggested evidence: assert the exact child read first in the next pass (the first unread index), and cover the `pending_children` path (queued child path) at the ceiling.

### T3. CLI rewrite of `RestoredSessions` is unprotected (severity 3, introduced)
- Location: `backend/cmd/agentic-lens/capture.go`, `retainCaptureChildLinks` (`result.RestoredSessions = restored.Sessions` and the `Observed` wiring).
- Expectation: design says the helper runs on `RefreshedSessions` and `RestoredSessions` against `baseline.Current`.
- Trigger: `--capture-auto` publishes an owner restored from registry state whose delegations lack an accepted link.
- Consequence: restored parents could publish with unlinked children while the refreshed path is correct.
- Evidence: scratch mutation replacing the restored assignment with `_ = restored`. `TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState` still passed. The fixture never produces restored sessions.
- Suggested evidence: a cross-machine case where the parent is restored, not refreshed.

### T4. Fixture shape leaves some oracles easier than the incident (severity 2, suggestion)
- Every unrelated rollout in `codexstoretest.WriteLargeStore` is created before the root. No unrelated conventional rollout falls inside the family window, so the sweep oracle `familyHeaders + non-date header` never has to include later-but-unrelated header reads or show they stay within the quarter. The irregular names sit in pruned old day directories (only the non-date file is reachable by the sweep), so the "unconventional names inside a listed directory are read" sweep rule is covered only by the exact-lookup oracle.
- A regression that over-read unrelated later rollouts, or mishandled an irregular name in a listed directory, would not move these totals.

### T5. Minor oracle and robustness notes (severity 1, suggestion)
- The header-cost model (4 KiB chunk, 64 KiB bound) is restated in `discovery_window_test.go` and `reconcile_large_store_test.go`. It mirrors the implementation, so a deliberate change must update three places; a silent drift fails loudly rather than escaping. Acceptable.
- `TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout` has an 8 s wall-clock bound against a 10 s lifetime. Measured whole test 2.7 s for three worker runs here, and the "lifetime exhausted" text check is the primary oracle, so flake risk is low but nonzero on a heavily loaded host.
- Manual `generate` non-regression is only a precondition step inside the CLI retention test (a full manual generation links all four children). There is no case showing manual generation still drops or replaces links from fresh source when an older ledger exists. `retainCaptureChildLinks` is only called under `--capture-auto`, so the leak risk is low.
- CLI-level contradiction clearing is covered by domain and capture tests only, not through `--capture-auto`.

## Hypotheses investigated

| Case | Disposition |
|---|---|
| Partial refresh keeping all links | Effective. Capture-level test checks registry snapshot, refreshed, published and unread-child snapshots against accepted values, partial diagnostic, coverage summary text, spawn events, two queueing variants. Author mutation M1 and my dedup mutation B killed. |
| Idempotence | Effective at three levels: domain second application on every table row plus the sibling-release regression, repeated partial capture pass (snapshots identical), repeated CLI run (ledger bytes identical). |
| Contradiction clearing | Effective: domain rows for fresh/accepted contradictions, capture task-name and explicit-session variants (links cleared, stays cleared on repeat). Warning is deliberately not asserted, per design Known limits. |
| Cross-machine `--capture-auto` | Effective for refreshed owners: real binary, real git commit, empty registry, unrelated Codex owner and other-harness fact preserved, unregistered children counted, partial diagnostic recorded (my mutation L killed). Restored path unprotected (T3). |
| Manual full linking | Covered as a precondition only (T5). |
| Large-store budget | Effective: exact byte and attempt totals from fixture sizes at max_sessions 4096 and 256, two passes, reservation ceiling (my mutation C, removing the ceiling, was killed), unused-share release, deferral and refund, deferred-first ordering in both key orders. |
| Window, tolerance and directory rules | Effective: unit boundaries at exactly and 1 ms beyond the tolerance, margin asserted as 14 h 5 s, year/month/day, CET/CEST and UTC-12/UTC+14 cases, non-v7, malformed and out-of-tree directories, zero window, plus sweep-level local-midnight tests. |
| Exact lookup | Effective: claim and byte oracles, archived same-name copy, duplicate stays ambiguous, mismatched metadata, empty store, live lookup unchanged (claims every rollout). |
| Header cache | Effective: same state cached, size change and mtime-only change re-read, refused read not cached, nil cache, small file costs its size. |
| Reservation, release, deferral | Effective, see large-store row. Authors' M5 and M6 are plausible; I did not rerun them. |
| Archived relocation and isolation | Effective: seven table cases with typed reasons, no adoption of other facts, non-not-found errors still fail, recovered row stays unassigned and can still be bound with its facts. |
| Worker isolation | Effective: real worker process, other ticket still publishes, row keeps facts, later archive recovery retries. Author mutation M4 killed. |
| `regroupArchiveFamily` retention | Not protected (T1). |
| Queued-path / sweep resume at the ceiling | Not protected (T2). |

## Changed existing oracles

- `TestReconcileCodexRecoversNestedSpawnWithoutChildCallbacks`: only fixture dates, directory and file names moved to agree with the UUIDv7 instants (2026-01-27T15:00:18.048Z). The old fixture put 2026-01 UUIDs in `sessions/2026/01/01`, which real Codex never writes and which the 14 h directory rule now correctly prunes. Assertions are untouched and the test still passes (verified). Not a weakening. The "directory disagrees with UUID" situation is covered elsewhere (window unit tests, local-midnight sweep tests).
- `TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities`: the old oracle `SourceBytes == allowance` held only because work equalled the cap. The new oracle derives the exact sum (every candidate header at whole-file cost, asserted below 4 KiB, plus the bound owner's extraction) at a tight and a double allowance, with all original assertions kept (attempts 1, no disclosure of unassigned IDs, five local rows). Stronger, not weaker. Verified passing.

## Fixture quality

- Self-contained: every store is under `t.TempDir()` with no `~/.codex`, ledger or network access. No `t.Skip`, `testing.Short`, permission tricks or domain mocks in any changed or added test (grep over the diff). The CLI tests drive the real built binary and real git.
- `codexstoretest` is a sound generator: stdlib-only, deterministic (seeded PCG), no production importer (`go list` over `cmd` and `internal`), and has its own invariants test (sizes equal stat, exact header size, UUIDv7 instants on the correct side of the root, machine-local day directory, each irregular name once). Residual points: it is a non-test package importing `testing` and `time/tzdata` (acceptable for test-support); the sparse-file approach assumes a sparse-capable filesystem (about 30 MB allocated at the default divisor 4, larger apparent size); see T4 for shape limits.

## Runtime proportionality

Proportionate. Verified here: domain, codex and codexstoretest packages 2.6 s; focused capture set (retention, archived, large-store, nested, unassigned-roots) 11.5 s; focused cmd retention and worker tests 7 s. Reported package deltas are about +34 s (capture), +71 s (cmd) and +31 s (race), against a baseline measurement fixture that cost about 74 s under race. The default divisor 4 keeps the large-store cases cheap, and divisor 1 is opt-in.

## Verification performed

- Reran focused tests at 0f73f759: domain, codex, codexstoretest; capture `TestReconcileRetains...`, `TestReconcileFreshContradicting...`, `TestReconcileArchivedUnassignedCodexChild`, `TestReconcileUnassignedRoots...`, `TestReconcileCodexRecoversNestedSpawn...`, `TestLargeStore...`; cmd `TestCaptureWorkerIsolates...` and `TestCaptureAutoKeepsCommittedChildLinks...`. All passed.
- Disposable mutations in a copy under the scratchpad (checkout untouched). The cmd tests needed the gitignored embedded `dist` copied into the scratch tree:

| Mutation | Result |
|---|---|
| A: regroup retention removed | survived the full capture package (T1) |
| B: retained-link dedup removed | killed |
| C: identification ceiling removed | killed (ceiling test) |
| D: queued-path cursor restore removed | survived (T2) |
| G: starts-sweep resume cursor removed | survived (T2) |
| J: CLI restored rewrite removed | survived the CLI retention test (T3) |
| L: CLI diagnostic not recorded | killed |

- Not rerun: whole package suites, author mutations M1 to M6.

## Gaps

- Race behavior, the integrated `make` gate and the whole-package `cmd` suite were not reexamined; I reuse evidence/TESTS.md for those.
- The divisor 1 incident-scale run was not repeated.
