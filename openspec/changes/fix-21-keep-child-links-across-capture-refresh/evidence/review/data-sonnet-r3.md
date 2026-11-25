# FIX-21 data integrity re-review (data-sonnet-r3)

- Run: data-sonnet-r3, phase re_review (round 3), focus reviewer-data (data integrity and failure handling).
- Requested profile: Claude Sonnet 5.5 / high. Runtime model ID reported to the agent: claude-sonnet-5-5. Effort is not observable; usage and cost are not available.
- Reviewed revision: frozen head `5d4df7ee` (checked out HEAD `5d4df7ee0dc6`). Corrections delta `git diff 86eb6dae 5d4df7ee`; full change `133100bb..5d4df7ee`.
- Verdict: APPROVED (scoped to data integrity). No blocking defect found. Three non-blocking observations (N1 to N3) are listed for disposition. N1 is the one worth a conscious decision.

## Round-2 findings in this focus

| ID | Status | Evidence |
|---|---|---|
| R2-01 | Resolved | `reconcileOwnerChildLinks` pairs accepted position i with fresh position i inside each key group and applies the unique-delegation rules at that pair. A partly re-read group keeps its unread links. Property probes (below) found 0 silent losses in 300k random worlds and 0 lost or misplaced links in 100k grow-only multi-pass simulations. The CLI second trigger is covered by `TestCaptureAutoAlignsAPartlyResolvedGroupWithTheCommittedLedger` (run: pass). |
| R2-02 | Resolved | An untimed unique key is a group of one. A different fresh child that is not confirmed no longer replaces the accepted link (`kept_pending_confirmation`); a confirmed one replaces it. Covered by the domain table and the fuzz stability check. |
| R2-03 | Resolved | A prior key with no fresh group is stored in `unplaced` as `unmatched` and is reported by `refusedChildLinks` unless its child is linked in the final result or contradicted. Probe: every accepted, uncontradicted link missing from the output was in `Report.Refused`, except the legitimate replacement by a confirmed fresh child. The later loss of such a link is recorded in design Known limits. |
| R2-04 | Resolved | Decisions contain only retained, retained_by_position, kept_pending_confirmation, replaced, cleared, refused_contradicted and refused_held_by_sibling. Complete groups never produce `ambiguous` or `unknown_provenance`; the fuzz check (complete groups produce only held_by_sibling refusals) held in 300k cases. `held_by_sibling` is still a refusal decision inside complete groups, which the design text allows ("Held by another delegation"). |
| R2-05 | Resolved | `partial` is "." unless `retention.Added > 0` (`cmd/agentic-lens/capture.go`). `TestCaptureAutoReportsOnlyClearedLinksWithoutPartial` passes. |
| R2-06 | Resolved by code reading | `retainCaptureChildLinks` exits `capturePublicationBusyExit` for `IsRegistryContention(err) || errors.Is(err, context.DeadlineExceeded)`. The exit happens before pending sync, merge and publication, so nothing is published after a busy outcome and a retry recomputes the same retention. The CLI mapping itself is not unit-tested because the helper calls `os.Exit` (see gaps). |
| R2-07 | Resolved | `design.md` and `tasks.md` name `capture.RetainAcceptedChildLinks`, `Known`, the file-name dedup hint, per-owner application and the three evidence states. No stale API name remains outside the superseded Steering history (`ChildParentNotEvaluated` and `ChildLinkRefusedAmbiguous` appear only in the R2 sync note that records their removal). |
| R2-T1 | Verified, not rerun | The contention test now creates a separate `retryCtx`. I ran the capture retention tests once (pass) but did not rerun `-race`; the 3 of 3 and 20 of 20 race results in `evidence/TESTS.md` are reused. |
| R2-T2, T3, T8 (domain side) | Verified | `TestRetainAcceptedChildLinksAlignsGroupsByPosition` and `...OrdinalGroups` pass. Their cases (interleaved and offset groups, contradicted member of a refused group, no decisions for complete groups) match the behavior my probes derive independently. |
| R2-T4, T5, T6, T7 | Out of data focus | Not investigated beyond noting that the capture tests that exercise retention and the cmd tests pass in a focused run. |

## Investigation of the rewritten rule

Method: the unchanged `child_links.go` was copied into a scratch module with minimal type stubs (scratchpad directory `data-sonnet-r3`, no repository file touched), then driven by random worlds. Worlds mix agent types, descriptions, `invoked_at` values (including unknown), repeated keys, shuffled fresh order, truncation, extra delegations and child records that confirm, contradict or are absent (observed, known and accepted records).

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | The rule invents a link or places an accepted link at a position other than the paired one in the same key group. | Rejected. 300k random worlds: every output link equals the fresh link at that position or the accepted link at the position-paired slot of the same key group. Delegation fields other than the child are never changed. |
| H2 | A child is placed on two delegations beyond what the fresh read already did. | Rejected. 300k worlds: no duplicate child was introduced by the rule. |
| H3 | A contradicted child survives in the output. | Rejected. 300k worlds. |
| H4 | The rule is not idempotent. | Rejected. 300k worlds: applying the rule to its own output with the same accepted facts returns the identical delegations. |
| H5 | Feedback instability: publishing the output as the next accepted ledger changes the result of the same fresh read. | Rejected. 200k worlds: `f(fresh, f(fresh, A0)) == f(fresh, A0)`. |
| H6 | Realistic grow-only history loses or misplaces a link over several partial passes. | Rejected. 100k simulations: a transcript that only grows, a ground-truth child per delegation, a random subset revealed per pass (child records sometimes present), with each pass's output as the next accepted ledger. Over 6 passes no revealed link was ever lost, no link sat at a wrong position and no refusal was emitted. |
| H7 | An accepted link disappears without a report. | Rejected. 300k random worlds, excluding links whose paired fresh child is confirmed (legitimate replacement) and contradicted links: 0 silent losses. The probe's definition of loss was checked against `Report.Refused`, which is where smaller groups, unmatched keys and held-by-sibling land. |
| H8 | A reported refused link is actually linked in the result (false unresolved). | Observed, low. See N2. Only `held_by_sibling` does this; smaller-group and unmatched refusals are reported only when the child is not linked in the final result. |
| H9 | Chronological ledger sorting breaks positional alignment. | Rejected. `chronologicalFacts` keeps unknown-time rows in their slots and sorts stably by instant; one key group shares one timestamp string, so its internal order cannot change. Transcript order inside a key group is deterministic (`occurrences.go` appends in call order and de-duplicates by call ID). |
| H10 | Per-owner snapshot application loses a link that the family annotation then cannot restore. | Rejected for the supported window. The first commit applies the rule against `previousSnapshots`, so an interruption before annotation keeps all placeable links (DATA-01, test `TestReconcileInterruptedAnnotationKeepsRegistryLinks` runs and passes in the focused capture run). A link the rule refuses (smaller group or unmatched key) is not written at that first commit either, and its report is made only in the annotation. If a pass stops in between, the refused link is neither kept nor reported by the next pass from the registry. This is the same one-time-report limit that design Known limits already records, extended to an interrupted pass; the `--capture-auto` stage re-derives the report from the ledger, which still has the link. |
| H11 | Dedup counting double counts. | Rejected. `countedChildLinks` is seeded from `pending_starts` and `pending_children` (with the file-name hint) and extended by the adapter's retained and refused links and then by the helper. The adapter adds its pairs once at the end of `ReconcileAt`. The helper reports `Kept` as distinct links and removes links that are both kept and refused. The delta did not touch `reconcile.go`; the counting behavior is unchanged since round 2. |
| H12 | Known precedence flips a decision. | Low. `evidence` checks fresh and observed records, then Known, then Accepted; a Known record with no parent is unobserved and shadows a contradicting accepted record. This is the approved DATA-04 rule and the same as in round 2. |
| H13 | The stderr line or the busy exit hides an incomplete pass. | Rejected. The wording now says "remains partial" only when `Added > 0`; when a link was kept but already counted, the pass is already partial through the existing pending rows or the registry-stage count. A busy or failed retention exits before publication, so no partial ledger is published. |

## Non-blocking observations

### N1. A confirmed fresh duplicate overwrites an accepted link when two delegations share a task name (severity 3, pre-existing root cause R2-P1)

- Location: `backend/internal/domain/child_links.go`, `reconcileOwnerChildLinks`, the `Replaced` branch, together with `codex/source_occurrences.go` `resolveSourceChildren` (unchanged).
- Trigger: two delegations of one parent share agent type, description and `invoked_at`, and the pass reads only one of the two children, X. `resolveSourceChildren` points both delegations at X (R2-P1). The accepted ledger has `[X, Y]`.
- Evidence (scratch probe, accepted `[X,Y]`, fresh `[X,X]`, X observed with parent P): output `[X, X]`, decisions: one `replaced`, `Refused` empty, `UnresolvedLinks` empty. Y is dropped from the registry and the ledger with no unresolved evidence, and X now sits on two delegations.
- Consequence: this is exactly the partial-refresh case FIX-21 targets, and the rule does not protect Y here because a confirmed fresh child replaces the accepted link by approved rule (D2). The loss and the duplicate pre-date FIX-21; the change just does not repair them.
- Why it is not blocking: design Known limits already records R2-P1 with a follow-up. A narrow follow-up rule would be to treat a fresh child that is linked from more than one fresh delegation of the parent as unconfirmed at those positions. The no-invention and idempotence properties would be unaffected. It needs the same approval route as the other rule changes, so I leave it to the parent's disposition.

### N2. `refused_held_by_sibling` can report a link that is linked in the result (severity 1, by design)

- Trigger: an accepted link at an empty fresh position whose child is already held by another fresh delegation (for example, the child resolved at a different position). Seen in about 15 percent of my random worlds, almost all of them artificial (duplicated accepted children, shuffled order).
- Consequence: the child is counted as unresolved child evidence even though a delegation holds it, so the pass is partial once more than strictly necessary. The next pass has the child in the accepted set on its holder, and the report disappears. This matches the spec text ("another delegation of the parent already holds the child") and matches the round-2 observation. No action.

### N3. Retention is skipped at the CLI stage when the accepted ledger is empty (severity 1)

- `capture.RetainAcceptedChildLinks` returns early for `len(accepted) == 0`, so the CLI stage does not apply the contradicted-child clearing on a first publication. The registry stages (owner commit, annotation, regroup) call the domain rule even without accepted facts and clear contradicted fresh links, so the practical difference is limited to evidence only present in `Known` at the CLI stage. Not a regression; note only.

## Verification performed

- Read: `domain/child_links.go` in full; `capture/reconcile.go` (`refreshTicketOwners`, `finishCaptureAnnotations`, `addUnresolvedChildLinks`, `RetainAcceptedChildLinks`, diagnostic handling), `archive_projection.go` call, `children.go` `pendingChildLinks`, `store.go` `IsRegistryContention`, `cmd/agentic-lens/capture.go` helper and its `main.go` call site, the codex resolver and occurrence code, `adapters/fact_ledger_order.go`, the retention spec, `design.md`, `tasks.md`, the round-2 data reports and the round-2 section of `REVIEW.md`.
- Ran, in the shared checkout and without editing it:
  - `go test ./internal/domain -run TestRetainAcceptedChildLinks -count=1`: pass.
  - `go test ./internal/adapters/capture -run 'TestRetainAcceptedChildLinks|ChildLinks|AmbiguousGroup|InterruptedAnnotation' -count=1`: pass.
  - `go test ./cmd/agentic-lens -run 'TestCaptureAuto(Keeps|Aligns|Reports|Child)|ChildLink' -count=1 -v`: three tests pass (`KeepsCommittedChildLinksWithoutCaptureState`, `AlignsAPartlyResolvedGroupWithTheCommittedLedger`, `ReportsOnlyClearedLinksWithoutPartial`).
- Scratch probes (scratchpad `data-sonnet-r3`, a copy of `child_links.go` plus type stubs, no repository file touched): property probe of 300k random worlds (no invented or misplaced link, no introduced duplicate, no contradicted child, idempotence, complete groups produce no ambiguity refusals, retained links present, refused links not linked except held-by-sibling, no silent loss); feedback stability over 200k worlds; grow-only multi-pass simulation over 100k histories; the R2-P1 interaction probe (N1); an example dump for `held_by_sibling` (N2).
- Reused `evidence/TESTS.md` for race, large-store and mutation results. No package suite was rerun.

## Gaps

- The busy-exit mapping for `DeadlineExceeded` lives in a helper that calls `os.Exit`; no test reaches it, and I did not exercise it. I traced the code only.
- I did not rerun `-race` for the contention test (R2-T1); I relied on the recorded results.
- The probes model the domain rule in isolation. I did not build a registry fixture for N1 or run a two-machine ledger exchange with clock skew. A cross-machine clock skew that makes one machine's cutoff drop a trailing delegation would present as a smaller group and refuse that whole key group; I judged it unlikely and did not test it.
- I did not exercise OpenCode, the gz or header-cache changes, cursor semantics or the discovery window (outside my focus).
