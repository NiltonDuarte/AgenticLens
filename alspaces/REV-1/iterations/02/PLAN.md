# Implementation Plan: REV-1 Iteration 02
## Strict Fallback Elimination and Explicit Error Handling for Workspaces Directory Resolution

### 1. Goal and Problem Statement
In response to steering feedback on ticket REV-1, this iteration eliminates runtime panics and removes all heuristic fallback cascades from workspaces directory resolution in the Go boundary adapters. 

Previously, `ResolveWorkspacesDir` attempted heuristic cascades, including container volume inspection at `/app/workspaces`, directory probing for `./workspaces` in the current working directory, and parent directory traversal before defaulting to the string literal `workspaces`. A recent modification eliminated some heuristics but introduced an unhandled `panic`, causing test failures when configuration was absent.

Under this iteration, `ResolveWorkspacesDir` becomes a deterministic boundary operation governed strictly by two inputs:
1. An explicit command-line override or argument.
2. A valid project configuration file containing `workspaces.dir`.

If neither is supplied or resolvable, `ResolveWorkspacesDir` must fail fast by returning a typed sentinel error (`ErrWorkspacesDirNotFound`). It must never panic and must never fall back to heuristic directory probing. All callers across CLI commands, filesystem scanners, ticket loaders, and batch migration engines must be updated to propagate or handle this error.

---

### 2. Architectural Delta Specification

#### 1. Sentinel Error Definition
- Export `ErrWorkspacesDirNotFound = errors.New("workspaces directory not found")` in package `adapters` (`backend/internal/adapters/repository.go`).

#### 2. Signature Refactoring
- Refactor `ResolveWorkspacesDir(overridePath string) string` to `ResolveWorkspacesDir(overridePath string) (string, error)`.

#### 3. Resolution Logic
- Evaluate `overridePath`: if non-empty, return `overridePath` and a nil error immediately.
- Evaluate project configuration: call `config.Load("")`. If `config.Load` succeeds and `cfg.Workspaces.Dir` is non-empty, return `cfg.Workspaces.Dir` and a nil error.
- Remove all fallback mechanisms:
  - Remove container volume probing at `/app/workspaces`.
  - Remove working directory inspection for `./workspaces`.
  - Remove parent directory climbing.
  - Remove default fallback string literal `workspaces`.
  - Remove `panic("could not find workspaces.dir")`.
- If neither override nor valid configuration directory is present, return an empty string and `ErrWorkspacesDirNotFound`.

#### 4. Call-Site Propagation and Containment
- In `backend/internal/adapters/repository.go`:
  - `ScanWorkspaces(basePath string) domain.ScanReport`: check error from `ResolveWorkspacesDir(basePath)`. If non-nil, return an empty `ScanReport` containing an empty ledger slice and record the error message in the `Errors` slice.
  - `ReadTicketArtifacts(ticketID string, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) domain.TicketArtifacts`: check error from `ResolveWorkspacesDir(basePath)`. If non-nil, return an empty `domain.TicketArtifacts` struct.
  - `LoadTicketWorkspace(ticketID, basePath string) (*domain.AgenticMetricsLedger, error)`: check error from `ResolveWorkspacesDir(basePath)`. If non-nil, return nil and wrap the error with context.
- In `backend/cmd/agentic-lens/main.go`:
  - In `runGenerate`: check error from `ResolveWorkspacesDir(*workspacesFlag)`. If non-nil, exit with error message.
  - In `runServer`: capture error from `ResolveWorkspacesDir(*workspacesFlag)`. If non-nil, print diagnostic to standard error and exit with status code 1.
  - In `runFleet`: capture error from `ResolveWorkspacesDir(opts.WorkspacesDir)` and exit with status code 1 on failure.
  - In `runBackfill`: capture error from `ResolveWorkspacesDir(opts.WorkspacesDir)` and exit with status code 1 on failure.
  - In `runMigrate`: capture error from `ResolveWorkspacesDir(opts.WorkspacesDir)` and exit with status code 1 on failure.
  - In `runRoot`: capture error from `ResolveWorkspacesDir(*workspacesFlag)` and exit with status code 1 on failure.
- In `backend/internal/adapters/backfill/engine.go`:
  - In `Execute(opts BackfillOptions)`: capture error from `ResolveWorkspacesDir(opts.WorkspacesDir)` and return wrapped error when resolution fails.
- In `backend/internal/adapters/backfill/migration_runner.go`:
  - In `Run(opts MigrationOptions)`: capture error from `ResolveWorkspacesDir(opts.WorkspacesDir)` and return wrapped error when resolution fails.

---

### 3. Grounded Reference Architecture & File Anchors

#### Go Boundary Adapters (`backend/internal/adapters/`)
- `backend/internal/adapters/repository.go`:
  - Sentinel error `ErrWorkspacesDirNotFound`
  - Function `ResolveWorkspacesDir(overridePath string) (string, error)`
  - Function `ScanWorkspaces(basePath string) domain.ScanReport`
  - Function `ReadTicketArtifacts(ticketID string, basePath string, configuredArtifacts ...[]domain.WorkspaceArtifactConfig) domain.TicketArtifacts`
  - Function `LoadTicketWorkspace(ticketID, basePath string) (*domain.AgenticMetricsLedger, error)`
- `backend/internal/adapters/repository_test.go`:
  - Test function `TestResolveWorkspacesDir`
  - Test function `TestScanWorkspaces_RoleMappingFactsPreservation`
- `backend/internal/adapters/backfill/engine.go`:
  - Method `(e *BackfillEngine) Execute(opts BackfillOptions) (*BackfillReport, error)`
- `backend/internal/adapters/backfill/engine_test.go`:
  - Engine execution failure tests
- `backend/internal/adapters/backfill/migration_runner.go`:
  - Method `(r *MigrationRunner) Run(opts MigrationOptions) (*SchemaMigrationReport, error)`
- `backend/internal/adapters/backfill/migration_runner_test.go`:
  - Runner execution failure tests

#### Go CLI Entrypoint (`backend/cmd/agentic-lens/`)
- `backend/cmd/agentic-lens/main.go`:
  - Function `runGenerate(args []string)`
  - Function `runServer(args []string)`
  - Function `runFleet(args []string)`
  - Function `runBackfill(args []string)`
  - Function `runMigrate(args []string)`
  - Function `runRoot(args []string)`
- `backend/cmd/agentic-lens/main_test.go`:
  - Integration execution tests without explicit flags

---

### 4. Phased Implementation Tasks for Coder

#### Phase 1: Sentinel Error Export and Function Refactor (`backend/internal/adapters/repository.go`)
- Define and export `ErrWorkspacesDirNotFound = errors.New("workspaces directory not found")`.
- Update signature of `ResolveWorkspacesDir` to accept `overridePath string` and return `(string, error)`.
- If `overridePath != ""`, immediately return `overridePath, nil`.
- Call `config.Load("")`. If `err == nil` and `cfg.Workspaces.Dir != ""`, return `cfg.Workspaces.Dir, nil`.
- If neither condition matches, return empty string and `ErrWorkspacesDirNotFound`.
- Ensure zero calls to `panic`, zero checks against `/app/workspaces`, zero checks for `./workspaces`, and zero parent directory crawling.

#### Phase 2: Internal Adapter Call-Site Updates (`backend/internal/adapters/repository.go`)
- Update `ScanWorkspaces`: call `ResolveWorkspacesDir(basePath)`. If error is returned, construct and return a `domain.ScanReport` with empty ledgers and append formatted resolution error string into `Errors`.
- Update `ReadTicketArtifacts`: call `ResolveWorkspacesDir(basePath)`. If error is returned, return an empty `domain.TicketArtifacts` struct containing `TicketID` and empty items slice.
- Update `LoadTicketWorkspace`: call `ResolveWorkspacesDir(basePath)`. If error is returned, return nil and wrap error using `fmt.Errorf("resolving workspaces directory: %w", err)`.

#### Phase 3: Backfill Adapter Call-Site Updates (`backend/internal/adapters/backfill/`)
- In `backend/internal/adapters/backfill/engine.go` inside `Execute`: call `adapters.ResolveWorkspacesDir(opts.WorkspacesDir)`. On error, return nil and wrap error with context.
- In `backend/internal/adapters/backfill/migration_runner.go` inside `Run`: call `adapters.ResolveWorkspacesDir(opts.WorkspacesDir)`. On error, return nil and wrap error with context.

#### Phase 4: CLI Entrypoint Call-Site Updates (`backend/cmd/agentic-lens/main.go`)
- In `runGenerate`: invoke `adapters.ResolveWorkspacesDir(*workspacesFlag)`. If resolution fails with error, print message to standard error and exit with status code 1.
- In `runServer`: invoke `adapters.ResolveWorkspacesDir(*workspacesFlag)`. If resolution fails with error, print message to standard error and exit with status code 1.
- In `runFleet`: invoke `adapters.ResolveWorkspacesDir(opts.WorkspacesDir)`. If resolution fails with error, print message to standard error and exit with status code 1.
- In `runBackfill`: validate `opts.WorkspacesDir` via `adapters.ResolveWorkspacesDir`. If resolution fails, print error to standard error and exit with status code 1.
- In `runMigrate`: validate `opts.WorkspacesDir` via `adapters.ResolveWorkspacesDir`. If resolution fails, print error to standard error and exit with status code 1.
- In `runRoot`: invoke `adapters.ResolveWorkspacesDir(*workspacesFlag)`. If resolution fails, print error to standard error and exit with status code 1.

#### Phase 5: Test Suite Overhaul (`backend/internal/adapters/repository_test.go`)
- Update `TestResolveWorkspacesDir` subtests to consume the two-value return signature `(string, error)`:
  - Subtest `explicit override`: assert `err == nil` and returned path equals override string.
  - Subtest `config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `yaml config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `root yaml config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `root yml config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `root json config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `nested execution ascends to root config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `nested execution ascends to legacy config file resolution`: assert `err == nil` and returned path matches configured directory.
  - Subtest `root config workspaces dir overrides legacy config workspaces dir`: assert `err == nil` and returned path matches root configured directory.
  - Subtest `precedence between override and config`: assert override takes precedence over config, and both succeed with nil error.
  - Subtest `fails fast when no command flag and no config exists`: test strict failure. When no override is passed and no config file exists in the directory hierarchy, assert `ResolveWorkspacesDir("")` returns empty string and an error satisfying `errors.Is(err, adapters.ErrWorkspacesDirNotFound)`. Assert that even if a `./workspaces` directory is present on disk, it is ignored and fails fast.
  - Add subtest verifying container path `/app/workspaces` is never selected when unconfigured.
- Update line 1948 in `TestScanWorkspaces_RoleMappingFactsPreservation` to handle error return from `ResolveWorkspacesDir("")` and fail test immediately if error is non-nil.

#### Phase 6: Backfill and Migration Runner Test Updates
- In `backend/internal/adapters/backfill/engine_test.go`: add test case verifying `Execute` fails with `adapters.ErrWorkspacesDirNotFound` when `WorkspacesDir` is empty and no configuration file is present.
- In `backend/internal/adapters/backfill/migration_runner_test.go`: add test case verifying `Run` fails with `adapters.ErrWorkspacesDirNotFound` when `WorkspacesDir` is empty and no configuration file is present.

#### Phase 7: CLI Integration Test Suite Verification (`backend/cmd/agentic-lens/main_test.go`)
- Inspect tests executing CLI without `--workspaces` flag. Ensure tests verifying missing configuration expectations exit with code 1 and clean error output rather than panicking.

---

### 5. Verification and Adversarial Testing Tasks for Tester
- Execute `go test -v -race ./backend/internal/adapters/...` to verify all adapter unit tests pass without panic.
- Adversarial test: Call `ResolveWorkspacesDir("")` inside an empty temporary directory with a local `./workspaces` directory present. Confirm it returns `ErrWorkspacesDirNotFound` and does not fall back to `./workspaces`.
- Adversarial test: Call `ResolveWorkspacesDir("")` inside an environment where `/app/workspaces` exists. Confirm it returns `ErrWorkspacesDirNotFound` rather than returning `/app/workspaces`.
- Adversarial test: Call `ScanWorkspaces("")` in an unconfigured directory. Confirm it returns a `domain.ScanReport` containing an error entry and zero ledgers without panicking.
- Adversarial test: Call `LoadTicketWorkspace("AL-1", "")` in an unconfigured directory. Confirm it returns a non-nil error wrapping `ErrWorkspacesDirNotFound`.
- Execute `go test -v -race ./backend/cmd/agentic-lens/...` to ensure CLI commands fail gracefully with status 1 on missing configuration.
- Execute root verification: `make check && make test && make build`.

---

### 6. Audit Criteria for Reviewer
- Spec Alignment: Did `ResolveWorkspacesDir` completely remove all fallbacks (cwd `./workspaces`, container `/app/workspaces`, parent traversal, and default string `workspaces`)?
- Panic Elimination: Are all occurrences of `panic` in `ResolveWorkspacesDir` removed?
- Error Typing: Is `ErrWorkspacesDirNotFound` exported and returned as a typed sentinel error?
- Signature Migration: Did all callers across `adapters`, `backfill`, and `cmd` properly adapt to `(string, error)`?
- Test Cleanliness: Are all adapter, backfill, and CLI tests passing with zero panics and full assertions against `ErrWorkspacesDirNotFound`?
- Quality Discipline: Are there zero emojis and zero `any` types introduced in Go code?
