h3. Title: AL-48: Interactive Multi-Metric Ticket Scatter Plot with Outlier Exclusion and Dynamic Scaling in Comparison View

*TARGET*: AgenticLens (presentation / web / components / views / tests)

*CONTEXT*
 
The Cross-Ticket Comparison view (`web/src/views/TicketComparisons.tsx`) provides KPI summary cards and a tabular matrix comparing workspaces across 15+ dimensions.
However, it lacks an interactive visual correlation tool. Operators and engineering leaders cannot visually inspect cross-ticket trade-offs, such as plotting Spend against Net Lines of Code, Iterations against Duration, or Cache Hit Rate against Cost. Furthermore, software metrics suffer from power-law distributions where a single runaway ticket compresses 95% of tickets into an indistinguishable clump at the origin unless outlier exclusion and logarithmic scaling are supported.

*GOAL*
 
Add an interactive, pure SVG multi-metric scatter plot (`TicketScatterPlot.tsx`) to the Comparison view, defaulting to X = Total Cost and Y = Net Lines of Code, with a ticket exclusion filter to manually remove outliers, dynamic scale recalculation, selectable linear or logarithmic axes, and click navigation to Ticket Deep Dive autopsies.

*VERTICAL SLICE SCOPE*
 
* Domain Core: Pure scaling and metric extraction utilities in `web/src/domain/` or component helpers mapping `TicketComparisonRow` to numeric coordinate tuples `(x, y)` supporting linear and log scales, dynamic min/max recalculation, zero-safe log transforms (`log10(val + 1)`), and 10% outer padding.
* Boundary Adapter: Zero backend changes; operates directly on fetched `TicketComparisonRow` dataset.
* Presentation:
  - Create `web/src/components/TicketScatterPlot.tsx` using native SVG (zero external charting dependencies).
  - Add dropdown selectors for X-Axis and Y-Axis:
    - Default X-Axis: `cost_usd` (Total Cost USD)
    - Default Y-Axis: `net_lines_added` (Net Lines of Code)
    - Additional metrics: `active_duration_seconds`, `lead_duration_seconds`, `sdlc_iterations`, `cache_hit_rate_pct`, `lines_added`, `lines_deleted`, `tool_success_rate_pct`, `orchestrator_saturation_pct`.
  - Add ticket exclusion filter (multi-select / checklist of all tickets) allowing operators to manually exclude outlier tickets from the scatter plot.
  - Dynamic scale selector: allow switching between linear scale and logarithmic scale for wide-span metrics, with automatic recommendation when max / min ratio exceeds 50x.
  - Graph scale dynamically recalculates strictly from non-excluded visible tickets.
  - Render points with status color coding (Completed = green/blue, In Progress = amber/accent).
  - Hover tooltip showing Ticket ID, Title, X-metric value, and Y-metric value.
  - Point click handler invoking `onSelectTicket(ticketId)`.
* Quality: Comprehensive unit tests in `TicketScatterPlot.test.tsx` and integration tests in `TicketComparisons.test.tsx` verifying:
  - Default metrics X=cost, Y=net_lines_added.
  - Outlier exclusion filtering and dynamic scale recalculation.
  - Linear and logarithmic scaling transforms.
  - SVG circle element rendering, tooltip appearance on hover, and click routing.

*ACCEPTANCE CRITERIA*
 
* Comparison page features an interactive scatter plot titled "Cross-Ticket Metric Correlation".
* X-axis defaults to Total Cost (`cost_usd`) and Y-axis defaults to Net Lines of Code (`net_lines_added`).
* Operators can select any valid numerical metric for X and Y axes independently.
* Includes a ticket exclusion filter allowing operators to check/uncheck tickets to manually remove outliers.
* Graph scale automatically updates to fit only the active, non-excluded tickets.
* Supports linear and logarithmic scale modes depending on metric span.
* Hovering over a ticket point displays a tooltip with ticket identifier, title, and exact formatted values.
* Clicking a ticket point navigates to the Ticket Deep Dive page for that ticket.
* Pure SVG implementation with zero new npm package dependencies.
* `npm run typecheck`, `npm run lint`, and `npm test` pass with zero errors.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
