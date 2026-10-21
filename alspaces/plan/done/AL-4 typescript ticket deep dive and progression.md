h3. Title: AL-4: Implement Ticket Deep Dive, Progression Charts, and Comparisons

*TARGET*: AgenticLens (presentation | web | tests)

*CONTEXT*

Streamlit views include Ticket Deep Dive (context health, subagent turns, tool executions), Metric Progression (daily/weekly/cumulative trends), and Ticket Comparisons. These analytical views need to be implemented in the TypeScript frontend.

*GOAL*

Deliver Ticket Deep Dive, Metric Progression, and Ticket Comparison views in the TypeScript dashboard connected to backend API endpoints.

*VERTICAL SLICE SCOPE*

* Domain Logic: `web/src/domain/formatters.ts` and progression rollup utilities matching Go calculations.
* Presentation:
  - Ticket Deep Dive view: Context window health gauges, turn velocity by role, tool execution success/error ratios, and thread breakdown.
  - Metric Progression view: Trend charts for cost, tokens, and code line churn grouped by daily, weekly, or cumulative buckets.
  - Comparisons view: Multi-ticket tabular comparison matrix with sorting.
* Quality: Unit tests for progression bucket calculations and table sorting; responsive render tests.

*ACCEPTANCE CRITERIA*

* Deep dive renders all workflow efficiency metrics accurately for any selected ticket.
* Context health status correctly indicates `HEALTHY`, `ELEVATED`, `BLOATED`, or `CRITICAL` according to token ceiling rules.
* Metric progression allows toggling between daily, weekly, and cumulative views.
* Zero emojis in labels, axes, legends, or values.
* `npm run typecheck` and `npm test` pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`npm run typecheck`, `npm run lint`, `npm test`).
* TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
