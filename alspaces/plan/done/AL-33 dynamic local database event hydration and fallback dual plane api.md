h3. Title: AL-33: End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall

*TARGET*: AgenticLens (domain | adapters | http | presentation | web | tests)

*CONTEXT*

With committed ledgers sanitized, routine events pruned, and database error strings stripped to bare failure indicators, full diagnostic context (prompts, tool inputs/outputs, LLM reasoning thoughts, raw error logs) is reserved strictly for local hydration.
Per the Dual-Plane Architecture in `RFC-telemetry-sanitization-and-local-hydration.md`, the backend server (`agentic-lens serve`) must dynamically hydrate rich event traces in memory from the local OpenCode SQLite database when available, while falling back gracefully to the committed bare ledger when running in CI, Docker, or remote environments without local database access.
Furthermore, the TypeScript web dashboard (`TraceWaterfall.tsx` and autopsy views) must deliver a seamless experience: rendering expandable prompt, reasoning, and tool drawers when hydrated locally (`is_hydrated: true`), and displaying a clean, informative bare error timeline with an offline indicator banner when running in unhydrated fallback mode (`is_hydrated: false`).

*GOAL*

Deliver an end-to-end vertical slice for dynamic local database hydration spanning Go backend adapters/HTTP and TypeScript frontend components, querying the local OpenCode SQLite database to dynamically hydrate rich in-memory event traces when present, seamlessly degrading to committed bare events when offline, and rendering rich interactive drawers or compact fallback error waterfalls without persisting transient data to disk.

*VERTICAL SLICE SCOPE*

* Domain Core (`internal/domain/`):
  - Define dual-plane event models in `internal/domain/models.go`:
    - `WorkspaceEventsResponse`: `TicketID string`, `IsHydrated bool`, `Events []EventFact`, `DataSource string` (`"local_sqlite"` or `"committed_ledger"`).
* Boundary Adapter & HTTP (`internal/adapters/http/`, `internal/adapters/live/` or `internal/adapters/opencode/`):
  - Implement non-locking SQLite query reader to fetch full chronological messages, parts, reasoning thoughts, and tool inputs/outputs by ticket session ID(s).
  - Open SQLite connection with non-blocking pragmas: `PRAGMA query_only = ON;`, `PRAGMA busy_timeout = 5000;`, and WAL mode.
  - Expose/enhance HTTP endpoint `GET /api/workspaces/:id/events`:
    - Checks whether ticket session exists in local OpenCode SQLite database.
    - If found: dynamically constructs rich hydrated event stream and returns with `is_hydrated: true` and `data_source: "local_sqlite"`.
    - If not found or SQLite inaccessible: reads bare events from `workspaces/<ID>/agentic_metrics.json` and returns with `is_hydrated: false` and `data_source: "committed_ledger"`.
  - Strictly enforce zero disk writes of transient hydrated data.
* Presentation & Web Components (`web/src/components/`, `web/src/views/`):
  - Update `TraceWaterfall.tsx` and autopsy views:
    - Detect `is_hydrated` flag from API response.
    - When `is_hydrated == true`: render full waterfall with expandable drawers for message prompts, LLM reasoning thought blocks, and formatted tool parameters/outputs.
    - When `is_hydrated == false`: render compact bare error waterfall with a clear badge/banner ("Committed Git Ledger — Local Database Offline"), informing the user that full prompts, tool inputs, and error contexts are available only when running locally against the session database.
  - Detail drawers/modal for inspecting tool arguments and outputs in hydrated mode adhering strictly to zero-emoji design.
  - Update `api.ts` with strict TypeScript types for `WorkspaceEventsResponse` and `is_hydrated`.
* Quality & Tests:
  - Unit tests for `WorkspaceEventsResponse` serialization with `is_hydrated` true and false.
  - Mock-free SQLite integration tests verifying rich event reconstruction against temporary SQLite databases.
  - Fallback tests verifying graceful degradation when SQLite file does not exist or database is locked.
  - Vitest component tests testing both `is_hydrated: true` (rich drawer view) and `is_hydrated: false` (compact fallback view).
  - Concurrency and race tests with `go test -race ./...` and `npm run typecheck`.

*ACCEPTANCE CRITERIA*

* `GET /api/workspaces/:id/events` returns rich events with `is_hydrated: true` when local SQLite database contains the session.
* `GET /api/workspaces/:id/events` returns bare committed events with `is_hydrated: false` when local database is missing.
* Database access uses non-blocking read-only pragmas, never conflicting with active agent sessions.
* Zero transient trace data is written to disk or workspace directories.
* `TraceWaterfall` renders rich prompt/reasoning/tool drawers when `is_hydrated` is true.
* `TraceWaterfall` renders compact bare error timeline with an informational offline banner when `is_hydrated` is false.
* All backend tests (`go test -race ./...`) and frontend tests (`npm test`, `npm run typecheck`) pass cleanly with zero `any` and zero emojis.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
