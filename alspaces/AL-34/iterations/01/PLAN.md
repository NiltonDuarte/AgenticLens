# Implementation Plan: AL-34 Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

## 1. Goal and Iteration Scope

Deliver complete implementation of ticket AL-34 across Go functional core, Go HTTP boundary adapters, TypeScript domain layer, and TypeScript presentation components:
- Remove GranularityCumulative from Go domain enums and update ParseTimeGranularity.
- Update Go HTTP progression handler to validate hourly, daily, and weekly, rejecting cumulative with HTTP 400.
- Update Go domain models and aggregations to calculate and expose CumulativeNetLinesAdded.
- Update TypeScript TimeGranularity type to strictly 'hourly' | 'daily' | 'weekly'.
- Extend TypeScript TimeBucketMetric and aggregateMetricsOverTime to derive cumulative_net_lines_added.
- Update TypeScript formatters to remove cumulative timestamp formatting and add compact count formatters (formatCompactCount, formatLinesCompact) with sign prefixes.
- Update TrendChart component to support bi-directional diverging bars around a zero baseline, dedicated yAxisFormatter, integer tick rounding, and overflow containment.
- Update MetricProgression view to remove Cumulative toggle, retain Chart 1 static title and composite spend series, retain Chart 2 token series, and upgrade Chart 4 into a bi-directional diverging code churn visualization.
- Ensure strict static typing (zero any), zero emojis, and complete test suites across Go and TypeScript.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against concrete existing files and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/enums.go:
  - Remove GranularityCumulative constant.
  - Update ParseTimeGranularity to accept only hourly, daily, and weekly.
- backend/internal/domain/models.go:
  - Add CumulativeNetLinesAdded int64 to TimeBucketMetric struct with JSON tag cumulative_net_lines_added.
- backend/internal/domain/aggregations.go:
  - Update AggregateMetricsOverTime docstring to state hourly, daily, or weekly buckets.
  - Calculate CumulativeNetLinesAdded as runningCumLinesAdded minus runningCumLinesDeleted.
- backend/internal/domain/aggregations_test.go:
  - In TestParseTimeGranularity, update test case for cumulative to assert valid is false and expected is empty.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/http/handlers.go:
  - In handleProgression, update error message when ParseTimeGranularity fails to state that granularity must be hourly, daily, or weekly.
- backend/internal/adapters/http/server_test.go:
  - Add test case verifying GET /api/progression?granularity=cumulative returns HTTP 400 Bad Request with error code bad_request and descriptive message.
  - Add cumulative to invalidGranularities slice.

### TypeScript Domain & Pure Formatters (web/src/domain/)
- web/src/domain/types.ts:
  - Update TimeGranularity type definition to 'hourly' | 'daily' | 'weekly'.
  - Add readonly cumulative_net_lines_added: number to TimeBucketMetric interface.
- web/src/domain/progression.ts:
  - In getBucketKey, remove any lingering cumulative assumptions.
  - In aggregateMetricsOverTime, compute and attach cumulative_net_lines_added: runningCumLinesAdded - runningCumLinesDeleted.
- web/src/domain/progression.test.ts:
  - Update test cases passing 'cumulative' as granularity to valid temporal granularities.
  - Add assertions verifying cumulative_net_lines_added across chronological buckets.
- web/src/domain/formatters.ts:
  - In formatChartAxisTimestamp, remove cumulative references.
  - Implement formatCompactCount(count: number, options?: { showSign?: boolean; explicitPlus?: boolean }): string.
  - Implement formatLinesCompact(lines: number, options?: { explicitPlus?: boolean }): string.
- web/src/domain/formatters.test.ts:
  - Remove cumulative test case from formatChartAxisTimestamp.
  - Add unit tests for formatCompactCount and formatLinesCompact covering positive, negative, zero, thousands, millions, and sign prefixes.

### TypeScript Presentation (web/src/components/ & web/src/views/)
- web/src/components/TrendChart.tsx:
  - Extend TrendChartProps with yAxisFormatter?: (val: number) => string and integerTicks?: boolean.
  - Update Y scale calculation to detect negative values in series data and reference lines, setting minY below zero when minDataValue < 0.
  - Compute yZero = getYCoord(0).
  - In bar rendering, calculate bar rectangle from yZero: positive bars extend upward from yZero to getYCoord(val); negative bars extend downward from yZero to getYCoord(val).
  - Render horizontal zero baseline line when minY < 0.
  - In horizontal grid line rendering, use yAxisFormatter when provided, and round tick values when integerTicks is true.
  - Verify layout padding and text alignment prevent tick labels from overflowing the card.
- web/src/components/TrendChart.test.tsx:
  - Create new unit test suite testing bi-directional diverging bars, yAxisFormatter usage, integer tick intervals, and zero baseline rendering.
- web/src/views/MetricProgression.tsx:
  - Remove Cumulative button from granularity-toggle-group.
  - In chartData mapping, include lines_deleted: -Math.abs(b.lines_deleted) and cumulative_net_lines_added.
  - For Chart 1 (Financial Spend & Cumulative Trajectory), keep title static and retain period spend bar plus cumulative spend line series across all active granularities.
  - For Chart 2 (Token Volume Distribution), retain composite token series.
  - For Chart 4 (Delivered Code Line Churn & Net Growth), configure lines_added as positive bar, lines_deleted as negative bar, cumulative_net_lines_added as net lines growth line, yAxisFormatter using formatLinesCompact, and detailed tooltip value formatting.
- web/src/views/MetricProgression.test.tsx:
  - Update granularity toggle tests: assert Hourly, Daily, Weekly buttons exist, assert Cumulative button does not exist.
  - Assert Chart 1 maintains static title across granularity switches.
  - Assert Chart 4 renders bi-directional bars and cumulative net growth line.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core & HTTP Boundary Adapter
- Task 1.1: Refactor domain enums in backend/internal/domain/enums.go:
  - Remove GranularityCumulative constant from TimeGranularity.
  - In ParseTimeGranularity, remove case string(GranularityCumulative), returning empty string and false.
- Task 1.2: Update domain models and aggregations in backend/internal/domain/models.go and backend/internal/domain/aggregations.go:
  - In backend/internal/domain/models.go, add CumulativeNetLinesAdded int64 with JSON tag cumulative_net_lines_added to TimeBucketMetric.
  - In backend/internal/domain/aggregations.go, update AggregateMetricsOverTime docstring to state hourly, daily, or weekly bucketing.
  - In AggregateMetricsOverTime, populate CumulativeNetLinesAdded: runningCumLinesAdded - runningCumLinesDeleted in the TimeBucketMetric instantiation.
- Task 1.3: Update HTTP progression handler in backend/internal/adapters/http/handlers.go:
  - In handleProgression, update the error response for invalid granularity to: invalid granularity "%s": must be hourly, daily, or weekly.
- Task 1.4: Update backend tests in backend/internal/domain/aggregations_test.go and backend/internal/adapters/http/server_test.go:
  - In backend/internal/domain/aggregations_test.go, update TestParseTimeGranularity to verify that cumulative input returns valid: false and expected: "".
  - In backend/internal/adapters/http/server_test.go, add a test case verifying GET /api/progression?granularity=cumulative returns HTTP 400 Bad Request with error code bad_request and proper error message.
  - Run go test ./... to confirm Go tests pass.

### Phase 2: TypeScript Domain Layer & Pure Formatters
- Task 2.1: Update TypeScript types in web/src/domain/types.ts:
  - Change TimeGranularity to 'hourly' | 'daily' | 'weekly'.
  - Add readonly cumulative_net_lines_added: number to TimeBucketMetric.
- Task 2.2: Update progression aggregation in web/src/domain/progression.ts:
  - In getBucketKey, ensure clean handling of TimeGranularity ('hourly', 'weekly', or daily fallback).
  - In aggregateMetricsOverTime, assign cumulative_net_lines_added: runningCumLinesAdded - runningCumLinesDeleted to each TimeBucketMetric.
- Task 2.3: Implement compact count formatters in web/src/domain/formatters.ts:
  - In formatChartAxisTimestamp, remove any handling for 'cumulative'.
  - Implement formatCompactCount(count: number, options?: { showSign?: boolean; explicitPlus?: boolean }): string:
    - Return '0' or '+0' when count is 0.
    - Format thousands with 'k' (e.g. 1k, 10k, 30.5k) and millions with 'M' (e.g. 1M, 1.5M).
    - Attach '-' for negative numbers and '+' when explicitPlus is true for positive numbers.
  - Implement formatLinesCompact(lines: number, options?: { explicitPlus?: boolean }): string:
    - Format integer lines compactly using formatCompactCount with integer rounding.
- Task 2.4: Update TypeScript domain tests in web/src/domain/progression.test.ts and web/src/domain/formatters.test.ts:
  - In web/src/domain/progression.test.ts, update tests that used 'cumulative' to use 'daily', and add assertions checking cumulative_net_lines_added.
  - In web/src/domain/formatters.test.ts, remove 'cumulative' test from formatChartAxisTimestamp, and add test suite for formatCompactCount and formatLinesCompact.
  - Run npm test to verify domain test suites pass.

### Phase 3: TrendChart SVG Component Upgrades
- Task 3.1: Update TrendChart props in web/src/components/TrendChart.tsx:
  - Add yAxisFormatter?: (val: number) => string.
  - Add integerTicks?: boolean.
- Task 3.2: Implement bi-directional Y-axis scaling:
  - Iterate through data rows and series to find minDataValue and maxDataValue.
  - Determine minY: if yAxisMin is provided and minDataValue >= 0, use yAxisMin. If minDataValue < 0, set minY to minDataValue * 1.12 (or min(yAxisMin ?? 0, minDataValue * 1.12)).
  - Determine maxY: maxDataValue > 0 ? maxDataValue * 1.12 : 10.
  - Ensure yRange handles negative to positive span cleanly.
- Task 3.3: Implement bi-directional bar rendering and zero baseline:
  - Calculate yZero = getYCoord(0).
  - When rendering bars:
    - If val >= 0: y coordinate is getYCoord(val) and height is Math.max(0, yZero - getYCoord(val)).
    - If val < 0: y coordinate is yZero and height is Math.max(0, getYCoord(val) - yZero).
  - If minY < 0, render a horizontal zero baseline line across the chart width at yZero with subtle stroke styling.
  - For area fills, anchor bottom to yZero instead of chart bottom.
- Task 3.4: Implement integer tick formatting and label containment:
  - In Y-axis tick generation, if integerTicks is true, round each tick value to the nearest integer.
  - Use yAxisFormatter for tick labels when provided; fallback to valueFormatter.
  - Ensure padding.left (64px) and right (28px) comfortably contain compact labels (e.g. +30k, -10k, 0) without overflowing the SVG boundary.
- Task 3.5: Create TrendChart unit test suite in web/src/components/TrendChart.test.tsx:
  - Test bi-directional bar rendering with positive and negative data values.
  - Test custom yAxisFormatter rendering on Y-axis labels.
  - Test integer tick generation without decimal fractions.
  - Test zero baseline element rendering when data contains negative values.

### Phase 4: Metric Progression Presentation View Integration
- Task 4.1: Update granularity toggle in web/src/views/MetricProgression.tsx:
  - Remove the Cumulative button from the granularity toggle group.
  - Retain Hourly, Daily, Weekly buttons.
- Task 4.2: Update Chart 1 and Chart 2:
  - Set Chart 1 title statically to "Financial Spend & Cumulative Trajectory ($ USD)".
  - Configure Chart 1 series with period spend bar and cumulative spend line across all active granularities.
  - Retain Chart 2 composite token distribution bars.
- Task 4.3: Update Chart 4 (Delivered Code Line Churn & Net Growth):
  - In chartData, map lines_added as positive integer, lines_deleted as negative integer (-Math.abs(b.lines_deleted)), and cumulative_net_lines_added as cumulative net lines.
  - In Chart 4 series configuration:
    - Lines Added (+): key lines_added, color CHART_COLORS.linesAdded, type bar.
    - Lines Deleted (-): key lines_deleted, color CHART_COLORS.linesDeleted, type bar.
    - Net Lines Growth: key cumulative_net_lines_added, color CHART_COLORS.linesNet, type line.
  - Provide yAxisFormatter={(v) => formatLinesCompact(v, { explicitPlus: true })} to display compact integer ticks (+30k, -10k, 0).
  - Pass integerTicks={true}.
  - Provide valueFormatter formatted with sign prefixes and lines suffix for tooltip display.
- Task 4.4: Update MetricProgression test suite in web/src/views/MetricProgression.test.tsx:
  - Update granularity toggle tests to verify only Hourly, Daily, Weekly buttons are present and assert Cumulative button is absent.
  - Verify switching between Hourly, Daily, Weekly preserves Chart 1 composite series and static title.
  - Verify Chart 4 renders bi-directional bars and cumulative net growth line.

### Phase 5: Verification & Quality Gates
- Task 5.1: Run Go backend verification:
  - Run go test -race ./...
  - Run go vet ./...
- Task 5.2: Run TypeScript frontend verification:
  - Run npm test
  - Run npm run typecheck
  - Run npm run lint
- Task 5.3: Run full single-binary build:
  - Run make build or npm run build followed by go build ./cmd/agentic-lens.
- Task 5.4: Verify zero any and zero emojis across all modified files.
