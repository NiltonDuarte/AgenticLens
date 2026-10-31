# AL-57 iteration 04: Cost-card and benchmark verification

Status: PASS for focused frontend verification and coverage audit; parent integrated gate pending. Baseline 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242 was initially clean. The direct user-approved delta preserves the frozen cost contract except the explicitly scoped comparison Average Spend benchmark now averages eligible complete ticket costs in the current filtered rows.

## Existing coverage and planned recipients

Inspected actual KPICard/InfoTooltip behavior and existing Components, TicketDeepDive, FleetOverview, MetricProgression and TicketComparisons test bodies. Existing Components cases already establish zero/dash behavior and tooltip focus/blur/Escape interaction; do not duplicate placement mechanics. Existing three-view tests own normal values, selected datasets, time-window refetch, empty recovery and non-cost cards, but did not independently assert short financial headline versus coverage text.

The frontend coder plans shared CostKPICard state/content cases, representative partial-view wiring in all three affected views, and mixed comparison cost/search/reset behavior. After frontend coder handoff, the tester inspected those additions and extended only the missing composed-filter and changed-response benchmark guards described below.

The existing TicketComparisons literal-skill OR/type/prefix/search-empty recovery test is a coherent recipient for cost benchmark assertions across composed filters. Its existing type/prefix/refetch tests protect actual response selection. There is no comparison time-window control in current production: it requests window all and refetches on type/prefix changes. Test changing server-selected responses through those real controls; do not invent a time-window UI. Fleet/progression retain their actual time-window owners.

Critical oracles: primary financial value is currency/dash only; known partial subtotal remains a concise explicit secondary state; detailed invocation counts are reachable through the existing accessible tooltip, not in the large headline. Observed zero survives. Comparison eligible costs are complete finite non-null values, including zero; unknown and partial tickets remain in rows and non-cost denominators, and no known subtotal acts as a complete ticket cost. Search, type, prefix and skill composition must update both average and eligible/all ticket counts; all-unknown/empty recover after filters clear, and changed fetch results replace prior populations.

The tester remained read-only until frontend coder handoff, then changed only web/src/views/TicketComparisons.test.tsx and this report. No production, backend, preview/browser, telemetry or delivery changes were made by the tester. Tests use self-contained fixtures; DOM accessibility interactions establish content and focus behavior, not browser geometry. Parent runs the integrated make/build/embed/binary gate after final review.


## Actual recipients and changes

- Components.test.tsx, CostKPICard presentation: coder-added shared case proves short currency primary versus explicit partial/state/context secondary and detailed count tooltip; keyboard focus reaches the tooltip. Complete zero, unknown and empty are distinct. Existing KPICard and InfoTooltip owners retain generic zero/dash and focus/blur/Escape behavior; their mechanics were not duplicated across all views.
- TicketDeepDive.test.tsx, FleetOverview.test.tsx and MetricProgression.test.tsx: coder-added representative partial-card cases scope assertions to the actual named card. They verify correct ledger/fleet/bucket summary counts, concise primary, explicit partial secondary and retained view context. Progression uses actual invocation facts and bucket aggregation.
- TicketComparisons.test.tsx, averages complete filtered ticket costs including zero and recovers after unknown and empty searches: coder-added case independently verifies 6/0/partial-known-4 produces 3 USD over 2 of 3 tickets, known and zero searches, partial-only unavailability, empty search and restoration. Partial known subtotal never enters the complete-ticket denominator.
- Existing literal skill-name OR/type/prefix/search-empty recovery case: tester changed its self-contained costs to 6, 0, partial 4 and unavailable. Assertions now check average/population after skill OR, type and prefix composition, empty search, partial-only selection and clearing filters. Retained unknown/partial rows stay in Benchmarked Workspaces and non-cost averages: all three feature rows give duration 6m 0s and iterations 6.0, while spend uses only two complete tickets. Existing scatter point assertions were updated for the deliberately missing-cost fixture; unknown/partial rows remain present while their cost-axis points are omitted. This preserves the existing scatter availability contract rather than globally removing cost-ineligible rows.
- New comparison changed-response case: distinct asynchronous mock setup verifies actual prefix/type refetches replace the dataset and cost population. Values transition from 6 USD (1 of 2), to observed zero (1 of 3), to Unavailable (0 of 2), then 4 USD (2 of 2) after clearing the prefix. It checks the actual all-time prefix query and unchanged row populations. There is no new comparison time-window control.

No tests were removed. No generic state matrix or test-count quota was applied. Shared state semantics and existing tooltip mechanics cover zero/unknown/empty once; actual view wiring and filter/refetch integration have independent meaningful assertions. The finite/nonnegative/complete eligibility condition was inspected in production; malformed source numbers remain rejected by the existing API boundary.

## Verification

Baseline: 52234b3c099fe1a6c1b2e1dd3d62c1db08ad7242 plus the shared frontend implementation and tester changes. Current backend/web source-tree SHA256: 2aba254e521adf59dd0ebfc5d79b3fba6c9af117a2aefee3cc1902d94d01335f across 220 existing paths. Method: sorted unique git ls-files -co --exclude-standard paths under backend and web, hashing each existing path, NUL, bytes, NUL; ignored generated assets/dependencies excluded.

- Frontend coder's completed state before tester additions: typecheck, lint and five affected suites passed 146 tests; final build/embed verification remains coder/parent-owned. Those unchanged card/view test checks are reused rather than duplicated here.
- npm test -- --run src/views/TicketComparisons.test.tsx: all 17 tests passed after the tester's composed-filter/refetch additions, zero skipped.
- npm run typecheck: passed after final additions.
- npm run lint: passed after final additions.
- git diff --check: passed.

No production defect was found. No full suite was repeated by the tester because the parent performs the final integrated make gate after review; affected new interaction coverage passed. No browser geometry, standalone-binary behavior, source extraction, final telemetry or PR delivery is claimed by this testing report.
