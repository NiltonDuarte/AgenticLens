# feat(server): implement Go HTTP server, embedded static FS, and telemetry REST API

## Overview
This PR implements ticket AL-2: the native Go HTTP server, embedded static frontend filesystem container, and read-only telemetry REST API endpoints for AgenticLens. It enables serving both REST API endpoints and embedded static SPA assets from a single binary with zero Node.js dependencies at runtime.

For architecture and specifications, see [CONTRACT.md](workspaces/AL-2/CONTRACT.md).
For work breakdown and execution telemetry, see [SUMMARY.md](workspaces/AL-2/SUMMARY.md).

## Key Changes

### Domain Core (`internal/domain`)
- Added typed enums `TimeWindow` and `LedgerSortField` with string parsers and deterministic fallback handling.
- Added `WorkspaceDetail` struct encapsulating `AgenticMetricsLedger` and `TicketArtifacts`.
- Added pure functions `FilterLedgersByTimeWindow` and `SortLedgers` with value semantics, zero mutations, zero I/O, and reference timestamp parameters.

### Boundary Adapters (`internal/adapters`)
- Added embedded static filesystem container (`internal/adapters/web`) embedding `dist/index.html` via `embed.FS` and `fs.Sub`.
- Added HTTP REST adapter (`internal/adapters/http`):
  - Endpoints: `GET /api/workspaces` (with `window`, `sort`, `order` query filters), `GET /api/workspaces/{ticketId}` (with artifact reading and traversal protection), and `GET /api/fleet`.
  - CORS middleware supporting cross-origin requests for local Vite development (`OPTIONS` preflight returns 204).
  - SPA static file handler with client-side routing fallback to `index.html` for UI paths and strict JSON 404 guards for `/api/*` routes.
  - Strongly-typed response payload constraint interface `ResponsePayload` with zero `any` and handled JSON serialization errors.

### Presentation Shell (`cmd/agentic-lens`)
- Added `server` subcommand with `--port` (default 8080) and `--workspaces` flags.
- Added graceful shutdown on OS interrupt signals (`SIGINT`, `SIGTERM`).
- Preserved existing CLI scan and report commands (`--list`, `--ticket-id`, `--append`, `--json-output`) with zero regressions.

## Invariant Compliance
- Strict Functional Core, Imperative Shell (FCIS) separation: pure domain logic, zero outer imports, zero system clock calls in domain.
- Zero `any` and zero `interface{}` across Go backend and tests.
- Zero emojis across all code, log messages, error responses, and workspace artifacts.
