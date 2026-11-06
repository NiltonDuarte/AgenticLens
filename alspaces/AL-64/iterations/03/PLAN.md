# AL-64 iteration 03: lean numeric provisional duration

Status: Approved for implementation on 2026-11-05. The user explicitly authorized this numeric-only plan; no renewed approval is needed within these boundaries. Baseline c68fbc90. Existing draft PR 121 contains the reviewed initial implementation, not this lean representation. CONTRACT.md remains unchanged; this delta records the latest acceptance boundary without quoting user messages.

## Agreed scope

Preserve active_duration_seconds as the recorded total. Add a flat provisional_seconds integer only when positive; accept input zero, but omit it on canonical writes. The provisional contribution is part of the total, not an additional amount. Require nonnegative integers and provisional_seconds no greater than active_duration_seconds. Absence means no recorded provisional amount; it does not prove all source activity completed.

Remove the proposed duration_reasoning field, configuration setting, CLI controls and console/summary explanation sections. Canonical writes contain no duration_evidence blob, duplicate completed_seconds, state/basis/cutoff objects or duration-specific aggregate cutoff/count graphs. No reasoning feature is introduced anywhere. Ordinary source parsing, identity validation and error reporting remain intact.

Summary and frontend show the recorded total and positive provisional portion with concise numeric labels. They do not parse or display reasoning. Backend API projections provide numeric provisional values; legacy conversion belongs only to backend readers. Cutoff remains explicit transient calculation input, while existing ledger timestamps retain their existing meanings.

Source-established open intervals accrue provisionally through the supplied cutoff. A process stopped without a native terminal marker can overestimate elapsed time; the provisional classification carries that accepted estimate limitation. The unapproved iteration 02 liveness/caller-proof/poke proposal is superseded. No mandatory liveness checks, session messages, status provider, daemon or decision protocol is required.

Preserve all six corrected duration defects, source-family semantics, future/invalid boundary handling, rounding, exact-owner replacement, repeated-refresh stability, unavailable-owner retention and other-harness facts. Preserve integrated AL-66 optional-category field presence/configuration and AL-77 shared-session warnings.

## Compatibility and persistence

Backend readers continue to accept the draft duration_evidence shape, validate its component/state/cutoff consistency and normalize its numeric provisional contribution. If legacy and flat values coexist, require numeric equality and reject conflicts. Canonical new writes retain numeric facts and omit the retired explanatory metadata. Do not manufacture completed-coverage claims after normalization.

Apply normalization consistently through DTOs, explicit migrations, retained-history publication and source-backed refresh. Validate retained identity and numeric facts before the target output projection; do not mutate the retained baseline, rewrite unrelated ledgers or add private receipt/source requirements. Unavailable sources preserve recorded numeric values. Complete source refresh replaces cumulative snapshots, including completion replacing provisional contribution without duplication.

Expanded readers accept original v6, draft evidence and lean numeric inputs. Existing strict-reader rollout constraints remain documented; this task does not release binaries. No duration-reasoning config key is added, and existing generation.metrics skills/delegations/anomalies settings retain their scope and defaults. Generate, registered capture and backfill use the same numeric representation without new option precedence or suppression behavior.

## Grounded implementation

Coder owns application changes; tester subsequently owns regression tests. Parent owns integration, artifacts, telemetry and PR updates. Shared-checkout agents preserve others' changes. The agentic-lens-dev implementation gate is satisfied by the explicit approval recorded above.

- Simplify backend/internal/domain/duration.go and models.go to total/provisional facts and pure numeric rollups, retaining transient arithmetic and invalid-boundary handling. Do not keep the retired rich object as a hidden mandatory parallel model.
- Update adapters/invocation_fact_dto.go, fact_ledger_dto.go, metric/HTTP projection DTOs and domain/migration decoding/validation. Confine old-wire compatibility to backend normalization; preserve numeric totals and owner identity.
- Update codex/rollout.go and OpenCode v1/v2 extraction mappings, generation/capture and backfill copies. Preserve established source boundaries and all six fixes; no new liveness policy or source lookup is required.
- Simplify adapters/formatters.go and derived duration projections to neutral recorded/provisional numeric text. Do not add a reasoning block, toggle or per-invocation narrative output.
- Simplify web/src/domain/types.ts, duration.ts, progression.ts, api/validation.ts and affected consumers to numeric provisional values. Remove their rich duration-evidence and cutoff/count dependencies; add no reasoning type, validator, rendering or controls. Preserve other UI behavior and rebuild embedded assets.

## Verification and review

Reuse prior evidence for unchanged behavior: six boundary fixes, native OpenCode generation/refresh, integrated AL-66/AL-77 behavior, full make with 863 frontend tests and standalone assets. Current native Codex generation succeeded at publication; a complete native active-to-completed refresh cycle remains unverified. Existing results do not verify the new representation.

Extend coherent tests for rounding and component invariants, positive/zero/absent provisional values, old-to-lean numeric preservation, conflicting dual forms, source completion replacement, fixed-cutoff repeat and unavailable history. Assert canonical saved JSON has no rich duration or reasoning fields, and summary/API/frontend show only numeric recorded/provisional values. Preserve optional-category presence and overlap behavior. No new reasoning-toggle or liveness test matrix is needed.

Run affected Go race, CLI, DTO/migration and frontend checks, then required integrated make and standalone embedded asset verification. Use self-contained fixtures, no live ticket workspace fixtures, domain mocks or silent skips.

All four review focuses are affected: behavior (numeric output/labels), data (legacy normalization/retention), architecture (lean contracts/pure arithmetic) and test effectiveness (discriminating compatibility/refresh oracles). Run each on gpt-6.1-sol/high and gpt-6-luna/medium against one frozen revision in capacity-limited waves. Unique reports remain REVIEW_BEHAVIOR_SOL.md, REVIEW_BEHAVIOR_LUNA.md, REVIEW_DATA_SOL.md, REVIEW_DATA_LUNA.md, REVIEW_ARCHITECTURE_SOL.md, REVIEW_ARCHITECTURE_LUNA.md, REVIEW_TESTS_SOL.md and REVIEW_TESTS_LUNA.md under this iteration. Reviewers own reports only; parent consolidates attribution and verdict.

This clarification removes the optional reasoning proposal within iteration 03 rather than creating another plan. The existing draft PR is updated after implementation and verification. Final telemetry remains delayed until that delivery step; no code changes or new verification are claimed by this planning/example checkpoint.

Completed review evidence is retained in [REVIEW_SUMMARY.json](REVIEW_SUMMARY.json) and the ticket SUMMARY.md. Detailed report filenames above describe the planned working artifacts; those reports were removed after consolidated approval under the canonical retention policy.
