# AL-57 iteration 04: Readable cost cards and filtered average spend

Status: APPROVED for implementation by direct user steering. Baseline: 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242, initially clean. Existing PR #89 remains the delivery PR. Frozen CONTRACT.md remains unchanged; this delta explicitly supersedes its blanket unavailable-average rule only for the comparison benchmark described below.

## Steering and acceptance delta

Source: user reports the cost cards in ticket deep dive, metric progression and fleet overview are unreadable, and directs comparison Average Spend to exclude tickets with unavailable costs, average the remainder, and respect filters. This authorizes the bounded presentation and benchmark fixes without another approval question.

Changed acceptance: cost card primary values contain short currency or a dash, never full coverage prose. Partial cost stays clearly identified as a known subtotal through concise secondary/status text. Detailed coverage remains accessible through the existing information tooltip. Comparison Average Spend uses only complete numeric ticket costs from the current filtered rows and states its eligible population.

Preserved guarantees: observed zero is eligible; partial subtotals never act as complete ticket costs; unknown/empty is explicit; invocation coverage and historical numeric-observed policy remain; table/CLI/schema/backend behavior and source ownership remain unchanged. No estimate, new financial dataset, chart semantics rewrite, or global design refactor is authorized.

Superseded decision: an unavailable ticket no longer makes this particular comparison average unavailable when another filtered ticket has complete cost. This is a benchmark over the eligible observed-ticket subset, not a full-population cost average or an average of known partial sums. Other aggregates and ratios retain their accepted availability rules.

## Grounded problem and ownership

- web/src/domain/formatters.ts formatCostUSD returns full sentences with cost coverage. That is suitable for existing verbose displays but is passed directly into KPICard.value in TicketDeepDive.tsx, MetricProgression.tsx and FleetOverview.tsx.
- components/KPICard.tsx renders value in .kpi-value mono-number, with separate secondary and tooltip props already available. index.css defines the large number style. The fix is to use these semantic areas; do not reduce the entire primary font, clip/ellipsis the evidence, or alter unrelated cards to accommodate a metadata sentence.
- TicketComparisons.tsx filteredRows already composes the currently fetched dataset with ticket type, prefix, skills and search. Comparison requests the all-time dataset and has no time-window control; type/prefix selections can refetch that dataset. Its benchmarks useMemo instead null-propagates totCost across every row and divides by all rows, causing Average Spend to disappear for mixed availability.

## Implementation plan

1. Add a small shared cost-card presentation helper/component used by the three affected views. Reuse KPICard and existing InfoTooltip. Return/render a short primary value, concise state/subtitle, and detailed accessible tooltip independently. Keep the existing formatCostUSD table formatter unchanged. No new dependency or backend change is needed.
2. Complete state: primary formatted amount, including $0.00; keep the view's useful existing secondary context. Partial state: primary formatted known subtotal, secondary explicitly identifies Known subtotal / Partial without repeating all counts. Unavailable state: dash primary with Unavailable secondary. Empty state: dash primary with No invocation data secondary. State must remain understandable without relying on color or hover alone. Preserve any existing view-context meaning in concise secondary text or tooltip; do not replace important scope with a generic label.
3. Tooltip combines the existing view's explanation with full invocation coverage counts and the distinction between complete total and partial subtotal. Reuse keyboard-focusable InfoTooltip and its accessible name; long counts must not enter the large primary number. Respect existing theme and responsive layout; use existing styling unless a focused component style is necessary. Do not introduce nested tooltip triggers or inaccessible title-only evidence.
4. In TicketComparisons benchmarks, independently accumulate cost sum and eligible cost-ticket count from filteredRows. Eligible means complete cost with a non-null finite numeric cost_usd; accepted wire types normally encode incomplete costs as null. Include numeric zero. Exclude partial/unavailable amounts and never substitute known_subtotal_usd. Average is sum divided by eligible count only when count is positive; otherwise render Unavailable for both all-unavailable and empty selections.
5. Keep totalTickets, duration, iterations, tool reliability, rows and scatter populations based on all filteredRows. Only Average Spend changes its denominator. Its concise secondary states the eligible population, for example “2 of 3 tickets with cost”; tooltip explains excluded unavailable/partial tickets and filtered scope. Preserve its clear title. Recompute on every existing filtering/search path and refreshed/changed ledgers through filteredRows dependencies. Do not globally remove cost-ineligible tickets from the view.

## Acceptance examples

| Current filtered ticket costs | Average Spend | Eligible subtitle |
| --- | --- | --- |
| 6, 0, null | 3 USD | 2 of 3 tickets with cost |
| 6, null with known partial subtotal 4 | 6 USD | 1 of 2 tickets with cost |
| 0, null | 0 USD | 1 of 2 tickets with cost |
| null, null | Unavailable | 0 of 2 tickets with cost |
| No rows | Unavailable | 0 of 0 tickets with cost |

Filtering the first example to only the 6 USD ticket yields 6 USD and 1 of 1; selecting only its unavailable ticket yields Unavailable and 0 of 1. Clearing the filter restores the eligible average and all original rows. A newly fetched dataset must replace these results rather than retain the previous count.

## Targeted verification and recipients

- Extend web/src/domain/formatters.test.ts if the presentation helper is pure, or a focused shared component test if it owns accessible rendering. Exercise primary/secondary separation for complete, partial, unavailable, empty and observed zero; meaningful coverage details must remain outside the large value and accessible on keyboard focus. Reuse existing InfoTooltip tests if present rather than duplicating tooltip placement mechanics.
- In TicketDeepDive.test.tsx, FleetOverview.test.tsx and MetricProgression.test.tsx, scope assertions to the actual named cost card and verify the shared card is wired to the correct selected ledger/fleet/bucket summary. Cover representative partial and unavailable/zero transitions across these view owners without repeating a full state matrix in every suite. Preserve existing non-cost cards, empty recovery controls and async loading assertions.
- Extend TicketComparisons.test.tsx's existing search/reset, ticket-type/prefix and literal-skill composition cases with mixed complete/non-null zero/partial/unavailable rows and independent expected averages/counts. Check each active filter affects the benchmark, other metrics/table keep the full filtered population, all-unavailable and empty states recover, and changed refetched data updates the denominator through the actual type/prefix refetch path. Do not add a comparison time-window control. Use real user interactions and asynchronous mocked response changes; testing only an isolated average helper does not prove filter integration.
- Existing progression time-window/filter tests and fleet refetch tests remain the owners of selected-scope behavior. Preserve the chart gap, cumulative availability and table cost formatting assertions unchanged unless an actual correction requires an explicit rationale. Tests use self-contained fixtures; no live ticket artifacts or source database.

Run focused frontend tests/typecheck/lint during fixes, then required integrated make and frontend build/embedded asset/binary checks for final delivery. Record tested source, commands, results and skips in iteration 04 review. DOM/accessibility tests establish content and interactions; they do not prove browser geometry. Do not use preview tools to verify these edits or claim browser layout inspection. Parent owns integration, telemetry, commits, PR update and handoff; planner writes only this plan.


Grounding correction from tester: comparison currently fetches the all-time dataset and has no time-window control. Verify its actual search/type/prefix/skill filters and data replacement; existing progression/fleet suites retain ownership of their own window controls. This corrects a plan assumption without expanding scope or requiring renewed approval.


Review correction: user explicitly requested removing the "time buckets" subtitle from the progression cost card. Remove that secondary context only; retain visible cost state, accessible coverage, date-range explanation and all chart/time aggregation behavior. This supersedes the earlier requirement to retain that particular context. Update the existing partial-card expectation without adding new tests for a reversible copy-only change. Keep this correction in iteration 04.
