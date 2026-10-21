# PR: PO-7: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization

## Summary
Defines the product specification, architectural contracts, and vertical tracer-bullet Jira ticket (AL-34) addressing usability and domain issues in the Metric Progression view: removing the "Cumulative" option from time granularity, maintaining composite multi-series spend and token charts, and upgrading code churn visualization into a bi-directional diverging chart centered on a zero baseline with cumulative net code growth tracking.

## Initiatives Covered

1. **Purge "Cumulative" from Time Granularity**:
   - Eliminates the category error conflating bucket granularity (`Hourly`, `Daily`, `Weekly`) with running accumulation.
   - Cleans up `TimeGranularity` across the Go backend (`enums.go`, `ParseTimeGranularity`, `aggregations.go`), HTTP API (`/api/progression`), and TypeScript domain (`types.ts`, `progression.ts`, `formatters.ts`).
   - Removes the "Cumulative" toggle button from `MetricProgression.tsx` controls bar.
   - Prevents dead interactive states where 4 out of 5 charts on the page ignored the toggle.

2. **Keep Spend and Token Charts Composite**:
   - Chart 1 ("Financial Spend & Cumulative Trajectory"): renders both Period Spend (bar) and Cumulative Spend (line) across all active granularities without requiring toggles; title remains static.
   - Chart 2 ("Token Volume Distribution"): preserves composite breakdown of Cache Read, Input, and Output tokens.

3. **Bi-Directional Code Churn, Cumulative Net Growth, and Compact Y-Axis Scaling**:
   - Upgrades `TrendChart.tsx` to handle bi-directional diverging bar charts centered on a zero baseline (`yZero`) when values span negative territory.
   - Eliminates fractional discrete counts (e.g. `33,306.56 lines`) via integer rounding on tick intervals.
   - Introduces compact scaling (`k`, `M`) on Chart 4's Y-axis to prevent label overflow outside the chart card boundaries.
   - Chart 4 ("Delivered Code Line Churn & Net Growth"):
     - Green bars grow upward from zero for code additions (`lines_added`).
     - Red bars grow downward from zero for code deletions (`-lines_deleted`).
     - Line series tracks cumulative net code growth across the selected time window (`cumulative_net_lines_added`).
     - Tooltip clearly indicates exact integer additions (+), deletions (-), and cumulative net growth (+/-).

## Ticket Breakdown
- Ready: `AL-34: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization`

## Pointers
- Specification: `workspaces/PO-7/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Development Ticket: `workspaces/plan/to-do/AL-34 eliminate cumulative granularity and implement bidirectional code churn visualization.md`
- Session Summary: `workspaces/PO-7/SUMMARY.md`
