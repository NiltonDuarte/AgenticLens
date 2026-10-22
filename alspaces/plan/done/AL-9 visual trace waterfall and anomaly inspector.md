h3. Title: AL-9: Visual Trace Waterfall and Anomaly Inspector

*TARGET*: AgenticLens (presentation | web | tests)

*CONTEXT*

Debugging agent stalls and context bloat currently requires reading terminal logs or markdown tables. Developers and leads need an intuitive visual timeline showing the chronological trajectory of subagent turns, tool executions, and warning flags in the web dashboard.

*GOAL*

Build an interactive, chronological trace waterfall and anomaly inspector drawer in the TypeScript SPA Ticket Deep Dive view.

*VERTICAL SLICE SCOPE*

* Domain Core & API Client: Strictly typed TypeScript models in `web/src/types/telemetry.ts` for `Anomaly`, `EventFact`, `Severity`, and `FleetAnomalySummary`. API client functions for fetching events and anomalies.
* UI Components (`web/src/components/`):
  - `TraceWaterfall`: Horizontal/vertical timeline mapping turns and tool calls chronologically.
  - `AnomalyBadge`: Color-coded indicator (amber for warning, red for critical) positioned on anomalous turns/events.
  - `AnomalyDrawer`: Expandable slide-out or modal inspecting tool error tracebacks, argument payloads, and token deltas for selected turns.
  - `FleetAnomalyCard`: Summary card on Fleet Overview showing anomaly frequencies across workspaces.
* Quality: Vitest unit tests verifying timeline rendering, badge color coding, drawer expansion, and graceful fallback when viewing legacy v1 workspaces with no event stream.

*ACCEPTANCE CRITERIA*

* Opening a v2 ticket in the web dashboard displays the Trace Waterfall component with sequential turns and tool executions.
* Turns with detected anomalies display visual anomaly badges (amber warning, red critical).
* Clicking a flagged turn opens the Anomaly Drawer displaying failure details and error logs.
* Legacy v1 tickets display an informative banner stating event stream is unavailable while showing v1 aggregate anomalies.
* TypeScript strict compilation passes with zero type errors and zero `any`.
* `npm run test` or `npm run build` succeeds cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`npm run build`, `npm run test`).
* TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
