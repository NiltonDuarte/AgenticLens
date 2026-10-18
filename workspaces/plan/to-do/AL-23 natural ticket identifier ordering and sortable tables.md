h3. Title: AL-23: Natural Ticket Identifier Ordering and Sortable Tables

*TARGET*: AgenticLens (domain | presentation | adapters | tests)

*CONTEXT*

Currently, sorting ticket identifiers relies on standard alphanumeric string comparison (`sort.Strings` in Go and `localeCompare` in TypeScript), producing distorted ordering sequences such as `AL-1`, `AL-10`, `AL-11`, `AL-2`, `AL-22`. Users navigating workspaces across the Landing Page, Fleet Overview, Ticket Deep Dive, and Comparisons matrix experience broken ordering where low-numbered tickets are scattered. Furthermore, the Quick-Launch Workspaces directory on the Landing Page is completely static and cannot be sorted by column headers.

*GOAL*

Implement pure domain natural ticket identifier parsing and comparison (`<PREFIX>-NN`, `<PREFIX>NN`) across Go and TypeScript, wire interactive column sorting to the Landing Page Quick-Launch Workspaces table, and enforce natural ordering across all dashboard tables and ticket selectors.

*VERTICAL SLICE SCOPE*

* Domain Core:
  - Implement pure Go comparator `CompareTicketIDs(a, b string) int` parsing ticket prefix (alphabetical) and numeric suffix (numerical integer comparison). Handle variations like `AL-1`, `AL1`, `GEN-42`, `PO-5`.
  - Add `SortFieldTicketID` to Go domain sorting functions.
* Boundary Adapter:
  - Wire natural ticket ordering into `adapters.ScanWorkspaces` and `/api/workspaces` response generation.
* Presentation:
  - Implement TypeScript natural comparator `compareTicketIds(a: string, b: string): number` in `web/src/domain/formatters.ts`.
  - Add interactive column sorting to `LandingPage.tsx` Quick-Launch Workspaces table (`Ticket ID`, `Title`, `Status`, `Iterations`, `Active Duration`), using natural sort for `Ticket ID`.
  - Update `TicketTable.tsx` to sort by `ticket_id` using `compareTicketIds`.
  - Update `TicketSelector.tsx` to display tickets in natural order by default.
  - Update `deriveTicketComparisonRows` tie-breaker to sort by natural ticket ID order.
* Quality:
  - Go table-driven unit tests for `CompareTicketIDs` covering edge cases (`AL-2` vs `AL-10`, `AL-1` vs `AL-01`, mixed prefixes, missing numbers, empty strings).
  - TypeScript unit tests in `formatters.test.ts`, `LandingPage.test.tsx`, `TicketTable.test.tsx`, and `TicketSelector.test.tsx`.

*ACCEPTANCE CRITERIA*

* Sorting tickets with identical prefix produces strictly numerical order: `AL-1`, `AL-2`, `AL-3`, ..., `AL-9`, `AL-10`, `AL-11`, `AL-22`.
* Multi-prefix lists group by prefix alphabetically and then sort numerically within each prefix (`AL-1`, `AL-2`, `GEN-1`, `GEN-10`, `PO-1`).
* Landing Page Quick-Launch Workspaces table headers are clickable, sortable ascending and descending with visible indicators.
* Fleet Overview Ticket Workspaces Directory sorts by `<PREFIX>-NN` when sorted by Ticket ID.
* Ticket Deep Dive Workspace Autopsy dropdown presents tickets in `<PREFIX>-NN` order by default.
* `make test-backend`, `npm test`, `make check`, and `make build` pass cleanly.

*DEFINITION OF DONE*

* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
