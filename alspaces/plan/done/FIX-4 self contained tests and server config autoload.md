# FIX-4: Self-Contained Test Suites and Server Configuration Autoload

## Metadata
- **Ticket ID**: FIX-4
- **Subsystem**: Domain / Adapters / HTTP / CLI / Tests
- **Status**: Done
- **Type**: fix

## Problem Statement
Renaming the workspace folder to `alspaces` caused parity and HTTP server tests to break due to hardcoded dependencies on outer `workspaces/` paths. Furthermore, the `server` CLI command failed to load configuration if `--workspaces` was not passed, and did not properly enforce failure when no configuration file was present.

## Solution
1. Decouple all tests from outer workspace directories by placing isolated fixtures in `testdata/parity_workspaces/`.
2. Restructure server startup in `main.go` to unconditionally load `.agenticlens-config.yaml` and fail if absent.
3. Automatically resolve the workspace directory from `workspaces.dir` in configuration when the `--workspaces` flag is omitted.
