# FIX-21 review round 3: architecture (arch-opus-r3)

- **Run:** arch-opus-r3, phase re_review round 3, focus architecture and maintainability.
- **Requested profile:** Claude Opus 5.5 / medium. Runtime model ID claude-opus-5-5; effort not independently observable.
- **Reviewed revision:** frozen head `5d4df7ee`. The corrections delta is `86eb6dae..5d4df7ee`.
- **Verdict:** CHANGES_REQUESTED. One new severity-3 finding, A3-01, which can be fixed in the domain rule or recorded as a limit. All round-2 findings in this focus are resolved.

## Round-2 findings

| ID | Status at 5d4df7ee | Evidence |
|---|---|---|
| R2-01 | Resolved | One alignment path in `reconcileOwnerChildLinks` (`domain/child_links.go`): accepted position i is paired with fresh position i for every key in `priorOrder`. Groups are refused only when the fresh group is smaller (`ambiguous` or `unknown_provenance`) or has no fresh delegation (`unmatched`). The `alignmentNone/Unique/Position/Refused` enum and the all-empty precondition are removed. |
| R2-02 | Resolved | An untimed unique key goes through the same per-position rules. An unconfirmed different child is `movable`, so it becomes `kept_pending_confirmation`. The domain table case "an untimed unique delegation keeps the accepted link against an unconfirmed child" covers this. |
| R2-03 | Resolved | `case len(fresh) == 0` returns `ChildLinkRefusalUnmatched`. `refusedChildLinks` reports the link only while no final delegation links the child. The domain table case covers this, and the one-time reporting is under Known limits. |
| R2-04 | Resolved | `ChildLinkRefusedAmbiguous` and `ChildParentNotEvaluated` are deleted. No code path emits a decision for a refused group. Neither name is left in the backend or the active spec or tasks. The only mentions are the history in design.md's "R2 implementation sync". |
| R2-05 | Resolved | `capture.go:810-814` appends ", so the reconciliation remains partial." only when `retention.Added > 0`. `TestCaptureAutoReportsOnlyClearedLinksWithoutPartial` covers this. |
| R2-06 | Resolved, with a nit (A3-05) | `capture.go:802` maps `errors.Is(err, context.DeadlineExceeded)` to `capturePublicationBusyExit`. No test exercises this `os.Exit` branch. |
| R2-07 | Mostly resolved; residual in A3-04 | The design "Pure domain rule", "Invocation" and "Coverage" sections now name `ChildLinkInput` with `Known`, per-owner application, the single `capture.RetainAcceptedChildLinks` entry point and the file-name deduplication hint. No live text names `AddRetainedChildLinks` or `ReconciliationDiagnostic`. |
| R2-T1 to R2-T8 | Present (the tests focus owns their adequacy) | The retry gets its own `retryCtx` (T1). Interleaved and offset group cases exist (T2). "a contradicted member of a refused group is not reported" (T3). `TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure` (T4). `TestHeaderSessionContextRejectsConflictingAncestry` (T5). `TestQueuedChildPathsResumeAtTheRefusedPath` (T6). `go list -e` in the import guard (T7). `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger`, `...ReportsALinkKeptAndRefusedAsKept` and "a complete group emits no refusal decisions" (T8). |

## New findings

### A3-01 (severity 3): a duplicated fresh child replaces an accepted link without any report

- **Classification:** regression in reporting introduced by position alignment. It interacts with the pre-existing limit R2-P1, which is behaviour and data cross-focus.
- **Location:** `backend/internal/domain/child_links.go`, the per-position switch in `reconcileOwnerChildLinks`, at `case current != "" && evidence(current, parent) == ChildParentConfirmed`.
- **Trigger:**
  - An accepted ledger links a same-key group as `[A, B]`. That is possible when an earlier full read resolved each delegation by `agent_id`.
  - A bounded pass reads only child B.
  - `resolveSourceChildren` falls back to `ChildReference` for the unread `agent_id`, so both delegations point at B (R2-P1). The fresh group is `[B, B]`.
- **Consequence:**
  - At position 0, B is confirmed and differs from A, so the rule records `replaced`. A disappears from both the ledger and the registry snapshot.
  - Nothing goes into `Refused` or `UnresolvedLinks`, so `PendingChildren` does not count A, and the attempt can be recorded as complete.
  - At 86eb6dae the same input refused the group and reported A as `ambiguous` unresolved evidence. The final links were the same, but the loss was visible.
  - The `replaced` decision also overwrites `ChildSessionID` with the fresh ID (see A3-02), so the decision is recorded as "replaced B with B" and A's ID is gone from the report.
  - On the next pass (full read, where both share an agent path and neither resolves), `[B, ""]` is published. A is permanently lost.
- **Evidence:** a probe through an overlay test (`go test -overlay`, scratch only, product tree untouched):
  - At 5d4df7ee: `links=["B" "B"] decisions=[{ChildSessionID:B FreshChildID:B Outcome:replaced Evidence:confirmed Delegation:0}] refused=[] unresolved=[]`.
  - The same input with 86eb6dae's `child_links.go`: `refused=[{P A ambiguous}] unresolved=[{P A}]`.
- **Why it is in scope:** the user-approved rule "a different child whose parent is confirmed replaces it" assumes the fresh child identifies that delegation. A child linked from two delegations of the same owner proves the parent relation but not the position. The design's R2-P1 limit says both delegations point at one child, but not that the alignment now discards the accepted link silently.
- **Suggested fix:**
  - **Smallest domain change:** a fresh child held by more than one delegation of the owner does not count as a confirmed replacement. Treat it as unconfirmed: the position becomes `movable`, A moves back, and the result is `[A, B]` with A counted as kept.
  - **Alternative:** extend the R2-P1 Known limit to say that, within a same-key group, the accepted link is replaced without being reported, and add a domain case that pins the behaviour.

### A3-02 (severity 1): the meaning of `ChildLinkDecision.ChildSessionID` is undocumented and changes by outcome

- **Classification:** maintainability, pre-existing, unchanged in the delta.
- **Location:** `child_links.go`, the `ChildLinkDecision` struct and the replace case (`entry.decision.ChildSessionID = current`).
- **Detail:**
  - For `retained`, `kept_pending_confirmation`, `refused_*` and `retained_by_position`, `ChildSessionID` is the accepted child.
  - For `replaced`, it is overwritten with the fresh child and equals `FreshChildID`, so the replaced accepted ID is lost.
  - For `cleared`, it is empty.
- **Consequence:** report consumers and future diagnostics cannot tell what was replaced (see A3-01's "B with B"). Document the field per outcome, or keep the accepted ID and rely on `FreshChildID`.

### A3-03 (severity 1): two report doc comments do not match the code

- **Classification:** doc accuracy.
- **Decision order:** `ChildLinkReport` says decisions are "in owner and delegation order". In each owner, all `cleared` decisions come first, then candidate decisions. Probe H3 shows `cleared` at delegation 1 listed before `retained` at delegation 0. This is pre-existing; reword the comment or merge the orders.
- **`Refused` scope:** `Refused` is documented as links whose "child is not linked from the delegation they had". A `held_by_sibling` refusal always concerns a child that another final delegation of the same owner links, even one the fresh read confirms (probe H4: `X` freshly confirmed at the sibling still counts as unresolved). `unplaced` refusals, by contrast, are suppressed when any final delegation links the child.
  - The spec's "Accepted link refused" scenario explicitly includes the held-by-sibling case, so this is per spec.
  - The two reasons do use opposite "already linked" semantics, so this pass is reported partial although the pair was re-proved. It converges on the next pass.
  - Optional: suppress a `held_by_sibling` refusal when the holder's link is confirmed, or state the asymmetry in the comment.
- **Formatting:** the `RetainAcceptedChildLinks` doc comment has one unwrapped line ("...reports them. Owners of other harnesses and owners with nil"). This is cosmetic.

### A3-04 (severity 1): a completed task still describes the superseded rule

- **Classification:** artifact consistency (R2-07 residual).
- **Location:** `tasks.md:34`, "Tester (ordinal group retention)".
- **Detail:** it still lists "Cover refusal and reporting when a fresh link in the group is set". Under position alignment, a set fresh link no longer refuses the group. Design history entries are explicitly superseded by "Position alignment", but this checked task carries no such note.
- **Fix:** add a one-line superseded note, or a pointer to the round-2 task on line 35.

### A3-05 (severity 1): the busy-exit comment overstates its premise

- **Classification:** comment accuracy (nit).
- **Location:** `cmd/agentic-lens/capture.go:801`.
- **Detail:** "The bounded wait expires only while the registry is held elsewhere." The 2 s context also covers `adaptersconfig.Load`, state-directory resolution, opening the registry and the diagnostic insert. Slow I/O without contention would also report busy.
- **Consequence:** the busy exit is retryable, so this is acceptable behaviour, but the comment should say "is treated as contention" rather than claim exclusivity.

## Other checks (no finding)

- **Single entry point:** the CLI calls only `capture.RetainAcceptedChildLinks`. The adapter applies the domain rule at the owner snapshot (`reconcile.go:1092`), the family annotation and `archive_projection.go:110`. No `mergeSessions` fallback exists.
- **Typed states:** three evidence values, seven outcomes and four refusal reasons. Each value has a producer. `ChildLinkRefusedHeldBySibling` (the outcome) and `ChildLinkRefusalHeldBySibling` (the reason) live at different layers, and both are used.
- **Determinism and termination:** the fixpoint loop is deterministic in index order, and each delegation moves at most once.
- **Idempotence on the probes:**
  - Applying the rule again to its own result with the same `Accepted` changed nothing in H1 to H4.
  - Feedback with the published result as `Accepted` converged after one pass (H1 pass 2 gives `[B, ""]`, then stays stable).
- **`refused_contradicted` after `cleared`:** when the cleared fresh child is also the accepted child at the same delegation, both decisions are emitted (probe H2). This is redundant but harmless, and pre-existing.
- **Kept versus refused:** `capture.ChildLinkRetention` removes links that are kept and also refused from `Refused`, which matches the stderr counts.

## Verification performed

- Read the shared brief, round-2 consolidated findings and the corrections delta for `child_links.go`, `capture.go`, the design, spec and tasks. Traced the consumers (`reconcile.go:390-450` and `:1090-1100`, `archive_projection.go:110`) and `resolveSourceChildren` in `codex/source_occurrences.go`.
- Ran the domain probe `TestArchR3Probe` (H1 to H4) through `go test -overlay` at 5d4df7ee, then with 86eb6dae's `child_links.go` and its test file overlaid. Scratch files are under `scratchpad/arch-opus-r3/`.
  - At one point a probe file was copied briefly into `backend/internal/domain/`. It was removed at once, before anything ran, and `git status` was clean afterwards.
- Did not rerun package suites. I reused the evidence in TESTS.md.

## Gaps

- No end-to-end reproduction of A3-01 through Codex rollouts. It is shown at domain level, combined with a code reading of `resolveSourceChildren`'s `ChildReference` fallback (the same path data-opus-r2 described for R2-P1).
- I did not judge whether the R2-T tests are strong enough; that belongs to the tests focus. No test exercises the deadline busy-exit branch (R2-06).
