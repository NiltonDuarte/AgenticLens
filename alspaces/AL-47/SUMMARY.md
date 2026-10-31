# AL-47 Work Summary

## Session baseline and approval

- Baseline: `1d20117711aab96a230679d3cdb9406853fe3f04`; initial Git status clean.
- Branch: `feat/AL-47-live-trace-dynamic-refresh-active-session-polling`.
- User approved CONTRACT.md and iteration 01 PLAN.md on 2026-10-31. Implementation stayed within those boundaries.
- Parent orchestrated planner, coder, tester, and reviewer; parent owns integration and delivery.

## Domain

Complete meaningful trace comparison detects in-place content, reasoning, tool, token, child-link, and anomaly changes. Frontend pure helpers identify active statuses and unchanged snapshots.

## Adapters

A stable read-only SQLite observer supplies connection-scoped source generations. Shared bounded trace snapshots eliminate repeated full-history reconstruction on unchanged tree, REST, and SSE reads. Source changes or uncertain observations invalidate reuse; failed reads do not become empty successful snapshots. Existing SSE protocol, nonblocking delivery, configured backend cadence, and source-independent tree status calculation remain intact.

## Presentation

Selected running/active traces recover through a coalesced 2000 ms REST loop alongside SSE. Selection/lifecycle ownership ignores obsolete successes and errors; newer SSE snapshots supersede earlier REST requests. Completion, stall, missing selection, pause, close, and unmount stop automatic refresh. Manual refresh and resume reconcile explicitly. Trace failures retain visible data and survive unrelated successful tree traffic. Stable message identity preserves reasoning choices, and reading anchors preserve position for same-message growth and new turns in both existing orderings.

## Verification

Tester evidence: [iterations/01/TESTING.md](iterations/01/TESTING.md). Full Go race suite passed. Full frontend suite passed 720 tests before two final stopped-polling cases; final focused view/inspector rerun passed 49 tests. Typecheck, lint, focused API tests, and diff checks passed. No tests removed or silently skipped. Existing React act warnings remain in unrelated MetricProgression coverage.

Independent review approved the diff. Implementation commit `f531208` was merged with `origin/main` without conflicts at tested head `6968687122d6146010e4255ae46f3ea0f1b6936b`. Upstream changed planning documents only; its AL-47 clarification matches the accepted scope. Full integrated `make` passed: Go vet/race tests, frontend typecheck/lint, all 34 frontend files and 722 tests, frontend build, and linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 builds. Embedded `TestGetFS` passed within that gate. `diff -qr web/dist backend/internal/adapters/web/dist` passed. The standalone linux/amd64 binary served byte-identical index and JS/CSS over HTTP from an isolated workspace with Node absent from PATH, then shut down cleanly. Built generate help and version JSON were checked. No generated assets were force-added.

The specification moved to plan/done and AL-47 was removed from the ticket index; no downstream ticket declares AL-47 as a blocker. Final Codex telemetry was extracted and verified; PR delivery follows. DOM checks verify controlled reading-anchor arithmetic and state, not browser geometry; deterministic mutation exactly between cache generation observations is untested. Raw source transcripts stay outside Git.


## Telemetry provenance

New ticket ledger: no prior AL-47 ledger or other-harness sessions to preserve. Actual Codex session metadata under the normal ~/.codex/sessions store establishes root `01a191f5-413e-7250-a3bf-0b3f4e5be953` and four direct children: planner `01a191f5-8f30-72b1-8e01-07a9f3fc3825`, coder `01a191fa-d821-7490-9edc-e6a10fdeeb77`, tester `01a19200-2d16-7db3-b0ec-8f788079261f`, reviewer `01a19203-e0c8-7d41-a003-809db2aa2c8e`. Final extraction must retain all five with Codex provenance. Source transcripts are not committed.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 18m 39s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 91.5k tokens (46% of 200k healthy ceiling, 0 compactions) | Subagents peak 110.7k (Tester)
- **Role Effort Distribution (Subagents):** N/A
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 1.5:1 (+440 test lines / +297 prod lines)
- **Code Delivered:** +737 / -94 lines across 12 code files (excluding docs)
- **Reasoning Effort:** 8,008 reasoning tokens (15.4% of generated output)
- **Tool Health & Self-Correction:** 165 tool calls (17 self-corrections, 89.7% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+66.2k tokens) — Context tokens surged by 66.2k to 66.2k tokens.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 79,194 | 6,348 | 786,048 | 90.8% | 83.6k | HEALTHY | $0.00 |
| **coder** | 1 | 74,494 | 13,374 | 2,423,680 | 97.0% | 106.9k | ELEVATED | $0.00 |
| **tester** | 1 | 61,577 | 9,868 | 2,120,064 | 97.2% | 110.7k | ELEVATED | $0.00 |
| **reviewer** | 1 | 61,100 | 3,419 | 1,020,800 | 94.4% | 105.4k | ELEVATED | $0.00 |
| **rootagent** | 1 | 91,934 | 10,853 | 5,158,144 | 98.2% | 91.5k | HEALTHY | $0.00 |
| **Total** | **5** | **368,299** | **43,862** | **11,508,736** | **96.9%** | **110.7k** | **ELEVATED** | **$0.00** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 368,299 | 43,862 | 11,508,736 | 96.9% | $0.00 |
| **Total** | **5** | **368,299** | **43,862** | **11,508,736** | **96.9%** | **$0.00** |


## Final telemetry validation

The built CLI reported five Codex rollouts. The generated ledger was checked against all five expected session IDs, Codex source labels, and four child-to-root links; all matched. There was no pre-existing AL-47 ledger, so no other-harness facts were replaced. Generation used a literal nonce, explicit root ID, source codex, type feat, and ticket title. Generated metrics record 18m 39s active duration, 96.9% cache hit rate, and elevated context health with zero root compactions. The renderer prints $0.00, but the extracted sessions carry no source cost fields; this is unavailable billing evidence, not proof that the run was free. No pricing estimate is asserted. Extraction succeeded without a source error. Raw transcripts remain outside Git.
