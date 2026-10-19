h3. Title: AL-34: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

*TARGET*: AgenticLens (domain | adapters | presentation | tests)

*CONTEXT*

The Metric Progression temporal analysis view currently exposes "Cumulative" as a time granularity alongside Hourly, Daily, and Weekly. Granularity represents bucket duration, while cumulative is an accumulation method. Presenting "Cumulative" as a granularity breaks domain semantics, silently defaults to daily bucketing, and leaves 4 out of 5 charts on the page unresponsive. Furthermore, Chart 4 ("Delivered Code Line Churn & Net Growth") currently renders lines deleted as upward-pointing positive bars, obscuring the distinction between net codebase growth, code replacement, and refactoring velocity. In addition, Chart 4's Y-axis currently formats tick values with verbose fractional strings (e.g. `33,306.56 lines`), which overflow the chart container boundary and make no mathematical sense for discrete line counts.

*GOAL*

Refactor the Metric Progression granularity model to purely temporal intervals (`hourly`, `daily`, `weekly`), maintain composite views for spend and token distributions, enhance the code churn chart into a bi-directional diverging visualization centered on a zero baseline with upward additions (green), downward deletions (red), and a running cumulative net code growth trajectory line, and resolve Y-axis label container overflow by introducing compact integer metric formatting (`k`, `M`) and integer tick rounding.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Remove `GranularityCumulative` from `TimeGranularity` enum in Go `backend/internal/domain/enums.go` and update `ParseTimeGranularity`.
  - Update `backend/internal/domain/aggregations.go` to support purely temporal bucketing (`hourly`, `daily`, `weekly`).
  - Update TypeScript `TimeGranularity` type in `web/src/domain/types.ts` to `'hourly' | 'daily' | 'weekly'`.
  - Compute `cumulative_net_lines_added` (`cumulative_lines_added - cumulative_lines_deleted`) in `TimeBucketMetric` in `web/src/domain/progression.ts`.
  - Update `web/src/domain/formatters.ts` to remove `'cumulative'` timestamp formatting, and implement compact integer count formatting (`formatLinesCompact` / `formatCompactCount`) supporting `k`, `M`, and sign prefixes (`+` / `-`).
* Boundary Adapter:
  - Update `handleProgression` in `backend/internal/adapters/http/handlers.go` to reject non-temporal granularities with HTTP 400 (`must be hourly, daily, or weekly`).
  - Update Go HTTP routing tests in `server_test.go` and domain tests in `aggregations_test.go`.
* Presentation:
  - Remove the "Cumulative" toggle button from `MetricProgression.tsx` controls bar.
  - Keep Chart 1 ("Financial Spend & Cumulative Trajectory") composite across all granularities with period spend bars and cumulative spend line; title remains static.
  - Keep Chart 2 ("Token Volume Distribution") composite across input, output, and cache read tokens.
  - Upgrade `TrendChart.tsx` to support:
    - Bi-directional diverging bar charts centered on a zero baseline when data contains negative values (e.g. `minY < 0`), rendering positive bars upward and negative bars downward from `y = getYCoord(0)`.
    - Dedicated `yAxisFormatter?: (val: number) => string` prop so axes can use compact notation while tooltips retain full precision (`valueFormatter`).
    - Integer rounding on Y-axis tick intervals for discrete count metrics, eliminating fractional ticks (e.g. `33,306.56`).
    - Layout containment preventing axis labels from overflowing the chart card boundary.
  - Update Chart 4 ("Delivered Code Line Churn & Net Growth") in `MetricProgression.tsx`:
    - Lines Added: positive green bars growing upward from zero baseline.
    - Lines Deleted: negative red bars growing downward from zero baseline.
    - Net Lines Growth: line tracing `cumulative_net_lines_added` across the time window.
    - Use `formatLinesCompact` for Y-axis tick labels and precise formatting in tooltips.
    - Update tooltip to format positive additions (`+N lines`), negative deletions (`-N lines`), and cumulative net lines (`+N / -N lines`) clearly.
* Quality:
  - Go unit tests in `aggregations_test.go` verifying valid granularities and rejecting removed cumulative granularity.
  - Go HTTP tests in `server_test.go` ensuring `granularity=cumulative` returns HTTP 400.
  - TypeScript unit tests in `progression.test.ts`, `formatters.test.ts`, `MetricProgression.test.tsx`, and `TrendChart.test.tsx` testing the removed toggle, composite spend chart, bi-directional churn chart rendering, compact axis scaling, and overflow prevention.

*ACCEPTANCE CRITERIA*

* Metric Progression controls bar displays only `Hourly`, `Daily`, and `Weekly` granularities, defaulting to `Daily`.
* Passing `granularity=cumulative` to `/api/progression` returns HTTP 400 with a descriptive error.
* Chart 1 ("Financial Spend & Cumulative Trajectory ($ USD)") renders period spend bars and cumulative spend trajectory across all active granularities.
* Chart 2 ("Token Volume Distribution") continues to render composite token series (Cache Read, Input, Output).
* Chart 4 ("Delivered Code Line Churn & Net Growth") renders a centered zero baseline:
  - Lines added bars grow upward in green (+).
  - Lines deleted bars grow downward in red (-).
  - Net code growth line tracks running cumulative net lines (`cumulative_lines_added - cumulative_lines_deleted`).
* Chart 4 Y-axis displays compact integer tick marks using `k` and `M` (e.g. `+30k`, `-10k`, `0`), strictly eliminating fractional numbers (e.g. `33,306.56 lines`).
* Chart 4 axis labels remain strictly within the chart card margins without overflowing the bounding box.
* Tooltip on Chart 4 clearly displays added lines (+), deleted lines (-), and cumulative net growth with exact integer formatting.
* `TrendChart` SVG handles both positive and negative values cleanly without axis or label distortion.
* `make test-backend` and `npm test` pass cleanly with zero failures.
* Zero `any` in Go and TypeScript. Zero emojis in code, tests, and UI strings.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
