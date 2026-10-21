h3. Title: AL-3: Build TypeScript SPA Shell and Fleet Overview Dashboard

*TARGET*: AgenticLens (presentation | web | tests)

*CONTEXT*

The Python Streamlit app provides a Fleet Overview showing KPI metric cards and role/model cost distributions. We need the TypeScript web frontend application reproducing this overview with fast rendering and strict typing.

*GOAL*

Scaffold the `web/` application using Vite, TypeScript, and a modern component architecture, providing API boundary types, a dual-mode network client (live `/api/*` with mock fixture fallback), and the Fleet Overview dashboard. Fully decoupled from Go server completion.

*VERTICAL SLICE SCOPE*

* Domain Types: `web/src/domain/`: strictly typed interfaces matching Go contracts (`AgenticMetricsLedger`, `FleetSummary`, `GroupedMetric`, `RoleMetric`, `ModelMetric`), zero any.
* Boundary Adapter & Mock Harness: `web/src/api/client.ts`: type-safe HTTP client for `/api/*` with error narrowing and mock fixture harness (`web/src/api/mockData.ts`) allowing standalone development and testing without running Go backend.
* Presentation: Top-level navigation bar, theme, and Fleet Overview view:
  - KPI cards: Total Cost, Invocations, Total Tokens, Cache Hit Rate %, Active Duration.
  - Role and Model distribution cards/charts.
  - Ticket directory table with search, sorting, and links to deep dive.
* Quality: TypeScript unit tests for pure formatting functions; Vitest component tests against mock fixtures; zero typecheck errors.

*ACCEPTANCE CRITERIA*

* `npm run build` generates static assets in `web/dist/`.
* `npm run typecheck` passes with zero errors under strict mode.
* App runs stand-alone in Vite dev mode using mock fixtures when backend is absent.
* Fleet Overview displays correct aggregated figures retrieved from `/api/fleet` when live backend is connected.
* Empty dataset handles gracefully without runtime exceptions.
* Zero emojis anywhere in UI text, labels, or tooltips.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`npm run typecheck`, `npm run lint`, `npm test`).
* TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
