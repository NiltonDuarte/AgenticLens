# Implementation Plan: AL-33 Iteration 04 Trace Waterfall Multi-Select and "No filter" Labeling

## 1. Goal and Iteration Scope

Iteration 04 addresses user steering feedback regarding the execution trace waterfall filtering semantics in the web dashboard:
1. **Labeling Invariant**: The unselected, lack-of-filtering state for Tool, Anomaly, and Role filters must explicitly say "No filter" instead of "All tools", "All anomalies", or "All Roles". Unselected means events without any tool or anomaly are retained rather than filtered out.
2. **Multi-Selection Invariant**: Users must be able to select multiple tools, multiple anomaly rule identifiers, and multiple subagent roles simultaneously to isolate complex cross-cutting failure patterns.
3. **Component Reusability**: Extend the existing `MultiSelectDropdown` component with configurable unselected labels and accessible attributes, replacing the single-select HTML elements in `TraceWaterfall` while preserving complete backward compatibility across existing views (`FleetOverview`, `MetricProgression`, `TicketComparisons`).
4. **Strict Conjunction Filtering**: Compose multi-select filter arrays with the existing `filterMode` toggle using logical conjunction.

The scope is strictly confined to frontend components, styling, and unit test suites:
- `web/src/components/MultiSelectDropdown.tsx`
- `web/src/components/MultiSelectDropdown.test.tsx`
- `web/src/components/TraceWaterfall.tsx`
- `web/src/components/TraceWaterfall.test.tsx`
- `web/src/index.css`

Go backend packages in `backend/internal/domain/` and `backend/internal/adapters/` are untouched.

---

## 2. Grounded Analysis of Existing Codebase

### MultiSelectDropdown (`web/src/components/MultiSelectDropdown.tsx`)
- Current API: Accepts `id`, `label`, `options` (`MultiSelectOption` array), `selectedValues` (string array), `onChange` handler, and optional `placeholder` (defaulting to `'All'`).
- Hardcoded Unselected Text: The top option in the dropdown list hardcodes a checkbox labeled with a span containing `'All'` and an aria-label set to `"Select all [label]"`. The trigger button `displaySummary` returns `placeholder` when `isAllSelected` evaluates to true.
- Existing Usages: Used by `FleetOverview.tsx` (Prefix, Type), `MetricProgression.tsx` (Prefix, Ticket Type), and `TicketComparisons.tsx` (Prefix, Type). These call sites rely on the default `'All'` placeholder and summary.

### Trace Waterfall (`web/src/components/TraceWaterfall.tsx`)
- Current Implementation: Uses three independent single-select HTML select elements for `roleFilter`, `toolFilter`, and `anomalyFilter`, each holding a single string initialized to `'all'`.
- Misleading Semantics: Options are titled "All Roles", "All Tools", and "All Anomalies". An event without any `tool_name` passes when `toolFilter` is `'all'`, but users intuitively interpret "All Tools" as "must have a tool". Naming the unselected state "No filter" accurately describes the pass-through behavior.
- Single-Select Limitation: A user cannot concurrently inspect `bash` and `edit` failures, nor view events flagged by either `TOOL_RETRY_STREAK` or `CONTEXT_SATURATION_SPIKE`.

---

## 3. Detailed Architectural Specifications

### 3.1 MultiSelectDropdown Enhancements (`web/src/components/MultiSelectDropdown.tsx`)
1. Interface Extensions on `MultiSelectDropdownProps`:
   - `allLabel` (optional string): Defines the label text for the unselected pass-through state. Defaults to `'All'` when omitted. When provided as `"No filter"`, the top option displays `"No filter"`.
   - `placeholder` (optional string): Defaults to `allLabel` if `allLabel` is provided, otherwise defaults to `'All'`.
   - `ariaLabel` (optional string): Overrides the default combobox trigger button aria-label. Defaults to `"${label} selector"`.
   - `testId` (optional string): Assigned to `data-testid` on the trigger button. Defaults to `customId` or `autoId`.
   - `className` (optional string): Custom CSS class string appended to the container element for layout positioning.
2. Top Option Label and Accessibility:
   - Checkbox label text inside the dropdown list renders `allLabel` instead of hardcoded `'All'`.
   - Checkbox aria-label evaluates to `allLabel` when `allLabel` is passed, or `"Select all ${label}"` when omitted, preserving existing test expectations in `MultiSelectDropdown.test.tsx`.
3. Trigger Summary Display:
   - When `isAllSelected` is true, `displaySummary` returns `placeholder` (which evaluates to `allLabel` or `'All'`).
   - When 1 item is selected, returns the label of that item.
   - When 2 items are selected, returns comma-separated item labels.
   - When 3 or more items are selected, returns `"${count} selected"`.
4. Backward Compatibility Invariant:
   - If `allLabel`, `ariaLabel`, `testId`, and `className` are omitted, component behavior and rendered DOM are 100% identical to prior iterations.

### 3.2 TraceWaterfall Multi-Select State and Options (`web/src/components/TraceWaterfall.tsx`)
1. State Migration:
   - `selectedRoles`: Readonly array of strings, initialized to `['all']`.
   - `selectedTools`: Readonly array of strings, initialized to `['all']`.
   - `selectedAnomalies`: Readonly array of strings, initialized to `['all']`.
   - `filterMode`: Retains `'all'`, `'tools'`, `'errors_anomalies'` (defaulting to `'errors_anomalies'`).
2. Derived Option Collections:
   - `availableRoles`: Memoized extraction of unique non-empty role strings from `events`, sorted alphabetically via `localeCompare`.
   - `roleOptions`: Mapped to `MultiSelectOption` objects with `value` and `label` set to role.
   - `availableTools`: Memoized extraction of unique non-null, non-empty, trimmed `tool_name` strings from `events`, sorted alphabetically.
   - `toolOptions`: Mapped to `MultiSelectOption` objects with `value` and `label` set to tool name.
   - `availableAnomalies`: Memoized extraction of unique non-null, non-empty, trimmed `rule_id` strings from `anomalies`, sorted alphabetically.
   - `anomalyOptions`: Mapped to `MultiSelectOption` objects with `value` and `label` set to rule ID.

### 3.3 Multi-Dimensional Conjunction Filtering Engine
The `filteredEvents` memoization in `TraceWaterfall.tsx` filters `sortedEvents` through four sequential criteria combined with boolean AND:
1. **Role Filter Criterion**:
   - Active Check: `selectedRoles` is active if length is greater than 0 and does not include `'all'` or `'ALL'`.
   - If active: Event must have non-empty role and `selectedRoles` must include `event.role`. Events without a role evaluate to false.
   - If inactive ("No filter"): Condition passes for all events.
2. **Filter Mode Criterion**:
   - If `filterMode` is `'tools'`: Event `event_type` must equal `'tool_call'`.
   - If `filterMode` is `'errors_anomalies'`: Event `status` must equal `'error'` OR `findMatchingAnomaly(event, anomalies)` must return a truthy anomaly.
   - If `filterMode` is `'all'`: Condition passes unconditionally.
3. **Tool Filter Criterion**:
   - Active Check: `selectedTools` is active if length is greater than 0 and does not include `'all'` or `'ALL'`.
   - If active: Event `tool_name` must be non-null, non-empty when trimmed, and `selectedTools` must include trimmed `event.tool_name`. Events without a `tool_name` evaluate to false.
   - If inactive ("No filter"): Condition passes for all events, including `turn_message` and `compaction` events.
4. **Anomaly Filter Criterion**:
   - Active Check: `selectedAnomalies` is active if length is greater than 0 and does not include `'all'` or `'ALL'`.
   - If active: Filter anomalies to candidates where `candidate.rule_id` is included in `selectedAnomalies`. Condition passes if `findMatchingAnomaly(event, candidateAnomalies)` returns a truthy anomaly.
   - If inactive ("No filter"): Condition passes for all events, including events with zero anomalies.

### 3.4 Event Counter and Empty State Semantics
- Counter format: `"Showing ${filteredEvents.length} of ${events.length} events"`.
- Total events denominator remains invariant regardless of filter state.
- Empty states:
  - If `events` array is empty or undefined: `"Zero chronological events recorded for this execution trace."`
  - If filters exclude all events: `"No events match the selected filters."`

### 3.5 Accessibility & Design System
- Semantic combobox and listbox ARIA attributes maintained by `MultiSelectDropdown`.
- Explicit `ariaLabel` on each combobox trigger ("Filter by subagent role", "Filter by tool", "Filter by anomaly rule").
- Visible labels "Role:", "Tool:", "Anomaly:" associated with combobox trigger buttons.
- Zero Emojis Invariant: Absolute prohibition of emoji characters in labels, options, trigger summaries, and test assertions.

---

## 4. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Extend MultiSelectDropdown Component
- **Task 1.1**: Update `MultiSelectDropdownProps` in `web/src/components/MultiSelectDropdown.tsx` to accept optional `allLabel`, `ariaLabel`, `testId`, and `className` props.
- **Task 1.2**: Set default value of `placeholder` to `allLabel` when `allLabel` is provided, falling back to `'All'`.
- **Task 1.3**: Update top option in dropdown list to render `allLabel` (or `'All'`) as text content. Set its input aria-label to `allLabel` when present, or `"Select all ${label}"` when omitted.
- **Task 1.4**: Update combobox trigger button to use `ariaLabel` prop (falling back to `"${label} selector"`), `data-testid` from `testId` prop, and merge `className` on the root container.
- **Task 1.5**: Expand `web/src/components/MultiSelectDropdown.test.tsx` with unit tests verifying `allLabel="No filter"` behavior, placeholder fallback, and backward compatibility.

### Phase 2: Migrate TraceWaterfall Filter Controls
- **Task 2.1**: In `web/src/components/TraceWaterfall.tsx`, replace `roleFilter`, `toolFilter`, and `anomalyFilter` string states with `selectedRoles`, `selectedTools`, and `selectedAnomalies` string array states, each initialized to `['all']`.
- **Task 2.2**: Compute `roleOptions`, `toolOptions`, and `anomalyOptions` using `useMemo` from `availableRoles`, `availableTools`, and `availableAnomalies`.
- **Task 2.3**: In `waterfall-toolbar`, replace the three HTML select elements with `MultiSelectDropdown` instances:
  - Role dropdown: `id="trace-role-filter"`, `testId="trace-role-filter"`, `label="Role"`, `allLabel="No filter"`, `ariaLabel="Filter by subagent role"`, `className="role-filter-group"`, `options={roleOptions}`, `selectedValues={selectedRoles}`, `onChange={setSelectedRoles}`.
  - Tool dropdown: `id="trace-tool-filter"`, `testId="trace-tool-filter"`, `label="Tool"`, `allLabel="No filter"`, `ariaLabel="Filter by tool"`, `className="tool-filter-group"`, `options={toolOptions}`, `selectedValues={selectedTools}`, `onChange={setSelectedTools}`.
  - Anomaly dropdown: `id="trace-anomaly-filter"`, `testId="trace-anomaly-filter"`, `label="Anomaly"`, `allLabel="No filter"`, `ariaLabel="Filter by anomaly rule"`, `className="anomaly-filter-group"`, `options={anomalyOptions}`, `selectedValues={selectedAnomalies}`, `onChange={setSelectedAnomalies}`.
- **Task 2.4**: Update `filteredEvents` memoization to evaluate active status and apply multi-selection membership checks across `selectedRoles`, `selectedTools`, and `selectedAnomalies` in conjunction with `filterMode`.

### Phase 3: CSS Layout & Styling Adjustments
- **Task 3.1**: In `web/src/index.css`, review `.waterfall-toolbar` layout to ensure multi-select dropdown containers align horizontally with `flex-wrap` and `gap: 16px`.
- **Task 3.2**: Verify `.multiselect-trigger` and `.multiselect-dropdown` render cleanly within the waterfall card with appropriate z-index (`z-index: 100`), avoiding clipping inside `.trace-waterfall-card`.
- **Task 3.3**: Ensure typography and padding of multiselect triggers match existing toolbar controls.

### Phase 4: Test Suite Overhaul in TraceWaterfall.test.tsx
- **Task 4.1**: Update existing filter tests to interact with `MultiSelectDropdown` components (clicking trigger combobox, selecting checkboxes) rather than changing HTML select values.
- **Task 4.2**: Add adversarial multi-select tests:
  - Multi-tool selection (e.g. `bash` + `edit`): verify events for both tools appear while other tools and non-tool events are excluded.
  - Multi-anomaly selection (e.g. `TOOL_RETRY_STREAK` + `CONTEXT_SATURATION_SPIKE`): verify events matching either rule appear.
  - Multi-role selection (e.g. `coder` + `tester`): verify events for both roles appear while orchestrator events are excluded.
  - Composite multi-select: combine multiple tools with multiple roles under `errors_anomalies` mode.
- **Task 4.3**: Add explicit "No filter" verification tests:
  - Verify trigger summary displays "No filter" on initial mount for Role, Tool, and Anomaly.
  - Verify unselected "No filter" state allows events without tools (`turn_message`, `compaction`) and events without anomalies to pass through.
  - Verify toggling an option on and then off restores "No filter" state.
- **Task 4.4**: Verify zero emoji invariant across all rendered DOM text.

---

## 5. Verification Checkpoints for Tester Subagent

### Test Suite 1: MultiSelectDropdown Unit Tests (`web/src/components/MultiSelectDropdown.test.tsx`)
- Checkpoint 1.1: Default rendering without `allLabel` maintains placeholder `'All'` and checkbox `"Select all [label]"`.
- Checkpoint 1.2: Passing `allLabel="No filter"` renders trigger summary `"No filter"` when `selectedValues` is `['all']`.
- Checkpoint 1.3: Opening dropdown with `allLabel="No filter"` renders checkbox with accessible label `"No filter"`.
- Checkpoint 1.4: Toggling individual options updates selection and reverting last option restores `['all']` with `"No filter"` summary.

### Test Suite 2: "No filter" Labeling and Pass-Through Verification (`web/src/components/TraceWaterfall.test.tsx`)
- Checkpoint 2.1: On initial mount, Role, Tool, and Anomaly dropdown triggers display `"No filter"`.
- Checkpoint 2.2: In All Events mode with `"No filter"` on all dropdowns, all events (including non-tool and non-anomaly events) are rendered.
- Checkpoint 2.3: Selecting a tool excludes non-tool events; deselecting the tool restores `"No filter"` and non-tool events reappear.

### Test Suite 3: Multi-Selection Filtering Verification
- Checkpoint 3.1: Tool Multi-Selection: Selecting `bash` and `edit` displays only `bash` and `edit` events. Trigger summary displays `"bash, edit"`.
- Checkpoint 3.2: Anomaly Multi-Selection: Selecting two anomaly rules displays events matching either rule.
- Checkpoint 3.3: Role Multi-Selection: Selecting `coder` and `tester` displays events from both roles and excludes orchestrator.
- Checkpoint 3.4: Three or more selections display count format (e.g. `"3 selected"`).

### Test Suite 4: Multi-Dimensional Conjunction Composition
- Checkpoint 4.1: Conjunction of FilterMode `'tools'` + Roles `['coder']` + Tools `['bash', 'edit']` renders only tool calls executed by coder using bash or edit.
- Checkpoint 4.2: Conjunction of FilterMode `'errors_anomalies'` + Anomalies `['TOOL_RETRY_STREAK']` renders only events flagged by `TOOL_RETRY_STREAK`.
- Checkpoint 4.3: Counter reflects exact filtered count (e.g. `"Showing 2 of 10 events"`).
- Checkpoint 4.4: Conflicting filters with zero matches display `"No events match the selected filters."` and counter `"Showing 0 of 10 events"`.

### Test Suite 5: Quality & Regression Gates
- Checkpoint 5.1: `npm run typecheck` passes with zero errors and zero `any`.
- Checkpoint 5.2: `npm test` passes all Vitest test suites across the repository.
- Checkpoint 5.3: `go test -race ./...` passes cleanly with zero failures.
- Checkpoint 5.4: Regex check confirms zero emojis in code and test files.

---

## 6. Audit Checkpoints for Reviewer Subagent

1. Steering Alignment:
   - Does unselected state explicitly display "No filter" across Role, Tool, and Anomaly dropdowns?
   - Can users select multiple tools, multiple anomaly rules, and multiple roles simultaneously?
   - Do events without tools pass through when Tool is "No filter", and do events without anomalies pass through when Anomaly is "No filter"?
2. Architecture & FCIS Boundary Discipline:
   - Are changes strictly confined to presentation components and tests?
   - Does `MultiSelectDropdown` maintain backwards compatibility for `FleetOverview`, `MetricProgression`, and `TicketComparisons`?
3. Typing & Design Invariants:
   - Are all props and state variables strictly typed without `any` or type assertions?
   - Is the zero-emoji invariant strictly maintained across all labels, summaries, and tests?
