# Implementation Plan: AL-20 Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector

## 1. Goal and Iteration Scope

The goal of iteration 01 is to deliver ticket AL-20: implementing the reactive Live Observability Dashboard and Multi-Root Session Inspector in the TypeScript frontend.

This iteration encompasses expanding application navigation with a dedicated Live Observability tab, creating comprehensive mock fixtures for live trees and traces, implementing a robust reactive SSE streaming client with automatic reconnect and fallback REST polling, building the hierarchical multi-root session tree component, constructing the under-the-hood message trace inspector with collapsible reasoning blocks, developing high-prominence failed tool call alert cards, designing the real-time anomaly notification banner, assembling the main Live Observability view, and verifying the entire feature with exhaustive Vitest component tests while adhering to zero-any and zero-emoji invariants.

## 2. Grounded Reference Architecture & File Paths

The implementation maps directly to the following repository locations:

- Domain Types & Formatters:
  - web/src/domain/types.ts: Extend ActiveTab union to include 'live'.
  - web/src/domain/live.ts: Existing pure domain models (LiveSessionNode, LiveSessionTree, LiveTraceView, LiveMessageFact, LivePartFact, LiveAnomalyAlert).
  - web/src/domain/formatters.ts: Reusable pure formatters (formatTokensK, formatDurationSeconds, formatPercentage, compareTicketIds).
- Client Network & Boundary Layer:
  - web/src/api/liveApi.ts: New SSE streaming client, EventSource lifecycle manager, exponential reconnect backoff, fallback REST polling, and fetchLiveTrace.
  - web/src/api/mockData.ts: New mock fixtures for live session trees, multi-root sessions, traces, reasoning turns, failed tool calls, and anomaly alerts.
  - web/src/api/liveApi.test.ts: Vitest unit tests for SSE handling, error recovery, fallback polling, and trace retrieval.
- UI Presentation Components:
  - web/src/components/Navbar.tsx: Add Live Observability navigation tab button with active state handling.
  - web/src/components/Navbar.test.tsx: Update tests to assert 5 operational tabs including Live Observability.
  - web/src/components/LiveAnomalyBanner.tsx: New real-time alert banner for active session anomalies.
  - web/src/components/LiveAnomalyBanner.test.tsx: Vitest component tests for anomaly rendering, severity color-coding, and click-to-inspect session focus.
  - web/src/components/FailedToolAlertCard.tsx: New prominent error card highlighting failed tool calls, red accent borders, error snippets, and retry counts.
  - web/src/components/FailedToolAlertCard.test.tsx: Vitest component tests for error badge display, monospace log rendering, and expand/collapse behavior.
  - web/src/components/LiveSessionTree.tsx: New multi-root session hierarchy tree with role badges, status pills, duration timers, guessed ticket badges, and anomaly tags.
  - web/src/components/LiveSessionTree.test.tsx: Vitest component tests for multi-root layout, recursive sub-session expansion, candidate badge rendering, and node selection.
  - web/src/components/LiveTraceInspector.tsx: New "Under the Hood" drawer/pane displaying chronological turns, token summaries, collapsible reasoning blocks, and tool parts.
  - web/src/components/LiveTraceInspector.test.tsx: Vitest component tests for message sequencing, reasoning toggle, token display, and failed tool call prioritization.
- Main Dashboard View & Application Wiring:
  - web/src/views/LiveObservability.tsx: New primary view container coordinating streaming state, split layout, inspector drawer, status bar, and mock toggle.
  - web/src/views/LiveObservability.test.tsx: Vitest integration tests for view mounting, connection status transitions, node selection, and fallback mode.
  - web/src/App.tsx: Mount LiveObservability view when activeTab is 'live'.
  - web/src/index.css: Scoped CSS rules for live dashboard layout, session cards, tree connectors, trace drawer, reasoning bubbles, and anomaly banners.
- Verification & Quality Gates:
  - Makefile: Verification via make test-frontend, npm run typecheck, npm run lint, and make build.

## 3. Phased Step-by-Step Implementation Tasks for Coder

### Phase 1: Domain Navigation & Mock Fixtures Scaffolding
- Task 1.1: In web/src/domain/types.ts, extend the ActiveTab type union by appending 'live'. Ensure no existing union members are altered.
- Task 1.2: In web/src/api/mockData.ts, define comprehensive mock live fixtures adhering to the types in web/src/domain/live.ts:
  - Create mockLiveSessionTree containing at least two concurrent root sessions: an orchestrator root session with child sub-sessions (planner, coder, tester) and a separate independent root session (such as product owner ticket inception).
  - Include realistic guessed ticket identifiers (e.g. AL-20 and PO-2) with confidence scores (1.0 and 0.85).
  - Populate child sub-sessions with varied status values (running, stalled, completed), token usages, and attached anomalies.
  - Create mockLiveTraceViews for active sessions containing chronological message turns with both user prompts, assistant turns, token breakdowns, reasoning content, successful tool invocations, and failed tool invocations (such as bash compilation error and missing file error).
  - Create mockLiveAnomalyAlerts covering multiple canonical rules (ConsecutiveToolErrors, ContextCeilingApproach, TokenVelocityRunaway, ToolThrashing) across CRITICAL, HIGH, and WARNING severities.
- Task 1.3: Update web/src/components/Navbar.tsx:
  - Add a button for Live Observability inside the nav-links navigation container.
  - Bind click handler to onSelectTab with argument 'live'.
  - Apply the active CSS class when activeTab is 'live'.
- Task 1.4: Update web/src/components/Navbar.test.tsx:
  - Update tests expecting 4 tabs to assert all 5 operational navigation tabs (Fleet Overview, Ticket Deep Dive, Metric Progression, Comparisons, Live Observability).
  - Add assertion verifying that the Live Observability tab button receives the active class when activeTab is 'live'.
  - Update the zero-emoji test to include 'live' in the array of tested tabs.

### Phase 2: Reactive Streaming and Fallback Polling Client
- Task 2.1: Implement web/src/api/liveApi.ts:
  - Define ConnectionStatus type as union: 'connecting', 'connected', 'fallback-polling', 'paused', 'error'.
  - Define LiveStreamCallbacks interface with callback properties for onSessionTree, onAnomalyAlert, onTraceUpdate, onStatusChange, and onError.
  - Implement subscribeToLiveStream function:
    - Accepts callbacks and optional isMock boolean flag.
    - If isMock is true, dispatches initial mock tree and mock anomalies immediately, starts a mock interval timer simulating periodic heartbeats, and returns a cleanup teardown function.
    - In live mode, initializes native EventSource targeting /api/live/stream.
    - Attaches event listeners for open, error, and named SSE events: session_tree, anomaly_alert, and trace_update.
    - Deserializes event data safely with typed JSON parsing, validating shape before invoking corresponding callbacks.
    - Implements automatic reconnection with exponential backoff if the SSE stream encounters errors or closes unexpectedly.
    - Implements fallback polling mechanism: if consecutive SSE errors exceed threshold or EventSource is unavailable, transitions status to 'fallback-polling' and initiates periodic GET requests to /api/live/sessions every 2000 milliseconds.
    - Returns a cleanup function that closes the EventSource connection, clears any pending reconnect timers, and stops polling intervals.
- Task 2.2: Implement REST fetch helpers in web/src/api/liveApi.ts:
  - Implement fetchLiveSessionTree(isMock?: boolean): Promise of LiveSessionTree. Calls GET /api/live/sessions, returning parsed response or fallback mock tree on failure.
  - Implement fetchLiveTrace(sessionId: string, isMock?: boolean): Promise of LiveTraceView. Calls GET /api/live/sessions/{id}/trace, returning parsed response or fallback mock trace for the session.

### Phase 3: Presentation Component Implementation
- Task 3.1: Implement FailedToolAlertCard in web/src/components/FailedToolAlertCard.tsx:
  - Accepts props: part (LivePartFact), consecutiveErrors (optional number), onInspect (optional callback).
  - Renders a prominent card container with red accent border and elevated dark background.
  - Displays tool name, failure status badge, and severity indicator (CRITICAL, HIGH, MEDIUM, LOW).
  - Displays consecutive error streak count if greater than 1.
  - Renders input snippet and error output snippet in a styled monospace box.
  - Provides an expand/collapse toggle button for long error outputs exceeding 200 characters, defaulting to collapsed view.
- Task 3.2: Implement LiveAnomalyBanner in web/src/components/LiveAnomalyBanner.tsx:
  - Accepts props: anomalies (readonly LiveAnomalyAlert array), onSelectSession (callback taking session ID).
  - If anomalies array is empty, renders null.
  - Renders top-pinned banner container with distinct styling based on highest active severity (red for CRITICAL, orange-red for HIGH, amber for WARNING).
  - Displays count of active anomalies and lists active alerts with rule name, timestamp, and human-readable message.
  - Clicking an anomaly alert invokes onSelectSession with the associated session ID to focus the inspector on that session.
  - Includes a collapse/expand toggle if more than 2 anomalies are present simultaneously.
- Task 3.3: Implement LiveSessionTree in web/src/components/LiveSessionTree.tsx:
  - Accepts props: tree (LiveSessionTree), selectedSessionId (string or null), onSelectSession (callback taking string), onSelectTicket (optional callback taking ticket ID).
  - If root_sessions is empty, renders an empty state card with informative message indicating no active agent sessions.
  - Renders multiple root sessions side-by-side or stacked with distinct visual separation.
  - For each session node, recursively renders:
    - Session header with Agent name, canonical Role badge (orchestrator, planner, coder, tester, reviewer, po), Model identifier, and execution status pill (running, stalled, completed).
    - Session active duration calculated from time_created and time_updated using formatDurationSeconds.
    - Guessed ticket badge: when guessed_ticket_id is present, renders candidate badge showing ticket ID and confidence percentage. Clicking the badge invokes onSelectTicket to navigate to Deep Dive autopsy.
    - Anomaly count indicator badge if the node has attached anomalies.
    - Expand/collapse toggle button for child sub-sessions if sub_sessions array is non-empty.
    - Clicking the session card body invokes onSelectSession to trigger inspection in the trace pane.
    - Visual selection indicator (accent blue border and glow) when node ID matches selectedSessionId.
- Task 3.4: Implement LiveTraceInspector in web/src/components/LiveTraceInspector.tsx:
  - Accepts props: sessionNode (LiveSessionNode or null), trace (LiveTraceView or null), loading (boolean), onClose (callback).
  - If sessionNode is null, renders an empty placeholder pane prompting the user to select an active session from the tree.
  - Renders inspector header displaying selected session ID, agent persona, role badge, total tool calls count, and failed tool calls count.
  - Renders prominent Failed Tools section at the top of the inspector if failed_tool_calls is greater than zero or recent_errors is non-empty, rendering a FailedToolAlertCard for each failed execution.
  - Renders chronological message stream:
    - Renders each LiveMessageFact in chronological order with role indicator (user, assistant, system) and timestamp.
    - Displays token usage metrics (context tokens, input tokens, output tokens, reasoning tokens) formatted via formatTokensK.
    - If reasoning_content is present, renders a distinct collapsible reasoning block with monospace font, subtle border accent, and a toggle button to show or hide the thought process.
    - Renders tool execution parts with tool name, status, input snippet, and output snippet.

### Phase 4: Main Live Observability View & App Wiring
- Task 4.1: Implement LiveObservability view in web/src/views/LiveObservability.tsx:
  - Accepts props: onSelectTicket (callback taking ticket ID), onMockChange (optional callback taking boolean).
  - Manages state: sessionTree (LiveSessionTree), selectedSessionId (string or null), traceView (LiveTraceView or null), connectionStatus (ConnectionStatus), isPaused (boolean), isMock (boolean), errorMessage (string or null).
  - Sets up reactive streaming via subscribeToLiveStream on component mount, passing state setter callbacks.
  - When selectedSessionId changes, calls fetchLiveTrace to load execution trace for the selected session.
  - Renders top status bar displaying:
    - Connection status pill with color indicator (green for connected, amber for fallback-polling, blue for connecting, red for error, gray for paused).
    - Total active sessions counter.
    - Active anomalies counter.
    - Last refreshed timestamp.
    - Action controls: Pause/Resume streaming toggle, Manual Refresh button, and Mock Mode toggle.
  - Renders LiveAnomalyBanner when active_anomalies has elements.
  - Renders split-pane layout:
    - Main/Left column: LiveSessionTree.
    - Right column / Drawer: LiveTraceInspector.
- Task 4.2: Update web/src/App.tsx:
  - Import LiveObservability from ./views/LiveObservability.
  - In main content rendering, add branch: when activeTab is 'live', render LiveObservability, passing onSelectTicket and onMockChange props.

### Phase 5: Styling & Design System Integration
- Task 5.1: In web/src/index.css, add scoped styles for live observability components:
  - Styles for live dashboard container, status header bar, connection pills, and control buttons.
  - Styles for live session tree cards, hierarchical indentation connectors, role badges, and candidate ticket badges.
  - Styles for LiveAnomalyBanner with severity color tiers (CRITICAL in red, HIGH in orange-red, WARNING in amber).
  - Styles for FailedToolAlertCard with prominent red border, error badge, and monospace code blocks.
  - Styles for LiveTraceInspector drawer, message turn items, token chips, and collapsible reasoning containers.
  - Ensure full responsiveness and accessibility across viewports.

## 4. Adversarial Testing Strategy for Tester

### Unit Tests for Streaming & Polling Client (web/src/api/liveApi.test.ts)
- Test Case 1: SSE Subscription Lifecycle
  - Mock native EventSource. Verify that subscribeToLiveStream initializes EventSource with /api/live/stream.
  - Verify that dispatched session_tree event triggers onSessionTree callback with deserialized LiveSessionTree.
  - Verify that dispatched anomaly_alert event triggers onAnomalyAlert callback.
  - Verify that closing the subscription invokes EventSource.close() and clears timers.
- Test Case 2: Automatic Reconnect on Connection Error
  - Simulate EventSource onerror. Verify that status transitions to 'connecting' and reconnection attempt is scheduled with backoff.
- Test Case 3: Fallback Polling Activation
  - Simulate persistent SSE failure. Verify that client falls back to periodic REST polling via /api/live/sessions.
  - Verify that polling response correctly updates tree state.
- Test Case 4: On-Demand Trace Fetching
  - Verify fetchLiveTrace executes GET request to /api/live/sessions/{id}/trace with proper URL encoding and returns parsed LiveTraceView.
  - Verify fetchLiveTrace handles 404 or network error by falling back gracefully without uncaught rejection.
- Test Case 5: Mock Mode Subscription
  - Call subscribeToLiveStream with isMock set to true. Verify that mock tree and mock anomalies are immediately emitted without opening network sockets.

### Component Tests for LiveSessionTree (web/src/components/LiveSessionTree.test.tsx)
- Test Case 6: Empty Session Tree Rendering
  - Render LiveSessionTree with zero root sessions. Assert informative empty state is displayed without errors.
- Test Case 7: Multi-Root Session Rendering
  - Render tree with multiple concurrent root sessions. Assert both root sessions appear with correct agent titles and role badges.
- Test Case 8: Hierarchical Child Sub-Session Expansion
  - Render root session with nested child sub-sessions.
  - Assert expand/collapse button toggles visibility of child nodes.
- Test Case 9: Guessed Ticket Candidate Badge
  - Render node with guessed_ticket_id AL-20 and confidence 1.0.
  - Assert ticket badge renders with candidate label and confidence indicator.
  - Click ticket badge and verify onSelectTicket callback is invoked with AL-20.
- Test Case 10: Session Node Selection
  - Click session card body. Verify onSelectSession callback is invoked with the clicked session ID.
  - Assert selected session card applies visual selected styling.

### Component Tests for LiveTraceInspector (web/src/components/LiveTraceInspector.test.tsx)
- Test Case 11: Empty Selection State
  - Render LiveTraceInspector with sessionNode set to null. Assert placeholder prompt is rendered.
- Test Case 12: Chronological Message Turn Rendering
  - Render trace with multiple message turns. Assert messages render in chronological sequence with role labels and formatted token counts.
- Test Case 13: Collapsible Reasoning Blocks
  - Render message fact containing reasoning_content.
  - Assert reasoning block is rendered with monospace styling.
  - Test clicking "Show Reasoning" / "Hide Reasoning" toggle expands and collapses the content.
- Test Case 14: Failed Tool Call Prioritization
  - Render trace containing failed tool calls.
  - Assert prominent Failed Tools section renders at the top of the inspector above message turns.
  - Assert FailedToolAlertCard is rendered for each failed tool execution.

### Component Tests for FailedToolAlertCard (web/src/components/FailedToolAlertCard.test.tsx)
- Test Case 15: Error Snippet and Severity Rendering
  - Render card with high severity failed bash tool call. Assert tool name, error status, and high severity tag render with red accent styling.
- Test Case 16: Monospace Log Expand/Collapse
  - Render card with long error output exceeding 200 characters.
  - Assert output is initially truncated/clamped. Click expand button and assert full text is revealed.
- Test Case 17: Consecutive Failure Streak Display
  - Render card with consecutiveErrors set to 4. Assert streak counter indicator is prominently displayed.

### Component Tests for LiveAnomalyBanner (web/src/components/LiveAnomalyBanner.test.tsx)
- Test Case 18: Severity-Based Styling
  - Render banner with CRITICAL anomaly. Assert critical styling is applied.
  - Render banner with WARNING anomaly. Assert warning styling is applied.
- Test Case 19: Anomaly Click-to-Inspect Focus
  - Click anomaly alert inside banner. Verify onSelectSession callback is invoked with the alert's session ID.
- Test Case 20: Hidden on Zero Anomalies
  - Render banner with empty anomalies array. Assert component renders null.

### Integration Tests for LiveObservability View (web/src/views/LiveObservability.test.tsx)
- Test Case 21: Mount and Initial Data Load
  - Render LiveObservability. Assert status bar displays connection status, active session count, and session tree.
- Test Case 22: Selecting Session Updates Inspector
  - Select session node from tree. Verify trace inspector updates with the selected session's details and message stream.
- Test Case 23: Pause and Resume Controls
  - Click Pause button. Verify status changes to paused and streaming halts. Click Resume and verify stream reconnects.
- Test Case 24: Mock Mode Toggle
  - Click Mock toggle. Assert mock fixtures load and onMockChange callback is triggered.

### Static Invariant and Quality Gates
- Verify zero any across all new and modified TypeScript files using static analysis and tsc.
- Verify zero emojis across all components, test files, and CSS rules using regex scanning.
- Run make test-frontend, npm run typecheck, and npm run lint.
- Verify single-binary bundling via make build.

## 5. Review Criteria for Reviewer

### Axis 1: Specification and Contract Alignment
- Does the Live Observability view provide real-time monitoring of concurrent root sessions and child sub-sessions without requiring manual page refresh?
- Does the client establish an SSE connection to /api/live/stream and seamlessly fall back to REST polling when SSE fails?
- Are guessed ticket badges displayed beside sessions with confidence scores, allowing one-click navigation to Ticket Deep Dive?
- Does the "Under the Hood" trace inspector render chronological turns with collapsible reasoning blocks?
- Are failed tool calls prominently highlighted in dedicated cards with red accent borders and error logs?
- Does the LiveAnomalyBanner dynamically display active health anomalies categorized by severity?
- Is the new tab integrated into Navbar.tsx and App.tsx under the /live route?

### Axis 2: Engineering Standards and Technical Invariants
- FCIS Separation: Are domain models strictly isolated in web/src/domain/live.ts without DOM or network dependencies? Does network streaming remain isolated in web/src/api/liveApi.ts?
- Strict Typing: Is there zero usage of any or untyped unknown across all TypeScript files? Are discriminated unions used for connection and navigation states?
- Immutability: Are all props and domain data structures typed with readonly fields?
- Resource Safety: Does the streaming client cleanly abort connections and timers upon component unmount to prevent memory leaks?
- Emoji Prohibition: Are all components, labels, badges, tooltips, and tests 100% free of emojis?
- Quality Gates: Do make test-frontend, npm run typecheck, npm run lint, and make build pass cleanly with zero warnings or errors?
