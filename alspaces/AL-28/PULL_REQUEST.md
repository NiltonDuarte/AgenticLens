# Pull Request: AL-28 Multi-Dimensional Filtering by Ticket Type and Workspace Prefix

## Title
feat(filtering): multi-dimensional filtering by ticket type and workspace prefix (AL-28)

## Overview
Implements multi-dimensional filtering across ticket types and workspace prefixes throughout the AgenticLens Go backend and TypeScript SPA dashboard, defaulting non-destructively to `all` and supporting arbitrary multi-selection across both dimensions.

Refer to architectural baseline in [CONTRACT.md](workspaces/AL-28/CONTRACT.md) and layer-by-layer audit in [SUMMARY.md](workspaces/AL-28/SUMMARY.md).

## Key Changes by Layer

### 1. Functional Core (Go)
- `backend/internal/domain/models.go`: Added `TicketType *string` to `TicketComparisonRow` with `ticket_type,omitempty` JSON tag.
- `backend/internal/domain/aggregations.go`:
  - `CanonicalTicketType`: normalizes conventional aliases (`feat` -> `feature`, `bug`/`fix`/`bugfix` -> `bug`, `doc`/`docs` -> `docs`).
  - `ExtractTicketPrefix`: extracts uppercase ticket prefix from conforming ticket identifiers.
  - `ExtractDistinctPrefixes`: extracts and sorts distinct non-empty uppercase ticket prefixes.
  - `FilterLedgersByTicketTypes`: pure filter supporting multi-token matching, case-insensitivity, and pass-through on `all` or empty slice.
  - `FilterLedgersByPrefixes`: pure filter matching workspace prefixes case-insensitively with `all` pass-through.
  - Updated `BuildComparisonRows` to propagate `TicketType`.

### 2. Imperative Shell Adapters (Go)
- `backend/internal/adapters/http/dto.go`: Added `[]domain.TicketComparisonRow` to `ResponsePayload` and added `WriteComparisonResponse`.
- `backend/internal/adapters/http/handlers.go`:
  - Added `parseMultiQueryParam` helper handling comma-separated lists and repeated query keys.
  - Extended `WorkspacesHandler` and `FleetHandler` to support `ticket_type` and `prefix` filtering.
  - Added `ComparisonHandler` for `GET /api/comparison` and `ProgressionHandler` for `GET /api/progression`.
- `backend/internal/adapters/http/server.go`: Registered dedicated routes `GET /api/comparison` and `GET /api/progression`.

### 3. Presentation & Client (TypeScript / React)
- `web/src/domain/types.ts`: Defined `StandardTicketType` union and added `ticket_type?: string` to `TicketComparisonRow`.
- `web/src/domain/formatters.ts`: Pure mirror functions for alias normalization, prefix extraction, and multi-selection filtering.
- `web/src/api/client.ts`: Extended API client with multi-parameter URL formatting and fallback filtering for comparison and progression endpoints.
- `web/src/components/MultiSelectDropdown.tsx`: Reusable accessible combobox dropdown with Escape key handling, outside click detection, and "Select All" toggle.
- `web/src/views/FleetOverview.tsx`: Integrated multi-select controls into filter bar.
- `web/src/views/TicketComparisons.tsx`: Unconditionally mounted `table-controls-bar` above empty states and table rows; interactive re-selection without page reload.
- `web/src/views/MetricProgression.tsx`: Unconditionally mounted `progression-controls-bar` inside root container above empty state; filter dropdowns remain interactive during zero-match queries for seamless recovery without page reload.
