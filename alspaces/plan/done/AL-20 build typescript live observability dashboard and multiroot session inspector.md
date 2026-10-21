h3. Title: AL-20: Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector

*TARGET*: AgenticLens (presentation | web | tests)

*CONTEXT*

With backend live streaming and non-locking endpoints available, developers require an intuitive, real-time UI dashboard to monitor active agent execution. Users must be able to follow multiple concurrent root sessions and their child sub-sessions, view guessed ticket identifiers, inspect underlying messages and reasoning blocks, review prominent failed tool calls, and observe live anomaly alerts without page reloads.

*GOAL*

Build a reactive Live Observability view in the TypeScript SPA with real-time SSE streaming, presenting a hierarchical multi-root session tree, guessed ticket badges, an under-the-hood message and reasoning inspector, prominent failed tool call alerting cards, and live anomaly notification banners.

*VERTICAL SLICE SCOPE*

* Presentation & Web Components:
  - Add `/live` route and `LiveObservability.tsx` view with live navigation tab in `Navbar.tsx`.
  - Create `LiveSessionTree.tsx`: displays multiple active root sessions side-by-side or as an expandable tree, listing child sub-sessions with agent roles, active duration, and guessed ticket badges.
  - Create `LiveTraceInspector.tsx`: "Under the Hood" drawer/pane displaying chronological inter-agent messages, expandable thinking/reasoning blocks, and tool executions.
  - Create `FailedToolAlertCard.tsx`: high-prominence error card highlighting failed tool calls, error stack/output snippets, and retry counts.
  - Create `LiveAnomalyBanner.tsx`: real-time alert bar indicating active session anomalies (consecutive failures, context ceiling warnings, looping).
* Client State & Streaming:
  - Add `liveApi.ts` client connecting to `/api/live/stream` with automatic reconnect and fallback polling.
  - Strict TypeScript types (`LiveSessionNode`, `LiveTraceView`, `LiveAnomalyAlert`) with zero `any`.
* Quality:
  - Vitest unit and component tests for session tree rendering, ticket guessing badge display, reasoning collapse/expand, and failed tool call prioritization.
  - Accessibility and responsiveness checks.

*ACCEPTANCE CRITERIA*

* The Live Observability dashboard connects to the live stream and renders active root sessions and nested sub-sessions.
* Guessed ticket identifiers are displayed alongside each active session with a clear candidate indicator.
* Clicking a session opens the "Under the Hood" inspector showing inter-agent messages and collapsible reasoning blocks.
* Failed tool calls are visually emphasized with high-importance error badges, red accent borders, and expandable error logs.
* Active anomaly alerts appear dynamically when triggered by the backend without requiring manual refresh.
* All frontend tests pass (`npm test`) and type checking succeeds with zero `any`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make test-frontend`, `npm run typecheck`, `npm run lint`).
* Zero any and zero emojis across all frontend files.
* Single-binary embedding verified with `make build`.
