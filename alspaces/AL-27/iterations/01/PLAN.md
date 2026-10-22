# Implementation Plan: AL-27 Metric Progression Custom Window, 1-Day Window, and Hourly Granularity

## 1. Goal and Iteration Scope

Deliver complete implementation of ticket AL-27 across Go functional core, Go HTTP boundary adapters, and TypeScript presentation view:
- Extend Go domain enums with hourly granularity, 1-day window, and custom window identifiers.
- Update Go AggregateMetricsOverTime to parse full ISO timestamps and bucket into YYYY-MM-DDTHH:00:00Z buckets.
- Implement /api/progression HTTP handler in Go boundary adapter with query validation and integration tests.
- Extend TypeScript domain models, ISO timestamp parsers, and progression bucket derivation.
- Add hourly granularity toggle, 1-day window button, and custom date-time pickers to Metric Progression view.
- Update TrendChart X-axis tick formatting to render high-resolution hourly labels without overlap.
- Ensure strict static typing (zero any), zero emojis, and complete test suites across Go and TypeScript.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against concrete existing files and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/enums.go:
  - Add GranularityHourly to TimeGranularity enum.
  - Add TimeWindow1D and TimeWindowCustom to TimeWindow enum.
  - Add TimeWindowLast24Hours to TimeWindowPreset enum.
  - Update ParseTimeWindow to map 1d and custom.
- backend/internal/domain/aggregations.go:
  - Update invocation and ledger timestamp parsing in AggregateMetricsOverTime to utilize parseTimestamp instead of date-only parseISODate.
  - Extend bucket key derivation in AggregateMetricsOverTime to output YYYY-MM-DDTHH:00:00Z when granularity is hourly.
  - Extend FilterLedgersByTimeWindow and ResolveDatePreset to resolve 24-hour windows relative to reference timestamp.
- backend/internal/domain/models.go:
  - Verify TimeBucketMetric structure accommodates hourly bucket keys.
- backend/internal/domain/aggregations_test.go:
  - Pure unit tests for hourly aggregation, 24-hour filtering, and timestamp boundaries.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/http/dto.go:
  - Add slice of domain.TimeBucketMetric to generic ResponsePayload interface union.
  - Add WriteProgressionResponse helper function.
- backend/internal/adapters/http/handlers.go:
  - Extend parseDateBoundary to support YYYY-MM-DDTHH:mm format.
  - Extend resolveTemporalFilter to support 1d (24-hour offset from UTC now) and custom.
  - Implement ProgressionHandler with query parameter extraction, validation, and JSON response.
- backend/internal/adapters/http/server.go:
  - Register route GET /api/progression.
- backend/internal/adapters/http/server_test.go:
  - HTTP integration tests for /api/progression, 1-day window, hourly granularity, and error responses.

### TypeScript Domain & API (web/src/domain/ & web/src/api/)
- web/src/domain/types.ts:
  - Add hourly to TimeGranularity union.
  - Add 1d and custom to TimeWindowPreset union.
- web/src/domain/progression.ts:
  - Update timestamp parser to preserve hours, minutes, and seconds from ISO strings.
  - Extend getBucketKey to output YYYY-MM-DDTHH:00:00Z for hourly granularity.
  - Update aggregateMetricsOverTime to aggregate invocations and code into hourly buckets.
- web/src/domain/progression.test.ts:
  - Unit tests verifying hourly bucketing, timestamp boundaries, and cumulative rollups.
- web/src/domain/formatters.ts:
  - Implement formatChartAxisTimestamp to format hourly ISO strings into clean tick labels.
- web/src/domain/formatters.test.ts:
  - Unit tests for axis timestamp formatting.
- web/src/api/client.ts:
  - Add fetchProgressionMetrics with network fetch and client-side fallback.
  - Support 1d and custom parameters in fetchWorkspaceLedgers and fetchFleetSummary.

### TypeScript Presentation (web/src/views/ & web/src/components/)
- web/src/views/MetricProgression.tsx:
  - Add state for customSince and customUntil.
  - Add Hourly toggle button to Granularity controls.
  - Add 24 Hours (1d) and Custom Range (custom) buttons to Window controls.
  - Conditionally render date-time inputs when custom is active.
  - Update date boundary calculation to compute 24-hour offset for 1d and pass date-times.
  - Preserve default states: Project Lifetime window and Daily granularity.
- web/src/views/MetricProgression.test.tsx:
  - Component tests for 1-day filtering, custom range inputs, hourly granularity, and zero emojis.
- web/src/components/TrendChart.tsx:
  - Update X-axis label rendering to format hourly timestamps cleanly via formatChartAxisTimestamp.
  - Prevent tick label collision on dense time series.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Implementation
- Task 1.1: In backend/internal/domain/enums.go:
  - Add constant GranularityHourly TimeGranularity = "hourly".
  - Add constant TimeWindow1D TimeWindow = "1d" and TimeWindowCustom TimeWindow = "custom".
  - Add constant TimeWindowLast24Hours TimeWindowPreset = "Last 24 Hours".
  - Update ParseTimeWindow(val string) TimeWindow to return TimeWindow1D when input is "1d", and TimeWindowCustom when input is "custom".
- Task 1.2: In backend/internal/domain/aggregations.go:
  - In ResolveDatePreset: add case for TimeWindowLast24Hours, evaluating start timestamp as 24 hours prior to reference timestamp (refDate.Add(-24 * time.Hour)), and end timestamp as reference timestamp.
  - In FilterLedgersByTimeWindow: add case for TimeWindow1D, filtering ledgers whose effective completion date is between refDate.Add(-24 * time.Hour) and refDate (inclusive).
  - Refactor AggregateMetricsOverTime:
    - Replace invocation date parsing call from date-only parseISODate to full timestamp parseTimestamp, preserving hour and minute precision in UTC.
    - Check if granularity equals GranularityHourly (case-insensitive comparison).
    - If hourly, construct bucket key as invDate.UTC().Format("2006-01-02T15:00:00Z").
    - If weekly, construct bucket key using existing ISO Monday calculation formatted as YYYY-MM-DD.
    - Otherwise (daily and cumulative), construct bucket key formatted as YYYY-MM-DD.
    - Parse ledger code delivery date using parseTimestamp and apply identical bucket key derivation.
    - Accumulate metrics into bucket accumulator map and construct sorted slice of TimeBucketMetric with running cumulative metrics and prompt cache hit rate percentage.
- Task 1.3: In backend/internal/domain/aggregations_test.go:
  - Add table-driven unit tests for AggregateMetricsOverTime with GranularityHourly:
    - Assert events within same hour combine into single bucket.
    - Assert events in distinct hours generate chronological buckets.
    - Assert timestamp bounds (since and until) correctly include and exclude hourly events.
    - Assert code delivery facts attach to matching hourly buckets.
    - Assert empty ledgers return empty slice.
  - Add unit test for FilterLedgersByTimeWindow with TimeWindow1D.
  - Add unit test for ResolveDatePreset with TimeWindowLast24Hours.

### Phase 2: Go Boundary Shell Adapters & Endpoints
- Task 2.1: In backend/internal/adapters/http/dto.go:
  - Extend ResponsePayload interface type constraint to include slice of domain.TimeBucketMetric.
  - Add function WriteProgressionResponse(w http.ResponseWriter, buckets []domain.TimeBucketMetric) calling WriteJSON(w, http.StatusOK, buckets).
- Task 2.2: In backend/internal/adapters/http/handlers.go:
  - In parseDateBoundary: add parsing support for layout "2006-01-02T15:04" (standard HTML datetime-local output format without seconds) resolving to UTC.
  - In resolveTemporalFilter:
    - Add case for windowParam == "1d": evaluate current time as time.Now().UTC(), set start bound to now.Add(-24 * time.Hour), and set end bound to now.
    - For windowParam == "custom": ensure since and until parameters are parsed directly.
  - Implement ProgressionHandler(workspacesDir string) http.HandlerFunc:
    - Extract query parameters: window, granularity, since, until.
    - Default granularity to daily if parameter is omitted.
    - Validate granularity: if not one of hourly, daily, weekly, cumulative, write HTTP 400 Bad Request error.
    - Resolve temporal bounds using resolveTemporalFilter. If parsing error occurs, write HTTP 400 Bad Request error.
    - Scan workspace ledgers via adapters.ScanWorkspaces(workspacesDir).
    - Invoke domain.AggregateMetricsOverTime(report.Ledgers, granularity, since, until).
    - Write response via WriteProgressionResponse.
- Task 2.3: In backend/internal/adapters/http/server.go:
  - Register route mux.HandleFunc("GET /api/progression", ProgressionHandler(workspacesDir)).
- Task 2.4: In backend/internal/adapters/http/server_test.go:
  - Write test verifying GET /api/progression default returns daily buckets for project lifetime.
  - Write test verifying GET /api/progression?granularity=hourly&window=1d returns hourly buckets.
  - Write test verifying GET /api/progression?window=custom&since=2026-10-15T00:00:00Z&until=2026-10-15T23:59:59Z&granularity=hourly.
  - Write test verifying invalid granularity returns HTTP 400 Bad Request.
  - Write test verifying invalid date string returns HTTP 400 Bad Request.

### Phase 3: TypeScript Domain & API Enhancements
- Task 3.1: In web/src/domain/types.ts:
  - Update TimeGranularity to 'hourly' | 'daily' | 'weekly' | 'cumulative'.
  - Update TimeWindowPreset to 'all' | '1d' | '7d' | '30d' | '90d' | 'custom'.
- Task 3.2: In web/src/domain/progression.ts:
  - Implement or update parseIsoTimestampToUtc (and update parseIsoDateToUtc):
    - When string contains T, parse full date and time components, constructing UTC Date preserving exact hours, minutes, and seconds.
    - When string is date-only (YYYY-MM-DD), construct UTC Date at 00:00:00.000 UTC.
    - Validate date validity defensively (rejecting invalid month days, roll-over dates, and non-leap years).
  - Update getBucketKey:
    - Support hourly granularity: format Date as YYYY-MM-DDTHH:00:00Z using UTC year, month, date, and hour.
    - Support weekly granularity: format ISO Monday as YYYY-MM-DD.
    - Support daily granularity: format as YYYY-MM-DD.
  - Update aggregateMetricsOverTime:
    - Parse invocation timestamps and ledger code timestamps using full timestamp parser.
    - Apply boundary checks against start and end boundaries.
    - Accumulate metrics into hourly, daily, or weekly buckets.
    - Compute running cumulative metrics across sorted keys.
- Task 3.3: In web/src/domain/formatters.ts:
  - Implement formatChartAxisTimestamp(label: string): string:
    - If label contains T (e.g. 2026-10-19T14:00:00Z), extract month-day and hour (10-19 14:00).
    - If label length is 10 (YYYY-MM-DD), extract month-day (10-19).
    - Return string without trailing seconds or time-zone tokens.
- Task 3.4: In web/src/api/client.ts:
  - Implement fetchProgressionMetrics(windowParam?: string, granularityParam?: string, sinceParam?: string, untilParam?: string):
    - Construct URL search parameters.
    - Fetch /api/progression.
    - If fetch fails or runs in mock mode, fetch ledgers and compute aggregateMetricsOverTime client-side as fallback.
  - Update fetchWorkspaceLedgers and fetchFleetSummary to support 1d and custom window parameters.

### Phase 4: TypeScript Presentation & Chart Axis Formatting
- Task 4.1: In web/src/views/MetricProgression.tsx:
  - Introduce states:
    - customSince: string defaulting to empty string.
    - customUntil: string defaulting to empty string.
  - Ensure initial states remain preset = 'all' and granularity = 'daily'.
  - In controls bar:
    - Add Hourly button to granularity-toggle-group before Daily.
    - Add 24 Hours button (preset === '1d') to preset-toggle-group.
    - Add Custom Range button (preset === 'custom') to preset-toggle-group.
  - Conditionally render custom date-time inputs when preset === 'custom':
    - Container class custom-date-inputs.
    - Input for Since with type datetime-local and label Since:.
    - Input for Until with type datetime-local and label Until:.
  - In useMemo boundary calculation:
    - If preset === '1d', find latest timestamp across ledgers; set endStr to latest ISO timestamp, and startStr to timestamp 24 hours prior (latestMs - 86_400_000).
    - If preset === 'custom', set startDate to customSince and endDate to customUntil.
    - If preset === 'all', set boundaries to null.
    - For 7d, 30d, 90d, calculate day offsets.
  - Update chartData mapping: pass formatted bucket_date to TrendChart.
- Task 4.2: In web/src/components/TrendChart.tsx:
  - Update X-axis tick label rendering:
    - Apply formatChartAxisTimestamp to raw bucket date label.
    - Ensure dynamic xTickStep guarantees no more than 6-7 tick marks render across chart width, avoiding tick label collisions for 24-hour periods.

---

## 4. Adversarial Testing Strategy for Tester

### Pure Go Domain Tests (backend/internal/domain/aggregations_test.go)
- Table-driven TestAggregateMetricsOverTime_Hourly:
  - Verify multiple invocations within the same hour bucket together into single YYYY-MM-DDTHH:00:00Z entry.
  - Verify invocations across adjacent hours produce distinct consecutive hourly buckets.
  - Verify inclusive start and end timestamp boundaries at hourly resolution.
  - Verify delivered code lines added and deleted attach to the hour bucket of the ledger completion date.
  - Verify cumulative cost, cumulative invocations, and cumulative tokens calculate monotonically across hourly buckets.
  - Verify prompt cache hit rate calculates accurately for hourly buckets.
  - Verify empty ledgers or ledgers with no matching invocations return empty slice without panic.
- Unit test TestFilterLedgersByTimeWindow_1D:
  - Verify ledgers completed within 24 hours of reference time are included.
  - Verify ledgers completed 24 hours and 1 second prior are excluded.
- Unit test TestResolveDatePreset_Last24Hours:
  - Verify start timestamp is exactly 24 hours prior to reference timestamp.

### Go HTTP Adapter Tests (backend/internal/adapters/http/server_test.go)
- Test GET /api/progression default:
  - Verify status code 200 and daily bucket metrics across all fixtures.
- Test GET /api/progression?granularity=hourly&window=1d:
  - Verify status code 200 and hourly buckets within last 24 hours.
- Test GET /api/progression?window=custom&since=2026-10-15T00:00:00Z&until=2026-10-15T12:00:00Z&granularity=hourly:
  - Verify date-time range boundaries.
- Test GET /api/progression?granularity=unsupported:
  - Verify status code 400 Bad Request with structured error JSON.
- Test GET /api/progression?since=invalid-timestamp:
  - Verify status code 400 Bad Request with structured error JSON.
- Test GET /api/fleet?window=1d and GET /api/workspaces?window=1d:
  - Verify status code 200 with 24-hour filtered data.

### TypeScript Unit & Formatter Tests (web/src/domain/progression.test.ts & formatters.test.ts)
- progression.test.ts:
  - Verify parseIsoTimestampToUtc parses ISO strings with hours and minutes in UTC.
  - Verify getBucketKey returns YYYY-MM-DDTHH:00:00Z when hourly granularity is active.
  - Verify aggregateMetricsOverTime produces hourly buckets matching expected sums.
  - Verify 1-day window calculation restricts to 24-hour interval.
  - Verify custom date-time interval restricts to specified interval.
- formatters.test.ts:
  - Verify formatChartAxisTimestamp formats 2026-10-19T14:00:00Z to 10-19 14:00.
  - Verify formatChartAxisTimestamp formats 2026-10-19 to 10-19.

### TypeScript Component & View Tests (web/src/views/MetricProgression.test.tsx)
- Verify Metric Progression defaults to Project Lifetime window and Daily granularity.
- Verify clicking Hourly granularity toggle activates hourly mode and renders updated chart data.
- Verify clicking 24 Hours window toggle activates 1d preset and restricts data to past 24 hours.
- Verify clicking Custom Range reveals Since and Until date-time input fields.
- Verify entering custom dates triggers reactive metric updates.
- Verify TrendChart SVG elements render cleanly without tick label collision or overflow.
- Verify zero emojis in document body across all views, controls, and tooltips.

---

## 5. Verification and Quality Gates

Execute full validation sequence before submitting iteration deliverables:
- Backend static analysis: go vet ./...
- Backend unit and race test suite: go test -race ./...
- Frontend typechecking: npm run typecheck
- Frontend linting: npm run lint
- Frontend unit tests: npm test
- Single binary build verification: make check && make test && make build
- Assert zero any across all modified Go and TypeScript files.
- Assert zero emojis across code, tests, UI strings, and artifacts.
