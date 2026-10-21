# feat(web): TypeScript SPA shell, Fleet Overview dashboard, and single-binary embedding (AL-3)

## Summary of Changes
Delivers the complete modern TypeScript web frontend application (`web/`) for AgenticLens, providing the Fleet Overview analytics dashboard and single-binary distribution via Go `embed.FS`.

- **Domain Core (`web/src/domain/`)**: Strictly typed domain models matching Go contracts (`internal/domain/models.go`) with value semantics and zero `any`. Implemented pure formatters for compact token notation, currency, percentages, and duration with zero-division defenses.
- **Boundary Adapters (`web/src/api/`)**: Resilient dual-mode API client querying live endpoints (`/api/fleet`, `/api/workspaces`, `/api/workspaces/:ticketId`) with automatic fallback to deterministic mock fixtures (`mockData.ts`) during offline development or network failure.
- **Presentation Shell (`web/src/components/`, `web/src/views/`)**:
  - `Navbar.tsx`: View routing state and live/mock connection pill.
  - `KPICard.tsx`: Metric cards for Cost, Invocations, Tokens, Cache Hit %, and Duration.
  - `DistributionCard.tsx`: Role and model cost distributions with proportional visual bars.
  - `TicketTable.tsx`: Full ticket directory table with real-time text search, multi-column sorting, and health badges.
  - `EmptyState.tsx`: Reusable zero-dataset placeholder.
  - `FleetOverview.tsx`: Root dashboard view coordinating parallel fetching, error handling with retry, and defensive empty states.
  - Application shell in `App.tsx` and styling in `index.css` using modern neutral dark theme.
- **Single-Binary Embedding (`internal/adapters/web/`)**: Frontend production build (`web/dist/`) bundled into Go binary via `embed.FS` in `internal/adapters/web/embed.go`. Tested single-binary execution (`agentic-lens server`) serving SPA assets and live REST API endpoints without Node.js runtime.

## References
- Architectural Contract: `workspaces/AL-3/CONTRACT.md`
- Iteration Plan: `workspaces/AL-3/iterations/01/PLAN.md`
- Review Audit: `workspaces/AL-3/iterations/01/REVIEW.md`
- Work & Telemetry Summary: `workspaces/AL-3/SUMMARY.md`
