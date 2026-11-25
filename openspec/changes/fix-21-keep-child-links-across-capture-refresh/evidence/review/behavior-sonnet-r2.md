# FIX-21 behavior and integration re-review, round 2 (beh-sonnet-r2)

- Run: FIX-21 re_review `beh-sonnet-r2`, focus behavior and integration. Requested identity: Claude Sonnet 5.5, effort high. Runtime metadata was not inspected beyond the requested identity.
- Reviewed revision: frozen head `86eb6dae` (full change `133100bb..86eb6dae`, corrections delta `0f73f759..86eb6dae`). The worktree had no tracked or product modifications; the only untracked files were review reports.
- Scope: end-to-end automatic capture (worker -> `Reconcile`/`ReconcileAt` -> `generate --capture-auto` -> publication), the BEH-01 and BEH-02 fixes, ordinal group retention, and parity of the manual, backfill, live, OpenCode and Claude Code paths.
- Verdict: **APPROVED** for the assigned focus. No finding blocks delivery. One residual limit (N1, severity 3) is broader than the recorded Known limits and should be recorded in delivery; it is an approved rule behaving as specified.

## Round-1 findings in this focus

| Finding | Status at `86eb6dae` | Evidence |
|---|---|---|
| BEH-01 (resumed sweep closed with cursor `''`) | Resolved | `children.go`: `last := cursor`, and a brand-new sweep that admitted nothing `continue`s without writing a cursor. I re-ran the round-1 reproduction (two parents, four hints each, budgets 12 KB to 300 KB, six small `ReconcileAt` passes, then one 64 MiB pass) on a scratch copy of `86eb6dae`. Every budget now keeps both sweeps open (`cur=2`, `disc=T`) through the small passes and ends with all 10 owners found and `pendingChildren=0`. Round 1 ended with 6 of 10 owners and closed cursors. Also passes: `TestResumedSweepAtTheCeilingKeepsItsFileCursor`, `TestBrandNewSweepAdmittingNothingWritesNoCursor`. |
| BEH-02 / ARCH-03 (busy handling, one capture-side call) | Resolved | `retainCaptureChildLinks` makes one call to `capture.RetainAcceptedChildLinks`. That call opens the registry with `OpenProjectContext` only when the unresolved count grew, and records the partial diagnostic itself. The CLI maps `capture.IsRegistryContention` to `capturePublicationBusyExit` before any publication step, as `main.go` does for reconcile. `TestRetainAcceptedChildLinksReportsRegistryContention` shows a held writer surfacing as contention with the earlier completed diagnostic untouched, and a later request recording the partial state. The process-level exit mapping is read, not exercised (see N4). |
| DATA-02 (refused links not signalled) | Resolved | `UnresolvedLinks()` (retained plus refused) feeds `finishCaptureAnnotations`, `regroupArchiveFamily` and the CLI helper. The CLI line reports kept, refused and cleared counts. The refused-link case is covered by `TestRetainAcceptedChildLinksCountsRefusedLinksOnceBesideFreshLinks` and the domain refusal table. |
| DATA-01 (interrupted pass before annotation) | Resolved | The owner snapshot commit now applies `RetainAcceptedChildLinks` against `previousSnapshots` before `RetainIncompleteGroupSummaries`. `TestReconcileInterruptedAnnotationKeepsRegistryLinks` passes. |
| DATA-03, DATA-04 | Resolved | The name-based hint in `pendingChildLinks` only dedups. The same pair also reaches `PendingChildren` through the queued row, so a mismatched name can only under-count identities, never hide a pending item. `ChildLinkInput.Known` takes precedence over `Accepted` for child evidence, and `TestRetainAcceptedChildLinksUsesRegistrySessionsAsChildEvidence` passes its three subcases. |
| ARCH-02 / ARCH-04 / ARCH-05 | Resolved | `exactSessionContext` keeps the old order: required identity and working directory first, then the conflicting-parent check. `SessionContext` returns the same two error classes as before, and `readSessionContext` is equivalent. `.gz` header admission now requires the whole 64 KiB bound. |
| ARCH-01 | Dispositioned correctly | Accepted debt, recorded as a limit. |

## Ordinal group retention (new behavior)

I traced it through the domain rule, stage 1 (owner snapshot and annotation against registry `previousSnapshots`) and stage 2 (CLI against the committed ledger).

- Position carry is correct under the append-only premise. The ledger writer (`chronologicalFacts`) sorts only timestamped rows, stably, and leaves untimed rows in their slots. Relative order inside one key group therefore survives publication, so accepted position i maps to fresh position i.
- The rule never invents a link. A move needs an accepted link, an empty fresh slot, no contradiction and no sibling holding the child. Repeated application is stable. Probes: repeated ordinal carry, and carry applied to its own output, both give identical links and counts.
- The end-to-end test `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses` passes. The registry and published links survive two partial passes, and the count is 2 each time.
- Counting does not double. The adapter counts stage-1 carries once, and the CLI stage, which sees already-filled slots, adds nothing, so `Added == 0` and the diagnostic is not rewritten.
- Parity holds for the manual, backfill, live, OpenCode and Claude Code paths:
  - Manual and backfill still call `codex.DiscoverRolloutPaths`.
  - Live hydration is unchanged.
  - The only new generation call sits under `--capture-auto`, behind `Harness: Codex`.
  - `git diff 0f73f759 86eb6dae` touches no `cmd/` file, so the worker and the `main.go` call site are as reviewed in round 1.

## New findings

### N1 (severity 3, residual limit, not a defect against the spec) - a group with any fresh link loses all its unread accepted links, with a wider trigger than the Known limits record

- Location: `domain/child_links.go`, `reconcileOwnerChildLinks`, the `!empty || len(fresh) < len(accepted)` branch.
- Trigger: one key group (identical agent type, description and invoked_at, or unknown invoked_at) in which this pass resolves at least one fresh link. Task-name and `agent_id` results resolve per call, so a link appears at its own position whenever its child is read. Realistic triggers:
  - a parent refreshed together with one still-running child while its finished siblings are not re-read;
  - a newly spawned trailing delegation whose child is read first.
- Consequence: no accepted link in the group is carried. The unread ones are reported as unresolved once (`PendingChildren` +1 for each) and then disappear from the registry snapshot and the published ledger. On the next pass the accepted facts no longer contain them, so nothing is reported. They return only when the child is read in the same call as its parent. This is the approved refusal branch ("Ambiguous delegation provenance, refusal"), but the design's Known limits do not say that a partially read group triggers it.
- Evidence (domain probe on a scratch copy of `86eb6dae`, group of identical delegations):
  - accepted `[C1 C2 C3]`, fresh `["" C2 ""]` gives links `["" C2 ""]` with C1 and C3 refused.
  - accepted `[C1 C2]`, fresh `["" "" C3]` gives links `["" "" C3]` with C1 and C2 refused.
  - Two-stage case: the registry holds `[C1 C2]` and the ledger `[C1 C2 C3]` (for example after a manual generate on the same machine). Stage 1 fills positions 0 and 1 by position. Stage 2 then sees a non-empty group and refuses C3, which is reported once and dropped.
  - End-to-end probe (adapter, two same-key spawns, child 1 large, child 2 small, parent read with child 2): pass 1 publishes `["" child-2]`, registry the same, `PendingChildren=1`. Passes 2 and 3 report `PendingChildren=0` with the link still empty. A final full-budget pass restores `[child-1 child-2]`.
- Suggestion for the parent: record this under Known limits and in delivery. Narrowing the rule so that only empty positions below the accepted group's length are carried, when the non-empty fresh links agree with their positions, would remove it, but that changes the approved rule and needs renewed approval.

### N2 (severity 2) - accepted untimed group with a fresh group that has a known invoked_at is dropped without being reported

- Location: `refusedChildLinks`, which reports only keys marked `alignmentRefused`.
- Trigger: the accepted delegation has no invoked_at and the fresh read of the same delegation has one. The keys differ, so there is no counterpart group and the link is neither carried nor reported.
- Evidence: domain probe, accepted `[C1]` with unknown invoked_at, fresh `[""]` with a known time, gives links `[""]`, no decisions, no refusals.
- Reachability: low. Migrated ledgers preserve `link.Timestamp`, and a rollout supplies the same timestamp each time. It needs a ledger written without delegation timestamps.
- Consequence: a silent drop of the kind DATA-02 closed for the group case. Report only if cheap.

### N3 (severity 1) - `refused_ambiguous` decisions are emitted for complete timed groups

- Location: the `key.timed` loop in the refusal branch.
- Evidence: accepted `[C1 C2 C3]` against an identical fresh group yields three `refused_ambiguous` decisions and no refusals.
- Consequence: counts and links are unaffected, because those decisions are neither retained nor unresolved. The `ChildLinkReport` comment ("kept as read are omitted") is not literally true for ambiguous groups. An owner with only such decisions is also rewritten with an unchanged copy. Harmless, and a comment fix at most.

### N4 (severity 1) - busy mapping not exercised at process level

- `retainCaptureChildLinks` uses a 2 s context. If that deadline expires while the diagnostic write waits, the error would be a context error rather than contention and would exit 1 instead of 75, which makes the worker back off. The adapter test covers contention, but nothing runs the CLI exit. This is the same behavior as the neighbouring helpers.

## Hypotheses investigated and not found

| Hypothesis | Result |
|---|---|
| Stage 1 and stage 2 double count a carried group | Refuted. Stage 2 sees the filled slots and adds nothing. |
| Ordinal carry moves a link to a wrong position after publication reordering | Refuted by reading `chronologicalFacts`. Order within a key group is stable. |
| A contradicted or already-held child is carried inside a group | Refuted by the position-candidate guards. |
| `last := cursor` can stall or loop a sweep | Refuted. A completed resumed sweep still closes. An exhausted one rewrites the same cursor. A fresh sweep that admits nothing leaves no cursor. |
| Retention and `Known` change events or other harnesses | Refuted. Only Codex `Delegations` are rewritten. |
| `readSessionContext` / `SessionContext` behavior changed in the refactor | Refuted by order-of-checks comparison and focused codex tests. |

## Verification performed

- Read in full: `domain/child_links.go`; `capture/reconcile.go` (stage-1 retention, annotation, `RetainAcceptedChildLinks`, diagnostics); `children.go` (sweep, queued path, `pendingChildLinks`); `archive_projection.go`; `codex/context.go`, `discovery_budget.go`; `cmd/agentic-lens/capture.go` and `main.go` call site; `capture_worker.go`; `fact_ledger_order.go`; `source_occurrences.go` (`resolveSourceChildren`).
- Focused runs on a scratch copy (`git archive 86eb6dae`), all passing:
  - capture: `TestResumedSweepAtTheCeilingKeepsItsFileCursor`, `TestBrandNewSweepAdmittingNothingWritesNoCursor`, `TestRetainAcceptedChildLinks*`, `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`, `TestReconcileInterruptedAnnotationKeepsRegistryLinks`, `TestReconcileFreshContradictingChild*` and `Nested*`.
  - codex: the `SessionContext`, `Provenance`, `Nested`, `Mismatch` and `Discovery` test subsets.
- Focused run in the worktree: `go test ./cmd/agentic-lens -run TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState`, with all five subtests passing.
- Probes (scratch copy only; none of this was added to the worktree): the BEH-01 budget sweep; the domain probe for N1, N2 and N3 and for idempotence; the adapter-level partial-group probe.
- Reused, not rerun: whole-package results, race runs, large-store test and mutations M1 to M19 in `evidence/TESTS.md`.

## Gaps and limits

- I did not run the full `capture`, `codex` or `cmd/agentic-lens` suites, per the brief.
- I did not execute the worker process or the CLI busy exit end to end.
- N1 was demonstrated on synthetic same-key fixtures. How often real Codex transcripts produce identical-key groups was not measured.
- OpenCode and Claude Code were checked by reading the call sites. No new probe ran against them.
- No other round-2 reports were read.
