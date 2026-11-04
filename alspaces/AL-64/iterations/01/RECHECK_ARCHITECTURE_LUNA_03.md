# AL-64 candidate 03 architecture correction review

Verdict: **APPROVED for the bounded architecture correction scope**, conditional on the separate pending user decision about source-state provisional intervals versus process-liveness proof. This is not consolidated delivery approval.

## Identity, revision and scope

- Responsibility: architecture and maintainability correction re-review. Run `recheck_arch_luna_03`; candidate `AL64-candidate03`.
- Requested and acknowledged identity: gpt-6-luna / medium, as specified in the review assignment and current runtime. No independent provider metadata, usage or cost is available.
- Baseline: `a84adbe5be46c2ef52ede0987a417562715d38e1`; branch `fix/AL-64-active-turn-duration`.
- Frozen manifest: `/tmp/al64-review-03/fingerprint.json`; digest `f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49`. I independently verified the digest and all 46 listed file hashes before the verdict: zero mismatches.
- Read the assigned architecture persona, shared review context, canonical review procedure, technical guidelines, contract, plan, investigation, test report, and runtime evidence. The candidate is a correction review, so supplied finding/evidence records were read and checked against code. No sibling review conclusions were used as a substitute for source inspection.
- Scope is DUR-01 through DUR-06 correction behavior and relevant source/refresh architecture. No product or test files were edited. Parent owns integration, final consolidation, delivery, and the pending policy gate.

## Findings and dispositions

No new distinct architecture finding was demonstrated. The following supplied findings are confirmed as corrected in the inspected candidate; discovery credit remains with the original discovering roles.

| Finding | Severity | Disposition |
| --- | ---: | --- |
| DUR-01: invalid Codex completion could silently certify a known subtotal as complete | 5/10 | **Fixed.** A matching invalid/reversed completion records a limitation rather than extending or certifying the interval. Valid completed milliseconds remain in the subtotal; evidence becomes partial and summary coverage remains unknown. |
| DUR-02: v1 generic progress update could hide native open activity | 6/10 | **Fixed.** The v1 activity query uses native running-tool start (with source record creation as the legacy fallback) and timed streaming start, bounded by the operation cutoff. Generic update time does not supersede the native activity boundary. |
| DUR-03: a post-cutoff v1 message could suppress an earlier open interval | 6/10 | **Fixed.** The newer-message predicate includes `newer.time_created <= cutoff`; future rows cannot alter the earlier snapshot. |
| DUR-04: fallback OpenCode recovery could use a different cutoff from the operation | 4/10 | **Fixed.** The engine's immutable operation cutoff reaches exact-owner extraction and explicitly enabled fallback tree extraction, including source event extraction. Exact-owner replacement and fallback overlap protections remain in place. |
| DUR-05: no-tool streaming paths lacked v1/v2 active evidence guards | 5/10 | **Protection verified.** v1 timed text/reasoning and v2 `time_idle IS NULL` support are present. Supplied path-removal mutants make their corresponding tests fail, discriminating the behavior from an inert test. |
| DUR-06: missing/malformed Codex start could make a completed subtotal appear complete | 5/10 | **Fixed.** The invalid-start branch now records an explicit limitation without inserting a start or inventing an interval. A later same-ID completion cannot erase the limitation; known completed duration and completion remain, state is partial, cutoff is absent, and unknown coverage is counted. |

The candidate02-to-candidate03 change is confined to the DUR-06 parser limitation branch and its regression cases. Source tracing confirms the new branch only runs when start parsing fails inside the unseen-turn identity guard. Valid duplicate identity behavior remains unchanged. The future-lifecycle cutoff gate executes before mutation, so future starts remain excluded. Finalization sums only parsed starts; the limitation cannot manufacture provisional milliseconds. The prior supported completion is retained and the normalized session still carries the partial evidence.

## Architecture assessment and verification

The correction keeps interval arithmetic in the clock-free domain function and lifecycle proof in the Codex/OpenCode source adapters. The backfill and generation shells carry one explicit cutoff through source reads. The DUR-06 fix reuses the existing evidence limitation channel and does not create another lifecycle state or caller-side reconstruction. I found no new layering, mutable-state ownership, or data-portability issue in the bounded changes.

I inspected the retained lifecycle tests and supplied disposable probes. The reviewer-side verification record for DUR-06 reports nine lifecycle table cases and six disposable controls passing through normalization; tester evidence shows both new regressions fail with the four-line fix removed and the corrected Codex race suite passes. I reused the reported affected Codex/OpenCode/backfill race checks, candidate01 full Go race/frontend 834/typecheck/lint checks for unchanged scopes, the red/green overlays for DUR-01 through DUR-04, and the DUR-05 streaming path-removal mutants. I did not repeat full suites or native runtime probes because those scopes were unchanged by candidate03 and reruns were not needed to evaluate the architecture correction.

## Remaining limits

- User clarification remains pending on whether persisted source-state intervals are acceptable as provisional with disclosed process-liveness uncertainty or require extra live-process proof. This review does not decide that policy; final delivery approval remains incomplete until it is resolved.
- Native OpenCode live generation, active refresh, completion refresh, and unchanged-repeat evidence is documented. Native Codex source generation/refresh remains unverified because the isolated copied source was correctly rejected by project ownership validation. Deterministic tests do not replace that runtime check.
- Integrated `make` and final standalone embedded-asset verification remain parent-owned and pending. The reader-first v6 compatibility limit remains: old strict readers reject the new evidence field, so consumers must upgrade before writer rollout. No release is authorized by this scoped review.
