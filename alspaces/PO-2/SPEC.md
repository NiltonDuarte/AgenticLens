# Feature Specification: Schema Expansion v2, Anomaly Engine, and Trace Waterfall

## Problem Statement
The current telemetry ledger (`agentic_metrics.json`) only captures aggregate rollups per subagent invocation and aggregate call/error counts per tool. It lacks chronological sequence ordering and schema versioning. Consequently, the platform cannot distinguish between three isolated tool errors spread across 100 turns versus a catastrophic three-turn consecutive retry loop. Furthermore, developers inspecting complex agent runs cannot pinpoint exactly when context window saturation spiked or which tool calls triggered rework cycles.

Introducing schema versioning with explicit backward compatibility (unversioned payloads defaulted to version 1) alongside a version 2 schema capturing chronological events enables deterministic anomaly detection (repeated tool failure streaks, context spikes, agent oscillation) and powers a visual trace waterfall timeline.

## Target Personas & Use Cases
- **AI Systems Engineers & Agent Authors**: Need deterministic alerts when agents enter infinite tool retry loops, thrash context windows, or repeat failing commands.
- **Tech Leads & Fleet Reviewers**: Need a turn-by-turn trace waterfall to inspect agent decision trajectories, understand why costs spiked, and identify brittle tools across tickets.

## User Stories
1. As an AI systems engineer, I want `agentic-lens` to parse both legacy unversioned ledgers (as v1) and versioned ledgers (v2) without breaking backward compatibility.
2. As an agent developer, I want AgenticLens to flag consecutive tool failure streaks (e.g. 3+ failed commands in a row) and context saturation spikes so that I can catch stuck runs and prompt regressions.
3. As a tech lead, I want to view a chronological trace waterfall in the web dashboard with clear anomaly indicators and an inspector drawer for error logs.

## Functional Requirements
- **Schema Versioning & Event Extraction (v2)**:
  - Add `schema_version` integer field to `agentic_metrics.json`.
  - Treat missing/omitted or null `schema_version` as Version 1 (`v1`).
  - Define Version 2 (`v2`) schema adding chronological `events` array containing: timestamp, session ID, turn number, role, event type (`tool_call`, `turn_start`, `compaction`), tool name, arguments summary, status (`success` or `error`), output snippet, duration ms, and context token count at event time.
  - Update `scripts/extract_agentic_metrics.py` to extract chronological tool events from SQLite database into v2 ledgers.
- **Go Domain Anomaly Engine**:
  - Implement pure functions in `internal/domain` to evaluate deterministic anomaly rules:
    - `TOOL_RETRY_STREAK`: 3+ consecutive errors for the same tool or command.
    - `TOOL_HIGH_FAILURE_RATE`: Tool error rate exceeding 20% with minimum 5 calls.
    - `CONTEXT_SATURATION_SPIKE`: Single-turn context jump exceeding 40k tokens or reaching `CRITICAL` (>160k tokens) ceiling.
    - `SUBAGENT_REWORK_LOOP`: Repeated cyclic invocations of the same worker roles (e.g., coder -> reviewer -> coder -> reviewer) exceeding threshold.
  - Return typed `[]Anomaly` structures with severity (`WARNING`, `CRITICAL`), target entity, and contextual explanation.
- **CLI & Presentation**:
  - Update Go CLI (`cmd/agentic-lens`) to display anomaly summaries in terminal output and append an "Anomalies & Health Flags" section to Markdown tables.
- **HTTP API & Endpoints**:
  - Expose anomalies on `/api/workspaces/:ticketId`.
  - Expose chronological events on `/api/workspaces/:ticketId/events`.
  - Aggregate fleet-level anomaly rates on `/api/fleet`.
- **TypeScript Visual Trace Waterfall**:
  - Implement turn-by-turn waterfall timeline in the Ticket Deep Dive view.
  - Render amber/red anomaly warning badges directly on affected turns and tool executions.
  - Provide an expandable inspector drawer showing raw error messages, tool inputs/outputs, and token increments.

## Non-Functional & Architectural Invariants
- **Layer Flow (FCIS)**: Anomaly detection rules remain pure functions in `internal/domain/`. Adapters in `internal/adapters/` isolate schema versions and disk/HTTP I/O. Presentation in `web/` and `cmd/agentic-lens/` consumes typed domain models.
- **Backward Compatibility**: Any `agentic_metrics.json` lacking `schema_version` must be transparently recognized and decoded as version 1 without error.
- **Strict Typing**: Zero `any` or empty `interface{}` across Go and TypeScript.
- **Zero Emojis**: Strictly no emojis in code, UI badges, tables, logs, or workspace markdown.
- **Performance**: Zero re-computation overhead; streaming chronological event parsing with bounded memory allocations.

## Decomposition & Delivery Roadmap
- [ ] Ticket AL-6: `AL-6: Versioned Schema Migration and Chronological Event Extraction` (Unblocked)
- [ ] Ticket AL-7: `AL-7: Domain Anomaly Detection and Rule Engine` (Blocked by AL-6)
- [ ] Ticket AL-8: `AL-8: Anomaly API Endpoints and Fleet Health Aggregation` (Blocked by AL-7)
- [ ] Ticket AL-9: `AL-9: Visual Trace Waterfall and Anomaly Inspector` (Blocked by AL-8)
