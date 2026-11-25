# FIX-21 architecture re-review, run arch-sonnet-r2

- Focus: architecture and maintainability (reviewer-architecture), phase re_review round 2.
- Requested profile: Claude Sonnet 5.5 / high. Runtime model ID claude-sonnet-5-5; effort not observable.
- Reviewed revision: head 86eb6dae (change 133100bb..86eb6dae; corrections delta 0f73f759..86eb6dae).
- Scope: ARCH-02..05 dispositions, `capture.RetainAcceptedChildLinks` and the `main.go` footprint, `ChildLinkInput.Known` and ordinal grouping in the pure rule, `RolloutNameSessionID` as a counting hint, comment and dispatch accuracy.
- Independence: no `*-r2.md` sibling report was read.

## Verdict: APPROVED (scoped)

No blocking architectural defect. Four non-blocking findings (ARCH-R2-01..04) and one optional cleanup. ARCH-R2-01 (design and tasks name removed APIs) should be corrected before delivery, because `design.md` is the owning record of technical decisions.

## Round-1 disposition verification

| ID | Status at 86eb6dae | Evidence |
|---|---|---|
| ARCH-02 | Resolved | `exactSessionContext` and `conflictingNativeParent` (`codex/context.go`) serve both `readSessionContext` and `RolloutHeader.SessionContext`; `headerFromMetadata` now calls `conflictingNativeParent`. Check order is identical to the old code in both callers (first-record gate, then id/cwd, then conflict) and the error texts are unchanged. Focused codex Header/SessionContext/Conflicting tests pass. |
| ARCH-03 | Resolved | One exported entry point, `capture.RetainAcceptedChildLinks(ctx, dir, project, result, accepted)`, does retention, deduplicated counting and the partial re-record. `AddRetainedChildLinks` and `ReconciliationDiagnostic` are now unexported (`addUnresolvedChildLinks`, `reconciliationDiagnostic`), so a caller cannot reorder the steps. A second call on the same result returns `Added == 0` and does not re-record (asserted in `retain_child_links_test.go`). |
| ARCH-04 | Resolved | `ReadHeader` requires the full 64 KiB bound for `.gz` paths. The predicate (`strings.HasSuffix(path, ".gz")`) is the same one `readRolloutHeader` uses to choose decompression, so admission and charging agree. The new test and its mutation protect it. |
| ARCH-05 | Resolved, with one design-record gap | `readMetadata` comment is now Codex-specific and says OpenCode pages still use slots. The deferral comment in `refreshTicketOwners` no longer claims the reservation is the only cause. `refused_ambiguous` carries `ChildParentNotEvaluated` (a fourth evidence state). Test-only exports are left in place, and the codexstoretest import guard now enforces the production boundary. The design record was not updated for the fourth state (see ARCH-R2-01). |
| ARCH-01 | Accepted debt, recorded | Listed under Known limits in `design.md`. |

## main.go and CLI footprint

`main.go` changes by one line against 133100bb (`retainCaptureChildLinks(...)`, no error branch). The helper in `cmd/agentic-lens/capture.go` shrinks to timeout context, one `capture.RetainAcceptedChildLinks` call, busy or failure exit and one stderr line. Busy mapping reuses `capture.IsRegistryContention` and `capturePublicationBusyExit`, matching the surrounding reconcile error handling. The CLI stays free of retention bookkeeping, and the overlap with PR #162 is minimal. Dependency direction is clean: domain pure, capture adapter orchestrates, cmd formats.

## RolloutNameSessionID as a counting hint

There is one production use (`capture/children.go:696`), inside `pendingChildLinks`, and it only feeds the dedup set that `addUnresolvedChildLinks` consults. It never reaches identity, binding, registration or any owner read. Failure modes are bounded and benign:
- An unparseable name leaves the child empty, so a retained link to that same child is counted twice (over-report).
- A renamed rollout whose name UUID equals another retained child can suppress one count (under-report by one). The figure stays above zero because the queued row itself counts, so the attempt is still partial.

The helper being exported from `codex` is acceptable: capture already imports codex, and the function wraps the existing `parseRolloutName` so name parsing keeps a single owner. Verified as a hint only.

## ChildLinkInput.Known and ordinal grouping

- Typed throughout: string-typed enums (`ChildParentEvidence`, `ChildLinkOutcome`, `ChildLinkRefusal`), a typed `childLinkKey`, an int-typed alignment enum. No `any` or `interface{}` in `child_links.go`. gofmt and `go vet` are clean on touched packages.
- Evidence precedence (fresh and observed, then known, then accepted) is documented on `ChildLinkInput` and implemented once in the `evidence` closure.
- Pure rule: no I/O. Inputs are copied, not mutated. The fixpoint comment (each delegation moves at most once) matches the code.
- The ordinal rule matches the approved Steering text for the all-empty case.
- Dispatch accuracy: three adapter call sites (`refreshTicketOwners` first commit, `finishCaptureAnnotations`, `regroupArchiveFamily`) plus the CLI entry share one rule and one `Accepted` source per site. The first commit applies the rule to the single-owner read against pass-start snapshots, and the annotation re-applies it to the family read against the same snapshots. The two results are independent, not compounding.

## Findings

### ARCH-R2-01 (sev 2, introduced, documentation)

**Design and tasks name removed APIs and omit the corrections.**
- `design.md` (Coverage, "When the count is added") and `tasks.md` (lines 18 and 25) say the CLI helper calls `ReconcileResult.AddRetainedChildLinks` and `ReconcileResult.ReconciliationDiagnostic`. Both were made unexported in the corrections, and the CLI now makes one call to `capture.RetainAcceptedChildLinks`.
- `design.md` still says "Child parent evidence has three typed states", "the rule runs in two places only" (it also runs at each owner snapshot, DATA-01), and that the rule "takes the accepted sessions, the fresh sessions and the target harness" (no mention of `Known` or `Observed`).
- The dedup list omits the name-based hint (DATA-03).
- The Owners list omits `codex/context.go`, which ARCH-02 changed.
- No steering entry records the round-1 fixes (DATA-01, DATA-04, BEH-01, BEH-02, ARCH-02..05).

Consequence: a reader following the design looks for APIs that no longer exist and misses the owner-snapshot retention site and the `Known` evidence tier. This is cheap to fix and AGENTS.md requires `design.md` to hold the current technical account and steering record. Fix: update those passages and add a short review-round-1 steering note with links.

### ARCH-R2-02 (sev 2, introduced, report contract)

**`refused_ambiguous` decisions are emitted for groups that were fully resolved.** `ChildLinkReport` documents "Delegations whose fresh link is kept as read are omitted", and `ChildLinkRefusedAmbiguous` is documented as "a fresh link of that group is already set or the fresh group is smaller". The emission loop (`child_links.go` lines 346 to 353) tests `child != current` per delegation, not whether the accepted child is linked anywhere in the final group.

Probe (scratch copy of the module, domain package only): accepted `P=[C1,C2]` with one shared key, fresh `P=[C1,C2]` (both children read, nothing to carry).
- Result: links `[C1 C2]` (correct), but two `refused_ambiguous` decisions with `Evidence=not_evaluated`, and `Refused` empty.
- Same noise on a second pass over a published `["", C2]` (one `refused_ambiguous` at delegation 0 with no refused link).
- Asymmetry: untimed groups skip the loop (`if !key.timed { continue }`), so the same situation reports differently depending on whether `invoked_at` is known.

Consequence: no wrong observable behavior today, because the only production reads of `Decisions` count `Cleared` and the capture layer derives unresolved links from `RetainedLinks` and `Refused`. The report is nonetheless misleading for the next consumer, and it contradicts the documented contract. Fix (optional but cheap): emit the decision only when the accepted child is not linked in the final group, or drop the decision in favor of `Refused` alone.

### ARCH-R2-03 (sev 3, cross-focus, approved-boundary question)

**A mixed ambiguous group loses an accepted link after one pass, and the unresolved signal is single-shot.** The approved rule refuses the whole group when any fresh link in it is set. This is the case the Steering "Problem" statement called out (the ledger lost links, so later passes had nothing to report), and it remains for mixed groups.

Probe: accepted `P=[C1,C2]` with one shared key, fresh `P=["",C2]` (the pass read C2 only).
- Pass 1: links `["" C2]`, `Refused=[{P C1 ambiguous}]`, so C1 is counted unresolved once.
- Pass 2 (the published result as accepted): links `["" C2]`, no `Refused` entry. C1 is gone from the ledger and nothing reports it.

The case sits inside the primary scenario of the ticket (a pass reads the parent and only some children) whenever a group has duplicate keys or unknown invocation time. A carry of accepted position i onto an empty fresh position i, when the set fresh links agree with the accepted positions, would be safe and keep the link. This is as approved, so it is reported as an observation for the parent. Either record it in design Known limits ("a mixed group drops its unplaced accepted links after one reported pass") or confirm that the user accepted this behavior when choosing "Manter o grupo".

### ARCH-R2-04 (sev 1, introduced, design record)

**Two small inconsistencies in the rule that the design does not state.**
- A timed single delegation with a different, unconfirmed fresh child keeps the accepted link (`kept_pending_confirmation`, probe `["C1"]`). An untimed single delegation in the same situation takes the fresh child and reports the accepted link as `unknown_provenance` (probe `["X"]`). Design says "a unique key is the one-element case", which is true only for the empty-link case.
- An accepted key with no matching fresh key (for example, a changed description) has alignment `none`, so its links are dropped without any `Refused` entry (probe P3: `[""]`, no refusal). Codex transcripts append, so this is unlikely, but the signal contract ("refused links are reported") does not cover it.

Fix: state both in the design, or align them.

### Optional cleanup (sev 1)

- `alignmentUnique` is assigned (`child_links.go` line 315) but never compared; only `alignmentPosition` and `alignmentRefused` are read.
- `ChildLinkRetention.Added` is exported but read outside the package only by tests. It gates the re-record internally. Keep it for the assertions or make it unexported.
- `reconcileOwnerChildLinks` is about 135 lines with a local `candidate` type. It is cohesive and well commented, and the tests cover it, so no split is requested. A comment line in the `RetainAcceptedChildLinks` doc (line 168) is overlong and could be rewrapped.

## Verification performed

- Read: `domain/child_links.go` in full, the `capture/reconcile.go` retention paths (first-commit, annotation, `RetainAcceptedChildLinks`), `capture/children.go` (sweep cursor, `pendingChildLinks`), `capture/archive_projection.go`, `codex/context.go`, `codex/discovery_budget.go` and `codex/discovery.go` diffs, CLI helper and `main.go` diff, `design.md`, `tasks.md` and the spec delta.
- Grep of all production consumers of `Report.Decisions`, `RolloutNameSessionID`, `pendingChildLinks`, `IsRegistryContention` and the renamed methods.
- Focused runs on the real tree, all passing: `go test ./internal/domain -run RetainAcceptedChildLinks`, `go test ./internal/adapters/codex -run 'Header|SessionContext|ConflictingParent|Conflicting'`, `go test ./internal/adapters/capture -run 'RetainAcceptedChildLinks|CompressedHeader'`. `gofmt -l` clean and `go vet` clean on the touched packages.
- Isolated probes (P1 to P7) in a scratch copy of the backend under the session scratchpad; the product tree was not modified. I did not rerun whole suites, and I reused `evidence/TESTS.md` for race, large-store and mutation results.

## Gaps

- BEH-02 end to end through the CLI is not exercised (acknowledged in TESTS.md); I verified the mapping by reading the code only.
- I did not probe the capture-level behavior of ARCH-R2-03 through a worker pass. The domain probe shows the rule outcome, and the adapter wiring passes `previousSnapshots` as read.
- The first-commit retention does not add its links to `retainedChildLinks`; the annotation does. If the annotation were skipped for an owner with no transcripts, those links would not be counted as unresolved. I found no path that reaches that state for a refreshed Codex owner, but did not prove it exhaustively.
