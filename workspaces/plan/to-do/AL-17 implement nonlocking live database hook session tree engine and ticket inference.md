h3. Title: AL-17: Implement Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry

*TARGET*: AgenticLens (adapters | domain | http | tests)

*CONTEXT*

During multi-agent development sessions, operators have no live visibility into executing sessions or their internal mechanics. Directly polling the active SQLite database risks locking or delaying running OpenCode agents. Furthermore, concurrent workflows may have multiple root sessions running concurrently, each spawning child sub-sessions (e.g. planner, coder, tester, reviewer). Developers need to discover active sessions, reconstruct hierarchical root/sub-session trees, heuristically infer which ticket is being worked on from session prompts and git context, and inspect live message contents, LLM reasoning streams, and prominent failed tool calls in real time without persisting transient data into disk ledgers.
(Note: Merged with former AL-18 to unify the live backend extraction vertical slice).

*GOAL*

Deliver a comprehensive read-only, non-locking SQLite database hook in Go using non-blocking pragmas, an active session discovery and tree reconstruction engine, a config-driven ticket inference algorithm, and HTTP endpoints (`/api/live/sessions`, `/api/live/sessions/:id/trace`, and SSE `/api/live/stream`) to serve live session hierarchies, message reasoning streams, and prominent tool failure telemetry without database contention or disk persistence.

*VERTICAL SLICE SCOPE*

* Domain Core (`internal/domain/live.go`):
  - Session Tree Models:
    - `LiveSessionNode` (`ID`, `ParentID`, `Agent`, `Model`, `TimeCreated`, `TimeUpdated`, `Status`, `GuessedTicketID`, `SubSessions []*LiveSessionNode`).
    - `LiveSessionTree` (`RootSessions []*LiveSessionNode`, `LastRefreshedAt string`).
  - Live Trace Models:
    - `LiveMessageFact` (`ID`, `Role`, `Content`, `Timestamp`, `ReasoningContent *string`, `Parts []LivePartFact`).
    - `LivePartFact` (`Type`, `ToolName`, `InputSnippet`, `OutputSnippet`, `Status`, `IsError`, `ErrorMessage *string`, `Severity string`).
    - `LiveTraceView` (`SessionID`, `Messages []LiveMessageFact`, `TotalToolCalls int`, `FailedToolCalls int`, `RecentErrors []LivePartFact`).
  - Pure tool failure classifier: determines error severity (`high`, `medium`, `low`) based on failure signatures (command failure, syntax error, exception traceback, repeated failure).
  - Config-driven ticket inference heuristic:
    - Reads configured ticket prefixes (e.g. `AL`, `GEN`, `PO`) from project configuration.
    - Matches case-insensitively across user prompts, session titles, and git branch names (e.g., `feat/al-12-...`, `al-4`, `GEN-2`).
    - Normalizes detected identifiers to canonical uppercase format (`AL-12`).
    - Scores candidate matches against existing workspace directories discovered via config.
* Boundary Adapter & HTTP (`internal/adapters/live/sqlite.go`, `internal/adapters/http/handlers.go`):
  - Opens SQLite connection with read-only and non-locking pragmas: `PRAGMA query_only = ON;`, `PRAGMA busy_timeout = 5000;`, and WAL mode queries.
  - Queries active sessions (updated within the configured active window, defaulting to 300s, or with incomplete status) and builds the root/sub-session hierarchy.
  - Queries messages and parts for specific sessions, extracting reasoning content blocks and tool part states (`error`, `completed`).
  - Transient in-memory cache / ring buffer holding recent trace snapshots.
  - Expose HTTP endpoints in `internal/adapters/http/handlers.go`:
    - `GET /api/live/sessions`: returns active session tree with guessed tickets.
    - `GET /api/live/sessions/:id/trace`: returns `LiveTraceView` for an active session.
    - `GET /api/live/stream`: Server-Sent Events (SSE) endpoint pushing live session tree and trace changes.
* Quality:
  - Table-driven unit tests for ticket guessing heuristics covering upper-case, lower-case, mixed-case, and branch name formats.
  - Table-driven unit tests for message and reasoning parsing, tool failure classification, and transient buffer boundaries.
  - Concurrency test ensuring database queries never lock SQLite write transactions or return busy errors.
  - Zero disk persistence assertions (guaranteeing nothing is written to `workspaces/` or database).
  - Race condition tests with `go test -race ./...`.

*ACCEPTANCE CRITERIA*

* Querying `/api/live/sessions` returns active sessions grouped into root sessions with nested sub-sessions and guessed ticket IDs.
* Querying `/api/live/sessions/:id/trace` returns full message contents, thinking/reasoning blocks, and tool executions.
* Failed tool calls are tagged with `is_error: true`, elevated severity, and clear error message extracts.
* SQLite connection operates strictly read-only with non-blocking pragmas, never locking active agent database writes.
* Normalized canonical ticket IDs (e.g. `AL-4`) are assigned with confidence scoring.
* SSE endpoint `/api/live/stream` emits session hierarchy and trace updates when new sessions or turns are detected.
* All live trace data remains strictly in-memory and transient, with zero disk writes to `agentic_metrics.json` or database tables.
* Go tests and static checks pass cleanly with zero errors and zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified non-locking behavior under simulated write load.
* Telemetry tested with `agentic-lens` CLI.
