# feat(presentation): contextual hover information icons and responsive table containment (AL-25)

## Overview
Adds a reusable, accessible contextual information hover component (`InfoTooltip` / `InfoIcon`) and integrates it across all dashboard KPI cards, distribution cards, longitudinal trend charts, and tabular headers. Provides instant clarification on telemetry metrics, mathematical rollups, and column definitions across AgenticLens, with full horizontal overflow containment and responsive scrolling.

Detailed design constraints and architectural boundaries are documented in `workspaces/AL-25/CONTRACT.md`. The unified change log and execution telemetry ledger are recorded in `workspaces/AL-25/SUMMARY.md`.

## Key Changes

### Presentation Components (`web/src/components/`)
- Created `InfoTooltip.tsx` exporting `InfoTooltip` and `InfoIcon` alias with strictly typed placement (`'top' | 'bottom' | 'left' | 'right'`), inline SVG glyph (`aria-hidden="true"`), full keyboard support (focus/blur/Escape), semantic ARIA hooks (`role="tooltip"`, `aria-describedby`), and `event.stopPropagation()` on click.
- `KPICard.tsx`: Removed native HTML container `title` attribute; embedded `InfoTooltip` inside `.kpi-header` beside `.kpi-title`.
- `DistributionCard.tsx`: Added optional `tooltip` prop and `.distribution-header` flex container. Defensively guarded against zero cost datasets.
- `TrendChart.tsx`: Added optional `tooltip` prop; renders `InfoTooltip` in `.trend-title-group` across populated and empty states.
- `SaturationBar.tsx`: Added optional `tooltip` prop explaining context saturation against the 200,000 token ceiling.
- `ThreadViewer.tsx`: Added `InfoTooltip` to section header and all 10 column headers in `thread-models-table`.

### Views & Table Integrations
- `TicketTable.tsx`: Integrated `InfoTooltip` across all 8 column headers with `placement="bottom"`.
- `TicketDeepDive.tsx`: Added tooltips to Subagent Turn Velocity, Role Resource Consumption title (`placement="left"`) + 5 column headers (`placement="bottom"`), Model Resource Consumption title (`placement="left"`) + 5 column headers (`placement="bottom"`), and refined KPI tooltips.
- `MetricProgression.tsx`: Attached concrete tooltips across all 5 longitudinal trend charts.
- `LandingPage.tsx`: Integrated `InfoTooltip` across 4 pulse cards, 4 feature navigation cards, and all 6 Quick-Launch table headers.
- `FleetOverview.tsx`: Added tooltips to Role and Model Cost Distribution cards and KPI cards.
- `TicketComparisons.tsx`: Integrated `InfoTooltip` across all 15 comparison matrix column headers.

### Layout, Responsive Table Scrolling & Tooltip Containment (`web/src/index.css`)
- Defined global `.table-responsive` utility class (`width: 100%; overflow-x: auto; -webkit-overflow-scrolling: touch;`).
- Set `.data-table` to `min-width: 480px;` and `.thread-models-table` to `min-width: 850px;` with `white-space: nowrap;` on cells, ensuring tables establish smooth horizontal scrollbars rather than spilling out of bounds on constrained viewports.
- Changed `.table-card` and `.thread-accordion-item` to `overflow: visible;` so tooltip bubbles float freely without container clipping.
- Separated coordinate spaces: card headers remain outside `.table-responsive` with unclipped `placement="left"`, while table column headers remain inside with `placement="bottom"` scrolling synchronously with table contents.

### Telemetry Semantic Clarifications
- Clarified distinction between Active Compute Duration (cumulative agent compute time across all subagent invocations, analogous to man-hours, which can exceed calendar elapsed lead time due to concurrent execution) and Turnaround Lead Time (wall-clock elapsed time from inception to resolution).

## Architecture & Compliance
- **FCIS Separation**: Go domain core remains untouched. Presentation shell purely renders visual cues.
- **Offline Independence**: Zero external floating-UI runtime libraries or CDN dependencies. Pure React and CSS.
- **Typing & Zero Any**: Strict TypeScript mode enforced with zero `any` or `unknown`.
- **Zero Emojis**: Scanned and verified zero emojis in code, markup, CSS, and documentation.
