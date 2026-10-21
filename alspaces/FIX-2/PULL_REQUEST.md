# fix(web,api): remove mockData from source and establish testdata workspaces harness (FIX-2)

## Overview
Eliminates over 2,100 lines of hardcoded synthetic mock fixtures from the frontend production source tree (`web/src/api/mockData.ts`) and removes silent mock fallback logic from `web/src/api/client.ts`. Previously, API failures silently returned synthetic fixtures tagged with `isMock: true`, masking operational failures and bundling mock fixtures directly into production binaries.

This PR establishes a dedicated filesystem test harness under `testdata/workspaces/` (for running the Go server in test/mock mode via `--workspaces testdata/workspaces`), relocates test fixtures strictly to `web/src/test/fixtures/mockData.ts` for Vitest, and enforces strict isolation: when pointing to production `workspaces.dir`, mock data is never loaded under any circumstances.

- Contract: [workspaces/FIX-2/CONTRACT.md](workspaces/FIX-2/CONTRACT.md)
- Summary: [workspaces/FIX-2/SUMMARY.md](workspaces/FIX-2/SUMMARY.md)

## Summary of Changes

### Boundary Adapters Layer (`web/src/api/`, `backend/internal/adapters/`)
- Deleted `web/src/api/mockData.ts` completely from the production source tree.
- Relocated static test fixtures to `web/src/test/fixtures/mockData.ts` for unit test suites.
- Refactored `web/src/api/client.ts`: removed all mock data imports, eliminated `getFilteredMockLedgers`, and converted all silent fallback returns into typed `Error` rejections (`HTTP error <status>: <statusText>` or structured domain errors).
- Created dedicated filesystem test data harness under `testdata/workspaces/` with valid schema v3 ledgers, contracts, and summaries for `AL-1`, `AL-2`, `AL-3`, `PO-1`, and `GEN-1`.
- Added HTTP server integration tests in `backend/internal/adapters/http/server_test.go` verifying strict filesystem isolation between test workspaces and production workspaces.

### Presentation & Application Layer (`web/src/`, `backend/cmd/`)
- In `web/src/App.tsx`, derived `isMock` from `projectConfig?.workspaces_dir` (inspecting whether the configured workspace path targets `testdata` or `mock`).
- In `web/src/components/Navbar.tsx`, retained Live Server vs. Mock Mode indicator accurately reflecting backend server configuration.
- Added CLI integration tests in `backend/cmd/agentic-lens/main_test.go` validating that `--workspaces testdata/workspaces` scans test workspaces and detects anomalies without touching production data.
- Updated all 13 frontend test files to import fixtures from `../test/fixtures/mockData`.
- Rewrote `web/src/api/client.test.ts` to assert explicit error rejections on network drop, HTTP 500, HTTP 404, and aborted requests.
- Rebuilt embedded frontend bundle (`web/dist/`), dropping bundle size from ~620 kB to 446 kB with zero mock fixture bytes.

### Standards & Quality
- Adheres strictly to FCIS (Functional Core, Imperative Shell).
- Zero `any` or empty `interface{}` across Go backend and TypeScript frontend.
- Zero emojis across all code, tests, UI strings, and documentation.
