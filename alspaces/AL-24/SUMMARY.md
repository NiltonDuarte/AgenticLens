# Ticket AL-24: Dynamic Role and Model Aggregations and Presentation

## Summary of Completed Work

### Domain Core Layer (`backend/internal/domain`)
- **Dynamic Role Classification & Ordering (`enums.go`)**:
  - Implemented pure domain helpers `IsSubagentRole`, `OrderRoles`, and `OrderSubagentRoles`.
  - Preserved canonical role order (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`) while dynamically sorting non-canonical roles alphabetically ascending.
- **Dynamic Subagent Aggregations (`aggregations.go`)**:
  - Refactored `ComputeWorkflowEfficiency` to eliminate fixed iteration over `SubagentRoleOrder`, classifying any role satisfying `IsSubagentRole` as a subagent.
  - Dynamically aggregated subagent effort distribution percentages (`RoleEffortDistributionPct`) across all discovered roles with positive costs.
  - Dynamically aggregated subagent turn velocity (`SubagentTurnVelocity`) across all subagent invocations, calculating average turns per run, peak turns, and per-role turn metrics.
  - Implemented deterministic peak context tie-breaking in `ContextWindowHealth`: canonical subagent precedence followed by alphabetical ordering for custom subagents.
- **Model Metric Aggregation & Deterministic Sorting (`aggregations.go`)**:
  - Refactored `ComputeModelMetrics` to dynamically aggregate all unique model identifiers without truncation or hardcoded filters.
  - Implemented primary sorting on `CostUSD` descending with deterministic secondary tie-breaking on `Model` string ascending.
  - Updated `ComputeRoleMetrics` and `ComputeThreadSummaries` to maintain canonical role priority and secondary alphabetical model tie-breaking.
- **Configured Role Mapping Transformation (`aggregations.go`)**:
  - Extended `DeriveMetricsLedger(factLedger, roleMappings...)` to project mapped copies for metrics derivation while preserving untouched raw facts in `AgenticMetricsLedger.Facts`.
  - Mapped roles using `RoleMappings.ResolveRole(...)`: raw personas (e.g. `build`) roll up under canonical mapped roles (e.g. `orchestrator`) across derived totals, role metrics, workflow efficiency, and thread summaries.
  - Roles mapping to `RoleOrchestrator` are automatically excluded from subagent turn velocity and effort distribution calculations.

### Adapters & CLI Layer (`backend/internal/adapters`, `backend/cmd/`)
- **Raw Persona Fact Extraction (`opencode/extractor.go`)**:
  - Refactored `ExtractFacts` to record the unmutated persona from SQLite `session.agent` into `InvocationFact.Role` and `EventFact.Role` without premature coercion to `orchestrator`.
  - Added clean fallback for blank agents (`orchestrator` for root sessions, `unassigned` for child sessions).
- **Repository Integration (`repository.go`)**:
  - Updated `ScanWorkspaces` and `LoadTicketWorkspace` to load `.agenticlens/config.yaml` and pass `cfg.RoleMappings` to `DeriveMetricsLedger`.
- **CLI Generation Pipeline (`main.go`)**:
  - Updated `runGenerate` to persist raw facts in `agentic_metrics.json` via `adapters.SaveFactLedger` while rendering mapped roles in markdown summary tables (`FormatMarkdownMetricsTable`).
- **Dynamic Markdown Telemetry Formatter (`formatters.go`)**:
  - Formats `Role Effort Distribution (Subagents)` and `Subagent Turn Velocity & Thrashing` using `OrderSubagentRoles`.

### Presentation Layer (`web/src`)
- **Domain Types & Theme Tokens (`types.ts`, `theme.ts`)**:
  - Widened `AgentRole` from a closed union to `type AgentRole = string` with helper `isCanonicalRole`.
  - Implemented `hashStringToColor`, `getRoleColor`, and `getModelColor` with a 12-color high-contrast accessible palette (`DYNAMIC_PALETTE`), mapping canonical roles to established theme tokens and arbitrary strings to deterministic colors.
- **Interactive UI Components (`components/`, `views/`)**:
  - `DistributionCard.tsx`: Added color dot indicators and dynamic progress track styling with defensive guards against zero costs and NaN.
  - `SaturationBar.tsx`: Rendered all roles present in `peak_by_role` with role-specific color badges and progress fills, supporting dynamic `subagents_peak_role` labels.
  - `TicketDeepDive.tsx`: Updated `SDLC Iterations` tooltip to reflect all participating roles; styled turn velocity and resource consumption tables with dynamic role and model colors.

### Quality & Verification
- **Adversarial Domain Tests (`aggregations_test.go`)**: Verified role classification, dynamic subagent distributions, deterministic model tie-breaking, role mapping execution invariance, raw fact immutability, and zero-division edge cases.
- **Extractor Boundary Tests (`opencode/extractor_test.go`)**: Verified raw persona preservation for `build` and custom agents, plus empty-agent fallbacks.
- **Repository Tests (`repository_test.go`)**: Verified workspace loading applies configured role mappings while preserving raw facts in `Facts`.
- **CLI Integration Tests (`main_test.go`)**: Verified raw facts in `agentic_metrics.json` vs mapped roles in `SUMMARY.md`.
- **Frontend Vitest Suites (`theme.test.ts`, `Components.test.tsx`, `TicketDeepDive.test.tsx`)**: Verified color hashing determinism, component rendering, and zero NaN values across 15 test suites and 237 tests.
- **Static Invariants**: Zero `any` or `interface{}` across Go and TypeScript, zero emojis, and zero skipped tests.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 4h 30m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 169.8k tokens (85% of 200k healthy ceiling, 2 compactions) | Subagents peak 168.2k (Coder)
- **Role Effort Distribution (Subagents):** Planner 17.8%, Coder 52.6%, Tester 20.6%, Reviewer 9.0%
- **Subagent Turn Velocity & Thrashing:** 77.1 avg turns/run (peak: 119 turns) | Planner 65.5, Coder 100.3, Tester 71.5, Reviewer 48.0
- **Test-to-Production Code Ratio:** 4.8:1 (+2,290 test lines / +475 prod lines)
- **Code Delivered:** +2,765 / -2,499 lines across 40 code files (excluding docs)
- **Reasoning Effort:** 254,135 reasoning tokens (59.1% of generated output)
- **Tool Health & Self-Correction:** 985 tool calls (18 self-corrections, 98.2% success rate)

### Anomalies & Health Flags
- **Status:** 8 execution anomalies detected (4 CRITICAL, 4 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.3k tokens) — Context tokens reached 162.3k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (162.0k tokens) — Context tokens reached 162.0k, breaching the 160k critical ceiling.
- [CRITICAL] **TOOL_HIGH_FAILURE_RATE**: High tool failure rate: mcp-gateway_mcp-exec (100.0%) — Tool 'mcp-gateway_mcp-exec' failed 5 out of 5 calls (100.0% failure rate, threshold: >=50%).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+72.7k tokens) — Context tokens surged by 72.7k to 72.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+58.4k tokens) — Context tokens surged by 58.4k to 58.4k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+78.7k tokens) — Context tokens surged by 78.7k to 78.7k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (3 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 1,693,880 | 41,709 | 14,912,640 | 89.8% | 169.8k | ELEVATED | $2.65 |
| **planner** | 2 | 1,089,283 | 24,049 | 9,057,616 | 89.3% | 144.3k | ELEVATED | $1.82 |
| **coder** | 4 | 3,010,548 | 68,848 | 32,949,110 | 91.6% | 168.2k | ELEVATED | $5.38 |
| **tester** | 2 | 1,352,125 | 31,929 | 11,356,136 | 89.4% | 132.7k | ELEVATED | $2.11 |
| **reviewer** | 2 | 571,743 | 9,531 | 4,892,742 | 89.5% | 104.5k | ELEVATED | $0.93 |
| **Total** | **11** | **7,717,579** | **176,066** | **73,168,244** | **90.5%** | **169.8k** | **ELEVATED** | **$12.89** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 11 | 7,717,579 | 176,066 | 73,168,244 | 90.5% | $12.89 |
| **Total** | **11** | **7,717,579** | **176,066** | **73,168,244** | **90.5%** | **$12.89** |
