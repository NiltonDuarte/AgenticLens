# feat(events): Chronological event extraction and schema evolution in Go (AL-6)

## Overview
Implements schema evolution and chronological event extraction across Go backend and TypeScript frontend for AL-6. Extracts granular event facts (`turn_message`, `tool_call`, `compaction`) from OpenCode SQLite databases with microsecond ISO 8601 timestamps, context token tracking, error message excerpts, and execution durations.

For complete design details and lifecycle telemetry, see:
- Contract: [workspaces/AL-6/CONTRACT.md](workspaces/AL-6/CONTRACT.md)
- Telemetry Summary: [workspaces/AL-6/SUMMARY.md](workspaces/AL-6/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- Added `EventType` string enum with constants (`EventTypeToolCall`, `EventTypeTurnMessage`, `EventTypeCompaction`) in `enums.go`.
- Added immutable `EventFact` struct and optional `Events []EventFact` to `AgenticFactLedger` in `models.go`.

### Boundary Adapters (`backend/internal/adapters/`)
- Added `EventFactDTO` and bidirectional conversion mapping in `schemas.go`.
- Updated `AgenticFactLedgerDTO` and `ExistingLedgerFacts` in `repository.go` to support backward-compatible decoding of ledgers lacking events and round-trip serialization of populated events.
- Updated `opencode.ExtractFacts` in `extractor.go` to extract ordered `turn_message`, `tool_call`, and `compaction` events with durations, status, error excerpts, and context tokens. Added schema introspection for SQLite column compatibility.
- Updated `backfill.Engine` in `engine.go` to detect ledgers needing events under `--recompute-db`, extract chronological events from local SQLite sessions, and persist updated schema version 2 ledgers non-destructively.

### CLI Application (`backend/cmd/agentic-lens/`)
- Updated `generate` subcommand in `main.go` to persist extracted events into `agentic_metrics.json`.

### Frontend Contracts (`web/src/domain/`)
- Added `EventType` union and `EventFact` interface to `types.ts`, and updated `AgenticFactLedger` with optional `events` array maintaining strict type parity.

### Existing Workspaces Backfill
- Backfilled chronological events across all existing workspaces in `workspaces/` with matching local OpenCode SQLite session records.
