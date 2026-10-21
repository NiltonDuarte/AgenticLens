# Contract: AL-17 Implement Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry

## 1. Problem Statement & Scope Boundaries

During autonomous multi-agent development sessions, operators have zero live visibility into running sessions, hierarchical sub-agent execution, or reasoning thought processes. Directly polling the active OpenCode SQLite database with default driver parameters risks contention, locking the database and causing running agent instances to fail with busy errors. Furthermore, multi-agent workflows frequently run multiple concurrent root sessions, each spawning specialized child sub-sessions (planner, coder, tester, reviewer).

Developers need an automated, non-locking SQLite database hook, an in-memory session tree reconstruction engine, a streamlined title-based ticket inference heuristic, and HTTP endpoints (REST and Server-Sent Events) to inspect live traces and tool failures. Relying on Git branch detection or workspace directory traversal is deliberately avoided because live sessions frequently execute inside detached worktrees where inspecting Git or filesystem state is undesirable and unreliable. If a ticket cannot be extracted directly from the session title, the system leaves the guessed ticket identifier empty, allowing users to override it manually in the UI. Crucially, all live trace telemetry must remain transient and in-memory, guaranteeing zero disk writes to workspaces or persistence ledgers.

### In Scope
1. Domain Functional Core (backend/internal/domain/live.go):
   - Pure domain models: LiveSessionNode, LiveSessionTree, LivePartFact, LiveMessageFact, and LiveTraceView.
   - Pure tool failure classifier: categorizes tool execution failures into severity tiers (high, medium, low, none) based on failure signatures, tracebacks, exit codes, and consecutive failure counts.
   - Pure session title ticket inference heuristic: inspects the session title exclusively, matches candidate ticket identifiers case-insensitively against configured ticket prefixes, and normalizes matches to canonical uppercase format (e.g. AL-17). Leaves the identifier empty (nil) with zero confidence when absent.
   - Pure tree hierarchy construction: groups root sessions and nests sub-sessions recursively based on parent-child relationships.
2. Imperative Shell Adapters (backend/internal/adapters/live/):
   - Read-only SQLite hook opening connection with non-blocking pragmas: PRAGMA query_only = ON;, PRAGMA busy_timeout = 5000;, and PRAGMA read_uncommitted = ON; in WAL mode.
   - Session discovery querying active sessions updated within the configured window (live.active_threshold_seconds, default 300s) or with active execution states.
   - Message and part extraction querying message and part tables, parsing reasoning content, and extracting tool execution states.
   - Transient in-memory ring buffer holding recent trace messages bounded by live.max_trace_buffer_messages (default 500).
   - Live service coordinator with background polling timer (live.poll_interval_ms, default 2000ms) and pub/sub broadcaster for Server-Sent Events.
3. HTTP Service Layer (backend/internal/adapters/http/):
   - GET /api/live/sessions: returns active LiveSessionTree with title-inferred tickets.
   - GET /api/live/sessions/:id/trace: returns LiveTraceView for a specific session.
   - GET /api/live/stream: Server-Sent Events (SSE) endpoint streaming session tree and trace snapshots.
   - DTO integration: concrete response typing extending ResponsePayload without using any.
4. CLI Server Wiring (backend/cmd/agentic-lens/main.go):
   - Wires live database hook and service inside runServer, accepting database parameters from .agenticlens/config.yaml with --db-path flag override.
5. Testing & Verification:
   - Table-driven unit tests for title-based ticket inference, tool failure classification, and session tree assembly.
   - Boundary tests verifying read-only enforcement and non-blocking reads under simulated SQLite write transactions.
   - Concurrency tests, transient buffer bounds checks, and race condition verification.

### Out of Scope
1. Git branch inspection, repository worktree traversal, or workspace folder scans for live ticket inference.
2. Frontend dashboard components and UI rendering (governed by AL-20).
3. Live anomaly detection rule evaluation (governed by AL-19).
4. Disk ledger writes, schema migration, or updates to workspaces/ files (strictly forbidden for live traces).
5. Direct schema migration or mutation of the OpenCode SQLite database.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

Dependencies flow strictly inward across layers:

### Domain Core Layer (backend/internal/domain/live.go)
- 100% pure functional core. Operates exclusively on immutable structs, slices, and primitive types.
- Pure value semantics: functions receive immutable data and return new domain models or explicit error indicators.
- Absolute isolation: zero imports of database/sql, os, io, net/http, git commands, or external logging libraries.
- No system clocks: functions receive timestamps (epoch milliseconds int64 or ISO strings) as parameters; time.Now() is strictly forbidden in the domain core.

### Imperative Shell Layer (backend/internal/adapters/live/, backend/internal/adapters/http/, backend/cmd/)
- Handles SQLite queries, connection pooling, background timer loops, memory buffers, and HTTP transport.
- SQLite connection configured with read-only, non-blocking pragmas.
- Memory isolation: live trace telemetry resides exclusively in transient memory buffers, guaranteeing zero disk writes.
- Errors wrapped with descriptive domain context (fmt.Errorf("...: %w", err)).

### Typing & Value Semantics
- Target Go 1.23+. Zero any or empty interface{} across domain and adapters.
- Use concrete structs, typed string enums (FailureSeverity), and parameterized generics where appropriate.

### Zero Emojis Mandate
- No emojis permitted in code, comments, docstrings, JSON keys, log lines, or error strings.

---

## 3. Domain Model Specifications (backend/internal/domain/live.go)

### Struct LiveSessionNode
- Field ID (string, JSON key id): unique session identifier.
- Field ParentID (pointer to string, JSON key parent_id, omitempty): parent session identifier for sub-sessions; null for root sessions.
- Field Agent (string, JSON key agent): raw persona name recorded on the session.
- Field Role (AgentRole, JSON key role): canonical agent role resolved via RoleMappings.
- Field Model (string, JSON key model): primary model identifier.
- Field Title (string, JSON key title): human-readable title of the session.
- Field TimeCreated (string, JSON key time_created): ISO 8601 creation timestamp.
- Field TimeUpdated (string, JSON key time_updated): ISO 8601 latest update timestamp.
- Field Status (string, JSON key status): active execution status (running, stalled, or completed).
- Field GuessedTicketID (pointer to string, JSON key guessed_ticket_id, omitempty): inferred canonical ticket identifier extracted from title; null if unassigned.
- Field TicketConfidence (float64, JSON key ticket_confidence): heuristic confidence score (1.0 for valid title match, 0.0 for unassigned).
- Field SubSessions (slice of pointers to LiveSessionNode, JSON key sub_sessions): nested child sub-sessions.

### Struct LiveSessionTree
- Field RootSessions (slice of pointers to LiveSessionNode, JSON key root_sessions): list of top-level root sessions.
- Field LastRefreshedAt (string, JSON key last_refreshed_at): ISO 8601 refresh timestamp.
- Field TotalActiveSessions (int, JSON key total_active_sessions): total count of active sessions across the tree.

### Struct LivePartFact
- Field ID (string, JSON key id): unique part identifier.
- Field Type (string, JSON key type): part kind (tool, reasoning, text, or compaction).
- Field ToolName (pointer to string, JSON key tool_name, omitempty): name of invoked tool.
- Field InputSnippet (string, JSON key input_snippet): truncated summary of tool input arguments (max 500 characters).
- Field OutputSnippet (string, JSON key output_snippet): truncated summary of tool execution output (max 1000 characters).
- Field Status (string, JSON key status): execution status (completed, error, or running).
- Field IsError (bool, JSON key is_error): true if execution failed or returned error output.
- Field ErrorMessage (pointer to string, JSON key error_message, omitempty): sanitized error excerpt.
- Field Severity (string, JSON key severity): error severity rating (high, medium, low, or none).

### Struct LiveMessageFact
- Field ID (string, JSON key id): unique message identifier.
- Field Role (string, JSON key role): message author role (user, assistant, or system).
- Field Content (string, JSON key content): primary textual content.
- Field Timestamp (string, JSON key timestamp): ISO 8601 creation timestamp.
- Field ReasoningContent (pointer to string, JSON key reasoning_content, omitempty): extracted chain-of-thought text.
- Field Parts (slice of LivePartFact, JSON key parts): execution parts associated with this message.

### Struct LiveTraceView
- Field SessionID (string, JSON key session_id): target session identifier.
- Field Messages (slice of LiveMessageFact, JSON key messages): full sequence of messages and execution steps.
- Field TotalToolCalls (int, JSON key total_tool_calls): aggregate count of tool calls.
- Field FailedToolCalls (int, JSON key failed_tool_calls): aggregate count of failed tool calls.
- Field RecentErrors (slice of LivePartFact, JSON key recent_errors): slice of recent failed tool calls ordered most recent first.

---

## 4. Pure Domain Algorithms

### Tool Failure Classifier (domain.ClassifyToolFailure)
Specification: Receives toolName as string, status as string, output as string, and consecutiveErrors as integer, returning isError as boolean, severity as string, and errorMessage as optional pointer to string.

Classification Rules:
1. High Severity:
   - Output contains tracebacks: Traceback (most recent call last), panic:, fatal error:, UnhandledPromiseRejection.
   - Tool execution fails compilation or syntax verification: go build errors, npm build parse errors, syntax errors.
   - Critical system faults: segmentation fault, killed, out of memory, permission denied.
   - Repeated failures: consecutiveErrors parameter is greater than or equal to 2.
2. Medium Severity:
   - Tool returns explicit error status or standard non-zero exit codes (exit status 1, command failed).
   - Missing resources: no such file or directory, file not found, path does not exist.
   - Network or I/O timeouts: connection refused, i/o timeout, context deadline exceeded.
3. Low Severity:
   - Benign empty searches (grep or glob returning no matches with non-zero exit code).
   - User cancelations or soft warnings.
4. None:
   - Status is completed and output lacks error signatures. Returns isError false, severity "none", and nil error message.

### Simplified Title-Based Ticket Inference Heuristic (domain.InferTicketID)
Specification: Receives sessionTitle as string and cfg as TicketConfig, returning guessedID as optional pointer to string and confidence as float64.

Heuristic Rules:
1. Input Sanitation:
   - Trims whitespace from sessionTitle. If empty, returns nil and confidence 0.0.
2. Prefix Matching & Normalization:
   - Scans sessionTitle using domain.NormalizeTicketID(sessionTitle, cfg).
   - Evaluates configured prefixes case-insensitively (e.g. al-17, AL-17, gen-4, po-2).
   - On match, formats the identifier to canonical uppercase prefix and digits (e.g. AL-17).
   - Returns pointer to the canonical string and confidence 1.0.
3. Fallback on Non-Match:
   - If no valid configured ticket pattern is found in the title, returns nil and confidence 0.0.
   - Leaves the field empty on the LiveSessionNode, preserving the raw session title for operators and allowing manual override in the UI.

---

## 5. Non-Locking SQLite Database Hook Contract

### Concurrency and Read-Only Guarantees (backend/internal/adapters/live/sqlite.go)
1. Connection String & Pragmas:
   - Connection DSN opened as file:<db_path>?mode=ro&_busy_timeout=5000.
   - Connection executes setup statements immediately upon opening:
     - PRAGMA query_only = ON; (prevents any data modification statements).
     - PRAGMA busy_timeout = 5000; (waits up to 5000ms if WAL index is momentarily busy).
     - PRAGMA read_uncommitted = ON; (enables non-blocking reads in WAL mode).
2. Pool Limits:
   - Max open connections set to 5; max idle connections set to 2.
3. Safe Failure on Mutation:
   - Any write query (INSERT, UPDATE, DELETE, CREATE) returns an explicit attempt to write a readonly database error from SQLite.
4. Non-Locking Concurrency Invariant:
   - Reader transactions must never block concurrent writer transactions executing on OpenCode.
   - Reads must succeed even when a write transaction is actively held by an external agent process.

### Zero Persistence Invariant
- The live engine is strictly read-only and in-memory.
- Zero writes to SQLite database files, WAL files, temporary tables, or journals.
- Zero writes to disk files, workspaces, or agentic_metrics.json.
- All extracted trace facts are maintained in transient in-memory ring buffers capped by live.max_trace_buffer_messages.

---

## 6. HTTP API Endpoints Contract

### Route 1: GET /api/live/sessions
- Handler: adaptershttp.LiveSessionsHandler(liveService)
- Description: Queries active sessions and returns the assembled hierarchical session tree with title-inferred ticket IDs.
- Response 200: JSON object conforming to domain.LiveSessionTree.
- Response 500: JSON ErrorResponse if database access fails.

### Route 2: GET /api/live/sessions/{id}/trace
- Handler: adaptershttp.LiveTraceHandler(liveService)
- Description: Fetches messages, reasoning blocks, and tool executions for the specified session ID.
- Response 200: JSON object conforming to domain.LiveTraceView.
- Response 400: JSON ErrorResponse if session ID contains invalid path traversal characters.
- Response 404: JSON ErrorResponse if the session ID does not exist in the database.

### Route 3: GET /api/live/stream
- Handler: adaptershttp.LiveStreamHandler(liveService)
- Description: Server-Sent Events (SSE) streaming endpoint.
- Headers:
  - Content-Type: text/event-stream
  - Cache-Control: no-cache
  - Connection: keep-alive
  - X-Accel-Buffering: no
- Lifecycle:
  - Pushes initial session_tree event immediately upon connection.
  - Subscribes to internal event broadcaster; flushes session_tree or trace_update events when session turns advance.
  - Listens to request context r.Context().Done() to unregister subscriber and prevent goroutine leaks.

---

## 7. Acceptance Criteria

1. Querying /api/live/sessions returns active sessions organized into root nodes with nested sub-sessions and inferred ticket IDs.
2. Querying /api/live/sessions/:id/trace returns full message contents, thinking/reasoning blocks, and tool execution parts.
3. Failed tool calls are tagged with is_error: true, assigned appropriate severity (high, medium, low), and provide sanitized error messages.
4. SQLite connection runs with read-only pragmas, never locks writer transactions, and fails immediately if writes are attempted.
5. Inferred ticket IDs are normalized from session titles into canonical uppercase format (AL-17, GEN-4) with confidence 1.0; sessions without matching titles return null for guessed ticket ID and confidence 0.0.
6. The SSE endpoint /api/live/stream streams live tree and trace updates to subscribers.
7. All live trace data remains strictly in-memory and transient; zero writes occur to workspaces/ or database files.
8. Zero any or empty interface{} across Go packages; zero emojis across all code, tests, logs, and artifacts.
9. All checks pass cleanly via go vet ./... and go test -race ./....
