# GEN-22: Human-readable workspace artifacts

Artifacts now have distinct purposes and concise human-facing content. PO specs
and tickets lead with goal, context and acceptance; optional technical notes come
last in the same document. AGENTS.md owns repository policy, skills own workflows
and artifact definitions, and personas own assigned duties.

The development skill restores operational detail after the user rejected its
compressed overview: startup/resume checks, assignment/return contracts, parent
checks, approval boundaries, correction routing and ordered delivery gates.
Technical rules, detailed review mechanics and exact commands stay at their
existing sources. The other instruction reductions remain. Across eight files,
the final net reduction from iteration 04's baseline is 3,882 whitespace-delimited
words (37%), counting Markdown/frontmatter/examples without relocating the bulk.

Development is version 0.1.1; PO/review are 0.1.1, and management gained its first
VERSION record at 0.1.1. All four version records pass format/content checks.
Scope follows the [contract](CONTRACT.md) and current
[development refinement](iterations/05/PLAN.md), preserving the prior
[PO format](iterations/02/PLAN.md), [skill ownership](iterations/03/PLAN.md) and
[instruction reduction](iterations/04/PLAN.md). Application code, historical
tickets and metric formatting remain unchanged.

Verification: `make` passed on `d16803aa` after integration with origin/main
(already current): Go race suites, frontend units/integration, static and Python
workflow checks, embedded assets and platform builds. Development skill
links/anchors, metadata, version format and `git diff --check` passed. The
independent walkthrough's content fingerprint matches the committed skill.

An independent development walkthrough (gpt-6.1-sol/high, source-observed)
resolved all four scenarios: new feature with an unresolved policy choice,
resumed correction, failed/incomplete review and an integration conflict changing
reviewed behavior. Each had a clear owner, assignment, return, parent check and
gate; no material issue was found. The earlier general-policy walkthrough also
passed. These are documentation usability checks, not proof of production
execution quality or application correctness. Hosted CI has not been observed
for this update.

Telemetry source metadata confirms root `01a1bbce-df16-7632-ac2d-9a565b47e918`
and both independent-walkthrough children with native parent lineage. Final
extraction matches all three sessions and preserves prior identities. There was no other-harness history.
Unavailable roles/pricing, the provisional open turn and post-extraction delivery
calls remain coverage limits.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 11
- **Active Agent Duration:** 44m 44s (recorded duration including 424s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 203.7k tokens (45% of 450k healthy ceiling, 0 compactions) | Subagents peak 37.4k (Default)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Default 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 16,502 reasoning tokens (22.0% of generated output)
- **Tool Health & Self-Correction:** 133 tool calls (4 self-corrections, 97.0% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 162,664 | 55,432 | 15,426,816 | 99.0% | 203.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **default** | 2 | 54,827 | 3,242 | 303,360 | 84.7% | 37.4k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **3** | **217,491** | **58,674** | **15,730,176** | **98.6%** | **203.7k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 3 | 217,491 | 58,674 | 15,730,176 | 98.6% | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Total** | **3** | **217,491** | **58,674** | **15,730,176** | **98.6%** | **Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations)** |

Capture observation finalized: 2026-11-08T16:55:25.597827926Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 2; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 2; reserved bytes: 3375246. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bbce-df16-7632-ac2d-9a565b47e918: current; last successful source read 2026-11-08T16:55:25.579452976Z; accepted source checkpoint 2026-11-08T16:55:25.377Z.
Owner codex/01a1bbe9-614e-7450-9a71-4d311f68c211: current; last successful source read 2026-11-08T16:55:25.514011111Z; accepted source checkpoint 2026-11-08T16:07:34.196Z.
