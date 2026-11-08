# AL-79 independent data/failure review — data-luna-01

## Identity and scope

- **Requested profile:** gpt-6-luna / medium. The execution context identifies this run as gpt-6-luna / medium; no source-observed usage or cost data is available.
- **Run:** `data-luna-01`; independent first pass.
- **Candidate:** AL79-C2, base `4df45967c801824500ac2ea39a8f52fa9cbdcf66`; frozen source/guidance/test diff `/tmp/al79-review-coordinator-b67f2d/CANDIDATE_C2.diff`, SHA-256 `84aa66679b8b2b643ccdfd3c17148b77a369ea383a546eee372397778218a7de` (ticket artifacts excluded).
- **Requirements:** AL-79 contract and iteration 01 plan; AL-79 specification and PO-20 F1-F6; portable ledger contract and regression checklist.
- **Focus:** exact source/session/parent-child ownership, partial vs complete snapshots, portable history, concurrency/retry/deleted destinations, immutable worker artifacts, and installation transaction recovery.
- **Read-only paths inspected:** native dispatch and Codex context normalization; capture registration, reconciliation, child recovery and publication; retained executable and worker admission/launch; user installation and rollback journal; relevant frozen diff and supplied raw test/native evidence.

## Finding

### Severity 4/10 — Introduced: executable can be retained from a concurrent in-place rebuild without proving a stable complete copy

**Locations:** [retained.go](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/backend/internal/adapters/capture/retained.go:17), [capture_worker_fence.go](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/backend/cmd/agentic-lens/capture_worker_fence.go:17), [worker.go](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/backend/internal/adapters/capture/worker.go:65).

**Expectation:** An admitted worker and its later collection/publication children must use one complete immutable executable identity until normal exit. A concurrent rebuild at the mutable development path must not turn that identity into truncated or mixed bytes.

**Trigger:** The resolver validates/selects a development executable. Before or during `newCaptureWorkerFence` or `LaunchWorker` calls `RetainExecutable`, another build rewrites that same output path in place. `RetainExecutable` opens the source once and streams it into a temporary file, but neither verifies source identity/size/timestamps after copying nor validates the retained artifact's native format and required interfaces before installing/returning the digest-named path. If the copy overlaps the writer, it can retain an incomplete or mixed image as a stable digest artifact.

**Consequence:** The worker launch or a later child can fail despite the selected build having passed resolver validation. The existing worker flow appears to leave publication work unacknowledged on launch/pass failure, so this is recoverable on a later trigger, but it violates the frozen admitted-build guarantee and can strand automatic publication until that trigger. This is a concurrency reliability defect, not a claim that ordinary atomic-output builds always expose the race.

**Evidence / classification:** Static source trace through the actual worker callers and `io.Copy` implementation; introduced in the new retained-artifact path. The supplied real-process test replaces the mutable path after retention and verifies that the retained A bytes remain stable, but does not overlap a source rewrite with the copy or validate the copied artifact. No destructive source-tree probe was run. A defensive fix should establish a stable source snapshot (for example, compare identity and size/metadata before and after copying, retry/fail if changed) and validate the retained native bytes before committing the digest path; avoid claiming retention succeeded for an invalid image.

## Hypotheses investigated and disposition

- **Codex child ownership and callback identity:** No remaining counterexample found in the reviewed paths. `NormalizeCodexSubagentStart` distinguishes the parent callback ID from a source-proven child, defers unresolved child metadata, and `ReceiveCodexUser` persists child pointers for later verification. Supplied native evidence records the original failure and the corrected normal-trust root/child re-check with exact parent/child IDs and separate worktree publication.
- **Complete, partial, and unavailable source snapshots:** No defect found in the inspected owner refresh path. Failed/unavailable extraction preserves prior snapshots; accepted empty extraction remains pending when source session data has not flushed; successful exact-owner refresh recomputes associated summaries. Supplied focused tests cover smaller/empty complete snapshots, partial parsing, unavailable sources and child recovery. Full integrated suite remains parent-owned.
- **Unassigned local collection and later binding:** No defect found in the reviewed flow. Reconciliation uses the unassigned ticket scope, preserves exact source identity, does not infer ticket attribution, and publication selects bound ticket owners. The plan/evidence distinguish local collection from ticket publication.
- **Portable history:** The reviewed publication path saves through `SaveFactLedgerPreservingHistory`, and the contract explicitly preserves Git-shared facts without local receipts/source logs. No anti-edit receipt gate was identified in this focus. I did not independently rerun cross-machine fixtures.
- **Install transaction recovery / deleted destination / pending retry:** Read-only inspection found journaled rollback that refuses to overwrite bytes changed since interruption, plus separate machine-root and Codex-home locks. No concrete failure counterexample established. Windows runtime remains unavailable; no claim of native Windows transaction validation is made.

## Verification and limits

- Reused `TEST_EVIDENCE.md`, `DISTRIBUTION_EVIDENCE.md`, and `NATIVE_EVIDENCE.md`; their reported scoped Go/POSIX checks and native source-matched root/child publication are evidence for those exact candidate paths. I did not rerun them.
- Verified the frozen diff fingerprint with `sha256sum`; it matches `CANDIDATE_C2.json`.
- No source/test changes or probes were made. Full integrated `make`/race gate is pending parent; Windows runtime is unavailable and only cross-compilation is reported. The native evidence records an explicit SessionEnd visibility limit.

## Scoped outcome

**CHANGES_REQUESTED** for this data/failure focus because the retained executable copy can race a same-path rebuild and is committed without validating the artifact. Severity is bounded by the recoverable pending-publication path. This verdict covers the stated AL-79-C2 focus only and does not certify the integrated candidate or other review focuses.
