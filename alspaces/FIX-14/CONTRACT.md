# FIX-14: Restore the CI Single-Binary Build

Ticket: FIX-14. Classification: bugfix; conventional commit and telemetry type: `fix`.

## Problem

The frontend bundle builds, but its copy step fails when `backend/internal/adapters/web/dist/` does not exist. Go embeds this tree with `//go:embed all:dist`, so the directory and generated assets must be staged before Go binaries are compiled.

## Contract

- `make build` succeeds from a clean checkout when frontend dependencies are installed.
- The frontend build creates its staging destination before clearing and copying the generated bundle.
- Repeated builds refresh the staged `index.html` and assets required by the embedded filesystem.
- Existing behavior when frontend dependencies are absent remains unchanged.

## Scope

- Change only the frontend build staging script unless verification reveals a direct need for another build-layer adjustment.
- Keep Go embed behavior, runtime asset serving, application UI behavior, and release packaging unchanged.

## Acceptance

1. Missing destination no longer causes the frontend copy step to fail.
2. A root `make build` stages frontend assets before Go compilation.
3. Staged assets include non-empty HTML, JS, and CSS files.
4. Repeated builds succeed and refresh generated assets.
