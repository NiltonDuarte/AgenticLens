# FIX-21 test-effectiveness re-review: tests-sonnet-r2

- Run: FIX-21 re_review run tests-sonnet-r2, focus reviewer-tests, requested profile Claude Sonnet 5.5 / high (runtime model self-report: claude-sonnet-5-5; effort not independently verifiable).
- Reviewed revision: 86eb6dae (frozen head), corrections delta `0f73f759..86eb6dae`, base 133100bb. Working tree clean at start.
- Independence: other round-2 reports (`*-r2.md`) were not read. Round-1 REVIEW.md and my own round-1 report were read as allowed.
- Write scope: this report only. All mutations ran in scratch copies of `backend/` under the session scratchpad (checkout untouched, no `~/.codex`, `~/.agenticlens` or live ledger access).

## Verdict: CHANGES_REQUESTED (low severity, test gaps only; no product defect found)

Every round-1 finding in my focus is closed or correctly dispositioned, and every regression test I mutated is killed by a plausible fault. The ordinal group retention is well protected at domain level and at the capture boundary. I found four residual or new test gaps, all severity 3 or below, none a blocker on its own. Three are surviving mutants in code this change introduced or touched. I recommend closing or consciously accepting them before delivery.

## Round-1 finding verification (my focus)

| ID | Status at 86eb6dae | Evidence |
|---|---|---|
| TEST-01 regroup retention | Closed | `TestReconcileArchivedRegroupKeepsAcceptedChildLink` asserts refreshed and published links, count once, no link on the spawn event. My mutation B14 (`owners = retention.Sessions` removed) fails it. |
| TEST-02 CLI fresh-plus-retained | Closed | CLI case "some children read with the parent on the second run" asserts exact per-run counts (4/4, then 2/2); passes when run (focused, 10 s). |
| TEST-03 non-v7 window | Closed | `TestChildSweepWindowFromNonV7ParentTimestamp` with exact byte total. Author M15 not rerun; test reads as protective (exact total, earlier rollouts costed 0). |
| TEST-04 restored rewrite | Closed | `TestRetainAcceptedChildLinksRewritesRestoredSessions`. My mutation M16 (restored assignment removed) fails it. |
| TEST-05 fixture shape | Closed | Large-store spec adds six later unrelated rollouts and two in-window irregular names; oracles use `sweepOnlyHeaders()` and a derived sum, not a literal. `TestWriteLargeStoreInvariants` guards the generator. |
| TEST-06 minor set | Closed, with one new caveat (N3) | Exact `[C1 C2]` asserted with outcomes. Import guard exists. Worker test now judged by outcome (see below). Manual missing-child and CLI contradiction cases exist and pass. |

Fix regression tests (mutations were mine unless marked author):

| Finding | Test | Result |
|---|---|---|
| DATA-01 | `TestReconcileInterruptedAnnotationKeepsRegistryLinks` | M7 (per-owner retention removed, `compact.Sessions` stored): killed. The trigger fires on the annotation write only (OLD input 41 and NEW input 41), so the refreshed-snapshot commit is the interrupted state. Sound seam. |
| BEH-01 | `TestResumedSweepAtTheCeilingKeepsItsFileCursor`, `TestBrandNewSweepAdmittingNothingWritesNoCursor` | S1 (no-row branch removed): killed by the brand-new test. S2 (`last = previous` replaced by `_ = previous`; round-1 mutation G): now killed by the resumed test. Also 12 repeated runs of the three sweep tests: stable. Author M8a (`last := cursor` alone) is a genuine equivalent mutant: the `last == ""` continue branch preserves the saved row. |
| BEH-02 / ARCH-03 | `TestRetainAcceptedChildLinksReportsRegistryContention` | Partly protective, see N1. |
| DATA-02 | Domain `...ReportsRefusedLinks`, capture `...CountsRefusedLinksOnce...`, CLI ambiguous case | E3 (swapped timed/untimed reason): killed. Counting dedup mutants killed (B5, R3). See N2 for the lost decision-outcome oracle. |
| DATA-03 | CLI "queued path of an unregistered child" | M12 (name-based dedup removed): killed (5 instead of 4). |
| DATA-04 | domain `...PrefersKnownChildEvidence`, capture `...UsesRegistrySessionsAsChildEvidence` | B13 (`Known` dropped): killed, two subcases. |
| ARCH-04 | `TestHeaderCacheAdmitsCompressedHeaderOnlyWithTheWholeBound` | M11 (gz admission by compressed size): killed. The fixture is sound: compressed size below allowance, allowance one byte under the decompressed cost, and an "almost 64 KiB" probe. |
| ARCH-02 | none new | Not directly guarded, see N4. |

## Ordinal retention tests and the seven changed domain oracles

I checked the rationale in TESTS.md against the diff. The three main-table rows and their three refusal-table twins changed from refusal to `retained_by_position`; the contradicted-child row now carries the uncontradicted sibling. These are exactly the cases the approved design flipped (design Steering "Ordinal group retention"), the new expectations match the rule text, and the refusal branch keeps independent rows (fresh link in group, smaller group timed and untimed, held by sibling). None weakens an assertion without a reason; the replaced oracles are stronger (outcome, evidence, delegation index, unresolved list).

Domain mutations, all against `go test ./internal/domain/`:

| Mutation | Result |
|---|---|
| A: ordinal branch disabled (`if true`) | killed (3 tests) |
| D1: `<` to `<=` on group size | killed |
| D2: emptiness condition removed (carry over a set fresh link) | killed by the refusal table only ("a fresh link set in the group refuses the whole group"); the ordinal test does not cover it, acceptable |
| D3: position reversed | killed |
| E1: empty accepted child not skipped | killed by the repeated-pass subtest only (the "no link is invented" subtest checks links, not decisions; adequate because links are what ship) |
| E2: held-by-sibling guard removed | killed (3 tests) |
| E3: refusal reason swapped | killed |
| E4: per-delegation `refused_ambiguous` decision made dead code | **survives** (N2) |

The loosened second-application check in the refusal table (facts only) is justified: the main table already uses the same standard, and a second application after an ordinal carry legitimately refuses a partly linked group with nothing left to report. The property that matters across passes (the carried links stay and are reported every time) is asserted separately by the domain "repeated pass" subtest and by `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`, which A (ordinal disabled) and R4 (`retained_by_position` dropped from `RetainedLinks`) both kill. Nothing identified requires restoring the stricter comparison.

Capture-level checks: A fails `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`. The CLI end to end case for an ordinal carry does not exist (the CLI ambiguous case is a smaller-group refusal), but the carry runs through the same `RetainAcceptedChildLinks` domain function with ledger accepted facts, so I rate that only a suggestion (S1).

## Worker test robustness

`TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout` now has no wall-clock assertion. The harness allows 30 s against the policy's 10 s lifetime and fails on a non-zero exit or "lifetime exhausted". Mutation W1 (`ErrOwnerUnavailable` no longer isolated) made the worker spin to its lifetime and fail the test (11.3 s total), so the oracle is the worker's own outcome and is load-independent apart from the 10 s worker lifetime. Passes unmutated (focused cmd run: 10 s for three cmd tests). Residual: a legitimate pass slower than the 10 s worker lifetime on a saturated host would still fail it, which is unchanged policy, not a weakness of this test.

## New and residual findings

### N1. A failed diagnostic write inside `RetainAcceptedChildLinks` is unprotected (severity 3, introduced)
- Location: `backend/internal/adapters/capture/reconcile.go`, final lines of `RetainAcceptedChildLinks` (`errors.Join(store.RecordDiagnostic(...), store.Close())`); test `retain_child_links_test.go` `TestRetainAcceptedChildLinksReportsRegistryContention`.
- Expectation: BEH-02 says contention while recording the partial diagnostic ends generation with the busy code, and nothing is published as complete.
- Trigger: the registry writer is taken after the project registry opens and before `RecordDiagnostic`, or the insert fails for any reason.
- Consequence: a refactor that swallows or retypes the `RecordDiagnostic` error leaves the CLI publishing a ledger whose reconciliation diagnostic still says completed, with no test noticing.
- Evidence: the test's `BEGIN IMMEDIATE` holder makes `OpenProjectContext` itself fail (database is locked, SQLITE_BUSY); mutation M10 (retype the open error) is killed, but M10b (retype the `RecordDiagnostic` error) and M10c (discard the `RecordDiagnostic` error, `_ = store.RecordDiagnostic(...)`) both survive the test. The CLI mapping to the busy exit is also verified by reading only (acknowledged in TESTS.md Limits).
- Suggested evidence: a registry trigger on the diagnostics insert (the DATA-01 test already uses this technique) asserting a returned error, and that a contention-coded abort surfaces through `IsRegistryContention`. This does not need a new product seam.

### N2. The per-delegation `refused_ambiguous` decision has no assertion left (severity 2, introduced by the oracle change)
- Location: `domain/child_links.go` `reconcileOwnerChildLinks` (the `matches[j] = &candidate{... ChildLinkRefusedAmbiguous ..., ChildParentNotEvaluated ...}` branch), and `domain/child_links_test.go`.
- Expectation: the report lists a decision per refused timed delegation (`ChildLinkRefusedAmbiguous`, evidence `not_evaluated`), as the enum and report doc say. Round 1 asserted it through three table rows that the ordinal rule turned into carries.
- Trigger: none needed; it is a static coverage hole. No test (`grep` for `RefusedAmbiguous` finds none) asserts these decisions or the `ChildParentNotEvaluated` constant.
- Consequence: the decisions can be dropped or changed silently. No current production consumer reads them (retention counts only `Refused` and unresolved links), so impact is a lost audit trace, not user-visible behavior.
- Evidence: mutation E4 (decision creation made dead code) passes the whole domain package.
- Suggested evidence: add the expected outcome list to the "fresh link set" and "smaller group" rows of `...ReportsRefusedLinks` (one extra field), or drop the dead decision if it is not meant to be reported.

### N3. The new import guard needs the gitignored embedded frontend (severity 2, introduced)
- Location: `codex/codexstoretest/import_guard_test.go` (`go list ./...` run from the module root).
- Trigger: `go test ./internal/adapters/codex/...` on a checkout without `backend/internal/adapters/web/dist` (gitignored; created by the `web-assets` step).
- Consequence: the guard fails with `go list ./...: exit status 1` before it checks anything, so a codex-only run, or any CI step that runs these packages before the frontend build, fails on an unrelated precondition.
- Evidence: copy of `backend/` with `dist` removed: the test fails exactly that way; with `dist` present it passes. The canonical `make` target orders `web-assets` before backend integration tests, so the standard gate is unaffected.
- Suggested evidence: use `go list -e` and fail only if the guarded import appears, or accept the dependency and note it.

### N4. Parity of the unified exact-metadata rule on the header path is unguarded (severity 3, pre-existing gap now touched by ARCH-02)
- Location: `codex/discovery_budget.go` `headerFromMetadata` (`header.ConflictingParent = conflictingNativeParent(metadata)`), `codex/context.go` `exactSessionContext`.
- Expectation: ARCH-02 merged two copies of the ancestry-agreement and required-field rules; behavior must not drift.
- Trigger: a child header whose direct parent and subagent `thread_spawn` parent disagree, read through the header cache (`RolloutHeader.SessionContext`, used by `validateReadOwner` and the sweep).
- Consequence: a regression that stops flagging the conflict on the header path is invisible to tests; the exact path (`readSessionContext`) is guarded.
- Evidence: mutation G2 (conflict flag forced false in `readSessionContext`) is killed by `TestNativeSessionContextKeepsChildIdentityAndStructuredParent`. Mutation G1 (`header.ConflictingParent = false`) survives the entire `codex`, `codexstoretest`, `live` and `capture` package runs (the only failures are the environment-only `live` wire fixtures and `TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject`, both failing unmutated in the copy because it lacks `web/` and `scripts/`). `grep` finds no test of `ConflictingParent` or `conflicting native parent` on this path. The same gap existed before the delta, so I classify it pre-existing, but the delta rewrote that line.
- Suggested evidence: one `HeaderCache.ReadHeader` plus `SessionContext` case with conflicting ancestry spellings, mirroring the existing exact-path test.

### N5. Queued-path cursor restore at the identification ceiling is still unasserted (severity 3, residual of my round-1 T2, introduced)
- Location: `capture/children.go`, `recoverChildren` queued-path loop (`saveChildCursor(... previous, true)` on `ErrCaptureBudget`).
- Expectation: "identification stops at its ceiling with discovery pending", and the unread candidate comes first in the next pass (comment in code).
- Trigger: identification reaches its share in the middle of the `pending_children` list (queued child paths).
- Consequence: the unread queued path is skipped until the cursor wraps, delaying that child's discovery. Bounded, not data loss.
- Evidence: mutation Q1 (restore replaced by `_ = previous; break`) passes the full capture package (117 s; only the environment failure). `grep` finds no test touching `queuedCursor` or `pending_children`. Round 1's G (starts sweep) is now killed (S2) but D (this path) was not addressed, and REVIEW.md maps BEH-01 only to the starts sweep.
- Suggested evidence: a queued-path variant of `TestResumedSweepAtTheCeilingKeepsItsFileCursor` asserting the "queued" cursor equals the last admitted path.

### Suggestions (severity 1)
- S1: no end-to-end CLI case carries an ordinal group (covered by domain and capture tests through the same function).
- S2: `for link := range kept { delete(refused, link) }` in `capture.RetainAcceptedChildLinks` can be removed (mutation R2) with every focused test passing. A pair appearing both kept and refused in the refreshed and restored reports needs the same parent in both lists, which I could not reach in a real flow; the effect would be a one-off inflated stderr/`Refused` count, since unresolved counting dedups separately.

## Fixture and test-hygiene observations
- No `t.Skip`, `testing.Short`, or domain mocks in the added tests; fixtures under `t.TempDir()`. The CLI tests drive the real binary and real git.
- `TestResumedSweepAtTheCeilingKeepsItsFileCursor` orders its parents by `IdentityKey` to match the registry's sweep order; 12 consecutive runs with fresh temp paths were stable.
- The retention and sweep tests need no sleeps or wall-clock bounds.

## Verification performed
- Read: brief, persona and shared context, REVIEW.md, TESTS.md, my round-1 report, changed product files (`domain/child_links.go`, `capture/reconcile.go`, `children.go`, `archive_projection.go`, `codex/context.go`, `discovery_budget.go`, `discovery.go`, `cmd/agentic-lens/capture.go`, `main.go`) and the changed or added tests in the corrections delta.
- Focused runs in the unmutated scratch copy: domain package; `TestResumedSweep...`, `TestBrandNewSweep...`, `TestChildSweepWindowFromNonV7...` (12 repetitions); cmd `TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState`, `TestManualGenerateLeavesLinkToMissingChildEmpty`, `TestCaptureWorkerIsolates...`. All passed.
- Mutation table (scratch copies; killed unless noted): A, D1, D2, D3, E1, E2, E3 (domain, killed); E4 (survived); B14, B7, B13, B5, R3, R4, M16 (capture, killed); M10 (killed), M10b and M10c (survived); R2 (survived); M12 (cmd, killed); M11 (codex, killed); S1 and S2 (killed); Q1 (survived); G1 (survived), G2 (killed); W1 (worker, killed).
- Not done: whole-package suites except the capture package (Q1 and G1, 117 s and 242 s), `-race`, the incident-scale divisor 1 run, and the integrated `make` gate. I reuse TESTS.md for those. The environment-only failures (`TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject`, the `live` wire fixtures) occurred in scratch copies lacking `scripts/` and `web/` and are not attributable to the change.

## Gaps
- N1 and N5 could be closed with registry-level tests using existing seams; no product change is needed. N2 and N4 are one assertion each. N3 is a one-line choice (use `go list -e` or accept).
- I did not verify author mutations M1 to M6, M8c, M9, M13 beyond my own variants, nor M15, M17, M18, M19 individually (A covers M19's effect).
- No effort level was verifiable at runtime.
