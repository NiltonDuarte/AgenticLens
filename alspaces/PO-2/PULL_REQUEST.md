# PR: PO-2: Schema Expansion v2, Anomaly Engine, and Visual Waterfall

## Summary
Defines the product specification, architectural contracts, and 4 vertical tracer-bullet Jira tickets for Phase 2 of AgenticLens: expanding the telemetry ledger schema with versioning, detecting execution anomalies (consecutive tool failures, context saturation jumps, rework loops) via a pure Go domain rule engine, and delivering a visual trace waterfall timeline with an anomaly inspector in the TypeScript web UI.

## Ticket Breakdown
New tickets created under `workspaces/plan/to-do/`:
- `AL-6`: Versioned Schema Migration and Chronological Event Extraction
- `AL-7`: Domain Anomaly Detection and Rule Engine
- `AL-8`: Anomaly API Endpoints and Fleet Health Aggregation
- `AL-9`: Visual Trace Waterfall and Anomaly Inspector

## Architectural Notes
- Schema versioning invariant: Missing, null, or zero `schema_version` is recognized and decoded as version 1 (`v1`), ensuring complete backward compatibility with existing ledgers.
- Version 2 (`v2`) introduces an ordered chronological event stream capturing tool invocations, turn markers, and token snapshots.
- Deterministic anomaly rules evaluate inside pure functions in `internal/domain`, maintaining strict Functional Core, Imperative Shell (FCIS) separation.
- Zero any across Go and TypeScript.
- Zero emojis across all code, documentation, UI strings, and artifacts.

## Pointers
- Specification: `workspaces/PO-2/SPEC.md`
- Shared Roadmap Index: `workspaces/plan/TICKETS.md`
- Session Summary: `workspaces/PO-2/SUMMARY.md`
