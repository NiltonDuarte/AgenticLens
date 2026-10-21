# Contract: AL-9 Visual Trace Waterfall and Anomaly Inspector

## 1. Problem Statement

Debugging agent stalls, retry loops, context bloat, and subagent rework currently requires inspecting terminal output, scanning raw JSON ledgers, or reading static markdown tables. While ticket AL-6 established chronological turn-level event extraction in Go (EventType, EventFact) and ticket AL-7 implemented the domain anomaly rule engine (DetectAnomalies), developers and engineering leads have no visual, interactive medium in the web dashboard to inspect session trajectories.

Ticket AL-8 completed the backend HTTP layer by exposing GET /api/workspaces/{ticketId}/events, embedding anomalies in GET /api/workspaces/{ticketId}, and computing fleet-wide anomaly aggregations in GET /api/fleet. 

Ticket AL-9 bridges the presentation layer by delivering:
1. A strictly typed telemetry and anomaly type contract in the TypeScript frontend.
2. An interactive chronological Trace Waterfall component in TicketDeepDive.tsx mapping subagent turns, tool executions, and compactions.
3. Color-coded AnomalyBadge components highlighting anomalous turns and brittle tool executions.
4. An expandable AnomalyDrawer component inspecting tool error tracebacks, execution arguments, token deltas, and rule diagnostics.
5. A graceful legacy v1 fallback banner when viewing older workspace autopsies lacking event streams, while preserving aggregate anomaly visibility.
6. A FleetAnomalyCard component on FleetOverview.tsx presenting fleet-wide anomaly frequencies, severity counts, and brittle tool rankings.
7. An adversarial testing suite verifying timeline rendering, color coding, drawer expansion, edge-case resilience, and zero silent skips.

---

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All frontend and backend components must strictly follow TECHNICAL_GUIDELINES.md:

### Inward Dependency Flow
1. Domain Types and Contracts (web/src/types/telemetry.ts and web/src/domain/types.ts):
   - Pure type definitions, discriminated unions, and value contracts mirroring backend Go domain models.
   - Zero external library dependencies, zero side-effects, and zero network calls.
2. API Boundary Client (web/src/api/client.ts):
   - Handles network transport, JSON deserialization, query serialization, and fallback mapping to mock data.
   - Validates response shapes at boundaries and wraps network errors gracefully.
3. Pure Presentation Components (web/src/components/):
   - Components (TraceWaterfall, AnomalyBadge, AnomalyDrawer, FleetAnomalyCard) receive immutable props and emit user intentions via callbacks.
   - Components contain zero direct HTTP requests, zero file system access, and zero state mutation outside local UI concerns (e.g. drawer open/closed toggle).
4. View Controllers (web/src/views/):
   - TicketDeepDive.tsx and FleetOverview.tsx orchestrate data fetching via web/src/api/client.ts, maintain view-level loading/error states, and wire presentation components together.

### Typing and Boundary Rules
1. Zero Any Mandate:
   - The any type is strictly forbidden across all TypeScript files.
   - All event facts, anomalies, severity classifications, and rule identifiers must be strictly typed.
   - Boundary responses from fetch must be narrowed safely without unsafe casts.
2. Value Semantics and Immutability:
   - Data structures passed across components must use readonly properties and readonly arrays.
3. Defensive Presentation:
   - Presentation components must handle null, undefined, and empty arrays without throwing runtime exceptions.
   - Missing event streams must render clear informational banners rather than crashing or rendering blank space.
4. Single Binary Parity:
   - The TypeScript frontend must build cleanly (npm run build) into web/dist and embed directly into the Go executable via embed.FS with zero Node.js production runtime dependencies.
5. Strict Zero-Emoji Policy:
   - Zero emojis are permitted in code, comments, docstrings, UI labels, tooltips, badges, or workspace artifacts. Textual indicators, color tokens, and SVG icons (from lucide-react) must be used instead.

---

## 3. Scope Boundaries: Go Backend and TypeScript Frontend

### In Scope
1. Telemetry and Anomaly Types (web/src/types/telemetry.ts):
   - Define EventFact, EventType, Anomaly, AnomalySeverity, AnomalyRuleID, AnomalyContext, BrittleToolSummary, and FleetAnomalySummary.
   - Export and align with web/src/domain/types.ts ensuring 100% parity with backend domain structs in backend/internal/domain/models.go and backend/internal/domain/anomalies.go.
2. API Client Extensions (web/src/api/client.ts):
   - Export fetchWorkspaceEvents(ticketId: string) returning ApiResponse of readonly EventFact array.
   - Export fetchWorkspaceAnomalies(ticketId: string) returning ApiResponse of readonly Anomaly array.
   - Update mock data in web/src/api/mockData.ts to provide realistic event sequences, anomaly fixtures, and legacy v1 test cases.
3. Presentation Components (web/src/components/):
   - TraceWaterfall.tsx: Chronological sequence of turns, subagent roles, tool calls, compactions, durations, and anomaly badges.
   - AnomalyBadge.tsx: Amber warning badge and red critical badge with count and label options.
   - AnomalyDrawer.tsx: Slide-out drawer displaying anomaly diagnosis, rule ID, tool error messages, and context metadata.
   - FleetAnomalyCard.tsx: Fleet-level summary card rendering anomaly frequencies, severity counts, rule distributions, and brittle tools table.
4. View Integration:
   - TicketDeepDive.tsx: Integrate TraceWaterfall, AnomalyDrawer, and legacy v1 fallback banner.
   - FleetOverview.tsx: Integrate FleetAnomalyCard into the operational dashboard.
5. CSS and Styling (web/src/index.css):
   - Trace waterfall timeline track, turn nodes, tool execution cards, status indicators, drawer slide-out animations, and anomaly badge styles.
6. Vitest Adversarial Test Suite:
   - Unit tests for TraceWaterfall, AnomalyBadge, AnomalyDrawer, and FleetAnomalyCard.
   - Integration tests in TicketDeepDive.test.tsx and FleetOverview.test.tsx.
   - Regression tests in web/src/api/client.test.ts.

### Out of Scope
1. Go backend domain modifications: backend/internal/domain/anomalies.go, models.go, and aggregations.go are already fully implemented and verified under AL-7 and AL-8.
2. Go HTTP handler modifications: backend/internal/adapters/http/handlers.go and server.go already expose /api/workspaces/{ticketId}/events, /api/workspaces/{ticketId}, and /api/fleet.
3. Real-time WebSocket or SSE streaming: Live trace inspection and live anomaly streaming belong to Phase 4 tickets (AL-17 through AL-20).
4. Markdown document rendering: Markdown contract and summary rendering is already complete under AL-22.

---

## 4. Architectural Specifications and Data Contracts

### Data Contracts Specification

1. EventType Enumeration:
   - String literal union of tool_call, turn_message, compaction.

2. EventFact Interface:
   - timestamp: string (ISO 8601 UTC timestamp).
   - session_id: string.
   - turn: number (1-based integer).
   - role: string (canonical role name: orchestrator, planner, coder, tester, reviewer).
   - event_type: EventType.
   - tool_name: optional string or null.
   - status: string (e.g. success, error).
   - error_message: optional string or null.
   - duration_ms: optional number or null.
   - context_tokens: number (context token count at event time).

3. AnomalySeverity Enumeration:
   - String literal union of WARNING and CRITICAL.

4. AnomalyRuleID Enumeration:
   - String literal union of TOOL_RETRY_STREAK, TOOL_HIGH_FAILURE_RATE, CONTEXT_SATURATION_SPIKE, and SUBAGENT_REWORK_LOOP.

5. AnomalyContext Interface:
   - tool_name: optional string.
   - session_id: optional string.
   - turn: optional number.
   - role: optional string.
   - secondary_role: optional string.
   - count: optional number.
   - total_calls: optional number.
   - token_delta: optional number.
   - context_tokens: optional number.
   - threshold: optional string.
   - failure_rate_pct: optional number.

6. Anomaly Interface:
   - rule_id: AnomalyRuleID.
   - severity: AnomalySeverity.
   - title: string.
   - description: string.
   - context: AnomalyContext.

7. BrittleToolSummary Interface:
   - tool_name: string.
   - anomaly_count: number.
   - total_calls: number.
   - error_count: number.
   - failure_rate_pct: number.

8. FleetAnomalySummary Interface:
   - total_anomalies: number.
   - warning_count: number.
   - critical_count: number.
   - severity_counts: Readonly Record of string to number.
   - rule_counts: Readonly Record of string to number.
   - total_tickets: number.
   - tickets_with_anomalies: number.
   - anomaly_rate_per_ticket: number.
   - brittle_tools: readonly array of BrittleToolSummary.

### Presentation Component Contracts

1. TraceWaterfall Component Contract:
   - Location: web/src/components/TraceWaterfall.tsx.
   - Props Interface:
     - events: readonly array of EventFact.
     - anomalies: optional readonly array of Anomaly.
     - selectedEvent: optional EventFact or null.
     - onSelectEvent: optional callback receiving an EventFact and an optional matching Anomaly.
   - Behavior:
     - Groups or maps events in chronological order.
     - Displays turn number, role badge, timestamp, duration in milliseconds, and status.
     - For tool calls, highlights tool name and indicates success or error.
     - Correlates anomalies to events matching turn, session ID, or tool name, and renders AnomalyBadge.
     - Clicking an event or badge triggers onSelectEvent.
     - Renders empty state when events list is empty.

2. AnomalyBadge Component Contract:
   - Location: web/src/components/AnomalyBadge.tsx.
   - Props Interface:
     - severity: AnomalySeverity.
     - count: optional number.
     - label: optional string.
     - size: optional union of sm, md, lg (defaulting to md).
     - onClick: optional callback function.
   - Behavior:
     - Renders amber/yellow styling for WARNING severity.
     - Renders crimson/red styling for CRITICAL severity.
     - Accessible with ARIA status role or button role when clickable.
     - Displays severity label or count without emojis.

3. AnomalyDrawer Component Contract:
   - Location: web/src/components/AnomalyDrawer.tsx.
   - Props Interface:
     - isOpen: boolean.
     - onClose: callback function.
     - anomaly: optional Anomaly or null.
     - event: optional EventFact or null.
   - Behavior:
     - Slide-out side drawer with backdrop overlay.
     - Displays anomaly title, severity badge, rule identifier, and narrative description.
     - Renders diagnostic context parameters (turn, session, tool name, threshold, token delta).
     - Displays error message or traceback in a formatted monospace card when present.
     - Closes on backdrop click, close button click, or Escape key press.
     - Accessible dialog attributes (role dialog, aria-modal true).

4. FleetAnomalyCard Component Contract:
   - Location: web/src/components/FleetAnomalyCard.tsx.
   - Props Interface:
     - summary: FleetAnomalySummary.
     - onViewTicket: optional callback receiving a ticket ID string.
   - Behavior:
     - Displays high-level anomaly metrics: total anomalies, warning count, critical count, tickets with anomalies ratio, anomaly rate per ticket.
     - Displays rule distribution breakdown across the four anomaly rules.
     - Displays brittle tools table showing tool name, anomaly count, total invocations, errors, and failure rate percentage.
     - Renders positive empty state when total anomalies is zero.

### View Integration Contracts

1. TicketDeepDive.tsx Integration:
   - Fetches workspace detail via fetchWorkspaceById and workspace event stream via fetchWorkspaceEvents.
   - Fallback check: If facts.events is undefined or empty, renders an informative warning banner stating: "Chronological trace waterfall is unavailable for legacy v1 telemetry ledgers. Aggregate workspace anomalies and thread records are displayed below."
   - When events are present, renders TraceWaterfall component above or alongside thread records.
   - Maintains selectedAnomaly and selectedEvent state; renders AnomalyDrawer and opens it when a flagged event is clicked.
2. FleetOverview.tsx Integration:
   - Ingests summary.anomaly_summary from fetchFleetSummary.
   - Renders FleetAnomalyCard in the dashboard layout.

---

## 5. Acceptance Criteria

1. Navigating to a v2 ticket in TicketDeepDive renders the TraceWaterfall component displaying chronological turns and tool calls.
2. Turns or tool calls with detected anomalies display an AnomalyBadge with proper color coding (amber for WARNING, red for CRITICAL).
3. Clicking a flagged turn or anomaly badge opens the AnomalyDrawer displaying rule diagnosis, error details, and execution context.
4. Viewing a legacy v1 ticket lacking an event stream displays an informative legacy banner while continuing to display aggregate anomalies and thread logs.
5. FleetOverview displays the FleetAnomalyCard summarizing fleet anomaly counts, rule distributions, and brittle tools.
6. TypeScript strict compilation passes cleanly with zero type errors and zero any.
7. Automated tests (npm test and npm run build) pass cleanly with zero failures and zero silent skips.
8. Zero emojis throughout all frontend code, components, CSS, tests, and documentation.
