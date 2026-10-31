# AL-43 Work Summary

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
- **Software Development Life Cycle Iterations:** 10
- **Active Agent Duration:** 55m 30s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 112.0k tokens (56% of 200k healthy ceiling, 0 compactions) | Subagents peak 146.1k (Reviewer)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 2.5, Tester 2.0, Reviewer 2.5
- **Test-to-Production Code Ratio:** 0.4:1 (+92 test lines / +226 prod lines)
- **Code Delivered:** +318 / -710 lines across 36 code files (excluding docs)
- **Reasoning Effort:** 12,215 reasoning tokens (11.7% of generated output)
- **Tool Health & Self-Correction:** 364 tool calls (16 self-corrections, 95.6% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 CRITICAL).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer (5 cycles) — Subagent roles 'tester' and 'reviewer' alternated for 5 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 188,371 | 8,273 | 1,302,912 | 87.4% | 89.2k | HEALTHY | $0.00 |
| **coder** | 2 | 186,327 | 26,939 | 7,701,248 | 97.6% | 132.7k | ELEVATED | $0.00 |
| **tester** | 2 | 166,207 | 17,976 | 5,859,200 | 97.2% | 119.5k | ELEVATED | $0.00 |
| **reviewer** | 2 | 215,606 | 9,762 | 4,748,416 | 95.7% | 146.1k | ELEVATED | $0.00 |
| **rootagent** | 2 | 213,906 | 29,215 | 15,780,736 | 98.7% | 112.0k | ELEVATED | $0.00 |
| **Total** | **10** | **970,417** | **92,165** | **35,392,512** | **97.3%** | **146.1k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 10 | 970,417 | 92,165 | 35,392,512 | 97.3% | $0.00 |
| **Total** | **10** | **970,417** | **92,165** | **35,392,512** | **97.3%** | **$0.00** |

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
