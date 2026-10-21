# Implementation Plan: AL-33 End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall

## 1. Goal and Iteration Scope

The goal of Iteration 01 is to deliver an end-to-end vertical slice for dynamic local database event hydration and the dual-plane waterfall view across Go backend adapters/HTTP and TypeScript frontend components.

When running locally with the OpenCode SQLite database accessible, the backend queries messages, reasoning thoughts, and tool inputs/outputs to dynamically hydrate rich in-memory event traces (is_hydrated: true). When offline (CI, Docker, or remote clones without database access), the backend gracefully degrades to committed bare events (is_hydrated: false). The web dashboard updates TraceWaterfall.tsx and TicketDeepDive.tsx to render rich interactive inspection drawers when hydrated and a clean offline indicator banner when unhydrated, strictly enforcing zero disk persistence and zero emojis.

---

## 2. Grounded File Map & Symbol References

### Pure Domain Functional Core
- backend/internal/domain/models.go:
  - Struct WorkspaceEventsResponse: fields TicketID string, IsHydrated bool, Events slice of EventFact, DataSource string.
  - Struct EventFact: extend with optional pointers PromptContent pointer to string, ReasoningContent pointer to string, ToolInput pointer to string, ToolOutput pointer to string.
- backend/internal/domain/enums.go:
  - Constants DataSourceLocalSQLite = "local_sqlite" and DataSourceCommittedLedger = "committed_ledger".
- backend/internal/domain/events.go:
  - Function PruneEventsForPersistence: ensure PromptContent, ReasoningContent, ToolInput, and ToolOutput are set to nil on retained events.
- backend/internal/domain/events_test.go:
  - Unit tests for WorkspaceEventsResponse serialization and pruning sanitization invariants.

### Imperative Shell Boundary Adapters & HTTP
- backend/internal/adapters/live/sqlite.go:
  - Method FetchSessionTraceMulti(sessionIDs slice of string): queries chronological messages and parts across all session IDs in the invocation tree.
  - Non-blocking pragmas: PRAGMA query_only = ON;, PRAGMA busy_timeout = 5000;, PRAGMA read_uncommitted = ON;.
- backend/internal/adapters/live/service.go:
  - Interface WorkspaceEventHydrator implementation: method HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error).
  - Reconstructs rich events with prompts, reasoning, and tool inputs/outputs from SQLite messages and parts.
- backend/internal/adapters/http/dto.go:
  - Extend ResponsePayload constraint interface with domain.WorkspaceEventsResponse.
  - Update WriteEventsResponse helper to write domain.WorkspaceEventsResponse.
- backend/internal/adapters/http/handlers.go:
  - Define WorkspaceEventHydrator interface.
  - Update WorkspaceEventsHandler(workspacesDir string, hydrator ...WorkspaceEventHydrator) to orchestrate local hydration with graceful committed ledger fallback.
- backend/internal/adapters/http/server.go:
  - Wire WorkspaceEventHydrator from LiveService into WorkspaceEventsHandler.
- backend/internal/adapters/http/server_test.go:
  - Integration tests for GET /api/workspaces/{ticketId}/events verifying dual-plane response structures, fallback degradation, and traversal protection.

### Presentation & Web Dashboard
- web/src/types/telemetry.ts and web/src/domain/types.ts:
  - Interface WorkspaceEventsResponse: fields ticket_id string, is_hydrated boolean, events readonly array of EventFact, data_source DataSource.
  - Type DataSource: 'local_sqlite' | 'committed_ledger'.
  - Extend EventFact: optional fields prompt_content optional string or null, reasoning_content optional string or null, tool_input optional string or null, tool_output optional string or null.
- web/src/api/client.ts:
  - Update fetchWorkspaceEvents(ticketId string): Promise of ApiResponse of WorkspaceEventsResponse:
    - Parses dual-plane response object.
    - Normalizes legacy/mock array payloads into fallback WorkspaceEventsResponse.
- web/src/components/TraceWaterfall.tsx:
  - Extend TraceWaterfallProps with isHydrated optional boolean, dataSource optional string.
  - Render offline fallback banner ("Committed Git Ledger — Local Database Offline") when isHydrated is false.
  - Render expandable prompt, reasoning, and tool inspection drawers when isHydrated is true.
  - Detail drawer/modal for inspecting tool inputs/outputs in hydrated mode.
  - Zero emojis across all UI elements.
- web/src/views/TicketDeepDive.tsx:
  - Update fetchWorkspaceEvents consumption to store hydration metadata and pass isHydrated and dataSource to TraceWaterfall.
- web/src/components/TraceWaterfall.test.tsx, web/src/api/client.test.ts, web/src/views/TicketDeepDive.test.tsx:
  - Vitest test suites verifying dual-plane rendering, fallback banner, drawer interactions, and API normalization.

---

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Pure Domain Functional Core
- Task 1.1: Update backend/internal/domain/models.go:
  - Add struct WorkspaceEventsResponse with fields TicketID string (json ticket_id), IsHydrated bool (json is_hydrated), Events slice of EventFact (json events), and DataSource string (json data_source).
  - Extend struct EventFact with optional pointer fields PromptContent pointer to string (json prompt_content, omitempty), ReasoningContent pointer to string (json reasoning_content, omitempty), ToolInput pointer to string (json tool_input, omitempty), and ToolOutput pointer to string (json tool_output, omitempty).
- Task 1.2: Update backend/internal/domain/enums.go:
  - Define string constants DataSourceLocalSQLite = "local_sqlite" and DataSourceCommittedLedger = "committed_ledger".
- Task 1.3: Update backend/internal/domain/events.go:
  - In PruneEventsForPersistence, ensure PromptContent, ReasoningContent, ToolInput, and ToolOutput are explicitly set to nil on every retained event, guaranteeing that no sensitive data can be committed to disk ledgers.
- Task 1.4: Update backend/internal/domain/events_test.go:
  - Add table-driven unit tests verifying serialization of WorkspaceEventsResponse with IsHydrated: true and IsHydrated: false.
  - Assert that PruneEventsForPersistence strips all prompt, reasoning, and tool input/output fields to nil.

### Phase 2: Boundary SQLite Reader & Live Service Hydration
- Task 2.1: Enhance backend/internal/adapters/live/sqlite.go:
  - Add method FetchSessionTraceMulti(sessionIDs slice of string) to SQLiteHook.
  - Query message and part records for all provided session IDs in a single query or sorted union ordered by time_created ascending.
  - Add method ResolveChildSessionIDs(parentSessionIDs slice of string) to find child sub-sessions where parent_id matches.
- Task 2.2: Implement HydrateWorkspaceEvents in backend/internal/adapters/live/service.go:
  - Collect session IDs from ledger.Facts.Invocations and ledger.Facts.Events. If empty, invoke opencode.ResolveActiveRootSession or opencode.ResolveSessionByWorkspacePath.
  - Query child sessions to gather the complete thread execution tree.
  - Verify session existence via SQLiteHook.SessionExists. If absent, return fallback WorkspaceEventsResponse with IsHydrated: false and DataSource: DataSourceCommittedLedger.
  - Extract messages and parts via FetchSessionTraceMulti.
  - Reconstruct rich slice of domain.EventFact:
    - For assistant messages: extract PromptContent, ReasoningContent, and token metrics.
    - For tool calls: extract tool name, ToolInput, ToolOutput, execution duration, and error status via domain.ClassifyToolFailure.
    - For compactions: generate compaction event facts.
  - Return WorkspaceEventsResponse with IsHydrated: true, DataSource: DataSourceLocalSQLite, and sorted events.
  - Guarantee zero disk writes during hydration.

### Phase 3: HTTP Handlers, DTO Integration & Server Wiring
- Task 3.1: Update backend/internal/adapters/http/dto.go:
  - Add domain.WorkspaceEventsResponse to the ResponsePayload generic interface union.
  - Update WriteEventsResponse(w http.ResponseWriter, resp domain.WorkspaceEventsResponse) to serialize the dual-plane response object.
- Task 3.2: Update backend/internal/adapters/http/handlers.go:
  - Define interface WorkspaceEventHydrator with method HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error).
  - Update WorkspaceEventsHandler(workspacesDir string, hydrator ...WorkspaceEventHydrator):
    - Validate ticketId against path traversal.
    - Load workspace ledger via adapters.LoadTicketWorkspace(ticketID, workspacesDir). Return 404 if not found.
    - Prepare sorted fallback events via domain.SortEvents(targetLedger.Facts.Events).
    - If hydrator is provided, call HydrateWorkspaceEvents.
    - If hydrator is nil, or if hydration returns unhydrated/error, write WorkspaceEventsResponse with is_hydrated: false, data_source: "committed_ledger", and fallback events.
- Task 3.3: Update backend/internal/adapters/http/server.go:
  - In NewServer, check if liveService implements WorkspaceEventHydrator and pass it to WorkspaceEventsHandler.
- Task 3.4: Update backend/internal/adapters/http/server_test.go:
  - Update existing /events tests to decode domain.WorkspaceEventsResponse instead of raw array.
  - Add integration tests verifying HTTP response when SQLite database is present (is_hydrated: true, data_source: "local_sqlite") and when missing (is_hydrated: false, data_source: "committed_ledger").

### Phase 4: Frontend Types, API Client & Presentation Waterfall
- Task 4.1: Update TypeScript types in web/src/types/telemetry.ts and web/src/domain/types.ts:
  - Define type DataSource = 'local_sqlite' | 'committed_ledger'.
  - Define interface WorkspaceEventsResponse with ticket_id string, is_hydrated boolean, events readonly array of EventFact, data_source DataSource.
  - Extend EventFact with optional prompt_content string or null, reasoning_content string or null, tool_input string or null, tool_output string or null.
- Task 4.2: Update web/src/api/client.ts:
  - Update fetchWorkspaceEvents(ticketId string): Promise of ApiResponse of WorkspaceEventsResponse.
  - Normalize incoming responses: if response is an object with is_hydrated, return directly; if legacy array, wrap into WorkspaceEventsResponse with is_hydrated: false and data_source: 'committed_ledger'.
  - In mock fallbacks, return structured WorkspaceEventsResponse.
- Task 4.3: Enhance web/src/components/TraceWaterfall.tsx:
  - Accept isHydrated optional boolean and dataSource optional string in TraceWaterfallProps.
  - When isHydrated is false:
    - Display informational offline banner:
      - Title: "Committed Git Ledger — Local Database Offline"
      - Description: "Full prompts, tool inputs, and error contexts are available only when running locally against the session database."
      - Styled without emojis using neutral badge palette.
    - Display compact bare error and compaction timeline.
  - When isHydrated is true:
    - Suppress offline banner.
    - For message turns with prompt_content: provide expandable prompt preview drawer.
    - For turns with reasoning_content: render expandable chain-of-thought block drawer.
    - For tool calls with tool_input or tool_output: render expandable inspection drawer/modal displaying formatted tool input parameters and output results.
  - Detail drawer/modal for inspecting tool executions in hydrated mode.
  - Zero emojis across all labels, buttons, drawers, and banners.
- Task 4.4: Update web/src/views/TicketDeepDive.tsx:
  - Update event loading to receive WorkspaceEventsResponse.
  - Store hydration state and pass isHydrated and dataSource to TraceWaterfall.

### Phase 5: Verification & Quality Gates
- Task 5.1: Run Go tests with race detection:
  - go test -race ./backend/...
- Task 5.2: Run TypeScript type checking and Vitest tests:
  - npm run typecheck (zero any violations).
  - npm test (all unit and component tests passing).
- Task 5.3: Verify single binary build:
  - make build or npm run build && go build -o bin/agentic-lens ./backend/cmd/agentic-lens.
- Task 5.4: Verify zero emojis in code, comments, UI strings, and docstrings.

---

## 4. Verification Checkpoints for Tester Subagent

1. Dual-Plane API Verification:
   - Request GET /api/workspaces/{ticketId}/events with valid ticket and active local SQLite database: assert status 200, is_hydrated: true, data_source: "local_sqlite", and populated prompt_content / tool_input.
   - Request GET /api/workspaces/{ticketId}/events with valid ticket and absent/locked SQLite database: assert status 200, is_hydrated: false, data_source: "committed_ledger", and bare committed events.
   - Request GET /api/workspaces/NON-EXISTENT/events: assert status 404 Not Found.
   - Request with hostile path traversal (..%2F..%2Fetc%2Fpasswd): assert status 400 Bad Request.
2. Zero Disk Persistence Verification:
   - Capture directory snapshot of workspaces/ before and after invoking hydration.
   - Assert zero new files created and zero existing files modified in workspaces/.
3. UI Component Verification:
   - In TraceWaterfall.test.tsx, render with isHydrated: false: assert presence of banner "Committed Git Ledger — Local Database Offline", absence of rich tool drawers.
   - In TraceWaterfall.test.tsx, render with isHydrated: true: assert absence of offline banner, presence of expandable prompt, reasoning, and tool drawers.
   - Assert zero emojis rendered in the component tree.
4. Static Typing & Race Conditions:
   - Assert npm run typecheck completes with zero errors.
   - Assert go test -race ./... completes with zero race conditions.
