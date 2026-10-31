# Iteration 01 Plan: Prepare Embedded Assets Before Backend CI Checks

## Grounding

- `.github/workflows/ci.yml` currently runs Go vet and Go race tests in `backend-check` without building the frontend. The separate `frontend-check` job runs concurrently, so its output is not available to the backend job.
- `backend/internal/adapters/web/embed.go` declares `//go:embed all:dist`.
- `backend/internal/adapters/web/embed_test.go` checks for embedded JavaScript and CSS bundles, so a placeholder directory alone would make backend tests fail.
- `web/package.json` defines `npm run build` to compile the dashboard and copy its generated assets into `backend/internal/adapters/web/dist/`.
- `.github/workflows/ci.yml` already provisions Node 20 and runs `npm ci` in `frontend-check`.

## Implementation Tasks

1. Update `backend-check` in `.github/workflows/ci.yml` to provision Node.js 20, install dependencies from `web/package-lock.json`, and run the existing frontend build before Go vet and race tests.
2. Keep the existing Go toolchain setup, Go checks, frontend-check job, and downstream binary build behavior unchanged.
3. Do not change the embed directive, embed tests, runtime serving behavior, or the frontend build script.

## Verification

- Confirm workflow ordering places frontend compilation before `go vet ./...` and `go test -v -race ./...` in `backend-check`.
- Run the existing frontend build, then run `go vet ./...` and `go test ./...` from `backend/`.
- Confirm Go tests observe actual JS/CSS assets via `embed_test.go`.

## Acceptance

A clean CI checkout has the compiled frontend at the Go embed path before backend package analysis and tests run, while production continues serving assets embedded in the single binary.
