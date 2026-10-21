# Implementation Plan: AL-28 Multi-Dimensional Filtering by Ticket Type and Workspace Prefix

## 1. Goal and Iteration Scope

Deliver complete vertical implementation of ticket AL-28 incorporating user steering:
- Implement pure domain multi-filtering functions FilterLedgersByTicketTypes, FilterLedgersByPrefixes, and ExtractDistinctPrefixes in Go with case-insensitive matching and alias resolution.
- Extend /api/fleet and /api/workspaces HTTP handlers with multi-value ticket_type and prefix query parameters supporting comma separation and repeated keys.
- Implement ComparisonHandler and ProgressionHandler in backend/internal/adapters/http/handlers.go and register routes in server.go.
- Implement matching pure multi-filtering and extraction functions in TypeScript domain formatters.
- Update TypeScript API client with multi-value parameter support and new endpoints.
- Integrate multi-select Ticket Type and Workspace Prefix dropdown controls across FleetOverview.tsx, TicketComparisons.tsx, and MetricProgression.tsx, defaulting both to "all".
- Guarantee zero any, zero emojis, and robust test suites across Go and TypeScript.

---

## 2. Grounded Reference Architecture & File Paths

The plan grounds against existing files and symbols:

### Go Functional Core (backend/internal/domain/)
- backend/internal/domain/models.go: Extend TicketComparisonRow struct with TicketType *string field tagged with json:"ticket_type,omitempty".
- backend/internal/domain/aggregations.go:
  - Implement CanonicalTicketType(raw string) string.
  - Implement FilterLedgersByTicketTypes(ledgers []AgenticMetricsLedger, ticketTypes []string) []AgenticMetricsLedger.
  - Implement FilterLedgersByTicketType convenience wrapper.
  - Implement ExtractTicketPrefix(ticketID string) string.
  - Implement FilterLedgersByPrefixes(ledgers []AgenticMetricsLedger, prefixes []string) []AgenticMetricsLedger.
  - Implement FilterLedgersByPrefix convenience wrapper.
  - Implement ExtractDistinctPrefixes(ledgers []AgenticMetricsLedger) []string.
  - Update BuildComparisonRows to propagate TicketType from ledger into TicketComparisonRow.
- backend/internal/domain/aggregations_test.go: Adversarial unit tests for multi-value ticket type aliases, case insensitivity, multi-prefix extraction, non-conforming IDs, empty and "all" pass-through, distinct prefix deduplication and sorting.

### Go Boundary Adapters (backend/internal/adapters/)
- backend/internal/adapters/http/dto.go: Update ResponsePayload union interface to include []domain.TicketComparisonRow; add WriteComparisonResponse helper.
- backend/internal/adapters/http/handlers.go:
  - Add parseMultiQueryParam helper to parse comma-separated and repeated query values into clean string slices.
  - Update FleetHandler and WorkspacesHandler to parse multi-value ticket_type and prefix query parameters and apply domain multi-filters.
  - Implement ComparisonHandler to handle GET /api/comparison.
  - Implement ProgressionHandler to handle GET /api/progression.
- backend/internal/adapters/http/server.go: Register routes GET /api/comparison and GET /api/progression.
- backend/internal/adapters/http/server_test.go: HTTP integration tests verifying multi-value query parameter handling (comma-delimited and repeated), filtering accuracy, and response payloads.

### TypeScript Domain & API (web/src/domain/ & web/src/api/)
- web/src/domain/types.ts: Define StandardTicketType union; add ticket_type to TicketComparisonRow interface.
- web/src/domain/formatters.ts:
  - Implement canonicalTicketType.
  - Implement extractTicketPrefix.
  - Implement filterLedgersByTicketTypes supporting multi-selection and "all" pass-through.
  - Implement filterLedgersByPrefixes supporting multi-selection and "all" pass-through.
  - Implement extractDistinctPrefixes.
  - Implement extractDistinctTicketTypes.
  - Update deriveTicketComparisonRows to populate ticket_type on rows.
- web/src/domain/formatters.test.ts: Unit tests for multi-alias equivalence, multi-prefix extraction, and multi-filtering.
- web/src/api/client.ts:
  - Update fetchFleetSummary, fetchWorkspaceLedgers, and fetchWorkspaces signatures to accept multi-value ticketTypesParam and prefixesParam.
  - Implement fetchComparisonRows and fetchProgressionLedgers with mock fallbacks.
- web/src/api/client.test.ts: Verify multi-value URL query strings and mock fallbacks.

### TypeScript Presentation (web/src/views/ & web/src/components/)
- web/src/components/MultiSelectDropdown.tsx: Reusable multi-select dropdown control supporting option checkboxes, selection summary label, "All" auto-toggle, and keyboard accessibility.
- web/src/views/FleetOverview.tsx:
  - Add selectedTicketTypes state defaulting to ["all"].
  - Add selectedPrefixes state defaulting to ["all"].
  - Discover distinct prefixes and dynamic ticket types across workspaces on mount.
  - Render Ticket Type and Workspace Prefix multi-select controls within filter bar.
  - Pass filter parameters to fetch calls in loadData.
- web/src/views/FleetOverview.test.tsx: Integration tests for multi-select filter controls, default "all" selection, and zero emojis.
- web/src/views/TicketComparisons.tsx:
  - Add selectedTicketTypes and selectedPrefixes multi-select states defaulting to ["all"].
  - Render multi-select controls above comparison matrix.
  - Filter comparison rows and update benchmark stats reactively.
- web/src/views/TicketComparisons.test.tsx: Integration tests verifying table filtering on multi-select toggles.
- web/src/views/MetricProgression.tsx:
  - Add selectedTicketTypes and selectedPrefixes multi-select states defaulting to ["all"].
  - Render multi-select controls alongside granularity and preset controls.
  - Apply pure domain multi-filtering to ledgers before computing progression buckets via aggregateMetricsOverTime.
- web/src/views/MetricProgression.test.tsx: Integration tests verifying chart and KPI reactivity.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Go Domain Core Implementation
- Task 1.1: In backend/internal/domain/models.go:
  - Add TicketType *string to TicketComparisonRow with JSON tag "ticket_type,omitempty".
- Task 1.2: In backend/internal/domain/aggregations.go:
  - Implement CanonicalTicketType(raw string) string:
    - Trim whitespace, lowercase.
    - Map "feat" and "feature" to "feature".
    - Map "bug", "fix", and "bugfix" to "bug".
    - Map "doc" and "docs" to "docs".
    - Return trimmed lowercase string for any other input.
  - Implement FilterLedgersByTicketTypes(ledgers []AgenticMetricsLedger, ticketTypes []string) []AgenticMetricsLedger:
    - Return empty slice if ledgers is empty.
    - If ticketTypes is empty, or contains "all" (case-insensitive) or an empty token, return a newly allocated copy of ledgers.
    - Build set of canonical target types using CanonicalTicketType.
    - If set is empty, return copy of ledgers.
    - Filter ledgers: keep ledger if ledger.TicketType is not nil and CanonicalTicketType(*ledger.TicketType) exists in target set.
    - Return newly allocated filtered slice.
  - Implement FilterLedgersByTicketType(ledgers []AgenticMetricsLedger, ticketType string) []AgenticMetricsLedger:
    - Split ticketType by comma if present, trim whitespace, and delegate to FilterLedgersByTicketTypes.
  - Implement ExtractTicketPrefix(ticketID string) string:
    - Trim whitespace. Return empty string if empty.
    - Apply ticketIDPattern regex. If match found, return strings.ToUpper(match[1]).
    - Fallback: check for delimiter "-" or "_", extract leading token if alphabetic.
    - Return uppercase string or empty string.
  - Implement FilterLedgersByPrefixes(ledgers []AgenticMetricsLedger, prefixes []string) []AgenticMetricsLedger:
    - Return empty slice if ledgers is empty.
    - If prefixes is empty, or contains "all" (case-insensitive) or an empty token, return a newly allocated copy of ledgers.
    - Build set of cleaned uppercase target prefixes.
    - If set is empty, return copy of ledgers.
    - Filter ledgers: keep ledger if ledger.TicketID is not nil and ExtractTicketPrefix(*ledger.TicketID) exists in target set.
    - Return newly allocated filtered slice.
  - Implement FilterLedgersByPrefix(ledgers []AgenticMetricsLedger, prefix string) []AgenticMetricsLedger:
    - Split prefix by comma if present, trim whitespace, and delegate to FilterLedgersByPrefixes.
  - Implement ExtractDistinctPrefixes(ledgers []AgenticMetricsLedger) []string:
    - Return empty slice if ledgers is empty.
    - Collect unique non-empty prefixes into map.
    - Extract keys into slice and sort with sort.Strings.
    - Return sorted slice.
  - In BuildComparisonRows:
    - Populate TicketType: ledger.TicketType on each TicketComparisonRow.

### Phase 2: Go Boundary Shell HTTP Handlers and Server Routes
- Task 2.1: In backend/internal/adapters/http/dto.go:
  - Extend ResponsePayload union interface to include []domain.TicketComparisonRow.
  - Implement WriteComparisonResponse(w http.ResponseWriter, rows []domain.TicketComparisonRow).
- Task 2.2: In backend/internal/adapters/http/handlers.go:
  - Implement parseMultiQueryParam(values []string) []string:
    - Iterate over values slice (supporting repeated query parameters).
    - For each item, split by comma to support comma-delimited tokens.
    - Trim whitespace, discard empty strings.
    - Return slice of cleaned tokens.
  - In WorkspacesHandler:
    - Extract ticket_type and prefix query slices via r.URL.Query()["ticket_type"] and r.URL.Query()["prefix"].
    - Parse using parseMultiQueryParam.
    - Apply domain.FilterLedgersByTicketTypes and domain.FilterLedgersByPrefixes after temporal filtering.
  - In FleetHandler:
    - Extract ticket_type and prefix query slices via r.URL.Query()["ticket_type"] and r.URL.Query()["prefix"].
    - Parse using parseMultiQueryParam.
    - Apply domain.FilterLedgersByTicketTypes and domain.FilterLedgersByPrefixes before CalculateFleetSummary.
  - Implement ComparisonHandler(workspacesDir string) http.HandlerFunc:
    - Scan workspaces, parse temporal filters.
    - Parse ticket_type and prefix via parseMultiQueryParam.
    - Apply domain.FilterLedgersByCompletionDate, domain.FilterLedgersByTicketTypes, and domain.FilterLedgersByPrefixes.
    - Compute domain.BuildComparisonRows and write response with WriteComparisonResponse.
  - Implement ProgressionHandler(workspacesDir string) http.HandlerFunc:
    - Scan workspaces, parse temporal filters.
    - Parse ticket_type and prefix via parseMultiQueryParam.
    - Apply domain.FilterLedgersByCompletionDate, domain.FilterLedgersByTicketTypes, and domain.FilterLedgersByPrefixes.
    - Write response with WriteWorkspacesResponse.
- Task 2.3: In backend/internal/adapters/http/server.go:
  - Register mux.HandleFunc("GET /api/comparison", ComparisonHandler(workspacesDir)).
  - Register mux.HandleFunc("GET /api/progression", ProgressionHandler(workspacesDir)).

### Phase 3: TypeScript Domain and API Client Enhancements
- Task 3.1: In web/src/domain/types.ts:
  - Define export type StandardTicketType = 'feature' | 'bug' | 'chore' | 'docs' | 'refactor' | 'all';.
  - Add readonly ticket_type?: string | null; to TicketComparisonRow interface.
- Task 3.2: In web/src/domain/formatters.ts:
  - Implement canonicalTicketType(raw: string): string matching Go alias rules.
  - Implement extractTicketPrefix(ticketId: string | null | undefined): string matching Go extraction rules.
  - Implement filterLedgersByTicketTypes<T extends { readonly ticket_type?: string | null }>(ledgers: readonly T[], ticketTypes?: readonly string[] | string | null): T[]:
    - Support string or array. If contains "all", null, or empty, return copy of ledgers.
    - Replicate canonical set membership check.
  - Implement filterLedgersByPrefixes<T extends { readonly ticket_id?: string | null }>(ledgers: readonly T[], prefixes?: readonly string[] | string | null): T[]:
    - Support string or array. If contains "all", null, or empty, return copy of ledgers.
    - Replicate uppercase prefix set membership check.
  - Implement extractDistinctPrefixes(items: readonly { readonly ticket_id?: string | null }[]): string[].
  - Implement extractDistinctTicketTypes(items: readonly { readonly ticket_type?: string | null }[]): string[].
  - Update deriveTicketComparisonRows to map ticket_type: ledger.ticket_type.
- Task 3.3: In web/src/api/client.ts:
  - Update fetchFleetSummary to accept ticketTypesParam?: readonly string[] | string and prefixesParam?: readonly string[] | string:
    - Format arrays into comma-separated query values if not "all".
  - Update fetchWorkspaceLedgers to accept multi-value ticketTypesParam and prefixesParam, filtering mock fallback data via pure formatters.
  - Update fetchWorkspaces to accept multi-value ticketTypesParam and prefixesParam, filtering mock fallback data.
  - Implement fetchComparisonRows with multi-value parameters querying /api/comparison.
  - Implement fetchProgressionLedgers with multi-value parameters querying /api/progression.

### Phase 4: TypeScript Presentation Layer
- Task 4.1: In web/src/components/MultiSelectDropdown.tsx:
  - Create accessible multi-select dropdown component:
    - Props: id, label, options (value and label), selectedValues (string[]), onChange (values: string[]) => void.
    - Summary trigger button showing "All" when ["all"] is selected, or comma-separated selected labels, or badge count when multiple are selected.
    - Dropdown panel containing checkboxes for each option.
    - Toggling "all" clears other selections. Toggling individual option deselects "all". If all items are deselected, resets to ["all"].
- Task 4.2: In web/src/views/FleetOverview.tsx:
  - Introduce selectedTicketTypes state defaulting to ["all"].
  - Introduce selectedPrefixes state defaulting to ["all"].
  - Discover distinct prefixes and dynamic ticket types across workspaces on mount.
  - Render Ticket Type and Workspace Prefix multi-select controls within filter bar.
  - Pass filter parameters to fetchFleetSummary and fetchWorkspaces in loadData.
  - Trigger loadData reactively when either filter selection updates.
- Task 4.3: In web/src/views/TicketComparisons.tsx:
  - Introduce selectedTicketTypes and selectedPrefixes states defaulting to ["all"].
  - Retain full unfiltered ledger list for prefix and type discovery.
  - Render Ticket Type and Workspace Prefix multi-select controls above comparison matrix.
  - Apply pure multi-filtering before deriving comparison rows and benchmark stats.
  - Update benchmark summary cards and comparison matrix reactively.
- Task 4.4: In web/src/views/MetricProgression.tsx:
  - Introduce selectedTicketTypes and selectedPrefixes states defaulting to ["all"].
  - Retain full unfiltered ledger list for prefix and type discovery.
  - Render Ticket Type and Workspace Prefix multi-select controls alongside granularity and preset controls.
  - Apply pure domain multi-filtering to ledgers before computing progression buckets via aggregateMetricsOverTime.
  - Update summary KPIs, SVG trend charts, and ceiling lines reactively.

---

## 4. Adversarial Testing Strategy for Tester

### Pure Go Domain Tests (backend/internal/domain/aggregations_test.go)
- Table-driven TestFilterLedgersByTicketTypes:
  - Default / pass-through: empty slice, slice with "all", slice with mixed "all" and concrete types returns all ledgers.
  - Single selection: ["feature"] matches ledgers with "feature" and "feat".
  - Multi-selection: ["feature", "bug"] matches ledgers with "feature", "feat", "bug", "fix", and "bugfix".
  - Case insensitivity: ["FEATURE", "fIx"] matches correctly.
  - Disjoint set: querying ["docs"] when only feature ledgers exist returns empty slice.
  - Nil handling: ledgers with nil TicketType are excluded gracefully on specific queries.
  - Immutability: assert modifying returned slice does not mutate input ledgers.
- Table-driven TestFilterLedgersByPrefixes:
  - Default / pass-through: empty slice, slice with "all" returns all ledgers.
  - Multi-prefix selection: ["AL", "GEN"] matches AL-1, AL-28, GEN-4 while excluding PO-2.
  - Case insensitivity: ["al", "gen"] matches uppercase and lowercase identifiers.
  - Disjoint set: querying non-existent prefix returns empty slice.
  - Nil / non-conforming ticket IDs: ignored gracefully without panic.
- Table-driven TestExtractDistinctPrefixes:
  - Deduplication: multiple AL-* tickets produce single "AL" entry.
  - Sorting: mixed prefix ledgers produce strictly ascending alphabetical slice.
  - Nil and empty ticket IDs: ignored gracefully.
  - Empty input: returns empty slice.

### Go Boundary Adapter Tests (backend/internal/adapters/http/server_test.go)
- Test /api/fleet with comma-separated types: /api/fleet?ticket_type=feature,bug verifies summary aggregates both feature and bug ledgers.
- Test /api/fleet with repeated query parameters: /api/fleet?ticket_type=feature&ticket_type=bug verifies identical behavior to comma separation.
- Test /api/workspaces with multi-prefix: /api/workspaces?prefix=AL,GEN verifies returned ledgers are filtered to specified prefixes.
- Test /api/comparison endpoint: verify GET /api/comparison?ticket_type=feature,chore&prefix=AL returns HTTP 200 with []TicketComparisonRow computed from filtered workspaces.
- Test /api/progression endpoint: verify GET /api/progression?prefix=AL,PO returns HTTP 200 with filtered []AgenticMetricsLedger.

### TypeScript Domain & Formatter Tests (web/src/domain/formatters.test.ts)
- Test filterLedgersByTicketTypes:
  - Verify multi-selection matching across aliases ("feat" + "fix").
  - Verify "all" and empty inputs return complete copy.
  - Verify unaliased types match cleanly.
- Test filterLedgersByPrefixes:
  - Verify multi-selection matching across prefixes (["AL", "GEN"]).
  - Verify "all" and empty inputs return complete copy.
- Test extractDistinctPrefixes and extractDistinctTicketTypes:
  - Verify deduplication and alphabetical sorting.

### TypeScript API Client Tests (web/src/api/client.test.ts)
- Verify fetchFleetSummary appends multi-value ticket_type and prefix query parameters as comma-separated values.
- Verify fetchWorkspaces appends multi-value parameters.
- Verify fetchWorkspaceLedgers appends multi-value parameters.
- Verify fetchComparisonRows and fetchProgressionLedgers construct accurate URLs and handle mock fallbacks.

### TypeScript View Integration Tests
- web/src/views/FleetOverview.test.tsx:
  - Verify initial render defaults to ticket_type=["all"] and prefix=["all"], displaying all tickets.
  - Verify toggling ticket type options updates API client calls with comma-separated list.
  - Verify toggling prefix options updates API client calls.
  - Verify zero emojis in rendered DOM.
- web/src/views/TicketComparisons.test.tsx:
  - Verify initial load displays all comparison rows under default "all".
  - Verify selecting multiple ticket types filters rows and updates benchmark summary cards.
  - Verify selecting multiple prefixes filters rows accordingly.
  - Verify zero emojis in rendered DOM.
- web/src/views/MetricProgression.test.tsx:
  - Verify initial load aggregates progression across all ledgers under default "all".
  - Verify toggling multiple ticket types or prefixes updates progression buckets and trend chart curves.
  - Verify zero emojis in rendered DOM.

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
