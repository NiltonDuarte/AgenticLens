# Iteration 02 Implementation Plan: AL-25 Graph and Table Contextual Tooltip Enrichment

## 1. Goal and Iteration Scope

Iteration 02 addresses user feedback following the Iteration 01 delivery of the core `InfoTooltip` component. While primary summary cards and top-level directory tables received contextual tooltips in Iteration 01, detailed analytical tables and graphical visualizations throughout the dashboard remain unannotated. Furthermore, operators require conceptual clarification regarding why Active Compute Duration often exceeds Turnaround Lead Time in multi-agent telemetry.

This iteration expands contextual hover information icons to:
1. All graphical telemetry visualizations: `TrendChart` component, all five longitudinal charts in `MetricProgression`, the Subagent Turn Velocity card, and the Context Window Saturation bar in `TicketDeepDive`.
2. All analytical drill-down tables: Role Resource Consumption and Model Resource Consumption tables in `TicketDeepDive` (card titles and column headers), and the session thread breakdown table in `ThreadViewer`.
3. Tooltip copy refinement across all views, explicitly clarifying the distinction between cumulative agent compute time (man-hours) and wall-clock turnaround lead time.

---

## 2. Telemetry Architecture Clarification: Active Compute Duration vs Turnaround Lead Time

### Root Cause of the Metric Disparity
In single-threaded, sequential human workflows, total active working duration cannot exceed total calendar elapsed time. However, in autonomous multi-agent software engineering systems:

1. Concurrent Subagent Invocations:
   When an orchestrator dispatches tasks to subagents, multiple model inference calls or tool executions may run concurrently or in rapid overlapping sessions. The durations of these individual invocations are summed together to produce Active Compute Duration.

2. Cumulative Worker Effort (Man-Hours Analogy):
   Active Compute Duration represents cumulative agent compute time across all invocations, subagents, and feedback loops (equivalent to contractor man-hours). For example, if three subagents each perform 2 minutes of compute within a 3-minute wall-clock span, the Active Compute Duration is 6 minutes, whereas the Turnaround Lead Time is only 3 minutes.

3. Elapsed Wall-Clock Clock Time:
   Turnaround Lead Time measures real-world wall-clock elapsed time strictly from ticket inception (the timestamp of the initial user prompt or first recorded event) to resolution (the timestamp of the final commit or iteration completion).

### Operational Definition Invariant
All tooltip descriptions and interface labels must reinforce this distinction:
- Active Compute Duration / Active Time: Cumulative compute execution time across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time when subagents execute concurrently.
- Turnaround Lead Time / Lead Time: Total wall-clock elapsed time from workspace inception to completion.

---

## 3. Grounded Reference Architecture and Concrete Paths

The changes strictly touch the presentation shell without mutating the pure Go domain core or TypeScript domain formatters:

### Presentation Layer Components (`web/src/components/`)
- `web/src/components/TrendChart.tsx`: Update `TrendChartProps` to accept an optional `tooltip` property. Import `InfoTooltip` and render it beside the chart title in `.trend-header` and `.trend-card.empty`.
- `web/src/components/SaturationBar.tsx`: Update `SaturationBarProps` to accept an optional `tooltip` property. Import `InfoTooltip` and render it inside `.saturation-title-group` with default explanatory copy describing context window thresholds against the 200k token ceiling.
- `web/src/components/ThreadViewer.tsx`: Import `InfoTooltip`. Add `InfoTooltip` to the card title in `.thread-viewer-header`. Add `InfoTooltip` with `placement="bottom"` to all ten column headers of `.thread-models-table`.
- `web/src/components/TicketTable.tsx`: Update the Active Duration column header tooltip copy to explicitly mention cumulative compute (man-hours) and concurrency.
- `web/src/components/Components.test.tsx`: Add unit tests asserting tooltip rendering on `TrendChart`, `SaturationBar`, and `ThreadViewer`.

### Presentation Layer Views (`web/src/views/`)
- `web/src/views/MetricProgression.tsx`: Supply distinct, domain-accurate tooltip strings to all five `TrendChart` instances (Spend Trajectory, Token Volume Distribution, Cache Hit Rate Trajectory, Delivered Code Line Churn, and Peak Context Saturation).
- `web/src/views/TicketDeepDive.tsx`:
  - Supply `InfoTooltip` to the Subagent Turn Velocity card header (`.turn-velocity-header`).
  - Wrap Role Resource Consumption and Model Resource Consumption card titles with header containers and attach `InfoTooltip`.
  - Add `InfoTooltip` with `placement="bottom"` to all column headers in both Role and Model resource consumption tables (`Role`, `Model`, `Invocations`, `Tokens`, `Cache Hit`, and `Cost`).
  - Refine tooltip copy for the Turnaround Lead Time and Active Compute Duration `KPICard` elements.
- `web/src/views/TicketComparisons.tsx`: Refine tooltip copy for the Average Active Duration `KPICard`, the Lead Time table header, and the Active Time table header.
- `web/src/views/LandingPage.tsx`: Refine tooltip copy for the Quick-Launch table Active Duration header.
- `web/src/views/FleetOverview.tsx`: Refine tooltip copy for the Active Duration `KPICard`.
- `web/src/views/TicketDeepDive.test.tsx`: Add assertions verifying the presence of information tooltips on tables, turn velocity, and refined KPI tooltips.
- `web/src/views/MetricProgression.test.tsx`: Add assertions verifying the presence of information tooltips across all five trend charts.

### Presentation Styles (`web/src/index.css`)
- `web/src/index.css`: Add styles for `.trend-title-group`, `.table-card-header`, and adjust `.data-table th` and `.thread-models-table th` flex alignments to support tooltips on right-aligned numerical headers without layout shift.

---

## 4. Phased Implementation Tasks

### Phase 1: Trend Chart Component Tooltip Integration (`web/src/components/TrendChart.tsx`)
- Task 1.1: Import `InfoTooltip` from `./InfoTooltip` in `web/src/components/TrendChart.tsx`.
- Task 1.2: Update `TrendChartProps` interface to include `readonly tooltip?: string | undefined;`.
- Task 1.3: In `.trend-header`, wrap `.trend-title` inside a container with class `.trend-title-group`.
- Task 1.4: Conditionally render `InfoTooltip` beside the title when `tooltip` is defined, with `ariaLabel` set to `Information about ${title}`.
- Task 1.5: In the empty state branch (`.trend-card.empty`), wrap `.trend-title` similarly and display `InfoTooltip` if `tooltip` is supplied.

### Phase 2: Metric Progression Longitudinal Charts (`web/src/views/MetricProgression.tsx`)
- Task 2.1: Supply explanatory tooltip to Chart 1 (Financial Trajectory):
  "Historical model inference spend per time period alongside cumulative trajectory in US dollars."
- Task 2.2: Supply explanatory tooltip to Chart 2 (Token Volume Distribution):
  "Distribution of token throughput across input prompts, output completions, and key-value prompt cache reads."
- Task 2.3: Supply explanatory tooltip to Chart 3 (Prompt Cache Hit Rate Trajectory):
  "Percentage of prompt tokens served from prompt cache per bucket against baseline (50%) and target (80%) thresholds."
- Task 2.4: Supply explanatory tooltip to Chart 4 (Delivered Code Line Churn & Net Growth):
  "Volume of production and test code lines added, deleted, and net code growth delivered per time bucket."
- Task 2.5: Supply explanatory tooltip to Chart 5 (Peak Context Window Saturation vs Ceiling):
  "Peak active context tokens per time bucket measured against the standard 200,000 token healthy context ceiling."

### Phase 3: Saturation Bar & Turn Velocity Visualizations (`web/src/components/SaturationBar.tsx` and `web/src/views/TicketDeepDive.tsx`)
- Task 3.1: In `web/src/components/SaturationBar.tsx`:
  - Import `InfoTooltip` from `./InfoTooltip`.
  - Update `SaturationBarProps` to accept `readonly tooltip?: string | undefined;`.
  - In `.saturation-title-group`, render `InfoTooltip` beside `.saturation-title`. Use `tooltip` prop if provided, or default to: "Context window utilization of the orchestrator and subagents evaluated against the 200,000 token healthy operating ceiling."
- Task 3.2: In `web/src/views/TicketDeepDive.tsx`:
  - Import `InfoTooltip` from `../components/InfoTooltip`.
  - In `.turn-velocity-header`, wrap `.section-title` in a title group and attach `InfoTooltip` with content: "Average and peak conversational turns executed per agent invocation, categorized by subagent persona role."

### Phase 4: Ticket Deep Dive Detailed Resource Tables (`web/src/views/TicketDeepDive.tsx`)
- Task 4.1: Refactor Role Resource Consumption card title:
  - Wrap title in a `.table-card-header` container.
  - Render `InfoTooltip` with content: "Telemetry breakdown of model invocations, token volume, cache hit rate, and spend by agent persona role."
- Task 4.2: Update Role Resource Consumption table headers with `InfoTooltip` using `placement="bottom"`:
  - Role: "Agent persona role (orchestrator, coder, tester, reviewer, planner)"
  - Invocations: "Total count of model invocations initiated by this role"
  - Tokens: "Total input and output tokens consumed by this role"
  - Cache Hit: "Percentage of prompt tokens served from prompt cache for this role"
  - Cost: "Total model inference spend in US dollars for this role"
- Task 4.3: Refactor Model Resource Consumption card title:
  - Wrap title in a `.table-card-header` container.
  - Render `InfoTooltip` with content: "Telemetry breakdown of invocations, token volume, cache hit rate, and spend by model family and version."
- Task 4.4: Update Model Resource Consumption table headers with `InfoTooltip` using `placement="bottom"`:
  - Model: "Model identifier and version invoked"
  - Invocations: "Total count of invocations executed on this model"
  - Tokens: "Total input and output tokens consumed by this model"
  - Cache Hit: "Percentage of prompt tokens served from prompt cache for this model"
  - Cost: "Total model inference spend in US dollars for this model"

### Phase 5: Thread Viewer Table and Header Integration (`web/src/components/ThreadViewer.tsx`)
- Task 5.1: Import `InfoTooltip` from `./InfoTooltip` in `web/src/components/ThreadViewer.tsx`.
- Task 5.2: In `.thread-viewer-header`, render `InfoTooltip` beside `.section-title` with content: "Hierarchical session execution threads tracking multi-agent turns, context compactions, and model usage per role."
- Task 5.3: Update all ten column headers of `.thread-models-table` with `InfoTooltip` using `placement="bottom"`:
  - Role: "Agent persona role executing in this thread"
  - Model: "Model family and version used in this session"
  - Invocations: "Total count of model invocations in this thread"
  - Input: "Input prompt tokens sent to the model"
  - Output: "Completion output tokens generated by the model"
  - Cache Read: "Prompt tokens read directly from prompt cache"
  - Cache Hit: "Percentage of prompt context served from cache"
  - Cost: "Total inference cost in US dollars for this model"
  - Peak Context: "Maximum single-turn context window size encountered in this thread"
  - Context Health: "Context window saturation status against healthy ceiling (HEALTHY, ELEVATED, BLOATED, CRITICAL)"

### Phase 6: Tooltip Copy Refinements (Active Compute Duration vs Lead Time)
- Task 6.1: In `web/src/views/TicketDeepDive.tsx`:
  - Turnaround Lead Time KPICard: "Total wall-clock elapsed time from ticket inception to completion. Can be less than Active Compute Duration when subagents execute in parallel."
  - Active Compute Duration KPICard: "Cumulative active compute duration across all agent invocations (analogous to cumulative man-hours). Can exceed wall-clock lead time due to concurrent subagent execution."
- Task 6.2: In `web/src/components/TicketTable.tsx`:
  - Active Duration header: "Cumulative active compute duration across all agent invocations (cumulative man-hours), which can exceed elapsed wall-clock time if agents run concurrently."
- Task 6.3: In `web/src/views/TicketComparisons.tsx`:
  - Average Active Duration KPICard: "Average cumulative active compute time across all agent invocations per ticket (cumulative man-hours)."
  - Lead Time header: "Total wall-clock elapsed time from first event to completion."
  - Active Time header: "Cumulative active compute duration across all agent invocations (cumulative man-hours); can exceed elapsed lead time when subagents run in parallel."
- Task 6.4: In `web/src/views/LandingPage.tsx`:
  - Quick-Launch table Active Duration header: "Cumulative active invocation duration spent on this ticket (cumulative man-hours)."
- Task 6.5: In `web/src/views/FleetOverview.tsx`:
  - Active Duration KPICard: "Cumulative active agent execution time across all tickets (cumulative man-hours)."

### Phase 7: Styling and Layout Adaptations (`web/src/index.css`)
- Task 7.1: Add `.trend-title-group`:
  - Display: inline-flex
  - Align-items: center
  - Gap: 8px
- Task 7.2: Add `.table-card-header`:
  - Display: flex
  - Justify-content: space-between
  - Align-items: center
  - Margin-bottom: 14px
  - Border-bottom: 1px solid var(--border-subtle)
  - Padding-bottom: 8px
- Task 7.3: Adjust `.table-card-title`:
  - Reset border-bottom and padding-bottom when child of `.table-card-header` to prevent double borders.
- Task 7.4: Support `.th-content` in `.data-table th` and `.thread-models-table th`:
  - Display: inline-flex
  - Align-items: center
  - Gap: 6px
  - For `.num-col .th-content`: justify-content flex-end, ensuring numerical column alignment is preserved.

### Phase 8: Testing, Verification, and Build
- Task 8.1: Update `web/src/components/Components.test.tsx`:
  - Add test verifying `TrendChart` displays `InfoTooltip` when `tooltip` prop is provided.
  - Add test verifying `SaturationBar` renders `InfoTooltip` with ceiling explanation.
  - Add test verifying `ThreadViewer` renders `InfoTooltip` on session header and thread model column headers.
- Task 8.2: Update `web/src/views/TicketDeepDive.test.tsx`:
  - Verify that Role and Model tables render information icons on titles and column headers.
  - Verify Subagent Turn Velocity card displays information icon.
  - Verify the clarified tooltip texts for Turnaround Lead Time and Active Compute Duration.
- Task 8.3: Update `web/src/views/MetricProgression.test.tsx`:
  - Verify all five trend charts render `InfoTooltip` triggers.
- Task 8.4: Execute `npm run typecheck` in `web/` to confirm strict typing and zero `any`.
- Task 8.5: Execute `npm run lint` in `web/` to verify ESLint compliance.
- Task 8.6: Execute `npm test` in `web/` to ensure all tests pass with zero regressions.
- Task 8.7: Verify complete absence of emojis across all code, tests, and documentation.

---

## 5. Execution Dependencies and Delivery Sequence

1. Phase 1 (`TrendChart.tsx`) and Phase 3 (`SaturationBar.tsx`) unblock Phase 2 (`MetricProgression.tsx`) and part of Phase 4 (`TicketDeepDive.tsx`).
2. Phase 7 (`index.css`) should accompany Phase 1, Phase 4, and Phase 5 to ensure header and table alignment remains pixel-perfect.
3. Phase 4 (`TicketDeepDive.tsx`) and Phase 5 (`ThreadViewer.tsx`) establish comprehensive table coverage.
4. Phase 6 (copy refinement) applies across all relevant views and components.
5. Phase 8 (testing) validates end-to-end functionality, accessibility attributes, and regression resilience.

---

## 6. Acceptance Criteria

1. Every `TrendChart` instance (including all five charts in `MetricProgression`) displays an accessible `InfoTooltip` beside its title.
2. The `SaturationBar` component displays an `InfoTooltip` explaining context saturation relative to the 200k ceiling.
3. The Subagent Turn Velocity card in `TicketDeepDive` displays an `InfoTooltip` explaining turn velocity metrics.
4. Both Role Resource Consumption and Model Resource Consumption tables in `TicketDeepDive` display an `InfoTooltip` on the card header and on every column header.
5. The `ThreadViewer` component displays an `InfoTooltip` on the section header and on every column header of the thread models table.
6. All tooltips for Active Duration, Active Compute, and Turnaround Lead Time explicitly explain that Active Compute Duration represents cumulative worker effort (man-hours) that can exceed wall-clock Turnaround Lead Time due to concurrent subagent execution.
7. Information icon clicks on table headers stop click propagation and do not trigger row selection or column sort toggling.
8. Zero emojis appear in any markup, icon, tooltip copy, CSS, or test file.
9. Full test suite (`npm test`), type check (`npm run typecheck`), and linter (`npm run lint`) pass with zero errors.
