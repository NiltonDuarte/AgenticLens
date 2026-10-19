# AL-27 Summary: Metric Progression Custom Window and Hourly Granularity

## Overview
Implemented high-resolution temporal tracking across AgenticLens, adding a 1-day (24-hour) window preset, custom datetime selection with sub-day precision, and an hourly aggregation granularity (`YYYY-MM-DDTHH:00:00Z`). Changes span the Go functional core and HTTP adapters as well as the TypeScript domain, API client, and React UI components.

## Work Completed
1. Go Domain:
   - Added `TimeWindow1D` ("1d") and `TimeWindowCustom` ("custom") to `TimeWindow`.
   - Added `GranularityHourly` ("hourly") to `TimeGranularity`.
   - Added `TimeWindowLast24Hours` ("Last 24 Hours") to `TimeWindowPreset`.
   - Implemented 24-hour window preset calculation in `ResolveDatePreset`.
   - Implemented 24-hour window filtering in `FilterLedgersByTimeWindow`.
   - Extended `parseTimestamp` to parse `2006-01-02T15:04` layouts.
   - Enhanced `AggregateMetricsOverTime` to aggregate invocations and code metrics chronologically into hourly buckets formatted as `YYYY-MM-DDTHH:00:00Z`, preserving hour/minute precision in UTC.
2. Go Adapters & HTTP Handlers:
   - Extended `parseDateBoundary` in HTTP handlers to parse datetime strings formatted as `2006-01-02T15:04`, properly setting end-of-minute boundaries for inclusive filtering.
   - Updated `resolveTemporalFilter` to support the `1d` window query parameter and custom date boundaries.
   - Implemented `ProgressionHandler` in `backend/internal/adapters/http/handlers.go` and wired route `GET /api/progression` in `server.go`.
   - Added `WriteProgressionResponse` in `dto.go` enforcing generic response payload typing.
   - Added integration tests in `server_test.go` and unit tests in `aggregations_test.go`.
3. TypeScript Domain & API Client:
   - Updated `TimeGranularity` and `TimeWindowPreset` union types in `web/src/domain/types.ts`.
   - Enhanced `parseIsoDateToUtc` and `parseIsoBoundary` in `web/src/domain/progression.ts` to parse full ISO timestamps with time precision and validate dates defensively.
   - Updated `getBucketKey` and `aggregateMetricsOverTime` to support hourly buckets formatted as `YYYY-MM-DDTHH:00:00Z` and handle sub-day boundary filtering.
   - Exported pure formatter `formatChartAxisTimestamp` in `web/src/domain/formatters.ts`.
   - Implemented `fetchProgressionMetrics` in `web/src/api/client.ts` with local pure fallback.
4. React UI & CSS:
   - Added `24 Hours` ('1d') and `Custom` ('custom') toggle buttons in `MetricProgression.tsx`.
   - Added `Hourly` ('hourly') granularity toggle button in `MetricProgression.tsx`.
   - Added custom datetime input controls (`type="datetime-local"`) rendered when `Custom` window is active, with validation guarding against start > end.
   - Integrated `granularity={granularity}` into `TrendChart` instances.
   - Added CSS styles for `.custom-datetime-bar`, `.datetime-input`, and `.datetime-warning`.
   - Maintained zero emojis across all code, docstrings, UI labels, buttons, and tests.

## Verification
- Backend tests: `go test -count=1 -race ./...` (All 10 packages passed).
- Frontend typecheck: `npm run typecheck` (Clean, zero errors).
- Frontend lint: `npm run lint` (Clean, zero warnings).
- Frontend tests: `npm run test` (19 test files, 358 tests passed).
- Production build: `make build` (Successfully bundled Vite production assets and compiled Linux/macOS multi-platform Go binaries).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 1
- **Active Agent Duration:** 2h 47m (excludes user wait time)
- **Context Window Health:** ELEVATED — Orchestrator peak 168.1k tokens (84% of 200k healthy ceiling, 3 compactions) | Subagents peak 167.1k (Coder)
- **Role Effort Distribution (Subagents):** Planner 10.0%, Coder 59.0%, Tester 15.8%, Reviewer 15.2%
- **Subagent Turn Velocity & Thrashing:** 116.8 avg turns/run (peak: 244 turns) | Planner 59.0, Coder 244.0, Tester 76.0, Reviewer 88.0
- **Test-to-Production Code Ratio:** 2.2:1 (+1,260 test lines / +568 prod lines)
- **Code Delivered:** +1,828 / -165 lines across 24 code files (excluding docs)
- **Reasoning Effort:** 138,111 reasoning tokens (54.6% of generated output)
- **Tool Health & Self-Correction:** 618 tool calls (5 self-corrections, 99.2% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (3 CRITICAL, 1 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 31 (session: ses_eaba3d891ffeGWFDD5HOZkKLft), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.8k tokens) — Context tokens reached 160.8k in turn 58 (session: ses_eabcd9f47ffeLS0GWyjygcTYQF), breaching 160k critical ceiling
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.5k tokens) — Context tokens reached 161.5k in turn 88 (session: ses_eabf926d0ffeOXK3XrdCjUZ03N), breaching 160k critical ceiling
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+67.4k tokens) — Context tokens surged by 67.4k tokens in turn 98 (session: ses_eabf926d0ffeOXK3XrdCjUZ03N, total: 67.4k tokens, threshold: 40000)

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,009,248 | 30,568 | 9,020,707 | 81.8% | 168.1k | ELEVATED | $2.38 |
| **planner** | 1 | 263,501 | 18,433 | 3,583,563 | 93.2% | 122.5k | ELEVATED | $0.63 |
| **coder** | 1 | 2,109,550 | 41,896 | 23,071,593 | 91.6% | 167.1k | ELEVATED | $3.69 |
| **tester** | 1 | 587,991 | 17,407 | 5,517,950 | 90.4% | 114.9k | ELEVATED | $0.99 |
| **reviewer** | 1 | 616,888 | 6,617 | 5,426,813 | 89.8% | 165.4k | ELEVATED | $0.95 |
| **Total** | **5** | **5,587,178** | **114,921** | **46,620,626** | **89.3%** | **168.1k** | **ELEVATED** | **$8.64** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 5 | 5,587,178 | 114,921 | 46,620,626 | 89.3% | $8.64 |
| **Total** | **5** | **5,587,178** | **114,921** | **46,620,626** | **89.3%** | **$8.64** |
