# Feature Specification: Python to Go + TypeScript Rewrite

## Problem Statement
The current Python application depends on Streamlit, Pandas, and Pydantic. It incurs high startup latency, requires an external Python 3.14 runtime environment with heavy dependencies, and lacks strict client-server decoupling. Rewriting the system into a compiled Go core and a static TypeScript frontend creates a standalone, single-binary distribution with sub-millisecond aggregation latency, zero Node.js/Python production dependencies, and strict type safety across all boundaries.

## Target Personas & Use Cases
- **AI Systems Engineers & Agents**: Need instant CLI generation of Markdown summary tables (`--append workspaces/<ID>/SUMMARY.md`) without runtime overhead.
- **Tech Leads & Fleet Operators**: Need an interactive, low-latency web dashboard to inspect context window saturation, token costs, tool failure distributions, and multi-agent turn velocities across hundreds of tickets.

## User Stories
1. As an AI engineer or orchestrator agent, I want to run `agentic-lens --append workspaces/AL-42/SUMMARY.md` to update session summaries instantly without loading Python interpreters.
2. As an engineering manager, I want to launch `agentic-lens server --port 8080` to inspect fleet KPIs and ticket deep dives from an embedded web dashboard without running Vite or Node.js in production.
3. As a developer debugging token bloat, I want to examine context health metrics, compaction events, and subagent turn distributions with zero presentation lag.

## Functional Requirements
- **Data Ingestion & Extraction**: Scan `workspaces/<TICKET_ID>/agentic_metrics.json` ledgers, parse atomic facts (`InvocationFact`, `ToolFact`, `CodeFacts`), and load workspace artifacts (`SUMMARY.md`, `PULL_REQUEST.md`) in `internal/adapters`.
- **Domain Aggregations**: Pure functional rollups in `internal/domain` for:
  - Totals: token breakdown, cache hit rates, cost derivations, context health categories (`HEALTHY`, `ELEVATED`, `BLOATED`, `CRITICAL`).
  - Workflow Efficiency: active durations, role effort distribution, subagent turn velocity, test-to-code ratios, tool success rates.
  - Fleet Summaries & Time Buckets: daily, weekly, cumulative rollups across date presets (`Last 7 Days`, `Last 30 Days`, `Project Lifetime`).
- **Visual Presentation**:
  - Go CLI (`cmd/agentic-lens`): commands for `scan`, `--list`, `--ticket-id`, `--json-output`, `--append`, and `server`.
  - Go HTTP Server: REST endpoints (`/api/workspaces`, `/api/workspaces/:ticketId`, `/api/fleet`) and SPA static asset handler using `embed.FS`.
  - TypeScript SPA (`web/`): Vite + React with strict TypeScript, displaying Fleet Overview, Ticket Deep Dive, Metric Progression, and Ticket Comparisons.

## Non-Functional & Architectural Invariants
- **Layer Flow (FCIS)**: Go domain remains pure (`internal/domain/`); Go adapters isolate filesystem, HTTP, and schemas (`internal/adapters/`); TypeScript/CLI handles presentation (`web/`, `cmd/agentic-lens/`).
- **Language Standards**: Go 1.23+ value semantics, TypeScript 5.x+ strict mode, zero any across both.
- **Single Binary Distribution**: Compiled TypeScript frontend (`web/dist`) embeds directly into the Go executable via `embed.FS`. Zero Node.js runtime in production.
- **Zero Emojis**: Strictly no emojis in code, UI strings, tables, logs, or workspace markdown.
- **Performance**: Sub-millisecond aggregation latency, memory-efficient data structures, zero redundant recomputation.

## Decomposition & Delivery Roadmap
- [ ] Ticket 1: `AL-1: Implement Go Domain Aggregations, Workspace Scanner, and CLI Engine` (Unblocked)
- [ ] Ticket 2: `AL-2: Implement Go HTTP Server, Embedded Static FS, and Telemetry REST API` (Blocked by AL-1)
- [ ] Ticket 3: `AL-3: Build TypeScript SPA Shell and Fleet Overview Dashboard` (Blocked by AL-2)
- [ ] Ticket 4: `AL-4: Implement Ticket Deep Dive, Progression Charts, and Comparisons` (Blocked by AL-3)
- [ ] Ticket 5: `AL-5: Integrate Single Binary Build, Verify Parity, and Deprecate Python App` (Blocked by AL-4)
