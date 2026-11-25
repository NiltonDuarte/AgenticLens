# FIX-21 data integrity re-review (data-sonnet-r2)

- Run: data-sonnet-r2, phase re_review (round 2), focus reviewer-data (data integrity and failure handling).
- Requested profile: Claude Sonnet 5.5 / high. Runtime model ID reported to the agent: claude-sonnet-5-5. Effort is not observable; usage and cost are not available.
- Reviewed revision: frozen head `86eb6dae`. Corrections delta `git diff 0f73f759 86eb6dae`; full change `133100bb..86eb6dae`.
- Verdict: APPROVED (scoped to data integrity), with one new design-level finding (F1, severity 4) that needs a disposition: fix it, or record it under Known limits. No defect against the approved spec text was found.

## Round-1 findings in this focus

| ID | Status | Evidence |
|---|---|---|
| DATA-01 | Resolved | `refreshTicketOwners` now applies `RetainAcceptedChildLinks` against `previousSnapshots` when it builds the per-owner snapshot (`reconcile.go`, saveCaptureRead callback). A pass that stops before the family annotation keeps every accepted link. `TestReconcileInterruptedAnnotationKeepsRegistryLinks` interrupts the annotation with a SQLite trigger. It asserts the parent's own snapshot committed, the links remain, and the next partial pass still reports the unread children (`PendingChildren == 2`). I ran it: pass. Residual: a link the rule refuses is not written at the first commit either, so an interruption cannot restore it. That is the same rule as F1, not a new loss. |
| DATA-02 | Resolved | `ChildLinkReport.Refused` and `UnresolvedLinks()`. The adapter counts them (`addUnresolvedChildLinks` after `pendingChildWork`). The CLI helper counts them through `capture.RetainAcceptedChildLinks`, and records the partial diagnostic when `Added > 0`. The stderr line prints the refused count. Probe below: no silent loss. |
| DATA-03 | Resolved | `pendingChildLinks` falls back to `codex.RolloutNameSessionID(filepath.Base(path))` for an unregistered child's queued path, so retained and queued pairs share one key. It is only a dedup hint. The worst case is a renamed rollout whose file-name UUID differs, which would undercount by one while `pending_children` still keeps the pass partial. |
| DATA-04 | Resolved | `ChildLinkInput.Known` takes precedence over `Accepted` for child evidence, and the capture helper passes `result.Sessions`. `TestRetainAcceptedChildLinksUsesRegistrySessionsAsChildEvidence` passes. A Known record without a parent is unobserved, which matches the accepted-record rule. |
| BEH-01 (data side) | Resolved | `last := cursor` keeps a resumed sweep's cursor. I traced the end states. A completed walk (`visited <= maxEntries && !exhausted`) still closes with `""`. An exhausted sweep that admitted nothing and had no cursor stays unstarted (`continue`, no row). A resumed one rewrites its own cursor. |
| ARCH-04 | Resolved | A `.gz` header read needs the full 64 KiB bound before it starts, so charged decompressed bytes cannot abandon a read. Charges remain within the total. |
| ARCH-02 | Resolved (parity checked) | `exactSessionContext` and `conflictingNativeParent` keep the previous order: missing identity or working directory first, then the conflict check. Both `readSessionContext` and `RolloutHeader.SessionContext` give the same results as before. |
| BEH-02 / ARCH-03 | Resolved (data side) | A single `capture.RetainAcceptedChildLinks` runs once, before pending sync, merge, restore acknowledgement and publication. `IsRegistryContention` covers the `errors.Join` form, so a busy registry exits 75. On any error the in-memory mutation is discarded because the process exits before anything is published. |

## Investigation of the corrections delta and ordinal retention

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | Ordinal carry invents or misplaces a link. | Rejected. Realistic worlds: the accepted delegations were re-read with empty links and extra spawns appended, with random keys, untimed and typed mixes and children missing. In 300k cases every accepted link stays at its own position, no appended delegation gets a link, and nothing is refused. I ran this before my scratch probe directory was overwritten by another process; I could not re-run it, so the number rests on that first run. Random unrelated worlds (400k, with contradicting or confirming observed parents): every output link is a fresh link or an accepted link of the same owner. No child is placed on two delegations beyond the fresh input. |
| H2 | Not idempotent, or not stable across passes with the published output as the accepted ledger. | Rejected. 400k random cases: `f(f(x))` equals `f(x)`, and `f(fresh, A=f(fresh, A0))` equals `f(fresh, A0)`. The same holds in the realistic worlds. |
| H3 | An accepted link is lost without being reported. | Rejected. 400k random cases, excluding the key groups absent from the fresh read and the legitimate replacement by a confirmed fresh child: every accepted, non-contradicted link missing from the output is in `Report.Refused`. |
| H4 | The ledger's chronological sort breaks positional alignment. | Rejected. `chronologicalFacts` is a stable sort that keeps unknown-time rows in their slots. A key group shares one timestamp string (or none), so its internal order is preserved. |
| H5 | Double counting. | Rejected. One `countedChildLinks` map seeds from pending starts and paths and is extended by the adapter's retained and refused links and by the helper. `Kept` is the number of distinct kept links, and refused links that are also kept are removed from the refused set. |
| H6 | Interrupted passes lose retained links or hide them. | Rejected for the supported window (DATA-01). The next pass recounts them from the registry. |
| H7 | Known evidence flips a decision against the ledger. | Low risk. A registry child record without `ParentID` hides a ledger record that names this parent or another one. It only matters for contradiction and replacement, and it matches the approved "no parent is unobserved" rule. Not reported as a finding. |
| H8 | Cursor or gz changes lose a candidate. | Rejected (see the table above). |

## Findings

### F1. A key group that is only partly read loses its unread accepted links from the persisted facts (severity 4, design-level, introduced)

- Location: `backend/internal/domain/child_links.go`, `reconcileOwnerChildLinks`, the `!empty || len(fresh) < len(accepted)` branch (`alignmentRefused`). It acts at all three application points: the first owner commit, the family annotation (`finishCaptureAnnotations` and `regroupArchiveFamily`) and the `--capture-auto` helper.
- Classification: introduced by FIX-21. The behavior is what the approved ordinal rule says ("When any fresh delegation in the group has a link ... nothing in the group is carried"). The finding is that this choice recreates, in a smaller case, the loss the ordinal rule was approved to remove.
- Trigger: a parent has several spawns with the same agent type, description and invoked_at. A description is the optional `task_name`, so many spawns have `""`. Some of those children are loaded in the pass (their `agent_id` resolves the delegation, `source_occurrences.go` lines 100 to 104) and others are not.
- Evidence: probe in an overlay test (domain package, no product file changed). Accepted group `[C1 C2 C3]`, one key:

  ```
  fresh [C1 "" ""]   -> out [C1 "" ""]   refused 2 unresolved 2 | next pass refused 0 unresolved 0
  fresh ["" C2 ""]   -> out ["" C2 ""]   refused 2 unresolved 2 | next pass refused 0 unresolved 0
  fresh [C1 C2 ""]   -> out [C1 C2 ""]   refused 1 unresolved 1 | next pass refused 0 unresolved 0
  fresh ["" "" ""]   -> out [C1 C2 C3]   refused 0 unresolved 3 | next pass refused 0 unresolved 3
  ```

- Consequence: the unread links are reported as unresolved once and then removed from the registry snapshot (the annotation overwrites the delegations) and from the published ledger (the helper applies the same rule). The next pass has them in neither accepted set, so retention reports nothing about them. The pair stays visible only while a `pending_starts` or `pending_children` row exists. If the children are read later, task-name and agent-ID resolution relinks them, so the loss is limited to the time they stay unread. On a machine that can never read them again the ledger permanently loses accepted links, which is the case the ordinal approval was meant to cover.
- Why a fix would stay within the approved safety properties: in each probe row the non-empty fresh links equal the accepted link at the same position. A rule of "carry position i only where the fresh link is empty and every non-empty fresh link of the group equals the accepted link at its position" keeps the no-invention and idempotence properties. A different fresh link at a position would still refuse the whole group.
- Disposition options: (a) planner and user decide whether to refine the rule, with the same approval route as the ordinal change; (b) record it under Known limits ("a partly read ambiguous group is reported, then dropped") with a scenario in the spec. Both are acceptable. Silent acceptance is not, because the design's Known limits currently do not list it.

## Observations (no finding)

- `refused_held_by_sibling` counts the child as unresolved although a sibling holds it in the output. This is transient: the next pass has the child on the sibling in the accepted set and reports nothing.
- An accepted delegation whose key group is absent from the fresh read is neither carried nor reported. The fresh read is authoritative for which delegations exist, as before.
- Stage-1 and stage-2 retention both use the registry snapshot taken at the start of the work-class round. A concurrent pass that wrote newer links in between is overwritten. This predates FIX-21 (`RetainIncompleteGroupSummaries` uses the same snapshot).
- A `.gz` owner needs 64 KiB of free extraction allowance for its header. With a very small `max_source_bytes` it stays pending, which is the recorded Known limit.

## Verification performed

- Read: the corrections delta for `domain/child_links.go`, `capture/reconcile.go`, `children.go`, `archive_projection.go`, `cmd/agentic-lens/capture.go`, `main.go`, `codex/context.go`, `discovery.go`, `discovery_budget.go`; the unchanged `resolveSourceChildren`, `completeOccurrence`, `chronologicalFacts`, `IsRegistryContention`, `OpenProjectContext`; round-1 `REVIEW.md`, `design.md`, the retention spec.
- Ran focused tests: `TestReconcileInterruptedAnnotationKeepsRegistryLinks`, `TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses`, `TestRetainAcceptedChildLinks*` (capture): pass. `TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState` and `TestManualGenerateLeavesLinkToMissingChildEmpty` (cmd): pass.
- Property probes in `go test -overlay` test files kept outside the tree (scratchpad `data-sonnet-r2-probe`): random worlds, idempotence, published stability, no invention, no silent loss, and the mixed-group table above.
- Reused `evidence/TESTS.md`. No package suite was rerun.

## Gaps

- F1 is proven at the domain rule and traced through the adapter and the helper; I did not build a full registry fixture with one of several same-key children loaded.
- The first realistic-world probe run (300k cases) was lost to a shared scratchpad overwrite and not repeated; the later runs cover random worlds, stability and no silent loss.
- I did not exercise OpenCode identification, the worker's busy backoff end to end, or cross-machine pending-snapshot recovery.
