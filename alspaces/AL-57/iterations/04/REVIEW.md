# AL-57 iteration 04: Independent presentation and benchmark audit

Status: APPROVED. No open findings. The parent integration/delivery gate remains separate.

## Reviewed authority and source

Baseline: 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242. Reviewed approved iterations/04/PLAN.md, frozen CONTRACT.md and the direct user correction: readable cost cards in deep dive, progression and fleet; comparison Average Spend excludes unavailable costs and averages the current filtered eligible remainder. The plan explicitly supersedes the old blanket unavailable-average behavior only for this comparison benchmark. No repeated scope approval is required.

Reviewed changed production, documentation and test bodies, including new CostKPICard.tsx and the tester's filter/refetch additions. Independently recomputed backend/web source SHA256: 2aba254e521adf59dd0ebfc5d79b3fba6c9af117a2aefee3cc1902d94d01335f across 220 existing files. Method: sorted unique git ls-files -co --exclude-standard -- backend web paths; hash each existing path, NUL, file bytes, NUL. Ignored dependencies/generated assets are excluded. This matches the tester's final source snapshot. No production or test edits by this reviewer.

## Axis 1: Acceptance alignment

CostKPICard reuses KPICard and InfoTooltip. Its large primary contains formatted currency for complete cost or a known partial contribution, otherwise a dash. The secondary visibly identifies Known subtotal (partial), Unavailable or No invocation data and preserves each view's existing scope context. A partial number cannot be mistaken for a complete total without ignoring that visible label. Observed zero remains currency zero. Full count prose no longer enters the large primary.

The tooltip combines the view explanation, complete/partial distinction and fully observed/partial/unavailable/total invocation counts. It retains the existing focusable button, accessible name and aria-describedby relationship; coverage does not depend on color or hover alone. No new nested triggers, CSS clipping, ellipsis, global typography reduction or hidden coverage workaround is introduced. Existing secondary text has normal wrapping. This is a source/DOM/accessibility review, not a claim of measured browser geometry.

Verified all three actual view integrations:

- TicketDeepDive passes its selected ledger total, summary and invocation count; retains invocation scope.
- FleetOverview passes the returned fleet complete amount, summary and total invocation count; retains active-ticket scope.
- MetricProgression passes its selected-bucket merged cost summary and invocation count; retains time-bucket context. Existing bucket/filter/chart calculations are unchanged.

The comparison benchmark independently sums/counts complete finite nonnegative ticket amounts from filteredRows, including numeric zero and compatible numeric legacy rows without a summary. It excludes null/nonfinite/partial amounts and never substitutes a known partial subtotal. The denominator is eligible tickets, not invocations or all displayed tickets. Zero eligible tickets produces Unavailable for both all-unknown and empty populations. The visible eligible/full ticket caption and tooltip disclose the selected population.

Traced filteredRows through current ledgers, type/prefix/skill composition, search and response replacement. TicketComparisons actually requests the all-time dataset and has no time-window control; type/prefix selection refetches that dataset. Memo dependencies include the resulting filteredRows, so both average and counts update with those real controls and changed responses. Other duration/iteration/reliability benchmarks, table rows and scatter inputs still receive the full filtered population. Cost-ineligible tickets are not globally removed.

Backend/schema/typed JSON behavior, table and CLI cost formatting, migration, source ownership and chart-gap semantics are unchanged. No estimate, historical special state or unrelated design refactor is added. docs/COST_SEMANTICS.md describes the benchmark subset exception without changing invocation coverage semantics.

## Tests inspected and verification reused

Inspected bodies and existing recipients rather than imposing a test-count matrix:

- Components.test.tsx exercises short primary versus visible partial state/context, large invocation counts outside the primary, keyboard focus exposing the existing tooltip, and zero/unavailable/empty transitions. Existing InfoTooltip cases retain focus/blur/Escape behavior.
- TicketDeepDive.test.tsx and FleetOverview.test.tsx scope assertions to the actual named cost card, checking its selected summary and complete counts. MetricProgression.test.tsx derives the card from actual invocation buckets containing a partial contribution and observed zero.
- TicketComparisons.test.tsx extends the existing literal-skill OR/type/prefix/search-empty recovery owner with 6/0/partial/unavailable tickets. Its expected averages and eligible/full counts change at each interaction. Duration/iteration expected values deliberately retain all filtered tickets; cost-axis omissions do not erase table rows.
- The dedicated existing-suite search case checks selecting a positive ticket, observed zero, only partial cost, no rows and clearing search. A distinct asynchronous setup exercises changed prefix/type refetch payloads, moving the average 6 -> 0 -> Unavailable -> 4 and replacing eligible/full counts through the real all-time query path. No artificial comparison window control or live workspace fixture is introduced.

Coder reports typecheck, lint, all 146 focused tests in five affected suites, production build and matching ignored embed tree passed. Independent tester reports all 17 TicketComparisons tests plus typecheck/lint passed after the final composition/refetch additions, with no skips or production findings. These valid results were reused rather than duplicated without a concrete concern. git diff --check passed during this audit.

## Remaining delivery boundaries

No browser geometry, screenshot or preview verification was performed or claimed. Parent owns the required integrated make gate, current frontend/embedded asset build and standalone binary checks, final real telemetry, commits and update of the existing PR. Earlier full-suite results apply to their recorded snapshots; they do not replace that final gate.

Final verdict: APPROVED.
