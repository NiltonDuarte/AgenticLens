# Technical Contract: Ticket FIX-4

## Problem Statement
The project workspace folder was renamed from workspaces to alspaces and the global configuration in .agenticlens-config.yaml was updated accordingly. Consequently:
1. Backend test suites broke because parity_test.go and server_test.go relied on hardcoded references to the outer workspaces directory instead of isolated, self-contained fixtures.
2. The server command startup logic in backend/cmd/agentic-lens/main.go coupled workspaces directory resolution before validating project configuration and failed to cleanly derive the active workspaces folder from configuration when the command-line flag was omitted.

## Core Architectural Invariants

### 1. Test Isolation and Self-Containment Mandate
- No unit, integration, or parity test may depend on the outer repository workspace folder (alspaces or workspaces) or any live project telemetry ledgers.
- All test fixtures must reside strictly under the testdata directory or within ephemeral directories created via t.TempDir.
- Any test suite reading workspaces fixtures (including parity and HTTP boundary tests) must resolve paths relative to the test file location or package root pointing inside testdata.

### 2. Functional Core, Imperative Shell (FCIS) Boundaries
- Pure Functional Core (backend/internal/domain):
  - Domain models, aggregations, anomaly detection, and fleet derivations remain pure functions with zero filesystem or network interactions.
  - No changes permitted to pure mathematical rules.
- Imperative Adapter Shell (backend/internal/adapters):
  - Configuration loaders (backend/internal/adapters/config/loader.go) and repository scanners (backend/internal/adapters/repository.go) remain responsible for filesystem traversal and schema translation.
  - HTTP server constructors (backend/internal/adapters/http/server.go) must guarantee consistency between the resolved workspaces directory string and the ProjectConfig model passed to HTTP route handlers.
- CLI Entrypoint (backend/cmd/agentic-lens/main.go):
  - Command parsing and process lifecycle boundaries.
  - Must validate the existence of configuration before proceeding with server startup.

### 3. Server Startup and Configuration Resolution Rules
- Server startup must load the configuration file via adaptersconfig.Load before initializing filesystem scanners or HTTP servers.
- When the --workspaces flag is omitted or empty:
  - If no configuration file is present (adaptersconfig.ErrConfigNotFound), the server must immediately abort startup with an exit code of 1 and output an explicit error message directing the user to initialize the project.
  - If a configuration file is present, the server must extract the configured workspaces directory from the configuration model (workspaces.dir) and resolve it relative to the configuration file root.
- When the --workspaces flag is provided:
  - The server must prioritize the explicitly passed directory as the active workspaces directory.
  - If a configuration file is present, the configuration model must have its workspaces directory property synchronized to match the explicit flag value so that /api/config and downstream handlers report consistent paths.
  - If no configuration file is present, the server must fail startup with the standard missing configuration error.

## Scope Boundaries

### Go Backend Scope
- backend/cmd/agentic-lens/main.go: Refactor runServer to enforce configuration loading priority and conditional workspace resolution based on flag presence.
- backend/internal/adapters/http/server.go: Ensure NewServer and NewServerWithConfig synchronize the ProjectConfig Workspaces.Dir field with the active workspaces directory argument.
- backend/internal/adapters/parity_test.go: Decouple from live repository workspaces by redirecting ledger discovery to isolated test fixtures located in testdata/parity_workspaces.
- backend/internal/adapters/http/server_test.go: Update TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces to use an isolated test fixture rather than navigating upwards to the outer project workspace folder.
- testdata/: Populate testdata/parity_workspaces with the twelve reference ticket ledgers necessary for cross-engine parity and comparison validation.

### TypeScript Frontend Scope
- None. The TypeScript SPA shell (web/src) consumes configuration and workspaces dynamically through /api/config and /api/workspaces, and is already fully decoupled from local filesystem directory names.
