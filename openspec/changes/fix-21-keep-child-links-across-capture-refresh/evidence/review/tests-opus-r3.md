# FIX-21 re-review round 3: test effectiveness

- **Run:** tests-opus-r3, phase re_review (round 3)
- **Requested model / effort:** Claude Opus 5.5 / medium. The runtime reports the model as claude-opus-5-5. Usage and cost are not available.
- **Reviewed revision:** `5d4df7ee`. The corrections delta is `86eb6dae..5d4df7ee`. Product and test files were read at HEAD, with a clean working tree.
- **Verdict:** APPROVED for the test-effectiveness focus. Two minor findings (severity 1) remain, and neither blocks delivery.

## Round 2 test findings at 5d4df7ee

| ID | Status | Evidence |
|---|---|---|
| R2-T1 | Resolved | The retry now uses its own 2 s `retryCtx` (`retain_child_links_test.go`). I ran the contention test and the diagnostic-failure test with `-race -count=10` in the scratch copy. All 20 runs passed and no data race was reported. |
| R2-T2 | Resolved | There are two interleaved and offset cases in `TestRetainAcceptedChildLinksAlignsGroupsByPosition`. I reproduced M21b, which uses the absolute accepted index in place of the group position. Both cases fail with exactly the reported links: `["C1" "C2" ""]` and `["C1" "C3" "C2" ""]`. |
| R2-T3 | Resolved | The case "a contradicted member of a refused group is not reported" protects this. I reproduced M24, which removes the contradicted filter in `add`: that case fails. |
| R2-T4 | Resolved | `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure` uses a real SQLite trigger, so no domain behavior is mocked. I reproduced M28, which discards the `RecordDiagnostic` error. The test fails with "got `<nil>`". |
| R2-T5 | Resolved, by reading | `TestHeaderSessionContextRejectsConflictingAncestry` goes through the real `HeaderCache.ReadHeader` and covers four ancestry shapes. I did not reproduce M29. |
| R2-T6 | Resolved, by reading | `TestQueuedChildPathsResumeAtTheRefusedPath` asserts the cursor value and checks that the next one-read pass reads the refused path. I did not reproduce M30. |
| R2-T7 | Resolved | My scratch copy (`git archive` of `backend/`) has no `web/dist`. The guard passes there. I reproduced M18 by adding a production import into `cmd/agentic-lens`. The guard then fails and names `agenticlens/cmd/agentic-lens`. |
| R2-T8 | Resolved | There is a CLI alignment end-to-end case (`TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger`). It runs twice and checks ledger byte identity, the partial attempt and `PendingChildren == 1`. I reproduced M31, which removes the `delete(refused)` guard. The test reports `Kept:1 Refused:1`. The `refused_ambiguous` decision no longer exists, so that part of the finding is moot. |

R2-06 still has no test. That is recorded as a limit in TESTS.md, because `retainCaptureChildLinks` creates its context internally. I verified the mapping by reading it.

## The two changed domain oracles

Both renamed cases in `TestRetainAcceptedChildLinksReportsRefusedLinks` match the "Position alignment" steering decision. In each, the unread position keeps its accepted link, nothing is refused, and the unresolved set is correct:
- In "an unconfirmed fresh link in a group keeps the accepted link at its position", C1 and C2 are unresolved.
- In "a partly re-read group keeps the unread position", only C2 is unresolved.

They are not weakened oracles. Each one asserts the new contract exactly, and mutations M20 and M25 make them fail. I reproduced M25: six subtests fail, including "a partly re-read group keeps the unread position".

## Mutation spot checks I ran

Each mutation was applied in `scratchpad/tests-opus-r3/backend`, which is a `git archive` of `5d4df7ee`. The domain package was run in full for the domain mutations; capture mutations used focused `-run`.

| Mutation | Result |
|---|---|
| M21b, M23, M24, M25, M28, M31, M18 without `web/dist` | All fail as TESTS.md reports |
| MX1: in a group, label a kept-pending link as `retained_by_position` | Killed by the untimed-unique case |
| MX2: report `unknown_provenance` as `ambiguous` | Killed: "a smaller fresh group with unknown invocation time is refused" |
| MX3: remove the `!linked` filter for unplaced links | Killed: 2 subtests |
| MX4: `grouped := false` | Killed: 5 subtests |
| MX5: `grouped` ignores an unknown invocation time | Killed: "unknown invocation time is carried by position" |
| MX6: run the sibling-release fixpoint only once | Killed: `TestRetainAcceptedChildLinksIsIdempotentAfterSiblingRelease` |
| MX7: align a smaller fresh group instead of refusing it | Killed: "a smaller fresh group is refused" |

## Regression hunt in the rewritten rule

I wrote a throwaway randomized property probe in the scratch copy and deleted it afterwards. It ran 200,000 cases with:
- keys that are timed, untimed or of a second key, with 0 to 4 delegations on each side;
- children whose parent is confirmed, contradicted or unobserved, in the accepted and fresh records.

All of the following properties held:
- **Idempotent:** feeding the result back as fresh input leaves the links unchanged.
- **No invented link:** a link the rule places always comes from an accepted delegation with the same description and invocation time.
- **No new duplicate:** the rule never links a child more times than the fresh read did.
- **Refusals exclude linked children:** no `ambiguous`, `unknown_provenance` or `unmatched` refusal names a child that the result links.
- **No silent loss:** every accepted, uncontradicted child that is not re-read is linked, unresolved, or covered by a `replaced` decision on its key.

One observation is not a FIX-21 regression. A `replaced` decision overwrites `ChildSessionID` with the fresh child, so the report does not name the accepted child that was replaced. That behavior was the same at `86eb6dae`, and replacement by a confirmed child is by design.

## New findings

**T3-1 (severity 1, test gap). The positive "remains partial" wording is unasserted.**
- **Location:** `backend/cmd/agentic-lens/capture.go` (R2-05 branch) and `capture_child_links_test.go`.
- **Gap:** only the cleared-only case checks the wording, and it asserts that "remains partial" is absent. No test asserts that the phrase appears when `Added > 0`. The alignment case checks "Capture kept 1" and the partial attempt, but not the suffix.
- **Consequence:** inverting the condition, or always printing ".", would pass every test. The partial state itself is still protected through the recorded diagnostic.
- **Fix:** add `strings.Contains(out, "remains partial")` to the alignment CLI case.

**T3-2 (severity 1, stale test comment).**
- **Location:** the idempotence block in `TestRetainAcceptedChildLinksReportsRefusedLinks`, `backend/internal/domain/child_links_test.go` (around line 543).
- **Gap:** the comment says a filled ordinal group "is then partly linked, so it is refused as a whole with nothing left to report". Whole-group refusal no longer exists. The assertion itself is still correct, because it compares facts only.
- **Consequence:** the comment misleads readers about the current rule.

**Minor bookkeeping.** The TESTS.md mutation table skips M27. R2-06 has no mutation, so the number is presumably unused, but the gap is not explained.

## Verification performed

- Read the corrections delta, the domain rule, the capture `RetainAcceptedChildLinks`, the CLI `retainCaptureChildLinks`, and every test added or changed in the delta.
- `go test -count=1 ./internal/domain/`: PASS (scratch copy).
- `go test -race -count=10 -run 'ReportsRegistryContention|ReturnsDiagnosticWriteFailure' ./internal/adapters/capture/`: PASS, 20 of 20.
- `go test -race -count=5` on the domain alignment and refusal tests: PASS.
- Import guard without `web/dist`: PASS, and M18 is detected.
- Mutations: the 7 reproduced and 7 new ones listed above.

## Gaps

- I did not rerun whole package suites, as the brief instructed. The full CLI suite, the capture suite under `-race` and the large-store run are reused from TESTS.md.
- I did not reproduce M20, M22, M26, M29 or M30. They are verified by reading the tests only.
- R2-06 (busy exit on the deadline) has no automated test; it is an accepted limit.

## Cross-focus note for the parent

The CLI maps any `context.DeadlineExceeded` from the retention call to the busy exit. The code comment says the bounded wait expires only during contention. A deadline caused by slow I/O would be reported as busy rather than as a failure. This is low impact and belongs to the behavior focus.
