# docs(PO-26): define the v3 automatic capture release gate

## Outcome

Added [AL-96](../../../alspaces/plan/to-do/AL-96%20verify%20automatic%20ticket%20metrics%20capture%20end%20to%20end%20for%20v3.md) to the [roadmap](../../../alspaces/plan/TICKETS.md) as Ready and blocking the v3 stable release. The test begins with an empty destination, exercises supported native capture integrations, and compares the published session tree against the harness source. Capture defects found by the test must be fixed and the corrected release candidate must pass.

## Verification and limits

`make -j2` passed on `5209b46b` with current `origin/main` at `9308ac77`. It covered Go race tests, 231 frontend unit tests, 819 frontend integration tests, distribution checks, and native builds. The subsequent provisional-seconds requirement changed only Markdown. OpenSpec 2.0.0 strict validation, the roadmap and artifact link checks, and `git diff --check` passed for the revised planning artifacts. A separate AL-96 development session owns the test and any fixes. V3 capture behavior remains unverified.

In this worktree, native capture accepted the PO-26 binding and source read. It failed first publication because no ledger existed and still failed after an exact-ticket reconciliation retry. Explicit generation created the first ledger. The worker then automatically refreshed the one verified Codex root and committed the ledger and summary. This verifies refresh after initial generation and reproduces the first-publication gap AL-96 must cover. There were no child sessions or prior ledger owners. The source still reported an open turn, and cost was unavailable.

## Pull request

Add AL-96 as a v3 release gate. A real agent and its delegated sessions must publish to a fresh ticket ledger through the installed Codex and supported OpenCode integrations. The gate checks source-backed session identity and parentage, first publication without manual generation, and preservation without duplicate metrics on refresh. It also measures how quickly provisional seconds settle after native completion. Fix and retest any capture gaps before shipping v3.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 5
- **Active Agent Duration:** 16m 45s (recorded duration including 807s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 101.8k tokens (23% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 4,767 reasoning tokens (18.8% of generated output)
- **Tool Health & Self-Correction:** 89 tool calls (6 self-corrections, 93.3% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 112,731 | 20,616 | 5,549,056 | 98.0% | 101.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **112,731** | **20,616** | **5,549,056** | **98.0%** | **101.8k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.2-sol** | 1 | 112,731 | 20,616 | 5,549,056 | 98.0% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **112,731** | **20,616** | **5,549,056** | **98.0%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-19T12:35:09.269789944Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 109300924. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1f3b9-71fd-7c70-b57f-161476901e46: current; last successful source read 2026-11-19T12:35:09.262801513Z; accepted source checkpoint 2026-11-19T12:35:02.857Z.
<!-- agentic-lens:metrics:end -->
