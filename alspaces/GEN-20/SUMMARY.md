# GEN-20: Clean up deleted-thread worktrees

Initial state: 33 registered checkouts; clean current checkout at 8d494d2. T3 retained only two top-level threads. Read-only T3 state inspection found only this thread with an active provider session; surviving child threads had idle provider sessions.

Scope: remove 27 clean unused checkouts without force; retain main, this thread, MGT-2, and three dirty checkouts (GEN-1 guide, AL-62, release-v2-0-0). Keep all local branches and commits. Preserve AL-68 ignored capture data under `/home/niltonduarte/.t3/scratch/GEN-20/t3-16a70bd4-capture/` outside Git. Detailed initial inventory and removal outcomes are in iterations/01.

T3's worktree handoff tool explicitly states: “The worktree stays after the thread is deleted.” This establishes documented behavior, not a cleanup failure. The tool does not document the design rationale. T3 database records and application code were not modified.

Verification: all 27 ordinary removal commands succeeded. Six registered worktrees remain and no unregistered directories remain under the AgenticLens worktree root. All 33 initial branch refs still resolve. Rechecked the three dirty checkouts against the initial status inventory; unchanged. No T3 state database mutations, application changes, or delegated sessions.

Integration: origin/main fetched and already contained in the feature branch at the cleanup record commit 01b1109. Full make passed on 01b1109 plus summary-only edits: static checks, release/installation/capture script tests, Go race tests, all 790 frontend tests across 36 files, frontend build, and binary builds. Generated assets follow the existing ignore policy; Git status contains only this ticket summary. Final telemetry follows immediately before the final commit.

Telemetry expectation: Codex root 01a1a390-8781-7182-95b1-8e17c84ad767, confirmed from CODEX_THREAD_ID and local rollout session metadata matching this checkout. No children launched and no existing GEN-20 ledger or other-harness sessions to preserve.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 0m 0s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 79.9k tokens (40% of 200k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 1,812 reasoning tokens (18.0% of generated output)
- **Tool Health & Self-Correction:** 33 tool calls (6 self-corrections, 81.8% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 67,685 | 8,278 | 2,165,760 | 97.0% | 79.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **67,685** | **8,278** | **2,165,760** | **97.0%** | **79.9k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 67,685 | 8,278 | 2,165,760 | 97.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **67,685** | **8,278** | **2,165,760** | **97.0%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-03T22:46:22.874851857Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.

Final telemetry verified: exactly the expected Codex root, one session and no children. No prior ledger existed. The native CLI does not support the instruction template flags --al-offline/--al-project; generation used its supported flags from this checkout with explicit title, type, source and literal nonce. Pricing evidence remains unavailable where reported in the ledger; no zero cost was manufactured. Generation covers calls through its snapshot; subsequent verification and delivery calls are outside that snapshot.
