# feat(api): anomaly API endpoints and fleet health aggregation (AL-8)

## Overview
Implements full-stack anomaly API endpoints and fleet health aggregations across Go backend domain, boundary adapters, HTTP server, and TypeScript web client for AL-8. Connects the pure-functional anomaly detection engine built in AL-7 to fleet-level health rollups, exposes workspace anomaly lists and chronological events timelines via HTTP, and ensures strict parity across the Go backend and TypeScript frontend.

For complete design details and lifecycle telemetry, see:
- Contract: [workspaces/AL-8/CONTRACT.md](workspaces/AL-8/CONTRACT.md)
- Telemetry Summary: [workspaces/AL-8/SUMMARY.md](workspaces/AL-8/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- Added `BrittleToolSummary` and `FleetAnomalySummary` structs to `models.go`.
- Added `AnomalySummary FleetAnomalySummary` to `FleetSummary` in `models.go`.
- Implemented pure-functional `ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary` in `anomalies.go`:
  - Computes fleet-wide anomaly totals, warning counts, critical counts, severity counts, and rule counts.
  - Computes total tickets, tickets with anomalies, and anomaly rate per ticket with zero-division safety.
  - Aggregates and stably sorts brittle tools by anomaly count descending, error count descending, failure rate percentage descending, and tool name ascending.
- Integrated anomaly rollup into `CalculateFleetSummary` in `aggregations.go`.
- Added comprehensive unit tests in `anomalies_test.go` covering empty inputs, clean ledgers, multi-ticket rollups, tie-breaking, and legacy v1 fallback.

### Boundary Adapters & HTTP Server (`backend/internal/adapters/`)
- Added `BrittleToolSummaryDTO` and `FleetAnomalySummaryDTO` with bidirectional domain conversion mappings in `schemas.go`.
- Added unit tests for bidirectional mapping and JSON round-tripping in `schemas_test.go`.
- Added `Anomalies []domain.Anomaly` to `WorkspaceDetailResponse` in `http/dto.go`, serialized as `[]` when empty (never omitted or null).
- Added `WriteEventsResponse` in `http/dto.go`.
- Registered `GET /api/workspaces/{ticketId}/events` in `http/server.go`.
- Implemented `WorkspaceEventsHandler` in `http/handlers.go`, returning chronologically sorted events (empty array for legacy v1 workspaces, 404 for missing workspaces).
- Updated `WorkspaceDetailHandler` in `http/handlers.go` to always return an initialized `Anomalies` array.
- Added HTTP endpoint integration tests in `http/server_test.go` verifying events response, 404 handling, anomalies array serialization, and path traversal guards.

### Frontend Contracts & API Client (`web/src/`)
- Added `BrittleToolSummary` and `FleetAnomalySummary` interfaces to `domain/types.ts`.
- Updated `FleetSummary` with `anomaly_summary: FleetAnomalySummary` and `WorkspaceDetail` with `anomalies: readonly Anomaly[]` in `domain/types.ts`.
- Implemented `fetchWorkspaceEvents(ticketId: string)` in `api/client.ts` with error handling, 404 handling, and mock fallback.
- Updated mock data in `api/mockData.ts` and test fixtures in `LandingPage.test.tsx` and `TicketDeepDive.test.tsx`.
- Added client unit tests in `api/client.test.ts` covering 200 responses, 404 handling, and network fallback.

### Verification
- `go test -count=1 ./...`: All Go packages pass cleanly.
- `npm run typecheck && npm test`: All 163 TypeScript tests pass across 12 test suites.
- `npm run build`: Production build compiles and copies embedded assets cleanly.
- `go build -o ../bin/agentic-lens ./cmd/agentic-lens`: Binary builds cleanly with embedded assets.
