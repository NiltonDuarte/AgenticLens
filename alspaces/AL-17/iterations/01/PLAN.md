# Implementation Plan: AL-17 Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry

## 1. Goal and Iteration Scope

The goal of Iteration 01 is to implement ticket AL-17: delivering a non-locking, read-only SQLite database hook for OpenCode, an active session discovery and tree reconstruction engine, a streamlined title-only ticket inference heuristic, an in-memory transient trace extraction buffer, and HTTP/SSE endpoints (/api/live/sessions, /api/live/sessions/:id/trace, /api/live/stream).

Per steering feedback, ticket inference strictly inspects session titles and eliminates dependencies on Git branches or workspace directory scanning. This slice spans Go domain modeling, boundary SQLite querying with non-blocking pragmas, HTTP route integration, and adversarial table tests with race detection.

---

## 2. Grounded Reference Architecture & File Map

- Pure Domain Functional Core:
  - backend/internal/domain/live.go: Pure domain models (LiveSessionNode, LiveSessionTree, LivePartFact, LiveMessageFact, LiveTraceView), pure tool failure classifier (ClassifyToolFailure), pure title-based ticket inference heuristic (InferTicketID), and pure tree hierarchy builder (BuildSessionTree).
  - backend/internal/domain/live_test.go: Pure unit tests for title ticket inference, tool failure severity classification, and tree assembly without I/O, Git, or mocks.
- Imperative Shell Boundary Adapters:
  - backend/internal/adapters/live/sqlite.go: Read-only SQLite hook opening connection with non-blocking pragmas (query_only, busy_timeout, read_uncommitted), active session discovery, and message/part trace querying.
  - backend/internal/adapters/live/cache.go: Thread-safe in-memory ring buffer holding recent trace messages bounded by live.max_trace_buffer_messages.
  - backend/internal/adapters/live/service.go: Live observability coordinator polling SQLite on live.poll_interval_ms, coordinating tree updates via session titles, and broadcasting events to SSE subscribers.
  - backend/internal/adapters/live/sqlite_test.go: Boundary tests for read-only pragma verification, non-locking reads during concurrent write load, transient buffer bounds, and zero disk persistence assertions.
- HTTP Service & DTOs:
  - backend/internal/adapters/http/dto.go: Extend ResponsePayload generic interface to include domain.LiveSessionTree and domain.LiveTraceView; add typed serialization helpers.
  - backend/internal/adapters/http/live_handlers.go: HTTP handlers for session tree, trace view, and SSE streaming.
  - backend/internal/adapters/http/server.go: Register /api/live/sessions, /api/live/sessions/{id}/trace, and /api/live/stream routes.
  - backend/internal/adapters/http/live_handlers_test.go: Integration tests for HTTP live endpoints and SSE stream disconnection handling.
- Service CLI Wiring:
  - backend/cmd/agentic-lens/main.go: In runServer, initialize live database hook and service from project config, passing service to HTTP server with graceful shutdown.

---

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Domain Functional Core (backend/internal/domain/live.go)
- Task 1.1: Define domain structs in backend/internal/domain/live.go:
  - LiveSessionNode with fields ID string, ParentID pointer to string, Agent string, Role AgentRole, Model string, Title string, TimeCreated string, TimeUpdated string, Status string, GuessedTicketID pointer to string, TicketConfidence float64, and SubSessions slice of pointers to LiveSessionNode.
  - LiveSessionTree with fields RootSessions slice of pointers to LiveSessionNode, LastRefreshedAt string, and TotalActiveSessions int.
  - LivePartFact with fields ID string, Type string, ToolName pointer to string, InputSnippet string, OutputSnippet string, Status string, IsError bool, ErrorMessage pointer to string, and Severity string.
  - LiveMessageFact with fields ID string, Role string, Content string, Timestamp string, ReasoningContent pointer to string, and Parts slice of LivePartFact.
  - LiveTraceView with fields SessionID string, Messages slice of LiveMessageFact, TotalToolCalls int, FailedToolCalls int, and RecentErrors slice of LivePartFact.
- Task 1.2: Implement pure function ClassifyToolFailure:
  - Accepts toolName string, status string, output string, and consecutiveErrors int.
  - Returns isError bool, severity string, and errorMessage pointer to string.
  - Inspects status and output text for exception tracebacks (Traceback (most recent call last), panic:, fatal error:, UnhandledPromiseRejection), syntax/compiler errors, and critical system faults (out of memory, killed, permission denied), assigning severity high.
  - Escalates severity to high if consecutiveErrors is greater than or equal to 2.
  - Categorizes non-zero exit codes, missing files, and timeouts as severity medium.
  - Treats empty search results and soft cancellations as severity low.
  - Returns sanitized error message excerpt capped at 2048 characters.
- Task 1.3: Implement pure function InferTicketID:
  - Accepts sessionTitle string and cfg TicketConfig.
  - Returns guessedID pointer to string and confidence float64.
  - Trims whitespace from sessionTitle. If empty, returns nil and 0.0.
  - Invokes NormalizeTicketID(sessionTitle, cfg) to parse ticket identifiers against configured prefixes case-insensitively.
  - On match, returns pointer to canonical uppercase string (e.g. AL-17) and confidence 1.0.
  - If no match found, returns nil and confidence 0.0.
- Task 1.4: Implement pure function BuildSessionTree:
  - Accepts slice of flat session records and returns assembled slice of root LiveSessionNode pointers.
  - Groups nodes with ParentID nil (or whose parent is not in the set) as roots.
  - Recursively attaches child sessions to their respective parent nodes.
  - Sorts root sessions by TimeUpdated descending, and child sessions by TimeCreated ascending.

### Phase 2: Non-Locking SQLite Adapter & Transient Buffer (backend/internal/adapters/live/)
- Task 2.1: Implement SQLiteHook in backend/internal/adapters/live/sqlite.go:
  - Define SQLiteHook struct wrapping sql.DB pointer and database path string.
  - Constructor OpenReadOnlyHook(dbPath string, busyTimeoutMs int) returning pointer to SQLiteHook and error, opening connection using DSN file:<path>?mode=ro&_busy_timeout=<timeout>.
  - Executes initialization pragmas on connection: PRAGMA query_only = ON;, PRAGMA busy_timeout = 5000;, and PRAGMA read_uncommitted = ON;.
  - Sets connection pool limits: SetMaxOpenConns(5), SetMaxIdleConns(2).
  - Implements method FetchActiveSessions(activeThresholdMs int64, nowMs int64) returning slice of raw session records (including title) and error, querying session table for sessions updated within threshold or with active status.
  - Implements method FetchSessionTrace(sessionID string) returning slice of raw message records, slice of raw part records, and error, querying message and part tables for given session ID ordered by time_created ascending.
- Task 2.2: Implement TraceBuffer in backend/internal/adapters/live/cache.go:
  - Thread-safe in-memory cache wrapping sync.RWMutex and a map of session ID to slice of domain.LiveMessageFact.
  - Enforces max_trace_buffer_messages cap per session, discarding oldest messages when capacity is exceeded.
  - Guarantees zero disk persistence: strictly in-memory data structures with no filesystem operations.
- Task 2.3: Implement LiveService in backend/internal/adapters/live/service.go:
  - Coordinates polling, tree reconstruction, and real-time broadcasting.
  - Holds references to SQLiteHook, TraceBuffer, and domain.ProjectConfig.
  - Method GetSessionTree(ctx context.Context) returning domain.LiveSessionTree and error: queries active sessions, runs InferTicketID directly on each session's Title, and calls BuildSessionTree. Zero Git calls, zero workspace scans.
  - Method GetSessionTrace(ctx context.Context, sessionID string) returning domain.LiveTraceView and error: queries messages and parts, classifies tool failures with ClassifyToolFailure, populates buffer, and builds LiveTraceView.
  - Polling loop running on live.poll_interval_ms: detects updates and pushes events to registered SSE subscriber channels.
  - Method Subscribe() returning channel of LiveEvent and cancellation cleanup function: registers an SSE subscriber channel.

### Phase 3: HTTP API Handlers, DTO Integration & Server-Sent Events (backend/internal/adapters/http/)
- Task 3.1: Update backend/internal/adapters/http/dto.go:
  - Extend generic ResponsePayload constraint interface to permit domain.LiveSessionTree and domain.LiveTraceView.
  - Add helper WriteLiveSessionTreeResponse(w http.ResponseWriter, tree domain.LiveSessionTree).
  - Add helper WriteLiveTraceResponse(w http.ResponseWriter, trace domain.LiveTraceView).
- Task 3.2: Implement HTTP handlers in backend/internal/adapters/http/live_handlers.go:
  - LiveSessionsHandler(service LiveServiceInterface) returning http.HandlerFunc: calls service.GetSessionTree and writes JSON response.
  - LiveTraceHandler(service LiveServiceInterface) returning http.HandlerFunc: extracts session ID from path parameter, validates against path traversal characters, calls service.GetSessionTrace, and writes JSON response or 404 error.
  - LiveStreamHandler(service LiveServiceInterface) returning http.HandlerFunc: sets SSE headers (text/event-stream, no-cache, keep-alive, X-Accel-Buffering: no), flushes initial LiveSessionTree, subscribes to event channel, loops on incoming events flushing them via http.Flusher, and listens to r.Context().Done() to unregister subscriber cleanly.
- Task 3.3: Register routes in backend/internal/adapters/http/server.go:
  - Update NewServer constructor to accept optional LiveServiceInterface.
  - Register GET /api/live/sessions, GET /api/live/sessions/{id}/trace, and GET /api/live/stream.

### Phase 4: CLI Server Command Integration (backend/cmd/agentic-lens/main.go)
- Task 4.1: Update runServer in backend/cmd/agentic-lens/main.go:
  - Add --db-path flag to serverFlags to allow overriding database path.
  - Load project config using adaptersconfig.Load("").
  - Resolve database path from flag or config (projCfg.Database.Path).
  - If database path exists or is configured, initialize live.OpenReadOnlyHook, create live.NewLiveService, and start background polling worker.
  - Pass live service into adaptershttp.NewServer.
  - In graceful shutdown signal handler, cancel live service context, stop poller, and close SQLite hook.

### Phase 5: Adversarial Testing & Verification
- Task 5.1: Pure domain unit tests in backend/internal/domain/live_test.go:
  - Test InferTicketID with session titles: uppercase (AL-17: implement live hook), lowercase (fix bug on al-2), alternate configured prefixes (gen-42 refactoring, PO-2 requirements), and titles without tickets (Exploration session, empty string, whitespace) asserting nil and 0.0 confidence.
  - Test ClassifyToolFailure against traceback strings, exit codes, missing files, repeated consecutive errors, and clean outputs.
  - Test BuildSessionTree with multi-level hierarchies, multiple concurrent roots, and orphaned children.
- Task 5.2: SQLite boundary and concurrency tests in backend/internal/adapters/live/sqlite_test.go:
  - Test read-only pragma enforcement: assert that attempting write statements returns a readonly error.
  - Concurrency test: start a concurrent writer goroutine holding an active write transaction in an in-memory/temporary SQLite database, and assert that the read-only hook executes queries concurrently without returning busy errors or blocking.
  - Transient buffer test: assert that memory buffer caps at max_trace_buffer_messages and does not write to disk.
  - Zero disk persistence test: assert that workspace directories and ledger files remain completely untouched during trace querying and tree generation.
- Task 5.3: HTTP live handlers integration tests in backend/internal/adapters/http/live_handlers_test.go:
  - Test GET /api/live/sessions returning status 200 with session tree JSON.
  - Test GET /api/live/sessions/{id}/trace returning 200 for valid session and 404 for missing session.
  - Test GET /api/live/stream verifying SSE headers, initial event push, and clean connection termination on client disconnect.
- Task 5.4: Run static checks and race detector:
  - Run go vet ./... and go test -race ./....
  - Verify zero any or empty interface{} across all added files.
  - Verify zero emojis across all code, tests, and documentation.

---

## 4. Adversarial Testing Strategy for Tester

### Pure Domain Unit Tests (backend/internal/domain/live_test.go)
- Test Case 1: Ticket Inference - Session Title Exact Prefix
  - Input: Session title "Task AL-17: implement live database hook". Assert guessed ID is "AL-17" with confidence 1.0.
- Test Case 2: Ticket Inference - Session Title Case-Insensitive Matching
  - Inputs: "working on al-17", "GEN-42 bugfix", "po-3 spec decomposition". Assert normalized to "AL-17", "GEN-42", "PO-3" with confidence 1.0.
- Test Case 3: Ticket Inference - Session Title Missing Ticket Identifier
  - Inputs: "General debugging session", "Quick question", "", "   ". Assert guessed ID is nil and confidence is 0.0.
- Test Case 4: Ticket Inference - Unconfigured Prefix
  - Input: "PROJ-1234 other project task". Assert guessed ID is nil and confidence is 0.0.
- Test Case 5: Tool Failure Classification - Python & Go Tracebacks
  - Input: Output with Traceback (most recent call last): or panic: runtime error. Assert isError = true, severity = "high".
- Test Case 6: Tool Failure Classification - Consecutive Failure Escalation
  - Input: Command failure with exit status 1 and consecutiveErrors = 2. Assert severity is escalated to high.
- Test Case 7: Tool Failure Classification - Resource Missing
  - Input: Output cat: file.txt: No such file or directory. Assert isError = true, severity = "medium".
- Test Case 8: Tool Failure Classification - Successful Execution
  - Input: Status completed with clean stdout. Assert isError = false, severity = "none", errorMessage = nil.
- Test Case 9: Session Tree Assembly - Hierarchy & Sorting
  - Input: Flat session list with two roots and multiple children. Assert roots sorted by time_updated descending and children nested under appropriate parent sorted by time_created ascending.

### Boundary & Concurrency Tests (backend/internal/adapters/live/sqlite_test.go)
- Test Case 10: Read-Only Pragma Enforcement
  - Attempt executing INSERT INTO session ... on read-only hook. Assert error contains attempt to write a readonly database.
- Test Case 11: Non-Locking Concurrency Under Write Load
  - Open a second connection in read-write mode. Begin an active exclusive write transaction modifying data. Simultaneously call FetchActiveSessions on the read-only hook. Assert query completes within 500ms without returning database is locked or busy error.
- Test Case 12: Transient Memory Buffer Boundary
  - Push 600 messages into a buffer configured with max_trace_buffer_messages = 500. Assert buffer contains exactly 500 messages and oldest 100 were dropped.
- Test Case 13: Zero Disk Persistence Verification
  - Record directory contents and file modification timestamps in workspaces/. Execute session tree generation, trace extraction, and SSE streaming. Re-scan workspaces/ and assert zero bytes were written and zero files were modified.

### HTTP Integration Tests (backend/internal/adapters/http/live_handlers_test.go)
- Test Case 14: GET /api/live/sessions
  - Issue request with httptest. Assert HTTP 200, valid Content-Type application/json, and response body unmarshals to LiveSessionTree.
- Test Case 15: GET /api/live/sessions/:id/trace - Valid and Invalid
  - Query valid ID: assert HTTP 200 and unmarshals to LiveTraceView.
  - Query non-existent ID: assert HTTP 404 with structured error response.
  - Query ID with path traversal (../etc): assert HTTP 400.
- Test Case 16: GET /api/live/stream - SSE Protocol & Context Cancellation
  - Open request, verify headers (text/event-stream), read initial session_tree event. Cancel request context and assert handler exits immediately without leaking goroutines.

---

## 5. Review Criteria for Reviewer

### Axis 1: Specification & Contract Alignment
- Does backend/internal/domain/live.go define all required models (LiveSessionNode, LiveSessionTree, LivePartFact, LiveMessageFact, LiveTraceView)?
- Does ClassifyToolFailure correctly identify high, medium, and low severity errors with consecutive failure escalation?
- Does InferTicketID inspect session titles exclusively, supporting case-insensitive prefix matching and returning nil / 0.0 confidence when no ticket is found in the title?
- Does sqlite.go open SQLite with read-only pragmas (query_only, busy_timeout, read_uncommitted)?
- Are the three HTTP endpoints (GET /api/live/sessions, GET /api/live/sessions/:id/trace, GET /api/live/stream) implemented with concrete typed responses?
- Is all live trace data strictly transient in-memory with zero disk persistence?

### Axis 2: Engineering Standards & Technical Invariants
- FCIS Discipline: Is backend/internal/domain/live.go a pure functional core with zero database, network, clock, or disk imports?
- Strong Typing: Is there zero usage of any or empty interface{} across Go packages?
- Value Semantics: Are domain models immutable structs with value receivers?
- Error Handling: Are all errors wrapped with %w and checked defensively?
- Emoji Prohibition: Are all files, comments, docstrings, logs, and artifacts 100% free of emojis?
- Quality Gates: Do go vet ./... and go test -race ./... pass with zero failures or warnings?
