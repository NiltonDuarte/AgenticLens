# Contract: AL-27 Metric Progression Custom Window, 1-Day Window, and Hourly Granularity

## 1. Problem Statement

Across AgenticLens, the Metric Progression temporal view currently supports limited window presets (all, 7d, 30d, 90d) and fixed granularities (daily, weekly, cumulative). In the Go functional core (backend/internal/domain/aggregations.go), timestamps parsed from invocations are truncated to calendar dates (YYYY-MM-DD), removing intra-day time data.

When investigating sudden token spikes, runaway subagent rework loops, or active debugging sessions, daily bucketing aggregates too coarsely to pinpoint the exact hours when anomalies began. Developers require high-resolution temporal tracking: a 1-day (24-hour) window option, an arbitrary custom date-time range picker, and hourly aggregation granularity (YYYY-MM-DDTHH:00:00Z), while preserving defaults to project lifetime window and daily granularity.

Furthermore, the HTTP adapter layer lacks a dedicated progression endpoint (/api/progression), and existing endpoints (/api/workspaces, /api/fleet) do not handle 1-day windows or custom range parameters. In the frontend, TrendChart.tsx assumes date-only string slices for X-axis tick labels, leading to label collision and overflow when rendering hourly ISO timestamps.

Ticket AL-27 resolves these limitations across the Go domain engine, HTTP boundary adapters, and TypeScript presentation view while maintaining strict Functional Core, Imperative Shell (FCIS) separation and zero-any typing.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architectural changes must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Domain Functional Core (backend/internal/domain/):
   - Pure aggregation function AggregateMetricsOverTime in aggregations.go accepts a slice of AgenticMetricsLedger, a TimeGranularity enum, and optional start and end boundary timestamps.
   - When hourly granularity is requested, timestamps are grouped into UTC hour buckets formatted as YYYY-MM-DDTHH:00:00Z.
   - Invocations and deliverable code statistics are aggregated without mutating input structs.
   - Absolute isolation: zero imports of net/http, os, io, or disk scanning packages.
   - Zero non-deterministic clock calls: functions must never call time.Now(). Reference timestamps must be passed as explicit arguments.
   - Defensive mathematical calculations: prompt cache hit rates guard against zero token totals, returning 0.0 without divide-by-zero panics. Cumulative sums accrue chronologically over sorted bucket keys.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Boundary HTTP handlers ProgressionHandler, FleetHandler, and WorkspacesHandler in backend/internal/adapters/http/handlers.go parse and validate query parameters (window, granularity, since, until).
   - The shell handles wall-clock evaluation (time.Now().UTC()), converting relative window presets (such as 1d, 7d, 30d, 90d) into concrete UTC boundary timestamps before invoking pure domain functions.
   - HTTP routing and response serialization are wired in server.go and dto.go.
3. TypeScript Presentation and Formatters (web/src/domain/ and web/src/):
   - Pure domain functions in web/src/domain/progression.ts mirror Go domain bucketing with UTC timestamp parsing and ISO hour formatting.
   - Pure formatting functions in web/src/domain/formatters.ts convert hourly ISO strings into clean chart tick representations without layout collisions.
   - web/src/views/MetricProgression.tsx manages reactive filter state, preserving defaults to Project Lifetime and Daily granularity.
   - web/src/components/TrendChart.tsx renders high-density time series across SVG viewports with dynamic tick spacing.

### Strong Typing and Zero Any Invariant
1. Zero usage of any or interface{} across Go backend and TypeScript frontend.
2. All time granularities and window presets are strictly modeled as typed enums and literal string unions (TimeGranularity, TimeWindow, TimeWindowPreset).
3. Concrete DTO interfaces in Go dto.go enforce typed response structures via Go generics.

### Zero Emojis Mandate
No emojis are permitted in code, comments, docstrings, commit messages, PR drafts, UI labels, select options, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Functional Core (backend/internal/domain/):
   - Add GranularityHourly constant to TimeGranularity enum in enums.go.
   - Add TimeWindow1D and TimeWindowCustom constants to TimeWindow enum in enums.go.
   - Add TimeWindowLast24Hours constant to TimeWindowPreset in enums.go.
   - Update ParseTimeWindow to map string inputs 1d and custom.
   - Refactor timestamp parsing in aggregations.go from date-only truncation to full ISO 8601 and RFC 3339 resolution preserving hours and minutes.
   - Extend AggregateMetricsOverTime in aggregations.go to construct hourly buckets in format YYYY-MM-DDTHH:00:00Z.
   - Extend FilterLedgersByTimeWindow and ResolveDatePreset to resolve 24-hour windows relative to reference timestamp.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Update resolveTemporalFilter in handlers.go to handle 1d (evaluating start timestamp to 24 hours prior to current UTC time) and custom.
   - Extend parseDateBoundary in handlers.go to support ISO date-time inputs without seconds (such as HTML datetime-local format YYYY-MM-DDTHH:mm).
   - Implement ProgressionHandler in handlers.go accepting window, granularity, since, and until query parameters, delegating to domain.AggregateMetricsOverTime.
   - Register route GET /api/progression in server.go.
   - Extend ResponsePayload in dto.go to include slice of domain.TimeBucketMetric and implement WriteProgressionResponse.
3. TypeScript Domain and API (web/src/domain/ and web/src/api/):
   - Update TimeGranularity type in types.ts to include literal union member hourly.
   - Update TimeWindowPreset type in types.ts to include literal union members 1d and custom.
   - Update parseIsoDateToUtc in progression.ts to parse full ISO timestamps without discarding hour/minute precision.
   - Update getBucketKey in progression.ts to output YYYY-MM-DDTHH:00:00Z for hourly granularity.
   - Update aggregateMetricsOverTime in progression.ts to support hourly bucketing and date-time boundary filtering.
   - Implement formatChartAxisTimestamp in formatters.ts to transform ISO timestamps into concise tick labels (MM-DD HH:00 or HH:00).
   - Implement fetchProgressionMetrics in client.ts calling /api/progression, with client-side fallback calculation when running in mock mode.
4. TypeScript Presentation Views and Components (web/src/):
   - In MetricProgression.tsx, add Hourly toggle button to Granularity controls.
   - In MetricProgression.tsx, add 24 Hours (1d) and Custom Range (custom) toggle options to Window controls.
   - In MetricProgression.tsx, conditionally render date-time input controls for Since and Until when Custom Range is selected.
   - Update TrendChart.tsx to handle hourly bucket labels cleanly via dynamic formatting, preventing tick label overlap and SVG clipping.
5. Quality and Adversarial Testing:
   - Pure Go domain unit tests in aggregations_test.go verifying hourly bucketing, timestamp boundaries, and empty bucket handling.
   - Go HTTP adapter tests in server_test.go verifying /api/progression with window=1d, granularity=hourly, custom date ranges, and invalid query inputs.
   - TypeScript unit tests in progression.test.ts verifying hourly aggregation and boundary conditions.
   - TypeScript component tests in MetricProgression.test.tsx verifying default states, 1-day filtering, custom range input visibility, and zero emojis.

### Out of Scope
1. Altering OpenCode SQLite database schemas or raw telemetry collection logic.
2. Modifying ticket type classification or workspace prefix filtering (governed by AL-28 and AL-29).
3. Modifying Fleet Overview completion date filtering or autopsy sorting (governed by AL-26).
4. Permanent deletion or on-disk mutation of existing workspace ledgers.

---

## 4. Architectural Contracts and Data Specifications

### Hourly Bucketing Contract
1. Bucket Key Format: YYYY-MM-DDTHH:00:00Z.
2. Bucket Date Value: Identical to Bucket Key.
3. Hour Boundary Assignment:
   - All invocations and code events with effective timestamps between hour T at minute 00 second 00 through minute 59 second 59.999999999 UTC belong to the bucket for hour T.
   - For example, an invocation at 2026-10-19T14:23:45Z belongs to bucket 2026-10-19T14:00:00Z.
4. Empty Bucket Invariant: Time buckets with zero matching invocations and zero code churn are omitted from the returned chronological bucket array.
5. Chronological Order: Returned buckets are sorted in ascending chronological order by bucket key.
6. Cumulative Metric Monotonicity: Cumulative metrics (cumulative_cost_usd, cumulative_input_tokens, cumulative_output_tokens, cumulative_lines_added, cumulative_lines_deleted, cumulative_invocations) increase or remain constant across chronological buckets.

### Window Resolution Contract
1. Preset all (Project Lifetime): Evaluates start and end temporal bounds to nil, aggregating all available ledgers.
2. Preset 1d (Last 24 Hours): Evaluates start bound to exactly 24 hours prior to reference anchor timestamp, and end bound to reference anchor timestamp. In the HTTP boundary shell, reference anchor is time.Now().UTC(). In presentation view with static ledgers, reference anchor is latest recorded ledger timestamp.
3. Presets 7d, 30d, 90d: Evaluates start bound to 7, 30, or 90 calendar days prior to reference anchor.
4. Preset custom: Evaluates start bound to user-supplied since parameter and end bound to user-supplied until parameter. If since is supplied without until, window is open-ended until infinity. If until is supplied without since, window starts from the earliest recorded event.

### API Endpoint Query Parameter Contracts
1. GET /api/progression:
   - window: Optional string (all, 1d, 7d, 30d, 90d, custom). Defaults to all.
   - granularity: Optional string (hourly, daily, weekly, cumulative). Defaults to daily.
   - since: Optional ISO 8601 date or date-time string.
   - until: Optional ISO 8601 date or date-time string.
   - Status 200: Returns JSON array of TimeBucketMetric objects matching requested window and granularity.
   - Status 400: Returns structured ErrorResponse if date formatting is invalid or granularity is unrecognized.
2. GET /api/workspaces and GET /api/fleet:
   - Support window=1d and window=custom in addition to existing presets.

### Chart Axis Rendering Contract
1. Hourly Tick Formatting:
   - When bucket label is an hourly timestamp (YYYY-MM-DDTHH:00:00Z), the display tick renders as MM-DD HH:00 (or HH:00 when the visible window is 24 hours or less).
   - Tick label string length must not exceed 11 characters, ensuring clean spacing without rotation or overlap.
2. Step Calculation:
   - X-axis tick step dynamically samples points so that no more than 7 tick labels render simultaneously across the SVG canvas width (640px).
