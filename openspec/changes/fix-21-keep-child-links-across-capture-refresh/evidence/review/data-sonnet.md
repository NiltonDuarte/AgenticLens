# FIX-21 data integrity review (data-sonnet)

Run: FIX-21 discovery run data-sonnet, reviewer-data focus. Requested profile: Claude Sonnet 5.5 / high (runtime model reported as claude-sonnet-5-5; effort not independently observable).
Reviewed revision: base 133100bb .. head 0f73f759 (`git diff 133100bb 0f73f759`). Product code read-only; probes ran only in a scratch copy of `internal/domain` under the session scratchpad.

## Verdict: APPROVED (scoped to data integrity), with six low-severity findings and no blocking defect

No path was found that invents, moves or resurrects a never-accepted or contradicted link, or that loses an accepted link from the published ledger through the registry snapshot, archive family, ledger merge, pending-snapshot or restore paths, except where the approved rule intentionally refuses (ambiguity, contradiction, sibling holder). The findings are about silent or fragile edges around that rule, not violations of it.

## Hypotheses investigated

| # | Hypothesis | Disposition |
|---|---|---|
| H1 | `RetainAcceptedChildLinks` invents or moves a link | Refuted. Randomized property probe (400k cases, mixed accepted/fresh/observed worlds, unknown/duplicate keys, contradicting parents): every output link is either the fresh link as read or the accepted link of the same owner with an identical unique (type, description, invoked_at) key; no child ends on two delegations unless the fresh input already did; delegation count, order and non-link fields unchanged; nil delegations untouched; contradicted children never kept or carried. |
| H2 | Not idempotent | Refuted. Re-applying the rule to its own output (same accepted) changed nothing in 400k cases. Across-pass stability also holds: publish the output as the new accepted ledger, rerun with the same fresh read, output identical (300k cases). The held-by-sibling fixpoint terminates (bounded by len(delegations)). |
| H3 | Resurrects a legitimately removed link | Refuted for supported flows. A cleared (contradicted) link is gone from the registry snapshot and from the ledger on the same pass, so no later accepted baseline carries it. A link is carried only onto an existing fresh delegation; it never re-creates a dropped delegation row. Residual: the carry cannot distinguish "child not read" from "child permanently gone", by design (see F5). |
| H4 | Registry snapshot / archive family merge loses links | Partially open. `finishCaptureAnnotations` and `regroupArchiveFamily` both apply the rule against `previousSnapshots`/`group.previous`, and `ReconcileAt` adds the regroup links before the pending count. But the first-stage snapshot is link-less (F2). |
| H5 | `mergeSessions` / `fillMissingCaptureSessions` / `retainUnrefreshedGenerationOwners` replace a retained owner wholesale | Refuted. The CLI helper rewrites `RefreshedSessions` and `RestoredSessions` before `syncGenerationPendingSnapshots`, `mergeSessions(Restored, Refreshed)`, `completeMaterializedCaptureRestores` and the publication records, so expected facts, pending snapshots and the ledger all see the same retained owner. Unrefreshed owners keep the ledger/History DTO. `fillMissingCaptureSessions` only fills absent owners. |
| H6 | Pending snapshots or restore acknowledgement disagree with the retained facts | Refuted. `SyncPendingSnapshots` builds `accepted` from the post-helper Restored/Refreshed lists; `matchesRestorationFacts` compares the same post-helper values with the canonical ledger. Residual: pending snapshots are not an input to retention (F2). |
| H7 | Cross-machine ledger (B refreshes A's committed ledger with empty registry) loses links | Refuted. The helper uses `baseline.Current` (the committed ledger) as the accepted set, applies it to Refreshed and Restored, and counts the links as unresolved child evidence with a recorded partial diagnostic. Nothing requires private capture state. |
| H8 | Budget split, deferral and header cache lose facts or stall the pass | Refuted for facts: a deferred owner is saved as pending with "previous facts preserved", its session slot and attempt are refunded, it is dropped from `attempted` (so restore logic sees it as unattempted), and it sorts first in the leftover refresh, which runs only if sessions and bytes remain. `extractionReserved()` is true only when the reservation, not the pass total, limits extraction, so a genuinely oversized owner is not deferred forever. Identification stops at its share with discovery pending (visible). Budget errors and cancellation are never cached (`ReadHeader`), and a file that changes during a read is keyed by its new state. |
| H9 | Filename and window pre-filters drop a real child or owner | Refuted for the stated clock model. Directory rule: a skipped directory needs parent - end > 14h + 5s while a child created at or after parent - 5s lies before real local day end <= end + 12h (UTC offsets -12..+14), so no admissible child directory is skipped. Exact-lookup filter keeps same-UUID copies in sessions and archived_sessions (duplicates stay ambiguous) and every unconventional name. Residuals: renamed rollout (documented Known limit) and a backward clock step above 5 s between parent and child creation (WSL2 hosts can step; a pruned child leaves its `pending_starts` row, so the loss stays visible as pending children). |
| H10 | Archived unassigned relocation or isolation weakens attribution | Refuted. Relocation requires the exact unique `session_meta`, same registered store (`ValidateRolloutStore`/`CanonicalStore`), project proof (`validateCodexOwnerMetadata` -> `ValidateProject`) and `keepsAcceptedCaptureParent`. The row stays unassigned and no ticket is inferred. A foreign-project, ambiguous or conflicting rollout is isolated as pending with facts preserved; budget outcomes and non-not-found I/O errors are not isolated. `updateErr == errCaptureReadObservationChanged` is not reported as isolated, so the worker still treats it as a retry. |
| H11 | Cancellation or partial failure closes a sweep early or skips a candidate | No new defect. New budget exhaustion rewinds the sweep and queued cursors to the candidate that was not admitted. Cancellation behaves as before: a swallowed cancelled read is followed by a failing cursor write or `ctx.Err()` check, so the sweep is not recorded complete. |
| H12 | Retained links make coverage claim "complete" | Refuted. Adapter path adds retained pairs after `pendingChildWork`; CLI path adds them through the deduplicating `AddRetainedChildLinks` and re-records the diagnostic as partial with a higher sequence, which is what `capture_diagnostics` reads as latest. |

## Findings

### F1. A refused accepted link is dropped silently (severity 3, introduced; consequence of an approved rule)
- Location: `backend/internal/domain/child_links.go` (`reconcileOwnerChildLinks`: unknown key, `ChildLinkRefusedAmbiguous`, `ChildLinkRefusedHeldBySibling`), `backend/cmd/agentic-lens/capture.go` `retainCaptureChildLinks`, `capture/reconcile.go` `finishCaptureAnnotations`.
- Expectation: AGENTS.md rule 6 and the spec ("keep that partial coverage visible"): an accepted `child_session_id` that a partial pass cannot carry should not vanish without a signal.
- Trigger: accepted parent has two delegations with the same type, description and invoked_at (or an unknown invoked_at), both linked; a pass refreshes the parent without those children.
- Consequence: the published ledger loses those `child_session_id` values; the report records `refused_*` decisions, but nothing consumes them. The helper prints only kept and cleared counts, and `PendingChildren` does not include them, so a pass that lost links can show no unresolved child evidence unless pending starts exist. A later full read (all children present) relinks task-name results, so the loss persists only while the children stay unread.
- Evidence: probe confirmed refusal leaves both links empty; `rg 'ChildLinkRefused|Report.Decisions'` shows no non-test consumer beyond `ChildLinkCleared` counting.
- Suggestion: count refused-ambiguous/unknown-key accepted links as unresolved child evidence (or print them) so the loss is visible; contradicted and held-by-sibling are legitimate.

### F2. The first-stage registry snapshot is link-less and durable; registry-only recovery is not covered (severity 3, pre-existing exposure, not closed by this change)
- Location: `capture/reconcile.go` `refreshTicketOwners` -> `saveCaptureRead` (`current.Snapshot = RetainIncompleteGroupSummaries(previousSnapshots, compact.Sessions)`), second stage in `finishCaptureAnnotations`.
- Expectation: accepted facts, including pending-snapshot facts, survive an interrupted or concurrent pass.
- Trigger: the owner loop saves each refreshed parent (delegations from a single transcript carry no links) in its own transaction, and annotations run once after the loop. A pass timeout or kill between the two (the incident was a timeout-class store scan) leaves a registered, `Current` parent snapshot without links. The next pass takes that as `previousSnapshots`, so adapter retention cannot recover. Concurrent passes see the same intermediate state.
- Consequence: the committed ledger is still protected, because the CLI helper uses `baseline.Current` as the accepted set. Registry-only facts are not: a ticket with no ledger yet whose only copy is a pending snapshot (pending snapshots are not an input to retention) loses links on the next partial pass. Narrow conjunction of interruption, no ledger and partial pass.
- Evidence: code reading (`reconcile.go` stage-1 callback; `RetainUnrecordedCategories` only preserves nil, not empty-but-linkless, delegations). Not probed with an injected cancellation.
- Suggestion: apply the retention rule when stage 1 writes the snapshot, or include pending-snapshot facts in the accepted set.

### F3. Diagnostic write failure aborts publication with a hard failure (severity 3, introduced)
- Location: `backend/cmd/agentic-lens/main.go` call site and `capture.go` `retainCaptureChildLinks` (`OpenProject`, `RecordDiagnostic` with a 2 s timeout) -> `os.Exit(1)`.
- Expectation: transient registry contention maps to the existing busy outcome (`capturePublicationBusyExit`), as the reconcile error path does with `capture.IsRegistryContention`.
- Trigger: the registry is locked for more than 2 s while the helper records the partial diagnostic (the worker and hooks contend on it).
- Consequence: generation exits 1 before publishing, so the worker counts a failure and backs off instead of a busy retry. Accepted facts are preserved; only delivery is delayed and misclassified.
- Suggestion: route the helper error through the same contention classification.

### F4. CLI helper ignores registry child snapshots as evidence (severity 2, introduced)
- Location: `retainCaptureChildLinks` passes only Refreshed/Restored as fresh/observed and the ledger as accepted; `captureResult.Sessions` (all bound registry snapshots) is not consulted.
- Trigger: registry snapshot links P -> C (resolved earlier, ledger not yet updated) while the ledger links the same unique key to C', and C is not read this pass.
- Consequence: stage 2 reports C as unobserved and keeps C' (kept_pending_confirmation), so a registry-proven link is replaced by the older ledger link in the published result. Requires two different children for one unique (type, description, invoked_at); two-stage random probe confirms stage 2 can override stage 1 only in such divergent worlds. Not a loss of an accepted fact (C' is accepted).
- Suggestion: pass registry child snapshots as `Observed`, or document the precedence.

### F5. Retained links keep a ticket partial forever on a machine that cannot re-read the children (severity 2, suggestion)
- Trigger: cross-machine ledger where the children live on the other machine, or children deleted from the store.
- Consequence: every pass counts the links as unresolved and the status keeps showing "Recorded child evidence remains unresolved; continue capture reconcile", with no action that can clear it. Visible incompleteness is intended (D5); the lack of an end state is the only gap. No facts are at risk.

### F6. Tiny `max_source_bytes` can stall identification (severity 2, introduced)
- Location: `newBudgetSplit` (quarter reserved), `HeaderCache.ReadHeader` admission (`min(size, 64 KiB)` must fit).
- Trigger: a configured allowance below about four header sizes, so the identification share cannot admit any header; the final release only widens extraction, not identification.
- Consequence: child discovery stays pending every pass (visible via `DiscoveryPending`; links still retained). The previous flat 64 KiB charge could still read one header per pass at the same allowance.
- Suggestion: if validation does not require a minimum, let identification use leftover extraction share when extraction has nothing admitted, or document a minimum.

## Verification performed
- Read the design, the three specs, the diff of all assigned entry points and the unchanged consumers (`mergeSessions`, `fillMissingCaptureSessions`, `retainUnrefreshedGenerationOwners`, pending-snapshot sync/promotion and restore completion, `resolveSourceChildren`, worker pass accounting).
- Scratch probes (copy of `internal/domain` at 0f73f759 in the session scratchpad, product tree untouched): randomized property tests for no-invention, contradiction, duplicate-child, order/length, idempotence, across-pass stability and two-stage divergence. All properties held; the two-stage test only showed the F4 override in inconsistent worlds.
- Did not rerun package suites (reused `evidence/TESTS.md` results for domain, codex, capture, cmd, race and mutations M1-M6).

## Gaps
- F2 and the interrupted-pass behavior were not exercised with an injected cancellation; they rest on code reading.
- The WSL2 backward clock-step risk to the 5 s window is reasoned, not measured.
- Real Codex fork/resume file naming beyond the author's 1909/1909 measurement was not independently checked.
- Dashboard readers of delegations were not traced.
