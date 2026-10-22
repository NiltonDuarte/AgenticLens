# Contract: AL-2 Go HTTP Server, Embedded Static FS, and Telemetry REST API

## 1. Problem Statement
AgenticLens currently operates exclusively as a command line scanner and table renderer. To power the planned TypeScript web dashboard without requiring a Node.js runtime in production or separate backend infrastructure, AgenticLens requires a built-in Go HTTP server. This server must expose read-only REST API endpoints for scanned telemetry data, serve embedded static frontend assets, support client-side single page application routing fallbacks, and enable cross-origin resource sharing for local Vite development.

The goal of AL-2 is to implement the Go HTTP server layer, embed a baseline static filesystem, define pure domain filtering and sorting utilities, and wire the `server` subcommand into the `agentic-lens` CLI executable.

## 2. Functional Core, Imperative Shell (FCIS) Invariants
All implementation must strictly comply with TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward: CLI entrypoint in `cmd/agentic-lens` depends on adapters and domain; HTTP adapter in `internal/adapters/http` and web asset container in `internal/adapters/web` depend on domain and standard library; domain in `internal/domain` has zero external dependencies and imports no outer packages or I/O primitives.

### Functional Core Isolation
Package `internal/domain` must contain pure filtering predicates and sorting functions operating on immutable slices of domain models.
The domain layer must not import disk, network, or HTTP packages such as `net/http` or `os`. It must not invoke system clocks such as `time.Now` directly within rollup calculations; reference timestamps must be passed as arguments.
All sorting functions must be deterministic and pure, accepting a slice and returning a new sorted slice without mutating inputs in-place.

### Imperative Boundary Shell
Package `internal/adapters/http` must isolate all HTTP routing, header management, CORS negotiation, parameter decoding, JSON response encoding, status code mapping, and static filesystem delivery.
External parameters from HTTP query strings or URL paths must be parsed into typed domain representations or rejected with structured error responses before invoking domain logic.
Filesystem access for static assets must operate through the `io/fs` abstract filesystem interface (`fs.FS`), allowing `internal/adapters/web` to supply embedded production assets and tests to inject in-memory fixtures (`fstest.MapFS`).
Path traversal attacks through static file requests or ticket artifact endpoints must be defensively rejected.

### Presentation Shell
Package `cmd/agentic-lens` must extend standard library flag parsing to support a `server` subcommand with flags for port and workspace path configuration.
Server lifecycle logging and startup diagnostics must have zero emojis.

## 3. Scope Boundaries: Go Backend and TypeScript Frontend
AL-2 delivers the Go HTTP backend and embedded serving infrastructure. It establishes the backend foundation for subsequent frontend development.

### In Scope for AL-2
- Pure domain models for workspace details, time window enums, and ledger sorting fields in `internal/domain/models.go` and `internal/domain/enums.go`.
- Pure domain filtering and sorting functions in `internal/domain/aggregations.go`.
- Embedded static filesystem container in `internal/adapters/web` containing a minimal static distribution with `index.html`.
- HTTP server configuration, router, and middleware in `internal/adapters/http`.
- REST API handlers for `GET /api/workspaces`, `GET /api/workspaces/{ticketId}`, and `GET /api/fleet`.
- CORS middleware supporting local Vite development with cross-origin requests.
- Single page application file server handler serving static files with fallback to `index.html` for non-API routes.
- CLI subcommand `server` in `cmd/agentic-lens/main.go` accepting `--port` and `--workspaces` flags.
- Comprehensive table-driven unit tests for domain filters and HTTP integration tests using `net/http/httptest`.

### Out of Scope for AL-2
- TypeScript web application scaffolding, React components, and Vite configuration (deferred to AL-3).
- TypeScript data models and API client implementations (deferred to AL-3).
- Fleet overview dashboard and ticket deep dive user interfaces (deferred to AL-3 and AL-4).
- Production build embedding of compiled TypeScript bundles (deferred to AL-5).
- Telemetry extraction Python scripts deprecation (deferred to AL-5).

## 4. Architectural Contracts and Data Structures

### Domain Core Specifications (`internal/domain`)
- `TimeWindow`: Typed string enum representing temporal query filters: `TimeWindowAll`, `TimeWindow7D`, `TimeWindow30D`, and `TimeWindow90D`.
- `LedgerSortField`: Typed string enum representing sorting dimensions: `SortFieldUpdatedAt`, `SortFieldCost`, `SortFieldDuration`, and `SortFieldInvocations`.
- `WorkspaceDetail`: Pure domain struct combining an `AgenticMetricsLedger` with its associated `TicketArtifacts`.
- `FilterLedgersByTimeWindow`: Pure function receiving a slice of `AgenticMetricsLedger`, a `TimeWindow`, and a reference `time.Time`, returning filtered ledgers based on ticket completion, start, or update timestamps.
- `SortLedgers`: Pure function receiving a slice of `AgenticMetricsLedger`, a `LedgerSortField`, and a boolean ascending flag, returning sorted ledgers.

### Boundary Adapters Specifications (`internal/adapters/http` and `internal/adapters/web`)
- `ErrorResponse`: Structured error payload containing `error` (short machine code), `message` (human-readable description), and `status_code` (integer HTTP code).
- `WorkspaceDetailResponse`: Concrete response model matching `WorkspaceDetail` serialization.
- `Server`: HTTP server wrapper holding router, workspace directory path, and static assets filesystem (`fs.FS`).
- `NewServer`: Factory constructor accepting workspace path, static filesystem, and returning configured `Server`.
- `Routes`:
  - `GET /api/workspaces`: Reads scanned ledgers, applies optional `window`, `sort`, and `order` query parameters, and serializes JSON array.
  - `GET /api/workspaces/{ticketId}`: Extracts `ticketId` from URL path, locates matching ledger from scan report, loads artifacts via `ReadTicketArtifacts`, and returns `WorkspaceDetailResponse`. Returns HTTP 404 with `ErrorResponse` if ticket ID does not match.
  - `GET /api/fleet`: Computes `FleetSummary` across scanned ledgers (with optional `window` filter) and serializes JSON response.
  - `GET /*`: Serves static assets from embedded `fs.FS`. If requested file does not exist and path does not start with `/api/`, serves `index.html` with status 200.
- `CORSMiddleware`: Sets `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, and `Access-Control-Allow-Headers`. Responds with HTTP 204 to `OPTIONS` preflight requests.
- `internal/adapters/web`: Declares embedded filesystem using `//go:embed` on package directory `dist/` containing minimal placeholder `index.html`.

### Presentation Shell (`cmd/agentic-lens`)
- Subcommand `server` parsed via `flag.NewFlagSet` or initial argument inspection:
  - `--port`: Integer port number (default: 8080).
  - `--workspaces`: String path to workspaces directory (default: resolved via `adapters.ResolveWorkspacesDir`).
- Clean startup message emitted to standard output indicating host, port, and resolved workspaces directory path.
- Graceful shutdown handling listening for OS interrupt signals (`SIGINT`, `SIGTERM`).

## 5. Quality Standards and Verification Gates
- Zero `any` and zero `interface{}` across all handlers, middleware, models, and test assertions. Concrete structs only.
- Strict error wrapping with `%w` verb and contextual metadata. Zero discarded errors.
- HTTP tests implemented using standard library `httptest.NewServer` and `httptest.ResponseRecorder` without third-party frameworks.
- Verification gates passing: `go vet ./...` and `go test -race ./...`.
- Zero emojis across all code, log messages, error responses, and workspace artifacts.
