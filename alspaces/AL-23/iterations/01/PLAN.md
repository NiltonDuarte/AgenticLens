# Implementation Plan: AL-23 Natural Ticket Identifier Ordering and Sortable Tables

## 1. Goal and Iteration Scope

Deliver complete functional and architectural implementation of ticket AL-23 across the Go backend and TypeScript frontend:
- Implement pure natural ticket identifier comparison in the Go functional core.
- Extend Go domain sorting to support SortFieldTicketID.
- Wire natural ordering into the Go workspace filesystem scanner and REST API handler.
- Implement pure natural ticket comparison in the TypeScript domain layer.
- Add interactive column sorting with visual direction indicators to the Landing Page Quick-Launch Workspaces table.
- Update Fleet Overview ticket table, ticket autopsy dropdown, and comparison row tie-breaking to enforce natural order.
- Provide comprehensive table-driven adversarial unit and integration tests across Go and TypeScript.

---

## 2. Grounded Reference Architecture & File Paths

The implementation grounds against concrete existing paths and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/enums.go: Add SortFieldTicketID constant to LedgerSortField enum; update ParseLedgerSortField to recognise "ticket_id".
- backend/internal/domain/aggregations.go: Implement pure function CompareTicketIDs(a, b string) int; update SortLedgers to handle SortFieldTicketID.
- backend/internal/domain/aggregations_test.go: Table-driven unit tests for CompareTicketIDs covering standard formatting, missing delimiters, mixed prefixes, zero-padding, empty values, and non-conforming inputs; unit tests for SortLedgers with SortFieldTicketID.

### Go Imperative Boundary Shell (backend/internal/adapters/)
- backend/internal/adapters/repository.go: In ScanWorkspaces, sort collected ledgers with domain.SortLedgers(ledgers, domain.SortFieldTicketID, true) before returning ScanReport.
- backend/internal/adapters/http/handlers.go: In WorkspacesHandler, ensure query parameter sort=ticket_id delegates to domain.SortLedgers.
- backend/internal/adapters/repository_test.go: Verify ScanWorkspaces returns workspaces in natural ticket order (e.g. AL-1, AL-2, AL-10, AL-22).
- backend/internal/adapters/http/server_test.go: Verify GET /api/workspaces?sort=ticket_id&order=asc and order=desc return correctly ordered ledgers.

### TypeScript Domain Layer (web/src/domain/)
- web/src/domain/formatters.ts: Implement pure function compareTicketIds(a: string, b: string): number; update deriveTicketComparisonRows to use compareTicketIds for ticket_id sorting and tie-breaking.
- web/src/domain/formatters.test.ts: Unit tests for compareTicketIds; update deriveTicketComparisonRows assertions to verify natural ordering.

### TypeScript Presentation Layer (web/src/components/ & web/src/views/)
- web/src/views/LandingPage.tsx: Add sorting state (sortField and sortAscending), header click handlers, visual indicators, and row sorting memoization to the Quick-Launch Workspaces table.
- web/src/views/LandingPage.test.tsx: Test suite verifying column header clicks, sort direction toggling, and natural ticket identifier ordering on the landing page.
- web/src/components/TicketTable.tsx: Update ticket_id sort case to use compareTicketIds.
- web/src/components/TicketTable.test.tsx: Add and update tests verifying natural ticket ordering (AL-2 before AL-10) for default sort and ticket_id header clicks.
- web/src/components/TicketSelector.tsx: Wrap tickets in memoized natural sort using compareTicketIds so autopsy dropdown options are always ordered naturally.
- web/src/components/TicketSelector.test.tsx (or web/src/components/Components.test.tsx): Verify that TicketSelector renders options in natural order regardless of input slice order.
- web/src/index.css: Add cursor: pointer and hover rules for .quick-launch-table th.sortable.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Functional Core Implementation
- Task 1.1: In backend/internal/domain/enums.go:
  - Add SortFieldTicketID LedgerSortField = "ticket_id" to the const block.
  - In ParseLedgerSortField, add case "ticket_id": return SortFieldTicketID.
- Task 1.2: In backend/internal/domain/aggregations.go:
  - Implement pure comparator CompareTicketIDs(a, b string) int:
    - Trim spaces from inputs.
    - If a == b, return 0.
    - Handle empty string edge cases: empty strings sort before non-empty strings.
    - Parse prefix (letters), optional delimiter (- or _), and numeric sequence digits using regular expressions or rune scanning.
    - If both match, compare prefixes using case-insensitive comparison (strings.ToUpper). If prefixes differ, return comparison result.
    - If prefixes match, convert numeric digit strings to int64. If numbers differ, return -1 if a < b else 1.
    - If numbers are equal (e.g. AL-1 vs AL-01), compare original raw string lengths or contents as tie-breaker.
    - If only one input conforms to the ticket pattern, conforming ticket sorts first.
    - If neither input conforms, fall back to strings.Compare(a, b).
  - In SortLedgers, add case SortFieldTicketID:
    - Safely extract ticket ID strings from a.TicketID and b.TicketID (defaulting nil to empty string).
    - Call CompareTicketIDs.
    - If comparison returns 0, tie-break using a.UpdatedAt < b.UpdatedAt.
    - Set less variable based on comparison result.

### Phase 2: Go Boundary Shell Adapter Updates
- Task 2.1: In backend/internal/adapters/repository.go:
  - In ScanWorkspaces, after all ledgers are parsed and constructed from metric files, sort ledgers using domain.SortLedgers(ledgers, domain.SortFieldTicketID, true).
  - Return ScanReport with naturally ordered Ledgers slice.
- Task 2.2: In backend/internal/adapters/http/handlers.go:
  - In WorkspacesHandler, verify that sortParam and orderParam processing handles SortFieldTicketID without regressions.

### Phase 3: TypeScript Domain Formatter Implementation
- Task 3.1: In web/src/domain/formatters.ts:
  - Implement export function compareTicketIds(a: string, b: string): number:
    - Trim spaces.
    - If a === b, return 0.
    - Handle empty strings: empty strings sort before non-empty strings.
    - Use regular expression to extract alphabetic prefix, numeric string, and trailing characters.
    - Compare prefixes case-insensitively using localeCompare. If non-zero, return result.
    - Parse numeric strings with parseInt(..., 10). If numeric values differ, return numA - numB.
    - If numeric values match, tie-break with a.localeCompare(b).
    - If only one conforms to ticket pattern, conforming string sorts before non-conforming.
    - If neither conforms, return a.localeCompare(b).
  - In deriveTicketComparisonRows:
    - In case 'ticket_id': replace a.ticket_id.localeCompare(b.ticket_id) with compareTicketIds(a.ticket_id, b.ticket_id).
    - In the default equality tie-breaker (line 243): replace a.ticket_id.localeCompare(b.ticket_id) with compareTicketIds(a.ticket_id, b.ticket_id).
    - In the unselected sortField block (line 179): replace b.ticket_id.localeCompare(a.ticket_id) with compareTicketIds(b.ticket_id, a.ticket_id).

### Phase 4: TypeScript Presentation and Component Enhancements
- Task 4.1: In web/src/components/TicketTable.tsx:
  - In the sort switch statement, under case 'ticket_id', replace a.ticket_id.localeCompare(b.ticket_id) with compareTicketIds(a.ticket_id, b.ticket_id).
- Task 4.2: In web/src/components/TicketSelector.tsx:
  - Import compareTicketIds from ../domain/formatters.
  - Memoize tickets array sorting: compute sortedTickets by sorting tickets with compareTicketIds(a.ticket_id, b.ticket_id).
  - Map over sortedTickets when rendering option elements.
- Task 4.3: In web/src/views/LandingPage.tsx:
  - Define type QuickLaunchSortField as union of 'ticket_id', 'ticket_title', 'context_health_status', 'sdlc_iterations', 'active_duration_seconds'.
  - Define sortField state initialized to 'ticket_id'.
  - Define sortAscending state initialized to true.
  - Implement handleSort(field: QuickLaunchSortField): if current sortField equals field, toggle sortAscending; otherwise set sortField to field and sortAscending to true.
  - Implement getSortIndicator(field: QuickLaunchSortField): returns ' (asc)' if active ascending, ' (desc)' if active descending, or empty string if inactive.
  - Memoize sortedWorkspaces array:
    - For 'ticket_id': compare with compareTicketIds.
    - For 'ticket_title': compare with a.ticket_title.localeCompare(b.ticket_title).
    - For 'context_health_status': compare with a.context_health_status.localeCompare(b.context_health_status).
    - For 'sdlc_iterations': compare with a.sdlc_iterations - b.sdlc_iterations.
    - For 'active_duration_seconds': compare with a.active_duration_seconds - b.active_duration_seconds.
    - Apply sortAscending sign multiplier.
  - In the table header, mark sortable headers with className="sortable", attach onClick={() => handleSort(...)}, and append getSortIndicator(...).
  - Render sortedWorkspaces in the table body.
- Task 4.4: In web/src/index.css:
  - Add styles for .quick-launch-table th.sortable: set cursor: pointer and user-select: none.
  - Add styles for .quick-launch-table th.sortable:hover: set color: #ffffff.

---

## 5. Adversarial Testing Strategy for Tester

### Pure Go Domain Tests (backend/internal/domain/aggregations_test.go)
- Table-driven TestCompareTicketIDs:
  - Identical prefix natural progression: verify AL-1 < AL-2 < AL-9 < AL-10 < AL-11 < AL-22.
  - Prefix without hyphen: verify AL1 < AL2 < AL10.
  - Multiple distinct prefixes: verify AL-10 < GEN-1 < GEN-2 < PO-1.
  - Leading zeroes tie-breaking: verify AL-1 and AL-01 do not produce inconsistent or cyclic results.
  - Non-conforming and partial formats: verify AL- < AL-1, ticket-xyz vs AL-1, UNKNOWN vs AL-2.
  - Empty strings: verify "" < "AL-1" and "" == "".
  - Case insensitivity: verify al-2 < AL-10 and AL-2 < al-10.
- TestSortLedgers_TicketID:
  - Create ledgers with ticket IDs: AL-10, AL-2, AL-1, GEN-5, AL-22, PO-1.
  - Sort ascending: assert order AL-1, AL-2, AL-10, AL-22, GEN-5, PO-1.
  - Sort descending: assert order PO-1, GEN-5, AL-22, AL-10, AL-2, AL-1.
  - Immutability check: verify input slice remains unmodified.

### Go Adapter Integration Tests
- backend/internal/adapters/repository_test.go:
  - In TestScanWorkspaces, create workspaces out of order (AL-10, AL-2, AL-1, GEN-1).
  - Call adapters.ScanWorkspaces; assert returned Ledgers slice is naturally ordered.
- backend/internal/adapters/http/server_test.go:
  - Call GET /api/workspaces?sort=ticket_id&order=asc; assert status 200 and natural ascending ticket order.
  - Call GET /api/workspaces?sort=ticket_id&order=desc; assert status 200 and natural descending ticket order.

### Pure TypeScript Formatter Tests (web/src/domain/formatters.test.ts)
- describe('compareTicketIds'):
  - Assert compareTicketIds('AL-2', 'AL-10') < 0.
  - Assert compareTicketIds('AL-10', 'AL-2') > 0.
  - Assert compareTicketIds('AL-1', 'AL-1') === 0.
  - Assert compareTicketIds('AL-1', 'AL-01') !== 0 (deterministic tie-breaking).
  - Assert multi-prefix ordering: AL-10 before GEN-1 before PO-1.
  - Assert format without hyphen: AL1 before AL10.
  - Assert case insensitivity: al-2 before AL-10.
- describe('deriveTicketComparisonRows natural ordering'):
  - Verify sort by 'ticket_id' ascending produces AL-1 before AL-2 before AL-10.

### TypeScript Component & View Tests
- web/src/components/TicketTable.test.tsx:
  - Test default sorting order with sample rows (AL-10, AL-2, GEN-9): assert first row contains AL-2 and second row contains AL-10.
  - Test clicking Ticket ID header: toggles from ascending (AL-2, AL-10, GEN-9) to descending (GEN-9, AL-10, AL-2).
- web/src/components/TicketSelector.test.tsx:
  - Pass unsorted tickets (AL-10, AL-2, GEN-1, AL-1) to TicketSelector.
  - Inspect rendered option elements and assert options appear in order: AL-1, AL-2, AL-10, GEN-1.
- web/src/views/LandingPage.test.tsx:
  - Verify Quick-Launch table headers are rendered with sortable capability.
  - Verify initial render displays workspaces naturally sorted by Ticket ID with " (asc)" indicator on Ticket ID header.
  - Click Ticket ID header: assert sort indicator becomes " (desc)" and order inverts.
  - Click Iterations header: assert sort indicator shifts to Iterations header and rows sort by iteration count.
  - Click Active Duration header: assert rows sort by active duration.

---

## 6. Verification and Quality Gates

Execute full validation sequence before submitting iteration deliverables:
- Backend static analysis: go vet ./...
- Backend unit and race test suite: go test -race ./...
- Frontend typechecking: npm run typecheck
- Frontend linting: npm run lint
- Frontend unit tests: npm test
- Automated build gate: make check && make test && make build
- Assert zero any across Go and TypeScript files.
- Assert zero emojis across code, tests, UI strings, and artifacts.
