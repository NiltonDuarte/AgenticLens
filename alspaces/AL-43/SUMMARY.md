# AL-43 Work Summary

## Iteration 02: user-requested feature revert

Source: the user instructed "revert this feature" after reviewing PR #78's stack-specific output classification. This explicitly approves removal of the feature; iteration 01 below remains historical evidence, not current acceptance.

Initial Git status was clean at 9a3796ebadad7f5eb002b764b43c28e0526efb86 on GEN-13/content-classification-pr78. Work continues under the existing AL-43 ticket on chore/AL-43-revert-content-classification. Configured workspace root remains alspaces. PR #78 is merged, so this iteration requires a new PR.

Remove automatic content classification and its domain, aggregation, extraction, API, and dashboard surfaces. Restore execution-error behavior from before PR #78 and preserve unrelated merged work. Accept retired fields only when decoding existing ledgers so previously written sessions remain readable. Do not migrate or rewrite historical ledgers as part of the revert.

Implementation, verification, and final telemetry evidence are recorded below when completed. Historical iteration 01 records follow.

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
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 44m 2s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 112.0k tokens (56% of 200k healthy ceiling, 0 compactions) | Subagents peak 132.7k (Coder)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 3.3 avg turns/run (peak: 4 turns) | Planner 3.0, Coder 4.0, Tester 3.0, Reviewer 3.0
- **Test-to-Production Code Ratio:** 0.9:1 (+396 test lines / +450 prod lines)
- **Code Delivered:** +846 / -216 lines across 36 code files (excluding docs)
- **Reasoning Effort:** 7,376 reasoning tokens (10.7% of generated output)
- **Tool Health & Self-Correction:** 246 tool calls (12 self-corrections, 95.1% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 105,329 | 4,449 | 691,328 | 86.8% | 66.0k | HEALTHY | $0.00 |
| **coder** | 1 | 98,808 | 19,964 | 5,947,136 | 98.4% | 132.7k | ELEVATED | $0.00 |
| **tester** | 1 | 83,024 | 13,909 | 3,691,776 | 97.8% | 113.4k | ELEVATED | $0.00 |
| **reviewer** | 1 | 106,116 | 5,746 | 2,517,760 | 96.0% | 131.2k | ELEVATED | $0.00 |
| **rootagent** | 1 | 102,232 | 17,329 | 9,797,120 | 99.0% | 112.0k | ELEVATED | $0.00 |
| **Total** | **5** | **495,509** | **61,397** | **22,645,120** | **97.9%** | **132.7k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 495,509 | 61,397 | 22,645,120 | 97.9% | $0.00 |
| **Total** | **5** | **495,509** | **61,397** | **22,645,120** | **97.9%** | **$0.00** |

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
