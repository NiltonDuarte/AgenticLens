# fix(AL-71): keep dashboard requests and live data reliable

## Outcome

Comparisons and Progression now fence asynchronous results through the current request. Fleet retains paired summary and workspace results. Progression shows a UTC window anchored to the newest usable ledger in the selection, so a filter that leaves the selection intact also leaves its historical chart intact.

REST and SSE reject malformed live data while retaining the last valid observation. Fallback polling keeps one request in flight, reports the actual failure and recovers at the configured cadence. Trace warnings stay with the affected session. An older REST trace cannot erase a newer SSE warning.

## Verification and review

Integrated revision `3e5377a9` passed `make -j2` with one Vitest worker and a local 20-second timeout for frontend integration. Earlier attempts timed out under host load. The successful run passed all 235 frontend unit tests, all 849 frontend integration tests, the Go race suites, static checks, installation and distribution checks, and all native builds. Earlier standalone reruns of the two capture assertions passed as well.

After merging main through FIX-21 (`be630e36`), integrated revision `0efb03d1` ran `make -j2`.

- **Passed:**
  - 235 frontend unit and 859 frontend integration tests;
  - the domain race suites, vet, typecheck and lint;
  - the installation, distribution, platform and orchestration checks;
  - all eight native builds;
  - every backend integration package except two, including `http` and `live`, the only backend packages this change touches.
- **Not passed locally:** `internal/adapters/capture` and `cmd/agentic-lens` exceeded the 10-minute default.
  - A serial rerun with a 90-minute timeout failed a different set of worker-timing tests, with `SQLITE_BUSY` and deadline errors at a host load average near 15.
  - This change does not touch those packages, and CI passed them on `be630e36`. CI on the pull request is the deciding run for them.

An Opus review of the merged candidate found no blocking defects. It confirmed that the stricter live-data validation still fits main's new `claude` session source. Three minor follow-ups remain:
- a malformed pushed trace can discard a valid trace read that is still loading;
- trace warnings can persist after polling fallback or Retry;
- a malformed alert is labeled as a connection error.

The native workspace handler and SQLite-backed REST/SSE producer verified the shared wire fixtures. Progression boundary checks passed in non-UTC time zones. The Linux manager and standalone engine reported their versions. From an isolated configured project, the engine served the dashboard and JS/CSS byte-for-byte against both generated asset trees and answered the workspace HTTP request.

Behavior, data integrity, architecture and test effectiveness were reviewed independently with Sol 6.2 high and Luna 6 medium. The first passes found diagnostic recovery, timestamp and numeric validity, and regression-coverage gaps. Focused rechecks verified the corrections against the final application and test fingerprint `8e153a7d89492dc2ee27903078330879ec40848f0941ea984696da2e52281487`. [Review findings and attribution](evidence/REVIEW_SUMMARY.json)

## Limits

Automatic capture reconciliation reported a degraded collection. Exact-session recovery and final root extraction verified all 21 expected Codex sessions, including the root and nine source-proven children, and preserved all 16 previously accepted sessions. Five independent sessions lacked verified automatic bindings and were recovered from their exact local source. API-equivalent costs remain unavailable.

**Restored ledger.** An automatic capture refresh by an older engine had removed 7 of the 9 `child_session_id` links on root `01a205df-6c09-72a2-a8a4-e546a27a0fe7`. That rewrite was never committed. After FIX-21, `generate --capture-auto` from an engine built at `0efb03d1` was run against this ledger:
- all 21 Codex sessions and all 9 root links match the accepted state at `1dfb0fdf`;
- no session was lost;
- only the root's final completion time, active duration and token totals were updated;
- all 16 AL-71 capture registrations are current.

**Limit.** Five unassigned capture rows still report that their rollout path is missing. These are the archived independent sessions; their facts remain in the ledger. The rows carry no parent proof, so FIX-21 leaves them isolated rather than recovered.

## Pull request

### Changes

- Keep dashboard results, warnings and errors tied to the active filter request.
- Use the newest selected ledger for Progression presets and display the effective UTC dates.
- Reject malformed REST and SSE snapshots, preserve usable data, and serialize fallback polling with scoped recovery.

### Verification

The integrated `make -j2` gate passed with 235 frontend unit and 849 frontend integration tests, plus Go race and distribution checks. The integration timeout was adjusted locally for host load. The standalone engine served the current embedded assets and workspace API.

After merging main (FIX-21), `0efb03d1` passed:
- 235 frontend unit and 859 frontend integration tests;
- static checks and native builds;
- every backend integration package except capture and the CLI.

Those two timed out or failed timing tests on the overloaded host, so CI decides them. The AL-71 ledger was restored with the FIX-21 engine and again carries all 9 child links and all 21 sessions.

[Pull request #160](https://github.com/NiltonDuarte/AgenticLens/pull/160)

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 46
- **Active Agent Duration:** 3h 6m (recorded duration including 57s provisional)
- **Context Window Usage:** Root peak 306.8k tokens (68% of 450k reference ceiling, 0 compactions) | Subagents peak 228.1k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.4 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 3.0, Tester 2.5, Reviewer 1.0, Reviewer-architecture 3.0, Reviewer-behavior 3.0, Reviewer-data 3.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 2.2:1 (+798 test lines / +365 prod lines)
- **Code Delivered:** +1,163 / -413 lines across 18 code files (excluding docs)
- **Reasoning Effort:** 97,898 reasoning tokens (25.2% of generated output)
- **Tool Health & Self-Correction:** 917 tool calls (66 self-corrections, 92.8% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (3 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (10 cycles) — Subagent roles 'coder' and 'tester' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-architecture (7 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-architecture' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.8k tokens) — Context tokens reached 250.8k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.
Anomaly analysis is not recorded for every required owner/group; recorded findings remain visible. A complete source-backed refresh is required for recovery.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 141,882 | 1,447 | 668,672 | 82.5% | 140.8k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 140,881 | 30,388 | 13,778,688 | 99.0% | 195.0k | Not recorded | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 2 | 235,464 | 34,505 | 22,832,384 | 99.0% | 228.1k | Not recorded | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer** | 1 | 54 | 16,984 | 1,914,711 | 100.0% | 98.5k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 13 | 1,020,454 | 161,671 | 86,301,449 | 98.8% | 306.8k | Not recorded | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **reviewer-architecture** | 1 | 139,353 | 9,584 | 2,298,240 | 94.3% | 127.7k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 154,311 | 13,366 | 3,968,640 | 96.3% | 149.3k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 123,860 | 11,511 | 2,869,888 | 95.9% | 133.5k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 114,513 | 10,799 | 2,731,776 | 96.0% | 125.2k | Not recorded | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **23** | **2,070,772** | **290,255** | **137,364,448** | **98.5%** | **306.8k** | **Not recorded** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 2 | 278 | 60,678 | 12,744,544 | 100.0% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **gpt-6-luna** | 11 | 581,302 | 37,423 | 5,636,224 | 90.7% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6.2-sol** | 10 | 1,489,192 | 192,154 | 118,983,680 | 98.8% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **Total** | **23** | **2,070,772** | **290,255** | **137,364,448** | **98.5%** | **Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations)** |

Retained recorded history: 21 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
