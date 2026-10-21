# Contract: AL-23 Natural Ticket Identifier Ordering and Sortable Tables

## 1. Problem Statement

Across the AgenticLens observability platform, ticket workspaces are identified by structured codes consisting of an alphabetic prefix, an optional delimiter, and a numeric sequence number (such as AL-1, AL-10, GEN-42, PO-5).

Sorting ticket identifiers currently relies on standard lexicographical string comparison (sort.Strings in Go and localeCompare in TypeScript). This produces distorted ordering sequences:
1. Low-numbered tickets are scattered and displaced by double-digit tickets, producing sequences such as AL-1, AL-10, AL-11, AL-2, AL-20, AL-21, AL-22, AL-3.
2. In the Fleet Overview, Ticket Deep Dive dropdown, Ticket Comparisons matrix, and the CLI workspace scanner, users experience broken ordering where logical ticket progression is jumbled.
3. On the Landing Page, the Quick-Launch Workspaces directory table is completely static. Users cannot click column headers to reorder by ticket identifier, title, health status, iteration count, or active compute duration.

Ticket AL-23 resolves this architectural defect across the entire stack. It introduces pure natural ticket identifier comparison logic in the Go domain core and TypeScript domain formatters, integrates natural ordering as the default sequence across directory scanning and API handlers, and provides interactive, multi-column sorting with visible indicators for the Landing Page Quick-Launch Workspaces table.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architectural changes must strictly adhere to TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Domain Core (backend/internal/domain/):
   - Pure comparator function CompareTicketIDs(a, b string) int parses ticket prefixes and numerical suffixes to return a deterministic three-way integer comparison (-1, 0, 1).
   - Domain sorting function SortLedgers is extended with SortFieldTicketID to order slices of AgenticMetricsLedger.
   - The domain core maintains absolute isolation: zero imports of os, io, net/http, database/sql, or external sorting packages.
2. Go Imperative Boundary Adapters (backend/internal/adapters/):
   - adapters.ScanWorkspaces in backend/internal/adapters/repository.go wires domain.SortLedgers with SortFieldTicketID to ensure all returned scan reports present ledgers in natural order.
   - HTTP handler WorkspacesHandler in backend/internal/adapters/http/handlers.go parses query parameter sort=ticket_id and delegates to domain.SortLedgers.
3. TypeScript Presentation & Pure Formatters (web/src/domain/ and web/src/components/):
   - web/src/domain/formatters.ts houses pure comparator compareTicketIds(a: string, b: string): number mirroring the Go domain algorithm.
   - web/src/domain/formatters.ts updates deriveTicketComparisonRows to tie-break and sort by ticket_id using natural ordering.
   - web/src/views/LandingPage.tsx and web/src/components/TicketTable.tsx manage UI state and user click events, delegating ordering logic to pure domain comparators.
   - web/src/components/TicketSelector.tsx guarantees natural ordering of dropdown options by default.

### Strong Typing and Zero Any Invariant
1. Zero usage of any or interface{} in Go or TypeScript.
2. All sort fields, directions, and row models are strictly typed via TypeScript string literal unions and Go enum types (LedgerSortField).
3. Value semantics: sorting functions in Go and TypeScript return newly allocated slices or arrays without mutating inputs.

### Zero Emojis Mandate
No emojis are permitted in code, comments, docstrings, commit messages, PR drafts, UI labels, table headers, sort indicators, or workspace documentation.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Functional Core (backend/internal/domain/):
   - Add SortFieldTicketID to LedgerSortField enum in backend/internal/domain/enums.go.
   - Update ParseLedgerSortField to parse ticket_id into SortFieldTicketID.
   - Implement pure function CompareTicketIDs(a, b string) int in backend/internal/domain/aggregations.go supporting standard prefixes (AL-1, AL1, GEN-42, PO-5, AL-01) and fallback comparisons.
   - Extend SortLedgers in backend/internal/domain/aggregations.go to handle SortFieldTicketID ascending and descending, using UpdatedAt as deterministic secondary tie-breaker.
2. Go Imperative Adapters (backend/internal/adapters/):
   - Update adapters.ScanWorkspaces in backend/internal/adapters/repository.go to sort ledgers naturally via SortFieldTicketID before returning domain.ScanReport.
   - Validate and ensure backend/internal/adapters/http/handlers.go WorkspacesHandler supports sort=ticket_id.
3. TypeScript Domain Formatters (web/src/domain/):
   - Implement pure natural comparator compareTicketIds(a: string, b: string): number in web/src/domain/formatters.ts.
   - Update deriveTicketComparisonRows in web/src/domain/formatters.ts to use compareTicketIds for explicit ticket_id sorting and tie-breaking.
4. TypeScript Presentation Views and Components (web/src/):
   - Upgrade web/src/views/LandingPage.tsx Quick-Launch Workspaces table with interactive column sorting across Ticket ID, Title, Status, Iterations, and Active Duration, displaying visible asc and desc indicators.
   - Refactor web/src/components/TicketTable.tsx to use compareTicketIds when sorting by ticket_id.
   - Refactor web/src/components/TicketSelector.tsx to present ticket options in natural identifier order by default.
   - Update web/src/index.css with styles for sortable table headers in the Quick-Launch table (.quick-launch-table th.sortable).
5. Comprehensive Test Suites:
   - Go table-driven unit tests for CompareTicketIDs and SortLedgers in backend/internal/domain/aggregations_test.go.
   - Go integration and adapter tests in backend/internal/adapters/repository_test.go and backend/internal/adapters/http/server_test.go.
   - TypeScript unit tests in web/src/domain/formatters.test.ts, web/src/views/LandingPage.test.tsx, web/src/components/TicketTable.test.tsx, and web/src/components/TicketSelector.test.tsx (or Components.test.tsx).

### Out of Scope
1. Changes to database extraction logic or SQLite queries in backend/internal/adapters/opencode/.
2. Modifying ticket directory filesystem paths or disk folder naming schemes.
3. Modifying git commit diff parsing or code volume metrics.
4. Multi-dimensional filtering by workspace prefix (governed by AL-28).

---

## 4. Architectural Contracts and Data Specifications

### Natural Ticket Identifier Parsing Rules
A ticket identifier string is partitioned into three logical segments:
1. Prefix Segment: Leading alphabetic characters (case-insensitive ASCII letters).
2. Delimiter: Optional separator character (hyphen or underscore).
3. Sequence Number Segment: Consecutive ASCII numeric digits parsed as a non-negative integer.
4. Suffix Segment: Any remaining trailing characters.

### Comparison Algorithm Specification
Given inputs a and b:
1. Whitespace normalization: Trim leading and trailing whitespace from both inputs.
2. Identity check: If trimmed strings are equal, return 0.
3. Empty check: If either string is empty, empty sorts before non-empty.
4. Pattern extraction:
   - Parse prefix, numeric value, and raw numeric string from a and b.
   - If both inputs possess valid prefixes and numbers:
     - Compare prefixes alphabetically (case-insensitive). If unequal, return -1 if a is less than b, else 1.
     - If prefixes match, compare numerical values. If unequal, return -1 if numeric value of a is less than b, else 1.
     - If numerical values are equal (such as AL-1 versus AL-01): compare raw numeric strings or fall back to standard string comparison to ensure strict deterministic ordering.
   - If only one input conforms to the ticket pattern: conforming ticket identifier sorts before non-conforming string.
   - If neither input conforms to the ticket pattern: fall back to standard lexicographical string comparison.

### Go Domain Contracts (backend/internal/domain/enums.go & aggregations.go)
- LedgerSortField: string type alias.
  - SortFieldUpdatedAt = "updated_at"
  - SortFieldCost = "cost"
  - SortFieldDuration = "duration"
  - SortFieldInvocations = "invocations"
  - SortFieldTicketID = "ticket_id"
- ParseLedgerSortField(val string) LedgerSortField: parses string into enum, mapping "ticket_id" to SortFieldTicketID and unknown strings to SortFieldUpdatedAt.
- CompareTicketIDs(a, b string) int: pure function returning -1 when a < b, 0 when a == b, and 1 when a > b.
- SortLedgers(ledgers []AgenticMetricsLedger, field LedgerSortField, ascending bool) []AgenticMetricsLedger: pure function returning a newly allocated, sorted slice.

### TypeScript Presentation Contracts (web/src/domain/formatters.ts)
- compareTicketIds(a: string, b: string): number: pure function returning negative number when a < b, 0 when a === b, and positive number when a > b.
- deriveTicketComparisonRows(ledgers, sortField, sortOrder): uses compareTicketIds when sortField is 'ticket_id' and as secondary tie-breaker.

### Quick-Launch Table Sortable Columns Contract (web/src/views/LandingPage.tsx)
- Sortable Column Identifiers:
  - ticket_id: sorted via compareTicketIds.
  - ticket_title: sorted via string comparison localeCompare.
  - context_health_status: sorted via string comparison localeCompare.
  - sdlc_iterations: sorted via numeric difference.
  - active_duration_seconds: sorted via numeric difference.
- Non-sortable Columns:
  - Action button column.
- Sort Indicator Display:
  - Active ascending header displays text label " (asc)".
  - Active descending header displays text label " (desc)".
  - Inactive headers display no indicator.
