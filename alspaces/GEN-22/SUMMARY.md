# GEN-22: Human-readable workspace artifacts

Added [Writing workspace artifacts](../../docs/WORKSPACE_ARTIFACTS.md) as the
canonical guide. Each document has one purpose, practical length targets and
rules for linking evidence. Summaries report current outcomes instead of
accumulating routine execution history.

Aligned contributor guidance, planner/orchestrator personas, relevant skills and
both PO templates. Removed mandatory layer inventories, repeated delivery
checklists, the PR ban on test evidence and automatic new iterations on resume.

PO specs and tickets now lead with the goal, context and observable acceptance.
Necessary implementation notes follow in the same document. Shared initiative
context is linked rather than copied into every ticket; separate human/agent
specifications are not the default.

Scope follows the [contract](CONTRACT.md) and the user's
[PO-ticket clarification](iterations/02/PLAN.md). Historical tickets, application
code and generated metrics formatting remain unchanged. Required review credit,
verification and unresolved limits remain protected.

Verification: `make` passed on `2ff67d95` after integration with origin/main
(already current): Go race suites, frontend units/integration, static and Python
workflow checks, embedded assets and platform builds. Changed-document links,
the new PO section anchor, PO skill metadata and `git diff --check` also passed.
Earlier iteration checks remain valid for unchanged guidance.

The documentation coordinator (gpt-6.1-sol/high) checked both templates against
recent PO-20/AL-79/AL-72 examples for goal visibility and duplication; no confirmed
issues remain. There was no independent application review or behavioral forward
test. Length targets guide future writing, without automatic enforcement.
Hosted CI has not been observed for this update.

Telemetry: source Codex; expected root `01a1bbce-df16-7632-ac2d-9a565b47e918`
was confirmed in local source metadata, with no source-linked children and no
subagents dispatched. Generated membership matches that root exactly. No prior
GEN-22 ledger or other-harness history existed. The open turn remains provisional;
role and pricing are unavailable, and post-extraction delivery calls are excluded.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 14m 19s (recorded duration including 355s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 99.9k tokens (22% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 3,071 reasoning tokens (16.3% of generated output)
- **Tool Health & Self-Correction:** 53 tool calls (2 self-corrections, 96.2% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 84,556 | 15,810 | 3,504,896 | 97.6% | 99.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **84,556** | **15,810** | **3,504,896** | **97.6%** | **99.9k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 84,556 | 15,810 | 3,504,896 | 97.6% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **84,556** | **15,810** | **3,504,896** | **97.6%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-08T15:52:00.053927004Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 1350153. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bbce-df16-7632-ac2d-9a565b47e918: current; last successful source read 2026-11-08T15:52:00.041727262Z; accepted source checkpoint 2026-11-08T15:51:59.904Z.
