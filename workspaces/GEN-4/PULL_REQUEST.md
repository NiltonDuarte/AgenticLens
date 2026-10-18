# GEN-4: Update TICKETS.md Status for Merged Workspaces and Configure Release-Please

## Summary
Updates `workspaces/plan/TICKETS.md` and ticket specification files for completed tickets AL-15 and AL-30 whose workspaces and implementations have already been merged into `main`. Configures automated semantic versioning and changelog generation via Google's `release-please` GitHub Action.

## Changes
- Moved `workspaces/plan/to-do/AL-15 implement universal multiversion schema migration engine for telemetry ledgers.md` to `workspaces/plan/done/`.
- Moved `workspaces/plan/to-do/AL-30 fail telemetry generation on missing root session.md` to `workspaces/plan/done/`.
- Updated `workspaces/plan/TICKETS.md`:
  - Updated individual ticket definition links to reflect `done/` locations for AL-15 and AL-30.
  - Removed AL-15 and AL-30 from `## Ready`.
  - Added AL-15 and AL-30 to `## Done` with `Status: Done` and updated file paths.
- Added `.release-please-config.json` configuring Go release type with `extra-files` for `web/package.json`.
- Added `.release-please-manifest.json` setting the root baseline version to `0.1.0`.
- Added `.github/workflows/release-please.yml` running on pushes to `main`.
- Added root `CHANGELOG.md` seed file for release-please automated tracking.
- Optimized `.github/workflows/release-candidate.yml` with path filtering (`backend/**`, `web/**`, `Makefile`, `go.work*`) and release commit filtering to prevent generating pre-releases on docs-only commits.
- Added sliding-window pruning to `.github/workflows/release-candidate.yml` keeping the latest 3 candidate pre-releases and cleaning older tags via GitHub CLI.
- Added automated sweep to `.github/workflows/release-stable.yml` purging all remaining pre-release candidates upon publishing a stable release.
- Optimized `.github/workflows/ci.yml` with concurrency cancellation on branches/PRs while preserving `main` runs, and restricted triggers strictly to `main` pushes and pull requests.
- Added cross-platform compilation to `Makefile` producing `bin/agentic-lens-linux-amd64`, `bin/agentic-lens-linux-arm64`, `bin/agentic-lens-darwin-arm64`, and `bin/agentic-lens-darwin-amd64`.
- Added architecture dispatcher wrapper script at `scripts/agentic-lens-wrapper.sh`, deployed to `bin/agentic-lens` during `make build` to seamlessly route executions on both host macOS and Linux containers.
- Allocated ticket workspace under `workspaces/GEN-4/`.
