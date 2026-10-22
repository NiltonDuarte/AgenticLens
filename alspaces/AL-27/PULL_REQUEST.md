# feat(progression): high-resolution temporal tracking with 24h window, custom range, and hourly granularity

## Summary
Implements high-resolution temporal tracking across AgenticLens for ticket AL-27:
- Extends time window presets to include a 1-day (24-hour) window preset and arbitrary custom date-time windows.
- Introduces hourly aggregation granularity formatting bucket keys as `YYYY-MM-DDTHH:00:00Z` in UTC.
- Adds HTTP endpoint `GET /api/progression` supporting `window`, `granularity`, `startDate`, and `endDate` query parameters.
- Enhances frontend MetricProgression view with toggle controls for 24 Hours, Custom Range datetime inputs with validation, and Hourly granularity toggle.
- Updates TrendChart SVG visualization with X-axis timestamp formatting based on active granularity.

For details, refer to:
- Contract: [CONTRACT.md](workspaces/AL-27/CONTRACT.md)
- Work Summary & Telemetry: [SUMMARY.md](workspaces/AL-27/SUMMARY.md)

## Key Changes by Layer

### Domain Layer (Go Functional Core & TypeScript Domain)
- `backend/internal/domain/enums.go`: Added `TimeWindow1D`, `TimeWindowCustom`, `GranularityHourly`, and `TimeWindowLast24Hours`.
- `backend/internal/domain/aggregations.go`: Extended `ResolveDatePreset` for 24-hour window, `FilterLedgersByTimeWindow` for `1d`, `parseTimestamp` for `2006-01-02T15:04`, and `AggregateMetricsOverTime` for hourly bucketing and sub-day boundary filtering.
- `web/src/domain/types.ts`: Extended `TimeGranularity` with `'hourly'` and `TimeWindowPreset` with `'1d'`.
- `web/src/domain/progression.ts`: Updated `parseIsoDateToUtc` and `parseIsoBoundary` for ISO datetime parsing, `getBucketKey` and `aggregateMetricsOverTime` for hourly bucketing.
- `web/src/domain/formatters.ts`: Exported `formatChartAxisTimestamp` pure formatter.

### Adapters Layer (Go Boundary Shell & Frontend API Client)
- `backend/internal/adapters/http/dto.go`: Added `WriteProgressionResponse` typed with generic `ResponsePayload`.
- `backend/internal/adapters/http/handlers.go`: Updated `parseDateBoundary` and `resolveTemporalFilter` for datetime strings, implemented `ProgressionHandler`.
- `backend/internal/adapters/http/server.go`: Registered `GET /api/progression` route.
- `web/src/api/client.ts`: Added `fetchProgressionMetrics` with offline mock fallback.

### Presentation Layer (React Views, Components & Styling)
- `web/src/views/MetricProgression.tsx`: Added 24 Hours and Custom preset buttons, custom datetime inputs with validation, Hourly granularity button, and linked `granularity` prop to `TrendChart`.
- `web/src/components/TrendChart.tsx`: Formatted X-axis labels using `formatChartAxisTimestamp` according to selected granularity.
- `web/src/index.css`: Added styles for `.custom-datetime-bar`, `.datetime-input`, and `.datetime-warning`.

## Invariant Adherence
- FCIS architecture preserved: Zero I/O, clocks, or external imports in `domain`.
- Zero `any` and zero `interface{}` across Go and TypeScript.
- Zero emojis across code, comments, docstrings, UI labels, buttons, and logs.
- Backward compatibility: Metric Progression defaults to Project Lifetime ("all") window and Daily ("daily") granularity.
