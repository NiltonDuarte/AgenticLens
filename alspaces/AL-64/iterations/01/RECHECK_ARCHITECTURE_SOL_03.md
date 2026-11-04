# AL-64 candidate 03 architecture correction review

Verdict: APPROVED for the bounded correction scope. DUR-06 is fixed; no new blocking finding was demonstrated. This is not consolidated delivery approval. The pending user decision about source-state provisional intervals versus live-process proof remains an independent delivery gate.

## Identity, revision and permissions

- Responsibility: architecture and maintainability, correction re-review. Run: recheck_arch_sol_03; local candidate AL64-candidate03.
- Model/effort: gpt-6.1-sol / high, inherited and acknowledged from parent dispatch. Independent provider/runtime metadata, usage and cost are unavailable; identity is not inferred from the persona name.
- Baseline: a84adbe5be46c2ef52ede0987a417562715d38e1; branch fix/AL-64-active-turn-duration.
- Frozen manifest: /tmp/al64-review-03/fingerprint.json; SHA256 f9b8bb31e5d7a3776adf9fef2b2f3f456711e76bfbceb7f276880e7e4634eb49. Independently verified manifest digest and all 46 file hashes at entry and immediately before this verdict: 46 matched, zero mismatches.
- Compared candidate02 and candidate03 manifests and frozen diffs. Exactly rollout.go and rollout_test.go differ. The production difference is the four-line invalid-start limitation branch; the test difference adds missing/malformed starts and strengthens partial-boundary assertions. All other producer, domain, DTO, retention, aggregate and UI paths are identical.
- Read the assigned persona/shared context, canonical review skill, technical/regression standards, contract, plan and supplied investigation/test/runtime evidence. Read my prior candidate02 architecture report to reuse its correction verification; no sibling persona or detailed sibling review report was loaded.
- Initial Git status was recorded and all shared changes preserved. Only this assigned report was written in the checkout. Product/tests remained read-only; isolated probe source/overlay is under /tmp/al64-arch-recheck-sol-03. No children, commits, telemetry or publication.

## Findings and correction dispositions

No new distinct finding. DUR-06 verification retains the original discovery credit: architecture reviewer gpt-6.1-sol/high during candidate02 correction review. This review confirms that supplied finding rather than claiming another discovery.

| Finding | Severity | Disposition and evidence |
| --- | --- | --- |
| DUR-06: invalid Codex start falsely certifies complete coverage | 5/10 | Fixed. At rollout.go:427-430 a missing or malformed timestamp now sets durationLimit without inserting a turnStarts entry. A later same-ID completion cannot create a measured interval. finalizeFacts passes the limitation to MeasuredDuration; the earlier supported four seconds remain completed seconds, state becomes partial, provisional seconds remain zero, cutoff remains empty, completed_at remains the earlier proven 00:00:05 and DurationForInvocation counts unknown coverage. |
| DUR-01: invalid completion drops coverage limitation | 5/10 | Prior fixed disposition carried forward. Completion error branch, domain evidence and normalization are unchanged; the retained lifecycle table independently passed reversed/malformed completion again in this run. |
| DUR-02: v1 generic update hides native open activity | 6/10 | Prior fixed disposition carried forward. Native-start cutoff proof and its tests are byte-identical to candidate02; its independent absent/zero/future native-start controls and tester red/green evidence remain applicable. |
| DUR-03: future message suppresses prior activity | 6/10 | Prior fixed disposition carried forward. At/before-cutoff supersession SQL and dependencies are byte-identical; prior before/after cutoff controls remain applicable. |
| DUR-04: fallback recovery chooses another cutoff | 4/10 | Prior fixed disposition carried forward. recoverOpenCodeAt/readOpenCodeTree/strict extraction propagation, engine cutoff selection and exact-owner replacement are byte-identical. Prior fixed-cutoff regression and red/green evidence remain applicable. |
| DUR-05: missing no-tool streaming guards | 5/10 | Prior protection disposition carried forward. Both family-specific streaming tests and production paths are byte-identical; retained v1 text/reasoning and v2 idle-null guards plus path-removal failures remain applicable. |

The original DUR-06 trigger and consequence were an AL-64-introduced contract defect: malformed/missing native start followed by matching completion silently labelled a known subtotal complete. The expected result is supported subtotal plus an explicit limitation, never an invented interval. The correction is local to native parsing and reuses the existing typed evidence channel; callers need no parallel lifecycle reconstruction or new validator. It introduces no clock, I/O or dependency into the functional core, no persistent journal, and no change to owner refresh/portability authority.

## Hypotheses assessed and checks

The new limitation branch is inside the existing unseen-ID guard, so valid duplicates retain their established identity semantics. It executes only after parsing fails; valid future timestamps still return through the earlier lifecycle cutoff gate before mutating seen/start state. Finalization sums only successfully parsed starts; the new limitation cannot invent provisional milliseconds. A later completion has no start to pair with and preserves the prior proven completion. SessionFacts normalization accepts the corrected partial evidence rather than discarding supported duration.

Executed from backend:

    go test -overlay /tmp/al64-arch-recheck-sol-03/overlay.json ./internal/adapters/codex -run 'TestRolloutActiveDurationCutoffAndTerminalBoundaries|TestRecheck03StartControls' -count=1 -v

Exit 0. All nine retained lifecycle cases passed. Six disposable controls also passed: duplicate open start (14 total/10 provisional), duplicate after completion (9 complete), future start preserving an existing open interval (14/10), future start creating no interval (4 complete), malformed start without completion (4 partial) and missing start without completion (4 partial). Each disposable case additionally traversed SessionFacts normalization. Partial controls require an explicit limitation, empty cutoff and the earlier proven completion. The two retained DUR-06 cases include a valid later matching completion and assert partial/known subtotal/unknown summary coverage.

Reused evidence:

- Tester DUR-06 pre-fix overlay removes only the four-line branch. Inspected /tmp/al64-tester-dur06/pre-fix.log: both new cases fail their coverage oracle. Corrected full Codex race log reports exit 0 (1.148s). This is supplied-finding verification, with tester verification credit distinct from discovery.
- Candidate02 architecture source tracing and isolated boundary probes for DUR-01 through DUR-04, and inspection/mutant evidence for DUR-05. The manifest/diff comparison establishes why those unchanged paths/dependencies remain covered.
- TEST_REPORT.md records passed affected Codex/OpenCode/backfill race suites, failed pre-correction overlays for DUR-01 through DUR-04 and failed streaming path-removal mutants for DUR-05. Candidate01 full Go race, 834 frontend tests, typecheck and lint remain reusable for unchanged scopes. No redundant broad suite was run.
- Recorded native OpenCode v1/v2 live generation, active source refresh, completion and unchanged-repeat evidence is reused, not independently rerun.

## Remaining limits

- The source-state provisional implementation is assessed conditionally. The explicit user liveness policy decision remains pending; no process-liveness guarantee or final delivery approval is implied.
- Native Codex lifecycle shapes were observed, but its copied-source native generation was blocked by correct project ownership validation. Native Codex generation, active refresh, completion refresh and unchanged-repeat remain unverified; deterministic parser/engine evidence does not replace that runtime result.
- Parent retains integrated make, final embedded standalone verification, consolidated verdict and publication. The accepted reader-first optional v6 compatibility boundary remains: old strict readers reject emitted evidence, so consumers must upgrade before writer rollout. No release is authorized here.
