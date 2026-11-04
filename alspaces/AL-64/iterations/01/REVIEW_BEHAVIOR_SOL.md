# AL-64 behavior and integration discovery review

Outcome: CHANGES_REQUESTED.

Run: review_behavior_sol; independent first-pass discovery. Assigned profile: gpt-6.1-sol/high inherited from parent; runtime model/usage metadata was not independently queried. Reviewed candidate: AL64-candidate-01, working application/test changes over a84adbe5be46c2ef52ede0987a417562715d38e1 on fix/AL-64-active-turn-duration. Fingerprint manifest SHA-256: 6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695. All 46 candidate paths matched their recorded hashes immediately before this report. Scope: duration generation, registered capture, source-backed backfill and refresh, persistence and visible duration projections for Codex/OpenCode v1/v2. Product and test files remained read-only. No sibling persona or review conclusion was loaded.

## Findings

### B-SOL-01: Future v1 messages erase activity at an earlier cutoff

Severity: 6/10. Classification: introduced correctness defect. Location: backend/internal/adapters/opencode/schema_v1_session_extraction.go:195-201, especially the NOT EXISTS newer-message subquery at line 201.

Trigger: extract a v1 assistant with a source-supported running tool at a fixed cutoff, while a later user/assistant message exists strictly after that cutoff. The subquery considers every newer message without bounding newer.time_created to the cutoff. Its running activity evidence disappears, even though that later message cannot supersede the assistant at the requested earlier snapshot.

Expected: records after the supplied cutoff must not change duration at that cutoff. Preserve the supported provisional contribution until a superseding record at or before the cutoff. Consequence: generation/capture/backfill can publish a smaller duration merely because a source grows after the operation began. This also makes the v1 cutoff semantics disagree with v2, whose later-message check explicitly compares later.created to cutoff.

Evidence: disposable overlay test TestAL64ReviewFutureMessageMustNotSuppressPastActivity calls the public ExtractSessionsStrictWithSourceAt API with assistant start 1760000001000 and cutoff 1760000011000. Before adding the future record, the invocation is 10 seconds, provisional, with that cutoff. Inserting only a user message at 1760000021000 changes the same invocation to 0 seconds, partial, limitation "unfinished activity lacks reliable current execution evidence". The fixed-cutoff equality assertion fails. No source interval or running part was modified. The newer-message guard needs to honor the cutoff; retain a regression reaching this actual extraction query.

### B-SOL-02: Invalid Codex completion is labeled as complete measured coverage

Severity: 4/10. Classification: introduced coverage-fidelity defect over pre-existing handling that excluded negative intervals. Location: backend/internal/adapters/codex/rollout.go:440-448 and finalizeFacts at lines 932-933.

Trigger: task_started has a valid timestamp and matching task_complete has an earlier timestamp, or an unparsable terminal timestamp. The guarded calculation is skipped, but the pending start is deleted unconditionally. No warning or durationLimit is recorded. finalizeFacts therefore emits state complete with zero seconds and no limitation.

Expected: invalid lifecycle sequences preserve supported measurements and explicitly report partial/unsupported coverage, as required by the duration contract; they must not appear wholly measured and complete. Consequence: CLI/dashboard descriptions say "completed measurements" despite the source interval being unmeasurable. The successful parser result has no warnings, so generation and authoritative refresh accept this misleading evidence.

Evidence: disposable overlay TestAL64ReviewInvalidCompletionCoverage parses task_started at 2026-01-01T00:00:10Z followed by matching task_complete at 00:00:05Z, cutoff 00:00:20Z. ParseRolloutAt succeeds, Warnings is empty, ActiveDurationSeconds is 0, and DurationEvidence is {State:complete, CompletedSeconds:0, ProvisionalSeconds:0, Cutoff:"", Basis:"codex:lifecycle-turn", Limitation:""}. The assertion requiring partial coverage fails. Mark the invalid completion explicitly while retaining any other known measured intervals.

## Consequential hypotheses assessed

- Repeated cumulative extraction could add provisional snapshots to completed totals: refuted in the inspected paths. Generation reconciles complete fresh owners; backfill replaceRecoveredOwners replaces exact successfully extracted owners; capture uses its exact owner snapshot. Domain workflow aggregation sums the resulting invocations, rather than adding a fresh total to the prior aggregate. Reused the parent-supplied three-family owner-retention/replacement matrix and native v1/v2 generate/recompute/complete/repeat observations. No new duplication counterexample was found.
- Unavailable sources could advance a stored cutoff or erase other owners: the unchanged portable recovery paths retain unrefreshed owners. Duration evidence is carried by InvocationFactDTO and migration copies; reads calculate summaries from persisted evidence and have no clock-based duration advancement. The supplied backfill matrix includes an unavailable owner in the same harness and another harness, verifies exact preservation, unchanged child ancestry and repeated snapshots. No new retention defect found.
- Completion after cutoff could be reported completed: Codex duration lifecycle events after cutoff are ignored, leaving the pending interval provisional. Both OpenCode families use DurationInterval to clip explicit future completions to provisional duration. The concrete v1 future-supersession exception is B-SOL-01.
- Explicit supported terminal markers could extend indefinitely: Codex abort/interruption/shutdown/thread-end clear pending starts and bound measured duration; v2 later messages suppress open evidence. Native v2 completion/cancellation supplies assistant completion and idle observations. The invalid Codex terminal case is B-SOL-02. Source-state/process-liveness uncertainty remains the policy dependency below.
- Presentation could silently turn mixed/legacy values into verified completed duration: workflow/fleet/comparison/progression summaries retain provisional totals, unknown counts and distinct cutoffs. CLI/Markdown use DurationDescription; ticket, fleet, comparison and scatter consumers use duration descriptions/labels, while invocation evidence remains available in detail DTOs. Legacy evidence absence maps to unknown invocation coverage. No additional scoped presentation regression was established. Browser layout and assistive-technology behavior were not exercised.
- Optional v6 expansion could break old strict readers: confirmed as an explicitly accepted reader-first rollout limitation supplied by the parent. New readers accept old v6 facts and preserve optional evidence. This review does not authorize release or claim old-binary forward compatibility.

## Pending acceptance decision

The original approved contract requires reliable current-activity proof and says stale activity without such proof must remain unsupported. The candidate treats open source lifecycle state as sufficient provisional evidence: an unmatched Codex start, a v1 latest-message running tool/open streaming part, or v2 open session/running tool can extend to the operation cutoff. RUNTIME_EVIDENCE.md explicitly records that SIGINT can leave v1 assistant/tool source state unfinished after process termination. The emitted provisional evidence has no corresponding liveness limitation. Consequently the original stronger current-execution reading is not established by this implementation. The parent's pending user clarification must settle whether explicit source-state provisional semantics with surfaced liveness uncertainty is accepted, or additional live-process evidence is required. This is a known unresolved contract decision, not claimed as a new independently discovered implementation defect. No assumption of user approval was made here.

## Verification and limits

Executed against the matching candidate:

- go test -overlay=/tmp/al64-behavior-sol-m8nnm6bx/overlay.json ./internal/adapters/codex ./internal/adapters/opencode -run '^TestAL64Review' -v: both counterexample assertions failed with the values recorded above. Overlay/test inputs are disposable files under /tmp; no repository test source was added.
- go test ./internal/adapters/codex ./internal/adapters/opencode -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestSchemaV[12]ActiveDurationSnapshotAndCompletion' -count=1: passed both packages. These existing tests exercise normal active advance/completion/repeat behavior and Codex abort/future completion; their oracles do not catch the two counterexamples.
- Candidate manifest validation: 46 paths checked, zero mismatches; manifest hash matches the assigned candidate.

Reused parent-provided DTO/migration/UI and owner-retention matrix evidence, and RUNTIME_EVIDENCE.md native v1 3 provisional -> 4 provisional -> 21 complete -> 21 unchanged and v2 2 -> 3 -> 20 -> 20 observations. These are supplied evidence, not independently rerun native sessions. Full Go race/frontend gates were assigned to the independent tester and were not duplicated. No full make, browser accessibility, standalone asset inspection, live Codex generation or native harness rerun was performed by this reviewer. Coverage is bounded to the assigned behavior/integration risks. Both concrete findings and the pending acceptance decision require resolution before this review can approve the candidate.
