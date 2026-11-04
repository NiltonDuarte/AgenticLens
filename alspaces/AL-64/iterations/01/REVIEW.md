# AL-64 iteration 01 consolidated review

Status: INCOMPLETE; six confirmed issues corrected and all correction review cells complete. Integrated make and standalone assets pass; user source-state policy clarification remains pending.

Candidate: AL64-candidate-01 on a84adbe5be46c2ef52ede0987a417562715d38e1; app/test manifest SHA-256 6aab550b873d4b5de1242b889c26bb3b7c6838d4900ecae6d418b426c2a35695. Independent discovery is complete. Coder corrections and tester regressions are now authorized.

## Review roster

| Responsibility | Model / effort | Phase | Outcome | Report |
| --- | --- | --- | --- | --- |
| reviewer-behavior | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_BEHAVIOR_SOL.md](REVIEW_BEHAVIOR_SOL.md) |
| reviewer-behavior | gpt-6-luna / medium | discovery | APPROVED | [REVIEW_BEHAVIOR_LUNA.md](REVIEW_BEHAVIOR_LUNA.md) |
| reviewer-data | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_DATA_SOL.md](REVIEW_DATA_SOL.md) |
| reviewer-data | gpt-6-luna / medium | discovery | INCOMPLETE | [REVIEW_DATA_LUNA.md](REVIEW_DATA_LUNA.md) |
| reviewer-architecture | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_SOL.md](REVIEW_ARCHITECTURE_SOL.md) |
| reviewer-architecture | gpt-6-luna / medium | discovery | CHANGES_REQUESTED | [REVIEW_ARCHITECTURE_LUNA.md](REVIEW_ARCHITECTURE_LUNA.md) |
| reviewer-tests | gpt-6.1-sol / high | discovery | CHANGES_REQUESTED | [REVIEW_TESTS_SOL.md](REVIEW_TESTS_SOL.md) |
| reviewer-tests | gpt-6-luna / medium | discovery | APPROVED | [REVIEW_TESTS_LUNA.md](REVIEW_TESTS_LUNA.md) |

## Confirmed corrections

- DUR-01, severity 5: invalid Codex completion removes its open start without marking unsupported coverage. Shared discovery: behavior, data, architecture and test-effectiveness reviewers, each gpt-6.1-sol/high. Parent confirmed the missing guard/limitation by source trace.
- DUR-02, severity 6: a v1 part update later than cutoff hides its proven open activity interval. Discovered by architecture reviewer, gpt-6.1-sol/high, with an isolated SQLite counterexample.
- DUR-03, severity 6: a v1 message created after cutoff suppresses an interval that was active at cutoff. Shared discovery: behavior reviewer gpt-6.1-sol/high (isolated probe) and architecture reviewer gpt-6-luna/medium (source trace).
- DUR-04, severity 4: fallback OpenCode recovery uses its own clock instead of the operation cutoff. Discovered by data reviewer gpt-6.1-sol/high, reproduced with an isolated fallback probe and confirmed by parent source trace.
- DUR-05, severity 5: no deterministic no-tool streaming regression protects either OpenCode native streaming path. Discovered by test-effectiveness reviewer gpt-6.1-sol/high; both path-removal mutants pass existing package suites, and isolated new oracles kill both.

All findings are open; no correction or final approval is claimed. Working run/finding records remain in /tmp/al64-review-01. Preserve all discovery credit when consolidating later results.

## Verification and remaining conditions

Independent tester passed full Go race, 834 frontend tests, typecheck and lint on candidate 01. Native OpenCode generation, active --recompute-db, completion and unchanged repeat passed both pinned versions. These pre-correction results do not refute demonstrated boundary failures.

The user decision on provisional explicitly open source intervals versus live-process proof remains pending. No reviewer resolves that policy. Existing scope approval remains valid for other work; corrections are proceeding after frozen first-pass discovery completed. Integrated make, affected re-review, final embedded binary verification, telemetry and publication remain pending.

## Candidate 02 corrections

Coder gpt-6.1-sol/high corrected DUR-01 through DUR-04. Tester gpt-6.1-sol/high added DUR-01 through DUR-05 oracles, verified original-code overlays fail and corrected controls pass, and verified path-removal mutants fail the new streaming guards. Affected Codex/OpenCode/backfill race suites pass. This is verification credit, not additional discovery.

Corrected manifest: 027324fa31f45b9933adb35bd34835b54918757047556354e8b2c3ce82f4c24c (46 app/test paths). Product/test edits frozen. Focused correction reviews under both profiles are pending. Native Codex source-copy generation was blocked by project-ownership validation; native lifecycle records are observed, native Codex fresh generation/refresh is unverified.

Candidate 02 correction review outcomes: behavior and data reviewers gpt-6.1-sol/high verified DUR-01 through DUR-05 with no new findings. Architecture reviewer gpt-6.1-sol/high verified those corrections and independently discovered DUR-06 (severity 5) during invalid-timestamp re-review: missing/malformed Codex starts can still label a known subtotal complete. This distinct branch defect is verification-phase discovery, not shared credit for DUR-01. Coder/tester corrections are in progress; no final approval.

## Candidate 03

Coder gpt-6.1-sol/high corrected DUR-06 with an explicit invalid-start limitation; tester gpt-6.1-sol/high proved missing/malformed start oracles fail pre-fix and corrected full Codex race passes. Manifest f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49 (46 paths). Architecture/test-effectiveness Sol re-review is running; all four Luna correction reviews remain queued.

Candidate02 behavior/data Sol approvals carry forward for their verified correction scopes: only the Codex invalid-start parse-error branch and its table cases changed; existing partial-evidence representation is reused, and v1/fallback/DTO/aggregation/UI/retention paths are identical. Architecture/test-effectiveness reviewers inspect the new branch on candidate03. The unresolved user policy remains a separate material delivery condition.

Candidate03 architecture and test-effectiveness reviewers gpt-6.1-sol/high issued scoped APPROVED after exact hash checks and targeted controls; no new findings. All six findings are fixed and regression-verified. Current delivery remains incomplete pending Luna correction coverage, user policy decision and required integrated checks. Detailed findings/reports remain active.

Candidate03 Luna behavior, data and architecture correction reviews issued scoped APPROVED after exact46-path checks with no new distinct findings. Data also independently passed affected race tests. The final Luna test-effectiveness correction cell is running. Existing corrected findings are not unresolved blockers; policy/verification gaps keep the consolidated verdict INCOMPLETE.

All applicable candidate03 persona/profile cells are now scoped APPROVED, with Sol behavior/data carried forward as justified above. The Luna test-effectiveness review verified the discriminating regression oracles and exact frozen file hashes. No confirmed finding remains open. Consolidated delivery remains INCOMPLETE pending the source-interval policy and integrated verification. Parent will commit the reviewed candidate, merge origin/main and run make; these reversible verification steps do not publish or approve the unresolved policy.

## Integrated gate checkpoint

Reviewed implementation commit 44db8658097e9fb27c7ed1149ee132906903d862; origin/main integration was already up to date and introduced no code changes. Parent ran required make successfully, including full race/frontend/check/build gates, and verified the host standalone binary serves exact current canonical HTML/JS/CSS. Source/test scopes remain identical to the frozen candidate03 manifest, so review evidence remains applicable. Full evidence is in SUMMARY.md.

No confirmed bug remains open. Final consolidated approval remains INCOMPLETE because user clarification on source-state/liveness semantics is still pending. No review cleanup is performed: all reports and compact findings remain available. Final telemetry, final delivery commit, push/PR and ticket lifecycle synchronization are deferred until this acceptance boundary is resolved.
