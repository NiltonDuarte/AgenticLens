# Contract: AL-3 TypeScript SPA Shell, Fleet Overview Dashboard, and Single-Binary Integration

## 1. Problem Statement
The legacy Python Streamlit prototype coupled presentation state with runtime execution, suffered high memory and process overhead, required an active Python interpreter, and could not be compiled into a self-contained executable.

With AL-2 now complete, the Go HTTP server, REST endpoints (/api/fleet, /api/workspaces, /api/workspaces/{ticketId}), and embedded filesystem adapter (internal/adapters/web/embed.go) are operational. AL-3 completes the vertical slice by delivering the modern TypeScript web frontend application (web/), compiling and embedding the production SPA bundle into the Go single binary (internal/adapters/web/dist), wiring live API client communication with mock fallback resilience, and verifying end-to-end server delivery with zero Node.js runtime dependency in production.

## 2. Functional Core, Imperative Shell (FCIS) Invariants
All implementation must strictly comply with TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward:
1. Presentation Views and Components (web/src/views/, web/src/components/) depend on Boundary Adapters and Domain Types.
2. Boundary Adapters and API Client (web/src/api/) depend on Domain Types.
3. Domain Types and Pure Formatters (web/src/domain/) have zero external dependencies and zero awareness of HTTP, browser DOM, or UI frameworks.
4. Single-binary embedding shell (internal/adapters/web/) treats web/dist output as an opaque static asset tree.

### Pure Domain Core Isolation (web/src/domain/)
1. All domain models, contracts, and pure calculation or formatting utilities reside in web/src/domain/.
2. Formatters (compact token notation, currency formatting, duration formatting, percentage formatting) are deterministic pure functions.
3. Domain utilities do not invoke browser DOM APIs, network calls, local storage, or system clocks.
4. All division operations (cache hit rate, saturation percentages, cost distribution ratios) defend against zero denominators.

### Imperative Boundary Shell (web/src/api/ and internal/adapters/web/)
1. Network access and JSON deserialization are strictly confined to web/src/api/.
2. The API client implements dual-mode operation: it communicates with live backend HTTP endpoints (/api/fleet, /api/workspaces) and seamlessly falls back to deterministic mock fixtures (web/src/api/mockData.ts) during disconnected development or network interruptions.
3. Asynchronous network states are modeled as type-safe discriminated unions: idle, loading, success, and error.
4. Runtime error responses and malformed payloads are narrowed and mapped to descriptive typed error models without uncaught promise rejections.
5. The compiled SPA assets must be placed into internal/adapters/web/dist/ to satisfy Go embed.FS embedding directives in internal/adapters/web/embed.go.

### Presentation Shell (web/src/components/, web/src/views/)
1. Components are purely presentational, accepting typed props and emitting typed events.
2. Leaf UI components never initiate network requests or mock fallbacks directly; data fetching belongs exclusively to container views or custom hooks.
3. Every component and view handles empty or zero-value states defensively with clear explanatory empty-state messages.
4. Strict Zero Emojis Mandate: No emojis in page titles, navigation tabs, metric cards, table headers, status tags, distribution charts, or tooltips.

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope for AL-3
1. Scaffolding the web/ application directory with Vite, React, TypeScript 5.x+, ESLint, and Vitest.
2. Mirroring Go backend domain contracts from internal/domain/models.go and internal/domain/enums.go into strictly typed TypeScript interfaces under web/src/domain/types.ts.
3. Implementing pure domain formatters under web/src/domain/formatters.ts with parity to Go formatters in internal/domain/constants.go and internal/domain/aggregations.go.
4. Implementing the type-safe boundary HTTP client and mock fixture harness in web/src/api/client.ts and web/src/api/mockData.ts.
5. Building reusable UI components under web/src/components/:
   - Top-level navigation bar with active view state and connection status.
   - KPI metric cards for Total Cost, Invocations, Total Tokens, Cache Hit Rate, and Active Duration.
   - Role and Model cost distribution cards with percentage breakdowns.
   - Ticket directory table with search filtering and column sorting.
   - Reusable empty-state display component.
6. Building the Fleet Overview view under web/src/views/FleetOverview.tsx and top-level application shell under web/src/App.tsx.
7. Unit and component tests in web/src/domain/formatters.test.ts, web/src/api/client.test.ts, and web/src/views/FleetOverview.test.tsx.
8. Compiling production SPA assets via Vite directly targeting or synchronizing into internal/adapters/web/dist/, replacing the AL-2 placeholder index.html.
9. Integrating the frontend build step into root Makefile so that make build produces a self-contained bin/agentic-lens executable with embedded SPA assets.
10. Live backend API integration with the Go HTTP server (/api/fleet, /api/workspaces) and end-to-end verification of agentic-lens server serving the compiled SPA and live API data without Node.js runtime.

### Out of Scope for AL-3
1. Ticket Deep Dive view, subagent turn velocity breakdown, and context window waterfall inspection (governed by AL-4).
2. Anomaly detection rule engine and alert visualization (governed by AL-7 and AL-8).
3. WebSockets or Server-Sent Events for real-time telemetry push (future streaming ticket).

## 4. Architectural Contracts and Data Structures

### Domain Contracts (web/src/domain/types.ts)
1. InvocationTokens: input (number), output (number), cache_read (number), reasoning (number), peak_context (number).
2. InvocationFact: id (string), parent_id (string or null), role (string), model (string), user (string), started_at (string or null), completed_at (string or null), active_duration_seconds (number), turns (number), tokens (InvocationTokens), cost_usd (number), compactions (number).
3. ToolFact: name (string), calls (number), errors (number).
4. CodeFacts: lines_added (number), lines_deleted (number), test_lines_added (number), test_lines_deleted (number), prod_lines_added (number), prod_lines_deleted (number), files_changed (number), sdlc_iterations (number).
5. AgenticFactLedger: ticket_id (string or null), ticket_title (string or null), ticket_type (string or null), started_at (string or null), completed_at (string or null), updated_at (string), code (CodeFacts), invocations (array of InvocationFact), tools (array of ToolFact).
6. GroupedMetric: invocations (number), input_tokens (number), output_tokens (number), cache_read_tokens (number), cache_hit_rate_pct (number), cost_usd (number), peak_context_tokens (number), context_health (string).
7. TestToCodeRatio: ratio (string), test_lines_added (number), prod_lines_added (number).
8. SubagentTurnVelocity: avg_turns_per_invocation (number), peak_turns (number), by_role (record of string to number).
9. CodeDeliveredStats: lines_added (number), lines_deleted (number), files_changed (number).
10. ToolExecutionHealth: total_calls (number), errors_and_corrections (number), success_rate_pct (number).
11. ContextWindowHealth: healthy_ceiling_tokens (number), status (string), orchestrator_peak_tokens (number), orchestrator_saturation_pct (number), subagents_peak_tokens (number), subagents_peak_role (string), compactions (number), peak_by_role (record of string to number).
12. WorkflowEfficiency: sdlc_iterations (number), active_duration_seconds (number), active_duration_formatted (string), role_effort_distribution_pct (record of string to number), subagent_turn_velocity (SubagentTurnVelocity), test_to_production_ratio (TestToCodeRatio), reasoning_tokens (number), reasoning_ratio_pct (number), code_delivered (CodeDeliveredStats), tool_execution (ToolExecutionHealth), context_window_health (ContextWindowHealth).
13. ModelMetric and RoleMetric: role or model identifier (string), invocations (number), input_tokens (number), output_tokens (number), cache_read_tokens (number), cache_hit_rate_pct (number), cost_usd (number), peak_context_tokens (number), context_health (string).
14. ThreadRoleMetric: role (string), models (array of ModelMetric).
15. ThreadExecutionRecord: session_id (string), user (string), roles (array of ThreadRoleMetric), compactions (number).
16. AgenticMetricsLedger: ticket_id (string or null), ticket_title (string or null), ticket_type (string or null), started_at (string or null), completed_at (string or null), updated_at (string), facts (AgenticFactLedger), workflow_efficiency (WorkflowEfficiency), threads (array of ThreadExecutionRecord), totals (GroupedMetric), roles (array of RoleMetric), models (array of ModelMetric).
17. FleetSummary: total_tickets (number), total_cost_usd (number), total_invocations (number), total_input_tokens (number), total_output_tokens (number), total_cache_read_tokens (number), overall_cache_hit_rate_pct (number), total_duration_seconds (number), formatted_total_duration (string), total_lines_added (number), total_lines_deleted (number), total_files_changed (number), total_reasoning_tokens (number), overall_reasoning_ratio_pct (number), role_costs (record of string to number), model_costs (record of string to number).
18. TicketComparisonRow: ticket_id (string), updated_at (string), sdlc_iterations (number), active_duration_seconds (number), active_duration_formatted (string), cost_usd (number), cache_hit_rate_pct (number), test_to_code_ratio (string), lines_added (number), lines_deleted (number), files_changed (number), tool_success_rate_pct (number), context_health_status (string), orchestrator_saturation_pct (number), peak_context_tokens (number), ticket_title (string), started_at (string or null), completed_at (string or null), lead_duration_seconds (number or null), lead_duration_formatted (string).

### Domain Pure Formatters (web/src/domain/formatters.ts)
1. formatTokensK: Converts token numbers to compact notation (>= 1,000,000 to X.XM or XM, >= 1,000 to X.Xk or Xk, otherwise raw integer string).
2. formatCurrencyUSD: Formats cost numbers to two decimal places (or four for micro-costs), prefixed with dollar sign.
3. formatPercentage: Formats ratios and percentages to one decimal place appended with percent sign.
4. formatDurationSeconds: Converts seconds to human-readable strings (hours and minutes, or minutes and seconds).
5. formatLeadTime: Formats optional lead time seconds into days, hours, or minutes, defaulting to In Progress when null.

### Boundary API Client (web/src/api/client.ts)
1. fetchFleetSummary: Retrieves aggregated fleet data. Queries GET /api/fleet. On HTTP or network failure, returns mock fleet summary with isMock set to true.
2. fetchWorkspaces: Retrieves ticket directory rows. Queries GET /api/workspaces. On failure, returns mock workspace records with isMock set to true.
3. fetchWorkspaceById: Retrieves individual workspace ledger. Queries GET /api/workspaces/:ticketId. On failure, returns matching mock ledger or throws structured typed error.

### Go Single-Binary Embedding Boundary (internal/adapters/web/)
1. The production build of the React SPA is placed in internal/adapters/web/dist/, replacing the temporary baseline index.html.
2. internal/adapters/web/embed.go embeds all files in dist/ into distFS embed.FS.
3. internal/adapters/http/spa.go serves these assets at the root route / and falls back to index.html for client routes while maintaining strict 404 boundaries for missing /api/* routes.

## 5. Acceptance Criteria
1. Frontend build target npm run build produces optimized static bundle in web/dist/ and synchronizes or outputs to internal/adapters/web/dist/.
2. Static type checker npm run typecheck passes with zero errors under strict mode with zero any.
3. Linter npm run lint passes with zero warnings or errors.
4. Frontend test runner npm test passes all domain formatter, API client, and component tests.
5. Root make build compiles the Go executable bin/agentic-lens with the real React SPA assets embedded via embed.FS.
6. Execution of agentic-lens server launches the HTTP server and correctly serves the compiled React application at http://localhost:8080/ without requiring a running Node.js process.
7. Fleet Overview dashboard displays live aggregated telemetry retrieved from /api/fleet and /api/workspaces when the Go backend is serving workspaces.
8. SPA runs stand-alone in Vite development mode (npm run dev) using deterministic mock fixtures when backend is absent, clearly indicating mock mode in the UI.
9. Empty datasets render defensive zero states without unhandled runtime exceptions.
10. Absolute zero emojis across all UI strings, navigation headers, metric cards, tables, distribution charts, and tooltips.
