# feat(fleet): completion date filtering and autopsy date sorting (AL-26)

## Overview
Implements fleet-wide completion date filtering on Fleet Overview and completion date sorting for the Workspace Autopsy selector in Ticket Deep Dive.

- **Contract Baseline**: `workspaces/AL-26/CONTRACT.md`
- **Detailed Summary & Telemetry**: `workspaces/AL-26/SUMMARY.md`

## Key Changes
### Domain Core (`backend/internal/domain/`)
- Added `SortFieldCompletedAt` to `LedgerSortField` and updated `ParseLedgerSortField`.
- Implemented pure `GetLedgerCompletionDate` with deterministic fallback hierarchy (`completed_at` -> `updated_at` -> `started_at`).
- Implemented pure `FilterLedgersByCompletionDate` evaluating inclusive temporal bounds.
- Extended `SortLedgers` with `SortFieldCompletedAt` sorting descending by default, positioning nulls last, and tie-breaking on identical timestamps by natural ticket ID.

### Boundary Adapters (`backend/internal/adapters/http/`)
- Updated `/api/fleet` and `/api/workspaces` endpoints to parse optional `window`, `since`, and `until` query parameters.
- Validated date parameters (RFC 3339, RFC 3339 Nano, YYYY-MM-DD), returning structured HTTP 400 Bad Request on invalid date formats.
- Supported `sort=completed_at` in `/api/workspaces`.

### Presentation Layer (`web/src/`)
- Added date range filter control bar to Fleet Overview (`FleetOverview.tsx`) supporting presets (Project Lifetime [default], Last 7 Days, Last 30 Days, Last 90 Days, Custom Range) and custom date pickers with reactive reloads.
- Updated Ticket Deep Dive Workspace Autopsy selector (`TicketSelector.tsx`) with an interactive sorting toggle between Natural Ticket ID (`<PREFIX>-NN`, default) and Completion Date (`completed_at` descending, nulls last).
- Added pure date comparator `compareTicketCompletionDate` and client-side filter helper `filterLedgersByCompletionDate` in `formatters.ts`.
