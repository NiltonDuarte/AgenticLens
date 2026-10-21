# Contract: FIX-2 Elimination of In-Source Mock Data and Establishment of Dedicated Test Workspace Harness

## 1. Problem Statement

A major architectural defect and boundary leak exists in the web frontend:
1. Over 2,100 lines of hardcoded synthetic mock fixtures currently reside directly inside the production source code tree at web/src/api/mockData.ts.
2. The boundary HTTP client at web/src/api/client.ts imports from web/src/api/mockData.ts and intercepts network rejections, HTTP errors (such as 404 Not Found or 500 Internal Server Error), and JSON parse exceptions across every endpoint (fetchFleetSummary, fetchWorkspaces, fetchWorkspaceById, fetchWorkspaceLedgers, fetchWorkspaceEvents, fetchWorkspaceAnomalies, fetchVersion, fetchProjectConfig, fetchComparisonRows, fetchProgressionLedgers, fetchProgressionMetrics), silently replacing failed requests with synthetic fixtures tagged with isMock true.
3. Because web/src/api/client.ts imports web/src/api/mockData.ts directly, Vite bundles the entire mock fixture payload into the production JavaScript artifact (web/dist/assets/index-*.js), which is then embedded into the Go executable via embed.FS in backend/internal/adapters/web/embed.go.
4. When running AgenticLens against real production workspaces (workspaces.dir), network or backend errors cause the UI to silently mask real failures and display fictitious mock data (such as synthetic AL-1, AL-2, PO-1, and GEN-1 workspaces), violating operator trust and preventing proper diagnosis.
5. In addition, 13 frontend test suites in web/src/ import test fixtures directly from the production source path web/src/api/mockData.

Ticket FIX-2 eliminates mockData from production source code, removes all silent mock fallback branches from the API client, relocates test data into a dedicated filesystem folder (testdata/workspaces) for server execution, establishes an isolated test fixture harness for Vitest (web/src/test/fixtures/mockData.ts), and enforces strict isolation: when pointing to production workspace-dir, mocked data must never be loaded under any circumstances. When users wish to run with mock data, they must start the server pointing workspaces.dir (or --workspaces <folder>) to the test data folder.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes must strictly respect TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Functional Core (backend/internal/domain/):
   - Zero I/O, zero network, zero filesystem access, zero dependencies on outer adapters.
   - Domain aggregation, rollups, anomaly scoring, and model derivation functions (DeriveMetricsLedger, CalculateFleetSummary, ComputeTotals) operate as pure functions over immutable inputs.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Filesystem scanner (backend/internal/adapters/repository.go) scans exclusively the directory passed to ResolveWorkspacesDir. It never mixes directories or falls back across disparate roots.
   - HTTP server (backend/internal/adapters/http/server.go) serves endpoints (/api/workspaces, /api/fleet, /api/comparison, /api/progression, /api/config) backed entirely by the configured workspaces directory.
3. TypeScript Boundary Shell (web/src/api/):
   - Pure HTTP client: client.ts communicates with backend endpoints via fetch.
   - Zero synthetic fallback: on fetch failure or HTTP error status, client.ts throws or propagates typed Error instances. It never returns mock data.
   - Zero mock imports: client.ts and all production UI files have zero imports of mockData.
4. TypeScript Test Boundary (web/src/test/fixtures/):
   - Test fixtures are placed strictly under web/src/test/fixtures/, completely decoupled from the production application entry point (web/src/main.tsx).
   - The production bundler (Vite) excludes web/src/test/ from the production build output.
5. Filesystem Test Data Harness (testdata/workspaces/):
   - A dedicated test data directory structured as genuine ticket workspaces containing agentic_metrics.json ledgers and documentation artifacts (CONTRACT.md, SUMMARY.md).
   - Used when running the Go server in mock/test mode via --workspaces testdata/workspaces.

### Strong Typing & Zero Any Invariant
1. Zero usage of any or empty interface{} across Go backend and TypeScript frontend.
2. Error states are represented by concrete Error instances in TypeScript and wrapped errors via %w in Go.
3. ApiResponse<T> interface retains typed data of type T and isMock boolean. For all live backend responses, isMock is strictly false.

### Zero Emojis Mandate
Zero emojis permitted across all Go code, TypeScript code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Dedicated Filesystem Test Data Harness:
   - Create root directory testdata/workspaces/ containing genuine workspace subdirectories (AL-1, AL-2, AL-3, PO-1, GEN-1).
   - In each test workspace subdirectory, place the corresponding agentic_metrics.json schema v3 ledger, CONTRACT.md, and SUMMARY.md.
   - Verify that running the Go CLI or server pointing to testdata/workspaces correctly loads and aggregates all test workspaces.
2. Removal of Mock Data from Frontend Source Code:
   - Delete web/src/api/mockData.ts from production source code.
   - Create web/src/test/fixtures/mockData.ts to host deterministic typed test fixtures exclusively for Vitest unit test suites.
   - Verify that web/src/test/ is excluded from Vite production builds and zero mock bytes appear in web/dist/.
3. Boundary API Client Refactoring (web/src/api/client.ts):
   - Remove all imports from mockData.ts.
   - Remove getFilteredMockLedgers and all synthetic in-memory fallback branches.
   - In fetchFleetSummary, fetchWorkspaceLedgers, fetchWorkspaces, fetchComparisonRows, fetchProgressionLedgers, fetchWorkspaceById, fetchVersion, fetchProjectConfig, fetchWorkspaceEvents, fetchWorkspaceAnomalies, and fetchProgressionMetrics:
     - On successful response (res.ok is true): parse JSON payload and return ApiResponse with isMock set to false.
     - On HTTP error response (res.ok is false): throw descriptive Error (e.g. Workspace <ID> not found on 404, or HTTP error <status>: <statusText>).
     - On network rejection or parse exception: propagate or rethrow the Error with descriptive failure context.
4. UI Error Handling and Status Indicator Verification:
   - Verify that LandingPage, FleetOverview, TicketDeepDive, TicketComparisons, and MetricProgression catch thrown API errors and transition cleanly to error states displaying descriptive error banners and Retry Connection buttons.
   - Update App.tsx and Navbar.tsx: derive mock mode status from projectConfig.workspaces_dir (for example, whether the active workspace directory path contains testdata or mock). When pointing to production workspaces, status pill displays Live Server; when pointing to testdata/workspaces, status pill displays Mock Mode.
5. Unit and Integration Test Updates:
   - Update all 13 Vitest test suites (web/src/api/client.test.ts, web/src/api/liveApi.test.ts, web/src/components/LiveSessionTree.test.tsx, web/src/components/LiveTraceInspector.test.tsx, web/src/components/Navbar.test.tsx, web/src/domain/formatters.test.ts, web/src/domain/progression.test.ts, web/src/views/LandingPage.test.tsx, web/src/views/FleetOverview.test.tsx, web/src/views/TicketDeepDive.test.tsx, web/src/views/TicketComparisons.test.tsx, web/src/views/MetricProgression.test.tsx, web/src/views/LiveObservability.test.tsx) to import fixtures from web/src/test/fixtures/mockData.ts.
   - In web/src/api/client.test.ts, replace obsolete fallback assertions with error rejection assertions (verifying that network errors, 500 errors, and 404 errors reject with expected Error instances).
   - Add Go integration test in backend/cmd/agentic-lens/main_test.go verifying that starting the server with --workspaces testdata/workspaces serves the mock test data, whereas starting without --workspaces (or pointing to production workspaces) strictly serves production data and never touches testdata.

### Out of Scope
1. Modifying live telemetry streaming protocols or SQLite hooks in backend/internal/adapters/live/.
2. Modifying Git commit diff analysis, inference rules, or code classification patterns in backend/internal/adapters/git/.
3. Changing domain rollup calculation formulas in backend/internal/domain/aggregations.go.
4. Altering the schema version or migration engine in backend/internal/domain/migration/.

---

## 4. Architectural Contracts and Interface Specifications

### Server Workspaces Resolution and Strict Isolation Contract
1. Default / Production Execution:
   - When the server runs with default configuration or --workspaces workspaces, ResolveWorkspacesDir resolves to the production workspaces directory.
   - The scanner scans exclusively files within that directory.
   - Mocked data in testdata/workspaces is never touched, loaded, or exposed.
2. Mock / Test Data Execution:
   - When running against mocked data, the server must be launched pointing workspaces.dir to the test data folder (e.g. ./bin/agentic-lens server --workspaces testdata/workspaces or setting workspaces.dir: testdata/workspaces in .agenticlens-config.yaml).
   - The scanner discovers all valid workspaces under testdata/workspaces/ and calculates all domain aggregates, fleet summaries, comparison rows, and progression buckets dynamically through the Go domain engine.
   - GET /api/config returns workspaces_dir reflecting the configured test data path (e.g. testdata/workspaces).

### Filesystem Test Data Structure Contract: testdata/workspaces/
The test data folder must mirror genuine workspace layouts:
1. Root Folder: testdata/workspaces/
2. Ticket Folders:
   - testdata/workspaces/AL-1/
   - testdata/workspaces/AL-2/
   - testdata/workspaces/AL-3/
   - testdata/workspaces/PO-1/
   - testdata/workspaces/GEN-1/
3. Artifacts per Ticket:
   - agentic_metrics.json: Valid schema v3 JSON ledger matching AgenticFactLedgerDTO.
   - CONTRACT.md: Markdown technical contract artifact.
   - SUMMARY.md: Markdown work summary artifact.

### Frontend Test Fixture Harness Contract: web/src/test/fixtures/mockData.ts
1. Location: web/src/test/fixtures/mockData.ts
2. Exported Symbols (mirroring previous exports for test compatibility):
   - mockBuildInfo: BuildInfo
   - mockProjectConfig: ProjectConfigResponse
   - mockAL2Anomalies: readonly Anomaly[]
   - mockGEN1Anomalies: readonly Anomaly[]
   - mockAL2Events: readonly EventFact[]
   - mockFleetSummary: FleetSummary
   - mockTicketComparisonRows: readonly TicketComparisonRow[]
   - mockAgenticMetricsLedgers: Readonly<Record<string, AgenticMetricsLedger>>
   - mockWorkspaceDetails: Readonly<Record<string, WorkspaceDetail>>
   - mockLiveAnomalyAlerts: readonly LiveAnomalyAlert[]
   - mockLiveSessionTree: LiveSessionTree
   - mockLiveTraceViews: Readonly<Record<string, LiveTraceView>>
   - mockLiveTraceView: LiveTraceView
3. Build Isolation Guarantee:
   - File is imported exclusively by *.test.ts and *.test.tsx files.
   - File is excluded from the production bundle graph rooted at web/src/main.tsx.

### Boundary API Client Contract: web/src/api/client.ts
1. General Return Signature:
   - Every fetch function returns Promise<ApiResponse<T>> where ApiResponse<T> has data: T and isMock: boolean.
   - On successful HTTP response: data contains parsed JSON, isMock is false.
2. Error Handling Policy:
   - No catch block may return synthetic fallback fixtures.
   - If response status is 404: throw Error(`Workspace ${ticketId} not found`) or Error(`Resource not found: ${url}`).
   - If response status is non-2xx: throw Error(`HTTP error ${response.status}: ${response.statusText}`).
   - If response JSON parsing fails: throw Error(`Failed to parse JSON response from ${url}`).
   - If fetch rejects (network drop, connection refused, DNS error): throw Error(`Network error: ${error.message}`).

### UI Live Server vs. Mock Mode Status Contract
1. App Component (web/src/App.tsx):
   - Queries fetchProjectConfig on initialization.
   - Derives isMock: boolean based on whether projectConfig.workspaces_dir contains "testdata" or "mock".
   - Passes derived isMock to Navbar.
2. Navbar Component (web/src/components/Navbar.tsx):
   - When isMock is true: displays "Mock Mode" status pill with title "Running with test data folder (<workspaces_dir>)".
   - When isMock is false: displays "Live Server" status pill with title "Connected to live AgenticLens Go HTTP server".
3. Error Presentation:
   - When API calls throw, the calling views (LandingPage, FleetOverview, TicketDeepDive, TicketComparisons, MetricProgression) transition to their error states, rendering the error banner with an actionable Retry Connection button.
