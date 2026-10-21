# Implementation Plan: FIX-2 Iteration 04 - Strict Fail-Fast Server Configuration and Explicit Domain Ledger Configuration Standard

## 1. Architectural Delta and Rationale

Iteration 04 addresses two foundational architectural hazards identified in the Go backend:

### Principle 1: Strict Fail-Fast HTTP Server Configuration Resolution
In iteration 03, the HTTP server constructor NewServer attempted to resolve project configuration by calling config.Load on the provided workspace directory or the current working directory. However, when both discovery attempts failed, NewServer silently fell back to domain.DefaultProjectConfig().

This silent fallback violates boundary discipline and operator trust:
- A bare default configuration lacks critical project context (such as custom ticket prefixes, specific role orderings, custom artifact definitions, and code metric file inclusion patterns).
- In production deployment, running a server without discovering its intended configuration hides deployment defects, missing volume mounts, and misconfigured environment paths behind a silent, degraded fallback.
- The configuration loader in backend/internal/adapters/config/loader.go is already responsible for discovering configuration files and running mergeConfig to populate baseline domain defaults for any fields omitted from disk.
- If no configuration file can be discovered or if parsing fails, NewServer must fail fast rather than silently substituting a synthetic default.
- The constructor signature for NewServer must therefore return a pair: a pointer to Server and an error.
- Callers requiring custom in-memory configuration without filesystem dependency (such as pure unit tests or pre-parsed CLI invocations) must explicitly invoke NewServerWithConfig, preserving clear behavioral separation between filesystem-discovering constructors and explicit dependency-injecting constructors.

### Principle 2: Strict Standardization on Explicit ProjectConfig in DeriveMetricsLedger
Previously, the domain package exposed two parallel functions for computing metric ledgers from atomic fact ledgers:
1. DeriveMetricsLedger: accepted an AgenticFactLedger and an optional variadic slice of RoleMappings. It constructed domain.DefaultProjectConfig() internally and dropped role ordering or other project configuration fields.
2. DeriveMetricsLedgerWithConfig: accepted an AgenticFactLedger and an explicit ProjectConfig.

Maintaining a parameterless or variadic variant that silently manufactures a default configuration violates Functional Core, Imperative Shell (FCIS) discipline:
- In pure functional cores, calculations should not guess or manufacture contextual projection rules via hidden defaults.
- Callers frequently invoked DeriveMetricsLedger under the assumption that it would apply project-wide settings, only to discover that custom role ordering was discarded because an empty baseline default was constructed under the hood.
- To eliminate ambiguity and enforce architectural rigor, DeriveMetricsLedger is standardized to strictly require ProjectConfig: func DeriveMetricsLedger(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger.
- The older DeriveMetricsLedgerWithConfig function is retained solely as an alias forwarding to DeriveMetricsLedger.
- All callers across the repository (filesystem scanner, schema backfill engine, schema migration runner, CLI generate command, and all domain and adapter test suites) are updated to supply ProjectConfig explicitly.

---

## 2. Core Domain Changes (backend/internal/domain/aggregations.go)

### Function Signature and Purity
- Update the primary derivation function in backend/internal/domain/aggregations.go:
  - Signature: func DeriveMetricsLedger(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger.
  - Remove the variadic roleMappings parameter entirely.
  - Remove the internal invocation of DefaultProjectConfig() within DeriveMetricsLedger.
- Purity and Boundary Invariants:
  - backend/internal/domain/aggregations.go must maintain strictly zero imports of adapters, operating system facilities, filesystem packages, network libraries, or external clocks.
  - Value semantics must be strictly observed: inputs factLedger and cfg are accepted by value; the derived AgenticMetricsLedger is returned by value.
  - Zero usage of any or empty interface{} constructs.

### Projection Implementation Mechanics
- Map extraction:
  - Extract mappings from cfg.RoleMappings. If cfg.RoleMappings is nil, initialize mappings to domain.RoleMappings{}.
- Invocation and event role resolution:
  - Iterate factLedger.Invocations, transforming each role via mappings.ResolveRole(inv.Role).
  - If factLedger.Events is non-nil, iterate factLedger.Events, transforming each event role via mappings.ResolveRole(ev.Role).
  - Preserve all untouched raw facts in the returned AgenticMetricsLedger.Facts structure to guarantee non-destructive auditing.
- Downstream rollup dispatching:
  - Pass mapped invocations to ComputeTotals(mappedInvocations).
  - Pass mapped fact ledger and cfg.RoleOrder to ComputeWorkflowEfficiency(mappedFactLedger, cfg.RoleOrder).
  - Pass mapped invocations and cfg.RoleOrder to ComputeRoleMetrics(mappedInvocations, cfg.RoleOrder).
  - Pass mapped invocations to ComputeModelMetrics(mappedInvocations).
  - Pass mapped invocations and cfg.RoleOrder to ComputeThreadSummaries(mappedInvocations, cfg.RoleOrder).
  - Pass mapped fact ledger to DetectAnomalies(mappedFactLedger).
- Struct composition:
  - Populate TicketID, TicketTitle, TicketType, StartedAt, CompletedAt, and UpdatedAt directly from factLedger.
  - Assign raw factLedger to Facts.
  - Assign computed rollups to WorkflowEfficiency, Threads, Totals, Roles, Models, and Anomalies.

### Backward Compatibility Alias
- Keep DeriveMetricsLedgerWithConfig in backend/internal/domain/aggregations.go as an exported 1-line forwarding alias:
  - Signature: func DeriveMetricsLedgerWithConfig(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger.
  - Implementation: return DeriveMetricsLedger(factLedger, cfg).
  - Add docstring noting that DeriveMetricsLedgerWithConfig is maintained for backwards compatibility and that DeriveMetricsLedger should be preferred.

---

## 3. HTTP Server Changes (backend/internal/adapters/http/server.go)

### Signature and Error Return Contract
- Modify the constructor signature of NewServer in backend/internal/adapters/http/server.go:
  - Updated Signature: func NewServer(workspacesDir string, staticFS fs.FS, liveService ...LiveServiceInterface) (*Server, error).
  - Return type transitions from a single pointer *Server to (*Server, error).

### Configuration Discovery and Fail-Fast Logic
- Step 1: Attempt configuration resolution from the specified workspacesDir via config.Load(workspacesDir).
- Step 2: If config.Load(workspacesDir) returns a non-nil error, attempt fallback resolution from the current directory via config.Load("").
- Step 3: If both configuration discovery attempts fail:
  - Do NOT assign domain.DefaultProjectConfig().
  - Immediately return nil and a formatted error wrapping the underlying failure (for example: fmt.Errorf("loading project configuration for server: %w", err)).
- Step 4: If configuration resolution succeeds:
  - Delegate instantiation to NewServerWithConfig(workspacesDir, staticFS, cfg, liveService...).
  - Return the instantiated *Server pointer and a nil error.

### Invariant on NewServerWithConfig
- Retain NewServerWithConfig unchanged:
  - Signature: func NewServerWithConfig(workspacesDir string, staticFS fs.FS, cfg domain.ProjectConfig, liveService ...LiveServiceInterface) *Server.
  - Behavior: accepts an explicit domain.ProjectConfig by value, builds router endpoints, registers SPA static serving, wraps CORS middleware, and returns *Server directly. It performs zero filesystem discovery.

---

## 4. Adapter Callers

### Caller 1: Workspace Repository Scanner (backend/internal/adapters/repository.go)
- ScanWorkspaces:
  - At line 120, configuration is loaded via cfg, _ := config.Load(basePath).
  - At line 149, update the metric ledger derivation from domain.DeriveMetricsLedgerWithConfig(factLedger, cfg) to domain.DeriveMetricsLedger(factLedger, cfg).
- LoadTicketWorkspace:
  - At line 306, configuration is loaded via cfg, _ := config.Load(basePath).
  - At line 307, update the metric ledger derivation from domain.DeriveMetricsLedgerWithConfig(factLedger, cfg) to domain.DeriveMetricsLedger(factLedger, cfg).

### Caller 2: Backfill Engine (backend/internal/adapters/backfill/engine.go)
- BackfillEngine.Execute:
  - In Execute(opts BackfillOptions), after resolving targetDir via adapters.ResolveWorkspacesDir(opts.WorkspacesDir), load project configuration via cfg, _ := config.Load(targetDir).
  - If loading returns an error, attempt fallback via cfg, _ = config.Load("").
  - Pass the resolved cfg down to processSingleWorkspace.
- BackfillEngine.processSingleWorkspace:
  - Extend parameter list to receive cfg domain.ProjectConfig.
  - At line 390, where SUMMARY.md updates are processed under opts.UpdateSummaries, replace domain.DeriveMetricsLedger(factLedger) with domain.DeriveMetricsLedger(factLedger, cfg).

### Caller 3: Universal Migration Runner (backend/internal/adapters/backfill/migration_runner.go)
- MigrationRunner.Run:
  - In Run(opts MigrationOptions), after resolving targetDir via adapters.ResolveWorkspacesDir(opts.WorkspacesDir), load project configuration via cfg, _ := config.Load(targetDir).
  - If loading returns an error, attempt fallback via cfg, _ = config.Load("").
  - At line 183, where SUMMARY.md updates are processed under opts.UpdateSummaries, replace domain.DeriveMetricsLedger(migrated) with domain.DeriveMetricsLedger(migrated, cfg).

### Caller 4: CLI Application Entrypoint (backend/cmd/agentic-lens/main.go)
- runGenerate:
  - At line 306, replace domain.DeriveMetricsLedgerWithConfig(factLedger, projCfg) with domain.DeriveMetricsLedger(factLedger, projCfg).
- runServer:
  - Line 434 already uses adaptershttp.NewServerWithConfig(workspacesDir, staticFS, projCfg, liveService). Verify this call remains intact, passing the pre-loaded projCfg directly to NewServerWithConfig.

---

## 5. Test Impact and Test Suites Updates

### Suite 1: Domain Aggregations (backend/internal/domain/aggregations_test.go)
- TestDeriveMetricsLedger_EventsPreservationAndInvariants:
  - Update all calls to DeriveMetricsLedger(ledgerNilEvents), DeriveMetricsLedger(ledgerEmptyEvents), DeriveMetricsLedger(ledgerWithEvents), DeriveMetricsLedger(baseLedger), and DeriveMetricsLedger(emptyLedger) to pass domain.DefaultProjectConfig() as the second parameter.
- TestDeriveMetricsLedger_RoleMappings:
  - In subtests with customMappings, construct cfg := domain.DefaultProjectConfig(), assign cfg.RoleMappings = customMappings, and invoke domain.DeriveMetricsLedger(factLedger, cfg).
- TestDeriveMetricsLedgerWithConfig_RoleOrder:
  - In custom role order subtests, replace domain.DeriveMetricsLedgerWithConfig(factLedger, cfg) with domain.DeriveMetricsLedger(factLedger, cfg).
  - In parity subtest, assert that domain.DeriveMetricsLedgerWithConfig(factLedger, cfg) returns an identical structure to domain.DeriveMetricsLedger(factLedger, cfg), verifying alias parity.
- TestDeriveMetricsLedger_SubagentClassificationUnderRoleMapping:
  - In Scenarios A, B, and C, construct cfg := domain.DefaultProjectConfig(), set cfg.RoleMappings = mappings, and invoke domain.DeriveMetricsLedger(baseLedger, cfg).
- TestDeriveMetricsLedger_AdversarialDeepImmutabilityAndEdgeCases:
  - In deep immutability tests, empty ledger test, zero invocation test, and role mapping override tests, construct explicit domain.ProjectConfig instances (using domain.DefaultProjectConfig() as the baseline) and pass to domain.DeriveMetricsLedger.
  - In the table-driven subagent classification matrix (lines 2740-2770), pass cfg := domain.DefaultProjectConfig() with cfg.RoleMappings = mappings directly to domain.DeriveMetricsLedger(factLedger, cfg).

### Suite 2: Domain Anomalies (backend/internal/domain/anomalies_test.go)
- TestDeriveMetricsLedger_IntegratesAnomalies:
  - At line 1749, update call to metrics := domain.DeriveMetricsLedger(factLedger, domain.DefaultProjectConfig()).
  - Verify that anomaly detection rules evaluate identically under the explicit default configuration.

### Suite 3: Domain Events (backend/internal/domain/events_test.go)
- TestPruneEventsForPersistence_Invariants:
  - At lines 611-612, update both derivations to pass domain.DefaultProjectConfig():
    - derivedUnpruned := domain.DeriveMetricsLedger(ledgerUnpruned, domain.DefaultProjectConfig())
    - derivedPruned := domain.DeriveMetricsLedger(ledgerPruned, domain.DefaultProjectConfig())
  - Verify parity checks between unpruned and pruned derivations remain completely satisfied.

### Suite 4: HTTP Server Handlers (backend/internal/adapters/http/server_test.go)
- Update all 16 existing NewServer call sites across server_test.go to handle the new (*Server, error) signature.
- Introduce a local test helper at the top of server_test.go:
  - Name: mustNewServer.
  - Signature: func mustNewServer(t *testing.T, workspacesDir string, staticFS fs.FS, liveService ...LiveServiceInterface) *Server.
  - Implementation: invokes adaptershttp.NewServer(workspacesDir, staticFS, liveService...); if err is non-nil, invokes t.Fatalf with descriptive failure details; returns *Server.
- Replace direct assignments server := adaptershttp.NewServer(...) with mustNewServer(t, ...) across:
  - TestWorkspacesHandler (line 165)
  - TestWorkspaceDetailHandler (line 239)
  - TestWorkspaceDetailHandler_EmptyWorkspacesDir (line 850)
  - TestWorkspaceDetailHandler_ZeroValueWorkspacesDir (line 950)
  - TestFleetHandler (line 1482)
  - TestWorkspaceEventsHandler (line 2077)
  - TestComparisonHandler (line 2285)
  - TestProgressionHandler (line 2486)
  - TestVersionHandler (line 2737)
  - TestSPAHandler_SinglePageAppRouting (lines 2941, 2987)
  - TestCORSMiddleware (line 3059)
  - TestConfigHandler_DefaultConfig (line 3157)
  - TestConfigHandler_NonGETMethods (line 3441)
  - TestServerWithTestDataWorkspaces (line 3462)
  - TestProductionWorkspacesStrictIsolation (line 3627)
- Add new adversarial test case TestNewServer_MissingConfigFailsFast:
  - Create an isolated temporary directory via t.TempDir().
  - Change working directory to the isolated temporary directory via t.Chdir(tempDir).
  - Invoke adaptershttp.NewServer(filepath.Join(tempDir, "workspaces"), fstest.MapFS{}).
  - Assert that returned *Server is nil.
  - Assert that returned error is non-nil and unwraps or matches config.ErrConfigNotFound.
  - Verify that no silent fallback to domain.DefaultProjectConfig() occurs.

### Suite 5: HTTP Live Handlers (backend/internal/adapters/http/live_handlers_test.go)
- In setupLiveTestServer(mock *mockLiveService) *httptest.Server:
  - Replace the single-return call server := adaptershttp.NewServer("workspaces", mockFS, svc) with:
    - server, err := adaptershttp.NewServer("workspaces", mockFS, svc)
    - If err is non-nil, fallback to adaptershttp.NewServerWithConfig("workspaces", mockFS, domain.DefaultProjectConfig(), svc) to guarantee test hermeticity, or fail fast if within an integration context.
  - This immediately satisfies all 11 test suites in live_handlers_test.go without signature churn across individual test cases.

---

## 6. Verification Gates

### Gate 1: Go Backend Compilation and Signature Uniformity
- Target: Verify that all packages compile cleanly under the new signatures.
- Execution: cd backend && go test -c ./internal/domain && go test -c ./internal/adapters/http && go test -c ./internal/adapters/backfill && go test -c ./cmd/agentic-lens
- Pass Criteria: Zero compilation errors; zero unused variables; zero signature mismatch diagnostics.

### Gate 2: Pure Domain Aggregations Suite
- Target: Validate DeriveMetricsLedger behavior under explicit ProjectConfig and verify alias parity.
- Execution: cd backend && go test -v -race -run "TestDeriveMetricsLedger" ./internal/domain/...
- Pass Criteria: All role mapping, role ordering, event preservation, immutability, and alias parity tests pass with zero failures.

### Gate 3: Domain Anomalies and Event Pruning Parity
- Target: Verify anomaly detection and event stream invariants under explicit configuration.
- Execution: cd backend && go test -v -race -run "TestDeriveMetricsLedger_IntegratesAnomalies|TestPruneEventsForPersistence" ./internal/domain/...
- Pass Criteria: Anomaly integration and persistence parity tests pass with zero regressions.

### Gate 4: HTTP Server Fail-Fast and Endpoint Suite
- Target: Validate that NewServer fails fast when config is missing and serves accurately when config is found.
- Execution: cd backend && go test -v -race -run "TestNewServer|TestConfigHandler|TestServerWithTestDataWorkspaces" ./internal/adapters/http/...
- Pass Criteria: TestNewServer_MissingConfigFailsFast passes, confirming no silent fallback. TestConfigHandler_DefaultConfig passes, confirming valid config loading.

### Gate 5: HTTP Live Handlers Integration
- Target: Validate live session endpoints under the updated server constructors.
- Execution: cd backend && go test -v -race ./internal/adapters/http -run "TestLive"
- Pass Criteria: All 11 live handler tests pass without panic or connection drop.

### Gate 6: Repository and Backfill Adapter Suites
- Target: Validate workspace scanning and summary backfilling under explicit configuration passing.
- Execution: cd backend && go test -v -race ./internal/adapters/...
- Pass Criteria: Scanner and backfill tests pass with zero errors.

### Gate 7: CLI Integration Suite
- Target: Validate generate, init, and server commands in the CLI package.
- Execution: cd backend && go test -v -race ./cmd/agentic-lens/...
- Pass Criteria: CLI tests pass, confirming flag parsing, ledger generation, and server wiring.

### Gate 8: Full Repository Verification Gate
- Target: Validate entire repository across formatting, typechecking, tests with race detection, and production build.
- Execution: make check && make test && make build
- Pass Criteria: Linting, formatting, backend tests, frontend Vitest suites, Vite build, and Go binary compilation succeed with exit code 0.

### Gate 9: Working Tree Non-Commit Invariant
- Target: Verify user directive to maintain working tree uncommitted.
- Execution: git status --porcelain
- Pass Criteria: Modified files exist in working tree for inspection. Zero new git commits created.

---

## 7. Reviewer Audit Checklist

### Axis 1: Spec and User Decision Alignment
- Check 1: Does NewServer return (*Server, error) instead of *Server?
- Check 2: Does NewServer fail fast and return an error when neither config.Load(workspacesDir) nor config.Load("") finds a configuration file?
- Check 3: Is domain.DefaultProjectConfig() completely removed as a silent fallback inside NewServer?
- Check 4: Does DeriveMetricsLedger strictly require ProjectConfig: func DeriveMetricsLedger(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger?
- Check 5: Is the parameterless/variadic DeriveMetricsLedger(factLedger, roleMappings...) eliminated?
- Check 6: Is DeriveMetricsLedgerWithConfig maintained as a pure forwarding alias to DeriveMetricsLedger?
- Check 7: Are all callers in repository.go, backfill/engine.go, backfill/migration_runner.go, and main.go updated to pass ProjectConfig explicitly?
- Check 8: Are all test suites in domain, http, and backfill updated to pass explicit configuration and handle (*Server, error)?
- Check 9: Did the coder refrain from authoring git commits, keeping the working tree uncommitted as instructed?

### Axis 2: Engineering Standards and Invariants
- Check 10: FCIS Boundary Discipline: Does backend/internal/domain/aggregations.go remain 100% pure, with zero imports of adapters, os, net/http, or external clocks?
- Check 11: Value Semantics: Does DeriveMetricsLedger accept inputs and return outputs by value without mutable reference leaks?
- Check 12: Strict Typing: Are there strictly zero uses of any or empty interface{} in Go code?
- Check 13: Zero Emojis Mandate: Are there zero emojis across all code files, comments, docstrings, plan files, and review reports?
- Check 14: Quality Gates: Do make check, make test, and make build pass cleanly with zero errors?
