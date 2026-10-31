# AL-43 Work Summary

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
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 30m 48s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 94.6k tokens (47% of 200k healthy ceiling, 0 compactions) | Subagents peak 117.1k (Coder)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 3 turns) | Planner 2.0, Coder 3.0, Tester 2.0, Reviewer 2.0
- **Test-to-Production Code Ratio:** 0.8:1 (+375 test lines / +449 prod lines)
- **Code Delivered:** +824 / -213 lines across 36 code files (excluding docs)
- **Reasoning Effort:** 5,949 reasoning tokens (11.1% of generated output)
- **Tool Health & Self-Correction:** 184 tool calls (8 self-corrections, 95.7% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 101,784 | 3,814 | 500,736 | 83.1% | 62.3k | HEALTHY | $0.00 |
| **coder** | 1 | 83,716 | 17,664 | 4,425,216 | 98.1% | 117.1k | ELEVATED | $0.00 |
| **tester** | 1 | 65,603 | 10,622 | 2,432,128 | 97.4% | 95.0k | HEALTHY | $0.00 |
| **reviewer** | 1 | 78,627 | 3,655 | 1,454,208 | 94.9% | 102.9k | ELEVATED | $0.00 |
| **rootagent** | 1 | 85,398 | 12,006 | 6,253,184 | 98.7% | 94.6k | HEALTHY | $0.00 |
| **Total** | **5** | **415,128** | **47,761** | **15,065,472** | **97.3%** | **117.1k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 415,128 | 47,761 | 15,065,472 | 97.3% | $0.00 |
| **Total** | **5** | **415,128** | **47,761** | **15,065,472** | **97.3%** | **$0.00** |

## Management review corrections F3/F4

User-authorized MGT-1 findings correct the existing accepted scope in iteration 01. Starting head: 8c4e2dc7a25586b13af6a7af0c0859b73edef65b, clean worktree. Upstream integration is explicitly held until the manager confirms #75 and #77 merged; no merge or auto-merge is authorized for this ticket owner.

F3: Matching older recorded events now gain positive hydrated content-warning metadata in both Go and TypeScript. False/absent enrichment cannot retract recorded diagnostics; missing/empty diagnostic details retain existing evidence. Execution state and event identity are preserved.

F4: Codex classification decodes supported text/input_text arrays, MCP content envelopes, string-wrapped JSON and command output wrappers separately from unchanged ToolOutput evidence. Unknown/nontext shapes do not manufacture diagnostics. Decoding is bounded to eight nesting levels.

Source-shape evidence: coder reported a bounded shape-only inspection of 20 recent local Codex rollout files (1,047 response outputs): 672 outputs used input_text arrays, and 14 text payloads decoded to content/isError envelopes; command output JSON wrappers were also observed. One sampled source was rollout-2026-10-31T13-51-26-01a19203-d42b-7020-9684-5937117880ea.jsonl in the local Codex session store. This is observed envelope-shape evidence from the coder session, not proof that real compiler failures occurred in those records or that all Codex versions emit them. Compiler-content tests are self-contained synthetic fixtures through production parsing; raw transcripts remain outside Git.

Correction verification: full `make` passed on the working diff from 8c4e2dc7a25586b13af6a7af0c0859b73edef65b, subsequently committed unchanged as 004523114326650864f02a66dc64fd1bbd593d19. All Go race suites, typecheck/lint, 721 frontend tests and four platform builds passed. Frontend/embed trees match and standalone index/JS/CSS match the current build. Build metadata identifies the pre-commit 8c4e2dc-dirty snapshot because make started before commit; product source tested is the 0045231 correction tree. Independent review approved F3/F4. Existing npm audit and React act warnings remain non-failing. No upstream merge was performed; later cross-ticket inspector/navigation/refresh-anchor integration is untested and held for manager sequencing.
