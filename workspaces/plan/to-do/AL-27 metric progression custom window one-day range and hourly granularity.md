h3. Title: AL-27: Metric Progression Custom Window, 1-Day Window, and Hourly Granularity

*TARGET*: AgenticLens (domain | adapters | presentation | tests)

*CONTEXT*

The Metric Progression temporal analysis view currently supports limited preset windows (`all`, `7d`, `30d`, `90d`) and fixed granularities (`daily`, `weekly`, `cumulative`). Developers investigating acute cost spikes, short-lived runaways, or active debugging sessions require high-resolution temporal tracking: a 1-day (24-hour) window option, an arbitrary custom date range picker, and hourly aggregation granularity, while preserving defaults to lifetime window and daily granularity.

*GOAL*

Extend the Metric Progression aggregation engine and UI with custom window date pickers, a 1-day preset window, and hourly temporal granularity while maintaining backward compatibility and defaults.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Add `GranularityHourly` (`"hourly"`) to `TimeGranularity` enum in Go `internal/domain`.
  - Update `AggregateMetricsOverTime` to support hourly bucketing (`YYYY-MM-DDTHH:00:00Z`).
  - Add `TimeWindow1D` (`"1d"`) and custom range parsing to domain window utilities.
* Boundary Adapter:
  - Update `/api/progression` HTTP handler to accept `window=1d`, `window=custom`, `since`, `until`, and `granularity=hourly`.
* Presentation:
  - Add `hourly` option to granularity selector in `MetricProgression.tsx` (`Hourly`, `Daily` [default], `Weekly`, `Cumulative`).
  - Add `1d` option (`Last 24 Hours`) and `custom` option (`Custom Range`) to Window dropdown.
  - Implement date-time picker inputs in `MetricProgression.tsx` that appear when `custom` is selected.
  - Update `TrendChart.tsx` and formatters to render hourly timestamps cleanly on chart axes.
* Quality:
  - Pure Go domain tests in `aggregations_test.go` verifying hourly bucketing, timestamp boundaries, and empty bucket handling.
  - TypeScript unit tests in `MetricProgression.test.tsx` and `progression.test.ts` testing 1-day windows, custom ranges, and hourly charts.

*ACCEPTANCE CRITERIA*

* Metric Progression defaults to Project Lifetime window and Daily granularity.
* Selecting `1d` window restricts telemetry to the past 24 hours.
* Selecting `Custom Range` reveals date inputs and restricts progression data to specified interval.
* Selecting `Hourly` granularity aggregates tokens, cost, invocations, and code churn into 1-hour time buckets.
* All charts render without layout distortion or tick label overlap at hourly resolution.
* `make test-backend` and `npm test` pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
