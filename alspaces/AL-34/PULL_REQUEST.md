# feat(progression): eliminate cumulative granularity and implement bidirectional code churn visualization (AL-34)

## Summary of Changes

This pull request resolves ticket AL-34 by eliminating cumulative granularity from temporal progression views and introducing a bi-directional diverging code churn visualization centered on a zero baseline.

### Core Changes
1. **Domain Core (Go & TypeScript)**:
   - Eliminated `GranularityCumulative` from `TimeGranularity` enum in `backend/internal/domain/enums.go` and `web/src/domain/types.ts`.
   - Updated `ParseTimeGranularity` to validate strictly `hourly`, `daily`, and `weekly`.
   - Added `CumulativeNetLinesAdded` (`cumulative_net_lines_added`) to `TimeBucketMetric` in Go and TypeScript, calculated as `runningCumLinesAdded - runningCumLinesDeleted`.
   - Implemented `formatCompactCount` and `formatLinesCompact` in `web/src/domain/formatters.ts` supporting `k`, `M`, and sign prefixes (`+` / `-`).
2. **Boundary Adapters (Go HTTP)**:
   - Updated `handleProgression` in `backend/internal/adapters/http/handlers.go` to reject non-temporal granularities (including `cumulative`) with HTTP 400 Bad Request.
   - Added adversarial test suites in `backend/internal/adapters/http/server_test.go` and `backend/internal/domain/aggregations_test.go`.
3. **Presentation (TypeScript & SVG)**:
   - Upgraded `TrendChart.tsx` to support bi-directional diverging bars centered on `y = getYCoord(0)`, custom `yAxisFormatter`, integer tick rounding, and container overflow containment.
   - Removed "Cumulative" toggle button from `MetricProgression.tsx`.
   - Maintained Chart 1 as a composite spend view with static title and period spend bars plus cumulative trajectory line.
   - Configured Chart 4 ("Delivered Code Line Churn & Net Growth") to render positive additions upward (green), negative deletions downward (red), and running cumulative net growth as a trajectory line with compact Y-axis tick formatting.

## Architecture & Invariants Compliance
- **FCIS**: Strict separation between pure domain aggregations (`backend/internal/domain/`) and imperative HTTP adapters (`backend/internal/adapters/`).
- **Typing**: Zero `any` or `interface{}` across Go and TypeScript codebases.
- **Tone & Style**: Zero emojis across code, comments, test names, and UI elements.

## References
- Architectural Contract: `workspaces/AL-34/CONTRACT.md`
- Work Summary: `workspaces/AL-34/SUMMARY.md`
- Plan & Reviews: `workspaces/AL-34/iterations/01/PLAN.md` and `workspaces/AL-34/iterations/01/REVIEW.md`
