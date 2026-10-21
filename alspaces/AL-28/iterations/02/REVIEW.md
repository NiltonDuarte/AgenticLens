# Code Review Report: AL-28 (Iteration 02) Persistent Filter Controls on Zero Progression Telemetry

**Target Artifacts**: `workspaces/AL-28/CONTRACT.md`, `workspaces/AL-28/iterations/02/PLAN.md`, `TECHNICAL_GUIDELINES.md`  
**Iteration**: 02  
**Verdict**: `APPROVED`  

---

## 1. Executive Summary

Iteration 02 resolves the control bar unmounting bug identified during user steering. Previously, selecting a ticket type or prefix combination matching zero records triggered top-level early returns in `MetricProgression.tsx` and `TicketComparisons.tsx`, stripping the controls bar (`progression-controls-bar` / `table-controls-bar`) and trapping the user on the `EmptyState` banner without a full page refresh.

The implementation restructures the component tree in both views to mount the filter controls bar unconditionally at the top of the container. When zero telemetry records match, `EmptyState` renders directly beneath the interactive controls bar. Filter state and known prefixes remain preserved, allowing immediate recovery and reactive re-querying.

---

## 2. Axis 1: Spec & Contract Alignment

### 1.1 Requirements Verification
- **Persistent Controls in `MetricProgression.tsx`**:
  - The root element is now `div.progression-container`.
  - `div.progression-controls-bar` is rendered unconditionally as the first child, retaining Window presets, Ticket Type dropdown (`#prog-ticket-type`), Prefix dropdown (`#prog-prefix`), and Granularity toggles.
  - `div.custom-datetime-bar` renders conditionally below the controls bar when `preset === 'custom'`.
  - View states (`loading`, `error`, `EmptyState`, and active charts grid) render conditionally in the content slot beneath the controls bar.
  - When `ledgers.length === 0 || filteredLedgers.length === 0`, `EmptyState` renders gracefully without unmounting the controls.
- **Alignment in `TicketComparisons.tsx`**:
  - Top-level early return on `ledgers.length === 0` has been removed.
  - `comparisons-container` mounts `kpi-grid` and `table-controls-bar` unconditionally.
  - When `ledgers.length === 0`, `EmptyState` renders beneath `table-controls-bar`.
  - When `ledgers.length > 0` but `filteredRows.length === 0`, the table fallback row (`No tickets match the selected filters.`) displays beneath the controls.
  - Search input, `#comp-ticket-type`, and `#comp-prefix` remain interactive across all empty states.
- **Zero Scope Creep**:
  - Changes strictly target layout hierarchy and regression coverage for filter persistence across empty states. No speculative refactorings or unrequested features were introduced.

### 1.2 Boundary & Error Resilience
- **Defensive Benchmarks Rollup**:
  - In `TicketComparisons.tsx`, zero-division protection (`total === 0`) safely produces default zero-value metrics (`$0.00`, `0m 0s`, `0.0%`) when ledgers are empty.
- **Prefix Discovery Retention**:
  - `knownPrefixes` preserves discovered prefix sets across temporary zero-match queries via the `if (ledgers.length > 0)` guard with static fallback `['AL', 'GEN', 'PO']`.
- **Zero Skips**:
  - All test suites execute active assertions without `.skip` or `.todo`.

---

## 3. Axis 2: Engineering Standards & Architecture

### 2.1 Functional Core, Imperative Shell (FCIS) Isolation
- Both components strictly consume pure domain functions from `../domain/formatters` (`filterLedgersByTicketTypes`, `filterLedgersByPrefixes`, `extractTicketPrefix`, `deriveTicketComparisonRows`).
- Zero domain logic or arithmetic leakage inside presentation event handlers.
- Pure Go domain core (`backend/internal/domain/`) remains entirely untouched and isolated from outer layers or I/O.

### 2.2 Strict Typing & Zero Any
- Go backend: Zero `any` or untyped `interface{}`.
- TypeScript frontend: Zero `any` across production and test code.
- Caught errors in both views use `err: unknown` safely narrowed via `err instanceof Error`.

### 2.3 UI & Presentation Invariants
- Zero emojis present across UI labels, tooltips, error banners, test names, and code comments.
- Both components render accessible `EmptyState` banners without unhandled runtime exceptions.

---

## 4. Test Suite Diagnostics & Coverage

| Test File | Test Cases Added / Verified | Status |
|---|---|---|
| `web/src/views/MetricProgression.test.tsx` | - Persistent controls bar when `fetchWorkspaceLedgers` returns `[]`<br>- Interactive recovery from empty filter selection without reload<br>- Controls bar persistence on client-side zero-match filtering<br>- Re-triggering data fetch on prefix change from empty state | PASS |
| `web/src/views/TicketComparisons.test.tsx` | - Persistent table controls bar when zero ledgers match filters<br>- Interactive recovery from empty filter selection<br>- Re-triggering data fetch from empty state | PASS |

---

## 5. Final Audit Verdict

Status: **`APPROVED`**

The implementation faithfully satisfies `CONTRACT.md` and `iterations/02/PLAN.md`. Layout hierarchy restructuring cleanly preserves control interactivity across empty telemetry states while maintaining strict typing, FCIS separation, and zero emoji invariants.
