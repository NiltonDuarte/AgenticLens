# Implementation Plan: AL-3 TypeScript SPA Shell, Fleet Overview Dashboard, and Single-Binary Integration

## 1. Goal and Iteration Scope
The goal of iteration 01 is to deliver the complete vertical slice of the AgenticLens web frontend: scaffolding the web/ application using Vite, React, and strict TypeScript; implementing pure domain types and formatting functions mirroring Go backend contracts; building the dual-mode API client with live server communication and mock fixture fallback; rendering the Fleet Overview dashboard; compiling and embedding the production SPA bundle into the Go backend (internal/adapters/web/dist); and verifying single-binary execution via agentic-lens server without a Node.js runtime.

Upon completion of this iteration, running make check, make test, and make build will execute frontend and backend verification and produce a single standalone executable bin/agentic-lens capable of serving the complete UI and live telemetry.

## 2. Grounded Reference Architecture & Paths
The implementation maps directly to the following repository locations:
- Frontend Manifest & Tooling: web/package.json, web/vite.config.ts, web/tsconfig.json, web/tsconfig.node.json, web/index.html
- Domain Contracts: web/src/domain/types.ts (mirroring internal/domain/models.go and internal/domain/enums.go)
- Pure Formatters: web/src/domain/formatters.ts (mirroring internal/domain/constants.go and internal/domain/aggregations.go)
- Boundary Adapters & API Client: web/src/api/client.ts, web/src/api/types.ts
- Mock Fixtures: web/src/api/mockData.ts (deterministic multi-ticket fixtures matching sample workspaces)
- UI Presentation Components:
  - Navigation: web/src/components/Navbar.tsx
  - Metric Card: web/src/components/KPICard.tsx
  - Cost Distribution: web/src/components/DistributionCard.tsx
  - Ticket Directory: web/src/components/TicketTable.tsx
  - Empty State: web/src/components/EmptyState.tsx
- Views & Shell: web/src/views/FleetOverview.tsx, web/src/App.tsx, web/src/main.tsx, web/src/index.css
- Automated Tests:
  - web/src/domain/formatters.test.ts
  - web/src/api/client.test.ts
  - web/src/views/FleetOverview.test.tsx
- Go Embedding & Server Shell:
  - Static Embedding: internal/adapters/web/embed.go, internal/adapters/web/dist/
  - HTTP Server & SPA Handler: internal/adapters/http/server.go, internal/adapters/http/spa.go
  - CLI Server Entrypoint: cmd/agentic-lens/main.go
  - Root Build Script: Makefile

## 3. Phased Implementation Tasks

### Phase 1: Vite Project Scaffolding and Tooling Configuration
- Task 1.1: Create directory web/ with package.json specifying name agenticlens-web, private flag, dependencies (React 18 or 19, Lucide React or inline SVG geometric icons without emojis), and devDependencies (Vite, TypeScript 5.x+, Vitest, Testing Library, jsdom, ESLint with TypeScript plugins).
- Task 1.2: Define mandatory scripts in web/package.json satisfying the Makefile:
  - dev: starts Vite development server.
  - build: runs typecheck followed by vite build.
  - typecheck: executes tsc --noEmit.
  - lint: executes ESLint across TypeScript source files.
  - test: executes Vitest in run mode.
- Task 1.3: Configure web/tsconfig.json with strict compiler flags: strict: true, noImplicitAny: true, exactOptionalPropertyTypes: true, target: ES2022, module: ESNext, moduleResolution: bundler, jsx: react-jsx.
- Task 1.4: Configure web/vite.config.ts with React plugin, development proxy routing /api to http://localhost:8080, Vitest configuration with jsdom environment, and build output configuration targeting web/dist.
- Task 1.5: Create web/index.html with document title AgenticLens, root mount container, and viewport meta tags.

### Phase 2: Domain Types and Pure Formatting Functions (web/src/domain/)
- Task 2.1: Implement web/src/domain/types.ts:
  - Define strictly typed interfaces matching Go contracts with value semantics and zero any: InvocationTokens, InvocationFact, ToolFact, CodeFacts, AgenticFactLedger, GroupedMetric, TestToCodeRatio, SubagentTurnVelocity, CodeDeliveredStats, ToolExecutionHealth, ContextWindowHealth, WorkflowEfficiency, ModelMetric, RoleMetric, ThreadRoleMetric, ThreadExecutionRecord, AgenticMetricsLedger, FleetSummary, TicketComparisonRow.
  - Define string union types for ContextHealthStatus (HEALTHY, ELEVATED, BLOATED, CRITICAL), AgentRole (orchestrator, planner, coder, tester, reviewer), and canonical display order arrays.
- Task 2.2: Implement web/src/domain/formatters.ts:
  - Implement formatTokensK converting integer token amounts into compact notation with k and M suffixes matching Go FormatTokensK.
  - Implement formatCurrencyUSD formatting numbers as currency with four decimal places for amounts under one dollar and two decimal places otherwise.
  - Implement formatPercentage formatting floats with one decimal place.
  - Implement formatDurationSeconds formatting durations into hours and minutes, or minutes and seconds matching Go ComputeWorkflowEfficiency.
  - Implement formatLeadTime matching Go FormatLeadTime.
  - Implement deriveTicketComparisonRows pure helper mapping ledgers into sorted comparison rows.
  - Defensively guard all ratio and percentage calculations against division by zero.

### Phase 3: Boundary Adapters, Dual-Mode API Client, and Mock Data (web/src/api/)
- Task 3.1: Implement web/src/api/types.ts:
  - Define discriminated union type for asynchronous data fetching state: AsyncState<T> with variants idle, loading, success, and error.
  - Define ApiConfig interface capturing base URL and mock fallback preference.
  - Define structured ApiError interface with error message, optional HTTP status code, and fallback indicator.
- Task 3.2: Implement web/src/api/mockData.ts:
  - Generate deterministic mock data fixtures matching real AgenticLens workspace telemetry (covering tickets such as AL-1, AL-2, PO-1, GEN-1).
  - Provide complete mockFleetSummary, mockTicketComparisonRows, and mockAgenticMetricsLedgers.
  - Include representative variance: healthy, elevated, and bloated context states, multiple roles, and realistic token and cost figures.
  - Strict zero emojis in all mock titles, descriptions, and metadata.
- Task 3.3: Implement web/src/api/client.ts:
  - Implement fetchFleetSummary: makes fetch request to /api/fleet; on network exception or HTTP failure, logs descriptive info and returns mockFleetSummary with indicator isMock: true.
  - Implement fetchWorkspaces: makes fetch request to /api/workspaces; on failure falls back to mockTicketComparisonRows with indicator isMock: true.
  - Implement fetchWorkspaceById: makes fetch request to /api/workspaces/:ticketId; on failure resolves from mock ledgers or returns structured error.
  - Ensure zero unhandled promise rejections and clean typing without any.

### Phase 4: Reusable Presentation Components (web/src/components/)
- Task 4.1: Implement web/src/components/Navbar.tsx:
  - Render top navigation header with application title AgenticLens.
  - Render active view tab indicator for Fleet Overview and inactive link for Ticket Deep Dive.
  - Render data source indicator pill (Live Server or Mock Mode) based on API status.
  - Strict zero emojis across all navigation text and badges.
- Task 4.2: Implement web/src/components/KPICard.tsx:
  - Render card displaying metric label, formatted primary value, optional secondary context (such as percentage of total or change), and tooltip text.
  - Guard against undefined or zero values with fallback dashes.
- Task 4.3: Implement web/src/components/DistributionCard.tsx:
  - Render distribution panel displaying categorized items (roles or models).
  - For each category row: display category name, formatted cost in USD, percentage share, and a proportionate progress fill bar.
  - Sort items descending by cost.
  - Handle zero total cost gracefully by rendering zero percentages without visual distortion.
- Task 4.4: Implement web/src/components/TicketTable.tsx:
  - Render directory table of tickets with columns: Ticket ID, Title, Status, Duration, Cost (USD), Cache Hit Rate, Code Delivered, and SDLC Iterations.
  - Provide search input field filtering rows by Ticket ID and Ticket Title in real time.
  - Provide column header sorting controls for Ticket ID, Cost, Duration, and Iterations.
  - Render colored health status badges (HEALTHY in green, ELEVATED in blue, BLOATED in yellow, CRITICAL in red) using styled text badges with zero emojis.
  - Handle empty search results with descriptive empty-table message.
- Task 4.5: Implement web/src/components/EmptyState.tsx:
  - Render centered card with clear title, descriptive explanation, and optional action button for empty datasets or missing workspaces.

### Phase 5: Views and Application Shell (web/src/views/, web/src/App.tsx)
- Task 5.1: Implement web/src/views/FleetOverview.tsx:
  - Orchestrate data fetching for fleet summary and workspace rows via web/src/api/client.ts.
  - Render loading indicator during initial data retrieval.
  - Render error banner with retry button on unexpected unrecoverable errors.
  - Render top-row KPI cards: Total Cost, Invocations, Total Tokens, Cache Hit Rate, and Active Duration.
  - Render middle-row distribution cards side by side: Role Cost Distribution and Model Cost Distribution.
  - Render bottom-row Ticket Directory Table with search and sorting.
  - Defensively handle completely empty fleet summaries (zero tickets) by showing the EmptyState component.
- Task 5.2: Implement web/src/App.tsx:
  - Mount application shell containing Navbar, main container layout, and FleetOverview view.
  - Maintain top-level state for active route or selected ticket identifier.
- Task 5.3: Implement web/src/main.tsx and styling:
  - Mount React root into root DOM node.
  - Establish CSS styling with modern neutral dark theme, monospace fonts for numbers, responsive grid layouts, and clean borders.

### Phase 6: Unit and Component Testing
- Task 6.1: Create web/src/domain/formatters.test.ts:
  - Test formatTokensK for zero, values under 1000, exact thousands, decimal thousands, exact millions, and decimal millions.
  - Test formatCurrencyUSD for zero, sub-cent micro-costs, normal dollar amounts, and large values.
  - Test formatPercentage for zero, whole numbers, and rounding.
  - Test formatDurationSeconds for zero seconds, sub-minute seconds, minutes and seconds, and hours and minutes.
  - Test formatLeadTime for null (In Progress), seconds, minutes, hours, and days.
  - Assert zero division by zero exceptions.
- Task 6.2: Create web/src/api/client.test.ts:
  - Test fetchFleetSummary returning live data on successful fetch.
  - Test fetchFleetSummary falling back to mock fixtures on network failure.
  - Test fetchWorkspaces fallback behavior and payload structure.
- Task 6.3: Create web/src/views/FleetOverview.test.tsx:
  - Test rendering of KPI cards with mock values.
  - Test rendering of Role and Model distribution tables.
  - Test Ticket Table search filtering by ticket identifier.
  - Test rendering of empty state when ticket list is empty.
  - Assert absence of emojis in rendered DOM text.

### Phase 7: Static Asset Embedding and Build Pipeline Integration
- Task 7.1: Configure the build pipeline so that executing npm run build in web/ synchronizes or outputs the compiled production distribution into internal/adapters/web/dist/, replacing the AL-2 placeholder index.html.
- Task 7.2: Verify that internal/adapters/web/embed.go compiles cleanly with the populated internal/adapters/web/dist/ directory containing assets, javascript bundles, and index.html.
- Task 7.3: Update root Makefile build target to ensure the frontend bundle is built and copied/output to internal/adapters/web/dist/ before invoking go build -o bin/agentic-lens ./cmd/agentic-lens.
- Task 7.4: Verify that internal/adapters/web/embed_test.go passes with the populated distribution assets.

### Phase 8: End-to-End Server Verification and Delivery Sync
- Task 8.1: Verify single-binary server startup using bin/agentic-lens server --port 8080 --workspaces workspaces.
- Task 8.2: Verify that GET / serves the compiled React application index.html with appropriate content-type headers.
- Task 8.3: Verify that compiled javascript and css assets under /assets/ are served with correct MIME types and HTTP 200 responses.
- Task 8.4: Verify that client-side routes (e.g. /overview, /tickets) fall back to index.html while /api/* routes strictly return JSON 404 on missing endpoints.
- Task 8.5: Verify live data ingestion: navigate to http://localhost:8080, confirm that the SPA displays live telemetry from /api/fleet and /api/workspaces, and confirm that the Navbar indicates Live Server.
- Task 8.6: Run full verification suite via make check, make test, and make build.
- Task 8.7: Ingest session telemetry facts using python3 scripts/extract_agentic_metrics.py.

## 4. Key Risks and Defensive Mitigations
- Risk: Divergence between web/dist and internal/adapters/web/dist during local builds.
  Mitigation: The build script in web/package.json or Makefile explicitly copies web/dist contents into internal/adapters/web/dist prior to Go compilation, ensuring embed.FS always packages the current build.
- Risk: Dual-mode API client silently masking backend errors in production.
  Mitigation: The API client tags responses with isMock: true and Navbar visually indicates connection mode. When running from the embedded Go server, the API endpoints /api/fleet and /api/workspaces are served on the same origin, avoiding CORS and proxy mismatch.
- Risk: Single-page application route refresh returning 404 from Go server.
  Mitigation: Verified in AL-2 via SPAHandler in internal/adapters/http/spa.go: any non-file, non-API GET request automatically falls back to index.html with HTTP 200.
- Risk: Accidental emoji usage in icons, labels, or badges.
  Mitigation: Restrict visual indicators to clean SVG geometric shapes and plain text status tags. Execute grep audit for emoji unicode sequences across web/src/.
