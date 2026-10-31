# AL-49 Iteration 02 Review

Baseline: `1d20117711aab96a230679d3cdb9406853fe3f04`. Documentation-only reconciliation; no application, test, schema, or runtime-script changes. Initial Git status was clean. This is a direct documentation audit, not a new application SDLC cycle.

## Findings and disposition

| Ticket | Disposition and evidence |
| --- | --- |
| AL-49 | Removed stale Ready entry. Merged PR #59, commit `ba53cfc`; specification already in done/. Current inspector has latest-first ordering and navigation controls. |
| AL-50 | Removed stale Ready entry. Merged PR #60, commit `791f948`; specification already in done/. |
| AL-43 | Retained Ready. Semantic content-error facts/UI are not implemented; `ClassifyToolExecution` deliberately handles process failures only. AL-42 does not complete this separate slice. |
| AL-47 | Retained Ready; annotated partial coverage. `live/service.go` emits trace updates and `LiveObservability.tsx` consumes them. No selected-active-session polling fallback is established by those paths. AL-49 supplies scroll handling; full acceptance remains unverified. |
| AL-52 | Synchronized specification to Ready with no blockers. AL-51 PR #62 shipped metadata, explicitly excluding skill/task UI. Badge/navigation acceptance remains open. |
| AL-53 | Synchronized specification to Ready with no blockers. Dedicated panel remains open. Replaced obsolete ordinary-v3-ingestion wording with supported-v4 optional-data handling, preserving the existing AL-56 runtime boundary. |
| AL-54 | Synchronized specification to Ready with no blockers. Skill adoption aggregation, fleet cards, and comparison filtering remain open; AL-51 metadata alone does not deliver them. |
| AL-55 | Retained Blocked by AL-52/53/54; removed completed AL-51 blocker. Added replan notice: AL-56 PRs #38/#43 shipped v4 and source-backed migration boundaries; current `StepV3ToV4UserIterations` is not the proposed required-array normalization. Original scope retained as historical proposal, not an approved replacement design. |
| AL-57 | Moved to Ready. Existing specification and AL-56 summary explicitly say independent of AL-55. Current `CostUSD float64` and TS `cost_usd: number` still cannot express unavailable cost. |
| AL2-1 | Retained Ready. Existing frontend diagnostic export is not the specified CLI prompt command, HTTP debug-prompt endpoint, and interactive drawer slice. No merged delivery evidence found. |
| AL2-2 | Retained Ready. Archive CLI, archive-backed scanner, and archived UI contract not delivered. |
| AL2-3 | Retained Ready. Existing source/configuration selection is not the specified multi-project schema, per-project scanner, and cross-project rollups. |
| AL2-4 | Retained Blocked by AL2-3. Remote cache/project API not delivered. |
| AL2-5 | Retained Blocked by AL2-3/4. Multi-project selector and persistent project selection not delivered. |
| AL2-6 | Retained Ready. Rate-card configuration, estimated-cost provenance, and calculation contract not delivered. |
| AL2-7 | Retained Blocked by AL2-6. Existing scatter plot does not deliver cost-mode switching and rate-card simulation. |
| AL2-8 | Retained Ready. Invocation facts and frontend model breakdowns do not expose the specified reasoning-effort variant contract. |

Merged states checked using GitHub CLI; current source and ticket contracts checked against baseline history. Links: [AL-49 #59](https://github.com/NiltonDuarte/AgenticLens/pull/59), [AL-50 #60](https://github.com/NiltonDuarte/AgenticLens/pull/60), [AL-51 #62](https://github.com/NiltonDuarte/AgenticLens/pull/62), [AL-56 #43](https://github.com/NiltonDuarte/AgenticLens/pull/43).

## Verification

- Passed read-only consistency check: 15 to-do specifications, 17 unique indexed tickets including 2 discarded; every referenced file exists; every to-do file is indexed exactly once; Ready/Blocked sections agree with statuses and outstanding dependencies; explicit spec headers agree with index; blocker relationships are reciprocal.
- `git diff --check`: passed.
- Full `make`: pending final result and upstream synchronization.
- No new automated tests: this change edits only planning documents. No claim that remaining product acceptance cases were executed.

## Applicable regression-prevention cases

Applied configured workspace root, preserved baseline and historical implementation evidence, distinguished partially implemented features from delivered acceptance, and kept migration shape separate from source recovery. AL-55 requires replanning; this audit neither marks it complete nor approves changing runtime ingestion.
