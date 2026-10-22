h3. Title: AL-26: Fleet Overview Completion Date Filtering and Autopsy Date Sorting

*TARGET*: AgenticLens (domain | adapters | presentation | tests)

*CONTEXT*

Currently, Fleet Overview calculates fleet-wide metrics across all recorded workspaces for project lifetime with zero temporal filtering. Users cannot isolate metrics to a specific release cycle, sprint, or calendar range. In addition, when browsing Workspace Autopsy in Ticket Deep Dive, users need the ability to toggle sorting between natural ticket identifier sequence (`<PREFIX>-NN`) and actual completion date (`completed_at`).

*GOAL*

Introduce completion date filtering on Fleet Overview defaulting to project lifetime, and add an optional completion date sorting toggle for the Ticket Deep Dive Workspace Autopsy selector.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Add date range filtering functions in Go `internal/domain`: `FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger`.
  - Treat `completed_at` as the primary date; fallback to `updated_at` (or `started_at`) if `completed_at` is null.
  - Implement completion date sorting for ticket lists (`completed_at` descending, nulls last).
* Boundary Adapter:
  - Update `/api/fleet` and `/api/workspaces` endpoints to accept optional `window`, `since`, and `until` query parameters.
  - Wire repository scanner to filter ledgers according to query parameters.
* Presentation:
  - Add date range filter dropdown in `FleetOverview.tsx` (presets: Project Lifetime [default], Last 7 Days, Last 30 Days, Last 90 Days, Custom Range).
  - Update `TicketDeepDive.tsx` and `TicketSelector.tsx` to support sorting by natural ticket ID (`<PREFIX>-NN`, default) or by completion date (`completed_at` descending).
  - Ensure reactive reloading and state synchronization when date filters change.
* Quality:
  - Go unit tests verifying date boundary conditions (inclusive timestamps, null `completed_at` handling, leap years, timezone offsets).
  - TypeScript tests in `FleetOverview.test.tsx` and `TicketDeepDive.test.tsx` verifying filter change handlers and autopsy sorting options.

*ACCEPTANCE CRITERIA*

* Fleet Overview defaults to Project Lifetime and correctly filters KPIs, distributions, and directory table when a date range is selected.
* Workspaces with null `completed_at` fall back gracefully to `updated_at` without throwing runtime panics or errors.
* Workspace Autopsy selector in Ticket Deep Dive sorts by `<PREFIX>-NN` by default and allows sorting by completion date upon toggle.
* Zero `any` types; all Go and TypeScript tests pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
