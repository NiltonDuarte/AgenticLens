# AL-48 Iteration 01 Plan

## Grounding

- `web/src/views/TicketComparisons.tsx` owns ledger loading, current type/prefix/search filters, derives `allRows` and `filteredRows`, renders the comparison table, and already accepts `onSelectTicket`.
- `web/src/domain/types.ts` defines `TicketComparisonRow`; it includes cost, line additions/deletions, active duration, iterations, cache hit rate, tool success rate, `root_saturation_pct`, nullable lead duration, title, and completion timestamp.
- `web/src/domain/formatters.ts` provides `deriveTicketComparisonRows`, `formatCurrencyUSD`, `formatPercentage`, `formatDurationSeconds`, and lead-time formatting. Net lines are not currently a row field and must be derived as additions minus deletions. Requested orchestrator saturation maps to existing `root_saturation_pct`.
- `web/src/views/TicketComparisons.test.tsx` already tests comparison loading, filters, table behavior, and `onSelectTicket`; extend this file for view-level integration.
- `web/src/test/fixtures/mockData.ts` contains typed comparison fixtures. `web/package.json` exposes `typecheck`, `lint`, and `test` scripts; Vitest and Testing Library are already installed.

## Work Items

1. Add a pure scatter-plot metric/scaling module under `web/src/domain/` with a closed metric key definition and display metadata for all requested metrics. Extract numeric coordinates from `TicketComparisonRow`, deriving net lines and mapping saturation to `root_saturation_pct`; represent unavailable lead duration as absent.
2. In the pure module, filter excluded and unplottable rows before calculating axis recommendation and bounds. Implement linear and zero-safe logarithmic transforms, the >50x span recommendation for non-negative values, and 10% padded extents with explicit handling for no points and constant-valued ranges. Keep negative values plottable in linear mode and prevent invalid/non-finite logarithmic positions.
3. Build `web/src/components/TicketScatterPlot.tsx` using React and native SVG. Include the exact title, independent axis selectors, scale mode control with non-forcing recommendation, accessible exclusion checklist, completion/in-progress color distinction, axes/labels, and hover/focus details. Render only included rows with valid coordinates and invoke `onSelectTicket` from point activation.
4. Integrate the chart into `web/src/views/TicketComparisons.tsx` using the current comparison rows after existing ticket-type, prefix, and search filters. Preserve the existing table and thread `onSelectTicket` through to the chart.
5. Add `TicketScatterPlot.test.tsx` for metric extraction, defaults, scale transforms/recommendation, padding and degenerate/empty inputs, exclusion-driven coordinates and bounds, nullable metric handling, SVG points, accessible tooltip content, status colors, and point selection.
6. Extend `TicketComparisons.test.tsx` to verify title/default controls and ticket click routing in the integrated view, using existing fetch mocks and comparison fixtures.

## Verification

- Run focused component and view tests during implementation.
- Run `npm run typecheck`, `npm run lint`, and `npm test` from `web/`.
- Run repository `make` before delivery to cover the configured quality gate.
- Confirm no dependencies or backend/API files changed and no TypeScript `any` was introduced.

## Acceptance Boundary

Iteration 01 is complete when the chart meets the contract's interaction and scaling behavior, the view integrates the active comparison rows and navigation callback, tests cover the requested cases, and the listed quality gates pass.
