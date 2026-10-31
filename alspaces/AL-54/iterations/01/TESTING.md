# AL-54 Iteration 01 Verification

Verified working diff against baseline 85a2af94b5036b0918aaf9a7a89aed3a990f2650 after coder release. Tester changed only relevant test owners; no product code changes. Full integrated make and standalone-binary checks remain parent-owned delivery gates.

## Meaningful coverage

Existing Go/TypeScript adoption tests protect repeated invocation multiplicity, distinct ticket counts, explicit success/failure rather than error text, missing/empty facts, and unidentified ledgers. Extended domain coverage protects input nonmutation, invocation rank before lexical tie order, Unicode ordering parity, all-failure rates, empty server projections, and browser skill-name deduplication and sorting.

FleetOverview test proves the displayed success rate is invocation-weighted (72.7 percent for 8 successes in 11 invocations), exactly three ranked skills appear, all four unique skills are counted, and a filtered response omitting optional facts replaces the card with no-data/unavailable copy. Existing tests already cover date/type/prefix request parameters.

TicketComparisons test exercises literal all and ALL as distinct recorded names, OR selection, all-known-skill selection excluding missing-fact tickets, AND composition with type/prefix/search, zero-result options remaining available, and restoring missing-fact tickets by clearing selection. Chart point count and benchmark workspace count track the selected population. Existing filter/sorting/recovery suites remain intact. Scatter plot test verifies recorded skill names in accessible labels and native SVG title tooltips, plus explicit missing-data copy.

No tests removed. All fixtures are self-contained and do not read ticket workspaces. Domain tests use zero mocks.

## Commands and results

- Focused Vitest: domain aggregations, domain formatters, FleetOverview, TicketComparisons, and TicketScatterPlot: 5 files / 108 tests passed.
- npm run typecheck and npm run lint: passed after test-query correction.
- make test: passed. Canonical frontend build succeeded and copied index/JS/CSS into ignored embedded assets; all 12 Go packages passed with race detector; 35 frontend files / 718 tests passed, none skipped.
- git diff --check: passed.

Initial new-test runs exposed two test author assumptions, both corrected: membership assertions originally assumed row sort order, and role-query options used an unsupported exact property. These were test-only corrections. Full final test execution passed.

The unchanged MetricProgression tests print React act warnings during full Vitest; they pass. No product behavior defect was found. Browser geometry and standalone binary execution were not claimed by these DOM/domain tests; parent owns embedding/binary delivery verification.
