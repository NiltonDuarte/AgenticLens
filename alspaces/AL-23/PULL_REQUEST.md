# Pull Request: AL-23 Natural Ticket Identifier Ordering and Sortable Tables

## Description
This pull request introduces pure natural ticket identifier comparison across the Go domain and TypeScript frontend, sets natural ordering as the default sequence for workspace discovery and REST API responses, and provides interactive multi-column sorting with visible direction indicators for the Landing Page Quick-Launch table.

For detailed architecture, contracts, and telemetry, refer to:
- Contract: [workspaces/AL-23/CONTRACT.md](workspaces/AL-23/CONTRACT.md)
- Summary: [workspaces/AL-23/SUMMARY.md](workspaces/AL-23/SUMMARY.md)

## Summary of Changes

### Domain
- Added `SortFieldTicketID` to `LedgerSortField` enum in Go (`backend/internal/domain/enums.go`) and updated `ParseLedgerSortField`.
- Implemented pure alphanumeric comparator `CompareTicketIDs(a, b string) int` in Go (`backend/internal/domain/aggregations.go`) parsing ticket prefix and numerical sequence to ensure natural progression (`AL-1`, `AL-2`, `AL-10`, `AL-22`).
- Extended Go `SortLedgers` to support `SortFieldTicketID` and updated `BuildComparisonRows` with natural tie-breaking.
- Implemented pure comparator `compareTicketIds(a: string, b: string): number` in TypeScript (`web/src/domain/formatters.ts`) mirroring Go logic.
- Updated `deriveTicketComparisonRows` to sort and tie-break comparison rows via `compareTicketIds`.

### Adapters
- Updated `ScanWorkspaces` in `backend/internal/adapters/repository.go` to sort discovered workspaces naturally by ticket ID (`SortFieldTicketID`, ascending) before returning `ScanReport`.
- Updated `WorkspacesHandler` in `backend/internal/adapters/http/handlers.go` to default to ascending order for `SortFieldTicketID`.
- Rebuilt embedded web assets into `backend/internal/adapters/web/dist/` for single-binary distribution.

### Presentation
- Updated `LandingPage.tsx` Quick-Launch table with interactive column sorting across Ticket ID, Title, Status, Iterations, and Active Duration with textual sort direction indicators (` (asc)`, ` (desc)`) and `aria-sort` attributes.
- Updated `TicketTable.tsx` to use `compareTicketIds` for ticket identifier column sorting and equal metric tie-breaking.
- Updated `TicketSelector.tsx` to memoize and sort autopsy dropdown options in natural order by default with fallback handling for untitled workspaces.
- Added table header styles in `web/src/index.css` for sortable columns.
