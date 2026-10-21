h3. Title: AL-28: Multi-Dimensional Filtering by Ticket Type and Workspace Prefix

*TARGET*: AgenticLens (domain | adapters | presentation | tests)

*CONTEXT*

When engineering fleets scale across diverse ticket initiatives, metrics from feature development, bug fixes, automated refactorings, and documentation/PO sessions become mixed together. Users browsing Fleet Overview, Ticket Comparisons, and Metric Progression currently have no way to isolate metrics by ticket type (e.g., `feature`, `fix`, `chore`, `docs`, `refactor`) or workspace prefix (e.g., `AL`, `GEN`, `PO`). Adding multi-dimensional filtering defaulting to `ticket_type: "feature"` allows teams to inspect focused development velocity.

*GOAL*

Implement filtering by `ticket_type` and `workspace prefix` across Fleet Overview, Ticket Comparisons, and Metric Progression, defaulting `ticket_type` to `"feature"` while allowing full exploration of all ticket categories.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Add domain filtering functions: `FilterLedgersByTicketType(ledgers []AgenticMetricsLedger, ticketType string) []AgenticMetricsLedger` (case-insensitive with conventional aliases e.g. `feat`/`feature`, `bug`/`fix`).
  - Add domain filtering functions: `FilterLedgersByPrefix(ledgers []AgenticMetricsLedger, prefix string) []AgenticMetricsLedger`.
* Boundary Adapter:
  - Add `ticket_type` and `prefix` query parameters to `/api/fleet`, `/api/workspaces`, `/api/comparison`, and `/api/progression` endpoints.
* Presentation:
  - Add `Ticket Type` filter dropdown to Fleet Overview, Ticket Comparisons, and Metric Progression, defaulting to `"feature"`.
  - Add `Workspace Prefix` filter dropdown (dynamically populated with distinct prefixes discovered from ledgers, e.g. `All`, `AL`, `GEN`, `PO`), defaulting to `All`.
  - Update view state, fetch calls, and table/chart renders to respect both filters in unison.
* Quality:
  - Pure Go domain unit tests for ticket type and prefix filtering (aliasing, case sensitivity, empty inputs, non-matching tickets).
  - TypeScript integration tests for filter dropdowns across `FleetOverview.test.tsx`, `TicketComparisons.test.tsx`, and `MetricProgression.test.tsx`.

*ACCEPTANCE CRITERIA*

* Fleet Overview, Comparisons, and Metric Progression default to `ticket_type: "feature"` upon initial load.
* Users can change `ticket_type` to `All Types`, `bug`/`fix`, `chore`, `docs`, `refactor`, or any custom type present in ledgers.
* Users can filter by workspace prefix (`AL`, `GEN`, `PO`, or `All Prefixes`).
* Filter states update KPI metrics, distribution cards, comparison rows, and progression trends reactively.
* Zero `any` types; all quality gates pass.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
