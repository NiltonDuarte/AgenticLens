# Pull Request: AL-5 Single Binary Integration Parity and Deprecation

## Description
This pull request completes AL-5 by verifying cross-engine numerical and categorical parity between the Go functional core and the legacy Python implementation, relocating the Go codebase into a dedicated `backend/` directory, updating the build system and Dockerfile to produce a standalone single binary with embedded web assets, and permanently removing all Python artifacts from the repository.

## Key Changes
1. **Cross-Engine Parity Verification**:
   - Implemented table-driven parity test suite in `backend/internal/adapters/parity_test.go`.
   - Validated 100% numerical and categorical match against 12 real workspace ledgers (`AL-1`, `AL-2`, `AL-3`, `AL-4`, `AL-10`, `AL-11`, `AL-12`, `GEN-1`, `GEN-2`, `PO-1`, `PO-2`, `PO-3`) and synthetic edge cases.
   - Tested totals, workflow efficiency metrics, fleet summaries, and comparison tables.
2. **Go Codebase Relocation to `backend/`**:
   - Moved `cmd/` -> `backend/cmd/`.
   - Moved `internal/` -> `backend/internal/`.
   - Moved `go.mod` and `go.sum` -> `backend/`.
   - Updated `.agenticlens/config.json` `src_folders` to `["web", "backend"]`.
   - Updated root `Makefile` targets (`vet`, `test-backend`, `build`, `server`) to build inside `backend/` and output the binary to root `bin/agentic-lens`.
   - Updated `web/package.json` build step to target `backend/internal/adapters/web/dist`.
   - Updated `.gitignore` to protect embedded dist files under `!backend/internal/adapters/web/dist/`.
   - Updated Dockerfile to a multi-stage build (Node Alpine -> Go Alpine -> Alpine runtime) with zero Python and Node.js in the final container.
3. **Decommissioning Legacy Python**:
   - Deleted `agentic_lens/` presentation, domain, and adapter packages.
   - Deleted `scripts/extract_agentic_metrics.py`.
   - Deleted `tests/` Python test suite.
   - Deleted `pyproject.toml`, `uv.lock`, and related cache files.
4. **Standalone Binary Serving**:
   - Confirmed `./bin/agentic-lens server` serves embedded frontend assets, handles client-side SPA fallback routing, and serves JSON API endpoints with zero runtime external dependencies.

## Verification
- `make check`: Passed (`git diff --check`, Go vet).
- `make test-backend`: Passed with race detector enabled (`go test -race ./...` in `backend/`).
- `make build`: Standalone binary compiled to `bin/agentic-lens`.
- Integration tests verified via `curl` against `GET /`, `GET /deep-dive`, `GET /assets/*`, `GET /api/workspaces`, and `GET /api/nonexistent`.
