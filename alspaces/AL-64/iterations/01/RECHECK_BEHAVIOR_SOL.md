# AL-64 candidate 02 behavior correction re-review

Outcome: APPROVED for the assigned correction scope, subject to the explicit pending product condition below. No new distinct finding. This is not consolidated approval or delivery authorization.

## Identity and candidate

- Responsibility: behavior and integration reviewer; correction verification, not new first-pass discovery.
- Run: recheck_behavior_sol; phase: candidate 02 correction re-review.
- Acknowledged inherited profile: gpt-6.1-sol / high, supplied by the parent assignment. Runtime model/effort introspection is unavailable; no independent provider metadata or usage claim is made.
- Baseline: a84adbe5be46c2ef52ede0987a417562715d38e1; branch: fix/AL-64-active-turn-duration.
- Frozen manifest: /tmp/al64-review-02/fingerprint.json, SHA-256 027324fa31f45b9933adb35bd34835b54918757047556354e8b2c3ce82f4c24c. Verified all 46 application/test path hashes at entry and immediately before this verdict; zero mismatches.
- Product/tests remained read-only. Only this assigned report and disposable /tmp probe artifacts were written. Existing staged and unstaged work was preserved.

## Correction dispositions

| Finding | Original severity | Disposition and evidence |
| --- | ---: | --- |
| DUR-01 | 5 | Corrected. In codex/rollout.go readEvent, a matching reversed or unparsable completion sets an explicit duration limitation and removes the invalid open interval. finalizeFacts passes the limitation into MeasuredDuration, retaining measured completed milliseconds while exposing partial coverage. The retained lifecycle table checks both invalid cases after a valid 4-second interval and requires unknown summary coverage. Supplied pre-correction overlays fail both assertions. Valid completion, future completion and interruption controls remain covered. |
| DUR-02 | 6 | Corrected. The v1 activity query uses native running-tool state.time.start, with record creation fallback when the native start is absent, and native streaming time.start. Generic progress time_updated no longer disqualifies an earlier activity start. The permanent later-progress regression discriminates original and corrected production. Supplemental probe confirms fallback and future-start rejection. |
| DUR-03 | 6 | Corrected. The v1 NOT EXISTS supersession guard requires the newer message to be created at or before cutoff. The permanent future-message regression fails the original code and passes the correction. Supplemental probes confirm legitimate newer messages before and exactly at cutoff still suppress ongoing evidence, while a newer message after cutoff preserves it. Superseded unsupported activity remains partial rather than becoming completed. |
| DUR-04 | 4 | Corrected. recoverOpenCodeAt passes its immutable cutoff through readOpenCodeTree to ExtractSessionsStrictWithSourceAt. The fallback test verifies the persisted 3-second provisional contribution and exact cutoff, retained unavailable exact owner, discovered owner and incomplete recovery status. Original-code overlay instead records a wall-clock-sized contribution. No fallback authority or owner-overlap guard was weakened. |
| DUR-05 | 5 | Corrected regression gap. Permanent family tests now independently reach v1 timed text and timed reasoning without tools, and v2 text with time_idle NULL and no tool. Advancing/fixed cutoffs, no fabricated completion, completion replacement and completed repeat are asserted. Supplied path-removal mutants fail these cases; control passes. This verifies supplied test corrections and claims no additional discovery credit. |

These are verification dispositions for the original findings retained by the parent; original discovering reviewers keep their credit. No new introduced defect, pre-existing defect or optional suggestion is reported in this bounded pass.

## Behavioral investigation

Traced explicit generation's single snapshotCutoff through Codex extraction, OpenCode extraction and registered reconciliation, and backfill Execute's cutoff through exact-owner extraction and heuristic fallback. Codex parsing gates only duration lifecycle events after cutoff, preserving unrelated token/user facts. Invalid lifecycle completion contributes neither completed nor provisional duration; known completed values survive with partial coverage.

For OpenCode v1, native part activity establishes ongoing evidence while DurationInterval retains assistant-message accounting. Missing activity proof, future activity starts and genuine supersession do not create provisional duration. Explicit message completion remains authoritative independently of the ongoing-proof query. The completion projection requires supported completed bounds and no provisional/unsupported coverage. The unchanged v2 path similarly uses explicit completion, idle/tool activity and cutoff-bounded supersession; the added no-tool test reaches the idle-null branch.

Followed backfill recovery into replaceRecoveredOwners: successful exact owners replace their snapshots; unavailable owners remain; fallback cannot overlap stored exact identities or certify unavailable historical recovery. The correction passes a cutoff without changing that authority boundary. The retained source-family end-to-end matrix saves/reloads each stage and verifies advancing/fixed snapshots, completion replacing provisional time, unchanged children, unavailable stored cutoff, other-harness facts, additive totals and no duplicate owners. Component arithmetic and summary propagation remain governed by MeasuredDuration and DurationForInvocation; no DTO/UI/aggregate implementation changed in these corrections.

## Verification

Executed from backend:

`go test -overlay /tmp/al64-recheck-behavior-sol/overlay.json ./internal/adapters/opencode -run '^TestRecheckV1ActivityCutoffGuards$' -count=1 -v`

Result: exit 0, all nine independent disposable cases passed. Cases cover legacy tool creation fallback; native tool start after cutoff; absent native start with future row creation; newer message before, at and after cutoff; streaming progress after cutoff; future streaming start; and closed text without ongoing proof. Every result also validates coherent duration evidence and absent fabricated completion. The overlay appends tests to an external copy of the existing test file; shared product/test files were unchanged.

Reused TEST_REPORT.md candidate 02 affected-package race results, original-code red/green overlay evidence and streaming-removal mutant evidence. Inspected the private failing logs under /tmp/al64-tester-prefix: failure outputs demonstrate the intended wrong subtotal/state/cutoff, rather than unrelated fixture setup failures. Reused candidate 01 full Go race, 834 frontend tests, typecheck and lint for unchanged scopes. No redundant full-suite execution was warranted. Native OpenCode generated/persisted active, completion and repeat observations for both pinned schema families remain separately recorded in RUNTIME_EVIDENCE.md; their narrow subsequent corrections are covered deterministically here.

## Remaining conditions and limits

- The user decision between explicitly open persisted source intervals as provisional with disclosed process-liveness uncertainty, and additional live-process proof, is still pending. This review assesses current source-state semantics only and does not decide that policy. Final delivery remains conditional on that decision and any resulting implementation/verification change.
- Native Codex lifecycle records are observed, but native Codex CLI fresh generation, active refresh, completion refresh and repeat remain unverified: the disposable whole-source copy was correctly rejected by project ownership validation. Deterministic Codex coverage does not substitute for that native boundary.
- Parent retains integrated make, final embedded standalone verification, consolidation, telemetry and publication. This report does not certify those gates.
- Review is limited to candidate 02 corrections and consequential behavior dependencies, not the whole repository or proof of process liveness. Any later application/test change invalidates the corresponding frozen-path evidence until reassessed.
