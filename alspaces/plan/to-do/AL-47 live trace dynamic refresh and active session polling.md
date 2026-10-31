h3. Title: AL-47: Real-Time Live Trace Dynamic Refresh and Active Session Polling

*TARGET*: AgenticLens (backend / live / http / presentation / web / tests)

*CONTEXT*
 
In Live Observability (`div#root > div.app-container > main.main-content > div.live-observability-view > div.live-split-layout > div.live-split-pane.right-inspector-pane`), the trace inspector does not update when an active agent session executes new turns, tool invocations, or deliberations.
Currently, `LiveTraceHandler` serves `/api/live/sessions/:id/trace` on demand, but `LiveObservability.tsx` only calls `fetchLiveTrace` once upon initial node selection or manual full-page refresh. While the SSE connection broadcasts `session_tree` when session nodes update, the backend never emits `trace_update`, and the frontend does not re-fetch the active session trace when the session receives new events. Operators watching an ongoing run are left staring at a static snapshot unless they refresh the browser.

*GOAL*
 
Enable seamless, real-time live updates in the right inspector pane for actively running sessions, so that incoming message turns, model reasoning, token counts, and tool errors stream dynamically into the view without requiring a manual page refresh.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Pure comparison logic determining if a trace view has advanced (e.g. message count increased, token totals changed, or newest timestamp advanced) to avoid unnecessary re-renders.
* Boundary Adapter:
  - In `backend/internal/adapters/live/service.go`: track session trace revisions or updated timestamps. When polling SQLite hook, broadcast `LiveEventTraceUpdate` when an active session's trace content changes, or ensure session nodes expose an updated revision timestamp.
  - In `backend/internal/adapters/http/live_handlers.go`: verify SSE streaming flushes trace events correctly without blocking subscribers.
* Presentation:
  - In `web/src/views/LiveObservability.tsx`:
    - Listen for `onTraceUpdate` in SSE stream to update `traceView` when `trace.session_id === selectedSessionId`.
    - As a robust dual-plane guarantee: when `selectedSessionId` corresponds to an active session (`status === 'running'` or `status === 'active'`), automatically re-fetch `fetchLiveTrace(selectedSessionId)` on incoming `session_tree` updates or on a 2000ms active interval if the session node's `time_updated` has changed.
    - Preserve user scroll position and reasoning toggle states (`showReasoning`) during incremental trace updates so active reading is not interrupted.
* Quality:
  - Backend integration test in `live_handlers_test.go` verifying trace update emission or on-demand retrieval.
  - TypeScript integration tests in `LiveObservability.test.tsx` verifying that incoming SSE trace updates or active polling dynamically update the inspector messages list without full component remount.

*ACCEPTANCE CRITERIA*
 
* When an active session is selected in `LiveObservability`, new turns and tool calls appear in the right inspector pane automatically without user intervention or manual page reload.
* Token metrics (total, context, input, output, reasoning) and failed tool count update in real time.
* If a session has completed (`status === 'completed'`), automatic polling for that session ceases to prevent unnecessary network or database traffic.
* Incremental trace updates do not reset or jump the user's scroll position or collapse open reasoning bubbles.
* SSE disconnect/fallback-polling cleanly continues updating active traces.
* All Go tests (`go test ./...`) and TypeScript tests (`npm test`) pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.

*PLANNING STATUS (2026-10-31)*

Still open. The current backend emits `trace_update` and `LiveObservability.tsx` already handles `onTraceUpdate`, so the context above describes the original report rather than the entire current implementation. AL-49 supplies inspector scroll preservation. Before implementation, verify the remaining active-session refresh, completed-session cessation, SSE fallback, and reading-state acceptance cases; existing streaming alone does not establish completion.
