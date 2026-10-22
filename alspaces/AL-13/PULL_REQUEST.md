# Pull Request: AL-13 Semantic Versioning and Automated CI Pipeline

## Description
This pull request implements ticket AL-13, establishing compile-time semantic versioning, an HTTP version endpoint, CLI version reporting with multiple output formats, a frontend navigation version badge with tooltip and fallback, and an automated GitHub Actions CI workflow.

## References
- Architectural Contract: `workspaces/AL-13/CONTRACT.md`
- Detailed Delivery Summary: `workspaces/AL-13/SUMMARY.md`

## Summary of Changes
- **Domain Core (`backend/internal/domain/`)**: Added immutable `BuildInfo` value struct and pure formatters (`FormatBuildInfo`, `FormatShortVersion`, `GetBuildInfo`).
- **HTTP Endpoint (`backend/internal/adapters/http/`)**: Added `GET /api/version` handler returning structured JSON `BuildInfoDTO` (`version`, `commit`, `build_date`).
- **CLI Subcommand & Root Flags (`backend/cmd/agentic-lens/`)**: Added `version` subcommand supporting `--short` and `--json` flags, as well as `-v` and `--version` root flags.
- **Build System (`Makefile`)**: Added compile-time `-ldflags` injection for `Version`, `GitCommit`, and `BuildDate` into `agenticlens/internal/domain`.
- **Frontend Presentation (`web/`)**: Added `BuildInfo` type, `fetchVersion()` API client with offline fallback, and `.version-badge` element in `Navbar` with commit/build date tooltip.
- **CI Pipeline (`.github/workflows/ci.yml`)**: Added multi-job GitHub Actions workflow executing backend checks (`go vet`, `go test -race ./...`), frontend checks (`typecheck`, `eslint`, `vitest`), and binary compilation with version verification on pull requests and pushes to `main`.
