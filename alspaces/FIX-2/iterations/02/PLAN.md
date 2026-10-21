# Implementation Plan: FIX-2 Iteration 02 - Mandatory Workspaces Configuration Enforcement, Complete Elimination of Mock Mode Branching, and Strict Zero-Guess Presentation

## 1. Goal and Iteration Scope

Iteration 02 addresses critical user steering feedback regarding workspace configuration enforcement and mock indicator clean-up:
- Enforce mandatory workspace directory configuration across the Go backend: when running the server or workspace-dependent CLI commands, workspaces directory must be explicitly configured in project configuration (workspaces.dir) or provided via the --workspaces flag. The server cannot start without this information. Zero guesses and zero silent fallbacks to "workspaces".
- Fail fast on server startup: if neither --workspaces flag nor projCfg.Workspaces.Dir is set (or if empty/whitespace), runServer must fail immediately with exit code 1 and output the exact error:
  Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information.
- Validate directory existence: if the specified workspaces directory does not exist on disk, runServer must fail immediately with exit code 1 and output the exact error:
  Error: workspaces directory %q does not exist. Server cannot start.
- Eliminate guessing in the React frontend: in Navbar.tsx and App.tsx, do NOT default or guess "workspaces". The active workspaces directory is obtained strictly from projectConfig.workspaces_dir returned by /api/config. If projectConfig is loading or unavailable, display a loading placeholder. Never guess or substitute a default.
- Complete the elimination of isMock and onMockChange across the entire frontend: remove isMock from ApiResponse, AsyncState, ApiConfig, ApiError, App.tsx, Navbar.tsx, and all view components. Remove onMockChange callbacks and mock tracking refs across all views.
- Synchronize all frontend and backend tests: ensure all test servers and configs explicitly provide workspaces_dir, assert that server startup fails when workspaces.dir is missing/empty or nonexistent, and update Navbar.test.tsx to assert that Navbar renders projectConfig.workspaces_dir as provided and does NOT guess "workspaces" when missing or loading.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against existing concrete packages, files, interfaces, and test suites:

### Go Backend (backend/)
- backend/cmd/agentic-lens/main.go:
  - runServer(args []string): CLI entrypoint where flags are parsed, project configuration is loaded, workspaces directory presence and existence are validated, and the HTTP server is launched.
  - Root command execution (flag.Parse()): handles --list, --ticket-id, --anomalies, and metrics reporting requiring a valid workspaces directory.
- backend/cmd/agentic-lens/main_test.go:
  - CLI integration tests validating server startup failures when workspaces configuration is omitted or points to nonexistent paths, and testing valid server startup with explicit workspace directories.
- backend/internal/adapters/repository.go:
  - ResolveWorkspacesDir(overridePath string) string: directory resolver. Must be updated to eliminate fallback to "workspaces". If neither overridePath nor cfg.Workspaces.Dir is provided, returns empty string.
  - ScanWorkspaces(basePath string) domain.ScanReport: scans workspaces directory; returns clear error report if resolved directory is empty or missing.
  - LoadTicketWorkspace(ticketID string, basePath string) (domain.AgenticMetricsLedger, error): loads a specific ticket ledger; returns error if directory is empty or missing.
  - ReadTicketArtifacts(ticketID string, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) domain.TicketArtifacts: reads ticket artifacts; returns empty artifacts if directory is empty or missing.
- backend/internal/adapters/repository_test.go:
  - Unit tests for ResolveWorkspacesDir and ScanWorkspaces, asserting that missing configuration returns empty string without defaulting to "workspaces".
- backend/internal/adapters/http/config_handler.go:
  - ConfigHandler(cfg domain.ProjectConfig, workspacesDir string) http.HandlerFunc: endpoint serving GET /api/config with sanitized project metadata and workspaces_dir.
- backend/internal/adapters/http/server.go:
  - NewServer and NewServerWithConfig: constructor initializing router, SPA static handler, and API routes.
- backend/internal/adapters/http/server_test.go:
  - HTTP integration tests verifying /api/config responses across explicit configurations.

### TypeScript API Client & Types (web/src/api/)
- web/src/api/client.ts:
  - ApiResponse<T>: interface defining readonly data: T (isMock removed).
  - Exported functions: fetchFleetSummary, fetchWorkspaceLedgers, fetchWorkspaces, fetchComparisonRows, fetchProgressionLedgers, fetchWorkspaceById, fetchVersion, fetchProjectConfig, fetchWorkspaceEvents, fetchWorkspaceAnomalies, and fetchProgressionMetrics.
- web/src/api/types.ts:
  - AsyncState<T>: discriminated union with isMock removed from success state.
  - ApiConfig: interface with preferMock removed.
  - ApiError: interface with isFallback removed.
- web/src/api/client.test.ts:
  - Client test suite with all isMock assertions removed.

### TypeScript Components & Layout (web/src/components/ & web/src/)
- web/src/App.tsx:
  - Application root component. Removes isMock calculation, removes handleMockChange, and passes projectConfig to Navbar without any fallback or default.
- web/src/components/Navbar.tsx:
  - Navigation header. Removes isMock prop, removes Live Server and Mock Mode status pills, extracts workspaces_dir strictly from projectConfig, renders a loading placeholder when not yet loaded, and renders the exact workspace directory when available without guessing "workspaces".
- web/src/index.css:
  - Stylesheet replacing .status-pill.live and .status-pill.mock with .status-pill.workspace-pill and loading state styling.

### TypeScript Views (web/src/views/)
- web/src/views/LandingPage.tsx:
  - Removes onMockChange from LandingPageProps interface.
  - Removes onMockChange from component props destructuring.
  - In loadData, remove isAnyMock check and remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- web/src/views/FleetOverview.tsx:
  - Removes onMockChange from FleetOverviewProps interface.
  - Removes onMockChange from component props destructuring.
  - In loadData, remove isAnyMock check and remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- web/src/views/TicketDeepDive.tsx:
  - Removes onMockChange from TicketDeepDiveProps interface.
  - Removes onMockChange from component props destructuring.
  - Removes isMockFleetRef ref object.
  - In loadData, remove isMockFleet and isMockEvents tracking variables and remove all onMockChange invocations.
  - In useCallback dependency array, remove onMockChange.
- web/src/views/MetricProgression.tsx:
  - Removes onMockChange from MetricProgressionProps interface.
  - Removes onMockChange from component props destructuring.
  - In loadLedgers, remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- web/src/views/TicketComparisons.tsx:
  - Removes onMockChange from TicketComparisonsProps interface.
  - Removes onMockChange from component props destructuring.
  - In loadLedgers, remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.

### Frontend Test Suites (web/src/**/*.test.ts, web/src/**/*.test.tsx)
- web/src/components/Navbar.test.tsx:
  - Tests rendering of projectConfig.workspaces_dir, asserts loading placeholder when projectConfig is missing/loading, and verifies absence of "workspaces" guessing.
- web/src/components/Components.test.tsx:
  - Updates Navbar rendering tests to verify workspace pill and absence of guessing.
- web/src/views/LandingPage.test.tsx:
  - Removes isMock from resolved mocks and removes onMockChange test case.
- web/src/views/FleetOverview.test.tsx:
  - Removes isMock from resolved mocks and removes onMockChange test case.
- web/src/views/TicketDeepDive.test.tsx:
  - Removes isMock flags from mock resolved returns.
- web/src/views/MetricProgression.test.tsx:
  - Removes isMock flags from mock resolved returns.
- web/src/views/TicketComparisons.test.tsx:
  - Removes isMock flags from mock resolved returns.

---

## 3. Concrete Step-by-Step Implementation Tasks for Coder

### Phase 1: Go Backend Mandatory Workspaces Configuration & Startup Validation
- Task 1.1: Refactor ResolveWorkspacesDir in backend/internal/adapters/repository.go:
  - Inspect overridePath parameter: if strings.TrimSpace(overridePath) is non-empty, clean and resolve it.
  - If overridePath is empty: load project configuration via config.Load("").
  - Inspect cfg.Workspaces.Dir: if strings.TrimSpace(cfg.Workspaces.Dir) is empty, do NOT default to "workspaces". Return empty string immediately.
  - If cfg.Workspaces.Dir is non-empty: resolve candidate path relative to configuration file directory, container mount (/app/<dir>), or current working directory.
  - Return resolved path if directory exists on disk, or return the configured candidate path so callers can validate existence. Never guess "workspaces".
- Task 1.2: Update ScanWorkspaces, LoadTicketWorkspace, and ReadTicketArtifacts in backend/internal/adapters/repository.go:
  - In ScanWorkspaces: if ResolveWorkspacesDir returns empty string, return a domain.ScanReport with empty ledgers and an error string: "Workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag."
  - In LoadTicketWorkspace: if ResolveWorkspacesDir returns empty string, return domain.AgenticMetricsLedger{} and an error: fmt.Errorf("workspaces directory is not configured").
  - In ReadTicketArtifacts: if ResolveWorkspacesDir returns empty string, return domain.TicketArtifacts{Artifacts: make(map[string]domain.ArtifactContent)}.
- Task 1.3: Update runServer in backend/cmd/agentic-lens/main.go:
  - Parse server flags (serverFlags.Parse(args)).
  - Load project configuration first via adaptersconfig.Load(""). If loading fails (ErrConfigNotFound or invalid configuration), output error to os.Stderr and exit with code 1.
  - Determine raw workspaces directory setting:
    - Check flag value: strings.TrimSpace(*workspacesFlag).
    - If flag value is empty, check configuration value: strings.TrimSpace(projCfg.Workspaces.Dir).
    - If both flag and configuration are empty or whitespace:
      Output to os.Stderr: "Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information."
      Call os.Exit(1).
  - Resolve the workspaces directory path via adapters.ResolveWorkspacesDir(rawWorkspacesDir).
  - Validate directory existence on disk:
    - Perform fi, err := os.Stat(workspacesDir).
    - If err != nil or !fi.IsDir():
      Output to os.Stderr: fmt.Sprintf("Error: workspaces directory %q does not exist. Server cannot start.\n", rawWorkspacesDir).
      Call os.Exit(1).
  - Pass the validated workspacesDir into adaptershttp.NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService...).
- Task 1.4: Update root CLI command execution in backend/cmd/agentic-lens/main.go:
  - For commands that require workspaces (--list, --ticket-id, --anomalies, and default ledger summary):
    - Load project configuration via adaptersconfig.Load("").
    - If both *workspacesFlag and projCfg.Workspaces.Dir are empty/whitespace, output error to os.Stderr indicating workspaces directory must be configured or provided via --workspaces flag, and exit with code 1.
    - If the resolved directory does not exist on disk, output error to os.Stderr that the workspaces directory does not exist, and exit with code 1.
- Task 1.5: Update ConfigHandler in backend/internal/adapters/http/config_handler.go:
  - Ensure WorkspacesDir in ProjectConfigDTO strictly reflects the supplied workspacesDir parameter or trimmed cfg.Workspaces.Dir.
  - Do NOT supply or fallback to "workspaces" when empty.

### Phase 2: Frontend API Client & Async Types Clean-Up (Eliminating isMock)
- Task 2.1: In web/src/api/types.ts:
  - In AsyncState<T>, remove readonly isMock: boolean from the success variant.
  - In ApiConfig, remove readonly preferMock: boolean.
  - In ApiError, remove readonly isFallback: boolean.
- Task 2.2: In web/src/api/client.ts:
  - In ApiResponse<T>, remove readonly isMock: boolean, leaving readonly data: T.
  - In fetchFleetSummary, return { data }.
  - In fetchWorkspaceLedgers, return { data: ledgers }.
  - In fetchWorkspaces, return { data: rows }.
  - In fetchComparisonRows, return { data: rows }.
  - In fetchProgressionLedgers, return { data: ledgers }.
  - In fetchWorkspaceById, return { data }.
  - In fetchVersion, return { data }.
  - In fetchProjectConfig, return { data }.
  - In fetchWorkspaceEvents, return { data: resp } or default unhydrated events object without isMock.
  - In fetchWorkspaceAnomalies, return { data: detailRes.data.anomalies ?? [] }.
  - In fetchProgressionMetrics, return { data }.

### Phase 3: View Components Clean-Up (Eliminating onMockChange and Mock State)
- Task 3.1: In web/src/views/LandingPage.tsx:
  - Remove onMockChange from LandingPageProps interface.
  - Remove onMockChange from component props destructuring.
  - In loadData, remove isAnyMock check and remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- Task 3.2: In web/src/views/FleetOverview.tsx:
  - Remove onMockChange from FleetOverviewProps interface.
  - Remove onMockChange from component props destructuring.
  - In loadData, remove isAnyMock check and remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- Task 3.3: In web/src/views/TicketDeepDive.tsx:
  - Remove onMockChange from TicketDeepDiveProps interface.
  - Remove onMockChange from component props destructuring.
  - Remove isMockFleetRef ref object.
  - In loadData, remove isMockFleet and isMockEvents tracking variables and remove all onMockChange invocations.
  - In useCallback dependency array, remove onMockChange.
- Task 3.4: In web/src/views/MetricProgression.tsx:
  - Remove onMockChange from MetricProgressionProps interface.
  - Remove onMockChange from component props destructuring.
  - In loadLedgers, remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.
- Task 3.5: In web/src/views/TicketComparisons.tsx:
  - Remove onMockChange from TicketComparisonsProps interface.
  - Remove onMockChange from component props destructuring.
  - In loadLedgers, remove onMockChange invocation.
  - In useCallback dependency array, remove onMockChange.

### Phase 4: Application Shell & Navbar Workspaces Directory Presentation (Zero Guessing)
- Task 4.1: In web/src/App.tsx:
  - Remove isMock useMemo hook completely.
  - Remove handleMockChange callback function completely.
  - Remove isMock prop passed to Navbar.
  - Remove onMockChange={handleMockChange} prop passed to LandingPage, FleetOverview, TicketDeepDive, MetricProgression, and TicketComparisons.
  - Pass projectConfig={projectConfig ?? undefined} to Navbar.
  - Ensure zero guessing or defaulting to "workspaces" anywhere in App.tsx.
- Task 4.2: In web/src/components/Navbar.tsx:
  - In NavbarProps, remove readonly isMock: boolean.
  - In component parameter destructuring, remove isMock.
  - Determine active workspaces directory strictly from projectConfig?.workspaces_dir or internalConfig?.workspaces_dir. Clean with trim.
  - Under NO circumstances default or guess "workspaces".
  - In the navbar-actions container:
    - If active workspaces directory is non-empty: render a workspace directory badge/pill:
      An element with class status-pill workspace-pill, title describing the active workspaces directory, a status-dot child element, and the exact workspace directory text.
    - If active workspaces directory is empty, null, or undefined (e.g. while configuration is loading or pending): render a loading placeholder element:
      An element with class status-pill workspace-pill loading, title indicating loading workspaces configuration, a status-dot loading child element, and text "Loading...".
    - Completely eliminate the isMock conditional check and remove Live Server / Mock Mode status pills.
- Task 4.3: In web/src/index.css:
  - Replace .status-pill.live and .status-pill.mock styles with .status-pill.workspace-pill styles.
  - Style .status-pill.workspace-pill with monospace font family for directory paths, subtle surface background, neutral border, and overflow ellipsis truncation.
  - Add styles for .status-pill.workspace-pill.loading with subtle pulsating or muted dot for the loading state.

### Phase 5: Test Suite Synchronization (Backend and Frontend)
- Task 5.1: In backend/cmd/agentic-lens/main_test.go:
  - Add test case TestServer_FailsWithoutWorkspacesConfig:
    - Run server command in a temporary directory with a configuration file where workspaces.dir is empty or omitted, without passing the --workspaces flag.
    - Assert that the command exits with non-zero exit code (1).
    - Assert that standard error contains:
      "Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information."
  - Add test case TestServer_FailsWhenWorkspacesDirDoesNotExist:
    - Run server command with --workspaces /nonexistent/workspaces/dir/xyz.
    - Assert that the command exits with non-zero exit code (1).
    - Assert that standard error contains:
      "Error: workspaces directory \"/nonexistent/workspaces/dir/xyz\" does not exist. Server cannot start."
  - Verify all existing server tests in main_test.go explicitly supply a valid workspace directory via --workspaces or project configuration.
- Task 5.2: In backend/internal/adapters/repository_test.go:
  - In TestResolveWorkspacesDir, add sub-test asserting that when overridePath is empty and configuration has no workspaces.dir, ResolveWorkspacesDir returns empty string "" and does NOT return "workspaces".
  - Verify existing sub-tests that configure custom or yaml workspaces directories continue to pass.
- Task 5.3: In web/src/components/Navbar.test.tsx:
  - Remove isMock prop from all render(<Navbar ...>) and rerender(<Navbar ...>) calls across all tests.
  - Add test: asserts that Navbar renders the active workspace directory text when projectConfig with workspaces_dir is provided (e.g. "testdata/workspaces" or "custom/workspaces").
  - Add test: asserts that Navbar renders the loading placeholder ("Loading...") and does NOT render or guess "workspaces" when projectConfig is undefined or workspaces_dir is empty.
  - Remove obsolete tests: "renders Live Server status pill when isMock is false" and "renders Mock Mode status pill when isMock is true".
  - Update zero emojis test to verify clean text across the workspace directory pill and loading state.
- Task 5.4: In web/src/components/Components.test.tsx:
  - Remove isMock prop from Navbar renders.
  - Update test assertions to verify workspace pill rendering and verify absence of Live Server and Mock Mode.
  - Assert that when configuration is missing or empty, Navbar displays loading state and does NOT default to "workspaces".
- Task 5.5: In web/src/api/client.test.ts:
  - Remove all 14 expect(res.isMock).toBe(false) assertions.
  - Update test titles that mention isMock false to clean descriptive titles.
- Task 5.6: In view test suites (LandingPage.test.tsx, FleetOverview.test.tsx, TicketDeepDive.test.tsx, MetricProgression.test.tsx, TicketComparisons.test.tsx):
  - In LandingPage.test.tsx: remove isMock from mock resolved values; delete test "notifies onMockChange callback with true when either endpoint falls back to mock".
  - In FleetOverview.test.tsx: remove isMock from mock resolved values; delete test "notifies onMockChange callback with true if either summary or workspaces is mock".
  - In TicketDeepDive.test.tsx: remove isMock properties from mock resolved values.
  - In MetricProgression.test.tsx: remove isMock properties from mock resolved values.
  - In TicketComparisons.test.tsx: remove isMock properties from mock resolved values.

### Phase 6: Verification Gates
- Task 6.1: Run frontend unit tests: cd web && npm test.
- Task 6.2: Run frontend typecheck: cd web && npm run typecheck.
- Task 6.3: Run frontend linter: cd web && npm run lint.
- Task 6.4: Run backend unit and integration test suite: cd backend && go test -race ./...
- Task 6.5: Run full repository verification: make check, make test, and make build.

---

## 4. Verification Criteria for Tester

The tester must systematically execute and record evidence for the following test gates:

### Gate 1: Server Startup Failure on Missing Workspaces Directory
1. Prepare a project directory with .agenticlens-config.yaml omitting workspaces.dir (or setting dir: "").
2. Execute: ./bin/agentic-lens server --port 19123
3. Verify that the process immediately exits with exit code 1.
4. Verify standard error output matches exactly:
   Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information.

### Gate 2: Server Startup Failure on Nonexistent Workspaces Directory
1. Execute: ./bin/agentic-lens server --port 19123 --workspaces /path/to/definitely/nonexistent/directory
2. Verify that the process immediately exits with exit code 1.
3. Verify standard error output matches:
   Error: workspaces directory "/path/to/definitely/nonexistent/directory" does not exist. Server cannot start.

### Gate 3: Navbar Rendering and Zero-Guess Workspaces Directory Display
1. In web/src/components/Navbar.test.tsx:
   - When projectConfig is supplied with workspaces_dir: "testdata/workspaces", verify the DOM contains an element with text "testdata/workspaces".
   - When projectConfig is supplied with workspaces_dir: "workspaces", verify the DOM contains an element with text "workspaces".
   - When projectConfig is omitted or has empty workspaces_dir, verify the DOM displays the loading placeholder "Loading..." and verify queryByText("workspaces") is null.
   - Verify that neither "Live Server" nor "Mock Mode" exists in the rendered output.
2. In web/src/components/Components.test.tsx:
   - Verify identical zero-guess behavior.

### Gate 4: Complete Removal of isMock and onMockChange across Frontend
1. Search production frontend code in web/src/:
   - Execute grep search for isMock in web/src/ excluding web/src/test/.
   - Expected result: zero matches found.
   - Execute grep search for onMockChange in web/src/ excluding web/src/test/.
   - Expected result: zero matches found.
2. Execute production build: cd web && npm run build.
   - Inspect build bundle in web/dist/assets/:
   - Search bundle for isMock: zero occurrences.
   - Search bundle for Live Server: zero occurrences.
   - Search bundle for Mock Mode: zero occurrences.

### Gate 5: Backend & Frontend Test Suites Clean Pass
1. Run cd web && npm test:
   - All tests pass with zero failures.
2. Run cd backend && go test -race ./...:
   - All tests pass with zero failures and zero race condition warnings.

### Gate 6: Typecheck and Linter Verification
1. Run cd web && npm run typecheck:
   - Zero TypeScript diagnostic errors; zero any types.
2. Run cd web && npm run lint:
   - Zero ESLint errors and zero warnings.

---

## 5. Audit Checklist for Reviewer

The reviewer must evaluate the implementation against the following checklist:

### Axis 1: Spec Alignment
- [ ] Does runServer in backend/cmd/agentic-lens/main.go refuse to start if neither --workspaces flag nor projCfg.Workspaces.Dir is set?
- [ ] Does runServer fail with exit code 1 and output the exact error message: "Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information."?
- [ ] Does runServer fail with exit code 1 and output "Error: workspaces directory %q does not exist. Server cannot start." if the directory does not exist on disk?
- [ ] Does ResolveWorkspacesDir in backend/internal/adapters/repository.go return empty string instead of guessing "workspaces" when neither overridePath nor config is provided?
- [ ] Does Navbar.tsx render the active workspace directory strictly from projectConfig.workspaces_dir without guessing "workspaces"?
- [ ] When projectConfig is loading or missing workspaces_dir, does Navbar.tsx display a loading placeholder instead of guessing a default?
- [ ] Is isMock completely removed from ApiResponse<T>, AsyncState<T>, ApiConfig, ApiError, client.ts, App.tsx, Navbar.tsx, and all views?
- [ ] Is onMockChange completely removed from all view interfaces, parameters, and callback invocations?
- [ ] Are all frontend and backend tests updated to explicitly supply workspaces_dir and assert failure when missing?
- [ ] Does Navbar.test.tsx verify that Navbar does NOT guess "workspaces" when missing or loading?

### Axis 2: Engineering Standards & Invariants
- [ ] Does the Go backend maintain Functional Core, Imperative Shell (FCIS) discipline, keeping pure functions free of I/O?
- [ ] Does the TypeScript codebase strictly enforce typing with zero usage of any?
- [ ] Does the Go codebase strictly enforce typing with zero usage of interface{} or any?
- [ ] Are there strictly zero code blocks in CONTRACT.md and PLAN.md?
- [ ] Are there strictly zero emojis across code, comments, docstrings, commit messages, and UI text?
- [ ] Do all automated verification gates pass cleanly (make check, make test, make build)?
