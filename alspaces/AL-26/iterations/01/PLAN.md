# Implementation Plan: AL-26 Fleet Overview Completion Date Filtering and Autopsy Date Sorting

## 1. Goal and Iteration Scope

Deliver complete implementation of ticket AL-26 across the Go functional core, imperative boundary adapters, and TypeScript frontend views:
- Implement pure domain date filtering FilterLedgersByCompletionDate in Go with timestamp fallback hierarchy.
- Add completion date sorting SortFieldCompletedAt with nulls-last positioning and natural ticket ID tie-breaking.
- Extend /api/fleet and /api/workspaces HTTP handlers to accept window, since, and until query parameters.
- Add date range filtering dropdown and custom range picker to Fleet Overview, defaulting to Project Lifetime.
- Update Ticket Deep Dive Workspace Autopsy selector with a sorting toggle between natural ticket ID and completion date.
- Guarantee zero any, zero emojis, and robust test coverage across Go and TypeScript.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against concrete existing files and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/enums.go: Add SortFieldCompletedAt constant to LedgerSortField; update ParseLedgerSortField to recognize "completed_at".
- backend/internal/domain/aggregations.go:
  - Implement GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool).
  - Implement FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger.
  - Extend SortLedgers to handle SortFieldCompletedAt descending by default, placing nulls last, tie-breaking via CompareTicketIDs.
- backend/internal/domain/aggregations_test.go: Adversarial unit tests for date filtering, inclusive bounds, leap years, timezone offsets, and nulls-last sorting.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/http/handlers.go:
  - Add date query parameter parsing helper for window, since, and until.
  - Update FleetHandler to apply date filtering before invoking CalculateFleetSummary.
  - Update WorkspacesHandler to apply date filtering and support sort=completed_at.
- backend/internal/adapters/http/server_test.go: Integration tests verifying /api/fleet and /api/workspaces with date filter parameters, sorting flags, and error responses on invalid dates.

### TypeScript Domain & API (web/src/domain/ & web/src/api/)
- web/src/domain/types.ts: Update ComparisonSortField to ensure completed_at is supported; add AutopsySortMode type definition.
- web/src/domain/formatters.ts:
  - Implement pure helper compareTicketCompletionDate.
  - Implement pure helper filterLedgersByCompletionDate for client-side and mock state filtering.
- web/src/domain/formatters.test.ts: Unit tests verifying compareTicketCompletionDate descending sort, fallback hierarchy, nulls-last behavior, and natural ID tie-breaking.
- web/src/api/client.ts: Update fetchFleetSummary, fetchWorkspaces, and fetchWorkspaceLedgers to pass window, since, and until query parameters.
- web/src/api/client.test.ts: Verify query string generation for date filter parameters.

### TypeScript Presentation (web/src/views/ & web/src/components/)
- web/src/views/FleetOverview.tsx:
  - Add date range filter state (windowPreset, customSince, customUntil).
  - Render date filter control bar above KPI cards with preset dropdown (Project Lifetime, Last 7 Days, Last 30 Days, Last 90 Days, Custom Range).
  - Render start and end date pickers when Custom Range is active.
  - Wire reactive reload on filter change.
- web/src/views/FleetOverview.test.tsx: Component tests verifying default Project Lifetime render, preset changes, custom date inputs, and zero emojis.
- web/src/components/TicketSelector.tsx:
  - Extend TicketSelectorOption interface with completed_at, updated_at, started_at.
  - Add sort mode toggle between Natural Ticket ID and Completion Date.
  - Apply compareTicketCompletionDate when completion date sort is active.
- web/src/components/TicketSelector.test.tsx: Verify options ordering for both sort modes.
- web/src/views/TicketDeepDive.tsx: Pass date fields (completed_at, updated_at, started_at) into ticketOptions.
- web/src/views/TicketDeepDive.test.tsx: Test autopsy selector sorting interaction.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Implementation
- Task 1.1: In backend/internal/domain/enums.go:
  - Define SortFieldCompletedAt LedgerSortField = "completed_at".
  - Update ParseLedgerSortField(val string) LedgerSortField to return SortFieldCompletedAt when val == "completed_at".
- Task 1.2: In backend/internal/domain/aggregations.go:
  - Implement timestamp parsing utility supporting RFC 3339 Nano, RFC 3339, and date-only YYYY-MM-DD formats in UTC.
  - Implement pure function GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool):
    - Attempt to parse ledger.CompletedAt. If valid, return parsed timestamp and true.
    - If ledger.CompletedAt is nil or unparseable, attempt to parse ledger.UpdatedAt. If valid, return timestamp and true.
    - If ledger.UpdatedAt is empty or unparseable, attempt to parse ledger.StartedAt. If valid, return timestamp and true.
    - If none valid, return zero time.Time and false.
  - Implement pure function FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger:
    - If since == nil and until == nil, return a newly allocated copy of ledgers.
    - Iterate through ledgers, resolving effective completion date with GetLedgerCompletionDate.
    - Exclude ledgers without a valid date.
    - Exclude ledgers where effective date is before *since (when since != nil).
    - Exclude ledgers where effective date is after *until (when until != nil).
    - Return filtered slice.
  - Extend SortLedgers with case SortFieldCompletedAt:
    - Extract dates for candidate pair a and b via GetLedgerCompletionDate.
    - Enforce nulls-last invariant: if only one candidate has a date, the candidate with a date precedes the null candidate regardless of sort direction.
    - If both candidates have dates, compare chronological timestamps (descending by default).
    - If timestamps match, tie-break using CompareTicketIDs(tidA, tidB).

### Phase 2: Go Boundary Shell HTTP Handlers
- Task 2.1: In backend/internal/adapters/http/handlers.go:
  - Implement helper to parse since and until query parameters into *time.Time:
    - Parse date-only string (YYYY-MM-DD) as start of day 00:00:00 UTC for since and end of day 23:59:59.999999999 UTC for until.
    - Parse full ISO timestamp as exact UTC timestamp.
    - Return explicit error if formatting is invalid.
  - Implement helper to resolve window parameter (7d, 30d, 90d, all) relative to reference time into since and until pointers when explicit boundaries are omitted.
  - In FleetHandler:
    - Parse window, since, and until. Return HTTP 400 Bad Request via WriteError if parsing fails.
    - Apply domain.FilterLedgersByCompletionDate.
    - Compute fleet summary across filtered ledgers via domain.CalculateFleetSummary.
    - Return HTTP 200 with JSON payload.
  - In WorkspacesHandler:
    - Parse window, since, until, sort, and order. Return HTTP 400 on parse error.
    - Apply domain.FilterLedgersByCompletionDate.
    - Apply domain.SortLedgers with requested sort field and direction (defaulting to natural ticket ID if sort is unspecified).
    - Return HTTP 200 with filtered and sorted ledgers.

### Phase 3: TypeScript Domain & API Enhancements
- Task 3.1: In web/src/domain/types.ts:
  - Define export type AutopsySortMode = 'ticket_id' | 'completed_at';.
  - Ensure ComparisonSortField includes 'completed_at'.
- Task 3.2: In web/src/domain/formatters.ts:
  - Implement pure function compareTicketCompletionDate:
    - Evaluate effective completion date using completed_at -> updated_at -> started_at.
    - Enforce nulls-last ordering in descending sequence.
    - For candidates with matching dates, tie-break with compareTicketIds.
  - Implement pure client-side filter helper filterLedgersByCompletionDate.
- Task 3.3: In web/src/api/client.ts:
  - Update fetchFleetSummary(windowParam?: string, sinceParam?: string, untilParam?: string).
  - Update fetchWorkspaces(windowParam?: string, sortParam?: string, orderParam?: string, sinceParam?: string, untilParam?: string).
  - Update fetchWorkspaceLedgers(windowParam?: string, sinceParam?: string, untilParam?: string).
  - Build query string parameters dynamically.
  - Apply client-side fallback filtering in mock handlers to ensure mock mode behaves identically to production backend.

### Phase 4: TypeScript Presentation Layer
- Task 4.1: In web/src/views/FleetOverview.tsx:
  - Declare state variables: windowPreset (default 'all'), customSince (default ''), customUntil (default '').
  - Render a date filter control section above KPI cards:
    - Label: Date Range Filter:
    - Select dropdown with options: Project Lifetime (all), Last 7 Days (7d), Last 30 Days (30d), Last 90 Days (90d), Custom Range (custom).
    - When custom is selected, render date inputs for Since and Until.
  - In loadData, pass filter parameters to fetchFleetSummary and fetchWorkspaces.
  - Trigger loadData reactively whenever windowPreset, customSince, or customUntil updates.
  - Handle empty filtered results cleanly with the existing EmptyState component.
- Task 4.2: In web/src/components/TicketSelector.tsx:
  - Update TicketSelectorOption interface to include completed_at?: string | null; updated_at?: string; started_at?: string | null;.
  - Introduce sortMode state defaulting to 'ticket_id'.
  - Add an autopsy sort toggle control (select dropdown or toggle button) allowing users to switch between Sort: Ticket ID and Sort: Completion Date.
  - Memoize sortedTickets: sort using compareTicketIds when sortMode === 'ticket_id', or compareTicketCompletionDate when sortMode === 'completed_at'.
  - Render options using sortedTickets.
- Task 4.3: In web/src/views/TicketDeepDive.tsx:
  - Ensure ticketOptions extracts completed_at, updated_at, and started_at from each ledger in ledgers.

---

## 4. Adversarial Testing Strategy for Tester

### Pure Go Domain Tests (backend/internal/domain/aggregations_test.go)
- Table-driven TestFilterLedgersByCompletionDate:
  - Inclusive boundary assertions: verify ledgers timestamped exactly at since or until are included.
  - Fallback validation: verify ledger with null completed_at but matching updated_at is included.
  - Secondary fallback: verify ledger with null completed_at and null updated_at uses started_at.
  - All-null exclusion: verify ledgers without any parseable timestamps are excluded when date range is active.
  - Open-ended ranges: verify since != nil && until == nil and since == nil && until != nil.
  - Leap year handling: verify ledgers on February 29 (leap year 2024 or 2028) filter correctly.
  - Timezone offset handling: verify timestamps in UTC, EDT (-04:00), and JST (+09:00) resolve accurately to UTC.
  - Inverted bounds: verify since after until returns an empty slice without panic.
  - Empty input slice: verify empty slice returns empty slice without error.
- Table-driven TestSortLedgers_CompletedAt:
  - Verify descending sort orders recent completions before older completions.
  - Verify nulls-last positioning: ledgers with null completion dates appear at the end of the slice.
  - Verify tie-breaking: ledgers with identical completion dates are ordered by natural ticket ID.

### Go Boundary Adapter Tests (backend/internal/adapters/http/server_test.go)
- Test /api/fleet?window=7d: verify summary aggregates only ledgers within the last 7 days.
- Test /api/fleet?since=2026-10-01&until=2026-10-15: verify date range filtering.
- Test /api/workspaces?since=2026-10-01&until=2026-10-15: verify workspace list is filtered.
- Test /api/workspaces?sort=completed_at&order=desc: verify returned ledgers are ordered by completion date with nulls last.
- Test invalid date input /api/fleet?since=invalid-date: verify HTTP 400 Bad Request with structured error payload.

### TypeScript Unit & Formatter Tests (web/src/domain/formatters.test.ts)
- Unit test compareTicketCompletionDate:
  - Verify descending order: newer date before older date.
  - Verify fallback from null completed_at to updated_at.
  - Verify nulls-last rule: candidate without date sorts after candidate with date.
  - Verify natural ticket ID tie-breaking when dates are identical.

### TypeScript Component & View Tests
- web/src/views/FleetOverview.test.tsx:
  - Verify Fleet Overview defaults to Project Lifetime (all).
  - Verify selecting preset 7d triggers API client calls with window=7d.
  - Verify selecting Custom Range displays date input fields.
  - Verify changing custom dates triggers API client calls with since and until.
  - Verify zero emojis in rendered DOM.
- web/src/components/TicketSelector.test.tsx:
  - Verify selector defaults to Natural Ticket ID order (AL-1, AL-2, AL-10).
  - Verify toggling sort mode to Completion Date reorders options chronologically (completed_at descending) with nulls last.
  - Verify zero emojis in rendered DOM.
- web/src/views/TicketDeepDive.test.tsx:
  - Verify autopsy selector integration and ticket selection synchronization.

---

## 5. Verification and Quality Gates

Execute full validation sequence before submitting iteration deliverables:
- Backend static analysis: go vet ./...
- Backend unit and race test suite: go test -race ./...
- Frontend typechecking: npm run typecheck
- Frontend linting: npm run lint
- Frontend unit tests: npm test
- Single binary build verification: make check && make test && make build
- Assert zero any across all modified Go and TypeScript files.
- Assert zero emojis across code, tests, UI strings, and artifacts.
