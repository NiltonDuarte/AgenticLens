# feat(live,web): build typescript live observability dashboard and multiroot session inspector (AL-20)

## Summary of Changes

Delivers a reactive Live Observability Dashboard and Multi-Root Session Inspector in the TypeScript web frontend for ticket AL-20:

1. **Reactive Streaming & Boundary Client**:
   - Implemented `web/src/api/liveApi.ts` managing real-time Server-Sent Events (SSE) connections at `/api/live/stream` with automatic exponential backoff reconnection.
   - Built seamless fallback REST polling (`GET /api/live/sessions` every 2000ms) for environments where SSE connections drop or are unsupported.
   - Implemented `fetchLiveTrace` for on-demand session execution trace retrieval with path traversal validation.
   - Provided deterministic mock fixtures in `web/src/api/mockData.ts` for offline demo environments and automated testing.

2. **Multi-Root Session Tree & Inferred Ticket Navigation**:
   - Built `LiveSessionTree.tsx` displaying concurrent root sessions with agent roles, models, active durations, and status pills.
   - Implemented recursive child sub-session lineage rendering with depth indentation and expand/collapse toggles.
   - Added candidate ticket inference badges with confidence indicators, enabling direct one-click navigation to Ticket Deep Dive autopsy views.

3. **Under-the-Hood Trace Inspector & Reasoning Stream**:
   - Built `LiveTraceInspector.tsx` rendering chronological turns across user, assistant, and system roles with token usage chips.
   - Added collapsible deliberation/reasoning blocks in monospace format with show/hide toggle buttons.
   - Elevated failed tool calls into a dedicated high-priority section at the top of the inspector.

4. **Prominent Tool Failure Alerting & Live Anomaly Banners**:
   - Built `FailedToolAlertCard.tsx` with red accent borders (`var(--status-red)`), failure badges, severity tags, monospace log clamping, and consecutive error streak counters.
   - Built `LiveAnomalyBanner.tsx` dynamically styled by highest active severity (CRITICAL, HIGH, WARNING) with click-to-focus session navigation.

5. **Application Wiring & Design System**:
   - Extended `ActiveTab` with `'live'` in `web/src/domain/types.ts`.
   - Added "Live Observability" navigation tab to `Navbar.tsx` and routed view in `App.tsx`.
   - Added scoped dark-mode styling in `web/src/index.css`.

## Architectural Invariants
- FCIS compliance: pure domain models isolated from imperative browser I/O.
- Strict typing: zero `any` or `interface{}` across all TypeScript modules.
- Resource safety: complete teardown of SSE sockets and interval timers on unmount.
- Design standards: zero emojis across code, tests, and UI components.

## Documentation References
- Contract: [CONTRACT.md](workspaces/AL-20/CONTRACT.md)
- Work Summary: [SUMMARY.md](workspaces/AL-20/SUMMARY.md)
