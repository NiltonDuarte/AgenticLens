# Implementation Plan: AL-4 TypeScript Ticket Deep Dive, Progression Charts, and Comparisons

## 1. Goal and Iteration Scope
Deliver the complete analytical suite in the AgenticLens TypeScript web frontend:
1. Implement pure domain progression rollups in web/src/domain/progression.ts with exact mathematical parity to Go AggregateMetricsOverTime.
2. Expand web/src/domain/types.ts and web/src/domain/formatters.ts with progression and comparison structures.
3. Enhance boundary adapters in web/src/api/client.ts and web/src/api/mockData.ts.
4. Build reusable UI visualization components (pure SVG responsive charts, saturation gauges, thread logs, ticket selector).
5. Build the three core views: TicketDeepDive.tsx, MetricProgression.tsx, and TicketComparisons.tsx.
6. Wire seamless tab navigation and ticket selection routing across web/src/App.tsx and web/src/components/Navbar.tsx.
7. Verify static typing, linting, pure domain unit tests, component render tests, and embedded Go binary compilation.

Zero code blocks are permitted in this plan. Implementation belongs exclusively to the coder subagent.

## 2. Grounded Reference Architecture & Symbol Map
- Backend Domain Aggregations: internal/domain/aggregations.go (AggregateMetricsOverTime, GetBaseHealthStatus, ParseISODate, ComputeWorkflowEfficiency)
- Backend Domain Models: internal/domain/models.go (TimeBucketMetric, AgenticMetricsLedger, WorkflowEfficiency, ContextWindowHealth, ThreadExecutionRecord)
- Backend HTTP Handlers: internal/adapters/http/handlers.go (GET /api/fleet, GET /api/workspaces, GET /api/workspaces/{ticketId})
- Frontend Scaffolding: web/package.json, web/tsconfig.json, web/vite.config.ts
- Frontend Domain Types: web/src/domain/types.ts
- Frontend Pure Formatters & Aggregations: web/src/domain/formatters.ts, web/src/domain/progression.ts
- Frontend Boundary Adapters: web/src/api/client.ts, web/src/api/mockData.ts, web/src/api/types.ts
- Frontend Presentation Components:
  - Navigation: web/src/components/Navbar.tsx
  - Trend Visualizer: web/src/components/TrendChart.tsx
  - Context Gauge & Saturation: web/src/components/SaturationBar.tsx
  - Threads Log: web/src/components/ThreadViewer.tsx
  - Workspace Artifacts: web/src/components/ArtifactViewer.tsx
  - Ticket Selector: web/src/components/TicketSelector.tsx
  - Ticket Directory Table: web/src/components/TicketTable.tsx
- Frontend Presentation Views:
  - web/src/views/FleetOverview.tsx
  - web/src/views/TicketDeepDive.tsx
  - web/src/views/MetricProgression.tsx
  - web/src/views/TicketComparisons.tsx
- Application Shell: web/src/App.tsx, web/src/index.css
- Automated Tests:
  - web/src/domain/progression.test.ts
  - web/src/domain/formatters.test.ts
  - web/src/views/TicketDeepDive.test.tsx
  - web/src/views/MetricProgression.test.tsx
  - web/src/views/TicketComparisons.test.tsx
  - web/src/components/Navbar.test.tsx

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Domain Models and Pure Progression Rollups (web/src/domain/)
- Task 1.1: Extend web/src/domain/types.ts:
  - Define TimeBucketMetric interface mirroring internal/domain/models.go with immutable readonly properties.
  - Define TimeGranularity union ('daily', 'weekly', 'cumulative').
  - Define TimeWindowPreset union ('all', '7d', '30d', '90d', 'custom').
  - Define ActiveTab union ('overview', 'deep-dive', 'progression', 'comparisons').
  - Define ComparisonSortField union and ComparisonSortOrder union ('asc', 'desc').
- Task 1.2: Implement web/src/domain/progression.ts:
  - Implement pure helper parseIsoDateToUtc converting ISO date strings to UTC Date objects safely, returning null on invalid input.
  - Implement pure helper getBucketKey: for daily returns YYYY-MM-DD; for weekly finds ISO Monday of the week and returns YYYY-MM-DD.
  - Implement pure function aggregateMetricsOverTime:
    - Traverses invocations across ledgers, extracting authoritative completion timestamp (with fallback chain: invocation completed_at, invocation started_at, ledger completed_at, ledger started_at, ledger updated_at).
    - Filters by start and end date boundaries when supplied.
    - Aggregates invocations, cost, input tokens, output tokens, cache read tokens, reasoning tokens, active duration, and peak context into bucket accumulators.
    - Accrues deliverable code lines added, lines deleted, and files changed into matching date bucket accumulators.
    - Sorts bucket keys chronologically.
    - Computes non-zero guarded prompt cache hit rates per bucket.
    - Computes running cumulative trajectories for cost, input tokens, output tokens, code lines added, code lines deleted, and invocations.
    - Returns readonly array of TimeBucketMetric objects.
    - Defensively handles empty inputs by returning an empty array.
- Task 1.3: Extend web/src/domain/formatters.ts:
  - Implement formatIsoDateOnly formatting ISO timestamps into YYYY-MM-DD.
  - Implement formatIsoDateTimeShort formatting ISO timestamps into readable date and time strings with zero emojis.
  - Extend deriveTicketComparisonRows to support dynamic sorting by arbitrary ComparisonSortField and ascending/descending order.
  - Ensure zero division by zero vulnerabilities in all helper calculations.

### Phase 2: Boundary Adapters and API Client Enhancements (web/src/api/)
- Task 2.1: Update web/src/api/client.ts:
  - Implement fetchWorkspaceLedgers: makes GET /api/workspaces, returning ApiResponse containing raw readonly AgenticMetricsLedger array. On network error or non-200 status, returns mockAgenticMetricsLedgers array with isMock set to true.
  - Ensure fetchWorkspaceById cleanly retrieves workspace detail and artifacts, falling back to mockWorkspaceDetails fixture on network disconnection.
  - Ensure zero any typing across all client calls.
- Task 2.2: Expand web/src/api/mockData.ts:
  - Enrich mockAgenticMetricsLedgers to include complete invocation records, session threads (ThreadExecutionRecord), and tool usages for AL-1, AL-2, PO-1, and GEN-1.
  - Enrich mockWorkspaceDetails with authentic markdown content for CONTRACT.md and SUMMARY.md for each mock ticket.
  - Verify zero emojis across all mock titles, descriptions, and text fields.

### Phase 3: Reusable UI Presentation Components (web/src/components/)
- Task 3.1: Implement web/src/components/TrendChart.tsx:
  - Pure SVG responsive line, area, and multi-bar visualization component without external libraries.
  - Props: title, series definitions, data points, value formatter, height, optional baseline target lines (e.g. 50% baseline, 80% target).
  - Renders SVG viewBox with responsive layout, horizontal grid reference lines, axis ticks, formatted labels, and hovered point tooltip overlays.
  - Handles empty series gracefully with descriptive fallback message.
  - Strict zero emojis in chart labels, tooltips, and legends.
- Task 3.2: Implement web/src/components/SaturationBar.tsx:
  - Visual context window health component.
  - Displays context health status badge (HEALTHY, ELEVATED, BLOATED, CRITICAL) using distinct CSS color variables.
  - Displays orchestrator saturation percentage and token count relative to 200,000 ceiling.
  - Displays normalized progress bar with indicator tick at ceiling mark.
  - Renders subagents peak role and token volume.
- Task 3.3: Implement web/src/components/ThreadViewer.tsx:
  - Displays session thread execution records with collapsible accordion sections.
  - Displays session ID, user, compactions count, and role/model metrics breakdown table.
- Task 3.4: Implement web/src/components/ArtifactViewer.tsx:
  - Tabbed or sectional viewer for CONTRACT.md and SUMMARY.md workspace artifacts.
  - Displays pre-formatted markdown text with monospace styling.
  - Renders clean empty state when artifacts are missing.
- Task 3.5: Implement web/src/components/TicketSelector.tsx:
  - Dropdown selector listing tickets with ID, title, and context health pill.
  - Emits onSelectTicket callback when user picks a workspace.
- Task 3.6: Update web/src/components/Navbar.tsx:
  - Add navigation tabs for Fleet Overview, Ticket Deep Dive, Metric Progression, and Comparisons.
  - Remove disabled flag from Ticket Deep Dive.
  - Highlight active tab with active CSS class.
  - Maintain status pill for Live Server vs Mock Mode.

### Phase 4: Analytical Views and Application Shell (web/src/views/, web/src/App.tsx)
- Task 4.1: Implement web/src/views/TicketDeepDive.tsx:
  - Ticket selector at top allowing user to choose any loaded workspace.
  - Header showing ticket ID, title, ticket type, started at, completed at, and extraction update timestamps.
  - Top row KPI cards: SDLC iterations, lead time, active duration, total cost, prompt cache hit rate, test:prod ratio.
  - Second row cards: Code delivery stats (lines added, deleted, files changed, test vs prod breakdown), reasoning tokens and ratio, tool execution reliability (calls, corrections, success rate %).
  - Context Window Health section with SaturationBar and Peak Context Tokens by Role horizontal bar distribution.
  - Subagent Turn Velocity section: average turns, peak turns, and turns by role distribution.
  - Role and Model metrics breakdown tables.
  - Session Thread Records section using ThreadViewer.
  - Workspace Artifacts section using ArtifactViewer.
  - Handles missing or unselected ticket with EmptyState.
- Task 4.2: Implement web/src/views/MetricProgression.tsx:
  - Top control bar: time window preset selector (Lifetime, 7D, 30D, 90D) and granularity radio toggle (Daily, Weekly, Cumulative).
  - Active window KPI banner: Filtered Spend, Invocations, Total Tokens, Overall Cache Hit Rate, Net Code Delivered.
  - Tabbed or sectional SVG trend charts:
    - Financial & Invocations: Period spend vs cumulative spend, invocations count.
    - Token Throughput & Cache Efficiency: Prompt input, completion output, cache read tokens, and cache hit rate trajectory with 50% and 80% reference lines.
    - Code Delivery & Growth: Periodic lines added and deleted, net code line growth.
    - Context Saturation: Peak context window tokens across buckets evaluated against 200,000 ceiling.
  - Handles empty date ranges with clear empty message.
- Task 4.3: Implement web/src/views/TicketComparisons.tsx:
  - Benchmarking summary banner: total tickets, average active duration, average iterations, fleet tool reliability, average cost.
  - Interactive search bar filtering tickets by ID or title.
  - Comprehensive comparison matrix table: Ticket ID, Title, Status, Completed At, Started At, Lead Time, Active Duration, Iterations, Cost, Cache Hit Rate, Test:Prod Ratio, Code Delivered (+/- lines, files), Tool Success Rate, Context Health, Orchestrator Saturation.
  - Column header sorting with ascending/descending toggle and visual direction indicator.
  - Authoritative default sort: deliverable completion date descending, followed by active tickets.
  - Row click triggers onSelectTicket, transitioning the dashboard directly to Ticket Deep Dive for that ticket.
- Task 4.4: Update web/src/App.tsx:
  - Top-level state: activeTab ('overview' | 'deep-dive' | 'progression' | 'comparisons'), selectedTicketId (string), isMock (boolean).
  - Navigation handlers: switching tabs, selecting ticket from Fleet Overview or Ticket Comparisons switches activeTab to 'deep-dive' with selectedTicketId set.
  - Passes isMock state to Navbar.
- Task 4.5: Update web/src/index.css:
  - Add styles for SVG trend charts (lines, fills, axes, grid, tooltips).
  - Add styles for context health gauges, saturation progress bars, ticket selector, and thread accordions.
  - Maintain dark theme consistency, monospace numbers, and zero emojis.

### Phase 5: Automated Test Suite (Tester Subagent Tasks)
- Task 5.1: Create web/src/domain/progression.test.ts:
  - Test aggregateMetricsOverTime with daily granularity: verifies date bucketing, token sums, cost rounding, cache hit calculation.
  - Test aggregateMetricsOverTime with weekly granularity: verifies ISO Monday bucketing across month boundaries.
  - Test aggregateMetricsOverTime with cumulative granularity: verifies running sums of cost, tokens, and code lines.
  - Test date window boundary filtering (start date, end date, out-of-range exclusion).
  - Test division by zero defense (zero input tokens, zero cache read tokens).
  - Test edge cases: empty ledger list, missing completion timestamps, single invocation.
- Task 5.2: Create web/src/views/TicketDeepDive.test.tsx:
  - Test rendering all core KPI cards with mock ticket data.
  - Test rendering context window saturation bar and role distribution.
  - Test rendering subagent turn velocity breakdown.
  - Test rendering tool execution success rate.
  - Test rendering role and model tables.
  - Test switching selected ticket via ticket selector.
  - Test rendering workspace artifacts (CONTRACT.md, SUMMARY.md).
  - Test graceful rendering of empty ticket state.
  - Assert zero emojis in rendered DOM text.
- Task 5.3: Create web/src/views/MetricProgression.test.tsx:
  - Test rendering granularity controls (Daily, Weekly, Cumulative).
  - Test active window KPI summary calculation.
  - Test rendering SVG trend chart elements (polyline, path, rect).
  - Test toggling granularity updates chart data points.
  - Test rendering empty state when no ledgers match window.
  - Assert zero emojis in rendered DOM text.
- Task 5.4: Create web/src/views/TicketComparisons.test.tsx:
  - Test rendering comparison matrix table with mock rows.
  - Test column sorting by cost, iterations, duration, and completion date.
  - Test search input filtering rows by ticket ID and title.
  - Test row click invoking onSelectTicket callback.
  - Assert zero emojis in rendered DOM text.
- Task 5.5: Create web/src/components/Navbar.test.tsx:
  - Test rendering all 4 navigation tabs.
  - Test clicking tabs invokes onSelectTab callback.
  - Test status pill rendering for Live Server and Mock Mode.
  - Assert zero emojis across all navigation elements.

### Phase 6: Build Verification, Single Binary Parity, and Delivery Sync
- Task 6.1: Run npm run typecheck in web/ verifying 100% strict TypeScript compilation with zero errors and zero any.
- Task 6.2: Run npm run lint in web/ verifying zero ESLint warnings or errors.
- Task 6.3: Run npm test in web/ verifying all unit, component, and integration tests pass cleanly.
- Task 6.4: Run npm run build in web/ verifying production SPA assets compile and copy to internal/adapters/web/dist/.
- Task 6.5: Run root make check, make test, and make build verifying backend tests, embedded filesystem tests, and compilation of bin/agentic-lens.
- Task 6.6: Verify single-binary server startup using bin/agentic-lens server --port 8080 --workspaces workspaces, confirming all 4 views function over HTTP.
- Task 6.7: Ingest development telemetry facts using python3 scripts/extract_agentic_metrics.py.

## 4. Key Defensive Mitigations
- Zero Denominators in Progression: When both input tokens and cache read tokens are zero, prompt cache hit rate defaults to 0.0% without NaN or Infinity.
- Date Parsing Resilience: Event timestamps may appear in multiple formats or be absent in unfinished sessions. The fallback chain (completed_at -> started_at -> updated_at) guarantees valid bucketing without throwing exceptions.
- SVG Chart Responsiveness: Charts use scalable viewBox coordinates and percentage widths to adapt cleanly across screen widths without clipping labels or axes.
- Zero Emojis Auditing: Continuous test assertions and code review checks guarantee no emojis enter titles, badges, tooltips, or legends.
