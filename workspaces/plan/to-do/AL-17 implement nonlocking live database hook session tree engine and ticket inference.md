h3. Title: AL-17: Implement Non-Locking Live Database Hook, Session Tree Engine, and Config-Driven Ticket Inference

*TARGET*: AgenticLens (adapters | domain | http | tests)

*CONTEXT*

During multi-agent development sessions, operators have no live visibility into executing sessions. Directly polling the active SQLite database risks locking or delaying running OpenCode agents. Furthermore, concurrent workflows may have multiple root sessions running concurrently, each spawning child sub-sessions (e.g. planner, coder, tester, reviewer). A non-locking SQLite hook is needed to discover active sessions, reconstruct hierarchical root/sub-session trees, and heuristically infer which ticket is being worked on from session prompts and git context. To maximize accuracy, the inference engine must leverage the project configuration (`.agenticlens/config.yaml` / `config.json`) and support case-insensitive ticket recognition (handling lower-case tickets like `al-12` or `gen-2`).

*GOAL*

Deliver a read-only, non-locking SQLite database hook in Go using non-blocking pragmas, an active session discovery engine grouping concurrent root sessions with their child sub-sessions, a config-driven case-insensitive ticket inference algorithm, and HTTP endpoints (`/api/live/sessions` and SSE `/api/live/stream`) to serve live session hierarchies without database contention.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Create `internal/domain/live.go` defining live session tree models:
    - `LiveSessionNode` (`ID`, `ParentID`, `Agent`, `Model`, `TimeCreated`, `TimeUpdated`, `Status`, `GuessedTicketID`, `SubSessions []*LiveSessionNode`).
    - `LiveSessionTree` (`RootSessions []*LiveSessionNode`, `LastRefreshedAt string`).
  - Implement config-driven ticket inference heuristic:
    - Reads configured ticket prefixes (e.g. `AL`, `GEN`) and pattern rules from project configuration.
    - Matches case-insensitively across user prompts, session titles, and git branch names (e.g., `feat/al-12-...`, `al-4`, `GEN-2`).
    - Normalizes detected identifiers to canonical uppercase format (`AL-12`).
    - Scores candidate matches against existing workspace directories discovered via config.
* Boundary Adapter & HTTP:
  - Create `internal/adapters/live/sqlite.go`:
    - Opens SQLite connection with read-only and non-locking pragmas: `PRAGMA query_only = ON;`, `PRAGMA busy_timeout = 5000;`, WAL mode queries.
    - Queries active sessions (updated within the configured active window, defaulting to 300s, or with incomplete status) and builds the root/sub-session hierarchy.
  - Expose HTTP endpoints in `internal/adapters/http/handlers.go`:
    - `GET /api/live/sessions`: returns active session tree with guessed tickets.
    - `GET /api/live/stream`: Server-Sent Events (SSE) endpoint pushing live session tree changes.
* Quality:
  - Table-driven unit tests for ticket guessing heuristics covering upper-case, lower-case, mixed-case, and branch name formats.
  - Concurrency test ensuring database queries never lock SQLite write transactions or return busy errors.
  - Race condition tests with `go test -race ./...`.

*ACCEPTANCE CRITERIA*

* Querying `/api/live/sessions` returns active sessions grouped into root sessions with nested sub-sessions.
* SQLite connection operates strictly read-only with non-blocking pragmas, never locking active agent database writes.
* The ticket guessing algorithm correctly infers ticket IDs case-insensitively (e.g. `al-4`, `GEN-2`, `feat/al-12-migration`) utilizing configuration prefixes.
* Normalized canonical ticket IDs (e.g. `AL-4`) are assigned with confidence scoring.
* SSE endpoint `/api/live/stream` emits session hierarchy updates when new sessions or turns are detected.
* Go tests and static checks pass with zero errors and zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Verified non-locking behavior under simulated write load.
* Telemetry tested with `agentic-lens` CLI.
