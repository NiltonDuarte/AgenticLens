# AL-48 Technical Contract

## Problem

The Ticket Comparisons view presents cross-ticket metrics in a table but does not provide a visual way to compare relationships between two metrics or isolate outliers. This ticket adds an interactive scatter plot to the existing comparison view so operators can select metrics, exclude tickets, change scale mode, inspect values, and navigate to ticket details.

## Scope

- Add a native SVG scatter plot in `web/src/components/TicketScatterPlot.tsx`, integrated into `web/src/views/TicketComparisons.tsx`.
- Use the comparison rows already derived from fetched workspace ledgers. No backend, API, schema, or npm dependency changes are in scope.
- Provide selectable X and Y metrics, linear and logarithmic scales, manual ticket exclusion, hover details, and ticket selection through the existing `onSelectTicket` callback.
- Default the X metric to Total Cost (`cost_usd`) and the Y metric to Net Lines of Code (computed as `lines_added - lines_deleted`).
- Support the requested metrics: cost, net lines, active duration, lead duration, SDLC iterations, cache hit rate, added lines, deleted lines, tool success rate, and orchestrator saturation.

## Data and Metric Contract

- `TicketComparisonRow` in `web/src/domain/types.ts` is the chart's typed data boundary. `web/src/domain/formatters.ts` derives rows from existing ledgers.
- Net lines are a pure derived metric: added lines minus deleted lines. Do not add redundant persisted or API fields.
- The existing row field `root_saturation_pct` supplies the requested orchestrator saturation metric; present it with the chart label “Orchestrator Saturation”.
- `lead_duration_seconds` is nullable. A missing lead duration is not a numeric zero and must not be plotted as a valid coordinate.
- Ticket completion state is represented by `completed_at`: non-null means Completed; null means In Progress.
- Metric extraction and scale calculations must be deterministic and side-effect free. Excluded tickets and rows lacking either selected metric must not influence visible points, auto-recommendation, or axis bounds.

## Functional Requirements

- Show the chart with the exact title “Cross-Ticket Metric Correlation”.
- Let each axis independently select any supported metric. Defaults are `cost_usd` on X and net lines on Y.
- Allow operators to exclude and re-include any displayed ticket using an accessible multi-ticket checklist.
- Support linear and logarithmic scaling using `log10(value + 1)` for non-negative metrics. Avoid invalid logarithms and non-finite coordinates; negative net-line values remain valid in linear mode and must be handled safely if log mode is selected.
- Recommend logarithmic scaling when the active values span more than 50x, without overriding the operator's explicit selection. Recompute recommendation and padded bounds from currently included, plottable rows.
- Apply 10% outer padding to non-degenerate axis ranges and define a stable useful position for single-valued or empty ranges.
- Render completion state with distinct colors, render one SVG circle per visible plottable ticket, and expose ticket ID, title, and exact formatted X/Y values on hover or keyboard focus.
- Clicking a point calls `onSelectTicket(ticketId)`.

## Architecture and Boundaries

- UI and SVG rendering belong in the presentation layer under `web/src/components/` and `web/src/views/`.
- Metric extraction, transforms, recommendations, and bounds belong in pure TypeScript helpers in `web/src/domain/` (or a narrowly scoped pure module colocated with the component if that better matches existing conventions).
- Reuse available formatters in `web/src/domain/formatters.ts` where appropriate. Do not add I/O, network access, or clock reads to metric helpers.
- Keep TypeScript strict, use explicit metric-key/value unions, avoid `any`, and add no external charting packages.

## Acceptance Boundaries

- Component and view tests cover defaults, each supported metric mapping, exclusion effects on points and scales, linear/log transforms, 50x recommendation, padding and degenerate ranges, unavailable nullable metrics, SVG rendering, tooltip details, completion colors, and click routing.
- `TicketComparisons.test.tsx` verifies chart integration using its existing mocked ledger fetch setup.
- `npm run typecheck`, `npm run lint`, and `npm test` pass; the repository `make` quality gate is also required at delivery.
- No Go/backend changes, API shape changes, or package dependency changes.
