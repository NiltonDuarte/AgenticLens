# Work Summary: AL-20 Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector

## Overview
Ticket AL-20 delivers a reactive Live Observability Dashboard and Multi-Root Session Inspector in the TypeScript frontend. The dashboard connects to the backend live streaming engine via Server-Sent Events (SSE) at `/api/live/stream` and automatically falls back to REST polling at `/api/live/sessions` if streaming encounters persistent errors. It provides multi-root session tree visualization with child sub-session lineages, candidate ticket inference badges with confidence ratings, an "Under the Hood" trace inspector with collapsible deliberation/reasoning blocks, prioritized failed tool execution cards, and real-time anomaly notification banners.

In Iteration 02, runtime mock mode was completely eliminated from the live observability subsystem. The dashboard now interacts strictly with real backend endpoints. When network connectivity fails or endpoints are unreachable, the UI transitions to an explicit error state displaying a prominent warning banner with an actionable "Retry Connection" button, while preserving genuine zero-session empty states when connected to an idle server. Static mock fixtures in `mockData.ts` are retained exclusively for Vitest test assertions.

## Changes by Layer

### Domain Layer (`web/src/domain/`)
- Extended `ActiveTab` type union in `web/src/domain/types.ts` with `'live'`.
- Leveraged pure domain models in `web/src/domain/live.ts` (`LiveSessionNode`, `LiveSessionTree`, `LiveTraceView`, `LiveMessageFact`, `LivePartFact`, `LiveAnomalyAlert`).
- Consumed pure formatters in `web/src/domain/formatters.ts` (`formatTokensK`, `formatDurationSeconds`, `formatPercentage`).

### Boundary Adapters & API Client Layer (`web/src/api/`)
- `web/src/api/liveApi.ts`:
  - Implemented `subscribeToLiveStream` handling SSE events (`session_tree`, `anomaly_alert`, `trace_update`) via native `EventSource`.
  - Added exponential backoff reconnection logic (1s to 16s cap) for resilient socket recovery.
  - Implemented seamless fallback REST polling (`GET /api/live/sessions` every 2000ms) upon repeated SSE connection failures or in restricted environments.
  - Implemented `fetchLiveSessionTree` and `fetchLiveTrace` REST endpoints with defensive payload sanitization and path traversal prevention.
  - Iteration 02: Purged all mock imports (`mockLiveAnomalyAlerts`, `mockLiveSessionTree`, `mockLiveTraceView`, `mockLiveTraceViews`) and excised the `isMock` parameter from all live API functions.
  - Iteration 02: Replaced silent synthetic mock fallbacks with explicit typed `Error` throwing on non-200 HTTP responses and network rejections.
  - Provided clean resource disposal handles (`close()` and timer cancellation) preventing memory leaks.
- `web/src/api/mockData.ts`:
  - Isolated mock live fixtures strictly for Vitest component tests.

### Presentation & Web Dashboard (`web/src/`)
- `web/src/components/Navbar.tsx`:
  - Added "Live Observability" navigation tab button linking to `/live`.
  - Wired active tab state styling.
- `web/src/components/LiveSessionTree.tsx`:
  - Displays concurrent root sessions with agent roles, models, and execution status pills.
  - Recursively renders hierarchical child sub-session lineages with depth indentation and expand/collapse toggles.
  - Displays guessed ticket candidate badges with confidence ratings and direct navigation links to Ticket Deep Dive.
  - Displays attached anomaly indicators and active durations.
- `web/src/components/LiveTraceInspector.tsx`:
  - "Under the Hood" drawer rendering chronological message turns across system, user, and assistant roles.
  - Displays formatted token usage chips (Context, Input, Output, Reasoning, Total).
  - Collapsible deliberation/reasoning containers with monospace formatting and character length threshold toggling.
  - Elevated "Failed Tool Calls" section prioritizing execution errors at the top of the inspector.
- `web/src/components/FailedToolAlertCard.tsx`:
  - High-priority error cards with red accent borders (`var(--status-red)`).
  - Displays tool names, error status badges, severity tags (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`), input snippets, and consecutive error streak counters.
  - Monospace output clamping (>200 chars) with interactive expand/collapse toggle.
- `web/src/components/LiveAnomalyBanner.tsx`:
  - Real-time alert bar dynamically styled according to highest active anomaly severity.
  - Lists anomaly rules, formatted timestamps, and messages with click-to-focus session navigation.
- `web/src/views/LiveObservability.tsx`:
  - Main view container coordinating split-pane layout (tree on left, trace inspector on right).
  - Status header displaying connection status pills, active sessions counter, anomaly counter, and last refreshed timestamp.
  - Interactive controls for pause/resume stream and manual refresh.
  - Iteration 02: Removed runtime mock state and the "Mock Mode" button.
  - Iteration 02: Implemented explicit connection failure banner (`data-testid="live-connection-error-banner"`, `role="alert"`) stating "Unable to connect to live AgenticLens server. Ensure ./bin/agentic-lens server is running." with an active "Retry Connection" button.
  - Preserved genuine empty state card ("No Active Sessions") when connected to an idle backend.
- `web/src/App.tsx`:
  - Wired `LiveObservability` view when `activeTab === 'live'`.
- `web/src/index.css`:
  - Scoped dark-theme styling for live dashboard layouts, session tree connectors, reasoning bubbles, anomaly banners, and live connection error alerts.

## Verification Status
- Frontend Test Suite: All 578 Vitest tests across 28 test files passed cleanly with zero silent skips (`npm test`).
- Backend Test Suite: All 11 Go packages passed with race detection enabled (`go test -race ./...`).
- TypeScript Quality: `npm run typecheck` passed with zero errors and zero `any` usage.
- Linter Quality: `npm run lint` passed with zero warnings and zero errors.
- Repository Checks: `make check` passed with zero git diff anomalies.
- Single-Binary Build: `make` compiled multi-platform binaries with embedded web assets successfully.
- Code Invariants: 100% compliant with FCIS separation, zero `any` or `interface{}`, and zero emojis.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 2
- **Active Agent Duration:** 2h 2m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 163.1k tokens (82% of 200k healthy ceiling, 2 compactions) | Subagents peak 180.4k (Coder)
- **Role Effort Distribution (Subagents):** Planner 16.6%, Coder 41.2%, Tester 30.8%, Reviewer 11.5%
- **Subagent Turn Velocity & Thrashing:** 51.4 avg turns/run (peak: 101 turns) | Planner 38.5, Coder 79.5, Tester 62.5, Reviewer 25.0
- **Test-to-Production Code Ratio:** 0.9:1 (+2,445 test lines / +2,804 prod lines)
- **Code Delivered:** +5,250 / -131 lines across 22 code files (excluding docs)
- **Reasoning Effort:** 146,845 reasoning tokens (45.7% of generated output)
- **Tool Health & Self-Correction:** 0 tool calls (0 self-corrections, 100.0% success rate)

### Anomalies & Health Flags
- **Status:** 15 execution anomalies detected (9 CRITICAL, 6 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.5k tokens) — Context tokens reached 160.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (175.0k tokens) — Context tokens reached 175.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.5k tokens) — Context tokens reached 161.5k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.1k tokens) — Context tokens reached 163.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (180.4k tokens) — Context tokens reached 180.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (175.0k tokens) — Context tokens reached 175.0k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.1k tokens) — Context tokens reached 163.1k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (180.4k tokens) — Context tokens reached 180.4k, breaching the 160k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (175.0k tokens) — Context tokens reached 175.0k, breaching the 160k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+41.8k tokens) — Context tokens surged by 41.8k to 89.5k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+43.7k tokens) — Context tokens surged by 43.7k to 133.2k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+70.7k tokens) — Context tokens surged by 70.7k to 70.7k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+71.4k tokens) — Context tokens surged by 71.4k to 71.4k tokens.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: mcp-gateway_mcp-exec (3 consecutive errors) — Tool 'mcp-gateway_mcp-exec' encountered 3 consecutive execution failures.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: read (3 consecutive errors) — Tool 'read' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **orchestrator** | 1 | 2,412,016 | 41,508 | 8,062,555 | 77.0% | 163.1k | ELEVATED | $2.65 |
| **planner** | 2 | 576,868 | 19,942 | 3,736,960 | 86.6% | 134.5k | ELEVATED | $0.88 |
| **coder** | 2 | 1,083,919 | 65,409 | 12,189,843 | 91.8% | 180.4k | ELEVATED | $2.19 |
| **tester** | 2 | 1,014,990 | 40,363 | 8,233,424 | 89.0% | 175.0k | ELEVATED | $1.64 |
| **reviewer** | 2 | 469,786 | 7,398 | 2,396,706 | 83.6% | 118.9k | ELEVATED | $0.61 |
| **Total** | **9** | **5,557,579** | **174,620** | **34,619,488** | **86.2%** | **180.4k** | **ELEVATED** | **$7.97** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gemini-3.8-flash** | 9 | 5,557,579 | 174,620 | 34,619,488 | 86.2% | $7.97 |
| **Total** | **9** | **5,557,579** | **174,620** | **34,619,488** | **86.2%** | **$7.97** |
