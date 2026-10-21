# Contract: AL-20 Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector

## 1. Problem Statement

With backend live streaming and non-locking endpoints delivered in AL-17 and AL-19, the AgenticLens Go HTTP server provides real-time Server-Sent Events (SSE) at `/api/live/stream`, snapshot hierarchical session trees at `/api/live/sessions`, and chronological execution traces at `/api/live/sessions/{id}/trace`. However, the frontend presentation layer currently lacks a dedicated Live Observability view. Users can only observe finished sessions via post-mortem autopsy reports in Fleet Overview and Ticket Deep Dive.

During autonomous multi-agent execution, operators need real-time situational awareness:
1. Multi-Root Concurrency: Agents execute concurrently across disparate worktrees, tasks, and roles. Operators must follow multiple active root sessions side-by-side without page reloads.
2. Hierarchical Traceability: Orchestrator sessions delegate to specialized child subagents (planner, coder, tester, reviewer). Operators must see nested execution parent-child lineages with agent roles, active durations, and inferred ticket candidates.
3. Under-the-Hood Inspection: Operators require immediate access to chronological turn messages, token accumulations, and expandable thinking/reasoning blocks to understand agent deliberations.
4. Tool Failure Visibility: Failed tool executions (compilation faults, path traversal traps, nonzero exit codes) are easily lost in high-frequency message streams. They must be prominently extracted into dedicated high-priority cards with red accent borders and error logs.
5. Real-Time Anomaly Notifications: Health degradations flagged by the Go live anomaly engine (consecutive tool errors, context window ceiling approach, token velocity runaway, tool thrashing) must dynamically surface in a top-level alert banner without user polling.

Building a reactive Live Observability dashboard with an SSE streaming client, resilient fallback polling, multi-root session tree, under-the-hood message inspector, prominent failed tool alert cards, and live anomaly banners fulfills this operational gap.

## 2. Functional Core, Imperative Shell (FCIS) Invariants

All implementation must strictly comply with TECHNICAL_GUIDELINES.md, AGENTS.md, and project invariants.

### Inward Dependency Flow
Dependencies flow strictly inward across layers:
1. Functional Core Domain Layer (`web/src/domain/live.ts` and `web/src/domain/formatters.ts`):
   - Pure domain models (`LiveSessionNode`, `LiveSessionTree`, `LiveTraceView`, `LiveMessageFact`, `LivePartFact`, `LiveAnomalyAlert`).
   - Pure formatting and mathematical calculations (`formatTokensK`, `formatDurationSeconds`, `formatPercentage`, `compareTicketIds`).
   - Absolute isolation: zero DOM APIs, zero React hooks, zero network calls (`fetch`, `EventSource`), and zero side-effects.
2. Imperative Network & Streaming Adapter Layer (`web/src/api/liveApi.ts`):
   - Manages external browser I/O: SSE connection lifecycle via `EventSource`, event listeners, retry timers, network reconnects, and fallback REST polling (`/api/live/sessions`).
   - Translates raw boundary JSON into validated domain models.
   - Provides deterministic mock fallback fixtures in `web/src/api/mockData.ts` when running offline or in mock mode.
   - Exposes clean subscription and unsubscription handles to prevent memory leaks and orphaned network connections.
3. UI Presentation Layer (`web/src/views/LiveObservability.tsx` and `web/src/components/`):
   - Declarative React components consuming immutable domain objects passed via props.
   - Pure presentation logic separated from network transport.

### Strict Typing and Zero Any Invariant
1. Zero usage of `any` or untyped `unknown` across all TypeScript modules.
2. Discriminated unions for connection status: `'connecting' | 'connected' | 'fallback-polling' | 'paused' | 'error'`.
3. Discriminated unions for tab navigation: `'landing' | 'overview' | 'deep-dive' | 'progression' | 'comparisons' | 'live'`.
4. String literals for anomaly severity: `'WARNING' | 'HIGH' | 'CRITICAL'`.
5. Readonly object fields and readonly arrays across all live interfaces.

### Zero Emojis Mandate
No emojis permitted in source code, component templates, CSS class names, test assertions, button labels, badges, tooltips, or log messages.

## 3. Scope Boundaries

### In Scope for AL-20
1. Application Routing & Navigation:
   - Extend `ActiveTab` type in `web/src/domain/types.ts` to include `'live'`.
   - Add "Live Observability" nav-tab button in `web/src/components/Navbar.tsx`.
   - Mount `LiveObservability` view in `web/src/App.tsx`.
   - Support navigation from guessed ticket badges directly into `TicketDeepDive`.
2. Reactive Live Streaming & Polling Client (`web/src/api/liveApi.ts`):
   - Connection to `/api/live/stream` via `EventSource`.
   - Routing of incoming SSE events: `session_tree`, `anomaly_alert`, and `trace_update`.
   - Resilient automatic reconnection with exponential backoff on connection termination.
   - Graceful fallback polling: if `EventSource` fails or is unsupported, automatically fall back to polling `/api/live/sessions` at the configured interval (2000ms).
   - On-demand trace retrieval: `fetchLiveTrace(sessionId)` fetching from `/api/live/sessions/{id}/trace`.
   - Comprehensive mock fallback fixtures in `web/src/api/mockData.ts` for mock mode and testing.
   - Clean resource disposal: `close()` and unsubscribe callback aborting active connections and clearing interval timers.
3. Presentation Components:
   - Main View (`web/src/views/LiveObservability.tsx`): split-pane layout with status bar (connection pill, active sessions count, anomaly count, last refreshed time), pause/resume controls, and mock mode toggle.
   - Session Tree (`web/src/components/LiveSessionTree.tsx`): displays multiple active root sessions side-by-side or stacked, expandable child sub-sessions with agent roles, models, active durations, guessed ticket badges, and anomaly tags.
   - Trace Inspector (`web/src/components/LiveTraceInspector.tsx`): "Under the Hood" drawer/pane displaying chronological message turns, token usages, collapsible thinking/reasoning blocks, and tool executions.
   - Failed Tool Alert Card (`web/src/components/FailedToolAlertCard.tsx`): high-prominence error card with red accent borders, error status badge, severity tags, error logs, and consecutive failure counts.
   - Anomaly Alert Banner (`web/src/components/LiveAnomalyBanner.tsx`): real-time alert bar with severity badges (CRITICAL, HIGH, WARNING), rule names, messages, timestamps, and click-to-inspect session focus.
4. Styling:
   - Modern dark-theme CSS in `web/src/index.css` maintaining visual consistency with existing AgenticLens design tokens.
5. Automated Test Suites:
   - Vitest component and unit tests covering `liveApi.ts`, `LiveSessionTree`, `LiveTraceInspector`, `FailedToolAlertCard`, `LiveAnomalyBanner`, `LiveObservability`, and updated `Navbar`.

### Out of Scope for AL-20
1. Modifying Go backend domain models or HTTP handlers (completed in AL-17 and AL-19).
2. Direct SQLite database mutations or query schema alterations.
3. Historical autopsy telemetry backfill (governed by AL-12 and AL-32).

## 4. Architectural Contracts and Component Hierarchy

### Component Hierarchy
- `App` (`web/src/App.tsx`)
  - `Navbar` (`web/src/components/Navbar.tsx`) [Includes 'Live Observability' tab]
  - `LiveObservability` (`web/src/views/LiveObservability.tsx`) [Active when activeTab === 'live']
    - Status Bar [Connection status pill, active count, anomaly count, last refreshed timestamp, pause/resume, mock toggle]
    - `LiveAnomalyBanner` (`web/src/components/LiveAnomalyBanner.tsx`) [Pinned at top when active anomalies exist]
    - Split Layout:
      - Left Pane: `LiveSessionTree` (`web/src/components/LiveSessionTree.tsx`)
        - Multi-root session container
        - Expandable sub-session nodes
        - Guessed ticket badges
        - Role pills and duration indicators
      - Right Pane: `LiveTraceInspector` (`web/src/components/LiveTraceInspector.tsx`)
        - Session header metadata
        - Failed Tool Calls section rendering `FailedToolAlertCard` (`web/src/components/FailedToolAlertCard.tsx`)
        - Chronological message turn stream with collapsible reasoning blocks
        - Tool execution parts

### Client Streaming Contract (`web/src/api/liveApi.ts`)
1. Types:
   - `ConnectionStatus = 'connecting' | 'connected' | 'fallback-polling' | 'paused' | 'error'`
   - `LiveStreamCallbacks`:
     - `onSessionTree: (tree: LiveSessionNode[]) => void`
     - `onAnomalyAlert: (alert: LiveAnomalyAlert) => void`
     - `onStatusChange: (status: ConnectionStatus) => void`
     - `onError: (error: Error) => void`
2. Functions:
   - `subscribeToLiveStream(callbacks: LiveStreamCallbacks, isMock?: boolean): () => void`
   - `fetchLiveSessionTree(isMock?: boolean): Promise<LiveSessionTree>`
   - `fetchLiveTrace(sessionId: string, isMock?: boolean): Promise<LiveTraceView>`

### Guessed Ticket Badge Contract
When `LiveSessionNode.guessed_ticket_id` is populated:
- Display ticket badge with label matching the candidate identifier (e.g. `AL-20`).
- Display candidate confidence percentage indicator (e.g. `100% confidence`).
- Provide click handler navigating directly to the ticket's Deep Dive autopsy view via `onSelectTicket(ticketId)`.

### Failed Tool Alert Contract (`FailedToolAlertCard.tsx`)
When a tool part has `is_error === true`:
- Card rendered with high visual priority: red border accent (`var(--status-red)`), dark elevated surface, and error badge.
- Severity tag: `CRITICAL` or `HIGH` (bright red), `MEDIUM` (amber), `LOW` (subtle red/gray).
- Monospace box for error output snippet, capped with expand/collapse toggle for outputs exceeding 200 characters.
- Display tool name, execution status, and consecutive error streak count when consecutive errors are present.

### Collapsible Reasoning Contract
When `LiveMessageFact.reasoning_content` is present:
- Rendered in a distinct thought bubble container with subtle purple or blue borders.
- Collapsed by default if length exceeds 300 characters, showing first 2 lines with a toggle button labeled "Show Reasoning" / "Hide Reasoning".
- Monospace font style to distinguish model internal reasoning from conversational responses.

## 5. Backward Compatibility & Resilience Guarantees

1. Graceful Connection Degradation:
   If the browser environment fails to establish an SSE connection or receives HTTP 404/500 from `/api/live/stream`, the client smoothly switches to polling mode via `GET /api/live/sessions` without throwing unhandled exceptions or disrupting the user experience.
2. Offline / Mock Mode Parity:
   When `isMock` is toggled or when running in local offline demo environments without a live Go backend, `liveApi.ts` yields deterministic mock multi-root trees and trace records from `mockData.ts`.
3. Navigation Safety:
   Switching between `/live` and other dashboard views cleanly tears down the `EventSource` connection and clears background timers, preventing memory leaks and orphaned network requests.

## 6. Acceptance Criteria

1. Route `/live` is registered and accessible via the "Live Observability" navigation tab in `Navbar.tsx`.
2. The Live Observability view establishes an SSE stream with `/api/live/stream` and dynamically updates when new sessions or events arrive.
3. Fallback polling activates seamlessly if SSE connection cannot be established or encounters sustained network errors.
4. Multiple concurrent root sessions are rendered side-by-side or stacked, with nested sub-sessions expandable in a tree hierarchy.
5. Inferred ticket candidate badges appear beside active sessions with candidate confidence percentages, supporting one-click navigation to Deep Dive.
6. Clicking any session node opens the "Under the Hood" Trace Inspector pane showing chronological turns, message metadata, and collapsible reasoning blocks.
7. Failed tool executions are highlighted in high-prominence `FailedToolAlertCard` components with red accent borders and expandable error details.
8. Active anomaly alerts trigger immediate updates to `LiveAnomalyBanner` without requiring manual page refresh.
9. All frontend unit and component tests pass cleanly via `npm test`.
10. Strict type-checking succeeds via `npm run typecheck` with zero `any`.
11. Linter passes cleanly via `npm run lint` with zero warnings.
12. Single-binary build succeeds via `make build`.
13. Zero emojis exist in code, comments, logs, or UI strings.
