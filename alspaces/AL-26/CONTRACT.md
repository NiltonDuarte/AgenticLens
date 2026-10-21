# Contract: AL-26 Fleet Overview Completion Date Filtering and Autopsy Date Sorting

## 1. Problem Statement

Across the AgenticLens observability platform, the Fleet Overview view aggregates metrics across every recorded workspace for project lifetime with zero temporal filtering. Users cannot isolate metrics to a specific release cycle, sprint, calendar month, or arbitrary date window.

Furthermore, the HTTP endpoints for /api/fleet and /api/workspaces lack support for explicit temporal boundaries (since and until query parameters). In the Ticket Deep Dive view, the Workspace Autopsy selector sorts tickets strictly by natural ticket identifier (PREFIX-NN). Users cannot order workspaces by completion date (completed_at), preventing rapid inspection of recently resolved workloads or chronologically ordered post-mortems.

Ticket AL-26 resolves these limitations end-to-end. It introduces pure completion date filtering in the Go functional core with deterministic fallback semantics for uncompleted or legacy workspaces, equips the HTTP API adapters with window, since, and until query parameters, introduces interactive date preset and custom range filtering on Fleet Overview defaulting to Project Lifetime, and adds an autopsy date sorting toggle (completed_at descending, nulls last) in the Ticket Deep Dive workspace selector.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architectural changes must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Domain Functional Core (backend/internal/domain/):
   - Pure filtering function FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger evaluates effective completion dates against inclusive temporal boundaries.
   - Primary completion date is derived from completed_at, with deterministic fallback to updated_at, and secondary fallback to started_at when completed_at is null or unparseable.
   - Pure sorting function SortLedgers is extended with SortFieldCompletedAt to order slices of AgenticMetricsLedger by completion date descending with nulls placed last.
   - Absolute isolation: zero imports of net/http, os, io, or outer packages. No non-deterministic clock calls (such as time.Now) inside domain core functions; reference timestamps must be passed as explicit arguments.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Boundary HTTP handlers FleetHandler and WorkspacesHandler in backend/internal/adapters/http/handlers.go parse and validate query parameters (window, since, until, sort, order).
   - Adapters handle date format deserialization (RFC 3339, RFC 3339 Nano, ISO 8601 YYYY-MM-DD), time window conversions relative to current wall clock, and parameter error formatting before delegating to pure domain filtering.
3. TypeScript Presentation and Formatters (web/src/domain/ and web/src/):
   - Pure comparator compareTicketCompletionDate in web/src/domain/formatters.ts mirrors the Go domain algorithm with identical fallback hierarchy and nulls-last positioning.
   - Pure client-side filtering utility in web/src/domain/formatters.ts provides consistent fallback filtering for offline mock state.
   - web/src/views/FleetOverview.tsx manages date filter state (preset selection and custom date pickers), triggering reactive data reloads.
   - web/src/components/TicketSelector.tsx provides an interactive sorting toggle between natural ticket ID sequence and completion date.

### Strong Typing and Zero Any Invariant
1. Zero usage of any or interface{} across Go backend and TypeScript frontend.
2. All date window presets and sort fields are strictly modeled via concrete string literal unions and Go enum types (TimeWindowPreset, LedgerSortField).
3. Value semantics: filtering and sorting functions in Go and TypeScript return newly allocated slices without mutating input arguments.

### Zero Emojis Mandate
No emojis are permitted in code, comments, docstrings, commit messages, PR drafts, UI labels, select options, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Functional Core (backend/internal/domain/):
   - Add SortFieldCompletedAt to LedgerSortField enum in backend/internal/domain/enums.go.
   - Update ParseLedgerSortField to map string "completed_at" to SortFieldCompletedAt.
   - Implement GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool) in backend/internal/domain/aggregations.go resolving completed_at with fallback to updated_at and started_at.
   - Implement FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger in backend/internal/domain/aggregations.go.
   - Extend SortLedgers in backend/internal/domain/aggregations.go to support SortFieldCompletedAt with descending order by default, placing null completion dates last, tie-breaking by natural ticket ID.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Update FleetHandler in backend/internal/adapters/http/handlers.go to parse optional window, since, and until query parameters, validate temporal boundaries, filter ledgers via domain.FilterLedgersByCompletionDate, and compute domain.CalculateFleetSummary.
   - Update WorkspacesHandler in backend/internal/adapters/http/handlers.go to support window, since, and until parameters alongside existing sort and order parameters.
   - Return structured HTTP 400 Bad Request error when unparseable date formats are supplied for since or until.
3. TypeScript API Client and Types (web/src/api/ and web/src/domain/):
   - Extend fetchFleetSummary, fetchWorkspaces, and fetchWorkspaceLedgers in web/src/api/client.ts to accept date filtering parameters (window, since, until).
   - Implement pure helper compareTicketCompletionDate in web/src/domain/formatters.ts.
   - Update TicketSelectorOption in web/src/components/TicketSelector.tsx to include completed_at, updated_at, and started_at.
4. TypeScript Presentation Views and Components (web/src/):
   - Add date range filter dropdown to web/src/views/FleetOverview.tsx with presets: Project Lifetime (default), Last 7 Days, Last 30 Days, Last 90 Days, and Custom Range.
   - Render start date and end date picker controls when Custom Range is selected.
   - Update web/src/components/TicketSelector.tsx with a sort mode toggle supporting Natural Ticket ID (PREFIX-NN, default) and Completion Date (completed_at descending, nulls last).
   - Ensure reactive data fetching and state synchronization across Fleet Overview KPI cards, distribution cards, and ticket directory tables.
5. Quality and Adversarial Testing:
   - Go unit tests in backend/internal/domain/aggregations_test.go covering boundary timestamps, null fallbacks, leap years, timezone offsets, and nulls-last sorting.
   - Go HTTP adapter tests in backend/internal/adapters/http/server_test.go verifying query parameters and error responses.
   - TypeScript unit and component tests in web/src/domain/formatters.test.ts, web/src/views/FleetOverview.test.tsx, and web/src/components/TicketSelector.test.tsx.

### Out of Scope
1. Altering OpenCode SQLite database schemas or raw telemetry collection logic.
2. Modifying ticket type classification or workspace prefix filtering (governed by AL-28 and AL-29).
3. Metric progression custom window hourly granularity derivations (governed by AL-27).
4. Permanent deletion or on-disk mutation of existing workspace ledgers.

---

## 4. Architectural Contracts and Data Specifications

### Date Resolution Hierarchy
When filtering or sorting by completion date, the effective timestamp for an AgenticMetricsLedger is evaluated in the following strict precedence order:
1. CompletedAt (completed_at): If present and successfully parsed as an ISO 8601 / RFC 3339 string, use this timestamp.
2. UpdatedAt (updated_at): If CompletedAt is null, empty, or unparseable, attempt to parse UpdatedAt.
3. StartedAt (started_at): If both CompletedAt and UpdatedAt are null or unparseable, attempt to parse StartedAt.
4. If all three fields are missing or unparseable, the ledger has no valid temporal timestamp.

### Inclusive Boundary Invariant
For any date range defined by since and until:
1. since: Represents the inclusive start timestamp. If provided as a calendar date (YYYY-MM-DD), it evaluates to 00:00:00.000000000 UTC on that date.
2. until: Represents the inclusive end timestamp. If provided as a calendar date (YYYY-MM-DD), it evaluates to 23:59:59.999999999 UTC on that date.
3. Ledgers matching since <= effective_date <= until are retained.
4. Ledgers with no valid temporal timestamp are excluded when date boundaries are active.
5. If both since and until are omitted or nil, all ledgers are retained (Project Lifetime).

### Completion Date Sorting Contract
1. Field identifier: completed_at.
2. Sort order: Descending by default (latest completed first).
3. Nulls last rule: Ledgers lacking a resolvable date sort after all ledgers with valid dates, regardless of whether ascending or descending sort is requested.
4. Secondary tie-breaker: When two ledgers share identical effective completion dates, order is resolved deterministically by natural ticket ID comparison via CompareTicketIDs.

### API Endpoint Query Parameter Contracts
1. GET /api/fleet:
   - window: Optional string (all, 7d, 30d, 90d). Defaults to all.
   - since: Optional ISO date/datetime string.
   - until: Optional ISO date/datetime string.
   - Status 200: Returns FleetSummary calculated across filtered ledgers.
   - Status 400: Returns structured ErrorResponse if since or until is invalid.
2. GET /api/workspaces:
   - window: Optional string (all, 7d, 30d, 90d). Defaults to all.
   - since: Optional ISO date/datetime string.
   - until: Optional ISO date/datetime string.
   - sort: Optional string (completed_at, ticket_id, updated_at, cost, duration, invocations).
   - order: Optional string (asc, desc).
   - Status 200: Returns slice of AgenticMetricsLedger filtered and sorted.
   - Status 400: Returns structured ErrorResponse if query parameters are invalid.
