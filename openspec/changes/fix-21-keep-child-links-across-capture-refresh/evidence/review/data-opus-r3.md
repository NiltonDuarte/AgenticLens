# FIX-21 data-integrity re-review (data-opus-r3)

- **Run:** data-opus-r3, re_review round 3. Focus: reviewer-data (data integrity and failure handling).
- **Requested profile:** Claude Opus 5.5 / medium. The runtime model ID is claude-opus-5-5. Effort, usage and cost are not visible to the reviewer.
- **Reviewed revision:** frozen head `5d4df7ee`. I read the corrections delta `git diff 86eb6dae 5d4df7ee`, and the full change is `133100bb..5d4df7ee`.
- **Verdict:** APPROVED for the data focus. I found no new defect. Two non-blocking observations are listed below.

## Round-2 findings in focus

| ID | Status at 5d4df7ee | Evidence |
|---|---|---|
| R2-01 / DATA-R2-01, first trigger (partly re-read group) | Resolved | `reconcileOwnerChildLinks` now pairs accepted position i with fresh position i whenever the fresh group is at least as large. An unread position keeps its accepted link (`retained_by_position`) and counts as unresolved. The adapter test `ambiguousGroupPartialPasses/second child re-read` checks the refreshed and published links over two passes, and it passes when run here. The property probe below finds no silent loss and no feedback drift. |
| R2-01 / DATA-R2-01, second trigger (two-layer CLI) | Resolved | The registry output `[C1,C2,_]` compared with the ledger `[C1,C2,C3]` now pairs position 2, so C3 is carried. `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger` passes. The two-layer probe confirms it. |
| R2-02 (untimed unique delegation, unconfirmed different child) | Resolved | An untimed group of one is paired, and an unconfirmed fresh child becomes `kept_pending_confirmation`. A domain table case covers it. |
| R2-03 (accepted key with no fresh counterpart) | Resolved | `unplaced[key] = ChildLinkRefusalUnmatched` is set, and it is reported only while no final delegation links the child. The one-time reporting is recorded under Known limits, as the user decided. |
| R2-04 (refusal decisions in complete groups) | Resolved | Unplaced groups emit no decisions, only `Report.Refused`. My probe check 5 finds no refused entry for a complete group, apart from `held_by_sibling`. |
| R2-05 (stderr says "remains partial") | Resolved, as the user specified | The suffix is gated on `retention.Added > 0`. See observation O2. |
| R2-06 (deadline maps to busy) | Resolved | `capture.go` maps `IsRegistryContention(err) \|\| errors.Is(err, context.DeadlineExceeded)` to `capturePublicationBusyExit`. This runs before any ledger write (main.go L601), so a busy exit leaves the published ledger untouched. A deadline from a slow but uncontended open also maps to busy. That code means retry, so no fact is lost. |
| DATA-01 (per-owner snapshot) | Still resolved | Retention at reconcile.go L896 is unchanged. Because the rule is idempotent under feedback, the carried first-commit snapshot used as the next baseline keeps re-reporting an unread child. |
| DATA-02 / dedup counting | Still resolved | The refreshed and restored `UnresolvedLinks()` go through the shared `countedChildLinks` set. `TestRetainAcceptedChildLinksReportsALinkKeptAndRefusedAsKept` shows that a link both kept and refused is counted once, as kept. |
| DATA-04 (Known precedence) | Still resolved | `evidence` checks fresh/observed first, then Known, then Accepted. A Known record without a parent is unobserved and does not fall through to Accepted. My probe models this order exactly and finds no discrepancy. |

## Investigation of the rewritten rule

I ran a randomized property probe through an overlay test (no tracked file changed). It covers 300,000 cases: four key shapes (timed, untimed, a different description, and a nil versus typed agent type), groups that are appended, truncated (smaller) or re-keyed (unmatched), and fresh links that are random, equal to accepted, or empty. Each child gets confirmed, contradicted, accepted-confirmed or Known-without-parent evidence. Zero failures for:

1. **Invented or misplaced links.** Every output child is either the fresh child at that delegation or the accepted child at the same group position.
2. **Contradicted children.** No contradicted child appears in the output.
3. **Idempotence of facts and reports.** `f(f(x,A),A)` equals `f(x,A)`, including the delegations and `Report.Refused`.
4. **Feedback stability.** `f(x, f(x,A))` produces the same delegations as `f(x,A)`, and when nothing is refused the unresolved count is also unchanged.
5. **No silent loss.** Every accepted link that is not contradicted is in the output, in `Refused`, or at a position replaced by a confirmed fresh child.
6. **Refusals only where expected.** `Refused` holds only links from smaller or unmatched groups, plus `held_by_sibling`.
7. **No new duplicates.** No child is duplicated beyond the duplicates the fresh read already had.

A two-layer probe (200,000 cases) applies registry retention, then ledger retention, as the CLI does. It finds no invented link. It surfaced observation O1.

## Observations (non-blocking, no finding)

- **O1, registry and ledger disagree at one position (severity 2, by design).** Suppose the registry snapshot and the committed ledger hold different unconfirmed children at the same group position, for example registry C4 and ledger C6, and the fresh link is empty. The registry layer carries C4. The CLI ledger layer then sees C4 as an unconfirmed different fresh child and keeps C6 (`kept_pending_confirmation`). The published ledger shows C6, and C4 appears in neither the output nor the CLI report.
  - This follows the documented precedence: the committed ledger is the CLI's accepted baseline, and the registry acts only as Known evidence. C4 stays in the registry snapshot, and the registry pass already counted it as unresolved, so the pass is still partial and nothing committed is lost.
  - Two unconfirmed children at the same transcript position can only come from a task-name fallback (R2-P1) or a divergent cross-machine read.
  - Record it only if the parent wants the precedence written in the design.
- **O2, stderr wording.** When `Kept > 0` but `Added == 0`, because the registry pass already counted those links, `PendingChildren > 0` and the attempt is partial, yet the line ends with "." rather than "remains partial". The line still prints the unresolved count, and the user decision named `Added > 0` explicitly, so I am not raising this as a defect.

## Verification performed

- Read `domain/child_links.go` at head in full, the CLI `retainCaptureChildLinks` delta, `capture.RetainAcceptedChildLinks` / `addUnresolvedChildLinks`, and the retention call sites at reconcile.go L896 and L1092 and in archive_projection.go. I also read the design sections "Position alignment", "R2 implementation sync" and "Known limits", and the spec requirements.
- Focused runs at 5d4df7ee, all passing:
  - `go test ./internal/domain/ -run ChildLink`
  - `go test ./internal/adapters/capture/ -run 'RetainAccepted|Ambiguous|PartlyReRead|QueuedChildPaths|Interrupted|Contention'` (55 s)
  - `go test ./cmd/agentic-lens/ -run CaptureAuto`
- Overlay probes. Their files are under the scratchpad subdirectory `data-opus-r3/` (`probe_test.go`, `probe2_test.go`, `overlay.json`). They ran with `go test -overlay <scratch>/data-opus-r3/overlay.json ./internal/domain/ -run 'TestProbeR3'`.
- I did not rerun full package suites or `-race`. I reused `evidence/TESTS.md` as the brief instructs.

## Gaps

- I did not exercise O1 end to end through `generate --capture-auto`. It comes from the domain probe and from reading the CLI order.
- I did not probe cross-machine pending-snapshot recovery or OpenCode paging, which are unchanged by the delta.
