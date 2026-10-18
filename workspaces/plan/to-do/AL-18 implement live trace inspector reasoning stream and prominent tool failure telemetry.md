h3. Title: AL-18: Implement Live Trace Inspector, Reasoning Stream, and Prominent Tool Failure Telemetry

*TARGET*: AgenticLens (domain | adapters | http | tests)

*CONTEXT*

When monitoring active agent workflows, developers need to inspect what is happening under the hood in real time without persisting transient data into disk ledgers. This includes inspecting the actual message contents exchanged between the user and agents, reasoning/thinking blocks produced by LLMs, and tool calls. Crucially, failed tool calls must be highlighted with high importance and severity so developers can immediately detect agent errors and execution loops.

*GOAL*

Build an in-memory transient trace extraction engine and HTTP API in Go to stream live message contents, reasoning thoughts, and tool executions for any active session, categorizing failed tool calls with high importance and delivering detailed error traces without persisting transient data to disk.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Create live trace models in `internal/domain/live.go`:
    - `LiveMessageFact` (`ID`, `Role`, `Content`, `Timestamp`, `ReasoningContent *string`, `Parts []LivePartFact`).
    - `LivePartFact` (`Type`, `ToolName`, `InputSnippet`, `OutputSnippet`, `Status`, `IsError`, `ErrorMessage *string`, `Severity string`).
    - `LiveTraceView` (`SessionID`, `Messages []LiveMessageFact`, `TotalToolCalls int`, `FailedToolCalls int`, `RecentErrors []LivePartFact`).
  - Pure tool failure classifier: determines severity (`high`, `medium`, `low`) based on error signatures (e.g. command failure, syntax error, exception traceback, repeated failure).
* Boundary Adapter & HTTP:
  - Add query methods in `internal/adapters/live/` to fetch messages and parts for a specific session ID without modifying DB state.
  - Parse `reasoning` content blocks and `tool` part states (`error`, `completed`) in memory.
  - Implement transient in-memory cache/ring buffer holding recent trace snapshots.
  - Expose HTTP endpoint: `GET /api/live/sessions/:id/trace` returning the structured `LiveTraceView`.
* Quality:
  - Table-driven Go tests for message and reasoning parsing, tool failure classification, and transient buffer boundaries.
  - Zero disk persistence assertions (guaranteeing nothing is written to `workspaces/` or database).
  - Race condition tests with `go test -race ./...`.

*ACCEPTANCE CRITERIA*

* Querying `/api/live/sessions/:id/trace` returns full message contents, thinking/reasoning blocks, and tool executions.
* Failed tool calls are tagged with `is_error: true`, elevated severity, and clear error message extracts.
* All live trace data remains strictly in-memory and transient, with zero disk writes to `agentic_metrics.json` or database tables.
* Database queries execute using read-only non-blocking queries.
* Go checks pass cleanly (`go vet ./...`, `go test -race ./...`) with zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified error prioritization logic with table-driven tests.
* Telemetry tested with `agentic-lens` CLI.
