# FIX-15 Work Summary

## Changes

### Build and CI orchestration

- Added a `web-assets` Make target that builds the Vite frontend bundle needed by Go's `//go:embed all:dist` directive.
- Made `vet` depend on `web-assets`, so `make check` prepares embedded assets before Go vet.
- Added `backend-check` as the local and CI backend-only check entry point, combining vet and race-enabled tests.
- Made `test-backend` depend on `web-assets` so local Go tests work from a clean asset state.
- Updated the CI backend job to use `make backend-check` rather than maintaining a separate asset build and direct Go commands.
- Reviewed all four checked-in GitHub Actions workflows. Findings are recorded in `iterations/01/REVIEW.md` and `iterations/02/REVIEW.md`.

## Verification

- `npm ci` and `make check` passed in iteration 01.
- With generated embed assets moved aside, `make backend-check` rebuilt them, passed `go vet ./...`, and passed `go test -race ./...` across all backend packages in iteration 02.
- Frontend tests were not run.
