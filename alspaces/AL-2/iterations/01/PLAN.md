# Implementation Plan: AL-2 Go HTTP Server, Embedded Static FS, and Telemetry REST API

## 1. Goal and Iteration Scope
The goal of iteration 01 is to implement the native Go HTTP server, embed static baseline assets, provide the REST API for telemetry data, and wire the `server` subcommand into the CLI.

Upon completion of this plan, running `agentic-lens server --port 8080` will serve the REST endpoints (`/api/workspaces`, `/api/workspaces/{ticketId}`, `/api/fleet`), handle CORS requests for local frontend development, and serve embedded static assets with SPA fallback routing for any browser navigation.

## 2. Grounded Reference Architecture
The implementation builds directly on top of the domain and adapter foundation established in AL-1:
- Existing Domain Models: `internal/domain/models.go` (models `AgenticMetricsLedger`, `FleetSummary`, `TicketArtifacts`, `ScanReport`)
- Existing Pure Aggregations: `internal/domain/aggregations.go` (functions `CalculateFleetSummary`, `FilterLedgersByDateRange`, `ResolveDatePreset`)
- Existing Workspace Repository: `internal/adapters/repository.go` (functions `ResolveWorkspacesDir`, `ScanWorkspaces`, `ReadTicketArtifacts`)
- New Domain Additions: `internal/domain/enums.go` (typed enums `TimeWindow`, `LedgerSortField`) and `internal/domain/models.go` (`WorkspaceDetail`)
- New Web Asset Adapter: `internal/adapters/web/embed.go` (embedding `dist/index.html`)
- New HTTP Adapter Package: `internal/adapters/http` (files `server.go`, `handlers.go`, `middleware.go`, `spa.go`, `server_test.go`)
- CLI Subcommand Integration: `cmd/agentic-lens/main.go`

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Models and Filter Predicates (`internal/domain`)
- Task 1.1: Update `internal/domain/enums.go`:
  - Define `TimeWindow` type as a typed string enum with constants: `TimeWindowAll` ("all"), `TimeWindow7D` ("7d"), `TimeWindow30D` ("30d"), and `TimeWindow90D` ("90d").
  - Define `LedgerSortField` type as a typed string enum with constants: `SortFieldUpdatedAt` ("updated_at"), `SortFieldCost` ("cost"), `SortFieldDuration` ("duration"), and `SortFieldInvocations` ("invocations").
  - Implement helper functions `ParseTimeWindow` and `ParseLedgerSortField` validating input strings and returning typed constants with fallback defaults.
- Task 1.2: Update `internal/domain/models.go`:
  - Define `WorkspaceDetail` struct containing `Ledger` (`AgenticMetricsLedger`) and `Artifacts` (`TicketArtifacts`) with JSON field tags.
- Task 1.3: Update `internal/domain/aggregations.go`:
  - Implement pure function `FilterLedgersByTimeWindow`: accepts a slice of `AgenticMetricsLedger`, a `TimeWindow`, and a reference `time.Time` timestamp. Calculates boundary dates (7 days, 30 days, 90 days before reference) and filters ledgers using `GetTicketEffectiveDate`. For `TimeWindowAll`, returns a clone of all ledgers.
  - Implement pure function `SortLedgers`: accepts a slice of `AgenticMetricsLedger`, a `LedgerSortField`, and a boolean `ascending` flag. Performs deterministic sorting on copy of slice according to targeted field:
    - `SortFieldUpdatedAt`: compares `UpdatedAt` strings.
    - `SortFieldCost`: compares `Totals.CostUSD`.
    - `SortFieldDuration`: compares `WorkflowEfficiency.ActiveDurationSeconds`.
    - `SortFieldInvocations`: compares `Totals.Invocations`.
  - Add unit tests in `internal/domain/aggregations_test.go` covering all time window boundaries and sort dimensions with zero mocks.

### Phase 2: Embedded Static Assets Container (`internal/adapters/web`)
- Task 2.1: Create package directory `internal/adapters/web` and subfolder `internal/adapters/web/dist`.
- Task 2.2: Add a minimal, valid `dist/index.html` containing HTML5 boilerplate, page title "AgenticLens", and root container element to ensure successful Go compilation.
- Task 2.3: Implement `internal/adapters/web/embed.go`:
  - Declare package `web`.
  - Declare embedded filesystem variable using `go:embed all:dist`.
  - Implement accessor function `GetFS` returning an `fs.FS` sub-filesystem rooted at `dist` using `fs.Sub`.

### Phase 3: HTTP Adapter, Handlers, Middleware, and Router (`internal/adapters/http`)
- Task 3.1: Define structured HTTP response schemas in `internal/adapters/http/dto.go`:
  - Define `ErrorResponse` struct with fields `Error` (string), `Message` (string), and `StatusCode` (int).
  - Implement helper `WriteJSON` to serialize payloads with appropriate `Content-Type: application/json` headers and status codes.
  - Implement helper `WriteError` to serialize `ErrorResponse` with appropriate HTTP status code.
- Task 3.2: Implement middleware in `internal/adapters/http/middleware.go`:
  - Implement `CORSMiddleware`: adds `Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods: GET, OPTIONS`, and `Access-Control-Allow-Headers: Content-Type, Authorization`. Short-circuits `OPTIONS` requests with status 204 No Content.
  - Implement `LoggingMiddleware`: logs incoming request method, path, and remote address without emojis.
- Task 3.3: Implement SPA static asset handler in `internal/adapters/http/spa.go`:
  - Implement `SPAHandler` struct implementing `http.Handler`. Holds an `fs.FS` instance.
  - In `ServeHTTP`:
    - Clean request URL path.
    - If path is root or empty, serve `index.html`.
    - Attempt to open file in `fs.FS`.
    - If file exists and is not a directory, serve via `http.ServeContent` or `http.FileServer`.
    - If file does not exist, check if request path starts with `/api/`. If it is an API route, do not fallback to HTML; return 404 JSON error.
    - For non-API routes, fallback to serving `index.html` with status 200 and `Content-Type: text/html; charset=utf-8`.
- Task 3.4: Implement API handlers in `internal/adapters/http/handlers.go`:
  - `HandleWorkspaces`: scans workspaces directory using `adapters.ScanWorkspaces`, extracts query parameters `window`, `sort`, and `order`, applies domain `FilterLedgersByTimeWindow` and `SortLedgers`, and serializes the resulting ledger slice as JSON.
  - `HandleWorkspaceDetail`: parses `ticketId` from URL path (e.g. `/api/workspaces/{ticketId}`). Scans workspaces, searches for matching ticket ID. If not found, returns 404 `ErrorResponse`. If found, loads artifacts using `adapters.ReadTicketArtifacts`, combines into `WorkspaceDetail`, and returns JSON.
  - `HandleFleet`: scans workspaces, applies optional `window` filter, executes `domain.CalculateFleetSummary`, and serializes `FleetSummary` as JSON.
- Task 3.5: Implement server constructor and routing in `internal/adapters/http/server.go`:
  - Define `Server` struct storing HTTP server instance, router, workspaces directory path, and static filesystem.
  - Implement `NewServer(workspacesDir string, staticFS fs.FS) *Server`:
    - Uses Go 1.22+ `http.NewServeMux` with pattern matching (`GET /api/workspaces`, `GET /api/workspaces/{ticketId}`, `GET /api/fleet`).
    - Wraps API handlers with CORS middleware.
    - Registers SPA static handler on root pattern `/`.
    - Returns configured server instance.

### Phase 4: CLI Presentation Subcommand Wiring (`cmd/agentic-lens`)
- Task 4.1: Update `cmd/agentic-lens/main.go`:
  - Check `os.Args` for `server` subcommand as first positional argument.
  - If `server` is present:
    - Create a dedicated `flag.FlagSet` for server configuration.
    - Define `--port` flag (integer, default 8080).
    - Define `--workspaces` flag (string, default empty, resolved via `adapters.ResolveWorkspacesDir`).
    - Parse remaining arguments.
    - Initialize static filesystem from `web.GetFS()`.
    - Initialize HTTP server via `httpadapter.NewServer(workspacesDir, staticFS)`.
    - Print startup information to stdout (host, port, workspaces directory, zero emojis).
    - Start server listener with support for graceful shutdown on interrupt signal (`SIGINT`, `SIGTERM`).
  - If `server` is not present, preserve existing flag parsing and CLI scan behavior (`--list`, `--ticket-id`, `--append`, `--json-output`).

### Phase 5: Verification, Testing, and Quality Gates
- Task 5.1: Create table-driven tests in `internal/domain/aggregations_test.go`:
  - Test `FilterLedgersByTimeWindow` with boundary dates across 7-day, 30-day, 90-day, and all windows.
  - Test `SortLedgers` across all sort dimensions in both ascending and descending order.
  - Test empty ledger slices and single-item slices.
- Task 5.2: Create HTTP integration tests in `internal/adapters/http/server_test.go`:
  - Setup temporary workspaces fixtures with sample `agentic_metrics.json`, `CONTRACT.md`, and `SUMMARY.md` files.
  - Create test server using `httptest.NewServer`.
  - Test `GET /api/workspaces`: verify status 200, array length, JSON keys, and filtering query parameters.
  - Test `GET /api/workspaces/{ticketId}` with existing ticket: verify status 200, ledger fields, and artifact contents.
  - Test `GET /api/workspaces/{ticketId}` with non-existent ticket: verify status 404 and structured `ErrorResponse`.
  - Test `GET /api/fleet`: verify status 200 and fleet aggregate numbers match domain calculations.
  - Test CORS preflight: send `OPTIONS /api/workspaces` and verify 204 No Content and allowed headers.
  - Test static asset delivery: verify requesting existing file returns file content.
  - Test SPA fallback: verify requesting non-existent UI route returns `index.html` with status 200.
  - Test API 404 guard: verify requesting non-existent `/api/unknown` returns 404 JSON, not `index.html`.
- Task 5.3: Update `cmd/agentic-lens/main_test.go`:
  - Add test verifying `server` subcommand flag parsing and invalid argument handling.
- Task 5.4: Execute verification gates:
  - Run `go vet ./...`.
  - Run `go test -v -race ./...`.
  - Build binary: `go build -o bin/agentic-lens ./cmd/agentic-lens`.
  - Run `bin/agentic-lens --list` to verify regression safety on existing CLI.
  - Verify zero `any` or `interface{}` across all new files.
  - Verify zero emojis across all code, strings, log messages, and error payloads.

## 4. Key Risks and Mitigations
- Risk: SPA fallback routing mistakenly catching `/api/*` requests and returning HTML instead of 404 JSON.
  Mitigation: Explicit path prefix guard in SPA handler ensuring any route beginning with `/api/` immediately delegates to JSON 404 error handler rather than falling back to `index.html`.
- Risk: Missing static frontend files failing Go compilation if `web/dist` is not yet generated by Vite.
  Mitigation: Provide a self-contained baseline `internal/adapters/web/dist/index.html` embedded directly into the binary, decoupled from the external `web/` build process during early iterations.
- Risk: Cross-origin resource sharing errors during Vite frontend development in AL-3.
  Mitigation: Full CORS middleware in `internal/adapters/http` permitting all origins and handling HTTP `OPTIONS` preflight requests cleanly.
- Risk: Concurrent read race conditions when scanning workspaces while HTTP requests arrive.
  Mitigation: Read-only repository scanning and value semantics for all domain entities passed across goroutines. Verified under `go test -race ./...`.
