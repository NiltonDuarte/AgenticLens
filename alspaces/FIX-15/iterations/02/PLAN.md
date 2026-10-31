# FIX-15 Iteration 02: Prepare Embed Assets Before Backend Go Tests

Classification: `fix`. Baseline: `alspaces/FIX-15/CONTRACT.md` and iteration 01.

## Finding

A clean checkout running standalone `go test ./...` in `backend/` fails during package setup because `//go:embed all:dist` has no matching asset directory. The first iteration made `make check` prepare assets before Go vet, but it did not align the standalone Go test path or a backend-only Make target.

## Implementation

1. Add a backend-only Go check target that prepares frontend assets before invoking Go vet and tests.
2. Ensure `make test-backend` shares the same asset prerequisite while preserving its race detector behavior.
3. Keep direct `go test ./...` documented as requiring generated embed assets; update CI to use the shared backend-only target where it currently invokes Go tooling directly, or otherwise ensure the same prerequisite is established first.
4. Update the workflow review with the revised local and CI invocation contract.

## Verification

- Remove only generated embedded assets and invoke the backend-only check target to verify it recreates them before Go package loading.
- Run the backend-only check and race-enabled test target.
- Confirm CI uses the shared prerequisite path and review all workflows again.
