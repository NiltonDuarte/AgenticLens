h3. Title: AL2-7: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation

*TARGET*: AgenticLens (presentation | web | components | state | analytics | tests)

*CONTEXT*

With foundational rate-card configuration and domain cost approximation established in AL2-6, operators require multi-dimensional visualization and interactive controls across the web dashboard. Specifically, fleet overview aggregations, progression timelines, and cross-ticket comparison scatter plots currently operate solely on measured spend, preventing comparative ROI evaluations between teams using pay-per-token API keys and teams operating under flat-rate subscription models.

*GOAL*

Deliver rich interactive cost observability across the TypeScript dashboard, introducing a global Cost Mode toggle (`Measured Spend`, `Estimated Equivalent`, `Combined Hybrid`), dynamic scatter plot and progression re-projections, and an interactive rate-card simulation drawer allowing engineers to model pricing changes on historical token traces.

*VERTICAL SLICE SCOPE*

* Domain Core & Aggregations: Extend fleet rollouts and comparison projection helpers in `internal/domain` to support parameterized cost mode selection (`CostDisplayMode`). Provide pure projection functions for fleet-level measured vs estimated vs hybrid totals.
* Boundary Adapter & API: Extend HTTP query parameters on `/api/fleet`, `/api/workspaces`, and comparison endpoints to support `?cost_mode=measured|estimated|hybrid` and simulation overrides.
* Web Presentation & State: Implement global `CostModeSelector` component in `AppHeader.tsx` or view action bars, persisting preference in local storage. Update `FleetOverview.tsx`, `TicketProgressionChart.tsx`, and `TicketComparisons.tsx` (scatter plot X-axis and comparison cards) to dynamically switch axes and series data based on selected cost mode. Add `RateCardDrawer.tsx` to simulate alternate model pricing scenarios in real time.
* Quality: Vitest component and interaction tests for `CostModeSelector`, chart re-projections, and `RateCardDrawer`. Full typecheck and verification with `make`.

*ACCEPTANCE CRITERIA*

* Operators can toggle between `Measured Spend`, `Estimated Equivalent`, and `Combined Hybrid` across Fleet Overview and Ticket Comparisons.
* Scatter plot X-axis and data points dynamically re-scale and update when switching cost modes.
* Chart tooltips and comparison cards explicitly indicate whether displayed values reflect measured spend, estimated rate cards, or a hybrid blend.
* Rate card simulation drawer allows testing alternate rate values against workspace token totals without mutating configuration files.
* TypeScript strict mode passes with zero `any`; full test suite passes with `make`.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make`).
* Go and TypeScript type checking clean with zero any.
* End-to-end dashboard verified with embedded binary.
