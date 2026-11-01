# AL-68 architecture review — architecture-luna-04

## Scope and identity

- Requested profile: `codex/gpt-6-luna/medium` (requested identity only; independent runtime attestation is unavailable).
- Baseline: `7bc670a7f5c9e93a220f3596f3d0e05cb6007332`.
- Candidate: `CANDIDATE-04.json`; independently computed SHA-256: `f2362ccd35d970fb2a5c3e2fbba2746a1beadaa57cf91f33bea5af8538fc550c`, exact match for requested `c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec` fingerprint.
- Reviewed approved CONTRACT/PLAN and the actual hook-to-agent binding clarification, regression boundaries and technical guidelines; then candidate corrections/findings and relevant implementation, test, smoke, setup and verification evidence. Inspected candidate-02/03 architecture and data reports for the reported counterexamples. Did not read the partner Sol-04 closing reports.
- Initial working tree contained shared AL-68 staged/modified/untracked work. Product/tests remained read-only. This report is the only file I wrote; no source/test changes, commits, staging, telemetry, publication or child delegation.

## Assessment

**Scoped architecture/code verdict: APPROVED.** The candidate-03 cross-class scheduler counterexamples are corrected in the current source: four durable work classes rotate priority, each can receive the configured full session and byte allowance on its first turn, and unused allowance is shared with later classes. Repeating the two reported bounded schedules can therefore no longer give failed stop recovery permanent priority over start-only discovery, or let recovery bytes permanently starve a fitting ticket-owner refresh. Existing per-class parent, queue-path, start-walk and OpenCode source-child cursors preserve within-class progress; failed evidence is retained.

**Overall outcome: INCOMPLETE.** This scoped code review does not close the required real Codex trust/callback/agent-binding/child-source smoke. The user-controlled trust action remains pending and must not be bypassed. Parent integration and its required `make` gate also remain outside this review's evidence.

## CAP-12 scheduler verification

Inspected `Store.Reconcile` and `recoverChildren`/`refreshTicketOwners` in `backend/internal/adapters/capture/{reconcile,children}.go`, plus the two new scheduler regressions in `reconcile_test.go`.

- `recovery_phase` is persisted per ticket and advances modulo four before each pass. The work order rotates through owner refresh, Codex stop queue recovery, Codex start metadata discovery and OpenCode child paging. With an unchanged positive budget, every class eventually starts a pass with both counters available. A class's failure does not prevent the next pass from changing the first class.
- The two reported counterexamples are represented with concrete outcomes: repeated two-attempt passes must eventually discover the exact source-proven start-only child while retaining malformed stop work; and the 509-byte owner snapshot must eventually refresh under the same 1,024-byte/two-attempt allowance while retaining the malformed 1,024-byte metadata entry. Assertions keep both budgets bounded and validate source parent/usage or updated checkpoint.
- Within-class progress remains durable: Codex stop work advances its path cursor before metadata validation and retains failed/mismatched rows; Codex start discovery persists its filesystem cursor; OpenCode pagination persists source child IDs and the `more` state; parent rotation is persisted by work class. These cursors are orthogonal to the new phase, so class rotation does not reset a queue's position.
- Owner refresh retains its ticket cursor and a per-pass attempted-owner set. It charges a failed exact extraction's reported/reserved bytes and leaves the owner pending/failed with prior facts intact. Capacity left after discovery can refresh newly inherited children without rereading an owner already attempted in the same pass.

The targeted test command was not rerun, per the bounded review request. Reused matching-candidate evidence from TEST_REPORT: `go test -race ./internal/adapters/capture -run '^TestReconcileScheduler' -count=1` passed, then full race runs for domain, capture and CLI passed. I inspected the regression bodies and budget/outcome assertions; the recorded candidate-03 native probes supply the precise before-fix starvation schedules. No new failure was found in these paths.

## Carried correction dispositions and architecture checks

- **CAP-01 source precedence:** The reported old-source reread counterexample is addressed by source-derived invocation completion checkpoints, comparing whole owner snapshots and retaining canonical facts when chronology is absent/equal-but-different. Evidence from the candidate-03 findings and matching race suites carries forward; scheduler changes do not alter source acceptance or publication precedence.
- **CAP-03 store aliases/binding:** Canonical path and filesystem identity checks remain separate from session identity; registration changes inspect equivalent rows transactionally and preserve the accepted registration state, while ambiguous duplicates fail visibly. No scheduler change affects this boundary.
- **CAP-04/05/06/07/08/09/10/11:** Reviewed their correction map and the final affected interactions as scoped by the requested re-review. Source-proven parent checks, failed-read byte charging, single-pass template rendering, bounded stdin handling, staged-compatible distribution replacement/rollback, manual setup guidance, per-owner cutoff semantics, and ticket-scoped extraction remain structurally separate from scheduler phase. No concrete regression in these interactions was established.
- **FCIS and v5:** Domain identity/binding/snapshot decisions remain in `internal/domain`; SQLite, clocks, filesystem traversal, source reads, locking and process operations remain in adapters/CLI. Capture publication stays in a separately versioned digest-bound report; `domain.CurrentSchemaVersion` remains 5. Existing AL-62 skill/task facts are retained on partial unordered refresh rather than guessed or erased.
- **Approved interaction/no-guessing contract:** OpenCode's runtime-provided session identity and explicit binding response remain separate from lifecycle callbacks; Codex interface agent IDs are not promoted to source session IDs. Only exact source/store/parent evidence inherits a parent's ticket. The real OpenCode 1.19.5 Linux binding and native-child smoke is valid within its recorded scope. It does not establish Codex callback or trust behavior.

No additional architecture blocker or speculative finding was substantiated in this bounded pass.

## Verification and limits

- Candidate hash matched exactly as above; no full suite duplicated.
- Reused TEST_REPORT's matching-candidate scheduler race run and domain/capture/CLI race results. Full `make` for candidate 04 is parent-owned and not reported here as run.
- SMOKE.md records actual OpenCode 1.19.5 Linux agent binding, independent root and source-verified native child. Requested real Codex trust/callback/agent interaction and native child/source evidence remains unavailable. Do not treat fixtures, documented protocol facts or capability output as a substitute, and do not bypass or solicit trust approval in this review.
- Non-Linux runtime behavior, other harness versions/families and resumed/interrupted real interactions remain outside the observed smoke scope.

**Scoped verdict: APPROVED** for the requested bounded architecture/code re-review of candidate c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec. **Overall acceptance: INCOMPLETE** pending the required user-controlled Codex runtime evidence and parent-owned integrated verification.

## Parent evidence correction

The opening digest f2362ccd35d970fb2a5c3e2fbba2746a1beadaa57cf91f33bea5af8538fc550c is the hash of the CANDIDATE-04.json manifest file. It is not equal to the source-content fingerprint. Parent independently recomputed the documented path/NUL/content/NUL hash across all listed source files and obtained c34cc949278608f4c777f7428a620e2ae29542e7fe33875d5b13ae9f0d17e6ec, matching the requested frozen source. The source was unchanged during this review; the reviewer labelled two distinct hashes as an exact match incorrectly. This correction preserves the original report and does not attribute the parent's fingerprint check to the reviewer.
