# fix(server,tests): self contained test suites and server config autoload (FIX-4)

## Overview
Remediates broken backend test suites caused by renaming the project workspace directory to `alspaces` and isolates all tests from the repository outer filesystem structure. Enforces strict server configuration loading at startup so that `agentic-lens server` uses the configured workspace directory when `--workspaces` is omitted and fails immediately if no configuration file is present.

- Contract: [alspaces/FIX-4/CONTRACT.md](alspaces/FIX-4/CONTRACT.md)
- Summary: [alspaces/FIX-4/SUMMARY.md](alspaces/FIX-4/SUMMARY.md)

## Summary of Changes

### Boundary Adapters (`backend/internal/adapters/`)
- In `backend/internal/adapters/parity_test.go`, replaced external workspace search with `findParityWorkspacesDir()`, resolving isolated test fixtures inside `testdata/parity_workspaces` relative to the test file location via `runtime.Caller(0)`.
- In `backend/internal/adapters/http/server_test.go`, updated `TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces` to target `testdata/parity_workspaces` instead of navigating to outer `../../../../workspaces`.
- Added `TestServerWorkspaceDirConfigSynchronization` to verify consistency between workspace path arguments and `ProjectConfig.Workspaces.Dir`.
- In `backend/internal/adapters/http/server.go`, synchronized `cfg.Workspaces.Dir` with the active workspace directory in `NewServer` and `NewServerWithConfig`.

### CLI & Presentation Layer (`backend/cmd/agentic-lens/`, `testdata/`)
- In `backend/cmd/agentic-lens/main.go`, restructured `runServer` to invoke `adaptersconfig.Load("")` immediately upon execution. If no configuration file is found (`ErrConfigNotFound`), the process terminates with exit code 1 directing the user to run `agentic-lens init`.
- When `--workspaces` is omitted, `runServer` derives the workspace directory from `projCfg.Workspaces.Dir` relative to the config file path.
- When `--workspaces` is provided, `runServer` prioritizes the flag and synchronizes `projCfg.Workspaces.Dir`.
- In `backend/cmd/agentic-lens/main_test.go`, added integration assertions verifying startup failure when configuration is missing and correct workspace resolution with and without the `--workspaces` flag.
- Created `testdata/parity_workspaces/` containing 12 reference ticket fixtures (`AL-1`, `AL-2`, `AL-3`, `AL-4`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) for cross-engine parity validation, keeping `testdata/workspaces/` intact for 5-ticket assertions.
