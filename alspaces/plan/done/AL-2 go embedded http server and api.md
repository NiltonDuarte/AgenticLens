h3. Title: AL-2: Implement Go HTTP Server, Embedded Static FS, and Telemetry REST API

*TARGET*: AgenticLens (adapters | cli | tests)

*CONTEXT*

AgenticLens needs an HTTP service layer to expose scanned telemetry data to the web frontend and serve static SPA assets as a single executable without Node.js.

*GOAL*

Implement the HTTP server command `agentic-lens server` that serves REST API endpoints (`/api/workspaces`, `/api/workspaces/:ticketId`, `/api/fleet`) and embeds static frontend assets via `embed.FS` with SPA fallback routing.

*VERTICAL SLICE SCOPE*

* Domain Core: Serialization-ready domain view models and filter predicates in `internal/domain` (time window filtering, ticket sorting).
* Boundary Adapter: `internal/adapters/http`: HTTP handlers for `/api/workspaces`, `/api/workspaces/{ticketId}`, `/api/fleet`, CORS middleware, structured error responses, and SPA file server with `index.html` fallback. `internal/adapters/web`: `embed.FS` container.
* Presentation: Subcommand or flag `agentic-lens server --port <port> --workspaces <path>`.
* Quality: HTTP integration tests using `httptest.NewServer` verifying status codes, payload shapes, 404 handling, and fallback asset routing.

*ACCEPTANCE CRITERIA*

* `GET /api/workspaces` returns JSON array of all scanned workspace ledgers.
* `GET /api/workspaces/{ticketId}` returns the target ledger and artifact contents, or 404 if missing.
* `GET /api/fleet` returns aggregated fleet metrics across all tickets.
* Non-API routes serve embedded static assets or fall back to `index.html`.
* `go test -race ./...` passes.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Zero any in Go handlers and payloads.
* Telemetry tested with `agentic-lens` CLI.
