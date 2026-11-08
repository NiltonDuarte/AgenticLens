# AL-63 planning outcome

Refined the existing capability-audit specification so an operator can compare OpenCode v1, OpenCode v2 and Codex without mistaking declared support or fixture coverage for native runtime verification. All eight capability areas remain in scope. Evidence must identify versions, operations, source identities, results and limits, with historical evidence reused only within its stated applicability.

The specification now explicitly covers partial/unavailable hydration, complete-empty traces, unavailable same-harness owners, smaller authoritative snapshots and aggregate/idempotence behavior under the existing portable-ledger contract. FIX-18 retains live-support corrections and REV-7 retains read-contract investigation. No capability parity, new runtime support or changed backfill default is promised.

This is planning delivery under the supplied AL-63 identifier. No application, tests, runtime scripts, developer contract or new tickets changed. The audit itself remains Ready and unexecuted; AL-65 remains blocked. No native compatibility, successful backfill or original-ledger preservation probe is claimed. [Iteration plan](iterations/01/PLAN.md) records scope and verification boundaries; acceptance lives only in the [ticket specification](../plan/to-do/AL-63%20verify%20source%20capabilities%20across%20harnesses.md).

Document checks passed for local links, three source families, eight capabilities, unchanged registry and documentation-only scope. At planning commit 0a8cfcee, origin/main integration was a no-op and full `make` passed (exit 0): static checks, backend race/integration tests, 190 frontend unit and 678 integration tests, installation/capture/release/orchestration checks and multi-platform builds. No required stage skipped. npm reported 12 dependency advisories; bundle-size and React act warnings remain outside this planning scope. Initial checkout was clean at eb7d169c; fetched origin/main matched it. Author inspection confirmed source enums identify OpenCode/Codex, historical hydration dispatches per source, and current live SQLite queries reference v1 tables. These source observations motivate separate matrix columns; they do not certify runtime behavior.

Telemetry coverage baseline: native Codex root 01a1bd1f-69c1-7043-9bf4-8d4fe9a53c9b, matching this worktree; source metadata inventory found no native children and no delegates were launched. No pre-existing AL-63 ledger or other-harness sessions require merging. Final extraction must retain this identity; unavailable measurements remain unknown.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 3
- **Active Agent Duration:** 6m 48s (recorded duration including 408s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 63.5k tokens (14% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 1,419 reasoning tokens (15.6% of generated output)
- **Tool Health & Self-Correction:** 29 tool calls (3 self-corrections, 89.7% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 60,001 | 7,651 | 1,563,648 | 96.3% | 63.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **60,001** | **7,651** | **1,563,648** | **96.3%** | **63.5k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 1 | 60,001 | 7,651 | 1,563,648 | 96.3% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **60,001** | **7,651** | **1,563,648** | **96.3%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

Capture observation finalized: 2026-11-08T21:52:04.509643571Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.

Final extraction verified schema 6, docs classification and the exact single Codex root with no parent/children. Native cost and role are unavailable/unknown; the active turn has provisional duration at the extraction cutoff. No completed-source or current-caller claim is made. The CLI required an explicit ticket title; generation then succeeded without source/configuration overrides.
