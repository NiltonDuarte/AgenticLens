# FIX-13 Contract: Fix Go CI Checks for Embedded Frontend Assets

## Problem

The backend CI job runs `go vet ./...` and `go test ./...` from a clean checkout before any frontend bundle exists. `backend/internal/adapters/web/embed.go` embeds `dist` using `//go:embed all:dist`, so Go rejects package loading when that directory is missing. The embed test also expects compiled JavaScript and CSS assets.

## Acceptance Boundaries

- Backend CI must generate the frontend bundle before running Go vet and backend tests.
- The production frontend remains embedded through `embed.FS`; no runtime asset lookup or weakened embed behavior is introduced.
- Existing frontend checks and the downstream single-binary build remain intact.
- Keep changes limited to CI workflow configuration and the required ticket artifacts.

## Architecture

This is an imperative build orchestration change. Frontend compilation remains in the existing `web` package build script, and the backend embedding boundary remains `backend/internal/adapters/web/embed.go`. No domain logic or application behavior changes.

## Verification

Run the frontend build followed by `go vet ./...` and `go test ./...` from `backend/`, matching the CI order. Review the workflow to ensure the asset generation precedes both Go checks.
