# FIX-21 behavior review, round 3: behavior-sonnet-r3

- **Run ID:** beh-sonnet-r3. **Phase:** re_review (round 3). **Focus:** end-to-end automatic capture behavior.
- **Requested profile:** Claude Sonnet 5.5 / high. The runtime reports model `claude-sonnet-5-5`; it cannot confirm the effort level.
- **Reviewed revision:** `5d4df7ee` (branch `FIX-21/fix-capture-refresh-child-links`, clean at start). Corrections delta `86eb6dae..5d4df7ee` read in full. Production files in the delta: `backend/internal/domain/child_links.go` and `backend/cmd/agentic-lens/capture.go`. Nothing else in product code changed, so cursor, `.gz` accounting and `readSessionContext` are exactly as round 2 reviewed them.
- **Verdict (behavior scope): APPROVED.** No defect that blocks delivery. Two low-severity observations below.

## Round-2 findings in my focus

| ID | Result | Evidence |
|---|---|---|
| R2-01 | Resolved | CLI probes (below): same-key group with machine B reading 2 of 3 children keeps all 3 links on three runs, stable; a group that grew by one spawn gives `[c1 c2 c3 ""]` with 1 unresolved; a group with distinct timestamps (unique keys) behaves the same. A property run over 300,000 random inputs and a 200,000-sequence append-only model (growing transcript prefix, random subset of children resolved per pass, previous output as next accepted ledger) found 0 misplaced links, 0 lost links, 0 invented links, 0 duplicated accepted children, and idempotence on every case. |
| R2-02 | Resolved | Code: an unconfirmed different fresh child at the same position takes the `movable` branch and keeps the accepted link (`kept_pending_confirmation`); a confirmed one replaces it. Focused `-run RetainAcceptedChildLinks` domain tests pass. |
| R2-03 | Resolved | Probe "B zero spawns": 3 accepted links are placed nowhere, stderr says "could not place 3", unresolved is 3 and the attempt is partial. The next run is complete. That is the documented one-time reporting limit. |
| R2-04 | Resolved | `reconcileOwnerChildLinks` appends decisions only for pairs whose link differs; a complete group with nothing to move returns no decisions. `ChildLinkRefusedAmbiguous` and `ChildParentNotEvaluated` are gone (grep finds none outside historical Steering text). |
| R2-05 | Resolved | `Added > 0` gives "...so the reconciliation remains partial."; every probe run that added links shows it. The cleared-only case (`TestCaptureAutoReportsOnlyClearedLinksWithoutPartial`) ends with "." and passes. |
| R2-06 | Resolved by reading, with a note | `errors.Is(err, context.DeadlineExceeded)` now maps to exit 75 next to `IsRegistryContention`. Probe: with another connection holding `BEGIN IMMEDIATE`, `RecordDiagnostic` returns `SQLITE_BUSY` after a fixed ~1.0 s even under a 300 ms context, and the registry open retry returns the SQLite error after 1 s. Both already map to busy. The new branch is reached only for the joined `ctx.Err()` case when the 2 s context expires first. It is correct and harmless; its end-to-end untestability is recorded in TESTS.md Limits. |
| R2-07 | Resolved | `design.md`, `tasks.md` and the specs name `capture.RetainAcceptedChildLinks`, `Known`, the file-name dedup hint, per-owner application, and the three evidence states. No removed API name remains outside the historical Steering entries. |
| R2-T1..T8 | Present and stable | Test names in TESTS.md exist at head. `-count=5 -run TestRetainAcceptedChildLinks ./internal/adapters/capture/` passed (6.8 s), including the contention test. The three CLI cases (`...KeepsCommitted...`, `...AlignsAPartlyResolved...`, `...ReportsOnlyCleared...`) plus `TestManualGenerateLeavesLinkToMissingChildEmpty` and the capture tests `TestReconcileAmbiguousGroup*`, `TestReconcileInterruptedAnnotation*`, `TestReconcileArchivedRegroupKeeps*` passed. Test strength is for the tests reviewers. |

## New findings

### BEH3-01 (severity 2, observation, pre-existing cause R2-P1): a fresh child duplicated across positions silently replaces an accepted link

- **Location:** `domain/child_links.go`, the `ChildLinkReplaced` branch of `reconcileOwnerChildLinks`.
- **Classification:** limit of the approved rule (D2 plus position alignment), driven by pre-existing R2-P1. Not a regression.
- **Trigger:** two delegations share agent type, task name and timestamp. Only one child is read, so `resolveSourceChildren` points both delegations at it (R2-P1). Accepted `[C1 C2]`, fresh `[C2 C2]` with C2 confirmed.
- **Consequence:** probe output is `[C2 C2]`, decision `replaced` at position 0, `Refused=[]`, `UnresolvedLinks=[]`. C1 is dropped with no unresolved evidence, so the pass can look complete. The pre-FIX-21 result was the same `[C2 C2]`, so nothing got worse.
- **Evidence:** scratch probe `TestProbeR2P1` in `beh-sonnet-r3/internal/domain/probe3_test.go` (scratchpad, not in the repo).
- **Suggested handling:** add one sentence to the R2-P1 limit saying the accepted link is replaced silently in that state. The follow-up that fixes `resolveSourceChildren` removes it. No code change is needed for FIX-21.

### BEH3-02 (severity 1, documentation nits)

- A child held by a sibling at an empty position is counted as unresolved while it stays linked at the sibling. TESTS.md Limits records this over-report for one pass; `design.md` Known limits does not. A fresh duplicate child can also cause it, so the limit belongs next to R2-P1.
- The `RetainAcceptedChildLinks` doc comment in `child_links.go` has one over-long line ("... and reports them. Owners of other harnesses and owners with nil"). Cosmetic.

## Verification performed

- Read the full corrections delta, `domain/child_links.go`, the retention call sites (`reconcile.go` owner snapshot, `finishCaptureAnnotations`, `regroupArchiveFamily`, `capture.RetainAcceptedChildLinks`, `cmd/agentic-lens/capture.go`, `main.go` call position), and the unchanged callers (`mergeSessions` consumers, pending snapshot sync, manual generation branch, worker isolation loop).
- Isolated copy of the domain package under `scratchpad/beh-sonnet-r3/`:
  - 300,000-input property run: inputs never mutated, no invented link, only the child field changes, idempotent on its own output, no duplicate accepted child, and every accepted uncontradicted link missing from the output is either reported or replaced at its own pair by a confirmed fresh child.
  - 200,000-sequence append-only multi-pass model (6 passes each): 0 misplaced and 0 lost links.
- Isolated `git archive` copy of `backend/` at `5d4df7ee` with a stub `web/dist`, using `t.TempDir()` stores only. Real CLI binary, machine A manual ledger committed, machine B `generate --capture-auto` run three times per case:

  | Case | Published links | Attempt / unresolved | Repeat |
  |---|---|---|---|
  | Group grew by one spawn (same timestamp) | `[c1 c2 c3 ""]` | partial / 1 | stable on runs 2 and 3 |
  | Group shrank by one spawn | `[c1 c2]` | partial / 1, then completed / 0 | c3 reported once, as designed |
  | Same size, distinct timestamps | `[c1 c2 c3]` | partial / 1 | stable |
  | Same size, nothing registered | `[c1 c2 c3]` | partial / 3 | stable |
  | Zero spawns on B | `[]` | partial / 3, then completed / 0 | reported once, as designed |

- Probe of the `RecordDiagnostic` busy and deadline behavior under a held writer lock (see R2-06).
- Reused TESTS.md "Candidate verification" for package suites, race runs, large-store and mutations M1 to M31. I did not rerun whole suites, only the focused tests above.
- No product, test or repository file was written. The only repo file I wrote is this report.

## Gaps

- R2-06's end-to-end busy exit through a real contended CLI run is still reasoned, not observed; there is no seam between reconcile and retention.
- I did not run the `-race` suites or the large-store test; I relied on the recorded results.
- The real `~/.codex` store, `~/.agenticlens` and live ledgers were not touched, so real-store incidence of the R2-P1 state is unknown.

## Verdict

APPROVED for the end-to-end automatic capture behavior at `5d4df7ee`. Position alignment pairs links correctly under append-only transcripts, reports every refusal it makes, is idempotent across passes and across the registry and ledger stages, and the unchanged manual, backfill, OpenCode and other-harness paths are untouched. BEH3-01 and BEH3-02 are non-blocking.
