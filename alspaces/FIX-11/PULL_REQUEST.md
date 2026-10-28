# fix(config): configurable context ceilings and source roots consolidation (FIX-11)

## Summary

- Add `healthy_ceiling_tokens` directly to root `ProjectConfig`, defaulting to 200,000 tokens, eliminating the redundant `MetricsConfig` wrapper.
- Promote `anomaly_thresholds` to root `ProjectConfig` so offline ledger generation and live observability share a single source of truth.
- Standardize context ceiling saturation rules using explicit fixed token boundaries: `context_ceiling_warning_tokens` (160,000) and `context_ceiling_critical_tokens` (200,000) across offline and live anomaly evaluators.
- Dynamically compute elevated context health at `ceiling / 2` across domain health evaluators and workflow efficiency aggregations.
- Support configurable anomaly thresholds for tool retry streaks, failure rates, context spikes, and tool thrashing.
- Consolidate source code root boundaries under `code_metrics.source_roots` with bidirectional backward-compatible fallback for `workspaces.src_folders`.
- Update git diff classification and path monitoring to respect `code_metrics.source_roots`.
- Retain backward-compatible migration in configuration loader for legacy manifests using `metrics.healthy_ceiling_tokens` or `live.anomaly_thresholds`.

## Context

- Contract: `alspaces/FIX-11/CONTRACT.md`
- Summary: `alspaces/FIX-11/SUMMARY.md`
- Iteration Plans: `alspaces/FIX-11/iterations/01/PLAN.md`, `alspaces/FIX-11/iterations/02/PLAN.md`
- Iteration Reviews: `alspaces/FIX-11/iterations/01/REVIEW.md`, `alspaces/FIX-11/iterations/02/REVIEW.md`

## Verification

- Full backend test suite with race detector: `go test -race ./...` (PASS)
- Frontend test suite: `vitest run` (32 files, 734 tests passed) (PASS)
- Linting and vetting: `make check` (PASS)
- Multi-platform binary compilation: `make build` (PASS)
