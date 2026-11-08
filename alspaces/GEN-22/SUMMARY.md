# GEN-22: Human-readable workspace artifacts

Shortened eight active instruction files by 4,601 whitespace-delimited words
(44% of their iteration-04 baseline), including Markdown/frontmatter/examples.
This removes duplicated workflow and policy rather than relocating it to new files.

AGENTS.md owns repository policy, skills own workflows/artifact definitions and
personas contain only assigned duties. The development skill no longer repeats
command recipes or delegation prompts; the orchestrator follows that workflow.
PO specs/tickets still lead with goal, context and observable acceptance, with
optional technical notes last. Review/installation protocols and application code
are unchanged; required approval, evidence, credit and delivery gates remain.

Scope follows the [contract](CONTRACT.md) and current
[instruction-reduction plan](iterations/04/PLAN.md), which preserves the prior
[PO format](iterations/02/PLAN.md) and [skill ownership](iterations/03/PLAN.md).
Historical tickets and metric formatting remain unchanged.

Verification: `make` passed on integrated revision `4c88edf3`: Go race suites,
frontend units/integration, static and Python workflow checks, embedded assets
and platform builds. Integration added only PO-21 planning artifacts from main.
Changed-document links/anchors, all four edited skills' metadata and
`git diff --check` passed. No active link to the removed guide remains.

An independent instruction walkthrough (gpt-6.1-sol/high, inherited settings)
resolved four cases without contradictions: resumed approval with incomplete
review, ad hoc delivery, PO draft consolidation/session separation and management
merge authority. It did not inspect live ticket/PR readiness or prove production
execution quality. Iteration 04 links/anchors, skill metadata and whitespace
checks passed; integrated make is pending.

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
