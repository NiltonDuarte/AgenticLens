# AL-63 capability audit

Updated on November11 after merging origin/main4e5c4583 at d32d1836. The [capability report](CAPABILITIES.md) now distinguishes the current integrated behavior from the original November8 observations. Original AL-63 acceptance remains unchanged in plan/done; AL-65 stays Ready. Roadmap conflict resolution also preserves upstream AL-80 Ready and AL-83 In Progress. No application changes are introduced relative to main.

Current fixture revalidation confirms configured Codex-home hydration is fixed by AL-70 and failed native child lookup now returns typed unavailable diagnostics while preserving trace content (AL-69). Mixed recovery's old table-existence selector is replaced by complete-family validation, which explicitly rejects unsupported mixed layouts. OpenCode metadata child-link loss and Codex explicit function-output error counters still reproduce. Live v2 and Codex support remain limited; malformed-part diagnostics remain a static residual. [Candidate attribution](iterations/02/REVIEW_SUMMARY.json) preserves all original discovery credit and updates dispositions without rewriting historical reviewer reports.

The three focused live tests passed (0.891s): child-lookup diagnostics, structured v1 execution parity, and mixed-source hydration/preservation. A private current-module probe verified configured-home complete hydration/one event, empty metadata child linkage, explicit mixed-layout rejection, and zero summary errors for failed/error function outputs. These are constructed fixtures; no fresh native source sessions or backfill runs occurred. Original native counts (Codex232events/5owners; OpenCode27/4events) and native Codex copy-backfill/repeat results remain November8 evidence. Original private probe directories are no longer available; they were not represented as rerun.

The eight independent original review reports and initial INCOMPLETE outcomes remain in the [review record](iterations/02/REVIEW.md). This maintenance is a bounded source/test/probe refresh, not a new broad correctness review or native-parity certification. Full make after integration at d32d1836 passed (exit0): 227 frontend unit and 751 integration tests, backend race/integration, static/release/install/capture/orchestration checks and multi-platform builds. The built Linux CLI version command succeeded; later edits are audit documentation only. Final telemetry is refreshed with that current built CLI because the shared pinned installed engine predates the integrated configuration contract; no installation, pin or local capture policy is changed. Native usage is limited to the established root/child graph; four original Luna service-session links remain unproven.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 8
- **Active Agent Duration:** 48m 55s (recorded duration including 657s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 188.8k tokens (42% of 450k healthy ceiling, 0 compactions) | Subagents peak 129.5k (Reviewer-behavior)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 24,948 reasoning tokens (21.1% of generated output)
- **Tool Health & Self-Correction:** 250 tool calls (23 self-corrections, 90.8% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 156,469 | 51,297 | 12,643,328 | 98.8% | 188.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 1 | 96,524 | 11,312 | 1,920,640 | 95.2% | 103.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 108,653 | 13,625 | 2,883,968 | 96.4% | 129.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 92,417 | 10,347 | 2,242,432 | 96.0% | 110.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 85,512 | 6,449 | 1,026,048 | 92.3% | 102.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **5** | **539,575** | **93,030** | **20,716,416** | **97.5%** | **188.8k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 5 | 539,575 | 93,030 | 20,716,416 | 97.5% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **5** | **539,575** | **93,030** | **20,716,416** | **97.5%** | **Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations)** |

Capture observation finalized: 2026-11-08T22:25:44.895732985Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 1; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 1; reserved bytes: 2646003. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1bd1f-69c1-7043-9bf4-8d4fe9a53c9b: current; last successful source read 2026-11-08T22:25:44.880012534Z; accepted source checkpoint 2026-11-08T22:25:44.784Z.
<!-- agentic-lens:metrics:end -->
