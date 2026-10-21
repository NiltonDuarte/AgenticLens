# feat(anomalies): Domain anomaly detection and rule engine (AL-7)

## Overview
Implements pure-functional execution anomaly detection and operational health rules across the Go backend, CLI diagnostics, markdown formatters, and TypeScript domain types for AL-7. During metric projection, the anomaly engine inspects tool retry loops, error rate spikes, context saturation jumps, and subagent ping-pong pinging to flag operational hazards.

For complete design details and lifecycle telemetry, see:
- Contract: [workspaces/AL-7/CONTRACT.md](workspaces/AL-7/CONTRACT.md)
- Telemetry Summary: [workspaces/AL-7/SUMMARY.md](workspaces/AL-7/SUMMARY.md)

## Summary of Changes

### Domain Core (`backend/internal/domain/`)
- Added `AnomalySeverity` (`WARNING`, `CRITICAL`) and `AnomalyRuleID` (`TOOL_RETRY_STREAK`, `TOOL_HIGH_FAILURE_RATE`, `CONTEXT_SATURATION_SPIKE`, `SUBAGENT_REWORK_LOOP`) enums in `enums.go`.
- Added immutable `Anomaly` and `AnomalyContext` structs, and added `Anomalies []Anomaly` to `AgenticMetricsLedger` in `models.go`.
- Implemented pure functional rule evaluators and engine in `anomalies.go`:
  - `TOOL_RETRY_STREAK`: Detects consecutive tool failures (warning at 3, critical at 5).
  - `TOOL_HIGH_FAILURE_RATE`: Detects elevated tool failure rates for tools with >= 5 calls (warning at >20%, critical at >=50%).
  - `CONTEXT_SATURATION_SPIKE`: Detects rapid context expansion (warning at >40k delta, critical at >80k delta or >=160k ceiling breach).
  - `SUBAGENT_REWORK_LOOP`: Detects ping-pong rework cycling between non-orchestrator subagent roles (warning at >=3 cycles, critical at >=5 cycles).
  - Deterministic multi-key sorting prioritizing severity descending, rule ID ascending, turn/tool ascending, title ascending.
- Wired anomaly evaluation into `DeriveMetricsLedger` in `aggregations.go`.
- Comprehensive adversarial unit tests in `anomalies_test.go`.

### Boundary Adapters (`backend/internal/adapters/`)
- Added `AnomalyContextDTO`, `AnomalyDTO`, and `AgenticMetricsLedgerDTO` with bidirectional domain conversion mappings in `schemas.go`.
- Updated `FormatMarkdownMetricsTable` and `UpdateSummaryFile` in `formatters.go` to render the `### Anomalies & Health Flags` section with severity badges and clean session status banners.

### CLI Application (`backend/cmd/agentic-lens/`)
- Added `--anomalies` flag to display formatted tabular anomaly diagnostics for target ticket workspaces.
- Added `Anomalies` summary count column to the `--list` command table.
- Added end-to-end integration tests in `main_test.go`.

### Frontend Contracts (`web/src/domain/`)
- Added `AnomalySeverity`, `AnomalyRuleID`, `AnomalyContext`, and `Anomaly` interfaces to `types.ts`, and updated `AgenticMetricsLedger` with optional `anomalies` slice for strict Go-TypeScript type parity.

### Documentation (`README.md`)
- Added documentation covering anomaly rules, severity thresholds, CLI diagnostic usage, and markdown summary generation.
