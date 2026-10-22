h3. Title: AL-25: Contextual Hover Information Icons for Cards and Tables

*TARGET*: AgenticLens (presentation | components | tests)

*CONTEXT*

The AgenticLens dashboard presents dense operational metrics, such as context ceiling saturation, prompt cache hit ratios, test-to-production code ratios, and subagent turn velocities. Users unfamiliar with multi-agent telemetry or specific mathematical rollups have no immediate way to understand what these metrics represent or how they are computed. Adding explanatory hover icons to cards and tables will provide instant, self-service clarification across the interface while strictly adhering to repository standards (zero emojis).

*GOAL*

Build a reusable, zero-emoji contextual hover information icon component (`InfoTooltip` / `InfoIcon`) and integrate it across all dashboard KPI cards, distribution cards, feature navigation cards, and table headers to display clear explanations of displayed content on hover.

*VERTICAL SLICE SCOPE*

* Presentation Components:
  - Create `InfoIcon.tsx` (or `InfoTooltip.tsx`) rendering an accessible information symbol (clean text `[i]` or SVG information circular glyph, strictly zero emojis).
  - Support configurable tooltip placement (top, bottom, left, right) and accessible ARIA attributes (`aria-label`, `role="tooltip"`).
  - Add smooth CSS hover transitions and high-contrast tooltip styling in `index.css`.
* View Integrations:
  - Add information icons to all `KPICard` instances across Fleet Overview, Ticket Deep Dive, and Metric Progression.
  - Add information icons to `DistributionCard` (Role Cost Distribution, Model Cost Distribution).
  - Add information icons to Landing Page pulse cards and feature navigation cards.
  - Add information icons to table column headers in `TicketTable.tsx`, `LandingPage.tsx` Quick-Launch table, and `TicketComparisons.tsx` matrix table.
* Quality:
  - Unit tests in `Components.test.tsx` verifying that `InfoIcon` renders correctly, expands tooltip content on hover/focus, and contains appropriate ARIA attributes.
  - Visual regression and keyboard accessibility tests ensuring tooltips do not overflow viewport bounds.

*ACCEPTANCE CRITERIA*

* Every KPI card, distribution card, and major table header possesses an information icon.
* Hovering over or focusing on the information icon reveals an explanatory tooltip describing the metric or column calculation.
* Zero emojis used in icon rendering, tooltip text, or markup.
* Tooltips remain responsive and readable across desktop and tablet viewports without clipping.
* `npm test`, `npm run lint`, and `npm run typecheck` pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
