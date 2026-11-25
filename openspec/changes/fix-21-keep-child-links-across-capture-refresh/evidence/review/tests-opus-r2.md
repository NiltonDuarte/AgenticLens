# FIX-21 re-review round 2: test effectiveness (tests-opus-r2)

- **Run:** tests-opus-r2, phase re_review (round 2), focus reviewer-tests.
- **Reviewed revision:** `86eb6dae` (corrections delta `0f73f759..86eb6dae`; full change `133100bb..86eb6dae`).
- **Requested profile:** Claude Opus 5.5, medium effort. Runtime model metadata: claude-opus-5-5. Usage and cost not available to me.
- **Verdict:** CHANGES_REQUESTED (one reproduced flaky test; two small oracle gaps). Every round 1 test finding in this focus is closed.

## Round 1 findings in focus

| Finding | Status at 86eb6dae | Evidence |
|---|---|---|
| TEST-01 regroup retention | Closed | `TestReconcileArchivedRegroupKeepsAcceptedChildLink` asserts the link in both refreshed and published owners, count 1, and no child on the spawn event. M14 claim not rerun; the assertion targets exactly the deleted call. |
| TEST-02 CLI fresh plus retained | Closed | CLI case "some children read with the parent on the second run" asserts exact `{kept 2, pending 2}` on run 2, which a double count would turn into 3. |
| TEST-03 non-v7 fallback | Closed | `TestChildSweepWindowFromNonV7ParentTimestamp` asserts the exact byte total (passes under `-race` here). |
| TEST-04 RestoredSessions rewrite | Closed | `TestRetainAcceptedChildLinksRewritesRestoredSessions` asserts `[C1 ""]` on restored, with the refreshed C2 unchanged. |
| TEST-05 later unrelated rollouts | Closed by claim (M17); not re-run (large store). |
| TEST-06 minor | Closed | The idempotence test asserts exact `[C1 C2]` and its outcomes. There is an import guard. The worker test is now judged by outcome: I re-ran **M4** (isolated-row exclusion removed) against the loosened test, and it still fails with "native worker lifetime exhausted", so removing the 8 s bound lost no detection. |
| DATA-01 regression | Protected | **M7 reproduced**: registry links `["" "" "" ""]`. |
| BEH-01 regression | Protected | **M8c reproduced**: cursor moved to `""` and the brand-new sweep wrote `""`. The M8a equivalence argument holds by reading `children.go:493,560`. |
| DATA-02, DATA-03, DATA-04 | Protected | Domain, capture API and CLI cases assert exact counts and typed reasons. M9, M12 and M13 claims accepted; the assertions are exact. |
| BEH-02/ARCH-03 | Protected at the API boundary | `IsRegistryContention` is asserted. The CLI exit-code mapping is untested end to end, which is a documented limit. See T-R2-1 for this test's robustness. |
| ARCH-04 | Protected | **M11 reproduced**: `TestHeaderCacheAdmitsCompressedHeaderOnlyWithTheWholeBound` fails. |

## Changed domain oracles

I checked the seven renamed or flipped oracles against the approved design Steering "Ordinal group retention".
- The three former refusals are now `retained_by_position` carries: duplicate, larger fresh group, unknown time.
- The refusal table now holds three carried cases with no refusal.
- "Contradicted child in an ordinal group is skipped" now carries its sibling.

Each matches the approved rule. The refusal branch keeps independent cases: a fresh link set, a smaller group (timed and untimed) and held by sibling.

Spot-checked mutations, applied in a scratch copy:
- **M19** (ordinal branch disabled): 14 FAIL lines in the domain package. The capture two-pass test fails with `["" ""]`.
- Dropping the size check: "a smaller fresh group is refused" fails, with an index panic.
- Dropping the empty check: two refusal cases fail.
- Swapping the unknown-provenance reason: detected.
- Relabelling position outcomes as retained: detected.

**The loosened second-application check is justified.** I traced it: after an ordinal carry into a larger fresh group, f(x) is partly linked, so f(f(x)) goes to the refusal branch. That branch emits `refused_ambiguous` decisions but no `Refused` links, because every accepted child is linked. No production consumer reads those decisions for counting: `capture/reconcile.go:424-440` uses only `RetainedLinks`, `Refused`, `Cleared` and `UnresolvedLinks`. Comparing facts only is therefore the right oracle.

The same second application really happens in production. `capture.RetainAcceptedChildLinks` (CLI) runs on `RefreshedSessions` that capture already carried. Only the domain idempotence subtest covers it, not a CLI case (see gaps).

## New findings

### T-R2-1: the contention test is flaky under load (reproduced). Severity 3.

- **Location:** `backend/internal/adapters/capture/retain_child_links_test.go:224-253`, `TestRetainAcceptedChildLinksReportsRegistryContention`.
- **Classification:** test robustness.
- **Trigger:** `bounded` is a 2 s context created before the contended call. The contended open waits about 1 s, the constructor's busy allowance. The same `bounded` context is then reused, after a full `ReconcileAt`, for the retry's `RetainAcceptedChildLinks` (line 250).
- **Consequence:** on a loaded or `-race` host, the deadline expires before the retry and the test fails with `context deadline exceeded`. This is not a product defect.
- **Evidence:** `go test -race -count=3 -run TestRetainAcceptedChildLinksReportsRegistryContention ./internal/adapters/capture/` failed 1 of 3 with `retain_child_links_test.go:251: context deadline exceeded` (2.58 s). A further `-count=6` passed at 1.60 to 1.98 s against the 2.0 s deadline, so the margin is thin. Overall, 1 failure in 9 race runs.
- The evidence claim "no data race" for capture rests on a round 1 `-race` run that predates this file. The other new capture tests pass under `-race` here (focused run, 9.9 s).
- **Remedy (tester):** give the retry its own bounded context.

### T-R2-2: the ordinal placement oracle cannot tell group position from absolute index. Severity 2.

- **Location:** `backend/internal/domain/child_links.go:350` (`j := fresh[position]`); tests in `child_links_test.go` (`TestRetainAcceptedChildLinksOrdinalGroups` and the refusal table) and `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`.
- **Trigger:** mutation `j := i`, which uses the accepted absolute index with a bounds guard. It **survives** the whole domain package and both capture ordinal and interrupted tests.
- **Consequence:** in every fixture, the group occupies the same absolute positions on both sides. A regression that places a link on the delegation at the accepted absolute index, rather than at position i of the key group, goes undetected. Under strict append-only transcripts the two coincide. They differ as soon as the accepted and fresh sequences are not prefix-aligned, for example with an accepted ledger from another machine or version. The result would be a misplaced link on a delegation of another key.
- **Remedy (tester):** add one case whose group is interleaved with another key and offset between the two sides. One example: accepted `[A(C1) A(C2)]`, fresh `[B("") A("") A("")]`, expecting `["" C1 C2]`.

### T-R2-3: the contradicted-refusal filter is unprotected. Severity 2.

- **Location:** `child_links.go:443` (`evidence(child, parent) == ChildParentContradicted` in `refusedChildLinks.add`).
- **Trigger:** removing that condition **survives** the domain package.
- **Consequence:** in a refused group (a fresh link set, or a smaller fresh group), an accepted child that evidence contradicts would be reported as a refused link. That over-counts unresolved child evidence and keeps the attempt partial. The spec says contradicted links are not refusals.
- No case combines a refused group with a contradicted accepted child. The existing "contradicted child in an ordinal group" case goes through the carry branch, and held-by-sibling candidates are never contradicted.
- **Remedy (tester):** add a refused-group case with a contradicted member, expecting it absent from `Refused`.

Minor, not scored:
- The untimed refused branch records no `refused_ambiguous` decision, and adding one survives. That is unobservable to consumers.
- `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses` says "published owners" but asserts `RefreshedSessions`, not `result.Sessions`. The archived regroup test does check both.

## Verification performed

- Focused runs at the head:
  - `TestRetainAcceptedChildLinks*` (capture): PASS.
  - The contention test under `-race`: 1 failure in 9 runs.
  - The new capture and codex FIX-21 tests under `-race`: PASS.
- Mutations, each in a scratch copy of `git archive 86eb6dae backend` (with the embedded dist copied in for cmd). I restored each one by file copy and confirmed it with `diff -q`. Product and test files in the checkout were untouched.
  - Reproduced: M4 (against the new worker test), M7, M8c, M11 and M19.
  - New: absolute-index placement, which survives; dropping the size check, which is killed; dropping the empty check, which is killed; reason swap, which is killed; outcome relabel, which is killed; dropping the contradicted filter, which survives; recording untimed decisions, which survives and is unobservable.
- Read: `domain/child_links.go`, the domain test diff, `capture/retain_child_links_test.go`, the capture reconcile and archived test diffs, the CLI and worker test diffs, and the product diffs for reconcile, children, discovery_budget and context.

## Gaps

- I did not re-run M12 to M18 or the large-store tests (host load). I accepted them from TESTS.md where the assertion visibly targets the change.
- No CLI case covers the production double application, where capture already carried an ordinal group into a larger fresh group and the ledger retention is then applied. The domain idempotence subtest covers the facts.
- The ARCH-02 refactor (`exactSessionContext`/`conflictingNativeParent`) was checked for parity by reading only, which is outside this focus.
