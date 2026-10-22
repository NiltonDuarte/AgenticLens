# Implementation Plan: AL-13 Semantic Versioning and Automated CI for Pull Requests and Main Merges

## 1. Goal and Iteration Scope

Deliver compiled-in semantic versioning with build-time linker flag injection, CLI version commands, HTTP version endpoints, and web UI version indicators, backed by an automated GitHub Actions CI workflow that enforces code checks, tests, and binary build verification on pull requests and pushes to main. This iteration covers the complete scope of ticket AL-13.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following existing modules and newly established files:

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/version.go: Define BuildInfo struct, package-level variables (Version, GitCommit, BuildDate), GetBuildInfo constructor, FormatBuildInfo, and FormatShortVersion formatters.
- backend/internal/domain/version_test.go: Table-driven unit tests verifying build info retrieval, custom values, fallback defaults, and formatting functions. Zero mocks.

### Boundary Adapters Layer (backend/internal/adapters/http/)
- backend/internal/adapters/http/dto.go: Define BuildInfoDTO and add it to the ResponsePayload interface union constraint.
- backend/internal/adapters/http/handlers.go: Implement VersionHandler returning JSON BuildInfoDTO.
- backend/internal/adapters/http/server.go: Register GET /api/version route in NewServer multiplexer.
- backend/internal/adapters/http/server_test.go: Integration tests querying GET /api/version on test server verifying status 200, Content-Type, and JSON payload.

### CLI Application Layer (backend/cmd/agentic-lens/)
- backend/cmd/agentic-lens/main.go: Add version subcommand and handle -v and --version flags at root CLI entrypoint. Add --json and --short flag support.
- backend/cmd/agentic-lens/main_test.go: Integration tests verifying version command outputs, JSON output parsing, and flag variations.

### Build Automation Layer (Makefile)
- Makefile: Define VERSION, GIT_COMMIT, BUILD_DATE, and LDFLAGS variables. Update the build target to compile with -ldflags "$(LDFLAGS)".

### TypeScript Presentation Layer (web/src/)
- web/src/domain/types.ts: Define BuildInfo interface.
- web/src/api/mockData.ts: Add mockBuildInfo fixture.
- web/src/api/client.ts: Implement fetchVersion returning ApiResponse<BuildInfo> with fallback on network failure.
- web/src/components/Navbar.tsx: Render version badge indicator in navigation header with title tooltip for commit and build date. Support optional version prop and fallback state.
- web/src/index.css: Add dark theme styling for .version-badge.
- web/src/components/Navbar.test.tsx: Vitest unit tests verifying badge rendering, tooltips, mock fallbacks, and zero emojis.

### GitHub Actions Workflow (.github/workflows/)
- .github/workflows/ci.yml: Multi-step CI workflow running on pull_request targeting main and push to main, running make check, make test, make build, and binary verification.

---

## 3. Phased Implementation Tasks

### Phase 1: Pure Domain Core Version Models and Formatters (backend/internal/domain/)
- Task 1.1: Create backend/internal/domain/version.go:
  - Declare package domain.
  - Define package variables Version initialized to "v0.1.0-dev", GitCommit initialized to "unknown", and BuildDate initialized to "unknown".
  - Define BuildInfo struct with fields Version of type string (json tag version), GitCommit of type string (json tag commit), and BuildDate of type string (json tag build_date).
  - Implement GetBuildInfo() BuildInfo returning an immutable BuildInfo struct, falling back to defaults if any package variable is set to empty string.
  - Implement FormatBuildInfo(info BuildInfo) string returning formatted string: agentic-lens version <version> (commit: <commit>, built: <build_date>).
  - Implement FormatShortVersion(info BuildInfo) string returning info.Version.
- Task 1.2: Create backend/internal/domain/version_test.go:
  - Test GetBuildInfo default values.
  - Table-driven tests for FormatBuildInfo verifying standard formatting, dirty tags, and empty fallbacks.
  - Table-driven tests for FormatShortVersion.
  - Verify zero external dependencies, zero clocks, and zero I/O.

### Phase 2: Boundary Adapters and HTTP Version Route (backend/internal/adapters/http/)
- Task 2.1: Update backend/internal/adapters/http/dto.go:
  - Define BuildInfoDTO struct with fields Version of type string (json tag version), Commit of type string (json tag commit), and BuildDate of type string (json tag build_date).
  - Update ResponsePayload interface constraint to include BuildInfoDTO in the union type.
  - Implement WriteVersionResponse(w http.ResponseWriter, info domain.BuildInfo) helper mapping domain.BuildInfo to BuildInfoDTO and invoking WriteJSON with HTTP 200.
- Task 2.2: Update backend/internal/adapters/http/handlers.go:
  - Implement VersionHandler() http.HandlerFunc: invokes domain.GetBuildInfo() and calls WriteVersionResponse.
- Task 2.3: Update backend/internal/adapters/http/server.go:
  - Register route mux.HandleFunc("GET /api/version", VersionHandler()) inside NewServer before SPA static file handler.
- Task 2.4: Update backend/internal/adapters/http/server_test.go:
  - Add test querying GET /api/version using httptest.NewRecorder.
  - Assert response status is 200 OK.
  - Assert Content-Type header is application/json; charset=utf-8.
  - Parse JSON response into BuildInfoDTO and assert non-empty version, commit, and build_date fields.

### Phase 3: CLI Subcommand and Flags (backend/cmd/agentic-lens/)
- Task 3.1: Update backend/cmd/agentic-lens/main.go:
  - Implement runVersion(args []string):
    - Parse version-specific flags: --json (bool) and --short (bool).
    - If --json is set: marshal domain.GetBuildInfo() to formatted JSON and print to stdout.
    - If --short is set: print domain.FormatShortVersion(domain.GetBuildInfo()) to stdout.
    - Default: print domain.FormatBuildInfo(domain.GetBuildInfo()) to stdout.
  - In main():
    - Check if the first argument is "version": invoke runVersion(os.Args[2:]) and return.
    - Check if the first argument is "-v" or "--version": invoke runVersion(nil) and return.
- Task 3.2: Update backend/cmd/agentic-lens/main_test.go:
  - Add integration tests executing CLI version subcommand, -v flag, --version flag, --short flag, and --json flag.
  - Assert stdout matches expected format and exit code is 0.

### Phase 4: Build System and Linker Flag Injection (Makefile)
- Task 4.1: Update Makefile:
  - Add VERSION variable using git describe --tags --always --dirty with fallback to v0.1.0-dev.
  - Add GIT_COMMIT variable using git rev-parse --short HEAD with fallback to unknown.
  - Add BUILD_DATE variable using date -u +"%Y-%m-%dT%H:%M:%SZ" with fallback to unknown.
  - Add LDFLAGS variable containing -X assignments for agenticlens/internal/domain.Version, agenticlens/internal/domain.GitCommit, and agenticlens/internal/domain.BuildDate.
  - Update build target to compile with -ldflags "$(LDFLAGS)".
- Task 4.2: Verify local Makefile targets:
  - Run make vet, make lint, make check, make test-backend, and make build.
  - Verify generated binary ./bin/agentic-lens version reflects git commit and current timestamp.

### Phase 5: Frontend Domain Types, API Client, and Navbar Version Indicator (web/src/)
- Task 5.1: Update web/src/domain/types.ts:
  - Define interface BuildInfo with readonly fields: version of type string, commit of type string, and build_date of type string.
- Task 5.2: Update web/src/api/mockData.ts:
  - Export mockBuildInfo of type BuildInfo with sample values.
- Task 5.3: Update web/src/api/client.ts:
  - Implement fetchVersion(): Promise<ApiResponse<BuildInfo>>:
    - Attempt fetch('/api/version').
    - If successful, parse JSON into BuildInfo and return with isMock: false.
    - On HTTP error or exception, catch safely and return mockBuildInfo with isMock: true.
- Task 5.4: Update web/src/components/Navbar.tsx:
  - Extend NavbarProps to accept optional versionInfo of type BuildInfo.
  - If versionInfo prop is not provided, fetch version info asynchronously on component mount using fetchVersion, storing in local state.
  - Render version badge span element with class version-badge adjacent to the brand title or status pill.
  - Populate title attribute on the badge with Commit: <commit> | Built: <build_date>.
  - Render fallback text v0.1.0-dev if version is not yet resolved.
- Task 5.5: Update web/src/index.css:
  - Add styles for .version-badge: small font size, subtle border, rounded corners, muted text color, monospaced font family, matching existing dark theme navigation elements.
- Task 5.6: Update web/src/components/Navbar.test.tsx:
  - Test Navbar renders version badge when versionInfo prop is provided.
  - Test Navbar renders fallback version badge when versionInfo is omitted.
  - Test tooltip title attribute contains commit and build date.
  - Test zero emojis exist across all rendered elements.

### Phase 6: Automated Continuous Integration Workflow (.github/workflows/)
- Task 6.1: Create directory .github/workflows if not already present.
- Task 6.2: Create .github/workflows/ci.yml:
  - Define workflow name: CI.
  - Configure on triggers: pull_request on branch main, push on branch main.
  - Define job build-and-test running on ubuntu-latest.
  - Configure step: Checkout code using actions/checkout@v4 with fetch-depth: 0.
  - Configure step: Set up Go using actions/setup-go@v5 with go-version-file: backend/go.mod and cache: true (cache-dependency-path: backend/go.sum).
  - Configure step: Set up Node.js using actions/setup-node@v4 with node-version: 20 and cache: npm (cache-dependency-path: web/package-lock.json).
  - Configure step: Install frontend dependencies using npm ci in web/.
  - Configure step: Run static quality checks using make check.
  - Configure step: Run automated test suites using make test.
  - Configure step: Build standalone executable using make build.
  - Configure step: Verify compiled binary version output using ./bin/agentic-lens version and ./bin/agentic-lens version --json.

### Phase 7: Verification, Adversarial Testing, and Compliance Verification
- Task 7.1: Go Unit and Integration Test Verification:
  - Execute cd backend && go test -v -race ./... verifying all domain, adapter, and CLI tests pass without race conditions.
  - Assert zero unhandled errors.
- Task 7.2: Frontend Test and Typecheck Verification:
  - Execute cd web && npm run typecheck asserting zero TypeScript compilation errors.
  - Execute cd web && npm run lint asserting zero lint warnings or errors.
  - Execute cd web && npm test asserting all Vitest suites pass including Navbar tests.
- Task 7.3: Build Pipeline and CLI Output Verification:
  - Execute make clean && make build from repository root.
  - Verify ./bin/agentic-lens is generated.
  - Verify ./bin/agentic-lens version output matches injected git commit SHA and UTC build date.
  - Verify ./bin/agentic-lens version --json outputs parseable JSON.
- Task 7.4: Static Compliance Audits:
  - Verify zero any across all Go files (internal/domain, internal/adapters, cmd).
  - Verify zero any across all TypeScript files (web/src).
  - Verify zero emojis across all code files, comments, docstrings, UI templates, test files, and workflow files.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Domain Core Version Models and Formatters) -> unblocks Phase 2 (HTTP Adapter) and Phase 3 (CLI Entrypoint).
2. Phase 2 (HTTP Version Route) -> unblocks Phase 5 (Frontend API Client and Navbar Badge).
3. Phase 3 (CLI Version Subcommand and Flags) -> unblocks CLI verification.
4. Phase 4 (Makefile Linker Flag Injection) -> unblocks local binary verification and Phase 6 (CI Workflow).
5. Phase 5 (Frontend Domain Types, API Client, Navbar Component, and Tests) -> unblocks full frontend build.
6. Phase 6 (GitHub Actions CI Configuration) -> unblocks automated verification for PRs and main merges.
7. Phase 7 (Full Verification and Compliance Audit) -> validates acceptance criteria and prepares delivery artifacts.

---

## 5. Verification Plan

### Test Suites and Quality Commands
- Go Domain Tests: cd backend && go test -v -race ./internal/domain/...
- Go HTTP Adapter Tests: cd backend && go test -v -race ./internal/adapters/http/...
- Go CLI Integration Tests: cd backend && go test -v -race ./cmd/agentic-lens/...
- Go Static Analysis: cd backend && go vet ./...
- TypeScript Typecheck: cd web && npm run typecheck
- TypeScript Linter: cd web && npm run lint
- TypeScript Component Tests: cd web && npm test
- Complete Pipeline: make check && make test && make build
- Binary CLI Inspection:
  - ./bin/agentic-lens version
  - ./bin/agentic-lens -v
  - ./bin/agentic-lens --version
  - ./bin/agentic-lens version --short
  - ./bin/agentic-lens version --json
- HTTP Endpoint Inspection:
  - Start server: ./bin/agentic-lens server --port 8085 &
  - Query endpoint: curl -s http://localhost:8085/api/version
  - Terminate server process.
- Workflow Syntax Validation:
  - Validate .github/workflows/ci.yml structure and YAML indentation.

### Compliance Checks
- Zero Any Check: Scan with ripgrep for any or interface{} in Go backend and TypeScript frontend.
- Zero Emoji Check: Scan with regex for unicode emoji ranges across all repository files.
- Single Binary Check: Verify the compiled Go executable runs self-contained without Node.js or external static file dependencies.
