# fix(web,api): eliminate isMock branching and enforce mandatory workspaces dir configuration (FIX-2)

## Overview
Eliminates synthetic mock mode branching and the `isMock` boolean flag across the entire web application and API client, enforces mandatory workspace directory configuration in the Go server with zero fallback guessing, and displays the active workspace directory directly in the navigation bar.

The platform operates identically regardless of whether the configured workspace directory points to production (`workspaces`) or test data (`testdata/workspaces`), treating both uniformly as standard workspace directories without platform-level awareness or branching between mock and real data.

- Contract: [workspaces/FIX-2/CONTRACT.md](workspaces/FIX-2/CONTRACT.md)
- Summary: [workspaces/FIX-2/SUMMARY.md](workspaces/FIX-2/SUMMARY.md)

## Summary of Changes

### Go Backend & CLI Layer (`backend/cmd/`, `backend/internal/adapters/`)
- Enforced mandatory `workspaces.dir` configuration on server startup: `runServer` exits immediately with code 1 if `workspaces.dir` is omitted and `--workspaces` is not provided, or if the specified directory does not exist on disk.
- Updated `ResolveWorkspacesDir` in `repository.go`: returns `""` when neither flag nor config is set, eliminating implicit fallback guessing to `"workspaces"`.
- Updated `ScanWorkspaces`, `LoadTicketWorkspace`, and `ReadTicketArtifacts` to return explicit errors when the workspaces directory is unconfigured.
- Added CLI integration tests in `main_test.go` and unit tests in `repository_test.go` validating fail-fast startup and existence checks.

### Boundary Adapters & API Client (`web/src/api/`)
- Deleted `web/src/api/mockData.ts` and relocated test fixtures strictly to `web/src/test/fixtures/mockData.ts` for Vitest test suites.
- Removed `isMock` from `ApiResponse<T>`, `AsyncState<T>`, `ApiConfig`, and `ApiError`. All API methods return `{ data }` directly.
- All 11 endpoint functions throw typed `Error` instances on HTTP errors or network drops without silent synthetic fallbacks.

### Presentation & Application Layer (`web/src/`, `web/src/components/`, `web/src/views/`)
- Completely removed `isMock` state, `useMemo` calculation, and `handleMockChange` callback from `App.tsx`.
- Completely removed `onMockChange` prop and mock tracking references across `LandingPage`, `FleetOverview`, `TicketDeepDive`, `MetricProgression`, and `TicketComparisons`.
- Updated `Navbar.tsx`: removed "Live Server" and "Mock Mode" status pills. Displays the active workspace directory in a monospace workspace pill directly from `projectConfig.workspaces_dir`. Displays `"Loading..."` when missing or pending without guessing defaults.
- Updated styling in `web/src/index.css` for `.status-pill.workspace-pill` and loading state.
- Rebuilt embedded frontend bundle (`web/dist/`) with zero mock data and zero mock mode strings.

### Standards & Quality
- Adheres strictly to FCIS (Functional Core, Imperative Shell).
- Zero `any` or empty `interface{}` across Go backend and TypeScript frontend.
- Zero emojis across all code, tests, UI strings, and documentation.
