# Contract: AL-33 End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall

## 1. Problem Statement & Scope Boundaries

Committed Git telemetry ledgers (workspaces/<TICKET_ID>/agentic_metrics.json) are intentionally sanitized and pruned per the Dual-Plane Telemetry RFC: routine turn messages and successful tool calls are stripped, and database error strings are reduced to bare failure indicators to prevent leaking secrets, credentials, internal file paths, or proprietary code into Git history. Consequently, committed ledgers provide safe, portable high-level metrics and bare error timelines, but lack deep diagnostic context.

Full execution context—including user prompts, LLM reasoning thought blocks, tool input arguments, and complete tool outputs—resides exclusively in the developer's local OpenCode SQLite database. The AgenticLens server (agentic-lens serve) and web dashboard must deliver a unified dual-plane experience:

1. When running on a developer workstation where the local OpenCode SQLite database contains the ticket session data, the system dynamically hydrates rich event traces in memory without writing transient data to disk.
2. When running in offline or remote environments (CI runners, Docker containers, remote clones without database access), the system gracefully degrades to reading the committed bare ledger, rendering an informational offline banner and a compact error timeline.
3. The TypeScript web dashboard (TraceWaterfall.tsx and TicketDeepDive.tsx) must adapt seamlessly based on the dual-plane response, displaying rich interactive inspection drawers when hydrated and a clean fallback notice when offline.

### In Scope
1. Domain Core Models (backend/internal/domain/models.go, backend/internal/domain/events.go, backend/internal/domain/enums.go):
   - Define WorkspaceEventsResponse struct holding TicketID string, IsHydrated bool, Events slice of EventFact, and DataSource string.
   - Extend EventFact with optional in-memory diagnostic pointers: PromptContent pointer to string, ReasoningContent pointer to string, ToolInput pointer to string, and ToolOutput pointer to string.
   - Define canonical data source constants: DataSourceLocalSQLite ("local_sqlite") and DataSourceCommittedLedger ("committed_ledger").
   - Update PruneEventsForPersistence to explicitly set PromptContent, ReasoningContent, ToolInput, and ToolOutput to nil on all retained events, reinforcing the FCIS invariant that sensitive context is never serialized to disk ledgers.
2. Boundary Adapters & SQLite Query Engine (backend/internal/adapters/live/sqlite.go, backend/internal/adapters/live/service.go, backend/internal/adapters/opencode/resolver.go):
   - Implement read-only, non-locking SQLite queries to extract full chronological messages, parts, reasoning thoughts, and tool inputs/outputs across all session IDs associated with a ticket.
   - Enforce non-blocking database pragmas: PRAGMA query_only = ON;, PRAGMA busy_timeout = 5000;, PRAGMA read_uncommitted = ON;, and WAL mode.
   - Resolve ticket session IDs from invocation records in agentic_metrics.json or fallback resolution via workspace path matching.
   - Strictly enforce zero disk persistence: transient hydrated data remains 100% in-memory.
3. HTTP Service Layer (backend/internal/adapters/http/dto.go, backend/internal/adapters/http/handlers.go, backend/internal/adapters/http/server.go):
   - Extend ResponsePayload constraint interface with domain.WorkspaceEventsResponse.
   - Update WriteEventsResponse to serialize domain.WorkspaceEventsResponse.
   - Update WorkspaceEventsHandler to consume a WorkspaceEventHydrator boundary interface, attempting local SQLite hydration when available and degrading gracefully to committed ledger events.
4. Frontend Types & Presentation Components (web/src/types/telemetry.ts, web/src/domain/types.ts, web/src/api/client.ts, web/src/components/TraceWaterfall.tsx, web/src/views/TicketDeepDive.tsx):
   - Define TypeScript interfaces WorkspaceEventsResponse and DataSource.
   - Update api/client.ts to fetch and parse WorkspaceEventsResponse, supporting both hydrated and fallback payloads.
   - Update TraceWaterfall.tsx to detect is_hydrated:
     - When is_hydrated is true: render full waterfall with expandable drawers for message prompts, LLM reasoning thought blocks, and formatted tool parameters/outputs.
     - When is_hydrated is false: render compact bare error waterfall with an informational offline banner: "Committed Git Ledger — Local Database Offline".
   - Detail drawers and modal inspection views adhering strictly to the zero-emoji design system.
5. Quality Gates & Automated Tests:
   - Go unit tests verifying WorkspaceEventsResponse serialization, data source values, and pruning security.
   - Go integration tests verifying non-locking SQLite hydration against a temporary SQLite database, as well as offline fallback degradation.
   - TypeScript Vitest component tests verifying hydrated and unhydrated rendering in TraceWaterfall.test.tsx and TicketDeepDive.test.tsx.
   - Typecheck and race-detector verification (go test -race ./..., npm run typecheck, npm test).

### Out of Scope
1. Direct mutation or write operations to the local OpenCode SQLite database.
2. Persisting hydrated trace events to workspaces/ or any local disk file.
3. Modifying the CLI agentic-lens generate command or re-introducing routine events into committed ledgers.
4. Implementing live streaming SSE for committed historical tickets (governed separately by AL-17/AL-20 for active live sessions).

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

Dependencies flow strictly inward across layers:

### Go Functional Core (backend/internal/domain/)
- 100% pure functional core. Operates exclusively on immutable structs, slices, and primitive types.
- Zero external imports: no database/sql, os, io, net/http, or outer adapter references.
- Pure event sorting: SortEvents sorts stably by timestamp ascending, then turn ascending, returning a new allocated slice.
- Pure pruning invariant: PruneEventsForPersistence guarantees that ErrorMessage, PromptContent, ReasoningContent, ToolInput, and ToolOutput are explicitly set to nil for any event retained in memory before persistence.

### Go Imperative Shell (backend/internal/adapters/, backend/cmd/)
- Handles SQLite connection pooling, non-blocking pragma execution, filesystem reading of workspaces/<TICKET_ID>/agentic_metrics.json, and HTTP transport serialization.
- Non-blocking SQLite reader: configured with PRAGMA query_only = ON;, PRAGMA busy_timeout = 5000;, and PRAGMA read_uncommitted = ON; in WAL mode.
- Memory isolation: all hydrated traces are constructed on-the-fly and streamed directly over HTTP. Zero transient trace data is written to disk or workspace directories.
- Graceful degradation: any SQLite connectivity failure, locked database, missing table, or absent session ID must be caught and logged as non-fatal, immediately falling back to the committed ledger.

### TypeScript Strictness & Presentation (web/src/)
- Strict TypeScript with strict: true, noImplicitAny: true, and exactOptionalPropertyTypes: true.
- Zero any across all domain types, API clients, and UI components.
- Zero emojis across all UI strings, badges, drawer headers, buttons, tooltips, and console outputs.
- Defensive presentation: handles empty event arrays, missing tool inputs, or long text snippets gracefully with proper text truncation and overflow handling.

---

## 3. Layer Contracts and Data Models

### Domain Core Layer (backend/internal/domain/models.go)

Struct WorkspaceEventsResponse:
- Field TicketID (string, JSON key ticket_id): unique ticket identifier.
- Field IsHydrated (bool, JSON key is_hydrated): true if events were dynamically hydrated from the local SQLite database; false if loaded from committed offline ledger.
- Field Events (slice of EventFact, JSON key events): chronological list of execution events.
- Field DataSource (string, JSON key data_source): data source indicator ("local_sqlite" or "committed_ledger").

Struct EventFact (extended with optional fields):
- Field Timestamp (string, JSON key timestamp): ISO 8601 UTC timestamp.
- Field SessionID (string, JSON key session_id): session identifier.
- Field Turn (int64, JSON key turn): sequential turn number.
- Field Role (string, JSON key role): agent persona role.
- Field EventType (EventType, JSON key event_type): tool_call, turn_message, or compaction.
- Field ToolName (pointer to string, JSON key tool_name, omitempty): tool name.
- Field Status (string, JSON key status): completed, error, or failed.
- Field ErrorMessage (pointer to string, JSON key error_message, omitempty): error details.
- Field DurationMs (pointer to int64, JSON key duration_ms, omitempty): execution duration in milliseconds.
- Field ContextTokens (int64, JSON key context_tokens): context window token count.
- Field PromptContent (pointer to string, JSON key prompt_content, omitempty): user prompt or message turn text.
- Field ReasoningContent (pointer to string, JSON key reasoning_content, omitempty): chain-of-thought reasoning text.
- Field ToolInput (pointer to string, JSON key tool_input, omitempty): formatted tool input parameters.
- Field ToolOutput (pointer to string, JSON key tool_output, omitempty): formatted tool execution output.

Constants (backend/internal/domain/enums.go):
- DataSourceLocalSQLite: string constant "local_sqlite".
- DataSourceCommittedLedger: string constant "committed_ledger".

### Boundary Adapter Layer (backend/internal/adapters/live/)

Interface WorkspaceEventHydrator (backend/internal/adapters/http/handlers.go):
- Method HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error).

Implementation in LiveService (backend/internal/adapters/live/service.go):
- Resolves session IDs from ledger.Facts.Invocations and ledger.Facts.Events. If missing, falls back to opencode.ResolveActiveRootSession or opencode.ResolveSessionByWorkspacePath.
- Checks session existence in SQLite via SQLiteHook.SessionExists.
- Queries messages and parts across all related session IDs ordered chronologically.
- Parses user prompts, assistant messages, reasoning thoughts, and tool inputs/outputs.
- Assembles rich slice of domain.EventFact and returns WorkspaceEventsResponse with is_hydrated: true and data_source: "local_sqlite".
- If sessions are not found or an error occurs, returns WorkspaceEventsResponse with is_hydrated: false and data_source: "committed_ledger".

### HTTP Service Layer (backend/internal/adapters/http/)

Endpoint: GET /api/workspaces/{ticketId}/events
- Input: ticketId path parameter. Validates against directory traversal characters (.., /, \).
- Status 400 Bad Request: if ticketId is invalid or contains traversal characters.
- Status 404 Not Found: if workspaces/<ticketId>/agentic_metrics.json does not exist on disk.
- Status 200 OK: returns JSON object conforming to domain.WorkspaceEventsResponse.
  - When hydrated: is_hydrated: true, data_source: "local_sqlite", full event list with prompts, reasoning, and tool details.
  - When unhydrated: is_hydrated: false, data_source: "committed_ledger", compact bare events.

### Frontend TypeScript Types (web/src/types/telemetry.ts, web/src/domain/types.ts)

Type DataSource:
- String union: 'local_sqlite' | 'committed_ledger'.

Interface WorkspaceEventsResponse:
- Field ticket_id (string): ticket identifier.
- Field is_hydrated (boolean): hydration status.
- Field events (readonly array of EventFact): execution events.
- Field data_source (DataSource): data source indicator.

Interface EventFact (extended):
- Optional fields: prompt_content?: string | null | undefined, reasoning_content?: string | null | undefined, tool_input?: string | null | undefined, tool_output?: string | null | undefined.

### Presentation Components (web/src/components/TraceWaterfall.tsx)

Props TraceWaterfallProps:
- events: readonly EventFact[]
- anomalies?: readonly Anomaly[] | undefined
- selectedEvent?: EventFact | null | undefined
- onSelectEvent?: ((event: EventFact, anomaly?: Anomaly | undefined) => void) | undefined
- isHydrated?: boolean | undefined
- dataSource?: string | undefined

UI Behaviors:
1. Fallback Offline Banner (when isHydrated is false):
   - Title: "Committed Git Ledger — Local Database Offline"
   - Body: "Full prompts, tool inputs, and error contexts are available only when running locally against the session database."
   - Visual: styled informational callout with distinct border and neutral theme palette, strictly zero emojis.
2. Hydrated Waterfall (when isHydrated is true):
   - Banner is suppressed.
   - Message turns with prompt_content: expandable prompt preview drawer.
   - Reasoning turns with reasoning_content: expandable reasoning thought drawer.
   - Tool call turns with tool_input and tool_output: interactive inspection drawer/modal displaying formatted inputs, outputs, and execution metadata.
3. Event Selection and Drawer Inspector:
   - Clicking an event in either mode opens the event inspector drawer.
   - When hydrated, displays full prompt, reasoning, and tool parameter/output tabs.
   - When offline, displays bare metadata, status, context tokens, and anomaly indicators.

---

## 4. Error Handling & Degradation Semantics

1. Path Traversal & Hostile Input:
   - Any ticketId containing .., /, \, %2e, %2f, or %5c must immediately return HTTP 400 Bad Request with structured ErrorResponse.
2. Workspace File Absent:
   - If workspaces/<ticketId>/agentic_metrics.json is missing or unreadable, return HTTP 404 Not Found with structured ErrorResponse.
3. Database Inaccessibility:
   - If SQLite file does not exist, has invalid permissions, is locked, or encounters busy timeout, catch the error, log a diagnostic notice, and return HTTP 200 with is_hydrated: false and data_source: "committed_ledger".
4. Empty Events Array:
   - If a workspace has an empty or null events array, return HTTP 200 with events: []. Never return null or HTTP 500.
5. Zero Disk Writes:
   - Under no circumstances shall transient hydrated data be persisted to workspaces/ or any local file. All hydration occurs purely in memory.

---

## 5. Quality Gates & Test Requirements

1. Backend Unit Tests (backend/internal/domain/events_test.go):
   - Verify WorkspaceEventsResponse JSON marshaling with is_hydrated: true and is_hydrated: false.
   - Verify PruneEventsForPersistence strips PromptContent, ReasoningContent, ToolInput, and ToolOutput to nil.
2. Backend Integration Tests (backend/internal/adapters/live/sqlite_test.go, backend/internal/adapters/http/server_test.go):
   - Test SQLite event hydration against a temporary SQLite database populated with messages, parts, reasoning, and tool calls.
   - Verify non-locking read performance under active concurrent write transactions.
   - Test graceful fallback when SQLite database is absent or ticket session does not exist.
   - Test HTTP endpoint GET /api/workspaces/{ticketId}/events returning both hydrated and unhydrated dual-plane responses.
   - Verify zero files created or modified in workspaces/ during hydration.
3. Frontend Vitest Tests (web/src/components/TraceWaterfall.test.tsx, web/src/api/client.test.ts, web/src/views/TicketDeepDive.test.tsx):
   - Test TraceWaterfall renders the offline fallback banner when isHydrated is false.
   - Test TraceWaterfall renders prompt, reasoning, and tool drawers when isHydrated is true.
   - Test fetchWorkspaceEvents parses WorkspaceEventsResponse and handles legacy/mock fallbacks cleanly.
   - Test TicketDeepDive correctly passes hydration props to TraceWaterfall.
4. Full Automated Verification:
   - go test -race ./... passes cleanly with zero race conditions.
   - npm run typecheck passes cleanly with zero any.
   - npm test passes all Vitest test suites.
   - Single binary build verification (make build).
