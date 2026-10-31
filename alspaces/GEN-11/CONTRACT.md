# GEN-11 Technical Contract

## Problem

The generated frontend bundle under `backend/internal/adapters/web/dist/` is currently tracked. The task is to remove its generated files from version control and ignore future output at that exact path.

## Acceptance boundaries

- Remove tracked files under `backend/internal/adapters/web/dist/` from the repository.
- Add a `.gitignore` rule scoped to `backend/internal/adapters/web/dist/`, without ignoring other directories named `dist`.
- Keep the frontend source and its build process unchanged.
- Do not change the separate `AL-50` checkout. Implement this as the isolated `GEN-11` chore.

## Architecture and build constraints

- `backend/internal/adapters/web/embed.go` embeds `all:dist`, so Go compilation requires this directory to exist.
- `web/package.json` builds frontend assets into that embedded directory; the root `Makefile` runs the frontend build during build targets when `web/node_modules` exists.
- The cleanup must preserve this generation-before-compilation relationship. Do not change Go embedding behavior or add runtime dependencies as part of this ticket.

## Scope boundaries

- No Go or TypeScript source changes are required.
- No domain, adapter, HTTP, or UI behavior changes are permitted.
- Do not remove frontend source assets or generated output outside the specified embedded directory.
