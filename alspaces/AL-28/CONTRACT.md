# Contract: AL-28 Multi-Dimensional Filtering by Ticket Type and Workspace Prefix

## 1. Problem Statement

Across the AgenticLens observability platform, engineering fleets scale across diverse ticket initiatives. Workspaces record metrics from feature implementation, bug fixes, automated refactorings, chore automation, and documentation or Product Owner sessions. Currently, Fleet Overview, Ticket Comparisons, and Metric Progression aggregate and display metrics across all recorded workspaces indiscriminately. Users cannot isolate metrics to specific ticket categories or partition by initiative prefix.

Furthermore, HTTP endpoints (/api/fleet and /api/workspaces) lack filtering parameters for ticket type or workspace prefix, while dedicated cross-ticket comparison (/api/comparison) and longitudinal progression (/api/progression) endpoints are not yet surfaced in the HTTP adapter.

Based on steering feedback, both ticket type and workspace prefix dimensions must default to "all" (All Types, All Prefixes) and must support multi-selection simultaneously. Users must be able to select any arbitrary subset of ticket types (for example, simultaneously viewing features and bugfixes while excluding chores and documentation) and any subset of workspace prefixes (for example, combining "AL" and "GEN" while omitting "PO").

Ticket AL-28 resolves these limitations vertically across backend and frontend. It introduces pure multi-value ticket type and workspace prefix filtering in the Go functional core with case-insensitive alias resolution and sorted prefix extraction, equips the HTTP boundary shell with query parameter support accepting comma-separated and repeated parameters, registers dedicated comparison and progression routes, updates the TypeScript API client and domain formatters, and integrates reactive multi-select filter controls across Fleet Overview, Ticket Comparisons, and Metric Progression.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All architectural additions must strictly satisfy TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Go Domain Functional Core (backend/internal/domain/):
   - Pure filtering function FilterLedgersByTicketTypes evaluates candidate ledgers against a slice of queried ticket types with symmetric alias expansion (e.g. "feat" matches "feature", "fix" matches "bug" and "bugfix", "doc" matches "docs"). When ticketTypes is empty, or contains "all" or an empty string, all ledgers are retained.
   - Pure filtering function FilterLedgersByPrefixes extracts the workspace prefix from ledger.TicketID and matches against a slice of requested prefix strings using case-insensitive set membership. When prefixes is empty, or contains "all" or an empty string, all ledgers are retained.
   - Pure extraction function ExtractDistinctPrefixes gathers, deduplicates, and sorts ticket prefixes present across candidate ledgers in ascending alphabetical order.
   - Value semantics: domain filtering and extraction functions allocate and return new slices without mutating input arguments.
   - Absolute isolation: zero imports of net/http, os, io, or non-deterministic clock packages inside domain functions.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Boundary HTTP handlers in backend/internal/adapters/http/handlers.go parse query parameters (ticket_type, prefix, window, since, until, sort, order), supporting both comma-delimited strings (e.g. ticket_type=feature,bug) and repeated query keys (e.g. ticket_type=feature&ticket_type=bug).
   - Server route multiplexer in backend/internal/adapters/http/server.go registers GET /api/comparison and GET /api/progression.
   - Handlers delegate filtering to pure domain functions and serialize responses via typed helper functions.
3. TypeScript Presentation and Domain (web/src/domain/ and web/src/):
   - Pure client-side filtering functions in web/src/domain/formatters.ts replicate Go domain multi-selection logic and alias rules for offline mock resilience.
   - View components (FleetOverview.tsx, TicketComparisons.tsx, MetricProgression.tsx) manage multi-select filter state defaulting to "all", dispatch queries, and synchronize KPIs, charts, distributions, and tables in unison.

### Strong Typing and Zero Any Invariant
1. Zero usage of any or interface{} across Go backend and TypeScript frontend.
2. Concrete response types enforced via Go generics on ResponsePayload in backend/internal/adapters/http/dto.go.
3. Concrete string literal unions and typed arrays model multi-select filter options in TypeScript.

### Zero Emojis Mandate
No emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI labels, select options, or workspace artifacts.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Go Functional Core (backend/internal/domain/):
   - Implement CanonicalTicketType(raw string) string in backend/internal/domain/aggregations.go to normalize conventional aliases.
   - Implement FilterLedgersByTicketTypes(ledgers []AgenticMetricsLedger, ticketTypes []string) []AgenticMetricsLedger.
   - Implement single-value convenience wrapper FilterLedgersByTicketType(ledgers []AgenticMetricsLedger, ticketType string) []AgenticMetricsLedger.
   - Implement ExtractTicketPrefix(ticketID string) string.
   - Implement FilterLedgersByPrefixes(ledgers []AgenticMetricsLedger, prefixes []string) []AgenticMetricsLedger.
   - Implement single-value convenience wrapper FilterLedgersByPrefix(ledgers []AgenticMetricsLedger, prefix string) []AgenticMetricsLedger.
   - Implement ExtractDistinctPrefixes(ledgers []AgenticMetricsLedger) []string.
   - Ensure TicketComparisonRow in models.go models TicketType *string so comparison rows carry ticket type metadata.
   - Update BuildComparisonRows to propagate TicketType to TicketComparisonRow.
2. Go Imperative Boundary Shell (backend/internal/adapters/):
   - Add multi-parameter parsing helper parseMultiQueryParam in backend/internal/adapters/http/handlers.go supporting comma-separated tokens and repeated query parameters.
   - Update FleetHandler and WorkspacesHandler to parse multi-value ticket_type and prefix query parameters and apply domain multi-filters.
   - Implement ComparisonHandler in backend/internal/adapters/http/handlers.go to handle GET /api/comparison, filtering ledgers by temporal, ticket_type, and prefix parameters before building comparison rows via domain.BuildComparisonRows.
   - Implement ProgressionHandler in backend/internal/adapters/http/handlers.go to handle GET /api/progression, returning filtered []AgenticMetricsLedger.
   - Register routes in backend/internal/adapters/http/server.go.
   - Update ResponsePayload union in backend/internal/adapters/http/dto.go to include []domain.TicketComparisonRow and add WriteComparisonResponse.
3. TypeScript Domain & API (web/src/domain/ & web/src/api/):
   - Add StandardTicketType union and ensure TicketComparisonRow includes ticket_type in web/src/domain/types.ts.
   - Implement canonicalTicketType, extractTicketPrefix, filterLedgersByTicketTypes, filterLedgersByPrefixes, extractDistinctPrefixes, and extractDistinctTicketTypes in web/src/domain/formatters.ts.
   - Update fetchFleetSummary, fetchWorkspaceLedgers, and fetchWorkspaces in web/src/api/client.ts to accept multi-value ticketTypesParam and prefixesParam (accepting readonly string[] or comma-delimited strings).
   - Implement fetchComparisonRows and fetchProgressionLedgers in web/src/api/client.ts with mock fallbacks.
4. TypeScript Presentation Views & Components (web/src/views/ & web/src/components/):
   - Add MultiSelectDropdown component in web/src/components/MultiSelectDropdown.tsx (or inline controls) supporting multi-selection checkboxes, "All" toggle logic, and selection badges.
   - Add Ticket Type and Workspace Prefix multi-select controls to FleetOverview.tsx, defaulting to ["all"] and ["all"].
   - Add Ticket Type and Workspace Prefix multi-select controls to TicketComparisons.tsx, defaulting to ["all"] and ["all"].
   - Add Ticket Type and Workspace Prefix multi-select controls to MetricProgression.tsx, defaulting to ["all"] and ["all"].
   - Reactively reload or re-filter views on filter adjustments.
5. Quality and Verification:
   - Pure Go unit tests in backend/internal/domain/aggregations_test.go covering case insensitivity, multi-alias mapping, multi-prefix extraction, empty and "all" pass-through, non-conforming IDs, and distinct prefix sorting.
   - Go HTTP adapter tests in backend/internal/adapters/http/server_test.go covering comma-delimited and repeated query parameter filtering for /api/fleet, /api/workspaces, /api/comparison, and /api/progression.
   - TypeScript unit tests in web/src/domain/formatters.test.ts and web/src/api/client.test.ts.
   - TypeScript component integration tests in FleetOverview.test.tsx, TicketComparisons.test.tsx, and MetricProgression.test.tsx verifying multi-select toggles and zero emojis.

### Out of Scope
1. Altering OpenCode SQLite database schemas or ingestion logic.
2. Modifying telemetry generation CLI commands (governed by AL-29).
3. Metric progression hourly granularity derivations (governed by AL-27).
4. Mutating existing workspace ledger files on disk.

---

## 4. Architectural Contracts and Data Specifications

### Multi-Value Ticket Type Aliasing and Matching Contract
When matching ledgers by ticket type:
1. Normalization: All queried ticket types and candidate ledger ticket_type values are trimmed of whitespace and converted to lowercase.
2. Canonical Equivalence Mapping:
   - "feat" and "feature" map to canonical "feature".
   - "bug", "fix", and "bugfix" map to canonical "bug".
   - "doc" and "docs" map to canonical "docs".
   - Unaliased strings (e.g. "chore", "refactor", "ci", "test") map to their trimmed lowercase representation.
3. Multi-Match Evaluation:
   - If queried types slice is empty, or contains "all" or an empty string, all ledgers match (pass-through).
   - If queried types specify concrete types (e.g. ["feature", "bug"]), candidate ledgers whose canonical ticket type exists in the queried canonical set match.
   - Ledgers with null or empty ticket_type do not match specific category queries.

### Multi-Value Workspace Prefix Extraction and Matching Contract
When extracting and filtering by ticket prefix:
1. Extraction:
   - Evaluated against ledger.TicketID using the conforming pattern: leading letters, optional hyphen or underscore, followed by numeric digits.
   - Leading alphabetic characters are extracted and converted to uppercase (e.g. "al-1" yields "AL", "GEN_4" yields "GEN", "PO-2" yields "PO").
   - If a ticket identifier lacks digits but contains delimiters, the leading token is used if alphabetic. Otherwise, empty string is returned.
2. Multi-Match Evaluation:
   - If queried prefixes slice is empty, or contains "all" or an empty string, all ledgers match (pass-through).
   - If queried prefixes specify concrete prefixes (e.g. ["AL", "GEN"]), candidate ledgers whose extracted uppercase prefix exists in the queried uppercase set match.
   - Ledgers with null TicketID or unresolvable prefixes do not match specific prefix queries.
3. Distinct Prefix Ordering:
   - ExtractDistinctPrefixes returns an array of non-empty, uppercase prefix strings sorted in ascending alphabetical order without duplicates.

### API Endpoint Query Parameter Contracts
All query endpoints accept ticket_type and prefix as comma-separated lists (e.g. ticket_type=feature,bug) or repeated keys (e.g. ticket_type=feature&ticket_type=bug):
1. GET /api/fleet:
   - ticket_type: Optional string or repeated parameter (default all).
   - prefix: Optional string or repeated parameter (default all).
   - window, since, until: Optional temporal filters.
   - Returns FleetSummary computed over matching ledgers.
2. GET /api/workspaces:
   - ticket_type: Optional string or repeated parameter (default all).
   - prefix: Optional string or repeated parameter (default all).
   - window, since, until, sort, order: Optional parameters.
   - Returns []AgenticMetricsLedger matching filters.
3. GET /api/comparison:
   - ticket_type: Optional string or repeated parameter (default all).
   - prefix: Optional string or repeated parameter (default all).
   - window, since, until: Optional temporal filters.
   - Returns []TicketComparisonRow computed from matching ledgers.
4. GET /api/progression:
   - ticket_type: Optional string or repeated parameter (default all).
   - prefix: Optional string or repeated parameter (default all).
   - window, since, until: Optional temporal filters.
   - Returns []AgenticMetricsLedger matching filters.

### Multi-Select Reactive Presentation Layer Contracts
1. Initial Default State:
   - selectedTicketTypes state initializes to ["all"] across Fleet Overview, Ticket Comparisons, and Metric Progression.
   - selectedPrefixes state initializes to ["all"] across all three views.
2. Multi-Select Toggle Semantics:
   - Toggling "all" clears any individual selections and selects "all".
   - Toggling an individual item (e.g. "feature") deselects "all" and adds/removes that specific item.
   - Deselecting the last active item automatically resets the filter back to ["all"].
   - When ["all"] is active, all candidate workspaces pass through.
3. Filter Option Discovery:
   - Ticket types list offers: "all" (All Types), "feature", "bug", "chore", "docs", "refactor", plus any dynamic types discovered across workspaces.
   - Prefixes list offers: "all" (All Prefixes), plus sorted distinct prefixes discovered across all workspaces.
4. Reactivity:
   - Toggling any option immediately triggers data re-fetch or re-filtering.
   - KPIs, charts, distributions, and tables update in synchrony.
