# AL-64 candidate 02 architecture correction review

Verdict: CHANGES_REQUESTED. DUR-01 through DUR-04 are corrected in the inspected candidate. One distinct invalid-start path still falsely certifies complete coverage. This is a scoped correction review, not delivery approval.

## Identity, revision and scope

- Reviewer responsibility: architecture and maintainability; correction phase. Parent task: recheck_arch_sol. Local review identity: AL64-candidate02-architecture-sol-correction.
- Model/effort: gpt-6.1-sol / high, inherited and acknowledged from parent dispatch. Independent provider/runtime metadata and usage/cost are unavailable; no further identity is inferred from the persona name.
- Baseline: a84adbe5be46c2ef52ede0987a417562715d38e1; branch fix/AL-64-active-turn-duration.
- Frozen candidate manifest: /tmp/al64-review-02/fingerprint.json, SHA256 027324fa31f45b9933adb35bd34835b54918757047556354e8b2c3ce82f4c24c. Independently recomputed all 46 listed file hashes immediately before this verdict: 46 matched, zero mismatches. Candidate product/tests were not edited.
- Inspected the three correction diffs, relevant Codex parse/finalization/source-replacement callers, v1/v2 extraction, domain duration accounting/validation, and OpenCode fallback/exact-owner recovery. Read the assigned persona/shared context, canonical review skill, technical/regression standards, contract/plan and parent-supplied investigation/test/runtime evidence. No sibling detailed review reports were loaded.
- Only this report is written in the checkout; disposable probes reside under /tmp/al64-arch-recheck-sol. Initial shared status included the existing AL-64 application/test and workspace changes; they were preserved. No children, implementation, commits, telemetry or publication.

## New distinct finding

### DUR-06 candidate: invalid Codex start silently becomes complete coverage

Severity: 5/10. Type: introduced coverage-contract defect in AL-64; discovered during this correction review by architecture reviewer gpt-6.1-sol/high. This is distinct from DUR-01's invalid completion after a valid start.

Location: backend/internal/adapters/codex/rollout.go:411-427 and :453-454; projected at :936 and accepted by SessionFacts at :981.

Trigger: after a valid measured four-second turn, a new task_started/turn_started has a missing or unparsable timestamp, followed by a valid completion with the same turn ID. At start, seenTurns is set before timestamp parsing. A parse failure neither records the start nor sets durationLimit. The subsequent completion sees no start but also sees the ID in seenTurns, so it bypasses the orphan-completion limitation.

Observed consequence: ParseRolloutAt returns no error or warnings and emits active duration 4, state complete, completed seconds 4, empty limitation. The supported subtotal is preserved, but the unmeasurable additional lifecycle disappears from coverage. DurationForInvocation therefore records no unknown contribution. Read/replacement callers only reject actual parse failures/warnings; this path can publish or replace a source owner's snapshot with falsely complete evidence.

Expected: retain the known four seconds and mark partial coverage with a limitation for the unmeasurable start. The contract expressly requires invalid lifecycle sequences to preserve known duration and report their limitation. The caller should not need to reconstruct parser-private seen/start state to enforce that distinction.

Evidence: a Go test overlay replacing only rollout_test.go with a disposable copy plus TestRecheckStartBoundaryCoverage. Both malformed_start (timestamp not-a-date) and missing_start (empty timestamp) failed the expected partial/4 assertion and logged complete/4 with warnings=[]. The valid_start_control passed complete/9. Command from backend:

    go test -overlay /tmp/al64-arch-recheck-sol/overlay.json ./internal/adapters/codex ./internal/adapters/opencode -run 'TestRecheckStartBoundaryCoverage|TestRecheckV1ActivityProofBoundaries' -count=1 -v

Exit 1, attributable only to the two invalid-start cases. No shared source was mutated. Fix the start-parse failure branch and add a focused regression in the existing lifecycle table; retain valid completion, duplicate-ID and known-subtotal behavior. Parent was notified and retains correction ownership.

## Correction dispositions and architecture assessment

| Finding | Disposition | Evidence and side-effect assessment |
| --- | --- | --- |
| DUR-01, severity 5 | Supplied invalid-completion defect fixed | Matching completion with parse failure or end before start sets durationLimit before deleting the open start. finalizeFacts passes that limitation to pure MeasuredDuration, so the known completed subtotal remains and state becomes partial. Corrected tests explicitly assert subtotal 4, limitation and unknown summary coverage; tester's original-code overlay failed both cases and corrected control passed. The distinct invalid-start branch above remains open. |
| DUR-02, severity 6 | Fixed | v1 proof now bounds a running tool by native state.time.start, falling back to part.time_created only when native start is absent. Generic part.time_updated is no longer a proof cutoff. Existing corrected late-progress regression reaches this query. My disposable probes additionally passed legacy absent native start with late update (10 provisional), explicit zero start (0 partial) and future native start (0 partial). |
| DUR-03, severity 6 | Fixed | Supersession NOT EXISTS now considers only newer messages created at/before cutoff. Existing future-newer-message regression exercises retained activity. My disposable controls passed newer-before-cutoff (0 partial) and newer-after-cutoff (10 provisional), showing legitimate supersession still terminates proof. Completed assistant intervals remain measured independently of runningMessages. |
| DUR-04, severity 4 | Fixed | recoverOpenCodeAt passes the supplied operation cutoff into readOpenCodeTree, which calls ExtractSessionsStrictWithSourceAt. The shared extraction dispatcher passes the same immutable value to both schema families. Engine selects its cutoff once. Existing fallback test asserts exactly 3 seconds and the fixed recorded cutoff; tester's pre-correction overlay fails this assertion. Exact-owner capture still uses CaptureSessionSnapshotAt; fallback overlap guards and unavailable-owner preservation are unchanged. |
| DUR-05, severity 5 | Added protection; no architecture blocker | Read the retained no-tool v1 text/reasoning and v2 idle-null streaming cases and tester path-removal evidence. These exercise the native activity proof paths and completion replacement without requiring a running tool. Reused their passed corrected controls and failed path-removal mutants; no redundant suite was run. |

The corrected responsibilities remain narrow: adapters parse native source-specific activity and timestamps; the shell supplies operation cutoffs; domain functions receive values and remain deterministic. DurationEvidence persists a compact contribution/state/basis/cutoff/limitation value rather than an event journal. MeasuredDuration preserves the existing invocation rounding boundary and reconciling components. No new clock, I/O, outer-layer dependency or mutable source cache was introduced in the domain correction. No new abstraction is warranted by these fixes. The remaining invalid-start finding belongs in the existing parser transition rather than a second caller-side validator.

## Executed and reused verification

Executed independently:

- All 46 frozen manifest hashes and manifest digest matched.
- Disposable boundary overlay above: five OpenCode activity-proof cases passed; valid Codex start control passed; malformed/missing start cases failed their required partial-coverage oracle and substantiated DUR-06. Private fixtures/overlay are /tmp/al64-arch-recheck-sol/{rollout_test.go,v1_test.go,overlay.json}.
- Static trace of corrected completion, native proof, supersession and fallback cutoff through consumers and exact-owner replacement.

Reused from TEST_REPORT.md and the supplied correction evidence:

- Corrected Codex/OpenCode/backfill focused controls and full affected package race suites passed on unchanged candidate02 production/test content.
- Original-code overlays fail DUR-01 through DUR-04 oracles; both no-tool streaming path-removal mutants fail DUR-05 oracles.
- Candidate01 full Go race suite, 834 frontend tests, typecheck and lint remain relevant to unchanged scopes. No UI/DTO/aggregation correction was made in candidate02, so a redundant full suite was not run here.
- Native OpenCode v1/v2 live generation, active source refresh, completion and unchanged repeat are recorded runtime evidence, not independently rerun by this reviewer.

## Remaining limits and delivery conditions

- DUR-06 is unresolved in this frozen candidate and blocks this scoped approval.
- User clarification on explicitly open persisted source intervals versus extra live-process proof remains pending. This review assesses the current source-state provisional implementation and does not decide that product policy or certify process liveness. The eventual explicit user decision remains a delivery condition independently of DUR-06.
- Native Codex lifecycle shape was observed, but its isolated native generation was blocked by correct project-ownership validation. Native Codex generation/active-refresh/completion/repeat remains unverified; deterministic lifecycle and engine coverage does not replace that runtime result.
- The parent retains integrated make, final standalone embedded-asset verification, consolidated approval and publication. Reader-first optional v6 expansion remains the accepted compatibility boundary; this report authorizes no writer rollout or release.
