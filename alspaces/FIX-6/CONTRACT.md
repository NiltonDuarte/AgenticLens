# Technical Contract: Ticket FIX-6

## Problem Statement
Release Please manages automated releases and semantic versioning in AgenticLens. However, when Release Please created the v1.0.0 release (commit `968aa18`), no multi-platform binary assets, archives, or checksums were attached to the GitHub release.

Two architectural gaps caused this issue:
1. `.github/workflows/release-please.yml` only executed `googleapis/release-please-action@v4` without building or uploading assets.
2. Tagging semantics differed: Release Please was configured with `"package-name": "agentic-lens"` without `"include-component-in-tag": false`, causing tags to be created as `agentic-lens-v1.0.0` rather than `v1.0.0`. Even if tags were named `v*`, GitHub Actions does not trigger workflows from events initiated by GitHub Actions tokens (`GITHUB_TOKEN`), meaning downstream workflows such as `release-stable.yml` do not trigger on Release Please release commits or tags.

## Core Architectural Invariants

### 1. Functional Core, Imperative Shell (FCIS) Boundaries
- Pure Domain (`backend/internal/domain/`): Domain logic, metric aggregations, and session models remain untouched.
- Imperative Shell (`.github/workflows/release-please.yml`, `.release-please-config.json`): CI/CD release workflow acts as the automation shell. When a release is produced, it checks out the codebase, runs test and static analysis suites (`make check`, `make test`), cross-compiles release binaries (`make release-build`), packages archives (`make release-package`), verifies artifact signatures (`make release-verify`), and uploads them to the GitHub release.

### 2. Manual Dispatch & Historical Tag Support
- The updated workflow includes `workflow_dispatch` with an optional `force_tag` parameter. This allows operators to build and attach artifacts to existing releases (such as `agentic-lens-v1.0.0` or `v1.0.0`) retroactively if ever needed.

### 3. Verification & Safety Invariants
- All unit, boundary, and workflow verification tests in `backend/cmd/agentic-lens/release_verification_test.go` must pass without regressions.
- Zero emojis across workflow files, configuration, commit messages, and workspace deliverables.

## Scope Boundaries

### In-Scope Files
- `.github/workflows/release-please.yml`: Add build, package, verification, upload, floating latest tag update, and stale candidate pruning steps gated on release creation or manual dispatch.
- `.release-please-config.json`: Add `"include-component-in-tag": false` to align future release tags with standard `vX.Y.Z` semantics.
- `alspaces/FIX-6/`: Workspace deliverables (`CONTRACT.md`, `SUMMARY.md`, `PULL_REQUEST.md`, `agentic_metrics.json`).

### Out-of-Scope Files
- Backend application code and domain modules (`backend/internal/domain/`, `backend/internal/adapters/`).
- Frontend source files (`web/src/`).

## Acceptance Criteria
1. Release Please Workflow Asset Pipeline: `.github/workflows/release-please.yml` defines the complete asset build, package, verification, and upload lifecycle when a release is created or when manually dispatched.
2. Tag Alignment: `.release-please-config.json` specifies `"include-component-in-tag": false`.
3. Quality Gate: `make check` and `make test` pass cleanly with zero warnings and zero errors.
4. Workflow Verification Tests: `go test ./cmd/agentic-lens/...` passes all workflow invariant assertions.
5. Zero Emojis: Zero emojis across all modified code, configuration, and workspace files.
