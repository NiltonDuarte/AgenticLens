# Contract: AL-13 Semantic Versioning and Automated CI for Pull Requests and Main Merges

## 1. Problem Statement

AgenticLens currently operates without compiled-in semantic versioning and lacks an automated continuous integration pipeline on GitHub Actions. This creates two distinct operational vulnerabilities:

First, without embedded build metadata, users, downstream automation pipelines, and monitoring agents cannot inspect the active release version, git commit hash, or build timestamp via CLI commands or HTTP endpoints. Troubleshooting production issues requires external heuristics rather than deterministic runtime introspection.

Second, without automated CI running against pull requests and pushes to the main branch, code regressions can slip into the primary branch unchecked. Race conditions in Go concurrency, TypeScript compilation and lint failures, formatting anomalies, broken embedded static assets, and single-binary packaging defects are caught only during manual local verifications.

Ticket AL-13 establishes software semantic versioning with compile-time linker flag injection across the Go backend, exposes version endpoints and CLI commands, surfaces the active version in the TypeScript web navigation header, and configures an automated GitHub Actions CI workflow enforcing all static checks, unit tests, race detector runs, and binary compilation for pull requests and main branch merges.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All changes across Go backend, TypeScript frontend, build tooling, and workflow configuration must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. CLI Application Layer (backend/cmd/agentic-lens/main.go):
   - Handles command dispatch for the version subcommand and command-line flags (-v, --version, --json, --short).
   - Formats build metadata for terminal output.
   - Contains zero build metadata derivation or mutation logic.
2. Imperative Adapters Layer (backend/internal/adapters/http/):
   - Registers GET /api/version route in the HTTP multiplexer.
   - Serializes domain BuildInfo into BuildInfoDTO JSON responses.
   - Sets appropriate HTTP status codes and headers (application/json; charset=utf-8).
   - Contains zero hardcoded version strings or compilation timestamps.
3. Functional Core Domain Layer (backend/internal/domain/version.go):
   - Houses the pure immutable BuildInfo domain model and compile-time package variables.
   - Implements pure deterministic formatting functions (FormatBuildInfo, FormatShortVersion).
   - Remains 100% pure functional: zero disk I/O, zero network calls, zero external processes, and zero system clocks. Default values serve as fallback when linker flags are omitted.
4. Presentation and Client Layer (web/src/domain/types.ts, web/src/api/client.ts, web/src/components/Navbar.tsx):
   - Strict TypeScript interface BuildInfo mirroring the Go domain contract.
   - Pure UI component rendering in Navbar displaying active version badge and tooltip with git commit and build date.
   - Safe client-side fallback when operating in offline preview mode or when the API endpoint is unreachable.

### Pure Domain Core Isolation
1. BuildInfo is defined as an immutable value object in backend/internal/domain/version.go containing string fields: Version, GitCommit, and BuildDate.
2. Package-level string variables (Version, GitCommit, BuildDate) reside in backend/internal/domain/version.go and receive linker flag assignments during go build.
3. Default variable values are deterministic and non-empty: Version defaults to v0.1.0-dev, GitCommit defaults to unknown, and BuildDate defaults to unknown.
4. GetBuildInfo() BuildInfo is a pure constructor returning an immutable BuildInfo value object constructed from package variables, applying fallback defaults if empty strings are encountered.
5. FormatBuildInfo(info BuildInfo) string and FormatShortVersion(info BuildInfo) string are pure deterministic string formatters with zero side effects.

### Imperative Shell Boundary Rules
1. HTTP REST Endpoint: GET /api/version must return HTTP status 200 with JSON payload containing version, commit, and build_date fields.
2. Typed Response Constraints: BuildInfoDTO in backend/internal/adapters/http/dto.go must be added to the generic ResponsePayload interface type union.
3. Single Binary Packaging: Versioning must not alter the embedded asset pipeline (embed.FS in backend/internal/adapters/web/embed.go).
4. Zero Any Mandate: Zero any or empty interface{} permitted across Go backend packages and TypeScript frontend code.
5. Zero Emoji Mandate: Zero emojis across all source code, comments, log lines, CLI outputs, UI badges, tooltips, JSON payloads, workflow files, and markdown artifacts.

---

## 3. Scope Boundaries: Go Backend, TypeScript Frontend, Build System, and CI

### In Scope
1. Domain Core Models and Functions (backend/internal/domain/):
   - Create backend/internal/domain/version.go defining BuildInfo struct, package variables, GetBuildInfo constructor, and pure string formatting functions.
   - Create backend/internal/domain/version_test.go with table-driven tests asserting pure formatting behavior, default values, and edge cases.
2. HTTP Boundary Adapters (backend/internal/adapters/http/):
   - Define BuildInfoDTO in backend/internal/adapters/http/dto.go and add it to the ResponsePayload type union.
   - Implement VersionHandler in backend/internal/adapters/http/handlers.go.
   - Register GET /api/version route in backend/internal/adapters/http/server.go.
   - Add integration tests in backend/internal/adapters/http/server_test.go asserting status 200, Content-Type, and JSON response schema.
3. CLI Application Integration (backend/cmd/agentic-lens/):
   - Add version subcommand in backend/cmd/agentic-lens/main.go.
   - Add support for -v and --version flags at root CLI invocation.
   - Add flag support for --json and --short within the version command.
   - Add CLI integration tests in backend/cmd/agentic-lens/main_test.go.
4. Build Tooling and Linker Flag Injection (Makefile):
   - Define VERSION, GIT_COMMIT, BUILD_DATE, and LDFLAGS variables in Makefile using git tags, commit hashes, and UTC timestamps with fallback defaults.
   - Update build target in Makefile to inject LDFLAGS into go build targeting package agenticlens/internal/domain.
   - Verify make check, make test, make test-backend, make test-frontend, and make build targets execute cleanly.
5. TypeScript Frontend Integration (web/src/):
   - Add BuildInfo interface in web/src/domain/types.ts.
   - Add mockBuildInfo fixture in web/src/api/mockData.ts.
   - Implement fetchVersion in web/src/api/client.ts with fallback to mock data on network error.
   - Update web/src/components/Navbar.tsx to accept optional version prop or fetch active build info, rendering a styled version badge.
   - Add CSS styling for .version-badge in web/src/index.css adhering to dark theme aesthetic.
   - Update web/src/components/Navbar.test.tsx with Vitest tests asserting badge display, tooltips, fallback states, and zero emojis.
6. Continuous Integration Workflow (.github/workflows/):
   - Create .github/workflows/ci.yml configured to trigger on pull_request targeting main and push to main.
   - Configure environment matrix/setup for Go (reading version from backend/go.mod) and Node.js 20.
   - Run make check (vet, lint, git diff check), make test (Go race-detected test suite and Vitest suite), and make build (single binary build with embedded UI and ldflags).
   - Execute binary verification step invoking ./bin/agentic-lens version and ./bin/agentic-lens version --json.

### Out of Scope
1. Automated release tagging, GitHub Releases creation, changelog extraction, and multi-architecture binary asset packaging, which belong to ticket AL-14.
2. Telemetry schema migrations or ledger transformations, which belong to ticket AL-15.
3. OpenCode database live hooks and session tree inference, which belong to ticket AL-17.

---

## 4. Architectural Specifications

### Linker Flag Injection Format
Build metadata is injected at compile time via the Go linker using -ldflags with -X options targeting variables in the agenticlens/internal/domain package:
- agenticlens/internal/domain.Version: Derived from git describe --tags --always --dirty, defaulting to v0.1.0-dev.
- agenticlens/internal/domain.GitCommit: Derived from git rev-parse --short HEAD, defaulting to unknown.
- agenticlens/internal/domain.BuildDate: Derived from date -u +"%Y-%m-%dT%H:%M:%SZ", defaulting to unknown.

### HTTP Version Endpoint Contract
- Method: GET
- Path: /api/version
- Status: 200 OK
- Content-Type: application/json; charset=utf-8
- Response JSON Schema:
  - version: string (semantic release tag or dev identifier)
  - commit: string (short git commit SHA or unknown)
  - build_date: string (ISO 8601 UTC timestamp or unknown)

### CLI Version Commands and Flags
1. agentic-lens version: Outputs standard formatted string: agentic-lens version <version> (commit: <commit>, built: <build_date>).
2. agentic-lens -v or agentic-lens --version: Outputs standard formatted version string identical to the version subcommand.
3. agentic-lens version --short: Outputs short version string: <version>.
4. agentic-lens version --json: Outputs JSON payload containing version, commit, and build_date keys.

### GitHub Actions CI Workflow Specification
- Workflow File: .github/workflows/ci.yml
- Name: CI
- Triggers:
  - pull_request: branches [ main ]
  - push: branches [ main ]
- Job Structure:
  - Name: build-and-test
  - Runs On: ubuntu-latest
  - Steps:
    1. Checkout repository with actions/checkout@v4 using fetch-depth: 0 to enable git describe and commit resolution.
    2. Setup Go environment with actions/setup-go@v5 using go-version-file: backend/go.mod and caching enabled.
    3. Setup Node.js environment with actions/setup-node@v4 using node-version: 20 and caching npm dependencies for web/package-lock.json.
    4. Install web dependencies using npm ci inside web/.
    5. Run quality checks using make check.
    6. Run backend and frontend test suites using make test.
    7. Build standalone binary using make build.
    8. Verify binary execution by running ./bin/agentic-lens version and checking zero exit status.

---

## 5. Acceptance Criteria

1. Running agentic-lens version, agentic-lens -v, or agentic-lens --version prints the semantic version, git commit SHA, and build timestamp to standard output with exit code 0.
2. Running agentic-lens version --json prints valid JSON matching the BuildInfo schema.
3. Querying GET /api/version on the HTTP server returns status 200 with Content-Type application/json; charset=utf-8 and body containing version, commit, and build_date fields.
4. The TypeScript web dashboard displays the active semantic version in the navigation header, styled cleanly as a version badge with commit and build date in the title tooltip.
5. When running in offline or mock mode, the web dashboard displays a safe fallback version indicator without layout distortion or uncaught errors.
6. Makefile includes VERSION, GIT_COMMIT, BUILD_DATE, and LDFLAGS variables, and passes them to go build.
7. make check, make test (including go test -race ./... and Vitest), and make build pass cleanly with zero errors.
8. .github/workflows/ci.yml triggers on pull requests targeting main and pushes to main, executing checkout, dependency setup, make check, make test, make build, and binary verification.
9. All Go and TypeScript code strictly enforces zero any, zero unhandled errors, and zero emojis.
