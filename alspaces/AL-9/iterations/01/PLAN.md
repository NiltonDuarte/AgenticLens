# Implementation Plan: AL-9 Visual Trace Waterfall and Anomaly Inspector

## 1. Goal and Iteration Scope

Deliver the full vertical slice for ticket AL-9 in the TypeScript frontend:
1. Define strictly typed domain and telemetry contracts in web/src/types/telemetry.ts and synchronize with web/src/domain/types.ts.
2. Add API client functions in web/src/api/client.ts and update mock datasets in web/src/api/mockData.ts.
3. Build four presentation components in web/src/components/: TraceWaterfall, AnomalyBadge, AnomalyDrawer, and FleetAnomalyCard.
4. Add styling in web/src/index.css matching existing design tokens, dark theme surfaces, and color semantics.
5. Integrate TraceWaterfall, AnomalyDrawer, and the legacy v1 fallback banner into web/src/views/TicketDeepDive.tsx.
6. Integrate FleetAnomalyCard into web/src/views/FleetOverview.tsx.
7. Implement adversarial unit and integration tests across Vitest verifying timeline ordering, badge color coding, drawer expansion, legacy fallback behavior, and zero silent skips.

This plan contains zero code blocks in compliance with the Zero-Code Mandate.

---

## 2. Grounded Reference Architecture and Concrete File Paths

### Types and API Boundary
- web/src/types/telemetry.ts: New file. Author canonical telemetry type definitions (EventFact, EventType, Anomaly, AnomalySeverity, AnomalyRuleID, AnomalyContext, BrittleToolSummary, FleetAnomalySummary).
- web/src/domain/types.ts: Existing file. Re-export and consolidate telemetry definitions from web/src/types/telemetry.ts to eliminate duplication and ensure inward compatibility.
- web/src/api/client.ts: Existing file. Export fetchWorkspaceEvents and fetchWorkspaceAnomalies with typed responses.
- web/src/api/mockData.ts: Existing file. Add mock EventFact sequences for v2 ticket AL-2, keep AL-1 clean as a legacy v1 ticket without events, and ensure mock anomalies are populated.
- web/src/api/client.test.ts: Existing file. Add tests for fetchWorkspaceEvents and fetchWorkspaceAnomalies.

### Presentation Components
- web/src/components/AnomalyBadge.tsx: New file. Pure badge component for severity rendering.
- web/src/components/AnomalyDrawer.tsx: New file. Slide-out inspection drawer for turn anomalies and error payloads.
- web/src/components/TraceWaterfall.tsx: New file. Chronological timeline mapping turns and tool executions.
- web/src/components/FleetAnomalyCard.tsx: New file. Fleet overview card summarizing operational anomalies and brittle tools.
- web/src/index.css: Existing file. Styling rules for waterfall layout, timeline nodes, tool tags, badges, and drawer animations.

### Views and Integration
- web/src/views/TicketDeepDive.tsx: Existing file. Integrate trace waterfall timeline, drawer state, and legacy v1 fallback banner.
- web/src/views/FleetOverview.tsx: Existing file. Integrate fleet anomaly summary card.

### Test Suites
- web/src/components/AnomalyBadge.test.tsx: New file. Unit tests for badge variants, colors, and clicks.
- web/src/components/AnomalyDrawer.test.tsx: New file. Unit tests for drawer rendering, diagnostics, and keyboard escape.
- web/src/components/TraceWaterfall.test.tsx: New file. Unit tests for chronological ordering, tool call display, anomaly badge presence, and empty state.
- web/src/components/FleetAnomalyCard.test.tsx: New file. Unit tests for summary numbers, rule counts, brittle tools table, and zero-anomaly empty state.
- web/src/views/TicketDeepDive.test.tsx: Existing file. Update and expand tests for v2 waterfall rendering, v1 fallback banner, and drawer interactions.
- web/src/views/FleetOverview.test.tsx: Existing file. Update tests asserting FleetAnomalyCard presence and interaction.

---

## 3. Phased Implementation Tasks for Coder Subagent

### Phase 1: Telemetry Types, Boundary Client, and Mock Data

Task 1.1: Author web/src/types/telemetry.ts
- Define EventType as string literal union: tool_call, turn_message, compaction.
- Define EventFact interface with readonly fields: timestamp (string), session_id (string), turn (number), role (string), event_type (EventType), tool_name (optional string or null), status (string), error_message (optional string or null), duration_ms (optional number or null), context_tokens (number).
- Define AnomalySeverity as string literal union: WARNING, CRITICAL.
- Define AnomalyRuleID as string literal union: TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, CONTEXT_SATURATION_SPIKE, SUBAGENT_REWORK_LOOP.
- Define AnomalyContext interface with optional readonly fields: tool_name, session_id, turn, role, secondary_role, count, total_calls, token_delta, context_tokens, threshold, failure_rate_pct.
- Define Anomaly interface with readonly fields: rule_id (AnomalyRuleID), severity (AnomalySeverity), title (string), description (string), context (AnomalyContext).
- Define BrittleToolSummary interface with readonly fields: tool_name (string), anomaly_count (number), total_calls (number), error_count (number), failure_rate_pct (number).
- Define FleetAnomalySummary interface with readonly fields: total_anomalies (number), warning_count (number), critical_count (number), severity_counts (Readonly Record of string to number), rule_counts (Readonly Record of string to number), total_tickets (number), tickets_with_anomalies (number), anomaly_rate_per_ticket (number), brittle_tools (readonly array of BrittleToolSummary).

Task 1.2: Reconcile web/src/domain/types.ts
- Import and re-export telemetry types from web/src/types/telemetry.ts.
- Ensure AgenticMetricsLedger, WorkspaceDetail, and FleetSummary reference the consolidated types.
- Ensure strict type-checking with zero any and zero circular references.

Task 1.3: Update web/src/api/client.ts
- Verify fetchWorkspaceEvents(ticketId: string): Promise of ApiResponse of readonly EventFact array. Ensure it calls GET /api/workspaces/{ticketId}/events and handles 404 and network errors gracefully by falling back to mock events.
- Implement fetchWorkspaceAnomalies(ticketId: string): Promise of ApiResponse of readonly Anomaly array. It calls fetchWorkspaceById(ticketId) and extracts anomalies from the workspace detail response.
- Export all functions with explicit return type signatures.

Task 1.4: Update web/src/api/mockData.ts
- Create a realistic array of mock EventFact objects for workspace AL-2 depicting orchestrator turn start, coder tool executions (including 3 consecutive bash errors triggering TOOL_RETRY_STREAK), and compaction events.
- Assign these mock events to mockAgenticMetricsLedgers['AL-2'].facts.events.
- Ensure workspace AL-1 has facts.events set to undefined or an empty array to serve as the canonical legacy v1 fixture.
- Ensure mockWorkspaceDetails['AL-2'] contains matching anomalies and events.

### Phase 2: Presentation Components and Styling

Task 2.1: Implement AnomalyBadge in web/src/components/AnomalyBadge.tsx
- Props: severity (AnomalySeverity), count (optional number), label (optional string), size (optional sm, md, lg), onClick (optional function).
- Style conditionally: amber background, amber border, and amber text for WARNING; crimson background, red border, and red text for CRITICAL.
- Render text label or count. If count is provided without label, render severity followed by count.
- Add accessible role status or role button if onClick is supplied.
- Zero emojis.

Task 2.2: Implement AnomalyDrawer in web/src/components/AnomalyDrawer.tsx
- Props: isOpen (boolean), onClose (function), anomaly (optional Anomaly or null), event (optional EventFact or null).
- If isOpen is false, return null or render hidden.
- Render semi-transparent backdrop overlay and slide-out panel from the right edge.
- Header: Anomaly title, AnomalyBadge with anomaly severity, and close button with textual Close or multiplication symbol x.
- Body sections:
  - Rule Diagnosis: rule_id, description.
  - Context Metadata Grid: turn number, session ID, tool name, subagent role, failure count, threshold, token delta, context token volume.
  - Error Payload Box: if event.error_message or anomaly context contains error information, render a monospace, border-accented block with error message.
  - Timestamp and execution duration in milliseconds.
- Keyboard handler: listen for keydown event matching Escape to trigger onClose.
- Overlay click handler: clicking outside the drawer content invokes onClose.
- Zero emojis.

Task 2.3: Implement TraceWaterfall in web/src/components/TraceWaterfall.tsx
- Props: events (readonly array of EventFact), anomalies (optional readonly array of Anomaly), selectedEvent (optional EventFact or null), onSelectEvent (optional callback function).
- If events is empty or undefined, render an empty state message: "Zero chronological events recorded for this execution trace."
- Sort events chronologically by timestamp ascending and turn ascending.
- Grouping/Rendering:
  - Timeline vertical rail with connector lines.
  - Each event node displays:
    - Turn pill (e.g. Turn 1, Turn 2).
    - Role pill with distinct color scheme for orchestrator, planner, coder, tester, reviewer.
    - Event type badge (Tool Call, Message, Compaction).
    - Tool details (if tool_call): tool name in monospace, execution status badge (Success in green, Error in red), execution duration formatted as milliseconds or seconds.
    - Context token gauge/tag showing context_tokens.
    - Anomaly correlation: find any anomaly whose context matches event turn and tool name or session ID. If match found, render AnomalyBadge.
  - Selection: clicking the event row or its anomaly badge invokes onSelectEvent with the event and associated anomaly.
  - Highlight selected event with active border and elevated surface background.
- Filter toolbar:
  - Optional quick filter toggles: All Events, Tool Calls Only, Errors and Anomalies Only.
  - Role filter dropdown or pill buttons.

Task 2.4: Implement FleetAnomalyCard in web/src/components/FleetAnomalyCard.tsx
- Props: summary (FleetAnomalySummary), onViewTicket (optional callback receiving ticketId).
- Header: section title "Fleet Anomaly and Reliability Profile" with subtitle detailing operational stability.
- Metric Row:
  - Total Anomalies count with warning and critical breakdown.
  - Affected Tickets metric (tickets_with_anomalies out of total_tickets).
  - Fleet Anomaly Rate per ticket.
- Rule Frequency Grid:
  - Cards or progress meters for each rule: Tool Retry Streak, Tool High Failure Rate, Context Saturation Spike, Subagent Rework Loop.
- Brittle Tools Section:
  - Table displaying brittle tools sorted by anomaly count and failure rate.
  - Columns: Tool Name, Anomalies, Total Calls, Errors, Failure Rate %.
  - If brittle_tools is empty, display "Zero brittle tools identified across the workspace fleet."
- Zero-anomaly state: If summary.total_anomalies is 0, render a clean green-bordered card stating "Fleet Operating Cleanly - Zero Anomalies Detected."
- Zero emojis.

Task 2.5: Add Styling to web/src/index.css
- Define CSS classes for .trace-waterfall-container, .waterfall-timeline, .timeline-item, .timeline-node, .timeline-connector, .timeline-content.
- Define role-specific pill styles (.role-pill.orchestrator, .role-pill.coder, etc.).
- Define event status pill styles (.status-success, .status-error).
- Define .anomaly-badge (.warning, .critical, .sm, .md, .lg).
- Define .anomaly-drawer-backdrop, .anomaly-drawer, .drawer-header, .drawer-body, .drawer-meta-grid, .drawer-error-box.
- Define .fleet-anomaly-card, .rule-frequency-grid, .brittle-tools-table.
- Ensure all styles utilize CSS variables from :root (--bg-surface, --border-color, --status-yellow, --status-red, etc.).

### Phase 3: View Integration

Task 3.1: Integrate TraceWaterfall, AnomalyDrawer, and Legacy Fallback in web/src/views/TicketDeepDive.tsx
- Add component state:
  - events: readonly array of EventFact, initialized to empty.
  - eventsLoading: boolean.
  - selectedEvent: EventFact or null.
  - selectedAnomaly: Anomaly or null.
  - isDrawerOpen: boolean.
- Update data loading effect:
  - Along with fetchWorkspaceById(targetId), call fetchWorkspaceEvents(targetId).
  - Store returned events in state.
- Render Legacy v1 Fallback Banner:
  - Condition: if (!events || events.length === 0) and workspace detail is loaded.
  - Banner markup: warning-styled notification card with clear text:
    "Chronological trace waterfall is unavailable for legacy v1 telemetry ledgers. Aggregate workspace anomalies and thread records are displayed below."
  - When anomalies exist on the workspace (detail.anomalies.length > 0), display an aggregate anomalies summary list inside the banner or directly beneath it so anomaly visibility is never lost.
- Render TraceWaterfall:
  - Condition: when events and events.length > 0.
  - Pass events, detail.anomalies, selectedEvent, and onSelectEvent handler.
- Wire AnomalyDrawer:
  - Render AnomalyDrawer passing isOpen={isDrawerOpen}, onClose={() => setIsDrawerOpen(false)}, anomaly={selectedAnomaly}, event={selectedEvent}.
  - onSelectEvent handler sets selectedEvent, selectedAnomaly (matching from detail.anomalies if available), and sets isDrawerOpen to true.

Task 3.2: Integrate FleetAnomalyCard in web/src/views/FleetOverview.tsx
- Import FleetAnomalyCard.
- In FleetOverview render method, position FleetAnomalyCard below the KPI row and above the distribution cards / ticket table.
- Pass summary.anomaly_summary and onSelectTicket.

---

## 4. Adversarial Testing Plan for Tester Subagent

The tester subagent must execute adversarial tests across all new and updated components with zero silent skips (test.skip and it.skip are forbidden).

### Test Suite 1: web/src/components/AnomalyBadge.test.tsx
1. Renders WARNING severity with amber styling classes and accessible role status.
2. Renders CRITICAL severity with red styling classes and accessible role status.
3. Renders count accurately when count prop is provided.
4. Renders custom label string when label prop is provided.
5. Invokes onClick callback when clicked, and verifies role is button when onClick is provided.
6. Renders small, medium, and large sizes according to size prop.

### Test Suite 2: web/src/components/AnomalyDrawer.test.tsx
1. Returns null or remains hidden when isOpen is false.
2. Renders title, severity badge, rule ID, and description when isOpen is true and anomaly is provided.
3. Renders diagnostic context fields: turn number, session ID, tool name, threshold, and token delta.
4. Renders error message container in monospace font when event.error_message is present.
5. Invokes onClose when close button is clicked.
6. Invokes onClose when backdrop overlay is clicked.
7. Invokes onClose when Escape key is pressed.
8. Handles missing anomaly context or missing event gracefully without runtime errors.

### Test Suite 3: web/src/components/TraceWaterfall.test.tsx
1. Renders empty message when events array is empty.
2. Renders chronological sequence of events sorted by timestamp and turn.
3. Displays role pills with proper role names for orchestrator, coder, tester, etc.
4. For tool calls, displays tool name, formatted duration in milliseconds, and status indicator.
5. Correlates anomaly by turn/tool name and renders AnomalyBadge on the corresponding turn item.
6. Invokes onSelectEvent callback with event and anomaly when a timeline item or badge is clicked.
7. Filters events by role when role filter is changed.
8. Filters events to only errors and anomalies when error filter toggle is activated.
9. Highlights the active event when selectedEvent prop matches an event in the list.

### Test Suite 4: web/src/components/FleetAnomalyCard.test.tsx
1. Displays total anomalies, warning count, critical count, and anomaly rate per ticket.
2. Displays affected tickets ratio (e.g. 2 of 4 tickets).
3. Renders rule count values for all four anomaly rules.
4. Renders brittle tools table with tool names, anomaly counts, total calls, and formatted failure rate percentages.
5. Handles division by zero defensively when a brittle tool has zero calls.
6. Displays positive empty state message when total_anomalies is zero.
7. Handles empty brittle_tools array without table errors.

### Test Suite 5: web/src/views/TicketDeepDive.test.tsx (Integration)
1. Renders legacy v1 fallback banner when viewing a ticket whose events array is empty (e.g. AL-1).
2. Verifies that aggregate anomalies for v1 tickets are still displayed even when waterfall timeline is unavailable.
3. Renders TraceWaterfall component when viewing a v2 ticket with populated events (e.g. AL-2).
4. Verifies that clicking an anomalous event in the waterfall opens the AnomalyDrawer with diagnostic details.
5. Verifies closing the AnomalyDrawer restores normal view state.

### Test Suite 6: web/src/views/FleetOverview.test.tsx (Integration)
1. Renders FleetAnomalyCard in the fleet overview layout.
2. Displays fleet anomaly metrics from the API response.
3. Navigates or calls ticket selection when a ticket row or brittle tool is selected.

### Test Suite 7: web/src/api/client.test.ts (API Boundary)
1. fetchWorkspaceEvents calls GET /api/workspaces/{ticketId}/events and returns data with isMock false when fetch succeeds.
2. fetchWorkspaceEvents falls back to mock events when fetch fails.
3. fetchWorkspaceAnomalies returns anomalies from workspace detail.
4. fetchWorkspaceEvents throws or falls back appropriately for non-existent tickets.

### Boundary and Quality Gates
- Execute npm run typecheck: must pass with zero type errors and zero any.
- Execute npm test: all Vitest suites must pass with zero skipped tests.
- Execute npm run build: frontend bundles into web/dist without errors.
- Execute go vet ./... and go test -race ./...: ensure full single-binary embedding and backend integrity.

---

## 5. Reviewer Audit Checklist

The reviewer subagent must audit all changes along two strict axes:

### Axis 1: Spec and Contract Alignment
- Trace waterfall maps turns and tool calls chronologically.
- Visual anomaly badges indicate amber warning and red critical severities.
- Anomaly drawer opens upon clicking flagged turns and displays error tracebacks, argument payloads, and token deltas.
- Legacy v1 workspaces lacking event streams display an informative fallback banner while preserving aggregate anomaly visibility.
- FleetOverview displays the FleetAnomalyCard summary.
- Zero silent skips across all test suites.

### Axis 2: Engineering Standards (TECHNICAL_GUIDELINES.md)
- Functional Core, Imperative Shell separation respected: presentation components are pure, network fetching is isolated to client.ts and views.
- Strict typing: zero any across all TypeScript files.
- Immutability: readonly modifiers on interface fields and arrays.
- Single binary asset bundling: npm run build populates web/dist for embed.FS in Go backend.
- Zero emojis across all code files, comments, docstrings, JSX elements, CSS, and test assertions.
