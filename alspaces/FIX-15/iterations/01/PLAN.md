# FIX-15 Iteration 01: Audit and Repair CI Asset Preparation

Classification: `fix`. Baseline: `alspaces/FIX-15/CONTRACT.md`.

## Grounding

- `backend/internal/adapters/web/embed.go` embeds `all:dist`, requiring the directory to exist whenever Go loads that package.
- `make check` invokes `go vet ./...` and currently does not build or stage web assets.
- Review each workflow under `.github/workflows/`, including CI, release candidate, stable release, and release-please, for clean-checkout ordering and conditional execution.

## Implementation

1. Compare each workflow's install, frontend build, Go vet/test/build, and release-build steps.
2. Make `make check` prepare the frontend assets before Go vet, so release workflows cannot omit the required step.
3. Keep generated assets untracked and fail early with a clear instruction when frontend dependencies are unavailable.
4. Record workflow-by-workflow findings in the iteration review.

## Verification

- Reproduce the clean-state failure by removing only generated embedded assets, then run the selected check path.
- Run `make check` after the fix.
- Review all workflow steps and validate YAML syntax where a local validator is available.
- Run relevant build checks to ensure the embed tree remains available to binary compilation.
