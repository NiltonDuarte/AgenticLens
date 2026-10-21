# Contract: AL-6 Chronological Event Extraction and Schema Evolution in Go

## 1. Problem Statement
Active session telemetry in AgenticLens is generated and backfilled natively by the Go engine across backend/cmd/agentic-lens, backend/internal/adapters/opencode, and backend/internal/adapters/backfill. While workspace ledgers (agentic_metrics.json) are already at schema version 2 with migration metadata, they record only coarse aggregate numbers per invocation and overall tool execution and error counts.

Detecting fine-grained execution anomalies (such as consecutive tool failure streaks, subagent rework loops, and single-turn context window spikes planned in AL-7, AL-8, and AL-9) requires ordered, turn-level event telemetry.

Furthermore, existing workspaces on disk (under workspaces/) currently lack this ordered chronological event stream. To prevent fleet telemetry fragmentation, the new chronological event extraction must not only serve future sessions, but must also be actively backfilled across all existing workspaces where sessions exist in the OpenCode SQLite database.

Ticket AL-6 delivers native Go chronological event extraction, schema evolution, and workspace backfill:
1. Defines pure domain event models (EventFact) representing atomic turn events, tool executions, and compaction milestones.
2. Updates OpenCode SQLite extraction in backend/internal/adapters/opencode/extractor.go to extract, correlate, and chronologically sort events from session, message, and part tables.
3. Provides seamless bidirectional serialization and backward compatibility in backend/internal/adapters/schemas.go and backend/internal/adapters/repository.go, ensuring historical ledgers lacking events deserialize cleanly without regression.
4. Integrates event capture into both agentic-lens generate and agentic-lens backfill --recompute-db.
5. Backfills existing workspaces on disk to populate chronological events from local OpenCode SQLite session traces.
6. Establishes TypeScript domain contract parity in web/src/domain/types.ts.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation must strictly comply with architectural boundaries established in TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
Dependencies flow strictly inward across layers:
1. CLI Subcommand Layer (backend/cmd/agentic-lens/main.go):
   - Application entrypoint, CLI flag parsing (generate, backfill, server), console output rendering, and adapter wiring.
2. Imperative Adapters Layer (backend/internal/adapters/):
   - External boundary serialization and DTO mapping (backend/internal/adapters/schemas.go).
   - Filesystem workspace traversal and raw ledger reading (backend/internal/adapters/repository.go).
   - SQLite read-only queries and extraction from OpenCode database (backend/internal/adapters/opencode/).
   - Multi-workspace backfill coordination (backend/internal/adapters/backfill/).
3. Functional Core Domain Layer (backend/internal/domain/):
   - Houses pure domain models (EventFact, AgenticFactLedger, AgenticMetricsLedger), typed event enums (EventType), and pure aggregation algorithms.
   - Must contain zero disk I/O, zero network calls, zero SQL queries, and zero system clocks. All timestamps, durations, and external identifiers must be passed as arguments from the imperative shell.

### Pure Domain Core Isolation
1. EventType is defined as a typed string enum in backend/internal/domain/enums.go with explicit constants for tool calls, turn messages, and compactions.
2. EventFact is an immutable domain value object capturing timestamp, session identifier, turn index, agent role, event type, optional tool name, execution status, optional error message, optional duration in milliseconds, and context token count.
3. Adding Events []EventFact to AgenticFactLedger in backend/internal/domain/models.go must not mutate or invalidate any existing pure mathematical aggregations in backend/internal/domain/aggregations.go.
4. Domain logic remains deterministic and resilient against zero values, empty event slices, and nil pointers.

### Imperative Boundary Shell Rules
1. Seamless Backward Compatibility Mandate: Existing historical ledgers lacking an events array must deserialize without error, panic, or data loss. In Go, omitted events fields must map to an empty or nil slice. In existing aggregations, ledgers without events must behave identically to prior releases.
2. Non-Destructive Preservation Mandate: If agentic-lens backfill or agentic-lens generate encounters an existing ledger on disk, but local OpenCode SQLite session records are missing on that machine, pre-existing invocations, tools, and events must be preserved intact rather than erased.
3. Workspace Backfill Mandate: The backfill engine (backend/internal/adapters/backfill/engine.go) must be upgraded to extract and populate chronological events when --recompute-db is specified. All existing workspaces on disk must be backfilled to populate events where session records reside in the local database.
4. Read-Only SQLite Access: SQLite queries across session, message, and part tables must execute in read-only mode (mode=ro), preventing locks against concurrent development processes.
5. Static Single-Binary Compilation: SQLite access relies exclusively on the pure Go driver modernc.org/sqlite with CGO_ENABLED=0.
6. Strict Zero-Emoji Mandate: Zero emojis across log statements, CLI outputs, generated JSON files, summary tables, and code comments.
7. Zero any or empty interface{} across all Go backend and TypeScript frontend code.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Domain Core Models (backend/internal/domain/):
   - Define EventType enum type in backend/internal/domain/enums.go with values: tool_call, turn_message, and compaction.
   - Define EventFact struct in backend/internal/domain/models.go.
   - Extend AgenticFactLedger in backend/internal/domain/models.go with optional Events []EventFact.
2. Boundary DTOs and Serialization (backend/internal/adapters/):
   - Implement EventFactDTO in backend/internal/adapters/schemas.go with bidirectional mapping to domain.EventFact.
   - Update AgenticFactLedgerDTO in backend/internal/adapters/schemas.go with Events []EventFactDTO (json tag events,omitempty).
   - Update backend/internal/adapters/repository.go (ExistingLedgerFacts and rawExistingLedger) to read and retain events when loading ledgers.
3. OpenCode SQLite Extractor (backend/internal/adapters/opencode/):
   - Update backend/internal/adapters/opencode/extractor.go to query messages and parts for resolved sessions, construct chronologically sorted EventFact items, calculate turn indices, capture tool execution status and error messages, and calculate duration in milliseconds.
4. CLI Subcommands and Backfill Integration:
   - Wire event extraction into runGenerate in backend/cmd/agentic-lens/main.go.
   - Wire event extraction into processSingleWorkspace in backend/internal/adapters/backfill/engine.go when --recompute-db is executed.
   - Ensure the backfill engine recognizes ledgers lacking events as candidates for DB recomputation when --recompute-db is enabled or when --force is passed.
5. Existing Workspace Backfill Execution:
   - Run backfill across all target workspace directories on disk (workspaces/) to upgrade existing ledgers with chronological events.
6. TypeScript Domain Contract (web/src/domain/):
   - Add EventFact interface in web/src/domain/types.ts.
   - Add optional readonly events?: readonly EventFact[]; to AgenticFactLedger in web/src/domain/types.ts.
7. Automated Unit and Adversarial Test Suites:
   - Unit tests in backend/internal/adapters/opencode/extractor_test.go asserting accurate event extraction, chronological sort ordering, turn correlation, tool error capture, and compaction detection from SQLite fixtures.
   - Unit tests in backend/internal/adapters/repository_test.go and backend/internal/adapters/schemas_test.go validating backward-compatible decoding of legacy files and round-trip fidelity of files with events.
   - Integration tests in backend/cmd/agentic-lens/main_test.go and backend/internal/adapters/backfill/engine_test.go.

### Out of Scope
1. Domain anomaly detection rule algorithms (e.g. TOOL_RETRY_STREAK, CONTEXT_SATURATION_SPIKE, SUBAGENT_REWORK_LOOP), which belong to AL-7.
2. REST/SSE API endpoints specifically for anomaly queries or trace streaming, which belong to AL-8.
3. Visual timeline and waterfall UI components in the web dashboard, which belong to AL-9.
4. Any Python scripts or runtime dependencies (Python has been removed from the repository).

---

## 4. Event Telemetry Data Architecture & Extraction Lifecycle

### Event Fact Structure
Each EventFact represents a discrete, ordered lifecycle occurrence:
- timestamp: ISO 8601 formatted timestamp string in UTC (microsecond precision).
- session_id: Unique identifier of the session where the event occurred.
- turn: Sequential 1-based turn number within the session.
- role: Agent role associated with the session (e.g. orchestrator, planner, coder, tester, reviewer).
- event_type: Categorical identifier (tool_call, turn_message, compaction).
- tool_name: Optional string containing the tool name (e.g. bash, read, edit, grep, write), present for tool calls.
- status: Execution status string (e.g. completed, error).
- error_message: Optional string containing the error excerpt if the event encountered an error.
- duration_ms: Optional int64 measuring execution duration in milliseconds.
- context_tokens: Cumulative context token volume (input tokens plus cache read tokens) at the point of the event.

### Chronological Correlation Algorithm
1. Query session records for target root threads and child invocations, establishing session role and model mapping.
2. Query message records for all sessions ordered by creation timestamp:
   - Assign sequential 1-based turn numbers to messages within each session.
   - Extract input and cache tokens to compute context_tokens.
   - Generate turn_message event facts.
3. Query part records for all sessions ordered by creation timestamp:
   - For parts where data LIKE '%"type":"tool"%':
     - Parse tool name, state status, and output.
     - Detect tool errors (matching status error, presence of Traceback, or output beginning with Error:).
     - Extract truncated error message excerpt when status is error.
     - Calculate duration_ms from time_updated - time_created.
     - Correlate turn index and context tokens from parent message ID.
     - Generate tool_call event facts.
   - For parts where data LIKE '%"type":"compaction"%':
     - Correlate turn index from parent message ID.
     - Generate compaction event facts.
4. Merge all extracted events into a single slice.
5. Sort chronologically by timestamp in ascending order, using turn as secondary tie-breaker.

### Existing Workspace Backfill Migration
1. Discover all workspace directories in workspaces/ containing agentic_metrics.json.
2. For each workspace:
   - Load existing ledger facts.
   - When --recompute-db is passed, if the ledger lacks events or has incomplete telemetry, query the local OpenCode SQLite database using session resolution strategies (cached session ID, ticket ID matching, workspace path correlation).
   - If session records exist: extract full chronological events, update invocations and tools, and write updated agentic_metrics.json with populated events.
   - If session records do not exist locally: retain existing invocations, tools, and duration facts verbatim, marking pending steps as appropriate.

---

## 5. Acceptance Criteria

1. Legacy agentic_metrics.json ledgers omitting events deserialize successfully without error or panic.
2. All existing domain aggregations (ComputeTotals, ComputeWorkflowEfficiency, ComputeRoleMetrics, ComputeModelMetrics, DeriveMetricsLedger) continue to pass without regression.
3. Ledgers containing events deserialize with populated, chronologically sorted Events slices.
4. Native Go extraction via agentic-lens generate populates ordered events in agentic_metrics.json when connected to an OpenCode SQLite database containing session records.
5. Native Go backfill via agentic-lens backfill --recompute-db populates events for tickets whose sessions exist in the SQLite database, while preserving existing telemetry for tickets whose sessions do not exist locally.
6. Existing workspaces in workspaces/ on disk are backfilled so that agentic_metrics.json contains populated chronological events for all tickets with sessions in the local OpenCode database.
7. Extraction correctly identifies tool execution status (completed vs error), extracts truncated error messages for failures, calculates duration in milliseconds, and associates accurate turn numbers and context tokens.
8. TypeScript interfaces in web/src/domain/types.ts mirror Go models with strict typing and zero any.
9. Quality checks pass cleanly:
   - go vet ./... succeeds with zero errors.
   - go test -race ./... passes across all packages with zero race conditions or failures.
   - npm run typecheck succeeds in web/ with zero errors.
10. Strict zero-emoji policy enforced across all code, tests, documentation, and serialized JSON ledgers.
