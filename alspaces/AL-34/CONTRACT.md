# Contract: AL-34 Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

## 1. Problem Statement

In the AgenticLens Metric Progression view, the temporal analysis controls currently expose Cumulative as a granularity option alongside Hourly, Daily, and Weekly. In temporal domain modeling, granularity represents bucket duration, whereas cumulative is an aggregation trajectory. Treating Cumulative as a granularity breaks domain semantics, defaults under the hood to daily bucketing, and renders 4 out of 5 charts unresponsive.

In addition, Chart 4 (Delivered Code Line Churn & Net Growth) displays lines deleted as positive upward bars. This conceals the distinction between net codebase growth, code replacement, and refactoring velocity. Furthermore, Chart 4 Y-axis tick values format verbose fractional strings such as 33,306.56 lines, overflowing the chart container boundary and misrepresenting discrete line counts.

Ticket AL-34 eliminates cumulative granularity from both Go backend and TypeScript frontend, restricts temporal granularity strictly to hourly, daily, and weekly intervals, upgrades Chart 4 into a centered bi-directional diverging churn chart, introduces compact count formatting with integer tick intervals, and enforces layout containment preventing axis label overflow.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architectural changes must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Functional Core (backend/internal/domain/):
   - Pure domain models in models.go and enums in enums.go define TimeGranularity as an immutable string enum with values hourly, daily, and weekly.
   - AggregateMetricsOverTime in aggregations.go performs pure chronological bucketing over immutable slices of AgenticMetricsLedger.
   - Zero external dependencies: no imports of net/http, os, database/sql, or disk traversal packages.
   - Zero non-deterministic operations: no time.Now() calls or internal clocks. Timestamps are supplied as explicit arguments.
   - Pure mathematical calculations: cumulative rollups accrue monotonically across sorted buckets, and net line calculations remain deterministic.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Boundary HTTP handler handleProgression in backend/internal/adapters/http/handlers.go parses query parameters, rejects non-temporal granularities (including cumulative) with HTTP 400 Bad Request, and serializes domain results to JSON.
3. TypeScript Domain & Formatters (web/src/domain/):
   - Pure domain logic in web/src/domain/progression.ts mirrors Go domain bucketing, computing cumulative_net_lines_added sequentially.
   - Pure formatters in web/src/domain/formatters.ts implement compact integer count formatting without side effects.
4. TypeScript Presentation Shell (web/src/views/ and web/src/components/):
   - Pure presentation components in web/src/views/MetricProgression.tsx and web/src/components/TrendChart.tsx render reactive SVG visuals without direct API access.

### Strong Typing & Zero Any Invariant
1. Zero usage of any or interface{} across Go backend and TypeScript frontend.
2. TimeGranularity is strictly typed as an enum in Go and a string literal union in TypeScript.
3. Explicit type contracts govern all chart props, formatters, and dataset rows.

### Zero Emojis Mandate
No emojis permitted in Go code, TypeScript code, comments, docstrings, commit messages, PR drafts, UI labels, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Functional Core (backend/internal/domain/):
   - Remove GranularityCumulative constant from TimeGranularity in backend/internal/domain/enums.go.
   - Update ParseTimeGranularity in backend/internal/domain/enums.go to accept only hourly, daily, and weekly (case-insensitive), returning false for cumulative or invalid strings.
   - Add CumulativeNetLinesAdded field to TimeBucketMetric struct in backend/internal/domain/models.go with JSON tag cumulative_net_lines_added.
   - Update AggregateMetricsOverTime in backend/internal/domain/aggregations.go to calculate and populate CumulativeNetLinesAdded as running cumulative lines added minus running cumulative lines deleted.
   - Update aggregations_test.go to verify valid granularities and assert cumulative is rejected.
2. Go Boundary Adapters (backend/internal/adapters/):
   - Update handleProgression in backend/internal/adapters/http/handlers.go to return HTTP 400 Bad Request with an error message stating granularity must be hourly, daily, or weekly when validation fails.
   - Update server_test.go to verify that GET /api/progression?granularity=cumulative returns HTTP 400 Bad Request.
3. TypeScript Domain & Formatters (web/src/domain/):
   - Update TimeGranularity in web/src/domain/types.ts to 'hourly' | 'daily' | 'weekly'.
   - Add cumulative_net_lines_added to TimeBucketMetric interface in web/src/domain/types.ts.
   - Update aggregateMetricsOverTime in web/src/domain/progression.ts to compute cumulative_net_lines_added for each bucket.
   - Update getBucketKey in web/src/domain/progression.ts to clean up type checks.
   - Remove 'cumulative' branch in formatChartAxisTimestamp in web/src/domain/formatters.ts.
   - Implement formatCompactCount and formatLinesCompact in web/src/domain/formatters.ts supporting magnitude abbreviations (k, M) and sign prefixes (+, -).
   - Update unit tests in progression.test.ts and formatters.test.ts.
4. TypeScript Presentation (web/src/views/ and web/src/components/):
   - Remove the Cumulative button from the Granularity toggle group in web/src/views/MetricProgression.tsx.
   - Retain static title and composite period spend bars plus cumulative spend line for Chart 1 across all granularities.
   - Retain composite series for Chart 2 (Cache Read, Input, Output).
   - Upgrade TrendChart.tsx with yAxisFormatter prop, bi-directional diverging bar rendering centered on y = getYCoord(0) when minY < 0, integer tick rounding, and zero-baseline marker.
   - Upgrade Chart 4 in MetricProgression.tsx: positive green bars upward for lines added, negative red bars downward for lines deleted, line series for cumulative_net_lines_added, compact Y-axis tick formatting, and detailed signed tooltip formatting.
   - Create comprehensive unit test suite in TrendChart.test.tsx.
   - Update component tests in MetricProgression.test.tsx.

### Out of Scope
1. Altering OpenCode SQLite database schema, live event ingestion, or telemetry generation scripts.
2. Modifying ticket prefix filtering or ticket type classification.
3. Modifying Fleet Overview completion date filtering or autopsy sorting.

---

## 4. Architectural Contracts and Data Specifications

### Temporal Granularity Contract
1. Valid Granularities: hourly, daily, weekly.
2. Default Granularity: daily.
3. Rejection of Cumulative: Passing granularity=cumulative to /api/progression returns HTTP 400 Bad Request with error code bad_request and message specifying that granularity must be hourly, daily, or weekly.

### Code Churn Metrics Contract
1. Lines Added: Absolute positive integer count representing lines of code added in the bucket.
2. Lines Deleted: Rendered as a negative integer bar value (-Math.abs(lines_deleted)) in the diverging chart representation.
3. Net Lines Growth: Line series tracking cumulative_net_lines_added across the time window, where cumulative_net_lines_added equals cumulative_lines_added minus cumulative_lines_deleted.

### Compact Count Formatting Contract
1. formatCompactCount(count, options):
   - Values below 1,000 render as plain integers.
   - Values from 1,000 to 999,999 render with suffix k (e.g. 1k, 10k, 30k, 30.5k).
   - Values of 1,000,000 and above render with suffix M (e.g. 1M, 1.5M).
   - Positive values include a + prefix when explicitPlus option is true.
   - Negative values include a - prefix.
   - Zero renders as 0 without sign unless explicitPlus is true (which renders +0).
2. formatLinesCompact(lines, options):
   - Formats discrete line counts for chart axes and metric badges using formatCompactCount.
   - Guarantees integer representations for tick marks without decimal fractions.

### TrendChart Diverging Visualization Contract
1. Bi-Directional Zero Baseline:
   - When data contains negative values or minY is less than 0, yZero is computed via getYCoord(0).
   - Positive bars render upward from yZero to getYCoord(val) with height yZero - getYCoord(val).
   - Negative bars render downward from yZero to getYCoord(val) with height getYCoord(val) - yZero.
   - A distinct horizontal zero baseline renders at yZero to visually split additions and deletions.
2. Dual Formatting:
   - yAxisFormatter prop formats Y-axis tick values (compact notation, e.g. +30k, -10k, 0).
   - valueFormatter prop formats tooltip values with full precision (e.g. +33,307 lines, -12,410 lines).
3. Integer Ticks & Containment:
   - When integerTicks is true or discrete counts are passed, Y-axis ticks are rounded to integers.
   - Left padding and SVG text anchors prevent Y-axis tick labels from overflowing the card boundary.
