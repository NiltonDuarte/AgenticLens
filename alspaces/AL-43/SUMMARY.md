# AL-43 Work Summary

## Iteration 03: PR #88 execution-classifier correction

Current acceptance is the user's exact instruction routed through MGT-1: "retain main’s execution classifier while removing content classification, and add these regression cases." This explicitly supersedes iteration 02's mechanical pre-PR #78 execution restoration. No repeat permission is required. F1 is a review correction to preserve execution semantics, not a replacement feature or expansion of execution policy.

Before writing, verified cwd and Git toplevel /home/niltonduarte/.t3/worktrees/AgenticLens/t3-3f89bda8, branch chore/AL-43-revert-content-classification, clean status and HEAD e91cf0d05348537657c210601c6acc93c01f967d. Fetched origin/main at 3857fa22f49f825c82ced39c3f81f963b7447545. This is PR #88's owner checkout; no manager or original PR #78 checkout is used.

Preserve current-main ClassifyToolExecution exactly. Protect successful read output containing timeoutMs, successful grep output quoting panic, and explicitly zero-exit output quoting permission denied. Keep content-classification removal, strict decode-only retired-field compatibility, raw Codex evidence, AL-55 recovery and nil-array defaults, AL-47 subscription/refresh, and AL-53 navigation. Historical iteration 02 evidence below is superseded only on execution-restoration acceptance. Independent tester and reviewer approved the correction; F1 is closed. The classifier exactly matches main 3857fa22f49f825c82ced39c3f81f963b7447545, with no content-feature reintroduction. The three supplied successful-output cases are protected in the existing classifier table and real SQLite live/offline parity fixture. The fixture measures eight calls and three genuine failures, retains raw output and no failure metadata for successes, confirms live streak reset/non-escalation, and rejects an offline retry-streak anomaly. Explicit failures, nonzero exits, actual crash, exact timeout/permission and no-match guards remain.

Correction commit 3828f537cd6c448b8b811688e2b0cd693b0a6cd5 was fetched/merged against actual origin/main 3857fa22f49f825c82ced39c3f81f963b7447545 (already up to date). Full make then passed: dependency stages, Go vet/race suites, TypeScript typecheck/lint, 741 frontend tests in 35 files, canonical frontend embedding, and all four Linux/macOS amd64/arm64 builds. Independent focused race verification additionally covered compatibility, Codex evidence and recovery/nil-array/live boundaries. Full integrated application diff against that main reference has SHA-256 5d45600be782aee45c3b243c8e49f35456f88f1b4134cdcf82addb00c403c13c, defined by git diff base...HEAD -- backend web/src. This remains stable through later artifact-only commits.

Generated source/embedded asset trees match and remain ignored; host binary version/help passed. Frontend sources and generated asset names remain unchanged from iteration 02, so its standalone HTTP asset and UI interaction evidence is reused. No new browser/geometry/HTTP measurement of F1 is claimed. Existing non-failing npm audit findings and React act diagnostics remain outside this correction. Final telemetry refresh uses this owner root 01a1928e-9694-7d53-96a0-99bcd82dbdc7, not the manager root; all ten previous session identities, parent links and existing source labels were verified preserved after extraction. The selected owner root and four expected children match source metadata; no manager root or fabricated session was added, no source errors occurred, and no previous other-harness sessions existed. Raw transcripts remain outside Git. Telemetry is an active-session snapshot that excludes subsequent delivery calls; source-projected zero cost is not confirmed billing data. The user's later short brief reaffirmed the same approved scope, retained current progress, and reserved the merge gate for the user.

## Iteration 02: user-requested feature revert

Source: the user instructed "revert this feature" after reviewing PR #78's stack-specific output classification. This explicitly approves removal of the feature; iteration 01 below remains historical evidence, not current acceptance.

Initial Git status was clean at 9a3796ebadad7f5eb002b764b43c28e0526efb86 on GEN-13/content-classification-pr78. Work continues under the existing AL-43 ticket on chore/AL-43-revert-content-classification. Configured workspace root remains alspaces. PR #78 is merged, so this iteration requires a new PR.

Remove automatic content classification and its domain, aggregation, extraction, API, and dashboard surfaces. Restore execution-error behavior from before PR #78 and preserve unrelated merged work. Accept retired fields only when decoding existing ledgers so previously written sessions remain readable. Do not migrate or rewrite historical ledgers as part of the revert.

Implementation complete: automatic content classification, content-only ledger/domain/API fields and counts, extraction wiring, and dashboard warning badges/counters/filters were removed. Raw Codex output is retained. Retired nested fields are accepted only on ledger reads and omitted from output. Execution classification is restored to the pre-PR #78 behavior, including its original crash/permission/timeout heuristics. The strict compatibility regression and existing source/UI tests cover retained behavior. Test retirement rationale is in iterations/02/IMPLEMENTATION.md.

Independent tester and reviewer approved the application diff. Implementation commit ac376a7 was followed by clean integration of origin/main 3857fa2 (AL-55 backfill) in 4af99a8. Integration re-review approved the overlapping extraction/schema paths. Full make passed both before and after integration; the final gate at 4af99a8 ran Go vet/race suites, TypeScript typecheck/lint, all 741 frontend tests (35 files), canonical frontend embedding and all four Linux/macOS amd64/arm64 builds. Application diff against integrated origin/main has SHA-256 61eaecb49fce16593612ffab3b0308dba2d10f25a0c5c131950f60b5261be745.

Generated web/dist and backend/internal/adapters/web/dist are identical and ignored. The integrated standalone host binary served byte-identical index, JS and CSS over HTTP; the temporary server was terminated. Version/help and diff checks passed. Browser geometry was not tested. Existing npm audit findings and non-failing React act warnings are unchanged. Historical AL-43 artifacts and completed plan lifecycle were preserved; no new specification or index entry was created.

Final explicit Codex telemetry merged this root and its four expected children with the five previously recorded Codex sessions (10 sessions total). Existing ledger has no other-harness sessions. Source metadata establishes current root 01a1928e-9694-7d53-96a0-99bcd82dbdc7 and planner 01a19293-8606-7401-abff-d598a17fee74, coder 01a19295-8da6-7e02-ac9e-9fd3bbeaaba3, tester 01a19298-81ef-7c82-8ece-ee6582a77bfd, reviewer 01a19299-6442-75d1-aea4-dd71f3506153. After generation, all five prior session IDs were verified preserved; the selected root, all four child IDs, their parent links and Codex source labels matched the real source metadata. No source errors occurred. Raw transcripts remain outside Git; current root and later handoff activity are snapshot limits. Zero source-projected costs are not confirmed billing data.

Historical iteration 01 records follow.

## Historical iteration 01

User pre-approved the plan and contract and authorized full delivery. Initial Git status was clean at 85a2af94b5036b0918aaf9a7a89aed3a990f2650 on AL-43/surface-semantic-content-errors. Configured workspace: alspaces.

## Domain

Added pure semantic content classification and independent tool content-error counts, event diagnostics, canonical session mappings, and aggregate totals. Completed calls retain execution success when output contains semantic diagnostics; explicit execution failures remain separate.

## Adapters

Propagated diagnostics through supported OpenCode schema families, Codex result handlers, live parsing, serialization and HTTP projections. Compact persistence retains content-warning evidence, including when optional source hydration is unavailable. Legacy missing fields default to zero/false.

## Presentation

Added amber CONTENT ERROR and red EXEC ERROR indicators, independent live counters, content-aware trace filtering and diagnostic copy, and Deep Dive totals and table values.

## Verification

Independent testing and review approved the implementation. Integrated `make` passed at implementation commit 873343d after fetching and merging origin/main (already up to date): Go vet, all Go race suites, frontend typecheck/lint, all 714 frontend tests, frontend build and Linux/macOS amd64/arm64 builds. `diff -qr web/dist backend/internal/adapters/web/dist` passed. Standalone binary served byte-identical embedded index, JS and CSS assets; temporary verification server was stopped. Generated assets remain ignored. Browser geometry was not tested. Existing npm dependency audit reports 21 vulnerabilities; unrelated dependency updates were outside this ticket. Frontend tests emitted non-failing React act warnings in MetricProgression.

Specification moved to plan/done and completed ticket removed from index; no downstream dependencies reference AL-43. Final telemetry evidence follows.

## Telemetry provenance and limits

Final explicit Codex extraction found root 01a19201-e9f1-7001-922f-3ff5d6344a64 and all four expected planner/coder/tester/reviewer children. IDs were verified against real rollout metadata; ledger parent links point to that root and all sources are codex. No earlier ledger or other-harness sessions existed to preserve. Raw transcripts remain outside Git. The current root is still active, so its completion/duration and post-extraction handoff calls are snapshot limits. Reported cost is zero in the source projection; actual billing/pricing is unavailable and zero must not be interpreted as confirmed free usage. The rendered rework anomaly is recorded telemetry, not an unresolved test failure.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 12
- **Active Agent Duration:** 1h 13m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 147.6k tokens (74% of 200k healthy ceiling, 0 compactions) | Subagents peak 163.8k (Reviewer)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 2.8 avg turns/run (peak: 4 turns) | Planner 2.5, Coder 3.0, Tester 2.5, Reviewer 3.0
- **Test-to-Production Code Ratio:** 0.6:1 (+142 test lines / +220 prod lines)
- **Code Delivered:** +362 / -707 lines across 36 code files (excluding docs)
- **Reasoning Effort:** 17,469 reasoning tokens (13.1% of generated output)
- **Tool Health & Self-Correction:** 429 tool calls (16 self-corrections, 96.3% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (1 CRITICAL, 1 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer (5 cycles) — Subagent roles 'tester' and 'reviewer' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.5k tokens) — Context tokens reached 162.5k, breaching the 160k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 283,773 | 10,586 | 1,695,488 | 85.7% | 102.2k | ELEVATED | $0.00 |
| **coder** | 2 | 297,713 | 32,774 | 9,584,000 | 97.0% | 148.9k | ELEVATED | $0.00 |
| **tester** | 2 | 191,320 | 21,583 | 7,603,968 | 97.5% | 142.8k | ELEVATED | $0.00 |
| **reviewer** | 2 | 234,357 | 11,852 | 5,836,288 | 96.1% | 163.8k | ELEVATED | $0.00 |
| **rootagent** | 2 | 245,089 | 39,495 | 20,917,632 | 98.8% | 147.6k | ELEVATED | $0.00 |
| **Total** | **10** | **1,252,252** | **116,290** | **45,637,376** | **97.3%** | **163.8k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 10 | 1,252,252 | 116,290 | 45,637,376 | 97.3% | $0.00 |
| **Total** | **10** | **1,252,252** | **116,290** | **45,637,376** | **97.3%** | **$0.00** |

## Management review corrections F3/F4

User-authorized MGT-1 findings correct the existing accepted scope in iteration 01. Starting head: 8c4e2dc7a25586b13af6a7af0c0859b73edef65b, clean worktree. Upstream integration is explicitly held until the manager confirms #75 and #77 merged; no merge or auto-merge is authorized for this ticket owner.

F3: Matching older recorded events now gain positive hydrated content-warning metadata in both Go and TypeScript. False/absent enrichment cannot retract recorded diagnostics; missing/empty diagnostic details retain existing evidence. Execution state and event identity are preserved.

F4: Codex classification decodes supported text/input_text arrays, MCP content envelopes, string-wrapped JSON and command output wrappers separately from unchanged ToolOutput evidence. Unknown/nontext shapes do not manufacture diagnostics. Decoding is bounded to eight nesting levels.

Source-shape evidence: coder reported a bounded shape-only inspection of 20 recent local Codex rollout files (1,047 response outputs): 672 outputs used input_text arrays, and 14 text payloads decoded to content/isError envelopes; command output JSON wrappers were also observed. One sampled source was rollout-2026-10-31T13-51-26-01a19203-d42b-7020-9684-5937117880ea.jsonl in the local Codex session store. This is observed envelope-shape evidence from the coder session, not proof that real compiler failures occurred in those records or that all Codex versions emit them. Compiler-content tests are self-contained synthetic fixtures through production parsing; raw transcripts remain outside Git.

Correction verification: full `make` passed on the working diff from 8c4e2dc7a25586b13af6a7af0c0859b73edef65b, subsequently committed unchanged as 004523114326650864f02a66dc64fd1bbd593d19. All Go race suites, typecheck/lint, 721 frontend tests and four platform builds passed. Frontend/embed trees match and standalone index/JS/CSS match the current build. Build metadata identifies the pre-commit 8c4e2dc-dirty snapshot because make started before commit; product source tested is the 0045231 correction tree. Independent review approved F3/F4. Existing npm audit and React act warnings remain non-failing. No upstream merge was performed; later cross-ticket inspector/navigation/refresh-anchor integration is untested and held for manager sequencing.

## Actual main integration

The manager explicitly released the hold after verifying #75 and #77 merged. Actual origin/main bdef2d0ef98e37fe259921bc30ba2daefbfda717 was fetched and merged into starting AL-43 head bf575e1afd66ba17a82c5fbc1c4f3b3a2e65fcc8 as 71f1203b284dec59de18b1c8d5637af08ba68f09. Four conflicts in the live inspector, its tests, waterfall tests and CSS were resolved with both behaviors retained. Invocation badges/child navigation coexist with amber content warnings and red execution errors; an existing combined test also protects reasoning/failure expansion and diagnostic copy.

Upstream useLiveTrace ownership/cancellation/reading anchors, shared focusSession/panel navigation and fleet filters/adoption remain intact. The merged #82 Subscribe-before-headers/snapshot implementation and concrete SSE regressions are unchanged from actual main; no timeout or fixture bypass was introduced. F3/F4 product corrections are unchanged from the reviewed starting head. AL-55 remains Ready/None in index and specification; all completed-ticket removals remain preserved. No AL-55 implementation or PR merge/auto-merge was performed.

Focused integration checks passed: 259 frontend tests, typecheck, Go domain/Codex race suites, and concrete SSE protocol/SQLite recovery/subscription cleanup race regressions. Full make and final asset verification follow on the actual merge. PR merge sequencing remains manager-owned.

Actual integration final gate: `make` PASS on merge 71f1203b284dec59de18b1c8d5637af08ba68f09 against base bdef2d0ef98e37fe259921bc30ba2daefbfda717. All Go race suites, vet, typecheck/lint, 35 frontend files / 750 tests and all four binary builds passed. Embedded asset trees are byte-identical; the standalone 71f1203 binary served the current index, JS and CSS exactly, and the temporary server was stopped. Independent combined review APPROVED. Browser geometry and deployment were not tested. Existing dependency audit and non-failing React act warnings remain recorded limitations. Post-gate changes are AL-43 evidence/telemetry only.
