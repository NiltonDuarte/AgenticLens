ma# Pull Request: AL-17 Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry

## Overview
This PR implements live session discovery, hierarchical session tree reconstruction, title-based ticket inference, and transient trace streaming for AgenticLens.

## Key Changes
1. **Domain Functional Core**:
   - `backend/internal/domain/live.go`: Core data models (`LiveSessionNode`, `LiveSessionTree`, `LivePartFact`, `LiveMessageFact`, `LiveTraceView`), pure tool failure classifier (`ClassifyToolFailure`), title-based ticket inference heuristic (`InferTicketID`), and hierarchical tree builder (`BuildSessionTree`).
   - `backend/internal/domain/live_test.go`: Unit test suite covering failure classification severity tiers, case-insensitive title parsing, and tree hierarchy.
2. **SQLite Live Hook & Transient Ring Buffer**:
   - `backend/internal/adapters/live/sqlite.go`: Read-only hook using non-blocking pragmas (`query_only = ON`, `busy_timeout = 5000`, `read_uncommitted = ON`) and read-only DSN configuration to query active sessions and message/part facts without locking active agent sessions.
   - `backend/internal/adapters/live/cache.go`: Thread-safe transient ring buffer bounded by `live.max_trace_buffer_messages` guaranteeing strictly zero disk persistence.
   - `backend/internal/adapters/live/service.go`: Live observability coordinator providing background polling, status derivation, SSE channel subscription, and trace extraction.
   - `backend/internal/adapters/live/sqlite_test.go`: Boundary, concurrency, pragma enforcement, buffer boundary, and zero-disk persistence tests.
3. **HTTP REST & SSE API**:
   - `backend/internal/adapters/http/dto.go`: Type union expansion and typed response writers.
   - `backend/internal/adapters/http/live_handlers.go`: Handlers for `GET /api/live/sessions`, `GET /api/live/sessions/{id}/trace`, and `GET /api/live/stream`.
   - `backend/internal/adapters/http/server.go`: Endpoint registration in HTTP server.
   - `backend/internal/adapters/http/live_handlers_test.go`: Integration tests for live endpoints and SSE client disconnections.
4. **CLI Server Command Integration**:
   - `backend/cmd/agentic-lens/main.go`: Added `--db-path` flag to `server` command, automatic configuration fallback, and live service lifecycle management during server startup and shutdown.

## Verification
- `go test -race ./...`: All domain, live adapter, HTTP, and CLI tests pass.
- `go vet ./...`: 0 warnings.
- Zero `any` or empty `interface{}` across Go packages.
- Zero emojis across all code, tests, and documentation.
