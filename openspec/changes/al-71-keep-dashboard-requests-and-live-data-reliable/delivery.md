# fix(AL-71): keep dashboard requests and live data reliable

## Outcome

Comparisons and Progression now fence asynchronous results through the current request. Fleet retains paired summary and workspace results. Progression shows a UTC window anchored to the newest usable ledger in the selection, so a filter that leaves the selection intact also leaves its historical chart intact.

REST and SSE reject malformed live data while retaining the last valid observation. Fallback polling keeps one request in flight, reports the actual failure and recovers at the configured cadence. Trace warnings stay with the affected session. An older REST trace cannot erase a newer SSE warning.

## Verification and review

Integrated revision `3e5377a9` passed `make -j2` with one Vitest worker and a local 20-second timeout for frontend integration. Earlier attempts timed out under host load. The successful run passed all 235 frontend unit tests, all 849 frontend integration tests, the Go race suites, static checks, installation and distribution checks, and all native builds. Earlier standalone reruns of the two capture assertions passed as well.

The native workspace handler and SQLite-backed REST/SSE producer verified the shared wire fixtures. Progression boundary checks passed in non-UTC time zones. The Linux manager and standalone engine reported their versions. From an isolated configured project, the engine served the dashboard and JS/CSS byte-for-byte against both generated asset trees and answered the workspace HTTP request.

Behavior, data integrity, architecture and test effectiveness were reviewed independently with Sol 6.2 high and Luna 6 medium. The first passes found diagnostic recovery, timestamp and numeric validity, and regression-coverage gaps. Focused rechecks verified the corrections against the final application and test fingerprint `8e153a7d89492dc2ee27903078330879ec40848f0941ea984696da2e52281487`. [Review findings and attribution](evidence/REVIEW_SUMMARY.json)

## Limits

Automatic capture reconciliation reported a degraded collection. Exact-session recovery and final root extraction verified all 21 expected Codex sessions, including the root and nine source-proven children, and preserved all 16 previously accepted sessions. Five independent sessions lacked verified automatic bindings and were recovered from their exact local source. API-equivalent costs remain unavailable.

## Pull request

### Changes

- Keep dashboard results, warnings and errors tied to the active filter request.
- Use the newest selected ledger for Progression presets and display the effective UTC dates.
- Reject malformed REST and SSE snapshots, preserve usable data, and serialize fallback polling with scoped recovery.

### Verification

The integrated `make -j2` gate passed with 235 frontend unit and 849 frontend integration tests, plus Go race and distribution checks. The integration timeout was adjusted locally for host load. The standalone engine served the current embedded assets and workspace API.

[Pull request #160](https://github.com/NiltonDuarte/AgenticLens/pull/160)

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 41
- **Active Agent Duration:** 2h 41m (recorded duration including 3965s provisional)
- **Context Window Usage:** Root peak 303.0k tokens (67% of 450k reference ceiling, 0 compactions) | Subagents peak 228.1k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.6 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 3.0, Tester 2.5, Reviewer-architecture 3.0, Reviewer-behavior 3.0, Reviewer-data 3.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 2.2:1 (+798 test lines / +365 prod lines)
- **Code Delivered:** +1,163 / -413 lines across 18 code files (excluding docs)
- **Reasoning Effort:** 78,313 reasoning tokens (25.6% of generated output)
- **Tool Health & Self-Correction:** 787 tool calls (61 self-corrections, 92.2% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-architecture (7 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-architecture' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (10 cycles) — Subagent roles 'coder' and 'tester' alternated for 10 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.8k tokens) — Context tokens reached 250.8k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
Anomaly analysis is not recorded for every required owner/group; recorded findings remain visible. A complete source-backed refresh is required for recovery.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 141,882 | 1,447 | 668,672 | 82.5% | 140.8k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 140,881 | 30,388 | 13,778,688 | 99.0% | 195.0k | Not recorded | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 2 | 235,464 | 34,505 | 22,832,384 | 99.0% | 228.1k | Not recorded | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 12 | 1,013,696 | 115,880 | 71,816,448 | 98.6% | 303.0k | Not recorded | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **reviewer-architecture** | 1 | 139,353 | 9,584 | 2,298,240 | 94.3% | 127.7k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 154,311 | 13,366 | 3,968,640 | 96.3% | 149.3k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 123,860 | 11,511 | 2,869,888 | 95.9% | 133.5k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 114,513 | 10,799 | 2,731,776 | 96.0% | 125.2k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **21** | **2,063,960** | **227,480** | **120,964,736** | **98.3%** | **303.0k** | **Not recorded** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 11 | 581,302 | 37,423 | 5,636,224 | 90.7% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6.2-sol** | 10 | 1,482,658 | 190,057 | 115,328,512 | 98.7% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **Total** | **21** | **2,063,960** | **227,480** | **120,964,736** | **98.3%** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

Retained recorded history: 11 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
