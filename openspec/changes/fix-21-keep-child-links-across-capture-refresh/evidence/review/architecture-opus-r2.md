# FIX-21 architecture re-review (arch-opus-r2, round 2)

- Role: reviewer-architecture (architecture and maintainability).
- Run: arch-opus-r2, phase re_review. Requested profile: Claude Opus 5.5, medium effort. The runtime reports model claude-opus-5-5. Actual effort is not exposed.
- Reviewed revision: head `86eb6dae7d28529902653faa3326c5df89027969`. The corrections delta is `0f73f759..86eb6dae`, inside the full change `133100bb..86eb6dae`. Product code was read-only.
- Verdict: **CHANGES_REQUESTED**, for one design gap in the ordinal group rule (R2-ARCH-01). All round-1 architecture findings in scope are resolved or correctly dispositioned.

## Round-1 findings

| ID | Status at 86eb6dae | Evidence |
|---|---|---|
| ARCH-03 / BEH-02 | Resolved | `capture.RetainAcceptedChildLinks` (`capture/reconcile.go:414`) now owns both domain calls, the kept/refused/cleared counting, the unresolved dedup and the diagnostic write. `AddRetainedChildLinks` and `ReconciliationDiagnostic` are now unexported. No caller can record "completed" out of order. The registry is opened through `OpenProjectContext` only when `Added > 0`. The CLI helper (`cmd/agentic-lens/capture.go:795`) is down to the call, the busy exit (`IsRegistryContention`, then `capturePublicationBusyExit`) and stderr. `errors.Join` of a single busy cause still classifies as contention. |
| main.go footprint | Minimal | One line at `main.go:601`. It does not overlap `runBackfill`. |
| ARCH-04 | Resolved | `ReadHeader` admits a `.gz` rollout only when the full 64 KiB bound fits (`discovery_budget.go:260-266`). Suffix detection matches `readRolloutHeader` (line 133). |
| ARCH-02 | Resolved | `exactSessionContext` and `conflictingNativeParent` (`codex/context.go`) are shared by `readSessionContext`, `RolloutHeader.SessionContext` and `headerFromMetadata`. Behaviour is unchanged: the required-field error comes first and the conflict error second, with the same messages and the same `ErrSessionContextMismatch` wrapping. |
| ARCH-05 | Mostly resolved | The deferral comment (`reconcile.go:854`) and the Codex-only `readMetadata` comment are corrected. `refused_ambiguous` now carries `ChildParentNotEvaluated`. Residual: the disposition says the test-only exports stay "with a rationale", but no rationale is recorded. `DiscoverExactRolloutPathWithBudget` (`discovery.go:492`) still has no doc comment and no production caller, and `HeaderRead.Cached` has no production reader. Optional, 1/10. |
| ARCH-01 | Accepted debt | Recorded under Known limits in design.md. |

## Hint use and typed states

- **Name hint (DATA-03).** `codex.RolloutNameSessionID` has one production caller, `pendingChildLinks` (`children.go:696`). It is used only as a key in the dedup set. A wrong name can only merge two counted links into one, and the queued path is still counted as pending work by `pendingChildWork`. A wrong name therefore cannot make a pass "completed". Identity still comes from `session_meta`.
- **`ChildLinkInput.Known` (DATA-04).**
  - The precedence is fresh or observed, then Known, then Accepted. A Known record with no parent is unobserved and does not fall back to Accepted, as documented.
  - Only the CLI entry point passes Known (`result.Sessions`). Adapter-internal sites already use registry snapshots as Accepted. The layering is coherent.
- **Typed states.**
  - `childLinkKey` keeps `typed` and `timed` flags, so a nil agent type stays distinct from an empty one, and an unknown invocation time is its own key.
  - `childLinkAlignment` is a typed enum, and the outcome and refusal enums are typed strings.
  - A scan of the lines added since `133100bb` finds no `any` or `interface{}`.
  - `alignmentNone` and `alignmentUnique` are written but never read. This is cosmetic.

## New findings

### R2-ARCH-01. The partial re-proof refusal still drops links from the published ledger (design gap, 4/10)

- **Where:** `domain/child_links.go`, `reconcileOwnerChildLinks`, refusal branch `if !empty || len(fresh) < len(accepted)`. The spec makes the same choice ("Otherwise it SHALL carry none of the group").
- **Trigger:** a key group where the fresh read re-proves some links but not others. For example, accepted `[A, B]` and fresh `[A, ""]`. This happens when a bounded pass reads child A but not child B through `agent_id` resolution, which is the scenario FIX-21 targets.
- **Consequence:**
  - Pass 1 refuses the whole group, reports B once as unresolved (partial), and publishes `[A, ""]`.
  - Pass 2 then has accepted `[A, ""]`. B is gone with no unresolved evidence, and coverage can be reported complete.
  - This is the same failure class that the approved "Ordinal group retention" steering names as its Problem ("the published ledger then lost those links, so later passes had nothing to report").
  - On a machine that can never re-read B (the portable-ledger case), the loss is permanent.
  - The branch is not recorded under Known limits.
- **Evidence:** a go-overlay probe. It adds a virtual `zz_arch_probe_test.go` in `internal/domain` and leaves the tree unchanged.
  ```
  pass1 links=[A ] refused=[{{p B} ambiguous}] unresolved=[{p B}]
  pass2 links=[A ] refused=[] unresolved=[]
  ```
- **Suggestion:** treat a group as position-alignable when every set fresh link equals the accepted link at the same position. A consistent prefix contradicts nothing. Keep refusing when a set fresh link disagrees with its position or the fresh group is smaller. Otherwise, record this branch explicitly as a Known limit. Changing the acceptance condition is a spec change, so the parent should confirm the steering scope with the user.

### R2-ARCH-02. `refused_ambiguous` decisions in steady state contradict the report's contract (introduced in round 1 and kept, cosmetic, 1/10)

- **Where:** the same refusal branch, decision loop.
- **Trigger:** applying the rule to a fully linked group, for example accepted `[A, B]` and fresh `[A, B]`.
- **Effect:**
  - It emits `refused_ambiguous` decisions for both delegations, with an empty `ChildSessionID`, even though nothing was declined.
  - `ChildLinkReport`'s doc says delegations "whose fresh link is kept as read are omitted".
  - The owner's delegations are also rewritten with an identical copy.
  - Untimed refused groups emit no decision, while timed ones do.
- **Probe:** `steady decisions=[{... FreshChildID:A Outcome:refused_ambiguous ...} {... FreshChildID:B Outcome:refused_ambiguous ...}]`.
- **Impact:** none today, because the only consumer of `Decisions` counts `ChildLinkCleared`. A future diagnostic built on `Decisions` would over-report.
- **Suggestion:** emit the decision only for an accepted link that is not linked in the final delegations, like `refusedChildLinks` does, or correct the doc.

### R2-ARCH-03. The stderr wording asserts "remains partial" when only links were cleared (cosmetic, 1/10)

- **Where:** `cmd/agentic-lens/capture.go:809`.
- **Effect:** the message prints when `Cleared > 0` even if `Kept == Refused == 0` and `PendingChildren == 0`. The reconciliation may then be completed, yet the message says it remains partial.

## Regression hypotheses checked

| Hypothesis | Disposition |
|---|---|
| Ordinal carry invents or misplaces a link | Rejected within the append-only premise. A carry needs every fresh link in the group to be empty and the fresh group to be at least as large, and it applies the per-child contradiction and held-by-sibling guards. No code in the repository reorders delegations (searched for sort calls), and Codex appends delegations in call order (`occurrences.go`). |
| Ordinal carry is not idempotent | Rejected for sessions: re-applying the rule to its result leaves the delegations unchanged. Reports differ (R2-ARCH-02). |
| The first-commit retention (DATA-01) leaves uncounted links | Rejected for the normal flow. `finishCaptureAnnotations` re-applies the rule from the same pre-commit `previousSnapshots` and counts the result. Owners marked stale by `keepCurrentCaptureReads` become pending. |
| Double counting between adapter and CLI | Rejected. Both go through `countedChildLinks` in `addUnresolvedChildLinks`. In the CLI, kept links are removed from the refused set. |
| `readSessionContext` behaviour parity | Confirmed. The same checks run in the same order with the same errors. |
| Busy handling | Confirmed. Contention exits 75, and an expired context exits 1, which is the same policy as reconciliation. |

## Verification performed

- I read `git diff 0f73f759 86eb6dae` for all product files in my focus. I read the head versions of `domain/child_links.go`, `capture/reconcile.go` (the entry point, the first commit, the annotation and counting), `capture/children.go`, `capture/archive_projection.go`, `codex/context.go`, `codex/discovery*.go` and the CLI call sites.
- `go vet ./internal/domain ./internal/adapters/capture ./internal/adapters/codex ./cmd/agentic-lens`: clean.
- `go test ./internal/domain -run ChildLink -count=1`: ok.
- Overlay probe `go test -overlay ... ./internal/domain -run TestArchProbe -v`: output as quoted above. The probe file stays in the session scratchpad, not in the tree.
- I reused `evidence/TESTS.md` and did not rerun package suites.

## Gaps

- I did not probe the partial re-proof case end to end through capture and CLI publication. The domain probe plus code reading shows that the published `RefreshedSessions` carry `[A, ""]`.
- I did not assess test strength or the large-store design. Those belong to other reviewers.
