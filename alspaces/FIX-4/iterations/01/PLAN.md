# Iteration 01 Plan: Self-Contained Test Suites and Server Configuration Resolution

## Objective
Remediate broken backend test suites caused by the workspaces-to-alspaces rename, eliminate all test dependencies on the outer project directory by establishing isolated testdata fixtures, and align server command startup so that configuration is strictly enforced and used as the default workspace source when the command-line flag is absent.

## Grounded Context References
- backend/cmd/agentic-lens/main.go: runServer function handling CLI flag parsing, configuration loading, database hook setup, and server startup.
- backend/internal/adapters/http/server.go: NewServer and NewServerWithConfig constructor functions and route registrations.
- backend/internal/adapters/parity_test.go: Parity tests asserting ledger metrics against Python reference baselines.
- backend/internal/adapters/http/server_test.go: HTTP server integration tests, specifically TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces.
- testdata/workspaces: Existing five-ticket fixture directory containing AL-1, AL-2, AL-3, GEN-1, and PO-1.
- alspaces: Live repository workspaces directory containing thirty-plus ticket ledgers including reference tickets AL-10, AL-11, AL-12, AL-4, GEN-2, PO-2, and PO-3.

## Phased Implementation Tasks

### Phase 1: Test Fixture Isolation in testdata
1. Create a dedicated parity fixtures directory at testdata/parity_workspaces.
2. Copy the twelve reference ticket workspace fixtures (AL-1, AL-2, AL-3, AL-4, AL-10, AL-11, AL-12, GEN-1, GEN-2, PO-1, PO-2, PO-3) from alspaces into testdata/parity_workspaces. Each ticket folder must contain its agentic_metrics.json, CONTRACT.md, and SUMMARY.md artifacts.
3. Preserve the existing testdata/workspaces folder with its original five tickets (AL-1, AL-2, AL-3, GEN-1, PO-1) so that existing tests asserting exact five-ticket counts remain completely stable.

### Phase 2: Decouple backend/internal/adapters/parity_test.go
1. Remove the findWorkspacesDir function candidate list that searched for ../../workspaces and ../../../workspaces.
2. Implement a self-contained helper function in parity_test.go that resolves the absolute path to testdata/parity_workspaces relative to the test file location.
3. Update TestCrossEngineParityOnWorkspaceLedgers, TestFleetSummaryAndComparisonParity, and TestStructuralClassificationParityOnWorkspaces to execute against testdata/parity_workspaces.
4. Verify that all twelve reference tickets evaluate with zero scan errors and achieve full numerical parity against the embedded reference JSON.

### Phase 3: Decouple backend/internal/adapters/http/server_test.go
1. Refactor TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces in server_test.go.
2. Replace the hardcoded path ../../../../workspaces with a path pointing to testdata/parity_workspaces (which contains AL-10 and more than five ledgers).
3. Ensure the test verifies isolation between testdata/workspaces and testdata/parity_workspaces without ever traversing to root project workspace directories.

### Phase 4: Server Startup Configuration Resolution in backend/cmd/agentic-lens/main.go
1. Modify runServer in main.go to restructure the startup sequence:
   - Attempt to load project configuration via adaptersconfig.Load immediately after parsing flags.
   - If loading configuration yields adaptersconfig.ErrConfigNotFound:
     - Print the standard error message instructing the user to run agentic-lens init.
     - Terminate the process with exit code 1.
   - If loading configuration fails with any other error:
     - Print the configuration loading error message.
     - Terminate the process with exit code 1.
2. Resolve the active workspaces directory based on flag presence:
   - If the --workspaces flag is provided and non-empty:
     - Use the flag value as workspacesDir.
     - Synchronize the loaded project configuration by setting projCfg.Workspaces.Dir to workspacesDir.
   - If the --workspaces flag is empty or omitted:
     - Retrieve the directory name from projCfg.Workspaces.Dir.
     - Resolve the path relative to the configuration file location or current working directory.
     - Assign this resolved path to workspacesDir.
3. Pass the resolved workspacesDir and synchronized projCfg into adaptershttp.NewServerWithConfig.

### Phase 5: Adapter Server Configuration Synchronization in backend/internal/adapters/http/server.go
1. In NewServer, inspect the incoming workspacesDir parameter:
   - If workspacesDir is non-empty, ensure that the loaded configuration model has its Workspaces.Dir field overridden with workspacesDir.
   - If workspacesDir is empty and a configuration is successfully loaded, populate workspacesDir from the configuration model.
2. Ensure that handlers registered inside NewServerWithConfig receive a consistent workspaces directory across both /api/config and data-serving endpoints.

### Phase 6: Full Verification and Quality Gate
1. Execute the entire Go backend test suite across all packages (internal/domain, internal/adapters, cmd/agentic-lens) to verify zero failures.
2. Run end-to-end CLI tests in main_test.go verifying:
   - Server fails when invoked in a directory without configuration.
   - Server succeeds when started without --workspaces flag, correctly reading workspaces from .agenticlens-config.yaml.
   - Server succeeds when started with an explicit --workspaces flag, overriding the configured directory.
3. Run the TypeScript test suite (npm test in web/) to ensure complete frontend contract compatibility.
