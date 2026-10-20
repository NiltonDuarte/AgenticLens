# Implementation Plan: AL-33 Iteration 03 Trace Waterfall Multi-Dimensional Filtering

## 1. Goal and Iteration Scope

The goal of Iteration 03 is to enhance the Execution Trace Waterfall view in the web dashboard by adding multi-dimensional filtering capabilities. Users will be able to filter trace events by specific tool name (for example: task, read, bash, edit) and by specific anomaly rule identifier (for example: TOOL_RETRY_STREAK, CONTEXT_SATURATION_SPIKE, SUBAGENT_REWORK_LOOP, TOOL_HIGH_FAILURE_RATE).

These new filters must seamlessly compose with existing filters (event type FilterMode buttons and roleFilter dropdown), allowing users to pinpoint granular execution anomalies across agent personas and tool invocations.

The scope of this iteration is confined to the frontend presentation layer (web/src/components/TraceWaterfall.tsx, web/src/index.css) and comprehensive automated unit testing (web/src/components/TraceWaterfall.test.tsx). The Go domain core and boundary adapters already provide complete event facts and anomaly classifications, requiring zero backend schema or API modifications.

---

## 2. Grounded Analysis of Existing Architecture

### Existing Filtering Mechanisms in TraceWaterfall.tsx
- State filterMode: Typed as FilterMode ('all' | 'tools' | 'errors_anomalies'), defaulting to 'errors_anomalies'. It toggles between showing all events, only tool_call event types, or only events with status 'error' or matching an anomaly.
- State roleFilter: Typed as string, defaulting to 'all'. Computed availableRoles extracts unique, non-empty role strings from the events prop and sorts them alphabetically. When set to a specific role, non-matching events are discarded.
- Memoized Pipeline:
  - sortedEvents: Orders raw events chronologically by timestamp ascending, then turn ascending.
  - filteredEvents: Evaluates sortedEvents against roleFilter and filterMode.
  - findMatchingAnomaly: Correlates an EventFact against the anomalies prop using tool name, turn number, session ID, and role context.

### Identified Enhancement Opportunities
- Tool Filtering: Events containing tool_name are currently only filtered coarsely via the 'tools' mode button. There is no way to isolate a specific tool (such as bash failures versus read calls).
- Anomaly Filtering: Events matching anomalies are lumped together under the 'errors_anomalies' filter button. There is no mechanism to isolate events flagged by a specific rule (such as focusing strictly on CONTEXT_SATURATION_SPIKE or TOOL_RETRY_STREAK).
- Composition: Users need to drill down along all dimensions at once (for example: coder role + bash tool + errors/anomalies mode).

---

## 3. Detailed Technical Specifications

### 3.1 Tool Filter Specification
1. State Management:
   - State variable: toolFilter.
   - Type: string.
   - Initial value: 'all'.
2. Available Options Derivation (availableTools):
   - Memoized computation using useMemo dependent on events.
   - Iterates through events and extracts e.tool_name where tool_name is non-null, non-undefined, and trimmed to non-empty.
   - Deduplicates names into a unique set and returns an alphabetically sorted array of strings.
3. Filtering Predicate:
   - When toolFilter is 'all', all events pass this condition.
   - When toolFilter is set to a specific tool name, an event matches if and only if e.tool_name is present and exactly matches toolFilter (case-sensitive or normalized).
   - Events without a tool name (such as turn_message or compaction) do not match when a specific toolFilter is active.
4. UI Control:
   - Container class: tool-filter-group.
   - Label: Associated with select input via htmlFor="trace-tool-filter", label text "Tool:".
   - Select element: id="trace-tool-filter", aria-label="Filter by tool", data-testid="trace-tool-filter", class="tool-filter-select".
   - Default option: value="all", text "All Tools".
   - Tool options: Rendered dynamically from availableTools with unique key and matching value and label.

### 3.2 Anomaly Filter Specification
1. State Management:
   - State variable: anomalyFilter.
   - Type: string.
   - Initial value: 'all'.
2. Available Options Derivation (availableAnomalies):
   - Memoized computation using useMemo dependent on anomalies.
   - If anomalies is null, undefined, or empty, evaluates to an empty array.
   - Extracts unique a.rule_id values across all anomalies and returns an alphabetically sorted array of rule ID strings.
3. Filtering Predicate:
   - When anomalyFilter is 'all', all events pass this condition.
   - When anomalyFilter is set to a specific rule ID, the event must correlate with an anomaly whose rule_id matches anomalyFilter.
   - Evaluated by checking if findMatchingAnomaly(e, anomalies) returns an anomaly whose rule_id equals anomalyFilter, or by filtering anomalies to only those matching anomalyFilter before invoking findMatchingAnomaly. This guarantees that events triggering multiple anomaly rules resolve accurately.
4. UI Control:
   - Container class: anomaly-filter-group.
   - Label: Associated with select input via htmlFor="trace-anomaly-filter", label text "Anomaly:".
   - Select element: id="trace-anomaly-filter", aria-label="Filter by anomaly rule", data-testid="trace-anomaly-filter", class="anomaly-filter-select".
   - Default option: value="all", text "All Anomalies".
   - Anomaly options: Rendered dynamically from availableAnomalies with unique key and matching value and label.

### 3.3 Multi-Filter Composition Engine
The filteredEvents memoization in TraceWaterfall.tsx must evaluate every event against all four criteria simultaneously using boolean conjunction (logical AND):
1. Role Filter Criterion:
   - Passes if roleFilter === 'all' OR e.role === roleFilter.
2. Filter Mode Criterion:
   - If filterMode === 'tools', passes if e.event_type === 'tool_call'.
   - If filterMode === 'errors_anomalies', passes if e.status === 'error' OR findMatchingAnomaly(e, anomalies) is truthy.
   - If filterMode === 'all', passes unconditionally.
3. Tool Filter Criterion:
   - Passes if toolFilter === 'all' OR e.tool_name === toolFilter.
4. Anomaly Filter Criterion:
   - Passes if anomalyFilter === 'all' OR findMatchingAnomaly(e, anomalies with rule_id === anomalyFilter) is truthy.

Only events that evaluate to true across all four criteria are included in filteredEvents.
The counter in waterfall-counter automatically displays the number of events satisfying the combined filters out of total events. If zero events match, the existing empty state message ("No events match the selected filters.") is displayed.

### 3.4 Accessibility & UI Design System
- Semantic form elements with explicit htmlFor and id bindings.
- Distinct aria-label attributes on all dropdowns for screen reader clarity.
- Design system consistency:
  - Both tool and anomaly selects must match the existing role-filter-select visual styling.
  - Background: var(--bg-surface-elevated).
  - Border: 1px solid var(--border-color).
  - Text color: var(--text-primary).
  - Padding: 6px 12px.
  - Border radius: 6px.
- Zero Emojis Invariant:
  - Absolute prohibition of emoji characters in labels, options, placeholders, titles, and test fixtures.

---

## 4. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Presentation Logic & State in TraceWaterfall.tsx
- Task 1.1: Add toolFilter state initialized to 'all'.
- Task 1.2: Add anomalyFilter state initialized to 'all'.
- Task 1.3: Compute availableTools via useMemo:
  - Collect unique non-null and non-empty e.tool_name values from events.
  - Sort alphabetically using localeCompare.
- Task 1.4: Compute availableAnomalies via useMemo:
  - Collect unique a.rule_id values from anomalies (handling empty or undefined arrays safely).
  - Sort alphabetically using localeCompare.
- Task 1.5: Update filteredEvents memoized filtering pipeline:
  - Evaluate roleFilter equality.
  - Evaluate filterMode conditions ('all', 'tools', 'errors_anomalies').
  - Evaluate toolFilter equality on e.tool_name.
  - Evaluate anomalyFilter matching against candidate anomalies.
  - Return true only when all four filter conditions are satisfied simultaneously.

### Phase 2: Toolbar UI & Accessible Controls in TraceWaterfall.tsx
- Task 2.1: Inside waterfall-toolbar, directly following the role filter group, add the Tool filter group:
  - Container div with class tool-filter-group.
  - Label with text "Tool:" and htmlFor="trace-tool-filter".
  - Select with id="trace-tool-filter", class="tool-filter-select", aria-label="Filter by tool", data-testid="trace-tool-filter", bound to toolFilter state and onChange handler.
  - Default option with value "all" and text "All Tools".
  - Mapped options from availableTools.
- Task 2.2: Directly following the Tool filter group, add the Anomaly filter group:
  - Container div with class anomaly-filter-group.
  - Label with text "Anomaly:" and htmlFor="trace-anomaly-filter".
  - Select with id="trace-anomaly-filter", class="anomaly-filter-select", aria-label="Filter by anomaly rule", data-testid="trace-anomaly-filter", bound to anomalyFilter state and onChange handler.
  - Default option with value "all" and text "All Anomalies".
  - Mapped options from availableAnomalies.
- Task 2.3: Verify that reset to 'all' on all dropdowns restores the default view without page reload.

### Phase 3: CSS Styling in web/src/index.css
- Task 3.1: Add layout styling for tool-filter-group and anomaly-filter-group matching role-filter-group (display flex, align-items center, gap 8px).
- Task 3.2: Add label styling for tool-filter-label and anomaly-filter-label matching role-filter-label (font-size 0.85rem, color var(--text-secondary)).
- Task 3.3: Add select element styling for tool-filter-select and anomaly-filter-select matching role-filter-select (background-color var(--bg-surface-elevated), border 1px solid var(--border-color), color var(--text-primary), padding 6px 12px, border-radius 6px, cursor pointer).

---

## 5. Verification Checkpoints for Tester Subagent

### Phase 1: Unit & Component Tests in TraceWaterfall.test.tsx
- Test Suite 1: Tool Filter Rendering and Filtering:
  - Verify tool dropdown renders with default "All Tools" option and unique sorted tools present in the events fixture.
  - Verify selecting a specific tool (for example: "bash") displays only events matching tool_name "bash" and hides events for other tools or turn messages.
  - Verify switching tool dropdown back to "All Tools" restores the full list under the active mode.
- Test Suite 2: Anomaly Filter Rendering and Filtering:
  - Verify anomaly dropdown renders with default "All Anomalies" option and unique sorted rule IDs present in the anomalies fixture.
  - Verify selecting a specific anomaly rule (for example: "TOOL_RETRY_STREAK") displays only events correlated with that rule ID.
  - Verify selecting an anomaly rule with no matching events renders the empty state message.
  - Verify switching anomaly dropdown back to "All Anomalies" restores the event list.
- Test Suite 3: Multi-Filter Composition:
  - Activate multiple filters simultaneously (for example: filterMode 'tools' + roleFilter 'coder' + toolFilter 'edit' + anomalyFilter 'TOOL_RETRY_STREAK').
  - Assert that only events satisfying all active criteria are rendered.
  - Assert that the counter updates accurately (for example: "Showing 1 of 4 events").
- Test Suite 4: Edge Cases and Resilience:
  - Provide events with null or undefined tool_name: verify no runtime errors and clean fallback.
  - Provide empty anomalies array: verify anomaly dropdown renders with only "All Anomalies" option.
  - Provide empty events array: verify component gracefully renders the empty trace message.
- Test Suite 5: Accessibility and Zero Emojis:
  - Verify aria-label and testid attributes are present on tool and anomaly selects.
  - Verify regex check confirms zero emoji characters in rendered container text.

### Phase 2: Full Regression & Automated Gate Verification
- Run npm run typecheck to ensure zero TypeScript errors and zero any leaks.
- Run npm test to verify all frontend test suites pass cleanly.
- Run go test -race ./... to ensure Go backend remains completely unaffected and passing.

---

## 6. Audit Checkpoints for Reviewer Subagent

1. Specification Alignment:
   - Does TraceWaterfall allow independent and combined filtering by tool name and anomaly rule ID?
   - Are options dynamically derived from event and anomaly data?
   - Does the event counter reflect the composite filtered set?
2. Architectural Compliance (FCIS & Typing):
   - Are pure functional rollups in domain untouched?
   - Is TypeScript strictly typed with zero any or interface{} assertions?
   - Are optional fields handled defensively without unsafe non-null assertions?
3. Zero Emojis Invariant:
   - Confirm zero emojis across all UI strings, labels, option text, CSS, and test assertions.
4. Clean Code Standards:
   - Proper component decomposition, idiomatic React hooks (useState, useMemo), and semantic HTML layout.
