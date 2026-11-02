# GEN-16 iteration 02 data/failure re-review

Verdict: **CHANGES_REQUESTED**. The supplied GEN16-01 and GEN16-02 corrections are sound in the inspected paths, but the new source-role projection has an adjacent partial-failure regression.

Run: `data_sol_6_1_high_recheck`; phase: `re_review`; responsibility: data integrity, source ownership, failure handling and snapshot consistency. Requested/configured identity: `gpt-6.1-sol`, reasoning effort `high`, provider `codex`; source-observed model/effort metadata was not independently validated. This run does not claim discovery credit for the supplied GEN16-01/02 findings.

Base: `99fbcd7be0d0ed05639e4a0cb3be8744a3d5371e`. Reviewed candidate: `7b720b474b53e3fecc292b8c7de0ea4575dd00282c12da63ea8d2c3b08f06b06`, independently reproduced using the REVIEW_RUNS method: SHA-256 of `git diff HEAD --binary -- backend web`, followed by sorted untracked backend/web path bytes, NUL and file contents. Requirements: CONTRACT.md with controlling iterations/02/PLAN.md unknown-child delta. Initial Git status contained the parent/coder/tester changes; they were preserved. Product and tests remained read-only. Only this assigned report was written in the checkout.

## Blocking finding

### D-SOL-R1 — P2: failed persona metadata reads can erase a committed event role

Locations: `backend/internal/adapters/live/service.go:943`, `:962`, `:838`; `backend/internal/domain/events.go:133`.

Trigger: a committed OpenCode SessionFact has a recorded event with a known persona and no nonblank invocation role to populate `sessionRoleMap`. The local session and trace exist, but `SELECT agent FROM session` fails. A degraded source with the agent column unavailable provides a deterministic example. Empty invocation collections are an accepted fact shape, including in the existing hydration-failure fixtures.

The failed metadata query is silently omitted from `sessionAgents`, so that session never enters persona resolution. Its successfully fetched messages are nevertheless converted to local events with an empty role. The new `ReconcileSourceEvents` call treats that empty role as authoritative and replaces the matched committed persona. The response reports `IsHydrated=true`. Thus unavailable persona evidence is presented as a successful Unknown resolution, violating preservation of committed facts when optional hydration fails. The ledger file and input are not mutated; the observable defect is the hydrated read projection.

Disposable probe: `/tmp/gen16-data-recheck-f7p4_ffx/overlay.json` appends `TestReviewMetadataReadFailurePreservesMatchedEventPersona` to an external copy of the existing live SQLite tests. It creates a real isolated source containing session `owner` and assistant message `msg`, removes `session.agent`, and supplies a committed matching event `opencode:message:msg` with role `reviewer` and no invocations. The expected retained persona becomes empty.

- Current candidate: `go test -overlay /tmp/gen16-data-recheck-f7p4_ffx/overlay.json ./internal/adapters/live -run '^TestReviewMetadataReadFailurePreservesMatchedEventPersona$' -count=1` **failed** in 0.057s, with one matched event, `Role:""`, `IsHydrated:true`.
- Same probe, changing only the hydration call back to ordinary `ReconcileEvents` through `/tmp/gen16-data-recheck-f7p4_ffx/overlay_generic.json`: **passed** in 0.051s. This isolates the newly introduced replacement behavior; it is not a proposed fix because ordinary reconciliation would reintroduce GEN16-01.
- Current correction with the metadata column intact, through `/tmp/gen16-data-recheck-f7p4_ffx/overlay_control.json`: **passed** in 0.054s.

Required correction: distinguish metadata-read failure from a successful read resolving to Unknown. Failed persona reads must preserve matched committed roles or fall back for that source/session; successful blank/conflicting source resolution must continue to support GEN16-01. Add the partial-failure case to a coherent durable adapter test. Do not use a blanket nonblank-only overwrite, which would prevent valid conflict-to-Unknown projection.

Discovery actor: data/failure re-reviewer, configured `gpt-6.1-sol/high`; source-observed execution identity remains unvalidated. New candidate ID is local to this run; the parent owns consolidation.

## Supplied corrections and adjacent hypotheses

- GEN16-01: source-persona overwrite is explicitly opt-in. Ordinary `ReconcileEvents` keeps its prior behavior. The shared reconciler retains stable-ID matching, conservative unique legacy occurrence matching and unmatched occurrence preservation; role replacement is a value assignment on the copied event. Inspected the updated stable-ID, legacy unique, appended, conflicting-to-Unknown, raw explicit precedence and input-byte preservation assertions. The new failure above concerns whether a caller has established successful persona evidence, not the reconciler's requested replacement semantics.
- GEN16-02: `traceSnapshot` contains both the trace and evidence derived from the same `rawMsgs`. `GetSessionTree` retains this pair instead of doing a second cache lookup. Global invalidation replaces the cache map while already accepted snapshot pointers remain reachable and unchanged. `traceMu` still serializes revision/cache operations; construction does not mutate prior entries. Inspected the deterministic other-session invalidation regression and public unchanged-reuse/in-place-update/failure regression.
- Ownership/fallback: source partitioning, Codex expected-root/child validation and failure fallback remain present. Matched updates leave IDs, topology, timestamps, tokens and costs under their existing ownership; unmatched events remain present. Metadata reads and source trace reads require distinct success handling, as D-SOL-R1 demonstrates.
- GEN16-03: global source-revision invalidation and reconstruction after unrelated changes are pre-existing baseline behavior. The correction derives evidence from the already fetched raw messages and adds no independent history read. This review makes no per-session changed-only polling or performance claim.
- Orphan topology: reviewed the production builder test's new retained-parent assertions; snapshot and event corrections do not change ancestry.

## Verification and limits

Executed from backend: `go test ./internal/adapters/live -run 'Test(HydrationFailuresRetainRecordedEventsAndLedgerBytes|HydrateWorkspaceEvents_(MixedSources|CodexFailsClosedWhenAnyRootIsMissing|CodexFailsClosedWhenChildRolloutIsMissing|SourceLessLedgerStaysUnhydrated|HostileDegradationAndFailureModes)|LiveTraceSnapshotReuseAndInPlaceInvalidation)$' -count=1` **passed**, 0.293s. The `MixedSources` branch of this expression does not match the actual mixed-source test name; that test was inspected and prior valid evidence reused, not claimed executed here. `git diff --check` passed.

Reused TEST_REPORT correction evidence: targeted race run for paired snapshots, matched hydration roles, conservative reconciliation and orphan ancestry passed. Its before-fix hydration overlay failed and corrected fixture passed. Parent-reported original hydration probe passed; bounded 20-poll concurrency probe under race passed in 1.889s. The original 2,000-poll/race/count=3 attempt timed out at 40s and is not a pass. These supplied results were not independently rerun here.

Unchanged extractor/frontend verification is reused. No full suite was duplicated. Integrated make, rebuilt standalone embedded assets, source identity validation and publication remain parent-owned gates. This scoped verdict does not approve the whole repository or authorize delivery. Disposable probe files are outside Git; the report records their actual commands and limitations.
