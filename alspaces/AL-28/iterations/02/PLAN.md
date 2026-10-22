# Implementation Plan: AL-28 (Iteration 02) Persistent Filter Controls on Zero Progression Telemetry

## 1. Goal and Root Cause Analysis

### User Steering Scenario
When a user selects Ticket Type and Prefix filter combinations that match zero tickets (such as selecting a ticket type or prefix with no recorded workspaces), Fleet Overview retains its filter controls and displays an empty state beneath them. In Metric Progression, however, selecting an unmatched filter combination causes the controls bar to vanish entirely, displaying only the empty state banner "Zero Progression Telemetry Available". The user cannot adjust their selection, reset filters to "All", or pick another ticket type or prefix without performing a full browser reload.

### Root Cause Analysis in `MetricProgression.tsx`
1. Early Return Preempts Container Rendering:
   In `web/src/views/MetricProgression.tsx` (lines 230-237), an early return triggers immediately when `ledgers.length === 0`. The component returns the `EmptyState` component directly, terminating render execution before returning the outer container `progression-container`.
2. Controls Bar Positioned Downstream:
   The top filter and controls bar (`progression-controls-bar` and conditional `custom-datetime-bar`) is declared inside `progression-container` at line 256. Because of the early return on empty ledgers, `progression-controls-bar` is unmounted from the DOM whenever the backend or client mock returns zero matching ledgers.
3. Contrast with `FleetOverview.tsx`:
   In `web/src/views/FleetOverview.tsx`, the filter control bar (`dateFilterControl`) is rendered unconditionally at the top of the component view, preceding the ternary evaluation of loading, error, and empty states (`!summary || summary.total_tickets === 0`). When no tickets match, `EmptyState` renders beneath `dateFilterControl`, leaving all dropdowns and presets interactive.
4. Alignment Analysis for `TicketComparisons.tsx`:
   In `web/src/views/TicketComparisons.tsx` (lines 213-220), a similar early return (`if (ledgers.length === 0) return <EmptyState ... />`) preempts the rendering of `comparisons-container` and its `table-controls-bar`. Aligning `TicketComparisons.tsx` ensures uniform behavior across all dashboard views.

---

## 2. Grounded Reference Architecture and Concrete Changes

### Scope Boundaries
- `web/src/views/MetricProgression.tsx`: Restructure render hierarchy so `progression-controls-bar` and `custom-datetime-bar` remain permanently mounted above view states.
- `web/src/views/TicketComparisons.tsx`: Restructure render hierarchy so `table-controls-bar` remains mounted and interactive when filter selections yield zero matching workspaces.
- `web/src/views/MetricProgression.test.tsx`: Add regression test cases verifying controls bar presence and dropdown interactivity when filters produce zero matching ledgers.
- `web/src/views/TicketComparisons.test.tsx`: Add test cases verifying filter controls persistence when filters yield zero matching ledgers.

---

## 3. Phased Implementation Tasks for Coder

### Phase 1: Restructure `MetricProgression.tsx` Render Hierarchy
- Task 1.1: Reorganize Component Return Structure:
  - Eliminate top-level early returns for `loading`, `error`, and `ledgers.length === 0` that bypass the outer container.
  - Return `div.progression-container` as the root wrapper element of the component.
  - Render `div.progression-controls-bar` unconditionally as the first child of `div.progression-container`.
  - Render `div.custom-datetime-bar` conditionally immediately beneath `div.progression-controls-bar` when `preset === 'custom'`.
- Task 1.2: Conditional View Content Area Beneath Controls:
  - Inside `div.progression-container`, below the controls and custom datetime bars, render the content body using conditional branches:
    - If `loading` is true and data has not previously loaded, render `div.loading-container` containing the spinner and loading text.
    - Else if `error` is not null, render `div.error-banner` with error message and retry button.
    - Else if `ledgers.length === 0` or `filteredLedgers.length === 0`, render `EmptyState` with title "Zero Progression Telemetry Available" and message "No workspace telemetry files were found. Run an agentic workflow to populate longitudinal progression charts."
    - Else, render the active telemetry progression view: `div.kpi-grid` (Period Spend, Total Invocations, Total Token Throughput, Window Cache Hit Rate, Net Code Delivered) followed by either `div.empty-progression-notice` (if buckets array is empty) or `div.progression-charts-grid` (with the five SVG trend charts).
- Task 1.3: Maintain Filter Option Interactivity on Empty Ledgers:
  - Verify `knownPrefixes` state persistence: existing logic guards updates with `if (ledgers.length > 0)`, ensuring previously discovered prefixes are not cleared when empty ledgers arrive. Fallback to default prefixes `['AL', 'GEN', 'PO']` ensures options remain available.
  - Verify `TICKET_TYPE_OPTIONS` static list remains fully interactive.
  - Ensure `MultiSelectDropdown` controls for Ticket Type (`#prog-ticket-type`) and Prefix (`#prog-prefix`) receive current `selectedTicketTypes` and `selectedPrefixes` and continue invoking `loadLedgers` upon user selection changes without needing page refresh.

### Phase 2: Align `TicketComparisons.tsx` Render Hierarchy
- Task 2.1: Reorganize `TicketComparisons.tsx` Return Structure:
  - Eliminate top-level early return on `ledgers.length === 0` that bypasses `comparisons-container`.
  - Ensure `div.table-controls-bar` (containing search input, `#comp-ticket-type`, `#comp-prefix`, and row count display) remains mounted when `ledgers.length === 0` or `filteredRows.length === 0`.
  - When `ledgers.length === 0` or `filteredRows.length === 0`, keep benchmarking cards and controls bar visible, and render the empty fallback row ("No tickets match the selected filters.") or `EmptyState` below the controls bar.
  - Ensure `#comp-ticket-type` and `#comp-prefix` remain interactive so users can reselect filters.

---

## 4. Testing Strategy for Tester

### Unit and Integration Tests in `MetricProgression.test.tsx`
- Test 4.1: Persistent Controls Bar on Empty Filter Results:
  - Mock `apiClient.fetchWorkspaceLedgers` to resolve with empty array (`data: []`).
  - Render `MetricProgression`.
  - Assert that `EmptyState` title "Zero Progression Telemetry Available" is present.
  - Assert that `div.progression-controls-bar` is present in document.
  - Assert that Ticket Type combobox trigger and Prefix combobox trigger are present in document and accessible.
  - Assert that window preset buttons (Lifetime, 24 Hours, 7 Days, etc.) and granularity buttons remain present.
- Test 4.2: Interactive Recovery from Empty Filter Selection:
  - Mock `apiClient.fetchWorkspaceLedgers` initially returning mock ledgers.
  - Render `MetricProgression` and wait for KPI cards to display.
  - Simulate selecting a filter combination that yields zero matching tickets (e.g. Bug).
  - Assert that `EmptyState` appears and controls bar remains visible.
  - Click the Ticket Type dropdown trigger, click the checkbox for "Select all Ticket Type" or "Feature".
  - Assert that `fetchWorkspaceLedgers` is invoked with updated parameters and telemetry charts re-render upon data arrival.
- Test 4.3: Regression Assurance:
  - Verify existing tests continue passing: loading state, SVG trend chart rendering, granularity toggles, time window presets, tooltip hover and focus interactions, and zero emojis check.

### Integration Tests in `TicketComparisons.test.tsx`
- Test 4.4: Persistent Filter Bar on Zero Comparison Matches:
  - Verify that when `fetchWorkspaceLedgers` returns zero ledgers matching filter parameters, the controls bar (`table-controls-bar`) remains in the DOM and interactive.

---

## 5. Non-Negotiables and Quality Gates

1. Strong Typing: Zero `any` or empty `interface{}` across all TypeScript files.
2. Zero Emojis: Strictly zero emojis across UI labels, tooltips, empty state messages, test names, and code comments.
3. Functional Core Isolation: Presentation components consume pure domain formatters without performing ad-hoc arithmetic or mutation.
4. Quality Gates:
   - `npm run typecheck` passes with zero errors.
   - `npm run lint` passes with zero warnings.
   - `npm test` passes all tests across `FleetOverview`, `TicketComparisons`, and `MetricProgression`.
   - `make check && make test && make build` succeeds.
