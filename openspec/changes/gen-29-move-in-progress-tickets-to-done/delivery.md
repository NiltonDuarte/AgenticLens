# chore(GEN-29): move in-progress tickets to done

## Outcome

The AL-96 and FIX-20 specifications moved to `alspaces/plan/done/` and left the active roadmap. Links to those specifications now point to the completed files. AL-97 moved to Ready.

## Delivery evidence

- AL-96 merged as `83f6f6c1`. The installed Linux amd64 capture gate passed on Codex 0.170.0 and OpenCode 1.19.5, and the archived-session binding follow-up passed focused race checks and integrated Make. The stable release decision remains separate. [Delivery](../al-96-verify-automatic-ticket-metrics-capture-end-to-end-for-v3/delivery.md) · [PR #156](https://github.com/NiltonDuarte/AgenticLens/pull/156)
- FIX-20 merged as `dda513dd`. Its delivery records cancellation and preservation checks for native source reads and backfill, independent review, and integrated verification. [Delivery](../fix-20-stop-source-reads-when-callers-cancel/delivery.md) · [PR #157](https://github.com/NiltonDuarte/AgenticLens/pull/157)

## Verification

At `22791284` over main `343b1b00`, the roadmap check confirmed two byte-identical moves, the expected active ticket set, AL-97 readiness and 39 local links. `git diff --check`, strict OpenSpec validation and the three OpenSpec workflow tests passed. The integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed.

Final extraction matched the expected Codex root with no child sessions. The source still showed an open turn at extraction, so its duration remains provisional. The ledger and generated summary preserve the available usage and pricing coverage.

## Pull request

Move the delivered AL-96 and FIX-20 specifications to `alspaces/plan/done/`, remove their entries from the active roadmap, and update links to the moved files. Promote AL-97 to Ready now that its sole listed dependency, AL-96, has been delivered.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 15m 37s (recorded duration including 937s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 82.7k tokens (18% of 450k healthy ceiling, 0 compactions) | Subagents peak 0 (None)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 0.0 avg turns/run (peak: 0 turns)
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 2,241 reasoning tokens (15.4% of generated output)
- **Tool Health & Self-Correction:** 69 tool calls (5 self-corrections, 92.8% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **Unknown** | 1 | 98,654 | 12,278 | 5,172,352 | 98.1% | 82.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **98,654** | **12,278** | **5,172,352** | **98.1%** | **82.7k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.2-sol** | 1 | 98,654 | 12,278 | 5,172,352 | 98.1% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **1** | **98,654** | **12,278** | **5,172,352** | **98.1%** | **Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations)** |
<!-- agentic-lens:metrics:end -->
