# GEN-22: Human-readable workspace artifacts

Added [Writing workspace artifacts](../../docs/WORKSPACE_ARTIFACTS.md) as the
canonical guide. Each document has one purpose, practical length targets and
rules for linking evidence. Summaries report current outcomes instead of
accumulating routine execution history.

Aligned AGENTS.md, contributor/setup guidance, planner/orchestrator personas,
development/management/PO/review skills and the specification template. Removed
mandatory layer inventories, the PR ban on test evidence and automatic new
iterations on resume. Required review attribution, verification, unresolved limits
and generated metrics remain intact.

Scope follows the user's opening request and [contract](CONTRACT.md).
[Iteration 01](iterations/01/PLAN.md) records the clean baseline and delivery plan.
Historical ticket artifacts and application code were not changed.

Verification: `make` passed on `938e7e44` after integration with origin/main
(already current): Go race suites, frontend units/integration, static and Python
workflow checks, embedded assets and platform builds. Changed-document links,
four edited skills' metadata validation and `git diff --check` also passed.

The documentation coordinator (gpt-6.1-sol/high) checked instruction consistency;
no confirmed issues remain. There was no independent application review or
behavioral forward test. Length targets guide future writing; this change does
not enforce them automatically. Hosted CI has not been observed.

Telemetry: source Codex; expected root `01a1bbce-df16-7632-ac2d-9a565b47e918`
was confirmed in local source metadata, with no source-linked children and no
subagents dispatched. Generated membership matches that root exactly. No prior
GEN-22 ledger or other-harness history existed. The open turn remains provisional;
role and pricing are unavailable, and post-extraction delivery calls are excluded.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 7m 34s (recorded duration including 454s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 77.1k tokens (17% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 1,434 reasoning tokens (13.8% of generated output)
- **Tool Health & Self-Correction:** 35 tool calls (1 self-corrections, 97.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 65,591 | 8,946 | 1,645,440 | 96.2% | 77.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **65,591** | **8,946** | **1,645,440** | **96.2%** | **77.1k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 65,591 | 8,946 | 1,645,440 | 96.2% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **65,591** | **8,946** | **1,645,440** | **96.2%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-08T15:45:14.233653449Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
