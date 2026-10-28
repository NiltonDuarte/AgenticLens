# fix(config): configurable context ceilings and source roots consolidation (FIX-11)

## Summary

- Add `metrics.healthy_ceiling_tokens` to project configuration, defaulting to 200,000 tokens when non-positive.
- Dynamically compute elevated context health at `ceiling / 2` across domain health evaluators and workflow efficiency aggregations.
- Support configurable live anomaly detection thresholds for tool error streaks, ceiling approach ratios, token velocity runaway, and tool thrashing.
- Consolidate source code root boundaries under `code_metrics.source_roots` with bidirectional backward-compatible fallback for `workspaces.src_folders`.
- Update git diff classification and path monitoring to respect `code_metrics.source_roots`.

## Context

- Contract: `alspaces/FIX-11/CONTRACT.md`
- Summary: `alspaces/FIX-11/SUMMARY.md`
- Iteration Plan: `alspaces/FIX-11/iterations/01/PLAN.md`
- Iteration Review: `alspaces/FIX-11/iterations/01/REVIEW.md`

## Verification

- Full backend test suite with race detector: `go test -race ./...` (PASS)
- Frontend test suite: `vitest run` (32 files, 734 tests passed) (PASS)
- Linting and vetting: `make check` (PASS)
- Multi-platform binary compilation: `make build` (PASS)
