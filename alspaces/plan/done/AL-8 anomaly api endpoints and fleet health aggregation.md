h3. Title: AL-8: Anomaly API Endpoints and Fleet Health Aggregation

*TARGET*: AgenticLens (adapters | http | server | tests)

*CONTEXT*

With the domain anomaly engine and v2 event facts in place, the HTTP server must expose anomaly findings and chronological events to API consumers, and roll up fleet-wide anomaly metrics to help team leads assess agent stability across tickets.

*GOAL*

Expose REST endpoints for ticket anomalies and event timelines, and integrate fleet-wide anomaly statistics into the `/api/fleet` endpoint.

*VERTICAL SLICE SCOPE*

* Domain Core: Pure fleet anomaly rollup function `ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary` calculating total anomalies, anomaly rates per ticket, most brittle tools, and severity breakdown.
* Boundary Adapter & HTTP Server:
  - Update `internal/adapters/http`:
    - Extend `GET /api/workspaces/:ticketId` response DTO to include `anomalies: []AnomalyDTO`.
    - Implement `GET /api/workspaces/:ticketId/events` returning chronologically sorted event facts.
    - Extend `GET /api/fleet` response DTO to include `anomaly_summary: FleetAnomalySummaryDTO`.
* Quality: HTTP integration tests using `net/http/httptest` asserting accurate JSON responses for ticket anomalies, event streams, and fleet rollups. Zero domain mocks.

*ACCEPTANCE CRITERIA*

* `GET /api/workspaces/:ticketId` includes evaluated anomalies array in response payload.
* `GET /api/workspaces/:ticketId/events` returns ordered chronological event list for v2 workspaces and empty list for legacy v1 workspaces.
* `GET /api/fleet` reports fleet-level anomaly counts and high-frequency failure tools.
* Unmatched workspace returns clean 404 response.
* `go test -race ./...` and `go vet ./...` pass with zero errors.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`go vet ./...`, `go test -race ./...`).
* Go type checking clean with zero any or empty `interface{}`.
* Telemetry tested with `agentic-lens` CLI.
