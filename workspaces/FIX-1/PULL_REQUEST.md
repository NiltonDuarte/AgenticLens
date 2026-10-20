# fix(cli): enforce mandatory ticket title and id with diagnostic warnings (FIX-1)

## Overview
Hardens metadata validation in the `agentic-lens generate` CLI subcommand and persistence boundaries. Previously, invoking `generate` without `--ticket-title` or `--ticket-id` could result in `null` metadata fields being serialized into `agentic_metrics.json`, degrading downstream observability dashboards and summaries. Furthermore, missing ticket types and empty invocation lists lacked clear operator visibility.

The CLI now strictly fails with exit code 1 if `ticket-id` or `ticket-title` is missing and cannot be resolved from existing cached facts. It emits prominent warnings to standard error when `ticket-type` cannot be resolved in non-strict mode or when zero invocations are recorded. In addition, the `FIX` prefix is registered in `.agenticlens-config.yaml` to automatically infer conventional commit type `fix`.

- Contract: [workspaces/FIX-1/CONTRACT.md](workspaces/FIX-1/CONTRACT.md)
- Summary: [workspaces/FIX-1/SUMMARY.md](workspaces/FIX-1/SUMMARY.md)

## Summary of Changes

### CLI Application Layer (`backend/cmd/agentic-lens/`)
- Enforced mandatory `activeTicketID`: terminates with exit code 1 and writes `Error: missing ticket ID. Specify --ticket-id explicitly.` to standard error when absent.
- Enforced mandatory `activeTicketTitle`: trims whitespace, inspects cached ledger facts fallback, and terminates with exit code 1 writing `Error: missing ticket title for ticket '<ID>'. Specify --ticket-title explicitly.` to standard error when absent or blank.
- Preserved non-strict ticket type warning: outputs `WARNING: Missing ticket type for ticket <ID>. Specify --ticket-type explicitly.` on standard error without halting execution.
- Added empty invocations warning: outputs `WARNING: Zero invocations recorded for ticket <ID>. This is not expected.` on standard error when `len(invocations) == 0`.
- Added integration test suite `TestCLIGenerate_TicketMetadataValidation` in `main_test.go` covering all validation gates, fallback paths, and warnings.
- Updated existing integration test fixtures to provide `--ticket-title`.

### Boundary Adapters Layer (`backend/internal/adapters/`)
- Added defensive validation in `SaveFactLedger` (`repository.go`): prevents writing fact ledgers if `TicketID` or `TicketTitle` is nil, empty, or whitespace-only.
- Added unit tests in `repository_test.go` covering boundary rejection of invalid IDs and titles.
- Added `FIX` ticket prefix to `.agenticlens-config.yaml` with `ticket_type: fix`, validated via `loader_test.go`.

### Standards & Quality
- Adheres strictly to FCIS (Functional Core, Imperative Shell) architecture.
- Zero `any` or empty `interface{}` across Go packages.
- Zero emojis across all code, tests, logs, and artifacts.
