# Implementation Plan: FIX-2 Elimination of In-Source Mock Data and Dedicated Test Workspace Harness

## 1. Goal and Iteration Scope

Deliver the complete elimination of hardcoded mock data from production source code and establish a dedicated test workspace directory across backend and frontend:
- Create testdata/workspaces/ containing genuine ticket workspace folders (AL-1, AL-2, AL-3, PO-1, GEN-1) with valid schema v3 agentic_metrics.json ledgers and markdown artifacts (CONTRACT.md, SUMMARY.md).
- Relocate mock fixtures from web/src/api/mockData.ts to web/src/test/fixtures/mockData.ts.
- Delete web/src/api/mockData.ts from production source code.
- Refactor web/src/api/client.ts to remove all mock data imports, remove getFilteredMockLedgers, and eliminate all synthetic fallback returns so errors are cleanly thrown and propagated.
- Update web/src/App.tsx to derive mock mode status based on whether projectConfig.workspaces_dir points to a test data or mock directory.
- Update all 13 Vitest unit test files to import fixtures from web/src/test/fixtures/mockData.ts and update client.test.ts to assert error rejection instead of mock fallbacks.
- Add backend integration tests verifying server behavior when pointed to testdata/workspaces versus production workspaces.
- Verify that npm run build produces a clean bundle in web/dist/ with zero mock data bundled.
- Verify full test suite passes with make check, make test, and make build.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against existing concrete packages, files, and symbols:

### Go Backend (backend/)
- backend/internal/adapters/repository.go:
  - ResolveWorkspacesDir(overridePath string) string: resolves workspace root from argument, environment, config, or directory traversal.
  - ScanWorkspaces(basePath string) domain.ScanReport: traverses workspaces directory to discover and load agentic_metrics.json ledgers.
  - LoadTicketWorkspace(ticketID string, basePath string) (domain.AgenticMetricsLedger, error): loads a specific ticket ledger.
  - ReadTicketArtifacts(ticketID string, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) domain.TicketArtifacts: reads markdown artifacts.
- backend/cmd/agentic-lens/main.go:
  - runServer(args []string): CLI command supporting --workspaces flag to specify workspaces directory.
- backend/cmd/agentic-lens/main_test.go:
  - CLI integration tests validating server and list commands across custom workspaces directories.
- testdata/workspaces/ (new directory):
  - Dedicated test workspace directory containing mock workspaces for AL-1, AL-2, AL-3, PO-1, GEN-1.

### TypeScript Frontend (web/src/)
- web/src/api/mockData.ts:
  - Existing file to be DELETED from production source tree.
- web/src/test/fixtures/mockData.ts (new file):
  - Isolated test fixture file hosting typed fixtures for Vitest tests.
- web/src/api/client.ts:
  - Refactored boundary client with all mock fallback branches eliminated.
  - Exports: fetchFleetSummary, fetchWorkspaceLedgers, fetchWorkspaces, fetchComparisonRows, fetchProgressionLedgers, fetchWorkspaceById, fetchVersion, fetchProjectConfig, fetchWorkspaceEvents, fetchWorkspaceAnomalies, fetchProgressionMetrics.
- web/src/App.tsx:
  - Main application component hydrating projectConfig and computing isMock from workspaces_dir.
- web/src/components/Navbar.tsx:
  - Navigation header rendering Live Server or Mock Mode status pill.
- web/src/api/client.test.ts:
  - Client unit test suite updated to test error rejections.
- Test suites importing fixtures:
  - web/src/api/liveApi.test.ts
  - web/src/components/LiveSessionTree.test.tsx
  - web/src/components/LiveTraceInspector.test.tsx
  - web/src/components/Navbar.test.tsx
  - web/src/domain/formatters.test.ts
  - web/src/domain/progression.test.ts
  - web/src/views/LandingPage.test.tsx
  - web/src/views/FleetOverview.test.tsx
  - web/src/views/TicketDeepDive.test.tsx
  - web/src/views/TicketComparisons.test.tsx
  - web/src/views/MetricProgression.test.tsx
  - web/src/views/LiveObservability.test.tsx

---

## 3. Concrete Step-by-Step Implementation Tasks for Coder

### Phase 1: Test Workspace Filesystem Directory Setup
- Task 1.1: Create testdata/workspaces/ directory in the repository root.
- Task 1.2: Populate test workspaces for tickets AL-1, AL-2, AL-3, PO-1, and GEN-1:
  - Create directory testdata/workspaces/AL-1/ with agentic_metrics.json (schema v3 ledger with invocations, code facts, and totals), CONTRACT.md, and SUMMARY.md.
  - Create directory testdata/workspaces/AL-2/ with agentic_metrics.json (schema v3 ledger with event facts and tool retry anomaly), CONTRACT.md, and SUMMARY.md.
  - Create directory testdata/workspaces/AL-3/ with agentic_metrics.json, CONTRACT.md, and SUMMARY.md.
  - Create directory testdata/workspaces/PO-1/ with agentic_metrics.json (PO documentation ticket), CONTRACT.md, and SUMMARY.md.
  - Create directory testdata/workspaces/GEN-1/ with agentic_metrics.json (chore ticket with context saturation anomaly), CONTRACT.md, and SUMMARY.md.
- Task 1.3: Verify that all agentic_metrics.json files adhere strictly to AgenticFactLedgerDTO schema version 3 so backend/internal/adapters/repository.go unmarshals them without errors.

### Phase 2: Frontend Test Fixtures Relocation
- Task 2.1: Create directory web/src/test/fixtures/.
- Task 2.2: Move and adapt the full fixture definitions from web/src/api/mockData.ts into web/src/test/fixtures/mockData.ts.
- Task 2.3: Ensure web/src/test/fixtures/mockData.ts exports all existing fixture symbols:
  - mockBuildInfo
  - mockProjectConfig
  - mockAL2Anomalies
  - mockGEN1Anomalies
  - mockAL2Events
  - mockFleetSummary
  - mockTicketComparisonRows
  - mockAgenticMetricsLedgers
  - mockWorkspaceDetails
  - mockLiveAnomalyAlerts
  - mockLiveSessionTree
  - mockLiveTraceViews
  - mockLiveTraceView
- Task 2.4: Delete web/src/api/mockData.ts completely from the repository.

### Phase 3: Boundary API Client Refactoring (web/src/api/client.ts)
- Task 3.1: In web/src/api/client.ts, remove the import of mock data symbols from ./mockData.
- Task 3.2: Remove getFilteredMockLedgers helper function completely.
- Task 3.3: Refactor fetchFleetSummary:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok: throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch fleet summary'}`).
  - Catch block: rethrow caught error or throw new Error with failure message. Do not return mockFleetSummary.
- Task 3.4: Refactor fetchWorkspaceLedgers:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok: throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace ledgers'}`).
  - Catch block: throw Error. Do not return mock ledgers.
- Task 3.5: Refactor fetchWorkspaces:
  - On res.ok: parse ledgers, compute rows via deriveTicketComparisonRows, and return { data: rows, isMock: false }.
  - On !res.ok: throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspaces'}`).
  - Catch block: throw Error. Do not return mock rows.
- Task 3.6: Refactor fetchComparisonRows:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok: throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch comparison rows'}`).
  - Catch block: throw Error. Do not return mock rows.
- Task 3.7: Refactor fetchProgressionLedgers:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok: throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch progression ledgers'}`).
  - Catch block: throw Error. Do not return mock ledgers.
- Task 3.8: Refactor fetchWorkspaceById:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok: if res.status === 404, throw new Error(`Workspace ${ticketId} not found`), else throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace'}`).
  - Catch block: throw Error. Do not return mockWorkspaceDetails or mockAgenticMetricsLedgers.
- Task 3.9: Refactor fetchVersion:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok or catch: throw new Error(`Failed to fetch build version info`).
- Task 3.10: Refactor fetchProjectConfig:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok or catch: throw new Error(`Failed to fetch project configuration`).
- Task 3.11: Refactor fetchWorkspaceEvents:
  - On res.ok: parse and return { data: parsedResponse, isMock: false }.
  - On !res.ok: if res.status === 404, throw new Error(`Workspace ${ticketId} not found`), else throw new Error(`HTTP error ${res.status}: ${res.statusText || 'Failed to fetch workspace events'}`).
  - Catch block: throw Error. Do not return mock events.
- Task 3.12: Refactor fetchWorkspaceAnomalies:
  - Calls fetchWorkspaceById; returns { data: detailRes.data.anomalies ?? [], isMock: false }.
  - On failure: catches and throws Error.
- Task 3.13: Refactor fetchProgressionMetrics:
  - On res.ok: return { data: await res.json(), isMock: false }.
  - On !res.ok or catch: throw new Error(`Failed to fetch progression metrics`). Do not fall back to local aggregation over mock ledgers.

### Phase 4: App and Navbar Live/Mock Status Wiring
- Task 4.1: In web/src/App.tsx, inspect projectConfig returned from fetchProjectConfig:
  - Compute isMock as a boolean flag: true if projectConfig.workspaces_dir contains "testdata" or "mock", otherwise false.
  - Pass the computed isMock prop to Navbar.
  - Ensure view callbacks onMockChange do not force mock mode when real server errors happen.
- Task 4.2: In web/src/components/Navbar.tsx:
  - When isMock is true, render the Mock Mode status pill with tooltip title indicating test data workspace execution.
  - When isMock is false, render the Live Server status pill with tooltip title indicating connected live server execution.

### Phase 5: Vitest Test Suites Migration
- Task 5.1: In web/src/api/client.test.ts:
  - Update import path from ./mockData to ../test/fixtures/mockData.
  - Update tests that previously verified silent fallback on network errors, 500 errors, 404 errors, and malformed JSON:
    - Assert that fetchFleetSummary rejects with an Error on network failure and HTTP 500.
    - Assert that fetchWorkspaces rejects with an Error on network failure and HTTP 500.
    - Assert that fetchWorkspaceById rejects with an Error containing "Workspace <ID> not found" on HTTP 404.
    - Assert that fetchWorkspaceLedgers rejects with an Error on network failure and HTTP 500.
    - Assert that fetchWorkspaceEvents rejects with an Error on network failure and HTTP 404.
    - Assert that fetchVersion rejects with an Error on network failure and HTTP 500.
    - Assert that fetchProjectConfig rejects with an Error on network failure and HTTP 500.
    - Assert that fetchProgressionMetrics rejects with an Error on network failure and HTTP 500.
    - Keep successful response tests, verifying isMock is false.
- Task 5.2: Update import paths from ../api/mockData to ../test/fixtures/mockData in:
  - web/src/api/liveApi.test.ts
  - web/src/components/LiveSessionTree.test.tsx
  - web/src/components/LiveTraceInspector.test.tsx
  - web/src/components/Navbar.test.tsx
  - web/src/domain/formatters.test.ts
  - web/src/domain/progression.test.ts
  - web/src/views/LandingPage.test.tsx
  - web/src/views/FleetOverview.test.tsx
  - web/src/views/TicketDeepDive.test.tsx
  - web/src/views/TicketComparisons.test.tsx
  - web/src/views/MetricProgression.test.tsx
  - web/src/views/LiveObservability.test.tsx

### Phase 6: Go Backend Integration Tests
- Task 6.1: In backend/cmd/agentic-lens/main_test.go, add test cases:
  - Verify agentic-lens --workspaces testdata/workspaces --list discovers the test workspaces AL-1, AL-2, AL-3, PO-1, GEN-1.
  - Verify agentic-lens server --workspaces testdata/workspaces responds to GET /api/fleet, GET /api/workspaces, and GET /api/config with workspaces_dir set to testdata/workspaces.
  - Verify that running the server pointing to production workspaces (the default) never scans or loads testdata/workspaces.

### Phase 7: Production Build Verification & Verification Gates
- Task 7.1: Run cd web && npm run build.
- Task 7.2: Verify that web/dist/assets/index-*.js contains zero references to mock fixtures (e.g. grep for unique strings from mock data such as "ses-al20-coder" or "breaching 160k critical ceiling" inside web/dist/assets/).
- Task 7.3: Run full automated checks from repository root: make check, make test, make build.

---

## 4. Verification Criteria for Tester

The tester must systematically execute and record evidence for the following test gates:

### Gate 1: Frontend Unit Test Verification (Vitest)
1. Run cd web && npm test.
2. Verify all test files pass with 0 failures:
   - web/src/api/client.test.ts: all error rejection tests and success tests pass.
   - web/src/api/liveApi.test.ts: passes cleanly.
   - web/src/components/Navbar.test.tsx: passes Live Server and Mock Mode pill checks.
   - web/src/views/LandingPage.test.tsx: passes cleanly.
   - web/src/views/FleetOverview.test.tsx: passes cleanly.
   - web/src/views/TicketDeepDive.test.tsx: passes cleanly.
   - web/src/views/TicketComparisons.test.tsx: passes cleanly.
   - web/src/views/MetricProgression.test.tsx: passes cleanly.
   - web/src/views/LiveObservability.test.tsx: passes cleanly.
   - web/src/domain/formatters.test.ts: passes cleanly.
   - web/src/domain/progression.test.ts: passes cleanly.

### Gate 2: Frontend Typecheck and Linting
1. Run cd web && npm run typecheck.
2. Verify zero TypeScript diagnostic errors and zero any usage.
3. Run cd web && npm run lint.
4. Verify zero ESLint warnings and zero errors.

### Gate 3: Production Bundle Leak Audit
1. Run cd web && npm run build.
2. Inspect web/dist/assets/:
   - Verify web/src/api/mockData.ts is not present.
   - Verify web/src/test/fixtures/mockData.ts is NOT included in web/dist/.
   - Search the bundle output for unique test strings (e.g. grep -rn "Tool retry streak: bash" web/dist/) and verify zero matches found.

### Gate 4: Go Backend Test Suite & Race Detection
1. Run cd backend && go test -race ./...
2. Verify all packages pass with zero race conditions.
3. Run go test -race ./cmd/agentic-lens -run TestServerWithTestdataWorkspaces or new integration test.
4. Verify clean pass.

### Gate 5: End-to-End Server Isolation Check
1. Start Go server with test workspaces:
   ./bin/agentic-lens server --port 18090 --workspaces testdata/workspaces &
2. Query curl http://localhost:18090/api/config:
   - Verify workspaces_dir is testdata/workspaces.
3. Query curl http://localhost:18090/api/workspaces:
   - Verify AL-1, AL-2, AL-3, PO-1, GEN-1 are returned.
4. Stop server.
5. Start Go server with production workspaces:
   ./bin/agentic-lens server --port 18090 --workspaces workspaces &
6. Query curl http://localhost:18090/api/config:
   - Verify workspaces_dir is workspaces.
7. Query curl http://localhost:18090/api/workspaces:
   - Verify only genuine production workspaces are returned; zero testdata items.
8. Stop server.

---

## 5. Audit Checklist for Reviewer

The reviewer must evaluate the changes along the two core review axes:

### Axis 1: Spec Alignment
- [ ] Is web/src/api/mockData.ts completely removed from the production source tree?
- [ ] Are mock fixtures relocated to web/src/test/fixtures/mockData.ts and used strictly by Vitest?
- [ ] Is testdata/workspaces/ created and populated with valid schema v3 workspaces (AL-1, AL-2, AL-3, PO-1, GEN-1)?
- [ ] Are silent mock fallbacks completely eliminated from web/src/api/client.ts?
- [ ] Do API client calls throw or propagate typed Error instances on network failures, 404s, and 500s?
- [ ] When pointed to production workspaces.dir, does the server exclusively scan production workspaces and never load testdata?
- [ ] When pointed to testdata/workspaces, does the server serve mock workspaces through the standard REST endpoints?
- [ ] Does the UI render proper error banners with Retry Connection buttons when the server is unreachable?
- [ ] Does the Navbar display Mock Mode when running against testdata/workspaces and Live Server when running against production workspaces?

### Axis 2: Engineering Standards & Invariants
- [ ] Does the code adhere to Functional Core, Imperative Shell (FCIS)?
- [ ] Are Go domain packages free of I/O, network, and disk dependencies?
- [ ] Is TypeScript strict typing maintained with zero any across all files?
- [ ] Is Go static typing maintained with zero interface{} or any?
- [ ] Is the single binary distribution preserved with web/dist embedded into the Go binary?
- [ ] Is the production bundle completely free of mock data?
- [ ] Are all error messages descriptive and wrapped with domain context?
- [ ] Are there strictly zero emojis across code, comments, docstrings, commit messages, and UI text?
- [ ] Are there zero triple-backtick code blocks in CONTRACT.md and PLAN.md?
