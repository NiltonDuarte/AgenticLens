# fix(config): enforce critical context ceilings and consolidate source roots (FIX-11)

## Summary

- Complete architectural excision of `healthy_ceiling_tokens` across domain, configuration loader/writer, and `.agenticlens-config.yaml`.
- Enforce fail-fast configuration validation requiring positive `anomaly_thresholds.context_ceiling_critical_tokens` and `anomaly_thresholds.context_ceiling_warning_tokens`, rejecting non-positive, missing, or inverted threshold relationships.
- Establish `AnomalyThresholds.ContextCeilingCriticalTokens` as the single authoritative source of truth for context saturation percentage and health status grading across all domain aggregations and anomaly detection.
- Promote `anomaly_thresholds` to root `ProjectConfig` so offline ledger generation and live observability share a single source of truth.
- Standardize context ceiling saturation rules using explicit fixed token boundaries: `context_ceiling_warning_tokens` (160,000) and `context_ceiling_critical_tokens` (200,000) across offline and live anomaly evaluators.
- Dynamically compute elevated context health at `critical_ceiling / 2` across domain health evaluators and workflow efficiency aggregations.
- Support configurable anomaly thresholds for tool retry streaks, failure rates, context spikes, and tool thrashing.
- Consolidate source code root boundaries under `code_metrics.source_roots` with bidirectional backward-compatible fallback for `workspaces.src_folders`.
- Update git diff classification and path monitoring to respect `code_metrics.source_roots`.

## Context

- Contract: `alspaces/FIX-11/CONTRACT.md`
- Summary: `alspaces/FIX-11/SUMMARY.md`
- Iteration Plans: `alspaces/FIX-11/iterations/01/PLAN.md`, `alspaces/FIX-11/iterations/02/PLAN.md`, `alspaces/FIX-11/iterations/03/PLAN.md`
- Iteration Reviews: `alspaces/FIX-11/iterations/01/REVIEW.md`, `alspaces/FIX-11/iterations/02/REVIEW.md`, `alspaces/FIX-11/iterations/03/REVIEW.md`

## Verification

- Full backend test suite with race detector: `go test -race -count=1 ./...` (12 packages PASS)
- Frontend test suite: `vitest run` (32 files, 734 tests passed) (PASS)
- Linting and vetting: `make check` (PASS)
- Multi-platform binary compilation: `make build` (PASS)
