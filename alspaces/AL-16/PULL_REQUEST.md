# refactor(codebase): optimize test performance, modernize skills, and decouple adapters (AL-16)

## Overview
Comprehensive codebase refactoring, test performance optimizations, agent skill modernization, and architectural decoupling across AgenticLens. Decouples HTTP handlers from filesystem I/O via targeted workspace loading, extracts pure domain event sorting, eliminates duplicated invocation aggregation loops, encapsulates filesystem ledger persistence in repository adapters, parallelizes CLI subprocess tests, optimizes frontend fleet caching, removes machine-specific paths from agent personas, and centralizes frontend color tokens into a shared theme module.

- Contract: [workspaces/AL-16/CONTRACT.md](workspaces/AL-16/CONTRACT.md)
- Telemetry: [workspaces/AL-16/SUMMARY.md](workspaces/AL-16/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- Implemented `domain.SortEvents` in `events.go` with value-copy immutability, ascending timestamp/turn sorting, and stable ordering for ties.
- Eliminated triplicated token, cost, and cache loops across `ComputeTotals`, `ComputeRoleMetrics`, and `ComputeModelMetrics` via pure helper `accumulateInvocations` in `aggregations.go`.
- Added parallel unit tests in `events_test.go` and verified arithmetic parity in `aggregations_test.go`.

### Adapters & Decoupling Layer (`backend/internal/adapters/`)
- Refactored `WorkspaceEventsHandler` and `WorkspaceDetailHandler` in `http/handlers.go` to use targeted `adapters.LoadTicketWorkspace` rather than full directory traversals.
- Implemented `SaveFactLedger` and `SaveMetricsLedger` in `repository.go` with parent directory creation, indented JSON formatting, and trailing newlines, delegating ledger writes from `cmd/agentic-lens/main.go`.
- Relocated `UpdateSummaryFile` into `repository.go` to keep `formatters.go` 100% pure string formatting.
- Added unit tests in `repository_test.go`, `formatters_test.go`, and `http/server_test.go`.

### Frontend Presentation (`web/src/`)
- Optimized `TicketDeepDive.tsx` to cache fleet catalogs in memory, eliminating redundant `fetchWorkspaceLedgers` calls on dropdown ticket changes.
- Centralized theme and chart color tokens in `web/src/utils/theme.ts`.
- Consolidated hardcoded hex colors in `MetricProgression.tsx` and `SaturationBar.tsx`.
- Added unit tests in `TicketDeepDive.test.tsx` and `theme.test.ts`.

### Test Performance & Concurrency (`backend/`, `web/`)
- Added `t.Parallel()` across independent CLI subprocess tests in `cmd/agentic-lens/main_test.go` and unit tests.
- Replaced unconditional 100ms sleeps in server polling tests with low-latency 10ms retries.
- Added `// @vitest-environment node` to pure domain frontend test files.

### Agent Environment (`.opencode/`)
- Updated agent personas and dev skill with `backend/` directory prefixes.
- Removed hardcoded local developer paths from `planner.md` and `SKILL.md`.
- Standardized configuration references to `.agenticlens/config.yaml`.
