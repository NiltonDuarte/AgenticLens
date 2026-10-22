# feat(api,web): dynamic config rest api endpoint and dashboard branding (AL-40)

## Overview

Implements the dynamic configuration REST API endpoint `GET /api/config` and observability dashboard branding for AgenticLens, resolving pre-release source code audit recommendations (AL-35 Section 3 & Section 4 Recommendation 2). The HTTP server now exposes sanitized project configuration (`project_name`, `prefixes`, `workspaces_dir`, `role_order`, and `version`), while the frontend dynamically hydrates this configuration, displays the observed project name alongside permanent `AgenticLens Telemetry` platform branding, and populates multi-select prefix dropdowns with both configured and discovered prefixes.

- Contract: [workspaces/AL-40/CONTRACT.md](workspaces/AL-40/CONTRACT.md)
- Telemetry: [workspaces/AL-40/SUMMARY.md](workspaces/AL-40/SUMMARY.md)

## Summary of Changes

### Adapters & HTTP Server Layer (`backend/internal/adapters/http/`, `backend/cmd/`)
- `dto.go`: Added `ProjectConfigDTO` with JSON serialization tags and added it to the `ResponsePayload` type union with typed `WriteProjectConfigResponse`.
- `config_handler.go`: Created `ConfigHandler` returning sanitized project parameters and build version while omitting sensitive internals (database paths, query timeouts, git mappings).
- `server.go`: Added `projectConfig` to `Server`, introduced `NewServerWithConfig`, registered `GET /api/config`, and preserved backwards compatibility in `NewServer`.
- `backend/cmd/agentic-lens/main.go`: Instantiated server via `NewServerWithConfig` passing parsed `projCfg`.

### TypeScript Domain & Boundary Client (`web/src/domain/`, `web/src/api/`)
- `types.ts`: Defined `TicketPrefixConfig` and `ProjectConfigResponse` domain interfaces.
- `formatters.ts`: Implemented `combineDistinctPrefixes` uniting discovered and configured prefixes with deduplication, case normalization, and default fallback (`['AL', 'GEN', 'PO']`).
- `mockData.ts`: Defined `mockProjectConfig` matching `ProjectConfigResponse`.
- `client.ts`: Implemented `fetchProjectConfig()` with graceful fallback to mock data on network errors, aborts, 500, or 404 responses.

### Presentation Shell & Views (`web/src/components/`, `web/src/views/`, `web/src/`)
- `Navbar.tsx`: Retained permanent `AgenticLens Telemetry` core platform brand and rendered observed project name in adjacent styled badge (`.project-badge`) with fallback fetch hydration.
- `LandingPage.tsx`: Rendered observed project context banner in the hero section when configured.
- `FleetOverview.tsx`, `TicketComparisons.tsx`, `MetricProgression.tsx`: Integrated `combineDistinctPrefixes` uniting known and configured prefixes for filter controls.
- `App.tsx`: Hydrated configuration on startup and wired project context to navbar, landing page, and analytical views.

### Quality & Verification
- Go HTTP tests in `server_test.go` covering 200 OK responses, default fallbacks, custom configurations, whitespace trimming, and non-GET rejection.
- TypeScript test suites in `formatters.test.ts`, `client.test.ts`, `Navbar.test.tsx`, `LandingPage.test.tsx`, `FleetOverview.test.tsx`, `TicketComparisons.test.tsx`, and `MetricProgression.test.tsx`.
- All quality gates pass (`make check`, `make test`, `make build`) with zero `any` and zero emojis.
