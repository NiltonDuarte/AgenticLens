# FIX-21 re-review round 3: behavior (beh-opus-r3)

- **Run:** beh-opus-r3, phase re_review round 3, focus behavior and integration.
- **Requested profile:** Claude Opus 5.5 / medium. Runtime model ID reported by the harness: claude-opus-5-5. Usage and cost are not available to this run.
- **Reviewed revision:** `5d4df7ee` (frozen head). Corrections delta `86eb6dae..5d4df7ee`. Full change `133100bb..5d4df7ee`.
- **Verdict:** APPROVED for the behavior focus. There are two non-blocking observations (severity 2 and 1).

## Round-2 findings at 5d4df7ee

| ID | Status | Evidence |
|---|---|---|
| R2-01 | Resolved | `reconcileOwnerChildLinks` now pairs accepted position i with fresh position i in each key group. It refuses a group only when the fresh group is smaller or has no fresh delegation. Probe `mixed-groups` (interleaved untimed and timed groups, offset between sides) keeps `[U A B]` at the right positions. `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger` passes in a focused run: the registry stage fills children 1 and 2, the ledger stage keeps child 3 by position, two runs give byte-identical ledgers, and the latest reconciliation is partial with 1 unresolved child. The reconcile test's "second child re-read" subtest passes too. |
| R2-02 | Resolved | An untimed key is always `grouped`. An unconfirmed different fresh child becomes `movable`, which gives `kept_pending_confirmation`. Probe `untimed-unconfirmed` keeps `A` over `X`, reports `A` as unresolved, and is stable on the next run. |
| R2-03 | Resolved; the later loss is a recorded limit | When a key has no fresh delegation, its links get `ChildLinkRefusalUnmatched`. Probe `time-became-known` reports `{P A} unmatched` once and counts it as unresolved. On the next run nothing is reported, which matches the design Known limit "Refused and unmatched links are reported only once". |
| R2-04 | Resolved | `refused_ambiguous` was removed. Decisions now carry only `refused_contradicted` and `refused_held_by_sibling`. A smaller group yields `Refused` entries and no decisions (probe `shrink`: `decisions=[]`). |
| R2-05 | Resolved | `capture.go` appends ", so the reconciliation remains partial." only when `retention.Added > 0`. `TestCaptureAutoReportsOnlyClearedLinksWithoutPartial` passes. |
| R2-06 | Resolved | An overlay probe (no tree edits) held `BEGIN IMMEDIATE` on a fresh retention fixture and called `capture.RetainAcceptedChildLinks` with 1 ms, 20 ms, 200 ms and 2 s deadlines, three times each. At 2 s the error is a plain `SQLITE_BUSY`, and `IsRegistryContention` is true. At 20 and 200 ms it is `context deadline exceeded` joined with `SQLITE_BUSY`: `IsRegistryContention` is false and `errors.Is(err, context.DeadlineExceeded)` is true. At 1 ms it is a bare deadline. The CLI's `||` branch maps every one of these to `capturePublicationBusyExit` (75). The worker treats 75 as busy-only (`capture_worker.go:501`). There is no end-to-end CLI test; TESTS.md records this as a limit. |
| R2-07 | Resolved | `design.md` and `tasks.md` no longer name `AddRetainedChildLinks`, the exported `ReconciliationDiagnostic`, `ChildParentNotEvaluated` or `refused_ambiguous`. The design describes `Known`/`Observed`, the three evidence states, the owner-snapshot/family/generation invocation sites, the file-name deduplication hint and position alignment. The Known limits cover one-time refusal reporting and R2-P1. |
| R2-T1 | Resolved | The retry uses its own `retryCtx` with a 2 s timeout. |
| R2-T2, T3, T8 | Resolved | New domain cases: interleaved/offset groups, a contradicted member of a refused group, an unmatched key linked elsewhere, and a complete group with no decisions. CLI alignment end to end (above). `TestRetainAcceptedChildLinksReportsALinkKeptAndRefusedAsKept` protects the `delete(refused, kept)` guard. |
| R2-T4 | Resolved | `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure` (trigger-rejected insert) passes. |
| R2-T5 | Resolved | `TestHeaderSessionContextRejectsConflictingAncestry` passes. |
| R2-T6 | Resolved | `TestQueuedChildPathsResumeAtTheRefusedPath` passes. |
| R2-T7 | Resolved | The guard uses `go list -e`, so it no longer needs the gitignored frontend embed. |

## Regression hunt in the rewritten rule

Domain overlay probe (`go test -overlay`, package `domain`). For each case it applies the rule, re-applies it to its own result, and runs a realistic next pass in which the published result becomes the accepted ledger and the fresh read is unchanged.

| Case | Result | Assessment |
|---|---|---|
| grow (accepted 1, fresh 2 with the same key) | `[A ""]`, `retained_by_position` | Correct. |
| shrink (accepted `[A B C]`, fresh `[A ""]`) | `[A ""]`; B and C refused as `ambiguous`; A not reported | Matches the user decision. |
| dup-accepted (accepted `[A A]`, an R2-P1 artifact) | `[A ""]`; second position `refused_held_by_sibling` | No invented link. |
| mixed-groups (interleaved, offset) | `[U A B]` | Correct. |
| shift (accepted `[A B]`, fresh `["" A]` with A confirmed) | `["" A]`; A reported refused `held_by_sibling`; B dropped through `replaced` with no report | See OBS-1. |
| swap (accepted `[A B]`, fresh `[B ""]` with B confirmed) | `[B ""]`; B reported refused; A dropped without a report | See OBS-1. |

In every case, re-applying the rule to its own output gives the same result, and the next pass invents and moves no links. For double counting, `Kept`/`Refused` in the adapter are link sets, and refused links that are also kept are removed. `addUnresolvedChildLinks` deduplicates through `countedChildLinks`, so each link adds at most one to `PendingChildren` per request. The cursor, `.gz` admission and session-context code are unchanged in this delta (only their tests changed). The focused tests for them pass.

**Unchanged non-capture paths.** `domain.RetainAcceptedChildLinks` is called only from capture: owner snapshot `reconcile.go:896`, family annotation `reconcile.go:1092`, archive `archive_projection.go:108`, and generation through `capture.RetainAcceptedChildLinks`. The CLI calls it only under `--capture-auto` (`main.go:601`). Manual generation, OpenCode owners and owners with nil delegations do not reach the rule.

## New findings

### OBS-1 (severity 2, existing round-1 semantics now reachable inside groups; non-blocking)

**Location:** `backend/internal/domain/child_links.go`, the `ChildLinkReplaced` branch of `reconcileOwnerChildLinks`, and the `held_by_sibling` path of `refusedChildLinks`.

**Trigger:** A fresh read places a confirmed child at a different position of a key group than the accepted ledger did. For example, accepted `[A B]` and fresh `["" A]`, where A was re-read with this parent.

**Consequence:** The decision at position 1 is `replaced` (A over B). B is dropped from the ledger, and no refusal or unresolved entry names it. Position 0 reports A as `held_by_sibling` and counts `(P, A)` as unresolved, although A is linked on the owner and confirmed. That pass is recorded as partial for a pair that is present, while the lost pair B goes unreported. On the next pass nothing is reported and B is gone.

By construction, every `held_by_sibling` refusal names a child that another final delegation holds, so reporting it as unresolved is accurate only while that holder is itself unconfirmed.

**Reachability:** Low. Transcript order is stable, so fresh and accepted order agree when both come from the same parser. The mismatch needs a ledger from a differently resolving read, or the pre-existing R2-P1 task-name resolution.

**Disposition suggested:** Record it under Known limits next to R2-P1, or in the R2-P1 follow-up report the displaced accepted child (`decision.ChildSessionID` before replacement) instead of the held child. No change is required for this ticket's approved scope.

### OBS-2 (severity 1, comment accuracy)

**Location:** `backend/cmd/agentic-lens/capture.go:801`.

The comment "The bounded wait expires only while the registry is held elsewhere" is stronger than the evidence. The probe shows that a bare `context deadline exceeded`, with no SQLite cause, also takes the busy exit, and slow I/O could cause one too. `capture.IsRegistryContention`'s own documentation says an expired context "retains ordinary failure policy". The behavior is benign, because exit 75 only makes the worker retry later, and the user decision asks for exactly this mapping. Only the comment, or a cross-reference to the decision, could be clearer.

## Verification performed

- Read `domain/child_links.go` in full at `5d4df7ee`, plus the `capture.RetainAcceptedChildLinks`/`ChildLinkRetention` adapter, the CLI helper, the `main.go` call site, the worker busy handling and `IsRegistryContention`.
- Read the delta to `design.md`, `spec.md` and `tasks.md`, and grepped them for removed API names.
- Focused tests, all passing:
  - `go test -count=1 -run 'TestRetainAcceptedChildLinks|TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses|TestQueuedChildPathsResumeAtTheRefusedPath|TestHeaderSessionContextRejectsConflictingAncestry|TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger|TestCaptureAutoReportsOnlyClearedLinksWithoutPartial' ./internal/domain/ ./internal/adapters/capture/ ./internal/adapters/codex/ ./cmd/agentic-lens/`
- Overlay probes, kept in the run's scratch directory and not in the tree:
  - deadline/contention error classification (12 calls, all busy-mapped);
  - eight domain alignment cases with self-application and next-pass stability.

## Gaps

- No whole-suite or `-race` reruns; this run reuses TESTS.md "Candidate verification" as the brief instructs.
- No end-to-end CLI assertion of exit 75 on the retention deadline. The mapping is verified by a probe at the API boundary and by reading the code.
- The session-context, cursor and `.gz` admission product code is unchanged since round 2. It was checked only through the new focused tests, not re-traced.
