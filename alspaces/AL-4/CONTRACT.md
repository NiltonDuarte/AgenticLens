# Contract: AL-4 TypeScript Ticket Deep Dive, Progression Charts, and Comparisons

## 1. Problem Statement
AL-3 delivered the initial TypeScript SPA shell, the embedded Go single-binary packaging, and the Fleet Overview dashboard. The legacy Python Streamlit prototype also contained three deep analytical perspectives:
1. Ticket Deep Dive: Single-ticket autopsy examining context window saturation, subagent turn velocity, tool reliability, role/model cost distributions, execution thread logs, and workspace artifacts.
2. Metric Progression: Longitudinal temporal velocity analysis tracking period spend, invocation volume, token throughput, cache hit rate trajectories, and net code growth across daily, weekly, and cumulative horizons.
3. Cross-Ticket Comparisons: Multi-ticket benchmarking matrix comparing turnaround lead time, active compute duration, test-to-production ratios, cost efficiency, tool success rates, and context saturation.

AL-4 migrates these three analytical views into the TypeScript frontend, connecting them to Go backend API endpoints and deterministic fallback fixtures, maintaining strict Functional Core, Imperative Shell (FCIS) separation, zero any types, zero emojis, and pure SVG/CSS responsive visualization with zero external charting dependencies.

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation adheres to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward:
1. Presentation Views and Components (web/src/views/, web/src/components/) depend on Boundary Adapters (web/src/api/) and Domain Types (web/src/domain/).
2. Boundary Adapters and API Client (web/src/api/) depend on Domain Types.
3. Domain Models, Aggregations, and Formatters (web/src/domain/) have zero external dependencies. They contain no DOM references, no HTTP calls, no local storage, and no system clock access.
4. Single-binary embedding shell (internal/adapters/web/) treats the web/dist output as an opaque static asset tree.

### Pure Domain Core Isolation (web/src/domain/)
1. Progression temporal rollup algorithms reside in web/src/domain/progression.ts, mirroring Go AggregateMetricsOverTime from internal/domain/aggregations.go with mathematical parity.
2. Cross-ticket comparison sorting and projection utilities reside in web/src/domain/formatters.ts.
3. Pure domain formatters and calculation functions are deterministic. Any timestamp boundaries or presets are passed as explicit parameters rather than querying system clocks inside core functions.
4. All division operations (prompt cache hit rates, reasoning ratios, tool success rates, saturation percentages) defend against zero denominators with explicit zero fallbacks.

### Imperative Boundary Shell (web/src/api/)
1. Network access and JSON deserialization are strictly confined to web/src/api/.
2. Dual-mode resilience: the client talks to live Go HTTP endpoints (/api/workspaces, /api/workspaces/:ticketId, /api/fleet) and falls back to deterministic mock fixtures (web/src/api/mockData.ts) during disconnected development or network errors.
3. Network fetch states are typed as discriminated unions (idle, loading, success, error) with zero uncaught promise rejections.
4. Payload deserialization uses unknown narrowing with runtime validation. No any types allowed.

### Presentation Shell (web/src/components/, web/src/views/)
1. Components are purely presentational: accepting readonly typed props and emitting typed callbacks.
2. Data fetching belongs exclusively to container views or custom hooks. Leaf UI components never initiate HTTP requests or mock fallbacks directly.
3. Zero External Charting Dependencies: Trend charts, progress bars, and gauges are implemented using responsive SVG elements and modern CSS. No heavy charting libraries (e.g. Chart.js, Recharts, Altair) permitted.
4. Defensive Presentation: Every view handles empty datasets, unselected tickets, and zero-value states with explicit fallback messages.
5. Strict Zero Emojis Mandate: Zero emojis across all labels, titles, table headers, status badges, chart axes, legends, tooltips, and console output.

## 3. Scope Boundaries

### In Scope for AL-4
1. Domain Types: Adding TimeBucketMetric, TimeGranularity, TimeWindowPreset, ComparisonSortField, and ViewTab to web/src/domain/types.ts matching Go domain models.
2. Pure Domain Aggregations: Implementing aggregateMetricsOverTime in web/src/domain/progression.ts with daily, weekly, and cumulative bucketing, matching internal/domain/aggregations.go.
3. Boundary Client Enhancements: Adding fetchWorkspaceLedgers to web/src/api/client.ts and expanding mock data fixtures in web/src/api/mockData.ts with detailed thread and artifact payloads.
4. Reusable Presentation Components in web/src/components/:
   - SVG Trend Chart component (line, area, and bar modes with grid lines, axes, and hover tooltips).
   - Context Health Gauge and Saturation Bar with color-coded token ceiling indicator.
   - Subagent Turn Velocity horizontal bar distribution.
   - Session Threads viewer with per-role model metrics.
   - Ticket Selector dropdown with status pills and title display.
   - Workspace Artifacts viewer (tabs for CONTRACT.md and SUMMARY.md).
5. Presentation Views in web/src/views/:
   - TicketDeepDive.tsx: Granular single-ticket autopsy (KPIs, context health, turn velocity, tool reliability, roles, models, threads, artifacts).
   - MetricProgression.tsx: Longitudinal velocity trends (time window presets, daily/weekly/cumulative granularity toggle, spend chart, token volume breakdown, cache hit trajectory, net code line churn).
   - TicketComparisons.tsx: Multi-ticket benchmarking matrix with multi-column sorting, filtering, and click-through to Ticket Deep Dive.
6. Navigation Wiring: Updating web/src/App.tsx and web/src/components/Navbar.tsx to support 4 active tabs (Fleet Overview, Ticket Deep Dive, Metric Progression, Comparisons) and pass ticket selection state between views.
7. Automated Testing: Pure domain unit tests for progression bucketing and comparison sorting, component render tests for all three views, and navigation integration tests.
8. Single-Binary Integration: Building frontend bundle into internal/adapters/web/dist and verifying single-binary execution via make build and agentic-lens server.

### Out of Scope for AL-4
1. Visual Trace Waterfall and Span Inspector (slated for AL-9).
2. Domain Anomaly Detection and Rule Engine (slated for AL-7 and AL-8).
3. Telemetry Extractor or Session Generation in Go (slated for AL-11).
4. WebSockets or Server-Sent Events (SSE) streaming.

## 4. Architectural Contracts and Data Structures

### Domain Types (web/src/domain/types.ts)
1. TimeBucketMetric:
   - bucket_key: string
   - bucket_date: string
   - invocations: number
   - cost_usd: number
   - input_tokens: number
   - output_tokens: number
   - cache_read_tokens: number
   - reasoning_tokens: number
   - cache_hit_rate_pct: number
   - lines_added: number
   - lines_deleted: number
   - files_changed: number
   - net_lines_added: number
   - peak_context_tokens: number
   - active_duration_seconds: number
   - cumulative_cost_usd: number
   - cumulative_input_tokens: number
   - cumulative_output_tokens: number
   - cumulative_lines_added: number
   - cumulative_lines_deleted: number
   - cumulative_invocations: number
2. TimeGranularity: 'daily' | 'weekly' | 'cumulative'
3. TimeWindowPreset: 'all' | '7d' | '30d' | '90d' | 'custom'
4. ActiveTab: 'overview' | 'deep-dive' | 'progression' | 'comparisons'
5. ComparisonSortField: 'completed_at' | 'started_at' | 'ticket_id' | 'cost_usd' | 'active_duration_seconds' | 'sdlc_iterations' | 'cache_hit_rate_pct' | 'tool_success_rate_pct' | 'lines_added' | 'orchestrator_saturation_pct'

### Pure Progression Aggregator (web/src/domain/progression.ts)
- aggregateMetricsOverTime(ledgers: readonly AgenticMetricsLedger[], granularity: TimeGranularity, startDate?: string | null, endDate?: string | null): readonly TimeBucketMetric[]
  - Buckets each invocation by completed_at (fallback: started_at, ledger completed_at, ledger started_at, ledger updated_at).
  - Daily: YYYY-MM-DD.
  - Weekly: Monday of the corresponding ISO week (YYYY-MM-DD).
  - Cumulative: Running totals accumulated chronologically over sorted bucket keys.
  - Code statistics bucketed per ledger deliverable completion timestamp.
  - Yields empty array on empty inputs or zero matches within range.

## 5. Acceptance Criteria
1. Deep Dive view displays full efficiency metrics for any selected workspace: SDLC iterations, lead time, active duration, cost, cache hit rate, test:prod ratio, code delivered, reasoning ratio, tool execution reliability, context health gauges, role/model distributions, session threads, and workspace artifacts.
2. Context window health accurately indicates HEALTHY, ELEVATED, BLOATED, or CRITICAL matching token ceiling thresholds, with orchestrator saturation percentage and ceiling normalized progress bar.
3. Metric Progression view supports toggling between Daily, Weekly, and Cumulative granularities, correctly updating summary banner KPIs and SVG trend charts for spend, tokens, cache hit rate, and code line churn.
4. Ticket Comparisons view renders all tickets in a tabular matrix with default sorting by deliverable completion date (most recent first, in-progress last), provides interactive column sorting and search filtering, and allows clicking a row to open the Ticket Deep Dive for that ticket.
5. Navigation between Fleet Overview, Ticket Deep Dive, Metric Progression, and Comparisons operates smoothly with Navbar visual active states and Live/Mock indicator.
6. Zero emojis present across all views, labels, tables, charts, axes, legends, and tooltips.
7. TypeScript strict type checking passes with zero errors and zero any.
8. Vitest suite passes all unit and component tests with zero skipped tests.
9. Root make build compiles the Go executable bin/agentic-lens with the embedded SPA containing all four operational views.