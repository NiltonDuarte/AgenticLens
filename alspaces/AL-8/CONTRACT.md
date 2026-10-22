# Contract: AL-8 Anomaly API Endpoints and Fleet Health Aggregation

## 1. Problem Statement

Ticket AL-7 introduced the domain anomaly rule engine (DetectAnomalies) evaluating tool retry streaks, tool failure rates, context window spikes, and subagent rework loops against raw fact ledgers. Ticket AL-6 delivered chronological turn-level event extraction (EventFact) captured in the v2 schema.

However, the HTTP server layer currently isolates these analytical capabilities from API consumers:
1. Workspace Detail Incompleteness: GET /api/workspaces/{ticketId} returns an AgenticMetricsLedger and TicketArtifacts, but lacks explicit, top-level serialization of evaluated anomalies in its response contract.
2. Missing Event Timelines: There is no HTTP endpoint allowing web dashboards or API clients to fetch ordered chronological event streams for a workspace. UI waterfall rendering planned in AL-9 cannot proceed without this endpoint.
3. Fleet Health Blindness: GET /api/fleet aggregates token volumes, runtimes, lines of code, and model costs, but provides zero insight into fleet-wide operational health, tool brittleness, or ticket anomaly frequencies. Engineering leads cannot determine whether agent failure patterns are systemic across the workspace fleet.

Ticket AL-8 bridges this gap by:
- Constructing a pure functional fleet anomaly rollup in the Go domain: ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary.
- Exposing GET /api/workspaces/{ticketId}/events returning chronologically sorted EventFact records (clean empty array for legacy v1 workspaces, 404 for missing workspaces).
- Extending the GET /api/workspaces/{ticketId} response DTO to include evaluated anomalies.
- Extending GET /api/fleet with anomaly_summary: FleetAnomalySummaryDTO.
- Establishing complete TypeScript type and API client contract parity across web/src/domain/types.ts and web/src/api/.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation across Go backend and TypeScript frontend must adhere strictly to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. HTTP Server and Handler Layer (backend/internal/adapters/http/):
   - Handles route matching, URL parameter validation, path traversal guards, HTTP status writing, and JSON serialization.
   - Contains zero anomaly detection heuristics, threshold checks, or mathematical calculations.
2. Boundary Adapters and Schemas Layer (backend/internal/adapters/):
   - Maps between domain entities and JSON DTOs (schemas.go).
   - Scans workspaces and reads files safely from disk (repository.go).
   - Does not perform domain business logic or metric aggregation.
3. Functional Core Domain Layer (backend/internal/domain/):
   - Houses pure models (FleetAnomalySummary, BrittleToolSummary) and pure aggregation logic (ComputeFleetAnomalies).
   - Operates purely on immutable structs passed as value parameters and returns calculated value objects.
   - Contains zero external dependencies: no net/http, no os, no io, no database/sql, and no system clocks.
   - Defensively guards against division by zero (empty ledgers slice, zero ticket count, zero tool calls).

### Pure Domain Core Rules
1. ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary:
   - Must be a pure deterministic function. Given identical inputs, it must return identical output.
   - Must evaluate DetectAnomalies for each ledger, accumulating total anomaly counts, severity breakdown (WARNING, CRITICAL), and rule counts.
   - Must track tickets containing at least one anomaly and compute the fleet anomaly rate per ticket rounded to two decimal places.
   - Must identify brittle tools across the fleet: aggregate tool calls, execution errors, and triggered anomalies per tool, calculating failure rates and sorting deterministically by anomaly count descending, error count descending, failure rate descending, and tool name ascending.
   - Must handle empty ledgers slices gracefully, returning a zero-value FleetAnomalySummary with empty slices and empty maps rather than nil pointers.
2. CalculateFleetSummary(ledgers []AgenticMetricsLedger) FleetSummary:
   - Must incorporate AnomalySummary: ComputeFleetAnomalies(factLedgers) by extracting Facts from each ledger.
   - Existing volume, duration, token, and cost metrics calculations must remain untouched and backward compatible.

### Imperative Shell Boundary Rules
1. Workspace Events Endpoint (GET /api/workspaces/{ticketId}/events):
   - Must validate ticketId to reject empty strings, parent traversals (..), and slash delimiters with HTTP 400 Bad Request.
   - Must scan workspaces to locate the requested ticketId. If not found, must return HTTP 404 Not Found with structured ErrorResponse.
   - For existing v2 workspaces, must return chronologically sorted []EventFact.
   - For existing legacy v1 workspaces lacking events, must return an empty JSON array [] with HTTP 200 OK.
2. Workspace Detail Endpoint (GET /api/workspaces/{ticketId}):
   - Response DTO (WorkspaceDetailResponse) must include anomalies: []domain.Anomaly.
   - If no anomalies are detected, the field must serialize as an empty JSON array [] rather than null.
3. Fleet Endpoint (GET /api/fleet):
   - Response DTO must include anomaly_summary: FleetAnomalySummaryDTO (or domain.FleetAnomalySummary).
   - Supports existing time window filtering (?window=7d, 30d, 90d, all) consistently across both volume metrics and anomaly aggregations.
4. Static Single-Binary and Zero Any:
   - Single Go binary distribution without CGO or Node.js runtime dependencies.
   - Zero any or empty interface{} across all Go backend and TypeScript frontend code.
   - HTTP response writer enforces type parameters constrained by ResponsePayload.
5. Strict Zero-Emoji Mandate:
   - Zero emojis across all source files, HTTP routes, DTOs, JSON payloads, test names, log messages, and documentation.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Domain Core Models and Rollup Engine (backend/internal/domain/):
   - Define BrittleToolSummary struct in backend/internal/domain/models.go.
   - Define FleetAnomalySummary struct in backend/internal/domain/models.go.
   - Extend FleetSummary in backend/internal/domain/models.go with AnomalySummary FleetAnomalySummary.
   - Implement pure ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary in backend/internal/domain/anomalies.go.
   - Update CalculateFleetSummary in backend/internal/domain/aggregations.go to populate AnomalySummary via ComputeFleetAnomalies.
2. Boundary Adapters and Serialization (backend/internal/adapters/):
   - Define BrittleToolSummaryDTO and FleetAnomalySummaryDTO in backend/internal/adapters/schemas.go with bidirectional domain mapping.
   - Update FleetSummaryDTO if applicable in backend/internal/adapters/schemas.go.
3. HTTP Server and Handler Layer (backend/internal/adapters/http/):
   - Update WorkspaceDetailResponse in backend/internal/adapters/http/dto.go to include Anomalies []domain.Anomaly.
   - Update ResponsePayload type constraint interface in backend/internal/adapters/http/dto.go to permit []domain.EventFact.
   - Add WriteEventsResponse in backend/internal/adapters/http/dto.go.
   - Update WorkspaceDetailHandler in backend/internal/adapters/http/handlers.go to populate the top-level anomalies field.
   - Implement WorkspaceEventsHandler(workspacesDir string) http.HandlerFunc in backend/internal/adapters/http/handlers.go.
   - Register route GET /api/workspaces/{ticketId}/events in backend/internal/adapters/http/server.go.
4. TypeScript Presentation Layer Parity (web/src/):
   - Define BrittleToolSummary interface in web/src/domain/types.ts.
   - Define FleetAnomalySummary interface in web/src/domain/types.ts.
   - Extend FleetSummary in web/src/domain/types.ts with readonly anomaly_summary: FleetAnomalySummary.
   - Extend WorkspaceDetail in web/src/domain/types.ts with readonly anomalies: readonly Anomaly[].
   - Add fetchWorkspaceEvents(ticketId: string): Promise<ApiResponse<readonly EventFact[]>> to web/src/api/client.ts.
   - Update web/src/api/mockData.ts to supply compliant mock anomaly summaries for fleet and workspace details.
5. Comprehensive Test Suites:
   - Pure domain unit tests in backend/internal/domain/anomalies_test.go covering ComputeFleetAnomalies with empty ledgers, clean executions, anomaly mixtures, brittle tool sorting, and division-by-zero guards. Zero mocks.
   - Domain aggregation regression tests in backend/internal/domain/aggregations_test.go and backend/internal/adapters/parity_test.go verifying CalculateFleetSummary parity.
   - HTTP integration tests in backend/internal/adapters/http/server_test.go using httptest verifying:
     - GET /api/workspaces/{ticketId} includes anomalies array.
     - GET /api/workspaces/{ticketId}/events returns sorted events for v2 workspaces.
     - GET /api/workspaces/{ticketId}/events returns empty array for v1 workspaces.
     - GET /api/workspaces/{ticketId}/events returns 404 for non-existent ticket.
     - GET /api/workspaces/{ticketId}/events returns 400 for path traversal input.
     - GET /api/fleet returns anomaly_summary with brittle tools and counts.
   - TypeScript verification via npm run typecheck and npm test.

### Out of Scope
1. Visual waterfall timeline UI component and interactive anomaly badge inspector view (Ticket AL-9 scope).
2. SQLite database extraction alterations or schema backfill changes (Tickets AL-6, AL-11, and AL-12 scope).
3. CLI flag changes or Markdown summary changes (Ticket AL-7 scope).

---

## 4. Architectural Specifications and Typed Pseudo-Contracts

### Domain Models Pseudo-Contract (backend/internal/domain/)

Pseudo-contract for BrittleToolSummary:
- Struct Name: BrittleToolSummary
- Fields:
  - ToolName: string, json tag "tool_name"
  - AnomalyCount: int64, json tag "anomaly_count"
  - TotalCalls: int64, json tag "total_calls"
  - ErrorCount: int64, json tag "error_count"
  - FailureRatePct: float64, json tag "failure_rate_pct"

Pseudo-contract for FleetAnomalySummary:
- Struct Name: FleetAnomalySummary
- Fields:
  - TotalAnomalies: int64, json tag "total_anomalies"
  - WarningCount: int64, json tag "warning_count"
  - CriticalCount: int64, json tag "critical_count"
  - SeverityCounts: map of string to int64, json tag "severity_counts"
  - RuleCounts: map of string to int64, json tag "rule_counts"
  - TotalTickets: int64, json tag "total_tickets"
  - TicketsWithAnomalies: int64, json tag "tickets_with_anomalies"
  - AnomalyRatePerTicket: float64, json tag "anomaly_rate_per_ticket"
  - BrittleTools: slice of BrittleToolSummary, json tag "brittle_tools"

Pseudo-contract for FleetSummary Extension:
- In struct FleetSummary:
  - Add field AnomalySummary of type FleetAnomalySummary, json tag "anomaly_summary"

### Fleet Anomaly Computation Heuristic (ComputeFleetAnomalies)
1. Initialize counters:
   - totalAnomalies = 0
   - warningCount = 0
   - criticalCount = 0
   - ticketsWithAnomalies = 0
   - severityCounts = map {"WARNING": 0, "CRITICAL": 0}
   - ruleCounts = map {"TOOL_RETRY_STREAK": 0, "TOOL_HIGH_FAILURE_RATE": 0, "CONTEXT_SATURATION_SPIKE": 0, "SUBAGENT_REWORK_LOOP": 0}
   - toolStats = map of tool name to {calls: int64, errors: int64, anomalies: int64}
2. For each ledger in ledgers:
   - Evaluate ticket anomalies: ticketAnomalies = DetectAnomalies(ledger)
   - If length of ticketAnomalies > 0: increment ticketsWithAnomalies
   - For each anomaly in ticketAnomalies:
     - Increment totalAnomalies
     - If severity is WARNING: increment warningCount and severityCounts["WARNING"]
     - If severity is CRITICAL: increment criticalCount and severityCounts["CRITICAL"]
     - Increment ruleCounts for string value of anomaly RuleID
     - If anomaly.Context.ToolName is non-nil and non-empty: increment toolStats[toolName].anomalies
   - Accumulate tool execution volume:
     - If ledger has Events: iterate tool_call events, incrementing calls and error counts per tool name
     - If ledger Events is empty: iterate ledger.Tools slice, incrementing calls and error counts per tool name
3. Compute anomalyRatePerTicket:
   - If totalTickets > 0: round (totalAnomalies / totalTickets) to 2 decimal places
   - Else: 0.0
4. Build and sort BrittleTools slice:
   - For each tool in toolStats where errors > 0 or anomalies > 0:
     - Compute failureRatePct = round (errors / calls * 1000) / 10 if calls > 0, else 0.0
     - Construct BrittleToolSummary
   - Sort slice stably:
     - AnomalyCount descending
     - ErrorCount descending
     - FailureRatePct descending
     - ToolName ascending
5. Return FleetAnomalySummary value object.

### HTTP Layer Pseudo-Contract (backend/internal/adapters/http/)

Pseudo-contract for WorkspaceDetailResponse:
- In struct WorkspaceDetailResponse:
  - Ledger: domain.AgenticMetricsLedger, json tag "ledger"
  - Artifacts: domain.TicketArtifacts, json tag "artifacts"
  - Anomalies: slice of domain.Anomaly, json tag "anomalies"

Pseudo-contract for ResponsePayload:
- Interface constraint: ErrorResponse | []domain.AgenticMetricsLedger | WorkspaceDetailResponse | domain.FleetSummary | []domain.EventFact

Route Registrations:
- GET /api/workspaces -> WorkspacesHandler
- GET /api/workspaces/{ticketId} -> WorkspaceDetailHandler
- GET /api/workspaces/{ticketId}/events -> WorkspaceEventsHandler
- GET /api/fleet -> FleetHandler

### TypeScript Frontend Pseudo-Contract (web/src/domain/types.ts)

Pseudo-contract for BrittleToolSummary:
- Interface: BrittleToolSummary
- Properties:
  - readonly tool_name: string
  - readonly anomaly_count: number
  - readonly total_calls: number
  - readonly error_count: number
  - readonly failure_rate_pct: number

Pseudo-contract for FleetAnomalySummary:
- Interface: FleetAnomalySummary
- Properties:
  - readonly total_anomalies: number
  - readonly warning_count: number
  - readonly critical_count: number
  - readonly severity_counts: Readonly<Record<string, number>>
  - readonly rule_counts: Readonly<Record<string, number>>
  - readonly total_tickets: number
  - readonly tickets_with_anomalies: number
  - readonly anomaly_rate_per_ticket: number
  - readonly brittle_tools: readonly BrittleToolSummary[]

---

## 5. Acceptance Criteria

1. GET /api/workspaces/{ticketId} response payload includes evaluated anomalies in both ledger.anomalies and top-level anomalies array.
2. GET /api/workspaces/{ticketId}/events returns chronologically sorted event list with HTTP 200 OK for v2 workspaces.
3. GET /api/workspaces/{ticketId}/events returns empty array [] with HTTP 200 OK for legacy v1 workspaces lacking events.
4. GET /api/workspaces/{ticketId}/events returns HTTP 404 Not Found with structured error JSON when ticketId does not exist.
5. GET /api/workspaces/{ticketId}/events returns HTTP 400 Bad Request when ticketId contains path traversal characters.
6. GET /api/fleet includes anomaly_summary reporting total anomalies, warning count, critical count, rule counts, tickets with anomalies, anomaly rate per ticket, and sorted brittle tools list.
7. ComputeFleetAnomalies safely handles nil or empty input without runtime panic, returning zero-value summary structures.
8. TypeScript interfaces in web/src/domain/types.ts provide 100% parity with backend domain models.
9. TypeScript client function fetchWorkspaceEvents is exported from web/src/api/client.ts.
10. Quality gates pass cleanly:
    - go vet ./... passes with zero warnings.
    - go test -race ./... passes across all packages with zero race conditions.
    - npm run typecheck in web/ passes with zero errors.
11. Strict zero-emoji policy enforced across all code, tests, documentation, DTOs, and HTTP responses.
