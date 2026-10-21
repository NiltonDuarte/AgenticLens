# Implementation Plan: AL-25 Contextual Hover Information Icons for Cards and Tables

## 1. Goal and Iteration Scope

Implement a reusable, zero-emoji, accessible `InfoTooltip` component in `web/src/components/InfoTooltip.tsx`, provide comprehensive styling in `web/src/index.css`, and integrate contextual information icons across all dashboard cards (`KPICard`, `DistributionCard`, `PulseCard`, `FeatureNavCard`) and tables (`TicketTable`, `QuickLaunchTable`, `TicketComparisonsMatrix`). Deliver the full scope of ticket AL-25 in this iteration.

---

## 2. Grounded Reference Architecture and Concrete Paths

The implementation maps to the following existing packages and modules:

### Presentation Layer Components (`web/src/components/`)
- `web/src/components/InfoTooltip.tsx`: New component file rendering the accessible information trigger and positioned tooltip bubble.
- `web/src/components/KPICard.tsx`: Update to embed `InfoTooltip` inside `.kpi-header` when `tooltip` prop is supplied.
- `web/src/components/DistributionCard.tsx`: Update to accept optional `tooltip` prop and render `InfoTooltip` inside card header.
- `web/src/components/TicketTable.tsx`: Update column headers to include `InfoTooltip` for every column with metric calculation explanations.
- `web/src/components/Components.test.tsx`: Add dedicated unit tests for `InfoTooltip` rendering, keyboard focus, hover visibility, placement classes, and zero-emoji assertion. Update existing component tests where header structures evolve.

### Presentation Layer Views (`web/src/views/`)
- `web/src/views/LandingPage.tsx`: Add `InfoTooltip` to pulse cards, feature navigation cards, and Quick-Launch table headers.
- `web/src/views/FleetOverview.tsx`: Supply explanatory tooltip strings to Role Cost Distribution and Model Cost Distribution cards.
- `web/src/views/TicketComparisons.tsx`: Add `InfoTooltip` to all 15 comparison matrix table headers.
- `web/src/views/TicketDeepDive.tsx`: Audit and ensure all existing `KPICard`s pass clear, mathematically accurate tooltip descriptions.
- `web/src/views/MetricProgression.tsx`: Audit and ensure all existing `KPICard`s pass clear, mathematically accurate tooltip descriptions.

### Presentation Styles (`web/src/index.css`)
- `web/src/index.css`: Add styling rules for `.info-tooltip-container`, `.info-tooltip-trigger`, `.info-tooltip-glyph`, `.info-tooltip-bubble`, placement variants, caret arrows, and flex header layouts.

### Verification & Build
- `web/package.json`: Scripts `typecheck`, `lint`, `test`, `build`.
- `backend/internal/adapters/web/embed.go`: Target Go embedding package for `web/dist/`.

---

## 3. Phased Implementation Tasks

### Phase 1: Core Presentation Component (`web/src/components/InfoTooltip.tsx`)
- Task 1.1: Create `web/src/components/InfoTooltip.tsx`.
- Task 1.2: Define `TooltipPlacement` type as union of literals: `'top' | 'bottom' | 'left' | 'right'`.
- Task 1.3: Define `InfoTooltipProps` interface with:
  - `readonly content: string`
  - `readonly placement?: TooltipPlacement`
  - `readonly ariaLabel?: string`
  - `readonly className?: string`
- Task 1.4: Implement component with local state or CSS-driven hover/focus visibility.
- Task 1.5: Render trigger button element:
  - `type="button"`
  - `className="info-tooltip-trigger"`
  - `aria-label` with fallback to "Metric details"
  - `onClick` with `event.stopPropagation` to prevent parent row or sort triggers
- Task 1.6: Render SVG information icon:
  - Clean circular information glyph (`circle cx=12 cy=12 r=10`, `line x1=12 y1=16 x2=12 y2=12`, `line x1=12 y1=8 x2=12 y2=8`)
  - `aria-hidden="true"`
  - Class name `"info-tooltip-glyph"`
- Task 1.7: Render tooltip bubble:
  - `role="tooltip"`
  - `className` combining `"info-tooltip-bubble"` and placement modifier (such as `"placement-top"`)
  - Inner text displaying `content`
- Task 1.8: Export `InfoTooltip` as primary component and `InfoIcon` as alias.

### Phase 2: Card Component Integrations (`web/src/components/`)
- Task 2.1: Refactor `web/src/components/KPICard.tsx`:
  - Import `InfoTooltip` from `./InfoTooltip`.
  - In `.kpi-header`, render `InfoTooltip` beside `.kpi-title` when `tooltip` prop is present, with `ariaLabel` set to `"Information about " + title`.
  - Remove outer container `title` attribute to avoid duplicate browser default tooltip overlay.
- Task 2.2: Refactor `web/src/components/DistributionCard.tsx`:
  - Update `DistributionCardProps` to accept `readonly tooltip?: string`.
  - Wrap card title in a `.distribution-header` container with display flex and space-between alignment.
  - Render `InfoTooltip` beside title when `tooltip` prop is present, with `ariaLabel` set to `"Information about " + title`.

### Phase 3: Table Column Headers Integration (`web/src/components/TicketTable.tsx`)
- Task 3.1: Import `InfoTooltip` in `web/src/components/TicketTable.tsx`.
- Task 3.2: Update `thead th` elements with `InfoTooltip` using `placement="bottom"`:
  - Ticket ID: "Unique identifier for the development ticket and workspace directory"
  - Title: "Summary title of the feature, bugfix, or task tracked by this ticket"
  - Status: "Context window saturation health status (HEALTHY, ELEVATED, BLOATED, or CRITICAL)"
  - Active Duration: "Cumulative active invocation duration spent on this ticket"
  - Cost (USD): "Total model inference spend in US dollars"
  - Cache Hit Rate: "Percentage of prompt tokens read from key-value prompt cache"
  - Code Delivered: "Lines of code added and deleted, and total files changed"
  - SDLC Iterations: "Count of completed development iterations for this ticket"
- Task 3.3: Ensure header text and sort indicators align cleanly with the information icon using inline-flex layout.

### Phase 4: Landing Page Cards & Table Integration (`web/src/views/LandingPage.tsx`)
- Task 4.1: Import `InfoTooltip` in `web/src/views/LandingPage.tsx`.
- Task 4.2: Update Pulse Metrics section cards:
  - Total Ticket Workspaces: "Total number of registered agentic development workspaces with recorded telemetry"
  - Cumulative Fleet Spend: "Total monetary expenditure for LLM model inference across all workspaces"
  - Total Iterations: "Total number of completed SDLC development iterations across all workspaces"
  - Cache Hit Efficiency: "Overall percentage of prompt context tokens served from prompt cache"
- Task 4.3: Update Feature Navigation cards:
  - Fleet Overview: "Macro fleet-level analytics, cumulative spend, token throughput, and role/model distributions"
  - Ticket Deep Dive: "Detailed autopsy of single ticket workspaces, context saturation, subagent turns, and artifacts"
  - Metric Progression: "Longitudinal temporal progression tracking period spend, invocations, tokens, and code churn"
  - Ticket Comparisons: "Cross-workspace benchmarking matrix comparing lead time, compute, test ratios, and health"
- Task 4.4: Update Quick-Launch table headers with `placement="bottom"`:
  - Ticket ID: "Unique identifier for the development ticket and workspace directory"
  - Title: "Summary title of the feature, bugfix, or task tracked by this ticket"
  - Status: "Context window saturation health status (HEALTHY, ELEVATED, BLOATED, or CRITICAL)"
  - Iterations: "Number of SDLC feedback cycles executed for this ticket"
  - Active Duration: "Cumulative active invocation duration spent on this ticket"
  - Action: "Direct navigation action to inspect single ticket autopsy"

### Phase 5: View-Level Tooltip Enrichment (`web/src/views/`)
- Task 5.1: In `web/src/views/FleetOverview.tsx`:
  - Pass `tooltip="Breakdown of cumulative LLM inference spend categorized by agent persona role (orchestrator, coder, tester, reviewer, planner)"` to Role Cost Distribution card.
  - Pass `tooltip="Breakdown of cumulative LLM inference spend categorized by model family and version"` to Model Cost Distribution card.
- Task 5.2: In `web/src/views/TicketComparisons.tsx`:
  - Import `InfoTooltip`.
  - Add `InfoTooltip` with `placement="bottom"` to all 15 matrix table headers:
    - Ticket ID: "Unique identifier for the development ticket workspace"
    - Title: "Descriptive title of the ticket or work item"
    - Status: "Workspace lifecycle status (Completed or In Progress)"
    - Completed: "Timestamp when the ticket workspace execution completed"
    - Started: "Timestamp when the ticket workspace execution began"
    - Lead Time: "Total elapsed clock time from first recorded event to completion"
    - Active Time: "Cumulative active compute duration across all model calls"
    - Iterations: "Total number of SDLC feedback cycles executed"
    - Cost (USD): "Total LLM inference spend in US dollars"
    - Cache Hit: "Percentage of prompt tokens read from KV prompt cache"
    - Test:Prod: "Ratio of test lines added to production lines added"
    - Code Lines: "Total code churn: lines added, lines deleted, and files changed"
    - Tool Success: "Percentage of subagent tool executions that succeeded without error"
    - Context Health: "Context window saturation evaluation status (HEALTHY, ELEVATED, BLOATED, CRITICAL)"
    - Saturation: "Peak context window utilization percentage relative to the 200k token ceiling"

### Phase 6: Dark Theme CSS & Positioning Styles (`web/src/index.css`)
- Task 6.1: Add layout rules for `.info-tooltip-container`:
  - `display: inline-flex`
  - `align-items: center`
  - `position: relative`
  - `vertical-align: middle`
  - `margin-left: 6px`
- Task 6.2: Add styles for `.info-tooltip-trigger`:
  - `background: transparent`
  - `border: none`
  - `padding: 2px`
  - `cursor: pointer`
  - `color: var(--text-muted)`
  - `border-radius: 4px`
  - `line-height: 1`
  - `transition: color 0.15s ease, background-color 0.15s ease`
  - hover state: `color: var(--text-primary)`
  - focus-visible state: `outline: 2px solid var(--accent-blue)`, `outline-offset: 2px`
- Task 6.3: Add styles for `.info-tooltip-glyph`:
  - `width: 14px`
  - `height: 14px`
  - `display: block`
- Task 6.4: Add styles for `.info-tooltip-bubble`:
  - `position: absolute`
  - `z-index: 100`
  - `min-width: 160px`
  - `max-width: 260px`
  - `background-color: var(--bg-surface-elevated)`
  - `color: var(--text-primary)`
  - `border: 1px solid var(--border-color)`
  - `border-radius: 6px`
  - `padding: 6px 10px`
  - `font-size: 0.75rem`
  - `font-weight: 400`
  - `line-height: 1.4`
  - `box-shadow: 0 4px 14px rgba(0, 0, 0, 0.45)`
  - `pointer-events: none`
  - `opacity: 0`
  - `visibility: hidden`
  - `transition: opacity 0.15s ease, visibility 0.15s ease`
- Task 6.5: Add trigger hover and focus rules:
  - `.info-tooltip-trigger:hover + .info-tooltip-bubble, .info-tooltip-trigger:focus + .info-tooltip-bubble, .info-tooltip-container:hover .info-tooltip-bubble: opacity 1, visibility: visible`
- Task 6.6: Add placement positioning rules:
  - `.placement-top: bottom: 100%; left: 50%; transform: translateX(-50%) translateY(-6px);`
  - `.placement-bottom: top: 100%; left: 50%; transform: translateX(-50%) translateY(6px);`
  - `.placement-left: right: 100%; top: 50%; transform: translateY(-50%) translateX(-6px);`
  - `.placement-right: left: 100%; top: 50%; transform: translateY(-50%) translateX(6px);`
- Task 6.7: Add subtle arrow carets for each placement using CSS pseudo-elements (`::after`).
- Task 6.8: Add header layout rules:
  - `.distribution-header: display flex, justify-content space-between, align-items center, margin-bottom 16px, border-bottom 1px solid var(--border-subtle), padding-bottom 8px.`
  - Update table `th` elements to maintain flex or inline-flex alignment between label, sort indicator, and info icon.

### Phase 7: Automated Testing and Quality Verification (`web/src/`)
- Task 7.1: Add unit tests in `web/src/components/Components.test.tsx`:
  - Verify `InfoTooltip` renders button with `aria-label`.
  - Verify `InfoTooltip` renders tooltip bubble with `role="tooltip"` and matching content.
  - Verify mouseenter and mouseleave / focus and blur state transitions.
  - Verify placement classes are applied correctly.
  - Verify zero emojis in rendered content or attributes.
- Task 7.2: Verify `KPICard` tests in `Components.test.tsx` assert presence of `InfoTooltip` when `tooltip` prop is passed.
- Task 7.3: Verify `DistributionCard` tests assert rendering of `InfoTooltip` when `tooltip` prop is passed.
- Task 7.4: Run full test suite: `npm test` in `web/`. Ensure 100% test pass rate.
- Task 7.5: Run typecheck: `npm run typecheck` in `web/`. Ensure zero TypeScript compilation errors.
- Task 7.6: Run lint: `npm run lint` in `web/`. Ensure zero ESLint warnings.

### Phase 8: Repository Build Verification
- Task 8.1: Run `make check` from root.
- Task 8.2: Run `make test` from root.
- Task 8.3: Run `make build` from root to verify single binary compilation with embedded frontend assets.
- Task 8.4: Verify zero emojis across all modified code, CSS, tests, and documentation.

---

## 4. Execution Dependencies and Delivery Sequence

1. Phase 1 (Core InfoTooltip Component) -> unblocks Phase 2, Phase 3, Phase 4, and Phase 5 integrations.
2. Phase 6 (CSS Theme Styling) -> can proceed concurrently with or immediately following Phase 1.
3. Phase 2 (KPICard & DistributionCard) -> enables card tooltips across all views.
4. Phase 3 & Phase 4 & Phase 5 (Table Headers and View Integrations) -> completes all surface area coverage.
5. Phase 7 (Testing Suite) -> verifies accessibility attributes, behavioral contracts, and regressions.
6. Phase 8 (Quality Gates & Build) -> confirms production readiness and single-binary packaging.
