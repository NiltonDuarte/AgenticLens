# GEN-22: Human-readable workspace artifacts

Artifact definitions now live in their producing skills:
[development](../../.agents/skills/agentic-lens-dev/SKILL.md#setup--workspace-structure),
[PO](../../.agents/skills/agentic-lens-po/SKILL.md#po-artifact-definitions),
[management](../../.agents/skills/agentic-lens-management/SKILL.md#management-artifact-definitions)
and [review](../../.agents/skills/agentic-lens-review/SKILL.md#evidence-and-outcomes).
AGENTS.md retains short shared writing rules; the standalone docs guide is removed.

Each document has one purpose and practical prose targets. Summaries report
current outcomes. PO specs/tickets lead with goal, context and observable
acceptance, with optional technical notes last in the same document. Shared
context is linked rather than copied; forced layer inventories and repeated
delivery checklists are removed.

Scope follows the [contract](CONTRACT.md), [PO-ticket clarification](iterations/02/PLAN.md)
and [skill-ownership decision](iterations/03/PLAN.md). Historical tickets,
application code and metric formatting remain unchanged. Required review credit,
verification and unresolved limits remain protected.

Verification: `make` passed on integrated revision `4c88edf3`: Go race suites,
frontend units/integration, static and Python workflow checks, embedded assets
and platform builds. Integration added only PO-21 planning artifacts from main.
Changed-document links/anchors, all four edited skills' metadata and
`git diff --check` passed. No active link to the removed guide remains.

The documentation coordinator (gpt-6.1-sol/high) checked ownership and instruction
consistency; no confirmed issues remain. Prior PO-template checks remain valid.
There was no independent application review or behavioral forward test. Length
targets guide writing without automatic enforcement. Hosted CI has not been
observed for this update.

Telemetry: source Codex; expected root `01a1bbce-df16-7632-ac2d-9a565b47e918`
was confirmed in local source metadata, with no source-linked children and no
subagents dispatched. Generated membership matches that root exactly. No prior
GEN-22 ledger or other-harness history existed. The open turn remains provisional;
role and pricing are unavailable, and post-extraction delivery calls are excluded.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 4
- **Active Agent Duration:** 21m 31s (recorded duration including 383s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 122.5k tokens (27% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 5,688 reasoning tokens (18.8% of generated output)
- **Tool Health & Self-Correction:** 70 tool calls (2 self-corrections, 97.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 101,285 | 24,517 | 5,625,088 | 98.2% | 122.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **101,285** | **24,517** | **5,625,088** | **98.2%** | **122.5k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 101,285 | 24,517 | 5,625,088 | 98.2% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **101,285** | **24,517** | **5,625,088** | **98.2%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-08T16:02:51.140744662Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 1695023. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bbce-df16-7632-ac2d-9a565b47e918: current; last successful source read 2026-11-08T16:02:51.132302957Z; accepted source checkpoint 2026-11-08T16:02:51.011Z.
