# FIX-21 architecture review (arch-opus, first pass)

- Role: reviewer-architecture (architecture and maintainability)
- Run: arch-opus, discovery phase. Requested profile: Claude Opus 5.5, medium effort. Runtime reports model claude-opus-5-5; actual effort is not exposed at runtime.
- Reviewed revision: base 133100bb .. head 0f73f759 (HEAD confirmed `0f73f7598fcbf3ee81ad9dc5c0b526b7a630d3db`). Product code was treated as read-only.
- Verdict: **APPROVED** for the architecture scope, with suggestions. No finding blocks delivery.

## Questions answered

| Question | Answer |
|---|---|
| Is link retention pure, immutable, typed and free of `any`? | Yes. `domain/child_links.go` imports nothing. It copies `Fresh` and each rewritten owner's delegations before changing them. States and outcomes are typed string enums. The diff adds no `any` or `interface{}`. The only gap is that `ChildLinkRefusedAmbiguous` decisions carry an empty `Evidence`, which is outside the typed set (cosmetic). |
| Are I/O and policy in the right layers? | Yes. The domain rule does no I/O. The capture adapter applies it in `finishCaptureAnnotations` and `regroupArchiveFamily` against `previousSnapshots`. The CLI helper `retainCaptureChildLinks` in `capture.go` does the stderr output and the diagnostic write. See S4 for one place where the layering could be tighter. |
| Is the capture-only discovery API separated from manual, backfill and live paths? | Yes. Every capture caller moved to `DiscoverExactCaptureRolloutPath`, which is used in `source_database.go` (3 calls) and `codex_context.go`. Live hydration still uses `DiscoverExactRolloutPathContext` (`live/service.go:1060`). Manual generate and backfill still use the unchanged `DiscoverRolloutPaths`/`listRolloutPaths`. `HeaderCache` and `CreationWindow` are created only in capture code. The cost is duplication (S1, S2), and one exported function is left with only test callers (S3). |
| Is the budget split and deferral model coherent? | Yes. Every byte charge goes through `chargeIdentification` or `chargeExtraction`. No direct `b.charge` caller bypasses the split, so the two shares always add up to `b.bytes` until `releaseSplit`. Identification remains capped at its share after release. The leftover refresh is the only thing that uses the released share. The rule that each owner is attempted at most once, except deferred owners, is stated where the work-class loop runs. Minor naming and comment issues are in S5 and S6. |
| Is `codexstoretest` test-only? | Yes. It imports `testing` and `time/tzdata`. Only `_test.go` files import it. `go list -deps ./cmd/agentic-lens` does not contain it. |
| Is the `main.go` footprint minimal given PR #162? | Yes. The change adds four lines at `runGeneration` around line 598: one call plus its error exit. PR #162 changes `main.go` only inside `runBackfill` (hunks starting at lines 1315, 1342 and 1377), and does not touch `codex/discovery*` or `context.go`. No textual or semantic overlap was found. |

## Findings

None are blockers. All are maintainability suggestions or existing debt made more visible.

### S1. Exact-lookup selection logic is duplicated (introduced, suggestion, 3/10)

- **Where:** `codex/discovery.go`. `DiscoverExactCaptureRolloutPath` plus `listFilteredRolloutPaths` repeat `discoverExactRolloutPath` plus `listRolloutPaths`:
  - the home and budget validation;
  - the `sessions`/`archived_sessions` walk and the suffix filter;
  - the "incomplete discovery for exact owner" rule;
  - the ambiguity rule and the unavailable rule.
- **Difference between the copies:** the new list does not claim paths while walking; it claims them per candidate.
- **Expectation:** both lookups share one selection core, with the name pre-filter and header reader injected.
- **Consequence:** a future fix to ambiguity, incomplete-discovery or store-walk semantics (for example symlink or error wrapping) has to be made twice. Otherwise manual and live lookups will silently diverge from capture lookups. `ambiguousOwnerError` is already shared, which shows the two are meant to agree.
- **Suggestion:** implement `listRolloutPaths` as `listFilteredRolloutPaths` with a keep-all filter plus claim. Factor the candidate loop over a `readID(path)` function.

### S2. Session-metadata validation now exists twice (introduced, suggestion, 3/10)

- **Where:**
  - `codex/discovery_budget.go`: `headerFromMetadata` and `RolloutHeader.SessionContext`;
  - `codex/context.go`: `readSessionContext`.
- **What is copied:**
  - the nested `subagent.thread_spawn.parent_thread_id` conflict check, verbatim;
  - the "ID and cwd required" rule.
- **Who uses which:**
  - `ReadSessionContextWithinBudget` remains in use in `capture/codex_context.go:36` and `capture/project.go:317`;
  - the capture reconcile owner preflight now uses `RolloutHeader.SessionContext`.
- **Expectation:** one function derives `SessionMetadataContext` from decoded metadata.
- **Consequence:** the hook and project paths and the reconcile preflight can disagree about whether an owner is valid if the ancestry rule changes in only one copy.

### S3. `DiscoverExactRolloutPathWithBudget` is left as test-only production API (introduced by migration, suggestion, 2/10)

- **Where:** `codex/discovery.go:492`.
- **Callers:**
  - at base, its only callers were the capture adapter (`source_database.go` ×3 and `codex_context.go`);
  - at head, its only callers are tests: `discovery_budget_test.go`, `discovery_window_test.go:317` and `reconcile_large_store_test.go:311`. The last two use it as the "before" baseline.
- **Consequence:** an exported entry point with no production consumer invites new capture code to bypass the pre-filter and header cache by accident.
- **Suggestion:** either remove it and keep the baseline inside the tests through `DiscoverExactRolloutPathContext`, or document it as the unfiltered reference.

### S4. Retention bookkeeping is split between the adapter and the CLI (introduced, suggestion, 2/10)

- **Where:** `cmd/agentic-lens/capture.go`, `retainCaptureChildLinks`.
- **What the CLI helper does:**
  - calls the domain rule twice, swapping `Fresh` and `Observed`;
  - merges the reports;
  - calls the exported mutator `ReconcileResult.AddRetainedChildLinks`;
  - re-derives the diagnostic with `ReconciliationDiagnostic`;
  - reopens the store to record it.
- **Duplication:** the adapter's `ReconcileAt` runs the same "count retained links, then mark the diagnostic partial" sequence internally.
- **Consequence:** `ReconcileResult` now carries request state (`countedChildLinks`, `reconciliation`, `retainedChildLinks`) and a public mutation API that a caller must use in the right order. A future caller that calls `AddRetainedChildLinks` without re-recording the diagnostic would persist a "completed" reconciliation while `PendingChildren > 0`.
- **Suggestion:** one capture method, for example `(*Store).RetainAcceptedChildLinks(ctx, *ReconcileResult, accepted)`, would own the domain call, the counting and the diagnostic write. The CLI helper would shrink to that call plus stderr output.
- **Why this is not a defect:** the design approved this placement (D3-A, D5), and today's single caller is correct.

### S5. OpenCode child pages remain asymmetric with Codex header reads (existing behavior, now inconsistent with the stated model, 2/10)

- **Where:** `capture/children.go`, around lines 366-394.
- **The asymmetry:**
  - Design and spec classify OpenCode child pages as identification, and they are now charged to the identification share.
  - They still do `budget.sessions--` and `budget.attempts++`.
  - Codex header reads were deliberately moved off session slots and attempts (refinement 2).
  - The `readMetadata` comment ("consumes no owner-extraction session slot and so no source attempt") is true only for Codex.
- **Consequence:** on mixed stores, "Source attempts" counts some identification work. OpenCode discovery can also starve owner extraction slots, which the reserved split was meant to prevent.
- **Suggestion:** align the two, or record the exception in the budget spec.

### S6. Smaller wording mismatches (introduced, cosmetic, 1-2/10)

1. **`.gz` headers can still be abandoned partway** (`codex/discovery_budget.go`, `ReadHeader`).
   - The admission check uses `min(key.size, 64 KiB)`, where `key.size` is the compressed size for `.jsonl.gz`.
   - `readRolloutHeader` charges decompressed bytes.
   - So a gzip header can still be abandoned partway with `ErrDiscoveryBudget`. That contradicts "a header is never abandoned part way through" and the spec's admission rule.
   - Reach: only the exact lookup, because child sweeps list only `.jsonl`. The cost stays bounded by the allowance, so there is no overspend.
2. **The deferral predicate is broader than its comment** (`capture/reconcile.go`).
   - The comment says "refused only because extraction is holding identification's reserved share".
   - `extractionReserved()` is also true for an owner that would not fit even in the released total.
   - Effect: such an owner is still deferred, sorted first in the leftover refresh, and spends a leftover session slot to fail again. This is harmless, but the comment overstates the precision.
3. **Possible PendingChildren overcount after regrouping** (`capture/archive_projection.go`).
   - `regroupArchiveFamily` appends its own retained links.
   - Links already appended by `finishCaptureAnnotations` for the same group are not withdrawn when regrouping clears or replaces them.
   - The count is deduplicated but can overstate `PendingChildren`. That is conservative: it can only keep a pass "partial".
4. **`HeaderRead.Cached` has no production reader.** It is used only by tests.

## Hypotheses investigated

| Hypothesis | Disposition |
|---|---|
| The domain rule changes its inputs (shared delegation slices). | Rejected. `append([]T(nil), ...)` copies both the fresh slice and each owner's delegations, and the `Accepted` and `Observed` maps are only read. The focused tests `go test ./internal/domain -run ChildLink` pass. |
| Some byte charge bypasses the split, letting one share borrow the other. | Rejected. Every capture call site uses `chargeIdentification` or `chargeExtraction`, and plain `charge` is called only inside them. |
| Identification draws on the released share in the leftover refresh. | Rejected. `identificationRemaining` stays capped by `split.identification`, and `recoverChildren` does not run after `releaseSplit`. |
| Capture pre-filters leak into manual, backfill or live lookup. | Rejected. See the separation answer above. |
| `codexstoretest` is linked into the engine. | Rejected. It is absent from `go list -deps ./cmd/agentic-lens`. |
| The `main.go` call conflicts with PR #162. | Rejected. The hunks are disjoint (`runGeneration` here, `runBackfill` there). |
| `regroupArchiveFamily` should use the augmented `previous`, not `group.previous`. | Rejected. Synthesized records come after the originals, and the accepted map keeps the first record per ID, so the evidence is identical. |
| Header-based `readMetadata` loses nested-parent children. | Rejected. `decodeSessionMetadata` folds `thread_spawn.parent_thread_id` into `ParentID`. |
| `any` or `interface{}` was introduced. | Rejected. A search of the added lines found none. |

## Verification performed

- `git diff 133100bb 0f73f759` for all product files in scope. I read the head versions of `domain/child_links.go`, `codex/discovery_budget.go`, `codex/context.go`, the related parts of `codex/discovery.go`, `capture/reconcile.go`, `capture/children.go`, `capture/source_database.go` and `capture/archive_projection.go`, and the CLI call sites.
- `go vet ./internal/domain/ ./internal/adapters/codex/...`: clean.
- `go test ./internal/domain/ -run ChildLink -count=1`: ok.
- `go list -deps ./cmd/agentic-lens | grep codexstoretest`: 0 matches.
- `gh pr view 162` and `gh pr diff 162`: lists the files and `main.go` hunks. The diff was saved only to the session scratchpad.
- I reused the test evidence in `evidence/TESTS.md` and did not rerun package suites.

## Gaps

- I wrote no probe for the `.gz` admission mismatch (S6.1). It is established by reading the code only.
- I did not assess test strength or the large-store test design. Those belong to other reviewers.
- I did not trace the worker's isolation semantics beyond the call site.
