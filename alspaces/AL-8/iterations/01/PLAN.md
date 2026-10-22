# Implementation Plan: AL-8 Anomaly API Endpoints and Fleet Health Aggregation

## 1. Goal and Iteration Scope

Implement the complete vertical slice for ticket AL-8:
1. Pure domain fleet anomaly rollup engine ComputeFleetAnomalies in backend/internal/domain/anomalies.go.
2. Domain model extensions (FleetAnomalySummary, BrittleToolSummary, FleetSummary.AnomalySummary) in backend/internal/domain/models.go.
3. Integration of fleet anomaly calculations into CalculateFleetSummary in backend/internal/domain/aggregations.go.
4. Boundary serialization DTOs (FleetAnomalySummaryDTO, BrittleToolSummaryDTO) in backend/internal/adapters/schemas.go.
5. HTTP DTO extensions, handler implementation, and route wiring for GET /api/workspaces/{ticketId}/events, GET /api/workspaces/{ticketId}, and GET /api/fleet in backend/internal/adapters/http/.
6. TypeScript domain types, API response types, API client methods, and mock data parity in web/src/.
7. Adversarial unit and HTTP integration test suites with zero domain mocks.

This plan contains zero code blocks in compliance with the Zero-Code Mandate.

---

## 2. Grounded Reference Architecture and Concrete Paths

### Domain Core Layer (backend/internal/domain/)
- backend/internal/domain/models.go: Define BrittleToolSummary, FleetAnomalySummary structs. Extend FleetSummary with AnomalySummary field.
- backend/internal/domain/anomalies.go: Implement pure function ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary.
- backend/internal/domain/aggregations.go: Update CalculateFleetSummary to invoke ComputeFleetAnomalies and assign AnomalySummary to FleetSummary.
- backend/internal/domain/anomalies_test.go: Add table-driven unit tests for ComputeFleetAnomalies covering clean executions, multiple anomalies, brittle tools sorting, empty inputs, and v1/v2 mixtures.

### Boundary Adapters Layer (backend/internal/adapters/)
- backend/internal/adapters/schemas.go: Define BrittleToolSummaryDTO, FleetAnomalySummaryDTO with bidirectional conversion methods ToDomain and FromDomain. Update FleetSummaryDTO if present.
- backend/internal/adapters/repository.go: Verify existing workspace loading and event retention logic.

### HTTP Server Layer (backend/internal/adapters/http/)
- backend/internal/adapters/http/dto.go: Extend WorkspaceDetailResponse with Anomalies []domain.Anomaly. Expand ResponsePayload type constraint to accept []domain.EventFact. Add helper WriteEventsResponse.
- backend/internal/adapters/http/handlers.go: Update WorkspaceDetailHandler to populate top-level Anomalies. Implement WorkspaceEventsHandler(workspacesDir string) http.HandlerFunc with path traversal checks, workspace lookup, chronological sorting, and v1 fallback.
- backend/internal/adapters/http/server.go: Register GET /api/workspaces/{ticketId}/events in NewServer router.
- backend/internal/adapters/http/server_test.go: Add HTTP integration tests asserting responses for workspace events (v1, v2, 404, 400), workspace detail anomalies, and fleet anomaly summary.

### TypeScript Presentation Layer (web/src/)
- web/src/domain/types.ts: Add BrittleToolSummary, FleetAnomalySummary interfaces. Extend FleetSummary with anomaly_summary. Extend WorkspaceDetail with anomalies.
- web/src/api/types.ts: Ensure type definitions and re-exports match domain types.
- web/src/api/client.ts: Export fetchWorkspaceEvents(ticketId: string): Promise<ApiResponse<readonly EventFact[]>>.
- web/src/api/mockData.ts: Update mockFleetSummary to provide mock anomaly_summary, and update mockWorkspaceDetails to supply mock anomalies.

---

## 3. Phased Implementation Tasks

### Phase 1: Domain Core Models and Fleet Rollup Engine (backend/internal/domain/)
- Task 1.1: Define BrittleToolSummary struct in backend/internal/domain/models.go with fields:
  - ToolName (string, json tag "tool_name")
  - AnomalyCount (int64, json tag "anomaly_count")
  - TotalCalls (int64, json tag "total_calls")
  - ErrorCount (int64, json tag "error_count")
  - FailureRatePct (float64, json tag "failure_rate_pct")
- Task 1.2: Define FleetAnomalySummary struct in backend/internal/domain/models.go with fields:
  - TotalAnomalies (int64, json tag "total_anomalies")
  - WarningCount (int64, json tag "warning_count")
  - CriticalCount (int64, json tag "critical_count")
  - SeverityCounts (map[string]int64, json tag "severity_counts")
  - RuleCounts (map[string]int64, json tag "rule_counts")
  - TotalTickets (int64, json tag "total_tickets")
  - TicketsWithAnomalies (int64, json tag "tickets_with_anomalies")
  - AnomalyRatePerTicket (float64, json tag "anomaly_rate_per_ticket")
  - BrittleTools ([]BrittleToolSummary, json tag "brittle_tools")
- Task 1.3: Update FleetSummary struct in backend/internal/domain/models.go:
  - Add field AnomalySummary of type FleetAnomalySummary (json tag "anomaly_summary")
- Task 1.4: Implement pure function ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary in backend/internal/domain/anomalies.go:
  - Handle empty ledgers input: return zeroed struct with initialized non-nil maps and empty slice for BrittleTools.
  - Traverse ledgers: evaluate DetectAnomalies(ledger) for each ticket.
  - Accumulate severity counts (WARNING, CRITICAL), rule counts (all four rule IDs), total anomalies, and tickets containing at least one anomaly.
  - Aggregate tool calls, tool errors, and tool-associated anomalies across all ledgers (evaluating both Events and Tools).
  - Compute anomaly rate per ticket safely avoiding division by zero.
  - Identify tools having errors > 0 or anomalies > 0 as brittle tools. Compute failure rate percentage per tool.
  - Sort brittle tools deterministically: AnomalyCount descending, ErrorCount descending, FailureRatePct descending, ToolName ascending.
- Task 1.5: Update CalculateFleetSummary in backend/internal/domain/aggregations.go:
  - Extract fact ledgers from ledgers slice.
  - Invoke ComputeFleetAnomalies(factLedgers) and assign to AnomalySummary on FleetSummary.

### Phase 2: Boundary Adapters and Schema DTOs (backend/internal/adapters/)
- Task 2.1: Implement BrittleToolSummaryDTO in backend/internal/adapters/schemas.go:
  - Fields matching domain model with JSON tags.
  - Methods ToDomain() domain.BrittleToolSummary and standalone BrittleToolSummaryDTOFromDomain(domain.BrittleToolSummary) BrittleToolSummaryDTO.
- Task 2.2: Implement FleetAnomalySummaryDTO in backend/internal/adapters/schemas.go:
  - Fields matching domain model with JSON tags.
  - Methods ToDomain() domain.FleetAnomalySummary and standalone FleetAnomalySummaryDTOFromDomain(domain.FleetAnomalySummary) FleetAnomalySummaryDTO.
- Task 2.3: Ensure bidirectional mapping between domain types and boundary DTOs is verified and covered by existing or updated schemas tests.

### Phase 3: HTTP Handlers, DTO Responses, and Route Wiring (backend/internal/adapters/http/)
- Task 3.1: Update WorkspaceDetailResponse in backend/internal/adapters/http/dto.go:
  - Add field Anomalies of type []domain.Anomaly with json tag "anomalies".
- Task 3.2: Update ResponsePayload constraint in backend/internal/adapters/http/dto.go:
  - Add []domain.EventFact to the allowed type union: ErrorResponse | []domain.AgenticMetricsLedger | WorkspaceDetailResponse | domain.FleetSummary | []domain.EventFact.
- Task 3.3: Add WriteEventsResponse in backend/internal/adapters/http/dto.go:
  - Accepts http.ResponseWriter and []domain.EventFact, ensuring nil slices map to empty slice before calling WriteJSON with HTTP 200.
- Task 3.4: Update WorkspaceDetailHandler in backend/internal/adapters/http/handlers.go:
  - Pass targetLedger.Anomalies (defaulting to empty slice if nil) into WorkspaceDetailResponse.
- Task 3.5: Implement WorkspaceEventsHandler(workspacesDir string) http.HandlerFunc in backend/internal/adapters/http/handlers.go:
  - Extract ticketId via r.PathValue("ticketId").
  - Validate ticketId: reject empty strings, parent traversal elements (..), slashes, and backslashes with HTTP 400 Bad Request.
  - Scan workspaces using adapters.ScanWorkspaces(workspacesDir) to locate matching ticketId.
  - If ticket not found: write HTTP 404 Not Found error response.
  - If ticket found: extract Events from targetLedger.Facts.
  - If Events is nil or empty: return empty slice []domain.EventFact{}.
  - If Events has entries: verify chronological ordering (sort by timestamp ascending, turn ascending), and return via WriteEventsResponse.
- Task 3.6: Update NewServer in backend/internal/adapters/http/server.go:
  - Register route mux.HandleFunc("GET /api/workspaces/{ticketId}/events", WorkspaceEventsHandler(workspacesDir)).

### Phase 4: TypeScript Presentation Layer Parity and API Client (web/src/)
- Task 4.1: Update web/src/domain/types.ts:
  - Define interface BrittleToolSummary with readonly fields: tool_name, anomaly_count, total_calls, error_count, failure_rate_pct.
  - Define interface FleetAnomalySummary with readonly fields: total_anomalies, warning_count, critical_count, severity_counts, rule_counts, total_tickets, tickets_with_anomalies, anomaly_rate_per_ticket, brittle_tools.
  - Extend FleetSummary interface with readonly anomaly_summary: FleetAnomalySummary.
  - Extend WorkspaceDetail interface with readonly anomalies: readonly Anomaly[].
- Task 4.2: Update web/src/api/types.ts:
  - Re-export or verify parity for any response wrapper types if necessary.
- Task 4.3: Update web/src/api/mockData.ts:
  - Add mock anomaly_summary to mockFleetSummary with realistic sample counts and brittle tools.
  - Ensure mockWorkspaceDetails entries have sample anomalies populated.
- Task 4.4: Implement fetchWorkspaceEvents in web/src/api/client.ts:
  - Function signature: fetchWorkspaceEvents(ticketId: string): Promise<ApiResponse<readonly EventFact[]>>.
  - Calls GET /api/workspaces/{ticketId}/events.
  - On 404 status: throws descriptive not found error.
  - On network error or other failure: falls back to empty array with isMock: true.
  - On 200 OK: parses JSON as readonly EventFact[] with isMock: false.

### Phase 5: Adversarial Unit, Integration, and Quality Verification
- Task 5.1: Pure domain unit tests in backend/internal/domain/anomalies_test.go:
  - Test ComputeFleetAnomalies with nil and empty slices (asserts zero totals, non-nil maps, empty brittle tools).
  - Test ComputeFleetAnomalies with clean tickets (zero anomalies, zero brittle tools).
  - Test ComputeFleetAnomalies with tickets exhibiting tool retry streaks, tool failure rates, context spikes, and rework loops.
  - Test severity count accumulation and rule counts breakdown.
  - Test anomaly rate per ticket calculation with precision rounding.
  - Test brittle tools sorting: tool with higher anomaly count ranks first; tie-breakers on error count, failure rate, and name.
  - Test mixed v1 and v2 ledgers: verifying that v1 ledgers contribute tool failure rate anomalies and peak context spikes without errors.
  - Zero mocks used throughout domain tests.
- Task 5.2: Aggregation regression tests:
  - Verify TestCalculateFleetSummary in backend/internal/domain/aggregations_test.go passes.
  - Verify TestFleetSummaryAndComparisonParity in backend/internal/adapters/parity_test.go passes without regression.
- Task 5.3: HTTP integration tests in backend/internal/adapters/http/server_test.go:
  - Test GET /api/workspaces/{ticketId} returns 200 with populated anomalies field.
  - Test GET /api/workspaces/{ticketId}/events returns 200 with ordered chronological event list for v2 workspaces.
  - Test GET /api/workspaces/{ticketId}/events returns 200 with empty array for legacy v1 workspaces.
  - Test GET /api/workspaces/{ticketId}/events returns 404 for non-existent ticket.
  - Test GET /api/workspaces/{ticketId}/events returns 400 for path traversal input.
  - Test GET /api/fleet returns 200 with anomaly_summary object, asserting non-nil brittle_tools and accurate severity counts.
- Task 5.4: Full quality checks:
  - Run go vet ./... across the repository.
  - Run go test -race ./... across all backend packages.
  - Run npm run typecheck in web/.
  - Run npm test in web/.
  - Verify zero emojis across all code, tests, docstrings, and response structures.
  - Verify zero any or empty interface{} across Go and TypeScript.

---

## 4. Execution Dependencies and Delivery Sequence

1. Task 1.1 to 1.5 (Domain models, ComputeFleetAnomalies, CalculateFleetSummary) -> Unblocks domain unit tests and boundary adapters.
2. Task 2.1 to 2.3 (Boundary schemas and DTOs) -> Unblocks HTTP layer.
3. Task 3.1 to 3.6 (HTTP handlers, DTOs, route registration) -> Unblocks HTTP integration tests.
4. Task 4.1 to 4.4 (TypeScript types, API client, mock data) -> Unblocks frontend typecheck and AL-9 visual waterfall.
5. Task 5.1 to 5.4 (Unit tests, HTTP integration tests, quality gates) -> Completes iteration verification.
