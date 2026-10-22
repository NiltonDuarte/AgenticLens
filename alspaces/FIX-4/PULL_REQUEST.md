# fix(server,tests): self contained test suites and server config autoload (FIX-4)

## Overview
Remediates broken backend test suites caused by renaming the project workspace directory to `alspaces` and isolates tests from outer filesystem structures. Enforces strict server configuration loading at startup so that `agentic-lens server` uses the configured workspace directory when `--workspaces` is omitted and fails immediately if no configuration file is present or if `workspaces.dir` is unconfigured. Completely eliminates hardcoded `"workspaces"` fallbacks and purges obsolete `.agenticlens` directory references.

- Contract: [alspaces/FIX-4/CONTRACT.md](alspaces/FIX-4/CONTRACT.md)
- Summary: [alspaces/FIX-4/SUMMARY.md](alspaces/FIX-4/SUMMARY.md)

## Summary of Changes

### Boundary Adapters (`backend/internal/adapters/`)
- In `config/loader.go`, removed default `"workspaces"` string from `mergeConfig`, leaving unconfigured `Workspaces.Dir` empty. Removed legacy `.agenticlens` subfolder candidate lookups in `FindConfigFile`.
- In `repository.go`, eliminated fallback to `"workspaces"` in `ResolveWorkspacesDir`, returning `""` when unconfigured. Removed container probe `/app/workspaces` and `.agenticlens` directory checks. Added empty workspace directory guard in `ScanWorkspaces`.
- In `parity_test.go`, replaced external workspace search with `findParityWorkspacesDir()`, resolving isolated test fixtures inside `testdata/parity_workspaces` relative to the test file location via `runtime.Caller(0)`.
- In `http/server_test.go`, updated `TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces` to target `testdata/parity_workspaces` instead of outer `../../../../workspaces`.
- Added `TestServerWorkspaceDirConfigSynchronization` to verify consistency between workspace path arguments and `ProjectConfig.Workspaces.Dir`.
- In `http/server.go`, synchronized `cfg.Workspaces.Dir` with the active workspace directory in `NewServer` and `NewServerWithConfig`.

### CLI & Presentation Layer (`backend/cmd/agentic-lens/`, `testdata/`)
- In `main.go`, eliminated `if configuredDir == "" { configuredDir = "workspaces" }` and obsolete `.agenticlens` directory checks. Process fails with exit code 1 when `workspaces.dir` is unconfigured and `--workspaces` is omitted.
- In `main_test.go`, added integration assertions verifying startup failure when configuration or `workspaces.dir` is missing and correct workspace resolution with and without `--workspaces`.
- Created `testdata/parity_workspaces/` containing 12 reference ticket fixtures (`AL-1`, `AL-2`, `AL-3`, `AL-4`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) for cross-engine parity validation, keeping `testdata/workspaces/` intact for 5-ticket assertions.
