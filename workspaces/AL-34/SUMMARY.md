# Work Summary: AL-34 Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

## Overview

Ticket AL-34 eliminates "Cumulative" from the Metric Progression granularity options across both Go backend and TypeScript frontend, strictly restricting temporal granularity to `hourly`, `daily`, and `weekly`. Furthermore, it enhances Chart 4 ("Delivered Code Line Churn & Net Growth") into a centered bi-directional diverging visualization with additions extending upward in green, deletions extending downward in red, and a cumulative net line growth trajectory line, backed by compact integer tick formatting on the Y-axis.

---

## Changes Grouped by Layer

### Domain Layer
- **Go Domain Enums & Models (`backend/internal/domain/enums.go`, `backend/internal/domain/models.go`)**:
  - Removed `GranularityCumulative` constant from `TimeGranularity` enum.
  - Updated `ParseTimeGranularity` to validate strictly case-insensitive `hourly`, `daily`, and `weekly`, rejecting `cumulative` and non-temporal values.
  - Added `CumulativeNetLinesAdded int64` (`cumulative_net_lines_added`) to `TimeBucketMetric` struct.
- **Go Domain Aggregations (`backend/internal/domain/aggregations.go`)**:
  - In `AggregateMetricsOverTime`, populated `CumulativeNetLinesAdded` as running cumulative lines added minus running cumulative lines deleted across chronological buckets.
- **TypeScript Types & Domain Logic (`web/src/domain/types.ts`, `web/src/domain/progression.ts`)**:
  - Constrained `TimeGranularity` type union to `'hourly' | 'daily' | 'weekly'`.
  - Added `cumulative_net_lines_added: number` to `TimeBucketMetric` interface.
  - In `aggregateMetricsOverTime`, computed `cumulative_net_lines_added` across sorted chronological buckets.
- **Pure Formatters (`web/src/domain/formatters.ts`)**:
  - Removed `'cumulative'` branch from `formatChartAxisTimestamp`.
  - Implemented `formatCompactCount` supporting magnitude abbreviations (`k`, `M`) and optional sign prefixes (`+` / `-`).
  - Implemented `formatLinesCompact` formatting discrete line counts without decimal fractions.

### Boundary Adapters Layer
- **Go HTTP Handler (`backend/internal/adapters/http/handlers.go`)**:
  - Updated `handleProgression` to reject non-temporal granularities (including `cumulative`) with HTTP 400 Bad Request and descriptive error message (`must be hourly, daily, or weekly`).
- **HTTP Server Routing & Adversarial Tests (`backend/internal/adapters/http/server_test.go`)**:
  - Added `TestProgressionHandler_AL34_Adversarial` testing rejection of `cumulative`, casing variations, whitespace, path traversal, and script payloads with HTTP 400.
  - Asserted valid granularities return HTTP 200 with matching `cumulative_net_lines_added`.
- **Domain Adversarial Tests (`backend/internal/domain/aggregations_test.go`)**:
  - Added `TestAL34_NetLinesGrowthAndGranularityAdversarial` covering running net code growth calculation, granularity rejection, and empty/zero-token boundary states.

### Presentation Layer
- **TrendChart Component (`web/src/components/TrendChart.tsx`)**:
  - Added `yAxisFormatter`, `integerTicks`, and `divergingZeroBaseline` props.
  - Implemented bi-directional Y scaling when negative values are present (`minY < 0`).
  - Rendered positive bars extending upward from zero baseline (`yZero = getYCoord(0)`) and negative bars extending downward.
  - Rendered horizontal zero baseline `<line className="chart-zero-baseline" ... />`.
  - Rounded Y-axis ticks to integers when `integerTicks` is enabled, and formatted tick labels using `yAxisFormatter`.
  - Contained labels inside left padding (`64px`) to prevent container overflow.
- **Metric Progression View (`web/src/views/MetricProgression.tsx`)**:
  - Removed "Cumulative" toggle button from Granularity controls.
  - Maintained static title "Financial Spend & Cumulative Trajectory ($ USD)" and dual series (period spend bar + cumulative spend line) for Chart 1 across all granularities.
  - Maintained composite series for Chart 2 (Token Volume Distribution).
  - Configured Chart 4 ("Delivered Code Line Churn & Net Growth") as a diverging bar chart with green positive additions, red negative deletions (`-Math.abs(lines_deleted)`), running cumulative net lines trajectory line, integer ticks, and compact signed formatting (`+30k`, `-10k`, `0`).
- **Component & View Tests (`web/src/components/TrendChart.test.tsx`, `web/src/views/MetricProgression.test.tsx`)**:
  - Authored comprehensive test suite for `TrendChart` verifying diverging bars, zero baseline position, negative bar coordinates, and tick formatting.
  - Updated `MetricProgression.test.tsx` verifying absence of Cumulative button and correct bi-directional churn rendering.

---

## Verification Status

- Go Backend Domain & HTTP Tests: `go test -race ./cmd/... ./internal/domain/... ./internal/adapters/http/...` passed cleanly.
- Go Vet: `go vet ./...` passed with zero warnings.
- TypeScript Tests: `vitest run` passed across 21 test files (412 tests).
- TypeScript Type Checking: `tsc --noEmit` passed with zero errors.
- ESLint: `eslint` passed with zero warnings.
- Frontend Production Build: Single binary bundle built and embedded into Go web assets via `embed.FS`.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 2h 50m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.2k tokens (84% of 200k healthy ceiling, 6 compactions) | Subagents peak 168.9k (Tester)
- **Role Effort Distribution (Subagents):** Planner 7.5%, Coder 57.1%, Tester 25.9%, Reviewer 9.5%
- **Subagent Turn Velocity & Thrashing:** 83.1 avg turns/run (peak: 285 turns) | Planner 62.0, Coder 81.2, Tester 144.0, Reviewer 53.0
- **Test-to-Production Code Ratio:** 5.5:1 (+803 test lines / +145 prod lines)
- **Code Delivered:** +976 / -142 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 180,712 reasoning tokens (60.9% of generated output)
- **Tool Health & Self-Correction:** 953 tool calls (9 self-corrections, 99.1% success rate)

### Anomalies & Health Flags
- **Status:** 10 execution anomalies detected (4 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 28 (session: ses_eaa6a1fe4ffeRETrbx7cZmcvfS), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.1k tokens) — Context tokens reached 160.1k in turn 79 (session: ses_eaa7dfb7effe8KcGclwVTrLvku), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.1k tokens) — Context tokens reached 161.1k in turn 89 (session: ses_eaa96a6b0ffeA5GEReuDcVsYzx), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.0k tokens) — Context tokens reached 160.0k in turn 114 (session: ses_eaaa516a0ffeGJNNcwC4miQtfo), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.6k tokens) — Context tokens surged by 61.6k tokens in turn 6 (session: ses_eaa6a1fe4ffeRETrbx7cZmcvfS, total: 81.1k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+61.7k tokens) — Context tokens surged by 61.7k tokens in turn 95 (session: ses_eaa7dfb7effe8KcGclwVTrLvku, total: 61.7k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+57.1k tokens) — Context tokens surged by 57.1k tokens in turn 100 (session: ses_eaa96a6b0ffeA5GEReuDcVsYzx, total: 57.1k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+75.5k tokens) — Context tokens surged by 75.5k tokens in turn 125 (session: ses_eaaa516a0ffeGJNNcwC4miQtfo, total: 75.5k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+51.3k tokens) — Context tokens surged by 51.3k tokens in turn 201 (session: ses_eaa96a6b0ffeA5GEReuDcVsYzx, total: 51.3k tokens, threshold: 40000)
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+48.5k tokens) — Context tokens surged by 48.5k tokens in turn 284 (session: ses_eaaa516a0ffeGJNNcwC4miQtfo, total: 48.5k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,501,487 | 45,714 | 26,894,315 | 91.5% | 168.2k | ELEVATED | $4.21 |
| **planner** | 1 | 306,062 | 8,808 | 2,268,837 | 88.1% | 81.0k | HEALTHY | $0.50 |
| **coder** | 5 | 2,055,616 | 40,153 | 24,249,736 | 92.2% | 168.6k | ELEVATED | $3.81 |
| **tester** | 1 | 1,094,070 | 17,008 | 9,530,395 | 89.7% | 168.9k | ELEVATED | $1.73 |
| **reviewer** | 1 | 356,264 | 4,354 | 4,239,730 | 92.2% | 168.6k | ELEVATED | $0.64 |
| **Total** | **9** | **6,313,499** | **116,037** | **67,183,013** | **91.4%** | **168.9k** | **ELEVATED** | **$10.89** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 9 | 6,313,499 | 116,037 | 67,183,013 | 91.4% | $10.89 |
| **Total** | **9** | **6,313,499** | **116,037** | **67,183,013** | **91.4%** | **$10.89** |
