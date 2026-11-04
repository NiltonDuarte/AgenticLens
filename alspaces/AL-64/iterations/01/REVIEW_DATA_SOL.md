# AL-64 data and failure review

Verdict: CHANGES_REQUESTED. Independent first-pass discovery, data/failure focus only. Parent owns consolidation and delivery.

Reviewed revision: AL64-candidate-01, working application/test diff over a84adbe5be46c2ef52ede0987a417562715d38e1 on fix/AL-64-active-turn-duration. Requested/inherited identity supplied by parent: reviewer-data, gpt-6.1-sol, high. Independent provider usage metadata was unavailable; no cost claim is made. All 46 files in /tmp/al64-review-01/fingerprint.json matched their recorded SHA256 values; manifest SHA256 6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695. This includes the actual new duration source/test files. No sibling review reports or conclusions were read. Product and test files remained unchanged; probes used disposable Go overlays under /tmp.

## Blocking findings

### DATA-SOL-1: invalid Codex completion is certified as complete coverage

Severity: 5/10. Introduced evidence defect; dropping the malformed interval numerically was pre-existing.

Location: backend/internal/adapters/codex/rollout.go:440-448, propagated by finalizeFacts at 932 and domain.MeasuredDuration.

Trigger: a known turn start has a completion timestamp earlier than its start (also reachable with an unparsable completion timestamp). The completion branch rejects the elapsed measurement but deletes the pending start without recording durationLimit or a warning. A valid earlier turn remains in the accumulator.

Consequence: emitted and persisted duration_evidence claims state complete, with no limitation and no provisional component, despite an unsupported interval. Workflow/API/UI consumers consequently present the supported subtotal as wholly completed coverage. A complete parser result also passes the existing owner-replacement warning gate. The numerical subtotal can remain valid; its claimed coverage is false.

Expectation: retain known valid elapsed measurements, exclude invalid elapsed measurement, and explicitly report partial/unsupported coverage. This follows CONTRACT.md: ambiguous, negative or invalid sequences preserve supported known duration and report their limitation. The existing interruption branch already reports invalid terminal boundaries, which does not guard this completion branch.

Independent probe: go test -overlay /tmp/al64-data-overlay.json ./internal/adapters/codex -run TestAL64DataInvalidLifecycleProbe -count=1 -v. A valid 00:00:01-00:00:05 turn followed by start 00:00:10 / completion 00:00:09, cutoff 00:00:20, emitted total=4, State=complete, CompletedSeconds=4, Limitation empty, warnings empty. The assertion requiring partial coverage failed. Probe source: /tmp/al64-data-codex_test.go; no production mutation.

Required correction: record the invalid completion limitation before removing that interval, and cover the complete-to-partial evidence transition with a focused regression.

### DATA-SOL-2: explicit OpenCode fallback ignores the operation cutoff

Severity: 4/10. Introduced propagation omission in an unchanged helper.

Location: backend/internal/adapters/backfill/recovery.go:296 -> readOpenCodeTree at 203-208 -> opencode.ExtractSessionsStrict -> extractSessions -> time.Now().

Trigger: --allow-fallbacks is enabled, all exact recorded OpenCode owners are unavailable, and the existing ticket/workspace heuristic discovers a different supported active tree. recoverOpenCodeAt receives the operation cutoff, but the fallback invokes readOpenCodeTree without it. That helper enters the clock-selecting legacy wrapper.

Consequence: newly added provisional owners use a later cutoff and larger contribution than the immutable cutoff selected for this backfill operation. Fixed-cutoff recovery is not stable across repeated attempts and duration events/completions between operation start and fallback read can leak into that snapshot. Original missing owners are preserved by the overlap guard; this finding does not allege deletion or unauthorized exact-owner replacement.

Expectation: all extraction paths in one operation, including explicitly enabled heuristic enrichment, must use its selected cutoff. Existing mixed historical cutoffs are permitted, but do not justify selecting another observation time for a newly extracted owner in this operation.

Independent probe: go test -overlay /tmp/al64-data-overlay.json ./internal/adapters/backfill -run TestAL64DataFallbackCutoffProbe -count=1 -v. A concrete SQLite v1 fixture retains a missing exact owner and resolves a distinct AL-64 root with a native running tool. Requested cutoff 2026-11-04T19:20:06.491Z was 3 seconds after assistant creation. Added owner instead used 2026-11-04T19:20:12.525588727Z and total 9 seconds. The cutoff-equality assertion failed. Probe source: /tmp/al64-data-backfill_test.go; no production mutation.

Required correction: propagate cutoff through readOpenCodeTree and its extraction call. Test the explicitly enabled fallback branch with a fixed supplied cutoff.

## Consequential hypotheses assessed

- Duplicate contributions during ordinary exact-owner refresh: refuted for inspected paths. replaceRecoveredOwners discards the previous complete matching session snapshot and preserves owners outside the refresh set; it does not union old invocation rows or add cumulative duration. Codex failed parses/warnings are excluded before replacement. OpenCode discovery/extraction/parent-linkage failures preserve the old exact owner. Existing family refresh matrix passed independently.
- Unavailable-source loss or advancing historical cutoff: refuted for the tested retention path. Domain/DTO and migration mappings carry DurationEvidence, PersistSourceAnomalySummaries and RetainIncompleteGroupSummaries change anomaly collections without reconstructing invocation fields, and aggregation reads stored evidence without a clock. The matrix preserves unavailable provisional owner/cutoff, other harness and child facts while a successful local root snapshot changes.
- Evidence loss or invented certainty through DTO/migration: round-trip/current-schema migration and legacy-no-evidence tests passed. DurationEvidence is copied into both DTO directions and migration input/domain conversion; the invariant validator compares it. Legacy absence remains absent and aggregates as unknown. ValidateDuration checks nonnegative reconciliation, states, cutoff parsing and required limitation combinations. Complete/provisional rounding occurs once per invocation and components reconcile.
- Schema family misselection for native v1 auxiliary session_message: inspected detector checks supported empty auxiliary shape and rejects populated/malformed competing families; this change does not move session ownership to an auxiliary family.
- Future completion extending measured time: DurationInterval clips a completion beyond cutoff to a provisional contribution; Codex gates future lifecycle records. Mutable SQLite snapshots are not historical event journals, and future-source observation limitations still apply.
- Persisted lifecycle state proving process liveness: not established. Native v1 SIGINT evidence in RUNTIME_EVIDENCE.md explicitly shows the same open assistant/running tool after process exit. Candidate source-state provisional policy counts it on a later read. CONTRACT.md currently requires current-activity proof and limits stale intervals; the requested user clarification remained pending at dispatch. This is an unresolved acceptance-policy gap separate from the two deterministic defects. Resolve the policy and preserve truthful uncertainty before delivery; this review does not silently approve a new liveness policy.
- Additive v6 compatibility: current reader paths accept legacy absence and new evidence. Old strict binaries reject the new member, so mixed-deployment safety requires the proposed reader-first rollout decision; this report authorizes no release or writer rollout.

## Verification and limits

Independently executed: focused production tests TestBackfillActiveDurationReplacesSnapshotsAcrossSourceFamilies and TestDurationEvidenceRoundTripLegacyAndValidation passed in adapters/backfill and adapters. The same command included domain/migration, which reported no matching tests; that package was not independently retested here. Both independent defect probes above failed for the stated contract assertions, with successful compilation and actual production-path execution.

Reused TEST_REPORT.md evidence for this frozen application/test candidate: go test -race ./... passed all 13 test-bearing packages; frontend 37 files/834 tests, typecheck, lint and diff whitespace checks passed. Reused RUNTIME_EVIDENCE.md for real native v1/v2 generation, active --recompute-db, completion replacement and unchanged-repeat output; these are not experiments independently rerun by this reviewer. No full-suite repetition, native-process experiment, frontend interaction or required post-integration make was executed here. Parent retains embedded binary/build, integration, policy resolution and delivery gates.

Scope conclusion: ordinary exact-owner preservation/replacement has meaningful positive evidence, but the two reproducible duration-evidence/cutoff defects require correction. A subsequent review must verify affected changes on a new frozen fingerprint and resolve the source-state acceptance boundary.
